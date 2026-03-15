package reporting

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/signals"
)

// ReviewGroup holds findings grouped by a common key (owner, type, directory).
type ReviewGroup struct {
	Key      string
	Signals  []models.Signal
	TopRisk  models.RiskBand
	Count    int
}

// GroupSignalsByOwner groups signals by their owner field.
func GroupSignalsByOwner(signals []models.Signal) []ReviewGroup {
	byOwner := map[string][]models.Signal{}
	for _, s := range signals {
		owner := s.Owner
		if owner == "" {
			owner = "unknown"
		}
		byOwner[owner] = append(byOwner[owner], s)
	}
	return toSortedGroups(byOwner)
}

// GroupSignalsByType groups signals by their type.
func GroupSignalsByType(signals []models.Signal) []ReviewGroup {
	byType := map[string][]models.Signal{}
	for _, s := range signals {
		byType[string(s.Type)] = append(byType[string(s.Type)], s)
	}
	return toSortedGroups(byType)
}

// GroupSignalsByDirectory groups signals by the directory of their file location.
func GroupSignalsByDirectory(signals []models.Signal) []ReviewGroup {
	byDir := map[string][]models.Signal{}
	for _, s := range signals {
		dir := directoryOf(s.Location.File)
		if dir == "" {
			dir = "(repo-level)"
		}
		byDir[dir] = append(byDir[dir], s)
	}
	return toSortedGroups(byDir)
}

// GroupSignalsByCategory groups signals by their category.
func GroupSignalsByCategory(signals []models.Signal) []ReviewGroup {
	byCat := map[string][]models.Signal{}
	for _, s := range signals {
		byCat[string(s.Category)] = append(byCat[string(s.Category)], s)
	}
	return toSortedGroups(byCat)
}

// MigrationBlockers filters signals to migration-related types.
func MigrationBlockers(sigs []models.Signal) []models.Signal {
	var result []models.Signal
	for _, s := range sigs {
		if signals.IsMigrationSignal(s.Type) {
			result = append(result, s)
		}
	}
	return result
}

func toSortedGroups(m map[string][]models.Signal) []ReviewGroup {
	groups := make([]ReviewGroup, 0, len(m))
	for key, sigs := range m {
		groups = append(groups, ReviewGroup{
			Key:     key,
			Signals: sigs,
			Count:   len(sigs),
		})
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Count != groups[j].Count {
			return groups[i].Count > groups[j].Count
		}
		return groups[i].Key < groups[j].Key
	})
	return groups
}

func directoryOf(path string) string {
	if path == "" {
		return ""
	}
	parts := strings.Split(path, "/")
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], "/")
}

// RenderReviewSections appends review-oriented sections to the analyze report.
func RenderReviewSections(w io.Writer, snap *models.TestSuiteSnapshot) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	if len(snap.Signals) == 0 {
		return
	}

	// Top risk areas
	if len(snap.Risk) > 0 {
		line("Highest-Risk Areas")
		line(strings.Repeat("-", 40))
		limit := 5
		if len(snap.Risk) < limit {
			limit = len(snap.Risk)
		}
		for _, r := range snap.Risk[:limit] {
			line("  [%s] %s — %s: %s", r.Band, r.Type, r.Scope, r.ScopeName)
			if r.Explanation != "" {
				line("    %s", r.Explanation)
			}
		}
		blank()
	}

	// Review by owner
	ownerGroups := GroupSignalsByOwner(snap.Signals)
	if len(ownerGroups) > 0 {
		line("Review by Owner")
		line(strings.Repeat("-", 40))
		for _, g := range ownerGroups {
			line("  %-20s %d findings", g.Key, g.Count)
		}
		blank()
	}

	// Review by directory
	dirGroups := GroupSignalsByDirectory(snap.Signals)
	if len(dirGroups) > 1 { // Only show if there are multiple directories
		line("Review by Directory")
		line(strings.Repeat("-", 40))
		limit := 5
		if len(dirGroups) < limit {
			limit = len(dirGroups)
		}
		for _, g := range dirGroups[:limit] {
			line("  %-30s %d findings", g.Key, g.Count)
		}
		if len(dirGroups) > 5 {
			line("  ... and %d more directories", len(dirGroups)-5)
		}
		blank()
	}

	// Migration blockers summary
	blockers := MigrationBlockers(snap.Signals)
	if len(blockers) > 0 {
		line("Migration Blockers")
		line(strings.Repeat("-", 40))
		blockerGroups := GroupSignalsByType(blockers)
		for _, g := range blockerGroups {
			line("  %-26s %d", g.Key, g.Count)
		}
		blank()
	}

	// Coverage by type summary
	if snap.CoverageSummary != nil {
		cs := snap.CoverageSummary
		line("Coverage by Type")
		line(strings.Repeat("-", 40))
		if cs.TotalCodeUnits > 0 {
			line("  Total code units:          %d", cs.TotalCodeUnits)
			line("  Covered by unit tests:     %d", cs.CoveredByUnitTests)
			line("  Covered by e2e:            %d", cs.CoveredByE2E)
			line("  Covered only by e2e:       %d", cs.CoveredOnlyByE2E)
			line("  Uncovered exported:        %d", cs.UncoveredExported)
		}
		if cs.LineCoveragePct > 0 {
			line("  Line coverage:             %.1f%%", cs.LineCoveragePct)
		}
		blank()
	}

	// Test identity summary
	if len(snap.TestCases) > 0 {
		line("Test Identity")
		line(strings.Repeat("-", 40))
		typeCounts := map[string]int{}
		for _, tc := range snap.TestCases {
			t := tc.TestType
			if t == "" {
				t = "unknown"
			}
			typeCounts[t]++
		}
		line("  Total test cases:          %d", len(snap.TestCases))
		for _, tt := range []string{"unit", "integration", "e2e", "unknown"} {
			if c, ok := typeCounts[tt]; ok {
				line("  %-28s %d", tt+":", c)
			}
		}

		// Health signals with test ID
		testIDCount := 0
		for _, s := range snap.Signals {
			if s.Category == models.CategoryHealth {
				if _, ok := s.Metadata["testId"].(string); ok {
					testIDCount++
				}
			}
		}
		if testIDCount > 0 {
			line("  Health signals with testId: %d", testIDCount)
		}
		blank()
	}
}
