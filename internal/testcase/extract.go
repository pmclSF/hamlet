package testcase

import (
	"os"
	"regexp"
	"strconv"
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
	`'((?:\\.|[^'\\])*)'` + `|` + // single-quoted (supports escapes)
	`"((?:\\.|[^"\\])*)"` + `|` + // double-quoted (supports escapes)
	"`((?:\\\\.|[^`\\\\])*)`" + // backtick (supports escaped backticks)
	`)`

func extractJSStringMatch(m []string) string {
	// Capture groups for single/double/backtick are the last three entries.
	if len(m) < 4 {
		return ""
	}
	start := len(m) - 3
	if m[start] != "" {
		return decodeQuotedString(m[start], '\'')
	}
	if m[start+1] != "" {
		return decodeQuotedString(m[start+1], '"')
	}
	if m[start+2] != "" {
		return decodeTemplateLiteral(m[start+2])
	}
	return ""
}

func decodeQuotedString(raw string, quote byte) string {
	if quote == '\'' {
		decoded := strings.ReplaceAll(raw, `\\`, `\`)
		decoded = strings.ReplaceAll(decoded, `\'`, `'`)
		return decoded
	}
	quoted := string(quote) + raw + string(quote)
	decoded, err := strconv.Unquote(quoted)
	if err != nil {
		return raw
	}
	return decoded
}

func decodeTemplateLiteral(raw string) string {
	// Best-effort template literal decoding for escaped backticks and slashes.
	raw = strings.ReplaceAll(raw, "\\`", "`")
	raw = strings.ReplaceAll(raw, "\\\\", "\\")
	return raw
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
	inBlockComment := false

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Count braces for scope tracking while ignoring string literals and comments.
		sanitized, nextBlockComment := stripJSForBraceCounting(line, inBlockComment)
		opens := strings.Count(sanitized, "{")
		closes := strings.Count(sanitized, "}")
		inBlockComment = nextBlockComment

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
			paramName, estimatedInstances := extractEachTestInfo(lines, lineNum)
			if paramName != "" {
				kind := em[1]
				if kind == "describe" {
					// Parameterized describe — track as suite.
					suiteStack = append(suiteStack, scopeEntry{
						name:       paramName,
						braceDepth: braceDepth,
					})
				} else {
					// Enumerate concrete instances when table cardinality can be
					// determined statically; otherwise keep template-level fallback.
					if estimatedInstances > 0 {
						const maxEnumeratedInstances = 100
						if estimatedInstances > maxEnumeratedInstances {
							estimatedInstances = maxEnumeratedInstances
						}
						for i := 1; i <= estimatedInstances; i++ {
							tc := TestCase{
								TestName:       paramName,
								SuiteHierarchy: suiteNames(suiteStack),
								Line:           lineNum + 1,
								ExtractionKind: ExtractionStatic,
								Confidence:     0.8,
								Parameterized: &ParameterizationInfo{
									IsTemplate:         false,
									ParamSignature:     "case_" + strconv.Itoa(i),
									EstimatedInstances: estimatedInstances,
								},
							}
							cases = append(cases, tc)
						}
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

func stripJSForBraceCounting(line string, inBlockComment bool) (string, bool) {
	var out strings.Builder
	inSingle := false
	inDouble := false
	inTemplate := false
	templateExprDepth := 0
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

		if inTemplate {
			if templateExprDepth == 0 {
				if escaped {
					escaped = false
					continue
				}
				if ch == '\\' {
					escaped = true
					continue
				}
				if ch == '`' {
					inTemplate = false
					continue
				}
				if ch == '$' && next == '{' {
					templateExprDepth = 1
					out.WriteByte('{')
					i++
					continue
				}
				continue
			}

			// Inside `${...}` expression: keep brace accounting active.
			if ch == '{' {
				templateExprDepth++
				out.WriteByte(ch)
				continue
			}
			if ch == '}' {
				templateExprDepth--
				out.WriteByte(ch)
				if templateExprDepth == 0 {
					escaped = false
				}
				continue
			}
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
		if ch == '`' {
			inTemplate = true
			escaped = false
			continue
		}
		out.WriteByte(ch)
	}

	return out.String(), inBlockComment
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

func extractEachTestInfo(lines []string, startLine int) (name string, estimatedInstances int) {
	name = extractEachTestName(lines, startLine)
	if name == "" {
		return "", 0
	}
	return name, estimateEachInstances(lines, startLine)
}

func estimateEachInstances(lines []string, startLine int) int {
	end := startLine + 12
	if end > len(lines) {
		end = len(lines)
	}
	combined := strings.Join(lines[startLine:end], "\n")
	args := extractEachArgs(combined)
	if args == "" {
		return 0
	}
	firstArg := firstTopLevelArg(args)
	if firstArg == "" {
		return 0
	}
	return estimateArrayLiteralElements(firstArg)
}

func extractEachArgs(src string) string {
	const marker = ".each("
	idx := strings.Index(src, marker)
	if idx < 0 {
		return ""
	}
	input := src[idx+len(marker):]

	depth := 1
	inSingle := false
	inDouble := false
	inTemplate := false
	escaped := false

	for i := 0; i < len(input); i++ {
		ch := input[i]

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
		if inTemplate {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '`' {
				inTemplate = false
			}
			continue
		}

		switch ch {
		case '\'':
			inSingle = true
		case '"':
			inDouble = true
		case '`':
			inTemplate = true
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return strings.TrimSpace(input[:i])
			}
		}
	}

	return ""
}

