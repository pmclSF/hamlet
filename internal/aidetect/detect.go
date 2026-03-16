package aidetect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// DetectResult holds all AI/ML detection findings for a repository.
type DetectResult struct {
	// Frameworks detected in the repository.
	Frameworks []Framework `json:"frameworks"`

	// EvalConfigs are paths to eval framework config files found.
	EvalConfigs []string `json:"evalConfigs,omitempty"`

	// PromptFiles are source files containing prompt patterns.
	PromptFiles []string `json:"promptFiles,omitempty"`

	// DatasetFiles are source files containing dataset patterns.
	DatasetFiles []string `json:"datasetFiles,omitempty"`

	// ModelFiles are source files with model invocation patterns.
	ModelFiles []string `json:"modelFiles,omitempty"`
}

// Detect scans a repository root for AI/ML frameworks, prompt patterns,
// dataset usage, and model invocations. No configuration required.
func Detect(root string) *DetectResult {
	result := &DetectResult{}

	// Phase 1: Check config files.
	detectConfigFiles(root, result)

	// Phase 2: Check dependency manifests.
	detectDependencies(root, result)

	// Phase 3: Scan source files for import patterns and AI code patterns.
	detectFromSource(root, result)

	// Deduplicate frameworks by name, keeping highest confidence.
	result.Frameworks = deduplicateFrameworks(result.Frameworks)

	return result
}

// detectConfigFiles looks for framework-specific config files at the repo root.
func detectConfigFiles(root string, result *DetectResult) {
	for _, sig := range KnownFrameworks {
		for _, cfgName := range sig.ConfigFiles {
			// Check root and common subdirectories.
			candidates := []string{
				filepath.Join(root, cfgName),
				filepath.Join(root, ".config", cfgName),
			}
			for _, path := range candidates {
				if _, err := os.Stat(path); err == nil {
					rel, _ := filepath.Rel(root, path)
					result.Frameworks = append(result.Frameworks, Framework{
						Name:       sig.Name,
						Source:     "config",
						ConfigFile: rel,
						Confidence: 0.95,
					})
					result.EvalConfigs = append(result.EvalConfigs, rel)
				}
			}
		}
	}
}

// detectDependencies checks package.json and pyproject.toml for AI framework deps.
func detectDependencies(root string, result *DetectResult) {
	// Check package.json
	pkgPath := filepath.Join(root, "package.json")
	if data, err := os.ReadFile(pkgPath); err == nil {
		var pkg map[string]interface{}
		if json.Unmarshal(data, &pkg) == nil {
			allDeps := map[string]string{}
			for _, field := range []string{"dependencies", "devDependencies", "peerDependencies"} {
				if deps, ok := pkg[field].(map[string]interface{}); ok {
					for k, v := range deps {
						if vs, ok := v.(string); ok {
							allDeps[k] = vs
						}
					}
				}
			}

			for _, sig := range KnownFrameworks {
				for _, key := range sig.DependencyKeys {
					if ver, ok := allDeps[key]; ok {
						result.Frameworks = append(result.Frameworks, Framework{
							Name:       sig.Name,
							Version:    ver,
							Source:     "dependency",
							ConfigFile: "package.json",
							Confidence: 0.9,
						})
					}
				}
			}
		}
	}

	// Check pyproject.toml (simple string matching, no TOML parser to avoid deps).
	pyPath := filepath.Join(root, "pyproject.toml")
	if data, err := os.ReadFile(pyPath); err == nil {
		content := string(data)
		for _, sig := range KnownFrameworks {
			for _, key := range sig.DependencyKeys {
				if strings.Contains(content, key) {
					result.Frameworks = append(result.Frameworks, Framework{
						Name:       sig.Name,
						Source:     "dependency",
						ConfigFile: "pyproject.toml",
						Confidence: 0.85,
					})
				}
			}
		}
	}

	// Check requirements.txt
	reqPath := filepath.Join(root, "requirements.txt")
	if data, err := os.ReadFile(reqPath); err == nil {
		content := string(data)
		for _, sig := range KnownFrameworks {
			for _, key := range sig.DependencyKeys {
				if strings.Contains(content, key) {
					result.Frameworks = append(result.Frameworks, Framework{
						Name:       sig.Name,
						Source:     "dependency",
						ConfigFile: "requirements.txt",
						Confidence: 0.8,
					})
				}
			}
		}
	}
}

