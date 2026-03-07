// Package quality implements quality-focused signal detectors for Hamlet.
package quality

import "github.com/pmclSF/hamlet/internal/models"

// WeakAssertionDetector identifies test files with low assertion density.
//
// Heuristic:
//   - A test file with tests but fewer assertions than tests is suspicious.
//   - Files where assertion-to-test ratio < 1.0 are flagged as medium severity.
//   - Files with zero assertions but at least one test are flagged as high severity.
//
// Limitations:
//   - Cannot distinguish strong assertions from trivial ones (e.g. toBeDefined).
//   - Regex-based counting may miss some assertion patterns.
//   - Does not evaluate assertion quality, only density.
type WeakAssertionDetector struct{}

// Detect scans test files for weak assertion patterns.
func (d *WeakAssertionDetector) Detect(snap *models.TestSuiteSnapshot) []models.Signal {
	var signals []models.Signal

	for _, tf := range snap.TestFiles {
		if tf.TestCount == 0 {
			continue
		}

		ratio := float64(tf.AssertionCount) / float64(tf.TestCount)

		if tf.AssertionCount == 0 {
			signals = append(signals, models.Signal{
				Type:             "weakAssertion",
				Category:         models.CategoryQuality,
				Severity:         models.SeverityHigh,
				Confidence:       0.8,
				EvidenceStrength: models.EvidenceModerate,
				EvidenceSource:   models.SourceStructuralPattern,
				Location:         models.SignalLocation{File: tf.Path},
				Explanation: "No assertions detected in file with " +
					itoa(tf.TestCount) + " test(s). Tests execute code but do not verify behavior.",
				SuggestedAction: "Add assertions on returned values, state transitions, or side effects.",
			})
		} else if ratio < 1.0 {
			signals = append(signals, models.Signal{
				Type:             "weakAssertion",
				Category:         models.CategoryQuality,
				Severity:         models.SeverityMedium,
				Confidence:       0.6,
				EvidenceStrength: models.EvidenceWeak,
				EvidenceSource:   models.SourceStructuralPattern,
				Location:         models.SignalLocation{File: tf.Path},
				Explanation: "Low assertion density: " + itoa(tf.AssertionCount) +
					" assertion(s) across " + itoa(tf.TestCount) +
					" test(s). Some tests may not meaningfully verify behavior.",
				SuggestedAction: "Add assertions on outputs, state changes, or user-visible behavior.",
			})
		}
	}

	return signals
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
