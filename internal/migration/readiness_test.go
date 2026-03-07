package migration

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestComputeReadiness_NoBlockers(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/a.test.js"},
			{Path: "test/b.test.js"},
		},
		Signals: []models.Signal{
			{Type: "weakAssertion"},
		},
	}

	r := ComputeReadiness(snap)
	if r.ReadinessLevel != "high" {
		t.Errorf("readiness = %q, want high", r.ReadinessLevel)
	}
	if r.TotalBlockers != 0 {
		t.Errorf("totalBlockers = %d, want 0", r.TotalBlockers)
	}
}

func TestComputeReadiness_FewBlockers(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 20),
		Signals: []models.Signal{
			{Type: "deprecatedTestPattern", Metadata: map[string]any{"blockerType": "deprecated-pattern"}},
		},
	}

	r := ComputeReadiness(snap)
	if r.ReadinessLevel != "high" {
		t.Errorf("readiness = %q, want high (1/20 = 5%%)", r.ReadinessLevel)
	}
}

func TestComputeReadiness_MediumBlockers(t *testing.T) {
	signals := make([]models.Signal, 0)
	for i := 0; i < 4; i++ {
		signals = append(signals, models.Signal{
			Type:     "deprecatedTestPattern",
			Metadata: map[string]any{"blockerType": "deprecated-pattern"},
		})
	}

	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 20),
		Signals:   signals,
	}

	r := ComputeReadiness(snap)
	if r.ReadinessLevel != "medium" {
		t.Errorf("readiness = %q, want medium (4/20 = 20%%)", r.ReadinessLevel)
	}
}

func TestComputeReadiness_ManyBlockers(t *testing.T) {
	signals := make([]models.Signal, 0)
	for i := 0; i < 8; i++ {
		signals = append(signals, models.Signal{
			Type:     "migrationBlocker",
			Metadata: map[string]any{"blockerType": "custom-matcher"},
		})
	}

	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 10),
		Signals:   signals,
	}

	r := ComputeReadiness(snap)
	if r.ReadinessLevel != "low" {
		t.Errorf("readiness = %q, want low (8/10 = 80%%)", r.ReadinessLevel)
	}
}

func TestComputeReadiness_NoTestFiles(t *testing.T) {
	snap := &models.TestSuiteSnapshot{}

	r := ComputeReadiness(snap)
	if r.ReadinessLevel != "unknown" {
		t.Errorf("readiness = %q, want unknown for empty repo", r.ReadinessLevel)
	}
}

func TestComputeReadiness_BlockersByType(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 10),
		Signals: []models.Signal{
			{Type: "deprecatedTestPattern", Metadata: map[string]any{"blockerType": "deprecated-pattern"}},
			{Type: "customMatcherRisk", Metadata: map[string]any{"blockerType": "custom-matcher"}},
			{Type: "customMatcherRisk", Metadata: map[string]any{"blockerType": "custom-matcher"}},
		},
	}

	r := ComputeReadiness(snap)
	if r.BlockersByType["custom-matcher"] != 2 {
		t.Errorf("custom-matcher count = %d, want 2", r.BlockersByType["custom-matcher"])
	}
	if r.BlockersByType["deprecated-pattern"] != 1 {
		t.Errorf("deprecated-pattern count = %d, want 1", r.BlockersByType["deprecated-pattern"])
	}
}

// --- Quality factor tests ---

func TestComputeReadiness_QualityFactors_WeakAssertionsInBlockerFiles(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/legacy.test.js"},
			{Path: "src/modern.test.js"},
		},
		Signals: []models.Signal{
			{
				Type:     "deprecatedTestPattern",
				Category: models.CategoryMigration,
				Location: models.SignalLocation{File: "src/legacy.test.js"},
				Metadata: map[string]any{"blockerType": "deprecated-pattern"},
			},
			{
				Type:     "weakAssertion",
				Category: models.CategoryQuality,
				Location: models.SignalLocation{File: "src/legacy.test.js"},
			},
		},
	}

	r := ComputeReadiness(snap)
	if len(r.QualityFactors) != 1 {
		t.Fatalf("qualityFactors count = %d, want 1", len(r.QualityFactors))
	}
	if r.QualityFactors[0].SignalType != "weakAssertion" {
		t.Errorf("signalType = %q, want weakAssertion", r.QualityFactors[0].SignalType)
	}
	if r.QualityFactors[0].AffectedFiles != 1 {
		t.Errorf("affectedFiles = %d, want 1", r.QualityFactors[0].AffectedFiles)
	}
}

