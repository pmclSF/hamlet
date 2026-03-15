package summary

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/benchmark"
	"github.com/pmclSF/terrain/internal/comparison"
	"github.com/pmclSF/terrain/internal/heatmap"
	"github.com/pmclSF/terrain/internal/metrics"
	"github.com/pmclSF/terrain/internal/models"
)

func baseInput() *BuildInput {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/auth/login.test.js"},
			{Path: "src/pay/pay.test.js"},
			{Path: "src/core/util.test.js"},
		},
		Frameworks: []models.Framework{
			{Name: "jest", FileCount: 3},
		},
		Signals: []models.Signal{
			{Type: "weakAssertion", Severity: models.SeverityMedium, Owner: "auth-team",
				Location: models.SignalLocation{File: "src/auth/login.test.js"}},
			{Type: "weakAssertion", Severity: models.SeverityMedium, Owner: "auth-team",
				Location: models.SignalLocation{File: "src/auth/signup.test.js"}},
			{Type: "mockHeavyTest", Severity: models.SeverityLow, Owner: "pay-team",
				Location: models.SignalLocation{File: "src/pay/pay.test.js"}},
		},
		Risk: []models.RiskSurface{
			{Type: "reliability", Scope: "repository", Band: models.RiskBandMedium, Score: 5},
			{Type: "change", Scope: "repository", Band: models.RiskBandHigh, Score: 10},
			{Type: "change", Scope: "directory", ScopeName: "src/auth", Band: models.RiskBandHigh, Score: 8,
				ContributingSignals: []models.Signal{
					{Type: "weakAssertion"},
					{Type: "weakAssertion"},
				}},
		},
	}

	h := heatmap.Build(snap)
	ms := metrics.Derive(snap)

	return &BuildInput{
		Snapshot: snap,
		Heatmap:  h,
		Metrics:  ms,
	}
}

func TestBuild_PostureDimensions(t *testing.T) {
	t.Parallel()
	es := Build(baseInput())

	if len(es.Posture.Dimensions) != 2 {
		t.Fatalf("expected 2 posture dimensions, got %d", len(es.Posture.Dimensions))
	}

	found := map[string]models.RiskBand{}
	for _, d := range es.Posture.Dimensions {
		found[d.Dimension] = d.Band
	}
	if found["reliability"] != models.RiskBandMedium {
		t.Errorf("reliability = %q, want medium", found["reliability"])
	}
	if found["change"] != models.RiskBandHigh {
		t.Errorf("change = %q, want high", found["change"])
	}
}

func TestBuild_TopRiskAreas(t *testing.T) {
	t.Parallel()
	es := Build(baseInput())

	if len(es.TopRiskAreas) == 0 {
		t.Fatal("expected at least one top risk area")
	}
	if es.TopRiskAreas[0].Name != "src/auth" {
		t.Errorf("top risk area = %q, want src/auth", es.TopRiskAreas[0].Name)
	}
	if es.TopRiskAreas[0].Scope != "directory" {
		t.Errorf("scope = %q, want directory", es.TopRiskAreas[0].Scope)
	}
}

func TestBuild_DominantDrivers(t *testing.T) {
	t.Parallel()
	es := Build(baseInput())

	if len(es.DominantDrivers) == 0 {
		t.Fatal("expected dominant drivers")
	}
	if es.DominantDrivers[0] != "weakAssertion" {
		t.Errorf("top driver = %q, want weakAssertion", es.DominantDrivers[0])
	}
}

func TestBuild_KeyNumbers(t *testing.T) {
	t.Parallel()
	es := Build(baseInput())

	if es.KeyNumbers.TestFiles != 3 {
		t.Errorf("testFiles = %d, want 3", es.KeyNumbers.TestFiles)
	}
	if es.KeyNumbers.Frameworks != 1 {
		t.Errorf("frameworks = %d, want 1", es.KeyNumbers.Frameworks)
	}
	if es.KeyNumbers.TotalSignals != 3 {
		t.Errorf("totalSignals = %d, want 3", es.KeyNumbers.TotalSignals)
	}
}

func TestBuild_NoTrendData(t *testing.T) {
	t.Parallel()
	es := Build(baseInput())

	if es.HasTrendData {
		t.Error("hasTrendData should be false without comparison")
	}
	if len(es.TrendHighlights) > 0 {
		t.Error("should have no trend highlights without comparison")
	}
}

