package analysis

import (
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// Detector ID constants — aliases to the canonical constants in models.
// The old names (DetectorAST*) are preserved as package-level names
// so that existing references and tests continue to compile.
const (
	DetectorASTMessageArray  = models.DetectorASTMessageArray
	DetectorASTSystemPrompt  = models.DetectorASTSystemPrompt
	DetectorASTFewShot       = models.DetectorASTFewShot
	DetectorASTPromptBuilder = models.DetectorASTPromptBuilder
	DetectorASTTemplateCall  = models.DetectorASTTemplateCall
)

// ParsePromptAST performs AST-level detection of AI prompts and contexts.
//
// For Python and JS/TS, this uses a deep structural analysis that tracks:
//   - Object/dict literal shapes ({role: ..., content: ...})
//   - Framework constructors (SystemMessage, ChatPromptTemplate, etc.)
//   - Assignment targets for system prompts
//   - Function return types that produce message payloads
//
// For Go, this uses go/parser for real AST analysis.
//
// Detection tier: TierStructural for AST-verified patterns.
// All detections carry evidence metadata for traceability.
func ParsePromptAST(relPath, src, lang string) []models.CodeSurface {
	switch lang {
	case "js":
		return parsePromptASTJS(relPath, src)
	case "python":
		return parsePromptASTPython(relPath, src)
	case "go":
		return parsePromptASTGo(relPath, src)
	default:
		return nil
	}
}

// --- JS/TS AST-level prompt detection ---

// --- Capability-based AI patterns (framework-agnostic) ---

// SDK client constructors: new OpenAI(), new Anthropic(), new Cohere(), etc.
// Capability: "instantiates an LLM provider client"
var jsSDKConstructor = regexp.MustCompile(
	`\bnew\s+(?:OpenAI|Anthropic|Cohere|GoogleGenerativeAI|Mistral|Groq|Together)\s*\(`)

// Completion/generation calls: .create(), generateText(), streamText()
// Capability: "invokes an LLM for text generation"
var jsGenerationCall = regexp.MustCompile(
	`\b(?:` +
		// OpenAI/Anthropic style: client.chat.completions.create(
		`(?:openai|client|api|anthropic)\.(?:chat\.completions|completions|messages|ChatCompletion)\.create|` +
		// Vercel AI SDK: generateText(, streamText(, generateObject(
		`(?:generateText|streamText|generateObject|streamObject)\s*\(|` +
		// Vercel AI SDK React hooks: useChat(, useCompletion(
		`(?:useChat|useCompletion|useAssistant)\s*\()`)

// Structured output patterns: response_format, tool definitions
// Capability: "constrains LLM output to a schema"
var jsStructuredOutput = regexp.MustCompile(
	`\b(?:` +
		// OpenAI JSON mode / structured outputs
		`response_format\s*:\s*\{\s*type\s*:\s*["']json|` +
		// Vercel AI SDK tool(): tool({ description:, parameters: })
		`\btool\s*\(\s*\{\s*(?:description|parameters)\s*:|` +
		// Generic tools array with function type
		`type\s*:\s*["']function["']\s*,\s*function\s*:\s*\{)`)

// Framework message constructors (LangChain, LlamaIndex).
var jsFrameworkConstructors = regexp.MustCompile(
	`\bnew\s+(?:SystemMessage|HumanMessage|AIMessage|ChatMessage|` +
		`SystemChatMessage|HumanChatMessage|AIChatMessage)\s*\(`)

// Template factories (LangChain).
var jsTemplateFactories = regexp.MustCompile(
	`\b(?:ChatPromptTemplate|PromptTemplate|FewShotPromptTemplate|` +
		`FewShotChatMessagePromptTemplate|MessagePromptTemplate|` +
		`SystemMessagePromptTemplate|HumanMessagePromptTemplate)` +
		`\.(?:from[A-Z]\w*|of)\s*\(`)

// OpenAI-style API call (kept for backward compat, subsumed by jsGenerationCall).
var jsOpenAICallPattern = regexp.MustCompile(
	`\b(?:openai|client|api)\.(?:chat\.completions|completions|ChatCompletion)\.create\s*\(`)

// Anthropic-style (kept for backward compat).
var jsAnthropicCallPattern = regexp.MustCompile(
	`\b(?:anthropic|client)\.messages\.create\s*\(`)

func parsePromptASTJS(relPath, src string) []models.CodeSurface {
	var surfaces []models.CodeSurface
	pkg := inferSurfacePackage(relPath)
	seen := map[string]bool{}
	lines := strings.Split(src, "\n")

	add := func(name string, kind models.CodeSurfaceKind, detectorID, reason string, line int, confidence float64) {
		sid := models.BuildSurfaceID(relPath, name+"_L"+strconv.Itoa(line), "")
		if seen[sid] {
			return
		}
		seen[sid] = true
		surfaces = append(surfaces, models.CodeSurface{
			SurfaceID:     sid,
			Name:          name,
			Path:          relPath,
			Kind:          kind,
			Language:      "js",
			Package:       pkg,
			Line:          line,
			Exported:      false,
			DetectionTier: models.TierStructural,
			Confidence:    confidence,
			Reason:        "[" + detectorID + "] " + reason,
		})
	}

	// 1. Detect message arrays by scanning for arrays of objects with role+content shape.
	arrayRanges := findJSObjectArrayRanges(src, lines)
	for _, ar := range arrayRanges {
		if ar.hasRoleContent && ar.objectCount >= 2 {
			// Determine kind: if it has a system role entry, it's a prompt.
			kind := models.SurfacePrompt
			if ar.hasSystemRole {
				kind = models.SurfaceContext
			}
			add("message_array_"+ar.varName, kind, DetectorASTMessageArray,
				"AST: array of "+strconv.Itoa(ar.objectCount)+" objects with role/content shape in '"+ar.varName+"'",
				ar.line, 0.97)
		}
	}

	// 2. Detect framework constructors (LangChain, etc.).
	for _, m := range jsFrameworkConstructors.FindAllStringIndex(src, -1) {
		line := 1 + strings.Count(src[:m[0]], "\n")
		constructorName := extractConstructorName(src[m[0]:m[1]])
		isSystem := strings.Contains(strings.ToLower(constructorName), "system")
		kind := models.SurfacePrompt
		if isSystem {
			kind = models.SurfaceContext
		}
		add("framework_msg_"+sanitizeVarName(constructorName), kind, DetectorASTSystemPrompt,
			"AST: framework message constructor "+constructorName,
			line, 0.95)
	}

	// 3. Detect template factory calls.
	for _, m := range jsTemplateFactories.FindAllStringIndex(src, -1) {
		line := 1 + strings.Count(src[:m[0]], "\n")
		factoryName := extractFactoryName(src[m[0]:m[1]])
		add("template_prompt_"+sanitizeVarName(factoryName), models.SurfacePrompt, DetectorASTTemplateCall,
			"AST: prompt template factory "+factoryName,
			line, 0.95)
	}

	// 4. Detect OpenAI/Anthropic API calls with messages parameter.
	for _, pat := range []*regexp.Regexp{jsOpenAICallPattern, jsAnthropicCallPattern} {
		for _, m := range pat.FindAllStringIndex(src, -1) {
			line := 1 + strings.Count(src[:m[0]], "\n")
			// Check if the call body contains a messages key.
			callBody := extractBracketContent(src, lines, line-1, '(', ')')
			if callBody != "" && strings.Contains(callBody, "messages") {
				varName := extractCallAssignTarget(lines, line-1)
				if varName == "" {
					varName = "api_call"
				}
				add("api_prompt_"+varName, models.SurfacePrompt, DetectorASTMessageArray,
					"AST: API call with messages parameter",
					line, 0.93)
			}
		}
	}

	// 4b. Detect SDK client constructors (OpenAI, Anthropic, Cohere, etc.).
	// Capability: "instantiates an LLM provider client"
	for _, m := range jsSDKConstructor.FindAllStringIndex(src, -1) {
		line := 1 + strings.Count(src[:m[0]], "\n")
		name := extractConstructorName(src[m[0]:m[1]])
		add("sdk_client_"+sanitizeVarName(name), models.SurfaceAgent, DetectorASTPromptBuilder,
			"AST: LLM SDK client constructor "+name,
			line, 0.92)
	}

	// 4c. Detect generation/streaming calls (Vercel AI SDK, OpenAI, etc.).
	// Capability: "invokes an LLM for text generation"
	for _, m := range jsGenerationCall.FindAllStringIndex(src, -1) {
		line := 1 + strings.Count(src[:m[0]], "\n")
		callBody := extractBracketContent(src, lines, line-1, '(', ')')
		varName := extractCallAssignTarget(lines, line-1)
		if varName == "" {
			varName = "generation"
		}
		kind := models.SurfacePrompt
		if callBody != "" && (strings.Contains(callBody, "system") || strings.Contains(callBody, "messages")) {
			kind = models.SurfaceContext
		}
		add("llm_call_"+varName, kind, DetectorASTMessageArray,
			"AST: LLM generation/streaming call",
			line, 0.90)
	}

	// 4d. Detect structured output patterns (tool definitions, JSON mode).
	// Capability: "constrains LLM output to a schema"
	for _, m := range jsStructuredOutput.FindAllStringIndex(src, -1) {
		line := 1 + strings.Count(src[:m[0]], "\n")
		add("structured_output", models.SurfaceToolDef, DetectorASTTemplateCall,
			"AST: structured output / tool definition pattern",
			line, 0.88)
	}

	// 5. Detect system prompt string assignments.
	// Look for: const systemPrompt = "You are..." or const SYSTEM_PROMPT = `...`
	for i, line := range lines {
		varName, content := extractStringAssignment(line)
		if varName == "" || len(content) < 30 {
			continue
		}
		if !isSystemPromptVarName(varName) {
			continue
		}
		// Require at least 1 AI marker (lower threshold for named system prompts).
		if len(aiStringMarkers.FindAllString(content, -1)) >= 1 {
			add("system_prompt_"+varName, models.SurfaceContext, DetectorASTSystemPrompt,
				"AST: system prompt assignment '"+varName+"' with AI instruction content",
				i+1, 0.92)
		}
	}

	// 6. Detect multi-line template literals assigned to prompt-named variables.
	for i, line := range lines {
		varName := extractTemplateAssignTarget(line)
		if varName == "" || !isPromptVarName(varName) {
			continue
		}
		// Extract the template literal content.
		templateContent := extractBacktickContent(src, lines, i)
		if len(templateContent) < 40 {
			continue
		}
		if len(aiStringMarkers.FindAllString(templateContent, -1)) >= 1 {
			add("template_prompt_"+varName, models.SurfacePrompt, DetectorASTTemplateCall,
				"AST: template literal prompt '"+varName+"'",
				i+1, 0.90)
		}
	}

	return surfaces
}

// --- Python AST-level prompt detection ---

// --- Python capability-based patterns ---

// Python SDK client constructors.
var pySDKConstructor = regexp.MustCompile(
	`\b(?:OpenAI|Anthropic|Cohere|GoogleGenerativeAI|Mistral|Groq|Together)\s*\(`)

// Python generation calls (OpenAI, Anthropic, Instructor).
var pyGenerationCall = regexp.MustCompile(
	`\b(?:` +
		`(?:openai|client|api|anthropic)\.(?:chat\.completions|completions|messages|ChatCompletion)\.create|` +
		// Instructor: client.chat.completions.create(response_model=...)
		`response_model\s*=)`)

// Python AI decorators: @tool, @prompt_template, @marvin.fn, @openai.call, etc.
// Capability: "declares a function as an AI-callable tool or prompt"
var pyAIDecorator = regexp.MustCompile(
	`@(?:` +
		// Generic tool decorators
		`tool|function_tool|register_tool|` +
		// Mirascope: @openai.call, @anthropic.call, @prompt_template
		`(?:openai|anthropic|mistral|cohere|groq)\.call|` +
		`mirascope\.prompt_template|prompt_template|` +
		// Marvin: @marvin.fn, @marvin.model
		`marvin\.(?:fn|model)|` +
		// Instructor: @instructor.patch (rare but possible)
		`instructor\.(?:patch|tool))\b`)

// Python DSPy patterns: dspy.Signature, dspy.Module, dspy.ChainOfThought, dspy.Retrieve.
var pyDSPyPattern = regexp.MustCompile(
	`\b(?:dspy\.(?:Signature|Module|ChainOfThought|Retrieve|Predict|Assert|Suggest|` +
		`ReAct|ProgramOfThought|MultiChainComparison))\b`)

// Python structured output: Pydantic BaseModel used with response_model or create_model.
var pyStructuredOutput = regexp.MustCompile(
	`\b(?:` +
		// Instructor: response_model=MyModel
		`response_model\s*=\s*\w+|` +
		// Marvin: marvin.extract, marvin.classify, marvin.cast
		`marvin\.(?:extract|classify|cast|generate)\s*\(|` +
		// Generic: create_model, model_validate
		`(?:create_model|model_validate)\s*\()`)

// Python framework constructors (LangChain, LlamaIndex — kept for backward compat).
var pyFrameworkConstructors = regexp.MustCompile(
	`\b(?:SystemMessage|HumanMessage|AIMessage|ChatMessage|` +
		`SystemMessagePromptTemplate|HumanMessagePromptTemplate|` +
		`ChatPromptTemplate|PromptTemplate|FewShotPromptTemplate|` +
		`FewShotChatMessagePromptTemplate)\s*\(`)

// Python template factories (LangChain — kept for backward compat).
var pyTemplateFactories = regexp.MustCompile(
	`\b(?:ChatPromptTemplate|PromptTemplate|FewShotPromptTemplate)` +
		`\.(?:from_messages|from_template|from_examples)\s*\(`)

// Python OpenAI/Anthropic call (kept for backward compat, subsumed by pyGenerationCall).
var pyOpenAICallPattern = regexp.MustCompile(
	`\b(?:openai\.ChatCompletion\.create|client\.chat\.completions\.create|` +
		`anthropic\.messages\.create|client\.messages\.create)\s*\(`)

func parsePromptASTPython(relPath, src string) []models.CodeSurface {
	var surfaces []models.CodeSurface
	pkg := inferSurfacePackage(relPath)
	seen := map[string]bool{}
	lines := strings.Split(src, "\n")

	add := func(name string, kind models.CodeSurfaceKind, detectorID, reason string, line int, confidence float64) {
		sid := models.BuildSurfaceID(relPath, name+"_L"+strconv.Itoa(line), "")
		if seen[sid] {
			return
		}
		seen[sid] = true
		surfaces = append(surfaces, models.CodeSurface{
			SurfaceID:     sid,
			Name:          name,
			Path:          relPath,
			Kind:          kind,
			Language:      "python",
			Package:       pkg,
			Line:          line,
			Exported:      false,
			DetectionTier: models.TierStructural,
			Confidence:    confidence,
			Reason:        "[" + detectorID + "] " + reason,
		})
	}

	// 1. Detect message lists: X = [{"role": "system", ...}, {"role": "user", ...}]
	pyArrayRanges := findPyDictListRanges(src, lines)
	for _, ar := range pyArrayRanges {
		if ar.hasRoleContent && ar.objectCount >= 2 {
			kind := models.SurfacePrompt
			if ar.hasSystemRole {
				kind = models.SurfaceContext
			}
			add("message_list_"+ar.varName, kind, DetectorASTMessageArray,
				"AST: list of "+strconv.Itoa(ar.objectCount)+" dicts with role/content shape in '"+ar.varName+"'",
				ar.line, 0.97)
		}
	}

	// 2. Detect framework constructors.
	for _, m := range pyFrameworkConstructors.FindAllStringIndex(src, -1) {
		line := 1 + strings.Count(src[:m[0]], "\n")
		constructorName := extractPyConstructorName(src[m[0]:m[1]])
		isSystem := strings.Contains(strings.ToLower(constructorName), "system")
		kind := models.SurfacePrompt
		if isSystem {
			kind = models.SurfaceContext
		}
		add("framework_msg_"+sanitizeVarName(constructorName), kind, DetectorASTSystemPrompt,
			"AST: framework constructor "+constructorName,
			line, 0.95)
	}

	// 3. Detect template factory calls.
	for _, m := range pyTemplateFactories.FindAllStringIndex(src, -1) {
		line := 1 + strings.Count(src[:m[0]], "\n")
		factoryName := extractPyFactoryName(src[m[0]:m[1]])
		add("template_prompt_"+sanitizeVarName(factoryName), models.SurfacePrompt, DetectorASTTemplateCall,
			"AST: prompt template factory "+factoryName,
			line, 0.95)
	}

	// 4. Detect OpenAI/Anthropic API calls.
	for _, m := range pyOpenAICallPattern.FindAllStringIndex(src, -1) {
		line := 1 + strings.Count(src[:m[0]], "\n")
		callBody := extractBracketContent(src, lines, line-1, '(', ')')
		if callBody != "" && strings.Contains(callBody, "messages") {
			varName := extractPyCallAssignTarget(lines, line-1)
			if varName == "" {
				varName = "api_call"
			}
			add("api_prompt_"+varName, models.SurfacePrompt, DetectorASTMessageArray,
				"AST: API call with messages parameter",
				line, 0.93)
		}
	}

	// 4b. Detect SDK client constructors.
	for _, m := range pySDKConstructor.FindAllStringIndex(src, -1) {
		line := 1 + strings.Count(src[:m[0]], "\n")
		name := extractPyConstructorName(src[m[0]:m[1]])
		add("sdk_client_"+sanitizeVarName(name), models.SurfaceAgent, DetectorASTPromptBuilder,
			"AST: LLM SDK client constructor "+name,
			line, 0.92)
	}

	// 4c. Detect AI decorators (@tool, @openai.call, @marvin.fn, etc.).
	for _, m := range pyAIDecorator.FindAllStringIndex(src, -1) {
		line := 1 + strings.Count(src[:m[0]], "\n")
		decoratorText := src[m[0]:m[1]]
		kind := models.SurfaceToolDef
		if strings.Contains(decoratorText, "call") || strings.Contains(decoratorText, "prompt") {
			kind = models.SurfacePrompt
		}
		add("ai_decorator_L"+strconv.Itoa(line), kind, DetectorASTTemplateCall,
			"AST: AI decorator "+decoratorText,
			line, 0.93)
	}

	// 4d. Detect DSPy module/signature patterns.
	for _, m := range pyDSPyPattern.FindAllStringIndex(src, -1) {
		line := 1 + strings.Count(src[:m[0]], "\n")
		patternText := src[m[0]:m[1]]
		kind := models.SurfacePrompt
		if strings.Contains(patternText, "Retrieve") {
			kind = models.SurfaceRetrieval
		} else if strings.Contains(patternText, "Module") {
			kind = models.SurfaceAgent
		}
		add("dspy_"+sanitizeVarName(patternText), kind, DetectorASTTemplateCall,
			"AST: DSPy pattern "+patternText,
			line, 0.90)
	}

	// 4e. Detect structured output patterns (Instructor, Marvin).
	for _, m := range pyStructuredOutput.FindAllStringIndex(src, -1) {
		line := 1 + strings.Count(src[:m[0]], "\n")
		add("structured_output_L"+strconv.Itoa(line), models.SurfaceToolDef, DetectorASTTemplateCall,
			"AST: structured output pattern (Instructor/Marvin)",
			line, 0.88)
	}

	// 5. Detect system prompt string assignments.
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		varName, content := extractPyStringAssignment(trimmed)
		if varName == "" || len(content) < 30 {
			continue
		}
		if !isSystemPromptVarName(varName) {
			continue
		}
		if len(aiStringMarkers.FindAllString(content, -1)) >= 1 {
			add("system_prompt_"+varName, models.SurfaceContext, DetectorASTSystemPrompt,
				"AST: system prompt assignment '"+varName+"' with AI instruction content",
				i+1, 0.92)
		}
	}

	// 6. Detect triple-quoted prompt strings assigned to prompt-named variables.
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		varName := extractPyTripleQuoteAssignTarget(trimmed)
		if varName == "" {
			continue
		}
		if !isPromptVarName(varName) && !isSystemPromptVarName(varName) {
			continue
		}
		// Extract content up to closing triple-quote.
		content := extractTripleQuoteContent(lines, i)
		if len(content) < 40 {
			continue
		}
		if len(aiStringMarkers.FindAllString(content, -1)) >= 1 {
			kind := models.SurfacePrompt
			detectorID := DetectorASTTemplateCall
			namePrefix := "template_prompt_"
			if isSystemPromptVarName(varName) {
				kind = models.SurfaceContext
				detectorID = DetectorASTSystemPrompt
				namePrefix = "system_prompt_"
			}
			add(namePrefix+varName, kind, detectorID,
				"AST: triple-quoted prompt '"+varName+"'",
				i+1, 0.92)
		}
	}

	// 7. Detect few-shot example lists with input/output dict structure.
	for _, ar := range pyArrayRanges {
		if ar.hasFewShotShape && ar.objectCount >= 2 && !ar.hasRoleContent {
			add("few_shot_"+ar.varName, models.SurfaceContext, DetectorASTFewShot,
				"AST: list of "+strconv.Itoa(ar.objectCount)+" dicts with input/output shape in '"+ar.varName+"'",
				ar.line, 0.93)
		}
	}

	return surfaces
}

