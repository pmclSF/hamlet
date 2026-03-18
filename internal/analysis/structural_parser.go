package analysis

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/pmclSF/terrain/internal/models"
)

// StructuralPromptParser performs bracket-aware, token-level parsing of source
// code to detect AI prompts and contexts. Unlike flat regex, this parser:
//
//   - Tracks bracket nesting to find real array/object boundaries
//   - Identifies assignment targets (variable names)
//   - Understands function boundaries
//   - Distinguishes string content from code structure
//
// This is NOT a full AST parser (no tree-sitter dependency), but it provides
// structural guarantees that regex cannot: bracket matching, scope tracking,
// and assignment-target resolution.
//
// Detection tier: TierStructural for bracket-matched patterns,
// TierSemantic for framework-pattern matches within structural context.

// DetectorID constants — aliases to the canonical constants in models.
const (
	DetectorStructuralMessageArray = models.DetectorBracketMessageArray
	DetectorStructuralFewShot      = models.DetectorBracketFewShot
	DetectorStructuralPromptAssign = models.DetectorBracketPromptAssign
	DetectorStructuralPromptFunc   = models.DetectorPromptBuilderFunc
	DetectorSemanticTemplatePrompt = models.DetectorTemplatePrompt
	DetectorContentStringPrompt    = models.DetectorContentString
)

// ParseStructural performs bracket-aware detection of AI prompts/contexts.
func ParseStructural(relPath, src, lang string) []models.CodeSurface {
	switch lang {
	case "js":
		return parseStructuralJS(relPath, src)
	case "python":
		return parseStructuralPython(relPath, src)
	default:
		return nil
	}
}

// --- JS/TS structural parser ---

func parseStructuralJS(relPath, src string) []models.CodeSurface {
	var surfaces []models.CodeSurface
	pkg := inferSurfacePackage(relPath)
	seen := map[string]bool{}
	lines := strings.Split(src, "\n")

	add := func(name, detectorID, reason string, line int, confidence float64, tier string) {
		sid := models.BuildSurfaceID(relPath, name+"_L"+strconv.Itoa(line), "")
		if seen[sid] {
			return
		}
		seen[sid] = true
		surfaces = append(surfaces, models.CodeSurface{
			SurfaceID:     sid,
			Name:          name,
			Path:          relPath,
			Kind:          models.SurfaceContext,
			Language:      "js",
			Package:       pkg,
			Line:          line,
			Exported:      false,
			DetectionTier: tier,
			Confidence:    confidence,
			Reason:        "[" + detectorID + "] " + reason,
		})
	}

	// 1. Find message arrays via bracket-aware scanning.
	// Look for: variable = [ { role: "system", ... }, { role: "user", ... } ]
	msgArrays := findJSMessageArrays(src, lines)
	for _, ma := range msgArrays {
		add("message_array_"+ma.varName, DetectorStructuralMessageArray,
			"bracket-matched message array '"+ma.varName+"' with "+strconv.Itoa(ma.roleCount)+" role entries",
			ma.line, 0.95, models.TierStructural)
	}

	// 2. Find few-shot example arrays.
	fewShots := findJSFewShotArrays(src, lines)
	for _, fs := range fewShots {
		add("few_shot_"+fs.varName, DetectorStructuralFewShot,
			"bracket-matched few-shot array '"+fs.varName+"' with "+strconv.Itoa(fs.exampleCount)+" examples",
			fs.line, 0.90, models.TierStructural)
	}

	// 3. Find exported/assigned prompt constants.
	prompts := findJSPromptAssignments(lines)
	for _, p := range prompts {
		add("prompt_const_"+p.varName, DetectorStructuralPromptAssign,
			"assigned prompt constant '"+p.varName+"'",
			p.line, 0.88, models.TierStructural)
	}

	// 4. Find prompt builder functions (functions returning message-like content).
	builders := findJSPromptBuilderFuncs(src, lines)
	for _, b := range builders {
		add("prompt_builder_"+b.funcName, DetectorStructuralPromptFunc,
			"function '"+b.funcName+"' returns AI message payload",
			b.line, 0.85, models.TierSemantic)
	}

	return surfaces
}

