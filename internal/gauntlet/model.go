// Package gauntlet provides ingestion for Gauntlet AI eval execution artifacts.
//
// Gauntlet (github.com/pmclsf/gauntlet) is Terrain's first AI execution
// provider. It runs eval scenarios deterministically and produces structured
// JSON result artifacts that Terrain ingests for reasoning, coverage analysis,
// and baseline comparison.
//
// Terrain owns: scenario selection, reasoning, coverage.
// Gauntlet owns: deterministic execution, baseline comparison.
// The boundary is the artifact file.
package gauntlet

// Artifact is the top-level Gauntlet result file.
type Artifact struct {
	// Version is the schema version. Currently "1".
	Version string `json:"version"`

	// Provider identifies the execution provider. Always "gauntlet".
	Provider string `json:"provider"`

	// Timestamp is the ISO 8601 execution timestamp.
	Timestamp string `json:"timestamp"`

	// Repository is the repository name, if provided by Gauntlet.
	Repository string `json:"repository,omitempty"`

	// Scenarios contains per-scenario execution results.
	Scenarios []ScenarioResult `json:"scenarios"`

	// Summary contains aggregate execution counts.
	Summary Summary `json:"summary"`
}

// ScenarioResult holds execution results for a single eval scenario.
type ScenarioResult struct {
	// ScenarioID matches Terrain's Scenario.ScenarioID for joining.
	ScenarioID string `json:"scenarioId"`

	// Name is the human-readable scenario name.
	Name string `json:"name"`

	// Status is the execution outcome: "passed", "failed", "skipped", "error".
	Status string `json:"status"`

	// DurationMs is the execution time in milliseconds.
	DurationMs float64 `json:"durationMs,omitempty"`

	// Metrics holds key-value metric results (accuracy, latency, etc.).
	Metrics map[string]float64 `json:"metrics,omitempty"`

	// ModelVersion is the model version under evaluation.
	ModelVersion string `json:"modelVersion,omitempty"`

	// Baseline holds previous baseline metric values for comparison.
	Baseline map[string]float64 `json:"baseline,omitempty"`

	// Regressions lists metric names that regressed vs the baseline.
	Regressions []string `json:"regressions,omitempty"`
}

// Summary holds aggregate execution counts.
type Summary struct {
	Total      int     `json:"total"`
	Passed     int     `json:"passed"`
	Failed     int     `json:"failed"`
	Skipped    int     `json:"skipped"`
	DurationMs float64 `json:"durationMs,omitempty"`
}
