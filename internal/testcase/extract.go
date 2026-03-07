package testcase

import (
	"os"
	"regexp"
	"strings"

	"github.com/pmclSF/hamlet/internal/identity"
)

// Extract discovers individual test cases from a file and assigns stable IDs.
func Extract(root, relPath, framework string) []TestCase {
	absPath := root + "/" + relPath
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil
	}

	lang := FrameworkLanguage(framework)
	src := string(content)

	var cases []TestCase
	switch lang {
	case "js":
		cases = extractJS(src, relPath, framework)
	case "go":
		cases = extractGo(src, relPath, framework)
	case "python":
		cases = extractPython(src, relPath, framework)
	case "java":
		cases = extractJava(src, relPath, framework)
	}

	// Assign stable IDs to all extracted cases.
	for i := range cases {
		tc := &cases[i]
		tc.FilePath = identity.NormalizePath(relPath)
		tc.Framework = framework
		tc.Language = lang
		tc.CanonicalIdentity = identity.BuildCanonical(
			tc.FilePath,
			tc.SuiteHierarchy,
			tc.TestName,
			paramSignature(tc),
		)
		tc.TestID = identity.GenerateID(tc.CanonicalIdentity)
	}

	return cases
}

func paramSignature(tc *TestCase) string {
	if tc.Parameterized != nil && tc.Parameterized.ParamSignature != "" {
		return tc.Parameterized.ParamSignature
	}
	return ""
}

// --- JS/TS extraction ---

// jsDescribePattern matches describe('name', ...) or describe("name", ...)
// and describe(`name`, ...).
var jsDescribePattern = regexp.MustCompile(
	`\b(?:describe|context|suite)\s*\(\s*` + jsStringLiteral,
)

// jsTestPattern matches it('name', ...) or test('name', ...).
var jsTestPattern = regexp.MustCompile(
	`\b(it|test)\s*\(\s*` + jsStringLiteral,
)

// jsTestEachPattern matches it.each or test.each or describe.each.
var jsTestEachPattern = regexp.MustCompile(
	`\b(it|test|describe)\.each\s*[\(\[]`,
)

// jsStringLiteral captures a single-quoted, double-quoted, or backtick string.
const jsStringLiteral = `(?:` +
	`'([^']*)'` + `|` + // single-quoted
	`"([^"]*)"` + `|` + // double-quoted
	"`([^`]*)`" + // backtick
	`)`

func extractJSStringMatch(m []string) string {
	// Capture groups for single/double/backtick are at offsets depending on pattern.
	// For jsTestPattern: groups are [full, keyword, sq, dq, bt]
	// For jsDescribePattern: groups are [full, sq, dq, bt]
	for i := len(m) - 3; i < len(m); i++ {
		if i >= 0 && m[i] != "" {
			return m[i]
		}
	}
	return ""
}

// scopeEntry tracks a describe/suite scope for JS extraction.
type scopeEntry struct {
	name       string
	braceDepth int
}