type messageArrayDetection struct {
	varName   string
	line      int
	roleCount int
}

type fewShotDetection struct {
	varName      string
	line         int
	exampleCount int
}

type promptAssignment struct {
	varName string
	line    int
}

type promptBuilderFunc struct {
	funcName string
	line     int
}

// findJSMessageArrays scans for bracket-matched arrays containing role/content objects.
func findJSMessageArrays(src string, lines []string) []messageArrayDetection {
	var results []messageArrayDetection

	for i, line := range lines {
		// Look for: const/let/var X = [
		varName := extractAssignmentTarget(line, "[")
		if varName == "" {
			continue
		}

		// Find the matching bracket, counting roles within.
		arrayContent := extractBracketContent(src, lines, i, '[', ']')
		if arrayContent == "" {
			continue
		}

		// Count role entries within the matched brackets.
		roleCount := countRoleEntries(arrayContent)
		if roleCount >= 2 {
			results = append(results, messageArrayDetection{
				varName:   varName,
				line:      i + 1,
				roleCount: roleCount,
			})
		}
	}

	return results
}

// findJSFewShotArrays scans for arrays of objects with input/output structure.
func findJSFewShotArrays(src string, lines []string) []fewShotDetection {
	var results []fewShotDetection

	for i, line := range lines {
		varName := extractAssignmentTarget(line, "[")
		if varName == "" {
			continue
		}

		arrayContent := extractBracketContent(src, lines, i, '[', ']')
		if arrayContent == "" {
			continue
		}

		exampleCount := countFewShotEntries(arrayContent)
		if exampleCount >= 2 {
			results = append(results, fewShotDetection{
				varName:      varName,
				line:         i + 1,
				exampleCount: exampleCount,
			})
		}
	}

	return results
}

// findJSPromptAssignments finds variables assigned long AI-instructed strings.
func findJSPromptAssignments(lines []string) []promptAssignment {
	var results []promptAssignment
	aiMarkers := aiStringMarkers // reuse from prompt_parser.go

	for i, line := range lines {
		// Look for: const X = "..." or const X = `...`
		varName, content := extractStringAssignment(line)
		if varName == "" || len(content) < 60 {
			continue
		}
		if len(aiMarkers.FindAllString(content, -1)) >= 2 {
			results = append(results, promptAssignment{varName: varName, line: i + 1})
		}
	}
	return results
}

// findJSPromptBuilderFuncs finds functions whose name suggests prompt building
// and that contain message/role patterns in their body.
func findJSPromptBuilderFuncs(src string, lines []string) []promptBuilderFunc {
	var results []promptBuilderFunc
	promptFuncNames := []string{"build", "create", "generate", "format", "construct", "assemble", "compose"}

	for i, line := range lines {
		// Look for: function X( or const X = (
		funcName := extractFuncName(line)
		if funcName == "" {
			continue
		}
		lower := strings.ToLower(funcName)
		isPromptFunc := false
		for _, prefix := range promptFuncNames {
			if strings.HasPrefix(lower, prefix) && (strings.Contains(lower, "prompt") || strings.Contains(lower, "message") || strings.Contains(lower, "context")) {
				isPromptFunc = true
				break
			}
		}
		if !isPromptFunc {
			continue
		}

		// Check if the function body contains role/message patterns.
		body := extractBracketContent(src, lines, i, '{', '}')
		if body != "" && (strings.Contains(body, "role") || strings.Contains(body, "system") || strings.Contains(body, "assistant")) {
			results = append(results, promptBuilderFunc{funcName: funcName, line: i + 1})
		}
	}
	return results
}

// --- Python structural parser ---

