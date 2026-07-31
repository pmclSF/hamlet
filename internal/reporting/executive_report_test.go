package reporting

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/benchmark"
	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/summary"
)

func TestRenderExecutiveSummary_AllSections(t *testing.T) {
	t.Parallel()
	es := &summary.ExecutiveSummary{
		Posture: summary.PostureSummary{
			OverallBand:      models.RiskBandHigh,
			OverallStatement: "High risk detected.",
			Dimensions: []summary.DimensionPosture{
				{Dimension: "reliability", Band: models.RiskBandMedium},
				{Dimension: "change", Band: models.RiskBandHigh},
			},
		},
		TopRiskAreas: []summary.FocusArea{
			{Name: "src/auth", Scope: "directory", Band: models.RiskBandHigh, RiskType: "quality", SignalCount: 5},
		},
		TrendHighlights: []summary.TrendCallout{
			{Description: "Weak Assertion findings decreased (-3)", Direction: "improved"},
			{Description: "Flaky Test findings increased (+2)", Direction: "worsened"},
		},
		HasTrendData:     true,
		DominantDrivers:  []string{"weakAssertion", "mockHeavyTest"},
		RecommendedFocus: "Start with: Weak Assertion — 2 findings › terrain explain weakAssertion",
		BenchmarkReadiness: summary.BenchmarkReadinessSummary{
			ReadyDimensions: []string{"test structure", "quality metrics"},
			LimitedDimensions: []summary.BenchmarkLimitation{
				{Dimension: "speed comparison", Reason: "runtime data is partial"},
			},
			Segment: &benchmark.Segment{
				PrimaryLanguage:  "javascript",
				PrimaryFramework: "jest",
				TestFileBucket:   "small",
			},
		},
		KeyNumbers: summary.KeyNumbers{
			TestFiles:        10,
			Frameworks:       2,
			TotalSignals:     15,
			CriticalFindings: 1,
			HighRiskAreas:    3,
		},
	}

	var buf bytes.Buffer
	RenderExecutiveSummary(&buf, es)
	output := buf.String()

	expected := []string{
		"Terrain · Executive Summary",
		"Overall Posture",
		"reliability:",
		"change:",
		"Key Numbers",
		"Test files:",
		"Top Risk Areas",
		"src/auth",
		"Trend Highlights",
		"Weak Assertion findings decreased",
		"Flaky Test findings increased",
		"Dominant Drivers",
		"Weak Assertion",
		"Mock-Heavy Test",
		"Recommended Focus",
		"Start with: Weak Assertion",
		"Benchmark Readiness",
		"test structure",
		"speed comparison",
	}

	for _, s := range expected {
		if !strings.Contains(output, s) {
			t.Errorf("output missing %q", s)
		}
	}
}

func TestRenderExecutiveSummary_NoTrendData(t *testing.T) {
	t.Parallel()
	es := &summary.ExecutiveSummary{
		Posture: summary.PostureSummary{
			OverallBand:      models.RiskBandLow,
			OverallStatement: "Low risk.",
		},
		HasTrendData: false,
		BenchmarkReadiness: summary.BenchmarkReadinessSummary{
			ReadyDimensions: []string{"test structure"},
		},
		KeyNumbers: summary.KeyNumbers{TestFiles: 5},
	}

	var buf bytes.Buffer
	RenderExecutiveSummary(&buf, es)
	output := buf.String()

	if !strings.Contains(output, "first analysis") {
		t.Error("expected first-analysis baseline message")
	}
	if !strings.Contains(output, "write-snapshot") {
		t.Error("expected hint about write-snapshot")
	}
}

func TestRenderExecutiveSummary_Empty(t *testing.T) {
	t.Parallel()
	es := &summary.ExecutiveSummary{
		Posture: summary.PostureSummary{
			OverallBand:      models.RiskBandLow,
			OverallStatement: "Clean.",
		},
		BenchmarkReadiness: summary.BenchmarkReadinessSummary{
			ReadyDimensions: []string{"test structure"},
		},
	}

	var buf bytes.Buffer
	RenderExecutiveSummary(&buf, es)
	output := buf.String()

	if !strings.Contains(output, "Terrain · Executive Summary") {
		t.Error("expected header")
	}
	// Should not have empty sections crashing
	if strings.Contains(output, "Top Risk Areas") {
		t.Error("should not show Top Risk Areas when empty")
	}
}

func TestRenderExecutiveSummary_TrendDirectionIcons(t *testing.T) {
	t.Parallel()
	es := &summary.ExecutiveSummary{
		Posture: summary.PostureSummary{
			OverallBand:      models.RiskBandMedium,
			OverallStatement: "Moderate.",
		},
		TrendHighlights: []summary.TrendCallout{
			{Description: "improved thing", Direction: "improved"},
			{Description: "worsened thing", Direction: "worsened"},
		},
		HasTrendData: true,
		BenchmarkReadiness: summary.BenchmarkReadinessSummary{
			ReadyDimensions: []string{"test structure"},
		},
	}

	var buf bytes.Buffer
	RenderExecutiveSummary(&buf, es)
	output := buf.String()

	if !strings.Contains(output, "↓ improved thing") {
		t.Error("expected down arrow for improved")
	}
	if !strings.Contains(output, "↑ worsened thing") {
		t.Error("expected up arrow for worsened")
	}
}

