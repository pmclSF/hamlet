package changescope

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pmclSF/terrain/internal/impact"
	"github.com/pmclSF/terrain/internal/models"
)

// AnalyzePR performs a PR/change-scoped analysis.
// It runs impact analysis on the change scope and produces a focused
// PRAnalysis with findings, posture delta, and recommendations.
func AnalyzePR(scope *impact.ChangeScope, snap *models.TestSuiteSnapshot) *PRAnalysis {
	result := impact.Analyze(scope, snap)

	pr := &PRAnalysis{
		SchemaVersion: PRAnalysisSchemaVersion,
		Scope:         *scope,
		ImpactResult:  result,
	}

	// Count changed files by type.
	for _, cf := range scope.ChangedFiles {
		pr.ChangedFileCount++
		if cf.IsTestFile {
			pr.ChangedTestCount++
		} else if impact.IsAnalyzableSourceFile(cf.Path) {
			pr.ChangedSourceCount++
		}
		// Non-analyzable, non-test files (docs, config, CI) are counted
		// in ChangedFileCount but not in source or test counts.
	}

	pr.ImpactedUnitCount = len(result.ImpactedUnits)
	pr.ProtectionGapCount = len(result.ProtectionGaps)
	pr.PostureBand = result.Posture.Band
	pr.AffectedOwners = result.ImpactedOwners
	pr.Limitations = result.Limitations
	pr.TotalTestCount = len(snap.TestFiles)

	// Extract recommended tests with reasoning.
	populateTestSelections(pr, result)

	// Build change-scoped findings from protection gaps and signals (deduplicated).
	pr.NewFindings = DeduplicateFindings(buildChangeScopedFindings(result, snap))

	// Build posture delta.
	pr.PostureDelta = buildPostureDelta(result)

	// Build AI validation summary.
	pr.AI = buildAIValidationSummary(result, snap)

	// Build summary.
	pr.Summary = buildPRSummary(pr)

	return pr
}

// AnalyzePRFromChangeSet performs a PR/change-scoped analysis starting from
// a ChangeSet. This is the preferred entry point for new code.
func AnalyzePRFromChangeSet(cs *models.ChangeSet, snap *models.TestSuiteSnapshot) *PRAnalysis {
	result := impact.AnalyzeChangeSet(cs, snap)

	pr := &PRAnalysis{
		SchemaVersion: PRAnalysisSchemaVersion,
		Scope:         result.Scope,
		ImpactResult:  result,
	}

	for _, cf := range cs.ChangedFiles {
		pr.ChangedFileCount++
		if cf.IsTestFile {
			pr.ChangedTestCount++
		} else if impact.IsAnalyzableSourceFile(cf.Path) {
			pr.ChangedSourceCount++
		}
	}

	pr.ImpactedUnitCount = len(result.ImpactedUnits)
	pr.ProtectionGapCount = len(result.ProtectionGaps)
	pr.PostureBand = result.Posture.Band
	pr.AffectedOwners = result.ImpactedOwners
	pr.Limitations = result.Limitations
	pr.TotalTestCount = len(snap.TestFiles)

	populateTestSelections(pr, result)

	pr.NewFindings = DeduplicateFindings(buildChangeScopedFindings(result, snap))
	pr.PostureDelta = buildPostureDelta(result)
	pr.AI = buildAIValidationSummary(result, snap)
	pr.Summary = buildPRSummary(pr)

	return pr
}

func populateTestSelections(pr *PRAnalysis, result *impact.ImpactResult) {
	if result.ProtectiveSet != nil {
		pr.SelectionStrategy = result.ProtectiveSet.SetKind
		pr.SelectionExplanation = result.ProtectiveSet.Explanation
		for _, st := range result.ProtectiveSet.Tests {
			pr.RecommendedTests = append(pr.RecommendedTests, st.Path)
			ts := TestSelection{
				Path:        st.Path,
				Confidence:  string(st.ImpactConfidence),
				Relevance:   st.Relevance,
				CoversUnits: st.CoversUnits,
			}
			for _, r := range st.Reasons {
				ts.Reasons = append(ts.Reasons, r.Reason)
			}
			pr.TestSelections = append(pr.TestSelections, ts)
		}
	} else {
		for _, t := range result.SelectedTests {
			pr.RecommendedTests = append(pr.RecommendedTests, t.Path)
			pr.TestSelections = append(pr.TestSelections, TestSelection{
				Path:        t.Path,
				Confidence:  string(t.ImpactConfidence),
				Relevance:   t.Relevance,
				CoversUnits: t.CoversUnits,
			})
		}
	}
}

// AnalyzeChangedPaths is a convenience function that creates a change scope
// from explicit paths and runs PR analysis.
func AnalyzeChangedPaths(paths []string, changeKind impact.ChangeKind, snap *models.TestSuiteSnapshot) *PRAnalysis {
	scope := impact.ChangeScopeFromPaths(paths, changeKind)
	return AnalyzePR(scope, snap)
}

