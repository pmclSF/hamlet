package measurement

import (
	"fmt"

	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/signals"
)

// CoverageDepthMeasurements returns measurement definitions for the
// coverage_depth dimension.
func CoverageDepthMeasurements() []Definition {
	return []Definition{
		{
			ID:          "coverage_depth.uncovered_exports",
			Dimension:   DimensionCoverageDepth,
			Description: "Share of exported code units without any linked tests.",
			Units:       UnitsRatio,
			Inputs:      []string{string(signals.SignalUntestedExport)},
			Compute:     computeUncoveredExports,
		},
		{
			ID:          "coverage_depth.weak_assertion_share",
			Dimension:   DimensionCoverageDepth,
			Description: "Share of test files with weak assertion density.",
			Units:       UnitsRatio,
			Inputs:      []string{string(signals.SignalWeakAssertion)},
			Compute:     computeWeakAssertionShare,
		},
		{
			ID:          "coverage_depth.coverage_breach_share",
			Dimension:   DimensionCoverageDepth,
			Description: "Share of coverage areas below threshold.",
			Units:       UnitsRatio,
			Inputs:      []string{string(signals.SignalCoverageThresholdBreak)},
			Compute:     computeCoverageBreachShare,
		},
	}
}

// CoverageDiversityMeasurements returns measurement definitions for the
// coverage_diversity dimension.
func CoverageDiversityMeasurements() []Definition {
	return []Definition{
		{
			ID:          "coverage_diversity.mock_heavy_share",
			Dimension:   DimensionCoverageDiversity,
			Description: "Share of test files dominated by mocks over assertions.",
			Units:       UnitsRatio,
			Inputs:      []string{string(signals.SignalMockHeavyTest)},
			Compute:     computeMockHeavyShare,
		},
		{
			ID:          "coverage_diversity.framework_fragmentation",
			Dimension:   DimensionCoverageDiversity,
			Description: "Framework count relative to test suite size.",
			Units:       UnitsRatio,
			Inputs:      []string{"frameworks"},
			Compute:     computeFrameworkFragmentation,
		},
		{
			ID:          "coverage_diversity.e2e_concentration",
			Dimension:   DimensionCoverageDiversity,
			Description: "Share of test files using E2E frameworks.",
			Units:       UnitsRatio,
			Inputs:      []string{"frameworks"},
			Compute:     computeE2EConcentration,
		},
		{
			ID:          "coverage_diversity.e2e_only_units",
			Dimension:   DimensionCoverageDiversity,
			Description: "Share of code units covered only by e2e tests.",
			Units:       UnitsRatio,
			Inputs:      []string{"coverageSummary"},
			Compute:     computeE2EOnlyUnits,
		},
		{
			ID:          "coverage_diversity.unit_test_coverage",
			Dimension:   DimensionCoverageDiversity,
			Description: "Share of code units covered by unit tests.",
			Units:       UnitsRatio,
			Inputs:      []string{"coverageSummary"},
			Compute:     computeUnitTestCoverage,
		},
	}
}

func computeUncoveredExports(snap *models.TestSuiteSnapshot) Result {
	exported := 0
	for _, cu := range snap.CodeUnits {
		if cu.Exported {
			exported++
		}
	}

	if exported == 0 {
		return Result{
			ID: "coverage_depth.uncovered_exports", Dimension: DimensionCoverageDepth,
			Value: 0, Units: UnitsRatio, Band: "strong",
			Evidence:    EvidenceNone,
			Explanation: "No exported code units detected.",
			Limitations: []string{"Code unit discovery may be incomplete."},
		}
	}

	untestedCount := countSignals(snap, signals.SignalUntestedExport)
	ratio := float64(untestedCount) / float64(exported)
	band := ratioToBand(ratio, 0.10, 0.30, 0.50)

	return Result{
		ID: "coverage_depth.uncovered_exports", Dimension: DimensionCoverageDepth,
		Value: ratio, Units: UnitsRatio, Band: band,
		Evidence:    EvidencePartial,
		Explanation: fmt.Sprintf("%d of %d exported code unit(s) appear untested (%.0f%%).", untestedCount, exported, ratio*100),
		Inputs:      []string{"untestedExport"},
		Limitations: []string{"Test linkage is heuristic-based; some coverage may exist but not be detected."},
	}
}

