package engine

import (
	"testing"

	"github.com/pmclSF/terrain/internal/measurement"
	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/scoring"
	"github.com/pmclSF/terrain/internal/testdata"
)

func runPipelineSteps(t *testing.T, snap *models.TestSuiteSnapshot) {
	t.Helper()
	registry := DefaultRegistry(Config{RepoRoot: "."})
	registry.Run(snap)
	snap.Risk = scoring.ComputeRisk(snap)
	measRegistry := measurement.DefaultRegistry()
	measSnap := measRegistry.ComputeSnapshot(snap)
	snap.Measurements = measSnap.ToModel()
	models.SortSnapshot(snap)
}

func TestPipeline_MixedFrameworks(t *testing.T) {
	t.Parallel()
	snap := testdata.MixedFrameworkSnapshot()
	runPipelineSteps(t, snap)

	if snap.Measurements == nil {
		t.Fatal("expected measurements")
	}
	if len(snap.Frameworks) < 5 {
		t.Errorf("expected 5+ frameworks, got %d", len(snap.Frameworks))
	}
}

func TestPipeline_ZeroSignals(t *testing.T) {
	t.Parallel()
	snap := testdata.ZeroSignalSnapshot()
	runPipelineSteps(t, snap)

	// A well-tested repo with good assertions should produce few or no quality signals.
	// (Some detectors may still fire based on structural heuristics.)
	if snap.Measurements == nil {
		t.Fatal("expected measurements")
	}
}

func TestPipeline_AllSignalTypes(t *testing.T) {
	t.Parallel()
	snap := testdata.AllSignalTypesSnapshot()
	runPipelineSteps(t, snap)

	// Should have pre-loaded signals plus any detector-generated ones.
	if len(snap.Signals) < 25 {
		t.Errorf("expected at least 25 pre-loaded signals, got %d", len(snap.Signals))
	}

	// Risk surfaces should be computed from the many signals.
	if len(snap.Risk) == 0 {
		t.Error("expected risk surfaces from pre-loaded signals")
	}
}

func TestPipeline_DeepNesting(t *testing.T) {
	t.Parallel()
	snap := testdata.DeepNestingSnapshot()
	runPipelineSteps(t, snap)

	if snap.Measurements == nil {
		t.Fatal("expected measurements")
	}
}

func TestPipeline_VeryLargeScale(t *testing.T) {
	t.Parallel()
	snap := testdata.VeryLargeSnapshot()
	runPipelineSteps(t, snap)

	if snap.Measurements == nil {
		t.Fatal("expected measurements")
	}
	if len(snap.TestFiles) != 2000 {
		t.Errorf("expected 2000 test files, got %d", len(snap.TestFiles))
	}
}

func TestPipeline_EmptyFrameworks(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		SnapshotMeta: models.SnapshotMeta{SchemaVersion: models.SnapshotSchemaVersion},
		Repository:   models.RepositoryMetadata{Name: "no-frameworks"},
		GeneratedAt:  testdata.FixedTime,
	}
	runPipelineSteps(t, snap)

	if snap.Measurements == nil {
		t.Fatal("expected measurements even with no frameworks")
	}
}

func TestPipeline_ValidationAfterRun(t *testing.T) {
	t.Parallel()
	snap := testdata.HealthyBalancedSnapshot()
	snap.SnapshotMeta = models.SnapshotMeta{SchemaVersion: models.SnapshotSchemaVersion}
	runPipelineSteps(t, snap)

	// Validate all signals produced by the pipeline.
	for i, s := range snap.Signals {
		if err := models.ValidateSignal(s); err != nil {
			t.Errorf("signal[%d] (%s) failed validation: %v", i, s.Type, err)
		}
	}
}
