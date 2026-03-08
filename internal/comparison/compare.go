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

	// TestCaseDeltas summarizes changes to individual test cases.
	TestCaseDeltas *TestCaseDeltas `json:"testCaseDeltas,omitempty"`

	// CoverageDelta summarizes changes to coverage metrics.
	CoverageDelta *CoverageDelta `json:"coverageDelta,omitempty"`
}

// TestCaseDeltas summarizes changes in test case identity across snapshots.
type TestCaseDeltas struct {
	// Added is the number of new test cases (IDs in to but not from).
	Added int `json:"added"`

	// Removed is the number of removed test cases (IDs in from but not to).
	Removed int `json:"removed"`

	// Stable is the number of unchanged test cases (IDs in both).
	Stable int `json:"stable"`

	// AddedExamples are representative new test names (up to 5).
	AddedExamples []string `json:"addedExamples,omitempty"`

	// RemovedExamples are representative removed test names (up to 5).
	RemovedExamples []string `json:"removedExamples,omitempty"`
}

// CoverageDelta summarizes changes to coverage metrics across snapshots.
type CoverageDelta struct {
	// LineCoverageBefore is the previous line coverage percentage.
	LineCoverageBefore float64 `json:"lineCoverageBefore"`

	// LineCoverageAfter is the current line coverage percentage.
	LineCoverageAfter float64 `json:"lineCoverageAfter"`

	// LineCoverageDelta is the change in line coverage percentage.
	LineCoverageDelta float64 `json:"lineCoverageDelta"`

	// UncoveredExportedBefore is the previous count of uncovered exports.
	UncoveredExportedBefore int `json:"uncoveredExportedBefore"`

	// UncoveredExportedAfter is the current count of uncovered exports.
	UncoveredExportedAfter int `json:"uncoveredExportedAfter"`

	// CoveredOnlyByE2EBefore is the previous count of e2e-only coverage.
	CoveredOnlyByE2EBefore int `json:"coveredOnlyByE2eBefore"`

	// CoveredOnlyByE2EAfter is the current count of e2e-only coverage.
	CoveredOnlyByE2EAfter int `json:"coveredOnlyByE2eAfter"`

	// UnitTestCoverageBefore is the previous count of units covered by unit tests.
	UnitTestCoverageBefore int `json:"unitTestCoverageBefore,omitempty"`

	// UnitTestCoverageAfter is the current count of units covered by unit tests.
	UnitTestCoverageAfter int `json:"unitTestCoverageAfter,omitempty"`
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
	comp.TestCaseDeltas = compareTestCases(from.TestCases, to.TestCases)
	comp.CoverageDelta = compareCoverage(from.CoverageSummary, to.CoverageSummary)

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
	if c.TestCaseDeltas != nil && (c.TestCaseDeltas.Added > 0 || c.TestCaseDeltas.Removed > 0) {
		return true
	}
	if c.CoverageDelta != nil && c.CoverageDelta.LineCoverageDelta != 0 {
		return true
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

	// Sort types for deterministic output.
	sortedTypes := make([]models.SignalType, 0, len(allTypes))
	for t := range allTypes {
		sortedTypes = append(sortedTypes, t)
	}
	sort.Slice(sortedTypes, func(i, j int) bool {
		return sortedTypes[i] < sortedTypes[j]
	})

	var deltas []SignalDelta
	for _, t := range sortedTypes {
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

	// Sort by absolute delta descending, then by type for determinism.
	sort.Slice(deltas, func(i, j int) bool {
		ai := deltas[i].Delta
		if ai < 0 {
			ai = -ai
		}
		aj := deltas[j].Delta
		if aj < 0 {
			aj = -aj
		}
		if ai != aj {
			return ai > aj
		}
		return deltas[i].Type < deltas[j].Type
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

	// Sort keys for deterministic output.
	sortedKeys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	var deltas []RiskDelta
	for _, key := range sortedKeys {
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
	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Change != changes[j].Change {
			return changes[i].Change < changes[j].Change
		}
		return changes[i].Name < changes[j].Name
	})
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

func compareTestCases(from, to []models.TestCase) *TestCaseDeltas {
	if len(from) == 0 && len(to) == 0 {
		return nil
	}

	fromIDs := map[string]string{} // testID → testName
	for _, tc := range from {
		fromIDs[tc.TestID] = tc.TestName
	}
	toIDs := map[string]string{}
	for _, tc := range to {
		toIDs[tc.TestID] = tc.TestName
	}

	d := &TestCaseDeltas{}
	for id, name := range toIDs {
		if _, ok := fromIDs[id]; ok {
			d.Stable++
		} else {
			d.Added++
			if len(d.AddedExamples) < 5 {
				d.AddedExamples = append(d.AddedExamples, name)
			}
		}
	}
	for id, name := range fromIDs {
		if _, ok := toIDs[id]; !ok {
			d.Removed++
			if len(d.RemovedExamples) < 5 {
				d.RemovedExamples = append(d.RemovedExamples, name)
			}
		}
	}

	// Sort examples for determinism.
	sort.Strings(d.AddedExamples)
	sort.Strings(d.RemovedExamples)

	return d
}

func compareCoverage(from, to *models.CoverageSummary) *CoverageDelta {
	if from == nil && to == nil {
		return nil
	}

	d := &CoverageDelta{}
	if from != nil {
		d.LineCoverageBefore = from.LineCoveragePct
		d.UncoveredExportedBefore = from.UncoveredExported
		d.CoveredOnlyByE2EBefore = from.CoveredOnlyByE2E
		d.UnitTestCoverageBefore = from.CoveredByUnitTests
	}
	if to != nil {
		d.LineCoverageAfter = to.LineCoveragePct
		d.UncoveredExportedAfter = to.UncoveredExported
		d.CoveredOnlyByE2EAfter = to.CoveredOnlyByE2E
		d.UnitTestCoverageAfter = to.CoveredByUnitTests
	}
	d.LineCoverageDelta = d.LineCoverageAfter - d.LineCoverageBefore

	return d
}
