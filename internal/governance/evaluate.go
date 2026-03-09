// Package governance evaluates repository state against local policy
// and emits canonical governance signals for violations.
package governance

import (
	"fmt"
	"sort"
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

	topFiles := topFilesForType(snap.Signals, "skippedTest", 5)
	if len(topFiles) == 0 {
		return nil
	}

	skippedCount := 0
	for _, tf := range topFiles {
		skippedCount += tf.count
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
			"Policy disallows skipped tests, but %d skipped test signal(s) were detected. Top files: %s.",
			skippedCount, formatTopFiles(topFiles),
		),
		SuggestedAction: "Restore or remove skipped tests.",
		Metadata: map[string]any{
			"skippedCount": skippedCount,
			"rule":         "disallow_skipped_tests",
			"topFiles":     topFileNames(topFiles),
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

	topFiles := topFilesForType(snap.Signals, "weakAssertion", 5)
	count := 0
	for _, tf := range topFiles {
		count += tf.count
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
			"Found %d weakAssertion signal(s), exceeding policy maximum of %d. Top files: %s.",
			count, max, formatTopFiles(topFiles),
		),
		SuggestedAction: "Add meaningful assertions to test files with weak or missing assertions.",
		Metadata: map[string]any{
			"count":    count,
			"max":      max,
			"rule":     "max_weak_assertions",
			"topFiles": topFileNames(topFiles),
		},
	}}
}

// checkMockHeavyThreshold emits a policyViolation if mockHeavyTest
// signal count exceeds the configured maximum.
func checkMockHeavyThreshold(snap *models.TestSuiteSnapshot, cfg *policy.Config) []models.Signal {
	if cfg.Rules.MaxMockHeavyTests == nil {
		return nil
	}

	topFiles := topFilesForType(snap.Signals, "mockHeavyTest", 5)
	count := 0
	for _, tf := range topFiles {
		count += tf.count
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
			"Found %d mockHeavyTest signal(s), exceeding policy maximum of %d. Top files: %s.",
			count, max, formatTopFiles(topFiles),
		),
		SuggestedAction: "Reduce mock usage in favor of real implementations in tests.",
		Metadata: map[string]any{
			"count":    count,
			"max":      max,
			"rule":     "max_mock_heavy_tests",
			"topFiles": topFileNames(topFiles),
		},
	}}
}

// fileCount tracks signal count per file for governance reporting.
type fileCount struct {
	file  string
	count int
}

// topFilesForType returns the top N files with the most signals of the given type.
func topFilesForType(signals []models.Signal, signalType models.SignalType, limit int) []fileCount {
	counts := map[string]int{}
	for _, s := range signals {
		if s.Type == signalType && s.Location.File != "" {
			counts[s.Location.File]++
		}
	}
	// Also count signals without file location.
	total := 0
	for _, s := range signals {
		if s.Type == signalType {
			total++
		}
	}

	if total == 0 {
		return nil
	}

	// If no file-level signals, return a single entry with empty file.
	if len(counts) == 0 {
		return []fileCount{{file: "", count: total}}
	}

	result := make([]fileCount, 0, len(counts))
	for f, c := range counts {
		result = append(result, fileCount{file: f, count: c})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].count != result[j].count {
			return result[i].count > result[j].count
		}
		return result[i].file < result[j].file
	})
	if len(result) > limit {
		result = result[:limit]
	}
	return result
}

// formatTopFiles formats file counts for human-readable explanations.
func formatTopFiles(files []fileCount) string {
	parts := make([]string, 0, len(files))
	for _, f := range files {
		if f.file == "" {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s (%d)", f.file, f.count))
	}
	if len(parts) == 0 {
		return "(no file-level detail)"
	}
	return strings.Join(parts, ", ")
}

// topFileNames extracts just the file paths for metadata.
func topFileNames(files []fileCount) []string {
	names := make([]string, 0, len(files))
	for _, f := range files {
		if f.file != "" {
			names = append(names, f.file)
		}
	}
	return names
}
