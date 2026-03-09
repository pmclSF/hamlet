# Structural Risk Measurements

## Overview

Structural risk measurements identify where the test codebase is fragile due to migration blockers, deprecated patterns, or dynamic test generation. They feed the `structural_risk` posture dimension.

## Key Distinction

Structural risk is grounded in **technical fragility**, not vague "complexity." It measures concrete obstacles to modernization and maintenance, not subjective code quality.

## Measurements

### structural_risk.migration_blocker_density

**What:** Density of migration blockers relative to test files.

**How:** Counts signals of type `migrationBlocker`, `deprecatedTestPattern`, `dynamicTestGeneration`, and `customMatcherRisk`, divides by total test file count.

**Evidence:** Always strong.

**Thresholds:**
| Ratio | Band |
|-------|------|
| ≤ 5% | strong |
| ≤ 15% | moderate |
| ≤ 30% | weak |
| > 30% | critical |

**Why it matters:** Migration blockers are the obstacles that prevent framework modernization. High density means the suite will resist migration efforts and require significant rework.

### structural_risk.deprecated_pattern_share

**What:** Share of test files using deprecated patterns.

**How:** Counts `deprecatedTestPattern` signals, divides by total test file count.

**Evidence:** Always strong.

**Thresholds:**
| Ratio | Band |
|-------|------|
| ≤ 5% | strong |
| ≤ 15% | moderate |
| ≤ 30% | weak |
| > 30% | critical |

**Why it matters:** Deprecated patterns will stop working in future framework versions. Proactively updating them reduces migration risk and prevents surprise breakage.

### structural_risk.dynamic_generation_share

**What:** Share of test files using dynamic test generation (runtime-generated test cases).

**How:** Counts `dynamicTestGeneration` signals, divides by total test file count.

**Evidence:** Always strong.

**Thresholds:**
| Ratio | Band |
|-------|------|
| ≤ 5% | strong |
| ≤ 10% | moderate |
| ≤ 20% | weak |
| > 20% | critical |

**Why it matters:** Dynamic test generation makes test identity unstable across runs, complicating longitudinal tracking, flakiness analysis, and migration automation. These tests are harder to reason about statically.

## Overlap Documentation

Structural risk shares some input signals with other dimensions:

| Signal | Also used in | Distinction |
|--------|-------------|-------------|
| `deprecatedTestPattern` | Migration readiness | Structural risk measures density; migration measures blocker impact |
| `customMatcherRisk` | Migration readiness | Same |
| `dynamicTestGeneration` | Test identity analysis | Structural risk measures share; identity tracks stability |

This overlap is intentional — the same signal can indicate risk in multiple dimensions. The measurements are independent computations.

## Test Fixtures

### Concentrated auth/payments risk

```
50 test files
Migration blockers: 12 (8 in src/auth/, 4 in src/payments/)
→ structural_risk.migration_blocker_density = 0.24 (weak)
→ Structural risk posture: weak
```

### Diffuse low-risk repo

```
200 test files
Deprecated patterns: 3, dynamic generation: 1
→ structural_risk.deprecated_pattern_share = 0.015 (strong)
→ structural_risk.dynamic_generation_share = 0.005 (strong)
→ Structural risk posture: strong
```

### Strong public-surface fragility

```
30 test files
Migration blockers: 15 (includes custom matchers and deprecated patterns)
→ structural_risk.migration_blocker_density = 0.50 (critical)
→ Structural risk posture: critical
```

## File

`internal/measurement/structural_risk.go`
