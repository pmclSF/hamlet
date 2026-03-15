# Coverage Depth Measurements

## Overview

Coverage depth measurements assess whether code is structurally covered — not just by lines executed, but by meaningful assertions against exported surfaces. They feed the `coverage_depth` posture dimension.

## Key Distinction

Coverage depth is **not** raw line coverage. A repo with 85% line coverage but 40% of its exported functions untested has weak coverage depth. Line coverage measures execution breadth; coverage depth measures structural thoroughness.

## Measurements

### coverage_depth.uncovered_exports

**What:** Share of exported code units without any linked tests.

**How:** Counts exported `CodeUnit` entries, counts `untestedExport` signals, computes ratio.

**Evidence:** Partial — test linkage is heuristic-based (import tracing, naming conventions). Some coverage may exist but not be detected.

**Thresholds:**
| Ratio | Band |
|-------|------|
| ≤ 10% | strong |
| ≤ 30% | moderate |
| ≤ 50% | weak |
| > 50% | critical |

**Why it matters:** Exported functions are the public API surface. Untested exports are the most likely source of regression when other code depends on them.

**Limitations:** "Test linkage is heuristic-based; some coverage may exist but not be detected."

### coverage_depth.weak_assertion_share

**What:** Share of test files with weak assertion density (few assertions relative to test complexity).

**How:** Counts `weakAssertion` signals, divides by total test file count.

**Evidence:** Always strong — assertion density is a static analysis finding.

**Thresholds:**
| Ratio | Band |
|-------|------|
| ≤ 10% | strong |
| ≤ 25% | moderate |
| ≤ 50% | weak |
| > 50% | critical |

**Why it matters:** Tests with few assertions may pass without actually verifying behavior. They contribute to line coverage without providing safety.

### coverage_depth.coverage_breach_share

**What:** Share of coverage areas below configured thresholds.

**How:** Counts `coverageThresholdBreak` signals, divides by total test file count.

**Evidence:** Strong if coverage data is available. Weak if no coverage artifacts were provided.

**Thresholds:**
| Ratio | Band |
|-------|------|
| ≤ 5% | strong |
| ≤ 15% | moderate |
| ≤ 30% | weak |
| > 30% | critical |

**Why it matters:** Coverage thresholds represent the team's own standards. Breaches indicate regression below accepted minimums.

**Limitations:** Without coverage data, this measurement reports "strong" with evidence: weak. Provide `--coverage` artifacts for accurate results.

## Test Fixtures

### Healthy deep coverage

```
Exported code units: 50
Untested export signals: 3
Weak assertion signals: 2 / 80 test files
→ coverage_depth.uncovered_exports = 0.06 (strong)
→ coverage_depth.weak_assertion_share = 0.025 (strong)
→ Coverage depth posture: strong
```

### Broad line coverage but weak structural depth

```
Line coverage: 82%
Exported code units: 40
Untested export signals: 15
→ coverage_depth.uncovered_exports = 0.375 (weak)
→ Coverage depth posture: weak
```

### Missing branch data

```
No coverage artifacts provided.
→ coverage_depth.coverage_breach_share = 0.0 (strong), evidence: weak
→ Limitation: "No coverage data available"
```

### Public API weak coverage

```
Exported code units: 30
Untested export signals: 20
→ coverage_depth.uncovered_exports = 0.67 (critical)
→ Coverage depth posture: critical
```

## Artifact Limitations

Coverage depth depends on:

1. **Code unit discovery** — Terrain's static analysis may not find all exported symbols, especially in dynamic languages.
2. **Test linkage** — The heuristic linking tests to code units may miss indirect coverage.
3. **Coverage artifacts** — Providing `--coverage` improves breach detection accuracy.

The measurement is explicit about these limitations via the `evidence` and `limitations` fields.

## File

`internal/measurement/coverage.go` (first three functions)
