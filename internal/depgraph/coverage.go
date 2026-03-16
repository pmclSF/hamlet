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

	// Tests that reach this source file indirectly (via transitive imports).
	IndirectTests []string `json:"indirectTests"`

	// Coverage classification.
	Band CoverageBand `json:"band"`
}

// AnalyzeCoverage computes reverse coverage: for each source file, find all
// tests that cover it (directly or indirectly via the graph).
//
// Direct coverage: test file → (imports) → source file.
// Indirect coverage: transitive path: sourceB ← imports source, test ← imports sourceB
func AnalyzeCoverage(g *Graph) CoverageResult {
	sources := g.NodesByType(NodeSourceFile)

	result := CoverageResult{
		Sources:    make([]SourceCoverage, 0, len(sources)),
		BandCounts: map[CoverageBand]int{},
	}

	for _, src := range sources {
		directTests := map[string]bool{}
		indirectTests := map[string]bool{}

		// Single pass over incoming edges: handle direct imports and
		// transitive source imports in one iteration.
		for _, e := range g.Incoming(src.ID) {
			switch e.Type {
			case EdgeImportsModule:
				// Direct: test file imports this source.
				fromNode := g.Node(e.From)
				if fromNode != nil && fromNode.Type == NodeTestFile {
					for _, testID := range collectTestsInFile(g, e.From) {
						directTests[testID] = true
					}
				}
			case EdgeSourceImportsSource:
				// Indirect: another source imports this source.
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
