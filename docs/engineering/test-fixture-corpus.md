# Test Fixture Corpus

Hamlet's test suite uses code-based fixture factories rather than static JSON files.
All fixtures live in `internal/testdata/` across two source files: `fixtures.go`
(core scenarios) and `adversarial.go` (edge-case and adversarial scenarios).

## Design Principles

- **Realistic scenarios.** Each fixture models a pattern observed in real-world
  repositories: healthy codebases, migration-in-progress repos, flaky E2E suites,
  polyglot monorepos, and scale extremes.
- **Deterministic via FixedTime.** Every factory sets `GeneratedAt` to
  `FixedTime` (2025-01-15T12:00:00Z), eliminating time-dependent variance in
  golden files, determinism tests, and benchmark baselines.
- **Code-based factories.** Factories are Go functions returning
  `*models.TestSuiteSnapshot`. This makes fixtures type-safe, composable, and
  impossible to drift from the schema. No JSON fixture files to maintain.
- **Single-concern snapshots.** Each fixture isolates one risk pattern or
  structural characteristic so tests can assert on specific behaviors without
  confounding variables.

## Fixture Manifest

| Factory | File | What It Represents | Key Subsystems Exercised | Expected Behavior |
|---|---|---|---|---|
| `EmptySnapshot` | fixtures.go | Valid but completely empty repo | Metrics, reporting, schema | Zero counts, no panics, graceful empty output |
| `MinimalSnapshot` | fixtures.go | One framework, one test file, one code unit | Metrics, golden, schema, measurement | Baseline non-zero metrics, clean golden output |
| `HealthyBalancedSnapshot` | fixtures.go | Well-tested repo with unit + E2E coverage and ownership | Metrics, heatmap, risk, determinism, schema | Low risk scores, strong posture, balanced heatmap |
| `FlakyConcentratedSnapshot` | fixtures.go | Flaky E2E tests with low pass rates and high retry rates | Risk scoring, portfolio, golden | High risk signals, flaky test detection, concentrated ownership warnings |
| `E2EHeavySnapshot` | fixtures.go | Over-reliance on E2E with shallow unit tests | Migration readiness, risk, portfolio | E2E-heavy signals, migration recommendations |
| `MigrationRiskSnapshot` | fixtures.go | Mid-migration from Jasmine/Protractor to Jest with coverage data | Migration, quality signals, coverage | Dual-framework warnings, migration blockers, coverage insights |
| `LargeScaleSnapshot` | fixtures.go | 550 test files across 3 frameworks with 200 code units | Performance benchmarks, determinism at scale | Sub-second metrics derivation, stable deterministic output |
| `MixedFrameworkSnapshot` | adversarial.go | 4 languages, 6 frameworks including Go and Java | Framework detection, migration readiness | Polyglot handling, cross-language signal generation |
| `ZeroSignalSnapshot` | adversarial.go | Perfect repo that should produce zero quality signals | Signal generation, scoring | Empty signal list, clean posture |
| `AllSignalTypesSnapshot` | adversarial.go | Pre-loaded with one signal per category/severity combination (25 total) | Renderers, filters, signal aggregation | All categories rendered, severity filtering works |
| `DeepNestingSnapshot` | adversarial.go | 10 test files at directory depths 1 through 10 | Path handling, heatmap grouping | No path truncation bugs, correct grouping |
| `VeryLargeSnapshot` | adversarial.go | 2000 test files, 500 code units across 3 languages | Memory pressure, benchmark ceiling | No OOM, acceptable benchmark times |

## Planned Additions

### SuppressionHeavySnapshot

A repo where a significant fraction of tests are suppressed (skipped, quarantined,
or annotated with `@disable`). This exercises the suppression detection pipeline,
quarantine reporting, and ensures posture correctly penalizes heavy suppression
rather than treating skipped tests as invisible.

Key characteristics:
- 40+ test files, 30% marked as suppressed via various mechanisms
- Mixed suppression reasons (flaky, blocked, deprecated)
- Ownership data to test per-team suppression rates

### OwnershipFragmentedSnapshot

A repo where CODEOWNERS mappings are fragmented: many code units have no owner,
some have multiple overlapping owners, and ownership boundaries do not align with
directory structure. This exercises ownership attribution, governance signals,
and the orphan-code-unit detection path.

Key characteristics:
- 20+ code units with partial or absent ownership
- Overlapping owner rules that create ambiguity
- Test files that span multiple ownership boundaries

## Usage in Tests

Fixtures are used across all test categories:

```go
// Golden tests
snap := MinimalSnapshot()
ms := metrics.Derive(snap)
assertGolden(t, "metrics-minimal.json", data)

// Determinism tests
snap := HealthyBalancedSnapshot()
for i := 0; i < 10; i++ {
    ms := metrics.Derive(snap)
    // compare JSON identity across runs
}

// Adversarial tests
snap := &models.TestSuiteSnapshot{Signals: []models.Signal{}}
risks := scoring.ComputeRisk(snap)

// Benchmarks
snap := LargeScaleSnapshot()
b.ResetTimer()
for i := 0; i < b.N; i++ {
    metrics.Derive(snap)
}
```

When adding a new fixture, place it in the appropriate file (`fixtures.go` for
standard scenarios, `adversarial.go` for edge cases), follow the existing naming
convention (`<Adjective><Noun>Snapshot`), and always set `GeneratedAt: FixedTime`.
