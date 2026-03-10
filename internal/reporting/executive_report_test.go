package reporting

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pmclSF/hamlet/internal/benchmark"
	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/summary"
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
			{Description: "weakAssertion findings decreased (-3)", Direction: "improved"},
			{Description: "flakyTest findings increased (+2)", Direction: "worsened"},
		},
		HasTrendData:     true,
		DominantDrivers:  []string{"weakAssertion", "mockHeavyTest"},
		RecommendedFocus: "Address quality risk in src/auth; reduce weakAssertion findings.",
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
		"Hamlet Executive Summary",
		"Overall Posture",
		"reliability:",
		"change:",
		"Key Numbers",
		"Test files:",
		"Top Risk Areas",
		"src/auth",
		"Trend Highlights",
		"weakAssertion findings decreased",
		"flakyTest findings increased",
		"Dominant Drivers",
		"weakAssertion",
		"Recommended Focus",
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

	if !strings.Contains(output, "Hamlet Executive Summary") {
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
