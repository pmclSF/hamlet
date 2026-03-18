# Feature Matrix

> **Status:** Current (validated against implementation 2026-03-15)
> **Purpose:** Map Terrain's capabilities to the personas who use them. Each cell shows the support level for that persona's specific workflow. Status column distinguishes implemented, partial, and future.

**See also:** [Persona Matrix](persona-matrix.md), [Persona Journeys](../architecture/17-persona-journeys.md), [Canonical User Journeys](canonical-user-journeys.md)

## Legend

| Symbol | Meaning |
|--------|---------|
| **Strong** | Capability is fully implemented and directly serves this persona's key workflow |
| **Useful** | Capability is implemented and relevant, but not purpose-built for this persona |
| **Partial** | Capability exists but has gaps for this persona's specific needs |
| **---** | Not relevant to this persona's primary workflows |
| **Future** | Not yet implemented; identified as a future extension |

## Core Capabilities x Personas

| Capability | Status | Frontend | Backend | AI / ML | SRE / Platform | QA / SDET |
|---|---|---|---|---|---|---|
| **Static analysis** (`analyze`) | Implemented | Strong | Strong | Strong | Useful | Strong |
| **Validation inventory** (in `analyze`) | Implemented | Strong | Strong | Strong | Useful | Strong |
| **Prioritized findings** (`insights`) | Implemented | Strong | Strong | Strong | Useful | Strong |
| **Change impact** (`impact`) | Implemented | Strong | Strong | Strong | Strong | Strong |
| **Scenario impact** (in `impact`) | Implemented | --- | --- | Strong | --- | Useful |
| **Explainability** (`explain`) | Implemented | Strong | Strong | Strong | Strong | Strong |
| **Scenario explain** (in `explain`) | Implemented | --- | --- | Strong | --- | Useful |
| **Test selection** (`select-tests`) | Implemented | Strong | Strong | Partial | Strong | Useful |
| **PR analysis** (`pr`) | Implemented | Strong | Strong | Partial | Strong | Useful |
| **Executive summary** (`summary`) | Implemented | Useful | Useful | Useful | Useful | Strong |
| **Quality posture** (`posture`) | Implemented | Useful | Useful | Useful | Useful | Strong |
| **Portfolio view** (`portfolio`) | Implemented | --- | --- | --- | Useful | Strong |
| **Metrics export** (`metrics`) | Implemented | --- | --- | --- | Strong | Strong |
| **Snapshot comparison** (`compare`) | Implemented | Useful | Useful | Useful | Useful | Strong |
| **Migration readiness** (`migration`) | Implemented | Strong | Useful | --- | --- | Strong |
| **Policy enforcement** (`policy`) | Implemented | Useful | Useful | Useful | Useful | Strong |
| **Focus / next actions** (`focus`) | Implemented | Useful | Useful | Useful | Useful | Strong |
| **Benchmark export** (`export benchmark`) | Implemented | --- | --- | --- | Useful | Strong |
| **AI list** (`ai list`) | Implemented | --- | --- | Strong | --- | Useful |
| **AI doctor** (`ai doctor`) | Implemented | --- | --- | Strong | --- | Useful |
| **AI run** (`ai run`) | Future | --- | --- | Future | --- | --- |
| **AI record** (`ai record`) | Future | --- | --- | Future | --- | --- |
| **AI baseline** (`ai baseline`) | Future | --- | --- | Future | --- | --- |

## Analysis Engines x Personas

| Engine / Detector | Status | Frontend | Backend | AI / ML | SRE / Platform | QA / SDET |
|---|---|---|---|---|---|---|
| **Weak assertion detection** | Implemented | Strong | Strong | Useful | --- | Strong |
| **Mock-heavy detection** | Implemented | Strong | Strong | Useful | --- | Strong |
| **Snapshot-heavy detection** | Implemented | Strong | --- | --- | --- | Strong |
| **Untested export detection** | Implemented | Strong | Strong | Useful | Useful | Strong |
| **Coverage threshold check** | Implemented | Strong | Strong | Useful | Strong | Strong |
| **Coverage blind spot detection** | Implemented | Strong | Strong | Useful | Useful | Strong |
| **Slow test detection** | Implemented | Useful | Strong | Useful | Strong | Strong |
| **Flaky test detection** | Implemented | Useful | Strong | Useful | Strong | Strong |
| **Skipped test detection** | Implemented | Useful | Useful | Useful | Strong | Strong |
| **Dead test detection** | Implemented | Useful | Useful | Useful | Useful | Strong |
| **Unstable suite detection** | Implemented | Useful | Strong | Useful | Strong | Strong |
| **Deprecated pattern detection** | Implemented | Strong | Useful | --- | --- | Strong |
| **Dynamic test generation detection** | Implemented | Useful | Useful | Useful | --- | Useful |
| **Custom matcher risk** | Implemented | Strong | Useful | --- | --- | Strong |
| **Policy evaluation** | Implemented | Useful | Useful | Useful | Useful | Strong |
| **Prompt surface detection** | Implemented | --- | --- | Strong | --- | Useful |
| **Dataset surface detection** | Implemented | --- | --- | Strong | --- | Useful |
| **Scenario duplication detection** | Implemented | --- | --- | Strong | --- | Useful |
| **Behavior redundancy detection** | Implemented | Useful | Useful | Useful | Useful | Strong |
| **Stability clustering** | Implemented | Useful | Strong | Useful | Strong | Strong |
| **Environment matrix analysis** | Implemented | --- | --- | Useful | Strong | Strong |

