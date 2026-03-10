package suppression

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestDetect_NilSnapshot(t *testing.T) {
	t.Parallel()
	result := Detect(nil)
	if result == nil {
		t.Fatal("expected non-nil result for nil snapshot")
	}
	if len(result.Suppressions) != 0 {
		t.Errorf("expected 0 suppressions, got %d", len(result.Suppressions))
	}
	if result.TotalSuppressedTests != 0 {
		t.Errorf("expected 0 total suppressed tests, got %d", result.TotalSuppressedTests)
	}
}

func TestDetect_EmptySnapshot(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{}
	result := Detect(snap)
	if len(result.Suppressions) != 0 {
		t.Errorf("expected 0 suppressions for empty snapshot, got %d", len(result.Suppressions))
	}
}

func TestDetect_QuarantinedNamingConvention(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "tests/quarantine/login_test.go"},
			{Path: "tests/quarantined_checkout_test.go"},
			{Path: "tests/normal_test.go"},
		},
	}

	result := Detect(snap)

	if result.QuarantinedCount != 2 {
		t.Errorf("expected 2 quarantined, got %d", result.QuarantinedCount)
	}

	for _, s := range result.Suppressions {
		if s.Kind == KindQuarantined {
			if s.Source != SourceNaming {
				t.Errorf("expected naming source, got %s", s.Source)
			}
			if s.Intent != IntentChronic {
				t.Errorf("expected chronic intent for quarantined test, got %s", s.Intent)
			}
		}
	}
}

func TestDetect_SkipNamingConvention(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "tests/login.skip.test.js"},
			{Path: "tests/skip.checkout_test.py"},
			{Path: "tests/disabled_payment_test.go"},
		},
	}

	result := Detect(snap)

	if result.SkipDisableCount != 3 {
		t.Errorf("expected 3 skip/disable, got %d", result.SkipDisableCount)
	}

	for _, s := range result.Suppressions {
		if s.Kind != KindSkipDisable {
			t.Errorf("expected skip_disable kind, got %s", s.Kind)
		}
		if s.Source != SourceNaming {
			t.Errorf("expected naming source, got %s", s.Source)
		}
	}
}

func TestDetect_SkippedTestSignal(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Signals: []models.Signal{
			{
				Type:     "skippedTest",
				Location: models.SignalLocation{File: "tests/auth_test.go"},
			},
			{
				Type:     "skippedTest",
				Location: models.SignalLocation{File: "tests/payment_test.go"},
			},
			{
				Type:     "lowCoverage",
				Location: models.SignalLocation{File: "tests/other_test.go"},
			},
		},
	}

	result := Detect(snap)

	if result.SkipDisableCount != 2 {
		t.Errorf("expected 2 skip/disable from signals, got %d", result.SkipDisableCount)
	}

	for _, s := range result.Suppressions {
		if s.Source != SourceSignal {
			t.Errorf("expected signal source, got %s", s.Source)
		}
		if s.Confidence != 0.9 {
			t.Errorf("expected 0.9 confidence, got %f", s.Confidence)
		}
	}
}

func TestDetect_RetryWrapper(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path: "tests/flaky_test.go",
				RuntimeStats: &models.RuntimeStats{
					RetryRate: 0.5,
					PassRate:  0.8,
				},
			},
			{
				Path: "tests/stable_test.go",
				RuntimeStats: &models.RuntimeStats{
					RetryRate: 0.05,
					PassRate:  0.99,
				},
			},
		},
	}

	result := Detect(snap)

	if result.RetryWrapperCount != 1 {
		t.Errorf("expected 1 retry wrapper, got %d", result.RetryWrapperCount)
	}

	found := false
	for _, s := range result.Suppressions {
		if s.Kind == KindRetryWrapper {
			found = true
			if s.TestFilePath != "tests/flaky_test.go" {
				t.Errorf("expected flaky_test.go, got %s", s.TestFilePath)
			}
			if s.Source != SourceRuntimeData {
				t.Errorf("expected runtime_data source, got %s", s.Source)
			}
			if s.Intent != IntentChronic {
				t.Errorf("expected chronic intent for high retry rate, got %s", s.Intent)
			}
			if rate, ok := s.Metadata["retryRate"].(float64); !ok || rate != 0.5 {
				t.Errorf("expected retryRate 0.5 in metadata, got %v", s.Metadata["retryRate"])
			}
		}
	}
	if !found {
		t.Error("expected to find a retry wrapper suppression")
	}
}

