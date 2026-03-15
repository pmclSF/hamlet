// Package graph provides an in-memory analysis graph that indexes
// TestSuiteSnapshot data for cross-cutting queries.
//
// The graph is NOT a persistent store — it is built on demand from a
// snapshot and provides O(1) lookups across tests, code units, owners,
// files, signals, and coverage links that would otherwise require
// repeated linear scans.
//
// Design constraints:
//   - Lightweight: no graph database, no external dependencies
//   - Read-only after construction: Build() then query
//   - Deterministic: iteration order is defined where it matters
//   - No circular imports: depends only on models
package graph

import (
	"path/filepath"
	"sort"

	"github.com/pmclSF/terrain/internal/models"
)

// Graph is the normalized analysis graph built from a TestSuiteSnapshot.
//
// All indexes are populated during Build(). After construction the graph
// is read-only and safe for concurrent reads.
type Graph struct {
	// Source snapshot.
	snap *models.TestSuiteSnapshot

	// --- Test identity indexes ---

	// TestByID maps TestCase.TestID → *TestCase.
	TestByID map[string]*models.TestCase

	// TestsByFile maps file path → test case IDs in that file.
	TestsByFile map[string][]string

	// TestsByOwner maps owner → test case IDs owned by that owner.
	TestsByOwner map[string][]string

	// TestsByType maps test type (unit, integration, e2e) → test case IDs.
	TestsByType map[string][]string

	// --- Code unit indexes ---

	// UnitByID maps CodeUnit.UnitID → *CodeUnit.
	UnitByID map[string]*models.CodeUnit

	// UnitsByFile maps file path → unit IDs in that file.
	UnitsByFile map[string][]string

	// UnitsByOwner maps owner → unit IDs owned by that owner.
	UnitsByOwner map[string][]string

	// ExportedUnits lists unit IDs for exported code units.
	ExportedUnits []string

	// --- Signal indexes ---

	// SignalsByFile maps file path → signals for that file.
	SignalsByFile map[string][]models.Signal

	// SignalsByOwner maps owner → signals for that owner.
	SignalsByOwner map[string][]models.Signal

	// SignalsByType maps signal type → signals of that type.
	SignalsByType map[models.SignalType][]models.Signal

	// HealthSignalsByTestID maps test ID → health signals referencing that test.
	// Populated when health signals carry testId metadata.
	HealthSignalsByTestID map[string][]models.Signal

	// --- File/directory indexes ---

	// FileOwner maps file path → owner.
	FileOwner map[string]string

	// DirectoryFiles maps directory → file paths.
	DirectoryFiles map[string][]string

	// --- Coverage indexes ---

	// UncoveredExportedUnits lists exported unit IDs with no coverage.
	UncoveredExportedUnits []string

	// E2EOnlyUnits lists unit IDs covered only by e2e tests.
	E2EOnlyUnits []string

	// UnitCoverageType maps unit ID → set of test types that cover it.
	// Populated from CoverageSummary and code unit coverage data.
	UnitCoverageType map[string]map[string]bool
}

// Build constructs a Graph from a snapshot. The graph is read-only after
// this call. Build is O(n) in the size of the snapshot data.
func Build(snap *models.TestSuiteSnapshot) *Graph {
	g := &Graph{
		snap:                  snap,
		TestByID:              make(map[string]*models.TestCase, len(snap.TestCases)),
		TestsByFile:           make(map[string][]string),
		TestsByOwner:          make(map[string][]string),
		TestsByType:           make(map[string][]string),
		UnitByID:              make(map[string]*models.CodeUnit, len(snap.CodeUnits)),
		UnitsByFile:           make(map[string][]string),
		UnitsByOwner:          make(map[string][]string),
		SignalsByFile:         make(map[string][]models.Signal),
		SignalsByOwner:        make(map[string][]models.Signal),
		SignalsByType:         make(map[models.SignalType][]models.Signal),
		HealthSignalsByTestID: make(map[string][]models.Signal),
		FileOwner:             make(map[string]string),
		DirectoryFiles:        make(map[string][]string),
		UnitCoverageType:      make(map[string]map[string]bool),
	}

	g.indexTests()
	g.indexCodeUnits()
	g.indexSignals()
	g.indexFiles()
	g.indexCoverage()

	return g
}

