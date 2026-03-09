# Impact Benchmark Safety

## Overview

`impact.Aggregate` provides privacy-safe aggregate statistics from an `ImpactResult`. It contains no raw file paths, symbol names, test IDs, or owner names — only counts, ratios, and bands suitable for cross-repository comparison.

## Privacy Threshold

```go
const PrivacyThreshold = 3
```

When the number of impacted units is below 3, breakdown fields (`ProtectionCounts`, `ConfidenceCounts`) are suppressed (set to empty maps) and `IsSparse` is set to `true`. This prevents identification of individual files or code units from small change sets.

## Aggregate Fields

### Always Present

| Field | Description |
|-------|-------------|
| `ChangedFileCount` | Number of files in scope |
| `ChangedTestFileCount` | Number of test files changed |
| `ImpactedUnitCount` | Number of impacted code units |
| `ExportedUnitCount` | Number of exported/public units |
| `GapCount` | Total protection gaps |
| `HighSeverityGapCount` | High-severity gaps |
| `ImpactedTestCount` | Total relevant tests |
| `SelectedTestCount` | Tests in protective set |
| `OwnerCount` | Distinct owners affected |
| `Posture` | Change-risk posture band |

### Above Privacy Threshold Only

| Field | Description |
|-------|-------------|
| `ProtectionCounts` | Breakdown by protection status |
| `ConfidenceCounts` | Breakdown by confidence level |
| `ProtectionRatio` | Ratio of protected units |
| `ExactConfidenceRatio` | Ratio of exact-confidence mappings |

### Optional Enrichment

| Field | Description |
|-------|-------------|
| `SelectionSetKind` | Type of protective set (exact/near_minimal/fallback_broad) |
| `GraphStats` | Impact graph quality metrics (counts only) |
| `IsSparse` | Whether data is below privacy threshold |

## What Is NOT Included

- Raw file paths
- Code unit names or IDs
- Test file paths or IDs
- Owner names
- Signal explanations or suggested actions
- Any string that could identify specific code

## GraphStats

`GraphStats` contains only aggregate counts and is always privacy-safe:
- `TotalEdges`, `ExactEdges`, `InferredEdges`, `WeakEdges`
- `ConnectedUnits`, `IsolatedUnits`, `ConnectedTests`
