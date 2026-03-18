package truthcheck

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// CalibrationResult holds measured detection accuracy across fixture repos.
type CalibrationResult struct {
	// ByKind holds precision/recall per CodeSurfaceKind.
	ByKind map[models.CodeSurfaceKind]*KindCalibration `json:"byKind"`

	// ByTier holds aggregate precision/recall per detection tier.
	ByTier map[string]*TierCalibration `json:"byTier"`

	// ByConfidenceBand shows observed correctness in confidence bands.
	ByConfidenceBand []ConfidenceBand `json:"byConfidenceBand"`

	// FixtureCount is how many fixtures contributed to calibration.
	FixtureCount int `json:"fixtureCount"`

	// TotalSurfaces is the total surfaces evaluated.
	TotalSurfaces int `json:"totalSurfaces"`
}

// KindCalibration holds measured metrics for one CodeSurfaceKind.
type KindCalibration struct {
	Kind      string  `json:"kind"`
	Detected  int     `json:"detected"`  // total detected
	Correct   int     `json:"correct"`   // true positives
	Spurious  int     `json:"spurious"`  // false positives
	Missed    int     `json:"missed"`    // false negatives (from truth spec)
	Precision float64 `json:"precision"` // correct / (correct + spurious)
	Recall    float64 `json:"recall"`    // correct / (correct + missed)
	F1        float64 `json:"f1"`

	// AssignedConfidence is the current hardcoded confidence for this kind.
	AssignedConfidence float64 `json:"assignedConfidence"`

	// ObservedAccuracy is the empirically measured precision.
	// When this diverges significantly from AssignedConfidence, the
	// assigned score should be adjusted.
	ObservedAccuracy float64 `json:"observedAccuracy"`

	// Basis is "calibrated" if enough data points exist, "heuristic" otherwise.
	Basis string `json:"basis"`
}

// TierCalibration holds aggregate metrics for one detection tier.
type TierCalibration struct {
	Tier      string  `json:"tier"`
	Detected  int     `json:"detected"`
	Correct   int     `json:"correct"`
	Precision float64 `json:"precision"`
}

// ConfidenceBand shows observed correctness within a confidence range.
type ConfidenceBand struct {
	RangeLabel        string  `json:"rangeLabel"`        // e.g., "0.90-1.00"
	Min               float64 `json:"min"`
	Max               float64 `json:"max"`
	Count             int     `json:"count"`             // surfaces in this band
	CorrectCount      int     `json:"correctCount"`      // correctly detected
	ObservedPrecision float64 `json:"observedPrecision"` // correctCount / count
}

// CalibrateFromFixtures runs calibration using pre-analyzed snapshots and
// their truth specs. This does not re-run the pipeline — it takes already
// computed snapshots and truth expectations, making it fast and testable.
//
// For each fixture, it compares detected CodeSurfaces against truth-expected
// surfaces to measure per-kind precision. Surfaces that appear in the
// snapshot but not in the truth spec are counted as "unverified" (not
// necessarily wrong — the truth spec may be incomplete).
func CalibrateFromFixtures(fixtures []CalibrationInput) *CalibrationResult {
	result := &CalibrationResult{
		ByKind: map[models.CodeSurfaceKind]*KindCalibration{},
		ByTier: map[string]*TierCalibration{},
	}

	for _, f := range fixtures {
		calibrateOneFixture(f, result)
		result.FixtureCount++
	}

	// Compute derived metrics.
	for _, kc := range result.ByKind {
		if kc.Detected > 0 {
			kc.Precision = float64(kc.Correct) / float64(kc.Correct+kc.Spurious)
		}
		if kc.Correct+kc.Missed > 0 {
			kc.Recall = float64(kc.Correct) / float64(kc.Correct+kc.Missed)
		}
		if kc.Precision+kc.Recall > 0 {
			kc.F1 = 2 * kc.Precision * kc.Recall / (kc.Precision + kc.Recall)
		}
		kc.ObservedAccuracy = kc.Precision

		// Mark as calibrated if we have enough data points (>=5 detections).
		if kc.Detected >= 5 {
			kc.Basis = models.ConfidenceBasisCalibrated
		} else {
			kc.Basis = models.ConfidenceBasisHeuristic
		}
	}

	for _, tc := range result.ByTier {
		if tc.Detected > 0 {
			tc.Precision = float64(tc.Correct) / float64(tc.Detected)
		}
	}

	// Build confidence bands.
	result.ByConfidenceBand = buildConfidenceBands(fixtures)

	return result
}

