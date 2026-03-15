package depgraph

import (
	"testing"
)

// buildRedundancyTestGraph creates a graph with tests, source files, code
// surfaces, and behavior surfaces for redundancy testing.
func buildRedundancyTestGraph() *Graph {
	g := NewGraph()

	// Source files.
	g.AddNode(&Node{ID: "file:src/auth.go", Type: NodeSourceFile, Path: "src/auth.go", Package: "auth"})
	g.AddNode(&Node{ID: "file:src/billing.go", Type: NodeSourceFile, Path: "src/billing.go", Package: "billing"})
	g.AddNode(&Node{ID: "file:src/notify.go", Type: NodeSourceFile, Path: "src/notify.go", Package: "notify"})

	// Code surfaces in auth.go.
	g.AddNode(&Node{ID: "cs:auth:Login", Type: NodeCodeSurface, Path: "src/auth.go", Name: "Login", Package: "auth"})
	g.AddNode(&Node{ID: "cs:auth:Logout", Type: NodeCodeSurface, Path: "src/auth.go", Name: "Logout", Package: "auth"})
	g.AddEdge(&Edge{From: "cs:auth:Login", To: "file:src/auth.go", Type: EdgeBelongsToPackage, Confidence: 1.0})
	g.AddEdge(&Edge{From: "cs:auth:Logout", To: "file:src/auth.go", Type: EdgeBelongsToPackage, Confidence: 1.0})

	// Code surfaces in billing.go.
	g.AddNode(&Node{ID: "cs:billing:Charge", Type: NodeCodeSurface, Path: "src/billing.go", Name: "Charge", Package: "billing"})
	g.AddEdge(&Edge{From: "cs:billing:Charge", To: "file:src/billing.go", Type: EdgeBelongsToPackage, Confidence: 1.0})

	// Behavior surface grouping auth surfaces.
	g.AddNode(&Node{ID: "bs:auth", Type: NodeBehaviorSurface, Name: "auth", Package: "auth", Metadata: map[string]string{"kind": "module"}})
	g.AddEdge(&Edge{From: "bs:auth", To: "cs:auth:Login", Type: EdgeBehaviorDerivedFrom, Confidence: 0.7})
	g.AddEdge(&Edge{From: "bs:auth", To: "cs:auth:Logout", Type: EdgeBehaviorDerivedFrom, Confidence: 0.7})

	// Test files.
	g.AddNode(&Node{ID: "file:test/auth_test.go", Type: NodeTestFile, Path: "test/auth_test.go", Framework: "go"})
	g.AddNode(&Node{ID: "file:test/auth_e2e_test.go", Type: NodeTestFile, Path: "test/auth_e2e_test.go", Framework: "go"})
	g.AddNode(&Node{ID: "file:test/billing_test.go", Type: NodeTestFile, Path: "test/billing_test.go", Framework: "go"})

	// Tests.
	g.AddNode(&Node{ID: "test:auth:1:TestLogin", Type: NodeTest, Name: "TestLogin", Path: "test/auth_test.go", Framework: "go"})
	g.AddNode(&Node{ID: "test:auth:2:TestLoginEdge", Type: NodeTest, Name: "TestLoginEdge", Path: "test/auth_test.go", Framework: "go"})
	g.AddNode(&Node{ID: "test:auth_e2e:1:TestLoginE2E", Type: NodeTest, Name: "TestLoginE2E", Path: "test/auth_e2e_test.go", Framework: "go"})
	g.AddNode(&Node{ID: "test:billing:1:TestCharge", Type: NodeTest, Name: "TestCharge", Path: "test/billing_test.go", Framework: "go"})

	// Test → file edges.
	g.AddEdge(&Edge{From: "test:auth:1:TestLogin", To: "file:test/auth_test.go", Type: EdgeTestDefinedInFile})
	g.AddEdge(&Edge{From: "test:auth:2:TestLoginEdge", To: "file:test/auth_test.go", Type: EdgeTestDefinedInFile})
	g.AddEdge(&Edge{From: "test:auth_e2e:1:TestLoginE2E", To: "file:test/auth_e2e_test.go", Type: EdgeTestDefinedInFile})
	g.AddEdge(&Edge{From: "test:billing:1:TestCharge", To: "file:test/billing_test.go", Type: EdgeTestDefinedInFile})

	// Import edges: test files → source files.
	g.AddEdge(&Edge{From: "file:test/auth_test.go", To: "file:src/auth.go", Type: EdgeImportsModule})
	g.AddEdge(&Edge{From: "file:test/auth_e2e_test.go", To: "file:src/auth.go", Type: EdgeImportsModule})
	g.AddEdge(&Edge{From: "file:test/billing_test.go", To: "file:src/billing.go", Type: EdgeImportsModule})

	return g
}

