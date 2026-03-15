package impact

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestAnalyze_WellProtectedChange(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/auth/__tests__/login.test.js", Framework: "jest", LinkedCodeUnits: []string{"AuthService"}},
			{Path: "src/auth/__tests__/session.test.js", Framework: "jest", LinkedCodeUnits: []string{"SessionManager"}},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "AuthService", Path: "src/auth/service.js", Kind: "class", Exported: true},
			{Name: "SessionManager", Path: "src/auth/session.js", Kind: "class", Exported: true},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 2},
		},
	}

	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "src/auth/service.js", ChangeKind: ChangeModified, IsTestFile: false},
		},
		Source: "explicit",
	}

	result := Analyze(scope, snap)

	if len(result.ImpactedUnits) != 1 {
		t.Errorf("expected 1 impacted unit, got %d", len(result.ImpactedUnits))
	}
	if result.ImpactedUnits[0].Name != "AuthService" {
		t.Errorf("expected AuthService, got %s", result.ImpactedUnits[0].Name)
	}
	if result.ImpactedUnits[0].ProtectionStatus != ProtectionStrong {
		t.Errorf("expected strong protection, got %s", result.ImpactedUnits[0].ProtectionStatus)
	}
	if len(result.ProtectionGaps) != 0 {
		t.Errorf("expected 0 protection gaps, got %d", len(result.ProtectionGaps))
	}
	// Posture is partially_protected because exported unit changes carry exposure risk,
	// even when coverage is strong.
	if result.Posture.Band != "partially_protected" {
		t.Errorf("expected partially_protected posture (exported unit), got %s", result.Posture.Band)
	}
}

func TestAnalyze_UntestedExportGap(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/__tests__/utils.test.js", Framework: "jest"},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "ApiClient", Path: "src/api/client.js", Kind: "class", Exported: true},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 1},
		},
	}

	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "src/api/client.js", ChangeKind: ChangeModified, IsTestFile: false},
		},
	}

	result := Analyze(scope, snap)

	if len(result.ImpactedUnits) != 1 {
		t.Fatalf("expected 1 impacted unit, got %d", len(result.ImpactedUnits))
	}
	if result.ImpactedUnits[0].ProtectionStatus != ProtectionNone {
		t.Errorf("expected no protection, got %s", result.ImpactedUnits[0].ProtectionStatus)
	}
	if len(result.ProtectionGaps) == 0 {
		t.Fatal("expected protection gaps")
	}
	if result.ProtectionGaps[0].GapType != "untested_export" {
		t.Errorf("expected untested_export gap, got %s", result.ProtectionGaps[0].GapType)
	}
	if result.ProtectionGaps[0].Severity != "high" {
		t.Errorf("expected high severity, got %s", result.ProtectionGaps[0].Severity)
	}
}

func TestAnalyze_NameOnlyLinkingRequiresUniqueSymbol(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/handler.test.js", Framework: "jest", LinkedCodeUnits: []string{"Handler"}},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "Handler", Path: "src/a/handler.js", Kind: "function", Exported: true},
			{Name: "Handler", Path: "src/b/handler.js", Kind: "function", Exported: true},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 1},
		},
	}

	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "src/a/handler.js", ChangeKind: ChangeModified, IsTestFile: false},
		},
	}

	result := Analyze(scope, snap)
	if len(result.ImpactedUnits) != 1 {
		t.Fatalf("expected 1 impacted unit, got %d", len(result.ImpactedUnits))
	}
	if result.ImpactedUnits[0].ProtectionStatus != ProtectionNone {
		t.Fatalf("expected no protection for ambiguous name-only linkage, got %s", result.ImpactedUnits[0].ProtectionStatus)
	}
	if len(result.ImpactedUnits[0].CoveringTests) != 0 {
		t.Fatalf("expected 0 covering tests for ambiguous linkage, got %d", len(result.ImpactedUnits[0].CoveringTests))
	}
}

func TestAnalyze_DirectlyChangedTest(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/__tests__/auth.test.js", Framework: "jest"},
		},
	}

	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "src/__tests__/auth.test.js", ChangeKind: ChangeModified, IsTestFile: true},
		},
	}

	result := Analyze(scope, snap)

	if len(result.ImpactedTests) == 0 {
		t.Fatal("expected impacted tests for directly changed test file")
	}
	found := false
	for _, it := range result.ImpactedTests {
		if it.IsDirectlyChanged {
			found = true
		}
	}
	if !found {
		t.Error("expected a directly changed test")
	}
}

func TestAnalyze_DirectoryProximity_IsPathAware(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/apple/apple.test.js", Framework: "jest"},
		},
	}

	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "src/api/service.js", ChangeKind: ChangeModified, IsTestFile: false},
		},
	}

	result := Analyze(scope, snap)
	if len(result.ImpactedTests) != 0 {
		t.Fatalf("expected 0 impacted tests, got %d (%v)", len(result.ImpactedTests), result.ImpactedTests)
	}
}

