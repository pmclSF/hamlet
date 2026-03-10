package depgraph

// EdgeType classifies the relationship between two nodes.
type EdgeType string

const (
	// Test structure edges.
	EdgeTestDefinedInFile EdgeType = "test_defined_in_file"
	EdgeSuiteContainsTest EdgeType = "suite_contains_test"

	// Dependency edges.
	EdgeImportsModule        EdgeType = "imports_module"
	EdgeSourceImportsSource  EdgeType = "source_imports_source"
	EdgeTestUsesFixture      EdgeType = "test_uses_fixture"
	EdgeTestUsesHelper       EdgeType = "test_uses_helper"
	EdgeFixtureImportsSource EdgeType = "fixture_imports_source"
	EdgeHelperImportsSource  EdgeType = "helper_imports_source"

	// Package edges.
	EdgeBelongsToPackage EdgeType = "belongs_to_package"
)

// EvidenceType describes how an edge was discovered.
type EvidenceType string

const (
	EvidenceStaticAnalysis EvidenceType = "static_analysis"
	EvidenceConvention     EvidenceType = "convention"
	EvidenceInferred       EvidenceType = "inferred"
	EvidenceManual         EvidenceType = "manual"
)

// Edge is a directed relationship between two nodes.
type Edge struct {
	// From is the source node ID.
	From string `json:"from"`

	// To is the target node ID.
	To string `json:"to"`

	// Type classifies this relationship.
	Type EdgeType `json:"type"`

	// Confidence is the strength of this relationship (0.0–1.0).
	Confidence float64 `json:"confidence"`

	// EvidenceType describes how this edge was discovered.
	EvidenceType EvidenceType `json:"evidenceType"`
}
