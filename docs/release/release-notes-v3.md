# Hamlet Release Notes

## Overview

Hamlet's current engine is a ground-up rewrite of the analysis engine in Go. It introduces a signal-first architecture that maps test suites into structured snapshots, evaluates them across five posture dimensions backed by 18 measurements, and surfaces portfolio intelligence, impact analysis, and migration readiness -- all from a single CLI.

This is a new product line, not an upgrade to the legacy converter. The legacy JavaScript converter engine (test framework conversion across 16 frameworks) remains functional and unchanged.

## Key Capabilities

### 10-Step Analysis Pipeline

Every `hamlet analyze` invocation runs a deterministic pipeline:

1. Static analysis (file discovery, framework detection, code unit extraction)
2. Policy loading
3. Signal detection via registry-based detector plugins
4. Ownership resolution (CODEOWNERS)
5. Runtime ingestion (optional: JUnit XML, Jest JSON)
6. Risk scoring
7. Coverage ingestion (optional: LCOV, Istanbul JSON)
8. Measurement-layer posture computation
9. Portfolio intelligence
10. Deterministic sorting and snapshot generation

### 18 Measurements Across 5 Posture Dimensions

Each dimension answers a specific question about your test suite:

| Dimension | Question | Measurements |
|-----------|----------|-------------|
| Health | Are our tests themselves reliable? | flaky share, skip density, dead test share, slow test share |
| Coverage Depth | Is our exported code actually tested? | uncovered exports, weak assertion share, coverage breach share |
| Coverage Diversity | Are we testing with the right mix? | mock-heavy share, framework fragmentation, E2E concentration, E2E-only units, unit test coverage |
| Structural Risk | Are we blocked from modernizing? | migration blocker density, deprecated pattern share, dynamic generation share |
| Operational Risk | Are we following our own rules? | policy violation density, legacy framework share, runtime budget breach share |

Every measurement produces a value, a band (strong/moderate/weak/elevated/critical), evidence strength, an explanation, and known limitations. Hamlet never pretends to know more than it does.

### Portfolio Intelligence

`hamlet portfolio` treats your test suite as an investment portfolio:

- **High-leverage tests**: tests that protect disproportionate code surface
- **Redundancy candidates**: overlapping tests that could be consolidated
- **Overbroad tests**: tests covering too many modules with shallow assertions
- **Low-value high-cost tests**: slow tests with minimal protection breadth
- **Runtime concentration**: what percentage of CI time is consumed by the top 20% of tests

### Impact Analysis

`hamlet impact` integrates with git to assess which changed files are tested, which are not, and where test gaps exist in the current diff.

- Drill-down views: `--show units`, `--show gaps`, `--show tests`, `--show owners`
- Owner filtering: `--owner "Team Platform"`
- Base ref control: `--base HEAD~5` or `--base main`

### Migration Readiness

`hamlet migration readiness` evaluates how prepared your codebase is for a framework migration:

- Per-area assessments with blocker counts and quality cross-references
- Blocker breakdown by type (API usage, dynamic generation, custom matchers, deprecated patterns)
- File-level and scope-level migration previews

### Trend Tracking

Save snapshots with `hamlet analyze --write-snapshot`, then compare them:

```
hamlet compare
hamlet compare --from baseline.json --to current.json
```

Trend highlights surface in `hamlet summary` automatically when prior snapshots exist.

### Policy Governance

Define rules in `.hamlet/policy.yaml` and enforce them:

```
hamlet policy check
```

Exit code 1 on violations, suitable for CI gates.

### Benchmark Exports

```
hamlet export benchmark
```

Produces a privacy-safe JSON export with no file paths, no source code, and no identifying information. Contains aggregate posture bands, measurement values, and portfolio summaries suitable for cross-repository comparison.

## Advanced Test Intelligence

### Test Lifecycle Model

Hamlet tracks test identity across snapshots using a two-phase approach:

