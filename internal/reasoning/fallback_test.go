package reasoning

import (
	"testing"

	"github.com/pmclSF/terrain/internal/depgraph"
)

func buildFallbackTestGraph() *depgraph.Graph {
	g := depgraph.NewGraph()
	g.AddNode(&depgraph.Node{ID: "src:a", Type: depgraph.NodeSourceFile, Path: "pkg/auth/login.go", Package: "pkg:auth"})
	g.AddNode(&depgraph.Node{ID: "test:1", Type: depgraph.NodeTest, Path: "pkg/auth/login_test.go", Package: "pkg:auth", Name: "TestLogin"})
	g.AddNode(&depgraph.Node{ID: "test:2", Type: depgraph.NodeTest, Path: "pkg/auth/signup_test.go", Package: "pkg:auth", Name: "TestSignup"})
	g.AddNode(&depgraph.Node{ID: "test:3", Type: depgraph.NodeTest, Path: "pkg/billing/charge_test.go", Package: "pkg:billing", Name: "TestCharge"})
	return g
}

func TestExpandFallback_NoFallbackNeeded(t *testing.T) {
	g := buildFallbackTestGraph()
	result := ExpandFallback(g, []string{"src:a"}, []string{"test:1"}, DefaultFallbackConfig())
	if result.Strategy != FallbackNone {
		t.Errorf("expected no fallback, got %v", result.Strategy)
	}
}

func TestExpandFallback_PackageFallback(t *testing.T) {
	g := buildFallbackTestGraph()
	result := ExpandFallback(g, []string{"src:a"}, nil, DefaultFallbackConfig())

	if result.Strategy != FallbackPackage {
		t.Errorf("expected package fallback, got %v", result.Strategy)
	}
	if len(result.NodeIDs) == 0 {
		t.Fatal("expected fallback to find tests")
	}

	// Should find tests in pkg:auth but not pkg:billing.
	found := map[string]bool{}
	for _, id := range result.NodeIDs {
		found[id] = true
	}
	if !found["test:1"] || !found["test:2"] {
		t.Error("expected test:1 and test:2 in package fallback")
	}
	if found["test:3"] {
		t.Error("test:3 (different package) should not be in package fallback")
	}
}

func TestExpandFallback_DirectoryFallback(t *testing.T) {
	g := buildFallbackTestGraph()
	cfg := FallbackConfig{
		MinResults: 1,
		Strategies: []FallbackStrategy{FallbackDirectory},
	}

	result := ExpandFallback(g, []string{"src:a"}, nil, cfg)
	if result.Strategy != FallbackDirectory {
		t.Errorf("expected directory fallback, got %v", result.Strategy)
	}
	if len(result.NodeIDs) == 0 {
		t.Fatal("expected fallback to find tests in same directory")
	}
}

func TestExpandFallback_NoResults(t *testing.T) {
	g := depgraph.NewGraph()
	g.AddNode(&depgraph.Node{ID: "src:a", Type: depgraph.NodeSourceFile, Path: "isolated.go", Package: "pkg:isolated"})

	result := ExpandFallback(g, []string{"src:a"}, nil, DefaultFallbackConfig())
	if result.Strategy != FallbackNone {
		t.Errorf("expected no fallback when no tests exist, got %v", result.Strategy)
	}
}

func TestExpandFallback_ExcludesSeeds(t *testing.T) {
	g := buildFallbackTestGraph()
	// Exclude test:1 and test:2 as already in results.
	result := ExpandFallback(g, []string{"src:a"}, []string{"test:1", "test:2"}, FallbackConfig{
		MinResults: 3,
		Strategies: []FallbackStrategy{FallbackPackage},
	})

	// Package fallback should find no additional tests in pkg:auth.
	for _, id := range result.NodeIDs {
		if id == "test:1" || id == "test:2" {
			t.Errorf("excluded node %s should not appear in fallback", id)
		}
	}
}
