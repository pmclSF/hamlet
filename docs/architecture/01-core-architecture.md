# Core Architecture

> **Status:** Implemented
> **Purpose:** Detailed walkthrough of the signal pipeline layers and dependency graph analysis within the Go engine.
> **Key decisions:**
> - Five-layer signal pipeline: static analysis, runtime ingestion, signal detection, risk scoring, reporting
> - Registry-based detector plugins for extensible signal detection
> - Dependency graph is read-only after construction — build once, then query
> - Insight engines are pure functions over the graph (no side effects, no cross-engine dependencies)

## Signal Pipeline (Go)

The signal pipeline is built as a layered pipeline. Each layer transforms data and passes it to the next.

### Layer 1: Static Analysis

Repository scanning, framework detection, and test file discovery.

- **Framework detection** identifies test frameworks by analyzing package manifests, config files, and import patterns.
- **Test file discovery** finds test files using framework-specific naming conventions and directory structures.
- **Code unit extraction** normalizes source code into queryable units (functions, classes, exports).
- **Test identity** produces deterministic hashes for each test case to enable stable cross-snapshot references.

### Layer 2: Runtime Ingestion

Optional ingestion of runtime artifacts for health signals.

- JUnit XML results (test durations, failures, pass/fail status)
- Jest JSON results (same, plus snapshot metadata)
- Coverage data (LCOV, Istanbul JSON) for structural coverage attribution

### Layer 3: Signal Detection

Registry-based detector plugins scan the repository and runtime artifacts to emit structured signals.

Signal categories:
- **Health:** slowTest, flakyTest, skippedTest, deadTest, unstableSuite
- **Quality:** untestedExport, weakAssertion, mockHeavyTest, snapshotHeavyTest, coverageBlindSpot
- **Migration:** frameworkMigration, migrationBlocker, deprecatedTestPattern
- **Governance:** policyViolation, legacyFrameworkUsage, runtimeBudgetExceeded

Each signal includes: type, category, severity, confidence, location (file + line), evidence, explanation, and suggested action.

### Layer 4: Risk Scoring

Signals are aggregated into risk surfaces across three dimensions:

- **Reliability risk** — derived from flaky tests, unstable suites, dead tests
- **Change risk** — derived from migration blockers, deprecated patterns, framework fragmentation
- **Speed risk** — derived from slow tests, runtime budget violations

Risk is computed at multiple levels: repository, package, directory, and owner.

### Layer 5: Reporting

Snapshot serialization, human-readable reports, JSON output, and executive summaries.

The canonical artifact is `TestSuiteSnapshot` — a serialized representation of the entire analysis. Snapshots enable trend comparison across time.

## Dependency Graph Layer (Go)

The dependency graph layer operates on a different model: a typed dependency graph rather than a signal pipeline. It lives in `internal/depgraph/` (graph construction and traversal) and `internal/graph/` (snapshot index for flat lookups).

### Graph Construction

1. **Test discovery** — find all test files using framework-specific patterns
2. **Graph population** — create nodes for test files, suites, and individual tests
3. **Import analysis** — resolve import statements to build edges between nodes
4. **File classification** — identify fixtures, helpers, and source files
5. **Fixture/helper detection** — add specialized edges for test-to-fixture and test-to-helper relationships

### Insight Engines

Each engine receives the constructed graph and produces a typed result:

- **Coverage engine** (`internal/depgraph/coverage.go`) — reverse traversal from source files to find covering tests
- **Duplicate engine** (`internal/depgraph/duplicate.go`) — structural fingerprinting to detect redundant test clusters
- **Fanout engine** (`internal/depgraph/fanout.go`) — BFS to measure transitive dependency counts
- **Impact engine** (`internal/depgraph/impact.go`) — forward traversal from changed files to find affected tests

### Analysis Layer

Built on top of engine results:

- **Repository profiling** — classify the repo across eight dimensions
- **Edge case detection** — identify 14 conditions that affect analysis reliability
- **Policy application** — adjust behavior based on detected edge cases
- **Explainability** — trace dependency paths with human-readable reasons

## Package Map

```
cmd/terrain/              CLI entry point (Go)
internal/
  analysis/              Repository scanning, framework detection
  benchmark/             Privacy-safe benchmark export
  comparison/            Snapshot trend comparison
  coverage/              Coverage ingestion and attribution
  depgraph/              Typed dependency graph: nodes, edges, traversal,
                           insight engines (coverage, duplicate, fanout, impact),
                           repository profiling, edge case detection
  engine/                Pipeline orchestration, detector registry
  governance/            Policy evaluation
  graph/                 Snapshot index (flat lookup by test ID, file, etc.)
  health/                Health detectors (slow, flaky, skipped)
  heatmap/               Risk concentration model
  identity/              Test identity hashing
  impact/                Change-scope impact analysis
  measurement/           Posture measurement framework
  metrics/               Aggregate metric derivation
  migration/             Migration detectors, readiness model
  models/                Canonical data models
  ownership/             Ownership resolution
  policy/                Policy config and YAML loader
  quality/               Quality signal detectors
  reporting/             Report renderers
  runtime/               Runtime artifact ingestion
  scoring/               Explainable risk engine
  signals/               Signal detector interface and registry
  summary/               Executive summary builder
  testcase/              Test case extraction
  testtype/              Test type inference

src/                     Legacy JavaScript converter engine (ES modules)
  core/                  BaseConverter, ConverterFactory, PatternEngine, etc.
  converters/            E2E converter classes (Cypress/Playwright/Selenium)
  languages/             Framework definitions by language (Java, JS, Python)
  utils/                 Helpers, reporter
```
