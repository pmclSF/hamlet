# Integration Map: Test Identity & Coverage-by-Type

How stable test IDs and coverage-by-type flow through every Terrain subsystem.

## Test Identity Flow

Stable test IDs connect runtime observations back to the extracted test inventory.

```
[analysis/]  TestCase extraction → snap.TestCases (with stable IDs)
      |
[runtime/]   Ingest JUnit/Jest → []TestResult (raw, no TestID)
      |
[runtime/resolve]  ResolveTestIDs(results, testCases)
      |              ├─ Exact: file + name match
      |              ├─ Suffix: runtime path ends with TestCase relative path
      |              └─ Fuzzy: runtime name contains TestCase name
      |
[health/]    SlowTestDetector, FlakyTestDetector, SkippedTestDetector
      |        → Signals with metadata["testId"] when resolved
      |
[graph/]     HealthSignalsByTestID index → O(1) lookup
      |
[summary/]   Concentrated instability detection
      |        → "N tests have multiple health signals"
      |
[comparison/]  TestCase diff (added/removed/stable)
      |
[reporting/]   Test Identity Changes section in comparison report
      |
[benchmark/]   TestIdentityStats aggregate (counts + type distribution)
                 → Privacy-safe: no test names or file paths
```

### Graceful Degradation

When test ID resolution fails (no matching TestCase), signals degrade:
- `metadata["testId"]` is omitted (not set to empty string)
- Health signals still carry file and symbol for file-level attribution
- `HealthSignalsByTestID` index simply has no entry for that signal
- Summary concentrated-instability detection skips unresolved signals
- Comparison report still shows test case diffs from static extraction

## Coverage-by-Type Flow

Coverage-by-type distinguishes unit test coverage from e2e coverage.

```
[coverage/]    Labeled artifacts (--coverage unit:path, e2e:path)
      |          → CoverageArtifact with Type field
      |
[coverage/]    ComputeByType(artifacts, codeUnits)
      |          → per-unit coverage type attribution
      |
[coverage/]    BuildRepoSummary → CoverageSummary
      |          ├─ CoveredByUnitTests
      |          ├─ CoveredByE2E
      |          ├─ CoveredOnlyByE2E
      |          └─ UncoveredExported
      |
[coverage/]    DeriveInsights → CoverageInsight[]
      |          └─ Type: "e2e_only_coverage" with file path
      |
[graph/]       UncoveredExportedUnits, E2EOnlyUnits indexes
      |
[measurement/] coverage_diversity.e2e_only_units
      |         coverage_diversity.unit_test_coverage
      |          → Posture bands: strong / moderate / weak / unknown
      |
[summary/]     Coverage-backed recommendations
      |          ├─ "N code units covered only by e2e"
      |          ├─ "N exported functions have no test coverage"
      |          └─ "Unit test coverage below 50%"
      |
[migration/]   CoverageGuidance enhanced with e2e-only directories
      |          → High-priority: "N units covered only by e2e — no fast feedback"
      |
[comparison/]  CoverageDelta with UnitTestCoverage before/after
      |
[reporting/]   "Coverage by Type" section in summary report
      |         "Coverage Trend" section in comparison report
      |
[benchmark/]   CoverageByType aggregate (percentages + diversity band)
                 → Privacy-safe: no file paths
```

### Graceful Degradation

When coverage data is absent (`CoverageSummary` is nil):
- `coverage_diversity.*` measurements report band "unknown"
- Graph `UncoveredExportedUnits` and `E2EOnlyUnits` are empty slices
- Summary skips coverage-backed recommendations
- Migration guidance omits e2e-only directory flags
- Comparison `CoverageDelta` fields are zero-valued
- Benchmark `CoverageByType` is nil (omitted from JSON)

## Subsystem Integration Matrix

| Subsystem | Test Identity | Coverage-by-Type | New in Current Schema |
|-----------|:---:|:---:|:---:|
| `internal/graph` | HealthSignalsByTestID | UncoveredExportedUnits, E2EOnlyUnits | Yes |
| `internal/health` | testId in signal metadata | - | Modified |
| `internal/runtime` | ResolveTestIDs | - | New file |
| `internal/measurement` | - | e2e_only_units, unit_test_coverage | Modified |
| `internal/summary` | Concentrated instability | Coverage recommendations | Modified |
| `internal/comparison` | TestCase diff | UnitTestCoverage delta | Modified |
| `internal/migration` | - | e2e-only directory guidance | Modified |
| `internal/reporting` | Identity changes section | Coverage-by-type sections | Modified |
| `internal/benchmark` | TestIdentityStats | CoverageByType aggregate | Current schema |
| `internal/engine` | ResolveTestIDs in pipeline | - | Modified |

## Privacy Boundary

All external-facing outputs (benchmark export) contain only:
- **Counts**: total tests, tests by type, health signal count
- **Percentages**: unit coverage %, e2e-only %, uncovered exported %
- **Bands**: diversity band (strong/moderate/weak)

Never exported: file paths, test names, symbol names, owner identifiers, or source code.

## Related Documents

- [Analysis Graph](analysis-graph.md) — graph indexes and query methods
- [Integration Map: Signals, Quality, Migration](integration-map.md) — signal type flow
- [Test Identity](test-identity.md) — test case extraction and stable ID design
- [Coverage Ingestion](coverage-ingestion.md) — coverage artifact parsing
- [Coverage Attribution](coverage-attribution.md) — code unit coverage mapping
