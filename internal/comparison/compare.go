// Package comparison implements snapshot-to-snapshot comparison
// for local trend detection.
//
// Comparison works at the aggregate level: signal counts, risk bands,
// and framework changes. It does not attempt perfect per-signal identity
// matching — meaningful aggregate deltas are more useful than fragile diffs.
package comparison

import (
	"sort"
	"strings"

	"github.com/pmclSF/hamlet/internal/lifecycle"
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

	// OwnershipDelta summarizes changes to ownership coverage.
	OwnershipDelta *OwnershipDelta `json:"ownershipDelta,omitempty"`

	// PostureDeltas summarizes changes to posture dimensions.
	PostureDeltas []PostureDelta `json:"postureDeltas,omitempty"`

	// MeasurementDeltas summarizes changes to individual measurements.
	MeasurementDeltas []MeasurementDelta `json:"measurementDeltas,omitempty"`

	// LifecycleContinuity holds test lifecycle analysis between the two snapshots.
	LifecycleContinuity *lifecycle.ContinuityResult `json:"lifecycleContinuity,omitempty"`

	// MethodologyCompatible indicates whether snapshots are directly comparable
	// for methodology-sensitive deltas (risk/posture/measurements).
	MethodologyCompatible bool `json:"methodologyCompatible"`

	// MethodologyNotes explains methodology compatibility decisions.
	MethodologyNotes []string `json:"methodologyNotes,omitempty"`
}

