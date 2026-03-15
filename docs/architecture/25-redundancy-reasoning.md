# Redundancy Reasoning

> **Status:** Implemented
> **Purpose:** Detect duplicate or redundant validation in a behavior-centric way, distinguishing healthy overlap from wasteful duplication.
> **Key decisions:**
> - Redundancy is behavior-aware: tests are redundant when they exercise identical behavior surfaces, not just when they look structurally similar
> - Three overlap kinds: wasteful (same framework, same behavior), cross-framework (migration artifact), cross-level (defense in depth)
> - Cross-level overlap is classified as healthy and given low confidence — defense in depth is intentional
> - This complements (not replaces) the existing structural `DetectDuplicates()` analysis

**See also:** [02-graph-schema.md](02-graph-schema.md), [08-test-similarity-structural-fingerprints.md](08-test-similarity-structural-fingerprints.md), [14-evidence-scoring-and-confidence-model.md](14-evidence-scoring-and-confidence-model.md)

## Problem

The existing `DetectDuplicates()` analysis finds structurally similar tests using fingerprinting (shared fixtures, helpers, suite paths, assertion patterns). This catches copy-paste duplication but misses a deeper question: **do these tests validate the same behavior?**

Two tests can be structurally different (different fixtures, different assertion patterns) yet exercise identical behavior surfaces — the same API endpoints, the same domain logic, the same code paths. This is genuine redundancy that wastes CI time.

Conversely, two tests can look structurally similar yet serve different purposes:
- A unit test and an e2e test covering the same surface = defense in depth (healthy)
- A Jest test and a Vitest test covering the same surface = migration artifact (temporary)
- Two Jest tests in the same suite covering the same surface = wasteful overlap

Terrain needs to reason about **what** tests validate, not just **how** they look.

## Algorithm

### Step 1: Build Test-to-Surface Mapping

For each test node, trace the dependency path through the graph:

```
Test → TestFile (EdgeTestDefinedInFile)
     → SourceFile (EdgeImportsModule)
     → CodeSurface (EdgeBelongsToPackage, reverse)
     → BehaviorSurface (EdgeBehaviorDerivedFrom, reverse)
```

This produces a set of code surfaces and behavior surfaces that each test "exercises" — the behaviors it validates.

### Step 2: Surface-Based Candidate Generation

Build a reverse index: surface → tests that exercise it. Use this as blocking keys to generate candidate pairs — two tests are candidates if they share at least one surface.

Blocks larger than 500 members are skipped (too coarse for meaningful redundancy analysis).

### Step 3: Overlap Scoring

For each candidate pair, compute behavior overlap using Jaccard similarity on their surface sets. Pairs with < 50% overlap are discarded — partial overlap is expected and not indicative of redundancy.

### Step 4: Clustering and Classification

Pairs exceeding the overlap threshold are clustered via union-find. Each cluster is classified:

| Overlap Kind | Condition | Meaning |
|-------------|-----------|---------|
| `wasteful` | Single framework, single test level | Tests in the same framework exercise identical surfaces — consolidation candidate |
| `cross_framework` | Multiple frameworks | Tests in different frameworks (e.g., Jest + Vitest) exercise the same surfaces — migration artifact |
| `cross_level` | Multiple test types (e.g., unit + e2e) | Defense in depth — typically healthy |

Clusters are sorted: wasteful first (most actionable), then cross-framework, then cross-level.

## Relationship to DetectDuplicates

| Dimension | DetectDuplicates | AnalyzeRedundancy |
|-----------|-----------------|-------------------|
| Signal | Structural similarity | Behavior overlap |
| Inputs | Fixtures, helpers, suite paths, test names | Code surfaces, behavior surfaces, framework, test type |
| Question | "Do these tests look the same?" | "Do these tests validate the same behavior?" |
| Algorithm | Fingerprint → blocking → scoring → union-find | Surface mapping → blocking → Jaccard → union-find |
| Threshold | 60% weighted composite | 50% surface Jaccard |
| Output | `DuplicateResult` with `SimilaritySignals` | `RedundancyResult` with `OverlapKind` + `Rationale` |

