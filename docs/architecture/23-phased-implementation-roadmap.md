# Phased Implementation Roadmap

> **Status:** Reference
> **Purpose:** Document the implementation phases for Terrain's architecture evolution, from current state to full vision.
> **Key decisions:**
> - Phases are additive — each phase builds on the previous, nothing is discarded
> - Core principles (inference-first, explainability, conservative under uncertainty) are invariant across all phases
> - Each phase must be independently useful — no phase exists solely to enable a future phase

**See also:** [00-overview.md](00-overview.md), [01-core-architecture.md](01-core-architecture.md), [09-cli-spec.md](09-cli-spec.md)

## Overview

Terrain's architecture is designed to evolve incrementally. The vision described across the architecture documents spans multiple phases of implementation, from the signal engine that exists today to organization-level intelligence that is aspirational. This document maps that trajectory and links each capability to the architecture document that defines it.

Each phase is independently valuable. A team using Terrain in Phase 1 gets actionable risk assessment and coverage analysis. Each subsequent phase deepens the analysis without invalidating previous results.

## Phase 1: Signal Engine and Dependency Graph (Implemented)

The foundation. Terrain scans a repository, detects frameworks, builds a dependency graph, and produces structured findings.

### Capabilities

- **Repository scanning and framework detection** — identifies test frameworks by analyzing package manifests, config files, and import patterns across JavaScript, TypeScript, Python, Go, and Java ecosystems
- **Signal detection pipeline** — registry-based detectors emit structured signals across four categories: health, quality, migration, and governance
- **Dependency graph construction** — builds a directed graph of imports, fixtures, helpers, and test-to-source relationships with confidence-weighted edges
- **Coverage analysis** — maps test files to source files through graph traversal, identifying covered and uncovered code units
- **Duplicate detection** — identifies structurally similar tests using fingerprinting and similarity thresholds
- **Fanout analysis** — identifies high-fanout nodes (fixtures, helpers) that create disproportionate CI pressure
- **Impact analysis** — given a set of changed files, determines which tests are affected through graph traversal
- **Risk scoring** — aggregates signals into reliability, change, and speed risk dimensions with explainable breakdowns
- **Executive summary** — produces a human-readable summary of repository health with key findings and recommendations
- **CLI with four canonical journeys** — `analyze`, `explain`, `insights`, `impact` commands for different investigation paths
- **Snapshot workflow** — save analysis results and compare across snapshots for trend detection
- **Policy evaluation** — evaluate repository state against configurable governance policies
- **PR/CI integration** — GitHub Actions workflow for automated PR analysis and test selection

### Architecture Documents

| Document | Coverage |
|----------|----------|
| [01-core-architecture.md](01-core-architecture.md) | Signal engine layers, pipeline structure |
| [02-graph-schema.md](02-graph-schema.md) | Node and edge types, graph construction |
| [04-deterministic-test-identity.md](04-deterministic-test-identity.md) | Test hashing for stable cross-snapshot references |
| [05-insight-engine-framework.md](05-insight-engine-framework.md) | Insight generation from signals and graph |
| [07-pr-ci-integration.md](07-pr-ci-integration.md) | PR analysis workflow and test selection |
| [09-cli-spec.md](09-cli-spec.md) | CLI commands and output formats |
| [10-json-artifact-schemas.md](10-json-artifact-schemas.md) | JSON output schema definitions |
| [12-risk-and-coverage-taxonomy.md](12-risk-and-coverage-taxonomy.md) | Risk dimensions and coverage bands |
| [13-graph-traversal-algorithms.md](13-graph-traversal-algorithms.md) | BFS/DFS traversal, cycle detection, path finding |
| [14-evidence-scoring-and-confidence-model.md](14-evidence-scoring-and-confidence-model.md) | Edge confidence, hop decay, fanout penalties |
| [15-edge-case-handling.md](15-edge-case-handling.md) | Graceful degradation for unusual codebases |

## Phase 2: Unified Graph and Behavior Surfaces (In Progress)

Deepens analysis by unifying graph representations and introducing behavior-surface-level coverage assessment.

### Capabilities

- **Unified graph schema** — merge the dependency graph, impact graph, and analysis graph into a single queryable structure with consistent node and edge types
- **Behavior surface derivation** — automatically identify the behavior surfaces (exported functions, API endpoints, event handlers, UI components) that tests should validate, without requiring user enumeration
- **Reasoning engine** — produce composable, serializable reason chains for every finding so that every score and recommendation can be traced to concrete evidence
- **Enhanced explain command** — full reason chain output showing observation-rule-conclusion steps with confidence at each stage
- **Manual coverage modeling** — represent manual test coverage (TestRail suites, QA checklists, exploratory testing) as graph nodes that supplement automated coverage in risk assessment

### Architecture Documents

| Document | Coverage |
|----------|----------|
| [16-unified-graph-schema.md](16-unified-graph-schema.md) | Merged graph model and migration path |
| [21-behavior-surface-derivation.md](21-behavior-surface-derivation.md) | Surface type detection and gap analysis |
| [22-reasoning-engine.md](22-reasoning-engine.md) | Reason chain structure and composition |
| [20-manual-coverage-and-governance-model.md](20-manual-coverage-and-governance-model.md) | Manual coverage nodes and governance policies |

