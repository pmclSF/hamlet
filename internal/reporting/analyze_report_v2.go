package reporting

import (
	"fmt"
	"io"
	"strings"

	"github.com/pmclSF/terrain/internal/analyze"
)

// RenderAnalyzeReportV2 writes a human-readable analysis report from the
// structured analyze.Report. This is the "wow moment" first-run output.
func RenderAnalyzeReportV2(w io.Writer, r *analyze.Report) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	// Header
	line("Terrain — Test Suite Analysis")
	line(strings.Repeat("=", 60))
	blank()

	// Repository profile (the "wow" section — leads with insight)
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

	// Tests detected
	line("Tests Detected")
	line(strings.Repeat("-", 60))
	line("  Test files:     %d", r.TestsDetected.TestFileCount)
	line("  Test cases:     %d", r.TestsDetected.TestCaseCount)
	line("  Code units:     %d", r.TestsDetected.CodeUnitCount)
	if len(r.TestsDetected.Frameworks) > 0 {
		line("  Frameworks:")
		for _, fw := range r.TestsDetected.Frameworks {
			typeBadge := ""
			if fw.Type != "" && fw.Type != "unknown" {
				typeBadge = fmt.Sprintf(" [%s]", fw.Type)
			}
			line("    %-20s %4d files%s", fw.Name, fw.FileCount, typeBadge)
		}
	}
	blank()

	// Coverage confidence summary
	if r.CoverageConfidence.TotalFiles > 0 {
		line("Coverage Confidence")
		line(strings.Repeat("-", 60))
		total := r.CoverageConfidence.TotalFiles
		line("  High:    %d files (%d%%)", r.CoverageConfidence.HighCount, pct(r.CoverageConfidence.HighCount, total))
		line("  Medium:  %d files (%d%%)", r.CoverageConfidence.MediumCount, pct(r.CoverageConfidence.MediumCount, total))
		line("  Low:     %d files (%d%%)", r.CoverageConfidence.LowCount, pct(r.CoverageConfidence.LowCount, total))
		blank()
	}

	// Duplicate clusters
	if r.DuplicateClusters.ClusterCount > 0 {
		line("Duplicate Clusters")
		line(strings.Repeat("-", 60))
		line("  Clusters:        %d", r.DuplicateClusters.ClusterCount)
		line("  Redundant tests: %d", r.DuplicateClusters.RedundantTestCount)
		if r.DuplicateClusters.HighestSimilarity > 0 {
			line("  Max similarity:  %.0f%%", r.DuplicateClusters.HighestSimilarity*100)
		}
		blank()
	}

	// High-fanout fixtures/helpers
	if r.HighFanout.FlaggedCount > 0 {
		line("High-Fanout Nodes")
		line(strings.Repeat("-", 60))
		line("  Flagged: %d (threshold: %d)", r.HighFanout.FlaggedCount, r.HighFanout.Threshold)
		for _, n := range r.HighFanout.TopNodes {
			line("    %s  (%s, %d dependents)", n.Path, n.NodeType, n.TransitiveFanout)
		}
		blank()
	}

	// Skipped test burden
	if r.SkippedTestBurden.SkippedCount > 0 {
		line("Skipped Test Burden")
		line(strings.Repeat("-", 60))
		line("  Skipped: %d / %d tests (%.0f%%)",
			r.SkippedTestBurden.SkippedCount,
			r.SkippedTestBurden.TotalTests,
			r.SkippedTestBurden.SkipRatio*100)
		blank()
	}

	// Weak coverage areas
	if len(r.WeakCoverageAreas) > 0 {
		line("Weak Coverage Areas")
		line(strings.Repeat("-", 60))
		for _, a := range r.WeakCoverageAreas {
			if a.TestCount == 0 {
				line("  %-40s no structural coverage", a.Path)
			} else {
				line("  %-40s %d test(s)", a.Path, a.TestCount)
			}
		}
		blank()
	}

	// CI optimization potential
	if r.CIOptimization.Recommendation != "" {
		line("CI Optimization Potential")
		line(strings.Repeat("-", 60))
		if r.CIOptimization.DuplicateTestsRemovable > 0 {
			line("  Duplicate tests removable:  %d", r.CIOptimization.DuplicateTestsRemovable)
		}
		if r.CIOptimization.SkippedTestsReviewable > 0 {
			line("  Skipped tests reviewable:   %d", r.CIOptimization.SkippedTestsReviewable)
		}
		if r.CIOptimization.HighFanoutNodes > 0 {
			line("  High-fanout nodes:          %d", r.CIOptimization.HighFanoutNodes)
		}
		line("  %s", r.CIOptimization.Recommendation)
		blank()
	}

	// Top insight (the headline)
	line("Top Insight")
	line(strings.Repeat("-", 60))
	line("  %s", r.TopInsight)
	blank()

	// Risk posture
	if len(r.RiskPosture) > 0 {
		line("Risk Posture")
		line(strings.Repeat("-", 60))
		for _, d := range r.RiskPosture {
			line("  %-24s %s", d.Dimension+":", strings.ToUpper(d.Band))
		}
		blank()
	}

	// Signals
	line("Signals: %d total", r.SignalSummary.Total)
	if r.SignalSummary.Total > 0 {
		parts := []string{}
		if r.SignalSummary.Critical > 0 {
			parts = append(parts, fmt.Sprintf("%d critical", r.SignalSummary.Critical))
		}
		if r.SignalSummary.High > 0 {
			parts = append(parts, fmt.Sprintf("%d high", r.SignalSummary.High))
		}
		if r.SignalSummary.Medium > 0 {
			parts = append(parts, fmt.Sprintf("%d medium", r.SignalSummary.Medium))
		}
		if r.SignalSummary.Low > 0 {
			parts = append(parts, fmt.Sprintf("%d low", r.SignalSummary.Low))
		}
		if len(parts) > 0 {
			line("  %s", strings.Join(parts, ", "))
		}
	}
	blank()

	// Behavior redundancy
	if r.BehaviorRedundancy != nil && len(r.BehaviorRedundancy.Clusters) > 0 {
		br := r.BehaviorRedundancy
		line("Behavior Redundancy")
		line(strings.Repeat("-", 60))
		line("  Redundant tests:  %d across %d clusters", br.RedundantTestCount, len(br.Clusters))
		if br.CrossFrameworkOverlaps > 0 {
			line("  Cross-framework:  %d cluster(s)", br.CrossFrameworkOverlaps)
		}
		limit := 5
		if len(br.Clusters) < limit {
			limit = len(br.Clusters)
		}
		for _, c := range br.Clusters[:limit] {
			fwLabel := ""
			if len(c.Frameworks) > 0 {
				fwLabel = fmt.Sprintf(" [%s]", strings.Join(c.Frameworks, ", "))
			}
			line("  [%s] %d tests, %d shared surfaces (%.0f%% confidence)%s",
				c.OverlapKind, len(c.Tests), len(c.SharedSurfaces), c.Confidence*100, fwLabel)
			line("         %s", c.Rationale)
		}
		if len(br.Clusters) > 5 {
			line("  ... and %d more cluster(s)", len(br.Clusters)-5)
		}
		blank()
	}

	// Stability clusters
	if r.StabilityClusters != nil && len(r.StabilityClusters.Clusters) > 0 {
		sc := r.StabilityClusters
		line("Stability Clusters")
		line(strings.Repeat("-", 60))
		line("  Unstable tests:  %d (%d clustered around shared dependencies)", sc.UnstableTestCount, sc.ClusteredTestCount)
		limit := 5
		if len(sc.Clusters) < limit {
			limit = len(sc.Clusters)
		}
		for _, c := range sc.Clusters[:limit] {
			line("  [%s] %s  (%d tests, %.0f%% confidence)",
				c.CauseKind, c.CauseName, len(c.Members), c.Confidence*100)
			line("         %s", c.Remediation)
		}
		if len(sc.Clusters) > 5 {
			line("  ... and %d more cluster(s)", len(sc.Clusters)-5)
		}
		blank()
	}

	// Matrix coverage
	if r.MatrixCoverage != nil && len(r.MatrixCoverage.Classes) > 0 {
		mc := r.MatrixCoverage
		line("Matrix Coverage")
		line(strings.Repeat("-", 60))
		line("  Classes analyzed:  %d", mc.ClassesAnalyzed)
		line("  Test files:        %d", mc.TestsAnalyzed)
		for _, cc := range mc.Classes {
			line("  [%s] %s: %d/%d members covered (%.0f%%)",
				cc.Dimension, cc.ClassName, cc.CoveredMembers, cc.TotalMembers, cc.CoverageRatio*100)
		}
		if len(mc.Gaps) > 0 {
			line("  Gaps:  %d uncovered members", len(mc.Gaps))
			limit := 3
			if len(mc.Gaps) < limit {
				limit = len(mc.Gaps)
			}
			for _, gap := range mc.Gaps[:limit] {
				line("    - %s (%s/%s)", gap.MemberName, gap.ClassName, gap.Dimension)
			}
			if len(mc.Gaps) > 3 {
				line("    ... and %d more", len(mc.Gaps)-3)
			}
		}
		if len(mc.Recommendations) > 0 {
			line("  Recommendations:  %d devices/environments to consider", len(mc.Recommendations))
			limit := 3
			if len(mc.Recommendations) < limit {
				limit = len(mc.Recommendations)
			}
			for _, rec := range mc.Recommendations[:limit] {
				line("    %d. %s — %s", rec.Priority, rec.MemberName, rec.Reason)
			}
		}
		blank()
	}

	// Manual coverage overlay
	if r.ManualCoverage != nil && r.ManualCoverage.ArtifactCount > 0 {
		line("Manual Coverage Overlay")
		line(strings.Repeat("-", 60))
		line("  Artifacts:  %d (not executable — supplements CI coverage)", r.ManualCoverage.ArtifactCount)
		if len(r.ManualCoverage.BySource) > 0 {
			parts := []string{}
			for src, count := range r.ManualCoverage.BySource {
				parts = append(parts, fmt.Sprintf("%s: %d", src, count))
			}
			line("  Sources:    %s", strings.Join(parts, ", "))
		}
		if len(r.ManualCoverage.ByCriticality) > 0 {
			parts := []string{}
			for crit, count := range r.ManualCoverage.ByCriticality {
				parts = append(parts, fmt.Sprintf("%s: %d", crit, count))
			}
			line("  Criticality: %s", strings.Join(parts, ", "))
		}
		if len(r.ManualCoverage.Areas) > 0 {
			line("  Areas:      %s", strings.Join(r.ManualCoverage.Areas, ", "))
		}
		if r.ManualCoverage.StaleCount > 0 {
			line("  Stale:      %d artifact(s) have no recent execution date", r.ManualCoverage.StaleCount)
		}
		blank()
	}

	// Edge cases
	if len(r.EdgeCases) > 0 {
		line("Edge Cases")
		line(strings.Repeat("-", 60))
		for _, ec := range r.EdgeCases {
			line("  [%s] %s", ec.Severity, ec.Description)
		}
		blank()
	}

	// Policy
	if r.Policy != nil && len(r.Policy.Recommendations) > 0 {
		line("Policy Recommendations")
		line(strings.Repeat("-", 60))
		for _, pr := range r.Policy.Recommendations {
			line("  • %s", pr)
		}
		blank()
	}

	// Data completeness
	line("Data Completeness")
	line(strings.Repeat("-", 60))
	for _, ds := range r.DataCompleteness {
		line("  [%-9s] %s", completenessBadge(ds.Available), ds.Name)
	}
	blank()

	// Limitations
	if len(r.Limitations) > 0 {
		line("Limitations")
		line(strings.Repeat("-", 60))
		for _, lim := range r.Limitations {
			line("  * %s", lim)
		}
		blank()
	}

	// Next steps
	line("Next steps:")
	line("  terrain analyze --verbose           show all findings")
	line("  terrain impact                      what tests matter for this change?")
	line("  terrain explain <test-path>         why was a test selected?")
	line("  terrain insights                    deeper analysis with recommendations")
	blank()
}

func pct(n, total int) int {
	if total == 0 {
		return 0
	}
	return 100 * n / total
}
