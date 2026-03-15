# Graph Schema

> **Status:** Implemented
> **Purpose:** Data model for the unified dependency graph — node families, node types, edge types, evidence types, and ID conventions.
> **Key decisions:**
> - Unified graph model with six node families: system, validation, behavior, environment, execution, governance
> - Typed nodes and edges with string-based type constants for JSON serialization
> - Edge direction follows dependency direction: `from` depends on / uses / is defined in `to`
> - Evidence types on edges enable confidence scoring (static analysis > convention > inferred)
> - Node IDs use type prefixes (`test:`, `file:`, `pkg:`, etc.) for unambiguous identification
> - Graph supports `MarshalJSON`/`UnmarshalJSON` for deterministic serialization and deserialization

The dependency graph is the structural backbone of Terrain's graph layer. It models relationships between all components of a test system across six conceptual topology layers. The implementation lives in `internal/depgraph/`.

## Node Families

Nodes are organized into six families representing the conceptual layers of the unified graph:

| Family | Constant | Description |
|--------|----------|-------------|
| System | `FamilySystem` | Structural elements of the codebase |
| Validation | `FamilyValidation` | Test system components |
| Behavior | `FamilyBehavior` | Inferred behavioral surfaces |
| Environment | `FamilyEnvironment` | Execution contexts and external dependencies |
| Execution | `FamilyExecution` | Runtime execution state |
| Governance | `FamilyGovernance` | Ownership, policy, and release governance |

## Node Types

### System Topology

| Node Type | Constant | String Value | Description |
|-----------|----------|-------------|-------------|
| SourceFile | `NodeSourceFile` | `"source_file"` | Production source code file |
| Package | `NodePackage` | `"package"` | Package or module boundary |
| Service | `NodeService` | `"service"` | External service dependency |
| Contract | `NodeContract` | `"contract"` | API contract or interface definition |
| ConfigArtifact | `NodeConfigArtifact` | `"config_artifact"` | Configuration file |
| GeneratedArtifact | `NodeGeneratedArtifact` | `"generated_artifact"` | Build output, compiled asset, or generated config |
| CodeSurface | `NodeCodeSurface` | `"code_surface"` | Exported function, method, or public API boundary |

### Validation Topology

| Node Type | Constant | String Value | Description |
|-----------|----------|-------------|-------------|
| ValidationTarget | `NodeValidationTarget` | `"validation_target"` | Abstract validation objective |
| Test | `NodeTest` | `"test"` | Individual test case (an `it()` or `test()` block) |
| Scenario | `NodeScenario` | `"scenario"` | Multi-step behavioral scenario |
| ManualCoverage | `NodeManualCoverage` | `"manual_coverage"` | Manual test suite (TestRail, QA checklist) |
| Suite | `NodeSuite` | `"suite"` | Test suite (a `describe()` block containing tests) |
| TestFile | `NodeTestFile` | `"test_file"` | A file containing test definitions |
| Fixture | `NodeFixture` | `"fixture"` | Shared test setup resource (database seeds, auth contexts) |
| Helper | `NodeHelper` | `"helper"` | Test utility module (custom assertions, request builders) |
| AssertionHelper | `NodeAssertionHelper` | `"assertion_helper"` | Specialized assertion helper |

### Behavior Topology

| Node Type | Constant | String Value | Description |
|-----------|----------|-------------|-------------|
| BehaviorSurface | `NodeBehaviorSurface` | `"behavior_surface"` | Derived behavioral grouping of related CodeSurfaces (by route prefix, class, or module) |

### Environment Topology

| Node Type | Constant | String Value | Description |
|-----------|----------|-------------|-------------|
| Environment | `NodeEnvironment` | `"environment"` | Execution environment (staging, prod, CI) |
| EnvironmentClass | `NodeEnvironmentClass` | `"environment_class"` | Group of similar environments |
| DeviceConfig | `NodeDeviceConfig` | `"device_config"` | Device configuration for mobile/cross-platform |
| FeatureFlag | `NodeFeatureFlag` | `"feature_flag"` | Feature flag affecting test behavior |
| Dataset | `NodeDataset` | `"dataset"` | Test dataset or data fixture |
| Model | `NodeModel` | `"model"` | ML model under evaluation |
| Prompt | `NodePrompt` | `"prompt"` | AI prompt template under evaluation |
| ExternalService | `NodeExternalService` | `"external_service"` | External service dependency |
| Resource | `NodeResource` | `"resource"` | Generic resource dependency |