func computeWeakAssertionShare(snap *models.TestSuiteSnapshot) Result {
	total := len(snap.TestFiles)
	if total == 0 {
		return Result{
			ID: "coverage_depth.weak_assertion_share", Dimension: DimensionCoverageDepth,
			Value: 0, Units: UnitsRatio, Band: "strong",
			Evidence: EvidenceNone, Explanation: "No test files detected.",
		}
	}

	count := countSignals(snap, signals.SignalWeakAssertion)
	ratio := float64(count) / float64(total)
	band := ratioToBand(ratio, 0.10, 0.25, 0.50)

	return Result{
		ID: "coverage_depth.weak_assertion_share", Dimension: DimensionCoverageDepth,
		Value: ratio, Units: UnitsRatio, Band: band,
		Evidence:    EvidenceStrong,
		Explanation: fmt.Sprintf("%d of %d test file(s) have weak assertion density (%.0f%%).", count, total, ratio*100),
		Inputs:      []string{"weakAssertion"},
	}
}

func computeCoverageBreachShare(snap *models.TestSuiteSnapshot) Result {
	total := len(snap.TestFiles)
	if total == 0 {
		return Result{
			ID: "coverage_depth.coverage_breach_share", Dimension: DimensionCoverageDepth,
			Value: 0, Units: UnitsRatio, Band: "strong",
			Evidence: EvidenceNone, Explanation: "No test files detected.",
		}
	}

	count := countSignals(snap, signals.SignalCoverageThresholdBreak)
	if count == 0 {
		// Distinguish between "no breaches" and "no coverage data."
		hasCoverage := false
		for _, s := range snap.Signals {
			if s.Type == signals.SignalCoverageThresholdBreak || s.Type == signals.SignalCoverageBlindSpot {
				hasCoverage = true
				break
			}
		}
		evidence := EvidenceStrong
		limitations := []string(nil)
		if !hasCoverage {
			evidence = EvidenceWeak
			limitations = []string{"No coverage data available; result may improve with coverage artifacts."}
		}
		return Result{
			ID: "coverage_depth.coverage_breach_share", Dimension: DimensionCoverageDepth,
			Value: 0, Units: UnitsRatio, Band: "strong",
			Evidence: evidence, Explanation: "No coverage threshold breaches detected.",
			Inputs: []string{"coverageThresholdBreak"}, Limitations: limitations,
		}
	}

	ratio := float64(count) / float64(total)
	band := ratioToBand(ratio, 0.05, 0.15, 0.30)

	return Result{
		ID: "coverage_depth.coverage_breach_share", Dimension: DimensionCoverageDepth,
		Value: ratio, Units: UnitsRatio, Band: band,
		Evidence:    EvidenceStrong,
		Explanation: fmt.Sprintf("%d coverage threshold breach(es) detected across %d test file(s) (%.0f%%).", count, total, ratio*100),
		Inputs:      []string{"coverageThresholdBreak"},
	}
}

func computeMockHeavyShare(snap *models.TestSuiteSnapshot) Result {
	total := len(snap.TestFiles)
	if total == 0 {
		return Result{
			ID: "coverage_diversity.mock_heavy_share", Dimension: DimensionCoverageDiversity,
			Value: 0, Units: UnitsRatio, Band: "strong",
			Evidence: EvidenceNone, Explanation: "No test files detected.",
		}
	}

	count := countSignals(snap, signals.SignalMockHeavyTest)
	ratio := float64(count) / float64(total)
	band := ratioToBand(ratio, 0.10, 0.25, 0.40)

	return Result{
		ID: "coverage_diversity.mock_heavy_share", Dimension: DimensionCoverageDiversity,
		Value: ratio, Units: UnitsRatio, Band: band,
		Evidence:    EvidenceStrong,
		Explanation: fmt.Sprintf("%d of %d test file(s) are mock-heavy (%.0f%%).", count, total, ratio*100),
		Inputs:      []string{"mockHeavyTest"},
	}
}

func computeFrameworkFragmentation(snap *models.TestSuiteSnapshot) Result {
	total := len(snap.TestFiles)
	fwCount := len(snap.Frameworks)

	if total == 0 || fwCount == 0 {
		return Result{
			ID: "coverage_diversity.framework_fragmentation", Dimension: DimensionCoverageDiversity,
			Value: 0, Units: UnitsRatio, Band: "strong",
			Evidence: EvidenceNone, Explanation: "No frameworks detected.",
		}
	}

	// Fragmentation: many frameworks relative to suite size.
	// 1 framework = no fragmentation, 5+ frameworks in a small suite = high.
	ratio := float64(fwCount) / float64(total)
	band := "strong"
	if fwCount >= 3 {
		band = "moderate"
	}
	if fwCount >= 5 || ratio > 0.3 {
		band = "weak"
	}

	return Result{
		ID: "coverage_diversity.framework_fragmentation", Dimension: DimensionCoverageDiversity,
		Value: ratio, Units: UnitsRatio, Band: band,
		Evidence:    EvidenceStrong,
		Explanation: fmt.Sprintf("%d framework(s) across %d test file(s).", fwCount, total),
		Inputs:      []string{"frameworks"},
	}
}

