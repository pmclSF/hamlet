# Behavior Surface Derivation

> **Status:** Partially Implemented (core inference implemented, gap analysis planned)
> **Purpose:** Define how Terrain derives the behavior surfaces that tests are supposed to validate, without requiring users to enumerate them.
> **Key decisions:**
> - Behavior surfaces are derived from code structure, not declared by users — no manual YAML or config required
> - Inference uses regex-based pattern matching per language (JS/TS, Go, Python, Java)
> - Five surface kinds: `function`, `method`, `handler`, `route`, `class`
> - Surfaces are integrated into the dependency graph as `NodeCodeSurface` nodes
> - Gap severity is based on structural position: boundary functions score higher than internal utilities (planned)
> - Behavior surfaces are a subset of code units, filtered by visibility and structural role

**See also:** [01-core-architecture.md](01-core-architecture.md), [12-risk-and-coverage-taxonomy.md](12-risk-and-coverage-taxonomy.md), [19-ai-scenario-and-eval-model.md](19-ai-scenario-and-eval-model.md), [16-unified-graph-schema.md](16-unified-graph-schema.md)

## Problem

Teams want to know two things: "what behaviors does our system have?" and "which of those behaviors are validated?" But asking teams to enumerate their behaviors is impractical. Behavior lists drift from reality within days of being written. Documentation goes stale. Requirements databases become graveyards.

Meanwhile, the code itself is the authoritative record of what the system does. Exported functions, API endpoints, event handlers, UI components, CLI commands — these are the concrete surfaces where behavior is exposed. They change when the code changes. They are always up to date.

Terrain derives behavior surfaces from code structure so that coverage analysis starts from what the system actually does, not what someone once wrote down.

## Behavior Surface Types

Terrain recognizes the following behavior surface types, ordered by typical criticality:

| Surface Type | Description | Detection Method | Default Criticality |
|-------------|-------------|-----------------|---------------------|
| API endpoint | HTTP route handler (REST, GraphQL) | Framework-specific patterns (Express, Fastify, Flask, etc.) | High |
| CLI command | Command-line interface entry point | Commander, yargs, argparse patterns | High |
| Event handler | Pub/sub, webhook, or message queue handler | Framework-specific listener patterns | High |
| Exported function | Public function exported from a module | Export statement analysis | Medium |
| UI component | Rendered component (React, Vue, Angular, etc.) | Component definition and export patterns | Medium |
| State transition | State machine transition or reducer | Framework-specific patterns (Redux, XState, etc.) | Medium |
| Data transformation | Function that transforms data between shapes | Heuristic: input/output type analysis, naming patterns | Low |
| Error path | Explicit error handling branch | Catch blocks, error callbacks, error boundary patterns | Medium |
| Integration boundary | External service call or SDK usage | Import analysis + call-site detection for known SDKs | High |

## How Behavior Surfaces Are Derived

### Inference-First Philosophy

Terrain's core value proposition is that behavior surfaces are inferred automatically from code. Users never need to define surfaces manually — no YAML definitions, no annotation requirements, no manual enumeration. The code is the authoritative record of system behavior, and Terrain reads it directly.

This is implemented in `internal/analysis/code_surface.go` via the `SurfaceExtractor` interface, with one implementation per supported language.

### Step 1: Language Detection and File Walking

`InferCodeSurfaces()` walks the repository, skipping test files (already identified by the analyzer), vendor directories, and build artifacts. Each source file is dispatched to the appropriate language extractor based on file extension.

### Step 2: Per-Language Surface Extraction (Implemented)

Each language extractor reads the file content and applies regex patterns to detect surfaces. The extractors are:

| Extractor | Languages | Patterns |
|-----------|-----------|----------|
| `jsSurfaceExtractor` | `.js`, `.ts`, `.jsx`, `.tsx`, `.mjs` | Express-style routes (`app.get/post/...`), exported functions with handler/middleware/controller naming, `export function`, `export class` |
| `goSurfaceExtractor` | `.go` | `http.HandleFunc`, mux/router patterns, functions with `http.ResponseWriter` parameter, exported functions and methods with receivers |
| `pythonSurfaceExtractor` | `.py` | Flask/FastAPI route decorators (`@app.route/get/post`), handler-named functions, public functions (no `_` prefix), classes |
| `javaSurfaceExtractor` | `.java` | Spring annotations (`@GetMapping`, `@PostMapping`, `@RequestMapping`), `@RestController`/`@Controller` classes, public methods |

