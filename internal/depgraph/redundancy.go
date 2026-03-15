package depgraph

import (
	"fmt"
	"sort"
	"strings"
)

// RedundancyResult contains behavior-aware redundancy analysis.
type RedundancyResult struct {
	// Clusters are groups of tests exercising overlapping behavior surfaces.
	Clusters []RedundancyCluster `json:"clusters"`

	// TestsAnalyzed is the total number of tests evaluated.
	TestsAnalyzed int `json:"testsAnalyzed"`

	// RedundantTestCount is the total number of tests in redundancy clusters.
	RedundantTestCount int `json:"redundantTestCount"`

	// CrossFrameworkOverlaps counts clusters where tests span multiple frameworks.
	CrossFrameworkOverlaps int `json:"crossFrameworkOverlaps"`

	// Skipped indicates the analysis was skipped for scale.
	Skipped    bool   `json:"skipped,omitempty"`
	SkipReason string `json:"skipReason,omitempty"`
}

// RedundancyCluster is a group of tests with overlapping behavior coverage.
type RedundancyCluster struct {
	// ID is a deterministic cluster identifier.
	ID string `json:"id"`

	// Tests are the test node IDs in this cluster.
	Tests []string `json:"tests"`

	// SharedSurfaces are the behavior/code surface IDs that all cluster
	// members exercise.
	SharedSurfaces []string `json:"sharedSurfaces"`

	// SurfaceNames maps surface IDs to human-readable names.
	SurfaceNames map[string]string `json:"surfaceNames,omitempty"`

	// Confidence is the strength of the redundancy signal (0.0–1.0).
	Confidence float64 `json:"confidence"`

	// OverlapKind classifies the type of overlap.
	OverlapKind OverlapKind `json:"overlapKind"`

	// Rationale explains why this cluster was flagged.
	Rationale string `json:"rationale"`

	// Frameworks lists the distinct frameworks present in the cluster.
	Frameworks []string `json:"frameworks,omitempty"`
}

// OverlapKind classifies the nature of test overlap.
type OverlapKind string

const (
	// OverlapWasteful means tests in the same framework exercise the same
	// behavior surfaces with similar structure — likely candidates for
	// consolidation.
	OverlapWasteful OverlapKind = "wasteful"

	// OverlapCrossFramework means tests from different frameworks exercise
	// the same behavior surfaces. This may be intentional (migration
	// period) or may indicate consolidation opportunity.
	OverlapCrossFramework OverlapKind = "cross_framework"

	// OverlapCrossLevel means tests at different levels (e.g., unit and
	// e2e) exercise the same surfaces. This is typically healthy defense
	// in depth.
	OverlapCrossLevel OverlapKind = "cross_level"
)

// maxRedundancyTests caps the analysis to avoid expensive graph traversals.
const maxRedundancyTests = 25000

// minBehaviorOverlap is the minimum fraction of shared surfaces required
// to consider two tests as exercising overlapping behavior.
const minBehaviorOverlap = 0.5

// AnalyzeRedundancy detects behavior-aware redundancy in the test suite.
//
// Unlike DetectDuplicates (which uses structural fingerprinting), this
// analysis reasons about what behaviors each test exercises:
//  1. Trace test → imported source files → code surfaces → behavior surfaces
//  2. Group tests by shared behavior/code surface coverage
//  3. Classify overlap as wasteful, cross-framework, or cross-level
//  4. Emit explainable clusters with rationale
func AnalyzeRedundancy(g *Graph) RedundancyResult {
	tests := g.NodesByType(NodeTest)
	if len(tests) == 0 {
		return RedundancyResult{Clusters: []RedundancyCluster{}}
	}
	if len(tests) > maxRedundancyTests {
		return RedundancyResult{
			Clusters:      []RedundancyCluster{},
			TestsAnalyzed: len(tests),
			Skipped:       true,
			SkipReason: fmt.Sprintf(
				"redundancy analysis skipped for %d tests (limit %d)",
				len(tests), maxRedundancyTests),
		}
	}

	// Step 1: Build test → exercised surfaces mapping.
	testSurfaces := buildTestSurfaceMap(g, tests)

	// Step 2: Build surface → tests reverse index for candidate generation.
	surfaceToTests := buildSurfaceToTestsIndex(testSurfaces)

	// Step 3: Generate candidate pairs via shared surface blocking.
	pairs := generateRedundancyCandidates(testSurfaces, surfaceToTests)

	// Step 4: Score pairs and build clusters.
	clusters := scoreAndClusterRedundancy(g, tests, testSurfaces, pairs)

	// Count cross-framework overlaps.
	crossFW := 0
	for _, c := range clusters {
		if c.OverlapKind == OverlapCrossFramework {
			crossFW++
		}
	}

	totalRedundant := 0
	for _, c := range clusters {
		totalRedundant += len(c.Tests)
	}

	return RedundancyResult{
		Clusters:               clusters,
		TestsAnalyzed:          len(tests),
		RedundantTestCount:     totalRedundant,
		CrossFrameworkOverlaps: crossFW,
	}
}