func parseStructuralPython(relPath, src string) []models.CodeSurface {
	var surfaces []models.CodeSurface
	pkg := inferSurfacePackage(relPath)
	seen := map[string]bool{}
	lines := strings.Split(src, "\n")

	add := func(name, detectorID, reason string, line int, confidence float64, tier string) {
		sid := models.BuildSurfaceID(relPath, name+"_L"+strconv.Itoa(line), "")
		if seen[sid] {
			return
		}
		seen[sid] = true
		surfaces = append(surfaces, models.CodeSurface{
			SurfaceID:     sid,
			Name:          name,
			Path:          relPath,
			Kind:          models.SurfaceContext,
			Language:      "python",
			Package:       pkg,
			Line:          line,
			Exported:      false,
			DetectionTier: tier,
			Confidence:    confidence,
			Reason:        "[" + detectorID + "] " + reason,
		})
	}

	// 1. Message arrays: X = [{"role": "system", ...}, ...]
	for i, line := range lines {
		varName := extractPyAssignmentTarget(line, "[")
		if varName == "" {
			continue
		}
		arrayContent := extractBracketContent(src, lines, i, '[', ']')
		if arrayContent == "" {
			continue
		}
		roleCount := countPyRoleEntries(arrayContent)
		if roleCount >= 2 {
			add("message_array_"+varName, DetectorStructuralMessageArray,
				"bracket-matched message list '"+varName+"' with "+strconv.Itoa(roleCount)+" role entries",
				i+1, 0.95, models.TierStructural)
		}
	}

	// 2. Few-shot arrays.
	for i, line := range lines {
		varName := extractPyAssignmentTarget(line, "[")
		if varName == "" {
			continue
		}
		arrayContent := extractBracketContent(src, lines, i, '[', ']')
		if arrayContent == "" {
			continue
		}
		if countPyFewShotEntries(arrayContent) >= 2 {
			add("few_shot_"+varName, DetectorStructuralFewShot,
				"bracket-matched few-shot list '"+varName+"'",
				i+1, 0.90, models.TierStructural)
		}
	}

	// 3. Functions returning message lists.
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "def ") {
			continue
		}
		funcName := extractPyFuncName(trimmed)
		if funcName == "" {
			continue
		}
		lower := strings.ToLower(funcName)
		if !strings.Contains(lower, "prompt") && !strings.Contains(lower, "message") && !strings.Contains(lower, "context") {
			continue
		}
		// Check body for return with role pattern.
		body := extractPythonFuncBody(lines, i)
		if strings.Contains(body, "role") && strings.Contains(body, "return") {
			add("prompt_builder_"+funcName, DetectorStructuralPromptFunc,
				"function '"+funcName+"' returns AI message payload",
				i+1, 0.85, models.TierSemantic)
		}
	}

	return surfaces
}

// --- Shared structural helpers ---

// extractAssignmentTarget extracts the variable name from "const X = [" or "let X = [".
func extractAssignmentTarget(line, openToken string) string {
	trimmed := strings.TrimSpace(line)
	// const/let/var X = [
	for _, kw := range []string{"const ", "let ", "var "} {
		if strings.HasPrefix(trimmed, kw) {
			rest := trimmed[len(kw):]
			eqIdx := strings.Index(rest, "=")
			if eqIdx < 0 {
				continue
			}
			name := strings.TrimSpace(rest[:eqIdx])
			afterEq := strings.TrimSpace(rest[eqIdx+1:])
			if strings.HasPrefix(afterEq, openToken) {
				return name
			}
		}
	}
	return ""
}

