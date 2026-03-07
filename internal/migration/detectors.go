// Package migration implements detectors for migration-related signals.
//
// Migration signals help teams understand what makes framework migration
// difficult, what patterns are deprecated, and where modernization effort
// should be concentrated.
//
// These detectors analyze test file content for patterns that indicate
// migration complexity. They are heuristic-based and conservative.
package migration

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
)

// Blocker taxonomy categories.
// These help group migration issues for prioritization.
const (
	BlockerCustomMatcher    = "custom-matcher"
	BlockerDynamicGeneration = "dynamic-generation"
	BlockerDeprecatedPattern = "deprecated-pattern"
	BlockerFrameworkHelper   = "framework-helper"
	BlockerUnsupportedSetup  = "unsupported-setup"
)

// DeprecatedPatternDetector identifies deprecated or outdated test patterns
// that complicate migration and should be modernized.
type DeprecatedPatternDetector struct {
	RepoRoot string
}

func (d *DeprecatedPatternDetector) Detect(snap *models.TestSuiteSnapshot) []models.Signal {
	var signals []models.Signal

	for _, tf := range snap.TestFiles {
		lang := frameworkLanguage(tf.Framework)
		if lang != "js" {
			continue
		}

		content := readFile(d.RepoRoot, tf.Path)
		if content == "" {
			continue
		}

		found := detectDeprecatedJS(content)
		for _, pattern := range found {
			signals = append(signals, models.Signal{
				Type:       "deprecatedTestPattern",
				Category:   models.CategoryMigration,
				Severity:   models.SeverityMedium,
				Confidence: 0.7,
				Location:   models.SignalLocation{File: tf.Path},
				Owner:      tf.Owner,
				Explanation: fmt.Sprintf(
					"Deprecated pattern '%s' found in %s.",
					pattern, tf.Path,
				),
				SuggestedAction: "Update to modern testing patterns before migration.",
				Metadata: map[string]any{
					"pattern":       pattern,
					"blockerType":   BlockerDeprecatedPattern,
				},
			})
		}
	}
	return signals
}

// DynamicTestGenerationDetector identifies tests that use dynamic test
// generation patterns, which reduce migration predictability.
type DynamicTestGenerationDetector struct {
	RepoRoot string
}

func (d *DynamicTestGenerationDetector) Detect(snap *models.TestSuiteSnapshot) []models.Signal {
	var signals []models.Signal

	for _, tf := range snap.TestFiles {
		lang := frameworkLanguage(tf.Framework)
		if lang != "js" {
			continue
		}

		content := readFile(d.RepoRoot, tf.Path)
		if content == "" {
			continue
		}

		if hasDynamicTestGeneration(content) {
			signals = append(signals, models.Signal{
				Type:       "dynamicTestGeneration",
				Category:   models.CategoryMigration,
				Severity:   models.SeverityMedium,
				Confidence: 0.6,
				Location:   models.SignalLocation{File: tf.Path},
				Owner:      tf.Owner,
				Explanation: fmt.Sprintf(
					"Dynamic test generation detected in %s. This reduces migration predictability.",
					tf.Path,
				),
				SuggestedAction: "Review dynamic test generation for migration compatibility.",
				Metadata: map[string]any{
					"blockerType": BlockerDynamicGeneration,
				},
			})
		}
	}
	return signals
}

// CustomMatcherDetector identifies custom matcher/helper patterns that
// complicate framework migration.
type CustomMatcherDetector struct {
	RepoRoot string
}

func (d *CustomMatcherDetector) Detect(snap *models.TestSuiteSnapshot) []models.Signal {
	var signals []models.Signal

	for _, tf := range snap.TestFiles {
		lang := frameworkLanguage(tf.Framework)
		if lang != "js" {
			continue
		}

		content := readFile(d.RepoRoot, tf.Path)
		if content == "" {
			continue
		}

		if hasCustomMatchers(content) {
			signals = append(signals, models.Signal{
				Type:       "customMatcherRisk",
				Category:   models.CategoryMigration,
				Severity:   models.SeverityLow,
				Confidence: 0.5,
				Location:   models.SignalLocation{File: tf.Path},
				Owner:      tf.Owner,
				Explanation: fmt.Sprintf(
					"Custom matcher or assertion helper usage in %s may need manual migration.",
					tf.Path,
				),
				SuggestedAction: "Document custom matchers and plan manual migration strategy.",
				Metadata: map[string]any{
					"blockerType": BlockerCustomMatcher,
				},
			})
		}
	}
	return signals
}

