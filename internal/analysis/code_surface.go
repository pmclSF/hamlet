package analysis

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// inferSurfacePackage extracts a package identifier from a file path.
// Same logic as depgraph.inferPackage but local to avoid cross-package deps.
func inferSurfacePackage(filePath string) string {
	parts := strings.Split(filepath.ToSlash(filePath), "/")
	if len(parts) <= 1 {
		return ""
	}
	switch parts[0] {
	case "packages", "libs", "apps", "modules":
		if len(parts) > 1 {
			return parts[0] + "/" + parts[1]
		}
	}
	return parts[0]
}

// SurfaceExtractor infers CodeSurface nodes from a source file.
// Each language provides its own implementation.
type SurfaceExtractor interface {
	// Language returns the canonical language identifier.
	Language() string

	// ExtractSurfaces reads a source file and returns inferred behavior anchors.
	ExtractSurfaces(root, relPath string) []models.CodeSurface
}

// surfaceRegistry maps language identifiers to their extractors.
var surfaceRegistry = map[string]SurfaceExtractor{}

func init() {
	registerSurfaceExtractor(&jsSurfaceExtractor{})
	registerSurfaceExtractor(&goSurfaceExtractor{})
	registerSurfaceExtractor(&pythonSurfaceExtractor{})
	registerSurfaceExtractor(&javaSurfaceExtractor{})
}

func registerSurfaceExtractor(e SurfaceExtractor) {
	surfaceRegistry[e.Language()] = e
}

// InferCodeSurfaces scans all source files in root and returns inferred
// behavior anchors. Test files are excluded. No configuration required.
func InferCodeSurfaces(root string, testFiles []models.TestFile) []models.CodeSurface {
	return InferCodeSurfacesFromList(root, testFiles, collectSourceFiles(root))
}

// InferCodeSurfacesFromList is like InferCodeSurfaces but uses a pre-collected
// file list to avoid redundant directory walks in large repos.
func InferCodeSurfacesFromList(root string, testFiles []models.TestFile, sourceFiles []string) []models.CodeSurface {
	testPaths := map[string]bool{}
	for _, tf := range testFiles {
		testPaths[tf.Path] = true
	}
	surfacesByFile := make([][]models.CodeSurface, len(sourceFiles))
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
		if e, exists := surfaceRegistry[lang]; exists {
			surfacesByFile[i] = e.ExtractSurfaces(root, relPath)
		}
	})

	var surfaces []models.CodeSurface
	for _, s := range surfacesByFile {
		surfaces = append(surfaces, s...)
	}

	// Assign detection tier and confidence based on surface kind and evidence.
	assignInferenceMetadata(surfaces)

	return surfaces
}

// assignInferenceMetadata sets DetectionTier and Confidence on surfaces
// based on the detection method used for each kind.
func assignInferenceMetadata(surfaces []models.CodeSurface) {
	for i := range surfaces {
		s := &surfaces[i]

		// Skip if already set (e.g., by content-based inference).
		if s.DetectionTier != "" {
			// Mark content/semantic detections that already have tier+confidence
			// as heuristic unless they explicitly set a basis.
			if s.ConfidenceBasis == "" {
				s.ConfidenceBasis = models.ConfidenceBasisHeuristic
			}
			continue
		}

		// Default basis is heuristic — overridden below for calibrated cases.
		basis := models.ConfidenceBasisHeuristic

		switch s.Kind {
		case models.SurfaceRoute:
			s.DetectionTier = models.TierPattern
			s.Confidence = 0.95
		case models.SurfaceHandler:
			s.DetectionTier = models.TierPattern
			s.Confidence = 0.9
		case models.SurfacePrompt, models.SurfaceContext:
			s.DetectionTier = models.TierPattern
			s.Confidence = 0.85
		case models.SurfaceDataset:
			s.DetectionTier = models.TierPattern
			s.Confidence = 0.85
		case models.SurfaceToolDef:
			s.DetectionTier = models.TierPattern
			s.Confidence = 0.85
		case models.SurfaceRetrieval:
			s.DetectionTier = models.TierPattern
			s.Confidence = 0.8
		case models.SurfaceAgent:
			s.DetectionTier = models.TierPattern
			s.Confidence = 0.8
		case models.SurfaceEvalDef:
			s.DetectionTier = models.TierPattern
			s.Confidence = 0.8
		case models.SurfaceFunction:
			if s.Language == "go" {
				// Go capitalization rule: exported = starts with uppercase.
				// This is a language-level guarantee, not a heuristic.
				// FP rate is provably 0% for correctly parsed Go files.
				s.DetectionTier = models.TierStructural
				s.Confidence = 0.99
				basis = models.ConfidenceBasisCalibrated
			} else {
				s.DetectionTier = models.TierPattern
				s.Confidence = 0.9
			}
		case models.SurfaceMethod:
			if s.Language == "go" {
				s.DetectionTier = models.TierStructural
				s.Confidence = 0.99
				basis = models.ConfidenceBasisCalibrated
			} else {
				s.DetectionTier = models.TierPattern
				s.Confidence = 0.9
			}
		case models.SurfaceClass:
			s.DetectionTier = models.TierPattern
			s.Confidence = 0.9
		case models.SurfaceFixture:
			s.DetectionTier = models.TierPattern
			s.Confidence = 0.85
		default:
			s.DetectionTier = models.TierPattern
			s.Confidence = 0.7
		}

		s.ConfidenceBasis = basis
	}
}

