// Package runtime provides ingestion of test runtime artifacts (JUnit XML,
// Jest/Vitest JSON) into a normalized model for health signal detection.
//
// The package is intentionally format-agnostic at its boundary: parsers
// produce []TestResult values, which downstream health detectors consume
// without knowing the original artifact format.
package runtime

// TestResult is the normalized representation of a single test execution.
// It is backend-format-agnostic and suitable for health detection.
type TestResult struct {
	// Name is the test name as reported by the runner.
	Name string

	// Suite is the test suite or class name if available.
	Suite string

	// File is the source file path if available.
	File string

	// DurationMs is the execution duration in milliseconds.
	DurationMs float64

	// Status is the execution outcome.
	Status TestStatus

	// Retried indicates whether this test was retried (evidence for flakiness).
	Retried bool

	// RetryAttempt is the retry sequence number (0 = first run).
	RetryAttempt int

	// Message is the failure or error message if applicable.
	Message string
}

// TestStatus represents the outcome of a test execution.
type TestStatus string

const (
	StatusPassed  TestStatus = "passed"
	StatusFailed  TestStatus = "failed"
	StatusSkipped TestStatus = "skipped"
	StatusError   TestStatus = "error"
)

// IngestionResult holds the normalized output of parsing one or more
// runtime artifacts.
type IngestionResult struct {
	// Results are the individual test execution records.
	Results []TestResult

	// Format describes the source format (e.g., "junit-xml", "jest-json").
	Format string

	// SourcePath is the file path of the ingested artifact.
	SourcePath string
}