// testSurfaceInfo captures the surfaces a test exercises and its metadata.
type testSurfaceInfo struct {
	nodeID    string
	framework string
	testType  string // from metadata: unit, integration, e2e
	surfaces  map[string]bool
}

// buildTestSurfaceMap traces each test through the import graph to discover
// which code surfaces and behavior surfaces it exercises.
func buildTestSurfaceMap(g *Graph, tests []*Node) map[string]*testSurfaceInfo {
	// Pre-build indexes for efficient lookup.
	// source file → code surfaces contained in it
	srcToCodeSurfaces := map[string][]string{}
	for _, e := range g.Edges() {
		if e.Type == EdgeBelongsToPackage {
			fromNode := g.Node(e.From)
			if fromNode != nil && fromNode.Type == NodeCodeSurface {
				srcToCodeSurfaces[e.To] = append(srcToCodeSurfaces[e.To], e.From)
			}
		}
	}

	// code surface → behavior surfaces it belongs to
	csToBehavior := map[string][]string{}
	for _, e := range g.Edges() {
		if e.Type == EdgeBehaviorDerivedFrom {
			fromNode := g.Node(e.From)
			if fromNode != nil && fromNode.Type == NodeBehaviorSurface {
				csToBehavior[e.To] = append(csToBehavior[e.To], e.From)
			}
		}
	}

	// test file → imported source files
	testFileImports := map[string][]string{}
	for _, e := range g.Edges() {
		if e.Type == EdgeImportsModule {
			testFileImports[e.From] = append(testFileImports[e.From], e.To)
		}
	}

	result := map[string]*testSurfaceInfo{}
	for _, test := range tests {
		info := &testSurfaceInfo{
			nodeID:    test.ID,
			framework: test.Framework,
			surfaces:  map[string]bool{},
		}
		if test.Metadata != nil {
			info.testType = test.Metadata["testType"]
		}

		fileID := "file:" + test.Path

		// Trace: test file → imported sources → code surfaces → behavior surfaces
		for _, srcID := range testFileImports[fileID] {
			for _, csID := range srcToCodeSurfaces[srcID] {
				info.surfaces[csID] = true
				for _, bsID := range csToBehavior[csID] {
					info.surfaces[bsID] = true
				}
			}
		}

		// Only track tests that exercise at least one surface.
		if len(info.surfaces) > 0 {
			result[test.ID] = info
		}
	}

	return result
}

// buildSurfaceToTestsIndex creates a reverse index from surface → tests.
func buildSurfaceToTestsIndex(testSurfaces map[string]*testSurfaceInfo) map[string][]string {
	index := map[string][]string{}
	for testID, info := range testSurfaces {
		for surfaceID := range info.surfaces {
			index[surfaceID] = append(index[surfaceID], testID)
		}
	}
	return index
}

// generateRedundancyCandidates produces test pairs that share at least one
// behavior or code surface. Uses surface-based blocking to avoid O(n²).
func generateRedundancyCandidates(
	testSurfaces map[string]*testSurfaceInfo,
	surfaceToTests map[string][]string,
) map[[2]string]bool {
	pairs := map[[2]string]bool{}

	for _, testIDs := range surfaceToTests {
		if len(testIDs) < 2 || len(testIDs) > maxBlockSize {
			continue
		}
		sort.Strings(testIDs)
		for a := 0; a < len(testIDs); a++ {
			for b := a + 1; b < len(testIDs); b++ {
				pair := [2]string{testIDs[a], testIDs[b]}
				pairs[pair] = true
			}
		}
	}

	return pairs
}