- **Exact continuity**: tests with identical `TestID` across snapshots are matched with 1.0 confidence.
- **Heuristic continuity**: unmatched tests are scored using LCS-based similarity on test name, suite hierarchy, file path, and canonical identity. Classifications include `likely_rename`, `likely_move`, `likely_split`, `likely_merge`, and `ambiguous`, each with a confidence score and evidence basis.

Lifecycle analysis is the foundation for stability classification and trend tracking. It enables Hamlet to tell you "this test was renamed" rather than "a test was removed and a new one was added."

### Historical Stability Classes

Given 3+ snapshots, Hamlet classifies each test into one of 7 stability classes:

| Class | Meaning |
|-------|---------|
| `consistently_stable` | Low failure rate, low flake rate across history |
| `newly_unstable` | Was stable in early observations, recently started failing |
| `chronically_flaky` | Persistent flaky signals across multiple snapshots |
| `intermittently_slow` | Slow signals present in some but not all observations |
| `improving` | Recent observations are better than historical pattern |
| `quarantined_or_suppressed` | Skipped in majority of observations |
| `data_insufficient` | Fewer than 3 observations available |

Each classification includes a confidence score, history depth, and recent trend direction (improving, worsening, stable, insufficient).

### Quarantine and Suppression Detection

Hamlet detects 4 kinds of test suppression:

- **Quarantined**: tests in quarantine directories or with quarantine markers in file names
- **Expected failure**: tests with persistently low pass rates that continue to be present
- **Skip/disable**: tests with `skippedTest` signals, `xdescribe`/`xit` blocks, or skip patterns in paths
- **Retry wrapper**: tests with high retry rates (>=0.3) suggesting retry-as-policy patterns

Each suppression is classified by intent: **tactical** (likely temporary) vs. **chronic** (likely debt) vs. **unknown**. Detection sources include signals, runtime data, naming conventions, and config files.

### Failure Taxonomy

Test failures are classified into 8 categories using priority-ordered pattern matching on error messages and stack traces:

| Category | Confidence | Example Patterns |
|----------|-----------|-----------------|
| `snapshot_mismatch` | 0.95 | toMatchSnapshot, snapshot changed |
| `selector_or_ui_fragility` | 0.90 | element not found, stale element, cy.get() |
| `infrastructure_or_environment` | 0.95 | ENOMEM, permission denied, SIGSEGV |
| `dependency_or_service_failure` | 0.90 | ECONNREFUSED, 503, socket hang up |
| `timeout` | 0.90 | timed out, deadline exceeded |
| `setup_or_fixture_failure` | 0.75 | beforeEach, setup failed, fixture |
| `assertion_failure` | 0.80 | expect(), assertEqual, toBe |
| `unknown` | 0.20 | No pattern matched |

The dominant failure category is surfaced in the taxonomy result, enabling teams to prioritize remediation by failure type rather than by individual test.

### Common-Cause Clustering

Hamlet detects 5 types of common-cause clusters where shared dependencies drive broad test problems:

- **Shared import dependency**: a code unit linked by 3+ test files, making it a blast radius risk
- **Dominant slow helper**: a code unit shared by multiple slow tests, suggesting it contributes to CI slowness
- **Dominant flaky fixture**: a code unit shared by multiple flaky tests, a candidate root cause for non-determinism
- **Global setup path**: a directory where many test files share the same signal type, suggesting shared setup issues
- **Repeated failure pattern**: concentrated signal patterns within a directory

Each cluster includes a cause path, affected test count, confidence score, impact metric (e.g., total runtime affected), and a human-readable explanation.

### Assertion Strength Analysis

Each test file is assessed for assertion strength based on:

- **Assertion density**: assertions per test (strong >=3.0, moderate >=1.5, weak <1.0)
- **Mock ratio**: mocks vs. assertions (mock-heavy tests are classified as weak)
- **Snapshot dominance**: tests dominated by snapshot assertions with low density are classified as weak
- **E2E adjustment**: E2E frameworks (Cypress, Playwright, etc.) receive lower density thresholds

Categories include `strong`, `moderate`, `weak`, and `unclear`. The overall suite strength is an aggregate across all files.

