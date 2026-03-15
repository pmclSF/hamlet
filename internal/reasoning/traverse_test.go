package reasoning

import (
	"testing"

	"github.com/pmclSF/terrain/internal/depgraph"
)

// buildTestGraph creates a simple graph for traversal tests:
//
//	src:a --(ImportsModule, 0.9)--> src:b --(ImportsModule, 0.8)--> test:c
//	src:a --(ImportsModule, 0.7)--> src:d
func buildTestGraph() *depgraph.Graph {
	g := depgraph.NewGraph()
	g.AddNode(&depgraph.Node{ID: "src:a", Type: depgraph.NodeSourceFile, Path: "a.go"})
	g.AddNode(&depgraph.Node{ID: "src:b", Type: depgraph.NodeSourceFile, Path: "b.go"})
	g.AddNode(&depgraph.Node{ID: "test:c", Type: depgraph.NodeTest, Path: "c_test.go", Name: "TestC"})
	g.AddNode(&depgraph.Node{ID: "src:d", Type: depgraph.NodeSourceFile, Path: "d.go"})

	g.AddEdge(&depgraph.Edge{From: "src:a", To: "src:b", Type: depgraph.EdgeImportsModule, Confidence: 0.9})
	g.AddEdge(&depgraph.Edge{From: "src:b", To: "test:c", Type: depgraph.EdgeImportsModule, Confidence: 0.8})
	g.AddEdge(&depgraph.Edge{From: "src:a", To: "src:d", Type: depgraph.EdgeImportsModule, Confidence: 0.7})

	return g
}

func TestReachable_ForwardTraversal(t *testing.T) {
	g := buildTestGraph()
	cfg := DefaultTraversalConfig()
	cfg.Direction = "forward"

	results := Reachable(g, []string{"src:a"}, cfg)
	if len(results) == 0 {
		t.Fatal("expected results from forward traversal")
	}

	// Should reach src:b, test:c, src:d.
	ids := map[string]bool{}
	for _, r := range results {
		ids[r.NodeID] = true
	}
	for _, want := range []string{"src:b", "test:c", "src:d"} {
		if !ids[want] {
			t.Errorf("expected %s in results", want)
		}
	}

	// Start node should NOT be in results.
	if ids["src:a"] {
		t.Error("start node src:a should not be in results")
	}
}

func TestReachable_ReverseTraversal(t *testing.T) {
	g := buildTestGraph()
	cfg := DefaultTraversalConfig()
	cfg.Direction = "reverse"

	results := Reachable(g, []string{"test:c"}, cfg)
	ids := map[string]bool{}
	for _, r := range results {
		ids[r.NodeID] = true
	}

	// Reverse from test:c should find src:b (direct) and src:a (transitive).
	if !ids["src:b"] {
		t.Error("expected src:b in reverse results")
	}
	if !ids["src:a"] {
		t.Error("expected src:a in reverse results")
	}
}

func TestReachable_NilGraph(t *testing.T) {
	results := Reachable(nil, []string{"x"}, DefaultTraversalConfig())
	if results != nil {
		t.Errorf("expected nil for nil graph, got %v", results)
	}
}

func TestReachable_EmptyStartNodes(t *testing.T) {
	g := buildTestGraph()
	results := Reachable(g, nil, DefaultTraversalConfig())
	if results != nil {
		t.Errorf("expected nil for empty start nodes, got %v", results)
	}
}

func TestReachable_NonexistentStartNode(t *testing.T) {
	g := buildTestGraph()
	results := Reachable(g, []string{"nonexistent"}, DefaultTraversalConfig())
	if len(results) != 0 {
		t.Errorf("expected 0 results for nonexistent start, got %d", len(results))
	}
}