func buildChangeScopedFindings(result *impact.ImpactResult, snap *models.TestSuiteSnapshot) []ChangeScopedFinding {
	var findings []ChangeScopedFinding

	// Build set of directly changed file paths for scope classification.
	directPaths := map[string]bool{}
	for _, cf := range result.Scope.ChangedFiles {
		directPaths[cf.Path] = true
	}

	// Convert protection gaps to findings, classifying each as direct or indirect.
	for _, gap := range result.ProtectionGaps {
		scope := "indirect"
		if directPaths[gap.Path] {
			scope = "direct"
		}
		findings = append(findings, ChangeScopedFinding{
			Type:            "protection_gap",
			Scope:           scope,
			Path:            gap.Path,
			Severity:        gap.Severity,
			Explanation:     gap.Explanation,
			SuggestedAction: gap.SuggestedAction,
		})
	}

	// Check for signals on changed files, but skip signals that duplicate
	// protection gaps already surfaced above.
	gapPaths := map[string]bool{}
	for _, gap := range result.ProtectionGaps {
		gapPaths[gap.Path] = true
	}

	for _, sig := range snap.Signals {
		if !directPaths[sig.Location.File] {
			continue
		}
		if sig.Type == "untestedExport" && gapPaths[sig.Location.File] {
			continue
		}
		findings = append(findings, ChangeScopedFinding{
			Type:        "existing_signal",
			Scope:       "direct",
			Path:        sig.Location.File,
			Severity:    string(sig.Severity),
			Explanation: fmt.Sprintf("[%s] %s", sig.Type, sig.Explanation),
		})
	}

	return findings
}

func buildPostureDelta(result *impact.ImpactResult) *PostureDelta {
	delta := &PostureDelta{
		NewGapCount: len(result.ProtectionGaps),
	}

	switch result.Posture.Band {
	case "well_protected":
		delta.OverallDirection = "unchanged"
		delta.Explanation = "Change is well protected by existing tests."
	case "partially_protected":
		delta.OverallDirection = "unchanged"
		delta.Explanation = fmt.Sprintf("Change has partial protection. %d gap(s) found.", len(result.ProtectionGaps))
	case "weakly_protected", "high_risk":
		delta.OverallDirection = "worsened"
		delta.Explanation = fmt.Sprintf("Change introduces risk. %d protection gap(s) found.", len(result.ProtectionGaps))
	default:
		delta.OverallDirection = "unchanged"
		delta.Explanation = "Unable to determine posture change."
	}

	return delta
}

func buildPRSummary(pr *PRAnalysis) string {
	parts := []string{
		fmt.Sprintf("%d file(s) changed", pr.ChangedFileCount),
	}
	if pr.ImpactedUnitCount > 0 {
		parts = append(parts, fmt.Sprintf("%d unit(s) impacted", pr.ImpactedUnitCount))
	}
	if pr.ProtectionGapCount > 0 {
		parts = append(parts, fmt.Sprintf("%d gap(s)", pr.ProtectionGapCount))
	}
	if len(pr.RecommendedTests) > 0 {
		parts = append(parts, fmt.Sprintf("%d test(s) recommended", len(pr.RecommendedTests)))
	}

	return strings.Join(parts, ", ") + ". Posture: " + pr.PostureBand + "."
}

// buildAIValidationSummary extracts AI-specific validation data from the
// impact result and snapshot. Returns nil if no AI content is relevant.
func buildAIValidationSummary(result *impact.ImpactResult, snap *models.TestSuiteSnapshot) *AIValidationSummary {
	if len(snap.Scenarios) == 0 && len(result.ImpactedScenarios) == 0 {
		return nil
	}

	ai := &AIValidationSummary{
		TotalScenarios:    len(snap.Scenarios),
		SelectedScenarios: len(result.ImpactedScenarios),
	}

	// Collect impacted capabilities and scenario summaries.
	capSet := map[string]bool{}
	for _, is := range result.ImpactedScenarios {
		ai.Scenarios = append(ai.Scenarios, AIScenarioSummary{
			Name:       is.Name,
			Capability: is.Capability,
			Reason:     is.Relevance,
		})
		if is.Capability != "" {
			capSet[is.Capability] = true
		}
	}
	for cap := range capSet {
		ai.ImpactedCapabilities = append(ai.ImpactedCapabilities, cap)
	}
	sort.Strings(ai.ImpactedCapabilities)

	// Collect AI signals, split into blocking vs warning.
	for _, sig := range snap.Signals {
		if sig.Category != models.CategoryAI {
			continue
		}
		entry := AISignalSummary{
			Type: string(sig.Type), Severity: string(sig.Severity),
			Explanation: sig.Explanation,
		}
		if sig.Severity == models.SeverityCritical || sig.Severity == models.SeverityHigh {
			ai.BlockingSignals = append(ai.BlockingSignals, entry)
		} else {
			ai.WarningSignals = append(ai.WarningSignals, entry)
		}
	}

	// Find changed context surfaces that lack scenario coverage.
	coveredIDs := map[string]bool{}
	for _, sc := range snap.Scenarios {
		for _, sid := range sc.CoveredSurfaceIDs {
			coveredIDs[sid] = true
		}
	}
	changedPaths := map[string]bool{}
	for _, cf := range result.Scope.ChangedFiles {
		changedPaths[cf.Path] = true
	}
	for _, cs := range snap.CodeSurfaces {
		if cs.Kind == models.SurfaceContext && changedPaths[cs.Path] && !coveredIDs[cs.SurfaceID] {
			ai.UncoveredContexts = append(ai.UncoveredContexts, cs.Name+" ("+cs.Path+")")
		}
	}
	sort.Strings(ai.UncoveredContexts)

	// Return nil if nothing AI-relevant.
	if ai.SelectedScenarios == 0 && len(ai.BlockingSignals) == 0 &&
		len(ai.WarningSignals) == 0 && len(ai.UncoveredContexts) == 0 {
		return nil
	}

	return ai
}
