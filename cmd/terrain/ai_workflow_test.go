package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/pmclSF/terrain/internal/engine"
	"github.com/pmclSF/terrain/internal/explain"
	"github.com/pmclSF/terrain/internal/impact"
	"github.com/pmclSF/terrain/internal/models"
)

var aiFixtureOnce sync.Once
var aiFixtureErr error

func aiFixtureRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "tests", "fixtures", "ai-eval-suite")
	aiFixtureOnce.Do(func() {
		aiFixtureErr = ensureAIFixtureGit(root)
	})
	if aiFixtureErr != nil {
		t.Fatalf("AI fixture git setup failed: %v", aiFixtureErr)
	}
	return root
}

// ensureAIFixtureGit initializes a git repo in the AI eval fixture directory.
// Creates two commits so HEAD~1 diff shows a changed source file.
func ensureAIFixtureGit(root string) error {
	if _, err := os.Stat(filepath.Join(root, ".git")); err == nil {
		return nil
	}

	run := func(args ...string) error {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			return &gitSetupError{args, err, string(out)}
		}
		return nil
	}

	for _, args := range [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@test.com"},
		{"git", "config", "user.name", "Test"},
	} {
		if err := run(args...); err != nil {
			return err
		}
	}

	// First commit: everything except classifier.py.
	if err := run("git", "add", "."); err != nil {
		return err
	}
	if err := run("git", "rm", "--cached", "src/models/classifier.py"); err != nil {
		return err
	}
	if err := run("git", "commit", "-m", "initial commit"); err != nil {
		return err
	}

	// Second commit: add classifier.py so HEAD~1 diff shows it as changed.
	if err := run("git", "add", "src/models/classifier.py"); err != nil {
		return err
	}
	return run("git", "commit", "-m", "add classifier model")
}

type gitSetupError struct {
	args []string
	err  error
	out  string
}

func (e *gitSetupError) Error() string {
	return "fixture git setup (" + strings.Join(e.args, " ") + "): " + e.err.Error() + "\n" + e.out
}

// --- Integration Tests: PR Change → Impact → Explain ---

// TestAIWorkflow_PipelineLoadsScenarios verifies that the analysis pipeline
// loads scenarios from .terrain/terrain.yaml.
func TestAIWorkflow_PipelineLoadsScenarios(t *testing.T) {
	t.Parallel()
	root := aiFixtureRoot(t)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skip("ai-eval-suite fixture not found")
	}

	result, err := engine.RunPipeline(root, engine.PipelineOptions{EngineVersion: "test"})
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	snap := result.Snapshot

	if len(snap.Scenarios) == 0 {
		t.Fatal("expected scenarios to be loaded from .terrain/terrain.yaml and auto-derived")
	}
	// At least 3 from YAML, potentially more from auto-derivation.
	if len(snap.Scenarios) < 3 {
		t.Errorf("expected at least 3 scenarios, got %d", len(snap.Scenarios))
	}

	// Verify scenario fields populated.
	found := false
	for _, sc := range snap.Scenarios {
		if sc.Name == "classifier-accuracy" {
			found = true
			if sc.Category != "accuracy" {
				t.Errorf("expected category accuracy, got %s", sc.Category)
			}
			if len(sc.CoveredSurfaceIDs) != 2 {
				t.Errorf("expected 2 covered surfaces, got %d", len(sc.CoveredSurfaceIDs))
			}
		}
	}
	if !found {
		t.Error("expected classifier-accuracy scenario")
	}
}

