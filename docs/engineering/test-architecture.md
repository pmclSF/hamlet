# Test Architecture

Hamlet's V3 engine (Go) uses a layered test strategy. Each layer targets a different failure class, from pure logic bugs through schema drift to performance regressions. All test infrastructure lives under `internal/testdata/`.

## File Organization

```
internal/testdata/
  fixtures.go          # Reusable snapshot factories (EmptySnapshot … LargeScaleSnapshot)
  adversarial.go       # Adversarial/edge-case factories (MixedFramework … VeryLargeSnapshot)
  golden/              # Snapshot-tested reference outputs
    metrics-minimal.json
    export-minimal.json
    analyze-minimal.txt
    summary-minimal.txt
    posture-minimal.txt
    portfolio-flaky.txt
  golden_test.go       # Golden output assertions (-update flag to refresh)
  determinism_test.go  # N-run identity checks
  bench_test.go        # Go benchmarks (Benchmark*)
  adversarial_test.go  # Nil/empty/malformed input tests
  schema_test.go       # JSON round-trip and forward compatibility
  e2e_test.go          # Full pipeline scenario tests
  cli_test.go          # Binary compilation and command regression

internal/<subsystem>/
  *_test.go            # Unit tests colocated with source
```

## Test Layers

### 1. Unit Tests

Colocated `*_test.go` files alongside source. Table-driven where inputs/outputs are enumerable. Covers identity resolution, lifecycle classification, ownership resolution/propagation/aggregation, coverage ingestion/attribution/insights, measurement computation, scoring, impact selection/filtering, heatmap construction, comparison diffing, portfolio analysis, signal detection, model validation/sorting, graph building, migration readiness, policy evaluation, and CLI shorthands.

Run: `go test ./internal/...`

### 2. Integration/Fixture Tests

Tests that compose multiple subsystems using standardized factories from `internal/testdata/fixtures.go` and `adversarial.go`. Each factory returns a `*models.TestSuiteSnapshot` representing a specific real-world scenario:

| Factory | Scenario |
|---------|----------|
| `EmptySnapshot` | Valid but completely empty repo |
| `MinimalSnapshot` | One framework, one test file, one code unit |
| `HealthyBalancedSnapshot` | Well-tested repo with strong coverage, multiple teams |
| `FlakyConcentratedSnapshot` | Flaky E2E tests, concentrated ownership |
| `E2EHeavySnapshot` | Over-reliance on E2E with shallow unit tests |
| `MigrationRiskSnapshot` | Mid-migration repo with dual frameworks and coverage gaps |
| `LargeScaleSnapshot` | 550 test files for scale testing |
| `MixedFrameworkSnapshot` | 5+ frameworks across 4 languages |
| `ZeroSignalSnapshot` | Clean repo that should produce zero quality signals |
| `AllSignalTypesSnapshot` | One signal per category/severity combination |
| `DeepNestingSnapshot` | Deeply nested directory structures (10 levels) |
| `VeryLargeSnapshot` | 2000+ test files for memory pressure testing |

All factories use `FixedTime` (2025-01-15T12:00:00Z) for deterministic timestamps.

### 3. Golden Output Tests

`internal/testdata/golden_test.go` compares rendered output byte-for-byte against reference files in `golden/`. Covers: metrics JSON, export JSON, analyze text, portfolio text, summary text, posture text.

Update golden files: `go test ./internal/testdata/ -run TestGolden -update`

Golden tests pin the output contract. Any change to rendering or computation that alters output must update the golden files explicitly.

### 4. Schema Compatibility Tests

`internal/testdata/schema_test.go` validates serialization contracts:

- **Round-trip**: Marshal a `TestSuiteSnapshot`, unmarshal it, verify field equality (frameworks, test files, code units, ownership, measurements).
- **Forward compatibility**: Unknown JSON fields are silently ignored -- ensures older consumers can read newer snapshots.
- **Empty snapshot**: Verifies zero-value snapshots serialize and deserialize cleanly.
- **Measurement snapshot**: Validates posture dimensions and bands survive round-trip.

### 5. Determinism Tests

`internal/testdata/determinism_test.go` runs each computation N times (10 for core paths, 5 for large-scale) with identical input and asserts JSON-identical output across all runs. Covers:

- `metrics.Derive` (10 runs)
- `measurement.ComputeSnapshot` (10 runs)
- `heatmap.Build` (10 runs)
- `scoring.ComputeRisk` (10 runs)
- Large-scale measurements (5 runs on 550-file snapshot)

Any map iteration non-determinism, timestamp leak, or random seed issue surfaces here.

### 6. Performance/Scale Benchmarks

`internal/testdata/bench_test.go` provides Go benchmarks for hot paths:

| Benchmark | Fixture |
|-----------|---------|
| `BenchmarkMetrics_Minimal` | MinimalSnapshot |
| `BenchmarkMetrics_LargeScale` | LargeScaleSnapshot |
| `BenchmarkMeasurements_Minimal` | MinimalSnapshot |
| `BenchmarkMeasurements_LargeScale` | LargeScaleSnapshot |
| `BenchmarkHeatmap_LargeScale` | LargeScaleSnapshot |
| `BenchmarkRiskScoring_LargeScale` | LargeScaleSnapshot |
| `BenchmarkImpactAnalysis` | HealthyBalancedSnapshot |

