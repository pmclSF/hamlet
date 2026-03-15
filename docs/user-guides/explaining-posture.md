# User Guide: Understanding and Explaining Posture

## What Is Posture?

When you run `terrain summary` or `terrain posture`, Terrain reports the state of your test suite across five dimensions:

| Dimension | Question It Answers |
|-----------|-------------------|
| Health | Are the tests themselves reliable? |
| Coverage Depth | Is exported code actually tested? |
| Coverage Diversity | Is coverage structurally diversified? |
| Structural Risk | Can the suite be modernized? |
| Operational Risk | Are operational controls in place? |

Each dimension gets a **band**: strong, moderate, weak, elevated, or critical.

## Viewing Posture

### Quick View

```bash
terrain summary
```

Shows posture as part of the executive summary:

```
Posture Dimensions
--------------------------------------------------
  health:                    STRONG
  coverage_depth:            WEAK
    15 of 40 exported code unit(s) appear untested (38%).
  coverage_diversity:        MODERATE
  structural_risk:           STRONG
  operational_risk:          STRONG
```

### Detailed View

```bash
terrain posture
```

Shows full evidence for every measurement:

```
COVERAGE_DEPTH
  Posture: WEAK
  coverage_depth posture is weak. Driven by: coverage_depth.uncovered_exports.

  Driving measurements: coverage_depth.uncovered_exports
  Measurements:
    coverage_depth.uncovered_exports         37.5% [weak]
      Evidence: partial
      15 of 40 exported code unit(s) appear untested (38%).
      * Test linkage is heuristic-based; some coverage may exist but not be detected.
    coverage_depth.weak_assertion_share      5.0% [strong]
      Evidence: strong
      4 of 80 test file(s) have weak assertion density (5%).
```

### Machine-Readable

```bash
terrain posture --json
```

Returns the full `MeasurementSnapshot` as JSON.

## Understanding the Evidence

Every measurement includes an **evidence** level:

| Evidence | What It Means |
|----------|---------------|
| Strong | Based on direct observation (e.g., signal counts, runtime data) |
| Partial | Some data available but gaps noted |
| Weak | Limited data; the result may improve with more artifacts |
| None | No data available for this measurement |

If you see `evidence: weak` with a limitation like "No runtime data available", you can improve the result by providing runtime artifacts:

```bash
terrain analyze --runtime path/to/junit.xml
```

## Tracking Changes Over Time

Save snapshots:
```bash
terrain analyze --write-snapshot
```

Compare them:
```bash
terrain compare
```

The comparison now includes posture and measurement changes:

```
Posture Changes
----------------------------------------
  health                     STRONG → WEAK

Measurement Changes
----------------------------------------
  health.flaky_share                     +23.0%
    band: strong → weak
```

## What Drives Each Dimension?

### Health — "Are tests reliable?"
- **Flaky test share** — tests that pass/fail unpredictably
- **Skip density** — tests marked as skipped
- **Dead test share** — unreachable test code
- **Slow test share** — tests exceeding performance thresholds

### Coverage Depth — "Is exported code tested?"
- **Uncovered exports** — exported functions without test linkage
- **Weak assertion share** — test files with few assertions
- **Coverage breach share** — areas below coverage thresholds

### Coverage Diversity — "Right mix of test types?"
- **Mock-heavy share** — tests dominated by mocks
- **Framework fragmentation** — many frameworks in a small suite
- **E2E concentration** — share of tests using E2E frameworks
- **E2E-only units** — code covered only by broad tests
- **Unit test coverage** — code covered by fast unit tests

### Structural Risk — "Can we modernize?"
- **Migration blocker density** — obstacles to framework migration
- **Deprecated pattern share** — patterns that will break in future versions
- **Dynamic generation share** — runtime-generated tests

### Operational Risk — "Are we following our rules?"
- **Policy violation density** — violations of team policy
- **Legacy framework share** — outdated framework usage
- **Runtime budget breach share** — tests exceeding time budgets

## Improving a Weak Posture

1. **Look at the driving measurements** — these are the specific measurements pulling the band down
2. **Read the explanation** — it tells you what was measured and the result
3. **Use `terrain analyze`** to find specific instances of the underlying signals
4. **Fix the most concentrated issues first** — the posture report tells you which measurements have the most impact
5. **Re-run and compare** — save a snapshot before and after to verify improvement

## Example: Fixing Weak Health

```
HEALTH
  Posture: WEAK
  Driven by: health.flaky_share

  health.flaky_share    25.0% [weak]
    5 of 20 test files flagged as flaky or unstable
```

Action plan:
1. Run `terrain analyze --json` and filter for `flakyTest` signals
2. Identify the 5 affected test files
3. Fix flakiness (add retries, fix timing issues, mock external deps)
4. Re-run `terrain posture` to verify improvement
