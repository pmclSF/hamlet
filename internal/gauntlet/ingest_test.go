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
	if snap.Signals[0].Type != "evalFailure" {
		t.Errorf("signal type = %s, want evalFailure", snap.Signals[0].Type)
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
	if snap.Signals[0].Type != "evalRegression" {
		t.Errorf("signal type = %s, want evalRegression", snap.Signals[0].Type)
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

func writeArtifact(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "gauntlet-results.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	return path
}
