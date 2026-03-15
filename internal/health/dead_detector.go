package health

import (
	"fmt"

	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/runtime"
)

// DeadTestDetector identifies tests that are observed only in skipped state.
//
// This detector is intentionally conservative: a test is considered dead only
// when all observed outcomes are skipped/pending with no pass/fail evidence.
type DeadTestDetector struct{}

type deadTestStats struct {
	file       string
	name       string
	testID     string
	skipped    int
	nonSkipped int
}

// Detect scans runtime results and emits deadTest signals.
func (d *DeadTestDetector) Detect(results []runtime.TestResult) []models.Signal {
	statsByKey := map[string]*deadTestStats{}

	for _, r := range results {
		key := r.File + "::" + r.Suite + "::" + r.Name
		if r.TestID != "" {
			key = r.TestID
		}
		s := statsByKey[key]
		if s == nil {
			s = &deadTestStats{file: r.File, name: r.Name, testID: r.TestID}
			statsByKey[key] = s
		}
		if r.Status == runtime.StatusSkipped {
			s.skipped++
		} else {
			s.nonSkipped++
		}
	}

	var signals []models.Signal
	for _, s := range statsByKey {
		if s.skipped == 0 || s.nonSkipped > 0 {
			continue
		}
		sev := models.SeverityLow
		conf := 0.7
		if s.skipped >= 3 {
			sev = models.SeverityMedium
			conf = 0.85
		}
		meta := map[string]any{
			"skippedObservations": s.skipped,
		}
		if s.testID != "" {
			meta["testId"] = s.testID
		}

		signals = append(signals, models.Signal{
			Type:             "deadTest",
			Category:         models.CategoryHealth,
			Severity:         sev,
			Confidence:       conf,
			EvidenceStrength: models.EvidenceModerate,
			EvidenceSource:   models.SourceRuntime,
			Location: models.SignalLocation{
				File:   s.file,
				Symbol: s.name,
			},
			Explanation: fmt.Sprintf(
				"Test observed only in skipped/pending state across %d observation(s).",
				s.skipped,
			),
			SuggestedAction: "Remove dead tests or re-enable them with explicit assertions and ownership.",
			Metadata:        meta,
		})
	}

	return signals
}
