package quality

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pmclSF/hamlet/internal/models"
)

// CoverageThresholdDetector checks whether coverage is below a declared
// threshold.
//
// This detector looks for coverage summary data in standard locations:
//   - coverage/coverage-summary.json (Istanbul/nyc format)
//   - coverage-summary.json
//
// If no coverage data is found, no signals are emitted.
//
// Default threshold: 80% for lines coverage.
//
// Limitations:
//   - Only supports Istanbul/nyc JSON summary format in this stage.
//   - Does not parse lcov, clover, or other formats yet.
//   - Only checks aggregate (total) coverage, not per-file.
type CoverageThresholdDetector struct {
	// Threshold is the minimum acceptable coverage percentage (0-100).
	// Default: 80.
	Threshold float64
}

// istanbulSummary represents the Istanbul coverage-summary.json format.
type istanbulSummary struct {
	Total struct {
		Lines      *istanbulMetric `json:"lines"`
		Branches   *istanbulMetric `json:"branches"`
		Functions  *istanbulMetric `json:"functions"`
		Statements *istanbulMetric `json:"statements"`
	} `json:"total"`
}

type istanbulMetric struct {
	Pct float64 `json:"pct"`
}

// Detect looks for coverage data and checks against threshold.
func (d *CoverageThresholdDetector) Detect(snap *models.TestSuiteSnapshot) []models.Signal {
	threshold := d.Threshold
	if threshold == 0 {
		threshold = 80.0
	}

	root := snap.Repository.RootPath

	// Try standard coverage summary locations.
	paths := []string{
		filepath.Join(root, "coverage", "coverage-summary.json"),
		filepath.Join(root, "coverage-summary.json"),
	}

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}

		var summary istanbulSummary
		if err := json.Unmarshal(data, &summary); err != nil {
			continue
		}

		return d.checkThreshold(summary, threshold)
	}

	// No coverage data found — no signals to emit.
	// Currently supports Istanbul JSON format only.
	return nil
}

func (d *CoverageThresholdDetector) checkThreshold(summary istanbulSummary, threshold float64) []models.Signal {
	var signals []models.Signal

	type metric struct {
		name    string
		pct     float64
		present bool
	}

	metrics := []metric{
		{"lines", metricPct(summary.Total.Lines), summary.Total.Lines != nil},
		{"branches", metricPct(summary.Total.Branches), summary.Total.Branches != nil},
		{"functions", metricPct(summary.Total.Functions), summary.Total.Functions != nil},
		{"statements", metricPct(summary.Total.Statements), summary.Total.Statements != nil},
	}

	for _, m := range metrics {
		if !m.present {
			continue
		}
		if m.pct < threshold {
			sev := models.SeverityMedium
			if m.pct < threshold-20 {
				sev = models.SeverityHigh
			}

			signals = append(signals, models.Signal{
				Type:             "coverageThresholdBreak",
				Category:         models.CategoryQuality,
				Severity:         sev,
				Confidence:       0.9,
				EvidenceStrength: models.EvidenceStrong,
				EvidenceSource:   models.SourceCoverage,
				Location:         models.SignalLocation{Repository: "total"},
				Explanation: m.name + " coverage is " + formatPct(m.pct) +
					"%, below threshold of " + formatPct(threshold) + "%.",
				SuggestedAction: "Identify concentrated coverage gaps and target high-risk modules first.",
				Metadata: map[string]any{
					"metric":    m.name,
					"coverage":  m.pct,
					"threshold": threshold,
				},
			})
		}
	}

	return signals
}

func metricPct(m *istanbulMetric) float64 {
	if m == nil {
		return 0
	}
	return m.Pct
}

func formatPct(v float64) string {
	// Simple float formatting without importing strconv for a lightweight package.
	whole := int(v)
	frac := int((v - float64(whole)) * 10)
	if frac == 0 {
		return itoa(whole)
	}
	return itoa(whole) + "." + itoa(frac)
}
