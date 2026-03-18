package changescope

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// RenderPRSummaryMarkdown writes a PR-ready markdown summary optimized for
// human review and merge decisions.
//
// Structure:
//   - Header with posture and merge recommendation
//   - Compact metrics table
//   - New risks introduced by this PR (max 10)
//   - Pre-existing gaps touched by this change
//   - Test recommendations (grouped if large)
//   - Execution summary
func RenderPRSummaryMarkdown(w io.Writer, pr *PRAnalysis) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}

	// Deduplicate findings at render time (safeguard).
	findings := DeduplicateFindings(pr.NewFindings)
	directRisk, indirectRisk, existingDebt := ClassifyFindingsDetailed(findings)

	// Merge recommendation.
	mergeRec, mergeExpl := MergeRecommendation(pr.PostureBand, findings)

	// --- Header ---
	badge := postureBadge(pr.PostureBand)
	line("## %s Terrain — %s", badge, mergeRec)
	line("")
	line("*%s*", mergeExpl)
	line("")

	// --- Compact metrics ---
	line("| Metric | Value |")
	line("|--------|-------|")
	line("| Changed files | %d (%d source, %d test) |", pr.ChangedFileCount, pr.ChangedSourceCount, pr.ChangedTestCount)
	line("| Impacted units | %d |", pr.ImpactedUnitCount)
	line("| Protection gaps | %d |", pr.ProtectionGapCount)
	if len(pr.RecommendedTests) > 0 {
		if pr.TotalTestCount > 0 {
			pct := 100 * len(pr.RecommendedTests) / pr.TotalTestCount
			line("| Tests to run | %d of %d (%d%% of suite) |", len(pr.RecommendedTests), pr.TotalTestCount, pct)
		} else {
			line("| Tests to run | %d |", len(pr.RecommendedTests))
		}
	}
	line("")

	// --- New risks introduced by this PR (directly changed files) ---
	if len(directRisk) > 0 {
		line("### New Risks (directly changed)")
		line("")
		renderFindingsLimited(line, directRisk, 10)
		line("")
	}

	// --- Indirectly impacted gaps ---
	if len(indirectRisk) > 0 {
		line("<details><summary>Indirectly impacted protection gaps (%d)</summary>", len(indirectRisk))
		line("")
		renderFindingsLimited(line, indirectRisk, 5)
		line("")
		line("</details>")
		line("")
	}

	// --- Pre-existing gaps touched by this change ---
	if len(existingDebt) > 0 {
		line("<details><summary>Pre-existing issues on changed files (%d)</summary>", len(existingDebt))
		line("")
		limit := 5
		if len(existingDebt) < limit {
			limit = len(existingDebt)
		}
		for _, f := range existingDebt[:limit] {
			line("- `%s`: %s", f.Path, f.Explanation)
		}
		if len(existingDebt) > limit {
			line("- ... and %d more", len(existingDebt)-limit)
		}
		line("")
		line("</details>")
		line("")
	}

	// --- Test recommendations ---
	renderTestRecommendations(line, pr)

	// --- AI Validation ---
	renderAISection(line, pr)

	// --- Execution summary ---
	if len(pr.AffectedOwners) > 0 {
		line("**Owners:** %s", strings.Join(pr.AffectedOwners, ", "))
		line("")
	}

	// --- Limitations ---
	if len(pr.Limitations) > 0 {
		line("<details><summary>Limitations</summary>")
		line("")
		for _, l := range pr.Limitations {
			line("- %s", l)
		}
		line("")
		line("</details>")
		line("")
	}

	line("---")
	line("*[Terrain](https://github.com/pmclSF/terrain) — `terrain pr --json` for full machine-readable results*")
}

// renderFindingsLimited renders up to limit findings, then summarizes overflow.
func renderFindingsLimited(line func(string, ...any), findings []ChangeScopedFinding, limit int) {
	if len(findings) < limit {
		limit = len(findings)
	}
	for _, f := range findings[:limit] {
		icon := severityIcon(f.Severity)
		line("- %s `%s`: %s", icon, f.Path, f.Explanation)
	}
	if len(findings) > limit {
		overflow := SummarizeFindingsBySeverity(findings[limit:])
		parts := formatSeverityCounts(overflow)
		line("- ... and %d more (%s)", len(findings)-limit, strings.Join(parts, ", "))
	}
}

