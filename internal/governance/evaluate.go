// Package governance evaluates repository state against local policy
// and emits canonical governance signals for violations.
package governance

import (
	"fmt"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/policy"
)

// Result holds the outcome of a policy evaluation.
type Result struct {
	Violations []models.Signal
	Pass       bool
}

// Evaluate checks the snapshot against the given policy and returns
// governance signals for any violations found.
//
// The evaluation is deterministic and transparent — each violation
// explains exactly what policy was violated and what evidence triggered it.
func Evaluate(snap *models.TestSuiteSnapshot, cfg *policy.Config) *Result {
	var violations []models.Signal

	if cfg == nil || cfg.IsEmpty() {
		return &Result{Pass: true}
	}

	violations = append(violations, checkDisallowedFrameworks(snap, cfg)...)
	violations = append(violations, checkSkippedTests(snap, cfg)...)
	violations = append(violations, checkRuntimeBudget(snap, cfg)...)
	violations = append(violations, checkCoverageThreshold(snap, cfg)...)
	violations = append(violations, checkWeakAssertionThreshold(snap, cfg)...)
	violations = append(violations, checkMockHeavyThreshold(snap, cfg)...)

	return &Result{
		Violations: violations,
		Pass:       len(violations) == 0,
	}
}

// checkDisallowedFrameworks emits legacyFrameworkUsage signals for each
// disallowed framework found in the repository.
func checkDisallowedFrameworks(snap *models.TestSuiteSnapshot, cfg *policy.Config) []models.Signal {
	if len(cfg.Rules.DisallowFrameworks) == 0 {
		return nil
	}

	disallowed := map[string]bool{}
	for _, f := range cfg.Rules.DisallowFrameworks {
		disallowed[strings.ToLower(f)] = true
	}

	var signals []models.Signal
	for _, fw := range snap.Frameworks {
		if disallowed[strings.ToLower(fw.Name)] {
			signals = append(signals, models.Signal{
				Type:     "legacyFrameworkUsage",
				Category: models.CategoryGovernance,
				Severity: models.SeverityHigh,
				Confidence: 1.0,
				Location: models.SignalLocation{
					Repository: snap.Repository.Name,
				},
				Explanation: fmt.Sprintf(
					"Policy disallows framework '%s', but %d test files were detected.",
					fw.Name, fw.FileCount,
				),
				SuggestedAction: fmt.Sprintf(
					"Migrate or remove '%s' framework usage.", fw.Name,
				),
				Metadata: map[string]any{
					"framework": fw.Name,
					"fileCount": fw.FileCount,
					"rule":      "disallow_frameworks",
				},
			})
		}
	}
	return signals
}

// checkSkippedTests emits a policyViolation signal if skipped tests
// are found when policy disallows them.
func checkSkippedTests(snap *models.TestSuiteSnapshot, cfg *policy.Config) []models.Signal {
	if cfg.Rules.DisallowSkippedTests == nil || !*cfg.Rules.DisallowSkippedTests {
		return nil
	}

	// Count existing skippedTest signals from the analysis phase
	var skippedCount int
	for _, s := range snap.Signals {
		if s.Type == "skippedTest" {
			skippedCount++
		}
	}

	if skippedCount == 0 {
		return nil
	}

	return []models.Signal{{
		Type:       "policyViolation",
		Category:   models.CategoryGovernance,
		Severity:   models.SeverityMedium,
		Confidence: 1.0,
		Location: models.SignalLocation{
			Repository: snap.Repository.Name,
		},
		Explanation: fmt.Sprintf(
			"Policy disallows skipped tests, but %d skipped test signal(s) were detected.",
			skippedCount,
		),
		SuggestedAction: "Restore or remove skipped tests.",
		Metadata: map[string]any{
			"skippedCount": skippedCount,
			"rule":         "disallow_skipped_tests",
		},
	}}
}

