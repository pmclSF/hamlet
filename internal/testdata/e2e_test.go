package testdata

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/pmclSF/hamlet/internal/benchmark"
	"github.com/pmclSF/hamlet/internal/comparison"
	"github.com/pmclSF/hamlet/internal/heatmap"
	"github.com/pmclSF/hamlet/internal/impact"
	"github.com/pmclSF/hamlet/internal/measurement"
	"github.com/pmclSF/hamlet/internal/metrics"
	"github.com/pmclSF/hamlet/internal/reporting"
	"github.com/pmclSF/hamlet/internal/scoring"
	"github.com/pmclSF/hamlet/internal/summary"
)

// TestE2E_FullAnalysisToSummary exercises the flagship flow:
// analyze → heatmap → metrics → measurements → summary → render.
func TestE2E_FullAnalysisToSummary(t *testing.T) {
	snap := HealthyBalancedSnapshot()

	// Step 1: Compute risk.
	snap.Risk = scoring.ComputeRisk(snap)

	// Step 2: Compute measurements.
	measReg := measurement.DefaultRegistry()
	snap.Measurements = measReg.ComputeSnapshot(snap).ToModel()

	// Step 3: Build heatmap and metrics.
	h := heatmap.Build(snap)
	ms := metrics.Derive(snap)

	// Step 4: Build benchmark export.
	export := benchmark.BuildExport(snap, ms, false)
	if export.SchemaVersion != "2" {
		t.Errorf("export schema version: got %q, want %q", export.SchemaVersion, "2")
	}

	// Step 5: Build executive summary.
	es := summary.Build(&summary.BuildInput{
		Snapshot:  snap,
		Heatmap:   h,
		Metrics:   ms,
		HasPolicy: false,
	})
	if es == nil {
		t.Fatal("expected non-nil executive summary")
	}

	// Step 6: Render reports.
	var buf bytes.Buffer
	reporting.RenderSummaryReport(&buf, snap, h)
	if buf.Len() == 0 {
		t.Error("expected non-empty summary report")
	}

	buf.Reset()
	reporting.RenderPostureReport(&buf, snap)
	if buf.Len() == 0 {
		t.Error("expected non-empty posture report")
	}

	buf.Reset()
	reporting.RenderAnalyzeReport(&buf, snap)
	if buf.Len() == 0 {
		t.Error("expected non-empty analyze report")
	}
}

// TestE2E_ComparisonWorkflow exercises snapshot → compare → render.
func TestE2E_ComparisonWorkflow(t *testing.T) {
	from := FlakyConcentratedSnapshot()
	to := HealthyBalancedSnapshot()

	comp := comparison.Compare(from, to)

	if comp == nil {
		t.Fatal("expected non-nil comparison")
	}

	var buf bytes.Buffer
	reporting.RenderComparisonReport(&buf, comp)
	if buf.Len() == 0 {
		t.Error("expected non-empty comparison report")
	}
}

// TestE2E_ImpactWorkflow exercises change → analyze → impact → render.
func TestE2E_ImpactWorkflow(t *testing.T) {
	snap := HealthyBalancedSnapshot()

	// Compute measurements for a full snapshot.
	measReg := measurement.DefaultRegistry()
	snap.Measurements = measReg.ComputeSnapshot(snap).ToModel()

	// Create a change scope.
	scope := impact.ChangeScopeFromPaths(
		[]string{"src/auth.js", "src/payment.js", "src/__tests__/auth.test.js"},
		impact.ChangeModified,
	)

	// Analyze impact.
	result := impact.Analyze(scope, snap)

	if result.Summary == "" {
		t.Error("expected non-empty impact summary")
	}
	if result.Posture.Band == "" {
		t.Error("expected non-empty posture band")
	}

	// Render full report.
	var buf bytes.Buffer
	reporting.RenderImpactReport(&buf, result)
	if buf.Len() == 0 {
		t.Error("expected non-empty impact report")
	}

	// Render drill-downs.
	buf.Reset()
	reporting.RenderImpactUnits(&buf, result)
	if buf.Len() == 0 {
		t.Error("expected non-empty units drill-down")
	}

	buf.Reset()
	reporting.RenderImpactGaps(&buf, result)
	if buf.Len() == 0 {
		t.Error("expected non-empty gaps drill-down")
	}

	// Build aggregate.
	agg := impact.BuildAggregate(result)
	if agg.ChangedFileCount != 3 {
		t.Errorf("expected 3 changed files, got %d", agg.ChangedFileCount)
	}

	// Verify aggregate serializes.
	data, err := json.Marshal(agg)
	if err != nil {
		t.Fatalf("marshal aggregate: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty aggregate JSON")
	}
}

// TestE2E_MigrationRiskFlow exercises migration-specific analysis.
func TestE2E_MigrationRiskFlow(t *testing.T) {
	snap := MigrationRiskSnapshot()

	// Compute risk and measurements.
	snap.Risk = scoring.ComputeRisk(snap)
	measReg := measurement.DefaultRegistry()
	snap.Measurements = measReg.ComputeSnapshot(snap).ToModel()

	ms := metrics.Derive(snap)

	// Multi-framework repo should have framework count > 1.
	if ms.Structure.FrameworkCount <= 1 {
		t.Errorf("expected multiple frameworks, got %d", ms.Structure.FrameworkCount)
	}

	// Should have posture dimensions.
	if snap.Measurements == nil || len(snap.Measurements.Posture) == 0 {
		t.Error("expected posture dimensions")
	}
}

// TestE2E_ExportPrivacySafe verifies the export contains no raw paths.
func TestE2E_ExportPrivacySafe(t *testing.T) {
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
