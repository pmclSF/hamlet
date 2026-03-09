package analysis

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
)

// ImportGraph maps test file paths to the source module paths they import.
// All paths are repository-relative (matching TestFile.Path and CodeUnit.Path).
type ImportGraph struct {
	// TestImports maps test file relative path → set of imported source relative paths.
	TestImports map[string]map[string]bool
}

// ImportedModules returns the set of all source module paths imported by any test file.
func (g *ImportGraph) ImportedModules() map[string]bool {
	result := map[string]bool{}
	for _, imports := range g.TestImports {
		for mod := range imports {
			result[mod] = true
		}
	}
	return result
}

// BuildImportGraph scans test files for import/require statements and resolves
// them to repository-relative source file paths.
//
// This enables precise test-to-code linkage: instead of guessing from directory
// proximity or filename stems, we know which source modules each test actually imports.
//
// Supported patterns:
//   - JS/TS: import ... from './path', require('./path'), dynamic import('./path')
//   - Python: from .module import ..., from ..module import ...
//   - Go: test files implicitly test their own package (same directory)
func BuildImportGraph(root string, testFiles []models.TestFile) *ImportGraph {
	graph := &ImportGraph{
		TestImports: map[string]map[string]bool{},
	}

	for _, tf := range testFiles {
		ext := strings.ToLower(filepath.Ext(tf.Path))
		var imports map[string]bool

		switch {
		case isJSExt(ext):
			imports = extractJSImports(root, tf.Path)
		case ext == ".py":
			imports = extractPythonImports(root, tf.Path)
		case ext == ".go":
			imports = extractGoImports(root, tf.Path)
		}

		if len(imports) > 0 {
			graph.TestImports[tf.Path] = imports
		}
	}

	return graph
}

// JS/TS import patterns.
var (
	// import ... from './foo' or from "../foo"
	jsImportFromPattern = regexp.MustCompile(`(?:import|export)\s+.*?\s+from\s+['"](\.[^'"]+)['"]`)
	// require('./foo') or require("../foo")
	jsRequirePattern = regexp.MustCompile(`require\s*\(\s*['"](\.[^'"]+)['"]\s*\)`)
	// dynamic import('./foo')
	jsDynamicImportPattern = regexp.MustCompile(`import\s*\(\s*['"](\.[^'"]+)['"]\s*\)`)
)

// extractJSImports extracts relative import paths from a JS/TS test file.
func extractJSImports(root, relPath string) map[string]bool {
	absPath := filepath.Join(root, relPath)
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil
	}

	src := string(content)
	imports := map[string]bool{}
	testDir := filepath.Dir(relPath)

	// Collect all relative import paths.
	for _, pat := range []*regexp.Regexp{jsImportFromPattern, jsRequirePattern, jsDynamicImportPattern} {
		for _, match := range pat.FindAllStringSubmatch(src, -1) {
			if len(match) < 2 {
				continue
			}
			importPath := match[1]
			// Skip non-relative imports (node_modules, node: builtins).
			if !strings.HasPrefix(importPath, ".") {
				continue
			}
			resolved := resolveJSImport(root, testDir, importPath)
			for _, r := range resolved {
				imports[r] = true
			}
		}
	}

	return imports
}

// resolveJSImport resolves a relative import path to one or more source file paths.
// It tries common JS/TS extensions and index file conventions.
func resolveJSImport(root, fromDir, importPath string) []string {
	// Resolve relative to the importing file's directory.
	resolved := filepath.Join(fromDir, importPath)
	resolved = filepath.Clean(resolved)
	// Normalize to forward slashes for consistency.
	resolved = filepath.ToSlash(resolved)

	// If it already has an extension, check directly.
	if filepath.Ext(resolved) != "" {
		if fileExists(filepath.Join(root, resolved)) {
			return []string{resolved}
		}
		return nil
	}

	// Try common extensions.
	extensions := []string{".js", ".ts", ".jsx", ".tsx", ".mjs", ".mts"}
	for _, ext := range extensions {
		candidate := resolved + ext
		if fileExists(filepath.Join(root, candidate)) {
			return []string{candidate}
		}
	}

	// Try index files (directory import).
	for _, ext := range extensions {
		candidate := resolved + "/index" + ext
		if fileExists(filepath.Join(root, candidate)) {
			return []string{candidate}
		}
	}

	return nil
}

// Python relative import pattern: from .module import ... or from ..pkg.module import ...
var pyRelativeImportPattern = regexp.MustCompile(`from\s+(\.+\w[\w.]*)\s+import`)

// extractPythonImports extracts relative imports from a Python test file.
func extractPythonImports(root, relPath string) map[string]bool {
	absPath := filepath.Join(root, relPath)
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil
	}

	src := string(content)
	imports := map[string]bool{}
	testDir := filepath.Dir(relPath)

	for _, match := range pyRelativeImportPattern.FindAllStringSubmatch(src, -1) {
		if len(match) < 2 {
			continue
		}
		modPath := match[1]

		// Count leading dots for relative depth.
		dots := 0
		for _, c := range modPath {
			if c == '.' {
				dots++
			} else {
				break
			}
		}
		moduleName := modPath[dots:]

		// Resolve: each dot goes up one directory.
		base := testDir
		for i := 1; i < dots; i++ {
			base = filepath.Dir(base)
		}

		// Convert module.name to module/name.py
		parts := strings.Split(moduleName, ".")
		modFile := filepath.ToSlash(filepath.Join(base, filepath.Join(parts...))) + ".py"
		if fileExists(filepath.Join(root, modFile)) {
			imports[modFile] = true
			continue
		}
		// Try __init__.py for package imports.
		pkgInit := filepath.ToSlash(filepath.Join(base, filepath.Join(parts...), "__init__.py"))
		if fileExists(filepath.Join(root, pkgInit)) {
			imports[pkgInit] = true
		}
	}

	return imports
}

// extractGoImports handles Go test files by linking to all source files in the same package.
func extractGoImports(root, relPath string) map[string]bool {
	testDir := filepath.Dir(relPath)
	absDir := filepath.Join(root, testDir)

	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil
	}

	imports := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			imports[filepath.ToSlash(filepath.Join(testDir, name))] = true
		}
	}

	return imports
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
