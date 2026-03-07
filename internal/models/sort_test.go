package models

import (
	"sort"
	"testing"
)

func TestSortSnapshot_Nil(t *testing.T) {
	// Must not panic on nil.
	SortSnapshot(nil)
}

func TestSortSnapshot_Empty(t *testing.T) {
	snap := &TestSuiteSnapshot{}
	SortSnapshot(snap)
	// No fields to check, just verify no panic.
}

func TestSortSnapshot_SignalOrder(t *testing.T) {
	snap := &TestSuiteSnapshot{
		Signals: []Signal{
			{Category: CategoryMigration, Type: "migration.deprecated-pattern", Location: SignalLocation{File: "b.js", Line: 10}},
			{Category: CategoryQuality, Type: "quality.weak-assertion", Location: SignalLocation{File: "a.js", Line: 5}},
			{Category: CategoryMigration, Type: "migration.custom-matcher", Location: SignalLocation{File: "a.js", Line: 1}},
			{Category: CategoryQuality, Type: "quality.weak-assertion", Location: SignalLocation{File: "a.js", Line: 1}},
		},
	}

	SortSnapshot(snap)

	// Signals should be sorted by category, type, file, line.
	if !sort.SliceIsSorted(snap.Signals, func(i, j int) bool {
		a, b := snap.Signals[i], snap.Signals[j]
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		if a.Type != b.Type {
			return a.Type < b.Type
		}
		if a.Location.File != b.Location.File {
			return a.Location.File < b.Location.File
		}
		return a.Location.Line < b.Location.Line
	}) {
		t.Error("signals not in canonical order")
	}

	// migration.* should come before quality.* (lexicographic on category)
	if snap.Signals[0].Category != CategoryMigration {
		t.Errorf("expected migration first, got %s", snap.Signals[0].Category)
	}
}

func TestSortSnapshot_TestFileOrder(t *testing.T) {
	snap := &TestSuiteSnapshot{
		TestFiles: []TestFile{
			{Path: "z/test.go"},
			{Path: "a/test.go"},
			{Path: "m/test.go"},
		},
	}

	SortSnapshot(snap)

	for i := 1; i < len(snap.TestFiles); i++ {
		if snap.TestFiles[i].Path < snap.TestFiles[i-1].Path {
			t.Errorf("test files not sorted: %s before %s",
				snap.TestFiles[i-1].Path, snap.TestFiles[i].Path)
		}
	}
}

func TestSortSnapshot_CodeUnitOrder(t *testing.T) {
	snap := &TestSuiteSnapshot{
		CodeUnits: []CodeUnit{
			{UnitID: "b:foo", Path: "b.js", Name: "foo"},
			{UnitID: "a:bar", Path: "a.js", Name: "bar"},
			{UnitID: "a:bar", Path: "a.js", Name: "baz"}, // same UnitID, diff name
		},
	}

	SortSnapshot(snap)

	if snap.CodeUnits[0].UnitID != "a:bar" {
		t.Errorf("expected a:bar first, got %s", snap.CodeUnits[0].UnitID)
	}
	// Two with same UnitID and Path should sort by Name.
	if snap.CodeUnits[0].Name != "bar" || snap.CodeUnits[1].Name != "baz" {
		t.Error("code units with same UnitID not sorted by Name")
	}
}

func TestSortSnapshot_FrameworkOrder(t *testing.T) {
	snap := &TestSuiteSnapshot{
		Frameworks: []Framework{
			{Name: "vitest"},
			{Name: "jest"},
			{Name: "playwright"},
		},
	}

	SortSnapshot(snap)

	if snap.Frameworks[0].Name != "jest" {
		t.Errorf("expected jest first, got %s", snap.Frameworks[0].Name)
	}
}

func TestSortSnapshot_Idempotent(t *testing.T) {
	snap := &TestSuiteSnapshot{
		Signals: []Signal{
			{Category: CategoryQuality, Type: "quality.weak-assertion", Location: SignalLocation{File: "a.js"}},
			{Category: CategoryMigration, Type: "migration.deprecated-pattern", Location: SignalLocation{File: "b.js"}},
		},
		TestFiles: []TestFile{
			{Path: "z.js"},
			{Path: "a.js"},
		},
	}

	SortSnapshot(snap)
	firstSignal := snap.Signals[0].Type
	firstFile := snap.TestFiles[0].Path

	// Sort again — should produce identical result.
	SortSnapshot(snap)
	if snap.Signals[0].Type != firstSignal {
		t.Error("sort is not idempotent for signals")
	}
	if snap.TestFiles[0].Path != firstFile {
		t.Error("sort is not idempotent for test files")
	}
}
