// Package scoring implements Terrain's explainable risk engine.
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
	"math"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// RiskModelVersion increments when scoring methodology changes.
const RiskModelVersion = "2.0.0"

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
	"migrationBlocker":       true,
	"deprecatedTestPattern":  true,
	"dynamicTestGeneration":  true,
	"customMatcherRisk":      true,
	"unsupportedSetup":       true,
}

var speedSignals = map[models.SignalType]bool{
	"slowTest":              true,
	"runtimeBudgetExceeded": true,
}

var governanceSignals = map[models.SignalType]bool{
	"policyViolation":       true,
	"legacyFrameworkUsage":  true,
	"skippedTestsInCI":      true,
	"runtimeBudgetExceeded": true,
}

// ComputeRisk generates risk surfaces from the signals in the snapshot.
//
// Risk uses a hybrid score:
//   - density score (weighted issues per 10 files)
//   - absolute burden score (log-scaled weight and count)
//
// The final score is the maximum of those two signals, which preserves
// local concentration while avoiding under-reporting severe absolute burden
// in very large repositories.
//
// Currently computes:
//   - repository-level reliability, change, speed, and governance risk
//   - directory-level change risk rollups
func ComputeRisk(snap *models.TestSuiteSnapshot) []models.RiskSurface {
	var surfaces []models.RiskSurface

	totalFiles := len(snap.TestFiles)
	previousRepoBands := previousRepositoryBands(snap.Risk)

	// Repository-level risk
	surfaces = append(surfaces, computeRepoRisk(
		snap.Signals, "reliability", reliabilitySignals, totalFiles, previousRepoBands["reliability"],
	)...)
	surfaces = append(surfaces, computeRepoRisk(
		snap.Signals, "change", changeRiskSignals, totalFiles, previousRepoBands["change"],
	)...)
	surfaces = append(surfaces, computeRepoRisk(
		snap.Signals, "speed", speedSignals, totalFiles, previousRepoBands["speed"],
	)...)
	surfaces = append(surfaces, computeRepoRisk(
		snap.Signals, "governance", governanceSignals, totalFiles, previousRepoBands["governance"],
	)...)

	// Directory-level change risk rollups
	surfaces = append(surfaces, computeDirectoryRisk(snap)...)

	return surfaces
}

// computeRepoRisk computes a single risk dimension at repo scope.
// totalFiles is used for density normalization; absolute burden is always
// considered to avoid masking severe issues in large suites.
func computeRepoRisk(
	signals []models.Signal,
	riskType string,
	relevant map[models.SignalType]bool,
	totalFiles int,
	previousBand models.RiskBand,
) []models.RiskSurface {
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

	score := computeHybridScore(totalWeight, len(contributing), totalFiles)
	if riskType == "governance" && score < 4 && hasGovernanceFloorTrigger(contributing) {
		score = 4
	}
	band := scoreToBandWithHysteresis(score, previousBand)

	return []models.RiskSurface{{
		Type:                riskType,
		Scope:               "repository",
		ScopeName:           "repo",
		Band:                band,
		Score:               score,
		ContributingSignals: contributing,
		Explanation:         buildExplanation(riskType, band, contributing, totalFiles, totalWeight, score),
		SuggestedAction:     buildSuggestedAction(riskType, band),
	}}
}

func previousRepositoryBands(risk []models.RiskSurface) map[string]models.RiskBand {
	byType := map[string]models.RiskBand{}
	for _, r := range risk {
		if r.Scope != "repository" || r.Type == "" {
			continue
		}
		byType[r.Type] = r.Band
	}
	return byType
}

func hasGovernanceFloorTrigger(signals []models.Signal) bool {
	for _, s := range signals {
		if s.Type == "policyViolation" || severityRank(s.Severity) >= severityRank(models.SeverityHigh) {
			return true
		}
	}
	return false
}

func severityRank(sev models.SignalSeverity) int {
	switch sev {
	case models.SeverityCritical:
		return 5
	case models.SeverityHigh:
		return 4
	case models.SeverityMedium:
		return 3
	case models.SeverityLow:
		return 2
	case models.SeverityInfo:
		return 1
	default:
		return 0
	}
}

