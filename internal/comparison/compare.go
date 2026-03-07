// Package comparison implements snapshot-to-snapshot comparison
// for local trend detection.
//
// Comparison works at the aggregate level: signal counts, risk bands,
// and framework changes. It does not attempt perfect per-signal identity
// matching — meaningful aggregate deltas are more useful than fragile diffs.
package comparison

import (
	"sort"

	"github.com/pmclSF/hamlet/internal/models"
)

// SnapshotComparison holds the result of comparing two snapshots.
type SnapshotComparison struct {
	// FromTime and ToTime identify the compared snapshots.
	FromTime string `json:"fromTime"`
	ToTime   string `json:"toTime"`

	// SignalDeltas shows how signal counts changed by type.
	SignalDeltas []SignalDelta `json:"signalDeltas,omitempty"`

	// RiskDeltas shows how risk surfaces changed.
	RiskDeltas []RiskDelta `json:"riskDeltas,omitempty"`

	// FrameworkChanges notes added/removed frameworks.
	FrameworkChanges []FrameworkChange `json:"frameworkChanges,omitempty"`

	// TestFileCountDelta is the change in total test file count.
	TestFileCountDelta int `json:"testFileCountDelta"`

	// NewSignalExamples shows representative new signals (up to 5).
	NewSignalExamples []SignalExample `json:"newSignalExamples,omitempty"`

	// ResolvedSignalExamples shows representative resolved signals (up to 5).
	ResolvedSignalExamples []SignalExample `json:"resolvedSignalExamples,omitempty"`
}

// SignalDelta represents the change in count for a signal type.
type SignalDelta struct {
	Type     models.SignalType     `json:"type"`
	Category models.SignalCategory `json:"category"`
	Before   int                   `json:"before"`
	After    int                   `json:"after"`
	Delta    int                   `json:"delta"` // positive = increased
}

// RiskDelta represents a change in a risk surface.
type RiskDelta struct {
	Type      string          `json:"type"`
	Scope     string          `json:"scope"`
	ScopeName string          `json:"scopeName"`
	Before    models.RiskBand `json:"before"`
	After     models.RiskBand `json:"after"`
	Changed   bool            `json:"changed"`
}

// FrameworkChange notes a framework added or removed.
type FrameworkChange struct {
	Name    string `json:"name"`
	Change  string `json:"change"` // "added" or "removed"
	Files   int    `json:"files"`
}

// SignalExample is a representative signal for display in comparison output.
type SignalExample struct {
	Type        models.SignalType `json:"type"`
	File        string            `json:"file,omitempty"`
	Explanation string            `json:"explanation"`
}

// Compare produces a SnapshotComparison between two snapshots.
//
// The "from" snapshot is the older/baseline, "to" is the current.
func Compare(from, to *models.TestSuiteSnapshot) *SnapshotComparison {
	comp := &SnapshotComparison{
		FromTime:           from.GeneratedAt.Format("2006-01-02 15:04:05 UTC"),
		ToTime:             to.GeneratedAt.Format("2006-01-02 15:04:05 UTC"),
		TestFileCountDelta: len(to.TestFiles) - len(from.TestFiles),
	}

	comp.SignalDeltas = compareSignals(from.Signals, to.Signals)
	comp.RiskDeltas = compareRisk(from.Risk, to.Risk)
	comp.FrameworkChanges = compareFrameworks(from.Frameworks, to.Frameworks)
	comp.NewSignalExamples, comp.ResolvedSignalExamples = findRepresentativeChanges(from.Signals, to.Signals)

	return comp
}

// HasMeaningfulChanges returns true if the comparison contains any notable changes.
func (c *SnapshotComparison) HasMeaningfulChanges() bool {
	for _, d := range c.SignalDeltas {
		if d.Delta != 0 {
			return true
		}
	}
	for _, r := range c.RiskDeltas {
		if r.Changed {
			return true
		}
	}
	return len(c.FrameworkChanges) > 0 || c.TestFileCountDelta != 0
}

