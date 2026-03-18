// Package impact implements Terrain's impact analysis framework.
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

import (
	"fmt"

	"github.com/pmclSF/terrain/internal/models"
)

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

	// Kind is the code unit kind (function, method, class, module).
	Kind string `json:"kind,omitempty"`

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

	// CoverageTypes describes the mix of coverage types for this unit.
	CoverageTypes *CoverageTypeInfo `json:"coverageTypes,omitempty"`

	// Complexity is the unit's complexity if known.
	Complexity float64 `json:"complexity,omitempty"`
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

// ImpactedScenario represents an AI/eval scenario whose covered surfaces
// were changed. This enables the prompt/dataset → scenario impact path.
type ImpactedScenario struct {
	// ScenarioID is the scenario identifier.
	ScenarioID string `json:"scenarioId"`

	// Name is the human-readable scenario name.
	Name string `json:"name"`

	// Category is the scenario classification (safety, accuracy, etc.).
	Category string `json:"category,omitempty"`

	// Framework is the eval framework (promptfoo, deepeval, etc.).
	Framework string `json:"framework,omitempty"`

	// Relevance explains why this scenario is impacted.
	Relevance string `json:"relevance"`

	// ImpactConfidence describes mapping confidence.
	ImpactConfidence Confidence `json:"impactConfidence"`

	// CoversSurfaces lists the changed surface IDs this scenario validates.
	CoversSurfaces []string `json:"coversSurfaces,omitempty"`

	// Capability is the inferred business capability this scenario validates.
	Capability string `json:"capability,omitempty"`
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

// ChangedArea groups changed code surfaces by domain area.
type ChangedArea struct {
	// Area is the domain area label (package, directory, or module name).
	Area string `json:"area"`

	// Surfaces lists the changed code surfaces in this area.
	Surfaces []ChangedSurface `json:"surfaces"`
}

// ChangedSurface is a code surface affected by the change.
type ChangedSurface struct {
	// SurfaceID is the stable surface identifier.
	SurfaceID string `json:"surfaceId"`

	// Name is the human-readable surface name.
	Name string `json:"name"`

	// Path is the file path.
	Path string `json:"path"`

	// Kind is the surface kind (function, method, handler, route, class).
	Kind string `json:"kind"`

	// ChangeKind is how the surface was affected.
	ChangeKind ChangeKind `json:"changeKind"`
}

// AffectedBehavior is a behavior surface impacted by the change.
type AffectedBehavior struct {
	// BehaviorID is the stable behavior identifier.
	BehaviorID string `json:"behaviorId"`

	// Label is the human-readable behavior label.
	Label string `json:"label"`

	// Kind is the derivation strategy (route_prefix, class, module, domain, naming).
	Kind string `json:"kind"`

	// ChangedSurfaceCount is how many of this behavior's surfaces were changed.
	ChangedSurfaceCount int `json:"changedSurfaceCount"`

	// TotalSurfaceCount is the total surfaces in this behavior group.
	TotalSurfaceCount int `json:"totalSurfaceCount"`
}

// ReasonCategories counts impacted tests by reason category.
type ReasonCategories struct {
	DirectDependency  int `json:"directDependency"`
	FixtureDependency int `json:"fixtureDependency"`
	DirectlyChanged   int `json:"directlyChanged"`
	DirectoryProximity int `json:"directoryProximity"`
}

// FallbackInfo describes the fallback strategy used, if any.
type FallbackInfo struct {
	// Level is the fallback strategy used ("none", "package", "directory", "all").
	Level string `json:"level"`

	// Reason explains why fallback was triggered.
	Reason string `json:"reason,omitempty"`

	// AdditionalTests is the count of tests added by fallback.
	AdditionalTests int `json:"additionalTests"`
}

// ImpactResult is the output of impact analysis.
type ImpactResult struct {
	// ChangeSet is the normalized change input when constructed via AnalyzeChangeSet.
	// Nil when constructed via the legacy Analyze() path.
	ChangeSet *models.ChangeSet `json:"changeSet,omitempty"`

	// Scope is the input change scope (legacy; preserved for backward compatibility).
	Scope ChangeScope `json:"scope"`

	// ChangedAreas groups changed code surfaces by domain area.
	ChangedAreas []ChangedArea `json:"changedAreas,omitempty"`

	// AffectedBehaviors lists behavior surfaces impacted by the change.
	AffectedBehaviors []AffectedBehavior `json:"affectedBehaviors,omitempty"`

	// ImpactedUnits lists affected code units.
	ImpactedUnits []ImpactedCodeUnit `json:"impactedUnits,omitempty"`

	// ImpactedTests lists tests relevant to the change.
	ImpactedTests []ImpactedTest `json:"impactedTests,omitempty"`

	// ImpactedScenarios lists AI/eval scenarios whose covered surfaces were changed.
	ImpactedScenarios []ImpactedScenario `json:"impactedScenarios,omitempty"`

	// ProtectionGaps identifies coverage gaps in the changed area.
	ProtectionGaps []ProtectionGap `json:"protectionGaps,omitempty"`

	// SelectedTests is the recommended protective test set.
	SelectedTests []ImpactedTest `json:"selectedTests,omitempty"`

	// ProtectiveSet is the enhanced protective test set with explanations.
	ProtectiveSet *ProtectiveTestSet `json:"protectiveSet,omitempty"`

	// Graph is the impact graph connecting code units to tests.
	Graph *ImpactGraph `json:"graph,omitempty"`

	// Posture is the change-risk assessment.
	Posture ChangeRiskPosture `json:"posture"`

	// CoverageConfidence is the overall coverage confidence band for the change.
	// Values: "high", "medium", "low".
	CoverageConfidence string `json:"coverageConfidence"`

	// ReasonCategories counts impacted tests by reason category.
	ReasonCategories ReasonCategories `json:"reasonCategories"`

	// Fallback describes the fallback strategy used, if any.
	Fallback FallbackInfo `json:"fallback"`

	// ImpactedOwners lists owners with impacted code.
	ImpactedOwners []string `json:"impactedOwners,omitempty"`

	// TotalTestCount is the total number of tests in the repository (for suite size context).
	TotalTestCount int `json:"totalTestCount,omitempty"`

	// Summary is a human-readable impact summary.
	Summary string `json:"summary"`

	// Limitations describes data gaps affecting the analysis.
	Limitations []string `json:"limitations,omitempty"`

	// PolicyApplied indicates whether an edge-case policy was applied.
	PolicyApplied bool `json:"policyApplied,omitempty"`

	// PolicyNotes explains how the policy affected this result.
	PolicyNotes []string `json:"policyNotes,omitempty"`
}

// ApplyEdgeCasePolicy adjusts an ImpactResult based on repo-level edge case
// policy. This is called after analysis to downgrade confidence and add
// warnings when the repo profile indicates reduced analysis reliability.
func (r *ImpactResult) ApplyEdgeCasePolicy(confidenceAdjustment float64, riskElevated bool, recommendations []string) {
	if confidenceAdjustment >= 1.0 && !riskElevated && len(recommendations) == 0 {
		return
	}

	r.PolicyApplied = true

	// Downgrade coverage confidence if adjustment is significant.
	if confidenceAdjustment < 0.7 {
		if r.CoverageConfidence == "high" {
			r.CoverageConfidence = "medium"
			r.PolicyNotes = append(r.PolicyNotes,
				"Coverage confidence downgraded from high to medium due to repo edge cases.")
		} else if r.CoverageConfidence == "medium" {
			r.CoverageConfidence = "low"
			r.PolicyNotes = append(r.PolicyNotes,
				"Coverage confidence downgraded from medium to low due to repo edge cases.")
		}
	}

	if riskElevated {
		r.PolicyNotes = append(r.PolicyNotes,
			"Risk elevated due to repo structural anomalies. Recommendations may be conservative.")
	}

	// Add policy recommendations as limitations.
	for _, rec := range recommendations {
		r.Limitations = append(r.Limitations, rec)
	}
}

// ApplyManualCoverageOverlay annotates the impact result with manual
// coverage information for changed areas. Manual coverage does NOT
// participate as executable CI validation — it is informational only,
// indicating that human QA covers areas where automated tests may be weak.
func (r *ImpactResult) ApplyManualCoverageOverlay(artifacts []models.ManualCoverageArtifact) {
	if len(artifacts) == 0 || len(r.ProtectionGaps) == 0 {
		return
	}

	// Build area index from manual coverage artifacts.
	areaArtifacts := map[string][]models.ManualCoverageArtifact{}
	for _, mc := range artifacts {
		if mc.Area != "" {
			areaArtifacts[mc.Area] = append(areaArtifacts[mc.Area], mc)
		}
	}

	// Check if any protection gaps overlap with manually covered areas.
	for _, gap := range r.ProtectionGaps {
		for area, mcs := range areaArtifacts {
			if matchesArea(gap.Path, area) {
				for _, mc := range mcs {
					r.PolicyNotes = append(r.PolicyNotes,
						fmt.Sprintf("Manual coverage exists for %s: %q (%s, %s criticality). Not executable — verify manually.",
							area, mc.Name, mc.Source, mc.Criticality))
				}
				break
			}
		}
	}
}

// matchesArea checks if a file path falls within a coverage area.
// Areas can be exact prefixes ("billing-core") or glob-like ("checkout/*").
func matchesArea(filePath, area string) bool {
	// Strip trailing wildcard for prefix match.
	prefix := area
	if len(prefix) > 0 && prefix[len(prefix)-1] == '*' {
		prefix = prefix[:len(prefix)-1]
	}
	// Prefix match on the file path or any of its directory components.
	return len(filePath) >= len(prefix) && filePath[:len(prefix)] == prefix
}

// AnalyzeChangeSet performs impact analysis starting from a ChangeSet.
// This is the preferred entry point — it normalizes the change input and
// carries ChangeSet metadata (SHAs, packages, services, limitations)
// through to the result.
func AnalyzeChangeSet(cs *models.ChangeSet, snap *models.TestSuiteSnapshot) *ImpactResult {
	scope := ChangeSetToScope(cs)
	result := analyzeFromScope(scope, snap)
	result.ChangeSet = cs

	// Merge ChangeSet limitations into result limitations.
	if len(cs.Limitations) > 0 {
		result.Limitations = append(cs.Limitations, result.Limitations...)
	}

	return result
}

// Analyze performs impact analysis given a change scope and snapshot.
// For new code, prefer AnalyzeChangeSet which provides richer metadata.
func Analyze(scope *ChangeScope, snap *models.TestSuiteSnapshot) *ImpactResult {
	return analyzeFromScope(scope, snap)
}

// analyzeFromScope is the shared implementation for both entry points.
func analyzeFromScope(scope *ChangeScope, snap *models.TestSuiteSnapshot) *ImpactResult {
	result := &ImpactResult{
		Scope:          *scope,
		TotalTestCount: len(snap.TestFiles),
	}

	// Map changed files to code surfaces and behavior surfaces.
	result.ChangedAreas = mapChangedSurfaces(scope, snap)
	result.AffectedBehaviors = mapAffectedBehaviors(result.ChangedAreas, snap)

	// Build impact graph for relationship lookups.
	result.Graph = BuildImpactGraph(snap)

	// Map changed files to code units.
	result.ImpactedUnits = mapChangedUnits(scope, snap)

	// Find impacted tests (using graph when available).
	result.ImpactedTests = findImpactedTests(scope, snap, result.ImpactedUnits)

	// Find impacted scenarios (prompt/dataset → scenario path).
	result.ImpactedScenarios = findImpactedScenarios(result.ChangedAreas, snap)

	// Identify protection gaps (enhanced with coverage diversity).
	result.ProtectionGaps = findProtectionGaps(result.ImpactedUnits, result.ImpactedTests, snap)

	// AI-specific protection gaps for changed AI surfaces without scenario coverage.
	result.ProtectionGaps = append(result.ProtectionGaps, findAIProtectionGaps(result, snap)...)

	// Select protective test set (legacy flat list for backward compatibility).
	result.SelectedTests = selectProtectiveTests(result.ImpactedTests, result.ImpactedUnits)

	// Build enhanced protective set with explanations.
	result.ProtectiveSet = buildProtectiveSet(result)

	// Compute change-risk posture.
	result.Posture = computeChangeRiskPosture(result)

	// Compute coverage confidence band.
	result.CoverageConfidence = computeCoverageConfidence(result)

	// Compute reason categories.
	result.ReasonCategories = computeReasonCategories(result.ImpactedTests)

	// Compute fallback info.
	result.Fallback = computeFallbackInfo(result)

	// Collect impacted owners.
	result.ImpactedOwners = collectOwners(result.ImpactedUnits)

	// Build summary.
	result.Summary = buildImpactSummary(result)

	// Note limitations.
	result.Limitations = identifyLimitations(scope, snap, result)

	return result
}
