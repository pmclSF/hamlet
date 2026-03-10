package migration

import (
	"fmt"
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestComputeReadiness_BlockerRatios(t *testing.T) {
	t.Parallel()
	buildSignals := func(count int, signalType models.SignalType, blockerType string) []models.Signal {
		out := make([]models.Signal, 0, count)
		for i := 0; i < count; i++ {
			out = append(out, models.Signal{
				Type:     signalType,
				Metadata: map[string]any{"blockerType": blockerType},
			})
		}
		return out
	}

	tests := []struct {
		name             string
		testFileCount    int
		signals          []models.Signal
		wantReadiness    string
		wantTotalBlocker int
	}{
		{
			name:          "no blockers",
			testFileCount: 2,
			signals: []models.Signal{
				{Type: "weakAssertion"},
			},
			wantReadiness:    "high",
			wantTotalBlocker: 0,
		},
		{
			name:             "few blockers",
			testFileCount:    20,
			signals:          buildSignals(1, "deprecatedTestPattern", "deprecated-pattern"),
			wantReadiness:    "high", // 5%
			wantTotalBlocker: 1,
		},
		{
			name:             "medium blockers",
			testFileCount:    20,
			signals:          buildSignals(4, "deprecatedTestPattern", "deprecated-pattern"),
			wantReadiness:    "medium", // 20%
			wantTotalBlocker: 4,
		},
		{
			name:             "many blockers",
			testFileCount:    10,
			signals:          buildSignals(8, "migrationBlocker", "custom-matcher"),
			wantReadiness:    "low", // 80%
			wantTotalBlocker: 8,
		},
		{
			name:             "no test files",
			testFileCount:    0,
			signals:          nil,
			wantReadiness:    "unknown",
			wantTotalBlocker: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			snap := &models.TestSuiteSnapshot{
				TestFiles: make([]models.TestFile, tt.testFileCount),
				Signals:   tt.signals,
			}
			for i := range snap.TestFiles {
				snap.TestFiles[i].Path = fmt.Sprintf("test/%d.test.js", i)
			}

			r := ComputeReadiness(snap)
			if r.ReadinessLevel != tt.wantReadiness {
				t.Errorf("readiness = %q, want %q", r.ReadinessLevel, tt.wantReadiness)
			}
			if r.TotalBlockers != tt.wantTotalBlocker {
				t.Errorf("totalBlockers = %d, want %d", r.TotalBlockers, tt.wantTotalBlocker)
			}
		})
	}
}

func TestComputeReadiness_BlockersByType(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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

// --- Tier taxonomy tests ---

func TestTierForSignal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		signal models.Signal
		want   string
	}{
		{
			name: "enzyme usage is hard blocker by pattern override",
			signal: models.Signal{
				Type:     "deprecatedTestPattern",
				Metadata: map[string]any{"pattern": "enzyme-usage", "blockerType": BlockerDeprecatedPattern},
			},
			want: TierHardBlocker,
		},
		{
			name: "dynamic generation defaults to advisory by blocker type",
			signal: models.Signal{
				Type:     "dynamicTestGeneration",
				Metadata: map[string]any{"blockerType": BlockerDynamicGeneration},
			},
			want: TierAdvisory,
		},
		{
			name: "missing metadata falls back to soft blocker",
			signal: models.Signal{
				Type:     "migrationBlocker",
				Metadata: map[string]any{},
			},
			want: TierSoftBlocker,
		},
		{
			name: "setTimeout deprecation is advisory",
			signal: models.Signal{
				Type:     "deprecatedTestPattern",
				Metadata: map[string]any{"pattern": "setTimeout-in-test", "blockerType": BlockerDeprecatedPattern},
			},
			want: TierAdvisory,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tier := TierForSignal(tt.signal); tier != tt.want {
				t.Errorf("TierForSignal(%v) = %q, want %q", tt.signal.Metadata, tier, tt.want)
			}
		})
	}
}

func TestComputeReadiness_TierCounting(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 20),
		Signals: []models.Signal{
			// Hard blocker: enzyme
			{Type: "deprecatedTestPattern", Category: models.CategoryMigration, Metadata: map[string]any{"pattern": "enzyme-usage", "blockerType": BlockerDeprecatedPattern}},
			// Soft blocker: done-callback
			{Type: "deprecatedTestPattern", Category: models.CategoryMigration, Metadata: map[string]any{"pattern": "done-callback", "blockerType": BlockerDeprecatedPattern}},
			// Advisory: setTimeout
			{Type: "deprecatedTestPattern", Category: models.CategoryMigration, Metadata: map[string]any{"pattern": "setTimeout-in-test", "blockerType": BlockerDeprecatedPattern}},
			// Advisory: dynamic generation
			{Type: "dynamicTestGeneration", Category: models.CategoryMigration, Metadata: map[string]any{"blockerType": BlockerDynamicGeneration}},
		},
	}

	r := ComputeReadiness(snap)
	if r.HardBlockers != 1 {
		t.Errorf("HardBlockers = %d, want 1", r.HardBlockers)
	}
	if r.SoftBlockers != 1 {
		t.Errorf("SoftBlockers = %d, want 1", r.SoftBlockers)
	}
	if r.Advisories != 2 {
		t.Errorf("Advisories = %d, want 2", r.Advisories)
	}
	// TotalBlockers = hard + soft only.
	if r.TotalBlockers != 2 {
		t.Errorf("TotalBlockers = %d, want 2 (hard+soft only)", r.TotalBlockers)
	}
	if r.BlockersByTier[TierHardBlocker] != 1 {
		t.Errorf("BlockersByTier[hard-blocker] = %d, want 1", r.BlockersByTier[TierHardBlocker])
	}
	if r.BlockersByTier[TierAdvisory] != 2 {
		t.Errorf("BlockersByTier[advisory] = %d, want 2", r.BlockersByTier[TierAdvisory])
	}
}

