package models

// ValidationKind describes the nature of a validation entity.
type ValidationKind string

const (
	// ValidationKindTest is an automated test case executed in CI.
	ValidationKindTest ValidationKind = "test"

	// ValidationKindScenario is a behavioral scenario (AI eval, multi-step
	// workflow, or derived behavior specification).
	ValidationKindScenario ValidationKind = "scenario"

	// ValidationKindManual is a manual coverage artifact — a QA checklist,
	// TestRail suite, or exploratory test session that exists outside CI.
	// Manual coverage is an overlay: it supplements automated coverage
	// but is never treated as executable CI coverage.
	ValidationKindManual ValidationKind = "manual"
)

// ValidationTarget is the common interface for all validation-bearing
// entities in the Terrain model. It unifies automated tests, AI evaluation
// scenarios, and manual coverage artifacts under a shared abstraction
// while preserving their type-specific metadata.
//
// This enables impact and coverage logic to operate generically over
// "things that validate behavior" without requiring type switches for
// every consumer. Concrete types (TestCase, Scenario, ManualCoverageArtifact)
// retain their full metadata — the interface does not flatten it away.
//
// Design principle: the interface exposes the minimal shared surface.
// Consumers that need type-specific data use a type assertion.
type ValidationTarget interface {
	// ValidationID returns a stable, deterministic identifier.
	ValidationID() string

	// ValidationName returns a human-readable label.
	ValidationName() string

	// ValidationKindOf returns the kind of validation this represents.
	ValidationKindOf() ValidationKind

	// ValidationPath returns the repository-relative file path, if applicable.
	// Returns "" for validation targets not tied to a specific file.
	ValidationPath() string

	// ValidationOwner returns the owner (team or individual), if known.
	// Returns "" when ownership is unresolved.
	ValidationOwner() string

	// IsExecutable returns true if this validation can be executed in CI.
	// Manual coverage artifacts return false.
	IsExecutable() bool
}

// --- TestCase implements ValidationTarget ---

func (tc TestCase) ValidationID() string        { return tc.TestID }
func (tc TestCase) ValidationName() string      { return tc.TestName }
func (tc TestCase) ValidationKindOf() ValidationKind { return ValidationKindTest }
func (tc TestCase) ValidationPath() string      { return tc.FilePath }
func (tc TestCase) ValidationOwner() string     { return "" } // Ownership resolved via TestFile.Owner
func (tc TestCase) IsExecutable() bool          { return true }

// Scenario represents a behavioral scenario — a multi-step workflow,
// AI evaluation case, or derived behavior specification that validates
// system behavior.
//
// Scenarios differ from TestCases in that they may be derived (inferred
// from code structure or AI-generated) rather than hand-written, and they
// may validate cross-cutting behavioral concerns rather than a single unit.
type Scenario struct {
	// ScenarioID is a stable identifier.
	// Format: "scenario:<path>:<name>" or "scenario:<category>:<hash>".
	ScenarioID string `json:"scenarioId"`

	// Name is a human-readable label for the scenario.
	Name string `json:"name"`

	// Description is a longer explanation of what behavior this scenario validates.
	Description string `json:"description,omitempty"`

	// Category classifies the scenario.
	// Examples: "happy_path", "edge_case", "adversarial", "safety", "regression".
	Category string `json:"category,omitempty"`

	// Path is the repository-relative file path, if tied to a specific file.
	Path string `json:"path,omitempty"`

	// Framework is the eval/test framework.
	// Examples: "deepeval", "promptfoo", "custom".
	Framework string `json:"framework,omitempty"`

	// Owner is the team or individual responsible for this scenario.
	Owner string `json:"owner,omitempty"`

	// CoveredSurfaceIDs lists the CodeSurface or BehaviorSurface IDs this
	// scenario is believed to exercise.
	CoveredSurfaceIDs []string `json:"coveredSurfaceIds,omitempty"`

	// EnvironmentIDs lists the environments this scenario targets.
	// Format: "env:<canonical-name>" matching Environment.EnvironmentID.
	EnvironmentIDs []string `json:"environmentIds,omitempty"`

	// Steps describes the ordered steps of the scenario, if applicable.
	Steps []string `json:"steps,omitempty"`

	// Executable indicates whether this scenario can be run in CI.
	// AI eval scenarios are typically executable; derived behavioral specs may not be.
	Executable bool `json:"executable"`
}

