package impact

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
)

// mapChangedUnits maps changed files to impacted code units.
func mapChangedUnits(scope *ChangeScope, snap *models.TestSuiteSnapshot) []ImpactedCodeUnit {
	// Build code-unit index by file path.
	unitsByFile := map[string][]models.CodeUnit{}
	for _, cu := range snap.CodeUnits {
		unitsByFile[cu.Path] = append(unitsByFile[cu.Path], cu)
	}

	var impacted []ImpactedCodeUnit

	for _, cf := range scope.ChangedFiles {
		if cf.IsTestFile {
			continue // test files handled separately
		}

		units, found := unitsByFile[cf.Path]
		if !found {
			// File changed but no known code units — file-level impact.
			impacted = append(impacted, ImpactedCodeUnit{
				UnitID:           cf.Path,
				Name:             filepath.Base(cf.Path),
				Path:             cf.Path,
				ChangeKind:       cf.ChangeKind,
				ImpactConfidence: ConfidenceWeak,
				ProtectionStatus: classifyProtection(cf.Path, snap),
			})
			continue
		}

		for _, cu := range units {
			confidence := ConfidenceInferred
			if cf.ChangeKind == ChangeAdded || cf.ChangeKind == ChangeDeleted {
				confidence = ConfidenceExact
			}

			iu := ImpactedCodeUnit{
				UnitID:           cu.Path + ":" + cu.Name,
				Name:             cu.Name,
				Path:             cu.Path,
				ChangeKind:       cf.ChangeKind,
				Exported:         cu.Exported,
				ImpactConfidence: confidence,
				ProtectionStatus: classifyUnitProtection(cu, snap),
				CoveringTests:    findCoveringTests(cu, snap),
			}

			// Resolve owner from snapshot ownership map.
			if snap.Ownership != nil {
				if owners, ok := snap.Ownership[cu.Path]; ok && len(owners) > 0 {
					iu.Owner = owners[0]
				}
			}

			impacted = append(impacted, iu)
		}
	}

	return impacted
}

// findImpactedTests finds tests relevant to the change.
func findImpactedTests(scope *ChangeScope, snap *models.TestSuiteSnapshot, units []ImpactedCodeUnit) []ImpactedTest {
	// Build set of changed source files and their directories.
	changedDirs := map[string]bool{}
	changedSourceFiles := map[string]bool{}
	for _, cf := range scope.ChangedFiles {
		if !cf.IsTestFile {
			changedSourceFiles[cf.Path] = true
			changedDirs[filepath.Dir(cf.Path)] = true
		}
	}

	// Build set of covering test paths from impacted units.
	coveringTestPaths := map[string]bool{}
	for _, iu := range units {
		for _, tp := range iu.CoveringTests {
			coveringTestPaths[tp] = true
		}
	}

	var tests []ImpactedTest

	for _, tf := range snap.TestFiles {
		isDirectlyChanged := false
		for _, cf := range scope.ChangedFiles {
			if cf.Path == tf.Path {
				isDirectlyChanged = true
				break
			}
		}

		// Direct coverage link.
		if coveringTestPaths[tf.Path] {
			it := ImpactedTest{
				Path:              tf.Path,
				Framework:         tf.Framework,
				Relevance:         "covers impacted code unit",
				ImpactConfidence:  ConfidenceExact,
				IsDirectlyChanged: isDirectlyChanged,
			}
			// Find which units this test covers.
			for _, iu := range units {
				for _, ct := range iu.CoveringTests {
					if ct == tf.Path {
						it.CoversUnits = append(it.CoversUnits, iu.UnitID)
					}
				}
			}
			tests = append(tests, it)
			continue
		}

		// Directly changed test.
		if isDirectlyChanged {
			tests = append(tests, ImpactedTest{
				Path:              tf.Path,
				Framework:         tf.Framework,
				Relevance:         "test file directly changed",
				ImpactConfidence:  ConfidenceExact,
				IsDirectlyChanged: true,
			})
			continue
		}

		// Directory proximity heuristic.
		testDir := filepath.Dir(tf.Path)
		for dir := range changedDirs {
			if strings.HasPrefix(testDir, dir) || strings.HasPrefix(dir, testDir) {
				tests = append(tests, ImpactedTest{
					Path:             tf.Path,
					Framework:        tf.Framework,
					Relevance:        "in same directory tree as changed code",
					ImpactConfidence: ConfidenceInferred,
				})
				break
			}
		}
	}

	// Sort by confidence (exact first), then path.
	sort.Slice(tests, func(i, j int) bool {
		ci, cj := confidenceOrder(tests[i].ImpactConfidence), confidenceOrder(tests[j].ImpactConfidence)
		if ci != cj {
			return ci < cj
		}
		return tests[i].Path < tests[j].Path
	})

	return tests
}

