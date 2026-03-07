package measurement

import (
	"fmt"

	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/signals"
)

// HealthMeasurements returns measurement definitions for the health dimension.
func HealthMeasurements() []Definition {
	return []Definition{
		{
			ID:          "health.flaky_share",
			Dimension:   DimensionHealth,
			Description: "Share of test files flagged as flaky or unstable.",
			Units:       UnitsRatio,
			Inputs:      []string{string(signals.SignalFlakyTest), string(signals.SignalUnstableSuite)},
			Compute:     computeFlakyShare,
		},
		{
			ID:          "health.skip_density",
			Dimension:   DimensionHealth,
			Description: "Share of test files with skipped tests.",
			Units:       UnitsRatio,
			Inputs:      []string{string(signals.SignalSkippedTest)},
			Compute:     computeSkipDensity,
		},
		{
			ID:          "health.dead_test_share",
			Dimension:   DimensionHealth,
			Description: "Share of test files containing dead or unreachable tests.",
			Units:       UnitsRatio,
			Inputs:      []string{string(signals.SignalDeadTest)},
			Compute:     computeDeadTestShare,
		},
		{
			ID:          "health.slow_test_share",
			Dimension:   DimensionHealth,
			Description: "Share of test files flagged as slow.",
			Units:       UnitsRatio,
			Inputs:      []string{string(signals.SignalSlowTest)},
			Compute:     computeSlowTestShare,
		},
	}
}

func computeFlakyShare(snap *models.TestSuiteSnapshot) Result {
	total := len(snap.TestFiles)
	if total == 0 {
		return Result{
			ID: "health.flaky_share", Dimension: DimensionHealth,
			Value: 0, Units: UnitsRatio, Band: "strong",
			Evidence: EvidenceNone, Explanation: "No test files detected.",
		}
	}

	count := countSignals(snap, signals.SignalFlakyTest, signals.SignalUnstableSuite)
	ratio := float64(count) / float64(total)
	band := ratioToBand(ratio, 0.05, 0.15, 0.30)
	evidence := runtimeEvidence(snap)

	return Result{
		ID: "health.flaky_share", Dimension: DimensionHealth,
		Value: ratio, Units: UnitsRatio, Band: band,
		Evidence:    evidence,
		Explanation: fmt.Sprintf("%d of %d test file(s) flagged as flaky or unstable (%.0f%%).", count, total, ratio*100),
		Inputs:      []string{"flakyTest", "unstableSuite"},
		Limitations: evidenceLimitations(evidence),
	}
}

func computeSkipDensity(snap *models.TestSuiteSnapshot) Result {
	total := len(snap.TestFiles)
	if total == 0 {
		return Result{
			ID: "health.skip_density", Dimension: DimensionHealth,
			Value: 0, Units: UnitsRatio, Band: "strong",
			Evidence: EvidenceNone, Explanation: "No test files detected.",
		}
	}

	count := countSignals(snap, signals.SignalSkippedTest)
	ratio := float64(count) / float64(total)
	band := ratioToBand(ratio, 0.05, 0.15, 0.30)

	return Result{
		ID: "health.skip_density", Dimension: DimensionHealth,
		Value: ratio, Units: UnitsRatio, Band: band,
		Evidence:    EvidenceStrong,
		Explanation: fmt.Sprintf("%d of %d test file(s) contain skipped tests (%.0f%%).", count, total, ratio*100),
		Inputs:      []string{"skippedTest"},
	}
}

func computeDeadTestShare(snap *models.TestSuiteSnapshot) Result {
	total := len(snap.TestFiles)
	if total == 0 {
		return Result{
			ID: "health.dead_test_share", Dimension: DimensionHealth,
			Value: 0, Units: UnitsRatio, Band: "strong",
			Evidence: EvidenceNone, Explanation: "No test files detected.",
		}
	}

	count := countSignals(snap, signals.SignalDeadTest)
	ratio := float64(count) / float64(total)
	band := ratioToBand(ratio, 0.02, 0.10, 0.20)

	return Result{
		ID: "health.dead_test_share", Dimension: DimensionHealth,
		Value: ratio, Units: UnitsRatio, Band: band,
		Evidence:    EvidenceStrong,
		Explanation: fmt.Sprintf("%d of %d test file(s) contain dead tests (%.0f%%).", count, total, ratio*100),
		Inputs:      []string{"deadTest"},
	}
}

func computeSlowTestShare(snap *models.TestSuiteSnapshot) Result {
	total := len(snap.TestFiles)
	if total == 0 {
		return Result{
			ID: "health.slow_test_share", Dimension: DimensionHealth,
			Value: 0, Units: UnitsRatio, Band: "strong",
			Evidence: EvidenceNone, Explanation: "No test files detected.",
		}
	}

	count := countSignals(snap, signals.SignalSlowTest)
	ratio := float64(count) / float64(total)
	band := ratioToBand(ratio, 0.10, 0.25, 0.50)
	evidence := runtimeEvidence(snap)

	return Result{
		ID: "health.slow_test_share", Dimension: DimensionHealth,
		Value: ratio, Units: UnitsRatio, Band: band,
		Evidence:    evidence,
		Explanation: fmt.Sprintf("%d of %d test file(s) flagged as slow (%.0f%%).", count, total, ratio*100),
		Inputs:      []string{"slowTest"},
		Limitations: evidenceLimitations(evidence),
	}
}
