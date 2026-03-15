# Terrain — Product Overview

> **Terrain is a test system intelligence platform. It maps your test terrain.**

Terrain reads your repository — test code, source structure, coverage data, runtime artifacts, ownership files, and local policy — and builds a structural model of how your tests relate to your code. From that model it surfaces risk, quality gaps, redundancy, fragile dependencies, and migration readiness, all without running a single test.

## Design Principles

**Inference-first.** Terrain infers structure from what already exists in your repo — import graphs, file naming, coverage artifacts, runtime results. No configuration required to get started. `terrain init` detects available data sources; `terrain analyze` produces a complete assessment.

**Explainability-first.** Every finding carries an evidence chain. `terrain explain` traces any decision back to the signals, dependency paths, and scoring rules that produced it. No black boxes, no magic numbers.

**Signal-first.** The core abstraction is the signal — a typed, located, severity-scored, confidence-weighted finding with an explanation and suggested action. All 22 signal types flow through the same registry, making the system extensible and auditable.

**Evidence-graded.** Terrain explicitly states the strength of its evidence (strong, partial, weak, none) and adjusts confidence accordingly. It tells you what it knows, what it inferred, and what it cannot see.

## The Four Canonical Workflows

Everything in Terrain is organized around four questions:

| Command | Question | Decision Moment |
|---------|----------|-----------------|
| `terrain analyze` | What is the state of our test system? | Sprint planning, onboarding, audits |
| `terrain insights` | What should we fix or improve? | Backlog grooming, tech debt triage |
| `terrain impact` | What validations matter for this change? | PR review, release readiness |
| `terrain explain` | Why did Terrain say this? | Debugging findings, building trust |

These four commands are the product front door. All other commands — `summary`, `posture`, `focus`, `portfolio`, `metrics`, `compare`, `migration`, `policy`, `select-tests`, `pr`, `show`, `export` — are supporting views that answer follow-up questions from the same underlying model.

## What Terrain Analyzes

**Languages:** JavaScript/TypeScript, Go, Python, Java.

**Frameworks:** 19 test frameworks across 4 languages — Jest, Vitest, Mocha, Jasmine, Playwright, Cypress, Puppeteer, WebdriverIO, TestCafe, Node Test, pytest, unittest, nose2, go-testing, JUnit 4, JUnit 5, TestNG.

**Data sources:**
- Source code and test code (static analysis — always available)
- Coverage artifacts (LCOV, Istanbul JSON — optional, enriches analysis)
- Runtime artifacts (JUnit XML, Jest/Vitest JSON — optional, enables health signals)
- CODEOWNERS files (ownership attribution)
- `.terrain/terrain.yaml` (manual coverage overlay, external QA tool mapping)
- `.terrain/policy.yaml` (governance rules)
- Git history (change impact analysis)

## Signal Categories

| Category | Signals | What They Detect |
|----------|---------|------------------|
| **Health** | slowTest, flakyTest, skippedTest, deadTest, unstableSuite | Runtime health and operational risk |
| **Quality** | untestedExport, weakAssertion, mockHeavyTest, testsOnlyMocks, snapshotHeavyTest, coverageBlindSpot | Structural quality and coverage depth |
| **Migration** | frameworkMigration, migrationBlocker, deprecatedTestPattern, dynamicTestGeneration, customMatcherRisk | Migration readiness and blockers |
| **Governance** | policyViolation, legacyFrameworkUsage, skippedTestsInCI, runtimeBudgetExceeded | Policy compliance and organizational standards |

Each signal includes: type, category, severity (Low/Medium/High), confidence (0.0–1.0), location (file + symbol), explanation, suggested action, and metadata.

## Measurement Model

Terrain computes posture across five dimensions:

1. **Health** — Skip burden, flaky rate, dead test share, runtime stability
2. **Coverage Depth** — Untested export share, weak assertion share, threshold compliance
3. **Coverage Diversity** — Mock concentration, framework fragmentation, E2E-only coverage, unit test reach
4. **Structural Risk** — Fanout burden (high-dependency modules), duplicate test clusters
5. **Operational Risk** — Policy violations, governance compliance

