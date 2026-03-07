package migration

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
)

// ReadinessSummary summarizes migration readiness for a repository.
type ReadinessSummary struct {
	// Frameworks lists detected frameworks with file counts.
	Frameworks []models.Framework `json:"frameworks"`

	// TotalBlockers is the count of migration-related signals.
	TotalBlockers int `json:"totalBlockers"`

	// BlockersByType groups blocker counts by blocker taxonomy category.
	BlockersByType map[string]int `json:"blockersByType"`

	// RepresentativeBlockers shows a few example blockers.
	RepresentativeBlockers []BlockerExample `json:"representativeBlockers,omitempty"`

	// ReadinessLevel is a qualitative assessment: low, medium, high.
	// Derived from visible blocker patterns and risk — not a magic score.
	ReadinessLevel string `json:"readinessLevel"`

	// Explanation describes why the readiness level was assigned.
	Explanation string `json:"explanation"`

	// QualityFactors describes how quality signals compound migration risk.
	// Present only when quality signals co-occur with migration targets.
	QualityFactors []QualityFactor `json:"qualityFactors,omitempty"`

	// AreaAssessments classifies directories by migration safety,
	// combining blocker presence with quality signal density.
	AreaAssessments []AreaAssessment `json:"areaAssessments,omitempty"`

	// CoverageGuidance suggests where additional test coverage would
	// most reduce migration risk.
	CoverageGuidance []CoverageGuidanceItem `json:"coverageGuidance,omitempty"`
}

// BlockerExample is a representative migration blocker for display.
type BlockerExample struct {
	Type        string `json:"type"`
	File        string `json:"file"`
	Explanation string `json:"explanation"`
}

// QualityFactor describes how a quality signal type compounds migration risk.
type QualityFactor struct {
	// SignalType is the quality signal (e.g. "weakAssertion").
	SignalType string `json:"signalType"`

	// AffectedFiles is the count of migration-target files also affected
	// by this quality issue.
	AffectedFiles int `json:"affectedFiles"`

	// Explanation describes the compounding effect.
	Explanation string `json:"explanation"`
}

// AreaAssessment classifies a directory's migration safety based on
// the combination of migration blockers and quality signals.
type AreaAssessment struct {
	// Directory is the assessed directory path.
	Directory string `json:"directory"`

	// Classification is one of: "safe", "caution", "risky".
	Classification string `json:"classification"`

	// MigrationBlockers is the count of migration signals in this directory.
	MigrationBlockers int `json:"migrationBlockers"`

	// QualityIssues is the count of quality signals in this directory.
	QualityIssues int `json:"qualityIssues"`

	// TestFileCount is the number of test files in this directory.
	TestFileCount int `json:"testFileCount"`

	// Explanation describes why the classification was assigned.
	Explanation string `json:"explanation"`
}

// CoverageGuidanceItem suggests where additional coverage would reduce
// migration risk the most.
type CoverageGuidanceItem struct {
	// Directory is the area that would benefit from more coverage.
	Directory string `json:"directory"`

	// Reason explains why coverage matters here for migration.
	Reason string `json:"reason"`

	// Priority is "high", "medium", or "low".
	Priority string `json:"priority"`
}

// Signal type sets used for classification.
var (
	migrationTypes = map[models.SignalType]bool{
		"frameworkMigration":    true,
		"migrationBlocker":     true,
		"deprecatedTestPattern": true,
		"dynamicTestGeneration": true,
		"customMatcherRisk":     true,
	}

	qualityTypes = map[models.SignalType]bool{
		"weakAssertion":         true,
		"mockHeavyTest":         true,
		"untestedExport":        true,
		"coverageThresholdBreak": true,
		"coverageBlindSpot":     true,
	}
)

// ComputeReadiness derives a migration readiness summary from the snapshot.
//
// Readiness levels:
//   - "high": few or no migration blockers
//   - "medium": some blockers but manageable
//   - "low": many blockers requiring significant effort
//
// The summary also cross-references quality signals with migration targets
// to surface areas where poor test quality amplifies migration risk.
func ComputeReadiness(snap *models.TestSuiteSnapshot) *ReadinessSummary {
	var blockers []models.Signal
	blockersByType := map[string]int{}

	for _, s := range snap.Signals {
		if !migrationTypes[s.Type] {
			continue
		}
		blockers = append(blockers, s)
		bt := "other"
		if m, ok := s.Metadata["blockerType"]; ok {
			if str, ok := m.(string); ok {
				bt = str
			}
		}
		blockersByType[bt]++
	}

	// Build representative examples (up to 5)
	var examples []BlockerExample
	limit := 5
	if len(blockers) < limit {
		limit = len(blockers)
	}
	for _, b := range blockers[:limit] {
		examples = append(examples, BlockerExample{
			Type:        string(b.Type),
			File:        b.Location.File,
			Explanation: b.Explanation,
		})
	}

	// Derive readiness level from blocker count relative to test files
	totalFiles := len(snap.TestFiles)
	readiness, explanation := deriveReadiness(len(blockers), totalFiles, blockersByType)

	// Cross-reference quality signals with migration targets
	qualityFactors := computeQualityFactors(snap, blockers)
	areas := computeAreaAssessments(snap)
	guidance := computeCoverageGuidance(areas, snap)

	return &ReadinessSummary{
		Frameworks:             snap.Frameworks,
		TotalBlockers:          len(blockers),
		BlockersByType:         blockersByType,
		RepresentativeBlockers: examples,
		ReadinessLevel:         readiness,
		Explanation:            explanation,
		QualityFactors:         qualityFactors,
		AreaAssessments:        areas,
		CoverageGuidance:       guidance,
	}
}

