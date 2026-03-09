package ownership

import (
	"os"
	"os/exec"
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
		{"packages/other/foo.test.js", "packages"},          // directory fallback
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
		{"src/api/routes.test.js", "test-infra"}, // wildcard matches last
		{"src/ui/button.ts", "frontend-team"},
		{"src/api/models.go", "backend-team"},
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

func TestResolver_ResolveAssignment_Provenance(t *testing.T) {
	dir := t.TempDir()
	hamletDir := filepath.Join(dir, ".hamlet")
	os.MkdirAll(hamletDir, 0o755)

	config := `ownership:
  rules:
    - path: "src/auth/"
      owner: "@team-auth"
`
	os.WriteFile(filepath.Join(hamletDir, "ownership.yaml"), []byte(config), 0o644)

	r := NewResolver(dir)
	a := r.ResolveAssignment("src/auth/login.js")

	if a.Source != SourceExplicitConfig {
		t.Errorf("source = %q, want %q", a.Source, SourceExplicitConfig)
	}
	if a.Confidence != ConfidenceHigh {
		t.Errorf("confidence = %q, want %q", a.Confidence, ConfidenceHigh)
	}
	if a.Inheritance != InheritanceDirect {
		t.Errorf("inheritance = %q, want %q", a.Inheritance, InheritanceDirect)
	}
	if a.PrimaryOwnerID() != "team-auth" {
		t.Errorf("owner = %q, want %q", a.PrimaryOwnerID(), "team-auth")
	}
	if a.SourceFile != ".hamlet/ownership.yaml" {
		t.Errorf("sourceFile = %q, want %q", a.SourceFile, ".hamlet/ownership.yaml")
	}
}

func TestResolver_ResolveAssignment_CodeownersProvenance(t *testing.T) {
	dir := t.TempDir()
	githubDir := filepath.Join(dir, ".github")
	os.MkdirAll(githubDir, 0o755)

	os.WriteFile(filepath.Join(githubDir, "CODEOWNERS"), []byte("/src/api/ @team-api @team-backend\n"), 0o644)

	r := NewResolver(dir)
	a := r.ResolveAssignment("src/api/handler.go")

	if a.Source != SourceCodeowners {
		t.Errorf("source = %q, want %q", a.Source, SourceCodeowners)
	}
	if len(a.Owners) != 2 {
		t.Fatalf("got %d owners, want 2", len(a.Owners))
	}
	if a.Owners[0].ID != "team-api" {
		t.Errorf("owner[0] = %q, want %q", a.Owners[0].ID, "team-api")
	}
	if a.Owners[1].ID != "team-backend" {
		t.Errorf("owner[1] = %q, want %q", a.Owners[1].ID, "team-backend")
	}
	if a.MatchedRule != "/src/api/" {
		t.Errorf("matchedRule = %q, want %q", a.MatchedRule, "/src/api/")
	}
}

func TestResolver_ResolveAssignment_DirectoryFallback(t *testing.T) {
	dir := t.TempDir()
	r := NewResolver(dir)

	a := r.ResolveAssignment("src/utils/helpers.js")
	if a.Source != SourceDirectoryFallback {
		t.Errorf("source = %q, want %q", a.Source, SourceDirectoryFallback)
	}
	if a.Confidence != ConfidenceLow {
		t.Errorf("confidence = %q, want %q", a.Confidence, ConfidenceLow)
	}
}

func TestResolver_ResolveAssignment_Unknown(t *testing.T) {
	dir := t.TempDir()
	r := NewResolver(dir)

	a := r.ResolveAssignment("standalone.js")
	if a.Source != SourceUnknown {
		t.Errorf("source = %q, want %q", a.Source, SourceUnknown)
	}
	if a.Confidence != ConfidenceNone {
		t.Errorf("confidence = %q, want %q", a.Confidence, ConfidenceNone)
	}
	if !a.IsUnowned() {
		t.Error("standalone file should be unowned")
	}
}

func TestResolver_PathMappings(t *testing.T) {
	dir := t.TempDir()
	hamletDir := filepath.Join(dir, ".hamlet")
	os.MkdirAll(hamletDir, 0o755)

	config := `ownership:
  rules: []
  path_mappings:
    - prefix: "lib/payments/"
      owners: ["@team-pay", "@team-billing"]
    - prefix: "lib/auth/"
      owners: ["@team-auth"]
`
	os.WriteFile(filepath.Join(hamletDir, "ownership.yaml"), []byte(config), 0o644)

	r := NewResolver(dir)
	a := r.ResolveAssignment("lib/payments/stripe.go")

	if a.Source != SourcePathMapping {
		t.Errorf("source = %q, want %q", a.Source, SourcePathMapping)
	}
	if len(a.Owners) != 2 {
		t.Fatalf("got %d owners, want 2", len(a.Owners))
	}
	if a.Owners[0].ID != "team-pay" {
		t.Errorf("owner[0] = %q, want %q", a.Owners[0].ID, "team-pay")
	}
	if a.Confidence != ConfidenceMedium {
		t.Errorf("confidence = %q, want %q", a.Confidence, ConfidenceMedium)
	}
}

func TestResolver_InheritFrom(t *testing.T) {
	parent := OwnershipAssignment{
		Owners:      []Owner{{ID: "team-auth"}},
		Source:      SourceCodeowners,
		Confidence:  ConfidenceHigh,
		Inheritance: InheritanceDirect,
		MatchedRule: "/src/auth/",
		SourceFile:  ".github/CODEOWNERS",
	}

	child := InheritFrom(parent)
	if child.Inheritance != InheritanceInherited {
		t.Errorf("inheritance = %q, want %q", child.Inheritance, InheritanceInherited)
	}
	if child.PrimaryOwnerID() != "team-auth" {
		t.Errorf("owner = %q, want %q", child.PrimaryOwnerID(), "team-auth")
	}
	if child.Source != SourceCodeowners {
		t.Errorf("source = %q, want %q", child.Source, SourceCodeowners)
	}
}

func TestResolver_SourcesUsed(t *testing.T) {
	dir := t.TempDir()

	// Set up CODEOWNERS.
	githubDir := filepath.Join(dir, ".github")
	os.MkdirAll(githubDir, 0o755)
	os.WriteFile(filepath.Join(githubDir, "CODEOWNERS"), []byte("* @owner\n"), 0o644)

	// Set up explicit config.
	hamletDir := filepath.Join(dir, ".hamlet")
	os.MkdirAll(hamletDir, 0o755)
	config := `ownership:
  rules:
    - path: "src/"
      owner: "explicit"
`
	os.WriteFile(filepath.Join(hamletDir, "ownership.yaml"), []byte(config), 0o644)

	r := NewResolver(dir)
	sources := r.SourcesUsed()

	if len(sources) < 2 {
		t.Errorf("expected at least 2 sources, got %d", len(sources))
	}
	if !r.HasCodeowners() {
		t.Error("expected HasCodeowners to be true")
	}
}

func TestResolver_ResolveAll(t *testing.T) {
	dir := t.TempDir()
	r := NewResolver(dir)

	paths := []string{"src/a.js", "lib/b.go", "standalone.js"}
	result := r.ResolveAll(paths)

	if len(result) != 3 {
		t.Errorf("got %d results, want 3", len(result))
	}
	a := result["src/a.js"]
	if a.PrimaryOwnerID() != "src" {
		t.Errorf("src/a.js owner = %q, want %q", a.PrimaryOwnerID(), "src")
	}
}

func TestResolver_GitHistoryFallback(t *testing.T) {
	requireGit(t)

	dir := t.TempDir()
	initTestRepo(t, dir)
	writeAndCommitFile(t, dir, "src/service/handler.go", "package service\n", "alice")

	hamletDir := filepath.Join(dir, ".hamlet")
	if err := os.MkdirAll(hamletDir, 0o755); err != nil {
		t.Fatalf("mkdir .hamlet: %v", err)
	}
	config := `ownership:
  git_history:
    enabled: true
    max_commits: 50
`
	if err := os.WriteFile(filepath.Join(hamletDir, "ownership.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write ownership config: %v", err)
	}

	r := NewResolver(dir)
	a := r.ResolveAssignment("src/service/handler.go")

	if a.Source != SourceGitHistory {
		t.Fatalf("source = %q, want %q", a.Source, SourceGitHistory)
	}
	if a.PrimaryOwnerID() != "alice" {
		t.Fatalf("owner = %q, want %q", a.PrimaryOwnerID(), "alice")
	}
	if a.Confidence != ConfidenceLow {
		t.Fatalf("confidence = %q, want %q", a.Confidence, ConfidenceLow)
	}
}

func TestResolver_GitHistoryDisabledUsesDirectoryFallback(t *testing.T) {
	requireGit(t)

	dir := t.TempDir()
	initTestRepo(t, dir)
	writeAndCommitFile(t, dir, "src/service/handler.go", "package service\n", "alice")

	r := NewResolver(dir)
	a := r.ResolveAssignment("src/service/handler.go")
	if a.Source != SourceDirectoryFallback {
		t.Fatalf("source = %q, want %q", a.Source, SourceDirectoryFallback)
	}
	if a.PrimaryOwnerID() != "src" {
		t.Fatalf("owner = %q, want %q", a.PrimaryOwnerID(), "src")
	}
}

func TestResolver_GitHistoryPrecedence(t *testing.T) {
	requireGit(t)

	dir := t.TempDir()
	initTestRepo(t, dir)

	writeAndCommitFile(t, dir, "src/explicit/a.go", "package explicit\n", "alice")
	writeAndCommitFile(t, dir, "src/mapped/b.go", "package mapped\n", "bob")
	writeAndCommitFile(t, dir, "src/codeowned/c.go", "package codeowned\n", "carol")
	writeAndCommitFile(t, dir, "src/history/d.go", "package history\n", "diana")

	githubDir := filepath.Join(dir, ".github")
	if err := os.MkdirAll(githubDir, 0o755); err != nil {
		t.Fatalf("mkdir .github: %v", err)
	}
	if err := os.WriteFile(filepath.Join(githubDir, "CODEOWNERS"), []byte("src/codeowned/ @codeowners-team\n"), 0o644); err != nil {
		t.Fatalf("write CODEOWNERS: %v", err)
	}

	hamletDir := filepath.Join(dir, ".hamlet")
	if err := os.MkdirAll(hamletDir, 0o755); err != nil {
		t.Fatalf("mkdir .hamlet: %v", err)
	}
	config := `ownership:
  rules:
    - path: "src/explicit/"
      owner: "explicit-team"
  path_mappings:
    - prefix: "src/mapped/"
      owners: ["mapped-team"]
  git_history:
    enabled: true
`
	if err := os.WriteFile(filepath.Join(hamletDir, "ownership.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write ownership config: %v", err)
	}

	r := NewResolver(dir)

	explicit := r.ResolveAssignment("src/explicit/a.go")
	if explicit.Source != SourceExplicitConfig || explicit.PrimaryOwnerID() != "explicit-team" {
		t.Fatalf("explicit resolution = (%q, %q), want (%q, %q)", explicit.Source, explicit.PrimaryOwnerID(), SourceExplicitConfig, "explicit-team")
	}

	mapped := r.ResolveAssignment("src/mapped/b.go")
	if mapped.Source != SourcePathMapping || mapped.PrimaryOwnerID() != "mapped-team" {
		t.Fatalf("path mapping resolution = (%q, %q), want (%q, %q)", mapped.Source, mapped.PrimaryOwnerID(), SourcePathMapping, "mapped-team")
	}

	codeowned := r.ResolveAssignment("src/codeowned/c.go")
	if codeowned.Source != SourceCodeowners || codeowned.PrimaryOwnerID() != "codeowners-team" {
		t.Fatalf("CODEOWNERS resolution = (%q, %q), want (%q, %q)", codeowned.Source, codeowned.PrimaryOwnerID(), SourceCodeowners, "codeowners-team")
	}

	history := r.ResolveAssignment("src/history/d.go")
	if history.Source != SourceGitHistory || history.PrimaryOwnerID() != "diana" {
		t.Fatalf("git history resolution = (%q, %q), want (%q, %q)", history.Source, history.PrimaryOwnerID(), SourceGitHistory, "diana")
	}
}

func TestResolver_GitHistorySourceReported(t *testing.T) {
	dir := t.TempDir()
	hamletDir := filepath.Join(dir, ".hamlet")
	if err := os.MkdirAll(hamletDir, 0o755); err != nil {
		t.Fatalf("mkdir .hamlet: %v", err)
	}
	config := `ownership:
  git_history:
    enabled: true
`
	if err := os.WriteFile(filepath.Join(hamletDir, "ownership.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write ownership config: %v", err)
	}

	r := NewResolver(dir)
	sources := r.SourcesUsed()
	if !containsSource(sources, SourceGitHistory) {
		t.Fatalf("expected sources to include %q, got %v", SourceGitHistory, sources)
	}
}

func TestResolver_GitHistoryDiagnosticsWhenNotRepo(t *testing.T) {
	requireGit(t)

	dir := t.TempDir()
	hamletDir := filepath.Join(dir, ".hamlet")
	if err := os.MkdirAll(hamletDir, 0o755); err != nil {
		t.Fatalf("mkdir .hamlet: %v", err)
	}
	config := `ownership:
  git_history:
    enabled: true
`
	if err := os.WriteFile(filepath.Join(hamletDir, "ownership.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("write ownership config: %v", err)
	}

	r := NewResolver(dir)
	a := r.ResolveAssignment("src/whatever.go")
	if a.Source != SourceDirectoryFallback {
		t.Fatalf("source = %q, want %q", a.Source, SourceDirectoryFallback)
	}

	diags := r.Diagnostics()
	found := false
	for _, d := range diags {
		if d.Source == "git-history" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected git-history diagnostic, got %v", diags)
	}
}

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

func initTestRepo(t *testing.T, dir string) {
	t.Helper()
	runGit(t, dir, nil, "init")
}

func writeAndCommitFile(t *testing.T, dir, relPath, content, author string) {
	t.Helper()

	absPath := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", filepath.Dir(relPath), err)
	}
	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q: %v", relPath, err)
	}

	runGit(t, dir, nil, "add", relPath)
	authorEnv := []string{
		"GIT_AUTHOR_NAME=" + author,
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=" + author,
		"GIT_COMMITTER_EMAIL=test@example.com",
	}
	runGit(t, dir, authorEnv, "commit", "-m", "add "+relPath)
}

func runGit(t *testing.T, dir string, extraEnv []string, args ...string) {
	t.Helper()
	gitArgs := append([]string{"-C", dir}, args...)
	cmd := exec.Command("git", gitArgs...)
	cmd.Env = append(os.Environ(), extraEnv...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}

func containsSource(sources []SourceType, want SourceType) bool {
	for _, source := range sources {
		if source == want {
			return true
		}
	}
	return false
}
