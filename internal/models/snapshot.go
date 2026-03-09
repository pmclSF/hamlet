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

// SnapshotSchemaVersion is the current schema version for TestSuiteSnapshot.
// Increment this when the snapshot JSON shape changes in a breaking way.
const SnapshotSchemaVersion = "1.0.0"

// SnapshotMeta holds machine-readable provenance for the snapshot artifact.
type SnapshotMeta struct {
	// SchemaVersion identifies the snapshot JSON schema version.
	SchemaVersion string `json:"schemaVersion"`

	// EngineVersion is the hamlet binary version that produced this snapshot.
	EngineVersion string `json:"engineVersion,omitempty"`

	// DetectorCount is the number of detectors that ran during analysis.
	DetectorCount int `json:"detectorCount,omitempty"`

	// Detectors lists the IDs of detectors that ran during analysis.
	Detectors []string `json:"detectors,omitempty"`
}

// TestSuiteSnapshot is the canonical output artifact of Hamlet analysis.
//
// This is the main serialization boundary for:
//   - CLI JSON output
//   - local snapshot persistence
//   - extension rendering
//   - future hosted ingestion
type TestSuiteSnapshot struct {
	// SnapshotMeta holds schema version and engine provenance.
	SnapshotMeta SnapshotMeta `json:"snapshotMeta"`

	Repository RepositoryMetadata `json:"repository"`

	Frameworks []Framework `json:"frameworks,omitempty"`

	TestFiles []TestFile `json:"testFiles,omitempty"`

	// TestCases contains individually identified test cases with stable IDs.
	// This enables longitudinal tracking across snapshots.
	TestCases []TestCase `json:"testCases,omitempty"`

	CodeUnits []CodeUnit `json:"codeUnits,omitempty"`

	Signals []Signal `json:"signals,omitempty"`

	Risk []RiskSurface `json:"risk,omitempty"`

	// Measurements contains the measurement-layer snapshot when computed.
	Measurements *MeasurementSnapshot `json:"measurements,omitempty"`

	// Portfolio contains portfolio intelligence results when computed.
	Portfolio *PortfolioSnapshot `json:"portfolio,omitempty"`

	// CoverageSummary holds aggregated coverage statistics when coverage
	// artifacts have been ingested.
	CoverageSummary *CoverageSummary `json:"coverageSummary,omitempty"`

	// CoverageInsights holds actionable findings derived from coverage analysis.
	CoverageInsights []CoverageInsight `json:"coverageInsights,omitempty"`

	Ownership map[string][]string `json:"ownership,omitempty"`

	Policies map[string]any `json:"policies,omitempty"`

	Metadata map[string]any `json:"metadata,omitempty"`

	GeneratedAt time.Time `json:"generatedAt"`
}

// CoverageSummary holds aggregated coverage statistics for the snapshot.
type CoverageSummary struct {
	// TotalCodeUnits is the total number of discovered code units.
	TotalCodeUnits int `json:"totalCodeUnits"`

	// CoveredByUnitTests is the count covered by unit tests.
	CoveredByUnitTests int `json:"coveredByUnitTests"`

	// CoveredByIntegration is the count covered by integration tests.
	CoveredByIntegration int `json:"coveredByIntegration"`

	// CoveredByE2E is the count covered by e2e tests.
	CoveredByE2E int `json:"coveredByE2e"`

	// CoveredOnlyByE2E is the count covered exclusively by e2e tests.
	CoveredOnlyByE2E int `json:"coveredOnlyByE2e"`

	// UncoveredExported is the count of exported units with no coverage.
	UncoveredExported int `json:"uncoveredExported"`

	// Uncovered is the total count of units with no coverage.
	Uncovered int `json:"uncovered"`

	// LineCoveragePct is the overall line coverage percentage.
	LineCoveragePct float64 `json:"lineCoveragePct,omitempty"`

	// BranchCoveragePct is the overall branch coverage percentage.
	BranchCoveragePct float64 `json:"branchCoveragePct,omitempty"`
}

// CoverageInsight represents an actionable finding from coverage analysis.
type CoverageInsight struct {
	Type            string `json:"type"`
	Severity        string `json:"severity"`
	Description     string `json:"description"`
	Path            string `json:"path,omitempty"`
	UnitID          string `json:"unitId,omitempty"`
	SuggestedAction string `json:"suggestedAction,omitempty"`
}
