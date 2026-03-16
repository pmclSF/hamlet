# Behavior Inference

> **Status:** Implemented
> **Purpose:** Document how Terrain infers structure from code — functions, handlers, prompts, dataset usage, and test definitions — without configuration. BehaviorSurface is derived from CodeSurface, never required as input.

**See also:** [21-behavior-surface-derivation.md](21-behavior-surface-derivation.md), [01-core-architecture.md](01-core-architecture.md), [16-unified-graph-schema.md](16-unified-graph-schema.md)

## Design Principle

Terrain is inference-first. It reads your repository and infers structure from what already exists — import graphs, file naming, coverage artifacts, runtime results. No configuration required to get started.

All behavior anchors (CodeSurface) and behavior groupings (BehaviorSurface) are **derived from source code**, not declared in YAML or configuration. The inference pipeline is deterministic and auditable — every detected surface traces to a specific regex pattern and source line.

## Inference Pipeline

```
Source Files
    │
    ├─ Code Unit Extraction         → CodeUnit[]    (exported symbols)
    ├─ Code Surface Inference       → CodeSurface[] (behavior anchors)
    ├─ Test Case Extraction         → TestCase[]    (test definitions)
    └─ Import Graph Construction    → ImportLink[]  (test→source edges)
                │
                ▼
    Behavior Surface Derivation     → BehaviorSurface[] (groupings)
                │
                ▼
    depgraph.Build()                → Graph (nodes + edges)
```

Each stage is a pure function. No stage requires output from a later stage. BehaviorSurface derivation is the last inference step before graph construction and is entirely optional — the graph, coverage analysis, impact analysis, and all other engines work correctly without it.

## What Gets Inferred

### 1. Functions

Exported functions are the primary behavior anchors. Detection is language-specific:

| Language | Pattern | Example |
|----------|---------|---------|
| JS/TS | `export function name()`, `export default function name()`, `export const name` | `export function login() {}` |
| Go | `func Name()` (uppercase first letter) | `func HandleRequest() {}` |
| Python | `def name()` (not starting with `_`), respects `__all__` | `def process_order():` |
| Java | `public ... name()` (excluding constructors) | `public void createUser()` |

**Surface kind:** `function`

### 2. Handlers

HTTP handlers and middleware are detected by naming convention and signature:

| Language | Detection Method | Example |
|----------|-----------------|---------|
| JS/TS | Function name contains `Handler`, `Middleware`, or `Controller` | `export async function loginHandler()` |
| Go | Name suffix + methods accepting `http.ResponseWriter` | `func (h *API) GetUser(w http.ResponseWriter, r *http.Request)` |
| Python | Function name contains `handler`, `view`, `endpoint`, `controller` | `def login_handler(request):` |
| Java | Class name contains `Controller`, `Resource`, `Endpoint`, `Handler` | `public class UserController` |

**Surface kind:** `handler`

### 3. Routes

HTTP route registrations are detected from framework-specific patterns:

| Language | Pattern | Example |
|----------|---------|---------|
| JS/TS | `app.get('/path')`, `router.post('/path')`, `server.delete('/path')` | `app.get('/api/users', getUsers)` |
| Go | `http.HandleFunc("/path")`, `mux.Handle("/path")`, `router.Get("/path")` | `router.Get("/api/users", handler)` |
| Python | `@app.route('/path')`, `@blueprint.post('/path')` | `@app.get('/api/users')` |
| Java | `@GetMapping("/path")`, `@PostMapping("/path")`, `@RequestMapping("/path")` | `@GetMapping("/api/users")` |

**Surface kind:** `route` — includes HTTP method and route path metadata.

### 4. Prompts

AI prompt templates and prompt-building functions are detected by naming convention:

| Language | Pattern | Example |
|----------|---------|---------|
| JS/TS | Export name contains `Prompt`, `Template`, or `PROMPT` | `export const systemPrompt = "..."` |
| JS/TS | Export function name contains `Prompt` or `Template` | `export function buildUserPrompt()` |
| Python | Function/variable name contains `prompt` or `template` | `def build_prompt(context):` |

**Surface kind:** `prompt`

### 5. Dataset Usage

Dataset loaders and data pipeline entry points are detected by naming convention:

| Language | Pattern | Example |
|----------|---------|---------|
| JS/TS | Export name contains `Dataset`, `Dataloader`, `TrainingData`, `EvalData` | `export const trainingDataset = [...]` |
| JS/TS | Export function name matches dataset patterns | `export function loadEvalData()` |
| Python | Function/variable name contains `dataset`, `dataloader`, `training_data`, `eval_data`, `load_data` | `def load_dataset(path):` |