// --- JavaScript/TypeScript Surface Extractor ---

type jsSurfaceExtractor struct{}

func (e *jsSurfaceExtractor) Language() string { return "js" }

// Route registration patterns: app.get('/path', ...), router.post('/path', ...)
var (
	jsRoutePattern   = regexp.MustCompile(`\b(?:app|router|server)\.(get|post|put|patch|delete|all|use)\s*\(\s*['"]([^'"]+)['"]`)
	jsHandlerPattern = regexp.MustCompile(`export\s+(?:async\s+)?function\s+(\w*(?:[Hh]andler|[Mm]iddleware|[Cc]ontroller)\w*)\s*\(`)
	// Prompt pattern: matches *Prompt*, PROMPT*, and *Template only when preceded
	// by an AI-related prefix (prompt, chat, system, user, completion, llm, ai, model).
	// This avoids false positives like emailTemplate, htmlTemplate, pageTemplate.
	jsPromptPattern = regexp.MustCompile(`export\s+(?:const|let|var|async\s+function|function)\s+(\w*(?:[Pp]rompt|PROMPT|(?:[Pp]rompt|[Cc]hat|[Ss]ystem|[Uu]ser|[Cc]ompletion|[Ll]lm|[Aa]i|[Mm]odel)[Tt]emplate)\w*)`)
	// Context pattern: matches AI behavioral contracts. Patterns are specific
	// to avoid false positives (e.g., "persona" requires AI prefix, "instruction"
	// requires "dynamic" or "system" prefix, "contextTemplate" requires AI prefix).
	jsContextPattern = regexp.MustCompile(`export\s+(?:const|let|var|async\s+function|function)\s+(\w*(?:[Ss]ystem[Mm]essage|[Ss]ystem[Pp]rompt|[Uu]ser[Mm]essage|[Ff]ew[Ss]hot|[Ss]afety[Oo]verlay|[Pp]olicy[Tt]ext|[Pp]olicy[Bb]lock|[Aa]i[Pp]ersona|[Ss]ystem[Pp]ersona|[Cc]hat[Pp]ersona|[Cc]ontext[Bb]uilder|[Aa]i[Cc]ontext|[Pp]rompt[Cc]ontext|[Cc]ustomer[Cc]ontext|[Aa]ccount[Cc]ontext|[Dd]ynamic[Ii]nstruction|[Ss]ystem[Ii]nstruction)\w*)`)
	jsDatasetPattern = regexp.MustCompile(`export\s+(?:const|let|var|async\s+function|function)\s+(\w*(?:[Dd]ataset|[Dd]ataloader|[Tt]raining[Dd]ata|[Ee]val[Dd]ata|[Ff]ixtures?Data)\w*)`)

	// AI tool/function-calling patterns: schemas, descriptions, routing, parsing,
	// guardrails, availability, budgets, retry/fallback for tool calls.
	// outputSchema requires "llm", "ai", "tool", or "function" prefix to avoid
	// matching REST API response schemas.
	jsToolDefPattern = regexp.MustCompile(`export\s+(?:const|let|var|async\s+function|function)\s+(\w*(?:[Tt]ool[Ss]chema|[Tt]ool[Dd]ef|[Ff]unction[Ss]chema|[Ff]unction[Dd]ef|[Tt]ool[Dd]escription|(?:[Ll]lm|[Aa]i|[Tt]ool|[Ff]unction)[Oo]utput[Ss]chema|[Ss]tructured[Oo]utput|[Tt]ool[Rr]outing|[Pp]arser[Ss]chema|[Vv]alidator[Ss]chema|[Tt]ool[Gg]uardrail|[Tt]ool[Aa]vailabilit|[Tt]ool[Bb]udget|[Tt]ool[Rr]etry|[Tt]ool[Ff]allback|[Tt]ool[Ff]ilter|[Tt]ool[Pp]ermission|[Aa]llowed[Tt]ool)\w*)`)

	// Retrieval/RAG patterns — covers all 12 RAG surface types:
	// retriever, query rewriting, chunking, doc splitting, embedding config,
	// vector store, retrieval filters, top-k/recall, reranker, citation,
	// context window assembly, source selection.
	jsRetrievalPattern = regexp.MustCompile(`export\s+(?:const|let|var|async\s+function|function)\s+(\w*(?:[Rr]etriever|[Rr]etrieval|[Vv]ector[Ss]tore|[Ee]mbedding[A-Z]|[Ee]mbedding[Cc]onfig|[Cc]hunking|[Cc]hunk[Ss]ize|[Cc]hunk[Cc]onfig|[Cc]hunk[Oo]verlap|[Dd]oc[Ss]plit|[Dd]ocument[Ss]plit|[Rr]eranker|[Rr]erank[Cc]onfig|[Dd]ocument[Ll]oader|[Cc]ontext[Aa]ssembl|[Cc]ontext[Ww]indow|[Rr]agConfig|[Rr]ag[Pp]ipeline|[Ss]earch[Qq]uery|[Qq]uery[Rr]ewrit|[Qq]uery[Bb]uild|[Ii]ndex[Bb]uilder|[Cc]itation|[Tt]op[Kk]|[Rr]ecall[Ss]etting|[Rr]etrieval[Ff]ilter|[Ss]ource[Ss]elect)\w*)`)

	// Agent/orchestration patterns: routers, planners, tool choice, budgets,
	// memory, handoff, retry, fallback, guardrails, step limits.
	jsAgentPattern = regexp.MustCompile(`export\s+(?:const|let|var|async\s+function|function)\s+(\w*(?:[Aa]gent[Rr]outer|[Aa]gent[Pp]lanner|[Tt]ool[Cc]hoice|[Ss]tep[Bb]udget|[Ss]tep[Ll]imit|[Mm]emory[Ww]indow|[Hh]andoff|[Ff]allback[Mm]odel|[Ff]allback[Pp]rovider|[Ff]allback[Ss]trategy|[Oo]rchestrat|[Pp]lanner[Cc]onfig|[Rr]etry[Pp]olicy|[Rr]etry[Cc]onfig|[Mm]ax[Rr]ecursion|[Mm]ax[Ss]teps|[Aa]gent[Cc]onfig|[Aa]gent[Gg]uardrail|[Ee]xecution[Bb]udget)\w*)`)

	// Eval definition patterns.
	jsEvalDefPattern = regexp.MustCompile(`export\s+(?:const|let|var|async\s+function|function)\s+(\w*(?:[Rr]ubric|[Ee]val[Mm]etric|[Ee]val[Cc]onfig|[Bb]aseline[Ss]chema|[Ee]xpected[Oo]utput|[Gg]rading[Cc]riteria|[Ss]coring[Ff]unc|[Ee]val[Ss]uite|[Ee]val[Rr]unner)\w*)`)
)

