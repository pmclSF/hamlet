// Package benchmark provides the scaffolding for benchmark-safe data export.
//
// The export model is designed to be:
//   - privacy-safe: no raw file paths, symbol names, or source code
//   - segmented: tagged with repo characteristics for meaningful comparison
//   - versioned: includes the analysis version for compatibility
//   - hosted-ready: structured for future aggregation without schema changes
//
// This package defines the export format and segmentation primitives.
// The actual hosted benchmarking service is out of scope — this package
// only produces the local export artifact.
package benchmark

import (
	"time"

	"github.com/pmclSF/hamlet/internal/metrics"
	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/signals"
)

// Export is the benchmark-safe artifact that can be shared for comparison.
//
// It intentionally contains only aggregate data and segmentation tags.
// No raw file paths, symbol names, source code, or user identity.
type Export struct {
	// SchemaVersion identifies the export format version.
	SchemaVersion string `json:"schemaVersion"`

	// ExportedAt is when this export was created.
	ExportedAt time.Time `json:"exportedAt"`

	// Segment contains the segmentation tags for meaningful comparison.
	Segment Segment `json:"segment"`

	// Metrics contains the aggregate metrics.
	Metrics metrics.Snapshot `json:"metrics"`

	// PostureBands contains per-dimension posture bands from the measurement layer.
	// Privacy-safe: only band values, no raw data.
	PostureBands map[string]string `json:"postureBands,omitempty"`
}

// Segment contains tags that allow meaningful benchmark grouping.
//
// Segments enable comparing "like with like" — a 50-file Jest project
// should be compared against other small JS unit-test projects, not
// against a 5000-file multi-framework monorepo.
type Segment struct {
	// PrimaryLanguage is the dominant language in the test suite.
	PrimaryLanguage string `json:"primaryLanguage"`

	// PrimaryFramework is the most-used testing framework.
	PrimaryFramework string `json:"primaryFramework"`

	// TestFileBucket categorizes the repo by test suite size.
	// Values: "small" (<50), "medium" (50-500), "large" (>500)
	TestFileBucket string `json:"testFileBucket"`

	// FrameworkCount is the number of distinct frameworks.
	FrameworkCount int `json:"frameworkCount"`

	// HasCoverage indicates whether coverage data was available.
	HasCoverage bool `json:"hasCoverage"`

	// HasRuntimeData indicates whether runtime data was available.
	HasRuntimeData bool `json:"hasRuntimeData"`

	// HasPolicy indicates whether a policy file was present.
	HasPolicy bool `json:"hasPolicy"`
}

// BuildExport creates a benchmark-safe Export from a snapshot and derived metrics.
func BuildExport(snap *models.TestSuiteSnapshot, ms *metrics.Snapshot, hasPolicy bool) *Export {
	e := &Export{
		SchemaVersion: "2",
		ExportedAt:    time.Now().UTC(),
		Segment:       buildSegment(snap, ms, hasPolicy),
		Metrics:       *ms,
	}

	// Include posture bands if measurements are available.
	if snap.Measurements != nil {
		bands := map[string]string{}
		for _, p := range snap.Measurements.Posture {
			bands[p.Dimension] = p.Band
		}
		if len(bands) > 0 {
			e.PostureBands = bands
		}
	}

	return e
}

func buildSegment(snap *models.TestSuiteSnapshot, ms *metrics.Snapshot, hasPolicy bool) Segment {
	seg := Segment{
		FrameworkCount: ms.Structure.FrameworkCount,
		HasPolicy:      hasPolicy,
	}

	// Primary language
	if len(ms.Structure.Languages) > 0 {
		seg.PrimaryLanguage = ms.Structure.Languages[0]
	}

	// Primary framework (most test files)
	if len(snap.Frameworks) > 0 {
		best := snap.Frameworks[0]
		for _, fw := range snap.Frameworks[1:] {
			if fw.FileCount > best.FileCount {
				best = fw
			}
		}
		seg.PrimaryFramework = best.Name
	}

	// Test file bucket
	total := ms.Structure.TotalTestFiles
	switch {
	case total > 500:
		seg.TestFileBucket = "large"
	case total >= 50:
		seg.TestFileBucket = "medium"
	default:
		seg.TestFileBucket = "small"
	}

	// Coverage detection
	seg.HasCoverage = ms.Quality.CoverageThresholdBreakCount > 0 || hasCoverageSignals(snap)

	// Runtime detection
	for _, tf := range snap.TestFiles {
		if tf.RuntimeStats != nil && tf.RuntimeStats.AvgRuntimeMs > 0 {
			seg.HasRuntimeData = true
			break
		}
	}

	return seg
}

func hasCoverageSignals(snap *models.TestSuiteSnapshot) bool {
	for _, s := range snap.Signals {
		if s.Type == signals.SignalCoverageThresholdBreak || s.Type == signals.SignalCoverageBlindSpot {
			return true
		}
	}
	return false
}
