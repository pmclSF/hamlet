package changescope

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pmclSF/hamlet/internal/impact"
	"github.com/pmclSF/hamlet/internal/models"
)

func testSnapshot() *models.TestSuiteSnapshot {
	return &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{Name: "test-repo"},
		Frameworks: []models.Framework{
			{Name: "jest", Type: models.FrameworkTypeUnit, FileCount: 4, TestCount: 20},
		},
		TestFiles: []models.TestFile{
			{Path: "src/__tests__/auth.test.js", Framework: "jest", TestCount: 5, AssertionCount: 8, LinkedCodeUnits: []string{"AuthService"}},
			{Path: "src/__tests__/user.test.js", Framework: "jest", TestCount: 4, AssertionCount: 6, LinkedCodeUnits: []string{"UserService"}},
			{Path: "src/__tests__/payment.test.js", Framework: "jest", TestCount: 5, AssertionCount: 7, LinkedCodeUnits: []string{"PaymentProcessor"}},
			{Path: "src/__tests__/config.test.js", Framework: "jest", TestCount: 3, AssertionCount: 4},
		},
		CodeUnits: []models.CodeUnit{
			{Name: "AuthService", Path: "src/auth.js", Kind: models.CodeUnitKindClass, Exported: true},
			{Name: "UserService", Path: "src/user.js", Kind: models.CodeUnitKindClass, Exported: true},
			{Name: "PaymentProcessor", Path: "src/payment.js", Kind: models.CodeUnitKindClass, Exported: true},
			{Name: "helperFn", Path: "src/utils.js", Kind: models.CodeUnitKindFunction, Exported: false},
		},
		Ownership: map[string][]string{
			"src/auth.js":    {"team-platform"},
			"src/payment.js": {"team-payments"},
		},
	}
}

func TestAnalyzePR_NarrowChange(t *testing.T) {
	t.Parallel()
	snap := testSnapshot()
	scope := impact.ChangeScopeFromPaths([]string{"src/auth.js"}, impact.ChangeModified)

	pr := AnalyzePR(scope, snap)

	if pr.ChangedFileCount != 1 {
		t.Errorf("changed file count = %d, want 1", pr.ChangedFileCount)
	}
	if pr.ChangedSourceCount != 1 {
		t.Errorf("changed source count = %d, want 1", pr.ChangedSourceCount)
	}
	if pr.ChangedTestCount != 0 {
		t.Errorf("changed test count = %d, want 0", pr.ChangedTestCount)
	}
	if pr.PostureBand == "" {
		t.Error("expected non-empty posture band")
	}
	if pr.Summary == "" {
		t.Error("expected non-empty summary")
	}
	if pr.ImpactedUnitCount == 0 {
		t.Error("expected impacted units for auth.js change")
	}
}

func TestAnalyzePR_BroadChange(t *testing.T) {
	t.Parallel()
	snap := testSnapshot()
	scope := impact.ChangeScopeFromPaths(
		[]string{"src/auth.js", "src/payment.js", "src/utils.js"},
		impact.ChangeModified,
	)

	pr := AnalyzePR(scope, snap)

	if pr.ChangedSourceCount != 3 {
		t.Errorf("changed source count = %d, want 3", pr.ChangedSourceCount)
	}
	if len(pr.AffectedOwners) == 0 {
		t.Error("expected affected owners")
	}
}

func TestAnalyzePR_WithTestChanges(t *testing.T) {
	t.Parallel()
	snap := testSnapshot()
	scope := impact.ChangeScopeFromPaths(
		[]string{"src/auth.js", "src/__tests__/auth.test.js"},
		impact.ChangeModified,
	)

	pr := AnalyzePR(scope, snap)

	if pr.ChangedTestCount != 1 {
		t.Errorf("changed test count = %d, want 1", pr.ChangedTestCount)
	}
	if pr.ChangedSourceCount != 1 {
		t.Errorf("changed source count = %d, want 1", pr.ChangedSourceCount)
	}
}

