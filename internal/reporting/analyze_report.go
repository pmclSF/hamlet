// Package reporting renders TestSuiteSnapshot data into user-facing outputs.
package reporting

import (
	"fmt"
	"io"
	"strings"

	"github.com/pmclSF/hamlet/internal/models"
)

// RenderAnalyzeReport writes a human-readable analysis report to w.
func RenderAnalyzeReport(w io.Writer, snap *models.TestSuiteSnapshot) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	// Header
	line("Hamlet — Test Suite Analysis")
	line(strings.Repeat("=", 40))
	blank()

	// Repository
	line("Repository:  %s", snap.Repository.Name)
	line("Root:        %s", snap.Repository.RootPath)
	if snap.Repository.Branch != "" {
		line("Branch:      %s", snap.Repository.Branch)
	}
	if snap.Repository.CommitSHA != "" {
		sha := snap.Repository.CommitSHA
		if len(sha) > 10 {
			sha = sha[:10]
		}
		line("Commit:      %s", sha)
	}
	if len(snap.Repository.Languages) > 0 {
		line("Languages:   %s", strings.Join(snap.Repository.Languages, ", "))
	}
	if len(snap.Repository.PackageManagers) > 0 {
		line("Packages:    %s", strings.Join(snap.Repository.PackageManagers, ", "))
	}
	if len(snap.Repository.CISystems) > 0 {
		line("CI:          %s", strings.Join(snap.Repository.CISystems, ", "))
	}
	blank()

	// Frameworks
	line("Frameworks")
	line(strings.Repeat("-", 40))
	if len(snap.Frameworks) == 0 {
		line("  (no test frameworks detected)")
	} else {
		for _, fw := range snap.Frameworks {
			typeBadge := ""
			if fw.Type != "" && fw.Type != models.FrameworkTypeUnknown {
				typeBadge = fmt.Sprintf(" [%s]", fw.Type)
			}
			line("  %-20s %4d files%s", fw.Name, fw.FileCount, typeBadge)
		}
	}
	blank()

	// Test files summary
	line("Test Files")
	line(strings.Repeat("-", 40))
	line("  Discovered:  %d", len(snap.TestFiles))
	if len(snap.TestFiles) > 0 {
		// Show a few representative examples
		limit := 5
		if len(snap.TestFiles) < limit {
			limit = len(snap.TestFiles)
		}
		for _, tf := range snap.TestFiles[:limit] {
			fw := tf.Framework
			if fw == "" {
				fw = "?"
			}
			line("    %s  (%s)", tf.Path, fw)
		}
		if len(snap.TestFiles) > 5 {
			line("    ... and %d more", len(snap.TestFiles)-5)
		}
	}
	blank()

	// Signals
	line("Signals")
	line(strings.Repeat("-", 40))
	if len(snap.Signals) == 0 {
		line("  No signals detected.")
	} else {
		counts := map[models.SignalCategory]int{}
		byType := map[models.SignalType]int{}
		for _, s := range snap.Signals {
			counts[s.Category]++
			byType[s.Type]++
		}
		for _, cat := range []models.SignalCategory{
			models.CategoryHealth,
			models.CategoryQuality,
			models.CategoryMigration,
			models.CategoryGovernance,
		} {
			if c := counts[cat]; c > 0 {
				line("  %-14s %d", cat, c)
			}
		}
		blank()

		// Show breakdown by signal type
		line("  Breakdown:")
		for _, st := range []models.SignalType{
			"weakAssertion", "mockHeavyTest", "untestedExport",
			"coverageThresholdBreak", "flakyTest", "slowTest",
			"skippedTest", "deadTest",
		} {
			if c := byType[st]; c > 0 {
				line("    %-26s %d", st, c)
			}
		}

		// Show a few representative findings
		blank()
		line("  Top findings:")
		limit := 5
		if len(snap.Signals) < limit {
			limit = len(snap.Signals)
		}
		for _, s := range snap.Signals[:limit] {
			loc := s.Location.File
			if loc == "" {
				loc = s.Location.Repository
			}
			line("    [%s] %s", s.Severity, s.Explanation)
			if loc != "" {
				line("           %s", loc)
			}
		}
		if len(snap.Signals) > 5 {
			line("    ... and %d more signals", len(snap.Signals)-5)
		}
	}
	blank()

	// Posture (measurement layer)
	if snap.Measurements != nil && len(snap.Measurements.Posture) > 0 {
		line("Posture")
		line(strings.Repeat("-", 40))
		for _, p := range snap.Measurements.Posture {
			line("  %-24s %s", p.Dimension+":", strings.ToUpper(p.Band))
		}
		blank()
	}

	// Risk
	line("Risk")
	line(strings.Repeat("-", 40))
	if len(snap.Risk) == 0 {
		line("  No risk surfaces detected.")
	} else {
		for _, r := range snap.Risk {
			line("  [%s] %s — %s: %s", r.Band, r.Type, r.Scope, r.ScopeName)
		}
	}
	blank()

	// Portfolio intelligence summary
	RenderPortfolioSection(w, snap.Portfolio)

	// Review sections (owner grouping, directory grouping, migration blockers)
	RenderReviewSections(w, snap)

	// Footer
	line("Generated:   %s", snap.GeneratedAt.Format("2006-01-02 15:04:05 UTC"))
	blank()

	// Next command hints
	line("Next steps:")
	line("  hamlet summary       leadership-ready overview")
	line("  hamlet posture       detailed posture with evidence")
	line("  hamlet portfolio     cost, leverage, and redundancy insights")
	line("  hamlet analyze --write-snapshot   save for trend tracking")
	blank()
}