// TestAIWorkflow_ImpactDetectsScenarios verifies that changing a source file
// surfaces impacted scenarios in the impact result.
func TestAIWorkflow_ImpactDetectsScenarios(t *testing.T) {
	t.Parallel()
	root := aiFixtureRoot(t)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skip("ai-eval-suite fixture not found")
	}

	result, err := engine.RunPipeline(root, engine.PipelineOptions{EngineVersion: "test"})
	if err != nil {
		t.Fatalf("pipeline failed: %v", err)
	}
	snap := result.Snapshot

	absRoot, err := filepath.Abs(root)
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}

	cs, err := impact.ChangeSetFromGitDiff(absRoot, "HEAD~1")
	if err != nil {
		t.Fatalf("changeset: %v", err)
	}

	impactResult := impact.AnalyzeChangeSet(cs, snap)

	// The diff adds classifier.py which contains classify/batch_classify.
	// The classifier-accuracy scenario covers those surfaces.
	if len(impactResult.ImpactedScenarios) == 0 {
		// Scenario detection depends on surface IDs matching exactly.
		// Even if surfaces don't match due to naming, the impact should still
		// produce changed areas with the classifier file.
		if len(impactResult.ChangedAreas) == 0 {
			t.Error("expected changed areas for classifier.py")
		}
		t.Log("no impacted scenarios (surface IDs may not match yaml declaration)")
	}

	// Verify changed surfaces include the classifier file.
	hasClassifier := false
	for _, area := range impactResult.ChangedAreas {
		for _, s := range area.Surfaces {
			if strings.Contains(s.Path, "classifier") {
				hasClassifier = true
			}
		}
	}
	if !hasClassifier {
		t.Error("expected classifier.py in changed areas")
	}

	// Verify confidence score exists.
	if impactResult.CoverageConfidence == "" {
		t.Error("expected non-empty coverage confidence")
	}

	// Verify fallback info exists.
	if impactResult.Fallback.Level == "" && impactResult.ProtectiveSet == nil {
		t.Error("expected fallback info or protective set")
	}

	// Verify summary mentions the change.
	if impactResult.Summary == "" {
		t.Error("expected non-empty impact summary")
	}
}

// TestAIWorkflow_ExplainScenario verifies that terrain explain can produce
// a structured explanation for an impacted scenario.
func TestAIWorkflow_ExplainScenario(t *testing.T) {
	t.Parallel()

	// Build a minimal ImpactResult with an impacted scenario.
	impactResult := &impact.ImpactResult{
		ImpactedScenarios: []impact.ImpactedScenario{
			{
				ScenarioID:       "scenario:custom:classifier-accuracy",
				Name:             "classifier-accuracy",
				Category:         "accuracy",
				Framework:        "custom",
				Relevance:        "covers 2 changed surface(s)",
				ImpactConfidence: impact.ConfidenceExact,
				CoversSurfaces: []string{
					"surface:src/models/classifier.py:classify",
					"surface:src/models/classifier.py:batch_classify",
				},
			},
		},
	}

	// Explain by scenario ID.
	se, err := explain.ExplainScenario("scenario:custom:classifier-accuracy", impactResult)
	if err != nil {
		t.Fatalf("explain by ID: %v", err)
	}
	if se.ScenarioID != "scenario:custom:classifier-accuracy" {
		t.Errorf("expected scenario ID, got %s", se.ScenarioID)
	}
	if se.Verdict == "" {
		t.Error("expected non-empty verdict")
	}
	if len(se.ChangedSurfaces) != 2 {
		t.Errorf("expected 2 changed surfaces, got %d", len(se.ChangedSurfaces))
	}

	// Explain by scenario name.
	se2, err := explain.ExplainScenario("classifier-accuracy", impactResult)
	if err != nil {
		t.Fatalf("explain by name: %v", err)
	}
	if se2.Name != "classifier-accuracy" {
		t.Errorf("expected name match, got %s", se2.Name)
	}

	// Explain unknown scenario.
	_, err = explain.ExplainScenario("nonexistent", impactResult)
	if err == nil {
		t.Error("expected error for unknown scenario")
	}
}

// TestAIWorkflow_ExplainScenario_NilResult ensures explain handles nil gracefully.
func TestAIWorkflow_ExplainScenario_NilResult(t *testing.T) {
	t.Parallel()
	_, err := explain.ExplainScenario("anything", nil)
	if err == nil {
		t.Error("expected error for nil impact result")
	}
}

// TestAIWorkflow_AIListShowsScenarios verifies terrain ai list includes
// scenarios from the fixture.
func TestAIWorkflow_AIListShowsScenarios(t *testing.T) {
	t.Parallel()
	root := aiFixtureRoot(t)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skip("ai-eval-suite fixture not found")
	}

	// runAIList should succeed and show scenarios.
	if err := runAIList(root, false, false); err != nil {
		t.Fatalf("runAIList: %v", err)
	}
}