// --- Go AST-level prompt detection ---

func parsePromptASTGo(relPath, src string) []models.CodeSurface {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, relPath, src, parser.ParseComments)
	if err != nil {
		return nil
	}

	var surfaces []models.CodeSurface
	pkg := inferSurfacePackage(relPath)
	seen := map[string]bool{}

	add := func(name string, kind models.CodeSurfaceKind, detectorID, reason string, line int, confidence float64) {
		sid := models.BuildSurfaceID(relPath, name+"_L"+strconv.Itoa(line), "")
		if seen[sid] {
			return
		}
		seen[sid] = true
		surfaces = append(surfaces, models.CodeSurface{
			SurfaceID:     sid,
			Name:          name,
			Path:          relPath,
			Kind:          kind,
			Language:      "go",
			Package:       pkg,
			Line:          line,
			Exported:      false,
			DetectionTier: models.TierStructural,
			Confidence:    confidence,
			Reason:        "[" + detectorID + "] " + reason,
		})
	}

	// Walk AST looking for composite literals with "role" fields.
	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CompositeLit:
			// Check for slice of structs/maps with "Role" or "role" field.
			if arrayType, ok := node.Type.(*ast.ArrayType); ok {
				_ = arrayType
				roleCount := 0
				hasSystem := false
				for _, elt := range node.Elts {
					if compLit, ok := elt.(*ast.CompositeLit); ok {
						for _, kv := range compLit.Elts {
							if kvExpr, ok := kv.(*ast.KeyValueExpr); ok {
								keyName := identName(kvExpr.Key)
								if strings.EqualFold(keyName, "role") {
									roleCount++
									if valStr := basicLitString(kvExpr.Value); strings.EqualFold(valStr, "system") {
										hasSystem = true
									}
								}
							}
						}
					}
				}
				if roleCount >= 2 {
					line := fset.Position(node.Pos()).Line
					kind := models.SurfacePrompt
					if hasSystem {
						kind = models.SurfaceContext
					}
					add("message_slice", kind, DetectorASTMessageArray,
						"AST: Go slice literal with "+strconv.Itoa(roleCount)+" role entries",
						line, 0.97)
				}
			}

		case *ast.GenDecl:
			// Check for const/var with prompt-like names containing long strings.
			for _, spec := range node.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok || len(vs.Names) == 0 || len(vs.Values) == 0 {
					continue
				}
				varName := vs.Names[0].Name
				if !isPromptVarName(varName) && !isSystemPromptVarName(varName) {
					continue
				}
				for _, val := range vs.Values {
					if bl, ok := val.(*ast.BasicLit); ok && bl.Kind == token.STRING {
						content, _ := strconv.Unquote(bl.Value)
						if content == "" && len(bl.Value) > 60 {
							// Raw backtick string — extract manually.
							content = strings.Trim(bl.Value, "`")
						}
						if len(content) >= 30 && len(aiStringMarkers.FindAllString(content, -1)) >= 1 {
							line := fset.Position(bl.Pos()).Line
							kind := models.SurfacePrompt
							if isSystemPromptVarName(varName) {
								kind = models.SurfaceContext
							}
							add("prompt_const_"+varName, kind, DetectorASTSystemPrompt,
								"AST: Go const/var '"+varName+"' with AI instruction content",
								line, 0.95)
						}
					}
				}
			}
		}
		return true
	})

	return surfaces
}

