# Contributor Architecture Map

This document provides a visual map of how Hamlet's components connect.
Use it to orient yourself when adding a new detector, report, or command.

## Command Surface

```
hamlet analyze       -->  engine.RunPipeline  -->  reporting.RenderAnalyzeReport
hamlet summary       -->  engine.RunPipeline  -->  summary.Build  -->  reporting.RenderExecutiveSummary
hamlet posture       -->  engine.RunPipeline  -->  reporting.RenderPostureReport
hamlet metrics       -->  engine.RunPipeline  -->  metrics.Derive  -->  reporting.RenderMetricsReport
hamlet impact        -->  engine.RunPipeline  -->  impact.Analyze  -->  reporting.RenderImpactReport
hamlet compare       -->  loadSnapshot x2  -->  comparison.Compare  -->  reporting.RenderComparisonReport
hamlet migration *   -->  engine.RunPipeline  -->  migration.*  -->  reporting.RenderMigration*
hamlet policy check  -->  analysis + quality  -->  governance.Evaluate  -->  reporting.RenderPolicyReport
hamlet export bench  -->  engine.RunPipeline  -->  benchmark.BuildExport
```

## Pipeline Flow

```
cmd/hamlet/main.go
  |
  v
engine.RunPipeline(root)
  |
  |-- 1. analysis.New(root).Analyze()          --> TestSuiteSnapshot (frameworks, test files, code units)
  |-- 2. policy.Load(root)                     --> PolicyConfig (optional)
  |-- 3. engine.DefaultRegistry(cfg).Run(snap) --> Signals appended to snapshot
  |      |-- quality detectors (4)
  |      |-- migration detectors (5)
  |      |-- governance detector (1, if policy present)
  |-- 4. ownership.Resolve()                   --> Owner labels on signals
  |-- 5. runtime.Ingest() (optional)           --> Health signals (slow, flaky, skipped)
  |-- 6. scoring.ComputeRisk()                 --> Risk surfaces
  |-- 7. coverage.Ingest() (optional)          --> CoverageSummary + insights
  |-- 8. measurement.ComputeSnapshot()         --> Posture measurements
  |
  v
PipelineResult { Snapshot, HasPolicy }
```

## Snapshot Contract

The `TestSuiteSnapshot` is the serialized boundary between engine and consumers:

```
TestSuiteSnapshot
  ├── Repository         metadata (name, root, languages, commit, branch)
  ├── Frameworks[]       detected frameworks with file counts
  ├── TestFiles[]        discovered test files with stats
  ├── TestCases[]        extracted test case identities
  ├── CodeUnits[]        exported functions/classes from source files
  ├── Signals[]          structured findings (the core product)
  ├── Risk[]             computed risk surfaces by dimension
  ├── Ownership{}        file -> owner mappings
  ├── CoverageSummary    aggregate coverage if ingested
  ├── CoverageInsights[] actionable coverage findings
  ├── Measurements       posture bands by dimension
  └── GeneratedAt        timestamp
```

## Detector Registry

Detectors implement `signals.Detector` and are registered in `engine.DefaultRegistry()`:

```
Quality Domain:
  quality.weak-assertion         WeakAssertionDetector
  quality.mock-heavy             MockHeavyDetector
  quality.untested-export        UntestedExportDetector
  quality.coverage-threshold     CoverageThresholdDetector

Migration Domain:
  migration.deprecated-pattern        DeprecatedPatternDetector
  migration.dynamic-test-generation   DynamicTestGenerationDetector
  migration.custom-matcher            CustomMatcherDetector
  migration.unsupported-setup         UnsupportedSetupDetector
  migration.framework-migration       FrameworkMigrationDetector

Governance Domain (conditional):
  governance.policy              GovernanceDetector (only if policy loaded)

Health (runtime-backed, run separately):
  SlowTestDetector
  FlakyTestDetector
  SkippedTestDetector
```

To add a detector, see [writing-a-detector.md](../contributing/writing-a-detector.md).

## Report Pipeline

```
Snapshot + derived data  -->  Renderer  -->  human-readable or JSON

Key renderers in internal/reporting/:
  RenderAnalyzeReport()
  RenderExecutiveSummary()
  RenderPostureReport()
  RenderMetricsReport()
  RenderComparisonReport()
  RenderImpactReport() + drill-down variants
  RenderMigrationReport() / RenderMigrationBlockers() / RenderMigrationPreview()
  RenderPolicyReport()
```

## Extension Boundary

```
extension/vscode/
  extension.ts  -->  execFile("hamlet", ["analyze", "--json"])
                      |
                      v
                 TestSuiteSnapshot (parsed JSON)
                      |
                      v
                 TreeDataProviders render sidebar views
```

The extension never computes signals, risk, or summaries.
It reads the snapshot and renders. That is the boundary.

## Package Map

```
cmd/hamlet/          CLI entry point and command routing
internal/
  analysis/          Repository scanning, framework detection, test file discovery
  benchmark/         Privacy-safe benchmark export and segmentation
  comparison/        Snapshot-to-snapshot trend comparison
  coverage/          Coverage artifact ingestion (LCOV, Istanbul) and attribution
  engine/            Pipeline orchestration and detector registry
  governance/        Policy evaluation and governance signals
  health/            Runtime-backed health detectors (slow, flaky, skipped)
  heatmap/           Risk concentration model (directory and owner hotspots)
  identity/          Test identity hashing and normalization
  impact/            Change-scope impact analysis
  measurement/       Posture measurement framework
  metrics/           Aggregate metric derivation
  migration/         Migration detectors, readiness model, preview boundary
  models/            Canonical data models (Signal, Snapshot, Risk, Framework, etc.)
  ownership/         Ownership resolution (CODEOWNERS, config, directory fallback)
  policy/            Policy config model and YAML loader
  quality/           Quality signal detectors
  reporting/         Human-readable report renderers
  runtime/           Runtime artifact ingestion (JUnit XML, Jest JSON)
  scoring/           Explainable risk engine (reliability, change, speed)
  signals/           Signal detector interface, registry, runner
  summary/           Executive summary builder (posture, trends, focus, recommendations)
  testcase/          Test case extraction and identity collision detection
  testtype/          Test type inference (unit, integration, e2e)
```
