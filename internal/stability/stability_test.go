package stability

import (
	"testing"

	"github.com/pmclSF/hamlet/internal/models"
)

func stableHistory(id string, n int) TestHistory {
	h := TestHistory{TestID: id, TestName: "stable test", FilePath: "test.js"}
	for i := 0; i < n; i++ {
		h.Observations = append(h.Observations, Observation{
			SnapshotIndex: i,
			Present:       true,
			Passed:        true,
			HasRuntime:    true,
		})
	}
	return h
}

func TestClassify_ConsistentlyStable(t *testing.T) {
	h := stableHistory("test-1", 5)
	result := Classify([]TestHistory{h})
	if len(result.Classifications) != 1 {
		t.Fatalf("expected 1 classification, got %d", len(result.Classifications))
	}
	c := result.Classifications[0]
	if c.Class != ClassConsistentlyStable {
		t.Errorf("class = %s, want consistently_stable", c.Class)
	}
	if c.Confidence < 0.7 {
		t.Errorf("confidence = %f, want >= 0.7", c.Confidence)
	}
}

func TestClassify_NewlyUnstable(t *testing.T) {
	h := TestHistory{TestID: "test-2", TestName: "unstable test", FilePath: "test.js"}
	// 6 stable observations, then 4 failures.
	for i := 0; i < 6; i++ {
		h.Observations = append(h.Observations, Observation{
			SnapshotIndex: i, Present: true, Passed: true, HasRuntime: true,
		})
	}
	for i := 6; i < 10; i++ {
		h.Observations = append(h.Observations, Observation{
			SnapshotIndex: i, Present: true, Failed: true, HasRuntime: true,
		})
	}
	result := Classify([]TestHistory{h})
	c := result.Classifications[0]
	if c.Class != ClassNewlyUnstable {
		t.Errorf("class = %s, want newly_unstable", c.Class)
	}
}

func TestClassify_ChronicallyFlaky(t *testing.T) {
	h := TestHistory{TestID: "test-3", TestName: "flaky test", FilePath: "test.js"}
	for i := 0; i < 5; i++ {
		h.Observations = append(h.Observations, Observation{
			SnapshotIndex: i, Present: true, FlakySignal: true, HasRuntime: true,
		})
	}
	result := Classify([]TestHistory{h})
	c := result.Classifications[0]
	if c.Class != ClassChronicallyFlaky {
		t.Errorf("class = %s, want chronically_flaky", c.Class)
	}
}

func TestClassify_IntermittentlySlow(t *testing.T) {
	h := TestHistory{TestID: "test-4", TestName: "slow test", FilePath: "test.js"}
	for i := 0; i < 5; i++ {
		obs := Observation{
			SnapshotIndex: i, Present: true, Passed: true, HasRuntime: true,
		}
		if i%2 == 0 {
			obs.SlowSignal = true
		}
		h.Observations = append(h.Observations, obs)
	}
	result := Classify([]TestHistory{h})
	c := result.Classifications[0]
	if c.Class != ClassIntermittentlySlow {
		t.Errorf("class = %s, want intermittently_slow", c.Class)
	}
}

func TestClassify_Improving(t *testing.T) {
	h := TestHistory{TestID: "test-5", TestName: "improving test", FilePath: "test.js"}
	// Early: failures and flaky
	for i := 0; i < 4; i++ {
		h.Observations = append(h.Observations, Observation{
			SnapshotIndex: i, Present: true, Failed: true, FlakySignal: true, HasRuntime: true,
		})
	}
	// Late: stable
	for i := 4; i < 8; i++ {
		h.Observations = append(h.Observations, Observation{
			SnapshotIndex: i, Present: true, Passed: true, HasRuntime: true,
		})
	}
	result := Classify([]TestHistory{h})
	c := result.Classifications[0]
	if c.Class != ClassImproving {
		t.Errorf("class = %s, want improving", c.Class)
	}
}

func TestClassify_QuarantinedSuppressed(t *testing.T) {
	h := TestHistory{TestID: "test-6", TestName: "quarantined test", FilePath: "test.js"}
	for i := 0; i < 5; i++ {
		h.Observations = append(h.Observations, Observation{
			SnapshotIndex: i, Present: true, Skipped: true,
		})
	}
	result := Classify([]TestHistory{h})
	c := result.Classifications[0]
	if c.Class != ClassQuarantinedSuppressed {
		t.Errorf("class = %s, want quarantined_or_suppressed", c.Class)
	}
}

func TestClassify_DataInsufficient(t *testing.T) {
	h := TestHistory{TestID: "test-7", TestName: "new test", FilePath: "test.js"}
	h.Observations = append(h.Observations, Observation{
		SnapshotIndex: 0, Present: true, Passed: true,
	})
	result := Classify([]TestHistory{h})
	c := result.Classifications[0]
	if c.Class != ClassDataInsufficient {
		t.Errorf("class = %s, want data_insufficient", c.Class)
	}
}

func TestClassify_EmptyHistories(t *testing.T) {
	result := Classify(nil)
	if len(result.Classifications) != 0 {
		t.Errorf("expected 0 classifications, got %d", len(result.Classifications))
	}
}

func TestBuildHistories(t *testing.T) {
	tc := models.TestCase{
		TestID:   "abc123",
		TestName: "should work",
		FilePath: "src/test.js",
	}
	snap1 := &models.TestSuiteSnapshot{
		TestCases: []models.TestCase{tc},
		TestFiles: []models.TestFile{
			{Path: "src/test.js", RuntimeStats: &models.RuntimeStats{
				AvgRuntimeMs: 100, PassRate: 1.0,
			}},
		},
	}
	snap2 := &models.TestSuiteSnapshot{
		TestCases: []models.TestCase{tc},
		TestFiles: []models.TestFile{
			{Path: "src/test.js", RuntimeStats: &models.RuntimeStats{
				AvgRuntimeMs: 200, PassRate: 0.8,
			}},
		},
	}

	histories := BuildHistories([]*models.TestSuiteSnapshot{snap1, snap2})
	if len(histories) != 1 {
		t.Fatalf("expected 1 history, got %d", len(histories))
	}
	if len(histories[0].Observations) != 2 {
		t.Errorf("expected 2 observations, got %d", len(histories[0].Observations))
	}
}

func TestTrend(t *testing.T) {
	// Improving trend
	obs := make([]Observation, 6)
	for i := 0; i < 3; i++ {
		obs[i] = Observation{Present: true, Failed: true, FlakySignal: true}
	}
	for i := 3; i < 6; i++ {
		obs[i] = Observation{Present: true, Passed: true}
	}
	if tr := trend(obs); tr != "improving" {
		t.Errorf("trend = %s, want improving", tr)
	}
}