// --- Object/Array shape analysis (JS) ---

type arrayShapeInfo struct {
	varName        string
	line           int
	objectCount    int
	hasRoleContent bool
	hasSystemRole  bool
	hasFewShotShape bool
}

// findJSObjectArrayRanges scans for variable assignments to arrays of objects,
// analyzing the shape of each object to determine if it's a message array
// or few-shot example array.
func findJSObjectArrayRanges(src string, lines []string) []arrayShapeInfo {
	var results []arrayShapeInfo

	for i, line := range lines {
		varName := extractAssignmentTarget(line, "[")
		if varName == "" {
			continue
		}

		content := extractBracketContent(src, lines, i, '[', ']')
		if content == "" {
			continue
		}

		info := analyzeArrayShape(content)
		if info.objectCount >= 2 {
			info.varName = varName
			info.line = i + 1
			results = append(results, info)
		}
	}

	return results
}

// findPyDictListRanges scans Python code for list-of-dict assignments.
func findPyDictListRanges(src string, lines []string) []arrayShapeInfo {
	var results []arrayShapeInfo

	for i, line := range lines {
		varName := extractPyAssignmentTarget(line, "[")
		if varName == "" {
			continue
		}

		content := extractBracketContent(src, lines, i, '[', ']')
		if content == "" {
			continue
		}

		info := analyzeArrayShape(content)
		if info.objectCount >= 2 {
			info.varName = varName
			info.line = i + 1
			results = append(results, info)
		}
	}

	return results
}

