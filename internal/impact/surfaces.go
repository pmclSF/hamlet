package impact

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
)

// mapChangedSurfaces identifies code surfaces in changed files and groups
// them by domain area. When no code surfaces exist for a changed file,
// it falls back to file-level representation.
func mapChangedSurfaces(scope *ChangeScope, snap *models.TestSuiteSnapshot) []ChangedArea {
	if snap == nil {
		return nil
	}

	// Index surfaces by file path.
	surfacesByPath := map[string][]models.CodeSurface{}
	for _, cs := range snap.CodeSurfaces {
		surfacesByPath[cs.Path] = append(surfacesByPath[cs.Path], cs)
	}

	// Collect changed surfaces grouped by area.
	areaMap := map[string][]ChangedSurface{}

	for _, cf := range scope.ChangedFiles {
		if cf.IsTestFile {
			continue
		}
		if !IsAnalyzableSourceFile(cf.Path) {
			continue
		}

		surfaces, hasSurfaces := surfacesByPath[cf.Path]
		area := inferArea(cf.Path)

		if hasSurfaces {
			for _, s := range surfaces {
				areaMap[area] = append(areaMap[area], ChangedSurface{
					SurfaceID:  s.SurfaceID,
					Name:       s.Name,
					Path:       s.Path,
					Kind:       string(s.Kind),
					ChangeKind: cf.ChangeKind,
				})
			}
		} else {
			// File-level fallback when no surfaces are identified.
			areaMap[area] = append(areaMap[area], ChangedSurface{
				SurfaceID:  "file:" + cf.Path,
				Name:       filepath.Base(cf.Path),
				Path:       cf.Path,
				Kind:       "file",
				ChangeKind: cf.ChangeKind,
			})
		}
	}

	// Convert to sorted slice.
	var areas []ChangedArea
	for area, surfaces := range areaMap {
		sort.Slice(surfaces, func(i, j int) bool {
			return surfaces[i].SurfaceID < surfaces[j].SurfaceID
		})
		areas = append(areas, ChangedArea{
			Area:     area,
			Surfaces: surfaces,
		})
	}
	sort.Slice(areas, func(i, j int) bool {
		return areas[i].Area < areas[j].Area
	})

	return areas
}

// mapAffectedBehaviors finds behavior surfaces that contain changed code surfaces.
func mapAffectedBehaviors(changedAreas []ChangedArea, snap *models.TestSuiteSnapshot) []AffectedBehavior {
	if snap == nil || len(snap.BehaviorSurfaces) == 0 || len(changedAreas) == 0 {
		return nil
	}

	// Build set of changed surface IDs.
	changedSurfaceIDs := map[string]bool{}
	for _, area := range changedAreas {
		for _, cs := range area.Surfaces {
			changedSurfaceIDs[cs.SurfaceID] = true
		}
	}

	var affected []AffectedBehavior

	for _, bs := range snap.BehaviorSurfaces {
		changedCount := 0
		for _, sid := range bs.CodeSurfaceIDs {
			if changedSurfaceIDs[sid] {
				changedCount++
			}
		}
		if changedCount > 0 {
			affected = append(affected, AffectedBehavior{
				BehaviorID:          bs.BehaviorID,
				Label:               bs.Label,
				Kind:                string(bs.Kind),
				ChangedSurfaceCount: changedCount,
				TotalSurfaceCount:   len(bs.CodeSurfaceIDs),
			})
		}
	}

	sort.Slice(affected, func(i, j int) bool {
		return affected[i].BehaviorID < affected[j].BehaviorID
	})

	return affected
}

// computeReasonCategories counts impacted tests by reason category.
func computeReasonCategories(tests []ImpactedTest) ReasonCategories {
	var cats ReasonCategories
	for _, t := range tests {
		switch {
		case t.IsDirectlyChanged:
			cats.DirectlyChanged++
		case t.ImpactConfidence == ConfidenceExact:
			cats.DirectDependency++
		case t.Relevance == "in same directory tree as changed code":
			cats.DirectoryProximity++
		default:
			// Fixture/helper-mediated or other inferred paths.
			cats.FixtureDependency++
		}
	}
	return cats
}

// computeCoverageConfidence derives an overall coverage confidence band
// from the protective test set and impacted units.
func computeCoverageConfidence(result *ImpactResult) string {
	if len(result.ImpactedUnits) == 0 && len(result.ImpactedTests) == 0 {
		return "low" // no data → conservative
	}

	if result.ProtectiveSet == nil || len(result.ProtectiveSet.Tests) == 0 {
		return "low"
	}

	// Count exact-confidence tests.
	exactCount := 0
	for _, t := range result.ProtectiveSet.Tests {
		if t.ImpactConfidence == ConfidenceExact {
			exactCount++
		}
	}

	totalTests := len(result.ProtectiveSet.Tests)
	if totalTests == 0 {
		return "low"
	}

	ratio := float64(exactCount) / float64(totalTests)
	if ratio >= 0.7 {
		return "high"
	}
	if ratio >= 0.3 {
		return "medium"
	}
	return "low"
}

// computeFallbackInfo determines what fallback strategy was used.
func computeFallbackInfo(result *ImpactResult) FallbackInfo {
	if result.ProtectiveSet == nil {
		return FallbackInfo{Level: "none"}
	}

	switch result.ProtectiveSet.SetKind {
	case "exact":
		return FallbackInfo{Level: "none"}
	case "near_minimal":
		return FallbackInfo{
			Level:           "package",
			Reason:          "no exact coverage lineage; structural heuristics used",
			AdditionalTests: len(result.ProtectiveSet.Tests),
		}
	case "fallback_broad":
		return FallbackInfo{
			Level:           "all",
			Reason:          "no impacted tests identified; full suite fallback",
			AdditionalTests: len(result.ProtectiveSet.Tests),
		}
	default:
		return FallbackInfo{Level: "none"}
	}
}

// inferArea derives a domain area label from a file path.
// Uses the deepest meaningful directory segment.
func inferArea(path string) string {
	dir := filepath.Dir(path)
	dir = filepath.ToSlash(dir)

	// Skip common non-semantic prefixes.
	for _, prefix := range []string{"src/", "lib/", "internal/", "pkg/", "app/"} {
		if strings.HasPrefix(dir, prefix) {
			dir = dir[len(prefix):]
			break
		}
	}

	// Use the first path segment as the area.
	if idx := strings.Index(dir, "/"); idx > 0 {
		return dir[:idx]
	}
	if dir == "." || dir == "" {
		return "root"
	}
	return dir
}
