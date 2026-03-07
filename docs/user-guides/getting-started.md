# Getting Started with Hamlet

## Install

```bash
go install github.com/pmclSF/hamlet/cmd/hamlet@latest
```

## First run

Navigate to any repository with tests and run:

```bash
hamlet analyze
```

Hamlet will discover test files, detect frameworks, emit signals, compute risk surfaces, and produce a posture assessment — all from static analysis.

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
hamlet summary     # leadership-ready overview
hamlet posture     # detailed posture with evidence per measurement
hamlet metrics     # aggregate scorecard
```

## Saving snapshots for trend tracking

```bash
hamlet analyze --write-snapshot
```

This saves the snapshot to `.hamlet/snapshots/`. After multiple snapshots, compare them:

```bash
hamlet compare
```

## JSON output

All commands support `--json` for machine-readable output:

```bash
hamlet analyze --json
hamlet summary --json
hamlet posture --json
```

## Policy enforcement

Create `.hamlet/policy.yaml` to define rules, then:

```bash
hamlet policy check
```

Returns exit code 1 if violations are found — useful in CI gates.
