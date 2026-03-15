package ownership

import (
	"testing"

	"github.com/pmclSF/terrain/internal/models"
	"github.com/pmclSF/terrain/internal/signals"
)

func TestBuildHealthSummaries(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Signals: []models.Signal{
			{Type: signals.SignalFlakyTest, Category: models.CategoryHealth, Owner: "team-auth", Location: models.SignalLocation{File: "src/auth/login.test.js"}},
			{Type: signals.SignalFlakyTest, Category: models.CategoryHealth, Owner: "team-auth", Location: models.SignalLocation{File: "src/auth/login.test.js"}},
			{Type: signals.SignalSlowTest, Category: models.CategoryHealth, Owner: "team-pay", Location: models.SignalLocation{File: "src/pay/checkout.test.js"}},
			{Type: "weakAssertion", Category: models.CategoryQuality, Owner: "team-auth"},
		},
	}

	summaries := BuildHealthSummaries(snap)

	if len(summaries) != 2 {
		t.Fatalf("got %d summaries, want 2", len(summaries))
	}

	// team-auth should be first (more health signals).
	if summaries[0].Owner != "team-auth" {
		t.Errorf("first owner = %q, want %q", summaries[0].Owner, "team-auth")
	}
	if summaries[0].FlakyCount != 2 {
		t.Errorf("flaky count = %d, want 2", summaries[0].FlakyCount)
	}
	if summaries[0].Concentration != "localized" {
		t.Errorf("concentration = %q, want %q", summaries[0].Concentration, "localized")
	}
}

func TestBuildQualitySummaries(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Signals: []models.Signal{
			{Type: signals.SignalWeakAssertion, Category: models.CategoryQuality, Owner: "team-auth"},
			{Type: signals.SignalWeakAssertion, Category: models.CategoryQuality, Owner: "team-auth"},
			{Type: signals.SignalWeakAssertion, Category: models.CategoryQuality, Owner: "team-auth"},
			{Type: signals.SignalUntestedExport, Category: models.CategoryQuality, Owner: "team-pay"},
			{Type: signals.SignalFlakyTest, Category: models.CategoryHealth, Owner: "team-auth"}, // non-quality
		},
	}

	summaries := BuildQualitySummaries(snap)
	if len(summaries) != 2 {
		t.Fatalf("got %d summaries, want 2", len(summaries))
	}

	// team-auth should be first (more quality signals).
	if summaries[0].Owner != "team-auth" {
		t.Errorf("first owner = %q, want %q", summaries[0].Owner, "team-auth")
	}
	if summaries[0].WeakAssertionCount != 3 {
		t.Errorf("weak assertion count = %d, want 3", summaries[0].WeakAssertionCount)
	}
	if summaries[0].QualityPosture != "weak" {
		t.Errorf("posture = %q, want %q", summaries[0].QualityPosture, "weak")
	}
}

func TestBuildMigrationSummaries(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Signals: []models.Signal{
			{Type: signals.SignalDeprecatedTestPattern, Category: models.CategoryMigration, Owner: "team-auth"},
			{Type: signals.SignalCustomMatcherRisk, Category: models.CategoryMigration, Owner: "team-auth"},
			{Type: signals.SignalDeprecatedTestPattern, Category: models.CategoryMigration, Owner: "team-pay"},
		},
	}

	summaries := BuildMigrationSummaries(snap)
	if len(summaries) != 2 {
		t.Fatalf("got %d summaries, want 2", len(summaries))
	}
	if summaries[0].Owner != "team-auth" {
		t.Errorf("first owner = %q, want %q", summaries[0].Owner, "team-auth")
	}
	if summaries[0].BlockerCount != 2 {
		t.Errorf("blocker count = %d, want 2", summaries[0].BlockerCount)
	}
}

func TestComputeMigrationCoordinationRisk(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		summaries []OwnerMigrationSummary
		wantLevel string
	}{
		{
			name:      "no blockers",
			summaries: nil,
			wantLevel: "low",
		},
		{
			name: "one owner",
			summaries: []OwnerMigrationSummary{
				{Owner: "team-auth", BlockerCount: 5},
			},
			wantLevel: "low",
		},
		{
			name: "two owners",
			summaries: []OwnerMigrationSummary{
				{Owner: "team-auth", BlockerCount: 5},
				{Owner: "team-pay", BlockerCount: 3},
			},
			wantLevel: "medium",
		},
		{
			name: "many owners",
			summaries: []OwnerMigrationSummary{
				{Owner: "team-a", BlockerCount: 2},
				{Owner: "team-b", BlockerCount: 2},
				{Owner: "team-c", BlockerCount: 2},
				{Owner: "team-d", BlockerCount: 2},
			},
			wantLevel: "high",
		},
		{
			name: "unowned blockers",
			summaries: []OwnerMigrationSummary{
				{Owner: "unknown", BlockerCount: 3},
			},
			wantLevel: "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			risk := ComputeMigrationCoordinationRisk(tt.summaries)
			if risk.Level != tt.wantLevel {
				t.Errorf("level = %q, want %q", risk.Level, tt.wantLevel)
			}
		})
	}
}

