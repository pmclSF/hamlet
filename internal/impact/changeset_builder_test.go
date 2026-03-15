package impact

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

// --- ChangeSetFromGitDiff tests ---

func TestChangeSetFromGitDiff_ResolvesHead(t *testing.T) {
	t.Parallel()
	requireGit(t)
	repo := initImpactTestRepo(t)

	// Modify a file to create a diff.
	if err := os.WriteFile(filepath.Join(repo, "src", "app.js"), []byte("v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cs, err := ChangeSetFromGitDiff(repo, "")
	if err != nil {
		t.Fatalf("ChangeSetFromGitDiff: %v", err)
	}

	if cs.HeadSHA == "" {
		t.Error("expected HeadSHA to be resolved")
	}
	if cs.Repository != repo {
		t.Errorf("expected repository %q, got %q", repo, cs.Repository)
	}
	if cs.Source != "git-diff-working-tree" {
		t.Errorf("expected source git-diff-working-tree, got %q", cs.Source)
	}
	if len(cs.ChangedFiles) != 1 {
		t.Fatalf("expected 1 changed file, got %d", len(cs.ChangedFiles))
	}
	if cs.ChangedFiles[0].Path != "src/app.js" {
		t.Errorf("expected src/app.js, got %s", cs.ChangedFiles[0].Path)
	}
	if cs.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestChangeSetFromGitDiff_WithBaseRef(t *testing.T) {
	t.Parallel()
	requireGit(t)
	repo := initImpactTestRepo(t)

	// Create a second commit.
	newFile := filepath.Join(repo, "src", "new.js")
	if err := os.WriteFile(newFile, []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-m", "add new.js")

	cs, err := ChangeSetFromGitDiff(repo, "HEAD~1")
	if err != nil {
		t.Fatalf("ChangeSetFromGitDiff: %v", err)
	}

	if cs.Source != "git-diff" {
		t.Errorf("expected source git-diff, got %q", cs.Source)
	}
	if cs.BaseSHA == "" {
		t.Error("expected BaseSHA to be resolved")
	}
	if cs.HeadSHA == "" {
		t.Error("expected HeadSHA to be resolved")
	}
	if cs.BaseSHA == cs.HeadSHA {
		t.Error("BaseSHA and HeadSHA should differ")
	}

	found := false
	for _, f := range cs.ChangedFiles {
		if f.Path == "src/new.js" && f.ChangeKind == models.ChangeAdded {
			found = true
		}
	}
	if !found {
		t.Error("expected src/new.js as added file")
	}
}

func TestChangeSetFromGitDiff_InvalidBaseRef(t *testing.T) {
	t.Parallel()
	requireGit(t)
	repo := initImpactTestRepo(t)

	_, err := ChangeSetFromGitDiff(repo, "NONEXISTENT_REF")
	if err == nil {
		t.Fatal("expected error for invalid base ref")
	}
}

func TestChangeSetFromGitDiff_InfersPackages(t *testing.T) {
	t.Parallel()
	requireGit(t)
	repo := initImpactTestRepo(t)

	// Create files in multiple packages.
	authDir := filepath.Join(repo, "internal", "auth")
	dbDir := filepath.Join(repo, "internal", "db")
	if err := os.MkdirAll(authDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(authDir, "auth.go"), []byte("package auth\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dbDir, "db.go"), []byte("package db\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-m", "add packages")

	// Modify both.
	if err := os.WriteFile(filepath.Join(authDir, "auth.go"), []byte("package auth\n// v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dbDir, "db.go"), []byte("package db\n// v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cs, err := ChangeSetFromGitDiff(repo, "")
	if err != nil {
		t.Fatalf("ChangeSetFromGitDiff: %v", err)
	}

	if len(cs.ChangedPackages) != 2 {
		t.Fatalf("expected 2 changed packages, got %d: %v", len(cs.ChangedPackages), cs.ChangedPackages)
	}
}

// --- ChangeSetFromPaths tests ---

func TestChangeSetFromPaths_Basic(t *testing.T) {
	t.Parallel()
	cs := ChangeSetFromPaths(
		[]string{"src/foo.js", "test/foo.test.js", "services/auth/handler.go"},
		models.ChangeModified,
	)

	if cs.Source != "explicit" {
		t.Errorf("expected source explicit, got %q", cs.Source)
	}
	if len(cs.ChangedFiles) != 3 {
		t.Fatalf("expected 3 files, got %d", len(cs.ChangedFiles))
	}
	if cs.ChangedFiles[0].IsTestFile {
		t.Error("src/foo.js should not be a test file")
	}
	if !cs.ChangedFiles[1].IsTestFile {
		t.Error("test/foo.test.js should be a test file")
	}

	// Should infer services.
	if len(cs.ChangedServices) != 1 || cs.ChangedServices[0] != "auth" {
		t.Errorf("expected service [auth], got %v", cs.ChangedServices)
	}
}

// --- ChangeSetFromCIList tests ---

func TestChangeSetFromCIList_Basic(t *testing.T) {
	t.Parallel()
	cs := ChangeSetFromCIList("src/auth.js\ntest/auth.test.js\n\n  src/db.js  \n", "/repo")

	if cs.Source != "ci-changed-files" {
		t.Errorf("expected source ci-changed-files, got %q", cs.Source)
	}
	if len(cs.ChangedFiles) != 3 {
		t.Fatalf("expected 3 files, got %d", len(cs.ChangedFiles))
	}
	for _, cf := range cs.ChangedFiles {
		if cf.ChangeKind != models.ChangeModified {
			t.Errorf("expected modified kind, got %s", cf.ChangeKind)
		}
	}
}

func TestChangeSetFromCIList_Empty(t *testing.T) {
	t.Parallel()
	cs := ChangeSetFromCIList("", "")
	if len(cs.ChangedFiles) != 0 {
		t.Errorf("expected 0 files for empty input, got %d", len(cs.ChangedFiles))
	}
}

// --- ChangeSetFromComparison tests ---

func TestChangeSetFromComparison_Diff(t *testing.T) {
	t.Parallel()
	from := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{{Path: "test/a.test.js"}, {Path: "test/b.test.js"}},
		CodeUnits: []models.CodeUnit{{Name: "A", Path: "src/a.js"}},
	}
	to := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{{Path: "test/a.test.js"}, {Path: "test/c.test.js"}},
		CodeUnits: []models.CodeUnit{{Name: "A", Path: "src/a.js"}, {Name: "D", Path: "src/d.js"}},
	}

	cs := ChangeSetFromComparison(from, to)

	if cs.Source != "snapshot-compare" {
		t.Errorf("expected snapshot-compare, got %s", cs.Source)
	}

	byPath := map[string]models.ChangeKind{}
	for _, cf := range cs.ChangedFiles {
		byPath[cf.Path] = cf.ChangeKind
	}

	if byPath["src/d.js"] != models.ChangeAdded {
		t.Errorf("expected src/d.js added, got %s", byPath["src/d.js"])
	}
	if byPath["test/b.test.js"] != models.ChangeDeleted {
		t.Errorf("expected test/b.test.js deleted, got %s", byPath["test/b.test.js"])
	}
	if byPath["src/a.js"] != models.ChangeModified {
		t.Errorf("expected src/a.js modified, got %s", byPath["src/a.js"])
	}
}

func TestChangeSetFromComparison_NilInputs(t *testing.T) {
	t.Parallel()
	cs := ChangeSetFromComparison(nil, nil)
	if len(cs.ChangedFiles) != 0 {
		t.Errorf("expected 0 files, got %d", len(cs.ChangedFiles))
	}
}

// --- ChangeSetToScope bridge tests ---

func TestChangeSetToScope_Roundtrip(t *testing.T) {
	t.Parallel()
	cs := &models.ChangeSet{
		BaseRef: "main",
		HeadSHA: "abc123",
		Source:  "git-diff",
		ChangedFiles: []models.ChangedFile{
			{Path: "src/auth.js", ChangeKind: models.ChangeModified, IsTestFile: false},
			{Path: "test/auth.test.js", ChangeKind: models.ChangeModified, IsTestFile: true},
		},
	}

	scope := ChangeSetToScope(cs)

	if scope.Source != "git-diff" {
		t.Errorf("expected source git-diff, got %q", scope.Source)
	}
	if scope.BaselineRef != "main" {
		t.Errorf("expected baseline ref main, got %q", scope.BaselineRef)
	}
	if len(scope.ChangedFiles) != 2 {
		t.Fatalf("expected 2 files, got %d", len(scope.ChangedFiles))
	}
	if scope.ChangedFiles[0].Path != "src/auth.js" {
		t.Errorf("expected src/auth.js, got %s", scope.ChangedFiles[0].Path)
	}
	if string(scope.ChangedFiles[0].ChangeKind) != "modified" {
		t.Errorf("expected modified, got %s", scope.ChangedFiles[0].ChangeKind)
	}
}

// --- AnalyzeChangeSet tests ---

func TestAnalyzeChangeSet_CarriesMetadata(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/auth.test.js", Framework: "jest", LinkedCodeUnits: []string{"AuthService"}},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "AuthService", Path: "src/auth.js", Kind: "class", Exported: true},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 1},
		},
	}

	cs := &models.ChangeSet{
		Source:  "explicit",
		BaseSHA: "abc",
		HeadSHA: "def",
		ChangedFiles: []models.ChangedFile{
			{Path: "src/auth.js", ChangeKind: models.ChangeModified},
		},
		ChangedPackages: []string{"src/auth"},
		Limitations:     []string{"test limitation"},
	}

	result := AnalyzeChangeSet(cs, snap)

	if result.ChangeSet == nil {
		t.Fatal("expected ChangeSet on result")
	}
	if result.ChangeSet.BaseSHA != "abc" {
		t.Errorf("expected BaseSHA abc, got %s", result.ChangeSet.BaseSHA)
	}
	if result.ChangeSet.HeadSHA != "def" {
		t.Errorf("expected HeadSHA def, got %s", result.ChangeSet.HeadSHA)
	}

	// ChangeSet limitations should be merged into result limitations.
	foundCSLimit := false
	for _, l := range result.Limitations {
		if l == "test limitation" {
			foundCSLimit = true
		}
	}
	if !foundCSLimit {
		t.Error("expected ChangeSet limitation to appear in result.Limitations")
	}

	// Impact analysis should still work normally.
	if len(result.ImpactedUnits) != 1 {
		t.Errorf("expected 1 impacted unit, got %d", len(result.ImpactedUnits))
	}
}

func TestAnalyzeChangeSet_BackwardCompatibleScope(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{}

	cs := &models.ChangeSet{
		Source: "explicit",
		ChangedFiles: []models.ChangedFile{
			{Path: "src/foo.js", ChangeKind: models.ChangeModified},
		},
	}

	result := AnalyzeChangeSet(cs, snap)

	// The legacy Scope field should still be populated.
	if len(result.Scope.ChangedFiles) != 1 {
		t.Errorf("expected 1 scope file, got %d", len(result.Scope.ChangedFiles))
	}
	if result.Scope.Source != "explicit" {
		t.Errorf("expected scope source explicit, got %s", result.Scope.Source)
	}
}

// --- inference utility tests ---

func TestInferChangedPackages(t *testing.T) {
	t.Parallel()
	files := []models.ChangedFile{
		{Path: "internal/auth/handler.go", IsTestFile: false},
		{Path: "internal/db/pool.go", IsTestFile: false},
		{Path: "internal/auth/middleware.go", IsTestFile: false},
		{Path: "test/auth_test.go", IsTestFile: true}, // should be excluded
	}

	pkgs := inferChangedPackages(files)

	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d: %v", len(pkgs), pkgs)
	}
	// Go files: package = directory
	expected := map[string]bool{"internal/auth": true, "internal/db": true}
	for _, p := range pkgs {
		if !expected[p] {
			t.Errorf("unexpected package: %s", p)
		}
	}
}

func TestInferChangedServices(t *testing.T) {
	t.Parallel()
	files := []models.ChangedFile{
		{Path: "services/auth/handler.go"},
		{Path: "services/auth/middleware.go"},
		{Path: "cmd/api/main.go"},
		{Path: "apps/web/index.ts"},
		{Path: "src/utils.js"}, // not in a service dir
	}

	svcs := inferChangedServices(files)

	expected := map[string]bool{"auth": true, "api": true, "web": true}
	if len(svcs) != 3 {
		t.Fatalf("expected 3 services, got %d: %v", len(svcs), svcs)
	}
	for _, s := range svcs {
		if !expected[s] {
			t.Errorf("unexpected service: %s", s)
		}
	}
}

func TestCollectChangedConfigs(t *testing.T) {
	t.Parallel()
	files := []models.ChangedFile{
		{Path: "src/auth.js"},
		{Path: "package.json"},
		{Path: ".github/workflows/ci.yml"},
		{Path: "go.mod"},
		{Path: "Dockerfile"},
		{Path: "terraform/main.tf"},
	}

	configs := collectChangedConfigs(files)

	// Should include package.json, ci.yml, go.mod, Dockerfile, main.tf — but NOT auth.js
	if len(configs) != 5 {
		t.Fatalf("expected 5 configs, got %d: %v", len(configs), configs)
	}
	for _, c := range configs {
		if c == "src/auth.js" {
			t.Error("src/auth.js should not be classified as config")
		}
	}
}

func TestIsConfigFile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		want bool
	}{
		{"package.json", true},
		{"go.mod", true},
		{"Dockerfile", true},
		{"Makefile", true},
		{".github/workflows/ci.yml", true},
		{"terraform/main.tf", true},
		{"config.yaml", true},
		{"src/auth.js", false},
		{"internal/handler.go", false},
		{"test/foo.test.js", false},
	}

	for _, tt := range tests {
		got := isConfigFile(tt.path)
		if got != tt.want {
			t.Errorf("isConfigFile(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

// --- shallow clone detection ---

func TestChangeSetFromGitDiff_ShallowDetection(t *testing.T) {
	t.Parallel()
	requireGit(t)
	repo := initImpactTestRepo(t)

	// A normal repo (not shallow) should set IsShallow = false.
	if err := os.WriteFile(filepath.Join(repo, "src", "app.js"), []byte("v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cs, err := ChangeSetFromGitDiff(repo, "")
	if err != nil {
		t.Fatalf("ChangeSetFromGitDiff: %v", err)
	}
	if cs.IsShallow {
		t.Error("expected IsShallow=false for a normal repo")
	}
}

// --- parseGitDiffToChangedFiles tests ---

func TestParseGitDiffToChangedFiles(t *testing.T) {
	t.Parallel()
	output := "M\tsrc/auth.js\nA\tsrc/new.js\nD\tsrc/old.js\nR100\tsrc/from.js\tsrc/to.js"

	files := parseGitDiffToChangedFiles(output, "/repo")

	if len(files) != 4 {
		t.Fatalf("expected 4 files, got %d", len(files))
	}

	checks := []struct {
		idx  int
		kind models.ChangeKind
		path string
	}{
		{0, models.ChangeModified, "src/auth.js"},
		{1, models.ChangeAdded, "src/new.js"},
		{2, models.ChangeDeleted, "src/old.js"},
		{3, models.ChangeRenamed, "src/to.js"},
	}

	for _, c := range checks {
		if files[c.idx].ChangeKind != c.kind {
			t.Errorf("file %d: expected kind %s, got %s", c.idx, c.kind, files[c.idx].ChangeKind)
		}
		if files[c.idx].Path != c.path {
			t.Errorf("file %d: expected path %s, got %s", c.idx, c.path, files[c.idx].Path)
		}
	}
}

// helper: create a shallow clone for testing
func initShallowClone(t *testing.T) string {
	t.Helper()
	requireGit(t)

	// Create origin repo with 2 commits.
	origin := t.TempDir()
	runGit(t, origin, "init", "--bare")

	work := t.TempDir()
	runGit(t, work, "init")
	runGit(t, work, "config", "user.name", "Test")
	runGit(t, work, "config", "user.email", "test@test.com")
	if err := os.WriteFile(filepath.Join(work, "a.txt"), []byte("v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, work, "add", ".")
	runGit(t, work, "commit", "-m", "first")
	if err := os.WriteFile(filepath.Join(work, "b.txt"), []byte("v1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, work, "add", ".")
	runGit(t, work, "commit", "-m", "second")
	runGit(t, work, "remote", "add", "origin", origin)
	runGit(t, work, "push", "origin", "HEAD:main")

	// Shallow clone with depth=1.
	clone := t.TempDir()
	cmd := exec.Command("git", "clone", "--depth=1", "file://"+origin, clone)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("shallow clone failed: %v\n%s", err, out)
	}

	return clone
}

func TestChangeSetFromGitDiff_ShallowClone(t *testing.T) {
	t.Parallel()
	clone := initShallowClone(t)

	// Some CI environments or git versions don't produce true shallow clones
	// from file:// protocol. Skip if the clone isn't actually shallow.
	if !isShallowClone(clone) {
		t.Skip("git did not produce a shallow clone from file:// (CI git version may not support this)")
	}

	// Modify a file.
	if err := os.WriteFile(filepath.Join(clone, "b.txt"), []byte("v2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cs, err := ChangeSetFromGitDiff(clone, "")
	if err != nil {
		t.Fatalf("ChangeSetFromGitDiff: %v", err)
	}

	if !cs.IsShallow {
		t.Error("expected IsShallow=true for shallow clone")
	}
	if len(cs.Limitations) == 0 {
		t.Error("expected limitations for shallow clone")
	}
}
