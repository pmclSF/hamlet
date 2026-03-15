package analyze

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

var updateGolden = flag.Bool("update-golden", false, "update golden snapshot files")

func goldenPath(t *testing.T, name string) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "testdata", name+".golden")
}

func compareGolden(t *testing.T, name string, data any) {
	t.Helper()
	golden := goldenPath(t, name)

	actual, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	if *updateGolden {
		if err := os.WriteFile(golden, actual, 0o644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		t.Logf("updated golden file: %s", golden)
		return
	}

	expected, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("golden file not found: %s\nRun with -update-golden to create it.", golden)
	}

	actualStr := strings.TrimSpace(string(actual))
	expectedStr := strings.TrimSpace(string(expected))

	if actualStr != expectedStr {
		t.Errorf("golden mismatch for %s\n\nExpected:\n%s\n\nActual:\n%s\n\nRun with -update-golden to update.",
			name, expectedStr, actualStr)
	}
}

// goldenReport extracts a stable subset for golden comparison,
// omitting fields that may vary (like graph density floats, limitations
// that depend on runtime state).
type goldenReport struct {
	TestsDetected      TestSummary         `json:"testsDetected"`
	RepoProfile        ProfileSummary      `json:"repoProfile"`
	CoverageConfidence CoverageSummary     `json:"coverageConfidence"`
	DuplicateClusters  DuplicateSummary    `json:"duplicateClusters"`
	HighFanout         FanoutSummary       `json:"highFanout"`
	SkippedTestBurden  SkipSummary         `json:"skippedTestBurden"`
	WeakAreaCount      int                 `json:"weakAreaCount"`
	CIOptimization     CIOptimizationSummary `json:"ciOptimization"`
	SignalSummary      SignalBreakdown     `json:"signalSummary"`
	HasTopInsight      bool                `json:"hasTopInsight"`
	RiskDimensionCount int                 `json:"riskDimensionCount"`
}

func toGoldenReport(r *Report) goldenReport {
	return goldenReport{
		TestsDetected:      r.TestsDetected,
		RepoProfile:        r.RepoProfile,
		CoverageConfidence: r.CoverageConfidence,
		DuplicateClusters:  r.DuplicateClusters,
		HighFanout:         r.HighFanout,
		SkippedTestBurden:  r.SkippedTestBurden,
		WeakAreaCount:      len(r.WeakCoverageAreas),
		CIOptimization:     r.CIOptimization,
		SignalSummary:      r.SignalSummary,
		HasTopInsight:      r.TopInsight != "",
		RiskDimensionCount: len(r.RiskPosture),
	}
}

// --- Fixture: Small repo with direct coverage ---

func smallRepoSnapshot() *models.TestSuiteSnapshot {
	return &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{
			Name:      "myapp",
			Languages: []string{"javascript"},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 3},
		},
		TestFiles: []models.TestFile{
			{Path: "test/auth.test.js", Framework: "jest", TestCount: 3,
				LinkedCodeUnits: []string{"src/auth.js:login", "src/auth.js:logout"}},
			{Path: "test/db.test.js", Framework: "jest", TestCount: 2,
				LinkedCodeUnits: []string{"src/db.js:connect"}},
			{Path: "test/utils.test.js", Framework: "jest", TestCount: 4},
		},
		TestCases: []models.TestCase{
			{TestID: "t1", TestName: "login works", FilePath: "test/auth.test.js", Framework: "jest"},
			{TestID: "t2", TestName: "logout works", FilePath: "test/auth.test.js", Framework: "jest"},
			{TestID: "t3", TestName: "token valid", FilePath: "test/auth.test.js", Framework: "jest"},
			{TestID: "t4", TestName: "connect", FilePath: "test/db.test.js", Framework: "jest"},
			{TestID: "t5", TestName: "query", FilePath: "test/db.test.js", Framework: "jest"},
		},
		CodeUnits: []models.CodeUnit{
			{UnitID: "src/auth.js:login", Name: "login", Path: "src/auth.js", Kind: models.CodeUnitKindFunction, Exported: true},
			{UnitID: "src/auth.js:logout", Name: "logout", Path: "src/auth.js", Kind: models.CodeUnitKindFunction, Exported: true},
			{UnitID: "src/db.js:connect", Name: "connect", Path: "src/db.js", Kind: models.CodeUnitKindFunction, Exported: true},
			{UnitID: "src/utils.js:format", Name: "format", Path: "src/utils.js", Kind: models.CodeUnitKindFunction, Exported: true},
		},
		CodeSurfaces: []models.CodeSurface{
			{SurfaceID: "surface:src/auth.js:login", Name: "login", Path: "src/auth.js", Kind: models.SurfaceFunction, Exported: true},
		},
		Signals: []models.Signal{
			{Type: "weakAssertion", Category: models.CategoryQuality, Severity: models.SeverityMedium,
				Location: models.SignalLocation{File: "test/utils.test.js"}, Explanation: "Uses toBeTruthy instead of specific matcher"},
			{Type: "untestedExport", Category: models.CategoryQuality, Severity: models.SeverityMedium,
				Location: models.SignalLocation{File: "src/utils.js"}, Explanation: "format is exported but not covered by tests"},
		},
	}
}

func TestGolden_AnalyzeReport_SmallRepo(t *testing.T) {
	t.Parallel()
	snap := smallRepoSnapshot()
	report := Build(&BuildInput{Snapshot: snap})
	compareGolden(t, "analyze-small-repo", toGoldenReport(report))
}

// --- Fixture: Empty repo ---

func TestGolden_AnalyzeReport_EmptyRepo(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{Name: "empty"},
	}
	report := Build(&BuildInput{Snapshot: snap})
	compareGolden(t, "analyze-empty-repo", toGoldenReport(report))
}

// --- Fixture: Repo with signals ---

func signalHeavySnapshot() *models.TestSuiteSnapshot {
	snap := smallRepoSnapshot()
	snap.Repository.Name = "signal-heavy"
	// Add more signals.
	snap.Signals = append(snap.Signals,
		models.Signal{Type: "slowTest", Category: models.CategoryHealth, Severity: models.SeverityHigh,
			Location: models.SignalLocation{File: "test/auth.test.js"}, Explanation: "Test takes >5s"},
		models.Signal{Type: "mockHeavyTest", Category: models.CategoryQuality, Severity: models.SeverityLow,
			Location: models.SignalLocation{File: "test/db.test.js"}, Explanation: "3 mocks in test file"},
	)
	return snap
}

func TestGolden_AnalyzeReport_SignalHeavy(t *testing.T) {
	t.Parallel()
	snap := signalHeavySnapshot()
	report := Build(&BuildInput{Snapshot: snap})
	compareGolden(t, "analyze-signal-heavy", toGoldenReport(report))
}
