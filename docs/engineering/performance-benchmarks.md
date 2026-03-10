# Performance Benchmarks

Hamlet includes Go benchmarks for all major computation pipelines. These
benchmarks track performance characteristics over time and guard against
regressions as the codebase grows.

## Go Benchmark Pattern

All benchmarks use Go's `testing.B` framework with `b.ResetTimer()` to exclude
fixture setup from the measurement:

```go
func BenchmarkMetrics_LargeScale(b *testing.B) {
    snap := LargeScaleSnapshot()   // setup: excluded from timing
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        metrics.Derive(snap)       // measured: N iterations
    }
}
```

The `b.N` loop is controlled by the test runner, which increases N until the
timing is statistically stable. This produces reliable ns/op measurements.

## Running Benchmarks

```bash
# All benchmarks in testdata
go test -bench . -benchmem ./internal/testdata/

# Specific benchmark
go test -bench BenchmarkMetrics_LargeScale -benchmem ./internal/testdata/

# With count for statistical stability
go test -bench . -benchmem -count=5 ./internal/testdata/

# Compare against a saved baseline (requires benchstat)
go test -bench . -benchmem -count=5 ./internal/testdata/ > new.txt
benchstat old.txt new.txt
```

The `-benchmem` flag reports allocations per operation (allocs/op and B/op),
which is often more actionable than raw timing for identifying regressions.

## Benchmark Inventory

All benchmarks live in `internal/testdata/bench_test.go`.

| Benchmark | Fixture | What It Measures |
|---|---|---|
| `BenchmarkMetrics_Minimal` | MinimalSnapshot | Baseline cost of metrics derivation |
| `BenchmarkMetrics_LargeScale` | LargeScaleSnapshot (550 files) | Metrics derivation at scale |
| `BenchmarkMeasurements_Minimal` | MinimalSnapshot | Baseline cost of measurement computation |
| `BenchmarkMeasurements_LargeScale` | LargeScaleSnapshot | Measurement computation at scale |
| `BenchmarkHeatmap_LargeScale` | LargeScaleSnapshot + risk | Heatmap construction including risk scoring |
| `BenchmarkRiskScoring_LargeScale` | LargeScaleSnapshot | Risk scoring at scale |
| `BenchmarkImpactAnalysis` | HealthyBalancedSnapshot + 3-file scope | Impact analysis for a typical PR |

### Planned Benchmarks

| Benchmark | Fixture | Purpose |
|---|---|---|
| `BenchmarkComparison_TwoSnapshots` | Two LargeScaleSnapshots | Comparison report generation between snapshot pairs |
| `BenchmarkPortfolio_LargeScale` | LargeScaleSnapshot | Portfolio analysis at scale |
| `BenchmarkFullPipeline_Minimal` | MinimalSnapshot | End-to-end pipeline: metrics + measurements + risk + heatmap + portfolio |
| `BenchmarkFullPipeline_VeryLarge` | VeryLargeSnapshot (2000 files) | Full pipeline under memory and compute pressure |

## Baseline Expectations

These are approximate baseline expectations, not hard thresholds. Actual
numbers depend on hardware, but relative relationships should hold:

| Tier | Example | Expected Order |
|---|---|---|
| Minimal fixtures (1-3 files) | BenchmarkMetrics_Minimal | < 100 us/op |
| Large scale (550 files) | BenchmarkMetrics_LargeScale | < 10 ms/op |
| Very large (2000 files) | Planned full pipeline | < 100 ms/op |

Allocations matter more than wall-clock time for Hamlet's use case (CLI tool,
CI integration). Excessive allocations cause GC pressure that manifests as
latency spikes in real usage.

## Interpreting Results

A typical benchmark run produces output like:

```
BenchmarkMetrics_Minimal-10         250000    4800 ns/op    3200 B/op    45 allocs/op
BenchmarkMetrics_LargeScale-10        2000  650000 ns/op  480000 B/op  5200 allocs/op
```

Key columns:
- **ns/op**: Nanoseconds per operation. Lower is better.
- **B/op**: Bytes allocated per operation. Lower means less GC pressure.
- **allocs/op**: Number of heap allocations per operation. Fewer is better.

## Regression Detection

To detect regressions:

1. Save a baseline: `go test -bench . -benchmem -count=5 ./internal/testdata/ > baseline.txt`
2. Make changes.
3. Run again: `go test -bench . -benchmem -count=5 ./internal/testdata/ > current.txt`
4. Compare: `benchstat baseline.txt current.txt`

`benchstat` reports the percentage change and statistical significance. A
regression of >20% in ns/op or >30% in allocs/op warrants investigation.

### CI Benchmark Comparison

For pull requests, CI runs a benchmark comparison job that:

1. Runs core Go benchmarks on the PR head.
2. Checks out the PR base commit and reruns the same benchmarks.
3. Compares results with `benchstat`.
4. Publishes the comparison in the GitHub Actions step summary.

This gives an immediate signal when performance drifts between base and head,
without requiring manual baseline capture for every PR.

## Adding New Benchmarks

When adding a benchmark:

1. Use an existing fixture factory. Do not construct fixtures inside the `b.N` loop.
2. Call `b.ResetTimer()` after all setup.
3. Name it `Benchmark<Subsystem>_<Fixture>` to match the existing convention.
4. Include both a minimal and large-scale variant if the subsystem is
   input-size-sensitive.
5. Run with `-benchmem` to capture allocation data from the start.
