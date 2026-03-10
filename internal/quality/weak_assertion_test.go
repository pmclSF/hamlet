package quality

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestWeakAssertionDetector_NoAssertions(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/a.test.js", TestCount: 3, AssertionCount: 0},
		},
	}

	d := &WeakAssertionDetector{}
	signals := d.Detect(snap)

	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}
	if signals[0].Type != "weakAssertion" {
		t.Errorf("type = %q, want weakAssertion", signals[0].Type)
	}
	if signals[0].Severity != models.SeverityHigh {
		t.Errorf("severity = %q, want high", signals[0].Severity)
	}
}

func TestWeakAssertionDetector_LowDensity(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/b.test.js", TestCount: 5, AssertionCount: 3},
		},
	}

	d := &WeakAssertionDetector{}
	signals := d.Detect(snap)

	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}
	if signals[0].Severity != models.SeverityMedium {
		t.Errorf("severity = %q, want medium", signals[0].Severity)
	}
}

func TestWeakAssertionDetector_GoodDensity(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/good.test.js", TestCount: 5, AssertionCount: 10},
		},
	}

	d := &WeakAssertionDetector{}
	signals := d.Detect(snap)

	if len(signals) != 0 {
		t.Errorf("expected 0 signals for good density, got %d", len(signals))
	}
}

func TestWeakAssertionDetector_NoTests(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/empty.test.js", TestCount: 0, AssertionCount: 0},
		},
	}

	d := &WeakAssertionDetector{}
	signals := d.Detect(snap)

	if len(signals) != 0 {
		t.Errorf("expected 0 signals for file with no tests, got %d", len(signals))
	}
}

func TestWeakAssertionDetector_SnapshotDominatedLowDensity(t *testing.T) {
	t.Parallel()

	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{
				Path:           "test/snapshot-heavy.test.js",
				TestCount:      10,
				AssertionCount: 4,
				SnapshotCount:  4, // 100% snapshot assertions, density 0.4
			},
		},
	}

	d := &WeakAssertionDetector{}
	signals := d.Detect(snap)
	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}
	if signals[0].Severity != models.SeverityInfo {
		t.Fatalf("severity = %q, want info", signals[0].Severity)
	}
	if signals[0].Type != "weakAssertion" {
		t.Fatalf("type = %q, want weakAssertion", signals[0].Type)
	}
}
