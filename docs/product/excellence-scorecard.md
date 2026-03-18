# Engineering Excellence Scorecard

> **Date:** 2026-03-16
> **Method:** End-to-end re-audit of all CLI commands, documented personas, and user stories after the v3 improvement cycle.
> **Test suite:** 38 Go packages, 0 failures. Truthcheck: 7/7 categories, 100% F1.

---

## Scoring Key

| Rating | Meaning |
|--------|---------|
| **A** | Production-ready. Documented, tested, deterministic. |
| **B** | Works correctly. Minor polish or coverage gaps remain. |
| **C** | Functional but degraded. Documented caveats or missing sub-features. |
| **D** | Scaffolded or aspirational. Not usable without workaround. |
| **--** | Not applicable to this persona. |

---

## Canonical Workflow Scores

| Workflow | Command | Clarity | Correctness | Trustworthiness | Test Coverage | Release Ready | Overall |
|----------|---------|---------|-------------|-----------------|---------------|---------------|---------|
| Understand test system | `analyze` | A | A | A | A (golden + snapshot) | A | **A** |
| Find improvement opportunities | `insights` | A | A | A | A (golden + dedup) | A | **A** |
| Understand a code change | `impact` | A | A | A | A (golden + decay) | A | **A** |
| Explain a decision | `explain` | A | A | B | B (6 types, scenario fallback new) | A | **A-** |
| PR reviewer summary | `pr` | A | A | A | A (dedup + 3-way + suite size) | A | **A** |
| AI validation | `ai list/doctor` | A | A | B | B (integration tests) | A | **A-** |
| Executive view | `summary` | A | A | A | B (golden) | A | **A** |
| Policy enforcement | `policy check` | A | A | A | A | A | **A** |
| Trend tracking | `compare` | B | A | A | B | A | **B+** |
| Benchmark export | `export benchmark` | B | A | A | A (privacy test) | A | **A-** |

---

## Per-Persona Scores

### Frontend Engineer

| Dimension | Score | Evidence |
|-----------|-------|----------|
| Clarity | A | Framework fragmentation visible. Coverage bands clear. |
| Correctness | A | 17 frameworks detected. Import graph accurate for JS/TS. |
| Trustworthiness | A | Deterministic golden tests. Signal sort stable. |
| Test coverage | A | Golden tests, snapshot tests, framework detection tests. |
| Release readiness | A | All commands produce clean output. |
| **Overall** | **A** | |

**Strengths:** Framework detection covers 9 JS/TS frameworks. Snapshot-heavy detection. PR comment shows suite size reduction. Fanout detection actionable.

**Remaining gaps:**
- P2: No visual regression detection (Storybook, Percy)
- P2: No component-level coverage (file-level only)
- P2: No React/Vue component tree awareness

---

### Backend Engineer

| Dimension | Score | Evidence |
|-----------|-------|----------|
| Clarity | A | Impact shows "N of M total" with confidence bands. |
| Correctness | A | BFS with 0.85 decay per hop. Fanout penalty via log2. Protection gap dedup. |
| Trustworthiness | A | Confidence decay validated by dedicated tests. |
| Test coverage | A | Impact golden tests, decay regression tests, dedup tests. |
| Release readiness | A | All commands clean. |
| **Overall** | **A** | |

**Strengths:** Import graph covers JS/TS, Go, Python, Java. Untested export detection. Handler/route surface detection. Runtime artifact ingestion (JUnit XML, Jest JSON).

**Remaining gaps:**
- P1: Code unit extraction is regex-based (may miss complex export patterns)
- P2: No database schema change tracking
- P2: No API contract testing detection

---

### Mobile / Device Engineer

| Dimension | Score | Evidence |
|-----------|-------|----------|
| Clarity | B | Static skip detection now works. Matrix analysis conditional. |
| Correctness | B | Skip detection accurate for 4 languages. Environment matrix requires config. |
| Trustworthiness | B | Static skips detected; runtime skips require --runtime. |
| Test coverage | B | Skip detector tested (4 languages). Matrix not tested with real CI config. |
| Release readiness | B | Functional but matrix value limited without manual config. |
| **Overall** | **B** | |

**Strengths:** Static skip detection (`.skip()`, `xit()`, `@skip`, `@Disabled`) works without `--runtime`. Skip severity thresholds (>20% medium, >50% high). Truthcheck stability passes.

**Remaining gaps:**
- P1: Environment matrix returns empty without `.terrain/terrain.yaml` or CI matrix config
- P2: No platform-aware test filtering

---

### QA / SDET

| Dimension | Score | Evidence |
|-----------|-------|----------|
| Clarity | A | Health grade, headline, ranked recommendations. |
| Correctness | A | 10 finding builders. Dedup across builders. Insights ranking by severity then category. |
| Trustworthiness | A | Deterministic output tested. |
| Test coverage | A | Insights golden, dedup, ranking, determinism tests. |
| Release readiness | A | All commands clean. Policy check with exit codes. |
| **Overall** | **A** | |

**Strengths:** 5 posture dimensions. Manual coverage overlay via YAML. Policy enforcement (6 rule types). Compare with auto-snapshot detection. Privacy-safe benchmark export (tested).

**Remaining gaps:**
- P2: No TestRail/Jira/Xray integration
- P2: No automated trend alerting
- P2: No multi-repo aggregation

---

### SRE / Platform Engineer