// Model invocation patterns in source code.
var modelCallPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)openai|ChatCompletion|chat\.completions\.create`),
	regexp.MustCompile(`(?i)anthropic|messages\.create|claude`),
	regexp.MustCompile(`(?i)\.invoke\(|\.call\(|\.predict\(|\.generate\(`),
	regexp.MustCompile(`(?i)LLMChain|ChatPromptTemplate|PromptTemplate`),
	regexp.MustCompile(`(?i)pipeline\(|AutoModel|AutoTokenizer`),
}

// detectFromSource walks source files looking for AI import patterns.
func detectFromSource(root string, result *DetectResult) {
	// Build pattern index for framework detection.
	type patternEntry struct {
		framework string
		pattern   string
	}
	var patterns []patternEntry
	for _, sig := range KnownFrameworks {
		for _, p := range sig.ImportPatterns {
			patterns = append(patterns, patternEntry{sig.Name, p})
		}
	}

	promptFiles := map[string]bool{}
	datasetFiles := map[string]bool{}
	modelFiles := map[string]bool{}
	frameworkHits := map[string]bool{}

	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			name := d.Name()
			if name == "node_modules" || name == ".git" || name == "__pycache__" || name == ".venv" || name == "venv" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".ts" && ext != ".js" && ext != ".py" && ext != ".tsx" && ext != ".jsx" && ext != ".mjs" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil || len(data) > 256*1024 { // skip files > 256KB
			return nil
		}
		content := string(data)
		rel, _ := filepath.Rel(root, path)

		// Check framework import patterns.
		for _, pe := range patterns {
			if strings.Contains(content, pe.pattern) && !frameworkHits[pe.framework+":"+rel] {
				frameworkHits[pe.framework+":"+rel] = true
				result.Frameworks = append(result.Frameworks, Framework{
					Name:       pe.framework,
					Source:     "import",
					ConfigFile: rel,
					Confidence: 0.75,
				})
			}
		}

		// Check for prompt patterns (broader than code surface detection).
		if hasPromptPatterns(content) {
			promptFiles[rel] = true
		}

		// Check for dataset patterns.
		if hasDatasetPatterns(content) {
			datasetFiles[rel] = true
		}

		// Check for model invocation patterns.
		for _, pat := range modelCallPatterns {
			if pat.MatchString(content) {
				modelFiles[rel] = true
				break
			}
		}

		return nil
	})

	result.PromptFiles = sortedKeyList(promptFiles)
	result.DatasetFiles = sortedKeyList(datasetFiles)
	result.ModelFiles = sortedKeyList(modelFiles)
}

func hasPromptPatterns(content string) bool {
	patterns := []string{
		"PromptTemplate", "ChatPromptTemplate", "SystemMessage",
		"HumanMessage", "AIMessage", "prompt_template",
		"system_prompt", "systemPrompt", "buildPrompt",
		"PROMPT", "prompt =", "template =",
	}
	for _, p := range patterns {
		if strings.Contains(content, p) {
			return true
		}
	}
	return false
}

func hasDatasetPatterns(content string) bool {
	patterns := []string{
		"load_dataset", "Dataset", "DataLoader", "dataloader",
		"training_data", "eval_data", "test_data",
		"evalDataset", "trainingDataset",
	}
	for _, p := range patterns {
		if strings.Contains(content, p) {
			return true
		}
	}
	return false
}

func deduplicateFrameworks(frameworks []Framework) []Framework {
	best := map[string]Framework{}
	for _, f := range frameworks {
		if existing, ok := best[f.Name]; !ok || f.Confidence > existing.Confidence {
			best[f.Name] = f
		}
	}
	out := make([]Framework, 0, len(best))
	for _, f := range best {
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Confidence != out[j].Confidence {
			return out[i].Confidence > out[j].Confidence
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func sortedKeyList(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