// computeDirectoryRisk groups signals by directory and computes per-directory risk.
// Uses test file count per directory for density normalization.
func computeDirectoryRisk(snap *models.TestSuiteSnapshot) []models.RiskSurface {
	riskType := "change"
	relevant := changeRiskSignals

	dirSignals := map[string][]models.Signal{}
	dirWeights := map[string]float64{}

	for _, s := range snap.Signals {
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

	// Count test files per directory for normalization.
	dirFileCount := map[string]int{}
	for _, tf := range snap.TestFiles {
		dir := filepath.Dir(tf.Path)
		dirFileCount[dir]++
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
		// Flag directories with multiple signals, or single signals at
		// high/critical severity — a lone high-severity issue in a
		// directory still represents concentrated risk worth surfacing.
		if len(sigs) < 2 && !hasHighSeveritySignal(sigs) {
			continue
		}
		// Normalize by directory file count.
		fileCount := dirFileCount[dir]
		normalizedScore := dirWeights[dir]
		if fileCount > 0 {
			normalizedScore = (dirWeights[dir] / float64(fileCount)) * 10.0
		}
		band := scoreToBand(normalizedScore)
		surfaces = append(surfaces, models.RiskSurface{
			Type:                riskType,
			Scope:               "directory",
			ScopeName:           dir,
			Band:                band,
			Score:               normalizedScore,
			ContributingSignals: sigs,
			Explanation: fmt.Sprintf("%s risk in %s: %d signals across %d test file(s).",
				titleRiskType(riskType), dir, len(sigs), fileCount),
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

func scoreToBandWithHysteresis(score float64, previousBand models.RiskBand) models.RiskBand {
	if previousBand == "" {
		return scoreToBand(score)
	}

	// Deadband around thresholds to reduce band flapping near boundaries.
	const hysteresis = 0.5
	lowUp := 4.0 + hysteresis
	mediumDown := 4.0 - hysteresis
	mediumUp := 9.0 + hysteresis
	highDown := 9.0 - hysteresis
	highUp := 16.0 + hysteresis
	criticalDown := 16.0 - hysteresis

	switch previousBand {
	case models.RiskBandLow:
		switch {
		case score >= highUp:
			return models.RiskBandCritical
		case score >= mediumUp:
			return models.RiskBandHigh
		case score >= lowUp:
			return models.RiskBandMedium
		default:
			return models.RiskBandLow
		}
	case models.RiskBandMedium:
		switch {
		case score >= highUp:
			return models.RiskBandCritical
		case score >= mediumUp:
			return models.RiskBandHigh
		case score < mediumDown:
			return models.RiskBandLow
		default:
			return models.RiskBandMedium
		}
	case models.RiskBandHigh:
		switch {
		case score >= highUp:
			return models.RiskBandCritical
		case score < highDown:
			if score < mediumDown {
				return models.RiskBandLow
			}
			return models.RiskBandMedium
		default:
			return models.RiskBandHigh
		}
	case models.RiskBandCritical:
		if score < criticalDown {
			if score < highDown {
				if score < mediumDown {
					return models.RiskBandLow
				}
				return models.RiskBandMedium
			}
			return models.RiskBandHigh
		}
		return models.RiskBandCritical
	default:
		return scoreToBand(score)
	}
}

func computeHybridScore(totalWeight float64, signalCount, totalFiles int) float64 {
	densityScore := totalWeight
	if totalFiles > 0 {
		densityScore = (totalWeight / float64(totalFiles)) * 10.0
	}

	// Absolute burden: log-scaled so large repos are comparable while still
	// surfacing substantial issue volume even at low density.
	absoluteScore := (math.Log1p(totalWeight) * 1.2) + (math.Log1p(float64(signalCount)) * 0.8)
	if densityScore > absoluteScore {
		return densityScore
	}
	return absoluteScore
}

func buildExplanation(riskType string, band models.RiskBand, signals []models.Signal, totalFiles int, totalWeight, score float64) string {
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

	density := "n/a"
	if totalFiles > 0 {
		densityScore := (totalWeight / float64(totalFiles)) * 10.0
		density = fmt.Sprintf("%.2f/10 across %d test files", densityScore, totalFiles)
	}
	absoluteScore := (math.Log1p(totalWeight) * 1.2) + (math.Log1p(float64(len(signals))) * 0.8)

	return fmt.Sprintf("%s risk is %s (score %.2f) from %d signal(s): density=%s, absolute=%.2f, types=%s.",
		titleRiskType(riskType), band, score, len(signals), density, absoluteScore, strings.Join(parts, ", "))
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
	case "governance":
		return "Address policy violations and governance findings before expanding test investments."
	default:
		return "Review contributing signals and address highest-severity items first."
	}
}

func titleRiskType(riskType string) string {
	if riskType == "" {
		return riskType
	}
	return strings.ToUpper(riskType[:1]) + riskType[1:]
}

func hasHighSeveritySignal(signals []models.Signal) bool {
	for _, s := range signals {
		if s.Severity == models.SeverityHigh || s.Severity == models.SeverityCritical {
			return true
		}
	}
	return false
}
