package coverage

import (
	"sort"

	"github.com/pmclSF/hamlet/internal/models"
)

// TypeCoverage describes coverage by test type for a code unit.
type TypeCoverage struct {
	// UnitID is the stable code unit identifier.
	UnitID string `json:"unitId"`

	// Name is the code unit name.
	Name string `json:"name"`

	// Path is the source file path.
	Path string `json:"path"`

	// CoveredByTypes maps test type labels to coverage status.
	CoveredByTypes map[string]bool `json:"coveredByTypes"`

	// ExclusiveType is the single type that covers this unit, if only one does.
	// Empty if covered by multiple types or uncovered.
	ExclusiveType string `json:"exclusiveType,omitempty"`

	// Uncovered indicates no test type covers this unit.
	Uncovered bool `json:"uncovered"`
}

// FileSummary summarizes coverage by type for a single file.
type FileSummary struct {
	Path                    string  `json:"path"`
	TotalUnits              int     `json:"totalUnits"`
	CoveredByUnit           int     `json:"coveredByUnit"`
	CoveredByIntegration    int     `json:"coveredByIntegration"`
	CoveredByE2E            int     `json:"coveredByE2E"`
	CoveredOnlyByE2E        int     `json:"coveredOnlyByE2E"`
	Uncovered               int     `json:"uncovered"`
	UnitCoveragePct         float64 `json:"unitCoveragePct"`
}

// RepoSummary summarizes coverage by type across the entire repository.
type RepoSummary struct {
	TotalCodeUnits          int     `json:"totalCodeUnits"`
	ExportedCodeUnits       int     `json:"exportedCodeUnits"`
	CoveredByUnitTests      int     `json:"coveredByUnitTests"`
	CoveredByIntegration    int     `json:"coveredByIntegration"`
	CoveredByE2E            int     `json:"coveredByE2E"`
	CoveredOnlyByE2E        int     `json:"coveredOnlyByE2E"`
	UncoveredExported       int     `json:"uncoveredExported"`
	UnitCoveragePct         float64 `json:"unitCoveragePct"`
	TopRiskyAreas           []FileSummary `json:"topRiskyAreas,omitempty"`
}

// ComputeByType computes per-unit coverage by test type from labeled coverage runs.
func ComputeByType(artifacts []CoverageArtifact, units []models.CodeUnit) []TypeCoverage {
	// Group artifacts by run label.
	byLabel := map[string][]CoverageArtifact{}
	for _, art := range artifacts {
		label := art.RunLabel
		if label == "" {
			label = "unknown"
		}
		byLabel[label] = append(byLabel[label], art)
	}

	// For each label, merge and attribute.
	labelCoverage := map[string]map[string]bool{} // unitID -> covered
	for label, arts := range byLabel {
		merged := Merge(arts)
		attributed := AttributeToCodeUnits(merged, units)
		for _, uc := range attributed {
			if uc.CoveredAny {
				if labelCoverage[uc.UnitID] == nil {
					labelCoverage[uc.UnitID] = map[string]bool{}
				}
				labelCoverage[uc.UnitID][label] = true
			}
		}
	}

	// Build result.
	var result []TypeCoverage
	for _, cu := range units {
		types := labelCoverage[cu.UnitID]
		tc := TypeCoverage{
			UnitID:         cu.UnitID,
			Name:           cu.Name,
			Path:           cu.Path,
			CoveredByTypes: types,
			Uncovered:      len(types) == 0,
		}
		if len(types) == 1 {
			for t := range types {
				tc.ExclusiveType = t
				break
			}
		}
		result = append(result, tc)
	}

	return result
}

// BuildRepoSummary summarizes coverage by type at the repository level.
func BuildRepoSummary(typeCov []TypeCoverage, units []models.CodeUnit) *RepoSummary {
	exported := map[string]bool{}
	for _, cu := range units {
		if cu.Exported {
			exported[cu.UnitID] = true
		}
	}

	rs := &RepoSummary{
		TotalCodeUnits:    len(typeCov),
		ExportedCodeUnits: len(exported),
	}

	fileCounts := map[string]*FileSummary{}

	for _, tc := range typeCov {
		// File summary tracking.
		fs, ok := fileCounts[tc.Path]
		if !ok {
			fs = &FileSummary{Path: tc.Path}
			fileCounts[tc.Path] = fs
		}
		fs.TotalUnits++

		if tc.CoveredByTypes["unit"] {
			rs.CoveredByUnitTests++
			fs.CoveredByUnit++
		}
		if tc.CoveredByTypes["integration"] {
			rs.CoveredByIntegration++
			fs.CoveredByIntegration++
		}
		if tc.CoveredByTypes["e2e"] {
			rs.CoveredByE2E++
			fs.CoveredByE2E++
		}

		// Only e2e: covered by e2e but not unit or integration.
		if tc.CoveredByTypes["e2e"] && !tc.CoveredByTypes["unit"] && !tc.CoveredByTypes["integration"] {
			rs.CoveredOnlyByE2E++
			fs.CoveredOnlyByE2E++
		}

		if tc.Uncovered && exported[tc.UnitID] {
			rs.UncoveredExported++
			fs.Uncovered++
		}
	}

	if rs.TotalCodeUnits > 0 {
		rs.UnitCoveragePct = float64(rs.CoveredByUnitTests) / float64(rs.TotalCodeUnits) * 100.0
	}

	// Compute per-file pct and find risky areas.
	var fileSummaries []FileSummary
	for _, fs := range fileCounts {
		if fs.TotalUnits > 0 {
			fs.UnitCoveragePct = float64(fs.CoveredByUnit) / float64(fs.TotalUnits) * 100.0
		}
		if fs.CoveredOnlyByE2E > 0 || fs.Uncovered > 0 {
			fileSummaries = append(fileSummaries, *fs)
		}
	}

	sort.Slice(fileSummaries, func(i, j int) bool {
		// Sort by risk: most "only e2e" + uncovered first.
		ri := fileSummaries[i].CoveredOnlyByE2E + fileSummaries[i].Uncovered
		rj := fileSummaries[j].CoveredOnlyByE2E + fileSummaries[j].Uncovered
		if ri != rj {
			return ri > rj
		}
		return fileSummaries[i].Path < fileSummaries[j].Path
	})

	limit := 10
	if len(fileSummaries) < limit {
		limit = len(fileSummaries)
	}
	rs.TopRiskyAreas = fileSummaries[:limit]

	return rs
}
