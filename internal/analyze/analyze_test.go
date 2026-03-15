package analyze

import (
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestBuild_EmptySnapshot(t *testing.T) {
	report := Build(&BuildInput{Snapshot: &models.TestSuiteSnapshot{}})
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.TestsDetected.TestFileCount != 0 {
		t.Errorf("expected 0 test files, got %d", report.TestsDetected.TestFileCount)
	}
	if report.TopInsight == "" {
		t.Error("expected non-empty top insight even with no data")
	}
}

func TestBuildSignalSummary(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		Signals: []models.Signal{
			{Severity: models.SeverityCritical, Category: models.CategoryHealth},
			{Severity: models.SeverityHigh, Category: models.CategoryHealth},
			{Severity: models.SeverityMedium, Category: models.CategoryQuality},
			{Severity: models.SeverityLow, Category: models.CategoryQuality},
			{Severity: models.SeverityLow, Category: models.CategoryMigration},
		},
	}
	sb := buildSignalSummary(snap)
	if sb.Total != 5 {
		t.Errorf("total = %d, want 5", sb.Total)
	}
	if sb.Critical != 1 {
		t.Errorf("critical = %d, want 1", sb.Critical)
	}
	if sb.High != 1 {
		t.Errorf("high = %d, want 1", sb.High)
	}
	if sb.ByCategory["health"] != 2 {
		t.Errorf("health = %d, want 2", sb.ByCategory["health"])
	}
}

func TestBuildSkipSummary(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{TestCount: 10},
			{TestCount: 5},
		},
		Signals: []models.Signal{
			{Type: "skippedTest"},
			{Type: "skippedTest"},
			{Type: "weakAssertion"},
		},
	}
	ss := buildSkipSummary(snap)
	if ss.SkippedCount != 2 {
		t.Errorf("skipped = %d, want 2", ss.SkippedCount)
	}
	if ss.TotalTests != 15 {
		t.Errorf("total = %d, want 15", ss.TotalTests)
	}
}

func TestBuildLimitations(t *testing.T) {
	snap := &models.TestSuiteSnapshot{}
	lims := buildLimitations(snap, false)
	if len(lims) == 0 {
		t.Error("expected limitations for empty snapshot")
	}
}

func TestBuild_WithSignals(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{Name: "test"},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 2},
		},
		TestFiles: []models.TestFile{
			{Path: "test/a.test.js", Framework: "jest", TestCount: 3},
			{Path: "test/b.test.js", Framework: "jest", TestCount: 2},
		},
		Signals: []models.Signal{
			{Type: "weakAssertion", Category: models.CategoryQuality, Severity: models.SeverityMedium},
		},
	}
	report := Build(&BuildInput{Snapshot: snap})
	if report.TestsDetected.TestFileCount != 2 {
		t.Errorf("test files = %d, want 2", report.TestsDetected.TestFileCount)
	}
	if report.SignalSummary.Total != 1 {
		t.Errorf("signals = %d, want 1", report.SignalSummary.Total)
	}
	if len(report.TestsDetected.Frameworks) != 1 {
		t.Errorf("frameworks = %d, want 1", len(report.TestsDetected.Frameworks))
	}
}

func TestBuild_TopInsight(t *testing.T) {
	snap := smallRepoSnapshot()
	report := Build(&BuildInput{Snapshot: snap})
	if report.TopInsight == "" {
		t.Error("expected non-empty top insight")
	}
}
