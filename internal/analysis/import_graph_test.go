package analysis

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestBuildImportGraph_JSImports(t *testing.T) {
	t.Parallel()
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

func TestBuildImportGraph_JSMultilineImportList(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "auth.js"), []byte("export const login = () => true"), 0o644); err != nil {
		t.Fatalf("write src/auth.js: %v", err)
	}

	testDir := filepath.Join(dir, "tests")
	if err := os.MkdirAll(testDir, 0o755); err != nil {
		t.Fatalf("mkdir tests: %v", err)
	}
	testContent := `import {
  login
} from '../src/auth'

describe('auth', () => {
  it('works', () => {
    expect(login()).toBe(true)
  })
})`
	if err := os.WriteFile(filepath.Join(testDir, "auth.test.js"), []byte(testContent), 0o644); err != nil {
		t.Fatalf("write tests/auth.test.js: %v", err)
	}

	graph := BuildImportGraph(dir, []models.TestFile{
		{Path: "tests/auth.test.js", Framework: "jest"},
	})
	imports := graph.TestImports["tests/auth.test.js"]
	if !imports["src/auth.js"] {
		t.Fatalf("expected multiline import to resolve src/auth.js, got %v", imports)
	}
}

func TestBuildImportGraph_GoPackage(t *testing.T) {
	t.Parallel()
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

func TestBuildImportGraph_GoCrossPackageImport(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/app\n\ngo 1.23\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	authDir := filepath.Join(dir, "pkg", "auth")
	if err := os.MkdirAll(authDir, 0o755); err != nil {
		t.Fatalf("mkdir pkg/auth: %v", err)
	}
	if err := os.WriteFile(filepath.Join(authDir, "auth.go"), []byte("package auth\nfunc Login() {}"), 0o644); err != nil {
		t.Fatalf("write auth.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(authDir, "auth_test.go"), []byte(`package auth
import (
	"testing"
	"example.com/app/pkg/common"
)
func TestLogin(t *testing.T) { common.Helper() }`), 0o644); err != nil {
		t.Fatalf("write auth_test.go: %v", err)
	}

	commonDir := filepath.Join(dir, "pkg", "common")
	if err := os.MkdirAll(commonDir, 0o755); err != nil {
		t.Fatalf("mkdir pkg/common: %v", err)
	}
	if err := os.WriteFile(filepath.Join(commonDir, "common.go"), []byte("package common\nfunc Helper() {}"), 0o644); err != nil {
		t.Fatalf("write common.go: %v", err)
	}

	graph := BuildImportGraph(dir, []models.TestFile{
		{Path: "pkg/auth/auth_test.go", Framework: "go-testing"},
	})
	imports := graph.TestImports["pkg/auth/auth_test.go"]

	if !imports["pkg/auth/auth.go"] {
		t.Fatalf("expected same-package source link, got %v", imports)
	}
	if !imports["pkg/common/common.go"] {
		t.Fatalf("expected cross-package source link, got %v", imports)
	}
}

func TestBuildImportGraph_IndexResolution(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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

func TestBuildImportGraph_TSConfigPathAlias(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte(`{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"]
    }
  }
}`), 0o644); err != nil {
		t.Fatalf("write tsconfig: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "src", "core"), 0o755); err != nil {
		t.Fatalf("mkdir src/core: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "src", "core", "auth.ts"), []byte("export const login = () => true"), 0o644); err != nil {
		t.Fatalf("write src/core/auth.ts: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatalf("mkdir tests: %v", err)
	}
	testContent := `import { login } from '@/core/auth'
describe('auth', () => { it('works', () => { expect(login()).toBe(true) }) })`
	if err := os.WriteFile(filepath.Join(dir, "tests", "auth.test.ts"), []byte(testContent), 0o644); err != nil {
		t.Fatalf("write tests/auth.test.ts: %v", err)
	}

	testFiles := []models.TestFile{
		{Path: "tests/auth.test.ts", Framework: "vitest"},
	}
	graph := BuildImportGraph(dir, testFiles)
	imports := graph.TestImports["tests/auth.test.ts"]

	if !imports["src/core/auth.ts"] {
		t.Fatalf("expected alias import to resolve to src/core/auth.ts, got %v", imports)
	}
}

func TestBuildImportGraph_PythonAbsoluteImport(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	if err := os.MkdirAll(filepath.Join(dir, "src", "mypackage"), 0o755); err != nil {
		t.Fatalf("mkdir src/mypackage: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "src", "mypackage", "__init__.py"), []byte(""), 0o644); err != nil {
		t.Fatalf("write __init__.py: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "src", "mypackage", "utils.py"), []byte("def helper(): pass"), 0o644); err != nil {
		t.Fatalf("write utils.py: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatalf("mkdir tests: %v", err)
	}
	content := `from mypackage.utils import helper
import mypackage`
	if err := os.WriteFile(filepath.Join(dir, "tests", "test_utils.py"), []byte(content), 0o644); err != nil {
		t.Fatalf("write test_utils.py: %v", err)
	}

	testFiles := []models.TestFile{
		{Path: "tests/test_utils.py", Framework: "pytest"},
	}
	graph := BuildImportGraph(dir, testFiles)
	imports := graph.TestImports["tests/test_utils.py"]
	if !imports["src/mypackage/utils.py"] {
		t.Fatalf("expected absolute module import to resolve, got %v", imports)
	}
	if !imports["src/mypackage/__init__.py"] {
		t.Fatalf("expected package import to resolve __init__.py, got %v", imports)
	}
}

func TestBuildImportGraph_WorkspacePackageAlias(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	rootPkg := `{
  "name": "root",
  "workspaces": ["packages/*"]
}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(rootPkg), 0o644); err != nil {
		t.Fatalf("write root package.json: %v", err)
	}

	pkgDir := filepath.Join(dir, "packages", "utils")
	if err := os.MkdirAll(filepath.Join(pkgDir, "src"), 0o755); err != nil {
		t.Fatalf("mkdir workspace pkg: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "package.json"), []byte(`{"name":"@org/utils"}`), 0o644); err != nil {
		t.Fatalf("write workspace package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "src", "index.ts"), []byte("export const x = 1"), 0o644); err != nil {
		t.Fatalf("write workspace source: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatalf("mkdir tests: %v", err)
	}
	testContent := `import { x } from '@org/utils/src/index'`
	if err := os.WriteFile(filepath.Join(dir, "tests", "ws.test.ts"), []byte(testContent), 0o644); err != nil {
		t.Fatalf("write ws.test.ts: %v", err)
	}

	testFiles := []models.TestFile{
		{Path: "tests/ws.test.ts", Framework: "vitest"},
	}
	graph := BuildImportGraph(dir, testFiles)
	imports := graph.TestImports["tests/ws.test.ts"]
	if !imports["packages/utils/src/index.ts"] {
		t.Fatalf("expected workspace alias to resolve to package source, got %v", imports)
	}
}

func TestBuildImportGraph_PNPMWorkspaceAlias(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pnpm-workspace.yaml"), []byte("packages:\n  - packages/*\n"), 0o644); err != nil {
		t.Fatalf("write pnpm-workspace.yaml: %v", err)
	}

	pkgDir := filepath.Join(dir, "packages", "ui")
	if err := os.MkdirAll(filepath.Join(pkgDir, "src"), 0o755); err != nil {
		t.Fatalf("mkdir workspace pkg: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "package.json"), []byte(`{"name":"@org/ui"}`), 0o644); err != nil {
		t.Fatalf("write workspace package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "src", "button.ts"), []byte("export const Button = () => null"), 0o644); err != nil {
		t.Fatalf("write workspace source: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatalf("mkdir tests: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tests", "ui.test.ts"), []byte(`import { Button } from '@org/ui/src/button'`), 0o644); err != nil {
		t.Fatalf("write ui.test.ts: %v", err)
	}

	graph := BuildImportGraph(dir, []models.TestFile{
		{Path: "tests/ui.test.ts", Framework: "vitest"},
	})
	imports := graph.TestImports["tests/ui.test.ts"]
	if !imports["packages/ui/src/button.ts"] {
		t.Fatalf("expected pnpm workspace alias to resolve, got %v", imports)
	}
}

func TestBuildImportGraph_LernaWorkspaceAlias(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "lerna.json"), []byte(`{"packages":["modules/*"]}`), 0o644); err != nil {
		t.Fatalf("write lerna.json: %v", err)
	}

	pkgDir := filepath.Join(dir, "modules", "core")
	if err := os.MkdirAll(filepath.Join(pkgDir, "src"), 0o755); err != nil {
		t.Fatalf("mkdir workspace pkg: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "package.json"), []byte(`{"name":"@org/core"}`), 0o644); err != nil {
		t.Fatalf("write workspace package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "src", "index.ts"), []byte("export const core = true"), 0o644); err != nil {
		t.Fatalf("write workspace source: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatalf("mkdir tests: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tests", "core.test.ts"), []byte(`import { core } from '@org/core/src/index'`), 0o644); err != nil {
		t.Fatalf("write core.test.ts: %v", err)
	}

	graph := BuildImportGraph(dir, []models.TestFile{
		{Path: "tests/core.test.ts", Framework: "vitest"},
	})
	imports := graph.TestImports["tests/core.test.ts"]
	if !imports["modules/core/src/index.ts"] {
		t.Fatalf("expected lerna workspace alias to resolve, got %v", imports)
	}
}

func TestBuildImportGraph_PackageImportsAlias(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	rootPkg := `{
  "name": "repo",
  "imports": {
    "#internal/*": "./src/internal/*.js",
    "#shared": "./src/shared/index.ts"
  }
}`
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(rootPkg), 0o644); err != nil {
		t.Fatalf("write package.json: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "src", "internal"), 0o755); err != nil {
		t.Fatalf("mkdir src/internal: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "src", "internal", "config.js"), []byte("export const cfg = {}"), 0o644); err != nil {
		t.Fatalf("write src/internal/config.js: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "src", "shared"), 0o755); err != nil {
		t.Fatalf("mkdir src/shared: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "src", "shared", "index.ts"), []byte("export const shared = true"), 0o644); err != nil {
		t.Fatalf("write src/shared/index.ts: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatalf("mkdir tests: %v", err)
	}
	testContent := `import { cfg } from '#internal/config'
import { shared } from '#shared'`
	if err := os.WriteFile(filepath.Join(dir, "tests", "imports.test.ts"), []byte(testContent), 0o644); err != nil {
		t.Fatalf("write imports.test.ts: %v", err)
	}

	graph := BuildImportGraph(dir, []models.TestFile{
		{Path: "tests/imports.test.ts", Framework: "vitest"},
	})
	imports := graph.TestImports["tests/imports.test.ts"]
	if !imports["src/internal/config.js"] {
		t.Fatalf("expected #internal/* alias to resolve, got %v", imports)
	}
	if !imports["src/shared/index.ts"] {
		t.Fatalf("expected #shared alias to resolve, got %v", imports)
	}
}
