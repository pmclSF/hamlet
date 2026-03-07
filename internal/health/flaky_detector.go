package health

import (
	"fmt"

	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/runtime"
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
		if r.Retried && !seen[r.Name] {
			seen[r.Name] = true
			signals = append(signals, buildFlakySignal(r, "Retry behavior detected in test result artifact."))
		}
	}

	// Strategy 2: Mixed outcomes for same test name (pass + fail).
	type outcome struct {
		passed bool
		failed bool
		file   string
		name   string
	}
	outcomes := map[string]*outcome{}
	for _, r := range results {
		key := r.Suite + "::" + r.Name
		o, ok := outcomes[key]
		if !ok {
			o = &outcome{file: r.File, name: r.Name}
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
		if o.passed && o.failed && !seen[o.name] {
			seen[o.name] = true
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
			})
		}
	}

	return signals
}

func buildFlakySignal(r runtime.TestResult, explanation string) models.Signal {
	sev := models.SeverityMedium
	confidence := 0.8

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
		Metadata: map[string]any{
			"retryAttempt": r.RetryAttempt,
		},
	}
}
