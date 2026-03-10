# Hamlet JSON Schemas

Hamlet emits structured JSON for all top-level commands. This document is the
canonical reference for field-level semantics.

## Stability

- `snapshotMeta.schemaVersion` is the schema compatibility key for snapshot JSON.
- Unknown fields should be ignored by consumers for forward compatibility.
- Optional fields are omitted when no evidence is available.

## `hamlet analyze --json`

Top-level type: `models.TestSuiteSnapshot`.

Key fields:

- `snapshotMeta`: schema and methodology provenance.
- `repository`: repository metadata (`name`, `rootPath`, `languages`, `commitSHA`).
- `frameworks`: detected framework inventory.
- `testFiles`: per-file metrics and linkage.
- `testCases`: stable test identities.
- `codeUnits`: exported/public units.
- `signals`: canonical findings.
- `risk`: explainable risk surfaces.
- `measurements`: posture-layer outputs.
- `portfolio`: portfolio intelligence outputs.
- `coverageSummary` / `coverageInsights`: coverage-derived intelligence.
- `ownership`: resolved ownership map.
- `policies`: loaded policy configuration (if present).
- `metadata`: pipeline summary metadata.
- `dataSources`: runtime/coverage/policy availability and impacts.
- `generatedAt`: snapshot generation timestamp.

## `hamlet compare --json`

Top-level type: `comparison.SnapshotComparison`.

Key fields:

- `fromTime`, `toTime`: compared snapshot timestamps.
- `methodologyCompatible`: whether methodology-sensitive deltas are comparable.
- `methodologyNotes`: rationale for compatibility decision.
- `signalDeltas`, `riskDeltas`, `frameworkChanges`.
- `newSignalExamples`, `resolvedSignalExamples`.
- `testCaseDeltas`, `coverageDelta`, `ownershipDelta`.
- `postureDeltas`, `measurementDeltas`.
- `lifecycleContinuity`.

## `hamlet metrics --json`

Top-level type: `metrics.Snapshot`.

Sections:

- `structure`
- `health`
- `quality`
- `changeReadiness`
- `governance`
- `risk`
- `notes`

## `hamlet policy check --json`

Top-level object:

- `policyFile`: string or `null`.
- `pass`: boolean.
- `violations`: `[]models.Signal`.
- `message`: human-readable policy status.

## `hamlet impact --json`

Top-level type: `impact.ImpactResult`.

Includes:

- changed scope
- impacted code units/tests
- protection gaps
- impacted owners
- change-risk posture
- recommendations and limitations