func computeE2EConcentration(snap *models.TestSuiteSnapshot) Result {
	total := len(snap.TestFiles)
	if total == 0 {
		return Result{
			ID: "coverage_diversity.e2e_concentration", Dimension: DimensionCoverageDiversity,
			Value: 0, Units: UnitsRatio, Band: "strong",
			Evidence: EvidenceNone, Explanation: "No test files detected.",
		}
	}

	e2eFrameworks := map[string]bool{}
	for _, fw := range snap.Frameworks {
		if fw.Type == models.FrameworkTypeE2E {
			e2eFrameworks[fw.Name] = true
		}
	}

	e2eCount := 0
	for _, tf := range snap.TestFiles {
		if e2eFrameworks[tf.Framework] {
			e2eCount++
		}
	}

	ratio := float64(e2eCount) / float64(total)
	band := "strong"
	if ratio > 0.50 {
		band = "moderate"
	}
	if ratio > 0.80 {
		band = "weak"
	}

	return Result{
		ID: "coverage_diversity.e2e_concentration", Dimension: DimensionCoverageDiversity,
		Value: ratio, Units: UnitsRatio, Band: band,
		Evidence:    EvidenceStrong,
		Explanation: fmt.Sprintf("%d of %d test file(s) use E2E frameworks (%.0f%%).", e2eCount, total, ratio*100),
		Inputs:      []string{"frameworks"},
	}
}

func computeE2EOnlyUnits(snap *models.TestSuiteSnapshot) Result {
	if snap.CoverageSummary == nil || snap.CoverageSummary.TotalCodeUnits == 0 {
		return Result{
			ID: "coverage_diversity.e2e_only_units", Dimension: DimensionCoverageDiversity,
			Value: 0, Units: UnitsRatio, Band: "unknown",
			Evidence:    EvidenceNone,
			Explanation: "No coverage data available.",
			Limitations: []string{"Provide labeled coverage artifacts (--coverage unit:path, --coverage e2e:path) for coverage-by-type analysis."},
		}
	}

	total := snap.CoverageSummary.TotalCodeUnits
	e2eOnly := snap.CoverageSummary.CoveredOnlyByE2E
	ratio := float64(e2eOnly) / float64(total)
	band := ratioToBand(ratio, 0.05, 0.15, 0.30)

	return Result{
		ID: "coverage_diversity.e2e_only_units", Dimension: DimensionCoverageDiversity,
		Value: ratio, Units: UnitsRatio, Band: band,
		Evidence:    EvidenceStrong,
		Explanation: fmt.Sprintf("%d of %d code unit(s) covered only by e2e tests (%.0f%%).", e2eOnly, total, ratio*100),
		Inputs:      []string{"coverageSummary"},
	}
}

func computeUnitTestCoverage(snap *models.TestSuiteSnapshot) Result {
	if snap.CoverageSummary == nil || snap.CoverageSummary.TotalCodeUnits == 0 {
		return Result{
			ID: "coverage_diversity.unit_test_coverage", Dimension: DimensionCoverageDiversity,
			Value: 0, Units: UnitsRatio, Band: "unknown",
			Evidence:    EvidenceNone,
			Explanation: "No coverage data available.",
			Limitations: []string{"Provide labeled coverage artifacts for coverage-by-type analysis."},
		}
	}

	total := snap.CoverageSummary.TotalCodeUnits
	covered := snap.CoverageSummary.CoveredByUnitTests
	ratio := float64(covered) / float64(total)

	// For unit test coverage, higher is better — invert the band logic.
	band := "strong"
	if ratio < 0.50 {
		band = "weak"
	} else if ratio < 0.70 {
		band = "moderate"
	}

	return Result{
		ID: "coverage_diversity.unit_test_coverage", Dimension: DimensionCoverageDiversity,
		Value: ratio, Units: UnitsRatio, Band: band,
		Evidence:    EvidenceStrong,
		Explanation: fmt.Sprintf("%d of %d code unit(s) covered by unit tests (%.0f%%).", covered, total, ratio*100),
		Inputs:      []string{"coverageSummary"},
	}
}