func (e *jsSurfaceExtractor) ExtractSurfaces(root, relPath string) []models.CodeSurface {
	content, err := os.ReadFile(filepath.Join(root, relPath))
	if err != nil {
		return nil
	}
	src := string(content)
	lines := strings.Split(src, "\n")
	pkg := inferSurfacePackage(relPath)

	var surfaces []models.CodeSurface
	seen := map[string]bool{}

	add := func(s models.CodeSurface) {
		if seen[s.SurfaceID] {
			return
		}
		seen[s.SurfaceID] = true
		surfaces = append(surfaces, s)
	}

	// Pass 1: Detect route registrations.
	for i, line := range lines {
		if m := jsRoutePattern.FindStringSubmatch(line); m != nil {
			method := strings.ToUpper(m[1])
			route := m[2]
			name := method + " " + route
			add(models.CodeSurface{
				SurfaceID:  models.BuildSurfaceID(relPath, name, ""),
				Name:       name,
				Path:       relPath,
				Kind:       models.SurfaceRoute,
				Language:   "js",
				Package:    pkg,
				Line:       i + 1,
				Route:      route,
				HTTPMethod: method,
				Exported:   true,
			})
		}
	}

	// Pass 2: Detect handler/middleware functions.
	for i, line := range lines {
		if m := jsHandlerPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			add(models.CodeSurface{
				SurfaceID:      models.BuildSurfaceID(relPath, name, ""),
				Name:           name,
				Path:           relPath,
				Kind:           models.SurfaceHandler,
				Language:       "js",
				Package:        pkg,
				Line:           i + 1,
				Exported:       true,
				LinkedCodeUnit: buildUnitID(relPath, name, ""),
			})
		}
	}

	// Pass 3a: Detect context surfaces (system messages, policy, few-shot, safety).
	// Checked before prompts so context-specific names like systemPrompt are
	// classified as context, not generic prompt.
	for i, line := range lines {
		if m := jsContextPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			sid := models.BuildSurfaceID(relPath, name, "")
			if seen[sid] {
				continue
			}
			add(models.CodeSurface{
				SurfaceID:      sid,
				Name:           name,
				Path:           relPath,
				Kind:           models.SurfaceContext,
				Language:       "js",
				Package:        pkg,
				Line:           i + 1,
				Exported:       true,
				LinkedCodeUnit: buildUnitID(relPath, name, ""),
			})
		}
	}

	// Pass 3b: Detect prompt definitions (templates, builders).
	for i, line := range lines {
		if m := jsPromptPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			sid := models.BuildSurfaceID(relPath, name, "")
			if seen[sid] {
				continue
			}
			add(models.CodeSurface{
				SurfaceID:      sid,
				Name:           name,
				Path:           relPath,
				Kind:           models.SurfacePrompt,
				Language:       "js",
				Package:        pkg,
				Line:           i + 1,
				Exported:       true,
				LinkedCodeUnit: buildUnitID(relPath, name, ""),
			})
		}
	}

	// Pass 4: Detect dataset definitions.
	for i, line := range lines {
		if m := jsDatasetPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			sid := models.BuildSurfaceID(relPath, name, "")
			if seen[sid] {
				continue
			}
			add(models.CodeSurface{
				SurfaceID:      sid,
				Name:           name,
				Path:           relPath,
				Kind:           models.SurfaceDataset,
				Language:       "js",
				Package:        pkg,
				Line:           i + 1,
				Exported:       true,
				LinkedCodeUnit: buildUnitID(relPath, name, ""),
			})
		}
	}

	// Pass 5: Tool definitions.
	for i, line := range lines {
		if m := jsToolDefPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			sid := models.BuildSurfaceID(relPath, name, "")
			if seen[sid] {
				continue
			}
			add(models.CodeSurface{
				SurfaceID: sid, Name: name, Path: relPath,
				Kind: models.SurfaceToolDef, Language: "js",
				Package: pkg, Line: i + 1, Exported: true,
				LinkedCodeUnit: buildUnitID(relPath, name, ""),
			})
		}
	}

	// Pass 6: Retrieval/RAG surfaces.
	for i, line := range lines {
		if m := jsRetrievalPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			sid := models.BuildSurfaceID(relPath, name, "")
			if seen[sid] {
				continue
			}
			add(models.CodeSurface{
				SurfaceID: sid, Name: name, Path: relPath,
				Kind: models.SurfaceRetrieval, Language: "js",
				Package: pkg, Line: i + 1, Exported: true,
				LinkedCodeUnit: buildUnitID(relPath, name, ""),
			})
		}
	}

	// Pass 7: Agent/orchestration surfaces.
	for i, line := range lines {
		if m := jsAgentPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			sid := models.BuildSurfaceID(relPath, name, "")
			if seen[sid] {
				continue
			}
			add(models.CodeSurface{
				SurfaceID: sid, Name: name, Path: relPath,
				Kind: models.SurfaceAgent, Language: "js",
				Package: pkg, Line: i + 1, Exported: true,
				LinkedCodeUnit: buildUnitID(relPath, name, ""),
			})
		}
	}

	// Pass 8: Eval definitions.
	for i, line := range lines {
		if m := jsEvalDefPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			sid := models.BuildSurfaceID(relPath, name, "")
			if seen[sid] {
				continue
			}
			add(models.CodeSurface{
				SurfaceID: sid, Name: name, Path: relPath,
				Kind: models.SurfaceEvalDef, Language: "js",
				Package: pkg, Line: i + 1, Exported: true,
				LinkedCodeUnit: buildUnitID(relPath, name, ""),
			})
		}
	}

	// Pass 9: Exported functions and classes (same patterns as code unit extraction).
	for i, line := range lines {
		if m := jsExportFuncPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			// Skip if already captured as handler.
			sid := models.BuildSurfaceID(relPath, name, "")
			if seen[sid] {
				continue
			}
			add(models.CodeSurface{
				SurfaceID:      sid,
				Name:           name,
				Path:           relPath,
				Kind:           models.SurfaceFunction,
				Language:       "js",
				Package:        pkg,
				Line:           i + 1,
				Exported:       true,
				LinkedCodeUnit: buildUnitID(relPath, name, ""),
			})
		}
		if m := jsExportDefaultFuncPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			sid := models.BuildSurfaceID(relPath, name, "")
			if seen[sid] {
				continue
			}
			add(models.CodeSurface{
				SurfaceID:      sid,
				Name:           name,
				Path:           relPath,
				Kind:           models.SurfaceFunction,
				Language:       "js",
				Package:        pkg,
				Line:           i + 1,
				Exported:       true,
				LinkedCodeUnit: buildUnitID(relPath, name, ""),
			})
		}
		if m := jsExportClassPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			add(models.CodeSurface{
				SurfaceID:      models.BuildSurfaceID(relPath, name, ""),
				Name:           name,
				Path:           relPath,
				Kind:           models.SurfaceClass,
				Language:       "js",
				Package:        pkg,
				Line:           i + 1,
				Exported:       true,
				LinkedCodeUnit: buildUnitID(relPath, name, ""),
			})
		}
		if m := jsExportDefaultClassPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			add(models.CodeSurface{
				SurfaceID:      models.BuildSurfaceID(relPath, name, ""),
				Name:           name,
				Path:           relPath,
				Kind:           models.SurfaceClass,
				Language:       "js",
				Package:        pkg,
				Line:           i + 1,
				Exported:       true,
				LinkedCodeUnit: buildUnitID(relPath, name, ""),
			})
		}
	}

	return surfaces
}

