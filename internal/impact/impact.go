// Package impact implements Hamlet's impact analysis framework.
//
// Impact analysis answers: "If this code changes, which tests matter,
// what protection exists, and where are the gaps?"
//
// Core concepts:
//   - ChangeScope: what changed (files, code units, tests)
//   - ImpactResult: what is affected and what protection exists
//   - ImpactedCodeUnit: a code unit affected by the change
//   - ImpactedTest: a test relevant to the change
//   - ProtectionGap: where changed code lacks adequate test coverage
//   - ChangeRiskPosture: overall risk assessment for the change
//
// Impact analysis sits above:
//   - CodeUnit inventory
//   - test file/framework detection
//   - coverage lineage (when available)
//   - ownership resolution
//   - posture/measurement systems
//
// It differs from:
//   - Repo-wide posture: impact is scoped to a specific change
//   - PR reporting: impact provides the data; PR reporting renders it
//   - Portfolio intelligence: portfolio is cross-repo; impact is intra-repo
package impact

import "github.com/pmclSF/hamlet/internal/models"

// ChangeKind describes how an entity was changed.
type ChangeKind string

const (
	ChangeAdded    ChangeKind = "added"
	ChangeModified ChangeKind = "modified"
	ChangeDeleted  ChangeKind = "deleted"
	ChangeRenamed  ChangeKind = "renamed"
)

// Confidence describes how confident the impact mapping is.
type Confidence string

const (
	ConfidenceExact    Confidence = "exact"    // direct coverage lineage
	ConfidenceInferred Confidence = "inferred" // structural/heuristic mapping
	ConfidenceWeak     Confidence = "weak"     // best-effort fallback
)

// ChangeScope defines what changed in the repository.
type ChangeScope struct {
	// ChangedFiles lists files that were added, modified, deleted, or renamed.
	ChangedFiles []ChangedFile `json:"changedFiles"`

	// BaselineRef is the git ref or snapshot used as the baseline.
	BaselineRef string `json:"baselineRef,omitempty"`

	// CurrentRef is the git ref or snapshot representing the current state.
	CurrentRef string `json:"currentRef,omitempty"`

	// Source describes how the change scope was determined.
	// Values: "git-diff", "explicit", "ci-changed-files", "snapshot-compare"
	Source string `json:"source,omitempty"`
}

// ChangedFile represents a single changed file.
type ChangedFile struct {
	Path       string     `json:"path"`
	ChangeKind ChangeKind `json:"changeKind"`

	// OldPath is set when ChangeKind is "renamed".
	OldPath string `json:"oldPath,omitempty"`

	// IsTestFile indicates if this file is a test file.
	IsTestFile bool `json:"isTestFile"`
}

// ImpactedCodeUnit is a code unit affected by the change.
type ImpactedCodeUnit struct {
	// UnitID is the stable code unit identifier.
	UnitID string `json:"unitId"`

	// Name is the human-readable name.
	Name string `json:"name"`

	// Path is the file containing the code unit.
	Path string `json:"path"`

	// ChangeKind describes how the unit was affected.
	ChangeKind ChangeKind `json:"changeKind"`

	// Exported indicates if the unit is publicly visible.
	Exported bool `json:"exported"`

	// Owner is the resolved owner if known.
	Owner string `json:"owner,omitempty"`

	// ImpactConfidence describes mapping confidence.
	ImpactConfidence Confidence `json:"impactConfidence"`

	// ProtectionStatus summarizes test coverage for this unit.
	ProtectionStatus ProtectionStatus `json:"protectionStatus"`

	// CoveringTests lists test IDs that cover this unit.
	CoveringTests []string `json:"coveringTests,omitempty"`
}

// ProtectionStatus summarizes test protection for a code unit.
type ProtectionStatus string

const (
	ProtectionStrong  ProtectionStatus = "strong"  // unit + integration coverage
	ProtectionPartial ProtectionStatus = "partial" // some coverage but gaps exist
	ProtectionWeak    ProtectionStatus = "weak"    // only e2e or indirect coverage
	ProtectionNone    ProtectionStatus = "none"    // no observed coverage
)

