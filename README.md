# Hamlet

**Signal-first test intelligence for engineering teams**

Hamlet analyzes repository structure, test code, runtime artifacts, coverage data, and local policy to surface the real state of your test system — risk, quality, migration readiness, and governance — without running tests.

## Quick Start

```bash
# Install
go install github.com/pmclSF/hamlet/cmd/hamlet@latest

# Or build from source
git clone https://github.com/pmclSF/hamlet.git
cd hamlet
go build -o hamlet ./cmd/hamlet

# Analyze the current repository
hamlet analyze

# Executive summary with risk, trends, and benchmark readiness
hamlet summary

# Aggregate metrics scorecard
hamlet metrics

# JSON output for any command
hamlet analyze --json
hamlet summary --json
hamlet metrics --json
```

### Requirements

- Go 1.23 or later

## Commands

| Command | Description |
|---------|-------------|
| `hamlet analyze` | Analyze repository test suite — frameworks, signals, risk |
| `hamlet summary` | Executive summary — posture, trends, focus, benchmark readiness |
| `hamlet metrics` | Aggregate metrics scorecard (privacy-safe, benchmark-ready) |
| `hamlet compare` | Compare two snapshots and show trend changes |
| `hamlet policy check` | Evaluate repository against local policy rules |
| `hamlet export benchmark` | Output benchmark-safe JSON export for future comparison |

Run `hamlet --help` for full flag documentation.

## What Hamlet Reveals

### Structure
Framework inventory, test file discovery, code-to-test relationships, ownership mapping.

### Health
Flaky tests, slow tests, skipped tests, dead tests, unstable suites.

### Quality
Weak assertions, mock-heavy tests, untested exports, coverage blind spots.

### Migration Readiness
Migration blockers, deprecated patterns, legacy framework drift, framework fragmentation.

### Policy and Governance
Local policy rules, violation tracking, compliance enforcement in CI.

### Risk
Explainable risk surfaces by dimension (reliability, change, speed) with directory and owner concentration.

### Benchmark-Safe Exports
Privacy-safe aggregate metrics for future cross-repo comparison — no raw paths or source code exposed.

## Snapshot Workflow

Hamlet supports local snapshot history for trend tracking:

```bash
# Save a snapshot
hamlet analyze --write-snapshot

# Later, save another snapshot
hamlet analyze --write-snapshot

# Compare the two most recent snapshots
hamlet compare

# Executive summary automatically includes trend highlights
hamlet summary
```

Snapshots are stored in `.hamlet/snapshots/` as timestamped JSON files.

## Policy

Define local policy rules in `.hamlet/policy.yaml`:

```yaml
rules:
  disallow_skipped_tests: true
  max_weak_assertions: 10
  max_mock_heavy_tests: 5
```

Then check compliance:

```bash
hamlet policy check        # human-readable output
hamlet policy check --json # JSON output for CI
```

Exit code 0 = pass, 1 = violations found.

## Architecture

Hamlet is built around a signal-first architecture:

```
Repository scan → Signal detection → Risk modeling → Reporting
```

- **Signals** are the core abstraction — every finding is a structured signal
- **Snapshots** are the canonical serialized artifact (`TestSuiteSnapshot`)
- **Risk surfaces** are derived from signals with explainable scoring
- **Reports** synthesize signals, risk, trends, and benchmark readiness

See [DESIGN.md](DESIGN.md) for architecture overview and [docs/](docs/) for detailed documentation.

## Project Structure (V3 Go Engine)

```
cmd/hamlet/          CLI entry point
internal/
├── analysis/        Repository scanning and test file discovery
├── benchmark/       Benchmark-safe export and segmentation
├── comparison/      Snapshot-to-snapshot trend comparison
├── governance/      Policy evaluation and governance signals
├── heatmap/         Risk concentration model
├── metrics/         Aggregate metrics extraction
├── migration/       Migration detectors and readiness model
├── models/          Canonical data models (Signal, Snapshot, Risk, etc.)
├── ownership/       Ownership resolution (config, CODEOWNERS, directory)
├── policy/          Policy config model and YAML loader
├── quality/         Quality signal detectors
├── reporting/       Human-readable report renderers
├── scoring/         Explainable risk engine
├── signals/         Signal detector interface and runner
└── summary/         Executive summary builder
```

## Legacy Converter Engine

Hamlet originated as a multi-framework test converter (V2, JavaScript ES modules). That engine is preserved in `src/`, `bin/`, and `test/` and remains functional. V3 reframes migration as one dimension of broader test intelligence. See [docs/legacy/](docs/legacy/) for historical architecture docs and [CLAUDE.md](CLAUDE.md) for legacy code conventions.

## Development

```bash
# Build
go build -o hamlet ./cmd/hamlet

# Test all Go packages
go test ./internal/... ./cmd/...

# Test with verbose output
go test -v ./internal/...

# Legacy JavaScript tests (requires Node.js 22+)
npm test
```

## Principles

- Signals are the core abstraction
- Analysis comes before automation
- Risk must be explainable
- Hamlet must be useful locally, without SaaS
- Privacy boundary: aggregate metrics never expose raw paths or source code
- Hamlet measures system health, not individual developer productivity

## Status

Hamlet V3 is in active development. The Go engine implements:
- repository analysis and signal detection
- explainable risk modeling
- local policy and governance
- ownership-aware review and triage
- migration intelligence
- snapshot history and trend comparison
- benchmark-ready metrics
- executive summary reporting

The JSON contract (`TestSuiteSnapshot`) is stabilizing but may evolve.

## License

MIT License — see [LICENSE](LICENSE) for details.