Both analyses run independently. A test may appear in both a structural duplicate cluster and a behavior redundancy cluster — the two signals reinforce each other when they converge.

## Confidence Model

Confidence combines three factors:

1. **Overlap strength** (70% weight): Higher Jaccard similarity on surfaces = higher confidence
2. **Surface count boost**: ≥3 shared surfaces = +0.15, ≥1 = +0.10
3. **Kind adjustment**: Wasteful = +0.15, cross-framework = +0.10, cross-level = +0.00

Cross-level overlap gets zero kind boost because it represents intentional defense in depth, not waste.

## Integration

### `terrain analyze`

The analyze report includes a "Behavior Redundancy" section:

```
Behavior Redundancy
------------------------------------------------------------
  Redundant tests:  18 across 4 clusters
  Cross-framework:  1 cluster(s)
  [wasteful] 6 tests, 3 shared surfaces (85% confidence) [jest]
         Tests in the same framework exercise 3 identical behavior surfaces (85% overlap). Consolidation would reduce CI cost without losing coverage.
  [cross_framework] 4 tests, 2 shared surfaces (72% confidence) [jest, vitest]
         Tests in jest and vitest exercise 2 identical surfaces (80% overlap). If migrating, remove old-framework tests after migration completes.
```

### `terrain insights`

Insights surfaces behavior redundancy as optimization findings:

- **Wasteful clusters**: Medium/High severity finding under optimization category
- **Cross-framework overlaps**: Low severity finding noting migration cleanup opportunity
- Both contribute to recommendations

### JSON Output

Both `terrain analyze --json` and `terrain insights --json` include the full `RedundancyResult`:

```json
{
  "behaviorRedundancy": {
    "clusters": [
      {
        "id": "redundancy:cs:auth:Login+cs:auth:Logout+bs:auth",
        "tests": ["test:auth:1:TestLogin", "test:auth:2:TestLoginEdge"],
        "sharedSurfaces": ["bs:auth", "cs:auth:Login", "cs:auth:Logout"],
        "surfaceNames": {"bs:auth": "auth", "cs:auth:Login": "Login", "cs:auth:Logout": "Logout"},
        "confidence": 0.85,
        "overlapKind": "wasteful",
        "rationale": "Tests in the same framework exercise 3 identical behavior surfaces (100% overlap). Consolidation would reduce CI cost without losing coverage.",
        "frameworks": ["go"]
      }
    ],
    "testsAnalyzed": 50,
    "redundantTestCount": 12,
    "crossFrameworkOverlaps": 1
  }
}
```

## Implementation

| File | Purpose |
|------|---------|
| `internal/depgraph/redundancy.go` | Core algorithm: `AnalyzeRedundancy(g)` |
| `internal/depgraph/redundancy_test.go` | 14 tests covering same-behavior, cross-framework, wasteful, empty graph, determinism |
| `internal/analyze/analyze.go` | Wires `BehaviorRedundancy` into analyze report |
| `internal/insights/insights.go` | Wires `BehaviorRedundancy` into insights; adds wasteful + cross-framework findings |
| `internal/reporting/analyze_report_v2.go` | Renders "Behavior Redundancy" section |
| `internal/reporting/insights_report_v2.go` | Renders redundancy summary in insights |
| `cmd/terrain/main.go` | Calls `AnalyzeRedundancy()` in `runInsights()` |

## Scale Safety

- Hard cap at 25,000 tests (same as structural duplicate analysis)
- Surface-based blocking avoids O(n²) comparisons
- Block size capped at 500 members
- Tests with no exercised surfaces are excluded from analysis

## Future Work

- **Direct test-to-behavior edges**: Currently, test-to-surface mapping is inferred via the import graph. Adding explicit `EdgeTestExercises` edges (from runtime coverage data or assertion analysis) would improve precision.
- **Assertion-aware overlap**: Two tests may import the same source but assert different properties. Incorporating assertion targets would distinguish validation scope more precisely.
- **Portfolio integration**: Redundancy findings could feed into portfolio-level recommendations for test suite optimization.