// checkRuntimeBudget emits runtimeBudgetExceeded signals for test files
// whose average runtime exceeds the configured maximum.
func checkRuntimeBudget(snap *models.TestSuiteSnapshot, cfg *policy.Config) []models.Signal {
	if cfg.Rules.MaxTestRuntimeMs == nil {
		return nil
	}
	maxMs := *cfg.Rules.MaxTestRuntimeMs

	var signals []models.Signal
	for _, tf := range snap.TestFiles {
		if tf.RuntimeStats != nil && tf.RuntimeStats.AvgRuntimeMs > 0 && tf.RuntimeStats.AvgRuntimeMs > maxMs {
			signals = append(signals, models.Signal{
				Type:       "runtimeBudgetExceeded",
				Category:   models.CategoryGovernance,
				Severity:   models.SeverityMedium,
				Confidence: 1.0,
				Location: models.SignalLocation{
					File: tf.Path,
				},
				Explanation: fmt.Sprintf(
					"Observed runtime %.0fms exceeds configured max_test_runtime_ms of %.0f.",
					tf.RuntimeStats.AvgRuntimeMs, maxMs,
				),
				SuggestedAction: "Reduce runtime hotspots or adjust policy intentionally.",
				Metadata: map[string]any{
					"observedMs": tf.RuntimeStats.AvgRuntimeMs,
					"budgetMs":   maxMs,
					"rule":       "max_test_runtime_ms",
				},
			})
		}
	}
	return signals
}

// checkCoverageThreshold emits a policyViolation if coverage threshold
// break signals exist and policy requires minimum coverage.
func checkCoverageThreshold(snap *models.TestSuiteSnapshot, cfg *policy.Config) []models.Signal {
	if cfg.Rules.MinimumCoveragePercent == nil {
		return nil
	}

	var coverageBreaks int
	for _, s := range snap.Signals {
		if s.Type == "coverageThresholdBreak" {
			coverageBreaks++
		}
	}

	if coverageBreaks == 0 {
		return nil
	}

	return []models.Signal{{
		Type:       "policyViolation",
		Category:   models.CategoryGovernance,
		Severity:   models.SeverityHigh,
		Confidence: 1.0,
		Location: models.SignalLocation{
			Repository: snap.Repository.Name,
		},
		Explanation: fmt.Sprintf(
			"Policy requires minimum %.0f%% coverage, but %d coverage threshold break(s) detected.",
			*cfg.Rules.MinimumCoveragePercent, coverageBreaks,
		),
		SuggestedAction: "Increase test coverage to meet the configured threshold.",
		Metadata: map[string]any{
			"coverageBreaks": coverageBreaks,
			"threshold":      *cfg.Rules.MinimumCoveragePercent,
			"rule":           "minimum_coverage_percent",
		},
	}}
}

// checkWeakAssertionThreshold emits a policyViolation if weakAssertion
// signal count exceeds the configured maximum.
func checkWeakAssertionThreshold(snap *models.TestSuiteSnapshot, cfg *policy.Config) []models.Signal {
	if cfg.Rules.MaxWeakAssertions == nil {
		return nil
	}

	var count int
	for _, s := range snap.Signals {
		if s.Type == "weakAssertion" {
			count++
		}
	}

	max := *cfg.Rules.MaxWeakAssertions
	if count <= max {
		return nil
	}

	return []models.Signal{{
		Type:       "policyViolation",
		Category:   models.CategoryGovernance,
		Severity:   models.SeverityMedium,
		Confidence: 1.0,
		Location: models.SignalLocation{
			Repository: snap.Repository.Name,
		},
		Explanation: fmt.Sprintf(
			"Found %d weakAssertion signal(s), exceeding policy maximum of %d.",
			count, max,
		),
		SuggestedAction: "Add meaningful assertions to test files with weak or missing assertions.",
		Metadata: map[string]any{
			"count": count,
			"max":   max,
			"rule":  "max_weak_assertions",
		},
	}}
}

// checkMockHeavyThreshold emits a policyViolation if mockHeavyTest
// signal count exceeds the configured maximum.
func checkMockHeavyThreshold(snap *models.TestSuiteSnapshot, cfg *policy.Config) []models.Signal {
	if cfg.Rules.MaxMockHeavyTests == nil {
		return nil
	}

	var count int
	for _, s := range snap.Signals {
		if s.Type == "mockHeavyTest" {
			count++
		}
	}

	max := *cfg.Rules.MaxMockHeavyTests
	if count <= max {
		return nil
	}

	return []models.Signal{{
		Type:       "policyViolation",
		Category:   models.CategoryGovernance,
		Severity:   models.SeverityMedium,
		Confidence: 1.0,
		Location: models.SignalLocation{
			Repository: snap.Repository.Name,
		},
		Explanation: fmt.Sprintf(
			"Found %d mockHeavyTest signal(s), exceeding policy maximum of %d.",
			count, max,
		),
		SuggestedAction: "Reduce mock usage in favor of real implementations in tests.",
		Metadata: map[string]any{
			"count": count,
			"max":   max,
			"rule":  "max_mock_heavy_tests",
		},
	}}
}
