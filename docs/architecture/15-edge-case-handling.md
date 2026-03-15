# Edge Case Handling

> **Status:** Implemented
> **Purpose:** Detect repository conditions that reduce analysis reliability and adapt behavior via a fallback strategy ladder.
> **Key decisions:**
> - 14 named edge-case conditions profiled across 8 repository dimensions before engines run
> - Conservative under uncertainty: edge cases widen the recommended test set rather than narrowing it
> - Five-level fallback ladder from direct-dependency tests up to full suite, selected by the most severe active edge case
> - All confidence adjustments are multiplicative and visible in the profile report — no hidden heuristics

**See also:** [12-risk-and-coverage-taxonomy.md](12-risk-and-coverage-taxonomy.md), [14-evidence-scoring-and-confidence-model.md](14-evidence-scoring-and-confidence-model.md), [06-monorepo-graph-scaling.md](06-monorepo-graph-scaling.md)

Terrain adapts its behavior based on the characteristics of each repository. Not every repository benefits from the same analysis approach. A 10-test hobby project needs different treatment than a 5000-test monorepo with three frameworks and a QA team running manual regression suites.

## Repository Profiling

Before running insight engines, Terrain profiles the repository across eight dimensions:

| Dimension | Values | How It's Measured |
|-----------|--------|-------------------|
| Test volume | tiny / small / medium / large | Count of test nodes (≤10 / ≤100 / ≤1000 / >1000) |
| CI pressure | low / medium / high | CI duration if known (≤60s / ≤300s / >300s), or estimated from test count |
| Coverage confidence | low / medium / high | Low-band ratio and graph density |
| Redundancy level | low / medium / high | Duplicate ratio from duplicate engine (>30% high, >10% medium) |
| Flake burden | low / medium / high | Flaky test ratio (>10% high, >2% medium) |
| Skip burden | low / medium / high | Skipped test ratio (>20% high, >5% medium) |
| Fanout burden | low / medium / high | Flagged node ratio from fanout engine (>15% high, >5% medium) |
| Manual coverage presence | none / partial / substantial | Count and criticality of manual coverage entries |

The profile is computed from graph structure, engine outputs, and optional metadata from `terrain.yaml` (CI duration, manual coverage entries).

## Edge Cases

Terrain detects 14 edge-case conditions that affect analysis reliability:

### FEW_TESTS

**Condition:** Test volume is "tiny" (≤10 tests).

**Behavior:** CI optimization is disabled. With so few tests, running all of them is faster than computing which ones to skip. Terrain reports the full test suite as the recommendation.

### FAST_CI_ALREADY

**Condition:** CI pressure is "low" (≤60 seconds or ≤50 tests with no duration data).

**Behavior:** Terrain recommends focusing on duplicate reduction and coverage gaps rather than test selection optimization. Optimization adds complexity with minimal time savings.

### REDUNDANT_TEST_SUITE

**Condition:** Redundancy level is "high" (>30% of tests are in duplicate clusters).

**Behavior:** Terrain elevates duplicate reduction as the primary recommendation. Fallback behavior widens to fixture expansion to ensure deduplicated tests still cover the necessary code.

### HIGH_SKIP_BURDEN

**Condition:** Skip burden is "high" (>20% of tests are skipped).

**Behavior:** Coverage confidence is reduced by 0.8x. Skipped tests represent invisible gaps — the denominator looks smaller than it really is. Terrain recommends addressing skipped tests before trusting coverage metrics.

### HIGH_FLAKE_BURDEN

**Condition:** Flake burden is "high" (>10% of tests are flaky).

**Behavior:** CI optimization is disabled. Flaky tests undermine the reliability of test selection — a "passed" result from a flaky test is not evidence of correctness. Terrain recommends stabilizing flaky tests first.

### HIGH_FANOUT_FIXTURE

**Condition:** Fanout burden is "high" (>15% of nodes exceed the fanout threshold).

**Behavior:** Coverage confidence is reduced by 0.7x. High-fanout fixtures make impact analysis noisy — every change appears to affect many tests. Fallback behavior widens to package-level tests.

### LOW_GRAPH_VISIBILITY

**Condition:** Coverage confidence is "low" AND the edge-to-test ratio is less than 2 (very few edges per test).

**Behavior:** Coverage confidence is reduced by 0.6x. A sparse graph means many dependency relationships are not captured. Impact analysis may miss affected tests. The fallback ladder escalates to full suite in combination with FEW_TESTS.

### EXTERNAL_SERVICE_HEAVY

**Condition:** More than 5 external service nodes in the graph (from graph `ExternalService` nodes plus snapshot-level counts).

