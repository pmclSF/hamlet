package health

import (
	"fmt"

	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/runtime"
)

// UnstableSuiteDetector identifies suites with instability patterns.
//
// Evidence combines elevated failure rate with retries/skips, indicating
// unreliable suite behavior rather than isolated flaky tests.
type UnstableSuiteDetector struct{}

type suiteStats struct {
	file    string
	suite   string
	total   int
	fail    int
	error   int
	skipped int
	retried int
}

// Detect scans runtime results for unstable suite behavior.
func (d *UnstableSuiteDetector) Detect(results []runtime.TestResult) []models.Signal {
	bySuite := map[string]*suiteStats{}
	for _, r := range results {
		suiteName := r.Suite
		if suiteName == "" {
			suiteName = r.File
		}
		key := r.File + "::" + suiteName
		s := bySuite[key]
		if s == nil {
			s = &suiteStats{file: r.File, suite: suiteName}
			bySuite[key] = s
		}
		s.total++
		switch r.Status {
		case runtime.StatusFailed:
			s.fail++
		case runtime.StatusError:
			s.error++
		case runtime.StatusSkipped:
			s.skipped++
		}
		if r.Retried {
			s.retried++
		}
	}

	var signals []models.Signal
	for _, s := range bySuite {
		if s.total < 3 {
			continue
		}
		failRate := float64(s.fail+s.error) / float64(s.total)
		skipRate := float64(s.skipped) / float64(s.total)
		retryRate := float64(s.retried) / float64(s.total)

		unstable := (failRate >= 0.2 && (retryRate >= 0.1 || skipRate >= 0.2)) ||
			(retryRate >= 0.25 && skipRate >= 0.25)
		if !unstable {
			continue
		}

		sev := models.SeverityMedium
		if failRate >= 0.5 || (failRate >= 0.3 && retryRate >= 0.2) {
			sev = models.SeverityHigh
		}

		signals = append(signals, models.Signal{
			Type:             "unstableSuite",
			Category:         models.CategoryHealth,
			Severity:         sev,
			Confidence:       0.8,
			EvidenceStrength: models.EvidenceModerate,
			EvidenceSource:   models.SourceRuntime,
			Location: models.SignalLocation{
				File:   s.file,
				Symbol: s.suite,
			},
			Explanation: fmt.Sprintf(
				"Suite instability detected: fail/error %.0f%%, retry %.0f%%, skipped %.0f%% across %d test result(s).",
				failRate*100, retryRate*100, skipRate*100, s.total,
			),
			SuggestedAction: "Stabilize shared fixtures, isolate order-dependent tests, and triage recurring suite-level failures.",
			Metadata: map[string]any{
				"failRate":  failRate,
				"skipRate":  skipRate,
				"retryRate": retryRate,
				"total":     s.total,
			},
		})
	}

	return signals
}
