package testcase

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestDetectCollisions_NoCollisions(t *testing.T) {
	t.Parallel()
	cases := []models.TestCase{
		{TestID: "a", CanonicalIdentity: "path::suite::test1"},
		{TestID: "b", CanonicalIdentity: "path::suite::test2"},
	}
	result, diagnostics := DetectAndResolveCollisions(cases)
	if len(diagnostics) != 0 {
		t.Errorf("expected no diagnostics, got %d", len(diagnostics))
	}
	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}
}

func TestDetectCollisions_WithCollision(t *testing.T) {
	t.Parallel()
	cases := []models.TestCase{
		{TestID: "a", CanonicalIdentity: "path::suite::test", Line: 10},
		{TestID: "a", CanonicalIdentity: "path::suite::test", Line: 20},
		{TestID: "b", CanonicalIdentity: "path::suite::other"},
	}
	result, diagnostics := DetectAndResolveCollisions(cases)
	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diagnostics))
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 results, got %d", len(result))
	}

	// First occurrence keeps original ID, second gets disambiguated.
	ids := map[string]bool{}
	for _, tc := range result {
		if ids[tc.TestID] {
			t.Errorf("duplicate TestID after disambiguation: %s", tc.TestID)
		}
		ids[tc.TestID] = true
	}

	// Check diagnostic.
	d := diagnostics[0]
	if len(d.Occurrences) != 2 {
		t.Errorf("expected 2 occurrences, got %d", len(d.Occurrences))
	}
}

func TestDetectCollisions_Deterministic(t *testing.T) {
	t.Parallel()
	cases := []models.TestCase{
		{TestID: "x", CanonicalIdentity: "path::suite::dup", Line: 30},
		{TestID: "x", CanonicalIdentity: "path::suite::dup", Line: 10},
		{TestID: "x", CanonicalIdentity: "path::suite::dup", Line: 20},
	}
	r1, _ := DetectAndResolveCollisions(cases)
	r2, _ := DetectAndResolveCollisions(cases)

	if len(r1) != len(r2) {
		t.Fatal("different result lengths")
	}
	for i := range r1 {
		if r1[i].TestID != r2[i].TestID {
			t.Errorf("non-deterministic: r1[%d].TestID=%s, r2[%d].TestID=%s", i, r1[i].TestID, i, r2[i].TestID)
		}
	}
}

func TestDuplicateNamesInDifferentSuites(t *testing.T) {
	t.Parallel()
	src := `
describe('Suite A', () => {
  it('works', () => {});
});
	describe('Suite B', () => {
  it('works', () => {});
});
`
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "test.js"), []byte(src), 0644); err != nil {
		t.Fatalf("write test.js: %v", err)
	}

	cases := Extract(dir, "test.js", "jest")
	if len(cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(cases))
	}

	// Different suites should produce different IDs even with same test name.
	if cases[0].TestID == cases[1].TestID {
		t.Error("same test name in different suites should have different IDs")
	}
}

func TestDuplicateNamesInSameFile_SameSuite(t *testing.T) {
	t.Parallel()
	// This is a real-world antipattern: two tests with identical name in same suite.
	src := `
describe('Suite', () => {
  it('works', () => {});
  it('works', () => {});
});
`
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "dup.test.js"), []byte(src), 0644); err != nil {
		t.Fatalf("write dup.test.js: %v", err)
	}

	cases := Extract(dir, "dup.test.js", "jest")
	if len(cases) != 2 {
		t.Fatalf("expected 2 cases, got %d", len(cases))
	}

	// Will have same canonical identity — collision detection should resolve.
	modelCases := ToModels(cases)
	resolved, diagnostics := DetectAndResolveCollisions(modelCases)

	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 collision diagnostic, got %d", len(diagnostics))
	}
	if len(resolved) != 2 {
		t.Fatalf("expected 2 resolved cases, got %d", len(resolved))
	}
	if resolved[0].TestID == resolved[1].TestID {
		t.Error("collision resolution should produce unique IDs")
	}
}

func TestDynamicTestGeneration(t *testing.T) {
	t.Parallel()
	src := `
const cases = [1, 2, 3];
cases.forEach(n => {
  it('handles ' + n, () => {});
});
`
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "dynamic.test.js"), []byte(src), 0644); err != nil {
		t.Fatalf("write dynamic.test.js: %v", err)
	}

	cases := Extract(dir, "dynamic.test.js", "jest")
	// Dynamic tests may or may not be extractable.
	// The forEach pattern may capture the it() inside.
	// What matters is that extraction doesn't crash.
	sort.Slice(cases, func(i, j int) bool { return cases[i].Line < cases[j].Line })
	for _, c := range cases {
		if c.TestID == "" {
			t.Error("all extracted cases should have non-empty TestID")
		}
	}
}