// extractPyAssignmentTarget extracts variable name from "X = [" in Python.
func extractPyAssignmentTarget(line, openToken string) string {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "def ") || strings.HasPrefix(trimmed, "class ") {
		return ""
	}
	eqIdx := strings.Index(trimmed, "=")
	if eqIdx < 1 {
		return ""
	}
	// Ensure it's not ==, !=, <=, >=
	if eqIdx > 0 && (trimmed[eqIdx-1] == '!' || trimmed[eqIdx-1] == '<' || trimmed[eqIdx-1] == '>') {
		return ""
	}
	if eqIdx+1 < len(trimmed) && trimmed[eqIdx+1] == '=' {
		return ""
	}
	name := strings.TrimSpace(trimmed[:eqIdx])
	if !isValidIdentifier(name) {
		return ""
	}
	afterEq := strings.TrimSpace(trimmed[eqIdx+1:])
	if strings.HasPrefix(afterEq, openToken) {
		return name
	}
	return ""
}

// extractBracketContent finds the content between matching brackets starting
// from the given line. Returns empty string if brackets don't match within 30 lines.
func extractBracketContent(src string, lines []string, startLine int, open, close byte) string {
	depth := 0
	inString := byte(0)
	var content strings.Builder
	started := false

	for i := startLine; i < len(lines) && i < startLine+30; i++ {
		for j := 0; j < len(lines[i]); j++ {
			ch := lines[i][j]

			// Simple string tracking (not perfect but good enough for structure).
			if inString != 0 {
				if ch == inString && (j == 0 || lines[i][j-1] != '\\') {
					inString = 0
				}
				if started {
					content.WriteByte(ch)
				}
				continue
			}
			if ch == '"' || ch == '\'' || ch == '`' {
				inString = ch
				if started {
					content.WriteByte(ch)
				}
				continue
			}

			if ch == open {
				depth++
				if !started {
					started = true
					continue // Don't include the opening bracket.
				}
			}
			if ch == close {
				depth--
				if depth == 0 {
					return content.String()
				}
			}
			if started {
				content.WriteByte(ch)
			}
		}
		if started {
			content.WriteByte('\n')
		}
	}
	return "" // Unmatched.
}

func countRoleEntries(content string) int {
	count := 0
	// Count { role: "system"|"user"|"assistant" } patterns.
	idx := 0
	for idx < len(content) {
		pos := strings.Index(content[idx:], "role")
		if pos < 0 {
			break
		}
		idx += pos + 4
		// Look for : followed by "system", "user", or "assistant".
		rest := strings.TrimSpace(content[idx:])
		if len(rest) > 0 && rest[0] == ':' {
			rest = strings.TrimSpace(rest[1:])
			if strings.HasPrefix(rest, `"system"`) || strings.HasPrefix(rest, `'system'`) ||
				strings.HasPrefix(rest, `"user"`) || strings.HasPrefix(rest, `'user'`) ||
				strings.HasPrefix(rest, `"assistant"`) || strings.HasPrefix(rest, `'assistant'`) {
				count++
			}
		}
	}
	return count
}

func countPyRoleEntries(content string) int {
	count := 0
	idx := 0
	for idx < len(content) {
		// Look for "role": "system"|"user"|"assistant" or 'role': ...
		pos := strings.Index(content[idx:], "role")
		if pos < 0 {
			break
		}
		idx += pos + 4
		rest := strings.TrimSpace(content[idx:])
		// Skip quote around "role" key.
		if len(rest) > 0 && (rest[0] == '"' || rest[0] == '\'') {
			rest = rest[1:]
		}
		rest = strings.TrimSpace(rest)
		if len(rest) > 0 && rest[0] == ':' {
			rest = strings.TrimSpace(rest[1:])
			if strings.HasPrefix(rest, `"system"`) || strings.HasPrefix(rest, `'system'`) ||
				strings.HasPrefix(rest, `"user"`) || strings.HasPrefix(rest, `'user'`) ||
				strings.HasPrefix(rest, `"assistant"`) || strings.HasPrefix(rest, `'assistant'`) {
				count++
			}
		}
	}
	return count
}

