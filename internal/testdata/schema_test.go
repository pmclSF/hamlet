package testdata

import (
	"encoding/json"
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

// TestSchema_SnapshotRoundTrip verifies that TestSuiteSnapshot serializes
// and deserializes without data loss.
func TestSchema_SnapshotRoundTrip(t *testing.T) {
	snap := HealthyBalancedSnapshot()

	data, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded models.TestSuiteSnapshot
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Repository.Name != snap.Repository.Name {
		t.Errorf("name: got %q, want %q", decoded.Repository.Name, snap.Repository.Name)
	}
	if len(decoded.Frameworks) != len(snap.Frameworks) {
		t.Errorf("frameworks: got %d, want %d", len(decoded.Frameworks), len(snap.Frameworks))
	}
	if len(decoded.TestFiles) != len(snap.TestFiles) {
		t.Errorf("test files: got %d, want %d", len(decoded.TestFiles), len(snap.TestFiles))
	}
	if len(decoded.CodeUnits) != len(snap.CodeUnits) {
		t.Errorf("code units: got %d, want %d", len(decoded.CodeUnits), len(snap.CodeUnits))
	}
	if len(decoded.Ownership) != len(snap.Ownership) {
		t.Errorf("ownership: got %d, want %d", len(decoded.Ownership), len(snap.Ownership))
	}
}

// TestSchema_SnapshotWithMeasurements verifies measurement snapshot round-trips.
func TestSchema_SnapshotWithMeasurements(t *testing.T) {
	snap := MinimalSnapshot()
	snap.Measurements = &models.MeasurementSnapshot{
		Posture: []models.DimensionPostureResult{
			{Dimension: "health", Band: "strong", Explanation: "test"},
		},
		Measurements: []models.MeasurementResult{
			{ID: "m1", Dimension: "health", Value: 0.95, Units: "ratio", Band: "strong", Evidence: "strong"},
		},
	}

	data, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}

	var decoded models.TestSuiteSnapshot
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.Measurements == nil {
		t.Fatal("measurements lost during round-trip")
	}
	if len(decoded.Measurements.Posture) != 1 {
		t.Errorf("posture: got %d, want 1", len(decoded.Measurements.Posture))
	}
	if decoded.Measurements.Posture[0].Band != "strong" {
		t.Errorf("band: got %q, want %q", decoded.Measurements.Posture[0].Band, "strong")
	}
}

// TestSchema_ForwardCompatibility verifies that extra JSON fields don't cause errors.
func TestSchema_ForwardCompatibility(t *testing.T) {
	rawJSON := `{
		"repository": {"name": "test-repo"},
		"frameworks": [],
		"testFiles": [],
		"generatedAt": "2025-01-15T12:00:00Z",
		"futureField": "this should be ignored",
		"nested": {"unknown": true}
	}`

	var snap models.TestSuiteSnapshot
	if err := json.Unmarshal([]byte(rawJSON), &snap); err != nil {
		t.Fatalf("forward compat: %v", err)
	}
	if snap.Repository.Name != "test-repo" {
		t.Errorf("name: got %q, want %q", snap.Repository.Name, "test-repo")
	}
}

// TestSchema_EmptySnapshot verifies empty snapshot serializes correctly.
func TestSchema_EmptySnapshot(t *testing.T) {
	snap := EmptySnapshot()

	data, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}

	var decoded models.TestSuiteSnapshot
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.Repository.Name != "empty-repo" {
		t.Errorf("name: got %q", decoded.Repository.Name)
	}
	if len(decoded.TestFiles) != 0 {
		t.Errorf("expected no test files, got %d", len(decoded.TestFiles))
	}
}