func TestComputeReadiness_QualityFactors_NoOverlapNoFactors(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/legacy.test.js"},
			{Path: "src/modern.test.js"},
		},
		Signals: []models.Signal{
			{
				Type:     "deprecatedTestPattern",
				Category: models.CategoryMigration,
				Location: models.SignalLocation{File: "src/legacy.test.js"},
				Metadata: map[string]any{"blockerType": "deprecated-pattern"},
			},
			{
				Type:     "weakAssertion",
				Category: models.CategoryQuality,
				Location: models.SignalLocation{File: "src/modern.test.js"},
			},
		},
	}

	r := ComputeReadiness(snap)
	if len(r.QualityFactors) != 0 {
		t.Errorf("qualityFactors count = %d, want 0 (no overlap)", len(r.QualityFactors))
	}
}

func TestComputeReadiness_QualityFactors_MultipleTypes(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/old.test.js"},
		},
		Signals: []models.Signal{
			{
				Type:     "deprecatedTestPattern",
				Category: models.CategoryMigration,
				Location: models.SignalLocation{File: "src/old.test.js"},
				Metadata: map[string]any{"blockerType": "deprecated-pattern"},
			},
			{
				Type:     "weakAssertion",
				Category: models.CategoryQuality,
				Location: models.SignalLocation{File: "src/old.test.js"},
			},
			{
				Type:     "mockHeavyTest",
				Category: models.CategoryQuality,
				Location: models.SignalLocation{File: "src/old.test.js"},
			},
		},
	}

	r := ComputeReadiness(snap)
	if len(r.QualityFactors) != 2 {
		t.Fatalf("qualityFactors count = %d, want 2", len(r.QualityFactors))
	}
	// Should be sorted by affected count (tie) then alphabetical.
	types := make([]string, len(r.QualityFactors))
	for i, qf := range r.QualityFactors {
		types[i] = qf.SignalType
	}
	if types[0] != "mockHeavyTest" || types[1] != "weakAssertion" {
		t.Errorf("qualityFactors order = %v, want [mockHeavyTest, weakAssertion]", types)
	}
}

// --- Area assessment tests ---

func TestComputeReadiness_AreaAssessments_RiskyArea(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "legacy/old.test.js"},
		},
		Signals: []models.Signal{
			{
				Type:     "deprecatedTestPattern",
				Category: models.CategoryMigration,
				Location: models.SignalLocation{File: "legacy/old.test.js"},
				Metadata: map[string]any{"blockerType": "deprecated-pattern"},
			},
			{
				Type:     "weakAssertion",
				Category: models.CategoryQuality,
				Location: models.SignalLocation{File: "legacy/old.test.js"},
			},
		},
	}

	r := ComputeReadiness(snap)
	if len(r.AreaAssessments) != 1 {
		t.Fatalf("areaAssessments count = %d, want 1", len(r.AreaAssessments))
	}
	area := r.AreaAssessments[0]
	if area.Classification != "risky" {
		t.Errorf("classification = %q, want risky", area.Classification)
	}
	if area.MigrationBlockers != 1 {
		t.Errorf("migrationBlockers = %d, want 1", area.MigrationBlockers)
	}
	if area.QualityIssues != 1 {
		t.Errorf("qualityIssues = %d, want 1", area.QualityIssues)
	}
}

func TestComputeReadiness_AreaAssessments_SafeArea(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "modern/clean.test.js"},
		},
		Signals: []models.Signal{},
	}

	r := ComputeReadiness(snap)
	if len(r.AreaAssessments) != 1 {
		t.Fatalf("areaAssessments count = %d, want 1", len(r.AreaAssessments))
	}
	if r.AreaAssessments[0].Classification != "safe" {
		t.Errorf("classification = %q, want safe", r.AreaAssessments[0].Classification)
	}
}

func TestComputeReadiness_AreaAssessments_CautionBlockersOnly(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "mid/file.test.js"},
		},
		Signals: []models.Signal{
			{
				Type:     "deprecatedTestPattern",
				Category: models.CategoryMigration,
				Location: models.SignalLocation{File: "mid/file.test.js"},
				Metadata: map[string]any{"blockerType": "deprecated-pattern"},
			},
		},
	}

	r := ComputeReadiness(snap)
	if len(r.AreaAssessments) != 1 {
		t.Fatalf("areaAssessments count = %d, want 1", len(r.AreaAssessments))
	}
	if r.AreaAssessments[0].Classification != "caution" {
		t.Errorf("classification = %q, want caution", r.AreaAssessments[0].Classification)
	}
}