// Snapshot returns the underlying snapshot.
func (g *Graph) Snapshot() *models.TestSuiteSnapshot {
	return g.snap
}

func (g *Graph) indexTests() {
	// Build file → owner map from test files.
	fileOwners := make(map[string]string, len(g.snap.TestFiles))
	for i := range g.snap.TestFiles {
		tf := &g.snap.TestFiles[i]
		if tf.Owner != "" {
			fileOwners[tf.Path] = tf.Owner
		}
	}

	for i := range g.snap.TestCases {
		tc := &g.snap.TestCases[i]
		g.TestByID[tc.TestID] = tc
		g.TestsByFile[tc.FilePath] = append(g.TestsByFile[tc.FilePath], tc.TestID)

		if tc.TestType != "" {
			g.TestsByType[tc.TestType] = append(g.TestsByType[tc.TestType], tc.TestID)
		}

		owner := fileOwners[tc.FilePath]
		if owner == "" {
			owner = "unknown"
		}
		g.TestsByOwner[owner] = append(g.TestsByOwner[owner], tc.TestID)
	}
}

func (g *Graph) indexCodeUnits() {
	for i := range g.snap.CodeUnits {
		cu := &g.snap.CodeUnits[i]
		if cu.UnitID == "" {
			continue
		}
		g.UnitByID[cu.UnitID] = cu
		g.UnitsByFile[cu.Path] = append(g.UnitsByFile[cu.Path], cu.UnitID)

		owner := cu.Owner
		if owner == "" {
			owner = "unknown"
		}
		g.UnitsByOwner[owner] = append(g.UnitsByOwner[owner], cu.UnitID)

		if cu.Exported {
			g.ExportedUnits = append(g.ExportedUnits, cu.UnitID)
			if cu.Coverage <= 0 {
				g.UncoveredExportedUnits = append(g.UncoveredExportedUnits, cu.UnitID)
			}
		}
	}
	sort.Strings(g.ExportedUnits)
	sort.Strings(g.UncoveredExportedUnits)
}

func (g *Graph) indexSignals() {
	healthTypes := map[models.SignalType]bool{
		"slowTest":    true,
		"flakyTest":   true,
		"skippedTest": true,
	}

	for _, s := range g.snap.Signals {
		file := s.Location.File
		if file != "" {
			g.SignalsByFile[file] = append(g.SignalsByFile[file], s)
		}

		owner := s.Owner
		if owner == "" {
			owner = "unknown"
		}
		g.SignalsByOwner[owner] = append(g.SignalsByOwner[owner], s)
		g.SignalsByType[s.Type] = append(g.SignalsByType[s.Type], s)

		// Index health signals by test ID if available.
		if healthTypes[s.Type] {
			if testID, ok := s.Metadata["testId"].(string); ok && testID != "" {
				g.HealthSignalsByTestID[testID] = append(g.HealthSignalsByTestID[testID], s)
			}
		}
	}
}

func (g *Graph) indexFiles() {
	seen := make(map[string]bool)

	// Index test files.
	for _, tf := range g.snap.TestFiles {
		if tf.Owner != "" {
			g.FileOwner[tf.Path] = tf.Owner
		}
		dir := filepath.Dir(tf.Path)
		if !seen[tf.Path] {
			g.DirectoryFiles[dir] = append(g.DirectoryFiles[dir], tf.Path)
			seen[tf.Path] = true
		}
	}

	// Index code unit files.
	for _, cu := range g.snap.CodeUnits {
		if cu.Owner != "" && g.FileOwner[cu.Path] == "" {
			g.FileOwner[cu.Path] = cu.Owner
		}
		dir := filepath.Dir(cu.Path)
		if !seen[cu.Path] {
			g.DirectoryFiles[dir] = append(g.DirectoryFiles[dir], cu.Path)
			seen[cu.Path] = true
		}
	}
}