func compareSignals(from, to []models.Signal) []SignalDelta {
	fromCounts := countByType(from)
	toCounts := countByType(to)

	// Collect all types
	allTypes := map[models.SignalType]bool{}
	for t := range fromCounts {
		allTypes[t] = true
	}
	for t := range toCounts {
		allTypes[t] = true
	}

	var deltas []SignalDelta
	for t := range allTypes {
		before := fromCounts[t]
		after := toCounts[t]
		if before != after {
			cat := findCategory(from, to, t)
			deltas = append(deltas, SignalDelta{
				Type:     t,
				Category: cat,
				Before:   before,
				After:    after,
				Delta:    after - before,
			})
		}
	}

	// Sort by absolute delta descending
	sort.Slice(deltas, func(i, j int) bool {
		ai := deltas[i].Delta
		if ai < 0 {
			ai = -ai
		}
		aj := deltas[j].Delta
		if aj < 0 {
			aj = -aj
		}
		return ai > aj
	})

	return deltas
}

func compareRisk(from, to []models.RiskSurface) []RiskDelta {
	fromMap := map[string]models.RiskSurface{}
	for _, r := range from {
		key := r.Type + ":" + r.Scope + ":" + r.ScopeName
		fromMap[key] = r
	}

	toMap := map[string]models.RiskSurface{}
	for _, r := range to {
		key := r.Type + ":" + r.Scope + ":" + r.ScopeName
		toMap[key] = r
	}

	// All keys
	allKeys := map[string]bool{}
	for k := range fromMap {
		allKeys[k] = true
	}
	for k := range toMap {
		allKeys[k] = true
	}

	var deltas []RiskDelta
	for key := range allKeys {
		fromR := fromMap[key]
		toR := toMap[key]

		// Use the non-empty surface for metadata
		ref := toR
		if ref.Type == "" {
			ref = fromR
		}

		deltas = append(deltas, RiskDelta{
			Type:      ref.Type,
			Scope:     ref.Scope,
			ScopeName: ref.ScopeName,
			Before:    fromR.Band,
			After:     toR.Band,
			Changed:   fromR.Band != toR.Band,
		})
	}

	// Sort: changed first, then by type
	sort.Slice(deltas, func(i, j int) bool {
		if deltas[i].Changed != deltas[j].Changed {
			return deltas[i].Changed
		}
		return deltas[i].Type < deltas[j].Type
	})

	return deltas
}

func compareFrameworks(from, to []models.Framework) []FrameworkChange {
	fromSet := map[string]int{}
	for _, fw := range from {
		fromSet[fw.Name] = fw.FileCount
	}
	toSet := map[string]int{}
	for _, fw := range to {
		toSet[fw.Name] = fw.FileCount
	}

	var changes []FrameworkChange
	for name, files := range toSet {
		if _, existed := fromSet[name]; !existed {
			changes = append(changes, FrameworkChange{Name: name, Change: "added", Files: files})
		}
	}
	for name, files := range fromSet {
		if _, exists := toSet[name]; !exists {
			changes = append(changes, FrameworkChange{Name: name, Change: "removed", Files: files})
		}
	}
	return changes
}

func findRepresentativeChanges(from, to []models.Signal) (newExamples, resolvedExamples []SignalExample) {
	// Build sets of signal keys for rough matching
	fromKeys := map[string]bool{}
	for _, s := range from {
		fromKeys[string(s.Type)+":"+s.Location.File] = true
	}
	toKeys := map[string]bool{}
	for _, s := range to {
		toKeys[string(s.Type)+":"+s.Location.File] = true
	}

	// Find new signals (in to but not from)
	for _, s := range to {
		key := string(s.Type) + ":" + s.Location.File
		if !fromKeys[key] && len(newExamples) < 5 {
			newExamples = append(newExamples, SignalExample{
				Type:        s.Type,
				File:        s.Location.File,
				Explanation: s.Explanation,
			})
		}
	}

	// Find resolved signals (in from but not to)
	for _, s := range from {
		key := string(s.Type) + ":" + s.Location.File
		if !toKeys[key] && len(resolvedExamples) < 5 {
			resolvedExamples = append(resolvedExamples, SignalExample{
				Type:        s.Type,
				File:        s.Location.File,
				Explanation: s.Explanation,
			})
		}
	}

	return
}

func countByType(signals []models.Signal) map[models.SignalType]int {
	counts := map[models.SignalType]int{}
	for _, s := range signals {
		counts[s.Type]++
	}
	return counts
}

func findCategory(from, to []models.Signal, t models.SignalType) models.SignalCategory {
	for _, s := range to {
		if s.Type == t {
			return s.Category
		}
	}
	for _, s := range from {
		if s.Type == t {
			return s.Category
		}
	}
	return ""
}
