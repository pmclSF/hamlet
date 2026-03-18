package analysis

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// ExtractFixtures scans test files and returns detected fixture surfaces.
// It reads each test file's content and applies language-specific fixture
// detection patterns.
func ExtractFixtures(root string, testFiles []models.TestFile) []models.FixtureSurface {
	results := make([][]models.FixtureSurface, len(testFiles))
	parallelForEachIndex(len(testFiles), func(i int) {
		tf := &testFiles[i]
		content, err := os.ReadFile(filepath.Join(root, tf.Path))
		if err != nil {
			return
		}
		lang := frameworkLanguage(tf.Framework)
		results[i] = detectFixtures(string(content), tf.Path, lang, tf.Framework)
	})

	var fixtures []models.FixtureSurface
	for _, batch := range results {
		fixtures = append(fixtures, batch...)
	}
	// Ensure every fixture carries a Reason field for evidence traceability.
	assignFixtureReasons(fixtures)
	return fixtures
}

// assignFixtureReasons fills in the Reason field on fixtures that don't
// have one, using a standardized format based on kind and detection tier.
func assignFixtureReasons(fixtures []models.FixtureSurface) {
	for i := range fixtures {
		if fixtures[i].Reason != "" {
			continue
		}
		fs := &fixtures[i]
		detectorID := fixtureDetectorID(fs)
		desc := string(fs.Kind) + " '" + fs.Name + "'"
		if fs.Scope != "" && fs.Scope != "unknown" {
			desc += " (scope: " + fs.Scope + ")"
		}
		fs.Reason = models.FormatReason(detectorID, desc)
	}
}

func fixtureDetectorID(fs *models.FixtureSurface) string {
	switch {
	case fs.DetectionTier == models.TierStructural && fs.Language == "go" && fs.Name == "TestMain":
		return models.DetectorFixtureGoTestMain
	case fs.DetectionTier == models.TierStructural && fs.Language == "go":
		return models.DetectorFixtureGoHelper
	case fs.DetectionTier == models.TierStructural && fs.Language == "python":
		return models.DetectorFixturePyFixture
	case fs.DetectionTier == models.TierStructural && fs.Language == "java":
		return models.DetectorFixtureJavaLifecycle
	case fs.Kind == models.FixtureSetupHook || fs.Kind == models.FixtureTeardownHook:
		return models.DetectorFixtureLifecycleHook
	case fs.Kind == models.FixtureBuilder:
		return models.DetectorFixtureBuilder
	case fs.Kind == models.FixtureMockProvider:
		return models.DetectorFixtureMockProvider
	case fs.Kind == models.FixtureDataLoader:
		return models.DetectorFixtureDataLoader
	default:
		return models.DetectorFixtureBuilder
	}
}

// detectFixtures dispatches to language-specific fixture detection.
func detectFixtures(src, relPath, lang, framework string) []models.FixtureSurface {
	switch lang {
	case "js":
		return detectJSFixtures(src, relPath, framework)
	case "python":
		return detectPythonFixtures(src, relPath, framework)
	case "go":
		return detectGoFixtures(src, relPath)
	case "java":
		return detectJavaFixtures(src, relPath)
	default:
		return nil
	}
}

// --- JavaScript/TypeScript fixture detection ---

