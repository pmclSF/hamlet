# Excellence Gap Audit

> **Date:** 2026-03-16
> **Method:** Code-level verification of all documented personas, user stories, and feature claims against actual implementation in `cmd/`, `internal/`, and test suites.
> **Source of truth:** docs/product/terrain-overview.md, docs/product/feature-matrix.md, docs/product/persona-matrix.md, docs/product/canonical-user-journeys.md, docs/architecture/17-persona-journeys.md, README.md, docs/cli-spec.md

## Summary

| Status | Count |
|--------|-------|
| Fully implemented | 42 |
| Partial / degraded | 8 |
| Missing (documented as planned) | 12 |
| Stale / inconsistent across docs | 4 |

---

## Canonical Workflows

| Journey | Command | Status | Evidence | Gap |
|---------|---------|--------|----------|-----|
| Understand test system | `terrain analyze` | **Implemented** | `runAnalyze()`, all renderers, golden tests | None |
| Improve test suite | `terrain insights` | **Implemented** | `runInsights()`, 10 finding builders | None |
| Understand a change | `terrain impact` | **Implemented** | `runImpact()`, BFS+confidence, fallback ladder | None |
| Explain a decision | `terrain explain` | **Implemented** | `runExplain()`, 6 entity types + scenario explain | None |

---

## Per-Persona Gap Analysis

### Frontend Engineer

| User Story | Command | Status | Evidence | Gap | Priority |
|---|---|---|---|---|---|
| See framework fragmentation | `analyze` | **Implemented** | Framework detection, 17 frameworks | None | - |
| Check coverage of changed components | `impact --show units` | **Implemented** | ImpactedCodeUnit with CoverageTypes | None | - |
| Find snapshot-heavy files | `insights` | **Implemented** | SnapshotHeavyDetector | None | - |
| Run only affected tests on PR | `select-tests` | **Implemented** | ProtectiveTestSet with fallback | None | - |
| Visual regression detection | - | **Missing** | No Storybook/Percy/Chromatic detection | No detector for visual regression frameworks | P2 |
| Component-level coverage mapping | - | **Missing** | Coverage is file-level via import graph | No component tree traversal | P2 |
| React/Vue/Svelte component tree awareness | - | **Missing** | Import graph follows modules, not JSX composition | Would require framework-specific AST analysis | P2 |

### Backend Engineer

| User Story | Command | Status | Evidence | Gap | Priority |
|---|---|---|---|---|---|
| Find high-fanout modules | `insights`, `debug fanout` | **Implemented** | `AnalyzeFanout()`, Kahn's algorithm | None | - |
| Trace change impact through deps | `impact` | **Implemented** | BFS with confidence decay, 0.85 length decay | None | - |
| Find untested service boundaries | `analyze` | **Implemented** | `untestedExport` signal, coverage band analysis | None | - |
| Check coverage with runtime data | `analyze --runtime` | **Implemented** | JUnit XML + Jest JSON ingestion | None | - |
| HTTP handler/route detection | `analyze` | **Implemented** | Handler/route surface detection for JS, Go, Python, Java | None | - |
| Database schema change tracking | - | **Missing** | No migration file detection | No detector for schema changes | P2 |
| API contract testing detection | - | **Missing** | No OpenAPI/gRPC schema awareness | No contract test framework detection | P2 |
| Code unit extraction accuracy | - | **Partial** | Regex-based, no AST | May miss destructured re-exports, complex patterns | P1 |

### Mobile / Device Engineer

| User Story | Command | Status | Evidence | Gap | Priority |
|---|---|---|---|---|---|
| Audit skip debt | `analyze` | **Implemented** | SkippedTest signal (requires `--runtime`) | Static skip detection absent | P1 |
| Platform-conditional gap finding | `insights` | **Partial** | Environment matrix exists but requires EnvironmentClass nodes | Matrix analysis returns empty without explicit env metadata | P1 |
| Select tests for platform build | `impact` | **Implemented** | Standard impact analysis works | No platform-aware filtering | P2 |
| Device/environment matrix awareness | `analyze` | **Partial** | `internal/matrix/` exists, NodeDeviceConfig defined | Only useful if `.terrain/terrain.yaml` declares environments or CI config is parsed | P1 |

