package governance

import (
	"testing"

	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/policy"
)

func boolPtr(v bool) *bool          { return &v }
func float64Ptr(v float64) *float64 { return &v }
func intPtr(v int) *int             { return &v }

func TestEvaluate_NoPolicy(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{}
	result := Evaluate(snap, nil)
	if !result.Pass {
		t.Error("expected PASS with nil policy")
	}
	if len(result.Violations) != 0 {
		t.Errorf("expected 0 violations, got %d", len(result.Violations))
	}
}

func TestEvaluate_EmptyPolicy(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{}
	cfg := &policy.Config{}
	result := Evaluate(snap, cfg)
	if !result.Pass {
		t.Error("expected PASS with empty policy")
	}
}

func TestEvaluate_DisallowedFramework(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{Name: "test-repo"},
		Frameworks: []models.Framework{
			{Name: "jest", FileCount: 42},
			{Name: "vitest", FileCount: 10},
		},
	}
	cfg := &policy.Config{
		Rules: policy.Rules{
			DisallowFrameworks: []string{"jest"},
		},
	}

	result := Evaluate(snap, cfg)
	if result.Pass {
		t.Error("expected FAIL when disallowed framework is present")
	}
	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(result.Violations))
	}
	v := result.Violations[0]
	if v.Type != "legacyFrameworkUsage" {
		t.Errorf("type = %q, want legacyFrameworkUsage", v.Type)
	}
	if v.Category != models.CategoryGovernance {
		t.Errorf("category = %q, want governance", v.Category)
	}
}

func TestEvaluate_DisallowedFramework_CaseInsensitive(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{
			{Name: "Jest", FileCount: 5},
		},
	}
	cfg := &policy.Config{
		Rules: policy.Rules{
			DisallowFrameworks: []string{"jest"},
		},
	}

	result := Evaluate(snap, cfg)
	if result.Pass {
		t.Error("expected FAIL (case-insensitive match)")
	}
}

func TestEvaluate_SkippedTests(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{Name: "test-repo"},
		Signals: []models.Signal{
			{Type: "skippedTest", Category: models.CategoryHealth},
			{Type: "skippedTest", Category: models.CategoryHealth},
		},
	}
	cfg := &policy.Config{
		Rules: policy.Rules{
			DisallowSkippedTests: boolPtr(true),
		},
	}

	result := Evaluate(snap, cfg)
	if result.Pass {
		t.Error("expected FAIL when skipped tests exist")
	}
	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(result.Violations))
	}
	if result.Violations[0].Type != "skippedTestsInCI" {
		t.Errorf("type = %q, want skippedTestsInCI", result.Violations[0].Type)
	}
}

func TestEvaluate_SkippedTests_NonePresent(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{}
	cfg := &policy.Config{
		Rules: policy.Rules{
			DisallowSkippedTests: boolPtr(true),
		},
	}

	result := Evaluate(snap, cfg)
	if !result.Pass {
		t.Error("expected PASS when no skipped tests present")
	}
}

func TestEvaluate_SkippedTests_CountsBeyondTopFiveFiles(t *testing.T) {
	t.Parallel()
	signals := make([]models.Signal, 0, 6)
	for i := 0; i < 6; i++ {
		signals = append(signals, models.Signal{
			Type: "skippedTest",
			Location: models.SignalLocation{
				File: "test/skip" + string(rune('A'+i)) + ".test.js",
			},
		})
	}

	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{Name: "test-repo"},
		Signals:    signals,
	}
	cfg := &policy.Config{
		Rules: policy.Rules{
			DisallowSkippedTests: boolPtr(true),
		},
	}

	result := Evaluate(snap, cfg)
	if result.Pass {
		t.Fatal("expected FAIL when skipped tests exist across more than five files")
	}
	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(result.Violations))
	}
	gotCount, ok := result.Violations[0].Metadata["skippedCount"].(int)
	if !ok {
		t.Fatalf("expected skippedCount metadata as int, got %T", result.Violations[0].Metadata["skippedCount"])
	}
	if gotCount != 6 {
		t.Fatalf("skippedCount metadata = %d, want 6", gotCount)
	}
}

func TestEvaluate_SkippedTests_CountsMixedRepoAndFileLevel(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{Name: "test-repo"},
		Signals: []models.Signal{
			{Type: "skippedTest", Location: models.SignalLocation{File: "test/a.test.js"}},
			{Type: "skippedTest", Location: models.SignalLocation{Repository: "test-repo"}},
		},
	}
	cfg := &policy.Config{
		Rules: policy.Rules{
			DisallowSkippedTests: boolPtr(true),
		},
	}

	result := Evaluate(snap, cfg)
	if result.Pass {
		t.Fatal("expected FAIL when skipped tests exist")
	}
	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(result.Violations))
	}
	gotCount, ok := result.Violations[0].Metadata["skippedCount"].(int)
	if !ok {
		t.Fatalf("expected skippedCount metadata as int, got %T", result.Violations[0].Metadata["skippedCount"])
	}
	if gotCount != 2 {
		t.Fatalf("skippedCount metadata = %d, want 2", gotCount)
	}
}