var (
	// beforeEach(() => { ... }) or beforeEach(function() { ... })
	jsBeforeEachPattern = regexp.MustCompile(`\b(beforeEach|beforeAll|afterEach|afterAll)\s*\(`)

	// Shared helper/builder/factory functions in test files:
	// function createUser(...), const buildOrder = ..., export function makeFixture(...)
	jsTestHelperPattern = regexp.MustCompile(`(?:export\s+)?(?:(?:async\s+)?function\s+|(?:const|let|var)\s+)(\w*(?:[Bb]uild|[Cc]reate|[Mm]ake|[Ss]etup|[Mm]ock|[Ss]tub|[Ff]ake|[Ff]ixture|[Ff]actory|[Ss]eed|[Hh]elper|[Pp]rovid)\w*)\s*[=(]`)

	// jest.fn() / vi.fn() / sinon.stub() mock providers
	jsMockProviderPattern = regexp.MustCompile(`(?:const|let|var)\s+(\w+)\s*=\s*(?:jest\.fn|vi\.fn|sinon\.(?:stub|mock|spy)|new\s+Mock\w*)\s*\(`)

	// Shared test data: const testData = [...], const fixtures = {...}
	jsTestDataPattern = regexp.MustCompile(`(?:const|let|var)\s+(\w*(?:[Tt]est[Dd]ata|[Ff]ixtures?|[Ss]ample[Dd]ata|[Mm]ock[Dd]ata|[Ss]eed[Dd]ata|[Ss]tub[Dd]ata)\w*)\s*=`)
)

func detectJSFixtures(src, relPath, framework string) []models.FixtureSurface {
	lines := strings.Split(src, "\n")
	var fixtures []models.FixtureSurface
	seen := map[string]bool{}

	add := func(f models.FixtureSurface) {
		if seen[f.FixtureID] {
			return
		}
		seen[f.FixtureID] = true
		fixtures = append(fixtures, f)
	}

	// Track describe nesting for scope inference.
	describeDepth := 0
	isSharedFile := isSharedFixtureFile(relPath)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track describe blocks for scope.
		describeDepth += strings.Count(trimmed, "describe(")
		describeDepth += strings.Count(trimmed, "describe.each(")

		// Detect lifecycle hooks.
		if m := jsBeforeEachPattern.FindStringSubmatch(trimmed); m != nil {
			hookName := m[1]
			kind := models.FixtureSetupHook
			if strings.HasPrefix(hookName, "after") {
				kind = models.FixtureTeardownHook
			}
			scope := "test"
			if strings.HasSuffix(hookName, "All") {
				scope = "suite"
			}
			add(models.FixtureSurface{
				FixtureID:     models.BuildFixtureID(relPath, hookName, scope),
				Name:          hookName,
				Path:          relPath,
				Kind:          kind,
				Scope:         scope,
				Language:      "js",
				Framework:     framework,
				Line:          i + 1,
				Shared:        isSharedFile || describeDepth == 0,
				DetectionTier: models.TierPattern,
				Confidence:    0.95,
			})
		}

		// Detect shared test data (checked before helpers — more specific).
		if m := jsTestDataPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			add(models.FixtureSurface{
				FixtureID:     models.BuildFixtureID(relPath, name, ""),
				Name:          name,
				Path:          relPath,
				Kind:          models.FixtureDataLoader,
				Scope:         "module",
				Language:      "js",
				Framework:     framework,
				Line:          i + 1,
				Shared:        isSharedFile || strings.Contains(line, "export"),
				DetectionTier: models.TierPattern,
				Confidence:    0.85,
			})
		}

		// Detect mock providers (jest.fn, vi.fn, sinon.stub).
		if m := jsMockProviderPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			add(models.FixtureSurface{
				FixtureID:     models.BuildFixtureID(relPath, name, ""),
				Name:          name,
				Path:          relPath,
				Kind:          models.FixtureMockProvider,
				Scope:         "unknown",
				Language:      "js",
				Framework:     framework,
				Line:          i + 1,
				Shared:        isSharedFile,
				DetectionTier: models.TierPattern,
				Confidence:    0.9,
			})
		}

		// Detect helper/builder functions.
		if m := jsTestHelperPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			kind := classifyJSHelper(name)
			add(models.FixtureSurface{
				FixtureID:     models.BuildFixtureID(relPath, name, ""),
				Name:          name,
				Path:          relPath,
				Kind:          kind,
				Scope:         "unknown",
				Language:      "js",
				Framework:     framework,
				Line:          i + 1,
				Shared:        isSharedFile || strings.Contains(line, "export"),
				DetectionTier: models.TierPattern,
				Confidence:    0.85,
			})
		}
	}

	return fixtures
}

