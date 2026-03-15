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
	testPaths := map[string]bool{}
	for _, tf := range testFiles {
		testPaths[tf.Path] = true
	}

	sourceFiles := collectSourceFiles(root)
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
	return surfaces
}

// --- JavaScript/TypeScript Surface Extractor ---

type jsSurfaceExtractor struct{}

func (e *jsSurfaceExtractor) Language() string { return "js" }

// Route registration patterns: app.get('/path', ...), router.post('/path', ...)
var (
	jsRoutePattern   = regexp.MustCompile(`\b(?:app|router|server)\.(get|post|put|patch|delete|all|use)\s*\(\s*['"]([^'"]+)['"]`)
	jsHandlerPattern = regexp.MustCompile(`export\s+(?:async\s+)?function\s+(\w*(?:[Hh]andler|[Mm]iddleware|[Cc]ontroller)\w*)\s*\(`)
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

	// Pass 3: Exported functions and classes (same patterns as code unit extraction).
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

// Python route/handler patterns.
var (
	pyRouteDecoratorPattern = regexp.MustCompile(`@(?:app|router|blueprint|bp)\.(route|get|post|put|patch|delete)\s*\(\s*['"]([^'"]+)['"]`)
	pyHandlerPattern        = regexp.MustCompile(`^def\s+(\w*(?:handler|view|endpoint|controller)\w*)\s*\(`)
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
