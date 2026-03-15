package benchmark

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// --- Config tests ---

func TestLoadBenchmarkRepos(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "repos.json")
	content := `{
		"repos": [
			{"name": "test-repo", "path": "/tmp/test", "type": "fixture"},
			{"name": "real-repo", "path": "/opt/code", "type": "real", "description": "A real repo"}
		]
	}`
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	repos, err := LoadBenchmarkRepos(cfgPath)
	if err != nil {
		t.Fatalf("LoadBenchmarkRepos error: %v", err)
	}

	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}
	if repos[0].Name != "test-repo" {
		t.Errorf("expected name test-repo, got %s", repos[0].Name)
	}
	if repos[1].Type != "real" {
		t.Errorf("expected type real, got %s", repos[1].Type)
	}
	if repos[1].Description != "A real repo" {
		t.Errorf("expected description, got %s", repos[1].Description)
	}
}

func TestLoadBenchmarkRepos_InvalidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(cfgPath, []byte("{invalid}"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadBenchmarkRepos(cfgPath)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadBenchmarkRepos_MissingFile(t *testing.T) {
	t.Parallel()
	_, err := LoadBenchmarkRepos("/nonexistent/path/repos.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// --- Repo loader tests ---

func TestResolveRepo(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := Repo{
		Name:        "test",
		Path:        dir,
		Type:        "fixture",
		Description: "test repo",
	}

	meta, err := ResolveRepo(cfg, "/unused")
	if err != nil {
		t.Fatalf("ResolveRepo error: %v", err)
	}
	if meta.Name != "test" {
		t.Errorf("expected name test, got %s", meta.Name)
	}
	if meta.IsGitRepo {
		t.Error("temp dir should not be detected as git repo")
	}

	// Create .git to simulate git repo.
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	meta, err = ResolveRepo(cfg, "/unused")
	if err != nil {
		t.Fatal(err)
	}
	if !meta.IsGitRepo {
		t.Error("expected git repo detection after .git created")
	}
}

func TestResolveRepo_RelativePath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	subDir := filepath.Join(dir, "myrepo")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfg := Repo{
		Name: "relative",
		Path: "myrepo",
		Type: "fixture",
	}

	meta, err := ResolveRepo(cfg, dir)
	if err != nil {
		t.Fatalf("ResolveRepo error: %v", err)
	}
	if meta.AbsPath != subDir {
		t.Errorf("expected abs path %s, got %s", subDir, meta.AbsPath)
	}
}

func TestResolveRepo_MissingPath(t *testing.T) {
	t.Parallel()

	cfg := Repo{
		Name: "missing",
		Path: "/nonexistent/path/to/repo",
		Type: "fixture",
	}

	_, err := ResolveRepo(cfg, "/tmp")
	if err == nil {
		t.Error("expected error for missing repo path")
	}
}

func TestDetectLanguages(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	srcDir := filepath.Join(dir, "src")
	if err := os.Mkdir(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a.js", "b.js", "c.ts", "d.ts"} {
		if err := os.WriteFile(filepath.Join(srcDir, name), []byte("//"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	langs := detectLanguages(dir)
	found := map[string]bool{}
	for _, l := range langs {
		found[l] = true
	}

	if !found["javascript"] {
		t.Error("expected javascript detected")
	}
	if !found["typescript"] {
		t.Error("expected typescript detected")
	}
}

func TestDiscoverRepos(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	for _, name := range []string{"repo-a", "repo-b"} {
		if err := os.Mkdir(filepath.Join(dir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// Create a file (should be ignored).
	if err := os.WriteFile(filepath.Join(dir, "readme.md"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	repos, err := DiscoverRepos(dir)
	if err != nil {
		t.Fatalf("DiscoverRepos error: %v", err)
	}

	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}
}

// --- Assessment tests ---

func TestAssessCommand_Success(t *testing.T) {
	t.Parallel()

	cr := CommandResult{
		Command:   "analyze",
		RepoName:  "test-repo",
		ExitCode:  0,
		RuntimeMs: 1500,
		Stdout:    `{"testCases": [{"testId": "t1"}], "repoProfile": {"volume": "medium"}, "coverageConfidence": "high", "duplicateTests": 3, "highFanout": [{"node": "a"}], "weakCoverage": []}`,
	}

	a := AssessCommand(cr)

	if !a.Success {
		t.Error("expected success")
	}
	if !a.OutputNonEmpty {
		t.Error("expected non-empty output")
	}
	if !a.ParsedJSON {
		t.Error("expected JSON parsed")
	}
	if a.CredibilityScore <= 0 {
		t.Errorf("expected positive credibility, got %d", a.CredibilityScore)
	}
	if len(a.ExpectedSections) < 3 {
		t.Errorf("expected at least 3 found sections, got %d: %v", len(a.ExpectedSections), a.ExpectedSections)
	}
}

func TestAssessCommand_Failure(t *testing.T) {
	t.Parallel()

	cr := CommandResult{
		Command:   "analyze",
		RepoName:  "test-repo",
		ExitCode:  1,
		RuntimeMs: 50,
		Stderr:    "panic: something went wrong",
	}

	a := AssessCommand(cr)

	if a.Success {
		t.Error("expected failure")
	}
	if a.CredibilityScore > 0 {
		t.Errorf("expected zero or negative credibility for failed command, got %d", a.CredibilityScore)
	}
	if len(a.WarningFlags) == 0 {
		t.Error("expected warning flags for stderr output")
	}
}

func TestAssessCommand_EmptyOutput(t *testing.T) {
	t.Parallel()

	cr := CommandResult{
		Command:  "analyze",
		RepoName: "test-repo",
		ExitCode: 0,
		Stdout:   "",
	}

	a := AssessCommand(cr)

	if a.OutputNonEmpty {
		t.Error("expected empty output detection")
	}

	found := false
	for _, note := range a.Notes {
		if strings.Contains(note, "empty") {
			found = true
		}
	}
	if !found {
		t.Error("expected note about empty output")
	}
}

func TestAssessCommand_ImpactSections(t *testing.T) {
	t.Parallel()

	cr := CommandResult{
		Command:   "impact",
		RepoName:  "test-repo",
		ExitCode:  0,
		RuntimeMs: 800,
		Stdout:    `{"changedFiles": ["src/auth.js"], "impactedTests": [{"id": "t1", "confidence": 0.9}], "risk": "medium", "reason": "direct dependency"}`,
	}

	a := AssessCommand(cr)

	if len(a.ExpectedSections) < 3 {
		t.Errorf("expected at least 3 impact sections found, got %d: %v",
			len(a.ExpectedSections), a.ExpectedSections)
	}
}

func TestAssessCommand_CredibilityClamp(t *testing.T) {
	t.Parallel()

	cr := CommandResult{
		Command:  "analyze",
		ExitCode: 1,
		Stderr:   "error\nerror\nerror\nerror\nerror\nerror\nerror\nerror\nerror\nerror\nerror",
		Error:    "crash",
	}

	a := AssessCommand(cr)

	if a.CredibilityScore < 0 {
		t.Errorf("credibility should be clamped to 0, got %d", a.CredibilityScore)
	}
}

func TestAssessResults(t *testing.T) {
	t.Parallel()

	br := BenchResult{
		Repo: RepoMeta{Name: "test-repo", Type: "fixture"},
		Commands: []CommandResult{
			{Command: "analyze", RepoName: "test-repo", ExitCode: 0, Stdout: `{"tests": 5}`},
			{Command: "impact", RepoName: "test-repo", ExitCode: 0, Stdout: `{"impacted": 3}`},
		},
	}

	ra := AssessResults(br)
	if ra.Repo.Name != "test-repo" {
		t.Errorf("expected repo name test-repo, got %s", ra.Repo.Name)
	}
	if len(ra.Assessments) != 2 {
		t.Fatalf("expected 2 assessments, got %d", len(ra.Assessments))
	}
	if ra.OverallScore <= 0 {
		t.Errorf("expected positive overall score, got %d", ra.OverallScore)
	}
}

// --- Report writer tests ---

func TestGenerateSummary(t *testing.T) {
	t.Parallel()

	assessments := []RepoAssessment{
		{
			Repo: RepoMeta{Name: "test-repo", Type: "fixture", IsGitRepo: true, Languages: []string{"javascript"}},
			Assessments: []CommandAssessment{
				{
					Command:          "analyze",
					Success:          true,
					RuntimeMs:        1500,
					CredibilityScore: 80,
					ExpectedSections: []string{"tests detected", "repo profile"},
					MissingSections:  []string{"weak coverage"},
					Notes:            []string{"command succeeded", "tests detected"},
				},
				{
					Command:          "impact",
					Success:          true,
					RuntimeMs:        800,
					CredibilityScore: 70,
					ExpectedSections: []string{"changed files"},
					Notes:            []string{"command succeeded"},
				},
			},
			OverallScore: 75,
		},
	}

	md := GenerateSummary(assessments)

	checks := []string{
		"# Terrain CLI Benchmark Summary",
		"## Repo: test-repo",
		"### analyze",
		"### impact",
		"## Cross-Repo Summary",
		"Command Strength Ranking",
		"Frequently Missing Sections",
		"weak coverage",
	}
	for _, c := range checks {
		if !strings.Contains(md, c) {
			t.Errorf("missing expected content: %q", c)
		}
	}
}

func TestGenerateSummary_Empty(t *testing.T) {
	t.Parallel()

	md := GenerateSummary(nil)
	if !strings.Contains(md, "# Terrain CLI Benchmark Summary") {
		t.Error("empty summary should still have title")
	}
}

func TestWriteResults(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	results := []BenchResult{
		{Repo: RepoMeta{Name: "test"}},
	}
	assessments := []RepoAssessment{
		{Repo: RepoMeta{Name: "test"}, OverallScore: 50},
	}

	err := WriteResults(dir, results, assessments)
	if err != nil {
		t.Fatalf("WriteResults error: %v", err)
	}

	for _, name := range []string{"cli-benchmark-results.json", "cli-benchmark-assessment.json", "cli-benchmark-summary.md"} {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected output file %s to exist", name)
		}
	}
}

// --- ExtractTestID tests ---

func TestExtractTestID_Shape1(t *testing.T) {
	t.Parallel()
	j := `{"impactedTests": [{"testId": "test:auth.test.js:10:login", "confidence": 0.9}]}`
	id := ExtractTestID(j)
	if id != "test:auth.test.js:10:login" {
		t.Errorf("expected test:auth.test.js:10:login, got %q", id)
	}
}

func TestExtractTestID_Shape2(t *testing.T) {
	t.Parallel()
	j := `{"tests": [{"id": "test-123"}]}`
	id := ExtractTestID(j)
	if id != "test-123" {
		t.Errorf("expected test-123, got %q", id)
	}
}

func TestExtractTestID_Shape3(t *testing.T) {
	t.Parallel()
	j := `{"selectedTests": ["test-abc", "test-def"]}`
	id := ExtractTestID(j)
	if id != "test-abc" {
		t.Errorf("expected test-abc, got %q", id)
	}
}

func TestExtractTestID_Shape4(t *testing.T) {
	t.Parallel()
	j := `{"impact": {"impacted": [{"testId": "test-xyz"}]}}`
	id := ExtractTestID(j)
	if id != "test-xyz" {
		t.Errorf("expected test-xyz, got %q", id)
	}
}

func TestExtractTestID_NoTests(t *testing.T) {
	t.Parallel()
	j := `{"noTests": true}`
	id := ExtractTestID(j)
	if id != "" {
		t.Errorf("expected empty, got %q", id)
	}
}

func TestExtractTestID_InvalidJSON(t *testing.T) {
	t.Parallel()
	id := ExtractTestID("not json")
	if id != "" {
		t.Errorf("expected empty for invalid JSON, got %q", id)
	}
}

// --- extractTestIDFromAnalyze tests ---

func TestExtractTestIDFromAnalyze_Found(t *testing.T) {
	t.Parallel()
	j := `{"testCases": [{"testId": "abc123", "testName": "test_login"}]}`
	id := extractTestIDFromAnalyze(j)
	if id != "abc123" {
		t.Errorf("expected abc123, got %q", id)
	}
}

func TestExtractTestIDFromAnalyze_Empty(t *testing.T) {
	t.Parallel()
	j := `{"testCases": []}`
	id := extractTestIDFromAnalyze(j)
	if id != "" {
		t.Errorf("expected empty, got %q", id)
	}
}

func TestExtractTestIDFromAnalyze_NoField(t *testing.T) {
	t.Parallel()
	j := `{"frameworks": ["jest"]}`
	id := extractTestIDFromAnalyze(j)
	if id != "" {
		t.Errorf("expected empty, got %q", id)
	}
}

func TestExtractTestIDFromAnalyze_SkipsMissingID(t *testing.T) {
	t.Parallel()
	// First test case has no testId/id field, second one does.
	// Should iterate past the first and find the second.
	j := `{"testCases": [{"testName": "no_id_here"}, {"testId": "found_me"}]}`
	id := extractTestIDFromAnalyze(j)
	if id != "found_me" {
		t.Errorf("expected found_me, got %q", id)
	}
}

func TestExtractTestIDFromAnalyze_FallbackToID(t *testing.T) {
	t.Parallel()
	// Uses "id" field instead of "testId".
	j := `{"testCases": [{"id": "id_field_test"}]}`
	id := extractTestIDFromAnalyze(j)
	if id != "id_field_test" {
		t.Errorf("expected id_field_test, got %q", id)
	}
}

// --- Assessment: truncation and timeout tests ---

func TestAssessCommand_Truncation(t *testing.T) {
	t.Parallel()

	cr := CommandResult{
		Command:         "analyze",
		RepoName:        "test-repo",
		ExitCode:        0,
		Stdout:          `{"tests": 5}`,
		StdoutTruncated: true,
		StdoutBytes:     1024 * 1024,
	}

	a := AssessCommand(cr)

	foundTrunc := false
	for _, w := range a.WarningFlags {
		if strings.Contains(w, "stdout truncated") {
			foundTrunc = true
		}
	}
	if !foundTrunc {
		t.Error("expected stdout truncated warning flag")
	}
}

func TestAssessCommand_Timeout(t *testing.T) {
	t.Parallel()

	cr := CommandResult{
		Command:   "analyze",
		RepoName:  "test-repo",
		ExitCode:  -1,
		TimedOut:   true,
		RuntimeMs: 60000,
		Error:     "timed out after 1m0s",
	}

	a := AssessCommand(cr)

	foundTimeout := false
	for _, w := range a.WarningFlags {
		if strings.Contains(w, "timed out") {
			foundTimeout = true
		}
	}
	if !foundTimeout {
		t.Error("expected timeout warning flag")
	}
}

// --- Bounded capture tests ---

func TestBoundedCapture_UnderLimit(t *testing.T) {
	t.Parallel()

	bc := newBoundedCapture(1024)
	data := []byte("hello world")
	n, err := bc.Write(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != len(data) {
		t.Errorf("expected %d bytes written, got %d", len(data), n)
	}
	if bc.truncated {
		t.Error("expected no truncation")
	}
	if bc.String() != "hello world" {
		t.Errorf("expected 'hello world', got %q", bc.String())
	}
	if bc.AnnotatedString() != "hello world" {
		t.Errorf("annotated string should match raw when not truncated")
	}
}

func TestBoundedCapture_OverLimit(t *testing.T) {
	t.Parallel()

	bc := newBoundedCapture(1024) // minimum enforced
	// Write more than 1024 bytes.
	chunk := strings.Repeat("x", 600)
	bc.Write([]byte(chunk))
	bc.Write([]byte(chunk)) // total 1200 > 1024

	if !bc.truncated {
		t.Error("expected truncation")
	}
	if len(bc.data) != 1024 {
		t.Errorf("expected 1024 bytes captured, got %d", len(bc.data))
	}
	if bc.total != 1200 {
		t.Errorf("expected 1200 total bytes, got %d", bc.total)
	}
	if !strings.Contains(bc.AnnotatedString(), "output truncated") {
		t.Error("annotated string should contain truncation notice")
	}
}

func TestBoundedCapture_MinimumSize(t *testing.T) {
	t.Parallel()

	bc := newBoundedCapture(10) // should be bumped to 1024
	if bc.maxBytes != 1024 {
		t.Errorf("expected minimum 1024, got %d", bc.maxBytes)
	}
}

// --- Timeout factor tests ---

func TestTimeoutForCommand(t *testing.T) {
	t.Parallel()

	// Save and restore globals.
	origTimeout := PerCommandTimeout
	origFactor := HeavyCommandTimeoutFactor
	defer func() {
		PerCommandTimeout = origTimeout
		HeavyCommandTimeoutFactor = origFactor
	}()

	PerCommandTimeout = 60 * time.Second
	HeavyCommandTimeoutFactor = 4.0

	// Regular command: no scaling.
	if d := timeoutForCommand("analyze"); d != 60*time.Second {
		t.Errorf("analyze timeout = %v, want 60s", d)
	}
	// Insights: factor * 1.5.
	if d := timeoutForCommand("insights"); d != time.Duration(float64(60*time.Second)*6.0) {
		t.Errorf("insights timeout = %v, want 360s", d)
	}
	// Debug: factor.
	if d := timeoutForCommand("debug:graph"); d != time.Duration(float64(60*time.Second)*4.0) {
		t.Errorf("debug:graph timeout = %v, want 240s", d)
	}
	// Explain: 1.5x.
	if d := timeoutForCommand("explain"); d != time.Duration(float64(60*time.Second)*1.5) {
		t.Errorf("explain timeout = %v, want 90s", d)
	}
}

// --- Helper tests ---

func TestTruncate(t *testing.T) {
	t.Parallel()

	if truncate("short", 10) != "short" {
		t.Error("short string should not be truncated")
	}
	if truncate("a long string that should be truncated", 10) != "a long str..." {
		t.Errorf("got: %s", truncate("a long string that should be truncated", 10))
	}
}
