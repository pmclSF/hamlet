# Public Benchmark Matrix

## Purpose

This benchmark matrix runs Hamlet against real-world public repositories to verify:

1. **Functional correctness** — Hamlet produces meaningful output on diverse codebases
2. **Determinism** — Identical inputs produce identical structured outputs
3. **Performance** — Analysis completes in reasonable time at scale
4. **Degradation** — Hamlet degrades gracefully on unsupported or edge-case repos

## Repo selection criteria

Each repo was chosen to exercise specific Hamlet capabilities:

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

### Smoke (< 2 minutes total)
- **express**, **fastify**
- Quick feedback loop for CI or local development
- Should always pass without issues

### Full (5-15 minutes total)
- All smoke repos + **jest**, **playwright**, **vue**, **flask**, **next.js**
- The main representative matrix
- Covers language diversity, scale range, and framework variety

### Stress (15-30+ minutes total)
- All full repos + **storybook**
- For pre-release validation and scale regression testing
- Failures here are informational, not blocking

## What gets tested per repo

For each repo, the runner executes:

| Command | Purpose | Output captured |
|---------|---------|----------------|
| `hamlet analyze --json` | Core analysis | JSON snapshot, exit code, timing |
| `hamlet analyze` | Human-readable output | Text output, exit code |
| `hamlet summary` | Executive summary | Text output, exit code |
| `hamlet posture` | Posture breakdown | Text output, exit code |
| `hamlet metrics --json` | Metrics snapshot | JSON metrics, exit code |
| `hamlet export benchmark` | Privacy-safe export | JSON export, exit code |

### Determinism check
- `hamlet analyze --json` is run twice per repo
- Outputs are compared with timestamps stripped
- Any semantic difference is flagged

### Expectation checks
- Each repo has optional expectations in `benchmarks/expectations/<id>.yaml`
- Checks include: minimum test files, minimum code units, required frameworks
- These catch obvious regressions without hardcoding exact outputs

## Adding a new repo

1. Add an entry to `benchmarks/public-repos.yaml`
2. Create `benchmarks/expectations/<id>.yaml` with baseline expectations
3. Run `make benchmark-fetch` to download
4. Run `make benchmark-smoke` or `make benchmark-full` to verify
5. Update this document's table

## Interpreting failures

| Failure type | Meaning | Action |
|-------------|---------|--------|
| Command exit non-zero | Hamlet crashed or errored | Investigate — likely a bug |
| Expectation miss | Fewer tests/units than expected | Check if repo changed or Hamlet regressed |
| Determinism mismatch | Non-deterministic output | Investigate — timestamps, map ordering, etc. |
| Timeout | Analysis too slow | Check for performance regression or mark as stress |
| Clone failure | Network or repo issue | Retry; if persistent, mark repo as flaky |