### QA / SDET

| User Story | Command | Status | Evidence | Gap | Priority |
|---|---|---|---|---|---|
| Full system health assessment | `analyze` | **Implemented** | 22 signals, 5 posture dimensions, 18 measurements | None | - |
| Manual coverage overlay | `analyze` | **Implemented** | `.terrain/terrain.yaml` manual_coverage loaded in pipeline | None | - |
| Policy enforcement | `policy check` | **Implemented** | 6 rule types, exit code 0/1/2 | None | - |
| Trend tracking | `compare` | **Implemented** | Auto-detects recent snapshots, shows posture delta | None | - |
| Executive reporting | `summary` | **Implemented** | `runSummary()` with risk, trends, benchmark readiness | None | - |
| TestRail/Jira/Xray integration | - | **Missing** | Manual coverage is YAML-only | No external tool sync | P2 |
| Automated trend alerting | - | **Missing** | Compare requires explicit execution | No threshold-based alerting | P2 |
| Multi-repo aggregation | - | **Missing** | Portfolio is single-repo | No cross-repo dashboard | P2 |

### SRE / Platform Engineer

| User Story | Command | Status | Evidence | Gap | Priority |
|---|---|---|---|---|---|
| PR test selection | `pr`, `select-tests` | **Implemented** | Both produce ProtectiveTestSet | `select-tests` is redundant with `impact --show selected` | P2 |
| Identify slow tests | `analyze --runtime` | **Implemented** | SlowTestDetector with configurable threshold | Requires `--runtime` artifacts | - |
| Identify flaky tests | `analyze --runtime` | **Implemented** | FlakyTestDetector with retry detection | Requires `--runtime` artifacts | - |
| CI optimization assessment | `analyze` | **Implemented** | CI optimization section in report | None | - |
| System metrics for monitoring | `metrics --json` | **Implemented** | 6 metric categories, JSON output | None | - |
| GitLab/Jenkins/CircleCI integration | - | **Missing** | CLI works; no native plugin | By design — CLI-first | P2 |
| Test parallelization recommendations | - | **Missing** | No shard strategy analysis | Would need runtime duration data | P2 |
| Historical CI duration tracking | - | **Missing** | Per-snapshot only | Would need time-series storage | P2 |

### AI / ML Engineer

| User Story | Command | Status | Evidence | Gap | Priority |
|---|---|---|---|---|---|
| Audit eval suite | `analyze` | **Implemented** | Validation inventory shows scenarios, prompts, datasets | None | - |
| Find eval gaps | `insights` | **Implemented** | Scenario duplication finding | None | - |
| Check impact of model changes | `impact` | **Implemented** | `findImpactedScenarios()` maps surfaces to scenarios | None | - |
| List scenarios/prompts/datasets | `ai list` | **Implemented** | Shows frameworks, auto-derived scenarios, surfaces | None | - |
| Validate eval setup | `ai doctor` | **Implemented** | 5-point diagnostic | None | - |
| Execute eval scenarios | `ai run` | **Implemented** | Detects framework, builds and executes command | None | - |
| Record eval baseline | `ai record` | **Implemented** | Saves to `.terrain/baselines/latest.json` | None | - |
| Manage baselines | `ai baseline` | **Implemented** | Reads and displays baseline | None | - |
| Auto-detect eval frameworks | `ai list`, pipeline | **Implemented** | `aidetect.Detect()` — 12 frameworks, 3 detection methods | None | - |
| Auto-derive scenarios from code | pipeline | **Implemented** | `aidetect.DeriveScenarios()` — eval dirs + AI imports | None | - |
| Populate AI graph nodes | - | **Missing** | NodePrompt/NodeDataset/NodeModel/NodeEvalMetric defined but never created | Graph nodes exist as types only; no build stage creates them | P1 |
| Eval-specific signals | - | **Missing** | No eval_safety_failure, eval_accuracy_regression detectors | Would need eval result parsing | P1 |
| Model version tracking | - | **Missing** | No model version metadata | Would need model registry integration | P2 |
| Dataset drift detection | - | **Missing** | No data distribution analysis | Would need dataset fingerprinting | P2 |

