# Analysis Graph

The analysis graph (`internal/graph`) is a lightweight in-memory index that connects Terrain's core entities for cross-cutting queries. It is built on demand from a `TestSuiteSnapshot` and provides O(1) lookups that would otherwise require repeated linear scans.

## Design Principles

- **Not a database.** The graph is an in-memory read-only index. No persistence, no query language.
- **Built once, query many.** `graph.Build(snapshot)` constructs all indexes in a single O(n) pass. After construction, the graph is immutable.
- **No circular imports.** Depends only on `internal/models`. Downstream consumers (reporting, summary, heatmap) import the graph; the graph never imports them.
- **Deterministic iteration.** Sorted results where iteration order matters.

## What the Graph Indexes

| Entity | Index | Query Example |
|--------|-------|---------------|
| TestCase | `TestByID`, `TestsByFile`, `TestsByOwner`, `TestsByType` | "All unit tests owned by @team-auth" |
| CodeUnit | `UnitByID`, `UnitsByFile`, `UnitsByOwner`, `ExportedUnits` | "All exported units in src/api/" |
| Signal | `SignalsByFile`, `SignalsByOwner`, `SignalsByType` | "All quality signals for @team-api" |
| Health | `HealthSignalsByTestID` | "Health signals for test t4" |
| Coverage | `UncoveredExportedUnits`, `E2EOnlyUnits` | "Exported units with no coverage" |
| Files | `FileOwner`, `DirectoryFiles` | "Who owns src/auth/auth.js?" |

## Query Methods

```go
g := graph.Build(snapshot)

// Tests touching a module
testIDs := g.TestsInModule("src/auth")

// Code units covered by type
unitIDs := g.UnitsByOwner["@team-api"]

// Owners with weakly covered units
uncovered := g.UncoveredExportedForOwner("@team-auth")

// Tests associated with failures/slowdowns
top := g.TopFailingTestIDs(5)

// Per-owner risk aggregation
summaries := g.OwnerRiskSummaries()

// Per-directory coverage quality
modules := g.ModuleCoverageSummaries()
```

## Integration Points

The graph is consumed by:

- **Executive summary** (`internal/summary`): coverage-backed recommendations, concentrated instability detection
- **Heatmap** (`internal/heatmap`): owner hotspot enrichment with coverage data
- **Reporting** (`internal/reporting`): comparison reports with test identity and coverage trends
- **Benchmark export** (`internal/benchmark`): privacy-safe coverage-by-type aggregates

## Pipeline Position

```
Raw Artifacts (source files, coverage, runtime)
        |
   [Analysis Layer]     internal/analysis, internal/coverage, internal/runtime
        |
   [Signal Detection]   internal/signals, internal/health, internal/quality
        |
   [Analysis Graph]     internal/graph  <-- indexes normalized facts
        |
   [Measurement Layer]  internal/measurement
        |
   [Reporting Layer]    internal/reporting, internal/summary, internal/benchmark
```

The graph sits between raw signal detection and the reporting/measurement layers. It provides the normalized facts that downstream systems query without duplicating join logic.

## Test Identity Integration

Health signals (slowTest, flakyTest, skippedTest) now carry `testId` in their metadata when runtime results can be joined to extracted test cases. The join happens in `runtime.ResolveTestIDs()` during pipeline execution:

1. **Exact match**: runtime file + name matches TestCase.FilePath + TestName
2. **Suffix match**: runtime absolute path ends with TestCase relative path
3. **Fuzzy match**: runtime name contains TestCase name (parameterized tests)

When resolution fails, `TestID` remains empty and health signals degrade to file/suite-level attribution.

## Coverage-by-Type Integration

Coverage-by-type data flows through:

1. **Ingestion**: labeled coverage artifacts (`--coverage unit:path`) produce per-type attribution
2. **Snapshot**: `CoverageSummary` captures aggregate by-type counts
3. **Graph**: indexes `UncoveredExportedUnits` and `E2EOnlyUnits`
4. **Measurement**: `coverage_diversity.e2e_only_units` and `coverage_diversity.unit_test_coverage`
5. **Recommendations**: executive summary surfaces e2e-only and uncovered export findings
6. **Migration**: coverage guidance includes e2e-only data for migration risk
7. **Benchmark**: privacy-safe percentages exported without file paths
