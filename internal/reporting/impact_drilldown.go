package reporting

import (
	"fmt"
	"io"
	"strings"

	"github.com/pmclSF/hamlet/internal/impact"
)

// RenderImpactUnits writes a focused view of impacted code units.
func RenderImpactUnits(w io.Writer, result *impact.ImpactResult) {
	line := func(format string, args ...any) { fmt.Fprintf(w, format+"\n", args...) }
	blank := func() { fmt.Fprintln(w) }

	line("Impacted Code Units (%d)", len(result.ImpactedUnits))
	line(strings.Repeat("=", 60))
	blank()

	if len(result.ImpactedUnits) == 0 {
		line("  No impacted code units identified.")
		blank()
		return
	}

	for _, iu := range result.ImpactedUnits {
		exported := ""
		if iu.Exported {
			exported = " [exported]"
		}
		line("  %-30s %s  protection: %s%s", iu.Name, iu.ChangeKind, iu.ProtectionStatus, exported)
		if iu.Owner != "" {
			line("    Owner: %s", iu.Owner)
		}
		line("    Confidence: %s", iu.ImpactConfidence)
		if len(iu.CoveringTests) > 0 {
			line("    Covering tests: %s", strings.Join(iu.CoveringTests, ", "))
		}
		line("    Path: %s", iu.Path)
		blank()
	}
}

// RenderImpactGaps writes a focused view of protection gaps.
func RenderImpactGaps(w io.Writer, result *impact.ImpactResult) {
	line := func(format string, args ...any) { fmt.Fprintf(w, format+"\n", args...) }
	blank := func() { fmt.Fprintln(w) }

	line("Protection Gaps (%d)", len(result.ProtectionGaps))
	line(strings.Repeat("=", 60))
	blank()

	if len(result.ProtectionGaps) == 0 {
		line("  No protection gaps identified. All changed code appears covered.")
		blank()
		return
	}

	// Group by severity.
	bySeverity := map[string][]impact.ProtectionGap{}
	for _, gap := range result.ProtectionGaps {
		bySeverity[gap.Severity] = append(bySeverity[gap.Severity], gap)
	}

	for _, sev := range []string{"high", "medium", "low"} {
		gaps := bySeverity[sev]
		if len(gaps) == 0 {
			continue
		}
		line("  %s severity (%d)", strings.ToUpper(sev), len(gaps))
		line("  " + strings.Repeat("-", 40))
		for _, gap := range gaps {
			line("    [%s] %s", gap.GapType, gap.Explanation)
			line("      Path: %s", gap.Path)
			if gap.SuggestedAction != "" {
				line("      Action: %s", gap.SuggestedAction)
			}
		}
		blank()
	}
}

// RenderImpactTests writes a focused view of impacted and selected tests.
func RenderImpactTests(w io.Writer, result *impact.ImpactResult) {
	line := func(format string, args ...any) { fmt.Fprintf(w, format+"\n", args...) }
	blank := func() { fmt.Fprintln(w) }

	line("Impacted Tests (%d total, %d selected)", len(result.ImpactedTests), len(result.SelectedTests))
	line(strings.Repeat("=", 60))
	blank()

	if len(result.SelectedTests) > 0 {
		line("  Recommended (run these first)")
		line("  " + strings.Repeat("-", 40))
		for _, t := range result.SelectedTests {
			changed := ""
			if t.IsDirectlyChanged {
				changed = " [changed]"
			}
			line("    %s  [%s]%s", t.Path, t.ImpactConfidence, changed)
			line("      %s", t.Relevance)
			if len(t.CoversUnits) > 0 {
				line("      Covers: %s", strings.Join(t.CoversUnits, ", "))
			}
		}
		blank()
	}

	// Show non-selected tests if any.
	selectedPaths := map[string]bool{}
	for _, t := range result.SelectedTests {
		selectedPaths[t.Path] = true
	}
	var other []impact.ImpactedTest
	for _, t := range result.ImpactedTests {
		if !selectedPaths[t.Path] {
			other = append(other, t)
		}
	}

	if len(other) > 0 {
		line("  Additional relevant tests")
		line("  " + strings.Repeat("-", 40))
		for _, t := range other {
			line("    %s  [%s]", t.Path, t.ImpactConfidence)
			line("      %s", t.Relevance)
		}
		blank()
	}

	if len(result.ImpactedTests) == 0 {
		line("  No impacted tests identified.")
		blank()
	}
}

// RenderImpactOwners writes a focused view of impacted owners.
func RenderImpactOwners(w io.Writer, result *impact.ImpactResult) {
	line := func(format string, args ...any) { fmt.Fprintf(w, format+"\n", args...) }
	blank := func() { fmt.Fprintln(w) }

	line("Impacted Owners (%d)", len(result.ImpactedOwners))
	line(strings.Repeat("=", 60))
	blank()

	if len(result.ImpactedOwners) == 0 {
		line("  No ownership data available.")
		blank()
		return
	}

	// Group units by owner.
	byOwner := map[string][]impact.ImpactedCodeUnit{}
	for _, iu := range result.ImpactedUnits {
		if iu.Owner != "" {
			byOwner[iu.Owner] = append(byOwner[iu.Owner], iu)
		}
	}

	for _, owner := range result.ImpactedOwners {
		units := byOwner[owner]
		line("  %s (%d unit(s))", owner, len(units))
		line("  " + strings.Repeat("-", 40))
		for _, iu := range units {
			line("    %-30s %s  %s", iu.Name, iu.ProtectionStatus, iu.ChangeKind)
		}
		blank()
	}
}