func (g *Graph) indexCoverage() {
	// Derive e2e-only units from CoverageSummary and code unit data.
	// If we have CoverageInsights, use them to identify e2e-only units.
	for _, insight := range g.snap.CoverageInsights {
		if insight.Type == "e2e_only_coverage" && insight.UnitID != "" {
			g.E2EOnlyUnits = append(g.E2EOnlyUnits, insight.UnitID)
		}
	}
	sort.Strings(g.E2EOnlyUnits)
}

// --- Query methods ---

// TestsInModule returns test IDs for tests in files under the given directory.
func (g *Graph) TestsInModule(dir string) []string {
	var ids []string
	for file, testIDs := range g.TestsByFile {
		if filepath.Dir(file) == dir || hasPrefix(file, dir) {
			ids = append(ids, testIDs...)
		}
	}
	sort.Strings(ids)
	return ids
}

// UnitsForOwner returns code unit IDs owned by the given owner.
func (g *Graph) UnitsForOwner(owner string) []string {
	return g.UnitsByOwner[owner]
}

// UncoveredExportedForOwner returns exported unit IDs with no coverage
// that belong to the given owner.
func (g *Graph) UncoveredExportedForOwner(owner string) []string {
	ownerUnits := make(map[string]bool, len(g.UnitsByOwner[owner]))
	for _, uid := range g.UnitsByOwner[owner] {
		ownerUnits[uid] = true
	}

	var result []string
	for _, uid := range g.UncoveredExportedUnits {
		if ownerUnits[uid] {
			result = append(result, uid)
		}
	}
	return result
}

// E2EOnlyForOwner returns unit IDs covered only by e2e that belong to the owner.
func (g *Graph) E2EOnlyForOwner(owner string) []string {
	ownerUnits := make(map[string]bool, len(g.UnitsByOwner[owner]))
	for _, uid := range g.UnitsByOwner[owner] {
		ownerUnits[uid] = true
	}

	var result []string
	for _, uid := range g.E2EOnlyUnits {
		if ownerUnits[uid] {
			result = append(result, uid)
		}
	}
	return result
}

// HealthSignalsForOwner returns health-category signals for tests owned
// by the given owner.
func (g *Graph) HealthSignalsForOwner(owner string) []models.Signal {
	var result []models.Signal
	for _, s := range g.SignalsByOwner[owner] {
		if s.Category == models.CategoryHealth {
			result = append(result, s)
		}
	}
	return result
}

// TopFailingTestIDs returns up to n test IDs with the most health signals,
// sorted by signal count descending.
func (g *Graph) TopFailingTestIDs(n int) []TestHealthSummary {
	var summaries []TestHealthSummary
	for testID, sigs := range g.HealthSignalsByTestID {
		tc := g.TestByID[testID]
		name := testID
		file := ""
		if tc != nil {
			name = tc.TestName
			file = tc.FilePath
		}
		summaries = append(summaries, TestHealthSummary{
			TestID:      testID,
			TestName:    name,
			FilePath:    file,
			SignalCount: len(sigs),
			Signals:     sigs,
		})
	}
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].SignalCount != summaries[j].SignalCount {
			return summaries[i].SignalCount > summaries[j].SignalCount
		}
		return summaries[i].TestID < summaries[j].TestID
	})
	if len(summaries) > n {
		summaries = summaries[:n]
	}
	return summaries
}

// TestHealthSummary describes health signal concentration for a single test.
type TestHealthSummary struct {
	TestID      string          `json:"testId"`
	TestName    string          `json:"testName"`
	FilePath    string          `json:"filePath"`
	SignalCount int             `json:"signalCount"`
	Signals     []models.Signal `json:"signals,omitempty"`
}

// OwnerRiskSummary aggregates risk data for a single owner.
type OwnerRiskSummary struct {
	Owner                string `json:"owner"`
	TotalSignals         int    `json:"totalSignals"`
	HealthSignals        int    `json:"healthSignals"`
	QualitySignals       int    `json:"qualitySignals"`
	UncoveredExported    int    `json:"uncoveredExported"`
	E2EOnlyUnits         int    `json:"e2eOnlyUnits"`
	TestCount            int    `json:"testCount"`
	CodeUnitCount        int    `json:"codeUnitCount"`
}

