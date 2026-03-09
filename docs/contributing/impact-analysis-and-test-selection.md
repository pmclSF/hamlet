# Contributing: Impact Analysis and Test Selection

This guide covers how to extend the impact analysis and protective test selection subsystems in Hamlet V3.

## Package Layout

```
internal/impact/
  impact.go         # Types: ChangeScope, ImpactedCodeUnit, ImpactedTest, ProtectionGap, etc.
  analysis.go       # Core logic: mapChangedUnits, findImpactedTests, findProtectionGaps, etc.
  graph.go          # ImpactGraph: edge model, construction, query methods
  changescope.go    # Input adapters: git diff, explicit paths, CI list, snapshot comparison
  filter.go         # FilterByOwner: narrows results to a single owner
  aggregate.go      # BuildAggregate: privacy-safe export with suppression
```

## How to Add a New Protection Gap Type

1. Add gap detection in `analysis.go`, either in `findProtectionGaps()` or a new helper:

```go
func findMyNewGaps(units []ImpactedCodeUnit) []ProtectionGap {
    var gaps []ProtectionGap
    for _, iu := range units {
        if shouldFlag(iu) {
            gaps = append(gaps, ProtectionGap{
                GapType:         "my_new_gap",
                CodeUnitID:      iu.UnitID,
                Path:            iu.Path,
                Explanation:     fmt.Sprintf("factual observation about %s.", iu.Name),
                Severity:        "medium",
                SuggestedAction: fmt.Sprintf("Add tests for %s.", iu.Name),
            })
        }
    }
    return gaps
}
```

2. Call it from `findProtectionGaps()`:

```go
gaps = append(gaps, findMyNewGaps(units)...)
```

3. Add tests covering the new gap type.

## How to Add a New Change-Risk Dimension

1. Add a `computeNewDimension` function in `analysis.go`:

```go
func computeNewDimension(result *ImpactResult) ChangeRiskDimension {
    // Analyze the result and return a dimension with band and explanation.
}
```

2. Add it to the `dims` slice in `computeChangeRiskPosture()`:

```go
dims := []ChangeRiskDimension{
    computeProtectionDimension(result),
    computeExposureDimension(result),
    computeCoordinationDimension(result),
    computeInstabilityDimension(result),
    computeNewDimension(result),
}
```

3. Add tests for the new dimension with well_protected and high_risk scenarios.

## How to Add a New Scope Input Adapter

1. Add a function in `changescope.go`:

```go
func ChangeScopeFromNewSource(input MyInput) *ChangeScope {
    scope := &ChangeScope{Source: "new-source"}
    // Populate scope.ChangedFiles from input.
    return scope
}
```

2. Normalize all paths to repo-relative using `filepath.Rel()`.
3. Set `IsTestFile` using `isTestFilePath()`.
4. Add tests for empty input, single file, and multi-file scenarios.

## How to Add a New Impact Graph Edge Type

1. Add the constant in `graph.go`:

```go
const EdgeNewType EdgeKind = "new_type"
```

2. Add edge construction logic in `BuildImpactGraph()`:

```go
// Strategy N: New edge type.
for _, tf := range snap.TestFiles {
    if matchesNewCriteria(tf) {
        g.addEdge(unitID, tf.Path, EdgeNewType, ConfidenceInferred, "new_source", covType)
    }
}
```

3. Edges are deduplicated by `addEdge()` — stronger confidence upgrades existing edges.

## Testing Conventions

- Build test fixtures inline in the test file.
- Test nil/empty inputs — `Analyze(scope, nil)` and empty snapshots.
- Test deterministic output — run twice, compare results.
- Test edge cases: no code units, no tests, all weak confidence.
- Use `t.Errorf` for non-fatal assertions, `t.Fatalf` only when continuation is meaningless.

## Style Guide

### Explanations

- Factual: describe what was observed
- Specific: include names, counts, paths
- Lowercase start: sentence fragments
- Under 120 characters when possible

### Suggested Actions

- Start with a verb: Add, Remove, Investigate, Consider
- Name the specific file, unit, or test
- Use "Consider" for judgment calls, imperative for clear fixes
