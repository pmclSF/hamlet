# Feature Matrix

> **Status:** Current
> **Purpose:** Map Terrain's capabilities to the personas who use them. Each cell shows whether the capability is implemented, partial, or planned — and at what support level for that persona's specific workflow.

**See also:** [Persona Journeys](../architecture/17-persona-journeys.md), [Canonical User Journeys](canonical-user-journeys.md), [CLI Spec](../architecture/09-cli-spec.md)

## Legend

| Symbol | Meaning |
|--------|---------|
| **Strong** | Capability is fully implemented and directly serves this persona's key workflow |
| **Useful** | Capability is implemented and relevant, but not purpose-built for this persona |
| **Partial** | Capability exists but has gaps for this persona's specific needs |
| **—** | Not relevant to this persona's primary workflows |
| **Planned** | Not yet implemented; identified as a future extension |

## Core Capabilities × Personas

| Capability | Frontend Engineer | Backend Engineer | Mobile / Device | QA / SDET | SRE / Platform | AI / ML Engineer |
|---|---|---|---|---|---|---|
| **Static analysis** (`analyze`) | Strong | Strong | Strong | Strong | Useful | Useful |
| **Prioritized findings** (`insights`) | Strong | Strong | Strong | Strong | Useful | Useful |
| **Change impact** (`impact`) | Strong | Strong | Strong | Strong | Strong | Useful |
| **Explainability** (`explain`) | Strong | Strong | Strong | Strong | Strong | Useful |
| **Test selection** (`select-tests`) | Strong | Strong | Useful | Useful | Strong | Partial |
| **PR analysis** (`pr`) | Strong | Strong | Useful | Useful | Strong | Partial |
| **Executive summary** (`summary`) | Useful | Useful | Useful | Strong | Useful | Useful |
| **Quality posture** (`posture`) | Useful | Useful | Useful | Strong | Useful | Useful |
| **Portfolio view** (`portfolio`) | — | — | — | Strong | Useful | — |
| **Metrics export** (`metrics`) | — | — | — | Strong | Strong | — |
| **Snapshot comparison** (`compare`) | Useful | Useful | Useful | Strong | Useful | Useful |
| **Migration readiness** (`migration`) | Strong | Useful | Partial | Strong | — | — |
| **Policy enforcement** (`policy`) | Useful | Useful | Useful | Strong | Useful | Useful |
| **Focus / next actions** (`focus`) | Useful | Useful | Useful | Strong | Useful | Useful |
| **Benchmark export** (`export benchmark`) | — | — | — | Strong | Useful | — |

## Analysis Engines × Personas

| Engine / Detector | Frontend Engineer | Backend Engineer | Mobile / Device | QA / SDET | SRE / Platform | AI / ML Engineer |
|---|---|---|---|---|---|---|
| **Weak assertion detection** | Strong | Strong | Useful | Strong | — | Useful |
| **Mock-heavy detection** | Strong | Strong | Useful | Strong | — | Useful |
| **Snapshot-heavy detection** | Strong | — | — | Strong | — | — |
| **Untested export detection** | Strong | Strong | Strong | Strong | Useful | Useful |
| **Coverage threshold check** | Strong | Strong | Strong | Strong | Strong | Useful |
| **Coverage blind spot detection** | Strong | Strong | Useful | Strong | Useful | Useful |
| **Slow test detection** | Useful | Strong | Strong | Strong | Strong | Useful |
| **Flaky test detection** | Useful | Strong | Strong | Strong | Strong | Useful |
| **Skipped test detection** | Useful | Useful | Strong | Strong | Strong | Useful |
| **Dead test detection** | Useful | Useful | Useful | Strong | Useful | Useful |
| **Unstable suite detection** | Useful | Strong | Strong | Strong | Strong | Useful |
| **Deprecated pattern detection** | Strong | Useful | Useful | Strong | — | — |
| **Dynamic test generation detection** | Useful | Useful | — | Useful | — | Useful |
| **Custom matcher risk** | Strong | Useful | Useful | Strong | — | — |
| **Policy evaluation** | Useful | Useful | Useful | Strong | Useful | Useful |

