package depgraph

import (
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestNewGraph(t *testing.T) {
	t.Parallel()
	g := NewGraph()
	if g.NodeCount() != 0 {
		t.Errorf("expected 0 nodes, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 0 {
		t.Errorf("expected 0 edges, got %d", g.EdgeCount())
	}
}

func TestAddNodeAndEdge(t *testing.T) {
	t.Parallel()
	g := NewGraph()

	g.AddNode(&Node{ID: "file:a.js", Type: NodeTestFile, Path: "a.js"})
	g.AddNode(&Node{ID: "file:b.js", Type: NodeSourceFile, Path: "b.js"})
	g.AddEdge(&Edge{
		From:       "file:a.js",
		To:         "file:b.js",
		Type:       EdgeImportsModule,
		Confidence: 1.0,
	})

	if g.NodeCount() != 2 {
		t.Errorf("expected 2 nodes, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 1 {
		t.Errorf("expected 1 edge, got %d", g.EdgeCount())
	}

	n := g.Node("file:a.js")
	if n == nil {
		t.Fatal("node file:a.js not found")
	}
	if n.Type != NodeTestFile {
		t.Errorf("expected test_file type, got %s", n.Type)
	}
}

func TestNeighbors(t *testing.T) {
	t.Parallel()
	g := NewGraph()

	g.AddNode(&Node{ID: "a", Type: NodeTestFile})
	g.AddNode(&Node{ID: "b", Type: NodeSourceFile})
	g.AddNode(&Node{ID: "c", Type: NodeSourceFile})
	g.AddEdge(&Edge{From: "a", To: "b", Type: EdgeImportsModule})
	g.AddEdge(&Edge{From: "a", To: "c", Type: EdgeImportsModule})

	neighbors := g.Neighbors("a")
	if len(neighbors) != 2 {
		t.Fatalf("expected 2 neighbors, got %d", len(neighbors))
	}
	// Should be sorted.
	if neighbors[0] != "b" || neighbors[1] != "c" {
		t.Errorf("expected [b c], got %v", neighbors)
	}

	rev := g.ReverseNeighbors("b")
	if len(rev) != 1 || rev[0] != "a" {
		t.Errorf("expected [a], got %v", rev)
	}
}

func TestNodesByType(t *testing.T) {
	t.Parallel()
	g := NewGraph()

	g.AddNode(&Node{ID: "f1", Type: NodeTestFile})
	g.AddNode(&Node{ID: "f2", Type: NodeTestFile})
	g.AddNode(&Node{ID: "s1", Type: NodeSourceFile})

	testFiles := g.NodesByType(NodeTestFile)
	if len(testFiles) != 2 {
		t.Errorf("expected 2 test files, got %d", len(testFiles))
	}

	sources := g.NodesByType(NodeSourceFile)
	if len(sources) != 1 {
		t.Errorf("expected 1 source file, got %d", len(sources))
	}
}

func TestStats(t *testing.T) {
	t.Parallel()
	g := NewGraph()

	g.AddNode(&Node{ID: "a", Type: NodeTest})
	g.AddNode(&Node{ID: "b", Type: NodeTest})
	g.AddNode(&Node{ID: "c", Type: NodeSourceFile})
	g.AddEdge(&Edge{From: "a", To: "c", Type: EdgeImportsModule})
	g.AddEdge(&Edge{From: "b", To: "c", Type: EdgeImportsModule})

	s := g.Stats()
	if s.NodeCount != 3 {
		t.Errorf("expected 3 nodes, got %d", s.NodeCount)
	}
	if s.EdgeCount != 2 {
		t.Errorf("expected 2 edges, got %d", s.EdgeCount)
	}
	if s.NodesByType["test"] != 2 {
		t.Errorf("expected 2 test nodes, got %d", s.NodesByType["test"])
	}
	if s.Density <= 0 {
		t.Errorf("expected positive density, got %f", s.Density)
	}
}

func TestBuild_EmptySnapshot(t *testing.T) {
	t.Parallel()
	g := Build(nil)
	if g.NodeCount() != 0 {
		t.Errorf("expected 0 nodes for nil snapshot, got %d", g.NodeCount())
	}

	g = Build(&models.TestSuiteSnapshot{})
	if g.NodeCount() != 0 {
		t.Errorf("expected 0 nodes for empty snapshot, got %d", g.NodeCount())
	}
}

func TestBuild_TestStructure(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/auth.test.js", Framework: "jest", TestCount: 3},
		},
		TestCases: []models.TestCase{
			{
				FilePath:       "test/auth.test.js",
				TestName:       "should login",
				SuiteHierarchy: []string{"AuthService"},
				Framework:      "jest",
				Line:           10,
			},
			{
				FilePath:       "test/auth.test.js",
				TestName:       "should logout",
				SuiteHierarchy: []string{"AuthService"},
				Framework:      "jest",
				Line:           20,
			},
			{
				FilePath:       "test/auth.test.js",
				TestName:       "should hash password",
				SuiteHierarchy: []string{"AuthService", "password"},
				Framework:      "jest",
				Line:           30,
			},
		},
	}

	g := Build(snap)

	// 1 test file + 3 tests + 2 suites (AuthService, AuthService::password).
	testFiles := g.NodesByType(NodeTestFile)
	if len(testFiles) != 1 {
		t.Errorf("expected 1 test file, got %d", len(testFiles))
	}

	tests := g.NodesByType(NodeTest)
	if len(tests) != 3 {
		t.Errorf("expected 3 tests, got %d", len(tests))
	}

	suites := g.NodesByType(NodeSuite)
	if len(suites) != 2 {
		t.Errorf("expected 2 suites, got %d", len(suites))
	}

	// Each test should have a TEST_DEFINED_IN_FILE edge.
	definedEdges := g.EdgesByType(EdgeTestDefinedInFile)
	if len(definedEdges) != 3 {
		t.Errorf("expected 3 defined-in-file edges, got %d", len(definedEdges))
	}
}

func TestBuild_ImportGraph(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/auth.test.js", Framework: "jest"},
		},
		ImportGraph: map[string]map[string]bool{
			"test/auth.test.js": {
				"src/auth/service.js": true,
				"src/auth/util.js":    true,
			},
		},
	}

	g := Build(snap)

	// Should have test file + 2 source files.
	sources := g.NodesByType(NodeSourceFile)
	if len(sources) != 2 {
		t.Errorf("expected 2 source files, got %d", len(sources))
	}

	// Should have 2 import edges.
	imports := g.EdgesByType(EdgeImportsModule)
	if len(imports) != 2 {
		t.Errorf("expected 2 import edges, got %d", len(imports))
	}

	// Source files in the same package should have a source→source edge.
	s2s := g.EdgesByType(EdgeSourceImportsSource)
	if len(s2s) != 1 {
		t.Errorf("expected 1 source-to-source edge, got %d", len(s2s))
	}
}

