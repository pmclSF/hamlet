package signals

import "github.com/pmclSF/terrain/internal/models"

// Definition describes a known signal type in the Terrain registry.
//
// The registry is the canonical catalog of signals supported by the product.
// It exists to keep code, docs, UI copy, and future scoring/policy logic
// aligned around the same signal vocabulary.
type Definition struct {
	Type        models.SignalType     `json:"type"`
	Category    models.SignalCategory `json:"category"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
}

// Registry contains the current signal catalog.
//
// This should remain in sync with docs/signal-catalog.md.
var Registry = map[models.SignalType]Definition{
	SignalSlowTest: {
		Type:        SignalSlowTest,
		Category:    models.CategoryHealth,
		Title:       "Slow Test",
		Description: "A test or suite consistently exceeds an expected runtime threshold.",
	},
	SignalFlakyTest: {
		Type:        SignalFlakyTest,
		Category:    models.CategoryHealth,
		Title:       "Flaky Test",
		Description: "A test demonstrates intermittent failures or elevated retry behavior.",
	},
	SignalSkippedTest: {
		Type:        SignalSkippedTest,
		Category:    models.CategoryHealth,
		Title:       "Skipped Test",
		Description: "A test is disabled, skipped, or pending.",
	},
	SignalDeadTest: {
		Type:        SignalDeadTest,
		Category:    models.CategoryHealth,
		Title:       "Dead Test",
		Description: "A test appears disconnected from live behavior, modules, or execution paths.",
	},
	SignalUnstableSuite: {
		Type:        SignalUnstableSuite,
		Category:    models.CategoryHealth,
		Title:       "Unstable Suite",
		Description: "A suite exhibits unusually high variance or inconsistency as a group.",
	},
	SignalUntestedExport: {
		Type:        SignalUntestedExport,
		Category:    models.CategoryQuality,
		Title:       "Untested Export",
		Description: "A public code unit appears to have weak or missing direct test coverage.",
	},
	SignalWeakAssertion: {
		Type:        SignalWeakAssertion,
		Category:    models.CategoryQuality,
		Title:       "Weak Assertion",
		Description: "A test has low or weak assertion strength relative to its scope.",
	},
	SignalMockHeavyTest: {
		Type:        SignalMockHeavyTest,
		Category:    models.CategoryQuality,
		Title:       "Mock-Heavy Test",
		Description: "A test relies heavily on mocks relative to real interactions.",
	},
	SignalTestsOnlyMocks: {
		Type:        SignalTestsOnlyMocks,
		Category:    models.CategoryQuality,
		Title:       "Tests Only Mocks",
		Description: "Assertions primarily validate mock interactions rather than business outcomes.",
	},
	SignalSnapshotHeavyTest: {
		Type:        SignalSnapshotHeavyTest,
		Category:    models.CategoryQuality,
		Title:       "Snapshot-Heavy Test",
		Description: "A test file depends heavily on snapshots relative to direct semantic assertions.",
	},
	SignalCoverageBlindSpot: {
		Type:        SignalCoverageBlindSpot,
		Category:    models.CategoryQuality,
		Title:       "Coverage Blind Spot",
		Description: "Coverage exists, but high-risk paths remain weakly exercised.",
	},
	SignalCoverageThresholdBreak: {
		Type:        SignalCoverageThresholdBreak,
		Category:    models.CategoryQuality,
		Title:       "Coverage Threshold Break",
		Description: "Coverage is below a declared threshold.",
	},
	SignalFrameworkMigration: {
		Type:        SignalFrameworkMigration,
		Category:    models.CategoryMigration,
		Title:       "Framework Migration Opportunity",
		Description: "The repository or package appears suitable for migration to a target framework.",
	},
	SignalMigrationBlocker: {
		Type:        SignalMigrationBlocker,
		Category:    models.CategoryMigration,
		Title:       "Migration Blocker",
		Description: "A pattern makes automated or safe migration difficult.",
	},
	SignalDeprecatedTestPattern: {
		Type:        SignalDeprecatedTestPattern,
		Category:    models.CategoryMigration,
		Title:       "Deprecated Test Pattern",
		Description: "A test pattern is outdated or poorly aligned with future standards.",
	},
	SignalDynamicTestGeneration: {
		Type:        SignalDynamicTestGeneration,
		Category:    models.CategoryMigration,
		Title:       "Dynamic Test Generation",
		Description: "Dynamic generation patterns reduce migration predictability.",
	},
	SignalCustomMatcherRisk: {
		Type:        SignalCustomMatcherRisk,
		Category:    models.CategoryMigration,
		Title:       "Custom Matcher Risk",
		Description: "Custom matchers or wrappers complicate portability and migration safety.",
	},
	SignalPolicyViolation: {
		Type:        SignalPolicyViolation,
		Category:    models.CategoryGovernance,
		Title:       "Policy Violation",
		Description: "Current repository state violates declared Terrain policy.",
	},
	SignalLegacyFrameworkUsage: {
		Type:        SignalLegacyFrameworkUsage,
		Category:    models.CategoryGovernance,
		Title:       "Legacy Framework Usage",
		Description: "Legacy or disallowed framework usage persists or is reintroduced.",
	},
	SignalSkippedTestsInCI: {
		Type:        SignalSkippedTestsInCI,
		Category:    models.CategoryGovernance,
		Title:       "Skipped Tests In CI",
		Description: "Skipped tests are present where CI policy disallows them.",
	},
	SignalRuntimeBudgetExceeded: {
		Type:        SignalRuntimeBudgetExceeded,
		Category:    models.CategoryGovernance,
		Title:       "Runtime Budget Exceeded",
		Description: "Tests or suites exceed configured runtime budgets.",
	},
}
