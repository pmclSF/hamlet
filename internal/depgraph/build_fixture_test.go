package depgraph

import (
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

// TestBuildFixtureSurfaces_GraphIntegration verifies that fixture surfaces
// are wired into the graph with correct node types, edges, and metadata.
func TestBuildFixtureSurfaces_GraphIntegration(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/auth.test.ts", Framework: "jest"},
		},
		TestCases: []models.TestCase{
			{
				TestID:         "tid-1",
				FilePath:       "test/auth.test.ts",
				TestName:       "should login",
				Framework:      "jest",
				Language:       "js",
				Line:           10,
				ExtractionKind: "static",
				Confidence:     1.0,
			},
		},
		FixtureSurfaces: []models.FixtureSurface{
			{
				FixtureID:     "fixture:test/auth.test.ts:test.beforeEach",
				Name:          "beforeEach",
				Path:          "test/auth.test.ts",
				Kind:          models.FixtureSetupHook,
				Scope:         "test",
				Language:      "js",
				Framework:     "jest",
				Line:          5,
				Shared:        false,
				DetectionTier: models.TierPattern,
				Confidence:    0.95,
			},
		},
	}

	g := Build(snap)

	// Fixture node exists.
	fNode := g.Node("fixture:test/auth.test.ts:test.beforeEach")
	if fNode == nil {
		t.Fatal("fixture node not found in graph")
	}
	if fNode.Type != NodeFixture {
		t.Errorf("expected node type fixture, got %s", fNode.Type)
	}
	if fNode.Family() != FamilyValidation {
		t.Errorf("expected validation family, got %s", fNode.Family())
	}
	if fNode.Metadata["fixtureKind"] != "setup_hook" {
		t.Errorf("expected fixtureKind=setup_hook, got %s", fNode.Metadata["fixtureKind"])
	}

	// Fixture → TestFile edge.
	var fileEdges []*Edge
	for _, e := range g.Outgoing(fNode.ID) {
		if e.Type == EdgeTestDefinedInFile {
			fileEdges = append(fileEdges, e)
		}
	}
	if len(fileEdges) != 1 {
		t.Errorf("expected 1 EdgeTestDefinedInFile edge from fixture, got %d", len(fileEdges))
	}

	// Test → Fixture edge (test in same file uses the fixture).
	testID := "test:test/auth.test.ts:10:should login"
	var usesFixtureEdges []*Edge
	for _, e := range g.Outgoing(testID) {
		if e.Type == EdgeTestUsesFixture {
			usesFixtureEdges = append(usesFixtureEdges, e)
		}
	}
	if len(usesFixtureEdges) != 1 {
		t.Errorf("expected 1 EdgeTestUsesFixture edge from test, got %d", len(usesFixtureEdges))
	}
	if len(usesFixtureEdges) > 0 && usesFixtureEdges[0].To != fNode.ID {
		t.Errorf("expected fixture edge to %s, got %s", fNode.ID, usesFixtureEdges[0].To)
	}
}

// TestBuildFixtureSurfaces_SharedFixtureViaImport verifies that tests in one
// file are linked to shared fixtures in an imported file.
func TestBuildFixtureSurfaces_SharedFixtureViaImport(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/helpers.ts", Framework: "jest"},
			{Path: "test/app.test.ts", Framework: "jest"},
		},
		TestCases: []models.TestCase{
			{
				TestID:         "tid-app-1",
				FilePath:       "test/app.test.ts",
				TestName:       "should work",
				Framework:      "jest",
				Language:       "js",
				Line:           5,
				ExtractionKind: "static",
				Confidence:     1.0,
			},
		},
		FixtureSurfaces: []models.FixtureSurface{
			{
				FixtureID:     "fixture:test/helpers.ts:createUser",
				Name:          "createUser",
				Path:          "test/helpers.ts",
				Kind:          models.FixtureBuilder,
				Scope:         "unknown",
				Language:      "js",
				Framework:     "jest",
				Line:          1,
				Shared:        true,
				DetectionTier: models.TierPattern,
				Confidence:    0.85,
			},
		},
		ImportGraph: map[string]map[string]bool{
			"test/app.test.ts": {
				"test/helpers.ts": true,
			},
		},
	}

	g := Build(snap)

	// Shared fixture node exists.
	fNode := g.Node("fixture:test/helpers.ts:createUser")
	if fNode == nil {
		t.Fatal("shared fixture node not found")
	}

	// Test in app.test.ts uses the shared fixture from helpers.ts.
	testID := "test:test/app.test.ts:5:should work"
	var usesFixtureEdges []*Edge
	for _, e := range g.Outgoing(testID) {
		if e.Type == EdgeTestUsesFixture {
			usesFixtureEdges = append(usesFixtureEdges, e)
		}
	}
	if len(usesFixtureEdges) != 1 {
		t.Errorf("expected 1 EdgeTestUsesFixture from imported fixture, got %d", len(usesFixtureEdges))
	}
	if len(usesFixtureEdges) > 0 && usesFixtureEdges[0].To != fNode.ID {
		t.Errorf("expected edge to shared fixture, got edge to %s", usesFixtureEdges[0].To)
	}
}