// --- Go Surface Extractor ---

type goSurfaceExtractor struct{}

func (e *goSurfaceExtractor) Language() string { return "go" }

// Go handler patterns: http.HandleFunc, mux.Handle, router.HandleFunc, etc.
var (
	goRoutePattern   = regexp.MustCompile(`\b(?:http\.HandleFunc|(?:mux|router|r|e|app)\.(?:HandleFunc|Handle|Get|Post|Put|Patch|Delete|Group|Route))\s*\(\s*"([^"]+)"`)
	goHandlerPattern = regexp.MustCompile(`^\s*func\s+([A-Z]\w*(?:Handler|Middleware|Controller)\w*)\s*\(`)
	goHTTPHandlerPattern = regexp.MustCompile(`^\s*func\s+\(\s*\w+\s+\*?\s*(\w+)\s*\)\s*([A-Z]\w*)\s*\(\s*\w+\s+http\.ResponseWriter`)
)

func (e *goSurfaceExtractor) ExtractSurfaces(root, relPath string) []models.CodeSurface {
	content, err := os.ReadFile(filepath.Join(root, relPath))
	if err != nil {
		return nil
	}
	src := string(content)
	lines := strings.Split(src, "\n")
	pkg := inferSurfacePackage(relPath)

	var surfaces []models.CodeSurface
	seen := map[string]bool{}

	add := func(s models.CodeSurface) {
		if seen[s.SurfaceID] {
			return
		}
		seen[s.SurfaceID] = true
		surfaces = append(surfaces, s)
	}

	// Pass 1: Route registrations.
	for i, line := range lines {
		if m := goRoutePattern.FindStringSubmatch(line); m != nil {
			route := m[1]
			add(models.CodeSurface{
				SurfaceID: models.BuildSurfaceID(relPath, route, ""),
				Name:      route,
				Path:      relPath,
				Kind:      models.SurfaceRoute,
				Language:  "go",
				Package:   pkg,
				Line:      i + 1,
				Route:     route,
				Exported:  true,
			})
		}
	}

	// Pass 2: Handler functions (named with Handler/Middleware suffix).
	for i, line := range lines {
		if m := goHandlerPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			add(models.CodeSurface{
				SurfaceID:      models.BuildSurfaceID(relPath, name, ""),
				Name:           name,
				Path:           relPath,
				Kind:           models.SurfaceHandler,
				Language:       "go",
				Package:        pkg,
				Line:           i + 1,
				Exported:       true,
				LinkedCodeUnit: buildUnitID(relPath, name, ""),
			})
		}
	}

	// Pass 3: Methods that accept http.ResponseWriter (HTTP handlers by signature).
	for i, line := range lines {
		if m := goHTTPHandlerPattern.FindStringSubmatch(line); m != nil {
			receiver := m[1]
			methodName := m[2]
			sid := models.BuildSurfaceID(relPath, methodName, receiver)
			if seen[sid] {
				continue
			}
			add(models.CodeSurface{
				SurfaceID:      sid,
				Name:           methodName,
				Path:           relPath,
				Kind:           models.SurfaceHandler,
				ParentName:     receiver,
				Language:       "go",
				Package:        pkg,
				Line:           i + 1,
				Receiver:       receiver,
				Exported:       true,
				LinkedCodeUnit: buildUnitID(relPath, methodName, receiver),
			})
		}
	}

	// Pass 4: Exported functions.
	for i, line := range lines {
		if m := goExportFuncPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			sid := models.BuildSurfaceID(relPath, name, "")
			if seen[sid] {
				continue
			}
			add(models.CodeSurface{
				SurfaceID:      sid,
				Name:           name,
				Path:           relPath,
				Kind:           models.SurfaceFunction,
				Language:       "go",
				Package:        pkg,
				Line:           i + 1,
				Exported:       true,
				LinkedCodeUnit: buildUnitID(relPath, name, ""),
			})
		}
	}

	// Pass 5: Exported methods.
	for i, line := range lines {
		if m := goExportMethodPattern.FindStringSubmatch(line); m != nil {
			receiver := m[1]
			methodName := m[2]
			sid := models.BuildSurfaceID(relPath, methodName, receiver)
			if seen[sid] {
				continue
			}
			add(models.CodeSurface{
				SurfaceID:      sid,
				Name:           methodName,
				Path:           relPath,
				Kind:           models.SurfaceMethod,
				ParentName:     receiver,
				Language:       "go",
				Package:        pkg,
				Line:           i + 1,
				Receiver:       receiver,
				Exported:       true,
				LinkedCodeUnit: buildUnitID(relPath, methodName, receiver),
			})
		}
	}

	return surfaces
}

