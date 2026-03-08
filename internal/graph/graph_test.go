package graph

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func testSnapshot() *models.TestSuiteSnapshot {
	return &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/auth/auth.test.js", Framework: "jest", Owner: "@team-auth", TestCount: 3},
			{Path: "src/api/api.test.js", Framework: "jest", Owner: "@team-api", TestCount: 2},
			{Path: "tests/e2e/checkout.spec.js", Framework: "playwright", Owner: "@team-api", TestCount: 1},
		},
		TestCases: []models.TestCase{
			{TestID: "t1", FilePath: "src/auth/auth.test.js", TestName: "validates token", TestType: "unit", Framework: "jest"},
			{TestID: "t2", FilePath: "src/auth/auth.test.js", TestName: "refreshes token", TestType: "unit", Framework: "jest"},
			{TestID: "t3", FilePath: "src/auth/auth.test.js", TestName: "handles expired", TestType: "integration", Framework: "jest"},
			{TestID: "t4", FilePath: "src/api/api.test.js", TestName: "creates user", TestType: "unit", Framework: "jest"},
			{TestID: "t5", FilePath: "src/api/api.test.js", TestName: "deletes user", TestType: "unit", Framework: "jest"},
			{TestID: "t6", FilePath: "tests/e2e/checkout.spec.js", TestName: "full checkout", TestType: "e2e", Framework: "playwright"},
		},
		CodeUnits: []models.CodeUnit{
			{UnitID: "u1", Name: "validateToken", Path: "src/auth/auth.js", Kind: "function", Exported: true, Owner: "@team-auth", Coverage: 0.8},
			{UnitID: "u2", Name: "refreshToken", Path: "src/auth/auth.js", Kind: "function", Exported: true, Owner: "@team-auth", Coverage: 0},
			{UnitID: "u3", Name: "createUser", Path: "src/api/users.js", Kind: "function", Exported: true, Owner: "@team-api", Coverage: 0.9},
			{UnitID: "u4", Name: "processPayment", Path: "src/api/checkout.js", Kind: "function", Exported: true, Owner: "@team-api", Coverage: 0.5},
			{UnitID: "u5", Name: "helperInternal", Path: "src/api/checkout.js", Kind: "function", Exported: false, Owner: "@team-api"},
		},
		Signals: []models.Signal{
			{Type: "weakAssertion", Category: models.CategoryQuality, Severity: models.SeverityMedium,
				Location: models.SignalLocation{File: "src/auth/auth.test.js"}, Owner: "@team-auth"},
			{Type: "slowTest", Category: models.CategoryHealth, Severity: models.SeverityMedium,
				Location: models.SignalLocation{File: "src/api/api.test.js", Symbol: "creates user"}, Owner: "@team-api",
				Metadata: map[string]any{"testId": "t4"}},
			{Type: "flakyTest", Category: models.CategoryHealth, Severity: models.SeverityMedium,
				Location: models.SignalLocation{File: "tests/e2e/checkout.spec.js", Symbol: "full checkout"}, Owner: "@team-api",
				Metadata: map[string]any{"testId": "t6"}},
			{Type: "untestedExport", Category: models.CategoryQuality, Severity: models.SeverityHigh,
				Location: models.SignalLocation{File: "src/auth/auth.js", Symbol: "refreshToken"}, Owner: "@team-auth"},
		},
		CoverageInsights: []models.CoverageInsight{
			{Type: "e2e_only_coverage", UnitID: "u4", Severity: "medium"},
		},
	}
}

func TestBuild(t *testing.T) {
	g := Build(testSnapshot())

	if len(g.TestByID) != 6 {
		t.Errorf("TestByID: got %d, want 6", len(g.TestByID))
	}
	if g.TestByID["t1"].TestName != "validates token" {
		t.Errorf("TestByID[t1]: got %q", g.TestByID["t1"].TestName)
	}
}

func TestTestsByFile(t *testing.T) {
	g := Build(testSnapshot())

	authTests := g.TestsByFile["src/auth/auth.test.js"]
	if len(authTests) != 3 {
		t.Errorf("TestsByFile[auth]: got %d, want 3", len(authTests))
	}
}

func TestTestsByType(t *testing.T) {
	g := Build(testSnapshot())

	unitTests := g.TestsByType["unit"]
	if len(unitTests) != 4 {
		t.Errorf("TestsByType[unit]: got %d, want 4", len(unitTests))
	}

	e2eTests := g.TestsByType["e2e"]
	if len(e2eTests) != 1 {
		t.Errorf("TestsByType[e2e]: got %d, want 1", len(e2eTests))
	}
}

func TestTestsByOwner(t *testing.T) {
	g := Build(testSnapshot())

	authTests := g.TestsByOwner["@team-auth"]
	if len(authTests) != 3 {
		t.Errorf("TestsByOwner[@team-auth]: got %d, want 3", len(authTests))
	}

	apiTests := g.TestsByOwner["@team-api"]
	if len(apiTests) != 3 {
		t.Errorf("TestsByOwner[@team-api]: got %d, want 3", len(apiTests))
	}
}

