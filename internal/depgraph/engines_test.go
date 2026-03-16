package depgraph

import (
	"testing"
)

// buildTestGraph creates a small but realistic graph for engine testing.
//
// Structure:
//
//	test:a.test.js:10:login → file:a.test.js → file:src/auth.js
//	test:a.test.js:20:logout → file:a.test.js → file:src/auth.js
//	test:b.test.js:10:signup → file:b.test.js → file:src/auth.js
//	test:b.test.js:20:verify → file:b.test.js → file:src/util.js
//	file:src/auth.js → file:src/util.js (source imports source)
func buildTestGraph() *Graph {
	g := NewGraph()

	// Test files.
	g.AddNode(&Node{ID: "file:a.test.js", Type: NodeTestFile, Path: "a.test.js", Package: "test"})
	g.AddNode(&Node{ID: "file:b.test.js", Type: NodeTestFile, Path: "b.test.js", Package: "test"})

	// Source files.
	g.AddNode(&Node{ID: "file:src/auth.js", Type: NodeSourceFile, Path: "src/auth.js", Package: "src"})
	g.AddNode(&Node{ID: "file:src/util.js", Type: NodeSourceFile, Path: "src/util.js", Package: "src"})

	// Tests.
	g.AddNode(&Node{ID: "test:a.test.js:10:login", Type: NodeTest, Path: "a.test.js", Name: "should login", Package: "test", Line: 10})
	g.AddNode(&Node{ID: "test:a.test.js:20:logout", Type: NodeTest, Path: "a.test.js", Name: "should logout", Package: "test", Line: 20})
	g.AddNode(&Node{ID: "test:b.test.js:10:signup", Type: NodeTest, Path: "b.test.js", Name: "should signup", Package: "test", Line: 10})
	g.AddNode(&Node{ID: "test:b.test.js:20:verify", Type: NodeTest, Path: "b.test.js", Name: "should verify email", Package: "test", Line: 20})

	// Test → file edges.
	g.AddEdge(&Edge{From: "test:a.test.js:10:login", To: "file:a.test.js", Type: EdgeTestDefinedInFile, Confidence: 1.0, EvidenceType: EvidenceStaticAnalysis})
	g.AddEdge(&Edge{From: "test:a.test.js:20:logout", To: "file:a.test.js", Type: EdgeTestDefinedInFile, Confidence: 1.0, EvidenceType: EvidenceStaticAnalysis})
	g.AddEdge(&Edge{From: "test:b.test.js:10:signup", To: "file:b.test.js", Type: EdgeTestDefinedInFile, Confidence: 1.0, EvidenceType: EvidenceStaticAnalysis})
	g.AddEdge(&Edge{From: "test:b.test.js:20:verify", To: "file:b.test.js", Type: EdgeTestDefinedInFile, Confidence: 1.0, EvidenceType: EvidenceStaticAnalysis})

	// Test file → source file imports.
	g.AddEdge(&Edge{From: "file:a.test.js", To: "file:src/auth.js", Type: EdgeImportsModule, Confidence: 1.0, EvidenceType: EvidenceStaticAnalysis})
	g.AddEdge(&Edge{From: "file:b.test.js", To: "file:src/auth.js", Type: EdgeImportsModule, Confidence: 1.0, EvidenceType: EvidenceStaticAnalysis})
	g.AddEdge(&Edge{From: "file:b.test.js", To: "file:src/util.js", Type: EdgeImportsModule, Confidence: 1.0, EvidenceType: EvidenceStaticAnalysis})

	// Source → source import.
	g.AddEdge(&Edge{From: "file:src/auth.js", To: "file:src/util.js", Type: EdgeSourceImportsSource, Confidence: 0.8, EvidenceType: EvidenceStaticAnalysis})

	return g
}

// --- Fanout Tests ---

