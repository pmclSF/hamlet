# Hamlet — Architecture Overview

Hamlet is a signal-first test intelligence platform. It analyzes repository structure, test code, runtime artifacts, and local policy to surface actionable findings about test suite health, quality, migration readiness, and risk.

## Core Principles

- **Signals are the core abstraction.** Every finding is a structured signal with type, severity, evidence, and location.
- **Snapshots are the canonical artifact.** The `TestSuiteSnapshot` is the serialized boundary between engine, reporting, and future aggregation.
- **Risk must be explainable.** Risk surfaces are derived from signals with transparent scoring, not opaque scores.
- **Local-first.** Hamlet is useful on a single machine, without accounts, SaaS, or network access.
- **Privacy boundary.** Aggregate metrics and benchmark exports never expose raw file paths or source code.

## Architecture

```
Repository scan
  -> Framework detection + test file discovery
    -> Signal detection (quality, health, migration, governance)
      -> Risk scoring (reliability, change, speed)
        -> Snapshot (TestSuiteSnapshot)
          -> Reporting (human-readable, JSON, executive summary)
```

See [docs/architecture.md](docs/architecture.md) for the full layered architecture.

## Key Design Documents

| Document | Purpose |
|----------|---------|
| [docs/architecture.md](docs/architecture.md) | Layered architecture: analysis, signals, risk, reporting |
| [docs/signal-model.md](docs/signal-model.md) | Signal abstraction and schema |
| [docs/signal-catalog.md](docs/signal-catalog.md) | All signal types and categories |
| [docs/roadmap.md](docs/roadmap.md) | Milestone history and future work |
| [docs/cli-spec.md](docs/cli-spec.md) | Full command and flag reference |
| [docs/engineering/detector-architecture.md](docs/engineering/detector-architecture.md) | Registry-based detector plugin system |
| [docs/engineering/architecture-map.md](docs/engineering/architecture-map.md) | Contributor-facing component and pipeline map |
| [docs/engineering/detector-audit.md](docs/engineering/detector-audit.md) | Evidence classification for all detectors |
| [docs/engineering/hosted-future.md](docs/engineering/hosted-future.md) | What remains for hosted/org product |
| [docs/contributing/writing-a-detector.md](docs/contributing/writing-a-detector.md) | How to add a new signal detector |

## Package Map

```
cmd/hamlet/          CLI entry point
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

See [docs/engineering/architecture-map.md](docs/engineering/architecture-map.md) for the full contributor-oriented component map.

## Migration Context

Hamlet originated as a multi-framework test converter (V2). That converter engine (JavaScript ES modules, `src/`) remains in the repository and is fully functional. V3 reframes migration as one dimension of broader test intelligence rather than the sole product.

The V2 converter architecture is documented in [docs/legacy/v2-converter-architecture.md](docs/legacy/v2-converter-architecture.md).

## Extension Architecture

The VS Code extension is intentionally thin. It invokes `hamlet analyze --json`, reads the snapshot, and renders views. No domain logic is duplicated in the extension. See [docs/vscode-extension.md](docs/vscode-extension.md).
