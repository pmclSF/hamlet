# Stability Reasoning

> **Status:** Implemented
> **Purpose:** Detect clusters of unstable tests that share common dependencies, surfacing likely root causes and remediation suggestions.
> **Key decisions:**
> - Clustering is deterministic — same graph + signals always produces the same clusters
> - Clusters are derived from graph structure, not heuristics on test names or file proximity
> - Cause attribution is ranked by dependency type (fixtures > helpers > external services > environments > resources > source files)
> - Stability clustering is additive — it enhances existing flaky/unstable findings without replacing them

**See also:** [02-graph-schema.md](02-graph-schema.md), [14-evidence-scoring-and-confidence-model.md](14-evidence-scoring-and-confidence-model.md), [15-edge-case-handling.md](15-edge-case-handling.md)

## Problem

When a test suite has many flaky or unstable tests, the signal count alone is overwhelming. Knowing "47 flaky tests detected" is not actionable. Teams need to know *why* tests are flaky and *where to start fixing*.

In many codebases, flaky tests cluster around shared infrastructure: a database fixture with teardown gaps, a helper with non-deterministic behavior, an external service dependency that times out intermittently. If Terrain can identify these shared dependencies among unstable tests, it can surface the likely root cause and recommend targeted fixes.

## Algorithm

### Step 1: Identify Unstable Tests

Collect all `flakyTest` and `unstableSuite` signals from the snapshot. Extract file paths from signal locations and map them to test node IDs in the dependency graph.

Both individual test nodes (`NodeTest`) and test file nodes (`NodeTestFile`) are resolved — if a signal targets a file, all tests defined in that file are included.

### Step 2: Collect Shared Dependencies

For each unstable test node, walk outgoing edges (and outgoing edges from its test file) to find infrastructure dependencies:

| Edge Type | Target Node Type | Cause Kind |
|-----------|-----------------|------------|
| `test_uses_fixture` | `fixture` | `shared_fixture` |
| `test_uses_helper` | `helper`, `assertion_helper` | `shared_helper` |
| `depends_on_service` | `external_service` | `external_service` |
| `targets_environment` | `environment` | `shared_environment` |
| `uses_resource` | `resource` | `shared_resource` |
| `imports_module` | `source_file` | `shared_source` |

The walk is one hop from each test/test-file node — deep transitive closure is not performed, keeping the algorithm linear in graph size.

### Step 3: Form Clusters

A cluster forms when **two or more unstable tests** share the same dependency node. Each cluster records:

- **Cause kind** — the type of shared dependency
- **Cause node** — the specific fixture, helper, service, etc.
- **Members** — the unstable test node IDs that share this dependency
- **Confidence** — how likely this shared dependency is the root cause
- **Remediation** — a targeted suggestion for stabilization

### Step 4: Rank and Deduplicate

Clusters are sorted by size (largest first), then confidence (highest first), then ID (for determinism). Tests may appear in multiple clusters — this is intentional, as a test may be unstable due to multiple shared dependencies.

## Confidence Model

Confidence combines two factors:

1. **Base confidence by cause kind** — some dependency types are more likely to cause flakiness:
   - External service: 0.85 (network dependencies are a primary flake source)
   - Fixture: 0.80 (shared mutable state is a classic flake cause)
   - Environment: 0.75 (resource contention and timing issues)
   - Helper / Resource: 0.70
   - Source file: 0.50 (correlation, not necessarily causation)

2. **Concentration boost** — what fraction of all unstable tests share this dependency:
   - `base + 0.15 × (cluster_size / total_unstable)`
   - Capped at 1.0

A fixture shared by 8 of 10 unstable tests gets higher confidence than one shared by 2 of 10.

## Remediation Suggestions

Each cause kind generates a targeted suggestion:

| Cause Kind | Remediation Pattern |
|-----------|-------------------|
| `shared_fixture` | Audit for shared mutable state, timing dependencies, or teardown gaps |
| `shared_helper` | Review for side effects or non-deterministic behavior |
| `external_service` | Consider stubbing or adding retry/circuit-breaker logic |
| `shared_environment` | Consider test isolation or dedicated environments |
| `shared_resource` | Ensure proper cleanup and per-test isolation |
| `shared_source` | Investigate recent changes for introduced non-determinism |

## Integration

### `terrain analyze`

The analyze report includes a "Stability Clusters" section when clusters are detected:

```
Stability Clusters
------------------------------------------------------------
  Unstable tests:  12 (9 clustered around shared dependencies)
  [shared_fixture] db-setup  (6 tests, 87% confidence)
         Audit fixture "db-setup" for shared mutable state, timing dependencies, or teardown gaps.
  [external_service] payments-api  (4 tests, 91% confidence)
         External service dependency "payments-api" is a likely flake source. Consider stubbing or adding retry/circuit-breaker logic.
```

### `terrain insights`

Insights surfaces cluster findings as reliability issues. When clusters are detected, a dedicated finding is raised:

```
Reliability Problems (2)
------------------------------------------------------------
  [HIGH] 12 flaky/unstable test signals detected
         Flaky tests erode developer trust and waste CI cycles on retries.
  [MEDIUM] 3 stability clusters detected — likely shared root causes
         9 of 12 unstable tests cluster around shared dependencies. Top cause: db-setup (shared_fixture, 6 tests).
```

The insights report also includes a "Stability Clusters" section with the top 3 clusters and their remediation suggestions.

### JSON Output

Both `terrain analyze --json` and `terrain insights --json` include the full `ClusterResult` in their output:

```json
{
  "stabilityClusters": {
    "clusters": [
      {
        "id": "stability:shared_fixture:fixture:db-setup",
        "causeKind": "shared_fixture",
        "causeNodeId": "fixture:db-setup",
        "causeName": "db-setup",
        "causePath": "test/fixtures/db.js",
        "members": ["test:a:1:testA", "test:b:1:testB", ...],
        "memberNames": ["testA", "testB", ...],
        "confidence": 0.87,
        "remediation": "Audit fixture \"db-setup\" for shared mutable state, timing dependencies, or teardown gaps."
      }
    ],
    "unstableTestCount": 12,
    "clusteredTestCount": 9
  }
}
```

## Implementation

| File | Purpose |
|------|---------|
| `internal/stability/cluster.go` | Core clustering algorithm: `DetectClusters(g, signals)` |
| `internal/stability/cluster_test.go` | 12 tests covering shared fixtures, external services, multiple clusters, no-cluster cases, determinism |
| `internal/analyze/analyze.go` | Wires `StabilityClusters` into analyze report |
| `internal/insights/insights.go` | Wires `StabilityClusters` into insights report + cluster findings |
| `internal/reporting/analyze_report_v2.go` | Renders "Stability Clusters" section in analyze output |
| `internal/reporting/insights_report_v2.go` | Renders "Stability Clusters" section in insights output |
| `cmd/terrain/main.go` | Calls `stability.DetectClusters()` in `runInsights()` |

## Relationship to Existing Stability Analysis

This feature complements but does not replace existing stability mechanisms:

- **Health detectors** (`internal/health/`) detect individual flaky/unstable signals from runtime data — these are the *inputs* to clustering
- **Stability classification** (`internal/stability/classify.go`) classifies longitudinal patterns per test — this is *historical* analysis
- **Stability clustering** (`internal/stability/cluster.go`) groups currently-unstable tests by shared graph structure — this is *structural* analysis that identifies *why* tests are flaky

The three work together: detectors find flaky tests, classification tracks their history, and clustering explains their shared causes.
