# Quarantine and Suppression Awareness (Stage 108)

## Overview

The `internal/suppression` package detects and models quarantined tests,
expected failures, skip policies, and retry wrappers within a Hamlet snapshot.
Its purpose is to surface hidden test debt that silently degrades suite
reliability.

## Suppression Model

A **Suppression** represents a detected suppression mechanism on a test or test
file. Each suppression carries:

| Field          | Description                                       |
|----------------|---------------------------------------------------|
| `TestFilePath` | Repository-relative path of the suppressed test   |
| `TestName`     | Specific test name when identifiable               |
| `Kind`         | Suppression mechanism type (see below)             |
| `Intent`       | Tactical vs chronic classification                 |
| `Source`       | How the suppression was detected                   |
| `Confidence`   | 0.0-1.0 detection confidence score                |
| `Explanation`  | Human-readable description                         |
| `Metadata`     | Additional context (rates, thresholds, etc.)       |

### Suppression Kinds

| Kind               | Meaning                                                    |
|--------------------|------------------------------------------------------------|
| `quarantined`      | Test explicitly placed in quarantine (folder or tag)       |
| `expected_failure`  | Test with persistently low pass rate accepted as-is        |
| `skip_disable`     | Test skipped or disabled via annotation, config, or naming |
| `retry_wrapper`    | Test masked by retry-as-policy pattern                     |

### Intent Classification

| Intent    | Meaning                                            |
|-----------|----------------------------------------------------|
| `tactical`| Recent, likely temporary suppression               |
| `chronic` | Persistent suppression, likely accumulated debt    |
| `unknown` | Insufficient evidence to classify                  |

## Detection Strategies

### 1. Signal-Based Detection

Scans existing Hamlet signals for `skippedTest` signal types. These are
high-confidence indicators (0.9) since they come from prior analysis.

### 2. Runtime Data Detection

Examines `RuntimeStats` on test files:

- **Retry wrapper**: `RetryRate >= 0.3` suggests retry-as-policy. Confidence
  0.7. Tests with `RetryRate >= 0.5` are classified as chronic.
- **Expected failure**: `0 < PassRate < 0.3` suggests the team has accepted
  persistent failure. Confidence 0.5. Always classified as chronic.

### 3. Naming Convention Detection

Matches test file paths against known suppression indicators:

- `quarantine`, `quarantined` -- mapped to `KindQuarantined`
- `.skip`, `skip.`, `disabled`, `.disabled`, `xdescribe`, `xit`, `pending`,
  `.pending` -- mapped to `KindSkipDisable`

Confidence 0.6. One match per file.

## Deduplication

Suppressions are deduplicated by `Kind + TestFilePath`. If multiple strategies
detect the same kind of suppression on the same file, only the first detected
instance is retained.

## Result Structure

`SuppressionResult` aggregates all detected suppressions with pre-computed
counts by kind and intent, plus `TotalSuppressedTests` (unique file count).

## Integration Points

The suppression package is designed to be called after snapshot construction:

```go
import "github.com/pmclSF/hamlet/internal/suppression"

result := suppression.Detect(snapshot)
```

Results can feed into:

- **Risk surfaces**: Chronic suppressions contribute to reliability risk.
- **Quality signals**: High suppression counts indicate test debt.
- **Migration readiness**: Quarantined tests may block migration confidence.
- **Governance policies**: Teams can set thresholds for maximum allowed
  suppressions.

## Future Extensions

- **Annotation-based detection**: Parse framework-specific annotations
  (`@Quarantined`, `@Disabled`, `pytest.mark.skip`).
- **Config-based detection**: Scan CI configuration files for quarantine lists.
- **Temporal analysis**: Compare suppressions across snapshots to detect newly
  chronic patterns.
- **Tactical intent inference**: Use git blame / commit age to distinguish
  recent tactical suppressions from old chronic ones.