## Language & Framework Support × Personas

| Language / Framework | Frontend Engineer | Backend Engineer | Mobile / Device | QA / SDET | SRE / Platform | AI / ML Engineer |
|---|---|---|---|---|---|---|
| **JavaScript / TypeScript** | Primary | — | Useful | Primary | — | — |
| Jest | Strong | — | — | Strong | — | — |
| Vitest | Strong | — | Useful | Strong | — | — |
| Mocha | Strong | — | — | Strong | — | — |
| Jasmine | Useful | — | — | Useful | — | — |
| Playwright | Strong | — | — | Strong | — | — |
| Cypress | Strong | — | — | Strong | — | — |
| Puppeteer | Useful | — | — | Useful | — | — |
| WebdriverIO | Useful | — | Useful | Useful | — | — |
| TestCafe | Useful | — | — | Useful | — | — |
| **Go** | — | Primary | — | Useful | Useful | — |
| go-testing | — | Strong | — | Useful | Useful | — |
| **Python** | — | Strong | — | Useful | — | Primary |
| pytest | — | Strong | — | Useful | — | Strong |
| unittest | — | Useful | — | Useful | — | Useful |
| nose2 | — | Useful | — | Useful | — | Useful |
| **Java** | — | Strong | Useful | Useful | — | — |
| JUnit 4 | — | Strong | Useful | Useful | — | — |
| JUnit 5 | — | Strong | Useful | Useful | — | — |
| TestNG | — | Strong | Useful | Useful | — | — |

## Data Source Support

| Data Source | Format | Status | Primary Personas |
|---|---|---|---|
| Source code (static) | Any supported language | **Implemented** | All |
| LCOV coverage | `*.lcov`, `lcov.info` | **Implemented** | Backend, QA/SDET, SRE |
| Istanbul coverage | `coverage-final.json` | **Implemented** | Frontend, QA/SDET |
| JUnit XML runtime | `*.xml` | **Implemented** | Backend, SRE, QA/SDET |
| Jest/Vitest JSON runtime | `*.json` | **Implemented** | Frontend, SRE |
| CODEOWNERS | `CODEOWNERS`, `.github/CODEOWNERS` | **Implemented** | QA/SDET, SRE |
| Manual coverage overlay | `.terrain/terrain.yaml` | **Implemented** | QA/SDET |
| Policy rules | `.terrain/policy.yaml` | **Implemented** | QA/SDET, SRE |
| Git history | Local `.git` | **Implemented** | All (for impact/PR) |

## Dependency Graph Capabilities

| Capability | Status | Primary Personas |
|---|---|---|
| Import graph construction | **Implemented** (JS, Go, Python, Java) | Backend, Frontend |
| Transitive dependency tracing | **Implemented** | Backend |
| High-fanout node detection | **Implemented** (threshold: 10 dependents) | Backend, SRE |
| Duplicate test cluster detection | **Implemented** (0.60 similarity threshold) | QA/SDET, Frontend |
| Structural reverse coverage | **Implemented** | All |
| Change-scoped impact profiling | **Implemented** | All (via `impact`) |
| Confidence decay by path length | **Implemented** | Backend, SRE |

## Current Gaps by Persona

Terrain is honest about what it cannot do. These gaps are documented here so that users know the boundaries and so that future work is prioritized by real need.

### Frontend Engineer
| Gap | Impact | Status |
|---|---|---|
| No visual regression framework detection (Storybook, Percy, Chromatic) | Cannot assess visual test coverage | Planned |
| No component-level coverage mapping | Coverage is file-level, not component-level | Planned |
| No React/Vue/Svelte component tree awareness | Import graph doesn't follow component composition | Planned |