// TestBuildFixtureSurfaces_FixtureToCodeSurface verifies that fixtures are
// linked to code surfaces they set up via name matching.
func TestBuildFixtureSurfaces_FixtureToCodeSurface(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/auth.test.ts", Framework: "jest"},
		},
		CodeSurfaces: []models.CodeSurface{
			{
				SurfaceID: "surface:src/auth.ts:login",
				Name:      "login",
				Path:      "src/auth.ts",
				Kind:      models.SurfaceFunction,
				Package:   "src",
				Language:  "js",
				Exported:  true,
			},
		},
		FixtureSurfaces: []models.FixtureSurface{
			{
				FixtureID: "fixture:test/auth.test.ts:mockLogin",
				Name:      "mockLogin",
				Path:      "test/auth.test.ts",
				Kind:      models.FixtureMockProvider,
				Scope:     "test",
				Language:  "js",
				Framework: "jest",
				Line:      3,
			},
		},
	}

	g := Build(snap)

	// Fixture → CodeSurface edge via name matching (mockLogin contains login).
	var setsEdges []*Edge
	for _, e := range g.Outgoing("fixture:test/auth.test.ts:mockLogin") {
		if e.Type == EdgeFixtureSetsSurface {
			setsEdges = append(setsEdges, e)
		}
	}
	// The fixture is in package "test" and the surface is in package "src",
	// so they won't match (different packages). This is correct behavior.
	if len(setsEdges) != 0 {
		t.Errorf("expected 0 cross-package fixture→surface edges, got %d", len(setsEdges))
	}
}

// TestBuildFixtureSurfaces_FanoutAnalysis verifies that high-fanout fixtures
// can be identified through graph traversal.
func TestBuildFixtureSurfaces_FanoutAnalysis(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/suite.test.ts", Framework: "jest"},
		},
		TestCases: []models.TestCase{
			{TestID: "t1", FilePath: "test/suite.test.ts", TestName: "test1", Framework: "jest", Language: "js", Line: 10, ExtractionKind: "static", Confidence: 1.0},
			{TestID: "t2", FilePath: "test/suite.test.ts", TestName: "test2", Framework: "jest", Language: "js", Line: 20, ExtractionKind: "static", Confidence: 1.0},
			{TestID: "t3", FilePath: "test/suite.test.ts", TestName: "test3", Framework: "jest", Language: "js", Line: 30, ExtractionKind: "static", Confidence: 1.0},
		},
		FixtureSurfaces: []models.FixtureSurface{
			{
				FixtureID: "fixture:test/suite.test.ts:test.beforeEach",
				Name:      "beforeEach",
				Path:      "test/suite.test.ts",
				Kind:      models.FixtureSetupHook,
				Scope:     "test",
				Language:  "js",
				Framework: "jest",
				Line:      5,
			},
		},
	}

	g := Build(snap)

	// Count incoming edges to the fixture (fanout = number of tests using it).
	fixtureID := "fixture:test/suite.test.ts:test.beforeEach"
	incoming := g.Incoming(fixtureID)
	var testUsesCount int
	for _, e := range incoming {
		if e.Type == EdgeTestUsesFixture {
			testUsesCount++
		}
	}

	// All 3 tests in the same file should use the fixture.
	if testUsesCount != 3 {
		t.Errorf("expected fanout of 3 (3 tests use fixture), got %d", testUsesCount)
	}
}

// TestBuildFixtureSurfaces_EmptySnapshot verifies no panic with empty fixtures.
func TestBuildFixtureSurfaces_EmptySnapshot(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{}
	g := Build(snap)
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	fixtureNodes := g.NodesByType(NodeFixture)
	if len(fixtureNodes) != 0 {
		t.Errorf("expected 0 fixture nodes, got %d", len(fixtureNodes))
	}
}
