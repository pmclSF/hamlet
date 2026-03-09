# Measurement Framework

## Overview

The measurement framework is a formal computation layer that sits between Hamlet's raw signals/derived facts and user-facing summaries, comparisons, and exports. It provides explicit, named, versioned measurements with evidence metadata that feed into posture dimensions.

## Philosophy

1. **Evidence-based.** Every measurement traces to concrete signals or observable facts. No measurement exists without inputs.
2. **Actionable.** Measurements should help users understand what to do, not just report a number.
3. **Stable.** Measurement IDs and semantics are durable across versions. Renaming or redefining a measurement requires a migration path.
4. **Honest about uncertainty.** Missing data reduces evidence strength, not the measurement value. A measurement with `evidence: none` is explicit about what it lacks.
5. **No fake precision.** Prefer bands and ratios over decimal scores. A ratio of 0.23 is meaningful; a "quality score of 72.4" is not.

## Architecture

```
Signals / Derived Facts
        │
        ▼
┌─────────────────────────┐
│  Measurement Definitions │  (18 definitions across 5 dimensions)
│  - ID, dimension, units  │
│  - Compute(snapshot)     │
│  - Evidence, limitations │
└──────────┬──────────────┘
           │
           ▼
┌─────────────────────────┐
│  Registry                │  Register(), Run(), ComputeSnapshot()
│  - Groups by dimension   │
│  - Duplicate ID check    │
│  - Deterministic order   │
└──────────┬──────────────┘
           │
           ▼
┌─────────────────────────┐
│  Posture Computation     │  Per-dimension band resolution
│  - Band: strong/moderate │
│    /weak/elevated/critical│
│  - Driving measurements  │
│  - Explanation            │
└──────────┬──────────────┘
           │
           ▼
┌─────────────────────────┐
│  Snapshot Persistence     │  Embedded in TestSuiteSnapshot.Measurements
│  - MeasurementSnapshot   │
│  - Serialized to JSON    │
│  - Supports compare/trend│
└─────────────────────────┘
```

## Pipeline Integration

Measurements run as **Step 8** in the engine pipeline (`internal/engine/pipeline.go`), after signal detection, ownership, runtime ingestion, risk scoring, and coverage ingestion:

```go
measRegistry := measurement.DefaultRegistry()
measSnap := measRegistry.ComputeSnapshot(snapshot)
snapshot.Measurements = measSnap.ToModel()
```

## Core Types

| Type | Package | Purpose |
|------|---------|---------|
| `Definition` | `measurement` | Declares a measurement: ID, dimension, inputs, compute function |
| `Result` | `measurement` | Output of a single measurement computation |
| `DimensionPosture` | `measurement` | Per-dimension posture with band and explanation |
| `Snapshot` | `measurement` | All measurements and posture for a point in time |
| `Registry` | `measurement` | Holds definitions, runs computations |
| `MeasurementSnapshot` | `models` | Serializable form embedded in TestSuiteSnapshot |
| `MeasurementResult` | `models` | Serializable form of a single measurement result |
| `DimensionPostureResult` | `models` | Serializable form of dimension posture |

## Evidence Model

Every measurement carries an `EvidenceStrength`:

| Level | Meaning |
|-------|---------|
| `strong` | Direct observation, high confidence (e.g., static analysis signals) |
| `partial` | Some data available but gaps noted (e.g., heuristic-based linkage) |
| `weak` | Limited data, best-effort (e.g., no runtime data for flakiness) |
| `none` | No data available (e.g., no test files found) |

Measurements also carry `Limitations` — human-readable strings describing specific data gaps.

## Inputs

Measurements may consume:

- **Raw signals** — counted by type from `snap.Signals`
- **Derived insights** — coverage summary, framework list, code units
- **Normalized facts** — test file counts, exported code unit counts
- **Snapshot history** — via comparison (measurement deltas across snapshots)

## File Map

```
internal/measurement/
├── measurement.go        # Core types: Dimension, Result, DimensionPosture, Definition
├── registry.go           # Registry, posture computation
├── default_registry.go   # Pre-populated registry with all 18 measurements
├── health.go             # Health dimension measurements (4)
├── coverage.go           # Coverage depth (3) and diversity (5) measurements
├── structural_risk.go    # Structural risk measurements (3)
├── operational_risk.go   # Operational risk measurements (3)
├── helpers.go            # Shared utilities (countSignals, ratioToBand, etc.)
└── measurement_test.go   # Comprehensive test suite
```

## Related Docs

- [Posture Dimensions](posture-dimensions.md) — what each dimension means
- [Measurement Registry](measurement-registry.md) — how registration works
- [Posture Computation](posture-computation.md) — how bands are derived
- [Measurement Persistence](measurement-persistence.md) — snapshot schema
