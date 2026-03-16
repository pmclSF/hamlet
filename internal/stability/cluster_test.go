package stability

import (
	"testing"

	"github.com/pmclSF/terrain/internal/depgraph"
	"github.com/pmclSF/terrain/internal/models"
)

// buildTestGraph creates a minimal graph with shared environments and source
// files for cluster detection testing.
func buildTestGraph() *depgraph.Graph {
	g := depgraph.NewGraph()

	// Shared environment targeted by test-a and test-b.
	g.AddNode(&depgraph.Node{ID: "env:staging", Type: depgraph.NodeEnvironment, Name: "staging"})
	// Shared source file imported by test-b and test-c.
	g.AddNode(&depgraph.Node{ID: "file:src/db.js", Type: depgraph.NodeSourceFile, Name: "db.js", Path: "src/db.js"})

	// Test files.
	g.AddNode(&depgraph.Node{ID: "file:test/a.test.js", Type: depgraph.NodeTestFile, Path: "test/a.test.js"})
	g.AddNode(&depgraph.Node{ID: "file:test/b.test.js", Type: depgraph.NodeTestFile, Path: "test/b.test.js"})
	g.AddNode(&depgraph.Node{ID: "file:test/c.test.js", Type: depgraph.NodeTestFile, Path: "test/c.test.js"})
	g.AddNode(&depgraph.Node{ID: "file:test/d.test.js", Type: depgraph.NodeTestFile, Path: "test/d.test.js"})

	// Tests.
	g.AddNode(&depgraph.Node{ID: "test:a:1:testA", Type: depgraph.NodeTest, Name: "testA", Path: "test/a.test.js"})
	g.AddNode(&depgraph.Node{ID: "test:b:1:testB", Type: depgraph.NodeTest, Name: "testB", Path: "test/b.test.js"})
	g.AddNode(&depgraph.Node{ID: "test:c:1:testC", Type: depgraph.NodeTest, Name: "testC", Path: "test/c.test.js"})
	g.AddNode(&depgraph.Node{ID: "test:d:1:testD", Type: depgraph.NodeTest, Name: "testD", Path: "test/d.test.js"})

	// Test → file edges.
	g.AddEdge(&depgraph.Edge{From: "test:a:1:testA", To: "file:test/a.test.js", Type: depgraph.EdgeTestDefinedInFile})
	g.AddEdge(&depgraph.Edge{From: "test:b:1:testB", To: "file:test/b.test.js", Type: depgraph.EdgeTestDefinedInFile})
	g.AddEdge(&depgraph.Edge{From: "test:c:1:testC", To: "file:test/c.test.js", Type: depgraph.EdgeTestDefinedInFile})
	g.AddEdge(&depgraph.Edge{From: "test:d:1:testD", To: "file:test/d.test.js", Type: depgraph.EdgeTestDefinedInFile})

	// Test/file → dependency edges.
	// Tests A and B target the same environment.
	g.AddEdge(&depgraph.Edge{From: "file:test/a.test.js", To: "env:staging", Type: depgraph.EdgeTargetsEnvironment})
	g.AddEdge(&depgraph.Edge{From: "file:test/b.test.js", To: "env:staging", Type: depgraph.EdgeTargetsEnvironment})
	// Tests B and C import the same source file.
	g.AddEdge(&depgraph.Edge{From: "file:test/b.test.js", To: "file:src/db.js", Type: depgraph.EdgeImportsModule})
	g.AddEdge(&depgraph.Edge{From: "file:test/c.test.js", To: "file:src/db.js", Type: depgraph.EdgeImportsModule})

	return g
}

func flakySignals(files ...string) []models.Signal {
	var sigs []models.Signal
	for _, f := range files {
		sigs = append(sigs, models.Signal{
			Type:     "flakyTest",
			Location: models.SignalLocation{File: f},
		})
	}
	return sigs
}

func TestDetectClusters_SharedEnvironment(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	// Tests A and B are flaky — they share env:staging.
	signals := flakySignals("test/a.test.js", "test/b.test.js")
	result := DetectClusters(g, signals)

	if result.UnstableTestCount < 2 {
		t.Fatalf("expected >= 2 unstable tests, got %d", result.UnstableTestCount)
	}
	if len(result.Clusters) == 0 {
		t.Fatal("expected at least one cluster")
	}

	// Should find the environment cluster.
	found := false
	for _, c := range result.Clusters {
		if c.CauseNodeID == "env:staging" {
			found = true
			if c.CauseKind != CauseEnvironment {
				t.Errorf("cause kind = %s, want shared_environment", c.CauseKind)
			}
			if len(c.Members) < 2 {
				t.Errorf("expected >= 2 members, got %d", len(c.Members))
			}
			if c.Confidence < 0.5 {
				t.Errorf("confidence = %f, want >= 0.5", c.Confidence)
			}
			if c.Remediation == "" {
				t.Error("expected non-empty remediation")
			}
		}
	}
	if !found {
		t.Error("expected cluster with cause env:staging")
	}
}

func TestDetectClusters_SharedSourceFile(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	// Tests B and C are flaky — they share file:src/db.js.
	signals := flakySignals("test/b.test.js", "test/c.test.js")
	result := DetectClusters(g, signals)

	found := false
	for _, c := range result.Clusters {
		if c.CauseNodeID == "file:src/db.js" {
			found = true
			if c.CauseKind != CauseSourceFile {
				t.Errorf("cause kind = %s, want shared_source", c.CauseKind)
			}
		}
	}
	if !found {
		t.Error("expected cluster with cause file:src/db.js")
	}
}

