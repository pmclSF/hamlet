package reporting

import (
	"fmt"
	"io"
	"strings"

	"github.com/pmclSF/terrain/internal/insights"
)

// RenderInsightsReport writes a human-readable health report from the
// structured insights.Report. This replaces the raw dump with a
// prioritized, categorized view.
func RenderInsightsReport(w io.Writer, r *insights.Report) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	// Header with health grade.
	line("Terrain — Test System Health Report")
	line(strings.Repeat("=", 60))
	blank()

	line("Health Grade: %s", r.HealthGrade)
	line("Headline: %s", r.Headline)
	blank()

	// Repository profile.
	line("Repository Profile")
	line(strings.Repeat("-", 60))
	line("  Test volume:          %s", r.RepoProfile.TestVolume)
	line("  CI pressure:          %s", r.RepoProfile.CIPressure)
	line("  Coverage confidence:  %s", r.RepoProfile.CoverageConfidence)
	line("  Redundancy level:     %s", r.RepoProfile.RedundancyLevel)
	line("  Fanout burden:        %s", r.RepoProfile.FanoutBurden)
	if r.RepoProfile.SkipBurden != "" {
		line("  Skip burden:          %s", r.RepoProfile.SkipBurden)
	}
	if r.RepoProfile.FlakeBurden != "" {
		line("  Flake burden:         %s", r.RepoProfile.FlakeBurden)
	}
	if r.RepoProfile.ManualCoveragePresence != "" && r.RepoProfile.ManualCoveragePresence != "none" {
		line("  Manual coverage:      %s", r.RepoProfile.ManualCoveragePresence)
	}
	blank()

	// Findings by category.
	categoryOrder := []insights.Category{
		insights.CategoryReliability,
		insights.CategoryArchitectureDebt,
		insights.CategoryCoverageDebt,
		insights.CategoryOptimization,
	}
	categoryLabels := map[insights.Category]string{
		insights.CategoryReliability:     "Reliability Problems",
		insights.CategoryArchitectureDebt: "Architecture Debt",
		insights.CategoryCoverageDebt:    "Coverage Debt",
		insights.CategoryOptimization:    "Optimization Opportunities",
	}

	for _, cat := range categoryOrder {
		cb, ok := r.CategorySummary[cat]
		if !ok || cb.Count == 0 {
			continue
		}
		label := categoryLabels[cat]
		line("%s (%d)", label, cb.Count)
		line(strings.Repeat("-", 60))

		for _, f := range r.Findings {
			if f.Category != cat {
				continue
			}
			line("  [%s] %s", strings.ToUpper(string(f.Severity)), f.Title)
			if f.Description != "" {
				// Wrap long descriptions.
				line("         %s", f.Description)
			}
		}
		blank()
	}

	// If no findings at all.
	if len(r.Findings) == 0 {
		line("No significant issues detected.")
		blank()
	}

	// Recommendations.
	if len(r.Recommendations) > 0 {
		line("Recommended Actions")
		line(strings.Repeat("-", 60))
		for _, rec := range r.Recommendations {
			badge := categoryBadge(rec.Category)
			line("  %d. [%s] %s", rec.Priority, badge, rec.Action)
			if rec.Rationale != "" {
				line("     why: %s", rec.Rationale)
			}
			if rec.Impact != "" {
				line("     impact: %s", rec.Impact)
			}
		}
		blank()
	}

	// Behavior redundancy.
	if r.BehaviorRedundancy != nil && len(r.BehaviorRedundancy.Clusters) > 0 {
		br := r.BehaviorRedundancy
		wasteful := 0
		for _, c := range br.Clusters {
			if c.OverlapKind == "wasteful" {
				wasteful++
			}
		}
		if wasteful > 0 || br.CrossFrameworkOverlaps > 0 {
			line("Behavior Redundancy")
			line(strings.Repeat("-", 60))
			if wasteful > 0 {
				line("  %d wasteful overlap clusters — tests exercise identical behavior surfaces", wasteful)
			}
			if br.CrossFrameworkOverlaps > 0 {
				line("  %d cross-framework overlaps — review after migration completes", br.CrossFrameworkOverlaps)
			}
			top := br.Clusters[0]
			line("  Top: %d tests share %d surfaces (%s)",
				len(top.Tests), len(top.SharedSurfaces), top.OverlapKind)
			blank()
		}
	}

	// Stability clusters.
	if r.StabilityClusters != nil && len(r.StabilityClusters.Clusters) > 0 {
		sc := r.StabilityClusters
		line("Stability Clusters")
		line(strings.Repeat("-", 60))
		line("  %d unstable tests cluster around %d shared dependencies",
			sc.ClusteredTestCount, len(sc.Clusters))
		limit := 3
		if len(sc.Clusters) < limit {
			limit = len(sc.Clusters)
		}
		for _, c := range sc.Clusters[:limit] {
			line("  [%s] %s  (%d tests)", c.CauseKind, c.CauseName, len(c.Members))
			line("         %s", c.Remediation)
		}
		if len(sc.Clusters) > 3 {
			line("  ... and %d more cluster(s)", len(sc.Clusters)-3)
		}
		blank()
	}

	// Matrix coverage.
	if r.MatrixCoverage != nil && len(r.MatrixCoverage.Classes) > 0 {
		mc := r.MatrixCoverage
		line("Matrix Coverage")
		line(strings.Repeat("-", 60))
		for _, cc := range mc.Classes {
			line("  [%s] %s: %d/%d covered (%.0f%%)",
				cc.Dimension, cc.ClassName, cc.CoveredMembers, cc.TotalMembers, cc.CoverageRatio*100)
		}
		if len(mc.Gaps) > 0 {
			line("  Gaps: %d uncovered members", len(mc.Gaps))
		}
		if len(mc.Recommendations) > 0 {
			line("  Top recommendation: %s — %s", mc.Recommendations[0].MemberName, mc.Recommendations[0].Reason)
		}
		blank()
	}

	// Edge cases.
	if len(r.EdgeCases) > 0 {
		line("Edge Cases")
		line(strings.Repeat("-", 60))
		for _, ec := range r.EdgeCases {
			line("  [%s] %s", ec.Severity, ec.Description)
		}
		blank()
	}

	// Policy.
	if len(r.Policy.Recommendations) > 0 {
		line("Policy Recommendations")
		line(strings.Repeat("-", 60))
		for _, pr := range r.Policy.Recommendations {
			line("  • %s", pr)
		}
		blank()
	}

	// Data completeness.
	line("Data Completeness")
	line(strings.Repeat("-", 60))
	for _, ds := range r.DataCompleteness {
		line("  [%-9s] %s", completenessBadge(ds.Available), ds.Name)
	}
	blank()

	// Limitations.
	if len(r.Limitations) > 0 {
		line("Limitations")
		line(strings.Repeat("-", 60))
		for _, lim := range r.Limitations {
			line("  * %s", lim)
		}
		blank()
	}

	// Next steps.
	line("Next steps:")
	line("  terrain analyze                     full suite analysis")
	line("  terrain impact                      what tests matter for this change?")
	line("  terrain explain <test-path>         why was a test selected?")
	blank()
}

func categoryBadge(c insights.Category) string {
	switch c {
	case insights.CategoryReliability:
		return "reliability"
	case insights.CategoryArchitectureDebt:
		return "architecture"
	case insights.CategoryCoverageDebt:
		return "coverage"
	case insights.CategoryOptimization:
		return "optimization"
	default:
		return string(c)
	}
}