## AI / ML Capabilities

| Capability | Status | Details |
|---|---|---|
| Scenario model (`Scenario` type) | **Implemented** | First-class validation target alongside tests |
| Scenario loading from `.terrain/terrain.yaml` | **Implemented** | Scenarios with name, category, framework, owner, surfaces |
| Prompt surface inference | **Implemented** | JS/TS and Python naming convention detection |
| Dataset surface inference | **Implemented** | JS/TS and Python naming convention detection |
| Scenario impact detection | **Implemented** | `findImpactedScenarios()` — changed surfaces mapped to scenarios |
| Scenario explain | **Implemented** | `ExplainScenario()` — verdict with changed surface list |
| Scenario duplication insights | **Implemented** | Detects >50% surface overlap between scenario pairs |
| Gauntlet artifact ingestion | **Implemented** | `--gauntlet` flag, signal generation for failures/regressions |
| AI list command | **Implemented** | Lists scenarios, prompts, datasets, eval files |
| AI doctor command | **Implemented** | 5-point diagnostic: scenarios, prompts, datasets, eval files, graph wiring |
| Graph node types (Prompt, Dataset, Model, EvalMetric) | **Partial** | Types defined in schema; not yet populated by inference |
| Eval framework auto-detection (12 frameworks) | **Implemented** | Config files, dependency manifests, and source imports |
| Auto-scenario derivation from eval files | **Implemented** | Eval test files + AI imports → scenarios without YAML |
| AI run / record / baseline | **Future** | Scaffolded with doc references |
| Eval-specific signals | **Future** | eval_safety_failure, eval_accuracy_regression, etc. |
| Model version tracking | **Future** | Correlate eval results to model versions |
| Dataset drift detection | **Future** | Detect training/eval data divergence |

## Language & Framework Support

| Language / Framework | Status | Frontend | Backend | AI / ML | SRE / Platform | QA / SDET |
|---|---|---|---|---|---|---|
| **JavaScript / TypeScript** | Implemented | Primary | --- | Useful | --- | Primary |
| Jest | Implemented | Strong | --- | --- | --- | Strong |
| Vitest | Implemented | Strong | --- | Useful | --- | Strong |
| Mocha | Implemented | Strong | --- | --- | --- | Strong |
| Jasmine | Implemented | Useful | --- | --- | --- | Useful |
| Playwright | Implemented | Strong | --- | --- | --- | Strong |
| Cypress | Implemented | Strong | --- | --- | --- | Strong |
| Puppeteer | Implemented | Useful | --- | --- | --- | Useful |
| WebdriverIO | Implemented | Useful | --- | --- | --- | Useful |
| TestCafe | Implemented | Useful | --- | --- | --- | Useful |
| Node Test | Implemented | Useful | --- | --- | --- | Useful |
| **Go** | Implemented | --- | Primary | --- | Useful | Useful |
| go-testing | Implemented | --- | Strong | --- | Useful | Useful |
| **Python** | Implemented | --- | Strong | Primary | --- | Useful |
| pytest | Implemented | --- | Strong | Strong | --- | Useful |
| unittest | Implemented | --- | Useful | Useful | --- | Useful |
| nose2 | Implemented | --- | Useful | Useful | --- | Useful |
| **Java** | Implemented | --- | Strong | --- | --- | Useful |
| JUnit 4 | Implemented | --- | Strong | --- | --- | Useful |
| JUnit 5 | Implemented | --- | Strong | --- | --- | Useful |
| TestNG | Implemented | --- | Strong | --- | --- | Useful |

## Data Source Support

| Data Source | Format | Status | Primary Personas |
|---|---|---|---|
| Source code (static) | Any supported language | **Implemented** | All |
| LCOV coverage | `*.lcov`, `lcov.info` | **Implemented** | Backend, QA/SDET, SRE |
| Istanbul coverage | `coverage-final.json` | **Implemented** | Frontend, QA/SDET |
| JUnit XML runtime | `*.xml` | **Implemented** | Backend, SRE, QA/SDET |
| Jest/Vitest JSON runtime | `*.json` | **Implemented** | Frontend, SRE |
| Gauntlet eval results | `*.json` (via `--gauntlet`) | **Implemented** | AI/ML |
| CODEOWNERS | `CODEOWNERS`, `.github/CODEOWNERS` | **Implemented** | QA/SDET, SRE |
| Manual coverage overlay | `.terrain/terrain.yaml` | **Implemented** | QA/SDET |
| Scenario definitions | `.terrain/terrain.yaml` | **Implemented** | AI/ML, QA/SDET |
| Policy rules | `.terrain/policy.yaml` | **Implemented** | QA/SDET, SRE |
| Git history | Local `.git` | **Implemented** | All (for impact/PR) |