// findProtectionGaps identifies where changed code lacks adequate coverage.
func findProtectionGaps(units []ImpactedCodeUnit, tests []ImpactedTest, snap *models.TestSuiteSnapshot) []ProtectionGap {
	var gaps []ProtectionGap

	for _, iu := range units {
		if iu.ProtectionStatus == ProtectionNone {
			severity := "medium"
			gapType := "no_coverage"
			explanation := fmt.Sprintf("%s has no observed test coverage.", iu.Name)
			action := fmt.Sprintf("Add unit tests for %s.", iu.Name)

			if iu.Exported {
				severity = "high"
				gapType = "untested_export"
				explanation = fmt.Sprintf("Exported function %s has no observed test coverage.", iu.Name)
				action = fmt.Sprintf("Add unit tests for exported function %s — this is public API surface.", iu.Name)
			}

			gaps = append(gaps, ProtectionGap{
				GapType:         gapType,
				CodeUnitID:      iu.UnitID,
				Path:            iu.Path,
				Explanation:     explanation,
				Severity:        severity,
				SuggestedAction: action,
			})
		}

		if iu.ProtectionStatus == ProtectionWeak && iu.Exported {
			gaps = append(gaps, ProtectionGap{
				GapType:         "weak_export_coverage",
				CodeUnitID:      iu.UnitID,
				Path:            iu.Path,
				Explanation:     fmt.Sprintf("Exported function %s is covered only by E2E or indirect tests.", iu.Name),
				Severity:        "medium",
				SuggestedAction: fmt.Sprintf("Add unit tests for %s to improve coverage diversity.", iu.Name),
			})
		}
	}

	return gaps
}

// selectProtectiveTests selects a focused protective test set.
func selectProtectiveTests(tests []ImpactedTest, units []ImpactedCodeUnit) []ImpactedTest {
	var selected []ImpactedTest

	// Always include exact-confidence tests and directly changed tests.
	for _, t := range tests {
		if t.ImpactConfidence == ConfidenceExact || t.IsDirectlyChanged {
			selected = append(selected, t)
		}
	}

	// If no exact tests, include inferred tests.
	if len(selected) == 0 {
		for _, t := range tests {
			if t.ImpactConfidence == ConfidenceInferred {
				selected = append(selected, t)
			}
		}
	}

	return selected
}

// computeChangeRiskPosture summarizes the risk posture.
func computeChangeRiskPosture(result *ImpactResult) ChangeRiskPosture {
	dims := []ChangeRiskDimension{
		computeProtectionDimension(result),
		computeExposureDimension(result),
		computeCoordinationDimension(result),
	}

	// Overall band is the worst dimension.
	bandOrder := map[string]int{"well_protected": 0, "partially_protected": 1, "weakly_protected": 2, "high_risk": 3}
	worst := "well_protected"
	for _, d := range dims {
		if bandOrder[d.Band] > bandOrder[worst] {
			worst = d.Band
		}
	}

	explanation := buildPostureExplanation(worst, result)

	return ChangeRiskPosture{
		Band:        worst,
		Explanation: explanation,
		Dimensions:  dims,
	}
}

