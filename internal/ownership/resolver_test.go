package ownership

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolver_ExplicitConfig(t *testing.T) {
	dir := t.TempDir()
	hamletDir := filepath.Join(dir, ".hamlet")
	os.MkdirAll(hamletDir, 0o755)

	config := `ownership:
  rules:
    - path: "packages/payments/"
      owner: "payments-team"
    - path: "packages/auth/"
      owner: "auth-team"
    - path: "packages/auth/mfa/"
      owner: "security-team"
`
	os.WriteFile(filepath.Join(hamletDir, "ownership.yaml"), []byte(config), 0o644)

	r := NewResolver(dir)

	tests := []struct {
		path string
		want string
	}{
		{"packages/payments/checkout.test.js", "payments-team"},
		{"packages/auth/login.test.js", "auth-team"},
		{"packages/auth/mfa/totp.test.js", "security-team"}, // longest match wins
		{"packages/other/foo.test.js", "packages"},           // directory fallback
	}

	for _, tt := range tests {
		got := r.Resolve(tt.path)
		if got != tt.want {
			t.Errorf("Resolve(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestResolver_CODEOWNERS(t *testing.T) {
	dir := t.TempDir()
	githubDir := filepath.Join(dir, ".github")
	os.MkdirAll(githubDir, 0o755)

	codeowners := `# CODEOWNERS
src/api/ @backend-team
src/ui/ @frontend-team
*.test.js @test-infra
`
	os.WriteFile(filepath.Join(githubDir, "CODEOWNERS"), []byte(codeowners), 0o644)

	r := NewResolver(dir)

	tests := []struct {
		path string
		want string
	}{
		{"src/api/routes.test.js", "@test-infra"}, // wildcard matches last
		{"src/ui/button.ts", "@frontend-team"},
		{"src/api/models.go", "@backend-team"},
	}

	for _, tt := range tests {
		got := r.Resolve(tt.path)
		if got != tt.want {
			t.Errorf("Resolve(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestResolver_DirectoryFallback(t *testing.T) {
	dir := t.TempDir()
	r := NewResolver(dir)

	got := r.Resolve("src/utils/helpers.test.js")
	if got != "src" {
		t.Errorf("expected directory fallback 'src', got %q", got)
	}
}

func TestResolver_Unknown(t *testing.T) {
	dir := t.TempDir()
	r := NewResolver(dir)

	got := r.Resolve("standalone.test.js")
	if got != "unknown" {
		t.Errorf("expected 'unknown' for root-level file, got %q", got)
	}
}

func TestResolver_Precedence(t *testing.T) {
	dir := t.TempDir()

	// Set up both CODEOWNERS and explicit config
	githubDir := filepath.Join(dir, ".github")
	os.MkdirAll(githubDir, 0o755)
	os.WriteFile(filepath.Join(githubDir, "CODEOWNERS"), []byte("src/ @code-owner\n"), 0o644)

	hamletDir := filepath.Join(dir, ".hamlet")
	os.MkdirAll(hamletDir, 0o755)
	config := `ownership:
  rules:
    - path: "src/"
      owner: "explicit-owner"
`
	os.WriteFile(filepath.Join(hamletDir, "ownership.yaml"), []byte(config), 0o644)

	r := NewResolver(dir)

	// Explicit config should take precedence over CODEOWNERS
	got := r.Resolve("src/foo.test.js")
	if got != "explicit-owner" {
		t.Errorf("expected explicit config to win, got %q", got)
	}
}

func TestResolver_NoConfigFiles(t *testing.T) {
	dir := t.TempDir()
	r := NewResolver(dir)

	// Should not panic, should return directory fallback or unknown
	got := r.Resolve("test/something.test.js")
	if got != "test" {
		t.Errorf("expected directory fallback 'test', got %q", got)
	}
}