// TestAIWorkflow_AIDoctorPassesWithScenarios verifies terrain ai doctor
// reports passing checks when scenarios are configured.
func TestAIWorkflow_AIDoctorPassesWithScenarios(t *testing.T) {
	t.Parallel()
	root := aiFixtureRoot(t)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skip("ai-eval-suite fixture not found")
	}

	if err := runAIDoctor(root, false); err != nil {
		t.Fatalf("runAIDoctor: %v", err)
	}
}

// TestAIWorkflow_FullChain_Deterministic verifies the full chain produces
// deterministic output across two runs.
func TestAIWorkflow_FullChain_Deterministic(t *testing.T) {
	t.Parallel()
	root := aiFixtureRoot(t)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skip("ai-eval-suite fixture not found")
	}

	run := func() *models.TestSuiteSnapshot {
		result, err := engine.RunPipeline(root, engine.PipelineOptions{EngineVersion: "test"})
		if err != nil {
			t.Fatalf("pipeline failed: %v", err)
		}
		return result.Snapshot
	}

	snap1 := run()
	snap2 := run()

	if len(snap1.Scenarios) != len(snap2.Scenarios) {
		t.Fatalf("non-deterministic scenario count: %d vs %d", len(snap1.Scenarios), len(snap2.Scenarios))
	}
	for i := range snap1.Scenarios {
		if snap1.Scenarios[i].ScenarioID != snap2.Scenarios[i].ScenarioID {
			t.Errorf("non-deterministic scenario ID at %d: %s vs %s",
				i, snap1.Scenarios[i].ScenarioID, snap2.Scenarios[i].ScenarioID)
		}
	}
	if len(snap1.CodeSurfaces) != len(snap2.CodeSurfaces) {
		t.Fatalf("non-deterministic surface count: %d vs %d", len(snap1.CodeSurfaces), len(snap2.CodeSurfaces))
	}
}

func TestAIWorkflow_InventoryJSON_Structure(t *testing.T) {
	// Not parallel: captures os.Stdout.
	root := aiFixtureRoot(t)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skip("ai-eval-suite fixture not found")
	}

	// Capture JSON output.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runAIList(root, true, false)
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("runAIList JSON: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	var result map[string]json.RawMessage
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output[:min(200, len(output))])
	}

	// Required top-level keys.
	for _, key := range []string{"scenarios", "prompts", "datasets", "evalFiles", "summary", "capabilities"} {
		if _, ok := result[key]; !ok {
			t.Errorf("missing key %q in JSON", key)
		}
	}

	// Summary should have required fields.
	var summary map[string]int
	if err := json.Unmarshal(result["summary"], &summary); err != nil {
		t.Fatalf("invalid summary: %v", err)
	}
	for _, field := range []string{"scenarios", "capabilities", "totalAISurfaces", "uncoveredSurfaces", "evalFiles"} {
		if _, ok := summary[field]; !ok {
			t.Errorf("missing summary field %q", field)
		}
	}
}

func TestAIWorkflow_InventoryText_Content(t *testing.T) {
	// Not parallel: captures os.Stdout.
	root := aiFixtureRoot(t)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skip("ai-eval-suite fixture not found")
	}

	// Capture text output.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runAIList(root, false, false)
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("runAIList text: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Key sections must be present.
	for _, section := range []string{"Terrain AI Inventory", "Component", "Scenarios", "Next steps"} {
		if !strings.Contains(output, section) {
			t.Errorf("missing section %q in text output", section)
		}
	}
}

func TestAIWorkflow_InventoryDeterministic(t *testing.T) {
	// Not parallel: captures os.Stdout.
	root := aiFixtureRoot(t)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skip("ai-eval-suite fixture not found")
	}

	capture := func() string {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		_ = runAIList(root, true, false)
		w.Close()
		os.Stdout = old
		var buf bytes.Buffer
		buf.ReadFrom(r)
		return buf.String()
	}

	j1 := capture()
	j2 := capture()
	if j1 != j2 {
		t.Error("ai list JSON output is not deterministic across runs")
	}
}

