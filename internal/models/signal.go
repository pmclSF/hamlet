package models

// SignalType is the canonical identifier for a Hamlet signal.
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
}

// Signal is the canonical structured insight type in Hamlet.
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

	// Confidence indicates how certain Hamlet is about the signal.
	// Expected range is 0.0 to 1.0.
	Confidence float64 `json:"confidence,omitempty"`

	Location SignalLocation `json:"location"`

	Owner string `json:"owner,omitempty"`

	Explanation string `json:"explanation"`

	SuggestedAction string `json:"suggestedAction,omitempty"`

	Metadata map[string]any `json:"metadata,omitempty"`
}