Each extractor produces `models.CodeSurface` values with:
- **SurfaceID** — deterministic ID: `surface:<path>:<name>` or `surface:<path>:<parent>.<name>`
- **Kind** — one of: `function`, `method`, `handler`, `route`, `class`
- **Metadata** — language, package, line number, HTTP method, route path, receiver, export status

### Step 3: Visibility Filtering

Not all code units are behavior surfaces. The extractors apply visibility rules per language:

- **JS/TS** — only `export`ed symbols are surfaces (routes are always included since they register behavior regardless of export)
- **Go** — only capitalized (exported) functions/methods are surfaces; unexported helpers are excluded
- **Python** — only public functions (no `_` prefix) are surfaces; `_internal_helper` is excluded
- **Java** — public classes and methods are surfaces; private/package-private are excluded

### Step 4: Graph Integration (Implemented)

Each inferred `CodeSurface` becomes a `NodeCodeSurface` node in the dependency graph (see `internal/depgraph/build.go`). The node is connected to:
- Its containing source file via an edge
- Its linked code unit (if one exists) via an `EdgeBehaviorDerivedFrom` edge

Surface metadata (kind, language, HTTP method, route, receiver) is stored in the node's `Metadata` map for downstream analysis.

### Step 5: Criticality Assignment (Planned)

Each behavior surface will be assigned a default criticality based on its type (see table above). Criticality overrides via configuration are planned but not yet implemented.

## BehaviorSurface Derivation (Implemented)

While CodeSurfaces identify individual behavior anchors (a single route, a single handler), BehaviorSurfaces group related anchors into cohesive behavioral units. This maps to how teams think about their system: "user authentication" rather than individual route handlers.

BehaviorSurfaces are derived automatically from CodeSurfaces — no manual definition required. They are optional: all analysis pipelines (coverage, impact, risk) work with or without them. When present, they provide higher-level explanations.

This is implemented in `internal/analysis/behavior_surface.go` via `DeriveBehaviorSurfaces()`.

### Derivation Strategies

Five strategies are applied, in priority order:

| Strategy | Kind | Grouping Key | Min Surfaces | Example |
|----------|------|-------------|-------------|---------|
| Route prefix | `route_prefix` | First two path segments | 2 | `POST /api/users`, `GET /api/users/:id`, `DELETE /api/users/:id` → behavior `/api/users/*` |
| Class/receiver | `class` | Parent class + source file | 2 | `UserController.GetUser`, `UserController.DeleteUser` → behavior `UserController` |
| Domain | `domain` | Directory boundary (first two dirs) | 3 across 2+ files | `src/auth/login.ts:login`, `src/auth/token.ts:validate` → behavior `auth` |
| Naming | `naming` | Shared PascalCase/snake_case prefix | 3 | `AuthLogin`, `AuthRegister`, `AuthLogout` → behavior `Auth*` |
| Module | `module` | Source file path | 2 | All exports from `auth.ts` → behavior `auth` |

Surfaces may appear in multiple groups (a route handler belongs to both its route group and its module group). This is intentional — it reflects how the same code participates in multiple behavioral concerns.

**Domain grouping** captures service/domain boundaries that span multiple files. It requires at least 3 surfaces from at least 2 files under the same second-level directory. This prevents single-file modules from generating redundant domain groups alongside module groups.

**Naming grouping** detects shared prefixes in function and method names, using PascalCase word boundaries (e.g., `AuthLogin` → `Auth`) and snake_case underscores (e.g., `hash_password` → `hash`). Route names (containing spaces or slashes) are excluded. The minimum threshold of 3 surfaces prevents noisy groupings from coincidental prefixes.

### Pruning

Single-surface groups are dropped when that surface already belongs to a multi-surface group. This prevents noise from trivial groupings while keeping meaningful singleton behaviors that have no other group.

### Explainability

Every BehaviorSurface contains:
- `CodeSurfaceIDs` — the concrete CodeSurface IDs that constitute the group (always non-empty)
- `Description` — human-readable explanation referencing the derivation source and count
- `Kind` — which strategy produced the grouping (`route_prefix`, `class`, `domain`, `naming`, `module`)

