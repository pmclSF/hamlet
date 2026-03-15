# Verification Strategy

## Philosophy

Terrain is a measurement and intelligence tool. Its outputs directly inform engineering decisions about test quality, migration readiness, risk posture, and resource allocation. Incorrect outputs erode trust faster than missing features.

The verification strategy prioritizes **confidence in critical invariants** over chasing a single global coverage number. Each subsystem has specific correctness properties, and each property is tested at the appropriate layer.

### Core Principles

1. **Determinism is non-negotiable.** Same inputs must produce identical outputs across runs, traversal orders, and execution modes. Determinism failures are P0 bugs.

2. **Schema contracts are versioned and tested.** Persisted artifacts (snapshots, exports, impact results) have explicit compatibility guarantees. Schema drift must be intentional and tested.

3. **Degradation must be honest.** When data is missing, malformed, or ambiguous, Terrain must produce reduced-confidence outputs — never silently wrong ones. Adversarial tests enforce this.

4. **Golden tests protect intentional behavior.** Machine-readable artifacts and user-facing outputs are snapshot-tested. Changes are reviewed, not accidental.

5. **Performance is a feature.** Scale regressions are caught with benchmarks, not discovered in production.

6. **Tests are documentation.** Well-structured tests communicate expected behavior to contributors. Fixture scenarios encode real-world patterns.

7. **Fixture-driven.** Tests use standardized fixtures representing real-world scenarios. Fixtures are reusable across test categories.

## What Must Be Deterministic

| Output | Why | Test layer |
|--------|-----|------------|
| Snapshots | Persisted, compared over time | Determinism suite |
| Measurement/posture results | Drive decisions, must be reproducible | Determinism suite |
| Risk scores and signal lists | Compared across runs | Determinism suite |
| Impact analysis results | Inform test selection | Determinism suite |
| Compare outputs | Diffs must be stable | Determinism suite |
| Benchmark-safe exports | Shared externally | Determinism suite |
| Test IDs | Identity system anchor | Unit tests |
| Code unit IDs | Impact graph anchor | Unit tests |
| Heatmap cells | Visualized, compared | Determinism suite |

## What Must Be Schema-Compatible

| Artifact | Compatibility guarantee | Test layer |
|----------|------------------------|------------|
| TestSuiteSnapshot | Forward-compatible (unknown fields ignored) | Schema suite |
| Benchmark exports | Versioned, backward-compatible within major | Schema suite |
| Impact aggregates | Additive fields only within minor version | Schema suite |
| MeasurementSnapshot | Forward-compatible | Schema suite |
| CoverageInsight | Additive fields only | Schema suite |
| PortfolioSnapshot | Forward-compatible | Schema suite |

## What Must Be Golden-Tested

| Output | Fixture | Format |
|--------|---------|--------|
| Metrics JSON | MinimalSnapshot | JSON |
| Benchmark export JSON | MinimalSnapshot | JSON |
| Analyze report | MinimalSnapshot | Text |
| Portfolio report | FlakyConcentratedSnapshot | Text |
| Summary report | MinimalSnapshot | Text |
| Posture report | MinimalSnapshot | Text |
| Impact report | HealthyBalancedSnapshot + change scope | Text |
| Compare report | FlakyConcentrated → HealthyBalanced | Text |
| Impact aggregate JSON | HealthyBalancedSnapshot + change scope | JSON |

Golden files are updated with `go test ./internal/testdata/ -update` and reviewed in PRs.

## What Must Be Performance-Regressed

| Scenario | Fixture | Measured |
|----------|---------|----------|
| Metrics derivation (small) | MinimalSnapshot | Runtime |
| Metrics derivation (large) | LargeScaleSnapshot | Runtime |
| Measurements (small/large) | Minimal / LargeScale | Runtime |
| Heatmap (large) | LargeScaleSnapshot | Runtime |
| Risk scoring (large) | LargeScaleSnapshot | Runtime |
| Impact analysis | HealthyBalancedSnapshot | Runtime |
| Full pipeline (very large) | VeryLargeSnapshot | Runtime |
| Comparison | FlakyConcentrated → HealthyBalanced | Runtime |
| Portfolio analysis | FlakyConcentratedSnapshot | Runtime |

Benchmarks run with `go test -bench . ./internal/testdata/` and are tracked over time.

## What Must Be Adversarially Tested

| Scenario | Risk | Expected behavior |
|----------|------|-------------------|
| Nil/empty snapshots | Panic, nil dereference | Graceful empty output |
| Zero test files | Division by zero | Zero-value metrics |
| Missing coverage data | Silent wrong output | Reduced confidence annotation |
| Malformed CODEOWNERS | Owner misattribution | Skip malformed entries |
| Files not in snapshot | Index out of bounds | File-level fallback, weak confidence |
| 1000+ signals | Performance degradation | Complete within threshold |
| Empty change scope | Unnecessary work | Empty result, no errors |
| Nonexistent owner filter | Silent empty result | Empty filtered set |
| Duplicate test IDs | Identity corruption | Deterministic dedup |
| No ownership data | Null dereference | Graceful skip, limitation noted |
| Deeply nested paths | Path normalization failure | Correct normalization |
| Mixed framework chaos | Detection confusion | All frameworks detected |

