package reporting

import (
	"fmt"
	"io"
	"strings"

	"github.com/pmclSF/terrain/internal/explain"
)

// RenderTestExplanation writes a human-readable explanation of why a test
// was selected.
func RenderTestExplanation(w io.Writer, te *explain.TestExplanation) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	line("Terrain Explain")
	line(strings.Repeat("=", 60))
	blank()

	// Target metadata.
	line("Target: %s", te.Target.Path)
	if te.Target.Framework != "" {
		line("Framework: %s", te.Target.Framework)
	}
	if te.Target.TestID != "" {
		line("Test ID: %s", te.Target.TestID)
	}
	if te.Target.Owner != "" {
		line("Owner: %s", te.Target.Owner)
	}
	blank()

	// Verdict.
	line("Verdict: %s", te.Verdict)
	blank()

	// Strongest path.
	if te.StrongestPath != nil {
		line("Strongest path (confidence: %.0f%% — %s)", te.StrongestPath.Confidence*100, te.StrongestPath.Band)
		line(strings.Repeat("-", 60))
		renderChain(w, te.StrongestPath)
		blank()
	}

	// Alternative paths.
	if len(te.AlternativePaths) > 0 {
		line("Alternative paths (%d)", len(te.AlternativePaths))
		line(strings.Repeat("-", 60))
		for i, alt := range te.AlternativePaths {
			if i > 0 {
				blank()
			}
			line("  Path %d (confidence: %.0f%% — %s)", i+1, alt.Confidence*100, alt.Band)
			renderChain(w, &alt)
		}
		blank()
	}

	// Confidence summary.
	line("Confidence: %s (%.0f%%)", te.ConfidenceBand, te.Confidence*100)
	line("Reason: %s", reasonCategoryLabel(te.ReasonCategory))
	blank()

	// Covers units.
	if len(te.CoversUnits) > 0 {
		line("Covers %d code unit(s):", len(te.CoversUnits))
		for _, u := range te.CoversUnits {
			line("  %s", u)
		}
		blank()
	}

	// Fallback.
	if te.FallbackUsed != nil {
		line("Fallback: %s", te.FallbackUsed.Level)
		if te.FallbackUsed.Reason != "" {
			line("  %s", te.FallbackUsed.Reason)
		}
		blank()
	}

	// Limitations.
	if len(te.Limitations) > 0 {
		line("Limitations")
		line(strings.Repeat("-", 60))
		for _, lim := range te.Limitations {
			line("  * %s", lim)
		}
		blank()
	}
}

// RenderSelectionExplanation writes a human-readable explanation of the
// overall test selection strategy.
func RenderSelectionExplanation(w io.Writer, sel *explain.SelectionExplanation) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	line("Terrain Explain — Test Selection")
	line(strings.Repeat("=", 60))
	blank()

	line("Summary: %s", sel.Summary)
	blank()

	line("Strategy: %s", sel.Strategy)
	line("Coverage confidence: %s", sel.CoverageConfidence)
	line("Tests selected: %d", sel.TotalSelected)
	if sel.GapCount > 0 {
		line("Protection gaps: %d", sel.GapCount)
	}
	blank()

	// Reason breakdown.
	hasReasons := false
	for _, v := range sel.ReasonBreakdown {
		if v > 0 {
			hasReasons = true
			break
		}
	}
	if hasReasons {
		line("Reason breakdown:")
		for reason, count := range sel.ReasonBreakdown {
			if count > 0 {
				line("  %-24s %d", reasonCategoryLabel(reason)+":", count)
			}
		}
		blank()
	}

	// High confidence tests.
	if len(sel.HighConfidenceTests) > 0 {
		line("High confidence (%d):", len(sel.HighConfidenceTests))
		for _, te := range sel.HighConfidenceTests {
			renderTestSummaryLine(w, &te)
		}
		blank()
	}

	// Medium confidence tests.
	if len(sel.MediumConfidenceTests) > 0 {
		line("Medium confidence (%d):", len(sel.MediumConfidenceTests))
		for _, te := range sel.MediumConfidenceTests {
			renderTestSummaryLine(w, &te)
		}
		blank()
	}

	// Low confidence tests.
	if len(sel.LowConfidenceTests) > 0 {
		line("Low confidence (%d):", len(sel.LowConfidenceTests))
		for _, te := range sel.LowConfidenceTests {
			renderTestSummaryLine(w, &te)
		}
		blank()
	}

	// Fallback.
	if sel.FallbackUsed != nil {
		line("Fallback: %s", sel.FallbackUsed.Level)
		if sel.FallbackUsed.Reason != "" {
			line("  %s", sel.FallbackUsed.Reason)
		}
		blank()
	}

	// Limitations.
	if len(sel.Limitations) > 0 {
		line("Limitations")
		line(strings.Repeat("-", 60))
		for _, lim := range sel.Limitations {
			line("  * %s", lim)
		}
		blank()
	}

	// Next steps.
	line("Next steps:")
	line("  terrain explain <test-path>       explain a specific test")
	line("  terrain impact --show selected     view protective test set")
	line("  terrain impact --json              machine-readable impact data")
	blank()
}

// renderChain renders a reason chain as indented steps.
func renderChain(w io.Writer, chain *explain.ReasonChain) {
	for i, step := range chain.Steps {
		prefix := "  "
		if i > 0 {
			prefix = "    → "
		}
		fmt.Fprintf(w, "%s%s\n", prefix, step.From)
		fmt.Fprintf(w, "    → %s  [%s, confidence: %.0f%%]\n",
			step.To, step.Relationship, step.EdgeConfidence*100)
	}
}

// renderTestSummaryLine renders a one-line test summary.
func renderTestSummaryLine(w io.Writer, te *explain.TestExplanation) {
	conf := fmt.Sprintf("%.0f%%", te.Confidence*100)
	reason := ""
	if te.StrongestPath != nil && len(te.StrongestPath.Steps) > 0 {
		step := te.StrongestPath.Steps[0]
		reason = fmt.Sprintf("via %s from %s", step.Relationship, step.From)
	}
	fmt.Fprintf(w, "  %-40s %5s  %s\n", te.Target.Path, conf, reason)
}

// reasonCategoryLabel converts a reason category key to a human label.
func reasonCategoryLabel(category string) string {
	switch category {
	case "directDependency":
		return "Direct code dependency"
	case "fixtureDependency":
		return "Fixture/helper dependency"
	case "directlyChanged":
		return "Directly changed"
	case "directoryProximity":
		return "Directory proximity"
	default:
		return category
	}
}
