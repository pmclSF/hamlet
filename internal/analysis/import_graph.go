package analysis

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
	"gopkg.in/yaml.v3"
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
//   - Python: relative and absolute imports (from .x import, from pkg.x import, import pkg.x)
//   - Go: test files implicitly test their own package (same directory)
func BuildImportGraph(root string, testFiles []models.TestFile) *ImportGraph {
	graph := &ImportGraph{
		TestImports: map[string]map[string]bool{},
	}
	if len(testFiles) == 0 {
		return graph
	}
	resolver := newJSImportResolver(root)

	type entry struct {
		testPath string
		imports  map[string]bool
	}
	entries := make([]entry, len(testFiles))
	parallelForEachIndex(len(testFiles), func(i int) {
		tf := testFiles[i]
		ext := strings.ToLower(filepath.Ext(tf.Path))
		var imports map[string]bool

		switch {
		case isJSExt(ext):
			imports = extractJSImports(root, tf.Path, resolver)
		case ext == ".py":
			imports = extractPythonImports(root, tf.Path)
		case ext == ".go":
			imports = extractGoImports(root, tf.Path)
		}

		if len(imports) > 0 {
			entries[i] = entry{
				testPath: tf.Path,
				imports:  imports,
			}
		}
	})
	for _, e := range entries {
		if len(e.imports) > 0 {
			graph.TestImports[e.testPath] = e.imports
		}
	}

	return graph
}