// OwnershipDelta summarizes changes to ownership metrics across snapshots.
type OwnershipDelta struct {
	// OwnerCountBefore is the number of distinct owners before.
	OwnerCountBefore int `json:"ownerCountBefore"`

	// OwnerCountAfter is the number of distinct owners after.
	OwnerCountAfter int `json:"ownerCountAfter"`

	// OwnedFilesBefore is the count of files with ownership before.
	OwnedFilesBefore int `json:"ownedFilesBefore"`

	// OwnedFilesAfter is the count of files with ownership after.
	OwnedFilesAfter int `json:"ownedFilesAfter"`

	// OwnershipImproved is true if ownership coverage increased.
	OwnershipImproved bool `json:"ownershipImproved"`
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

// PostureDelta represents a change in a posture dimension band.
type PostureDelta struct {
	// Dimension is the posture dimension name (e.g. "health").
	Dimension string `json:"dimension"`

	// Before is the band in the baseline snapshot.
	Before string `json:"before"`

	// After is the band in the current snapshot.
	After string `json:"after"`

	// Changed is true if the band changed.
	Changed bool `json:"changed"`
}

// MeasurementDelta represents a change in an individual measurement.
type MeasurementDelta struct {
	// ID is the measurement identifier (e.g. "health.flaky_share").
	ID string `json:"id"`

	// Dimension is the posture dimension this measurement feeds.
	Dimension string `json:"dimension"`

	// Before is the measurement value in the baseline snapshot.
	Before float64 `json:"before"`

	// After is the measurement value in the current snapshot.
	After float64 `json:"after"`

	// Delta is the change in value (After - Before).
	Delta float64 `json:"delta"`

	// BandBefore is the qualitative band before.
	BandBefore string `json:"bandBefore,omitempty"`

	// BandAfter is the qualitative band after.
	BandAfter string `json:"bandAfter,omitempty"`

	// BandChanged is true if the band changed.
	BandChanged bool `json:"bandChanged,omitempty"`
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
	Name   string `json:"name"`
	Change string `json:"change"` // "added" or "removed"
	Files  int    `json:"files"`
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
	if from == nil || to == nil {
		return &SnapshotComparison{
			MethodologyCompatible: false,
			MethodologyNotes:      []string{"cannot compare nil snapshot input"},
		}
	}
	models.MigrateSnapshotInPlace(from)
	models.MigrateSnapshotInPlace(to)

	fromTime := "unknown"
	toTime := "unknown"
	if !from.GeneratedAt.IsZero() {
		fromTime = from.GeneratedAt.Format("2006-01-02 15:04:05 UTC")
	}
	if !to.GeneratedAt.IsZero() {
		toTime = to.GeneratedAt.Format("2006-01-02 15:04:05 UTC")
	}

	methodologyCompatible, methodologyNotes := assessMethodologyCompatibility(from, to)
	comp := &SnapshotComparison{
		FromTime:              fromTime,
		ToTime:                toTime,
		TestFileCountDelta:    len(to.TestFiles) - len(from.TestFiles),
		MethodologyCompatible: methodologyCompatible,
		MethodologyNotes:      methodologyNotes,
	}

	comp.SignalDeltas = compareSignals(from.Signals, to.Signals)
	if methodologyCompatible {
		comp.RiskDeltas = compareRisk(from.Risk, to.Risk)
	}
	comp.FrameworkChanges = compareFrameworks(from.Frameworks, to.Frameworks)
	comp.NewSignalExamples, comp.ResolvedSignalExamples = findRepresentativeChanges(from.Signals, to.Signals)
	comp.TestCaseDeltas = compareTestCases(from.TestCases, to.TestCases)
	comp.CoverageDelta = compareCoverage(from.CoverageSummary, to.CoverageSummary)
	comp.OwnershipDelta = compareOwnership(from.Ownership, to.Ownership)
	if methodologyCompatible {
		comp.PostureDeltas, comp.MeasurementDeltas = compareMeasurements(from.Measurements, to.Measurements)
	}
	comp.LifecycleContinuity = lifecycle.InferContinuity(from, to)

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
	if len(c.PostureDeltas) > 0 {
		return true
	}
	return len(c.FrameworkChanges) > 0 || c.TestFileCountDelta != 0
}

func assessMethodologyCompatibility(from, to *models.TestSuiteSnapshot) (bool, []string) {
	var notes []string
	if from == nil || to == nil {
		return false, []string{"cannot assess methodology compatibility with nil snapshot input"}
	}
	if from.SnapshotMeta.SchemaVersion != "" && to.SnapshotMeta.SchemaVersion != "" &&
		from.SnapshotMeta.SchemaVersion != to.SnapshotMeta.SchemaVersion {
		notes = append(notes, "snapshot schema versions differ; methodology-sensitive deltas were suppressed")
		return false, notes
	}

	fromFP := strings.TrimSpace(from.SnapshotMeta.MethodologyFingerprint)
	toFP := strings.TrimSpace(to.SnapshotMeta.MethodologyFingerprint)
	if fromFP == "" || toFP == "" {
		notes = append(notes, "methodology fingerprint missing on one or both snapshots; compatibility assumed for backward compatibility")
		return true, notes
	}
	if fromFP != toFP {
		notes = append(notes, "methodology fingerprint differs; risk/posture/measurement deltas were suppressed")
		return false, notes
	}
	return true, nil
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
	// Build multi-sets of signal identities for precise matching.
	fromCounts := map[string]int{}
	for _, s := range from {
		fromCounts[signalIdentityKey(s)]++
	}
	for _, s := range to {
		key := signalIdentityKey(s)
		if fromCounts[key] > 0 {
			fromCounts[key]--
			continue
		}
		if len(newExamples) < 5 {
			newExamples = append(newExamples, SignalExample{
				Type:        s.Type,
				File:        s.Location.File,
				Explanation: s.Explanation,
			})
		}
	}

	toCounts := map[string]int{}
	for _, s := range to {
		toCounts[signalIdentityKey(s)]++
	}
	for _, s := range from {
		key := signalIdentityKey(s)
		if toCounts[key] > 0 {
			toCounts[key]--
			continue
		}
		if len(resolvedExamples) < 5 {
			resolvedExamples = append(resolvedExamples, SignalExample{
				Type:        s.Type,
				File:        s.Location.File,
				Explanation: s.Explanation,
			})
		}
	}

	return
}

func signalIdentityKey(s models.Signal) string {
	explanation := strings.TrimSpace(strings.Join(strings.Fields(s.Explanation), " "))
	if len(explanation) > 120 {
		explanation = explanation[:120]
	}
	return strings.Join([]string{
		string(s.Type),
		s.Location.Repository,
		s.Location.Package,
		s.Location.File,
		s.Location.Symbol,
		explanation,
	}, "|")
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

func compareMeasurements(from, to *models.MeasurementSnapshot) ([]PostureDelta, []MeasurementDelta) {
	if from == nil && to == nil {
		return nil, nil
	}

	// Build posture deltas.
	fromPosture := map[string]string{}
	if from != nil {
		for _, p := range from.Posture {
			fromPosture[p.Dimension] = p.Band
		}
	}
	toPosture := map[string]string{}
	if to != nil {
		for _, p := range to.Posture {
			toPosture[p.Dimension] = p.Band
		}
	}

	allDims := map[string]bool{}
	for d := range fromPosture {
		allDims[d] = true
	}
	for d := range toPosture {
		allDims[d] = true
	}

	sortedDims := make([]string, 0, len(allDims))
	for d := range allDims {
		sortedDims = append(sortedDims, d)
	}
	sort.Strings(sortedDims)

	var postureDeltas []PostureDelta
	for _, dim := range sortedDims {
		before := fromPosture[dim]
		after := toPosture[dim]
		if before != after {
			postureDeltas = append(postureDeltas, PostureDelta{
				Dimension: dim,
				Before:    before,
				After:     after,
				Changed:   true,
			})
		}
	}

	// Build measurement deltas.
	fromMeas := map[string]models.MeasurementResult{}
	if from != nil {
		for _, m := range from.Measurements {
			fromMeas[m.ID] = m
		}
	}
	toMeas := map[string]models.MeasurementResult{}
	if to != nil {
		for _, m := range to.Measurements {
			toMeas[m.ID] = m
		}
	}

	allIDs := map[string]bool{}
	for id := range fromMeas {
		allIDs[id] = true
	}
	for id := range toMeas {
		allIDs[id] = true
	}

	sortedIDs := make([]string, 0, len(allIDs))
	for id := range allIDs {
		sortedIDs = append(sortedIDs, id)
	}
	sort.Strings(sortedIDs)

	var measDeltas []MeasurementDelta
	for _, id := range sortedIDs {
		fm := fromMeas[id]
		tm := toMeas[id]
		delta := tm.Value - fm.Value
		if delta == 0 && fm.Band == tm.Band {
			continue
		}
		dim := tm.Dimension
		if dim == "" {
			dim = fm.Dimension
		}
		measDeltas = append(measDeltas, MeasurementDelta{
			ID:          id,
			Dimension:   dim,
			Before:      fm.Value,
			After:       tm.Value,
			Delta:       delta,
			BandBefore:  fm.Band,
			BandAfter:   tm.Band,
			BandChanged: fm.Band != tm.Band,
		})
	}

	// Sort by absolute delta descending, then by ID.
	sort.Slice(measDeltas, func(i, j int) bool {
		ai := measDeltas[i].Delta
		if ai < 0 {
			ai = -ai
		}
		aj := measDeltas[j].Delta
		if aj < 0 {
			aj = -aj
		}
		if ai != aj {
			return ai > aj
		}
		return measDeltas[i].ID < measDeltas[j].ID
	})

	return postureDeltas, measDeltas
}

func compareOwnership(from, to map[string][]string) *OwnershipDelta {
	if len(from) == 0 && len(to) == 0 {
		return nil
	}

	fromOwners := map[string]bool{}
	for _, owners := range from {
		for _, o := range owners {
			fromOwners[o] = true
		}
	}
	toOwners := map[string]bool{}
	for _, owners := range to {
		for _, o := range owners {
			toOwners[o] = true
		}
	}

	d := &OwnershipDelta{
		OwnerCountBefore:  len(fromOwners),
		OwnerCountAfter:   len(toOwners),
		OwnedFilesBefore:  len(from),
		OwnedFilesAfter:   len(to),
		OwnershipImproved: len(to) > len(from),
	}

	return d
}