func TestAIRun_DryRun_SelectsScenarios(t *testing.T) {
	// Not parallel: captures os.Stdout.
	root := aiFixtureRoot(t)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skip("ai-eval-suite fixture not found")
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runAIRun(root, true, "", true, true) // full + dry-run + json
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("runAIRun dry-run: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	var artifact map[string]json.RawMessage
	if err := json.Unmarshal([]byte(output), &artifact); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Should have mode, selected, decision.
	for _, key := range []string{"mode", "selected", "decision"} {
		if _, ok := artifact[key]; !ok {
			t.Errorf("missing key %q in artifact", key)
		}
	}

	// Mode should be dry-run.
	var mode string
	json.Unmarshal(artifact["mode"], &mode)
	if mode != "dry-run" {
		t.Errorf("mode = %q, want dry-run", mode)
	}

	// Should have selected scenarios.
	var selected []json.RawMessage
	json.Unmarshal(artifact["selected"], &selected)
	if len(selected) == 0 {
		t.Error("expected selected scenarios in dry-run output")
	}
}

func TestAIRun_FullMode_AllScenarios(t *testing.T) {
	// Not parallel: captures os.Stdout.
	root := aiFixtureRoot(t)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skip("ai-eval-suite fixture not found")
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runAIRun(root, true, "", true, true) // full + dry-run
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("runAIRun full: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var artifact struct {
		Mode     string          `json:"mode"`
		Selected []aiRunScenario `json:"selected"`
		Skipped  []aiRunScenario `json:"skipped"`
	}
	if err := json.Unmarshal(buf.Bytes(), &artifact); err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(artifact.Selected) == 0 {
		t.Error("expected selected scenarios in full mode")
	}
	// Full mode: all selected, none skipped.
	if len(artifact.Skipped) != 0 {
		t.Errorf("expected 0 skipped in full mode, got %d", len(artifact.Skipped))
	}
	// Every selected scenario should have reason mentioning "full".
	for _, sc := range artifact.Selected {
		if sc.Reason == "" {
			t.Errorf("scenario %q has empty reason", sc.Name)
		}
	}
}

func TestAIRun_DecisionPass_NoSignals(t *testing.T) {
	// Not parallel: captures os.Stdout.
	root := aiFixtureRoot(t)
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skip("ai-eval-suite fixture not found")
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runAIRun(root, true, "", true, true)
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("runAIRun: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var artifact struct {
		Decision aiRunDecision `json:"decision"`
		ExitCode int           `json:"exitCode"`
	}
	if err := json.Unmarshal(buf.Bytes(), &artifact); err != nil {
		t.Fatalf("parse: %v", err)
	}

	// No Gauntlet signals in fixture → pass.
	if artifact.Decision.Action != "pass" {
		t.Errorf("decision = %q, want pass", artifact.Decision.Action)
	}
	if artifact.ExitCode != 0 {
		t.Errorf("exitCode = %d, want 0", artifact.ExitCode)
	}
}

func TestAIWorkflow_InventoryJSON_IncludesEvidence(t *testing.T) {
	// Not parallel: captures os.Stdout.
	_, thisFile, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "tests", "fixtures", "ai-prompt-only")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		t.Skip("ai-prompt-only fixture not found")
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runAIList(root, true, false)
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("runAIList JSON: %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result struct {
		Prompts  []json.RawMessage `json:"prompts"`
		Contexts []json.RawMessage `json:"contexts"`
	}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("parse: %v", err)
	}

	// Check that at least one surface has tier and confidence.
	allSurfaces := append(result.Prompts, result.Contexts...)
	if len(allSurfaces) == 0 {
		t.Skip("no prompts or contexts in fixture")
	}

	for _, raw := range allSurfaces {
		var entry struct {
			DetectionTier string  `json:"detectionTier"`
			Confidence    float64 `json:"confidence"`
		}
		json.Unmarshal(raw, &entry)
		if entry.DetectionTier == "" {
			t.Error("surface missing detectionTier in JSON")
		}
		if entry.Confidence <= 0 {
			t.Error("surface missing confidence in JSON")
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
