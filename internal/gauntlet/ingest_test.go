package gauntlet

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestIngest_ValidArtifact(t *testing.T) {
	t.Parallel()
	path := writeArtifact(t, `{
		"version": "1",
		"provider": "gauntlet",
		"timestamp": "2026-03-15T12:00:00Z",
		"scenarios": [
			{"scenarioId": "eval:safety", "name": "safety-check", "status": "passed", "durationMs": 1200},
			{"scenarioId": "eval:accuracy", "name": "accuracy-check", "status": "failed", "durationMs": 3000}
		],
		"summary": {"total": 2, "passed": 1, "failed": 1}
	}`)

	art, err := Ingest(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if art.Version != "1" {
		t.Errorf("version = %q, want 1", art.Version)
	}
	if len(art.Scenarios) != 2 {
		t.Errorf("scenarios = %d, want 2", len(art.Scenarios))
	}
	if art.Summary.Total != 2 {
		t.Errorf("summary.total = %d, want 2", art.Summary.Total)
	}
}

func TestIngest_MissingVersion(t *testing.T) {
	t.Parallel()
	path := writeArtifact(t, `{"provider": "gauntlet", "scenarios": [{"scenarioId": "x", "name": "x", "status": "passed"}]}`)

	_, err := Ingest(path)
	if err == nil {
		t.Fatal("expected error for missing version")
	}
}

func TestIngest_WrongProvider(t *testing.T) {
	t.Parallel()
	path := writeArtifact(t, `{"version": "1", "provider": "other", "scenarios": [{"scenarioId": "x", "name": "x", "status": "passed"}]}`)

	_, err := Ingest(path)
	if err == nil {
		t.Fatal("expected error for wrong provider")
	}
}

func TestIngest_EmptyScenarios(t *testing.T) {
	t.Parallel()
	path := writeArtifact(t, `{"version": "1", "provider": "gauntlet", "scenarios": []}`)

	_, err := Ingest(path)
	if err == nil {
		t.Fatal("expected error for empty scenarios")
	}
}

func TestIngest_FileNotFound(t *testing.T) {
	t.Parallel()
	_, err := Ingest(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestIngest_InvalidJSON(t *testing.T) {
	t.Parallel()
	path := writeArtifact(t, `{invalid json}`)

	_, err := Ingest(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestApplyToSnapshot_MatchesScenarios(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{
		Scenarios: []models.Scenario{
			{ScenarioID: "eval:safety", Name: "safety-check"},
			{ScenarioID: "eval:accuracy", Name: "accuracy-check"},
		},
	}

	art := &Artifact{
		Scenarios: []ScenarioResult{
			{ScenarioID: "eval:safety", Name: "safety-check", Status: "passed"},
			{ScenarioID: "eval:accuracy", Name: "accuracy-check", Status: "failed", DurationMs: 3000},
			{ScenarioID: "eval:unknown", Name: "unknown", Status: "passed"},
		},
	}

	result := ApplyToSnapshot(snap, art)

	if result.TotalResults != 3 {
		t.Errorf("total = %d, want 3", result.TotalResults)
	}
	if result.MatchedCount != 2 {
		t.Errorf("matched = %d, want 2", result.MatchedCount)
	}
	if len(result.UnmatchedIDs) != 1 || result.UnmatchedIDs[0] != "eval:unknown" {
		t.Errorf("unmatched = %v, want [eval:unknown]", result.UnmatchedIDs)
	}
	if result.FailureCount != 1 {
		t.Errorf("failures = %d, want 1", result.FailureCount)
	}
}

func TestApplyToSnapshot_GeneratesSignals(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{}
	art := &Artifact{
		Scenarios: []ScenarioResult{
			{ScenarioID: "eval:safety", Name: "safety-check", Status: "failed", DurationMs: 1200},
			{ScenarioID: "eval:infra", Name: "infra-check", Status: "error", DurationMs: 500},
		},
	}

	ApplyToSnapshot(snap, art)

	if len(snap.Signals) != 2 {
		t.Fatalf("expected 2 signals, got %d", len(snap.Signals))
	}

	// Failed scenario should generate medium severity.
	if snap.Signals[0].Severity != models.SeverityMedium {
		t.Errorf("failed scenario severity = %s, want medium", snap.Signals[0].Severity)
	}
	// "safety-check" maps to safetyFailure via classifier.
	if snap.Signals[0].Type != "safetyFailure" {
		t.Errorf("signal type = %s, want safetyFailure", snap.Signals[0].Type)
	}

	// Error scenario should generate high severity.
	if snap.Signals[1].Severity != models.SeverityHigh {
		t.Errorf("error scenario severity = %s, want high", snap.Signals[1].Severity)
	}
}

func TestApplyToSnapshot_Regressions(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{}
	art := &Artifact{
		Scenarios: []ScenarioResult{
			{
				ScenarioID:  "eval:accuracy",
				Name:        "accuracy-check",
				Status:      "passed",
				Metrics:     map[string]float64{"accuracy": 0.85},
				Baseline:    map[string]float64{"accuracy": 0.92},
				Regressions: []string{"accuracy"},
			},
		},
	}

	result := ApplyToSnapshot(snap, art)

	if result.RegressionCount != 1 {
		t.Errorf("regressions = %d, want 1", result.RegressionCount)
	}
	// Should have 1 regression signal (no failure signal since status is passed).
	if len(snap.Signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(snap.Signals))
	}
	// "accuracy" metric maps to accuracyRegression via classifier.
	if snap.Signals[0].Type != "accuracyRegression" {
		t.Errorf("signal type = %s, want accuracyRegression", snap.Signals[0].Type)
	}
}

func TestApplyToSnapshot_EmptySnapshot(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{}
	art := &Artifact{
		Scenarios: []ScenarioResult{
			{ScenarioID: "eval:safety", Name: "safety-check", Status: "passed"},
		},
	}

	result := ApplyToSnapshot(snap, art)

	// No matching scenarios — all unmatched.
	if result.MatchedCount != 0 {
		t.Errorf("matched = %d, want 0", result.MatchedCount)
	}
	if len(result.UnmatchedIDs) != 1 {
		t.Errorf("unmatched = %d, want 1", len(result.UnmatchedIDs))
	}
	// Passed scenario generates no signal.
	if len(snap.Signals) != 0 {
		t.Errorf("expected 0 signals for passed scenario, got %d", len(snap.Signals))
	}
}

func TestClassifyFailureSignal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want models.SignalType
	}{
		{"safety-check", "safetyFailure"},
		{"hallucination-detection", "hallucinationDetected"},
		{"grounding-eval", "hallucinationDetected"},
		{"citation-completeness", "citationMissing"},
		{"retrieval-quality", "retrievalMiss"},
		{"search-relevance", "retrievalMiss"},
		{"tool-selection-accuracy", "toolSelectionError"},
		{"schema-validation", "schemaParseFailure"},
		{"policy-compliance", "aiPolicyViolation"},
		{"citation-mismatch-check", "citationMismatch"},
		{"wrong-source-selection", "wrongSourceSelected"},
		{"stale-source-detection", "staleSourceRisk"},
		{"chunking-quality-check", "chunkingRegression"},
		{"rerank-quality", "rerankerRegression"},
		{"tool-routing-accuracy", "toolRoutingError"},
		{"tool-guardrail-enforcement", "toolGuardrailViolation"},
		{"step-budget-exceeded", "toolBudgetExceeded"},
		{"agent-fallback-triggered", "agentFallbackTriggered"},
		{"generic-scenario", "evalFailure"},
	}
	for _, tt := range tests {
		got := classifyFailureSignal(ScenarioResult{Name: tt.name})
		if got != tt.want {
			t.Errorf("classifyFailureSignal(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestClassifyRegressionSignal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		metric string
		want   models.SignalType
	}{
		{"accuracy", "accuracyRegression"},
		{"f1_score", "accuracyRegression"},
		{"precision", "accuracyRegression"},
		// recall_at_k is now classified as topKRegression (RAG-specific, checked before accuracy).
		{"latency_p95_ms", "latencyRegression"},
		{"p99_duration", "latencyRegression"},
		{"cost_per_query", "costRegression"},
		{"token_usage", "costRegression"},
		{"citation_score", "citationMissing"},
		{"grounding_score", "answerGroundingFailure"},
		{"faithfulness", "answerGroundingFailure"},
		{"context_length", "contextOverflowRisk"},
		{"chunk_quality_score", "chunkingRegression"},
		{"chunk_size_ratio", "chunkingRegression"},
		{"rerank_ndcg", "rerankerRegression"},
		{"rerank_score", "rerankerRegression"},
		{"top_k_recall", "topKRegression"},
		{"mrr", "topKRegression"},
		{"recall_at_k", "topKRegression"},
		{"citation_match_rate", "citationMismatch"},
		{"citation_accuracy", "citationMismatch"},
		{"freshness_score", "staleSourceRisk"},
		{"source_relevance", "wrongSourceSelected"},
		{"tool_routing_accuracy", "toolRoutingError"},
		{"tool_selection_accuracy", "toolRoutingError"},
		{"step_count", "toolBudgetExceeded"},
		{"step_budget_used", "toolBudgetExceeded"},
		{"fallback_rate", "agentFallbackTriggered"},
		{"fallback_count", "agentFallbackTriggered"},
		{"custom_metric", "evalRegression"},
	}
	for _, tt := range tests {
		got := classifyRegressionSignal(tt.metric)
		if got != tt.want {
			t.Errorf("classifyRegressionSignal(%q) = %q, want %q", tt.metric, got, tt.want)
		}
	}
}

func TestIngest_AISignalTypes(t *testing.T) {
	t.Parallel()
	path := writeArtifact(t, `{
		"version": "1",
		"provider": "gauntlet",
		"scenarios": [
			{
				"scenarioId": "eval:safety-check",
				"name": "safety-check",
				"status": "failed",
				"durationMs": 100
			},
			{
				"scenarioId": "eval:accuracy",
				"name": "accuracy-eval",
				"status": "passed",
				"metrics": {"accuracy": 0.85},
				"baseline": {"accuracy": 0.93},
				"regressions": ["accuracy"]
			}
		]
	}`)

	snap := &models.TestSuiteSnapshot{
		Scenarios: []models.Scenario{
			{ScenarioID: "eval:safety-check"},
			{ScenarioID: "eval:accuracy"},
		},
	}

	art, err := Ingest(path)
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	result := ApplyToSnapshot(snap, art)

	if result.FailureCount != 1 {
		t.Errorf("failures = %d, want 1", result.FailureCount)
	}
	if result.RegressionCount != 1 {
		t.Errorf("regressions = %d, want 1", result.RegressionCount)
	}

	// Check signal types.
	var safetyFound, accuracyFound bool
	for _, sig := range snap.Signals {
		if sig.Type == "safetyFailure" {
			safetyFound = true
			if sig.Category != models.CategoryAI {
				t.Errorf("safetyFailure category = %s, want ai", sig.Category)
			}
			if sig.Location.ScenarioID != "eval:safety-check" {
				t.Errorf("safetyFailure scenarioID = %s", sig.Location.ScenarioID)
			}
		}
		if sig.Type == "accuracyRegression" {
			accuracyFound = true
			if sig.Category != models.CategoryAI {
				t.Errorf("accuracyRegression category = %s, want ai", sig.Category)
			}
		}
	}
	if !safetyFound {
		t.Error("expected safetyFailure signal")
	}
	if !accuracyFound {
		t.Error("expected accuracyRegression signal")
	}
}

func writeArtifact(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "gauntlet-results.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	return path
}
