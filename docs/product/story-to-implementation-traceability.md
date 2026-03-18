# Story-to-Implementation Traceability Matrix

> **Date:** 2026-03-16
> **Method:** Systematic trace from every documented user story in product docs to actual Go source, test coverage, and CLI behavior.
> **Confidence levels:** HIGH = code + tests + golden files confirm; MEDIUM = code exists but test coverage is thin or output unverified; LOW = code exists but functionality is degraded or conditional.

---

## Canonical Journey 1: Understand the Test System (`terrain analyze`)

| ID | User Goal | Key Packages | Tests | Confidence | UX Matches Docs? | Gap |
|---|---|---|---|---|---|---|
| J1-1 | Tests detected (count, frameworks) | `internal/analysis`, `internal/engine` | `TestSnapshot_Analyze`, golden file | HIGH | Yes | None |
| J1-2 | Repository profile (5 dimensions) | `internal/depgraph` (AnalyzeProfile) | Golden file validates profile fields | HIGH | Yes | None |
| J1-3 | Coverage confidence bands | `internal/depgraph` (AnalyzeCoverage) | Golden file validates band counts | HIGH | Yes | None |
| J1-4 | Duplicate cluster count | `internal/depgraph` (DetectDuplicates) | Golden file validates cluster count | HIGH | Yes | None |
| J1-5 | High-fanout count | `internal/depgraph` (AnalyzeFanout) | Golden file validates flagged count | HIGH | Yes | None |
| J1-6 | Skipped test burden | `internal/health` (SkippedTestDetector) | Conditional on `--runtime` | LOW | **No — appears only with runtime data** | Docs imply always present. Actually requires `--runtime`. No static `.skip()` detection. |
| J1-7 | Weak coverage areas | `internal/depgraph` (AnalyzeCoverage, BandLow) | Report lists uncovered paths | HIGH | Yes | None |
| J1-8 | CI optimization potential | `internal/analyze` (buildCIOptimization) | Present in report | HIGH | Yes | None |
| J1-9 | Top insight | `internal/analyze` (deriveTopInsight) | Present in report | HIGH | Yes | None |
| J1-10 | Risk posture (5 dimensions) | `internal/measurement` | 5 named dimensions confirmed | HIGH | Yes | None |
| J1-11 | Signal summary | `internal/analyze` (buildSignalSummary) | Total + by severity + by category | HIGH | Yes | None |

---

## Canonical Journey 2: Understand a Code Change (`terrain impact`)

| ID | User Goal | Key Packages | Tests | Confidence | UX Matches Docs? | Gap |
|---|---|---|---|---|---|---|
| J2-1 | Changed areas/packages | `internal/impact` (mapChangedSurfaces) | `TestSnapshot_Impact` | HIGH | Yes | None |
| J2-2 | Impacted tests count + total suite size | `internal/reporting/impact_report.go` | Report shows impacted count | MEDIUM | **No — total suite size not shown** | Docs promise "total tests for context" but output only shows impacted count |
| J2-3 | Coverage confidence for change scope | `internal/impact` (computeCoverageConfidence) | Golden file validates | HIGH | Yes | None |
| J2-4 | PR risk level | `internal/impact` (computeChangeRiskPosture) | Golden file validates posture band | HIGH | Yes | None |
| J2-5 | Top reason categories | `internal/impact` (computeReasonCategories) | 4 categories: direct, fixture, changed, proximity | HIGH | Yes | None |
| J2-6 | Insights affecting confidence | `internal/impact` (edge case policy) | Policy notes in output | MEDIUM | Partially — edge case notes appear but not as "insights" | Minor wording gap |
| J2-7 | Fallback level used | `internal/impact` (computeFallbackInfo) | Fallback section in report | HIGH | Yes | None |

---

## Canonical Journey 3: Improve the Test Suite (`terrain insights`)

