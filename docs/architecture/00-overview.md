# Architecture Overview

> **Status:** Implemented
> **Purpose:** High-level map of Terrain's architecture — signal pipeline, dependency graph, and how they relate.
> **Key decisions:**
> - Single Go engine with two complementary layers (signal pipeline + dependency graph) rather than separate engines
> - Signals are the core abstraction — every finding is structured with type, severity, evidence, and location
> - Inference-first — Terrain infers structure from what exists in the repo; no configuration required to get started
> - Explainability-first — every finding carries an evidence chain; `terrain explain` traces decisions to their source
> - Dependency graph enables precise impact analysis and structural coverage mapping
> - Local-first design — useful on a single machine without SaaS or network access

Terrain is a signal-first test intelligence platform built on a single Go engine with two complementary analysis layers.

## Two Layers

### Signal Pipeline (Go, `internal/engine/`)

The signal pipeline performs broad repository analysis. It scans test files, detects frameworks, extracts signals (health, quality, migration, governance), scores risk across multiple dimensions, and produces snapshots for trend tracking.

Pipeline:

```
Repository scan
  → Framework detection + test file discovery
    → Signal detection (quality, health, migration, governance)
      → Risk scoring (reliability, change, speed)
        → Snapshot (TestSuiteSnapshot)
          → Reporting (human-readable, JSON, executive summary)
```

### Dependency Graph Layer (Go, `internal/depgraph/` and `internal/graph/`)

The dependency graph layer constructs a typed dependency graph connecting tests, fixtures, helpers, source files, packages, services, and artifacts. Insight engines traverse this graph to answer structural questions: what tests are impacted by a change? Which source files lack coverage? Where are the high-fanout bottlenecks?

Pipeline:

```
Test discovery
  → Graph construction (nodes + edges)
    → Import analysis (static + heuristic)
      → Insight engines (coverage, duplicates, fanout, impact)
        → Repository profiling + edge case detection
          → Reporting (human-readable, JSON, artifacts)
```

## How They Relate

| Concern | Signal Pipeline | Dependency Graph |
|---------|----------------|-----------------|
| Health signals (flaky, slow, skipped) | Primary | — |
| Quality signals (weak assertions, mock-heavy) | Primary | — |
| Migration intelligence | Primary | — |
| Policy and governance | Primary | — |
| Dependency structure | — | Primary |
| Change impact analysis | Coarse (git diff) | Precise (graph traversal) |
| Coverage analysis | Ingested (LCOV) | Structural (reverse graph) |
| Duplicate detection | — | Primary |
| Fanout analysis | — | Primary |
| Repository profiling | — | Primary |

## Core Principles

- **Signals are the core abstraction.** Every finding is a structured signal with type, severity, evidence, and location.
- **The graph is the structural backbone.** Dependency relationships enable precise impact analysis and coverage mapping.
- **Risk must be explainable.** No opaque scores. Every recommendation includes evidence chains.
- **Conservative under uncertainty.** When confidence is low, run more tests, not fewer.
- **Local-first.** Useful on a single machine without SaaS or network access.
- **Privacy boundary.** Aggregate metrics never expose raw file paths or source code.

## Architecture Documents

| Document | Description |
|----------|-------------|
| [01-core-architecture.md](01-core-architecture.md) | Signal engine layers and pipeline |
| [02-graph-schema.md](02-graph-schema.md) | Graph data model: node types, edge types, evidence |
| [03-graph-storage-incremental-updates.md](03-graph-storage-incremental-updates.md) | Graph persistence and incremental rebuild |
| [04-deterministic-test-identity.md](04-deterministic-test-identity.md) | Test identity hashing for stable references |
| [05-insight-engine-framework.md](05-insight-engine-framework.md) | How insight engines traverse the graph |
| [06-monorepo-graph-scaling.md](06-monorepo-graph-scaling.md) | Scaling the graph for large repositories |
| [07-pr-ci-integration.md](07-pr-ci-integration.md) | PR comments, CI artifacts, GitHub Actions |
| [08-test-similarity-structural-fingerprints.md](08-test-similarity-structural-fingerprints.md) | Duplicate detection algorithm |
| [09-cli-spec.md](09-cli-spec.md) | CLI commands, flags, and output modes |
| [10-json-artifact-schemas.md](10-json-artifact-schemas.md) | JSON artifact envelope and schemas |
| [11-ui-requirements.md](11-ui-requirements.md) | VS Code extension and future UI |
| [12-risk-and-coverage-taxonomy.md](12-risk-and-coverage-taxonomy.md) | Risk dimensions and coverage bands |
| [13-graph-traversal-algorithms.md](13-graph-traversal-algorithms.md) | BFS, reverse reachability, path tracing |
| [14-evidence-scoring-and-confidence-model.md](14-evidence-scoring-and-confidence-model.md) | Edge confidence and evidence types |
| [15-edge-case-handling.md](15-edge-case-handling.md) | Repository profiling and adaptive behavior |
| [16-unified-graph-schema.md](16-unified-graph-schema.md) | Converged graph schema unifying snapshot index, depgraph, and impact graph |
| [17-persona-journeys.md](17-persona-journeys.md) | Canonical workflows mapped to user personas and decision contexts |
| [18-environment-and-device-model.md](18-environment-and-device-model.md) | Execution environments and device targets as graph nodes |
| [19-ai-scenario-and-eval-model.md](19-ai-scenario-and-eval-model.md) | AI/ML evaluation suites and behavioral scenarios |
| [20-manual-coverage-and-governance-model.md](20-manual-coverage-and-governance-model.md) | Manual test coverage, QA processes, and governance overlays |
| [21-behavior-surface-derivation.md](21-behavior-surface-derivation.md) | Deriving behavior surfaces from code structure |
| [22-reasoning-engine.md](22-reasoning-engine.md) | Explainable, traceable reasoning chains for findings |
| [23-phased-implementation-roadmap.md](23-phased-implementation-roadmap.md) | Implementation phases from current state to full vision |
