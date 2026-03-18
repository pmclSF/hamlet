# Terrain

**Map your test terrain.**

Terrain is a test system intelligence platform. It reads your repository — test code, source structure, coverage data, runtime artifacts, ownership files, and local policy — and builds a structural model of how your tests relate to your code. From that model it surfaces risk, quality gaps, redundancy, fragile dependencies, and migration readiness, all without running a single test.

The core idea: every codebase has a *test terrain* — the shape of its testing infrastructure, the density of coverage across areas, the hidden fault lines where a fixture change cascades into thousands of tests. Terrain makes that shape visible and navigable so you can make informed decisions about what to test, what to fix, and where to invest.

## What "Test Terrain" Means

Most teams know what tests they have. Few teams understand the *terrain* underneath:

- Which source files have no structural test coverage?
- Which shared fixtures fan out to thousands of tests, making any change to them a blast-radius problem?
- Which test clusters are near-duplicates burning CI time?
- Which areas have tests but weak assertions that wouldn't catch a real regression?
- When you change `auth/session.ts`, which 41 of your 18,000 tests actually matter?

Test terrain is the structural topology of your test system — the dependency graph, the coverage landscape, the duplication clusters, the fanout hotspots, the skip debt. Terrain maps it.

## Problems Terrain Solves

**"We don't know the state of our test system."** Teams inherit test suites they didn't write. Terrain gives a baseline in seconds: framework mix, coverage confidence, duplication, risk posture.

**"CI takes too long and we don't know what to cut."** Terrain identifies redundant tests, high-fanout fixtures, and confidence-based test selection — showing where CI time is wasted and what can be safely reduced.

**"We changed auth code — what tests should we worry about?"** Terrain traces your change through the import graph and structural dependencies, returning the impacted tests ranked by confidence, with reason chains explaining each selection.

**"A tool flagged something but won't explain why."** Every Terrain finding carries an evidence chain. `terrain explain` shows exactly what signals, dependency paths, and scoring rules produced each decision.

**"We're migrating frameworks and need to know what's blocking us."** Migration readiness, blockers by type, and preview-scoped difficulty assessment — all derived from static analysis.

## The Four Canonical Workflows

Terrain is organized around four questions. Everything else is a supporting view.

```
terrain analyze     "What is the state of our test system?"
terrain insights    "What should we fix in our test system?"
terrain impact      "What validations matter for this change?"
terrain explain     "Why did Terrain make this decision?"
```

### 1. Analyze — understand the test system

```bash
terrain analyze
```

```
Terrain Test Suite Analysis
══════════════════════════════════════════════════

Repository Profile
  Test volume:          very large
  CI pressure:          high
  Coverage confidence:  medium
  Redundancy level:     medium
  Fanout burden:        high

Tests detected:         52,341 across 1,047 test files
Frameworks:             pytest (100%)

Weak coverage areas:
  pandas/io/sas/          2 test files, no parametrize coverage for edge encodings
  pandas/core/internals/  block manager has 4 tests, low relative to complexity
  pandas/plotting/        matplotlib integration tests skip without display backend

CI optimization potential:
  Estimated 40% runtime reduction with confidence-based test selection
  187 tests marked @pytest.mark.slow — clustered in io/ and groupby/

Risk Posture
  Quality:     medium risk   (weak assertions in 23 test files)
  Reliability: high risk     (network-dependent tests, xfail clusters)
  Speed:       high risk     (slow markers, single_cpu constraints)
  Governance:  low risk

Signals: 1,204 total (38 critical, 187 high, 412 medium, 567 low)

Next: terrain insights    see what to improve
      terrain impact      analyze a specific change
```

### 2. Insights — find what to improve

```bash
terrain insights
```

```
Top improvement opportunities:
  1. Reduce conftest.py fixture fanout in tests/frame/
     why: dataframe_with_arrays fixture fans out to 3,100 tests —
          any change retriggers the entire frame/ suite
     where: pandas/tests/frame/conftest.py

  2. Add structural tests for pandas/core/internals/
     why: Block manager is critical infrastructure with minimal test density
     where: pandas/core/internals/

  3. Review 34 xfail(strict=False) markers older than 6 months
     why: Loose xfail masks real regressions — either fix or remove
     where: pandas/tests/io/, pandas/tests/indexing/

  4. Consolidate duplicate GroupBy aggregation tests
     why: 8 tests across 3 files with 0.91 similarity — redundant CI cost
     where: pandas/tests/groupby/

  5. Split network-dependent I/O tests into isolated suite
     why: @pytest.mark.network tests fail intermittently, blocking unrelated PRs
     where: pandas/tests/io/json/, pandas/tests/io/html/
```