## Phase 3: Environments and Extended Validation (Planned)

Extends analysis beyond single-environment, automated-only testing to account for the full validation landscape.

### Capabilities

- **Environment and device model** — represent target environments (OS, browser, device, cloud region) as graph nodes, enabling environment-aware analysis
- **Environment-aware coverage** — track which tests run in which environments and identify environment-specific coverage gaps (e.g., "auth flow tested in Chrome but not Safari")
- **Cross-environment gap analysis** — surface behavior surfaces that are validated in some environments but not others, ranked by environment criticality
- **AI/ML evaluation suite integration** — model AI evaluation suites (accuracy benchmarks, bias checks, regression tests) as first-class validation alongside traditional tests
- **Manual coverage API integration** — connect to TestRail, Xray, and Qase APIs for live ingestion of manual test execution data, replacing static YAML configuration
- **Execution evidence ingestion** — ingest CI execution history to detect flaky tests, slow suites, and execution trends over time with higher confidence than static analysis alone

### Architecture Documents

| Document | Coverage |
|----------|----------|
| [17-environment-and-device-model.md](17-environment-and-device-model.md) | Environment nodes, device matrix, coverage gaps |
| [19-ai-scenario-and-eval-model.md](19-ai-scenario-and-eval-model.md) | AI/ML evaluation integration |
| [20-manual-coverage-and-governance-model.md](20-manual-coverage-and-governance-model.md) | API integration for live manual coverage |

## Phase 4: Organization-Level Intelligence (Aspirational)

Scales analysis from individual repositories to organizational patterns and cross-repo insights.

### Capabilities

- **Multi-repo analysis** — analyze multiple repositories as a connected system, identifying shared dependencies, cross-repo test gaps, and organization-wide patterns
- **Organization-wide risk heatmaps** — aggregate risk scores across repositories to surface systemic issues (e.g., "all services using auth-sdk have low test coverage for token refresh")
- **Benchmark positioning** — compare a repository's health metrics against anonymized benchmarks from similar projects (by language, size, framework)
- **Web dashboard** — browser-based interface for exploring analysis results, trends, and cross-repo views beyond what the CLI can present
- **Team-level insights** — correlate code ownership (CODEOWNERS) with test health to surface team-specific patterns and recommendations

### Architecture Documents

| Document | Coverage |
|----------|----------|
| [06-monorepo-graph-scaling.md](06-monorepo-graph-scaling.md) | Graph scaling patterns applicable to multi-repo |
| [11-ui-requirements.md](11-ui-requirements.md) | Dashboard and visualization requirements |
| [03-graph-storage-incremental-updates.md](03-graph-storage-incremental-updates.md) | Storage patterns for cross-repo graph data |

## Architecture Document Map

The following table maps every architecture document to its primary implementation phase:

| # | Document | Phase |
|---|----------|-------|
| 00 | Overview | All |
| 01 | Core Architecture | 1 |
| 02 | Graph Schema | 1 |
| 03 | Graph Storage and Incremental Updates | 1, 4 |
| 04 | Deterministic Test Identity | 1 |
| 05 | Insight Engine Framework | 1 |
| 06 | Monorepo Graph Scaling | 1, 4 |
| 07 | PR/CI Integration | 1 |
| 08 | Test Similarity and Structural Fingerprints | 1 |
| 09 | CLI Spec | 1 |
| 10 | JSON Artifact Schemas | 1 |
| 11 | UI Requirements | 4 |
| 12 | Risk and Coverage Taxonomy | 1 |
| 13 | Graph Traversal Algorithms | 1 |
| 14 | Evidence Scoring and Confidence Model | 1 |
| 15 | Edge Case Handling | 1 |
| 16 | Unified Graph Schema | 2 |
| 17 | Environment and Device Model | 3 |
| 19 | AI Scenario and Eval Model | 3 |
| 20 | Manual Coverage and Governance Model | 2, 3 |
| 21 | Behavior Surface Derivation | 2 |
| 22 | Reasoning Engine | 2 |
| 23 | Phased Implementation Roadmap | All |

## Key Decisions

1. **Additive phases.** Each phase adds capabilities without removing or replacing previous ones. A team that upgraded from Phase 1 to Phase 2 keeps all Phase 1 functionality unchanged. This means phase boundaries are about new capabilities, not migrations.

2. **Invariant principles.** Terrain's core principles — inference from code as default, explainability over magic, conservative under uncertainty — apply identically in every phase. Phase 4's multi-repo analysis follows the same confidence model as Phase 1's single-file signal detection.

3. **Independent utility.** No phase exists solely as scaffolding for a future phase. Phase 1 is a complete, useful product. Phase 2 makes it deeper. Phase 3 makes it broader. Phase 4 makes it organizational. A team that never progresses past Phase 1 still gets substantial value.
