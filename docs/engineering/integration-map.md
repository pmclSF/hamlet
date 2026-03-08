# Integration Map: Signals, Quality, and Migration

How Hamlet's signal types flow through major subsystems.

## Canonical Signal Type Sets

Defined in `internal/signals/signal_types.go`:

| Set | Signal Types | Purpose |
|-----|-------------|---------|
| `MigrationSignalTypes` | frameworkMigration, migrationBlocker, deprecatedTestPattern, dynamicTestGeneration, customMatcherRisk | Migration readiness, blocker tracking |
| `QualitySignalTypes` | weakAssertion, mockHeavyTest, untestedExport, coverageThresholdBreak, coverageBlindSpot | Test quality assessment |

All subsystems import these sets from `signals` rather than defining local copies.

## Signal Flow

```
[Detectors]
  quality/   → weakAssertion, mockHeavyTest, untestedExport, coverageThresholdBreak
  migration/ → deprecatedTestPattern, dynamicTestGeneration, customMatcherRisk, frameworkMigration
  governance/ → policyViolation, legacyFrameworkUsage, runtimeBudgetExceeded
      |
      v
[Snapshot] (snap.Signals)
      |
      +---> [scoring/risk_engine] → snap.Risk (reliability, change, speed surfaces)
      |
      +---> [metrics/] → aggregate counts, ratios, bands (privacy-safe)
      |         |
      |         +---> qualityPostureBand (strong/moderate/weak)
      |         +---> migrationReadinessBand, safeAreaCount, riskyAreaCount
      |
      +---> [migration/readiness] → ReadinessSummary
      |         |
      |         +---> QualityFactors (quality issues compounding migration risk)
      |         +---> AreaAssessments (per-directory safe/caution/risky)
      |         +---> CoverageGuidance (where to add tests for migration safety)
      |
      +---> [heatmap/] → directory and owner risk concentration
      |
      +---> [summary/executive] → ExecutiveSummary (posture, focus, trends)
      |
      +---> [benchmark/export] → privacy-safe Export (schema v3)
      |
      +---> [reporting/] → human-readable CLI output
```

## Cross-Referencing

Migration readiness cross-references quality signals with migration blockers:

1. **QualityFactors**: Files with both migration blockers AND quality issues (e.g., a file with `deprecatedTestPattern` + `weakAssertion`) get flagged as compounded risk.

2. **AreaAssessments**: Each directory is classified as:
   - `safe`: no blockers, no quality issues
   - `caution`: blockers only OR quality issues only
   - `risky`: both blockers AND quality issues

3. **CoverageGuidance**: Directories with compounded risk or untested exports in migration targets get high-priority coverage recommendations.

## Privacy Boundary

The `metrics.Snapshot` and `benchmark.Export` contain only:
- Counts and ratios (not file lists)
- Bands and postures (not paths)
- Framework names (not source code)
- Schema version for forward compatibility

No raw file paths, symbol names, test names, or source code appear in benchmark-safe outputs. The `Export` struct is the serialization boundary between local analysis and any future hosted aggregation.

For test identity and coverage-by-type integration details, see [integration-map-test-identity-coverage.md](integration-map-test-identity-coverage.md).

## Adding a New Signal Type

When adding a signal type that belongs to the migration or quality category:

1. Add the constant to `internal/signals/signal_types.go`
2. Add it to the appropriate canonical set (`MigrationSignalTypes` or `QualitySignalTypes`)
3. All downstream subsystems (readiness, metrics, reporting) automatically pick it up
4. See [writing-a-detector.md](../contributing/writing-a-detector.md) for the full checklist
