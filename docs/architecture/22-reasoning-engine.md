# Reasoning Engine

> **Status:** Implemented (hardened)
> **Purpose:** Document how Terrain produces explainable, traceable reasoning for all findings and recommendations. Defines the five reasoning pipelines, their determinism guarantees, and explanation trace contracts.
> **Key decisions:**
> - Every finding has an explanation trace — no opaque scores or unexplained conclusions
> - Confidence is always visible, never hidden behind a simplified label
> - Reasoning is conservative: uncertainty is surfaced explicitly
> - Five reasoning pipelines, each domain-specialized with consistent determinism guarantees
> - All pipeline outputs are sorted for reproducibility

**See also:** [14-evidence-scoring-and-confidence-model.md](14-evidence-scoring-and-confidence-model.md), [13-graph-traversal-algorithms.md](13-graph-traversal-algorithms.md), [16-unified-graph-schema.md](16-unified-graph-schema.md)

## Design Principles

**Explainability over precision.** Every finding carries a reason chain. A developer can always ask "why?" and trace back to concrete evidence. Terrain never produces a score without showing how it was computed.

**Conservative under uncertainty.** When evidence is insufficient, Terrain says so rather than guessing. Fallback strategies are explicit and visible in the output.

**Deterministic output.** Given the same graph state, every pipeline produces byte-identical output. All results are sorted by documented criteria with tie-breaking by stable identifiers.

**Domain-specialized pipelines.** Each reasoning pipeline is tuned for its specific domain rather than sharing a generic framework. This keeps each pipeline simple, testable, and auditable.

## Five Reasoning Pipelines

### 1. Impact Reasoning

**Package:** `internal/depgraph/impact.go`
**Question:** Given a code change, which tests are affected and how confident are we?

**Algorithm:** BFS from changed source files through the dependency graph, propagating confidence with decay at each hop.

```
Confidence at each hop:
  newConf = current × edgeConfidence × lengthDecay × fanoutPenalty

  lengthDecay   = 0.85  (confidence drops 15% per hop)
  fanoutPenalty  = 1/log₂(outDegree + 1)  when outDegree > 5
  minConfidence = 0.1   (stop propagating below this)
  maxDepth      = 20    (BFS depth cap)
```

**Confidence bands:**
- High: ≥ 0.7 (direct import or single-hop path)
- Medium: ≥ 0.4 (2-3 hop path with good edge confidence)
- Low: ≥ 0.1 (transitive or fanout-penalized path)

**Explanation trace:**
```go
type ImpactedTest struct {
    TestID      string       // test node ID
    Confidence  float64      // composite confidence
    Level       string       // "high", "medium", "low"
    ChangedFile string       // source file that triggered this
    ReasonChain []ReasonStep // edge-by-edge path from source to test
}

type ReasonStep struct {
    From           string  // source node ID
    To             string  // target node ID
    EdgeType       string  // e.g., "imports_module"
    EdgeConfidence float64 // individual edge confidence
}
```

**Determinism:** Results sorted by confidence descending, then TestID ascending.

**Scale safety:** BFS capped at depth 20. Fanout penalty prevents explosion through high-connectivity nodes.

---

### 2. Coverage Reasoning

**Package:** `internal/depgraph/coverage.go`
**Question:** For each source file, which tests cover it and how thoroughly?

**Algorithm:** For each source file node, trace reverse edges to find covering tests via two pathways:

1. **Direct:** Test file imports source file (EdgeImportsModule)
2. **Indirect (transitive):** Another source imports this source, and a test imports that intermediate source (EdgeSourceImportsSource → EdgeImportsModule)

Tests found via both pathways are deduplicated, with direct tests taking precedence.

**Coverage bands:**
- High: ≥ 3 covering tests
- Medium: 1–2 covering tests
- Low: 0 covering tests

**Explanation trace:**
```go
type SourceCoverage struct {
    SourceID      string       // source file node ID
    Path          string       // file path
    TestCount     int          // total unique covering tests
    DirectTests   []string     // test IDs with direct import edges
    IndirectTests []string     // test IDs via transitive paths
    Band          CoverageBand // High/Medium/Low
}
```

The direct/indirect distinction is the explanation: the user can see exactly which tests cover a file directly vs. transitively.

**Determinism:** Sources sorted by TestCount ascending (worst coverage first), then SourceID. Test ID lists are sorted alphabetically.

---

### 3. Redundancy Reasoning

**Package:** `internal/depgraph/duplicate.go`
**Question:** Which tests are structurally similar enough to be redundant?

**Algorithm:** Four-stage pipeline:

1. **Fingerprinting** — Extract structural signature per test: package, suite path, normalized assertion pattern
2. **Blocking** — Generate candidate pairs using blocking keys (shared package) to avoid O(n²)
3. **Scoring** — Weighted composite similarity:
   ```
   composite = 0.25×fixtureOverlap + 0.25×helperOverlap
             + 0.30×suitePathLCS + 0.20×assertionTokenJaccard
   ```
4. **Clustering** — Union-find groups tests exceeding the 0.60 similarity threshold

**Similarity metrics:**
- Jaccard set similarity (intersection/union)
- Longest common subsequence ratio (suite path comparison)
- Token Jaccard (normalized test name overlap)

