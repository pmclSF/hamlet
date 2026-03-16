package aidetect

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// DeriveScenarios auto-generates AI/eval scenarios from detected code
// patterns. Users never need to write YAML — scenarios are inferred from:
//
//  1. Test files in eval directories that import prompt/model/dataset surfaces
//  2. Files importing AI framework libraries (langchain, openai, etc.)
//  3. Promptfoo config files (each test case becomes a scenario)
//  4. Code surfaces of kind SurfacePrompt/SurfaceDataset linked to test files
//
// Derived scenarios are merged with any manually-declared scenarios from
// .terrain/terrain.yaml (manual declarations take precedence on ID collision).
func DeriveScenarios(root string, detection *DetectResult, codeSurfaces []models.CodeSurface, testFiles []models.TestFile) []models.Scenario {
	var scenarios []models.Scenario

	// Build surface index by path for linking.
	surfacesByPath := map[string][]models.CodeSurface{}
	for _, cs := range codeSurfaces {
		surfacesByPath[cs.Path] = append(surfacesByPath[cs.Path], cs)
	}

	// Build prompt/dataset surface ID sets for linking.
	promptSurfaces := map[string]string{} // path:name → surfaceID
	datasetSurfaces := map[string]string{}
	for _, cs := range codeSurfaces {
		key := cs.Path + ":" + cs.Name
		switch cs.Kind {
		case models.SurfacePrompt:
			promptSurfaces[key] = cs.SurfaceID
		case models.SurfaceDataset:
			datasetSurfaces[key] = cs.SurfaceID
		}
	}

	// Strategy 1: Test files in eval directories that import AI surfaces.
	scenarios = append(scenarios, deriveFromEvalTests(root, testFiles, surfacesByPath, promptSurfaces, datasetSurfaces)...)

	// Strategy 2: Test files importing AI framework libraries.
	scenarios = append(scenarios, deriveFromAIImports(root, testFiles, detection, promptSurfaces, datasetSurfaces)...)

	// Strategy 3: Promptfoo config test cases.
	for _, cfg := range detection.EvalConfigs {
		scenarios = append(scenarios, deriveFromPromptfooConfig(root, cfg, promptSurfaces)...)
	}

	// Deduplicate by scenario ID.
	scenarios = deduplicateScenarios(scenarios)

	return scenarios
}

// deriveFromEvalTests creates scenarios from test files in eval-like directories.
func deriveFromEvalTests(root string, testFiles []models.TestFile, surfacesByPath map[string][]models.CodeSurface, prompts, datasets map[string]string) []models.Scenario {
	var scenarios []models.Scenario

	for _, tf := range testFiles {
		if !isEvalTestPath(tf.Path) {
			continue
		}

		// Read the test file to find which surfaces it imports.
		content, err := os.ReadFile(filepath.Join(root, tf.Path))
		if err != nil {
			continue
		}
		src := string(content)

		// Find imported source files.
		importedSurfaces := findImportedSurfaces(src, tf.Path, prompts, datasets)
		if len(importedSurfaces) == 0 {
			continue
		}

		// Classify scenario category from path and content.
		category := classifyScenarioCategory(tf.Path, src)

		scenarios = append(scenarios, models.Scenario{
			ScenarioID:        scenarioID(tf.Framework, tf.Path),
			Name:              scenarioNameFromPath(tf.Path),
			Category:          category,
			Path:              tf.Path,
			Framework:         tf.Framework,
			CoveredSurfaceIDs: importedSurfaces,
		})
	}

	return scenarios
}

// deriveFromAIImports creates scenarios from test files that import AI libraries.
func deriveFromAIImports(root string, testFiles []models.TestFile, detection *DetectResult, prompts, datasets map[string]string) []models.Scenario {
	if len(detection.Frameworks) == 0 {
		return nil
	}

	// Build set of framework import patterns.
	var importPatterns []string
	for _, fw := range detection.Frameworks {
		for _, sig := range KnownFrameworks {
			if sig.Name == fw.Name {
				importPatterns = append(importPatterns, sig.ImportPatterns...)
			}
		}
	}
	if len(importPatterns) == 0 {
		return nil
	}

	var scenarios []models.Scenario
	for _, tf := range testFiles {
		// Skip files already handled by eval directory detection.
		if isEvalTestPath(tf.Path) {
			continue
		}

		content, err := os.ReadFile(filepath.Join(root, tf.Path))
		if err != nil {
			continue
		}
		src := string(content)

		// Check if this test imports any AI framework.
		hasAIImport := false
		var detectedFramework string
		for _, pat := range importPatterns {
			if strings.Contains(src, pat) {
				hasAIImport = true
				// Find which framework this matches.
				for _, fw := range detection.Frameworks {
					for _, sig := range KnownFrameworks {
						if sig.Name == fw.Name {
							for _, sp := range sig.ImportPatterns {
								if sp == pat {
									detectedFramework = fw.Name
								}
							}
						}
					}
				}
				break
			}
		}
		if !hasAIImport {
			continue
		}

		// Find surfaces this test imports.
		surfaces := findImportedSurfaces(src, tf.Path, prompts, datasets)
		if len(surfaces) == 0 {
			// Even without specific surfaces, create a scenario for AI test files.
			surfaces = []string{} // empty but scenario still valid
		}

		category := classifyScenarioCategory(tf.Path, src)
		framework := detectedFramework
		if framework == "" {
			framework = tf.Framework
		}

		scenarios = append(scenarios, models.Scenario{
			ScenarioID:        scenarioID(framework, tf.Path),
			Name:              scenarioNameFromPath(tf.Path),
			Category:          category,
			Path:              tf.Path,
			Framework:         framework,
			CoveredSurfaceIDs: surfaces,
		})
	}

	return scenarios
}