func classifyJSHelper(name string) models.FixtureKind {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "mock") || strings.Contains(lower, "stub") || strings.Contains(lower, "fake"):
		return models.FixtureMockProvider
	case strings.Contains(lower, "seed") || strings.Contains(lower, "fixture"):
		return models.FixtureDataLoader
	case strings.Contains(lower, "setup"):
		return models.FixtureSetupHook
	default:
		return models.FixtureBuilder
	}
}

// --- Python fixture detection ---

var (
	// @pytest.fixture / @pytest.fixture(scope="session")
	pyFixtureDecoratorPattern = regexp.MustCompile(`@pytest\.fixture(?:\s*\(([^)]*)\))?`)
	pyFixtureScopePattern     = regexp.MustCompile(`scope\s*=\s*['"](\w+)['"]`)

	// class setUp/tearDown methods
	pySetUpPattern = regexp.MustCompile(`^\s+def\s+(setUp|tearDown|setUpClass|tearDownClass|setUpModule|tearDownModule)\s*\(`)

	// conftest.py helper functions: def create_user(...), def mock_db(...)
	pyTestHelperPattern = regexp.MustCompile(`^def\s+(\w*(?:create|build|make|setup|mock|stub|fake|fixture|factory|seed|helper|provide)\w*)\s*\(`)

	// conftest.py data: TEST_DATA = ..., SAMPLE_FIXTURES = ...
	pyTestDataPattern = regexp.MustCompile(`^(\w*(?:TEST_DATA|FIXTURES?|SAMPLE_DATA|MOCK_DATA|SEED_DATA|STUB_DATA)\w*)\s*=`)
)

func detectPythonFixtures(src, relPath, framework string) []models.FixtureSurface {
	lines := strings.Split(src, "\n")
	var fixtures []models.FixtureSurface
	seen := map[string]bool{}
	isSharedFile := isSharedFixtureFile(relPath)

	add := func(f models.FixtureSurface) {
		if seen[f.FixtureID] {
			return
		}
		seen[f.FixtureID] = true
		fixtures = append(fixtures, f)
	}

	// Track pending @pytest.fixture decorator.
	pendingFixtureScope := ""
	hasPendingFixture := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect @pytest.fixture decorators.
		if m := pyFixtureDecoratorPattern.FindStringSubmatch(trimmed); m != nil {
			hasPendingFixture = true
			pendingFixtureScope = "test" // default
			if len(m) > 1 && m[1] != "" {
				if sm := pyFixtureScopePattern.FindStringSubmatch(m[1]); sm != nil {
					pendingFixtureScope = sm[1]
				}
			}
			continue
		}

		// If we have a pending fixture decorator, the next def is the fixture.
		if hasPendingFixture {
			if m := pyDefPattern.FindStringSubmatch(trimmed); m != nil {
				name := m[1]
				add(models.FixtureSurface{
					FixtureID:     models.BuildFixtureID(relPath, name, pendingFixtureScope),
					Name:          name,
					Path:          relPath,
					Kind:          models.FixtureSetupHook,
					Scope:         pendingFixtureScope,
					Language:      "python",
					Framework:     framework,
					Line:          i + 1,
					Shared:        isSharedFile,
					DetectionTier: models.TierStructural,
					Confidence:    0.95,
				})
				hasPendingFixture = false
				continue
			}
			// Non-def line after decorator — reset.
			if trimmed != "" && !strings.HasPrefix(trimmed, "@") && !strings.HasPrefix(trimmed, "#") {
				hasPendingFixture = false
			}
		}

		// Detect setUp/tearDown methods.
		if m := pySetUpPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			kind := models.FixtureSetupHook
			scope := "test"
			if strings.Contains(name, "tearDown") {
				kind = models.FixtureTeardownHook
			}
			if strings.Contains(name, "Class") {
				scope = "suite"
			}
			if strings.Contains(name, "Module") {
				scope = "module"
			}
			add(models.FixtureSurface{
				FixtureID:     models.BuildFixtureID(relPath, name, scope),
				Name:          name,
				Path:          relPath,
				Kind:          kind,
				Scope:         scope,
				Language:      "python",
				Framework:     framework,
				Line:          i + 1,
				Shared:        false,
				DetectionTier: models.TierPattern,
				Confidence:    0.95,
			})
		}

		// Detect helper functions (in conftest or test helpers).
		if isSharedFile {
			if m := pyTestHelperPattern.FindStringSubmatch(trimmed); m != nil {
				name := m[1]
				kind := classifyPythonHelper(name)
				fid := models.BuildFixtureID(relPath, name, "")
				if !seen[fid] {
					add(models.FixtureSurface{
						FixtureID:     fid,
						Name:          name,
						Path:          relPath,
						Kind:          kind,
						Scope:         "unknown",
						Language:      "python",
						Framework:     framework,
						Line:          i + 1,
						Shared:        true,
						DetectionTier: models.TierPattern,
						Confidence:    0.8,
					})
				}
			}
		}

		// Detect shared test data constants.
		if m := pyTestDataPattern.FindStringSubmatch(trimmed); m != nil {
			name := m[1]
			add(models.FixtureSurface{
				FixtureID:     models.BuildFixtureID(relPath, name, "module"),
				Name:          name,
				Path:          relPath,
				Kind:          models.FixtureDataLoader,
				Scope:         "module",
				Language:      "python",
				Framework:     framework,
				Line:          i + 1,
				Shared:        isSharedFile,
				DetectionTier: models.TierPattern,
				Confidence:    0.8,
			})
		}
	}

	return fixtures
}

