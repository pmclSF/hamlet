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

	// Without runtime data and no signals, flakiness cannot be assessed —
	// report unknown rather than a false "strong".
	if evidence == EvidenceWeak && count == 0 {
		return Result{
			ID: "health.flaky_share", Dimension: DimensionHealth,
			Value: 0, Units: UnitsRatio, Band: "unknown",
			Evidence:    evidence,
			Explanation: "No runtime data available; flakiness cannot be assessed from static analysis alone.",
			Inputs:      []string{"flakyTest", "unstableSuite"},
			Limitations: evidenceLimitations(evidence),
		}
	}

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

	fileSet := map[string]bool{}
	for _, s := range snap.Signals {
		if s.Type != signals.SignalSkippedTest {
			continue
		}
		if s.Location.File != "" {
			fileSet[s.Location.File] = true
		}
	}
	count := len(fileSet)
	if count == 0 {
		// Backward compatibility for snapshots that only contain repo-level skipped signals.
		count = countSignals(snap, signals.SignalSkippedTest)
	}
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

	// Without runtime data and no signals, slowness cannot be assessed —
	// report unknown rather than a false "strong".
	if evidence == EvidenceWeak && count == 0 {
		return Result{
			ID: "health.slow_test_share", Dimension: DimensionHealth,
			Value: 0, Units: UnitsRatio, Band: "unknown",
			Evidence:    evidence,
			Explanation: "No runtime data available; test speed cannot be assessed from static analysis alone.",
			Inputs:      []string{"slowTest"},
			Limitations: evidenceLimitations(evidence),
		}
	}

	return Result{
		ID: "health.slow_test_share", Dimension: DimensionHealth,
		Value: ratio, Units: UnitsRatio, Band: band,
		Evidence:    evidence,
		Explanation: fmt.Sprintf("%d of %d test file(s) flagged as slow (%.0f%%).", count, total, ratio*100),
		Inputs:      []string{"slowTest"},
		Limitations: evidenceLimitations(evidence),
	}
}