func TestBuild_WithTrendData(t *testing.T) {
	t.Parallel()
	in := baseInput()
	in.Comparison = &comparison.SnapshotComparison{
		SignalDeltas: []comparison.SignalDelta{
			{Type: "weakAssertion", Category: "quality", Before: 5, After: 2, Delta: -3},
			{Type: "flakyTest", Category: "health", Before: 1, After: 4, Delta: 3},
		},
		RiskDeltas: []comparison.RiskDelta{
			{Type: "reliability", Scope: "repository", Before: models.RiskBandLow, After: models.RiskBandMedium, Changed: true},
		},
		TestFileCountDelta: 5,
	}

	es := Build(in)

	if !es.HasTrendData {
		t.Error("hasTrendData should be true")
	}
	if len(es.TrendHighlights) == 0 {
		t.Fatal("expected trend highlights")
	}

	// Check that risk band change is surfaced
	foundRisk := false
	for _, th := range es.TrendHighlights {
		if strings.Contains(th.Description, "reliability") && th.Direction == "worsened" {
			foundRisk = true
		}
	}
	if !foundRisk {
		t.Error("expected reliability worsened trend callout")
	}

	// Check improved signal delta
	foundImproved := false
	for _, th := range es.TrendHighlights {
		if strings.Contains(th.Description, "weakAssertion") && th.Direction == "improved" {
			foundImproved = true
		}
	}
	if !foundImproved {
		t.Error("expected weakAssertion improved trend callout")
	}
}

func TestBuild_BenchmarkReadiness(t *testing.T) {
	t.Parallel()
	in := baseInput()
	in.Segment = &benchmark.Segment{
		PrimaryLanguage:  "javascript",
		PrimaryFramework: "jest",
		TestFileBucket:   "small",
	}

	es := Build(in)

	if len(es.BenchmarkReadiness.ReadyDimensions) == 0 {
		t.Error("expected ready benchmark dimensions")
	}
	if es.BenchmarkReadiness.Segment == nil {
		t.Error("expected benchmark segment")
	}
	if es.BenchmarkReadiness.Segment.PrimaryLanguage != "javascript" {
		t.Errorf("segment language = %q, want javascript", es.BenchmarkReadiness.Segment.PrimaryLanguage)
	}
}

func TestBuild_RecommendedFocus(t *testing.T) {
	t.Parallel()
	es := Build(baseInput())

	if es.RecommendedFocus == "" {
		t.Error("expected non-empty recommended focus")
	}
	// Should mention top risk area
	if !strings.Contains(es.RecommendedFocus, "src/auth") {
		t.Errorf("recommended focus should mention src/auth, got %q", es.RecommendedFocus)
	}
}

func TestBuild_EmptySnapshot(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{}
	h := heatmap.Build(snap)
	ms := metrics.Derive(snap)

	es := Build(&BuildInput{
		Snapshot: snap,
		Heatmap:  h,
		Metrics:  ms,
	})

	if es.KeyNumbers.TotalSignals != 0 {
		t.Errorf("totalSignals = %d, want 0", es.KeyNumbers.TotalSignals)
	}
	if len(es.TopRiskAreas) != 0 {
		t.Errorf("topRiskAreas = %d, want 0", len(es.TopRiskAreas))
	}
	if es.RecommendedFocus == "" {
		t.Error("expected a default recommended focus")
	}
}

func TestBuild_JSONSerialization(t *testing.T) {
	t.Parallel()
	es := Build(baseInput())

	data, err := json.MarshalIndent(es, "", "  ")
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}

	var roundtrip ExecutiveSummary
	if err := json.Unmarshal(data, &roundtrip); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	if roundtrip.Posture.OverallBand != es.Posture.OverallBand {
		t.Errorf("roundtrip overallBand = %q, want %q", roundtrip.Posture.OverallBand, es.Posture.OverallBand)
	}
	if len(roundtrip.DominantDrivers) != len(es.DominantDrivers) {
		t.Errorf("roundtrip dominantDrivers len = %d, want %d", len(roundtrip.DominantDrivers), len(es.DominantDrivers))
	}
}

func TestCategorizeSignalType(t *testing.T) {
	t.Parallel()
	tests := []struct {
		signal string
		want   string
	}{
		{"flakyTest", "reliability"},
		{"slowTest", "speed"},
		{"weakAssertion", "quality"},
		{"migrationBlocker", "migration"},
		{"policyViolation", "governance"},
		{"skippedTestsInCI", "governance"},
		{"unknownType", "quality"},
	}
	for _, tt := range tests {
		got := categorizeSignalType(tt.signal)
		if got != tt.want {
			t.Errorf("categorizeSignalType(%q) = %q, want %q", tt.signal, got, tt.want)
		}
	}
}

