// Package matrix implements device and environment matrix intelligence.
//
// It answers questions like:
//   - "Which environment classes have coverage gaps?"
//   - "Are tests concentrated on a single device when the class has many?"
//   - "Which representative devices should we recommend for a given change?"
//
// The matrix engine operates on the dependency graph and produces
// recommendations that are conservative: it only recommends devices
// and environments where evidence suggests coverage is relevant.
package matrix

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/depgraph"
)

// MatrixResult is the output of environment/device matrix analysis.
type MatrixResult struct {
	// Classes contains coverage analysis per environment class.
	Classes []ClassCoverage `json:"classes"`

	// Gaps lists environment class members with no test coverage.
	Gaps []CoverageGap `json:"gaps,omitempty"`

	// Concentrations lists classes where coverage is heavily skewed.
	Concentrations []Concentration `json:"concentrations,omitempty"`

	// Recommendations suggests representative devices/environments to add.
	Recommendations []DeviceRecommendation `json:"recommendations,omitempty"`

	// TestsAnalyzed is how many test files were examined.
	TestsAnalyzed int `json:"testsAnalyzed"`

	// ClassesAnalyzed is how many environment classes were examined.
	ClassesAnalyzed int `json:"classesAnalyzed"`
}

// ClassCoverage describes test coverage across an environment class.
type ClassCoverage struct {
	// ClassID is the environment class identifier.
	ClassID string `json:"classId"`

	// ClassName is the human-readable name.
	ClassName string `json:"className"`

	// Dimension is what this class varies on (os, browser, device, runtime).
	Dimension string `json:"dimension"`

	// TotalMembers is the number of environments/devices in this class.
	TotalMembers int `json:"totalMembers"`

	// CoveredMembers is the number with at least one test targeting them.
	CoveredMembers int `json:"coveredMembers"`

	// CoverageRatio is CoveredMembers / TotalMembers.
	CoverageRatio float64 `json:"coverageRatio"`

	// Members lists each member with its coverage status.
	Members []MemberCoverage `json:"members"`
}

// MemberCoverage describes coverage for a single class member.
type MemberCoverage struct {
	// ID is the environment or device ID.
	ID string `json:"id"`

	// Name is the human-readable name.
	Name string `json:"name"`

	// TestCount is how many test files target this member.
	TestCount int `json:"testCount"`

	// Covered is true if TestCount > 0.
	Covered bool `json:"covered"`
}

// CoverageGap identifies an environment class member with no test coverage.
type CoverageGap struct {
	// ClassID is the environment class this gap belongs to.
	ClassID string `json:"classId"`

	// ClassName is the class's human-readable name.
	ClassName string `json:"className"`

	// Dimension is the class dimension.
	Dimension string `json:"dimension"`

	// MemberID is the uncovered member's ID.
	MemberID string `json:"memberId"`

	// MemberName is the uncovered member's name.
	MemberName string `json:"memberName"`
}

// Concentration identifies a class where test coverage is heavily skewed
// toward one member while others are neglected.
type Concentration struct {
	// ClassID is the concentrated class.
	ClassID string `json:"classId"`

	// ClassName is the class's human-readable name.
	ClassName string `json:"className"`

	// Dimension is the class dimension.
	Dimension string `json:"dimension"`

	// DominantMember is the member with the most test coverage.
	DominantMember string `json:"dominantMember"`

	// DominantName is the dominant member's human-readable name.
	DominantName string `json:"dominantName"`

	// DominantShare is the fraction of total test coverage on the dominant member.
	DominantShare float64 `json:"dominantShare"`

	// TotalMembers is the class size.
	TotalMembers int `json:"totalMembers"`

	// CoveredMembers is how many members have any coverage.
	CoveredMembers int `json:"coveredMembers"`
}

