package health

import (
	"fmt"

	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/runtime"
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

	for _, r := range results {
		totalCount++
		if r.Status == runtime.StatusSkipped {
			skippedCount++
		}
	}

	// Only emit if there's a meaningful number of skipped tests.
	if skippedCount == 0 || totalCount == 0 {
		return nil
	}

	ratio := float64(skippedCount) / float64(totalCount)
	sev := models.SeverityLow
	if ratio > 0.2 {
		sev = models.SeverityMedium
	}
	if ratio > 0.5 {
		sev = models.SeverityHigh
	}

	signals = append(signals, models.Signal{
		Type:     "skippedTest",
		Category: models.CategoryHealth,
		Severity: sev,
		Confidence: 0.9,
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
		},
	})

	return signals
}
