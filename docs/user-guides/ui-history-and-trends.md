# Stage 118 -- UI History, Trends, and Snapshot Browser

Design specification for the Hamlet extension UI's longitudinal views.
This document describes views and data access patterns, not implementation.

## Overview

Hamlet snapshots (`TestSuiteSnapshot`) are point-in-time artifacts. The history
UI lets users browse saved snapshots, compare posture across time, and surface
trends such as newly unstable tests, lost coverage, and portfolio drift. All
trend computation happens in the engine or at snapshot-diff time -- the UI
renders pre-computed or cheaply derivable comparisons.

## Snapshot Navigation

### Snapshot list

The snapshot browser presents a chronological list of saved snapshots from the
local `.hamlet/snapshots/` directory. Each entry shows:

- `GeneratedAt` timestamp
- `SnapshotMeta.EngineVersion`
- `SnapshotMeta.DetectorCount` (how many detectors ran)
- Repository name from `Repository`
- A posture summary row: one badge per `DimensionPostureResult.Band`

Users can pin two snapshots for side-by-side comparison (see below).

### Snapshot detail

Selecting a snapshot opens the standard drill-down hierarchy from Stage 117,
but in read-only mode (no triage actions). This allows reviewing historical
posture without modifying current triage state.

## Trend Views

Trends are derived by comparing the current (most recent) snapshot against one
or more prior snapshots. The UI computes lightweight diffs client-side using
the stable identifiers already present in the models.

### Posture Movement

For each `DimensionPostureResult.Dimension`, compare `Band` values across
snapshots. Display a timeline showing band transitions (e.g., coverage went
from `green` to `yellow` between snapshots 5 and 6).

Data source: `Measurements.Posture[].Band` and `Measurements.Posture[].Dimension`.

### Newly Unstable Tests

Cross-reference `TestCase.TestID` across snapshots. A test is "newly unstable"
if it appears in the current snapshot's signals with type `flakyTest` or
`unstableSuite` but did not have that signal in the prior snapshot. This aligns
with the `stability.ClassNewlyUnstable` classification.

Data source: `Signals[]` filtered by `SignalType == "flakyTest"` or
`"unstableSuite"`, keyed by `Signal.Location` (file + symbol).

The stability subsystem's `Classification.RecentTrend` field ("improving",
"worsening", "stable") provides a pre-computed trend when history depth
is sufficient (`HistoryDepth >= MinHistoryDepth`).

### Lost Coverage

Compare `CoverageSummary` fields across snapshots:
- `UncoveredExported` increasing means exported units lost coverage.
- `LineCoveragePct` or `BranchCoveragePct` dropping indicates regression.

Also diff `CoverageInsights[]` by `UnitID` to identify specific code units
that were covered in a prior snapshot but are uncovered now.

### Portfolio Drift

Compare `Portfolio.Aggregates` across snapshots:
- `RedundancyCandidateCount` increasing suggests new test overlap.
- `LowValueHighCostCount` increasing suggests tests becoming expensive without
  proportional coverage.
- `PortfolioPostureBand` changing indicates overall portfolio health shift.

Per-owner drift is visible through `PortfolioAggregates.ByOwner[]`:
compare each `PortfolioOwnerSummary` by owner ID across snapshots.

### Signal Count Trends

Aggregate `Signals[]` by `SignalType` and `SignalCategory` per snapshot.
Show a bar chart of signal counts over time, highlighting types with
monotonically increasing counts (worsening) or decreasing counts (improving).

## Comparison Summaries

When two snapshots are pinned for comparison, the UI generates a summary panel:

| Field                        | Source                                      |
|------------------------------|---------------------------------------------|
| Posture band changes         | `DimensionPostureResult.Band` diff          |
| New signals                  | Signals present in B but not A by location  |
| Resolved signals             | Signals present in A but not B by location  |
| Test count delta             | `len(TestCases)` diff                       |
| Coverage delta               | `CoverageSummary` field diffs               |
| New protection gaps          | `PostureDelta.NewGapCount` if available     |
| Owner changes                | `OwnershipSummary.OwnerCount` diff          |

### Drill-down from comparison

Each row in the comparison summary is clickable. Clicking "New signals" opens
a filtered signal list showing only signals whose `SignalLocation` key exists
in snapshot B but not in snapshot A.

## Handling Missing or Partial History

The UI must handle several data-gap scenarios gracefully:

### No prior snapshots

When only one snapshot exists, trend views are hidden. The snapshot browser
shows the single snapshot with a message: "Run `hamlet analyze` again to
begin tracking trends."

### Snapshots with different detector sets

`SnapshotMeta.Detectors` may differ across snapshots (e.g., coverage detector
was added later). When comparing, the UI shows a notice listing detectors
that were present in one snapshot but not the other, and excludes measurements
from those detectors from trend calculations.

### Missing runtime or coverage data

`PortfolioAsset.HasRuntimeData` and `HasCoverageData` flags indicate data
availability. When runtime data is absent, the UI grays out speed-related
trends and shows "No runtime data in this snapshot." The same applies to
coverage fields when `CoverageSummary` is nil.

### Schema version mismatch

If `SnapshotMeta.SchemaVersion` differs between snapshots, the UI warns that
comparison accuracy may be reduced and limits comparison to fields present
in both schema versions.

## Performance Considerations

### Thin UI principle

The extension loads snapshot JSON files and performs no analysis beyond
client-side diffing of scalar fields and set membership checks on signal
locations. There is no re-parsing of source code, no AST analysis, and
no re-computation of posture bands.

### Snapshot size

A typical snapshot for a medium repository (500 test files, 2000 code units)
serializes to roughly 1-3 MB of JSON. The UI should:

- Load only the current snapshot eagerly.
- Load comparison snapshots on demand (when the user pins a second snapshot).
- Cache parsed snapshot data in memory for the duration of the session.
- Never load all historical snapshots simultaneously.

### Index file

An optional `.hamlet/snapshots/index.json` file can cache lightweight metadata
(timestamp, posture bands, signal counts) for each snapshot, allowing the
snapshot browser to render the list without parsing full snapshot files.

## Open Questions

- Should trend views support configurable time windows (last N snapshots vs.
  last N days)?
- Should the UI support exporting comparison summaries as markdown for
  inclusion in PR descriptions?
- How should the UI handle snapshot files that are corrupted or fail JSON
  parsing?