// analyzeArrayShape examines the content between array brackets to determine
// the shape of contained objects.
func analyzeArrayShape(content string) arrayShapeInfo {
	info := arrayShapeInfo{}

	// Count objects with role+content keys.
	roleContentCount := 0
	fewShotCount := 0
	hasSystem := false

	// Simple object boundary tracking: count top-level { } blocks.
	depth := 0
	inString := byte(0)
	var currentObj strings.Builder
	objectCount := 0

	for j := 0; j < len(content); j++ {
		ch := content[j]

		if inString != 0 {
			if ch == inString && (j == 0 || content[j-1] != '\\') {
				inString = 0
			}
			currentObj.WriteByte(ch)
			continue
		}
		if ch == '"' || ch == '\'' || ch == '`' {
			inString = ch
			currentObj.WriteByte(ch)
			continue
		}

		if ch == '{' {
			depth++
			if depth == 1 {
				currentObj.Reset()
			}
			continue
		}
		if ch == '}' {
			depth--
			if depth == 0 {
				objectCount++
				obj := currentObj.String()
				objLower := strings.ToLower(obj)

				// Check for role+content shape.
				hasRole := strings.Contains(objLower, "role") &&
					(strings.Contains(objLower, `"system"`) || strings.Contains(objLower, `'system'`) ||
						strings.Contains(objLower, `"user"`) || strings.Contains(objLower, `'user'`) ||
						strings.Contains(objLower, `"assistant"`) || strings.Contains(objLower, `'assistant'`))
				hasContent := strings.Contains(objLower, "content")

				if hasRole && hasContent {
					roleContentCount++
					if strings.Contains(objLower, `"system"`) || strings.Contains(objLower, `'system'`) {
						hasSystem = true
					}
				}

				// Check for few-shot shape (input/output, question/answer).
				hasInput := strings.Contains(objLower, `"input"`) || strings.Contains(objLower, `'input'`) ||
					strings.Contains(objLower, `"question"`) || strings.Contains(objLower, `'question'`)
				hasOutput := strings.Contains(objLower, `"output"`) || strings.Contains(objLower, `'output'`) ||
					strings.Contains(objLower, `"answer"`) || strings.Contains(objLower, `'answer'`) ||
					strings.Contains(objLower, `"expected"`) || strings.Contains(objLower, `'expected'`)
				if hasInput && hasOutput {
					fewShotCount++
				}
			}
			continue
		}
		if depth >= 1 {
			currentObj.WriteByte(ch)
		}
	}

	info.objectCount = objectCount
	info.hasRoleContent = roleContentCount >= 2
	info.hasSystemRole = hasSystem
	info.hasFewShotShape = fewShotCount >= 2

	return info
}