func computeProtectionDimension(result *ImpactResult) ChangeRiskDimension {
	if len(result.ImpactedUnits) == 0 {
		return ChangeRiskDimension{
			Name: "protection", Band: "well_protected",
			Explanation: "No impacted code units identified.",
		}
	}

	unprotected := 0
	for _, iu := range result.ImpactedUnits {
		if iu.ProtectionStatus == ProtectionNone || iu.ProtectionStatus == ProtectionWeak {
			unprotected++
		}
	}

	ratio := float64(unprotected) / float64(len(result.ImpactedUnits))
	switch {
	case ratio == 0:
		return ChangeRiskDimension{
			Name: "protection", Band: "well_protected",
			Explanation: "All impacted code units have strong or partial coverage.",
		}
	case ratio < 0.3:
		return ChangeRiskDimension{
			Name: "protection", Band: "partially_protected",
			Explanation: fmt.Sprintf("%d of %d impacted unit(s) have weak or no coverage.", unprotected, len(result.ImpactedUnits)),
		}
	case ratio < 0.6:
		return ChangeRiskDimension{
			Name: "protection", Band: "weakly_protected",
			Explanation: fmt.Sprintf("%d of %d impacted unit(s) have weak or no coverage.", unprotected, len(result.ImpactedUnits)),
		}
	default:
		return ChangeRiskDimension{
			Name: "protection", Band: "high_risk",
			Explanation: fmt.Sprintf("%d of %d impacted unit(s) have weak or no coverage.", unprotected, len(result.ImpactedUnits)),
		}
	}
}

func computeExposureDimension(result *ImpactResult) ChangeRiskDimension {
	exportedCount := 0
	for _, iu := range result.ImpactedUnits {
		if iu.Exported {
			exportedCount++
		}
	}

	if exportedCount == 0 {
		return ChangeRiskDimension{
			Name: "exposure", Band: "well_protected",
			Explanation: "No exported/public code units affected.",
		}
	}

	band := "partially_protected"
	if exportedCount > 3 {
		band = "weakly_protected"
	}

	return ChangeRiskDimension{
		Name: "exposure", Band: band,
		Explanation: fmt.Sprintf("%d exported/public code unit(s) affected by this change.", exportedCount),
	}
}

func computeCoordinationDimension(result *ImpactResult) ChangeRiskDimension {
	ownerCount := len(result.ImpactedOwners)
	if ownerCount <= 1 {
		return ChangeRiskDimension{
			Name: "coordination", Band: "well_protected",
			Explanation: "Change affects a single owner area.",
		}
	}
	if ownerCount <= 3 {
		return ChangeRiskDimension{
			Name: "coordination", Band: "partially_protected",
			Explanation: fmt.Sprintf("Change spans %d owner areas.", ownerCount),
		}
	}
	return ChangeRiskDimension{
		Name: "coordination", Band: "weakly_protected",
		Explanation: fmt.Sprintf("Change spans %d owner areas — coordination risk is elevated.", ownerCount),
	}
}

func buildPostureExplanation(band string, result *ImpactResult) string {
	switch band {
	case "well_protected":
		return "This change appears well protected by existing tests."
	case "partially_protected":
		return fmt.Sprintf("This change has partial protection. %d protection gap(s) identified.", len(result.ProtectionGaps))
	case "weakly_protected":
		return fmt.Sprintf("This change has weak protection. %d protection gap(s) identified.", len(result.ProtectionGaps))
	case "high_risk":
		return fmt.Sprintf("This change has significant risk. %d protection gap(s) and weak coverage across impacted units.", len(result.ProtectionGaps))
	default:
		return "Unable to determine change-risk posture."
	}
}