// --- Scenario implements ValidationTarget ---

func (s Scenario) ValidationID() string        { return s.ScenarioID }
func (s Scenario) ValidationName() string      { return s.Name }
func (s Scenario) ValidationKindOf() ValidationKind { return ValidationKindScenario }
func (s Scenario) ValidationPath() string      { return s.Path }
func (s Scenario) ValidationOwner() string     { return s.Owner }
func (s Scenario) IsExecutable() bool          { return s.Executable }

// ManualCoverageArtifact represents a validation activity that exists
// outside automated CI — a QA checklist, TestRail regression suite,
// exploratory testing session, or release sign-off procedure.
//
// Manual coverage is an overlay: it supplements automated coverage but
// is never treated as executable CI coverage. It carries a confidence
// adjustment relative to automated coverage (see doc 20).
type ManualCoverageArtifact struct {
	// ArtifactID is a stable identifier.
	// Format: "manual:<source>:<name-hash>".
	ArtifactID string `json:"artifactId"`

	// Name is a human-readable label.
	Name string `json:"name"`

	// Area is the code area or behavior surface this coverage applies to.
	// Examples: "auth/login", "checkout/payment", "admin/*".
	Area string `json:"area"`

	// Source identifies the origin system.
	// Values: "testrail", "jira", "qase", "checklist", "exploratory", "manual".
	Source string `json:"source"`

	// Owner is the team or individual responsible for executing this coverage.
	Owner string `json:"owner,omitempty"`

	// Criticality indicates how critical this manual coverage is to release confidence.
	// Values: "high", "medium", "low".
	Criticality string `json:"criticality,omitempty"`

	// LastExecuted is when this manual coverage was last executed.
	// Empty if unknown. Used for staleness detection.
	LastExecuted string `json:"lastExecuted,omitempty"`

	// Frequency is the expected execution cadence.
	// Values: "per-release", "weekly", "monthly", "ad-hoc".
	Frequency string `json:"frequency,omitempty"`

	// CoveredSurfaceIDs lists the CodeSurface or BehaviorSurface IDs this
	// artifact covers.
	CoveredSurfaceIDs []string `json:"coveredSurfaceIds,omitempty"`
}

// --- ManualCoverageArtifact implements ValidationTarget ---

func (m ManualCoverageArtifact) ValidationID() string        { return m.ArtifactID }
func (m ManualCoverageArtifact) ValidationName() string      { return m.Name }
func (m ManualCoverageArtifact) ValidationKindOf() ValidationKind { return ValidationKindManual }
func (m ManualCoverageArtifact) ValidationPath() string      { return "" }
func (m ManualCoverageArtifact) ValidationOwner() string     { return m.Owner }
func (m ManualCoverageArtifact) IsExecutable() bool          { return false }

// CollectValidationTargets aggregates all validation-bearing entities from
// a snapshot into a single slice. This is the canonical way to iterate over
// all validation in the system regardless of kind.
//
// The result preserves insertion order: tests first, then scenarios, then
// manual coverage artifacts.
func CollectValidationTargets(snap *TestSuiteSnapshot) []ValidationTarget {
	if snap == nil {
		return nil
	}

	var targets []ValidationTarget
	for i := range snap.TestCases {
		targets = append(targets, snap.TestCases[i])
	}
	for i := range snap.Scenarios {
		targets = append(targets, snap.Scenarios[i])
	}
	for i := range snap.ManualCoverage {
		targets = append(targets, snap.ManualCoverage[i])
	}
	return targets
}
