package depgraph

import (
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

// TestBuildBehaviorSurfaces_GraphIntegration verifies that behavior surfaces
// are wired into the graph with correct node types and edges.
func TestBuildBehaviorSurfaces_GraphIntegration(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{
		CodeSurfaces: []models.CodeSurface{
			{SurfaceID: "surface:src/auth.ts:login", Name: "login", Path: "src/auth.ts", Kind: models.SurfaceFunction, Language: "typescript", Exported: true},
			{SurfaceID: "surface:src/auth.ts:register", Name: "register", Path: "src/auth.ts", Kind: models.SurfaceFunction, Language: "typescript", Exported: true},
		},
		BehaviorSurfaces: []models.BehaviorSurface{
			{
				BehaviorID:     "behavior:module:src/auth.ts",
				Label:          "auth",
				Kind:           models.BehaviorGroupModule,
				CodeSurfaceIDs: []string{"surface:src/auth.ts:login", "surface:src/auth.ts:register"},
				Language:       "typescript",
			},
		},
	}

	g := Build(snap)

	// Behavior node exists with correct type.
	bNode := g.Node("behavior:module:src/auth.ts")
	if bNode == nil {
		t.Fatal("behavior node not found in graph")
	}
	if bNode.Type != NodeBehaviorSurface {
		t.Errorf("expected node type behavior_surface, got %s", bNode.Type)
	}
	if bNode.Family() != FamilyBehavior {
		t.Errorf("expected behavior family, got %s", bNode.Family())
	}

	// Behavior → CodeSurface edges exist.
	outgoing := g.Outgoing("behavior:module:src/auth.ts")
	var derivedEdges []*Edge
	for _, e := range outgoing {
		if e.Type == EdgeBehaviorDerivedFrom {
			derivedEdges = append(derivedEdges, e)
		}
	}
	if len(derivedEdges) != 2 {
		t.Errorf("expected 2 BehaviorDerivedFrom edges, got %d", len(derivedEdges))
	}

	// Edges carry inferred evidence with 0.7 confidence.
	for _, e := range derivedEdges {
		if e.Confidence != 0.7 {
			t.Errorf("expected confidence 0.7, got %f", e.Confidence)
		}
		if e.EvidenceType != EvidenceInferred {
			t.Errorf("expected inferred evidence, got %s", e.EvidenceType)
		}
	}
}

// TestBuildWithoutBehaviorSurfaces verifies that the graph pipeline works
// correctly when no behavior surfaces are present. BehaviorSurface is optional.
func TestBuildWithoutBehaviorSurfaces(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:      "src/__tests__/auth.test.ts",
				Framework: "jest",
				TestCount: 1,
			},
		},
		TestCases: []models.TestCase{
			{TestName: "should login", FilePath: "src/__tests__/auth.test.ts", Framework: "jest", Language: "javascript", Line: 5, ExtractionKind: "static", Confidence: 1.0},
		},
		CodeSurfaces: []models.CodeSurface{
			{SurfaceID: "surface:src/auth.ts:login", Name: "login", Path: "src/auth.ts", Kind: models.SurfaceFunction, Language: "typescript", Exported: true},
		},
		// No BehaviorSurfaces — this is intentionally empty.
	}

	g := Build(snap)

	// Graph should still work.
	if g.NodeCount() == 0 {
		t.Fatal("expected nodes from test files and code surfaces")
	}

	// Code surface node should exist.
	csNode := g.Node("surface:src/auth.ts:login")
	if csNode == nil {
		t.Fatal("code surface node not found")
	}
	if csNode.Type != NodeCodeSurface {
		t.Errorf("expected code_surface type, got %s", csNode.Type)
	}

	// No behavior nodes.
	behaviorNodes := g.NodesByFamily(FamilyBehavior)
	if len(behaviorNodes) != 0 {
		t.Errorf("expected 0 behavior nodes when none provided, got %d", len(behaviorNodes))
	}

	// Coverage and fanout engines should still work.
	coverage := AnalyzeCoverage(g)
	if coverage.SourceCount == 0 {
		t.Error("coverage analysis should work without behavior surfaces")
	}

	fanout := AnalyzeFanout(g, 10)
	if fanout.NodeCount == 0 {
		t.Error("fanout analysis should work without behavior surfaces")
	}
}

