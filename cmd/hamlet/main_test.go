package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pmclSF/hamlet/internal/models"
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

func TestRunInit_InvalidRoot(t *testing.T) {
	t.Parallel()

	if err := runInit(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("expected error for missing root")
	}
}
