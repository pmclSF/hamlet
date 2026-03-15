# Example: `terrain posture` Output

This document shows representative output from `terrain posture` (the detailed posture breakdown with measurement evidence) run against a repository with E2E-heavy test distribution and limited unit test coverage.

## Terminal Output

```
Terrain Posture Report
============================================================

HEALTH
  Posture: STRONG
  All health measurements are within acceptable thresholds.

  Driving measurements: health.skip_density

  Measurements:
    health.flaky_share                       0.0% [strong]
      Evidence: weak
      0 of 38 test file(s) flagged as flaky or unstable (0%).
      * No runtime artifacts provided; flaky detection is heuristic-only.
    health.skip_density                      2.6% [strong]
      Evidence: strong
      1 of 38 test file(s) contain skipped tests (3%).
    health.dead_test_share                   0.0% [strong]
      Evidence: strong
      0 of 38 test file(s) contain dead tests (0%).
    health.slow_test_share                   0.0% [strong]
      Evidence: weak
      0 of 38 test file(s) flagged as slow (0%).
      * No runtime artifacts provided; slow test detection is heuristic-only.

------------------------------------------------------------

COVERAGE DEPTH
  Posture: WEAK
  Multiple exported code units have no linked tests.

  Driving measurements: coverage_depth.uncovered_exports

  Measurements:
    coverage_depth.uncovered_exports         35.7% [weak]
      Evidence: partial
      10 of 28 exported code unit(s) appear untested (36%).
      * Test linkage is heuristic-based; some coverage may exist but not be detected.
    coverage_depth.weak_assertion_share      13.2% [moderate]
      Evidence: strong
      5 of 38 test file(s) have weak assertion density (13%).
    coverage_depth.coverage_breach_share     0.0% [strong]
      Evidence: weak
      No coverage threshold breaches detected.
      * No coverage data available; result may improve with coverage artifacts.

------------------------------------------------------------

COVERAGE DIVERSITY
  Posture: MODERATE
  E2E tests account for a large share of the test suite.

  Driving measurements: coverage_diversity.e2e_concentration

  Measurements:
    coverage_diversity.mock_heavy_share      5.3% [strong]
      Evidence: strong
      2 of 38 test file(s) are mock-heavy (5%).
    coverage_diversity.framework_fragmentation  5.3% [strong]
      Evidence: strong
      2 framework(s) across 38 test file(s).
    coverage_diversity.e2e_concentration     57.9% [moderate]
      Evidence: strong
      22 of 38 test file(s) use E2E frameworks (58%).
    coverage_diversity.e2e_only_units        0.0% [unknown]
      Evidence: none
      No coverage data available.
      * Provide labeled coverage artifacts (--coverage unit:path, --coverage e2e:path) for coverage-by-type analysis.
    coverage_diversity.unit_test_coverage    0.0% [unknown]
      Evidence: none
      No coverage data available.
      * Provide labeled coverage artifacts for coverage-by-type analysis.

------------------------------------------------------------

STRUCTURAL RISK
  Posture: STRONG
  No significant migration blockers or deprecated patterns detected.

  Measurements:
    structural_risk.migration_blocker_density  0.0% [strong]
      Evidence: strong
      0 migration blocker(s) across 38 test file(s) (0%).
    structural_risk.deprecated_pattern_share  0.0% [strong]
      Evidence: strong
      0 of 38 test file(s) use deprecated patterns (0%).
    structural_risk.dynamic_generation_share  0.0% [strong]
      Evidence: strong
      0 of 38 test file(s) use dynamic test generation (0%).

------------------------------------------------------------

OPERATIONAL RISK
  Posture: STRONG
  No policy violations or legacy framework usage detected.

  Measurements:
    operational_risk.policy_violation_density  0.0% [strong]
      Evidence: strong
      0 policy violation(s) across 38 test file(s) (0%).
    operational_risk.legacy_framework_share  0.0% [strong]
      Evidence: strong
      0 of 38 test file(s) use legacy frameworks (0%).
    operational_risk.runtime_budget_breach_share  0.0% [strong]
      Evidence: weak
      0 of 38 test file(s) exceed runtime budgets (0%).
      * No runtime artifacts provided; runtime budget detection is heuristic-only.

------------------------------------------------------------

Next steps:
  terrain summary       leadership-ready overview
  terrain metrics       aggregate scorecard
  terrain posture --json   machine-readable posture data
```

## Section Annotations

### Dimension Headers

Each dimension is displayed in uppercase with its resolved posture band and a one-line explanation. The "driving measurements" line identifies which specific measurements determined the dimension's band. For example, coverage depth is WEAK because `coverage_depth.uncovered_exports` scored in the weak band.

### Measurement Rows

Each measurement shows:

- **ID and value**: The measurement identifier and its computed value (as a percentage for ratio-type measurements), with the band in brackets.
- **Evidence strength**: One of strong, partial, weak, or none. This tells you how much to trust the finding. "Strong" means direct observation from code analysis. "Partial" means some data but with known gaps (e.g., heuristic test linkage). "Weak" means limited data, typically because runtime or coverage artifacts were not provided. "None" means no data was available for this measurement.
- **Explanation**: A sentence describing the finding in concrete terms -- counts, percentages, and what was measured.
- **Limitations** (marked with `*`): Stated constraints on the measurement. These appear when evidence is partial or weak, explaining what additional data would improve confidence.

### Reading the Evidence

In this example, the health dimension is STRONG but two of its four measurements have weak evidence (flaky share and slow test share) because no runtime artifacts were provided. The measurements show 0% for both, but the limitations note that this is based on code-level heuristics only. Providing `--runtime` with JUnit XML or Jest JSON would give high-confidence runtime-backed signals.

Coverage depth is WEAK with partial evidence on uncovered exports. The 35.7% ratio is computed from heuristic test linkage, and the limitation explicitly states that "some coverage may exist but not be detected." This is an honest assessment -- the finding is likely directionally correct but the exact number may improve with coverage artifact ingestion.

### Dimensions Without Issues

Structural risk and operational risk are both STRONG with strong evidence across all measurements. When a dimension has no issues, the output is compact -- just the measurements with their zero values. This keeps the report scannable: problem areas get detail, clean areas get confirmation.

### Unknown Bands

Coverage diversity shows two measurements with `[unknown]` bands and `none` evidence. These are coverage-by-type measurements that require labeled coverage artifacts to compute. Rather than guessing or omitting them, Terrain shows them explicitly with a remediation hint. This transparency helps users understand what additional data would unlock.
