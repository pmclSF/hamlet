# Hamlet Test Intelligence -- VS Code Extension

Thin VS Code / Cursor extension that renders Hamlet analysis as sidebar views.
All intelligence lives in the CLI -- the extension is a read-only lens.

See [docs/vscode-extension.md](../../docs/vscode-extension.md) for the design.

## Architecture

```
hamlet analyze --json  -->  TestSuiteSnapshot  -->  TreeDataProviders  -->  Sidebar Views
```

The extension invokes the CLI, parses the JSON snapshot, and renders views.
No business logic is duplicated in TypeScript.

## Sidebar Views

| View | Content |
|------|---------|
| **Overview** | Repository name, framework count, test file count, signal count, risk surfaces, top issues |
| **Health** | Health signals grouped by type (slow, flaky, skipped). Empty state when no runtime data |
| **Quality** | Quality signals grouped by type (weak assertions, mock-heavy, untested exports) |
| **Migration** | Framework summary, blocker count, blockers by type, area assessments (safe/caution/risky) |
| **Review** | Medium+ severity findings grouped by type, owner, directory. Migration blockers surfaced |

## Commands

| Command | Description |
|---------|-------------|
| `Hamlet: Refresh Analysis` | Re-run `hamlet analyze --json` and refresh all views |
| `Hamlet: Open Executive Summary` | Open terminal with `hamlet summary` |
| `Hamlet: Show Migration Blockers` | Open terminal with `hamlet migration blockers` |
| `Hamlet: Reveal File` | Open the file associated with a finding |

## States

All views handle these states:

- **Empty**: No analysis run yet -- shows prompt to refresh
- **Loading**: Analysis in progress -- shows spinner
- **Error**: CLI failed -- shows error message and install guidance
- **Loaded**: Real data rendered from snapshot

## Configuration

| Setting | Default | Description |
|---------|---------|-------------|
| `hamlet.binaryPath` | `hamlet` | Path to the hamlet CLI binary |
| `hamlet.autoRefresh` | `false` | Auto-refresh when files change |

## Development

```bash
cd extension/vscode
npm install
npm run compile
# Press F5 in VS Code to launch Extension Development Host
```

## Source Files

| File | Purpose |
|------|---------|
| `src/extension.ts` | Extension entry point, TreeDataProviders, commands, CLI integration |
| `src/types.ts` | TypeScript types aligned with CLI JSON snapshot contract |
| `src/signal_renderer.ts` | Grouping/filtering helpers for transforming snapshot data |
| `src/views.ts` | View data builders for each sidebar view |
