package depgraph

import "sort"

// CoverageBand classifies how well a source file is covered by tests.
type CoverageBand string

const (
	CoverageBandHigh   CoverageBand = "High"
	CoverageBandMedium CoverageBand = "Medium"
	CoverageBandLow    CoverageBand = "Low"
)

// CoverageResult contains coverage analysis for all source files.
type CoverageResult struct {
	// Per-source coverage, sorted by TestCount ascending (least covered first).
	Sources []SourceCoverage `json:"sources"`

	// Total source files analyzed.
	SourceCount int `json:"sourceCount"`

	// Counts by band.
	BandCounts map[CoverageBand]int `json:"bandCounts"`
}

// SourceCoverage holds coverage metrics for a single source file.
type SourceCoverage struct {
	// Source file node ID.
	SourceID string `json:"sourceId"`

	// File path (without "file:" prefix).
	Path string `json:"path"`

	// Total unique test count (direct + indirect, deduplicated).
	TestCount int `json:"testCount"`

	// Tests that directly import this source file.
	DirectTests []string `json:"directTests"`

	// Tests that reach this source file indirectly (via helpers, fixtures,
	// or transitive imports).
	IndirectTests []string `json:"indirectTests"`

	// Coverage classification.
	Band CoverageBand `json:"band"`
}

// AnalyzeCoverage computes reverse coverage: for each source file, find all
// tests that cover it (directly or indirectly via the graph).
//
// Direct coverage: test file → (imports) → source file.
// Indirect coverage via 3 pathways:
//  1. Helper path: helper ← imports source, test ← uses helper
//  2. Fixture path: fixture ← imports source, test ← uses fixture
//  3. Transitive path: sourceB ← imports source, test ← imports sourceB
func AnalyzeCoverage(g *Graph) CoverageResult {
	sources := g.NodesByType(NodeSourceFile)

	// Build edge index: incoming edges grouped by (target, type).
	type edgeKey struct {
		target string
		etype  EdgeType
	}
	incoming := map[edgeKey][]*Edge{}
	for _, e := range g.Edges() {
		k := edgeKey{target: e.To, etype: e.Type}
		incoming[k] = append(incoming[k], e)
	}

	result := CoverageResult{
		Sources:    make([]SourceCoverage, 0, len(sources)),
		BandCounts: map[CoverageBand]int{},
	}

	for _, src := range sources {
		directTests := map[string]bool{}
		indirectTests := map[string]bool{}

		// Direct: test files that import this source.
		for _, e := range g.Incoming(src.ID) {
			if e.Type == EdgeImportsModule {
				fromNode := g.Node(e.From)
				if fromNode != nil && fromNode.Type == NodeTestFile {
					// Collect test IDs from this test file.
					for _, testID := range collectTestsInFile(g, e.From) {
						directTests[testID] = true
					}
				}
			}
		}

		// Indirect pathway 1: helpers that import this source.
		for _, e := range g.Incoming(src.ID) {
			if e.Type == EdgeHelperImportsSource || e.Type == EdgeImportsModule {
				fromNode := g.Node(e.From)
				if fromNode != nil && fromNode.Type == NodeHelper {
					// Find tests that use this helper.
					for _, he := range g.Incoming(e.From) {
						if he.Type == EdgeTestUsesHelper {
							for _, testID := range collectTestsInFile(g, he.From) {
								if !directTests[testID] {
									indirectTests[testID] = true
								}
							}
						}
					}
				}
			}
		}

		// Indirect pathway 2: fixtures that import this source.
		for _, e := range g.Incoming(src.ID) {
			if e.Type == EdgeFixtureImportsSource || e.Type == EdgeImportsModule {
				fromNode := g.Node(e.From)
				if fromNode != nil && fromNode.Type == NodeFixture {
					for _, fe := range g.Incoming(e.From) {
						if fe.Type == EdgeTestUsesFixture {
							for _, testID := range collectTestsInFile(g, fe.From) {
								if !directTests[testID] {
									indirectTests[testID] = true
								}
							}
						}
					}
				}
			}
		}

		// Indirect pathway 3: other sources that import this source.
		for _, e := range g.Incoming(src.ID) {
			if e.Type == EdgeSourceImportsSource {
				// Find tests that import the intermediate source.
				for _, ie := range g.Incoming(e.From) {
					if ie.Type == EdgeImportsModule {
						fromNode := g.Node(ie.From)
						if fromNode != nil && fromNode.Type == NodeTestFile {
							for _, testID := range collectTestsInFile(g, ie.From) {
								if !directTests[testID] {
									indirectTests[testID] = true
								}
							}
						}
					}
				}
			}
		}

		totalCount := len(directTests) + len(indirectTests)
		band := classifyBand(totalCount)

		directList := sortedKeys(directTests)
		indirectList := sortedKeys(indirectTests)

		result.Sources = append(result.Sources, SourceCoverage{
			SourceID:      src.ID,
			Path:          src.Path,
			TestCount:     totalCount,
			DirectTests:   directList,
			IndirectTests: indirectList,
			Band:          band,
		})
		result.BandCounts[band]++
	}

	result.SourceCount = len(result.Sources)

	// Sort by test count ascending (least covered first).
	sort.Slice(result.Sources, func(i, j int) bool {
		if result.Sources[i].TestCount != result.Sources[j].TestCount {
			return result.Sources[i].TestCount < result.Sources[j].TestCount
		}
		return result.Sources[i].SourceID < result.Sources[j].SourceID
	})

	return result
}

// collectTestsInFile returns the IDs of all Test nodes defined in a given
// file (identified by file node ID).
func collectTestsInFile(g *Graph, fileID string) []string {
	// Tests have edges test→file via TestDefinedInFile.
	// Look for incoming TestDefinedInFile edges to this file.
	var tests []string
	for _, e := range g.Incoming(fileID) {
		if e.Type == EdgeTestDefinedInFile {
			tests = append(tests, e.From)
		}
	}
	return tests
}

// classifyBand assigns a coverage band based on test count.
func classifyBand(testCount int) CoverageBand {
	if testCount >= 3 {
		return CoverageBandHigh
	}
	if testCount >= 1 {
		return CoverageBandMedium
	}
	return CoverageBandLow
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
