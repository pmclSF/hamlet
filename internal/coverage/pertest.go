package coverage

import "sort"

// PerTestCoverage maps individual test IDs to their covered code units.
// This is an optional enrichment layer — not all frameworks support it.
type PerTestCoverage struct {
	// TestID is the stable test identifier.
	TestID string `json:"testId"`

	// CoveredFiles maps file paths to line coverage for this test.
	CoveredFiles map[string][]int `json:"coveredFiles,omitempty"`

	// CoveredUnitIDs lists code unit IDs this test covers.
	CoveredUnitIDs []string `json:"coveredUnitIds,omitempty"`

	// ScopeBreadth is the number of distinct files this test touches.
	ScopeBreadth int `json:"scopeBreadth"`
}

// TestCoverageRecord is the raw per-test coverage input.
type TestCoverageRecord struct {
	// TestID is the stable test identifier.
	TestID string `json:"testId"`

	// FilePath is the source file covered by this test.
	FilePath string `json:"filePath"`

	// CoveredLines lists 1-based line numbers hit by this test.
	CoveredLines []int `json:"coveredLines"`
}

// BuildPerTestCoverage constructs PerTestCoverage from raw records.
// It joins test coverage records with code unit definitions to determine
// which units each test covers.
func BuildPerTestCoverage(records []TestCoverageRecord, unitsByFile map[string][]UnitSpan) []PerTestCoverage {
	// Group records by test ID.
	byTest := map[string][]TestCoverageRecord{}
	for _, r := range records {
		byTest[r.TestID] = append(byTest[r.TestID], r)
	}

	// Sort test IDs for deterministic output.
	testIDs := make([]string, 0, len(byTest))
	for id := range byTest {
		testIDs = append(testIDs, id)
	}
	sort.Strings(testIDs)

	var result []PerTestCoverage
	for _, testID := range testIDs {
		recs := byTest[testID]
		ptc := PerTestCoverage{
			TestID:       testID,
			CoveredFiles: map[string][]int{},
		}

		unitSet := map[string]bool{}
		for _, rec := range recs {
			ptc.CoveredFiles[rec.FilePath] = rec.CoveredLines

			// Match covered lines to code units.
			if spans, ok := unitsByFile[rec.FilePath]; ok {
				for _, span := range spans {
					if linesOverlap(rec.CoveredLines, span.StartLine, span.EndLine) {
						unitSet[span.UnitID] = true
					}
				}
			}
		}

		ptc.ScopeBreadth = len(ptc.CoveredFiles)
		for uid := range unitSet {
			ptc.CoveredUnitIDs = append(ptc.CoveredUnitIDs, uid)
		}
		sort.Strings(ptc.CoveredUnitIDs)

		result = append(result, ptc)
	}

	return result
}

// UnitSpan represents a code unit's line span for per-test matching.
type UnitSpan struct {
	UnitID    string
	StartLine int
	EndLine   int
}

// BuildUnitSpanIndex creates a file-to-spans index from code units.
func BuildUnitSpanIndex(unitCovs []UnitCoverage, startLines, endLines map[string]int) map[string][]UnitSpan {
	index := map[string][]UnitSpan{}
	for _, uc := range unitCovs {
		start := startLines[uc.UnitID]
		end := endLines[uc.UnitID]
		if start > 0 {
			if end == 0 {
				end = start + 50
			}
			index[uc.Path] = append(index[uc.Path], UnitSpan{
				UnitID:    uc.UnitID,
				StartLine: start,
				EndLine:   end,
			})
		}
	}
	return index
}

func linesOverlap(lines []int, start, end int) bool {
	for _, l := range lines {
		if l >= start && l <= end {
			return true
		}
	}
	return false
}