var promptfooTestPattern = regexp.MustCompile(`(?m)^\s*-\s+(?:description|vars|assert)`)

// deriveFromPromptfooConfig creates scenarios from promptfoo config files.
func deriveFromPromptfooConfig(root, configPath string, prompts map[string]string) []models.Scenario {
	data, err := os.ReadFile(filepath.Join(root, configPath))
	if err != nil {
		return nil
	}
	content := string(data)

	// Extract test descriptions from YAML (simple pattern matching).
	var scenarios []models.Scenario
	lines := strings.Split(content, "\n")
	testIdx := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "description:") {
			desc := strings.TrimSpace(strings.TrimPrefix(trimmed, "description:"))
			desc = strings.Trim(desc, "\"'")
			if desc == "" {
				continue
			}
			testIdx++

			// Link to all known prompt surfaces.
			var surfaces []string
			for _, sid := range prompts {
				surfaces = append(surfaces, sid)
			}
			sort.Strings(surfaces)

			scenarios = append(scenarios, models.Scenario{
				ScenarioID:        fmt.Sprintf("scenario:promptfoo:%s:%d", configPath, testIdx),
				Name:              desc,
				Category:          "eval",
				Path:              configPath,
				Framework:         "promptfoo",
				CoveredSurfaceIDs: surfaces,
			})
		}
	}

	return scenarios
}

// --- Helpers ---

var (
	// JS/TS imports: from 'path' or require('path')
	jsImportPattern = regexp.MustCompile(`(?:from\s+['"]([^'"]+)['"]|require\s*\(\s*['"]([^'"]+)['"])`)
	// Python imports: from module.path import name
	pyImportPattern = regexp.MustCompile(`from\s+([\w.]+)\s+import`)
)

func findImportedSurfaces(src, testPath string, prompts, datasets map[string]string) []string {
	surfaceIDs := map[string]bool{}

	// JS/TS imports.
	for _, m := range jsImportPattern.FindAllStringSubmatch(src, -1) {
		importPath := m[1]
		if importPath == "" {
			importPath = m[2]
		}
		resolved := resolveImportPath(testPath, importPath)
		if resolved == "" {
			continue
		}
		matchSurfaces(resolved, prompts, datasets, surfaceIDs)
	}

	// Python imports (from x.y.z import name → x/y/z).
	for _, m := range pyImportPattern.FindAllStringSubmatch(src, -1) {
		modulePath := strings.ReplaceAll(m[1], ".", "/")
		matchSurfaces(modulePath, prompts, datasets, surfaceIDs)
	}

	out := make([]string, 0, len(surfaceIDs))
	for sid := range surfaceIDs {
		out = append(out, sid)
	}
	sort.Strings(out)
	return out
}

func matchSurfaces(resolved string, prompts, datasets map[string]string, surfaceIDs map[string]bool) {
	for key, sid := range prompts {
		if strings.HasPrefix(key, resolved) {
			surfaceIDs[sid] = true
		}
	}
	for key, sid := range datasets {
		if strings.HasPrefix(key, resolved) {
			surfaceIDs[sid] = true
		}
	}
}

func resolveImportPath(fromFile, importPath string) string {
	if !strings.HasPrefix(importPath, ".") {
		return "" // skip non-relative imports
	}
	dir := filepath.Dir(fromFile)
	resolved := filepath.Join(dir, importPath)
	resolved = filepath.ToSlash(filepath.Clean(resolved))
	// Strip leading ./
	resolved = strings.TrimPrefix(resolved, "./")
	return resolved
}

func isEvalTestPath(path string) bool {
	lower := strings.ToLower(path)
	parts := strings.Split(strings.ReplaceAll(lower, "\\", "/"), "/")
	for _, p := range parts {
		switch p {
		case "eval", "evals", "evaluations", "__evals__", "benchmarks":
			return true
		}
	}
	return false
}

func classifyScenarioCategory(path, content string) string {
	lower := strings.ToLower(path + " " + content)
	switch {
	case strings.Contains(lower, "safety") || strings.Contains(lower, "harm") || strings.Contains(lower, "toxic"):
		return "safety"
	case strings.Contains(lower, "accuracy") || strings.Contains(lower, "precision") || strings.Contains(lower, "recall"):
		return "accuracy"
	case strings.Contains(lower, "regression") || strings.Contains(lower, "baseline"):
		return "regression"
	case strings.Contains(lower, "bias") || strings.Contains(lower, "fairness"):
		return "bias"
	case strings.Contains(lower, "latency") || strings.Contains(lower, "performance"):
		return "performance"
	default:
		return "eval"
	}
}

func scenarioID(framework, path string) string {
	h := sha256.Sum256([]byte(path))
	return fmt.Sprintf("scenario:%s:%x", framework, h[:6])
}

func scenarioNameFromPath(path string) string {
	base := filepath.Base(path)
	// Remove test prefixes/suffixes and extensions.
	name := strings.TrimSuffix(base, filepath.Ext(base))
	for _, prefix := range []string{"test_", "test-"} {
		name = strings.TrimPrefix(name, prefix)
	}
	for _, suffix := range []string{".test", ".spec", "_test", "-test"} {
		name = strings.TrimSuffix(name, suffix)
	}
	return strings.ReplaceAll(name, "_", "-")
}

func deduplicateScenarios(scenarios []models.Scenario) []models.Scenario {
	seen := map[string]bool{}
	var out []models.Scenario
	for _, s := range scenarios {
		if !seen[s.ScenarioID] {
			seen[s.ScenarioID] = true
			out = append(out, s)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ScenarioID < out[j].ScenarioID
	})
	return out
}
