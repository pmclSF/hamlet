// Package coverage implements coverage ingestion, attribution, and analysis.
//
// This package supports:
//   - ingesting LCOV and Istanbul JSON coverage artifacts
//   - normalizing coverage into internal CoverageRecord models
//   - attributing coverage to code units
//   - coverage-by-type analysis across labeled test runs
//   - per-test coverage when available
package coverage

// CoverageRecord represents normalized coverage data for a single file.
type CoverageRecord struct {
	// FilePath is the repository-relative path.
	FilePath string `json:"filePath"`

	// LineHits maps 1-based line numbers to hit counts.
	// A line not in this map was not instrumented.
	LineHits map[int]int `json:"lineHits,omitempty"`

	// BranchHits maps branch identifiers to hit counts.
	// Format is implementation-defined (e.g., "line:blockIndex").
	BranchHits map[string]int `json:"branchHits,omitempty"`

	// FunctionHits maps function names to hit counts.
	FunctionHits map[string]int `json:"functionHits,omitempty"`

	// LineCoveredCount is the number of instrumented lines that were executed.
	LineCoveredCount int `json:"lineCoveredCount"`

	// LineTotalCount is the total number of instrumented lines.
	LineTotalCount int `json:"lineTotalCount"`

	// BranchCoveredCount is the number of covered branches.
	BranchCoveredCount int `json:"branchCoveredCount,omitempty"`

	// BranchTotalCount is the total number of branches.
	BranchTotalCount int `json:"branchTotalCount,omitempty"`

	// FunctionCoveredCount is the number of covered functions.
	FunctionCoveredCount int `json:"functionCoveredCount,omitempty"`

	// FunctionTotalCount is the total number of functions.
	FunctionTotalCount int `json:"functionTotalCount,omitempty"`
}

// CoverageArtifact represents an ingested coverage artifact with provenance.
type CoverageArtifact struct {
	// Records are the per-file coverage records.
	Records []CoverageRecord `json:"records"`

	// Provenance describes where this coverage data came from.
	Provenance ArtifactProvenance `json:"provenance"`

	// RunLabel is the optional test bucket label (e.g., "unit", "e2e").
	RunLabel string `json:"runLabel,omitempty"`
}

// ArtifactProvenance captures where a coverage artifact came from.
type ArtifactProvenance struct {
	// SourceFile is the path to the original coverage file.
	SourceFile string `json:"sourceFile"`

	// Format is the detected format (e.g., "lcov", "istanbul").
	Format string `json:"format"`

	// RunLabel is the optional test type label if provided.
	RunLabel string `json:"runLabel,omitempty"`
}

// MergedCoverage contains the result of merging multiple coverage artifacts.
type MergedCoverage struct {
	// ByFile maps repository-relative file paths to merged coverage records.
	ByFile map[string]*CoverageRecord `json:"byFile"`

	// Artifacts lists the source artifacts that were merged.
	Artifacts []ArtifactProvenance `json:"artifacts"`
}

// Merge combines multiple coverage artifacts into a single merged view.
// Hit counts are summed across artifacts.
func Merge(artifacts []CoverageArtifact) *MergedCoverage {
	merged := &MergedCoverage{
		ByFile: map[string]*CoverageRecord{},
	}

	for _, art := range artifacts {
		merged.Artifacts = append(merged.Artifacts, art.Provenance)
		for _, rec := range art.Records {
			existing, ok := merged.ByFile[rec.FilePath]
			if !ok {
				// Clone the record.
				clone := rec
				clone.LineHits = cloneIntMap(rec.LineHits)
				clone.BranchHits = cloneStringIntMap(rec.BranchHits)
				clone.FunctionHits = cloneStringIntMap(rec.FunctionHits)
				merged.ByFile[rec.FilePath] = &clone
				continue
			}
			// Merge hit counts.
			for line, hits := range rec.LineHits {
				if existing.LineHits == nil {
					existing.LineHits = map[int]int{}
				}
				existing.LineHits[line] += hits
			}
			for branch, hits := range rec.BranchHits {
				if existing.BranchHits == nil {
					existing.BranchHits = map[string]int{}
				}
				existing.BranchHits[branch] += hits
			}
			for fn, hits := range rec.FunctionHits {
				if existing.FunctionHits == nil {
					existing.FunctionHits = map[string]int{}
				}
				existing.FunctionHits[fn] += hits
			}
			// Recompute summary counts.
			recomputeCounts(existing)
		}
	}

	return merged
}

func recomputeCounts(rec *CoverageRecord) {
	rec.LineTotalCount = len(rec.LineHits)
	rec.LineCoveredCount = 0
	for _, hits := range rec.LineHits {
		if hits > 0 {
			rec.LineCoveredCount++
		}
	}
	rec.BranchTotalCount = len(rec.BranchHits)
	rec.BranchCoveredCount = 0
	for _, hits := range rec.BranchHits {
		if hits > 0 {
			rec.BranchCoveredCount++
		}
	}
	rec.FunctionTotalCount = len(rec.FunctionHits)
	rec.FunctionCoveredCount = 0
	for _, hits := range rec.FunctionHits {
		if hits > 0 {
			rec.FunctionCoveredCount++
		}
	}
}

func cloneIntMap(m map[int]int) map[int]int {
	if m == nil {
		return nil
	}
	c := make(map[int]int, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

func cloneStringIntMap(m map[string]int) map[string]int {
	if m == nil {
		return nil
	}
	c := make(map[string]int, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}
