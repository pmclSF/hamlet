# Impact Analysis System Map

Stage 132 architecture document for Terrain's impact analysis subsystem.

## Subsystem Inventory

| Package | File | Primary Function | Purpose |
|---------|------|-----------------|---------|
| `internal/impact` | `impact.go` | `AnalyzeChangeSet`, `Analyze` | Entry points; `AnalyzeChangeSet` is preferred (starts from `models.ChangeSet`), `Analyze` is the legacy bridge (starts from `ChangeScope`). Defines domain types (`ChangeScope`, `ImpactResult`, `ImpactedCodeUnit`, `ProtectionGap`, `ChangeRiskPosture`) |
| `internal/impact` | `analysis.go` | `mapChangedUnits`, `findImpactedTests`, `findProtectionGaps`, `selectProtectiveTests`, `buildProtectiveSet`, `computeChangeRiskPosture` | Core analysis logic: unit mapping, test selection, gap detection, risk dimensions, coverage diversity |
| `internal/impact` | `graph.go` | `BuildImpactGraph` | Constructs a bidirectional impact graph connecting code units to tests via exact coverage, bucket coverage, structural links, and name conventions |
| `internal/impact` | `changeset_builder.go` | `ChangeSetFromGitDiff`, `ChangeSetFromPaths`, `ChangeSetFromCIList`, `ChangeSetFromComparison`, `ChangeSetToScope` | Builders that produce a `models.ChangeSet` from git diffs, explicit paths, CI file lists, or snapshot comparisons. Includes SHA resolution, shallow-clone detection, package/service/config inference. `ChangeSetToScope` bridges to legacy `ChangeScope` |
| `internal/impact` | `changescope.go` | `ChangeScopeFromGitDiff`, `ChangeScopeFromPaths`, `ChangeScopeFromCIList`, `ChangeScopeFromComparison` | Legacy adapters that produce a `ChangeScope` (preserved for backward compatibility) |
| `internal/models` | `changeset.go` | `ChangeSet` | Normalized, serializable representation of a code change: repo, SHAs, changed files/packages/services/configs, shallow-clone metadata |
| `internal/impact` | `filter.go` | `FilterByOwner` | Filters an `ImpactResult` to a single owner's code units, tests, and gaps |
| `internal/impact` | `aggregate.go` | `BuildAggregate` | Produces a privacy-safe `Aggregate` with counts and ratios; suppresses breakdowns below the privacy threshold (3) |
| `internal/impact` | `impact_test.go` | Tests | Inline-fixture tests covering all public functions and edge cases |
| `internal/changescope` | `model.go` | Types | `PRAnalysis`, `ChangeScopedFinding`, `PostureDelta` types for PR-level output |
| `internal/changescope` | `analyze.go` | `AnalyzePRFromChangeSet`, `AnalyzePR` | `AnalyzePRFromChangeSet` is preferred (starts from `models.ChangeSet`). `AnalyzePR` is the legacy path. Both wrap impact analysis with PR-specific findings, posture delta, and summary |
| `internal/changescope` | `render.go` | `RenderPRSummaryMarkdown`, `RenderPRCommentConcise`, `RenderCIAnnotation`, `RenderChangeScopedReport` | Rendering functions for PR markdown, concise comments, CI annotations, and terminal reports |
| `internal/reporting` | `impact_report.go` | `RenderImpactReport` | Full human-readable impact report for terminal output |
| `internal/reporting` | `impact_drilldown.go` | `RenderImpactUnits`, `RenderImpactGaps`, `RenderImpactTests`, `RenderImpactGraph`, `RenderProtectiveSet`, `RenderImpactOwners` | Drill-down views for each `--show` subview |

## Data Flow

