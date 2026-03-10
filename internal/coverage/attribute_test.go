package coverage

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestAttributeToCodeUnits_FullyCovered(t *testing.T) {
	t.Parallel()
	merged := &MergedCoverage{
		ByFile: map[string]*CoverageRecord{
			"src/utils.js": {
				FilePath:     "src/utils.js",
				LineHits:     map[int]int{1: 5, 2: 5, 3: 5, 4: 5, 5: 5},
				FunctionHits: map[string]int{"formatDate": 5, "parseDate": 3},
			},
		},
	}

	units := []models.CodeUnit{
		{UnitID: "src/utils.js:formatDate", Name: "formatDate", Path: "src/utils.js", StartLine: 1, EndLine: 3},
		{UnitID: "src/utils.js:parseDate", Name: "parseDate", Path: "src/utils.js", StartLine: 4, EndLine: 5},
	}

	result := AttributeToCodeUnits(merged, units)
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}

	if !result[0].CoveredAny {
		t.Error("formatDate should be covered")
	}
	if result[0].FunctionHit != 1 {
		t.Errorf("formatDate FunctionHit = %d, want 1", result[0].FunctionHit)
	}
	if result[0].LineCoveragePct != 100.0 {
		t.Errorf("formatDate LineCoveragePct = %f, want 100.0", result[0].LineCoveragePct)
	}
	if result[0].EvidenceQuality != "exact" {
		t.Errorf("formatDate evidence = %q, want exact", result[0].EvidenceQuality)
	}
}

func TestAttributeToCodeUnits_Uncovered(t *testing.T) {
	t.Parallel()
	merged := &MergedCoverage{
		ByFile: map[string]*CoverageRecord{
			"src/utils.js": {
				FilePath:     "src/utils.js",
				LineHits:     map[int]int{1: 0, 2: 0},
				FunctionHits: map[string]int{"dead": 0},
			},
		},
	}

	units := []models.CodeUnit{
		{UnitID: "src/utils.js:dead", Name: "dead", Path: "src/utils.js", StartLine: 1, EndLine: 2},
	}

	result := AttributeToCodeUnits(merged, units)
	if result[0].CoveredAny {
		t.Error("dead function should not be covered")
	}
	if result[0].FunctionHit != 0 {
		t.Errorf("FunctionHit = %d, want 0", result[0].FunctionHit)
	}
}

func TestAttributeToCodeUnits_NoData(t *testing.T) {
	t.Parallel()
	merged := &MergedCoverage{
		ByFile: map[string]*CoverageRecord{},
	}

	units := []models.CodeUnit{
		{UnitID: "src/missing.js:fn", Name: "fn", Path: "src/missing.js"},
	}

	result := AttributeToCodeUnits(merged, units)
	if result[0].EvidenceQuality != "unavailable" {
		t.Errorf("evidence = %q, want unavailable", result[0].EvidenceQuality)
	}
}

func TestAttributeToCodeUnits_InferEndLineFromNextUnit(t *testing.T) {
	t.Parallel()
	merged := &MergedCoverage{
		ByFile: map[string]*CoverageRecord{
			"src/service.ts": {
				FilePath: "src/service.ts",
				LineHits: map[int]int{
					1: 1,
					2: 1,
					3: 0,
					4: 1,
					5: 1,
					6: 0,
				},
			},
		},
	}

	units := []models.CodeUnit{
		{UnitID: "src/service.ts:first", Name: "first", Path: "src/service.ts", StartLine: 1, EndLine: 0},
		{UnitID: "src/service.ts:second", Name: "second", Path: "src/service.ts", StartLine: 4, EndLine: 0},
	}

	result := AttributeToCodeUnits(merged, units)
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}

	// First unit should infer end line from next start line (4 -> 3), not max line.
	if result[0].LineCoveragePct < 66.6 || result[0].LineCoveragePct > 66.7 {
		t.Errorf("first LineCoveragePct = %f, want approx 66.67", result[0].LineCoveragePct)
	}

	// Last unit should fall back to max instrumented line (line 6).
	if result[1].LineCoveragePct < 66.6 || result[1].LineCoveragePct > 66.7 {
		t.Errorf("second LineCoveragePct = %f, want approx 66.67", result[1].LineCoveragePct)
	}
}

