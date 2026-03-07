package testdata

import (
	"bytes"
	"testing"

	"github.com/pmclSF/hamlet/internal/heatmap"
	"github.com/pmclSF/hamlet/internal/impact"
	"github.com/pmclSF/hamlet/internal/measurement"
	"github.com/pmclSF/hamlet/internal/metrics"
	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/reporting"
	"github.com/pmclSF/hamlet/internal/scoring"
)

// TestAdversarial_NilMeasurements verifies reporting handles nil measurements.
func TestAdversarial_NilMeasurements(t *testing.T) {
	snap := MinimalSnapshot()
	snap.Measurements = nil

	h := heatmap.Build(snap)
	var buf bytes.Buffer
	reporting.RenderSummaryReport(&buf, snap, h)

	if buf.Len() == 0 {
		t.Error("expected non-empty output even with nil measurements")
	}
}

// TestAdversarial_EmptySignals verifies scoring handles zero signals.
func TestAdversarial_EmptySignals(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		Signals: []models.Signal{},
	}
	risks := scoring.ComputeRisk(snap)
	// Should not panic; may produce zero or empty risk surfaces.
	_ = risks
}

// TestAdversarial_ZeroTestFiles verifies metrics handles zero test files.
func TestAdversarial_ZeroTestFiles(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		GeneratedAt: FixedTime,
	}
	ms := metrics.Derive(snap)

	if ms.Structure.TotalTestFiles != 0 {
		t.Errorf("expected 0 test files, got %d", ms.Structure.TotalTestFiles)
	}
}

// TestAdversarial_MeasurementsOnEmpty verifies measurement handles empty snapshot.
func TestAdversarial_MeasurementsOnEmpty(t *testing.T) {
	snap := &models.TestSuiteSnapshot{}

	reg := measurement.DefaultRegistry()
	result := reg.ComputeSnapshot(snap)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Posture) != 5 {
		t.Errorf("expected 5 posture dimensions, got %d", len(result.Posture))
	}
}

// TestAdversarial_ImpactEmptyScope verifies impact analysis handles empty scope.
func TestAdversarial_ImpactEmptyScope(t *testing.T) {
	snap := HealthyBalancedSnapshot()
	scope := &impact.ChangeScope{
		ChangedFiles: []impact.ChangedFile{},
	}

	result := impact.Analyze(scope, snap)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.ImpactedUnits) != 0 {
		t.Errorf("expected 0 impacted units for empty scope, got %d", len(result.ImpactedUnits))
	}
}

// TestAdversarial_ImpactNonexistentFile verifies impact handles files not in snapshot.
func TestAdversarial_ImpactNonexistentFile(t *testing.T) {
	snap := MinimalSnapshot()
	scope := &impact.ChangeScope{
		ChangedFiles: []impact.ChangedFile{
			{Path: "does/not/exist.js", ChangeKind: impact.ChangeModified},
		},
	}

	result := impact.Analyze(scope, snap)

	// Should create a file-level impacted unit with weak confidence.
	if len(result.ImpactedUnits) != 1 {
		t.Fatalf("expected 1 file-level unit, got %d", len(result.ImpactedUnits))
	}
	if result.ImpactedUnits[0].ImpactConfidence != impact.ConfidenceWeak {
		t.Errorf("expected weak confidence, got %s", result.ImpactedUnits[0].ImpactConfidence)
	}
}

// TestAdversarial_HeatmapNoRisk verifies heatmap builds without risk data.
func TestAdversarial_HeatmapNoRisk(t *testing.T) {
	snap := MinimalSnapshot()
	snap.Risk = nil

	h := heatmap.Build(snap)
	if h == nil {
		t.Fatal("expected non-nil heatmap")
	}
}

// TestAdversarial_LargeSignalVolume verifies scoring handles many signals.
func TestAdversarial_LargeSignalVolume(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		GeneratedAt: FixedTime,
	}

	// Add 1000 signals.
	for i := 0; i < 1000; i++ {
		snap.Signals = append(snap.Signals, models.Signal{
			Type:     "test_signal",
			Category: models.CategoryHealth,
			Severity: models.SeverityMedium,
			Location: models.SignalLocation{File: "src/file.js"},
		})
	}

	risks := scoring.ComputeRisk(snap)
	_ = risks // Should not panic or hang.
}

// TestAdversarial_FilterByOwner_NoMatch verifies filter with nonexistent owner.
func TestAdversarial_FilterByOwner_NoMatch(t *testing.T) {
	result := &impact.ImpactResult{
		ImpactedUnits: []impact.ImpactedCodeUnit{
			{Name: "Foo", Owner: "team-a"},
		},
		ImpactedTests: []impact.ImpactedTest{
			{Path: "test/foo.test.js"},
		},
		Posture: impact.ChangeRiskPosture{Band: "well_protected"},
	}

	filtered := impact.FilterByOwner(result, "nonexistent-team")

	if len(filtered.ImpactedUnits) != 0 {
		t.Errorf("expected 0 units for nonexistent owner, got %d", len(filtered.ImpactedUnits))
	}
}
