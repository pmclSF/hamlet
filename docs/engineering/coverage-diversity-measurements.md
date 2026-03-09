# Coverage Diversity Measurements

## Overview

Coverage diversity measures whether the test suite is structurally diversified across test types or overreliant on broad tests (e2e, integration) or mocks. This is one of Hamlet's most distinctive product dimensions.

## Key Insight

Teams need to know not just whether code is exercised, but whether it is exercised only through coarse broad tests or through a healthy mix. Code covered only by e2e tests is covered, but fragily — those tests are slow, expensive, and harder to debug.

## Measurements

### coverage_diversity.mock_heavy_share

**What:** Share of test files dominated by mocks over assertions.

**How:** Counts `mockHeavyTest` signals, divides by total test file count.

**Evidence:** Always strong.

**Thresholds:**
| Ratio | Band |
|-------|------|
| ≤ 10% | strong |
| ≤ 25% | moderate |
| ≤ 40% | weak |
| > 40% | critical |

**Why it matters:** Mock-heavy tests verify wiring, not behavior. They can pass when behavior breaks and break when implementation changes.

### coverage_diversity.framework_fragmentation

**What:** Framework count relative to test suite size.

**How:** Ratio of distinct frameworks to total test files. Band escalation when ≥ 3 frameworks (moderate) or ≥ 5 / ratio > 0.3 (weak).

**Evidence:** Always strong.

**Why it matters:** Many frameworks in a small suite creates maintenance overhead, inconsistent patterns, and cognitive load.

### coverage_diversity.e2e_concentration

**What:** Share of test files using E2E frameworks (Cypress, Playwright, Selenium, etc.).

**How:** Identifies E2E frameworks by type, counts test files using them, computes ratio.

**Evidence:** Always strong.

**Thresholds:**
| Ratio | Band |
|-------|------|
| ≤ 50% | strong |
| ≤ 80% | moderate |
| > 80% | weak |

**Why it matters:** Suites dominated by E2E tests are slow, expensive, and prone to environmental flakiness. A healthy suite has most logic tested at the unit/integration level.

### coverage_diversity.e2e_only_units

**What:** Share of code units covered only by e2e tests (no unit or integration coverage).

**How:** Uses `CoverageSummary.CoveredOnlyByE2E` divided by `TotalCodeUnits`.

**Evidence:** Strong if coverage data is available. None if no coverage artifacts provided.

**Thresholds:**
| Ratio | Band |
|-------|------|
| ≤ 5% | strong |
| ≤ 15% | moderate |
| ≤ 30% | weak |
| > 30% | critical |

**Why it matters:** Code units covered only by e2e are the most fragile — if the e2e test breaks for environmental reasons, there's no faster test to catch regressions.

**Limitations:** Requires labeled coverage artifacts (`--coverage unit:path, --coverage e2e:path`).

### coverage_diversity.unit_test_coverage

**What:** Share of code units covered by unit tests.

**How:** Uses `CoverageSummary.CoveredByUnitTests` divided by `TotalCodeUnits`. Band logic is inverted (higher is better).

**Evidence:** Strong if coverage data available. None otherwise.

**Thresholds:**
| Ratio | Band |
|-------|------|
| ≥ 70% | strong |
| ≥ 50% | moderate |
| < 50% | weak |

**Why it matters:** Unit test coverage provides the fastest, most reliable feedback loop. Low unit coverage means the team depends on slower tests for regression detection.

## Test Fixtures

### Healthy layered coverage

```
100 test files: 70 unit, 20 integration, 10 e2e
Code units: 50, 40 covered by unit, 2 e2e-only
→ e2e_concentration = 0.10 (strong)
→ e2e_only_units = 0.04 (strong)
→ unit_test_coverage = 0.80 (strong)
→ Coverage diversity posture: strong
```

### E2E-heavy suite

```
50 test files: 5 unit, 5 integration, 40 e2e
Code units: 30, 20 e2e-only
→ e2e_concentration = 0.80 (moderate)
→ e2e_only_units = 0.67 (critical)
→ unit_test_coverage = 0.10 (weak)
→ Coverage diversity posture: critical
```

### Large unknown test-type share

```
No coverage artifacts provided.
→ e2e_only_units = 0.0 (unknown), evidence: none
→ unit_test_coverage = 0.0 (unknown), evidence: none
→ Limitation: "Provide labeled coverage artifacts for coverage-by-type analysis."
```

## Classification Confidence

Coverage diversity measurements are limited by test type classification confidence. If many test files have unknown type, the measurements may undercount certain categories. The `evidence` and `limitations` fields make this transparent.

## File

`internal/measurement/coverage.go` (diversity functions)