func TestAttributeToCodeUnits_InferEndLineWithSparseCoverage(t *testing.T) {
	t.Parallel()
	merged := &MergedCoverage{
		ByFile: map[string]*CoverageRecord{
			"pkg/core.go": {
				FilePath: "pkg/core.go",
				LineHits: map[int]int{
					10: 0,
					11: 3,
				},
			},
		},
	}

	units := []models.CodeUnit{
		{UnitID: "pkg/core.go:Compute", Name: "Compute", Path: "pkg/core.go", StartLine: 10, EndLine: 0},
	}

	result := AttributeToCodeUnits(merged, units)
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].LineCoveragePct < 49.9 || result[0].LineCoveragePct > 50.1 {
		t.Errorf("LineCoveragePct = %f, want approx 50.0", result[0].LineCoveragePct)
	}
}

func TestComputeByType(t *testing.T) {
	t.Parallel()
	unitArt := CoverageArtifact{
		RunLabel: "unit",
		Records: []CoverageRecord{
			{FilePath: "src/a.js", FunctionHits: map[string]int{"fn1": 5, "fn2": 0}, LineHits: map[int]int{1: 5}},
		},
		Provenance: ArtifactProvenance{Format: "lcov", RunLabel: "unit"},
	}
	e2eArt := CoverageArtifact{
		RunLabel: "e2e",
		Records: []CoverageRecord{
			{FilePath: "src/a.js", FunctionHits: map[string]int{"fn1": 0, "fn2": 3}, LineHits: map[int]int{10: 3}},
		},
		Provenance: ArtifactProvenance{Format: "lcov", RunLabel: "e2e"},
	}

	units := []models.CodeUnit{
		{UnitID: "src/a.js:fn1", Name: "fn1", Path: "src/a.js", StartLine: 1, EndLine: 5},
		{UnitID: "src/a.js:fn2", Name: "fn2", Path: "src/a.js", StartLine: 10, EndLine: 15},
	}

	result := ComputeByType([]CoverageArtifact{unitArt, e2eArt}, units)
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}

	// fn1: covered by unit only.
	fn1 := findTypeCov(result, "src/a.js:fn1")
	if fn1 == nil {
		t.Fatal("fn1 not found")
	}
	if !fn1.CoveredByTypes["unit"] {
		t.Error("fn1 should be covered by unit")
	}
	if fn1.CoveredByTypes["e2e"] {
		t.Error("fn1 should not be covered by e2e")
	}
	if fn1.ExclusiveType != "unit" {
		t.Errorf("fn1 exclusive = %q, want unit", fn1.ExclusiveType)
	}

	// fn2: covered by e2e only.
	fn2 := findTypeCov(result, "src/a.js:fn2")
	if fn2 == nil {
		t.Fatal("fn2 not found")
	}
	if fn2.CoveredByTypes["unit"] {
		t.Error("fn2 should not be covered by unit")
	}
	if !fn2.CoveredByTypes["e2e"] {
		t.Error("fn2 should be covered by e2e")
	}
	if fn2.ExclusiveType != "e2e" {
		t.Errorf("fn2 exclusive = %q, want e2e", fn2.ExclusiveType)
	}
}

func TestBuildRepoSummary(t *testing.T) {
	t.Parallel()
	typeCov := []TypeCoverage{
		{UnitID: "a:fn1", Name: "fn1", Path: "a.js", CoveredByTypes: map[string]bool{"unit": true}},
		{UnitID: "a:fn2", Name: "fn2", Path: "a.js", CoveredByTypes: map[string]bool{"e2e": true}},
		{UnitID: "b:fn3", Name: "fn3", Path: "b.js", CoveredByTypes: nil, Uncovered: true},
	}
	units := []models.CodeUnit{
		{UnitID: "a:fn1", Exported: true},
		{UnitID: "a:fn2", Exported: true},
		{UnitID: "b:fn3", Exported: true},
	}

	rs := BuildRepoSummary(typeCov, units)
	if rs.CoveredByUnitTests != 1 {
		t.Errorf("CoveredByUnitTests = %d, want 1", rs.CoveredByUnitTests)
	}
	if rs.CoveredOnlyByE2E != 1 {
		t.Errorf("CoveredOnlyByE2E = %d, want 1", rs.CoveredOnlyByE2E)
	}
	if rs.UncoveredExported != 1 {
		t.Errorf("UncoveredExported = %d, want 1", rs.UncoveredExported)
	}
}

func findTypeCov(results []TypeCoverage, unitID string) *TypeCoverage {
	for i, r := range results {
		if r.UnitID == unitID {
			return &results[i]
		}
	}
	return nil
}
