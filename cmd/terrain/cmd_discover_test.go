package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pmclSF/terrain/internal/uitokens"
)

// TestRunDiscover_EmptyRepo checks the no-args discovery report stays friendly
// on a repo with no AI surfaces and points at the full posture command.
func TestRunDiscover_EmptyRepo(t *testing.T) {
	tmp := t.TempDir()

	out := captureStdout(t, func() {
		if err := runDiscover(tmp); err != nil {
			t.Fatalf("runDiscover: %v", err)
		}
	})

	if !strings.Contains(out, "MAPPED") {
		t.Errorf("expected the MAPPED line, got: %s", out)
	}
	if !strings.Contains(out, "no AI surfaces here yet") {
		t.Errorf("expected the no-AI friendly line, got: %s", out)
	}
	if !strings.Contains(out, "terrain analyze") {
		t.Errorf("expected next-step pointer at `terrain analyze`, got: %s", out)
	}
}

// TestRunDiscover_AIRepo verifies the report leads with the MAPPED comprehension
// line, shows the HEALTH block, and offers next-step commands when AI surfaces
// are present.
func TestRunDiscover_AIRepo(t *testing.T) {
	tmp := t.TempDir()

	// Seed a minimal AI-shaped layout: a prompt file + an eval config + a
	// model call site importing an AI SDK (so the drift analyzer engages).
	mustWrite(t, filepath.Join(tmp, "prompts", "answer.txt"),
		"You are a helpful assistant. {{user_input}}")
	mustWrite(t, filepath.Join(tmp, "evals", "promptfoo.yaml"),
		"prompts:\n  - prompts/answer.txt\nproviders:\n  - openai:gpt-4\n")
	mustWrite(t, filepath.Join(tmp, "main.py"),
		"import openai\nopenai.ChatCompletion.create(model='gpt-4')\n")

	out := captureStdout(t, func() {
		if err := runDiscover(tmp); err != nil {
			t.Fatalf("runDiscover: %v", err)
		}
	})

	if !strings.Contains(out, "MAPPED") {
		t.Errorf("expected the MAPPED comprehension line, got: %s", out)
	}
	if strings.Contains(out, "no AI surfaces here yet") {
		t.Errorf("AI repo should not report zero surfaces, got: %s", out)
	}
	if !strings.Contains(out, "next") || !strings.Contains(out, "terrain analyze") {
		t.Errorf("expected next-step commands, got: %s", out)
	}
}

// TestRunDiscover_TestRepo verifies a repo with tests but no AI surfaces stays
// in the friendly no-AI state rather than fabricating an AI report.
func TestRunDiscover_TestRepo(t *testing.T) {
	tmp := t.TempDir()

	mustWrite(t, filepath.Join(tmp, "package.json"),
		`{"name": "x", "scripts": {"test": "jest"}}`)
	mustWrite(t, filepath.Join(tmp, "src", "thing.test.js"),
		"test('thing', () => { expect(1).toBe(1) })")

	out := captureStdout(t, func() {
		if err := runDiscover(tmp); err != nil {
			t.Fatalf("runDiscover: %v", err)
		}
	})

	if !strings.Contains(out, "no AI surfaces here yet") {
		t.Errorf("expected the no-AI state on a non-AI repo, got: %s", out)
	}
	if strings.Contains(out, "HEALTH") {
		t.Errorf("did not expect a HEALTH block on a non-AI repo, got: %s", out)
	}
}

// TestRunDiscover_DriftIssue verifies the report surfaces a real prompt↔schema
// drift as a curated issue and reflects it in the contracts health line.
func TestRunDiscover_DriftIssue(t *testing.T) {
	tmp := t.TempDir()
	mustWrite(t, filepath.Join(tmp, "models.py"),
		"from pydantic import BaseModel\n\nclass UserProfile(BaseModel):\n    name: str\n")
	mustWrite(t, filepath.Join(tmp, "prompt.py"),
		"import openai\nfrom models import UserProfile\n\n"+
			"def build(user: UserProfile) -> str:\n"+
			"    return f\"\"\"Hello {user.user_id}.\"\"\"\n")

	out := captureStdout(t, func() {
		if err := runDiscover(tmp); err != nil {
			t.Fatalf("runDiscover: %v", err)
		}
	})

	if !strings.Contains(out, "[drift]") {
		t.Errorf("expected a drift issue card, got: %s", out)
	}
	if !strings.Contains(out, "drifting") {
		t.Errorf("expected the contracts health line to show drifting, got: %s", out)
	}
	// The MAPPED line must show the actual parsed counts (the comprehension proof).
	if !strings.Contains(out, "1 prompt") || !strings.Contains(out, "1 schema") {
		t.Errorf("expected MAPPED counts '1 prompt', '1 schema', got: %s", out)
	}
	// This drift ({user.user_id} vs field name) is not a typo of an existing
	// field, so the producer declines — `terrain fix` must NOT be suggested.
	if strings.Contains(out, "terrain fix") {
		t.Errorf("non-fixable drift must not advertise `terrain fix`, got: %s", out)
	}
}

