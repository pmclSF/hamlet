package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/analysis"
)

// InitResult holds everything detected during `terrain init`.
type InitResult struct {
	// Root is the absolute repository root path.
	Root string

	// Frameworks detected at the project level.
	Frameworks []DetectedFramework

	// Languages detected from test files.
	Languages []string

	// Artifacts discovered at known paths.
	Artifacts *ArtifactDiscovery

	// HasTerrainDir is true if .terrain/ already exists.
	HasTerrainDir bool

	// HasPolicyFile is true if .terrain/policy.yaml already exists.
	HasPolicyFile bool

	// HasTerrainYAML is true if terrain.yaml already exists.
	HasTerrainYAML bool

	// TestFileCount is the number of test files discovered.
	TestFileCount int

	// ConfigPath is the path to the generated config, if any.
	ConfigPath string

	// PolicyPath is the path to the generated policy, if any.
	PolicyPath string
}

// DetectedFramework captures a framework found during init scanning.
type DetectedFramework struct {
	Name       string `json:"name"`
	Language   string `json:"language"`
	Source     string `json:"source"` // "config-file", "dependency", "convention"
	Confidence float64 `json:"confidence"`
}

// RunInit performs the full initialization scan. It detects frameworks,
// artifacts, and existing configuration, then generates config files.
// This is non-interactive by default — it scans, generates, and reports.
func RunInit(root string) (*InitResult, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve root path: %w", err)
	}

	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("invalid root path %q: %w", root, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("invalid root path %q: not a directory", root)
	}

	result := &InitResult{Root: absRoot}

	// Detect existing configuration.
	result.HasTerrainDir = dirExists(filepath.Join(absRoot, ".terrain"))
	result.HasPolicyFile = fileExists(filepath.Join(absRoot, ".terrain", "policy.yaml"))
	result.HasTerrainYAML = fileExists(filepath.Join(absRoot, ".terrain", "terrain.yaml")) ||
		fileExists(filepath.Join(absRoot, "terrain.yaml"))

	// Detect project-level frameworks.
	projectCtx := analysis.DetectProjectFrameworks(absRoot)
	for lang, fws := range projectCtx.Frameworks {
		for _, fw := range fws {
			result.Frameworks = append(result.Frameworks, DetectedFramework{
				Name:       fw.Name,
				Language:   lang,
				Source:     fw.Source,
				Confidence: fw.Confidence,
			})
		}
	}
	sort.Slice(result.Frameworks, func(i, j int) bool {
		return result.Frameworks[i].Confidence > result.Frameworks[j].Confidence
	})

	// Detect languages from frameworks.
	langSet := map[string]bool{}
	for _, fw := range result.Frameworks {
		langSet[fw.Language] = true
	}
	for l := range langSet {
		result.Languages = append(result.Languages, l)
	}
	sort.Strings(result.Languages)

	// Quick test file count (from project context, no full scan needed).
	// For a lightweight init, we count test files by pattern matching.
	result.TestFileCount = countTestFiles(absRoot)

	// Discover artifacts.
	result.Artifacts = DiscoverArtifacts(absRoot)

	// Generate .terrain/ directory if it doesn't exist.
	terrainDir := filepath.Join(absRoot, ".terrain")
	if !result.HasTerrainDir {
		if err := os.MkdirAll(terrainDir, 0o755); err != nil {
			return nil, fmt.Errorf("create .terrain/ directory: %w", err)
		}
	}

	// Generate policy.yaml with commented defaults if it doesn't exist.
	if !result.HasPolicyFile {
		policyPath := filepath.Join(terrainDir, "policy.yaml")
		if err := generatePolicyYAML(policyPath); err != nil {
			return nil, fmt.Errorf("generate policy.yaml: %w", err)
		}
		result.PolicyPath = policyPath
	}

	// Generate terrain.yaml only if it doesn't exist and we found interesting things.
	if !result.HasTerrainYAML && (len(result.Frameworks) > 0 || result.TestFileCount > 0) {
		configPath := filepath.Join(terrainDir, "terrain.yaml")
		if err := generateTerrainYAML(configPath, result); err != nil {
			return nil, fmt.Errorf("generate terrain.yaml: %w", err)
		}
		result.ConfigPath = configPath
	}

	return result, nil
}

