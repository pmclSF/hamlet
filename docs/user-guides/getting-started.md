# Getting Started with Terrain

## Install

```bash
go install github.com/pmclSF/terrain/cmd/terrain@latest
```

## First run

Navigate to any repository with tests and run:

```bash
terrain analyze
```

Terrain will discover test files, detect frameworks, emit signals, compute risk surfaces, and produce a posture assessment — all from static analysis.

## Understanding the output

The analyze report shows:

- **Repository** — languages, frameworks, CI systems detected
- **Frameworks** — which test frameworks and how many files each
- **Posture** — five dimensions (health, coverage depth, coverage diversity, structural risk, operational risk) rated strong/moderate/weak
- **Signals** — categorized findings: health, quality, migration, governance
- **Risk** — where risk concentrates by directory or owner

## Next commands

After `analyze`, try:

```bash
terrain summary     # leadership-ready overview
terrain posture     # detailed posture with evidence per measurement
terrain portfolio   # see which tests provide the most value and which waste resources
terrain metrics     # aggregate scorecard
```

## Saving snapshots for trend tracking

```bash
terrain analyze --write-snapshot
```

This saves the snapshot to `.terrain/snapshots/`. After multiple snapshots, compare them:

```bash
terrain compare
```

## JSON output

All commands support `--json` for machine-readable output:

```bash
terrain analyze --json
terrain summary --json
terrain posture --json
```

## Policy enforcement

Create `.terrain/policy.yaml` to define rules, then:

```bash
terrain policy check
```

Returns exit code 2 if violations are found — useful in CI gates.
