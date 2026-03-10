package comparison

import (
	"testing"
	"time"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestCompare_SignalDeltas(t *testing.T) {
	t.Parallel()
	from := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		Signals: []models.Signal{
			{Type: "weakAssertion", Category: models.CategoryQuality},
			{Type: "weakAssertion", Category: models.CategoryQuality},
			{Type: "flakyTest", Category: models.CategoryHealth},
		},
	}
	to := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC),
		Signals: []models.Signal{
			{Type: "weakAssertion", Category: models.CategoryQuality},
			{Type: "weakAssertion", Category: models.CategoryQuality},
			{Type: "weakAssertion", Category: models.CategoryQuality},
			{Type: "weakAssertion", Category: models.CategoryQuality},
		},
	}

	comp := Compare(from, to)

	// weakAssertion: 2 → 4 = +2
	// flakyTest: 1 → 0 = -1
	if len(comp.SignalDeltas) != 2 {
		t.Fatalf("expected 2 signal deltas, got %d", len(comp.SignalDeltas))
	}

	// Sorted by absolute delta, so weakAssertion (+2) should be first
	if comp.SignalDeltas[0].Type != "weakAssertion" || comp.SignalDeltas[0].Delta != 2 {
		t.Errorf("first delta = %s %+d, want weakAssertion +2", comp.SignalDeltas[0].Type, comp.SignalDeltas[0].Delta)
	}
	if comp.SignalDeltas[1].Type != "flakyTest" || comp.SignalDeltas[1].Delta != -1 {
		t.Errorf("second delta = %s %+d, want flakyTest -1", comp.SignalDeltas[1].Type, comp.SignalDeltas[1].Delta)
	}
}

func TestCompare_RiskDeltas(t *testing.T) {
	t.Parallel()
	from := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		Risk: []models.RiskSurface{
			{Type: "change", Scope: "repository", ScopeName: "repo", Band: models.RiskBandMedium},
			{Type: "speed", Scope: "repository", ScopeName: "repo", Band: models.RiskBandHigh},
		},
	}
	to := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC),
		Risk: []models.RiskSurface{
			{Type: "change", Scope: "repository", ScopeName: "repo", Band: models.RiskBandHigh},
			{Type: "speed", Scope: "repository", ScopeName: "repo", Band: models.RiskBandHigh},
		},
	}

	comp := Compare(from, to)

	var changeRisk, speedRisk *RiskDelta
	for i, r := range comp.RiskDeltas {
		if r.Type == "change" {
			changeRisk = &comp.RiskDeltas[i]
		}
		if r.Type == "speed" {
			speedRisk = &comp.RiskDeltas[i]
		}
	}

	if changeRisk == nil || !changeRisk.Changed {
		t.Error("expected change risk to be marked as changed")
	}
	if changeRisk != nil && changeRisk.Before != models.RiskBandMedium {
		t.Errorf("change risk before = %q, want medium", changeRisk.Before)
	}
	if changeRisk != nil && changeRisk.After != models.RiskBandHigh {
		t.Errorf("change risk after = %q, want high", changeRisk.After)
	}
	if speedRisk == nil || speedRisk.Changed {
		t.Error("expected speed risk to be unchanged")
	}
}

func TestCompare_NoChanges(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		Signals: []models.Signal{
			{Type: "weakAssertion"},
		},
	}

	comp := Compare(snap, snap)
	if comp.HasMeaningfulChanges() {
		t.Error("expected no meaningful changes when comparing same snapshot")
	}
}

