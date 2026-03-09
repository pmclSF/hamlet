package migration

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/signals"
)

// ReadinessSummary summarizes migration readiness for a repository.
type ReadinessSummary struct {
	// Frameworks lists detected frameworks with file counts.
	Frameworks []models.Framework `json:"frameworks"`

	// TotalBlockers is the count of hard + soft blockers (excludes advisories).
	TotalBlockers int `json:"totalBlockers"`

	// HardBlockers is the count of signals that genuinely prevent migration.
	HardBlockers int `json:"hardBlockers"`

	// SoftBlockers is the count of signals that require effort but have clear paths.
	SoftBlockers int `json:"softBlockers"`

	// Advisories is the count of signals worth noting but not affecting readiness.
	Advisories int `json:"advisories"`

	// BlockersByType groups blocker counts by blocker taxonomy category.
	BlockersByType map[string]int `json:"blockersByType"`

	// BlockersByTier groups counts by severity tier (hard-blocker, soft-blocker, advisory).
	BlockersByTier map[string]int `json:"blockersByTier"`

	// RepresentativeBlockers shows a few example blockers.
	RepresentativeBlockers []BlockerExample `json:"representativeBlockers,omitempty"`

	// ReadinessLevel is a qualitative assessment: low, medium, high, unknown.
	// Derived from visible blocker patterns and risk — not a magic score.
	// "unknown" is used when framework detection confidence is too low.
	ReadinessLevel string `json:"readinessLevel"`

	// Explanation describes why the readiness level was assigned.
	Explanation string `json:"explanation"`

	// QualityFactors describes how quality signals compound migration risk.
	// Present only when quality signals co-occur with migration targets.
	QualityFactors []QualityFactor `json:"qualityFactors,omitempty"`

	// AreaAssessments classifies directories by migration safety,
	// combining blocker presence with quality signal density.
	AreaAssessments []AreaAssessment `json:"areaAssessments,omitempty"`

	// CoverageGuidance suggests where additional test coverage would
	// most reduce migration risk.
	CoverageGuidance []CoverageGuidanceItem `json:"coverageGuidance,omitempty"`
}

// BlockerExample is a representative migration blocker for display.
type BlockerExample struct {
	Type        string `json:"type"`
	File        string `json:"file"`
	Explanation string `json:"explanation"`
}

// QualityFactor describes how a quality signal type compounds migration risk.
type QualityFactor struct {
	// SignalType is the quality signal (e.g. "weakAssertion").
	SignalType string `json:"signalType"`

	// AffectedFiles is the count of migration-target files also affected
	// by this quality issue.
	AffectedFiles int `json:"affectedFiles"`

	// Explanation describes the compounding effect.
	Explanation string `json:"explanation"`
}

// AreaAssessment classifies a directory's migration safety based on
// the combination of migration blockers and quality signals.
type AreaAssessment struct {
	// Directory is the assessed directory path.
	Directory string `json:"directory"`

	// Classification is one of: "safe", "caution", "risky".
	Classification string `json:"classification"`

	// MigrationBlockers is the count of migration signals in this directory.
	MigrationBlockers int `json:"migrationBlockers"`

	// QualityIssues is the count of quality signals in this directory.
	QualityIssues int `json:"qualityIssues"`

	// TestFileCount is the number of test files in this directory.
	TestFileCount int `json:"testFileCount"`

	// Explanation describes why the classification was assigned.
	Explanation string `json:"explanation"`
}

// CoverageGuidanceItem suggests where additional coverage would reduce
// migration risk the most.
type CoverageGuidanceItem struct {
	// Directory is the area that would benefit from more coverage.
	Directory string `json:"directory"`

	// Reason explains why coverage matters here for migration.
	Reason string `json:"reason"`

	// Priority is "high", "medium", or "low".
	Priority string `json:"priority"`
}

// Use the canonical signal type sets from the signals package.
var (
	migrationTypes = signals.MigrationSignalTypes
	qualityTypes   = signals.QualitySignalTypes
)

