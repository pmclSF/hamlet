# Adversarial and Degradation Testing

Adversarial tests verify that Terrain degrades gracefully when given incomplete,
malformed, or pathological input. The philosophy: real repositories are messy,
and trust comes from honest degradation rather than silent failure.

## Philosophy

A test intelligence platform that panics on missing data or produces confident
scores from garbage input is worse than useless. Adversarial tests encode the
principle that Terrain should always:

- **Produce output.** Never panic, never return nil where a caller expects a value.
- **Reduce confidence.** When data is missing or suspect, scores and posture bands
  should reflect reduced confidence rather than fabricating certainty.
- **Annotate limitations.** When output quality is degraded, the result should
  carry explanations or warnings that surface the limitation to the user.

## Adversarial Test Inventory

All adversarial tests live in `internal/testdata/adversarial_test.go`.

| Test Function | Input Condition | Expected Behavior |
|---|---|---|
| `TestAdversarial_NilMeasurements` | MinimalSnapshot with nil measurements | Summary report renders non-empty output |
| `TestAdversarial_EmptySignals` | Snapshot with empty signal slice | Risk scoring completes without panic |
| `TestAdversarial_ZeroTestFiles` | Snapshot with no test files | Metrics returns zero counts |
| `TestAdversarial_MeasurementsOnEmpty` | Completely empty snapshot | Measurement registry returns 5 posture dimensions |
| `TestAdversarial_ImpactEmptyScope` | HealthyBalancedSnapshot with zero changed files | Impact returns non-nil result with zero impacted units |
| `TestAdversarial_ImpactNonexistentFile` | MinimalSnapshot with a changed file not in snapshot | Impact returns file-level unit with weak confidence |
| `TestAdversarial_HeatmapNoRisk` | MinimalSnapshot with nil risk data | Heatmap builds without panic |
| `TestAdversarial_LargeSignalVolume` | Snapshot with 1000 identical signals | Risk scoring completes without panic or hang |
| `TestAdversarial_FilterByOwner_NoMatch` | Impact result filtered by nonexistent team | Filter returns empty unit list |

### Planned Adversarial Tests

| Test Function | Input Condition | Expected Behavior |
|---|---|---|
| `TestAdversarial_MalformedCoverage` | Snapshot with coverage percentages > 100% or negative | Metrics clamp to valid range, add warning |
| `TestAdversarial_ParserFailures` | Test files with syntax errors in path patterns | Framework detection returns unknown, does not panic |
| `TestAdversarial_DuplicateTests` | Snapshot with duplicate test file entries | Metrics deduplicates, counts correctly |
| `TestAdversarial_UnsupportedLanguages` | Snapshot with language "fortran" or "cobol" | Detection returns empty framework list, no panic |
| `TestAdversarial_PartialOwnership` | Snapshot where 80% of code units have no owner | Governance signals reflect orphan ratio honestly |
| `TestAdversarial_SparseHistory` | Snapshot with runtime stats on only 1 of 50 test files | Stability metrics report low confidence |

## Expected Behavior Patterns

### Graceful Output

Every subsystem must return a valid, non-nil result for any input. The caller
should never need to nil-check a return value from `metrics.Derive`,
`scoring.ComputeRisk`, `heatmap.Build`, or `impact.Analyze`. Empty input
produces empty-but-valid output.

### Reduced Confidence

When input data is sparse or missing, confidence indicators should reflect that:

- Impact analysis uses `ConfidenceWeak` for files not found in the snapshot.
- Measurement posture bands should trend toward "unknown" or "insufficient_data"
  rather than defaulting to "strong."
- Risk scores computed from zero signals should be zero or minimal, not
  artificially inflated.

### Limitation Annotations

Where possible, degraded results should carry machine-readable annotations:

- Coverage insights with `Severity: "medium"` for partial data.
- Signal explanations that mention missing inputs.
- Posture explanations that cite insufficient evidence.

These annotations flow through to CLI output and export JSON, giving users
visibility into why a result may be less reliable.

## Writing New Adversarial Tests

When adding a new adversarial test:

1. **Start from a known fixture.** Use an existing factory and then null out,
   empty, or corrupt a specific field.
2. **Assert non-panic first.** The minimum bar is "does not panic."
3. **Assert graceful output second.** Verify the result is non-nil and structurally
   valid.
4. **Assert honest degradation third.** Verify that confidence, scores, or
   annotations reflect the missing data.

Do not test for specific error messages -- the degradation behavior is the
contract, not the wording.

## Relationship to Other Test Categories

- **Golden tests** verify correctness of output for good input.
- **Adversarial tests** verify correctness of behavior for bad input.
- **Schema tests** verify structural integrity of serialized data.
- **Determinism tests** verify consistency; adversarial tests verify resilience.

Adversarial tests are Terrain's immune system. They ensure the platform remains
trustworthy even when the input does not deserve trust.
