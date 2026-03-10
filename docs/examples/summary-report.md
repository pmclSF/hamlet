# Example: `hamlet summary` Output

This document shows representative output from `hamlet summary` run against a mid-sized JavaScript repository with 2 frameworks (Jest + Cypress), 53 test files, and no prior snapshots.

## Terminal Output

```
Hamlet Executive Summary
==================================================

Overall Posture
--------------------------------------------------
  Health:              strong
  Coverage Depth:      weak
  Coverage Diversity:  moderate
  Structural Risk:     strong
  Operational Risk:    strong

Key Numbers
--------------------------------------------------
  Test files:          53
  Frameworks:          2
  Total signals:       18
  High-risk areas:     2

Top Risk Areas
--------------------------------------------------
  src/api/                    weak coverage_depth risk
  src/payments/               moderate health risk

Trend Highlights
--------------------------------------------------
  No prior snapshots available.
  Run `hamlet analyze --write-snapshot` to begin tracking trends.

Dominant Drivers
--------------------------------------------------
  untestedExport (7 signals)
  weakAssertion (5 signals)
  skippedTest (3 signals)

Recommended Focus
--------------------------------------------------
  Address untested exports in src/api/ — 7 exported functions have no linked
  tests. Coverage depth is the weakest dimension and the most actionable.

Prioritized Recommendations
--------------------------------------------------
  1. Add unit tests for untested exports in src/api/
     Why:      7 exported code units have no linked tests, leaving core API
               surface unverified.
     Where:    src/api/users.js, src/api/orders.js, src/api/billing.js
     Evidence: partial
  2. Strengthen assertions in test files with weak density
     Why:      5 test files have fewer than 1 assertion per test on average,
               reducing confidence that behavior is verified.
     Where:    src/payments/__tests__/, src/auth/__tests__/
     Evidence: strong
  3. Review skipped tests for relevance
     Why:      3 test files contain skipped tests that may indicate deferred
               work or stale test logic.
     Where:    src/api/__tests__/orders.test.js, src/auth/__tests__/
     Evidence: strong

Known Blind Spots
--------------------------------------------------
  Runtime health: No runtime artifacts provided. Flaky and slow test detection
    is based on code-level heuristics only.
    → Provide --runtime with JUnit XML or Jest JSON for runtime-backed signals.
  Coverage precision: No coverage data provided. Untested export detection uses
    heuristic test linkage.
    → Provide --coverage with LCOV or Istanbul JSON for precise coverage data.

Benchmark Readiness
--------------------------------------------------
  Ready:
    Health
    Coverage Diversity
    Structural Risk
    Operational Risk
  Limited:
    Coverage Depth (no coverage artifacts provided)

Next steps:
  hamlet posture       evidence behind each dimension
  hamlet analyze       full signal-level detail
  hamlet export benchmark   privacy-safe export
```

## Section Annotations

### Overall Posture

The five posture dimensions, each resolved to a band. This is the top-level answer to "how is our test suite doing?" Coverage depth is weak, meaning exported code units are not adequately tested. The other dimensions are strong or moderate.

### Key Numbers

Aggregate counts that anchor the summary. 53 test files across 2 frameworks, 18 total signals detected, 2 areas classified as high risk. These numbers provide scale context for the findings below.

### Top Risk Areas

The highest-concentration risk hotspots, identified by directory. `src/api/` has weak coverage depth (untested exports). `src/payments/` has moderate health risk (skipped tests or weak assertions concentrated there). These are the areas to investigate first.

### Trend Highlights

Shows changes since the previous snapshot. In this example, no prior snapshots exist, so Hamlet prompts the user to begin tracking. When snapshots are available, this section shows movement like "Coverage depth improved from weak to moderate" or "3 new untested exports detected."

### Dominant Drivers

The most frequently occurring signal types across the repository. This reveals what kinds of issues dominate. Here, untested exports are the primary concern (7 signals), followed by weak assertions (5) and skipped tests (3).

### Recommended Focus

A single-sentence prioritization recommendation derived from the analysis. It identifies the weakest dimension and the most impactful action to improve it.

### Prioritized Recommendations

Structured recommendations with what, why, where, and evidence strength. Each recommendation is traceable to specific signals and locations. Evidence strength indicates how confident the finding is: "strong" means direct observation, "partial" means some data with known gaps.

### Known Blind Spots

Areas where Hamlet lacks data to make confident assessments. This section is critical for honesty -- it tells the user what the tool cannot see and how to provide additional data. In this example, runtime and coverage artifacts are missing, which limits flaky test detection and coverage precision.

### Benchmark Readiness

Which posture dimensions have enough data for meaningful cross-repository comparison via `hamlet export benchmark`. Dimensions with limited data are flagged with the reason, so users know what to provide to improve benchmark quality.
