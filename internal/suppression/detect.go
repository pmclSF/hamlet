package suppression

import (
	"sort"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
)

// Detect analyzes a snapshot for suppression mechanisms.
func Detect(snap *models.TestSuiteSnapshot) *SuppressionResult {
	if snap == nil {
		return &SuppressionResult{}
	}

	var suppressions []Suppression

	// Strategy 1: Detect from existing signals (skippedTest signals).
	suppressions = append(suppressions, detectFromSignals(snap)...)

	// Strategy 2: Detect from runtime data (high skip/retry rates).
	suppressions = append(suppressions, detectFromRuntime(snap)...)

	// Strategy 3: Detect from test file naming conventions.
	suppressions = append(suppressions, detectFromNaming(snap)...)

	// Deduplicate by file+kind.
	suppressions = dedup(suppressions)

	// Classify intent for each suppression.
	for i := range suppressions {
		if suppressions[i].Intent == "" {
			suppressions[i].Intent = classifyIntent(suppressions[i], snap)
		}
	}

	// Sort for determinism.
	sort.Slice(suppressions, func(i, j int) bool {
		if suppressions[i].Kind != suppressions[j].Kind {
			return suppressions[i].Kind < suppressions[j].Kind
		}
		return suppressions[i].TestFilePath < suppressions[j].TestFilePath
	})

	return buildResult(suppressions)
}

func detectFromSignals(snap *models.TestSuiteSnapshot) []Suppression {
	var result []Suppression
	for _, sig := range snap.Signals {
		if sig.Type == "skippedTest" {
			result = append(result, Suppression{
				TestFilePath: sig.Location.File,
				Kind:         KindSkipDisable,
				Source:       SourceSignal,
				Confidence:   0.9,
				Explanation:  "skipped test detected via signal analysis",
			})
		}
	}
	return result
}

func detectFromRuntime(snap *models.TestSuiteSnapshot) []Suppression {
	var result []Suppression
	for _, tf := range snap.TestFiles {
		if tf.RuntimeStats == nil {
			continue
		}
		// High retry rate suggests retry wrapper policy.
		if tf.RuntimeStats.RetryRate >= 0.3 {
			result = append(result, Suppression{
				TestFilePath: tf.Path,
				Kind:         KindRetryWrapper,
				Source:       SourceRuntimeData,
				Confidence:   0.7,
				Explanation:  "high retry rate suggests retry-as-policy pattern",
				Metadata: map[string]any{
					"retryRate": tf.RuntimeStats.RetryRate,
				},
			})
		}
		// Very low pass rate with continued presence suggests expected failure.
		if tf.RuntimeStats.PassRate > 0 && tf.RuntimeStats.PassRate < 0.3 {
			result = append(result, Suppression{
				TestFilePath: tf.Path,
				Kind:         KindExpectedFailure,
				Source:       SourceRuntimeData,
				Confidence:   0.5,
				Explanation:  "persistently low pass rate may indicate expected-failure acceptance",
				Metadata: map[string]any{
					"passRate": tf.RuntimeStats.PassRate,
				},
			})
		}
	}
	return result
}

func detectFromNaming(snap *models.TestSuiteSnapshot) []Suppression {
	var result []Suppression
	quarantinePatterns := []string{
		"quarantine", "quarantined", ".skip", "skip.",
		"disabled", ".disabled", "xdescribe", "xit",
		"pending", ".pending",
	}
	for _, tf := range snap.TestFiles {
		lower := strings.ToLower(tf.Path)
		for _, pat := range quarantinePatterns {
			if strings.Contains(lower, pat) {
				kind := KindSkipDisable
				if strings.Contains(lower, "quarantine") {
					kind = KindQuarantined
				}
				result = append(result, Suppression{
					TestFilePath: tf.Path,
					Kind:         kind,
					Source:       SourceNaming,
					Confidence:   0.6,
					Explanation:  "file path contains suppression indicator: " + pat,
				})
				break // One match per file is enough.
			}
		}
	}
	return result
}

func classifyIntent(s Suppression, snap *models.TestSuiteSnapshot) SuppressionIntent {
	// If we have runtime data showing the pattern persists, it's chronic.
	for _, tf := range snap.TestFiles {
		if tf.Path == s.TestFilePath && tf.RuntimeStats != nil {
			// If it's been retried heavily or has very low pass rate, likely chronic.
			if s.Kind == KindRetryWrapper && tf.RuntimeStats.RetryRate >= 0.5 {
				return IntentChronic
			}
			if s.Kind == KindExpectedFailure {
				return IntentChronic
			}
		}
	}

	// Quarantined tests are chronic by nature unless explicitly marked tactical.
	if s.Kind == KindQuarantined {
		return IntentChronic
	}

	return IntentUnknown
}

func dedup(suppressions []Suppression) []Suppression {
	seen := map[string]bool{}
	var result []Suppression
	for _, s := range suppressions {
		key := string(s.Kind) + ":" + s.TestFilePath
		if !seen[key] {
			seen[key] = true
			result = append(result, s)
		}
	}
	return result
}

func buildResult(suppressions []Suppression) *SuppressionResult {
	r := &SuppressionResult{
		Suppressions: suppressions,
	}

	files := map[string]bool{}
	for _, s := range suppressions {
		files[s.TestFilePath] = true
		switch s.Kind {
		case KindQuarantined:
			r.QuarantinedCount++
		case KindExpectedFailure:
			r.ExpectedFailureCount++
		case KindSkipDisable:
			r.SkipDisableCount++
		case KindRetryWrapper:
			r.RetryWrapperCount++
		}
		switch s.Intent {
		case IntentTactical:
			r.TacticalCount++
		case IntentChronic:
			r.ChronicCount++
		default:
			r.UnknownCount++
		}
	}
	r.TotalSuppressedTests = len(files)

	return r
}
