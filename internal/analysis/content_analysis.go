package analysis

import (
	"os"
	"path/filepath"
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
	analyzeTestFileContentCached(tf, root)
}

// analyzeTestFileContentCached reads a test file, populates analysis counts,
// and returns the file content string for reuse by downstream stages.
func analyzeTestFileContentCached(tf *models.TestFile, root string) string {
	absPath := filepath.Join(root, tf.Path)
	content, err := os.ReadFile(absPath)
	if err != nil {
		return ""
	}
	src := string(content)

	tf.TestCount = countTests(src, tf.Framework)
	tf.AssertionCount = countAssertions(src, tf.Framework)
	tf.MockCount = countMocks(src, tf.Framework)
	tf.SnapshotCount = countSnapshots(src, tf.Framework)
	return src
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

	sourceFiles := collectSourceFiles(root)
	unitsByFile := make([][]models.CodeUnit, len(sourceFiles))
	parallelForEachIndex(len(sourceFiles), func(i int) {
		relPath := sourceFiles[i]
		if testPaths[relPath] {
			return
		}
		ext := strings.ToLower(relPathExt(relPath))
		lang, ok := languageForExt[ext]
		if !ok {
			return
		}
		if a := getLanguageAnalyzer(lang); a != nil {
			unitsByFile[i] = a.ExtractExports(root, relPath)
		}
	})

	units := make([]models.CodeUnit, 0, len(sourceFiles))
	for i := range unitsByFile {
		units = append(units, unitsByFile[i]...)
	}
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
	jsExportFuncPattern         = regexp.MustCompile(`export\s+(?:async\s+)?function\s+(\w+)`)
	jsExportDefaultFuncPattern  = regexp.MustCompile(`export\s+default\s+(?:async\s+)?function\s+(\w+)`)
	jsExportClassPattern        = regexp.MustCompile(`export\s+class\s+(\w+)`)
	jsExportDefaultClassPattern = regexp.MustCompile(`export\s+default\s+class\s+(\w+)`)
	jsExportConstPattern        = regexp.MustCompile(`export\s+(?:const|let|var)\s+(\w+)`)
	jsNamedExportListPattern    = regexp.MustCompile(`export\s*\{\s*([^}]*)\s*\}(?:\s*from\s*['"][^'"]+['"])?`)

	// CJS export patterns
	cjsNamedExportPattern  = regexp.MustCompile(`(?:module\.)?exports\.(\w+)\s*=`)
	cjsModuleExportPattern = regexp.MustCompile(`module\.exports\s*=\s*(\w+)\s*;?$`)

	goExportFuncPattern     = regexp.MustCompile(`^\s*func\s+([A-Z]\w*)\s*\(`)
	goExportMethodPattern   = regexp.MustCompile(`^\s*func\s+\(\s*[^)]*\*?\s*([A-Z]\w*)\s*\)\s*([A-Z]\w*)\s*\(`)
	goExportTypePattern     = regexp.MustCompile(`^\s*type\s+([A-Z]\w*)\s+`)
	goExportConstVarPattern = regexp.MustCompile(`^\s*(?:const|var)\s+([A-Z]\w*)\b`)
	javaExportTypePattern   = regexp.MustCompile(`\bpublic\s+(?:abstract\s+|final\s+)?(?:class|interface|enum)\s+(\w+)`)
	javaExportMethodPattern = regexp.MustCompile(`\bpublic\s+(?:static\s+)?(?:[\w\[\]<>?,]+\s+)+(\w+)\s*\(`)

	pyDefPattern = regexp.MustCompile(`^def\s+([a-z]\w+)\s*\(`)
	pyAllPattern = regexp.MustCompile(`(?s)__all__\s*=\s*[\[\(]([^\]\)]*)[\]\)]`)
	pyAllItem    = regexp.MustCompile(`['"]([A-Za-z_]\w*)['"]`)
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
		if m := jsExportDefaultFuncPattern.FindStringSubmatch(line); m != nil {
			addUnit(m[1], models.CodeUnitKindFunction, i+1)
		}
		if m := jsExportClassPattern.FindStringSubmatch(line); m != nil {
			addUnit(m[1], models.CodeUnitKindClass, i+1)
		}
		if m := jsExportDefaultClassPattern.FindStringSubmatch(line); m != nil {
			addUnit(m[1], models.CodeUnitKindClass, i+1)
		}
		if m := jsExportConstPattern.FindStringSubmatch(line); m != nil {
			addUnit(m[1], models.CodeUnitKindFunction, i+1)
		}
		if m := jsNamedExportListPattern.FindStringSubmatch(line); m != nil {
			for _, item := range strings.Split(m[1], ",") {
				exportName := parseJSNamedExport(item)
				if exportName == "" {
					continue
				}
				addUnit(exportName, models.CodeUnitKindFunction, i+1)
			}
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

func parseJSNamedExport(raw string) string {
	item := strings.TrimSpace(raw)
	if item == "" {
		return ""
	}
	parts := strings.Split(item, " as ")
	if len(parts) == 1 {
		return strings.TrimSpace(parts[0])
	}
	alias := strings.TrimSpace(parts[len(parts)-1])
	return alias
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
	seen := map[string]bool{}
	add := func(name string, kind models.CodeUnitKind, line int, parent string) {
		unitID := buildUnitID(relPath, name, parent)
		if seen[unitID] {
			return
		}
		seen[unitID] = true
		units = append(units, models.CodeUnit{
			UnitID:     unitID,
			Name:       name,
			Path:       relPath,
			Kind:       kind,
			Exported:   true,
			Language:   "go",
			ParentName: parent,
			StartLine:  line,
		})
	}

	for i, line := range lines {
		if m := goExportMethodPattern.FindStringSubmatch(line); m != nil {
			add(m[2], models.CodeUnitKindMethod, i+1, m[1])
		}
		if m := goExportFuncPattern.FindStringSubmatch(line); m != nil {
			add(m[1], models.CodeUnitKindFunction, i+1, "")
		}
		if m := goExportTypePattern.FindStringSubmatch(line); m != nil {
			add(m[1], models.CodeUnitKindClass, i+1, "")
		}
		if m := goExportConstVarPattern.FindStringSubmatch(line); m != nil {
			add(m[1], models.CodeUnitKindUnknown, i+1, "")
		}
	}
	return units
}

func extractPythonExports(root, relPath string) []models.CodeUnit {
	content, err := os.ReadFile(root + "/" + relPath)
	if err != nil {
		return nil
	}
	src := string(content)
	allowed := pythonAllExports(src)
	var units []models.CodeUnit
	for i, line := range strings.Split(src, "\n") {
		if m := pyDefPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			if strings.HasPrefix(name, "_") {
				continue
			}
			if len(allowed) > 0 && !allowed[name] {
				continue
			}
			units = append(units, models.CodeUnit{
				UnitID:    buildUnitID(relPath, name, ""),
				Name:      name,
				Path:      relPath,
				Kind:      models.CodeUnitKindFunction,
				Exported:  true,
				Language:  "python",
				StartLine: i + 1,
			})
		}
	}
	return units
}

func pythonAllExports(src string) map[string]bool {
	m := pyAllPattern.FindStringSubmatch(src)
	if len(m) < 2 {
		return nil
	}
	items := pyAllItem.FindAllStringSubmatch(m[1], -1)
	if len(items) == 0 {
		return nil
	}
	out := map[string]bool{}
	for _, item := range items {
		if len(item) >= 2 && item[1] != "" {
			out[item[1]] = true
		}
	}
	return out
}

func extractJavaExports(root, relPath string) []models.CodeUnit {
	content, err := os.ReadFile(root + "/" + relPath)
	if err != nil {
		return nil
	}
	lines := strings.Split(string(content), "\n")
	var units []models.CodeUnit
	seen := map[string]bool{}
	typeNames := map[string]bool{}
	add := func(name string, kind models.CodeUnitKind, line int, parent string) {
		if name == "" {
			return
		}
		unitID := buildUnitID(relPath, name, parent)
		if seen[unitID] {
			return
		}
		seen[unitID] = true
		units = append(units, models.CodeUnit{
			UnitID:     unitID,
			Name:       name,
			Path:       relPath,
			Kind:       kind,
			Exported:   true,
			Language:   "java",
			ParentName: parent,
			StartLine:  line,
		})
	}

	currentType := ""
	braceDepth := 0
	typeDepth := 0
	inBlockComment := false
	for i, line := range lines {
		if m := javaExportTypePattern.FindStringSubmatch(line); m != nil {
			currentType = m[1]
			typeDepth = braceDepth
			typeNames[currentType] = true
			add(currentType, models.CodeUnitKindClass, i+1, "")
		}

		if m := javaExportMethodPattern.FindStringSubmatch(line); m != nil {
			methodName := m[1]
			parent := currentType
			if typeNames[methodName] {
				// Constructor already represented by its type.
				parent = ""
			}
			if parent != "" && methodName != parent {
				add(methodName, models.CodeUnitKindMethod, i+1, parent)
			} else if parent == "" && !typeNames[methodName] {
				add(methodName, models.CodeUnitKindMethod, i+1, "")
			}
		}

		sanitized, nextBlockComment := stripJavaForBraceCounting(line, inBlockComment)
		inBlockComment = nextBlockComment
		braceDepth += strings.Count(sanitized, "{")
		braceDepth -= strings.Count(sanitized, "}")
		if currentType != "" && braceDepth <= typeDepth {
			currentType = ""
		}
	}
	return units
}

func stripJavaForBraceCounting(line string, inBlockComment bool) (string, bool) {
	var out strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	for i := 0; i < len(line); i++ {
		ch := line[i]
		next := byte(0)
		if i+1 < len(line) {
			next = line[i+1]
		}

		if inBlockComment {
			if ch == '*' && next == '/' {
				inBlockComment = false
				i++
			}
			continue
		}

		if inSingle {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '\'' {
				inSingle = false
			}
			continue
		}

		if inDouble {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inDouble = false
			}
			continue
		}

		if ch == '/' && next == '*' {
			inBlockComment = true
			i++
			continue
		}
		if ch == '/' && next == '/' {
			break
		}
		if ch == '\'' {
			inSingle = true
			escaped = false
			continue
		}
		if ch == '"' {
			inDouble = true
			escaped = false
			continue
		}

		out.WriteByte(ch)
	}

	return out.String(), inBlockComment
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

func collectSourceFiles(root string) []string {
	files := make([]string, 0, 128)
	walkSourceFiles(root, func(relPath string) {
		files = append(files, relPath)
	})
	return files
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
		fullPath = filepath.Join(root, rel)
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
		if e.Type()&os.ModeSymlink != 0 {
			// Skip symlinks to avoid filesystem cycles.
			continue
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