// collectOwners collects unique owners from impacted units.
func collectOwners(units []ImpactedCodeUnit) []string {
	seen := map[string]bool{}
	var owners []string
	for _, iu := range units {
		if iu.Owner != "" && !seen[iu.Owner] {
			seen[iu.Owner] = true
			owners = append(owners, iu.Owner)
		}
	}
	sort.Strings(owners)
	return owners
}

// buildImpactSummary creates a human-readable summary.
func buildImpactSummary(result *ImpactResult) string {
	parts := []string{
		fmt.Sprintf("%d file(s) changed", len(result.Scope.ChangedFiles)),
	}
	if len(result.ImpactedUnits) > 0 {
		parts = append(parts, fmt.Sprintf("%d code unit(s) impacted", len(result.ImpactedUnits)))
	}
	if len(result.ImpactedTests) > 0 {
		parts = append(parts, fmt.Sprintf("%d test(s) relevant", len(result.ImpactedTests)))
	}
	if len(result.ProtectionGaps) > 0 {
		parts = append(parts, fmt.Sprintf("%d protection gap(s)", len(result.ProtectionGaps)))
	}
	return strings.Join(parts, ", ") + ". Posture: " + result.Posture.Band + "."
}

// identifyLimitations notes data gaps.
func identifyLimitations(scope *ChangeScope, snap *models.TestSuiteSnapshot, result *ImpactResult) []string {
	var lims []string

	if len(snap.CodeUnits) == 0 {
		lims = append(lims, "No code units discovered; impact analysis is file-level only.")
	}

	hasLineage := false
	for _, tf := range snap.TestFiles {
		if len(tf.LinkedCodeUnits) > 0 {
			hasLineage = true
			break
		}
	}
	if !hasLineage {
		lims = append(lims, "No per-test coverage lineage available; test selection uses structural heuristics.")
	}

	if snap.Ownership == nil || len(snap.Ownership) == 0 {
		lims = append(lims, "No ownership data available; coordination risk may be underestimated.")
	}

	return lims
}

// --- helpers ---

func classifyProtection(filePath string, snap *models.TestSuiteSnapshot) ProtectionStatus {
	for _, tf := range snap.TestFiles {
		for _, linked := range tf.LinkedCodeUnits {
			if linked == filePath {
				return ProtectionPartial
			}
		}
	}
	return ProtectionNone
}

func classifyUnitProtection(cu models.CodeUnit, snap *models.TestSuiteSnapshot) ProtectionStatus {
	unitID := cu.Path + ":" + cu.Name
	hasUnit := false
	hasE2E := false

	// Build framework type index.
	fwTypes := map[string]models.FrameworkType{}
	for _, fw := range snap.Frameworks {
		fwTypes[fw.Name] = fw.Type
	}

	for _, tf := range snap.TestFiles {
		for _, linked := range tf.LinkedCodeUnits {
			if linked == unitID || linked == cu.Name {
				fwType := fwTypes[tf.Framework]
				if fwType == models.FrameworkTypeUnit {
					hasUnit = true
				} else if fwType == models.FrameworkTypeE2E {
					hasE2E = true
				} else {
					hasUnit = true // integration, etc. treated as unit-level
				}
			}
		}
	}

	if hasUnit {
		return ProtectionStrong
	}
	if hasE2E {
		return ProtectionWeak
	}
	return ProtectionNone
}

func findCoveringTests(cu models.CodeUnit, snap *models.TestSuiteSnapshot) []string {
	unitID := cu.Path + ":" + cu.Name
	var tests []string
	for _, tf := range snap.TestFiles {
		for _, linked := range tf.LinkedCodeUnits {
			if linked == unitID || linked == cu.Name {
				tests = append(tests, tf.Path)
				break
			}
		}
	}
	return tests
}

func confidenceOrder(c Confidence) int {
	switch c {
	case ConfidenceExact:
		return 0
	case ConfidenceInferred:
		return 1
	case ConfidenceWeak:
		return 2
	default:
		return 3
	}
}