func TestDetect_ExpectedFailure(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path: "tests/known_broken_test.go",
				RuntimeStats: &models.RuntimeStats{
					PassRate:  0.1,
					RetryRate: 0.0,
				},
			},
		},
	}

	result := Detect(snap)

	if result.ExpectedFailureCount != 1 {
		t.Errorf("expected 1 expected failure, got %d", result.ExpectedFailureCount)
	}

	for _, s := range result.Suppressions {
		if s.Kind == KindExpectedFailure {
			if s.Intent != IntentChronic {
				t.Errorf("expected chronic intent for expected failure, got %s", s.Intent)
			}
			if rate, ok := s.Metadata["passRate"].(float64); !ok || rate != 0.1 {
				t.Errorf("expected passRate 0.1 in metadata, got %v", s.Metadata["passRate"])
			}
		}
	}
}

func TestDetect_ExpectedFailure_ZeroPassRateExcluded(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path: "tests/always_fails_test.go",
				RuntimeStats: &models.RuntimeStats{
					PassRate: 0.0, // Exactly zero is excluded (no data vs always fails).
				},
			},
		},
	}

	result := Detect(snap)

	if result.ExpectedFailureCount != 0 {
		t.Errorf("expected 0 expected failures for zero pass rate, got %d", result.ExpectedFailureCount)
	}
}

func TestDetect_MixedSuppressions(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "tests/quarantine/old_test.go"},
			{
				Path: "tests/retry_heavy_test.go",
				RuntimeStats: &models.RuntimeStats{
					RetryRate: 0.6,
					PassRate:  0.7,
				},
			},
			{
				Path: "tests/low_pass_test.go",
				RuntimeStats: &models.RuntimeStats{
					PassRate:  0.2,
					RetryRate: 0.0,
				},
			},
		},
		Signals: []models.Signal{
			{
				Type:     "skippedTest",
				Location: models.SignalLocation{File: "tests/skipped_test.go"},
			},
		},
	}

	result := Detect(snap)

	if result.QuarantinedCount != 1 {
		t.Errorf("expected 1 quarantined, got %d", result.QuarantinedCount)
	}
	if result.RetryWrapperCount != 1 {
		t.Errorf("expected 1 retry wrapper, got %d", result.RetryWrapperCount)
	}
	if result.ExpectedFailureCount != 1 {
		t.Errorf("expected 1 expected failure, got %d", result.ExpectedFailureCount)
	}
	if result.SkipDisableCount != 1 {
		t.Errorf("expected 1 skip/disable, got %d", result.SkipDisableCount)
	}
	if result.TotalSuppressedTests != 4 {
		t.Errorf("expected 4 total suppressed tests, got %d", result.TotalSuppressedTests)
	}
}

func TestDetect_IntentClassification(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			// Retry rate >= 0.5 => chronic
			{
				Path: "tests/chronic_retry_test.go",
				RuntimeStats: &models.RuntimeStats{
					RetryRate: 0.55,
					PassRate:  0.9,
				},
			},
			// Retry rate 0.3-0.5 => unknown (not high enough for chronic)
			{
				Path: "tests/moderate_retry_test.go",
				RuntimeStats: &models.RuntimeStats{
					RetryRate: 0.35,
					PassRate:  0.85,
				},
			},
		},
	}

	result := Detect(snap)

	for _, s := range result.Suppressions {
		switch s.TestFilePath {
		case "tests/chronic_retry_test.go":
			if s.Intent != IntentChronic {
				t.Errorf("expected chronic for high retry, got %s", s.Intent)
			}
		case "tests/moderate_retry_test.go":
			if s.Intent != IntentUnknown {
				t.Errorf("expected unknown for moderate retry, got %s", s.Intent)
			}
		}
	}
}

func TestDetect_Deduplication(t *testing.T) {
	t.Parallel()
	// A file that matches both naming and signal should only appear once per kind.
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "tests/skip.auth_test.go"},
		},
		Signals: []models.Signal{
			{
				Type:     "skippedTest",
				Location: models.SignalLocation{File: "tests/skip.auth_test.go"},
			},
		},
	}

	result := Detect(snap)

	// Both naming and signal detect skip_disable for the same file, but dedup
	// should keep only one.
	if result.SkipDisableCount != 1 {
		t.Errorf("expected 1 skip/disable after dedup, got %d", result.SkipDisableCount)
	}
	if result.TotalSuppressedTests != 1 {
		t.Errorf("expected 1 total suppressed test after dedup, got %d", result.TotalSuppressedTests)
	}
}