func TestBuild_CodeSurfaces(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		CodeSurfaces: []models.CodeSurface{
			{
				SurfaceID:      "surface:src/routes/api.ts:GET /api/users",
				Name:           "GET /api/users",
				Path:           "src/routes/api.ts",
				Kind:           models.SurfaceRoute,
				Language:       "typescript",
				HTTPMethod:     "GET",
				Route:          "/api/users",
				Exported:       true,
				LinkedCodeUnit: "unit:src/routes/api.ts:setupRoutes",
			},
			{
				SurfaceID: "surface:src/handlers/auth.ts:loginHandler",
				Name:      "loginHandler",
				Path:      "src/handlers/auth.ts",
				Kind:      models.SurfaceHandler,
				Language:  "typescript",
				Exported:  true,
			},
		},
	}

	g := Build(snap)

	// Should have 2 code surface nodes.
	surfaces := g.NodesByType(NodeCodeSurface)
	if len(surfaces) != 2 {
		t.Fatalf("expected 2 code surface nodes, got %d", len(surfaces))
	}

	// Should have source file nodes created for each surface's path.
	sources := g.NodesByType(NodeSourceFile)
	if len(sources) != 2 {
		t.Errorf("expected 2 source file nodes, got %d", len(sources))
	}

	// Surface node should have metadata.
	routeNode := g.Node("surface:src/routes/api.ts:GET /api/users")
	if routeNode == nil {
		t.Fatal("expected route surface node to exist")
	}
	if routeNode.Metadata["kind"] != "route" {
		t.Errorf("expected kind=route, got %q", routeNode.Metadata["kind"])
	}
	if routeNode.Metadata["httpMethod"] != "GET" {
		t.Errorf("expected httpMethod=GET, got %q", routeNode.Metadata["httpMethod"])
	}

	// Surface with LinkedCodeUnit should have a behavior_derived_from edge.
	derivedEdges := g.EdgesByType(EdgeBehaviorDerivedFrom)
	if len(derivedEdges) != 1 {
		t.Fatalf("expected 1 behavior_derived_from edge, got %d", len(derivedEdges))
	}
	if derivedEdges[0].From != "surface:src/routes/api.ts:GET /api/users" {
		t.Errorf("unexpected edge from: %q", derivedEdges[0].From)
	}
}

