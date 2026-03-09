# Verification System Map

This document maps how Hamlet's verification components connect across test
layers, fixture flows, and CI integration points.

## Test Layer Architecture

```
+------------------------------------------------------------------+
|                        CI Pipeline                                |
|  +------------------------------------------------------------+  |
|  |  Tier 1: PR Gates                                          |  |
|  |  go vet --> go build --> go test (unit + golden)            |  |
|  +------------------------------------------------------------+  |
|  |  Tier 2: Release Gates                                     |  |
|  |  Tier 1 + determinism + schema + E2E + CLI + docs          |  |
|  +------------------------------------------------------------+  |
|  |  Tier 3: Nightly                                           |  |
|  |  benchmarks + large-scale + adversarial                    |  |
|  +------------------------------------------------------------+  |
+------------------------------------------------------------------+

+------------------------------------------------------------------+
|                      Test Pyramid                                 |
|                                                                   |
|                         /\                                        |
|                        /  \       E2E Scenario Tests              |
|                       / E2E\      (e2e_test.go)                   |
|                      /------\                                     |
|                     /  CLI   \    CLI Regression Tests             |
|                    /  Regress \   (cli_test.go)                    |
|                   /------------\                                   |
|                  / Integration  \  Fixture-based cross-subsystem   |
|                 /    Tests       \ tests per package               |
|                /------------------\                                |
|               /    Unit Tests      \ Table-driven, per-function   |
|              /    (largest layer)    \ tests across all packages   |
|             /------------------------\                             |
+------------------------------------------------------------------+
```

## Fixture Flow

```
internal/testdata/
  |
  +--> sample-repo/          Source fixture: real repo structure
  |      |
  |      +--> go files, JS files, Python files
  |      +--> config files (go.mod, package.json, etc.)
  |      +--> CODEOWNERS
  |
  +--> golden/               Golden reference files
  |      |
  |      +--> summary-report.golden
  |      +--> analyze-report.golden
  |      +--> posture-report.golden
  |      +--> impact-report.golden
  |
  +--> fixtures/             Pre-computed snapshots for fast loading
         |
         +--> full-analysis.json
         +--> comparison-pair.json
         +--> impact-scenario.json

Fixture lifecycle:
  Create fixture --> Load in test --> Run pipeline --> Assert output
                                                         |
                                          Compare to golden file (if golden test)
                                          Compare to self (if determinism test)
                                          Check counts/contains (if E2E test)
```

## CI Integration Points

```
GitHub Event          Gate                   Test Category
-----------          ----                   -------------
PR opened        --> Tier 1 gates      --> Unit tests
                                        --> Golden tests
                                        --> go vet + go build

PR approved      --> (same as above,       re-run on merge queue)

Release branch   --> Tier 2 gates      --> All of Tier 1
                                        --> E2E scenarios
                                        --> CLI regression
                                        --> Determinism suite
                                        --> Schema compatibility
                                        --> Docs consistency

Nightly cron     --> Tier 3 gates      --> Full benchmarks
                                        --> Large-scale tests
                                        --> Adversarial expansion
```

## Subsystem Coverage Matrix

Each subsystem is tested at multiple layers. An "x" indicates the layer
provides meaningful coverage for that subsystem.

```
Subsystem            Unit  Integration  Golden  E2E  CLI  Determinism
---------            ----  -----------  ------  ---  ---  -----------
Detectors             x        x                 x
Signals               x        x          x      x            x
Measurements          x        x                 x
Posture               x        x          x      x            x
Risk Scoring          x        x                 x
Renderers             x                   x      x
CLI Commands                                     x    x
Impact Analysis       x        x          x      x
Ownership             x        x                 x            x
Graph/Dependencies    x        x                 x
Export/Privacy        x        x                 x            x
Schema Compliance              x                 x
Migration Readiness   x        x                 x
Portfolio             x        x          x      x
Benchmarks            x                                       x
```

## Key Verification Contracts

| Contract | Enforced By | Failure Mode |
|----------|-------------|--------------|
| Deterministic JSON output | Determinism tests | Same input, different output across runs |
| Schema stability | Schema compatibility tests | Output breaks published JSON schema |
| CLI completeness | Docs-consistency checks | Command exists in code but not in help/docs |
| Golden output stability | assertGolden helper | Rendered output changes without explicit update |
| Privacy boundaries | Export privacy tests | Sensitive fields leak into public exports |
| Posture band thresholds | Posture unit tests | Score maps to wrong quality band |

## Related Documents

- `docs/engineering/e2e-scenario-testing.md` -- E2E test details
- `docs/engineering/cli-and-docs-regression-testing.md` -- CLI test details
- `docs/engineering/ui-regression-testing.md` -- View-model test details
- `docs/release/quality-bar-and-gates.md` -- Gate tier definitions
- `docs/contributing/testing-and-quality.md` -- Contributor testing guide
- `docs/engineering/determinism.md` -- Deterministic output contract
