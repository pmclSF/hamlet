package reporting

import (
	"fmt"
	"io"
	"strings"

	"github.com/pmclSF/hamlet/internal/heatmap"
	"github.com/pmclSF/hamlet/internal/models"
)

// RenderSummaryReport writes a leadership-oriented summary to w.
func RenderSummaryReport(w io.Writer, snap *models.TestSuiteSnapshot, h *heatmap.Heatmap) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	line("Hamlet Summary")
	line(strings.Repeat("=", 50))
	blank()

	// Posture
	line("Posture: %s", strings.ToUpper(string(h.PostureBand)))
	line("%s", h.PostureSummary)
	blank()

	// Key numbers
	line("Key Numbers")
	line(strings.Repeat("-", 50))
	line("  Test files:          %d", len(snap.TestFiles))
	line("  Frameworks:          %d", len(snap.Frameworks))
	line("  Total signals:       %d", h.TotalSignals)
	if h.CriticalCount > 0 {
		line("  Critical findings:   %d", h.CriticalCount)
	}
	if h.HighRiskAreaCount > 0 {
		line("  High-risk areas:     %d", h.HighRiskAreaCount)
	}
	blank()

	// Risk bands
	hasRisk := false
	for _, r := range snap.Risk {
		if r.Scope == "repository" {
			if !hasRisk {
				line("Risk Dimensions")
				line(strings.Repeat("-", 50))
				hasRisk = true
			}
			line("  %-20s %s", r.Type+":", strings.ToUpper(string(r.Band)))
		}
	}
	if hasRisk {
		blank()
	}

	// Directory hotspots (top 5)
	if len(h.DirectoryHotSpots) > 0 {
		line("Highest-Risk Directories")
		line(strings.Repeat("-", 50))
		limit := 5
		if len(h.DirectoryHotSpots) < limit {
			limit = len(h.DirectoryHotSpots)
		}
		for _, hs := range h.DirectoryHotSpots[:limit] {
			line("  %-30s %s  (%d signals)", hs.Name, strings.ToUpper(string(hs.Band)), hs.SignalCount)
		}
		blank()
	}

	// Owner hotspots (top 5)
	if len(h.OwnerHotSpots) > 0 {
		line("Highest-Risk Owners")
		line(strings.Repeat("-", 50))
		limit := 5
		if len(h.OwnerHotSpots) < limit {
			limit = len(h.OwnerHotSpots)
		}
		for _, hs := range h.OwnerHotSpots[:limit] {
			line("  %-30s %s  (%d signals)", hs.Name, strings.ToUpper(string(hs.Band)), hs.SignalCount)
		}
		blank()
	}
}