func TestDetectClusters_MultipleClusters(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	// All three tests (A, B, C) are flaky.
	signals := flakySignals("test/a.test.js", "test/b.test.js", "test/c.test.js")
	result := DetectClusters(g, signals)

	// Should find clusters for: env:staging and file:src/db.js.
	if len(result.Clusters) < 2 {
		t.Errorf("expected >= 2 clusters, got %d", len(result.Clusters))
	}

	kinds := map[CauseKind]bool{}
	for _, c := range result.Clusters {
		kinds[c.CauseKind] = true
	}
	for _, expected := range []CauseKind{CauseEnvironment, CauseSourceFile} {
		if !kinds[expected] {
			t.Errorf("expected cluster with cause kind %s", expected)
		}
	}

	// ClusteredTestCount should cover all 3.
	if result.ClusteredTestCount < 3 {
		t.Errorf("clustered test count = %d, want >= 3", result.ClusteredTestCount)
	}
}

func TestDetectClusters_NoSharedDeps(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	// Tests A and D are flaky, but D has no shared deps with A.
	signals := flakySignals("test/a.test.js", "test/d.test.js")
	result := DetectClusters(g, signals)

	// No cluster should include both A and D.
	for _, c := range result.Clusters {
		hasA := false
		hasD := false
		for _, m := range c.Members {
			if m == "test:a:1:testA" {
				hasA = true
			}
			if m == "test:d:1:testD" {
				hasD = true
			}
		}
		if hasA && hasD {
			t.Error("tests A and D should not be in the same cluster")
		}
	}
}

func TestDetectClusters_SingleUnstableTest(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	// Only one flaky test — cannot form clusters.
	signals := flakySignals("test/a.test.js")
	result := DetectClusters(g, signals)

	if len(result.Clusters) != 0 {
		t.Errorf("expected 0 clusters for single unstable test, got %d", len(result.Clusters))
	}
}

func TestDetectClusters_NoSignals(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	result := DetectClusters(g, nil)

	if len(result.Clusters) != 0 {
		t.Errorf("expected 0 clusters with no signals, got %d", len(result.Clusters))
	}
	if result.UnstableTestCount != 0 {
		t.Errorf("expected 0 unstable tests, got %d", result.UnstableTestCount)
	}
}

func TestDetectClusters_NilGraph(t *testing.T) {
	t.Parallel()
	signals := flakySignals("test/a.test.js")

	result := DetectClusters(nil, signals)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Clusters) != 0 {
		t.Errorf("expected 0 clusters, got %d", len(result.Clusters))
	}
}

func TestDetectClusters_Deterministic(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()
	signals := flakySignals("test/a.test.js", "test/b.test.js", "test/c.test.js")

	r1 := DetectClusters(g, signals)
	r2 := DetectClusters(g, signals)

	if len(r1.Clusters) != len(r2.Clusters) {
		t.Fatalf("non-deterministic: %d vs %d clusters", len(r1.Clusters), len(r2.Clusters))
	}
	for i := range r1.Clusters {
		if r1.Clusters[i].ID != r2.Clusters[i].ID {
			t.Errorf("non-deterministic cluster order at %d: %s vs %s", i, r1.Clusters[i].ID, r2.Clusters[i].ID)
		}
	}
}

func TestDetectClusters_UnstableSuiteSignal(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	// Mix of flakyTest and unstableSuite signals.
	signals := []models.Signal{
		{Type: "flakyTest", Location: models.SignalLocation{File: "test/a.test.js"}},
		{Type: "unstableSuite", Location: models.SignalLocation{File: "test/b.test.js"}},
	}
	result := DetectClusters(g, signals)

	if result.UnstableTestCount < 2 {
		t.Fatalf("expected >= 2 unstable tests, got %d", result.UnstableTestCount)
	}

	// Should still find the environment cluster.
	found := false
	for _, c := range result.Clusters {
		if c.CauseNodeID == "env:staging" {
			found = true
		}
	}
	if !found {
		t.Error("expected environment cluster from mixed signal types")
	}
}

func TestClusterConfidence_Scales(t *testing.T) {
	t.Parallel()

	// Higher concentration → higher confidence.
	c2of10 := clusterConfidence(CauseEnvironment, 2, 10)
	c8of10 := clusterConfidence(CauseEnvironment, 8, 10)
	if c8of10 <= c2of10 {
		t.Errorf("expected higher confidence for larger concentration: %f <= %f", c8of10, c2of10)
	}

	// Environment > source file base confidence.
	cEnv := clusterConfidence(CauseEnvironment, 3, 10)
	cSrc := clusterConfidence(CauseSourceFile, 3, 10)
	if cEnv <= cSrc {
		t.Errorf("expected environment > source_file: %f <= %f", cEnv, cSrc)
	}
}

func TestClusterRemediation_NonEmpty(t *testing.T) {
	t.Parallel()
	kinds := []CauseKind{CauseFixture, CauseHelper, CauseExternalService, CauseEnvironment, CauseSourceFile}
	for _, k := range kinds {
		r := clusterRemediation(k, "test-dep")
		if r == "" {
			t.Errorf("expected non-empty remediation for %s", k)
		}
	}
}

func TestSortedClusters_LargestFirst(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()
	signals := flakySignals("test/a.test.js", "test/b.test.js", "test/c.test.js")
	result := DetectClusters(g, signals)

	if len(result.Clusters) < 2 {
		t.Skip("need >= 2 clusters to test ordering")
	}

	for i := 1; i < len(result.Clusters); i++ {
		if len(result.Clusters[i].Members) > len(result.Clusters[i-1].Members) {
			t.Errorf("clusters not sorted by size: cluster[%d] has %d members > cluster[%d] has %d members",
				i, len(result.Clusters[i].Members), i-1, len(result.Clusters[i-1].Members))
		}
	}
}
