package analysis

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectJSFrameworkResult_NodeTest(t *testing.T) {
	t.Parallel()
	// Create a temporary file with node:test import.
	dir := t.TempDir()
	file := filepath.Join(dir, "example.test.js")
	content := `import { describe, it } from 'node:test';
import assert from 'node:assert';

describe('example', () => {
  it('works', () => {
    assert.ok(true);
  });
});
`
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result := detectJSFrameworkResult(file)
	if result.Framework != "node-test" {
		t.Errorf("framework = %q, want 'node-test'", result.Framework)
	}
	if result.Source != "import" {
		t.Errorf("source = %q, want 'import'", result.Source)
	}
	if result.Confidence < 0.9 {
		t.Errorf("confidence = %v, want >= 0.9", result.Confidence)
	}
}

func TestDetectJSFrameworkResult_RequireNodeTest(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := filepath.Join(dir, "example.test.js")
	content := `const { describe, it } = require('node:test');
const assert = require('node:assert');

describe('example', () => {
  it('works', () => {
    assert.ok(true);
  });
});
`
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result := detectJSFrameworkResult(file)
	if result.Framework != "node-test" {
		t.Errorf("framework = %q, want 'node-test'", result.Framework)
	}
}

func TestDetectFrameworkWithContext_Fallback(t *testing.T) {
	t.Parallel()
	// Create a JS test file with no framework indicators.
	dir := t.TempDir()
	file := filepath.Join(dir, "example.test.js")
	content := `// A plain test file with no framework imports
function add(a, b) { return a + b; }
`
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Without project context → unknown.
	result := detectFrameworkWithContext("example.test.js", file, nil)
	if result.Framework != "unknown" {
		t.Errorf("without context: framework = %q, want 'unknown'", result.Framework)
	}

	// With project context → falls back to project default.
	ctx := &ProjectContext{
		Frameworks: map[string][]ProjectFramework{
			"javascript": {{Name: "jest", Source: "dependency", Confidence: 0.85}},
		},
	}
	result = detectFrameworkWithContext("example.test.js", file, ctx)
	if result.Framework != "jest" {
		t.Errorf("with context: framework = %q, want 'jest'", result.Framework)
	}
	if result.Source != "project-fallback" {
		t.Errorf("source = %q, want 'project-fallback'", result.Source)
	}
	if result.Confidence != 0.4 {
		t.Errorf("confidence = %v, want 0.4", result.Confidence)
	}
}

func TestDetectProjectFrameworks_PackageJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	pkgJSON := `{
  "name": "test-project",
  "devDependencies": {
    "jest": "^29.0.0",
    "cypress": "^12.0.0"
  }
}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := DetectProjectFrameworks(dir)
	jsFws := ctx.Frameworks["javascript"]
	if len(jsFws) < 2 {
		t.Fatalf("expected at least 2 JS frameworks, got %d", len(jsFws))
	}

	names := map[string]bool{}
	for _, fw := range jsFws {
		names[fw.Name] = true
	}
	if !names["jest"] {
		t.Error("expected jest in project frameworks")
	}
	if !names["cypress"] {
		t.Error("expected cypress in project frameworks")
	}
}

func TestDetectProjectFrameworks_ConfigFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// Create a vitest config — should be detected with high confidence.
	if err := os.WriteFile(filepath.Join(dir, "vitest.config.ts"), []byte("export default {}"), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := DetectProjectFrameworks(dir)
	jsFws := ctx.Frameworks["javascript"]
	if len(jsFws) == 0 {
		t.Fatal("expected vitest in project frameworks")
	}
	if jsFws[0].Name != "vitest" {
		t.Errorf("framework = %q, want 'vitest'", jsFws[0].Name)
	}
	if jsFws[0].Source != "config-file" {
		t.Errorf("source = %q, want 'config-file'", jsFws[0].Source)
	}
	if jsFws[0].Confidence < 0.9 {
		t.Errorf("confidence = %v, want >= 0.9", jsFws[0].Confidence)
	}
}

func TestDetectProjectFrameworks_GoMod(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/foo\ngo 1.23\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ctx := DetectProjectFrameworks(dir)
	goFws := ctx.Frameworks["go"]
	if len(goFws) != 1 || goFws[0].Name != "go-testing" {
		t.Errorf("expected go-testing, got %v", goFws)
	}
}

func TestInferFrameworkType_NodeTest(t *testing.T) {
	t.Parallel()
	ft := inferFrameworkType("node-test")
	if ft != "unit" {
		t.Errorf("inferFrameworkType('node-test') = %q, want 'unit'", ft)
	}
}

func TestDiscoverTestFiles_WithNodeTest(t *testing.T) {
	t.Parallel()
	root := fixtureRoot(t)
	files, err := discoverTestFiles(root)
	if err != nil {
		t.Fatalf("discoverTestFiles failed: %v", err)
	}

	// Find the server.test.js file with node:test.
	found := false
	for _, f := range files {
		if filepath.Base(f.Path) == "server.test.js" {
			found = true
			if f.Framework != "node-test" {
				t.Errorf("server.test.js framework = %q, want 'node-test'", f.Framework)
			}
			if f.FrameworkConfidence < 0.9 {
				t.Errorf("server.test.js confidence = %v, want >= 0.9", f.FrameworkConfidence)
			}
		}
	}
	if !found {
		t.Error("expected to find server.test.js in test files")
	}
}
