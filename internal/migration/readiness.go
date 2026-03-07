package migration

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
)

// ReadinessSummary summarizes migration readiness for a repository.
type ReadinessSummary struct {
	// Frameworks lists detected frameworks with file counts.
	Frameworks []models.Framework `json:"frameworks"`

	// TotalBlockers is the count of migration-related signals.
	TotalBlockers int `json:"totalBlockers"`

	// BlockersByType groups blocker counts by blocker taxonomy category.
	BlockersByType map[string]int `json:"blockersByType"`

	// RepresentativeBlockers shows a few example blockers.
	RepresentativeBlockers []BlockerExample `json:"representativeBlockers,omitempty"`

	// ReadinessLevel is a qualitative assessment: low, medium, high.
	// Derived from visible blocker patterns and risk — not a magic score.
	ReadinessLevel string `json:"readinessLevel"`

	// Explanation describes why the readiness level was assigned.
	Explanation string `json:"explanation"`
}

// BlockerExample is a representative migration blocker for display.
type BlockerExample struct {
	Type        string `json:"type"`
	File        string `json:"file"`
	Explanation string `json:"explanation"`
}

// ComputeReadiness derives a migration readiness summary from the snapshot.
//
// Readiness levels:
//   - "high": few or no migration blockers
//   - "medium": some blockers but manageable
//   - "low": many blockers requiring significant effort
//
// These derive from visible blocker counts, not hidden logic.
func ComputeReadiness(snap *models.TestSuiteSnapshot) *ReadinessSummary {
	migrationTypes := map[models.SignalType]bool{
		"frameworkMigration":    true,
		"migrationBlocker":     true,
		"deprecatedTestPattern": true,
		"dynamicTestGeneration": true,
		"customMatcherRisk":     true,
	}

	var blockers []models.Signal
	blockersByType := map[string]int{}

	for _, s := range snap.Signals {
		if !migrationTypes[s.Type] {
			continue
		}
		blockers = append(blockers, s)
		bt := "other"
		if m, ok := s.Metadata["blockerType"]; ok {
			if str, ok := m.(string); ok {
				bt = str
			}
		}
		blockersByType[bt]++
	}

	// Build representative examples (up to 5)
	var examples []BlockerExample
	limit := 5
	if len(blockers) < limit {
		limit = len(blockers)
	}
	for _, b := range blockers[:limit] {
		examples = append(examples, BlockerExample{
			Type:        string(b.Type),
			File:        b.Location.File,
			Explanation: b.Explanation,
		})
	}

	// Derive readiness level from blocker count relative to test files
	totalFiles := len(snap.TestFiles)
	readiness, explanation := deriveReadiness(len(blockers), totalFiles, blockersByType)

	return &ReadinessSummary{
		Frameworks:             snap.Frameworks,
		TotalBlockers:          len(blockers),
		BlockersByType:         blockersByType,
		RepresentativeBlockers: examples,
		ReadinessLevel:         readiness,
		Explanation:            explanation,
	}
}

func deriveReadiness(blockerCount, totalFiles int, byType map[string]int) (string, string) {
	if totalFiles == 0 {
		return "unknown", "No test files detected."
	}

	ratio := float64(blockerCount) / float64(totalFiles)

	if blockerCount == 0 {
		return "high", "No migration blockers detected."
	}

	if ratio < 0.1 {
		topType := dominantType(byType)
		return "high", fmt.Sprintf(
			"Few migration blockers (%d across %d test files). Primary: %s.",
			blockerCount, totalFiles, topType,
		)
	}

	if ratio < 0.3 {
		topType := dominantType(byType)
		return "medium", fmt.Sprintf(
			"Some migration blockers (%d across %d test files). Focus on: %s.",
			blockerCount, totalFiles, topType,
		)
	}

	topType := dominantType(byType)
	return "low", fmt.Sprintf(
		"Many migration blockers (%d across %d test files). Major blocker: %s.",
		blockerCount, totalFiles, topType,
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
		return pairs[i].count > pairs[j].count
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
