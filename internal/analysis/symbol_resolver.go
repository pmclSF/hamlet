package analysis

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// SymbolLink represents a resolved reference from a test file to a specific
// code unit (function, method, or class). This enables symbol-level coverage
// mapping rather than file-level.
type SymbolLink struct {
	// TestPath is the repository-relative test file path.
	TestPath string

	// CodeUnitID is the deterministic unit ID (path:name or path:parent.name).
	CodeUnitID string

	// Confidence is how certain the linkage is (0.0–1.0).
	// 1.0 = AST-verified call, 0.8 = name match in test content.
	Confidence float64
}

// ResolveSymbolLinks scans test files for references to specific code units
// from imported source files. Returns per-test symbol-level links.
//
// Strategy per language:
//   - Go: AST parsing of test files for function calls matching exported symbols
//   - JS/TS: regex extraction of imported names + call-site matching
//   - Python: regex extraction of imported names + call-site matching
func ResolveSymbolLinks(root string, testFiles []models.TestFile, codeUnits []models.CodeUnit, importGraph *ImportGraph) []SymbolLink {
	if importGraph == nil || len(importGraph.TestImports) == 0 || len(codeUnits) == 0 {
		return nil
	}

	// Index code units by file path for fast lookup.
	unitsByFile := map[string][]models.CodeUnit{}
	for _, cu := range codeUnits {
		unitsByFile[cu.Path] = append(unitsByFile[cu.Path], cu)
	}

	// Resolve in parallel.
	linksByFile := make([][]SymbolLink, len(testFiles))
	parallelForEachIndex(len(testFiles), func(i int) {
		tf := testFiles[i]
		imports := importGraph.TestImports[tf.Path]
		if len(imports) == 0 {
			return
		}

		// Collect candidate code units from all imported files.
		var candidates []models.CodeUnit
		for srcPath := range imports {
			candidates = append(candidates, unitsByFile[srcPath]...)
		}
		if len(candidates) == 0 {
			return
		}

		ext := strings.ToLower(filepath.Ext(tf.Path))
		switch {
		case ext == ".go":
			linksByFile[i] = resolveGoSymbolLinks(root, tf.Path, candidates)
		case isJSExt(ext):
			linksByFile[i] = resolveJSSymbolLinks(root, tf.Path, candidates)
		case ext == ".py":
			linksByFile[i] = resolvePythonSymbolLinks(root, tf.Path, candidates)
		}
	})

	var allLinks []SymbolLink
	for _, links := range linksByFile {
		allLinks = append(allLinks, links...)
	}
	return allLinks
}

// --- Go symbol resolution (AST-based) ---

func resolveGoSymbolLinks(root, testPath string, candidates []models.CodeUnit) []SymbolLink {
	absPath := filepath.Join(root, testPath)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, absPath, nil, 0)
	if err != nil {
		// Fallback to regex if AST fails.
		return resolveByNameMatch(root, testPath, candidates)
	}

	// Collect all identifiers used as call targets or selector expressions.
	calledNames := map[string]bool{}
	referencedNames := map[string]bool{}

	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			switch fn := node.Fun.(type) {
			case *ast.Ident:
				// Direct function call: FuncName(...)
				calledNames[fn.Name] = true
			case *ast.SelectorExpr:
				// Method or qualified call: obj.Method(...) or pkg.Func(...)
				calledNames[fn.Sel.Name] = true
				if ident, ok := fn.X.(*ast.Ident); ok {
					// pkg.Func pattern — also record the full qualified name.
					calledNames[ident.Name+"."+fn.Sel.Name] = true
				}
			}
		case *ast.SelectorExpr:
			// Any selector access (field reads, type assertions, etc.).
			referencedNames[node.Sel.Name] = true
		case *ast.Ident:
			// Bare identifier references (type names, variable usage).
			referencedNames[node.Name] = true
		}
		return true
	})

	var links []SymbolLink
	seen := map[string]bool{}

	for _, cu := range candidates {
		uid := unitID(cu)
		if seen[uid] {
			continue
		}

		// Check direct call (highest confidence).
		if calledNames[cu.Name] {
			seen[uid] = true
			links = append(links, SymbolLink{
				TestPath:   testPath,
				CodeUnitID: uid,
				Confidence: 1.0,
			})
			continue
		}

		// Check method call: Parent.Method.
		if cu.ParentName != "" && calledNames[cu.ParentName+"."+cu.Name] {
			seen[uid] = true
			links = append(links, SymbolLink{
				TestPath:   testPath,
				CodeUnitID: uid,
				Confidence: 1.0,
			})
			continue
		}

		// Check type/class reference (used in construction or type assertion).
		if cu.Kind == models.CodeUnitKindClass && referencedNames[cu.Name] {
			seen[uid] = true
			links = append(links, SymbolLink{
				TestPath:   testPath,
				CodeUnitID: uid,
				Confidence: 0.9,
			})
			continue
		}

		// Check general reference (field access, variable usage).
		if referencedNames[cu.Name] {
			seen[uid] = true
			links = append(links, SymbolLink{
				TestPath:   testPath,
				CodeUnitID: uid,
				Confidence: 0.8,
			})
		}
	}

	return links
}

