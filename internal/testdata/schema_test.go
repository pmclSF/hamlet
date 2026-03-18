package testdata

import (
	"encoding/json"
	"testing"

	"github.com/pmclSF/terrain/internal/benchmark"
	"github.com/pmclSF/terrain/internal/impact"
	"github.com/pmclSF/terrain/internal/measurement"
	"github.com/pmclSF/terrain/internal/metrics"
	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/portfolio"
)

// TestSchema_SnapshotRoundTrip verifies that TestSuiteSnapshot serializes
// and deserializes without data loss.
func TestSchema_SnapshotRoundTrip(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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

// TestSchema_ImpactAggregateRoundTrip verifies impact aggregate JSON round-trip.
func TestSchema_ImpactAggregateRoundTrip(t *testing.T) {
	t.Parallel()
	snap := HealthyBalancedSnapshot()
	scope := impact.ChangeScopeFromPaths(
		[]string{"src/auth.js", "src/payment.js", "src/__tests__/auth.test.js"},
		impact.ChangeModified,
	)
	result := impact.Analyze(scope, snap)
	agg := impact.BuildAggregate(result)

	data, err := json.Marshal(agg)
	if err != nil {
		t.Fatal(err)
	}

	var decoded impact.Aggregate
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.ChangedFileCount != agg.ChangedFileCount {
		t.Errorf("changedFileCount: got %d, want %d", decoded.ChangedFileCount, agg.ChangedFileCount)
	}
	if decoded.Posture != agg.Posture {
		t.Errorf("posture: got %q, want %q", decoded.Posture, agg.Posture)
	}
}

// TestSchema_ImpactAggregateForwardCompat verifies extra fields are ignored.
func TestSchema_ImpactAggregateForwardCompat(t *testing.T) {
	t.Parallel()
	rawJSON := `{
		"changedFileCount": 5,
		"impactedUnitCount": 3,
		"posture": "well_protected",
		"futureField": "should be ignored",
		"protectionCounts": {},
		"confidenceCounts": {}
	}`
	var agg impact.Aggregate
	if err := json.Unmarshal([]byte(rawJSON), &agg); err != nil {
		t.Fatalf("forward compat: %v", err)
	}
	if agg.ChangedFileCount != 5 {
		t.Errorf("changedFileCount: got %d, want 5", agg.ChangedFileCount)
	}
}

// TestSchema_BenchmarkExportRoundTrip verifies export JSON round-trip.
func TestSchema_BenchmarkExportRoundTrip(t *testing.T) {
	t.Parallel()
	snap := MinimalSnapshot()
	ms := metrics.Derive(snap)
	ms.GeneratedAt = FixedTime

	measReg, mErr := measurement.DefaultRegistry(); if mErr != nil { t.Fatal(mErr) }
	snap.Measurements = measReg.ComputeSnapshot(snap).ToModel()
	snap.Portfolio = portfolio.Analyze(snap).ToModel()

	export := benchmark.BuildExport(snap, ms, false)
	export.ExportedAt = FixedTime

	data, err := json.Marshal(export)
	if err != nil {
		t.Fatal(err)
	}

	var decoded benchmark.Export
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.SchemaVersion != export.SchemaVersion {
		t.Errorf("schemaVersion: got %q, want %q", decoded.SchemaVersion, export.SchemaVersion)
	}
	if !decoded.ExportedAt.Equal(export.ExportedAt) {
		t.Errorf("exportedAt: got %v, want %v", decoded.ExportedAt, export.ExportedAt)
	}
}

// TestSchema_PortfolioRoundTrip verifies portfolio model round-trip.
func TestSchema_PortfolioRoundTrip(t *testing.T) {
	t.Parallel()
	snap := FlakyConcentratedSnapshot()
	ps := portfolio.Analyze(snap)
	model := ps.ToModel()

	data, err := json.Marshal(model)
	if err != nil {
		t.Fatal(err)
	}

	var decoded models.PortfolioSnapshot
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.Aggregates.TotalAssets != model.Aggregates.TotalAssets {
		t.Errorf("totalAssets: got %d, want %d", decoded.Aggregates.TotalAssets, model.Aggregates.TotalAssets)
	}
}

// TestSchema_CoverageInsightRoundTrip verifies coverage insight JSON round-trip.
func TestSchema_CoverageInsightRoundTrip(t *testing.T) {
	t.Parallel()
	insight := models.CoverageInsight{
		Type:        "e2e_only_coverage",
		Severity:    "medium",
		Path:        "src/db.js",
		UnitID:      "src/db.js:DbAdapter",
		Description: "covered only by E2E",
	}

	data, err := json.Marshal(insight)
	if err != nil {
		t.Fatal(err)
	}

	var decoded models.CoverageInsight
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.Type != insight.Type {
		t.Errorf("type: got %q, want %q", decoded.Type, insight.Type)
	}
	if decoded.UnitID != insight.UnitID {
		t.Errorf("unitID: got %q, want %q", decoded.UnitID, insight.UnitID)
	}
}
