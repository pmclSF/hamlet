package portfolio

import (
	"math"
	"sort"

	"github.com/pmclSF/hamlet/internal/models"
)

// Analyze runs the full portfolio intelligence pipeline on a snapshot.
// It builds test assets, detects findings, and computes aggregates.
func Analyze(snap *models.TestSuiteSnapshot) *PortfolioSummary {
	if snap == nil || len(snap.TestFiles) == 0 {
		return &PortfolioSummary{}
	}

	assets := BuildAssets(snap)

	// Detect findings.
	var findings []Finding
	findings = append(findings, detectRedundancy(assets)...)
	findings = append(findings, detectOverbroad(assets)...)
	findings = append(findings, detectLowValueHighCost(assets)...)
	findings = append(findings, detectHighLeverage(assets)...)

	// Sort findings by type then path for determinism.
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Type != findings[j].Type {
			return findings[i].Type < findings[j].Type
		}
		return findings[i].Path < findings[j].Path
	})

	agg := computeAggregates(assets, findings, snap)

	return &PortfolioSummary{
		Assets:     assets,
		Findings:   findings,
		Aggregates: agg,
	}
}

// computeAggregates derives summary statistics from assets and findings.
func computeAggregates(assets []TestAsset, findings []Finding, snap *models.TestSuiteSnapshot) PortfolioAggregates {
	agg := PortfolioAggregates{
		TotalAssets: len(assets),
	}

	// Runtime aggregation.
	var totalRuntime float64
	var runtimes []float64
	for _, a := range assets {
		if a.HasRuntimeData {
			agg.HasRuntimeData = true
			totalRuntime += a.RuntimeMs
			runtimes = append(runtimes, a.RuntimeMs)
		}
		if a.HasCoverageData {
			agg.HasCoverageData = true
		}
	}
	agg.TotalRuntimeMs = totalRuntime

	// Runtime concentration: share consumed by top 20%.
	if len(runtimes) >= 5 {
		sort.Float64s(runtimes)
		top20Count := int(math.Ceil(float64(len(runtimes)) * 0.20))
		top20Sum := 0.0
		for i := len(runtimes) - top20Count; i < len(runtimes); i++ {
			top20Sum += runtimes[i]
		}
		if totalRuntime > 0 {
			agg.RuntimeConcentration = top20Sum / totalRuntime
		}
	}

	// Finding counts.
	for _, f := range findings {
		switch f.Type {
		case FindingRedundancyCandidate:
			agg.RedundancyCandidateCount++
		case FindingOverbroad:
			agg.OverbroadCount++
		case FindingLowValueHighCost:
			agg.LowValueHighCostCount++
		case FindingHighLeverage:
			agg.HighLeverageCount++
		}
	}

	// Per-owner aggregations.
	agg.ByOwner = computeOwnerAggregates(assets, findings)

	return agg
}

// computeOwnerAggregates groups portfolio data by owner.
func computeOwnerAggregates(assets []TestAsset, findings []Finding) []OwnerPortfolioSummary {
	type ownerData struct {
		assetCount  int
		runtimeMs   float64
		redundancy  int
		overbroad   int
		lowValue    int
		highLev     int
	}

	byOwner := map[string]*ownerData{}
	for _, a := range assets {
		d, ok := byOwner[a.Owner]
		if !ok {
			d = &ownerData{}
			byOwner[a.Owner] = d
		}
		d.assetCount++
		d.runtimeMs += a.RuntimeMs
	}

	for _, f := range findings {
		d, ok := byOwner[f.Owner]
		if !ok {
			d = &ownerData{}
			byOwner[f.Owner] = d
		}
		switch f.Type {
		case FindingRedundancyCandidate:
			d.redundancy++
		case FindingOverbroad:
			d.overbroad++
		case FindingLowValueHighCost:
			d.lowValue++
		case FindingHighLeverage:
			d.highLev++
		}
	}

	result := make([]OwnerPortfolioSummary, 0, len(byOwner))
	for owner, d := range byOwner {
		result = append(result, OwnerPortfolioSummary{
			Owner:                    owner,
			AssetCount:               d.assetCount,
			TotalRuntimeMs:           d.runtimeMs,
			RedundancyCandidateCount: d.redundancy,
			OverbroadCount:           d.overbroad,
			LowValueHighCostCount:    d.lowValue,
			HighLeverageCount:        d.highLev,
		})
	}

	// Sort by owner name for determinism.
	sort.Slice(result, func(i, j int) bool {
		return result[i].Owner < result[j].Owner
	})

	return result
}

// BuildBenchmarkAggregate creates a privacy-safe aggregate for export.
// No file paths, owner names, or identifying details are included.
func BuildBenchmarkAggregate(summary *PortfolioSummary) *BenchmarkAggregate {
	if summary == nil || summary.Aggregates.TotalAssets == 0 {
		return nil
	}

	total := float64(summary.Aggregates.TotalAssets)

	return &BenchmarkAggregate{
		RuntimeConcentrationBand:    concentrationBand(summary.Aggregates.RuntimeConcentration),
		RedundancyCandidateShareBand: shareBand(float64(summary.Aggregates.RedundancyCandidateCount) / total),
		OverbroadShareBand:          shareBand(float64(summary.Aggregates.OverbroadCount) / total),
		LowValueHighCostShareBand:   shareBand(float64(summary.Aggregates.LowValueHighCostCount) / total),
		HighLeverageShareBand:       shareBand(float64(summary.Aggregates.HighLeverageCount) / total),
		PortfolioPostureBand:        computePortfolioPosture(summary),
	}
}

// concentrationBand classifies runtime concentration into bands.
func concentrationBand(ratio float64) string {
	switch {
	case ratio <= 0:
		return "unknown"
	case ratio <= 0.50:
		return "balanced"
	case ratio <= 0.70:
		return "moderate"
	case ratio <= 0.85:
		return "concentrated"
	default:
		return "highly_concentrated"
	}
}

// shareBand classifies a finding share into bands.
func shareBand(ratio float64) string {
	switch {
	case ratio <= 0.02:
		return "minimal"
	case ratio <= 0.10:
		return "low"
	case ratio <= 0.25:
		return "moderate"
	default:
		return "high"
	}
}

// computePortfolioPosture derives an overall portfolio posture band.
// It considers the ratio of problematic findings to total assets.
func computePortfolioPosture(summary *PortfolioSummary) string {
	total := summary.Aggregates.TotalAssets
	if total == 0 {
		return "unknown"
	}

	// Count problematic findings (exclude high_leverage which is positive).
	problems := summary.Aggregates.RedundancyCandidateCount +
		summary.Aggregates.OverbroadCount +
		summary.Aggregates.LowValueHighCostCount

	ratio := float64(problems) / float64(total)

	switch {
	case ratio <= 0.05:
		return "strong"
	case ratio <= 0.15:
		return "moderate"
	case ratio <= 0.30:
		return "weak"
	default:
		return "critical"
	}
}
