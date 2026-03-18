package analysis

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
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

func TestExtractJSExports_MultilineNamedExportList(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	relPath := "src/multi.js"
	absPath := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := `
const alpha = 1
const beta = 2

export {
  alpha as Alpha,
  beta,
} from './values'
`
	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	units := extractJSExports(root, relPath)
	got := map[string]bool{}
	for _, u := range units {
		got[u.Name] = true
	}

	for _, want := range []string{"Alpha", "beta"} {
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

func TestExtractGoExports_LowercaseReceiverType(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	relPath := "pkg/worker.go"
	absPath := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := `
package pkg

type reader struct{}

func (r *reader) Read() {}
func (reader) Write() {}
`
	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	units := extractGoExports(root, relPath)
	got := map[string]bool{}
	for _, u := range units {
		got[u.UnitID] = true
	}

	for _, id := range []string{
		"pkg/worker.go:reader.Read",
		"pkg/worker.go:reader.Write",
	} {
		if !got[id] {
			t.Fatalf("missing Go method %q in %+v", id, units)
		}
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

// TestExtractExports_FileAndContentPathsMatch verifies that file-reading
// extractors and their content-based counterparts produce identical results
// for all four languages.
func TestExtractExports_FileAndContentPathsMatch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		relPath string
		content string
		lang    string
	}{
		{
			name:    "JS",
			relPath: "src/lib.js",
			content: "export function greet() {}\nexport class Client {}\nexport const VERSION = '1';\nmodule.exports = legacy;\n",
			lang:    "js",
		},
		{
			name:    "Go",
			relPath: "pkg/svc.go",
			content: "package pkg\n\ntype Service struct{}\n\nfunc NewService() *Service { return nil }\nfunc (s *Service) Run() {}\nconst MaxRetries = 3\n",
			lang:    "go",
		},
		{
			name:    "Python",
			relPath: "pkg/util.py",
			content: "def compute():\n    pass\n\ndef _private():\n    pass\n",
			lang:    "python",
		},
		{
			name:    "Java",
			relPath: "src/Handler.java",
			content: "public class Handler {\n  public void handle() {}\n  private void internal() {}\n}\n",
			lang:    "java",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			absPath := filepath.Join(root, tc.relPath)
			if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
				t.Fatalf("mkdir: %v", err)
			}
			if err := os.WriteFile(absPath, []byte(tc.content), 0o644); err != nil {
				t.Fatalf("write: %v", err)
			}

			// File-reading path (uses os.ReadFile internally).
			a := getLanguageAnalyzer(tc.lang)
			fromFile := a.ExtractExports(root, tc.relPath)

			// Content-based path (no I/O).
			fromContent := extractCodeUnitsFromSource(tc.content, tc.relPath, tc.lang)

			if len(fromFile) != len(fromContent) {
				t.Fatalf("count mismatch: file-path=%d content-path=%d\nfile:    %+v\ncontent: %+v",
					len(fromFile), len(fromContent), fromFile, fromContent)
			}
			for i := range fromFile {
				if fromFile[i].UnitID != fromContent[i].UnitID {
					t.Errorf("unit[%d] UnitID mismatch: %q vs %q", i, fromFile[i].UnitID, fromContent[i].UnitID)
				}
				if fromFile[i].Kind != fromContent[i].Kind {
					t.Errorf("unit[%d] Kind mismatch: %q vs %q", i, fromFile[i].Kind, fromContent[i].Kind)
				}
				if fromFile[i].StartLine != fromContent[i].StartLine {
					t.Errorf("unit[%d] StartLine mismatch: %d vs %d", i, fromFile[i].StartLine, fromContent[i].StartLine)
				}
			}
		})
	}
}

func TestCountSkips_JS(t *testing.T) {
	t.Parallel()
	src := `
		it.skip('should handle edge case', () => {});
		test.skip('another skip', () => {});
		describe.skip('skipped suite', () => {});
		xit('old skip', () => {});
		xdescribe('old suite skip', () => {});
		it('normal test', () => {});
	`
	got := countSkips(src, "jest")
	if got != 5 {
		t.Errorf("expected 5 JS skips, got %d", got)
	}
}

func TestCountSkips_Go(t *testing.T) {
	t.Parallel()
	src := `
		func TestFoo(t *testing.T) {
			t.Skip("not ready")
		}
		func TestBar(t *testing.T) {
			t.Skipf("skipping because %s", reason)
		}
	`
	got := countSkips(src, "go-testing")
	if got != 2 {
		t.Errorf("expected 2 Go skips, got %d", got)
	}
}

func TestCountSkips_Python(t *testing.T) {
	t.Parallel()
	src := `
		@pytest.mark.skip(reason="not ready")
		def test_alpha():
			pass

		@unittest.skip("broken")
		def test_beta():
			pass

		def test_gamma():
			pytest.skip("conditional")
	`
	got := countSkips(src, "pytest")
	if got != 3 {
		t.Errorf("expected 3 Python skips, got %d", got)
	}
}

func TestCountSkips_Java(t *testing.T) {
	t.Parallel()
	src := `
		@Disabled("not ready")
		@Test
		void testAlpha() {}

		@Ignore
		@Test
		void testBeta() {}
	`
	got := countSkips(src, "junit5")
	if got != 2 {
		t.Errorf("expected 2 Java skips, got %d", got)
	}
}

func TestCountSkips_NoSkips(t *testing.T) {
	t.Parallel()
	src := `
		it('should work', () => { expect(true).toBe(true); });
		test('another', () => {});
	`
	got := countSkips(src, "jest")
	if got != 0 {
		t.Errorf("expected 0 skips, got %d", got)
	}
}