### 3. Impact — understand what a change affects

```bash
terrain impact --base main
```

```
Terrain Impact Analysis
══════════════════════════════════════════════════

Changed areas:
  core/groupby           pandas/core/groupby/groupby.py (+8 -2)

Impacted tests:          127 / 52,341

Selected tests (top 10):
  [high]   tests/groupby/test_groupby.py                confidence: 0.96
  [high]   tests/groupby/test_apply.py                   confidence: 0.93
  [high]   tests/groupby/test_grouper.py                 confidence: 0.88
  [medium] tests/resample/test_base.py                   confidence: 0.61
  [medium] tests/frame/methods/test_describe.py          confidence: 0.54
  ...and 117 more

Insights:
  conftest.py fixture path amplifies impact — 74 of 127 tests reached via
  shared fixtures, not direct imports. Consider targeted test run:
    pytest tests/groupby/ -x -q
```

### 4. Explain — understand why

```bash
terrain explain pandas/tests/io/json/test_pandas.py
```

```
Test File: pandas/tests/io/json/test_pandas.py
Framework: pytest
Tests: 84    Assertions: 312
Runtime: 4.2s    Pass rate: 96%    Retry rate: 4%

Signals (4):
  [high]   networkDependency: 12 tests use @pytest.mark.network — flaky in CI
  [medium] slowTest: 4.2s runtime exceeds 2s threshold
  [medium] weakAssertion: 8 bare assert statements without descriptive messages
  [low]    xfailAccumulation: 3 xfail markers older than 180 days
```

See [Canonical User Journeys](docs/product/canonical-user-journeys.md) for the full workflow specification and [example outputs](docs/examples/) for detailed report samples.

## Product Philosophy

**Inference first.** Terrain reads code. It parses imports, detects frameworks, resolves coverage relationships, and builds dependency graphs from what already exists in the repository. No annotations, no test tagging, no SDK integration required.

**Zero-config by default.** `terrain analyze` works on any repository with test files. Coverage data, runtime artifacts, ownership files, and policy rules are optional inputs that enrich the model — but the core analysis requires nothing beyond the code itself.

**Explainability over magic.** Every finding carries evidence: which signal type, what confidence level, what dependency path, what scoring rule. `terrain explain` exposes the full reasoning chain behind any decision. Teams should never wonder *why* Terrain said something.

**Conservative under uncertainty.** When Terrain encounters ambiguity — a dependency path with low confidence, a file that might or might not be a test — it flags the uncertainty rather than guessing. Impact analysis uses fallback policies with explicit confidence penalties rather than silently expanding scope.

**System health, not individual productivity.** Terrain measures the test system. It never attributes quality to individual developers. Ownership information is used for routing and triage, not scoring.

## Who Uses Terrain

Terrain is framework-agnostic and language-aware. The same analysis model applies across:

- **Frontend teams** — React/Vue component tests, Playwright/Cypress E2E suites, Vitest/Jest unit tests
- **Backend teams** — Go test suites, pytest collections, JUnit hierarchies, integration test infrastructure
- **Mobile teams** — XCTest, Espresso, and cross-platform test suites
- **QA / SDET** — test portfolio management, coverage gap analysis, migration planning across frameworks
- **SRE / Platform** — CI optimization, test selection for pipelines, policy enforcement
- **AI / ML evaluation** — evaluation suite structure, benchmark test management, coverage across model behaviors

The structural model is the same. The signals and recommendations adapt to the framework and test patterns detected.

## How CI Optimization Emerges

Terrain does not start with CI optimization. It starts with understanding.

When you run `terrain analyze`, Terrain builds a structural model: which tests exist, which source files they cover, how they depend on shared fixtures, where duplication lives. From that model, CI optimization *emerges*:

- **Test selection** — `terrain impact` traces a diff through the dependency graph and returns only the tests that structurally matter, ranked by confidence. This is not a heuristic skip list — it is a graph traversal with evidence.
- **Redundancy reduction** — `terrain insights` surfaces duplicate test clusters. Removing or consolidating them directly reduces CI time without reducing coverage.
- **Fanout control** — High-fanout fixtures that trigger thousands of tests on any change are identified and prioritized for splitting.
- **Confidence-based runs** — Impact analysis assigns confidence scores. CI pipelines can run high-confidence tests immediately and defer low-confidence tests to nightly runs.

The result is faster CI that comes from *understanding the test system*, not from skipping tests and hoping for the best.

## Quick Start

```bash
# Install
go install github.com/pmclSF/terrain/cmd/terrain@latest

# Or build from source
git clone https://github.com/pmclSF/terrain.git
cd terrain
go build -o terrain ./cmd/terrain

# Detect coverage/runtime data paths (recommended first step)
terrain init

# Analyze the current repository
terrain analyze

# JSON output for any command
terrain analyze --json
```

### Requirements

- Go 1.23 or later

### Legacy Test Converter (npm)

The npm package `terrain-testframework` provides a separate CLI for **test framework conversion** (migrating tests between Jest, Vitest, Playwright, Cypress, pytest, JUnit, and 10 more frameworks). It is distinct from the Go-based analysis CLI above.

```bash
# Install the converter
npm install -g terrain-testframework

# Convert tests
terrain-convert convert src/tests/ --from jest --to vitest -o converted/

# Shorthand aliases
terrain-convert cy2pw src/e2e/ -o converted/
```

The old `terrain` npm binary name is deprecated. Use `terrain-convert` instead to avoid conflicts with the Go CLI.

## Commands

### Primary commands

| Command | Question |
|---------|----------|
| `terrain analyze` | What is the state of our test system? |
| `terrain insights` | What should we fix in our test system? |
| `terrain impact` | What validations matter for this change? |
| `terrain explain <target>` | Why did Terrain make this decision? |

### Supporting commands

| Command | Purpose |
|---------|---------|
| `terrain init` | Detect data files and print recommended analyze command |
| `terrain summary` | Executive summary with risk, trends, benchmark readiness |
| `terrain focus` | Prioritized next actions |
| `terrain posture` | Detailed posture breakdown with measurement evidence |
| `terrain portfolio` | Portfolio intelligence: cost, breadth, leverage, redundancy |
| `terrain metrics` | Aggregate metrics scorecard |
| `terrain compare` | Compare two snapshots for trend tracking |
| `terrain select-tests` | Recommend protective test set for a change |
| `terrain pr` | PR/change-scoped analysis |
| `terrain show <type> <id>` | Drill into test, unit, owner, or finding |
| `terrain migration <sub>` | Migration readiness, blockers, or preview |
| `terrain policy check` | Evaluate local policy rules |
| `terrain export benchmark` | Privacy-safe JSON export for benchmarking |

### AI / eval

| Command | Purpose |
|---------|---------|
| `terrain ai list` | List detected scenarios, prompts, datasets, eval files |
| `terrain ai doctor` | Validate AI/eval setup and surface configuration issues |
| `terrain ai run` | Execute eval scenarios and collect results (planned) |
| `terrain ai record` | Record eval results as a baseline snapshot (planned) |
| `terrain ai baseline` | Manage eval baselines (planned) |

### Advanced / debug

| Command | Purpose |
|---------|---------|
| `terrain debug graph` | Dependency graph statistics |
| `terrain debug coverage` | Structural coverage analysis |
| `terrain debug fanout` | High-fanout node analysis |
| `terrain debug duplicates` | Duplicate test cluster analysis |
| `terrain debug depgraph` | Full dependency graph analysis (all engines) |

All commands support `--root PATH` and `--json` flags. Run `terrain --help` for full flag documentation.

## Architecture Overview

Terrain is built around a signal-first architecture:

```
Repository scan  →  Signal detection  →  Risk modeling  →  Reporting
     │                    │                    │               │
  test files         framework-specific    explainable     human-readable
  source files       pattern detectors     risk scoring    + JSON output
  coverage data      quality signals       with evidence
  runtime artifacts  health signals        chains
  ownership files    migration signals
  policy rules       governance signals
```

- **Signals** are the core abstraction — every finding is a structured signal with type, severity, confidence, evidence, and location
- **Snapshots** (`TestSuiteSnapshot`) are the canonical serialized artifact — the complete structural model of a test system at a point in time
- **Risk surfaces** are derived from signals with explainable scoring across dimensions (quality, reliability, speed, governance)
- **Dependency graphs** model import relationships, fixture fanout, and structural coverage
- **Reports** synthesize signals, risk, trends, and benchmark readiness into actionable output

