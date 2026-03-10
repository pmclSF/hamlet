package coverage

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestComputeByType_UnlabeledCoverageDefaultsToUnit(t *testing.T) {
	t.Parallel()
	units := []models.CodeUnit{
		{UnitID: "src/a.js:A", Name: "A", Path: "src/a.js", StartLine: 1, EndLine: 1},
	}
	arts := []CoverageArtifact{
		{
			RunLabel: "",
			Records: []CoverageRecord{
				{
					FilePath: "src/a.js",
					LineHits: map[int]int{1: 1},
				},
			},
			Provenance: ArtifactProvenance{SourceFile: "coverage-final.json"},
		},
	}

	typeCov := ComputeByType(arts, units)
	if len(typeCov) != 1 {
		t.Fatalf("expected 1 type coverage record, got %d", len(typeCov))
	}
	if !typeCov[0].CoveredByTypes["unit"] {
		t.Fatalf("expected unlabeled coverage to map to unit, got %#v", typeCov[0].CoveredByTypes)
	}

	repo := BuildRepoSummary(typeCov, units)
	if repo.CoveredByUnitTests != 1 {
		t.Fatalf("CoveredByUnitTests = %d, want 1", repo.CoveredByUnitTests)
	}
}

func TestComputeByType_InfersFromSourceFileName(t *testing.T) {
	t.Parallel()
	units := []models.CodeUnit{
		{UnitID: "src/a.js:A", Name: "A", Path: "src/a.js", StartLine: 1, EndLine: 1},
		{UnitID: "src/b.js:B", Name: "B", Path: "src/b.js", StartLine: 1, EndLine: 1},
	}
	arts := []CoverageArtifact{
		{
			RunLabel: "",
			Records: []CoverageRecord{
				{
					FilePath: "src/a.js",
					LineHits: map[int]int{1: 1},
				},
			},
			Provenance: ArtifactProvenance{SourceFile: "/tmp/coverage.e2e.lcov"},
		},
		{
			RunLabel: "integration",
			Records: []CoverageRecord{
				{
					FilePath: "src/b.js",
					LineHits: map[int]int{1: 1},
				},
			},
			Provenance: ArtifactProvenance{SourceFile: "/tmp/integration.json"},
		},
	}

	typeCov := ComputeByType(arts, units)
	if len(typeCov) != 2 {
		t.Fatalf("expected 2 type coverage records, got %d", len(typeCov))
	}

	var a, b TypeCoverage
	if typeCov[0].Path == "src/a.js" {
		a, b = typeCov[0], typeCov[1]
	} else {
		a, b = typeCov[1], typeCov[0]
	}
	if !a.CoveredByTypes["e2e"] {
		t.Fatalf("expected src/a.js to be covered by e2e, got %#v", a.CoveredByTypes)
	}
	if !b.CoveredByTypes["integration"] {
		t.Fatalf("expected src/b.js to be covered by integration, got %#v", b.CoveredByTypes)
	}
}
