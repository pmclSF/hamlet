package envdepth

// DepthClass classifies the environmental depth/realism of a test.
type DepthClass string

const (
	DepthHeavyMocking    DepthClass = "heavy_mocking"
	DepthModerateMocking DepthClass = "moderate_mocking"
	DepthRealDependency  DepthClass = "real_dependency_usage"
	DepthBrowserRuntime  DepthClass = "browser_runtime"
	DepthUnknown         DepthClass = "unknown"
)

// EnvironmentIndicator is a specific signal about test environment.
type EnvironmentIndicator string

const (
	IndicatorFakeClock      EnvironmentIndicator = "fake_clock"
	IndicatorStubbedNetwork EnvironmentIndicator = "stubbed_network"
	IndicatorInMemoryDB     EnvironmentIndicator = "in_memory_db"
	IndicatorBrowserDriver  EnvironmentIndicator = "browser_driver"
	IndicatorMockLibrary    EnvironmentIndicator = "mock_library"
	IndicatorRealHTTP       EnvironmentIndicator = "real_http"
)

// Assessment is the environment depth assessment for one test file.
type Assessment struct {
	FilePath    string
	Depth       DepthClass
	Indicators  []EnvironmentIndicator
	MockRatio   float64 // mocks / (mocks + assertions), 0.0-1.0
	Confidence  float64
	Explanation string
}

// AssessmentResult holds all environment depth assessments.
type AssessmentResult struct {
	Assessments  []Assessment
	ByDepth      map[DepthClass]int
	OverallDepth DepthClass
}
