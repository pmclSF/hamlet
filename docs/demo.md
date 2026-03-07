# Demo Walkthrough

This guide walks through Hamlet's core commands on a real repository.

## Prerequisites

```bash
# Build Hamlet
go build -o hamlet ./cmd/hamlet

# Or use go run
alias hamlet="go run ./cmd/hamlet"
```

## 1. Analyze

Run a full analysis of the current repository:

```bash
hamlet analyze
```

This produces a human-readable report showing:
- detected frameworks and test files
- health signals (flaky, slow, skipped, dead tests)
- quality signals (weak assertions, mock-heavy tests, untested exports)
- migration signals (deprecated patterns, custom matchers, dynamic generation)
- risk surfaces (reliability, change, speed)
- review sections (by owner, directory, type)

For JSON output:

```bash
hamlet analyze --json
```

## 2. Executive Summary

Get a leadership-oriented summary:

```bash
hamlet summary
```

This produces:
- overall posture by dimension
- key numbers
- top risk areas
- trend highlights (if prior snapshots exist)
- dominant signal drivers
- recommended focus
- benchmark readiness

For JSON:

```bash
hamlet summary --json
```

## 3. Metrics

Get an aggregate metrics scorecard:

```bash
hamlet metrics
```

This shows privacy-safe aggregate counts and ratios across structure, health, quality, change readiness, governance, and risk.

## 4. Snapshot and Compare

Save a snapshot for trend tracking:

```bash
hamlet analyze --write-snapshot
```

Make changes, then save another:

```bash
hamlet analyze --write-snapshot
```

Compare the two most recent snapshots:

```bash
hamlet compare
```

This shows signal count changes, risk band changes, framework changes, and representative new/resolved findings.

## 5. Policy Check

Create a policy file at `.hamlet/policy.yaml`:

```yaml
rules:
  disallow_skipped_tests: true
  max_weak_assertions: 10
```

Then check compliance:

```bash
hamlet policy check
```

Exit code 0 means pass, 1 means violations found. Use `--json` for CI integration.

## 6. Benchmark Export

Export a benchmark-safe artifact:

```bash
hamlet export benchmark
```

This outputs JSON with aggregate metrics and segmentation tags — no raw file paths or source code. Designed for future cross-repo comparison.

## Quick Demo

Run all major commands in sequence:

```bash
make demo
```

Or manually:

```bash
hamlet analyze
hamlet summary
hamlet metrics
hamlet policy check
```

## Sample Outputs

See `examples/sample-reports/` for example outputs:
- `executive-summary.txt` — human-readable executive summary
- `executive-summary.json` — JSON executive summary
- `policy-check-pass.txt` — policy check with no violations
- `policy-check-fail.txt` — policy check with violations