## Test Categories

| Category | Scope | Location | Purpose |
|----------|-------|----------|---------|
| Unit | Single function/type | `*_test.go` alongside source | Verify deterministic core logic |
| Golden | JSON/text output | `testdata/golden/` | Catch output regressions |
| Determinism | Pipeline reproducibility | `testdata/determinism_test.go` | Same input → same output |
| Schema | Persistence format | `testdata/schema_test.go` | Backward/forward compatibility |
| Adversarial | Edge cases | `testdata/adversarial_test.go` | Graceful degradation |
| Benchmark | Performance | `testdata/bench_test.go` | Detect performance regressions |
| E2E | Full pipeline flow | `testdata/e2e_test.go` | Flagship scenarios end-to-end |
| CLI | Command regression | `testdata/cli_test.go` | Output format, flags, exit codes |

## Fixture Architecture

```
internal/testdata/
  fixtures.go           — core scenario snapshot factories
  adversarial.go        — adversarial and edge-case snapshot factories
  golden/
    metrics-minimal.json     — expected metrics output
    export-minimal.json      — expected benchmark export
    analyze-minimal.txt      — expected analyze report
    portfolio-flaky.txt      — expected portfolio report
    summary-minimal.txt      — expected summary report
    posture-minimal.txt      — expected posture report
    impact-balanced.txt      — expected impact report
    compare-trend.txt        — expected compare report
    impact-aggregate.json    — expected impact aggregate
```

## Subsystem-to-Test-Layer Mapping

| Subsystem | Unit | Golden | Determinism | Schema | Adversarial | Benchmark | E2E |
|-----------|------|--------|-------------|--------|-------------|-----------|-----|
| Test identity | **required** | | **required** | | edge cases | | |
| Lifecycle tracking | **required** | | | | edge cases | | flow |
| Ownership | **required** | | | | malformed input | | flow |
| Coverage attribution | **required** | | | | partial data | | flow |
| Measurement/posture | **required** | **required** | **required** | **required** | empty/nil | **required** | flow |
| Portfolio intelligence | **required** | **required** | **required** | | edge cases | **required** | flow |
| Impact analysis | **required** | **required** | **required** | **required** | empty/missing | **required** | flow |
| CLI outputs | | **required** | | | bad flags | | **required** |
| Benchmark exports | | **required** | **required** | **required** | sparse data | | flow |
| Compare | | **required** | **required** | | nil snapshots | **required** | flow |
| Signals/scoring | **required** | | **required** | | zero/many | **required** | |
| Heatmap | **required** | | **required** | | no risk data | **required** | |

## Coverage Targets

Coverage targets are per-subsystem, not global. Focus on branch coverage for decision-heavy logic.

| Package | Target | Rationale |
|---------|--------|-----------|
| identity | 95%+ | Core correctness contract |
| models/validate | 90%+ | Schema boundary |
| measurement | 85%+ | Posture computation |
| impact | 85%+ | Selection and gap detection |
| scoring | 80%+ | Risk computation |
| lifecycle | 80%+ | Continuity inference |
| ownership | 80%+ | Precedence and propagation |
| coverage | 80%+ | Attribution logic |
| benchmark | 85%+ | Privacy-safe export |
| reporting | 70%+ | Rendering (lower priority) |

## Release Gate

A release candidate must pass all of:

1. `go vet ./cmd/... ./internal/...` — zero warnings
2. `go test ./internal/... ./cmd/...` — all tests green
3. `go build -o /dev/null ./cmd/terrain/` — clean build
4. Golden file tests match — no unexpected output changes
5. Determinism tests pass — 5-10 repeat runs identical
6. Schema tests pass — round-trip and forward compatibility
7. Benchmark tests within thresholds — no severe regressions
8. CLI regression tests pass — help text, commands, exit codes

See [quality-bar-and-gates.md](../release/quality-bar-and-gates.md) for the full gate definitions.

## Contributor Guidelines

When adding a new feature:

1. **Pure logic** (normalization, classification, computation) → add unit tests with edge cases
2. **New output format** → add golden test
3. **New persisted artifact** → add schema round-trip test
4. **New subsystem** → add adversarial test for nil/empty/partial inputs
5. **Performance-sensitive path** → add benchmark
6. **Cross-subsystem flow** → add E2E scenario test
7. **New CLI command** → add CLI regression test

See [test-architecture.md](test-architecture.md) for full layer definitions.
See [testing-and-quality.md](../contributing/testing-and-quality.md) for the contributor test guide.