func TestComputeReadiness_AreaAssessments_CautionQualityOnly(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/file.test.js"},
		},
		Signals: []models.Signal{
			{
				Type:     "weakAssertion",
				Category: models.CategoryQuality,
				Location: models.SignalLocation{File: "src/file.test.js"},
			},
		},
	}

	r := ComputeReadiness(snap)
	if len(r.AreaAssessments) != 1 {
		t.Fatalf("areaAssessments count = %d, want 1", len(r.AreaAssessments))
	}
	if r.AreaAssessments[0].Classification != "caution" {
		t.Errorf("classification = %q, want caution", r.AreaAssessments[0].Classification)
	}
}

func TestComputeReadiness_AreaAssessments_MixedRepo(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "legacy/old.test.js"},
			{Path: "modern/clean.test.js"},
			{Path: "mid/ok.test.js"},
		},
		Signals: []models.Signal{
			{
				Type:     "deprecatedTestPattern",
				Category: models.CategoryMigration,
				Location: models.SignalLocation{File: "legacy/old.test.js"},
				Metadata: map[string]any{"blockerType": "deprecated-pattern"},
			},
			{
				Type:     "weakAssertion",
				Category: models.CategoryQuality,
				Location: models.SignalLocation{File: "legacy/old.test.js"},
			},
			{
				Type:     "deprecatedTestPattern",
				Category: models.CategoryMigration,
				Location: models.SignalLocation{File: "mid/ok.test.js"},
				Metadata: map[string]any{"blockerType": "deprecated-pattern"},
			},
		},
	}

	r := ComputeReadiness(snap)
	if len(r.AreaAssessments) != 3 {
		t.Fatalf("areaAssessments count = %d, want 3", len(r.AreaAssessments))
	}

	// Should be sorted: risky first, then caution, then safe.
	classifications := make([]string, len(r.AreaAssessments))
	for i, a := range r.AreaAssessments {
		classifications[i] = a.Classification
	}
	if classifications[0] != "risky" {
		t.Errorf("first area classification = %q, want risky", classifications[0])
	}
	if classifications[1] != "caution" {
		t.Errorf("second area classification = %q, want caution", classifications[1])
	}
	if classifications[2] != "safe" {
		t.Errorf("third area classification = %q, want safe", classifications[2])
	}
}

// --- Coverage guidance tests ---

func TestComputeReadiness_CoverageGuidance_HighPriorityForRiskyArea(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "legacy/old.test.js"},
		},
		Signals: []models.Signal{
			{
				Type:     "deprecatedTestPattern",
				Category: models.CategoryMigration,
				Location: models.SignalLocation{File: "legacy/old.test.js"},
				Metadata: map[string]any{"blockerType": "deprecated-pattern"},
			},
			{
				Type:     "weakAssertion",
				Category: models.CategoryQuality,
				Location: models.SignalLocation{File: "legacy/old.test.js"},
			},
		},
	}

	r := ComputeReadiness(snap)
	if len(r.CoverageGuidance) == 0 {
		t.Fatal("expected coverage guidance for risky area")
	}
	if r.CoverageGuidance[0].Priority != "high" {
		t.Errorf("priority = %q, want high", r.CoverageGuidance[0].Priority)
	}
	if r.CoverageGuidance[0].Directory != "legacy" {
		t.Errorf("directory = %q, want legacy", r.CoverageGuidance[0].Directory)
	}
}

func TestComputeReadiness_CoverageGuidance_NoneForSafeArea(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "modern/clean.test.js"},
		},
		Signals: []models.Signal{},
	}

	r := ComputeReadiness(snap)
	if len(r.CoverageGuidance) != 0 {
		t.Errorf("expected no coverage guidance for safe area, got %d", len(r.CoverageGuidance))
	}
}

func TestComputeReadiness_CoverageGuidance_UntestedExportsHighPriority(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/api.test.js"},
		},
		Signals: []models.Signal{
			{
				Type:     "deprecatedTestPattern",
				Category: models.CategoryMigration,
				Location: models.SignalLocation{File: "src/api.test.js"},
				Metadata: map[string]any{"blockerType": "deprecated-pattern"},
			},
			{
				Type:     "untestedExport",
				Category: models.CategoryQuality,
				Location: models.SignalLocation{File: "src/helpers.js"},
			},
		},
	}

	r := ComputeReadiness(snap)
	found := false
	for _, cg := range r.CoverageGuidance {
		if cg.Directory == "src" && cg.Priority == "high" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected high-priority coverage guidance for src/ due to untested exports + migration blocker")
	}
}