// TestRunDiscover_FixableDriftSuggestsFix pairs with the above: a drift that IS
// a typo of a real field carries a validated fix, so `terrain fix` is suggested.
func TestRunDiscover_FixableDriftSuggestsFix(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "models.py"),
		"from pydantic import BaseModel\n\nclass UserProfile(BaseModel):\n    user_id: str\n")
	mustWrite(t, filepath.Join(root, "prompt.py"),
		"import openai\nfrom models import UserProfile\n\ndef build(user: UserProfile) -> str:\n    return f\"\"\"Hi {user.user_idx}.\"\"\"\n")

	out := captureStdout(t, func() {
		if err := runDiscover(root); err != nil {
			t.Fatalf("runDiscover: %v", err)
		}
	})
	if !strings.Contains(out, "terrain fix") {
		t.Errorf("a fixable typo drift should suggest `terrain fix`, got: %s", out)
	}
}

// TestRunDiscover_AllClearAndInSync locks the reward states: a consistent AI
// repo (schema + prompt, no drift) shows "all clear", the contracts line reads
// "in sync", and no drift artifacts appear.
func TestRunDiscover_AllClearAndInSync(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "models.py"),
		"from pydantic import BaseModel\n\nclass UserProfile(BaseModel):\n    user_id: str\n    name: str\n")
	mustWrite(t, filepath.Join(root, "prompt.py"),
		"import openai\nfrom models import UserProfile\n\ndef build(user: UserProfile) -> str:\n    return f\"\"\"Hi {user.name} ({user.user_id}).\"\"\"\n")

	out := captureStdout(t, func() {
		if err := runDiscover(root); err != nil {
			t.Fatalf("runDiscover: %v", err)
		}
	})
	if !strings.Contains(out, "all clear") {
		t.Errorf("consistent repo should show 'all clear', got: %s", out)
	}
	if !strings.Contains(out, "in sync") {
		t.Errorf("consistent repo contracts line should read 'in sync', got: %s", out)
	}
	if strings.Contains(out, "[drift]") || strings.Contains(out, "drifting") {
		t.Errorf("consistent repo must show no drift artifacts, got: %s", out)
	}
}

// TestRunDiscover_HonoursGitignore locks the report's .gitignore contract: a
// gitignored directory's prompts, schemas, and drift never reach MAPPED or
// ISSUES, while files that are merely untracked (on disk, not ignored) still
// count. Exercises all three walkers behind the report (aidetect,
// promptcontract, detectSchemaFiles) through the real entry point.
func TestRunDiscover_HonoursGitignore(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, ".gitignore"), "local/\n")

	// Gitignored: a drifting schema+prompt pair plus a schema file — none of
	// it may surface.
	mustWrite(t, filepath.Join(root, "local", "models.py"),
		"from pydantic import BaseModel\n\nclass Ignored(BaseModel):\n    name: str\n")
	mustWrite(t, filepath.Join(root, "local", "prompt.py"),
		"import openai\nfrom models import Ignored\n\n"+
			"def build(user: Ignored) -> str:\n    return f\"\"\"Hello {user.user_id}.\"\"\"\n")
	mustWrite(t, filepath.Join(root, "local", "api.schema.json"), "{}")

	// Untracked but not ignored: a consistent pair that must be counted.
	mustWrite(t, filepath.Join(root, "models.py"),
		"from pydantic import BaseModel\n\nclass UserProfile(BaseModel):\n    user_id: str\n")
	mustWrite(t, filepath.Join(root, "prompt.py"),
		"import openai\nfrom models import UserProfile\n\n"+
			"def build(user: UserProfile) -> str:\n    return f\"\"\"Hi {user.user_id}.\"\"\"\n")

	out := captureStdout(t, func() {
		if err := runDiscover(root); err != nil {
			t.Fatalf("runDiscover: %v", err)
		}
	})

	if strings.Contains(out, "local/") {
		t.Errorf("gitignored paths leaked into the report, got: %s", out)
	}
	// Only the kept pair counts: 1 prompt, 1 schema, no drift.
	if !strings.Contains(out, "1 prompt") || !strings.Contains(out, "1 schema") {
		t.Errorf("expected MAPPED counts '1 prompt', '1 schema' from the kept pair, got: %s", out)
	}
	if strings.Contains(out, "[drift]") {
		t.Errorf("drift from a gitignored directory must not surface, got: %s", out)
	}
	if !strings.Contains(out, "all clear") {
		t.Errorf("expected 'all clear' once ignored drift is excluded, got: %s", out)
	}
}

