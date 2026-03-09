// Package portfolio implements test portfolio intelligence for Hamlet.
//
// Portfolio intelligence treats the test suite as a portfolio of assets
// with costs, protections, and strategic value. It sits above raw signals
// and measurements, providing investment-oriented insights about efficiency,
// redundancy, breadth, and leverage.
//
// Key concepts:
//   - TestAsset: a test file viewed as an asset with cost, reach, and stability
//   - ProtectionBreadth: how much surface a test covers
//   - RuntimeCost: observed or inferred operational cost
//   - Redundancy: overlap between tests covering the same surface
//   - Overbreadth: tests that are excessively broad relative to their diagnostic value
//   - Leverage: tests that protect important surface efficiently
//
// Portfolio intelligence is grounded in observed evidence, not speculation.
package portfolio

import "github.com/pmclSF/hamlet/internal/models"

// BreadthClass describes the relative scope of a test.
type BreadthClass string

const (
	BreadthNarrow  BreadthClass = "narrow"  // targets specific units
	BreadthModerate BreadthClass = "moderate" // covers a module/package
	BreadthBroad   BreadthClass = "broad"   // spans multiple modules or owners
	BreadthUnknown BreadthClass = "unknown" // insufficient data
)

// CostClass describes the relative operational cost of a test.
type CostClass string

const (
	CostLow     CostClass = "low"
	CostModerate CostClass = "moderate"
	CostHigh    CostClass = "high"
	CostUnknown CostClass = "unknown"
)

// CandidateConfidence describes confidence in a portfolio finding.
type CandidateConfidence string

const (
	ConfidenceHigh   CandidateConfidence = "high"
	ConfidenceModerate CandidateConfidence = "moderate"
	ConfidenceLow    CandidateConfidence = "low"
)

// TestAsset represents a test file viewed as a portfolio asset with
// cost, protection reach, stability, and evidence metadata.
type TestAsset struct {
	// Path is the test file path (primary key).
	Path string `json:"path"`

	// Framework is the test framework name.
	Framework string `json:"framework"`

	// TestType is the inferred test type (unit, integration, e2e).
	TestType string `json:"testType"`

	// Owner is the resolved owner.
	Owner string `json:"owner"`

	// TestCount is the number of tests in this file.
	TestCount int `json:"testCount"`

	// --- Cost metrics ---

	// RuntimeMs is the observed average runtime in milliseconds (0 if unknown).
	RuntimeMs float64 `json:"runtimeMs,omitempty"`

	// RetryRate is the observed retry rate (0.0-1.0).
	RetryRate float64 `json:"retryRate,omitempty"`

	// PassRate is the observed pass rate (0.0-1.0, 0 if unknown).
	PassRate float64 `json:"passRate,omitempty"`

	// CostClass is the inferred cost classification.
	CostClass CostClass `json:"costClass"`

	// InstabilitySignals is the count of health signals (flaky, slow, etc.).
	InstabilitySignals int `json:"instabilitySignals,omitempty"`

	// --- Protection metrics ---

	// CoveredUnitCount is the number of code units this test covers.
	CoveredUnitCount int `json:"coveredUnitCount"`

	// CoveredModules is the set of directories/modules touched.
	CoveredModules []string `json:"coveredModules,omitempty"`

	// ExportedUnitsCovered is the number of exported code units covered.
	ExportedUnitsCovered int `json:"exportedUnitsCovered,omitempty"`

	// OwnersCovered is the set of distinct owners whose code is covered.
	OwnersCovered []string `json:"ownersCovered,omitempty"`

	// BreadthClass is the inferred breadth classification.
	BreadthClass BreadthClass `json:"breadthClass"`

	// ImportedSources lists source files this test imports (from import graph).
	// Used for precise redundancy detection based on actual code linkage.
	ImportedSources []string `json:"importedSources,omitempty"`

	// --- Evidence ---

	// HasRuntimeData is true if runtime data was available for cost estimation.
	HasRuntimeData bool `json:"hasRuntimeData"`

	// HasCoverageData is true if coverage linkage was available for reach estimation.
	HasCoverageData bool `json:"hasCoverageData"`
}

