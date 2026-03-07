# extension/vscode

Thin VS Code / Cursor extension that consumes `hamlet analyze --json` output. See [docs/vscode-extension.md](../../docs/vscode-extension.md) for the design.

## Architecture
The extension invokes the CLI and renders structured views. It does not re-implement business logic.

## Sidebar views

### Overview
Repository-level summary: frameworks, test file counts, top risk areas, top findings.

### Health
Health-category signals grouped by type (slow, flaky, skipped, dead tests).

### Quality
Quality-category signals grouped by type (weak assertions, mock-heavy tests, untested exports, coverage breaks).

### Review
Findings requiring human attention, grouped by:
- signal type
- owner
- directory
- severity

Supports empty states and unresolved ownership display.

### Migration
Migration readiness and blockers:
- blocker groups by type
- representative examples
- framework summary
- owner grouping for migration signals

## Source files

- `src/types.ts` — TypeScript types aligned with CLI JSON snapshot contract
- `src/signal_renderer.ts` — Grouping/filtering helpers for transforming snapshot data into view items
- `src/views.ts` — View data builders for each sidebar view

## Status
Scaffolded with real type definitions, grouping logic, and view builders. TreeDataProvider implementations pending.