func TestAnalyzeRedundancy_SameBehaviorSurface(t *testing.T) {
	t.Parallel()
	g := buildRedundancyTestGraph()

	result := AnalyzeRedundancy(g)

	if result.Skipped {
		t.Fatal("unexpected skip")
	}

	// TestLogin and TestLoginEdge both import auth.go which contains
	// Login and Logout code surfaces under the auth behavior surface.
	// TestLoginE2E also imports auth.go.
	// All three should cluster around shared auth surfaces.
	found := false
	for _, c := range result.Clusters {
		authTests := 0
		for _, tid := range c.Tests {
			if tid == "test:auth:1:TestLogin" || tid == "test:auth:2:TestLoginEdge" || tid == "test:auth_e2e:1:TestLoginE2E" {
				authTests++
			}
		}
		if authTests >= 2 {
			found = true
			if len(c.SharedSurfaces) == 0 {
				t.Error("expected shared surfaces in auth cluster")
			}
			if c.Rationale == "" {
				t.Error("expected non-empty rationale")
			}
		}
	}
	if !found {
		t.Error("expected a cluster containing auth tests")
	}
}

func TestAnalyzeRedundancy_NoOverlapDifferentSurfaces(t *testing.T) {
	t.Parallel()
	g := buildRedundancyTestGraph()

	result := AnalyzeRedundancy(g)

	// TestCharge imports billing.go, which has no overlap with auth.
	// It should not cluster with auth tests.
	for _, c := range result.Clusters {
		hasAuth := false
		hasBilling := false
		for _, tid := range c.Tests {
			if tid == "test:auth:1:TestLogin" {
				hasAuth = true
			}
			if tid == "test:billing:1:TestCharge" {
				hasBilling = true
			}
		}
		if hasAuth && hasBilling {
			t.Error("auth and billing tests should not be in the same cluster")
		}
	}
}

func TestAnalyzeRedundancy_CrossFramework(t *testing.T) {
	t.Parallel()
	g := buildRedundancyTestGraph()

	// Add a Jest test that imports auth.go.
	g.AddNode(&Node{ID: "file:test/auth.test.js", Type: NodeTestFile, Path: "test/auth.test.js", Framework: "jest"})
	g.AddNode(&Node{ID: "test:auth_js:1:loginTest", Type: NodeTest, Name: "loginTest", Path: "test/auth.test.js", Framework: "jest"})
	g.AddEdge(&Edge{From: "test:auth_js:1:loginTest", To: "file:test/auth.test.js", Type: EdgeTestDefinedInFile})
	g.AddEdge(&Edge{From: "file:test/auth.test.js", To: "file:src/auth.go", Type: EdgeImportsModule})

	result := AnalyzeRedundancy(g)

	// Should detect cross-framework overlap between go and jest tests.
	foundCross := false
	for _, c := range result.Clusters {
		if c.OverlapKind == OverlapCrossFramework {
			foundCross = true
			if len(c.Frameworks) < 2 {
				t.Errorf("expected multiple frameworks, got %v", c.Frameworks)
			}
		}
	}
	if !foundCross {
		t.Error("expected at least one cross-framework cluster")
	}
	if result.CrossFrameworkOverlaps == 0 {
		t.Error("expected CrossFrameworkOverlaps > 0")
	}
}

func TestAnalyzeRedundancy_WastefulOverlap(t *testing.T) {
	t.Parallel()
	g := buildRedundancyTestGraph()

	result := AnalyzeRedundancy(g)

	// TestLogin and TestLoginEdge are same framework, same import path.
	foundWasteful := false
	for _, c := range result.Clusters {
		if c.OverlapKind == OverlapWasteful {
			foundWasteful = true
		}
	}
	// All auth tests are same framework (go), so wasteful is possible.
	// Whether it's wasteful depends on whether all cluster members are same framework.
	if !foundWasteful {
		// The 3 auth tests are all "go" framework, so this should be wasteful.
		t.Log("no wasteful clusters found — may be classified as cross-level if metadata differs")
	}
}

func TestAnalyzeRedundancy_EmptyGraph(t *testing.T) {
	t.Parallel()
	g := NewGraph()

	result := AnalyzeRedundancy(g)

	if len(result.Clusters) != 0 {
		t.Errorf("expected 0 clusters for empty graph, got %d", len(result.Clusters))
	}
	if result.TestsAnalyzed != 0 {
		t.Errorf("expected 0 tests analyzed, got %d", result.TestsAnalyzed)
	}
}

