// Package metrics extracts normalized, benchmark-ready aggregate metrics
// from a TestSuiteSnapshot.
//
// These metrics are designed to be:
//   - aggregate (counts and ratios, not raw file lists)
//   - explainable (clear derivation from snapshot data)
//   - privacy-conscious (no raw source, file paths, or symbol names)
//   - locally useful as a repo health scorecard
//   - future-safe for hosted benchmarking aggregation
//
// Privacy boundary:
//   The metrics artifact intentionally excludes raw file paths, symbol names,
//   source code snippets, and user identity information. It contains only
//   aggregate counts, ratios, and qualitative bands. This makes it safe
//   for future anonymous aggregation without exposing proprietary code.
package metrics

import (
	"time"

	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/signals"
)

// Snapshot contains benchmark-ready aggregate metrics derived from
// a TestSuiteSnapshot.
//
// This is intentionally separate from the rich local snapshot to
// maintain a clear privacy boundary between local analysis data
// and data suitable for future aggregation.
type Snapshot struct {
	// GeneratedAt is when this metrics snapshot was created.
	GeneratedAt time.Time `json:"generatedAt"`

	// AnalysisVersion identifies the Hamlet version/stage.
	AnalysisVersion string `json:"analysisVersion"`

	Structure  StructureMetrics  `json:"structure"`
	Health     HealthMetrics     `json:"health"`
	Quality    QualityMetrics    `json:"quality"`
	Change     ChangeMetrics     `json:"changeReadiness"`
	Governance GovernanceMetrics `json:"governance"`
	Risk       RiskMetrics       `json:"risk"`

	// Notes describes missing inputs or limitations.
	Notes []string `json:"notes,omitempty"`
}

// StructureMetrics captures the shape of the test ecosystem.
type StructureMetrics struct {
	TotalTestFiles             int      `json:"totalTestFiles"`
	Frameworks                 []string `json:"frameworks"`
	FrameworkCount             int      `json:"frameworkCount"`
	FrameworkFragmentationRatio float64 `json:"frameworkFragmentationRatio"`
	Languages                  []string `json:"languages"`
}

// HealthMetrics captures reliability and runtime behavior.
type HealthMetrics struct {
	SlowTestCount    int     `json:"slowTestCount"`
	SlowTestRatio    float64 `json:"slowTestRatio"`
	FlakyTestCount   int     `json:"flakyTestCount"`
	FlakyTestRatio   float64 `json:"flakyTestRatio"`
	SkippedTestCount int     `json:"skippedTestCount"`
	SkippedTestRatio float64 `json:"skippedTestRatio"`
	DeadTestCount    int     `json:"deadTestCount"`
}

// QualityMetrics captures test quality characteristics.
type QualityMetrics struct {
	WeakAssertionCount         int     `json:"weakAssertionCount"`
	WeakAssertionRatio         float64 `json:"weakAssertionRatio"`
	MockHeavyTestCount         int     `json:"mockHeavyTestCount"`
	MockHeavyTestRatio         float64 `json:"mockHeavyTestRatio"`
	UntestedExportCount        int     `json:"untestedExportCount"`
	CoverageThresholdBreakCount int    `json:"coverageThresholdBreakCount"`
	SnapshotHeavyCount         int     `json:"snapshotHeavyCount"`

	// QualityPostureBand summarizes overall test quality as a band.
	// Derived from the proportion of files with quality issues.
	// Values: "strong", "moderate", "weak".
	QualityPostureBand string `json:"qualityPostureBand"`
}

// ChangeMetrics captures migration/modernization readiness.
type ChangeMetrics struct {
	MigrationBlockerCount     int            `json:"migrationBlockerCount"`
	DeprecatedPatternCount    int            `json:"deprecatedPatternCount"`
	DynamicGenerationCount    int            `json:"dynamicGenerationCount"`
	CustomMatcherRiskCount    int            `json:"customMatcherRiskCount"`
	BlockerCountByType        map[string]int `json:"blockerCountByType,omitempty"`

	// MigrationReadinessBand is the aggregate migration readiness level.
	// Values: "high", "medium", "low", "unknown".
	MigrationReadinessBand string `json:"migrationReadinessBand"`

	// SafeAreaCount is the number of directories classified as safe to modernize.
	SafeAreaCount int `json:"safeAreaCount"`

	// RiskyAreaCount is the number of directories with compounded migration
	// and quality risk.
	RiskyAreaCount int `json:"riskyAreaCount"`

	// QualityCompoundedBlockerCount is the number of migration blockers
	// in files that also have quality issues (weak assertions, mock-heavy, etc.).
	QualityCompoundedBlockerCount int `json:"qualityCompoundedBlockerCount"`
}

