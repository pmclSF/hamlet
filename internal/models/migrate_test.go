package models

import (
	"testing"
	"time"
)

func TestMigrateSnapshotInPlace_BackfillsLegacyFields(t *testing.T) {
	t.Parallel()

	ts := time.Date(2026, 3, 9, 10, 0, 0, 0, time.UTC)
	snap := &TestSuiteSnapshot{
		Repository: RepositoryMetadata{
			Name:              "repo",
			SnapshotTimestamp: ts,
		},
		CodeUnits: []CodeUnit{
			{Path: "src/auth.js", Name: "login"},
		},
	}

	notes := MigrateSnapshotInPlace(snap)
	if len(notes) == 0 {
		t.Fatal("expected compatibility notes for legacy snapshot")
	}
	if snap.SnapshotMeta.SchemaVersion != LegacySnapshotSchemaVersion {
		t.Fatalf("schema version = %q, want %q", snap.SnapshotMeta.SchemaVersion, LegacySnapshotSchemaVersion)
	}
	if snap.GeneratedAt.IsZero() {
		t.Fatal("expected generatedAt to be backfilled from repository snapshot timestamp")
	}
	if got := snap.CodeUnits[0].UnitID; got != "src/auth.js:login" {
		t.Fatalf("code unit ID = %q, want src/auth.js:login", got)
	}
	if snap.Metadata == nil {
		t.Fatal("expected metadata to include compatibility notes")
	}
}

func TestMigrateSnapshotInPlace_NoOpForCurrentSnapshot(t *testing.T) {
	t.Parallel()

	ts := time.Date(2026, 3, 9, 10, 0, 0, 0, time.UTC)
	snap := &TestSuiteSnapshot{
		SnapshotMeta: SnapshotMeta{
			SchemaVersion: SnapshotSchemaVersion,
		},
		Repository: RepositoryMetadata{
			Name:              "repo",
			SnapshotTimestamp: ts,
		},
		GeneratedAt: ts,
		CodeUnits: []CodeUnit{
			{Path: "src/auth.js", Name: "login", UnitID: "src/auth.js:login"},
		},
	}

	notes := MigrateSnapshotInPlace(snap)
	if len(notes) != 0 {
		t.Fatalf("expected no compatibility notes, got %v", notes)
	}
}

func TestMigrateSnapshotInPlace_SchemaCompatibilityNotes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		schemaVersion string
		wantNotes     bool
	}{
		{name: "newer major", schemaVersion: "2.1.0", wantNotes: true},
		{name: "same major", schemaVersion: SnapshotSchemaVersion, wantNotes: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			snap := &TestSuiteSnapshot{
				SnapshotMeta: SnapshotMeta{
					SchemaVersion: tc.schemaVersion,
				},
				Repository: RepositoryMetadata{Name: "repo"},
			}
			notes := MigrateSnapshotInPlace(snap)
			gotNotes := len(notes) > 0
			if gotNotes != tc.wantNotes {
				t.Fatalf("notes presence = %v, want %v (notes=%v)", gotNotes, tc.wantNotes, notes)
			}
		})
	}
}
