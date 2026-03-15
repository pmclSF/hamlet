package health

import (
	"fmt"

	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/runtime"
)

// FlakyTestDetector identifies tests with evidence of flakiness from
// runtime artifacts.
//
// Detection approach:
//   - Retry evidence: tests that were retried (explicit retry metadata)
//   - Mixed outcomes: same test name appears with both pass and fail status
//     within a single artifact (e.g., retry succeeded after initial failure)
//
// Limitations:
//   - Single-run data cannot establish statistical flake rates
//   - Multi-run history-based detection is future work
//   - Without explicit retry metadata, detection relies on duplicate test names
type FlakyTestDetector struct{}

// Detect scans runtime results for flakiness evidence.
func (d *FlakyTestDetector) Detect(results []runtime.TestResult) []models.Signal {
	var signals []models.Signal

	// Strategy 1: Explicit retry metadata.
	seen := map[string]bool{}
	for _, r := range results {
		key := r.Name
		if r.TestID != "" {
			key = r.TestID
		}
		if r.Retried && !seen[key] {
			seen[key] = true
			signals = append(signals, buildFlakySignal(r, "Retry behavior detected in test result artifact."))
		}
	}

	// Strategy 2: Mixed outcomes for same test name (pass + fail).
	type outcome struct {
		passed bool
		failed bool
		file   string
		name   string
		testID string
	}
	outcomes := map[string]*outcome{}
	for _, r := range results {
		key := r.Suite + "::" + r.Name
		o, ok := outcomes[key]
		if !ok {
			o = &outcome{file: r.File, name: r.Name, testID: r.TestID}
			outcomes[key] = o
		}
		if r.Status == runtime.StatusPassed {
			o.passed = true
		}
		if r.Status == runtime.StatusFailed || r.Status == runtime.StatusError {
			o.failed = true
		}
	}
	for _, o := range outcomes {
		key := o.name
		if o.testID != "" {
			key = o.testID
		}
		if o.passed && o.failed && !seen[key] {
			seen[key] = true
			meta := map[string]any{}
			if o.testID != "" {
				meta["testId"] = o.testID
			}
			signals = append(signals, models.Signal{
				Type:     "flakyTest",
				Category: models.CategoryHealth,
				Severity: models.SeverityMedium,
				Confidence: 0.7,
				Location: models.SignalLocation{
					File:   o.file,
					Symbol: o.name,
				},
				Explanation:      "Test appears to have failed before succeeding in the same run.",
				SuggestedAction:  "Investigate non-deterministic dependencies, timing issues, or shared mutable state.",
				EvidenceStrength: models.EvidenceModerate,
				EvidenceSource:   models.SourceRuntime,
				Metadata:         meta,
			})
		}
	}

	return signals
}

func buildFlakySignal(r runtime.TestResult, explanation string) models.Signal {
	sev := models.SeverityMedium
	confidence := 0.8

	meta := map[string]any{
		"retryAttempt": r.RetryAttempt,
	}
	if r.TestID != "" {
		meta["testId"] = r.TestID
	}

	return models.Signal{
		Type:       "flakyTest",
		Category:   models.CategoryHealth,
		Severity:   sev,
		Confidence: confidence,
		Location: models.SignalLocation{
			File:   r.File,
			Symbol: r.Name,
		},
		Explanation:     fmt.Sprintf("%s Test: %s", explanation, r.Name),
		SuggestedAction: "Investigate non-deterministic dependencies, timing issues, or shared mutable state.",
		EvidenceStrength: models.EvidenceModerate,
		EvidenceSource:   models.SourceRuntime,
		Metadata:         meta,
	}
}
