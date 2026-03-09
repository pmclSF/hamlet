package analysis

import (
	"os"
	"regexp"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
)

// analyzeTestFileContent reads a test file and populates counts for
// tests, assertions, mocks, and snapshots.
//
// This is a heuristic content analyzer, not a parser. It uses regex
// patterns to estimate counts. Limitations:
//   - Cannot distinguish test helpers from real tests
//   - May double-count in complex expressions
//   - Language-specific patterns are approximate
//
// These heuristics are sufficient for the V3 nucleus detector stage.
func analyzeTestFileContent(tf *models.TestFile, root string) {
	absPath := root + "/" + tf.Path
	content, err := os.ReadFile(absPath)
	if err != nil {
		return
	}
	src := string(content)

	tf.TestCount = countTests(src, tf.Framework)
	tf.AssertionCount = countAssertions(src, tf.Framework)
	tf.MockCount = countMocks(src, tf.Framework)
	tf.SnapshotCount = countSnapshots(src, tf.Framework)
}

// JS/TS patterns
var (
	jsTestPattern     = regexp.MustCompile(`\b(it|test)\s*\(`)
	jsExpectPattern   = regexp.MustCompile(`\bexpect\s*\(`)
	jsAssertPattern   = regexp.MustCompile(`\bassert\s*[\.(]`)
	jsMockPattern     = regexp.MustCompile(`\b(jest\.mock|jest\.fn|jest\.spyOn|vi\.mock|vi\.fn|vi\.spyOn|sinon\.stub|sinon\.mock|sinon\.spy|\.mockImplementation|\.mockReturnValue|\.mockResolvedValue)\b`)
	jsSnapshotPattern = regexp.MustCompile(`\b(toMatchSnapshot|toMatchInlineSnapshot|matchSnapshot)\b`)

	// Go patterns
	goTestPattern   = regexp.MustCompile(`func\s+Test\w+\s*\(`)
	goAssertPattern = regexp.MustCompile(`\b(t\.Error|t\.Errorf|t\.Fatal|t\.Fatalf|assert\.|require\.)\b`)

	// Python patterns
	pyTestPattern   = regexp.MustCompile(`\bdef\s+test_\w+`)
	pyAssertPattern = regexp.MustCompile(`\b(assert\s|self\.assert|pytest\.raises)\b`)
	pyMockPattern   = regexp.MustCompile(`\b(mock\.patch|Mock\(|MagicMock\(|@patch)\b`)

	// Java patterns
	javaTestPattern   = regexp.MustCompile(`@Test\b`)
	javaAssertPattern = regexp.MustCompile(`\b(assert\w+\s*\(|assertThat\s*\()`)
	javaMockPattern   = regexp.MustCompile(`\b(mock\(|when\(|verify\(|@Mock|Mockito\.)`)
)

func countTests(src, framework string) int {
	return getLanguageAnalyzer(frameworkLanguage(framework)).CountTests(src)
}

func countAssertions(src, framework string) int {
	return getLanguageAnalyzer(frameworkLanguage(framework)).CountAssertions(src)
}

func countMocks(src, framework string) int {
	return getLanguageAnalyzer(frameworkLanguage(framework)).CountMocks(src)
}

func countSnapshots(src, framework string) int {
	return getLanguageAnalyzer(frameworkLanguage(framework)).CountSnapshots(src)
}

func frameworkLanguage(framework string) string {
	switch framework {
	case "jest", "vitest", "mocha", "jasmine", "cypress", "playwright", "puppeteer", "webdriverio", "testcafe":
		return "js"
	case "go-testing":
		return "go"
	case "pytest", "unittest", "nose2":
		return "python"
	case "junit4", "junit5", "testng":
		return "java"
	default:
		return "js"
	}
}

// extractExportedCodeUnits performs a lightweight scan of source files
// to find exported/public code units.
//
// This is intentionally simple: it finds exported functions and classes
// via regex patterns. It does not perform full AST analysis.
//
// Limitations:
//   - May miss some export patterns (re-exports, default exports of complex expressions)
//   - Does not resolve type exports
//   - Go detection requires exported identifiers (uppercase first letter)
func extractExportedCodeUnits(root string, testFiles []models.TestFile) []models.CodeUnit {
	// Build set of test file paths to exclude
	testPaths := map[string]bool{}
	for _, tf := range testFiles {
		testPaths[tf.Path] = true
	}

	var units []models.CodeUnit

	// Walk the source tree looking for source files (not test files)
	walkSourceFiles(root, func(relPath string) {
		if testPaths[relPath] {
			return
		}
		ext := strings.ToLower(relPathExt(relPath))
		if lang, ok := languageForExt[ext]; ok {
			if a := getLanguageAnalyzer(lang); a != nil {
				units = append(units, a.ExtractExports(root, relPath)...)
			}
		}
	})

	return units
}

func relPathExt(p string) string {
	i := strings.LastIndex(p, ".")
	if i < 0 {
		return ""
	}
	return p[i:]
}