// ComputeReadiness derives a migration readiness summary from the snapshot.
//
// Readiness levels:
//   - "high": few or no migration blockers
//   - "medium": some blockers but manageable
//   - "low": many blockers requiring significant effort
//
// The summary also cross-references quality signals with migration targets
// to surface areas where poor test quality amplifies migration risk.
func ComputeReadiness(snap *models.TestSuiteSnapshot) *ReadinessSummary {
	var allMigrationSignals []models.Signal
	blockersByType := map[string]int{}
	blockersByTier := map[string]int{}
	hardCount, softCount, advisoryCount := 0, 0, 0

	for _, s := range snap.Signals {
		if !migrationTypes[s.Type] {
			continue
		}
		allMigrationSignals = append(allMigrationSignals, s)

		bt := "other"
		if m, ok := s.Metadata["blockerType"]; ok {
			if str, ok := m.(string); ok {
				bt = str
			}
		}
		blockersByType[bt]++

		tier := TierForSignal(s)
		blockersByTier[tier]++
		switch tier {
		case TierHardBlocker:
			hardCount++
		case TierSoftBlocker:
			softCount++
		case TierAdvisory:
			advisoryCount++
		}
	}

	// Only hard + soft blockers count toward readiness.
	effectiveBlockerCount := hardCount + softCount

	// Build representative examples (up to 5, prioritize hard blockers).
	var examples []BlockerExample
	addedCount := 0
	// Add hard blockers first, then soft, then advisory.
	for _, tier := range []string{TierHardBlocker, TierSoftBlocker, TierAdvisory} {
		for _, b := range allMigrationSignals {
			if addedCount >= 5 {
				break
			}
			if TierForSignal(b) == tier {
				examples = append(examples, BlockerExample{
					Type:        string(b.Type),
					File:        b.Location.File,
					Explanation: b.Explanation,
				})
				addedCount++
			}
		}
	}

	// Check framework detection confidence before deriving readiness.
	// If we can't confidently identify frameworks, readiness is unknown.
	totalFiles := len(snap.TestFiles)
	readiness, explanation := deriveReadinessWithTiers(
		hardCount, softCount, advisoryCount, totalFiles, blockersByType,
		frameworkConfidence(snap),
	)

	// Cross-reference quality signals with migration targets
	qualityFactors := computeQualityFactors(snap, allMigrationSignals)
	areas := computeAreaAssessments(snap)
	guidance := computeCoverageGuidance(areas, snap)

	return &ReadinessSummary{
		Frameworks:             snap.Frameworks,
		TotalBlockers:          effectiveBlockerCount,
		HardBlockers:           hardCount,
		SoftBlockers:           softCount,
		Advisories:             advisoryCount,
		BlockersByType:         blockersByType,
		BlockersByTier:         blockersByTier,
		RepresentativeBlockers: examples,
		ReadinessLevel:         readiness,
		Explanation:            explanation,
		QualityFactors:         qualityFactors,
		AreaAssessments:        areas,
		CoverageGuidance:       guidance,
	}
}

// frameworkConfidence returns the average confidence across detected frameworks.
// Returns 1.0 if no confidence data is available (backwards compatibility).
func frameworkConfidence(snap *models.TestSuiteSnapshot) float64 {
	if len(snap.Frameworks) == 0 {
		return 1.0 // No framework data available — assume confident (legacy behavior).
	}
	// If no framework has confidence set, assume full confidence (legacy behavior).
	hasConfidence := false
	total := 0.0
	for _, fw := range snap.Frameworks {
		if fw.Confidence > 0 {
			hasConfidence = true
			total += fw.Confidence
		}
	}
	if !hasConfidence {
		return 1.0
	}
	return total / float64(len(snap.Frameworks))
}

