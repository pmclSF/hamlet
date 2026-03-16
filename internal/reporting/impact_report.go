package reporting

import (
	"fmt"
	"io"
	"strings"

	"github.com/pmclSF/terrain/internal/impact"
)

// RenderImpactReport writes a human-readable impact analysis report.
func RenderImpactReport(w io.Writer, result *impact.ImpactResult) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	line("Terrain Impact Analysis")
	line(strings.Repeat("=", 60))
	blank()

	// Summary
	line("Summary: %s", result.Summary)
	blank()

	// Changed areas
	if len(result.ChangedAreas) > 0 {
		line("Changed areas:")
		for _, area := range result.ChangedAreas {
			for _, s := range area.Surfaces {
				line("  %-22s %s (%s)", area.Area, s.Path, s.ChangeKind)
			}
		}
		blank()
	}

	// Affected behaviors
	if len(result.AffectedBehaviors) > 0 {
		line("Affected behaviors:")
		for _, ab := range result.AffectedBehaviors {
			line("  %-30s %d/%d surfaces changed", ab.Label, ab.ChangedSurfaceCount, ab.TotalSurfaceCount)
		}
		blank()
	}

	// Impacted tests count
	if len(result.ImpactedTests) > 0 {
		line("Impacted tests:          %d", len(result.ImpactedTests))
	}

	// Coverage confidence
	line("Coverage confidence:     %s", strings.Title(result.CoverageConfidence))

	// PR risk
	line("PR risk:                 %s", strings.ToUpper(result.Posture.Band))
	blank()

	// Reason categories
	cats := result.ReasonCategories
	if cats.DirectDependency+cats.FixtureDependency+cats.DirectlyChanged+cats.DirectoryProximity > 0 {
		line("Reason categories:")
		if cats.DirectDependency > 0 {
			line("  Direct code dependency:  %d", cats.DirectDependency)
		}
		if cats.FixtureDependency > 0 {
			line("  Fixture dependency:      %d", cats.FixtureDependency)
		}
		if cats.DirectlyChanged > 0 {
			line("  Directly changed:        %d", cats.DirectlyChanged)
		}
		if cats.DirectoryProximity > 0 {
			line("  Directory proximity:     %d", cats.DirectoryProximity)
		}
		blank()
	}

	// Change-risk posture dimensions
	line("Change-Risk Posture: %s", strings.ToUpper(result.Posture.Band))
	line("  %s", result.Posture.Explanation)
	if len(result.Posture.Dimensions) > 0 {
		for _, d := range result.Posture.Dimensions {
			line("  %-20s %s", d.Name+":", d.Band)
		}
	}
	blank()

	// Selected protective tests
	if len(result.SelectedTests) > 0 {
		line("Recommended Tests (%d)", len(result.SelectedTests))
		line(strings.Repeat("-", 60))
		for _, t := range result.SelectedTests {
			conf := ""
			if t.ImpactConfidence != "" {
				conf = fmt.Sprintf(" [%s]", t.ImpactConfidence)
			}
			line("  %s%s", t.Path, conf)
			if t.Relevance != "" {
				line("    %s", t.Relevance)
			}
		}
		blank()
	}

	// Impacted scenarios (AI/eval)
	if len(result.ImpactedScenarios) > 0 {
		line("Impacted Scenarios (%d)", len(result.ImpactedScenarios))
		line(strings.Repeat("-", 60))
		for _, sc := range result.ImpactedScenarios {
			conf := ""
			if sc.ImpactConfidence != "" {
				conf = fmt.Sprintf(" [%s]", sc.ImpactConfidence)
			}
			label := sc.Name
			if sc.Category != "" {
				label += " (" + sc.Category + ")"
			}
			line("  %s%s", label, conf)
			line("    %s", sc.Relevance)
		}
		blank()
	}

	// Fallback
	if result.Fallback.Level != "none" && result.Fallback.Level != "" {
		line("Fallback:")
		line("  Level: %s", result.Fallback.Level)
		if result.Fallback.Reason != "" {
			line("  Reason: %s", result.Fallback.Reason)
		}
		if result.Fallback.AdditionalTests > 0 {
			line("  Additional tests: %d", result.Fallback.AdditionalTests)
		}
		blank()
	}

	// Protection gaps
	if len(result.ProtectionGaps) > 0 {
		line("Protection Gaps")
		line(strings.Repeat("-", 60))
		for _, gap := range result.ProtectionGaps {
			line("  [%s] %s", gap.Severity, gap.Explanation)
			if gap.SuggestedAction != "" {
				line("    Action: %s", gap.SuggestedAction)
			}
		}
		blank()
	}

	// Impacted owners
	if len(result.ImpactedOwners) > 0 {
		line("Impacted Owners: %s", strings.Join(result.ImpactedOwners, ", "))
		blank()
	}

	// Limitations
	if len(result.Limitations) > 0 {
		line("Limitations")
		line(strings.Repeat("-", 60))
		for _, lim := range result.Limitations {
			line("  * %s", lim)
		}
		blank()
	}

	// Next steps
	line("Next steps:")
	line("  terrain impact --show selected   view protective test set with reasoning")
	line("  terrain impact --show graph       see dependency graph")
	line("  terrain impact --json             machine-readable impact data")
	blank()
}
