# Unified Graph Schema

> **Status:** Implemented (hardened)
> **Purpose:** Canonical reference for the graph schema that powers Terrain's analysis engines. Defines all node types, edge types, evidence types, construction pipeline, and serialization contract.
> **Key decisions:**
> - Six node families: system, validation, behavior, environment, execution, governance
> - 20 node types and 15 edge types — unused types removed during schema hardening
> - All types defined as Go constants with stable JSON string values
> - Graph is immutable after `Build()` — construct once, then query
> - Deterministic serialization: nodes sorted by ID, edges preserve insertion order
> - Confidence model applies uniformly across all edge types

**See also:** [02-graph-schema.md](02-graph-schema.md), [13-graph-traversal-algorithms.md](13-graph-traversal-algorithms.md), [18-environment-and-device-model.md](18-environment-and-device-model.md), [19-ai-scenario-and-eval-model.md](19-ai-scenario-and-eval-model.md)

## Node Types (20)

### System Family (`FamilySystem`)

Structural elements of the codebase.

| Constant | JSON Value | Built By | Purpose |
|----------|-----------|----------|---------|
| `NodeSourceFile` | `source_file` | `buildImportEdges`, `buildCodeSurfaces` | Source code file |
| `NodeCodeSurface` | `code_surface` | `buildCodeSurfaces` | Function, method, endpoint, or route |

### Validation Family (`FamilyValidation`)

Test system elements.

| Constant | JSON Value | Built By | Purpose |
|----------|-----------|----------|---------|
| `NodeValidationTarget` | `validation_target` | — (reserved) | Abstract validation target |
| `NodeTest` | `test` | `buildTestStructure` | Individual test case |
| `NodeScenario` | `scenario` | `buildScenarios` | Behavioral scenario or manual test spec |
| `NodeManualCoverage` | `manual_coverage` | `buildManualCoverage` | Manual testing artifact (QA checklist, exploratory test) |
| `NodeSuite` | `suite` | `buildTestStructure` | Test suite / describe block |
| `NodeTestFile` | `test_file` | `buildTestStructure` | Test file containing tests and suites |

### Behavior Family (`FamilyBehavior`)

Inferred behavioral groupings.

| Constant | JSON Value | Built By | Purpose |
|----------|-----------|----------|---------|
| `NodeBehaviorSurface` | `behavior_surface` | `buildBehaviorSurfaces` | Behavioral grouping derived from code surfaces |

### Environment Family (`FamilyEnvironment`)

Execution environments, devices, and AI/ML resources.

| Constant | JSON Value | Built By | Purpose |
|----------|-----------|----------|---------|
| `NodeEnvironment` | `environment` | `buildEnvironments` | Execution environment (staging, prod, etc.) |
| `NodeEnvironmentClass` | `environment_class` | `buildEnvironmentClasses` | Grouping of related environments |
| `NodeDeviceConfig` | `device_config` | `buildDeviceConfigs` | Device/browser target |
| `NodeDataset` | `dataset` | — (AI reserved) | ML/AI dataset resource |
| `NodeModel` | `model` | — (AI reserved) | ML/AI model resource |
| `NodePrompt` | `prompt` | — (AI reserved) | AI prompt/template |
| `NodeEvalMetric` | `eval_metric` | — (AI reserved) | Evaluation metric (accuracy, BLEU, etc.) |

### Execution Family (`FamilyExecution`)

Runtime execution state.

| Constant | JSON Value | Built By | Purpose |
|----------|-----------|----------|---------|
| `NodeExecutionRun` | `execution_run` | — (reserved) | CI run or test execution instance |
| `NodeValidationExecution` | `validation_execution` | — (reserved) | Outcome of a test execution |

### Governance Family (`FamilyGovernance`)

Ownership and policy.

| Constant | JSON Value | Built By | Purpose |
|----------|-----------|----------|---------|
| `NodeOwner` | `owner` | `buildScenarios`, `buildManualCoverage` | Team/person responsible for a node |

## Edge Types (15)

### Test Structure