// DeviceRecommendation suggests a device or environment to add test coverage for.
type DeviceRecommendation struct {
	// MemberID is the recommended environment/device ID.
	MemberID string `json:"memberId"`

	// MemberName is the human-readable name.
	MemberName string `json:"memberName"`

	// ClassID is the class this recommendation belongs to.
	ClassID string `json:"classId"`

	// ClassName is the class's human-readable name.
	ClassName string `json:"className"`

	// Dimension is the class dimension.
	Dimension string `json:"dimension"`

	// Reason explains why this recommendation is made.
	Reason string `json:"reason"`

	// Priority ranks recommendations (1 = most important).
	Priority int `json:"priority"`
}

// Analyze performs environment/device matrix coverage analysis on the
// dependency graph. It examines environment classes, determines which
// members have test coverage, identifies gaps and concentrations, and
// produces conservative recommendations.
func Analyze(g *depgraph.Graph) *MatrixResult {
	result := &MatrixResult{}
	if g == nil {
		return result
	}

	// Count test files analyzed.
	testFiles := g.NodesByType(depgraph.NodeTestFile)
	result.TestsAnalyzed = len(testFiles)

	// Find all environment classes.
	classNodes := g.NodesByType(depgraph.NodeEnvironmentClass)
	result.ClassesAnalyzed = len(classNodes)
	if len(classNodes) == 0 {
		return result
	}

	// Build reverse index: member → test files that target it.
	memberTests := buildMemberTestIndex(g, testFiles)

	// Analyze each class.
	for _, classNode := range classNodes {
		cc := analyzeClass(g, classNode, memberTests)
		result.Classes = append(result.Classes, cc)
	}

	// Sort classes by coverage ratio ascending (worst coverage first).
	sort.SliceStable(result.Classes, func(i, j int) bool {
		return result.Classes[i].CoverageRatio < result.Classes[j].CoverageRatio
	})

	// Extract gaps and concentrations.
	result.Gaps = extractGaps(result.Classes)
	result.Concentrations = extractConcentrations(result.Classes)

	// Build recommendations.
	result.Recommendations = buildRecommendations(result.Gaps, result.Concentrations, result.Classes)

	return result
}

// RecommendForTests takes a set of impacted test file paths and recommends
// representative devices/environments for those tests. This is the entry
// point for change-scoped matrix recommendations.
func RecommendForTests(g *depgraph.Graph, testPaths []string) []DeviceRecommendation {
	if g == nil || len(testPaths) == 0 {
		return nil
	}

	// Find which environments/devices these tests currently target.
	targetedMembers := map[string]bool{}
	for _, path := range testPaths {
		fileID := "file:" + path
		for _, e := range g.Outgoing(fileID) {
			if e.Type == depgraph.EdgeTargetsEnvironment {
				targetedMembers[e.To] = true
			}
		}
	}

	// Find environment classes that contain targeted members.
	classNodes := g.NodesByType(depgraph.NodeEnvironmentClass)
	var recs []DeviceRecommendation

	for _, classNode := range classNodes {
		members := classMembers(g, classNode.ID)
		if len(members) < 2 {
			continue
		}

		// Check if any member of this class is targeted.
		classRelevant := false
		coveredInClass := 0
		for _, m := range members {
			if targetedMembers[m.ID] {
				classRelevant = true
				coveredInClass++
			}
		}
		if !classRelevant {
			continue
		}

		// Recommend uncovered members of relevant classes.
		dimension := classNode.Metadata["dimension"]
		for _, m := range members {
			if !targetedMembers[m.ID] {
				recs = append(recs, DeviceRecommendation{
					MemberID:   m.ID,
					MemberName: m.Name,
					ClassID:    classNode.ID,
					ClassName:  classNode.Name,
					Dimension:  dimension,
					Reason: fmt.Sprintf(
						"Tests target %d of %d %s members — %s is untested",
						coveredInClass, len(members), dimension, m.Name),
				})
			}
		}
	}

	// Sort by class, then member name for determinism.
	sort.SliceStable(recs, func(i, j int) bool {
		if recs[i].ClassID != recs[j].ClassID {
			return recs[i].ClassID < recs[j].ClassID
		}
		return recs[i].MemberID < recs[j].MemberID
	})

	for i := range recs {
		recs[i].Priority = i + 1
	}

	return recs
}