// JS/TS import patterns.
var (
	// import ... from './foo' or from "../foo"
	// (?s) allows multiline import lists:
	// import { a,\n b } from './mod'
	jsImportFromPattern = regexp.MustCompile(`(?s)(?:import|export)\s+.*?\s+from\s+['"]([^'"]+)['"]`)
	// require('./foo') or require("../foo")
	jsRequirePattern = regexp.MustCompile(`require\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	// dynamic import('./foo')
	jsDynamicImportPattern = regexp.MustCompile(`import\s*\(\s*['"]([^'"]+)['"]\s*\)`)
)

// extractJSImports extracts relative import paths from a JS/TS test file.
func extractJSImports(root, relPath string, resolver *jsImportResolver) map[string]bool {
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
			resolved := resolver.resolveImport(testDir, importPath)
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

type jsImportResolver struct {
	root              string
	aliasRule         []pathAlias
	packageAliasRule  []pathAlias
	workspacePackages map[string]string
}

type pathAlias struct {
	keyPrefix    string
	keySuffix    string
	targetPrefix string
	targetSuffix string
	hasWildcard  bool
	targetHasExt bool
}

func newJSImportResolver(root string) *jsImportResolver {
	return &jsImportResolver{
		root:              root,
		aliasRule:         loadTSPathAliases(root),
		packageAliasRule:  loadPackageImportAliases(root),
		workspacePackages: loadWorkspacePackages(root),
	}
}

func (r *jsImportResolver) resolveImport(fromDir, importPath string) []string {
	// Relative imports.
	if strings.HasPrefix(importPath, ".") {
		return resolveJSImport(r.root, fromDir, importPath)
	}
	// Node built-ins or URLs are not repository source links.
	if strings.HasPrefix(importPath, "node:") || strings.Contains(importPath, "://") {
		return nil
	}
	// Non-relative alias imports via tsconfig paths.
	if resolved := r.resolveAlias(importPath); len(resolved) > 0 {
		return resolved
	}
	// Node.js package import aliases (package.json#imports, often "#internal/*").
	if resolved := r.resolvePackageAlias(importPath); len(resolved) > 0 {
		return resolved
	}
	// Monorepo workspace package aliases.
	return r.resolveWorkspacePackage(importPath)
}

func (r *jsImportResolver) resolveAlias(importPath string) []string {
	for _, alias := range r.aliasRule {
		suffix, ok := matchAlias(alias, importPath)
		if !ok {
			continue
		}

		candidate := filepath.ToSlash(filepath.Clean(alias.targetPrefix + suffix + alias.targetSuffix))
		if alias.targetHasExt {
			if fileExists(filepath.Join(r.root, candidate)) {
				return []string{candidate}
			}
			continue
		}
		if resolved := resolveFromRoot(r.root, candidate); len(resolved) > 0 {
			return resolved
		}
	}
	return nil
}

func (r *jsImportResolver) resolvePackageAlias(importPath string) []string {
	for _, alias := range r.packageAliasRule {
		suffix, ok := matchAlias(alias, importPath)
		if !ok {
			continue
		}

		candidate := filepath.ToSlash(filepath.Clean(alias.targetPrefix + suffix + alias.targetSuffix))
		if alias.targetHasExt {
			if fileExists(filepath.Join(r.root, candidate)) {
				return []string{candidate}
			}
			continue
		}
		if resolved := resolveFromRoot(r.root, candidate); len(resolved) > 0 {
			return resolved
		}
	}
	return nil
}

func matchAlias(alias pathAlias, importPath string) (suffix string, ok bool) {
	if alias.hasWildcard {
		if !strings.HasPrefix(importPath, alias.keyPrefix) {
			return "", false
		}
		if alias.keySuffix != "" && !strings.HasSuffix(importPath, alias.keySuffix) {
			return "", false
		}
		suffix = strings.TrimPrefix(importPath, alias.keyPrefix)
		if alias.keySuffix != "" {
			suffix = strings.TrimSuffix(suffix, alias.keySuffix)
		}
		return suffix, true
	}
	if importPath != alias.keyPrefix {
		return "", false
	}
	return "", true
}

func (r *jsImportResolver) resolveWorkspacePackage(importPath string) []string {
	for pkgName, pkgDir := range r.workspacePackages {
		if importPath != pkgName && !strings.HasPrefix(importPath, pkgName+"/") {
			continue
		}
		subpath := strings.TrimPrefix(importPath, pkgName)
		subpath = strings.TrimPrefix(subpath, "/")
		candidate := pkgDir
		if subpath != "" {
			candidate = filepath.ToSlash(filepath.Clean(pkgDir + "/" + subpath))
		}
		if resolved := resolveFromRoot(r.root, candidate); len(resolved) > 0 {
			return resolved
		}
	}
	return nil
}

func resolveFromRoot(root, pathNoExt string) []string {
	if filepath.Ext(pathNoExt) != "" {
		if fileExists(filepath.Join(root, pathNoExt)) {
			return []string{pathNoExt}
		}
		return nil
	}

	extensions := []string{".js", ".ts", ".jsx", ".tsx", ".mjs", ".mts"}
	for _, ext := range extensions {
		candidate := pathNoExt + ext
		if fileExists(filepath.Join(root, candidate)) {
			return []string{candidate}
		}
	}
	for _, ext := range extensions {
		candidate := pathNoExt + "/index" + ext
		if fileExists(filepath.Join(root, candidate)) {
			return []string{candidate}
		}
	}
	return nil
}

func loadTSPathAliases(root string) []pathAlias {
	path := filepath.Join(root, "tsconfig.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var cfg struct {
		CompilerOptions struct {
			BaseURL string              `json:"baseUrl"`
			Paths   map[string][]string `json:"paths"`
		} `json:"compilerOptions"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil
	}
	if len(cfg.CompilerOptions.Paths) == 0 {
		return nil
	}

	baseURL := cfg.CompilerOptions.BaseURL
	if baseURL == "" {
		baseURL = "."
	}
	baseURL = filepath.ToSlash(filepath.Clean(baseURL))

	var aliases []pathAlias
	for key, targets := range cfg.CompilerOptions.Paths {
		if len(targets) == 0 {
			continue
		}
		target := targets[0]
		if key == "" || target == "" {
			continue
		}
		keyWildcard := strings.HasSuffix(key, "/*")
		targetWildcard := strings.HasSuffix(target, "/*")
		keyPrefix := strings.TrimSuffix(key, "/*")
		targetPrefix := strings.TrimSuffix(target, "/*")

		joined := normalizeAliasPrefix(filepath.Join(baseURL, targetPrefix), keyWildcard && targetWildcard)
		aliases = append(aliases, pathAlias{
			keyPrefix:    keyPrefix,
			keySuffix:    "",
			targetPrefix: joined,
			targetSuffix: "",
			hasWildcard:  keyWildcard && targetWildcard,
			targetHasExt: filepath.Ext(joined) != "",
		})
	}
	return aliases
}

func loadPackageImportAliases(root string) []pathAlias {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return nil
	}

	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil
	}
	raw, ok := doc["imports"]
	if !ok {
		return nil
	}
	imports, ok := raw.(map[string]any)
	if !ok {
		return nil
	}

	var aliases []pathAlias
	for key, rawTarget := range imports {
		key = strings.TrimSpace(key)
		if key == "" || !strings.HasPrefix(key, "#") {
			continue
		}
		target := packageImportTarget(rawTarget)
		if target == "" || !strings.HasPrefix(target, "./") {
			continue
		}
		target = strings.TrimPrefix(target, "./")

		keyPrefix, keySuffix, keyWildcard, keyOK := splitWildcardPattern(key)
		targetPrefix, targetSuffix, targetWildcard, targetOK := splitWildcardPattern(target)
		if !keyOK || !targetOK {
			continue
		}
		if keyWildcard != targetWildcard {
			continue
		}

		joinedPrefix := normalizeAliasPrefix(targetPrefix, keyWildcard)
		aliases = append(aliases, pathAlias{
			keyPrefix:    keyPrefix,
			keySuffix:    keySuffix,
			targetPrefix: joinedPrefix,
			targetSuffix: targetSuffix,
			hasWildcard:  keyWildcard,
			targetHasExt: filepath.Ext(joinedPrefix+targetSuffix) != "",
		})
	}

	return aliases
}