---

## Cross-Cutting Issues

### Documented Inconsistencies

| Issue | Where | Details | Recommended Action |
|---|---|---|---|
| `select-tests` vs `impact --show selected` | cli-spec.md, main.go | Both call identical `RenderProtectiveSet()`. `select-tests` is redundant. | Document as alias or deprecate `select-tests` in favor of `impact --show selected` | P2 |
| Health signals documented as "always available" | terrain-overview.md, feature-matrix.md | Health signals (slow, flaky, skipped, dead, unstable) **require `--runtime` artifacts**. Without runtime data, these signals are absent. Docs imply they work from static analysis. | Clarify in overview that health signals require runtime artifacts | P0 |
| Environment matrix "implemented" | feature-matrix.md | Matrix analysis code exists but returns empty without EnvironmentClass nodes. No code populates these from real CI configs. | Change status to "Partial" in feature-matrix.md | P1 |
| AI graph nodes "Partial" | feature-matrix.md, persona-matrix.md | feature-matrix.md says "Partial — types defined, not populated." This is accurate. But persona-matrix.md lists these under "Partial" without explaining why. | Align descriptions | P1 |

### Functional Gaps by Priority

#### P0 — Docs claim feature works but implementation has caveats

| Gap | Impact | Fix |
|---|---|---|
| Health signals require `--runtime` | Users expect skip/flaky detection from `analyze` alone. No static skip detection exists (e.g., `.skip()` patterns in code). | Either add static skip detection or prominently document the `--runtime` requirement in overview and feature-matrix |

#### P1 — Feature infrastructure exists but value is not delivered

| Gap | Impact | Fix |
|---|---|---|
| AI graph nodes not populated | NodePrompt/Dataset/Model/EvalMetric defined but never wired into Build(). No graph edges connect AI surfaces. | Add build stage to create AI nodes from CodeSurfaces with prompt/dataset kinds and Scenarios |
| Environment matrix empty without manual config | Mobile persona gets no value from matrix analysis on a real repo. | Add environment inference from CI config (GitHub Actions matrix), test file annotations, or framework metadata |
| Eval-specific signals not implemented | AI/ML persona can't see accuracy regression or safety failure signals. | Add eval signal detectors that trigger from Gauntlet artifacts or scenario execution results |
| Code unit extraction regex limitations | Backend persona may have false negatives for destructured exports, re-exports. | Document known limitation; consider tree-sitter for critical languages |

#### P2 — Planned features not yet started

| Gap | Count | Personas Affected |
|---|---|---|
| Visual regression / component coverage | 3 stories | Frontend |
| Database / API contract detection | 2 stories | Backend |
| External tool integration (TestRail, Jira) | 3 stories | QA/SDET |
| CI duration tracking / cost modeling | 3 stories | SRE |
| Model versioning / dataset drift | 2 stories | AI/ML |

---

## Command Audit Summary

