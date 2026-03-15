# Demo Walkthrough

This guide walks through Terrain's core commands on a real repository.

## Prerequisites

```bash
# Build Terrain
go build -o terrain ./cmd/terrain

# Or use go run
alias terrain="go run ./cmd/terrain"
```

## 1. Analyze

Run a full analysis of the current repository:

```bash
terrain analyze
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
terrain analyze --json
```

## 2. Executive Summary

Get a leadership-oriented summary:

```bash
terrain summary
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
terrain summary --json
```

## 3. Metrics

Get an aggregate metrics scorecard:

```bash
terrain metrics
```

This shows privacy-safe aggregate counts and ratios across structure, health, quality, change readiness, governance, and risk.

## 4. Snapshot and Compare

Save a snapshot for trend tracking:

```bash
terrain analyze --write-snapshot
```

Make changes, then save another:

```bash
terrain analyze --write-snapshot
```

Compare the two most recent snapshots:

```bash
terrain compare
```

This shows signal count changes, risk band changes, framework changes, and representative new/resolved findings.

## 5. Policy Check

Create a policy file at `.terrain/policy.yaml`:

```yaml
rules:
  disallow_skipped_tests: true
  max_weak_assertions: 10
```

Then check compliance:

```bash
terrain policy check
```

Exit code 0 means pass, 1 means violations found. Use `--json` for CI integration.

## 6. Migration Workflow

Assess migration readiness and inspect blockers:

```bash
# Overall readiness assessment
terrain migration readiness

# List specific blockers by type and area
terrain migration blockers

# Preview migration difficulty for a single file
terrain migration preview --file test/auth/login.test.js

# Preview migration difficulty across a directory
terrain migration preview --scope test/
```

The migration workflow bridges old Terrain pain ("how hard will this migration be?") with current intelligence ("where is the risk and what should we fix first?").

## 7. Impact Analysis

See what your recent changes affect:

```bash
# Impact against the last commit
terrain impact

# Impact against a specific base ref
terrain impact --base main

# Drill down into specific views
terrain impact --show gaps     # untested changed code
terrain impact --show owners   # impact by team
```

## 8. Benchmark Export

Export a benchmark-safe artifact:

```bash
terrain export benchmark
```

This outputs JSON with aggregate metrics and segmentation tags — no raw file paths or source code. Designed for future cross-repo comparison.

## Quick Demo

Run all major commands in sequence:

```bash
make demo
```

Or manually:

```bash
terrain analyze
terrain summary
terrain metrics
terrain migration readiness
terrain policy check
```

## Migration Journey

A complete migration workflow from assessment to verification:

```bash
# 1. Understand what you're working with
terrain analyze

# 2. Check migration readiness
terrain migration readiness

# 3. Inspect blockers
terrain migration blockers

# 4. Preview specific files
terrain migration preview --file test/auth/login.test.js

# 5. Save a baseline snapshot
terrain analyze --write-snapshot

# 6. Fix blockers and quality issues...

# 7. Save another snapshot and compare
terrain analyze --write-snapshot
terrain compare

# 8. Verify governance compliance
terrain policy check
```

## Sample Outputs

See `examples/sample-reports/` for example outputs:
- `executive-summary.txt` — human-readable executive summary
- `executive-summary.json` — JSON executive summary
- `migration-readiness.txt` — migration readiness assessment
- `migration-preview.txt` — file-level migration preview
- `migration-preview.json` — JSON migration preview
- `policy-check-pass.txt` — policy check with no violations
- `policy-check-fail.txt` — policy check with violations
