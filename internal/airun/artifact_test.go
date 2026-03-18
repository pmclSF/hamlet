package airun

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestComputeHashes_Deterministic(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, root, "src/prompts.ts", "export const systemPrompt = 'hello';")
	writeFile(t, root, "src/data.ts", "export const dataset = [1, 2, 3];")

	surfaces := []models.CodeSurface{
		{SurfaceID: "surface:src/prompts.ts:systemPrompt", Path: "src/prompts.ts", Kind: models.SurfacePrompt},
		{SurfaceID: "surface:src/data.ts:dataset", Path: "src/data.ts", Kind: models.SurfaceDataset},
	}

	h1 := ComputeHashes(root, surfaces)
	h2 := ComputeHashes(root, surfaces)

	if h1.Prompts["surface:src/prompts.ts:systemPrompt"] != h2.Prompts["surface:src/prompts.ts:systemPrompt"] {
		t.Error("prompt hash not deterministic")
	}
	if h1.Datasets["surface:src/data.ts:dataset"] != h2.Datasets["surface:src/data.ts:dataset"] {
		t.Error("dataset hash not deterministic")
	}
}

func TestComputeHashes_DetectsChange(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, root, "src/prompts.ts", "export const systemPrompt = 'v1';")

	surfaces := []models.CodeSurface{
		{SurfaceID: "s1", Path: "src/prompts.ts", Kind: models.SurfacePrompt},
	}
	h1 := ComputeHashes(root, surfaces)

	// Change file content.
	writeFile(t, root, "src/prompts.ts", "export const systemPrompt = 'v2';")
	h2 := ComputeHashes(root, surfaces)

	if h1.Prompts["s1"] == h2.Prompts["s1"] {
		t.Error("expected different hash after content change")
	}
}

func TestComputeHashes_AllSurfaceKinds(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, root, "src/a.ts", "prompt")
	writeFile(t, root, "src/b.ts", "context")
	writeFile(t, root, "src/c.ts", "dataset")
	writeFile(t, root, "src/d.ts", "tool")
	writeFile(t, root, "src/e.ts", "retrieval")

	surfaces := []models.CodeSurface{
		{SurfaceID: "s1", Path: "src/a.ts", Kind: models.SurfacePrompt},
		{SurfaceID: "s2", Path: "src/b.ts", Kind: models.SurfaceContext},
		{SurfaceID: "s3", Path: "src/c.ts", Kind: models.SurfaceDataset},
		{SurfaceID: "s4", Path: "src/d.ts", Kind: models.SurfaceToolDef},
		{SurfaceID: "s5", Path: "src/e.ts", Kind: models.SurfaceRetrieval},
	}
	h := ComputeHashes(root, surfaces)

	if h.TotalHashCount() != 5 {
		t.Errorf("expected 5 hashes, got %d", h.TotalHashCount())
	}
}

func TestReplay_ExactMatch(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, root, "src/prompts.ts", "export const p = 'hello';")

	surfaces := []models.CodeSurface{
		{SurfaceID: "s1", Path: "src/prompts.ts", Kind: models.SurfacePrompt},
	}

	// Create artifact with current hashes.
	hashes := ComputeHashes(root, surfaces)
	art := &Artifact{
		Version:  "1",
		Mode:     "full",
		Selected: []ScenarioEntry{{ID: "sc1", Name: "test"}},
		Hashes:   hashes,
	}

	artPath := filepath.Join(t.TempDir(), "artifact.json")
	data, _ := json.MarshalIndent(art, "", "  ")
	os.WriteFile(artPath, data, 0o644)

	// Replay against same state.
	result, err := Replay(artPath, root, surfaces, 1)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if !result.Match {
		t.Errorf("expected match, got mismatches: %+v", result.Mismatches)
	}
}

func TestReplay_DetectsContentChange(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, root, "src/prompts.ts", "v1")

	surfaces := []models.CodeSurface{
		{SurfaceID: "s1", Path: "src/prompts.ts", Kind: models.SurfacePrompt},
	}

	hashes := ComputeHashes(root, surfaces)
	art := &Artifact{
		Version:  "1",
		Selected: []ScenarioEntry{{ID: "sc1", Name: "test"}},
		Hashes:   hashes,
	}

	artPath := filepath.Join(t.TempDir(), "artifact.json")
	data, _ := json.MarshalIndent(art, "", "  ")
	os.WriteFile(artPath, data, 0o644)

	// Change content.
	writeFile(t, root, "src/prompts.ts", "v2")

	result, err := Replay(artPath, root, surfaces, 1)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if result.Match {
		t.Error("expected mismatch after content change")
	}
	if len(result.Mismatches) != 1 {
		t.Fatalf("expected 1 mismatch, got %d", len(result.Mismatches))
	}
	if result.Mismatches[0].Kind != "hash" {
		t.Errorf("expected hash mismatch, got %s", result.Mismatches[0].Kind)
	}
}

func TestReplay_DetectsScenarioCountChange(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	art := &Artifact{
		Version:  "1",
		Selected: []ScenarioEntry{{ID: "sc1"}, {ID: "sc2"}},
		Hashes:   ContentHashes{Prompts: map[string]string{}, Contexts: map[string]string{}, Datasets: map[string]string{}, ToolDefs: map[string]string{}, Retrievals: map[string]string{}},
	}
	artPath := filepath.Join(t.TempDir(), "artifact.json")
	data, _ := json.MarshalIndent(art, "", "  ")
	os.WriteFile(artPath, data, 0o644)

	result, err := Replay(artPath, root, nil, 3) // 3 current vs 2 original
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if result.Match {
		t.Error("expected mismatch for scenario count change")
	}
	found := false
	for _, m := range result.Mismatches {
		if m.Kind == "scenario" {
			found = true
		}
	}
	if !found {
		t.Error("expected scenario mismatch kind")
	}
}

func TestSaveArtifact_WritesFile(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	art := &Artifact{
		Mode:     "full",
		Selected: []ScenarioEntry{{ID: "sc1", Name: "test"}},
		Decision: Decision{Action: "pass"},
	}

	path, err := SaveArtifact(root, art)
	if err != nil {
		t.Fatalf("save: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	var loaded Artifact
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if loaded.Version != "1" {
		t.Errorf("version = %q, want 1", loaded.Version)
	}
	if loaded.CreatedAt == "" {
		t.Error("expected non-empty createdAt")
	}
	if len(loaded.Selected) != 1 {
		t.Errorf("selected = %d, want 1", len(loaded.Selected))
	}
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	abs := filepath.Join(root, rel)
	os.MkdirAll(filepath.Dir(abs), 0o755)
	os.WriteFile(abs, []byte(content), 0o644)
}
