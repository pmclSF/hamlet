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
| [docs/contributing/writing-a-detector.md](docs/contributing/writing-a-detector.md) | How to add a new signal detector |

## Package Map

```
cmd/hamlet/          CLI entry point
internal/
  engine/            Analysis pipeline orchestration, detector registry builder
  analysis/          Repository scanning, framework detection, test file discovery
  signals/           Signal detector interface, registry, runner
  quality/           Quality signal detectors (weak assertions, mock-heavy, untested exports)
  migration/         Migration detectors and readiness model
  governance/        Policy evaluation and governance signals
  scoring/           Explainable risk engine (reliability, change, speed)
  ownership/         Ownership resolution (CODEOWNERS, config, directory fallback)
  policy/            Policy config model and YAML loader
  comparison/        Snapshot-to-snapshot trend comparison
  metrics/           Aggregate metric derivation
  heatmap/           Risk concentration model (directory and owner hotspots)
  benchmark/         Privacy-safe benchmark export and segmentation
  summary/           Executive summary builder (posture, trends, focus, benchmark readiness)
  reporting/         Human-readable report renderers
  models/            Canonical data models (Signal, Snapshot, Risk, Framework, etc.)
```

## Migration Context

Hamlet originated as a multi-framework test converter (V2). That converter engine (JavaScript ES modules, `src/`) remains in the repository and is fully functional. V3 reframes migration as one dimension of broader test intelligence rather than the sole product.

The V2 converter architecture is documented in [docs/legacy/v2-converter-architecture.md](docs/legacy/v2-converter-architecture.md).

## Extension Architecture

The VS Code extension is intentionally thin. It invokes `hamlet analyze --json`, reads the snapshot, and renders views. No domain logic is duplicated in the extension. See [docs/vscode-extension.md](docs/vscode-extension.md).
