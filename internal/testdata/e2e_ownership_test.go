package testdata

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/pmclSF/terrain/internal/benchmark"
	"github.com/pmclSF/terrain/internal/comparison"
	"github.com/pmclSF/terrain/internal/heatmap"
	"github.com/pmclSF/terrain/internal/metrics"
	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/ownership"
	"github.com/pmclSF/terrain/internal/reporting"
	"github.com/pmclSF/terrain/internal/scoring"
)

// TestE2E_OwnershipPropagationAndAggregation exercises the full ownership
// flow: resolve -> propagate -> aggregate -> render -> benchmark-safe export.
func TestE2E_OwnershipPropagationAndAggregation(t *testing.T) {
	t.Parallel()
	snap := HealthyBalancedSnapshot()

	// Add ownership data to simulate CODEOWNERS resolution.
	snap.Ownership = map[string][]string{
		"src/auth.js":                   {"team-platform"},
		"src/user.js":                   {"team-platform"},
		"src/payment.js":                {"team-payments"},
		"src/config.js":                 {"team-platform"},
		"src/__tests__/auth.test.js":    {"team-platform"},
		"src/__tests__/user.test.js":    {"team-platform"},
		"src/__tests__/payment.test.js": {"team-payments"},
		"e2e/checkout.spec.js":          {"team-payments", "team-qa"},
	}

	// Set owners on test files and code units to match ownership map.
	for i := range snap.TestFiles {
		if owners, ok := snap.Ownership[snap.TestFiles[i].Path]; ok && len(owners) > 0 {
			snap.TestFiles[i].Owner = owners[0]
		}
	}
	for i := range snap.CodeUnits {
		if owners, ok := snap.Ownership[snap.CodeUnits[i].Path]; ok && len(owners) > 0 {
			snap.CodeUnits[i].Owner = owners[0]
		}
	}

	// Add signals with owners.
	snap.Signals = append(snap.Signals,
		models.Signal{Type: "weakAssertion", Category: models.CategoryQuality, Severity: models.SeverityMedium, Owner: "team-platform", Location: models.SignalLocation{File: "src/__tests__/auth.test.js"}},
		models.Signal{Type: "weakAssertion", Category: models.CategoryQuality, Severity: models.SeverityMedium, Owner: "team-platform", Location: models.SignalLocation{File: "src/__tests__/user.test.js"}},
		models.Signal{Type: "weakAssertion", Category: models.CategoryQuality, Severity: models.SeverityMedium, Owner: "team-payments", Location: models.SignalLocation{File: "src/__tests__/payment.test.js"}},
		models.Signal{Type: "flakyTest", Category: models.CategoryHealth, Severity: models.SeverityHigh, Owner: "team-payments", Location: models.SignalLocation{File: "e2e/checkout.spec.js"}},
	)
	snap.Risk = scoring.ComputeRisk(snap)

	// Build owner-aware health summaries.
	healthSummaries := ownership.BuildHealthSummaries(snap)
	if len(healthSummaries) == 0 {
		t.Error("expected health summaries")
	}
	// team-payments should have the flaky test.
	found := false
	for _, hs := range healthSummaries {
		if hs.Owner == "team-payments" && hs.FlakyCount > 0 {
			found = true
		}
	}
	if !found {
		t.Error("expected team-payments to have flaky test in health summary")
	}

	// Build owner-aware quality summaries.
	qualitySummaries := ownership.BuildQualitySummaries(snap)
	if len(qualitySummaries) == 0 {
		t.Error("expected quality summaries")
	}

	// Build focus items.
	ownerSummary := ownership.OwnershipSummary{
		OwnerCount: 3,
		Owners: []ownership.OwnerAggregate{
			{Owner: ownership.Owner{ID: "team-platform"}, SignalCount: 2},
			{Owner: ownership.Owner{ID: "team-payments"}, SignalCount: 2},
		},
	}
	focusItems := ownership.BuildFocusItems(ownerSummary, healthSummaries, qualitySummaries)
	// Focus items may or may not be generated depending on thresholds.
	t.Logf("generated %d focus items", len(focusItems))

	// Benchmark export should include ownership stats.
	ms := metrics.Derive(snap)
	export := benchmark.BuildExport(snap, ms, false)
	if export.OwnershipStats == nil {
		t.Error("expected ownership stats in benchmark export")
	} else {
		if export.OwnershipStats.OwnerCount == 0 {
			t.Error("expected non-zero owner count in export")
		}
		// Should not contain owner names.
		data, err := json.Marshal(export.OwnershipStats)
		if err != nil {
			t.Fatal(err)
		}
		output := string(data)
		if strings.Contains(output, "team-platform") || strings.Contains(output, "team-payments") {
			t.Errorf("ownership export contains owner names — privacy violation: %s", output)
		}
	}

	// Render summary report should include ownership section.
	var buf bytes.Buffer
	h := heatmap.Build(snap)
	reporting.RenderSummaryReport(&buf, snap, h)
	if !strings.Contains(buf.String(), "Ownership Coverage") {
		t.Error("expected 'Ownership Coverage' section in summary report")
	}
}

// TestE2E_OwnershipComparisonTrend exercises ownership delta in comparison.
func TestE2E_OwnershipComparisonTrend(t *testing.T) {
	t.Parallel()
	from := HealthyBalancedSnapshot()
	from.GeneratedAt = from.GeneratedAt.Add(-24 * time.Hour)
	// Simulate "before" with fewer owned files.
	from.Ownership = map[string][]string{
		"src/auth.js": {"team-platform"},
	}

	to := HealthyBalancedSnapshot()
	// "after" has more owned files.
	to.Ownership = map[string][]string{
		"src/auth.js":    {"team-platform"},
		"src/payment.js": {"team-payments"},
		"src/user.js":    {"team-platform"},
	}

	comp := comparison.Compare(from, to)
	if comp.OwnershipDelta == nil {
		t.Fatal("expected ownership delta")
	}
	if comp.OwnershipDelta.OwnedFilesAfter != 3 {
		t.Errorf("owned files after = %d, want 3", comp.OwnershipDelta.OwnedFilesAfter)
	}
	if comp.OwnershipDelta.OwnedFilesBefore != 1 {
		t.Errorf("owned files before = %d, want 1", comp.OwnershipDelta.OwnedFilesBefore)
	}
	if !comp.OwnershipDelta.OwnershipImproved {
		t.Error("expected ownership to be marked as improved")
	}
}
