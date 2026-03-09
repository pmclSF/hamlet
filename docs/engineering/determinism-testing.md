# Determinism Testing

Determinism tests verify that Hamlet's computation pipeline produces bit-identical
output across multiple runs with the same input. Non-determinism in a test
intelligence platform erodes trust: if the same snapshot produces different risk
scores or posture bands on consecutive runs, users cannot rely on the results.

## The N-Run Identity Pattern

Every determinism test follows the same structure:

1. Construct a fixture snapshot.
2. Run the computation N times (typically 5 or 10).
3. Serialize each result to JSON.
4. Assert that all N JSON strings are identical to run 0.

```go
func TestDeterminism_MetricsIdentical(t *testing.T) {
    snap := HealthyBalancedSnapshot()
    results := make([]string, 10)
    for i := 0; i < 10; i++ {
        ms := metrics.Derive(snap)
        ms.GeneratedAt = FixedTime  // normalize time
        data, _ := json.Marshal(ms)
        results[i] = string(data)
    }
    for i := 1; i < 10; i++ {
        if results[i] != results[0] {
            t.Errorf("metrics run %d differs from run 0", i)
        }
    }
}
```

JSON serialization is the comparison mechanism because it captures field ordering,
floating-point representation, and nil-vs-empty distinctions. If two JSON strings
match, the underlying structs are functionally identical.

## Determinism Test Inventory

| Test Function | Runs | Fixture | What It Validates |
|---|---|---|---|
| `TestDeterminism_MetricsIdentical` | 10 | HealthyBalancedSnapshot | metrics.Derive output stability |
| `TestDeterminism_MeasurementsIdentical` | 10 | HealthyBalancedSnapshot | measurement.ComputeSnapshot and posture band stability |
| `TestDeterminism_HeatmapIdentical` | 10 | HealthyBalancedSnapshot | heatmap.Build output with risk scoring |
| `TestDeterminism_RiskScoringIdentical` | 10 | HealthyBalancedSnapshot | scoring.ComputeRisk output stability |
| `TestDeterminism_LargeScaleStable` | 5 | LargeScaleSnapshot | Measurement stability at 550 test files |

### Planned Determinism Tests

| Test Function | Runs | Fixture | Purpose |
|---|---|---|---|
| `TestDeterminism_ImpactIdentical` | 10 | HealthyBalancedSnapshot | impact.Analyze output for a fixed change scope |
| `TestDeterminism_ComparisonIdentical` | 10 | Two snapshots | Comparison report between snapshot pairs |
| `TestDeterminism_PortfolioIdentical` | 10 | FlakyConcentratedSnapshot | portfolio.Analyze output stability |

## Sources of Non-Determinism

### Timestamps

The primary source of non-determinism is `time.Now()`. All fixtures use `FixedTime`
(2025-01-15T12:00:00Z), and any `GeneratedAt` field populated during computation
is overwritten with `FixedTime` before serialization.

### Map Iteration Order

Go maps iterate in random order. Any computation that iterates over a map and
produces ordered output (JSON arrays, sorted lists) must sort explicitly. The
metrics, measurement, and heatmap packages all sort their output slices before
returning.

### Reordered-Traversal Approach

Some tests deliberately reorder the input (shuffling test file slices or ownership
map entries) to verify that the computation is order-independent. This strengthens
the determinism guarantee beyond "same input, same output" to "equivalent input,
same output."

### Floating-Point Precision

Measurement values are `float64`. The computation avoids accumulation-order
sensitivity by using stable summation patterns. JSON serialization of floats uses
Go's default formatting, which is deterministic for the same binary.

## Relationship to Golden Tests

Determinism tests and golden tests are complementary:

- **Determinism tests** verify internal consistency: N runs produce the same output.
  They do not assert what that output is.
- **Golden tests** verify external correctness: the output matches a stored
  reference file. They run once per test execution.

A subsystem can be deterministic but produce wrong output (determinism test passes,
golden test fails). Conversely, a golden test can pass on one run but the
subsystem might be non-deterministic (unlikely to catch with a single run).

## When to Add a Determinism Test

Add a determinism test when:
- A new computation pipeline is introduced (new `Compute*` or `Derive*` function)
- A subsystem starts using maps, goroutines, or channels internally
- A scale variant is needed (existing test uses small fixture, need large-scale)

The bar for determinism is absolute: any failure is a real bug that must be fixed,
not a flaky test to be retried.