func TestCompare_MethodologyMismatchSuppressesMethodologySensitiveDeltas(t *testing.T) {
	t.Parallel()
	from := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		SnapshotMeta: models.SnapshotMeta{
			SchemaVersion:          models.SnapshotSchemaVersion,
			MethodologyFingerprint: "aaa",
		},
		Risk: []models.RiskSurface{
			{Type: "change", Scope: "repository", ScopeName: "repo", Band: models.RiskBandLow},
		},
		Measurements: &models.MeasurementSnapshot{
			Posture: []models.DimensionPostureResult{
				{Dimension: "health", Band: "strong"},
			},
			Measurements: []models.MeasurementResult{
				{ID: "health.flaky_share", Dimension: "health", Value: 0.01, Band: "strong"},
			},
		},
	}
	to := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC),
		SnapshotMeta: models.SnapshotMeta{
			SchemaVersion:          models.SnapshotSchemaVersion,
			MethodologyFingerprint: "bbb",
		},
		Risk: []models.RiskSurface{
			{Type: "change", Scope: "repository", ScopeName: "repo", Band: models.RiskBandHigh},
		},
		Measurements: &models.MeasurementSnapshot{
			Posture: []models.DimensionPostureResult{
				{Dimension: "health", Band: "weak"},
			},
			Measurements: []models.MeasurementResult{
				{ID: "health.flaky_share", Dimension: "health", Value: 0.30, Band: "weak"},
			},
		},
	}

	comp := Compare(from, to)
	if comp.MethodologyCompatible {
		t.Fatal("expected methodology mismatch to mark comparison as incompatible")
	}
	if len(comp.MethodologyNotes) == 0 {
		t.Fatal("expected methodology notes when incompatible")
	}
	if len(comp.RiskDeltas) != 0 || len(comp.PostureDeltas) != 0 || len(comp.MeasurementDeltas) != 0 {
		t.Fatalf("expected methodology-sensitive deltas to be suppressed, got risk=%d posture=%d measurements=%d",
			len(comp.RiskDeltas), len(comp.PostureDeltas), len(comp.MeasurementDeltas))
	}
}

func TestCompare_MissingMethodologyFingerprintBackCompat(t *testing.T) {
	t.Parallel()
	from := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}
	to := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC),
	}

	comp := Compare(from, to)
	if !comp.MethodologyCompatible {
		t.Fatal("expected compatibility to be assumed when fingerprints are missing")
	}
	if len(comp.MethodologyNotes) == 0 {
		t.Fatal("expected note about missing fingerprint")
	}
}

func TestCompare_TestFileCountDelta(t *testing.T) {
	t.Parallel()
	from := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		TestFiles:   make([]models.TestFile, 10),
	}
	to := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC),
		TestFiles:   make([]models.TestFile, 15),
	}

	comp := Compare(from, to)
	if comp.TestFileCountDelta != 5 {
		t.Errorf("testFileCountDelta = %d, want 5", comp.TestFileCountDelta)
	}
}

func TestCompare_FrameworkChanges(t *testing.T) {
	t.Parallel()
	from := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		Frameworks: []models.Framework{
			{Name: "jest", FileCount: 50},
			{Name: "mocha", FileCount: 10},
		},
	}
	to := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC),
		Frameworks: []models.Framework{
			{Name: "jest", FileCount: 55},
			{Name: "vitest", FileCount: 5},
		},
	}

	comp := Compare(from, to)
	if len(comp.FrameworkChanges) != 2 {
		t.Fatalf("expected 2 framework changes, got %d", len(comp.FrameworkChanges))
	}

	var added, removed bool
	for _, fc := range comp.FrameworkChanges {
		if fc.Name == "vitest" && fc.Change == "added" {
			added = true
		}
		if fc.Name == "mocha" && fc.Change == "removed" {
			removed = true
		}
	}
	if !added {
		t.Error("expected vitest to be flagged as added")
	}
	if !removed {
		t.Error("expected mocha to be flagged as removed")
	}
}

func TestCompare_RepresentativeExamples(t *testing.T) {
	t.Parallel()
	from := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		Signals: []models.Signal{
			{Type: "migrationBlocker", Location: models.SignalLocation{File: "old.test.js"}, Explanation: "old blocker"},
		},
	}
	to := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC),
		Signals: []models.Signal{
			{Type: "weakAssertion", Location: models.SignalLocation{File: "new.test.js"}, Explanation: "new finding"},
		},
	}

	comp := Compare(from, to)
	if len(comp.NewSignalExamples) != 1 {
		t.Errorf("expected 1 new example, got %d", len(comp.NewSignalExamples))
	}
	if len(comp.ResolvedSignalExamples) != 1 {
		t.Errorf("expected 1 resolved example, got %d", len(comp.ResolvedSignalExamples))
	}
}

