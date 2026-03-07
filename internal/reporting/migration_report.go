package reporting

import (
	"fmt"
	"io"
	"strings"

	"github.com/pmclSF/hamlet/internal/migration"
)

// RenderMigrationReport writes a migration readiness report to w.
func RenderMigrationReport(w io.Writer, readiness *migration.ReadinessSummary) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	line("Hamlet Migration Readiness")
	line(strings.Repeat("=", 40))
	blank()

	// Frameworks
	line("Frameworks")
	line(strings.Repeat("-", 40))
	if len(readiness.Frameworks) == 0 {
		line("  (no frameworks detected)")
	} else {
		for _, fw := range readiness.Frameworks {
			typeBadge := ""
			if fw.Type != "" {
				typeBadge = fmt.Sprintf(" [%s]", fw.Type)
			}
			line("  %-20s %4d files%s", fw.Name, fw.FileCount, typeBadge)
		}
	}
	blank()

	// Readiness
	line("Readiness Level: %s", strings.ToUpper(readiness.ReadinessLevel))
	line("  %s", readiness.Explanation)
	blank()

	// Blockers
	line("Migration Blockers: %d", readiness.TotalBlockers)
	line(strings.Repeat("-", 40))
	if readiness.TotalBlockers == 0 {
		line("  (none detected)")
	} else {
		for bt, count := range readiness.BlockersByType {
			line("  %-26s %d", bt, count)
		}
	}
	blank()

	// Representative examples
	if len(readiness.RepresentativeBlockers) > 0 {
		line("Representative Blockers")
		line(strings.Repeat("-", 40))
		for _, ex := range readiness.RepresentativeBlockers {
			line("  [%s] %s", ex.Type, ex.Explanation)
			if ex.File != "" {
				line("    %s", ex.File)
			}
		}
		blank()
	}

	// Quality factors compounding migration risk
	if len(readiness.QualityFactors) > 0 {
		line("Quality Factors Affecting Migration")
		line(strings.Repeat("-", 40))
		for _, qf := range readiness.QualityFactors {
			line("  %s", qf.Explanation)
		}
		blank()
	}

	// Area assessments
	if len(readiness.AreaAssessments) > 0 {
		line("Area Assessments")
		line(strings.Repeat("-", 40))
		for _, area := range readiness.AreaAssessments {
			badge := strings.ToUpper(area.Classification)
			line("  [%s] %s", badge, area.Directory)
			line("    %s", area.Explanation)
		}
		blank()
	}

	// Coverage guidance
	if len(readiness.CoverageGuidance) > 0 {
		line("Where Additional Coverage Reduces Migration Risk")
		line(strings.Repeat("-", 40))
		for _, cg := range readiness.CoverageGuidance {
			line("  [%s] %s", strings.ToUpper(cg.Priority), cg.Directory)
			line("    %s", cg.Reason)
		}
		blank()
	}
}

// RenderMigrationBlockers writes a focused migration blockers report to w.
// This is the output for `hamlet migration blockers`.
func RenderMigrationBlockers(w io.Writer, readiness *migration.ReadinessSummary) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	line("Hamlet Migration Blockers")
	line(strings.Repeat("=", 40))
	blank()

	line("Total Blockers: %d", readiness.TotalBlockers)
	blank()

	if readiness.TotalBlockers == 0 {
		line("No migration blockers detected.")
		blank()
		line("This repository appears ready for framework migration or standardization.")
		return
	}

	// Blockers by type
	line("By Type")
	line(strings.Repeat("-", 40))
	for bt, count := range readiness.BlockersByType {
		line("  %-26s %d", bt, count)
	}
	blank()

	// Representative examples
	if len(readiness.RepresentativeBlockers) > 0 {
		line("Examples")
		line(strings.Repeat("-", 40))
		for _, ex := range readiness.RepresentativeBlockers {
			line("  [%s] %s", ex.Type, ex.Explanation)
			if ex.File != "" {
				line("    %s", ex.File)
			}
		}
		blank()
	}

	// Area assessments — only risky and caution
	var risky []migration.AreaAssessment
	for _, area := range readiness.AreaAssessments {
		if area.Classification != "safe" {
			risky = append(risky, area)
		}
	}
	if len(risky) > 0 {
		line("Highest-Risk Areas")
		line(strings.Repeat("-", 40))
		for _, area := range risky {
			badge := strings.ToUpper(area.Classification)
			line("  [%s] %s", badge, area.Directory)
			line("    %s", area.Explanation)
		}
		blank()
	}

	line("Next steps:")
	line("  hamlet migration readiness    full readiness assessment with quality factors")
	line("  hamlet migration preview      preview migration for a specific file")
	line("  hamlet policy check           check against local policy rules")
}

