# Test Fixtures

Fixture repositories used for snapshot tests, benchmarks, and pipeline validation.

## Fixture Matrix

| Fixture | Persona / Edge Case | Language | Framework(s) | Key Signals Triggered |
|---------|---------------------|----------|--------------|----------------------|
| `sample-repo/` | General (all journeys) | TypeScript | Vitest | duplicates, skipped tests, fanout |
| `edgecases/` | Manual coverage overlay | YAML config only | — | manual coverage signals |
| `frontend-react/` | Frontend developer | TypeScript/JSX | Jest + Playwright + Enzyme | snapshot-heavy, multi-framework, deprecated pattern |
| `backend-api/` | Backend developer | Go | go-testing | coverage available, runtime available, untested exports |
| `mobile-cross-platform/` | Mobile / device-sensitive | TypeScript | Vitest | skipped tests (platform-conditional) |
| `qa-manual-overlay/` | QA / manual tester | TypeScript | Vitest | manual coverage, CODEOWNERS, policy evaluation |
| `ai-eval-suite/` | AI / ML evaluation | Python | pytest | parametrized tests, skipped tests, weak assertions |
| `high-fanout/` | High-fanout codebase | TypeScript | Vitest | high-fanout node (12 dependents), fixture concentration |
| `duplicate-heavy/` | Duplicate-heavy suite | TypeScript | Vitest | duplicate clusters (shared fixtures/helpers, >0.60 similarity) |
| `weak-coverage/` | Low coverage codebase | TypeScript | Vitest | untested exports, coverage threshold break, weak coverage band |
| `skipped-tests/` | Skip-debt-heavy suite | TypeScript | Vitest | 59% skip rate, high skip burden edge case |
| `legacy-mixed/` | Legacy mixed-style repo | JavaScript/JSX | Jest + Mocha + Cypress + Enzyme + Sinon | framework migration, migration blockers, deprecated patterns, mixed cultures |

## Fixture Details

### `sample-repo/`

Primary fixture for all 4 canonical user journeys. Has git history for impact analysis.

- **13 source files** (auth, api, db, cache, utils, config)
- **12 test files** (unit + integration, Vitest)
- **3 test fixture modules** + **3 helper modules**
- **Git history:** 2 commits (HEAD~1 adds package.json + skipped legacy tests)
- **Signals:** duplicate candidates (login + login-extended), 3 skipped tests, high-fanout db fixture

**Used by:** `cmd/terrain/snapshot_test.go` (analyze, insights, impact, explain golden tests)

### `edgecases/`

Minimal fixture for manual coverage overlay testing.

- `terrain.yaml` with 5 manual coverage entries (TestRail, Jira)
- Low CI duration (45s)

### `frontend-react/`

Snapshot-heavy React component tests with multi-framework detection.

- **4 source files** (3 components + auth hook)
- **4 test files** (3 Jest component tests + 1 Playwright E2E spec)
- **Snapshot ratio:** 7/14 assertions are `toMatchSnapshot()` (50%, above 40% threshold)
- **Frameworks:** Jest + Playwright + Enzyme (deprecated dep in package.json)
- **Signals:** `snapshotHeavyTest`, `frameworkMigration`, `deprecatedTestPattern`

### `backend-api/`

Go backend with coverage and runtime artifacts for data-completeness testing.

- **5 source files** (3 handlers + 2 middleware, Go)
- **4 test files** (Go testing package)
- **coverage.lcov:** partial coverage (handlers covered, `middleware/cors.go` at 0%)
- **junit.xml:** 11 tests, 0 failures, timing data
- **Signals:** `untestedExport` (cors functions), coverage/runtime both available

### `mobile-cross-platform/`

Device-conditional tests where many tests are skipped due to hardware requirements.

- **5 source files** (iOS, Android, shared platform, camera, GPS)
- **5 test files** (Vitest)
- **Skip ratio:** 8/15 tests skipped (platform-conditional: Face ID, APNS, biometric, FCM, GPS hardware)
- **Signals:** `skippedTest` (platform-conditional), skipped test burden

### `qa-manual-overlay/`

Thin automated coverage supplemented by QA manual test suites tracked in external tools.

