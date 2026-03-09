# Operational Risk Measurements

## Overview

Operational risk measurements capture whether the test suite is operationally manageable — policies followed, frameworks current, runtime budgets met. They feed the `operational_risk` posture dimension.

## Key Distinction

Operational risk is distinct from quality and health. It focuses on organizational and operational controls, not technical correctness or test reliability.

## Governance/Policy Decision

Governance and policy signals feed operational risk directly rather than existing as a separate dimension. Rationale:

1. Policy violations and legacy framework usage are operationally scoped
2. Keeping the dimension count small (5) prevents metric sprawl
3. If governance grows to need its own dimension, it can be split later

## Measurements

### operational_risk.policy_violation_density

**What:** Density of policy violations relative to test files.

**How:** Counts `policyViolation` signals, divides by total test file count.

**Evidence:** Always strong.

**Thresholds:**
| Ratio | Band |
|-------|------|
| = 0% | strong |
| ≤ 5% | moderate |
| ≤ 15% | weak |
| > 15% | critical |

**Why it matters:** Policy violations indicate the team's own rules are not being followed. These represent explicit team decisions about what is acceptable.

### operational_risk.legacy_framework_share

**What:** Share of test files using legacy frameworks.

**How:** Counts `legacyFrameworkUsage` signals, divides by total test file count.

**Evidence:** Always strong.

**Thresholds:**
| Ratio | Band |
|-------|------|
| ≤ 5% | strong |
| ≤ 15% | moderate |
| ≤ 30% | weak |
| > 30% | critical |

**Why it matters:** Legacy frameworks may lack security patches, community support, and modern features. They create hiring friction and limit tooling options.

### operational_risk.runtime_budget_breach_share

**What:** Share of test files exceeding runtime budgets.

**How:** Counts `runtimeBudgetExceeded` signals, divides by total test file count.

**Evidence:** Strong if runtime data is available. Weak otherwise.

**Thresholds:**
| Ratio | Band |
|-------|------|
| ≤ 5% | strong |
| ≤ 15% | moderate |
| ≤ 30% | weak |
| > 30% | critical |

**Why it matters:** Runtime budget breaches lengthen CI pipelines and increase infrastructure costs. They often indicate tests that should be restructured or moved to a different test tier.

**Limitations:** Without runtime data, this measurement is based on static analysis only.

## Phrasing

Operational risk reporting follows Hamlet's ownership-safe phrasing conventions:

| Do | Don't |
|----|-------|
| "Policy violations concentrated in payments area" | "Team-payments violates policy" |
| "Legacy framework usage in 15% of test files" | "Team is using outdated tools" |
| "Runtime budgets exceeded in auth module" | "Auth team's tests are too slow" |

## Test Fixtures

### Fully compliant repo

```
100 test files, 0 policy violations, 0 legacy frameworks
→ All measurements: strong
→ Operational risk posture: strong
```

### Fragmented ownership repo

```
80 test files, 12 policy violations, 8 legacy framework files
→ operational_risk.policy_violation_density = 0.15 (weak)
→ operational_risk.legacy_framework_share = 0.10 (moderate)
→ Operational risk posture: weak
```

### Governance-heavy repo

```
50 test files, 15 policy violations, 20 legacy framework files
→ operational_risk.policy_violation_density = 0.30 (critical)
→ operational_risk.legacy_framework_share = 0.40 (critical)
→ Operational risk posture: critical
```

## File

`internal/measurement/operational_risk.go`
