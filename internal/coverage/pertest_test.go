package coverage

import (
	"sort"
	"testing"
)

func TestBuildPerTestCoverage_Basic(t *testing.T) {
	records := []TestCoverageRecord{
		{TestID: "test-1", FilePath: "src/math.go", CoveredLines: []int{10, 11, 12}},
		{TestID: "test-1", FilePath: "src/utils.go", CoveredLines: []int{5, 6}},
		{TestID: "test-2", FilePath: "src/math.go", CoveredLines: []int{20, 21}},
	}

	unitsByFile := map[string][]UnitSpan{
		"src/math.go": {
			{UnitID: "src/math.go:Add", StartLine: 10, EndLine: 15},
			{UnitID: "src/math.go:Sub", StartLine: 18, EndLine: 25},
		},
		"src/utils.go": {
			{UnitID: "src/utils.go:Helper", StartLine: 1, EndLine: 10},
		},
	}

	result := BuildPerTestCoverage(records, unitsByFile)
	if len(result) != 2 {
		t.Fatalf("expected 2 per-test entries, got %d", len(result))
	}

	// Sort for deterministic assertions.
	sort.Slice(result, func(i, j int) bool { return result[i].TestID < result[j].TestID })

	ptc1 := result[0]
	if ptc1.TestID != "test-1" {
		t.Errorf("expected test-1, got %s", ptc1.TestID)
	}
	if ptc1.ScopeBreadth != 2 {
		t.Errorf("test-1 scope breadth = %d, want 2", ptc1.ScopeBreadth)
	}
	sort.Strings(ptc1.CoveredUnitIDs)
	if len(ptc1.CoveredUnitIDs) != 2 {
		t.Fatalf("test-1 covered units = %d, want 2", len(ptc1.CoveredUnitIDs))
	}
	if ptc1.CoveredUnitIDs[0] != "src/math.go:Add" || ptc1.CoveredUnitIDs[1] != "src/utils.go:Helper" {
		t.Errorf("test-1 units = %v", ptc1.CoveredUnitIDs)
	}

	ptc2 := result[1]
	if ptc2.TestID != "test-2" {
		t.Errorf("expected test-2, got %s", ptc2.TestID)
	}
	if ptc2.ScopeBreadth != 1 {
		t.Errorf("test-2 scope breadth = %d, want 1", ptc2.ScopeBreadth)
	}
	if len(ptc2.CoveredUnitIDs) != 1 || ptc2.CoveredUnitIDs[0] != "src/math.go:Sub" {
		t.Errorf("test-2 units = %v", ptc2.CoveredUnitIDs)
	}
}

func TestBuildPerTestCoverage_NoUnits(t *testing.T) {
	records := []TestCoverageRecord{
		{TestID: "test-1", FilePath: "src/orphan.go", CoveredLines: []int{1, 2, 3}},
	}

	result := BuildPerTestCoverage(records, map[string][]UnitSpan{})
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if len(result[0].CoveredUnitIDs) != 0 {
		t.Errorf("expected no covered units, got %v", result[0].CoveredUnitIDs)
	}
	if result[0].ScopeBreadth != 1 {
		t.Errorf("scope breadth = %d, want 1", result[0].ScopeBreadth)
	}
}

func TestBuildPerTestCoverage_Empty(t *testing.T) {
	result := BuildPerTestCoverage(nil, nil)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d", len(result))
	}
}

func TestLinesOverlap(t *testing.T) {
	tests := []struct {
		name  string
		lines []int
		start int
		end   int
		want  bool
	}{
		{"overlap at start", []int{10, 11}, 10, 20, true},
		{"overlap at end", []int{19, 20}, 10, 20, true},
		{"overlap in middle", []int{15}, 10, 20, true},
		{"no overlap before", []int{5, 8}, 10, 20, false},
		{"no overlap after", []int{21, 25}, 10, 20, false},
		{"empty lines", []int{}, 10, 20, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := linesOverlap(tt.lines, tt.start, tt.end)
			if got != tt.want {
				t.Errorf("linesOverlap(%v, %d, %d) = %v, want %v", tt.lines, tt.start, tt.end, got, tt.want)
			}
		})
	}
}

func TestBuildUnitSpanIndex(t *testing.T) {
	unitCovs := []UnitCoverage{
		{UnitID: "a.go:Foo", Path: "a.go"},
		{UnitID: "a.go:Bar", Path: "a.go"},
		{UnitID: "b.go:Baz", Path: "b.go"},
		{UnitID: "c.go:Skip", Path: "c.go"}, // no start line
	}

	startLines := map[string]int{
		"a.go:Foo": 10,
		"a.go:Bar": 30,
		"b.go:Baz": 5,
	}
	endLines := map[string]int{
		"a.go:Foo": 20,
		"a.go:Bar": 45,
		// b.go:Baz has no end — should default to start+50
	}

	index := BuildUnitSpanIndex(unitCovs, startLines, endLines)

	if len(index["a.go"]) != 2 {
		t.Errorf("a.go spans = %d, want 2", len(index["a.go"]))
	}
	if len(index["b.go"]) != 1 {
		t.Errorf("b.go spans = %d, want 1", len(index["b.go"]))
	}
	if len(index["c.go"]) != 0 {
		t.Errorf("c.go should have no spans (no start line), got %d", len(index["c.go"]))
	}

	// Check default end line.
	bSpan := index["b.go"][0]
	if bSpan.EndLine != 55 {
		t.Errorf("b.go:Baz end = %d, want 55 (start+50)", bSpan.EndLine)
	}
}