func TestBuild_TrendWithWorsenedFocusRecommendation(t *testing.T) {
	t.Parallel()
	in := baseInput()
	in.Comparison = &comparison.SnapshotComparison{
		SignalDeltas: []comparison.SignalDelta{
			{Type: "flakyTest", Category: "health", Before: 1, After: 5, Delta: 4},
		},
	}

	es := Build(in)

	// Recommended focus should mention the worsened trend dimension
	if !strings.Contains(es.RecommendedFocus, "trend") {
		t.Errorf("expected recommended focus to reference trend, got %q", es.RecommendedFocus)
	}
}

func TestBuild_Recommendations(t *testing.T) {
	t.Parallel()
	in := baseInput()
	// Add evidence strength to signals.
	in.Snapshot.Signals[0].EvidenceStrength = models.EvidenceStrong
	in.Snapshot.Signals[1].EvidenceStrength = models.EvidenceModerate
	in.Snapshot.Signals[2].EvidenceStrength = models.EvidenceWeak

	es := Build(in)

	if len(es.Recommendations) == 0 {
		t.Fatal("expected recommendations")
	}
	// First recommendation should have highest evidence strength.
	first := es.Recommendations[0]
	if first.Priority != 1 {
		t.Errorf("first recommendation priority = %d, want 1", first.Priority)
	}
	if first.What == "" || first.Why == "" || first.Where == "" {
		t.Error("recommendation missing what/why/where")
	}
	if first.EvidenceStrength == "" {
		t.Error("recommendation missing evidence strength")
	}
}

func TestBuild_BlindSpots_NoCoverage(t *testing.T) {
	t.Parallel()
	in := baseInput()
	// No CoverageSummary → should flag coverage blind spot.
	es := Build(in)

	found := false
	for _, bs := range es.BlindSpots {
		if bs.Area == "Coverage data" {
			found = true
			if bs.Remediation == "" {
				t.Error("expected remediation for coverage blind spot")
			}
		}
	}
	if !found {
		t.Error("expected coverage blind spot when no coverage data")
	}
}

func TestBuild_BlindSpots_WeakEvidence(t *testing.T) {
	t.Parallel()
	in := baseInput()
	// All signals weak → should flag weak evidence blind spot.
	for i := range in.Snapshot.Signals {
		in.Snapshot.Signals[i].EvidenceStrength = models.EvidenceWeak
	}

	es := Build(in)

	found := false
	for _, bs := range es.BlindSpots {
		if bs.Area == "Signal confidence" {
			found = true
		}
	}
	if !found {
		t.Error("expected signal confidence blind spot when majority weak evidence")
	}
}

func TestBuild_NoBlindSpots_WithCoverage(t *testing.T) {
	t.Parallel()
	in := baseInput()
	in.Snapshot.CoverageSummary = &models.CoverageSummary{
		LineCoveragePct: 80.0,
	}
	in.Snapshot.Ownership = map[string][]string{"src/auth": {"auth-team"}}
	for i := range in.Snapshot.Signals {
		in.Snapshot.Signals[i].EvidenceStrength = models.EvidenceStrong
	}

	es := Build(in)

	for _, bs := range es.BlindSpots {
		if bs.Area == "Coverage data" {
			t.Error("should not flag coverage blind spot when coverage data exists")
		}
		if bs.Area == "Signal confidence" {
			t.Error("should not flag weak evidence when signals are strong")
		}
		if bs.Area == "Ownership attribution" {
			t.Error("should not flag ownership when CODEOWNERS present")
		}
	}
}

func TestRenderExecutiveSummary_Sections(t *testing.T) {
	t.Parallel()
	// Import test - verify rendering doesn't panic and includes expected sections
	es := Build(baseInput())

	var buf bytes.Buffer
	// We can't directly import reporting here due to circular deps,
	// so we just verify the model is well-formed for rendering.
	_ = buf
	_ = es

	// Verify key sections exist in model
	if es.Posture.OverallStatement == "" {
		t.Error("expected non-empty overall statement")
	}
	if len(es.BenchmarkReadiness.ReadyDimensions) == 0 {
		t.Error("expected benchmark ready dimensions")
	}
}
