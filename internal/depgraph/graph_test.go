package depgraph

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
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