// --- Python Surface Extractor ---

type pythonSurfaceExtractor struct{}

func (e *pythonSurfaceExtractor) Language() string { return "python" }

// Python route/handler/prompt/dataset patterns.
var (
	pyRouteDecoratorPattern = regexp.MustCompile(`@(?:app|router|blueprint|bp)\.(route|get|post|put|patch|delete)\s*\(\s*['"]([^'"]+)['"]`)
	pyHandlerPattern        = regexp.MustCompile(`^def\s+(\w*(?:handler|view|endpoint|controller)\w*)\s*\(`)
	// Python prompt: template requires AI prefix (prompt_template, chat_template, etc.).
	pyPromptPattern         = regexp.MustCompile(`^(?:def\s+(\w*(?:prompt|PROMPT|(?:prompt|chat|system|user|completion|llm|ai|model)_template)\w*)|(\w*(?:prompt|PROMPT|(?:prompt|chat|system|user|completion|llm|ai|model)_template)\w*)\s*=)`)
	// Python context: persona requires ai_ prefix, instruction requires dynamic_/system_.
	pyContextPattern        = regexp.MustCompile(`^(?:def\s+(\w*(?:system_message|system_prompt|user_message|few_shot|safety_overlay|policy_text|policy_block|ai_persona|system_persona|chat_persona|context_builder|ai_context|prompt_context|customer_context|account_context|dynamic_instruction|system_instruction)\w*)|(\w*(?:system_message|system_prompt|few_shot|safety_overlay|policy_text|policy_block|ai_persona|system_persona|context_builder|ai_context|prompt_context)\w*)\s*=)`)
	pyDatasetPattern        = regexp.MustCompile(`^(?:def\s+(\w*(?:dataset|dataloader|training_data|eval_data|load_data)\w*)|(\w*(?:dataset|dataloader|training_data|eval_data)\w*)\s*=)`)
	pyToolDefPattern        = regexp.MustCompile(`^(?:def\s+(\w*(?:tool_schema|tool_def|function_schema|function_def|tool_description|output_schema|structured_output|tool_routing|parser_schema|validator_schema|tool_guardrail|tool_availability|tool_budget|tool_retry|tool_fallback|tool_filter|tool_permission|allowed_tool)\w*)|(\w*(?:tool_schema|tool_def|function_schema|output_schema|structured_output|tool_guardrail|allowed_tool)\w*)\s*=)`)
	pyRetrievalPattern      = regexp.MustCompile(`^(?:def\s+(\w*(?:retriever|retrieval|vector_store|embedding_config|embedding_model|chunking|chunk_size|chunk_config|chunk_overlap|doc_split|document_split|reranker|rerank_config|document_loader|context_assembl|context_window|rag_config|rag_pipeline|search_query|query_rewrit|query_build|index_builder|citation|top_k|recall_setting|retrieval_filter|source_select)\w*)|(\w*(?:retriever|vector_store|embedding_config|chunk_config|chunk_size|rerank_config|rag_config|top_k|retrieval_filter)\w*)\s*=)`)
	pyAgentPattern          = regexp.MustCompile(`^(?:def\s+(\w*(?:agent_router|agent_planner|tool_choice|step_budget|step_limit|memory_window|handoff|fallback_model|fallback_provider|fallback_strategy|orchestrat|planner_config|retry_policy|retry_config|max_recursion|max_steps|agent_config|agent_guardrail|execution_budget)\w*)|(\w*(?:agent_config|planner_config|tool_choice|agent_guardrail|execution_budget)\w*)\s*=)`)
	pyEvalDefPattern        = regexp.MustCompile(`^(?:def\s+(\w*(?:rubric|eval_metric|eval_config|baseline_schema|expected_output|grading_criteria|scoring_func|eval_suite|eval_runner)\w*)|(\w*(?:rubric|eval_config|grading_criteria|scoring_func)\w*)\s*=)`)
)

