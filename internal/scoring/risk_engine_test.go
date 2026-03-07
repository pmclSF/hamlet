package scoring

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestComputeRisk_NoSignals(t *testing.T) {
	snap := &models.TestSuiteSnapshot{}
	surfaces := ComputeRisk(snap)
	if len(surfaces) != 0 {
		t.Errorf("expected 0 risk surfaces, got %d", len(surfaces))
	}
}

func TestComputeRisk_ReliabilitySignals(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		Signals: []models.Signal{
			{Type: "flakyTest", Severity: models.SeverityHigh},
			{Type: "skippedTest", Severity: models.SeverityLow},
		},
	}
	surfaces := ComputeRisk(snap)

	var found bool
	for _, s := range surfaces {
		if s.Type == "reliability" && s.Scope == "repository" {
			found = true
			// high=3 + low=1 = 4 → medium band
			if s.Band != models.RiskBandMedium {
				t.Errorf("band = %q, want medium (score=%.1f)", s.Band, s.Score)
			}
			if s.Score != 4.0 {
				t.Errorf("score = %.1f, want 4.0", s.Score)
			}
		}
	}
	if !found {
		t.Error("expected a reliability risk surface")
	}
}

func TestComputeRisk_ChangeRiskSignals(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		Signals: []models.Signal{
			{Type: "weakAssertion", Severity: models.SeverityMedium},
			{Type: "untestedExport", Severity: models.SeverityMedium},
		},
	}
	surfaces := ComputeRisk(snap)

	var found bool
	for _, s := range surfaces {
		if s.Type == "change" && s.Scope == "repository" {
			found = true
			// medium=2 + medium=2 = 4 → medium band
			if s.Band != models.RiskBandMedium {
				t.Errorf("band = %q, want medium", s.Band)
			}
		}
	}
	if !found {
		t.Error("expected a change risk surface")
	}
}

func TestComputeRisk_DirectoryRollup(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		Signals: []models.Signal{
			{Type: "weakAssertion", Severity: models.SeverityMedium, Location: models.SignalLocation{File: "src/auth/login.test.js"}},
			{Type: "untestedExport", Severity: models.SeverityHigh, Location: models.SignalLocation{File: "src/auth/signup.js"}},
		},
	}
	surfaces := ComputeRisk(snap)

	var dirFound bool
	for _, s := range surfaces {
		if s.Type == "change" && s.Scope == "directory" && s.ScopeName == "src/auth" {
			dirFound = true
		}
	}
	if !dirFound {
		t.Error("expected a directory-level change risk surface for src/auth")
	}
}

func TestScoreToBand(t *testing.T) {
	tests := []struct {
		score float64
		want  models.RiskBand
	}{
		{0, models.RiskBandLow},
		{3, models.RiskBandLow},
		{4, models.RiskBandMedium},
		{8, models.RiskBandMedium},
		{9, models.RiskBandHigh},
		{15, models.RiskBandHigh},
		{16, models.RiskBandCritical},
		{100, models.RiskBandCritical},
	}
	for _, tt := range tests {
		got := scoreToBand(tt.score)
		if got != tt.want {
			t.Errorf("scoreToBand(%.1f) = %q, want %q", tt.score, got, tt.want)
		}
	}
}