### Test Environment Depth

Each test file is classified by environmental depth:

- **Browser runtime**: tests using Cypress, Playwright, Puppeteer, Selenium, or similar browser-backed frameworks
- **Real dependency usage**: tests with zero mocks using E2E or integration frameworks
- **Moderate mocking**: tests with mocks present but not dominant
- **Heavy mocking**: tests where mock count exceeds 2x assertion count or reaches 8+
- **Unknown**: insufficient evidence to classify

This is descriptive, not judgmental. Different depth classes carry different risk profiles: browser-runtime tests are more realistic but slower and more prone to environmental flakiness; heavy-mocking tests are fast but may miss integration issues.

### PR/Change-Scoped Workflows

The `hamlet pr` command performs change-scoped analysis against a git diff:

```
hamlet pr --base origin/main
```

It produces:
- **Posture band**: how well-protected is the changed code (well_protected, partially_protected, weakly_protected, high_risk)
- **Change-scoped findings**: protection gaps, existing signals on changed files, untested exports in the change area
- **Recommended tests**: which tests to run to validate the change
- **Affected owners**: which teams own the impacted code
- **Posture delta**: whether the change improves, worsens, or maintains test posture

### CLI Drill-Down Commands

The `hamlet show` command provides entity-level drill-downs:

```
hamlet show test src/__tests__/auth.test.js   # Test file details with signals
hamlet show unit AuthService                   # Code unit coverage and ownership
hamlet show owner team-platform                # Owner portfolio with signals
hamlet show finding untested_export            # Finding details with actions
```

Each entity view includes a "Next:" hint suggesting a related drill-down, creating a navigable investigation flow.

### Output Modes

The `hamlet pr` command supports multiple output formats:

| Format | Flag | Use Case |
|--------|------|----------|
| Terminal report | (default) | Human-readable CLI output |
| Markdown | `--format markdown` | GitHub/GitLab PR comments |
| CI annotation | `--format annotation` | GitHub Actions `::error`/`::warning` annotations |
| Concise comment | `--format comment` | One-line status for inline use |
| JSON | `--json` | Programmatic consumption |

All commands support `--json` for machine-readable output.

## What's New

- **Signal-first architecture**: all findings trace back to detected signals with evidence metadata
- **Measurement framework**: 18 measurements with transparent computation, evidence strength, and stated limitations
- **Portfolio intelligence**: cost/value/leverage analysis for test assets
- **Impact analysis framework**: formal impact graph connecting code units to tests with edge provenance, change-risk posture with 4 dimensions (protection, exposure, coordination, instability), protective test-set selection with explanation metadata, and privacy-safe aggregate exports
- **Impact graph**: durable code-unit-to-test graph built from coverage lineage, structural links, and naming conventions, with exact/inferred/weak confidence classification per edge
- **Protective test selection**: `hamlet select-tests` recommends a focused test set for a change, with selection reasoning and gap reporting
- **Test identity tracking**: stable identifiers for test files and code units across snapshots
- **Benchmark exports**: privacy-safe aggregate exports for cross-repo comparison
- **Registry-based detector plugins**: extensible signal detection without modifying the engine core
- **Snapshot comparison**: structural diff between analysis runs with trend detection
- **Ownership-aware analysis**: CODEOWNERS integration for per-team risk attribution
- **Coverage and runtime enrichment**: optional ingestion of LCOV, Istanbul JSON, JUnit XML, and Jest JSON artifacts
- **Advanced test intelligence**: 7 assessment subsystems (lifecycle, stability, suppression, failure taxonomy, clustering, assertion strength, environment depth) with confidence metadata and cautious classification
- **PR/change-scoped workflows**: `hamlet pr` with markdown, CI annotation, and concise comment output formats
- **Entity drill-down**: `hamlet show` for test files, code units, owners, and findings

## Commands Reference

