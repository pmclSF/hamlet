# Verification Strategy

## Principles

1. **Determinism first** — Given the same input, every analysis step produces identical output. No randomness, no time-dependent behavior in core logic.
2. **Fixture-driven** — Tests use standardized fixtures that represent real-world scenarios. Fixtures are versioned and reusable across test categories.
3. **Golden file testing** — Machine-readable outputs (JSON snapshots, exports, metrics) are compared against checked-in golden files. Changes to output format require explicit golden file updates.
4. **Layered coverage** — Unit tests for pure functions, integration tests for pipeline stages, E2E tests for CLI workflows.
5. **Adversarial inputs** — Tests include malformed, empty, and edge-case inputs to verify graceful degradation.

## Test categories

| Category | Scope | Location | Purpose |
|----------|-------|----------|---------|
| Unit | Single function/type | `*_test.go` alongside source | Verify deterministic core logic |
| Golden | JSON/text output | `testdata/golden/` | Catch output regressions |
| Determinism | Pipeline reproducibility | `internal/engine/*_test.go` | Same input → same output |
| Schema | Persistence format | `internal/models/*_test.go` | Backward/forward compatibility |
| Adversarial | Edge cases | `*_test.go` with adversarial fixtures | Graceful degradation |
| Benchmark | Performance | `*_bench_test.go` | Detect performance regressions |
| E2E | Full CLI workflow | `test/e2e/` | Flagship scenarios end-to-end |
| CLI | Command regression | `test/cli/` | Output format, flags, exit codes |

## Fixture architecture

```
internal/testdata/
  fixtures/
    minimal.json           — smallest valid snapshot
    healthy-balanced.json   — well-tested repo
    flaky-concentrated.json — repo with flaky, concentrated tests
    e2e-heavy.json          — repo relying on E2E tests
    migration-risk.json     — repo mid-migration
    empty.json              — empty snapshot (zero state)
    large-scale.json        — 500+ test files for scale testing
  golden/
    analyze-minimal.txt     — expected analyze output for minimal fixture
    metrics-minimal.json    — expected metrics output
    export-minimal.json     — expected benchmark export
    posture-minimal.txt     — expected posture output
    summary-minimal.txt     — expected summary output
```

## Coverage targets

| Package | Branch | Function | Line |
|---------|--------|----------|------|
| models | 80% | 90% | 90% |
| analysis | 70% | 80% | 80% |
| quality | 80% | 90% | 90% |
| measurement | 80% | 90% | 90% |
| impact | 80% | 90% | 90% |
| scoring | 70% | 80% | 80% |
| heatmap | 70% | 80% | 80% |
| reporting | 60% | 70% | 70% |
| engine | 70% | 80% | 80% |
| benchmark | 80% | 90% | 90% |

## Release gate

A release candidate must pass:
1. `go vet ./cmd/... ./internal/...` — zero warnings
2. `go test ./internal/...` — all tests pass
3. `go build -o /dev/null ./cmd/hamlet/` — clean build
4. Golden file tests match — no unexpected output changes
5. No test flakiness — determinism tests pass 10x in a row
