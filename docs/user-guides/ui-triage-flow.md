# Stage 117 -- UI Drill-Down Hierarchy and Triage Queue

Design specification for the Terrain extension UI's primary navigation and triage workflow.
This document describes views, not implementation. No Go code is introduced.

## Overview

The Terrain UI surfaces engine findings through a drill-down hierarchy that mirrors
how engineers reason about test risk: start with a posture dimension, narrow to a
cluster of related findings, identify the responsible owner or package, then inspect
individual test cases or code units. A companion triage queue lets users categorize
findings into actionable buckets without leaving the extension.

## Drill-Down Hierarchy

The hierarchy has four levels. Each level is a view backed by data already present
in `TestSuiteSnapshot`.

### Level 1 -- Posture Dimension

The entry point is the set of `DimensionPostureResult` entries from
`Measurements.Posture`. Each dimension (reliability, speed, coverage, governance,
migration) is rendered as a card showing its `Band`, `Explanation`, and the count
of signals and risk surfaces that fall under it.

Clicking a dimension opens Level 2.

### Level 2 -- Finding Cluster

Within a dimension, findings are grouped into clusters. Two cluster sources exist:

1. **Engine clusters** from the `clustering.Cluster` model -- these represent
   common-cause groups (e.g., `dominant_flaky_fixture`, `shared_import_dependency`).
   Each cluster has `CausePath`, `AffectedTests`, `Confidence`, and `Explanation`.
2. **Signal groups** -- signals from `TestSuiteSnapshot.Signals` grouped by
   `SignalType` within the selected dimension's category. For example, all
   `flakyTest` signals form one group, all `weakAssertion` signals form another.

Each cluster row shows: signal type or cluster type, affected test count,
severity distribution, and the top contributing owner.

Clicking a cluster opens Level 3.

### Level 3 -- Owner / Package

Within a cluster, findings are broken down by owner (from `Signal.Owner` and
`OwnershipAssignment`) or by `SignalLocation.Package`. This view uses the
`OwnerAggregate` model to show per-owner counts: `SignalCount`,
`CriticalSignalCount`, `HealthSignalCount`, and `UncoveredExportedCount`.

When ownership data is missing (confidence `none`), the view groups by package
path and renders an "unowned" badge with a suggestion to configure CODEOWNERS
or `.terrain/ownership.yaml`.

Clicking an owner row opens Level 4.

### Level 4 -- Test / Code Unit

The leaf level lists individual `Signal` entries filtered to the selected owner
and cluster. Each row shows:

- `Signal.Location` (file, symbol, line)
- `Signal.Severity` and `Signal.Confidence`
- `Signal.EvidenceStrength` (strong / moderate / weak)
- `Signal.Explanation` and `Signal.SuggestedAction`

Linked data is shown inline when available: the `TestCase` entry (with stable
`TestID` for longitudinal tracking), `CodeUnit` coverage status, and
`PortfolioAsset` cost and breadth classification.

## Triage Queue

The triage queue is a persistent categorization layer on top of the drill-down.
Users assign each finding (or batch of findings) to one of five buckets:

| Bucket              | Intent                                             |
|---------------------|----------------------------------------------------|
| Fix now             | Immediate remediation, blocks merge or release      |
| Investigate         | Needs more context before deciding action           |
| Assign ownership    | Finding is valid but no owner -- route to a team    |
| Add unit coverage   | Code unit is untested or only covered by E2E        |
| Review redundancy   | Portfolio finding suggests overlap or low-value test |

### Data model

Triage state is stored locally in `.terrain/triage.json` as an array of entries:

```
{ "signalType": SignalType, "locationKey": string, "bucket": string,
  "assignee": string?, "note": string?, "snapshotTimestamp": string }
```

The `locationKey` is derived from `SignalLocation` (repo + package + file + symbol).
This allows triage decisions to survive across snapshots as long as the signal
location is stable.

### Batch operations

From Level 2 (cluster view) or Level 3 (owner view), users can select all
visible findings and assign them to a bucket in one action. This is critical
for large repositories where individual triage is impractical.

## Relationship to Engine Logic

The UI does not recompute posture, clustering, or ownership. It reads the
serialized `TestSuiteSnapshot` and applies client-side filtering only. Key
principles:

- **Posture bands** come from `DimensionPostureResult.Band` -- the UI renders
  them, never recalculates them.
- **Clustering** comes from `ClusterResult.Clusters` -- the UI groups and
  sorts, never re-runs the clustering algorithm.
- **Ownership routing** comes from `OwnershipAssignment` -- the UI filters by
  `Owner.ID`, never re-resolves CODEOWNERS.
- **Signal severity and confidence** are engine-computed. The UI may sort and
  filter by these fields but does not adjust them.

The only mutable state the UI introduces is the triage queue itself, stored
separately from snapshot data.

## Wireframe Descriptions

### Dashboard (Level 1)

Five horizontal cards, one per posture dimension. Each card shows the dimension
name, band (color-coded: green/yellow/orange/red), driving measurement IDs, and
a sparkline of signal count by severity. A sidebar shows the triage queue summary
(counts per bucket).

### Cluster List (Level 2)

A table with columns: cluster name or signal type, affected count, severity
breakdown (as small colored dots), top owner, and a triage status indicator.
Sortable by any column. A filter bar at top allows scoping by owner or package.

### Owner Breakdown (Level 3)

A grouped list: each owner section shows their `OwnerAggregate` stats and a
nested list of signals. Unowned findings appear in a separate section at the
bottom with a call-to-action to assign ownership.

### Finding Detail (Level 4)

A detail panel showing the full `Signal` struct: location as a clickable file
link, explanation, suggested action, evidence strength badge, and linked
test case / code unit metadata. A triage dropdown lets the user assign the
finding to a bucket directly from this view.

## Open Questions

- Should triage state sync across team members via a shared file in the repo?
- Should the triage queue support custom buckets beyond the five defaults?
- How should the UI handle snapshots with no clustering data (engine ran without
  runtime artifacts)?