// --- Helper functions ---

func isSystemPromptVarName(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "system") && (strings.Contains(lower, "prompt") || strings.Contains(lower, "message")) ||
		lower == "system_prompt" || lower == "systemprompt" ||
		lower == "system_message" || lower == "systemmessage" ||
		lower == "system_instruction" || lower == "systeminstruction" ||
		strings.HasPrefix(lower, "system_") && len(lower) > 7
}

func isPromptVarName(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "prompt") || strings.Contains(lower, "instruction") ||
		strings.Contains(lower, "system_msg") || strings.Contains(lower, "systemmsg") ||
		strings.Contains(lower, "template") && (strings.Contains(lower, "prompt") || strings.Contains(lower, "chat") || strings.Contains(lower, "ai"))
}

func extractConstructorName(match string) string {
	// "new SystemMessage(" → "SystemMessage"
	match = strings.TrimPrefix(match, "new ")
	match = strings.TrimSpace(match)
	idx := strings.Index(match, "(")
	if idx > 0 {
		return match[:idx]
	}
	return match
}

func extractPyConstructorName(match string) string {
	idx := strings.Index(match, "(")
	if idx > 0 {
		return strings.TrimSpace(match[:idx])
	}
	return match
}

func extractFactoryName(match string) string {
	idx := strings.Index(match, "(")
	if idx > 0 {
		return strings.TrimSpace(match[:idx])
	}
	return match
}