// --- Internal helpers ---

// buildMemberTestIndex builds a reverse index: member ID → count of
// test files targeting it.
func buildMemberTestIndex(g *depgraph.Graph, testFiles []*depgraph.Node) map[string]int {
	index := map[string]int{}
	for _, tf := range testFiles {
		for _, e := range g.Outgoing(tf.ID) {
			if e.Type == depgraph.EdgeTargetsEnvironment {
				index[e.To]++
			}
		}
	}
	return index
}

type memberInfo struct {
	ID   string
	Name string
}

// classMembers returns all member nodes of an environment class.
func classMembers(g *depgraph.Graph, classID string) []memberInfo {
	var members []memberInfo
	for _, e := range g.Outgoing(classID) {
		if e.Type == depgraph.EdgeEnvironmentClassContains {
			n := g.Node(e.To)
			if n != nil {
				members = append(members, memberInfo{ID: n.ID, Name: n.Name})
			}
		}
	}
	// Sort for determinism.
	sort.Slice(members, func(i, j int) bool {
		return members[i].ID < members[j].ID
	})
	return members
}

// analyzeClass computes coverage for a single environment class.
func analyzeClass(g *depgraph.Graph, classNode *depgraph.Node, memberTests map[string]int) ClassCoverage {
	members := classMembers(g, classNode.ID)
	dimension := classNode.Metadata["dimension"]

	cc := ClassCoverage{
		ClassID:      classNode.ID,
		ClassName:    classNode.Name,
		Dimension:    dimension,
		TotalMembers: len(members),
	}

	for _, m := range members {
		count := memberTests[m.ID]
		mc := MemberCoverage{
			ID:        m.ID,
			Name:      m.Name,
			TestCount: count,
			Covered:   count > 0,
		}
		cc.Members = append(cc.Members, mc)
		if count > 0 {
			cc.CoveredMembers++
		}
	}

	if cc.TotalMembers > 0 {
		cc.CoverageRatio = float64(cc.CoveredMembers) / float64(cc.TotalMembers)
	}

	return cc
}

// extractGaps identifies uncovered members across all classes.
func extractGaps(classes []ClassCoverage) []CoverageGap {
	var gaps []CoverageGap
	for _, cc := range classes {
		for _, m := range cc.Members {
			if !m.Covered {
				gaps = append(gaps, CoverageGap{
					ClassID:    cc.ClassID,
					ClassName:  cc.ClassName,
					Dimension:  cc.Dimension,
					MemberID:   m.ID,
					MemberName: m.Name,
				})
			}
		}
	}
	return gaps
}

// extractConcentrations finds classes where coverage is heavily skewed.
// A class is concentrated if one member holds > 70% of all test coverage
// within the class and there are ≥ 2 members.
func extractConcentrations(classes []ClassCoverage) []Concentration {
	var concs []Concentration
	for _, cc := range classes {
		if cc.TotalMembers < 2 {
			continue
		}

		totalTests := 0
		maxTests := 0
		var dominant MemberCoverage
		for _, m := range cc.Members {
			totalTests += m.TestCount
			if m.TestCount > maxTests {
				maxTests = m.TestCount
				dominant = m
			}
		}

		if totalTests == 0 {
			continue
		}

		share := float64(maxTests) / float64(totalTests)
		if share > 0.70 && cc.CoveredMembers < cc.TotalMembers {
			concs = append(concs, Concentration{
				ClassID:        cc.ClassID,
				ClassName:      cc.ClassName,
				Dimension:      cc.Dimension,
				DominantMember: dominant.ID,
				DominantName:   dominant.Name,
				DominantShare:  share,
				TotalMembers:   cc.TotalMembers,
				CoveredMembers: cc.CoveredMembers,
			})
		}
	}
	return concs
}