func classifyPythonHelper(name string) models.FixtureKind {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "mock") || strings.Contains(lower, "stub") || strings.Contains(lower, "fake"):
		return models.FixtureMockProvider
	case strings.Contains(lower, "seed") || strings.Contains(lower, "fixture"):
		return models.FixtureDataLoader
	case strings.Contains(lower, "setup"):
		return models.FixtureSetupHook
	default:
		return models.FixtureBuilder
	}
}

// --- Go fixture detection ---

var (
	// Go test helper functions: func TestHelper(t *testing.T, ...) or
	// functions that accept *testing.T / *testing.B but don't start with Test/Benchmark.
	goTestHelperPattern = regexp.MustCompile(`^\s*func\s+([a-z]\w*)\s*\([^)]*\*testing\.[TB]`)

	// TestMain(m *testing.M) — module-scoped setup.
	goTestMainPattern = regexp.MustCompile(`^\s*func\s+TestMain\s*\(\s*\w+\s+\*testing\.M\s*\)`)

	// Helper builders: func newTestServer(...), func createTestDB(...)
	goTestBuilderPattern = regexp.MustCompile(`^\s*func\s+((?:new|create|make|build|setup|mock|stub|fake|seed|fixture|helper)\w*)\s*\(`)

	// Shared test data: var testFixtures = ..., var mockResponse = ...
	goTestDataPattern = regexp.MustCompile(`^\s*var\s+(\w*(?:test[A-Z]\w*|mock[A-Z]\w*|fake[A-Z]\w*|fixture[A-Z]\w*|sample[A-Z]\w*|seed[A-Z]\w*))\s*=`)
)

