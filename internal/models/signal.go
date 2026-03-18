package models

// SignalType is the canonical identifier for a Terrain signal.
// Signal type constants are defined in internal/signals for registry use,
// but the type itself lives here so the snapshot model is self-contained.
type SignalType string

// SignalCategory groups signal types into major product areas.
type SignalCategory string

const (
	CategoryStructure  SignalCategory = "structure"
	CategoryHealth     SignalCategory = "health"
	CategoryQuality    SignalCategory = "quality"
	CategoryMigration  SignalCategory = "migration"
	CategoryGovernance SignalCategory = "governance"
	CategoryAI         SignalCategory = "ai"
)

// SignalSeverity expresses how urgent or important a signal is.
type SignalSeverity string

const (
	SeverityInfo     SignalSeverity = "info"
	SeverityLow      SignalSeverity = "low"
	SeverityMedium   SignalSeverity = "medium"
	SeverityHigh     SignalSeverity = "high"
	SeverityCritical SignalSeverity = "critical"
)

// SignalLocation identifies where a signal applies.
type SignalLocation struct {
	Repository string `json:"repository,omitempty"`
	Package    string `json:"package,omitempty"`
	File       string `json:"file,omitempty"`
	Symbol     string `json:"symbol,omitempty"`
	Line       int    `json:"line,omitempty"`

	// ScenarioID links the signal to a specific AI/eval scenario.
	ScenarioID string `json:"scenarioId,omitempty"`

	// Capability links the signal to a business capability.
	Capability string `json:"capability,omitempty"`
}

// EvidenceStrength describes how robust the evidence behind a signal is.
type EvidenceStrength string

const (
	EvidenceStrong   EvidenceStrength = "strong"   // AST-backed, coverage data, runtime observation
	EvidenceModerate EvidenceStrength = "moderate" // structural pattern matching with context
	EvidencePartial  EvidenceStrength = "partial"  // some evidence available, with known gaps
	EvidenceWeak     EvidenceStrength = "weak"     // path/name heuristic only
	EvidenceNone     EvidenceStrength = "none"     // no supporting evidence available
)

// EvidenceSource describes how the signal was derived.
type EvidenceSource string

const (
	SourceAST               EvidenceSource = "ast"
	SourceStructuralPattern EvidenceSource = "structural-pattern"
	SourcePathName          EvidenceSource = "path-name"
	SourceRuntime           EvidenceSource = "runtime"
	SourceCoverage          EvidenceSource = "coverage"
	SourcePolicy            EvidenceSource = "policy"
	SourceCodeowners        EvidenceSource = "codeowners"
	SourceEvalExecution     EvidenceSource = "eval-execution"
)

// Signal is the canonical structured insight type in Terrain.
//
// Every meaningful user-facing finding should be representable as a Signal.
// This type lives in models because it is a core part of TestSuiteSnapshot
// and must be serializable without circular imports.
//
// Signals are designed to be:
//   - explainable
//   - serializable
//   - composable
//   - renderable in multiple surfaces (CLI, extension, CI)
type Signal struct {
	Type     SignalType     `json:"type"`
	Category SignalCategory `json:"category"`
	Severity SignalSeverity `json:"severity"`

	// Confidence indicates how certain Terrain is about the signal.
	// Expected range is 0.0 to 1.0.
	Confidence float64 `json:"confidence,omitempty"`

	// EvidenceStrength classifies the robustness of the evidence.
	// Weak-evidence signals are rendered with appropriate caveats.
	EvidenceStrength EvidenceStrength `json:"evidenceStrength,omitempty"`

	// EvidenceSource identifies how the signal was derived.
	EvidenceSource EvidenceSource `json:"evidenceSource,omitempty"`

	Location SignalLocation `json:"location"`

	Owner string `json:"owner,omitempty"`

	Explanation string `json:"explanation"`

	SuggestedAction string `json:"suggestedAction,omitempty"`

	Metadata map[string]any `json:"metadata,omitempty"`
}
