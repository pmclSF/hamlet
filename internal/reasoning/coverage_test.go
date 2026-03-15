package reasoning

import (
	"testing"

	"github.com/pmclSF/terrain/internal/depgraph"
)

func TestClassifyCoverageBand(t *testing.T) {
	tests := []struct {
		count int
		want  CoverageBand
	}{
		{0, CoverageLow},
		{1, CoverageMedium},
		{2, CoverageMedium},
		{3, CoverageHigh},
		{10, CoverageHigh},
	}
	for _, tt := range tests {
		got := ClassifyCoverageBand(tt.count)
		if got != tt.want {
			t.Errorf("ClassifyCoverageBand(%d) = %v, want %v", tt.count, got, tt.want)
		}
	}
}

func TestClassifyCoverageBandWithConfig(t *testing.T) {
	cfg := CoverageConfig{HighThreshold: 5, MediumThreshold: 2}
	if ClassifyCoverageBandWithConfig(1, cfg) != CoverageLow {
		t.Error("1 test should be Low with threshold 2")
	}
	if ClassifyCoverageBandWithConfig(3, cfg) != CoverageMedium {
		t.Error("3 tests should be Medium with threshold 5")
	}
	if ClassifyCoverageBandWithConfig(5, cfg) != CoverageHigh {
		t.Error("5 tests should be High with threshold 5")
	}
}

func buildCoverageTestGraph() *depgraph.Graph {
	g := depgraph.NewGraph()
	// Source file.
	g.AddNode(&depgraph.Node{ID: "file:src/auth.go", Type: depgraph.NodeSourceFile, Path: "src/auth.go"})
	// Test file that imports the source.
	g.AddNode(&depgraph.Node{ID: "file:test/auth_test.go", Type: depgraph.NodeTestFile, Path: "test/auth_test.go"})
	// Test defined in the test file.
	g.AddNode(&depgraph.Node{ID: "test:login", Type: depgraph.NodeTest, Path: "test/auth_test.go", Name: "TestLogin"})

	// test file → source file (import).
	g.AddEdge(&depgraph.Edge{From: "file:test/auth_test.go", To: "file:src/auth.go", Type: depgraph.EdgeImportsModule, Confidence: 0.9})
	// test → test file.
	g.AddEdge(&depgraph.Edge{From: "test:login", To: "file:test/auth_test.go", Type: depgraph.EdgeTestDefinedInFile, Confidence: 1.0})

	return g
}

func TestCollectCovering_DirectCoverage(t *testing.T) {
	g := buildCoverageTestGraph()
	cov := CollectCovering(g, "file:src/auth.go")

	if cov.TotalCount != 1 {
		t.Errorf("expected 1 covering test, got %d", cov.TotalCount)
	}
	if len(cov.DirectTests) != 1 || cov.DirectTests[0] != "test:login" {
		t.Errorf("expected direct test test:login, got %v", cov.DirectTests)
	}
	if cov.Band != CoverageMedium {
		t.Errorf("expected Medium band for 1 test, got %v", cov.Band)
	}
}

func TestCollectCovering_NilGraph(t *testing.T) {
	cov := CollectCovering(nil, "file:x")
	if cov.Band != CoverageLow {
		t.Errorf("expected Low band for nil graph, got %v", cov.Band)
	}
}

func TestCollectCovering_NonexistentNode(t *testing.T) {
	g := buildCoverageTestGraph()
	cov := CollectCovering(g, "nonexistent")
	if cov.TotalCount != 0 {
		t.Errorf("expected 0 tests for nonexistent node, got %d", cov.TotalCount)
	}
}

func TestFindCoverageGaps(t *testing.T) {
	g := depgraph.NewGraph()
	g.AddNode(&depgraph.Node{ID: "file:a.go", Type: depgraph.NodeSourceFile, Path: "a.go"})
	g.AddNode(&depgraph.Node{ID: "file:b.go", Type: depgraph.NodeSourceFile, Path: "b.go"})

	// a.go has no tests — should be a gap.
	// b.go has no tests — should be a gap.

	gaps := FindCoverageGaps(g, CoverageMedium)
	if len(gaps) != 2 {
		t.Fatalf("expected 2 gaps, got %d", len(gaps))
	}
	for _, gap := range gaps {
		if gap.Band != CoverageLow {
			t.Errorf("gap %s should have Low band, got %v", gap.NodeID, gap.Band)
		}
	}
}

func TestFindCoverageGaps_NilGraph(t *testing.T) {
	gaps := FindCoverageGaps(nil, CoverageMedium)
	if gaps != nil {
		t.Errorf("expected nil for nil graph, got %v", gaps)
	}
}