| ID | User Goal | Key Packages | Tests | Confidence | UX Matches Docs? | Gap |
|---|---|---|---|---|---|---|
| J3-1 | Duplicate clusters with detail | `internal/depgraph`, `internal/insights` | `TestSnapshot_Insights`, golden | HIGH | Yes | None |
| J3-2 | High-fanout with worst offender | `internal/insights` (fanoutFindings) | Finding includes top node path | HIGH | Yes | None |
| J3-3 | Weak coverage areas | `internal/insights` (coverageFindings) | Finding includes uncovered count | HIGH | Yes | None |
| J3-4 | Skipped test debt | `internal/insights` (skipFindings) | Conditional on runtime signals | LOW | **Only with `--runtime`** | Same as J1-6. No static skip detection. |
| J3-5 | Repo profile + edge cases | `internal/insights`, `internal/depgraph` | Profile and edge cases in output | HIGH | Yes | None |
| J3-6 | Ranked improvement opportunities | `internal/insights` (buildRecommendations) | Recommendations sorted by priority | HIGH | Yes | None |
| J3-7 | Policy recommendations | `internal/depgraph` (ApplyEdgeCasePolicy) | Policy section in output | HIGH | Yes | None |

---

## Canonical Journey 4: Explain Terrain's Reasoning (`terrain explain`)

| ID | User Goal | Key Packages | Tests | Confidence | UX Matches Docs? | Gap |
|---|---|---|---|---|---|---|
| J4-1 | Target metadata | `cmd/terrain/main.go` (renderTestDetail) | `TestSnapshot_Explain` | HIGH | Yes | None |
| J4-2 | Dependency paths / reason chains | `internal/explain` (ExplainTest, ReasonChain) | Explain golden test | HIGH | Yes — strongest path + alternatives shown | None |
| J4-3 | Confidence scores + evidence types | `internal/explain` (TestExplanation) | Confidence + band in output | HIGH | Yes | None |
| J4-4 | Connected signals and findings | `cmd/terrain/main.go` (renderTestDetail) | Signals shown for test file path | MEDIUM | **Partially — file-level signals shown for test files, but not for scenarios or code units** | Scenario explain does not show connected signals |
| J4-5 | Navigation hints | `cmd/terrain/main.go` (renderScenarioExplanation) | "Next steps" in scenario explain | MEDIUM | **Partially — scenario explain has next steps, test file detail does not** | Inconsistent across entity types |
| J4-6 | Explain by test file path | `cmd/terrain/main.go` (runExplain) | Works via snapshot lookup | HIGH | Yes | None |
| J4-7 | Explain by test case ID | `cmd/terrain/main.go` (runExplain) | Works via TestCase lookup | HIGH | Yes | None |
| J4-8 | Explain by code unit | `cmd/terrain/main.go` (runExplain) | Works via CodeUnit lookup | HIGH | Yes | None |
| J4-9 | Explain by owner | `cmd/terrain/main.go` (runExplain, showOwner) | Works via ownership lookup | HIGH | Yes | None |
| J4-10 | Explain by finding | `cmd/terrain/main.go` (runExplain, showFinding) | Works via portfolio finding lookup | HIGH | Yes | None |

---

## Output Determinism

| ID | Claim | Evidence | Confidence | Gap |
|---|---|---|---|---|
| Det-1 | analyze fully deterministic | `TestSnapshot_Analyze` golden | HIGH | None |
| Det-2 | insights fully deterministic | `TestSnapshot_Insights` golden | HIGH | None |
| Det-3 | impact deterministic except createdAt | `TestSnapshot_Impact` golden | HIGH | None |
| Det-4 | explain fully deterministic | `TestSnapshot_Explain` golden | HIGH | None |
| Det-5 | Signals sorted by severity then type | `models.SortSnapshot()` | HIGH | None |
| Det-6 | Test files sorted by path | `models.SortSnapshot()` | HIGH | None |
| Det-7 | Code units sorted by path then name | `models.SortSnapshot()` | HIGH | None |
| Det-8 | Risk dimensions in canonical order | Measurement dimension order hardcoded | HIGH | None |
| Det-9 | Floating-point fixed precision | Go default `json.Marshal` | LOW | **Overstated — no explicit fixed-precision formatting. Go uses full precision.** |

