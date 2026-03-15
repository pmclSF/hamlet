package explain

import (
	"testing"

	"github.com/pmclSF/terrain/internal/impact"
)

func TestExplainTest_NilResult(t *testing.T) {
	_, err := ExplainTest("test.js", nil)
	if err == nil {
		t.Fatal("expected error for nil result")
	}
}

func TestExplainTest_NotFound(t *testing.T) {
	result := &impact.ImpactResult{}
	_, err := ExplainTest("nonexistent.test.js", result)
	if err == nil {
		t.Fatal("expected error for missing test")
	}
}

func TestExplainSelection_NilResult(t *testing.T) {
	_, err := ExplainSelection(nil)
	if err == nil {
		t.Fatal("expected error for nil result")
	}
}

func TestExplainSelection_EmptyResult(t *testing.T) {
	result := &impact.ImpactResult{}
	sel, err := ExplainSelection(result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sel.TotalSelected != 0 {
		t.Errorf("expected 0 selected, got %d", sel.TotalSelected)
	}
	if sel.Strategy != "none" {
		t.Errorf("expected strategy 'none', got %q", sel.Strategy)
	}
}

func TestClassifyConfidenceBand(t *testing.T) {
	tests := []struct {
		conf float64
		want string
	}{
		{0.95, "high"},
		{0.70, "high"},
		{0.65, "medium"},
		{0.40, "medium"},
		{0.30, "low"},
		{0.0, "low"},
	}
	for _, tt := range tests {
		got := classifyConfidenceBand(tt.conf)
		if got != tt.want {
			t.Errorf("classifyConfidenceBand(%v) = %q, want %q", tt.conf, got, tt.want)
		}
	}
}

func TestConfidenceScore(t *testing.T) {
	tests := []struct {
		conf impact.Confidence
		want float64
	}{
		{impact.ConfidenceExact, 0.95},
		{impact.ConfidenceInferred, 0.65},
		{impact.ConfidenceWeak, 0.30},
	}
	for _, tt := range tests {
		got := confidenceScore(tt.conf)
		if got != tt.want {
			t.Errorf("confidenceScore(%q) = %v, want %v", tt.conf, got, tt.want)
		}
	}
}

func TestEdgeKindLabel(t *testing.T) {
	tests := []struct {
		kind impact.EdgeKind
		want string
	}{
		{impact.EdgeExactCoverage, "exact per-test coverage"},
		{impact.EdgeBucketCoverage, "file-level coverage link"},
		{impact.EdgeStructuralLink, "import/export dependency"},
		{impact.EdgeDirectoryProximity, "directory proximity"},
		{impact.EdgeNameConvention, "naming convention match"},
	}
	for _, tt := range tests {
		got := edgeKindLabel(tt.kind)
		if got != tt.want {
			t.Errorf("edgeKindLabel(%q) = %q, want %q", tt.kind, got, tt.want)
		}
	}
}

func TestClassifyReason(t *testing.T) {
	tests := []struct {
		name string
		test impact.ImpactedTest
		want string
	}{
		{"directly changed", impact.ImpactedTest{IsDirectlyChanged: true}, "directlyChanged"},
		{"exact confidence", impact.ImpactedTest{ImpactConfidence: impact.ConfidenceExact}, "directDependency"},
		{"directory proximity", impact.ImpactedTest{Relevance: "in same directory tree as changed code"}, "directoryProximity"},
		{"default", impact.ImpactedTest{ImpactConfidence: impact.ConfidenceInferred}, "fixtureDependency"},
	}
	for _, tt := range tests {
		got := classifyReason(&tt.test)
		if got != tt.want {
			t.Errorf("classifyReason(%s) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestFindTest_PartialMatch(t *testing.T) {
	result := &impact.ImpactResult{
		ImpactedTests: []impact.ImpactedTest{
			{Path: "test/integration/auth.test.js"},
		},
	}
	test, found := findTest("auth.test.js", result)
	if !found {
		t.Fatal("expected partial match to find test")
	}
	if test.Path != "test/integration/auth.test.js" {
		t.Errorf("got %q, want test/integration/auth.test.js", test.Path)
	}
}

func TestBuildVerdict_NoPath(t *testing.T) {
	te := &TestExplanation{
		Target:         TestTarget{Path: "test/foo.test.js"},
		ConfidenceBand: "low",
	}
	verdict := buildVerdict(te)
	if verdict == "" {
		t.Error("expected non-empty verdict")
	}
}