| Dimension | Score | Evidence |
|-----------|-------|----------|
| Clarity | A | PR comment optimized for merge decisions. Suite size context. |
| Correctness | A | 3-way finding classification (direct/indirect/existing). Fallback ladder (5 levels). |
| Trustworthiness | A | PR dedup. Confidence bands. Merge recommendation logic tested. |
| Test coverage | A | PR rendering tests, dedup tests, truncation tests, grouping tests. |
| Release readiness | A | GitHub Action uses CLI renderer. CI workflow tested. |
| **Overall** | **A** | |

**Strengths:** `terrain pr --format markdown` optimized for GitHub. Test savings percentage. Collapsible pre-existing debt. GitHub Action aligned with CLI output. Benchmark smoke tests in CI.

**Remaining gaps:**
- P2: No CI plugin model (CLI-first by design)
- P2: No test parallelization recommendations
- P2: No historical CI duration tracking

---

### AI / ML Engineer

| Dimension | Score | Evidence |
|-----------|-------|----------|
| Clarity | A | `ai list` shows surface counts. `ai doctor` 6-point diagnostic. |
| Correctness | A | 12 frameworks auto-detected. Scenarios auto-derived without YAML. Dedup works. |
| Trustworthiness | B | Auto-detection depends on naming conventions. Graph nodes (Prompt, Dataset, Model, EvalMetric) not populated. |
| Test coverage | B | AI workflow integration tests. Detection unit tests. No per-framework detection test. |
| Release readiness | A- | `ai run/record/baseline` work for detected frameworks. Scaffolded fallback for undetected. |
| **Overall** | **A-** | |

**Strengths:** Zero-config scenario discovery. Prompt/dataset surface inference (JS/TS, Python). Scenario impact via `terrain impact`. Scenario explain works with and without `--base`. Gauntlet artifact ingestion. Scenario duplication detection (>50% overlap).

**Remaining gaps:**
- P1: Graph nodes (Prompt, Dataset, Model, EvalMetric) defined but never populated by Build()
- P1: No eval-specific signals (eval_safety_failure, eval_accuracy_regression)
- P2: No model version tracking
- P2: No dataset drift detection

---

## Cross-Cutting Quality

| Dimension | Score | Evidence |
|-----------|-------|----------|
| **Determinism** | A | Golden tests for analyze, insights, explain, impact. Deterministic sort everywhere. |
| **Duplicate suppression** | A | Insights dedup, impact gap dedup, changescope dedup, scenario dedup. |
| **Output consistency** | A | All text reports follow same structure (header, sections, next steps). |
| **JSON/human parity** | A | `--json` available on all commands. Text and JSON produce equivalent data. |
| **Error handling** | A | Descriptive errors with remediation hints. No panics. |
| **Release packaging** | A | npm: 99 files, 246KB. Go: goreleaser per-platform. verify-pack.js with 18 prohibited patterns. |
| **Documentation accuracy** | A- | Feature matrix, persona matrix, gap audit all updated. One overstatement corrected (AI auto-detection was "Future" but implemented). |
| **Truthcheck infrastructure** | A | 7/7 categories, 100% F1 on terrain-world. Per-category subtests. |

---

## Summary: Improvements Made This Cycle

| Commit | Impact | Personas |
|--------|--------|----------|
| Deduplicate AI scenarios, fix fanout labels | Correctness | All |
| 3-way PR finding classification + suite size | Clarity, trust | SRE, Backend |
| Reasoning engine: insights dedup + impact gap dedup + decay tests | Correctness, trust | All |
| AI: correct doc overstatements, 6-point doctor, surface counts | Clarity, trust | AI/ML |
| Static skip detection (4 languages) | Closes P0 | Backend, Mobile, QA |
| Impact suite size in text report | Clarity | SRE, Backend |
| Explain test-file Next steps | Consistency | All |
| Scenario explain without --base | Usability | AI/ML |
| Per-category truthcheck tests + privacy safety | Trust | All (engineering) |
| Release packaging audit | Release quality | All (engineering) |

**Total new tests added this cycle:** ~40 across 8 packages.

---

## Recommended Next Milestones

### Next Release (v3.1)
- [ ] Wire AI graph nodes (NodePrompt, NodeDataset) into Build() from CodeSurface kinds
- [ ] Add eval-specific signals from Gauntlet artifacts (eval_safety_failure, eval_accuracy_regression)
- [ ] Add environment inference from GitHub Actions matrix config
- [ ] Document code unit extraction limitations in feature-matrix

### Next Quarter (v3.2+)
- [ ] Visual regression detection (Storybook, Percy)
- [ ] API contract testing detection (OpenAPI, gRPC)
- [ ] External tool integration (TestRail, Jira)
- [ ] Model version tracking / dataset drift detection
- [ ] Consider tree-sitter for code unit extraction accuracy

---

## Risk Register

| Risk | Severity | Mitigation |
|------|----------|------------|
| Code unit extraction regex may miss complex export patterns | Medium | Documented as known limitation. False negatives possible for destructured re-exports. |
| Environment matrix returns empty without manual config | Medium | Mobile persona gets limited value. Needs CI config inference. |
| AI graph nodes never populated | Low | Types reserved. No user-facing feature depends on them yet. |
| Go binary release verification is manual | Low | `Makefile release-gate` runs tests. No goreleaser output inspection script. |