---

## Frontend Engineer Persona

| ID | User Goal | Command | Tests | Confidence | UX Matches? | Gap |
|---|---|---|---|---|---|---|
| FE-W1 | Framework fragmentation | `analyze` | Framework detection tests, 17 frameworks | HIGH | Yes | None |
| FE-W2 | Coverage of changed components | `impact --show units` | Impact tests | HIGH | Yes | None |
| FE-W3 | Snapshot-heavy files | `insights` | SnapshotHeavy edge case at >40% + >20 count | HIGH | Yes | None |
| FE-W4 | Run only affected tests | `select-tests` | Protective test set tests | HIGH | Yes | None |
| FE-W5 | Why a test was selected | `explain <test>` | Explain tests + golden | HIGH | Yes | None |
| FE-Support-1 | 9 JS/TS frameworks detected | `analyze` | `framework_detection_test.go` | HIGH | Yes | None |

---

## Backend Engineer Persona

| ID | User Goal | Command | Tests | Confidence | UX Matches? | Gap |
|---|---|---|---|---|---|---|
| BE-W1 | High-fanout modules | `insights`, `debug fanout` | Fanout tests | HIGH | Yes | None |
| BE-W2 | Change impact with confidence | `impact` | BFS + decay tests | HIGH | Yes | None |
| BE-W3 | Untested service boundaries | `analyze` | UntestedExportDetector | HIGH | Yes | None |
| BE-W4 | Runtime health signals | `analyze --runtime` | Health detector tests | HIGH | Yes | None |
| BE-W5 | Policy compliance | `policy check` | Policy tests, exit codes | HIGH | Yes | None |
| BE-Support-1 | 4 languages supported | `analyze` | Language-specific tests | HIGH | Yes | None |
| BE-Support-4 | LCOV + Istanbul ingestion | `analyze --coverage` | Coverage ingest tests | HIGH | Yes | None |

---

## Mobile / Device Engineer Persona

| ID | User Goal | Command | Tests | Confidence | UX Matches? | Gap |
|---|---|---|---|---|---|---|
| Mobile-W1 | Audit skip debt | `analyze` | SkippedTestDetector | LOW | **No — requires `--runtime`** | No static skip detection from code patterns |
| Mobile-W2 | Platform-conditional gaps | `insights` | Edge case detection | LOW | **Partially — HIGH_SKIP_BURDEN edge case exists but skip signals require runtime** | Skip-related insights empty without runtime |
| Mobile-W3 | Select tests for platform | `impact` | Standard impact | MEDIUM | Yes (generic, not platform-aware) | No platform-specific filtering |
| Mobile-W4 | Per-platform coverage | `analyze --coverage` | Coverage ingest | HIGH | Yes | None |
| Mobile-Support-1 | Skip detection framework-agnostic | health detectors | Only runtime-based | LOW | **Overstated — only from runtime results, not code** | |

---

## QA / SDET Persona

| ID | User Goal | Command | Tests | Confidence | UX Matches? | Gap |
|---|---|---|---|---|---|---|
| QA-W1 | Full health assessment | `analyze` | Golden tests + signal detectors | HIGH | Yes | None |
| QA-W2 | Manual coverage overlay | `analyze` with yaml | terrain_config_test.go | HIGH | Yes | None |
| QA-W3 | Policy enforcement | `policy check` | Policy loader tests | HIGH | Yes | None |
| QA-W4 | Prioritized backlog | `insights` | Insights tests | HIGH | Yes | None |
| QA-W5 | Executive reporting | `summary`, `posture` | Reporting tests | HIGH | Yes | None |
| QA-W6 | Trend tracking | `compare` | Snapshot compare tests | HIGH | Yes | None |
| QA-W7 | Privacy-safe benchmarking | `export benchmark` | Export tests | HIGH | Yes | None |
| QA-Support-2 | 6 policy rule types | `policy check` | policy/config.go has exactly 6 fields | HIGH | Yes | None |
| QA-Support-3 | Evidence strength grading | `posture` | Measurement model | HIGH | Yes | None |