| Command | Description |
|---------|-------------|
| `hamlet analyze` | Full test suite analysis with human-readable output |
| `hamlet summary` | Executive summary with posture, risk areas, trends, and benchmark readiness |
| `hamlet posture` | Detailed posture breakdown with measurement evidence |
| `hamlet portfolio` | Portfolio intelligence: cost, breadth, leverage, redundancy |
| `hamlet impact` | Impact analysis for changed code against git diff |
| `hamlet impact --show graph` | View impact graph edges for impacted units |
| `hamlet impact --show selected` | View protective test set with selection reasoning |
| `hamlet select-tests` | Recommend protective test set for a change |
| `hamlet metrics` | Aggregate metrics scorecard |
| `hamlet migration readiness` | Migration readiness assessment |
| `hamlet migration blockers` | List migration blockers by type and area |
| `hamlet migration preview` | Preview migration for a file or scope |
| `hamlet compare` | Compare two snapshots for trend tracking |
| `hamlet policy check` | Evaluate local policy rules |
| `hamlet export benchmark` | Privacy-safe JSON export for benchmarking |
| `hamlet pr` | PR-scoped analysis with format options (markdown, annotation, comment) |
| `hamlet show test <path>` | Drill into a specific test file |
| `hamlet show unit <name>` | Drill into a specific code unit |
| `hamlet show owner <name>` | Drill into an owner's portfolio |
| `hamlet show finding <type>` | Drill into a finding type |

All commands support `--json` for machine-readable output and `--root PATH` to target a specific directory.

## Breaking Changes

None. This is a new engine and a new binary. The legacy JavaScript converter engine (`node bin/hamlet.js`) is unchanged and continues to work independently.

## Known Limitations

- **Static analysis confidence**: test-to-code linkage is heuristic-based (naming conventions, import analysis). Some coverage relationships may not be detected. Evidence strength metadata indicates where confidence is limited.
- **Flaky/slow test detection**: without runtime artifacts (`--runtime`), flaky and slow test signals rely on code-level heuristics (retry patterns, timeout values). Provide JUnit XML or Jest JSON for high-confidence runtime signals.
- **Coverage enrichment**: coverage-dependent measurements (E2E-only units, unit test coverage, coverage breach share) require separate LCOV or Istanbul JSON artifacts passed via `--coverage`.
- **No hosted benchmarking**: benchmark exports are local JSON files. Cross-repository comparison requires manual aggregation. A hosted benchmarking service is planned but not yet available.
- **Single-language heuristics**: code unit extraction and framework detection are strongest for JavaScript/TypeScript, Java, and Python. Other languages receive basic file-level analysis.

## Getting Started

### Install

```bash
go install github.com/pmclSF/hamlet/cmd/hamlet@latest
```

Or build from source:

```bash
git clone https://github.com/pmclSF/hamlet.git
cd hamlet
go build -o hamlet ./cmd/hamlet
```

### First Command

```bash
cd your-repo
hamlet analyze
```

Then progressively drill down:

```bash
hamlet summary           # leadership-ready overview
hamlet posture           # measurement evidence
hamlet portfolio         # cost and leverage analysis
hamlet impact            # what changed and is it tested?
```

## Verification System

This release includes a formal, layered verification system:

- **Unit tests** — deterministic core logic across identity, lifecycle, ownership, coverage, measurement, scoring, and impact
- **Golden tests** — 9 golden outputs protecting metrics, exports, reports, and impact analysis artifacts
- **Determinism tests** — 9 repeat-run tests ensuring identical outputs across runs for all critical subsystems
- **Schema tests** — 9 round-trip and forward-compatibility tests for persisted artifacts
- **Adversarial tests** — 19 tests for nil/empty/malformed/partial inputs with honest degradation
- **E2E scenario tests** — 14 flagship workflow tests from analysis to rendering
- **CLI regression tests** — 13 tests covering build, help, commands, output structure, and JSON modes
- **Benchmarks** — 11 Go benchmarks for metrics, measurements, heatmap, scoring, impact, comparison, portfolio, and full pipeline
- **15 fixture factories** — realistic scenarios from minimal repos to 2000+ test file stress tests

Release gates: `make pr-gate` (fast, every PR) and `make release-gate` (full verification before release).