// OwnerRiskSummaries returns per-owner risk aggregations, sorted by total
// signal count descending.
func (g *Graph) OwnerRiskSummaries() []OwnerRiskSummary {
	owners := make(map[string]*OwnerRiskSummary)

	ensureOwner := func(o string) *OwnerRiskSummary {
		if s, ok := owners[o]; ok {
			return s
		}
		s := &OwnerRiskSummary{Owner: o}
		owners[o] = s
		return s
	}

	// Signals by owner.
	for owner, sigs := range g.SignalsByOwner {
		s := ensureOwner(owner)
		s.TotalSignals = len(sigs)
		for _, sig := range sigs {
			switch sig.Category {
			case models.CategoryHealth:
				s.HealthSignals++
			case models.CategoryQuality:
				s.QualitySignals++
			}
		}
	}

	// Tests by owner.
	for owner, testIDs := range g.TestsByOwner {
		ensureOwner(owner).TestCount = len(testIDs)
	}

	// Code units by owner.
	for owner, unitIDs := range g.UnitsByOwner {
		ensureOwner(owner).CodeUnitCount = len(unitIDs)
	}

	// Uncovered exports by owner.
	for _, uid := range g.UncoveredExportedUnits {
		cu := g.UnitByID[uid]
		if cu == nil {
			continue
		}
		owner := cu.Owner
		if owner == "" {
			owner = "unknown"
		}
		ensureOwner(owner).UncoveredExported++
	}

	// E2E-only by owner.
	for _, uid := range g.E2EOnlyUnits {
		cu := g.UnitByID[uid]
		if cu == nil {
			continue
		}
		owner := cu.Owner
		if owner == "" {
			owner = "unknown"
		}
		ensureOwner(owner).E2EOnlyUnits++
	}

	// Collect and sort.
	result := make([]OwnerRiskSummary, 0, len(owners))
	for _, s := range owners {
		result = append(result, *s)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].TotalSignals != result[j].TotalSignals {
			return result[i].TotalSignals > result[j].TotalSignals
		}
		return result[i].Owner < result[j].Owner
	})
	return result
}

// ModuleCoverageSummary describes coverage quality for a directory.
type ModuleCoverageSummary struct {
	Directory         string `json:"directory"`
	TotalUnits        int    `json:"totalUnits"`
	ExportedUnits     int    `json:"exportedUnits"`
	UncoveredExported int    `json:"uncoveredExported"`
	E2EOnlyUnits      int    `json:"e2eOnlyUnits"`
	TestCount         int    `json:"testCount"`
}

// ModuleCoverageSummaries returns per-directory coverage aggregations.
func (g *Graph) ModuleCoverageSummaries() []ModuleCoverageSummary {
	dirs := make(map[string]*ModuleCoverageSummary)

	ensureDir := func(d string) *ModuleCoverageSummary {
		if s, ok := dirs[d]; ok {
			return s
		}
		s := &ModuleCoverageSummary{Directory: d}
		dirs[d] = s
		return s
	}

	for _, cu := range g.snap.CodeUnits {
		dir := filepath.Dir(cu.Path)
		s := ensureDir(dir)
		s.TotalUnits++
		if cu.Exported {
			s.ExportedUnits++
		}
	}

	for _, uid := range g.UncoveredExportedUnits {
		cu := g.UnitByID[uid]
		if cu != nil {
			ensureDir(filepath.Dir(cu.Path)).UncoveredExported++
		}
	}

	for _, uid := range g.E2EOnlyUnits {
		cu := g.UnitByID[uid]
		if cu != nil {
			ensureDir(filepath.Dir(cu.Path)).E2EOnlyUnits++
		}
	}

	for file, testIDs := range g.TestsByFile {
		ensureDir(filepath.Dir(file)).TestCount += len(testIDs)
	}

	result := make([]ModuleCoverageSummary, 0, len(dirs))
	for _, s := range dirs {
		result = append(result, *s)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Directory < result[j].Directory
	})
	return result
}

func hasPrefix(path, prefix string) bool {
	if len(prefix) == 0 {
		return true
	}
	if len(path) <= len(prefix) {
		return false
	}
	return path[:len(prefix)] == prefix && path[len(prefix)] == '/'
}
