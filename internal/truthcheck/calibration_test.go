package truthcheck

import (
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestCalibrateFromFixtures_BasicPrecision(t *testing.T) {
	t.Parallel()

	fixtures := []CalibrationInput{
		{
			Name: "test-repo",
			Snapshot: &models.TestSuiteSnapshot{
				CodeSurfaces: []models.CodeSurface{
					{SurfaceID: "surface:src/auth.ts:login", Name: "login", Path: "src/auth.ts", Kind: models.SurfaceFunction, DetectionTier: models.TierPattern, Confidence: 0.90},
					{SurfaceID: "surface:src/auth.ts:register", Name: "register", Path: "src/auth.ts", Kind: models.SurfaceFunction, DetectionTier: models.TierPattern, Confidence: 0.90},
					{SurfaceID: "surface:src/prompts.ts:buildPrompt", Name: "buildPrompt", Path: "src/prompts.ts", Kind: models.SurfacePrompt, DetectionTier: models.TierPattern, Confidence: 0.85},
				},
			},
			Truth: &TruthSpec{
				Coverage: &CoverageTruth{
					ExpectedUncovered: []CoverageItem{
						{Path: "src/auth.ts", Reason: "no tests"},
					},
				},
				AI: &AITruth{
					ExpectedPromptSurfaces: []string{"src/prompts.ts:buildPrompt"},
				},
			},
		},
	}

	result := CalibrateFromFixtures(fixtures)

	if result.FixtureCount != 1 {
		t.Errorf("expected 1 fixture, got %d", result.FixtureCount)
	}
	if result.TotalSurfaces != 3 {
		t.Errorf("expected 3 total surfaces, got %d", result.TotalSurfaces)
	}

	// Function surfaces: 2 detected, login + register both match src/auth.ts.
	funcKind := result.ByKind[models.SurfaceFunction]
	if funcKind == nil {
		t.Fatal("expected function kind calibration")
	}
	if funcKind.Detected != 2 {
		t.Errorf("function detected: want 2, got %d", funcKind.Detected)
	}

	// Prompt surfaces.
	promptKind := result.ByKind[models.SurfacePrompt]
	if promptKind == nil {
		t.Fatal("expected prompt kind calibration")
	}
	if promptKind.Detected != 1 {
		t.Errorf("prompt detected: want 1, got %d", promptKind.Detected)
	}
	if promptKind.Correct != 1 {
		t.Errorf("prompt correct: want 1, got %d", promptKind.Correct)
	}
}

func TestCalibrateFromFixtures_TierAggregation(t *testing.T) {
	t.Parallel()

	fixtures := []CalibrationInput{
		{
			Name: "go-repo",
			Snapshot: &models.TestSuiteSnapshot{
				CodeSurfaces: []models.CodeSurface{
					{SurfaceID: "surface:pkg/svc.go:Hello", Name: "Hello", Path: "pkg/svc.go", Kind: models.SurfaceFunction, Language: "go", DetectionTier: models.TierStructural, Confidence: 0.99},
					{SurfaceID: "surface:pkg/svc.go:World", Name: "World", Path: "pkg/svc.go", Kind: models.SurfaceFunction, Language: "go", DetectionTier: models.TierStructural, Confidence: 0.99},
				},
			},
			Truth: &TruthSpec{
				Coverage: &CoverageTruth{
					ExpectedUncovered: []CoverageItem{
						{Path: "pkg/svc.go", Reason: "no tests"},
					},
				},
			},
		},
	}

	result := CalibrateFromFixtures(fixtures)

	structTier := result.ByTier[models.TierStructural]
	if structTier == nil {
		t.Fatal("expected structural tier")
	}
	if structTier.Detected != 2 {
		t.Errorf("structural detected: want 2, got %d", structTier.Detected)
	}
}

func TestCalibrateFromFixtures_ConfidenceBands(t *testing.T) {
	t.Parallel()

	fixtures := []CalibrationInput{
		{
			Name: "mixed",
			Snapshot: &models.TestSuiteSnapshot{
				CodeSurfaces: []models.CodeSurface{
					{SurfaceID: "surface:a.go:A", Name: "A", Path: "a.go", Kind: models.SurfaceFunction, DetectionTier: models.TierStructural, Confidence: 0.99},
					{SurfaceID: "surface:b.ts:b", Name: "b", Path: "b.ts", Kind: models.SurfaceFunction, DetectionTier: models.TierPattern, Confidence: 0.90},
					{SurfaceID: "surface:c.ts:c", Name: "c", Path: "c.ts", Kind: models.SurfaceRetrieval, DetectionTier: models.TierPattern, Confidence: 0.80},
					{SurfaceID: "surface:d.ts:d", Name: "d", Path: "d.ts", Kind: models.SurfacePrompt, DetectionTier: models.TierContent, Confidence: 0.75},
				},
			},
			Truth: &TruthSpec{
				Coverage: &CoverageTruth{
					ExpectedUncovered: []CoverageItem{{Path: "a.go"}},
				},
			},
		},
	}

	result := CalibrateFromFixtures(fixtures)

	if len(result.ByConfidenceBand) == 0 {
		t.Fatal("expected confidence bands")
	}

	// The 0.95-1.00 band should have 1 surface (the 0.99 Go function).
	var highBand *ConfidenceBand
	for i := range result.ByConfidenceBand {
		if result.ByConfidenceBand[i].RangeLabel == "0.95-1.00" {
			highBand = &result.ByConfidenceBand[i]
		}
	}
	if highBand == nil {
		t.Fatal("expected 0.95-1.00 band")
	}
	if highBand.Count != 1 {
		t.Errorf("0.95-1.00 band count: want 1, got %d", highBand.Count)
	}
}

func TestCalibrateFromFixtures_BasisAssignment(t *testing.T) {
	t.Parallel()

	// Create enough surfaces to trigger calibrated basis (>=5).
	var surfaces []models.CodeSurface
	for i := 0; i < 6; i++ {
		surfaces = append(surfaces, models.CodeSurface{
			SurfaceID:     "surface:pkg.go:F" + string(rune('A'+i)),
			Name:          "F" + string(rune('A'+i)),
			Path:          "pkg.go",
			Kind:          models.SurfaceFunction,
			Language:      "go",
			DetectionTier: models.TierStructural,
			Confidence:    0.99,
		})
	}

	fixtures := []CalibrationInput{
		{
			Name:     "many-funcs",
			Snapshot: &models.TestSuiteSnapshot{CodeSurfaces: surfaces},
			Truth:    &TruthSpec{Coverage: &CoverageTruth{ExpectedUncovered: []CoverageItem{{Path: "pkg.go"}}}},
		},
	}

	result := CalibrateFromFixtures(fixtures)
	funcKind := result.ByKind[models.SurfaceFunction]
	if funcKind == nil {
		t.Fatal("expected function kind")
	}
	if funcKind.Basis != models.ConfidenceBasisCalibrated {
		t.Errorf("basis: want calibrated (>=5 detections), got %s", funcKind.Basis)
	}
}

func TestCalibrateFromFixtures_FewDetectionsAreHeuristic(t *testing.T) {
	t.Parallel()

	fixtures := []CalibrationInput{
		{
			Name: "small",
			Snapshot: &models.TestSuiteSnapshot{
				CodeSurfaces: []models.CodeSurface{
					{SurfaceID: "surface:a.ts:p", Name: "p", Path: "a.ts", Kind: models.SurfacePrompt, DetectionTier: models.TierPattern, Confidence: 0.85},
				},
			},
			Truth: &TruthSpec{AI: &AITruth{ExpectedPromptSurfaces: []string{"a.ts:p"}}},
		},
	}

	result := CalibrateFromFixtures(fixtures)
	promptKind := result.ByKind[models.SurfacePrompt]
	if promptKind == nil {
		t.Fatal("expected prompt kind")
	}
	if promptKind.Basis != models.ConfidenceBasisHeuristic {
		t.Errorf("basis: want heuristic (<5 detections), got %s", promptKind.Basis)
	}
}

func TestCalibrateFromFixtures_EmptyInput(t *testing.T) {
	t.Parallel()
	result := CalibrateFromFixtures(nil)
	if result.FixtureCount != 0 {
		t.Errorf("expected 0 fixtures, got %d", result.FixtureCount)
	}
	if result.TotalSurfaces != 0 {
		t.Errorf("expected 0 surfaces, got %d", result.TotalSurfaces)
	}
}

func TestFormatCalibrationReport(t *testing.T) {
	t.Parallel()

	result := &CalibrationResult{
		FixtureCount:  2,
		TotalSurfaces: 10,
		ByKind: map[models.CodeSurfaceKind]*KindCalibration{
			models.SurfaceFunction: {Kind: "function", Detected: 8, Correct: 7, Precision: 0.875, Recall: 1.0, F1: 0.933, Basis: models.ConfidenceBasisCalibrated},
			models.SurfacePrompt:   {Kind: "prompt", Detected: 2, Correct: 2, Precision: 1.0, Recall: 1.0, F1: 1.0, Basis: models.ConfidenceBasisHeuristic},
		},
		ByTier: map[string]*TierCalibration{
			models.TierStructural: {Tier: "structural", Detected: 5, Correct: 5, Precision: 1.0},
			models.TierPattern:    {Tier: "pattern", Detected: 5, Correct: 4, Precision: 0.8},
		},
		ByConfidenceBand: []ConfidenceBand{
			{RangeLabel: "0.95-1.00", Count: 5, CorrectCount: 5, ObservedPrecision: 1.0},
			{RangeLabel: "0.85-0.90", Count: 3, CorrectCount: 2, ObservedPrecision: 0.667},
		},
	}

	report := FormatCalibrationReport(result)

	if !strings.Contains(report, "Confidence Calibration Report") {
		t.Error("missing header")
	}
	if !strings.Contains(report, "structural") {
		t.Error("missing structural tier")
	}
	if !strings.Contains(report, "function") {
		t.Error("missing function kind")
	}
	if !strings.Contains(report, "calibrated") {
		t.Error("missing calibrated basis")
	}
	if !strings.Contains(report, "0.95-1.00") {
		t.Error("missing confidence band")
	}
}
