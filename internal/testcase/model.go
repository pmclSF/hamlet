// Package testcase defines the normalized test case model and extraction logic.
//
// Each discovered test case receives a deterministic, stable test ID
// suitable for snapshots, trend analysis, flake tracking, and coverage
// attribution. Identity is based on structural properties (path, suite
// hierarchy, test name), never on traversal order, line numbers, or
// random values.
package testcase

// TestCase represents a single discovered test case with a stable identity.
type TestCase struct {
	// TestID is the deterministic hash of CanonicalIdentity.
	TestID string `json:"testId"`

	// CanonicalIdentity is the human-readable normalized identity string.
	// Format: path::suite_hierarchy::test_name[::param_signature]
	CanonicalIdentity string `json:"canonicalIdentity"`

	// FilePath is the repository-relative path to the test file.
	FilePath string `json:"filePath"`

	// SuiteHierarchy is the normalized chain of describe/suite/class names.
	// For flat test files, this may be empty.
	SuiteHierarchy []string `json:"suiteHierarchy,omitempty"`

	// TestName is the normalized test name.
	TestName string `json:"testName"`

	// Framework is the detected framework for this test.
	Framework string `json:"framework"`

	// Language is the programming language.
	Language string `json:"language"`

	// Line is the approximate line number where the test is defined.
	// This is metadata only — NOT part of identity.
	Line int `json:"line,omitempty"`

	// ExtractionKind describes how the test was discovered.
	ExtractionKind ExtractionKind `json:"extractionKind"`

	// Parameterized contains parameterization metadata if applicable.
	Parameterized *ParameterizationInfo `json:"parameterized,omitempty"`

	// Confidence indicates how confident the extraction is.
	// 1.0 = high confidence static extraction.
	// Lower values for dynamic or ambiguous patterns.
	Confidence float64 `json:"confidence"`
}

// ExtractionKind describes how a test case was discovered.
type ExtractionKind string

const (
	// ExtractionStatic is a statically identifiable test definition.
	ExtractionStatic ExtractionKind = "static"

	// ExtractionParameterizedTemplate is the template of a parameterized test.
	ExtractionParameterizedTemplate ExtractionKind = "parameterized_template"

	// ExtractionDynamic is a dynamically generated test (e.g., in a loop).
	ExtractionDynamic ExtractionKind = "dynamic"

	// ExtractionAmbiguous is a test that could not be clearly classified.
	ExtractionAmbiguous ExtractionKind = "ambiguous"
)

// ParameterizationInfo captures metadata about parameterized tests.
type ParameterizationInfo struct {
	// IsTemplate indicates this is the template definition, not an instance.
	IsTemplate bool `json:"isTemplate"`

	// ParamSignature is the stable parameter descriptor if statically available.
	// For test.each, this might be the table header or value shape.
	ParamSignature string `json:"paramSignature,omitempty"`

	// EstimatedInstances is the estimated number of parameter combinations
	// if statically determinable.
	EstimatedInstances int `json:"estimatedInstances,omitempty"`
}
