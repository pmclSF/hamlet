package testdata

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/benchmark"
	"github.com/pmclSF/terrain/internal/impact"
	"github.com/pmclSF/terrain/internal/measurement"
	"github.com/pmclSF/terrain/internal/metrics"
	"github.com/pmclSF/terrain/internal/migration"
	"github.com/pmclSF/terrain/internal/portfolio"
	"github.com/pmclSF/terrain/internal/reporting"
	"github.com/pmclSF/terrain/internal/scoring"
)

// TestE2E_PortfolioIntelligenceFlow exercises analyze -> portfolio -> render -> export.
func TestE2E_PortfolioIntelligenceFlow(t *testing.T) {
	t.Parallel()
	snap := FlakyConcentratedSnapshot()
	snap.Risk = scoring.ComputeRisk(snap)
	measReg := measurement.DefaultRegistry()
	snap.Measurements = measReg.ComputeSnapshot(snap).ToModel()

	// Run portfolio analysis.
	ps := portfolio.Analyze(snap)
	snap.Portfolio = ps.ToModel()

	if snap.Portfolio == nil {
		t.Fatal("expected non-nil portfolio snapshot")
	}
	if snap.Portfolio.Aggregates.TotalAssets == 0 {
		t.Error("expected non-zero total assets")
	}

	// Flaky concentrated fixture has expensive E2E tests -> low-value findings.
	if snap.Portfolio.Aggregates.LowValueHighCostCount == 0 {
		t.Error("expected low-value high-cost findings for flaky E2E tests")
	}

	// Posture band should reflect the problems.
	if snap.Portfolio.Aggregates.PortfolioPostureBand == "" {
		t.Error("expected non-empty portfolio posture band")
	}

	// Render portfolio report.
	var buf bytes.Buffer
	reporting.RenderPortfolioReport(&buf, snap)
	output := buf.String()
	if !strings.Contains(output, "Portfolio Intelligence") {
		t.Error("expected 'Portfolio Intelligence' header in report")
	}
	if !strings.Contains(output, "LOW-VALUE") {
		t.Error("expected LOW-VALUE badges in portfolio report")
	}

	// Render analyze report — should include portfolio section.
	buf.Reset()
	reporting.RenderAnalyzeReport(&buf, snap)
	if !strings.Contains(buf.String(), "Portfolio Intelligence") {
		t.Error("expected portfolio section in analyze report")
	}

	// Benchmark export should include portfolio stats.
	ms := metrics.Derive(snap)
	export := benchmark.BuildExport(snap, ms, false)
	if export.PortfolioStats == nil {
		t.Error("expected portfolio stats in benchmark export")
	} else {
		if export.PortfolioStats.PortfolioPostureBand == "" {
			t.Error("expected non-empty portfolio posture band in export")
		}
		// Should not contain file paths.
		data, err := json.Marshal(export.PortfolioStats)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "cypress/e2e") {
			t.Error("portfolio export contains raw paths — privacy violation")
		}
	}
}

// TestE2E_ExportPrivacySafe verifies the export contains no raw paths.
func TestE2E_ExportPrivacySafe(t *testing.T) {
	t.Parallel()
	snap := HealthyBalancedSnapshot()
	ms := metrics.Derive(snap)
	export := benchmark.BuildExport(snap, ms, true)

	data, err := json.Marshal(export)
	if err != nil {
		t.Fatal(err)
	}
	output := string(data)

	// Must not contain any file paths from the fixture.
	forbidden := []string{"src/auth.js", "src/user.js", "__tests__", "e2e/"}
	for _, f := range forbidden {
		if bytes.Contains(data, []byte(f)) {
			t.Errorf("export contains raw path %q — privacy violation\nExport: %s", f, output[:min(len(output), 500)])
		}
	}
}

// TestE2E_ImpactSelectTestsFlow exercises impact -> select-tests -> aggregate.
func TestE2E_ImpactSelectTestsFlow(t *testing.T) {
	t.Parallel()
	snap, changedFiles := ChangeScopedPRSnapshot()

	// Compute measurements for a full snapshot.
	measReg := measurement.DefaultRegistry()
	snap.Measurements = measReg.ComputeSnapshot(snap).ToModel()

	// Create scope and analyze.
	scope := impact.ChangeScopeFromPaths(changedFiles, impact.ChangeModified)
	result := impact.Analyze(scope, snap)

	// Verify protective set.
	if result.ProtectiveSet == nil {
		t.Fatal("expected protective test set")
	}
	if len(result.ProtectiveSet.Tests) == 0 {
		t.Error("expected at least one protective test")
	}
	if result.ProtectiveSet.SetKind == "" {
		t.Error("expected selection strategy kind")
	}

	// Verify impact graph.
	if result.Graph == nil {
		t.Fatal("expected impact graph")
	}
	if result.Graph.Stats.TotalEdges == 0 {
		t.Error("expected non-zero graph edges")
	}

	// Build aggregate.
	agg := impact.BuildAggregate(result)
	if agg.SelectionSetKind == "" {
		t.Error("expected selection set kind in aggregate")
	}

	// Render protective set view.
	var buf bytes.Buffer
	reporting.RenderProtectiveSet(&buf, result)
	if buf.Len() == 0 {
		t.Error("expected non-empty protective set rendering")
	}

	// Render impact owners view.
	buf.Reset()
	reporting.RenderImpactOwners(&buf, result)
	if buf.Len() == 0 {
		t.Error("expected non-empty impact owners rendering")
	}
}

