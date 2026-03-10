package heatmap

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestBuild_Empty(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{}
	h := Build(snap)

	if h.TotalSignals != 0 {
		t.Errorf("totalSignals = %d, want 0", h.TotalSignals)
	}
	if h.PostureBand != models.RiskBandLow {
		t.Errorf("postureBand = %q, want low", h.PostureBand)
	}
	if h.PostureSummary == "" {
		t.Error("expected non-empty posture summary")
	}
}

func TestBuild_DirectoryHotSpots(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Risk: []models.RiskSurface{
			{
				Type:      "change",
				Scope:     "directory",
				ScopeName: "src/auth",
				Band:      models.RiskBandHigh,
				Score:     12.0,
				ContributingSignals: []models.Signal{
					{Type: "weakAssertion"},
					{Type: "weakAssertion"},
					{Type: "untestedExport"},
				},
			},
			{
				Type:      "change",
				Scope:     "directory",
				ScopeName: "src/payments",
				Band:      models.RiskBandMedium,
				Score:     5.0,
				ContributingSignals: []models.Signal{
					{Type: "mockHeavyTest"},
				},
			},
		},
	}

	h := Build(snap)

	if len(h.DirectoryHotSpots) != 2 {
		t.Fatalf("directoryHotSpots = %d, want 2", len(h.DirectoryHotSpots))
	}
	// Sorted by score descending
	if h.DirectoryHotSpots[0].Name != "src/auth" {
		t.Errorf("first hotspot = %q, want src/auth", h.DirectoryHotSpots[0].Name)
	}
	if h.DirectoryHotSpots[0].SignalCount != 3 {
		t.Errorf("signalCount = %d, want 3", h.DirectoryHotSpots[0].SignalCount)
	}
}

func TestBuild_OwnerHotSpots(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Signals: []models.Signal{
			{Type: "weakAssertion", Owner: "auth-team", Severity: models.SeverityMedium},
			{Type: "weakAssertion", Owner: "auth-team", Severity: models.SeverityMedium},
			{Type: "mockHeavyTest", Owner: "payments-team", Severity: models.SeverityLow},
			{Type: "weakAssertion", Owner: "", Severity: models.SeverityLow},
		},
	}

	h := Build(snap)

	if len(h.OwnerHotSpots) != 3 {
		t.Fatalf("ownerHotSpots = %d, want 3", len(h.OwnerHotSpots))
	}
	// auth-team has score 4.0 (2 * medium=2), should be first
	if h.OwnerHotSpots[0].Name != "auth-team" {
		t.Errorf("first owner = %q, want auth-team", h.OwnerHotSpots[0].Name)
	}
}

func TestBuild_CriticalPosture(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Risk: []models.RiskSurface{
			{Type: "reliability", Scope: "repository", Band: models.RiskBandCritical, Score: 20},
		},
		Signals: []models.Signal{
			{Type: "flakyTest", Severity: models.SeverityCritical},
		},
	}

	h := Build(snap)

	if h.PostureBand != models.RiskBandCritical {
		t.Errorf("postureBand = %q, want critical", h.PostureBand)
	}
	if h.CriticalCount != 1 {
		t.Errorf("criticalCount = %d, want 1", h.CriticalCount)
	}
}

func TestBuild_HighRiskAreaCount(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		Risk: []models.RiskSurface{
			{Type: "change", Scope: "directory", ScopeName: "src/a", Band: models.RiskBandHigh, Score: 10,
				ContributingSignals: []models.Signal{{Type: "weakAssertion"}}},
			{Type: "change", Scope: "directory", ScopeName: "src/b", Band: models.RiskBandLow, Score: 2,
				ContributingSignals: []models.Signal{{Type: "weakAssertion"}}},
		},
		Signals: []models.Signal{
			{Type: "weakAssertion", Owner: "team-a", Severity: models.SeverityHigh},
			{Type: "weakAssertion", Owner: "team-a", Severity: models.SeverityHigh},
			{Type: "weakAssertion", Owner: "team-a", Severity: models.SeverityHigh},
		},
	}

	h := Build(snap)

	// 1 high-risk directory + 1 high-risk owner (team-a has score 9)
	if h.HighRiskAreaCount != 2 {
		t.Errorf("highRiskAreaCount = %d, want 2", h.HighRiskAreaCount)
	}
}