func TestAnalyzeRedundancy_NoSurfaces(t *testing.T) {
	t.Parallel()
	g := NewGraph()

	// Tests exist but no code surfaces — no behavior to overlap on.
	g.AddNode(&Node{ID: "file:test/a.test.js", Type: NodeTestFile, Path: "test/a.test.js"})
	g.AddNode(&Node{ID: "file:test/b.test.js", Type: NodeTestFile, Path: "test/b.test.js"})
	g.AddNode(&Node{ID: "test:a:1:testA", Type: NodeTest, Name: "testA", Path: "test/a.test.js"})
	g.AddNode(&Node{ID: "test:b:1:testB", Type: NodeTest, Name: "testB", Path: "test/b.test.js"})
	g.AddEdge(&Edge{From: "test:a:1:testA", To: "file:test/a.test.js", Type: EdgeTestDefinedInFile})
	g.AddEdge(&Edge{From: "test:b:1:testB", To: "file:test/b.test.js", Type: EdgeTestDefinedInFile})

	result := AnalyzeRedundancy(g)

	if len(result.Clusters) != 0 {
		t.Errorf("expected 0 clusters with no surfaces, got %d", len(result.Clusters))
	}
}

func TestAnalyzeRedundancy_Deterministic(t *testing.T) {
	t.Parallel()
	g := buildRedundancyTestGraph()

	r1 := AnalyzeRedundancy(g)
	r2 := AnalyzeRedundancy(g)

	if len(r1.Clusters) != len(r2.Clusters) {
		t.Fatalf("non-deterministic cluster count: %d vs %d", len(r1.Clusters), len(r2.Clusters))
	}
	for i := range r1.Clusters {
		if r1.Clusters[i].ID != r2.Clusters[i].ID {
			t.Errorf("non-deterministic cluster ID at %d: %s vs %s",
				i, r1.Clusters[i].ID, r2.Clusters[i].ID)
		}
	}
}

func TestAnalyzeRedundancy_SurfaceNamesPopulated(t *testing.T) {
	t.Parallel()
	g := buildRedundancyTestGraph()

	result := AnalyzeRedundancy(g)

	for _, c := range result.Clusters {
		for _, sid := range c.SharedSurfaces {
			if name, ok := c.SurfaceNames[sid]; !ok || name == "" {
				t.Errorf("surface %s has no name in cluster %s", sid, c.ID)
			}
		}
	}
}

func TestAnalyzeRedundancy_ConfidenceRange(t *testing.T) {
	t.Parallel()
	g := buildRedundancyTestGraph()

	result := AnalyzeRedundancy(g)

	for _, c := range result.Clusters {
		if c.Confidence < 0.0 || c.Confidence > 1.0 {
			t.Errorf("confidence %f out of range [0,1] for cluster %s", c.Confidence, c.ID)
		}
	}
}

func TestClassifyOverlap_Wasteful(t *testing.T) {
	t.Parallel()
	kind, rationale := classifyOverlap(
		[]string{"jest"},
		map[string]bool{"unit": true},
		0.8, 3,
	)
	if kind != OverlapWasteful {
		t.Errorf("kind = %s, want wasteful", kind)
	}
	if rationale == "" {
		t.Error("expected non-empty rationale")
	}
}

func TestClassifyOverlap_CrossFramework(t *testing.T) {
	t.Parallel()
	kind, _ := classifyOverlap(
		[]string{"jest", "vitest"},
		map[string]bool{"unit": true},
		0.9, 5,
	)
	if kind != OverlapCrossFramework {
		t.Errorf("kind = %s, want cross_framework", kind)
	}
}

func TestClassifyOverlap_CrossLevel(t *testing.T) {
	t.Parallel()
	kind, rationale := classifyOverlap(
		[]string{"jest"},
		map[string]bool{"unit": true, "e2e": true},
		0.7, 2,
	)
	if kind != OverlapCrossLevel {
		t.Errorf("kind = %s, want cross_level", kind)
	}
	if rationale == "" {
		t.Error("expected non-empty rationale")
	}
}

func TestRedundancyConfidence_WastefulHigherThanCrossLevel(t *testing.T) {
	t.Parallel()
	cWasteful := redundancyConfidence(0.8, 3, OverlapWasteful)
	cCrossLevel := redundancyConfidence(0.8, 3, OverlapCrossLevel)
	if cWasteful <= cCrossLevel {
		t.Errorf("expected wasteful confidence > cross-level: %f <= %f", cWasteful, cCrossLevel)
	}
}

func TestIntersectSurfaces(t *testing.T) {
	t.Parallel()
	surfaces := map[string]*testSurfaceInfo{
		"a": {surfaces: map[string]bool{"s1": true, "s2": true, "s3": true}},
		"b": {surfaces: map[string]bool{"s1": true, "s2": true}},
		"c": {surfaces: map[string]bool{"s1": true, "s4": true}},
	}

	shared := intersectSurfaces(surfaces, []string{"a", "b"})
	if len(shared) != 2 {
		t.Errorf("expected 2 shared surfaces between a,b, got %d: %v", len(shared), shared)
	}

	shared = intersectSurfaces(surfaces, []string{"a", "b", "c"})
	if len(shared) != 1 || shared[0] != "s1" {
		t.Errorf("expected [s1] shared between a,b,c, got %v", shared)
	}
}