func TestAnalyzeFanout_Basic(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	result := AnalyzeFanout(g, 3)

	if result.NodeCount != 8 {
		t.Errorf("expected 8 nodes, got %d", result.NodeCount)
	}
	if result.Threshold != 3 {
		t.Errorf("expected threshold 3, got %d", result.Threshold)
	}

	// Entries should be sorted by transitive fanout descending.
	if len(result.Entries) != 8 {
		t.Fatalf("expected 8 entries, got %d", len(result.Entries))
	}

	// b.test.js imports auth.js and util.js; auth.js→util.js gives
	// transitive fanout of 3 from b.test.js (auth.js, util.js, and
	// the source-to-source connection).
	for i := 1; i < len(result.Entries); i++ {
		if result.Entries[i].TransitiveFanout > result.Entries[i-1].TransitiveFanout {
			t.Errorf("entries not sorted: [%d].transitive=%d > [%d].transitive=%d",
				i, result.Entries[i].TransitiveFanout, i-1, result.Entries[i-1].TransitiveFanout)
		}
	}

	// At least one node should be flagged (b.test.js has fanout ≥ 3).
	if result.FlaggedCount < 1 {
		t.Errorf("expected at least 1 flagged node, got %d", result.FlaggedCount)
	}

	// Verify flagged consistency.
	flagged := 0
	for _, e := range result.Entries {
		if e.Flagged {
			flagged++
			if e.TransitiveFanout < result.Threshold {
				t.Errorf("node %s flagged but transitive=%d < threshold=%d",
					e.NodeID, e.TransitiveFanout, result.Threshold)
			}
		}
	}
	if flagged != result.FlaggedCount {
		t.Errorf("flagged count mismatch: counted %d, reported %d", flagged, result.FlaggedCount)
	}
}

func TestAnalyzeFanout_DefaultThreshold(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	result := AnalyzeFanout(g, 0)
	if result.Threshold != DefaultFanoutThreshold {
		t.Errorf("expected default threshold %d, got %d", DefaultFanoutThreshold, result.Threshold)
	}
}

// --- Coverage Tests ---

func TestAnalyzeCoverage_Basic(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	result := AnalyzeCoverage(g)

	if result.SourceCount != 2 {
		t.Errorf("expected 2 sources, got %d", result.SourceCount)
	}

	// auth.js is imported by both test files → should have 4 direct tests.
	// util.js is imported by b.test.js → 2 direct + indirect via auth.js.
	totalBands := result.BandCounts[CoverageBandHigh] +
		result.BandCounts[CoverageBandMedium] +
		result.BandCounts[CoverageBandLow]
	if totalBands != result.SourceCount {
		t.Errorf("band counts (%d) don't sum to source count (%d)", totalBands, result.SourceCount)
	}

	// auth.js should have high coverage (4 tests import it).
	for _, src := range result.Sources {
		if src.Path == "src/auth.js" {
			if src.TestCount < 2 {
				t.Errorf("auth.js expected ≥2 tests, got %d", src.TestCount)
			}
			if src.Band != CoverageBandHigh {
				t.Errorf("auth.js expected High band, got %s", src.Band)
			}
		}
	}
}

func TestAnalyzeCoverage_SortedAscending(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	result := AnalyzeCoverage(g)

	for i := 1; i < len(result.Sources); i++ {
		if result.Sources[i].TestCount < result.Sources[i-1].TestCount {
			t.Errorf("sources not sorted ascending: [%d]=%d < [%d]=%d",
				i, result.Sources[i].TestCount, i-1, result.Sources[i-1].TestCount)
		}
	}
}

// --- Duplicate Tests ---

func TestDetectDuplicates_EmptyGraph(t *testing.T) {
	t.Parallel()
	g := NewGraph()

	result := DetectDuplicates(g)

	if result.TestsAnalyzed != 0 {
		t.Errorf("expected 0 tests analyzed, got %d", result.TestsAnalyzed)
	}
	if len(result.Clusters) != 0 {
		t.Errorf("expected 0 clusters, got %d", len(result.Clusters))
	}
}

func TestDetectDuplicates_SimilarTests(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	result := DetectDuplicates(g)

	if result.TestsAnalyzed != 4 {
		t.Errorf("expected 4 tests analyzed, got %d", result.TestsAnalyzed)
	}

	// login and logout are in the same file, same package → should be candidates.
	// Whether they cluster depends on similarity scoring.
	// At minimum, the engine should run without error and produce valid output.
	for _, c := range result.Clusters {
		if len(c.Tests) < 2 {
			t.Errorf("cluster %d has < 2 tests: %v", c.ID, c.Tests)
		}
		if c.Similarity < 0 || c.Similarity > 1 {
			t.Errorf("cluster %d similarity out of range: %f", c.ID, c.Similarity)
		}
	}
}

