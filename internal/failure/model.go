// Package failure provides taxonomy and classification of test failures
// into meaningful categories based on error messages and stack traces.
//
// The classifier uses pattern matching to assign each failure a category
// (e.g., assertion_failure, timeout, infrastructure_or_environment) along
// with a confidence level, enabling downstream consumers to prioritize
// remediation and detect systemic failure patterns.
package failure

// FailureCategory classifies the type of test failure.
type FailureCategory string

const (
	CategoryAssertionFailure  FailureCategory = "assertion_failure"
	CategoryTimeout           FailureCategory = "timeout"
	CategorySetupFixture      FailureCategory = "setup_or_fixture_failure"
	CategoryDependencyService FailureCategory = "dependency_or_service_failure"
	CategorySnapshotMismatch  FailureCategory = "snapshot_mismatch"
	CategorySelectorUI        FailureCategory = "selector_or_ui_fragility"
	CategoryInfraEnvironment  FailureCategory = "infrastructure_or_environment"
	CategoryUnknown           FailureCategory = "unknown"
)

// ClassificationConfidence indicates how confident the classification is.
type ClassificationConfidence string

const (
	ConfidenceExact    ClassificationConfidence = "exact"    // Unambiguous match
	ConfidenceInferred ClassificationConfidence = "inferred" // Pattern-based
	ConfidenceWeak     ClassificationConfidence = "weak"     // Low-signal guess
)

// FailureClassification represents a classified failure.
type FailureClassification struct {
	// TestFilePath is the file containing the failing test.
	TestFilePath string
	// TestName is the specific test, if identifiable.
	TestName string
	// Category is the failure classification.
	Category FailureCategory
	// Confidence in the classification.
	Confidence ClassificationConfidence
	// ConfidenceScore is a numeric confidence (0.0-1.0).
	ConfidenceScore float64
	// Explanation describes why this category was assigned.
	Explanation string
	// ErrorMessage is the original error text, if available.
	ErrorMessage string
	// StackTrace is the original stack, if available.
	StackTrace string
}

// TaxonomyResult holds all failure classifications for a snapshot.
type TaxonomyResult struct {
	Classifications []FailureClassification
	// ByCategory counts classifications by category.
	ByCategory map[FailureCategory]int
	// TotalFailures is the count of classified failures.
	TotalFailures int
	// DominantCategory is the most common failure category.
	DominantCategory FailureCategory
}

// FailureInput represents a single failure to be classified.
// This is the input boundary for the classifier, decoupled from any
// specific runtime artifact format.
type FailureInput struct {
	// TestFilePath is the file containing the failing test.
	TestFilePath string
	// TestName is the specific test, if identifiable.
	TestName string
	// ErrorMessage is the error or failure message.
	ErrorMessage string
	// StackTrace is the stack trace, if available.
	StackTrace string
}