var (
	// ESM export patterns
	jsExportFuncPattern  = regexp.MustCompile(`export\s+(?:async\s+)?function\s+(\w+)`)
	jsExportClassPattern = regexp.MustCompile(`export\s+class\s+(\w+)`)
	jsExportConstPattern = regexp.MustCompile(`export\s+(?:const|let|var)\s+(\w+)`)

	// CJS export patterns
	cjsNamedExportPattern  = regexp.MustCompile(`(?:module\.)?exports\.(\w+)\s*=`)
	cjsModuleExportPattern = regexp.MustCompile(`module\.exports\s*=\s*(\w+)\s*;?$`)

	goExportFuncPattern = regexp.MustCompile(`func\s+([A-Z]\w*)\s*\(`)

	pyDefPattern = regexp.MustCompile(`^def\s+([a-z]\w+)\s*\(`)
)

func extractJSExports(root, relPath string) []models.CodeUnit {
	content, err := os.ReadFile(root + "/" + relPath)
	if err != nil {
		return nil
	}
	src := string(content)
	lines := strings.Split(src, "\n")
	var units []models.CodeUnit
	seen := map[string]bool{}

	addUnit := func(name string, kind models.CodeUnitKind, line int) {
		if seen[name] {
			return
		}
		seen[name] = true
		units = append(units, models.CodeUnit{
			UnitID:    buildUnitID(relPath, name, ""),
			Name:      name,
			Path:      relPath,
			Kind:      kind,
			Exported:  true,
			Language:  "js",
			StartLine: line,
		})
	}

	// Line-aware matching for ESM exports.
	for i, line := range lines {
		if m := jsExportFuncPattern.FindStringSubmatch(line); m != nil {
			addUnit(m[1], models.CodeUnitKindFunction, i+1)
		}
		if m := jsExportClassPattern.FindStringSubmatch(line); m != nil {
			addUnit(m[1], models.CodeUnitKindClass, i+1)
		}
		if m := jsExportConstPattern.FindStringSubmatch(line); m != nil {
			addUnit(m[1], models.CodeUnitKindFunction, i+1)
		}
		if m := cjsNamedExportPattern.FindStringSubmatch(line); m != nil {
			addUnit(m[1], models.CodeUnitKindFunction, i+1)
		}
		if m := cjsModuleExportPattern.FindStringSubmatch(line); m != nil {
			addUnit(m[1], models.CodeUnitKindFunction, i+1)
		}
	}

	return units
}

// buildUnitID constructs a deterministic code unit ID.
// Format: path:name or path:parent.name for methods.
func buildUnitID(path, name, parent string) string {
	if parent != "" {
		return path + ":" + parent + "." + name
	}
	return path + ":" + name
}

func extractGoExports(root, relPath string) []models.CodeUnit {
	content, err := os.ReadFile(root + "/" + relPath)
	if err != nil {
		return nil
	}
	lines := strings.Split(string(content), "\n")
	var units []models.CodeUnit
	for i, line := range lines {
		if m := goExportFuncPattern.FindStringSubmatch(line); m != nil {
			units = append(units, models.CodeUnit{
				UnitID:    buildUnitID(relPath, m[1], ""),
				Name:      m[1],
				Path:      relPath,
				Kind:      models.CodeUnitKindFunction,
				Exported:  true,
				Language:  "go",
				StartLine: i + 1,
			})
		}
	}
	return units
}

func extractPythonExports(root, relPath string) []models.CodeUnit {
	content, err := os.ReadFile(root + "/" + relPath)
	if err != nil {
		return nil
	}
	var units []models.CodeUnit
	for i, line := range strings.Split(string(content), "\n") {
		if m := pyDefPattern.FindStringSubmatch(line); m != nil {
			if !strings.HasPrefix(m[1], "_") {
				units = append(units, models.CodeUnit{
					UnitID:    buildUnitID(relPath, m[1], ""),
					Name:      m[1],
					Path:      relPath,
					Kind:      models.CodeUnitKindFunction,
					Exported:  true,
					Language:  "python",
					StartLine: i + 1,
				})
			}
		}
	}
	return units
}

func walkSourceFiles(root string, fn func(relPath string)) {
	sourceExts := map[string]bool{
		".js": true, ".jsx": true, ".ts": true, ".tsx": true,
		".mjs": true, ".mts": true, ".go": true, ".py": true,
		".java": true,
	}

	_ = walkDir(root, func(relPath string, isDir bool) bool {
		if isDir {
			return skipDirs[relPathBase(relPath)]
		}
		ext := strings.ToLower(relPathExt(relPath))
		if sourceExts[ext] {
			fn(relPath)
		}
		return false
	})
}

func relPathBase(p string) string {
	i := strings.LastIndex(p, "/")
	if i < 0 {
		return p
	}
	return p[i+1:]
}

// walkDir is a simple recursive directory walker that uses relative paths.
// The callback returns true to skip a directory.
func walkDir(root string, fn func(relPath string, isDir bool) bool) error {
	return walkDirRec(root, "", fn)
}

func walkDirRec(root, rel string, fn func(relPath string, isDir bool) bool) error {
	fullPath := root
	if rel != "" {
		fullPath = root + "/" + rel
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil
	}

	for _, e := range entries {
		childRel := e.Name()
		if rel != "" {
			childRel = rel + "/" + e.Name()
		}

		if e.IsDir() {
			if fn(childRel, true) {
				continue // skip
			}
			_ = walkDirRec(root, childRel, fn)
		} else {
			fn(childRel, false)
		}
	}
	return nil
}
