# PR and CI Integration

> **Status:** Implemented
> **Purpose:** Define how Terrain integrates into GitHub PR workflows via Actions, JSON artifacts, and PR comments
> **Key decisions:**
> - Composite GitHub Action wraps graph build, impact analysis, and PR commenting into a single reusable step
> - All engines produce JSON artifacts in a standard envelope format for downstream consumption
> - PR comments are upserted (create or update) using a stable comment identifier to avoid duplication

See also: [10-json-artifact-schemas.md](10-json-artifact-schemas.md), [14-evidence-scoring-and-confidence-model.md](14-evidence-scoring-and-confidence-model.md)

## Overview

Terrain integrates into PR workflows through GitHub Actions, JSON artifacts, and PR comments. The goal is to surface test intelligence at the point where it matters most: during code review.

## GitHub Action

A composite GitHub Action is provided at `.github/actions/terrain-impact/`:

### Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `node-version` | No | `22` | Node.js version |
| `base-sha` | No | Auto-detected | Base commit SHA |
| `head-sha` | No | Auto-detected | Head commit SHA |
| `comment-header` | No | `terrain-impact` | PR comment identifier |
| `post-comment` | No | `true` | Whether to post/update PR comment |

### Outputs

| Output | Description |
|--------|-------------|
| `artifact-path` | Path to the generated JSON artifact |
| `test-count` | Number of impacted tests |
| `risk-level` | PR risk assessment (low, medium, high) |

### What the Action Does

1. Builds the dependency graph
2. Constructs a `ChangeSet` from the git diff between base and head SHAs (resolves SHAs, infers changed packages/services, detects shallow clones)
3. Runs impact analysis starting from the `ChangeSet` via `AnalyzeChangeSet()`
4. Generates a JSON artifact (`terrain-impact.json`) that includes the full `ChangeSet`
5. Posts or updates a PR comment with:
   - Risk badge (low/medium/high)
   - Impacted test count and coverage confidence
   - Key insights (high-fanout nodes, coverage gaps)
   - Impacted test details (collapsible)
6. Uploads the artifact for downstream consumption
7. Writes a job summary

### Usage

```yaml
- uses: ./.github/actions/terrain-impact
  with:
    post-comment: true
```

SHAs are auto-detected from the PR context (`github.event.pull_request.base.sha` and `github.event.pull_request.head.sha`).

## JSON Artifacts

All insight engines can produce JSON artifacts via the `--artifact` flag:

```bash
terrain impact --base main --artifact    # → .terrain/artifacts/terrain-impact.json
terrain coverage --artifact              # → .terrain/artifacts/terrain-coverage.json
terrain duplicates --artifact            # → .terrain/artifacts/terrain-duplicates.json
```

### Artifact Envelope

Every artifact is wrapped in a standard envelope:

```json
{
  "version": "1.0.0",
  "repo": "repository-name",
  "base_sha": "abc123",
  "head_sha": "def456",
  "generated_at": "2025-01-15T10:30:00Z",
  "results": { }
}
```

The `results` field contains the engine-specific payload (impact results, coverage results, etc.).

## CI Workflow Patterns

### Basic: Impact Check on PRs

```yaml
on:
  pull_request:
    branches: [main]

jobs:
  terrain:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: ./.github/actions/terrain-impact
```

### Advanced: Gate on Risk Level

```yaml
- uses: ./.github/actions/terrain-impact
  id: terrain
- run: |
    if [ "${{ steps.terrain.outputs.risk-level }}" = "high" ]; then
      echo "High risk PR — review terrain impact report"
      exit 1
    fi
```

### Advanced: Upload Artifact for Downstream Jobs

```yaml
- uses: ./.github/actions/terrain-impact
  id: terrain
- uses: actions/upload-artifact@v4
  with:
    name: terrain-impact
    path: ${{ steps.terrain.outputs.artifact-path }}
```

## PR Comment Format

The PR comment includes:

- **Risk badge** — color-coded (green/yellow/red) based on PR risk
- **Metrics table** — impacted test count, coverage confidence, changed file count
- **Key insights** — notable findings (high-fanout nodes, coverage gaps, edge cases)
- **Impacted tests** — collapsible detail section listing affected tests with confidence scores
