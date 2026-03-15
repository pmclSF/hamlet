# Measurement Explainability

## Overview

Terrain's measurement and posture model is designed to be inspectable and debuggable. Every measurement carries explanation metadata, and every posture band can be traced to specific driving measurements.

## Explainability Structures

### Per-Measurement

Each `MeasurementResult` includes:

| Field | Purpose |
|-------|---------|
| `Explanation` | Human-readable summary of what was measured and the result |
| `Evidence` | How strong the backing data is (strong/partial/weak/none) |
| `Inputs` | Signal types or data sources that fed this measurement |
| `Limitations` | Specific data gaps or caveats |
| `Band` | Qualitative interpretation of the value |

Example:
```json
{
  "id": "health.flaky_share",
  "value": 0.25,
  "band": "weak",
  "evidence": "weak",
  "explanation": "5 of 20 test file(s) flagged as flaky or unstable (25%).",
  "inputs": ["flakyTest", "unstableSuite"],
  "limitations": ["No runtime data available; result is based on static analysis only."]
}
```

### Per-Dimension

Each `DimensionPostureResult` includes:

| Field | Purpose |
|-------|---------|
| `Band` | The overall posture band for this dimension |
| `Explanation` | Why this band was assigned |
| `DrivingMeasurements` | Which measurement IDs most influenced the band |
| `Measurements` | All individual measurements for drill-down |

Example:
```json
{
  "dimension": "health",
  "band": "weak",
  "explanation": "health posture is weak. Driven by: health.flaky_share.",
  "drivingMeasurements": ["health.flaky_share"],
  "measurements": [...]
}
```

## CLI Access

### terrain posture

The `terrain posture` command renders a full posture breakdown with measurement evidence:

```
HEALTH
  Posture: WEAK
  health posture is weak. Driven by: health.flaky_share.

  Driving measurements: health.flaky_share
  Measurements:
    health.flaky_share                       25.0% [weak]
      Evidence: weak
      5 of 20 test file(s) flagged as flaky or unstable (25%).
      * No runtime data available; result is based on static analysis only.
    health.skip_density                      3.0% [strong]
      Evidence: strong
      ...
```

### terrain posture --json

Returns the full `MeasurementSnapshot` as JSON, suitable for programmatic consumption.

### terrain compare

Shows posture and measurement changes across snapshots:

```
Posture Changes
----------------------------------------
  health                     STRONG → WEAK

Measurement Changes
----------------------------------------
  health.flaky_share                     +23.0%
    band: strong → weak
```

## What Would Improve the Posture?

Users can inspect the driving measurements to understand what to fix:

1. Look at `drivingMeasurements` — these are the measurements pulling the band down
2. Read each measurement's `explanation` — it says what was measured
3. Check `inputs` — these are the signal types to investigate
4. Use `terrain analyze` to find specific instances of those signals

Example flow:
- "Health is weak" → driven by `health.flaky_share`
- `health.flaky_share` = 25% → 5 of 20 files flagged as flaky
- Input signals: `flakyTest`, `unstableSuite`
- `terrain analyze --json | jq '.signals[] | select(.type == "flakyTest")'` → see specific files

## Missing Data Transparency

When data is missing, measurements are explicit:

- **evidence: none** — "No test files detected." / "No coverage data available."
- **evidence: weak** — "No runtime data available; result is based on static analysis only."
- **Limitations** — specific strings describing what's missing and how to fix it

This prevents false confidence. A "strong" posture with evidence "none" is different from a "strong" posture with evidence "strong".

## Privacy Boundary

Explanations in benchmark exports are stripped. The export includes only:
- Posture bands (by dimension)
- Aggregate metrics (counts, ratios)
- No raw measurement explanations, file paths, or signal details

## File Map

```
internal/measurement/measurement.go     # Result type with explanation fields
internal/measurement/registry.go        # buildPostureExplanation()
internal/reporting/posture_report.go    # RenderPostureReport() — the "explain" view
internal/reporting/comparison_report.go # Posture/measurement delta rendering
```