| Command | Status | Notes |
|---|---|---|
| `terrain analyze` | Implemented | All flags work |
| `terrain insights` | Implemented | 10 finding builders |
| `terrain impact` | Implemented | BFS + confidence + fallback |
| `terrain explain` | Implemented | 6 entity types + scenarios |
| `terrain init` | Implemented | Artifact detection |
| `terrain summary` | Implemented | Executive view |
| `terrain posture` | Implemented | 5 dimensions, 18 measurements |
| `terrain focus` | Implemented | Risk-based prioritization |
| `terrain portfolio` | Implemented | Cost/breadth/leverage/redundancy |
| `terrain metrics` | Implemented | 6 metric categories |
| `terrain compare` | Implemented | Auto snapshot detection |
| `terrain select-tests` | Implemented | Redundant with `impact --show selected` |
| `terrain pr` | Implemented | 4 output formats (text, markdown, comment, annotation) |
| `terrain show` | Implemented | 5 entity types |
| `terrain policy check` | Implemented | 6 rule types, exit codes |
| `terrain export benchmark` | Implemented | Privacy-safe JSON |
| `terrain migration readiness` | Implemented | Area-by-area assessment |
| `terrain migration blockers` | Implemented | By type and risk |
| `terrain migration preview` | Implemented | Per-file/scope preview |
| `terrain ai list` | Implemented | Frameworks + auto-derived scenarios |
| `terrain ai doctor` | Implemented | 5-point diagnostic |
| `terrain ai run` | Implemented | Framework-delegated execution |
| `terrain ai record` | Implemented | Baseline snapshot |
| `terrain ai baseline` | Implemented | Baseline display |
| `terrain debug graph` | Implemented | Node/edge/density stats |
| `terrain debug coverage` | Implemented | Reverse coverage analysis |
| `terrain debug fanout` | Implemented | Transitive fanout |
| `terrain debug duplicates` | Implemented | Structural duplicate clusters |
| `terrain debug depgraph` | Implemented | All engines combined |
| `terrain version` | Implemented | Version/commit/date |

---

## Signal Detector Audit

All 22 documented signal types have working detectors:

| Signal | Detector | Requires Runtime | Status |
|---|---|---|---|
| slowTest | SlowTestDetector | Yes | Implemented |
| flakyTest | FlakyTestDetector | Yes | Implemented |
| skippedTest | SkippedTestDetector | Yes | Implemented |
| deadTest | DeadTestDetector | Yes | Implemented |
| unstableSuite | UnstableSuiteDetector | Yes | Implemented |
| weakAssertion | WeakAssertionDetector | No | Implemented |
| mockHeavyTest | MockHeavyDetector | No | Implemented |
| testsOnlyMocks | MockHeavyDetector | No | Implemented |
| snapshotHeavyTest | SnapshotHeavyDetector | No | Implemented |
| untestedExport | UntestedExportDetector | No | Implemented |
| coverageBlindSpot | CoverageBlindSpotDetector | No | Implemented |
| coverageThresholdBreak | CoverageThresholdDetector | No (needs coverage data) | Implemented |
| frameworkMigration | FrameworkMigrationDetector | No | Implemented |
| migrationBlocker | DeprecatedPatternDetector | No | Implemented |
| deprecatedTestPattern | DeprecatedPatternDetector | No | Implemented |
| dynamicTestGeneration | DynamicTestGenerationDetector | No | Implemented |
| customMatcherRisk | CustomMatcherDetector | No | Implemented |
| unsupportedSetup | UnsupportedSetupDetector | No | Implemented |
| policyViolation | GovernanceDetector | No (needs policy) | Implemented |
| legacyFrameworkUsage | GovernanceDetector | No | Implemented |
| skippedTestsInCI | GovernanceDetector | No | Implemented |
| runtimeBudgetExceeded | GovernanceDetector | No (needs runtime) | Implemented |

---

## Recommended Next Actions

1. **P0:** Add prominent "Requires `--runtime`" callout to health signal documentation in terrain-overview.md and feature-matrix.md
2. **P1:** Wire AI graph nodes into Build() — create NodePrompt/NodeDataset from CodeSurfaces with matching kinds
3. **P1:** Add static skip detection (`.skip()`, `@skip`, `xit()` patterns) as a code-level signal that doesn't require runtime data
4. **P1:** Add environment inference from GitHub Actions matrix config or package.json test scripts
5. **P2:** Deprecate `select-tests` as alias for `impact --show selected`, or document equivalence
6. **P2:** Align feature-matrix.md "environment matrix" status from "Implemented" to "Partial"
