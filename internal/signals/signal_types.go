package signals

import "github.com/pmclSF/hamlet/internal/models"

// Convenience aliases so callers can use signals.SignalSlowTest etc.
// The underlying type is models.SignalType.
const (
	SignalSlowTest               models.SignalType = "slowTest"
	SignalFlakyTest              models.SignalType = "flakyTest"
	SignalSkippedTest            models.SignalType = "skippedTest"
	SignalDeadTest               models.SignalType = "deadTest"
	SignalUnstableSuite          models.SignalType = "unstableSuite"
	SignalUntestedExport         models.SignalType = "untestedExport"
	SignalWeakAssertion          models.SignalType = "weakAssertion"
	SignalMockHeavyTest          models.SignalType = "mockHeavyTest"
	SignalTestsOnlyMocks         models.SignalType = "testsOnlyMocks"
	SignalSnapshotHeavyTest      models.SignalType = "snapshotHeavyTest"
	SignalCoverageBlindSpot      models.SignalType = "coverageBlindSpot"
	SignalCoverageThresholdBreak models.SignalType = "coverageThresholdBreak"
	SignalFrameworkMigration     models.SignalType = "frameworkMigration"
	SignalMigrationBlocker       models.SignalType = "migrationBlocker"
	SignalDeprecatedTestPattern  models.SignalType = "deprecatedTestPattern"
	SignalDynamicTestGeneration  models.SignalType = "dynamicTestGeneration"
	SignalCustomMatcherRisk      models.SignalType = "customMatcherRisk"
	SignalPolicyViolation        models.SignalType = "policyViolation"
	SignalLegacyFrameworkUsage   models.SignalType = "legacyFrameworkUsage"
	SignalSkippedTestsInCI       models.SignalType = "skippedTestsInCI"
	SignalRuntimeBudgetExceeded  models.SignalType = "runtimeBudgetExceeded"
)
