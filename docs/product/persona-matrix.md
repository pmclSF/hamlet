# Persona Coverage Matrix

> **Status:** Current (validated 2026-03-15)
> **Purpose:** For each persona, show exactly which Terrain features serve them, at what depth, and what gaps remain. This is the persona-first view of the [Feature Matrix](feature-matrix.md).

## Backend Engineer

**Primary concern:** Transitive dependency risk, service boundary coverage, handler/route awareness.

### What's Implemented

| Feature | Depth | Command |
|---|---|---|
| Static analysis with Go/Python/Java support | Strong | `terrain analyze` |
| Handler and route surface detection | Strong | Inferred from code (HTTP handlers, route registrations) |
| Import graph with transitive tracing | Strong | `terrain impact`, `terrain debug graph` |
| High-fanout node detection | Strong | `terrain analyze`, `terrain insights` |
| Coverage confidence (LCOV, Istanbul) | Strong | `terrain analyze --coverage` |
| Runtime health (JUnit XML) | Strong | `terrain analyze --runtime` |
| Flaky/slow test detection | Strong | `terrain insights` (with runtime data) |
| Stability clustering | Strong | `terrain analyze` (shared dependency root cause) |
| Impact analysis with confidence decay | Strong | `terrain impact --base main` |
| Change-risk posture | Strong | `terrain impact` (4 dimensions) |
| Behavior redundancy detection | Useful | `terrain analyze`, `terrain insights` |
| Code unit extraction (functions, methods, classes) | Strong | Regex-based, 4 languages |

### What's Partial or Future

| Feature | Status | Gap |
|---|---|---|
| API contract testing detection (OpenAPI, gRPC) | Future | Cannot assess service boundary coverage |
| Database schema change tracking | Future | Cannot detect migration-related test gaps |
| Distributed system test topology | Future | Cannot model cross-service dependencies |
| Code unit extraction accuracy | Known limitation | Regex-based; may miss complex patterns |

### Hero Workflow

```bash
terrain analyze --root . --coverage lcov.info --runtime junit.xml
terrain impact --base main
terrain explain src/handlers/auth_handler.go
```

---

## Frontend Engineer

**Primary concern:** Component coverage, snapshot debt, framework fragmentation, E2E test burden.

### What's Implemented

| Feature | Depth | Command |
|---|---|---|
| JS/TS analysis (Jest, Vitest, Playwright, Cypress + 5 more) | Strong | `terrain analyze` |
| Snapshot-heavy test detection | Strong | Signal detection |
| Mock-heavy test detection | Strong | Signal detection |
| E2E concentration measurement | Strong | `terrain posture` (coverage_diversity dimension) |
| E2E-only coverage gaps | Strong | `terrain insights` |
| Istanbul coverage ingestion | Strong | `terrain analyze --coverage coverage-final.json` |
| Migration readiness (framework migration) | Strong | `terrain migration readiness` |
| Weak assertion detection | Strong | Signal detection |
| Duplicate test cluster detection | Strong | `terrain analyze`, `terrain insights` |
| Import graph construction | Strong | JS/TS imports, CommonJS require |

### What's Partial or Future

| Feature | Status | Gap |
|---|---|---|
| Visual regression detection (Storybook, Percy) | Future | Cannot assess visual test coverage |
| Component-level coverage mapping | Future | Coverage is file-level only |
| React/Vue/Svelte component tree awareness | Future | Import graph follows modules, not component composition |

### Hero Workflow

```bash
terrain analyze --root . --coverage coverage-final.json
terrain insights
terrain migration readiness
```

---

## AI / ML Engineer

**Primary concern:** Eval coverage, prompt validation, dataset integrity, model regression.

### What's Implemented

| Feature | Depth | Command |
|---|---|---|
| Scenario model (first-class validation target) | Strong | `.terrain/terrain.yaml` scenarios |
| Scenario loading from config | Strong | Pipeline loads and wires into graph |
| Prompt surface inference (JS/TS, Python) | Strong | `terrain analyze` (naming convention detection) |
| Dataset surface inference (JS/TS, Python) | Strong | `terrain analyze` (naming convention detection) |
| Scenario impact detection | Strong | `terrain impact` (changed surfaces → impacted scenarios) |
| Scenario explain | Strong | `terrain explain <scenario-id>` |
| Scenario duplication insights | Strong | `terrain insights` (>50% surface overlap) |
| AI list command | Strong | `terrain ai list` (scenarios, prompts, datasets, eval files) |
| AI doctor command | Strong | `terrain ai doctor` (5-point diagnostic) |
| Gauntlet artifact ingestion | Strong | `terrain analyze --gauntlet results.json` |
| Eval file detection by path | Useful | Detects eval/, evals/, __evals__/, benchmarks/ directories |
| Validation inventory (prompts, datasets counted) | Strong | `terrain analyze` |
| Graph reasoning path (CodeSurface → Scenario → Environment) | Strong | Full graph wiring |

