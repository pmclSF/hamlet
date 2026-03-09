package stability

// StabilityClass classifies a test's historical stability pattern.
type StabilityClass string

const (
	ClassConsistentlyStable    StabilityClass = "consistently_stable"
	ClassNewlyUnstable         StabilityClass = "newly_unstable"
	ClassChronicallyFlaky      StabilityClass = "chronically_flaky"
	ClassIntermittentlySlow    StabilityClass = "intermittently_slow"
	ClassImproving             StabilityClass = "improving"
	ClassQuarantinedSuppressed StabilityClass = "quarantined_or_suppressed"
	ClassDataInsufficient      StabilityClass = "data_insufficient"
)

// MinHistoryDepth is the minimum number of snapshot observations
// required before classifying beyond data_insufficient.
const MinHistoryDepth = 3

// TestHistory represents longitudinal observations for a single test.
type TestHistory struct {
	TestID    string
	TestName  string
	FilePath  string
	Framework string
	Owner     string

	// Observations ordered from oldest to newest.
	Observations []Observation
}

// Observation captures a single snapshot's data for one test.
type Observation struct {
	// SnapshotIndex is the ordinal position (0 = oldest).
	SnapshotIndex int
	// Present is true if the test existed in this snapshot.
	Present bool
	// Passed is true if the test passed (only meaningful if Present && HasRuntime).
	Passed bool
	// Failed is true if the test failed.
	Failed bool
	// Skipped is true if the test was skipped/disabled.
	Skipped bool
	// RuntimeMs is the average runtime, if available.
	RuntimeMs float64
	// HasRuntime is true if runtime data was available.
	HasRuntime bool
	// RetryRate is the retry rate, if available.
	RetryRate float64
	// FlakySignal is true if a flakyTest signal was present.
	FlakySignal bool
	// SlowSignal is true if a slowTest signal was present.
	SlowSignal bool
}

// Classification is the result of stability classification for one test.
type Classification struct {
	TestID       string
	TestName     string
	FilePath     string
	Owner        string
	Class        StabilityClass
	Confidence   float64
	HistoryDepth int
	Explanation  string
	// RecentTrend describes the direction of recent observations.
	RecentTrend string // "improving", "worsening", "stable", "insufficient"
}

// ClassificationResult holds stability classes for all tests.
type ClassificationResult struct {
	Classifications []Classification
	// ByClass groups test IDs by stability class.
	ByClass map[StabilityClass]int
	// HistoryDepth is the number of snapshots analyzed.
	HistoryDepth int
}
