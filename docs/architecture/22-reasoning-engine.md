# Reasoning Engine

> **Status:** Implemented (core primitives)
> **Purpose:** Define how Terrain produces explainable, traceable reasoning chains for all findings and recommendations, and document the shared reasoning primitives in `internal/reasoning/`.
> **Key decisions:**
> - Every finding must have a reason chain — no opaque scores or unexplained conclusions
> - Reason chains are serializable and available in JSON output for programmatic consumption
> - Confidence is always visible to the user, never hidden behind a simplified label
> - Reasoning is conservative: uncertainty is surfaced explicitly rather than papered over with false precision
> - Six shared primitives (reachability, scoring, coverage, candidates, stability, fallback) live in `internal/reasoning/` and are used by all analysis engines

**See also:** [14-evidence-scoring-and-confidence-model.md](14-evidence-scoring-and-confidence-model.md), [01-core-architecture.md](01-core-architecture.md), [05-insight-engine-framework.md](05-insight-engine-framework.md), [13-graph-traversal-algorithms.md](13-graph-traversal-algorithms.md)

## Problem

Most developer tools produce opaque results. A coverage tool says "78%." A linter says "error on line 42." A quality gate says "failed." None of these tell the developer why the result matters, what evidence supports it, or how confident the tool is in its conclusion.

Opacity erodes trust. When a tool flags something and the developer cannot understand why, the flag gets ignored. When a tool produces a score and the developer cannot decompose it, the score becomes meaningless ritual.

Terrain's core differentiator is explainability. Every finding — every signal, every risk score, every recommendation — carries a complete reason chain that traces back to concrete evidence in the codebase. The developer can always ask "why?" and get a substantive answer.

Additionally, the analysis engines (impact, coverage, duplicates, fanout, stability) share common traversal, scoring, and aggregation patterns. Before the reasoning package, each engine reimplemented these patterns independently, leading to divergent behavior and duplicated logic.

## Implemented: `internal/reasoning/` Package

The reasoning package provides six shared primitives that all analysis engines can use. Each primitive is designed to be composable, configurable, and tested independently.

### Package Structure

| File | Purpose | Key Exports |
|------|---------|-------------|
| `doc.go` | Package documentation | — |
| `traverse.go` | BFS reachability with confidence decay | `Reachable`, `ReachableNodes`, `TraversalConfig`, `ReachResult`, `Step` |
| `score.go` | Hop scoring, band classification, result utilities | `ScoreHop`, `FanoutPenalty`, `CompoundConfidence`, `ClassifyBand`, `TopN`, `FilterByBand`, `BandCounts` |
| `coverage.go` | Coverage aggregation (direct + indirect) | `CollectCovering`, `FindCoverageGaps`, `ClassifyCoverageBand`, `CoverageSummary` |
| `candidates.go` | Redundancy candidate generation | `BuildFingerprints`, `GenerateCandidates`, `ScoreSimilarity`, `NormalizeTestName`, `JaccardSets`, `LCSRatio` |
| `stability.go` | Stability signal aggregation | `AggregateStability`, `StabilityAdjustedConfidence`, `StabilitySummary`, `StabilitySignal` |
| `fallback.go` | Fallback expansion strategies | `ExpandFallback`, `FallbackResult`, `FallbackConfig` |

### 1. Reachability (`traverse.go`)

BFS traversal through the dependency graph with configurable direction, confidence decay, stop conditions, and edge filtering.

```go
cfg := reasoning.DefaultTraversalConfig()
cfg.Direction = "reverse"  // follow incoming edges
cfg.StopAt = func(n *depgraph.Node) bool {
    return n.Type == depgraph.NodeTest
}

results := reasoning.Reachable(g, []string{"file:src/auth.go"}, cfg)
// Returns all reachable nodes with confidence scores and reason chains
```

**Configuration:**

| Parameter | Default | Description |
|-----------|---------|-------------|
| `MaxDepth` | 20 | Maximum hops |
| `MinConfidence` | 0.1 | Prune paths below this threshold |
| `LengthDecay` | 0.85 | Per-hop decay factor |
| `FanoutThreshold` | 5 | Fanout penalty trigger |
| `Direction` | `"reverse"` | `"reverse"` or `"forward"` |
| `StopAt` | nil | Stop traversal at matching nodes |
| `EdgeFilter` | nil | Select which edges to traverse |

### 2. Path Scoring (`score.go`)

The scoring formula matches the established model from `depgraph.AnalyzeImpact`:

```
newConfidence = currentConfidence × edgeConfidence × lengthDecay × fanoutPenalty
```

Where `fanoutPenalty = 1 / log₂(outDegree + 1)` when `outDegree > fanoutThreshold`, otherwise 1.0.

**Confidence bands:**

| Band | Threshold | Meaning |
|------|-----------|---------|
| High | ≥ 0.7 | Strong evidence, direct relationship |
| Medium | ≥ 0.4 | Moderate evidence, indirect relationship |
| Low | < 0.4 | Weak evidence, transitive inference |

