# Terrain CLI Benchmark System

## Purpose

The CLI benchmark system exercises Terrain's primary CLI commands against a set of local repositories, capturing both **performance** (runtime, exit status) and **result quality** (output structure, expected content sections, credibility scoring).

It validates that Terrain's CLI workflows behave sensibly on real repositories with different shapes — different languages, test frameworks, monorepo structures, and sizes.

## Quick Start

```bash
# Build and run the benchmark CLI
go run ./cmd/terrain-bench/

# Single repo
go run ./cmd/terrain-bench/ --repo terrain

# Single command across all repos
go run ./cmd/terrain-bench/ --command analyze

# Auto-discover repos from a directory
go run ./cmd/terrain-bench/ --discover /path/to/repos

# Use a specific terrain binary (skip auto-rebuild)
go run ./cmd/terrain-bench/ --terrain /usr/local/bin/terrain

# Skip rebuilding and use an existing binary
go run ./cmd/terrain-bench/ --rebuild=false

# Or build the binary first
go build -o terrain-bench ./cmd/terrain-bench/
./terrain-bench --repo terrain --command analyze

# Shell wrapper (delegates to cmd/terrain-bench/)
./benchmarks/run-cli-benchmarks.sh
./benchmarks/run-cli-benchmarks.sh --repo terrain --command analyze
```

## Configuration

### `benchmarks/repos.json`

The config file lists repositories to benchmark:

```json
{
  "repos": [
    {
      "name": "terrain",
      "path": ".",
      "type": "real",
      "description": "Terrain itself"
    },
    {
      "name": "express",
      "path": "benchmarks/repos/express",
      "type": "real",
      "description": "Node.js HTTP framework, mocha tests, CommonJS"
    }
  ]
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `name` | yes | Display name for the repo |
| `path` | yes | Path to repo (relative to project root or absolute) |
| `type` | yes | `"real"` or `"fixture"` |
| `description` | no | Human-readable description |

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `benchmarks/repos.json` | Path to config file |
| `--repo` | (all) | Filter to a single repo by name |
| `--command` | (all) | Filter to a single command (e.g., `analyze`, `impact`, `depgraph`) |
| `--discover` | (none) | Auto-discover repos from a directory |
| `--terrain` | (auto-detect) | Path to terrain binary |
| `--output` | `benchmarks/output` | Output directory for results |
| `--timeout` | `120` | Per-command timeout in seconds |
| `--sequential` | `false` | Run repos sequentially instead of in parallel |
| `--rebuild` | `true` | Rebuild terrain binary from source before running |
| `--heavy-timeout-factor` | `4.0` | Timeout multiplier for heavy commands (insights, debug) |
| `--stdout-capture-kb` | `512` | Max stdout to retain per command (KB) |
| `--stderr-capture-kb` | `128` | Max stderr to retain per command (KB) |

## Commands Benchmarked

### Primary Commands

| Benchmark Name | Actual Command | Notes |
|---------------|----------------|-------|
| `analyze` | `terrain analyze --json` | Runs in repo root |
| `impact` | `terrain impact --json --base HEAD~1` | Requires git repo |
| `insights` | `terrain insights --json` | Falls back to `summary --json` if unavailable |
| `explain` | `terrain explain <id> --json` | Multi-step: runs analyze/impact to find a test ID first |

### Debug Commands (auto-detected)

| Benchmark Name | Actual Command |
|---------------|----------------|
| `debug:graph` | `terrain debug graph --json` |
| `debug:coverage` | `terrain debug coverage --json` |
| `debug:fanout` | `terrain debug fanout --json` |
| `debug:duplicates` | `terrain debug duplicates --json` |

Debug commands use the native `terrain debug` namespace if available, falling back to `terrain depgraph --show <view>` for older binaries.

### Timeout Scaling

Heavy commands get scaled timeouts to avoid false timeout failures:

| Command | Timeout Factor |
|---------|---------------|
| `analyze`, `impact` | 1.0x (base timeout) |
| `explain` | 1.5x |
| `debug:*`, `depgraph:*` | heavy-timeout-factor (default 4.0x) |
| `insights` | heavy-timeout-factor × 1.5 (default 6.0x) |

With `--timeout 120`, insights gets 720s (12 minutes) by default.

## Output Capture

Command output is bounded to prevent memory and artifact bloat on large repos:

- **stdout**: 512 KB by default (configurable with `--stdout-capture-kb`)
- **stderr**: 128 KB by default (configurable with `--stderr-capture-kb`)

When output exceeds the limit, it is truncated and annotated:

```
...[output truncated: captured 524288 of 2097152 bytes]...
```

Truncation is tracked in the results JSON via `stdoutTruncated`, `stderrTruncated`, `stdoutBytes`, and `stderrBytes` fields.

## Assessment & Credibility Scoring

Each command execution is assessed using deterministic heuristics (no LLM involved).

### Scoring Model

| Condition | Points |
|-----------|--------|
| Command succeeded (exit 0) | +20 |
| Output non-empty | +10 |
| Parsed as valid JSON | +10 |
| Each expected section found | +10 |
| Stderr warning line | -5 |
| Trivially short output (<20 chars) | -10 |
| Empty output | -20 |
| Command timed out | -20 |
| Crash or exception | -50 |

Score is clamped to [0, 100].

### Expected Sections by Command

**analyze:** tests detected, repo profile, coverage confidence, duplicates, fanout, weak coverage

**impact:** changed files, impacted tests, coverage confidence, risk signal, dependency reasoning

**insights:** duplicate clusters, high-fanout nodes, weak coverage, recommendations

**explain:** test identifier, test location, confidence, framework

**debug:\*:** node/edge counts, coverage analysis, fanout analysis, duplicate analysis (per sub-command)

### Interpreting Credibility Scores

| Score Range | Meaning |
|------------|---------|
| 80–100 | Excellent — command produced rich, well-structured output |
| 60–79 | Good — output is useful but missing some expected sections |
| 40–59 | Fair — output exists but is incomplete or partially useful |
| 20–39 | Poor — significant issues (missing sections, warnings) |
| 0–19 | Bad — command failed or produced unusable output |

## Output Artifacts

Results are written to `benchmarks/output/`:

| File | Format | Contents |
|------|--------|----------|
| `cli-benchmark-results.json` | JSON | Raw command results (stdout, stderr, exit code, runtime, truncation) |
| `cli-benchmark-assessment.json` | JSON | Per-command assessments with credibility scores |
| `cli-benchmark-summary.md` | Markdown | Human-readable summary with cross-repo analysis |

### Example Summary Output

```markdown
## Repo: terrain

