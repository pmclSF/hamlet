# internal/

V3 engine packages. See [docs/architecture.md](../docs/architecture.md) for the layer model.

## Packages

| Package | Layer | Purpose |
|---------|-------|---------|
| models | Core | Domain types: TestSuiteSnapshot, Framework, TestFile, CodeUnit, Signal |
| analysis | Static Analysis | Repository scanning, test file discovery, framework detection |
| runtime | Runtime Ingestion | CI artifact parsing, coverage reports, runtime metrics |
| signals | Signal Engine | Signal registry, detector interfaces, canonical Signal type |
| health | Signal Engine | Health detectors: slow, flaky, skipped, dead tests |
| quality | Signal Engine | Quality detectors: weak assertions, mock-heavy, untested exports |
| migration | Signal Engine | Migration detectors: blockers, deprecated patterns |
| governance | Signal Engine | Governance detectors: policy violations, legacy usage |
| scoring | Risk Engine | Risk aggregation: reliability, change, speed, governance risk |
| reporting | Reporting | CLI output, JSON serialization, snapshot rendering |
| ownership | Risk Engine | Code ownership resolution |
| policy | Risk Engine | Policy model, local policy evaluation |
| util | Shared | Common utilities |

## Status
Scaffolded. Implementation begins at Stage 2 (models + signals).
