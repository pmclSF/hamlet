# Demo Fixtures

Terrain ships with canonical demo fixtures in `fixtures/demos/` that showcase its strongest insights in realistic, reproducible scenarios.

## Fixtures

### 1. healthy-balanced.json — Well-maintained repo

**Scenario:** A JavaScript API project with Jest unit tests and Playwright E2E tests. Few issues, good assertion density, clean posture.

**What it demonstrates:**
- Terrain produces useful output even on healthy repos
- Posture dimensions show "strong" with evidence
- The report is concise, not overwhelming

**Key wow moment:** "No significant issues — here's exactly why."

### 2. flaky-concentrated.json — Instability concentrated in auth

**Scenario:** A payments service where 3 auth test files account for all flaky/unstable signals. The rest of the suite is healthy.

**What it demonstrates:**
- Flakiness concentrates in a small number of files
- Those same files also have weak assertions (compounding risk)
- Runtime evidence backs the health assessment

**Key wow moment:** "3 test files in src/auth/ account for 100% of your instability."

### 3. e2e-heavy-shallow.json — E2E-heavy shallow coverage

**Scenario:** A storefront app with 25 Cypress tests and only 5 Jest unit tests. All 5 exported service classes have no linked unit tests.

**What it demonstrates:**
- E2E concentration (83% of test files are E2E)
- Public functions covered only by E2E
- Coverage diversity posture is weak

**Key wow moment:** "5 exported functions in your core services have no unit test coverage — only E2E tests exercise them."

### 4. fragmented-migration-risk.json — Multi-framework migration risk

**Scenario:** An enterprise platform with 5 test frameworks (Jest, Mocha, Jasmine, Cypress, Protractor), migration blockers in legacy tests, and quality issues compounding the risk.

**What it demonstrates:**
- Framework fragmentation (5 frameworks across 58 files)
- Migration blockers compounded by weak assertions
- Area assessment: legacy directory is "risky"
- Legacy framework governance signals

**Key wow moment:** "src/legacy/ has 2 migration blockers compounded by 2 quality issues. Address quality before migrating."

### 5. bloated-overlapping-tests.json — Overlapping broad E2E tests

**Scenario:** A DevTools Platform repo with broad E2E tests that overlap heavily, consuming disproportionate CI runtime while duplicating coverage across modules.

**What it demonstrates:**
- Redundancy detection (multiple E2E tests covering 90% the same modules)
- Overbreadth (individual tests covering far more than necessary)
- Runtime concentration (60% of CI time in 3 tests)
- High-leverage test identification (fast unit tests with broad coverage)

**Key wow moment:** "3 E2E tests cover 90% the same modules and consume 60% of CI runtime. Here are the redundancy candidates."

## Using fixtures

Run Terrain against a fixture to see the demo output:

```bash
# These are snapshot fixtures, not live repos.
# Use them to test rendering and validate output.
cat fixtures/demos/flaky-concentrated.json | terrain analyze --json | terrain summary --json
```

## Adding new fixtures

When adding a fixture:
1. Create a realistic scenario with believable file paths and signal counts
2. Focus on one primary insight per fixture
3. Document the expected wow moment
4. Ensure the fixture validates against the current TestSuiteSnapshot schema