// buildRecommendations produces conservative device/environment recommendations
// from gaps and concentrations. Gaps take priority over concentrations.
func buildRecommendations(gaps []CoverageGap, concs []Concentration, classes []ClassCoverage) []DeviceRecommendation {
	var recs []DeviceRecommendation
	seen := map[string]bool{}

	// Priority 1: gaps — uncovered members in classes that have some coverage.
	for _, gap := range gaps {
		// Only recommend if the class has at least one covered member
		// (otherwise the class may not be relevant to the project).
		for _, cc := range classes {
			if cc.ClassID == gap.ClassID && cc.CoveredMembers > 0 {
				key := gap.MemberID
				if !seen[key] {
					seen[key] = true
					recs = append(recs, DeviceRecommendation{
						MemberID:   gap.MemberID,
						MemberName: gap.MemberName,
						ClassID:    gap.ClassID,
						ClassName:  gap.ClassName,
						Dimension:  gap.Dimension,
						Reason: fmt.Sprintf(
							"No tests target %s — %d of %d %s class members have coverage",
							gap.MemberName, cc.CoveredMembers, cc.TotalMembers, gap.Dimension),
					})
				}
				break
			}
		}
	}

	// Priority 2: concentration — suggest underserved members.
	for _, conc := range concs {
		for _, cc := range classes {
			if cc.ClassID != conc.ClassID {
				continue
			}
			for _, m := range cc.Members {
				if m.Covered && m.ID != conc.DominantMember {
					continue // already has some coverage
				}
				if !m.Covered && !seen[m.ID] {
					seen[m.ID] = true
					recs = append(recs, DeviceRecommendation{
						MemberID:   m.ID,
						MemberName: m.Name,
						ClassID:    conc.ClassID,
						ClassName:  conc.ClassName,
						Dimension:  conc.Dimension,
						Reason: fmt.Sprintf(
							"%.0f%% of %s tests target only %s — diversify coverage",
							conc.DominantShare*100, conc.Dimension,
							conc.DominantName),
					})
				}
			}
			break
		}
	}

	// Cap at 10 recommendations.
	if len(recs) > 10 {
		recs = recs[:10]
	}

	for i := range recs {
		recs[i].Priority = i + 1
	}

	return recs
}

// FormatSummary produces a human-readable summary of the matrix result.
func FormatSummary(r *MatrixResult) string {
	if r == nil || len(r.Classes) == 0 {
		return "No environment classes defined — matrix analysis not applicable."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Matrix Coverage: %d classes, %d test files analyzed\n",
		r.ClassesAnalyzed, r.TestsAnalyzed))

	for _, cc := range r.Classes {
		sb.WriteString(fmt.Sprintf("  [%s] %s: %d/%d members covered (%.0f%%)\n",
			cc.Dimension, cc.ClassName, cc.CoveredMembers, cc.TotalMembers, cc.CoverageRatio*100))
	}

	if len(r.Gaps) > 0 {
		sb.WriteString(fmt.Sprintf("\nGaps: %d uncovered members\n", len(r.Gaps)))
		for _, gap := range r.Gaps {
			sb.WriteString(fmt.Sprintf("  - %s (%s/%s)\n", gap.MemberName, gap.ClassName, gap.Dimension))
		}
	}

	if len(r.Concentrations) > 0 {
		sb.WriteString(fmt.Sprintf("\nConcentrations: %d skewed classes\n", len(r.Concentrations)))
		for _, conc := range r.Concentrations {
			sb.WriteString(fmt.Sprintf("  - %s: %.0f%% of tests on %s\n",
				conc.ClassName, conc.DominantShare*100, conc.DominantName))
		}
	}

	if len(r.Recommendations) > 0 {
		sb.WriteString(fmt.Sprintf("\nRecommendations: %d devices/environments to consider\n", len(r.Recommendations)))
		for _, rec := range r.Recommendations {
			sb.WriteString(fmt.Sprintf("  %d. %s — %s\n", rec.Priority, rec.MemberName, rec.Reason))
		}
	}

	return sb.String()
}
