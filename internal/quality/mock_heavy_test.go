package quality

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestMockHeavyDetector_HighMockRatio(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/a.test.js", TestCount: 2, AssertionCount: 1, MockCount: 5},
		},
	}

	d := &MockHeavyDetector{}
	signals := d.Detect(snap)

	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}
	if signals[0].Type != "mockHeavyTest" {
		t.Errorf("type = %q, want mockHeavyTest", signals[0].Type)
	}
}

func TestMockHeavyDetector_LowMockCount(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/b.test.js", TestCount: 3, AssertionCount: 5, MockCount: 2},
		},
	}

	d := &MockHeavyDetector{}
	signals := d.Detect(snap)

	if len(signals) != 0 {
		t.Errorf("expected 0 signals for low mock count, got %d", len(signals))
	}
}

func TestMockHeavyDetector_ZeroMocks(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/c.test.js", TestCount: 5, AssertionCount: 10, MockCount: 0},
		},
	}

	d := &MockHeavyDetector{}
	signals := d.Detect(snap)

	if len(signals) != 0 {
		t.Errorf("expected 0 signals for zero mocks, got %d", len(signals))
	}
}

func TestMockHeavyDetector_NoAssertions(t *testing.T) {
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/d.test.js", TestCount: 2, AssertionCount: 0, MockCount: 4},
		},
	}

	d := &MockHeavyDetector{}
	signals := d.Detect(snap)

	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}
	if signals[0].Severity != models.SeverityHigh {
		t.Errorf("severity = %q, want high", signals[0].Severity)
	}
}