func TestDetect_SortOrder(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "tests/quarantine/z_test.go"},
			{Path: "tests/quarantine/a_test.go"},
			{Path: "tests/skip.b_test.go"},
		},
	}

	result := Detect(snap)

	if len(result.Suppressions) < 2 {
		t.Fatalf("expected at least 2 suppressions, got %d", len(result.Suppressions))
	}

	// Verify sorted by kind first, then by file path.
	for i := 1; i < len(result.Suppressions); i++ {
		prev := result.Suppressions[i-1]
		curr := result.Suppressions[i]
		if prev.Kind > curr.Kind {
			t.Errorf("suppressions not sorted by kind: %s > %s", prev.Kind, curr.Kind)
		}
		if prev.Kind == curr.Kind && prev.TestFilePath > curr.TestFilePath {
			t.Errorf("suppressions not sorted by path within kind: %s > %s",
				prev.TestFilePath, curr.TestFilePath)
		}
	}
}

func TestDetect_NoRuntimeStats(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "tests/normal_test.go"},
			{Path: "tests/another_test.go", RuntimeStats: nil},
		},
	}

	result := Detect(snap)

	if result.RetryWrapperCount != 0 {
		t.Errorf("expected 0 retry wrappers for files without runtime stats, got %d", result.RetryWrapperCount)
	}
	if result.ExpectedFailureCount != 0 {
		t.Errorf("expected 0 expected failures for files without runtime stats, got %d", result.ExpectedFailureCount)
	}
}

func TestDetect_RetryBelowThreshold(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path: "tests/low_retry_test.go",
				RuntimeStats: &models.RuntimeStats{
					RetryRate: 0.29, // Just below 0.3 threshold.
					PassRate:  0.95,
				},
			},
		},
	}

	result := Detect(snap)

	if result.RetryWrapperCount != 0 {
		t.Errorf("expected 0 retry wrappers for rate below threshold, got %d", result.RetryWrapperCount)
	}
}

func TestDetect_PassRateAtBoundary(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path: "tests/boundary_test.go",
				RuntimeStats: &models.RuntimeStats{
					PassRate: 0.3, // Exactly at boundary (>= 0.3 is not expected failure).
				},
			},
		},
	}

	result := Detect(snap)

	if result.ExpectedFailureCount != 0 {
		t.Errorf("expected 0 expected failures at boundary pass rate 0.3, got %d", result.ExpectedFailureCount)
	}
}

func TestBuildResult_CountsMatch(t *testing.T) {
	t.Parallel()
	suppressions := []Suppression{
		{Kind: KindQuarantined, Intent: IntentChronic, TestFilePath: "a.go"},
		{Kind: KindQuarantined, Intent: IntentChronic, TestFilePath: "b.go"},
		{Kind: KindRetryWrapper, Intent: IntentUnknown, TestFilePath: "c.go"},
		{Kind: KindExpectedFailure, Intent: IntentChronic, TestFilePath: "d.go"},
		{Kind: KindSkipDisable, Intent: IntentUnknown, TestFilePath: "e.go"},
	}

	result := buildResult(suppressions)

	if result.QuarantinedCount != 2 {
		t.Errorf("expected 2 quarantined, got %d", result.QuarantinedCount)
	}
	if result.RetryWrapperCount != 1 {
		t.Errorf("expected 1 retry wrapper, got %d", result.RetryWrapperCount)
	}
	if result.ExpectedFailureCount != 1 {
		t.Errorf("expected 1 expected failure, got %d", result.ExpectedFailureCount)
	}
	if result.SkipDisableCount != 1 {
		t.Errorf("expected 1 skip/disable, got %d", result.SkipDisableCount)
	}
	if result.ChronicCount != 3 {
		t.Errorf("expected 3 chronic, got %d", result.ChronicCount)
	}
	if result.UnknownCount != 2 {
		t.Errorf("expected 2 unknown, got %d", result.UnknownCount)
	}
	if result.TotalSuppressedTests != 5 {
		t.Errorf("expected 5 total suppressed tests, got %d", result.TotalSuppressedTests)
	}
}