func TestAnalyze_FileWithNoCodeUnits(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/config.test.js", Framework: "jest"},
		},
	}

	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "src/config.js", ChangeKind: ChangeModified, IsTestFile: false},
		},
	}

	result := Analyze(scope, snap)

	// Should still create a file-level impacted unit.
	if len(result.ImpactedUnits) != 1 {
		t.Fatalf("expected 1 file-level impacted unit, got %d", len(result.ImpactedUnits))
	}
	if result.ImpactedUnits[0].ImpactConfidence != ConfidenceWeak {
		t.Errorf("expected weak confidence for file-level unit, got %s", result.ImpactedUnits[0].ImpactConfidence)
	}
	if len(result.Limitations) == 0 {
		t.Error("expected limitations about missing code units")
	}
}

func TestAnalyze_MultipleOwners(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		CodeUnits: []models.CodeUnit{
			{Name: "AuthService", Path: "src/auth/service.js", Exported: true},
			{Name: "PaymentService", Path: "src/payments/service.js", Exported: true},
		},
		Ownership: map[string][]string{
			"src/auth/service.js":     {"team-auth"},
			"src/payments/service.js": {"team-payments"},
		},
	}

	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "src/auth/service.js", ChangeKind: ChangeModified},
			{Path: "src/payments/service.js", ChangeKind: ChangeModified},
		},
	}

	result := Analyze(scope, snap)

	if len(result.ImpactedOwners) != 2 {
		t.Errorf("expected 2 impacted owners, got %d", len(result.ImpactedOwners))
	}
}

func TestAnalyze_Summary(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/a.test.js", Framework: "jest"},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "Foo", Path: "src/foo.js", Exported: true},
		},
	}

	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "src/foo.js", ChangeKind: ChangeModified},
		},
	}

	result := Analyze(scope, snap)

	if result.Summary == "" {
		t.Error("expected non-empty summary")
	}
}

// --- changescope tests ---

func TestChangeScopeFromPaths(t *testing.T) {
	t.Parallel()
	scope := ChangeScopeFromPaths([]string{"src/foo.js", "test/foo.test.js"}, ChangeModified)

	if len(scope.ChangedFiles) != 2 {
		t.Fatalf("expected 2 changed files, got %d", len(scope.ChangedFiles))
	}
	if scope.ChangedFiles[0].IsTestFile {
		t.Error("src/foo.js should not be a test file")
	}
	if !scope.ChangedFiles[1].IsTestFile {
		t.Error("test/foo.test.js should be a test file")
	}
	if scope.Source != "explicit" {
		t.Errorf("expected source 'explicit', got %q", scope.Source)
	}
}

func TestParseGitDiffOutput(t *testing.T) {
	t.Parallel()
	output := `M	src/auth/service.js
A	src/api/new.js
D	src/legacy/old.js
R100	src/old/name.js	src/new/name.js`

	scope := parseGitDiffOutput(output, "/repo")

	if len(scope.ChangedFiles) != 4 {
		t.Fatalf("expected 4 changed files, got %d", len(scope.ChangedFiles))
	}

	checks := []struct {
		idx  int
		kind ChangeKind
		path string
	}{
		{0, ChangeModified, "src/auth/service.js"},
		{1, ChangeAdded, "src/api/new.js"},
		{2, ChangeDeleted, "src/legacy/old.js"},
		{3, ChangeRenamed, "src/new/name.js"},
	}

	for _, c := range checks {
		if scope.ChangedFiles[c.idx].ChangeKind != c.kind {
			t.Errorf("file %d: expected kind %s, got %s", c.idx, c.kind, scope.ChangedFiles[c.idx].ChangeKind)
		}
		if scope.ChangedFiles[c.idx].Path != c.path {
			t.Errorf("file %d: expected path %s, got %s", c.idx, c.path, scope.ChangedFiles[c.idx].Path)
		}
	}
}

func TestParseGitDiffOutput_PathWithSpaces(t *testing.T) {
	t.Parallel()
	output := "M\tsrc/with space/file name.js"
	scope := parseGitDiffOutput(output, "/repo")
	if len(scope.ChangedFiles) != 1 {
		t.Fatalf("expected 1 changed file, got %d", len(scope.ChangedFiles))
	}
	if scope.ChangedFiles[0].Path != "src/with space/file name.js" {
		t.Fatalf("path = %q, want %q", scope.ChangedFiles[0].Path, "src/with space/file name.js")
	}
}