These dimensions roll up into a posture band (strong, moderate, weak, critical) with explicit evidence grading for each dimension.

## Who Uses Terrain

Terrain serves six primary personas, each asking different questions of the same underlying graph:

| Persona | Primary Concern | Key Workflow |
|---------|----------------|--------------|
| **Frontend Engineer** | Component coverage, snapshot debt, framework fragmentation | `impact` on component refactors; `insights` for snapshot-heavy files |
| **Backend Engineer** | Transitive dependency risk, service boundary coverage | `impact` for blast radius; `debug fanout` for high-risk modules |
| **Mobile / Device Engineer** | Skip debt from platform dependencies, CI cost | `analyze` for skip burden; `insights` for platform-conditional gaps |
| **QA / SDET** | Test health trends, policy compliance, manual coverage overlay | `policy check`; `posture`; `compare` for trend tracking |
| **SRE / Platform Engineer** | CI duration, flaky tests, test selection | `pr` in CI; `select-tests` for fast feedback; `metrics` for dashboards |
| **AI / ML Engineer** | Eval suite coverage, parametrized test auditing | `analyze` for eval audit; `insights` for redundant evals |

See [Persona Journeys](../architecture/17-persona-journeys.md) and [Feature Matrix](feature-matrix.md) for detailed capability mapping.

## CI Integration

Terrain integrates into CI as a CLI step. The primary integration point is the GitHub Actions composite action at `.github/actions/terrain-impact/`, which:

1. Builds the Go binary
2. Runs impact analysis against the PR diff
3. Parses results (test count, risk level, coverage confidence)
4. Posts/updates a PR comment with the impact summary
5. Uploads the impact artifact for downstream consumption
6. Writes a job summary

Test selection uses a 5-level fallback ladder to ensure safety: exact match → inferred match → transitive match → heuristic match → full suite.

## Architecture

```
cmd/terrain/         CLI entry point and command routing
internal/
├── engine/          Pipeline orchestration, detector registry
├── analysis/        Repository scanning, content analysis, language detection
├── signals/         Signal type registry and construction
├── depgraph/        Dependency graph: coverage, fanout, duplicates
├── scoring/         Risk scoring engine
├── governance/      Policy evaluation
├── quality/         Quality detectors (assertions, mocks, snapshots, coverage)
├── health/          Health detectors (slow, flaky, skipped, dead, unstable)
├── migration/       Migration readiness and blocker detection
├── impact/          Change impact analysis and test selection
├── metrics/         Aggregate metrics computation
├── summary/         Executive summary generation
├── ownership/       CODEOWNERS parsing and attribution
├── reporting/       Output formatting (text, JSON, markdown)
└── testcase/        Test case extraction and identity
```

The pipeline runs in a fixed order: scan → analyze → detect signals → build dependency graph → score → report. Each detector registers with the engine and receives the accumulated state from prior stages.

## Legacy Converter Engine

The JavaScript converter engine (`src/`, `bin/terrain.js`) provides multi-framework test conversion across 16 frameworks and 25 conversion directions. It remains functional and is accessible via `terrain convert`. This engine predates the Go analysis engine and serves as the migration acquisition wedge — the pain of framework migration is what brings teams to Terrain, and the analysis engine turns that pain into broader test intelligence.

## Current State

- **22 signal types** across 4 categories, all implemented
- **17 analysis detectors** registered in the engine
- **19 test frameworks** detected across 4 languages
- **6 policy rule types** with size-aware thresholds
- **4 canonical commands** with full JSON and terminal output
- **12 supporting commands** covering summary, posture, portfolio, metrics, comparison, migration, policy, test selection, PR analysis, entity drill-down, focus, and benchmark export
- **6 debug commands** for graph, coverage, fanout, duplicates, depgraph, and impact profiling
- **CI integration** via GitHub Actions composite action with PR comments and artifact upload
- **Golden/snapshot tests** validating all 4 canonical journeys
- **12 fixture repositories** covering all 6 personas and key edge cases
- **Benchmark matrix** against real open-source repositories (pandas, fastapi, express, etc.)
