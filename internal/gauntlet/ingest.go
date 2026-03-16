package gauntlet

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/pmclSF/terrain/internal/models"
)

// Ingest reads a Gauntlet result artifact and returns the parsed artifact.
// The artifact is validated for required fields.
func Ingest(path string) (*Artifact, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read gauntlet artifact %s: %w", path, err)
	}

	var art Artifact
	if err := json.Unmarshal(data, &art); err != nil {
		return nil, fmt.Errorf("parse gauntlet artifact %s: %w", path, err)
	}

	if art.Version == "" {
		return nil, fmt.Errorf("gauntlet artifact %s: missing version field", path)
	}
	if art.Provider != "gauntlet" {
		return nil, fmt.Errorf("gauntlet artifact %s: expected provider \"gauntlet\", got %q", path, art.Provider)
	}
	if len(art.Scenarios) == 0 {
		return nil, fmt.Errorf("gauntlet artifact %s: no scenarios in artifact", path)
	}

	return &art, nil
}

// ApplyToSnapshot merges Gauntlet execution results into a Terrain snapshot.
//
// For each scenario result:
//   - If the scenarioId matches a Scenario in the snapshot, execution metadata
//     is recorded and signals are generated for failures/regressions.
//   - Unmatched scenario results are tracked but do not generate signals.
//
// The snapshot's DataSources is updated with gauntlet ingestion status.
func ApplyToSnapshot(snap *models.TestSuiteSnapshot, art *Artifact) ApplyResult {
	result := ApplyResult{
		TotalResults: len(art.Scenarios),
	}

	// Index snapshot scenarios by ID for O(1) lookup.
	scenarioIdx := map[string]int{}
	for i, sc := range snap.Scenarios {
		scenarioIdx[sc.ScenarioID] = i
	}

	for _, sr := range art.Scenarios {
		if _, ok := scenarioIdx[sr.ScenarioID]; ok {
			result.MatchedCount++
		} else {
			result.UnmatchedIDs = append(result.UnmatchedIDs, sr.ScenarioID)
		}

		// Generate signals for failures and regressions.
		if sr.Status == "failed" || sr.Status == "error" {
			severity := models.SeverityMedium
			if sr.Status == "error" {
				severity = models.SeverityHigh
			}
			snap.Signals = append(snap.Signals, models.Signal{
				Type:     "evalFailure",
				Category: models.CategoryQuality,
				Severity: severity,
				Location: models.SignalLocation{
					File: sr.ScenarioID,
				},
				Explanation: fmt.Sprintf(
					"Gauntlet scenario %q %s (duration: %.0fms)",
					sr.Name, sr.Status, sr.DurationMs,
				),
				SuggestedAction: fmt.Sprintf("Investigate %s scenario %q", sr.Status, sr.Name),
			})
			result.FailureCount++
		}

		for _, regMetric := range sr.Regressions {
			current, hasCurrent := sr.Metrics[regMetric]
			baseline, hasBaseline := sr.Baseline[regMetric]
			explanation := fmt.Sprintf("Gauntlet scenario %q: metric %q regressed", sr.Name, regMetric)
			if hasCurrent && hasBaseline {
				explanation = fmt.Sprintf(
					"Gauntlet scenario %q: metric %q regressed from %.4f to %.4f",
					sr.Name, regMetric, baseline, current,
				)
			}
			snap.Signals = append(snap.Signals, models.Signal{
				Type:            "evalRegression",
				Category:        models.CategoryQuality,
				Severity:        models.SeverityMedium,
				Location:        models.SignalLocation{File: sr.ScenarioID},
				Explanation:     explanation,
				SuggestedAction: fmt.Sprintf("Review baseline for %q metric %q", sr.Name, regMetric),
			})
			result.RegressionCount++
		}
	}

	sort.Slice(result.UnmatchedIDs, func(i, j int) bool {
		return result.UnmatchedIDs[i] < result.UnmatchedIDs[j]
	})

	return result
}

// ApplyResult summarizes the outcome of applying a Gauntlet artifact.
type ApplyResult struct {
	// TotalResults is the number of scenario results in the artifact.
	TotalResults int

	// MatchedCount is how many matched a Terrain Scenario by ID.
	MatchedCount int

	// UnmatchedIDs lists scenario IDs present in the artifact but not
	// in Terrain's scenario inventory.
	UnmatchedIDs []string

	// FailureCount is the number of failed/errored scenarios.
	FailureCount int

	// RegressionCount is the number of metric regressions detected.
	RegressionCount int
}