| Constant | JSON Value | Direction | Built By |
|----------|-----------|-----------|----------|
| `EdgeTestDefinedInFile` | `test_defined_in_file` | test → file | `buildTestStructure` |
| `EdgeSuiteContainsTest` | `suite_contains_test` | suite → test | `buildTestStructure` |

### Dependency

| Constant | JSON Value | Direction | Built By |
|----------|-----------|-----------|----------|
| `EdgeImportsModule` | `imports_module` | testfile → sourcefile | `buildImportEdges` |
| `EdgeSourceImportsSource` | `source_imports_source` | sourcefile → sourcefile | `buildSourceToSourceEdges` |

### Package

| Constant | JSON Value | Direction | Built By |
|----------|-----------|-----------|----------|
| `EdgeBelongsToPackage` | `belongs_to_package` | codesurface → sourcefile | `buildCodeSurfaces` |

### Validation

| Constant | JSON Value | Direction | Built By |
|----------|-----------|-----------|----------|
| `EdgeCoversCodeSurface` | `covers_code_surface` | scenario → surface | `buildScenarios` |
| `EdgeManualCovers` | `manual_covers` | manualcoverage → surface | `buildManualCoverage` |

### Behavior

| Constant | JSON Value | Direction | Built By |
|----------|-----------|-----------|----------|
| `EdgeBehaviorDerivedFrom` | `behavior_derived_from` | behavior/surface → surface/file | `buildCodeSurfaces`, `buildBehaviorSurfaces` |

### Environment

| Constant | JSON Value | Direction | Built By |
|----------|-----------|-----------|----------|
| `EdgeTargetsEnvironment` | `targets_environment` | testfile/scenario → environment/device | `buildEnvironmentEdges` |
| `EdgeEnvironmentClassContains` | `environment_class_contains` | class → environment/device | `buildEnvironmentClasses`, `buildDeviceConfigs` |
| `EdgeUsesDataset` | `uses_dataset` | test → dataset | — (AI reserved) |
| `EdgeUsesModel` | `uses_model` | test → model | — (AI reserved) |
| `EdgeUsesPrompt` | `uses_prompt` | test → prompt | — (AI reserved) |
| `EdgeEvaluatesMetric` | `evaluates_metric` | test → evalmetric | — (AI reserved) |

### Execution

| Constant | JSON Value | Direction | Built By |
|----------|-----------|-----------|----------|
| `EdgeExecutionRunsTest` | `execution_runs_test` | run → test | — (reserved) |

### Governance

| Constant | JSON Value | Direction | Built By |
|----------|-----------|-----------|----------|
| `EdgeOwns` | `owns` | owner → node | `buildScenarios`, `buildManualCoverage` |

## Evidence Types

Every edge carries a confidence score (0.0–1.0) and an evidence type describing how the relationship was discovered.

| Constant | JSON Value | Description |
|----------|-----------|-------------|
| `EvidenceStaticAnalysis` | `static_analysis` | Discovered through AST or import analysis |
| `EvidenceConvention` | `convention` | Inferred from naming conventions or file structure |
| `EvidenceInferred` | `inferred` | Heuristically inferred relationship |
| `EvidenceManual` | `manual` | Explicitly declared in `.terrain/` configuration |
| `EvidenceExecution` | `execution` | Observed during runtime execution |

## Graph Construction Pipeline

`Build(snap *TestSuiteSnapshot) *Graph` constructs the graph in 10 sequential stages:

```
Stage  Function                      Produces
─────  ────────────────────────────  ─────────────────────────────────────
 1     buildTestStructure()          NodeTestFile, NodeTest, NodeSuite
                                     EdgeTestDefinedInFile, EdgeSuiteContainsTest
 2     buildImportEdges()            NodeSourceFile, EdgeImportsModule
 3     buildSourceToSourceEdges()    EdgeSourceImportsSource
 4     buildCodeSurfaces()           NodeCodeSurface, EdgeBelongsToPackage,
                                     EdgeBehaviorDerivedFrom
 5     buildBehaviorSurfaces()       NodeBehaviorSurface, EdgeBehaviorDerivedFrom
 6     buildScenarios()              NodeScenario, NodeOwner,
                                     EdgeCoversCodeSurface, EdgeOwns
 7     buildManualCoverage()         NodeManualCoverage, NodeOwner,
                                     EdgeManualCovers, EdgeOwns
 8     buildEnvironments()           NodeEnvironment
 9     buildEnvironmentClasses()     NodeEnvironmentClass, EdgeEnvironmentClassContains
10     buildDeviceConfigs()          NodeDeviceConfig, EdgeEnvironmentClassContains
 +     buildEnvironmentEdges()       EdgeTargetsEnvironment
```