func extractPyFactoryName(match string) string {
	return extractFactoryName(match)
}

func sanitizeVarName(name string) string {
	// Replace dots and spaces with underscores for use in IDs.
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, " ", "_")
	return name
}

func extractCallAssignTarget(lines []string, lineIdx int) string {
	if lineIdx >= len(lines) {
		return ""
	}
	line := strings.TrimSpace(lines[lineIdx])
	for _, kw := range []string{"const ", "let ", "var "} {
		if strings.HasPrefix(line, kw) {
			rest := line[len(kw):]
			eqIdx := strings.Index(rest, "=")
			if eqIdx > 0 {
				return strings.TrimSpace(rest[:eqIdx])
			}
		}
	}
	return ""
}

func extractPyCallAssignTarget(lines []string, lineIdx int) string {
	if lineIdx >= len(lines) {
		return ""
	}
	line := strings.TrimSpace(lines[lineIdx])
	eqIdx := strings.Index(line, "=")
	if eqIdx > 0 && eqIdx+1 < len(line) && line[eqIdx+1] != '=' {
		name := strings.TrimSpace(line[:eqIdx])
		if isValidIdentifier(name) {
			return name
		}
	}
	return ""
}

func extractTemplateAssignTarget(line string) string {
	trimmed := strings.TrimSpace(line)
	for _, kw := range []string{"const ", "let ", "var ", "export const ", "export let "} {
		if strings.HasPrefix(trimmed, kw) {
			rest := trimmed[len(kw):]
			eqIdx := strings.Index(rest, "=")
			if eqIdx > 0 {
				afterEq := strings.TrimSpace(rest[eqIdx+1:])
				if strings.HasPrefix(afterEq, "`") {
					return strings.TrimSpace(rest[:eqIdx])
				}
			}
		}
	}
	return ""
}

