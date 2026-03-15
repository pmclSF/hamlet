package ownership

import (
	"sort"

	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/signals"
)

// PropagateResult holds the output of ownership propagation.
type PropagateResult struct {
	// Summary is the snapshot-level ownership overview.
	Summary OwnershipSummary

	// FileAssignments maps file paths to their resolved assignments.
	FileAssignments map[string]OwnershipAssignment
}

// Propagate resolves ownership for all entities in the snapshot and
// populates ownership fields on signals, code units, and test files.
//
// This is the main integration point: call it during pipeline execution
// after signal detection and before downstream aggregation.
func Propagate(resolver *Resolver, snap *models.TestSuiteSnapshot) *PropagateResult {
	result := &PropagateResult{
		FileAssignments: make(map[string]OwnershipAssignment),
	}

	// Phase 1: Resolve file-level ownership for all test files.
	for i := range snap.TestFiles {
		tf := &snap.TestFiles[i]
		a := resolver.ResolveAssignment(tf.Path)
		result.FileAssignments[tf.Path] = a
		tf.Owner = a.PrimaryOwnerID()
	}

	// Phase 2: Resolve code unit ownership — inherit from file if not directly set.
	for i := range snap.CodeUnits {
		cu := &snap.CodeUnits[i]
		if cu.Owner != "" {
			// Already directly assigned — keep it.
			continue
		}
		// Try to resolve directly first.
		a := resolver.ResolveAssignment(cu.Path)
		if a.Source != SourceUnknown {
			cu.Owner = a.PrimaryOwnerID()
			result.FileAssignments[cu.Path] = a
		}
	}

	// Phase 3: Propagate ownership to signals.
	for i := range snap.Signals {
		s := &snap.Signals[i]
		if s.Owner != "" {
			continue
		}
		if s.Location.File == "" {
			continue
		}
		// Check if we already resolved this file.
		if a, ok := result.FileAssignments[s.Location.File]; ok {
			s.Owner = a.PrimaryOwnerID()
		} else {
			a := resolver.ResolveAssignment(s.Location.File)
			result.FileAssignments[s.Location.File] = a
			s.Owner = a.PrimaryOwnerID()
		}
	}

	// Phase 4: Build the ownership map on the snapshot.
	ownershipMap := make(map[string][]string)
	for path, a := range result.FileAssignments {
		if a.IsUnowned() {
			continue
		}
		owners := make([]string, len(a.Owners))
		for i, o := range a.Owners {
			owners[i] = o.ID
		}
		ownershipMap[path] = owners
	}
	if len(ownershipMap) > 0 {
		snap.Ownership = ownershipMap
	}

	// Phase 5: Compute ownership summary.
	result.Summary = computeSummary(snap, result.FileAssignments, resolver)

	return result
}

// computeSummary builds the OwnershipSummary from the resolved state.
func computeSummary(snap *models.TestSuiteSnapshot, fileAssignments map[string]OwnershipAssignment, resolver *Resolver) OwnershipSummary {
	summary := OwnershipSummary{
		Sources: resolver.SourcesUsed(),
	}

	// Count file ownership.
	allFiles := map[string]bool{}
	for _, tf := range snap.TestFiles {
		allFiles[tf.Path] = true
	}
	for _, cu := range snap.CodeUnits {
		allFiles[cu.Path] = true
	}
	summary.TotalFiles = len(allFiles)
	for path := range allFiles {
		a, ok := fileAssignments[path]
		if ok && !a.IsUnowned() && a.Source != SourceUnknown {
			summary.OwnedFiles++
		}
	}
	summary.UnownedFiles = summary.TotalFiles - summary.OwnedFiles

	// Count code unit ownership.
	summary.TotalCodeUnits = len(snap.CodeUnits)
	for _, cu := range snap.CodeUnits {
		if cu.Owner != "" && cu.Owner != unknownOwner {
			summary.OwnedCodeUnits++
		}
	}

	// Count test case ownership (inherited from test file).
	summary.TotalTestCases = len(snap.TestCases)
	for _, tc := range snap.TestCases {
		if a, ok := fileAssignments[tc.FilePath]; ok && !a.IsUnowned() && a.Source != SourceUnknown {
			summary.OwnedTestCases++
		}
	}

	// Build per-owner aggregates.
	ownerAggs := map[string]*OwnerAggregate{}
	ensureOwner := func(id string) *OwnerAggregate {
		if agg, ok := ownerAggs[id]; ok {
			return agg
		}
		agg := &OwnerAggregate{Owner: Owner{ID: id}}
		ownerAggs[id] = agg
		return agg
	}

	// Files per owner.
	ownerFiles := map[string]map[string]bool{} // owner -> set of file paths
	for path, a := range fileAssignments {
		if a.IsUnowned() || a.Source == SourceUnknown {
			continue
		}
		for _, o := range a.Owners {
			if ownerFiles[o.ID] == nil {
				ownerFiles[o.ID] = map[string]bool{}
			}
			ownerFiles[o.ID][path] = true
		}
	}
	for id, files := range ownerFiles {
		agg := ensureOwner(id)
		agg.FileCount = len(files)
	}

	// Code units per owner.
	for _, cu := range snap.CodeUnits {
		owner := cu.Owner
		if owner == "" || owner == unknownOwner {
			continue
		}
		agg := ensureOwner(owner)
		agg.CodeUnitCount++
		if cu.Exported {
			agg.ExportedCodeUnitCount++
			if cu.Coverage == 0 && len(cu.LinkedTestFiles) == 0 {
				agg.UncoveredExportedCount++
			}
		}
	}

	// Test cases per owner (inherited from file).
	for _, tc := range snap.TestCases {
		if a, ok := fileAssignments[tc.FilePath]; ok && !a.IsUnowned() {
			for _, o := range a.Owners {
				agg := ensureOwner(o.ID)
				agg.TestCaseCount++
			}
		}
	}

	// Signals per owner.
	for _, s := range snap.Signals {
		owner := s.Owner
		if owner == "" || owner == unknownOwner {
			continue
		}
		agg := ensureOwner(owner)
		agg.SignalCount++
		if s.Severity == models.SeverityCritical {
			agg.CriticalSignalCount++
		}
		if s.Category == models.CategoryHealth {
			agg.HealthSignalCount++
		}
		if signals.IsMigrationSignal(s.Type) {
			agg.MigrationBlockerCount++
		}
	}

	// Convert to sorted slice.
	owners := make([]OwnerAggregate, 0, len(ownerAggs))
	for _, agg := range ownerAggs {
		owners = append(owners, *agg)
	}
	sort.Slice(owners, func(i, j int) bool {
		if owners[i].SignalCount != owners[j].SignalCount {
			return owners[i].SignalCount > owners[j].SignalCount
		}
		return owners[i].Owner.ID < owners[j].Owner.ID
	})

	summary.Owners = owners
	summary.OwnerCount = len(owners)
	summary.Diagnostics = resolver.Diagnostics()

	// Derive coverage posture.
	summary.CoveragePosture = deriveCoveragePosture(summary)

	return summary
}

// deriveCoveragePosture classifies ownership coverage.
func deriveCoveragePosture(s OwnershipSummary) string {
	if s.TotalFiles == 0 {
		return "none"
	}

	ratio := float64(s.OwnedFiles) / float64(s.TotalFiles)
	switch {
	case ratio >= 0.80:
		return "strong"
	case ratio >= 0.50:
		return "partial"
	case ratio > 0:
		return "weak"
	default:
		return "none"
	}
}