---

## SRE / Platform Engineer Persona

| ID | User Goal | Command | Tests | Confidence | UX Matches? | Gap |
|---|---|---|---|---|---|---|
| SRE-W1 | PR test selection | `pr`, `select-tests` | Changescope tests | HIGH | Yes | `select-tests` is functionally identical to `impact --show selected` |
| SRE-W2 | Slow test identification | `analyze --runtime` | SlowTestDetector tests | HIGH | Yes | None |
| SRE-W3 | Flaky test identification | `analyze --runtime` | FlakyTestDetector tests | HIGH | Yes | None |
| SRE-W4 | CI optimization | `analyze` | CI optimization section | HIGH | Yes | None |
| SRE-W5 | System metrics | `metrics --json` | Metrics tests | HIGH | Yes | None |
| SRE-W6 | Debug graph | `debug graph/fanout` | Engine tests | HIGH | Yes | None |
| SRE-Support-1 | GitHub Actions integration | CI workflows | terrain-pr.yml exists | HIGH | Yes | None |
| SRE-Support-2 | 5-level fallback ladder | impact | Fallback logic in analysis.go | HIGH | Yes | None |

---

## AI / ML Engineer Persona

| ID | User Goal | Command | Tests | Confidence | UX Matches? | Gap |
|---|---|---|---|---|---|---|
| AI-W1 | Audit eval suite | `analyze` | Validation inventory | HIGH | Yes | None |
| AI-W2 | Find eval gaps | `insights` | Scenario duplication finding | HIGH | Yes | None |
| AI-W3 | Impact of model changes | `impact` | `findImpactedScenarios()` | HIGH | Yes | None |
| AI-W4 | Detect duplicate evals | `insights` | scenarioDuplicationFindings | HIGH | Yes | None |
| AI-W5 | Eval policy compliance | `policy check` | Standard policy | HIGH | Yes | None |
| AI-W6 | List scenarios/prompts | `ai list` | AI workflow tests | HIGH | Yes | None |
| AI-W7 | Validate eval setup | `ai doctor` | AI workflow tests | HIGH | Yes | None |
| AI-W8 | Execute eval scenarios | `ai run` | Framework delegation | HIGH | Yes | None |
| AI-W9 | Record baseline | `ai record` | Baseline file written | HIGH | Yes | None |
| AI-W10 | Manage baselines | `ai baseline` | Baseline read | HIGH | Yes | None |
| AI-W11 | Auto-detect frameworks | `ai list`, pipeline | `aidetect.Detect()` | HIGH | Yes | None |
| AI-W12 | Auto-derive scenarios | pipeline | `aidetect.DeriveScenarios()` | HIGH | Yes | None |
| AI-GAP-1 | AI graph nodes populated | depgraph Build() | NodePrompt/Dataset/Model/EvalMetric never created | MISSING | **No** | Types exist but Build() never instantiates them |
| AI-GAP-2 | Eval-specific signals | - | No eval_safety_failure detector | MISSING | **No** | No eval-aware signal detectors |

---

## Capabilities Implemented but Not in Product Docs

| Capability | Package | What It Does | Missing From |
|---|---|---|---|
| PR comment deduplication | `internal/changescope/dedup.go` | Removes duplicate findings, classifies new vs existing risk | Not documented in any product doc |
| Merge recommendation | `internal/changescope/dedup.go` | Safe/Caution/Blocked/Informational | Not in feature-matrix or cli-spec |
| Test grouping by package | `internal/changescope/dedup.go` | Groups large test sets by directory | Not documented |
| Graph performance caching | `internal/depgraph/graph.go` | Type/family indexes, query caching, Seal() | Not in architecture docs |
| Parallel artifact ingestion | `internal/engine/pipeline.go` | Worker pool for runtime/coverage files | Not documented |
| Parallel duplicate scoring | `internal/depgraph/duplicate.go` | Worker pool for >100 candidate pairs | Not documented |
| CLI progress output | `cmd/terrain/progress.go` | [1/5] step-based TTY progress | Not in cli-spec |
| Truthcheck harness | `internal/truthcheck`, `cmd/terrain-truthcheck` | Ground truth validation with P/R scoring | Not in product docs (benchmark-only) |
| 14 ground-truth fixtures | `tests/fixtures/` | Comprehensive benchmark repos with truth specs | Not in product docs |