// Deprecated JS patterns
var (
	// done() callback pattern (use async/await instead)
	doneCallbackPattern = regexp.MustCompile(`\b(it|test)\s*\([^,]+,\s*function\s*\(\s*done\s*\)`)
	// .toBeTruthy()/.toBeFalsy() without specific value check
	weakTruthyPattern = regexp.MustCompile(`\.(toBeTruthy|toBeFalsy)\s*\(\s*\)`)
	// setTimeout in tests (fragile timing)
	setTimeoutInTest = regexp.MustCompile(`setTimeout\s*\(`)
	// Enzyme-specific patterns (deprecated React testing)
	enzymePattern = regexp.MustCompile(`\b(shallow|mount|render)\s*\(.*<`)
	// sinon standalone usage (older mocking)
	sinonPattern = regexp.MustCompile(`\bsinon\.(stub|mock|spy)\s*\(`)
)

func detectDeprecatedJS(content string) []string {
	var found []string
	if doneCallbackPattern.MatchString(content) {
		found = append(found, "done-callback")
	}
	if setTimeoutInTest.MatchString(content) {
		found = append(found, "setTimeout-in-test")
	}
	if enzymePattern.MatchString(content) {
		found = append(found, "enzyme-usage")
	}
	if sinonPattern.MatchString(content) {
		found = append(found, "sinon-standalone")
	}
	return found
}

// Dynamic test generation patterns
var (
	forEachTestPattern = regexp.MustCompile(`\.\s*forEach\s*\([^)]*\)\s*=>\s*\{[^}]*(it|test|describe)\s*\(`)
	mapTestPattern     = regexp.MustCompile(`\.\s*map\s*\([^)]*\)\s*=>\s*\{[^}]*(it|test|describe)\s*\(`)
	testEachPattern    = regexp.MustCompile(`\b(it|test|describe)\.each\s*[\(\[]`)
	forLoopTestPattern = regexp.MustCompile(`for\s*\([^)]+\)\s*\{[^}]*(it|test|describe)\s*\(`)
)

func hasDynamicTestGeneration(content string) bool {
	return forEachTestPattern.MatchString(content) ||
		mapTestPattern.MatchString(content) ||
		testEachPattern.MatchString(content) ||
		forLoopTestPattern.MatchString(content)
}

// Custom matcher patterns
var (
	addMatcherPattern    = regexp.MustCompile(`\bexpect\.extend\s*\(`)
	customMatcherPattern = regexp.MustCompile(`\bjest\.addMatchers?\s*\(`)
	chaiUsePattern       = regexp.MustCompile(`\bchai\.use\s*\(`)
	customAssertHelper   = regexp.MustCompile(`\bfunction\s+assert[A-Z]\w+\s*\(`)
)

func hasCustomMatchers(content string) bool {
	return addMatcherPattern.MatchString(content) ||
		customMatcherPattern.MatchString(content) ||
		chaiUsePattern.MatchString(content) ||
		customAssertHelper.MatchString(content)
}

func readFile(root, relPath string) string {
	path := root + "/" + relPath
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func frameworkLanguage(framework string) string {
	switch framework {
	case "jest", "vitest", "mocha", "jasmine", "cypress", "playwright", "puppeteer", "webdriverio", "testcafe":
		return "js"
	case "go-testing":
		return "go"
	case "pytest", "unittest", "nose2":
		return "python"
	case "junit4", "junit5", "testng":
		return "java"
	default:
		return "js"
	}
}

// FrameworkMigrationDetector detects multi-framework situations that
// suggest migration opportunity.
type FrameworkMigrationDetector struct{}

func (d *FrameworkMigrationDetector) Detect(snap *models.TestSuiteSnapshot) []models.Signal {
	if len(snap.Frameworks) < 2 {
		return nil
	}

	// Only flag if there are multiple unit-test frameworks (not e2e + unit mix)
	var unitFrameworks []models.Framework
	for _, fw := range snap.Frameworks {
		if fw.Type == models.FrameworkTypeUnit || fw.Type == models.FrameworkTypeUnknown {
			unitFrameworks = append(unitFrameworks, fw)
		}
	}

	if len(unitFrameworks) < 2 {
		return nil
	}

	names := make([]string, len(unitFrameworks))
	for i, fw := range unitFrameworks {
		names[i] = fw.Name
	}

	return []models.Signal{{
		Type:       "frameworkMigration",
		Category:   models.CategoryMigration,
		Severity:   models.SeverityInfo,
		Confidence: 0.8,
		Location: models.SignalLocation{
			Repository: snap.Repository.Name,
		},
		Explanation: fmt.Sprintf(
			"Multiple unit-test frameworks detected (%s). Consolidation may simplify maintenance.",
			strings.Join(names, ", "),
		),
		SuggestedAction: "Evaluate consolidating to a single test framework.",
		Metadata: map[string]any{
			"frameworks":     names,
			"frameworkCount": len(unitFrameworks),
		},
	}}
}