func TestChangeScopeFromGitDiff_DefaultsToWorkingTreeWhenHeadMinusOneMissing(t *testing.T) {
	t.Parallel()
	requireGit(t)
	repo := initImpactTestRepo(t)

	if err := os.WriteFile(filepath.Join(repo, "src", "app.js"), []byte("console.log('v2')\n"), 0o644); err != nil {
		t.Fatalf("write modified file: %v", err)
	}

	scope, err := ChangeScopeFromGitDiff(repo, "")
	if err != nil {
		t.Fatalf("ChangeScopeFromGitDiff returned error: %v", err)
	}
	if scope.Source != "git-diff-working-tree" {
		t.Fatalf("source = %q, want %q", scope.Source, "git-diff-working-tree")
	}
	if len(scope.ChangedFiles) != 1 {
		t.Fatalf("expected 1 changed file, got %d", len(scope.ChangedFiles))
	}
	if scope.ChangedFiles[0].Path != "src/app.js" {
		t.Fatalf("path = %q, want %q", scope.ChangedFiles[0].Path, "src/app.js")
	}
}

func TestChangeScopeFromGitDiff_InvalidBaseIncludesContext(t *testing.T) {
	t.Parallel()
	requireGit(t)
	repo := initImpactTestRepo(t)

	_, err := ChangeScopeFromGitDiff(repo, "DOES_NOT_EXIST")
	if err == nil {
		t.Fatal("expected error for invalid base ref")
	}
	if !strings.Contains(err.Error(), `against "DOES_NOT_EXIST"`) {
		t.Fatalf("expected contextual error, got: %v", err)
	}
}

func TestIsTestFilePath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		want bool
	}{
		{"src/auth/service.js", false},
		{"src/__tests__/auth.test.js", true},
		{"test/foo.test.js", true},
		{"test/res.type.js", true},            // top-level test/ dir, no .test. in name
		{"tests/integration/api.js", true},    // top-level tests/ dir
		{"src/foo.spec.js", true},
		{"internal/auth/auth_test.go", true},
		{"e2e/login.spec.js", true},           // top-level e2e/ dir
		{"src/utils/helpers.js", false},
		{"cypress/e2e/flow.cy.js", true},
	}

	for _, tt := range tests {
		got := isTestFilePath(tt.path)
		if got != tt.want {
			t.Errorf("isTestFilePath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

func initImpactTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.name", "Test User")
	runGit(t, dir, "config", "user.email", "test@example.com")

	srcDir := filepath.Join(dir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "app.js"), []byte("console.log('v1')\n"), 0o644); err != nil {
		t.Fatalf("write app.js: %v", err)
	}

	runGit(t, dir, "add", ".")
	runGit(t, dir, "commit", "-m", "initial")

	return dir
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(out))
	}
}

func TestSelectProtectiveTests_ExactFirst(t *testing.T) {
	t.Parallel()
	tests := []ImpactedTest{
		{Path: "a.test.js", ImpactConfidence: ConfidenceInferred},
		{Path: "b.test.js", ImpactConfidence: ConfidenceExact},
	}

	selected := selectProtectiveTests(tests, nil)

	if len(selected) != 1 {
		t.Fatalf("expected 1 selected test, got %d", len(selected))
	}
	if selected[0].Path != "b.test.js" {
		t.Errorf("expected exact test selected, got %s", selected[0].Path)
	}
}

func TestSelectProtectiveTests_FallbackToInferred(t *testing.T) {
	t.Parallel()
	tests := []ImpactedTest{
		{Path: "a.test.js", ImpactConfidence: ConfidenceInferred},
		{Path: "b.test.js", ImpactConfidence: ConfidenceInferred},
	}

	selected := selectProtectiveTests(tests, nil)

	if len(selected) != 2 {
		t.Errorf("expected 2 inferred tests selected, got %d", len(selected))
	}
}

// --- filter tests ---

func TestFilterByOwner(t *testing.T) {
	t.Parallel()
	result := &ImpactResult{
		Scope: ChangeScope{
			ChangedFiles: []ChangedFile{
				{Path: "src/auth/service.js", ChangeKind: ChangeModified},
				{Path: "src/payments/service.js", ChangeKind: ChangeModified},
			},
		},
		ImpactedUnits: []ImpactedCodeUnit{
			{UnitID: "auth:AuthService", Name: "AuthService", Path: "src/auth/service.js", Owner: "team-auth", Exported: true, ProtectionStatus: ProtectionStrong, CoveringTests: []string{"test/auth.test.js"}},
			{UnitID: "pay:PaymentService", Name: "PaymentService", Path: "src/payments/service.js", Owner: "team-payments", Exported: true, ProtectionStatus: ProtectionNone},
		},
		ImpactedTests: []ImpactedTest{
			{Path: "test/auth.test.js", ImpactConfidence: ConfidenceExact, Relevance: "covers impacted code unit"},
		},
		ProtectionGaps: []ProtectionGap{
			{GapType: "untested_export", CodeUnitID: "pay:PaymentService", Severity: "high"},
		},
		ImpactedOwners: []string{"team-auth", "team-payments"},
		Posture:        ChangeRiskPosture{Band: "partially_protected"},
	}

	filtered := FilterByOwner(result, "team-auth")

	if len(filtered.ImpactedUnits) != 1 {
		t.Fatalf("expected 1 filtered unit, got %d", len(filtered.ImpactedUnits))
	}
	if filtered.ImpactedUnits[0].Name != "AuthService" {
		t.Errorf("expected AuthService, got %s", filtered.ImpactedUnits[0].Name)
	}
	if len(filtered.ImpactedTests) != 1 {
		t.Errorf("expected 1 filtered test, got %d", len(filtered.ImpactedTests))
	}
	if len(filtered.ProtectionGaps) != 0 {
		t.Errorf("expected 0 filtered gaps, got %d", len(filtered.ProtectionGaps))
	}
}

