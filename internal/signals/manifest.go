package signals

import (
	"sort"

	"github.com/pmclSF/terrain/internal/models"
)

// SignalStatus expresses the lifecycle stage of a signal type.
type SignalStatus string

const (
	// StatusStable: at least one production detector emits this signal,
	// it has documented severity/confidence semantics, and the schema is locked.
	StatusStable SignalStatus = "stable"

	// StatusExperimental: detector exists and may emit, but its severity and
	// confidence semantics may still change.
	StatusExperimental SignalStatus = "experimental"

	// StatusPlanned: signal type is declared but no detector emits it today.
	// Documented to reserve the name and shape so future detectors don't
	// invent overlapping types. References from policy or measurement code
	// short-circuit to zero counts.
	StatusPlanned SignalStatus = "planned"
)

// SignalTier separates gate-relevant detectors (CI-blocking)
// from observability detectors (target silent failure modes that PR-revert
// proxy cannot measure, never block CI). Some detectors target failure
// classes the proxy cannot score but that still matter
// (aiEmbeddingModelChange, aiSafetyEvalMissing, uncoveredAISurface); the
// observability tier exists so those detectors can ship without abusing
// severity ladders.
//
// Severity-from-lift logic respects this tier:
//   - TierGate detectors: declared severity may be demoted per the lift CI
//     ladder; they gate CI (`--fail-on=high` selects them).
//   - TierObservability detectors: lift evidence informs explain output but
//     does NOT demote severity (since lift can't measure their failure
//     mode); severity is capped at Medium so they never gate CI.
//   - Tier is REQUIRED on every manifest entry. The contract is enforced
//     by TestManifest_AllEntriesHaveExplicitTier. Empty Tier is a build
//     error so no detector can silently fall through to either side of
//     the gate/observability split.
type SignalTier string

const (
	// TierGate detectors target code regressions that produce revert/hotfix-
	// shaped failures. The CI gate fires on these. Examples: blastRadiusHotspot,
	// aiModelDeprecationRisk, depsDriftRisk.
	TierGate SignalTier = "gate"

	// TierObservability detectors target structural conditions for *silent*
	// quality degradation (eval-score drift, hallucination-rate creep,
	// embedding-model-without-reindex). They never produce revert/hotfix
	// patterns because the failure mode is gradual. Validated by sampled
	// review and public-incident matching, NOT by PR-lift. Severity
	// capped at Medium; never gate-relevant. Examples:
	// aiSafetyEvalMissing, uncoveredAISurface, aiEmbeddingModelChange,
	// aiPromptVersioning.
	TierObservability SignalTier = "observability"
)

// ManifestEntry is the canonical record for a signal type. Every signal
// declared in signal_types.go must have a matching entry here, and every
// entry here must reference a real signal-type constant. Drift between the
// two is caught by TestManifest_MatchesSignalTypes.
//
// The manifest replaces three older registration layers over time:
//   - Registry (registry.go): superset; will be regenerated from this manifest
//   - typeInfoBySignal (signal_types.go): description/remediation pairs
//   - docs/signal-catalog.md: hand-edited list with persistent drift
type ManifestEntry struct {
	// Type is the canonical signal type string emitted in snapshots and JSON.
	Type models.SignalType

	// ConstName is the Go constant name (e.g. "SignalWeakAssertion"). Used by
	// the drift linter to validate one-to-one mapping with signal_types.go.
	ConstName string

	// Domain is the high-level category the signal belongs to. Maps to the
	// long-standing models.SignalCategory enum.
	Domain models.SignalCategory

	// Status: stable / experimental / planned.
	Status SignalStatus

	// Title is a short human-readable name (Title Case).
	Title string

	// Description is the one-line user-facing explanation. Pulled from
	// signal_types.go's typeInfoBySignal where present.
	Description string

	// Remediation is the suggested action to take.
	Remediation string

	// DefaultSeverity is the severity the producing detector emits in the
	// typical case. Detectors retain authority to escalate or de-escalate
	// per finding; this field documents the expected baseline.
	DefaultSeverity models.SignalSeverity

	// ConfidenceMin / ConfidenceMax bracket the typical confidence range
	// the detector emits. Today's values are descriptive (sourced from
	// detector code review) rather than tuned.
	ConfidenceMin float64
	ConfidenceMax float64

	// EvidenceSources lists the data inputs the detector consults.
	// Values: structural-pattern, path-name, runtime, coverage,
	// policy, codeowners, graph-traversal.
	EvidenceSources []string

	// RuleID is a stable identifier for documentation cross-references and
	// SARIF emission. Format: terrain/<category>/<rule-name>.
	RuleID string

	// RuleURI points to the canonical rule documentation page as an
	// in-repo path under docs/rules/. The explain surface renders it as a
	// GitHub blob URL; there is no external docs domain today.
	RuleURI string

	// PromotionPlan describes what is required to advance the entry's
	// status. Populated for experimental and planned entries; empty for
	// stable.
	PromotionPlan string

	// Tier classifies the detector as gate (CI-blocking)
	// or observability (silent-failure-mode, never blocks CI). Empty
	// defaults to TierObservability — a detector must opt in to
	// TierGate explicitly. See SignalTier comment.
	Tier SignalTier

	// DisabledByDefault marks a detector as off in the default config.
	// Used for detectors that need a structural redesign before they
	// are reliable enough to gate: the implementation stays in tree
	// so a future fix or an opt-in user can exercise it, but the
	// pipeline does not emit its findings unless the user opts in via
	// .terrain/policy.yaml.
	DisabledByDefault bool
}

// DefaultDisabledTypes returns the set of signal types disabled by
// default at the manifest level. The engine's emission path consults
// this and skips emission for any signal type in the set unless the
// user has explicitly enabled it via policy config.
func DefaultDisabledTypes() map[string]bool {
	out := map[string]bool{}
	for _, e := range allSignalManifest {
		if e.DisabledByDefault {
			out[string(e.Type)] = true
		}
	}
	return out
}

