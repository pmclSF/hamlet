package insights

import (
	"testing"

	"github.com/pmclSF/terrain/internal/depgraph"
	"github.com/pmclSF/terrain/internal/models"
)

func TestBuild_EmptySnapshot(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{}
	input := &BuildInput{
		Snapshot: snap,
		Coverage: depgraph.CoverageResult{BandCounts: map[depgraph.CoverageBand]int{}},
	}

	r := Build(input)

	if r.HealthGrade != "A" {
		t.Errorf("expected health grade A for empty snapshot, got %s", r.HealthGrade)
	}
	if len(r.Findings) != 0 {
		t.Errorf("expected 0 findings for empty snapshot, got %d", len(r.Findings))
	}
	if r.Headline == "" {
		t.Error("expected non-empty headline")
	}
}

func TestBuild_DuplicateFindings(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{}
	input := &BuildInput{
		Snapshot: snap,
		Coverage: depgraph.CoverageResult{BandCounts: map[depgraph.CoverageBand]int{}},
		Duplicates: depgraph.DuplicateResult{
			DuplicateCount: 150,
			TestsAnalyzed:  500,
			Clusters: []depgraph.DuplicateCluster{
				{ID: 1, Tests: []string{"t1", "t2", "t3"}, Similarity: 0.85},
				{ID: 2, Tests: []string{"t4", "t5"}, Similarity: 0.72},
			},
		},
	}

	r := Build(input)

	found := false
	for _, f := range r.Findings {
		if f.Category == CategoryOptimization {
			found = true
			if f.Severity != SeverityHigh {
				t.Errorf("expected high severity for 150 duplicates, got %s", f.Severity)
			}
		}
	}
	if !found {
		t.Error("expected optimization finding for duplicates")
	}
}

func TestBuild_FanoutFindings(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{}
	input := &BuildInput{
		Snapshot: snap,
		Coverage: depgraph.CoverageResult{BandCounts: map[depgraph.CoverageBand]int{}},
		Fanout: depgraph.FanoutResult{
			FlaggedCount: 3,
			Threshold:    10,
			NodeCount:    100,
			Entries: []depgraph.FanoutEntry{
				{NodeID: "n1", Path: "fixtures/auth.js", NodeType: "helper", TransitiveFanout: 200, Flagged: true},
				{NodeID: "n2", Path: "fixtures/db.js", NodeType: "helper", TransitiveFanout: 50, Flagged: true},
			},
		},
	}

	r := Build(input)

	found := false
	for _, f := range r.Findings {
		if f.Category == CategoryArchitectureDebt {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected architecture debt finding for high fanout")
	}
}

func TestBuild_CoverageFindings(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{}
	input := &BuildInput{
		Snapshot: snap,
		Coverage: depgraph.CoverageResult{
			SourceCount: 100,
			BandCounts: map[depgraph.CoverageBand]int{
				depgraph.CoverageBandLow:    60,
				depgraph.CoverageBandMedium: 25,
				depgraph.CoverageBandHigh:   15,
			},
			Sources: []depgraph.SourceCoverage{
				{Path: "src/billing.js", TestCount: 0, Band: depgraph.CoverageBandLow},
			},
		},
	}

	r := Build(input)

	found := false
	for _, f := range r.Findings {
		if f.Category == CategoryCoverageDebt {
			found = true
			if f.Severity != SeverityHigh {
				t.Errorf("expected high severity for 60%% uncovered, got %s", f.Severity)
			}
		}
	}
	if !found {
		t.Error("expected coverage debt finding")
	}
}

func TestBuild_SkipFindings(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/a.test.js", TestCount: 100},
		},
		Signals: []models.Signal{
			{Type: "skippedTest", Severity: models.SeverityMedium},
			{Type: "skippedTest", Severity: models.SeverityMedium},
			{Type: "skippedTest", Severity: models.SeverityMedium},
			{Type: "skippedTest", Severity: models.SeverityMedium},
			{Type: "skippedTest", Severity: models.SeverityMedium},
		},
	}
	input := &BuildInput{
		Snapshot: snap,
		Coverage: depgraph.CoverageResult{BandCounts: map[depgraph.CoverageBand]int{}},
	}

	r := Build(input)

	found := false
	for _, f := range r.Findings {
		if f.Category == CategoryReliability && f.Metric == "5 skipped" {
			found = true
		}
	}
	if !found {
		t.Error("expected reliability finding for skipped tests")
	}
}

func TestBuild_StabilityFindings(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Signals: []models.Signal{
			{Type: "flakyTest", Severity: models.SeverityHigh},
			{Type: "flakyTest", Severity: models.SeverityHigh},
			{Type: "unstableSuite", Severity: models.SeverityHigh},
		},
	}
	input := &BuildInput{
		Snapshot: snap,
		Coverage: depgraph.CoverageResult{BandCounts: map[depgraph.CoverageBand]int{}},
	}

	r := Build(input)

	found := false
	for _, f := range r.Findings {
		if f.Category == CategoryReliability && f.Metric == "3 flaky signals" {
			found = true
		}
	}
	if !found {
		t.Error("expected reliability finding for flaky tests")
	}
}

