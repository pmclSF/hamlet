package models

// KnownSignalTypes is the canonical signal vocabulary accepted by snapshot
// validation. Keep this in sync with internal/signals.
var KnownSignalTypes = map[SignalType]bool{
	"slowTest":               true,
	"flakyTest":              true,
	"skippedTest":            true,
	"deadTest":               true,
	"unstableSuite":          true,
	"untestedExport":         true,
	"weakAssertion":          true,
	"mockHeavyTest":          true,
	"testsOnlyMocks":         true,
	"snapshotHeavyTest":      true,
	"coverageBlindSpot":      true,
	"coverageThresholdBreak": true,
	"frameworkMigration":     true,
	"migrationBlocker":       true,
	"deprecatedTestPattern":  true,
	"dynamicTestGeneration":  true,
	"customMatcherRisk":      true,
	"unsupportedSetup":       true,
	"policyViolation":        true,
	"legacyFrameworkUsage":   true,
	"skippedTestsInCI":       true,
	"runtimeBudgetExceeded":  true,
}

// IsKnownSignalType reports whether t is part of Terrain's canonical catalog.
func IsKnownSignalType(t SignalType) bool {
	return KnownSignalTypes[t]
}
