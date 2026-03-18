package truthcheck

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadTruthSpec(t *testing.T) {
	t.Parallel()
	path := truthSpecPath(t)

	spec, err := LoadTruthSpec(path)
	if err != nil {
		t.Fatalf("failed to load truth spec: %v", err)
	}

	if spec.Impact == nil {
		t.Error("expected impact section")
	}
	if spec.Coverage == nil {
		t.Error("expected coverage section")
	}
	if spec.Redundancy == nil {
		t.Error("expected redundancy section")
	}
	if spec.Fanout == nil {
		t.Error("expected fanout section")
	}
	if spec.Stability == nil {
		t.Error("expected stability section")
	}
	if spec.AI == nil {
		t.Error("expected AI section")
	}
	if spec.Environment == nil {
		t.Error("expected environment section")
	}

	// Verify specific fields.
	if len(spec.Impact.Cases) < 2 {
		t.Errorf("expected >=2 impact cases, got %d", len(spec.Impact.Cases))
	}
	if spec.AI.ExpectedScenarios != 4 {
		t.Errorf("expected 4 AI scenarios, got %d", spec.AI.ExpectedScenarios)
	}
	if len(spec.Coverage.ExpectedUncovered) < 2 {
		t.Errorf("expected >=2 uncovered paths, got %d", len(spec.Coverage.ExpectedUncovered))
	}
}

func TestRun_TerrainWorld(t *testing.T) {
	t.Parallel()
	root := fixtureRoot(t)
	truthPath := truthSpecPath(t)

	report, err := Run(root, truthPath)
	if err != nil {
		t.Fatalf("truth check failed: %v", err)
	}

	if len(report.Categories) == 0 {
		t.Fatal("expected categories in report")
	}

	// Verify all 7 categories are present.
	catNames := map[string]bool{}
	for _, c := range report.Categories {
		catNames[c.Category] = true
		if c.Category == "" {
			t.Error("empty category name")
		}
	}
	for _, expected := range []string{"coverage", "redundancy", "fanout", "stability", "ai", "impact", "environment"} {
		if !catNames[expected] {
			t.Errorf("missing category: %s", expected)
		}
	}

	// AI category should pass (scenarios, prompts, datasets all present).
	for _, c := range report.Categories {
		if c.Category == "ai" {
			if c.Recall < 0.5 {
				t.Errorf("AI recall too low: %.2f", c.Recall)
			}
		}
	}

	// Summary should have reasonable scores.
	if report.Summary.TotalCategories != 7 {
		t.Errorf("expected 7 categories, got %d", report.Summary.TotalCategories)
	}
	if report.Summary.OverallScore < 0.1 {
		t.Errorf("overall score too low: %.2f", report.Summary.OverallScore)
	}
}

func TestComputeScores(t *testing.T) {
	t.Parallel()

	r := TruthCategoryResult{Expected: 10, Matched: 8}
	r.Unexpected = []string{"x", "y"}
	computeScores(&r)

	// Recall: 8/10 = 0.8
	if r.Recall < 0.79 || r.Recall > 0.81 {
		t.Errorf("recall = %.2f, want 0.80", r.Recall)
	}
	// Precision: 8/(8+2) = 0.8
	if r.Precision < 0.79 || r.Precision > 0.81 {
		t.Errorf("precision = %.2f, want 0.80", r.Precision)
	}
	// F1: 2*0.8*0.8/(0.8+0.8) = 0.8
	if r.Score < 0.79 || r.Score > 0.81 {
		t.Errorf("score = %.2f, want 0.80", r.Score)
	}
	if !r.Passed {
		t.Error("expected passed (recall >= 0.5)")
	}
}

func TestRun_TerrainWorld_PerCategory(t *testing.T) {
	t.Parallel()
	root := fixtureRoot(t)
	truthPath := truthSpecPath(t)

	report, err := Run(root, truthPath)
	if err != nil {
		t.Fatalf("truth check failed: %v", err)
	}

	// Each category should pass with recall >= 0.5.
	for _, c := range report.Categories {
		t.Run(c.Category, func(t *testing.T) {
			if !c.Passed {
				t.Errorf("category %s FAILED: recall=%.2f precision=%.2f score=%.2f missing=%v",
					c.Category, c.Recall, c.Precision, c.Score, c.Missing)
			}
			// Precision and recall should both be > 0 (or category is trivially empty).
			if c.Expected > 0 && c.Recall < 0.5 {
				t.Errorf("category %s recall %.2f below threshold 0.5", c.Category, c.Recall)
			}
		})
	}
}

