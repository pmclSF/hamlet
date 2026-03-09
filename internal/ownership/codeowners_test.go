package ownership

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCodeownersFile(t *testing.T) {
	dir := t.TempDir()
	coPath := filepath.Join(dir, "CODEOWNERS")
	content := `# Global owners
* @org/global

# Auth team
/src/auth/ @team-auth @team-security

# Payments
/src/payments/ @team-payments

# JS files
*.js @team-frontend

# Docs (single level)
docs/* @team-docs

# Nested test dirs
**/test/ @team-testing

# Malformed line (no owner)
/orphan/

# Pattern with no comment
/src/config/ @team-platform
`
	if err := os.WriteFile(coPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cf := ParseCodeownersFile(coPath, "CODEOWNERS")

	if cf.Path != "CODEOWNERS" {
		t.Errorf("path = %q, want %q", cf.Path, "CODEOWNERS")
	}

	// Should have parsed rules (excluding malformed line).
	if len(cf.Rules) != 7 {
		t.Errorf("got %d rules, want 7", len(cf.Rules))
		for i, r := range cf.Rules {
			t.Logf("  rule[%d]: pattern=%q owners=%v line=%d", i, r.Pattern, r.Owners, r.LineNumber)
		}
	}

	// Check owner normalization (@ stripped).
	if len(cf.Rules) > 0 {
		if cf.Rules[0].Owners[0] != "org/global" {
			t.Errorf("first rule owner = %q, want %q", cf.Rules[0].Owners[0], "org/global")
		}
	}

	// Multi-owner rule.
	if len(cf.Rules) > 1 {
		if len(cf.Rules[1].Owners) != 2 {
			t.Errorf("auth rule has %d owners, want 2", len(cf.Rules[1].Owners))
		}
	}

	// Should have diagnostic for malformed line.
	hasMalformedDiag := false
	for _, d := range cf.Diagnostics {
		if d.Level == "warning" && d.Line > 0 {
			hasMalformedDiag = true
		}
	}
	if !hasMalformedDiag {
		t.Error("expected diagnostic for malformed CODEOWNERS line")
	}
}

func TestMatchCodeowners_Precedence(t *testing.T) {
	rules := []CodeownersRule{
		{Pattern: "*", Owners: []string{"global"}, LineNumber: 1},
		{Pattern: "/src/", Owners: []string{"team-src"}, LineNumber: 2},
		{Pattern: "/src/auth/", Owners: []string{"team-auth"}, LineNumber: 3},
	}

	tests := []struct {
		path      string
		wantOwner string
		wantMatch bool
	}{
		{"README.md", "global", true},
		{"src/util.js", "team-src", true},
		{"src/auth/login.js", "team-auth", true},
		{"src/auth/deep/nested.js", "team-auth", true},
	}

	for _, tt := range tests {
		rule, matched := MatchCodeowners(rules, tt.path)
		if matched != tt.wantMatch {
			t.Errorf("MatchCodeowners(%q) matched=%v, want %v", tt.path, matched, tt.wantMatch)
			continue
		}
		if matched && rule.Owners[0] != tt.wantOwner {
			t.Errorf("MatchCodeowners(%q) owner=%q, want %q", tt.path, rule.Owners[0], tt.wantOwner)
		}
	}
}

func TestMatchCodeowners_WildcardExtension(t *testing.T) {
	rules := []CodeownersRule{
		{Pattern: "*.js", Owners: []string{"team-js"}, LineNumber: 1},
		{Pattern: "*.go", Owners: []string{"team-go"}, LineNumber: 2},
	}

	tests := []struct {
		path      string
		wantOwner string
		wantMatch bool
	}{
		{"src/app.js", "team-js", true},
		{"deep/nested/file.js", "team-js", true},
		{"main.go", "team-go", true},
		{"README.md", "", false},
	}

	for _, tt := range tests {
		rule, matched := MatchCodeowners(rules, tt.path)
		if matched != tt.wantMatch {
			t.Errorf("path=%q matched=%v, want %v", tt.path, matched, tt.wantMatch)
			continue
		}
		if matched && rule.Owners[0] != tt.wantOwner {
			t.Errorf("path=%q owner=%q, want %q", tt.path, rule.Owners[0], tt.wantOwner)
		}
	}
}

func TestMatchCodeowners_DoubleStarPattern(t *testing.T) {
	rules := []CodeownersRule{
		{Pattern: "**/test/", Owners: []string{"team-testing"}, LineNumber: 1},
	}

	tests := []struct {
		path      string
		wantMatch bool
	}{
		{"test/unit.js", true},
		{"src/test/helper.js", true},
		{"deep/nested/test/file.js", true},
		{"testing/other.js", false},
	}

	for _, tt := range tests {
		_, matched := MatchCodeowners(rules, tt.path)
		if matched != tt.wantMatch {
			t.Errorf("path=%q matched=%v, want %v", tt.path, matched, tt.wantMatch)
		}
	}
}

func TestMatchCodeowners_SingleLevelWildcard(t *testing.T) {
	rules := []CodeownersRule{
		{Pattern: "docs/*", Owners: []string{"team-docs"}, LineNumber: 1},
	}

	tests := []struct {
		path      string
		wantMatch bool
	}{
		{"docs/readme.md", true},
		{"docs/guide.txt", true},
		{"docs/sub/nested.md", false}, // single level only
	}

	for _, tt := range tests {
		_, matched := MatchCodeowners(rules, tt.path)
		if matched != tt.wantMatch {
			t.Errorf("path=%q matched=%v, want %v", tt.path, matched, tt.wantMatch)
		}
	}
}

func TestMatchCodeowners_NoRules(t *testing.T) {
	_, matched := MatchCodeowners(nil, "any/file.js")
	if matched {
		t.Error("empty rules should not match")
	}
}

func TestCodeownersRule_ToAssignment(t *testing.T) {
	rule := CodeownersRule{
		Pattern:    "/src/auth/",
		Owners:     []string{"team-auth", "team-security"},
		LineNumber: 5,
	}

	a := rule.ToAssignment(".github/CODEOWNERS")
	if len(a.Owners) != 2 {
		t.Fatalf("got %d owners, want 2", len(a.Owners))
	}
	if a.Owners[0].ID != "team-auth" {
		t.Errorf("owner[0] = %q, want %q", a.Owners[0].ID, "team-auth")
	}
	if a.Source != SourceCodeowners {
		t.Errorf("source = %q, want %q", a.Source, SourceCodeowners)
	}
	if a.Confidence != ConfidenceHigh {
		t.Errorf("confidence = %q, want %q", a.Confidence, ConfidenceHigh)
	}
	if a.MatchedRule != "/src/auth/" {
		t.Errorf("matchedRule = %q, want %q", a.MatchedRule, "/src/auth/")
	}
	if a.SourceFile != ".github/CODEOWNERS" {
		t.Errorf("sourceFile = %q, want %q", a.SourceFile, ".github/CODEOWNERS")
	}
}

func TestFindCodeownersFile(t *testing.T) {
	// Test with .github/CODEOWNERS
	dir := t.TempDir()
	ghDir := filepath.Join(dir, ".github")
	os.MkdirAll(ghDir, 0755)
	os.WriteFile(filepath.Join(ghDir, "CODEOWNERS"), []byte("* @owner"), 0644)

	_, relPath, found := FindCodeownersFile(dir)
	if !found {
		t.Fatal("expected to find CODEOWNERS")
	}
	if relPath != filepath.Join(".github", "CODEOWNERS") {
		t.Errorf("relPath = %q, want %q", relPath, filepath.Join(".github", "CODEOWNERS"))
	}

	// Test with no CODEOWNERS
	emptyDir := t.TempDir()
	_, _, found = FindCodeownersFile(emptyDir)
	if found {
		t.Error("should not find CODEOWNERS in empty dir")
	}
}
