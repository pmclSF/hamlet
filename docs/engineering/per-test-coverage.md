# Per-Test Coverage

Per-test coverage is an optional enrichment layer that maps individual test IDs to the code units they cover. This unlocks precise coverage lineage, redundancy detection, and impact analysis.

## Data Model

### TestCoverageRecord (Input)

Raw per-test coverage data, typically from specialized tooling:

```go
type TestCoverageRecord struct {
    TestID       string // stable test identifier
    FilePath     string // source file covered
    CoveredLines []int  // 1-based line numbers hit by this test
}
```

### PerTestCoverage (Output)

Enriched per-test coverage with code unit joins:

```go
type PerTestCoverage struct {
    TestID         string            // stable test ID
    CoveredFiles   map[string][]int  // file → covered lines
    CoveredUnitIDs []string          // matched code unit IDs
    ScopeBreadth   int               // number of distinct files touched
}
```

## How It Works

`BuildPerTestCoverage()`:

1. Groups raw `TestCoverageRecord`s by test ID
2. For each test, collects all covered files and lines
3. Matches covered lines against `UnitSpan` index (code unit line ranges)
4. Produces `PerTestCoverage` with file map, unit IDs, and scope breadth

## UnitSpan Index

`BuildUnitSpanIndex()` creates a file-to-spans mapping from code unit coverage data:

```go
type UnitSpan struct {
    UnitID    string
    StartLine int
    EndLine   int  // defaults to start+50 if unknown
}
```

Line overlap detection (`linesOverlap()`) checks whether any covered line falls within a unit's span.

## Capabilities Unlocked

When per-test coverage is available:

| Capability | Description |
|------------|-------------|
| **Coverage lineage** | Which tests cover which functions |
| **Redundancy detection** | Multiple tests covering identical code |
| **Impact analysis** | Changed function → affected tests |
| **Scope analysis** | Tests that touch too many files (overbroad) |
| **Flake impact** | Flaky test → impacted code surface |

## Graceful Degradation

Per-test coverage is strictly optional:

| Data Available | Behavior |
|----------------|----------|
| Per-test coverage records | Full per-test attribution |
| Bucket/run-level coverage only | Coverage-by-type (unit/integration/e2e) |
| No coverage at all | Coverage fields remain nil |

The system never fails or degrades incorrectly when per-test data is absent. All other coverage features (ingestion, attribution, coverage-by-type, insights) work independently.

## Current Status

The data model and join logic are implemented. Per-test coverage ingestion from specific tooling (e.g., Jest `--collectCoverageFrom` per test, Go `-coverprofile` per test) is documented as future work.
