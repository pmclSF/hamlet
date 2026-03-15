# Persona Journeys

> **Status:** Implemented (core journeys); Planned (extended personas)
> **Purpose:** Map Terrain's canonical workflows to specific user personas and their decision contexts. Each persona asks different questions of the same underlying graph — the CLI surfaces the right answer at the right altitude.
> **Key decisions:**
> - Four core commands (analyze, insights, impact, explain) map to distinct decision moments, not distinct users
> - Personas differ in which output dimensions they prioritize, not which commands they use
> - Journey transitions are first-class — commands produce cross-references that guide the next question
> - Extended commands (summary, focus, posture, portfolio, metrics, compare, policy, migration) serve specialized decision contexts

**See also:** [09-cli-spec.md](09-cli-spec.md), [docs/product/canonical-user-journeys.md](../product/canonical-user-journeys.md), [docs/product/feature-matrix.md](../product/feature-matrix.md)

## Core Journeys

Terrain's four canonical commands each answer a different question at a different moment in the development cycle:

| Command | Question | Decision Moment |
|---------|----------|-----------------|
| `terrain analyze` | What is the state of our test system? | Sprint planning, onboarding, audits |
| `terrain insights` | What should we fix or improve? | Backlog grooming, tech debt triage |
| `terrain impact` | What is affected by this change? | PR review, release readiness |
| `terrain explain` | Why does Terrain say this? | Debugging findings, building trust |

These commands are not exclusive to any persona. A frontend engineer and an engineering manager both use `terrain analyze` — but they read different parts of the output and take different actions.

## Personas

---

### Frontend Engineer

**Context:** Owns UI components and their associated integration tests. Cares about test reliability for visual and interaction behavior. Often works across multiple frameworks (Jest for units, Playwright or Cypress for E2E).

**Key concerns:**
- Component interaction paths are actually tested, not just rendered
- E2E tests don't flake and block PRs
- Snapshot tests aren't masking real coverage gaps
- Multiple framework configs (Jest, Vitest, Playwright, Cypress) are coherent

**Fears and risks:**
- "I refactored a component and nothing failed — were there even tests?"
- "Our snapshot tests pass but don't test real behavior"
- "We have 3 different test frameworks and nobody knows which to use for what"
- "E2E tests flake on CI but pass locally — I can't ship"

**Hero moment:** Runs `terrain impact` on a component refactor PR. Terrain shows that 4 E2E tests and 2 unit tests cover the changed component, identifies 1 interaction path with no coverage, and selects only the relevant tests — CI runs in 2 minutes instead of 15.

**Primary Terrain workflows:**

| Workflow | Command | What they see |
|----------|---------|---------------|
| Understand framework fragmentation | `terrain analyze` | Framework distribution, test file counts per framework, snapshot ratio |
| Check coverage of changed components | `terrain impact --show units` | Which code surfaces are covered vs. exposed by the change |
| Find snapshot-heavy test files | `terrain insights` | `snapshotHeavyTest` signals, duplicate cluster candidates |
| Run only affected tests on PR | `terrain select-tests` | Minimal test set for CI |
| Understand why a test was selected | `terrain explain <test>` | Dependency path from changed file to test |

**Current support level:** Strong. Terrain detects Jest, Vitest, Mocha, Jasmine, Playwright, Cypress, WebdriverIO, Puppeteer, TestCafe. Snapshot-heavy detection triggers at >40% snapshot assertions. Framework migration signals flag multi-framework repos. Import graph links component source to test files.

**Gaps:** No visual regression framework detection (Storybook visual tests, Percy, Chromatic). No component-level coverage mapping — coverage is file-level. No React/Vue/Svelte component tree awareness.

**Fixture repo:** `tests/fixtures/frontend-react/` — snapshot-heavy component tests with Jest + Playwright + Enzyme.

---

### Backend Engineer

**Context:** Owns services, APIs, and data layers. Tests are typically unit tests with some integration tests against databases or external services. Concerned with transitive dependency risk — changing a shared utility can break downstream consumers.

**Key concerns:**
- Service boundary coverage — are all API contracts tested?
- Transitive dependency risk — what breaks when a shared module changes?
- Integration test reliability — database and service dependencies
- High-fanout modules creating blast radius problems

**Fears and risks:**
- "I changed a utility function and 40 tests failed that I didn't know existed"
- "Our database module is imported everywhere — one change ripples through the whole suite"
- "We have no tests for the payment service boundary"
- "Tests pass in isolation but fail when run together (shared state)"

