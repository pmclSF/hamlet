package changescope

import (
	"fmt"
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
		Scope:        *scope,
		ImpactResult: result,
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

	// Extract recommended tests with reasoning.
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
		// Fall back to SelectedTests (no detailed reasons).
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

	// Build change-scoped findings from protection gaps and signals.
	pr.NewFindings = buildChangeScopedFindings(result, snap)

	// Build posture delta.
	pr.PostureDelta = buildPostureDelta(result)

	// Build summary.
	pr.Summary = buildPRSummary(pr)

	return pr
}

// AnalyzePRFromChangeSet performs a PR/change-scoped analysis starting from
// a ChangeSet. This is the preferred entry point for new code.
func AnalyzePRFromChangeSet(cs *models.ChangeSet, snap *models.TestSuiteSnapshot) *PRAnalysis {
	result := impact.AnalyzeChangeSet(cs, snap)

	pr := &PRAnalysis{
		Scope:        result.Scope,
		ImpactResult: result,
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

	pr.NewFindings = buildChangeScopedFindings(result, snap)
	pr.PostureDelta = buildPostureDelta(result)
	pr.Summary = buildPRSummary(pr)

	return pr
}

// AnalyzeChangedPaths is a convenience function that creates a change scope
// from explicit paths and runs PR analysis.
func AnalyzeChangedPaths(paths []string, changeKind impact.ChangeKind, snap *models.TestSuiteSnapshot) *PRAnalysis {
	scope := impact.ChangeScopeFromPaths(paths, changeKind)
	return AnalyzePR(scope, snap)
}

func buildChangeScopedFindings(result *impact.ImpactResult, snap *models.TestSuiteSnapshot) []ChangeScopedFinding {
	var findings []ChangeScopedFinding

	// Convert protection gaps to findings.
	for _, gap := range result.ProtectionGaps {
		findings = append(findings, ChangeScopedFinding{
			Type:            "protection_gap",
			Path:            gap.Path,
			Severity:        gap.Severity,
			Explanation:     gap.Explanation,
			SuggestedAction: gap.SuggestedAction,
		})
	}

	// Check for signals on changed files, but skip signals that duplicate
	// protection gaps already surfaced above (e.g., untestedExport signals
	// overlap with untested_export protection gaps for the same path).
	gapPaths := map[string]bool{}
	for _, gap := range result.ProtectionGaps {
		gapPaths[gap.Path] = true
	}

	changedPaths := map[string]bool{}
	for _, cf := range result.Scope.ChangedFiles {
		changedPaths[cf.Path] = true
	}
	for _, sig := range snap.Signals {
		if !changedPaths[sig.Location.File] {
			continue
		}
		// Skip untestedExport signals when a protection gap already covers this path.
		if sig.Type == "untestedExport" && gapPaths[sig.Location.File] {
			continue
		}
		findings = append(findings, ChangeScopedFinding{
			Type:        "existing_signal",
			Path:        sig.Location.File,
			Severity:    string(sig.Severity),
			Explanation: fmt.Sprintf("[%s] %s", sig.Type, sig.Explanation),
		})
	}

	// Note: untested exported units are already surfaced via protection gaps
	// (gapType "untested_export" with severity "high"). We don't duplicate
	// them here to avoid showing the same issue twice in PR comments.

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