```
ChangeSet (git-diff | explicit | ci-list | snapshot-compare)
    |  resolves SHAs, infers packages/services/configs,
    |  detects shallow clones
    |
    v
ChangeSetToScope(cs)             -- bridge to legacy ChangeScope
    |
    v
BuildImpactGraph(snap)           -- bidirectional unit<->test graph
    |
    v
mapChangedUnits(scope, snap)     -- changed files -> ImpactedCodeUnit[]
    |
    v
findImpactedTests(scope, snap, units)  -- coverage links + proximity heuristics
    |
    v
findProtectionGaps(units, tests, snap) -- untested exports, weak coverage, diversity gaps
    |
    v
selectProtectiveTests(tests, units)    -- exact-first, fallback to inferred
    |
    v
buildProtectiveSet(result)       -- enhanced set with SelectionReason per test
    |
    v
computeChangeRiskPosture(result) -- 4 dimensions: protection, exposure, coordination, instability
    |
    v
collectOwners / buildSummary / identifyLimitations
    |
    v
ImpactResult
```

## Integration Points

| Subsystem | Direction | Mechanism |
|-----------|-----------|-----------|
| **Coverage ingestion** | Input | `TestFile.LinkedCodeUnits` provides per-test coverage lineage used to build exact graph edges |
| **Ownership** | Input | `Snapshot.Ownership` map resolves file-to-owner; used by `FilterByOwner` and the coordination risk dimension |
| **Lifecycle / quality** | Input | `Snapshot.Signals` are surfaced as `existing_signal` findings in PR analysis |
| **ChangeScope** | Input | Adapters in `changescope.go` produce `ChangeScope` from git, CI, or snapshot sources |
| **Reporting** | Output | `impact_report.go` and `impact_drilldown.go` consume `ImpactResult` for terminal rendering |
| **PR / changescope** | Output | `changescope.AnalyzePR` wraps impact analysis for PR-specific outputs and markdown rendering |
| **Benchmark export** | Output | `BuildAggregate` produces privacy-safe counts for cross-repo benchmarking |

## CLI Command Routing

| Command | Flag | Handler | Renderer |
|---------|------|---------|----------|
| `terrain impact` | (none) | `runImpact` | `RenderImpactReport` |
| `terrain impact` | `--show units` | `runImpact` | `RenderImpactUnits` |
| `terrain impact` | `--show gaps` | `runImpact` | `RenderImpactGaps` |
| `terrain impact` | `--show tests` | `runImpact` | `RenderImpactTests` |
| `terrain impact` | `--show owners` | `runImpact` | `RenderImpactOwners` |
| `terrain impact` | `--show graph` | `runImpact` | `RenderImpactGraph` |
| `terrain impact` | `--show selected` | `runImpact` | `RenderProtectiveSet` |
| `terrain impact` | `--owner NAME` | `runImpact` | Applies `FilterByOwner` before rendering |
| `terrain impact` | `--json` | `runImpact` | JSON-encoded `ImpactResult` |
| `terrain select-tests` | (none) | `runSelectTests` | `RenderProtectiveSet` |
| `terrain select-tests` | `--json` | `runSelectTests` | JSON-encoded `ProtectiveTestSet` |
| `terrain pr` | (none) | `runPR` | `RenderChangeScopedReport` |
| `terrain pr` | `--format markdown` | `runPR` | `RenderPRSummaryMarkdown` |
| `terrain pr` | `--format comment` | `runPR` | `RenderPRCommentConcise` |
| `terrain pr` | `--format annotation` | `runPR` | `RenderCIAnnotation` |

## Rendering Layer

- **`internal/reporting/impact_report.go`** -- Top-level terminal report: summary, posture, units, gaps, selected tests, owners, limitations, and next-step hints.
- **`internal/reporting/impact_drilldown.go`** -- Six focused renderers (`RenderImpactUnits`, `RenderImpactGaps`, `RenderImpactTests`, `RenderImpactGraph`, `RenderProtectiveSet`, `RenderImpactOwners`) for `--show` subviews.
- **`internal/changescope/render.go`** -- Four PR renderers: full markdown summary with posture badge and stats table, concise one-line comment, CI `::error`/`::warning` annotations, and human-readable terminal report.

All renderers accept an `io.Writer` and the relevant result type. They share a consistent `line`/`blank` helper pattern for formatted output.
