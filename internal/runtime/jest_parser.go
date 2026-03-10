package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Jest/Vitest JSON result schema.
// Produced by `jest --json` or `vitest --reporter=json`.

type jestResults struct {
	NumTotalTests  int              `json:"numTotalTests"`
	NumPassedTests int              `json:"numPassedTests"`
	NumFailedTests int              `json:"numFailedTests"`
	Success        bool             `json:"success"`
	TestResults    []jestTestResult `json:"testResults"`
}

type jestTestResult struct {
	TestFilePath     string          `json:"testFilePath"`
	NumPassing       int             `json:"numPassingTests"`
	NumFailing       int             `json:"numFailingTests"`
	NumPending       int             `json:"numPendingTests"`
	AssertionResults []jestAssertion `json:"assertionResults"`
	// Some Jest versions use "testResults" inside each file result.
	TestResults []jestAssertion `json:"testResults,omitempty"`
}

type jestAssertion struct {
	FullName        string   `json:"fullName"`
	Title           string   `json:"title"`
	AncestorTitles  []string `json:"ancestorTitles"`
	Status          string   `json:"status"`   // "passed", "failed", "pending", "skipped"
	Duration        *float64 `json:"duration"` // milliseconds, nullable
	FailureMessages []string `json:"failureMessages"`
	// Vitest uses "retryCount" for retry information.
	RetryCount int `json:"retryCount,omitempty"`
}

// ParseJestJSON parses a Jest/Vitest JSON results file.
//
// Supported features:
//   - standard jest --json output format
//   - vitest --reporter=json output format
//   - per-test duration (already in ms)
//   - retry detection via retryCount field (Vitest)
//   - pending/skipped test detection
//
// Limitations:
//   - Jest does not natively report retry counts; only Vitest does
//   - File paths are absolute as reported by the runner
func ParseJestJSON(path string) (*IngestionResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read Jest JSON: %w", err)
	}

	var jr jestResults
	if err := json.Unmarshal(data, &jr); err != nil {
		return nil, fmt.Errorf("failed to parse Jest JSON from %s: %w", path, err)
	}

	var results []TestResult
	for _, fileResult := range jr.TestResults {
		assertions := fileResult.AssertionResults
		if len(assertions) == 0 {
			assertions = fileResult.TestResults
		}

		for _, a := range assertions {
			name := a.FullName
			if name == "" {
				name = a.Title
			}

			suite := ""
			if len(a.AncestorTitles) > 0 {
				suite = joinNonEmpty(a.AncestorTitles, " > ")
			}

			var durationMs float64
			if a.Duration != nil {
				durationMs = *a.Duration
			}

			msg := ""
			if len(a.FailureMessages) > 0 {
				msg = a.FailureMessages[0]
				// Truncate very long failure messages for signal display.
				if len(msg) > 500 {
					msg = msg[:500] + "..."
				}
			}

			results = append(results, TestResult{
				Name:         name,
				Suite:        suite,
				File:         fileResult.TestFilePath,
				DurationMs:   durationMs,
				Status:       jestStatus(a.Status),
				Retried:      a.RetryCount > 0,
				RetryAttempt: a.RetryCount,
				Message:      msg,
			})
		}
	}

	return &IngestionResult{
		Results:    results,
		Format:     "jest-json",
		SourcePath: path,
	}, nil
}

func joinNonEmpty(parts []string, sep string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return ""
	}
	return strings.Join(out, sep)
}

func jestStatus(s string) TestStatus {
	switch s {
	case "passed":
		return StatusPassed
	case "failed":
		return StatusFailed
	case "pending", "skipped", "todo":
		return StatusSkipped
	default:
		return StatusPassed
	}
}
