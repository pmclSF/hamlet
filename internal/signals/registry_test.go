package signals

import (
	"testing"

	"github.com/pmclSF/terrain/internal/models"
)

func TestRegistryContainsAllSignalTypes(t *testing.T) {
	t.Parallel()
	expected := []models.SignalType{
		SignalSlowTest,
		SignalFlakyTest,
		SignalSkippedTest,
		SignalDeadTest,
		SignalUnstableSuite,
		SignalUntestedExport,
		SignalWeakAssertion,
		SignalMockHeavyTest,
		SignalTestsOnlyMocks,
		SignalSnapshotHeavyTest,
		SignalCoverageBlindSpot,
		SignalCoverageThresholdBreak,
		SignalFrameworkMigration,
		SignalMigrationBlocker,
		SignalDeprecatedTestPattern,
		SignalDynamicTestGeneration,
		SignalCustomMatcherRisk,
		SignalPolicyViolation,
		SignalLegacyFrameworkUsage,
		SignalSkippedTestsInCI,
		SignalRuntimeBudgetExceeded,
	}

	for _, st := range expected {
		def, ok := Registry[st]
		if !ok {
			t.Errorf("Registry missing signal type %q", st)
			continue
		}
		if def.Type != st {
			t.Errorf("Registry[%q].Type = %q, want %q", st, def.Type, st)
		}
		if def.Title == "" {
			t.Errorf("Registry[%q].Title is empty", st)
		}
		if def.Description == "" {
			t.Errorf("Registry[%q].Description is empty", st)
		}
		if def.Category == "" {
			t.Errorf("Registry[%q].Category is empty", st)
		}
	}

	if len(Registry) != len(expected) {
		t.Errorf("Registry has %d entries, expected %d", len(Registry), len(expected))
	}
}

func TestRegistryCategoryCounts(t *testing.T) {
	t.Parallel()
	counts := map[models.SignalCategory]int{}
	for _, def := range Registry {
		counts[def.Category]++
	}

	if counts[models.CategoryHealth] != 5 {
		t.Errorf("health signals = %d, want 5", counts[models.CategoryHealth])
	}
	if counts[models.CategoryQuality] != 7 {
		t.Errorf("quality signals = %d, want 7", counts[models.CategoryQuality])
	}
	if counts[models.CategoryMigration] != 5 {
		t.Errorf("migration signals = %d, want 5", counts[models.CategoryMigration])
	}
	if counts[models.CategoryGovernance] != 4 {
		t.Errorf("governance signals = %d, want 4", counts[models.CategoryGovernance])
	}
}