func firstTopLevelArg(args string) string {
	inSingle := false
	inDouble := false
	inTemplate := false
	escaped := false
	depthParen, depthBrace, depthBracket := 0, 0, 0

	for i := 0; i < len(args); i++ {
		ch := args[i]
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
		if inTemplate {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '`' {
				inTemplate = false
			}
			continue
		}

		switch ch {
		case '\'':
			inSingle = true
		case '"':
			inDouble = true
		case '`':
			inTemplate = true
		case '(':
			depthParen++
		case ')':
			if depthParen > 0 {
				depthParen--
			}
		case '{':
			depthBrace++
		case '}':
			if depthBrace > 0 {
				depthBrace--
			}
		case '[':
			depthBracket++
		case ']':
			if depthBracket > 0 {
				depthBracket--
			}
		case ',':
			if depthParen == 0 && depthBrace == 0 && depthBracket == 0 {
				return strings.TrimSpace(args[:i])
			}
		}
	}
	return strings.TrimSpace(args)
}

func estimateArrayLiteralElements(arg string) int {
	trimmed := strings.TrimSpace(arg)
	if len(trimmed) < 2 || trimmed[0] != '[' {
		return 0
	}

	inSingle := false
	inDouble := false
	inTemplate := false
	escaped := false
	depthBracket, depthParen, depthBrace := 0, 0, 0
	elements := 0
	hasValue := false

	for i := 0; i < len(trimmed); i++ {
		ch := trimmed[i]

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
			if depthBracket == 1 {
				hasValue = true
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
			if depthBracket == 1 {
				hasValue = true
			}
			continue
		}
		if inTemplate {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '`' {
				inTemplate = false
			}
			if depthBracket == 1 {
				hasValue = true
			}
			continue
		}

		switch ch {
		case '\'':
			inSingle = true
			if depthBracket == 1 {
				hasValue = true
			}
		case '"':
			inDouble = true
			if depthBracket == 1 {
				hasValue = true
			}
		case '`':
			inTemplate = true
			if depthBracket == 1 {
				hasValue = true
			}
		case '[':
			depthBracket++
			if depthBracket > 1 && depthBracket == 2 {
				hasValue = true
			}
		case ']':
			if depthBracket == 1 {
				if hasValue {
					elements++
				}
				return elements
			}
			if depthBracket > 0 {
				depthBracket--
			}
		case '(':
			if depthBracket >= 1 {
				depthParen++
				hasValue = true
			}
		case ')':
			if depthParen > 0 {
				depthParen--
			}
		case '{':
			if depthBracket >= 1 {
				depthBrace++
				hasValue = true
			}
		case '}':
			if depthBrace > 0 {
				depthBrace--
			}
		case ',':
			if depthBracket == 1 && depthParen == 0 && depthBrace == 0 {
				if hasValue {
					elements++
				}
				hasValue = false
			}
		default:
			if depthBracket == 1 && !isWhitespaceByte(ch) {
				hasValue = true
			}
		}
	}

	return 0
}

