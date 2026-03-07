package measurement

import (
	"fmt"

	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/signals"
)

// StructuralRiskMeasurements returns measurement definitions for the
// structural_risk dimension.
func StructuralRiskMeasurements() []Definition {
	return []Definition{
		{
			ID:          "structural_risk.migration_blocker_density",
			Dimension:   DimensionStructuralRisk,
			Description: "Density of migration blockers relative to test files.",
			Units:       UnitsRatio,
			Inputs: []string{
				string(signals.SignalMigrationBlocker),
				string(signals.SignalDeprecatedTestPattern),
				string(signals.SignalDynamicTestGeneration),
				string(signals.SignalCustomMatcherRisk),
			},
			Compute: computeMigrationBlockerDensity,
		},
		{
			ID:          "structural_risk.deprecated_pattern_share",
			Dimension:   DimensionStructuralRisk,
			Description: "Share of test files using deprecated patterns.",
			Units:       UnitsRatio,
			Inputs:      []string{string(signals.SignalDeprecatedTestPattern)},
			Compute:     computeDeprecatedPatternShare,
		},
		{
			ID:          "structural_risk.dynamic_generation_share",
			Dimension:   DimensionStructuralRisk,
			Description: "Share of test files using dynamic test generation.",
			Units:       UnitsRatio,
			Inputs:      []string{string(signals.SignalDynamicTestGeneration)},
			Compute:     computeDynamicGenerationShare,
		},
	}
}

func computeMigrationBlockerDensity(snap *models.TestSuiteSnapshot) Result {
	total := len(snap.TestFiles)
	if total == 0 {
		return Result{
			ID: "structural_risk.migration_blocker_density", Dimension: DimensionStructuralRisk,
			Value: 0, Units: UnitsRatio, Band: "strong",
			Evidence: EvidenceNone, Explanation: "No test files detected.",
		}
	}

	count := countSignals(snap,
		signals.SignalMigrationBlocker,
		signals.SignalDeprecatedTestPattern,
		signals.SignalDynamicTestGeneration,
		signals.SignalCustomMatcherRisk,
	)
	ratio := float64(count) / float64(total)
	band := ratioToBand(ratio, 0.05, 0.15, 0.30)

	return Result{
		ID: "structural_risk.migration_blocker_density", Dimension: DimensionStructuralRisk,
		Value: ratio, Units: UnitsRatio, Band: band,
		Evidence:    EvidenceStrong,
		Explanation: fmt.Sprintf("%d migration blocker(s) across %d test file(s) (%.0f%%).", count, total, ratio*100),
		Inputs:      []string{"migrationBlocker", "deprecatedTestPattern", "dynamicTestGeneration", "customMatcherRisk"},
	}
}

func computeDeprecatedPatternShare(snap *models.TestSuiteSnapshot) Result {
	total := len(snap.TestFiles)
	if total == 0 {
		return Result{
			ID: "structural_risk.deprecated_pattern_share", Dimension: DimensionStructuralRisk,
			Value: 0, Units: UnitsRatio, Band: "strong",
			Evidence: EvidenceNone, Explanation: "No test files detected.",
		}
	}

	count := countSignals(snap, signals.SignalDeprecatedTestPattern)
	ratio := float64(count) / float64(total)
	band := ratioToBand(ratio, 0.05, 0.15, 0.30)

	return Result{
		ID: "structural_risk.deprecated_pattern_share", Dimension: DimensionStructuralRisk,
		Value: ratio, Units: UnitsRatio, Band: band,
		Evidence:    EvidenceStrong,
		Explanation: fmt.Sprintf("%d of %d test file(s) use deprecated patterns (%.0f%%).", count, total, ratio*100),
		Inputs:      []string{"deprecatedTestPattern"},
	}
}

func computeDynamicGenerationShare(snap *models.TestSuiteSnapshot) Result {
	total := len(snap.TestFiles)
	if total == 0 {
		return Result{
			ID: "structural_risk.dynamic_generation_share", Dimension: DimensionStructuralRisk,
			Value: 0, Units: UnitsRatio, Band: "strong",
			Evidence: EvidenceNone, Explanation: "No test files detected.",
		}
	}

	count := countSignals(snap, signals.SignalDynamicTestGeneration)
	ratio := float64(count) / float64(total)
	band := ratioToBand(ratio, 0.05, 0.10, 0.20)

	return Result{
		ID: "structural_risk.dynamic_generation_share", Dimension: DimensionStructuralRisk,
		Value: ratio, Units: UnitsRatio, Band: band,
		Evidence:    EvidenceStrong,
		Explanation: fmt.Sprintf("%d of %d test file(s) use dynamic test generation (%.0f%%).", count, total, ratio*100),
		Inputs:      []string{"dynamicTestGeneration"},
	}
}