- **4 source files** (billing, onboarding)
- **2 test files** (minimal automated tests)
- **`.terrain/terrain.yaml`** — 4 manual coverage entries (TestRail, Jira)
- **`.terrain/policy.yaml`** — 60% coverage threshold
- **`CODEOWNERS`** — billing-team, platform-team, qa-lead
- **Signals:** manual coverage presence, policy evaluation, ownership resolution

### `ai-eval-suite/`

Python ML evaluation tests with accuracy thresholds and regression checks.

- **4 source modules** (classifier, embeddings, metrics, dataset)
- **4 pytest test files** (accuracy thresholds, embedding similarity, output format, regression)
- **Patterns:** `@pytest.mark.parametrize`, `@pytest.mark.skip`, loose-bound assertions (`>= 0.6`)
- **Signals:** `skippedTest` (2 skipped: sarcasm detection, multilingual), pytest framework detection

### `high-fanout/`

Single shared database utility imported by 12 test files, exceeding the default fanout threshold of 10.

- **5 source files** (database utility, auth utility, 3 service modules)
- **12 test files** — all import `src/utils/database.ts`
- **Fanout:** `database.ts` has direct fanout of 12 (threshold default: 10)
- **Signals:** high-fanout node, high-fanout fixture edge case

### `duplicate-heavy/`

Structurally similar test files designed to exceed the 0.60 duplicate similarity threshold.

- **3 source validators** (email, phone, address)
- **1 shared fixture** (`valid-inputs.ts`) + **1 shared helper** (`validator-assertions.ts`)
- **6 test files** — all import same fixture + helper, follow identical describe/it structure
- **Similarity dimensions:** fixture overlap (shared `valid-inputs.ts`), helper overlap (shared `validator-assertions.ts`), suite path similarity (`tests/validators/*`), assertion pattern similarity (identical `expect`/`expectValid`/`expectNormalized` patterns)
- **Signals:** duplicate clusters, redundant suite edge case

### `weak-coverage/`

Many exported modules with only one test file, plus LCOV data showing the coverage gap.

- **6 source modules** (engine, parser, transformer, optimizer, serializer, helpers) — 18+ exported functions
- **1 test file** — covers only `engine.ts`
- **`coverage/lcov.info`** — engine.ts at 100%, all others at 0%
- **Signals:** `untestedExport` (5 untested modules), `coverageThresholdBreak`, weak coverage band

### `skipped-tests/`

Auth module suite where >50% of tests are skipped due to infrastructure dependencies.

- **5 source modules** (oauth, saml, ldap, mfa, basic auth)
- **5 test files** — 10 of 17 tests skipped (59% skip rate)
- **Fully skipped:** `saml.test.ts` (3/3), `ldap.test.ts` (4/4)
- **Partially skipped:** `oauth.test.ts` (2/4), `mfa.test.ts` (1/3)
- **Healthy baseline:** `basic.test.ts` (0/3 skipped)
- **Signals:** `skippedTest`, high skip burden edge case (>20% threshold)

### `legacy-mixed/`

Multi-framework repo representing a real-world legacy codebase mid-migration.

- **3 source modules** (app, header component, API utils)
- **6 test files** across 4 frameworks:
  - `app.test.js` — Jest
  - `app.mocha.js` — Mocha + Chai (duplicate coverage of same module)
  - `Header.enzyme.test.jsx` — Enzyme (deprecated)
  - `api-callback.test.js` — Jest with `done()` callbacks (deprecated pattern)
  - `api-sinon.test.js` — Jest with Sinon stubs (migration blocker)
  - `cypress/e2e/login.cy.js` — Cypress E2E
- **Config files:** `jest.config.js`, `.mocharc.yml`, `cypress.config.js`
- **Signals:** `frameworkMigration`, `migrationBlocker` (Enzyme, Sinon), `deprecatedTestPattern` (done callbacks), mixed test cultures edge case

## Adding New Fixtures

When adding a fixture repo:
1. Create a directory under `tests/fixtures/`
2. Initialize git if the fixture needs `terrain impact` (most don't)
3. Include a `package.json`, `go.mod`, or `requirements.txt` for framework detection
4. Keep fixtures small (5-20 files) but rich enough to trigger target signals
5. Document which signals the fixture triggers in this README
6. Add the fixture to `benchmarks/repos.json` if it should be benchmarked
