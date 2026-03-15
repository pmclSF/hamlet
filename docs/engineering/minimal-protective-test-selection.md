# Minimal Protective Test Selection

## Overview

Terrain's protective test selection recommends a focused set of tests for a change, explaining why each test was included and identifying coverage gaps in the selected set. This powers `terrain select-tests` and the recommended tests section in `terrain impact` and `terrain pr`.

## ProtectiveTestSet Model

```go
type ProtectiveTestSet struct {
    Tests              []SelectedTest
    SetKind            string   // "exact", "near_minimal", "fallback_broad"
    CoveredUnitCount   int
    UncoveredUnitCount int
    Explanation        string
}

type SelectedTest struct {
    ImpactedTest
    Reasons []SelectionReason
}

type SelectionReason struct {
    Reason     string
    CodeUnitID string
    EdgeKind   string
}
```

## Selection Strategies

### Exact (`exact`)

When at least one test has `ConfidenceExact` or `IsDirectlyChanged`, the set includes only exact-confidence and directly-changed tests. This is the highest-quality selection.

### Near-Minimal (`near_minimal`)

When no exact tests exist but inferred tests are available, the set includes all inferred-confidence tests. The name "near-minimal" reflects that structural heuristics may include more tests than strictly necessary.

### Fallback Broad (`fallback_broad`)

When no impacted tests are identified at all, the set is empty and the explanation recommends running the full test suite.

## Selection Reasoning

Each `SelectedTest` carries `Reasons[]` explaining why it was included:

- `"test file directly changed"` — the test was modified in the change scope
- `"exact coverage of impacted unit"` — per-test lineage links this test to a changed unit
- `"inferred structural relationship"` — directory proximity or naming heuristic

Reasons include `CodeUnitID` and `EdgeKind` where applicable, enabling UI drill-downs.

## Coverage Gap Awareness

The protective set reports:
- `CoveredUnitCount` — impacted units that have at least one test in the selected set
- `UncoveredUnitCount` — impacted units with no covering test in the selected set

When `UncoveredUnitCount > 0`, the explanation notes the gap explicitly.

## CLI Usage

```bash
terrain select-tests --base main         # human-readable output
terrain select-tests --base main --json  # JSON ProtectiveTestSet
```

The `terrain impact --show selected` drill-down also renders the protective set.