---

## Places Where Docs Overstate Maturity

| Claim | Source | Reality | Severity |
|---|---|---|---|
| Health signals always appear in analyze | terrain-overview.md, feature-matrix.md | 5 health signals (slow/flaky/skipped/dead/unstable) require `--runtime` artifacts. Without runtime data, health section is empty. | **HIGH** — users will run `analyze` without `--runtime` and see no health info |
| Environment matrix "Implemented" | feature-matrix.md | Returns empty result without EnvironmentClass nodes. No CI config inference exists. | **MEDIUM** — technically runs but produces no output on most repos |
| "Floating-point values use fixed precision" | canonical-user-journeys.md | Go's default `json.Marshal` used. No explicit formatting. | **LOW** — unlikely to cause problems in practice |
| "Impacted tests count and total tests for context" | canonical-user-journeys.md | Impact report shows impacted count only, not total suite size. | **LOW** — easy to add |
| Skip debt audit for Mobile persona | persona-matrix.md, 17-persona-journeys.md | Skip detection is runtime-only. Mobile persona hero moment requires `--runtime` which is not mentioned. | **MEDIUM** — mobile persona's primary value prop is gated on an undisclosed prerequisite |
| `terrain compare` auto-detects snapshots | cli-spec.md | Works but `.terrain/snapshots/` must exist (created by `--write-snapshot`). Not auto-created. | **LOW** — expected workflow creates it |

---

## Test Coverage Summary

| Area | Test File(s) | Test Count | Coverage of Documented Claims |
|---|---|---|---|
| Canonical journeys (golden) | `cmd/terrain/snapshot_test.go` | 4 | All 4 journeys validated |
| Impact analysis | `internal/impact/impact_test.go` | 48 | Strong — covers BFS, confidence, scenarios |
| Insights | `internal/insights/insights_test.go` | 13 | Moderate — covers findings, but not all edge cases |
| Explain | `internal/explain/explain_test.go` + golden | 14 | Good — covers test + scenario explain |
| PR/Changescope | `internal/changescope/*_test.go` | 26 | Strong — includes dedup, rendering, grouping |
| AI workflow | `cmd/terrain/ai_workflow_test.go` | 7 | Integration tests for scenario loading, impact, explain |
| AI detection | `internal/aidetect/detect_test.go` | 9 | Framework detection, scenario derivation |
| Gauntlet | `internal/gauntlet/ingest_test.go` | 10 | Parsing, signal generation, matching |
| Truthcheck | `internal/truthcheck/checker_test.go` | 3 | Spec loading, full run, scoring |
| Graph perf | `internal/depgraph/bench_test.go` | 12 | Benchmarks for all query/analysis ops |
| Determinism | `internal/depgraph/engines_test.go` | 4 | Duplicate/coverage/impact/fanout determinism |

---

## Recommended Priority Actions

| Priority | Action | Persona Impact | Effort |
|---|---|---|---|
| P0 | Add `--runtime` requirement callout to health signal docs | All personas | 30 min |
| P0 | Add static skip detection (`.skip()`, `xit()`, `@skip` patterns) as a code-level quality signal | Mobile, QA, SRE | 2-4 hours |
| P1 | Wire AI graph nodes into Build() from CodeSurfaces | AI/ML | 1-2 hours |
| P1 | Add total suite size to impact report output | All | 30 min |
| P1 | Add environment inference from CI config | Mobile, SRE | 4-8 hours |
| P2 | Document `select-tests` as alias for `impact --show selected` | SRE | 15 min |
| P2 | Add "Next steps" to test file explain output | All | 30 min |
| P2 | Document undocumented capabilities (dedup, merge rec, progress) | All | 1 hour |