func TestAnalyzePR_UntestedExport(t *testing.T) {
	t.Parallel()
	snap := testSnapshot()
	// Add a new exported function with no test coverage.
	snap.CodeUnits = append(snap.CodeUnits, models.CodeUnit{
		Name: "NewFeature", Path: "src/feature.js",
		Kind: models.CodeUnitKindFunction, Exported: true,
	})
	scope := impact.ChangeScopeFromPaths([]string{"src/feature.js"}, impact.ChangeAdded)

	pr := AnalyzePR(scope, snap)

	hasUntestedExport := false
	for _, f := range pr.NewFindings {
		if f.Type == "untested_export_in_change" {
			hasUntestedExport = true
		}
	}
	if !hasUntestedExport {
		t.Error("expected untested_export_in_change finding")
	}
}

func TestAnalyzeChangedPaths(t *testing.T) {
	t.Parallel()
	snap := testSnapshot()
	pr := AnalyzeChangedPaths([]string{"src/auth.js"}, impact.ChangeModified, snap)

	if pr == nil {
		t.Fatal("expected non-nil PRAnalysis")
	}
	if pr.ChangedFileCount != 1 {
		t.Errorf("changed file count = %d, want 1", pr.ChangedFileCount)
	}
}

func TestRenderPRSummaryMarkdown(t *testing.T) {
	t.Parallel()
	snap := testSnapshot()
	scope := impact.ChangeScopeFromPaths([]string{"src/auth.js"}, impact.ChangeModified)
	pr := AnalyzePR(scope, snap)

	var buf bytes.Buffer
	RenderPRSummaryMarkdown(&buf, pr)
	output := buf.String()

	if !strings.Contains(output, "## Hamlet") {
		t.Error("expected markdown header")
	}
	if !strings.Contains(output, "Posture") {
		t.Error("expected posture in output")
	}
	if !strings.Contains(output, "Changed files") {
		t.Error("expected stats table")
	}
}

func TestRenderPRCommentConcise(t *testing.T) {
	t.Parallel()
	snap := testSnapshot()
	scope := impact.ChangeScopeFromPaths([]string{"src/auth.js"}, impact.ChangeModified)
	pr := AnalyzePR(scope, snap)

	var buf bytes.Buffer
	RenderPRCommentConcise(&buf, pr)
	output := buf.String()

	if !strings.Contains(output, "Hamlet") {
		t.Error("expected Hamlet in concise comment")
	}
}

func TestRenderCIAnnotation(t *testing.T) {
	t.Parallel()
	snap := testSnapshot()
	// Add a new untested export to generate findings.
	snap.CodeUnits = append(snap.CodeUnits, models.CodeUnit{
		Name: "UntestedFn", Path: "src/untested.js",
		Kind: models.CodeUnitKindFunction, Exported: true,
	})
	scope := impact.ChangeScopeFromPaths([]string{"src/untested.js"}, impact.ChangeAdded)
	pr := AnalyzePR(scope, snap)

	var buf bytes.Buffer
	RenderCIAnnotation(&buf, pr)
	output := buf.String()

	if len(pr.NewFindings) > 0 && !strings.Contains(output, "::") {
		t.Error("expected CI annotation format")
	}
}

func TestRenderChangeScopedReport(t *testing.T) {
	t.Parallel()
	snap := testSnapshot()
	scope := impact.ChangeScopeFromPaths([]string{"src/auth.js", "src/payment.js"}, impact.ChangeModified)
	pr := AnalyzePR(scope, snap)

	var buf bytes.Buffer
	RenderChangeScopedReport(&buf, pr)
	output := buf.String()

	if !strings.Contains(output, "Change-Scoped Analysis") {
		t.Error("expected report header")
	}
	if !strings.Contains(output, "Posture:") {
		t.Error("expected posture line")
	}
}

func TestPostureBadge(t *testing.T) {
	t.Parallel()
	tests := []struct {
		band string
		want string
	}{
		{"well_protected", "[PASS]"},
		{"partially_protected", "[WARN]"},
		{"weakly_protected", "[RISK]"},
		{"high_risk", "[FAIL]"},
		{"unknown", "[????]"},
	}
	for _, tt := range tests {
		got := postureBadge(tt.band)
		if got != tt.want {
			t.Errorf("postureBadge(%q) = %q, want %q", tt.band, got, tt.want)
		}
	}
}
