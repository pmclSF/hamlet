package runtime

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// JUnit XML schema types.
// Supports the common JUnit XML format produced by most CI systems,
// Maven Surefire, Gradle, pytest --junitxml, Jest (via jest-junit), etc.

type junitTestSuites struct {
	XMLName xml.Name         `xml:"testsuites"`
	Suites  []junitTestSuite `xml:"testsuite"`
}

type junitTestSuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Errors    int             `xml:"errors,attr"`
	Skipped   int             `xml:"skipped,attr"`
	Time      string          `xml:"time,attr"`
	TestCases []junitTestCase `xml:"testcase"`
}

type junitTestCase struct {
	Name      string        `xml:"name,attr"`
	ClassName string        `xml:"classname,attr"`
	Time      string        `xml:"time,attr"`
	File      string        `xml:"file,attr"`
	Failure   *junitFailure `xml:"failure"`
	Error     *junitError   `xml:"error"`
	Skipped   *junitSkipped `xml:"skipped"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Body    string `xml:",chardata"`
}

type junitError struct {
	Message string `xml:"message,attr"`
	Body    string `xml:",chardata"`
}

type junitSkipped struct {
	Message string `xml:"message,attr"`
}

// ParseJUnitXML parses a JUnit XML file and returns normalized test results.
//
// Supported features:
//   - <testsuites> wrapper (common in CI)
//   - bare <testsuite> (single suite)
//   - test duration in seconds (converted to ms)
//   - failure, error, and skipped states
//   - classname and file attributes
//
// Limitations:
//   - retry/rerun detection is heuristic and based on repeated identity
//     (suite + file/class + test name), with pass-only duplicates treated as
//     duplicate reporting rather than retries
//   - no support for custom properties or attachments
func ParseJUnitXML(path string) (*IngestionResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read JUnit XML: %w", err)
	}

	suites, err := parseJUnitSuites(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JUnit XML from %s: %w", path, err)
	}

	var results []TestResult
	for _, suite := range suites {
		// Track test identity within a suite to detect retries conservatively.
		seenAttempts := map[string]int{}
		seenStatuses := map[string][]TestStatus{}
		for _, tc := range suite.TestCases {
			status := junitStatus(tc)
			key := junitRetryKey(suite.Name, tc)
			attempt := seenAttempts[key]
			retried := false
			if attempt > 0 {
				retried = shouldMarkRetry(seenStatuses[key], status)
			}
			seenAttempts[key]++
			seenStatuses[key] = append(seenStatuses[key], status)

			result := TestResult{
				Name:         tc.Name,
				Suite:        suite.Name,
				File:         resolveFile(tc.File, tc.ClassName),
				DurationMs:   parseSeconds(tc.Time) * 1000,
				Status:       status,
				Retried:      retried,
				RetryAttempt: attempt,
			}

			if tc.Failure != nil {
				result.Message = tc.Failure.Message
			} else if tc.Error != nil {
				result.Message = tc.Error.Message
			}

			results = append(results, result)
		}
	}

	return &IngestionResult{
		Results:    results,
		Format:     "junit-xml",
		SourcePath: path,
	}, nil
}

func junitRetryKey(suite string, tc junitTestCase) string {
	file := resolveFile(tc.File, tc.ClassName)
	if file != "" {
		file = filepath.ToSlash(file)
	}
	return suite + "|" + file + "|" + tc.ClassName + "|" + tc.Name
}

func shouldMarkRetry(previous []TestStatus, current TestStatus) bool {
	if len(previous) == 0 {
		return false
	}
	// If all observed results are passes (including the current one), this is
	// likely duplicate reporting (e.g., shard aggregation) rather than a retry.
	if current == StatusPassed {
		allPassed := true
		for _, s := range previous {
			if s != StatusPassed {
				allPassed = false
				break
			}
		}
		if allPassed {
			return false
		}
	}
	return true
}

func parseJUnitSuites(data []byte) ([]junitTestSuite, error) {
	// Try <testsuites> wrapper first.
	var suites junitTestSuites
	if err := xml.Unmarshal(data, &suites); err == nil && len(suites.Suites) > 0 {
		return suites.Suites, nil
	}

	// Try bare <testsuite>.
	var suite junitTestSuite
	if err := xml.Unmarshal(data, &suite); err == nil && len(suite.TestCases) > 0 {
		return []junitTestSuite{suite}, nil
	}

	return nil, fmt.Errorf("no test suites found in XML")
}

func junitStatus(tc junitTestCase) TestStatus {
	if tc.Skipped != nil {
		return StatusSkipped
	}
	if tc.Failure != nil {
		return StatusFailed
	}
	if tc.Error != nil {
		return StatusError
	}
	return StatusPassed
}

func parseSeconds(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// resolveFile returns the best available file path.
// Some JUnit producers set the file attribute; others only set classname
// which can be used as a fallback for display.
func resolveFile(file, className string) string {
	if file != "" {
		return file
	}
	return className
}
