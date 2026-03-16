package aidetect

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestDetect_EmptyDir(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	result := Detect(root)
	if len(result.Frameworks) != 0 {
		t.Errorf("expected 0 frameworks, got %d", len(result.Frameworks))
	}
}

func TestDetect_PackageJSON_LangChain(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "package.json"), []byte(`{
		"dependencies": {"@langchain/core": "^0.1.0", "openai": "^4.0.0"}
	}`), 0o644)

	result := Detect(root)
	names := map[string]bool{}
	for _, f := range result.Frameworks {
		names[f.Name] = true
	}
	if !names["langchain"] {
		t.Error("expected langchain detected from package.json")
	}
	if !names["openai"] {
		t.Error("expected openai detected from package.json")
	}
}

func TestDetect_PythonDeps(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "requirements.txt"), []byte("deepeval>=0.21\nragas>=0.1\nlangchain\n"), 0o644)

	result := Detect(root)
	names := map[string]bool{}
	for _, f := range result.Frameworks {
		names[f.Name] = true
	}
	if !names["deepeval"] {
		t.Error("expected deepeval detected from requirements.txt")
	}
	if !names["ragas"] {
		t.Error("expected ragas detected from requirements.txt")
	}
	if !names["langchain"] {
		t.Error("expected langchain detected from requirements.txt")
	}
}

func TestDetect_ConfigFile_Promptfoo(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "promptfooconfig.yaml"), []byte("prompts:\n  - prompt1\n"), 0o644)

	result := Detect(root)
	if len(result.Frameworks) == 0 {
		t.Fatal("expected frameworks detected")
	}
	if result.Frameworks[0].Name != "promptfoo" {
		t.Errorf("expected promptfoo, got %s", result.Frameworks[0].Name)
	}
	if result.Frameworks[0].Source != "config" {
		t.Errorf("expected config source, got %s", result.Frameworks[0].Source)
	}
}

func TestDetect_SourceImports(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "src"), 0o755)
	os.WriteFile(filepath.Join(root, "src", "chain.ts"), []byte(`
import { ChatOpenAI } from "@langchain/openai";
import Anthropic from "@anthropic-ai/sdk";
`), 0o644)

	result := Detect(root)
	names := map[string]bool{}
	for _, f := range result.Frameworks {
		names[f.Name] = true
	}
	if !names["langchain"] {
		t.Error("expected langchain from import")
	}
	if !names["anthropic"] {
		t.Error("expected anthropic from import")
	}
}

func TestDetect_ModelFiles(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "src"), 0o755)
	os.WriteFile(filepath.Join(root, "src", "inference.py"), []byte(`
from openai import OpenAI
client = OpenAI()
response = client.chat.completions.create(model="gpt-4")
`), 0o644)

	result := Detect(root)
	if len(result.ModelFiles) == 0 {
		t.Error("expected model files detected")
	}
}

func TestDeriveScenarios_FromEvalDir(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	// Create eval test file that imports a prompt surface.
	os.MkdirAll(filepath.Join(root, "tests", "eval", "safety"), 0o755)
	os.MkdirAll(filepath.Join(root, "src", "prompts"), 0o755)

	os.WriteFile(filepath.Join(root, "src", "prompts", "builder.py"), []byte(`
def build_safety_prompt(text):
    return "Safety: " + text
`), 0o644)

	os.WriteFile(filepath.Join(root, "tests", "eval", "safety", "test_safety.py"), []byte(`
from src.prompts.builder import build_safety_prompt

def test_safe_input():
    assert build_safety_prompt("hello")
`), 0o644)

	detection := Detect(root)
	surfaces := []models.CodeSurface{
		{SurfaceID: "surface:src/prompts/builder.py:build_safety_prompt", Name: "build_safety_prompt", Path: "src/prompts/builder.py", Kind: models.SurfacePrompt},
	}
	testFiles := []models.TestFile{
		{Path: "tests/eval/safety/test_safety.py", Framework: "pytest"},
	}

	scenarios := DeriveScenarios(root, detection, surfaces, testFiles)
	if len(scenarios) == 0 {
		t.Fatal("expected at least 1 auto-derived scenario")
	}

	found := false
	for _, sc := range scenarios {
		if sc.Category == "safety" {
			found = true
			if len(sc.CoveredSurfaceIDs) == 0 {
				t.Error("scenario should have linked surfaces")
			}
		}
	}
	if !found {
		t.Error("expected safety category scenario from eval/safety/ directory")
	}
}

func TestDeriveScenarios_EmptyRepo(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	detection := Detect(root)
	scenarios := DeriveScenarios(root, detection, nil, nil)
	if len(scenarios) != 0 {
		t.Errorf("expected 0 scenarios for empty repo, got %d", len(scenarios))
	}
}

func TestDeduplicateFrameworks(t *testing.T) {
	t.Parallel()
	fws := []Framework{
		{Name: "openai", Confidence: 0.75, Source: "import"},
		{Name: "openai", Confidence: 0.9, Source: "dependency"},
		{Name: "langchain", Confidence: 0.8, Source: "import"},
	}
	deduped := deduplicateFrameworks(fws)
	if len(deduped) != 2 {
		t.Fatalf("expected 2 unique frameworks, got %d", len(deduped))
	}
	// Highest confidence should win.
	for _, f := range deduped {
		if f.Name == "openai" && f.Confidence != 0.9 {
			t.Errorf("expected openai confidence 0.9 (dependency wins), got %f", f.Confidence)
		}
	}
}