func TestReachable_StopAt(t *testing.T) {
	g := buildTestGraph()
	cfg := DefaultTraversalConfig()
	cfg.Direction = "forward"
	cfg.StopAt = func(n *depgraph.Node) bool {
		return n.ID == "src:b"
	}

	results := Reachable(g, []string{"src:a"}, cfg)
	ids := map[string]bool{}
	for _, r := range results {
		ids[r.NodeID] = true
	}

	// src:b should be in results (recorded) but test:c should NOT (not traversed past src:b).
	if !ids["src:b"] {
		t.Error("expected src:b in results (stop node is still recorded)")
	}
	if ids["test:c"] {
		t.Error("test:c should not be reached (stop at src:b)")
	}
}

func TestReachable_EdgeFilter(t *testing.T) {
	g := buildTestGraph()
	cfg := DefaultTraversalConfig()
	cfg.Direction = "forward"
	// Only traverse edges with confidence >= 0.8.
	cfg.EdgeFilter = func(e *depgraph.Edge) bool {
		return e.Confidence >= 0.8
	}

	results := Reachable(g, []string{"src:a"}, cfg)
	ids := map[string]bool{}
	for _, r := range results {
		ids[r.NodeID] = true
	}

	// src:b (0.9) should be reached, src:d (0.7) should not.
	if !ids["src:b"] {
		t.Error("expected src:b (edge confidence 0.9)")
	}
	if ids["src:d"] {
		t.Error("src:d should be filtered out (edge confidence 0.7 < 0.8)")
	}
}

func TestReachable_ConfidenceDecay(t *testing.T) {
	g := buildTestGraph()
	cfg := DefaultTraversalConfig()
	cfg.Direction = "forward"

	results := Reachable(g, []string{"src:a"}, cfg)

	// Find src:b — should have confidence = 1.0 * 0.9 * 0.85 = 0.765.
	for _, r := range results {
		if r.NodeID == "src:b" {
			expected := ScoreHop(1.0, 0.9, 0, 2, 0.85, 5) // outDegree of src:b = 1 neighbor
			if r.Confidence < expected-0.01 || r.Confidence > expected+0.01 {
				t.Errorf("src:b confidence = %v, expected ~%v", r.Confidence, expected)
			}
			if r.Depth != 1 {
				t.Errorf("src:b depth = %d, want 1", r.Depth)
			}
		}
	}
}

func TestReachable_ChainTracking(t *testing.T) {
	g := buildTestGraph()
	cfg := DefaultTraversalConfig()
	cfg.Direction = "forward"

	results := Reachable(g, []string{"src:a"}, cfg)

	for _, r := range results {
		if r.NodeID == "test:c" {
			if len(r.Chain) != 2 {
				t.Fatalf("test:c chain length = %d, want 2", len(r.Chain))
			}
			if r.Chain[0].From != "src:a" || r.Chain[0].To != "src:b" {
				t.Errorf("chain[0] = %s→%s, want src:a→src:b", r.Chain[0].From, r.Chain[0].To)
			}
			if r.Chain[1].From != "src:b" || r.Chain[1].To != "test:c" {
				t.Errorf("chain[1] = %s→%s, want src:b→test:c", r.Chain[1].From, r.Chain[1].To)
			}
		}
	}
}

func TestReachable_MaxDepth(t *testing.T) {
	g := buildTestGraph()
	cfg := DefaultTraversalConfig()
	cfg.Direction = "forward"
	cfg.MaxDepth = 1

	results := Reachable(g, []string{"src:a"}, cfg)
	ids := map[string]bool{}
	for _, r := range results {
		ids[r.NodeID] = true
	}

	// With maxDepth=1, should reach src:b and src:d but not test:c (depth 2).
	if !ids["src:b"] {
		t.Error("expected src:b at depth 1")
	}
	if ids["test:c"] {
		t.Error("test:c at depth 2 should be excluded with maxDepth=1")
	}
}

func TestReachableNodes(t *testing.T) {
	g := buildTestGraph()
	cfg := DefaultTraversalConfig()
	cfg.Direction = "forward"

	ids := ReachableNodes(g, []string{"src:a"}, cfg)
	if len(ids) == 0 {
		t.Fatal("expected reachable nodes")
	}

	found := map[string]bool{}
	for _, id := range ids {
		found[id] = true
	}
	if !found["src:b"] {
		t.Error("expected src:b in reachable nodes")
	}
}
