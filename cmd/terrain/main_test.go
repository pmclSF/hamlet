package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pmclSF/terrain/internal/models"
)

func TestFindRecentSnapshots_SingleArchivePlusLatest(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	archive := filepath.Join(dir, "2026-03-09T10-00-00Z.json")
	latest := filepath.Join(dir, "latest.json")
	if err := os.WriteFile(archive, []byte("{}"), 0o644); err != nil {
		t.Fatalf("write archive: %v", err)
	}
	if err := os.WriteFile(latest, []byte("{}"), 0o644); err != nil {
		t.Fatalf("write latest: %v", err)
	}

	gotLatest, gotPrevious, err := findRecentSnapshots(dir)
	if err != nil {
		t.Fatalf("findRecentSnapshots returned error: %v", err)
	}
	if gotLatest != latest {
		t.Fatalf("latest = %q, want %q", gotLatest, latest)
	}
	if gotPrevious != archive {
		t.Fatalf("previous = %q, want %q", gotPrevious, archive)
	}
}

func TestLoadSnapshot_MigratesLegacyFields(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "legacy.json")
	json := `{
  "repository": {
    "name": "repo",
    "snapshotTimestamp": "2026-03-01T00:00:00Z"
  },
  "codeUnits": [
    {"path":"src/auth.js","name":"login","kind":"function","exported":true}
  ]
}`
	if err := os.WriteFile(path, []byte(json), 0o644); err != nil {
		t.Fatalf("write legacy snapshot: %v", err)
	}

	snap, err := loadSnapshot(path)
	if err != nil {
		t.Fatalf("loadSnapshot returned error: %v", err)
	}
	if snap.SnapshotMeta.SchemaVersion != models.LegacySnapshotSchemaVersion {
		t.Fatalf("schema version = %q, want %q", snap.SnapshotMeta.SchemaVersion, models.LegacySnapshotSchemaVersion)
	}
	if snap.GeneratedAt.IsZero() {
		t.Fatal("expected generatedAt to be backfilled")
	}
	want := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	if !snap.GeneratedAt.Equal(want) {
		t.Fatalf("generatedAt = %s, want %s", snap.GeneratedAt.UTC().Format(time.RFC3339), want.Format(time.RFC3339))
	}
	if len(snap.CodeUnits) != 1 || snap.CodeUnits[0].UnitID != "src/auth.js:login" {
		t.Fatalf("expected backfilled code unit id, got %+v", snap.CodeUnits)
	}
}

func TestDetectFirstExisting(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if got := detectFirstExisting(root, []string{"coverage/lcov.info"}); got != "" {
		t.Fatalf("expected empty result, got %q", got)
	}

	path := filepath.Join(root, "coverage", "lcov.info")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("TN:\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	got := detectFirstExisting(root, []string{"coverage/lcov.info", "coverage.out"})
	if got != path {
		t.Fatalf("detected path = %q, want %q", got, path)
	}
}

func TestInitLogging_ParsesFlag(t *testing.T) {
	// Not parallel: modifies global logging state.
	initLogging([]string{"analyze", "--log-level=debug", "--root", "."})
	// Should not panic; logger is reconfigured.

	initLogging([]string{"analyze", "--log-level", "quiet"})
	// Should not panic.

	initLogging([]string{"analyze"})
	// No --log-level: keeps default.
}

func TestRunInit_InvalidRoot(t *testing.T) {
	t.Parallel()

	if err := runInit(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected error for missing root")
	}
}

// --- AI command tests ---

func TestIsEvalPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path string
		want bool
	}{
		{"eval/safety.yaml", true},
		{"evals/accuracy.py", true},
		{"evaluations/suite.js", true},
		{"__evals__/prompt_test.py", true},
		{"benchmarks/speed.go", true},
		{"src/eval/runner.ts", true},
		{"test/auth.test.js", false},
		{"src/utils/helper.ts", false},
		{"evaluate.py", false},
		{"", false},
	}
	for _, tt := range tests {
		got := isEvalPath(tt.path)
		if got != tt.want {
			t.Errorf("isEvalPath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestRunAI_UnknownSubcommand(t *testing.T) {
	t.Parallel()
	err := runAI("nonexistent", ".", false)
	if err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
}

func TestRunAI_ScaffoldedCommandsReturnError(t *testing.T) {
	t.Parallel()
	for _, sub := range []string{"run", "record", "baseline"} {
		err := runAI(sub, ".", false)
		if err == nil {
			t.Errorf("terrain ai %s should return not-implemented error", sub)
		}
	}
}

func TestRunAIList_EmptyRepo(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// Empty directory — should produce zero counts without error.
	if err := runAIList(root, false, false); err != nil {
		t.Fatalf("runAIList on empty dir: %v", err)
	}
}

func TestRunAIDoctor_EmptyRepo(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := runAIDoctor(root, false); err != nil {
		t.Fatalf("runAIDoctor on empty dir: %v", err)
	}
}
