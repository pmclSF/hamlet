package envdepth

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestAssess_HeavyMocking(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:           "src/service.test.js",
				Framework:      "jest",
				TestCount:      5,
				AssertionCount: 3,
				MockCount:      10,
			},
		},
	}

	result := Assess(snap)

	if len(result.Assessments) != 1 {
		t.Fatalf("expected 1 assessment, got %d", len(result.Assessments))
	}
	a := result.Assessments[0]
	if a.Depth != DepthHeavyMocking {
		t.Errorf("depth = %s, want %s", a.Depth, DepthHeavyMocking)
	}
	if a.Confidence < 0.7 {
		t.Errorf("confidence = %f, want >= 0.7", a.Confidence)
	}
	if a.MockRatio < 0.5 {
		t.Errorf("mock ratio = %f, want >= 0.5", a.MockRatio)
	}
	if result.ByDepth[DepthHeavyMocking] != 1 {
		t.Errorf("ByDepth[heavy_mocking] = %d, want 1", result.ByDepth[DepthHeavyMocking])
	}
}

func TestAssess_HeavyMocking_HighAbsoluteCount(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:           "src/controller.test.js",
				Framework:      "jest",
				TestCount:      10,
				AssertionCount: 12,
				MockCount:      8,
			},
		},
	}

	result := Assess(snap)
	a := result.Assessments[0]
	if a.Depth != DepthHeavyMocking {
		t.Errorf("depth = %s, want %s (MockCount >= 8 triggers heavy mocking)", a.Depth, DepthHeavyMocking)
	}
}

func TestAssess_ModerateMocking(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:           "src/utils.test.js",
				Framework:      "jest",
				TestCount:      8,
				AssertionCount: 10,
				MockCount:      3,
			},
		},
	}

	result := Assess(snap)
	a := result.Assessments[0]
	if a.Depth != DepthModerateMocking {
		t.Errorf("depth = %s, want %s", a.Depth, DepthModerateMocking)
	}
	if a.Confidence < 0.6 {
		t.Errorf("confidence = %f, want >= 0.6", a.Confidence)
	}
}

func TestAssess_BrowserRuntime(t *testing.T) {
	t.Parallel()
	frameworks := []string{"cypress", "playwright", "puppeteer", "testcafe", "webdriverio", "selenium"}

	for _, fw := range frameworks {
		t.Run(fw, func(t *testing.T) {
			t.Parallel()
			snap := &models.TestSuiteSnapshot{
				TestFiles: []models.TestFile{
					{
						Path:           "e2e/login.test.js",
						Framework:      fw,
						TestCount:      3,
						AssertionCount: 5,
						MockCount:      0,
					},
				},
			}

			result := Assess(snap)
			a := result.Assessments[0]
			if a.Depth != DepthBrowserRuntime {
				t.Errorf("depth = %s, want %s for framework %s", a.Depth, DepthBrowserRuntime, fw)
			}
			if a.Confidence < 0.8 {
				t.Errorf("confidence = %f, want >= 0.8 for framework %s", a.Confidence, fw)
			}
			// Should have browser driver indicator.
			found := false
			for _, ind := range a.Indicators {
				if ind == IndicatorBrowserDriver {
					found = true
				}
			}
			if !found {
				t.Errorf("expected IndicatorBrowserDriver for framework %s", fw)
			}
		})
	}
}

func TestAssess_RealDependency(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{
			{Name: "supertest", Type: models.FrameworkTypeIntegration},
		},
		TestFiles: []models.TestFile{
			{
				Path:           "test/integration/api.test.js",
				Framework:      "supertest",
				TestCount:      5,
				AssertionCount: 8,
				MockCount:      0,
			},
		},
	}

	result := Assess(snap)
	a := result.Assessments[0]
	if a.Depth != DepthRealDependency {
		t.Errorf("depth = %s, want %s", a.Depth, DepthRealDependency)
	}
}

func TestAssess_RealDependency_E2EFrameworkType(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{
			{Name: "customfw", Type: models.FrameworkTypeE2E},
		},
		TestFiles: []models.TestFile{
			{
				Path:           "test/e2e/flow.test.js",
				Framework:      "customfw",
				TestCount:      3,
				AssertionCount: 4,
				MockCount:      0,
			},
		},
	}

	result := Assess(snap)
	a := result.Assessments[0]
	if a.Depth != DepthRealDependency {
		t.Errorf("depth = %s, want %s", a.Depth, DepthRealDependency)
	}
}

