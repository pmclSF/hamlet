package migration

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestPreviewFile_SimpleJestFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTestFile(t, dir, "test/simple.test.js", `
		import { sum } from '../src/math';

		describe('sum', () => {
			beforeEach(() => {
				// setup
			});

			it('adds numbers', () => {
				expect(sum(1, 2)).toBe(3);
			});
		});
	`)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/simple.test.js", Framework: "jest"},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 10},
		},
	}

	result := PreviewFile(snap, "test/simple.test.js", dir)

	if !result.PreviewAvailable {
		t.Fatal("expected preview to be available")
	}
	if result.SourceFramework != "jest" {
		t.Errorf("sourceFramework = %q, want jest", result.SourceFramework)
	}
	if result.SuggestedTarget != "vitest" {
		t.Errorf("suggestedTarget = %q, want vitest", result.SuggestedTarget)
	}
	if result.Difficulty != "low" {
		t.Errorf("difficulty = %q, want low", result.Difficulty)
	}
	if len(result.Blockers) != 0 {
		t.Errorf("expected 0 blockers, got %d", len(result.Blockers))
	}
	if len(result.SafePatterns) == 0 {
		t.Error("expected safe patterns to be identified")
	}
}

func TestPreviewFile_WithBlockers(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTestFile(t, dir, "test/complex.test.js", `
		expect.extend({
			toBeWithinRange(received, floor, ceiling) {
				return { pass: received >= floor && received <= ceiling };
			}
		});

		it('does something', function(done) {
			fetchData(function() {
				expect(true).toBe(true);
				done();
			});
		});
	`)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/complex.test.js", Framework: "jest"},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 10},
		},
	}

	result := PreviewFile(snap, "test/complex.test.js", dir)

	if !result.PreviewAvailable {
		t.Fatal("expected preview to be available")
	}
	if len(result.Blockers) < 2 {
		t.Fatalf("expected at least 2 blockers, got %d", len(result.Blockers))
	}
	if result.Difficulty == "low" {
		t.Error("expected difficulty > low for file with blockers")
	}
}

func TestPreviewFile_NotInSnapshot(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{},
	}

	result := PreviewFile(snap, "nonexistent.test.js", "/tmp")

	if result.PreviewAvailable {
		t.Error("expected preview not available for missing file")
	}
	if result.Difficulty != "unknown" {
		t.Errorf("difficulty = %q, want unknown", result.Difficulty)
	}
}

func TestPreviewFile_NonJSFramework(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test_main.go", Framework: "go-testing"},
		},
	}

	result := PreviewFile(snap, "test_main.go", "/tmp")

	if result.PreviewAvailable {
		t.Error("expected preview not available for Go framework")
	}
	if result.Difficulty != "unknown" {
		t.Errorf("difficulty = %q, want unknown", result.Difficulty)
	}
}

func TestPreviewFile_InferTarget_Cypress(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTestFile(t, dir, "e2e/login.cy.js", `
		describe('Login', () => {
			it('logs in', () => {
				cy.visit('/login');
				cy.get('#user').type('admin');
			});
		});
	`)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "e2e/login.cy.js", Framework: "cypress"},
		},
		Frameworks: []models.Framework{
			{Name: "cypress", Type: models.FrameworkTypeE2E, FileCount: 5},
		},
	}

	result := PreviewFile(snap, "e2e/login.cy.js", dir)

	if result.SuggestedTarget != "playwright" {
		t.Errorf("suggestedTarget = %q, want playwright", result.SuggestedTarget)
	}
}

func TestPreviewScope_MixedDifficulty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTestFile(t, dir, "test/easy.test.js", `
		test('simple', () => {
			expect(1).toBe(1);
		});
	`)
	writeTestFile(t, dir, "test/hard.test.js", `
		expect.extend({
			custom() { return { pass: true }; }
		});
		Cypress.Commands.add('foo', () => {});
		it('complex', function(done) { done(); });
	`)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/easy.test.js", Framework: "jest"},
			{Path: "test/hard.test.js", Framework: "cypress"},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 1},
			{Name: "cypress", Type: models.FrameworkTypeE2E, FileCount: 1},
		},
	}

	results := PreviewScope(snap, "test", dir)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Should be sorted by difficulty (hardest first)
	if results[0].Difficulty == "low" && results[1].Difficulty != "low" {
		t.Error("expected harder files to appear first")
	}
}

func TestClassifyDifficulty(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		blockers []PreviewBlocker
		want     string
	}{
		{"no blockers", nil, "low"},
		{"one low-impact", []PreviewBlocker{{Type: BlockerDeprecatedPattern}}, "low"},
		{"one high-impact", []PreviewBlocker{{Type: BlockerCustomMatcher}}, "medium"},
		{"two high-impact", []PreviewBlocker{
			{Type: BlockerCustomMatcher},
			{Type: BlockerUnsupportedSetup},
		}, "high"},
		{"four low-impact", []PreviewBlocker{
			{Type: BlockerDeprecatedPattern},
			{Type: BlockerDeprecatedPattern},
			{Type: BlockerDynamicGeneration},
			{Type: BlockerDeprecatedPattern},
		}, "high"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := classifyDifficulty(tt.blockers, &models.TestFile{})
			if got != tt.want {
				t.Errorf("classifyDifficulty() = %q, want %q", got, tt.want)
			}
		})
	}
}
