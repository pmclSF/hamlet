package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/analysis"
	"github.com/pmclSF/terrain/internal/measurement"
	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/scoring"
	"github.com/pmclSF/terrain/internal/testdata"
)

// TestPipelineSteps_Integration verifies each pipeline step works with
// standardized fixtures without requiring filesystem access.
func TestPipelineSteps_Integration(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	if meta.MethodologyFingerprint == "" {
		t.Error("expected non-empty methodology fingerprint")
	}
}

func TestRunPipeline_EngineVersionStamp_Default(t *testing.T) {
	t.Parallel()
	result, err := RunPipeline("../analysis/testdata/sample-repo")
	if err != nil {
		t.Fatalf("RunPipeline failed: %v", err)
	}
	if result.Snapshot.SnapshotMeta.EngineVersion != DefaultEngineVersion {
		t.Fatalf("engine version = %q, want %q", result.Snapshot.SnapshotMeta.EngineVersion, DefaultEngineVersion)
	}
}

func TestRunPipeline_EngineVersionStamp_FromOptions(t *testing.T) {
	t.Parallel()
	result, err := RunPipeline("../analysis/testdata/sample-repo", PipelineOptions{
		EngineVersion: "engine-test-version",
	})
	if err != nil {
		t.Fatalf("RunPipeline failed: %v", err)
	}
	if result.Snapshot.SnapshotMeta.EngineVersion != "engine-test-version" {
		t.Fatalf("engine version = %q, want %q", result.Snapshot.SnapshotMeta.EngineVersion, "engine-test-version")
	}
}

// Verify that analysis.New returns something usable even for a nonexistent repo.
func TestAnalyzerNewDoesNotPanic(t *testing.T) {
	t.Parallel()
	a := analysis.New("/nonexistent/path")
	if a == nil {
		t.Error("expected non-nil analyzer")
	}
}

// TestPipelineDeterminism verifies that running the pipeline twice on
// identical input produces byte-identical JSON output (excluding timestamps).
func TestPipelineDeterminism(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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

func TestAttachSignalsToTestFiles(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/a.test.js"},
			{Path: "test/b.test.js"},
		},
		Signals: []models.Signal{
			{
				Type:        "weakAssertion",
				Category:    models.CategoryQuality,
				Severity:    models.SeverityMedium,
				Location:    models.SignalLocation{File: "test/a.test.js"},
				Explanation: "weak assertions",
			},
			{
				Type:        "policyViolation",
				Category:    models.CategoryGovernance,
				Severity:    models.SeverityHigh,
				Location:    models.SignalLocation{Repository: "repo"},
				Explanation: "repo-level policy violation",
			},
		},
	}

	attachSignalsToTestFiles(snap)
	if got := len(snap.TestFiles[0].Signals); got != 1 {
		t.Fatalf("test/a.test.js file signals = %d, want 1", got)
	}
	if got := len(snap.TestFiles[1].Signals); got != 0 {
		t.Fatalf("test/b.test.js file signals = %d, want 0", got)
	}
}

func TestPopulateSnapshotMetadata(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/a.test.js", LinkedCodeUnits: []string{"src/a.js:A"}, Signals: []models.Signal{{Type: "weakAssertion"}}},
			{Path: "test/b.test.js"},
		},
		Signals: []models.Signal{
			{
				Type:        "weakAssertion",
				Category:    models.CategoryQuality,
				Severity:    models.SeverityMedium,
				Explanation: "weak assertions",
			},
		},
		DataSources: []models.DataSource{
			{Name: "runtime", Status: models.DataSourceAvailable},
			{Name: "coverage", Status: models.DataSourceError},
			{Name: "policy", Status: models.DataSourceUnavailable},
		},
	}

	populateSnapshotMetadata(snap, PipelineOptions{
		CoveragePath: "/tmp/coverage.json",
		RuntimePaths: []string{"/tmp/runtime.xml"},
	}, true)

	if snap.Metadata == nil {
		t.Fatal("expected metadata to be populated")
	}
	if got, ok := snap.Metadata["runtimeArtifactsProvided"].(int); !ok || got != 1 {
		t.Fatalf("runtimeArtifactsProvided = %#v, want 1", snap.Metadata["runtimeArtifactsProvided"])
	}
	if got, ok := snap.Metadata["testFilesWithLinkedUnits"].(int); !ok || got != 1 {
		t.Fatalf("testFilesWithLinkedUnits = %#v, want 1", snap.Metadata["testFilesWithLinkedUnits"])
	}
	if got, ok := snap.Metadata["dataSourcesError"].(int); !ok || got != 1 {
		t.Fatalf("dataSourcesError = %#v, want 1", snap.Metadata["dataSourcesError"])
	}
}

