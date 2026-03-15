# Common-Cause Clustering

## Overview

The `internal/clustering` package detects **common-cause clusters** — groups of tests that share a dependency (helper, fixture, setup path, or code unit) which appears responsible for broad instability, slowness, or concentrated signal patterns across the test suite.

Each cluster is a **candidate hypothesis**, not a proven root cause. The package surfaces evidence and confidence scores so downstream consumers (CLI, extension, reporting) can present findings with appropriate caveats.

## Cluster Types

| Type | Description |
|------|-------------|
| `dominant_slow_helper` | A code unit linked by multiple slow tests, suggesting it dominates runtime |
| `dominant_flaky_fixture` | A code unit linked by multiple flaky tests, suggesting shared non-determinism |
| `global_setup_path` | A directory where many test files share the same signal type, suggesting shared setup |
| `shared_import_dependency` | A code unit imported by many test files, creating broad blast radius for changes |
| `repeated_failure_pattern` | A directory with concentrated signals of the same type across multiple files |

## Detection Strategy

### Shared Import Clustering

Groups test files by their `LinkedCodeUnits`. When many test files depend on the same code unit, that unit has a large blast radius — any instability or performance regression in it will propagate broadly.

**Input:** `TestFile.LinkedCodeUnits`
**Threshold:** 3+ test files sharing the same code unit

### Slow Path Clustering

Filters test files to those with `slowTest` signals, then groups by shared code units. If multiple slow tests share the same dependency, that dependency is a candidate dominant slow helper.

**Input:** `TestFile.Signals` (type=slowTest), `TestFile.LinkedCodeUnits`, `TestFile.RuntimeStats`
**Threshold:** 3+ slow tests sharing the same code unit
**Impact metric:** Sum of `AvgRuntimeMs` across affected tests

### Flaky Fixture Clustering

Same approach as slow path clustering but for `flakyTest` signals. Shared code units across flaky tests are candidate sources of non-determinism.

**Input:** `TestFile.Signals` (type=flakyTest), `TestFile.LinkedCodeUnits`
**Threshold:** 3+ flaky tests sharing the same code unit

### Setup Path Clustering

Groups test files by directory and signal type. When all tests in a directory exhibit the same signal, the directory-level setup (shared fixtures, `TestMain`, `conftest.py`, etc.) is suspect.

**Input:** `TestFile.Path` (directory), `TestFile.Signals`
**Threshold:** 3+ test files in the same directory with the same signal type

### Repeated Failure Pattern

Groups snapshot-level signals by directory and type. Concentrated patterns suggest systemic issues rather than isolated occurrences.

**Input:** `Snapshot.Signals` (with `Location.File`)
**Threshold:** 3+ files in the same directory with the same signal type

## Confidence Scoring

Each cluster type uses a tailored confidence function:

- **Shared imports:** Based on the ratio of dependent tests to total tests. Higher ratio = higher confidence that changes will have broad impact.
- **Slow helpers:** Based on the ratio of affected slow tests to total slow tests. Clamped to [0.3, 0.95].
- **Flaky fixtures:** Based on the ratio of affected flaky tests to total flaky tests. Clamped to [0.4, 0.95].
- **Setup paths:** Scaled by count, clamped to [0.3, 0.85]. Lower ceiling because directory-level attribution is inherently less precise.
- **Repeated failures:** Scaled by count, clamped to [0.25, 0.80]. Lowest ceiling as this is purely pattern-based.

All confidence values stay below 1.0 to reflect that these are candidate hypotheses.

## Output

The `Detect` function returns a `ClusterResult` containing:

- `Clusters` — sorted by `AffectedCount` descending (ties broken by `CausePath` alphabetically)
- `TotalAffectedTests` — count of unique test files affected by any cluster

Each `Cluster` includes:

- Human-readable `Explanation` and `Evidence` strings
- Quantified `ImpactMetric` with `ImpactUnit` (e.g., total runtime in ms, affected file count)
- `Confidence` score in [0, 1]

## Usage

```go
import "github.com/pmclSF/terrain/internal/clustering"

result := clustering.Detect(snapshot)
for _, c := range result.Clusters {
    fmt.Printf("[%s] %s (confidence: %.0f%%, %d tests affected)\n",
        c.Type, c.CausePath, c.Confidence*100, c.AffectedCount)
}
```

## Design Decisions

1. **Minimum cluster size of 3.** A shared dependency affecting only 1-2 tests is normal and not actionable as a "cluster."

2. **Candidate-oriented language.** Explanations use phrases like "is a candidate root cause" and "may be the common cause" rather than asserting causation.

3. **No external dependencies.** The package uses only the standard library plus internal models/signals packages.

4. **Deterministic output.** Clusters are sorted by affected count (descending) then cause path (alphabetically) to ensure stable output across runs.
