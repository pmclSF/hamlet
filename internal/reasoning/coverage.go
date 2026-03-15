package reasoning

import (
	"sort"

	"github.com/pmclSF/terrain/internal/depgraph"
)

// CoverageBand classifies how well a node is covered.
type CoverageBand string

const (
	CoverageHigh   CoverageBand = "High"
	CoverageMedium CoverageBand = "Medium"
	CoverageLow    CoverageBand = "Low"
)

// CoverageSummary holds aggregated coverage for a single node.
type CoverageSummary struct {
	// NodeID is the covered node.
	NodeID string

	// DirectTests lists validation nodes that directly cover this node.
	DirectTests []string

	// IndirectTests lists validation nodes that reach this node
	// through intermediate edges (helpers, fixtures, transitive imports).
	IndirectTests []string

	// TotalCount is len(DirectTests) + len(IndirectTests).
	TotalCount int

	// Band is the coverage classification.
	Band CoverageBand
}

// ClassifyCoverageBand assigns a coverage band based on test count.
// Matches the thresholds used in depgraph.AnalyzeCoverage.
func ClassifyCoverageBand(testCount int) CoverageBand {
	if testCount >= 3 {
		return CoverageHigh
	}
	if testCount >= 1 {
		return CoverageMedium
	}
	return CoverageLow
}

// CoverageConfig controls coverage aggregation behavior.
type CoverageConfig struct {
	// HighThreshold is the minimum test count for "High" band. Default: 3.
	HighThreshold int

	// MediumThreshold is the minimum test count for "Medium" band. Default: 1.
	MediumThreshold int
}

// DefaultCoverageConfig returns standard coverage thresholds.
func DefaultCoverageConfig() CoverageConfig {
	return CoverageConfig{
		HighThreshold:   3,
		MediumThreshold: 1,
	}
}

// ClassifyCoverageBandWithConfig assigns a coverage band using custom thresholds.
func ClassifyCoverageBandWithConfig(testCount int, cfg CoverageConfig) CoverageBand {
	if testCount >= cfg.HighThreshold {
		return CoverageHigh
	}
	if testCount >= cfg.MediumThreshold {
		return CoverageMedium
	}
	return CoverageLow
}

// CollectCovering finds all validation nodes that cover the given target node
// by performing a reverse traversal through the graph.
//
// It identifies direct coverage (validation → target) and indirect coverage
// via helpers, fixtures, and transitive source imports.
func CollectCovering(g *depgraph.Graph, targetID string) CoverageSummary {
	if g == nil || g.Node(targetID) == nil {
		return CoverageSummary{NodeID: targetID, Band: CoverageLow}
	}

	direct := map[string]bool{}
	indirect := map[string]bool{}

	// Direct: validation nodes (test files) that import this target.
	for _, e := range g.Incoming(targetID) {
		if e.Type == depgraph.EdgeImportsModule {
			fromNode := g.Node(e.From)
			if fromNode != nil && fromNode.Type == depgraph.NodeTestFile {
				for _, testID := range collectTestsInFile(g, e.From) {
					direct[testID] = true
				}
			}
		}
	}

	// Indirect pathway 1: helpers that import this target.
	for _, e := range g.Incoming(targetID) {
		if e.Type == depgraph.EdgeHelperImportsSource || e.Type == depgraph.EdgeImportsModule {
			fromNode := g.Node(e.From)
			if fromNode != nil && fromNode.Type == depgraph.NodeHelper {
				for _, he := range g.Incoming(e.From) {
					if he.Type == depgraph.EdgeTestUsesHelper {
						for _, testID := range collectTestsInFile(g, he.From) {
							if !direct[testID] {
								indirect[testID] = true
							}
						}
					}
				}
			}
		}
	}

	// Indirect pathway 2: fixtures that import this target.
	for _, e := range g.Incoming(targetID) {
		if e.Type == depgraph.EdgeFixtureImportsSource || e.Type == depgraph.EdgeImportsModule {
			fromNode := g.Node(e.From)
			if fromNode != nil && fromNode.Type == depgraph.NodeFixture {
				for _, fe := range g.Incoming(e.From) {
					if fe.Type == depgraph.EdgeTestUsesFixture {
						for _, testID := range collectTestsInFile(g, fe.From) {
							if !direct[testID] {
								indirect[testID] = true
							}
						}
					}
				}
			}
		}
	}

	// Indirect pathway 3: transitive source imports.
	for _, e := range g.Incoming(targetID) {
		if e.Type == depgraph.EdgeSourceImportsSource {
			for _, ie := range g.Incoming(e.From) {
				if ie.Type == depgraph.EdgeImportsModule {
					fromNode := g.Node(ie.From)
					if fromNode != nil && fromNode.Type == depgraph.NodeTestFile {
						for _, testID := range collectTestsInFile(g, ie.From) {
							if !direct[testID] {
								indirect[testID] = true
							}
						}
					}
				}
			}
		}
	}

	directList := sortedStringKeys(direct)
	indirectList := sortedStringKeys(indirect)
	total := len(directList) + len(indirectList)

	return CoverageSummary{
		NodeID:        targetID,
		DirectTests:   directList,
		IndirectTests: indirectList,
		TotalCount:    total,
		Band:          ClassifyCoverageBand(total),
	}
}

// collectTestsInFile returns test node IDs defined in the given file node.
func collectTestsInFile(g *depgraph.Graph, fileID string) []string {
	var tests []string
	for _, e := range g.Incoming(fileID) {
		if e.Type == depgraph.EdgeTestDefinedInFile {
			tests = append(tests, e.From)
		}
	}
	return tests
}

// CoverageGap identifies a node with insufficient test coverage.
type CoverageGap struct {
	NodeID    string
	Path      string
	TestCount int
	Band      CoverageBand
}

// FindCoverageGaps returns all source nodes with coverage below the given band.
func FindCoverageGaps(g *depgraph.Graph, minBand CoverageBand) []CoverageGap {
	if g == nil {
		return nil
	}

	sources := g.NodesByType(depgraph.NodeSourceFile)
	var gaps []CoverageGap

	for _, src := range sources {
		cov := CollectCovering(g, src.ID)
		if isBandBelow(cov.Band, minBand) {
			gaps = append(gaps, CoverageGap{
				NodeID:    src.ID,
				Path:      src.Path,
				TestCount: cov.TotalCount,
				Band:      cov.Band,
			})
		}
	}

	// Sort by test count ascending (worst coverage first).
	sort.Slice(gaps, func(i, j int) bool {
		if gaps[i].TestCount != gaps[j].TestCount {
			return gaps[i].TestCount < gaps[j].TestCount
		}
		return gaps[i].NodeID < gaps[j].NodeID
	})

	return gaps
}

// isBandBelow returns true if actual is strictly below required.
func isBandBelow(actual, required CoverageBand) bool {
	order := map[CoverageBand]int{
		CoverageLow:    0,
		CoverageMedium: 1,
		CoverageHigh:   2,
	}
	return order[actual] < order[required]
}

func sortedStringKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