func TestNormalizeSignalMetadata(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{
		Signals: []models.Signal{
			{
				Type:           "weakAssertion",
				Category:       models.CategoryQuality,
				Severity:       models.SeverityMedium,
				Confidence:     0.7,
				EvidenceSource: models.SourceStructuralPattern,
				Location:       models.SignalLocation{File: "test/a.test.js"},
			},
		},
	}

	normalizeSignalMetadata(snap)
	md := snap.Signals[0].Metadata
	if md == nil {
		t.Fatal("expected metadata map to be populated")
	}
	if md["signalType"] != "weakAssertion" {
		t.Fatalf("signalType = %v, want weakAssertion", md["signalType"])
	}
	if md["scope"] != "file" {
		t.Fatalf("scope = %v, want file", md["scope"])
	}
	if md["confidence"] != 0.7 {
		t.Fatalf("confidence = %v, want 0.7", md["confidence"])
	}
}

func TestCompactRiskContributingSignals(t *testing.T) {
	t.Parallel()

	full := models.Signal{
		Type:             "weakAssertion",
		Category:         models.CategoryQuality,
		Severity:         models.SeverityMedium,
		Confidence:       0.8,
		EvidenceStrength: models.EvidenceStrong,
		EvidenceSource:   models.SourceStructuralPattern,
		Location:         models.SignalLocation{File: "test/a.test.js"},
		Explanation:      "weak assertions",
		SuggestedAction:  "add assertions",
		Metadata:         map[string]any{"k": "v"},
	}
	snap := &models.TestSuiteSnapshot{
		Risk: []models.RiskSurface{
			{
				Type:                "change",
				Scope:               "repository",
				ScopeName:           "repo",
				ContributingSignals: []models.Signal{full, full},
			},
		},
	}

	compactRiskContributingSignals(snap)
	contrib := snap.Risk[0].ContributingSignals
	if len(contrib) != 1 {
		t.Fatalf("expected deduped contributing signals length 1, got %d", len(contrib))
	}
	if contrib[0].Metadata != nil {
		t.Fatalf("expected compact contributing signal metadata to be nil, got %v", contrib[0].Metadata)
	}
	if contrib[0].SuggestedAction != "" {
		t.Fatalf("expected compact contributing signal suggested action empty, got %q", contrib[0].SuggestedAction)
	}
}

func TestDeriveDataCompleteness(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{{Path: "test/a.test.js"}},
		DataSources: []models.DataSource{
			{Name: "coverage", Status: models.DataSourceAvailable},
			{Name: "runtime", Status: models.DataSourceUnavailable},
			{Name: "policy", Status: models.DataSourceAvailable},
		},
	}
	dc := deriveDataCompleteness(snap, PipelineOptions{
		CoveragePath: "/tmp/coverage",
		RuntimePaths: []string{"/tmp/runtime"},
	}, true)

	if !dc.SourceAvailable {
		t.Fatal("expected source to be available")
	}
	if !dc.CoverageProvided || !dc.CoverageAvailable {
		t.Fatalf("expected coverage provided+available, got %+v", dc)
	}
	if !dc.RuntimeProvided || dc.RuntimeAvailable {
		t.Fatalf("expected runtime provided but unavailable, got %+v", dc)
	}
	if !dc.PolicyAvailable {
		t.Fatalf("expected policy available, got %+v", dc)
	}
}

func TestValidatePipelineOptions_CoverageRunLabel(t *testing.T) {
	t.Parallel()

	if err := validatePipelineOptions(".", PipelineOptions{CoverageRunLabel: "integration"}); err != nil {
		t.Fatalf("expected integration label to be valid, got %v", err)
	}
	if err := validatePipelineOptions(".", PipelineOptions{CoverageRunLabel: "E2E"}); err != nil {
		t.Fatalf("expected case-insensitive e2e label to be valid, got %v", err)
	}
	if err := validatePipelineOptions(".", PipelineOptions{CoverageRunLabel: "smoke"}); err == nil {
		t.Fatal("expected invalid coverage run label to return error")
	}
}

