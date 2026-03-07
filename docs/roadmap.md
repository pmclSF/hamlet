# Roadmap

## V3 Objective

Refactor Hamlet into a signal-first test intelligence platform while preserving the migration wedge.

---

## Milestone A -- Freeze strategy and repo structure

Deliverables:
- product docs
- architecture docs
- signal model
- signal catalog
- repo structure scaffold
- extension scaffold
- CI/config scaffold

Goal:
Create a stable base for implementation.

---

## Milestone B -- Core models and signal registry

Deliverables:
- TestSuiteSnapshot
- RepositoryMetadata
- Framework
- TestFile
- CodeUnit
- Signal schema
- signal registry

Goal:
Create the canonical language of the product.

---

## Milestone C -- Static analysis nucleus

Deliverables:
- repository scan
- test file discovery
- framework detection
- code-unit extraction
- test/code linkage

Goal:
Produce the first useful repository model.

---

## Milestone D -- Initial signal engine

Deliverables:
- slowTest
- flakyTest
- weakAssertion
- mockHeavyTest
- untestedExport
- coverageThresholdBreak

Goal:
Produce the first meaningful intelligence layer.

---

## Milestone E -- Risk engine

Deliverables:
- reliability risk
- change risk
- speed risk
- repo/package/module rollups

Goal:
Translate signals into engineering risk surfaces.

---

## Milestone F -- `hamlet analyze`

Deliverables:
- human-readable CLI output
- JSON snapshot output
- snapshot persistence

Goal:
Ship the first compelling user-facing vertical slice.

---

## Milestone G -- Extension prototype

Deliverables:
- Overview
- Health
- Quality
- Migration
- Review

Goal:
Render snapshot intelligence inside VS Code / Cursor.

---

## Milestone H -- Policy and CI checks [DONE]

Deliverables:
- policy config model (.hamlet/policy.yaml)
- policy loader with graceful missing-file handling
- governance signal evaluator
- `hamlet policy check` command with CI-friendly exit codes
- governance signals: policyViolation, legacyFrameworkUsage, runtimeBudgetExceeded
- policy rules: disallow_skipped_tests, disallow_frameworks, max_test_runtime_ms, minimum_coverage_percent, max_weak_assertions, max_mock_heavy_tests

Goal:
Prevent regression and support governance.

---

## Milestone I -- Ownership and review triage [DONE]

Deliverables:
- ownership resolver (explicit config, CODEOWNERS, directory fallback)
- ownership attached to test files and signals
- review grouping (by owner, type, directory, category)
- migration blocker grouping
- CLI analyze output with review sections
- extension Review and Migration view builders

Goal:
Turn findings into organized, triageable review work.

---

## Milestone J -- Migration intelligence [DONE]

Deliverables:
- migration detectors: DeprecatedPatternDetector, DynamicTestGenerationDetector,
  CustomMatcherDetector, FrameworkMigrationDetector
- blocker taxonomy: custom-matcher, dynamic-generation, deprecated-pattern,
  framework-helper, unsupported-setup
- migration readiness summary model (ReadinessSummary)
- migration reporting
- integrated into `hamlet analyze` pipeline

Goal:
Make migration intelligence concrete and actionable.

---

## Milestone K -- Historical comparison and trend detection [DONE]

Deliverables:
- timestamped snapshot archiving (.hamlet/snapshots/)
- comparison model (SnapshotComparison, SignalDelta, RiskDelta, FrameworkChange)
- `hamlet compare` command with --from/--to/--json flags
- human-readable and JSON comparison output
- representative new/resolved signal examples

Goal:
Enable local visibility into whether the test system is improving or deteriorating.

---

## Milestone L -- Benchmark-ready metrics [DONE]

Deliverables:
- aggregate metrics model (internal/metrics/metrics.go)
- metrics derivation from TestSuiteSnapshot (Derive function)
- `hamlet metrics` command with --root/--json flags
- human-readable metrics scorecard (RenderMetricsReport)
- privacy boundary: aggregate counts and ratios only, no raw paths or symbols
- metrics categories: structure, health, quality, change readiness, governance, risk

Goal:
Produce a normalized, privacy-conscious metrics artifact suitable for local
scorecards and future anonymous benchmarking aggregation.

---

## Milestone M -- Risk heatmap and leadership summary [DONE]

Deliverables:
- risk heatmap model (internal/heatmap/heatmap.go)
- directory and owner hotspot concentration
- overall posture band and summary
- leadership-oriented summary report (RenderSummaryReport)
- `hamlet summary` command with --root/--json flags
- tests for heatmap building, hotspot sorting, posture computation

Goal:
Provide leadership-oriented visibility into where risk concentrates and
the overall health posture of the test suite.

---

## Milestone N -- Benchmarking scaffolding [DONE]

Deliverables:
- benchmark export model (internal/benchmark/export.go)
- segmentation primitives (language, framework, size bucket, coverage/runtime/policy presence)
- benchmark-safe export builder (BuildExport)
- `hamlet export benchmark` command with --root flag
- privacy boundary enforced: aggregate-only, no raw paths or symbols
- tests for export building, segmentation, bucket computation

Goal:
Produce a benchmark-safe, segmented export artifact that enables future
cross-repo comparison without exposing proprietary code.

---

## Milestone O -- Executive summary and paid-product UX scaffolding [DONE]

Deliverables:
- executive summary model (internal/summary/executive.go)
- summary builder synthesizing heatmap + trends + benchmark readiness
- executive summary report renderer (RenderExecutiveSummary)
- `hamlet summary` enhanced with trend integration and benchmark readiness
- JSON executive summary output for future UI consumption
- trend highlights from local snapshot comparison (graceful degradation)
- benchmark readiness section (ready/limited dimensions, segmentation)
- evidence-based recommended focus generation
- example artifacts (executive-summary.txt, executive-summary.json)

Goal:
Make the local summary feel like a credible, paste-ready leadership report
and prepare the summary model for future paid-product UI reuse.

---

## Milestone P -- Release readiness [DONE]

Deliverables:
- README rewrite with quick start, commands, architecture, development
- docs navigation index (docs/README.md)
- demo walkthrough (docs/demo.md)
- release checklist (docs/release-checklist.md)
- Makefile with build/test/demo targets
- CLI help text polish
- extension coherence verification

Goal:
Make the repository coherent and approachable for first-time users
and contributors.

---

## Milestone Q -- Final pre-launch hardening [DONE]

Deliverables:
- stale/misleading messages removed
- compare error messages improved with actionable guidance
- CLI doc comment and help text aligned
- test expectations updated
- empty/error state review across all reports
- release checklist updated

Goal:
Leave the repository in a consistent, trustworthy release-candidate state.

---

## Later

- org-wide rollups
- benchmarking
- risk budgets
- cross-repo hosted product
