package depgraph

import (
	"encoding/json"
	"testing"
)

// --- NodeFamily Tests ---

func TestNodeTypeFamily(t *testing.T) {
	t.Parallel()

	tests := []struct {
		nodeType NodeType
		want     NodeFamily
	}{
		// System
		{NodeSourceFile, FamilySystem},
		{NodePackage, FamilySystem},
		{NodeService, FamilySystem},
		{NodeGeneratedArtifact, FamilySystem},
		{NodeCodeSurface, FamilySystem},

		// Validation
		{NodeValidationTarget, FamilyValidation},
		{NodeTest, FamilyValidation},
		{NodeScenario, FamilyValidation},
		{NodeManualCoverage, FamilyValidation},
		{NodeSuite, FamilyValidation},
		{NodeTestFile, FamilyValidation},
		{NodeFixture, FamilyValidation},
		{NodeHelper, FamilyValidation},

		// Behavior
		{NodeBehaviorSurface, FamilyBehavior},

		// Environment
		{NodeEnvironment, FamilyEnvironment},
		{NodeEnvironmentClass, FamilyEnvironment},
		{NodeDeviceConfig, FamilyEnvironment},
		{NodeExternalService, FamilyEnvironment},
		{NodeDataset, FamilyEnvironment},
		{NodeModel, FamilyEnvironment},
		{NodePrompt, FamilyEnvironment},
		{NodeEvalMetric, FamilyEnvironment},

		// Execution
		{NodeExecutionRun, FamilyExecution},
		{NodeValidationExecution, FamilyExecution},

		// Governance
		{NodeOwner, FamilyGovernance},

		// Unknown
		{NodeType("unknown_type"), ""},
	}

	for _, tt := range tests {
		got := NodeTypeFamily(tt.nodeType)
		if got != tt.want {
			t.Errorf("NodeTypeFamily(%q) = %q, want %q", tt.nodeType, got, tt.want)
		}
	}
}

func TestNodeFamily_Method(t *testing.T) {
	t.Parallel()
	n := &Node{ID: "test:a:1:login", Type: NodeTest}
	if n.Family() != FamilyValidation {
		t.Errorf("expected validation family, got %q", n.Family())
	}

	s := &Node{ID: "file:src/auth.go", Type: NodeSourceFile}
	if s.Family() != FamilySystem {
		t.Errorf("expected system family, got %q", s.Family())
	}
}

func TestAllNodeTypes_Coverage(t *testing.T) {
	t.Parallel()
	all := AllNodeTypes()

	// Every family should be present.
	families := []NodeFamily{
		FamilySystem, FamilyValidation, FamilyBehavior,
		FamilyEnvironment, FamilyExecution, FamilyGovernance,
	}
	for _, f := range families {
		types, ok := all[f]
		if !ok {
			t.Errorf("family %q missing from AllNodeTypes()", f)
			continue
		}
		if len(types) == 0 {
			t.Errorf("family %q has zero types", f)
		}

		// Verify each type maps back to its family.
		for _, nt := range types {
			if NodeTypeFamily(nt) != f {
				t.Errorf("NodeTypeFamily(%q) = %q, want %q", nt, NodeTypeFamily(nt), f)
			}
		}
	}
}

// --- NodesByFamily Tests ---

func TestNodesByFamily(t *testing.T) {
	t.Parallel()
	g := NewGraph()

	g.AddNode(&Node{ID: "file:src/auth.go", Type: NodeSourceFile})
	g.AddNode(&Node{ID: "pkg:auth", Type: NodePackage})
	g.AddNode(&Node{ID: "test:a:1:login", Type: NodeTest})
	g.AddNode(&Node{ID: "file:a.test.go", Type: NodeTestFile})
	g.AddNode(&Node{ID: "env:staging", Type: NodeEnvironment})
	g.AddNode(&Node{ID: "owner:team-alpha", Type: NodeOwner})

	system := g.NodesByFamily(FamilySystem)
	if len(system) != 2 {
		t.Errorf("expected 2 system nodes, got %d", len(system))
	}

	validation := g.NodesByFamily(FamilyValidation)
	if len(validation) != 2 {
		t.Errorf("expected 2 validation nodes, got %d", len(validation))
	}

	env := g.NodesByFamily(FamilyEnvironment)
	if len(env) != 1 {
		t.Errorf("expected 1 environment node, got %d", len(env))
	}

	gov := g.NodesByFamily(FamilyGovernance)
	if len(gov) != 1 {
		t.Errorf("expected 1 governance node, got %d", len(gov))
	}

	// Should be sorted by ID.
	if len(system) == 2 && system[0].ID >= system[1].ID {
		t.Errorf("system nodes not sorted: %q >= %q", system[0].ID, system[1].ID)
	}
}

