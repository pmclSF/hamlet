package impact

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

// goldenResult extracts a stable subset of ImpactResult for golden comparison,
// omitting fields that contain unstable data (like the full Graph).
type goldenResult struct {
	ChangedAreas       []ChangedArea      `json:"changedAreas,omitempty"`
	AffectedBehaviors  []AffectedBehavior `json:"affectedBehaviors,omitempty"`
	ImpactedUnitCount  int                `json:"impactedUnitCount"`
	ImpactedTestCount  int                `json:"impactedTestCount"`
	ProtectionGapCount int                `json:"protectionGapCount"`
	SelectedTestCount  int                `json:"selectedTestCount"`
	CoverageConfidence string             `json:"coverageConfidence"`
	ReasonCategories   ReasonCategories   `json:"reasonCategories"`
	Fallback           FallbackInfo       `json:"fallback"`
	PostureBand        string             `json:"postureBand"`
	Summary            string             `json:"summary"`
	Limitations        []string           `json:"limitations,omitempty"`
}

func toGoldenResult(r *ImpactResult) goldenResult {
	return goldenResult{
		ChangedAreas:       r.ChangedAreas,
		AffectedBehaviors:  r.AffectedBehaviors,
		ImpactedUnitCount:  len(r.ImpactedUnits),
		ImpactedTestCount:  len(r.ImpactedTests),
		ProtectionGapCount: len(r.ProtectionGaps),
		SelectedTestCount:  len(r.SelectedTests),
		CoverageConfidence: r.CoverageConfidence,
		ReasonCategories:   r.ReasonCategories,
		Fallback:           r.Fallback,
		PostureBand:        r.Posture.Band,
		Summary:            r.Summary,
		Limitations:        r.Limitations,
	}
}

// --- Fixture: Direct Dependency Impact ---
// A source file is modified and a test file directly imports it.

func directDependencySnapshot() *models.TestSuiteSnapshot {
	return &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit},
		},
		TestFiles: []models.TestFile{
			{Path: "test/auth.test.js", Framework: "jest", TestCount: 3,
				LinkedCodeUnits: []string{"src/auth.js:login", "src/auth.js:logout"}},
		},
		CodeUnits: []models.CodeUnit{
			{UnitID: "src/auth.js:login", Name: "login", Path: "src/auth.js", Kind: models.CodeUnitKindFunction, Exported: true},
			{UnitID: "src/auth.js:logout", Name: "logout", Path: "src/auth.js", Kind: models.CodeUnitKindFunction, Exported: true},
		},
		CodeSurfaces: []models.CodeSurface{
			{SurfaceID: "surface:src/auth.js:login", Name: "login", Path: "src/auth.js", Kind: models.SurfaceFunction, Exported: true},
			{SurfaceID: "surface:src/auth.js:logout", Name: "logout", Path: "src/auth.js", Kind: models.SurfaceFunction, Exported: true},
		},
	}
}

func TestGolden_DirectDependencyImpact(t *testing.T) {
	t.Parallel()
	snap := directDependencySnapshot()
	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "src/auth.js", ChangeKind: ChangeModified},
		},
	}

	result := Analyze(scope, snap)
	compareGolden(t, "direct-dependency", toGoldenResult(result))
}

// --- Fixture: Fixture-Mediated Impact ---
// A source file is modified. No test directly imports it, but a test
// uses a fixture that imports the source file.

func fixtureMediatedSnapshot() *models.TestSuiteSnapshot {
	return &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit},
		},
		TestFiles: []models.TestFile{
			{Path: "test/integration/api.test.js", Framework: "jest", TestCount: 2},
		},
		CodeUnits: []models.CodeUnit{
			{UnitID: "src/db/connection.js:connect", Name: "connect", Path: "src/db/connection.js", Kind: models.CodeUnitKindFunction, Exported: true},
		},
		CodeSurfaces: []models.CodeSurface{
			{SurfaceID: "surface:src/db/connection.js:connect", Name: "connect", Path: "src/db/connection.js", Kind: models.SurfaceFunction, Exported: true},
		},
	}
}

func TestGolden_FixtureMediatedImpact(t *testing.T) {
	t.Parallel()
	snap := fixtureMediatedSnapshot()
	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "src/db/connection.js", ChangeKind: ChangeModified},
		},
	}

	result := Analyze(scope, snap)
	compareGolden(t, "fixture-mediated", toGoldenResult(result))
}

// --- Fixture: High-Fanout Fallback ---
// Many test files exist but none directly cover the changed code.
// Triggers fallback to near_minimal or fallback_broad.

func highFanoutFallbackSnapshot() *models.TestSuiteSnapshot {
	snap := &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit},
		},
	}

	// Create many test files in various directories.
	for i := 0; i < 10; i++ {
		snap.TestFiles = append(snap.TestFiles, models.TestFile{
			Path:      filepath.Join("test", "module"+string(rune('a'+i)), "feature.test.js"),
			Framework: "jest",
			TestCount: 5,
		})
	}

	// Add a changed source file with code units but no coverage linkage.
	snap.CodeUnits = []models.CodeUnit{
		{UnitID: "src/core/engine.js:process", Name: "process", Path: "src/core/engine.js", Kind: models.CodeUnitKindFunction, Exported: true},
		{UnitID: "src/core/engine.js:validate", Name: "validate", Path: "src/core/engine.js", Kind: models.CodeUnitKindFunction, Exported: true},
	}
	snap.CodeSurfaces = []models.CodeSurface{
		{SurfaceID: "surface:src/core/engine.js:process", Name: "process", Path: "src/core/engine.js", Kind: models.SurfaceFunction, Exported: true},
		{SurfaceID: "surface:src/core/engine.js:validate", Name: "validate", Path: "src/core/engine.js", Kind: models.SurfaceFunction, Exported: true},
	}

	return snap
}

func TestGolden_HighFanoutFallback(t *testing.T) {
	t.Parallel()
	snap := highFanoutFallbackSnapshot()
	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "src/core/engine.js", ChangeKind: ChangeModified},
		},
	}

	result := Analyze(scope, snap)
	compareGolden(t, "high-fanout-fallback", toGoldenResult(result))
}

// --- Fixture: Low-Visibility Area ---
// Changed files have no code units, no surfaces, and limited data.
// The system should be conservative and report evidence_limited.

func lowVisibilitySnapshot() *models.TestSuiteSnapshot {
	return &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit},
		},
		TestFiles: []models.TestFile{
			{Path: "test/utils.test.js", Framework: "jest", TestCount: 1},
		},
	}
}

func TestGolden_LowVisibilityArea(t *testing.T) {
	t.Parallel()
	snap := lowVisibilitySnapshot()
	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "src/internal/scheduler.go", ChangeKind: ChangeModified},
		},
	}

	result := Analyze(scope, snap)
	compareGolden(t, "low-visibility", toGoldenResult(result))
}