### Backend Engineer
| Gap | Impact | Status |
|---|---|---|
| No database schema change tracking | Cannot detect migration-related test gaps | Planned |
| No API contract testing detection (OpenAPI, gRPC) | Cannot assess service boundary coverage | Planned |
| No distributed system test topology | Cannot model cross-service test dependencies | Planned |
| Code unit extraction is regex-based | May miss complex export patterns (destructured, re-exported) | Known limitation |

### Mobile / Device-Sensitive Engineer
| Gap | Impact | Status |
|---|---|---|
| No device/environment matrix awareness | Cannot model "runs on iOS but not Android" | Planned |
| No CI matrix analysis | Cannot parse GitHub Actions matrix or device farm configs | Planned |
| No platform-conditional skip classification | Skip detection is pattern-matching only | Planned |
| No test duration modeling by device type | Cannot estimate device farm cost impact | Planned |

### QA / SDET
| Gap | Impact | Status |
|---|---|---|
| No direct TestRail/Jira/Xray integration | Manual coverage is declared in YAML, not synced | Planned |
| No automated trend alerting | Compare requires two explicit snapshots | Planned |
| No multi-repo aggregation dashboard | Portfolio view is single-repo only | Planned |
| No test case management system sync | No bidirectional link to external QA tools | Planned |

### SRE / Platform Engineer
| Gap | Impact | Status |
|---|---|---|
| No CI plugin model | Terrain is invoked as a CLI step, not a native plugin | By design |
| No test parallelization recommendations | Cannot suggest optimal shard strategies | Planned |
| No historical CI duration tracking | Runtime data is per-snapshot, not time-series | Planned |
| No cost modeling (compute minutes, device farm billing) | Cannot estimate CI cost savings | Planned |

### AI / ML Engineer
| Gap | Impact | Status |
|---|---|---|
| No eval-specific semantics | Eval suites are treated as regular test suites | Planned |
| No accuracy/metric threshold tracking | Cannot monitor model performance thresholds | Planned |
| No model version awareness | Cannot correlate eval results to model versions | Planned |
| No dataset drift detection | Cannot detect training/eval data divergence | Planned |
| No LLM-specific patterns (prompt testing, safety evals) | Cannot distinguish behavioral evals from functional tests | Planned |

## CI Integration Matrix

| Integration Point | Status | Notes |
|---|---|---|
| GitHub Actions composite action | **Implemented** | `.github/actions/terrain-impact/` |
| PR comment with impact summary | **Implemented** | Via `peter-evans/create-or-update-comment` |
| Job summary output | **Implemented** | Writes to `$GITHUB_STEP_SUMMARY` |
| Artifact upload | **Implemented** | `terrain-impact.json` via `actions/upload-artifact` |
| Test selection for PR builds | **Implemented** | `terrain select-tests` with 5-level fallback |
| Exit code signaling | **Implemented** | 0 (clean), 1 (error), 2 (policy violation) |
| GitLab CI | **Not implemented** | CLI works; no native integration |
| Jenkins | **Not implemented** | CLI works; no native integration |
| CircleCI | **Not implemented** | CLI works; no native integration |

## Output Format Matrix

| Command | Terminal | JSON | Markdown | Comment | Annotation |
|---|---|---|---|---|---|
| `analyze` | Yes | Yes | — | — | — |
| `insights` | Yes | Yes | — | — | — |
| `impact` | Yes | Yes | — | — | — |
| `explain` | Yes | Yes | — | — | — |
| `pr` | Yes | Yes | Yes | Yes | Yes |
| `summary` | Yes | Yes | — | — | — |
| `posture` | Yes | Yes | — | — | — |
| `portfolio` | Yes | Yes | — | — | — |
| `metrics` | Yes | Yes | — | — | — |
| `compare` | Yes | Yes | — | — | — |
| `select-tests` | Yes | Yes | — | — | — |
| `policy check` | Yes | Yes | — | — | — |
| `migration *` | Yes | Yes | — | — | — |
| `focus` | Yes | Yes | — | — | — |
| `show` | Yes | Yes | — | — | — |
| `export benchmark` | — | Yes | — | — | — |
| `debug *` | Yes | Yes | — | — | — |
