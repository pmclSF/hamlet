package ownership

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/signals"
)

// OwnerHealthSummary aggregates health findings for a single owner.
type OwnerHealthSummary struct {
	Owner          string `json:"owner"`
	FlakyCount     int    `json:"flakyCount"`
	SlowCount      int    `json:"slowCount"`
	SkippedCount   int    `json:"skippedCount"`
	TotalHealth    int    `json:"totalHealth"`
	Concentration  string `json:"concentration"` // "localized", "distributed", "none"
	TopFiles       []string `json:"topFiles,omitempty"`
}

// OwnerQualitySummary aggregates quality and coverage findings for a single owner.
type OwnerQualitySummary struct {
	Owner                string `json:"owner"`
	UncoveredExported    int    `json:"uncoveredExported"`
	E2EOnlyUnits         int    `json:"e2eOnlyUnits"`
	WeakAssertionCount   int    `json:"weakAssertionCount"`
	MockHeavyCount       int    `json:"mockHeavyCount"`
	TotalQualitySignals  int    `json:"totalQualitySignals"`
	QualityPosture       string `json:"qualityPosture"` // "strong", "moderate", "weak"
}

// OwnerMigrationSummary aggregates migration findings for a single owner.
type OwnerMigrationSummary struct {
	Owner          string `json:"owner"`
	BlockerCount   int    `json:"blockerCount"`
	BlockerTypes   []string `json:"blockerTypes,omitempty"`
	RiskyAreaCount int    `json:"riskyAreaCount"`
}

// BuildHealthSummaries aggregates health signals by owner.
func BuildHealthSummaries(snap *models.TestSuiteSnapshot) []OwnerHealthSummary {
	type ownerData struct {
		flaky, slow, skipped int
		files                map[string]int
	}
	byOwner := map[string]*ownerData{}

	for _, s := range snap.Signals {
		if s.Category != models.CategoryHealth {
			continue
		}
		owner := s.Owner
		if owner == "" {
			owner = unknownOwner
		}
		d, ok := byOwner[owner]
		if !ok {
			d = &ownerData{files: map[string]int{}}
			byOwner[owner] = d
		}

		switch s.Type {
		case signals.SignalFlakyTest:
			d.flaky++
		case signals.SignalSlowTest:
			d.slow++
		case signals.SignalSkippedTest:
			d.skipped++
		}
		if s.Location.File != "" {
			d.files[s.Location.File]++
		}
	}

	summaries := make([]OwnerHealthSummary, 0, len(byOwner))
	for owner, d := range byOwner {
		total := d.flaky + d.slow + d.skipped
		s := OwnerHealthSummary{
			Owner:       owner,
			FlakyCount:  d.flaky,
			SlowCount:   d.slow,
			SkippedCount: d.skipped,
			TotalHealth: total,
			TopFiles:    topFilesByCount(d.files, 3),
		}

		// Determine concentration.
		if total == 0 {
			s.Concentration = "none"
		} else if len(d.files) <= 2 {
			s.Concentration = "localized"
		} else {
			s.Concentration = "distributed"
		}

		summaries = append(summaries, s)
	}

	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].TotalHealth != summaries[j].TotalHealth {
			return summaries[i].TotalHealth > summaries[j].TotalHealth
		}
		return summaries[i].Owner < summaries[j].Owner
	})

	return summaries
}

// BuildQualitySummaries aggregates quality and coverage findings by owner.
func BuildQualitySummaries(snap *models.TestSuiteSnapshot) []OwnerQualitySummary {
	type ownerData struct {
		uncoveredExported int
		e2eOnly           int
		weakAssertion     int
		mockHeavy         int
		totalQuality      int
	}
	byOwner := map[string]*ownerData{}

	// Count quality signals by owner.
	for _, s := range snap.Signals {
		if s.Category != models.CategoryQuality {
			continue
		}
		owner := s.Owner
		if owner == "" {
			owner = unknownOwner
		}
		d, ok := byOwner[owner]
		if !ok {
			d = &ownerData{}
			byOwner[owner] = d
		}
		d.totalQuality++

		switch s.Type {
		case signals.SignalUntestedExport:
			d.uncoveredExported++
		case signals.SignalWeakAssertion:
			d.weakAssertion++
		case signals.SignalMockHeavyTest:
			d.mockHeavy++
		}
	}

	// Count e2e-only code units by owner.
	for _, ci := range snap.CoverageInsights {
		if ci.Type != "e2e_only_coverage" {
			continue
		}
		// Find owner from code units.
		owner := unknownOwner
		for _, cu := range snap.CodeUnits {
			if cu.Path == ci.Path || cu.UnitID == ci.UnitID {
				if cu.Owner != "" {
					owner = cu.Owner
				}
				break
			}
		}
		d, ok := byOwner[owner]
		if !ok {
			d = &ownerData{}
			byOwner[owner] = d
		}
		d.e2eOnly++
	}

	summaries := make([]OwnerQualitySummary, 0, len(byOwner))
	for owner, d := range byOwner {
		s := OwnerQualitySummary{
			Owner:               owner,
			UncoveredExported:   d.uncoveredExported,
			E2EOnlyUnits:        d.e2eOnly,
			WeakAssertionCount:  d.weakAssertion,
			MockHeavyCount:      d.mockHeavy,
			TotalQualitySignals: d.totalQuality,
		}

		// Derive quality posture.
		if d.totalQuality == 0 && d.e2eOnly == 0 {
			s.QualityPosture = "strong"
		} else if d.totalQuality <= 2 && d.e2eOnly <= 1 {
			s.QualityPosture = "moderate"
		} else {
			s.QualityPosture = "weak"
		}

		summaries = append(summaries, s)
	}

	sort.Slice(summaries, func(i, j int) bool {
		ti := summaries[i].TotalQualitySignals + summaries[i].E2EOnlyUnits
		tj := summaries[j].TotalQualitySignals + summaries[j].E2EOnlyUnits
		if ti != tj {
			return ti > tj
		}
		return summaries[i].Owner < summaries[j].Owner
	})

	return summaries
}

