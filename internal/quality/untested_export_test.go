package quality

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestUntestedExportDetector_NoLinkedTests(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/__tests__/auth.test.js"},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "fetchData", Path: "src/services/api.js", Kind: models.CodeUnitKindFunction, Exported: true},
		},
	}

	d := &UntestedExportDetector{}
	signals := d.Detect(snap)

	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}
	if signals[0].Type != "untestedExport" {
		t.Errorf("type = %q, want untestedExport", signals[0].Type)
	}
	if signals[0].Location.Symbol != "fetchData" {
		t.Errorf("symbol = %q, want fetchData", signals[0].Location.Symbol)
	}
}

func TestUntestedExportDetector_HasNearbyTest(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/__tests__/auth.test.js"},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "authenticate", Path: "src/auth.js", Kind: models.CodeUnitKindFunction, Exported: true},
		},
	}

	d := &UntestedExportDetector{}
	signals := d.Detect(snap)

	// "auth" stem matches auth.test.js stem — should not flag
	if len(signals) != 0 {
		t.Errorf("expected 0 signals for code with nearby test, got %d", len(signals))
	}
}

func TestUntestedExportDetector_UnexportedIgnored(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{},
		CodeUnits: []models.CodeUnit{
			{Name: "internalHelper", Path: "src/util.js", Kind: models.CodeUnitKindFunction, Exported: false},
		},
	}

	d := &UntestedExportDetector{}
	signals := d.Detect(snap)

	if len(signals) != 0 {
		t.Errorf("expected 0 signals for unexported code unit, got %d", len(signals))
	}
}

func TestUntestedExportDetector_NoCodeUnits(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/a.test.js"},
		},
	}

	d := &UntestedExportDetector{}
	signals := d.Detect(snap)

	if len(signals) != 0 {
		t.Errorf("expected 0 signals with no code units, got %d", len(signals))
	}
}
