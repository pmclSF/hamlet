package models

import "sort"

// SortSnapshot sorts all slice fields in the snapshot into a canonical,
// deterministic order. This ensures that identical inputs always produce
// byte-identical JSON output regardless of map iteration order, goroutine
// scheduling, or filesystem traversal order.
//
// Sort keys are chosen to be stable and meaningful:
//   - TestFiles: by Path
//   - TestCases: by TestID
//   - CodeUnits: by UnitID, then Path+Name
//   - Signals: by Category, Type, File, Line, Explanation
//   - Frameworks: by Name
//   - Risk: by Type, Scope, ScopeName
//   - CoverageInsights: by Type, Path, UnitID
func SortSnapshot(snap *TestSuiteSnapshot) {
	if snap == nil {
		return
	}

	sort.Slice(snap.TestFiles, func(i, j int) bool {
		return snap.TestFiles[i].Path < snap.TestFiles[j].Path
	})

	sort.Slice(snap.TestCases, func(i, j int) bool {
		return snap.TestCases[i].TestID < snap.TestCases[j].TestID
	})

	sort.Slice(snap.CodeUnits, func(i, j int) bool {
		a, b := snap.CodeUnits[i], snap.CodeUnits[j]
		if a.UnitID != b.UnitID {
			return a.UnitID < b.UnitID
		}
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		return a.Name < b.Name
	})

	sort.Slice(snap.Frameworks, func(i, j int) bool {
		return snap.Frameworks[i].Name < snap.Frameworks[j].Name
	})

	sortSignals(snap.Signals)

	sort.Slice(snap.Risk, func(i, j int) bool {
		a, b := snap.Risk[i], snap.Risk[j]
		if a.Type != b.Type {
			return a.Type < b.Type
		}
		if a.Scope != b.Scope {
			return a.Scope < b.Scope
		}
		return a.ScopeName < b.ScopeName
	})

	sort.Slice(snap.CoverageInsights, func(i, j int) bool {
		a, b := snap.CoverageInsights[i], snap.CoverageInsights[j]
		if a.Type != b.Type {
			return a.Type < b.Type
		}
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		return a.UnitID < b.UnitID
	})
}

// sortSignals sorts a slice of signals into canonical order.
func sortSignals(signals []Signal) {
	sort.Slice(signals, func(i, j int) bool {
		a, b := signals[i], signals[j]
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		if a.Type != b.Type {
			return a.Type < b.Type
		}
		if a.Location.File != b.Location.File {
			return a.Location.File < b.Location.File
		}
		if a.Location.Line != b.Location.Line {
			return a.Location.Line < b.Location.Line
		}
		return a.Explanation < b.Explanation
	})
}