func detectGoFixtures(src, relPath string) []models.FixtureSurface {
	lines := strings.Split(src, "\n")
	var fixtures []models.FixtureSurface
	seen := map[string]bool{}
	isSharedFile := isSharedFixtureFile(relPath)

	add := func(f models.FixtureSurface) {
		if seen[f.FixtureID] {
			return
		}
		seen[f.FixtureID] = true
		fixtures = append(fixtures, f)
	}

	for i, line := range lines {
		// Detect TestMain.
		if goTestMainPattern.MatchString(line) {
			add(models.FixtureSurface{
				FixtureID:     models.BuildFixtureID(relPath, "TestMain", "module"),
				Name:          "TestMain",
				Path:          relPath,
				Kind:          models.FixtureSetupHook,
				Scope:         "module",
				Language:      "go",
				Framework:     "go-testing",
				Line:          i + 1,
				Shared:        true,
				DetectionTier: models.TierStructural,
				Confidence:    0.99,
			})
			continue
		}

		// Detect test helper functions (accept *testing.T but not Test*).
		if m := goTestHelperPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			add(models.FixtureSurface{
				FixtureID:     models.BuildFixtureID(relPath, name, ""),
				Name:          name,
				Path:          relPath,
				Kind:          models.FixtureHelper,
				Scope:         "test",
				Language:      "go",
				Framework:     "go-testing",
				Line:          i + 1,
				Shared:        isSharedFile,
				DetectionTier: models.TierStructural,
				Confidence:    0.95,
			})
		}

		// Detect builder/factory functions.
		if m := goTestBuilderPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			fid := models.BuildFixtureID(relPath, name, "")
			if seen[fid] {
				continue
			}
			kind := classifyGoHelper(name)
			add(models.FixtureSurface{
				FixtureID:     fid,
				Name:          name,
				Path:          relPath,
				Kind:          kind,
				Scope:         "unknown",
				Language:      "go",
				Framework:     "go-testing",
				Line:          i + 1,
				Shared:        isSharedFile,
				DetectionTier: models.TierPattern,
				Confidence:    0.85,
			})
		}

		// Detect shared test data variables.
		if m := goTestDataPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			add(models.FixtureSurface{
				FixtureID:     models.BuildFixtureID(relPath, name, "module"),
				Name:          name,
				Path:          relPath,
				Kind:          models.FixtureDataLoader,
				Scope:         "module",
				Language:      "go",
				Framework:     "go-testing",
				Line:          i + 1,
				Shared:        isSharedFile,
				DetectionTier: models.TierPattern,
				Confidence:    0.8,
			})
		}
	}

	return fixtures
}

func classifyGoHelper(name string) models.FixtureKind {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "mock") || strings.Contains(lower, "stub") || strings.Contains(lower, "fake"):
		return models.FixtureMockProvider
	case strings.Contains(lower, "seed") || strings.Contains(lower, "fixture"):
		return models.FixtureDataLoader
	case strings.Contains(lower, "setup"):
		return models.FixtureSetupHook
	default:
		return models.FixtureBuilder
	}
}

// --- Java fixture detection ---

var (
	// @Before, @After, @BeforeEach, @AfterEach, @BeforeAll, @AfterAll, @BeforeClass, @AfterClass
	javaLifecyclePattern = regexp.MustCompile(`@(Before|After|BeforeEach|AfterEach|BeforeAll|AfterAll|BeforeClass|AfterClass)\b`)

	// Lifecycle method after annotation — may lack access modifier (e.g., "static void initAll()").
	javaLifecycleMethodPattern = regexp.MustCompile(`\b(?:public\s+|private\s+|protected\s+)?(?:static\s+)?(?:[\w<>\[\]]+\s+)(\w+)\s*\(`)

	// Helper methods in test classes
	javaTestHelperPattern = regexp.MustCompile(`\b(?:private|protected|public)?\s*(?:static\s+)?(?:[\w<>\[\]]+\s+)(create\w+|build\w+|make\w+|setup\w+|mock\w+|stub\w+|fake\w+|fixture\w+|factory\w+|seed\w+|helper\w+)\s*\(`)
)