// BuildMigrationSummaries aggregates migration findings by owner.
func BuildMigrationSummaries(snap *models.TestSuiteSnapshot) []OwnerMigrationSummary {
	type ownerData struct {
		blockerCount int
		blockerTypes map[string]bool
	}
	byOwner := map[string]*ownerData{}

	for _, s := range snap.Signals {
		if !signals.IsMigrationSignal(s.Type) {
			continue
		}
		owner := s.Owner
		if owner == "" {
			owner = unknownOwner
		}
		d, ok := byOwner[owner]
		if !ok {
			d = &ownerData{blockerTypes: map[string]bool{}}
			byOwner[owner] = d
		}
		d.blockerCount++
		d.blockerTypes[string(s.Type)] = true
	}

	summaries := make([]OwnerMigrationSummary, 0, len(byOwner))
	for owner, d := range byOwner {
		types := make([]string, 0, len(d.blockerTypes))
		for t := range d.blockerTypes {
			types = append(types, t)
		}
		sort.Strings(types)

		summaries = append(summaries, OwnerMigrationSummary{
			Owner:        owner,
			BlockerCount: d.blockerCount,
			BlockerTypes: types,
		})
	}

	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].BlockerCount != summaries[j].BlockerCount {
			return summaries[i].BlockerCount > summaries[j].BlockerCount
		}
		return summaries[i].Owner < summaries[j].Owner
	})

	return summaries
}

// MigrationCoordinationRisk describes the organizational complexity of a migration.
type MigrationCoordinationRisk struct {
	// OwnerCount is the number of distinct owners with migration blockers.
	OwnerCount int `json:"ownerCount"`

	// TotalBlockers is the total number of migration blockers across all owners.
	TotalBlockers int `json:"totalBlockers"`

	// Level is "low", "medium", or "high".
	Level string `json:"level"`

	// Description summarizes the coordination risk.
	Description string `json:"description"`

	// UnownedBlockerCount is the number of blockers in unowned areas.
	UnownedBlockerCount int `json:"unownedBlockerCount"`
}

// ComputeMigrationCoordinationRisk assesses the organizational complexity
// of the migration based on how blockers distribute across owners.
func ComputeMigrationCoordinationRisk(summaries []OwnerMigrationSummary) MigrationCoordinationRisk {
	risk := MigrationCoordinationRisk{}
	for _, s := range summaries {
		risk.TotalBlockers += s.BlockerCount
		if s.Owner != unknownOwner {
			risk.OwnerCount++
		} else {
			risk.UnownedBlockerCount = s.BlockerCount
		}
	}

	switch {
	case risk.TotalBlockers == 0:
		risk.Level = "low"
		risk.Description = "No migration blockers detected."
	case risk.OwnerCount <= 1:
		risk.Level = "low"
		owner := unknownOwner
		for _, s := range summaries {
			if s.Owner != unknownOwner {
				owner = s.Owner
				break
			}
		}
		risk.Description = fmt.Sprintf("Migration blockers concentrated in one owner area (%s).", owner)
	case risk.OwnerCount <= 3:
		risk.Level = "medium"
		risk.Description = fmt.Sprintf("Migration blockers span %d owner areas — some coordination needed.", risk.OwnerCount)
	default:
		risk.Level = "high"
		risk.Description = fmt.Sprintf("Migration blockers span %d owner areas — significant coordination required.", risk.OwnerCount)
	}

	if risk.UnownedBlockerCount > 0 {
		risk.Description += fmt.Sprintf(" %d blocker(s) are in unowned areas.", risk.UnownedBlockerCount)
	}

	return risk
}

