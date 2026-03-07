package engine

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/pmclSF/hamlet/internal/analysis"
	"github.com/pmclSF/hamlet/internal/measurement"
	"github.com/pmclSF/hamlet/internal/models"
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

	// Verify schema version and detector manifest are populated.
	meta := result.Snapshot.SnapshotMeta
	if meta.SchemaVersion != models.SnapshotSchemaVersion {
		t.Errorf("expected schema version %s, got %s", models.SnapshotSchemaVersion, meta.SchemaVersion)
	}
	if meta.DetectorCount == 0 {
		t.Error("expected non-zero detector count")
	}
	if len(meta.Detectors) != meta.DetectorCount {
		t.Errorf("detector list length %d != count %d", len(meta.Detectors), meta.DetectorCount)
	}
}

// Verify that analysis.New returns something usable even for a nonexistent repo.
func TestAnalyzerNewDoesNotPanic(t *testing.T) {
	a := analysis.New("/nonexistent/path")
	if a == nil {
		t.Error("expected non-nil analyzer")
	}
}

// TestPipelineDeterminism verifies that running the pipeline twice on
// identical input produces byte-identical JSON output (excluding timestamps).
func TestPipelineDeterminism(t *testing.T) {
	run := func() string {
		snap := testdata.HealthyBalancedSnapshot()
		registry := DefaultRegistry(Config{RepoRoot: "."})
		registry.Run(snap)
		snap.Risk = scoring.ComputeRisk(snap)
		measRegistry := measurement.DefaultRegistry()
		measSnap := measRegistry.ComputeSnapshot(snap)
		snap.Measurements = measSnap.ToModel()
		models.SortSnapshot(snap)
		// Zero out timestamps for comparison.
		snap.GeneratedAt = testdata.FixedTime
		snap.Repository.SnapshotTimestamp = testdata.FixedTime
		out, err := json.Marshal(snap)
		if err != nil {
			t.Fatalf("marshal failed: %v", err)
		}
		return string(out)
	}

	a := run()
	b := run()
	if a != b {
		t.Error("pipeline output is not deterministic across identical runs")
	}
}

// TestPipelineOutputSorted verifies that pipeline output slices are sorted.
func TestPipelineOutputSorted(t *testing.T) {
	result, err := RunPipeline("../analysis/testdata/sample-repo")
	if err != nil {
		t.Fatalf("RunPipeline failed: %v", err)
	}
	snap := result.Snapshot

	if !sort.SliceIsSorted(snap.TestFiles, func(i, j int) bool {
		return snap.TestFiles[i].Path < snap.TestFiles[j].Path
	}) {
		t.Error("test files not sorted by path")
	}

	if !sort.SliceIsSorted(snap.Signals, func(i, j int) bool {
		a, b := snap.Signals[i], snap.Signals[j]
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		if a.Type != b.Type {
			return a.Type < b.Type
		}
		if a.Location.File != b.Location.File {
			return a.Location.File < b.Location.File
		}
		return a.Location.Line < b.Location.Line
	}) {
		t.Error("signals not sorted in canonical order")
	}

	if !sort.SliceIsSorted(snap.Frameworks, func(i, j int) bool {
		return snap.Frameworks[i].Name < snap.Frameworks[j].Name
	}) {
		t.Error("frameworks not sorted by name")
	}
}
