package migration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755)
	os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
}

func TestDeprecatedPatternDetector_DoneCallback(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "test/old.test.js", `
		it('does something', function(done) {
			fetchData(function() {
				expect(true).toBe(true);
				done();
			});
		});
	`)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/old.test.js", Framework: "jest"},
		},
	}

	d := &DeprecatedPatternDetector{RepoRoot: dir}
	signals := d.Detect(snap)

	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}
	if signals[0].Type != "deprecatedTestPattern" {
		t.Errorf("type = %q, want deprecatedTestPattern", signals[0].Type)
	}
}

func TestDeprecatedPatternDetector_NoDeprecated(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "test/modern.test.js", `
		it('does something', async () => {
			const data = await fetchData();
			expect(data).toBeDefined();
		});
	`)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/modern.test.js", Framework: "jest"},
		},
	}

	d := &DeprecatedPatternDetector{RepoRoot: dir}
	signals := d.Detect(snap)

	if len(signals) != 0 {
		t.Errorf("expected 0 signals for modern test, got %d", len(signals))
	}
}

func TestDynamicTestGenerationDetector(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "test/dynamic.test.js", `
		const cases = [1, 2, 3];
		test.each(cases)('works for %d', (n) => {
			expect(n).toBeGreaterThan(0);
		});
	`)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/dynamic.test.js", Framework: "jest"},
		},
	}

	d := &DynamicTestGenerationDetector{RepoRoot: dir}
	signals := d.Detect(snap)

	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}
	if signals[0].Type != "dynamicTestGeneration" {
		t.Errorf("type = %q, want dynamicTestGeneration", signals[0].Type)
	}
}

func TestDynamicTestGenerationDetector_NoDynamic(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "test/static.test.js", `
		test('works', () => {
			expect(1).toBe(1);
		});
	`)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/static.test.js", Framework: "jest"},
		},
	}

	d := &DynamicTestGenerationDetector{RepoRoot: dir}
	signals := d.Detect(snap)

	if len(signals) != 0 {
		t.Errorf("expected 0 signals, got %d", len(signals))
	}
}

func TestCustomMatcherDetector(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "test/custom.test.js", `
		expect.extend({
			toBeWithinRange(received, floor, ceiling) {
				const pass = received >= floor && received <= ceiling;
				return { pass, message: () => 'expected in range' };
			}
		});
	`)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/custom.test.js", Framework: "jest"},
		},
	}

	d := &CustomMatcherDetector{RepoRoot: dir}
	signals := d.Detect(snap)

	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}
	if signals[0].Type != "customMatcherRisk" {
		t.Errorf("type = %q, want customMatcherRisk", signals[0].Type)
	}
}

func TestCustomMatcherDetector_NoCustom(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "test/plain.test.js", `
		test('works', () => {
			expect(1).toBe(1);
		});
	`)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/plain.test.js", Framework: "jest"},
		},
	}

	d := &CustomMatcherDetector{RepoRoot: dir}
	signals := d.Detect(snap)

	if len(signals) != 0 {
		t.Errorf("expected 0 signals, got %d", len(signals))
	}
}

func TestFrameworkMigrationDetector_MultipleUnitFrameworks(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{Name: "test-repo"},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 50},
			{Name: "mocha", Type: models.FrameworkTypeUnit, FileCount: 30},
		},
	}

	d := &FrameworkMigrationDetector{}
	signals := d.Detect(snap)

	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}
	if signals[0].Type != "frameworkMigration" {
		t.Errorf("type = %q, want frameworkMigration", signals[0].Type)
	}
}

func TestFrameworkMigrationDetector_SingleFramework(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 50},
		},
	}

	d := &FrameworkMigrationDetector{}
	signals := d.Detect(snap)

	if len(signals) != 0 {
		t.Errorf("expected 0 signals for single framework, got %d", len(signals))
	}
}

func TestFrameworkMigrationDetector_UnitPlusE2E(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 50},
			{Name: "playwright", Type: models.FrameworkTypeE2E, FileCount: 10},
		},
	}

	d := &FrameworkMigrationDetector{}
	signals := d.Detect(snap)

	if len(signals) != 0 {
		t.Errorf("expected 0 signals for unit+e2e mix, got %d", len(signals))
	}
}