func TestBuild_BehaviorSurfaces(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		BehaviorSurfaces: []models.BehaviorSurface{
			{
				BehaviorID:     "behavior:route:/api/users",
				Label:          "/api/users/*",
				Description:    "API routes under /api/users (inferred from 2 endpoints)",
				Kind:           models.BehaviorGroupRoutePrefix,
				CodeSurfaceIDs: []string{"surface:src/api.ts:GET /api/users", "surface:src/api.ts:POST /api/users"},
				RoutePrefix:    "/api/users",
				Language:       "typescript",
			},
		},
	}

	g := Build(snap)

	// Should have 1 behavior surface node.
	bNodes := g.NodesByType(NodeBehaviorSurface)
	if len(bNodes) != 1 {
		t.Fatalf("expected 1 behavior surface node, got %d", len(bNodes))
	}
	if bNodes[0].Name != "/api/users/*" {
		t.Errorf("expected label '/api/users/*', got %q", bNodes[0].Name)
	}
	if bNodes[0].Metadata["kind"] != "route_prefix" {
		t.Errorf("expected kind=route_prefix, got %q", bNodes[0].Metadata["kind"])
	}

	// Should have 2 behavior_derived_from edges to code surfaces.
	derivedEdges := g.EdgesByType(EdgeBehaviorDerivedFrom)
	if len(derivedEdges) != 2 {
		t.Fatalf("expected 2 behavior_derived_from edges, got %d", len(derivedEdges))
	}
	for _, e := range derivedEdges {
		if e.From != "behavior:route:/api/users" {
			t.Errorf("expected edge from behavior node, got from %q", e.From)
		}
		if e.Confidence != 0.7 {
			t.Errorf("expected confidence 0.7, got %f", e.Confidence)
		}
	}
}

func TestBuild_Scenarios(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Scenarios: []models.Scenario{
			{
				ScenarioID:        "scenario:auth:login-flow",
				Name:              "Login flow",
				Category:          "happy_path",
				Framework:         "deepeval",
				Owner:             "ml-team",
				CoveredSurfaceIDs: []string{"surface:src/auth.ts:loginHandler"},
				Executable:        true,
			},
		},
	}

	g := Build(snap)

	// Should have 1 scenario node.
	scenarios := g.NodesByType(NodeScenario)
	if len(scenarios) != 1 {
		t.Fatalf("expected 1 scenario node, got %d", len(scenarios))
	}
	if scenarios[0].Name != "Login flow" {
		t.Errorf("expected name 'Login flow', got %q", scenarios[0].Name)
	}
	if scenarios[0].Metadata["category"] != "happy_path" {
		t.Errorf("expected category happy_path, got %q", scenarios[0].Metadata["category"])
	}

	// Should have a covers_code_surface edge.
	coverEdges := g.EdgesByType(EdgeCoversCodeSurface)
	if len(coverEdges) != 1 {
		t.Fatalf("expected 1 covers_code_surface edge, got %d", len(coverEdges))
	}
	if coverEdges[0].From != "scenario:auth:login-flow" {
		t.Errorf("expected edge from scenario, got %q", coverEdges[0].From)
	}

	// Should have an owner node and owns edge.
	owners := g.NodesByType(NodeOwner)
	if len(owners) != 1 {
		t.Fatalf("expected 1 owner node, got %d", len(owners))
	}
	if owners[0].Name != "ml-team" {
		t.Errorf("expected owner ml-team, got %q", owners[0].Name)
	}
}