**Behavior:** Fallback widens to fixture expansion. Coverage confidence is reduced by 0.85x. External service dependencies are opaque to static analysis — a service change can break tests without any file change being visible in the repository.

### GENERATED_ARTIFACT_CHANGES

**Condition:** Any `GeneratedArtifact` nodes present in the graph or snapshot data.

**Behavior:** Terrain recommends excluding generated files from impact scope to reduce noise. No fallback escalation — the edge case is informational.

### MIGRATION_OVERLAP

**Condition:** More than 10 migration signals detected in the snapshot (framework migration markers, dual-framework patterns).

**Behavior:** Fallback widens to package tests. Confidence is reduced by 0.8x. During a framework migration, duplicate coverage across old and new frameworks creates uncertainty. Terrain recommends completing migration before optimizing test selection.

### SNAPSHOT_HEAVY_SUITE

**Condition:** More than 40% of assertions are snapshot-based AND more than 20 snapshot assertions exist.

**Behavior:** Confidence is reduced by 0.9x. Snapshot-heavy suites inflate assertion counts without proportional confidence. Terrain recommends reviewing snapshot tests for value vs. churn cost.

### LEGACY_ZONE

**Condition:** More than 5 legacy framework signals detected in the snapshot.

**Behavior:** Terrain recommends migrating legacy tests before optimizing. Legacy test zones may not benefit from optimization due to outdated patterns and tooling.

### MIXED_TEST_CULTURES

**Condition:** 4+ frameworks detected, OR 3+ frameworks spanning 3+ distinct categories (unit, integration, e2e, etc.).

**Behavior:** Fallback widens to fixture expansion. Confidence is reduced by 0.85x. Mixed test cultures reduce cross-framework optimization confidence. Terrain recommends standardizing on fewer frameworks.

### LARGE_MANUAL_SUITE

**Condition:** Manual coverage presence is "significant".

**Behavior:** Terrain recommends factoring in manual QA when assessing risk. Automated analysis may underestimate total protection when significant manual coverage exists.

## Fallback Strategy Ladder

When edge cases reduce confidence, Terrain widens the set of tests it recommends running. The fallback ladder has five levels:

| Level | Name | What It Includes |
|-------|------|------------------|
| 1 | Direct dependency tests | Only tests that directly import the changed file |
| 2 | Fixture/helper expansion | Add tests that share fixtures or helpers with directly impacted tests |
| 3 | Package tests | Add all tests in the affected package(s) |
| 4 | Smoke/regression bundles | Add smoke tests and regression suites |
| 5 | Full test suite | Run everything |

### Level Selection

The fallback level is determined by the combination of detected edge cases:

- **No edge cases** → Level 1 (direct dependencies)
- **EXTERNAL_SERVICE_HEAVY, HIGH_FANOUT_FIXTURE, MIXED_TEST_CULTURES** → Level 2 (fixture expansion)
- **REDUNDANT_TEST_SUITE, HIGH_SKIP_BURDEN, HIGH_FLAKE_BURDEN, MIGRATION_OVERLAP** → Level 3 (package tests)
- **LOW_GRAPH_VISIBILITY** → Level 4 (smoke/regression)
- **FEW_TESTS** → Level 5 (full suite)

When multiple edge cases apply, the highest (most conservative) level wins.

## Policy Output

The edge case handler produces a `Policy` containing:

- **Edge cases** — list of detected conditions with reasons
- **Recommendations** — human-readable suggestions
- **Fallback level** — which level of the fallback ladder to use
- **Optimization disabled** — whether to skip CI optimization entirely
- **Confidence adjustment** — multiplicative factor to apply to confidence scores
- **Risk elevation** — whether to elevate risk assessment

This policy is wired into three commands:

- **`terrain analyze`** — edge cases and policy recommendations appear in the report alongside the repository profile
- **`terrain impact`** — edge case policy adjusts coverage confidence and adds limitations when confidence is low or risk is elevated
- **`terrain insights`** — edge cases and policy recommendations appear in the health report with severity badges

## Configuration

Edge case behavior is influenced by `terrain.yaml`:

```yaml
# Manual coverage entries affect LARGE_MANUAL_TEST_SUITE detection
manual_coverage:
  - name: billing regression suite
    area: billing-core
    source: testrail
    owner: qa-billing
    criticality: high

# CI duration affects FAST_CI_ALREADY detection
ci_duration_seconds: 45
```

## Design Philosophy

The edge case system follows Terrain's core principle: **conservative under uncertainty**. When the repository exhibits conditions that reduce analysis reliability, Terrain runs more tests rather than fewer, elevates risk rather than hiding it, and recommends addressing root causes rather than working around them.

Every adjustment is visible in the profile report. There are no hidden heuristics.
