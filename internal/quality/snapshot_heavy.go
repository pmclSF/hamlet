package quality

import (
	"fmt"

	"github.com/pmclSF/hamlet/internal/models"
)

// SnapshotHeavyDetector identifies test files that rely heavily on snapshots
// relative to direct semantic assertions.
type SnapshotHeavyDetector struct{}

// Detect emits snapshotHeavyTest signals for files where snapshot usage is
// high enough to indicate brittle, low-semantic verification.
func (d *SnapshotHeavyDetector) Detect(snap *models.TestSuiteSnapshot) []models.Signal {
	var out []models.Signal

	for _, tf := range snap.TestFiles {
		if tf.SnapshotCount <= 0 {
			continue
		}

		ratio := 0.0
		if tf.AssertionCount > 0 {
			ratio = float64(tf.SnapshotCount) / float64(tf.AssertionCount)
		}

		// Filter out incidental snapshot use.
		if tf.SnapshotCount < 3 && ratio < 1.0 {
			continue
		}

		severity := models.SeverityMedium
		if tf.AssertionCount == 0 || ratio >= 2.0 || tf.SnapshotCount >= 10 {
			severity = models.SeverityHigh
		}

		explanation := fmt.Sprintf(
			"%s uses %d snapshot assertion(s) with %d direct assertion(s) (snapshot/assertion ratio %.2f).",
			tf.Path, tf.SnapshotCount, tf.AssertionCount, ratio,
		)

		out = append(out, models.Signal{
			Type:             "snapshotHeavyTest",
			Category:         models.CategoryQuality,
			Severity:         severity,
			Confidence:       0.8,
			EvidenceStrength: models.EvidenceModerate,
			EvidenceSource:   models.SourceStructuralPattern,
			Location:         models.SignalLocation{File: tf.Path},
			Owner:            tf.Owner,
			Explanation:      explanation,
			SuggestedAction:  "Replace broad snapshot assertions with targeted semantic assertions for critical behavior.",
			Metadata: map[string]any{
				"snapshotCount":  tf.SnapshotCount,
				"assertionCount": tf.AssertionCount,
				"snapshotRatio":  ratio,
			},
		})
	}

	return out
}