// renderTestRecommendations renders the test recommendations section.
func renderTestRecommendations(line func(string, ...any), pr *PRAnalysis) {
	if len(pr.TestSelections) > 0 {
		line("### Recommended Tests")
		line("")
		if pr.SelectionExplanation != "" {
			line("*%s*", pr.SelectionExplanation)
			line("")
		}

		if len(pr.TestSelections) <= 15 {
			reasons := formatTestReasons(pr.TestSelections)
			line("| Test | Confidence | Why |")
			line("|------|------------|-----|")
			for _, t := range pr.TestSelections {
				line("| `%s` | %s | %s |", t.Path, t.Confidence, reasons[t.Path])
			}
		} else {
			paths := make([]string, len(pr.TestSelections))
			for i, t := range pr.TestSelections {
				paths[i] = t.Path
			}
			groups := GroupTestsByPackage(paths)
			line("| Package | Tests | Sample |")
			line("|---------|-------|--------|")
			for _, g := range groups {
				sample := g.Files[0]
				if len(g.Files) > 1 {
					sample += " ..."
				}
				line("| `%s` | %d | `%s` |", g.Package, g.Count, sample)
			}
		}
		line("")
	} else if len(pr.RecommendedTests) > 0 {
		line("### Recommended Tests")
		line("")
		if len(pr.RecommendedTests) <= 15 {
			for _, t := range pr.RecommendedTests {
				line("- `%s`", t)
			}
		} else {
			groups := GroupTestsByPackage(pr.RecommendedTests)
			for _, g := range groups {
				line("- `%s/` — %d test(s)", g.Package, g.Count)
			}
		}
		line("")
	}
}

// RenderPRCommentConcise writes a concise one-line PR comment.
func RenderPRCommentConcise(w io.Writer, pr *PRAnalysis) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}

	findings := DeduplicateFindings(pr.NewFindings)
	mergeRec, _ := MergeRecommendation(pr.PostureBand, findings)
	badge := postureBadge(pr.PostureBand)
	line("%s **Terrain:** %s — %s", badge, mergeRec, pr.Summary)

	highCount := 0
	for _, f := range findings {
		if f.Severity == "high" {
			highCount++
		}
	}
	if highCount > 0 {
		line("  - %d high-severity finding(s) require attention", highCount)
	}

	if len(pr.TestSelections) > 0 {
		if len(pr.TestSelections) <= 5 {
			paths := make([]string, len(pr.TestSelections))
			for i, t := range pr.TestSelections {
				paths[i] = t.Path
			}
			line("  - Run %d test(s): %s", len(paths), strings.Join(paths, ", "))
		} else {
			line("  - Run %d test(s) (see full comment for details)", len(pr.TestSelections))
		}
	}
}

// RenderCIAnnotation writes CI-annotation-style output.
func RenderCIAnnotation(w io.Writer, pr *PRAnalysis) {
	findings := DeduplicateFindings(pr.NewFindings)
	for _, f := range findings {
		level := "notice"
		switch f.Severity {
		case "high":
			level = "error"
		case "medium":
			level = "warning"
		}
		fmt.Fprintf(w, "::%s file=%s::%s\n", level, f.Path, f.Explanation)
	}
}

