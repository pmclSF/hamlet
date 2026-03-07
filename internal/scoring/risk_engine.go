// Package scoring implements Hamlet's explainable risk engine.
//
// Risk surfaces are derived from concrete signals. Every risk surface
// must be explainable, transparent, and actionable.
//
// The risk model is intentionally simple:
//   - risk is computed from signal severity weights
//   - weights are explicit and inspectable
//   - bands are computed from normalized scores
//   - contributing signals are listed for transparency
package scoring

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
)

// Severity weights for risk score computation.
// These are explicit so future stages can tune them transparently.
var severityWeight = map[models.SignalSeverity]float64{
	models.SeverityCritical: 4.0,
	models.SeverityHigh:     3.0,
	models.SeverityMedium:   2.0,
	models.SeverityLow:      1.0,
	models.SeverityInfo:     0.5,
}

// Signal types that feed each risk dimension.
var reliabilitySignals = map[models.SignalType]bool{
	"flakyTest":     true,
	"skippedTest":   true,
	"deadTest":      true,
	"unstableSuite": true,
	"slowTest":      true,
}

var changeRiskSignals = map[models.SignalType]bool{
	"weakAssertion":          true,
	"untestedExport":         true,
	"mockHeavyTest":          true,
	"testsOnlyMocks":         true,
	"coverageBlindSpot":      true,
	"coverageThresholdBreak": true,
	"migrationBlocker":          true,
	"deprecatedTestPattern":    true,
	"dynamicTestGeneration":    true,
	"customMatcherRisk":        true,
	"unsupportedSetup":         true,
}

var speedSignals = map[models.SignalType]bool{
	"slowTest":              true,
	"runtimeBudgetExceeded": true,
}

// ComputeRisk generates risk surfaces from the signals in the snapshot.
//
// Currently computes:
//   - repository-level reliability, change, and speed risk
//   - directory-level change risk rollups
func ComputeRisk(snap *models.TestSuiteSnapshot) []models.RiskSurface {
	var surfaces []models.RiskSurface

	// Repository-level risk
	surfaces = append(surfaces, computeRepoRisk(snap.Signals, "reliability", reliabilitySignals)...)
	surfaces = append(surfaces, computeRepoRisk(snap.Signals, "change", changeRiskSignals)...)
	surfaces = append(surfaces, computeRepoRisk(snap.Signals, "speed", speedSignals)...)

	// Directory-level change risk rollups
	surfaces = append(surfaces, computeDirectoryRisk(snap.Signals, "change", changeRiskSignals)...)

	return surfaces
}

// computeRepoRisk computes a single risk dimension at repo scope.
func computeRepoRisk(signals []models.Signal, riskType string, relevant map[models.SignalType]bool) []models.RiskSurface {
	var contributing []models.Signal
	var totalWeight float64

	for _, s := range signals {
		if relevant[s.Type] {
			contributing = append(contributing, s)
			totalWeight += severityWeight[s.Severity]
		}
	}

	if len(contributing) == 0 {
		return nil
	}

	band := scoreToBand(totalWeight)

	return []models.RiskSurface{{
		Type:                riskType,
		Scope:               "repository",
		ScopeName:           "repo",
		Band:                band,
		Score:               totalWeight,
		ContributingSignals: contributing,
		Explanation:         buildExplanation(riskType, band, contributing),
		SuggestedAction:     buildSuggestedAction(riskType, band),
	}}
}

// computeDirectoryRisk groups signals by directory and computes per-directory risk.
func computeDirectoryRisk(signals []models.Signal, riskType string, relevant map[models.SignalType]bool) []models.RiskSurface {
	dirSignals := map[string][]models.Signal{}
	dirWeights := map[string]float64{}

	for _, s := range signals {
		if !relevant[s.Type] {
			continue
		}
		dir := filepath.Dir(s.Location.File)
		if dir == "" || dir == "." {
			continue
		}
		dirSignals[dir] = append(dirSignals[dir], s)
		dirWeights[dir] += severityWeight[s.Severity]
	}

	// Sort directory keys for deterministic output.
	dirs := make([]string, 0, len(dirSignals))
	for dir := range dirSignals {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	var surfaces []models.RiskSurface
	for _, dir := range dirs {
		sigs := dirSignals[dir]
		if len(sigs) < 2 {
			continue // Only flag directories with multiple signals
		}
		band := scoreToBand(dirWeights[dir])
		surfaces = append(surfaces, models.RiskSurface{
			Type:                riskType,
			Scope:               "directory",
			ScopeName:           dir,
			Band:                band,
			Score:               dirWeights[dir],
			ContributingSignals: sigs,
			Explanation: fmt.Sprintf("%s risk in %s: %d signals with combined weight %.1f.",
				strings.Title(riskType), dir, len(sigs), dirWeights[dir]),
			SuggestedAction: fmt.Sprintf("Review %s for concentrated test quality issues.", dir),
		})
	}

	// Sort by score descending, then by name for deterministic output.
	sort.Slice(surfaces, func(i, j int) bool {
		if surfaces[i].Score != surfaces[j].Score {
			return surfaces[i].Score > surfaces[j].Score
		}
		return surfaces[i].ScopeName < surfaces[j].ScopeName
	})

	return surfaces
}

// scoreToBand maps a weighted signal score to a qualitative risk band.
//
// Thresholds:
//   - 0-3:  low
//   - 4-8:  medium
//   - 9-15: high
//   - 16+:  critical
//
// These thresholds are intentionally simple and inspectable.
func scoreToBand(score float64) models.RiskBand {
	switch {
	case score >= 16:
		return models.RiskBandCritical
	case score >= 9:
		return models.RiskBandHigh
	case score >= 4:
		return models.RiskBandMedium
	default:
		return models.RiskBandLow
	}
}

func buildExplanation(riskType string, band models.RiskBand, signals []models.Signal) string {
	// Count by type for a useful explanation
	typeCounts := map[models.SignalType]int{}
	for _, s := range signals {
		typeCounts[s.Type]++
	}

	parts := make([]string, 0, len(typeCounts))
	for t, c := range typeCounts {
		parts = append(parts, fmt.Sprintf("%d %s", c, t))
	}
	sort.Strings(parts)

	return fmt.Sprintf("%s risk is %s based on %d signals: %s.",
		strings.Title(riskType), band, len(signals), strings.Join(parts, ", "))
}

func buildSuggestedAction(riskType string, band models.RiskBand) string {
	switch riskType {
	case "reliability":
		return "Investigate flaky, skipped, or dead tests to improve suite reliability."
	case "change":
		if band == models.RiskBandHigh || band == models.RiskBandCritical {
			return "Prioritize adding assertions and test coverage for high-risk areas before making changes."
		}
		return "Improve test coverage and assertion quality to reduce change risk."
	case "speed":
		return "Identify and optimize slow tests to maintain fast feedback loops."
	default:
		return "Review contributing signals and address highest-severity items first."
	}
}