**Hero moment:** Runs `terrain impact` before merging a data model change. Terrain traces the import graph, identifies 12 tests affected through 3 transitive paths, flags 2 protection gaps in the payment service boundary, and recommends running the 12 tests plus 1 new integration test for the uncovered gap.

**Primary Terrain workflows:**

| Workflow | Command | What they see |
|----------|---------|---------------|
| Find high-fanout modules | `terrain insights` / `terrain debug fanout` | Nodes with excessive transitive dependents, refactoring candidates |
| Trace change impact through dependencies | `terrain impact` | Impact graph with confidence decay by path length |
| Find untested service boundaries | `terrain analyze` | `untestedExport` signals for uncovered public functions |
| Check coverage with runtime data | `terrain analyze --runtime junit.xml` | Slow test detection, flaky test flagging, dead test identification |
| Evaluate policy compliance | `terrain policy check` | Coverage threshold violations, framework restrictions |

**Current support level:** Strong. Go, Python, Java, and JS/TS all supported for framework detection and code unit extraction. Import graph construction works for all supported languages. Fanout analysis identifies high-risk shared modules. Coverage artifact ingestion (LCOV, Istanbul) provides concrete coverage data. JUnit XML ingestion enables slow/flaky/dead test detection.

**Gaps:** No database schema change tracking. No service contract / API schema testing detection (OpenAPI, gRPC). No distributed system test topology awareness. Code unit extraction is regex-based, not AST — may miss complex export patterns.

**Fixture repo:** `tests/fixtures/backend-api/` — Go backend with handlers, middleware, LCOV coverage, and JUnit runtime.

---

### Mobile / Device-Sensitive Engineer

**Context:** Targets multiple platforms and devices. Test matrices are large (OS versions, screen sizes, device capabilities). CI time is expensive because device farms are slow and costly.

**Key concerns:**
- Platform-conditional tests — which tests require specific hardware?
- Skip debt from device-dependent tests that can't run in CI
- Test matrix coverage — are all platform combinations exercised?
- CI cost — device farm minutes are expensive

**Fears and risks:**
- "Half our tests are skipped because they need a physical device"
- "We have no idea which platform combinations are actually tested"
- "A bug shipped on Android that was caught by an iOS-only test we skipped"
- "CI takes 45 minutes because we run everything on every platform"

**Hero moment:** Runs `terrain analyze` and sees that 59% of auth tests are skipped due to platform dependencies. `terrain insights` identifies 4 test files that could be refactored to use platform stubs, reducing the skip rate to 15% and enabling CI coverage for SAML and LDAP flows.

**Primary Terrain workflows:**

| Workflow | Command | What they see |
|----------|---------|---------------|
| Audit skip debt | `terrain analyze` | Skipped test count, skip burden classification |
| Find platform-conditional gaps | `terrain insights` | Skip patterns, edge case detection (high skip burden) |
| Select tests for a platform build | `terrain impact` | Tests relevant to changed platform module |
| Check coverage for specific platform | `terrain analyze --coverage coverage-ios.lcov` | Coverage data scoped to one platform's test run |

**Current support level:** Partial. Terrain detects skipped tests across all frameworks and flags high skip burden (>20%) as an edge case. The skip signal captures skip reasons where present in test source. Coverage and runtime artifacts can be provided per-platform via explicit `--coverage` and `--runtime` flags.

**Gaps:** No device/environment matrix awareness — Terrain cannot model "this test runs on iOS but not Android." No CI matrix analysis (GitHub Actions matrix, device farm configs). No platform-conditional skip classification beyond source pattern matching. No test duration modeling by device type.

**Fixture repo:** `tests/fixtures/mobile-cross-platform/` — platform-conditional tests with 53% skip rate.

---

### QA / SDET

**Context:** Owns test infrastructure and quality strategy. Responsible for coverage policy, test health metrics, and manual-to-automated migration. Often bridges automated tests with manual QA suites tracked in TestRail, Jira, or similar tools.

**Key concerns:**
- Overall test health and trend direction
- Manual coverage gaps — where does QA effort supplement automation?
- Policy compliance — are teams meeting coverage standards?
- Automation ROI — where should new automation investment go?

**Fears and risks:**
- "We have 200 manual tests but no visibility into what they cover"
- "Teams are skipping tests in CI and nobody notices"
- "Coverage numbers look good but the tests don't assert anything meaningful"
- "We can't answer 'are we ready to ship?' with data"