func TestEvaluate_RuntimeBudget(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/fast.test.js", RuntimeStats: &models.RuntimeStats{AvgRuntimeMs: 1000}},
			{Path: "test/slow.test.js", RuntimeStats: &models.RuntimeStats{AvgRuntimeMs: 8000}},
		},
	}
	cfg := &policy.Config{
		Rules: policy.Rules{
			MaxTestRuntimeMs: float64Ptr(5000),
		},
	}

	result := Evaluate(snap, cfg)
	if result.Pass {
		t.Error("expected FAIL when runtime exceeds budget")
	}
	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(result.Violations))
	}
	if result.Violations[0].Type != "runtimeBudgetExceeded" {
		t.Errorf("type = %q, want runtimeBudgetExceeded", result.Violations[0].Type)
	}
}

func TestEvaluate_RuntimeBudget_AllUnderBudget(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/fast.test.js", RuntimeStats: &models.RuntimeStats{AvgRuntimeMs: 1000}},
		},
	}
	cfg := &policy.Config{
		Rules: policy.Rules{
			MaxTestRuntimeMs: float64Ptr(5000),
		},
	}

	result := Evaluate(snap, cfg)
	if !result.Pass {
		t.Error("expected PASS when all runtimes are under budget")
	}
}

func TestEvaluate_CoverageThreshold(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{Name: "test-repo"},
		Signals: []models.Signal{
			{Type: "coverageThresholdBreak", Category: models.CategoryQuality},
		},
	}
	cfg := &policy.Config{
		Rules: policy.Rules{
			MinimumCoveragePercent: float64Ptr(80),
		},
	}

	result := Evaluate(snap, cfg)
	if result.Pass {
		t.Error("expected FAIL when coverage breaks exist")
	}
}

func TestEvaluate_WeakAssertionThreshold(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{Name: "test-repo"},
		Signals: []models.Signal{
			{Type: "weakAssertion"},
			{Type: "weakAssertion"},
			{Type: "weakAssertion"},
		},
	}
	cfg := &policy.Config{
		Rules: policy.Rules{
			MaxWeakAssertions: intPtr(2),
		},
	}

	result := Evaluate(snap, cfg)
	if result.Pass {
		t.Error("expected FAIL when weakAssertion count exceeds max")
	}
}

func TestEvaluate_WeakAssertionThreshold_CountsBeyondTopFiveFiles(t *testing.T) {
	t.Parallel()
	signals := make([]models.Signal, 0, 6)
	for i := 0; i < 6; i++ {
		signals = append(signals, models.Signal{
			Type: "weakAssertion",
			Location: models.SignalLocation{
				File: "test/file" + string(rune('A'+i)) + ".test.js",
			},
		})
	}

	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{Name: "test-repo"},
		Signals:    signals,
	}
	cfg := &policy.Config{
		Rules: policy.Rules{
			MaxWeakAssertions: intPtr(5),
		},
	}

	result := Evaluate(snap, cfg)
	if result.Pass {
		t.Fatal("expected FAIL when total weakAssertion count exceeds max across more than five files")
	}
	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(result.Violations))
	}
	gotCount, ok := result.Violations[0].Metadata["count"].(int)
	if !ok {
		t.Fatalf("expected count metadata as int, got %T", result.Violations[0].Metadata["count"])
	}
	if gotCount != 6 {
		t.Fatalf("count metadata = %d, want 6", gotCount)
	}
}

func TestEvaluate_WeakAssertionThreshold_UnderLimit(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Signals: []models.Signal{
			{Type: "weakAssertion"},
		},
	}
	cfg := &policy.Config{
		Rules: policy.Rules{
			MaxWeakAssertions: intPtr(5),
		},
	}

	result := Evaluate(snap, cfg)
	if !result.Pass {
		t.Error("expected PASS when weakAssertion count is under limit")
	}
}

func TestEvaluate_MockHeavyThreshold(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{Name: "test-repo"},
		Signals: []models.Signal{
			{Type: "mockHeavyTest"},
			{Type: "mockHeavyTest"},
		},
	}
	cfg := &policy.Config{
		Rules: policy.Rules{
			MaxMockHeavyTests: intPtr(1),
		},
	}

	result := Evaluate(snap, cfg)
	if result.Pass {
		t.Error("expected FAIL when mockHeavyTest count exceeds max")
	}
}

