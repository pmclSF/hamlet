# Stage 119 -- UI Filtering, Scoping, and Confidence Cues

Design specification for the Terrain extension UI's filtering system and
confidence communication. This document describes views and interaction
patterns, not implementation.

## Overview

Terrain signals carry structured metadata about where they apply, who owns
the affected code, how they were derived, and how confident the engine is
in the result. The filtering system lets users scope any view to a relevant
slice of findings. Confidence cues ensure users understand which findings
are grounded in strong evidence and which are heuristic inferences.

## Filter Dimensions

Every list view in the UI (signal list, cluster list, triage queue, trend
view) supports filtering by the following dimensions. Filters are additive
(AND composition) -- selecting owner `@team-auth` and dimension `coverage`
shows only coverage signals owned by team-auth.

### Owner

Filter by `Signal.Owner` or `OwnershipAssignment.Owners[].ID`. The filter
dropdown is populated from `OwnershipSummary.Owners[]` and includes an
"Unowned" option for findings where `OwnershipAssignment.IsUnowned()` is true.

### Package

Filter by `SignalLocation.Package`. The dropdown is populated from the
distinct `Package` values across all `Signal.Location` entries.

### Changed Area

When a `PRAnalysis` or `ImpactResult` is available (from `terrain pr` or
`terrain impact`), the filter can scope to only findings within the change
scope. This uses the `ImpactResult.ImpactedUnits[].UnitID` set and the
`ChangeScopedFinding` list. The UI calls `FilterByOwner` logic client-side,
matching `Signal.Location.File` against the impacted file set.

### Posture Dimension

Filter by `SignalCategory` (structure, health, quality, migration, governance)
or by the `DimensionPostureResult.Dimension` names from the measurement layer.
This maps to the Level 1 drill-down from Stage 117, but as a filter rather
than a navigation action.

### Test Type

Filter by test type when `PortfolioAsset.TestType` is populated (unit,
integration, e2e). This allows views like "show only findings affecting
e2e tests" or "show only unit test coverage gaps."

### Signal Type

Filter by specific `SignalType` values (e.g., `flakyTest`, `weakAssertion`,
`coverageBlindSpot`). Useful for targeted investigation of a single signal
category.

### Severity

Filter by `Signal.Severity` (info, low, medium, high, critical). The most
common use is filtering to high and critical to focus on urgent findings.

## Filter Composition

Filters compose with all views from Stages 117 and 118:

- **Triage queue** (Stage 117): filters narrow the finding list before
  batch triage operations. Example: triage all `weakAssertion` signals
  owned by `@team-payments` as "Add unit coverage."
- **Trend views** (Stage 118): filters scope trend comparisons. Example:
  show posture movement only for the `coverage` dimension in packages
  owned by `@team-auth`.
- **Drill-down hierarchy** (Stage 117): filters applied at Level 1 propagate
  down. Setting an owner filter at the dimension level means Levels 2-4
  only show findings for that owner.

Active filters are shown as removable chips at the top of every view.
A "Clear all" button resets to the unfiltered state.

## Confidence and Degradation Indicators

### Evidence strength badges

Every signal carries `EvidenceStrength` (strong, moderate, weak) and
`EvidenceSource` (ast, structural-pattern, path-name, runtime, coverage,
policy, codeowners). The UI renders these as visual badges:

| Evidence Strength | Badge Style          | Meaning                                |
|-------------------|----------------------|----------------------------------------|
| `strong`          | Solid filled icon    | AST-backed, coverage data, or runtime  |
| `moderate`        | Half-filled icon     | Structural pattern match with context   |
| `weak`            | Outline-only icon    | Path/name heuristic only               |

The `EvidenceSource` appears as a tooltip on hover, giving users the specific
derivation method without cluttering the default view.

### Ownership confidence

`OwnershipAssignment.Confidence` (high, medium, low, none) and
`OwnershipAssignment.Inheritance` (direct, inherited) are shown as badges
on owner labels:

| Confidence | Inheritance | Display                               |
|------------|-------------|---------------------------------------|
| high       | direct      | Owner name, no qualifier              |
| high       | inherited   | Owner name + "(inherited)" label      |
| medium     | any         | Owner name + dashed border            |
| low        | any         | Owner name + "inferred" badge         |
| none       | any         | "Unowned" with a suggestion tooltip   |

The `OwnershipAssignment.Source` and `MatchedRule` are available in a detail
popover so users can verify why a particular owner was assigned.

### Signal confidence score

`Signal.Confidence` (0.0-1.0) is rendered as a numeric percentage next to
findings. Findings below 0.5 confidence are visually dimmed (reduced opacity)
to signal uncertainty without hiding them.

### Cluster confidence

`Cluster.Confidence` from the clustering model is shown on cluster rows.
Low-confidence clusters (below 0.6) display a "tentative" badge and are
sorted below high-confidence clusters by default.

## Communicating Uncertainty Without Noise

### Progressive disclosure

The default view shows only `EvidenceStrength` badges. Users who want more
detail can expand to see `EvidenceSource`, `Confidence` score, and
`OwnershipAssignment.Source`. This three-level disclosure avoids overwhelming
users who trust the engine defaults while giving power users full transparency.

Level 1 (default): severity + evidence strength badge.
Level 2 (hover/click): confidence score + evidence source + ownership source.
Level 3 (detail panel): full `Signal` struct, `OwnershipAssignment` details,
`Cluster.Evidence`, and `MeasurementResult.Limitations[]`.

### Aggregation of weak evidence

When a view contains many weak-evidence signals, the UI shows a summary
banner: "N findings based on path/name heuristics -- consider adding
coverage data or CODEOWNERS for stronger evidence." This uses the count
of signals where `EvidenceStrength == "weak"` relative to total signals.

### Limitation callouts

`MeasurementResult.Limitations[]` and `PRAnalysis.Limitations[]` are
rendered as collapsible notes at the bottom of relevant views. These
explain data gaps (e.g., "No runtime data available -- speed measurements
are estimates") without interrupting the main finding flow.

### Data availability indicators

Views that depend on optional data sources show availability badges:

- `PortfolioAsset.HasRuntimeData` -- "Runtime" badge (green if true, gray if false)
- `PortfolioAsset.HasCoverageData` -- "Coverage" badge (green if true, gray if false)
- `PortfolioAggregates.HasRuntimeData` -- controls whether speed trends are shown

When data is unavailable, the UI hides dependent columns rather than showing
empty or misleading values.

## Filter State Persistence

Active filter combinations are saved to `.terrain/ui-state.json` per repository,
so users return to their last working context. The state includes:

- Active filter values (owner, package, dimension, severity, etc.)
- Drill-down path (which levels are expanded)
- Pinned comparison snapshot (from Stage 118)
- Sort order for list views

This file is local-only and excluded from version control.

## Open Questions

- Should filters support saved presets (e.g., "My team's critical findings")?
- Should the UI support regex-based package filtering for monorepos?
- How should the extension render findings when the snapshot has zero
  ownership data -- hide owner filters entirely or show them as disabled?