// Finding represents a portfolio-level insight or recommendation.
type Finding struct {
	// Type is the finding type (redundancy_candidate, overbroad, low_value_high_cost, high_leverage).
	Type string `json:"type"`

	// Path is the primary test file path.
	Path string `json:"path"`

	// RelatedPaths are other test file paths involved (e.g., for redundancy pairs).
	RelatedPaths []string `json:"relatedPaths,omitempty"`

	// Owner is the resolved owner of the primary path.
	Owner string `json:"owner"`

	// Confidence is how confident the finding is.
	Confidence CandidateConfidence `json:"confidence"`

	// Explanation describes what was observed and why it matters.
	Explanation string `json:"explanation"`

	// SuggestedAction describes what the team could do.
	SuggestedAction string `json:"suggestedAction"`

	// Metadata carries type-specific detail.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Finding types.
const (
	FindingRedundancyCandidate = "redundancy_candidate"
	FindingOverbroad           = "overbroad"
	FindingLowValueHighCost    = "low_value_high_cost"
	FindingHighLeverage        = "high_leverage"
)

// PortfolioSummary is the top-level output of portfolio analysis.
type PortfolioSummary struct {
	// Assets is the list of test assets.
	Assets []TestAsset `json:"assets"`

	// Findings is the list of portfolio findings (redundancy, overbreadth, etc.).
	Findings []Finding `json:"findings"`

	// Aggregates contains summary statistics.
	Aggregates PortfolioAggregates `json:"aggregates"`
}

// PortfolioAggregates contains summary statistics for the portfolio.
type PortfolioAggregates struct {
	// TotalAssets is the total number of test assets.
	TotalAssets int `json:"totalAssets"`

	// TotalRuntimeMs is the sum of all observed runtime (0 if no runtime data).
	TotalRuntimeMs float64 `json:"totalRuntimeMs,omitempty"`

	// RuntimeConcentration is the share of total runtime consumed by top 20% of tests.
	RuntimeConcentration float64 `json:"runtimeConcentration,omitempty"`

	// HasRuntimeData is true if any test asset has runtime data.
	HasRuntimeData bool `json:"hasRuntimeData"`

	// HasCoverageData is true if any test asset has coverage data.
	HasCoverageData bool `json:"hasCoverageData"`

	// RedundancyCandidateCount is the number of redundancy candidate findings.
	RedundancyCandidateCount int `json:"redundancyCandidateCount"`

	// OverbroadCount is the number of overbroad test findings.
	OverbroadCount int `json:"overbroadCount"`

	// LowValueHighCostCount is the number of low-value high-cost findings.
	LowValueHighCostCount int `json:"lowValueHighCostCount"`

	// HighLeverageCount is the number of high-leverage test findings.
	HighLeverageCount int `json:"highLeverageCount"`

	// ByOwner contains per-owner aggregations.
	ByOwner []OwnerPortfolioSummary `json:"byOwner,omitempty"`
}

// OwnerPortfolioSummary aggregates portfolio findings per owner.
type OwnerPortfolioSummary struct {
	Owner                    string  `json:"owner"`
	AssetCount               int     `json:"assetCount"`
	TotalRuntimeMs           float64 `json:"totalRuntimeMs,omitempty"`
	RedundancyCandidateCount int     `json:"redundancyCandidateCount,omitempty"`
	OverbroadCount           int     `json:"overbroadCount,omitempty"`
	LowValueHighCostCount    int     `json:"lowValueHighCostCount,omitempty"`
	HighLeverageCount        int     `json:"highLeverageCount,omitempty"`
}

// BenchmarkAggregate contains privacy-safe portfolio aggregates for export.
type BenchmarkAggregate struct {
	// RuntimeConcentrationBand describes how concentrated runtime is.
	RuntimeConcentrationBand string `json:"runtimeConcentrationBand"`

	// RedundancyCandidateShareBand describes the share of redundancy candidates.
	RedundancyCandidateShareBand string `json:"redundancyCandidateShareBand"`

	// OverbroadShareBand describes the share of overbroad tests.
	OverbroadShareBand string `json:"overbroadShareBand"`

	// LowValueHighCostShareBand describes the share of low-value high-cost tests.
	LowValueHighCostShareBand string `json:"lowValueHighCostShareBand"`

	// HighLeverageShareBand describes the share of high-leverage tests.
	HighLeverageShareBand string `json:"highLeverageShareBand"`

	// PortfolioPostureBand is the overall portfolio balance posture.
	PortfolioPostureBand string `json:"portfolioPostureBand"`
}

// frameworkType resolves the framework type from a snapshot for a given framework name.
func frameworkType(snap *models.TestSuiteSnapshot, name string) models.FrameworkType {
	for _, fw := range snap.Frameworks {
		if fw.Name == name {
			return fw.Type
		}
	}
	return ""
}