// CalibrationInput pairs a snapshot with its truth spec.
type CalibrationInput struct {
	Name     string
	Snapshot *models.TestSuiteSnapshot
	Truth    *TruthSpec
}

func calibrateOneFixture(input CalibrationInput, result *CalibrationResult) {
	snap := input.Snapshot
	truth := input.Truth
	if snap == nil || truth == nil {
		return
	}

	// Build set of expected surface paths from truth spec.
	// AI truth provides explicit surface expectations.
	expectedSurfaces := map[string]models.CodeSurfaceKind{}
	if truth.AI != nil {
		for _, p := range truth.AI.ExpectedPromptSurfaces {
			expectedSurfaces[normalizeSurfacePath(p)] = models.SurfacePrompt
		}
		for _, p := range truth.AI.ExpectedContextSurfaces {
			expectedSurfaces[normalizeSurfacePath(p)] = models.SurfaceContext
		}
		for _, p := range truth.AI.ExpectedDatasetSurfaces {
			expectedSurfaces[normalizeSurfacePath(p)] = models.SurfaceDataset
		}
	}

	// Coverage truth provides expected source file paths.
	if truth.Coverage != nil {
		for _, item := range truth.Coverage.ExpectedUncovered {
			// These are source files that should have code surfaces but no test coverage.
			// Their presence confirms the surface detection is correct.
			expectedSurfaces[item.Path] = models.SurfaceFunction
		}
		for _, item := range truth.Coverage.ExpectedWeak {
			expectedSurfaces[item.Path] = models.SurfaceFunction
		}
	}

	// Fanout truth provides expected node paths.
	if truth.Fanout != nil {
		for _, f := range truth.Fanout.ExpectedFlagged {
			expectedSurfaces[normalizeSurfacePath(f.Node)] = models.SurfaceFunction
		}
	}

	// Evaluate each detected surface.
	matchedExpected := map[string]bool{}

	for _, cs := range snap.CodeSurfaces {
		result.TotalSurfaces++

		kc := getOrCreateKind(result, cs.Kind)
		kc.Detected++
		kc.AssignedConfidence = cs.Confidence

		tc := getOrCreateTier(result, cs.DetectionTier)
		tc.Detected++

		// Check if this surface matches a truth expectation.
		matched := false
		for expPath := range expectedSurfaces {
			if surfaceMatchesTruth(cs, expPath) {
				matched = true
				matchedExpected[expPath] = true
				break
			}
		}

		if matched {
			kc.Correct++
			tc.Correct++
		}
		// Note: unmatched surfaces are NOT counted as spurious here
		// because truth specs are incomplete — they only specify
		// surfaces relevant to the truth category being tested.
		// We can only measure recall against what's specified.
	}

	// Count missed expectations (in truth but not detected).
	for expPath, expKind := range expectedSurfaces {
		if !matchedExpected[expPath] {
			kc := getOrCreateKind(result, expKind)
			kc.Missed++
		}
	}
}

func surfaceMatchesTruth(cs models.CodeSurface, truthPath string) bool {
	// Direct path match.
	if cs.Path == truthPath {
		return true
	}
	// SurfaceID contains the path: "surface:path:name"
	if strings.Contains(cs.SurfaceID, truthPath) {
		return true
	}
	// Truth path may be a surface ID format: "path:name"
	idCheck := cs.Path + ":" + cs.Name
	if idCheck == truthPath {
		return true
	}
	return false
}

func normalizeSurfacePath(p string) string {
	// Strip "surface:" prefix if present.
	if strings.HasPrefix(p, "surface:") {
		p = p[len("surface:"):]
	}
	return p
}

func getOrCreateKind(result *CalibrationResult, kind models.CodeSurfaceKind) *KindCalibration {
	if kc, ok := result.ByKind[kind]; ok {
		return kc
	}
	kc := &KindCalibration{Kind: string(kind)}
	result.ByKind[kind] = kc
	return kc
}

func getOrCreateTier(result *CalibrationResult, tier string) *TierCalibration {
	if tc, ok := result.ByTier[tier]; ok {
		return tc
	}
	tc := &TierCalibration{Tier: tier}
	result.ByTier[tier] = tc
	return tc
}

