package runtime

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Ingest parses a runtime artifact file and returns normalized results.
// The format is auto-detected from the file extension and content.
//
// Supported formats:
//   - .xml → JUnit XML
//   - .json → Jest/Vitest JSON
func Ingest(path string) (*IngestionResult, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".xml":
		return ParseJUnitXML(path)
	case ".json":
		return ParseJestJSON(path)
	default:
		return nil, fmt.Errorf("unsupported runtime artifact format: %s (expected .xml or .json)", ext)
	}
}

// ApplyToTestFiles merges runtime results into the snapshot's test files
// by matching file paths. This populates RuntimeStats on TestFile entries.
func ApplyToTestFiles(results []TestResult, testFiles []TestFileUpdate) {
	// Group results by file.
	byFile := map[string][]TestResult{}
	for _, r := range results {
		if r.File != "" {
			byFile[r.File] = append(byFile[r.File], r)
		}
	}

	for i := range testFiles {
		fileResults := findResultsForFile(byFile, testFiles[i].Path)
		if len(fileResults) == 0 {
			continue
		}

		var totalMs float64
		var maxMs float64
		var passCount, failCount, retryCount int
		for _, r := range fileResults {
			totalMs += r.DurationMs
			if r.DurationMs > maxMs {
				maxMs = r.DurationMs
			}
			if r.Status == StatusPassed {
				passCount++
			} else if r.Status == StatusFailed || r.Status == StatusError {
				failCount++
			}
			if r.Retried {
				retryCount++
			}
		}

		total := passCount + failCount
		var passRate float64
		if total > 0 {
			passRate = float64(passCount) / float64(total)
		}
		var retryRate float64
		if len(fileResults) > 0 {
			retryRate = float64(retryCount) / float64(len(fileResults))
		}

		testFiles[i].AvgRuntimeMs = totalMs / float64(len(fileResults))
		testFiles[i].P95RuntimeMs = maxMs // approximate: use max as P95 for single-run data
		testFiles[i].PassRate = passRate
		testFiles[i].RetryRate = retryRate
	}
}

// TestFileUpdate is a lightweight struct for applying runtime stats.
type TestFileUpdate struct {
	Path         string
	AvgRuntimeMs float64
	P95RuntimeMs float64
	PassRate     float64
	RetryRate    float64
}

// findResultsForFile matches runtime results to a test file path.
// Handles both exact matches and suffix-based matching (since CI artifacts
// may report absolute paths while the snapshot uses relative paths).
func findResultsForFile(byFile map[string][]TestResult, testFilePath string) []TestResult {
	// Exact match.
	if results, ok := byFile[testFilePath]; ok {
		return results
	}
	// Suffix match: runtime path might be absolute.
	for path, results := range byFile {
		if strings.HasSuffix(path, "/"+testFilePath) || strings.HasSuffix(path, "\\"+testFilePath) {
			return results
		}
	}
	return nil
}
