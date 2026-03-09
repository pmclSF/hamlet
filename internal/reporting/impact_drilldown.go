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

// RenderImpactGraph writes a summary of the impact graph.
func RenderImpactGraph(w io.Writer, result *impact.ImpactResult) {
	line := func(format string, args ...any) { fmt.Fprintf(w, format+"\n", args...) }
	blank := func() { fmt.Fprintln(w) }

	line("Impact Graph")
	line(strings.Repeat("=", 60))
	blank()

	if result.Graph == nil {
		line("  No impact graph available.")
		blank()
		return
	}

	g := result.Graph
	line("  Total edges:      %d", g.Stats.TotalEdges)
	line("  Exact edges:      %d", g.Stats.ExactEdges)
	line("  Inferred edges:   %d", g.Stats.InferredEdges)
	line("  Weak edges:       %d", g.Stats.WeakEdges)
	line("  Connected units:  %d", g.Stats.ConnectedUnits)
	line("  Isolated units:   %d", g.Stats.IsolatedUnits)
	line("  Connected tests:  %d", g.Stats.ConnectedTests)
	blank()

	// Show edges for impacted units.
	if len(result.ImpactedUnits) > 0 {
		line("Edges for impacted units")
		line(strings.Repeat("-", 60))
		for _, iu := range result.ImpactedUnits {
			edges := g.EdgesForUnit(iu.UnitID)
			if len(edges) == 0 {
				line("  %-30s (no edges)", iu.Name)
				continue
			}
			line("  %s", iu.Name)
			for _, e := range edges {
				line("    -> %-40s [%s] %s", e.TargetID, e.Confidence, e.Kind)
			}
		}
		blank()
	}

	line("Next: hamlet impact --show units   view impacted code units")
	blank()
}

// RenderProtectiveSet writes the enhanced protective test set.
func RenderProtectiveSet(w io.Writer, result *impact.ImpactResult) {
	line := func(format string, args ...any) { fmt.Fprintf(w, format+"\n", args...) }
	blank := func() { fmt.Fprintln(w) }

	line("Protective Test Set")
	line(strings.Repeat("=", 60))
	blank()

	if result.ProtectiveSet == nil || len(result.ProtectiveSet.Tests) == 0 {
		line("  No protective tests identified.")
		blank()
		return
	}

	ps := result.ProtectiveSet
	line("  Strategy:   %s", ps.SetKind)
	line("  Tests:      %d", len(ps.Tests))
	line("  Covered:    %d unit(s)", ps.CoveredUnitCount)
	line("  Uncovered:  %d unit(s)", ps.UncoveredUnitCount)
	blank()

	line("  %s", ps.Explanation)
	blank()

	line("Selected Tests")
	line(strings.Repeat("-", 60))
	for _, t := range ps.Tests {
		changed := ""
		if t.IsDirectlyChanged {
			changed = " [changed]"
		}
		line("  %s  [%s]%s", t.Path, t.ImpactConfidence, changed)
		for _, r := range t.Reasons {
			if r.CodeUnitID != "" {
				line("    - %s (%s)", r.Reason, r.CodeUnitID)
			} else {
				line("    - %s", r.Reason)
			}
		}
	}
	blank()

	if ps.UncoveredUnitCount > 0 {
		line("Warning: %d impacted unit(s) have no covering tests in the selected set.", ps.UncoveredUnitCount)
		line("Consider adding tests or running the full suite.")
		blank()
	}

	line("Next: hamlet impact --show gaps   view protection gaps")
	blank()
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