func TestBuild_ManualCoverage(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		ManualCoverage: []models.ManualCoverageArtifact{
			{
				ArtifactID:        "manual:testrail:login-suite",
				Name:              "Login regression suite",
				Area:              "auth/login",
				Source:             "testrail",
				Owner:              "qa-team",
				Criticality:        "high",
				Frequency:          "per-release",
				CoveredSurfaceIDs: []string{"surface:src/auth.ts:loginHandler", "surface:src/auth.ts:POST /api/login"},
			},
		},
	}

	g := Build(snap)

	// Should have 1 manual coverage node.
	manuals := g.NodesByType(NodeManualCoverage)
	if len(manuals) != 1 {
		t.Fatalf("expected 1 manual coverage node, got %d", len(manuals))
	}
	if manuals[0].Name != "Login regression suite" {
		t.Errorf("expected name, got %q", manuals[0].Name)
	}
	if manuals[0].Metadata["source"] != "testrail" {
		t.Errorf("expected source testrail, got %q", manuals[0].Metadata["source"])
	}
	if manuals[0].Metadata["criticality"] != "high" {
		t.Errorf("expected criticality high, got %q", manuals[0].Metadata["criticality"])
	}

	// Should have 2 manual_covers edges.
	manualEdges := g.EdgesByType(EdgeManualCovers)
	if len(manualEdges) != 2 {
		t.Fatalf("expected 2 manual_covers edges, got %d", len(manualEdges))
	}

	// Should have an owner node.
	owners := g.NodesByType(NodeOwner)
	if len(owners) != 1 {
		t.Fatalf("expected 1 owner node, got %d", len(owners))
	}
}

func TestBuild_ManualCoverage_AreaResolution(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		// Create code surfaces and behavior surfaces that area can match.
		CodeSurfaces: []models.CodeSurface{
			{SurfaceID: "surface:billing/payment.go:processPayment", Name: "processPayment", Package: "billing"},
			{SurfaceID: "surface:auth/login.go:handleLogin", Name: "handleLogin", Package: "auth"},
		},
		BehaviorSurfaces: []models.BehaviorSurface{
			{BehaviorID: "behavior:billing:checkout-flow", Label: "checkout flow", Package: "billing"},
		},
		ManualCoverage: []models.ManualCoverageArtifact{
			{
				ArtifactID:  "manual:testrail:billing-regression",
				Name:        "billing regression suite",
				Area:        "billing",
				Source:      "testrail",
				Criticality: "high",
				// No CoveredSurfaceIDs — should resolve via area.
			},
		},
	}

	g := Build(snap)

	// The manual coverage node should have edges to the billing code surface
	// and billing behavior surface (both have Package = "billing"), but NOT
	// to the auth surface.
	manualEdges := g.EdgesByType(EdgeManualCovers)
	if len(manualEdges) < 1 {
		t.Fatalf("expected at least 1 area-resolved manual_covers edge, got %d", len(manualEdges))
	}

	// Verify edges point to billing-related nodes only.
	for _, e := range manualEdges {
		target := g.Node(e.To)
		if target == nil {
			t.Errorf("edge target %q not found in graph", e.To)
			continue
		}
		if target.Package != "billing" {
			t.Errorf("expected edge to billing node, got edge to %q (package=%q)", e.To, target.Package)
		}
	}
}

func TestBuild_ManualCoverage_ExplicitSurfacesSkipAreaResolution(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		CodeSurfaces: []models.CodeSurface{
			{SurfaceID: "surface:billing/payment.go:process", Name: "process", Package: "billing"},
		},
		ManualCoverage: []models.ManualCoverageArtifact{
			{
				ArtifactID:        "manual:testrail:billing",
				Name:              "billing suite",
				Area:              "billing",
				Source:            "testrail",
				CoveredSurfaceIDs: []string{"surface:specific:target"},
			},
		},
	}

	g := Build(snap)

	manualEdges := g.EdgesByType(EdgeManualCovers)
	// Should only have the explicit edge, not area-resolved ones.
	if len(manualEdges) != 1 {
		t.Fatalf("expected 1 explicit edge (no area resolution), got %d", len(manualEdges))
	}
	if manualEdges[0].To != "surface:specific:target" {
		t.Errorf("expected edge to explicit target, got %q", manualEdges[0].To)
	}
}

func TestValidationTargets(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/auth.test.js", Framework: "jest"},
		},
		TestCases: []models.TestCase{
			{TestID: "t1", TestName: "login test", FilePath: "test/auth.test.js", Framework: "jest"},
		},
		Scenarios: []models.Scenario{
			{ScenarioID: "scenario:auth:flow", Name: "auth flow", Executable: true},
		},
		ManualCoverage: []models.ManualCoverageArtifact{
			{ArtifactID: "manual:testrail:auth", Name: "auth suite", Source: "testrail"},
		},
	}

	g := Build(snap)
	targets := g.ValidationTargets()

	// Should have: 1 test + 1 scenario + 1 manual = 3 validation targets.
	if len(targets) != 3 {
		t.Fatalf("expected 3 validation targets, got %d", len(targets))
	}

	// Check types are mixed.
	types := map[NodeType]int{}
	for _, n := range targets {
		types[n.Type]++
	}
	if types[NodeTest] != 1 {
		t.Errorf("expected 1 test, got %d", types[NodeTest])
	}
	if types[NodeScenario] != 1 {
		t.Errorf("expected 1 scenario, got %d", types[NodeScenario])
	}
	if types[NodeManualCoverage] != 1 {
		t.Errorf("expected 1 manual, got %d", types[NodeManualCoverage])
	}
}

