# Hamlet

**Signal-first test intelligence for engineering teams**

Hamlet analyzes repository structure, test code, runtime artifacts, coverage data, and local policy to surface the real state of your test system — risk, quality, migration readiness, and governance — without running tests.

## Quick Start

```bash
# Install
go install github.com/pmclSF/hamlet/cmd/hamlet@latest

# Or build from source
git clone https://github.com/pmclSF/hamlet.git
cd hamlet
go build -o hamlet ./cmd/hamlet

# Detect coverage/runtime paths (recommended first step)
hamlet init

# Analyze the current repository
hamlet analyze

# Executive summary with risk, trends, and benchmark readiness
hamlet summary

# Aggregate metrics scorecard
hamlet metrics

# JSON output for any command
hamlet analyze --json
hamlet summary --json
hamlet metrics --json
```

### Requirements

- Go 1.23 or later

## Commands

| Command | Description |
|---------|-------------|
| `hamlet analyze` | Analyze repository test suite — frameworks, signals, risk |
| `hamlet summary` | Executive summary — posture, trends, focus, benchmark readiness |
| `hamlet insights` | Prioritized improvement actions with rationale |
| `hamlet explain <entity>` | Evidence chain for a test, code unit, owner, or finding |
| `hamlet focus` | Where to concentrate testing effort based on risk and gaps |
| `hamlet portfolio` | Portfolio view — coverage breadth, test type distribution, risk allocation |
| `hamlet posture` | Detailed posture breakdown with measurement evidence |
| `hamlet metrics` | Aggregate metrics scorecard (privacy-safe, benchmark-ready) |
| `hamlet impact` | Impact analysis for changed code (git diff-aware) |
| `hamlet select-tests` | Protective test selection for a git diff |
| `hamlet pr` | PR analysis — impact, test selection, and risk for review |
| `hamlet show <type> <id>` | Drill into a specific test, unit, owner, or finding |
| `hamlet migration readiness` | Migration readiness assessment with quality factors |
| `hamlet migration blockers` | List migration blockers by type and area |
| `hamlet migration preview` | Preview migration difficulty for a file or scope |
| `hamlet compare` | Compare two snapshots and show trend changes |
| `hamlet policy check` | Evaluate repository against local policy rules |
| `hamlet export benchmark` | Output benchmark-safe JSON export for future comparison |
| `hamlet init` | Detect coverage/runtime paths and print recommended analyze command |
| `hamlet version` | Show version, commit, and build date |

Run `hamlet --help` for full flag documentation.

## Walkthrough: Using Hamlet on pandas