// computeQualityFactors identifies quality signals that co-occur with
// migration blocker files, amplifying migration risk.
func computeQualityFactors(snap *models.TestSuiteSnapshot, blockers []models.Signal) []QualityFactor {
	// Build set of files with migration blockers.
	blockerFiles := map[string]bool{}
	for _, b := range blockers {
		if b.Location.File != "" {
			blockerFiles[b.Location.File] = true
		}
	}

	if len(blockerFiles) == 0 {
		return nil
	}

	// Count quality signals that affect blocker files.
	type qualityHit struct {
		signalType string
		files      map[string]bool
	}
	hits := map[string]*qualityHit{}

	for _, s := range snap.Signals {
		if !qualityTypes[s.Type] {
			continue
		}
		file := s.Location.File
		if file == "" || !blockerFiles[file] {
			continue
		}
		st := string(s.Type)
		if hits[st] == nil {
			hits[st] = &qualityHit{signalType: st, files: map[string]bool{}}
		}
		hits[st].files[file] = true
	}

	var factors []QualityFactor
	for _, h := range hits {
		count := len(h.files)
		factors = append(factors, QualityFactor{
			SignalType:    h.signalType,
			AffectedFiles: count,
			Explanation:   qualityFactorExplanation(h.signalType, count),
		})
	}

	// Sort by affected file count descending for stable output.
	sort.Slice(factors, func(i, j int) bool {
		if factors[i].AffectedFiles != factors[j].AffectedFiles {
			return factors[i].AffectedFiles > factors[j].AffectedFiles
		}
		return factors[i].SignalType < factors[j].SignalType
	})

	return factors
}

func qualityFactorExplanation(signalType string, count int) string {
	switch signalType {
	case "weakAssertion":
		return fmt.Sprintf(
			"%d migration-target file(s) have weak assertions. Low assertion density means migrations may pass tests without verifying correctness.",
			count,
		)
	case "mockHeavyTest":
		return fmt.Sprintf(
			"%d migration-target file(s) are mock-heavy. Heavy mocking couples tests to implementation details that will change during migration.",
			count,
		)
	case "untestedExport":
		return fmt.Sprintf(
			"%d migration-target file(s) have untested exports. Untested public API makes it harder to verify migration correctness.",
			count,
		)
	case "coverageThresholdBreak":
		return fmt.Sprintf(
			"%d migration-target file(s) fall below coverage thresholds. Low coverage means less confidence in migration safety.",
			count,
		)
	case "coverageBlindSpot":
		return fmt.Sprintf(
			"%d migration-target file(s) have coverage blind spots. Uncovered code paths are invisible during migration validation.",
			count,
		)
	default:
		return fmt.Sprintf(
			"%d migration-target file(s) have quality issue '%s', adding risk to migration.",
			count, signalType,
		)
	}
}

// dirStats aggregates signal counts per directory.
type dirStats struct {
	dir               string
	migrationBlockers int
	qualityIssues     int
	testFileCount     int
}

// computeAreaAssessments classifies directories by migration safety.
func computeAreaAssessments(snap *models.TestSuiteSnapshot) []AreaAssessment {
	dirs := map[string]*dirStats{}

	// Count test files per directory.
	for _, tf := range snap.TestFiles {
		dir := filepath.Dir(tf.Path)
		if dirs[dir] == nil {
			dirs[dir] = &dirStats{dir: dir}
		}
		dirs[dir].testFileCount++
	}

	// Count signals per directory.
	for _, s := range snap.Signals {
		file := s.Location.File
		if file == "" {
			continue
		}
		dir := filepath.Dir(file)
		if dirs[dir] == nil {
			dirs[dir] = &dirStats{dir: dir}
		}
		if migrationTypes[s.Type] {
			dirs[dir].migrationBlockers++
		}
		if qualityTypes[s.Type] {
			dirs[dir].qualityIssues++
		}
	}

	// Only assess directories that have test files.
	var assessments []AreaAssessment
	for _, ds := range dirs {
		if ds.testFileCount == 0 {
			continue
		}

		classification, explanation := classifyArea(ds)
		assessments = append(assessments, AreaAssessment{
			Directory:         ds.dir,
			Classification:    classification,
			MigrationBlockers: ds.migrationBlockers,
			QualityIssues:     ds.qualityIssues,
			TestFileCount:     ds.testFileCount,
			Explanation:       explanation,
		})
	}

	// Sort: risky first, then caution, then safe. Within same class, by directory name.
	classOrder := map[string]int{"risky": 0, "caution": 1, "safe": 2}
	sort.Slice(assessments, func(i, j int) bool {
		ci, cj := classOrder[assessments[i].Classification], classOrder[assessments[j].Classification]
		if ci != cj {
			return ci < cj
		}
		return assessments[i].Directory < assessments[j].Directory
	})

	return assessments
}