func (e *pythonSurfaceExtractor) ExtractSurfaces(root, relPath string) []models.CodeSurface {
	content, err := os.ReadFile(filepath.Join(root, relPath))
	if err != nil {
		return nil
	}
	src := string(content)
	lines := strings.Split(src, "\n")
	allowed := pythonAllExports(src)
	pkg := inferSurfacePackage(relPath)

	var surfaces []models.CodeSurface
	seen := map[string]bool{}

	add := func(s models.CodeSurface) {
		if seen[s.SurfaceID] {
			return
		}
		seen[s.SurfaceID] = true
		surfaces = append(surfaces, s)
	}

	// Pass 1: Route decorators (@app.route('/path')).
	for i, line := range lines {
		if m := pyRouteDecoratorPattern.FindStringSubmatch(line); m != nil {
			method := strings.ToUpper(m[1])
			if method == "ROUTE" {
				method = ""
			}
			route := m[2]
			name := route
			if method != "" {
				name = method + " " + route
			}
			add(models.CodeSurface{
				SurfaceID:  models.BuildSurfaceID(relPath, name, ""),
				Name:       name,
				Path:       relPath,
				Kind:       models.SurfaceRoute,
				Language:   "python",
				Package:    pkg,
				Line:       i + 1,
				Route:      route,
				HTTPMethod: method,
				Exported:   true,
			})
		}
	}

	// Pass 2: Public functions.
	for i, line := range lines {
		if m := pyDefPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			if strings.HasPrefix(name, "_") {
				continue
			}
			if len(allowed) > 0 && !allowed[name] {
				continue
			}

			kind := models.SurfaceFunction
			if pyHandlerPattern.MatchString(line) {
				kind = models.SurfaceHandler
			} else if pyToolDefPattern.MatchString(line) {
				kind = models.SurfaceToolDef
			} else if pyRetrievalPattern.MatchString(line) {
				kind = models.SurfaceRetrieval
			} else if pyAgentPattern.MatchString(line) {
				kind = models.SurfaceAgent
			} else if pyEvalDefPattern.MatchString(line) {
				kind = models.SurfaceEvalDef
			} else if pyContextPattern.MatchString(line) {
				kind = models.SurfaceContext
			} else if pyPromptPattern.MatchString(line) {
				kind = models.SurfacePrompt
			} else if pyDatasetPattern.MatchString(line) {
				kind = models.SurfaceDataset
			}

			add(models.CodeSurface{
				SurfaceID:      models.BuildSurfaceID(relPath, name, ""),
				Name:           name,
				Path:           relPath,
				Kind:           kind,
				Language:       "python",
				Package:        pkg,
				Line:           i + 1,
				Exported:       true,
				LinkedCodeUnit: buildUnitID(relPath, name, ""),
			})
		}
	}

	return surfaces
}

