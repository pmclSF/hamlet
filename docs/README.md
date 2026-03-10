# Hamlet Documentation

**Signal-first test intelligence for engineering teams**

## Product Evolution

- **Legacy** was conversion-led: a multi-framework test converter (JS/Java/Python, 25 directions).
- **The current engine** is signal-led: a test intelligence platform that surfaces risk, quality, migration readiness, and governance from static and runtime analysis.
- **Migration remains the acquisition wedge** — the pain of framework migration is what brings teams to Hamlet. The current engine turns that pain into broader test intelligence.

The legacy converter engine is preserved and functional. See [legacy/](legacy/) for historical architecture docs.

## Start Here

- [Demo Walkthrough](demo.md) — try Hamlet in a few minutes
- [Canonical User Journeys](product/canonical-user-journeys.md) — primary workflows and expected outcomes
- [CLI Specification](cli-spec.md) — full command and flag reference
- [Architecture](architecture.md) — how Hamlet works internally

## Product

- [Vision](vision.md) — why Hamlet exists
- [Product Concept](product-concept.md) — what Hamlet does
- [Master Plan](MASTER_PLAN.md) — strategic direction
- [Paid Product](paid-product.md) — future commercial direction
- [UX Blueprint](ux-blueprint.md) — user experience design

## Technical

- [Signal Model](signal-model.md) — the core signal abstraction
- [Signal Catalog](signal-catalog.md) — all signal types
- [VS Code Extension](vscode-extension.md) — extension architecture and views
- [Roadmap](roadmap.md) — milestone history and future work
- [Implementation Workbook](implementation-workbook.md) — stage-by-stage implementation log

## Engineering

- [Architecture Map](engineering/architecture-map.md) — contributor-facing component map
- [Detector Architecture](engineering/detector-architecture.md) — registry-based detector plugin system
- [Detector Audit](engineering/detector-audit.md) — evidence classification for all detectors
- [Deterministic Output](engineering/determinism.md) — deterministic output contract and enforcement
- [Engineering Roadmap](engineering/future-work.md) — incremental analysis, parallelization, and layer separation
- [Hosted Future](engineering/hosted-future.md) — what remains for hosted/org product
- [Test Identity](engineering/test-identity.md) — deterministic test identity model
- [Test Type Inference](engineering/test-type-inference.md) — evidence-based test classification
- [Code Unit Inventory](engineering/code-unit-inventory.md) — normalized code structure model
- [Coverage Ingestion](engineering/coverage-ingestion.md) — LCOV/Istanbul ingestion and normalization
- [Coverage Attribution](engineering/coverage-attribution.md) — structural coverage for code units
- [Per-Test Coverage](engineering/per-test-coverage.md) — per-test coverage attribution model
- [Snapshot Test Lineage](engineering/snapshot-test-lineage.md) — longitudinal test tracking

## User Guides

- [Coverage by Type](user-guides/coverage-by-type.md) — analyze coverage by unit/integration/e2e

## Contributing

- [Writing a Detector](contributing/writing-a-detector.md) — how to add a new signal detector
- [Test Identity & Coverage](contributing/test-identity-and-coverage.md) — extending identity, inference, and coverage

## Release

- [Release Checklist](release-checklist.md) — launch readiness status
- [Release Process](releasing.md) — versioning and release workflow

## Legacy Converter Engine

Historical documentation for the JavaScript converter engine:

- [Converter Architecture (legacy)](legacy/converter-architecture-legacy.md)
- [Legacy Notes](legacy/legacy-notes.md)
- [Getting Started](guides/getting-started.md)
- [Migration Guide](guides/migration-guide.md)
- [CLI Reference (legacy)](api/cli.md)
- [Configuration (legacy)](api/configuration.md)
- [Conversion Process](api/conversion.md)
- [Jest ESM Strategy](adr/004-jest-esm-strategy.md)
