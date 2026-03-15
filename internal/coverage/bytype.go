package coverage

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
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
	Path                 string  `json:"path"`
	TotalUnits           int     `json:"totalUnits"`
	CoveredByUnit        int     `json:"coveredByUnit"`
	CoveredByIntegration int     `json:"coveredByIntegration"`
	CoveredByE2E         int     `json:"coveredByE2E"`
	CoveredOnlyByE2E     int     `json:"coveredOnlyByE2E"`
	Uncovered            int     `json:"uncovered"`
	UnitCoveragePct      float64 `json:"unitCoveragePct"`
}

// RepoSummary summarizes coverage by type across the entire repository.
type RepoSummary struct {
	TotalCodeUnits       int           `json:"totalCodeUnits"`
	ExportedCodeUnits    int           `json:"exportedCodeUnits"`
	CoveredByUnitTests   int           `json:"coveredByUnitTests"`
	CoveredByIntegration int           `json:"coveredByIntegration"`
	CoveredByE2E         int           `json:"coveredByE2E"`
	CoveredOnlyByE2E     int           `json:"coveredOnlyByE2E"`
	UncoveredExported    int           `json:"uncoveredExported"`
	UnitCoveragePct      float64       `json:"unitCoveragePct"`
	TopRiskyAreas        []FileSummary `json:"topRiskyAreas,omitempty"`
}

// ComputeByType computes per-unit coverage by test type from labeled coverage runs.
func ComputeByType(artifacts []CoverageArtifact, units []models.CodeUnit) []TypeCoverage {
	// Group artifacts by run label.
	byLabel := map[string][]CoverageArtifact{}
	for _, art := range artifacts {
		label := normalizeCoverageLabel(art.RunLabel, art.Provenance.SourceFile)
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

func normalizeCoverageLabel(label, sourceFile string) string {
	l := strings.ToLower(strings.TrimSpace(label))
	switch l {
	case "unit", "units", "jest", "vitest", "go-test", "go-testing", "pytest", "unittest":
		return "unit"
	case "integration", "integrations", "integ", "int":
		return "integration"
	case "e2e", "end-to-end", "end2end", "cypress", "playwright":
		return "e2e"
	}
	if l != "" {
		return l
	}

	file := strings.ToLower(filepath.Base(sourceFile))
	switch {
	case strings.Contains(file, "e2e"), strings.Contains(file, "playwright"), strings.Contains(file, "cypress"):
		return "e2e"
	case strings.Contains(file, "integration"), strings.Contains(file, "integ"), strings.Contains(file, "int"):
		return "integration"
	case strings.Contains(file, "unit"), strings.Contains(file, "jest"), strings.Contains(file, "vitest"):
		return "unit"
	default:
		// Unlabeled single-run coverage should still be useful for unit-coverage
		// measurements instead of being dropped as "unknown".
		return "unit"
	}
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
