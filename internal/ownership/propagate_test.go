package ownership

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestPropagate_BasicInheritance(t *testing.T) {
	dir := t.TempDir()
	githubDir := filepath.Join(dir, ".github")
	os.MkdirAll(githubDir, 0755)
	os.WriteFile(filepath.Join(githubDir, "CODEOWNERS"), []byte(`
/src/auth/ @team-auth
/src/payments/ @team-pay @team-billing
`), 0644)

	resolver := NewResolver(dir)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/auth/login.test.js", Owner: ""},
			{Path: "src/payments/checkout.test.js", Owner: ""},
			{Path: "src/unknown/misc.test.js", Owner: ""},
		},
		CodeUnits: []models.CodeUnit{
			{Path: "src/auth/service.js", Name: "AuthService", Exported: true},
			{Path: "src/payments/stripe.js", Name: "charge", Exported: true},
			{Path: "src/unknown/util.js", Name: "helper", Exported: false},
		},
		TestCases: []models.TestCase{
			{TestID: "tc1", FilePath: "src/auth/login.test.js", TestName: "login works"},
			{TestID: "tc2", FilePath: "src/payments/checkout.test.js", TestName: "checkout works"},
			{TestID: "tc3", FilePath: "src/unknown/misc.test.js", TestName: "misc works"},
		},
		Signals: []models.Signal{
			{Type: "weakAssertion", Category: models.CategoryQuality, Location: models.SignalLocation{File: "src/auth/login.test.js"}},
			{Type: "flakyTest", Category: models.CategoryHealth, Location: models.SignalLocation{File: "src/payments/checkout.test.js"}},
		},
	}

	result := Propagate(resolver, snap)

	// Test files should have ownership.
	if snap.TestFiles[0].Owner != "team-auth" {
		t.Errorf("auth test file owner = %q, want %q", snap.TestFiles[0].Owner, "team-auth")
	}
	if snap.TestFiles[1].Owner != "team-pay" {
		t.Errorf("payments test file owner = %q, want %q", snap.TestFiles[1].Owner, "team-pay")
	}

	// Code units should inherit ownership.
	if snap.CodeUnits[0].Owner != "team-auth" {
		t.Errorf("auth code unit owner = %q, want %q", snap.CodeUnits[0].Owner, "team-auth")
	}
	if snap.CodeUnits[1].Owner != "team-pay" {
		t.Errorf("payments code unit owner = %q, want %q", snap.CodeUnits[1].Owner, "team-pay")
	}

	// Signals should have ownership propagated.
	if snap.Signals[0].Owner != "team-auth" {
		t.Errorf("auth signal owner = %q, want %q", snap.Signals[0].Owner, "team-auth")
	}
	if snap.Signals[1].Owner != "team-pay" {
		t.Errorf("payments signal owner = %q, want %q", snap.Signals[1].Owner, "team-pay")
	}

	// Summary should be populated.
	s := result.Summary
	if s.TotalFiles == 0 {
		t.Error("expected non-zero total files")
	}
	if s.OwnerCount < 2 {
		t.Errorf("expected at least 2 owners, got %d", s.OwnerCount)
	}
	if s.CoveragePosture == "none" {
		t.Error("expected non-none coverage posture")
	}

	// Ownership map should be populated.
	if len(snap.Ownership) == 0 {
		t.Error("expected ownership map to be populated")
	}

	// Multi-owner should be in the map.
	if owners, ok := snap.Ownership["src/payments/checkout.test.js"]; ok {
		if len(owners) != 2 {
			t.Errorf("expected 2 owners for payments file, got %d", len(owners))
		}
	}
}

