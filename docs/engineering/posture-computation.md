# Posture Computation

## Overview

Posture computation turns a set of measurement results for a dimension into a single calibrated band with an explanation. It runs per-dimension, producing a `DimensionPosture` for each of Terrain's five dimensions.

## Posture Bands

| Band | Meaning |
|------|---------|
| `strong` | No significant issues detected |
| `moderate` | Minor issues; room for improvement |
| `weak` | Significant issues requiring attention |
| `elevated` | Widespread problems affecting reliability |
| `critical` | Immediate attention needed |
| `unknown` | No measurements available |

## Algorithm

The posture computation in `computeDimensionPosture()` follows these steps:

### Step 1: Collect Bands

Each measurement result carries a `Band` field (strong/moderate/weak/elevated/critical). The algorithm collects all non-empty bands from the dimension's measurements.

### Step 2: Identify Worst Band

Band precedence (highest to lowest severity):

```
critical (5) > elevated (4) > weak (3) > moderate (2) > strong (1)
```

The worst band among all measurements becomes the candidate posture.

### Step 3: Majority Escalation

If more than half of the measurements have band >= weak, the posture is escalated to at least `weak`. This prevents a single strong measurement from masking widespread moderate issues.

### Step 4: Evidence Cap

If no measurement has `strong` or `partial` evidence, the posture is capped at `moderate`. This prevents false confidence when evidence is limited.

### Step 5: Identify Drivers

Measurements with band `weak`, `elevated`, or `critical` are flagged as "driving measurements" — these are the specific measurements that explain why the posture is what it is.

### Step 6: Generate Explanation

A human-readable explanation is generated based on the band, drivers, and measurement count. Examples:

- "health posture is strong across 4 measurement(s)."
- "coverage_depth posture is weak. Driven by: coverage_depth.uncovered_exports."
- "structural_risk posture is critical. Immediate attention needed."

## Example Walkthrough

Given health measurements:
- `health.flaky_share`: 0.25 (weak)
- `health.skip_density`: 0.03 (strong)
- `health.dead_test_share`: 0.0 (strong)
- `health.slow_test_share`: 0.12 (moderate)

1. **Bands:** [weak, strong, strong, moderate]
2. **Worst:** weak (3)
3. **Majority:** 1/4 >= weak — no escalation
4. **Evidence:** strong evidence present — no cap
5. **Drivers:** ["health.flaky_share"]
6. **Result:** Band = weak, Explanation = "health posture is weak. Driven by: health.flaky_share."

## Calibration

### No Global Score

Terrain does not produce a single global repo score. Each dimension has its own posture. This is intentional — collapsing five dimensions into one number would lose the actionability that makes posture useful.

### Threshold Transparency

Each measurement uses explicit thresholds via `ratioToBand()`:

```go
func ratioToBand(ratio, low, mid, high float64) string
```

| Ratio range | Band |
|-------------|------|
| `<= low` | strong |
| `<= mid` | moderate |
| `<= high` | weak |
| `> high` | critical |

Thresholds are defined per-measurement and documented in the measurement's compute function.

### Stability

The posture computation is deterministic:
- Same inputs always produce the same output
- Dimension order is fixed
- Driver lists are sorted alphabetically
- No randomness or sampling

## Testing

Posture computation is tested in `internal/measurement/measurement_test.go`:

- `TestPostureStrong` — all clean measurements produce strong
- `TestPostureWeak` — weak measurements drive weak posture
- `TestResolvePostureBand` — band resolution edge cases
- `TestFullPipeline_EndToEnd` — integration test from snapshot through posture

## File Map

```
internal/measurement/registry.go  # computeDimensionPosture(), resolvePostureBand(), buildPostureExplanation()
```