func TestValidationsForSurface(t *testing.T) {
	t.Parallel()
	g := NewGraph()

	// Create a code surface.
	g.AddNode(&Node{ID: "surface:src/auth.ts:login", Type: NodeCodeSurface, Name: "login"})

	// Create a test that covers it.
	g.AddNode(&Node{ID: "test:auth:1:login", Type: NodeTest, Name: "login test"})
	g.AddEdge(&Edge{
		From: "test:auth:1:login", To: "surface:src/auth.ts:login",
		Type: EdgeCoversCodeSurface, Confidence: 0.9, EvidenceType: EvidenceStaticAnalysis,
	})

	// Create a scenario that covers it.
	g.AddNode(&Node{ID: "scenario:auth:flow", Type: NodeScenario, Name: "auth flow"})
	g.AddEdge(&Edge{
		From: "scenario:auth:flow", To: "surface:src/auth.ts:login",
		Type: EdgeCoversCodeSurface, Confidence: 0.8, EvidenceType: EvidenceInferred,
	})

	// Create a manual coverage that covers it.
	g.AddNode(&Node{ID: "manual:testrail:auth", Type: NodeManualCoverage, Name: "auth suite"})
	g.AddEdge(&Edge{
		From: "manual:testrail:auth", To: "surface:src/auth.ts:login",
		Type: EdgeManualCovers, Confidence: 0.7, EvidenceType: EvidenceManual,
	})

	// Create a non-validation node with an edge (should be excluded).
	g.AddNode(&Node{ID: "file:src/auth.ts", Type: NodeSourceFile})
	g.AddEdge(&Edge{
		From: "file:src/auth.ts", To: "surface:src/auth.ts:login",
		Type: EdgeBelongsToPackage, Confidence: 1.0, EvidenceType: EvidenceStaticAnalysis,
	})

	validations := g.ValidationsForSurface("surface:src/auth.ts:login")

	if len(validations) != 3 {
		t.Fatalf("expected 3 validations for surface, got %d", len(validations))
	}

	// Should include all three validation types.
	types := map[NodeType]bool{}
	for _, n := range validations {
		types[n.Type] = true
	}
	if !types[NodeTest] {
		t.Error("expected test in validations")
	}
	if !types[NodeScenario] {
		t.Error("expected scenario in validations")
	}
	if !types[NodeManualCoverage] {
		t.Error("expected manual coverage in validations")
	}
}

func TestValidationsForSurface_Empty(t *testing.T) {
	t.Parallel()
	g := NewGraph()
	g.AddNode(&Node{ID: "surface:src/auth.ts:login", Type: NodeCodeSurface})

	validations := g.ValidationsForSurface("surface:src/auth.ts:login")
	if len(validations) != 0 {
		t.Errorf("expected 0 validations for uncovered surface, got %d", len(validations))
	}
}

func TestIsValidationNode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		nodeType NodeType
		want     bool
	}{
		{NodeTest, true},
		{NodeScenario, true},
		{NodeManualCoverage, true},
		{NodeTestFile, false},
		{NodeSuite, false},
		{NodeSourceFile, false},
		{NodeCodeSurface, false},
		{NodeBehaviorSurface, false},
		{NodeOwner, false},
	}
	for _, tt := range tests {
		got := IsValidationNode(tt.nodeType)
		if got != tt.want {
			t.Errorf("IsValidationNode(%q) = %v, want %v", tt.nodeType, got, tt.want)
		}
	}
}

func TestInferPackage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		want string
	}{
		{"src/auth/service.js", "src"},
		{"packages/compiler-core/src/parse.ts", "packages/compiler-core"},
		{"libs/utils/helpers.js", "libs/utils"},
		{"index.js", ""},
		{"test/auth.test.js", "test"},
	}
	for _, tt := range tests {
		got := inferPackage(tt.path)
		if got != tt.want {
			t.Errorf("inferPackage(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}