Imagine you maintain [pandas](https://github.com/pandas-dev/pandas) — 1,000+ test files, ~52,000 test cases across pytest, and CI runs that take significant time. Here's what Hamlet tells you.

### Step 1: Analyze

```bash
cd pandas
hamlet analyze
```

```
Hamlet Test Suite Analysis
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

Next: hamlet insights    see what to improve
      hamlet impact      analyze a specific change
```

### Step 2: Insights — what to fix first

```bash
hamlet insights
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

### Step 3: Impact — what does your PR actually touch?

You're working on a fix in `pandas/core/groupby/groupby.py`. Before pushing:

```bash
hamlet impact --base main
```

```
Hamlet Impact Analysis
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

### Step 4: Explain — why did Hamlet flag something?

```bash
hamlet explain pandas/tests/io/json/test_pandas.py
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

### The pattern

```
hamlet analyze     →  "What is the state of our test system?"
hamlet insights    →  "What should we fix first?"
hamlet impact      →  "What tests matter for this PR?"
hamlet explain     →  "Why was this flagged?"
```

See [Canonical User Journeys](docs/product/canonical-user-journeys.md) for the full workflow and [example outputs](docs/examples/).

## What Hamlet Reveals

### Structure
Framework inventory, test file discovery, code-to-test relationships, ownership mapping.

### Health
Flaky tests, slow tests, skipped tests, dead tests, unstable suites.

### Quality
Weak assertions, mock-heavy tests, untested exports, coverage blind spots.

### Migration Readiness
Migration blockers, deprecated patterns, legacy framework drift, framework fragmentation.

### Policy and Governance
Local policy rules, violation tracking, compliance enforcement in CI.

### Risk
Explainable risk surfaces by dimension (reliability, change, speed) with directory and owner concentration.

### Benchmark-Safe Exports
Privacy-safe aggregate metrics for future cross-repo comparison — no raw paths or source code exposed.

## Snapshot Workflow

Hamlet supports local snapshot history for trend tracking:

```bash
# Save a snapshot
hamlet analyze --write-snapshot

# Later, save another snapshot
hamlet analyze --write-snapshot

# Compare the two most recent snapshots
hamlet compare

# Executive summary automatically includes trend highlights
hamlet summary
```

Snapshots are stored in `.hamlet/snapshots/` as timestamped JSON files.

## Policy

Define local policy rules in `.hamlet/policy.yaml`:

```yaml
rules:
  disallow_skipped_tests: true
  max_weak_assertions: 10
  max_mock_heavy_tests: 5
```

Then check compliance:

```bash
hamlet policy check        # human-readable output
hamlet policy check --json # JSON output for CI
```

Exit code 0 = pass, 2 = violations found, 1 = execution/error conditions.

## Migration Workflow

Hamlet started with migration pain — "how hard will this migration be?" The current engine turns that pain into broader test intelligence while keeping migration as a first-class workflow:

```bash
# Assess migration readiness
hamlet migration readiness

# List specific blockers
hamlet migration blockers

# Review policy and governance
hamlet policy check

# Save snapshot, fix issues, save another, then compare
hamlet analyze --write-snapshot
# ... fix blockers ...
hamlet analyze --write-snapshot
hamlet compare
```

## Architecture

Hamlet is built around a signal-first architecture:

```
Repository scan → Signal detection → Risk modeling → Reporting
```

- **Signals** are the core abstraction — every finding is a structured signal
- **Snapshots** are the canonical serialized artifact (`TestSuiteSnapshot`)
- **Risk surfaces** are derived from signals with explainable scoring
- **Reports** synthesize signals, risk, trends, and benchmark readiness

See [DESIGN.md](DESIGN.md) for architecture overview and [docs/](docs/) for detailed documentation.
JSON output structure is documented in [docs/json-schema.md](docs/json-schema.md).

## Project Structure (Go Engine)

```
cmd/hamlet/          CLI entry point
internal/
├── analysis/        Repository scanning, framework detection, test file discovery
├── benchmark/       Privacy-safe benchmark export and segmentation
├── comparison/      Snapshot-to-snapshot trend comparison
├── coverage/        Coverage ingestion (LCOV, Istanbul) and attribution
├── engine/          Pipeline orchestration and detector registry
├── governance/      Policy evaluation and governance signals
├── health/          Runtime-backed health detectors (slow, flaky, skipped)
├── heatmap/         Risk concentration model (directory and owner hotspots)
├── identity/        Test identity hashing and normalization
├── impact/          Change-scope impact analysis
├── measurement/     Posture measurement framework
├── metrics/         Aggregate metric derivation
├── migration/       Migration detectors, readiness model, preview boundary
├── models/          Canonical data models (Signal, Snapshot, Risk, etc.)
├── ownership/       Ownership resolution (CODEOWNERS, config, directory)
├── policy/          Policy config model and YAML loader
├── quality/         Quality signal detectors
├── reporting/       Human-readable report renderers
├── runtime/         Runtime artifact ingestion (JUnit XML, Jest JSON)
├── scoring/         Explainable risk engine (reliability, change, speed)
├── signals/         Signal detector interface, registry, runner
├── summary/         Executive summary builder
├── testcase/        Test case extraction and identity collision detection
└── testtype/        Test type inference (unit, integration, e2e)
```

## Legacy Converter Engine

Hamlet originated as a multi-framework test converter (legacy, JavaScript ES modules), published to npm as `hamlet-testframework`. That engine is preserved in `src/`, `bin/`, and `test/` and remains functional. The current engine reframes migration as one dimension of broader test intelligence. See [docs/legacy/](docs/legacy/) for historical architecture docs and [CLAUDE.md](CLAUDE.md) for legacy code conventions.

## Development

```bash
# Build
go build -o hamlet ./cmd/hamlet

# Test all Go packages
go test ./internal/... ./cmd/...

# Test with verbose output
go test -v ./internal/...

# Legacy JavaScript tests (requires Node.js 22+)
npm test
```

## Principles

- Signals are the core abstraction
- Analysis comes before automation
- Risk must be explainable
- Hamlet must be useful locally, without SaaS
- Privacy boundary: aggregate metrics never expose raw paths or source code
- Hamlet measures system health, not individual developer productivity

## Status

Hamlet's current engine is in active development. The Go engine implements:
- repository analysis and signal detection
- explainable risk modeling
- local policy and governance
- ownership-aware review and triage
- migration intelligence
- snapshot history and trend comparison
- benchmark-ready metrics
- executive summary reporting

The JSON contract (`TestSuiteSnapshot`) is stabilizing but may evolve.

## License

MIT License — see [LICENSE](LICENSE) for details.
