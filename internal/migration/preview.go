// Migration preview boundary.
//
// The preview system provides structured analysis of what migration
// would look like for a given file or narrow scope. It does NOT
// attempt full automated conversion. Instead it:
//
//   - identifies the source framework
//   - surfaces blockers specific to this file
//   - classifies migration difficulty
//   - suggests a target framework if inferable
//   - reports honestly when preview is not possible
//
// This is intentionally conservative. The preview boundary is useful
// even when migration cannot proceed automatically.
package migration

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/signals"
)

// PreviewResult is the structured output of a migration preview
// for a single file or narrow scope.
type PreviewResult struct {
	// File is the path that was previewed.
	File string `json:"file"`

	// SourceFramework is the detected framework for this file.
	SourceFramework string `json:"sourceFramework"`

	// SuggestedTarget is the inferred migration target, if any.
	// Empty string means no target could be inferred.
	SuggestedTarget string `json:"suggestedTarget,omitempty"`

	// Difficulty is the estimated migration difficulty: "low", "medium", "high".
	Difficulty string `json:"difficulty"`

	// Blockers lists migration blockers found in this file.
	Blockers []PreviewBlocker `json:"blockers,omitempty"`

	// SafePatterns lists patterns in this file that should migrate cleanly.
	SafePatterns []string `json:"safePatterns,omitempty"`

	// PreviewAvailable indicates whether a meaningful preview could be produced.
	PreviewAvailable bool `json:"previewAvailable"`

	// Explanation describes the preview result in human terms.
	Explanation string `json:"explanation"`

	// Limitations describes what this preview cannot assess.
	Limitations []string `json:"limitations,omitempty"`
}

// PreviewBlocker describes a specific migration obstacle in a file.
type PreviewBlocker struct {
	// Type is the blocker taxonomy category.
	Type string `json:"type"`

	// Pattern is the specific pattern detected.
	Pattern string `json:"pattern"`

	// Explanation describes why this blocks migration.
	Explanation string `json:"explanation"`

	// Remediation suggests what to do about it.
	Remediation string `json:"remediation"`
}

// PreviewFile generates a migration preview for a single file.
//
// It uses the snapshot to understand framework context and existing
// signals, then analyzes the file content for migration-specific patterns.
//
// This function is honest about limitations. If the file cannot be
// meaningfully previewed, the result explains why.
func PreviewFile(snap *models.TestSuiteSnapshot, filePath string, repoRoot string) *PreviewResult {
	result := &PreviewResult{
		File: filePath,
	}

	// Find the test file in the snapshot.
	var testFile *models.TestFile
	for i := range snap.TestFiles {
		if snap.TestFiles[i].Path == filePath {
			testFile = &snap.TestFiles[i]
			break
		}
	}

	if testFile == nil {
		result.PreviewAvailable = false
		result.Explanation = fmt.Sprintf("File %s not found in analysis snapshot. Run 'terrain analyze' first.", filePath)
		result.Difficulty = "unknown"
		return result
	}

	result.SourceFramework = testFile.Framework

	// Determine language support for preview.
	lang := frameworkLanguage(testFile.Framework)
	if lang != "js" {
		result.PreviewAvailable = false
		result.Difficulty = "unknown"
		result.Explanation = fmt.Sprintf(
			"Migration preview is currently supported for JavaScript/TypeScript frameworks. "+
				"Detected framework: %s (%s).",
			testFile.Framework, lang,
		)
		result.Limitations = []string{
			fmt.Sprintf("No preview support for %s frameworks yet.", lang),
		}
		return result
	}

	// Read and analyze the file.
	content := readFile(repoRoot, filePath)
	if content == "" {
		result.PreviewAvailable = false
		result.Difficulty = "unknown"
		result.Explanation = fmt.Sprintf("Could not read file %s.", filePath)
		return result
	}

	// Collect blockers from this file.
	result.Blockers = collectFileBlockers(content, testFile.Framework)

	// Collect existing signals for this file from the snapshot.
	var fileSignals []models.Signal
	for _, s := range snap.Signals {
		if s.Location.File == filePath && signals.IsMigrationSignal(s.Type) {
			fileSignals = append(fileSignals, s)
		}
	}

	// Add any snapshot-level blockers not already captured.
	for _, s := range fileSignals {
		if bt, ok := s.Metadata["blockerType"].(string); ok {
			found := false
			for _, b := range result.Blockers {
				if b.Type == bt {
					found = true
					break
				}
			}
			if !found {
				result.Blockers = append(result.Blockers, PreviewBlocker{
					Type:        bt,
					Pattern:     string(s.Type),
					Explanation: s.Explanation,
					Remediation: s.SuggestedAction,
				})
			}
		}
	}

	// Infer target framework.
	result.SuggestedTarget = inferTarget(testFile.Framework, snap)

	// Identify safe patterns.
	result.SafePatterns = identifySafePatterns(content, testFile.Framework)

	// Classify difficulty.
	result.Difficulty = classifyDifficulty(result.Blockers, testFile)

	// Build explanation and determine if preview is meaningful.
	result.PreviewAvailable = true
	result.Explanation = buildPreviewExplanation(result)

	// Standard limitations.
	result.Limitations = []string{
		"Preview is based on structural pattern analysis, not full AST parsing.",
		"Custom build configurations and plugins are not assessed.",
		"Runtime behavior differences between frameworks are not evaluated.",
	}

	return result
}