// scoreAndClusterRedundancy evaluates candidate pairs and forms clusters.
func scoreAndClusterRedundancy(
	g *Graph,
	tests []*Node,
	testSurfaces map[string]*testSurfaceInfo,
	pairs map[[2]string]bool,
) []RedundancyCluster {
	// Index test IDs for union-find.
	idToIdx := map[string]int{}
	idxToID := map[int]string{}
	for i, t := range tests {
		idToIdx[t.ID] = i
		idxToID[i] = t.ID
	}

	uf := newUnionFind(len(tests))

	type pairResult struct {
		overlap    float64
		sharedIDs  []string
	}
	pairData := map[[2]string]pairResult{}

	for pair := range pairs {
		infoA := testSurfaces[pair[0]]
		infoB := testSurfaces[pair[1]]
		if infoA == nil || infoB == nil {
			continue
		}

		// Compute surface overlap (Jaccard).
		overlap := jaccardSets(infoA.surfaces, infoB.surfaces)
		if overlap < minBehaviorOverlap {
			continue
		}

		// Collect shared surfaces.
		var shared []string
		for s := range infoA.surfaces {
			if infoB.surfaces[s] {
				shared = append(shared, s)
			}
		}
		sort.Strings(shared)

		idxA, okA := idToIdx[pair[0]]
		idxB, okB := idToIdx[pair[1]]
		if okA && okB {
			uf.union(idxA, idxB)
			pairData[pair] = pairResult{overlap: overlap, sharedIDs: shared}
		}
	}

	// Build clusters from union-find.
	clusterMap := map[int][]int{}
	for i := range tests {
		root := uf.find(i)
		clusterMap[root] = append(clusterMap[root], i)
	}

	var clusters []RedundancyCluster
	for _, members := range clusterMap {
		if len(members) < 2 {
			continue
		}

		// Collect test IDs and frameworks.
		testIDs := make([]string, len(members))
		fwSet := map[string]bool{}
		typeSet := map[string]bool{}
		for k, m := range members {
			testIDs[k] = idxToID[m]
			info := testSurfaces[idxToID[m]]
			if info != nil {
				if info.framework != "" {
					fwSet[info.framework] = true
				}
				if info.testType != "" {
					typeSet[info.testType] = true
				}
			}
		}
		sort.Strings(testIDs)

		// Find surfaces shared by ALL members (intersection).
		sharedSurfaces := intersectSurfaces(testSurfaces, testIDs)
		if len(sharedSurfaces) == 0 {
			continue
		}

		// Compute average pairwise overlap.
		var totalOverlap float64
		pairCount := 0
		for a := 0; a < len(testIDs); a++ {
			for b := a + 1; b < len(testIDs); b++ {
				pair := [2]string{testIDs[a], testIDs[b]}
				if testIDs[a] > testIDs[b] {
					pair = [2]string{testIDs[b], testIDs[a]}
				}
				if pd, ok := pairData[pair]; ok {
					totalOverlap += pd.overlap
					pairCount++
				}
			}
		}
		avgOverlap := 0.0
		if pairCount > 0 {
			avgOverlap = totalOverlap / float64(pairCount)
		}

		// Resolve surface names.
		surfaceNames := map[string]string{}
		for _, sid := range sharedSurfaces {
			if n := g.Node(sid); n != nil {
				surfaceNames[sid] = nodeLabelForRedundancy(n)
			}
		}

		// Classify overlap kind.
		frameworks := sortedStringKeys(fwSet)
		kind, rationale := classifyOverlap(frameworks, typeSet, avgOverlap, len(sharedSurfaces))

		// Confidence from overlap strength and surface count.
		confidence := redundancyConfidence(avgOverlap, len(sharedSurfaces), kind)

		clusterID := fmt.Sprintf("redundancy:%s", strings.Join(sharedSurfaces[:minInt(3, len(sharedSurfaces))], "+"))

		clusters = append(clusters, RedundancyCluster{
			ID:             clusterID,
			Tests:          testIDs,
			SharedSurfaces: sharedSurfaces,
			SurfaceNames:   surfaceNames,
			Confidence:     confidence,
			OverlapKind:    kind,
			Rationale:      rationale,
			Frameworks:     frameworks,
		})
	}

	// Sort: wasteful first, then by size desc, then confidence desc, then ID.
	sort.Slice(clusters, func(i, j int) bool {
		ki := overlapKindOrder(clusters[i].OverlapKind)
		kj := overlapKindOrder(clusters[j].OverlapKind)
		if ki != kj {
			return ki < kj
		}
		if len(clusters[i].Tests) != len(clusters[j].Tests) {
			return len(clusters[i].Tests) > len(clusters[j].Tests)
		}
		if clusters[i].Confidence != clusters[j].Confidence {
			return clusters[i].Confidence > clusters[j].Confidence
		}
		return clusters[i].ID < clusters[j].ID
	})

	return clusters
}

