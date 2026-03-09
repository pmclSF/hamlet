package analysis

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestBuildImportGraph_JSImports(t *testing.T) {
	dir := t.TempDir()

	// Create source file.
	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "auth.js"), []byte("export function login() {}"), 0644); err != nil {
		t.Fatalf("write auth.js: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "utils.js"), []byte("export function format() {}"), 0644); err != nil {
		t.Fatalf("write utils.js: %v", err)
	}

	// Create test file that imports the source.
	testDir := filepath.Join(dir, "src", "__tests__")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("mkdir __tests__: %v", err)
	}
	testContent := `import { login } from '../auth.js';
import { format } from '../utils';
const helper = require('../helper');

describe('auth', () => {
  it('should login', () => {
    expect(login()).toBeDefined();
  });
});
`
	if err := os.WriteFile(filepath.Join(testDir, "auth.test.js"), []byte(testContent), 0644); err != nil {
		t.Fatalf("write auth.test.js: %v", err)
	}

	// Create the helper file for require resolution.
	if err := os.WriteFile(filepath.Join(srcDir, "helper.js"), []byte("module.exports = {}"), 0644); err != nil {
		t.Fatalf("write helper.js: %v", err)
	}

	testFiles := []models.TestFile{
		{Path: "src/__tests__/auth.test.js", Framework: "jest"},
	}

	graph := BuildImportGraph(dir, testFiles)

	imports := graph.TestImports["src/__tests__/auth.test.js"]
	if imports == nil {
		t.Fatal("expected imports for auth.test.js")
	}

	// Should resolve ../auth.js → src/auth.js
	if !imports["src/auth.js"] {
		t.Errorf("expected import of src/auth.js, got: %v", imports)
	}

	// Should resolve ../utils → src/utils.js (extension resolution)
	if !imports["src/utils.js"] {
		t.Errorf("expected import of src/utils.js, got: %v", imports)
	}

	// Should resolve ../helper → src/helper.js
	if !imports["src/helper.js"] {
		t.Errorf("expected import of src/helper.js, got: %v", imports)
	}
}

func TestBuildImportGraph_GoPackage(t *testing.T) {
	dir := t.TempDir()

	// Create Go source and test files in the same package.
	pkgDir := filepath.Join(dir, "pkg", "auth")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatalf("mkdir pkg/auth: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "auth.go"), []byte("package auth\nfunc Login() {}"), 0644); err != nil {
		t.Fatalf("write auth.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "session.go"), []byte("package auth\nfunc Session() {}"), 0644); err != nil {
		t.Fatalf("write session.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "auth_test.go"), []byte("package auth\nimport \"testing\"\nfunc TestLogin(t *testing.T) {}"), 0644); err != nil {
		t.Fatalf("write auth_test.go: %v", err)
	}

	testFiles := []models.TestFile{
		{Path: "pkg/auth/auth_test.go", Framework: "go-testing"},
	}

	graph := BuildImportGraph(dir, testFiles)
	imports := graph.TestImports["pkg/auth/auth_test.go"]

	if !imports["pkg/auth/auth.go"] {
		t.Errorf("expected Go test to link to auth.go, got: %v", imports)
	}
	if !imports["pkg/auth/session.go"] {
		t.Errorf("expected Go test to link to session.go, got: %v", imports)
	}
}

func TestBuildImportGraph_IndexResolution(t *testing.T) {
	dir := t.TempDir()

	// Create a module with index.js.
	if err := os.MkdirAll(filepath.Join(dir, "src", "core"), 0755); err != nil {
		t.Fatalf("mkdir src/core: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "src", "core", "index.js"), []byte("export * from './engine'"), 0644); err != nil {
		t.Fatalf("write src/core/index.js: %v", err)
	}

	// Test that imports the directory.
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0755); err != nil {
		t.Fatalf("mkdir tests: %v", err)
	}
	testContent := `import { Engine } from '../src/core';
describe('core', () => { it('works', () => {}); });
`
	if err := os.WriteFile(filepath.Join(dir, "tests", "core.test.js"), []byte(testContent), 0644); err != nil {
		t.Fatalf("write tests/core.test.js: %v", err)
	}

	testFiles := []models.TestFile{
		{Path: "tests/core.test.js", Framework: "jest"},
	}

	graph := BuildImportGraph(dir, testFiles)
	imports := graph.TestImports["tests/core.test.js"]

	if !imports["src/core/index.js"] {
		t.Errorf("expected directory import to resolve to index.js, got: %v", imports)
	}
}

func TestImportedModules(t *testing.T) {
	graph := &ImportGraph{
		TestImports: map[string]map[string]bool{
			"tests/a.test.js": {"src/a.js": true, "src/b.js": true},
			"tests/c.test.js": {"src/c.js": true, "src/a.js": true},
		},
	}

	mods := graph.ImportedModules()
	if len(mods) != 3 {
		t.Errorf("expected 3 unique modules, got %d: %v", len(mods), mods)
	}
	for _, want := range []string{"src/a.js", "src/b.js", "src/c.js"} {
		if !mods[want] {
			t.Errorf("expected %s in imported modules", want)
		}
	}
}
