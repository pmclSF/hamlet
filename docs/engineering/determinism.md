# Deterministic Output Contract

Hamlet guarantees that running the same analysis on the same input produces
byte-identical JSON output (excluding timestamps). This contract is enforced
by tests and architectural rules.

## Why Determinism Matters

- **Snapshot comparison** (`hamlet compare`) depends on stable ordering to
  produce meaningful diffs.
- **CI gating** requires that repeated runs yield the same policy verdict.
- **Extension rendering** should not shuffle items between refreshes.
- **Benchmark exports** must be reproducible for external consumers.

## How Determinism Is Achieved

### Canonical Sort (`models.SortSnapshot`)

Before the pipeline returns, `models.SortSnapshot()` sorts all slice fields
in the `TestSuiteSnapshot` into a canonical order:

| Field | Sort Key |
|-------|----------|
| TestFiles | Path (ascending) |
| TestCases | TestID (ascending) |
| CodeUnits | UnitID, then Path, then Name |
| Frameworks | Name (ascending) |
| Signals | Category, Type, File, Line, Explanation |
| Risk | Type, Scope, ScopeName |
| CoverageInsights | Type, Path, UnitID |

### Upstream Sorts

Several packages independently sort their outputs:

- `analysis.detectLanguages` sorts language names
- `analysis.detectPackageManagers` sorts package manager names
- `analysis.detectCISystems` sorts CI system names
- `scoring.ComputeRisk` sorts risk surfaces by type
- `heatmap.Build` sorts hotspots by score descending, name ascending
- `comparison.Compare` sorts deltas by type

The pipeline-level `SortSnapshot` call is a safety net ensuring canonical
order even if upstream code changes.

### Timestamp Isolation

Timestamps (`GeneratedAt`, `Repository.SnapshotTimestamp`) are the only
fields expected to vary between runs. All other fields are deterministic.

## Nondeterminism Sources (Mitigated)

| Source | Mitigation |
|--------|-----------|
| Go map iteration | All map-derived slices are sorted before use |
| `filepath.WalkDir` | Returns lexical order on all major OS; `SortSnapshot` provides defense-in-depth |
| Detector execution order | Registry runs detectors in registration order; `SortSnapshot` normalizes final signal order |
| Goroutine scheduling | No concurrent mutation of shared slices |

## Test Enforcement

- `TestPipelineDeterminism` — runs pipeline twice on identical input, verifies byte-identical JSON
- `TestPipelineOutputSorted` — runs pipeline on real testdata, verifies all slices are sorted
- `TestSortSnapshot_*` — unit tests for each sort dimension and idempotency
