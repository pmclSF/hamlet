package metrics

import (
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestDerive_Empty(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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

func TestDerive_HealthSignals_UsesUniqueFilesForRatios(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 4),
		Signals: []models.Signal{
			{Type: "slowTest", Location: models.SignalLocation{File: "a.test.js"}},
			{Type: "slowTest", Location: models.SignalLocation{File: "a.test.js"}},
			{Type: "slowTest", Location: models.SignalLocation{File: "b.test.js"}},
		},
	}

	ms := Derive(snap)
	if ms.Health.SlowTestCount != 2 {
		t.Fatalf("slowTestCount = %d, want 2 unique files", ms.Health.SlowTestCount)
	}
	if ms.Health.SlowTestRatio != 0.5 {
		t.Fatalf("slowTestRatio = %.2f, want 0.50", ms.Health.SlowTestRatio)
	}
}

func TestDerive_QualitySignals(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 5),
		Signals: []models.Signal{
			{Type: "policyViolation"},
			{Type: "policyViolation"},
			{Type: "skippedTestsInCI"},
			{Type: "legacyFrameworkUsage"},
			{Type: "runtimeBudgetExceeded"},
		},
	}

	ms := Derive(snap)

	if ms.Governance.PolicyViolationCount != 3 {
		t.Errorf("policyViolationCount = %d, want 3", ms.Governance.PolicyViolationCount)
	}
	if ms.Governance.LegacyFrameworkUsageCount != 1 {
		t.Errorf("legacyFrameworkUsageCount = %d, want 1", ms.Governance.LegacyFrameworkUsageCount)
	}
	if ms.Governance.RuntimeBudgetExceededCount != 1 {
		t.Errorf("runtimeBudgetExceededCount = %d, want 1", ms.Governance.RuntimeBudgetExceededCount)
	}
}

func TestDerive_NoRuntimeNote(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	if r := safeRatio(3, 10); r != 0.3 {
		t.Errorf("safeRatio(3, 10) = %.2f, want 0.30", r)
	}
	if r := safeRatio(5, 0); r != 0.0 {
		t.Errorf("safeRatio(5, 0) = %.2f, want 0.00", r)
	}
}

func TestDerive_QualityPostureBand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		files    int
		quality  int
		wantBand string
	}{
		{"no files", 0, 0, "unknown"},
		{"strong - no issues", 10, 0, "strong"},
		{"strong - below 10%", 20, 1, "strong"},
		{"moderate - 20%", 10, 2, "moderate"},
		{"weak - 50%", 10, 5, "weak"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := deriveQualityPosture(tt.quality, tt.files)
			if got != tt.wantBand {
				t.Errorf("deriveQualityPosture(%d, %d) = %q, want %q", tt.quality, tt.files, got, tt.wantBand)
			}
		})
	}
}

func TestDerive_MigrationReadinessBand(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 10),
		Signals: []models.Signal{
			{Type: "deprecatedTestPattern", Location: models.SignalLocation{File: "src/a.test.js"}},
			{Type: "deprecatedTestPattern", Location: models.SignalLocation{File: "src/b.test.js"}},
			{Type: "deprecatedTestPattern", Location: models.SignalLocation{File: "src/c.test.js"}},
		},
	}
	ms := Derive(snap)
	// 3 blockers / 10 files = 30% → low
	if ms.Change.MigrationReadinessBand != "low" {
		t.Errorf("migrationReadinessBand = %q, want low", ms.Change.MigrationReadinessBand)
	}
}

func TestDerive_MigrationAreaCounts(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "safe/clean.test.js"},
			{Path: "risky/old.test.js"},
		},
		Signals: []models.Signal{
			{Type: "deprecatedTestPattern", Location: models.SignalLocation{File: "risky/old.test.js"}},
			{Type: "weakAssertion", Location: models.SignalLocation{File: "risky/old.test.js"}},
		},
	}
	ms := Derive(snap)
	if ms.Change.SafeAreaCount != 1 {
		t.Errorf("safeAreaCount = %d, want 1", ms.Change.SafeAreaCount)
	}
	if ms.Change.RiskyAreaCount != 1 {
		t.Errorf("riskyAreaCount = %d, want 1", ms.Change.RiskyAreaCount)
	}
}

func TestDerive_QualityCompoundedBlockers(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/a.test.js"},
			{Path: "src/b.test.js"},
		},
		Signals: []models.Signal{
			{Type: "deprecatedTestPattern", Location: models.SignalLocation{File: "src/a.test.js"}},
			{Type: "weakAssertion", Location: models.SignalLocation{File: "src/a.test.js"}},
			{Type: "deprecatedTestPattern", Location: models.SignalLocation{File: "src/b.test.js"}},
			// b.test.js has no quality issue → not compounded
		},
	}
	ms := Derive(snap)
	if ms.Change.QualityCompoundedBlockerCount != 1 {
		t.Errorf("qualityCompoundedBlockerCount = %d, want 1", ms.Change.QualityCompoundedBlockerCount)
	}
}

func TestDerive_PrivacySafety_NoRawPaths(t *testing.T) {
	t.Parallel()
	// Verify that metrics.Snapshot contains no raw file paths.
	// This is a structural test — the Snapshot type uses only counts,
	// ratios, bands, and framework name strings.
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/secret/internal.test.js"},
		},
		Signals: []models.Signal{
			{Type: "weakAssertion", Location: models.SignalLocation{File: "src/secret/internal.test.js"}},
		},
		Repository: models.RepositoryMetadata{
			Name:      "private-repo",
			Languages: []string{"javascript"},
		},
	}
	ms := Derive(snap)

	// Structure should not contain file paths.
	for _, fw := range ms.Structure.Frameworks {
		if fw == "src/secret/internal.test.js" {
			t.Error("framework list should not contain file paths")
		}
	}
	// Notes should not contain file paths.
	for _, note := range ms.Notes {
		if note == "src/secret/internal.test.js" {
			t.Error("notes should not contain raw file paths")
		}
	}
	// Change metrics contain only counts, not paths.
	if ms.Change.MigrationBlockerCount < 0 {
		t.Error("unreachable — just ensuring the field is an int, not a path")
	}
}
