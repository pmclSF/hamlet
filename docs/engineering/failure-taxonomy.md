# Failure Taxonomy and Runtime Failure Classification

**Package:** `internal/failure/`
**Stage:** 109

## Overview

The failure taxonomy classifies observed test failures into meaningful
categories based on error messages and stack traces. This enables downstream
consumers (health dashboards, migration readiness, stability analysis) to
understand _why_ tests fail, not just _that_ they fail.

## Failure Categories

| Category | Constant | Description |
|----------|----------|-------------|
| Assertion Failure | `assertion_failure` | Test expectation mismatch (expect, assert, should) |
| Timeout | `timeout` | Operation exceeded time limit (ETIMEDOUT, async callback) |
| Setup/Fixture | `setup_or_fixture_failure` | Failure in beforeEach, beforeAll, setUp, fixture |
| Dependency/Service | `dependency_or_service_failure` | External service unavailable (ECONNREFUSED, 503) |
| Snapshot Mismatch | `snapshot_mismatch` | Snapshot assertion divergence |
| Selector/UI | `selector_or_ui_fragility` | DOM element not found, stale reference, detached |
| Infrastructure | `infrastructure_or_environment` | OOM, disk space, permission denied |
| Unknown | `unknown` | No pattern matched |

## Classification Confidence

Each classification carries a confidence level:

- **exact** — Unambiguous keyword match (e.g., `ECONNREFUSED` -> dependency)
- **inferred** — Pattern-based match requiring interpretation (e.g., `expect` -> assertion)
- **weak** — Low-signal guess or no match

Numeric confidence scores range from 0.0 to 1.0 and provide finer-grained
ranking within each confidence band.

## Priority Resolution

When an error message matches multiple categories, the classifier uses a
priority system. More specific categories take precedence:

1. Snapshot mismatch (highest priority)
2. Selector/UI fragility
3. Infrastructure/environment
4. Dependency/service
5. Timeout
6. Setup/fixture
7. Assertion failure (lowest priority)

This prevents generic assertion keywords from masking more specific signals.
For example, `expect(received).toMatchSnapshot()` matches both snapshot and
assertion patterns but is classified as snapshot mismatch.

## Input Boundary

The classifier accepts `[]FailureInput` values, which are decoupled from any
specific runtime artifact format. This allows ingestion from JUnit XML, Jest
JSON, or any future format without coupling the classification logic to
parsers.

```go
type FailureInput struct {
    TestFilePath string
    TestName     string
    ErrorMessage string
    StackTrace   string
}
```

Both `ErrorMessage` and `StackTrace` are searched for pattern matches,
allowing classification even when only a stack trace is available.

## Output

The `TaxonomyResult` provides:

- Individual classifications with category, confidence, and explanation
- Aggregate counts by category (`ByCategory`)
- Total failure count
- Dominant category (most common failure type)
- Deterministic sort order (by file path, then test name)

## Usage

```go
inputs := []failure.FailureInput{
    {
        TestFilePath: "test/api.test.js",
        TestName:     "fetches user",
        ErrorMessage: "connect ECONNREFUSED 127.0.0.1:5432",
    },
}

result := failure.Classify(inputs)
// result.Classifications[0].Category == CategoryDependencyService
// result.DominantCategory == CategoryDependencyService
```

## Design Decisions

1. **Pattern-based, not ML.** Classification uses substring matching on
   lowercased text. This is deterministic, fast, and auditable.

2. **Decoupled from runtime parsers.** The `FailureInput` struct avoids
   importing `internal/runtime`, keeping the package independently testable.

3. **Priority over scoring.** When multiple categories match, the most
   specific category wins regardless of individual pattern scores. This avoids
   ambiguity in multi-signal error messages.

4. **Deterministic output.** Classifications are sorted by file path and test
   name to ensure stable output across runs.
