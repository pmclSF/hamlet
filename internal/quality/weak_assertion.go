// Package quality implements quality-focused signal detectors for Hamlet.
package quality

import (
	"fmt"

	"github.com/pmclSF/hamlet/internal/models"
)

// e2eFrameworks are frameworks where lower assertion density is expected
// because tests include implicit checks (page loads, element presence, navigation).
var e2eFrameworks = map[string]bool{
	"cypress": true, "playwright": true, "puppeteer": true,
	"selenium": true, "webdriverio": true, "testcafe": true,
}

// WeakAssertionDetector identifies test files with low assertion density.
//
// Detection is framework-aware:
//   - E2E frameworks use lower thresholds (density < 0.5 instead of < 1.0)
//     because implicit checks (navigation, element presence) provide validation
//   - Snapshot-dominated files (≥80% snapshot assertions) are flagged as info
//     rather than weak, since snapshots are a valid regression strategy
//   - Unit test frameworks use the standard threshold
//
// Limitations:
//   - Cannot distinguish strong assertions from trivial ones (e.g. toBeDefined).
//   - Regex-based counting may miss some assertion patterns.
//   - Does not evaluate assertion quality, only density.
type WeakAssertionDetector struct{}

// Detect scans test files for weak assertion patterns.
func (d *WeakAssertionDetector) Detect(snap *models.TestSuiteSnapshot) []models.Signal {
	// Build framework type lookup from snapshot.
	fwTypes := map[string]models.FrameworkType{}
	for _, fw := range snap.Frameworks {
		fwTypes[fw.Name] = fw.Type
	}

	var signals []models.Signal

	for _, tf := range snap.TestFiles {
		if tf.TestCount == 0 {
			continue
		}

		isE2E := e2eFrameworks[tf.Framework] || fwTypes[tf.Framework] == models.FrameworkTypeE2E

		ratio := float64(tf.AssertionCount) / float64(tf.TestCount)

		// Snapshot-dominated: ≥80% of assertions are snapshots.
		// This is a valid regression testing strategy, not a weakness.
		if tf.AssertionCount > 0 && tf.SnapshotCount > 0 {
			snapshotRatio := float64(tf.SnapshotCount) / float64(tf.AssertionCount)
			if snapshotRatio >= 0.8 {
				// Only flag if overall density is also very low.
				if ratio < 0.5 {
					signals = append(signals, models.Signal{
						Type:             "weakAssertion",
						Category:         models.CategoryQuality,
						Severity:         models.SeverityInfo,
						Confidence:       0.4,
						EvidenceStrength: models.EvidenceWeak,
						EvidenceSource:   models.SourceStructuralPattern,
						Location:         models.SignalLocation{File: tf.Path},
						Explanation: fmt.Sprintf(
							"Snapshot-dominated test file (%d snapshot assertions of %d total) with low density (%.1f/test). Snapshots provide structural regression coverage.",
							tf.SnapshotCount, tf.AssertionCount, ratio,
						),
						SuggestedAction: "Consider adding behavioral assertions alongside snapshots for critical logic.",
					})
				}
				continue
			}
		}

		if tf.AssertionCount == 0 {
			sev := models.SeverityHigh
			conf := 0.8
			if isE2E {
				// E2E with zero explicit assertions may still validate via implicit checks.
				sev = models.SeverityMedium
				conf = 0.5
			}
			signals = append(signals, models.Signal{
				Type:             "weakAssertion",
				Category:         models.CategoryQuality,
				Severity:         sev,
				Confidence:       conf,
				EvidenceStrength: models.EvidenceModerate,
				EvidenceSource:   models.SourceStructuralPattern,
				Location:         models.SignalLocation{File: tf.Path},
				Explanation: "No assertions detected in file with " +
					itoa(tf.TestCount) + " test(s). Tests execute code but do not verify behavior.",
				SuggestedAction: "Add assertions on returned values, state transitions, or side effects.",
			})
		} else {
			// E2E threshold: 0.5 assertions/test (implicit checks provide coverage).
			// Unit threshold: 1.0 assertions/test.
			threshold := 1.0
			if isE2E {
				threshold = 0.5
			}

			if ratio < threshold {
				sev := models.SeverityMedium
				conf := 0.6
				if isE2E {
					sev = models.SeverityLow
					conf = 0.4
				}
				signals = append(signals, models.Signal{
					Type:             "weakAssertion",
					Category:         models.CategoryQuality,
					Severity:         sev,
					Confidence:       conf,
					EvidenceStrength: models.EvidenceWeak,
					EvidenceSource:   models.SourceStructuralPattern,
					Location:         models.SignalLocation{File: tf.Path},
					Explanation: fmt.Sprintf(
						"Low assertion density: %d assertion(s) across %d test(s) (%.1f/test). Some tests may not meaningfully verify behavior.",
						tf.AssertionCount, tf.TestCount, ratio,
					),
					SuggestedAction: "Add assertions on outputs, state changes, or user-visible behavior.",
				})
			}
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
