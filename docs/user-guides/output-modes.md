# Output Modes

Hamlet supports multiple output modes across all commands.

## Human-readable (default)

All commands produce formatted terminal output by default. This output is:
- Scannable with clear sections and headings
- Copy-paste friendly (no ANSI color codes)
- Concise — key findings first, details below
- Includes "next steps" hints

```bash
hamlet analyze
hamlet summary
hamlet posture
```

## JSON (--json)

All commands support `--json` for machine-readable output:

```bash
hamlet analyze --json     # Full TestSuiteSnapshot
hamlet summary --json     # ExecutiveSummary object
hamlet posture --json     # MeasurementSnapshot object
hamlet metrics --json     # Metrics Snapshot
hamlet compare --json     # SnapshotComparison object
```

JSON output is stable and suitable for:
- CI pipeline integration
- Custom dashboards
- Programmatic analysis
- Diff/comparison tooling

## Privacy-safe export

```bash
hamlet export benchmark
```

Produces a benchmark-safe JSON artifact with only:
- Aggregate counts and ratios
- Qualitative bands (strong/moderate/weak)
- Segmentation tags (language, framework, suite size)
- Posture bands per dimension

No raw file paths, symbol names, or source code.
