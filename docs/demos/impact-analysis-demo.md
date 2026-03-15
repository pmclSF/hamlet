# Impact Analysis Demo Flow

This document walks through an end-to-end impact analysis workflow. It demonstrates how a developer would use the CLI to understand change risk, identify protection gaps, and select tests.

## Prerequisites

```bash
cd terrain
go build -o terrain ./cmd/terrain
```

## Step 1: Analyze a Repository

Start with a full analysis:

```bash
terrain analyze --root /path/to/your-repo
```

This produces the baseline snapshot with code units, test files, signals, and coverage data.

## Step 2: Run Impact Analysis

After making changes on a branch, run impact analysis against main:

```bash
terrain impact --base origin/main
```

Expected output (abbreviated):

```
Terrain Impact Analysis
============================================================

Summary: 3 file(s) changed, 2 code unit(s) impacted, 2 test(s) relevant, 1 protection gap(s). Posture: partially_protected.

Change-Risk Posture: PARTIALLY_PROTECTED
  This change has partial protection. 1 protection gap(s) identified.
  protection:          partially_protected
  exposure:            partially_protected
  coordination:        well_protected
  instability:         well_protected

Impacted Code Units
------------------------------------------------------------
  AuthService                    modified  strong [exported]
    Owner: team-auth
  validateToken                  modified  none [exported]

Protection Gaps
------------------------------------------------------------
  [high] Exported function validateToken has no observed test coverage.
    Action: Add unit tests for exported function validateToken — this is public API surface.

Recommended Tests (2)
------------------------------------------------------------
  test/auth.test.js [exact]
    covers impacted code unit
  test/integration/auth-flow.test.js [inferred]
    in same directory tree as changed code
```

## Step 3: View Impacted Units

Drill into the code units:

```bash
terrain impact --base origin/main --show units
```

```
Impacted Code Units (2)
============================================================

  AuthService                    modified  protection: strong [exported]
    Owner: team-auth
    Confidence: inferred
    Covering tests: test/auth.test.js
    Path: src/auth/service.js

  validateToken                  modified  protection: none [exported]
    Confidence: inferred
    Path: src/auth/service.js
```

## Step 4: Inspect Protection Gaps

```bash
terrain impact --base origin/main --show gaps
```

```
Protection Gaps (1)
============================================================

  HIGH severity (1)
  ----------------------------------------
    [untested_export] Exported function validateToken has no observed test coverage.
      Path: src/auth/service.js
      Action: Add unit tests for exported function validateToken — this is public API surface.
```

## Step 5: Select Protective Tests

```bash
terrain select-tests --base origin/main
```

```
Protective Test Set
============================================================

  Strategy:   exact
  Tests:      2
  Covered:    1 unit(s)
  Uncovered:  1 unit(s)

  2 test(s) with exact coverage of 1 impacted unit(s). 1 impacted unit(s) have no covering tests in the selected set.

Selected Tests
------------------------------------------------------------
  test/auth.test.js  [exact]
    - exact coverage of impacted unit (src/auth/service.js:AuthService)
  test/integration/auth-flow.test.js  [inferred]
    - inferred structural relationship

Warning: 1 impacted unit(s) have no covering tests in the selected set.
Consider adding tests or running the full suite.
```

## Step 6: View Impact Graph

```bash
terrain impact --base origin/main --show graph
```

```
Impact Graph
============================================================

  Total edges:      3
  Exact edges:      1
  Inferred edges:   1
  Weak edges:       1
  Connected units:  2
  Isolated units:   0
  Connected tests:  2

Edges for impacted units
------------------------------------------------------------
  AuthService
    -> test/auth.test.js                           [exact] exact_coverage
  validateToken
    -> test/auth.test.js                           [weak] name_convention
```

## Step 7: Generate PR Comment

```bash
terrain pr --base origin/main --format markdown
```

Produces GitHub-ready markdown with posture badge, metrics table, findings, recommended tests, and affected owners.

## CI Integration

```yaml
name: Terrain Impact Analysis
on: [pull_request]

jobs:
  terrain:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Build Terrain
        run: go build -o terrain ./cmd/terrain

      - name: Impact Annotations
        run: ./terrain pr --base origin/${{ github.base_ref }} --format annotation
        continue-on-error: true

      - name: Select Tests
        run: ./terrain select-tests --base origin/${{ github.base_ref }} --json > /tmp/tests.json

      - name: PR Comment
        if: github.event_name == 'pull_request'
        run: |
          ./terrain pr --base origin/${{ github.base_ref }} --format markdown > /tmp/comment.md
          gh pr comment ${{ github.event.number }} --body-file /tmp/comment.md
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Command Reference

| Command | Purpose |
|---------|---------|
| `terrain impact` | Full impact analysis |
| `terrain impact --show units` | Impacted code units |
| `terrain impact --show gaps` | Protection gaps |
| `terrain impact --show tests` | Impacted tests |
| `terrain impact --show owners` | Owner impact |
| `terrain impact --show graph` | Impact graph |
| `terrain impact --show selected` | Protective test set |
| `terrain impact --owner NAME` | Filter by owner |
| `terrain select-tests` | Recommend protective tests |
| `terrain pr --format markdown` | PR comment |
| `terrain pr --format annotation` | CI annotations |
