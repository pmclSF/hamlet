package engine

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/analysis"
	"github.com/pmclSF/hamlet/internal/measurement"
	"github.com/pmclSF/hamlet/internal/scoring"
	"github.com/pmclSF/hamlet/internal/testdata"
)

// TestPipelineSteps_Integration verifies each pipeline step works with
// standardized fixtures without requiring filesystem access.
func TestPipelineSteps_Integration(t *testing.T) {
	snap := testdata.HealthyBalancedSnapshot()

	// Step 3 equivalent: run detectors on an in-memory snapshot.
	registry := DefaultRegistry(Config{RepoRoot: "."})
	registry.Run(snap)

	// Step 5: compute risk surfaces.
	snap.Risk = scoring.ComputeRisk(snap)

	// Step 6: compute measurements.
	measRegistry := measurement.DefaultRegistry()
	measSnap := measRegistry.ComputeSnapshot(snap)
	snap.Measurements = measSnap.ToModel()

	// Verify pipeline produced expected artifacts.
	// Risk surfaces depend on signals; in-memory fixtures may produce zero
	// signals when file-reading detectors have no filesystem to inspect.
	// The important check is that the pipeline runs without error.
	if snap.Measurements == nil {
		t.Fatal("expected measurements to be computed")
	}
	if len(snap.Measurements.Posture) == 0 {
		t.Error("expected posture dimensions")
	}

	// Verify all 5 posture dimensions are present.
	dims := map[string]bool{}
	for _, p := range snap.Measurements.Posture {
		dims[p.Dimension] = true
	}
	expected := []string{"health", "coverage_depth", "coverage_diversity", "structural_risk", "operational_risk"}
	for _, d := range expected {
		if !dims[d] {
			t.Errorf("missing posture dimension: %s", d)
		}
	}
}

func TestPipelineSteps_EmptySnapshot(t *testing.T) {
	snap := testdata.EmptySnapshot()

	registry := DefaultRegistry(Config{RepoRoot: "."})
	registry.Run(snap)

	snap.Risk = scoring.ComputeRisk(snap)

	measRegistry := measurement.DefaultRegistry()
	measSnap := measRegistry.ComputeSnapshot(snap)
	snap.Measurements = measSnap.ToModel()

	if snap.Measurements == nil {
		t.Fatal("expected measurements even for empty snapshot")
	}
	if len(snap.Measurements.Posture) != 5 {
		t.Errorf("expected 5 posture dimensions, got %d", len(snap.Measurements.Posture))
	}
}

func TestPipelineSteps_LargeScale(t *testing.T) {
	snap := testdata.LargeScaleSnapshot()

	registry := DefaultRegistry(Config{RepoRoot: "."})
	registry.Run(snap)

	snap.Risk = scoring.ComputeRisk(snap)

	measRegistry := measurement.DefaultRegistry()
	measSnap := measRegistry.ComputeSnapshot(snap)
	snap.Measurements = measSnap.ToModel()

	if snap.Measurements == nil {
		t.Fatal("expected measurements")
	}
	if len(snap.TestFiles) != 550 {
		t.Errorf("expected 550 test files, got %d", len(snap.TestFiles))
	}
}

func TestRunPipeline_AnalysisTestdata(t *testing.T) {
	// Use the existing analysis testdata directory for a real pipeline run.
	result, err := RunPipeline("../analysis/testdata/sample-repo")
	if err != nil {
		t.Fatalf("RunPipeline failed: %v", err)
	}

	if result.Snapshot == nil {
		t.Fatal("expected snapshot")
	}
	if len(result.Snapshot.TestFiles) == 0 {
		t.Error("expected test files")
	}
	if result.Snapshot.Measurements == nil {
		t.Error("expected measurements")
	}
}

// Verify that analysis.New returns something usable even for a nonexistent repo.
func TestAnalyzerNewDoesNotPanic(t *testing.T) {
	a := analysis.New("/nonexistent/path")
	if a == nil {
		t.Error("expected non-nil analyzer")
	}
}