func extractJS(src, relPath, framework string) []TestCase {
	lines := strings.Split(src, "\n")
	var cases []TestCase

	// Track describe nesting via brace counting.
	var suiteStack []scopeEntry
	braceDepth := 0

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Count braces for scope tracking.
		// This is approximate but handles the common case.
		opens := strings.Count(line, "{")
		closes := strings.Count(line, "}")

		// Check for describe/context/suite.
		if dm := jsDescribePattern.FindStringSubmatch(trimmed); dm != nil {
			name := extractJSStringMatch(dm)
			suiteStack = append(suiteStack, scopeEntry{
				name:       name,
				braceDepth: braceDepth,
			})
		}

		// Check for test.each / it.each / describe.each (parameterized).
		if em := jsTestEachPattern.FindStringSubmatch(trimmed); em != nil {
			// Look for the test name on this line or the next few lines.
			paramName := extractEachTestName(lines, lineNum)
			if paramName != "" {
				kind := em[1]
				if kind == "describe" {
					// Parameterized describe — track as suite.
					suiteStack = append(suiteStack, scopeEntry{
						name:       paramName,
						braceDepth: braceDepth,
					})
				} else {
					tc := TestCase{
						TestName:       paramName,
						SuiteHierarchy: suiteNames(suiteStack),
						Line:           lineNum + 1,
						ExtractionKind: ExtractionParameterizedTemplate,
						Confidence:     0.7,
						Parameterized: &ParameterizationInfo{
							IsTemplate: true,
						},
					}
					cases = append(cases, tc)
				}
			}
		} else if tm := jsTestPattern.FindStringSubmatch(trimmed); tm != nil {
			// Regular it/test.
			name := extractJSStringMatch(tm)
			if name != "" {
				tc := TestCase{
					TestName:       name,
					SuiteHierarchy: suiteNames(suiteStack),
					Line:           lineNum + 1,
					ExtractionKind: ExtractionStatic,
					Confidence:     0.9,
				}
				cases = append(cases, tc)
			}
		}

		// Update brace depth.
		braceDepth += opens - closes

		// Pop suite stack when scope closes.
		for len(suiteStack) > 0 && braceDepth <= suiteStack[len(suiteStack)-1].braceDepth {
			suiteStack = suiteStack[:len(suiteStack)-1]
		}
	}

	return cases
}

// extractEachTestName tries to find the test name after a .each() call.
// Pattern: .each(...)('name', ...) or .each(...)("name", ...).
var eachNamePattern = regexp.MustCompile(
	`\)\s*\(\s*` + jsStringLiteral,
)

func extractEachTestName(lines []string, startLine int) string {
	// Search current line and next 5 lines for the test name.
	end := startLine + 6
	if end > len(lines) {
		end = len(lines)
	}
	combined := strings.Join(lines[startLine:end], " ")
	if m := eachNamePattern.FindStringSubmatch(combined); m != nil {
		return extractJSStringMatch(m)
	}
	return ""
}

func suiteNames(stack []scopeEntry) []string {
	if len(stack) == 0 {
		return nil
	}
	names := make([]string, len(stack))
	for i, s := range stack {
		names[i] = s.name
	}
	return names
}

// --- Go extraction ---

var goTestFuncPattern = regexp.MustCompile(`^func\s+(Test\w+)\s*\(`)
var goSubtestPattern = regexp.MustCompile(`\bt\.Run\s*\(\s*(?:"([^"]*)"` + "|`([^`]*)`" + `)`)

func extractGo(src, relPath, framework string) []TestCase {
	lines := strings.Split(src, "\n")
	var cases []TestCase
	var currentFunc string

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Top-level test function.
		if m := goTestFuncPattern.FindStringSubmatch(trimmed); m != nil {
			currentFunc = m[1]
			cases = append(cases, TestCase{
				TestName:       currentFunc,
				Line:           lineNum + 1,
				ExtractionKind: ExtractionStatic,
				Confidence:     0.95,
			})
			continue
		}

		// Subtest: t.Run("name", ...)
		if m := goSubtestPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			if name == "" {
				name = m[2]
			}
			var hierarchy []string
			if currentFunc != "" {
				hierarchy = []string{currentFunc}
			}
			cases = append(cases, TestCase{
				TestName:       name,
				SuiteHierarchy: hierarchy,
				Line:           lineNum + 1,
				ExtractionKind: ExtractionStatic,
				Confidence:     0.9,
			})
		}
	}

	return cases
}

// --- Python extraction ---

var pyClassPattern = regexp.MustCompile(`^class\s+(Test\w+)`)
var pyTestDefPattern = regexp.MustCompile(`^\s+def\s+(test_\w+)\s*\(`)
var pyTopTestDefPattern = regexp.MustCompile(`^def\s+(test_\w+)\s*\(`)
var pyParametrizePattern = regexp.MustCompile(`@pytest\.mark\.parametrize`)

