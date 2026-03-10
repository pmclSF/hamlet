package quality

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func TestSnapshotHeavyDetector_DetectsHighAndMediumCases(t *testing.T) {
	t.Parallel()
	snap := &models.TestSuiteSnapshot{
		TestFiles: []models.TestFile{
			{Path: "test/incidental.test.js", SnapshotCount: 1, AssertionCount: 4},
			{Path: "test/medium.test.js", SnapshotCount: 3, AssertionCount: 4},
			{Path: "test/high-no-assert.test.js", SnapshotCount: 4, AssertionCount: 0},
			{Path: "test/high-ratio.test.js", SnapshotCount: 6, AssertionCount: 2},
		},
	}

	d := &SnapshotHeavyDetector{}
	signals := d.Detect(snap)
	if len(signals) != 3 {
		t.Fatalf("expected 3 snapshotHeavyTest signals, got %d", len(signals))
	}

	byFile := map[string]models.Signal{}
	for _, s := range signals {
		byFile[s.Location.File] = s
		if s.Type != "snapshotHeavyTest" {
			t.Fatalf("unexpected signal type %q", s.Type)
		}
	}

	if _, ok := byFile["test/incidental.test.js"]; ok {
		t.Fatal("incidental snapshot usage should not be flagged")
	}
	if s, ok := byFile["test/medium.test.js"]; !ok || s.Severity != models.SeverityMedium {
		t.Fatalf("expected medium severity for medium.test.js, got %#v", s)
	}
	if s, ok := byFile["test/high-no-assert.test.js"]; !ok || s.Severity != models.SeverityHigh {
		t.Fatalf("expected high severity for high-no-assert.test.js, got %#v", s)
	}
	if s, ok := byFile["test/high-ratio.test.js"]; !ok || s.Severity != models.SeverityHigh {
		t.Fatalf("expected high severity for high-ratio.test.js, got %#v", s)
	}
}
