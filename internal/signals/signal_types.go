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
	SignalUnsupportedSetup      models.SignalType = "unsupportedSetup"
	SignalPolicyViolation        models.SignalType = "policyViolation"
	SignalLegacyFrameworkUsage   models.SignalType = "legacyFrameworkUsage"
	SignalSkippedTestsInCI       models.SignalType = "skippedTestsInCI"
	SignalRuntimeBudgetExceeded  models.SignalType = "runtimeBudgetExceeded"
)

// Canonical signal type sets. Import these rather than duplicating
// signal type maps across packages.

// MigrationSignalTypes is the canonical set of migration-related signal types.
var MigrationSignalTypes = map[models.SignalType]bool{
	SignalFrameworkMigration:    true,
	SignalMigrationBlocker:     true,
	SignalDeprecatedTestPattern: true,
	SignalDynamicTestGeneration: true,
	SignalCustomMatcherRisk:     true,
	SignalUnsupportedSetup:     true,
}

// QualitySignalTypes is the canonical set of quality-related signal types.
var QualitySignalTypes = map[models.SignalType]bool{
	SignalWeakAssertion:          true,
	SignalMockHeavyTest:          true,
	SignalUntestedExport:         true,
	SignalCoverageThresholdBreak: true,
	SignalCoverageBlindSpot:      true,
}

// IsMigrationSignal returns true if the signal type is migration-related.
func IsMigrationSignal(t models.SignalType) bool {
	return MigrationSignalTypes[t]
}

// IsQualitySignal returns true if the signal type is quality-related.
func IsQualitySignal(t models.SignalType) bool {
	return QualitySignalTypes[t]
}