// TestDetectSchemaFiles_GitignoreAndFixtures pins the schema walker's two
// independent exclusions — .gitignore and the fixture-path heuristic — while
// an untracked non-ignored schema is still returned.
func TestDetectSchemaFiles_GitignoreAndFixtures(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, ".gitignore"), "gen/\n")
	mustWrite(t, filepath.Join(root, "gen", "api.schema.json"), "{}")    // gitignored
	mustWrite(t, filepath.Join(root, "testdata", "s.schema.json"), "{}") // fixture
	mustWrite(t, filepath.Join(root, "api", "v1.schema.json"), "{}")     // kept

	got := detectSchemaFiles(root)
	if len(got) != 1 || got[0] != "api/v1.schema.json" {
		t.Errorf("detectSchemaFiles = %v, want [api/v1.schema.json]", got)
	}
}

// TestDiscoverHealthHelpers unit-tests the coverage-score logic directly (it is
// not exercised by the minimal fixtures, since aidetect doesn't classify their
// surfaces): the co-location coverage count, the meter band boundaries, and the
// path helpers.
func TestDiscoverHealthHelpers(t *testing.T) {
	// countCovered: a surface is covered when an eval config shares its top dir.
	surfaces := []string{"src/agent/chat.py", "src/agent/tools.py", "lib/util.py"}
	evals := []string{"src/agent/eval_chat.yaml"} // top dir "src"
	if got := countCovered(surfaces, evals); got != 2 {
		t.Errorf("countCovered = %d, want 2 (both src/ surfaces covered)", got)
	}
	if got := countCovered(surfaces, nil); got != 0 {
		t.Errorf("countCovered with no evals = %d, want 0", got)
	}

	// topDir + dedupePaths.
	if topDir("a/b/c.py") != "a" || topDir("solo.py") != "solo.py" {
		t.Errorf("topDir wrong: %q / %q", topDir("a/b/c.py"), topDir("solo.py"))
	}
	if got := dedupePaths([]string{"x", "x", "y"}); len(got) != 2 {
		t.Errorf("dedupePaths = %v, want 2 distinct", got)
	}

	// healthMeter: filled cells scale with ratio. Force the Unicode glyphs on so
	// the count is deterministic regardless of the test runner's locale.
	prevU := uitokens.UnicodeEnabled
	uitokens.UnicodeEnabled = true
	defer func() { uitokens.UnicodeEnabled = prevU }()
	if full := countRune(healthMeter(1.0), '●'); full != 10 {
		t.Errorf("healthMeter(1.0) filled cells = %d, want 10", full)
	}
	if empty := countRune(healthMeter(0.0), '○'); empty != 10 {
		t.Errorf("healthMeter(0.0) empty cells = %d, want 10", empty)
	}
}

func countRune(s string, r rune) int {
	n := 0
	for _, c := range s {
		if c == r {
			n++
		}
	}
	return n
}

// TestIsFixturePath spot-checks the heuristic that prevents the report from
// surfacing borrowed test-fixture content as if it were the adopter's code.
func TestIsFixturePath(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"benchmarks/cypress-playwright/cal.com/foo.ts", true},
		{"tests/fixtures/ai-eval-suite/foo.py", true},
		{"vendor/some-pkg/schemas/types.ts", true},
		{"src/app/main.py", false},
		{"schemas/finding.v1.json", false},
		{"docs/schema/analysis.schema.json", false},
	}
	for _, tc := range cases {
		if got := isFixturePath(tc.path); got != tc.want {
			t.Errorf("isFixturePath(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	fn()
	_ = w.Close()
	<-done
	return buf.String()
}

func mustWrite(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