func TestCompare_RepresentativeExamples_RepoLevelPrecision(t *testing.T) {
	t.Parallel()
	from := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}
	to := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC),
		Signals: []models.Signal{
			{
				Type:        "policyViolation",
				Location:    models.SignalLocation{Repository: "repo"},
				Explanation: "Policy A exceeded threshold",
			},
			{
				Type:        "policyViolation",
				Location:    models.SignalLocation{Repository: "repo"},
				Explanation: "Policy B exceeded threshold",
			},
		},
	}

	comp := Compare(from, to)
	if len(comp.NewSignalExamples) != 2 {
		t.Fatalf("expected 2 distinct repo-level new examples, got %d", len(comp.NewSignalExamples))
	}
}

func TestCompare_TestCaseDeltas(t *testing.T) {
	t.Parallel()
	from := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		TestCases: []models.TestCase{
			{TestID: "aaa", TestName: "test_login"},
			{TestID: "bbb", TestName: "test_logout"},
			{TestID: "ccc", TestName: "test_signup"},
		},
	}
	to := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC),
		TestCases: []models.TestCase{
			{TestID: "aaa", TestName: "test_login"},
			{TestID: "bbb", TestName: "test_logout"},
			{TestID: "ddd", TestName: "test_checkout"},
			{TestID: "eee", TestName: "test_payment"},
		},
	}

	comp := Compare(from, to)
	if comp.TestCaseDeltas == nil {
		t.Fatal("expected TestCaseDeltas to be populated")
	}
	d := comp.TestCaseDeltas
	if d.Stable != 2 {
		t.Errorf("stable = %d, want 2", d.Stable)
	}
	if d.Added != 2 {
		t.Errorf("added = %d, want 2", d.Added)
	}
	if d.Removed != 1 {
		t.Errorf("removed = %d, want 1", d.Removed)
	}
	if len(d.RemovedExamples) != 1 || d.RemovedExamples[0] != "test_signup" {
		t.Errorf("removed examples = %v, want [test_signup]", d.RemovedExamples)
	}
}

func TestCompare_TestCaseDeltas_Empty(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}
	comp := Compare(snap, snap)
	if comp.TestCaseDeltas != nil {
		t.Error("expected nil TestCaseDeltas when both snapshots have no test cases")
	}
}

func TestCompare_CoverageDelta(t *testing.T) {
	t.Parallel()
	from := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		CoverageSummary: &models.CoverageSummary{
			LineCoveragePct:   65.0,
			UncoveredExported: 10,
			CoveredOnlyByE2E:  5,
		},
	}
	to := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC),
		CoverageSummary: &models.CoverageSummary{
			LineCoveragePct:   72.0,
			UncoveredExported: 7,
			CoveredOnlyByE2E:  3,
		},
	}

	comp := Compare(from, to)
	if comp.CoverageDelta == nil {
		t.Fatal("expected CoverageDelta to be populated")
	}
	cd := comp.CoverageDelta
	if cd.LineCoverageBefore != 65.0 {
		t.Errorf("before = %f, want 65.0", cd.LineCoverageBefore)
	}
	if cd.LineCoverageAfter != 72.0 {
		t.Errorf("after = %f, want 72.0", cd.LineCoverageAfter)
	}
	if cd.LineCoverageDelta != 7.0 {
		t.Errorf("delta = %f, want 7.0", cd.LineCoverageDelta)
	}
	if cd.UncoveredExportedBefore != 10 || cd.UncoveredExportedAfter != 7 {
		t.Errorf("uncovered exported = %d→%d, want 10→7", cd.UncoveredExportedBefore, cd.UncoveredExportedAfter)
	}
}

func TestCompare_CoverageDelta_NilBoth(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}
	comp := Compare(snap, snap)
	if comp.CoverageDelta != nil {
		t.Error("expected nil CoverageDelta when neither snapshot has coverage")
	}
}

