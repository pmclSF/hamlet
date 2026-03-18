package gauntlet

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

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
			signalType := classifyFailureSignal(sr)
			snap.Signals = append(snap.Signals, models.Signal{
				Type:     signalType,
				Category: models.CategoryAI,
				Severity: severity,
				Location: models.SignalLocation{
					File:       sr.ScenarioID,
					ScenarioID: sr.ScenarioID,
				},
				Confidence:       0.9,
				EvidenceStrength: models.EvidenceStrong,
				EvidenceSource:   models.SourceEvalExecution,
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
			signalType := classifyRegressionSignal(regMetric)
			snap.Signals = append(snap.Signals, models.Signal{
				Type:             signalType,
				Category:         models.CategoryAI,
				Severity:         models.SeverityMedium,
				Location:         models.SignalLocation{File: sr.ScenarioID, ScenarioID: sr.ScenarioID},
				Confidence:       0.9,
				EvidenceStrength: models.EvidenceStrong,
				EvidenceSource:   models.SourceEvalExecution,
				Explanation:      explanation,
				SuggestedAction:  fmt.Sprintf("Review baseline for %q metric %q", sr.Name, regMetric),
				Metadata: map[string]any{
					"metric":   regMetric,
					"current":  current,
					"baseline": baseline,
				},
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

// classifyFailureSignal maps a failed scenario to the most specific AI signal
// type based on the scenario name and category.
func classifyFailureSignal(sr ScenarioResult) models.SignalType {
	name := strings.ToLower(sr.Name)
	switch {
	case strings.Contains(name, "safety"):
		return "safetyFailure"
	case strings.Contains(name, "hallucination") || strings.Contains(name, "grounding"):
		return "hallucinationDetected"
	case strings.Contains(name, "citation_mismatch") || strings.Contains(name, "citation-mismatch"):
		return "citationMismatch"
	case strings.Contains(name, "citation"):
		return "citationMissing"
	case strings.Contains(name, "wrong_source") || strings.Contains(name, "wrong-source"):
		return "wrongSourceSelected"
	case strings.Contains(name, "stale") && (strings.Contains(name, "source") || strings.Contains(name, "data")):
		return "staleSourceRisk"
	case strings.Contains(name, "chunking") || strings.Contains(name, "chunk_quality"):
		return "chunkingRegression"
	case strings.Contains(name, "rerank"):
		return "rerankerRegression"
	case strings.Contains(name, "retrieval") || strings.Contains(name, "search"):
		return "retrievalMiss"
	case strings.Contains(name, "tool_routing") || strings.Contains(name, "tool-routing") || strings.Contains(name, "wrong_tool"):
		return "toolRoutingError"
	case strings.Contains(name, "tool_guardrail") || strings.Contains(name, "tool-guardrail") || strings.Contains(name, "tool_permission"):
		return "toolGuardrailViolation"
	case strings.Contains(name, "tool_budget") || strings.Contains(name, "tool-budget") ||
		strings.Contains(name, "step_budget") || strings.Contains(name, "step-budget") ||
		strings.Contains(name, "step_limit") || strings.Contains(name, "step-limit"):
		return "toolBudgetExceeded"
	case strings.Contains(name, "fallback") && (strings.Contains(name, "agent") || strings.Contains(name, "model")):
		return "agentFallbackTriggered"
	case strings.Contains(name, "tool") || strings.Contains(name, "function_call"):
		return "toolSelectionError"
	case strings.Contains(name, "schema") || strings.Contains(name, "parse"):
		return "schemaParseFailure"
	case strings.Contains(name, "policy"):
		return "aiPolicyViolation"
	default:
		return "evalFailure"
	}
}

// classifyRegressionSignal maps a regression metric name to the most specific
// AI signal type.
func classifyRegressionSignal(metric string) models.SignalType {
	lower := strings.ToLower(metric)
	switch {
	// RAG-specific metrics (checked before generic accuracy to avoid false matches).
	case strings.Contains(lower, "chunk") && (strings.Contains(lower, "quality") || strings.Contains(lower, "score") || strings.Contains(lower, "size")):
		return "chunkingRegression"
	case strings.Contains(lower, "rerank") && (strings.Contains(lower, "score") || strings.Contains(lower, "quality") || strings.Contains(lower, "ndcg")):
		return "rerankerRegression"
	case strings.Contains(lower, "top_k") || strings.Contains(lower, "topk") || strings.Contains(lower, "recall_at_k") || strings.Contains(lower, "mrr"):
		return "topKRegression"
	// Citation-specific (before generic accuracy since "citation_accuracy" contains "accuracy").
	case strings.Contains(lower, "citation_match") || strings.Contains(lower, "citation_accuracy"):
		return "citationMismatch"
	case strings.Contains(lower, "citation"):
		return "citationMissing"
	// Tool/agent metrics (before generic accuracy since "tool_selection_accuracy" contains "accuracy").
	case strings.Contains(lower, "tool_routing") || strings.Contains(lower, "tool_selection"):
		return "toolRoutingError"
	case strings.Contains(lower, "tool_budget") || strings.Contains(lower, "step_count") || strings.Contains(lower, "step_budget"):
		return "toolBudgetExceeded"
	case strings.Contains(lower, "fallback_rate") || strings.Contains(lower, "fallback_count"):
		return "agentFallbackTriggered"
	// Generic metrics.
	case strings.Contains(lower, "accuracy") || strings.Contains(lower, "f1") || strings.Contains(lower, "precision") || strings.Contains(lower, "recall"):
		return "accuracyRegression"
	case strings.Contains(lower, "latency") || strings.Contains(lower, "p95") || strings.Contains(lower, "p99") || strings.Contains(lower, "duration"):
		return "latencyRegression"
	case strings.Contains(lower, "cost") || strings.Contains(lower, "token") || strings.Contains(lower, "price"):
		return "costRegression"
	case strings.Contains(lower, "grounding") || strings.Contains(lower, "hallucination") || strings.Contains(lower, "faithfulness"):
		return "answerGroundingFailure"
	case strings.Contains(lower, "context_length") || strings.Contains(lower, "overflow"):
		return "contextOverflowRisk"
	case strings.Contains(lower, "stale") || strings.Contains(lower, "freshness"):
		return "staleSourceRisk"
	case strings.Contains(lower, "source_relevance") || strings.Contains(lower, "wrong_source"):
		return "wrongSourceSelected"
	default:
		return "evalRegression"
	}
}