func TestPropagate_DirectAssignmentPreserved(t *testing.T) {
	dir := t.TempDir()
	resolver := NewResolver(dir)

	snap := &models.TestSuiteSnapshot{
		CodeUnits: []models.CodeUnit{
			{Path: "src/auth/service.js", Name: "AuthService", Owner: "direct-owner"},
		},
		Signals: []models.Signal{
			{Type: "weakAssertion", Category: models.CategoryQuality, Owner: "signal-owner", Location: models.SignalLocation{File: "src/auth/login.test.js"}},
		},
	}

	Propagate(resolver, snap)

	// Direct assignment should be preserved.
	if snap.CodeUnits[0].Owner != "direct-owner" {
		t.Errorf("code unit owner = %q, want %q (direct preserved)", snap.CodeUnits[0].Owner, "direct-owner")
	}
	if snap.Signals[0].Owner != "signal-owner" {
		t.Errorf("signal owner = %q, want %q (direct preserved)", snap.Signals[0].Owner, "signal-owner")
	}
}

func TestPropagate_UnownedTracking(t *testing.T) {
	dir := t.TempDir()
	resolver := NewResolver(dir)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "standalone.test.js"},
		},
		CodeUnits: []models.CodeUnit{
			{Path: "standalone.js", Name: "main", Exported: true},
		},
	}

	result := Propagate(resolver, snap)

	if result.Summary.UnownedFiles == 0 {
		t.Error("expected unowned files to be tracked")
	}
	if result.Summary.CoveragePosture != "none" {
		t.Errorf("expected 'none' posture for standalone files, got %q", result.Summary.CoveragePosture)
	}
}

func TestPropagate_OwnerAggregates(t *testing.T) {
	dir := t.TempDir()
	githubDir := filepath.Join(dir, ".github")
	os.MkdirAll(githubDir, 0755)
	os.WriteFile(filepath.Join(githubDir, "CODEOWNERS"), []byte(`
/src/auth/ @team-auth
`), 0644)

	resolver := NewResolver(dir)

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "src/auth/login.test.js"},
		},
		CodeUnits: []models.CodeUnit{
			{Path: "src/auth/service.js", Name: "login", Exported: true},
			{Path: "src/auth/session.js", Name: "session", Exported: true},
		},
		TestCases: []models.TestCase{
			{TestID: "tc1", FilePath: "src/auth/login.test.js"},
		},
		Signals: []models.Signal{
			{Type: "weakAssertion", Category: models.CategoryQuality, Severity: models.SeverityCritical, Location: models.SignalLocation{File: "src/auth/login.test.js"}},
			{Type: "flakyTest", Category: models.CategoryHealth, Severity: models.SeverityMedium, Location: models.SignalLocation{File: "src/auth/login.test.js"}},
		},
	}

	result := Propagate(resolver, snap)

	// Find team-auth aggregate.
	var authAgg *OwnerAggregate
	for i, agg := range result.Summary.Owners {
		if agg.Owner.ID == "team-auth" {
			authAgg = &result.Summary.Owners[i]
			break
		}
	}

	if authAgg == nil {
		t.Fatal("expected team-auth aggregate")
	}
	if authAgg.CodeUnitCount != 2 {
		t.Errorf("code units = %d, want 2", authAgg.CodeUnitCount)
	}
	if authAgg.ExportedCodeUnitCount != 2 {
		t.Errorf("exported code units = %d, want 2", authAgg.ExportedCodeUnitCount)
	}
	if authAgg.SignalCount != 2 {
		t.Errorf("signals = %d, want 2", authAgg.SignalCount)
	}
	if authAgg.CriticalSignalCount != 1 {
		t.Errorf("critical signals = %d, want 1", authAgg.CriticalSignalCount)
	}
	if authAgg.HealthSignalCount != 1 {
		t.Errorf("health signals = %d, want 1", authAgg.HealthSignalCount)
	}
}

func TestDeriveCoveragePosture(t *testing.T) {
	tests := []struct {
		total, owned int
		want         string
	}{
		{0, 0, "none"},
		{10, 0, "none"},
		{10, 2, "weak"},
		{10, 5, "partial"},
		{10, 8, "strong"},
		{10, 10, "strong"},
	}
	for _, tt := range tests {
		s := OwnershipSummary{TotalFiles: tt.total, OwnedFiles: tt.owned}
		got := deriveCoveragePosture(s)
		if got != tt.want {
			t.Errorf("deriveCoveragePosture(%d/%d) = %q, want %q", tt.owned, tt.total, got, tt.want)
		}
	}
}
