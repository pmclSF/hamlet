// Package measurement implements Hamlet's formal measurement layer.
//
// Measurements sit between raw signals/derived facts and user-facing
// summaries. They provide:
//   - explicit, named, versioned computations
//   - evidence strength and confidence metadata
//   - posture dimension attribution
//   - explanation metadata for diagnostics
//
// Philosophy:
//   - evidence-based: every measurement traces to concrete signals or facts
//   - actionable: measurements should suggest what to do, not just report
//   - stable: measurement IDs and semantics are durable across versions
//   - honest about uncertainty: missing data reduces confidence, not value
//   - no fake precision: prefer bands and ratios over decimal scores
package measurement

import "github.com/pmclSF/hamlet/internal/models"

// Dimension identifies which posture dimension a measurement feeds.
type Dimension string

const (
	DimensionHealth           Dimension = "health"
	DimensionCoverageDepth    Dimension = "coverage_depth"
	DimensionCoverageDiversity Dimension = "coverage_diversity"
	DimensionStructuralRisk   Dimension = "structural_risk"
	DimensionOperationalRisk  Dimension = "operational_risk"
)

// EvidenceStrength describes how much concrete evidence backs a measurement.
type EvidenceStrength string

const (
	EvidenceStrong  EvidenceStrength = "strong"  // direct observation, high confidence
	EvidencePartial EvidenceStrength = "partial" // some data available, gaps noted
	EvidenceWeak    EvidenceStrength = "weak"    // limited data, measurement is best-effort
	EvidenceNone    EvidenceStrength = "none"    // no data available for this measurement
)

// Units describes the type of value a measurement produces.
type Units string

const (
	UnitsRatio   Units = "ratio"   // 0.0 to 1.0
	UnitsCount   Units = "count"   // integer count
	UnitsBand    Units = "band"    // qualitative band (strong/moderate/weak/elevated/critical)
	UnitsPercent Units = "percent" // 0 to 100
)

// Result is the output of a single measurement computation.
type Result struct {
	// ID is the stable measurement identifier (e.g. "health.flaky_share").
	ID string `json:"id"`

	// Dimension is the posture dimension this measurement feeds.
	Dimension Dimension `json:"dimension"`

	// Value is the computed measurement value.
	Value float64 `json:"value"`

	// Units describes what the value represents.
	Units Units `json:"units"`

	// Band is an optional qualitative interpretation of the value.
	// Present when Units is "band" or when the measurement defines thresholds.
	Band string `json:"band,omitempty"`

	// Evidence describes the strength of the data backing this measurement.
	Evidence EvidenceStrength `json:"evidence"`

	// Explanation is a human-readable summary of what was measured and why.
	Explanation string `json:"explanation"`

	// Inputs lists the signal types or data sources that fed this measurement.
	Inputs []string `json:"inputs,omitempty"`

	// Limitations describes any data gaps or caveats.
	Limitations []string `json:"limitations,omitempty"`
}

// PostureBand represents the qualitative state of a posture dimension.
type PostureBand string

const (
	PostureStrong   PostureBand = "strong"
	PostureModerate PostureBand = "moderate"
	PostureWeak     PostureBand = "weak"
	PostureElevated PostureBand = "elevated"
	PostureCritical PostureBand = "critical"
	PostureUnknown  PostureBand = "unknown"
)

// DimensionPosture is the computed posture for a single dimension.
type DimensionPosture struct {
	// Dimension is the posture dimension.
	Dimension Dimension `json:"dimension"`

	// Band is the overall posture band for this dimension.
	Band PostureBand `json:"band"`

	// Explanation describes why this posture was assigned.
	Explanation string `json:"explanation"`

	// DrivingMeasurements lists the measurement IDs that most influenced
	// the posture computation.
	DrivingMeasurements []string `json:"drivingMeasurements,omitempty"`

	// Measurements contains all measurement results for this dimension.
	Measurements []Result `json:"measurements,omitempty"`
}

// Snapshot captures all measurements and posture results for a point in time.
// This is the measurement layer's output artifact.
type Snapshot struct {
	// Posture contains per-dimension posture results.
	Posture []DimensionPosture `json:"posture"`

	// Measurements contains all individual measurement results.
	Measurements []Result `json:"measurements"`
}

// Definition describes a measurement that can be computed from a snapshot.
type Definition struct {
	// ID is the stable measurement identifier.
	ID string

	// Dimension is the posture dimension this measurement feeds.
	Dimension Dimension

	// Description is a short human-readable summary.
	Description string

	// Units describes the output type.
	Units Units

	// Inputs lists the signal types or data sources this measurement reads.
	Inputs []string

	// Compute executes the measurement against a snapshot.
	Compute func(snap *models.TestSuiteSnapshot) Result
}
