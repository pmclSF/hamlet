package benchmark

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/metrics"
	"github.com/pmclSF/terrain/internal/models"
)

func TestBuildExport_Basic(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 10),
		Frameworks: []models.Framework{
			{Name: "jest", FileCount: 8},
			{Name: "mocha", FileCount: 2},
		},
		Repository: models.RepositoryMetadata{
			Languages: []string{"javascript", "typescript"},
		},
	}
	ms := metrics.Derive(snap)
	exp := BuildExport(snap, ms, false)

	if exp.SchemaVersion != "3" {
		t.Errorf("schemaVersion = %q, want 3", exp.SchemaVersion)
	}
	if exp.Segment.PrimaryLanguage != "javascript" {
		t.Errorf("primaryLanguage = %q, want javascript", exp.Segment.PrimaryLanguage)
	}
	if exp.Segment.PrimaryFramework != "jest" {
		t.Errorf("primaryFramework = %q, want jest", exp.Segment.PrimaryFramework)
	}
	if exp.Segment.TestFileBucket != "small" {
		t.Errorf("testFileBucket = %q, want small", exp.Segment.TestFileBucket)
	}
	if exp.Segment.FrameworkCount != 2 {
		t.Errorf("frameworkCount = %d, want 2", exp.Segment.FrameworkCount)
	}
	if exp.Segment.HasPolicy {
		t.Error("hasPolicy should be false")
	}
}

func TestBuildExport_WithPolicy(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 10),
	}
	ms := metrics.Derive(snap)
	exp := BuildExport(snap, ms, true)

	if !exp.Segment.HasPolicy {
		t.Error("hasPolicy should be true")
	}
}

func TestSegment_TestFileBuckets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		count int
		want  string
	}{
		{0, "small"},
		{49, "small"},
		{50, "medium"},
		{500, "medium"},
		{501, "large"},
	}

	for _, tt := range tests {
		snap := &models.TestSuiteSnapshot{
			TestFiles: make([]models.TestFile, tt.count),
		}
		ms := metrics.Derive(snap)
		exp := BuildExport(snap, ms, false)
		if exp.Segment.TestFileBucket != tt.want {
			t.Errorf("count=%d: bucket = %q, want %q", tt.count, exp.Segment.TestFileBucket, tt.want)
		}
	}
}

func TestSegment_RuntimeDetection(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "a.test.js", RuntimeStats: &models.RuntimeStats{AvgRuntimeMs: 100}},
		},
	}
	ms := metrics.Derive(snap)
	exp := BuildExport(snap, ms, false)

	if !exp.Segment.HasRuntimeData {
		t.Error("hasRuntimeData should be true")
	}
}

func TestSegment_NoRuntime(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "a.test.js"},
		},
	}
	ms := metrics.Derive(snap)
	exp := BuildExport(snap, ms, false)

	if exp.Segment.HasRuntimeData {
		t.Error("hasRuntimeData should be false")
	}
}

func TestSegment_CoverageDetection(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 5),
		Signals: []models.Signal{
			{Type: "coverageThresholdBreak"},
		},
	}
	ms := metrics.Derive(snap)
	exp := BuildExport(snap, ms, false)

	if !exp.Segment.HasCoverage {
		t.Error("hasCoverage should be true")
	}
}

func TestExport_MetricsIncluded(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 5),
		Signals: []models.Signal{
			{Type: "weakAssertion"},
			{Type: "weakAssertion"},
		},
	}
	ms := metrics.Derive(snap)
	exp := BuildExport(snap, ms, false)

	if exp.Metrics.Quality.WeakAssertionCount != 2 {
		t.Errorf("metrics weakAssertionCount = %d, want 2", exp.Metrics.Quality.WeakAssertionCount)
	}
	if exp.Metrics.Structure.TotalTestFiles != 5 {
		t.Errorf("metrics totalTestFiles = %d, want 5", exp.Metrics.Structure.TotalTestFiles)
	}
}

func TestExport_MigrationPostureIncluded(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "safe/a.test.js"},
			{Path: "risky/b.test.js"},
		},
		Signals: []models.Signal{
			{Type: "deprecatedTestPattern", Location: models.SignalLocation{File: "risky/b.test.js"}},
			{Type: "weakAssertion", Location: models.SignalLocation{File: "risky/b.test.js"}},
		},
	}
	ms := metrics.Derive(snap)
	exp := BuildExport(snap, ms, false)

	if exp.Metrics.Change.SafeAreaCount != 1 {
		t.Errorf("safeAreaCount = %d, want 1", exp.Metrics.Change.SafeAreaCount)
	}
	if exp.Metrics.Change.RiskyAreaCount != 1 {
		t.Errorf("riskyAreaCount = %d, want 1", exp.Metrics.Change.RiskyAreaCount)
	}
	if exp.Metrics.Quality.QualityPostureBand == "" {
		t.Error("qualityPostureBand should not be empty")
	}
}

func TestExport_PrivacySafety(t *testing.T) {
	t.Parallel()
	// The export should contain no raw file paths, symbol names, or test names.
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/internal/secret.test.js"},
		},
		Signals: []models.Signal{
			{
				Type:        "weakAssertion",
				Location:    models.SignalLocation{File: "src/internal/secret.test.js"},
				Explanation: "Weak assertions in src/internal/secret.test.js",
			},
		},
		Repository: models.RepositoryMetadata{
			Name:      "my-private-repo",
			Languages: []string{"javascript"},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "processPayment", Path: "src/internal/billing.js"},
		},
	}
	ms := metrics.Derive(snap)
	exp := BuildExport(snap, ms, false)

	// Export should only contain segment + metrics, no raw snapshot data.
	// Segment contains only language, framework, bucket — no paths.
	if exp.Segment.PrimaryLanguage != "javascript" {
		t.Errorf("primaryLanguage = %q, want javascript", exp.Segment.PrimaryLanguage)
	}

	// Metrics contain only counts and bands, no paths.
	if exp.Metrics.Structure.TotalTestFiles != 1 {
		t.Errorf("totalTestFiles = %d, want 1", exp.Metrics.Structure.TotalTestFiles)
	}

	// SchemaVersion should be 3.
	if exp.SchemaVersion != "3" {
		t.Errorf("schemaVersion = %q, want 3", exp.SchemaVersion)
	}

	// Serialize to JSON and verify no raw file paths or symbols leaked.
	data, err := json.Marshal(exp)
	if err != nil {
		t.Fatalf("json marshal: %v", err)
	}
	jsonStr := string(data)

	prohibited := []string{
		"src/internal/secret.test.js",
		"processPayment",
		"src/internal/billing.js",
		"my-private-repo",
	}
	for _, p := range prohibited {
		if strings.Contains(jsonStr, p) {
			t.Errorf("export JSON contains prohibited string %q — privacy leak", p)
		}
	}
}