// --- Serialization Tests ---

func TestGraphMarshalJSON(t *testing.T) {
	t.Parallel()
	g := NewGraph()

	g.AddNode(&Node{ID: "file:src/auth.go", Type: NodeSourceFile, Path: "src/auth.go", Name: "auth.go"})
	g.AddNode(&Node{ID: "test:a:1:login", Type: NodeTest, Path: "a.test.go", Name: "login", Line: 1})
	g.AddEdge(&Edge{
		From:         "test:a:1:login",
		To:           "file:src/auth.go",
		Type:         EdgeValidates,
		Confidence:   0.9,
		EvidenceType: EvidenceStaticAnalysis,
	})

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Verify it's valid JSON with expected structure.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if _, ok := raw["version"]; !ok {
		t.Error("missing version field")
	}
	if _, ok := raw["nodes"]; !ok {
		t.Error("missing nodes field")
	}
	if _, ok := raw["edges"]; !ok {
		t.Error("missing edges field")
	}
}

func TestGraphUnmarshalJSON(t *testing.T) {
	t.Parallel()

	input := `{
		"version": "1.0.0",
		"nodes": [
			{"id": "file:src/auth.go", "type": "source_file", "path": "src/auth.go"},
			{"id": "test:a:1:login", "type": "test", "path": "a.test.go", "name": "login", "line": 1}
		],
		"edges": [
			{"from": "test:a:1:login", "to": "file:src/auth.go", "type": "validates", "confidence": 0.9, "evidenceType": "static_analysis"}
		]
	}`

	g := &Graph{}
	if err := json.Unmarshal([]byte(input), g); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	if g.NodeCount() != 2 {
		t.Errorf("expected 2 nodes, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 1 {
		t.Errorf("expected 1 edge, got %d", g.EdgeCount())
	}

	// Verify adjacency indexes rebuilt.
	out := g.Outgoing("test:a:1:login")
	if len(out) != 1 {
		t.Fatalf("expected 1 outgoing edge, got %d", len(out))
	}
	if out[0].Type != EdgeValidates {
		t.Errorf("expected validates edge, got %s", out[0].Type)
	}

	inc := g.Incoming("file:src/auth.go")
	if len(inc) != 1 {
		t.Fatalf("expected 1 incoming edge, got %d", len(inc))
	}
}

func TestGraphSerializationRoundTrip(t *testing.T) {
	t.Parallel()
	original := NewGraph()

	// Add nodes from multiple families.
	original.AddNode(&Node{ID: "file:src/auth.go", Type: NodeSourceFile, Path: "src/auth.go"})
	original.AddNode(&Node{ID: "svc:auth-api", Type: NodeService, Name: "auth-api"})
	original.AddNode(&Node{ID: "test:a:1:login", Type: NodeTest, Path: "a.test.go", Name: "login", Line: 1, Framework: "go"})
	original.AddNode(&Node{ID: "env:staging", Type: NodeEnvironment, Name: "staging", Metadata: map[string]string{"region": "us-east-1"}})
	original.AddNode(&Node{ID: "behavior:auth-flow", Type: NodeBehaviorSurface, Name: "auth-flow"})
	original.AddNode(&Node{ID: "owner:team-alpha", Type: NodeOwner, Name: "team-alpha"})
	original.AddNode(&Node{ID: "run:ci-123", Type: NodeExecutionRun, Name: "ci-123"})

	// Add cross-family edges.
	original.AddEdge(&Edge{From: "test:a:1:login", To: "file:src/auth.go", Type: EdgeValidates, Confidence: 0.95, EvidenceType: EvidenceStaticAnalysis})
	original.AddEdge(&Edge{From: "test:a:1:login", To: "env:staging", Type: EdgeTargetsEnvironment, Confidence: 0.8, EvidenceType: EvidenceConvention})
	original.AddEdge(&Edge{From: "behavior:auth-flow", To: "file:src/auth.go", Type: EdgeBehaviorDerivedFrom, Confidence: 0.7, EvidenceType: EvidenceInferred})
	original.AddEdge(&Edge{From: "owner:team-alpha", To: "file:src/auth.go", Type: EdgeOwns, Confidence: 1.0, EvidenceType: EvidenceManual})
	original.AddEdge(&Edge{From: "run:ci-123", To: "test:a:1:login", Type: EdgeExecutionRunsTest, Confidence: 1.0, EvidenceType: EvidenceExecution})

	// Serialize.
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	// Deserialize.
	restored := &Graph{}
	if err := json.Unmarshal(data, restored); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	// Verify node count and types.
	if restored.NodeCount() != original.NodeCount() {
		t.Errorf("node count: got %d, want %d", restored.NodeCount(), original.NodeCount())
	}
	if restored.EdgeCount() != original.EdgeCount() {
		t.Errorf("edge count: got %d, want %d", restored.EdgeCount(), original.EdgeCount())
	}

	// Verify specific nodes survived.
	env := restored.Node("env:staging")
	if env == nil {
		t.Fatal("env:staging not found after roundtrip")
	}
	if env.Metadata["region"] != "us-east-1" {
		t.Errorf("metadata lost: got %q, want us-east-1", env.Metadata["region"])
	}
	if env.Family() != FamilyEnvironment {
		t.Errorf("family wrong after roundtrip: got %q", env.Family())
	}

	// Verify adjacency indexes rebuilt.
	outgoing := restored.Outgoing("test:a:1:login")
	if len(outgoing) != 2 {
		t.Errorf("expected 2 outgoing edges from test, got %d", len(outgoing))
	}

	incoming := restored.Incoming("file:src/auth.go")
	if len(incoming) != 3 {
		t.Errorf("expected 3 incoming edges to source file, got %d", len(incoming))
	}

	// Verify stats include family counts.
	stats := restored.Stats()
	if stats.NodesByFamily["system"] != 2 {
		t.Errorf("expected 2 system nodes in stats, got %d", stats.NodesByFamily["system"])
	}
	if stats.NodesByFamily["validation"] != 1 {
		t.Errorf("expected 1 validation node in stats, got %d", stats.NodesByFamily["validation"])
	}
}

func TestGraphMarshalJSON_Empty(t *testing.T) {
	t.Parallel()
	g := NewGraph()

	data, err := json.Marshal(g)
	if err != nil {
		t.Fatalf("MarshalJSON failed for empty graph: %v", err)
	}

	restored := &Graph{}
	if err := json.Unmarshal(data, restored); err != nil {
		t.Fatalf("UnmarshalJSON failed for empty graph: %v", err)
	}

	if restored.NodeCount() != 0 {
		t.Errorf("expected 0 nodes, got %d", restored.NodeCount())
	}
	if restored.EdgeCount() != 0 {
		t.Errorf("expected 0 edges, got %d", restored.EdgeCount())
	}
}

// --- Cross-family traversal tests ---

func TestCrossFamilyTraversal(t *testing.T) {
	t.Parallel()
	g := buildUnifiedTestGraph()

	// Traversal: from a changed source file, find all impacted tests
	// (validation family) through the graph.
	sourceID := "file:src/auth.go"
	incoming := g.Incoming(sourceID)

	var validationEdges []*Edge
	for _, e := range incoming {
		fromNode := g.Node(e.From)
		if fromNode != nil && fromNode.Family() == FamilyValidation {
			validationEdges = append(validationEdges, e)
		}
	}

	if len(validationEdges) != 2 {
		t.Errorf("expected 2 validation edges to source, got %d", len(validationEdges))
	}

	// Traversal: from a test, find all environments it targets.
	testID := "test:a:1:login"
	outgoing := g.Outgoing(testID)

	var envEdges []*Edge
	for _, e := range outgoing {
		toNode := g.Node(e.To)
		if toNode != nil && toNode.Family() == FamilyEnvironment {
			envEdges = append(envEdges, e)
		}
	}

	if len(envEdges) != 1 {
		t.Errorf("expected 1 environment edge from test, got %d", len(envEdges))
	}

	// Traversal: from an owner, find all owned source files.
	ownerID := "owner:team-alpha"
	ownedEdges := g.Outgoing(ownerID)
	var ownedSources []*Edge
	for _, e := range ownedEdges {
		if e.Type == EdgeOwns {
			ownedSources = append(ownedSources, e)
		}
	}
	if len(ownedSources) != 1 {
		t.Errorf("expected 1 owns edge, got %d", len(ownedSources))
	}
}

func TestStatsIncludesFamilyCounts(t *testing.T) {
	t.Parallel()
	g := buildUnifiedTestGraph()

	stats := g.Stats()

	if stats.NodesByFamily["system"] != 2 {
		t.Errorf("expected 2 system nodes, got %d", stats.NodesByFamily["system"])
	}
	if stats.NodesByFamily["validation"] != 3 {
		t.Errorf("expected 3 validation nodes, got %d", stats.NodesByFamily["validation"])
	}
	if stats.NodesByFamily["environment"] != 1 {
		t.Errorf("expected 1 environment node, got %d", stats.NodesByFamily["environment"])
	}
	if stats.NodesByFamily["behavior"] != 1 {
		t.Errorf("expected 1 behavior node, got %d", stats.NodesByFamily["behavior"])
	}
	if stats.NodesByFamily["governance"] != 1 {
		t.Errorf("expected 1 governance node, got %d", stats.NodesByFamily["governance"])
	}
	if stats.NodesByFamily["execution"] != 1 {
		t.Errorf("expected 1 execution node, got %d", stats.NodesByFamily["execution"])
	}
}

func TestNewNodeTypesInExistingEngines(t *testing.T) {
	t.Parallel()

	// Verify that adding new node types doesn't break existing analysis
	// engines. The engines only operate on types they recognize (NodeTest,
	// NodeTestFile, NodeSourceFile, etc.) and ignore unknown types.
	g := NewGraph()

	// Add traditional nodes.
	g.AddNode(&Node{ID: "file:a.test.js", Type: NodeTestFile, Path: "a.test.js", Package: "test"})
	g.AddNode(&Node{ID: "file:src/auth.js", Type: NodeSourceFile, Path: "src/auth.js", Package: "src"})
	g.AddNode(&Node{ID: "test:a:10:login", Type: NodeTest, Path: "a.test.js", Name: "login", Line: 10, Package: "test"})

	// Add new unified-model nodes alongside.
	g.AddNode(&Node{ID: "env:prod", Type: NodeEnvironment, Name: "production"})
	g.AddNode(&Node{ID: "behavior:auth", Type: NodeBehaviorSurface, Name: "auth-flow"})
	g.AddNode(&Node{ID: "owner:backend", Type: NodeOwner, Name: "backend-team"})

	// Traditional edges.
	g.AddEdge(&Edge{From: "test:a:10:login", To: "file:a.test.js", Type: EdgeTestDefinedInFile, Confidence: 1.0, EvidenceType: EvidenceStaticAnalysis})
	g.AddEdge(&Edge{From: "file:a.test.js", To: "file:src/auth.js", Type: EdgeImportsModule, Confidence: 1.0, EvidenceType: EvidenceStaticAnalysis})

	// New cross-family edges.
	g.AddEdge(&Edge{From: "test:a:10:login", To: "env:prod", Type: EdgeTargetsEnvironment, Confidence: 0.8, EvidenceType: EvidenceConvention})
	g.AddEdge(&Edge{From: "owner:backend", To: "file:src/auth.js", Type: EdgeOwns, Confidence: 1.0, EvidenceType: EvidenceManual})

	// Existing engines should still work correctly, ignoring the new types.
	coverage := AnalyzeCoverage(g)
	if coverage.SourceCount != 1 {
		t.Errorf("coverage: expected 1 source, got %d", coverage.SourceCount)
	}

	fanout := AnalyzeFanout(g, 10)
	if fanout.NodeCount != 6 {
		t.Errorf("fanout: expected 6 nodes, got %d", fanout.NodeCount)
	}

	impact := AnalyzeImpact(g, []string{"src/auth.js"})
	if len(impact.Tests) != 1 {
		t.Errorf("impact: expected 1 impacted test, got %d", len(impact.Tests))
	}

	dupes := DetectDuplicates(g)
	if dupes.TestsAnalyzed != 1 {
		t.Errorf("duplicates: expected 1 test analyzed, got %d", dupes.TestsAnalyzed)
	}
}

func TestEvidenceTypeExecution(t *testing.T) {
	t.Parallel()
	g := NewGraph()

	g.AddNode(&Node{ID: "run:ci-123", Type: NodeExecutionRun, Name: "ci-123"})
	g.AddNode(&Node{ID: "test:a:1:login", Type: NodeTest, Name: "login"})
	g.AddEdge(&Edge{
		From:         "run:ci-123",
		To:           "test:a:1:login",
		Type:         EdgeExecutionRunsTest,
		Confidence:   1.0,
		EvidenceType: EvidenceExecution,
	})

	edges := g.EdgesByType(EdgeExecutionRunsTest)
	if len(edges) != 1 {
		t.Fatalf("expected 1 execution edge, got %d", len(edges))
	}
	if edges[0].EvidenceType != EvidenceExecution {
		t.Errorf("expected execution evidence type, got %s", edges[0].EvidenceType)
	}
}

// buildUnifiedTestGraph creates a graph spanning all six node families
// with cross-family edges for traversal testing.
func buildUnifiedTestGraph() *Graph {
	g := NewGraph()

	// System nodes.
	g.AddNode(&Node{ID: "file:src/auth.go", Type: NodeSourceFile, Path: "src/auth.go", Package: "auth"})
	g.AddNode(&Node{ID: "svc:auth-api", Type: NodeService, Name: "auth-api"})

	// Validation nodes.
	g.AddNode(&Node{ID: "test:a:1:login", Type: NodeTest, Path: "a.test.go", Name: "login", Line: 1})
	g.AddNode(&Node{ID: "test:a:10:logout", Type: NodeTest, Path: "a.test.go", Name: "logout", Line: 10})
	g.AddNode(&Node{ID: "file:a.test.go", Type: NodeTestFile, Path: "a.test.go"})

	// Behavior nodes.
	g.AddNode(&Node{ID: "behavior:auth-flow", Type: NodeBehaviorSurface, Name: "auth-flow"})

	// Environment nodes.
	g.AddNode(&Node{ID: "env:staging", Type: NodeEnvironment, Name: "staging"})

	// Execution nodes.
	g.AddNode(&Node{ID: "run:ci-456", Type: NodeExecutionRun, Name: "ci-456"})

	// Governance nodes.
	g.AddNode(&Node{ID: "owner:team-alpha", Type: NodeOwner, Name: "team-alpha"})

	// Validation → System edges.
	g.AddEdge(&Edge{From: "test:a:1:login", To: "file:src/auth.go", Type: EdgeValidates, Confidence: 0.95, EvidenceType: EvidenceStaticAnalysis})
	g.AddEdge(&Edge{From: "test:a:10:logout", To: "file:src/auth.go", Type: EdgeValidates, Confidence: 0.9, EvidenceType: EvidenceStaticAnalysis})

	// Validation → Environment edges.
	g.AddEdge(&Edge{From: "test:a:1:login", To: "env:staging", Type: EdgeTargetsEnvironment, Confidence: 0.8, EvidenceType: EvidenceConvention})

	// Behavior → System edges.
	g.AddEdge(&Edge{From: "behavior:auth-flow", To: "file:src/auth.go", Type: EdgeBehaviorDerivedFrom, Confidence: 0.7, EvidenceType: EvidenceInferred})

	// Governance → System edges.
	g.AddEdge(&Edge{From: "owner:team-alpha", To: "file:src/auth.go", Type: EdgeOwns, Confidence: 1.0, EvidenceType: EvidenceManual})

	// Execution → Validation edges.
	g.AddEdge(&Edge{From: "run:ci-456", To: "test:a:1:login", Type: EdgeExecutionRunsTest, Confidence: 1.0, EvidenceType: EvidenceExecution})

	return g
}
