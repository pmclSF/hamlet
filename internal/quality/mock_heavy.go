package quality

import "github.com/pmclSF/hamlet/internal/models"

// MockHeavyDetector identifies test files with high mock usage relative
// to direct assertions.
//
// Heuristic:
//   - If mock count exceeds assertion count, the test is mock-heavy.
//   - If mock count is >= 5 and assertion count is <= mock count, flag it.
//   - Higher ratios get higher severity.
//
// Limitations:
//   - Regex-based mock counting may miss some mock patterns or
//     over-count mock-related helper calls.
//   - Does not distinguish between necessary isolation mocks and
//     excessive mocking of internals.
//   - This is not saying mocking is bad — it signals increased risk
//     of false confidence when mocks dominate.
type MockHeavyDetector struct{}

// Detect scans test files for high mock usage patterns.
func (d *MockHeavyDetector) Detect(snap *models.TestSuiteSnapshot) []models.Signal {
	var signals []models.Signal

	for _, tf := range snap.TestFiles {
		if tf.MockCount == 0 || tf.TestCount == 0 {
			continue
		}

		// Only flag files with meaningful mock usage (>= 3 mocks)
		if tf.MockCount < 3 {
			continue
		}

		// Mock-to-assertion ratio
		if tf.AssertionCount == 0 || tf.MockCount > tf.AssertionCount {
			sev := models.SeverityMedium
			conf := 0.7
			if tf.AssertionCount == 0 {
				sev = models.SeverityHigh
				conf = 0.8
			} else if tf.MockCount > 2*tf.AssertionCount {
				sev = models.SeverityHigh
				conf = 0.75
			}

			signals = append(signals, models.Signal{
				Type:     "mockHeavyTest",
				Category: models.CategoryQuality,
				Severity: sev,
				Confidence: conf,
				Location: models.SignalLocation{File: tf.Path},
				Explanation: "High mock usage detected: " + itoa(tf.MockCount) +
					" mock(s) vs " + itoa(tf.AssertionCount) +
					" assertion(s). Test behavior may be heavily isolated behind mocks.",
				SuggestedAction: "Consider adding assertions on real outputs or supplementing with integration coverage.",
			})
		}
	}

	return signals
}