### Execution Topology

| Node Type | Constant | String Value | Description |
|-----------|----------|-------------|-------------|
| ChangeSet | `NodeChangeSet` | `"change_set"` | Set of changed files triggering analysis |
| ExecutionRun | `NodeExecutionRun` | `"execution_run"` | CI/CD pipeline run |
| ValidationExecution | `NodeValidationExecution` | `"validation_execution"` | Single test execution within a run |
| FailureCluster | `NodeFailureCluster` | `"failure_cluster"` | Group of related failures |
| SkipRecord | `NodeSkipRecord` | `"skip_record"` | Record of a skipped test and reason |
| ExecutionBundle | `NodeExecutionBundle` | `"execution_bundle"` | Group of tests selected to run together |

### Governance Topology

| Node Type | Constant | String Value | Description |
|-----------|----------|-------------|-------------|
| Owner | `NodeOwner` | `"owner"` | Code owner (team or individual) |
| PolicyBundle | `NodePolicyBundle` | `"policy_bundle"` | Set of governance policies |
| ReleaseGate | `NodeReleaseGate` | `"release_gate"` | Release gate requiring test passage |

### Node Structure

```go
// Node is a vertex in the dependency graph (internal/depgraph/node.go).
type Node struct {
    ID        string            `json:"id"`                  // Unique identifier with type prefix
    Type      NodeType          `json:"type"`                // One of the node types above
    Path      string            `json:"path,omitempty"`      // Repository-relative file path
    Name      string            `json:"name,omitempty"`      // Human-readable label
    Line      int               `json:"line,omitempty"`      // Source line number (tests/suites)
    Package   string            `json:"package,omitempty"`   // Containing package or module
    Framework string            `json:"framework,omitempty"` // Test framework (jest, pytest, etc.)
    Metadata  map[string]string `json:"metadata,omitempty"`  // Extensible key-value pairs
}

// Family returns the NodeFamily this node belongs to.
func (n *Node) Family() NodeFamily
```

Metadata varies by node type. Test nodes include the test name. TestFile nodes include the file path and detected framework. ManualCoverage nodes include the coverage area, source system, and criticality.

## Edge Types

### Test Structure Edges

| Edge Type | Constant | String Value | From → To | Description |
|-----------|----------|-------------|-----------|-------------|
| TestDefinedInFile | `EdgeTestDefinedInFile` | `"test_defined_in_file"` | Test → TestFile | Test is defined inside a file |
| SuiteContainsTest | `EdgeSuiteContainsTest` | `"suite_contains_test"` | Suite → Test | Suite contains a test |

### Dependency Edges

| Edge Type | Constant | String Value | From → To | Description |
|-----------|----------|-------------|-----------|-------------|
| ImportsModule | `EdgeImportsModule` | `"imports_module"` | File → File | General import dependency |
| SourceImportsSource | `EdgeSourceImportsSource` | `"source_imports_source"` | SourceFile → SourceFile | Source file imports another source file |
| TestUsesFixture | `EdgeTestUsesFixture` | `"test_uses_fixture"` | TestFile → Fixture | Test file imports a fixture |
| TestUsesHelper | `EdgeTestUsesHelper` | `"test_uses_helper"` | TestFile → Helper | Test file imports a helper |
| FixtureImportsSource | `EdgeFixtureImportsSource` | `"fixture_imports_source"` | Fixture → SourceFile | Fixture imports production code |
| HelperImportsSource | `EdgeHelperImportsSource` | `"helper_imports_source"` | Helper → SourceFile | Helper imports production code |
| BelongsToPackage | `EdgeBelongsToPackage` | `"belongs_to_package"` | File → Package | File belongs to a package |

### Validation Edges