// computeQualityFactors identifies quality signals that co-occur with
// migration blocker files, amplifying migration risk.
func computeQualityFactors(snap *models.TestSuiteSnapshot, blockers []models.Signal) []QualityFactor {
	// Build set of files with migration blockers.
	blockerFiles := map[string]bool{}
	for _, b := range blockers {
		if b.Location.File != "" {
			blockerFiles[b.Location.File] = true
		}
	}

	if len(blockerFiles) == 0 {
		return nil
	}

	// Count quality signals that affect blocker files.
	type qualityHit struct {
		signalType string
		files      map[string]bool
	}
	hits := map[string]*qualityHit{}

	for _, s := range snap.Signals {
		if !qualityTypes[s.Type] {
			continue
		}
		file := s.Location.File
		if file == "" || !blockerFiles[file] {
			continue
		}
		st := string(s.Type)
		if hits[st] == nil {
			hits[st] = &qualityHit{signalType: st, files: map[string]bool{}}
		}
		hits[st].files[file] = true
	}

	var factors []QualityFactor
	for _, h := range hits {
		count := len(h.files)
		factors = append(factors, QualityFactor{
			SignalType:    h.signalType,
			AffectedFiles: count,
			Explanation:   qualityFactorExplanation(h.signalType, count),
		})
	}

	// Sort by affected file count descending for stable output.
	sort.Slice(factors, func(i, j int) bool {
		if factors[i].AffectedFiles != factors[j].AffectedFiles {
			return factors[i].AffectedFiles > factors[j].AffectedFiles
		}
		return factors[i].SignalType < factors[j].SignalType
	})

	return factors
}

func qualityFactorExplanation(signalType string, count int) string {
	switch signalType {
	case "weakAssertion":
		return fmt.Sprintf(
			"%d migration-target file(s) have weak assertions. Low assertion density means migrations may pass tests without verifying correctness.",
			count,
		)
	case "mockHeavyTest":
		return fmt.Sprintf(
			"%d migration-target file(s) are mock-heavy. Heavy mocking couples tests to implementation details that will change during migration.",
			count,
		)
	case "untestedExport":
		return fmt.Sprintf(
			"%d migration-target file(s) have untested exports. Untested public API makes it harder to verify migration correctness.",
			count,
		)
	case "coverageThresholdBreak":
		return fmt.Sprintf(
			"%d migration-target file(s) fall below coverage thresholds. Low coverage means less confidence in migration safety.",
			count,
		)
	case "coverageBlindSpot":
		return fmt.Sprintf(
			"%d migration-target file(s) have coverage blind spots. Uncovered code paths are invisible during migration validation.",
			count,
		)
	default:
		return fmt.Sprintf(
			"%d migration-target file(s) have quality issue '%s', adding risk to migration.",
			count, signalType,
		)
	}
}

// dirStats aggregates signal counts per directory.
type dirStats struct {
	dir               string
	migrationBlockers int
	qualityIssues     int
	testFileCount     int
}

// computeAreaAssessments classifies directories by migration safety.
func computeAreaAssessments(snap *models.TestSuiteSnapshot) []AreaAssessment {
	dirs := map[string]*dirStats{}

	// Count test files per directory.
	for _, tf := range snap.TestFiles {
		dir := filepath.Dir(tf.Path)
		if dirs[dir] == nil {
			dirs[dir] = &dirStats{dir: dir}
		}
		dirs[dir].testFileCount++
	}

	// Count signals per directory.
	for _, s := range snap.Signals {
		file := s.Location.File
		if file == "" {
			continue
		}
		dir := filepath.Dir(file)
		if dirs[dir] == nil {
			dirs[dir] = &dirStats{dir: dir}
		}
		if migrationTypes[s.Type] {
			dirs[dir].migrationBlockers++
		}
		if qualityTypes[s.Type] {
			dirs[dir].qualityIssues++
		}
	}

	// Only assess directories that have test files.
	var assessments []AreaAssessment
	for _, ds := range dirs {
		if ds.testFileCount == 0 {
			continue
		}

		classification, explanation := classifyArea(ds)
		assessments = append(assessments, AreaAssessment{
			Directory:         ds.dir,
			Classification:    classification,
			MigrationBlockers: ds.migrationBlockers,
			QualityIssues:     ds.qualityIssues,
			TestFileCount:     ds.testFileCount,
			Explanation:       explanation,
		})
	}

	// Sort: risky first, then caution, then safe. Within same class, by directory name.
	classOrder := map[string]int{"risky": 0, "caution": 1, "safe": 2}
	sort.Slice(assessments, func(i, j int) bool {
		ci, cj := classOrder[assessments[i].Classification], classOrder[assessments[j].Classification]
		if ci != cj {
			return ci < cj
		}
		return assessments[i].Directory < assessments[j].Directory
	})

	return assessments
}

