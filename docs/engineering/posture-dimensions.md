# Posture Dimensions

## Overview

Terrain reports posture across five core dimensions. Each dimension anchors a small set of measurements that can be summarized into a stable band (strong, moderate, weak, elevated, critical).

The dimension set is intentionally small and maps directly to user action.

## Dimensions

### 1. Health

**What it means:** Are the tests themselves healthy — reliable, not skipped, not dead, not slow?

**What it measures:**
- Flaky/unstable test share
- Skip density
- Dead test share
- Slow test share

**What it does NOT claim:** Health does not measure code quality, coverage, or correctness. A healthy test suite can still have weak coverage.

**Band semantics:**
| Band | Meaning |
|------|---------|
| Strong | Very few flaky, skipped, dead, or slow tests |
| Moderate | Some instability or maintenance debt present |
| Weak | Significant reliability or maintenance issues |
| Elevated | Widespread instability affecting multiple areas |
| Critical | Suite is unreliable enough to erode trust |

**Signals:** `flakyTest`, `unstableSuite`, `skippedTest`, `deadTest`, `slowTest`

### 2. Coverage Depth

**What it means:** How deeply is the code structurally covered — not just by lines, but by meaningful assertions against exported surfaces?

**What it measures:**
- Uncovered exported code units
- Weak assertion density
- Coverage threshold breaches

**What it does NOT claim:** Coverage depth does not measure coverage diversity (what types of tests cover the code). Deep coverage by only one test type may still be fragile.

**Band semantics:**
| Band | Meaning |
|------|---------|
| Strong | Most exported code is tested with meaningful assertions |
| Moderate | Some exported code lacks test coverage |
| Weak | Significant gaps in structural coverage |
| Elevated | Large portions of exported API are untested |
| Critical | Most exported code lacks test coverage |

**Signals:** `untestedExport`, `weakAssertion`, `coverageThresholdBreak`

### 3. Coverage Diversity

**What it means:** Is the test suite structurally diversified across test types, or does it overrely on broad tests (e2e, integration) or mocks?

**What it measures:**
- Mock-heavy test share
- Framework fragmentation
- E2E concentration
- Code units covered only by e2e
- Unit test coverage share

**What it does NOT claim:** There is no single "ideal mix" for every repo. Diversity measures imbalance and structural fragility, not conformity to a template.

**Band semantics:**
| Band | Meaning |
|------|---------|
| Strong | Healthy mix of test types; low mock and e2e concentration |
| Moderate | Some overreliance on broad tests or mocks |
| Weak | Significant concentration in one test type |
| Elevated | Most coverage comes from a single test type |
| Critical | Extreme overreliance creating structural fragility |

**Signals:** `mockHeavyTest`, frameworks, coverage summary

### 4. Structural Risk

**What it means:** Where is the repo fragile due to migration blockers, deprecated patterns, or dynamic test generation?

**What it measures:**
- Migration blocker density
- Deprecated pattern share
- Dynamic test generation share

**What it does NOT claim:** Structural risk does not measure runtime behavior or operational issues. It focuses on technical fragility in the test code itself.

**Band semantics:**
| Band | Meaning |
|------|---------|
| Strong | Few or no structural risk factors |
| Moderate | Some migration blockers or deprecated patterns |
| Weak | Significant structural obstacles to modernization |
| Elevated | Widespread technical debt blocking migration |
| Critical | Test infrastructure is structurally fragile |

**Signals:** `migrationBlocker`, `deprecatedTestPattern`, `dynamicTestGeneration`, `customMatcherRisk`

### 5. Operational Risk

**What it means:** Is the test suite operationally manageable — are policies followed, frameworks current, budgets met?

**What it measures:**
- Policy violation density
- Legacy framework share
- Runtime budget breach share

**What it does NOT claim:** Operational risk does not measure code quality or coverage. It focuses on whether organizational and operational controls are in place.

**Band semantics:**
| Band | Meaning |
|------|---------|
| Strong | Policies followed, frameworks current, budgets met |
| Moderate | Minor operational issues present |
| Weak | Significant policy violations or legacy framework usage |
| Elevated | Widespread operational problems |
| Critical | Operational controls are largely absent |

**Signals:** `policyViolation`, `legacyFrameworkUsage`, `runtimeBudgetExceeded`

## Governance/Policy Decision

Governance and policy signals currently feed into the **operational_risk** dimension rather than existing as a separate dimension. The rationale:

1. Governance signals (policy violations, legacy framework usage) are operationally scoped — they affect how manageable the suite is, not its technical correctness.
2. Keeping the dimension count small (5) prevents metric sprawl and makes posture summaries scannable.
3. If governance grows to include distinct concerns (e.g., compliance, audit trails), it can be split into a separate dimension in a future version.

This decision is documented here so future contributors can revisit it with evidence.

## Design Constraints

- **Intentionally small set.** Five dimensions is enough to be actionable without being overwhelming.
- **Each dimension maps to action.** "Health is weak" → fix flaky tests. "Coverage depth is weak" → test untested exports.
- **No vague dimensions.** Each dimension has clear scope and explicit exclusions.
- **Stable for trends.** Dimensions don't change across versions without migration. This enables meaningful longitudinal comparison.
