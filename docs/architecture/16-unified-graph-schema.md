# Unified Graph Schema

> **Status:** Implemented
> **Purpose:** Describe the converged graph schema that unifies the snapshot index (`internal/graph/`), typed dependency graph (`internal/depgraph/`), and impact graph (`internal/impact/`) into a single traversable structure.
> **Key decisions:**
> - Six node families (system, validation, behavior, environment, execution, governance) implemented in `internal/depgraph/`
> - All node types, edge types, and evidence types defined as constants with JSON serialization support
> - `NodeFamily` grouping enables family-level queries (`NodesByFamily`) alongside type-level queries
> - Graph supports `MarshalJSON`/`UnmarshalJSON` for deterministic serialization and deserialization
> - Three existing graph representations converge incrementally — unified types are in place; construction pipelines converge next
> - Confidence model applies uniformly across all edge types, including cross-family edges

**See also:** [02-graph-schema.md](02-graph-schema.md), [05-insight-engine-framework.md](05-insight-engine-framework.md), [13-graph-traversal-algorithms.md](13-graph-traversal-algorithms.md)

## Problem Statement

Terrain currently maintains three graph representations that model overlapping but distinct views of the test system:

1. **`graph.Graph`** (snapshot index) — a lightweight file-level adjacency structure built during repository scanning. Used for incremental update detection and change-set computation.
2. **`depgraph.Graph`** (typed dependency graph) — the full typed graph with node/edge metadata, confidence scores, and evidence types. Powers coverage, duplicate, and fanout analysis.
3. **`impact.ImpactGraph`** (bidirectional impact graph) — a bidirectional traversal structure derived from the dependency graph, optimized for impact analysis queries ("what tests are affected by this change?").

Each graph has its own construction path, its own traversal API, and its own assumptions about node identity. This creates three problems:

- **Translation overhead.** Engines must convert between representations when they need capabilities from more than one graph. Impact analysis rebuilds adjacency indexes that already exist in the dependency graph.
- **Inconsistent confidence.** The snapshot index has no confidence model. The impact graph applies hop decay but not fanout penalties. Only the dependency graph carries full evidence metadata.
- **Schema drift.** Adding a new node type (e.g., Environment) requires changes in three places with no shared contract enforcing consistency.

## Implemented: Unified Type System

The unified type system is implemented in `internal/depgraph/` and organizes all node types into six families:

### Node Families

| Family | Constant | Node Types | Status |
|--------|----------|------------|--------|
| System | `FamilySystem` | SourceFile, Package, Service, GeneratedArtifact, CodeSurface | Implemented |
| Validation | `FamilyValidation` | ValidationTarget, Test, Scenario, ManualCoverage, Suite, TestFile, Fixture, Helper | Implemented |
| Behavior | `FamilyBehavior` | BehaviorSurface | Implemented |
| Environment | `FamilyEnvironment` | Environment, EnvironmentClass, DeviceConfig, ExternalService, Dataset, Model, Prompt, EvalMetric | Implemented |
| Execution | `FamilyExecution` | ExecutionRun, ValidationExecution | Implemented |
| Governance | `FamilyGovernance` | Owner | Implemented |

All 26 node types are defined as constants with string values for JSON serialization. The `NodeTypeFamily()` function maps each type to its family. `AllNodeTypes()` returns the complete registry.

### Edge Types by Category

| Category | Edge Types | Status |
|----------|-----------|--------|
| Test Structure | TestDefinedInFile, SuiteContainsTest | Implemented, actively used |
| Dependency | ImportsModule, SourceImportsSource, TestUsesFixture, TestUsesHelper, FixtureImportsSource, HelperImportsSource | Implemented, actively used |
| Package | BelongsToPackage | Implemented, actively used |
| Validation | Validates, CoversCodeSurface, ManualCovers | Implemented |
| Behavior | BehaviorDerivedFrom, TestExercises | Implemented |
| Environment | TargetsEnvironment, EnvironmentClassContains, DependsOnService, UsesDataset, UsesModel, UsesPrompt, EvaluatesMetric | Implemented |
| Execution | ExecutionRunsTest | Implemented |
| Governance | Owns | Implemented, actively used |

### Evidence Types

| Evidence Type | Description | Status |
|---------------|-------------|--------|
| StaticAnalysis | Discovered through static code analysis | Implemented, actively used |
| Convention | Inferred from naming conventions | Implemented, actively used |
| Inferred | Heuristically inferred relationship | Implemented, actively used |
| Manual | Explicitly declared in configuration | Implemented, actively used |
| Execution | Observed during runtime execution | Implemented |

### Graph Operations

The `Graph` struct supports:

- **Node queries:** `Node(id)`, `Nodes()`, `NodesByType(t)`, `NodesByFamily(f)`
- **Edge queries:** `Edges()`, `EdgesByType(t)`, `Outgoing(id)`, `Incoming(id)`
- **Traversal:** `Neighbors(id)`, `ReverseNeighbors(id)`
- **Validation queries:** `ValidationTargets()` returns all validation-bearing nodes (Test, Scenario, ManualCoverage); `ValidationsForSurface(id)` returns validation nodes covering a given surface via reverse edge traversal; `IsValidationNode(t)` predicate checks node type membership
- **Statistics:** `Stats()` — includes per-type and per-family counts
- **Serialization:** `MarshalJSON()` / `UnmarshalJSON()` — deterministic JSON with adjacency index rebuild

### Backward Compatibility

All existing analysis engines (coverage, fanout, duplicates, impact) continue to work unchanged. They operate on the node types they recognize and ignore new types. This was verified by adding new node types alongside traditional ones in tests — all four engines produce correct results.

## Remaining Work: Graph Convergence

### Phase 1: Wrapper (Next)

Introduce a `UnifiedGraph` struct that embeds `depgraph.Graph` and adds snapshot metadata as node-level fields. The impact graph's bidirectional indexes are computed on demand from the unified structure. Existing engine function signatures remain unchanged.

### Phase 2: Consolidation

Remove `graph.Graph` and `impact.ImpactGraph` as separate types. All construction logic writes directly to `UnifiedGraph`. Snapshot hashing moves into node metadata. The incremental update path compares node metadata hashes instead of maintaining a parallel index.

### Phase 3: Construction Pipelines

Wire the new node/edge types into construction pipelines:
- **System nodes:** Populated during repository scanning (SourceFile, Package already done)
- **Validation nodes:** Populated during test discovery (Test, Suite, TestFile already done)
- **Behavior nodes:** Derived from code surface analysis
- **Environment nodes:** Extracted from CI configuration and test annotations
- **Execution nodes:** ExecutionRun and ValidationExecution are captured from CI run results and test execution logs
- **Governance nodes:** Owner nodes extracted from CODEOWNERS

## Why Unification Matters

- **Single traversal surface.** Every engine query — coverage, impact, duplicates, fanout, risk — operates on one graph with one API. No translation, no impedance mismatch.
- **Consistent confidence model.** Confidence scores, hop decay, and fanout penalties apply uniformly. An impact path that crosses from a test file through a fixture into a source file into an environment carries a single compounded confidence value.
- **Simpler engine contract.** Engines are pure functions over the graph. When the graph is one thing, the contract is one thing. Adding a new engine does not require deciding which graph representation to target.
- **Extensibility.** New node and edge types slot into the existing schema without structural changes. The graph grows horizontally (more types) without growing vertically (more representations).
