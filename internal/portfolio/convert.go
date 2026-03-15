package portfolio

import "github.com/pmclSF/terrain/internal/models"

// ToModel converts a PortfolioSummary to the serializable model type.
func (s *PortfolioSummary) ToModel() *models.PortfolioSnapshot {
	if s == nil {
		return nil
	}

	ps := &models.PortfolioSnapshot{
		Aggregates: models.PortfolioAggregates{
			TotalAssets:              s.Aggregates.TotalAssets,
			TotalRuntimeMs:           s.Aggregates.TotalRuntimeMs,
			RuntimeConcentration:     s.Aggregates.RuntimeConcentration,
			HasRuntimeData:           s.Aggregates.HasRuntimeData,
			HasCoverageData:          s.Aggregates.HasCoverageData,
			RedundancyCandidateCount: s.Aggregates.RedundancyCandidateCount,
			OverbroadCount:           s.Aggregates.OverbroadCount,
			LowValueHighCostCount:    s.Aggregates.LowValueHighCostCount,
			HighLeverageCount:        s.Aggregates.HighLeverageCount,
			PortfolioPostureBand:     computePortfolioPosture(s),
		},
	}

	for _, a := range s.Assets {
		ps.Assets = append(ps.Assets, models.PortfolioAsset{
			Path:                 a.Path,
			Framework:            a.Framework,
			TestType:             a.TestType,
			Owner:                a.Owner,
			TestCount:            a.TestCount,
			RuntimeMs:            a.RuntimeMs,
			RetryRate:            a.RetryRate,
			PassRate:             a.PassRate,
			CostClass:            string(a.CostClass),
			InstabilitySignals:   a.InstabilitySignals,
			CoveredUnitCount:     a.CoveredUnitCount,
			CoveredModules:       a.CoveredModules,
			ExportedUnitsCovered: a.ExportedUnitsCovered,
			OwnersCovered:        a.OwnersCovered,
			BreadthClass:         string(a.BreadthClass),
			HasRuntimeData:       a.HasRuntimeData,
			HasCoverageData:      a.HasCoverageData,
		})
	}

	for _, o := range s.Aggregates.ByOwner {
		ps.Aggregates.ByOwner = append(ps.Aggregates.ByOwner, models.PortfolioOwnerSummary{
			Owner:                    o.Owner,
			AssetCount:               o.AssetCount,
			TotalRuntimeMs:           o.TotalRuntimeMs,
			RedundancyCandidateCount: o.RedundancyCandidateCount,
			OverbroadCount:           o.OverbroadCount,
			LowValueHighCostCount:    o.LowValueHighCostCount,
			HighLeverageCount:        o.HighLeverageCount,
		})
	}

	for _, f := range s.Findings {
		ps.Findings = append(ps.Findings, models.PortfolioFinding{
			Type:            f.Type,
			Path:            f.Path,
			RelatedPaths:    f.RelatedPaths,
			Owner:           f.Owner,
			Confidence:      string(f.Confidence),
			Explanation:     f.Explanation,
			SuggestedAction: f.SuggestedAction,
			Metadata:        f.Metadata,
		})
	}

	return ps
}
