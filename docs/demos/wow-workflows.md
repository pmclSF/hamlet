# Wow Workflows

These are end-to-end workflows that reliably produce Terrain's strongest insights.

## 1. Concentrated instability

**Trigger:** `terrain analyze` on a repo with flaky tests

**What the user sees:**
```
Posture
  health:                    WEAK

Signals
  health           5
  Breakdown:
    flakyTest                3
    slowTest                 2
```

**Follow-up:** `terrain posture` shows health dimension driven by flaky_share measurement with 30% ratio.

**Action:** Fix the 3 flaky test files. They are listed by file path in the analyze output.

## 2. Public functions covered only by E2E

**Trigger:** `terrain analyze` on a repo with high E2E concentration

**What the user sees:**
```
Posture
  coverage_depth:            WEAK
  coverage_diversity:        WEAK

Signals
  quality          5
  Breakdown:
    untestedExport           5
```

**Follow-up:** `terrain posture` shows uncovered_exports at 100% (5 of 5 exported units untested) and e2e_concentration at 83%.

**Action:** Add unit tests for the 5 exported services. The E2E tests exist but do not exercise these functions at the unit level.

## 3. Migration risk compounded by quality issues

**Trigger:** `terrain analyze` on a repo with migration blockers

**What the user sees:**
```
Posture
  structural_risk:           WEAK
  coverage_depth:            MODERATE

Migration Blockers
  src/legacy/auth.spec.js     asyncPattern
  src/legacy/payments.spec.js customMatcher
```

**Follow-up:** `terrain posture` shows migration_blocker_density at 20%. The summary report shows area assessment: "src/legacy/ is risky — 2 migration blockers compounded by 2 quality issues."

**Action:** Address quality issues in legacy tests before attempting framework migration.

## 4. Framework fragmentation

**Trigger:** `terrain analyze` on a multi-framework repo

**What the user sees:**
```
Frameworks
  jest                 20 files [unit]
  mocha                15 files [unit]
  jasmine               8 files [unit]
  cypress              10 files [e2e]
  protractor            5 files [e2e]

Posture
  coverage_diversity:        WEAK
```

**Follow-up:** `terrain posture` shows framework_fragmentation at 0.086 with 5 frameworks. `terrain metrics` shows the full scorecard.

**Action:** Consolidate to fewer frameworks. The migration readiness assessment tells you where to start.

## 5. One area carries disproportionate risk

**Trigger:** `terrain summary` on a repo with concentrated signals

**What the user sees:**
```
Highest-Risk Directories
  src/auth/                    HIGH  (7 signals)
  src/api/                     MEDIUM  (3 signals)
  src/utils/                   LOW  (1 signal)
```

**Follow-up:** The auth directory has 70% of all signals despite containing only 20% of test files. `terrain analyze` shows the specific signals.

**Action:** Prioritize auth test remediation — it is the highest-leverage improvement.

## 6. Portfolio intelligence reveals CI waste

**Trigger:** `terrain portfolio` on a repo with overlapping E2E tests

**What the user sees:**
```
Portfolio Summary
  Total tests:              48
  Runtime concentration:    60% in 3 tests

Redundancy Candidates
  e2e/checkout-full.spec.ts     overlaps 90% with e2e/checkout-quick.spec.ts
  e2e/onboarding-full.spec.ts   overlaps 87% with e2e/onboarding-smoke.spec.ts

High-Leverage Tests
  unit/payment-service.test.ts   covers 12 modules in 0.4s
  unit/auth-validator.test.ts    covers 8 modules in 0.2s
```

**Follow-up:** `terrain analyze` shows the same redundancy findings surfaced as signals. `terrain posture` reflects the runtime concentration in the operational risk dimension.

**Demo fixture:** `bloated-overlapping-tests.json`

**Action:** Consolidate overlapping E2E tests or replace them with targeted integration tests. The high-leverage tests show where fast, focused tests already provide strong coverage.