func extractBacktickContent(src string, lines []string, startLine int) string {
	// Find the opening backtick on this line.
	line := lines[startLine]
	btIdx := strings.Index(line, "`")
	if btIdx < 0 {
		return ""
	}

	// Find content from after the backtick to the closing backtick.
	var content strings.Builder
	started := false
	for i := startLine; i < len(lines) && i < startLine+30; i++ {
		for j := 0; j < len(lines[i]); j++ {
			if !started {
				if i == startLine && j == btIdx {
					started = true
				}
				continue
			}
			if lines[i][j] == '`' {
				return content.String()
			}
			content.WriteByte(lines[i][j])
		}
		if started {
			content.WriteByte('\n')
		}
	}
	return ""
}

func extractPyStringAssignment(line string) (string, string) {
	eqIdx := strings.Index(line, "=")
	if eqIdx < 1 || (eqIdx+1 < len(line) && line[eqIdx+1] == '=') {
		return "", ""
	}
	if eqIdx > 0 && (line[eqIdx-1] == '!' || line[eqIdx-1] == '<' || line[eqIdx-1] == '>') {
		return "", ""
	}
	name := strings.TrimSpace(line[:eqIdx])
	if !isValidIdentifier(name) {
		return "", ""
	}
	afterEq := strings.TrimSpace(line[eqIdx+1:])
	if len(afterEq) > 1 {
		quote := afterEq[0]
		if quote == '"' || quote == '\'' {
			endIdx := strings.LastIndex(afterEq[1:], string(quote))
			if endIdx > 0 {
				return name, afterEq[1 : endIdx+1]
			}
		}
	}
	return "", ""
}

