package portfolio

import (
	"fmt"
	"sort"
)

// detectRedundancy finds pairs of test assets that cover substantially
// overlapping code surfaces. Two tests are redundancy candidates when:
//   - They share ≥70% of their covered code units
//   - Both have coverage data
//
// Redundancy candidates are not bugs — they are investment signals.
// The team may choose to keep both for different reasons (speed, isolation).
func detectRedundancy(assets []TestAsset) []Finding {
	// Build index of assets with coverage data.
	type assetUnits struct {
		idx   int
		units map[string]bool
	}
	var covered []assetUnits
	for i, a := range assets {
		if !a.HasCoverageData || a.CoveredUnitCount == 0 {
			continue
		}
		units := map[string]bool{}
		for _, m := range a.CoveredModules {
			units[m] = true
		}
		// Use path as a proxy unit if modules are empty.
		if len(units) == 0 {
			units[a.Path] = true
		}
		covered = append(covered, assetUnits{idx: i, units: units})
	}

	if len(covered) < 2 {
		return nil
	}

	// Compare all pairs. O(n²) is acceptable for typical suite sizes.
	seen := map[string]bool{}
	var findings []Finding

	for i := 0; i < len(covered); i++ {
		for j := i + 1; j < len(covered); j++ {
			a := covered[i]
			b := covered[j]

			overlap := intersectionCount(a.units, b.units)
			if overlap == 0 {
				continue
			}

			smaller := len(a.units)
			if len(b.units) < smaller {
				smaller = len(b.units)
			}
			if smaller == 0 {
				continue
			}

			ratio := float64(overlap) / float64(smaller)
			if ratio < 0.70 {
				continue
			}

			// Deduplicate by sorted path pair.
			pathA := assets[a.idx].Path
			pathB := assets[b.idx].Path
			if pathA > pathB {
				pathA, pathB = pathB, pathA
			}
			key := pathA + "|" + pathB
			if seen[key] {
				continue
			}
			seen[key] = true

			confidence := ConfidenceModerate
			if ratio >= 0.90 {
				confidence = ConfidenceHigh
			}

			findings = append(findings, Finding{
				Type:         FindingRedundancyCandidate,
				Path:         assets[a.idx].Path,
				RelatedPaths: []string{assets[b.idx].Path},
				Owner:        assets[a.idx].Owner,
				Confidence:   confidence,
				Explanation: fmt.Sprintf(
					"%s and %s cover %.0f%% overlapping modules (%d shared).",
					assets[a.idx].Path, assets[b.idx].Path, ratio*100, overlap,
				),
				SuggestedAction: "Consider consolidating or clarifying distinct purpose for each test.",
				Metadata: map[string]any{
					"overlapRatio": ratio,
					"sharedCount":  overlap,
				},
			})
		}
	}

	// Sort by overlap ratio descending.
	sort.Slice(findings, func(i, j int) bool {
		ri := findings[i].Metadata["overlapRatio"].(float64)
		rj := findings[j].Metadata["overlapRatio"].(float64)
		if ri != rj {
			return ri > rj
		}
		return findings[i].Path < findings[j].Path
	})

	return findings
}

func intersectionCount(a, b map[string]bool) int {
	count := 0
	// Iterate over the smaller set.
	if len(a) > len(b) {
		a, b = b, a
	}
	for k := range a {
		if b[k] {
			count++
		}
	}
	return count
}