// RenderMigrationPreview writes a migration preview report for a single file.
func RenderMigrationPreview(w io.Writer, preview *migration.PreviewResult) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	line("Hamlet Migration Preview")
	line(strings.Repeat("=", 40))
	blank()

	line("File: %s", preview.File)
	line("Source Framework: %s", preview.SourceFramework)
	if preview.SuggestedTarget != "" {
		line("Suggested Target: %s", preview.SuggestedTarget)
	}
	line("Difficulty: %s", strings.ToUpper(preview.Difficulty))
	blank()

	if !preview.PreviewAvailable {
		line("Preview Not Available")
		line(strings.Repeat("-", 40))
		line("  %s", preview.Explanation)
		blank()
		if len(preview.Limitations) > 0 {
			for _, l := range preview.Limitations {
				line("  - %s", l)
			}
			blank()
		}
		return
	}

	line("Assessment")
	line(strings.Repeat("-", 40))
	line("  %s", preview.Explanation)
	blank()

	// Blockers
	if len(preview.Blockers) > 0 {
		line("Migration Blockers (%d)", len(preview.Blockers))
		line(strings.Repeat("-", 40))
		for _, b := range preview.Blockers {
			line("  [%s] %s", b.Type, b.Pattern)
			line("    %s", b.Explanation)
			line("    -> %s", b.Remediation)
		}
		blank()
	} else {
		line("Migration Blockers: none")
		blank()
	}

	// Safe patterns
	if len(preview.SafePatterns) > 0 {
		line("Safe Patterns (should migrate cleanly)")
		line(strings.Repeat("-", 40))
		for _, p := range preview.SafePatterns {
			line("  + %s", p)
		}
		blank()
	}

	// Limitations
	if len(preview.Limitations) > 0 {
		line("Limitations")
		line(strings.Repeat("-", 40))
		for _, l := range preview.Limitations {
			line("  - %s", l)
		}
		blank()
	}
}

// RenderMigrationPreviewScope writes a scope-level migration preview summary.
func RenderMigrationPreviewScope(w io.Writer, previews []*migration.PreviewResult) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	line("Hamlet Migration Preview (scope)")
	line(strings.Repeat("=", 40))
	blank()

	if len(previews) == 0 {
		line("No test files found in scope.")
		return
	}

	// Summary counts
	low, medium, high, unknown := 0, 0, 0, 0
	for _, p := range previews {
		switch p.Difficulty {
		case "low":
			low++
		case "medium":
			medium++
		case "high":
			high++
		default:
			unknown++
		}
	}

	line("Files analyzed: %d", len(previews))
	line("  Low difficulty:    %d", low)
	line("  Medium difficulty: %d", medium)
	line("  High difficulty:   %d", high)
	if unknown > 0 {
		line("  Unknown:           %d", unknown)
	}
	blank()

	// Show high-difficulty files first
	if high > 0 {
		line("High-Difficulty Files (need manual review)")
		line(strings.Repeat("-", 40))
		for _, p := range previews {
			if p.Difficulty == "high" {
				line("  %s (%s)", p.File, p.SourceFramework)
				line("    %d blocker(s)", len(p.Blockers))
			}
		}
		blank()
	}

	// Medium
	if medium > 0 {
		line("Medium-Difficulty Files")
		line(strings.Repeat("-", 40))
		for _, p := range previews {
			if p.Difficulty == "medium" {
				line("  %s (%s)", p.File, p.SourceFramework)
				line("    %d blocker(s)", len(p.Blockers))
			}
		}
		blank()
	}

	// Low
	if low > 0 {
		line("Low-Difficulty Files (good candidates to start)")
		line(strings.Repeat("-", 40))
		for _, p := range previews {
			if p.Difficulty == "low" {
				line("  %s (%s)", p.File, p.SourceFramework)
			}
		}
		blank()
	}
}