**Hero moment:** Configures `.terrain/terrain.yaml` with manual coverage entries from TestRail. Runs `terrain analyze` and sees both automated and manual coverage in one view. `terrain insights` identifies 3 manual test suites that overlap with existing automation and 2 areas where manual tests are the only coverage — creating a targeted automation backlog.

**Primary Terrain workflows:**

| Workflow | Command | What they see |
|----------|---------|---------------|
| Full system health assessment | `terrain analyze` | Complete test system profile with all data sources |
| Manual coverage overlay | `terrain analyze` (with `.terrain/terrain.yaml`) | Manual test suites mapped to source files |
| Policy enforcement | `terrain policy check` | Violations: coverage thresholds, framework restrictions, skip limits |
| Prioritized improvement backlog | `terrain insights` | Ranked findings with actions: add coverage, remove duplicates, fix flaky |
| Executive reporting | `terrain summary` / `terrain posture` | Risk posture, evidence strength, limitations |
| Trend tracking | `terrain compare` | Before/after comparison across snapshots |
| Privacy-safe benchmarking | `terrain export benchmark` | Metrics without file paths or identifiers |

**Current support level:** Strong. Manual coverage overlay via `.terrain/terrain.yaml` maps external QA tools to source files with criticality levels. Policy engine supports 6 rule types. Posture reports explicitly state evidence strength (strong/partial/weak/none) and limitations. Snapshot comparison enables trend tracking. Metrics export is privacy-safe by design.

**Gaps:** No direct TestRail/Jira/Xray integration — manual coverage is declared in YAML, not synced. No automated trend alerting — compare requires two explicit snapshots. No multi-repo aggregation dashboard. No test case management system sync.

**Fixture repo:** `tests/fixtures/qa-manual-overlay/` — manual coverage YAML with TestRail/Jira references, CODEOWNERS, policy rules.

---

### SRE / Platform Engineer

**Context:** Owns CI/CD pipelines, test infrastructure, and environment provisioning. Cares about test execution efficiency and environment reliability. Evaluates Terrain as a CI optimization tool.

**Key concerns:**
- CI pipeline duration and cost
- Flaky tests causing false failures and re-runs
- Test selection for faster PR feedback
- Infrastructure reliability — test environments, fixtures, shared state

**Fears and risks:**
- "CI takes 20 minutes and developers stop waiting for it"
- "Flaky tests cause 30% of CI re-runs"
- "We run 5000 tests on every PR but only 50 are relevant"
- "Nobody knows which tests are slow because we never measured"

**Hero moment:** Integrates `terrain pr` into the CI workflow via `terrain-pr.yml`. PR builds drop from 18 minutes to 4 minutes by running only Terrain-selected tests. The posture check ensures coverage confidence stays high — no tests are silently dropped.

**Primary Terrain workflows:**

| Workflow | Command | What they see |
|----------|---------|---------------|
| PR test selection | `terrain pr` / `terrain select-tests` | Minimal protective test set with confidence |
| Identify slow tests | `terrain analyze --runtime junit.xml` | Slow test signals with actual timing data |
| Identify flaky tests | `terrain analyze --runtime junit.xml` | Flaky test signals with retry/failure patterns |
| CI optimization assessment | `terrain analyze` | CI optimization potential, suite runtime profile |
| System metrics for monitoring | `terrain metrics --json` | Aggregate metrics suitable for dashboards |
| Debug dependency graph issues | `terrain debug graph` / `terrain debug fanout` | Raw graph statistics, fanout analysis |

**Current support level:** Strong. GitHub Actions integration via `terrain-pr.yml` workflow and `terrain-impact` composite action. Test selection with 5-level fallback ladder ensures safety. Runtime artifact ingestion enables data-driven slow/flaky detection. Metrics export is designed for dashboard integration.

**Gaps:** No direct CI system integration beyond CLI invocation — Terrain is invoked as a step, not a plugin. No test parallelization recommendations. No historical CI duration tracking — runtime data is per-snapshot, not time-series. No cost modeling (compute minutes, device farm billing).

**Fixture repos:** `tests/fixtures/high-fanout/` (blast radius), `tests/fixtures/skipped-tests/` (skip burden).

---

### AI / ML Engineer

**Context:** Maintains model evaluation suites alongside traditional tests. Eval suites measure behavioral properties (accuracy, safety, bias, format compliance) rather than functional correctness. Test patterns include parametrized test matrices, golden output comparisons, and threshold-based assertions.

