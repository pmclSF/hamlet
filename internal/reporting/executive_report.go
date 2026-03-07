package reporting

import (
	"fmt"
	"io"
	"strings"

	"github.com/pmclSF/hamlet/internal/summary"
)

// RenderExecutiveSummary writes a concise, leadership-oriented summary to w.
//
// This report is designed to be paste-ready for leadership updates,
// technical debt reviews, and migration planning discussions.
func RenderExecutiveSummary(w io.Writer, es *summary.ExecutiveSummary) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	line("Hamlet Executive Summary")
	line(strings.Repeat("=", 50))
	blank()

	// Overall posture
	line("Overall Posture")
	line(strings.Repeat("-", 50))
	for _, d := range es.Posture.Dimensions {
		line("  %-20s %s", d.Dimension+":", strings.ToLower(string(d.Band)))
	}
	if len(es.Posture.Dimensions) == 0 {
		line("  (no risk surfaces computed)")
	}
	blank()

	// Key numbers
	line("Key Numbers")
	line(strings.Repeat("-", 50))
	line("  Test files:          %d", es.KeyNumbers.TestFiles)
	line("  Frameworks:          %d", es.KeyNumbers.Frameworks)
	line("  Total signals:       %d", es.KeyNumbers.TotalSignals)
	if es.KeyNumbers.CriticalFindings > 0 {
		line("  Critical findings:   %d", es.KeyNumbers.CriticalFindings)
	}
	if es.KeyNumbers.HighRiskAreas > 0 {
		line("  High-risk areas:     %d", es.KeyNumbers.HighRiskAreas)
	}
	blank()

	// Top risk areas
	if len(es.TopRiskAreas) > 0 {
		line("Top Risk Areas")
		line(strings.Repeat("-", 50))
		for _, a := range es.TopRiskAreas {
			line("  %-25s %s %s risk", a.Name, strings.ToLower(string(a.Band)), a.RiskType)
		}
		blank()
	}

	// Trend highlights
	if es.HasTrendData && len(es.TrendHighlights) > 0 {
		line("Trend Highlights")
		line(strings.Repeat("-", 50))
		for _, t := range es.TrendHighlights {
			icon := " "
			switch t.Direction {
			case "improved":
				icon = "↓"
			case "worsened":
				icon = "↑"
			}
			line("  %s %s", icon, t.Description)
		}
		blank()
	} else if !es.HasTrendData {
		line("Trend Highlights")
		line(strings.Repeat("-", 50))
		line("  No prior snapshots available.")
		line("  Run `hamlet analyze --write-snapshot` to begin tracking trends.")
		blank()
	}

	// Dominant drivers
	if len(es.DominantDrivers) > 0 {
		line("Dominant Drivers")
		line(strings.Repeat("-", 50))
		for _, d := range es.DominantDrivers {
			line("  %s", d)
		}
		blank()
	}

	// Recommended focus
	if es.RecommendedFocus != "" {
		line("Recommended Focus")
		line(strings.Repeat("-", 50))
		line("  %s", es.RecommendedFocus)
		blank()
	}

	// Structured recommendations
	if len(es.Recommendations) > 0 {
		line("Prioritized Recommendations")
		line(strings.Repeat("-", 50))
		for _, r := range es.Recommendations {
			strength := string(r.EvidenceStrength)
			if strength == "" {
				strength = "unknown"
			}
			line("  %d. %s", r.Priority, r.What)
			line("     Why:      %s", r.Why)
			line("     Where:    %s", r.Where)
			line("     Evidence: %s", strength)
		}
		blank()
	}

	// Blind spots
	if len(es.BlindSpots) > 0 {
		line("Known Blind Spots")
		line(strings.Repeat("-", 50))
		for _, b := range es.BlindSpots {
			line("  %s: %s", b.Area, b.Reason)
			if b.Remediation != "" {
				line("    → %s", b.Remediation)
			}
		}
		blank()
	}

	// Benchmark readiness
	line("Benchmark Readiness")
	line(strings.Repeat("-", 50))
	if len(es.BenchmarkReadiness.ReadyDimensions) > 0 {
		line("  Ready:")
		for _, d := range es.BenchmarkReadiness.ReadyDimensions {
			line("    %s", d)
		}
	}
	if len(es.BenchmarkReadiness.LimitedDimensions) > 0 {
		line("  Limited:")
		for _, l := range es.BenchmarkReadiness.LimitedDimensions {
			line("    %s (%s)", l.Dimension, l.Reason)
		}
	}
	blank()

	// Next command hints
	line("Next steps:")
	line("  hamlet posture       evidence behind each dimension")
	line("  hamlet analyze       full signal-level detail")
	line("  hamlet export benchmark   privacy-safe export")
	blank()
}
