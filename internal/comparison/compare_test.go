package comparison

import (
	"testing"
	"time"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestCompare_SignalDeltas(t *testing.T) {
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

func TestCompare_TestFileCountDelta(t *testing.T) {
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