// allSignalManifest is the canonical inventory. Order is significant for
// generated docs; do not sort. New entries go at the end of their domain
// section to keep RuleIDs stable.
//
// Convention for adding entries:
//  1. Add the constant to signal_types.go.
//  2. Add the manifest entry below in domain order.
//  3. If a detector emits it, set Status = StatusStable.
//  4. Otherwise mark StatusPlanned with a PromotionPlan that names the
//     milestone in docs/release/ that will ship the detector.
//  5. Run `go test ./internal/signals/... -run TestManifest`.
var allSignalManifest = []ManifestEntry{
	// ── Health ─────────────────────────────────────────────────
	{
		Type: SignalSlowTest, ConstName: "SignalSlowTest",
		Domain: models.CategoryHealth, Status: StatusStable,
		Title:           "Slow Test",
		Description:     "Tests exceed expected runtime budget and slow feedback loops.",
		Remediation:     "Profile slow paths and split or optimize expensive tests.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.85, ConfidenceMax: 0.95,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/health/slow-test",
		RuleURI:         "docs/rules/health/slow-test.md",
		Tier:            TierObservability,
	},
	{
		Type: SignalFlakyTest, ConstName: "SignalFlakyTest",
		Domain: models.CategoryHealth, Status: StatusStable,
		Title:           "Flaky Test",
		Description:     "Tests exhibit inconsistent pass/fail behavior across runs.",
		Remediation:     "Stabilize timing, shared state, and external dependency handling.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.7, ConfidenceMax: 0.85,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/health/flaky-test",
		RuleURI:         "docs/rules/health/flaky-test.md",
		PromotionPlan:   "Observability tier.",
		Tier:            TierObservability,
	},
	{
		Type: SignalSkippedTest, ConstName: "SignalSkippedTest",
		Domain: models.CategoryHealth, Status: StatusStable,
		Title:           "Skipped Test",
		Description:     "Tests are skipped and may hide latent regressions.",
		Remediation:     "Unskip, remove, or explicitly justify skipped tests in policy.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.85, ConfidenceMax: 0.95,
		EvidenceSources: []string{"runtime", "structural-pattern"},
		RuleID:          "terrain/health/skipped-test",
		RuleURI:         "docs/rules/health/skipped-test.md",
		Tier:            TierObservability,
	},
	{
		Type: SignalDeadTest, ConstName: "SignalDeadTest",
		Domain: models.CategoryHealth, Status: StatusStable,
		Title:           "Dead Test",
		Description:     "Tests may no longer validate meaningful behavior.",
		Remediation:     "Remove obsolete tests or reconnect them to active behavior.",
		DefaultSeverity: models.SeverityLow,
		ConfidenceMin:   0.6, ConfidenceMax: 0.8,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/health/dead-test",
		RuleURI:         "docs/rules/health/dead-test.md",
		Tier:            TierObservability,
	},
	{
		Type: SignalUnstableSuite, ConstName: "SignalUnstableSuite",
		Domain: models.CategoryHealth, Status: StatusStable,
		Title:           "Unstable Suite",
		Description:     "The suite has concentrated instability signals.",
		Remediation:     "Prioritize stabilization in the highest-instability areas.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.7, ConfidenceMax: 0.85,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/health/unstable-suite",
		RuleURI:         "docs/rules/health/unstable-suite.md",
		Tier:            TierObservability,
	},

	// ── Quality ────────────────────────────────────────────────
	{
		Type: SignalUntestedExport, ConstName: "SignalUntestedExport",
		Domain: models.CategoryQuality, Status: StatusStable,
		Title:           "Untested Export",
		Description:     "Exported code units are not directly covered by tests.",
		Remediation:     "Add direct tests for public exports to protect API behavior.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.5, ConfidenceMax: 0.7,
		EvidenceSources: []string{"path-name", "graph-traversal"},
		RuleID:          "terrain/coverage/untested-export",
		RuleURI:         "docs/rules/coverage/untested-export.md",
		Tier:            TierGate,
	},
	{
		Type: SignalWeakAssertion, ConstName: "SignalWeakAssertion",
		Domain: models.CategoryQuality, Status: StatusStable,
		Title:           "Weak Assertion",
		Description:     "Tests use weak or low-density assertions, reducing defect-catching power.",
		Remediation:     "Add behavior-focused assertions on outputs, state transitions, and side effects.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.4, ConfidenceMax: 0.8,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/hygiene/weak-assertion",
		RuleURI:         "docs/rules/hygiene/weak-assertion.md",
		// "Is this assertion strong enough?" is a value judgment with no
		// stable threshold across teams or framing. Preserved at
		// observability; not appropriate for CI gating without an
		// explicit user-declared policy threshold.
		Tier:          TierObservability,
		PromotionPlan: "Observability tier.",
	},
	{
		Type: SignalMockHeavyTest, ConstName: "SignalMockHeavyTest",
		Domain: models.CategoryQuality, Status: StatusExperimental,
		Title:           "Mock-Heavy Test",
		Description:     "Tests rely heavily on mocks and may miss integration-level regressions.",
		Remediation:     "Replace brittle mocks with real collaborators where practical.",
		DefaultSeverity: models.SeverityLow,
		// Demoted to experimental + low severity: raw mock count does not
		// distinguish brittle mocks from well-designed test doubles. A
		// rebuild needs a different underlying signal (e.g. mock-target
		// diversity, or a module-boundary vs callback-spy distinction).
		ConfidenceMin: 0.3, ConfidenceMax: 0.5,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/hygiene/mock-heavy",
		RuleURI:         "docs/rules/hygiene/mock-heavy.md",
		// Whether mocks outnumbering assertions is a defect or a design
		// choice is stylistic. Kept at the observability tier; a rebuild
		// would distinguish module-boundary mocks from callback-spy stubs.
		Tier:          TierObservability,
		PromotionPlan: "Observability tier.",
	},
	{
		Type: SignalTestsOnlyMocks, ConstName: "SignalTestsOnlyMocks",
		Domain: models.CategoryQuality, Status: StatusStable,
		Title:           "Tests Only Mocks",
		Description:     "Test files contain mock setup but zero assertions, verifying wiring only.",
		Remediation:     "Add assertions on outputs, state changes, or side effects to validate real behavior.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.7, ConfidenceMax: 0.9,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/quality/tests-only-mocks",
		RuleURI:         "docs/rules/quality/tests-only-mocks.md",
		// True "wiring-only" tests are a low-base-rate phenomenon, so
		// the per-file signal is structurally near-silent. Re-frame as
		// a repo-aggregate posture metric rather than a per-file
		// finding.
		Tier:          TierObservability,
		PromotionPlan: "Observability tier.",
	},
	{
		Type: SignalSnapshotHeavyTest, ConstName: "SignalSnapshotHeavyTest",
		Domain: models.CategoryQuality, Status: StatusStable,
		Title:           "Snapshot-Heavy Test",
		Description:     "Test files over-rely on snapshot assertions, reducing defect specificity.",
		Remediation:     "Supplement snapshots with targeted assertions on critical behavior.",
		DefaultSeverity: models.SeverityLow,
		ConfidenceMin:   0.5, ConfidenceMax: 0.75,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/hygiene/snapshot-heavy",
		RuleURI:         "docs/rules/hygiene/snapshot-heavy.md",
		// Whether snapshot use is heavy enough to flag is a value
		// judgment. Kept at the observability tier.
		Tier:          TierObservability,
		PromotionPlan: "Observability tier.",
	},
	{
		Type: SignalCoverageBlindSpot, ConstName: "SignalCoverageBlindSpot",
		Domain: models.CategoryQuality, Status: StatusStable,
		Title:           "Coverage Blind Spot",
		Description:     "Code units appear unprotected or weakly protected by current coverage mix.",
		Remediation:     "Add unit/integration tests where only broad or indirect coverage exists.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.5, ConfidenceMax: 0.8,
		EvidenceSources: []string{"coverage", "graph-traversal"},
		RuleID:          "terrain/coverage/blind-spot",
		RuleURI:         "docs/rules/coverage/blind-spot.md",
		Tier:            TierGate,
	},
	{
		Type: SignalCoverageThresholdBreak, ConstName: "SignalCoverageThresholdBreak",
		Domain: models.CategoryQuality, Status: StatusStable,
		Title:           "Coverage Threshold Break",
		Description:     "Measured coverage falls below configured thresholds.",
		Remediation:     "Target low-coverage, high-risk areas and raise meaningful coverage first.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.9, ConfidenceMax: 0.99,
		EvidenceSources: []string{"coverage"},
		RuleID:          "terrain/quality/coverage-threshold",
		RuleURI:         "docs/rules/quality/coverage-threshold.md",
		PromotionPlan:   "Stable.",
		Tier:            TierGate,
	},
	{
		Type: SignalStaticSkippedTest, ConstName: "SignalStaticSkippedTest",
		Domain: models.CategoryQuality, Status: StatusStable,
		Title:           "Static Skipped Test",
		Description:     "Tests are statically marked as skipped (it.skip, xit, @skip, etc.).",
		Remediation:     "Re-enable, replace, or document skip markers older than the policy threshold.",
		DefaultSeverity: models.SeverityLow,
		ConfidenceMin:   0.85, ConfidenceMax: 0.95,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/hygiene/permanently-skipped",
		RuleURI:         "docs/rules/hygiene/permanently-skipped.md",
		Tier:            TierGate,
	},
	{
		Type: SignalStaticSkippedTestUnconditional, ConstName: "SignalStaticSkippedTestUnconditional",
		Domain: models.CategoryQuality, Status: StatusPlanned,
		Title:           "Static Skipped Test — Unconditional",
		Description:     "A test is statically marked as skipped without any surrounding environment / feature-flag gate. The skip is permanent until the marker is removed.",
		Remediation:     "Re-enable, replace, or delete the test. Add a comment explaining why if the skip should persist.",
		DefaultSeverity: models.SeverityLow,
		ConfidenceMin:   0.9, ConfidenceMax: 0.95,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/hygiene/static-skip-unconditional",
		RuleURI:         "docs/rules/hygiene/static-skip-unconditional.md",
		PromotionPlan:   "Off by default; opt in via .terrain/policy.yaml.",
		// Tier deliberately observability while StatusPlanned. The
		// parent SignalStaticSkippedTest is TierGate; this split child
		// stays out of the gate decision until Status moves to
		// StatusStable.
		Tier: TierObservability,
	},
	{
		Type: SignalStaticSkippedTestConditionalGate, ConstName: "SignalStaticSkippedTestConditionalGate",
		Domain: models.CategoryQuality, Status: StatusPlanned,
		Title:           "Conditionally-Skipped Test (informational)",
		Description:     "A test is statically marked as skipped, but the skip is wrapped by an environment, feature-flag, or platform predicate. The skip is intentional and gated by code. This finding is informational — no action is required unless the gate condition itself is wrong.",
		Remediation:     "Audit the gate condition periodically. CI should run the test on platforms or branches where the gate evaluates to false.",
		DefaultSeverity: models.SeverityLow,
		ConfidenceMin:   0.75, ConfidenceMax: 0.9,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/hygiene/static-skip-conditional-gate",
		RuleURI:         "docs/rules/hygiene/static-skip-conditional-gate.md",
		PromotionPlan:   "Off by default; opt in via .terrain/policy.yaml.",
		// Informational by design — see the Description. Tier stays
		// observability even after promotion to StatusStable.
		Tier: TierObservability,
	},
	{
		Type: SignalAssertionFreeTest, ConstName: "SignalAssertionFreeTest",
		Domain: models.CategoryQuality, Status: StatusStable,
		Title:           "Assertion-Free Test",
		Description:     "Test files contain test function signatures but no detectable assertions.",
		Remediation:     "Add assertions to validate behavior — tests without assertions verify nothing.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.75, ConfidenceMax: 0.9,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/hygiene/no-assertions",
		RuleURI:         "docs/rules/hygiene/no-assertions.md",
		// The assertion counter is blind to several real assertion
		// dialects (self.assertX, np.testing.*, mock.assert_called_*),
		// which limits how reliably it fires until a multi-dialect
		// oracle plus a path-role gate are in place.
		Tier:          TierObservability,
		PromotionPlan: "Observability tier.",
	},
	{
		Type: SignalOrphanedTestFile, ConstName: "SignalOrphanedTestFile",
		Domain: models.CategoryQuality, Status: StatusStable,
		Title:           "Orphaned Test File",
		Description:     "Test files do not import any source modules from the repository.",
		Remediation:     "Connect orphaned tests to source code or remove if obsolete.",
		DefaultSeverity: models.SeverityLow,
		ConfidenceMin:   0.4, ConfidenceMax: 0.7,
		EvidenceSources: []string{"graph-traversal"},
		RuleID:          "terrain/hygiene/orphaned-test",
		RuleURI:         "docs/rules/hygiene/orphaned-test.md",
		Tier:            TierObservability,
	},
	{
		Type: SignalDepsDriftRisk, ConstName: "SignalDepsDriftRisk",
		Domain: models.CategoryQuality, Status: StatusExperimental,
		Title:           "Dependency Drift Risk",
		Description:     "A dependency manifest has a high share of moving-target version specs (caret, tilde, *, latest), making the repo silently susceptible to upstream regressions.",
		Remediation:     "Pin versions or add a lockfile-verification gate. Re-audit the manifest after pinning to confirm the moving-target share drops below the threshold.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.55, ConfidenceMax: 0.85,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/deps/drift-risk",
		RuleURI:         "docs/rules/deps/drift-risk.md",
		PromotionPlan:   "Off by default; opt in via .terrain/policy.yaml.",
		Tier:            TierGate,
	},
	{
		Type: SignalDepsDriftRiskStrictPin, ConstName: "SignalDepsDriftRiskStrictPin",
		Domain: models.CategoryQuality, Status: StatusPlanned,
		Title:           "Unpinned Dependency",
		Description:     "One or more dependencies are declared without an explicit version anchor (bare name, `*`, `latest`, or unversioned URL). The resolver picks whatever happens to be available at install time, so installs are not reproducible across runs.",
		Remediation:     "Add an explicit version, version range, or lockfile-verification gate so installs are reproducible.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.7, ConfidenceMax: 0.9,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/deps/drift-strict-pin",
		RuleURI:         "docs/rules/deps/drift-strict-pin.md",
		PromotionPlan:   "Off by default; opt in via .terrain/policy.yaml.",
		// Tier deliberately observability while StatusPlanned. The
		// parent SignalDepsDriftRisk is TierGate; this split child
		// stays out of the gate decision until promotion to
		// StatusStable.
		Tier: TierObservability,
	},
	{
		Type: SignalDepsDriftRiskCaretPolicy, ConstName: "SignalDepsDriftRiskCaretPolicy",
		Domain: models.CategoryQuality, Status: StatusPlanned,
		Title:           "Caret-Range Dependency Drift",
		Description:     "Dependencies use caret-range specs (`^x.y.z`) in an ecosystem where caret semantics let minor versions drift silently (npm, Poetry, and Cargo each interpret caret differently). The runtime version can change without a manifest edit.",
		Remediation:     "Adopt a stricter pinning policy (exact or commit-pinned) where minor-version drift would silently affect runtime behavior.",
		DefaultSeverity: models.SeverityLow,
		ConfidenceMin:   0.55, ConfidenceMax: 0.8,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/deps/drift-caret-policy",
		RuleURI:         "docs/rules/deps/drift-caret-policy.md",
		PromotionPlan:   "Off by default; opt in via .terrain/policy.yaml.",
		// Tier deliberately observability while StatusPlanned. See the
		// strict-pin counterpart above for rationale.
		Tier: TierObservability,
	},
	{
		Type: SignalConfigSchemaDrift, ConstName: "SignalConfigSchemaDrift",
		Domain: models.CategoryQuality, Status: StatusExperimental,
		Title:           "Config Schema Drift Risk",
		Description:     "An infra-config file (GitHub Actions workflow, docker-compose, Helm values, or k8s manifest) uses forward-compat hazards: mutable action refs, `:latest` or untagged image tags, deprecated apiVersions.",
		Remediation:     "Pin action refs to a SHA or tagged release. Replace `:latest` and untagged images with explicit versions. Upgrade deprecated apiVersions to their current replacement.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.5, ConfidenceMax: 0.8,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/config/schema-drift",
		RuleURI:         "docs/rules/config/schema-drift.md",
		PromotionPlan:   "Off by default; opt in via .terrain/policy.yaml.",
		Tier:            TierGate,
	},
	{
		Type: SignalPromptFileMissingEval, ConstName: "SignalPromptFileMissingEval",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "AI/ML Surface Without Eval Coverage",
		Description:     "An AI/ML surface (prompt, agent, tool definition, model context, or model artifact) has no eval scenario covering it — the dominant AI-testing failure mode.",
		Remediation:     "Add an eval scenario (promptfoo / DeepEval / Ragas / framework-specific) that exercises this surface. Use `terrain ai list` to see other uncovered surfaces in the same repo and batch-fix.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.55, ConfidenceMax: 0.85,
		EvidenceSources: []string{"graph-traversal"},
		RuleID:          "terrain/ai/surface-missing-eval",
		RuleURI:         "docs/rules/ai/surface-missing-eval.md",
		PromotionPlan:   "Observability tier.",
		Tier:            TierObservability,
	},

	// ── Migration ──────────────────────────────────────────────
	{
		Type: SignalFrameworkMigration, ConstName: "SignalFrameworkMigration",
		Domain: models.CategoryMigration, Status: StatusStable,
		Title:           "Framework Migration Opportunity",
		Description:     "The repository or package appears suitable for migration to a target framework.",
		Remediation:     "Evaluate candidates with `terrain migration readiness` and plan staged framework_migration.",
		DefaultSeverity: models.SeverityInfo,
		ConfidenceMin:   0.5, ConfidenceMax: 0.8,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/migration/framework-migration",
		RuleURI:         "docs/rules/migration/framework-migration.md",
		Tier:            TierGate,
	},
	{
		Type: SignalMigrationBlocker, ConstName: "SignalMigrationBlocker",
		Domain: models.CategoryMigration, Status: StatusStable,
		Title:           "Migration Blocker",
		Description:     "Detected patterns will complicate framework framework_migration.",
		Remediation:     "Address blockers incrementally before broad migration changes.",
		DefaultSeverity: models.SeverityLow,
		ConfidenceMin:   0.7, ConfidenceMax: 0.9,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/migration/migration-blocker",
		RuleURI:         "docs/rules/migration/migration-blocker.md",
		// The enzyme-usage sub-rule no longer matches modern repos —
		// enzyme as a migration target is historical. Capability
		// preserved by refreshing the
		// trigger set to living migration patterns (mocha->jest,
		// jasmine->jest, unittest->pytest).
		Tier:          TierObservability,
		PromotionPlan: "Observability tier.",
	},
	{
		Type: SignalDeprecatedTestPattern, ConstName: "SignalDeprecatedTestPattern",
		Domain: models.CategoryMigration, Status: StatusStable,
		Title:           "Deprecated Test Pattern",
		Description:     "Deprecated test patterns increase migration and maintenance risk.",
		Remediation:     "Replace deprecated APIs with supported alternatives.",
		DefaultSeverity: models.SeverityLow,
		ConfidenceMin:   0.7, ConfidenceMax: 0.9,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/migration/deprecated-pattern",
		RuleURI:         "docs/rules/migration/deprecated-pattern.md",
		// Two sub-rules need refresh: the enzyme-usage trigger no longer
		// matches modern repos and should be retired or refreshed; the
		// setTimeout-in-test trigger needs a scope/binding gate to
		// distinguish jest.setTimeout config from bare setTimeout in a
		// test body.
		Tier:          TierObservability,
		PromotionPlan: "Observability tier.",
	},
	{
		Type: SignalDynamicTestGeneration, ConstName: "SignalDynamicTestGeneration",
		Domain: models.CategoryMigration, Status: StatusStable,
		Title:           "Dynamic Test Generation",
		Description:     "Dynamic test generation may reduce migration and analysis confidence.",
		Remediation:     "Prefer explicit, static test declarations for critical paths.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.5, ConfidenceMax: 0.75,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/migration/dynamic-generation",
		RuleURI:         "docs/rules/migration/dynamic-generation.md",
		Tier:            TierObservability,
	},
	{
		Type: SignalCustomMatcherRisk, ConstName: "SignalCustomMatcherRisk",
		Domain: models.CategoryMigration, Status: StatusStable,
		Title:           "Custom Matcher Risk",
		Description:     "Custom matcher behavior can be difficult to migrate safely.",
		Remediation:     "Audit matcher semantics and provide migration-safe equivalents.",
		DefaultSeverity: models.SeverityLow,
		ConfidenceMin:   0.4, ConfidenceMax: 0.7,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/migration/custom-matcher",
		RuleURI:         "docs/rules/migration/custom-matcher.md",
		Tier:            TierObservability,
	},
	{
		Type: SignalUnsupportedSetup, ConstName: "SignalUnsupportedSetup",
		Domain: models.CategoryMigration, Status: StatusStable,
		Title:           "Unsupported Setup",
		Description:     "Setup/teardown patterns may not port cleanly to target frameworks.",
		Remediation:     "Refactor setup boundaries toward framework-agnostic patterns.",
		DefaultSeverity: models.SeverityLow,
		ConfidenceMin:   0.4, ConfidenceMax: 0.7,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/migration/unsupported-setup",
		RuleURI:         "docs/rules/migration/unsupported-setup.md",
		Tier:            TierObservability,
	},

	// ── Governance ─────────────────────────────────────────────
	{
		Type: SignalPolicyViolation, ConstName: "SignalPolicyViolation",
		Domain: models.CategoryGovernance, Status: StatusStable,
		Title:           "Policy Violation",
		Description:     "Repository state violates configured Terrain policy rules.",
		Remediation:     "Resolve violations or intentionally update policy thresholds.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   1.0, ConfidenceMax: 1.0,
		EvidenceSources: []string{"policy"},
		RuleID:          "terrain/governance/policy-violation",
		RuleURI:         "docs/rules/governance/policy-violation.md",
		Tier:            TierGate,
	},
	{
		Type: SignalLegacyFrameworkUsage, ConstName: "SignalLegacyFrameworkUsage",
		Domain: models.CategoryGovernance, Status: StatusStable,
		Title:           "Legacy Framework Usage",
		Description:     "Legacy framework usage remains where policy discourages it.",
		Remediation:     "Plan and execute incremental migration away from legacy frameworks.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   1.0, ConfidenceMax: 1.0,
		EvidenceSources: []string{"policy", "structural-pattern"},
		RuleID:          "terrain/governance/legacy-framework",
		RuleURI:         "docs/rules/governance/legacy-framework.md",
		Tier:            TierGate,
	},
	{
		Type: SignalSkippedTestsInCI, ConstName: "SignalSkippedTestsInCI",
		Domain: models.CategoryGovernance, Status: StatusStable,
		Title:           "Skipped Tests In CI",
		Description:     "Skipped tests are present where CI policy disallows them.",
		Remediation:     "Investigate skip conditions and re-enable tests or replace with targeted alternatives.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   1.0, ConfidenceMax: 1.0,
		EvidenceSources: []string{"policy", "structural-pattern"},
		RuleID:          "terrain/governance/skipped-in-ci",
		RuleURI:         "docs/rules/governance/skipped-in-ci.md",
		Tier:            TierGate,
	},
	{
		Type: SignalRuntimeBudgetExceeded, ConstName: "SignalRuntimeBudgetExceeded",
		Domain: models.CategoryGovernance, Status: StatusStable,
		Title:           "Runtime Budget Exceeded",
		Description:     "Observed runtimes exceed configured policy budget.",
		Remediation:     "Reduce runtime hotspots or adjust policy to reflect intentional tradeoffs.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   1.0, ConfidenceMax: 1.0,
		EvidenceSources: []string{"policy", "runtime"},
		RuleID:          "terrain/governance/runtime-budget",
		RuleURI:         "docs/rules/governance/runtime-budget.md",
		Tier:            TierGate,
	},

	// ── Structural (graph-powered) ─────────────────────────────
	{
		Type: SignalUncoveredAISurface, ConstName: "SignalUncoveredAISurface",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "Uncovered AI Surface",
		Description:     "AI surfaces (prompts, tools, datasets) have zero test or scenario coverage.",
		Remediation:     "Add eval scenarios that exercise this AI surface — untested prompts and tools can change behavior silently.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   0.7, ConfidenceMax: 0.9,
		EvidenceSources: []string{"graph-traversal", "structural-pattern"},
		RuleID:          "terrain/structural/uncovered-ai-surface",
		RuleURI:         "docs/rules/structural/uncovered-ai-surface.md",
		PromotionPlan:   "Observability tier.",
		Tier:            TierObservability,
	},
	{
		Type: SignalPhantomEvalScenario, ConstName: "SignalPhantomEvalScenario",
		Domain: models.CategoryAI, Status: StatusStable,
		Title:           "Phantom Eval Scenario",
		Description:     "Eval scenarios claim to validate AI surfaces but have no import-graph path to those surfaces — typically caused by a prompt/surface rename that wasn't propagated to the eval YAML.",
		Remediation:     "Verify the test file actually imports and exercises the target code, or correct the surface mapping.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   0.6, ConfidenceMax: 0.85,
		EvidenceSources: []string{"graph-traversal"},
		RuleID:          "terrain/structural/phantom-eval",
		RuleURI:         "docs/rules/structural/phantom-eval.md",
		PromotionPlan:   "Observability tier.",
		Tier:            TierObservability,
	},
	{
		Type: SignalUntestedPromptFlow, ConstName: "SignalUntestedPromptFlow",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "Untested Prompt Flow",
		Description:     "A prompt flows through multiple source files via imports with zero test coverage at any point in the chain.",
		Remediation:     "Add integration tests at the prompt's consumption points to catch behavioral regressions.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   0.6, ConfidenceMax: 0.85,
		EvidenceSources: []string{"graph-traversal"},
		RuleID:          "terrain/structural/untested-prompt-flow",
		RuleURI:         "docs/rules/structural/untested-prompt-flow.md",
		PromotionPlan:   "Observability tier.",
		Tier:            TierObservability,
	},
	{
		Type: SignalBlastRadiusHotspot, ConstName: "SignalBlastRadiusHotspot",
		Domain: models.CategoryStructure, Status: StatusStable,
		Title:           "Blast-Radius Hotspot",
		Description:     "Source files where a change would impact an unusually large number of tests.",
		Remediation:     "Ensure high direct test coverage and consider adding contract tests at interface boundaries.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.7, ConfidenceMax: 0.9,
		EvidenceSources: []string{"graph-traversal"},
		RuleID:          "terrain/structural/blast-radius",
		RuleURI:         "docs/rules/structural/blast-radius.md",
		Tier:            TierGate,
	},
	{
		Type: SignalFixtureFragilityHotspot, ConstName: "SignalFixtureFragilityHotspot",
		Domain: models.CategoryStructure, Status: StatusStable,
		Title:           "Fixture Fragility Hotspot",
		Description:     "Fixtures depended on by many tests, where a single change cascades widely.",
		Remediation:     "Extract smaller, focused fixtures to reduce cascading test failures.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.7, ConfidenceMax: 0.9,
		EvidenceSources: []string{"graph-traversal"},
		RuleID:          "terrain/structural/fixture-fragility",
		RuleURI:         "docs/rules/structural/fixture-fragility.md",
		Tier:            TierObservability,
	},
	{
		Type: SignalAssertionFreeImport, ConstName: "SignalAssertionFreeImport",
		Domain: models.CategoryStructure, Status: StatusStable,
		Title:           "Assertion-Free Import",
		Description:     "Test files import production code but contain zero assertions — exercising code without verifying behavior.",
		Remediation:     "Add assertions to validate behavior or remove tests that verify nothing.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.8, ConfidenceMax: 0.95,
		EvidenceSources: []string{"graph-traversal", "structural-pattern"},
		RuleID:          "terrain/structural/assertion-free-import",
		RuleURI:         "docs/rules/structural/assertion-free-import.md",
		// Shares the same assertion-counter blindness as the other
		// assertion detectors — misses self.assertX, np.testing,
		// mock.assert_called_*, and fluent helpers. Inherited-base-class
		// assertions also require cross-file resolution to fully fix.
		Tier:          TierObservability,
		PromotionPlan: "Observability tier.",
	},
	{
		Type: SignalCapabilityValidationGap, ConstName: "SignalCapabilityValidationGap",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "Capability Validation Gap",
		Description:     "Inferred AI capabilities have no eval scenarios validating them.",
		Remediation:     "Add eval scenarios that exercise this capability to ensure behavioral regression detection.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.5, ConfidenceMax: 0.8,
		EvidenceSources: []string{"graph-traversal", "structural-pattern"},
		RuleID:          "terrain/structural/capability-gap",
		RuleURI:         "docs/rules/structural/capability-gap.md",
		PromotionPlan:   "Observability tier.",
		Tier:            TierObservability,
	},

	// ── AI / Eval ──────────────────────────────────────────
	// All entries below are referenced by policy and measurement code so
	// that future detector wiring requires no plumbing change. Until then,
	// counts are zero and StatusPlanned is documented in feature-status.md.
	{
		Type: SignalEvalFailure, ConstName: "SignalEvalFailure",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title:           "Eval Failure",
		Description:     "An AI eval scenario reported a hard failure.",
		Remediation:     "Investigate the failing case in the eval framework's report and patch the prompt or guardrail.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   0.9, ConfidenceMax: 1.0,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/eval-failure",
		RuleURI:         "docs/rules/ai/eval-failure.md",
		// The airun eval-framework adapters (Promptfoo, DeepEval, Ragas)
		// emit per-case failure data into the snapshot's EvalRuns, but
		// the standalone evalFailure detector did not ship — failures
		// Today's per-case failures surface via more specific
		// detectors (hallucination-rate, cost-regression,
		// retrieval-regression).
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierGate,
	},
	{
		Type: SignalEvalRegression, ConstName: "SignalEvalRegression",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "Eval Regression",
		Description:     "An eval case's primary Score dropped from baseline to current past the configured threshold, OR the run's PrimaryMetric dropped across all matched cases. Identifies regressions before merge.",
		Remediation:     "Inspect the diff for prompt / model / retrieval changes that affect the regressing case(s). If intentional, update the baseline with `terrain ai record`.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   0.85, ConfidenceMax: 0.99,
		EvidenceSources:   []string{"eval-execution"},
		RuleID:            "terrain/regression/eval-regression",
		RuleURI:           "docs/rules/regression/eval-regression.md",
		Tier:              TierGate,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},
	{
		Type: SignalAccuracyRegression, ConstName: "SignalAccuracyRegression",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Accuracy Regression", DefaultSeverity: models.SeverityHigh,
		ConfidenceMin: 0.85, ConfidenceMax: 0.95,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/accuracy-regression", RuleURI: "docs/rules/ai/accuracy-regression.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierGate,
	},
	{
		Type: SignalCitationMissing, ConstName: "SignalCitationMissing",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Citation Missing", DefaultSeverity: models.SeverityMedium,
		ConfidenceMin: 0.6, ConfidenceMax: 0.85,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/citation-missing", RuleURI: "docs/rules/ai/citation-missing.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierGate,
	},
	{
		Type: SignalRetrievalMiss, ConstName: "SignalRetrievalMiss",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Retrieval Miss", DefaultSeverity: models.SeverityHigh,
		ConfidenceMin: 0.7, ConfidenceMax: 0.9,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/retrieval-miss", RuleURI: "docs/rules/ai/retrieval-miss.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierGate,
	},
	{
		Type: SignalAnswerGroundingFailure, ConstName: "SignalAnswerGroundingFailure",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Answer Grounding Failure", DefaultSeverity: models.SeverityHigh,
		ConfidenceMin: 0.7, ConfidenceMax: 0.9,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/grounding-failure", RuleURI: "docs/rules/ai/grounding-failure.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierGate,
	},
	{
		Type: SignalToolSelectionError, ConstName: "SignalToolSelectionError",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Tool Selection Error", DefaultSeverity: models.SeverityHigh,
		ConfidenceMin: 0.7, ConfidenceMax: 0.9,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/tool-selection-error", RuleURI: "docs/rules/ai/tool-selection-error.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierGate,
	},
	{
		Type: SignalSchemaParseFailure, ConstName: "SignalSchemaParseFailure",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Schema Parse Failure", DefaultSeverity: models.SeverityHigh,
		ConfidenceMin: 0.85, ConfidenceMax: 0.95,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/schema-parse-failure", RuleURI: "docs/rules/ai/schema-parse-failure.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierGate,
	},
	{
		Type: SignalSafetyFailure, ConstName: "SignalSafetyFailure",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Safety Failure", DefaultSeverity: models.SeverityCritical,
		ConfidenceMin: 0.9, ConfidenceMax: 1.0,
		EvidenceSources: []string{"runtime", "policy"},
		RuleID:          "terrain/ai/safety-failure", RuleURI: "docs/rules/ai/safety-failure.md",
		// The structural counterpart aiSafetyEvalMissing already
		// ships; this runtime variant fires when an eval framework
		// explicitly grades a case as a safety violation.
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierGate,
	},
	{
		Type: SignalAIPolicyViolation, ConstName: "SignalAIPolicyViolation",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "AI Policy Violation", DefaultSeverity: models.SeverityCritical,
		ConfidenceMin: 1.0, ConfidenceMax: 1.0,
		EvidenceSources: []string{"policy"},
		RuleID:          "terrain/ai/ai-policy-violation", RuleURI: "docs/rules/ai/ai-policy-violation.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierGate,
	},
	{
		Type: SignalHallucinationDetected, ConstName: "SignalHallucinationDetected",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Hallucination Detected", DefaultSeverity: models.SeverityCritical,
		ConfidenceMin: 0.6, ConfidenceMax: 0.85,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/hallucination", RuleURI: "docs/rules/ai/hallucination.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierGate,
	},
	{
		Type: SignalLatencyRegression, ConstName: "SignalLatencyRegression",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Latency Regression", DefaultSeverity: models.SeverityMedium,
		ConfidenceMin: 0.85, ConfidenceMax: 0.95,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/latency-regression", RuleURI: "docs/rules/ai/latency-regression.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierGate,
	},
	{
		Type: SignalCostRegression, ConstName: "SignalCostRegression",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Cost Regression (umbrella)", DefaultSeverity: models.SeverityMedium,
		ConfidenceMin: 0.85, ConfidenceMax: 0.95,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/cost-regression-umbrella", RuleURI: "docs/rules/ai/cost-regression-umbrella.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierGate,
	},
	{
		Type: SignalContextOverflowRisk, ConstName: "SignalContextOverflowRisk",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Context Overflow Risk", DefaultSeverity: models.SeverityMedium,
		ConfidenceMin: 0.6, ConfidenceMax: 0.85,
		EvidenceSources: []string{"structural-pattern", "runtime"},
		RuleID:          "terrain/ai/context-overflow", RuleURI: "docs/rules/ai/context-overflow.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierObservability,
	},
	{
		Type: SignalWrongSourceSelected, ConstName: "SignalWrongSourceSelected",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Wrong Source Selected", DefaultSeverity: models.SeverityMedium,
		ConfidenceMin: 0.6, ConfidenceMax: 0.85,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/wrong-source", RuleURI: "docs/rules/ai/wrong-source.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierObservability,
	},
	{
		Type: SignalCitationMismatch, ConstName: "SignalCitationMismatch",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Citation Mismatch", DefaultSeverity: models.SeverityHigh,
		ConfidenceMin: 0.7, ConfidenceMax: 0.9,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/citation-mismatch", RuleURI: "docs/rules/ai/citation-mismatch.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierGate,
	},
	{
		Type: SignalStaleSourceRisk, ConstName: "SignalStaleSourceRisk",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Stale Source Risk", DefaultSeverity: models.SeverityMedium,
		ConfidenceMin: 0.5, ConfidenceMax: 0.8,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/ai/stale-source", RuleURI: "docs/rules/ai/stale-source.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierObservability,
	},
	{
		Type: SignalChunkingRegression, ConstName: "SignalChunkingRegression",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Chunking Regression", DefaultSeverity: models.SeverityMedium,
		ConfidenceMin: 0.7, ConfidenceMax: 0.9,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/chunking-regression", RuleURI: "docs/rules/ai/chunking-regression.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierObservability,
	},
	{
		Type: SignalRerankerRegression, ConstName: "SignalRerankerRegression",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Reranker Regression", DefaultSeverity: models.SeverityHigh,
		ConfidenceMin: 0.7, ConfidenceMax: 0.9,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/reranker-regression", RuleURI: "docs/rules/ai/reranker-regression.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierGate,
	},
	{
		Type: SignalTopKRegression, ConstName: "SignalTopKRegression",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Top-K Regression", DefaultSeverity: models.SeverityMedium,
		ConfidenceMin: 0.7, ConfidenceMax: 0.9,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/topk-regression", RuleURI: "docs/rules/ai/topk-regression.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierObservability,
	},
	{
		Type: SignalToolRoutingError, ConstName: "SignalToolRoutingError",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Tool Routing Error", DefaultSeverity: models.SeverityHigh,
		ConfidenceMin: 0.7, ConfidenceMax: 0.9,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/tool-routing-error", RuleURI: "docs/rules/ai/tool-routing-error.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierGate,
	},
	{
		Type: SignalToolGuardrailViolation, ConstName: "SignalToolGuardrailViolation",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Tool Guardrail Violation", DefaultSeverity: models.SeverityCritical,
		ConfidenceMin: 0.85, ConfidenceMax: 0.95,
		EvidenceSources: []string{"runtime", "policy"},
		RuleID:          "terrain/ai/tool-guardrail", RuleURI: "docs/rules/ai/tool-guardrail.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierGate,
	},
	{
		Type: SignalToolBudgetExceeded, ConstName: "SignalToolBudgetExceeded",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Tool Budget Exceeded", DefaultSeverity: models.SeverityMedium,
		ConfidenceMin: 0.85, ConfidenceMax: 0.95,
		EvidenceSources: []string{"runtime", "policy"},
		RuleID:          "terrain/ai/tool-budget", RuleURI: "docs/rules/ai/tool-budget.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierObservability,
	},
	{
		Type: SignalAgentFallbackTriggered, ConstName: "SignalAgentFallbackTriggered",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title: "Agent Fallback Triggered", DefaultSeverity: models.SeverityMedium,
		ConfidenceMin: 0.7, ConfidenceMax: 0.9,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/agent-fallback", RuleURI: "docs/rules/ai/agent-fallback.md",
		PromotionPlan: "Off by default; opt in via .terrain/policy.yaml.",
		Tier:          TierObservability,
	},

	// ── AI signals ──────────────────────────────────────────────
	{
		Type: SignalAISafetyEvalMissing, ConstName: "SignalAISafetyEvalMissing",
		Domain: models.CategoryAI, Status: StatusStable,
		Title:           "AI Safety Eval Missing",
		Description:     "Agent or prompt has no eval scenario covering the documented safety category (jailbreak, harm, leak).",
		Remediation:     "Add an eval scenario tagged with the missing safety category and re-run the gauntlet.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   0.75, ConfidenceMax: 0.9,
		EvidenceSources: []string{"structural-pattern", "graph-traversal"},
		RuleID:          "terrain/ai/safety-eval-missing", RuleURI: "docs/rules/ai/safety-eval-missing.md",
		Tier: TierObservability,
	},
	{
		Type: SignalAIPromptVersioning, ConstName: "SignalAIPromptVersioning",
		Domain: models.CategoryAI, Status: StatusStable,
		Title:           "Prompt Versioning",
		Description:     "Prompt-kind surface ships without a recognizable version marker (filename suffix, inline `version:` field, or `# version:` comment). Future content changes will silently drift; consumers can't detect the change.",
		Remediation:     "Add a `version:` field, a `_v<N>` filename suffix, or a `# version: ...` comment so downstream consumers can detect content drift.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.75, ConfidenceMax: 0.92,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/ai/prompt-versioning", RuleURI: "docs/rules/ai/prompt-versioning.md",
		Tier:          TierObservability,
		PromotionPlan: "Observability tier.",
	},
	{
		Type: SignalAIPromptInjectionRisk, ConstName: "SignalAIPromptInjectionRisk",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "Prompt-Injection-Shaped Concatenation",
		Description:     "User-controlled input is concatenated into a prompt without escaping, system-prompt boundaries, or structured input boundaries.",
		Remediation:     "Use a prompt template with explicit user-content boundaries, or run user input through a sanitizer.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   0.6, ConfidenceMax: 0.85,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/ai/prompt-injection-risk", RuleURI: "docs/rules/ai/prompt-injection-risk.md",
		Tier:              TierObservability,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
		DisabledByDefault: true,
	},
	{
		Type: SignalAIHardcodedAPIKey, ConstName: "SignalAIHardcodedAPIKey",
		Domain: models.CategoryAI, Status: StatusStable,
		Title:           "Hard-Coded API Key in AI Configuration",
		Description:     "API-key-shaped string appears in an eval YAML, prompt config, or agent definition.",
		Remediation:     "Move the secret to an environment variable or secrets store and reference it through the runner's secret-resolution path.",
		DefaultSeverity: models.SeverityCritical,
		ConfidenceMin:   0.85, ConfidenceMax: 0.95,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/ai/hardcoded-api-key", RuleURI: "docs/rules/ai/hardcoded-api-key.md",
		Tier:              TierObservability,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
		DisabledByDefault: true,
	},
	{
		Type: SignalAIHardcodedAPIKeyLiteralShape, ConstName: "SignalAIHardcodedAPIKeyLiteralShape",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title:           "Hard-Coded API Key in Source",
		Description:     "An API-key-shaped string appears verbatim in an eval, prompt, or agent definition file. Pairs with the CI-coverage counterpart, secretScannerCoverageDegraded.",
		Remediation:     "Move the secret to an environment variable or secrets store and reference it via the runner's secret-resolution path.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   0.75, ConfidenceMax: 0.9,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/ai/hardcoded-api-key-literal-shape",
		RuleURI:         "docs/rules/ai/hardcoded-api-key-literal-shape.md",
		Tier:            TierObservability,
		PromotionPlan:   "Off by default; opt in via .terrain/policy.yaml.",
	},
	{
		Type: SignalSecretScannerCoverageDegraded, ConstName: "SignalSecretScannerCoverageDegraded",
		Domain: models.CategoryAI, Status: StatusPlanned,
		Title:           "No Secret Scanner in CI",
		Description:     "The repository configures or references AI surfaces that should be guarded by a secret scanner, but no secret-scanner CI integration (GitGuardian, GitHub secret scanning, gitleaks, trufflehog) is enabled. CI-coverage counterpart to aiHardcodedAPIKey-literal-shape.",
		Remediation:     "Enable a secret scanner in CI and document its coverage in the project README. Re-audit periodically.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.6, ConfidenceMax: 0.85,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/ai/secret-scanner-coverage-degraded",
		RuleURI:         "docs/rules/ai/secret-scanner-coverage-degraded.md",
		PromotionPlan:   "Off by default; opt in via .terrain/policy.yaml.",
		Tier:            TierObservability,
	},
	{
		Type: SignalAIToolWithoutSandbox, ConstName: "SignalAIToolWithoutSandbox",
		Domain: models.CategoryAI, Status: StatusStable,
		Title:           "Destructive Tool Without Sandbox",
		Description:     "An agent tool definition can perform an irreversible operation (delete, drop, exec) without an explicit approval gate, sandbox, or dry-run mode.",
		Remediation:     "Wrap the tool in an approval gate or restrict its capability surface to a sandbox.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   0.7, ConfidenceMax: 0.9,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/ai/tool-without-sandbox", RuleURI: "docs/rules/ai/tool-without-sandbox.md",
		Tier: TierGate,
	},
	{
		Type: SignalAINonDeterministicEval, ConstName: "SignalAINonDeterministicEval",
		Domain: models.CategoryAI, Status: StatusStable,
		Title:           "Non-Deterministic Eval Configuration",
		Description:     "An LLM eval runs without temperature pinned to 0 or a deterministic seed, so re-runs produce noisy comparisons.",
		Remediation:     "Pin temperature: 0 and a seed in the eval config, or document the non-determinism budget.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.9, ConfidenceMax: 0.98,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/ai/non-deterministic-eval", RuleURI: "docs/rules/ai/non-deterministic-eval.md",
		Tier: TierObservability,
	},
	{
		Type: SignalAIModelDeprecationRisk, ConstName: "SignalAIModelDeprecationRisk",
		Domain: models.CategoryAI, Status: StatusStable,
		Title:           "Model Pinned to Deprecated or Floating Tag",
		Description:     "Code references a model name that resolves to a deprecated version or a floating tag (e.g. `gpt-4`, `gpt-3.5-turbo`).",
		Remediation:     "Pin to a dated model variant or upgrade to a supported tier.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.8, ConfidenceMax: 0.95,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/ai/model-deprecation-risk", RuleURI: "docs/rules/ai/model-deprecation-risk.md",
		Tier: TierObservability,
	},
	{
		Type: SignalAICostRegression, ConstName: "SignalAICostRegression",
		Domain: models.CategoryAI, Status: StatusStable,
		Title:           "Prompt Token-Cost Regression",
		Description:     "A prompt change increases the token count by more than 25% versus the recorded baseline.",
		Remediation:     "Investigate the change for unintended bloat; bump the baseline if the increase is intentional.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.85, ConfidenceMax: 0.95,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/cost-regression", RuleURI: "docs/rules/ai/cost-regression.md",
		Tier: TierObservability,
	},
	{
		Type: SignalAIHallucinationRate, ConstName: "SignalAIHallucinationRate",
		Domain: models.CategoryAI, Status: StatusStable,
		// The detector does NOT judge hallucinations directly — it reads
		// hallucination-shaped failure metadata that the eval framework
		// (Promptfoo / DeepEval / Ragas) already produced and computes the
		// rate. The signal-type constant keeps its current name for
		// back-compat while the description and remediation carry the
		// accurate framing.
		Title:           "Eval-Flagged Hallucination Share",
		Description:     "The eval framework's own hallucination metadata reports a share of cases above the project-configured threshold (default 5%). Terrain reads this from the framework output (Promptfoo / DeepEval / Ragas) — Terrain does not judge hallucinations directly.",
		Remediation:     "Investigate the underlying eval-flagged cases; tighten retrieval or grounding before merging. If you disagree with the eval framework's classification, fix the eval scenario or raise the threshold (with a documented justification).",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   0.8, ConfidenceMax: 0.95,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/hallucination-rate", RuleURI: "docs/rules/ai/hallucination-rate.md",
		Tier: TierGate,
	},
	{
		Type: SignalAIFewShotContamination, ConstName: "SignalAIFewShotContamination",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "Few-Shot Contamination",
		Description:     "Few-shot examples in a prompt overlap verbatim with the inputs of eval scenarios that exercise that prompt, inflating reported scores.",
		Remediation:     "Hold out the contaminated examples from the prompt's few-shot block, or rewrite the eval input so it isn't a copy of an example. Re-run the eval after de-duplication.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.55, ConfidenceMax: 0.83,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/ai/few-shot-contamination", RuleURI: "docs/rules/ai/few-shot-contamination.md",
		PromotionPlan: "Observability tier.",
		Tier:          TierObservability,
	},
	{
		Type: SignalAIEmbeddingModelChange, ConstName: "SignalAIEmbeddingModelChange",
		Domain: models.CategoryAI, Status: StatusStable,
		Title:           "Embedding Model Swap Without Re-Evaluation",
		Description:     "A repository references an embedding model in source code without a retrieval-shaped eval scenario, so a future model swap will silently change retrieval quality.",
		Remediation:     "Add a retrieval eval scenario (Ragas, Promptfoo, or DeepEval) that exercises this surface so embedding swaps surface as a measurable regression.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.7, ConfidenceMax: 0.88,
		EvidenceSources: []string{"structural-pattern"},
		RuleID:          "terrain/ai/embedding-model-change", RuleURI: "docs/rules/ai/embedding-model-change.md",
		PromotionPlan: "Observability tier.",
		Tier:          TierObservability,
	},
	{
		Type: SignalAIRetrievalRegression, ConstName: "SignalAIRetrievalRegression",
		Domain: models.CategoryAI, Status: StatusStable,
		Title:           "Retrieval Quality Regression",
		Description:     "Context relevance, nDCG, or coverage dropped versus the recorded baseline.",
		Remediation:     "Investigate the regression; revert the offending change or re-tune retrieval before merging.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   0.85, ConfidenceMax: 0.95,
		EvidenceSources: []string{"runtime"},
		RuleID:          "terrain/ai/retrieval-regression", RuleURI: "docs/rules/ai/retrieval-regression.md",
		Tier: TierGate,
	},
	{
		Type: SignalAIPromptSchemaDrift, ConstName: "SignalAIPromptSchemaDrift",
		Domain: models.CategoryAI, Status: StatusStable,
		Title:           "Prompt Template References Changed Schema Field",
		Description:     "A prompt template references a schema field that this PR removed or whose declared type changed. The template will render with a missing value (or wrong type) once merged.",
		Remediation:     "Update the template to use the new schema field, restore the old field, or remove the variable reference.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   0.85, ConfidenceMax: 0.95,
		EvidenceSources: []string{"static"},
		RuleID:          "terrain/ai/prompt-schema-drift", RuleURI: "docs/rules/ai/prompt-schema-drift.md",
		Tier:          TierGate,
		PromotionPlan: "Stable.",
	},

	// ── Engine self-diagnostic signals ──────────────────────────────
	// Emitted by the pipeline itself (safeDetect's panic-recovery path)
	// rather than by a registered detector. Appears in the snapshot so
	// the user sees that something internal failed instead of a
	// silently half-empty result.
	{
		Type: SignalDetectorPanic, ConstName: "SignalDetectorPanic",
		Domain: models.CategoryQuality, Status: StatusStable,
		Title:           "Detector Panic",
		Description:     "A registered detector panicked during the run; safeDetect caught the panic and emitted this marker so the rest of the pipeline could continue.",
		Remediation:     "Re-run with --log-level=debug to capture the stack trace, then file an issue at https://github.com/pmclSF/terrain/issues with the detector ID and the input that triggered the panic.",
		DefaultSeverity: models.SeverityCritical,
		ConfidenceMin:   1.0, ConfidenceMax: 1.0,
		EvidenceSources: []string{"static"},
		RuleID:          "terrain/engine/detector-panic", RuleURI: "docs/rules/engine/detector-panic.md",
		Tier: TierObservability,
	},
	// Per-detector wall-clock timeout budgets. Emitted by the
	// pipeline (safeDetectWithBudget) when a detector exceeds its
	// DetectorMeta.Budget (default DefaultDetectorBudget). The
	// detector's signals from any post-budget completion are dropped
	// — this marker is the only signal returned for the abandoned
	// detector.
	{
		Type: SignalDetectorBudgetExceeded, ConstName: "SignalDetectorBudgetExceeded",
		Domain: models.CategoryQuality, Status: StatusStable,
		Title:           "Detector Budget Exceeded",
		Description:     "A registered detector exceeded its wall-clock budget and was abandoned by the pipeline. The rest of the pipeline continued without that detector's signals.",
		Remediation:     "If the detector is legitimately slow on your repo, raise DetectorMeta.Budget for it. If it should be fast, the runaway suggests a quadratic-or-worse code path or a hung I/O — re-run with --log-level=debug.",
		DefaultSeverity: models.SeverityCritical,
		ConfidenceMin:   1.0, ConfidenceMax: 1.0,
		EvidenceSources: []string{"static"},
		RuleID:          "terrain/engine/detector-budget", RuleURI: "docs/rules/engine/detector-budget.md",
		Tier: TierObservability,
	},
	// Emitted by safeDetectChecked when a detector's declared input
	// requirements (RequiresRuntime / RequiresBaseline /
	// RequiresEvalArtifact) aren't satisfied by the current snapshot.
	// Surfaces the gap so adopters know which flag to add rather than
	// seeing silent zero-output from the affected detector.
	{
		Type: SignalDetectorMissingInput, ConstName: "SignalDetectorMissingInput",
		Domain: models.CategoryQuality, Status: StatusStable,
		Title:           "Detector Missing Input",
		Description:     "A registered detector requires inputs (runtime artifacts, baseline snapshot, or eval-framework results) that the current snapshot doesn't carry. The detector was skipped; the rest of the pipeline ran normally.",
		Remediation:     "The marker explanation lists the specific flag(s) to pass to `terrain analyze` to provide the missing inputs. If you don't need this detector's signals, leave the inputs absent — the marker is informational.",
		DefaultSeverity: models.SeverityLow,
		ConfidenceMin:   1.0, ConfidenceMax: 1.0,
		EvidenceSources: []string{"static"},
		RuleID:          "terrain/engine/detector-missing-input", RuleURI: "docs/rules/engine/detector-missing-input.md",
		Tier: TierObservability,
	},
	{
		Type: SignalSuppressionExpired, ConstName: "SignalSuppressionExpired",
		Domain: models.CategoryGovernance, Status: StatusStable,
		Title:           "Suppression Expired",
		Description:     "A `.terrain/suppressions.yaml` entry has passed its `expires` date and is no longer in effect. The underlying findings will fire again until the entry is renewed or removed.",
		Remediation:     "Edit `.terrain/suppressions.yaml`: extend the `expires` date if the suppression is still warranted, or remove the entry if the underlying issue is resolved.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   1.0, ConfidenceMax: 1.0,
		EvidenceSources: []string{"policy"},
		RuleID:          "terrain/engine/suppression-expired", RuleURI: "docs/rules/engine/suppression-expired.md",
		Tier: TierObservability,
	},

	// ── Stable rules ─────────────────────────────────────────────
	// These implement the canonical stable taxonomy (regression /
	// coverage / hygiene / reproducibility / security / performance /
	// data). Domain stays as the closest existing SignalCategory until
	// the SignalCategory enum is extended in a separate change; the
	// rule ID encodes the taxonomy category.

	{
		Type: SignalVersionFloating, ConstName: "SignalVersionFloating",
		Domain: models.CategoryQuality, Status: StatusExperimental,
		Title:           "Floating Dependency Version",
		Description:     "A dependency is declared without a version pin (unpinned, range-only, or moving git/url reference). Subsequent installs may resolve to different versions, introducing non-determinism in test and eval runs.",
		Remediation:     "Pin the dependency to an exact version, commit a lockfile that records the resolved set, or use a content-addressed git SHA reference.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.85, ConfidenceMax: 0.99,
		EvidenceSources:   []string{"structural-pattern", "manifest"},
		RuleID:            "terrain/reproducibility/version-floating",
		RuleURI:           "docs/rules/reproducibility/version-floating.md",
		Tier:              TierObservability,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},

	// Remaining stable rules — declared as Planned. Each gets its
	// detector + doc page in subsequent commits as the implementations
	// land. The manifest-parity test requires every Signal* constant
	// to have an entry here.

	{
		Type: SignalSecretsInPrompt, ConstName: "SignalSecretsInPrompt",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "Secrets in Prompt",
		Description:     "A prompt-classified file contains embedded credentials (OpenAI / Anthropic / GitHub / Slack / AWS keys, JWT, bearer tokens). Anyone with read access to the prompt has access to the credential.",
		Remediation:     "Rotate the leaked credential immediately, then move it to an environment variable or secret manager.",
		DefaultSeverity: models.SeverityCritical,
		ConfidenceMin:   0.95, ConfidenceMax: 0.99,
		EvidenceSources:   []string{"structural-pattern"},
		RuleID:            "terrain/hygiene/secrets-in-prompt",
		RuleURI:           "docs/rules/hygiene/secrets-in-prompt.md",
		Tier:              TierGate,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},
	{
		Type: SignalNoTestsForCodeUnit, ConstName: "SignalNoTestsForCodeUnit",
		Domain: models.CategoryQuality, Status: StatusExperimental,
		Title:           "No Tests for Code Unit",
		Description:     "A code unit (exported function / method / class) exists in the codebase but no test in the snapshot's dependency graph covers it. Untested code reaches production undetected when changed.",
		Remediation:     "Add a test that imports the code unit and exercises its observable behavior. The rule defaults to exported symbols only; configure `include_private: true` to widen coverage.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.85, ConfidenceMax: 0.95,
		EvidenceSources:   []string{"graph-traversal"},
		RuleID:            "terrain/coverage/no-tests",
		RuleURI:           "docs/rules/coverage/no-tests.md",
		Tier:              TierObservability,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},
	{
		Type: SignalNoEvalForAISurface, ConstName: "SignalNoEvalForAISurface",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "No Eval for AI Surface",
		Description:     "An AI-typed CodeSurface (prompt / context / dataset / tool / retrieval / agent / eval_definition / model) has no Eval that claims to cover it. Model behavior can shift in production without any eval surfacing the regression.",
		Remediation:     "Add an eval scenario that exercises the surface and asserts on its output / metric / shape.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   0.8, ConfidenceMax: 0.95,
		EvidenceSources:   []string{"graph-traversal"},
		RuleID:            "terrain/coverage/no-eval",
		RuleURI:           "docs/rules/coverage/no-eval.md",
		Tier:              TierGate,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},
	{
		Type: SignalModelFixtureUnpinned, ConstName: "SignalModelFixtureUnpinned",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "Model Fixture Unpinned",
		Description:     "A model-loading call (from_pretrained / torch.load / joblib.load / load_model / snapshot_download) uses a path or revision that isn't content-addressed. The underlying weights may change without a code edit, regressing eval scores silently.",
		Remediation:     "Pin the load to a commit SHA (revision=\"<sha>\" for HuggingFace), a version-suffixed filename (model_v3.0.pt), or a .safetensors-with-checksum format.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   0.85, ConfidenceMax: 0.95,
		EvidenceSources:   []string{"structural-pattern"},
		RuleID:            "terrain/hygiene/model-fixture-unpinned",
		RuleURI:           "docs/rules/hygiene/model-fixture-unpinned.md",
		Tier:              TierGate,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},
	{
		Type: SignalEvalNoAssertion, ConstName: "SignalEvalNoAssertion",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "Eval Without Assertion",
		Description:     "An eval test function runs to completion without any assertion / score / metric call. The test cannot detect regressions because it accepts any model output.",
		Remediation:     "Add an assert / score check that fails when the eval output deviates from expectations.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   0.8, ConfidenceMax: 0.95,
		EvidenceSources:   []string{"structural-pattern"},
		RuleID:            "terrain/hygiene/eval-no-assertion",
		RuleURI:           "docs/rules/hygiene/eval-no-assertion.md",
		Tier:              TierGate,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},
	{
		Type: SignalNoSeed, ConstName: "SignalNoSeed",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "Missing Random Seed",
		Description:     "Stochastic library call (np.random / torch / random / tf.random) in an eval or training file without a preceding seed call. Run-to-run results vary, masking real regressions.",
		Remediation:     "Add a seed call at module scope or in a pytest fixture (np.random.seed(42), torch.manual_seed(42), or transformers.set_seed(42)).",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.75, ConfidenceMax: 0.9,
		EvidenceSources:   []string{"structural-pattern"},
		RuleID:            "terrain/reproducibility/no-seed",
		RuleURI:           "docs/rules/reproducibility/no-seed.md",
		Tier:              TierObservability,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},
	{
		Type: SignalMissingEnvPinning, ConstName: "SignalMissingEnvPinning",
		Domain: models.CategoryQuality, Status: StatusExperimental,
		Title:           "Missing Env Pinning",
		Description:     "An environment-variable read in eval / inference code lacks a default value. The same code produces different behavior depending on which environment runs it.",
		Remediation:     "Supply a default — os.environ.get(KEY, \"<pinned-value>\") — or fail fast with a clear error message when the variable is absent.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.7, ConfidenceMax: 0.9,
		EvidenceSources:   []string{"structural-pattern"},
		RuleID:            "terrain/reproducibility/missing-env-pinning",
		RuleURI:           "docs/rules/reproducibility/missing-env-pinning.md",
		Tier:              TierObservability,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},
	{
		Type: SignalPIIInEval, ConstName: "SignalPIIInEval",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "PII in Eval Dataset",
		Description:     "An eval-directory file contains PII-shaped values (emails, phone numbers, SSNs, credit card numbers, IPv4 addresses). Eval datasets that retain production PII expose customer data to anyone with repo access.",
		Remediation:     "Replace PII in the eval dataset with synthetic equivalents (Faker, Mimesis, mockaroo) or apply a redaction pass before committing.",
		DefaultSeverity: models.SeverityCritical,
		ConfidenceMin:   0.75, ConfidenceMax: 0.95,
		EvidenceSources:   []string{"structural-pattern"},
		RuleID:            "terrain/security/pii-in-eval",
		RuleURI:           "docs/rules/security/pii-in-eval.md",
		Tier:              TierGate,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},
	{
		Type: SignalInsecureDeserialize, ConstName: "SignalInsecureDeserialize",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "Insecure Deserialization",
		Description:     "A call into an unsafe deserialization primitive (pickle.load, torch.load without weights_only=True, joblib.load, yaml.load without SafeLoader, dill.load, marshal.load) executes arbitrary code on untrusted input.",
		Remediation:     "Switch to a safe format (JSON, msgpack, safetensors, ONNX). When the primitive is unavoidable, declare the explicit safe option (weights_only=True for torch.load, Loader=SafeLoader for yaml.load).",
		DefaultSeverity: models.SeverityCritical,
		ConfidenceMin:   0.9, ConfidenceMax: 0.99,
		EvidenceSources:   []string{"structural-pattern"},
		RuleID:            "terrain/security/insecure-deserialization",
		RuleURI:           "docs/rules/security/insecure-deserialization.md",
		Tier:              TierGate,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},
	{
		Type: SignalMissingPerfTest, ConstName: "SignalMissingPerfTest",
		Domain: models.CategoryQuality, Status: StatusExperimental,
		Title:           "Missing Performance Test",
		Description:     "A latency-critical AI surface (prompt / retrieval / agent / model / handler / route) has no benchmark or load test exercising it. Latency or throughput regressions ship silently.",
		Remediation:     "Add a benchmark under benchmarks/ or perf/ that records P50 / P95 latency for the surface.",
		DefaultSeverity: models.SeverityLow,
		ConfidenceMin:   0.7, ConfidenceMax: 0.85,
		EvidenceSources:   []string{"graph-traversal"},
		RuleID:            "terrain/performance/missing-perf-test",
		RuleURI:           "docs/rules/performance/missing-perf-test.md",
		Tier:              TierObservability,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},
	{
		Type: SignalDataLeakageSuspected, ConstName: "SignalDataLeakageSuspected",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "Data Leakage Suspected",
		Description:     "Source-level patterns associated with train/test contamination: preprocessing (scaler/encoder fit) applied before the split, or random train/test split applied to time-series data.",
		Remediation:     "Move scaler/encoder fits to AFTER the split. For time-series, use TimeSeriesSplit or a manual time-based cutoff.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   0.7, ConfidenceMax: 0.9,
		EvidenceSources:   []string{"structural-pattern"},
		RuleID:            "terrain/data/leakage-suspected",
		RuleURI:           "docs/rules/data/leakage-suspected.md",
		Tier:              TierGate,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},
	{
		Type: SignalMissingTrainTestSplit, ConstName: "SignalMissingTrainTestSplit",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "Missing Train/Test Split",
		Description:     "A training call (.fit() / .train() / .partial_fit()) appears in a training file without any preceding split helper (train_test_split, KFold, TimeSeriesSplit, cross_val_score, etc.). The model is fit on the full dataset; evaluation against the same data measures memorization, not generalization.",
		Remediation:     "Split the dataset before training (sklearn.model_selection.train_test_split, KFold for general use, TimeSeriesSplit for temporal data).",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   0.75, ConfidenceMax: 0.9,
		EvidenceSources:   []string{"structural-pattern"},
		RuleID:            "terrain/data/missing-train-test-split",
		RuleURI:           "docs/rules/data/missing-train-test-split.md",
		Tier:              TierGate,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},

	// ── Regression family (stable) ────────────────────────────────
	// Consume the Tier 1 eval-adapter foundation to compare baseline
	// and current runs.
	{
		Type: SignalBaselineNotSet, ConstName: "SignalBaselineNotSet",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "Baseline Not Set",
		Description:     "An EvalRun exists for the current PR but no baseline is recorded. Eval-regression detection is disabled until a baseline exists.",
		Remediation:     "Run `terrain ai record` on the current main-branch state to lock the baseline. Subsequent PRs will be compared against it.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.99, ConfidenceMax: 0.99,
		EvidenceSources:   []string{"eval-execution"},
		RuleID:            "terrain/regression/baseline-not-set",
		RuleURI:           "docs/rules/regression/baseline-not-set.md",
		Tier:              TierObservability,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},
	{
		Type: SignalPassRateDrop, ConstName: "SignalPassRateDrop",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "Eval Pass Rate Dropped",
		Description:     "The success / total ratio across eval cases dropped past the configured threshold from baseline to current. Distinct from eval-regression (continuous score deltas) — fires on discrete pass/fail count deltas.",
		Remediation:     "Inspect per-case eval-regression findings for cases that flipped from pass to fail. If intentional, update the baseline.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   0.9, ConfidenceMax: 0.99,
		EvidenceSources:   []string{"eval-execution"},
		RuleID:            "terrain/regression/pass-rate-drop",
		RuleURI:           "docs/rules/regression/pass-rate-drop.md",
		Tier:              TierGate,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},
	{
		Type: SignalSnapshotMismatch, ConstName: "SignalSnapshotMismatch",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "Snapshot Mismatch",
		Description:     "An eval case's recorded output snapshot diverged from baseline to current. Catches behavior changes the scalar score may not surface.",
		Remediation:     "Inspect the diff for prompt / model / retrieval changes affecting the case. If the new output is correct, accept it via `terrain ai record`.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.85, ConfidenceMax: 0.95,
		EvidenceSources:   []string{"eval-execution"},
		RuleID:            "terrain/regression/snapshot-mismatch",
		RuleURI:           "docs/rules/regression/snapshot-mismatch.md",
		Tier:              TierObservability,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},
	{
		Type: SignalTestFailed, ConstName: "SignalTestFailed",
		Domain: models.CategoryHealth, Status: StatusExperimental,
		Title:           "Impacted Test Failed",
		Description:     "A test selected by impact analysis as relevant to the current change failed. The change broke something the test suite already protects.",
		Remediation:     "Reproduce locally with `terrain test --selector regression/test-failed`. Fix the failure or, if the test is stale, update it deliberately.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   1.0, ConfidenceMax: 1.0,
		EvidenceSources:   []string{"runtime", "graph-traversal"},
		RuleID:            "terrain/regression/test-failed",
		RuleURI:           "docs/rules/regression/test-failed.md",
		Tier:              TierGate,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},
	{
		Type: SignalPerformanceRegression, ConstName: "SignalPerformanceRegression",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "Performance Regression",
		Description:     "An ML model's performance metric (accuracy / F1 / AUC / RMSE / etc.) regressed past the configured threshold from baseline to current. Same shape as eval-regression but applied to classical ML metrics rather than LLM rubric scores.",
		Remediation:     "Inspect the diff for training data / hyperparameter / feature changes. If the regression is intentional, update the baseline.",
		DefaultSeverity: models.SeverityHigh,
		ConfidenceMin:   0.9, ConfidenceMax: 0.99,
		EvidenceSources:   []string{"eval-execution"},
		RuleID:            "terrain/regression/performance-regression",
		RuleURI:           "docs/rules/regression/performance-regression.md",
		Tier:              TierGate,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},

	// ── Coverage family (stable) ──────────────────────────────────
	{
		Type: SignalMissingBaseline, ConstName: "SignalMissingBaseline",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title:           "Missing Coverage Baseline",
		Description:     "The repository has eval surfaces but no `.terrain/baselines/` directory exists. Eval regression detection is disabled at the coverage layer.",
		Remediation:     "Run `terrain ai record` to create the baseline directory. Commit `.terrain/baselines/latest.json` so subsequent PRs compare against it.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   1.0, ConfidenceMax: 1.0,
		EvidenceSources:   []string{"static"},
		RuleID:            "terrain/coverage/missing-baseline",
		RuleURI:           "docs/rules/coverage/missing-baseline.md",
		Tier:              TierObservability,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},
	{
		Type: SignalNoIntegrationTest, ConstName: "SignalNoIntegrationTest",
		Domain: models.CategoryQuality, Status: StatusExperimental,
		Title:           "No Integration Test",
		Description:     "A code unit reachable from a production entry point (handler / route) has no integration test exercising it through that entry point.",
		Remediation:     "Add an integration test that exercises the handler / route end-to-end. The unit test stays as a fast inner-loop check; the integration test ensures the cross-stack contract holds.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.8, ConfidenceMax: 0.95,
		EvidenceSources:   []string{"graph-traversal"},
		RuleID:            "terrain/coverage/no-integration-test",
		RuleURI:           "docs/rules/coverage/no-integration-test.md",
		Tier:              TierObservability,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},
	{
		Type: SignalNoDataValidation, ConstName: "SignalNoDataValidation",
		Domain: models.CategoryQuality, Status: StatusExperimental,
		Title:           "No Data Validation",
		Description:     "A data pipeline file (ETL / dbt model / training data loader) has no data-validation library import (Great Expectations, pandera, dbt-expectations, soda).",
		Remediation:     "Add data validation (GE expectations, pandera schemas, dbt-expectations tests) on the pipeline's output. Run the validation in CI on a fixed sample.",
		DefaultSeverity: models.SeverityMedium,
		ConfidenceMin:   0.75, ConfidenceMax: 0.9,
		EvidenceSources:   []string{"structural-pattern"},
		RuleID:            "terrain/coverage/no-data-validation",
		RuleURI:           "docs/rules/coverage/no-data-validation.md",
		Tier:              TierObservability,
		DisabledByDefault: true,
		PromotionPlan:     "Off by default; opt in via .terrain/policy.yaml.",
	},

	// ── Preview rules ────────────────────────────────────────────
	// Preview rules ship detection logic with a short-form doc page.
	// They're default-off; enable via .terrain/policy.yaml.
	//
	// Status=Experimental signals "detection works but not yet broadly
	// validated"; detectors land alongside these entries in
	// internal/preview/.

	{
		Type: SignalPromptBloat, ConstName: "SignalPromptBloat",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title: "Prompt Bloat", DefaultSeverity: models.SeverityLow,
		ConfidenceMin: 0.7, ConfidenceMax: 0.85, EvidenceSources: []string{"structural-pattern"},
		RuleID: "terrain/prompt-quality/prompt-bloat", RuleURI: "docs/rules/prompt-quality/prompt-bloat.md",
		PromotionPlan: "Fires when prompt token count exceeds the configured budget.",
		Tier:          TierObservability,
	},
	{
		Type: SignalPromptWithoutTemperature, ConstName: "SignalPromptWithoutTemperature",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title: "Prompt Without Temperature", DefaultSeverity: models.SeverityLow,
		ConfidenceMin: 0.85, ConfidenceMax: 0.95, EvidenceSources: []string{"structural-pattern"},
		RuleID: "terrain/prompt-quality/prompt-without-temperature", RuleURI: "docs/rules/prompt-quality/prompt-without-temperature.md",
		PromotionPlan: "Fires when an LLM SDK call has no temperature kwarg. Defaults differ across SDKs; an explicit value is reproducibility-critical.",
		Tier:          TierObservability,
	},
	{
		Type: SignalMissingPromptValidator, ConstName: "SignalMissingPromptValidator",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title: "Missing Prompt Validator", DefaultSeverity: models.SeverityMedium,
		ConfidenceMin: 0.7, ConfidenceMax: 0.85, EvidenceSources: []string{"structural-pattern"},
		RuleID: "terrain/prompt-quality/missing-validator", RuleURI: "docs/rules/prompt-quality/missing-validator.md",
		PromotionPlan: "Fires when a prompt template has no output-validator schema attached.",
		Tier:          TierObservability,
	},
	{
		Type: SignalPromptVersionSkew, ConstName: "SignalPromptVersionSkew",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title: "Prompt Version Skew", DefaultSeverity: models.SeverityMedium,
		ConfidenceMin: 0.75, ConfidenceMax: 0.9, EvidenceSources: []string{"graph-traversal"},
		RuleID: "terrain/prompt-quality/version-skew", RuleURI: "docs/rules/prompt-quality/version-skew.md",
		PromotionPlan: "Detects when the same prompt template is referenced by multiple eval scenarios under different version names.",
		Tier:          TierObservability,
	},
	{
		Type: SignalRetrievalWithoutRerank, ConstName: "SignalRetrievalWithoutRerank",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title: "Retrieval Without Rerank", DefaultSeverity: models.SeverityLow,
		ConfidenceMin: 0.65, ConfidenceMax: 0.8, EvidenceSources: []string{"structural-pattern"},
		RuleID: "terrain/retrieval-quality/no-rerank", RuleURI: "docs/rules/retrieval-quality/no-rerank.md",
		PromotionPlan: "Flags retrieval pipelines with top_k > 5 and no reranker.",
		Tier:          TierObservability,
	},
	{
		Type: SignalColdVectorStore, ConstName: "SignalColdVectorStore",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title: "Cold Vector Store", DefaultSeverity: models.SeverityLow,
		ConfidenceMin: 0.7, ConfidenceMax: 0.85, EvidenceSources: []string{"structural-pattern"},
		RuleID: "terrain/retrieval-quality/cold-store", RuleURI: "docs/rules/retrieval-quality/cold-store.md",
		PromotionPlan: "Fires when a vector store is initialized but no index-population call exists in the same module.",
		Tier:          TierObservability,
	},
	{
		Type: SignalAgentLoopRisk, ConstName: "SignalAgentLoopRisk",
		Domain: models.CategoryAI, Status: StatusStable,
		Title: "Agent Loop Risk", DefaultSeverity: models.SeverityHigh,
		ConfidenceMin: 0.7, ConfidenceMax: 0.85, EvidenceSources: []string{"structural-pattern"},
		RuleID: "terrain/agent-quality/loop-risk", RuleURI: "docs/rules/agent-quality/loop-risk.md",
		PromotionPlan: "Stable.",
		Tier:          TierGate,
	},
	{
		Type: SignalToolWithoutBudget, ConstName: "SignalToolWithoutBudget",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title: "Tool Without Budget", DefaultSeverity: models.SeverityMedium,
		ConfidenceMin: 0.75, ConfidenceMax: 0.9, EvidenceSources: []string{"structural-pattern"},
		RuleID: "terrain/agent-quality/tool-no-budget", RuleURI: "docs/rules/agent-quality/tool-no-budget.md",
		PromotionPlan: "Fires when a tool-call-enabled agent has no rate limit or cost ceiling configured.",
		Tier:          TierObservability,
	},
	{
		Type: SignalTargetLeakage, ConstName: "SignalTargetLeakage",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title: "Target Leakage", DefaultSeverity: models.SeverityHigh,
		ConfidenceMin: 0.7, ConfidenceMax: 0.85, EvidenceSources: []string{"structural-pattern"},
		RuleID: "terrain/data-quality/target-leakage", RuleURI: "docs/rules/data-quality/target-leakage.md",
		PromotionPlan: "Fires when a feature column is derived from the target column (e.g., y_lag1 in features after target encoding).",
		Tier:          TierGate,
	},
	{
		Type: SignalDuplicateEvalRows, ConstName: "SignalDuplicateEvalRows",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title: "Duplicate Eval Rows", DefaultSeverity: models.SeverityMedium,
		ConfidenceMin: 0.9, ConfidenceMax: 0.99, EvidenceSources: []string{"structural-pattern"},
		RuleID: "terrain/data-quality/duplicate-rows", RuleURI: "docs/rules/data-quality/duplicate-rows.md",
		PromotionPlan: "Fires when an eval dataset has more than 5% duplicate input rows.",
		Tier:          TierObservability,
	},
	{
		Type: SignalSchemaDrift, ConstName: "SignalSchemaDrift",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title: "Schema Drift", DefaultSeverity: models.SeverityHigh,
		ConfidenceMin: 0.85, ConfidenceMax: 0.95, EvidenceSources: []string{"structural-pattern"},
		RuleID: "terrain/data-quality/schema-drift", RuleURI: "docs/rules/data-quality/schema-drift.md",
		PromotionPlan: "Fires when the pipeline output schema has changed between baseline and current run.",
		Tier:          TierGate,
	},
	{
		Type: SignalMissingEvalCategories, ConstName: "SignalMissingEvalCategories",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title: "Missing Eval Categories", DefaultSeverity: models.SeverityLow,
		ConfidenceMin: 0.65, ConfidenceMax: 0.8, EvidenceSources: []string{"structural-pattern"},
		RuleID: "terrain/coverage/missing-eval-categories", RuleURI: "docs/rules/coverage/missing-eval-categories.md",
		PromotionPlan: "Fires when an eval suite has happy-path coverage but no adversarial or edge-case categories.",
		Tier:          TierObservability,
	},
	{
		Type: SignalOrphanedEval, ConstName: "SignalOrphanedEval",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title: "Orphaned Eval", DefaultSeverity: models.SeverityLow,
		ConfidenceMin: 0.8, ConfidenceMax: 0.95, EvidenceSources: []string{"graph-traversal"},
		RuleID: "terrain/coverage/orphaned-eval", RuleURI: "docs/rules/coverage/orphaned-eval.md",
		PromotionPlan: "Fires when an eval has no CoveredSurfaceIDs (references no surface).",
		Tier:          TierObservability,
	},
	{
		Type: SignalColdStartTime, ConstName: "SignalColdStartTime",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title: "Cold Start Time", DefaultSeverity: models.SeverityLow,
		ConfidenceMin: 0.8, ConfidenceMax: 0.95, EvidenceSources: []string{"runtime"},
		RuleID: "terrain/performance/cold-start-time", RuleURI: "docs/rules/performance/cold-start-time.md",
		PromotionPlan: "Fires when first-request latency exceeds the configured threshold (e.g., 2x P50).",
		Tier:          TierObservability,
	},
	{
		Type: SignalTokenCostBudget, ConstName: "SignalTokenCostBudget",
		Domain: models.CategoryAI, Status: StatusExperimental,
		Title: "Token Cost Budget Exceeded", DefaultSeverity: models.SeverityMedium,
		ConfidenceMin: 0.9, ConfidenceMax: 0.99, EvidenceSources: []string{"runtime"},
		RuleID: "terrain/performance/token-cost-budget", RuleURI: "docs/rules/performance/token-cost-budget.md",
		PromotionPlan: "Fires when per-run token cost exceeds the configured ceiling.",
		Tier:          TierObservability,
	},
}

