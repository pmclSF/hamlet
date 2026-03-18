package analyze

import (
	"testing"

	"github.com/pmclSF/terrain/internal/depgraph"
	"github.com/pmclSF/terrain/internal/stability"
)

func TestDeriveKeyFindings_PrioritizesBySeverity(t *testing.T) {
	t.Parallel()
	r := &Report{
		// High-fanout: medium severity
		HighFanout: FanoutSummary{FlaggedCount: 2},
		// Skip burden: high severity (>10%)
		SkippedTestBurden: SkipSummary{SkippedCount: 20, TotalTests: 100},
		// Weak coverage will be critical (>75%)
	}

	fanout := &depgraph.FanoutResult{FlaggedCount: 2}
	dupes := &depgraph.DuplicateResult{}
	cov := &depgraph.CoverageResult{
		BandCounts:  map[depgraph.CoverageBand]int{depgraph.CoverageBandLow: 80},
		SourceCount: 100,
	}

	findings, total := deriveKeyFindings(r, fanout, dupes, cov, nil)

	if total < 3 {
		t.Fatalf("expected at least 3 findings, got %d", total)
	}
	if len(findings) > 3 {
		t.Errorf("expected at most 3 key findings, got %d", len(findings))
	}

	// First finding should be the most severe (critical: weak coverage at 80%).
	if findings[0].Severity != "critical" {
		t.Errorf("first finding should be critical, got %s: %s", findings[0].Severity, findings[0].Title)
	}
}

func TestDeriveKeyFindings_MaxThree(t *testing.T) {
	t.Parallel()
	r := &Report{
		HighFanout:        FanoutSummary{FlaggedCount: 10},
		SkippedTestBurden: SkipSummary{SkippedCount: 50, TotalTests: 200},
		SignalSummary:     SignalBreakdown{Critical: 3, Total: 3},
		StabilityClusters: &stability.ClusterResult{
			ClusteredTestCount: 20,
			Clusters:           make([]stability.Cluster, 5),
		},
	}

	fanout := &depgraph.FanoutResult{FlaggedCount: 10}
	dupes := &depgraph.DuplicateResult{DuplicateCount: 200, Clusters: make([]depgraph.DuplicateCluster, 15)}
	cov := &depgraph.CoverageResult{
		BandCounts:  map[depgraph.CoverageBand]int{depgraph.CoverageBandLow: 30},
		SourceCount: 100,
	}

	findings, total := deriveKeyFindings(r, fanout, dupes, cov, nil)

	if len(findings) != 3 {
		t.Errorf("expected exactly 3 key findings, got %d", len(findings))
	}
	if total <= 3 {
		t.Errorf("expected total > 3 (many inputs), got %d", total)
	}
}

func TestDeriveKeyFindings_EmptyReport(t *testing.T) {
	t.Parallel()
	r := &Report{}
	fanout := &depgraph.FanoutResult{}
	dupes := &depgraph.DuplicateResult{}
	cov := &depgraph.CoverageResult{BandCounts: map[depgraph.CoverageBand]int{}}

	findings, total := deriveKeyFindings(r, fanout, dupes, cov, nil)

	if len(findings) != 0 {
		t.Errorf("expected 0 findings for empty report, got %d", len(findings))
	}
	if total != 0 {
		t.Errorf("expected total 0, got %d", total)
	}
}

func TestDeriveKeyFindings_Deterministic(t *testing.T) {
	t.Parallel()
	r := &Report{
		HighFanout:        FanoutSummary{FlaggedCount: 3},
		SkippedTestBurden: SkipSummary{SkippedCount: 10, TotalTests: 100},
	}
	fanout := &depgraph.FanoutResult{FlaggedCount: 3}
	dupes := &depgraph.DuplicateResult{DuplicateCount: 50, Clusters: make([]depgraph.DuplicateCluster, 5)}
	cov := &depgraph.CoverageResult{
		BandCounts:  map[depgraph.CoverageBand]int{depgraph.CoverageBandLow: 10},
		SourceCount: 100,
	}

	f1, t1 := deriveKeyFindings(r, fanout, dupes, cov, nil)
	f2, t2 := deriveKeyFindings(r, fanout, dupes, cov, nil)

	if t1 != t2 {
		t.Errorf("total not deterministic: %d vs %d", t1, t2)
	}
	if len(f1) != len(f2) {
		t.Fatalf("finding count not deterministic: %d vs %d", len(f1), len(f2))
	}
	for i := range f1 {
		if f1[i].Title != f2[i].Title {
			t.Errorf("finding[%d] title not deterministic: %q vs %q", i, f1[i].Title, f2[i].Title)
		}
		if f1[i].Severity != f2[i].Severity {
			t.Errorf("finding[%d] severity not deterministic: %q vs %q", i, f1[i].Severity, f2[i].Severity)
		}
	}
}

func TestDeriveKeyFindings_CategoryOrder(t *testing.T) {
	t.Parallel()
	// Two findings with same severity (medium) — reliability should rank before optimization.
	r := &Report{
		SkippedTestBurden: SkipSummary{SkippedCount: 5, TotalTests: 100}, // medium reliability
	}
	fanout := &depgraph.FanoutResult{}
	dupes := &depgraph.DuplicateResult{DuplicateCount: 30, Clusters: make([]depgraph.DuplicateCluster, 3)} // medium optimization
	cov := &depgraph.CoverageResult{BandCounts: map[depgraph.CoverageBand]int{}}

	findings, _ := deriveKeyFindings(r, fanout, dupes, cov, nil)

	if len(findings) < 2 {
		t.Fatalf("expected at least 2 findings, got %d", len(findings))
	}

	// Skip burden is reliability (category order 1), duplicates is optimization (order 4).
	// With same severity, reliability should come first.
	if findings[0].Category != "reliability" {
		t.Errorf("first finding should be reliability, got %s: %s", findings[0].Category, findings[0].Title)
	}
}

func TestDeriveKeyFindings_CriticalSignalsRankFirst(t *testing.T) {
	t.Parallel()
	r := &Report{
		SignalSummary:     SignalBreakdown{Critical: 2, Total: 5},
		SkippedTestBurden: SkipSummary{SkippedCount: 5, TotalTests: 100},
	}
	fanout := &depgraph.FanoutResult{}
	dupes := &depgraph.DuplicateResult{}
	cov := &depgraph.CoverageResult{BandCounts: map[depgraph.CoverageBand]int{}}

	findings, _ := deriveKeyFindings(r, fanout, dupes, cov, nil)

	if len(findings) == 0 {
		t.Fatal("expected findings")
	}
	if findings[0].Severity != "critical" {
		t.Errorf("critical signals should rank first, got %s: %s", findings[0].Severity, findings[0].Title)
	}
}

func TestKeyFinding_HasAllFields(t *testing.T) {
	t.Parallel()
	r := &Report{
		HighFanout: FanoutSummary{FlaggedCount: 5},
	}
	fanout := &depgraph.FanoutResult{FlaggedCount: 5}
	dupes := &depgraph.DuplicateResult{}
	cov := &depgraph.CoverageResult{BandCounts: map[depgraph.CoverageBand]int{}}

	findings, _ := deriveKeyFindings(r, fanout, dupes, cov, nil)
	if len(findings) == 0 {
		t.Fatal("expected at least 1 finding")
	}
	f := findings[0]
	if f.Title == "" {
		t.Error("finding missing Title")
	}
	if f.Severity == "" {
		t.Error("finding missing Severity")
	}
	if f.Category == "" {
		t.Error("finding missing Category")
	}
	if f.Metric == "" {
		t.Error("finding missing Metric")
	}
}
