package quality

import (
	"github.com/pmclSF/hamlet/internal/models"
)

// CoverageBlindSpotDetector translates coverage insights into canonical
// coverageBlindSpot signals for scoring and measurement layers.
type CoverageBlindSpotDetector struct{}

// Detect emits coverageBlindSpot signals from existing coverage insights.
func (d *CoverageBlindSpotDetector) Detect(snap *models.TestSuiteSnapshot) []models.Signal {
	if snap == nil || len(snap.CoverageInsights) == 0 {
		return nil
	}

	seen := map[string]bool{}
	var out []models.Signal
	for _, ci := range snap.CoverageInsights {
		if !isBlindSpotCoverageInsight(ci.Type) {
			continue
		}

		key := ci.Type + "|" + ci.Path + "|" + ci.UnitID + "|" + ci.Description
		if seen[key] {
			continue
		}
		seen[key] = true

		loc := models.SignalLocation{Repository: "coverage"}
		if ci.Path != "" {
			loc.File = ci.Path
			loc.Repository = ""
		}

		s := models.Signal{
			Type:             "coverageBlindSpot",
			Category:         models.CategoryQuality,
			Severity:         coverageInsightSeverity(ci.Severity),
			Confidence:       0.9,
			EvidenceStrength: models.EvidenceStrong,
			EvidenceSource:   models.SourceCoverage,
			Location:         loc,
			Explanation:      ci.Description,
			SuggestedAction:  ci.SuggestedAction,
			Metadata: map[string]any{
				"insightType": ci.Type,
			},
		}
		if ci.UnitID != "" {
			s.Metadata["unitId"] = ci.UnitID
		}
		out = append(out, s)
	}

	return out
}

func isBlindSpotCoverageInsight(insightType string) bool {
	switch insightType {
	case "only_e2e_coverage", "only_e2e_unit", "uncovered_exported", "weak_coverage_diversity":
		return true
	default:
		return false
	}
}

func coverageInsightSeverity(sev string) models.SignalSeverity {
	switch sev {
	case "critical":
		return models.SeverityCritical
	case "high":
		return models.SeverityHigh
	case "medium":
		return models.SeverityMedium
	case "low":
		return models.SeverityLow
	default:
		return models.SeverityInfo
	}
}
