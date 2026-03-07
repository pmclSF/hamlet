package models

// TestCase represents a single discovered test case with a stable identity.
//
// This model enables longitudinal tracking: every test can be compared
// across snapshots by its deterministic TestID.
type TestCase struct {
	// TestID is the deterministic hash of CanonicalIdentity.
	TestID string `json:"testId"`

	// CanonicalIdentity is the human-readable normalized identity string.
	// Format: path::suite_hierarchy::test_name[::param_signature]
	CanonicalIdentity string `json:"canonicalIdentity"`

	// FilePath is the repository-relative path to the test file.
	FilePath string `json:"filePath"`

	// SuiteHierarchy is the normalized chain of describe/suite/class names.
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
	// Values: "static", "parameterized_template", "dynamic", "ambiguous".
	ExtractionKind string `json:"extractionKind"`

	// Confidence indicates how confident the extraction is (0.0–1.0).
	Confidence float64 `json:"confidence"`

	// TestType is the inferred test type (unit, integration, e2e, etc.).
	// Empty if not yet inferred.
	TestType string `json:"testType,omitempty"`

	// TestTypeConfidence indicates confidence in the type inference.
	TestTypeConfidence float64 `json:"testTypeConfidence,omitempty"`

	// TestTypeEvidence lists reasons for the type classification.
	TestTypeEvidence []string `json:"testTypeEvidence,omitempty"`

	// Parameterized contains parameterization metadata if applicable.
	Parameterized *ParameterizationInfo `json:"parameterized,omitempty"`
}

// ParameterizationInfo captures metadata about parameterized tests.
type ParameterizationInfo struct {
	// IsTemplate indicates this is the template definition, not an instance.
	IsTemplate bool `json:"isTemplate"`

	// ParamSignature is the stable parameter descriptor if statically available.
	ParamSignature string `json:"paramSignature,omitempty"`

	// EstimatedInstances is the estimated number of parameter combinations.
	EstimatedInstances int `json:"estimatedInstances,omitempty"`
}
