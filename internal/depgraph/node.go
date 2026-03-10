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

// NodeType classifies graph nodes.
type NodeType string

const (
	NodeTest              NodeType = "test"
	NodeSuite             NodeType = "suite"
	NodeTestFile          NodeType = "test_file"
	NodeSourceFile        NodeType = "source_file"
	NodeFixture           NodeType = "fixture"
	NodeHelper            NodeType = "helper"
	NodePackage           NodeType = "package"
	NodeService           NodeType = "service"
	NodeGeneratedArtifact NodeType = "generated_artifact"
	NodeConfigArtifact    NodeType = "config_artifact"
	NodeManualCoverage    NodeType = "manual_coverage"
)

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
