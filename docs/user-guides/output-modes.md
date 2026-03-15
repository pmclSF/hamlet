# Output Modes

Terrain supports multiple output modes across all commands.

## Human-readable (default)

All commands produce formatted terminal output by default. This output is:
- Scannable with clear sections and headings
- Copy-paste friendly (no ANSI color codes)
- Concise — key findings first, details below
- Includes "next steps" hints

```bash
terrain analyze
terrain summary
terrain posture
terrain portfolio
```

## JSON (--json)

All commands support `--json` for machine-readable output:

```bash
terrain analyze --json     # Full TestSuiteSnapshot
terrain summary --json     # ExecutiveSummary object
terrain posture --json     # MeasurementSnapshot object
terrain metrics --json     # Metrics Snapshot
terrain portfolio --json   # PortfolioSnapshot object
terrain compare --json     # SnapshotComparison object
```

JSON output is stable and suitable for:
- CI pipeline integration
- Custom dashboards
- Programmatic analysis
- Diff/comparison tooling

## Privacy-safe export

```bash
terrain export benchmark
```

Produces a benchmark-safe JSON artifact with only:
- Aggregate counts and ratios
- Qualitative bands (strong/moderate/weak)
- Segmentation tags (language, framework, suite size)
- Posture bands per dimension
- Portfolio intelligence bands (redundancy, overbreadth, leverage)

No raw file paths, symbol names, or source code.
