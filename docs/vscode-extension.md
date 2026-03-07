# VS Code / Cursor Extension

## Purpose

The extension makes Hamlet intelligence explorable inside the editor.

It is not a separate engine.

The extension should remain a thin client over CLI JSON output.

## V1 / V3 architecture

Flow:
1. extension invokes `hamlet analyze --json`
2. CLI returns a TestSuiteSnapshot
3. extension renders structured views

## Sidebar views

### Overview
Repository-level summary:
- framework inventory
- top health issues
- top quality issues
- migration readiness
- highest-risk areas

### Health
Signals like:
- slow tests
- flaky tests
- skipped tests
- dead tests

### Quality
Signals like:
- untested exports
- weak assertions
- mock-heavy tests
- coverage threshold breaks

### Migration
- migration readiness
- blockers
- representative examples

### Review
Grouped triage for:
- blockers
- owners
- packages
- confidence/risk bands

## Editor interactions

- diagnostics based on signals
- hover explanations
- migration hints
- future diff previews

## Implementation status

Source files under `extension/vscode/src/`:
- `types.ts` — TypeScript types aligned with CLI JSON snapshot contract
- `signal_renderer.ts` — Grouping/filtering helpers (groupByType, groupByOwner, groupByDirectory, reviewWorthy, migrationSignals)
- `views.ts` — View data builders (buildOverview, buildHealth, buildQuality, buildReview, buildMigration)

TreeDataProvider implementations pending.

## Rules

- business logic stays in the CLI/core
- avoid duplicating detector logic
- prefer tree views and standard UI surfaces
- webviews should be minimal or deferred
