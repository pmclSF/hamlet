// Package reporting renders TestSuiteSnapshot data into user-facing outputs.
package reporting

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/signals"
)

// AnalyzeReportOptions configures analyze report rendering.
type AnalyzeReportOptions struct {
	Verbose bool
}

// RenderAnalyzeReport writes a human-readable analysis report to w.
func RenderAnalyzeReport(w io.Writer, snap *models.TestSuiteSnapshot, opts ...AnalyzeReportOptions) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }
	opt := AnalyzeReportOptions{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	// Header
	line("Terrain — Test Suite Analysis")
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

	// Data completeness
	line("Data Completeness")
	line(strings.Repeat("-", 40))
	sourceAvailable := len(snap.TestFiles) > 0 || len(snap.CodeUnits) > 0
	coverageStatus := dataSourceStatus(snap, "coverage")
	runtimeStatus := dataSourceStatus(snap, "runtime")
	policyStatus := dataSourceStatus(snap, "policy")
	line("  [%-9s] Source code", completenessBadge(sourceAvailable))
	line("  [%-9s] Coverage data", completenessBadge(coverageStatus == models.DataSourceAvailable))
	line("  [%-9s] Runtime data", completenessBadge(runtimeStatus == models.DataSourceAvailable))
	line("  [%-9s] Policy config", completenessBadge(policyStatus == models.DataSourceAvailable))
	if coverageStatus != models.DataSourceAvailable {
		line("  Coverage analysis skipped or partial. Provide coverage with:")
		line("    terrain analyze --root . --coverage path/to/lcov.info")
	}
	if runtimeStatus != models.DataSourceAvailable {
		line("  Runtime-dependent signals unavailable without runtime artifacts:")
		line("    slowTest, flakyTest, skippedTest, deadTest, unstableSuite")
		line("  Provide runtime data with:")
		line("    terrain analyze --root . --runtime path/to/results.xml")
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
	} else {
		line("  No test files detected.")
		line("  Terrain looks for patterns like *_test.go, *.test.js, *.spec.ts, test_*.py.")
	}
	blank()

	// Code unit summary
	line("Code Units")
	line(strings.Repeat("-", 40))
	line("  Extracted:   %d", len(snap.CodeUnits))
	if len(snap.CodeUnits) == 0 {
		line("  No source code functions/classes detected.")
		line("  Check that --root points to your source tree.")
	}
	blank()

	// Signals
	line("Signals")
	line(strings.Repeat("-", 40))
	if len(snap.Signals) == 0 {
		line("  No signals detected.")
		line("  This often means Terrain needs more runtime/coverage data to surface issues.")
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
		types := make([]models.SignalType, 0, len(byType))
		for st := range byType {
			types = append(types, st)
		}
		sort.Slice(types, func(i, j int) bool {
			ci, cj := byType[types[i]], byType[types[j]]
			if ci != cj {
				return ci > cj
			}
			return types[i] < types[j]
		})
		for _, st := range types {
			line("    %-26s %d", st, byType[st])
		}

		blank()
		line("  Top findings:")
		orderedSignals := rankedSignals(snap.Signals)
		limit := len(orderedSignals)
		if !opt.Verbose {
			limit = 5
			if len(orderedSignals) < limit {
				limit = len(orderedSignals)
			}
		}
		for _, s := range orderedSignals[:limit] {
			loc := s.Location.File
			if loc == "" {
				loc = s.Location.Repository
			}
			line("    [%s] %s", s.Severity, s.Explanation)
			if loc != "" {
				line("           %s", loc)
			}
			if info, ok := signals.Info(s.Type); ok {
				line("           %s: %s", s.Type, info.Description)
				if info.Remediation != "" {
					line("           Consider: %s", info.Remediation)
				}
			}
		}
		if !opt.Verbose && len(orderedSignals) > limit {
			remaining := orderedSignals[limit:]
			remainingCritical, remainingHigh := countHighSeverities(remaining)
			line("    ... and %d more signals", len(remaining))
			if remainingCritical > 0 || remainingHigh > 0 {
				line("        hidden severities: %d critical, %d high (use --verbose to view all)", remainingCritical, remainingHigh)
			} else {
				line("        use --verbose to view all findings")
			}
		}
	}
	blank()

	// What this means
	line("What This Means")
	line(strings.Repeat("-", 40))
	line("  Terrain found %d test files with %d signals.", len(snap.TestFiles), len(snap.Signals))
	line("  Test suite status: %s", suiteStatusBand(snap))
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
	line("  1) Add coverage data:")
	line("     terrain analyze --root . --coverage path/to/lcov.info")
	line("  2) Add runtime data:")
	line("     terrain analyze --root . --runtime path/to/test-results.xml")
	line("  3) Show full findings:")
	line("     terrain analyze --root . --verbose")
	line("  4) Save trend baseline:")
	line("     terrain analyze --write-snapshot")
	blank()
}

func dataSourceStatus(snap *models.TestSuiteSnapshot, name string) string {
	if snap == nil {
		return models.DataSourceUnavailable
	}
	for _, ds := range snap.DataSources {
		if ds.Name == name {
			return ds.Status
		}
	}
	return models.DataSourceUnavailable
}

func completenessBadge(ok bool) string {
	if ok {
		return "available"
	}
	return "missing"
}

func suiteStatusBand(snap *models.TestSuiteSnapshot) string {
	if snap == nil {
		return "unknown"
	}
	if len(snap.Signals) == 0 {
		return "good (limited evidence)"
	}
	high := 0
	critical := 0
	for _, s := range snap.Signals {
		if s.Severity == models.SeverityCritical {
			critical++
		}
		if s.Severity == models.SeverityHigh || s.Severity == models.SeverityCritical {
			high++
		}
	}
	switch {
	case critical > 0 || high >= 10:
		return "at risk"
	case high >= 1 || len(snap.Signals) >= 5:
		return "needs attention"
	default:
		return "good"
	}
}

func rankedSignals(signals []models.Signal) []models.Signal {
	if len(signals) == 0 {
		return nil
	}
	out := append([]models.Signal(nil), signals...)
	sort.Slice(out, func(i, j int) bool {
		ai := signalSeverityRank(out[i].Severity)
		aj := signalSeverityRank(out[j].Severity)
		if ai != aj {
			return ai > aj
		}
		if out[i].Confidence != out[j].Confidence {
			return out[i].Confidence > out[j].Confidence
		}
		if out[i].Category != out[j].Category {
			return out[i].Category < out[j].Category
		}
		if out[i].Type != out[j].Type {
			return out[i].Type < out[j].Type
		}
		if out[i].Location.File != out[j].Location.File {
			return out[i].Location.File < out[j].Location.File
		}
		return out[i].Location.Line < out[j].Location.Line
	})
	return out
}

func signalSeverityRank(sev models.SignalSeverity) int {
	switch sev {
	case models.SeverityCritical:
		return 5
	case models.SeverityHigh:
		return 4
	case models.SeverityMedium:
		return 3
	case models.SeverityLow:
		return 2
	case models.SeverityInfo:
		return 1
	default:
		return 0
	}
}

func countHighSeverities(signals []models.Signal) (critical, high int) {
	for _, s := range signals {
		if s.Severity == models.SeverityCritical {
			critical++
		}
		if s.Severity == models.SeverityHigh {
			high++
		}
	}
	return critical, high
}