func TestRun_TerrainWorld_Deterministic(t *testing.T) {
	t.Parallel()
	root := fixtureRoot(t)
	truthPath := truthSpecPath(t)

	r1, err := Run(root, truthPath)
	if err != nil {
		t.Fatalf("run 1 failed: %v", err)
	}
	r2, err := Run(root, truthPath)
	if err != nil {
		t.Fatalf("run 2 failed: %v", err)
	}

	if len(r1.Categories) != len(r2.Categories) {
		t.Fatalf("non-deterministic category count: %d vs %d", len(r1.Categories), len(r2.Categories))
	}
	for i := range r1.Categories {
		if r1.Categories[i].Category != r2.Categories[i].Category {
			t.Errorf("category %d differs: %s vs %s", i, r1.Categories[i].Category, r2.Categories[i].Category)
		}
		if r1.Categories[i].Score != r2.Categories[i].Score {
			t.Errorf("category %s score differs: %.4f vs %.4f",
				r1.Categories[i].Category, r1.Categories[i].Score, r2.Categories[i].Score)
		}
	}
}

func TestComputeScores_AllMatched(t *testing.T) {
	t.Parallel()
	r := TruthCategoryResult{Expected: 5, Matched: 5}
	computeScores(&r)
	if r.Score != 1.0 {
		t.Errorf("perfect match score = %.2f, want 1.0", r.Score)
	}
	if !r.Passed {
		t.Error("expected passed")
	}
}

func TestComputeScores_NoneMatched(t *testing.T) {
	t.Parallel()
	r := TruthCategoryResult{Expected: 5, Matched: 0}
	r.Missing = []string{"a", "b", "c", "d", "e"}
	computeScores(&r)
	if r.Recall != 0.0 {
		t.Errorf("zero-match recall = %.2f, want 0.0", r.Recall)
	}
	if r.Passed {
		t.Error("expected not passed")
	}
}

func TestComputeScores_EmptyExpected(t *testing.T) {
	t.Parallel()
	r := TruthCategoryResult{Expected: 0, Matched: 0}
	computeScores(&r)
	// No expectations means recall=0 (nothing to match), precision=1.0 (no false positives).
	// F1 = 0 (no recall). Passed = false (recall < 0.5).
	if r.Precision != 1.0 {
		t.Errorf("empty expected precision = %.2f, want 1.0", r.Precision)
	}
	if r.Recall != 0.0 {
		t.Errorf("empty expected recall = %.2f, want 0.0", r.Recall)
	}
}

func TestLoadTruthSpec_MissingCategory(t *testing.T) {
	t.Parallel()
	// A truth spec with only coverage should parse fine.
	yaml := `
coverage:
  expected_uncovered:
    - path: "src/auth.ts"
      reason: "No tests"
`
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "truth.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	spec, err := LoadTruthSpec(path)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}
	if spec.Coverage == nil {
		t.Error("expected coverage section")
	}
	if spec.Impact != nil {
		t.Error("expected nil impact section")
	}
	if spec.AI != nil {
		t.Error("expected nil AI section")
	}
}

func TestTruthCheckReport_SchemaVersion(t *testing.T) {
	t.Parallel()
	root := fixtureRoot(t)
	truthPath := truthSpecPath(t)

	report, err := Run(root, truthPath)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if report.SchemaVersion != TruthCheckReportSchemaVersion {
		t.Errorf("schemaVersion = %q, want %q", report.SchemaVersion, TruthCheckReportSchemaVersion)
	}
}

func fixtureRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "tests", "fixtures", "terrain-world")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skip("terrain-world fixture not found")
	}
	return root
}

func truthSpecPath(t *testing.T) string {
	t.Helper()
	root := fixtureRoot(t)
	return filepath.Join(root, "tests", "truth", "terrain_truth.yaml")
}