// TestBuildCodeSurfaces_GraphIntegration verifies that code surfaces
// create NodeCodeSurface nodes with correct edges to source files.
func TestBuildCodeSurfaces_GraphIntegration(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{
		CodeSurfaces: []models.CodeSurface{
			{
				SurfaceID:  "surface:src/routes/api.ts:GET /api/users",
				Name:       "GET /api/users",
				Path:       "src/routes/api.ts",
				Kind:       models.SurfaceRoute,
				Language:   "typescript",
				HTTPMethod: "GET",
				Route:      "/api/users",
				Exported:   true,
			},
			{
				SurfaceID:      "surface:handlers/auth.go:LoginHandler",
				Name:           "LoginHandler",
				Path:           "handlers/auth.go",
				Kind:           models.SurfaceHandler,
				Language:       "go",
				Exported:       true,
				LinkedCodeUnit: "handlers/auth.go:LoginHandler",
			},
		},
	}

	g := Build(snap)

	// Route surface should exist.
	routeNode := g.Node("surface:src/routes/api.ts:GET /api/users")
	if routeNode == nil {
		t.Fatal("route surface node not found")
	}
	if routeNode.Type != NodeCodeSurface {
		t.Errorf("expected code_surface type, got %s", routeNode.Type)
	}
	if routeNode.Metadata["kind"] != "route" {
		t.Errorf("expected kind=route in metadata, got %q", routeNode.Metadata["kind"])
	}
	if routeNode.Metadata["httpMethod"] != "GET" {
		t.Errorf("expected httpMethod=GET in metadata, got %q", routeNode.Metadata["httpMethod"])
	}

	// Handler surface should exist with linked code unit edge.
	handlerNode := g.Node("surface:handlers/auth.go:LoginHandler")
	if handlerNode == nil {
		t.Fatal("handler surface node not found")
	}

	outgoing := g.Outgoing("surface:handlers/auth.go:LoginHandler")
	var hasLinkedEdge bool
	for _, e := range outgoing {
		if e.Type == EdgeBehaviorDerivedFrom && e.To == "handlers/auth.go:LoginHandler" {
			hasLinkedEdge = true
		}
	}
	if !hasLinkedEdge {
		t.Error("expected BehaviorDerivedFrom edge to linked code unit")
	}

	// Source file nodes should be auto-created.
	srcNode := g.Node("file:src/routes/api.ts")
	if srcNode == nil {
		t.Fatal("source file node should be auto-created for code surface")
	}
	if srcNode.Type != NodeSourceFile {
		t.Errorf("expected source_file type, got %s", srcNode.Type)
	}
}

// TestBuildEndToEnd_InferenceChain verifies the full chain:
// test files + code surfaces + behavior surfaces → graph with correct
// cross-family traversal.
func TestBuildEndToEnd_InferenceChain(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:      "tests/auth.test.ts",
				Framework: "jest",
				TestCount: 2,
			},
		},
		TestCases: []models.TestCase{
			{TestName: "should login successfully", FilePath: "tests/auth.test.ts", Framework: "jest", Language: "javascript", Line: 10, ExtractionKind: "static", Confidence: 1.0},
			{TestName: "should reject bad password", FilePath: "tests/auth.test.ts", Framework: "jest", Language: "javascript", Line: 20, ExtractionKind: "static", Confidence: 1.0},
		},
		ImportGraph: map[string]map[string]bool{
			"tests/auth.test.ts": {"src/auth.ts": true},
		},
		CodeSurfaces: []models.CodeSurface{
			{SurfaceID: "surface:src/auth.ts:login", Name: "login", Path: "src/auth.ts", Kind: models.SurfaceFunction, Language: "typescript", Exported: true},
			{SurfaceID: "surface:src/auth.ts:register", Name: "register", Path: "src/auth.ts", Kind: models.SurfaceFunction, Language: "typescript", Exported: true},
			{SurfaceID: "surface:src/auth.ts:logout", Name: "logout", Path: "src/auth.ts", Kind: models.SurfaceFunction, Language: "typescript", Exported: true},
		},
		BehaviorSurfaces: []models.BehaviorSurface{
			{
				BehaviorID:     "behavior:module:src/auth.ts",
				Label:          "auth",
				Kind:           models.BehaviorGroupModule,
				CodeSurfaceIDs: []string{"surface:src/auth.ts:login", "surface:src/auth.ts:register", "surface:src/auth.ts:logout"},
				Language:       "typescript",
			},
		},
	}

	g := Build(snap)

	// All six families should be queryable.
	systemNodes := g.NodesByFamily(FamilySystem)
	validationNodes := g.NodesByFamily(FamilyValidation)
	behaviorNodes := g.NodesByFamily(FamilyBehavior)

	if len(systemNodes) == 0 {
		t.Error("expected system nodes (source files, code surfaces)")
	}
	if len(validationNodes) == 0 {
		t.Error("expected validation nodes (tests, test files)")
	}
	if len(behaviorNodes) != 1 {
		t.Errorf("expected 1 behavior node, got %d", len(behaviorNodes))
	}

	// Verify traversal: behavior → code surfaces → source file.
	behaviorID := "behavior:module:src/auth.ts"
	bOutgoing := g.Outgoing(behaviorID)
	if len(bOutgoing) == 0 {
		t.Fatal("behavior node should have outgoing edges to code surfaces")
	}

	// Follow one code surface to its source file.
	csID := bOutgoing[0].To
	csOutgoing := g.Outgoing(csID)
	var reachesSourceFile bool
	for _, e := range csOutgoing {
		target := g.Node(e.To)
		if target != nil && target.Type == NodeSourceFile {
			reachesSourceFile = true
			break
		}
	}
	if !reachesSourceFile {
		t.Error("code surface should connect to its source file")
	}

	// Verify impact: changing src/auth.ts should reach test nodes.
	impact := AnalyzeImpact(g, []string{"src/auth.ts"})
	if len(impact.Tests) == 0 {
		t.Error("changing src/auth.ts should impact test nodes")
	}
}

// TestBuildNilSnapshot verifies Build handles nil gracefully.
func TestBuildNilSnapshot(t *testing.T) {
	t.Parallel()
	g := Build(nil)
	if g == nil {
		t.Fatal("Build(nil) should return empty graph, not nil")
	}
	if g.NodeCount() != 0 {
		t.Errorf("expected 0 nodes, got %d", g.NodeCount())
	}
}