func classifyArea(ds *dirStats) (string, string) {
	hasBlockers := ds.migrationBlockers > 0
	hasQuality := ds.qualityIssues > 0

	if hasBlockers && hasQuality {
		return "risky", fmt.Sprintf(
			"%d migration blocker(s) compounded by %d quality issue(s) across %d test file(s). Address quality issues before migrating.",
			ds.migrationBlockers, ds.qualityIssues, ds.testFileCount,
		)
	}

	if hasBlockers {
		return "caution", fmt.Sprintf(
			"%d migration blocker(s) in %d test file(s), but test quality is adequate. Migration requires blocker remediation.",
			ds.migrationBlockers, ds.testFileCount,
		)
	}

	if hasQuality {
		// Quality issues without migration blockers — still worth noting
		// but not a migration barrier.
		return "caution", fmt.Sprintf(
			"No migration blockers, but %d quality issue(s) in %d test file(s). Consider improving test quality before or during migration.",
			ds.qualityIssues, ds.testFileCount,
		)
	}

	return "safe", fmt.Sprintf(
		"%d test file(s) with no migration blockers and no quality issues. Safe to modernize.",
		ds.testFileCount,
	)
}

// computeCoverageGuidance identifies directories where additional test
// coverage would most reduce migration risk.
func computeCoverageGuidance(areas []AreaAssessment, snap *models.TestSuiteSnapshot) []CoverageGuidanceItem {
	// Build a set of directories with untested exports.
	untestedDirs := map[string]int{}
	for _, s := range snap.Signals {
		if s.Type == "untestedExport" && s.Location.File != "" {
			dir := filepath.Dir(s.Location.File)
			untestedDirs[dir]++
		}
	}

	// Build a set of directories with e2e-only coverage from insights.
	e2eOnlyDirs := map[string]int{}
	for _, ci := range snap.CoverageInsights {
		if ci.Type == "e2e_only_coverage" && ci.Path != "" {
			dir := filepath.Dir(ci.Path)
			e2eOnlyDirs[dir]++
		}
	}

	var guidance []CoverageGuidanceItem

	for _, area := range areas {
		if area.Classification == "safe" {
			continue
		}

		priority := "medium"
		var reasons []string

		if area.MigrationBlockers > 0 && area.QualityIssues > 0 {
			priority = "high"
			reasons = append(reasons, "migration blockers combined with quality issues")
		}

		if count, ok := untestedDirs[area.Directory]; ok && count > 0 {
			priority = "high"
			reasons = append(reasons, fmt.Sprintf("%d untested export(s) in migration target", count))
		}

		if count, ok := e2eOnlyDirs[area.Directory]; ok && count > 0 {
			priority = "high"
			reasons = append(reasons, fmt.Sprintf("%d code unit(s) covered only by e2e — no fast feedback during migration", count))
		}

		if area.MigrationBlockers > 0 && area.QualityIssues == 0 {
			reasons = append(reasons, "migration blockers present — stronger tests would increase confidence in blocker remediation")
		}

		if area.QualityIssues > 0 && area.MigrationBlockers == 0 {
			reasons = append(reasons, "quality issues could become migration risks if framework changes are needed")
		}

		if len(reasons) == 0 {
			continue
		}

		guidance = append(guidance, CoverageGuidanceItem{
			Directory: area.Directory,
			Reason:    strings.Join(reasons, "; "),
			Priority:  priority,
		})
	}

	// Sort by priority (high first).
	priorityOrder := map[string]int{"high": 0, "medium": 1, "low": 2}
	sort.Slice(guidance, func(i, j int) bool {
		pi, pj := priorityOrder[guidance[i].Priority], priorityOrder[guidance[j].Priority]
		if pi != pj {
			return pi < pj
		}
		return guidance[i].Directory < guidance[j].Directory
	})

	return guidance
}

