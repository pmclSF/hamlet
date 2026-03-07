package benchmark

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/metrics"
	"github.com/pmclSF/hamlet/internal/models"
)

func TestBuildExport_Basic(t *testing.T) {
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

	if exp.SchemaVersion != "1" {
		t.Errorf("schemaVersion = %q, want 1", exp.SchemaVersion)
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