func extractPyTripleQuoteAssignTarget(line string) string {
	eqIdx := strings.Index(line, "=")
	if eqIdx < 1 {
		return ""
	}
	afterEq := strings.TrimSpace(line[eqIdx+1:])
	if strings.HasPrefix(afterEq, `"""`) || strings.HasPrefix(afterEq, `'''`) || strings.HasPrefix(afterEq, `f"""`) || strings.HasPrefix(afterEq, `f'''`) {
		name := strings.TrimSpace(line[:eqIdx])
		if isValidIdentifier(name) {
			return name
		}
	}
	return ""
}

func extractTripleQuoteContent(lines []string, startLine int) string {
	line := lines[startLine]
	// Find the opening triple quote.
	var quoteStyle string
	for _, q := range []string{`f"""`, `f'''`, `"""`, `'''`} {
		if idx := strings.Index(line, q); idx >= 0 {
			quoteStyle = q[len(q)-3:] // Always """ or '''
			break
		}
	}
	if quoteStyle == "" {
		return ""
	}

	// Find content between opening and closing triple quotes.
	openIdx := strings.Index(line, quoteStyle)
	afterOpen := line[openIdx+3:]
	// Check if closing is on the same line.
	if closeIdx := strings.Index(afterOpen, quoteStyle); closeIdx >= 0 {
		return afterOpen[:closeIdx]
	}

	var content strings.Builder
	content.WriteString(afterOpen)
	for i := startLine + 1; i < len(lines) && i < startLine+30; i++ {
		if closeIdx := strings.Index(lines[i], quoteStyle); closeIdx >= 0 {
			content.WriteString(lines[i][:closeIdx])
			return content.String()
		}
		content.WriteString(lines[i])
		content.WriteByte('\n')
	}
	return ""
}

// Go AST helpers.
func identName(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	if bl, ok := expr.(*ast.BasicLit); ok && bl.Kind == token.STRING {
		s, _ := strconv.Unquote(bl.Value)
		return s
	}
	return ""
}

func basicLitString(expr ast.Expr) string {
	if bl, ok := expr.(*ast.BasicLit); ok && bl.Kind == token.STRING {
		s, _ := strconv.Unquote(bl.Value)
		return s
	}
	return ""
}
