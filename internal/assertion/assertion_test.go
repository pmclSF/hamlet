package assertion

import (
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestAssess_StrongTargeted(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:           "src/math_test.go",
				Framework:      "jest",
				TestCount:      10,
				AssertionCount: 35,
				MockCount:      2,
				SnapshotCount:  0,
			},
		},
	}

	result := Assess(snap)
	if len(result.Assessments) != 1 {
		t.Fatalf("expected 1 assessment, got %d", len(result.Assessments))
	}
	a := result.Assessments[0]
	if a.Strength != StrengthStrong {
		t.Errorf("strength = %s, want strong", a.Strength)
	}
	if a.Density != 3.5 {
		t.Errorf("density = %f, want 3.5", a.Density)
	}
	if a.Confidence < 0.7 {
		t.Errorf("confidence = %f, want >= 0.7", a.Confidence)
	}
	if a.DominantCategory != CategoryBehavioral {
		t.Errorf("dominant = %s, want behavioral", a.DominantCategory)
	}
	if result.OverallStrength != StrengthStrong {
		t.Errorf("overall = %s, want strong", result.OverallStrength)
	}
}

func TestAssess_SnapshotOnly(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:           "src/ui_test.js",
				Framework:      "jest",
				TestCount:      5,
				AssertionCount: 5,
				MockCount:      0,
				SnapshotCount:  5,
			},
		},
	}

	result := Assess(snap)
	a := result.Assessments[0]
	if a.Strength != StrengthWeak {
		t.Errorf("strength = %s, want weak (snapshot-dominated)", a.Strength)
	}
	if a.Categories[CategorySnapshot] != 5 {
		t.Errorf("snapshot count = %d, want 5", a.Categories[CategorySnapshot])
	}
}

func TestAssess_StatusCodeOnly(t *testing.T) {
	t.Parallel()
	// E2E framework with low density — should still be moderate.
	snap := &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{
			{Name: "cypress", Type: models.FrameworkTypeE2E},
		},
		TestFiles: []models.TestFile{
			{
				Path:           "cypress/e2e/api_test.js",
				Framework:      "cypress",
				TestCount:      10,
				AssertionCount: 8,
				MockCount:      0,
				SnapshotCount:  0,
			},
		},
	}

	result := Assess(snap)
	a := result.Assessments[0]
	// E2E with density 0.8 — should be moderate (implicit checks).
	if a.Strength != StrengthModerate {
		t.Errorf("strength = %s, want moderate for E2E with low density", a.Strength)
	}
}

func TestAssess_NoAssertions(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:           "test/empty_test.js",
				Framework:      "jest",
				TestCount:      3,
				AssertionCount: 0,
				MockCount:      0,
				SnapshotCount:  0,
			},
		},
	}

	result := Assess(snap)
	a := result.Assessments[0]
	if a.Strength != StrengthWeak {
		t.Errorf("strength = %s, want weak", a.Strength)
	}
	if a.Explanation != "no assertions detected" {
		t.Errorf("explanation = %q, want 'no assertions detected'", a.Explanation)
	}
}

func TestAssess_MockHeavy(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:           "test/service_test.js",
				Framework:      "jest",
				TestCount:      5,
				AssertionCount: 4,
				MockCount:      10,
				SnapshotCount:  0,
			},
		},
	}

	result := Assess(snap)
	a := result.Assessments[0]
	if a.Strength != StrengthWeak {
		t.Errorf("strength = %s, want weak (mock-heavy)", a.Strength)
	}
	if a.Confidence < 0.6 {
		t.Errorf("confidence = %f, want >= 0.6", a.Confidence)
	}
}

func TestAssess_MixedFiles(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:           "test/strong_test.js",
				Framework:      "jest",
				TestCount:      10,
				AssertionCount: 40,
				MockCount:      1,
				SnapshotCount:  0,
			},
			{
				Path:           "test/weak_test.js",
				Framework:      "jest",
				TestCount:      10,
				AssertionCount: 3,
				MockCount:      0,
				SnapshotCount:  0,
			},
			{
				Path:           "test/moderate_test.js",
				Framework:      "jest",
				TestCount:      10,
				AssertionCount: 20,
				MockCount:      2,
				SnapshotCount:  0,
			},
		},
	}

	result := Assess(snap)
	if len(result.Assessments) != 3 {
		t.Fatalf("expected 3 assessments, got %d", len(result.Assessments))
	}

	strengthMap := make(map[string]StrengthClass)
	for _, a := range result.Assessments {
		strengthMap[a.FilePath] = a.Strength
	}

	if strengthMap["test/strong_test.js"] != StrengthStrong {
		t.Errorf("strong file = %s, want strong", strengthMap["test/strong_test.js"])
	}
	if strengthMap["test/weak_test.js"] != StrengthWeak {
		t.Errorf("weak file = %s, want weak", strengthMap["test/weak_test.js"])
	}
	if strengthMap["test/moderate_test.js"] != StrengthModerate {
		t.Errorf("moderate file = %s, want moderate", strengthMap["test/moderate_test.js"])
	}

	// Overall: 1 strong, 1 moderate, 1 weak — should be moderate.
	if result.OverallStrength != StrengthModerate {
		t.Errorf("overall = %s, want moderate", result.OverallStrength)
	}
}

