package stability

import "github.com/pmclSF/hamlet/internal/models"

// BuildHistories constructs TestHistory records from an ordered sequence
// of snapshots (oldest first). This function joins test cases across
// snapshots by TestID, attaching runtime and signal data where available.
func BuildHistories(snapshots []*models.TestSuiteSnapshot) []TestHistory {
	if len(snapshots) == 0 {
		return nil
	}

	// Collect all test IDs across all snapshots.
	allIDs := map[string]bool{}
	for _, snap := range snapshots {
		for _, tc := range snap.TestCases {
			if tc.TestID != "" {
				allIDs[tc.TestID] = true
			}
		}
	}

	// Build history for each test ID.
	histories := make([]TestHistory, 0, len(allIDs))
	for id := range allIDs {
		h := TestHistory{TestID: id}

		for snapIdx, snap := range snapshots {
			obs := Observation{
				SnapshotIndex: snapIdx,
			}

			// Find test case in this snapshot.
			tc := findTestCase(snap.TestCases, id)
			if tc == nil {
				obs.Present = false
			} else {
				obs.Present = true
				if h.TestName == "" {
					h.TestName = tc.TestName
					h.FilePath = tc.FilePath
					h.Framework = tc.Framework
				}

				// Attach runtime data from test file.
				tf := findTestFile(snap.TestFiles, tc.FilePath)
				if tf != nil && tf.RuntimeStats != nil {
					obs.HasRuntime = true
					obs.RuntimeMs = tf.RuntimeStats.AvgRuntimeMs
					obs.RetryRate = tf.RuntimeStats.RetryRate
					if tf.RuntimeStats.PassRate >= 0.95 {
						obs.Passed = true
					} else if tf.RuntimeStats.PassRate < 0.5 {
						obs.Failed = true
					}
					if h.Owner == "" && tf.Owner != "" {
						h.Owner = tf.Owner
					}
				}

				// Check for signals on this test's file.
				for _, sig := range snap.Signals {
					if sig.Location.File == tc.FilePath {
						switch sig.Type {
						case "flakyTest":
							obs.FlakySignal = true
						case "slowTest":
							obs.SlowSignal = true
						case "skippedTest":
							obs.Skipped = true
						}
					}
				}
			}

			h.Observations = append(h.Observations, obs)
		}

		histories = append(histories, h)
	}

	return histories
}

func findTestCase(cases []models.TestCase, testID string) *models.TestCase {
	for i := range cases {
		if cases[i].TestID == testID {
			return &cases[i]
		}
	}
	return nil
}

func findTestFile(files []models.TestFile, path string) *models.TestFile {
	for i := range files {
		if files[i].Path == path {
			return &files[i]
		}
	}
	return nil
}
