package analysis

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ProjectContext holds project-level framework detection results.
// This provides a fallback for per-file detection: if a test file's
// framework cannot be determined from its content, the project-level
// default is used instead of "unknown".
type ProjectContext struct {
	// Frameworks maps language -> list of detected frameworks (sorted by confidence descending).
	Frameworks map[string][]ProjectFramework
}

// ProjectFramework represents a framework detected at the project level.
type ProjectFramework struct {
	Name       string
	Source     string  // "config-file", "dependency", "convention"
	Confidence float64 // 0.0–1.0
}

// DefaultFramework returns the highest-confidence framework for a language,
// or empty string if none.
func (pc *ProjectContext) DefaultFramework(language string) string {
	fws := pc.Frameworks[language]
	if len(fws) == 0 {
		return ""
	}
	return fws[0].Name
}

// DetectProjectFrameworks scans the repository root for config files and
// dependency manifests to determine which test frameworks the project uses.
//
// This is Layer 1 of the layered detection system:
//   - Config files (jest.config.*, vitest.config.*, etc.) → high confidence
//   - Dependency manifests (package.json devDependencies) → medium-high confidence
//   - Convention (go.mod presence for Go) → medium confidence
func DetectProjectFrameworks(root string) *ProjectContext {
	ctx := &ProjectContext{
		Frameworks: map[string][]ProjectFramework{},
	}

	detectJSProjectFrameworks(root, ctx)
	detectPythonProjectFrameworks(root, ctx)
	detectGoProjectFrameworks(root, ctx)
	detectJavaProjectFrameworks(root, ctx)

	// Sort each language's frameworks by confidence descending.
	for lang := range ctx.Frameworks {
		sort.Slice(ctx.Frameworks[lang], func(i, j int) bool {
			return ctx.Frameworks[lang][i].Confidence > ctx.Frameworks[lang][j].Confidence
		})
	}

	return ctx
}

// detectJSProjectFrameworks detects JS/TS test frameworks from config files
// and package.json devDependencies.
func detectJSProjectFrameworks(root string, ctx *ProjectContext) {
	seen := map[string]bool{}

	// Config file detection — highest confidence.
	configIndicators := []struct {
		pattern   string
		framework string
	}{
		{"jest.config.*", "jest"},
		{"jest.config.js", "jest"},
		{"jest.config.ts", "jest"},
		{"jest.config.mjs", "jest"},
		{"jest.config.cjs", "jest"},
		{"vitest.config.*", "vitest"},
		{"vitest.config.ts", "vitest"},
		{"vitest.config.js", "vitest"},
		{"vitest.config.mts", "vitest"},
		{"playwright.config.*", "playwright"},
		{"playwright.config.ts", "playwright"},
		{"playwright.config.js", "playwright"},
		{"cypress.config.*", "cypress"},
		{"cypress.config.js", "cypress"},
		{"cypress.config.ts", "cypress"},
		{"cypress.json", "cypress"},
		{".mocharc.yml", "mocha"},
		{".mocharc.yaml", "mocha"},
		{".mocharc.json", "mocha"},
		{".mocharc.js", "mocha"},
		{".mocharc.cjs", "mocha"},
	}

	for _, ci := range configIndicators {
		if seen[ci.framework] {
			continue
		}
		matches, _ := filepath.Glob(filepath.Join(root, ci.pattern))
		if len(matches) > 0 {
			seen[ci.framework] = true
			ctx.Frameworks["javascript"] = append(ctx.Frameworks["javascript"], ProjectFramework{
				Name:       ci.framework,
				Source:     "config-file",
				Confidence: 0.95,
			})
		}
	}

	// package.json devDependencies — medium-high confidence.
	pkgPath := filepath.Join(root, "package.json")
	devDeps := readDevDependencies(pkgPath)

	depIndicators := map[string]string{
		"jest":              "jest",
		"@jest/core":        "jest",
		"vitest":            "vitest",
		"@playwright/test":  "playwright",
		"cypress":           "cypress",
		"mocha":             "mocha",
		"jasmine":           "jasmine",
		"@testing-library/jest-dom": "jest",
		"ts-jest":           "jest",
		"puppeteer":         "puppeteer",
		"webdriverio":       "webdriverio",
		"@wdio/cli":         "webdriverio",
		"testcafe":          "testcafe",
	}

	for dep, framework := range depIndicators {
		if seen[framework] {
			continue
		}
		if _, ok := devDeps[dep]; ok {
			seen[framework] = true
			ctx.Frameworks["javascript"] = append(ctx.Frameworks["javascript"], ProjectFramework{
				Name:       framework,
				Source:     "dependency",
				Confidence: 0.85,
			})
		}
	}

	// Also check for node:test usage — look for it in package.json scripts
	// or test configuration pointing to node --test.
	pkgData := readPackageJSON(pkgPath)
	if scripts, ok := pkgData["scripts"].(map[string]interface{}); ok {
		for _, v := range scripts {
			if s, ok := v.(string); ok {
				if strings.Contains(s, "node --test") || strings.Contains(s, "node:test") {
					if !seen["node-test"] {
						seen["node-test"] = true
						ctx.Frameworks["javascript"] = append(ctx.Frameworks["javascript"], ProjectFramework{
							Name:       "node-test",
							Source:     "dependency",
							Confidence: 0.90,
						})
					}
				}
			}
		}
	}
}