func TestCompare_PostureDeltas(t *testing.T) {
	t.Parallel()
	from := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		Measurements: &models.MeasurementSnapshot{
			Posture: []models.DimensionPostureResult{
				{Dimension: "health", Band: "strong"},
				{Dimension: "coverage_depth", Band: "moderate"},
			},
			Measurements: []models.MeasurementResult{
				{ID: "health.flaky_share", Dimension: "health", Value: 0.02, Band: "strong"},
				{ID: "coverage_depth.uncovered_exports", Dimension: "coverage_depth", Value: 0.20, Band: "moderate"},
			},
		},
	}
	to := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC),
		Measurements: &models.MeasurementSnapshot{
			Posture: []models.DimensionPostureResult{
				{Dimension: "health", Band: "weak"},
				{Dimension: "coverage_depth", Band: "moderate"},
			},
			Measurements: []models.MeasurementResult{
				{ID: "health.flaky_share", Dimension: "health", Value: 0.25, Band: "weak"},
				{ID: "coverage_depth.uncovered_exports", Dimension: "coverage_depth", Value: 0.20, Band: "moderate"},
			},
		},
	}

	comp := Compare(from, to)

	// Posture: health changed, coverage_depth did not.
	if len(comp.PostureDeltas) != 1 {
		t.Fatalf("expected 1 posture delta, got %d", len(comp.PostureDeltas))
	}
	if comp.PostureDeltas[0].Dimension != "health" {
		t.Errorf("posture delta dimension = %q, want %q", comp.PostureDeltas[0].Dimension, "health")
	}
	if comp.PostureDeltas[0].Before != "strong" || comp.PostureDeltas[0].After != "weak" {
		t.Errorf("posture delta = %q→%q, want strong→weak", comp.PostureDeltas[0].Before, comp.PostureDeltas[0].After)
	}

	// Measurement: flaky_share changed value and band, uncovered_exports unchanged.
	if len(comp.MeasurementDeltas) != 1 {
		t.Fatalf("expected 1 measurement delta, got %d", len(comp.MeasurementDeltas))
	}
	md := comp.MeasurementDeltas[0]
	if md.ID != "health.flaky_share" {
		t.Errorf("measurement delta ID = %q, want %q", md.ID, "health.flaky_share")
	}
	if md.Before != 0.02 || md.After != 0.25 {
		t.Errorf("measurement values = %.2f→%.2f, want 0.02→0.25", md.Before, md.After)
	}
	if !md.BandChanged || md.BandBefore != "strong" || md.BandAfter != "weak" {
		t.Errorf("band change = %v %q→%q, want true strong→weak", md.BandChanged, md.BandBefore, md.BandAfter)
	}

	// Should be meaningful.
	if !comp.HasMeaningfulChanges() {
		t.Error("expected meaningful changes when posture changed")
	}
}

func TestCompare_PostureDeltas_NilMeasurements(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}
	comp := Compare(snap, snap)
	if len(comp.PostureDeltas) != 0 {
		t.Errorf("expected no posture deltas when measurements are nil, got %d", len(comp.PostureDeltas))
	}
	if len(comp.MeasurementDeltas) != 0 {
		t.Errorf("expected no measurement deltas when measurements are nil, got %d", len(comp.MeasurementDeltas))
	}
}

func TestCompare_HasMeaningfulChanges_TestCases(t *testing.T) {
	t.Parallel()
	from := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		TestCases:   []models.TestCase{{TestID: "aaa", TestName: "test_a"}},
	}
	to := &models.TestSuiteSnapshot{
		GeneratedAt: time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC),
		TestCases:   []models.TestCase{{TestID: "bbb", TestName: "test_b"}},
	}
	comp := Compare(from, to)
	if !comp.HasMeaningfulChanges() {
		t.Error("expected meaningful changes when test cases differ")
	}
}

func TestCompare_NilSnapshots(t *testing.T) {
	t.Parallel()
	comp := Compare(nil, nil)
	if comp == nil {
		t.Fatal("expected non-nil comparison result for nil inputs")
	}
	if comp.MethodologyCompatible {
		t.Fatal("expected nil-input comparison to be incompatible")
	}
	if len(comp.MethodologyNotes) == 0 {
		t.Fatal("expected methodology notes for nil-input comparison")
	}
}

func TestCompare_BackfillsLegacySnapshotTimes(t *testing.T) {
	t.Parallel()
	from := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{
			Name:              "repo",
			SnapshotTimestamp: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	to := &models.TestSuiteSnapshot{
		Repository: models.RepositoryMetadata{
			Name:              "repo",
			SnapshotTimestamp: time.Date(2026, 3, 6, 0, 0, 0, 0, time.UTC),
		},
	}

	comp := Compare(from, to)
	if comp.FromTime == "unknown" || comp.ToTime == "unknown" {
		t.Fatalf("expected migrated snapshot times, got from=%q to=%q", comp.FromTime, comp.ToTime)
	}
}
