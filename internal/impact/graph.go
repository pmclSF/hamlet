package impact

import (
	"sort"

	"github.com/pmclSF/hamlet/internal/models"
)

// EdgeKind describes the type of impact relationship.
type EdgeKind string

const (
	EdgeExactCoverage      EdgeKind = "exact_coverage"      // per-test coverage lineage
	EdgeBucketCoverage     EdgeKind = "bucket_coverage"     // bucket/run-level coverage
	EdgeStructuralLink     EdgeKind = "structural_link"     // import/export relationship
	EdgeDirectoryProximity EdgeKind = "directory_proximity" // same directory tree
	EdgeNameConvention     EdgeKind = "name_convention"     // naming heuristic match
)

// ImpactEdge represents a single relationship in the impact graph.
type ImpactEdge struct {
	// SourceID is the code unit or test file ID on one end.
	SourceID string `json:"sourceId"`

	// TargetID is the code unit or test file ID on the other end.
	TargetID string `json:"targetId"`

	// Kind describes how this relationship was established.
	Kind EdgeKind `json:"kind"`

	// Confidence is the mapping quality (exact, inferred, weak).
	Confidence Confidence `json:"confidence"`

	// Provenance describes the data source for this edge.
	Provenance string `json:"provenance"`

	// CoverageType is the test type if known (unit, integration, e2e).
	CoverageType string `json:"coverageType,omitempty"`
}

// ImpactGraph is a bidirectional map connecting code units to tests.
type ImpactGraph struct {
	// Edges contains all impact relationships.
	Edges []ImpactEdge `json:"edges"`

	// UnitToTests maps code unit IDs to test file paths.
	UnitToTests map[string][]string `json:"-"`

	// TestToUnits maps test file paths to code unit IDs.
	TestToUnits map[string][]string `json:"-"`

	// EdgeIndex maps "sourceID->targetID" to the edge for lookup.
	EdgeIndex map[string]*ImpactEdge `json:"-"`

	// Stats summarizes graph construction quality.
	Stats GraphStats `json:"stats"`
}

// GraphStats captures aggregate quality metrics for the impact graph.
type GraphStats struct {
	TotalEdges     int `json:"totalEdges"`
	ExactEdges     int `json:"exactEdges"`
	InferredEdges  int `json:"inferredEdges"`
	WeakEdges      int `json:"weakEdges"`
	ConnectedUnits int `json:"connectedUnits"`
	IsolatedUnits  int `json:"isolatedUnits"`
	ConnectedTests int `json:"connectedTests"`
}

// BuildImpactGraph constructs an impact graph from snapshot data.
// It uses multiple strategies in priority order:
//  1. Per-test coverage lineage (exact)
//  2. LinkedCodeUnits on test files (bucket-level)
//  3. Structural/naming heuristics (inferred)
func BuildImpactGraph(snap *models.TestSuiteSnapshot) *ImpactGraph {
	g := &ImpactGraph{
		UnitToTests: make(map[string][]string),
		TestToUnits: make(map[string][]string),
		EdgeIndex:   make(map[string]*ImpactEdge),
	}

	if snap == nil {
		return g
	}

	// Build framework type index for coverage type classification.
	fwTypes := map[string]string{}
	for _, fw := range snap.Frameworks {
		switch fw.Type {
		case models.FrameworkTypeUnit:
			fwTypes[fw.Name] = "unit"
		case models.FrameworkTypeE2E:
			fwTypes[fw.Name] = "e2e"
		default:
			fwTypes[fw.Name] = "integration"
		}
	}

	// Build code unit ID set for validation.
	unitIDs := map[string]bool{}
	for _, cu := range snap.CodeUnits {
		unitIDs[cu.Path+":"+cu.Name] = true
	}

	// Strategy 1: LinkedCodeUnits on test files (bucket-level coverage).
	for _, tf := range snap.TestFiles {
		covType := fwTypes[tf.Framework]
		for _, linked := range tf.LinkedCodeUnits {
			kind := EdgeBucketCoverage
			conf := ConfidenceInferred
			prov := "linked_code_units"

			// If the linked value matches a known unit ID exactly, upgrade confidence.
			if unitIDs[linked] {
				conf = ConfidenceExact
				kind = EdgeExactCoverage
				prov = "exact_unit_link"
			}

			g.addEdge(linked, tf.Path, kind, conf, prov, covType)
		}
	}

	// Strategy 2: Name-convention matching.
	// Match test files to code units by naming patterns.
	unitsByName := map[string][]models.CodeUnit{}
	for _, cu := range snap.CodeUnits {
		unitsByName[cu.Name] = append(unitsByName[cu.Name], cu)
	}

	for _, tf := range snap.TestFiles {
		// Only add name-convention edges for tests not already connected.
		if len(g.TestToUnits[tf.Path]) > 0 {
			continue
		}
		covType := fwTypes[tf.Framework]
		// Try to match test file base name to code unit name.
		baseName := extractTestSubject(tf.Path)
		if baseName == "" {
			continue
		}
		if units, ok := unitsByName[baseName]; ok {
			for _, cu := range units {
				unitID := cu.Path + ":" + cu.Name
				g.addEdge(unitID, tf.Path, EdgeNameConvention, ConfidenceWeak, "name_convention", covType)
			}
		}
	}

	// Compute stats.
	g.computeStats(snap)

	// Sort edges for determinism.
	sort.Slice(g.Edges, func(i, j int) bool {
		if g.Edges[i].SourceID != g.Edges[j].SourceID {
			return g.Edges[i].SourceID < g.Edges[j].SourceID
		}
		return g.Edges[i].TargetID < g.Edges[j].TargetID
	})

	return g
}

