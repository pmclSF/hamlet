package suppression

// SuppressionKind classifies the type of suppression mechanism.
type SuppressionKind string

const (
	KindQuarantined     SuppressionKind = "quarantined"
	KindExpectedFailure SuppressionKind = "expected_failure"
	KindSkipDisable     SuppressionKind = "skip_disable"
	KindRetryWrapper    SuppressionKind = "retry_wrapper"
)

// SuppressionIntent classifies whether suppression is tactical or chronic.
type SuppressionIntent string

const (
	IntentTactical SuppressionIntent = "tactical"   // Recent, likely temporary
	IntentChronic  SuppressionIntent = "chronic"     // Persistent, likely debt
	IntentUnknown  SuppressionIntent = "unknown"
)

// DetectionSource describes how the suppression was inferred.
type DetectionSource string

const (
	SourceAnnotation  DetectionSource = "annotation"        // Test tag/annotation
	SourceConfig      DetectionSource = "config_file"       // CI/test config
	SourceNaming      DetectionSource = "naming_convention"  // skip/quarantine in name
	SourceRuntimeData DetectionSource = "runtime_data"      // Observed skip/retry patterns
	SourceSignal      DetectionSource = "signal"            // Existing Hamlet signal
)

// Suppression represents a detected suppression mechanism on a test.
type Suppression struct {
	// TestFilePath is the file containing the suppressed test.
	TestFilePath string
	// TestName is the specific test name, if identifiable.
	TestName string
	// Kind is the type of suppression.
	Kind SuppressionKind
	// Intent classifies tactical vs chronic suppression.
	Intent SuppressionIntent
	// Source describes how this was detected.
	Source DetectionSource
	// Confidence is 0.0-1.0 reflecting detection confidence.
	Confidence float64
	// Explanation is a human-readable description.
	Explanation string
	// Metadata holds additional context.
	Metadata map[string]any
}

// SuppressionResult holds all detected suppressions for a snapshot.
type SuppressionResult struct {
	Suppressions []Suppression
	// Counts by kind.
	QuarantinedCount     int
	ExpectedFailureCount int
	SkipDisableCount     int
	RetryWrapperCount    int
	// Counts by intent.
	TacticalCount int
	ChronicCount  int
	UnknownCount  int
	// TotalSuppressedTests is the number of unique test files with any suppression.
	TotalSuppressedTests int
}