### 3. Coverage Aggregation (`coverage.go`)

Finds all validation nodes that cover a target through three pathways:

1. **Direct:** test file → imports → target
2. **Indirect via helpers:** helper → imports → target, test → uses helper
3. **Indirect via fixtures:** fixture → imports → target, test → uses fixture
4. **Transitive:** sourceB → imports → target, test → imports → sourceB

```go
cov := reasoning.CollectCovering(g, "file:src/auth.go")
// cov.DirectTests, cov.IndirectTests, cov.TotalCount, cov.Band
```

`FindCoverageGaps` scans all source files and returns those below a minimum band.

### 4. Redundancy Candidate Generation (`candidates.go`)

Identifies structurally similar tests using:

1. **Fingerprinting:** Extract fixtures, helpers, suite path, assertion pattern for each test
2. **Blocking keys:** Group by package, fixture, helper to avoid O(n²) comparison
3. **Weighted similarity:** Jaccard on fixtures (0.25) + helpers (0.25) + LCS on suite path (0.30) + token Jaccard on assertion pattern (0.20)
4. **Threshold:** Pairs scoring ≥ 0.60 are candidates

Scale safety: `MaxBlockSize` (500) prevents degenerate blocking keys from generating excessive pairs.

### 5. Stability Aggregation (`stability.go`)

Aggregates execution-history signals (failure rate, flaky rate) into a stability band:

| Band | Condition |
|------|-----------|
| Stable | failure rate ≤ 0.20, flaky rate ≤ 0.05 |
| Unstable | failure rate > 0.20 or flaky rate > 0.05 |
| Critical | failure rate > 0.50 |
| Unknown | No signals available |

`StabilityAdjustedConfidence` applies a multiplier based on stability band: critical (×0.5), unstable (×0.75), unknown (×0.9), stable (×1.0).

### 6. Fallback Expansion (`fallback.go`)

When primary analysis yields insufficient results, fallback strategies progressively widen the search:

1. **Package:** Find tests in the same package(s) as the seed nodes
2. **Directory:** Find tests in the same directory
3. **Family:** Find validation nodes in the same node family
4. **All:** Return all validation targets (full suite fallback)

Strategies are tried in order; the first one that produces results is used.

## Reasoning Chain Structure

A reasoning chain is a sequence of evidence steps. Each step contains:

| Field | Type | Description |
|-------|------|-------------|
| `observation` | string | What was observed in the codebase or artifacts |
| `rule` | string | What rule or heuristic was applied to the observation |
| `conclusion` | string | What conclusion was drawn from applying the rule |
| `confidence` | float | How confident this step is (0.0-1.0) |
| `evidenceType` | enum | `static-analysis`, `import-graph`, `heuristic`, `explicit`, `runtime` |
| `source` | location | File path, line number, or artifact reference |

Steps are ordered causally: each step's conclusion feeds into the next step's observation. The final step produces the finding or recommendation.

## Example Chains

### Coverage Gap Finding

```
Step 1:
  observation: "src/auth/login.ts exports function validateCredentials()"
  rule: "Exported functions are behavior surfaces and should have test coverage"
  conclusion: "validateCredentials is a coverage target"
  confidence: 0.95
  evidenceType: static-analysis
  source: src/auth/login.ts:14

Step 2:
  observation: "No test file imports src/auth/login.ts or calls validateCredentials"
  rule: "Coverage targets without direct or transitive test coverage are gaps"
  conclusion: "validateCredentials has no test coverage"
  confidence: 0.90
  evidenceType: import-graph
  source: (graph traversal — 0 covering tests found)

Step 3:
  observation: "validateCredentials is imported by src/routes/auth.ts (API route handler)"
  rule: "Behavior surfaces in API request paths have high criticality"
  conclusion: "Coverage gap severity: HIGH"
  confidence: 0.90
  evidenceType: static-analysis
  source: src/routes/auth.ts:8
```

### Recommendation: Split High-Fanout Fixture

```
Step 1:
  observation: "test/fixtures/authSession.ts is imported by 47 test files"
  rule: "Fixtures imported by more than 20 test files have high fanout"
  conclusion: "authSession is a high-fanout fixture (47 dependents)"
  confidence: 1.0
  evidenceType: import-graph
  source: (graph: 47 edges from test files to authSession.ts)

Step 2:
  observation: "authSession.ts changes trigger re-evaluation of 2,400 transitive tests"
  rule: "High transitive impact indicates CI pressure risk"
  conclusion: "Changes to authSession create disproportionate CI pressure"
  confidence: 0.95
  evidenceType: import-graph
  source: (impact analysis: 2,400 transitive dependents)

Step 3:
  observation: "authSession exports 6 functions; test files use 1-2 each on average"
  rule: "Fixtures where consumers use a small fraction of exports are candidates for splitting"
  conclusion: "Recommend splitting authSession into focused fixtures"
  confidence: 0.85
  evidenceType: static-analysis
  source: src/test/fixtures/authSession.ts (6 exports, avg 1.5 used per consumer)
```

