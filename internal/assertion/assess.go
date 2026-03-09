package assertion

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
)

// e2eFrameworks lists framework names whose lower assertion density is expected.
var e2eFrameworks = map[string]bool{
	"cypress":     true,
	"playwright":  true,
	"puppeteer":   true,
	"selenium":    true,
	"webdriverio": true,
	"testcafe":    true,
}

// Assess evaluates assertion strength across all test files in the snapshot.
func Assess(snap *models.TestSuiteSnapshot) *AssessmentResult {
	result := &AssessmentResult{
		ByStrength: make(map[StrengthClass]int),
	}

	if snap == nil || len(snap.TestFiles) == 0 {
		result.OverallStrength = StrengthUnclear
		return result
	}

	// Build a lookup of framework types from the snapshot.
	frameworkTypes := buildFrameworkTypeLookup(snap)

	var totalDensity float64
	var filesWithTests int

	for i := range snap.TestFiles {
		tf := &snap.TestFiles[i]
		a := assessFile(tf, frameworkTypes)
		result.Assessments = append(result.Assessments, a)
		result.ByStrength[a.Strength]++
		if tf.TestCount > 0 {
			totalDensity += a.Density
			filesWithTests++
		}
	}

	// Sort for determinism.
	sort.Slice(result.Assessments, func(i, j int) bool {
		return result.Assessments[i].FilePath < result.Assessments[j].FilePath
	})

	if filesWithTests > 0 {
		result.AverageDensity = totalDensity / float64(filesWithTests)
	}

	result.OverallStrength = computeOverallStrength(result)

	return result
}

// buildFrameworkTypeLookup maps framework names to their FrameworkType.
func buildFrameworkTypeLookup(snap *models.TestSuiteSnapshot) map[string]models.FrameworkType {
	lookup := make(map[string]models.FrameworkType)
	for _, fw := range snap.Frameworks {
		lookup[strings.ToLower(fw.Name)] = fw.Type
	}
	return lookup
}

// assessFile produces an Assessment for a single test file.
func assessFile(tf *models.TestFile, frameworkTypes map[string]models.FrameworkType) Assessment {
	a := Assessment{
		FilePath:   tf.Path,
		Categories: make(map[AssertionCategory]int),
	}

	a.AssertionCount = tf.AssertionCount
	a.TestCount = tf.TestCount

	// If no tests, we cannot assess meaningfully.
	if tf.TestCount == 0 {
		a.Strength = StrengthUnclear
		a.Confidence = 0.2
		a.Explanation = "no tests detected in file"
		a.DominantCategory = CategoryUnknown
		return a
	}

	// Compute density.
	a.Density = float64(tf.AssertionCount) / float64(tf.TestCount)

	// Infer categories from available snapshot data.
	inferCategories(&a, tf)

	// Determine dominant category.
	a.DominantCategory = dominantCategory(a.Categories)

	// Check if this is an E2E framework.
	isE2E := isE2EFramework(tf.Framework, frameworkTypes)

	// Classify strength.
	classifyStrength(&a, tf, isE2E)

	return a
}

// inferCategories infers assertion categories from available snapshot fields.
func inferCategories(a *Assessment, tf *models.TestFile) {
	remaining := tf.AssertionCount

	// Snapshot assertions are directly available.
	if tf.SnapshotCount > 0 {
		a.Categories[CategorySnapshot] = tf.SnapshotCount
		remaining -= tf.SnapshotCount
		if remaining < 0 {
			remaining = 0
		}
	}

	// The remaining assertions are categorized as behavioral since we cannot
	// distinguish further from the snapshot data alone.
	if remaining > 0 {
		a.Categories[CategoryBehavioral] = remaining
	}

	// If zero assertions were detected, mark unknown.
	if tf.AssertionCount == 0 {
		a.Categories[CategoryUnknown] = 0
	}
}

// dominantCategory returns the category with the highest count.
func dominantCategory(cats map[AssertionCategory]int) AssertionCategory {
	best := CategoryUnknown
	bestCount := -1
	// Iterate in a deterministic order.
	keys := make([]AssertionCategory, 0, len(cats))
	for k := range cats {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return string(keys[i]) < string(keys[j])
	})
	for _, k := range keys {
		if cats[k] > bestCount {
			bestCount = cats[k]
			best = k
		}
	}
	return best
}

// isE2EFramework checks whether the framework is an E2E framework.
func isE2EFramework(framework string, frameworkTypes map[string]models.FrameworkType) bool {
	fw := strings.ToLower(framework)
	if e2eFrameworks[fw] {
		return true
	}
	if ft, ok := frameworkTypes[fw]; ok {
		return ft == models.FrameworkTypeE2E
	}
	return false
}

