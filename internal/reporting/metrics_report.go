package reporting

import (
	"fmt"
	"io"
	"strings"

	"github.com/pmclSF/terrain/internal/metrics"
)

// RenderMetricsReport writes a human-readable metrics scorecard to w.
func RenderMetricsReport(w io.Writer, ms *metrics.Snapshot) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	line("Terrain Metrics")
	line(strings.Repeat("=", 40))
	blank()

	// Structure
	line("Structure")
	line(strings.Repeat("-", 40))
	line("  Test files:     %d", ms.Structure.TotalTestFiles)
	if len(ms.Structure.Frameworks) > 0 {
		line("  Frameworks:     %s", strings.Join(ms.Structure.Frameworks, ", "))
	}
	if ms.Structure.FrameworkCount > 1 {
		line("  Fragmentation:  %.2f", ms.Structure.FrameworkFragmentationRatio)
	}
	if len(ms.Structure.Languages) > 0 {
		line("  Languages:      %s", strings.Join(ms.Structure.Languages, ", "))
	}
	blank()

	// Health
	line("Health")
	line(strings.Repeat("-", 40))
	line("  Slow tests:     %d (%.1f%%)", ms.Health.SlowTestCount, ms.Health.SlowTestRatio*100)
	line("  Flaky tests:    %d (%.1f%%)", ms.Health.FlakyTestCount, ms.Health.FlakyTestRatio*100)
	line("  Skipped tests:  %d (%.1f%%)", ms.Health.SkippedTestCount, ms.Health.SkippedTestRatio*100)
	if ms.Health.DeadTestCount > 0 {
		line("  Dead tests:     %d", ms.Health.DeadTestCount)
	}
	blank()

	// Quality
	line("Quality")
	line(strings.Repeat("-", 40))
	line("  Weak assertions:  %d (%.1f%%)", ms.Quality.WeakAssertionCount, ms.Quality.WeakAssertionRatio*100)
	line("  Mock-heavy tests: %d (%.1f%%)", ms.Quality.MockHeavyTestCount, ms.Quality.MockHeavyTestRatio*100)
	if ms.Quality.UntestedExportCount > 0 {
		line("  Untested exports: %d", ms.Quality.UntestedExportCount)
	}
	if ms.Quality.CoverageThresholdBreakCount > 0 {
		line("  Coverage breaks:  %d", ms.Quality.CoverageThresholdBreakCount)
	}
	blank()

	// Change readiness
	totalBlockers := ms.Change.MigrationBlockerCount +
		ms.Change.DeprecatedPatternCount +
		ms.Change.DynamicGenerationCount +
		ms.Change.CustomMatcherRiskCount
	if totalBlockers > 0 {
		line("Change Readiness")
		line(strings.Repeat("-", 40))
		if ms.Change.MigrationBlockerCount > 0 {
			line("  Migration blockers:    %d", ms.Change.MigrationBlockerCount)
		}
		if ms.Change.DeprecatedPatternCount > 0 {
			line("  Deprecated patterns:   %d", ms.Change.DeprecatedPatternCount)
		}
		if ms.Change.DynamicGenerationCount > 0 {
			line("  Dynamic generation:    %d", ms.Change.DynamicGenerationCount)
		}
		if ms.Change.CustomMatcherRiskCount > 0 {
			line("  Custom matcher risk:   %d", ms.Change.CustomMatcherRiskCount)
		}
		blank()
	}

	// Governance
	totalGov := ms.Governance.PolicyViolationCount +
		ms.Governance.LegacyFrameworkUsageCount +
		ms.Governance.RuntimeBudgetExceededCount
	if totalGov > 0 {
		line("Governance")
		line(strings.Repeat("-", 40))
		line("  Policy violations:     %d", ms.Governance.PolicyViolationCount)
		if ms.Governance.LegacyFrameworkUsageCount > 0 {
			line("  Legacy framework:      %d", ms.Governance.LegacyFrameworkUsageCount)
		}
		if ms.Governance.RuntimeBudgetExceededCount > 0 {
			line("  Runtime exceeded:      %d", ms.Governance.RuntimeBudgetExceededCount)
		}
		blank()
	}

	// Risk
	line("Risk")
	line(strings.Repeat("-", 40))
	if ms.Risk.ReliabilityBand != "" {
		line("  Reliability:    %s", ms.Risk.ReliabilityBand)
	}
	if ms.Risk.ChangeBand != "" {
		line("  Change:         %s", ms.Risk.ChangeBand)
	}
	if ms.Risk.SpeedBand != "" {
		line("  Speed:          %s", ms.Risk.SpeedBand)
	}
	if ms.Risk.ReliabilityBand == "" && ms.Risk.ChangeBand == "" && ms.Risk.SpeedBand == "" {
		line("  (no risk surfaces computed)")
	}
	blank()

	// Notes
	if len(ms.Notes) > 0 {
		line("Notes")
		line(strings.Repeat("-", 40))
		for _, note := range ms.Notes {
			line("  %s", note)
		}
		blank()
	}
}
