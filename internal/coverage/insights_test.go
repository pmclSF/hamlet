package coverage

import (
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestDeriveInsights_OnlyE2E(t *testing.T) {
	t.Parallel()
	typeCov := []TypeCoverage{
		{
			UnitID: "src/a.go:Foo",
			Name:   "Foo",
			Path:   "src/a.go",
			CoveredByTypes: map[string]bool{
				"e2e": true,
			},
		},
		{
			UnitID: "src/a.go:Bar",
			Name:   "Bar",
			Path:   "src/a.go",
			CoveredByTypes: map[string]bool{
				"unit": true,
			},
		},
	}

	units := []models.CodeUnit{
		{UnitID: "src/a.go:Foo", Exported: true},
		{UnitID: "src/a.go:Bar", Exported: true},
	}

	insights := DeriveInsights(typeCov, units)

	var foundSummary, foundUnit bool
	for _, ins := range insights {
		if ins.Type == "only_e2e_coverage" {
			foundSummary = true
			if ins.Severity != "medium" {
				t.Errorf("only_e2e_coverage severity = %q, want medium", ins.Severity)
			}
		}
		if ins.Type == "only_e2e_unit" {
			foundUnit = true
			if ins.Severity != "medium" {
				t.Errorf("exported only_e2e_unit severity = %q, want medium", ins.Severity)
			}
		}
	}
	if !foundSummary {
		t.Error("expected only_e2e_coverage insight")
	}
	if !foundUnit {
		t.Error("expected only_e2e_unit insight")
	}
}

func TestDeriveInsights_UncoveredExported(t *testing.T) {
	t.Parallel()
	typeCov := []TypeCoverage{
		{
			UnitID:    "src/b.go:Exported",
			Name:      "Exported",
			Path:      "src/b.go",
			Uncovered: true,
		},
	}

	units := []models.CodeUnit{
		{UnitID: "src/b.go:Exported", Exported: true},
	}

	insights := DeriveInsights(typeCov, units)

	var found bool
	for _, ins := range insights {
		if ins.Type == "uncovered_exported" {
			found = true
			if ins.Severity != "high" {
				t.Errorf("severity = %q, want high", ins.Severity)
			}
		}
	}
	if !found {
		t.Error("expected uncovered_exported insight")
	}
}

func TestDeriveInsights_WeakDiversity(t *testing.T) {
	t.Parallel()
	// Need >= 3 total units in a file with some e2e-only.
	typeCov := []TypeCoverage{
		{UnitID: "f.go:A", Name: "A", Path: "f.go", CoveredByTypes: map[string]bool{"e2e": true}},
		{UnitID: "f.go:B", Name: "B", Path: "f.go", CoveredByTypes: map[string]bool{"e2e": true}},
		{UnitID: "f.go:C", Name: "C", Path: "f.go", CoveredByTypes: map[string]bool{"unit": true}},
	}

	insights := DeriveInsights(typeCov, nil)

	var found bool
	for _, ins := range insights {
		if ins.Type == "weak_coverage_diversity" {
			found = true
			if ins.Path != "f.go" {
				t.Errorf("path = %q, want f.go", ins.Path)
			}
		}
	}
	if !found {
		t.Error("expected weak_coverage_diversity insight")
	}
}

func TestDeriveInsights_NoInsights(t *testing.T) {
	t.Parallel()
	// All units covered by unit tests, nothing uncovered.
	typeCov := []TypeCoverage{
		{UnitID: "x.go:A", Name: "A", Path: "x.go", CoveredByTypes: map[string]bool{"unit": true}},
	}

	insights := DeriveInsights(typeCov, nil)
	if len(insights) != 0 {
		t.Errorf("expected no insights, got %d: %v", len(insights), insights)
	}
}

func TestDeriveUnitInsights_NoBranch(t *testing.T) {
	t.Parallel()
	unitCov := []UnitCoverage{
		{CoveredAny: true, LineCoveragePct: 80, BranchCoveragePct: 0},
		{CoveredAny: true, LineCoveragePct: 60, BranchCoveragePct: 50},
	}

	insights := DeriveUnitInsights(unitCov)

	var found bool
	for _, ins := range insights {
		if ins.Type == "line_but_no_branch" {
			found = true
			if ins.Severity != "info" {
				t.Errorf("severity = %q, want info", ins.Severity)
			}
		}
	}
	if !found {
		t.Error("expected line_but_no_branch insight")
	}
}

func TestDeriveUnitInsights_PartiallyCovered(t *testing.T) {
	t.Parallel()
	unitCov := []UnitCoverage{
		{LineCoveragePct: 30},
		{LineCoveragePct: 45},
		{LineCoveragePct: 80},
	}

	insights := DeriveUnitInsights(unitCov)

	var found bool
	for _, ins := range insights {
		if ins.Type == "partially_covered" {
			found = true
		}
	}
	if !found {
		t.Error("expected partially_covered insight")
	}
}

func TestDeriveUnitInsights_Empty(t *testing.T) {
	t.Parallel()
	insights := DeriveUnitInsights(nil)
	if len(insights) != 0 {
		t.Errorf("expected no insights for empty input, got %d", len(insights))
	}
}