func extractPython(src, relPath, framework string) []TestCase {
	lines := strings.Split(src, "\n")
	var cases []TestCase
	var currentClass string
	pendingParametrize := false

	for lineNum, line := range lines {
		// Track class scope (simple indentation-based).
		if m := pyClassPattern.FindStringSubmatch(line); m != nil {
			currentClass = m[1]
			continue
		}

		// Detect parametrize decorator.
		if pyParametrizePattern.MatchString(line) {
			pendingParametrize = true
			continue
		}

		// Top-level test function outside class.
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			if line != "" && !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "class ") {
				currentClass = "" // Reset class scope at top-level non-blank line.
			}
		}

		// Method-level test.
		if m := pyTestDefPattern.FindStringSubmatch(line); m != nil {
			kind := ExtractionStatic
			confidence := 0.9
			var param *ParameterizationInfo
			if pendingParametrize {
				kind = ExtractionParameterizedTemplate
				confidence = 0.7
				param = &ParameterizationInfo{IsTemplate: true}
			}
			pendingParametrize = false

			var hierarchy []string
			if currentClass != "" {
				hierarchy = []string{currentClass}
			}
			cases = append(cases, TestCase{
				TestName:       m[1],
				SuiteHierarchy: hierarchy,
				Line:           lineNum + 1,
				ExtractionKind: kind,
				Confidence:     confidence,
				Parameterized:  param,
			})
			continue
		}

		// Top-level test function (pytest-style).
		if m := pyTopTestDefPattern.FindStringSubmatch(line); m != nil {
			kind := ExtractionStatic
			confidence := 0.9
			var param *ParameterizationInfo
			if pendingParametrize {
				kind = ExtractionParameterizedTemplate
				confidence = 0.7
				param = &ParameterizationInfo{IsTemplate: true}
			}
			pendingParametrize = false

			cases = append(cases, TestCase{
				TestName:       m[1],
				Line:           lineNum + 1,
				ExtractionKind: kind,
				Confidence:     confidence,
				Parameterized:  param,
			})
			continue
		}

		// Reset parametrize flag if line is not a decorator or blank.
		if pendingParametrize && strings.TrimSpace(line) != "" && !strings.HasPrefix(strings.TrimSpace(line), "@") {
			pendingParametrize = false
		}
	}

	return cases
}

// --- Java extraction ---

var javaClassPattern = regexp.MustCompile(`\bclass\s+(\w+)`)
var javaTestAnnotation = regexp.MustCompile(`@Test\b`)
var javaMethodPattern = regexp.MustCompile(`(?:public|protected|private)?\s*(?:static\s+)?void\s+(\w+)\s*\(`)
var javaParameterizedPattern = regexp.MustCompile(`@ParameterizedTest`)

func extractJava(src, relPath, framework string) []TestCase {
	lines := strings.Split(src, "\n")
	var cases []TestCase
	var currentClass string
	pendingTest := false
	pendingParameterized := false

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		if m := javaClassPattern.FindStringSubmatch(trimmed); m != nil {
			currentClass = m[1]
		}

		if javaTestAnnotation.MatchString(trimmed) {
			pendingTest = true
			continue
		}

		if javaParameterizedPattern.MatchString(trimmed) {
			pendingParameterized = true
			continue
		}

		if pendingTest || pendingParameterized {
			if m := javaMethodPattern.FindStringSubmatch(trimmed); m != nil {
				kind := ExtractionStatic
				confidence := 0.9
				var param *ParameterizationInfo
				if pendingParameterized {
					kind = ExtractionParameterizedTemplate
					confidence = 0.7
					param = &ParameterizationInfo{IsTemplate: true}
				}

				var hierarchy []string
				if currentClass != "" {
					hierarchy = []string{currentClass}
				}
				cases = append(cases, TestCase{
					TestName:       m[1],
					SuiteHierarchy: hierarchy,
					Line:           lineNum + 1,
					ExtractionKind: kind,
					Confidence:     confidence,
					Parameterized:  param,
				})
				pendingTest = false
				pendingParameterized = false
				continue
			}
		}
	}

	return cases
}

// FrameworkLanguage maps a framework name to its language identifier.
func FrameworkLanguage(framework string) string {
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
