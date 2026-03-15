package measurement

import (
	"fmt"

	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/signals"
)

// OperationalRiskMeasurements returns measurement definitions for the
// operational_risk dimension.
func OperationalRiskMeasurements() []Definition {
	return []Definition{
		{
			ID:          "operational_risk.policy_violation_density",
			Dimension:   DimensionOperationalRisk,
			Description: "Density of policy violations relative to test files.",
			Units:       UnitsRatio,
			Inputs:      []string{string(signals.SignalPolicyViolation)},
			Compute:     computePolicyViolationDensity,
		},
		{
			ID:          "operational_risk.legacy_framework_share",
			Dimension:   DimensionOperationalRisk,
			Description: "Share of test files using legacy frameworks.",
			Units:       UnitsRatio,
			Inputs:      []string{string(signals.SignalLegacyFrameworkUsage)},
			Compute:     computeLegacyFrameworkShare,
		},
		{
			ID:          "operational_risk.runtime_budget_breach_share",
			Dimension:   DimensionOperationalRisk,
			Description: "Share of test files exceeding runtime budgets.",
			Units:       UnitsRatio,
			Inputs:      []string{string(signals.SignalRuntimeBudgetExceeded)},
			Compute:     computeRuntimeBudgetBreachShare,
		},
	}
}

func computePolicyViolationDensity(snap *models.TestSuiteSnapshot) Result {
	total := len(snap.TestFiles)
	if total == 0 {
		return Result{
			ID: "operational_risk.policy_violation_density", Dimension: DimensionOperationalRisk,
			Value: 0, Units: UnitsRatio, Band: "strong",
			Evidence: EvidenceNone, Explanation: "No test files detected.",
		}
	}

	count := countSignals(snap, signals.SignalPolicyViolation)
	ratio := float64(count) / float64(total)
	band := ratioToBand(ratio, 0.0, 0.05, 0.15)

	return Result{
		ID: "operational_risk.policy_violation_density", Dimension: DimensionOperationalRisk,
		Value: ratio, Units: UnitsRatio, Band: band,
		Evidence:    EvidenceStrong,
		Explanation: fmt.Sprintf("%d policy violation(s) across %d test file(s) (%.0f%%).", count, total, ratio*100),
		Inputs:      []string{"policyViolation"},
	}
}

func computeLegacyFrameworkShare(snap *models.TestSuiteSnapshot) Result {
	total := len(snap.TestFiles)
	if total == 0 {
		return Result{
			ID: "operational_risk.legacy_framework_share", Dimension: DimensionOperationalRisk,
			Value: 0, Units: UnitsRatio, Band: "strong",
			Evidence: EvidenceNone, Explanation: "No test files detected.",
		}
	}

	count := countSignals(snap, signals.SignalLegacyFrameworkUsage)
	ratio := float64(count) / float64(total)
	band := ratioToBand(ratio, 0.05, 0.15, 0.30)

	return Result{
		ID: "operational_risk.legacy_framework_share", Dimension: DimensionOperationalRisk,
		Value: ratio, Units: UnitsRatio, Band: band,
		Evidence:    EvidenceStrong,
		Explanation: fmt.Sprintf("%d of %d test file(s) use legacy frameworks (%.0f%%).", count, total, ratio*100),
		Inputs:      []string{"legacyFrameworkUsage"},
	}
}

func computeRuntimeBudgetBreachShare(snap *models.TestSuiteSnapshot) Result {
	total := len(snap.TestFiles)
	if total == 0 {
		return Result{
			ID: "operational_risk.runtime_budget_breach_share", Dimension: DimensionOperationalRisk,
			Value: 0, Units: UnitsRatio, Band: "strong",
			Evidence: EvidenceNone, Explanation: "No test files detected.",
		}
	}

	count := countSignals(snap, signals.SignalRuntimeBudgetExceeded)
	ratio := float64(count) / float64(total)
	band := ratioToBand(ratio, 0.05, 0.15, 0.30)
	evidence := runtimeEvidence(snap)

	return Result{
		ID: "operational_risk.runtime_budget_breach_share", Dimension: DimensionOperationalRisk,
		Value: ratio, Units: UnitsRatio, Band: band,
		Evidence:    evidence,
		Explanation: fmt.Sprintf("%d of %d test file(s) exceed runtime budgets (%.0f%%).", count, total, ratio*100),
		Inputs:      []string{"runtimeBudgetExceeded"},
		Limitations: evidenceLimitations(evidence),
	}
}