| Edge Type | Constant | String Value | From → To | Description |
|-----------|----------|-------------|-----------|-------------|
| Validates | `EdgeValidates` | `"validates"` | Test → SourceFile | Test validates a source file |
| CoversCodeSurface | `EdgeCoversCodeSurface` | `"covers_code_surface"` | Test → CodeSurface | Test covers a code surface |
| ScenarioContains | `EdgeScenarioContains` | `"scenario_contains"` | Scenario → Test | Scenario contains a test step |
| ManualCovers | `EdgeManualCovers` | `"manual_covers"` | ManualCoverage → SourceFile | Manual coverage covers a source |

### Behavior Edges

| Edge Type | Constant | String Value | From → To | Description |
|-----------|----------|-------------|-----------|-------------|
| BehaviorDerivedFrom | `EdgeBehaviorDerivedFrom` | `"behavior_derived_from"` | BehaviorSurface → SourceFile | Behavior derived from source |
| TestExercises | `EdgeTestExercises` | `"test_exercises"` | Test → BehaviorSurface | Test exercises a behavior |

### Environment Edges

| Edge Type | Constant | String Value | From → To | Description |
|-----------|----------|-------------|-----------|-------------|
| TargetsEnvironment | `EdgeTargetsEnvironment` | `"targets_environment"` | Test → Environment | Test targets an environment |
| EnvironmentClassContains | `EdgeEnvironmentClassContains` | `"environment_class_contains"` | EnvironmentClass → Environment | Class contains an environment |
| RequiresFeatureFlag | `EdgeRequiresFeatureFlag` | `"requires_feature_flag"` | Test → FeatureFlag | Test requires a feature flag |
| UsesDataset | `EdgeUsesDataset` | `"uses_dataset"` | Test → Dataset | Test uses a dataset |
| UsesModel | `EdgeUsesModel` | `"uses_model"` | Test → Model | Test uses a model |
| UsesPrompt | `EdgeUsesPrompt` | `"uses_prompt"` | Test → Prompt | Test uses a prompt |
| DependsOnService | `EdgeDependsOnService` | `"depends_on_service"` | Test → ExternalService | Test depends on external service |
| UsesResource | `EdgeUsesResource` | `"uses_resource"` | Test → Resource | Test uses a resource |

### Execution Edges

| Edge Type | Constant | String Value | From → To | Description |
|-----------|----------|-------------|-----------|-------------|
| RunContainsExecution | `EdgeRunContainsExecution` | `"run_contains_execution"` | ExecutionRun → ValidationExecution | Run contains an execution |
| ExecutionRunsTest | `EdgeExecutionRunsTest` | `"execution_runs_test"` | ExecutionRun → Test | Run executes a test |
| FailureClusterContains | `EdgeFailureClusterContains` | `"failure_cluster_contains"` | FailureCluster → ValidationExecution | Cluster contains a failure |
| SkipRecordReferences | `EdgeSkipRecordReferences` | `"skip_record_references"` | SkipRecord → Test | Skip record references a test |
| BundleContainsExecution | `EdgeBundleContainsExecution` | `"bundle_contains_execution"` | ExecutionBundle → ValidationExecution | Bundle contains an execution |
| ChangeSetTriggersRun | `EdgeChangeSetTriggersRun` | `"change_set_triggers_run"` | ChangeSet → ExecutionRun | Change set triggers a run |

### Governance Edges

| Edge Type | Constant | String Value | From → To | Description |
|-----------|----------|-------------|-----------|-------------|
| Owns | `EdgeOwns` | `"owns"` | Owner → SourceFile | Owner owns a source file |
| PolicyGoverns | `EdgePolicyGoverns` | `"policy_governs"` | PolicyBundle → Package | Policy governs a package |
| ReleaseGateGuards | `EdgeReleaseGateGuards` | `"release_gate_guards"` | ReleaseGate → Package | Gate guards a package release |

### Edge Structure

```go
// Edge is a directed relationship between two nodes (internal/depgraph/edge.go).
type Edge struct {
    From         string       `json:"from"`         // Source node ID
    To           string       `json:"to"`           // Target node ID
    Type         EdgeType     `json:"type"`         // One of the edge types above
    Confidence   float64      `json:"confidence"`   // 0.0 to 1.0
    EvidenceType EvidenceType `json:"evidenceType"` // How the edge was discovered
}
```

### Edge Direction Convention

