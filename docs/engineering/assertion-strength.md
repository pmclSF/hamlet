# Assertion Strength and Oracle-Quality Analysis

## Overview

The `internal/assertion/` package assesses whether tests check behavior meaningfully by analyzing assertion density, category distribution, and mock/snapshot ratios. This is a static-analysis-based heuristic that operates on snapshot data — it does not parse test source code directly.

## Strength Classes

| Class | Meaning |
|-------|---------|
| `strong` | High assertion density with targeted behavioral checks and low mock ratio |
| `moderate` | Acceptable density; may rely on implicit checks (common in E2E) |
| `weak` | Low density, mock-heavy, snapshot-dominated, or no assertions |
| `unclear` | Insufficient data to classify (e.g., no tests detected in file) |

## Classification Logic

### Density Calculation

Assertion density is computed as `assertionCount / testCount` for each file. This is the primary signal for strength classification.

### Thresholds (Unit/Integration Frameworks)

- **Strong**: density >= 3.0 and mock ratio < 0.5
- **Moderate**: density >= 1.5
- **Weak**: density < 1.0, or mock count exceeds assertion count, or snapshot ratio >= 0.8 with low density
- **Unclear**: no tests detected

### E2E Framework Adjustments

E2E frameworks (Cypress, Playwright, Puppeteer, Selenium, WebDriverIO, TestCafe) receive adjusted thresholds because E2E tests often validate behavior through navigation and implicit checks rather than explicit assertions:

- **Strong**: density >= 2.0 and mock ratio < 0.3
- **Moderate**: density > 0 (even low density may indicate implicit validation)
- **Weak**: zero assertions

### Category Inference

Categories are inferred from available snapshot fields:

| Field | Inferred Category |
|-------|-------------------|
| `SnapshotCount > 0` | `snapshot` |
| Remaining assertions | `behavioral` (default) |

Future enhancements may add deeper category inference from AST analysis.

### Mock-Heavy Detection

When `MockCount > AssertionCount`, the test is classified as weak regardless of density. High mock counts relative to assertions suggest the test is verifying interaction patterns rather than behavioral outcomes.

## Overall Strength

The aggregate strength across all files uses a majority-based heuristic:

- **Strong**: >= 50% of assessed files are strong and < 20% are weak
- **Weak**: >= 50% of assessed files are weak
- **Moderate**: everything else
- **Unclear**: all files are unclear

## Data Sources

The assessment consumes `TestFile` fields from `models.TestSuiteSnapshot`:

- `TestCount` — estimated test count
- `AssertionCount` — estimated assertion count
- `MockCount` — estimated mock constructs
- `SnapshotCount` — snapshot assertion count
- `Framework` — framework name (used for E2E detection)

Framework metadata from `snap.Frameworks` is also used to determine framework type.

## Limitations

- Classification is based on counts, not semantic analysis of assertion targets
- Cannot distinguish `toBe(200)` (status check) from `toBe(expectedValue)` (behavioral)
- Snapshot count may not perfectly reflect snapshot-only tests
- Mock count heuristics may over-penalize legitimate test doubles
- Confidence values are calibrated conservatively