// --- Well-covered migration candidate (golden scenario) ---

func TestComputeReadiness_WellCoveredMigrationCandidate(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 40},
			{Name: "mocha", Type: models.FrameworkTypeUnit, FileCount: 10},
		},
		TestFiles: []models.TestFile{
			{Path: "src/auth/login.test.js", Framework: "jest", TestCount: 10, AssertionCount: 25},
			{Path: "src/auth/signup.test.js", Framework: "jest", TestCount: 8, AssertionCount: 20},
			{Path: "src/api/users.test.js", Framework: "jest", TestCount: 12, AssertionCount: 30},
			{Path: "src/legacy/old.test.js", Framework: "mocha", TestCount: 5, AssertionCount: 12},
		},
		Signals: []models.Signal{
			{
				Type:     "frameworkMigration",
				Category: models.CategoryMigration,
				Location: models.SignalLocation{Repository: "test-repo"},
			},
		},
	}

	r := ComputeReadiness(snap)

	// With 1 blocker across 4 files (25%), readiness is medium.
	// This is correct: frameworkMigration is a real blocker even in a well-covered repo.
	if r.ReadinessLevel != "medium" {
		t.Errorf("readiness = %q, want medium for well-covered repo with framework fragmentation", r.ReadinessLevel)
	}
	// No quality factors since frameworkMigration has no file location.
	if len(r.QualityFactors) != 0 {
		t.Errorf("qualityFactors = %d, want 0", len(r.QualityFactors))
	}
	// All areas should be safe (no per-file migration blockers).
	for _, area := range r.AreaAssessments {
		if area.Classification != "safe" {
			t.Errorf("area %s classification = %q, want safe", area.Directory, area.Classification)
		}
	}
}

// --- Shallowly tested migration risk (golden scenario) ---

func TestComputeReadiness_ShallowlyTestedMigrationRisk(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 5},
		},
		TestFiles: []models.TestFile{
			{Path: "src/core/engine.test.js", Framework: "jest", TestCount: 10, AssertionCount: 2, MockCount: 15},
			{Path: "src/core/parser.test.js", Framework: "jest", TestCount: 8, AssertionCount: 0},
			{Path: "src/util/helpers.test.js", Framework: "jest", TestCount: 3, AssertionCount: 8},
		},
		Signals: []models.Signal{
			{
				Type:     "deprecatedTestPattern",
				Category: models.CategoryMigration,
				Location: models.SignalLocation{File: "src/core/engine.test.js"},
				Metadata: map[string]any{"blockerType": "deprecated-pattern"},
			},
			{
				Type:     "dynamicTestGeneration",
				Category: models.CategoryMigration,
				Location: models.SignalLocation{File: "src/core/parser.test.js"},
				Metadata: map[string]any{"blockerType": "dynamic-generation"},
			},
			{
				Type:     "weakAssertion",
				Category: models.CategoryQuality,
				Location: models.SignalLocation{File: "src/core/engine.test.js"},
			},
			{
				Type:     "weakAssertion",
				Category: models.CategoryQuality,
				Location: models.SignalLocation{File: "src/core/parser.test.js"},
			},
			{
				Type:     "mockHeavyTest",
				Category: models.CategoryQuality,
				Location: models.SignalLocation{File: "src/core/engine.test.js"},
			},
			{
				Type:     "untestedExport",
				Category: models.CategoryQuality,
				Location: models.SignalLocation{File: "src/core/utils.js"},
			},
		},
	}

	r := ComputeReadiness(snap)

	// 2 blockers across 3 files = 67% → low readiness.
	if r.ReadinessLevel != "low" {
		t.Errorf("readiness = %q, want low for shallowly tested repo", r.ReadinessLevel)
	}

	// Should have quality factors: weakAssertion and mockHeavyTest overlap with blocker files.
	if len(r.QualityFactors) < 2 {
		t.Errorf("qualityFactors = %d, want >= 2", len(r.QualityFactors))
	}

	// src/core should be classified as risky.
	var coreArea *AreaAssessment
	for i, a := range r.AreaAssessments {
		if a.Directory == "src/core" {
			coreArea = &r.AreaAssessments[i]
			break
		}
	}
	if coreArea == nil {
		t.Fatal("expected area assessment for src/core")
	}
	if coreArea.Classification != "risky" {
		t.Errorf("src/core classification = %q, want risky", coreArea.Classification)
	}

	// src/util should be safe.
	var utilArea *AreaAssessment
	for i, a := range r.AreaAssessments {
		if a.Directory == "src/util" {
			utilArea = &r.AreaAssessments[i]
			break
		}
	}
	if utilArea == nil {
		t.Fatal("expected area assessment for src/util")
	}
	if utilArea.Classification != "safe" {
		t.Errorf("src/util classification = %q, want safe", utilArea.Classification)
	}

	// Coverage guidance should include src/core as high priority.
	foundHighCore := false
	for _, cg := range r.CoverageGuidance {
		if cg.Directory == "src/core" && cg.Priority == "high" {
			foundHighCore = true
			break
		}
	}
	if !foundHighCore {
		t.Error("expected high-priority coverage guidance for src/core")
	}
}