## Dependency Graph Capabilities

| Capability | Status | Primary Personas |
|---|---|---|
| Import graph construction | **Implemented** (JS, Go, Python, Java) | Backend, Frontend |
| Transitive dependency tracing | **Implemented** | Backend |
| High-fanout node detection | **Implemented** (threshold: 10 dependents) | Backend, SRE |
| Duplicate test cluster detection | **Implemented** (0.60 similarity threshold) | QA/SDET, Frontend |
| Behavior redundancy detection | **Implemented** | All |
| Structural reverse coverage | **Implemented** | All |
| Change-scoped impact profiling | **Implemented** | All (via `impact`) |
| Scenario-to-surface coverage | **Implemented** | AI/ML (via `impact`) |
| Confidence decay by path length | **Implemented** | Backend, SRE |
| Environment matrix analysis | **Implemented** | SRE, QA/SDET |

## Current Gaps by Persona

### Frontend Engineer
| Gap | Impact | Status |
|---|---|---|
| No visual regression detection (Storybook, Percy, Chromatic) | Cannot assess visual test coverage | Future |
| No component-level coverage mapping | Coverage is file-level, not component-level | Future |
| No React/Vue/Svelte component tree awareness | Import graph doesn't follow component composition | Future |

### Backend Engineer
| Gap | Impact | Status |
|---|---|---|
| No database schema change tracking | Cannot detect migration-related test gaps | Future |
| No API contract testing detection (OpenAPI, gRPC) | Cannot assess service boundary coverage | Future |
| No distributed system test topology | Cannot model cross-service test dependencies | Future |
| Code unit extraction is regex-based | May miss complex export patterns | Known limitation |

### AI / ML Engineer
| Gap | Impact | Status |
|---|---|---|
| Eval framework auto-detection | 12 frameworks detected via config, deps, and imports | **Implemented** |
| Auto-scenario derivation from eval test files | No YAML required for eval test files | **Implemented** |
| No eval-specific signals | No accuracy_regression, safety_failure signals | Future |
| No model version tracking | Cannot correlate evals to model versions | Future |
| No dataset drift detection | Cannot detect data divergence | Future |
| Graph node types (Prompt, Dataset, Model, EvalMetric) not populated | Types exist but not wired into inference | Partial |

### SRE / Platform Engineer
| Gap | Impact | Status |
|---|---|---|
| No CI plugin model | CLI step, not native plugin | By design |
| No test parallelization recommendations | Cannot suggest shard strategies | Future |
| No historical CI duration tracking | Per-snapshot only | Future |
| No cost modeling | Cannot estimate CI cost savings | Future |

### QA / SDET
| Gap | Impact | Status |
|---|---|---|
| No TestRail/Jira/Xray integration | Manual coverage via YAML only | Future |
| No automated trend alerting | Compare requires explicit snapshots | Future |
| No multi-repo aggregation | Portfolio is single-repo | Future |

## CI Integration Matrix

| Integration Point | Status | Notes |
|---|---|---|
| GitHub Actions composite action | **Implemented** | `.github/actions/terrain-impact/` |
| PR comment with impact summary | **Implemented** | Via `peter-evans/create-or-update-comment` |
| Job summary output | **Implemented** | Writes to `$GITHUB_STEP_SUMMARY` |
| Artifact upload | **Implemented** | `terrain-impact.json` via `actions/upload-artifact` |
| Test selection for PR builds | **Implemented** | `terrain select-tests` with fallback ladder |
| Exit code signaling | **Implemented** | 0 (clean), 1 (error), 2 (policy violation) |
| GitLab CI | **Future** | CLI works; no native integration |
| Jenkins | **Future** | CLI works; no native integration |
| CircleCI | **Future** | CLI works; no native integration |

## Output Format Matrix

| Command | Terminal | JSON | Status |
|---|---|---|---|
| `analyze` | Yes | Yes | Implemented |
| `insights` | Yes | Yes | Implemented |
| `impact` | Yes | Yes | Implemented |
| `explain` | Yes | Yes | Implemented |
| `pr` | Yes | Yes | Implemented |
| `summary` | Yes | Yes | Implemented |
| `posture` | Yes | Yes | Implemented |
| `portfolio` | Yes | Yes | Implemented |
| `metrics` | Yes | Yes | Implemented |
| `compare` | Yes | Yes | Implemented |
| `select-tests` | Yes | Yes | Implemented |
| `policy check` | Yes | Yes | Implemented |
| `migration *` | Yes | Yes | Implemented |
| `focus` | Yes | Yes | Implemented |
| `show` | Yes | Yes | Implemented |
| `export benchmark` | --- | Yes | Implemented |
| `debug *` | Yes | Yes | Implemented |
| `ai list` | Yes | Yes | Implemented |
| `ai doctor` | Yes | Yes | Implemented |
| `ai run` | --- | --- | Future |