Edges point in the **dependency direction**: `from` depends on / uses / is defined in `to`.

- A test defined in a file: `test:... → file:...` (test_defined_in_file)
- A test file using a fixture: `file:tests/... → file:fixtures/...` (test_uses_fixture)
- Source importing source: `file:src/a.go → file:src/b.go` (source_imports_source)

This means **reverse traversal** (following edges backward) answers "what tests cover this source file?" and **forward traversal** answers "what does this test depend on?"

## Evidence Types

Each edge carries an evidence type explaining how it was discovered:

| Evidence Type | Constant | String Value | Description |
|---------------|----------|-------------|-------------|
| StaticAnalysis | `EvidenceStaticAnalysis` | `"static_analysis"` | Discovered through static code analysis (AST, regex) |
| Convention | `EvidenceConvention` | `"convention"` | Inferred from naming conventions or directory structure |
| Inferred | `EvidenceInferred` | `"inferred"` | Heuristically inferred relationship |
| Manual | `EvidenceManual` | `"manual"` | Explicitly declared (e.g., in configuration) |
| Execution | `EvidenceExecution` | `"execution"` | Observed during runtime execution |

Evidence type affects confidence scoring. Static analysis evidence produces higher confidence than convention or inferred evidence. See [14-evidence-scoring-and-confidence-model.md](14-evidence-scoring-and-confidence-model.md) for the full scoring model.

## Graph Operations

The `Graph` struct (`internal/depgraph/graph.go`) provides:

- `AddNode(node)` / `Node(id)` — node management
- `AddEdge(edge)` — add a directed edge
- `Edges()` — all edges
- `Neighbors(nodeId)` — outgoing neighbors from a node
- `Incoming(nodeId)` — incoming edges to a node
- `Nodes()` — all nodes, sorted by ID
- `NodesByType(t)` — nodes filtered by type
- `NodesByFamily(f)` — nodes filtered by family
- `NodeCount()` / `EdgeCount()` — graph size metrics
- `Stats()` — summary statistics including per-type and per-family counts
- `ValidationTargets()` — all validation-bearing nodes (Test, Scenario, ManualCoverage), sorted by ID
- `ValidationsForSurface(surfaceID)` — validation nodes covering a given surface via reverse edge traversal (follows `EdgeCoversCodeSurface`, `EdgeTestExercises`, `EdgeManualCovers`, `EdgeValidates`)
- `IsValidationNode(t)` — predicate: true for Test, Scenario, ManualCoverage node types
- `MarshalJSON()` / `UnmarshalJSON()` — deterministic JSON serialization with adjacency index rebuild

## Graph Serialization

The graph serializes to JSON with a version envelope:

```json
{
  "version": "1.0.0",
  "nodes": [ ... ],
  "edges": [ ... ]
}
```

Nodes are sorted by ID for deterministic output. Deserialization rebuilds the adjacency indexes automatically.

## Graph Density

Graph density is computed as:

```
density = edgeCount / (nodeCount * (nodeCount - 1))
```

Density is used by the repository profiler to assess coverage confidence. Sparse graphs (density < 0.5) indicate that many dependency relationships may be missing, reducing confidence in coverage and impact analysis.

## Node ID Conventions

Node IDs use prefixes to distinguish types:

- `test:path/to/file.test.js:lineNumber:testName`
- `suite:path/to/file.test.js:lineNumber:suiteName`
- `file:path/to/file.js`
- `pkg:packageName`
- `svc:serviceName`
- `manual:coverageAreaName`
- `behavior:surfaceName`
- `env:environmentName`
- `owner:teamName`
- `run:runId`
- `contract:contractName`
- `dataset:datasetName`
- `model:modelName`

The test ID format includes file path, line number, and test name to produce deterministic, stable identifiers across runs.

## Related Documents

- [13-graph-traversal-algorithms.md](13-graph-traversal-algorithms.md) — BFS, reverse reachability, and path tracing algorithms used over this schema
- [14-evidence-scoring-and-confidence-model.md](14-evidence-scoring-and-confidence-model.md) — How evidence types map to confidence scores
- [16-unified-graph-schema.md](16-unified-graph-schema.md) — Convergence strategy for the three graph representations