func TestAssess_Unknown(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:      "test/something.test.js",
				TestCount: 1,
			},
		},
	}

	result := Assess(snap)
	a := result.Assessments[0]
	if a.Depth != DepthUnknown {
		t.Errorf("depth = %s, want %s", a.Depth, DepthUnknown)
	}
	if a.Confidence > 0.5 {
		t.Errorf("confidence = %f, want <= 0.5 for unknown depth", a.Confidence)
	}
}

func TestAssess_MixedFiles(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:           "src/unit.test.js",
				Framework:      "jest",
				TestCount:      5,
				AssertionCount: 2,
				MockCount:      10,
			},
			{
				Path:           "e2e/login.test.js",
				Framework:      "cypress",
				TestCount:      3,
				AssertionCount: 5,
				MockCount:      0,
			},
			{
				Path:           "src/helper.test.js",
				Framework:      "jest",
				TestCount:      4,
				AssertionCount: 6,
				MockCount:      2,
			},
		},
	}

	result := Assess(snap)
	if len(result.Assessments) != 3 {
		t.Fatalf("expected 3 assessments, got %d", len(result.Assessments))
	}

	depths := make(map[DepthClass]int)
	for _, a := range result.Assessments {
		depths[a.Depth]++
	}

	if depths[DepthHeavyMocking] != 1 {
		t.Errorf("expected 1 heavy_mocking, got %d", depths[DepthHeavyMocking])
	}
	if depths[DepthBrowserRuntime] != 1 {
		t.Errorf("expected 1 browser_runtime, got %d", depths[DepthBrowserRuntime])
	}
	if depths[DepthModerateMocking] != 1 {
		t.Errorf("expected 1 moderate_mocking, got %d", depths[DepthModerateMocking])
	}
}

func TestAssess_EmptySnapshot(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{}

	result := Assess(snap)
	if len(result.Assessments) != 0 {
		t.Errorf("expected 0 assessments, got %d", len(result.Assessments))
	}
	if result.OverallDepth != DepthUnknown {
		t.Errorf("overall depth = %s, want %s", result.OverallDepth, DepthUnknown)
	}
}

func TestAssess_NilSnapshot(t *testing.T) {
	t.Parallel()
	result := Assess(nil)
	if result == nil {
		t.Fatal("expected non-nil result for nil snapshot")
	}
	if result.OverallDepth != DepthUnknown {
		t.Errorf("overall depth = %s, want %s", result.OverallDepth, DepthUnknown)
	}
	if len(result.Assessments) != 0 {
		t.Errorf("expected 0 assessments, got %d", len(result.Assessments))
	}
}

func TestAssess_OverallDepth_MajorityWins(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "a.test.js", Framework: "cypress", TestCount: 1, AssertionCount: 1},
			{Path: "b.test.js", Framework: "cypress", TestCount: 1, AssertionCount: 1},
			{Path: "c.test.js", Framework: "jest", TestCount: 1, AssertionCount: 2, MockCount: 10},
		},
	}

	result := Assess(snap)
	if result.OverallDepth != DepthBrowserRuntime {
		t.Errorf("overall depth = %s, want %s (browser_runtime is majority)", result.OverallDepth, DepthBrowserRuntime)
	}
}

func TestAssess_MockRatioComputation(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:           "test.js",
				Framework:      "jest",
				TestCount:      5,
				AssertionCount: 6,
				MockCount:      4,
			},
		},
	}

	result := Assess(snap)
	a := result.Assessments[0]
	// MockRatio = 4 / (4 + 6) = 0.4
	expected := 0.4
	if a.MockRatio < expected-0.01 || a.MockRatio > expected+0.01 {
		t.Errorf("mock ratio = %f, want ~%f", a.MockRatio, expected)
	}
}

func TestAssess_BrowserWithMocksStillBrowser(t *testing.T) {
	t.Parallel()
	// Browser framework takes precedence even if mocks are present.
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:           "e2e/test.js",
				Framework:      "playwright",
				TestCount:      3,
				AssertionCount: 2,
				MockCount:      5,
			},
		},
	}

	result := Assess(snap)
	a := result.Assessments[0]
	if a.Depth != DepthBrowserRuntime {
		t.Errorf("depth = %s, want %s (browser framework takes precedence)", a.Depth, DepthBrowserRuntime)
	}
	// Should have both indicators.
	hasBrowser, hasMock := false, false
	for _, ind := range a.Indicators {
		if ind == IndicatorBrowserDriver {
			hasBrowser = true
		}
		if ind == IndicatorMockLibrary {
			hasMock = true
		}
	}
	if !hasBrowser {
		t.Error("expected IndicatorBrowserDriver")
	}
	if !hasMock {
		t.Error("expected IndicatorMockLibrary")
	}
}
