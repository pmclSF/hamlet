// Package depgraph provides a typed dependency graph that models relationships
// between tests, source files, fixtures, helpers, suites, and packages.
//
// Unlike the flat index in internal/graph (which indexes snapshot data for
// lookups), depgraph builds a true graph with typed nodes and edges that
// supports traversal algorithms: coverage analysis via reverse edges, impact
// propagation via BFS with confidence decay, fanout measurement via
// transitive closure, and structural duplicate detection via fingerprinting.
//
// Design constraints:
//   - Read-only after Build(): construct once, then query
//   - No external dependencies: depends only on models
//   - Deterministic: iteration order is canonical where it matters
//   - Lightweight: adjacency lists, no graph database
package depgraph

// NodeFamily groups node types into conceptual layers.
type NodeFamily string

const (
	FamilySystem     NodeFamily = "system"
	FamilyValidation NodeFamily = "validation"
	FamilyBehavior   NodeFamily = "behavior"
	FamilyEnvironment NodeFamily = "environment"
	FamilyExecution  NodeFamily = "execution"
	FamilyGovernance NodeFamily = "governance"
)

// NodeType classifies graph nodes.
type NodeType string

// --- System topology nodes ---
//
// Represent the structural elements of the codebase: source files
// and code surfaces (functions, methods, endpoints).
const (
	NodeSourceFile  NodeType = "source_file"
	NodeCodeSurface NodeType = "code_surface"
)

// --- Validation topology nodes ---
//
// Represent the test system: tests, suites, validation targets,
// and the files they validate.
const (
	NodeValidationTarget NodeType = "validation_target"
	NodeTest             NodeType = "test"
	NodeScenario         NodeType = "scenario"
	NodeManualCoverage   NodeType = "manual_coverage"
	NodeSuite            NodeType = "suite"
	NodeTestFile         NodeType = "test_file"
	NodeFixture          NodeType = "fixture"
)

// --- Behavior topology nodes ---
//
// Represent inferred behavioral surfaces derived from code structure.
const (
	NodeBehaviorSurface NodeType = "behavior_surface"
)

// --- Environment topology nodes ---
//
// Represent execution environments, device configurations,
// external services, and AI/ML-specific nodes.
const (
	NodeEnvironment      NodeType = "environment"
	NodeEnvironmentClass NodeType = "environment_class"
	NodeDeviceConfig     NodeType = "device_config"
	NodeDataset          NodeType = "dataset"
	NodeModel            NodeType = "model"
	NodePrompt           NodeType = "prompt"
	NodeEvalMetric       NodeType = "eval_metric"
)

// --- Execution topology nodes ---
//
// Represent runtime execution state: execution runs and validation results.
const (
	NodeExecutionRun        NodeType = "execution_run"
	NodeValidationExecution NodeType = "validation_execution"
)

// --- Governance topology nodes ---
//
// Represent ownership, policy, and release governance.
const (
	NodeOwner      NodeType = "owner"
	NodeCapability NodeType = "capability"
)

// NodeTypeFamily returns the family a node type belongs to.
func NodeTypeFamily(t NodeType) NodeFamily {
	switch t {
	// System
	case NodeSourceFile, NodeCodeSurface:
		return FamilySystem

	// Validation
	case NodeValidationTarget, NodeTest, NodeScenario, NodeManualCoverage,
		NodeSuite, NodeTestFile, NodeFixture:
		return FamilyValidation

	// Behavior
	case NodeBehaviorSurface:
		return FamilyBehavior

	// Environment
	case NodeEnvironment, NodeEnvironmentClass, NodeDeviceConfig,
		NodeDataset, NodeModel, NodePrompt, NodeEvalMetric:
		return FamilyEnvironment

	// Execution
	case NodeExecutionRun, NodeValidationExecution:
		return FamilyExecution

	// Governance
	case NodeOwner, NodeCapability:
		return FamilyGovernance

	default:
		return ""
	}
}

// AllNodeTypes returns all registered node types grouped by family.
// This is useful for documentation, validation, and schema export.
func AllNodeTypes() map[NodeFamily][]NodeType {
	return map[NodeFamily][]NodeType{
		FamilySystem: {
			NodeSourceFile, NodeCodeSurface,
		},
		FamilyValidation: {
			NodeValidationTarget, NodeTest, NodeScenario, NodeManualCoverage,
			NodeSuite, NodeTestFile, NodeFixture,
		},
		FamilyBehavior: {
			NodeBehaviorSurface,
		},
		FamilyEnvironment: {
			NodeEnvironment, NodeEnvironmentClass, NodeDeviceConfig,
			NodeDataset, NodeModel, NodePrompt, NodeEvalMetric,
		},
		FamilyExecution: {
			NodeExecutionRun, NodeValidationExecution,
		},
		FamilyGovernance: {
			NodeOwner,
		},
	}
}

// Node is a vertex in the dependency graph.
type Node struct {
	// ID uniquely identifies this node.
	// Convention: "type:path" or "type:path:line:name" for tests.
	ID string `json:"id"`

	// Type classifies this node.
	Type NodeType `json:"type"`

	// Path is the repository-relative file path, if applicable.
	Path string `json:"path,omitempty"`

	// Name is a human-readable label (test name, function name, etc.).
	Name string `json:"name,omitempty"`

	// Line is the source line number for tests and suites.
	Line int `json:"line,omitempty"`

	// Package is the containing package or module, if known.
	Package string `json:"package,omitempty"`

	// Framework is the test framework (jest, vitest, pytest, etc.).
	Framework string `json:"framework,omitempty"`

	// Metadata holds additional key-value pairs for extensibility.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Family returns the NodeFamily this node belongs to.
func (n *Node) Family() NodeFamily {
	return NodeTypeFamily(n.Type)
}
