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
	BlockerCustomMatcher     = "custom-matcher"
	BlockerDynamicGeneration = "dynamic-generation"
	BlockerDeprecatedPattern = "deprecated-pattern"
	BlockerFrameworkHelper   = "framework-helper"
	BlockerUnsupportedSetup  = "unsupported-setup"
)

// Blocker severity tiers control how signals affect migration readiness.
const (
	// TierHardBlocker: genuinely prevents migration without significant rework.
	// Examples: Enzyme usage (unmaintained library), Cypress plugin internals.
	TierHardBlocker = "hard-blocker"

	// TierSoftBlocker: requires effort but has a clear migration path.
	// Examples: custom matchers (need rewriting), done-callbacks (need async/await).
	TierSoftBlocker = "soft-blocker"

	// TierAdvisory: worth knowing about but doesn't affect migration readiness.
	// Examples: setTimeout in tests (works everywhere), dynamic test generation
	// (standard practice), snapshot testing (supported by all modern frameworks).
	TierAdvisory = "advisory"
)

// patternToTier maps specific deprecated patterns to their severity tier.
var patternToTier = map[string]string{
	// Hard blockers — genuinely block migration.
	"enzyme-usage":            TierHardBlocker,
	"cypress-custom-commands": TierHardBlocker,
	"cypress-plugin-events":   TierHardBlocker,

	// Soft blockers — require effort but have clear migration paths.
	"done-callback":          TierSoftBlocker,
	"sinon-standalone":       TierSoftBlocker,
	"jest-global-setup":      TierSoftBlocker,
	"mocha-root-hooks":       TierSoftBlocker,
	"framework-test-context": TierSoftBlocker,

	// Advisories — valid patterns that work across frameworks.
	"setTimeout-in-test": TierAdvisory,
}

// blockerTypeToDefaultTier maps blocker categories to their default tier
// when no specific pattern-level override exists.
var blockerTypeToDefaultTier = map[string]string{
	BlockerCustomMatcher:     TierSoftBlocker,
	BlockerDynamicGeneration: TierAdvisory,
	BlockerDeprecatedPattern: TierSoftBlocker,
	BlockerFrameworkHelper:   TierSoftBlocker,
	BlockerUnsupportedSetup:  TierSoftBlocker,
}

