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
		HasTrendData:    true,
		DominantDrivers: []string{"weakAssertion", "mockHeavyTest"},
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

	if !strings.Contains(output, "No prior snapshots available") {
		t.Error("expected no-trend-data message")
	}
	if !strings.Contains(output, "write-snapshot") {
		t.Error("expected hint about write-snapshot")
	}
}

func TestRenderExecutiveSummary_Empty(t *testing.T) {
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