func packageImportTarget(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case []any:
		for _, item := range t {
			if target := packageImportTarget(item); target != "" {
				return target
			}
		}
	case map[string]any:
		priority := []string{"default", "import", "require", "node"}
		for _, key := range priority {
			if child, ok := t[key]; ok {
				if target := packageImportTarget(child); target != "" {
					return target
				}
			}
		}
		keys := make([]string, 0, len(t))
		for key := range t {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if target := packageImportTarget(t[key]); target != "" {
				return target
			}
		}
	}
	return ""
}

func splitWildcardPattern(pattern string) (prefix string, suffix string, hasWildcard bool, ok bool) {
	if pattern == "" {
		return "", "", false, false
	}
	first := strings.Index(pattern, "*")
	if first == -1 {
		return pattern, "", false, true
	}
	if strings.LastIndex(pattern, "*") != first {
		return "", "", false, false
	}
	return pattern[:first], pattern[first+1:], true, true
}

func normalizeAliasPrefix(prefix string, wildcard bool) string {
	raw := filepath.ToSlash(prefix)
	normalized := filepath.ToSlash(filepath.Clean(prefix))
	if normalized == "." {
		normalized = ""
	}
	if wildcard && strings.HasSuffix(raw, "/") && normalized != "" && !strings.HasSuffix(normalized, "/") {
		normalized += "/"
	}
	return normalized
}

func loadWorkspacePackages(root string) map[string]string {
	patterns := workspacePatterns(root)
	if len(patterns) == 0 {
		return nil
	}

	out := map[string]string{}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(root, pattern))
		if err != nil {
			continue
		}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || !info.IsDir() {
				continue
			}
			pkgJSON := filepath.Join(match, "package.json")
			pkgData, err := os.ReadFile(pkgJSON)
			if err != nil {
				continue
			}
			var pkg struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal(pkgData, &pkg); err != nil || pkg.Name == "" {
				continue
			}
			rel, err := filepath.Rel(root, match)
			if err != nil {
				continue
			}
			out[pkg.Name] = filepath.ToSlash(rel)
		}
	}
	return out
}

func workspacePatterns(root string) []string {
	var patterns []string

	if data, err := os.ReadFile(filepath.Join(root, "package.json")); err == nil {
		patterns = append(patterns, parseWorkspacePatterns(data)...)
	}
	if data, err := os.ReadFile(filepath.Join(root, "pnpm-workspace.yaml")); err == nil {
		patterns = append(patterns, parsePNPMWorkspacePatterns(data)...)
	}
	if data, err := os.ReadFile(filepath.Join(root, "lerna.json")); err == nil {
		patterns = append(patterns, parseLernaWorkspacePatterns(data)...)
	}

	return dedupeStrings(patterns)
}