// --- JS/TS symbol resolution (regex-based) ---

var (
	// Match named imports: import { Foo, bar as baz } from '...'
	jsNamedImportPattern = regexp.MustCompile(`import\s*\{([^}]+)\}\s*from\s*['"]`)
	// Match default import: import Foo from '...'
	jsDefaultImportPattern = regexp.MustCompile(`import\s+(\w+)\s+from\s*['"]`)
	// Match require destructure: const { Foo, bar } = require('...')
	jsRequireDestructPattern = regexp.MustCompile(`(?:const|let|var)\s*\{([^}]+)\}\s*=\s*require\s*\(`)
)

func resolveJSSymbolLinks(root, testPath string, candidates []models.CodeUnit) []SymbolLink {
	content, err := os.ReadFile(filepath.Join(root, testPath))
	if err != nil {
		return nil
	}
	src := string(content)

	// Extract imported symbol names from the test file.
	importedNames := extractJSImportedNames(src)

	// Find which candidate code units are referenced by name in the test content.
	return matchCandidatesByName(testPath, src, candidates, importedNames)
}

func extractJSImportedNames(src string) map[string]bool {
	names := map[string]bool{}

	for _, m := range jsNamedImportPattern.FindAllStringSubmatch(src, -1) {
		for _, item := range strings.Split(m[1], ",") {
			name := parseImportName(item)
			if name != "" {
				names[name] = true
			}
		}
	}
	for _, m := range jsDefaultImportPattern.FindAllStringSubmatch(src, -1) {
		names[m[1]] = true
	}
	for _, m := range jsRequireDestructPattern.FindAllStringSubmatch(src, -1) {
		for _, item := range strings.Split(m[1], ",") {
			name := parseImportName(item)
			if name != "" {
				names[name] = true
			}
		}
	}

	return names
}

func parseImportName(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	// Handle "foo as bar" → bar (the local name).
	parts := strings.Split(trimmed, " as ")
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return strings.TrimSpace(parts[0])
}

// --- Python symbol resolution (regex-based) ---

var (
	// from module import name1, name2
	pyFromImportPattern = regexp.MustCompile(`from\s+\S+\s+import\s+(.+)`)
	// import module  → not useful for symbol-level, skip
)

func resolvePythonSymbolLinks(root, testPath string, candidates []models.CodeUnit) []SymbolLink {
	content, err := os.ReadFile(filepath.Join(root, testPath))
	if err != nil {
		return nil
	}
	src := string(content)

	importedNames := extractPythonImportedNames(src)
	return matchCandidatesByName(testPath, src, candidates, importedNames)
}

func extractPythonImportedNames(src string) map[string]bool {
	names := map[string]bool{}

	for _, m := range pyFromImportPattern.FindAllStringSubmatch(src, -1) {
		importList := m[1]
		// Handle multiline with backslash continuation.
		importList = strings.ReplaceAll(importList, "\\", "")
		// Handle parenthesized imports: from x import (a, b)
		importList = strings.Trim(importList, "()")

		for _, item := range strings.Split(importList, ",") {
			trimmed := strings.TrimSpace(item)
			if trimmed == "" {
				continue
			}
			// Handle "name as alias" → alias
			parts := strings.Fields(trimmed)
			if len(parts) >= 3 && parts[1] == "as" {
				names[parts[2]] = true
			} else {
				names[parts[0]] = true
			}
		}
	}

	return names
}

// --- Common matching logic ---

