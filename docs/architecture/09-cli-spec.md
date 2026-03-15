# CLI Specification

> **Status:** Implemented
> **Purpose:** Define the unified Go CLI command surface, flags, and output modes for the Terrain platform
> **Key decisions:**
> - Single unified Go binary with three-tier command structure: primary, supporting, and advanced/debug
> - All commands support `--json` for structured output and human-readable formatted output by default
> - Debug commands are namespaced under `terrain debug` to separate diagnostic tooling from user-facing workflows
> - Artifact generation uses `--artifact` flag to write JSON files alongside console output

See also: [10-json-artifact-schemas.md](10-json-artifact-schemas.md)

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Output results as JSON |
| `--verbose` | Show detailed progress information |
| `--root PATH` | Repository root (defaults to current directory) |

## Primary Commands

These are the canonical user journeys — the main entry points for interacting with Terrain.

### `terrain analyze`

Full repository analysis. Answers: "What is the state of our test system?"

```bash
terrain analyze [--root PATH] [--json] [--format json|text] [--verbose]
               [--write-snapshot] [--coverage PATH] [--coverage-run-label LABEL]
               [--runtime PATH] [--slow-threshold MS]
```

| Flag | Description |
|------|-------------|
| `--format` | Output format: `json` or `text` (default: text) |
| `--verbose` | Show all findings in analyze output |
| `--write-snapshot` | Persist analysis snapshot for later comparison |
| `--coverage PATH` | Path to coverage data file (LCOV or Istanbul JSON) |
| `--coverage-run-label LABEL` | Coverage run label: `unit`, `integration`, or `e2e` |
| `--runtime PATH` | Path to runtime artifact (JUnit XML or Jest JSON); comma-separated for multiple |
| `--slow-threshold MS` | Slow test threshold in milliseconds (default: 5000) |

### `terrain impact`

Analyze the impact of code changes on the test suite. Answers: "What validations matter for this change?"

```bash
terrain impact [--root PATH] [--base REF] [--json] [--show VIEW] [--owner NAME]
```

| Flag | Required | Default | Description |
|------|----------|---------|-------------|
| `--base` | No | `HEAD~1` | Git base ref for diff |
| `--show` | No | — | Drill-down view: `units`, `gaps`, `tests`, `owners`, `graph`, `selected` |
| `--owner` | No | — | Filter results by CODEOWNERS owner |

Output: changed files, impacted tests with confidence scores and impact chains, risk assessment.

### `terrain insights`

Surface actionable findings about the test system. Answers: "What should we fix in our test system?"

```bash
terrain insights [--root PATH] [--json]
```

### `terrain explain <target>`

Explain why Terrain made a decision or why a test/finding exists. Answers: "Why did Terrain make this decision?"

```bash
terrain explain <target> [--root PATH] [--base REF] [--json]
```

| Flag | Description |
|------|-------------|
| `--base` | Git base ref for diff (used when explaining impact-related decisions) |

Target can be: a test file path, test case ID, code unit, owner, or finding type. Use `terrain explain selection` to explain the overall test selection strategy.

Output: dependency paths, confidence scores, and human-readable reasons for each edge.

## Supporting Commands

### `terrain init`

Detect data files and print recommended analyze command.

```bash
terrain init [--root PATH]
```

### `terrain summary`

Executive summary with posture, trends, and benchmark readiness.

```bash
terrain summary [--root PATH] [--json]
```

### `terrain focus`

Prioritized next actions.

```bash
terrain focus [--root PATH] [--json]
```

### `terrain posture`

Detailed posture breakdown with measurement evidence.

```bash
terrain posture [--root PATH] [--json]
```

### `terrain portfolio`

Portfolio intelligence (cost, breadth, leverage, redundancy).

```bash
terrain portfolio [--root PATH] [--json]
```

### `terrain metrics`

Aggregate metrics scorecard (privacy-safe).

```bash
terrain metrics [--root PATH] [--json]
```

### `terrain compare`

Compare two snapshots for trend analysis.

```bash
terrain compare [--from PATH] [--to PATH] [--json]
```

### `terrain select-tests`

Recommend a protective test set for a change.

```bash
terrain select-tests [--root PATH] [--json]
```

### `terrain pr`

PR/change-scoped analysis.

```bash
terrain pr [--root PATH] [--json]
```

### `terrain show <entity> <id>`

Drill into a specific test, unit, owner, or finding.

```bash
terrain show <entity> <id> [--root PATH] [--json]
```

### `terrain migration readiness|blockers|preview`

Migration intelligence commands.

```bash
terrain migration readiness [--json]
terrain migration blockers [--json]
terrain migration preview --file PATH [--json]
```

### `terrain policy check`

Evaluate against local policy rules.

```bash
terrain policy check [--root PATH] [--json]
```

Exit codes: 0 = pass, 2 = violations found, 1 = error.

### `terrain export benchmark`

Benchmark-safe JSON export (no raw paths or source code).

```bash
terrain export benchmark [--root PATH]
```

## Advanced / Debug Commands

Debug commands provide low-level diagnostic access to internal engines and graph state. All are namespaced under `terrain debug`.

### `terrain debug graph`

Display dependency graph statistics.

```bash
terrain debug graph [--root PATH] [--json]
```

Output includes: node count, edge count, test count, fixture count, helper count, source file count, graph density.

### `terrain debug coverage`

Structural coverage analysis.

```bash
terrain debug coverage [--root PATH] [--json] [--band <band>] [--artifact]
```

| Flag | Description |
|------|-------------|
| `--band` | Filter by coverage band: High, Medium, or Low |
| `--artifact` | Write terrain-coverage.json to .terrain/artifacts/ |

Output: per-source-file coverage with test counts, direct/indirect test lists, and coverage band.

### `terrain debug fanout`

High-fanout node analysis.

```bash
terrain debug fanout [--root PATH] [--json] [--threshold <number>] [--flagged-only]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--threshold` | `10` | Flag nodes with transitive fanout above this value |
| `--flagged-only` | — | Only show flagged nodes |

Output: per-node fanout counts (direct and transitive), flagged status.

### `terrain debug duplicates`

Duplicate test cluster analysis.

```bash
terrain debug duplicates [--root PATH] [--json] [--threshold <number>] [--artifact]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--threshold` | `0.6` | Minimum similarity score (0-1) |
| `--artifact` | — | Write terrain-duplicates.json to .terrain/artifacts/ |

Output: duplicate clusters with member tests and similarity scores.

### `terrain debug depgraph`

Full dependency graph analysis (all engines).

```bash
terrain debug depgraph [--root PATH] [--json]
```

## Output Modes

All commands support two output modes:

- **Human-readable** (default) — formatted text with headers, tables, and color
- **JSON** (`--json`) — structured JSON for scripting and CI integration

Artifact commands (`--artifact`) write JSON files to `.terrain/artifacts/` in addition to console output.
