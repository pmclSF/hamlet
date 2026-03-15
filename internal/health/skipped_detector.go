package health

import (
	"fmt"
	"sort"

	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/runtime"
)

// SkippedTestDetector identifies skipped tests from runtime artifacts.
//
// Skipped tests are a reliability concern: they indicate deferred work
// that may hide regressions.
type SkippedTestDetector struct{}

// Detect scans runtime results for skipped tests.
func (d *SkippedTestDetector) Detect(results []runtime.TestResult) []models.Signal {
	var signals []models.Signal
	skippedCount := 0
	totalCount := 0
	type fileCount struct {
		total   int
		skipped int
	}
	byFile := map[string]*fileCount{}

	for _, r := range results {
		totalCount++
		if r.Status == runtime.StatusSkipped {
			skippedCount++
		}
		if r.File != "" {
			fc := byFile[r.File]
			if fc == nil {
				fc = &fileCount{}
				byFile[r.File] = fc
			}
			fc.total++
			if r.Status == runtime.StatusSkipped {
				fc.skipped++
			}
		}
	}

	// Only emit if there's a meaningful number of skipped tests.
	if skippedCount == 0 || totalCount == 0 {
		return nil
	}

	ratio := float64(skippedCount) / float64(totalCount)
	sev := skippedSeverity(ratio)

	signals = append(signals, models.Signal{
		Type:       "skippedTest",
		Category:   models.CategoryHealth,
		Severity:   sev,
		Confidence: 0.9,
		Location:   models.SignalLocation{Repository: "runtime"},
		Explanation: fmt.Sprintf(
			"%d of %d tests skipped (%.0f%%) in runtime artifacts.",
			skippedCount, totalCount, ratio*100,
		),
		SuggestedAction:  "Review skipped tests — restore, remove, or document the skip reason.",
		EvidenceStrength: models.EvidenceStrong,
		EvidenceSource:   models.SourceRuntime,
		Metadata: map[string]any{
			"skippedCount": skippedCount,
			"totalCount":   totalCount,
			"ratio":        ratio,
			"scope":        "repository",
		},
	})

	files := make([]string, 0, len(byFile))
	for file := range byFile {
		files = append(files, file)
	}
	sort.Strings(files)

	for _, file := range files {
		fc := byFile[file]
		if fc == nil || fc.total == 0 || fc.skipped == 0 {
			continue
		}
		fileRatio := float64(fc.skipped) / float64(fc.total)
		signals = append(signals, models.Signal{
			Type:             "skippedTest",
			Category:         models.CategoryHealth,
			Severity:         skippedSeverity(fileRatio),
			Confidence:       0.9,
			EvidenceStrength: models.EvidenceStrong,
			EvidenceSource:   models.SourceRuntime,
			Location:         models.SignalLocation{File: file},
			Explanation: fmt.Sprintf(
				"%d of %d tests skipped (%.0f%%) in %s.",
				fc.skipped, fc.total, fileRatio*100, file,
			),
			SuggestedAction: "Review skipped tests — restore, remove, or document the skip reason.",
			Metadata: map[string]any{
				"skippedCount": fc.skipped,
				"totalCount":   fc.total,
				"ratio":        fileRatio,
				"scope":        "file",
			},
		})
	}

	return signals
}

func skippedSeverity(ratio float64) models.SignalSeverity {
	if ratio > 0.5 {
		return models.SeverityHigh
	}
	if ratio > 0.2 {
		return models.SeverityMedium
	}
	return models.SeverityLow
}
