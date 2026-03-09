# Impact Graph

Stage 124 -- Bidirectional graph connecting code units to tests.

## Overview

The `ImpactGraph` is the relationship model at the center of impact analysis. It maps code units to tests and tests back to code units, with each edge carrying provenance metadata about how the relationship was established. Built by `BuildImpactGraph()` in `internal/impact/graph.go`.

## ImpactGraph Model

```go
type ImpactGraph struct {
    Edges       []ImpactEdge            // All impact relationships
    UnitToTests map[string][]string     // Code unit ID -> test file paths
    TestToUnits map[string][]string     // Test file path -> code unit IDs
    EdgeIndex   map[string]*ImpactEdge  // "sourceID->targetID" -> edge
    Stats       GraphStats              // Aggregate quality metrics
}
```

`UnitToTests`, `TestToUnits`, and `EdgeIndex` are internal indexes (excluded from JSON serialization via `json:"-"`). They are populated during construction and support the query methods.

## ImpactEdge

```go
type ImpactEdge struct {
    SourceID     string     // Code unit ID
    TargetID     string     // Test file path
    Kind         EdgeKind   // How this relationship was established
    Confidence   Confidence // Mapping quality (exact, inferred, weak)
    Provenance   string     // Data source for this edge
    CoverageType string     // Test type if known (unit, integration, e2e)
}
```

## Edge Types

| EdgeKind | Value | Confidence | Description |
|----------|-------|------------|-------------|
| Exact coverage | `"exact_coverage"` | exact | Per-test coverage lineage; `LinkedCodeUnits` matches a known unit ID |
| Bucket coverage | `"bucket_coverage"` | inferred | `LinkedCodeUnits` entry exists but does not match a known unit ID exactly |
| Structural link | `"structural_link"` | inferred | Import/export relationship between source and test |
| Directory proximity | `"directory_proximity"` | inferred | Test and source are in the same directory tree |
| Name convention | `"name_convention"` | weak | Test file name matches a code unit name (e.g., `AuthService.test.js` -> `AuthService`) |

## Graph Construction Strategies

`BuildImpactGraph()` applies strategies in priority order. Higher-confidence edges replace lower-confidence ones for the same unit-test pair via `addEdge()`:

### Strategy 1: LinkedCodeUnits (bucket or exact)

Iterates over all `TestFile` entries in the snapshot. For each `LinkedCodeUnits` entry:
- If the linked value matches a known `CodeUnit` ID (`path:name`), the edge is `EdgeExactCoverage` with `ConfidenceExact` and provenance `"exact_unit_link"`.
- Otherwise, the edge is `EdgeBucketCoverage` with `ConfidenceInferred` and provenance `"linked_code_units"`.

Coverage type is derived from the test file's framework type (unit, integration, or e2e).

### Strategy 2: Name-convention matching

For test files that have no edges after Strategy 1, the system extracts a subject name from the test file path using `extractTestSubject()`. For example, `src/__tests__/AuthService.test.js` yields `"AuthService"`. If a code unit with that name exists, a `EdgeNameConvention` edge is created with `ConfidenceWeak`.

`extractTestSubject()` handles common test file suffixes: `.test.js`, `.test.ts`, `.test.tsx`, `.spec.js`, `_test.go`, `_test.py`, and others.

### Edge Deduplication

When an edge already exists for a unit-test pair, `addEdge()` compares confidence levels. If the new edge has higher confidence (lower `confidenceOrder()`), the existing edge is upgraded in place. This ensures each pair has exactly one edge representing the strongest known relationship.

## GraphStats

```go
type GraphStats struct {
    TotalEdges     int  // Total number of edges
    ExactEdges     int  // Edges with exact confidence
    InferredEdges  int  // Edges with inferred confidence
    WeakEdges      int  // Edges with weak confidence
    ConnectedUnits int  // Code units with at least one test edge
    IsolatedUnits  int  // Code units with no test edges
    ConnectedTests int  // Tests connected to at least one code unit
}
```

`IsolatedUnits` is computed as `len(snap.CodeUnits) - ConnectedUnits`. A high isolation count signals poor coverage lineage data.

## Query Methods

### TestsForUnit

```go
func (g *ImpactGraph) TestsForUnit(unitID string) []string
```

Returns test paths covering a code unit, sorted by edge confidence (exact first, then inferred, then weak). Within the same confidence tier, paths are sorted alphabetically.

### UnitsForTest

```go
func (g *ImpactGraph) UnitsForTest(testPath string) []string
```

Returns code unit IDs covered by a test, sorted alphabetically.

### EdgeBetween

```go
func (g *ImpactGraph) EdgeBetween(unitID, testPath string) *ImpactEdge
```

Returns the edge between a specific unit and test, or `nil` if no relationship exists. Uses the `EdgeIndex` for O(1) lookup via the key `"unitID->testPath"`.

### EdgesForUnit

```go
func (g *ImpactGraph) EdgesForUnit(unitID string) []ImpactEdge
```

Returns all edges where the unit is the source. This is a linear scan over `Edges` (the index maps unit to test paths, not to edge structs).

## Determinism

Edges are sorted by `(SourceID, TargetID)` after construction. `ChangeScopeFromComparison` also sorts its output by path. This ensures identical inputs always produce identical graph output, which matters for snapshot testing and reproducible CI results.
