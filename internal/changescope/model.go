// Package changescope provides PR and change-scoped analysis workflows.
//
// It builds on the impact subsystem to produce focused outputs suitable
// for PR reviews, CI gating, and incremental development workflows.
package changescope

import "github.com/pmclSF/terrain/internal/impact"

// PRAnalysisSchemaVersion is the current schema version for PR analysis artifacts.
const PRAnalysisSchemaVersion = "2"

// PRAnalysis is the output of a PR/change-scoped analysis.
type PRAnalysis struct {
	// SchemaVersion identifies the PR analysis JSON schema version.
	SchemaVersion string `json:"schemaVersion"`

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

	// TotalTestCount is the total number of tests in the repository.
	TotalTestCount int `json:"totalTestCount"`

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

	// AI holds the AI validation summary for this PR.
	AI *AIValidationSummary `json:"ai,omitempty"`

	// ImpactResult is the full impact analysis result.
	ImpactResult *impact.ImpactResult `json:"-"`
}

// AIValidationSummary captures AI-specific validation state for a PR.
type AIValidationSummary struct {
	// ImpactedCapabilities lists business capabilities affected by this change.
	ImpactedCapabilities []string `json:"impactedCapabilities,omitempty"`

	// SelectedScenarios is the number of AI scenarios selected for this change.
	SelectedScenarios int `json:"selectedScenarios"`

	// TotalScenarios is the total number of AI scenarios in the repo.
	TotalScenarios int `json:"totalScenarios"`

	// Scenarios lists impacted scenarios with reasons.
	Scenarios []AIScenarioSummary `json:"scenarios,omitempty"`

	// BlockingSignals lists AI signals that block the merge.
	BlockingSignals []AISignalSummary `json:"blockingSignals,omitempty"`

	// WarningSignals lists AI signals that warn but don't block.
	WarningSignals []AISignalSummary `json:"warningSignals,omitempty"`

	// UncoveredContexts lists changed context surfaces with no scenario coverage.
	UncoveredContexts []string `json:"uncoveredContexts,omitempty"`
}

// AIScenarioSummary is a compact scenario entry for PR display.
type AIScenarioSummary struct {
	Name       string `json:"name"`
	Capability string `json:"capability,omitempty"`
	Reason     string `json:"reason"`
}

// AISignalSummary is a compact signal entry for PR display.
type AISignalSummary struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Explanation string `json:"explanation"`
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
	// Type is the finding type:
	//   "protection_gap"  — coverage gap directly from changed code
	//   "existing_signal" — pre-existing issue on a file touched by the change graph
	Type string `json:"type"`
	// Scope distinguishes change proximity:
	//   "direct"   — file was directly changed in this PR
	//   "indirect" — file was reached via the impact graph (transitive dependency)
	Scope string `json:"scope,omitempty"`
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
