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

	// EngineVersion is the terrain binary version that produced this snapshot.
	EngineVersion string `json:"engineVersion,omitempty"`

	// DetectorCount is the number of detectors that ran during analysis.
	DetectorCount int `json:"detectorCount,omitempty"`

	// Detectors lists the IDs of detectors that ran during analysis.
	Detectors []string `json:"detectors,omitempty"`

	// MethodologyFingerprint identifies detector/measurement/risk-model
	// methodology used to produce this snapshot.
	MethodologyFingerprint string `json:"methodologyFingerprint,omitempty"`
}

// TestSuiteSnapshot is the canonical output artifact of Terrain analysis.
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

	// CodeSurfaces contains inferred behavior anchors — the points in code
	// where observable behavior originates. Inferred automatically from
	// exported functions, methods, handlers, and routes.
	CodeSurfaces []CodeSurface `json:"codeSurfaces,omitempty"`

	// BehaviorSurfaces contains derived behavior groupings that aggregate
	// related CodeSurfaces into higher-level behavioral units. Optional —
	// all analysis pipelines work with or without them.
	BehaviorSurfaces []BehaviorSurface `json:"behaviorSurfaces,omitempty"`

	// FixtureSurfaces contains detected shared test fixtures — setup hooks,
	// builders, mock providers, and data loaders. Used for fanout analysis
	// to identify fragile validation structures.
	FixtureSurfaces []FixtureSurface `json:"fixtureSurfaces,omitempty"`

	// RAGPipelineSurfaces contains detected RAG pipeline components with
	// extracted configuration metadata. Enables structured reasoning about
	// retrieval pipelines: chunking strategy, top-k values, embedding models.
	RAGPipelineSurfaces []RAGPipelineSurface `json:"ragPipelineSurfaces,omitempty"`

	// Scenarios contains behavioral scenarios — AI evaluation cases,
	// multi-step workflows, or derived behavior specifications.
	Scenarios []Scenario `json:"scenarios,omitempty"`

	// InferredCapabilities lists AI capabilities detected in the codebase,
	// derived from code surface kinds rather than scenario naming.
	// This enables capability-level impact reporting independent of scenario coverage.
	InferredCapabilities []InferredCapability `json:"inferredCapabilities,omitempty"`

	// ManualCoverage contains manual coverage artifacts — QA checklists,
	// TestRail suites, and other validation that exists outside CI.
	// Manual coverage is an overlay, not executable CI coverage.
	ManualCoverage []ManualCoverageArtifact `json:"manualCoverage,omitempty"`

	// Environments contains concrete execution contexts where tests run.
	// Inferred from CI configuration files or declared in terrain.yaml.
	Environments []Environment `json:"environments,omitempty"`

	// EnvironmentClasses contains groups of related environments that
	// share common characteristics (e.g., all browsers, all OS variants).
	EnvironmentClasses []EnvironmentClass `json:"environmentClasses,omitempty"`

	// DeviceConfigs contains target device or browser configurations
	// where tests execute (phones, tablets, browsers, emulators).
	DeviceConfigs []DeviceConfig `json:"deviceConfigs,omitempty"`

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

	// ImportGraph maps test file paths to their resolved source module imports.
	// Used by quality detectors for precise test-to-code linkage.
	// Omitted from JSON output to keep snapshots compact.
	ImportGraph map[string]map[string]bool `json:"-"`

	// SourceImports maps source file paths to their resolved source imports.
	// Enables accurate transitive impact: A imports B imports C → change to C impacts A.
	// Omitted from JSON output.
	SourceImports map[string]map[string]bool `json:"-"`

	// DataSources tracks which data sources were attempted during analysis,
	// whether they succeeded, and what impact their absence has on results.
	DataSources []DataSource `json:"dataSources,omitempty"`

	Ownership map[string][]string `json:"ownership,omitempty"`

	Policies map[string]any `json:"policies,omitempty"`

	Metadata map[string]any `json:"metadata,omitempty"`

	GeneratedAt time.Time `json:"generatedAt"`
}

// DataSource tracks the status of a data source attempted during analysis.
type DataSource struct {
	// Name identifies the data source (e.g. "runtime", "coverage", "policy").
	Name string `json:"name"`

	// Status is "available", "unavailable", or "error".
	Status string `json:"status"`

	// Detail provides context (path attempted, error message, etc.).
	Detail string `json:"detail,omitempty"`

	// Impact describes what analysis is affected by this source's absence.
	Impact string `json:"impact,omitempty"`
}

// DataSource status constants.
const (
	DataSourceAvailable   = "available"
	DataSourceUnavailable = "unavailable"
	DataSourceError       = "error"
)

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