**Explanation trace:**
```go
type DuplicateCluster struct {
    ID         int              // cluster identifier
    Tests      []string         // test node IDs in cluster
    Similarity float64          // average pairwise similarity
    Signals    SimilaritySignals // per-dimension breakdown
}

type SimilaritySignals struct {
    FixtureOverlap             float64
    HelperOverlap              float64
    SuitePathSimilarity        float64
    AssertionPatternSimilarity float64
}
```

The `SimilaritySignals` breakdown shows which dimensions drove the similarity score, making the clustering decision auditable.

**Determinism:** Clusters sorted by similarity descending. Test IDs within clusters sorted alphabetically.

**Scale safety:** Skips clustering entirely when test count exceeds 25,000. Block size capped at 500.

---

### 4. Stability Reasoning

**Package:** `internal/stability/cluster.go`
**Question:** Which unstable tests share a common dependency that likely causes their instability?

**Algorithm:**

1. Collect unstable test IDs from flakyTest/unstableSuite signals
2. For each unstable test, walk 1-hop outgoing edges to find shared dependencies
3. Group unstable tests by shared dependency (require ≥2 tests sharing one)
4. Rank clusters by size and confidence

**Cause classification:**

| Cause Kind | Node Type | Base Confidence |
|------------|-----------|-----------------|
| `shared_environment` | Environment | 0.75 |
| `shared_source` | SourceFile | 0.50 |

Confidence is boosted by concentration: `base += 0.15 × (clusterSize / totalUnstable)`, capped at 1.0.

**Explanation trace:**
```go
type Cluster struct {
    ID          string    // deterministic cluster ID
    CauseKind   CauseKind // classification of shared dependency
    CauseNodeID string    // graph node ID of the shared dependency
    CauseName   string    // human-readable name
    CausePath   string    // file path if applicable
    Members     []string  // unstable test node IDs
    MemberNames []string  // human-readable test names
    Confidence  float64   // likelihood this is the root cause
    Remediation string    // suggested action
}
```

Every cluster has a cause (what), members (who), confidence (how sure), and remediation (what to do).

**Determinism:** Clusters sorted by size descending, then confidence descending, then ID ascending.

---

### 5. Environment Reasoning

**Package:** `internal/matrix/matrix.go`
**Question:** Is the test suite covering the intended environment/device matrix?

**Algorithm:**

1. Build reverse index: environment class member → test files targeting it
2. For each environment class, compute coverage ratio (covered members / total members)
3. Identify gaps (uncovered members) and concentrations (skewed coverage)
4. Generate recommendations for adding coverage to uncovered members of targeted classes

**Explanation trace:**
```go
type ClassCoverage struct {
    ClassID        string   // environment class node ID
    ClassName      string   // human-readable class name
    TotalMembers   int      // total members in class
    CoveredMembers int      // members with at least one test
    CoverageRatio  float64  // 0.0–1.0
    Members        []MemberCoverage // per-member detail
}

type CoverageGap struct {
    ClassID    string // which environment class
    ClassName  string
    MemberID   string // which member is uncovered
    MemberName string
}

type DeviceRecommendation struct {
    ClassID    string // which class to expand
    ClassName  string
    MemberID   string // which member to target
    MemberName string
    Reason     string // why this recommendation
}
```

Gaps and recommendations provide explicit "what's missing" traces.

**Determinism:** All results sorted by ClassID then MemberID.

## Explanation Flow Through the CLI

The reasoning pipelines feed into the CLI's explanation system via `internal/explain/`:

```
terrain impact     → ImpactResult.ImpactedTests[].ReasonChain
                   → explain.ExplainTest()  → TestExplanation (verdict + paths)
                   → explain.ExplainSelection() → SelectionExplanation (strategy + breakdown)

terrain insights   → DuplicateResult.Clusters[].Signals
                   → CoverageResult.Sources[].DirectTests/IndirectTests
                   → FanoutResult.FlaggedNodes[].TransitiveFanout

terrain explain    → Per-test: reason chains with confidence per hop
                   → Selection: strategy + confidence buckets + reason categories
                   → Fallback: level + reason when primary reasoning insufficient
```

`terrain explain <target>` auto-detects the target type (test file, test case, code unit, owner, finding) and produces the appropriate explanation from the underlying pipeline data.

## What Was Removed During Hardening

The `internal/reasoning/` package (13 files, ~1,300 lines) was removed. It contained extracted reasoning primitives (BFS traversal, scoring, coverage aggregation, similarity metrics, stability, fallback) that were never imported by any production code. The actual reasoning logic lives directly in the five pipeline packages listed above, where it is domain-specialized, tested, and production-proven.

Duplicate implementations that existed between `reasoning/` and the production packages:

| Duplicated Logic | `reasoning/` (removed) | Production Location |
|-----------------|------------------------|-------------------|
| BFS with confidence decay | `traverse.go` | `depgraph/impact.go` |
| Fanout penalty | `score.go` | `depgraph/impact.go` |
| Coverage aggregation | `coverage.go` | `depgraph/coverage.go` |
| Jaccard/LCS similarity | `candidates.go` | `depgraph/duplicate.go` |
| Stability aggregation | `stability.go` | `stability/cluster.go` |
| Fallback expansion | `fallback.go` | `impact/analysis.go` |
