package quality

import (
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestStaticSkipDetector_JSSkipPatterns(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "tests/auth.test.ts", Framework: "vitest", TestCount: 10, SkipCount: 3},
			{Path: "tests/billing.test.ts", Framework: "vitest", TestCount: 5, SkipCount: 0},
		},
	}

	d := &StaticSkipDetector{}
	signals := d.Detect(snap)

	if len(signals) == 0 {
		t.Fatal("expected signals for 3 skipped tests")
	}

	// Should have repo-level signal + file-level signal for auth.test.ts
	if len(signals) != 2 {
		t.Errorf("expected 2 signals (repo + file), got %d", len(signals))
	}

	// Repo-level signal
	if signals[0].Type != "staticSkippedTest" {
		t.Errorf("expected type staticSkippedTest, got %s", signals[0].Type)
	}
	if signals[0].Location.File != "" {
		t.Error("repo-level signal should not have file location")
	}

	// File-level signal
	if signals[1].Location.File != "tests/auth.test.ts" {
		t.Errorf("expected auth.test.ts, got %s", signals[1].Location.File)
	}
}

func TestStaticSkipDetector_NoSkips(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "tests/a.test.ts", Framework: "jest", TestCount: 10, SkipCount: 0},
		},
	}

	d := &StaticSkipDetector{}
	signals := d.Detect(snap)

	if len(signals) != 0 {
		t.Errorf("expected 0 signals for no skips, got %d", len(signals))
	}
}

func TestStaticSkipDetector_SeverityThresholds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		skips    int
		total    int
		wantSev  models.SignalSeverity
	}{
		{"low skip rate", 1, 20, models.SeverityLow},
		{"medium skip rate", 5, 20, models.SeverityMedium},
		{"high skip rate", 12, 20, models.SeverityHigh},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snap := &models.TestSuiteSnapshot{
				TestFiles: []models.TestFile{
					{Path: "tests/a.test.ts", Framework: "jest", TestCount: tt.total, SkipCount: tt.skips},
				},
			}
			d := &StaticSkipDetector{}
			signals := d.Detect(snap)
			if len(signals) == 0 {
				t.Fatal("expected signals")
			}
			if signals[0].Severity != tt.wantSev {
				t.Errorf("expected severity %s, got %s", tt.wantSev, signals[0].Severity)
			}
		})
	}
}

func TestStaticSkipDetector_DeterministicOrder(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "tests/z.test.ts", Framework: "jest", TestCount: 10, SkipCount: 2},
			{Path: "tests/a.test.ts", Framework: "jest", TestCount: 10, SkipCount: 5},
			{Path: "tests/m.test.ts", Framework: "jest", TestCount: 10, SkipCount: 2},
		},
	}

	d := &StaticSkipDetector{}
	s1 := d.Detect(snap)
	s2 := d.Detect(snap)

	if len(s1) != len(s2) {
		t.Fatalf("non-deterministic: %d vs %d", len(s1), len(s2))
	}
	for i := range s1 {
		if s1[i].Location.File != s2[i].Location.File {
			t.Errorf("non-deterministic order at %d: %s vs %s",
				i, s1[i].Location.File, s2[i].Location.File)
		}
	}

	// a.test.ts (50% skip) should come before z/m (20% skip)
	if len(s1) >= 3 && s1[1].Location.File != "tests/a.test.ts" {
		t.Errorf("expected highest skip ratio first, got %s", s1[1].Location.File)
	}
}
