# AI Scenario and Evaluation Model

> **Status:** Partially Implemented (Scenario model and graph integration implemented; eval suite inference planned)
> **Purpose:** Define how Terrain models AI/ML evaluation suites, behavioral scenarios, and model validation as part of the test terrain — treating evals as first-class validation alongside traditional tests.
> **Key decisions:**
> - Evaluation suites are first-class validation in the graph, not a separate analysis domain
> - Behavior scenarios are derived coverage targets, analogous to how source files are coverage targets for traditional tests
> - Model version is metadata on eval nodes, not a separate identity dimension — the graph models the test system, not the model lifecycle
> - Inference detects eval frameworks from config files and directory conventions (zero-config principle)
> - Conservative under uncertainty: unknown eval frameworks are flagged but not force-classified
> - Scenarios implement `ValidationTarget` — the shared interface that unifies tests, scenarios, and manual coverage artifacts (see `internal/models/validation_target.go`)

**See also:** [02-graph-schema.md](02-graph-schema.md), [16-unified-graph-schema.md](16-unified-graph-schema.md), [12-risk-and-coverage-taxonomy.md](12-risk-and-coverage-taxonomy.md)

## Problem

AI/ML teams maintain evaluation suites that function as validation but do not look like traditional tests. An eval suite might measure:

- **Accuracy:** Does the model produce correct outputs for a benchmark dataset?
- **Latency:** Does the model respond within acceptable time bounds?
- **Safety:** Does the model refuse harmful requests and avoid generating dangerous content?
- **Bias:** Does the model treat demographic groups equitably?
- **Regression:** Did a model update change behavior on previously stable inputs?

These are validation — they answer "is this system behaving correctly?" — but they use different frameworks (deepeval, promptfoo, ragas, custom scripts), different assertion patterns (threshold-based metrics rather than exact equality), and different organizational structures (eval datasets rather than test files).

Terrain's current model has no representation for eval suites. An AI team running `terrain analyze` sees their eval directory ignored or classified as generic test files with misleading coverage analysis. This creates a blind spot for a growing category of software validation.

## Proposed Node Types

### EvalSuite

An EvalSuite groups related evaluation cases, analogous to a test file grouping test cases.

```
Node Type: EvalSuite
ID Prefix: eval:
ID Format: eval:<relative-path>
```

#### Metadata

| Field | Type | Description |
|-------|------|-------------|
| `framework` | string | Eval framework (deepeval, promptfoo, ragas, custom) |
| `modelVersion` | string | Model version under evaluation, if detectable |
| `metricTypes` | []string | Types of metrics measured (accuracy, latency, safety, bias) |
| `caseCount` | int | Number of eval cases in the suite |
| `datasetPath` | string | Path to eval dataset if separate from code |

### EvalCase

An EvalCase is a single evaluation within a suite, analogous to an individual test case.

```
Node Type: EvalCase
ID Prefix: evalcase:
ID Format: evalcase:<suite-path>:<case-identifier>
```

#### Metadata

| Field | Type | Description |
|-------|------|-------------|
| `metric` | string | The metric this case measures |
| `threshold` | float | Pass/fail threshold for the metric |
| `category` | string | Behavioral category (happy_path, edge_case, adversarial, safety) |
| `input` | string | Eval input summary (truncated for readability) |

### BehaviorScenario

A BehaviorScenario represents a category of model behavior that eval cases validate. Behavior scenarios are coverage targets — they define what should be tested, and eval cases provide the evidence.

```
Node Type: Scenario
ID Prefix: scenario:
ID Format: scenario:<derived-category>:<description-hash>
```

#### Metadata

| Field | Type | Description |
|-------|------|-------------|
| `category` | string | Scenario type (happy_path, edge_case, adversarial, safety, bias, regression) |
| `description` | string | Human-readable behavior description |
| `criticality` | string | How critical this behavior is (critical, important, informational) |
| `coveredBy` | []string | EvalCase IDs that test this scenario |

## Proposed Edge Types

### `EVAL_TESTS_BEHAVIOR`

```
Direction: EvalCase → BehaviorScenario
Confidence: Based on eval case category matching (typically 0.7-0.9)
Evidence: eval_config, eval_annotation, directory_convention
```

This edge means "this eval case tests this behavioral scenario." An eval case that checks whether a model refuses to generate harmful content links to the "safety-refusal" behavior scenario.

### `SCENARIO_COVERS`

```
Direction: BehaviorScenario → SourceFile / ModelEndpoint
Confidence: Based on inference strength (typically 0.5-0.8)
Evidence: eval_config, code_pattern, manual_annotation
```

This edge means "this behavior scenario covers this source file or model endpoint." A safety scenario for a chat endpoint links to the endpoint's handler code, analogous to how a test file links to the source files it covers.

### `EVAL_DEFINED_IN_SUITE`

```
Direction: EvalCase → EvalSuite
Confidence: Structural (1.0)
Evidence: file_structure
```

This edge mirrors `TEST_DEFINED_IN_FILE` for the eval domain.

## Eval Results as Signals

When execution evidence is available (eval result files, CI output), Terrain maps eval outcomes to signals using the same signal framework as traditional test results:

