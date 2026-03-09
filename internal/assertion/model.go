package assertion

// StrengthClass classifies assertion strength for a test file.
type StrengthClass string

const (
	StrengthStrong   StrengthClass = "strong"
	StrengthModerate StrengthClass = "moderate"
	StrengthWeak     StrengthClass = "weak"
	StrengthUnclear  StrengthClass = "unclear"
)

// AssertionCategory classifies the type of assertion observed.
type AssertionCategory string

const (
	CategoryBehavioral AssertionCategory = "behavioral"  // Tests specific behavior/return values
	CategorySnapshot   AssertionCategory = "snapshot"     // toMatchSnapshot, toMatchInlineSnapshot
	CategoryExistence  AssertionCategory = "existence"    // toBeDefined, toExist, not null checks
	CategoryStatus     AssertionCategory = "status"       // Status code checks (200, 404)
	CategoryType       AssertionCategory = "type_check"   // instanceof, typeof checks
	CategoryException  AssertionCategory = "exception"    // toThrow, rejects
	CategoryUnknown    AssertionCategory = "unknown"
)

// Assessment is the assertion strength assessment for one test file.
type Assessment struct {
	// FilePath is the test file assessed.
	FilePath string
	// Strength is the overall assertion strength classification.
	Strength StrengthClass
	// AssertionCount is the total assertion count from static analysis.
	AssertionCount int
	// TestCount is the total test count.
	TestCount int
	// Density is assertions per test.
	Density float64
	// Categories breaks down assertions by type.
	Categories map[AssertionCategory]int
	// DominantCategory is the most common assertion type.
	DominantCategory AssertionCategory
	// Confidence is 0.0-1.0 reflecting assessment confidence.
	Confidence float64
	// Explanation describes the assessment.
	Explanation string
}

// AssessmentResult holds assertion strength for all test files.
type AssessmentResult struct {
	Assessments []Assessment
	// ByStrength counts files by strength class.
	ByStrength map[StrengthClass]int
	// OverallStrength is the aggregate strength classification.
	OverallStrength StrengthClass
	// AverageDensity is the mean assertions per test across all files.
	AverageDensity float64
}
