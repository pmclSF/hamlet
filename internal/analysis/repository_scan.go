package analysis

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
)

// skipDirs lists directories that should never be traversed during scanning.
var skipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"dist":         true,
	"build":        true,
	"coverage":     true,
	".next":        true,
	".turbo":       true,
	".nuxt":        true,
	"vendor":       true,
	"__pycache__":  true,
	".pytest_cache": true,
	".mypy_cache":  true,
	".tox":         true,
	".venv":        true,
	"venv":         true,
	".idea":        true,
	".vscode":      true,
	".hamlet":      true,
	"target":       true,
}

// testFilePatterns matches common test file naming conventions.
// The function isTestFile uses these plus directory-based heuristics.

// discoverTestFiles walks the repository tree and returns test files found.
// When projectCtx is provided, it is used as a fallback for files whose
// framework cannot be determined from content alone.
func discoverTestFiles(root string, projectCtx ...*ProjectContext) ([]models.TestFile, error) {
	var ctx *ProjectContext
	if len(projectCtx) > 0 {
		ctx = projectCtx[0]
	}

	var testFiles []models.TestFile

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible paths
		}

		if d.IsDir() {
			base := filepath.Base(path)
			if skipDirs[base] {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		if isTestFile(relPath) {
			result := detectFrameworkWithContext(relPath, path, ctx)
			testFiles = append(testFiles, models.TestFile{
				Path:                relPath,
				Framework:           result.Framework,
				FrameworkConfidence: result.Confidence,
				FrameworkSource:     result.Source,
			})
		}

		return nil
	})

	return testFiles, err
}

// isTestFile determines whether a file path looks like a test file
// based on naming conventions.
//
// Supported heuristics:
//   - *.test.{js,jsx,ts,tsx,mjs,cjs} — JS/TS test files
//   - *.spec.{js,jsx,ts,tsx,mjs,cjs} — JS/TS spec files
//   - *_test.go — Go test files
//   - test_*.py, *_test.py — Python test files
//   - *Test.java — Java test files
//   - files under __tests__/ directories
func isTestFile(relPath string) bool {
	base := filepath.Base(relPath)
	ext := strings.ToLower(filepath.Ext(base))
	nameWithoutExt := strings.TrimSuffix(base, ext)

	// JS/TS test and spec files
	jsExts := map[string]bool{
		".js": true, ".jsx": true, ".ts": true, ".tsx": true,
		".mjs": true, ".cjs": true, ".mts": true, ".cts": true,
	}
	if jsExts[ext] {
		if strings.HasSuffix(nameWithoutExt, ".test") || strings.HasSuffix(nameWithoutExt, ".spec") {
			return true
		}
		// Files inside __tests__/ directories
		if strings.Contains(relPath, "__tests__") {
			return true
		}
		// Files inside test/ or tests/ directories (e.g. Express, Mocha projects)
		parts := strings.Split(filepath.ToSlash(relPath), "/")
		for _, p := range parts[:len(parts)-1] {
			if p == "test" || p == "tests" {
				return true
			}
		}
	}

	// Go test files
	if ext == ".go" && strings.HasSuffix(nameWithoutExt, "_test") {
		return true
	}

	// Python test files
	if ext == ".py" {
		if strings.HasPrefix(base, "test_") || strings.HasSuffix(nameWithoutExt, "_test") {
			return true
		}
		// Files inside tests/ or test/ directories at any level
		parts := strings.Split(filepath.ToSlash(relPath), "/")
		for _, p := range parts[:len(parts)-1] {
			if p == "tests" || p == "test" {
				return true
			}
		}
	}

	// Java test files
	if ext == ".java" && strings.HasSuffix(nameWithoutExt, "Test") {
		return true
	}

	return false
}
