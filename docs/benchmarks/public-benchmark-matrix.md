# Public Benchmark Matrix

## Purpose

This benchmark matrix runs Terrain against real-world public repositories to verify:

1. **Functional correctness** — All Terrain commands produce meaningful output on diverse codebases
2. **Determinism** — Identical inputs produce identical structured outputs across all JSON commands
3. **Performance** — Analysis completes in reasonable time at scale
4. **Degradation** — Terrain degrades gracefully on unsupported or edge-case repos
5. **Feature coverage** — Portfolio, migration, posture, and metrics all work on real codebases

## Repo selection criteria

Each repo was chosen to exercise specific Terrain capabilities:

| Repo | Category | Why chosen | Key exercises |
|------|----------|-----------|---------------|
| express | Backend JS | Moderate mocha suite, clean structure | Unit-test detection, assertion counting, code-unit discovery |
| fastify | Backend JS | Uses tap/node:test, moderate size | Multi-framework detection, alternative test runners |
| jest | JS monorepo | Large, tests itself with Jest | Monorepo workspace analysis, high test-file volume, meta-testing |
| playwright | E2E TS | Mix of Jest + Playwright self-tests | E2E framework detection, TypeScript, browser-test patterns |
| vue | Frontend TS monorepo | Vitest, TypeScript packages | Vitest detection, monorepo workspaces, frontend patterns |
| flask | Python backend | pytest, non-JS | Cross-language analysis, graceful degradation |
| next.js | Large TS monorepo | Jest + Playwright, very large | Scale, multi-framework at volume |
| storybook | Massive JS monorepo | 100+ packages, Jest + Vitest | Extreme scale, framework fragmentation, memory |

## Tiers

### Smoke (< 5 minutes total)
- **express**, **fastify**
- Quick feedback loop for CI or local development
- Should always pass without issues

### Full (10-30 minutes total)
- All smoke repos + **jest**, **playwright**, **vue**, **flask**, **next.js**
- The main representative matrix
- Covers language diversity, scale range, and framework variety

### Stress (30-60+ minutes total)
- All full repos + **storybook**
- For pre-release validation and scale regression testing
- Failures here are informational, not blocking

## What gets tested per repo

### Core commands (14 total)

| Command | Purpose | Output captured |
|---------|---------|----------------|
| `terrain analyze --json` | Core analysis | JSON snapshot, exit code, timing |
| `terrain analyze` | Human-readable output | Text output, exit code |
| `terrain summary` | Executive summary | Text output, exit code |
| `terrain posture` | Posture breakdown | Text output, exit code |
| `terrain posture --json` | Machine-readable posture | JSON posture, exit code |
| `terrain portfolio` | Portfolio intelligence | Text output, exit code |
| `terrain portfolio --json` | Machine-readable portfolio | JSON portfolio, exit code |
| `terrain metrics` | Metrics scorecard | Text output, exit code |
| `terrain metrics --json` | Machine-readable metrics | JSON metrics, exit code |
| `terrain migration readiness` | Migration assessment | Text output, exit code |
| `terrain migration readiness --json` | Machine-readable migration | JSON migration, exit code |
| `terrain migration blockers` | Migration blocker list | Text output, exit code |
| `terrain policy check` | Policy evaluation | Text output, exit code |
| `terrain export benchmark` | Privacy-safe export | JSON export, exit code |

### Determinism checks (6 per repo)

Each of these commands is run twice and the JSON outputs are compared with timestamps stripped:
- `analyze --json`
- `metrics --json`
- `portfolio --json`
- `posture --json`
- `migration readiness --json`
- `export benchmark`

Any semantic difference is flagged as a determinism failure.

### Expectation checks

Each repo has expectations in `benchmarks/expectations/<id>.yaml`:
- Minimum test files detected
- Minimum code units detected
- Posture dimensions must exist
- Portfolio command must succeed
- Migration command must succeed

These catch obvious regressions without hardcoding exact outputs.

## Adding a new repo

1. Add an entry to `benchmarks/public-repos.yaml`
2. Create `benchmarks/expectations/<id>.yaml` with baseline expectations
3. Run `make benchmark-fetch` to download
4. Run `make benchmark-smoke` or `make benchmark-full` to verify
5. Update this document's table

## Interpreting failures

| Failure type | Meaning | Action |
|-------------|---------|--------|
| Command exit non-zero | Terrain crashed or errored | Investigate — likely a bug |
| Expectation miss | Fewer tests/units than expected | Check if repo changed or Terrain regressed |
| Determinism mismatch | Non-deterministic output | Investigate — timestamps, map ordering, etc. |
| Timeout | Analysis too slow | Check for performance regression or mark as stress |
| Clone failure | Network or repo issue | Retry; if persistent, mark repo as flaky |
| Policy "no file" | Repo has no `.terrain/policy.yaml` | Expected for public repos — not a failure |

## Total invocations per run

Per repo: 14 commands + 12 determinism runs (6 commands x 2) = 26 Terrain invocations

| Tier | Repos | Invocations | Estimated time |
|------|-------|-------------|----------------|
| Smoke | 2 | 52 | 2-5 min |
| Full | 7 | 182 | 10-30 min |
| Stress | 8 | 208 | 30-60 min |