// TestE2E_PostureExplainFlow exercises analyze -> posture -> explain -> render.
func TestE2E_PostureExplainFlow(t *testing.T) {
	t.Parallel()
	snap := HealthyBalancedSnapshot()
	snap.Risk = scoring.ComputeRisk(snap)

	measReg := measurement.DefaultRegistry()
	snap.Measurements = measReg.ComputeSnapshot(snap).ToModel()

	// Posture should have all 5 dimensions.
	if snap.Measurements == nil {
		t.Fatal("expected measurements")
	}
	if len(snap.Measurements.Posture) != 5 {
		t.Errorf("expected 5 posture dimensions, got %d", len(snap.Measurements.Posture))
	}

	// Render posture report.
	var buf bytes.Buffer
	reporting.RenderPostureReport(&buf, snap)
	output := buf.String()
	if !strings.Contains(output, "Terrain Posture") {
		t.Error("expected 'Terrain Posture' header")
	}
	// Should contain dimension names.
	for _, dim := range []string{"HEALTH", "COVERAGE_DEPTH", "STRUCTURAL_RISK"} {
		if !strings.Contains(output, dim) {
			t.Errorf("expected dimension %q in posture output", dim)
		}
	}
}

// TestE2E_MigrationReadinessFlow exercises migration -> readiness -> render.
func TestE2E_MigrationReadinessFlow(t *testing.T) {
	t.Parallel()
	snap := MigrationRiskSnapshot()
	snap.Risk = scoring.ComputeRisk(snap)

	readiness := migration.ComputeReadiness(snap)
	if readiness == nil {
		t.Fatal("expected non-nil readiness")
	}
	if len(readiness.AreaAssessments) == 0 {
		t.Error("expected area assessments")
	}

	var buf bytes.Buffer
	reporting.RenderMigrationReport(&buf, readiness)
	if buf.Len() == 0 {
		t.Error("expected non-empty migration report")
	}
}

// TestE2E_ViewModelDrillDowns exercises all impact drill-down renderers.
func TestE2E_ViewModelDrillDowns(t *testing.T) {
	t.Parallel()
	snap := HealthyBalancedSnapshot()
	measReg := measurement.DefaultRegistry()
	snap.Measurements = measReg.ComputeSnapshot(snap).ToModel()

	scope := impact.ChangeScopeFromPaths(
		[]string{"src/auth.js", "src/user.js", "src/payment.js"},
		impact.ChangeModified,
	)
	result := impact.Analyze(scope, snap)

	renderers := map[string]func(){
		"units":    func() { reporting.RenderImpactUnits(&bytes.Buffer{}, result) },
		"gaps":     func() { reporting.RenderImpactGaps(&bytes.Buffer{}, result) },
		"tests":    func() { reporting.RenderImpactTests(&bytes.Buffer{}, result) },
		"graph":    func() { reporting.RenderImpactGraph(&bytes.Buffer{}, result) },
		"selected": func() { reporting.RenderProtectiveSet(&bytes.Buffer{}, result) },
		"owners":   func() { reporting.RenderImpactOwners(&bytes.Buffer{}, result) },
		"full":     func() { reporting.RenderImpactReport(&bytes.Buffer{}, result) },
	}

	for name, render := range renderers {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			// Re-assign buffer to capture output.
			switch name {
			case "units":
				reporting.RenderImpactUnits(&buf, result)
			case "gaps":
				reporting.RenderImpactGaps(&buf, result)
			case "tests":
				reporting.RenderImpactTests(&buf, result)
			case "graph":
				reporting.RenderImpactGraph(&buf, result)
			case "selected":
				reporting.RenderProtectiveSet(&buf, result)
			case "owners":
				reporting.RenderImpactOwners(&buf, result)
			case "full":
				reporting.RenderImpactReport(&buf, result)
			default:
				render()
			}
			if buf.Len() == 0 {
				t.Errorf("expected non-empty output for %s drill-down", name)
			}
		})
	}
}