Run: `go test -bench=. ./internal/testdata/`

### 7. Adversarial Tests

`internal/testdata/adversarial_test.go` and `internal/engine/adversarial_test.go` feed nil, empty, malformed, and partial inputs into every subsystem:

- Nil measurements with summary rendering
- Empty signal list with risk scoring
- Zero test files with metrics derivation
- Empty snapshot with measurement computation
- Empty change scope with impact analysis
- Nonexistent file paths with impact analysis (expects weak confidence)
- Heatmap with nil risk data
- 1000-signal volume with risk scoring
- Nonexistent owner filter (expects zero results)
- Mixed frameworks, zero-signal repos, and all-signal-type combos through the full pipeline

### 8. E2E Scenario Tests

`internal/testdata/e2e_test.go` exercises full pipeline flows end-to-end:

- **Analyze-to-summary**: risk -> measurements -> heatmap -> metrics -> export -> executive summary -> render (analyze, posture, summary reports)
- **Comparison workflow**: two snapshots -> `comparison.Compare` -> render comparison report
- **Impact workflow**: change scope -> `impact.Analyze` -> render impact report, units drill-down, gaps drill-down -> build aggregate -> serialize
- **Migration flow**: migration-risk snapshot -> risk + measurements -> verify multi-framework detection, posture dimensions, area assessments, coverage guidance
- **Graph-enriched heatmap**: risk -> graph build -> `heatmap.BuildWithGraph` -> verify owner hotspots, E2E-only units, owner risk summaries, module coverage summaries
- **Review with coverage/identity**: signals + coverage summary -> render review sections -> verify coverage-by-type and E2E-only data
- **Ownership propagation**: resolve -> set owners -> signals -> health summaries -> quality summaries -> focus items -> benchmark export (verify owner names absent)
- **Ownership comparison trend**: before/after snapshots with different ownership maps -> verify delta and improvement flag
- **Portfolio intelligence**: flaky snapshot -> portfolio analyze -> verify low-value/high-cost findings, posture band, rendered report, export privacy
- **Export privacy**: verify benchmark export contains no raw file paths from fixtures

### 9. CLI Regression Tests

`internal/testdata/cli_test.go` runs against the compiled binary:

- `go build` succeeds for `cmd/hamlet/`
- `--help` exits cleanly and mentions "Hamlet"
- Help text lists all commands: analyze, summary, posture, metrics, compare, impact, policy check, export benchmark
- Unknown command exits non-zero
- `analyze --root internal/analysis/testdata/sample-repo` produces expected header and next-steps
- `analyze --json` produces valid JSON output
- `posture`, `metrics`, `summary` commands each produce expected headers

### 10. UI/View-Model Tests

`internal/summary/executive_test.go` and `internal/reporting/*_test.go` test view-model transformations:

- Executive summary `Build()` produces correct headline, risk band, top hotspots, and recommended actions from snapshot + heatmap + metrics inputs
- Report renderers produce correct section headers, badges, and data tables for each report type (analyze, summary, posture, comparison, impact, portfolio, review, migration)
- State handling: nil/empty measurements, zero signals, missing ownership, missing risk data all produce valid (possibly minimal) output
- Confidence cues: impact reports surface confidence levels (weak/moderate/strong) per impacted unit

### 11. Release-Gate Checks

Before release, all of the following must pass:

- `go test ./...` -- all unit, integration, golden, schema, determinism, adversarial, E2E, and CLI tests
- `go test -bench=. ./internal/testdata/` -- benchmarks must not regress (manual review)
- `go vet ./...` -- static analysis
- Golden files must be up to date (no `-update` needed)
- CLI tests verify binary compilation and all commands respond correctly

## Subsystem-to-Test-Layer Mapping

| Subsystem | Unit | Fixture | Golden | Schema | Determinism | Bench | Adversarial | E2E | CLI |
|-----------|:----:|:-------:|:------:|:------:|:-----------:|:-----:|:-----------:|:---:|:---:|
| metrics | x | x | x | | x | x | x | x | x |
| measurement | x | x | | x | x | x | x | x | |
| scoring | x | x | | | x | x | x | x | |
| heatmap | x | x | | | x | x | x | x | |
| impact | x | x | | | | x | x | x | |
| comparison | x | x | | | | | | x | |
| portfolio | x | x | x | | | | | x | |
| reporting | x | x | x | | | | x | x | |
| summary | x | x | x | | | | | x | x |
| ownership | x | x | | | | | | x | |
| migration | x | x | | | | | | x | |
| coverage | x | | | | | | | | |
| identity | x | | | | | | | | |
| lifecycle | x | | | | | | | | |
| graph | x | | | | | | | x | |
| signals | x | | | | | | | | |
| engine | x | x | | | | | x | | |
| models | x | | | x | | | | | |
| policy | x | | | | | | | | |
| benchmark | x | x | x | | | | | x | |
| CLI | | | | | | | | | x |
