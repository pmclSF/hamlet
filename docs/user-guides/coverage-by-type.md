# Coverage by Type

Terrain can analyze your test coverage by type — showing which functions are covered by unit tests vs. integration tests vs. e2e tests, and identifying risky areas where code is covered only by slow, broad tests.

## Quick Start

```bash
# Analyze with separate coverage runs
terrain analyze --coverage unit:./coverage/unit/lcov.info \
               --coverage e2e:./coverage/e2e/lcov.info \
               --coverage integration:./coverage/integration/lcov.info

# View coverage by type in the posture report
terrain posture

# View detailed coverage insights
terrain analyze --json | jq '.coverageInsights'
```

## How It Works

1. **Ingest** labeled coverage artifacts (LCOV or Istanbul JSON)
2. **Attribute** coverage to code units (functions/methods) per run label
3. **Classify** each code unit by which test types cover it
4. **Surface** actionable insights about coverage gaps

## Coverage Labels

When providing coverage data, label each artifact with the test type that produced it:

| Label | Meaning |
|-------|---------|
| `unit` | Unit test coverage run |
| `integration` | Integration test coverage run |
| `e2e` | End-to-end test coverage run |

If no label is provided, the coverage is marked as `unknown`.

## What You'll See

### Functions Covered Only by E2E

These functions have no fast feedback loop. If an e2e test breaks, you won't know which function is at fault:

```
⚠ 12 code unit(s) are covered only by e2e tests.
  → src/api/checkout.js:processPayment (covered only by e2e)
  → src/api/checkout.js:validateCart (covered only by e2e)
  → Add unit tests for code units that rely exclusively on e2e coverage.
```

### Uncovered Exported Functions

Public API surface with no test coverage at all:

```
🔴 5 exported/public function(s) have no test coverage.
  → src/utils/format.js:formatCurrency (no coverage)
  → src/auth/session.js:refreshToken (no coverage)
  → Prioritize adding tests for public API surface.
```

### Weak Coverage Diversity

Files where most functions are only covered by broad tests:

```
⚠ src/core/engine.js has 8 of 12 code units covered only by e2e tests.
  → Add unit tests to src/core/engine.js to reduce e2e dependency.
```

### Branch Coverage Gaps

Functions with line coverage but no branch coverage:

```
ℹ 15 function(s) have line coverage but no branch coverage.
  → Add tests that exercise conditional branches.
```

## Comparing Over Time

```bash
# Save a baseline snapshot
terrain analyze --coverage unit:./coverage/unit/lcov.info --json > baseline.json

# After improvements, compare
terrain analyze --coverage unit:./coverage/unit/lcov.info --json > current.json
terrain compare baseline.json current.json
```

The comparison shows:
- Line coverage change (e.g., 65% → 72%)
- Change in uncovered exports (e.g., 10 → 7)
- Change in e2e-only coverage (e.g., 5 → 3)

## Interpreting Results

| Finding | Risk Level | Action |
|---------|------------|--------|
| Public function with no coverage | **High** | Add unit tests immediately |
| Function covered only by e2e | **Medium** | Add unit tests for fast feedback |
| File with weak coverage diversity | **Medium** | Improve unit test coverage per file |
| Line coverage without branch coverage | **Low** | Add conditional path tests |
| Partially covered function (<50%) | **Low** | Expand test scenarios |

## Limitations

- Coverage-by-type requires **separate coverage runs** per test type. Most CI setups already produce these.
- Branch attribution is currently file-level, not function-level.
- Coverage precision depends on the quality of coverage artifacts (LCOV/Istanbul).
- Test type labels come from run labels, not per-test attribution (unless per-test coverage is provided).