**Surface kind:** `dataset`

### 6. Test Definitions

Test cases are extracted with framework-aware parsing:

| Language | Frameworks | Detection |
|----------|-----------|-----------|
| JS/TS | Jest, Vitest, Mocha, Jasmine, Playwright, Cypress | `describe()`/`it()`/`test()` with brace-tracked nesting, `test.each()` parametrization |
| Go | go-testing | `func TestXxx()` with `t.Run()` subtests |
| Python | pytest, unittest | `def test_*()`, `class Test*` with `@pytest.mark.parametrize` |
| Java | JUnit 4/5, TestNG | `@Test`, `@ParameterizedTest` annotations |

Each test case gets a stable deterministic ID (`TestID`) computed from path, suite hierarchy, name, and parameter signature. IDs survive line-number changes and file moves.

## BehaviorSurface Derivation

BehaviorSurface nodes are **derived from CodeSurface**, never provided as input. They group related code surfaces into cohesive behavioral units using five priority-ordered strategies:

| Priority | Strategy | Grouping Key | Minimum Surfaces | Example |
|----------|----------|-------------|-----------------|---------|
| 1 | Route Prefix | First 2 path segments of route | 2 | `GET /api/users` + `POST /api/users` → `behavior:route:/api/users` |
| 2 | Class | (file path, parent/receiver name) | 2 | `UserController.Get` + `UserController.Delete` → `behavior:class:UserController` |
| 3 | Domain | Directory path | 3 (from 2+ files) | `src/auth/login.ts` + `src/auth/register.ts` → `behavior:domain:src/auth` |
| 4 | Naming Prefix | Shared camelCase/snake_case prefix | 3 | `validateEmail` + `validatePassword` → `behavior:naming:validate*` |
| 5 | Module | Source file path | 2 | Two functions in `utils.ts` → `behavior:module:utils.ts` |

**Deduplication:** Singleton behaviors (1 surface) are pruned if the surface already appears in a multi-surface group from a higher-priority strategy.

### BehaviorSurface Is Optional

The graph construction pipeline, coverage analysis, impact analysis, duplicate detection, and all other engines work correctly when no BehaviorSurface nodes exist. This is verified by `TestBuildWithoutBehaviorSurfaces` in `internal/depgraph/build_inference_test.go`.

BehaviorSurface adds value when present — it enables behavior-level grouping in reports and portfolio analysis — but it is never a prerequisite for any engine.

## CodeSurface Kind Summary

| Kind | JSON Value | Description | Detection Method |
|------|-----------|-------------|-----------------|
| `SurfaceFunction` | `function` | Exported function | Language-specific export patterns |
| `SurfaceMethod` | `method` | Method on a type/class | Receiver/parent detection |
| `SurfaceHandler` | `handler` | HTTP handler or middleware | Naming convention + signature |
| `SurfaceRoute` | `route` | Registered HTTP route | Framework API call patterns |
| `SurfaceClass` | `class` | Class or struct | Export/public patterns |
| `SurfacePrompt` | `prompt` | AI prompt template | Naming convention (`*prompt*`, `*template*`) |
| `SurfaceDataset` | `dataset` | Dataset loader or data pipeline | Naming convention (`*dataset*`, `*dataloader*`) |

## Evidence Model

All inferred surfaces carry:
- **Source line number** — traces to the exact line where the surface was detected
- **Language** — the programming language of the source file
- **LinkedCodeUnit** — reference to the corresponding CodeUnit for coverage linkage
- **Exported flag** — whether the surface is publicly visible

When wired into the dependency graph, edges carry:
- **Confidence** — 0.7 for inferred behavior derivation, 1.0 for static analysis
- **EvidenceType** — `static_analysis` for direct detection, `inferred` for behavior grouping

## Test Coverage

Inference is tested at three levels:

1. **Unit tests per language** — Each extractor has dedicated tests for routes, handlers, functions, classes, prompts, and datasets (`internal/analysis/code_surface_test.go`)
2. **Behavior derivation tests** — All 5 grouping strategies tested with determinism verification (`internal/analysis/behavior_surface_test.go`)
3. **Graph integration tests** — Surfaces wired into depgraph with cross-family traversal, coverage, and impact engine verification (`internal/depgraph/build_inference_test.go`)

Key test: `TestBuildWithoutBehaviorSurfaces` confirms that BehaviorSurface is derived and optional.