### What's Partial or Future

| Feature | Status | Gap |
|---|---|---|
| Graph node population (Prompt, Dataset, Model, EvalMetric) | Partial | Types defined; not populated by inference yet |
| AI run / record / baseline commands | Future | Scaffolded with doc references |
| Eval framework auto-detection (12 frameworks) | **Implemented** | Config, deps, imports detected; scenarios auto-derived |
| Eval-specific signals (accuracy_regression, safety_failure) | Future | No eval-aware signal detectors |
| Model version tracking | Future | Cannot correlate evals to model versions |
| Dataset drift detection | Future | Cannot detect training/eval data divergence |

### Hero Workflow

```bash
terrain ai list
terrain ai doctor
terrain impact --base main    # shows impacted scenarios
terrain explain scenario:custom:safety-check
terrain analyze --gauntlet gauntlet-results.json
```

---

## SRE / Platform Engineer

**Primary concern:** CI duration, flaky tests, test selection efficiency, environment matrix.

### What's Implemented

| Feature | Depth | Command |
|---|---|---|
| Test selection for PR builds | Strong | `terrain select-tests`, `terrain pr` |
| Flaky/slow test detection | Strong | `terrain analyze --runtime` |
| Stability clustering (shared root causes) | Strong | `terrain analyze` |
| Environment/device matrix analysis | Strong | `terrain analyze` (gaps, concentrations, recommendations) |
| CI optimization potential | Strong | `terrain analyze` (duplicate removal, fanout reduction) |
| GitHub Actions integration | Strong | `.github/actions/terrain-impact/` |
| PR comment with impact summary | Strong | Composite action |
| Metrics export for dashboards | Strong | `terrain metrics --json` |
| Policy enforcement with exit codes | Strong | `terrain policy check` (0/1/2) |
| Benchmark export | Strong | `terrain export benchmark` |
| Impact analysis with fallback ladder | Strong | `terrain impact` (exact → structural → full suite) |

### What's Partial or Future

| Feature | Status | Gap |
|---|---|---|
| GitLab/Jenkins/CircleCI native integration | Future | CLI works; no native plugin |
| Test parallelization recommendations | Future | Cannot suggest shard strategies |
| Historical CI duration tracking | Future | Per-snapshot only |
| Cost modeling (compute minutes) | Future | Cannot estimate CI cost savings |

### Hero Workflow

```bash
terrain analyze --root . --runtime junit.xml --coverage lcov.info
terrain select-tests --base main --json
terrain metrics --json
terrain policy check
```

---

## QA / SDET

**Primary concern:** Test health trends, policy compliance, manual coverage overlay, portfolio management.

### What's Implemented

| Feature | Depth | Command |
|---|---|---|
| Executive summary (leadership view) | Strong | `terrain summary` |
| Quality posture (5 dimensions with evidence) | Strong | `terrain posture` |
| Portfolio intelligence (cost, breadth, leverage) | Strong | `terrain portfolio` |
| Manual coverage overlay | Strong | `.terrain/terrain.yaml` manual_coverage |
| Policy enforcement (6 rule types) | Strong | `terrain policy check` |
| Snapshot comparison (trend tracking) | Strong | `terrain compare` |
| Focus / next actions (prioritized) | Strong | `terrain focus` |
| Behavior redundancy (cross-framework overlap) | Strong | `terrain analyze`, `terrain insights` |
| All signal detectors (22 signal types) | Strong | Full signal pipeline |
| Benchmark export (privacy-safe) | Strong | `terrain export benchmark` |
| Entity drill-down | Strong | `terrain show test/unit/owner/finding` |
| Migration readiness | Strong | `terrain migration readiness/blockers/preview` |
| Scenario duplication detection | Useful | `terrain insights` |
| Environment matrix coverage | Strong | `terrain analyze` |

### What's Partial or Future

| Feature | Status | Gap |
|---|---|---|
| TestRail/Jira/Xray integration | Future | Manual coverage via YAML only |
| Automated trend alerting | Future | Compare requires explicit snapshots |
| Multi-repo aggregation | Future | Portfolio is single-repo |
| Test case management sync | Future | No bidirectional link to external QA tools |

### Hero Workflow

```bash
terrain analyze --root . --coverage lcov.info --runtime junit.xml
terrain summary
terrain posture
terrain portfolio
terrain compare
terrain policy check
```

---

## Coverage Summary

| Persona | Implemented | Partial | Future | Coverage |
|---|---|---|---|---|
| **Backend Engineer** | 12 | 0 | 4 | 75% |
| **Frontend Engineer** | 10 | 0 | 3 | 77% |
| **AI / ML Engineer** | 13 | 1 | 6 | 65% |
| **SRE / Platform** | 11 | 0 | 4 | 73% |
| **QA / SDET** | 14 | 0 | 4 | 78% |

All five personas have strong coverage of their primary workflows. The AI/ML persona has the most future items because it is the newest persona with the most recent infrastructure additions.