// detectPythonProjectFrameworks detects Python test frameworks from config files.
func detectPythonProjectFrameworks(root string, ctx *ProjectContext) {
	seen := map[string]bool{}

	// pytest config indicators.
	pytestConfigs := []string{
		"pytest.ini",
		"pyproject.toml",
		"setup.cfg",
		"conftest.py",
	}
	for _, f := range pytestConfigs {
		p := filepath.Join(root, f)
		if _, err := os.Stat(p); err == nil {
			// For pyproject.toml, check for [tool.pytest] section.
			if f == "pyproject.toml" {
				content, err := os.ReadFile(p)
				if err != nil || !strings.Contains(string(content), "[tool.pytest") {
					continue
				}
			}
			if !seen["pytest"] {
				seen["pytest"] = true
				ctx.Frameworks["python"] = append(ctx.Frameworks["python"], ProjectFramework{
					Name:       "pytest",
					Source:     "config-file",
					Confidence: 0.90,
				})
			}
			break
		}
	}

	// unittest doesn't have config files — it's a stdlib module.
	// nose2 has setup.cfg sections but that's rare.
}

// detectGoProjectFrameworks sets go-testing as the default for Go repos.
func detectGoProjectFrameworks(root string, ctx *ProjectContext) {
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
		ctx.Frameworks["go"] = append(ctx.Frameworks["go"], ProjectFramework{
			Name:       "go-testing",
			Source:     "convention",
			Confidence: 0.99,
		})
	}
}

// detectJavaProjectFrameworks detects Java test frameworks from build files.
func detectJavaProjectFrameworks(root string, ctx *ProjectContext) {
	// Check pom.xml or build.gradle for junit/testng dependencies.
	for _, buildFile := range []string{"pom.xml", "build.gradle", "build.gradle.kts"} {
		p := filepath.Join(root, buildFile)
		content, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		s := string(content)
		if strings.Contains(s, "junit-jupiter") || strings.Contains(s, "org.junit.jupiter") {
			ctx.Frameworks["java"] = append(ctx.Frameworks["java"], ProjectFramework{
				Name:       "junit5",
				Source:     "dependency",
				Confidence: 0.90,
			})
			return
		}
		if strings.Contains(s, "junit") || strings.Contains(s, "org.junit") {
			ctx.Frameworks["java"] = append(ctx.Frameworks["java"], ProjectFramework{
				Name:       "junit4",
				Source:     "dependency",
				Confidence: 0.85,
			})
			return
		}
		if strings.Contains(s, "testng") {
			ctx.Frameworks["java"] = append(ctx.Frameworks["java"], ProjectFramework{
				Name:       "testng",
				Source:     "dependency",
				Confidence: 0.90,
			})
			return
		}
	}
}

// readDevDependencies extracts the devDependencies map from package.json.
func readDevDependencies(path string) map[string]interface{} {
	pkg := readPackageJSON(path)
	if deps, ok := pkg["devDependencies"].(map[string]interface{}); ok {
		return deps
	}
	return map[string]interface{}{}
}

// readPackageJSON reads and parses a package.json file.
func readPackageJSON(path string) map[string]interface{} {
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]interface{}{}
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return map[string]interface{}{}
	}
	return result
}
