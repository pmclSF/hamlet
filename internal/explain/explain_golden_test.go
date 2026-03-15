package explain

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/impact"
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

// --- Fixture: Direct dependency ---
// A test directly imports the changed code unit via LinkedCodeUnits.

func directDependencyResult() *impact.ImpactResult {
	snap := &models.TestSuiteSnapshot{
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
	scope := &impact.ChangeScope{
		ChangedFiles: []impact.ChangedFile{
			{Path: "src/auth.js", ChangeKind: impact.ChangeModified},
		},
	}
	return impact.Analyze(scope, snap)
}

func TestGolden_ExplainTest_DirectDependency(t *testing.T) {
	t.Parallel()
	result := directDependencyResult()
	te, err := ExplainTest("test/auth.test.js", result)
	if err != nil {
		t.Fatalf("ExplainTest failed: %v", err)
	}
	compareGolden(t, "explain-test-direct", te)
}

// --- Fixture: Multi-path ---
// A test covers multiple changed code units via different edges.

func multiPathResult() *impact.ImpactResult {
	snap := &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit},
		},
		TestFiles: []models.TestFile{
			{Path: "test/api.test.js", Framework: "jest", TestCount: 5,
				LinkedCodeUnits: []string{"src/api/handler.js:handle", "src/api/validator.js:validate"}},
			{Path: "test/unit/handler.test.js", Framework: "jest", TestCount: 2,
				LinkedCodeUnits: []string{"src/api/handler.js:handle"}},
		},
		CodeUnits: []models.CodeUnit{
			{UnitID: "src/api/handler.js:handle", Name: "handle", Path: "src/api/handler.js", Kind: models.CodeUnitKindFunction, Exported: true},
			{UnitID: "src/api/validator.js:validate", Name: "validate", Path: "src/api/validator.js", Kind: models.CodeUnitKindFunction, Exported: true},
		},
		CodeSurfaces: []models.CodeSurface{
			{SurfaceID: "surface:src/api/handler.js:handle", Name: "handle", Path: "src/api/handler.js", Kind: models.SurfaceFunction, Exported: true},
			{SurfaceID: "surface:src/api/validator.js:validate", Name: "validate", Path: "src/api/validator.js", Kind: models.SurfaceFunction, Exported: true},
		},
	}
	scope := &impact.ChangeScope{
		ChangedFiles: []impact.ChangedFile{
			{Path: "src/api/handler.js", ChangeKind: impact.ChangeModified},
			{Path: "src/api/validator.js", ChangeKind: impact.ChangeModified},
		},
	}
	return impact.Analyze(scope, snap)
}

func TestGolden_ExplainTest_MultiPath(t *testing.T) {
	t.Parallel()
	result := multiPathResult()
	te, err := ExplainTest("test/api.test.js", result)
	if err != nil {
		t.Fatalf("ExplainTest failed: %v", err)
	}
	compareGolden(t, "explain-test-multipath", te)
}

// --- Fixture: Fallback ---
// No tests directly cover the changed code; fallback is triggered.

func fallbackResult() *impact.ImpactResult {
	snap := &models.TestSuiteSnapshot{
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit},
		},
		TestFiles: []models.TestFile{
			{Path: "test/utils.test.js", Framework: "jest", TestCount: 1},
		},
		CodeUnits: []models.CodeUnit{
			{UnitID: "src/core/engine.js:process", Name: "process", Path: "src/core/engine.js", Kind: models.CodeUnitKindFunction, Exported: true},
		},
		CodeSurfaces: []models.CodeSurface{
			{SurfaceID: "surface:src/core/engine.js:process", Name: "process", Path: "src/core/engine.js", Kind: models.SurfaceFunction, Exported: true},
		},
	}
	scope := &impact.ChangeScope{
		ChangedFiles: []impact.ChangedFile{
			{Path: "src/core/engine.js", ChangeKind: impact.ChangeModified},
		},
	}
	return impact.Analyze(scope, snap)
}

func TestGolden_ExplainTest_Fallback(t *testing.T) {
	t.Parallel()
	result := fallbackResult()
	// This test is not in the impact result, so ExplainTest should error.
	// Instead, test the selection explanation which covers the fallback case.
	sel, err := ExplainSelection(result)
	if err != nil {
		t.Fatalf("ExplainSelection failed: %v", err)
	}
	compareGolden(t, "explain-selection-fallback", sel)
}

// --- Fixture: Selection with direct dependency ---

func TestGolden_ExplainSelection_Direct(t *testing.T) {
	t.Parallel()
	result := directDependencyResult()
	sel, err := ExplainSelection(result)
	if err != nil {
		t.Fatalf("ExplainSelection failed: %v", err)
	}
	compareGolden(t, "explain-selection-direct", sel)
}
