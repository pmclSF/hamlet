# Test Stability Classes

## Overview

The `internal/stability/` package classifies each test's historical stability pattern by analyzing longitudinal observations across multiple snapshots. This enables Hamlet to distinguish between tests that are reliably stable, recently broken, chronically flaky, or improving over time.

## The 7 Stability Classes

| Class | Value | Description |
|-------|-------|-------------|
| Consistently Stable | `consistently_stable` | Test has been passing reliably across all observed snapshots with low failure and flake rates (both <= 10%). |
| Newly Unstable | `newly_unstable` | Test was stable in the first ~60% of its history but has started failing in the most recent ~40%. Indicates a recent regression. |
| Chronically Flaky | `chronically_flaky` | Flaky test signals are present in >= 40% of observations. The test has a persistent reliability problem. |
| Intermittently Slow | `intermittently_slow` | Slow test signals appear in 30-99% of observations. The test is not always slow, suggesting environment-sensitive performance. |
| Improving | `improving` | The test had failures or flaky signals historically, but recent observations show measurable improvement (problem rate dropped by > 0.2 between first and second half). |
| Quarantined/Suppressed | `quarantined_or_suppressed` | The test is skipped in >= 50% of observations. Likely quarantined, disabled, or conditionally excluded. |
| Data Insufficient | `data_insufficient` | Fewer than 3 observations are available, or signals are too mixed to assign a clear classification. |

## MinHistoryDepth Requirement

Classification requires at least **3 present observations** (`MinHistoryDepth = 3`). Tests with fewer observations are automatically classified as `data_insufficient` with low confidence (0.3). This prevents premature classification based on sparse data.

## Classification Priority Order

When a test matches multiple criteria, the classifier applies rules in priority order. The first match wins:

1. **Data Insufficient** -- fewer than 3 present observations
2. **Quarantined/Suppressed** -- skip rate >= 50%
3. **Improving** -- trend is "improving" AND the test has had failures or flaky signals
4. **Chronically Flaky** -- flaky signal rate >= 40%
5. **Newly Unstable** -- early 60% stable (fail rate <= 10%), late 40% failing (fail rate >= 30%)
6. **Intermittently Slow** -- slow signal rate between 30% and 100%
7. **Consistently Stable** -- failure rate <= 10% AND flaky rate <= 10%
8. **Data Insufficient (fallback)** -- mixed signals that don't match any clear pattern

## Trend Detection

Trend is computed by splitting present observations into two halves and comparing "problem rates" (failure rate + flaky signal rate):

- **Improving**: late-half problems are more than 0.2 lower than early-half problems
- **Worsening**: late-half problems are more than 0.2 higher than early-half problems
- **Stable**: difference is within +/- 0.2
- **Insufficient**: fewer than 3 observations available

This simple bisection approach is robust against short histories and avoids overfitting to individual data points.

## Confidence Scoring

Each classification includes a confidence score between 0.0 and 1.0:

- `data_insufficient`: 0.3 (few observations) or 0.4 (mixed signals)
- `quarantined_or_suppressed`: 0.7 + 0.3 * skip_rate (scales with how consistently skipped)
- `improving`: 0.6 (fixed -- trend detection is inherently less certain)
- `chronically_flaky`: 0.6 + 0.3 * flaky_rate (scales with flake prevalence)
- `newly_unstable`: 0.7 (fixed)
- `intermittently_slow`: 0.6 + 0.2 * slow_rate
- `consistently_stable`: 0.7 + 0.3 * (1 - fail_rate) (near 1.0 for zero failures)

## Building Histories from Snapshots

`BuildHistories()` constructs `TestHistory` records from an ordered sequence of `TestSuiteSnapshot` values (oldest first). For each unique test ID across all snapshots, it:

1. Creates one observation per snapshot (present or absent)
2. Attaches runtime data (pass/fail status derived from `PassRate` thresholds: >= 0.95 = passed, < 0.5 = failed)
3. Attaches signal data by matching file paths (flakyTest, slowTest, skippedTest signals)
4. Carries forward metadata (test name, file path, framework, owner) from the first snapshot where the test appears

## Integration Points

The stability classification system is designed to feed into several downstream consumers:

- **Health measurements**: Stability class distribution contributes to overall test suite health scores
- **Focus recommendations**: Newly unstable and chronically flaky tests are high-priority remediation targets
- **Trend reporting**: The improving/worsening trend feeds into longitudinal trend analysis
- **Portfolio intelligence**: Stability classes aggregate at the owner/team level for portfolio health views

## Known Limitations

- **File-level signal granularity**: Signals (flaky, slow, skipped) are matched by file path, not by individual test case. If a file contains multiple tests, all tests in that file share the same signal observations.
- **No per-test-case runtime**: Runtime stats come from `TestFile.RuntimeStats`, which is file-level. Individual test case durations are not yet available.
- **Binary pass/fail from PassRate**: Pass/fail is derived from PassRate thresholds (>= 0.95 and < 0.5), leaving a gray zone between 0.5 and 0.95 where neither Passed nor Failed is set.
- **No weighted recency**: All observations within each half are weighted equally. A failure 2 snapshots ago counts the same as one 10 snapshots ago within the same half.
- **Map iteration order**: `BuildHistories` iterates over a map of test IDs, so the output order of histories is non-deterministic. The `Classify` function sorts results by TestID for deterministic output.