// addEdge adds an edge to the graph, avoiding duplicates.
func (g *ImpactGraph) addEdge(unitID, testPath string, kind EdgeKind, conf Confidence, prov, covType string) {
	key := unitID + "->" + testPath
	if existing, ok := g.EdgeIndex[key]; ok {
		// Upgrade confidence if new edge is stronger.
		if confidenceOrder(conf) < confidenceOrder(existing.Confidence) {
			existing.Kind = kind
			existing.Confidence = conf
			existing.Provenance = prov
		}
		return
	}

	edge := ImpactEdge{
		SourceID:     unitID,
		TargetID:     testPath,
		Kind:         kind,
		Confidence:   conf,
		Provenance:   prov,
		CoverageType: covType,
	}
	g.Edges = append(g.Edges, edge)
	g.EdgeIndex[key] = &g.Edges[len(g.Edges)-1]

	g.UnitToTests[unitID] = append(g.UnitToTests[unitID], testPath)
	g.TestToUnits[testPath] = append(g.TestToUnits[testPath], unitID)
}

// TestsForUnit returns the test paths covering a code unit, sorted by edge confidence.
func (g *ImpactGraph) TestsForUnit(unitID string) []string {
	tests := g.UnitToTests[unitID]
	if len(tests) == 0 {
		return nil
	}

	// Sort by confidence of the edge (exact first).
	type testWithConf struct {
		path string
		conf Confidence
	}
	var sorted []testWithConf
	for _, tp := range tests {
		key := unitID + "->" + tp
		conf := ConfidenceWeak
		if edge, ok := g.EdgeIndex[key]; ok {
			conf = edge.Confidence
		}
		sorted = append(sorted, testWithConf{tp, conf})
	}
	sort.Slice(sorted, func(i, j int) bool {
		ci, cj := confidenceOrder(sorted[i].conf), confidenceOrder(sorted[j].conf)
		if ci != cj {
			return ci < cj
		}
		return sorted[i].path < sorted[j].path
	})

	result := make([]string, len(sorted))
	for i, s := range sorted {
		result[i] = s.path
	}
	return result
}

// UnitsForTest returns the code unit IDs covered by a test.
func (g *ImpactGraph) UnitsForTest(testPath string) []string {
	units := g.TestToUnits[testPath]
	if len(units) == 0 {
		return nil
	}
	result := make([]string, len(units))
	copy(result, units)
	sort.Strings(result)
	return result
}

// EdgesForUnit returns all edges connecting to a code unit.
func (g *ImpactGraph) EdgesForUnit(unitID string) []ImpactEdge {
	var edges []ImpactEdge
	for _, e := range g.Edges {
		if e.SourceID == unitID {
			edges = append(edges, e)
		}
	}
	return edges
}

// EdgeBetween returns the edge between a unit and test, if it exists.
func (g *ImpactGraph) EdgeBetween(unitID, testPath string) *ImpactEdge {
	key := unitID + "->" + testPath
	return g.EdgeIndex[key]
}

// computeStats aggregates graph quality metrics.
func (g *ImpactGraph) computeStats(snap *models.TestSuiteSnapshot) {
	g.Stats.TotalEdges = len(g.Edges)

	for _, e := range g.Edges {
		switch e.Confidence {
		case ConfidenceExact:
			g.Stats.ExactEdges++
		case ConfidenceInferred:
			g.Stats.InferredEdges++
		case ConfidenceWeak:
			g.Stats.WeakEdges++
		}
	}

	g.Stats.ConnectedUnits = len(g.UnitToTests)
	g.Stats.ConnectedTests = len(g.TestToUnits)

	// Count isolated units (units with no test edges).
	totalUnits := len(snap.CodeUnits)
	g.Stats.IsolatedUnits = totalUnits - g.Stats.ConnectedUnits
	if g.Stats.IsolatedUnits < 0 {
		g.Stats.IsolatedUnits = 0
	}
}

// extractTestSubject extracts the likely subject name from a test file path.
// E.g., "src/__tests__/AuthService.test.js" -> "AuthService"
func extractTestSubject(path string) string {
	// Find the base file name.
	base := path
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			base = path[i+1:]
			break
		}
	}

	// Remove test file suffixes.
	suffixes := []string{
		".test.js", ".test.ts", ".test.tsx", ".test.jsx",
		".spec.js", ".spec.ts", ".spec.tsx", ".spec.jsx",
		"_test.go", "_test.py",
		".test.java", ".spec.java",
	}
	for _, s := range suffixes {
		if len(base) > len(s) && base[len(base)-len(s):] == s {
			return base[:len(base)-len(s)]
		}
	}

	return ""
}