// ImpactedTest is a test relevant to the change.
type ImpactedTest struct {
	// TestID is the stable test identifier if available.
	TestID string `json:"testId,omitempty"`

	// Path is the test file path.
	Path string `json:"path"`

	// Framework is the test framework.
	Framework string `json:"framework,omitempty"`

	// Relevance describes why this test is relevant.
	Relevance string `json:"relevance"`

	// ImpactConfidence describes mapping confidence.
	ImpactConfidence Confidence `json:"impactConfidence"`

	// CoversUnits lists code unit IDs this test covers.
	CoversUnits []string `json:"coversUnits,omitempty"`

	// IsDirectlyChanged indicates if the test file itself was changed.
	IsDirectlyChanged bool `json:"isDirectlyChanged"`
}

// ProtectionGap identifies where changed code lacks adequate coverage.
type ProtectionGap struct {
	// GapType describes the kind of protection gap.
	GapType string `json:"gapType"`

	// CodeUnitID is the affected code unit if applicable.
	CodeUnitID string `json:"codeUnitId,omitempty"`

	// Path is the affected file.
	Path string `json:"path"`

	// Explanation describes the gap.
	Explanation string `json:"explanation"`

	// Severity is "high", "medium", or "low".
	Severity string `json:"severity"`

	// SuggestedAction recommends a remediation.
	SuggestedAction string `json:"suggestedAction,omitempty"`
}

// ChangeRiskPosture summarizes the risk posture for the change.
type ChangeRiskPosture struct {
	// Band is the overall change-risk band.
	Band string `json:"band"`

	// Explanation describes why this band was assigned.
	Explanation string `json:"explanation"`

	// Dimensions breaks down the assessment.
	Dimensions []ChangeRiskDimension `json:"dimensions,omitempty"`
}

// ChangeRiskDimension is one aspect of change-risk assessment.
type ChangeRiskDimension struct {
	Name        string `json:"name"`
	Band        string `json:"band"`
	Explanation string `json:"explanation"`
}

// ImpactResult is the output of impact analysis.
type ImpactResult struct {
	// Scope is the input change scope.
	Scope ChangeScope `json:"scope"`

	// ImpactedUnits lists affected code units.
	ImpactedUnits []ImpactedCodeUnit `json:"impactedUnits,omitempty"`

	// ImpactedTests lists tests relevant to the change.
	ImpactedTests []ImpactedTest `json:"impactedTests,omitempty"`

	// ProtectionGaps identifies coverage gaps in the changed area.
	ProtectionGaps []ProtectionGap `json:"protectionGaps,omitempty"`

	// SelectedTests is the recommended protective test set.
	SelectedTests []ImpactedTest `json:"selectedTests,omitempty"`

	// Posture is the change-risk assessment.
	Posture ChangeRiskPosture `json:"posture"`

	// ImpactedOwners lists owners with impacted code.
	ImpactedOwners []string `json:"impactedOwners,omitempty"`

	// Summary is a human-readable impact summary.
	Summary string `json:"summary"`

	// Limitations describes data gaps affecting the analysis.
	Limitations []string `json:"limitations,omitempty"`
}

// Analyze performs impact analysis given a change scope and snapshot.
func Analyze(scope *ChangeScope, snap *models.TestSuiteSnapshot) *ImpactResult {
	result := &ImpactResult{
		Scope: *scope,
	}

	// Map changed files to code units.
	result.ImpactedUnits = mapChangedUnits(scope, snap)

	// Find impacted tests.
	result.ImpactedTests = findImpactedTests(scope, snap, result.ImpactedUnits)

	// Identify protection gaps.
	result.ProtectionGaps = findProtectionGaps(result.ImpactedUnits, result.ImpactedTests, snap)

	// Select protective test set.
	result.SelectedTests = selectProtectiveTests(result.ImpactedTests, result.ImpactedUnits)

	// Compute change-risk posture.
	result.Posture = computeChangeRiskPosture(result)

	// Collect impacted owners.
	result.ImpactedOwners = collectOwners(result.ImpactedUnits)

	// Build summary.
	result.Summary = buildImpactSummary(result)

	// Note limitations.
	result.Limitations = identifyLimitations(scope, snap, result)

	return result
}
