# Contributing: Measurement and Posture System

This guide covers how to maintain and extend the measurement and posture system.

## Architecture Overview

```
Signals → Measurements → Posture → User-Facing Outputs
                                    ├── hamlet posture
                                    ├── hamlet summary
                                    ├── hamlet compare
                                    └── hamlet export benchmark
```

## Key Design Decisions

### 1. No single global score

Hamlet reports five independent dimensions. Do not add a combined score. If someone asks for one, explain why it would reduce actionability.

### 2. Evidence is first-class

Every measurement must carry evidence strength. Never produce a measurement without declaring how confident the data is. Missing data should degrade evidence, not produce a fake value.

### 3. Bands over decimals

Prefer qualitative bands (strong/moderate/weak/elevated/critical) over decimal scores. A band of "weak" is more actionable than a score of "0.37". Raw values are available for programmatic use but the band is the primary user-facing interpretation.

### 4. Stability over novelty

Measurement IDs and semantics must be stable across versions. Do not rename or redefine a measurement without a migration plan. Old snapshots must remain comparable to new ones.

### 5. Dimension boundaries are clear

Each measurement feeds exactly one dimension. Each dimension has explicit scope and exclusions. If a concern spans two dimensions, document the overlap rather than creating a cross-cutting measurement.

## Common Tasks

### Adding a measurement

See [Adding a Measurement](adding-a-measurement.md).

### Changing posture thresholds

Thresholds are defined per-measurement in the compute function via `ratioToBand()`. To change a threshold:

1. Modify the threshold arguments in the compute function
2. Update the corresponding documentation
3. Add a test verifying the new boundary
4. Consider whether existing snapshots will show different posture (this is acceptable but should be noted)

### Modifying posture computation

The posture computation algorithm is in `internal/measurement/registry.go`:

- `computeDimensionPosture()` — main algorithm
- `resolvePostureBand()` — worst-band resolution
- `buildPostureExplanation()` — explanation generation

Changes here affect all dimensions. Test thoroughly with `TestPostureStrong`, `TestPostureWeak`, and `TestResolvePostureBand`.

### Adding a dimension

1. Add a `Dimension` constant in `measurement.go`
2. Create a new file for the dimension's measurements
3. Add a measurement list function
4. Register in `DefaultRegistry()`
5. Add to the `dims` slice in `ComputeSnapshot()`
6. Update summary, comparison, and reporting to handle the new dimension
7. Update `docs/engineering/posture-dimensions.md`

Adding a dimension is a significant change. Keep the dimension count intentionally small.

## Testing Expectations

All measurement code is tested in `internal/measurement/measurement_test.go`. When modifying:

- Run all measurement tests: `go test ./internal/measurement/ -v`
- Run comparison tests: `go test ./internal/comparison/ -v`
- Run the full suite: `go test ./internal/... -count=1`
- Build everything: `go build ./internal/... ./cmd/...`

### Golden test

The golden test fixture at `internal/testdata/golden/posture-minimal.txt` verifies the posture report output format. If you change the report format, update the golden file.

## Privacy Rules

The benchmark export (`internal/benchmark/export.go`) must never include:
- Raw measurement explanations
- File paths or code unit names
- Signal details

Only dimension → band mappings and aggregate metrics are safe for export.

## Phrasing

Follow Hamlet's ownership-safe phrasing:

| Do | Don't |
|----|-------|
| "Coverage depth is weak" | "Coverage is bad" |
| "Driven by uncovered exports" | "Nobody tested these" |
| "Health posture is critical" | "Tests are broken" |
| "Structural risk is elevated in auth area" | "Auth team has too much tech debt" |

## File Map

```
internal/measurement/
├── measurement.go        # Core types
├── registry.go           # Registry, posture computation
├── default_registry.go   # Default setup
├── health.go             # Health measurements
├── coverage.go           # Coverage depth + diversity
├── structural_risk.go    # Structural risk
├── operational_risk.go   # Operational risk
├── helpers.go            # Shared utilities
└── measurement_test.go   # Tests

internal/models/measurement.go       # Serializable types
internal/engine/pipeline.go          # Pipeline integration (Step 8)
internal/comparison/compare.go       # Measurement deltas
internal/reporting/posture_report.go # Posture report rendering
internal/benchmark/export.go         # Privacy-safe export
```

## Related Documentation

- [Measurement Framework](../engineering/measurement-framework.md) — architecture overview
- [Posture Dimensions](../engineering/posture-dimensions.md) — dimension contracts
- [Posture Computation](../engineering/posture-computation.md) — band derivation algorithm
- [Measurement Registry](../engineering/measurement-registry.md) — registration system
- [Adding a Measurement](adding-a-measurement.md) — step-by-step guide
