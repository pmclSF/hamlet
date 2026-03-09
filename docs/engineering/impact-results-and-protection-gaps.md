# Impact Results and Protection Gaps

## Overview

`ImpactResult` is the complete output of `impact.Analyze()`. It aggregates impacted code units, impacted tests, protection gaps, a protective test set, the impact graph, and a change-risk posture into a single result that downstream consumers (PR analysis, CLI drill-downs, UI views) can render.

## ImpactResult Model

```go
type ImpactResult struct {
    Scope          ChangeScope
    ImpactedUnits  []ImpactedCodeUnit
    ImpactedTests  []ImpactedTest
    ProtectionGaps []ProtectionGap
    SelectedTests  []ImpactedTest       // legacy flat list
    ProtectiveSet  *ProtectiveTestSet   // enhanced set with reasoning
    Graph          *ImpactGraph
    Posture        ChangeRiskPosture
    ImpactedOwners []string
    Summary        string
    Limitations    []string
}
```

## Impacted Test Discovery

Tests are identified as relevant to a change through three strategies, applied in priority order:

1. **Coverage link** — test has a `LinkedCodeUnit` matching an impacted unit. Confidence: `exact`.
2. **Directly changed** — test file itself appears in the change scope. Confidence: `exact`.
3. **Directory proximity** — test is in the same directory tree as changed source code. Confidence: `inferred`.

Tests are sorted by confidence (exact first), then by path for determinism.

## Protection Gap Types

| Gap Type | Severity | Trigger |
|----------|----------|---------|
| `no_coverage` | medium | Non-exported unit with no test coverage |
| `untested_export` | high | Exported function/class with no test coverage |
| `weak_export_coverage` | medium | Exported unit covered only by E2E or indirect tests |
| `e2e_only_export` | medium | Exported unit with E2E coverage but no unit or integration tests |

### Gap Detection Logic

`findProtectionGaps()` iterates impacted units and checks:
- `ProtectionNone` creates `no_coverage` or `untested_export` (if exported)
- `ProtectionWeak` + exported creates `weak_export_coverage`

`findCoverageDiversityGaps()` checks `CoverageTypes`:
- E2E-only coverage on an exported unit creates `e2e_only_export`

Each gap includes:
- `Explanation` — factual description of the gap
- `SuggestedAction` — actionable remediation (starts with a verb)
- `CodeUnitID` — links back to the impacted unit

## Limitations

`identifyLimitations()` notes data gaps that affect analysis quality:

- No code units discovered — file-level impact only
- No per-test coverage lineage — structural heuristics used
- No ownership data — coordination risk underestimated

These appear in CLI output and PR comments to maintain honesty about analysis precision.