func isWhitespaceByte(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
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
var goSubtestPattern = regexp.MustCompile(`\bt\.Run\s*\(\s*(?:"((?:\\.|[^"\\])*)"` + "|`([^`]*)`" + `)`)

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
			} else {
				name = decodeQuotedString(name, '"')
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
	pendingParametrizeInstances := 0

	for lineNum, line := range lines {
		// Track class scope (simple indentation-based).
		if m := pyClassPattern.FindStringSubmatch(line); m != nil {
			currentClass = m[1]
			continue
		}

		// Detect parametrize decorator.
		if pyParametrizePattern.MatchString(line) {
			pendingParametrize = true
			pendingParametrizeInstances = estimatePythonParametrizeInstances(lines, lineNum)
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
			var params []*ParameterizationInfo
			if pendingParametrize {
				kind = ExtractionParameterizedTemplate
				confidence = 0.7
				if pendingParametrizeInstances > 0 {
					kind = ExtractionStatic
					confidence = 0.8
					for i := 1; i <= pendingParametrizeInstances; i++ {
						params = append(params, &ParameterizationInfo{
							IsTemplate:         false,
							ParamSignature:     "case_" + strconv.Itoa(i),
							EstimatedInstances: pendingParametrizeInstances,
						})
					}
				} else {
					params = append(params, &ParameterizationInfo{IsTemplate: true})
				}
			}
			pendingParametrize = false
			pendingParametrizeInstances = 0

			var hierarchy []string
			if currentClass != "" {
				hierarchy = []string{currentClass}
			}
			if len(params) == 0 {
				params = []*ParameterizationInfo{nil}
			}
			for _, param := range params {
				cases = append(cases, TestCase{
					TestName:       m[1],
					SuiteHierarchy: hierarchy,
					Line:           lineNum + 1,
					ExtractionKind: kind,
					Confidence:     confidence,
					Parameterized:  param,
				})
			}
			continue
		}

		// Top-level test function (pytest-style).
		if m := pyTopTestDefPattern.FindStringSubmatch(line); m != nil {
			kind := ExtractionStatic
			confidence := 0.9
			var params []*ParameterizationInfo
			if pendingParametrize {
				kind = ExtractionParameterizedTemplate
				confidence = 0.7
				if pendingParametrizeInstances > 0 {
					kind = ExtractionStatic
					confidence = 0.8
					for i := 1; i <= pendingParametrizeInstances; i++ {
						params = append(params, &ParameterizationInfo{
							IsTemplate:         false,
							ParamSignature:     "case_" + strconv.Itoa(i),
							EstimatedInstances: pendingParametrizeInstances,
						})
					}
				} else {
					params = append(params, &ParameterizationInfo{IsTemplate: true})
				}
			}
			pendingParametrize = false
			pendingParametrizeInstances = 0

			if len(params) == 0 {
				params = []*ParameterizationInfo{nil}
			}
			for _, param := range params {
				cases = append(cases, TestCase{
					TestName:       m[1],
					Line:           lineNum + 1,
					ExtractionKind: kind,
					Confidence:     confidence,
					Parameterized:  param,
				})
			}
			continue
		}

		// Reset parametrize flag if line is not a decorator or blank.
		if pendingParametrize && strings.TrimSpace(line) != "" && !strings.HasPrefix(strings.TrimSpace(line), "@") {
			pendingParametrize = false
			pendingParametrizeInstances = 0
		}
	}

	return cases
}

func estimatePythonParametrizeInstances(lines []string, startLine int) int {
	end := startLine + 8
	if end > len(lines) {
		end = len(lines)
	}
	combined := strings.Join(lines[startLine:end], " ")
	args := extractCallArgs(combined, "parametrize")
	if args == "" {
		return 0
	}
	valuesArg := secondTopLevelArg(args)
	if valuesArg == "" {
		return 0
	}
	return estimateSequenceLiteralElements(valuesArg)
}

func extractCallArgs(src, marker string) string {
	idx := strings.Index(src, marker+"(")
	if idx < 0 {
		return ""
	}
	input := src[idx+len(marker)+1:]

	depth := 1
	inSingle := false
	inDouble := false
	escaped := false
	for i := 0; i < len(input); i++ {
		ch := input[i]
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

		switch ch {
		case '\'':
			inSingle = true
		case '"':
			inDouble = true
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return strings.TrimSpace(input[:i])
			}
		}
	}
	return ""
}

func secondTopLevelArg(args string) string {
	first := firstTopLevelArg(args)
	if first == "" {
		return ""
	}
	rest := strings.TrimSpace(strings.TrimPrefix(args, first))
	if strings.HasPrefix(rest, ",") {
		rest = strings.TrimSpace(strings.TrimPrefix(rest, ","))
	}
	return firstTopLevelArg(rest)
}

func estimateSequenceLiteralElements(arg string) int {
	trimmed := strings.TrimSpace(arg)
	if strings.HasPrefix(trimmed, "[") {
		return estimateArrayLiteralElements(trimmed)
	}
	if strings.HasPrefix(trimmed, "(") && strings.HasSuffix(trimmed, ")") {
		mapped := "[" + strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "("), ")")) + "]"
		return estimateArrayLiteralElements(mapped)
	}
	return 0
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
