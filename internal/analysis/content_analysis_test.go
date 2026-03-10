package analysis

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestExtractJSExports_DefaultAndReExports(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	relPath := "src/module.js"
	absPath := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := `
export default function login() {}
export default class AuthClient {}
const foo = 1;
export { foo as Foo, bar };
export { baz } from './baz';
`
	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	units := extractJSExports(root, relPath)
	got := map[string]bool{}
	for _, u := range units {
		got[u.Name] = true
	}

	for _, want := range []string{"login", "AuthClient", "Foo", "bar", "baz"} {
		if !got[want] {
			t.Fatalf("missing exported unit %q in %+v", want, units)
		}
	}
}

func TestExtractGoExports_ExpandedKinds(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	relPath := "pkg/auth.go"
	absPath := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := `
package pkg

type Handler struct{}

const Version = "1.0.0"
var DefaultConfig = struct{}{}

func Exported() {}
func (h *Handler) Serve() {}
`
	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	units := extractGoExports(root, relPath)
	got := map[string]models.CodeUnit{}
	for _, u := range units {
		got[u.UnitID] = u
	}

	required := []string{
		"pkg/auth.go:Handler",
		"pkg/auth.go:Version",
		"pkg/auth.go:DefaultConfig",
		"pkg/auth.go:Exported",
		"pkg/auth.go:Handler.Serve",
	}
	for _, id := range required {
		if _, ok := got[id]; !ok {
			t.Fatalf("missing code unit %q in %+v", id, units)
		}
	}
	if got["pkg/auth.go:Handler.Serve"].Kind != models.CodeUnitKindMethod {
		t.Fatalf("Serve kind = %s, want method", got["pkg/auth.go:Handler.Serve"].Kind)
	}
}

func TestExtractJavaExports_PublicTypesAndMethods(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	relPath := "src/AuthService.java"
	absPath := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := `
public class AuthService {
  public void login() {}
  private void helper() {}
}

public interface Gateway {
  public void send();
}
`
	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	units := extractJavaExports(root, relPath)
	got := map[string]bool{}
	for _, u := range units {
		got[u.UnitID] = true
	}

	for _, id := range []string{
		"src/AuthService.java:AuthService",
		"src/AuthService.java:AuthService.login",
		"src/AuthService.java:Gateway",
		"src/AuthService.java:Gateway.send",
	} {
		if !got[id] {
			t.Fatalf("missing Java export %q in %+v", id, units)
		}
	}
}

func TestExtractJavaExports_IgnoresBracesInStringsAndComments(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	relPath := "src/Bracey.java"
	absPath := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := `
public class Bracey {
  public void first() {
    String json = "{\"k\": \"v\"}";
    // } should not close class
  }

  public void second() {}
}
`
	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	units := extractJavaExports(root, relPath)
	got := map[string]bool{}
	for _, u := range units {
		got[u.UnitID] = true
	}
	for _, id := range []string{
		"src/Bracey.java:Bracey",
		"src/Bracey.java:Bracey.first",
		"src/Bracey.java:Bracey.second",
	} {
		if !got[id] {
			t.Fatalf("missing Java export %q in %+v", id, units)
		}
	}
}

func TestWalkSourceFiles_SkipsSymlinkCycles(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	srcDir := filepath.Join(root, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "a.js"), []byte("export const A = 1"), 0o644); err != nil {
		t.Fatalf("write a.js: %v", err)
	}
	if err := os.Symlink(root, filepath.Join(srcDir, "cycle")); err != nil {
		t.Skipf("symlink not supported in this environment: %v", err)
	}

	var seen []string
	walkSourceFiles(root, func(relPath string) {
		seen = append(seen, relPath)
	})

	if len(seen) == 0 {
		t.Fatal("expected at least one source file discovered")
	}
	seenSet := map[string]int{}
	for _, p := range seen {
		seenSet[p]++
	}
	if seenSet["src/a.js"] != 1 {
		t.Fatalf("expected src/a.js discovered exactly once, got %d (%v)", seenSet["src/a.js"], seen)
	}
}

func TestExtractPythonExports_RespectsAll(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	relPath := "pkg/mod.py"
	absPath := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := `
__all__ = ["public_api"]

def public_api():
    return 1

def internal_helper():
    return 2
`
	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	units := extractPythonExports(root, relPath)
	if len(units) != 1 {
		t.Fatalf("expected exactly 1 exported Python unit from __all__, got %d (%+v)", len(units), units)
	}
	if units[0].Name != "public_api" {
		t.Fatalf("expected exported name public_api, got %q", units[0].Name)
	}
}
