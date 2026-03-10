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
// Levels are ordered by increasing conservatism via their numeric values.
type FallbackLevel int

const (
	FallbackDirectDeps      FallbackLevel = iota // least conservative
	FallbackFixtureExpand                        // expand fixture dependents
	FallbackPackageTests                         // run all package tests
	FallbackSmokeRegression                      // smoke + regression suite
	FallbackFullSuite                            // most conservative — run everything
)

// fallbackLevelNames maps FallbackLevel values to their JSON string representations.
var fallbackLevelNames = map[FallbackLevel]string{
	FallbackDirectDeps:      "DirectDeps",
	FallbackFixtureExpand:   "FixtureExpansion",
	FallbackPackageTests:    "PackageTests",
	FallbackSmokeRegression: "SmokeRegression",
	FallbackFullSuite:       "FullSuite",
}

// fallbackLevelValues maps JSON string representations to FallbackLevel values.
var fallbackLevelValues = map[string]FallbackLevel{
	"DirectDeps":      FallbackDirectDeps,
	"FixtureExpansion": FallbackFixtureExpand,
	"PackageTests":    FallbackPackageTests,
	"SmokeRegression": FallbackSmokeRegression,
	"FullSuite":       FallbackFullSuite,
}

// String returns the string representation of a FallbackLevel.
func (f FallbackLevel) String() string {
	if s, ok := fallbackLevelNames[f]; ok {
		return s
	}
	return fmt.Sprintf("FallbackLevel(%d)", int(f))
}

// MarshalText implements encoding.TextMarshaler for JSON serialization.
func (f FallbackLevel) MarshalText() ([]byte, error) {
	if s, ok := fallbackLevelNames[f]; ok {
		return []byte(s), nil
	}
	return nil, fmt.Errorf("unknown FallbackLevel: %d", int(f))
}

// UnmarshalText implements encoding.TextUnmarshaler for JSON deserialization.
func (f *FallbackLevel) UnmarshalText(text []byte) error {
	if v, ok := fallbackLevelValues[string(text)]; ok {
		*f = v
		return nil
	}
	return fmt.Errorf("unknown FallbackLevel: %q", string(text))
}

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

	if profile.SkipBurden == "high" {
		cases = append(cases, EdgeCase{
			Type:        EdgeCaseHighSkipBurden,
			Severity:    "caution",
			Description: "High proportion of skipped tests — optimization may select already-skipped tests.",
		})
	}

	if profile.FlakeBurden == "high" {
		cases = append(cases, EdgeCase{
			Type:        EdgeCaseHighFlakeBurden,
			Severity:    "caution",
			Description: "High proportion of flaky tests — selected tests may produce unreliable results.",
		})
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

		case EdgeCaseHighSkipBurden:
			if policy.FallbackLevel < FallbackPackageTests {
				policy.FallbackLevel = FallbackPackageTests
			}
			policy.ConfidenceAdjustment *= 0.85
			policy.Recommendations = append(policy.Recommendations,
				"High skip burden detected. Review skipped tests before relying on test selection.")

		case EdgeCaseHighFlakeBurden:
			if policy.FallbackLevel < FallbackPackageTests {
				policy.FallbackLevel = FallbackPackageTests
			}
			policy.ConfidenceAdjustment *= 0.75
			policy.RiskElevated = true
			policy.Recommendations = append(policy.Recommendations,
				"High flake burden undermines test reliability. Stabilize flaky tests to improve selection confidence.")

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