// --- aggregate tests ---

func TestBuildAggregate_SparseData(t *testing.T) {
	t.Parallel()
	result := &ImpactResult{
		Scope: ChangeScope{
			ChangedFiles: []ChangedFile{
				{Path: "src/foo.js", ChangeKind: ChangeModified, IsTestFile: false},
				{Path: "test/foo.test.js", ChangeKind: ChangeModified, IsTestFile: true},
			},
		},
		ImpactedUnits: []ImpactedCodeUnit{
			{UnitID: "foo:Foo", Name: "Foo", Exported: true, ImpactConfidence: ConfidenceExact, ProtectionStatus: ProtectionStrong},
			{UnitID: "bar:Bar", Name: "Bar", Exported: false, ImpactConfidence: ConfidenceInferred, ProtectionStatus: ProtectionNone},
		},
		ImpactedTests: []ImpactedTest{
			{Path: "test/foo.test.js", ImpactConfidence: ConfidenceExact},
			{Path: "test/bar.test.js", ImpactConfidence: ConfidenceInferred},
		},
		SelectedTests: []ImpactedTest{
			{Path: "test/foo.test.js", ImpactConfidence: ConfidenceExact},
		},
		ProtectionGaps: []ProtectionGap{
			{Severity: "high", GapType: "untested_export"},
			{Severity: "medium", GapType: "no_coverage"},
		},
		ImpactedOwners: []string{"team-a"},
		Posture:        ChangeRiskPosture{Band: "partially_protected"},
	}

	agg := BuildAggregate(result)

	if agg.ChangedFileCount != 2 {
		t.Errorf("expected 2 changed files, got %d", agg.ChangedFileCount)
	}
	if agg.ChangedTestFileCount != 1 {
		t.Errorf("expected 1 changed test file, got %d", agg.ChangedTestFileCount)
	}
	if agg.ImpactedUnitCount != 2 {
		t.Errorf("expected 2 impacted units, got %d", agg.ImpactedUnitCount)
	}
	if agg.GapCount != 2 {
		t.Errorf("expected 2 gaps, got %d", agg.GapCount)
	}
	if agg.Posture != "partially_protected" {
		t.Errorf("expected partially_protected posture, got %s", agg.Posture)
	}
	// Below privacy threshold (3), breakdowns are suppressed.
	if !agg.IsSparse {
		t.Error("expected IsSparse=true for below-threshold data")
	}
	if len(agg.ProtectionCounts) != 0 {
		t.Errorf("expected suppressed protection counts, got %v", agg.ProtectionCounts)
	}
}

func TestBuildAggregate_AboveThreshold(t *testing.T) {
	t.Parallel()
	result := &ImpactResult{
		Scope: ChangeScope{
			ChangedFiles: []ChangedFile{
				{Path: "src/a.js", ChangeKind: ChangeModified},
				{Path: "src/b.js", ChangeKind: ChangeModified},
				{Path: "src/c.js", ChangeKind: ChangeModified},
				{Path: "test/a.test.js", ChangeKind: ChangeModified, IsTestFile: true},
			},
		},
		ImpactedUnits: []ImpactedCodeUnit{
			{UnitID: "a:A", Name: "A", Exported: true, ImpactConfidence: ConfidenceExact, ProtectionStatus: ProtectionStrong},
			{UnitID: "b:B", Name: "B", Exported: true, ImpactConfidence: ConfidenceExact, ProtectionStatus: ProtectionPartial},
			{UnitID: "c:C", Name: "C", Exported: false, ImpactConfidence: ConfidenceInferred, ProtectionStatus: ProtectionNone},
		},
		ImpactedTests: []ImpactedTest{
			{Path: "test/a.test.js", ImpactConfidence: ConfidenceExact},
		},
		SelectedTests: []ImpactedTest{
			{Path: "test/a.test.js", ImpactConfidence: ConfidenceExact},
		},
		ProtectionGaps: []ProtectionGap{
			{Severity: "high", GapType: "untested_export"},
		},
		ImpactedOwners: []string{"team-a", "team-b", "team-c"},
		Posture:        ChangeRiskPosture{Band: "weakly_protected"},
	}

	agg := BuildAggregate(result)

	if agg.ImpactedUnitCount != 3 {
		t.Errorf("expected 3 impacted units, got %d", agg.ImpactedUnitCount)
	}
	if agg.ExportedUnitCount != 2 {
		t.Errorf("expected 2 exported, got %d", agg.ExportedUnitCount)
	}
	// Above threshold — breakdowns are present.
	if agg.IsSparse {
		t.Error("expected IsSparse=false for above-threshold data")
	}
	if agg.ProtectionCounts["strong"] != 1 {
		t.Errorf("expected 1 strong, got %d", agg.ProtectionCounts["strong"])
	}
	if agg.ProtectionCounts["none"] != 1 {
		t.Errorf("expected 1 none, got %d", agg.ProtectionCounts["none"])
	}
	if agg.ConfidenceCounts["exact"] != 2 {
		t.Errorf("expected 2 exact, got %d", agg.ConfidenceCounts["exact"])
	}
	// Ratios should be computed.
	if agg.ProtectionRatio == 0 {
		t.Error("expected non-zero protection ratio")
	}
	if agg.ExactConfidenceRatio == 0 {
		t.Error("expected non-zero exact confidence ratio")
	}
}

