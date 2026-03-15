// Package health implements runtime-backed health signal detectors.
//
// Health detectors consume normalized runtime data (from internal/runtime)
// and emit canonical models.Signal values for slowTest, flakyTest, etc.
//
// These detectors are distinct from static quality detectors because they
// require runtime evidence — they cannot operate on source code alone.
package health

import (
	"fmt"

	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/runtime"
)

// DefaultSlowThresholdMs is the default threshold for slow test detection.
// Tests exceeding this duration are flagged.
const DefaultSlowThresholdMs = 5000.0

// SlowTestDetector identifies tests whose runtime exceeds a configured threshold.
//
// Evidence: runtime artifact data (JUnit XML, Jest JSON)
// Limitations: single-run data; variance not available without multi-run history
type SlowTestDetector struct {
	ThresholdMs float64
}

// Detect scans runtime results for slow tests.
func (d *SlowTestDetector) Detect(results []runtime.TestResult) []models.Signal {
	threshold := d.ThresholdMs
	if threshold <= 0 {
		threshold = DefaultSlowThresholdMs
	}

	var signals []models.Signal
	for _, r := range results {
		if r.DurationMs <= 0 || r.Status == runtime.StatusSkipped {
			continue
		}
		if r.DurationMs > threshold {
			sev := slowSeverity(r.DurationMs, threshold)
			meta := map[string]any{
				"durationMs":  r.DurationMs,
				"thresholdMs": threshold,
				"suite":       r.Suite,
			}
			if r.TestID != "" {
				meta["testId"] = r.TestID
			}
			signals = append(signals, models.Signal{
				Type:     "slowTest",
				Category: models.CategoryHealth,
				Severity: sev,
				Confidence: 0.9,
				Location: models.SignalLocation{
					File:   r.File,
					Symbol: r.Name,
				},
				Explanation: fmt.Sprintf(
					"Observed runtime %.0fms exceeds threshold %.0fms.",
					r.DurationMs, threshold,
				),
				SuggestedAction: "Reduce fixture/setup cost, split expensive scenarios, or isolate integration-heavy behavior.",
				EvidenceStrength: models.EvidenceStrong,
				EvidenceSource:   models.SourceRuntime,
				Metadata: meta,
			})
		}
	}
	return signals
}

func slowSeverity(durationMs, thresholdMs float64) models.SignalSeverity {
	ratio := durationMs / thresholdMs
	switch {
	case ratio > 5:
		return models.SeverityHigh
	case ratio > 2:
		return models.SeverityMedium
	default:
		return models.SeverityLow
	}
}
