package reporting

import (
	"fmt"
	"io"
	"strings"

	"github.com/pmclSF/terrain/internal/measurement"
	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/signals"
	"github.com/pmclSF/terrain/internal/summary"
	"github.com/pmclSF/terrain/internal/uitokens"
)

// RenderExecutiveSummary writes a concise, leadership-oriented summary to w.
//
// This report is designed to be paste-ready for leadership updates,
// technical debt reviews, and migration planning discussions.
func RenderExecutiveSummary(w io.Writer, es *summary.ExecutiveSummary) {
	line, blank := reportHelpers(w)

	line(uitokens.Header("Executive Summary"))
	blank()

	// Overall posture — surface the underlying measurements alongside
	// the band. 0.2.0 polish: previously this section showed only the
	// band label ("Health: Strong"), which is a categorical
	// compression of the measurements that drove it. The reader had
	// to take the band on faith. Now the line is:
	//
	//   Health: Strong  (0.0% flaky · 3.6% skipped · 0.0% dead · 0.0% slow)
	//
	// — so the reader sees both the verdict (the band, polarity-
	// translated) and the concrete numbers. `terrain posture` retains
	// the full measurement breakdown with evidence + caveats; this
	// summary view trims to a one-line digest.
	line("Overall Posture")
	line(uitokens.H2Sep)
	unmeasured := 0
	for _, d := range es.Posture.Dimensions {
		if measurement.PostureBand(d.Band) == measurement.PostureUnknown {
			unmeasured++
		}
	}
	switch {
	case len(es.Posture.Dimensions) == 0:
		line("  (no risk surfaces computed)")
	case unmeasured == len(es.Posture.Dimensions):
		// Every dimension is unmeasured — say it once instead of
		// repeating "not yet measured" per row.
		line("  %s", unmeasuredLine(es.Posture.Dimensions))
	default:
		for _, d := range es.Posture.Dimensions {
			dim := measurement.Dimension(d.Dimension)
			label := measurement.DimensionDisplayName(dim)
			if measurement.PostureBand(d.Band) == measurement.PostureUnknown {
				line("  %-22s %s", label+":", unmeasuredCopy(d.NeedsInput))
				continue
			}
			band := measurement.BandDisplayForDimension(dim, measurement.PostureBand(d.Band))
			if len(d.KeyMeasurements) == 0 {
				line("  %-22s %s", label+":", band)
				continue
			}
			// Compact "value label" pairs joined by middle dot.
			parts := make([]string, 0, len(d.KeyMeasurements))
			for _, m := range d.KeyMeasurements {
				parts = append(parts, fmt.Sprintf("%s %s", m.FormattedValue, m.ShortLabel))
			}
			line("  %-22s %s  (%s)", label+":", band, strings.Join(parts, " · "))
		}
		line("  Dimension meaning:")
		line("    Health:             test-system reliability and signal cleanliness")
		line("    Coverage depth:     how thoroughly tests cover the source surface")
		line("    Coverage diversity: how many test layers (unit / integration / e2e) participate")
		line("    Structural risk:    blind-spots in the dependency graph (high fanout, weak coverage)")
		line("    Operational risk:   skip / flake / slow-test burden eroding signal")
	}
	blank()

	// Key numbers
	line("Key Numbers")
	line(uitokens.H2Sep)
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
		line(uitokens.H2Sep)
		for _, a := range es.TopRiskAreas {
			// "Top Risk Areas" is unambiguously risk-shaped output —
			// translate Strong → Low, Weak → Significant, etc. so
			// "low migration risk" / "critical quality risk" both
			// read naturally. Use a synthetic risk-polarity dim to
			// reuse the helper.
			band := measurement.BandDisplayForDimension(
				measurement.DimensionStructuralRisk,
				measurement.PostureBand(a.Band),
			)
			line("  %-25s %s %s risk", a.Name, band, a.RiskType)
		}
		blank()
	}

	// Trend highlights
	if es.HasTrendData && len(es.TrendHighlights) > 0 {
		line("Trend Highlights")
		line(uitokens.H2Sep)
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
		line(uitokens.H2Sep)
		line("  This is the first analysis — it establishes your baseline.")
		line("  Save it and re-run later to see trends:")
		line("    terrain analyze --write-snapshot    save this as baseline")
		line("  On subsequent runs, Terrain will show changes in risk, signals, and posture.")
		blank()
	}

	// Dominant drivers. The summary stores raw signal-type ids (stable
	// for JSON consumers); the human title from the signals manifest
	// leads, with the id dimmed after it for `terrain explain <id>`.
	if len(es.DominantDrivers) > 0 {
		line("Dominant Drivers")
		line(uitokens.H2Sep)
		for _, d := range es.DominantDrivers {
			title := signals.TitleForType(models.SignalType(d))
			if title == d {
				line("  %s", d)
				continue
			}
			line("  %s  %s", title, uitokens.Muted(d))
		}
		blank()
	}

	// Recommended focus
	if es.RecommendedFocus != "" {
		line("Recommended Focus")
		line(uitokens.H2Sep)
		line("  %s", es.RecommendedFocus)
		blank()
	}

	// Structured recommendations
	if len(es.Recommendations) > 0 {
		line("Prioritized Recommendations")
		line(uitokens.H2Sep)
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
		line(uitokens.H2Sep)
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
	line(uitokens.H2Sep)
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
	line("  terrain posture       evidence behind each dimension")
	line("  terrain analyze       full signal-level detail")
	line("  terrain export benchmark   privacy-safe export")
	blank()
}

// unmeasuredCopy renders the row-level text for an unmeasured
// dimension: "not yet measured — needs runtime data", or the plain
// form when the missing input could not be attributed.
func unmeasuredCopy(needsInput string) string {
	if needsInput == "" {
		return "not yet measured"
	}
	return fmt.Sprintf("not yet measured — needs %s", needsInput)
}

// unmeasuredLine renders the single collapsed line used when every
// posture dimension is unmeasured, naming each distinct missing input
// once in row order.
func unmeasuredLine(dims []summary.DimensionPosture) string {
	var needs []string
	seen := map[string]bool{}
	for _, d := range dims {
		if d.NeedsInput == "" || seen[d.NeedsInput] {
			continue
		}
		seen[d.NeedsInput] = true
		needs = append(needs, d.NeedsInput)
	}
	if len(needs) == 0 {
		return "Not yet measured."
	}
	return fmt.Sprintf("Not yet measured — needs %s.", strings.Join(needs, " and "))
}