func classifyArea(ds *dirStats) (string, string) {
	hasBlockers := ds.migrationBlockers > 0
	hasQuality := ds.qualityIssues > 0

	if hasBlockers && hasQuality {
		return "risky", fmt.Sprintf(
			"%d migration blocker(s) compounded by %d quality issue(s) across %d test file(s). Address quality issues before migrating.",
			ds.migrationBlockers, ds.qualityIssues, ds.testFileCount,
		)
	}

	if hasBlockers {
		return "caution", fmt.Sprintf(
			"%d migration blocker(s) in %d test file(s), but test quality is adequate. Migration requires blocker remediation.",
			ds.migrationBlockers, ds.testFileCount,
		)
	}

	if hasQuality {
		// Quality issues without migration blockers — still worth noting
		// but not a migration barrier.
		return "caution", fmt.Sprintf(
			"No migration blockers, but %d quality issue(s) in %d test file(s). Consider improving test quality before or during migration.",
			ds.qualityIssues, ds.testFileCount,
		)
	}

	return "safe", fmt.Sprintf(
		"%d test file(s) with no migration blockers and no quality issues. Safe to modernize.",
		ds.testFileCount,
	)
}

// computeCoverageGuidance identifies directories where additional test
// coverage would most reduce migration risk.
func computeCoverageGuidance(areas []AreaAssessment, snap *models.TestSuiteSnapshot) []CoverageGuidanceItem {
	// Build a set of directories with untested exports.
	untestedDirs := map[string]int{}
	for _, s := range snap.Signals {
		if s.Type == "untestedExport" && s.Location.File != "" {
			dir := filepath.Dir(s.Location.File)
			untestedDirs[dir]++
		}
	}

	var guidance []CoverageGuidanceItem

	for _, area := range areas {
		if area.Classification == "safe" {
			continue
		}

		priority := "medium"
		var reasons []string

		if area.MigrationBlockers > 0 && area.QualityIssues > 0 {
			priority = "high"
			reasons = append(reasons, "migration blockers combined with quality issues")
		}

		if count, ok := untestedDirs[area.Directory]; ok && count > 0 {
			priority = "high"
			reasons = append(reasons, fmt.Sprintf("%d untested export(s) in migration target", count))
		}

		if area.MigrationBlockers > 0 && area.QualityIssues == 0 {
			reasons = append(reasons, "migration blockers present — stronger tests would increase confidence in blocker remediation")
		}

		if area.QualityIssues > 0 && area.MigrationBlockers == 0 {
			reasons = append(reasons, "quality issues could become migration risks if framework changes are needed")
		}

		if len(reasons) == 0 {
			continue
		}

		guidance = append(guidance, CoverageGuidanceItem{
			Directory: area.Directory,
			Reason:    strings.Join(reasons, "; "),
			Priority:  priority,
		})
	}

	// Sort by priority (high first).
	priorityOrder := map[string]int{"high": 0, "medium": 1, "low": 2}
	sort.Slice(guidance, func(i, j int) bool {
		pi, pj := priorityOrder[guidance[i].Priority], priorityOrder[guidance[j].Priority]
		if pi != pj {
			return pi < pj
		}
		return guidance[i].Directory < guidance[j].Directory
	})

	return guidance
}

func deriveReadiness(blockerCount, totalFiles int, byType map[string]int) (string, string) {
	if totalFiles == 0 {
		return "unknown", "No test files detected."
	}

	ratio := float64(blockerCount) / float64(totalFiles)

	if blockerCount == 0 {
		return "high", "No migration blockers detected."
	}

	if ratio < 0.1 {
		topType := dominantType(byType)
		return "high", fmt.Sprintf(
			"Few migration blockers (%d across %d test files). Primary: %s.",
			blockerCount, totalFiles, topType,
		)
	}

	if ratio < 0.3 {
		topType := dominantType(byType)
		return "medium", fmt.Sprintf(
			"Some migration blockers (%d across %d test files). Focus on: %s.",
			blockerCount, totalFiles, topType,
		)
	}

	topType := dominantType(byType)
	return "low", fmt.Sprintf(
		"Many migration blockers (%d across %d test files). Major blocker: %s.",
		blockerCount, totalFiles, topType,
	)
}

func dominantType(byType map[string]int) string {
	type kv struct {
		key   string
		count int
	}
	var pairs []kv
	for k, v := range byType {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].count > pairs[j].count
	})
	if len(pairs) == 0 {
		return "unknown"
	}
	names := make([]string, 0, len(pairs))
	for _, p := range pairs {
		names = append(names, fmt.Sprintf("%s (%d)", p.key, p.count))
	}
	return strings.Join(names, ", ")
}