func TestBuild_DirectoryHotSpots_NormalizedByFileCount(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/zsmall/a.test.js"},
			{Path: "src/zsmall/b.test.js"},
			{Path: "src/alarge/a.test.js"},
			{Path: "src/alarge/b.test.js"},
			{Path: "src/alarge/c.test.js"},
			{Path: "src/alarge/d.test.js"},
			{Path: "src/alarge/e.test.js"},
			{Path: "src/alarge/f.test.js"},
			{Path: "src/alarge/g.test.js"},
			{Path: "src/alarge/h.test.js"},
		},
		Risk: []models.RiskSurface{
			{
				Type:      "change",
				Scope:     "directory",
				ScopeName: "src/alarge",
				Band:      models.RiskBandHigh,
				Score:     10,
				ContributingSignals: []models.Signal{
					{Type: "weakAssertion", Severity: models.SeverityMedium},
					{Type: "mockHeavyTest", Severity: models.SeverityMedium},
				},
			},
			{
				Type:      "change",
				Scope:     "directory",
				ScopeName: "src/zsmall",
				Band:      models.RiskBandHigh,
				Score:     10,
				ContributingSignals: []models.Signal{
					{Type: "weakAssertion", Severity: models.SeverityMedium},
					{Type: "mockHeavyTest", Severity: models.SeverityMedium},
				},
			},
		},
	}

	h := Build(snap)
	if len(h.DirectoryHotSpots) != 2 {
		t.Fatalf("directoryHotSpots = %d, want 2", len(h.DirectoryHotSpots))
	}
	// zsmall should rank first because same signal burden over fewer files.
	if h.DirectoryHotSpots[0].Name != "src/zsmall" {
		t.Fatalf("first hotspot = %q, want src/zsmall", h.DirectoryHotSpots[0].Name)
	}
	if h.DirectoryHotSpots[0].FileCount != 2 {
		t.Fatalf("zsmall fileCount = %d, want 2", h.DirectoryHotSpots[0].FileCount)
	}
	if h.DirectoryHotSpots[1].FileCount != 8 {
		t.Fatalf("alarge fileCount = %d, want 8", h.DirectoryHotSpots[1].FileCount)
	}
	if h.DirectoryHotSpots[0].Score <= h.DirectoryHotSpots[1].Score {
		t.Fatalf("expected normalized score ordering, got %.2f <= %.2f", h.DirectoryHotSpots[0].Score, h.DirectoryHotSpots[1].Score)
	}
}

func TestBuild_OwnerHotSpots_NormalizedByOwnedFileCount(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/a1.test.js", Owner: "team-a"},
			{Path: "src/a2.test.js", Owner: "team-a"},
			{Path: "src/a3.test.js", Owner: "team-a"},
			{Path: "src/b1.test.js", Owner: "team-b"},
		},
		Signals: []models.Signal{
			{Type: "weakAssertion", Owner: "team-a", Severity: models.SeverityMedium},
			{Type: "weakAssertion", Owner: "team-a", Severity: models.SeverityMedium},
			{Type: "weakAssertion", Owner: "team-a", Severity: models.SeverityMedium},
			{Type: "weakAssertion", Owner: "team-b", Severity: models.SeverityHigh},
		},
	}

	h := Build(snap)
	if len(h.OwnerHotSpots) < 2 {
		t.Fatalf("ownerHotSpots = %d, want at least 2", len(h.OwnerHotSpots))
	}

	var teamA, teamB *HotSpot
	for i := range h.OwnerHotSpots {
		switch h.OwnerHotSpots[i].Name {
		case "team-a":
			teamA = &h.OwnerHotSpots[i]
		case "team-b":
			teamB = &h.OwnerHotSpots[i]
		}
	}
	if teamA == nil || teamB == nil {
		t.Fatalf("missing expected owner hotspots: team-a=%v team-b=%v", teamA != nil, teamB != nil)
	}
	if teamA.FileCount != 3 || teamB.FileCount != 1 {
		t.Fatalf("unexpected file counts team-a=%d team-b=%d", teamA.FileCount, teamB.FileCount)
	}
	if teamB.Score <= teamA.Score {
		t.Fatalf("expected normalized owner score for team-b to exceed team-a, got %.2f <= %.2f", teamB.Score, teamA.Score)
	}
}

func TestTopTypes(t *testing.T) {
	t.Parallel()
	signals := []models.Signal{
		{Type: "weakAssertion"},
		{Type: "weakAssertion"},
		{Type: "weakAssertion"},
		{Type: "mockHeavyTest"},
		{Type: "mockHeavyTest"},
		{Type: "untestedExport"},
		{Type: "flakyTest"},
	}

	top := topTypes(signals, 2)
	if len(top) != 2 {
		t.Fatalf("topTypes len = %d, want 2", len(top))
	}
	if top[0] != "weakAssertion" {
		t.Errorf("top[0] = %q, want weakAssertion", top[0])
	}
}

func TestScoreToBand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		score float64
		want  models.RiskBand
	}{
		{0, models.RiskBandLow},
		{3, models.RiskBandLow},
		{4, models.RiskBandMedium},
		{9, models.RiskBandHigh},
		{16, models.RiskBandCritical},
	}
	for _, tt := range tests {
		got := scoreToBand(tt.score)
		if got != tt.want {
			t.Errorf("scoreToBand(%.1f) = %q, want %q", tt.score, got, tt.want)
		}
	}
}