### analyze
- success: yes
- runtime: 3.2s
- credibility: 82
- notes:
  - command succeeded
  - tests detected present
  - repo profile present
  - weak coverage missing
```

The cross-repo summary includes:
- **Command Strength Ranking** — average credibility per command
- **Frequently Missing Sections** — which output sections are most often absent
- **Repo Quality Ranking** — overall scores per repo

## Architecture

```
cmd/terrain-bench/
└── main.go                # CLI entry point: flag parsing, binary detection,
                           # repo loading, parallel/sequential orchestration

internal/benchmark/
├── config.go              # Repo type, BenchConfig, LoadBenchmarkRepos
├── repo_loader.go         # ResolveRepo, DiscoverRepos, detectLanguages
├── command_executor.go    # CommandSpec, CommandResult, DetectCommands,
│                          # RunCommand, RunExplain, bounded output capture,
│                          # per-command timeout scaling
├── runner.go              # BenchResult, RunBenchmark orchestration
├── assessment.go          # CommandAssessment, RepoAssessment, scoring heuristics
├── report_writer.go       # WriteResults, GenerateSummary (JSON + Markdown)
├── export.go              # Privacy-safe benchmark export model
├── bench_test.go          # Unit tests
└── export_test.go         # Export model tests

benchmarks/
├── repos.json             # Repository configuration
├── run-cli-benchmarks.sh  # Shell wrapper (delegates to cmd/terrain-bench/)
└── output/                # Generated output
```

## Running Tests

```bash
go test ./internal/benchmark/ -v
```

Tests cover: config loading, repo resolution, language detection, auto-discovery, assessment heuristics (success, failure, empty output, truncation, timeout, section detection, credibility clamping), summary generation, test ID extraction (4 JSON shapes), report writing, bounded output capture, timeout factor scaling, and string truncation.
