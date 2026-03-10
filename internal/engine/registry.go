// Package engine provides the analysis orchestration layer.
//
// It wires together detectors, ownership resolution, risk scoring,
// and governance evaluation into a single reusable pipeline.
package engine

import (
	"github.com/pmclSF/hamlet/internal/governance"
	"github.com/pmclSF/hamlet/internal/migration"
	"github.com/pmclSF/hamlet/internal/models"
	"github.com/pmclSF/hamlet/internal/policy"
	"github.com/pmclSF/hamlet/internal/quality"
	"github.com/pmclSF/hamlet/internal/signals"
)

// Config holds runtime configuration needed to construct detectors.
type Config struct {
	// RepoRoot is the repository root path (required for file-reading detectors).
	RepoRoot string

	// PolicyConfig is the loaded policy configuration (nil if no policy file).
	PolicyConfig *policy.Config
}

// DefaultRegistry returns a DetectorRegistry populated with all
// standard Hamlet detectors in the correct execution order.
//
// The order matters: governance detectors depend on signals from
// quality and migration detectors, so they are registered last.
func DefaultRegistry(cfg Config) *signals.DetectorRegistry {
	r := signals.NewRegistry()

	// Quality detectors (no dependencies on other signals).
	r.MustRegister(signals.DetectorRegistration{
		Meta: signals.DetectorMeta{
			ID:           "quality.weak-assertion",
			Domain:       signals.DomainQuality,
			EvidenceType: signals.EvidenceStructuralPattern,
			Description:  "Detect test files with weak or missing assertions.",
			SignalTypes:  []models.SignalType{signals.SignalWeakAssertion},
		},
		Detector: &quality.WeakAssertionDetector{},
	})
	r.MustRegister(signals.DetectorRegistration{
		Meta: signals.DetectorMeta{
			ID:           "quality.mock-heavy",
			Domain:       signals.DomainQuality,
			EvidenceType: signals.EvidenceStructuralPattern,
			Description:  "Detect test files with excessive mock usage.",
			SignalTypes:  []models.SignalType{signals.SignalMockHeavyTest},
		},
		Detector: &quality.MockHeavyDetector{},
	})
	r.MustRegister(signals.DetectorRegistration{
		Meta: signals.DetectorMeta{
			ID:           "quality.snapshot-heavy",
			Domain:       signals.DomainQuality,
			EvidenceType: signals.EvidenceStructuralPattern,
			Description:  "Detect test files that over-rely on snapshot assertions.",
			SignalTypes:  []models.SignalType{signals.SignalSnapshotHeavyTest},
		},
		Detector: &quality.SnapshotHeavyDetector{},
	})
	r.MustRegister(signals.DetectorRegistration{
		Meta: signals.DetectorMeta{
			ID:           "quality.untested-export",
			Domain:       signals.DomainQuality,
			EvidenceType: signals.EvidencePathName,
			Description:  "Detect exported code units without matching test files.",
			SignalTypes:  []models.SignalType{signals.SignalUntestedExport},
		},
		Detector: &quality.UntestedExportDetector{},
	})
	r.MustRegister(signals.DetectorRegistration{
		Meta: signals.DetectorMeta{
			ID:             "quality.coverage-threshold",
			Domain:         signals.DomainCoverage,
			EvidenceType:   signals.EvidenceCoverage,
			Description:    "Detect coverage below configured thresholds.",
			SignalTypes:    []models.SignalType{signals.SignalCoverageThresholdBreak},
			RequiresFileIO: true,
		},
		Detector: &quality.CoverageThresholdDetector{},
	})
	r.MustRegister(signals.DetectorRegistration{
		Meta: signals.DetectorMeta{
			ID:           "coverage.blind-spot",
			Domain:       signals.DomainCoverage,
			EvidenceType: signals.EvidenceCoverage,
			Description:  "Detect coverage lineage blind spots across discovered code units.",
			SignalTypes:  []models.SignalType{signals.SignalCoverageBlindSpot},
		},
		Detector: &quality.CoverageBlindSpotDetector{},
	})

	// Migration detectors (no dependencies on other signals).
	r.MustRegister(signals.DetectorRegistration{
		Meta: signals.DetectorMeta{
			ID:             "migration.deprecated-pattern",
			Domain:         signals.DomainMigration,
			EvidenceType:   signals.EvidenceStructuralPattern,
			Description:    "Detect deprecated test patterns that block migration.",
			SignalTypes:    []models.SignalType{signals.SignalDeprecatedTestPattern},
			RequiresFileIO: true,
		},
		Detector: &migration.DeprecatedPatternDetector{RepoRoot: cfg.RepoRoot},
	})
	r.MustRegister(signals.DetectorRegistration{
		Meta: signals.DetectorMeta{
			ID:             "migration.dynamic-test-generation",
			Domain:         signals.DomainMigration,
			EvidenceType:   signals.EvidenceStructuralPattern,
			Description:    "Detect dynamic test generation patterns.",
			SignalTypes:    []models.SignalType{signals.SignalDynamicTestGeneration},
			RequiresFileIO: true,
		},
		Detector: &migration.DynamicTestGenerationDetector{RepoRoot: cfg.RepoRoot},
	})
	r.MustRegister(signals.DetectorRegistration{
		Meta: signals.DetectorMeta{
			ID:             "migration.custom-matcher",
			Domain:         signals.DomainMigration,
			EvidenceType:   signals.EvidenceStructuralPattern,
			Description:    "Detect custom matchers that complicate migration.",
			SignalTypes:    []models.SignalType{signals.SignalCustomMatcherRisk},
			RequiresFileIO: true,
		},
		Detector: &migration.CustomMatcherDetector{RepoRoot: cfg.RepoRoot},
	})
	r.MustRegister(signals.DetectorRegistration{
		Meta: signals.DetectorMeta{
			ID:             "migration.unsupported-setup",
			Domain:         signals.DomainMigration,
			EvidenceType:   signals.EvidenceStructuralPattern,
			Description:    "Detect framework-specific setup/fixture patterns.",
			SignalTypes:    []models.SignalType{signals.SignalUnsupportedSetup},
			RequiresFileIO: true,
		},
		Detector: &migration.UnsupportedSetupDetector{RepoRoot: cfg.RepoRoot},
	})
	r.MustRegister(signals.DetectorRegistration{
		Meta: signals.DetectorMeta{
			ID:           "migration.framework-migration",
			Domain:       signals.DomainMigration,
			EvidenceType: signals.EvidenceStructuralPattern,
			Description:  "Detect multi-framework repos suitable for migration.",
			SignalTypes:  []models.SignalType{signals.SignalFrameworkMigration},
		},
		Detector: &migration.FrameworkMigrationDetector{},
	})

	// Governance detectors (depend on signals from quality/migration detectors).
	if cfg.PolicyConfig != nil && !cfg.PolicyConfig.IsEmpty() {
		r.MustRegister(signals.DetectorRegistration{
			Meta: signals.DetectorMeta{
				ID:               "governance.policy",
				Domain:           signals.DomainGovernance,
				EvidenceType:     signals.EvidencePolicy,
				Description:      "Evaluate repository state against local policy rules.",
				SignalTypes:      []models.SignalType{signals.SignalPolicyViolation, signals.SignalLegacyFrameworkUsage, signals.SignalRuntimeBudgetExceeded},
				DependsOnSignals: true,
			},
			Detector: &GovernanceDetector{Config: cfg.PolicyConfig},
		})
	}

	return r
}

// GovernanceDetector wraps governance.Evaluate as a signals.Detector.
//
// It must run after quality and migration detectors because some policy
// checks reference their signal counts.
type GovernanceDetector struct {
	Config *policy.Config
}

// Detect implements signals.Detector.
func (d *GovernanceDetector) Detect(snap *models.TestSuiteSnapshot) []models.Signal {
	result := governance.Evaluate(snap, d.Config)
	return result.Violations
}
