package portfolio

import (
	"fmt"
	"sort"
)

// detectLowValueHighCost finds tests that are expensive to run but
// protect little surface area. These are prime candidates for
// optimization, replacement, or removal.
//
// A test is low-value-high-cost when:
//   - CostClass is high AND BreadthClass is narrow or unknown
//   - OR RuntimeMs ≥ 10000 AND CoveredUnitCount ≤ 1
//   - OR RetryRate ≥ 0.3 AND CoveredUnitCount ≤ 2
func detectLowValueHighCost(assets []TestAsset) []Finding {
	var findings []Finding

	for _, a := range assets {
		if !isLowValueHighCost(a) {
			continue
		}

		confidence := ConfidenceLow
		if a.HasRuntimeData && a.HasCoverageData {
			confidence = ConfidenceHigh
		} else if a.HasRuntimeData || a.HasCoverageData {
			confidence = ConfidenceModerate
		}

		explanation := buildLVHCExplanation(a)

		findings = append(findings, Finding{
			Type:            FindingLowValueHighCost,
			Path:            a.Path,
			Owner:           a.Owner,
			Confidence:      confidence,
			Explanation:     explanation,
			SuggestedAction: "Review whether this test can be optimized, scoped down, or replaced with faster alternatives.",
			Metadata: map[string]any{
				"runtimeMs":    a.RuntimeMs,
				"retryRate":    a.RetryRate,
				"costClass":    string(a.CostClass),
				"breadthClass": string(a.BreadthClass),
				"unitCount":    a.CoveredUnitCount,
			},
		})
	}

	// Sort by runtime descending (most expensive first).
	sort.Slice(findings, func(i, j int) bool {
		ri := findings[i].Metadata["runtimeMs"].(float64)
		rj := findings[j].Metadata["runtimeMs"].(float64)
		if ri != rj {
			return ri > rj
		}
		return findings[i].Path < findings[j].Path
	})

	return findings
}

func isLowValueHighCost(a TestAsset) bool {
	// High cost + narrow/unknown breadth.
	if a.CostClass == CostHigh && (a.BreadthClass == BreadthNarrow || a.BreadthClass == BreadthUnknown) {
		return true
	}
	// Very slow + almost no coverage.
	if a.RuntimeMs >= 10000 && a.CoveredUnitCount <= 1 {
		return true
	}
	// High retry rate + minimal coverage.
	if a.RetryRate >= 0.3 && a.CoveredUnitCount <= 2 {
		return true
	}
	return false
}

func buildLVHCExplanation(a TestAsset) string {
	if a.HasRuntimeData && a.HasCoverageData {
		return fmt.Sprintf(
			"%s costs %.0fms with %.0f%% retry rate but covers only %d unit(s).",
			a.Path, a.RuntimeMs, a.RetryRate*100, a.CoveredUnitCount,
		)
	}
	if a.HasRuntimeData {
		return fmt.Sprintf(
			"%s costs %.0fms with %.0f%% retry rate; coverage data unavailable to confirm value.",
			a.Path, a.RuntimeMs, a.RetryRate*100,
		)
	}
	if a.HasCoverageData {
		return fmt.Sprintf(
			"%s is classified as %s cost (%s type) but covers only %d unit(s).",
			a.Path, string(a.CostClass), a.TestType, a.CoveredUnitCount,
		)
	}
	return fmt.Sprintf(
		"%s is classified as %s cost with %s breadth; limited data for precise assessment.",
		a.Path, string(a.CostClass), string(a.BreadthClass),
	)
}

// detectHighLeverage finds tests that efficiently protect important
// surface area. These are the portfolio's best assets — tests that
// cover many exported units across multiple modules at low cost.
//
// A test is high-leverage when:
//   - BreadthClass is moderate or broad
//   - CostClass is low or moderate
//   - ExportedUnitsCovered ≥ 3
//   - Has coverage data
func detectHighLeverage(assets []TestAsset) []Finding {
	var findings []Finding

	for _, a := range assets {
		if !isHighLeverage(a) {
			continue
		}

		confidence := ConfidenceModerate
		if a.HasRuntimeData && a.HasCoverageData {
			confidence = ConfidenceHigh
		}

		findings = append(findings, Finding{
			Type:       FindingHighLeverage,
			Path:       a.Path,
			Owner:      a.Owner,
			Confidence: confidence,
			Explanation: fmt.Sprintf(
				"%s covers %d exported unit(s) across %d module(s) at %s cost.",
				a.Path, a.ExportedUnitsCovered, len(a.CoveredModules), string(a.CostClass),
			),
			SuggestedAction: "Protect this test from degradation; it provides outsized value.",
			Metadata: map[string]any{
				"exportedUnits": a.ExportedUnitsCovered,
				"moduleCount":   len(a.CoveredModules),
				"costClass":     string(a.CostClass),
				"runtimeMs":     a.RuntimeMs,
			},
		})
	}

	// Sort by exported units descending (most valuable first).
	sort.Slice(findings, func(i, j int) bool {
		ei := findings[i].Metadata["exportedUnits"].(int)
		ej := findings[j].Metadata["exportedUnits"].(int)
		if ei != ej {
			return ei > ej
		}
		return findings[i].Path < findings[j].Path
	})

	return findings
}

func isHighLeverage(a TestAsset) bool {
	if !a.HasCoverageData {
		return false
	}
	if a.ExportedUnitsCovered < 3 {
		return false
	}
	if a.BreadthClass != BreadthModerate && a.BreadthClass != BreadthBroad {
		return false
	}
	if a.CostClass == CostHigh {
		return false
	}
	return true
}