// --- impact graph tests ---

func TestBuildImpactGraph_NilSnapshot(t *testing.T) {
	t.Parallel()
	g := BuildImpactGraph(nil)
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if g.Stats.TotalEdges != 0 {
		t.Errorf("expected 0 edges, got %d", g.Stats.TotalEdges)
	}
}

func TestBuildImpactGraph_LinkedCodeUnits(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/auth.test.js", Framework: "jest", LinkedCodeUnits: []string{"src/auth.js:AuthService"}},
			{Path: "test/db.test.js", Framework: "jest", LinkedCodeUnits: []string{"src/db.js:DBPool"}},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "AuthService", Path: "src/auth.js", Kind: "class", Exported: true},
			{Name: "DBPool", Path: "src/db.js", Kind: "class", Exported: true},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 2},
		},
	}

	g := BuildImpactGraph(snap)

	if g.Stats.TotalEdges != 2 {
		t.Errorf("expected 2 edges, got %d", g.Stats.TotalEdges)
	}
	if g.Stats.ConnectedUnits != 2 {
		t.Errorf("expected 2 connected units, got %d", g.Stats.ConnectedUnits)
	}
	if g.Stats.IsolatedUnits != 0 {
		t.Errorf("expected 0 isolated units, got %d", g.Stats.IsolatedUnits)
	}

	tests := g.TestsForUnit("src/auth.js:AuthService")
	if len(tests) != 1 || tests[0] != "test/auth.test.js" {
		t.Errorf("expected test/auth.test.js, got %v", tests)
	}

	units := g.UnitsForTest("test/db.test.js")
	if len(units) != 1 || units[0] != "src/db.js:DBPool" {
		t.Errorf("expected src/db.js:DBPool, got %v", units)
	}
}

func TestBuildImpactGraph_NameConvention(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/__tests__/AuthService.test.js", Framework: "jest"},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "AuthService", Path: "src/auth.js", Kind: "class", Exported: true},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 1},
		},
	}

	g := BuildImpactGraph(snap)

	if g.Stats.TotalEdges != 1 {
		t.Errorf("expected 1 name-convention edge, got %d", g.Stats.TotalEdges)
	}
	if g.Stats.WeakEdges != 1 {
		t.Errorf("expected 1 weak edge, got %d", g.Stats.WeakEdges)
	}
}

func TestBuildImpactGraph_EdgeBetween(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/a.test.js", Framework: "jest", LinkedCodeUnits: []string{"src/a.js:Foo"}},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "Foo", Path: "src/a.js", Kind: "function", Exported: true},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 1},
		},
	}

	g := BuildImpactGraph(snap)

	edge := g.EdgeBetween("src/a.js:Foo", "test/a.test.js")
	if edge == nil {
		t.Fatal("expected edge between unit and test")
	}
	if edge.Confidence != ConfidenceExact {
		t.Errorf("expected exact confidence, got %s", edge.Confidence)
	}

	noEdge := g.EdgeBetween("nonexistent", "test/a.test.js")
	if noEdge != nil {
		t.Error("expected nil for nonexistent edge")
	}
}

func TestBuildImpactGraph_DeterministicOutput(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/b.test.js", Framework: "jest", LinkedCodeUnits: []string{"src/b.js:B"}},
			{Path: "test/a.test.js", Framework: "jest", LinkedCodeUnits: []string{"src/a.js:A"}},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "B", Path: "src/b.js"},
			{Name: "A", Path: "src/a.js"},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 2},
		},
	}

	g1 := BuildImpactGraph(snap)
	g2 := BuildImpactGraph(snap)

	if len(g1.Edges) != len(g2.Edges) {
		t.Fatal("non-deterministic edge count")
	}
	for i := range g1.Edges {
		if g1.Edges[i].SourceID != g2.Edges[i].SourceID || g1.Edges[i].TargetID != g2.Edges[i].TargetID {
			t.Errorf("non-deterministic order at index %d", i)
		}
	}
}

// --- CI scope tests ---