// collectFileBlockers analyzes file content for migration-blocking patterns.
func collectFileBlockers(content string, framework string) []PreviewBlocker {
	var blockers []PreviewBlocker

	// Deprecated patterns
	for _, pattern := range detectDeprecatedJS(content, framework) {
		blockers = append(blockers, PreviewBlocker{
			Type:        BlockerDeprecatedPattern,
			Pattern:     pattern,
			Explanation: deprecatedPatternExplanation(pattern),
			Remediation: deprecatedPatternRemediation(pattern),
		})
	}

	// Dynamic test generation
	if hasDynamicTestGeneration(content) {
		blockers = append(blockers, PreviewBlocker{
			Type:        BlockerDynamicGeneration,
			Pattern:     "dynamic-test-generation",
			Explanation: "Dynamic test generation makes migration unpredictable. Each dynamically generated test may need individual review.",
			Remediation: "Consider converting dynamic tests to explicit test cases before migration.",
		})
	}

	// Custom matchers
	if hasCustomMatchers(content) {
		blockers = append(blockers, PreviewBlocker{
			Type:        BlockerCustomMatcher,
			Pattern:     "custom-matcher",
			Explanation: "Custom matchers are framework-specific and must be rewritten for the target framework.",
			Remediation: "Document custom matchers and create equivalent implementations for the target framework.",
		})
	}

	// Unsupported setup
	for _, pattern := range detectUnsupportedSetup(content, framework) {
		blockers = append(blockers, PreviewBlocker{
			Type:        BlockerUnsupportedSetup,
			Pattern:     pattern,
			Explanation: unsupportedSetupExplanation(pattern),
			Remediation: "Review setup/fixture patterns and find equivalents in the target framework.",
		})
	}

	return blockers
}

func deprecatedPatternExplanation(pattern string) string {
	switch pattern {
	case "done-callback":
		return "done() callbacks are deprecated in favor of async/await. Must be modernized before migration."
	case "setTimeout-in-test":
		return "setTimeout in tests creates fragile timing dependencies that may behave differently in other frameworks."
	case "enzyme-usage":
		return "Enzyme is deprecated. Migration to another framework should also migrate from Enzyme to a modern testing library."
	case "sinon-standalone":
		return "Sinon standalone usage needs replacement with the target framework's mocking capabilities."
	default:
		return fmt.Sprintf("Deprecated pattern '%s' should be modernized before migration.", pattern)
	}
}

func deprecatedPatternRemediation(pattern string) string {
	switch pattern {
	case "done-callback":
		return "Convert to async/await pattern."
	case "setTimeout-in-test":
		return "Replace with framework-provided waiting/timer utilities."
	case "enzyme-usage":
		return "Migrate to React Testing Library or the target framework's component testing."
	case "sinon-standalone":
		return "Replace with the target framework's built-in mocking (e.g., vi.fn() for Vitest)."
	default:
		return "Update to modern equivalent."
	}
}