// matchCandidatesByName checks which candidate code units are referenced in the
// test source, using imported names for high-confidence and bare name occurrence
// for medium-confidence matching.
func matchCandidatesByName(testPath, src string, candidates []models.CodeUnit, importedNames map[string]bool) []SymbolLink {
	var links []SymbolLink
	seen := map[string]bool{}

	for _, cu := range candidates {
		uid := unitID(cu)
		if seen[uid] {
			continue
		}

		// High confidence: name appears in import list AND in test body.
		if importedNames[cu.Name] && containsWordBoundary(src, cu.Name) {
			seen[uid] = true
			links = append(links, SymbolLink{
				TestPath:   testPath,
				CodeUnitID: uid,
				Confidence: 0.95,
			})
			continue
		}

		// Medium confidence: name appears as a call site in test body.
		// We look for name( pattern to identify function calls.
		if containsCallSite(src, cu.Name) {
			seen[uid] = true
			links = append(links, SymbolLink{
				TestPath:   testPath,
				CodeUnitID: uid,
				Confidence: 0.8,
			})
			continue
		}

		// Lower confidence: method call via instance.method().
		if cu.ParentName != "" && containsCallSite(src, cu.Name) {
			seen[uid] = true
			links = append(links, SymbolLink{
				TestPath:   testPath,
				CodeUnitID: uid,
				Confidence: 0.7,
			})
		}
	}

	return links
}

// resolveByNameMatch is a fallback when AST parsing fails.
func resolveByNameMatch(root, testPath string, candidates []models.CodeUnit) []SymbolLink {
	content, err := os.ReadFile(filepath.Join(root, testPath))
	if err != nil {
		return nil
	}
	return matchCandidatesByName(testPath, string(content), candidates, nil)
}

// containsWordBoundary checks if src contains name as a whole word.
var wordBoundaryCache = map[string]*regexp.Regexp{}

func containsWordBoundary(src, name string) bool {
	if name == "" || len(name) < 2 {
		return false
	}
	// Simple check: look for the name surrounded by non-alphanumeric chars.
	idx := strings.Index(src, name)
	for idx >= 0 {
		before := idx == 0 || !isAlphaNumeric(src[idx-1])
		after := idx+len(name) >= len(src) || !isAlphaNumeric(src[idx+len(name)])
		if before && after {
			return true
		}
		next := strings.Index(src[idx+1:], name)
		if next < 0 {
			break
		}
		idx = idx + 1 + next
	}
	return false
}

// containsCallSite checks if src contains "name(" as a function call.
func containsCallSite(src, name string) bool {
	if name == "" {
		return false
	}
	pattern := name + "("
	idx := strings.Index(src, pattern)
	for idx >= 0 {
		before := idx == 0 || !isAlphaNumeric(src[idx-1])
		if before {
			return true
		}
		next := strings.Index(src[idx+1:], pattern)
		if next < 0 {
			break
		}
		idx = idx + 1 + next
	}
	return false
}

func isAlphaNumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

func unitID(cu models.CodeUnit) string {
	if cu.UnitID != "" {
		return cu.UnitID
	}
	return buildUnitID(cu.Path, cu.Name, cu.ParentName)
}

// PopulateSymbolLinks resolves symbol-level links and populates
// TestFile.LinkedCodeUnits with precise symbol references instead of
// file-level blanket linkage. Falls back to file-level for test files
// where no symbol-level references could be resolved.
func PopulateSymbolLinks(root string, testFiles []models.TestFile, codeUnits []models.CodeUnit, importGraph *ImportGraph) {
	if importGraph == nil || len(importGraph.TestImports) == 0 || len(codeUnits) == 0 {
		return
	}

	// Resolve symbol-level links.
	links := ResolveSymbolLinks(root, testFiles, codeUnits, importGraph)

	// Group links by test path.
	linksByTest := map[string][]SymbolLink{}
	for _, link := range links {
		linksByTest[link.TestPath] = append(linksByTest[link.TestPath], link)
	}

	// Build file-level fallback: all units from imported files.
	unitsByFile := map[string][]models.CodeUnit{}
	for _, cu := range codeUnits {
		unitsByFile[cu.Path] = append(unitsByFile[cu.Path], cu)
	}

	for i := range testFiles {
		tf := &testFiles[i]
		symbolLinks := linksByTest[tf.Path]

		if len(symbolLinks) > 0 {
			// Use symbol-level links (more precise).
			seen := map[string]bool{}
			linked := make([]string, 0, len(symbolLinks))
			for _, sl := range symbolLinks {
				if !seen[sl.CodeUnitID] {
					seen[sl.CodeUnitID] = true
					linked = append(linked, sl.CodeUnitID)
				}
			}
			sort.Strings(linked)
			tf.LinkedCodeUnits = linked
		} else {
			// Fallback: file-level linkage (all units from imported files).
			imports := importGraph.TestImports[tf.Path]
			if len(imports) == 0 {
				continue
			}
			seen := map[string]bool{}
			var linked []string
			for srcPath := range imports {
				for _, cu := range unitsByFile[srcPath] {
					id := unitID(cu)
					if id != "" && !seen[id] {
						seen[id] = true
						linked = append(linked, id)
					}
				}
			}
			sort.Strings(linked)
			tf.LinkedCodeUnits = linked
		}
	}
}
