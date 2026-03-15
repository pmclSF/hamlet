# Detector Architecture

## Overview

Terrain's signal detection uses a registry-based plugin architecture. Detectors are self-describing components that examine a `TestSuiteSnapshot` and emit structured `Signal` objects.

## Core Abstractions

### Detector Interface

```go
type Detector interface {
    Detect(snap *models.TestSuiteSnapshot) []models.Signal
}
```

Every detector implements this single-method interface. Detectors should be stateless and deterministic given the same snapshot input.

### DetectorMeta

Each detector is registered with metadata describing its identity and capabilities:

| Field | Purpose |
|-------|---------|
| `ID` | Stable identifier (e.g., `quality.weak-assertion`) |
| `Domain` | Area of concern: quality, migration, governance, health, coverage |
| `EvidenceType` | How evidence is obtained: structural-pattern, path-name, runtime, coverage, policy, codeowners |
| `Description` | Human-readable summary |
| `SignalTypes` | Signal types this detector may emit |
| `RequiresFileIO` | Whether the detector reads files beyond the snapshot |
| `DependsOnSignals` | Whether it reads signals from prior detectors |

### DetectorRegistry

An ordered collection of `DetectorRegistration` entries (detector + metadata). Detectors execute in registration order, which matters for dependencies (governance detectors read signals from quality detectors).

Key operations:
- `Register(reg)` — add a detector
- `Run(snap)` — execute all detectors
- `RunDomain(snap, domain)` — execute detectors in one domain
- `ByDomain(domain)` — filter registrations by domain
- `All()` — list all registrations in order

## Execution Model

```
DefaultRegistry(config)
  -> Quality detectors (no dependencies)
  -> Migration detectors (no dependencies)
  -> Governance detector (depends on quality/migration signals)

registry.Run(snapshot)
  -> Detectors execute sequentially in registration order
  -> Each appends signals to snapshot.Signals
```

## Pipeline

The `engine.RunPipeline(root)` function orchestrates the full analysis:

1. Static analysis (file discovery, framework detection)
2. Policy loading (determines whether governance detector is registered)
3. Detector registry construction and execution
4. Ownership resolution
5. Risk scoring

This replaces the previously duplicated detector invocations across CLI commands.

## Current Detectors

| ID | Domain | Evidence | Signals |
|----|--------|----------|---------|
| `quality.weak-assertion` | quality | structural-pattern | weakAssertion |
| `quality.mock-heavy` | quality | structural-pattern | mockHeavyTest |
| `quality.untested-export` | quality | path-name | untestedExport |
| `quality.coverage-threshold` | coverage | coverage | coverageThresholdBreak |
| `migration.deprecated-pattern` | migration | structural-pattern | deprecatedTestPattern |
| `migration.dynamic-test-generation` | migration | structural-pattern | dynamicTestGeneration |
| `migration.custom-matcher` | migration | structural-pattern | customMatcherRisk |
| `migration.framework-migration` | migration | structural-pattern | frameworkMigration |
| `governance.policy` | governance | policy | policyViolation, legacyFrameworkUsage, runtimeBudgetExceeded |

## Adding a New Detector

See [contributing/writing-a-detector.md](../contributing/writing-a-detector.md).
