# Impact Analysis Framework

Stage 121 -- Core framework for change-scoped test intelligence.

## What Impact Analysis Is

Impact analysis answers: "If this code changes, which tests matter, what protection exists, and where are the gaps?" It operates on a specific change (a PR, a commit range, or an explicit file list) rather than the repository as a whole.

The entry point is `Analyze()` in `internal/impact/impact.go`. It accepts a `*ChangeScope` and a `*models.TestSuiteSnapshot` and returns a fully populated `*ImpactResult`.

## Why It Matters

Without impact analysis, teams either run every test on every change (slow, expensive) or guess which tests are relevant (risky). Impact analysis provides evidence-based test selection, surfaces protection gaps before merge, and quantifies change risk with confidence metadata so reviewers know how much to trust the assessment.

## Core Concepts

| Concept | Type | Description |
|---------|------|-------------|
| Changed entity | `ChangedFile` | A file that was added, modified, deleted, or renamed |
| Impacted code unit | `ImpactedCodeUnit` | A code unit (function, class, module) affected by the change |
| Impacted test | `ImpactedTest` | A test relevant to the change, with relevance reason |
| Protective coverage | `ProtectionStatus` | How well a code unit is covered: strong, partial, weak, none |
| Protection gap | `ProtectionGap` | Where changed code lacks adequate test coverage |
| Impacted owner area | `ImpactedOwners` on `ImpactResult` | Teams/individuals owning impacted code |
| Confidence | `Confidence` | Mapping quality: exact, inferred, or weak |
| Evidence | `ImpactEdge.Provenance` | Data source that established a relationship |

## Where Impact Analysis Sits

Impact analysis is a consumer of lower-level subsystems, not a replacement:

- **CodeUnit inventory** (`models.CodeUnit`) -- provides the units to map against changed files.
- **Test files** (`models.TestFile`) -- provides test paths, frameworks, and `LinkedCodeUnits`.
- **Coverage lineage** -- `LinkedCodeUnits` on test files establishes exact or bucket-level coverage links.
- **Lifecycle continuity** -- test identity and lifecycle states feed into instability scoring.
- **Posture/measurement** -- repo-wide posture computes health across the entire repo; impact analysis scopes that to a single change.

## How It Differs from Other Subsystems

| Subsystem | Scope | Purpose |
|-----------|-------|---------|
| Repo-wide posture | Entire repository | Overall health assessment across all code and tests |
| Impact analysis | Specific change | Change-scoped protection assessment and test selection |
| PR/change-scoped reporting | Specific PR | Renders impact results into human-readable review output |
| Portfolio intelligence | Cross-repo | Aggregate organizational view across multiple repositories |

## Data Flow

```
ChangeScope  -->  Analyze()  -->  ImpactResult  -->  (PR reporting / formatted output)
     |                |
     |                +-- BuildImpactGraph(snap)      --> ImpactGraph
     |                +-- mapChangedUnits(scope, snap) --> []ImpactedCodeUnit
     |                +-- findImpactedTests(...)       --> []ImpactedTest
     |                +-- findProtectionGaps(...)      --> []ProtectionGap
     |                +-- buildProtectiveSet(...)      --> *ProtectiveTestSet
     |                +-- computeChangeRiskPosture(..) --> ChangeRiskPosture
     |                +-- collectOwners(...)           --> []string
     |                +-- buildImpactSummary(...)      --> string
     |                +-- identifyLimitations(...)     --> []string
```

The `ImpactResult` is the single output artifact. Downstream consumers (PR comments, CLI output, benchmark aggregation) read from it without re-running analysis.

## The ImpactGraph Model

`ImpactGraph` is a bidirectional map connecting code units to tests. Each relationship is an `ImpactEdge` with:

- `Kind` -- how the edge was established (exact_coverage, bucket_coverage, structural_link, directory_proximity, name_convention)
- `Confidence` -- exact, inferred, or weak
- `Provenance` -- the data source (e.g., `"exact_unit_link"`, `"linked_code_units"`, `"name_convention"`)

The graph is constructed by `BuildImpactGraph()` using prioritized strategies. Higher-confidence edges replace lower-confidence ones for the same unit-test pair. Edges are sorted deterministically by source and target ID.

Internal indexes (`UnitToTests`, `TestToUnits`, `EdgeIndex`) support fast lookups:

```go
tests := graph.TestsForUnit("src/auth.js:AuthService")
units := graph.UnitsForTest("test/auth.test.js")
edge  := graph.EdgeBetween("src/auth.js:AuthService", "test/auth.test.js")
edges := graph.EdgesForUnit("src/auth.js:AuthService")
```

## Limitations

`Analyze()` automatically identifies data gaps and reports them in `ImpactResult.Limitations`. Common limitations include missing code unit inventory (file-level only analysis), absent coverage lineage (heuristic-only test selection), and missing ownership data (coordination risk underestimated).
