# Changed Code Unit Mapping

Stage 123 -- How changed files map to impacted code units.

## Overview

When a file changes, Hamlet needs to determine which code units inside that file are affected. The `mapChangedUnits()` function in `internal/impact/analysis.go` performs this mapping by joining `ChangeScope.ChangedFiles` against the snapshot's `CodeUnit` inventory using file path matching.

## Mapping Strategy

The algorithm is straightforward:

1. Build an index of `models.CodeUnit` entries keyed by `Path`.
2. For each `ChangedFile` in the scope (skipping test files):
   - Look up code units at that path.
   - If found, create one `ImpactedCodeUnit` per code unit in the file.
   - If not found, create a single file-level fallback entry.

```go
units, found := unitsByFile[cf.Path]
if !found {
    // File-level fallback
    impacted = append(impacted, ImpactedCodeUnit{
        UnitID:           cf.Path,
        Name:             filepath.Base(cf.Path),
        ImpactConfidence: ConfidenceWeak,
        ProtectionStatus: classifyProtection(cf.Path, snap),
    })
}
```

## ImpactedCodeUnit Model

```go
type ImpactedCodeUnit struct {
    UnitID           string           // Stable ID: "path:name" or file path for fallbacks
    Name             string           // Human-readable name
    Path             string           // File containing the code unit
    Kind             string           // function, method, class, module (optional)
    ChangeKind       ChangeKind       // How the unit was affected (added/modified/deleted/renamed)
    Exported         bool             // Whether the unit is publicly visible
    Owner            string           // Resolved owner from snapshot ownership map
    ImpactConfidence Confidence       // Mapping quality: exact, inferred, weak
    ProtectionStatus ProtectionStatus // Test coverage classification
    CoveringTests    []string         // Test file paths that cover this unit
    CoverageTypes    *CoverageTypeInfo // Mix of unit/integration/E2E coverage
    Complexity       float64          // Cyclomatic complexity if known
}
```

The `UnitID` follows the convention `"<path>:<name>"` for parsed code units (e.g., `"src/auth.js:AuthService"`). For file-level fallbacks, the `UnitID` is the file path itself.

## File-Level Fallback

When the snapshot contains no `CodeUnit` entries for a changed file, the system creates a single `ImpactedCodeUnit` representing the entire file:

- `UnitID` and `Path` are set to the file path.
- `Name` is the base filename (e.g., `config.js`).
- `ImpactConfidence` is `ConfidenceWeak` because no structural data is available.
- `ProtectionStatus` is determined by `classifyProtection()`, which checks whether any test file has the path in its `LinkedCodeUnits`.

This ensures every changed source file appears in the impact result, even without code-unit parsing. The `Limitations` field on the result will note "No code units discovered; impact analysis is file-level only."

## Protection Status Classification

Each impacted code unit receives a `ProtectionStatus` via `classifyUnitProtection()`:

| Status | Criteria |
|--------|----------|
| `strong` | Covered by a unit or integration test framework |
| `partial` | Some coverage exists but with gaps |
| `weak` | Covered only by E2E or indirect tests |
| `none` | No observed coverage from any test file |

Classification works by scanning all `TestFile` entries in the snapshot for `LinkedCodeUnits` that match either the full `UnitID` or the unit `Name`. The framework type of matching test files determines the coverage tier.

## Coverage Type Info

`CoverageTypeInfo` captures the diversity of test coverage for a code unit:

```go
type CoverageTypeInfo struct {
    HasUnitCoverage        bool
    HasIntegrationCoverage bool
    HasE2ECoverage         bool
}
```

This is computed by `classifyCoverageTypes()`, which maps each covering test file's framework to a framework type via `models.FrameworkType`. Coverage diversity matters because a unit covered only by E2E tests has slower feedback and is more fragile than one with unit-level coverage.

## Confidence Metadata

Confidence is assigned based on the change kind:

- **Exact** -- `ChangeAdded` or `ChangeDeleted`. The entire file is definitively new or removed, so all units are certainly affected.
- **Inferred** -- `ChangeModified` or `ChangeRenamed` with known code units. The file changed, so units are likely affected, but without line-level diffing this is an inference.
- **Weak** -- File-level fallback when no code units are parsed for the file.

## Owner Resolution

If the snapshot includes an `Ownership` map (`map[string][]string` keyed by file path), the first owner entry for the unit's path is assigned to `ImpactedCodeUnit.Owner`. This feeds into coordination risk assessment downstream.
