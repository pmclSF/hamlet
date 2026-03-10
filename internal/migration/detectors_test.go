package migration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", name, err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestDeprecatedPatternDetector_DoneCallback(t *testing.T) {
	t.Parallel()
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

func TestDeprecatedPatternDetector_DoneCallbackArrowFunction(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTestFile(t, dir, "test/old-arrow.test.js", `
		it('does something', (done) => {
			setImmediate(() => done());
		});
	`)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/old-arrow.test.js", Framework: "jest"},
		},
	}

	d := &DeprecatedPatternDetector{RepoRoot: dir}
	signals := d.Detect(snap)
	if len(signals) == 0 {
		t.Fatal("expected done-callback signal for arrow function syntax")
	}
}

func TestDeprecatedPatternDetector_NoDeprecated(t *testing.T) {
	t.Parallel()
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

func TestDeprecatedPatternDetector_IgnoresCommentsAndStrings(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTestFile(t, dir, "test/noise.test.js", `
		// it('fake', function(done) { done(); });
		const txt = "it('fake', (done) => done())";
		test('real', async () => {
			expect(txt).toContain('done');
		});
	`)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/noise.test.js", Framework: "jest"},
		},
	}

	d := &DeprecatedPatternDetector{RepoRoot: dir}
	signals := d.Detect(snap)
	if len(signals) != 0 {
		t.Fatalf("expected 0 signals for comment/string-only deprecated patterns, got %d", len(signals))
	}
}

func TestDynamicTestGenerationDetector(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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

func TestDynamicTestGenerationDetector_IgnoresCommentsAndStrings(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTestFile(t, dir, "test/noise-dynamic.test.js", `
		// test.each([1,2])('fake', () => {});
		const note = "cases.forEach(() => test('fake', () => {}))";
		test('real', () => {
			expect(note).toContain('each');
		});
	`)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/noise-dynamic.test.js", Framework: "jest"},
		},
	}

	d := &DynamicTestGenerationDetector{RepoRoot: dir}
	signals := d.Detect(snap)
	if len(signals) != 0 {
		t.Fatalf("expected 0 dynamic-generation signals for comment/string-only patterns, got %d", len(signals))
	}
}

func TestCustomMatcherDetector(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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

func TestCustomMatcherDetector_IgnoresCommentsAndStrings(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTestFile(t, dir, "test/noise-custom.test.js", `
		/* expect.extend({ toBeFoo() {} }) */
		const note = "chai.use(plugin)";
		test('real', () => {
			expect(note).toContain('plugin');
		});
	`)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/noise-custom.test.js", Framework: "jest"},
		},
	}

	d := &CustomMatcherDetector{RepoRoot: dir}
	signals := d.Detect(snap)
	if len(signals) != 0 {
		t.Fatalf("expected 0 custom-matcher signals for comment/string-only patterns, got %d", len(signals))
	}
}

func TestFrameworkMigrationDetector_MultipleUnitFrameworks(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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

func TestUnsupportedSetupDetector_CypressCommands(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTestFile(t, dir, "test/support/commands.js", `
		Cypress.Commands.add('login', (user, pass) => {
			cy.visit('/login');
			cy.get('#username').type(user);
			cy.get('#password').type(pass);
			cy.get('form').submit();
		});
	`)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/support/commands.js", Framework: "cypress"},
		},
	}

	d := &UnsupportedSetupDetector{RepoRoot: dir}
	signals := d.Detect(snap)

	// Cypress custom commands is a hard blocker, so we expect both
	// an unsupportedSetup signal and a migrationBlocker signal.
	if len(signals) != 2 {
		t.Fatalf("expected 2 signals (unsupportedSetup + migrationBlocker), got %d", len(signals))
	}
	if signals[0].Type != "unsupportedSetup" {
		t.Errorf("signals[0].Type = %q, want unsupportedSetup", signals[0].Type)
	}
	if signals[1].Type != "migrationBlocker" {
		t.Errorf("signals[1].Type = %q, want migrationBlocker", signals[1].Type)
	}
}

func TestUnsupportedSetupDetector_FrameworkTestContext(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTestFile(t, dir, "test/slow.test.js", `
		describe('slow suite', function() {
			this.timeout(10000);
			it('takes a while', function() {
				this.slow(5000);
			});
		});
	`)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/slow.test.js", Framework: "mocha"},
		},
	}

	d := &UnsupportedSetupDetector{RepoRoot: dir}
	signals := d.Detect(snap)

	if len(signals) == 0 {
		t.Fatal("expected at least 1 signal for framework test context")
	}
	if signals[0].Type != "unsupportedSetup" {
		t.Errorf("type = %q, want unsupportedSetup", signals[0].Type)
	}
}

func TestUnsupportedSetupDetector_FrameworkTestContext_NotFlaggedOutsideMocha(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTestFile(t, dir, "test/slow.test.js", `
		describe('suite', () => {
			const helper = "this.timeout(5000)";
			it('works', () => {
				expect(helper).toContain('timeout');
			});
		});
	`)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/slow.test.js", Framework: "jest"},
		},
	}

	d := &UnsupportedSetupDetector{RepoRoot: dir}
	signals := d.Detect(snap)
	if len(signals) != 0 {
		t.Fatalf("expected 0 signals for non-mocha framework context text, got %d", len(signals))
	}
}

func TestUnsupportedSetupDetector_NoUnsupported(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTestFile(t, dir, "test/clean.test.js", `
		describe('clean suite', () => {
			beforeEach(() => {
				// standard setup
			});
			it('works', () => {
				expect(true).toBe(true);
			});
		});
	`)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/clean.test.js", Framework: "jest"},
		},
	}

	d := &UnsupportedSetupDetector{RepoRoot: dir}
	signals := d.Detect(snap)

	if len(signals) != 0 {
		t.Errorf("expected 0 signals for clean test, got %d", len(signals))
	}
}

func TestUnsupportedSetupDetector_IgnoresCommentsAndStrings(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeTestFile(t, dir, "test/support/noise.js", `
		// Cypress.Commands.add('fake', () => {});
		const note = "Cypress.on('task', () => {})";
		describe('suite', () => {
			it('works', () => {
				expect(note).toContain('Cypress');
			});
		});
	`)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/support/noise.js", Framework: "cypress"},
		},
	}

	d := &UnsupportedSetupDetector{RepoRoot: dir}
	signals := d.Detect(snap)
	if len(signals) != 0 {
		t.Fatalf("expected 0 unsupported-setup signals for comment/string-only patterns, got %d", len(signals))
	}
}

func TestFrameworkMigrationDetector_UnitPlusE2E(t *testing.T) {
	t.Parallel()
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