func TestComputeReadiness_AdvisoryOnlyIsHigh(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 10),
		Signals: []models.Signal{
			{Type: "deprecatedTestPattern", Category: models.CategoryMigration, Metadata: map[string]any{"pattern": "setTimeout-in-test", "blockerType": BlockerDeprecatedPattern}},
			{Type: "dynamicTestGeneration", Category: models.CategoryMigration, Metadata: map[string]any{"blockerType": BlockerDynamicGeneration}},
		},
	}

	r := ComputeReadiness(snap)
	if r.ReadinessLevel != "high" {
		t.Errorf("readiness = %q, want high when only advisories present", r.ReadinessLevel)
	}
	if r.TotalBlockers != 0 {
		t.Errorf("TotalBlockers = %d, want 0 (advisories don't count)", r.TotalBlockers)
	}
}

func TestComputeReadiness_LowFrameworkConfidence(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{
			{Name: "unknown", Confidence: 0.3, FileCount: 10},
		},
		TestFiles: make([]models.TestFile, 10),
		Signals: []models.Signal{
			{Type: "deprecatedTestPattern", Category: models.CategoryMigration, Metadata: map[string]any{"blockerType": BlockerDeprecatedPattern}},
		},
	}

	r := ComputeReadiness(snap)
	if r.ReadinessLevel != "unknown" {
		t.Errorf("readiness = %q, want unknown when framework confidence < 0.5", r.ReadinessLevel)
	}
}

func TestComputeReadiness_UsesWeightedAverageFrameworkConfidence(t *testing.T) {
	t.Parallel()
	// With weighted average: (1.0*50 + 0.3*1) / 51 ≈ 0.99 — high confidence
	// because the low-confidence framework has very few files.
	snap := &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{
			{Name: "jest", Confidence: 1.0, FileCount: 50},
			{Name: "vitest", Confidence: 0.3, FileCount: 1},
		},
		TestFiles: make([]models.TestFile, 8),
	}

	r := ComputeReadiness(snap)
	if r.ReadinessLevel == "unknown" {
		t.Errorf("readiness = %q, want non-unknown when weighted confidence is high", r.ReadinessLevel)
	}
}

func TestComputeReadiness_RepresentativeBlockersPrioritizeHard(t *testing.T) {
	t.Parallel()
	signals := []models.Signal{}
	// Add 3 soft blockers first, then 2 hard blockers.
	for i := 0; i < 3; i++ {
		signals = append(signals, models.Signal{
			Type:     "deprecatedTestPattern",
			Category: models.CategoryMigration,
			Location: models.SignalLocation{File: fmt.Sprintf("src/soft%d.test.js", i)},
			Metadata: map[string]any{"pattern": "done-callback", "blockerType": BlockerDeprecatedPattern},
		})
	}
	for i := 0; i < 2; i++ {
		signals = append(signals, models.Signal{
			Type:     "deprecatedTestPattern",
			Category: models.CategoryMigration,
			Location: models.SignalLocation{File: fmt.Sprintf("src/hard%d.test.js", i)},
			Metadata: map[string]any{"pattern": "enzyme-usage", "blockerType": BlockerDeprecatedPattern},
		})
	}

	snap := &models.TestSuiteSnapshot{
		TestFiles: make([]models.TestFile, 20),
		Signals:   signals,
	}

	r := ComputeReadiness(snap)
	if len(r.RepresentativeBlockers) < 2 {
		t.Fatalf("representative blockers count = %d, want >= 2", len(r.RepresentativeBlockers))
	}
	// Hard blockers should appear first.
	if r.RepresentativeBlockers[0].File != "src/hard0.test.js" && r.RepresentativeBlockers[0].File != "src/hard1.test.js" {
		t.Errorf("first representative blocker file = %q, want a hard blocker file", r.RepresentativeBlockers[0].File)
	}
}

// --- Well-covered migration candidate (golden scenario) ---

func TestComputeReadiness_WellCoveredMigrationCandidate(t *testing.T) {
	t.Parallel()
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

	// frameworkMigration is advisory, so a well-covered repo remains high readiness.
	if r.ReadinessLevel != "high" {
		t.Errorf("readiness = %q, want high for well-covered repo with framework fragmentation", r.ReadinessLevel)
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
	t.Parallel()
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
	t.Parallel()
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