```
cmd/terrain/          CLI entry point and command routing
cmd/terrain-bench/    Benchmark harness for cross-repo CLI validation
internal/
├── analysis/        Repository scanning, framework detection, code surface inference
├── analyze/         Analyze report builder (depgraph aggregation)
├── benchmark/       Privacy-safe benchmark export and assessment scoring
├── comparison/      Snapshot-to-snapshot trend comparison
├── coverage/        Coverage ingestion (LCOV, Istanbul) and attribution
├── depgraph/        Dependency graph: 20 node types, 15 edge types, 5 reasoning engines
├── engine/          Pipeline orchestration and detector registry
├── explain/         Structured explanation builder (tests + scenarios)
├── gauntlet/        Gauntlet AI eval artifact ingestion
├── governance/      Policy evaluation and governance signals
├── graph/           Import graph construction
├── health/          Runtime-backed health detectors (slow, flaky, skipped)
├── heatmap/         Risk concentration model (directory and owner hotspots)
├── identity/        Test identity hashing and normalization
├── impact/          Change-scope impact analysis (tests + scenarios)
├── insights/        Prioritized health report and findings
├── matrix/          Environment/device matrix analysis
├── measurement/     Posture measurement framework (5 dimensions, 18 measurements)
├── metrics/         Aggregate metric derivation
├── migration/       Migration detectors, readiness model, preview boundary
├── models/          Canonical data models (Signal, Snapshot, CodeSurface, Scenario, etc.)
├── ownership/       Ownership resolution (CODEOWNERS, config, directory)
├── policy/          Policy + terrain.yaml config (scenarios, manual coverage)
├── portfolio/       Portfolio intelligence (cost, breadth, leverage, redundancy)
├── quality/         Quality signal detectors
├── reporting/       Report renderers (analyze, impact, insights, posture, etc.)
├── runtime/         Runtime artifact ingestion (JUnit XML, Jest JSON)
├── scoring/         Explainable risk engine (reliability, change, speed)
├── signals/         Signal detector interface, registry, runner
├── stability/       Stability clustering (shared root cause detection)
├── summary/         Executive summary builder
├── testcase/        Test case extraction and identity collision detection
└── testtype/        Test type inference (unit, integration, e2e)
```

See [DESIGN.md](DESIGN.md) for the full architecture overview, [docs/architecture/](docs/architecture/) for detailed design documents, and [docs/json-schema.md](docs/json-schema.md) for JSON output structure.

## Snapshot Workflow

Terrain supports local snapshot history for trend tracking:

```bash
# Save a snapshot
terrain analyze --write-snapshot

# Later, save another snapshot
terrain analyze --write-snapshot

# Compare the two most recent snapshots
terrain compare

# Executive summary automatically includes trend highlights
terrain summary
```

Snapshots are stored in `.terrain/snapshots/` as timestamped JSON files.

## Policy

Define local policy rules in `.terrain/policy.yaml`:

```yaml
rules:
  disallow_skipped_tests: true
  max_weak_assertions: 10
  max_mock_heavy_tests: 5
```

Then check compliance:

```bash
terrain policy check        # human-readable output
terrain policy check --json # JSON output for CI
```

Exit code 0 = pass, 2 = violations found, 1 = error.

## Development

```bash
# Build
go build -o terrain ./cmd/terrain

# Test all Go packages
go test ./internal/... ./cmd/...

# Test with verbose output
go test -v ./internal/...

# Legacy JavaScript tests (requires Node.js 22+)
npm test
```

## Documentation

- [Canonical User Journeys](docs/product/canonical-user-journeys.md) — primary workflows and expected outcomes
- [Example Reports](docs/examples/) — analyze, impact, insights, explain output samples
- [Architecture](docs/architecture/) — design documents and technical specifications
- [CLI Specification](docs/cli-spec.md) — full command and flag reference
- [Signal Model](docs/signal-model.md) — the core signal abstraction
- [Engineering](docs/engineering/) — contributor-facing architecture maps and implementation details
- [Legacy Converter](docs/legacy/) — historical JavaScript converter engine documentation

## License

MIT License — see [LICENSE](LICENSE) for details.