// --- Java Surface Extractor ---

type javaSurfaceExtractor struct{}

func (e *javaSurfaceExtractor) Language() string { return "java" }

// Java handler/route patterns.
var (
	javaRequestMappingPattern = regexp.MustCompile(`@(?:RequestMapping|GetMapping|PostMapping|PutMapping|PatchMapping|DeleteMapping)\s*\(\s*(?:value\s*=\s*)?["']([^"']+)["']`)
	javaControllerPattern     = regexp.MustCompile(`\bpublic\s+(?:abstract\s+|final\s+)?class\s+(\w*(?:Controller|Resource|Endpoint|Handler)\w*)`)
)

func (e *javaSurfaceExtractor) ExtractSurfaces(root, relPath string) []models.CodeSurface {
	content, err := os.ReadFile(filepath.Join(root, relPath))
	if err != nil {
		return nil
	}
	lines := strings.Split(string(content), "\n")
	pkg := inferSurfacePackage(relPath)

	var surfaces []models.CodeSurface
	seen := map[string]bool{}

	add := func(s models.CodeSurface) {
		if seen[s.SurfaceID] {
			return
		}
		seen[s.SurfaceID] = true
		surfaces = append(surfaces, s)
	}

	// Track current class context.
	currentType := ""

	// Pass 1: Route annotations and controller classes.
	for i, line := range lines {
		// Detect controller classes.
		if m := javaControllerPattern.FindStringSubmatch(line); m != nil {
			currentType = m[1]
			add(models.CodeSurface{
				SurfaceID:      models.BuildSurfaceID(relPath, currentType, ""),
				Name:           currentType,
				Path:           relPath,
				Kind:           models.SurfaceClass,
				Language:       "java",
				Package:        pkg,
				Line:           i + 1,
				Exported:       true,
				LinkedCodeUnit: buildUnitID(relPath, currentType, ""),
			})
		}

		// Detect route annotations.
		if m := javaRequestMappingPattern.FindStringSubmatch(line); m != nil {
			route := m[1]
			add(models.CodeSurface{
				SurfaceID: models.BuildSurfaceID(relPath, route, currentType),
				Name:      route,
				Path:      relPath,
				Kind:      models.SurfaceRoute,
				ParentName: currentType,
				Language:  "java",
				Package:   pkg,
				Line:      i + 1,
				Route:     route,
				Exported:  true,
			})
		}
	}

	// Pass 2: Public types.
	for i, line := range lines {
		if m := javaExportTypePattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			sid := models.BuildSurfaceID(relPath, name, "")
			if seen[sid] {
				continue
			}
			add(models.CodeSurface{
				SurfaceID:      sid,
				Name:           name,
				Path:           relPath,
				Kind:           models.SurfaceClass,
				Language:       "java",
				Package:        pkg,
				Line:           i + 1,
				Exported:       true,
				LinkedCodeUnit: buildUnitID(relPath, name, ""),
			})
		}
	}

	// Pass 3: Public methods.
	currentType = ""
	for i, line := range lines {
		if m := javaExportTypePattern.FindStringSubmatch(line); m != nil {
			currentType = m[1]
		}
		if m := javaExportMethodPattern.FindStringSubmatch(line); m != nil {
			name := m[1]
			if name == currentType {
				continue // constructor
			}
			parent := currentType
			sid := models.BuildSurfaceID(relPath, name, parent)
			if seen[sid] {
				continue
			}
			add(models.CodeSurface{
				SurfaceID:      sid,
				Name:           name,
				Path:           relPath,
				Kind:           models.SurfaceMethod,
				ParentName:     parent,
				Language:       "java",
				Package:        pkg,
				Line:           i + 1,
				Exported:       true,
				LinkedCodeUnit: buildUnitID(relPath, name, parent),
			})
		}
	}

	return surfaces
}
