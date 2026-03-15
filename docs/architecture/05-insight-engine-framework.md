# Insight Engine Framework

> **Status:** Implemented
> **Purpose:** How insight engines traverse the dependency graph to produce structured analysis results.
> **Key decisions:**
> - Engines are pure functions: receive a `*Graph`, return a typed result struct, no side effects
> - Each engine answers exactly one structural question about the test system
> - Scale-safety: engines skip analysis and return a `Skipped`/`SkipReason` when input exceeds safe thresholds
> - New engines are self-contained Go files in `internal/depgraph/` with no cross-engine dependencies

Insight engines are the analysis modules that traverse the dependency graph to produce structured results. Each engine answers a specific question about the test system.

## Engine Contract

Every insight engine follows the same pattern:

1. **Input:** The dependency graph (and optionally additional parameters)
2. **Processing:** Graph traversal using the appropriate algorithm
3. **Output:** A typed result struct

```go
func AnalyzeX(g *Graph) XResult
```

Engines are pure functions over the graph. They do not modify the graph or produce side effects. See [02-graph-schema.md](02-graph-schema.md) for the graph data model.

## Current Engines

### Coverage Engine

**Question:** For each source file, which tests cover it?

**Algorithm:** Reverse traversal from source file nodes. For each source file, find all test files that import it (directly or transitively through fixtures and helpers), then resolve individual test IDs within those files by walking the `test_defined_in_file` hierarchy.

**Output:**
```go
// CoverageResult from internal/depgraph/coverage.go
type CoverageResult struct {
    Sources     []SourceCoverage    `json:"sources"`     // Per-file coverage, sorted by TestCount ascending
    SourceCount int                 `json:"sourceCount"` // Total source files analyzed
    BandCounts  map[CoverageBand]int `json:"bandCounts"` // Counts by band (High, Medium, Low)
}

type SourceCoverage struct {
    SourceID      string       `json:"sourceId"`      // Source file node ID
    Path          string       `json:"path"`          // File path (without "file:" prefix)
    TestCount     int          `json:"testCount"`     // Total unique test count (direct + indirect)
    DirectTests   []string     `json:"directTests"`   // Tests that directly import this source
    IndirectTests []string     `json:"indirectTests"` // Tests reaching this source transitively
    Band          CoverageBand `json:"band"`          // Coverage classification
}
```

Coverage bands: High (3+ tests), Medium (1-2 tests), Low (0 tests).

### Duplicate Engine

**Question:** Which tests are structurally similar enough to be redundant?

**Algorithm:** Structural fingerprinting with blocking-key candidate generation. For each test, compute similarity signals (fixture overlap, helper overlap, suite path similarity, assertion pattern similarity), generate candidate pairs via shared blocking keys (package, fixture, helper), score with weighted Jaccard + LCS similarity, and cluster pairs exceeding the threshold using union-find.

**Output:**
```go
// DuplicateResult from internal/depgraph/duplicate.go
type DuplicateResult struct {
    Clusters       []DuplicateCluster `json:"clusters"`       // Groups of similar tests
    TestsAnalyzed  int                `json:"testsAnalyzed"`  // Total tests analyzed
    DuplicateCount int                `json:"duplicateCount"` // Tests flagged as duplicates
    Skipped        bool               `json:"skipped,omitempty"`    // True if skipped for scale-safety
    SkipReason     string             `json:"skipReason,omitempty"` // Reason when skipped
}

type DuplicateCluster struct {
    ID         int              `json:"id"`         // Cluster identifier
    Tests      []string         `json:"tests"`      // Test node IDs in this cluster
    Similarity float64          `json:"similarity"` // Overall similarity score (0-1)
    Signals    SimilaritySignals `json:"signals"`   // Breakdown of signal contributions
}
```

See [08-test-similarity-structural-fingerprints.md](08-test-similarity-structural-fingerprints.md) for the fingerprinting algorithm.

### Fanout Engine

**Question:** Which nodes have excessive transitive dependencies?

**Algorithm:** Reverse-topological traversal to compute transitive reachability for all nodes in a single O(n+e) pass. For DAGs this is exact; for graphs with cycles, it falls back to per-node BFS only for cycle members.

**Output:**
```go
// FanoutResult from internal/depgraph/fanout.go
type FanoutResult struct {
    Entries      []FanoutEntry `json:"entries"`      // Per-node fanout metrics, sorted by TransitiveFanout descending
    NodeCount    int           `json:"nodeCount"`    // Total nodes analyzed
    FlaggedCount int           `json:"flaggedCount"` // Nodes exceeding the threshold
    Threshold    int           `json:"threshold"`    // Threshold used for flagging
    Skipped      bool          `json:"skipped,omitempty"`    // True if skipped for scale-safety
    SkipReason   string        `json:"skipReason,omitempty"` // Reason when skipped
}

type FanoutEntry struct {
    NodeID           string `json:"nodeId"`           // Node ID
    NodeType         string `json:"nodeType"`         // Node type
    Path             string `json:"path,omitempty"`   // File path (if available)
    Fanout           int    `json:"fanout"`           // Direct outgoing dependency count
    TransitiveFanout int    `json:"transitiveFanout"` // Transitive dependency count (BFS)
    Flagged          bool   `json:"flagged"`          // Whether this node exceeds the threshold
}
```

### Impact Engine

**Question:** Given a code change, which tests are affected?

**Algorithm:** Identify changed files from git diff, then BFS via reverse edges from changed nodes to find all reachable test nodes. Confidence decays with path length (0.85 per hop) and is penalized for high-fanout intermediate nodes (divided by log2(fanout+1) when fanout > 5). Stops at test nodes or when confidence drops below 0.1.

**Output:**
```go
// ImpactResult from internal/depgraph/impact.go
type ImpactResult struct {
    ChangedFiles  []string         `json:"changedFiles"`  // Changed files that were analyzed
    Tests         []ImpactedTest   `json:"tests"`         // All impacted tests, sorted by confidence descending
    SelectedTests []string         `json:"selectedTests"` // Unique test IDs selected for re-run
    LevelCounts   map[string]int   `json:"levelCounts"`   // Summary counts by confidence level
}

type ImpactedTest struct {
    TestID      string       `json:"testId"`      // Test node ID
    Confidence  float64      `json:"confidence"`  // Composite confidence score (0-1)
    Level       string       `json:"level"`       // Classification: "high", "medium", "low"
    ChangedFile string       `json:"changedFile"` // The changed file that triggered this impact
    ReasonChain []ReasonStep `json:"reasonChain"` // Full chain of edges from changed file to test
}

type ReasonStep struct {
    From           string  `json:"from"`           // Source node in this hop
    To             string  `json:"to"`             // Target node in this hop
    EdgeType       string  `json:"edgeType"`       // Edge type for this hop
    EdgeConfidence float64 `json:"edgeConfidence"` // Edge confidence for this hop
}
```

## Adding a New Engine

1. Create a new analysis function in `internal/depgraph/` (e.g., `internal/depgraph/myengine.go`)
2. Define the result struct with JSON tags
3. Implement the analysis function taking `*Graph` as input
4. Add tests in `internal/depgraph/myengine_test.go`
5. Wire into the CLI in `cmd/terrain/main.go`

Engines should be self-contained. They receive the graph and return a result. They should not depend on other engines' results unless explicitly composed in the CLI layer.

## Related Documents

- [02-graph-schema.md](02-graph-schema.md) — Node types, edge types, and graph operations used by engines
- [13-graph-traversal-algorithms.md](13-graph-traversal-algorithms.md) — BFS, reverse reachability, and path tracing algorithms
