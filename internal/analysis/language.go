package analysis

import (
	"regexp"

	"github.com/pmclSF/hamlet/internal/models"
)

// LanguageAnalyzer defines the interface for language-specific content analysis.
// Each language provides its own set of patterns for counting tests, assertions,
// mocks, and snapshots, as well as for extracting exported code units.
type LanguageAnalyzer interface {
	// Language returns the canonical language identifier (e.g., "js", "go", "python", "java").
	Language() string

	// CountTests returns the estimated number of test definitions in src.
	CountTests(src string) int

	// CountAssertions returns the estimated number of assertions in src.
	CountAssertions(src string) int

	// CountMocks returns the estimated number of mock usages in src.
	CountMocks(src string) int

	// CountSnapshots returns the estimated number of snapshot assertions in src.
	CountSnapshots(src string) int

	// ExtractExports extracts exported code units from a source file.
	ExtractExports(root, relPath string) []models.CodeUnit
}

// languageRegistry maps language identifiers to their analyzers.
var languageRegistry = map[string]LanguageAnalyzer{}

func init() {
	registerLanguage(&jsAnalyzer{})
	registerLanguage(&goAnalyzer{})
	registerLanguage(&pythonAnalyzer{})
	registerLanguage(&javaAnalyzer{})
}

func registerLanguage(a LanguageAnalyzer) {
	languageRegistry[a.Language()] = a
}

func getLanguageAnalyzer(lang string) LanguageAnalyzer {
	if a, ok := languageRegistry[lang]; ok {
		return a
	}
	return languageRegistry["js"] // default fallback
}

// jsAnalyzer implements LanguageAnalyzer for JavaScript/TypeScript.
type jsAnalyzer struct{}

func (a *jsAnalyzer) Language() string { return "js" }

func (a *jsAnalyzer) CountTests(src string) int {
	return len(jsTestPattern.FindAllString(src, -1))
}

func (a *jsAnalyzer) CountAssertions(src string) int {
	return len(jsExpectPattern.FindAllString(src, -1)) +
		len(jsAssertPattern.FindAllString(src, -1))
}

func (a *jsAnalyzer) CountMocks(src string) int {
	return len(jsMockPattern.FindAllString(src, -1))
}

func (a *jsAnalyzer) CountSnapshots(src string) int {
	return len(jsSnapshotPattern.FindAllString(src, -1))
}

func (a *jsAnalyzer) ExtractExports(root, relPath string) []models.CodeUnit {
	return extractJSExports(root, relPath)
}

// goAnalyzer implements LanguageAnalyzer for Go.
type goAnalyzer struct{}

func (a *goAnalyzer) Language() string { return "go" }

func (a *goAnalyzer) CountTests(src string) int {
	return len(goTestPattern.FindAllString(src, -1))
}

func (a *goAnalyzer) CountAssertions(src string) int {
	return len(goAssertPattern.FindAllString(src, -1))
}

func (a *goAnalyzer) CountMocks(src string) int { return 0 }

func (a *goAnalyzer) CountSnapshots(src string) int { return 0 }

func (a *goAnalyzer) ExtractExports(root, relPath string) []models.CodeUnit {
	return extractGoExports(root, relPath)
}

// pythonAnalyzer implements LanguageAnalyzer for Python.
type pythonAnalyzer struct{}

func (a *pythonAnalyzer) Language() string { return "python" }

func (a *pythonAnalyzer) CountTests(src string) int {
	return len(pyTestPattern.FindAllString(src, -1))
}

func (a *pythonAnalyzer) CountAssertions(src string) int {
	return len(pyAssertPattern.FindAllString(src, -1))
}

func (a *pythonAnalyzer) CountMocks(src string) int {
	return len(pyMockPattern.FindAllString(src, -1))
}

func (a *pythonAnalyzer) CountSnapshots(src string) int { return 0 }

func (a *pythonAnalyzer) ExtractExports(root, relPath string) []models.CodeUnit {
	return extractPythonExports(root, relPath)
}

// javaAnalyzer implements LanguageAnalyzer for Java.
type javaAnalyzer struct{}

func (a *javaAnalyzer) Language() string { return "java" }

func (a *javaAnalyzer) CountTests(src string) int {
	return len(javaTestPattern.FindAllString(src, -1))
}

func (a *javaAnalyzer) CountAssertions(src string) int {
	return len(javaAssertPattern.FindAllString(src, -1))
}

func (a *javaAnalyzer) CountMocks(src string) int {
	return len(javaMockPattern.FindAllString(src, -1))
}

func (a *javaAnalyzer) CountSnapshots(src string) int { return 0 }

func (a *javaAnalyzer) ExtractExports(root, relPath string) []models.CodeUnit {
	// Java export extraction not yet implemented.
	return nil
}

// languageForExt maps file extensions to language analyzer keys.
var languageForExt = map[string]string{
	".js": "js", ".jsx": "js", ".ts": "js", ".tsx": "js",
	".mjs": "js", ".mts": "js", ".cjs": "js", ".cts": "js",
	".go":   "go",
	".py":   "python",
	".java": "java",
}

// Java export pattern (unused currently but reserved for future).
var _ = regexp.MustCompile(`public\s+(?:static\s+)?(?:class|interface)\s+(\w+)`)