func TestCompareOwnerSignals(t *testing.T) {
	t.Parallel()
	from := &models.TestSuiteSnapshot{
		Signals: []models.Signal{
			{Owner: "team-auth", Type: "weakAssertion"},
			{Owner: "team-auth", Type: "weakAssertion"},
			{Owner: "team-pay", Type: "weakAssertion"},
		},
	}
	to := &models.TestSuiteSnapshot{
		Signals: []models.Signal{
			{Owner: "team-auth", Type: "weakAssertion"},
			{Owner: "team-pay", Type: "weakAssertion"},
			{Owner: "team-pay", Type: "weakAssertion"},
			{Owner: "team-pay", Type: "weakAssertion"},
		},
	}

	trends := CompareOwnerSignals(from, to)
	if len(trends) != 2 {
		t.Fatalf("got %d trends, want 2", len(trends))
	}

	// team-pay should have worsened (1 -> 3).
	found := false
	for _, tr := range trends {
		if tr.Owner == "team-pay" {
			found = true
			if tr.Direction != "worsened" {
				t.Errorf("team-pay direction = %q, want %q", tr.Direction, "worsened")
			}
			if tr.SignalDelta != 2 {
				t.Errorf("team-pay delta = %d, want 2", tr.SignalDelta)
			}
		}
	}
	if !found {
		t.Error("expected team-pay trend")
	}
}

func TestBuildFocusItems(t *testing.T) {
	t.Parallel()
	summary := OwnershipSummary{
		Owners: []OwnerAggregate{
			{Owner: Owner{ID: "unknown"}, CriticalSignalCount: 2, UncoveredExportedCount: 3},
			{Owner: Owner{ID: "team-auth"}, SignalCount: 5},
		},
	}
	healthSummaries := []OwnerHealthSummary{
		{Owner: "team-auth", TotalHealth: 3, Concentration: "localized", TopFiles: []string{"src/auth/login.test.js"}},
	}
	qualitySummaries := []OwnerQualitySummary{
		{Owner: "team-auth", QualityPosture: "weak", WeakAssertionCount: 5, UncoveredExported: 2},
	}

	items := BuildFocusItems(summary, healthSummaries, qualitySummaries)
	if len(items) == 0 {
		t.Fatal("expected focus items")
	}
	// First item should be about unowned areas.
	if items[0].Owner != "unknown" {
		t.Errorf("first item owner = %q, want %q", items[0].Owner, "unknown")
	}
}

func TestBuildBenchmarkAggregate(t *testing.T) {
	t.Parallel()
	summary := OwnershipSummary{
		OwnerCount:      3,
		CoveragePosture: "partial",
		Owners: []OwnerAggregate{
			{Owner: Owner{ID: "team-auth"}, SignalCount: 5},
			{Owner: Owner{ID: "team-pay"}, SignalCount: 3},
			{Owner: Owner{ID: "team-infra"}, SignalCount: 2},
		},
	}
	snap := &models.TestSuiteSnapshot{
		CodeUnits: []models.CodeUnit{
			{Exported: true, Owner: "team-auth"},
			{Exported: true, Owner: "team-pay"},
			{Exported: true, Owner: "unknown"},
			{Exported: false, Owner: "team-auth"},
		},
	}

	agg := BuildBenchmarkAggregate(summary, snap)
	if agg == nil {
		t.Fatal("expected non-nil aggregate")
	}
	if agg.OwnerCount != 3 {
		t.Errorf("owner count = %d, want 3", agg.OwnerCount)
	}
	if agg.CoveragePosture != "partial" {
		t.Errorf("posture = %q, want %q", agg.CoveragePosture, "partial")
	}
	// Top owner has 5 out of 10 signals = 50%.
	if agg.TopOwnerRiskSharePct != 50.0 {
		t.Errorf("top risk share = %.1f%%, want 50.0%%", agg.TopOwnerRiskSharePct)
	}
	// 1 out of 3 exported units is unowned = 33.3%.
	if agg.UnownedCriticalPct < 33.0 || agg.UnownedCriticalPct > 34.0 {
		t.Errorf("unowned critical = %.1f%%, want ~33.3%%", agg.UnownedCriticalPct)
	}
	// Fragmentation should be > 0 (3 owners with different counts).
	if agg.FragmentationIndex <= 0 {
		t.Errorf("fragmentation = %.2f, want > 0", agg.FragmentationIndex)
	}
}

func TestBuildBenchmarkAggregate_NilWhenNoOwners(t *testing.T) {
	t.Parallel()
	summary := OwnershipSummary{OwnerCount: 0}
	snap := &models.TestSuiteSnapshot{}

	agg := BuildBenchmarkAggregate(summary, snap)
	if agg != nil {
		t.Error("expected nil aggregate when no owners")
	}
}

func TestComputeFragmentation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		owners []OwnerAggregate
		total  int
		want   float64
	}{
		{
			name:   "single owner",
			owners: []OwnerAggregate{{SignalCount: 10}},
			total:  10,
			want:   0,
		},
		{
			name:   "two owners even",
			owners: []OwnerAggregate{{SignalCount: 5}, {SignalCount: 5}},
			total:  10,
			want:   1.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := computeFragmentation(tt.owners, tt.total)
			if got < tt.want-0.01 || got > tt.want+0.01 {
				t.Errorf("fragmentation = %.2f, want %.2f", got, tt.want)
			}
		})
	}
}
