# Test Type Inference

Hamlet classifies each test case by type with explicit evidence and confidence, enabling coverage-by-type analysis.

## Supported Types

| Type | Description |
|------|-------------|
| `unit` | Fast, isolated, tests a single unit |
| `integration` | Tests interaction between components, may use DB/network |
| `e2e` | End-to-end, drives a browser or full system |
| `component` | Tests a UI component in isolation |
| `smoke` | Quick validation of basic functionality |
| `unknown` | Insufficient evidence for classification |

## Architecture

```
internal/testtype/infer.go    Inference engine with modular rules
```

`InferForTestCase()` applies multiple inference rules and returns the best result with accumulated evidence. `InferAll()` applies inference to an entire test case slice.

## Inference Rules

Rules are applied in priority order. The highest-confidence candidate wins:

### 1. Framework-based (strongest for e2e)

| Framework | Inferred Type | Confidence |
|-----------|---------------|------------|
| playwright, cypress, puppeteer, webdriverio, testcafe | e2e | 0.9 |
| jest, vitest, mocha, jasmine | unit | 0.5 |
| go-testing | unit | 0.5 |
| pytest, unittest, nose2, junit4, junit5, testng | unit | 0.5 |

### 2. Path-based

| Pattern | Inferred Type | Confidence |
|---------|---------------|------------|
| `/e2e/`, `/end-to-end/` directory | e2e | 0.85 |
| `/integration/`, `/integ/` directory | integration | 0.85 |
| `/smoke/` directory | smoke | 0.85 |
| `/unit/`, `/__tests__/` directory | unit | 0.75 |
| `.e2e.` in filename | e2e | 0.8 |
| `.integration.` in filename | integration | 0.8 |
| `.cy.js`/`.cy.ts` extension | e2e | 0.9 |

### 3. Suite hierarchy naming

| Pattern in suite name | Inferred Type | Confidence |
|-----------------------|---------------|------------|
| contains "integration" | integration | 0.7 |
| contains "e2e" or "end to end" | e2e | 0.7 |
| contains "component" | component | 0.6 |
| contains "smoke" | smoke | 0.7 |

### 4. Test name patterns

| Pattern | Inferred Type | Confidence |
|---------|---------------|------------|
| starts with "e2e" or "end-to-end" | e2e | 0.6 |
| starts with "integration" or "integ" | integration | 0.6 |
| starts with "smoke" | smoke | 0.6 |

## Conflict Handling

When multiple rules disagree (e.g., framework says "unit" but path says "integration"), the highest-confidence candidate wins and confidence is reduced by 20%. Evidence includes `"conflicting signals reduced confidence"`.

## Output

Each test case gets three fields populated:
- `TestType` — the inferred type string
- `TestTypeConfidence` — 0.0–1.0 confidence score
- `TestTypeEvidence` — list of human-readable reasons

## Example

A Jest test at `test/integration/db.test.js` in suite "Integration Tests > API":

```json
{
  "testType": "integration",
  "testTypeConfidence": 0.68,
  "testTypeEvidence": [
    "framework jest is typically used for unit tests",
    "path contains integration directory: integration",
    "suite name contains 'integration': Integration Tests",
    "conflicting signals reduced confidence"
  ]
}
```

## Limitations

- Test type inference is heuristic-based, not provably correct
- `unknown` is an honest answer when evidence is insufficient
- Framework-only inference for unit test frameworks has low confidence (0.5)
- Content-based signals (imports, fixtures, browser/DB usage) are not yet implemented
