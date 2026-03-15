# UI Requirements

> **Status:** Implemented for VS Code extension; Aspirational for dashboard
> **Purpose:** Define UI surface requirements for presenting Terrain analysis results to users.
> **Key decisions:**
> - VS Code extension is a thin, read-only client over CLI JSON output — no domain logic duplicated
> - Snapshot-driven rendering: views always reflect the latest CLI snapshot
> - Dashboard is planned as an aggregate-only view — no raw file paths or individual developer metrics exposed
> - PR integration designed around impact analysis and coverage confidence badges

**See also:** [09-cli-spec.md](09-cli-spec.md), [10-json-artifact-schemas.md](10-json-artifact-schemas.md)

## VS Code Extension

The VS Code extension is a thin client over Terrain's JSON output. It invokes CLI commands with `--json` and renders the results in sidebar views. No domain logic is duplicated in the extension.

### Views

| View | Source Command | Content |
|------|---------------|---------|
| Overview | `terrain summary --json` | Posture band, risk dimensions, key metrics |
| Health | `terrain analyze --json` | Health signals (flaky, slow, skipped, dead tests) |
| Quality | `terrain analyze --json` | Quality signals (weak assertions, mock-heavy, untested exports) |
| Migration | `terrain migration readiness --json` | Migration blockers, readiness assessment, area safety |
| Review | `terrain analyze --json` | Grouped findings by owner, directory, category |

### Design Principles

- **Read-only.** The extension displays findings. It does not modify code or run tests.
- **Snapshot-driven.** Views render from the latest snapshot. Refreshing re-runs the CLI command.
- **Thin client.** All intelligence lives in the CLI. The extension is presentation only.
- **Graceful degradation.** If no snapshot exists, the extension shows a "run terrain analyze" prompt.

## Future Dashboard (Hosted Product)

The dashboard is a planned web UI for organization-wide visibility. Key requirements:

### Org-Level Views

- Cross-repo posture comparison
- Risk heatmaps by team and service
- Trend charts across snapshot history
- Benchmark positioning against anonymized peer data

### PR Integration

- Inline impact analysis in PR review
- Coverage confidence badges
- Risk level indicators
- Test recommendation summaries

### Design Principles for Dashboard

- **Aggregate, never raw.** The dashboard shows aggregate metrics. No raw file paths, source code, or individual developer metrics.
- **Privacy boundary.** Same as CLI: benchmark exports contain only counts and ratios.
- **Snapshot as boundary.** The dashboard consumes snapshots. It does not run analysis directly.
