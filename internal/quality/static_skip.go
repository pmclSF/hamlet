package quality

import (
	"fmt"
	"sort"

	"github.com/pmclSF/terrain/internal/models"
)

// StaticSkipDetector identifies statically skipped tests from source code patterns.
//
// This detector finds skip markers in test files without requiring runtime artifacts:
//   - JS/TS: it.skip(), test.skip(), describe.skip(), xit(), xdescribe()
//   - Go: t.Skip(), t.Skipf(), t.SkipNow()
//   - Python: @pytest.mark.skip, @unittest.skip, pytest.skip()
//   - Java: @Disabled, @Ignore
//
// This closes the P0 gap where docs promise skip detection from `terrain analyze`
// but the runtime-based SkippedTestDetector requires --runtime artifacts.
type StaticSkipDetector struct{}

// Detect scans test files for static skip patterns.
func (d *StaticSkipDetector) Detect(snap *models.TestSuiteSnapshot) []models.Signal {
	var signals []models.Signal
	totalSkipped := 0
	totalTests := 0

	type fileSkip struct {
		path      string
		skips     int
		tests     int
		framework string
	}
	var skippedFiles []fileSkip

	for _, tf := range snap.TestFiles {
		if tf.TestCount == 0 {
			continue
		}
		totalTests += tf.TestCount
		if tf.SkipCount > 0 {
			totalSkipped += tf.SkipCount
			skippedFiles = append(skippedFiles, fileSkip{
				path:      tf.Path,
				skips:     tf.SkipCount,
				tests:     tf.TestCount,
				framework: tf.Framework,
			})
		}
	}

	if totalSkipped == 0 || totalTests == 0 {
		return nil
	}

	ratio := float64(totalSkipped) / float64(totalTests)
	sev := staticSkipSeverity(ratio)

	signals = append(signals, models.Signal{
		Type:       "staticSkippedTest",
		Category:   models.CategoryHealth,
		Severity:   sev,
		Confidence: 0.8,
		Location:   models.SignalLocation{Repository: "static"},
		Explanation: fmt.Sprintf(
			"%d of %d tests statically skipped (%.0f%%) via code markers (.skip, xit, @skip, etc.).",
			totalSkipped, totalTests, ratio*100,
		),
		SuggestedAction:  "Review skipped tests — restore, remove, or convert to conditional skips with documented reasons.",
		EvidenceStrength: models.EvidencePartial,
		EvidenceSource:   models.SourceStructuralPattern,
		Metadata: map[string]any{
			"skippedCount": totalSkipped,
			"totalCount":   totalTests,
			"ratio":        ratio,
			"scope":        "repository",
			"detection":    "static",
		},
	})

	// Sort by skip ratio descending for deterministic output.
	sort.Slice(skippedFiles, func(i, j int) bool {
		ri := float64(skippedFiles[i].skips) / float64(skippedFiles[i].tests)
		rj := float64(skippedFiles[j].skips) / float64(skippedFiles[j].tests)
		if ri != rj {
			return ri > rj
		}
		return skippedFiles[i].path < skippedFiles[j].path
	})

	for _, sf := range skippedFiles {
		fileRatio := float64(sf.skips) / float64(sf.tests)
		signals = append(signals, models.Signal{
			Type:             "staticSkippedTest",
			Category:         models.CategoryHealth,
			Severity:         staticSkipSeverity(fileRatio),
			Confidence:       0.8,
			EvidenceStrength: models.EvidencePartial,
			EvidenceSource:   models.SourceStructuralPattern,
			Location:         models.SignalLocation{File: sf.path},
			Explanation: fmt.Sprintf(
				"%d of %d tests statically skipped (%.0f%%) in %s.",
				sf.skips, sf.tests, fileRatio*100, sf.path,
			),
			SuggestedAction: "Review skipped tests — restore, remove, or document the skip reason.",
			Metadata: map[string]any{
				"skippedCount": sf.skips,
				"totalCount":   sf.tests,
				"ratio":        fileRatio,
				"scope":        "file",
				"detection":    "static",
			},
		})
	}

	return signals
}

func staticSkipSeverity(ratio float64) models.SignalSeverity {
	if ratio > 0.5 {
		return models.SeverityHigh
	}
	if ratio > 0.2 {
		return models.SeverityMedium
	}
	return models.SeverityLow
}
