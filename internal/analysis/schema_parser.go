package analysis

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// ParseToolSchemas performs structured detection of tool schemas, structured
// output contracts, and response model definitions in source code.
//
// Detections:
//
//   JS/TS:
//   - Zod schemas (z.object, z.string, etc.) assigned to tool/output variables
//   - OpenAI tool registration ({ type: "function", function: { name: ... } })
//   - LangChain output parsers (StructuredOutputParser, JsonOutputParser)
//   - Function-calling tool arrays
//
//   Python:
//   - Pydantic BaseModel subclasses used as response models
//   - Instructor response_model= patterns
//   - OpenAI tools=[{...}] registration
//   - LangChain output parsers
//
// Each detection carries tier, confidence, and evidence metadata.
func ParseToolSchemas(relPath, src, lang string) []models.CodeSurface {
	switch lang {
	case "js":
		return parseJSSchemas(relPath, src)
	case "python":
		return parsePythonSchemas(relPath, src)
	default:
		return nil
	}
}

// --- JS/TS schema detection ---

var (
	// Zod schema: const X = z.object({...})
	jsZodSchema = regexp.MustCompile(`(?:const|let|var)\s+(\w+)\s*=\s*z\.(?:object|string|number|array|enum|union|discriminatedUnion)\s*\(`)

	// OpenAI tool registration: { type: "function", function: { name: "..." } }
	jsToolRegistration = regexp.MustCompile(`type\s*:\s*["']function["']\s*,\s*function\s*:\s*\{`)

	// LangChain output parsers
	jsOutputParser = regexp.MustCompile(`(?:new\s+)?(?:StructuredOutputParser|JsonOutputParser|OutputFixingParser|CommaSeparatedListOutputParser|CustomListOutputParser)\s*[.(]`)

	// Function-calling tools array: tools: [{ ... }]
	jsToolsArray = regexp.MustCompile(`tools\s*:\s*\[`)

	// Response format / structured outputs: response_format: { type: "json_schema" }
	jsResponseFormat = regexp.MustCompile(`response_format\s*:\s*\{\s*type\s*:\s*["']json_schema["']`)
)

func parseJSSchemas(relPath, src string) []models.CodeSurface {
	var surfaces []models.CodeSurface
	pkg := inferSurfacePackage(relPath)
	seen := map[string]bool{}

	add := func(name, reason string, line int, confidence float64, tier string) {
		sid := models.BuildSurfaceID(relPath, name+"_L"+strconv.Itoa(line), "")
		if seen[sid] {
			return
		}
		seen[sid] = true
		surfaces = append(surfaces, models.CodeSurface{
			SurfaceID:     sid,
			Name:          name,
			Path:          relPath,
			Kind:          models.SurfaceToolDef,
			Language:      "js",
			Package:       pkg,
			Line:          line,
			Exported:      false,
			DetectionTier: tier,
			Confidence:    confidence,
			Reason:        reason,
		})
	}

	lines := strings.Split(src, "\n")

	// 1. Zod schemas — structural detection.
	for i, l := range lines {
		if m := jsZodSchema.FindStringSubmatch(l); m != nil {
			varName := m[1]
			// Only classify as AI tool schema if the variable name suggests
			// tool/output/response/schema purpose (avoid generic validation schemas).
			lower := strings.ToLower(varName)
			if containsAny(lower, "tool", "output", "response", "schema", "param", "arg", "input") {
				add("zod_"+varName, "Zod schema '"+varName+"' (structured validation for AI tool/output)", i+1, 0.88, models.TierSemantic)
			}
		}
	}

	// 2. OpenAI tool registration.
	if loc := jsToolRegistration.FindStringIndex(src); loc != nil {
		line := 1 + strings.Count(src[:loc[0]], "\n")
		add("openai_tool_registration", "OpenAI function-calling tool registration (type: 'function')", line, 0.95, models.TierSemantic)
	}

	// 3. Output parsers.
	for i, l := range lines {
		if jsOutputParser.MatchString(l) {
			add("output_parser", "LangChain output parser instantiation", i+1, 0.90, models.TierSemantic)
			break
		}
	}

	// 4. tools: [...] array with tool registration context.
	if loc := jsToolsArray.FindStringIndex(src); loc != nil {
		// Verify it's in an AI context (near chat.completions.create or similar).
		context := src
		if loc[0] > 200 {
			context = src[loc[0]-200:]
		}
		if strings.Contains(context, "completions") || strings.Contains(context, "openai") ||
			strings.Contains(context, "anthropic") || strings.Contains(context, "chat") {
			line := 1 + strings.Count(src[:loc[0]], "\n")
			add("tool_array", "tools array in API call context", line, 0.85, models.TierSemantic)
		}
	}

	// 5. Structured outputs (response_format: json_schema).
	if loc := jsResponseFormat.FindStringIndex(src); loc != nil {
		line := 1 + strings.Count(src[:loc[0]], "\n")
		add("structured_output_format", "structured output format (json_schema response_format)", line, 0.92, models.TierSemantic)
	}

	return surfaces
}

