package envdepth

import (
	"fmt"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
)

// browserFrameworks maps framework names that imply a browser or browser-like
// runtime environment. This is descriptive, not judgmental: browser-backed
// tests exercise a different environmental depth than pure unit tests.
var browserFrameworks = map[string]bool{
	"cypress":      true,
	"playwright":   true,
	"puppeteer":    true,
	"testcafe":     true,
	"webdriverio":  true,
	"selenium":     true,
	"seleniumjava": true,
}

// integrationFrameworks maps framework names commonly associated with
// integration-level testing. A test using one of these frameworks with low
// mock counts is more likely exercising real dependencies.
var integrationFrameworks = map[string]bool{
	"supertest":  true,
	"httptest":   true,
	"testserver": true,
}

// Assess evaluates the environmental depth of each test file in the snapshot.
//
// The function classifies files into depth categories based on observable
// evidence: mock counts, assertion counts, and framework type. The intent is
// descriptive: heavy mocking is not inherently bad, and browser-backed tests
// are not inherently better. Different depth classes carry different risk
// profiles and coverage semantics.
func Assess(snap *models.TestSuiteSnapshot) *AssessmentResult {
	result := &AssessmentResult{
		ByDepth: make(map[DepthClass]int),
	}

	if snap == nil {
		result.OverallDepth = DepthUnknown
		return result
	}

	// Build a lookup of framework types from the snapshot's framework list.
	frameworkTypes := make(map[string]models.FrameworkType)
	for _, fw := range snap.Frameworks {
		frameworkTypes[strings.ToLower(fw.Name)] = fw.Type
	}

	for _, tf := range snap.TestFiles {
		a := assessFile(tf, frameworkTypes)
		result.Assessments = append(result.Assessments, a)
		result.ByDepth[a.Depth]++
	}

	result.OverallDepth = computeOverallDepth(result.ByDepth)
	return result
}

// assessFile produces an Assessment for a single test file.
func assessFile(tf models.TestFile, frameworkTypes map[string]models.FrameworkType) Assessment {
	a := Assessment{
		FilePath: tf.Path,
	}

	fwLower := strings.ToLower(tf.Framework)

	// Compute mock ratio when both counts are available.
	total := tf.MockCount + tf.AssertionCount
	if total > 0 {
		a.MockRatio = float64(tf.MockCount) / float64(total)
	}

	// Collect environment indicators based on observable evidence.
	if browserFrameworks[fwLower] {
		a.Indicators = append(a.Indicators, IndicatorBrowserDriver)
	}
	if tf.MockCount > 0 {
		a.Indicators = append(a.Indicators, IndicatorMockLibrary)
	}

	// Classification logic — ordered from most specific to least.
	switch {
	case browserFrameworks[fwLower]:
		a.Depth = DepthBrowserRuntime
		a.Confidence = 0.85
		a.Explanation = fmt.Sprintf(
			"Framework %q implies browser-backed execution environment.",
			tf.Framework,
		)

	case isHeavyMocking(tf):
		a.Depth = DepthHeavyMocking
		a.Confidence = 0.80
		a.Explanation = fmt.Sprintf(
			"High mock usage (%d mocks vs %d assertions) suggests heavy isolation from real dependencies.",
			tf.MockCount, tf.AssertionCount,
		)

	case isModerateMocking(tf):
		a.Depth = DepthModerateMocking
		a.Confidence = 0.70
		a.Explanation = fmt.Sprintf(
			"Moderate mock usage (%d mocks, %d assertions) suggests partial isolation.",
			tf.MockCount, tf.AssertionCount,
		)

	case isRealDependency(tf, fwLower, frameworkTypes):
		a.Depth = DepthRealDependency
		a.Confidence = 0.65
		a.Explanation = "Low or zero mock usage with integration/E2E framework suggests real dependency usage."

	default:
		a.Depth = DepthUnknown
		a.Confidence = 0.30
		a.Explanation = "Insufficient evidence to determine environmental depth."
	}

	return a
}

// isHeavyMocking returns true when mock usage is disproportionately high
// relative to assertions.
//
// Criteria: MockCount > 2*AssertionCount OR MockCount >= 8.
func isHeavyMocking(tf models.TestFile) bool {
	if tf.MockCount >= 8 {
		return true
	}
	if tf.MockCount > 0 && tf.MockCount > 2*tf.AssertionCount {
		return true
	}
	return false
}

// isModerateMocking returns true when mocks are present but not dominant.
//
// Criteria: MockCount > 0 and MockCount <= AssertionCount.
func isModerateMocking(tf models.TestFile) bool {
	return tf.MockCount > 0 && tf.MockCount <= tf.AssertionCount
}

// isRealDependency returns true when the test appears to exercise real
// dependencies: an E2E or integration framework with low mock count.
func isRealDependency(tf models.TestFile, fwLower string, frameworkTypes map[string]models.FrameworkType) bool {
	if tf.MockCount > 0 {
		return false
	}
	// Check explicit integration frameworks.
	if integrationFrameworks[fwLower] {
		return true
	}
	// Check if the framework type is E2E or integration.
	if ft, ok := frameworkTypes[fwLower]; ok {
		return ft == models.FrameworkTypeE2E || ft == models.FrameworkTypeIntegration
	}
	return false
}

// computeOverallDepth determines the dominant depth class across all files.
// When there is a tie, it prefers the class that represents more environmental
// realism (browser > real_dependency > moderate > heavy > unknown).
func computeOverallDepth(byDepth map[DepthClass]int) DepthClass {
	if len(byDepth) == 0 {
		return DepthUnknown
	}

	best := DepthUnknown
	bestCount := 0
	for depth, count := range byDepth {
		if count > bestCount || (count == bestCount && depthPriority(depth) > depthPriority(best)) {
			best = depth
			bestCount = count
		}
	}
	return best
}

// depthPriority returns a numeric priority for tie-breaking. Higher values
// represent greater environmental realism. This ordering is descriptive,
// not a value judgment.
func depthPriority(d DepthClass) int {
	switch d {
	case DepthBrowserRuntime:
		return 4
	case DepthRealDependency:
		return 3
	case DepthModerateMocking:
		return 2
	case DepthHeavyMocking:
		return 1
	default:
		return 0
	}
}
