package coverage

import (
	"sort"
	"strconv"

	"github.com/pmclSF/terrain/internal/models"
)

// UnitCoverage represents coverage attribution for a single code unit.
type UnitCoverage struct {
	// UnitID is the stable code unit identifier.
	UnitID string `json:"unitId"`

	// Name is the code unit name.
	Name string `json:"name"`

	// Path is the source file path.
	Path string `json:"path"`

	// CoveredAny indicates whether the unit has any observed coverage.
	CoveredAny bool `json:"coveredAny"`

	// LineCoveragePct is the percentage of the unit's lines that are covered.
	// -1 if line-level data is not available for this unit.
	LineCoveragePct float64 `json:"lineCoveragePct"`

	// BranchCoveragePct is the percentage of branches covered.
	// -1 if branch data is not available.
	BranchCoveragePct float64 `json:"branchCoveragePct"`

	// FunctionHit indicates whether the function was directly hit.
	// -1 = unknown, 0 = not hit, 1 = hit.
	FunctionHit int `json:"functionHit"`

	// EvidenceQuality describes the quality of attribution evidence.
	EvidenceQuality string `json:"evidenceQuality"` // "exact", "approximate", "unavailable"

	// CoveredByTypes lists test types that cover this unit (from labeled runs).
	CoveredByTypes []string `json:"coveredByTypes,omitempty"`
}

// AttributeToCodeUnits maps coverage records onto code units.
func AttributeToCodeUnits(merged *MergedCoverage, units []models.CodeUnit) []UnitCoverage {
	var result []UnitCoverage
	implicitEndByUnit := buildImplicitEndLines(units)

	for _, cu := range units {
		rec, ok := merged.ByFile[cu.Path]
		if !ok {
			result = append(result, UnitCoverage{
				UnitID:            cu.UnitID,
				Name:              cu.Name,
				Path:              cu.Path,
				CoveredAny:        false,
				LineCoveragePct:   -1,
				BranchCoveragePct: -1,
				FunctionHit:       -1,
				EvidenceQuality:   "unavailable",
			})
			continue
		}

		uc := UnitCoverage{
			UnitID:          cu.UnitID,
			Name:            cu.Name,
			Path:            cu.Path,
			EvidenceQuality: "approximate",
		}

		// Function hit detection (exact if function-level data available).
		if len(rec.FunctionHits) > 0 {
			if hits, ok := rec.FunctionHits[cu.Name]; ok {
				uc.EvidenceQuality = "exact"
				if hits > 0 {
					uc.FunctionHit = 1
					uc.CoveredAny = true
				} else {
					uc.FunctionHit = 0
				}
			} else {
				uc.FunctionHit = -1
			}
		} else {
			uc.FunctionHit = -1
		}

		// Line coverage for the unit's span.
		if cu.StartLine > 0 && len(rec.LineHits) > 0 {
			endLine := cu.EndLine
			if endLine == 0 {
				if inferredEnd, ok := implicitEndByUnit[coverageUnitKey(cu)]; ok && inferredEnd >= cu.StartLine {
					endLine = inferredEnd
				} else {
					endLine = maxInstrumentedLine(rec.LineHits)
				}
			}
			if endLine < cu.StartLine {
				endLine = cu.StartLine
			}
			covered, total := countLineCoverage(rec.LineHits, cu.StartLine, endLine)
			if total > 0 {
				uc.LineCoveragePct = float64(covered) / float64(total) * 100.0
				if covered > 0 {
					uc.CoveredAny = true
				}
			} else {
				uc.LineCoveragePct = -1
			}
		} else {
			uc.LineCoveragePct = -1
		}

		// Branch coverage (file-level only for now).
		if rec.BranchTotalCount > 0 {
			uc.BranchCoveragePct = float64(rec.BranchCoveredCount) / float64(rec.BranchTotalCount) * 100.0
		} else {
			uc.BranchCoveragePct = -1
		}

		result = append(result, uc)
	}

	return result
}

func countLineCoverage(lineHits map[int]int, start, end int) (covered, total int) {
	for line, hits := range lineHits {
		if line >= start && line <= end {
			total++
			if hits > 0 {
				covered++
			}
		}
	}
	return
}

func buildImplicitEndLines(units []models.CodeUnit) map[string]int {
	byPath := map[string][]models.CodeUnit{}
	for _, cu := range units {
		if cu.StartLine <= 0 || cu.EndLine > 0 {
			continue
		}
		byPath[cu.Path] = append(byPath[cu.Path], cu)
	}

	endByUnit := map[string]int{}
	for _, fileUnits := range byPath {
		sort.Slice(fileUnits, func(i, j int) bool {
			if fileUnits[i].StartLine != fileUnits[j].StartLine {
				return fileUnits[i].StartLine < fileUnits[j].StartLine
			}
			return fileUnits[i].Name < fileUnits[j].Name
		})
		for i := range fileUnits {
			cur := fileUnits[i]
			if i+1 < len(fileUnits) {
				next := fileUnits[i+1]
				if next.StartLine > cur.StartLine {
					endByUnit[coverageUnitKey(cur)] = next.StartLine - 1
				}
			}
		}
	}

	return endByUnit
}

func coverageUnitKey(cu models.CodeUnit) string {
	if cu.UnitID != "" {
		return cu.UnitID
	}
	return cu.Path + ":" + cu.Name + ":" + strconv.Itoa(cu.StartLine)
}

func maxInstrumentedLine(lineHits map[int]int) int {
	max := 0
	for line := range lineHits {
		if line > max {
			max = line
		}
	}
	return max
}
