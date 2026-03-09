# Measurement Registry

## Overview

The measurement registry is a structured container for measurement definitions. It provides registration, grouping, introspection, and execution of measurements against a snapshot.

## Design

The registry follows a simple pattern:

1. **Register** measurement definitions at startup
2. **Run** all or per-dimension measurements against a snapshot
3. **ComputeSnapshot** runs everything and derives posture bands

```go
r := measurement.NewRegistry()
r.Register(def)            // Add a definition (panics on duplicate ID)
r.All()                    // All definitions in registration order
r.ByDimension(dim)         // Definitions for one dimension
r.Len()                    // Count of registered definitions
r.Run(snap)                // Execute all, return []Result
r.RunDimension(snap, dim)  // Execute one dimension
r.ComputeSnapshot(snap)    // Full computation with posture
```

## Default Registry

`DefaultRegistry()` returns a registry pre-populated with all 18 standard measurements:

| Dimension | Count | Source |
|-----------|-------|--------|
| health | 4 | `HealthMeasurements()` |
| coverage_depth | 3 | `CoverageDepthMeasurements()` |
| coverage_diversity | 5 | `CoverageDiversityMeasurements()` |
| structural_risk | 3 | `StructuralRiskMeasurements()` |
| operational_risk | 3 | `OperationalRiskMeasurements()` |

## Definition Structure

```go
type Definition struct {
    ID          string                                     // Stable identifier
    Dimension   Dimension                                  // Posture dimension
    Description string                                     // Human-readable summary
    Units       Units                                      // Output type (ratio, count, etc.)
    Inputs      []string                                   // Signal types or data sources
    Compute     func(snap *models.TestSuiteSnapshot) Result // Computation function
}
```

## Duplicate Detection

The registry panics on duplicate IDs to catch registration errors at startup rather than at runtime:

```go
r.Register(Definition{ID: "health.flaky_share", ...})
r.Register(Definition{ID: "health.flaky_share", ...}) // panics
```

## Introspection

The registry supports programmatic inspection:

```go
for _, def := range r.All() {
    fmt.Printf("%-40s %s (%s)\n", def.ID, def.Description, def.Units)
}
```

Per-dimension grouping:

```go
healthDefs := r.ByDimension(measurement.DimensionHealth)
```

## Determinism

- Definitions are stored in registration order
- `Run()` executes in registration order
- Results are returned in registration order
- Posture dimensions are computed in a fixed order: health, coverage_depth, coverage_diversity, structural_risk, operational_risk

## Extension

Dynamic plugin loading is not required. New measurements are added by:

1. Writing a `Compute` function in the appropriate dimension file
2. Adding the `Definition` to the dimension's measurement list function
3. The `DefaultRegistry()` picks it up automatically

See [Adding a Measurement](../contributing/adding-a-measurement.md) for the full guide.

## File Map

```
internal/measurement/
├── registry.go          # Registry struct, Run, ComputeSnapshot, posture computation
├── default_registry.go  # DefaultRegistry() with all 18 measurements
└── measurement_test.go  # Registry tests (register, duplicate, grouping)
```