func TestCodeUnitIndexes(t *testing.T) {
	g := Build(testSnapshot())

	if len(g.UnitByID) != 5 {
		t.Errorf("UnitByID: got %d, want 5", len(g.UnitByID))
	}
	if len(g.ExportedUnits) != 4 {
		t.Errorf("ExportedUnits: got %d, want 4", len(g.ExportedUnits))
	}
	if len(g.UncoveredExportedUnits) != 1 {
		t.Errorf("UncoveredExportedUnits: got %d, want 1 (u2=refreshToken)", len(g.UncoveredExportedUnits))
	}
}

func TestHealthSignalsByTestID(t *testing.T) {
	g := Build(testSnapshot())

	if len(g.HealthSignalsByTestID) != 2 {
		t.Errorf("HealthSignalsByTestID: got %d, want 2", len(g.HealthSignalsByTestID))
	}

	t4Sigs := g.HealthSignalsByTestID["t4"]
	if len(t4Sigs) != 1 || t4Sigs[0].Type != "slowTest" {
		t.Errorf("HealthSignalsByTestID[t4]: got %v", t4Sigs)
	}
}

func TestE2EOnlyUnits(t *testing.T) {
	g := Build(testSnapshot())

	if len(g.E2EOnlyUnits) != 1 {
		t.Errorf("E2EOnlyUnits: got %d, want 1", len(g.E2EOnlyUnits))
	}
	if g.E2EOnlyUnits[0] != "u4" {
		t.Errorf("E2EOnlyUnits[0]: got %q, want u4", g.E2EOnlyUnits[0])
	}
}

func TestTopFailingTestIDs(t *testing.T) {
	g := Build(testSnapshot())

	top := g.TopFailingTestIDs(5)
	if len(top) != 2 {
		t.Errorf("TopFailingTestIDs: got %d, want 2", len(top))
	}
}

func TestOwnerRiskSummaries(t *testing.T) {
	g := Build(testSnapshot())

	summaries := g.OwnerRiskSummaries()
	if len(summaries) == 0 {
		t.Fatal("OwnerRiskSummaries: got 0")
	}

	// Find @team-api and @team-auth.
	var apiSummary, authSummary *OwnerRiskSummary
	for i := range summaries {
		switch summaries[i].Owner {
		case "@team-api":
			apiSummary = &summaries[i]
		case "@team-auth":
			authSummary = &summaries[i]
		}
	}

	if apiSummary == nil {
		t.Fatal("missing @team-api summary")
	}
	if apiSummary.HealthSignals != 2 {
		t.Errorf("@team-api HealthSignals: got %d, want 2", apiSummary.HealthSignals)
	}
	if apiSummary.E2EOnlyUnits != 1 {
		t.Errorf("@team-api E2EOnlyUnits: got %d, want 1", apiSummary.E2EOnlyUnits)
	}

	if authSummary == nil {
		t.Fatal("missing @team-auth summary")
	}
	if authSummary.UncoveredExported != 1 {
		t.Errorf("@team-auth UncoveredExported: got %d, want 1", authSummary.UncoveredExported)
	}
}

func TestTestsInModule(t *testing.T) {
	g := Build(testSnapshot())

	authTests := g.TestsInModule("src/auth")
	if len(authTests) != 3 {
		t.Errorf("TestsInModule(src/auth): got %d, want 3", len(authTests))
	}
}

func TestModuleCoverageSummaries(t *testing.T) {
	g := Build(testSnapshot())

	summaries := g.ModuleCoverageSummaries()
	if len(summaries) == 0 {
		t.Fatal("ModuleCoverageSummaries: got 0")
	}
}

func TestUncoveredExportedForOwner(t *testing.T) {
	g := Build(testSnapshot())

	authUncovered := g.UncoveredExportedForOwner("@team-auth")
	if len(authUncovered) != 1 {
		t.Errorf("UncoveredExportedForOwner(@team-auth): got %d, want 1", len(authUncovered))
	}

	apiUncovered := g.UncoveredExportedForOwner("@team-api")
	if len(apiUncovered) != 0 {
		t.Errorf("UncoveredExportedForOwner(@team-api): got %d, want 0", len(apiUncovered))
	}
}

func TestEmptySnapshot(t *testing.T) {
	g := Build(&models.TestSuiteSnapshot{})

	if len(g.TestByID) != 0 {
		t.Error("expected empty TestByID")
	}
	if len(g.TopFailingTestIDs(5)) != 0 {
		t.Error("expected empty TopFailingTestIDs")
	}
	if len(g.OwnerRiskSummaries()) != 0 {
		t.Error("expected empty OwnerRiskSummaries")
	}
}