func TestEvaluate_MockHeavyThreshold_CountsBeyondTopFiveFiles(t *testing.T) {
	t.Parallel()
	signals := make([]models.Signal, 0, 6)
	for i := 0; i < 6; i++ {
		signals = append(signals, models.Signal{
			Type: "mockHeavyTest",
			Location: models.SignalLocation{
				File: "test/mock" + string(rune('A'+i)) + ".test.js",
			},
		})
	}

	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{Name: "test-repo"},
		Signals:    signals,
	}
	cfg := &policy.Config{
		Rules: policy.Rules{
			MaxMockHeavyTests: intPtr(5),
		},
	}

	result := Evaluate(snap, cfg)
	if result.Pass {
		t.Fatal("expected FAIL when total mockHeavyTest count exceeds max across more than five files")
	}
	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(result.Violations))
	}
	gotCount, ok := result.Violations[0].Metadata["count"].(int)
	if !ok {
		t.Fatalf("expected count metadata as int, got %T", result.Violations[0].Metadata["count"])
	}
	if gotCount != 6 {
		t.Fatalf("count metadata = %d, want 6", gotCount)
	}
}

func TestEvaluate_WeakAssertionThreshold_SizeAdjustedLargeRepo(t *testing.T) {
	t.Parallel()
	testFiles := make([]models.TestFile, 1000)
	signals := make([]models.Signal, 0, 50)
	for i := 0; i < 50; i++ {
		signals = append(signals, models.Signal{Type: "weakAssertion"})
	}

	snap := &models.TestSuiteSnapshot{
		TestFiles: testFiles,
		Signals:   signals,
	}
	cfg := &policy.Config{
		Rules: policy.Rules{
			MaxWeakAssertions: intPtr(10), // scales to 100 for 1000 files
		},
	}

	result := Evaluate(snap, cfg)
	if !result.Pass {
		t.Fatal("expected PASS when weakAssertion count is under size-adjusted threshold")
	}
}

func TestEvaluate_WeakAssertionThreshold_SizeAdjustedViolation(t *testing.T) {
	t.Parallel()
	testFiles := make([]models.TestFile, 1000)
	signals := make([]models.Signal, 0, 120)
	for i := 0; i < 120; i++ {
		signals = append(signals, models.Signal{Type: "weakAssertion"})
	}

	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{Name: "test-repo"},
		TestFiles:  testFiles,
		Signals:    signals,
	}
	cfg := &policy.Config{
		Rules: policy.Rules{
			MaxWeakAssertions: intPtr(10),
		},
	}

	result := Evaluate(snap, cfg)
	if result.Pass {
		t.Fatal("expected FAIL when weakAssertion count exceeds size-adjusted threshold")
	}
	if len(result.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(result.Violations))
	}
	if gotMax, ok := result.Violations[0].Metadata["max"].(int); !ok || gotMax != 100 {
		t.Fatalf("expected metadata max=100, got %#v", result.Violations[0].Metadata["max"])
	}
}

func TestEvaluate_MockHeavyThreshold_SizeAdjustedLargeRepo(t *testing.T) {
	t.Parallel()
	testFiles := make([]models.TestFile, 800)
	signals := make([]models.Signal, 0, 60)
	for i := 0; i < 60; i++ {
		signals = append(signals, models.Signal{Type: "mockHeavyTest"})
	}

	snap := &models.TestSuiteSnapshot{
		TestFiles: testFiles,
		Signals:   signals,
	}
	cfg := &policy.Config{
		Rules: policy.Rules{
			MaxMockHeavyTests: intPtr(10), // scales to 80 for 800 files
		},
	}

	result := Evaluate(snap, cfg)
	if !result.Pass {
		t.Fatal("expected PASS when mockHeavy count is under size-adjusted threshold")
	}
}

func TestEvaluate_NoViolations(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{
			{Name: "vitest", FileCount: 10},
		},
		TestFiles: []models.TestFile{
			{Path: "test/a.test.js", RuntimeStats: &models.RuntimeStats{AvgRuntimeMs: 200}},
		},
	}
	cfg := &policy.Config{
		Rules: policy.Rules{
			DisallowFrameworks: []string{"jest"},
			MaxTestRuntimeMs:   float64Ptr(5000),
		},
	}

	result := Evaluate(snap, cfg)
	if !result.Pass {
		t.Error("expected PASS when no violations exist")
	}
	if len(result.Violations) != 0 {
		t.Errorf("expected 0 violations, got %d", len(result.Violations))
	}
}
