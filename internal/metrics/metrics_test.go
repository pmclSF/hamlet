package metrics

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestDerive_Empty(t *testing.T) {
	snap := &models.TestSuiteSnapshot{}
	ms := Derive(snap)

	if ms.Structure.TotalTestFiles != 0 {
		t.Errorf("totalTestFiles = %d, want 0", ms.Structure.TotalTestFiles)
	}
	if len(ms.Notes) == 0 {
		t.Error("expected notes for empty snapshot")
	}
}

func TestDerive_Structure(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "a.test.js"},
			{Path: "b.test.js"},
			{Path: "c.test.js"},
		},
		Frameworks: []models.Framework{
			{Name: "jest"},
			{Name: "mocha"},
		},
		Repository: models.RepositoryMetadata{
			Languages: []string{"javascript"},
		},
	}

	ms := Derive(snap)

	if ms.Structure.TotalTestFiles != 3 {
		t.Errorf("totalTestFiles = %d, want 3", ms.Structure.TotalTestFiles)
	}
	if ms.Structure.FrameworkCount != 2 {
		t.Errorf("frameworkCount = %d, want 2", ms.Structure.FrameworkCount)
	}
	if len(ms.Structure.Frameworks) != 2 {
		t.Errorf("frameworks len = %d, want 2", len(ms.Structure.Frameworks))
	}
	// 2 frameworks / 3 files ≈ 0.667
	if ms.Structure.FrameworkFragmentationRatio < 0.66 || ms.Structure.FrameworkFragmentationRatio > 0.67 {
		t.Errorf("fragmentation = %.3f, want ~0.667", ms.Structure.FrameworkFragmentationRatio)
	}
	if len(ms.Structure.Languages) != 1 || ms.Structure.Languages[0] != "javascript" {
		t.Errorf("languages = %v, want [javascript]", ms.Structure.Languages)
	}
}

func TestDerive_HealthSignals(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 10),
		Signals: []models.Signal{
			{Type: "flakyTest"},
			{Type: "flakyTest"},
			{Type: "skippedTest"},
			{Type: "slowTest"},
			{Type: "deadTest"},
		},
	}

	ms := Derive(snap)

	if ms.Health.FlakyTestCount != 2 {
		t.Errorf("flakyTestCount = %d, want 2", ms.Health.FlakyTestCount)
	}
	if ms.Health.FlakyTestRatio != 0.2 {
		t.Errorf("flakyTestRatio = %.2f, want 0.20", ms.Health.FlakyTestRatio)
	}
	if ms.Health.SkippedTestCount != 1 {
		t.Errorf("skippedTestCount = %d, want 1", ms.Health.SkippedTestCount)
	}
	if ms.Health.SlowTestCount != 1 {
		t.Errorf("slowTestCount = %d, want 1", ms.Health.SlowTestCount)
	}
	if ms.Health.DeadTestCount != 1 {
		t.Errorf("deadTestCount = %d, want 1", ms.Health.DeadTestCount)
	}
}

func TestDerive_QualitySignals(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 5),
		Signals: []models.Signal{
			{Type: "weakAssertion"},
			{Type: "weakAssertion"},
			{Type: "mockHeavyTest"},
			{Type: "untestedExport"},
			{Type: "coverageThresholdBreak"},
		},
	}

	ms := Derive(snap)

	if ms.Quality.WeakAssertionCount != 2 {
		t.Errorf("weakAssertionCount = %d, want 2", ms.Quality.WeakAssertionCount)
	}
	if ms.Quality.WeakAssertionRatio != 0.4 {
		t.Errorf("weakAssertionRatio = %.2f, want 0.40", ms.Quality.WeakAssertionRatio)
	}
	if ms.Quality.MockHeavyTestCount != 1 {
		t.Errorf("mockHeavyTestCount = %d, want 1", ms.Quality.MockHeavyTestCount)
	}
	if ms.Quality.UntestedExportCount != 1 {
		t.Errorf("untestedExportCount = %d, want 1", ms.Quality.UntestedExportCount)
	}
	if ms.Quality.CoverageThresholdBreakCount != 1 {
		t.Errorf("coverageThresholdBreakCount = %d, want 1", ms.Quality.CoverageThresholdBreakCount)
	}
}