Explanations are always grounded in concrete code surfaces. A behavior like "API routes under /api/users" traces directly to the specific `GET`, `POST`, `DELETE` route surfaces it was derived from.

### Graph Integration

Each BehaviorSurface becomes a `NodeBehaviorSurface` node in the dependency graph, connected to its constituent `NodeCodeSurface` nodes via `EdgeBehaviorDerivedFrom` edges (confidence 0.7, evidence type `inferred`).

## Coverage Mapping

Each behavior surface becomes a coverage target in the graph. Tests that exercise the surface — directly by importing and calling it, or transitively through the dependency graph — provide coverage.

Coverage is attributed using the same graph traversal and confidence model described in [14-evidence-scoring-and-confidence-model.md](14-evidence-scoring-and-confidence-model.md). A test that directly imports and calls `validateToken()` provides higher-confidence coverage than a test that exercises it transitively through an API integration test.

### Direct Coverage

A test directly covers a behavior surface when:
- The test file imports the module containing the surface
- The test exercises the surface function/method/component (detected via call-site analysis or framework-specific test patterns)

Direct coverage confidence: 0.9-1.0 (based on evidence type).

### Transitive Coverage

A test transitively covers a behavior surface when:
- The test exercises a code unit that depends on the surface through the dependency graph
- The dependency path has sufficient confidence (above minimum threshold of 0.1)

Transitive coverage confidence: compounds through the path using the standard confidence propagation model (multiplicative with 0.85^depth decay).

## Gap Analysis

Behavior surfaces without covering tests are gaps. Terrain ranks gaps by severity using a composite score:

1. **Surface criticality** — high-criticality surfaces (API endpoints, integration boundaries) produce more severe gaps than low-criticality surfaces (data transformations)
2. **Structural position** — boundary functions (exported, route-registered) score higher than internal functions that happen to be exported
3. **Change frequency** — surfaces that change often in version control are higher-risk gaps (if git history is available)
4. **Dependency fanout** — surfaces with many dependents are higher-risk because a bug has wider blast radius

### Inference Example

Consider a module `src/auth/token.js` that exports three functions:

```javascript
export function validateToken(token) { /* ... */ }
export function refreshToken(token) { /* ... */ }
function parseTokenPayload(raw) { /* ... */ }
```

- `validateToken` — exported function, behavior surface (medium criticality). If imported by a route handler, elevated to high criticality as an integration boundary dependency.
- `refreshToken` — exported function, behavior surface (medium criticality).
- `parseTokenPayload` — not exported, not a behavior surface. Internal implementation detail.

If a test file imports `token.js` and calls `validateToken()` but never calls `refreshToken()`, Terrain reports `refreshToken` as a coverage gap.

## Relationship to Code Units

Behavior surfaces are a strict subset of code units. The full code unit set includes everything the static analysis layer extracts. Behavior surfaces are the code units that represent user-facing or system-boundary behavior — the things that matter for coverage analysis.

This distinction is important: Terrain does not flag every untested internal function as a gap. Only behavior surfaces — the exported, visible, boundary-level functions — are expected to have coverage. This keeps gap analysis focused and actionable rather than overwhelming teams with noise about untested private helpers.

## Key Decisions

1. **Derived, not declared.** Behavior surfaces come from code analysis, not user configuration. This eliminates documentation drift and means coverage targets update automatically as code changes. Users never need to enumerate surfaces manually — the code is the source of truth.

2. **Conservative inference.** Terrain only includes high-confidence surfaces: exported symbols, registered routes, and handler-named functions. Unexported functions, private methods, and `_`-prefixed Python functions are excluded. It is better to miss a real surface than to flag a false one — noise erodes trust.

3. **Structural position determines severity.** (Planned) A coverage gap on an exported API endpoint handler is more severe than a gap on an exported utility function. Terrain will use the behavior surface type and its position in the dependency graph to rank gap severity, not just whether coverage exists.

4. **Modular per-language extractors.** Each language has its own `SurfaceExtractor` implementation registered via `init()`. Adding support for a new language means implementing one interface method (`ExtractSurfaces`) and registering the extractor. No changes to the core inference loop are needed.

5. **Deterministic IDs.** Surface IDs are constructed deterministically from path + name + parent, producing stable identifiers that can be tracked across snapshots for longitudinal analysis. The format `surface:<path>:<name>` mirrors the existing `unit:<path>:<name>` convention for code units.
