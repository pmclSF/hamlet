package coverage

import (
	"fmt"
	"sort"

	"github.com/pmclSF/hamlet/internal/models"
)

// CoverageInsight represents an actionable finding from coverage analysis.
type CoverageInsight struct {
	// Type identifies the kind of insight.
	Type string `json:"type"`

	// Severity is the importance level (critical, high, medium, low, info).
	Severity string `json:"severity"`

	// Description is a human-readable explanation.
	Description string `json:"description"`

	// Path is the affected file path if applicable.
	Path string `json:"path,omitempty"`

	// UnitID is the affected code unit if applicable.
	UnitID string `json:"unitId,omitempty"`

	// SuggestedAction is a concrete next step.
	SuggestedAction string `json:"suggestedAction,omitempty"`
}

// DeriveInsights generates actionable coverage insights from type coverage data.
func DeriveInsights(typeCov []TypeCoverage, units []models.CodeUnit) []CoverageInsight {
	var insights []CoverageInsight

	exported := map[string]bool{}
	for _, cu := range units {
		if cu.Exported {
			exported[cu.UnitID] = true
		}
	}

	// 1. Functions covered only by e2e.
	var onlyE2E []TypeCoverage
	for _, tc := range typeCov {
		if tc.CoveredByTypes["e2e"] && !tc.CoveredByTypes["unit"] && !tc.CoveredByTypes["integration"] {
			onlyE2E = append(onlyE2E, tc)
		}
	}
	if len(onlyE2E) > 0 {
		insights = append(insights, CoverageInsight{
			Type:        "only_e2e_coverage",
			Severity:    "medium",
			Description: fmt.Sprintf("%d code unit(s) are covered only by e2e tests. These lack fast unit-level feedback.", len(onlyE2E)),
			SuggestedAction: "Add unit tests for code units that rely exclusively on e2e coverage.",
		})
		// Surface top 5 specific units.
		limit := 5
		if len(onlyE2E) < limit {
			limit = len(onlyE2E)
		}
		for _, tc := range onlyE2E[:limit] {
			sev := "low"
			if exported[tc.UnitID] {
				sev = "medium"
			}
			insights = append(insights, CoverageInsight{
				Type:        "only_e2e_unit",
				Severity:    sev,
				Description: fmt.Sprintf("%s (%s) is covered only by e2e tests.", tc.Name, tc.Path),
				Path:        tc.Path,
				UnitID:      tc.UnitID,
				SuggestedAction: fmt.Sprintf("Add unit tests for %s.", tc.Name),
			})
		}
	}

	// 2. Exported functions with no coverage at all.
	var uncoveredExported []TypeCoverage
	for _, tc := range typeCov {
		if tc.Uncovered && exported[tc.UnitID] {
			uncoveredExported = append(uncoveredExported, tc)
		}
	}
	if len(uncoveredExported) > 0 {
		insights = append(insights, CoverageInsight{
			Type:        "uncovered_exported",
			Severity:    "high",
			Description: fmt.Sprintf("%d exported/public function(s) have no test coverage.", len(uncoveredExported)),
			SuggestedAction: "Prioritize adding tests for public API surface.",
		})
		limit := 5
		if len(uncoveredExported) < limit {
			limit = len(uncoveredExported)
		}
		for _, tc := range uncoveredExported[:limit] {
			insights = append(insights, CoverageInsight{
				Type:        "uncovered_exported_unit",
				Severity:    "high",
				Description: fmt.Sprintf("Exported function %s (%s) has no test coverage.", tc.Name, tc.Path),
				Path:        tc.Path,
				UnitID:      tc.UnitID,
				SuggestedAction: fmt.Sprintf("Add tests for %s — this is public API surface.", tc.Name),
			})
		}
	}

	// 3. Files with weak coverage diversity.
	fileCov := map[string]struct{ total, unitCovered, e2eOnly int }{}
	for _, tc := range typeCov {
		fc := fileCov[tc.Path]
		fc.total++
		if tc.CoveredByTypes["unit"] {
			fc.unitCovered++
		}
		if tc.CoveredByTypes["e2e"] && !tc.CoveredByTypes["unit"] && !tc.CoveredByTypes["integration"] {
			fc.e2eOnly++
		}
		fileCov[tc.Path] = fc
	}

	type fileRisk struct {
		path      string
		e2eOnly   int
		total     int
	}
	var riskyFiles []fileRisk
	for path, fc := range fileCov {
		if fc.e2eOnly > 0 && fc.total >= 3 {
			riskyFiles = append(riskyFiles, fileRisk{path, fc.e2eOnly, fc.total})
		}
	}
	sort.Slice(riskyFiles, func(i, j int) bool {
		return riskyFiles[i].e2eOnly > riskyFiles[j].e2eOnly
	})
	limit := 3
	if len(riskyFiles) < limit {
		limit = len(riskyFiles)
	}
	for _, rf := range riskyFiles[:limit] {
		insights = append(insights, CoverageInsight{
			Type:     "weak_coverage_diversity",
			Severity: "medium",
			Description: fmt.Sprintf("%s has %d of %d code units covered only by e2e tests.",
				rf.path, rf.e2eOnly, rf.total),
			Path: rf.path,
			SuggestedAction: fmt.Sprintf("Add unit tests to %s to reduce e2e dependency.", rf.path),
		})
	}

	return insights
}

// DeriveUnitInsights generates insights from unit-level coverage attribution.
func DeriveUnitInsights(unitCov []UnitCoverage) []CoverageInsight {
	var insights []CoverageInsight

	// Functions with line coverage but no branch coverage.
	var noBranch int
	for _, uc := range unitCov {
		if uc.CoveredAny && uc.LineCoveragePct > 0 && uc.BranchCoveragePct == 0 {
			noBranch++
		}
	}
	if noBranch > 0 {
		insights = append(insights, CoverageInsight{
			Type:        "line_but_no_branch",
			Severity:    "info",
			Description: fmt.Sprintf("%d function(s) have line coverage but no branch coverage.", noBranch),
			SuggestedAction: "Add tests that exercise conditional branches.",
		})
	}

	// Partially covered functions (low line coverage pct).
	var partial int
	for _, uc := range unitCov {
		if uc.LineCoveragePct > 0 && uc.LineCoveragePct < 50 {
			partial++
		}
	}
	if partial > 0 {
		insights = append(insights, CoverageInsight{
			Type:        "partially_covered",
			Severity:    "low",
			Description: fmt.Sprintf("%d function(s) have less than 50%% line coverage.", partial),
			SuggestedAction: "Improve test coverage for partially covered functions.",
		})
	}

	return insights
}
