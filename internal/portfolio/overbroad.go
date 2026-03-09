package portfolio

import (
	"fmt"
	"sort"
)

// detectOverbroad finds tests that span an unusually large surface area
// relative to the suite. Overbroad tests are expensive to maintain,
// slow to diagnose, and often mask where failures originate.
//
// A test is overbroad when:
//   - It touches ≥3 modules AND ≥2 owners, OR
//   - It touches ≥5 modules
//   - AND it is classified as BreadthBroad
func detectOverbroad(assets []TestAsset) []Finding {
	var findings []Finding

	for _, a := range assets {
		if a.BreadthClass != BreadthBroad || !a.HasCoverageData {
			continue
		}

		moduleCount := len(a.CoveredModules)
		ownerCount := len(a.OwnersCovered)

		// Only flag the most extreme cases.
		if moduleCount < 3 {
			continue
		}

		confidence := ConfidenceModerate
		if moduleCount >= 5 || (moduleCount >= 3 && ownerCount >= 3) {
			confidence = ConfidenceHigh
		}

		findings = append(findings, Finding{
			Type:       FindingOverbroad,
			Path:       a.Path,
			Owner:      a.Owner,
			Confidence: confidence,
			Explanation: fmt.Sprintf(
				"%s spans %d modules and %d owner(s), making failures hard to diagnose.",
				a.Path, moduleCount, ownerCount,
			),
			SuggestedAction: "Consider splitting into focused tests per module or extracting shared setup.",
			Metadata: map[string]any{
				"moduleCount": moduleCount,
				"ownerCount":  ownerCount,
				"unitCount":   a.CoveredUnitCount,
			},
		})
	}

	// Sort by module count descending.
	sort.Slice(findings, func(i, j int) bool {
		mi := findings[i].Metadata["moduleCount"].(int)
		mj := findings[j].Metadata["moduleCount"].(int)
		if mi != mj {
			return mi > mj
		}
		return findings[i].Path < findings[j].Path
	})

	return findings
}