// TierForSignal determines the blocker tier for a migration signal
// based on its metadata.
func TierForSignal(s models.Signal) string {
	if s.Type == "frameworkMigration" {
		return TierAdvisory
	}
	// Check for explicit pattern-level tier first.
	if pattern, ok := s.Metadata["pattern"].(string); ok {
		if tier, ok := patternToTier[pattern]; ok {
			return tier
		}
	}
	// Fall back to blocker-type-level default.
	if bt, ok := s.Metadata["blockerType"].(string); ok {
		if tier, ok := blockerTypeToDefaultTier[bt]; ok {
			return tier
		}
	}
	return TierSoftBlocker // conservative default
}

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

		found := detectDeprecatedJS(content, tf.Framework)
		for _, pattern := range found {
			tier := patternToTier[pattern]
			if tier == "" {
				tier = blockerTypeToDefaultTier[BlockerDeprecatedPattern]
			}
			signals = append(signals, models.Signal{
				Type:             "deprecatedTestPattern",
				Category:         models.CategoryMigration,
				Severity:         models.SeverityMedium,
				Confidence:       0.7,
				EvidenceStrength: models.EvidenceModerate,
				EvidenceSource:   models.SourceStructuralPattern,
				Location:         models.SignalLocation{File: tf.Path},
				Owner:            tf.Owner,
				Explanation: fmt.Sprintf(
					"Deprecated pattern '%s' found in %s.",
					pattern, tf.Path,
				),
				SuggestedAction: "Update to modern testing patterns before migration.",
				Metadata: map[string]any{
					"pattern":     pattern,
					"blockerType": BlockerDeprecatedPattern,
				},
			})

			// Hard blockers additionally emit a migrationBlocker signal so
			// risk scoring and readiness assessment can aggregate them.
			if tier == TierHardBlocker {
				signals = append(signals, models.Signal{
					Type:             "migrationBlocker",
					Category:         models.CategoryMigration,
					Severity:         models.SeverityHigh,
					Confidence:       0.8,
					EvidenceStrength: models.EvidenceModerate,
					EvidenceSource:   models.SourceStructuralPattern,
					Location:         models.SignalLocation{File: tf.Path},
					Owner:            tf.Owner,
					Explanation: fmt.Sprintf(
						"Hard migration blocker: '%s' in %s requires significant rework before migration.",
						pattern, tf.Path,
					),
					SuggestedAction: "Address this blocker before attempting framework migration.",
					Metadata: map[string]any{
						"pattern":     pattern,
						"blockerType": BlockerDeprecatedPattern,
						"blockerTier": TierHardBlocker,
					},
				})
			}
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
				Type:             "dynamicTestGeneration",
				Category:         models.CategoryMigration,
				Severity:         models.SeverityLow,
				Confidence:       0.6,
				EvidenceStrength: models.EvidenceModerate,
				EvidenceSource:   models.SourceStructuralPattern,
				Location:         models.SignalLocation{File: tf.Path},
				Owner:            tf.Owner,
				Explanation: fmt.Sprintf(
					"Dynamic test generation detected in %s. Parameterized tests are standard practice but may need syntax adjustment during migration.",
					tf.Path,
				),
				SuggestedAction: "Review dynamic test generation for migration compatibility.",
				Metadata: map[string]any{
					"blockerType": BlockerDynamicGeneration,
					"blockerTier": TierAdvisory,
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
				Type:             "customMatcherRisk",
				Category:         models.CategoryMigration,
				Severity:         models.SeverityLow,
				Confidence:       0.5,
				EvidenceStrength: models.EvidenceWeak,
				EvidenceSource:   models.SourceStructuralPattern,
				Location:         models.SignalLocation{File: tf.Path},
				Owner:            tf.Owner,
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
	doneCallbackPattern = regexp.MustCompile(`\b(it|test)\s*\(\s*[^,]*,\s*(?:async\s*)?(?:function\s*\(\s*done\s*\)|\(\s*done\s*\)\s*=>|done\s*=>)`)
	// setTimeout in tests (fragile timing)
	setTimeoutInTest = regexp.MustCompile(`setTimeout\s*\(`)
	// Enzyme-specific patterns (deprecated React testing)
	enzymePattern = regexp.MustCompile(`(?s)\b(shallow|mount|render)\s*\([^)]*<`)
	// sinon standalone usage (older mocking)
	sinonPattern = regexp.MustCompile(`\bsinon\.(stub|mock|spy)\s*\(`)
)

func detectDeprecatedJS(content string, framework string) []string {
	content = stripJSNoise(content)
	var found []string
	// done-callback is idiomatic in mocha; only flag for other frameworks
	if framework != "mocha" && doneCallbackPattern.MatchString(content) {
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
	content = stripJSNoise(content)
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
	content = stripJSNoise(content)
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

// UnsupportedSetupDetector identifies framework-specific setup and fixture
// patterns that may not have equivalents in a target framework.
type UnsupportedSetupDetector struct {
	RepoRoot string
}

func (d *UnsupportedSetupDetector) Detect(snap *models.TestSuiteSnapshot) []models.Signal {
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

		found := detectUnsupportedSetup(content, tf.Framework)
		for _, pattern := range found {
			tier := patternToTier[pattern]
			if tier == "" {
				tier = blockerTypeToDefaultTier[BlockerUnsupportedSetup]
			}
			signals = append(signals, models.Signal{
				Type:             "unsupportedSetup",
				Category:         models.CategoryMigration,
				Severity:         models.SeverityMedium,
				Confidence:       0.6,
				EvidenceStrength: models.EvidenceModerate,
				EvidenceSource:   models.SourceStructuralPattern,
				Location:         models.SignalLocation{File: tf.Path},
				Owner:            tf.Owner,
				Explanation: fmt.Sprintf(
					"Framework-specific setup pattern '%s' in %s may not have a direct equivalent in other frameworks.",
					pattern, tf.Path,
				),
				SuggestedAction: "Review setup/fixture patterns for migration compatibility.",
				Metadata: map[string]any{
					"pattern":     pattern,
					"blockerType": BlockerUnsupportedSetup,
				},
			})

			// Hard blockers additionally emit a migrationBlocker signal.
			if tier == TierHardBlocker {
				signals = append(signals, models.Signal{
					Type:             "migrationBlocker",
					Category:         models.CategoryMigration,
					Severity:         models.SeverityHigh,
					Confidence:       0.7,
					EvidenceStrength: models.EvidenceModerate,
					EvidenceSource:   models.SourceStructuralPattern,
					Location:         models.SignalLocation{File: tf.Path},
					Owner:            tf.Owner,
					Explanation: fmt.Sprintf(
						"Hard migration blocker: '%s' in %s requires significant rework before migration.",
						pattern, tf.Path,
					),
					SuggestedAction: "Address this blocker before attempting framework migration.",
					Metadata: map[string]any{
						"pattern":     pattern,
						"blockerType": BlockerUnsupportedSetup,
						"blockerTier": TierHardBlocker,
					},
				})
			}
		}
	}
	return signals
}

// Unsupported setup/fixture patterns
var (
	// Jest-specific global setup
	jestGlobalSetup = regexp.MustCompile(`\bglobalSetup\b|\bglobalTeardown\b`)
	// Mocha-specific root hooks
	mochaRootHooks = regexp.MustCompile(`\brootHooks?\b|\bexports\.mochaHooks\b`)
	// Cypress-specific commands and plugins
	cypressCommands = regexp.MustCompile(`\bCypress\.Commands\.add\s*\(`)
	cypressPlugins  = regexp.MustCompile(`\bCypress\.on\s*\(\s*['"]`)
	// Framework-specific test context
	testContextPattern = regexp.MustCompile(`\bthis\.timeout\s*\(|\bthis\.retries\s*\(|\bthis\.slow\s*\(`)
)

func detectUnsupportedSetup(content string, framework string) []string {
	content = stripJSNoise(content)
	var found []string
	if framework != "jest" && jestGlobalSetup.MatchString(content) {
		found = append(found, "jest-global-setup")
	}
	if framework != "mocha" && mochaRootHooks.MatchString(content) {
		found = append(found, "mocha-root-hooks")
	}
	if framework == "cypress" && cypressCommands.MatchString(content) {
		found = append(found, "cypress-custom-commands")
	}
	if framework == "cypress" && cypressPlugins.MatchString(content) {
		found = append(found, "cypress-plugin-events")
	}
	if framework == "mocha" && testContextPattern.MatchString(content) {
		found = append(found, "framework-test-context")
	}
	return found
}

// stripJSNoise removes JS comments and string/template literals from content
// so structural regex detectors avoid matching inside non-code text.
func stripJSNoise(content string) string {
	var out strings.Builder
	inLineComment := false
	inBlockComment := false
	inSingle := false
	inDouble := false
	inTemplate := false
	escaped := false

	for i := 0; i < len(content); i++ {
		ch := content[i]
		next := byte(0)
		if i+1 < len(content) {
			next = content[i+1]
		}

		if inLineComment {
			if ch == '\n' {
				inLineComment = false
				out.WriteByte('\n')
			}
			continue
		}
		if inBlockComment {
			if ch == '*' && next == '/' {
				inBlockComment = false
				i++
				continue
			}
			if ch == '\n' {
				out.WriteByte('\n')
			}
			continue
		}
		if inSingle {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '\'' {
				inSingle = false
			}
			if ch == '\n' {
				out.WriteByte('\n')
			}
			continue
		}
		if inDouble {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inDouble = false
			}
			if ch == '\n' {
				out.WriteByte('\n')
			}
			continue
		}
		if inTemplate {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '`' {
				inTemplate = false
			}
			if ch == '\n' {
				out.WriteByte('\n')
			}
			continue
		}

		if ch == '/' && next == '/' {
			inLineComment = true
			i++
			continue
		}
		if ch == '/' && next == '*' {
			inBlockComment = true
			i++
			continue
		}
		if ch == '\'' {
			inSingle = true
			escaped = false
			continue
		}
		if ch == '"' {
			inDouble = true
			escaped = false
			continue
		}
		if ch == '`' {
			inTemplate = true
			escaped = false
			continue
		}

		out.WriteByte(ch)
	}

	return out.String()
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
		Type:             "frameworkMigration",
		Category:         models.CategoryMigration,
		Severity:         models.SeverityInfo,
		Confidence:       0.8,
		EvidenceStrength: models.EvidenceStrong,
		EvidenceSource:   models.SourceStructuralPattern,
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