func TestRunPipeline_GoldenRepos(t *testing.T) {
	t.Parallel()

	type expected struct {
		TestFiles       int     `json:"testFiles"`
		CodeUnits       int     `json:"codeUnits"`
		Signals         int     `json:"signals"`
		RiskSurfaces    int     `json:"riskSurfaces"`
		MaxRiskScore    float64 `json:"maxRiskScore"`
		SourceAvailable bool    `json:"sourceAvailable"`
		CoverageStatus  string  `json:"coverageStatus"`
		RuntimeStatus   string  `json:"runtimeStatus"`
		CoverageSummary bool    `json:"coverageSummary"`
	}
	var expectedByRepo map[string]expected
	data, err := os.ReadFile(filepath.Join("testdata", "golden_repos", "expectations.json"))
	if err != nil {
		t.Fatalf("read expectations: %v", err)
	}
	if err := json.Unmarshal(data, &expectedByRepo); err != nil {
		t.Fatalf("unmarshal expectations: %v", err)
	}

	tests := []struct {
		name           string
		coverageRel    string
		runtimeRelList []string
	}{
		{name: "minimal-js"},
		{
			name:           "go-with-runtime-and-coverage",
			coverageRel:    "coverage.lcov",
			runtimeRelList: []string{"junit.xml"},
		},
		{name: "empty"},
		{
			name:           "degraded-data",
			coverageRel:    "broken-coverage.json",
			runtimeRelList: []string{"broken-runtime.xml"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := filepath.Join("testdata", "golden_repos", tt.name)
			opt := PipelineOptions{}
			if tt.coverageRel != "" {
				opt.CoveragePath = filepath.Join(root, tt.coverageRel)
			}
			if len(tt.runtimeRelList) > 0 {
				opt.RuntimePaths = make([]string, 0, len(tt.runtimeRelList))
				for _, rel := range tt.runtimeRelList {
					opt.RuntimePaths = append(opt.RuntimePaths, filepath.Join(root, rel))
				}
			}
			result, err := RunPipeline(root, opt)
			if err != nil {
				t.Fatalf("RunPipeline(%s) failed: %v", tt.name, err)
			}
			want := expectedByRepo[tt.name]
			if got := len(result.Snapshot.TestFiles); got != want.TestFiles {
				t.Fatalf("test files = %d, want %d", got, want.TestFiles)
			}
			if got := len(result.Snapshot.CodeUnits); got != want.CodeUnits {
				t.Fatalf("code units = %d, want %d", got, want.CodeUnits)
			}
			if got := len(result.Snapshot.Signals); got != want.Signals {
				t.Fatalf("signals = %d, want %d", got, want.Signals)
			}
			if got := len(result.Snapshot.Risk); got != want.RiskSurfaces {
				t.Fatalf("risk surfaces = %d, want %d", got, want.RiskSurfaces)
			}
			maxRiskScore := 0.0
			for _, rs := range result.Snapshot.Risk {
				if rs.Score > maxRiskScore {
					maxRiskScore = rs.Score
				}
			}
			if maxRiskScore != want.MaxRiskScore {
				t.Fatalf("max risk score = %.4f, want %.4f", maxRiskScore, want.MaxRiskScore)
			}
			if result.DataCompleteness.SourceAvailable != want.SourceAvailable {
				t.Fatalf("source available = %v, want %v", result.DataCompleteness.SourceAvailable, want.SourceAvailable)
			}
			if status := dataSourceStatus(result.Snapshot, "coverage"); status != want.CoverageStatus {
				t.Fatalf("coverage status = %q, want %q", status, want.CoverageStatus)
			}
			if status := dataSourceStatus(result.Snapshot, "runtime"); status != want.RuntimeStatus {
				t.Fatalf("runtime status = %q, want %q", status, want.RuntimeStatus)
			}
			hasCoverageSummary := result.Snapshot.CoverageSummary != nil
			if hasCoverageSummary != want.CoverageSummary {
				t.Fatalf("coverage summary present = %v, want %v", hasCoverageSummary, want.CoverageSummary)
			}
			if hasCoverageSummary && tt.name == "go-with-runtime-and-coverage" && result.Snapshot.CoverageSummary.CoveredByUnitTests == 0 {
				t.Fatalf("coveredByUnitTests = %d, want > 0", result.Snapshot.CoverageSummary.CoveredByUnitTests)
			}
		})
	}
}

func dataSourceStatus(snapshot *models.TestSuiteSnapshot, name string) string {
	if snapshot == nil {
		return ""
	}
	for _, ds := range snapshot.DataSources {
		if strings.EqualFold(ds.Name, name) {
			return ds.Status
		}
	}
	return ""
}