func detectJavaFixtures(src, relPath string) []models.FixtureSurface {
	lines := strings.Split(src, "\n")
	var fixtures []models.FixtureSurface
	seen := map[string]bool{}

	add := func(f models.FixtureSurface) {
		if seen[f.FixtureID] {
			return
		}
		seen[f.FixtureID] = true
		fixtures = append(fixtures, f)
	}

	pendingAnnotation := ""
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect lifecycle annotations.
		if m := javaLifecyclePattern.FindStringSubmatch(trimmed); m != nil {
			pendingAnnotation = m[1]
			continue
		}

		// If we have a pending annotation, the next method is the fixture.
		if pendingAnnotation != "" {
			if m := javaLifecycleMethodPattern.FindStringSubmatch(line); m != nil {
				name := m[1]
				kind := models.FixtureSetupHook
				scope := "test"
				ann := strings.ToLower(pendingAnnotation)
				if strings.Contains(ann, "after") {
					kind = models.FixtureTeardownHook
				}
				if strings.Contains(ann, "all") || strings.Contains(ann, "class") {
					scope = "suite"
				}
				add(models.FixtureSurface{
					FixtureID:     models.BuildFixtureID(relPath, name, scope),
					Name:          name,
					Path:          relPath,
					Kind:          kind,
					Scope:         scope,
					Language:      "java",
					Framework:     "junit5",
					Line:          i + 1,
					Shared:        false,
					DetectionTier: models.TierStructural,
					Confidence:    0.95,
				})
				pendingAnnotation = ""
				continue
			}
			// Non-method line — annotation might apply to a field, reset.
			if trimmed != "" && !strings.HasPrefix(trimmed, "@") && !strings.HasPrefix(trimmed, "//") {
				pendingAnnotation = ""
			}
		}

		// Detect helper methods.
		if m := javaTestHelperPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			fid := models.BuildFixtureID(relPath, name, "")
			if !seen[fid] {
				kind := classifyJavaHelper(name)
				add(models.FixtureSurface{
					FixtureID:     fid,
					Name:          name,
					Path:          relPath,
					Kind:          kind,
					Scope:         "unknown",
					Language:      "java",
					Framework:     "junit5",
					Line:          i + 1,
					Shared:        false,
					DetectionTier: models.TierPattern,
					Confidence:    0.8,
				})
			}
		}
	}

	return fixtures
}

func classifyJavaHelper(name string) models.FixtureKind {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "mock") || strings.Contains(lower, "stub") || strings.Contains(lower, "fake"):
		return models.FixtureMockProvider
	case strings.Contains(lower, "seed") || strings.Contains(lower, "fixture"):
		return models.FixtureDataLoader
	case strings.Contains(lower, "setup"):
		return models.FixtureSetupHook
	default:
		return models.FixtureBuilder
	}
}

// --- Shared file detection ---

// isSharedFixtureFile returns true if the file path suggests it contains
// shared fixtures (conftest, setup files, helpers, factories, etc.).
func isSharedFixtureFile(relPath string) bool {
	base := strings.ToLower(filepath.Base(relPath))
	dir := strings.ToLower(filepath.Dir(relPath))

	// Python conftest files.
	if base == "conftest.py" {
		return true
	}

	// Common shared fixture directories.
	if strings.Contains(dir, "fixture") || strings.Contains(dir, "factory") ||
		strings.Contains(dir, "helper") || strings.Contains(dir, "support") ||
		strings.Contains(dir, "__fixtures__") {
		return true
	}

	// Common shared fixture file names.
	noExt := strings.TrimSuffix(base, filepath.Ext(base))
	sharedNames := map[string]bool{
		"setup": true, "helpers": true, "helper": true,
		"factories": true, "factory": true,
		"fixtures": true, "fixture": true,
		"test-utils": true, "testutils": true, "test_utils": true,
		"test-helpers": true, "testhelpers": true, "test_helpers": true,
		"mocks": true, "stubs": true, "fakes": true,
		"builders": true, "seeds": true,
	}
	return sharedNames[noExt]
}