func TestJaccardSets(t *testing.T) {
	t.Parallel()

	// Identical sets.
	a := map[string]bool{"x": true, "y": true}
	if j := jaccardSets(a, a); j != 1.0 {
		t.Errorf("identical sets: expected 1.0, got %f", j)
	}

	// Disjoint sets.
	b := map[string]bool{"z": true}
	if j := jaccardSets(a, b); j != 0.0 {
		t.Errorf("disjoint sets: expected 0.0, got %f", j)
	}

	// Partial overlap.
	c := map[string]bool{"x": true, "z": true}
	j := jaccardSets(a, c)
	if j < 0.3 || j > 0.4 {
		t.Errorf("partial overlap: expected ~0.33, got %f", j)
	}

	// Both empty — no evidence of similarity.
	if j := jaccardSets(nil, nil); j != 0.0 {
		t.Errorf("both empty: expected 0.0, got %f", j)
	}
}

func TestLCSRatio(t *testing.T) {
	t.Parallel()

	// Identical.
	if r := lcsRatio([]string{"a", "b", "c"}, []string{"a", "b", "c"}); r != 1.0 {
		t.Errorf("identical: expected 1.0, got %f", r)
	}

	// Completely different.
	if r := lcsRatio([]string{"a"}, []string{"b"}); r != 0.0 {
		t.Errorf("different: expected 0.0, got %f", r)
	}

	// Partial match.
	r := lcsRatio([]string{"a", "b", "c"}, []string{"a", "x", "c"})
	if r < 0.6 || r > 0.7 {
		t.Errorf("partial match: expected ~0.67, got %f", r)
	}

	// Both empty.
	if r := lcsRatio(nil, nil); r != 1.0 {
		t.Errorf("both empty: expected 1.0, got %f", r)
	}
}

