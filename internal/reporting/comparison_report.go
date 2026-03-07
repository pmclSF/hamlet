package reporting

import (
	"fmt"
	"io"
	"strings"

	"github.com/pmclSF/hamlet/internal/comparison"
)

// RenderComparisonReport writes a human-readable comparison report to w.
func RenderComparisonReport(w io.Writer, comp *comparison.SnapshotComparison) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	line("Hamlet Snapshot Comparison")
	line(strings.Repeat("=", 40))
	blank()

	// Compared
	line("Compared")
	line("  from: %s", comp.FromTime)
	line("  to:   %s", comp.ToTime)
	blank()

	if !comp.HasMeaningfulChanges() {
		line("No meaningful changes detected.")
		blank()
		return
	}

	// Test file count
	if comp.TestFileCountDelta != 0 {
		sign := "+"
		if comp.TestFileCountDelta < 0 {
			sign = ""
		}
		line("Test Files: %s%d", sign, comp.TestFileCountDelta)
		blank()
	}

	// Signal changes
	if len(comp.SignalDeltas) > 0 {
		line("Signal Changes")
		line(strings.Repeat("-", 40))
		for _, d := range comp.SignalDeltas {
			sign := "+"
			if d.Delta < 0 {
				sign = ""
			}
			line("  %-26s %s%d", d.Type, sign, d.Delta)
		}
		blank()
	}

	// Risk changes
	changedRisks := 0
	for _, r := range comp.RiskDeltas {
		if r.Changed {
			changedRisks++
		}
	}
	if len(comp.RiskDeltas) > 0 {
		line("Risk Changes")
		line(strings.Repeat("-", 40))
		for _, r := range comp.RiskDeltas {
			if r.Changed {
				before := string(r.Before)
				if before == "" {
					before = "(none)"
				}
				after := string(r.After)
				if after == "" {
					after = "(none)"
				}
				line("  %-20s %s → %s", r.Type+" ("+r.Scope+")", before, after)
			} else {
				line("  %-20s unchanged", r.Type+" ("+r.Scope+")")
			}
		}
		blank()
	}

	// Framework changes
	if len(comp.FrameworkChanges) > 0 {
		line("Framework Changes")
		line(strings.Repeat("-", 40))
		for _, fc := range comp.FrameworkChanges {
			line("  %s %s (%d files)", fc.Change, fc.Name, fc.Files)
		}
		blank()
	}

	// Representative changes
	if len(comp.NewSignalExamples) > 0 {
		line("New Findings")
		line(strings.Repeat("-", 40))
		for _, ex := range comp.NewSignalExamples {
			loc := ex.File
			if loc == "" {
				loc = "(repo-level)"
			}
			line("  [%s] %s", ex.Type, loc)
		}
		blank()
	}

	if len(comp.ResolvedSignalExamples) > 0 {
		line("Resolved")
		line(strings.Repeat("-", 40))
		for _, ex := range comp.ResolvedSignalExamples {
			loc := ex.File
			if loc == "" {
				loc = "(repo-level)"
			}
			line("  [%s] %s", ex.Type, loc)
		}
		blank()
	}
}