func TestDerive_ChangeReadiness(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 5),
		Signals: []models.Signal{
			{Type: "migrationBlocker", Metadata: map[string]any{"blockerType": "custom-matcher"}},
			{Type: "deprecatedTestPattern", Metadata: map[string]any{"blockerType": "deprecated-pattern"}},
			{Type: "deprecatedTestPattern", Metadata: map[string]any{"blockerType": "deprecated-pattern"}},
			{Type: "dynamicTestGeneration"},
			{Type: "customMatcherRisk"},
		},
	}

	ms := Derive(snap)

	if ms.Change.MigrationBlockerCount != 1 {
		t.Errorf("migrationBlockerCount = %d, want 1", ms.Change.MigrationBlockerCount)
	}
	if ms.Change.DeprecatedPatternCount != 2 {
		t.Errorf("deprecatedPatternCount = %d, want 2", ms.Change.DeprecatedPatternCount)
	}
	if ms.Change.DynamicGenerationCount != 1 {
		t.Errorf("dynamicGenerationCount = %d, want 1", ms.Change.DynamicGenerationCount)
	}
	if ms.Change.CustomMatcherRiskCount != 1 {
		t.Errorf("customMatcherRiskCount = %d, want 1", ms.Change.CustomMatcherRiskCount)
	}
	if ms.Change.BlockerCountByType["custom-matcher"] != 1 {
		t.Errorf("blockersByType[custom-matcher] = %d, want 1", ms.Change.BlockerCountByType["custom-matcher"])
	}
	if ms.Change.BlockerCountByType["deprecated-pattern"] != 2 {
		t.Errorf("blockersByType[deprecated-pattern] = %d, want 2", ms.Change.BlockerCountByType["deprecated-pattern"])
	}
}

func TestDerive_RiskMetrics(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 5),
		Risk: []models.RiskSurface{
			{Type: "reliability", Scope: "repository", Band: models.RiskBandMedium},
			{Type: "change", Scope: "repository", Band: models.RiskBandHigh},
			{Type: "change", Scope: "directory", ScopeName: "src/auth", Band: models.RiskBandCritical},
		},
		Signals: []models.Signal{
			{Type: "weakAssertion", Severity: models.SeverityCritical},
		},
	}

	ms := Derive(snap)

	if ms.Risk.ReliabilityBand != "medium" {
		t.Errorf("reliabilityBand = %q, want medium", ms.Risk.ReliabilityBand)
	}
	if ms.Risk.ChangeBand != "high" {
		t.Errorf("changeBand = %q, want high", ms.Risk.ChangeBand)
	}
	if ms.Risk.HighRiskAreaCount != 2 {
		t.Errorf("highRiskAreaCount = %d, want 2", ms.Risk.HighRiskAreaCount)
	}
	if ms.Risk.CriticalFindingCount != 1 {
		t.Errorf("criticalFindingCount = %d, want 1", ms.Risk.CriticalFindingCount)
	}
}

func TestDerive_GovernanceSignals(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 5),
		Signals: []models.Signal{
			{Type: "policyViolation"},
			{Type: "policyViolation"},
			{Type: "legacyFrameworkUsage"},
			{Type: "runtimeBudgetExceeded"},
		},
	}

	ms := Derive(snap)

	if ms.Governance.PolicyViolationCount != 2 {
		t.Errorf("policyViolationCount = %d, want 2", ms.Governance.PolicyViolationCount)
	}
	if ms.Governance.LegacyFrameworkUsageCount != 1 {
		t.Errorf("legacyFrameworkUsageCount = %d, want 1", ms.Governance.LegacyFrameworkUsageCount)
	}
	if ms.Governance.RuntimeBudgetExceededCount != 1 {
		t.Errorf("runtimeBudgetExceededCount = %d, want 1", ms.Governance.RuntimeBudgetExceededCount)
	}
}

func TestDerive_NoRuntimeNote(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "a.test.js"},
		},
	}

	ms := Derive(snap)

	found := false
	for _, note := range ms.Notes {
		if note == "No runtime artifacts detected; health metrics are static-analysis only." {
			found = true
		}
	}
	if !found {
		t.Error("expected no-runtime note")
	}
}

func TestDerive_WithRuntime(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:         "a.test.js",
				RuntimeStats: &models.RuntimeStats{AvgRuntimeMs: 100},
			},
		},
	}

	ms := Derive(snap)

	for _, note := range ms.Notes {
		if note == "No runtime artifacts detected; health metrics are static-analysis only." {
			t.Error("should not have no-runtime note when runtime data exists")
		}
	}
}

func TestSafeRatio(t *testing.T) {
	if r := safeRatio(3, 10); r != 0.3 {
		t.Errorf("safeRatio(3, 10) = %.2f, want 0.30", r)
	}
	if r := safeRatio(5, 0); r != 0.0 {
		t.Errorf("safeRatio(5, 0) = %.2f, want 0.00", r)
	}
}