func TestNormalizeTestName(t *testing.T) {
	t.Parallel()
	result := normalizeTestName("should handle login errors")
	if result == "" {
		t.Error("expected non-empty normalized name")
	}
	// "should" should be stripped.
	for _, tok := range []string{"should"} {
		if contains(result, tok) {
			t.Errorf("expected %q to be stripped from %q", tok, result)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsWord(s, sub))
}

func containsWord(s, word string) bool {
	for _, w := range splitWords(s) {
		if w == word {
			return true
		}
	}
	return false
}

func splitWords(s string) []string {
	var words []string
	current := ""
	for _, r := range s {
		if r == ' ' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}

// --- Impact Tests ---

func TestAnalyzeImpact_BasicChange(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	result := AnalyzeImpact(g, []string{"src/auth.js"})

	if len(result.ChangedFiles) != 1 {
		t.Errorf("expected 1 changed file, got %d", len(result.ChangedFiles))
	}
	if result.ChangedFiles[0] != "src/auth.js" {
		t.Errorf("expected src/auth.js, got %s", result.ChangedFiles[0])
	}

	// auth.js is imported by both test files → should impact all 4 tests.
	if len(result.Tests) < 2 {
		t.Errorf("expected ≥2 impacted tests, got %d", len(result.Tests))
	}

	// Verify all tests have valid confidence.
	for _, test := range result.Tests {
		if test.Confidence <= 0 || test.Confidence > 1 {
			t.Errorf("test %s confidence out of range: %f", test.TestID, test.Confidence)
		}
		if test.Level != "high" && test.Level != "medium" && test.Level != "low" {
			t.Errorf("test %s invalid level: %s", test.TestID, test.Level)
		}
	}

	// Level counts should sum to total tests.
	total := result.LevelCounts["high"] + result.LevelCounts["medium"] + result.LevelCounts["low"]
	if total != len(result.Tests) {
		t.Errorf("level counts (%d) don't sum to test count (%d)", total, len(result.Tests))
	}
}

func TestAnalyzeImpact_MultiFileChange(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	single := AnalyzeImpact(g, []string{"src/util.js"})
	multi := AnalyzeImpact(g, []string{"src/auth.js", "src/util.js"})

	// Multi-file change should impact at least as many tests as single.
	if len(multi.Tests) < len(single.Tests) {
		t.Errorf("multi-file change impacted fewer tests (%d) than single (%d)",
			len(multi.Tests), len(single.Tests))
	}
}

func TestAnalyzeImpact_UnknownFile(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	result := AnalyzeImpact(g, []string{"nonexistent.js"})

	if len(result.Tests) != 0 {
		t.Errorf("expected 0 impacted tests for unknown file, got %d", len(result.Tests))
	}
}

func TestAnalyzeImpact_EmptyFiles(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	result := AnalyzeImpact(g, nil)

	if len(result.Tests) != 0 {
		t.Errorf("expected 0 impacted tests for nil files, got %d", len(result.Tests))
	}
}

func TestAnalyzeImpact_ReasonChain(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	result := AnalyzeImpact(g, []string{"src/auth.js"})

	// Tests directly importing auth.js should have short reason chains.
	for _, test := range result.Tests {
		if len(test.ReasonChain) == 0 {
			t.Errorf("test %s has empty reason chain", test.TestID)
		}
	}
}

// --- Determinism tests for parallelized operations ---

func TestDetectDuplicates_Deterministic(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	r1 := DetectDuplicates(g)
	r2 := DetectDuplicates(g)

	if r1.TestsAnalyzed != r2.TestsAnalyzed {
		t.Fatalf("non-deterministic: tests analyzed %d vs %d", r1.TestsAnalyzed, r2.TestsAnalyzed)
	}
	if len(r1.Clusters) != len(r2.Clusters) {
		t.Fatalf("non-deterministic: clusters %d vs %d", len(r1.Clusters), len(r2.Clusters))
	}
	for i := range r1.Clusters {
		if r1.Clusters[i].Similarity != r2.Clusters[i].Similarity {
			t.Errorf("non-deterministic similarity at cluster %d: %f vs %f",
				i, r1.Clusters[i].Similarity, r2.Clusters[i].Similarity)
		}
	}
}

func TestAnalyzeCoverage_Deterministic(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	r1 := AnalyzeCoverage(g)
	r2 := AnalyzeCoverage(g)

	if r1.SourceCount != r2.SourceCount {
		t.Fatalf("non-deterministic source count: %d vs %d", r1.SourceCount, r2.SourceCount)
	}
	for i := range r1.Sources {
		if r1.Sources[i].SourceID != r2.Sources[i].SourceID {
			t.Errorf("non-deterministic source ordering at %d: %s vs %s",
				i, r1.Sources[i].SourceID, r2.Sources[i].SourceID)
		}
		if r1.Sources[i].TestCount != r2.Sources[i].TestCount {
			t.Errorf("non-deterministic test count for %s: %d vs %d",
				r1.Sources[i].SourceID, r1.Sources[i].TestCount, r2.Sources[i].TestCount)
		}
	}
}

func TestAnalyzeImpact_Deterministic(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()
	changed := []string{"src/auth.js"}

	r1 := AnalyzeImpact(g, changed)
	r2 := AnalyzeImpact(g, changed)

	if len(r1.Tests) != len(r2.Tests) {
		t.Fatalf("non-deterministic test count: %d vs %d", len(r1.Tests), len(r2.Tests))
	}
	for i := range r1.Tests {
		if r1.Tests[i].TestID != r2.Tests[i].TestID {
			t.Errorf("non-deterministic test ordering at %d: %s vs %s",
				i, r1.Tests[i].TestID, r2.Tests[i].TestID)
		}
		if r1.Tests[i].Confidence != r2.Tests[i].Confidence {
			t.Errorf("non-deterministic confidence for %s: %f vs %f",
				r1.Tests[i].TestID, r1.Tests[i].Confidence, r2.Tests[i].Confidence)
		}
	}
}

func TestAnalyzeFanout_Deterministic(t *testing.T) {
	t.Parallel()
	g := buildTestGraph()

	r1 := AnalyzeFanout(g, DefaultFanoutThreshold)
	r2 := AnalyzeFanout(g, DefaultFanoutThreshold)

	if r1.NodeCount != r2.NodeCount {
		t.Fatalf("non-deterministic node count: %d vs %d", r1.NodeCount, r2.NodeCount)
	}
	for i := range r1.Entries {
		if r1.Entries[i].NodeID != r2.Entries[i].NodeID {
			t.Errorf("non-deterministic fanout ordering at %d: %s vs %s",
				i, r1.Entries[i].NodeID, r2.Entries[i].NodeID)
		}
	}
}