// RenderChangeScopedReport writes a human-readable change-scoped report.
func RenderChangeScopedReport(w io.Writer, pr *PRAnalysis) {
	line := func(format string, args ...any) {
		fmt.Fprintf(w, format+"\n", args...)
	}
	blank := func() { fmt.Fprintln(w) }

	findings := DeduplicateFindings(pr.NewFindings)
	mergeRec, mergeExpl := MergeRecommendation(pr.PostureBand, findings)

	line("Terrain — Change-Scoped Analysis")
	line(strings.Repeat("=", 40))
	blank()

	line("Recommendation:  %s", mergeRec)
	line("Posture:         %s", strings.ToUpper(pr.PostureBand))
	line("Files:           %d changed (%d source, %d test)", pr.ChangedFileCount, pr.ChangedSourceCount, pr.ChangedTestCount)
	line("Units:           %d impacted", pr.ImpactedUnitCount)
	line("Gaps:            %d", pr.ProtectionGapCount)
	if len(pr.RecommendedTests) > 0 && pr.TotalTestCount > 0 {
		line("Tests:           %d of %d (%.0f%% reduction)", len(pr.RecommendedTests), pr.TotalTestCount,
			100.0-100.0*float64(len(pr.RecommendedTests))/float64(pr.TotalTestCount))
	}
	line("Reason:          %s", mergeExpl)
	blank()

	directRisk, indirectRisk, existingDebt := ClassifyFindingsDetailed(findings)

	if len(directRisk) > 0 {
		line("New Risks (directly changed)")
		line(strings.Repeat("-", 40))
		for _, f := range directRisk {
			line("  [%s] %s — %s", strings.ToUpper(f.Severity), f.Path, f.Explanation)
		}
		blank()
	}

	if len(indirectRisk) > 0 {
		line("Indirectly Impacted Gaps (%d)", len(indirectRisk))
		line(strings.Repeat("-", 40))
		for _, f := range indirectRisk {
			line("  [%s] %s — %s", strings.ToUpper(f.Severity), f.Path, f.Explanation)
		}
		blank()
	}

	if len(existingDebt) > 0 {
		line("Pre-Existing Issues")
		line(strings.Repeat("-", 40))
		for _, f := range existingDebt {
			line("  [%s] %s — %s", strings.ToUpper(f.Severity), f.Path, f.Explanation)
		}
		blank()
	}

	if len(pr.TestSelections) > 0 {
		line("Recommended Tests (%d)", len(pr.TestSelections))
		line(strings.Repeat("-", 40))
		if pr.SelectionExplanation != "" {
			line("  Strategy: %s", pr.SelectionExplanation)
			blank()
		}
		if len(pr.TestSelections) <= 20 {
			reasons := formatTestReasons(pr.TestSelections)
			for _, t := range pr.TestSelections {
				line("  [%s] %s", t.Confidence, t.Path)
				line("         %s", reasons[t.Path])
			}
		} else {
			paths := make([]string, len(pr.TestSelections))
			for i, t := range pr.TestSelections {
				paths[i] = t.Path
			}
			for _, g := range GroupTestsByPackage(paths) {
				line("  %s/ — %d test(s)", g.Package, g.Count)
			}
		}
		blank()
	}

	if len(pr.AffectedOwners) > 0 {
		line("Affected Owners: %s", strings.Join(pr.AffectedOwners, ", "))
		blank()
	}

	if len(pr.Limitations) > 0 {
		line("Limitations")
		line(strings.Repeat("-", 40))
		for _, l := range pr.Limitations {
			line("  %s", l)
		}
		blank()
	}
}

// renderAISection renders the AI validation summary in markdown.
func renderAISection(line func(string, ...any), pr *PRAnalysis) {
	ai := pr.AI
	if ai == nil {
		return
	}

	line("### AI Validation")
	line("")

	// Impacted capabilities.
	if len(ai.ImpactedCapabilities) > 0 {
		line("**Impacted capabilities:** %s", strings.Join(ai.ImpactedCapabilities, ", "))
		line("")
	}

	// Scenario selection summary.
	line("Scenarios: %d of %d selected", ai.SelectedScenarios, ai.TotalScenarios)
	line("")

	// Blocking signals.
	if len(ai.BlockingSignals) > 0 {
		line("**Blocking signals (%d):**", len(ai.BlockingSignals))
		line("")
		for _, s := range ai.BlockingSignals {
			line("- [%s] **%s**: %s", strings.ToUpper(s.Severity), s.Type, s.Explanation)
		}
		line("")
	}

	// Warning signals.
	if len(ai.WarningSignals) > 0 {
		line("<details><summary>Warning signals (%d)</summary>", len(ai.WarningSignals))
		line("")
		for _, s := range ai.WarningSignals {
			line("- [%s] %s: %s", s.Severity, s.Type, s.Explanation)
		}
		line("")
		line("</details>")
		line("")
	}

	// Impacted scenarios grouped by capability.
	if len(ai.Scenarios) > 0 {
		// Group by capability.
		byCap := map[string][]AIScenarioSummary{}
		var noCap []AIScenarioSummary
		for _, sc := range ai.Scenarios {
			if sc.Capability != "" {
				byCap[sc.Capability] = append(byCap[sc.Capability], sc)
			} else {
				noCap = append(noCap, sc)
			}
		}

		if len(byCap) > 0 || len(noCap) > 0 {
			line("<details><summary>Impacted scenarios (%d)</summary>", len(ai.Scenarios))
			line("")
			// Sort capability keys.
			caps := make([]string, 0, len(byCap))
			for c := range byCap {
				caps = append(caps, c)
			}
			sort.Strings(caps)
			for _, cap := range caps {
				line("**%s:**", cap)
				for _, sc := range byCap[cap] {
					line("- %s — %s", sc.Name, sc.Reason)
				}
			}
			if len(noCap) > 0 {
				for _, sc := range noCap {
					line("- %s — %s", sc.Name, sc.Reason)
				}
			}
			line("")
			line("</details>")
			line("")
		}
	}

	// Uncovered contexts.
	if len(ai.UncoveredContexts) > 0 {
		line("**Changed AI contexts without evaluation (%d):**", len(ai.UncoveredContexts))
		line("")
		for _, c := range ai.UncoveredContexts {
			line("- `%s`", c)
		}
		line("")
	}
}