**Key concerns:**
- Eval suite coverage — are all behavioral scenarios exercised?
- Regression detection — did model behavior change?
- Threshold drift — are accuracy/latency bounds still met?
- Eval-to-code linkage — which evals cover which model components?

**Fears and risks:**
- "We shipped a model update and missed a regression in edge-case handling"
- "Our eval suite tests the happy path but not adversarial inputs"
- "Nobody knows which eval tests cover which model behavior"
- "We added 50 eval cases but don't know if they're redundant"

**Hero moment:** Runs `terrain analyze` on the eval suite. Terrain detects pytest with parametrized tests, identifies 2 skipped eval scenarios (sarcasm detection, multilingual), flags 3 eval files as structurally similar (potential consolidation), and shows that model output format compliance has strong coverage but accuracy thresholds are only tested on the positive class.

**Primary Terrain workflows:**

| Workflow | Command | What they see |
|----------|---------|---------------|
| Audit eval suite coverage | `terrain analyze` | Test file count, framework detection, parametrized test count |
| Find eval gaps | `terrain insights` | Untested exports, skipped evals, weak assertion patterns |
| Check impact of model changes | `terrain impact` | Which eval files are affected by changes to model code |
| Detect duplicate evals | `terrain insights` | Structurally similar eval files (shared fixtures, similar assertions) |
| Verify eval policy compliance | `terrain policy check` | Coverage thresholds, skip restrictions |

**Current support level:** Partial. Terrain detects pytest (including `@pytest.mark.parametrize`, `@pytest.mark.skip`) and can analyze Python eval suites like any test suite. Import graph links eval files to model source. Duplicate detection can identify structurally similar evals. Coverage and runtime artifact ingestion works for Python test outputs.

**Gaps:** No eval-specific semantics — Terrain treats eval suites as regular test suites. No accuracy/metric threshold tracking. No model version awareness. No dataset drift detection. No distinction between behavioral evals and functional tests. No LLM-specific patterns (prompt testing, safety evals, hallucination checks). These are potential future extensions.

**Fixture repo:** `tests/fixtures/ai-eval-suite/` — Python pytest eval suite with accuracy thresholds, parametrized tests, and skip markers.

---

## Journey Transitions

Commands produce structured output that naturally leads to the next question. Terrain makes these transitions explicit:

- **analyze** surfaces risk dimensions. High change risk leads to **insights** for specific recommendations.
- **insights** identifies issues. Each issue links to **explain** for the evidence chain.
- **impact** shows affected tests. Unexpected impact paths lead to **explain** for the dependency chain.
- **explain** traces evidence. Understanding the evidence leads back to **analyze** to verify the fix.

```
analyze → insights → explain → impact
   │          │          │         │
   │          │          │         └─ PR workflow (CI integration)
   │          │          └─ Deep dive into any finding
   │          └─ Planning: what to fix next
   └─ First run: understand the system
```

In JSON output mode, each finding includes a `relatedCommands` field listing the natural next queries. In terminal output, Terrain prints suggested follow-up commands after each result.

## Extended Commands and Persona Fit

| Command | Primary Persona | Decision Context |
|---------|----------------|------------------|
| `terrain summary` | Engineering Manager | Executive briefing, stakeholder communication |
| `terrain focus` | SRE, QA/SDET | CI optimization, critical path identification |
| `terrain posture` | Engineering Manager, QA/SDET | Quality trend tracking, release gates |
| `terrain portfolio` | Engineering Manager | Cross-repository comparison, org-level investment |
| `terrain metrics` | SRE, QA/SDET | Execution performance, trend analysis |
| `terrain compare` | All personas | Before/after measurement, migration progress |
| `terrain policy` | QA/SDET, Engineering Manager | Governance enforcement, compliance reporting |
| `terrain migration` | Frontend Engineer, QA/SDET | Framework migration readiness, blockers |

These commands are not persona-exclusive. The table reflects the most common decision context — any engineer can run any command. Terrain's value comes from making the right information accessible at the right altitude for whoever is asking.

## Debug Commands

Internal engine views accessible via `terrain debug <engine>`:

| Debug Command | Underlying Engine |
|---------------|-------------------|
| `terrain debug graph` | Dependency graph statistics (nodes, edges, density) |
| `terrain debug coverage` | Structural reverse coverage analysis |
| `terrain debug fanout` | High-fanout node detection |
| `terrain debug duplicates` | Duplicate test cluster analysis |
| `terrain debug depgraph` | Full dependency graph analysis (all engines) |

These are intended for development and debugging, not end-user workflows.
