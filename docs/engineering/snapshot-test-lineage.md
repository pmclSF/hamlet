# Snapshot Test Lineage

Terrain persists test identities and coverage lineage in snapshots, enabling longitudinal tracking and structural regression detection.

## Snapshot Schema

The `TestSuiteSnapshot` includes:

```go
type TestSuiteSnapshot struct {
    SnapshotMeta     SnapshotMeta       // schema version, engine version, detectors
    Repository       RepositoryMetadata
    Frameworks       []Framework
    TestFiles        []TestFile
    TestCases        []TestCase         // ← individual tests with stable IDs
    CodeUnits        []CodeUnit         // ← code unit inventory
    Signals          []Signal
    Risk             []RiskSurface
    Measurements     *MeasurementSnapshot
    CoverageSummary  *CoverageSummary   // ← aggregate coverage metrics
    CoverageInsights []CoverageInsight  // ← actionable coverage findings
    Ownership        map[string][]string
    GeneratedAt      time.Time
}
```

### TestCase in Snapshots

Each test case carries its full identity:

```json
{
  "testId": "a1b2c3d4e5f67890",
  "canonicalIdentity": "src/__tests__/auth.test.js::AuthService::should login",
  "filePath": "src/__tests__/auth.test.js",
  "suiteHierarchy": ["AuthService"],
  "testName": "should login",
  "framework": "jest",
  "language": "js",
  "line": 15,
  "extractionKind": "static",
  "confidence": 0.9,
  "testType": "unit",
  "testTypeConfidence": 0.75,
  "testTypeEvidence": ["framework jest is typically used for unit tests", "path contains unit test directory: __tests__"]
}
```

## Comparison / Trend Support

The `comparison.Compare()` function detects test-level changes between snapshots:

### TestCaseDeltas

```go
type TestCaseDeltas struct {
    Added           int      // new test IDs
    Removed         int      // deleted test IDs
    Stable          int      // unchanged test IDs
    AddedExamples   []string // representative new test names
    RemovedExamples []string // representative removed test names
}
```

### CoverageDelta

```go
type CoverageDelta struct {
    LineCoverageBefore      float64
    LineCoverageAfter       float64
    LineCoverageDelta       float64
    UncoveredExportedBefore int
    UncoveredExportedAfter  int
    CoveredOnlyByE2EBefore  int
    CoveredOnlyByE2EAfter   int
}
```

## What Can Be Detected

| Change | Detection Method |
|--------|-----------------|
| Tests added | Test IDs in `to` but not `from` |
| Tests removed | Test IDs in `from` but not `to` |
| Tests renamed | Old ID removed + new ID added (heuristic) |
| Coverage regression | `LineCoverageDelta < 0` |
| E2E dependency increase | `CoveredOnlyByE2E` increase |
| Uncovered export increase | `UncoveredExported` increase |

## Schema Versioning

Snapshots include `SnapshotMeta.SchemaVersion` (currently `"1.0.0"`). When the schema changes incompatibly, the version increments. Comparison handles version mismatches by checking for field presence.

## Determinism

All snapshot slices are sorted into canonical order by `models.SortSnapshot()` before serialization. This ensures:
- Identical repo state → identical JSON output
- Diff-friendly snapshots
- Reliable comparison across runs

## Limitations

- Test rename detection is not yet implemented (requires similarity heuristics)
- Code unit add/remove tracking is aggregate-level, not per-unit
- Per-test coverage lineage is not yet persisted in snapshots (model exists, ingestion is future work)