// deriveReadiness is the legacy readiness function. Kept for backward compatibility
// with tests that don't use tiers.
func deriveReadiness(blockerCount, totalFiles int, byType map[string]int) (string, string) {
	return deriveReadinessWithTiers(blockerCount, 0, 0, totalFiles, byType, 1.0)
}

// deriveReadinessWithTiers computes readiness level using the blocker tier taxonomy.
//
// Only hard and soft blockers affect readiness. Advisories are informational.
// When framework detection confidence is below 0.5, readiness is "unknown"
// because we can't reliably assess migration complexity.
func deriveReadinessWithTiers(hardCount, softCount, advisoryCount, totalFiles int, byType map[string]int, fwConfidence float64) (string, string) {
	if totalFiles == 0 {
		return "unknown", "No test files detected."
	}

	// If framework detection confidence is too low, we can't assess readiness.
	if fwConfidence < 0.5 {
		return "unknown", fmt.Sprintf(
			"Framework detection confidence is low (%.0f%%); migration readiness cannot be reliably assessed. "+
				"Ensure test files have identifiable framework imports or add framework configuration.",
			fwConfidence*100,
		)
	}

	// Effective blockers: hard blockers count fully, soft blockers at half weight.
	effectiveCount := hardCount + softCount
	if effectiveCount == 0 {
		if advisoryCount > 0 {
			return "high", fmt.Sprintf(
				"No migration blockers detected. %d advisory note(s) for awareness.",
				advisoryCount,
			)
		}
		return "high", "No migration blockers detected."
	}

	ratio := float64(effectiveCount) / float64(totalFiles)

	tierSummary := ""
	if hardCount > 0 {
		tierSummary = fmt.Sprintf("%d hard blocker(s)", hardCount)
	}
	if softCount > 0 {
		if tierSummary != "" {
			tierSummary += ", "
		}
		tierSummary += fmt.Sprintf("%d soft blocker(s)", softCount)
	}
	if advisoryCount > 0 {
		if tierSummary != "" {
			tierSummary += ", "
		}
		tierSummary += fmt.Sprintf("%d advisory", advisoryCount)
	}

	if ratio < 0.1 {
		topType := dominantType(byType)
		return "high", fmt.Sprintf(
			"Few migration blockers (%s across %d test files). Primary: %s.",
			tierSummary, totalFiles, topType,
		)
	}

	if ratio < 0.3 {
		topType := dominantType(byType)
		return "medium", fmt.Sprintf(
			"Some migration blockers (%s across %d test files). Focus on: %s.",
			tierSummary, totalFiles, topType,
		)
	}

	topType := dominantType(byType)
	return "low", fmt.Sprintf(
		"Many migration blockers (%s across %d test files). Major blocker: %s.",
		tierSummary, totalFiles, topType,
	)
}

func dominantType(byType map[string]int) string {
	type kv struct {
		key   string
		count int
	}
	var pairs []kv
	for k, v := range byType {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count != pairs[j].count {
			return pairs[i].count > pairs[j].count
		}
		return pairs[i].key < pairs[j].key
	})
	if len(pairs) == 0 {
		return "unknown"
	}
	names := make([]string, 0, len(pairs))
	for _, p := range pairs {
		names = append(names, fmt.Sprintf("%s (%d)", p.key, p.count))
	}
	return strings.Join(names, ", ")
}