func parseWorkspacePatterns(packageJSON []byte) []string {
	var doc map[string]any
	if err := json.Unmarshal(packageJSON, &doc); err != nil {
		return nil
	}
	raw, ok := doc["workspaces"]
	if !ok {
		return nil
	}

	toStrings := func(v any) []string {
		arr, ok := v.([]any)
		if !ok {
			return nil
		}
		var out []string
		for _, x := range arr {
			if s, ok := x.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	}

	switch v := raw.(type) {
	case []any:
		return toStrings(v)
	case map[string]any:
		if pkgs, ok := v["packages"]; ok {
			return toStrings(pkgs)
		}
	}
	return nil
}

func parsePNPMWorkspacePatterns(doc []byte) []string {
	var ws struct {
		Packages []string `yaml:"packages"`
	}
	if err := yaml.Unmarshal(doc, &ws); err != nil {
		return nil
	}
	var out []string
	for _, p := range ws.Packages {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseLernaWorkspacePatterns(doc []byte) []string {
	var cfg struct {
		Packages []string `json:"packages"`
	}
	if err := json.Unmarshal(doc, &cfg); err != nil {
		return nil
	}
	var out []string
	for _, p := range cfg.Packages {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// Python import patterns.
var (
	pyRelativeFromImportPattern = regexp.MustCompile(`^\s*from\s+(\.+[\w.]*)\s+import\s+`)
	pyAbsoluteFromImportPattern = regexp.MustCompile(`^\s*from\s+([A-Za-z_]\w*(?:\.[A-Za-z_]\w*)*)\s+import\s+`)
	pyImportLinePattern         = regexp.MustCompile(`^\s*import\s+(.+)$`)
	pyImportAliasStripPattern   = regexp.MustCompile(`\s+as\s+\w+$`)
)

// extractPythonImports extracts relative imports from a Python test file.
func extractPythonImports(root, relPath string) map[string]bool {
	absPath := filepath.Join(root, relPath)
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil
	}

	imports := map[string]bool{}
	testDir := filepath.Dir(relPath)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if m := pyRelativeFromImportPattern.FindStringSubmatch(line); m != nil {
			module := strings.TrimSpace(m[1])
			dots := leadingDots(module)
			name := strings.TrimPrefix(module, strings.Repeat(".", dots))
			for _, p := range resolvePythonModule(root, testDir, name, dots) {
				imports[p] = true
			}
			continue
		}
		if m := pyAbsoluteFromImportPattern.FindStringSubmatch(line); m != nil {
			module := strings.TrimSpace(m[1])
			for _, p := range resolvePythonModule(root, testDir, module, 0) {
				imports[p] = true
			}
			continue
		}
		if m := pyImportLinePattern.FindStringSubmatch(line); m != nil {
			items := strings.Split(m[1], ",")
			for _, item := range items {
				module := strings.TrimSpace(item)
				module = pyImportAliasStripPattern.ReplaceAllString(module, "")
				if module == "" || strings.HasPrefix(module, ".") {
					continue
				}
				for _, p := range resolvePythonModule(root, testDir, module, 0) {
					imports[p] = true
				}
			}
		}
	}

	return imports
}

func resolvePythonModule(root, testDir, module string, dots int) []string {
	if module == "" {
		return nil
	}
	parts := strings.Split(module, ".")

	var bases []string
	if dots > 0 {
		base := testDir
		for i := 1; i < dots; i++ {
			base = filepath.Dir(base)
		}
		bases = append(bases, base)
	} else {
		bases = append(bases, ".", "src", "lib", "python")
	}

	var resolved []string
	for _, base := range bases {
		candidate := filepath.ToSlash(filepath.Join(base, filepath.Join(parts...)))
		pyFile := candidate + ".py"
		if fileExists(filepath.Join(root, pyFile)) {
			resolved = append(resolved, pyFile)
		}
		pkgInit := filepath.ToSlash(filepath.Join(candidate, "__init__.py"))
		if fileExists(filepath.Join(root, pkgInit)) {
			resolved = append(resolved, pkgInit)
		}
	}
	return dedupeStrings(resolved)
}

func leadingDots(s string) int {
	n := 0
	for _, ch := range s {
		if ch == '.' {
			n++
			continue
		}
		break
	}
	return n
}

func dedupeStrings(items []string) []string {
	if len(items) < 2 {
		return items
	}
	seen := map[string]bool{}
	var out []string
	for _, item := range items {
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}

// extractGoImports handles Go test files by linking to all source files in the same package.
func extractGoImports(root, relPath string) map[string]bool {
	testDir := filepath.Dir(relPath)
	imports := map[string]bool{}

	// Always link same-package source files.
	addGoSourceFiles(root, testDir, imports)

	// Also link local cross-package imports inside the same module.
	modulePath := loadGoModulePath(root)
	if modulePath != "" {
		absTestPath := filepath.Join(root, relPath)
		for _, imp := range parseGoImportPaths(absTestPath) {
			targetDir, ok := resolveGoImportDir(modulePath, imp)
			if !ok {
				continue
			}
			addGoSourceFiles(root, targetDir, imports)
		}
	}

	if len(imports) == 0 {
		return nil
	}
	return imports
}

func addGoSourceFiles(root, relDir string, imports map[string]bool) {
	absDir := filepath.Join(root, relDir)
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
			imports[filepath.ToSlash(filepath.Join(relDir, name))] = true
		}
	}
}

func parseGoImportPaths(absPath string) []string {
	fs := token.NewFileSet()
	f, err := parser.ParseFile(fs, absPath, nil, parser.ImportsOnly)
	if err != nil || f == nil {
		return nil
	}

	var out []string
	for _, spec := range f.Imports {
		if spec == nil || spec.Path == nil {
			continue
		}
		pathValue, err := strconv.Unquote(spec.Path.Value)
		if err != nil || pathValue == "" {
			continue
		}
		out = append(out, pathValue)
	}
	return dedupeStrings(out)
}

func loadGoModulePath(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

func resolveGoImportDir(modulePath, importPath string) (string, bool) {
	if modulePath == "" || importPath == "" {
		return "", false
	}
	if importPath == modulePath {
		return ".", true
	}
	prefix := modulePath + "/"
	if !strings.HasPrefix(importPath, prefix) {
		return "", false
	}
	rel := strings.TrimPrefix(importPath, prefix)
	rel = filepath.ToSlash(filepath.Clean(rel))
	if rel == "" || rel == "." || strings.HasPrefix(rel, "../") {
		return "", false
	}
	return rel, true
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