func unsupportedSetupExplanation(pattern string) string {
	switch pattern {
	case "jest-global-setup":
		return "Jest globalSetup/globalTeardown hooks need equivalent configuration in the target framework."
	case "mocha-root-hooks":
		return "Mocha root hooks need equivalent lifecycle configuration in the target framework."
	case "cypress-custom-commands":
		return "Cypress custom commands have no direct equivalent in other E2E frameworks."
	case "cypress-plugin-events":
		return "Cypress plugin events are framework-specific and need reimplementation."
	case "framework-test-context":
		return "Test context APIs (this.timeout, this.retries) are framework-specific."
	default:
		return fmt.Sprintf("Setup pattern '%s' may not have a direct equivalent.", pattern)
	}
}

// inferTarget determines the most likely migration target framework
// based on the source framework and what else exists in the repo.
func inferTarget(source string, snap *models.TestSuiteSnapshot) string {
	// Common migration paths
	commonTargets := map[string]string{
		"jest":    "vitest",
		"mocha":   "vitest",
		"jasmine": "vitest",
		"cypress": "playwright",
		"puppeteer": "playwright",
		"testcafe":  "playwright",
		"webdriverio": "playwright",
	}

	// If there's a common target and the repo already has it, that's strong signal.
	if target, ok := commonTargets[source]; ok {
		for _, fw := range snap.Frameworks {
			if fw.Name == target {
				return target
			}
		}
		// Still suggest the common target even if not present.
		return target
	}

	return ""
}

// identifySafePatterns finds patterns in the file that should migrate cleanly.
func identifySafePatterns(content string, framework string) []string {
	var safe []string

	// Standard describe/it/test blocks
	if strings.Contains(content, "describe(") || strings.Contains(content, "it(") || strings.Contains(content, "test(") {
		safe = append(safe, "standard test structure (describe/it/test)")
	}

	// Standard assertions
	if strings.Contains(content, "expect(") {
		safe = append(safe, "expect() assertions")
	}

	// beforeEach/afterEach
	if strings.Contains(content, "beforeEach(") || strings.Contains(content, "afterEach(") {
		safe = append(safe, "standard lifecycle hooks (beforeEach/afterEach)")
	}

	// Standard imports
	if strings.Contains(content, "import ") {
		safe = append(safe, "ES module imports")
	}

	return safe
}

// classifyDifficulty determines migration difficulty based on blockers and file complexity.
func classifyDifficulty(blockers []PreviewBlocker, tf *models.TestFile) string {
	if len(blockers) == 0 {
		return "low"
	}

	// Count high-impact blocker types
	highImpact := 0
	for _, b := range blockers {
		switch b.Type {
		case BlockerCustomMatcher, BlockerUnsupportedSetup, BlockerFrameworkHelper:
			highImpact++
		}
	}

	if highImpact >= 2 || len(blockers) >= 4 {
		return "high"
	}
	if highImpact >= 1 || len(blockers) >= 2 {
		return "medium"
	}
	return "low"
}

func buildPreviewExplanation(result *PreviewResult) string {
	var parts []string

	parts = append(parts, fmt.Sprintf("Source framework: %s.", result.SourceFramework))

	if result.SuggestedTarget != "" {
		parts = append(parts, fmt.Sprintf("Suggested target: %s.", result.SuggestedTarget))
	}

	if len(result.Blockers) == 0 {
		parts = append(parts, "No migration blockers detected. This file should be straightforward to migrate.")
	} else {
		parts = append(parts, fmt.Sprintf(
			"%d migration blocker(s) found. Difficulty: %s.",
			len(result.Blockers), result.Difficulty,
		))
	}

	if len(result.SafePatterns) > 0 {
		parts = append(parts, fmt.Sprintf(
			"%d standard pattern(s) detected that should migrate cleanly.",
			len(result.SafePatterns),
		))
	}

	return strings.Join(parts, " ")
}

// PreviewScope generates migration previews for all files matching a
// directory scope. Returns previews sorted by difficulty (high first).
func PreviewScope(snap *models.TestSuiteSnapshot, scope string, repoRoot string) []*PreviewResult {
	var results []*PreviewResult

	for _, tf := range snap.TestFiles {
		dir := filepath.Dir(tf.Path)
		if scope != "" && !strings.HasPrefix(dir, scope) && dir != scope {
			continue
		}

		preview := PreviewFile(snap, tf.Path, repoRoot)
		results = append(results, preview)
	}

	// Sort by difficulty: high → medium → low → unknown
	diffOrder := map[string]int{"high": 0, "medium": 1, "low": 2, "unknown": 3}
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if diffOrder[results[i].Difficulty] > diffOrder[results[j].Difficulty] {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}