func postureBadge(band string) string {
	switch band {
	case "well_protected":
		return "[PASS]"
	case "partially_protected":
		return "[WARN]"
	case "weakly_protected":
		return "[RISK]"
	case "high_risk":
		return "[FAIL]"
	case "evidence_limited":
		return "[INFO]"
	default:
		return "[????]"
	}
}

func severityIcon(severity string) string {
	switch severity {
	case "high":
		return "[HIGH]"
	case "medium":
		return "[MED]"
	case "low":
		return "[LOW]"
	default:
		return "[---]"
	}
}

func formatSeverityCounts(counts map[string]int) []string {
	var parts []string
	for _, sev := range []string{"high", "medium", "low"} {
		if c, ok := counts[sev]; ok && c > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", c, sev))
		}
	}
	return parts
}

func deduplicateReasons(reasons []string) string {
	seen := map[string]int{}
	var unique []string
	for _, r := range reasons {
		seen[r]++
		if seen[r] == 1 {
			unique = append(unique, r)
		}
	}
	var parts []string
	for _, r := range unique {
		if seen[r] > 1 {
			parts = append(parts, fmt.Sprintf("%s (%dx)", r, seen[r]))
		} else {
			parts = append(parts, r)
		}
	}
	return strings.Join(parts, "; ")
}

func formatTestReasons(selections []TestSelection) map[string]string {
	reasons := map[string]string{}
	unitTestCount := map[string]int{}
	for _, t := range selections {
		for _, uid := range t.CoversUnits {
			unitTestCount[uid]++
		}
	}
	totalTests := len(selections)
	for _, t := range selections {
		reasons[t.Path] = formatSingleTestReason(t, unitTestCount, totalTests)
	}
	return reasons
}

func formatSingleTestReason(t TestSelection, unitTestCount map[string]int, totalTests int) string {
	unitNames := extractUnitNames(t.CoversUnits)
	if len(unitNames) == 0 {
		if len(t.Reasons) > 0 {
			return deduplicateReasons(t.Reasons)
		}
		return t.Relevance
	}
	var unique, shared []string
	for _, uid := range t.CoversUnits {
		name := uid
		if idx := strings.LastIndex(uid, ":"); idx >= 0 {
			name = uid[idx+1:]
		}
		if unitTestCount[uid] == 1 {
			unique = append(unique, name)
		} else {
			shared = append(shared, name)
		}
	}
	unique = deduplicateStrings(unique)
	shared = deduplicateStrings(shared)
	label := "covers"
	if t.Confidence == "exact" {
		label = "exact coverage of"
	}
	const maxDisplay = 4
	if len(unique) > 0 {
		var why string
		if len(unique) <= maxDisplay {
			why = fmt.Sprintf("%s `%s`", label, strings.Join(unique, "`, `"))
		} else {
			why = fmt.Sprintf("%s `%s` + %d more", label, strings.Join(unique[:maxDisplay], "`, `"), len(unique)-maxDisplay)
		}
		if len(shared) > 0 {
			why += fmt.Sprintf(" (+ %d shared)", len(shared))
		}
		return why
	}
	if len(unitNames) <= maxDisplay {
		why := fmt.Sprintf("%s `%s`", label, strings.Join(unitNames, "`, `"))
		if totalTests > 1 {
			why += fmt.Sprintf(" (shared across %d tests)", unitTestCount[t.CoversUnits[0]])
		}
		return why
	}
	shown := strings.Join(unitNames[:maxDisplay], "`, `")
	why := fmt.Sprintf("%s `%s` + %d more", label, shown, len(unitNames)-maxDisplay)
	if totalTests > 1 {
		why += fmt.Sprintf(" (shared across %d tests)", unitTestCount[t.CoversUnits[0]])
	}
	return why
}

func deduplicateStrings(ss []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func extractUnitNames(unitIDs []string) []string {
	if len(unitIDs) == 0 {
		return nil
	}
	seen := map[string]bool{}
	var names []string
	for _, id := range unitIDs {
		name := id
		if idx := strings.LastIndex(id, ":"); idx >= 0 {
			name = id[idx+1:]
		}
		if name != "" && !seen[name] {
			seen[name] = true
			names = append(names, name)
		}
	}
	return names
}