func generatePolicyYAML(path string) error {
	content := `# Terrain policy configuration
# Uncomment rules to enforce them in CI via: terrain policy check
#
# See: docs/examples/policy-check.md

rules:
  # disallow_skipped_tests: true
  # disallow_frameworks:
  #   - jest
  # max_test_runtime_ms: 5000
  # minimum_coverage_percent: 80
  # max_weak_assertions: 5
  # max_mock_heavy_tests: 3

  # AI governance rules (for repos with AI/eval scenarios):
  # ai:
  #   block_on_safety_failure: true
  #   block_on_accuracy_regression: 5
  #   block_on_uncovered_context: true
  #   warn_on_latency_regression: true
  #   warn_on_cost_regression: true
`
	return os.WriteFile(path, []byte(content), 0o644)
}

func generateTerrainYAML(path string, result *InitResult) error {
	var b strings.Builder

	b.WriteString("# Terrain configuration\n")
	b.WriteString("# Generated by: terrain init\n")
	b.WriteString("# Edit this file to declare manual coverage, scenarios, and CI metadata.\n")
	b.WriteString("#\n")
	b.WriteString("# Detected frameworks: ")
	if len(result.Frameworks) > 0 {
		names := make([]string, 0, len(result.Frameworks))
		for _, fw := range result.Frameworks {
			names = append(names, fw.Name)
		}
		b.WriteString(strings.Join(names, ", "))
	} else {
		b.WriteString("(none detected)")
	}
	b.WriteString("\n")

	if result.Artifacts != nil && result.Artifacts.CoveragePath != "" {
		b.WriteString("# Detected coverage: " + result.Artifacts.CoveragePath + " (" + result.Artifacts.CoverageFormat + ")\n")
	}
	if result.Artifacts != nil && len(result.Artifacts.RuntimePaths) > 0 {
		b.WriteString("# Detected runtime: " + result.Artifacts.RuntimePaths[0] + " (" + result.Artifacts.RuntimeFormats[0] + ")\n")
	}
	b.WriteString("#\n")
	b.WriteString("# See: docs/examples/manual-coverage.md\n\n")

	// Manual coverage section (commented template).
	b.WriteString("# manual_coverage:\n")
	b.WriteString("#   - name: Regression suite\n")
	b.WriteString("#     area: src/billing\n")
	b.WriteString("#     source: testrail\n")
	b.WriteString("#     owner: qa-team\n")
	b.WriteString("#     criticality: high\n")
	b.WriteString("#     frequency: per-release\n\n")

	// Scenarios section (commented template).
	b.WriteString("# scenarios:\n")
	b.WriteString("#   - name: prompt-accuracy\n")
	b.WriteString("#     category: accuracy\n")
	b.WriteString("#     framework: custom\n")
	b.WriteString("#     surfaces:\n")
	b.WriteString("#       - surface:src/prompts.ts:systemPrompt\n\n")

	// CI duration hint.
	b.WriteString("# ci_duration_seconds: 120\n")

	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// countTestFiles does a lightweight count of test files without full analysis.
func countTestFiles(root string) int {
	count := 0
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := d.Name()
			if base == "node_modules" || base == ".git" || base == "vendor" ||
				base == "dist" || base == "build" || base == "__pycache__" ||
				base == ".venv" || base == "venv" || base == ".terrain" {
				return filepath.SkipDir
			}
			return nil
		}
		name := d.Name()
		if isTestFileName(name) {
			count++
		}
		return nil
	})
	return count
}

func isTestFileName(name string) bool {
	// JS/TS test files.
	if strings.Contains(name, ".test.") || strings.Contains(name, ".spec.") {
		return true
	}
	// Go test files.
	if strings.HasSuffix(name, "_test.go") {
		return true
	}
	// Python test files.
	if strings.HasPrefix(name, "test_") && strings.HasSuffix(name, ".py") {
		return true
	}
	if strings.HasSuffix(name, "_test.py") {
		return true
	}
	// Java test files.
	if strings.HasSuffix(name, "Test.java") {
		return true
	}
	return false
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
