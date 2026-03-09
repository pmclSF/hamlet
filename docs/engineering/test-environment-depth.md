# Test Environment Depth Analysis

## Overview

The `envdepth` package assesses the environmental depth — the degree of realism — in a repository's test suite. It classifies each test file into a depth category based on observable evidence: mock counts, assertion counts, and framework type.

This analysis is **descriptive, not judgmental**. Heavy mocking is a valid and often necessary strategy for unit-level isolation. Browser-backed tests are not inherently "better" — they carry different cost, speed, and maintenance tradeoffs. The purpose of depth analysis is to give teams visibility into the shape of their test environment so they can make informed decisions about coverage composition.

## Depth Classes

| Class | Description | Typical Indicators |
|---|---|---|
| `heavy_mocking` | Tests where mock usage significantly outweighs assertions, or absolute mock count is high (>= 8). | `mock_library` |
| `moderate_mocking` | Tests with mocks present but not dominant relative to assertions. | `mock_library` |
| `real_dependency_usage` | Tests using an integration or E2E framework with zero mocks, suggesting real service/dependency interaction. | `real_http`, `in_memory_db` |
| `browser_runtime` | Tests using a browser-backed framework (Cypress, Playwright, Puppeteer, etc.). | `browser_driver` |
| `unknown` | Insufficient data to classify. | — |

## Classification Logic

1. **Browser runtime** is checked first: if the test file's framework is a known browser-backed framework (cypress, playwright, puppeteer, testcafe, webdriverio, selenium), it is classified as `browser_runtime` regardless of mock count.

2. **Heavy mocking** is triggered when `MockCount > 2 * AssertionCount` or `MockCount >= 8`.

3. **Moderate mocking** is triggered when `MockCount > 0` and `MockCount <= AssertionCount`.

4. **Real dependency usage** applies when the framework is an integration or E2E type and mock count is zero.

5. **Unknown** is the fallback when none of the above criteria are met.

## Evidence and Confidence

Each assessment carries a confidence score (0.0–1.0) reflecting how strong the evidence is for the classification:

- Browser runtime: **0.85** — framework name is a strong signal.
- Heavy mocking: **0.80** — count-based heuristics are reasonably reliable.
- Moderate mocking: **0.70** — presence of mocks is observable, but boundary is softer.
- Real dependency: **0.65** — absence of mocks is suggestive but not conclusive.
- Unknown: **0.30** — insufficient evidence.

## Limitations

- **Mock counting is heuristic.** The underlying `MockCount` is derived from regex-based pattern matching, which may miss some mock patterns or over-count mock-related helper calls.
- **Framework name is the primary browser signal.** Custom or renamed browser frameworks may not be recognized.
- **No runtime evidence.** This analysis is purely structural. A test classified as `real_dependency_usage` may still use in-process fakes that are not detectable via static analysis.
- **Depth is not quality.** A heavily-mocked unit test suite may provide excellent fault isolation. A browser-backed E2E suite may be slow and flaky. Depth is one dimension of test intelligence, not a quality score.

## Integration

The `Assess` function accepts a `*models.TestSuiteSnapshot` and returns an `*AssessmentResult` containing per-file assessments, a depth distribution map, and an overall depth classification.

```go
result := envdepth.Assess(snapshot)
for _, a := range result.Assessments {
    fmt.Printf("%s: %s (confidence: %.2f)\n", a.FilePath, a.Depth, a.Confidence)
}
```
