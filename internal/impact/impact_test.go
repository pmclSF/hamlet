package impact

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestAnalyze_WellProtectedChange(t *testing.T) {
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

func TestAnalyze_DirectlyChangedTest(t *testing.T) {
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

func TestAnalyze_FileWithNoCodeUnits(t *testing.T) {
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

func TestIsTestFilePath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"src/auth/service.js", false},
		{"src/__tests__/auth.test.js", true},
		{"test/foo.test.js", true},
		{"src/foo.spec.js", true},
		{"internal/auth/auth_test.go", true},
		{"e2e/login.spec.js", true},
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

func TestSelectProtectiveTests_ExactFirst(t *testing.T) {
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

func TestBuildAggregate(t *testing.T) {
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
	if agg.ExportedUnitCount != 1 {
		t.Errorf("expected 1 exported unit, got %d", agg.ExportedUnitCount)
	}
	if agg.ProtectionCounts["strong"] != 1 {
		t.Errorf("expected 1 strong protection, got %d", agg.ProtectionCounts["strong"])
	}
	if agg.ProtectionCounts["none"] != 1 {
		t.Errorf("expected 1 none protection, got %d", agg.ProtectionCounts["none"])
	}
	if agg.GapCount != 2 {
		t.Errorf("expected 2 gaps, got %d", agg.GapCount)
	}
	if agg.HighSeverityGapCount != 1 {
		t.Errorf("expected 1 high severity gap, got %d", agg.HighSeverityGapCount)
	}
	if agg.SelectedTestCount != 1 {
		t.Errorf("expected 1 selected test, got %d", agg.SelectedTestCount)
	}
	if agg.OwnerCount != 1 {
		t.Errorf("expected 1 owner, got %d", agg.OwnerCount)
	}
	if agg.Posture != "partially_protected" {
		t.Errorf("expected partially_protected posture, got %s", agg.Posture)
	}
	if agg.ConfidenceCounts["exact"] != 1 {
		t.Errorf("expected 1 exact confidence, got %d", agg.ConfidenceCounts["exact"])
	}
}
