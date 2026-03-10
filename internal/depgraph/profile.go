package depgraph

// RepoProfile captures the structural characteristics of a repository's
// test system, enabling adaptive recommendations.
type RepoProfile struct {
	// TestVolume classifies the number of tests.
	TestVolume string `json:"testVolume"` // tiny, small, medium, large

	// CIPresure estimates CI runtime burden.
	CIPressure string `json:"ciPressure"` // low, medium, high

	// CoverageConfidence indicates how much of the codebase is structurally covered.
	CoverageConfidence string `json:"coverageConfidence"` // low, medium, high

	// RedundancyLevel measures test duplication.
	RedundancyLevel string `json:"redundancyLevel"` // low, medium, high

	// FanoutBurden measures the proportion of high-fanout nodes.
	FanoutBurden string `json:"fanoutBurden"` // low, medium, high

	// SkipBurden measures the proportion of skipped tests.
	SkipBurden string `json:"skipBurden"` // low, medium, high

	// FlakeBurden measures the proportion of flaky tests.
	FlakeBurden string `json:"flakeBurden"` // low, medium, high

	// GraphDensity is the raw edge-to-node ratio.
	GraphDensity float64 `json:"graphDensity"`
}

// ProfileInsights bundles engine results used for profiling.
type ProfileInsights struct {
	Coverage   *CoverageResult
	Duplicates *DuplicateResult
	Fanout     *FanoutResult
}

// AnalyzeProfile classifies a repository based on its graph structure
// and engine outputs.
func AnalyzeProfile(g *Graph, insights ProfileInsights) RepoProfile {
	stats := g.Stats()

	profile := RepoProfile{
		TestVolume:         classifyVolume(stats),
		CIPressure:         classifyCIPressure(stats),
		CoverageConfidence: classifyCoverageConfidence(insights.Coverage, stats),
		RedundancyLevel:    classifyRedundancy(insights.Duplicates),
		FanoutBurden:       classifyFanoutBurden(insights.Fanout),
		GraphDensity:       stats.Density,
	}

	return profile
}

func classifyVolume(stats Stats) string {
	testCount := stats.NodesByType[string(NodeTest)]
	switch {
	case testCount <= 10:
		return "tiny"
	case testCount <= 100:
		return "small"
	case testCount <= 1000:
		return "medium"
	default:
		return "large"
	}
}

func classifyCIPressure(stats Stats) string {
	testCount := stats.NodesByType[string(NodeTest)]
	switch {
	case testCount <= 50:
		return "low"
	case testCount <= 500:
		return "medium"
	default:
		return "high"
	}
}

func classifyCoverageConfidence(cov *CoverageResult, stats Stats) string {
	if cov == nil || cov.SourceCount == 0 {
		return "low"
	}

	lowPct := float64(cov.BandCounts[CoverageBandLow]) / float64(cov.SourceCount)

	switch {
	case lowPct > 0.6:
		return "low"
	case lowPct > 0.3 || stats.Density < 0.001:
		return "medium"
	default:
		return "high"
	}
}

func classifyRedundancy(dupes *DuplicateResult) string {
	if dupes == nil || dupes.TestsAnalyzed == 0 {
		return "low"
	}

	ratio := float64(dupes.DuplicateCount) / float64(dupes.TestsAnalyzed)
	switch {
	case ratio > 0.30:
		return "high"
	case ratio > 0.10:
		return "medium"
	default:
		return "low"
	}
}

// EnrichProfileWithHealthRatios sets skip and flake burden on a profile
// using ratios computed from runtime health metrics.
func EnrichProfileWithHealthRatios(profile *RepoProfile, skipRatio, flakeRatio float64) {
	profile.SkipBurden = classifyRatioBurden(skipRatio)
	profile.FlakeBurden = classifyRatioBurden(flakeRatio)
}

func classifyRatioBurden(ratio float64) string {
	switch {
	case ratio > 0.20:
		return "high"
	case ratio > 0.05:
		return "medium"
	default:
		return "low"
	}
}

func classifyFanoutBurden(fanout *FanoutResult) string {
	if fanout == nil || fanout.NodeCount == 0 {
		return "low"
	}

	ratio := float64(fanout.FlaggedCount) / float64(fanout.NodeCount)
	switch {
	case ratio > 0.30:
		return "high"
	case ratio > 0.10:
		return "medium"
	default:
		return "low"
	}
}