// Manifest returns a snapshot copy of the canonical signal manifest, sorted
// alphabetically by signal type. Callers should treat the result as read-only.
func Manifest() []ManifestEntry {
	out := make([]ManifestEntry, len(allSignalManifest))
	copy(out, allSignalManifest)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Type < out[j].Type
	})
	return out
}

// ManifestByType returns the manifest entry for a given signal type, or
// (zero, false) if no entry exists.
func ManifestByType(t models.SignalType) (ManifestEntry, bool) {
	for _, e := range allSignalManifest {
		if e.Type == t {
			return e, true
		}
	}
	return ManifestEntry{}, false
}

// TitleForType returns the manifest Title for a signal type, falling
// back to the raw type string when no entry exists. User-facing
// renderers use this so labels read as human names ("Weak Assertion")
// rather than machine ids ("weakAssertion").
func TitleForType(t models.SignalType) string {
	if e, ok := ManifestByType(t); ok && e.Title != "" {
		return e.Title
	}
	return string(t)
}

// AllSignalTypes returns every signal type currently declared in the manifest.
func AllSignalTypes() []models.SignalType {
	out := make([]models.SignalType, len(allSignalManifest))
	for i, e := range allSignalManifest {
		out[i] = e.Type
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// SignalTypesByStatus returns every signal type with the given status.
func SignalTypesByStatus(status SignalStatus) []models.SignalType {
	var out []models.SignalType
	for _, e := range allSignalManifest {
		if e.Status == status {
			out = append(out, e.Type)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