func TestRenderExecutiveSummary_Recommendations(t *testing.T) {
	t.Parallel()
	es := &summary.ExecutiveSummary{
		Posture: summary.PostureSummary{
			OverallBand:      models.RiskBandMedium,
			OverallStatement: "Moderate.",
		},
		Recommendations: []summary.Recommendation{
			{
				What:             "Reduce quality findings in src/auth (5 signals)",
				Why:              "High risk band with strong-confidence evidence",
				Where:            "src/auth",
				EvidenceStrength: "strong",
				Priority:         1,
			},
			{
				What:             "Reduce reliability findings in src/pay (2 signals)",
				Why:              "Medium risk band with weak-confidence evidence",
				Where:            "src/pay",
				EvidenceStrength: "weak",
				Priority:         2,
			},
		},
		BenchmarkReadiness: summary.BenchmarkReadinessSummary{
			ReadyDimensions: []string{"test structure"},
		},
	}

	var buf bytes.Buffer
	RenderExecutiveSummary(&buf, es)
	output := buf.String()

	expected := []string{
		"Prioritized Recommendations",
		"1. Reduce quality findings in src/auth",
		"Why:",
		"Where:    src/auth",
		"Evidence: strong",
		"2. Reduce reliability findings in src/pay",
		"Evidence: weak",
	}
	for _, s := range expected {
		if !strings.Contains(output, s) {
			t.Errorf("output missing %q", s)
		}
	}
}

func TestRenderExecutiveSummary_UnmeasuredDimensionRow(t *testing.T) {
	t.Parallel()
	es := &summary.ExecutiveSummary{
		Posture: summary.PostureSummary{
			OverallBand:      models.RiskBandLow,
			OverallStatement: "Mixed.",
			Dimensions: []summary.DimensionPosture{
				{Dimension: "coverage_depth", Band: "strong"},
				{Dimension: "health", Band: "unknown", NeedsInput: "runtime data"},
				{Dimension: "structural_risk", Band: "unknown"},
			},
		},
		BenchmarkReadiness: summary.BenchmarkReadinessSummary{
			ReadyDimensions: []string{"test structure"},
		},
	}

	var buf bytes.Buffer
	RenderExecutiveSummary(&buf, es)
	output := buf.String()

	if !strings.Contains(output, "not yet measured — needs runtime data") {
		t.Error("expected unmeasured row with the missing input named")
	}
	if !strings.Contains(output, "Structural risk:") || !strings.Contains(output, "not yet measured\n") {
		t.Error("expected plain 'not yet measured' when the missing input is unknown")
	}
	if strings.Contains(output, "Unknown") {
		t.Error("unmeasured rows must not render as 'Unknown'")
	}
	// Measured rows keep the band display and the legend stays.
	if !strings.Contains(output, "Coverage depth:") || !strings.Contains(output, "Strong") {
		t.Error("expected measured row to keep its band")
	}
	if !strings.Contains(output, "Dimension meaning:") {
		t.Error("expected dimension legend when some rows are measured")
	}
}

func TestRenderExecutiveSummary_AllDimensionsUnmeasured(t *testing.T) {
	t.Parallel()
	es := &summary.ExecutiveSummary{
		Posture: summary.PostureSummary{
			OverallBand:      models.RiskBandLow,
			OverallStatement: "Unmeasured.",
			Dimensions: []summary.DimensionPosture{
				{Dimension: "health", Band: "unknown", NeedsInput: "test files"},
				{Dimension: "coverage_depth", Band: "unknown", NeedsInput: "test files"},
				{Dimension: "coverage_diversity", Band: "unknown", NeedsInput: "test files"},
				{Dimension: "structural_risk", Band: "unknown", NeedsInput: "test files"},
			},
		},
		BenchmarkReadiness: summary.BenchmarkReadinessSummary{
			ReadyDimensions: []string{"test structure"},
		},
	}

	var buf bytes.Buffer
	RenderExecutiveSummary(&buf, es)
	output := buf.String()

	if !strings.Contains(output, "Not yet measured — needs test files.") {
		t.Error("expected single collapsed unmeasured line")
	}
	if got := strings.Count(output, "measured"); got != 1 {
		t.Errorf("unmeasured state should be stated once, got %d mentions", got)
	}
	if strings.Contains(output, "Unknown") {
		t.Error("unmeasured posture must not render as 'Unknown'")
	}
	if strings.Contains(output, "Dimension meaning:") {
		t.Error("legend should be omitted when nothing is measured")
	}
}

func TestRenderExecutiveSummary_BlindSpots(t *testing.T) {
	t.Parallel()
	es := &summary.ExecutiveSummary{
		Posture: summary.PostureSummary{
			OverallBand:      models.RiskBandLow,
			OverallStatement: "Low.",
		},
		BlindSpots: []summary.BlindSpot{
			{Area: "Coverage data", Reason: "No coverage artifacts were ingested", Remediation: "Run with --coverage"},
			{Area: "Ownership attribution", Reason: "No CODEOWNERS file detected"},
		},
		BenchmarkReadiness: summary.BenchmarkReadinessSummary{
			ReadyDimensions: []string{"test structure"},
		},
	}

	var buf bytes.Buffer
	RenderExecutiveSummary(&buf, es)
	output := buf.String()

	expected := []string{
		"Known Blind Spots",
		"Coverage data: No coverage artifacts",
		"→ Run with --coverage",
		"Ownership attribution: No CODEOWNERS",
	}
	for _, s := range expected {
		if !strings.Contains(output, s) {
			t.Errorf("output missing %q", s)
		}
	}
}
