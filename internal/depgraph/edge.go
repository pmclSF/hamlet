package depgraph

// EdgeType classifies the relationship between two nodes.
type EdgeType string

// --- Test structure edges ---
//
// Connect tests to the files and suites that contain them.
const (
	EdgeTestDefinedInFile EdgeType = "test_defined_in_file"
	EdgeSuiteContainsTest EdgeType = "suite_contains_test"
)

// --- Dependency edges ---
//
// Connect files and modules through import and usage relationships.
const (
	EdgeImportsModule       EdgeType = "imports_module"
	EdgeSourceImportsSource EdgeType = "source_imports_source"
)

// --- Package edges ---
const (
	EdgeBelongsToPackage EdgeType = "belongs_to_package"
)

// --- Validation edges ---
//
// Connect validation nodes to the system elements they validate.
const (
	EdgeCoversCodeSurface EdgeType = "covers_code_surface"
	EdgeManualCovers      EdgeType = "manual_covers"
)

// --- Behavior edges ---
//
// Connect behavior surfaces to the code they are derived from.
const (
	EdgeBehaviorDerivedFrom EdgeType = "behavior_derived_from"
)

// --- Fixture edges ---
//
// Connect fixtures to the tests that consume them and the code surfaces
// they set up. High in-degree on a fixture node indicates fragility risk.
const (
	EdgeTestUsesFixture    EdgeType = "test_uses_fixture"
	EdgeFixtureSetsSurface EdgeType = "fixture_sets_surface"
)

// --- Environment edges ---
//
// Connect tests and executions to their target environments,
// external services, and AI/ML resources.
const (
	EdgeTargetsEnvironment       EdgeType = "targets_environment"
	EdgeEnvironmentClassContains EdgeType = "environment_class_contains"
	EdgeUsesDataset              EdgeType = "uses_dataset"
	EdgeUsesModel                EdgeType = "uses_model"
	EdgeUsesPrompt               EdgeType = "uses_prompt"
	EdgeEvaluatesMetric          EdgeType = "evaluates_metric"
)

// --- Execution edges ---
//
// Connect execution runs to their constituent validation executions.
const (
	EdgeExecutionRunsTest EdgeType = "execution_runs_test"
)

// --- Governance edges ---
//
// Connect owners and capabilities to the elements they govern.
const (
	EdgeOwns                      EdgeType = "owns"
	EdgeScenarioValidatesCapability EdgeType = "scenario_validates_capability"
)

// EvidenceType describes how an edge was discovered.
type EvidenceType string

const (
	EvidenceStaticAnalysis EvidenceType = "static_analysis"
	EvidenceConvention     EvidenceType = "convention"
	EvidenceInferred       EvidenceType = "inferred"
	EvidenceManual         EvidenceType = "manual"
	EvidenceExecution      EvidenceType = "execution"
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
