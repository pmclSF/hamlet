package reporting

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/analyze"
)

func TestRenderAnalyzeReportV2_KeyFindings(t *testing.T) {
	t.Parallel()
	r := &analyze.Report{
		SchemaVersion: "1.0.0",
		Repository:    analyze.RepositoryInfo{Name: "test-repo"},
		RepoProfile: analyze.ProfileSummary{
			TestVolume:         "medium",
			CIPressure:         "low",
			CoverageConfidence: "low",
		},
		KeyFindings: []analyze.KeyFinding{
			{Title: "3 critical signals detected", Severity: "critical", Category: "reliability", Metric: "3 critical"},
			{Title: "80 source files have low coverage", Severity: "high", Category: "coverage_debt", Metric: "80 files"},
			{Title: "200 duplicate tests", Severity: "medium", Category: "optimization", Metric: "200 duplicates"},
		},
		TotalFindingCount: 6,
		TestsDetected:     analyze.TestSummary{TestFileCount: 42},
		SignalSummary:     analyze.SignalBreakdown{Total: 10, Critical: 3},
	}

	var buf bytes.Buffer
	RenderAnalyzeReportV2(&buf, r)
	output := buf.String()

	// Key Findings section should appear.
	if !strings.Contains(output, "Key Findings") {
		t.Error("output should contain 'Key Findings' section")
	}

	// All 3 findings should be rendered.
	if !strings.Contains(output, "[CRITICAL]") {
		t.Error("output should contain [CRITICAL] badge")
	}
	if !strings.Contains(output, "[HIGH]") {
		t.Error("output should contain [HIGH] badge")
	}
	if !strings.Contains(output, "[MEDIUM]") {
		t.Error("output should contain [MEDIUM] badge")
	}

	// Numbered list.
	if !strings.Contains(output, "1. [CRITICAL]") {
		t.Error("findings should be numbered starting at 1")
	}

	// Remaining count.
	if !strings.Contains(output, "3 more finding(s) available") {
		t.Error("output should show remaining finding count")
	}
	if !strings.Contains(output, "terrain insights") {
		t.Error("output should reference terrain insights command")
	}
}

func TestRenderAnalyzeReportV2_NoFindings(t *testing.T) {
	t.Parallel()
	r := &analyze.Report{
		SchemaVersion: "1.0.0",
		Repository:    analyze.RepositoryInfo{Name: "clean-repo"},
		RepoProfile: analyze.ProfileSummary{
			TestVolume: "small",
		},
		KeyFindings:       nil,
		TotalFindingCount: 0,
		TopInsight:        "No major issues detected.",
		TestsDetected:     analyze.TestSummary{TestFileCount: 5},
	}

	var buf bytes.Buffer
	RenderAnalyzeReportV2(&buf, r)
	output := buf.String()

	// Should fall back to TopInsight.
	if !strings.Contains(output, "Top Insight") {
		t.Error("output should fall back to 'Top Insight' when no key findings")
	}
	if !strings.Contains(output, "No major issues detected.") {
		t.Error("output should show the TopInsight text")
	}
}

func TestRenderAnalyzeReportV2_NextStepsShowsFindingCount(t *testing.T) {
	t.Parallel()
	r := &analyze.Report{
		SchemaVersion: "1.0.0",
		Repository:    analyze.RepositoryInfo{Name: "test-repo"},
		RepoProfile:   analyze.ProfileSummary{TestVolume: "medium"},
		KeyFindings: []analyze.KeyFinding{
			{Title: "issue 1", Severity: "high", Category: "reliability"},
		},
		TotalFindingCount: 5,
		TestsDetected:     analyze.TestSummary{TestFileCount: 10},
	}

	var buf bytes.Buffer
	RenderAnalyzeReportV2(&buf, r)
	output := buf.String()

	// Next steps should reference the total finding count.
	if !strings.Contains(output, "5 finding(s)") {
		t.Error("next steps should mention total finding count")
	}
}

func TestRenderAnalyzeReportV2_Deterministic(t *testing.T) {
	t.Parallel()
	r := &analyze.Report{
		SchemaVersion: "1.0.0",
		Repository:    analyze.RepositoryInfo{Name: "test-repo"},
		RepoProfile:   analyze.ProfileSummary{TestVolume: "medium"},
		KeyFindings: []analyze.KeyFinding{
			{Title: "a", Severity: "high", Category: "reliability"},
			{Title: "b", Severity: "medium", Category: "optimization"},
		},
		TotalFindingCount: 2,
		TestsDetected:     analyze.TestSummary{TestFileCount: 10},
	}

	var buf1, buf2 bytes.Buffer
	RenderAnalyzeReportV2(&buf1, r)
	RenderAnalyzeReportV2(&buf2, r)

	if buf1.String() != buf2.String() {
		t.Error("rendering should be deterministic")
	}
}
