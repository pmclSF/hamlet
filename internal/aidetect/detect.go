package aidetect

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/astguard"
	"github.com/pmclSF/terrain/internal/gitignore"
	"github.com/pmclSF/terrain/internal/saferead"
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

	// CallSites are AST-resolved AI SDK invocations. Each carries the
	// SDK identity, method shape, model arg (when static), and
	// confidence. Populated alongside ModelFiles during the source
	// scan; the regex-based ModelFiles entry remains as a safety net
	// until the AST detector reaches recall parity.
	CallSites []AICallSite `json:"callSites,omitempty"`
}

// maxSourceFileSize is the maximum file size (256 KB) to scan for AI patterns.
// Files larger than this are skipped to avoid slow analysis on generated code.
const maxSourceFileSize = 256 * 1024

// Detect scans a repository root for AI/ML frameworks, prompt patterns,
// dataset usage, and model invocations. No configuration required.
//
// This is a convenience wrapper that uses context.Background(). For
// cancellation support — required by callers driving Terrain from a
// CI workflow with a `--timeout`, or `terrain ai run` invoked from
// within an already-cancelling pipeline — use DetectContext.
func Detect(root string) *DetectResult {
	return DetectContext(context.Background(), root)
}

// DetectContext is like Detect but respects ctx for cancellation. The
// file-walking phase (Phase 3) checks ctx at each entry and aborts
// the walk cleanly when cancelled, returning whatever has been
// collected so far. Phases 1 and 2 (config-file probes and dependency
// manifest reads) are bounded — at most a few stat / open calls — so
// they don't need granular cancellation; we still check ctx between
// phases so a caller cancelling between Phase 2 and Phase 3 doesn't
// pay for the source walk.
//
// Honors caller cancellation through the AI detector path. The
// non-context shape (`Detect(root)` only) silently ignores ctx, so a
// slow AI scan would block until the walk completed even when the
// calling pipeline had already cancelled.
func DetectContext(ctx context.Context, root string) *DetectResult {
	result := &DetectResult{}

	// Phase 1: Check config files.
	detectConfigFiles(root, result)
	if ctx.Err() != nil {
		return result
	}

	// Phase 2: Check dependency manifests.
	detectDependencies(root, result)
	if ctx.Err() != nil {
		return result
	}

	// Phase 3: Scan source files for import patterns and AI code patterns.
	detectFromSourceCtx(ctx, root, result)

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
	if data, err := saferead.ReadFile(pkgPath); err == nil {
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
	if data, err := saferead.ReadFile(pyPath); err == nil {
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
	if data, err := saferead.ReadFile(reqPath); err == nil {
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

	// Check go.mod (Go module declarations).
	goModPath := filepath.Join(root, "go.mod")
	if data, err := saferead.ReadFile(goModPath); err == nil {
		content := string(data)
		for _, sig := range KnownFrameworks {
			for _, key := range sig.DependencyKeys {
				// go.mod entries are module paths like
				// github.com/sashabaranov/go-openai — substring match
				// against the DependencyKeys configured for the framework.
				if strings.Contains(content, key) {
					result.Frameworks = append(result.Frameworks, Framework{
						Name:       sig.Name,
						Source:     "dependency",
						ConfigFile: "go.mod",
						Confidence: 0.9,
					})
				}
			}
		}
	}

	// Check Maven pom.xml.
	pomPath := filepath.Join(root, "pom.xml")
	if data, err := saferead.ReadFile(pomPath); err == nil {
		content := string(data)
		for _, sig := range KnownFrameworks {
			for _, key := range sig.DependencyKeys {
				if strings.Contains(content, key) {
					result.Frameworks = append(result.Frameworks, Framework{
						Name:       sig.Name,
						Source:     "dependency",
						ConfigFile: "pom.xml",
						Confidence: 0.9,
					})
				}
			}
		}
	}

	// Check Gradle build.gradle / build.gradle.kts.
	for _, name := range []string{"build.gradle", "build.gradle.kts"} {
		gradlePath := filepath.Join(root, name)
		data, err := saferead.ReadFile(gradlePath)
		if err != nil {
			continue
		}
		content := string(data)
		for _, sig := range KnownFrameworks {
			for _, key := range sig.DependencyKeys {
				if strings.Contains(content, key) {
					result.Frameworks = append(result.Frameworks, Framework{
						Name:       sig.Name,
						Source:     "dependency",
						ConfigFile: name,
						Confidence: 0.9,
					})
				}
			}
		}
	}
}

// Model invocation patterns in source code.
var modelCallPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\bOpenAI\b|ChatCompletion|chat\.completions\.create`),
	regexp.MustCompile(`\bAnthropic\b|messages\.create|claude`),
	regexp.MustCompile(`\.(invoke|predict|generate)\s*\(`),
	regexp.MustCompile(`\bLLMChain\b|\bChatPromptTemplate\b|\bPromptTemplate\b`),
	regexp.MustCompile(`\bpipeline\s*\(|\bAutoModel\b|\bAutoTokenizer\b`),
}

// detectFromSource walks source files looking for AI import patterns.
func detectFromSource(root string, result *DetectResult) {
	detectFromSourceCtx(context.Background(), root, result)
}

func detectFromSourceCtx(ctx context.Context, root string, result *DetectResult) {
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

	// fileCount tracks how many files we've examined so the ctx check
	// fires every 64 files rather than on every entry — checking
	// ctx.Err() is a system call on some platforms, and AI detection
	// already walks ~the whole repo, so the per-entry overhead is
	// noticeable.
	fileCount := 0

	// Honour the repo .gitignore, matching the analysis walker. Without it,
	// the scan surfaces vendored or local-only trees the user has declared
	// off-limits — and counts their prompts/models as the repo's own.
	ign := gitignore.Load(root)

	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		// Honor cancellation. Returning a non-nil error from WalkDir
		// stops the walk; we use ctx.Err() so callers can distinguish
		// "user cancelled" from "filesystem error" if they choose to
		// inspect the walk error.
		fileCount++
		if fileCount&0x3F == 0 {
			if ctx.Err() != nil {
				return ctx.Err()
			}
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)

		if d.IsDir() {
			// Use the same canonical skip set as walkRepoForConfigs and
			// internal/analysis/repository_scan.go. A narrower skip set
			// would descend into dist/, build/, .terrain/, vendor/,
			// target/, etc. — a major contributor to multi-walk
			// amplification on real repos.
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			if ign.Match(rel, true) {
				return filepath.SkipDir
			}
			return nil
		}

		if ign.Match(rel, false) {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		isGo := ext == ".go"
		isJava := ext == ".java"
		if ext != ".ts" && ext != ".js" && ext != ".py" &&
			ext != ".tsx" && ext != ".jsx" && ext != ".mjs" && !isGo && !isJava {
			return nil
		}

		// Reject non-regular files (symlink to /dev/zero, devices, FIFOs) and
		// oversize files before reading. os.ReadFile follows symlinks, so
		// checking length only after the read would pull an unbounded stream
		// into memory first; stat-then-read caps it up front.
		if info, infoErr := d.Info(); infoErr != nil || !info.Mode().IsRegular() || info.Size() > maxSourceFileSize {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(data)

		// Skip source whose bracket nesting would make tree-sitter parsing and
		// per-node traversal pathologically slow (a crafted file of thousands of
		// nested/unclosed brackets). Real code never approaches the threshold.
		if astguard.LooksPathological(data) {
			return nil
		}

		// Go and Java take only the AST detection path. Substring
		// matching on those languages' source produces false positives
		// (Terrain's own test files include Python import strings as
		// fixtures, for instance); the AST detectors are strict about
		// actual import bindings.
		if isGo {
			result.CallSites = append(result.CallSites, DetectGoAISurfaces(data, rel)...)
			return nil
		}
		if isJava {
			result.CallSites = append(result.CallSites, DetectJavaAISurfaces(data, rel)...)
			return nil
		}

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

		// Check for model invocation patterns (regex fallback path).
		for _, pat := range modelCallPatterns {
			if pat.MatchString(content) {
				modelFiles[rel] = true
				break
			}
		}

		// AST-based call-site extraction. The AST path adds structural
		// detail (SDK identity, method shape, static model arg) that the
		// regex modelCallPatterns can't surface. Both run for now;
		// removing the regex path is gated on AST recall parity.
		switch ext {
		case ".py":
			result.CallSites = append(result.CallSites, DetectPythonAISurfaces(data, rel)...)
		case ".ts", ".tsx", ".js", ".jsx", ".mjs":
			result.CallSites = append(result.CallSites, DetectJSAISurfaces(data, rel)...)
		}

		return nil
	})

	result.PromptFiles = sortedKeyList(promptFiles)
	result.DatasetFiles = sortedKeyList(datasetFiles)
	result.ModelFiles = sortedKeyList(modelFiles)
}

func containsAny(content string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(content, p) {
			return true
		}
	}
	return false
}

var promptPatterns = []string{
	"PromptTemplate", "ChatPromptTemplate", "SystemMessage",
	"HumanMessage", "AIMessage", "prompt_template",
	"system_prompt", "systemPrompt", "buildPrompt",
	"PROMPT", "prompt =", "template =",
}

var datasetPatterns = []string{
	"load_dataset", "Dataset", "DataLoader", "dataloader",
	"training_data", "eval_data", "test_data",
	"evalDataset", "trainingDataset",
}

func hasPromptPatterns(content string) bool  { return containsAny(content, promptPatterns) }
func hasDatasetPatterns(content string) bool { return containsAny(content, datasetPatterns) }

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