## How `terrain explain` Works

The `explain` command assembles reason chains for a given entity:

1. **Identify the target.** Resolve the entity reference (file path, function name, signal ID) to a graph node.
2. **Collect signals.** Gather all signals that reference this entity or its transitive dependencies.
3. **Traverse dependency paths.** Walk the graph to find how this entity connects to tests, behavior surfaces, and other code units.
4. **Assemble reason chains.** For each finding or signal, construct the chain of evidence steps that produced it.
5. **Present with confidence.** Display the reason chains with confidence values and evidence types, ordered by severity.

The output is available in both human-readable (CLI) and machine-readable (JSON) formats. The JSON format preserves the full chain structure for integration with other tools.

## Composability

Reason chains compose. A high-level recommendation like "split the authSession fixture" is backed by multiple independent chains:

- **Fanout analysis chain** — documents the 47 direct dependents and 2,400 transitive dependents
- **Impact analysis chain** — documents the blast radius of changes to authSession
- **CI pressure chain** — documents how authSession changes affect CI execution time
- **Usage analysis chain** — documents the low utilization ratio (1.5 / 6 exports)

Each sub-chain is independently verifiable. A developer who trusts the fanout analysis but questions the CI pressure assessment can inspect that specific chain without wading through the others.

Composed chains reference their sub-chains by ID, creating a tree structure:

```
recommendation: "Split authSession fixture"
├── chain: fanout-analysis (confidence: 1.0)
├── chain: impact-analysis (confidence: 0.95)
├── chain: ci-pressure-assessment (confidence: 0.85)
└── chain: usage-analysis (confidence: 0.85)
overall confidence: 0.85 (minimum of sub-chain confidences)
```

## Evidence Types and Their Weight

Not all evidence is equally reliable. Terrain assigns base confidence ranges by evidence type:

| Evidence Type | Confidence Range | Description |
|--------------|-----------------|-------------|
| `static-analysis` | 0.85 - 1.0 | AST parsing, export analysis, import resolution |
| `import-graph` | 0.80 - 1.0 | Dependency relationships from resolved imports |
| `runtime` | 0.90 - 1.0 | JUnit XML, coverage data, test execution results |
| `heuristic` | 0.40 - 0.70 | Naming conventions, directory structure, co-location patterns |
| `explicit` | 1.0 | User-declared configuration in terrain.yaml |

When a reason chain combines evidence types, the overall chain confidence is bounded by its weakest link. A chain that starts with static analysis (0.95) but includes a heuristic step (0.6) produces an overall confidence no higher than 0.6.

## Confidence Propagation

Confidence flows through reason chains using the same multiplicative model as graph traversal (see [14-evidence-scoring-and-confidence-model.md](14-evidence-scoring-and-confidence-model.md)):

```
chainConfidence = step1.confidence * step2.confidence * ... * stepN.confidence
```

Longer chains naturally produce lower confidence, which reflects the reality that multi-step inferences are less certain than direct observations. This is intentional — it keeps Terrain conservative when reasoning requires many inferential leaps.

## Uncertainty Handling

When Terrain cannot reach a high-confidence conclusion, it says so explicitly:

- **Insufficient evidence:** "No test execution data available. Coverage assessment is based on static analysis only (confidence: 0.7)."
- **Conflicting signals:** "Static analysis suggests high coverage but no runtime execution data confirms it. Effective confidence reduced to 0.5."
- **Ambiguous classification:** "Function `processData` could be a data transformation (medium criticality) or an integration boundary (high criticality). Classified as medium with confidence 0.55."

Terrain never hides uncertainty behind a confident-sounding label. If the confidence is low, the output says so and explains why.

## Key Decisions

1. **Every finding has a chain.** This is non-negotiable. A finding without a reason chain is a bug, not a feature. Opaque results are worse than no results because they train users to ignore the tool.

2. **Serializable chains.** Reason chains are first-class data structures available in JSON output. This enables downstream tools (CI dashboards, PR comments, IDE integrations) to present explanations without re-deriving them.

3. **Visible confidence.** Users always see confidence values. The CLI formats them as qualitative labels (high/medium/low) backed by numeric values. The JSON output includes raw numbers. Hiding confidence would undermine the explainability promise.

4. **Conservative reasoning.** When evidence is ambiguous or insufficient, Terrain flags uncertainty rather than guessing. A false positive that a developer cannot understand or verify is more damaging than a missed finding that surfaces later with better evidence.

5. **Shared primitives.** The `internal/reasoning/` package extracts the six patterns that were previously reimplemented across engines (reachability, scoring, coverage, candidates, stability, fallback). This ensures consistent behavior and makes it easier to add new engines that reason over the graph.