// OwnerTrendSummary describes how an owner area changed between snapshots.
type OwnerTrendSummary struct {
	Owner         string `json:"owner"`
	SignalsBefore int    `json:"signalsBefore"`
	SignalsAfter  int    `json:"signalsAfter"`
	SignalDelta   int    `json:"signalDelta"`
	Direction     string `json:"direction"` // "improved", "worsened", "unchanged"
}

// CompareOwnerSignals compares signal counts by owner between two snapshots.
func CompareOwnerSignals(from, to *models.TestSuiteSnapshot) []OwnerTrendSummary {
	fromCounts := countSignalsByOwner(from.Signals)
	toCounts := countSignalsByOwner(to.Signals)

	allOwners := map[string]bool{}
	for o := range fromCounts {
		allOwners[o] = true
	}
	for o := range toCounts {
		allOwners[o] = true
	}

	var trends []OwnerTrendSummary
	for owner := range allOwners {
		before := fromCounts[owner]
		after := toCounts[owner]
		delta := after - before
		dir := "unchanged"
		if delta > 0 {
			dir = "worsened"
		} else if delta < 0 {
			dir = "improved"
		}

		trends = append(trends, OwnerTrendSummary{
			Owner:         owner,
			SignalsBefore: before,
			SignalsAfter:  after,
			SignalDelta:   delta,
			Direction:     dir,
		})
	}

	// Sort by absolute delta descending.
	sort.Slice(trends, func(i, j int) bool {
		ai := trends[i].SignalDelta
		if ai < 0 {
			ai = -ai
		}
		aj := trends[j].SignalDelta
		if aj < 0 {
			aj = -aj
		}
		if ai != aj {
			return ai > aj
		}
		return trends[i].Owner < trends[j].Owner
	})

	return trends
}

// FocusOwnerItem is an ownership-aware focus recommendation.
type FocusOwnerItem struct {
	// Owner is the relevant owner (or "unknown" for unowned areas).
	Owner string `json:"owner"`

	// Action is the recommended action.
	Action string `json:"action"`

	// Reason explains why this is recommended.
	Reason string `json:"reason"`

	// Priority is 1 (highest) to N.
	Priority int `json:"priority"`
}

// BuildFocusItems generates ownership-aware focus recommendations
// from the owner aggregates.
func BuildFocusItems(summary OwnershipSummary, healthSummaries []OwnerHealthSummary, qualitySummaries []OwnerQualitySummary) []FocusOwnerItem {
	var items []FocusOwnerItem
	priority := 1

	// Unowned critical areas.
	for _, agg := range summary.Owners {
		if agg.Owner.ID == unknownOwner && (agg.CriticalSignalCount > 0 || agg.UncoveredExportedCount > 0) {
			items = append(items, FocusOwnerItem{
				Owner:    unknownOwner,
				Action:   "Assign ownership before remediating risk in unowned areas",
				Reason:   fmt.Sprintf("%d critical signal(s) and %d uncovered export(s) have no owner", agg.CriticalSignalCount, agg.UncoveredExportedCount),
				Priority: priority,
			})
			priority++
			break
		}
	}

	// Concentrated instability.
	for _, hs := range healthSummaries {
		if hs.Concentration == "localized" && hs.TotalHealth >= 2 {
			action := fmt.Sprintf("Start with %s's concentrated instability", hs.Owner)
			if hs.Owner == unknownOwner {
				action = "Investigate concentrated instability in unowned area"
			}
			reason := fmt.Sprintf("%d health finding(s) concentrated in %s", hs.TotalHealth, strings.Join(hs.TopFiles, ", "))
			items = append(items, FocusOwnerItem{
				Owner:    hs.Owner,
				Action:   action,
				Reason:   reason,
				Priority: priority,
			})
			priority++
			if priority > 5 {
				break
			}
		}
	}

	// Quality gaps.
	for _, qs := range qualitySummaries {
		if qs.QualityPosture == "weak" {
			action := fmt.Sprintf("Improve test quality in %s's area", qs.Owner)
			if qs.Owner == unknownOwner {
				action = "Improve test quality in unowned area"
			}
			var parts []string
			if qs.UncoveredExported > 0 {
				parts = append(parts, fmt.Sprintf("%d uncovered export(s)", qs.UncoveredExported))
			}
			if qs.E2EOnlyUnits > 0 {
				parts = append(parts, fmt.Sprintf("%d e2e-only unit(s)", qs.E2EOnlyUnits))
			}
			if qs.WeakAssertionCount > 0 {
				parts = append(parts, fmt.Sprintf("%d weak assertion(s)", qs.WeakAssertionCount))
			}
			items = append(items, FocusOwnerItem{
				Owner:    qs.Owner,
				Action:   action,
				Reason:   strings.Join(parts, ", "),
				Priority: priority,
			})
			priority++
			if priority > 5 {
				break
			}
		}
	}

	return items
}