// GovernanceMetrics captures policy-related findings.
type GovernanceMetrics struct {
	PolicyViolationCount       int `json:"policyViolationCount"`
	LegacyFrameworkUsageCount  int `json:"legacyFrameworkUsageCount"`
	RuntimeBudgetExceededCount int `json:"runtimeBudgetExceededCount"`
}

// RiskMetrics captures the risk posture.
type RiskMetrics struct {
	ReliabilityBand     string `json:"reliabilityBand,omitempty"`
	ChangeBand          string `json:"changeBand,omitempty"`
	SpeedBand           string `json:"speedBand,omitempty"`
	HighRiskAreaCount   int    `json:"highRiskAreaCount"`
	CriticalFindingCount int   `json:"criticalFindingCount"`
}

// Derive computes a metrics Snapshot from a TestSuiteSnapshot.
//
// Each metric has a clear derivation:
//   - counts: number of signals of a given type
//   - ratios: signal count / total test files
//   - bands: from risk surfaces
//   - fragmentation: framework count / total test files
func Derive(snap *models.TestSuiteSnapshot) *Snapshot {
	totalFiles := len(snap.TestFiles)
	signalCounts := countSignalsByType(snap.Signals)

	ms := &Snapshot{
		GeneratedAt:     time.Now().UTC(),
		AnalysisVersion: "v3-nucleus",
	}

	// Structure
	frameworks := make([]string, len(snap.Frameworks))
	for i, fw := range snap.Frameworks {
		frameworks[i] = fw.Name
	}
	ms.Structure = StructureMetrics{
		TotalTestFiles:             totalFiles,
		Frameworks:                 frameworks,
		FrameworkCount:             len(snap.Frameworks),
		FrameworkFragmentationRatio: safeRatio(len(snap.Frameworks), totalFiles),
		Languages:                  snap.Repository.Languages,
	}

	// Health
	ms.Health = HealthMetrics{
		SlowTestCount:    signalCounts["slowTest"],
		SlowTestRatio:    safeRatio(signalCounts["slowTest"], totalFiles),
		FlakyTestCount:   signalCounts["flakyTest"],
		FlakyTestRatio:   safeRatio(signalCounts["flakyTest"], totalFiles),
		SkippedTestCount: signalCounts["skippedTest"],
		SkippedTestRatio: safeRatio(signalCounts["skippedTest"], totalFiles),
		DeadTestCount:    signalCounts["deadTest"],
	}

	// Quality
	qualityIssueCount := signalCounts["weakAssertion"] + signalCounts["mockHeavyTest"] +
		signalCounts["untestedExport"] + signalCounts["coverageThresholdBreak"]
	ms.Quality = QualityMetrics{
		WeakAssertionCount:         signalCounts["weakAssertion"],
		WeakAssertionRatio:         safeRatio(signalCounts["weakAssertion"], totalFiles),
		MockHeavyTestCount:         signalCounts["mockHeavyTest"],
		MockHeavyTestRatio:         safeRatio(signalCounts["mockHeavyTest"], totalFiles),
		UntestedExportCount:        signalCounts["untestedExport"],
		CoverageThresholdBreakCount: signalCounts["coverageThresholdBreak"],
		SnapshotHeavyCount:         signalCounts["snapshotHeavyTest"],
		QualityPostureBand:         deriveQualityPosture(qualityIssueCount, totalFiles),
	}

	// Change readiness
	blockersByType := map[string]int{}
	for _, s := range snap.Signals {
		if m, ok := s.Metadata["blockerType"]; ok {
			if str, ok := m.(string); ok {
				blockersByType[str]++
			}
		}
	}
	migrationStats := deriveMigrationPosture(snap)
	ms.Change = ChangeMetrics{
		MigrationBlockerCount:         signalCounts["migrationBlocker"],
		DeprecatedPatternCount:        signalCounts["deprecatedTestPattern"],
		DynamicGenerationCount:        signalCounts["dynamicTestGeneration"],
		CustomMatcherRiskCount:        signalCounts["customMatcherRisk"],
		BlockerCountByType:            blockersByType,
		MigrationReadinessBand:        migrationStats.readinessBand,
		SafeAreaCount:                 migrationStats.safeAreas,
		RiskyAreaCount:                migrationStats.riskyAreas,
		QualityCompoundedBlockerCount: migrationStats.compoundedBlockers,
	}

	// Governance
	ms.Governance = GovernanceMetrics{
		PolicyViolationCount:       signalCounts["policyViolation"],
		LegacyFrameworkUsageCount:  signalCounts["legacyFrameworkUsage"],
		RuntimeBudgetExceededCount: signalCounts["runtimeBudgetExceeded"],
	}

	// Risk
	ms.Risk = deriveRiskMetrics(snap)

	// Notes
	if totalFiles == 0 {
		ms.Notes = append(ms.Notes, "No test files detected.")
	}
	hasRuntime := false
	for _, tf := range snap.TestFiles {
		if tf.RuntimeStats != nil && tf.RuntimeStats.AvgRuntimeMs > 0 {
			hasRuntime = true
			break
		}
	}
	if !hasRuntime {
		ms.Notes = append(ms.Notes, "No runtime artifacts detected; health metrics are static-analysis only.")
	}

	return ms
}