// --- Python schema detection ---

var (
	// Pydantic BaseModel: class X(BaseModel):
	pyPydanticModel = regexp.MustCompile(`class\s+(\w+)\s*\(\s*(?:BaseModel|BaseSettings)\s*\)\s*:`)

	// Instructor response_model: response_model=X
	pyResponseModel = regexp.MustCompile(`response_model\s*=\s*(\w+)`)

	// OpenAI tools registration: tools=[{...}]
	pyToolsArray = regexp.MustCompile(`tools\s*=\s*\[`)

	// LangChain output parsers
	pyOutputParser = regexp.MustCompile(`(?:StructuredOutputParser|JsonOutputParser|PydanticOutputParser|OutputFixingParser)\s*[.(]`)

	// Function definition with tool decorator
	pyToolDecorator = regexp.MustCompile(`@(?:tool|function_tool|register_tool)\b`)
)

func parsePythonSchemas(relPath, src string) []models.CodeSurface {
	var surfaces []models.CodeSurface
	pkg := inferSurfacePackage(relPath)
	seen := map[string]bool{}

	add := func(name, reason string, line int, confidence float64, tier, kind string) {
		sid := models.BuildSurfaceID(relPath, name+"_L"+strconv.Itoa(line), "")
		if seen[sid] {
			return
		}
		seen[sid] = true
		surfaceKind := models.SurfaceToolDef
		if kind == "eval" {
			surfaceKind = models.SurfaceEvalDef
		}
		surfaces = append(surfaces, models.CodeSurface{
			SurfaceID:     sid,
			Name:          name,
			Path:          relPath,
			Kind:          surfaceKind,
			Language:      "python",
			Package:       pkg,
			Line:          line,
			Exported:      false,
			DetectionTier: tier,
			Confidence:    confidence,
			Reason:        reason,
		})
	}

	lines := strings.Split(src, "\n")

	// 1. Pydantic BaseModel classes.
	for i, l := range lines {
		if m := pyPydanticModel.FindStringSubmatch(l); m != nil {
			className := m[1]
			lower := strings.ToLower(className)
			// Classify based on name: Response/Output/Tool → tool_definition, else generic.
			if containsAny(lower, "response", "output", "tool", "schema", "result", "param", "arg") {
				add("pydantic_"+className, "Pydantic model '"+className+"' (structured output contract)", i+1, 0.90, models.TierSemantic, "tool")
			}
		}
	}

	// 2. Instructor response_model.
	for i, l := range lines {
		if m := pyResponseModel.FindStringSubmatch(l); m != nil {
			modelName := m[1]
			add("response_model_"+modelName, "Instructor response_model="+modelName, i+1, 0.92, models.TierSemantic, "tool")
			break
		}
	}

	// 3. Tool decorators.
	for i, l := range lines {
		if pyToolDecorator.MatchString(l) {
			// Get the function name from the next line.
			funcName := "decorated_tool"
			if i+1 < len(lines) {
				if m := regexp.MustCompile(`def\s+(\w+)`).FindStringSubmatch(lines[i+1]); m != nil {
					funcName = m[1]
				}
			}
			add("tool_"+funcName, "@tool decorator on function '"+funcName+"'", i+1, 0.95, models.TierSemantic, "tool")
		}
	}

	// 4. Output parsers.
	for i, l := range lines {
		if pyOutputParser.MatchString(l) {
			add("output_parser", "LangChain/output parser instantiation", i+1, 0.88, models.TierSemantic, "tool")
			break
		}
	}

	// 5. OpenAI tools array in API context.
	if loc := pyToolsArray.FindStringIndex(src); loc != nil {
		context := src
		if loc[0] > 200 {
			context = src[loc[0]-200:]
		}
		if strings.Contains(context, "completions") || strings.Contains(context, "openai") ||
			strings.Contains(context, "anthropic") || strings.Contains(context, "chat") {
			line := 1 + strings.Count(src[:loc[0]], "\n")
			add("tool_array", "tools array in API call context", line, 0.85, models.TierSemantic, "tool")
		}
	}

	return surfaces
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