// OwnershipBenchmarkAggregate holds privacy-safe ownership statistics
// for the benchmark export.
type OwnershipBenchmarkAggregate struct {
	// OwnerCount is the number of distinct owners.
	OwnerCount int `json:"ownerCount"`

	// CoveragePosture is the ownership coverage posture.
	CoveragePosture string `json:"coveragePosture"`

	// TopOwnerRiskSharePct is the percentage of all signals
	// concentrated in the single most-affected owner.
	TopOwnerRiskSharePct float64 `json:"topOwnerRiskSharePct"`

	// UnownedCriticalPct is the percentage of critical code units
	// that are in unowned areas.
	UnownedCriticalPct float64 `json:"unownedCriticalPct"`

	// FragmentationIndex measures how evenly risk is distributed
	// across owners (0 = all in one owner, 1 = perfectly even).
	FragmentationIndex float64 `json:"fragmentationIndex"`
}

// BuildBenchmarkAggregate creates a privacy-safe ownership aggregate
// for the benchmark export. No owner names or paths are included.
func BuildBenchmarkAggregate(summary OwnershipSummary, snap *models.TestSuiteSnapshot) *OwnershipBenchmarkAggregate {
	if summary.OwnerCount == 0 {
		return nil
	}

	agg := &OwnershipBenchmarkAggregate{
		OwnerCount:      summary.OwnerCount,
		CoveragePosture: summary.CoveragePosture,
	}

	// Top owner risk share.
	totalSignals := 0
	maxSignals := 0
	for _, o := range summary.Owners {
		totalSignals += o.SignalCount
		if o.SignalCount > maxSignals {
			maxSignals = o.SignalCount
		}
	}
	if totalSignals > 0 {
		agg.TopOwnerRiskSharePct = float64(maxSignals) / float64(totalSignals) * 100
	}

	// Unowned critical code units.
	totalExported := 0
	unownedExported := 0
	for _, cu := range snap.CodeUnits {
		if cu.Exported {
			totalExported++
			if cu.Owner == "" || cu.Owner == unknownOwner {
				unownedExported++
			}
		}
	}
	if totalExported > 0 {
		agg.UnownedCriticalPct = float64(unownedExported) / float64(totalExported) * 100
	}

	// Fragmentation index (normalized entropy).
	if totalSignals > 0 && summary.OwnerCount > 1 {
		agg.FragmentationIndex = computeFragmentation(summary.Owners, totalSignals)
	}

	return agg
}

// computeFragmentation calculates a normalized entropy-based fragmentation index.
// 0 = all signals in one owner, 1 = perfectly even distribution.
func computeFragmentation(owners []OwnerAggregate, totalSignals int) float64 {
	if totalSignals == 0 || len(owners) <= 1 {
		return 0
	}
	// Collect signal counts for active owners, sorted for deterministic
	// floating-point accumulation regardless of input order.
	var counts []int
	for _, o := range owners {
		if o.SignalCount > 0 {
			counts = append(counts, o.SignalCount)
		}
	}
	if len(counts) <= 1 {
		return 0
	}
	sort.Ints(counts)

	// Normalized Herfindahl index (inverted so higher = more fragmented).
	total := float64(totalSignals)
	sumSquares := 0.0
	for _, c := range counts {
		share := float64(c) / total
		sumSquares += share * share
	}
	// HHI ranges from 1/n (perfectly even) to 1 (all in one).
	// Normalize: 0 = concentrated, 1 = perfectly even.
	minHHI := 1.0 / float64(len(counts))
	if sumSquares >= 1.0 {
		return 0
	}
	return (1.0 - sumSquares) / (1.0 - minHHI)
}

func countSignalsByOwner(sigs []models.Signal) map[string]int {
	counts := map[string]int{}
	for _, s := range sigs {
		owner := s.Owner
		if owner == "" {
			owner = unknownOwner
		}
		counts[owner]++
	}
	return counts
}

func topFilesByCount(files map[string]int, n int) []string {
	type kv struct {
		file  string
		count int
	}
	pairs := make([]kv, 0, len(files))
	for f, c := range files {
		pairs = append(pairs, kv{f, c})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count != pairs[j].count {
			return pairs[i].count > pairs[j].count
		}
		return pairs[i].file < pairs[j].file
	})
	if len(pairs) > n {
		pairs = pairs[:n]
	}
	result := make([]string, len(pairs))
	for i, p := range pairs {
		result[i] = p.file
	}
	return result
}