func TestChangeScopeFromCIList(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		list            string
		root            string
		wantCount       int
		wantTestFileIdx int
		wantHasTestFile bool
	}{
		{
			name:            "standard list",
			list:            "src/auth.js\ntest/auth.test.js\nsrc/db.js\n",
			root:            "/repo",
			wantCount:       3,
			wantTestFileIdx: 1,
			wantHasTestFile: true,
		},
		{
			name:      "empty input",
			list:      "",
			root:      "/repo",
			wantCount: 0,
		},
		{
			name:      "whitespace lines are ignored",
			list:      "  src/foo.js  \n\n  \n  src/bar.js  \n",
			root:      "",
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			scope := ChangeScopeFromCIList(tt.list, tt.root)
			if len(scope.ChangedFiles) != tt.wantCount {
				t.Fatalf("changed files = %d, want %d", len(scope.ChangedFiles), tt.wantCount)
			}
			if scope.Source != "ci-changed-files" {
				t.Errorf("source = %q, want %q", scope.Source, "ci-changed-files")
			}
			for _, cf := range scope.ChangedFiles {
				if cf.ChangeKind != ChangeModified {
					t.Errorf("expected modified kind, got %s", cf.ChangeKind)
				}
			}
			if tt.wantHasTestFile {
				if tt.wantTestFileIdx >= len(scope.ChangedFiles) {
					t.Fatalf("test file index %d out of range", tt.wantTestFileIdx)
				}
				if !scope.ChangedFiles[tt.wantTestFileIdx].IsTestFile {
					t.Errorf("changed file at index %d should be a test file", tt.wantTestFileIdx)
				}
			}
		})
	}
}

// --- comparison scope tests ---

func TestChangeScopeFromComparison(t *testing.T) {
	t.Parallel()
	from := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/a.test.js"},
			{Path: "test/b.test.js"},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "A", Path: "src/a.js"},
		},
	}
	to := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/a.test.js"},
			{Path: "test/c.test.js"},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "A", Path: "src/a.js"},
			{Name: "D", Path: "src/d.js"},
		},
	}

	scope := ChangeScopeFromComparison(from, to)

	if scope.Source != "snapshot-compare" {
		t.Errorf("expected source snapshot-compare, got %s", scope.Source)
	}

	// Expect: src/d.js added, test/c.test.js added, test/b.test.js deleted,
	// src/a.js modified, test/a.test.js modified
	byPath := map[string]ChangeKind{}
	for _, cf := range scope.ChangedFiles {
		byPath[cf.Path] = cf.ChangeKind
	}

	if byPath["src/d.js"] != ChangeAdded {
		t.Errorf("expected src/d.js added, got %s", byPath["src/d.js"])
	}
	if byPath["test/c.test.js"] != ChangeAdded {
		t.Errorf("expected test/c.test.js added, got %s", byPath["test/c.test.js"])
	}
	if byPath["test/b.test.js"] != ChangeDeleted {
		t.Errorf("expected test/b.test.js deleted, got %s", byPath["test/b.test.js"])
	}
	if byPath["src/a.js"] != ChangeModified {
		t.Errorf("expected src/a.js modified, got %s", byPath["src/a.js"])
	}
}

func TestChangeScopeFromComparison_NilInputs(t *testing.T) {
	t.Parallel()
	scope := ChangeScopeFromComparison(nil, nil)
	if len(scope.ChangedFiles) != 0 {
		t.Errorf("expected 0 files for nil inputs, got %d", len(scope.ChangedFiles))
	}
}

// --- protective test set tests ---

func TestProtectiveSet_ExactStrategy(t *testing.T) {
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
	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "src/auth.js", ChangeKind: ChangeModified},
		},
	}

	result := Analyze(scope, snap)

	if result.ProtectiveSet == nil {
		t.Fatal("expected non-nil protective set")
	}
	if result.ProtectiveSet.SetKind != "exact" {
		t.Errorf("expected exact set kind, got %s", result.ProtectiveSet.SetKind)
	}
	if len(result.ProtectiveSet.Tests) == 0 {
		t.Fatal("expected at least 1 selected test")
	}
	if len(result.ProtectiveSet.Tests[0].Reasons) == 0 {
		t.Error("expected selection reasons")
	}
}

func TestProtectiveSet_FallbackBroad(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{}
	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "src/unknown.js", ChangeKind: ChangeModified},
		},
	}

	result := Analyze(scope, snap)

	if result.ProtectiveSet == nil {
		t.Fatal("expected non-nil protective set")
	}
	if result.ProtectiveSet.SetKind != "fallback_broad" {
		t.Errorf("expected fallback_broad, got %s", result.ProtectiveSet.SetKind)
	}
}

// --- instability dimension tests ---

func TestInstabilityDimension_WellProtected(t *testing.T) {
	t.Parallel()
	result := &ImpactResult{
		ImpactedUnits: []ImpactedCodeUnit{
			{ProtectionStatus: ProtectionStrong, Complexity: 2},
			{ProtectionStatus: ProtectionStrong, Complexity: 5},
		},
	}
	dim := computeInstabilityDimension(result)
	if dim.Band != "well_protected" {
		t.Errorf("expected well_protected, got %s", dim.Band)
	}
}

