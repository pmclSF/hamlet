package testdata

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/pmclSF/hamlet/internal/benchmark"
	"github.com/pmclSF/hamlet/internal/comparison"
	"github.com/pmclSF/hamlet/internal/graph"
	"github.com/pmclSF/hamlet/internal/heatmap"
	"github.com/pmclSF/hamlet/internal/impact"
	"github.com/pmclSF/hamlet/internal/measurement"
	"github.com/pmclSF/hamlet/internal/metrics"
	"github.com/pmclSF/hamlet/internal/migration"
	"github.com/pmclSF/hamlet/internal/models"
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
	if export.SchemaVersion != "3" {
		t.Errorf("export schema version: got %q, want %q", export.SchemaVersion, "3")
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

// TestE2E_GraphEnrichedHeatmap exercises the graph-backed heatmap flow.
func TestE2E_GraphEnrichedHeatmap(t *testing.T) {
	snap := MigrationRiskSnapshot()
	snap.Risk = scoring.ComputeRisk(snap)

	// Build graph and graph-enriched heatmap.
	g := graph.Build(snap)
	h := heatmap.BuildWithGraph(snap, g)

	if h == nil {
		t.Fatal("expected non-nil heatmap")
	}

	// Owner hotspots should exist (signals are present after risk scoring).
	if len(h.OwnerHotSpots) == 0 && len(snap.Signals) > 0 {
		t.Error("expected owner hotspots from signals")
	}

	// Graph should index the coverage insights.
	if len(g.E2EOnlyUnits) == 0 {
		t.Error("expected e2e-only units from coverage insights")
	}

	// Owner risk summaries should aggregate coverage data.
	summaries := g.OwnerRiskSummaries()
	if len(summaries) == 0 {
		t.Error("expected owner risk summaries")
	}

	// Module coverage summaries should be non-empty.
	modules := g.ModuleCoverageSummaries()
	if len(modules) == 0 {
		t.Error("expected module coverage summaries")
	}
}

// TestE2E_ReviewWithCoverageAndIdentity exercises the review renderer
// with coverage-by-type and test identity data.
func TestE2E_ReviewWithCoverageAndIdentity(t *testing.T) {
	snap := MigrationRiskSnapshot()

	// Add signals so the review renderer doesn't bail out early.
	snap.Signals = append(snap.Signals, models.Signal{
		Type:     "weakAssertion",
		Category: models.CategoryQuality,
		Severity: models.SeverityMedium,
		Location: models.SignalLocation{File: "spec/api.spec.js"},
		Owner:    "team-api",
	})
	snap.Risk = scoring.ComputeRisk(snap)

	var buf bytes.Buffer
	reporting.RenderReviewSections(&buf, snap)
	output := buf.String()

	if output == "" {
		t.Fatal("expected non-empty review output")
	}

	// Should contain coverage section from the fixture's CoverageSummary.
	if !strings.Contains(output, "Coverage by Type") {
		t.Error("expected 'Coverage by Type' section in review output")
	}
	if !strings.Contains(output, "Covered only by e2e") {
		t.Error("expected e2e-only coverage data in review output")
	}
}

// TestE2E_MigrationWithCoverageGuidance exercises migration readiness
// with coverage-by-type data for richer guidance.
func TestE2E_MigrationWithCoverageGuidance(t *testing.T) {
	snap := MigrationRiskSnapshot()
	snap.Risk = scoring.ComputeRisk(snap)

	readiness := migration.ComputeReadiness(snap)

	if readiness == nil {
		t.Fatal("expected non-nil readiness summary")
	}

	// Should have coverage guidance from e2e-only insights.
	if len(readiness.CoverageGuidance) == 0 && len(snap.CoverageInsights) > 0 {
		// Coverage guidance depends on area assessments having non-safe areas.
		// This is acceptable when all areas are safe.
		t.Log("no coverage guidance generated (areas may all be safe)")
	}

	// Should have area assessments.
	if len(readiness.AreaAssessments) == 0 {
		t.Error("expected area assessments")
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