// intersectSurfaces returns surfaces shared by ALL tests in the list.
func intersectSurfaces(testSurfaces map[string]*testSurfaceInfo, testIDs []string) []string {
	if len(testIDs) == 0 {
		return nil
	}

	first := testSurfaces[testIDs[0]]
	if first == nil {
		return nil
	}

	shared := map[string]bool{}
	for s := range first.surfaces {
		shared[s] = true
	}

	for _, tid := range testIDs[1:] {
		info := testSurfaces[tid]
		if info == nil {
			return nil
		}
		for s := range shared {
			if !info.surfaces[s] {
				delete(shared, s)
			}
		}
	}

	result := make([]string, 0, len(shared))
	for s := range shared {
		result = append(result, s)
	}
	sort.Strings(result)
	return result
}

// classifyOverlap determines whether overlap is wasteful, cross-framework,
// or cross-level (defense in depth).
func classifyOverlap(frameworks []string, testTypes map[string]bool, overlap float64, surfaceCount int) (OverlapKind, string) {
	multiFramework := len(frameworks) > 1
	multiLevel := len(testTypes) > 1

	if multiLevel {
		return OverlapCrossLevel, fmt.Sprintf(
			"Tests at different levels (%s) cover %d shared surfaces — this is defense in depth, typically healthy.",
			strings.Join(sortedStringKeys(testTypes), ", "), surfaceCount)
	}

	if multiFramework {
		return OverlapCrossFramework, fmt.Sprintf(
			"Tests in %s exercise %d identical surfaces (%.0f%% overlap). If migrating, remove old-framework tests after migration completes.",
			strings.Join(frameworks, " and "), surfaceCount, overlap*100)
	}

	return OverlapWasteful, fmt.Sprintf(
		"Tests in the same framework exercise %d identical behavior surfaces (%.0f%% overlap). Consolidation would reduce CI cost without losing coverage.",
		surfaceCount, overlap*100)
}

// redundancyConfidence computes the confidence score for a redundancy cluster.
func redundancyConfidence(avgOverlap float64, surfaceCount int, kind OverlapKind) float64 {
	// Base from overlap strength.
	base := avgOverlap * 0.7

	// Boost from surface count — more shared surfaces = stronger signal.
	surfaceBoost := 0.0
	if surfaceCount >= 3 {
		surfaceBoost = 0.15
	} else if surfaceCount >= 1 {
		surfaceBoost = 0.10
	}

	// Kind adjustment — wasteful overlap gets higher confidence.
	kindBoost := 0.0
	switch kind {
	case OverlapWasteful:
		kindBoost = 0.15
	case OverlapCrossFramework:
		kindBoost = 0.10
	case OverlapCrossLevel:
		kindBoost = 0.0 // cross-level is healthy, lower confidence of "redundancy"
	}

	confidence := base + surfaceBoost + kindBoost
	if confidence > 1.0 {
		confidence = 1.0
	}
	return confidence
}

// overlapKindOrder returns a sort order for overlap kinds.
// Wasteful first (most actionable), cross-level last (healthy overlap).
func overlapKindOrder(k OverlapKind) int {
	switch k {
	case OverlapWasteful:
		return 0
	case OverlapCrossFramework:
		return 1
	case OverlapCrossLevel:
		return 2
	default:
		return 3
	}
}

// nodeLabelForRedundancy returns a display label for a surface node.
func nodeLabelForRedundancy(n *Node) string {
	if n.Name != "" {
		return n.Name
	}
	if n.Path != "" {
		return n.Path
	}
	return n.ID
}

// sortedStringKeys returns the sorted keys of a string-bool map.
func sortedStringKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
