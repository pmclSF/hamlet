package models

import "time"

// RuntimeStats captures runtime evidence for a test file, suite, or other
// execution scope when CI or runner artifacts are available.
type RuntimeStats struct {
	AvgRuntimeMs    float64 `json:"avgRuntimeMs,omitempty"`
	P95RuntimeMs    float64 `json:"p95RuntimeMs,omitempty"`
	PassRate        float64 `json:"passRate,omitempty"`
	RetryRate       float64 `json:"retryRate,omitempty"`
	RuntimeVariance float64 `json:"runtimeVariance,omitempty"`
}

// RiskBand represents the qualitative severity of a risk surface.
type RiskBand string

const (
	RiskBandLow      RiskBand = "low"
	RiskBandMedium   RiskBand = "medium"
	RiskBandHigh     RiskBand = "high"
	RiskBandCritical RiskBand = "critical"
)

// RiskSurface represents an explainable risk output over a scope such as a
// file, package, module, team, or repository.
type RiskSurface struct {
	// Type is the risk dimension.
	// Examples: reliability, change, speed, governance.
	Type string `json:"type"`

	// Scope identifies where the risk applies.
	// Examples: repo, package, file, module, owner.
	Scope string `json:"scope"`

	// ScopeName names the concrete entity within the scope.
	ScopeName string `json:"scopeName"`

	// Band is the qualitative risk band.
	Band RiskBand `json:"band"`

	// Score is an optional normalized risk score.
	Score float64 `json:"score,omitempty"`

	// ContributingSignals are the signal identifiers or signal objects that
	// materially contribute to this risk surface.
	ContributingSignals []Signal `json:"contributingSignals,omitempty"`

	// Explanation summarizes why this risk surface exists.
	Explanation string `json:"explanation,omitempty"`

	// SuggestedAction gives the next useful remediation direction.
	SuggestedAction string `json:"suggestedAction,omitempty"`
}

// TestSuiteSnapshot is the canonical output artifact of Hamlet analysis.
//
// This is the main serialization boundary for:
//   - CLI JSON output
//   - local snapshot persistence
//   - extension rendering
//   - future hosted ingestion
type TestSuiteSnapshot struct {
	Repository RepositoryMetadata `json:"repository"`

	Frameworks []Framework `json:"frameworks,omitempty"`

	TestFiles []TestFile `json:"testFiles,omitempty"`

	CodeUnits []CodeUnit `json:"codeUnits,omitempty"`

	Signals []Signal `json:"signals,omitempty"`

	Risk []RiskSurface `json:"risk,omitempty"`

	// Measurements contains the measurement-layer snapshot when computed.
	Measurements *MeasurementSnapshot `json:"measurements,omitempty"`

	Ownership map[string][]string `json:"ownership,omitempty"`

	Policies map[string]any `json:"policies,omitempty"`

	Metadata map[string]any `json:"metadata,omitempty"`

	GeneratedAt time.Time `json:"generatedAt"`
}
