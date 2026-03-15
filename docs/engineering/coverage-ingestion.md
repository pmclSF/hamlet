# Coverage Ingestion

Terrain ingests structured coverage artifacts and normalizes them into an internal model suitable for attribution to code units and test types.

## Supported Formats

| Format | Detection | Line Hits | Branch Hits | Function Hits |
|--------|-----------|-----------|-------------|---------------|
| LCOV | `SF:` prefix, `.lcov` extension, `lcov.info` filename | ✅ DA: records | ✅ BRDA: records | ✅ FN:/FNDA: records |
| Istanbul JSON | JSON with `statementMap`/`s`/`fnMap`/`f`/`branchMap`/`b` | ✅ via statements | ✅ branch locations | ✅ function map |

Format is auto-detected by content inspection.

## Architecture

```
internal/coverage/
    model.go        CoverageRecord, CoverageArtifact, MergedCoverage
    ingest.go       IngestFile, IngestDirectory, LCOV/Istanbul parsers
    attribute.go    AttributeToCodeUnits
    bytype.go       ComputeByType, BuildRepoSummary
    pertest.go      Per-test coverage model
    insights.go     Derived coverage insights
```

## CoverageRecord Model

```go
type CoverageRecord struct {
    FilePath             string         // repository-relative path
    LineHits             map[int]int    // 1-based line → hit count
    BranchHits           map[string]int // branch ID → hit count
    FunctionHits         map[string]int // function name → hit count
    LineCoveredCount     int            // lines executed
    LineTotalCount       int            // lines instrumented
    BranchCoveredCount   int
    BranchTotalCount     int
    FunctionCoveredCount int
    FunctionTotalCount   int
}
```

## Artifact Provenance

Every ingested artifact carries provenance metadata:

```go
type ArtifactProvenance struct {
    SourceFile string // path to original coverage file
    Format     string // "lcov" or "istanbul"
    RunLabel   string // optional test type label (e.g., "unit", "e2e")
}
```

## Path Normalization

Coverage artifacts often contain absolute paths. `normalizeCoveragePath()` strips absolute prefixes by looking for common project root markers (`/src/`, `/lib/`, `/app/`, `/packages/`).

## Merging

`Merge()` combines multiple coverage artifacts into a unified view:
- Hit counts are summed across artifacts
- Summary counts are recomputed from merged hit maps
- Artifact provenance is preserved

## Labeled Runs

Coverage artifacts can be tagged with a `RunLabel` (e.g., `"unit"`, `"e2e"`, `"integration"`) to enable coverage-by-type analysis. Labels are passed via CLI flags when ingesting coverage.

## Pipeline Integration

Coverage ingestion is triggered by `--coverage` flag in the CLI:

```
engine.RunPipeline(root, PipelineOptions{CoveragePath: path})
    → ingestCoverage()
    → IngestFile/IngestDirectory
    → Merge artifacts
    → AttributeToCodeUnits
    → ComputeByType
    → DeriveInsights
    → Populate snapshot.CoverageSummary and snapshot.CoverageInsights
```

## Degradation Behavior

| Scenario | Behavior |
|----------|----------|
| Only line hits available | Line coverage computed, function/branch = -1 |
| No function-level data | Function hit = -1, evidence quality = "approximate" |
| Missing file in coverage | Unit marked as "unavailable" evidence quality |
| Unreadable artifact | Warning emitted, artifact skipped |
| No coverage path specified | Coverage fields remain nil in snapshot |