| Eval Outcome | Signal | Severity |
|-------------|--------|----------|
| Safety eval fails | `eval_safety_failure` | Critical |
| Accuracy drops below threshold | `eval_accuracy_regression` | High |
| Latency exceeds budget | `eval_latency_exceeded` | Medium |
| Bias metric exceeds threshold | `eval_bias_detected` | High |
| New behavior scenario has no eval cases | `eval_coverage_gap` | Medium |

Signal severity is influenced by the behavior scenario's criticality. A safety scenario failure is always critical. An informational edge-case failure is low severity.

## Coverage Model

Behavior scenarios function as coverage targets, parallel to how source files are coverage targets in traditional test analysis:

- **High coverage:** 3+ eval cases test the scenario across different input categories
- **Medium coverage:** 1-2 eval cases test the scenario
- **Low coverage:** No eval cases test the scenario

This reuses the coverage band model from [12-risk-and-coverage-taxonomy.md](12-risk-and-coverage-taxonomy.md) without special-casing. The insight engine treats behavior scenarios the same way it treats source files — it counts incoming validation edges and assigns bands.

Coverage gaps in behavior scenarios surface as insights: "The adversarial input scenario for your chat endpoint has no eval cases."

## Inference

Terrain detects eval suites from directory patterns and configuration files, following the zero-config principle.

### Directory Patterns

| Pattern | Inference |
|---------|-----------|
| `eval/`, `evals/`, `evaluations/` | Eval suite directory |
| `__evals__/` | Eval suite directory (Python convention) |
| `benchmarks/` with model config | Eval suite (distinguished from performance benchmarks by content) |

### Framework Detection

| Framework | Detection Signal |
|-----------|-----------------|
| deepeval | `deepeval.toml`, imports from `deepeval` |
| promptfoo | `promptfooconfig.yaml`, `promptfoo` in package.json |
| ragas | imports from `ragas`, `ragas` in requirements.txt |
| Custom | Python/JS files in eval directories with threshold assertions |

### Distinguishing Evals from Tests

Not every file in an `eval/` directory is an eval suite. Terrain uses content analysis to distinguish:

- Files with metric threshold assertions (accuracy > 0.95) are eval cases
- Files with standard test assertions (assertEqual, expect) are traditional tests that happen to live in an eval directory
- Files with dataset loading patterns (CSV, JSONL, HuggingFace datasets) are likely eval suites

When classification is ambiguous, Terrain marks the file as `classification: uncertain` and reports it with reduced confidence. The `explain` command shows exactly why Terrain classified (or declined to classify) each file, maintaining the explainability principle.

## Implemented: Scenario Model and ValidationTarget

The `Scenario` struct is implemented in `internal/models/validation_target.go` and represents a behavioral scenario — a multi-step workflow, AI evaluation case, or derived behavior specification that validates system behavior.

Scenarios implement the `ValidationTarget` interface alongside `TestCase` and `ManualCoverageArtifact`, enabling impact and coverage logic to operate generically over all validation-bearing entities. Key fields:

| Field | Type | Description |
|-------|------|-------------|
| `scenarioId` | string | Stable identifier (`scenario:<path>:<name>` or `scenario:<category>:<hash>`) |
| `name` | string | Human-readable label |
| `category` | string | Classification: `happy_path`, `edge_case`, `adversarial`, `safety`, `regression` |
| `framework` | string | Eval/test framework: `deepeval`, `promptfoo`, `custom` |
| `coveredSurfaceIds` | []string | CodeSurface or BehaviorSurface IDs this scenario exercises |
| `executable` | bool | Whether this scenario can be run in CI |

### Graph Integration (Implemented)

Scenarios are built into the dependency graph via `buildScenarios()` in `internal/depgraph/build.go`:

- Each `Scenario` becomes a `NodeScenario` node with metadata (category, framework, executable, description)
- Covered surfaces are connected via `EdgeCoversCodeSurface` edges (confidence 0.8, evidence `inferred`)
- If an owner is specified, a `NodeOwner` node is created with an `EdgeOwns` edge
- Scenarios are queryable via `Graph.ValidationTargets()` and `Graph.ValidationsForSurface()`

### ValidationTarget Interface

The `ValidationTarget` interface (`internal/models/validation_target.go`) provides a shared abstraction:

```go
type ValidationTarget interface {
    ValidationID() string
    ValidationName() string
    ValidationKindOf() ValidationKind  // "test", "scenario", "manual"
    ValidationPath() string
    ValidationOwner() string
    IsExecutable() bool
}
```

`CollectValidationTargets(snap)` aggregates all validation-bearing entities from a snapshot into a single slice, preserving insertion order: tests → scenarios → manual coverage.

## Relationship to Unified Graph

EvalSuite, EvalCase, and BehaviorScenario nodes participate in the unified graph schema defined in [16-unified-graph-schema.md](16-unified-graph-schema.md). They use the same traversal algorithms, confidence model, and insight engine contract as all other node types. Impact analysis works across domain boundaries: a change to a model endpoint's handler code surfaces affected eval cases through the `SCENARIO_COVERS` and `EVAL_TESTS_BEHAVIOR` edges, just as a change to a source file surfaces affected tests through `SOURCE_IMPORTS_SOURCE` and `TEST_USES_HELPER` edges.