func TestInstabilityDimension_HighRisk(t *testing.T) {
	t.Parallel()
	result := &ImpactResult{
		ImpactedUnits: []ImpactedCodeUnit{
			{ProtectionStatus: ProtectionNone, Complexity: 15},
			{ProtectionStatus: ProtectionWeak, Complexity: 20},
		},
	}
	dim := computeInstabilityDimension(result)
	if dim.Band != "high_risk" {
		t.Errorf("expected high_risk, got %s", dim.Band)
	}
}

// --- evidence_limited posture tests ---

func TestEvidenceLimited_NoUnitsNoTests(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{}
	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "README.md", ChangeKind: ChangeModified},
		},
	}

	result := Analyze(scope, snap)

	if result.Posture.Band != "evidence_limited" {
		t.Errorf("expected evidence_limited, got %s", result.Posture.Band)
	}
}

func TestEvidenceLimited_AllWeakConfidence(t *testing.T) {
	t.Parallel()
	result := &ImpactResult{
		ImpactedUnits: []ImpactedCodeUnit{
			{UnitID: "a", ImpactConfidence: ConfidenceWeak, ProtectionStatus: ProtectionNone},
		},
	}
	if !isEvidenceLimited(result) {
		t.Error("expected evidence limited for all-weak units")
	}
}

// --- coverage diversity gap tests ---

func TestCoverageDiversityGap_E2EOnlyExport(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "e2e/auth.spec.js", Framework: "playwright", LinkedCodeUnits: []string{"AuthService"}},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "AuthService", Path: "src/auth.js", Kind: "class", Exported: true},
		},
		Frameworks: []models.Framework{
			{Name: "playwright", Type: models.FrameworkTypeE2E, FileCount: 1},
		},
	}
	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "src/auth.js", ChangeKind: ChangeModified},
		},
	}

	result := Analyze(scope, snap)

	foundDiversityGap := false
	for _, gap := range result.ProtectionGaps {
		if gap.GapType == "e2e_only_export" {
			foundDiversityGap = true
		}
	}
	if !foundDiversityGap {
		t.Error("expected e2e_only_export gap for exported unit with only E2E coverage")
	}
}

// --- extract test subject tests ---

func TestExtractTestSubject(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		want string
	}{
		{"src/__tests__/AuthService.test.js", "AuthService"},
		{"test/DBPool.spec.ts", "DBPool"},
		{"internal/auth/auth_test.go", "auth"},
		{"src/utils/helpers.js", ""},
		{"test/foo.test.tsx", "foo"},
	}

	for _, tt := range tests {
		got := extractTestSubject(tt.path)
		if got != tt.want {
			t.Errorf("extractTestSubject(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

// --- graph integration in Analyze ---

func TestAnalyze_GraphIsBuilt(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/a.test.js", Framework: "jest", LinkedCodeUnits: []string{"src/a.js:A"}},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "A", Path: "src/a.js", Kind: "function", Exported: true},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 1},
		},
	}
	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "src/a.js", ChangeKind: ChangeModified},
		},
	}

	result := Analyze(scope, snap)

	if result.Graph == nil {
		t.Fatal("expected graph to be built")
	}
	if result.Graph.Stats.TotalEdges == 0 {
		t.Error("expected graph to have edges")
	}
}

// --- coverage type info tests ---

func TestCoverageTypeInfo_MixedCoverage(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/unit.test.js", Framework: "jest", LinkedCodeUnits: []string{"AuthService"}},
			{Path: "e2e/auth.spec.js", Framework: "playwright", LinkedCodeUnits: []string{"AuthService"}},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "AuthService", Path: "src/auth.js", Kind: "class", Exported: true},
		},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 1},
			{Name: "playwright", Type: models.FrameworkTypeE2E, FileCount: 1},
		},
	}
	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "src/auth.js", ChangeKind: ChangeModified},
		},
	}

	result := Analyze(scope, snap)

	if len(result.ImpactedUnits) == 0 {
		t.Fatal("expected impacted units")
	}
	ct := result.ImpactedUnits[0].CoverageTypes
	if ct == nil {
		t.Fatal("expected coverage type info")
	}
	if !ct.HasUnitCoverage {
		t.Error("expected unit coverage")
	}
	if !ct.HasE2ECoverage {
		t.Error("expected E2E coverage")
	}
}

// --- summary format tests ---

func TestImpactSummary_ContainsPosture(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		CodeUnits: []models.CodeUnit{
			{Name: "Foo", Path: "src/foo.js", Exported: true},
		},
	}
	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "src/foo.js", ChangeKind: ChangeModified},
		},
	}

	result := Analyze(scope, snap)

	if !strings.Contains(result.Summary, "Posture:") {
		t.Error("expected summary to contain posture band")
	}
}

// --- non-testable file filtering tests ---