After `Build()`, the graph is immutable. All queries are read-only.

## Graph API

### Node Queries

| Method | Returns | Ordering |
|--------|---------|----------|
| `Node(id)` | Single node by ID | — |
| `Nodes()` | All nodes | Sorted by ID |
| `NodesByType(t)` | Nodes of given type | Sorted by ID |
| `NodesByFamily(f)` | Nodes in given family | Sorted by ID |

### Edge Queries

| Method | Returns | Ordering |
|--------|---------|----------|
| `Edges()` | All edges | Insertion order |
| `EdgesByType(t)` | Edges of given type | Insertion order |
| `Outgoing(id)` | Edges from node | Insertion order |
| `Incoming(id)` | Edges to node | Insertion order |

### Traversal

| Method | Returns | Ordering |
|--------|---------|----------|
| `Neighbors(id)` | Outgoing neighbor IDs | Sorted |
| `ReverseNeighbors(id)` | Incoming neighbor IDs | Sorted |

### Validation Queries

| Method | Returns | Description |
|--------|---------|-------------|
| `IsValidationNode(t)` | bool | True for Test, Scenario, ManualCoverage |
| `ValidationTargets()` | Validation nodes | All validation-bearing nodes, sorted by ID |
| `ValidationsForSurface(id)` | Validation nodes | Reverse lookup: what validates this surface? |

### Serialization

| Method | Description |
|--------|-------------|
| `MarshalJSON()` | Deterministic JSON: version, nodes (sorted by ID), edges |
| `UnmarshalJSON()` | Deserialize and rebuild adjacency indexes |

Schema version: `1.0.0`

## Analysis Engines

All engines operate on `*Graph` as pure functions.

| Engine | Function | Uses |
|--------|----------|------|
| Coverage | `AnalyzeCoverage(g)` | Direct imports + transitive source imports |
| Fanout | `AnalyzeFanout(g, threshold)` | Reverse-topological traversal with Kahn's algorithm |
| Duplicates | `DetectDuplicates(g)` | Fingerprinting + blocking keys + weighted similarity |
| Redundancy | `AnalyzeRedundancy(g)` | Behavioral overlap analysis |
| Impact | `AnalyzeImpact(g, changedFiles)` | BFS from changed files to test nodes |
| Profile | `AnalyzeProfile(g, insights)` | Aggregate health metrics |
| Edge Cases | `DetectEdgeCases(profile, g, insights)` | Scale and quality issue detection |

## Determinism Guarantees

1. **Node iteration** — `Nodes()`, `NodesByType()`, `NodesByFamily()`, `ValidationTargets()` all return nodes sorted by ID.
2. **Neighbor iteration** — `Neighbors()` and `ReverseNeighbors()` return sorted ID lists.
3. **Serialization** — `MarshalJSON()` produces byte-identical output for identical graphs.
4. **Analysis output** — All engine result arrays are sorted by documented criteria (coverage by test count, fanout by dependent count, duplicates by similarity, etc.).

## Types Removed During Hardening

The following types were present in the schema but never created by `Build()` and never populated by any construction pipeline. They were removed to reduce schema surface area:

**Node types removed (6):** `package`, `service`, `generated_artifact`, `external_service`, `fixture`, `helper`

**Edge types removed (7):** `test_uses_fixture`, `test_uses_helper`, `fixture_imports_source`, `helper_imports_source`, `validates`, `test_exercises`, `depends_on_service`

Coverage analysis indirect pathways that referenced fixture/helper node types were also removed — these code paths were unreachable since Build() never created the corresponding nodes or edges. Coverage analysis now uses two pathways: direct imports and transitive source imports.