// --- Mixed framework with uneven coverage (golden scenario) ---

func TestComputeReadiness_MixedFrameworkUnevenCoverage(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 20},
			{Name: "mocha", Type: models.FrameworkTypeUnit, FileCount: 15},
			{Name: "cypress", Type: models.FrameworkTypeE2E, FileCount: 5},
		},
		TestFiles: []models.TestFile{
			{Path: "packages/auth/auth.test.js", Framework: "jest", TestCount: 15, AssertionCount: 30},
			{Path: "packages/auth/login.test.js", Framework: "jest", TestCount: 10, AssertionCount: 25},
			{Path: "packages/legacy-api/api.test.js", Framework: "mocha", TestCount: 8, AssertionCount: 3, MockCount: 12},
			{Path: "packages/legacy-api/routes.test.js", Framework: "mocha", TestCount: 5, AssertionCount: 1},
			{Path: "e2e/smoke.test.js", Framework: "cypress", TestCount: 3, AssertionCount: 5},
		},
		Signals: []models.Signal{
			{
				Type:     "frameworkMigration",
				Category: models.CategoryMigration,
				Location: models.SignalLocation{Repository: "test-repo"},
				Metadata: map[string]any{"frameworks": []string{"jest", "mocha"}, "frameworkCount": 2},
			},
			{
				Type:     "deprecatedTestPattern",
				Category: models.CategoryMigration,
				Location: models.SignalLocation{File: "packages/legacy-api/api.test.js"},
				Metadata: map[string]any{"blockerType": "deprecated-pattern"},
			},
			{
				Type:     "deprecatedTestPattern",
				Category: models.CategoryMigration,
				Location: models.SignalLocation{File: "packages/legacy-api/routes.test.js"},
				Metadata: map[string]any{"blockerType": "deprecated-pattern"},
			},
			{
				Type:     "weakAssertion",
				Category: models.CategoryQuality,
				Location: models.SignalLocation{File: "packages/legacy-api/api.test.js"},
			},
			{
				Type:     "weakAssertion",
				Category: models.CategoryQuality,
				Location: models.SignalLocation{File: "packages/legacy-api/routes.test.js"},
			},
			{
				Type:     "mockHeavyTest",
				Category: models.CategoryQuality,
				Location: models.SignalLocation{File: "packages/legacy-api/api.test.js"},
			},
		},
	}

	r := ComputeReadiness(snap)

	// 3 blockers (incl frameworkMigration) across 5 files = 60% → low readiness.
	if r.ReadinessLevel != "low" {
		t.Errorf("readiness = %q, want low for mixed repo", r.ReadinessLevel)
	}

	// Quality factors should show weak assertions + mock-heavy overlapping.
	if len(r.QualityFactors) < 2 {
		t.Errorf("qualityFactors = %d, want >= 2", len(r.QualityFactors))
	}

	// Area assessment: legacy-api risky, auth safe, e2e safe.
	areaMap := map[string]string{}
	for _, a := range r.AreaAssessments {
		areaMap[a.Directory] = a.Classification
	}
	if areaMap["packages/legacy-api"] != "risky" {
		t.Errorf("packages/legacy-api = %q, want risky", areaMap["packages/legacy-api"])
	}
	if areaMap["packages/auth"] != "safe" {
		t.Errorf("packages/auth = %q, want safe", areaMap["packages/auth"])
	}
	if areaMap["e2e"] != "safe" {
		t.Errorf("e2e = %q, want safe", areaMap["e2e"])
	}

	// Coverage guidance should prioritize legacy-api.
	if len(r.CoverageGuidance) == 0 {
		t.Fatal("expected coverage guidance")
	}
	if r.CoverageGuidance[0].Directory != "packages/legacy-api" {
		t.Errorf("top guidance directory = %q, want packages/legacy-api", r.CoverageGuidance[0].Directory)
	}
}