func countFewShotEntries(content string) int {
	// Count distinct objects with input-like keys.
	// Each { input: ... } or { "input": ... } = one example.
	count := 0
	lower := strings.ToLower(content)
	for _, pattern := range []string{
		`"input"`, `'input'`, "input:",
		`"user"`, `'user'`, "user:",
		`"question"`, `'question'`, "question:",
	} {
		count += strings.Count(lower, pattern)
	}
	// Approximate: "input" appears once per object in the array.
	// But the key may appear as both "input": and "input" in the same object.
	// Use max of (count of any single key pattern) as the example count.
	maxSingle := 0
	for _, key := range []string{"input", "user", "question"} {
		n := strings.Count(lower, key)
		if n > maxSingle {
			maxSingle = n
		}
	}
	return maxSingle
}

func countPyFewShotEntries(content string) int {
	return countFewShotEntries(content)
}

func extractStringAssignment(line string) (string, string) {
	trimmed := strings.TrimSpace(line)
	for _, kw := range []string{"const ", "let ", "var "} {
		if !strings.HasPrefix(trimmed, kw) {
			continue
		}
		rest := trimmed[len(kw):]
		eqIdx := strings.Index(rest, "=")
		if eqIdx < 0 {
			continue
		}
		name := strings.TrimSpace(rest[:eqIdx])
		afterEq := strings.TrimSpace(rest[eqIdx+1:])

		// Extract string content from "..." or `...`
		if len(afterEq) > 1 {
			quote := afterEq[0]
			if quote == '"' || quote == '\'' || quote == '`' {
				endIdx := strings.LastIndex(afterEq[1:], string(quote))
				if endIdx > 0 {
					return name, afterEq[1 : endIdx+1]
				}
			}
		}
	}
	return "", ""
}

func extractFuncName(line string) string {
	trimmed := strings.TrimSpace(line)
	// function X(
	if strings.HasPrefix(trimmed, "function ") || strings.HasPrefix(trimmed, "async function ") {
		rest := trimmed
		if strings.HasPrefix(rest, "async ") {
			rest = rest[6:]
		}
		rest = rest[9:] // "function "
		parenIdx := strings.Index(rest, "(")
		if parenIdx > 0 {
			return strings.TrimSpace(rest[:parenIdx])
		}
	}
	// const X = ( or const X = async (
	if strings.HasPrefix(trimmed, "export ") {
		trimmed = trimmed[7:]
	}
	for _, kw := range []string{"const ", "let "} {
		if strings.HasPrefix(trimmed, kw) {
			rest := trimmed[len(kw):]
			eqIdx := strings.Index(rest, "=")
			if eqIdx > 0 {
				name := strings.TrimSpace(rest[:eqIdx])
				afterEq := strings.TrimSpace(rest[eqIdx+1:])
				if strings.HasPrefix(afterEq, "(") || strings.HasPrefix(afterEq, "async") || strings.HasPrefix(afterEq, "function") {
					return name
				}
			}
		}
	}
	return ""
}

func extractPyFuncName(line string) string {
	// def X(
	if strings.HasPrefix(line, "def ") {
		rest := line[4:]
		parenIdx := strings.Index(rest, "(")
		if parenIdx > 0 {
			return strings.TrimSpace(rest[:parenIdx])
		}
	}
	return ""
}

func extractPythonFuncBody(lines []string, startLine int) string {
	// Simple: collect indented lines after the def line.
	var body strings.Builder
	baseIndent := countLeadingSpaces(lines[startLine])
	for i := startLine + 1; i < len(lines) && i < startLine+20; i++ {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		indent := countLeadingSpaces(lines[i])
		if indent <= baseIndent {
			break
		}
		body.WriteString(lines[i])
		body.WriteByte('\n')
	}
	return body.String()
}

func countLeadingSpaces(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else if ch == '\t' {
			count += 4
		} else {
			break
		}
	}
	return count
}

func isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, ch := range s {
		if i == 0 && !unicode.IsLetter(ch) && ch != '_' {
			return false
		}
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			return false
		}
	}
	return true
}