func TestAnalyze_NonTestableFilesExcluded(t *testing.T) {
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

	scope := &ChangeScope{
		ChangedFiles: []ChangedFile{
			{Path: "src/auth.js", ChangeKind: ChangeModified},
			{Path: "README.md", ChangeKind: ChangeModified},
			{Path: ".github/workflows/ci.yml", ChangeKind: ChangeModified},
			{Path: "CLAUDE.md", ChangeKind: ChangeModified},
			{Path: "Makefile", ChangeKind: ChangeModified},
			{Path: ".goreleaser.yaml", ChangeKind: ChangeModified},
			{Path: "package.json", ChangeKind: ChangeModified},
		},
	}

	result := Analyze(scope, snap)

	// Only src/auth.js should produce impacted units — all non-code files filtered.
	if len(result.ImpactedUnits) != 1 {
		t.Errorf("expected 1 impacted unit (AuthService only), got %d", len(result.ImpactedUnits))
		for _, iu := range result.ImpactedUnits {
			t.Logf("  unit: %s (%s)", iu.Name, iu.Path)
		}
	}

	// No protection gaps for non-code files.
	for _, gap := range result.ProtectionGaps {
		for _, ext := range []string{".md", ".yml", ".yaml", ".json"} {
			if strings.HasSuffix(gap.Path, ext) {
				t.Errorf("unexpected protection gap for non-code file: %s", gap.Path)
			}
		}
	}
}

func TestIsAnalyzableSourceFile_Exported(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		want bool
	}{
		{"src/auth.js", true},
		{"internal/foo/bar.go", true},
		{"src/app.tsx", true},
		{"src/utils.mjs", true},
		{"lib/core.py", true},
		{"Main.java", true},
		{"README.md", false},
		{".github/workflows/ci.yml", false},
		{"config.yaml", false},
		{"package.json", false},
		{"Makefile", false},
		{".goreleaser.yaml", false},
		{"docs/guide.md", false},
		{"tsconfig.json", false},
		{".eslintrc.json", false},
		{"Dockerfile", false},
		{"LICENSE", false},
	}

	for _, tt := range tests {
		got := IsAnalyzableSourceFile(tt.path)
		if got != tt.want {
			t.Errorf("IsAnalyzableSourceFile(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestApplyManualCoverageOverlay_AnnotatesGaps(t *testing.T) {
	t.Parallel()
	result := &ImpactResult{
		ProtectionGaps: []ProtectionGap{
			{Path: "billing-core/payment.go", GapType: "no_test_coverage", Severity: "high"},
			{Path: "auth/login.go", GapType: "no_test_coverage", Severity: "medium"},
		},
	}

	artifacts := []models.ManualCoverageArtifact{
		{ArtifactID: "manual:testrail:abc", Name: "billing regression", Area: "billing-core", Source: "testrail", Criticality: "high"},
	}

	result.ApplyManualCoverageOverlay(artifacts)

	if len(result.PolicyNotes) == 0 {
		t.Error("expected policy notes for billing-core gap")
	}
	found := false
	for _, note := range result.PolicyNotes {
		if strings.Contains(note, "billing regression") && strings.Contains(note, "billing-core") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected note about billing-core manual coverage, got: %v", result.PolicyNotes)
	}
}

func TestApplyManualCoverageOverlay_NoGaps(t *testing.T) {
	t.Parallel()
	result := &ImpactResult{}
	artifacts := []models.ManualCoverageArtifact{
		{ArtifactID: "manual:testrail:abc", Name: "billing regression", Area: "billing-core", Source: "testrail", Criticality: "high"},
	}

	result.ApplyManualCoverageOverlay(artifacts)

	if len(result.PolicyNotes) != 0 {
		t.Error("expected no policy notes when no protection gaps exist")
	}
}

func TestApplyManualCoverageOverlay_NoArtifacts(t *testing.T) {
	t.Parallel()
	result := &ImpactResult{
		ProtectionGaps: []ProtectionGap{
			{Path: "billing-core/payment.go", GapType: "no_test_coverage"},
		},
	}

	result.ApplyManualCoverageOverlay(nil)

	if len(result.PolicyNotes) != 0 {
		t.Error("expected no policy notes when no artifacts")
	}
}

func TestMatchesArea(t *testing.T) {
	t.Parallel()
	tests := []struct {
		filePath string
		area     string
		want     bool
	}{
		{"billing-core/payment.go", "billing-core", true},
		{"billing-core/sub/deep.go", "billing-core", true},
		{"auth/login.go", "billing-core", false},
		{"checkout/cart.go", "checkout/*", true},
		{"checkout/cart.go", "checkout/", true},
		{"other/file.go", "checkout/*", false},
	}

	for _, tt := range tests {
		got := matchesArea(tt.filePath, tt.area)
		if got != tt.want {
			t.Errorf("matchesArea(%q, %q) = %v, want %v", tt.filePath, tt.area, got, tt.want)
		}
	}
}
