# Coverage Attribution

Coverage attribution maps normalized coverage records onto code units to determine whether each function/method is covered and to what degree.

## UnitCoverage Model

```go
type UnitCoverage struct {
    UnitID            string   // stable code unit ID
    Name              string   // code unit name
    Path              string   // source file path
    CoveredAny        bool     // any observed execution
    LineCoveragePct   float64  // -1 if unavailable
    BranchCoveragePct float64  // -1 if unavailable
    FunctionHit       int      // -1=unknown, 0=not hit, 1=hit
    EvidenceQuality   string   // "exact", "approximate", "unavailable"
    CoveredByTypes    []string // test types covering this unit
}
```

## Attribution Logic

`AttributeToCodeUnits()` iterates over code units and matches them against merged coverage:

### 1. Function Hit Detection

If the coverage record has function-level data (`FunctionHits` map), Terrain looks up the code unit's name directly. This is the most precise attribution ("exact" evidence quality).

### 2. Line Coverage

For code units with known `StartLine`/`EndLine`, Terrain counts how many instrumented lines within that span were executed. If `EndLine` is unknown, a 50-line estimate is used.

### 3. Branch Coverage

Branch coverage is currently file-level (not scoped to individual code units). This is a known limitation documented as "approximate" evidence.

## Evidence Quality Levels

| Level | Meaning |
|-------|---------|
| `exact` | Function-level hit data available, direct match |
| `approximate` | Attribution is inferred from line overlap or file-level data |
| `unavailable` | No coverage data exists for this file/unit |

## Conservative Attribution

Terrain prefers conservative attribution over inflated claims:
- A unit is only marked `CoveredAny: true` if there is positive evidence of execution
- Missing data produces `-1` indicators, not zero
- Evidence quality is always explicit

## Coverage by Type

`ComputeByType()` extends attribution by using labeled coverage runs:

1. Group artifacts by `RunLabel` (e.g., "unit", "e2e", "integration")
2. Merge and attribute each group independently
3. For each code unit, record which test types cover it
4. Identify exclusive coverage (units covered only by one type)

### TypeCoverage Output

```go
type TypeCoverage struct {
    UnitID         string
    CoveredByTypes map[string]bool  // type → covered
    ExclusiveType  string           // single covering type, if only one
    Uncovered      bool
}
```

### Repository Summary

`BuildRepoSummary()` aggregates coverage-by-type across the repository:

- Total/exported code units
- Covered by unit tests / integration / e2e
- Covered only by e2e (risk indicator)
- Uncovered exported functions
- Top 10 risky files (most e2e-only + uncovered units)

## Limitations

- Branch attribution is file-level, not per-unit
- Line span estimates may be imprecise for complex nesting
- Function name matching depends on coverage format emitting function names
