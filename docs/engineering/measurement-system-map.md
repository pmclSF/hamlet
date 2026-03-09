# Measurement System Map

## Overview

This document maps the complete measurement and posture system for maintainers. It covers every component, how they connect, and where to make changes.

## System Diagram

```
┌─────────────────────────────────────────────────────────┐
│  Engine Pipeline (internal/engine/pipeline.go)          │
│  Step 1-7: Analysis, signals, ownership, coverage       │
│  Step 8: Measurement computation                        │
│  Step 9: Deterministic sorting                          │
└──────────┬──────────────────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────────────┐
│  Measurement Registry (internal/measurement/)           │
│                                                         │
│  DefaultRegistry() → 18 definitions                     │
│  ┌─────────────┐ ┌──────────────┐ ┌───────────────┐   │
│  │ health (4)   │ │ cov_depth(3) │ │ cov_div (5)   │   │
│  └─────────────┘ └──────────────┘ └───────────────┘   │
│  ┌──────────────┐ ┌───────────────┐                    │
│  │ struct_risk(3)│ │ oper_risk (3) │                    │
│  └──────────────┘ └───────────────┘                    │
│                                                         │
│  ComputeSnapshot(snap) → Snapshot                       │
│    ├── Run() → []Result (18 measurements)               │
│    └── computeDimensionPosture() × 5 → []DimensionPosture│
└──────────┬──────────────────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────────────┐
│  Snapshot Model (internal/models/measurement.go)        │
│                                                         │
│  TestSuiteSnapshot.Measurements → MeasurementSnapshot   │
│    ├── Posture: []DimensionPostureResult                 │
│    └── Measurements: []MeasurementResult                 │
└──────────┬──────────────────────────────────────────────┘
           │
     ┌─────┴──────────┬──────────────┬──────────────┐
     ▼                ▼              ▼              ▼
┌──────────┐  ┌────────────┐  ┌──────────┐  ┌──────────┐
│ Posture  │  │ Executive  │  │ Compare  │  │ Benchmark│
│ Report   │  │ Summary    │  │          │  │ Export   │
│          │  │            │  │ Posture  │  │          │
│ hamlet   │  │ Prefers    │  │ Deltas   │  │ Posture  │
│ posture  │  │ measurement│  │ Measure  │  │ Bands    │
│          │  │ posture    │  │ Deltas   │  │ (safe)   │
└──────────┘  └────────────┘  └──────────┘  └──────────┘
```

## Component Inventory

### Core Measurement Code

| File | Purpose | Key Functions |
|------|---------|---------------|
| `internal/measurement/measurement.go` | Core types | Dimension, Result, DimensionPosture, Definition |
| `internal/measurement/registry.go` | Registry + posture | Register, Run, ComputeSnapshot, computeDimensionPosture |
| `internal/measurement/default_registry.go` | Default setup | DefaultRegistry() |
| `internal/measurement/health.go` | Health measurements | computeFlakyShare, computeSkipDensity, computeDeadTestShare, computeSlowTestShare |
| `internal/measurement/coverage.go` | Coverage measurements | computeUncoveredExports, computeWeakAssertionShare, computeCoverageBreachShare, computeMockHeavyShare, computeFrameworkFragmentation, computeE2EConcentration, computeE2EOnlyUnits, computeUnitTestCoverage |
| `internal/measurement/structural_risk.go` | Structural risk | computeMigrationBlockerDensity, computeDeprecatedPatternShare, computeDynamicGenerationShare |
| `internal/measurement/operational_risk.go` | Operational risk | computePolicyViolationDensity, computeLegacyFrameworkShare, computeRuntimeBudgetBreachShare |
| `internal/measurement/helpers.go` | Shared utilities | countSignals, ratioToBand, runtimeEvidence, evidenceLimitations |
| `internal/measurement/measurement_test.go` | Tests | 36 test functions, 615 lines |

### Integration Points

| File | How It Uses Measurements |
|------|------------------------|
| `internal/engine/pipeline.go` | Step 8: runs registry, embeds in snapshot |
| `internal/models/measurement.go` | Serializable measurement types |
| `internal/summary/executive.go` | Prefers measurement posture for dimension display |
| `internal/reporting/posture_report.go` | Renders full posture + measurement evidence |
| `internal/reporting/comparison_report.go` | Renders posture and measurement deltas |
| `internal/comparison/compare.go` | compareMeasurements() produces deltas |
| `internal/benchmark/export.go` | Extracts PostureBands from measurements |
| `cmd/hamlet/main.go` | CLI: `hamlet posture`, `hamlet posture --json` |

### Test Coverage

| Test Location | What It Tests |
|---------------|---------------|
| `internal/measurement/measurement_test.go` | All 18 measurements, registry, posture computation |
| `internal/comparison/compare_test.go` | PostureDeltas, MeasurementDeltas |
| `internal/testdata/golden/posture-minimal.txt` | Golden output for posture report |

## Adding a New Measurement

1. Write a `compute` function in the appropriate dimension file
2. Add a `Definition` to the dimension's measurement list function
3. The `DefaultRegistry()` picks it up automatically
4. Add tests in `measurement_test.go`

See [Adding a Measurement](../contributing/adding-a-measurement.md).

## Adding a New Dimension

1. Add a `Dimension` constant in `measurement.go`
2. Create a new file for the dimension's measurements
3. Add a measurement list function (e.g., `NewDimensionMeasurements()`)
4. Register in `DefaultRegistry()`
5. Add to the `dims` slice in `ComputeSnapshot()`
6. Update comparison and reporting to handle the new dimension

## Known Limitations

1. **No measurement history query.** Measurements are compared pairwise across snapshots but not indexed for time-series queries.
2. **No weighted posture.** All measurements within a dimension are treated equally in posture computation.
3. **No cross-dimension correlation.** Measurements don't reference other dimensions' results.

These are documented here as potential future work, not current gaps.