func TestBuild_HealthGrade(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		signals  []models.Signal
		coverage depgraph.CoverageResult
		want     string
	}{
		{
			name: "grade A - no findings",
			want: "A",
			coverage: depgraph.CoverageResult{BandCounts: map[depgraph.CoverageBand]int{}},
		},
		{
			name: "grade D - critical signals",
			signals: []models.Signal{
				{Type: "flakyTest", Severity: models.SeverityCritical},
				{Type: "flakyTest", Severity: models.SeverityCritical},
			},
			coverage: depgraph.CoverageResult{BandCounts: map[depgraph.CoverageBand]int{}},
			want:     "D",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snap := &models.TestSuiteSnapshot{Signals: tt.signals}
			input := &BuildInput{
				Snapshot: snap,
				Coverage: tt.coverage,
			}
			r := Build(input)
			if r.HealthGrade != tt.want {
				t.Errorf("expected grade %s, got %s", tt.want, r.HealthGrade)
			}
		})
	}
}

func TestBuild_RecommendationsPrioritized(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/a.test.js", TestCount: 100},
		},
		Signals: []models.Signal{
			{Type: "skippedTest", Severity: models.SeverityMedium},
		},
	}
	input := &BuildInput{
		Snapshot: snap,
		Coverage: depgraph.CoverageResult{
			SourceCount: 10,
			BandCounts: map[depgraph.CoverageBand]int{
				depgraph.CoverageBandLow: 5,
			},
			Sources: []depgraph.SourceCoverage{
				{Path: "src/a.js", Band: depgraph.CoverageBandLow},
			},
		},
		Fanout: depgraph.FanoutResult{
			FlaggedCount: 2,
			Threshold:    10,
			NodeCount:    50,
			Entries: []depgraph.FanoutEntry{
				{Path: "helpers/auth.js", TransitiveFanout: 100, Flagged: true},
			},
		},
	}

	r := Build(input)

	if len(r.Recommendations) == 0 {
		t.Fatal("expected recommendations")
	}
	for i, rec := range r.Recommendations {
		if rec.Priority != i+1 {
			t.Errorf("recommendation %d has priority %d, expected %d", i, rec.Priority, i+1)
		}
	}
}

func TestBuild_Limitations(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{}
	input := &BuildInput{
		Snapshot: snap,
		Coverage: depgraph.CoverageResult{BandCounts: map[depgraph.CoverageBand]int{}},
	}

	r := Build(input)

	if len(r.Limitations) == 0 {
		t.Error("expected limitations for empty snapshot")
	}
}

func TestBuild_CategorySummary(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Signals: []models.Signal{
			{Type: "flakyTest", Severity: models.SeverityHigh},
		},
	}
	input := &BuildInput{
		Snapshot: snap,
		Coverage: depgraph.CoverageResult{
			SourceCount: 20,
			BandCounts:  map[depgraph.CoverageBand]int{depgraph.CoverageBandLow: 10},
			Sources: []depgraph.SourceCoverage{
				{Path: "src/a.js", Band: depgraph.CoverageBandLow},
			},
		},
	}

	r := Build(input)

	if len(r.CategorySummary) == 0 {
		t.Error("expected non-empty category summary")
	}
}

// --- AI Scenario Duplication ---

func TestBuild_ScenarioDuplication(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Scenarios: []models.Scenario{
			{
				ScenarioID:        "scenario:safety",
				Name:              "safety-check",
				CoveredSurfaceIDs: []string{"surface:prompts.ts:system", "surface:prompts.ts:user"},
			},
			{
				ScenarioID:        "scenario:accuracy",
				Name:              "accuracy-check",
				CoveredSurfaceIDs: []string{"surface:prompts.ts:system", "surface:prompts.ts:user"},
			},
			{
				ScenarioID:        "scenario:latency",
				Name:              "latency-check",
				CoveredSurfaceIDs: []string{"surface:api.ts:predict"},
			},
		},
	}
	input := &BuildInput{
		Snapshot: snap,
		Coverage: depgraph.CoverageResult{BandCounts: map[depgraph.CoverageBand]int{}},
	}

	r := Build(input)

	// safety and accuracy overlap 100% (2/2 shared surfaces).
	// latency has no overlap with either.
	found := false
	for _, f := range r.Findings {
		if f.Category == CategoryOptimization && f.Title != "" {
			if contains(f.Title, "scenario pair") {
				found = true
				if f.Severity != SeverityLow && f.Severity != SeverityMedium {
					t.Errorf("expected low or medium severity, got %s", f.Severity)
				}
			}
		}
	}
	if !found {
		t.Error("expected scenario duplication finding when 2 scenarios share >50% surfaces")
	}
}

func TestBuild_ScenarioDuplication_NoOverlap(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Scenarios: []models.Scenario{
			{ScenarioID: "scenario:a", CoveredSurfaceIDs: []string{"s1"}},
			{ScenarioID: "scenario:b", CoveredSurfaceIDs: []string{"s2"}},
		},
	}
	input := &BuildInput{
		Snapshot: snap,
		Coverage: depgraph.CoverageResult{BandCounts: map[depgraph.CoverageBand]int{}},
	}

	r := Build(input)

	for _, f := range r.Findings {
		if contains(f.Title, "scenario pair") {
			t.Error("should not report scenario duplication when no surfaces overlap")
		}
	}
}

func TestBuild_ScenarioDuplication_SingleScenario(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Scenarios: []models.Scenario{
			{ScenarioID: "scenario:only", CoveredSurfaceIDs: []string{"s1", "s2"}},
		},
	}
	input := &BuildInput{
		Snapshot: snap,
		Coverage: depgraph.CoverageResult{BandCounts: map[depgraph.CoverageBand]int{}},
	}

	r := Build(input)

	for _, f := range r.Findings {
		if contains(f.Title, "scenario pair") {
			t.Error("should not report scenario duplication with only 1 scenario")
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && searchString(s, substr)))
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
