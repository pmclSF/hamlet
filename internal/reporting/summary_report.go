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

	// Posture (measurement layer)
	if snap.Measurements != nil && len(snap.Measurements.Posture) > 0 {
		line("Posture Dimensions")
		line(strings.Repeat("-", 50))
		for _, p := range snap.Measurements.Posture {
			line("  %-24s %s", p.Dimension+":", strings.ToUpper(p.Band))
			if p.Explanation != "" {
				line("    %s", p.Explanation)
			}
		}
		blank()
	}

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

	// Coverage by type
	if snap.CoverageSummary != nil && snap.CoverageSummary.TotalCodeUnits > 0 {
		cs := snap.CoverageSummary
		line("Coverage by Type")
		line(strings.Repeat("-", 50))
		line("  Code units:          %d", cs.TotalCodeUnits)
		if cs.CoveredByUnitTests > 0 {
			line("  Covered by unit:     %d", cs.CoveredByUnitTests)
		}
		if cs.CoveredOnlyByE2E > 0 {
			line("  Covered only by e2e: %d", cs.CoveredOnlyByE2E)
		}
		if cs.UncoveredExported > 0 {
			line("  Uncovered exports:   %d", cs.UncoveredExported)
		}
		if cs.LineCoveragePct > 0 {
			line("  Line coverage:       %.1f%%", cs.LineCoveragePct)
		}
		blank()
	}

	// Test identity summary
	if len(snap.TestCases) > 0 {
		typeCounts := map[string]int{}
		for _, tc := range snap.TestCases {
			if tc.TestType != "" {
				typeCounts[tc.TestType]++
			}
		}
		if len(typeCounts) > 0 {
			line("Test Types")
			line(strings.Repeat("-", 50))
			line("  Total test cases:    %d", len(snap.TestCases))
			for _, t := range []string{"unit", "integration", "e2e"} {
				if c, ok := typeCounts[t]; ok {
					line("  %-20s %d", t+":", c)
				}
			}
			blank()
		}
	}

	// Next command hints
	line("Next steps:")
	line("  hamlet posture       evidence behind each dimension")
	line("  hamlet analyze       full signal-level detail")
	line("  hamlet compare       see what changed since last snapshot")
	blank()
}