// classifyStrength assigns strength, confidence, and explanation.
func classifyStrength(a *Assessment, tf *models.TestFile, isE2E bool) {
	mockRatio := float64(0)
	if tf.AssertionCount > 0 {
		mockRatio = float64(tf.MockCount) / float64(tf.AssertionCount)
	}

	snapshotRatio := float64(0)
	if tf.AssertionCount > 0 {
		snapshotRatio = float64(tf.SnapshotCount) / float64(tf.AssertionCount)
	}

	// No assertions at all.
	if tf.AssertionCount == 0 {
		a.Strength = StrengthWeak
		a.Confidence = 0.8
		a.Explanation = "no assertions detected"
		return
	}

	// Mock-heavy: more mocks than assertions suggests weak oracle.
	if tf.MockCount > tf.AssertionCount {
		a.Strength = StrengthWeak
		a.Confidence = 0.7
		a.Explanation = fmt.Sprintf(
			"mock-heavy test: %d mocks vs %d assertions",
			tf.MockCount, tf.AssertionCount,
		)
		return
	}

	// High snapshot ratio with low overall density.
	if snapshotRatio >= 0.8 && a.Density < 2.0 {
		a.Strength = StrengthWeak
		a.Confidence = 0.6
		a.Explanation = "dominated by snapshot assertions with low density"
		return
	}

	// E2E frameworks: lower density thresholds.
	if isE2E {
		classifyE2EStrength(a, mockRatio)
		return
	}

	// Strong: high density and low mock ratio.
	if a.Density >= 3.0 && mockRatio < 0.5 {
		a.Strength = StrengthStrong
		a.Confidence = 0.8
		a.Explanation = fmt.Sprintf(
			"high assertion density (%.1f/test) with low mock ratio",
			a.Density,
		)
		return
	}

	// Moderate: decent density.
	if a.Density >= 1.5 {
		a.Strength = StrengthModerate
		a.Confidence = 0.7
		a.Explanation = fmt.Sprintf(
			"moderate assertion density (%.1f/test)",
			a.Density,
		)
		return
	}

	// Low density.
	if a.Density < 1.0 {
		a.Strength = StrengthWeak
		a.Confidence = 0.6
		a.Explanation = fmt.Sprintf(
			"low assertion density (%.1f/test)",
			a.Density,
		)
		return
	}

	// Between 1.0 and 1.5 — borderline moderate.
	a.Strength = StrengthModerate
	a.Confidence = 0.5
	a.Explanation = fmt.Sprintf(
		"borderline assertion density (%.1f/test)",
		a.Density,
	)
}

// classifyE2EStrength applies adjusted thresholds for E2E frameworks,
// which typically have lower assertion density but may still be strong.
func classifyE2EStrength(a *Assessment, mockRatio float64) {
	if a.Density >= 2.0 && mockRatio < 0.3 {
		a.Strength = StrengthStrong
		a.Confidence = 0.7
		a.Explanation = fmt.Sprintf(
			"E2E test with good assertion density (%.1f/test)",
			a.Density,
		)
		return
	}

	if a.Density >= 1.0 {
		a.Strength = StrengthModerate
		a.Confidence = 0.6
		a.Explanation = fmt.Sprintf(
			"E2E test with acceptable assertion density (%.1f/test)",
			a.Density,
		)
		return
	}

	if a.Density > 0 {
		a.Strength = StrengthModerate
		a.Confidence = 0.5
		a.Explanation = fmt.Sprintf(
			"E2E test with low assertion density (%.1f/test); may rely on implicit checks",
			a.Density,
		)
		return
	}

	a.Strength = StrengthWeak
	a.Confidence = 0.5
	a.Explanation = "E2E test with no assertions detected"
}

// computeOverallStrength derives the aggregate strength from individual assessments.
func computeOverallStrength(result *AssessmentResult) StrengthClass {
	total := len(result.Assessments)
	if total == 0 {
		return StrengthUnclear
	}

	strong := result.ByStrength[StrengthStrong]
	moderate := result.ByStrength[StrengthModerate]
	weak := result.ByStrength[StrengthWeak]
	unclear := result.ByStrength[StrengthUnclear]

	// If everything is unclear, overall is unclear.
	if unclear == total {
		return StrengthUnclear
	}

	assessed := total - unclear
	if assessed == 0 {
		return StrengthUnclear
	}

	strongPct := float64(strong) / float64(assessed)
	weakPct := float64(weak) / float64(assessed)
	_ = moderate // used implicitly

	if strongPct >= 0.5 && weakPct < 0.2 {
		return StrengthStrong
	}
	if weakPct >= 0.5 {
		return StrengthWeak
	}
	return StrengthModerate
}