func buildConfidenceBands(fixtures []CalibrationInput) []ConfidenceBand {
	bands := []ConfidenceBand{
		{RangeLabel: "0.95-1.00", Min: 0.95, Max: 1.00},
		{RangeLabel: "0.90-0.95", Min: 0.90, Max: 0.95},
		{RangeLabel: "0.85-0.90", Min: 0.85, Max: 0.90},
		{RangeLabel: "0.80-0.85", Min: 0.80, Max: 0.85},
		{RangeLabel: "0.70-0.80", Min: 0.70, Max: 0.80},
		{RangeLabel: "0.00-0.70", Min: 0.00, Max: 0.70},
	}

	// Build expected surface sets per fixture.
	for _, f := range fixtures {
		if f.Snapshot == nil || f.Truth == nil {
			continue
		}

		expectedPaths := map[string]bool{}
		if f.Truth.AI != nil {
			for _, p := range f.Truth.AI.ExpectedPromptSurfaces {
				expectedPaths[normalizeSurfacePath(p)] = true
			}
			for _, p := range f.Truth.AI.ExpectedContextSurfaces {
				expectedPaths[normalizeSurfacePath(p)] = true
			}
			for _, p := range f.Truth.AI.ExpectedDatasetSurfaces {
				expectedPaths[normalizeSurfacePath(p)] = true
			}
		}
		if f.Truth.Coverage != nil {
			for _, item := range f.Truth.Coverage.ExpectedUncovered {
				expectedPaths[item.Path] = true
			}
			for _, item := range f.Truth.Coverage.ExpectedWeak {
				expectedPaths[item.Path] = true
			}
		}

		for _, cs := range f.Snapshot.CodeSurfaces {
			for i := range bands {
				if cs.Confidence >= bands[i].Min && cs.Confidence < bands[i].Max ||
					(bands[i].Max == 1.00 && cs.Confidence >= bands[i].Min) {
					bands[i].Count++
					// Check if this surface has truth backing.
					for ep := range expectedPaths {
						if surfaceMatchesTruth(cs, ep) {
							bands[i].CorrectCount++
							break
						}
					}
					break
				}
			}
		}
	}

	for i := range bands {
		if bands[i].Count > 0 {
			bands[i].ObservedPrecision = float64(bands[i].CorrectCount) / float64(bands[i].Count)
		}
	}

	return bands
}

// FormatCalibrationReport produces a human-readable calibration summary.
func FormatCalibrationReport(r *CalibrationResult) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Confidence Calibration Report (%d fixtures, %d surfaces)\n", r.FixtureCount, r.TotalSurfaces))
	b.WriteString(strings.Repeat("=", 70) + "\n\n")

	// By tier.
	b.WriteString("By Detection Tier\n")
	b.WriteString(strings.Repeat("-", 70) + "\n")
	tiers := []string{models.TierStructural, models.TierSemantic, models.TierPattern, models.TierContent}
	for _, tier := range tiers {
		tc, ok := r.ByTier[tier]
		if !ok {
			continue
		}
		b.WriteString(fmt.Sprintf("  %-12s  %4d detected  precision: %.0f%%\n",
			tier, tc.Detected, tc.Precision*100))
	}
	b.WriteString("\n")

	// By kind — sorted by detection count descending.
	b.WriteString("By Surface Kind\n")
	b.WriteString(strings.Repeat("-", 70) + "\n")
	b.WriteString(fmt.Sprintf("  %-20s %6s %9s %9s %6s  %s\n",
		"Kind", "Count", "Precision", "Recall", "F1", "Basis"))
	kinds := make([]models.CodeSurfaceKind, 0, len(r.ByKind))
	for k := range r.ByKind {
		kinds = append(kinds, k)
	}
	sort.Slice(kinds, func(i, j int) bool {
		return r.ByKind[kinds[i]].Detected > r.ByKind[kinds[j]].Detected
	})
	for _, k := range kinds {
		kc := r.ByKind[k]
		b.WriteString(fmt.Sprintf("  %-20s %6d %8.0f%% %8.0f%% %5.0f%%  %s\n",
			kc.Kind, kc.Detected, kc.Precision*100, kc.Recall*100, kc.F1*100, kc.Basis))
	}
	b.WriteString("\n")

	// By confidence band.
	b.WriteString("By Confidence Band (observed correctness)\n")
	b.WriteString(strings.Repeat("-", 70) + "\n")
	for _, band := range r.ByConfidenceBand {
		if band.Count == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("  %-12s  %4d surfaces  %3d verified  observed: %.0f%%\n",
			band.RangeLabel, band.Count, band.CorrectCount, band.ObservedPrecision*100))
	}

	return b.String()
}
