# Measurement Persistence

## Overview

Measurements and posture results are persisted into versioned snapshots as part of `TestSuiteSnapshot.Measurements`. This enables comparison, trend tracking, and benchmark-safe export across time.

## Schema

### MeasurementSnapshot

Embedded in `TestSuiteSnapshot` as the `Measurements` field:

```go
type MeasurementSnapshot struct {
    Posture      []DimensionPostureResult  `json:"posture"`
    Measurements []MeasurementResult        `json:"measurements"`
}
```

### DimensionPostureResult

```go
type DimensionPostureResult struct {
    Dimension           string              `json:"dimension"`
    Band                string              `json:"band"`
    Explanation         string              `json:"explanation"`
    DrivingMeasurements []string            `json:"drivingMeasurements,omitempty"`
    Measurements        []MeasurementResult `json:"measurements,omitempty"`
}
```

### MeasurementResult

```go
type MeasurementResult struct {
    ID          string   `json:"id"`
    Dimension   string   `json:"dimension"`
    Value       float64  `json:"value"`
    Units       string   `json:"units"`
    Band        string   `json:"band,omitempty"`
    Evidence    string   `json:"evidence"`
    Explanation string   `json:"explanation"`
    Inputs      []string `json:"inputs,omitempty"`
    Limitations []string `json:"limitations,omitempty"`
}
```

## Conversion

The measurement package uses its own internal types (`measurement.Result`, `measurement.DimensionPosture`), which are converted to model types via `Snapshot.ToModel()`:

```go
measSnap := registry.ComputeSnapshot(snapshot)
snapshot.Measurements = measSnap.ToModel()
```

This separation keeps the measurement package independent of the serialization model.

## Snapshot Storage

Snapshots are persisted to `.hamlet/snapshots/` when `--write-snapshot` is passed:

- `latest.json` — most recent snapshot (overwritten)
- `2026-03-08T12-00-00Z.json` — timestamped archive

Measurements are included in every snapshot automatically.

## Compare Support

The comparison system (`internal/comparison/compare.go`) produces:

- **PostureDeltas** — dimension band changes across snapshots
- **MeasurementDeltas** — individual measurement value and band changes

This enables trend tracking like "health went from strong to weak" or "flaky_share increased from 2% to 25%".

## Determinism

Persistence is fully deterministic:
- Measurements are stored in registry order
- Posture dimensions follow a fixed order
- All floating-point values are reproducible
- Driver lists are sorted alphabetically

## Compact Representation

The persisted form is compact by design:
- Only non-empty bands, inputs, and limitations are included (`omitempty`)
- Explanation strings are concise (one sentence)
- No raw signal data is duplicated — measurements reference signal types by name

## Version Safety

The `MeasurementSnapshot` is a struct with explicit fields. Adding new measurements to the registry automatically adds new entries to the `Measurements` array. Old snapshots that lack certain measurement IDs will produce `MeasurementDelta` entries with zero `Before` values when compared.

## File Map

```
internal/models/measurement.go       # Serializable types
internal/measurement/registry.go     # ToModel() conversion
internal/engine/pipeline.go          # Step 8: computation and persistence
internal/comparison/compare.go       # compareMeasurements() for deltas
```
