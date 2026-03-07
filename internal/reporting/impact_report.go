package reporting

import (
	"fmt"
	"io"
	"strings"

	"github.com/pmclSF/hamlet/internal/impact"
)

// RenderImpactReport writes a human-readable impact analysis report.
func RenderImpactReport(w io.Writer, result *impact.ImpactResult) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	line("Hamlet Impact Analysis")
	line(strings.Repeat("=", 60))
	blank()

	// Summary
	line("Summary: %s", result.Summary)
	blank()

	// Change-risk posture
	line("Change-Risk Posture: %s", strings.ToUpper(result.Posture.Band))
	line("  %s", result.Posture.Explanation)
	if len(result.Posture.Dimensions) > 0 {
		for _, d := range result.Posture.Dimensions {
			line("  %-20s %s", d.Name+":", d.Band)
		}
	}
	blank()

	// Impacted code units
	if len(result.ImpactedUnits) > 0 {
		line("Impacted Code Units")
		line(strings.Repeat("-", 60))
		for _, iu := range result.ImpactedUnits {
			exported := ""
			if iu.Exported {
				exported = " [exported]"
			}
			line("  %-30s %s  %s%s", iu.Name, iu.ChangeKind, iu.ProtectionStatus, exported)
			if iu.Owner != "" {
				line("    Owner: %s", iu.Owner)
			}
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
	line("  hamlet analyze       full repo analysis")
	line("  hamlet posture       evidence-backed posture")
	line("  hamlet impact --json   machine-readable impact data")
	blank()
}
