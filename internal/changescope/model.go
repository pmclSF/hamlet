// Package changescope provides PR and change-scoped analysis workflows.
//
// It builds on the impact subsystem to produce focused outputs suitable
// for PR reviews, CI gating, and incremental development workflows.
package changescope

import "github.com/pmclSF/hamlet/internal/impact"

// PRAnalysis is the output of a PR/change-scoped analysis.
type PRAnalysis struct {
	// Scope is the change scope used for analysis.
	Scope impact.ChangeScope `json:"scope"`

	// Summary is a concise one-line summary.
	Summary string `json:"summary"`

	// PostureBand is the change-risk posture band.
	PostureBand string `json:"postureBand"`

	// ChangedFileCount is the number of changed files.
	ChangedFileCount int `json:"changedFileCount"`

	// ChangedTestCount is the number of changed test files.
	ChangedTestCount int `json:"changedTestCount"`

	// ChangedSourceCount is the number of changed source files.
	ChangedSourceCount int `json:"changedSourceCount"`

	// ImpactedUnitCount is the number of impacted code units.
	ImpactedUnitCount int `json:"impactedUnitCount"`

	// ProtectionGapCount is the number of protection gaps.
	ProtectionGapCount int `json:"protectionGapCount"`

	// NewFindings are findings specific to the changed area.
	NewFindings []ChangeScopedFinding `json:"newFindings,omitempty"`

	// AffectedOwners lists owners with impacted code.
	AffectedOwners []string `json:"affectedOwners,omitempty"`

	// RecommendedTests are the tests to run for this change (paths only, for backward compat).
	RecommendedTests []string `json:"recommendedTests,omitempty"`

	// TestSelections are the recommended tests with reasoning.
	TestSelections []TestSelection `json:"testSelections,omitempty"`

	// SelectionStrategy describes how the test set was chosen.
	SelectionStrategy string `json:"selectionStrategy,omitempty"`

	// SelectionExplanation describes why this strategy was used.
	SelectionExplanation string `json:"selectionExplanation,omitempty"`

	// PostureDelta describes posture changes in the affected area.
	PostureDelta *PostureDelta `json:"postureDelta,omitempty"`

	// Limitations notes data gaps.
	Limitations []string `json:"limitations,omitempty"`

	// ImpactResult is the full impact analysis result.
	ImpactResult *impact.ImpactResult `json:"-"`
}

// TestSelection is a recommended test with reasoning about why it was selected.
type TestSelection struct {
	// Path is the test file path.
	Path string `json:"path"`
	// Confidence is the impact mapping confidence (exact, inferred, weak).
	Confidence string `json:"confidence"`
	// Relevance describes why this test is relevant.
	Relevance string `json:"relevance"`
	// CoversUnits lists code unit names this test protects.
	CoversUnits []string `json:"coversUnits,omitempty"`
	// Reasons are structured selection reasons.
	Reasons []string `json:"reasons,omitempty"`
}

// ChangeScopedFinding is a finding relevant to the changed area.
type ChangeScopedFinding struct {
	// Type is the finding type (e.g., "protection_gap", "new_signal", "worsened_coverage").
	Type string `json:"type"`
	// Path is the file path.
	Path string `json:"path"`
	// Severity is "high", "medium", or "low".
	Severity string `json:"severity"`
	// Explanation describes the finding.
	Explanation string `json:"explanation"`
	// SuggestedAction recommends a fix.
	SuggestedAction string `json:"suggestedAction,omitempty"`
}

// PostureDelta describes how posture changed in the affected area.
type PostureDelta struct {
	// OverallDirection is "improved", "worsened", or "unchanged".
	OverallDirection string `json:"overallDirection"`
	// Explanation describes the delta.
	Explanation string `json:"explanation"`
	// NewGapCount is the number of new protection gaps.
	NewGapCount int `json:"newGapCount"`
	// ResolvedGapCount is the number of resolved gaps (if baseline available).
	ResolvedGapCount int `json:"resolvedGapCount"`
}