func TestAssess_EmptySnapshot(t *testing.T) {
	t.Parallel()
	result := Assess(nil)
	if result.OverallStrength != StrengthUnclear {
		t.Errorf("overall = %s, want unclear for nil snapshot", result.OverallStrength)
	}
	if len(result.Assessments) != 0 {
		t.Errorf("expected 0 assessments, got %d", len(result.Assessments))
	}

	result2 := Assess(&models.TestSuiteSnapshot{})
	if result2.OverallStrength != StrengthUnclear {
		t.Errorf("overall = %s, want unclear for empty snapshot", result2.OverallStrength)
	}
}

func TestAssess_E2EStrongDensity(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{
			{Name: "playwright", Type: models.FrameworkTypeE2E},
		},
		TestFiles: []models.TestFile{
			{
				Path:           "e2e/checkout_test.ts",
				Framework:      "playwright",
				TestCount:      5,
				AssertionCount: 12,
				MockCount:      0,
				SnapshotCount:  0,
			},
		},
	}

	result := Assess(snap)
	a := result.Assessments[0]
	// E2E with density 2.4 and no mocks → strong.
	if a.Strength != StrengthStrong {
		t.Errorf("strength = %s, want strong for E2E with density 2.4", a.Strength)
	}
}

func TestAssess_E2ELowDensity(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:           "e2e/smoke_test.ts",
				Framework:      "cypress",
				TestCount:      10,
				AssertionCount: 5,
				MockCount:      0,
				SnapshotCount:  0,
			},
		},
	}

	result := Assess(snap)
	a := result.Assessments[0]
	// E2E with density 0.5 — moderate (implicit checks).
	if a.Strength != StrengthModerate {
		t.Errorf("strength = %s, want moderate for E2E with low density", a.Strength)
	}
}

func TestAssess_NoTestsInFile(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:           "test/helper.js",
				Framework:      "jest",
				TestCount:      0,
				AssertionCount: 0,
			},
		},
	}

	result := Assess(snap)
	a := result.Assessments[0]
	if a.Strength != StrengthUnclear {
		t.Errorf("strength = %s, want unclear for file with no tests", a.Strength)
	}
}

func TestAssess_AverageDensity(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:           "test/a_test.js",
				Framework:      "jest",
				TestCount:      10,
				AssertionCount: 30,
			},
			{
				Path:           "test/b_test.js",
				Framework:      "jest",
				TestCount:      10,
				AssertionCount: 10,
			},
		},
	}

	result := Assess(snap)
	// Average density: (3.0 + 1.0) / 2 = 2.0
	if result.AverageDensity != 2.0 {
		t.Errorf("average density = %f, want 2.0", result.AverageDensity)
	}
}

func TestAssess_ByStrengthCounts(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "a.js", Framework: "jest", TestCount: 10, AssertionCount: 40},
			{Path: "b.js", Framework: "jest", TestCount: 10, AssertionCount: 40},
			{Path: "c.js", Framework: "jest", TestCount: 10, AssertionCount: 2},
		},
	}

	result := Assess(snap)
	if result.ByStrength[StrengthStrong] != 2 {
		t.Errorf("strong count = %d, want 2", result.ByStrength[StrengthStrong])
	}
	if result.ByStrength[StrengthWeak] != 1 {
		t.Errorf("weak count = %d, want 1", result.ByStrength[StrengthWeak])
	}
}

func TestAssess_HighMockRatioWithHighDensity(t *testing.T) {
	t.Parallel()
	// High density but mock ratio >= 0.5 should not be strong.
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:           "test/mocked_test.js",
				Framework:      "jest",
				TestCount:      5,
				AssertionCount: 20,
				MockCount:      25, // More mocks than assertions → weak.
			},
		},
	}

	result := Assess(snap)
	a := result.Assessments[0]
	if a.Strength != StrengthWeak {
		t.Errorf("strength = %s, want weak (mock count exceeds assertions)", a.Strength)
	}
}
