# Contributing: Adding a Measurement

This guide covers how to add a new measurement to Terrain's measurement framework.

## Prerequisites

- Understand the five posture dimensions ([Posture Dimensions](../engineering/posture-dimensions.md))
- Know which dimension your measurement feeds
- Identify the signals or data sources it will consume

## Step 1: Write the Compute Function

Add your compute function to the appropriate dimension file:

- `internal/measurement/health.go` — health measurements
- `internal/measurement/coverage.go` — coverage depth and diversity
- `internal/measurement/structural_risk.go` — structural risk
- `internal/measurement/operational_risk.go` — operational risk

```go
func computeMyNewMeasurement(snap *models.TestSuiteSnapshot) Result {
    total := len(snap.TestFiles)
    if total == 0 {
        return Result{
            ID: "dimension.my_new_measurement", Dimension: DimensionHealth,
            Value: 0, Units: UnitsRatio, Band: "strong",
            Evidence: EvidenceNone, Explanation: "No test files detected.",
        }
    }

    count := countSignals(snap, signals.SignalMySignalType)
    ratio := float64(count) / float64(total)
    band := ratioToBand(ratio, 0.05, 0.15, 0.30)

    return Result{
        ID: "dimension.my_new_measurement", Dimension: DimensionHealth,
        Value: ratio, Units: UnitsRatio, Band: band,
        Evidence:    EvidenceStrong,
        Explanation: fmt.Sprintf("%d of %d test file(s) affected (%.0f%%).", count, total, ratio*100),
        Inputs:      []string{"mySignalType"},
    }
}
```

### Conventions

- **ID format:** `dimension.snake_case_name` (e.g., `health.flaky_share`)
- **Empty snapshot:** Return value 0, band "strong", evidence "none"
- **Explanation:** One sentence describing what was measured and the result
- **Inputs:** List the signal types or data sources consumed
- **Limitations:** Add strings describing data gaps when evidence is partial/weak

### Helper Functions

| Helper | Usage |
|--------|-------|
| `countSignals(snap, types...)` | Count signals matching any given type |
| `ratioToBand(ratio, low, mid, high)` | Map ratio to strong/moderate/weak/critical |
| `runtimeEvidence(snap)` | Check if runtime data is available |
| `evidenceLimitations(evidence)` | Standard limitation strings for evidence levels |

## Step 2: Add the Definition

Add your `Definition` to the dimension's measurement list function:

```go
func HealthMeasurements() []Definition {
    return []Definition{
        // ... existing measurements ...
        {
            ID:          "health.my_new_measurement",
            Dimension:   DimensionHealth,
            Description: "Short description of what this measures.",
            Units:       UnitsRatio,
            Inputs:      []string{"mySignalType"},
            Compute:     computeMyNewMeasurement,
        },
    }
}
```

The `DefaultRegistry()` will pick it up automatically.

## Step 3: Add Tests

Add tests in `internal/measurement/measurement_test.go`:

```go
func TestMyNewMeasurement_Empty(t *testing.T) {
    snap := makeSnap()
    r := DefaultRegistry()
    results := r.RunDimension(snap, DimensionHealth)
    // Find your measurement and verify empty-snapshot behavior
}

func TestMyNewMeasurement_Normal(t *testing.T) {
    snap := makeSnap(
        sig(signals.SignalMyType),
        sig(signals.SignalMyType),
    )
    snap.TestFiles = make([]models.TestFile, 10)
    r := DefaultRegistry()
    results := r.RunDimension(snap, DimensionHealth)
    // Verify value, band, evidence, explanation
}
```

Test coverage expectations:
- **Empty snapshot** — zero value, strong band, no-data evidence
- **Normal case** — correct ratio, band, evidence
- **Threshold boundaries** — verify band transitions
- **Evidence levels** — verify evidence degrades when data is missing

## Step 4: Verify

```bash
go test ./internal/measurement/ -v -run TestMyNew
go test ./internal/... -count=1
go build ./internal/... ./cmd/...
```

## Step 5: Document

Add a section to the appropriate measurement doc:
- `docs/engineering/health-measurements.md`
- `docs/engineering/coverage-depth-measurements.md`
- `docs/engineering/coverage-diversity-measurements.md`
- `docs/engineering/structural-risk-measurements.md`
- `docs/engineering/operational-risk-measurements.md`

Include: what, how, evidence, thresholds, why it matters, limitations.

## Checklist

- [ ] Compute function follows conventions (ID format, empty snapshot, explanation)
- [ ] Definition added to dimension's measurement list
- [ ] Tests cover empty, normal, threshold, and evidence cases
- [ ] No duplicate measurement ID (registry panics on duplicates)
- [ ] Documentation updated
- [ ] `go test ./internal/measurement/` passes
- [ ] `go build ./internal/... ./cmd/...` passes