func deriveRiskMetrics(snap *models.TestSuiteSnapshot) RiskMetrics {
	rm := RiskMetrics{}
	for _, r := range snap.Risk {
		if r.Scope == "repository" {
			switch r.Type {
			case "reliability":
				rm.ReliabilityBand = string(r.Band)
			case "change":
				rm.ChangeBand = string(r.Band)
			case "speed":
				rm.SpeedBand = string(r.Band)
			}
		}
		if r.Band == models.RiskBandHigh || r.Band == models.RiskBandCritical {
			rm.HighRiskAreaCount++
		}
	}
	for _, s := range snap.Signals {
		if s.Severity == models.SeverityCritical {
			rm.CriticalFindingCount++
		}
	}
	return rm
}

func countSignalsByType(signals []models.Signal) map[string]int {
	counts := map[string]int{}
	for _, s := range signals {
		counts[string(s.Type)]++
	}
	return counts
}

func safeRatio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func deriveQualityPosture(qualityIssueCount, totalFiles int) string {
	if totalFiles == 0 {
		return "unknown"
	}
	ratio := float64(qualityIssueCount) / float64(totalFiles)
	switch {
	case ratio < 0.1:
		return "strong"
	case ratio < 0.3:
		return "moderate"
	default:
		return "weak"
	}
}

// migrationPosture holds aggregate migration posture stats computed
// without importing the migration package (avoids potential cycles).
type migrationPosture struct {
	readinessBand      string
	safeAreas          int
	riskyAreas         int
	compoundedBlockers int
}

func deriveMigrationPosture(snap *models.TestSuiteSnapshot) migrationPosture {
	migrationTypes := signals.MigrationSignalTypes
	qualityTypes := signals.QualitySignalTypes

	totalFiles := len(snap.TestFiles)
	blockerCount := 0
	blockerFiles := map[string]bool{}
	qualityFiles := map[string]bool{}

	// Per-directory tracking.
	type dirInfo struct {
		hasBlocker bool
		hasQuality bool
		hasTests   bool
	}
	dirs := map[string]*dirInfo{}

	for _, tf := range snap.TestFiles {
		dir := dirOf(tf.Path)
		if dirs[dir] == nil {
			dirs[dir] = &dirInfo{}
		}
		dirs[dir].hasTests = true
	}

	for _, s := range snap.Signals {
		if migrationTypes[s.Type] {
			blockerCount++
			if s.Location.File != "" {
				blockerFiles[s.Location.File] = true
				dir := dirOf(s.Location.File)
				if dirs[dir] == nil {
					dirs[dir] = &dirInfo{}
				}
				dirs[dir].hasBlocker = true
			}
		}
		if qualityTypes[s.Type] {
			if s.Location.File != "" {
				qualityFiles[s.Location.File] = true
				dir := dirOf(s.Location.File)
				if dirs[dir] == nil {
					dirs[dir] = &dirInfo{}
				}
				dirs[dir].hasQuality = true
			}
		}
	}

	// Readiness band (same thresholds as migration.deriveReadiness).
	var band string
	if totalFiles == 0 {
		band = "unknown"
	} else {
		ratio := float64(blockerCount) / float64(totalFiles)
		switch {
		case blockerCount == 0:
			band = "high"
		case ratio < 0.1:
			band = "high"
		case ratio < 0.3:
			band = "medium"
		default:
			band = "low"
		}
	}

	// Area counts.
	var safe, risky int
	for _, di := range dirs {
		if !di.hasTests {
			continue
		}
		if di.hasBlocker && di.hasQuality {
			risky++
		} else if !di.hasBlocker && !di.hasQuality {
			safe++
		}
	}

	// Compounded blockers: migration blockers in files that also have quality issues.
	compounded := 0
	for f := range blockerFiles {
		if qualityFiles[f] {
			compounded++
		}
	}

	return migrationPosture{
		readinessBand:      band,
		safeAreas:          safe,
		riskyAreas:         risky,
		compoundedBlockers: compounded,
	}
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}
