# Implementation Workbook

This document defines the stage-by-stage implementation flow for Claude Code.

## Rules for Claude Code

- preserve working behavior where possible
- prefer adapters over destructive rewrites
- keep the repo buildable
- implement one vertical slice fully before the next
- avoid speculative "future architecture" code without purpose
- create useful stubs with comments and TODOs

## Stage sequence

### Stage 1
Docs, repo structure, scaffolding

### Stage 2
Core models and signal registry

### Stage 3
Static analysis nucleus

### Stage 4
Initial detectors

### Stage 5
Risk engine + analyze command

### Stage 6
Snapshot persistence + extension scaffold

### Stage 7
Policy checks

Implemented:
- internal/policy: Config model + YAML loader
- internal/governance: Evaluate() + governance signal construction
- cmd/terrain: `terrain policy check` subcommand with exit codes
- Governance signals: policyViolation, legacyFrameworkUsage, runtimeBudgetExceeded
- Policy rules: disallow_skipped_tests, disallow_frameworks, max_test_runtime_ms,
  minimum_coverage_percent, max_weak_assertions, max_mock_heavy_tests
- `terrain analyze` includes governance signals when policy file is present
- Example policy file at .terrain/policy.yaml
- Tests for policy parsing, governance evaluation, risk engine

### Stage 8
Ownership and review triage

Implemented:
- internal/ownership: Resolver with explicit config, CODEOWNERS, directory fallback
- Ownership attached to TestFile and Signal during analysis pipeline
- internal/reporting/review.go: GroupSignalsByOwner, GroupSignalsByType,
  GroupSignalsByDirectory, MigrationBlockers, RenderReviewSections
- CLI analyze output now includes review sections (highest risk, by owner, by directory, migration blockers)
- Extension scaffold: types.ts, signal_renderer.ts, views.ts with Review and Migration view builders
- Tests for ownership resolution, grouping logic, and review rendering

### Stage 9
Migration intelligence

Implemented:
- internal/migration: 4 detectors (DeprecatedPatternDetector, DynamicTestGenerationDetector,
  CustomMatcherDetector, FrameworkMigrationDetector)
- Blocker taxonomy constants (custom-matcher, dynamic-generation, deprecated-pattern,
  framework-helper, unsupported-setup)
- ReadinessSummary model with readiness levels (high/medium/low/unknown)
- Migration reporting (RenderMigrationReport)
- Detectors integrated into `terrain analyze` pipeline
- Tests for all detectors and readiness computation

### Stage 9.1
Migration preview boundary

Implemented:
- internal/migration: UnsupportedSetupDetector (5th detector for setup/fixture patterns)
- internal/migration: PreviewFile() and PreviewScope() for file-level migration preview
- PreviewResult model: sourceFramework, suggestedTarget, difficulty, blockers, safePatterns, limitations
- PreviewBlocker model: type, pattern, explanation, remediation
- Blocker taxonomy: unsupported-setup detector now active (was defined but unimplemented)
- internal/reporting: RenderMigrationPreview, RenderMigrationPreviewScope
- CLI: `terrain migration preview --file PATH --scope DIR --json`
- Extension: MigrationData now includes byDirectory, areaAssessments, previewAvailable
- Extension: ReviewData now includes migrationBlockers as first-class grouping
- Extension types: MigrationPreviewResult, MigrationReadiness types added
- Risk engine: migration signals (deprecated, dynamic, custom matcher, unsupported setup) now contribute to change risk
- Tests: 12 new tests (3 UnsupportedSetup detector, 7 preview boundary, 5 reporting)
- Example artifacts: migration-readiness.txt, migration-preview.txt, migration-preview.json

### Stage 10
Historical comparison and trend detection

Implemented:
- Timestamped snapshot archiving in persistSnapshot (latest.json + YYYY-MM-DDTHH-MM-SSZ.json)
- internal/comparison: Compare(), SnapshotComparison, SignalDelta, RiskDelta,
  FrameworkChange, SignalExample
- `terrain compare` command with --from/--to/--root/--json flags
- internal/reporting/comparison_report.go: RenderComparisonReport
- Auto-discovery of two most recent timestamped snapshots
- Tests for signal deltas, risk deltas, framework changes, representative examples

### Stage 11
Benchmark-ready metrics

Implemented:
- internal/metrics: Snapshot model with aggregate-only metrics (privacy boundary enforced)
- metrics.Derive() extracts structure, health, quality, change, governance, risk metrics
- `terrain metrics` command with --root/--json flags
- internal/reporting/metrics_report.go: RenderMetricsReport scorecard
- Notes system for missing inputs (no test files, no runtime data)
- Tests for all metric categories, safe ratio, runtime detection

### Stage 12
Risk heatmap and leadership summary

Implemented:
- internal/heatmap: Heatmap model with HotSpot, Build() function
- Directory hotspots derived from risk surfaces
- Owner hotspots derived from signal ownership
- Posture band (highest across all risk surfaces) and summary text
- internal/reporting/summary_report.go: RenderSummaryReport
- `terrain summary` command with --root/--json flags
- Tests for hotspot building, sorting, posture computation, band mapping

### Stage 13
Benchmarking scaffolding

Implemented:
- internal/benchmark: Export model with schema version, segmentation, metrics
- Segment primitives: primaryLanguage, primaryFramework, testFileBucket (small/medium/large),
  frameworkCount, hasCoverage, hasRuntimeData, hasPolicy
- BuildExport() creates benchmark-safe artifact from snapshot + derived metrics
- `terrain export benchmark` command (JSON-only, --root flag)
- Privacy boundary: no raw file paths, symbol names, or source code
- Tests for export building, segmentation buckets, runtime/coverage detection

### Stage 14
Executive summary and paid-product UX scaffolding

Implemented:
- internal/summary: ExecutiveSummary model with PostureSummary, FocusArea,
  TrendCallout, BenchmarkReadinessSummary, KeyNumbers
- summary.Build() synthesizes heatmap + comparison + benchmark into one artifact
- internal/reporting/executive_report.go: RenderExecutiveSummary
- `terrain summary` now integrates trend comparison (auto-loads prior snapshots)
  and benchmark readiness (segmentation + data completeness)
- JSON output for executive summary (future UI contract candidate)
- Graceful degradation when no snapshot history exists
- Example artifacts: executive-summary.txt, executive-summary.json
- Tests for posture, hotspots, trends, benchmark readiness, JSON serialization,
  rendering sections, direction icons, empty/partial states

### Stage 15
Release readiness

Implemented:
- README.md rewritten with quick start, commands table, snapshot workflow, policy,
  architecture, project structure, development instructions, status
- docs/README.md updated as navigation index (start here, product, technical, release, legacy)
- docs/demo.md walkthrough covering all major commands
- docs/release-checklist.md with product, engineering, UX, docs, packaging, honest gaps
- Makefile with build, test, test-v, test-cover, check, clean, demo targets
- CLI help text polished (summary description updated)
- Extension coherence verified against CLI JSON contract

### Stage 16
Final pre-launch hardening

Implemented:
- Removed stale messages ("not yet active", "current analysis nucleus" → "current")
- Improved compare error messages (actionable guidance for missing snapshots)
- CLI doc comment aligned with actual command surface
- Test expectations updated for output changes
- Empty/error states reviewed across all report renderers
- Release checklist updated with hardening section

## Acceptance principle

A stage is only complete if:
- it compiles or is structurally coherent
- it has tests where appropriate
- docs match reality
- next steps are explicit
