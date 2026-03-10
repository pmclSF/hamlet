package models

import (
	"strconv"
	"strings"
)

// LegacySnapshotSchemaVersion identifies snapshots written before schema
// version stamping was introduced.
const LegacySnapshotSchemaVersion = "0.0.0"

// MigrateSnapshotInPlace applies backward-compatible normalization for snapshots
// loaded from disk. It preserves existing values and only fills missing fields.
//
// The returned notes describe compatibility adjustments that were applied.
func MigrateSnapshotInPlace(snap *TestSuiteSnapshot) []string {
	if snap == nil {
		return nil
	}

	var notes []string

	if strings.TrimSpace(snap.SnapshotMeta.SchemaVersion) == "" {
		snap.SnapshotMeta.SchemaVersion = LegacySnapshotSchemaVersion
		notes = append(notes, "snapshotMeta.schemaVersion missing; treated as legacy schema 0.0.0")
	}

	if snap.GeneratedAt.IsZero() && !snap.Repository.SnapshotTimestamp.IsZero() {
		snap.GeneratedAt = snap.Repository.SnapshotTimestamp
		notes = append(notes, "generatedAt missing; backfilled from repository.snapshotTimestamp")
	} else if snap.Repository.SnapshotTimestamp.IsZero() && !snap.GeneratedAt.IsZero() {
		snap.Repository.SnapshotTimestamp = snap.GeneratedAt
		notes = append(notes, "repository.snapshotTimestamp missing; backfilled from generatedAt")
	}

	backfilledUnitIDs := 0
	for i := range snap.CodeUnits {
		cu := &snap.CodeUnits[i]
		if cu.UnitID != "" || cu.Path == "" || cu.Name == "" {
			continue
		}
		cu.UnitID = legacyUnitID(cu.Path, cu.Name, cu.ParentName)
		backfilledUnitIDs++
	}
	if backfilledUnitIDs > 0 {
		notes = append(notes, "code unit IDs backfilled for legacy snapshot compatibility")
	}

	// Preserve explicit provenance about whether this snapshot appears older/newer.
	if currentMajor, loadedMajor, ok := compareSchemaMajor(SnapshotSchemaVersion, snap.SnapshotMeta.SchemaVersion); ok {
		switch {
		case loadedMajor < currentMajor:
			notes = append(notes, "snapshot schema is older than current runtime schema; compatibility mode applied")
		case loadedMajor > currentMajor:
			notes = append(notes, "snapshot schema is newer than current runtime schema; compare results may be limited")
		}
	}

	if len(notes) > 0 {
		if snap.Metadata == nil {
			snap.Metadata = map[string]any{}
		}
		snap.Metadata["compatibilityNotes"] = append([]string(nil), notes...)
	}

	return notes
}

func legacyUnitID(path, name, parent string) string {
	if parent != "" {
		return path + ":" + parent + "." + name
	}
	return path + ":" + name
}

func compareSchemaMajor(current, loaded string) (currentMajor int, loadedMajor int, ok bool) {
	currentMajor, okCurrent := schemaMajor(current)
	loadedMajor, okLoaded := schemaMajor(loaded)
	if !okCurrent || !okLoaded {
		return 0, 0, false
	}
	return currentMajor, loadedMajor, true
}

func schemaMajor(v string) (int, bool) {
	parts := strings.SplitN(strings.TrimSpace(v), ".", 2)
	if len(parts) == 0 || parts[0] == "" {
		return 0, false
	}
	n, err := strconv.Atoi(parts[0])
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}
