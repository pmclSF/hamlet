package depgraph

import "fmt"

// EdgeCaseType identifies a specific edge case condition.
type EdgeCaseType string

const (
	EdgeCaseFewTests          EdgeCaseType = "FEW_TESTS"
	EdgeCaseFastCIAlready     EdgeCaseType = "FAST_CI_ALREADY"
	EdgeCaseRedundantSuite    EdgeCaseType = "REDUNDANT_TEST_SUITE"
	EdgeCaseHighSkipBurden    EdgeCaseType = "HIGH_SKIP_BURDEN"
	EdgeCaseHighFlakeBurden   EdgeCaseType = "HIGH_FLAKE_BURDEN"
	EdgeCaseHighFanoutFixture EdgeCaseType = "HIGH_FANOUT_FIXTURE"
	EdgeCaseLowGraphVisibility EdgeCaseType = "LOW_GRAPH_VISIBILITY"
)

// EdgeCase represents a detected edge case condition.
type EdgeCase struct {
	Type        EdgeCaseType `json:"type"`
	Severity    string       `json:"severity"` // warning, caution, critical
	Description string       `json:"description"`
}

// FallbackLevel indicates how conservative the system should be.
type FallbackLevel string

const (
	FallbackDirectDeps      FallbackLevel = "DirectDeps"
	FallbackFixtureExpand   FallbackLevel = "FixtureExpansion"
	FallbackPackageTests    FallbackLevel = "PackageTests"
	FallbackSmokeRegression FallbackLevel = "SmokeRegression"
	FallbackFullSuite       FallbackLevel = "FullSuite"
)

// Policy captures the recommendations derived from edge case analysis.
type Policy struct {
	// FallbackLevel indicates how conservative test selection should be.
	FallbackLevel FallbackLevel `json:"fallbackLevel"`

	// ConfidenceAdjustment is a multiplier (0–1) applied to confidence scores.
	ConfidenceAdjustment float64 `json:"confidenceAdjustment"`

	// OptimizationDisabled indicates whether test selection optimization
	// should be disabled entirely.
	OptimizationDisabled bool `json:"optimizationDisabled"`

	// RiskElevated indicates whether the risk flag should be raised.
	RiskElevated bool `json:"riskElevated"`

	// Recommendations contains human-readable guidance.
	Recommendations []string `json:"recommendations"`
}

// DetectEdgeCases identifies edge case conditions based on the repo profile,
// graph structure, and engine insights.
func DetectEdgeCases(profile RepoProfile, g *Graph, insights ProfileInsights) []EdgeCase {
	var cases []EdgeCase
	stats := g.Stats()
	testCount := stats.NodesByType[string(NodeTest)]

	if testCount <= 10 {
		cases = append(cases, EdgeCase{
			Type:        EdgeCaseFewTests,
			Severity:    "critical",
			Description: fmt.Sprintf("Only %d tests discovered — too few for meaningful optimization.", testCount),
		})
	}

	if profile.CIPressure == "low" {
		cases = append(cases, EdgeCase{
			Type:        EdgeCaseFastCIAlready,
			Severity:    "warning",
			Description: "CI is already fast — optimization may yield minimal benefit.",
		})
	}

	if profile.RedundancyLevel == "high" {
		cases = append(cases, EdgeCase{
			Type:        EdgeCaseRedundantSuite,
			Severity:    "caution",
			Description: "High test duplication detected — consider consolidating redundant tests before optimizing.",
		})
	}

	if insights.Fanout != nil && insights.Fanout.FlaggedCount > 0 {
		ratio := float64(insights.Fanout.FlaggedCount) / float64(insights.Fanout.NodeCount)
		if ratio > 0.3 {
			cases = append(cases, EdgeCase{
				Type:        EdgeCaseHighFanoutFixture,
				Severity:    "caution",
				Description: fmt.Sprintf("%.0f%% of nodes have excessive fanout — fragile test architecture.", ratio*100),
			})
		}
	}

	if profile.CoverageConfidence == "low" {
		cases = append(cases, EdgeCase{
			Type:        EdgeCaseLowGraphVisibility,
			Severity:    "warning",
			Description: "Low graph visibility — most source files have no structural test coverage.",
		})
	}

	return cases
}

// ApplyEdgeCasePolicy derives a policy from detected edge cases.
func ApplyEdgeCasePolicy(cases []EdgeCase, profile RepoProfile) Policy {
	policy := Policy{
		FallbackLevel:        FallbackDirectDeps,
		ConfidenceAdjustment: 1.0,
	}

	for _, c := range cases {
		switch c.Type {
		case EdgeCaseFewTests:
			policy.OptimizationDisabled = true
			policy.FallbackLevel = FallbackFullSuite
			policy.ConfidenceAdjustment *= 0.5
			policy.RiskElevated = true
			policy.Recommendations = append(policy.Recommendations,
				"Too few tests for meaningful optimization. Focus on expanding test coverage first.")

		case EdgeCaseFastCIAlready:
			policy.Recommendations = append(policy.Recommendations,
				"CI is already fast. Test selection would yield minimal time savings.")

		case EdgeCaseRedundantSuite:
			if policy.FallbackLevel < FallbackPackageTests {
				policy.FallbackLevel = FallbackPackageTests
			}
			policy.ConfidenceAdjustment *= 0.8
			policy.Recommendations = append(policy.Recommendations,
				"High test duplication detected. Consider consolidating redundant tests to reduce CI noise.")

		case EdgeCaseHighFanoutFixture:
			if policy.FallbackLevel < FallbackFixtureExpand {
				policy.FallbackLevel = FallbackFixtureExpand
			}
			policy.ConfidenceAdjustment *= 0.7
			policy.Recommendations = append(policy.Recommendations,
				"High-fanout fixtures create fragile dependencies. Consider breaking down shared fixtures.")

		case EdgeCaseLowGraphVisibility:
			if policy.FallbackLevel < FallbackSmokeRegression {
				policy.FallbackLevel = FallbackSmokeRegression
			}
			policy.ConfidenceAdjustment *= 0.6
			policy.RiskElevated = true
			policy.Recommendations = append(policy.Recommendations,
				"Low graph visibility limits confidence in impact analysis. Recommendations may be incomplete.")
		}
	}

	// Clamp confidence.
	if policy.ConfidenceAdjustment < 0.1 {
		policy.ConfidenceAdjustment = 0.1
	}

	return policy
}
