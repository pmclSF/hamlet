# Advanced Test Intelligence Demo Flow

This document walks through an end-to-end story using Hamlet's advanced assessment and workflow features. It demonstrates how a developer or tech lead would use the CLI to understand test quality, investigate specific areas, and integrate Hamlet into a PR workflow.

## Prerequisites

Build the Hamlet CLI:

```bash
cd hamlet
go build -o hamlet ./cmd/hamlet
```

For the richest output, prepare a repository with:
- Mixed test quality (some well-tested modules, some gaps)
- Runtime artifacts (JUnit XML or Jest JSON) for stability and failure data
- A CODEOWNERS file for ownership attribution
- At least one prior snapshot (from `hamlet analyze --write-snapshot`) for trend tracking

## Step 1: Analyze a Repository with Mixed Quality

Start with a full analysis to see the lay of the land:

```bash
hamlet analyze --root /path/to/your-repo
```

Expected output (abbreviated):

```
Hamlet — Test Suite Analysis
========================================

Repository: your-repo
Test files: 142    Code units: 387    Frameworks: jest, playwright

Signals (47)
  [HIGH] weakAssertion — src/__tests__/utils.test.js
    Weak assertion density (0.8/test) in file with 12 tests.
  [HIGH] untestedExport — src/services/AuthService.js
    Exported function "validateToken" has no covering test.
  [MEDIUM] flakyTest — src/__tests__/api/orders.test.js
    Retry pattern detected: test uses retry wrapper with 3 attempts.
  [MEDIUM] mockHeavy — src/__tests__/db/connection.test.js
    14 mocks vs 3 assertions in test file.
  [LOW] skippedTest — src/__tests__/legacy/old-parser.test.js
    Test file contains xdescribe/xit blocks.
  ... and 42 more signals
```

This gives a top-level view. Signals are sorted by severity and cover the full spectrum: assertion quality, untested exports, flaky behavior, mock density, and suppressed tests.

To enrich the analysis with runtime data:

```bash
hamlet analyze --root /path/to/your-repo \
  --runtime test-results/junit.xml \
  --coverage coverage/lcov.info
```

Runtime artifacts enable higher-confidence signals for flaky tests, slow tests, and pass-rate-based suppression detection.

## Step 2: PR/Change-Scoped Analysis

Switch to a feature branch and see what your changes mean for test quality:

```bash
hamlet pr --base origin/main
```

Expected output:

```
Hamlet -- Change-Scoped Analysis
========================================

Posture:   PARTIALLY_PROTECTED
Files:     8 changed (5 source, 3 test)
Units:     12 impacted
Gaps:      2

Findings
----------------------------------------
  [HIGH] src/services/PaymentService.js -- Exported function "processRefund" has no test coverage.
    Action: Add unit tests for processRefund before merging.
  [MEDIUM] src/__tests__/api/orders.test.js -- [flakyTest] Retry pattern detected: test uses retry wrapper with 3 attempts.

Recommended Tests
----------------------------------------
  src/__tests__/services/payment.test.js
  src/__tests__/api/orders.test.js
  src/__tests__/integration/checkout.test.js

Affected Owners: team-payments, team-platform
```

The PR analysis tells you:
- **Posture band**: how well-protected is the changed code?
- **Findings**: what specific risks exist in the changed area?
- **Recommended tests**: which tests should you run to validate this change?
- **Affected owners**: which teams own the impacted code?

## Step 3: Drill into an Owner

Investigate what team-platform owns and what signals affect them:

```bash
hamlet show owner team-platform
```

Expected output:

```
Owner: team-platform
Owned files: 34
Test files: 18
Signals: 12

Owned files:
  src/platform/auth.js
  src/platform/config.js
  src/platform/middleware.js
  src/platform/session.js
  ... and 30 more

Top signals:
  [HIGH] weakAssertion -- src/__tests__/platform/auth.test.js
  [HIGH] untestedExport -- src/platform/middleware.js
  [MEDIUM] mockHeavy -- src/__tests__/platform/session.test.js
  [MEDIUM] flakyTest -- src/__tests__/platform/config.test.js

Next: hamlet show test <path>   drill into a specific test file
```

This shows the team's test quality portfolio at a glance: 34 owned files, 18 test files, and 12 signals to address. The "Next:" hint guides the user toward deeper investigation.

## Step 4: Inspect a Specific Test

Follow the drill-down hint to inspect a test file:

```bash
hamlet show test src/__tests__/auth.test.js
```

Expected output:

```
Test File: src/__tests__/auth.test.js
Framework: jest
Owner: team-platform
Tests: 8    Assertions: 24
Mocks: 2
Runtime: 340ms    Pass rate: 100%    Retry rate: 0%
Covers: src/platform/auth.js:validateToken, src/platform/auth.js:createSession

Signals (1):
  [LOW] weakAssertion: moderate assertion density (3.0/test)

Next: hamlet impact --show tests   see impact analysis
```

This gives the full picture for a single test file: what it covers, how it performs at runtime, and what signals affect it. The assertion density of 3.0/test is on the border of moderate, which is why it triggered a low-severity signal.

## Step 5: Inspect a Code Unit

Check coverage for a specific code unit:

```bash
hamlet show unit AuthService
```

Expected output:

```
Code Unit: AuthService
Path: src/platform/auth.js
Kind: class
Exported: true
Owner: team-platform

Covering tests (2):
  src/__tests__/platform/auth.test.js
  src/__tests__/integration/auth-flow.test.js

Next: hamlet show test <path>   drill into a covering test
```

This confirms the code unit is covered by two test files. If no covering tests were detected, the output would say "No covering tests detected." -- a clear signal that tests need to be added.

## Step 6: Understanding Advanced Assessment Insights

The advanced assessment subsystems (lifecycle, stability, assertion, clustering, suppression, failure taxonomy, environment depth) produce insights that surface across multiple commands. Here is how each would appear in enriched outputs:

### Lifecycle continuity (in `hamlet compare`)

When comparing snapshots, lifecycle analysis tracks tests across renames and moves:

```
Lifecycle Continuity
  Exact matches: 128
  Likely renames: 3
    auth.test.js:validateUser -> auth.test.js:validateUserCredentials (confidence: 0.85)
  Likely moves: 1
    __tests__/old/parser.test.js -> __tests__/legacy/parser.test.js (confidence: 0.90)
  Added: 5
  Removed: 2
```

### Stability classes (in `hamlet compare`, `hamlet summary`)

With 3+ snapshots, stability classification identifies problem patterns:

```
Stability Classification (depth: 5 snapshots)
  Consistently stable: 120
  Newly unstable: 2
    src/__tests__/api/orders.test.js -- was stable, recently started failing
  Chronically flaky: 3
    src/__tests__/platform/config.test.js -- flaky signals in 4/5 snapshots
  Intermittently slow: 1
  Data insufficient: 16
```

### Assertion strength (in `hamlet posture`, `hamlet show test`)

Assertion assessment reveals which test files have meaningful verification:

```
Assertion Strength
  Strong (>=3.0/test, low mocks): 45 files
  Moderate (1.5-3.0/test): 62 files
  Weak (<1.0/test or mock-heavy): 18 files
  Unclear (no tests detected): 5 files
  Average density: 2.8 assertions/test
```

### Common-cause clustering (in `hamlet portfolio`)

Clustering identifies shared root causes for broad problems:

```
Common-Cause Clusters
  [shared_import_dependency] src/db/pool.js
    8 test files depend on this module. Changes or instability here impact 8 tests.
    Confidence: 0.72
  [dominant_flaky_fixture] src/test-utils/mockServer.js
    4 flaky tests share this dependency. Candidate root cause for non-deterministic behavior.
    Confidence: 0.65
```

### Suppression detection (in `hamlet analyze`, `hamlet posture`)

Suppression detection identifies tests that are quarantined, skipped, or masked by retry wrappers:

```
Suppression Detection
  Quarantined: 2 (chronic)
  Skip/disable: 5 (3 chronic, 2 unknown intent)
  Retry wrapper: 3 (1 chronic, 2 unknown intent)
  Expected failure: 1 (chronic)
```

### Environment depth (in `hamlet posture`)

Environment depth classifies how realistic each test's execution environment is:

```
Environment Depth
  Browser runtime: 12 files (playwright, cypress)
  Real dependency usage: 8 files
  Moderate mocking: 45 files
  Heavy mocking: 15 files
  Unknown: 4 files
```

## Step 7: Generate a PR Comment

For CI integration, generate a markdown-formatted PR comment:

```bash
hamlet pr --base origin/main --format markdown
```

This produces output suitable for posting as a GitHub PR comment:

```markdown
## Hamlet -- Change Analysis

**Posture:** [WARN] PARTIALLY_PROTECTED

| Metric | Count |
|--------|-------|
| Changed files | 8 |
| Changed source files | 5 |
| Changed test files | 3 |
| Impacted code units | 12 |
| Protection gaps | 2 |

### Findings

- [HIGH] **protection_gap** `src/services/PaymentService.js`: Exported function "processRefund" has no test coverage.
  - Action: Add unit tests for processRefund before merging.
- [MED] **existing_signal** `src/__tests__/api/orders.test.js`: [flakyTest] Retry pattern detected

### Recommended Tests

- `src/__tests__/services/payment.test.js`
- `src/__tests__/api/orders.test.js`
- `src/__tests__/integration/checkout.test.js`

### Affected Owners

- team-payments
- team-platform

---
*Generated by [Hamlet](https://github.com/pmclSF/hamlet) -- signal-first test intelligence*
```

### Other output formats

For CI annotations (GitHub Actions `::error`/`::warning` format):

```bash
hamlet pr --base origin/main --format annotation
```

Output:

```
::error file=src/services/PaymentService.js::Exported function "processRefund" has no test coverage.
::warning file=src/__tests__/api/orders.test.js::[flakyTest] Retry pattern detected
```

For a concise one-liner:

```bash
hamlet pr --base origin/main --format comment
```

Output:

```
[WARN] **Hamlet:** 8 file(s) changed, 12 unit(s) impacted, 2 gap(s), 3 test(s) recommended. Posture: partially_protected.
  - 1 high-severity finding(s) require attention
  - Run: src/__tests__/services/payment.test.js, src/__tests__/api/orders.test.js, src/__tests__/integration/checkout.test.js
```

For JSON (programmatic consumption):

```bash
hamlet pr --base origin/main --json
```

## CI Integration Example

Add Hamlet to your GitHub Actions workflow:

```yaml
name: Hamlet PR Analysis
on: [pull_request]

jobs:
  hamlet:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Full history needed for git diff

      - name: Build Hamlet
        run: go build -o hamlet ./cmd/hamlet

      - name: Hamlet CI Annotations
        run: ./hamlet pr --base origin/${{ github.base_ref }} --format annotation
        continue-on-error: true

      - name: Hamlet PR Comment
        if: github.event_name == 'pull_request'
        run: |
          ./hamlet pr --base origin/${{ github.base_ref }} --format markdown > /tmp/hamlet-comment.md
          gh pr comment ${{ github.event.number }} --body-file /tmp/hamlet-comment.md
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Command Reference

| Command | Purpose |
|---------|---------|
| `hamlet analyze` | Full test suite analysis |
| `hamlet pr --base REF` | PR/change-scoped analysis |
| `hamlet pr --format markdown` | PR comment in markdown |
| `hamlet pr --format annotation` | CI annotation output |
| `hamlet pr --format comment` | Concise one-line summary |
| `hamlet show test <path>` | Drill into a test file |
| `hamlet show unit <name>` | Drill into a code unit |
| `hamlet show owner <name>` | Drill into an owner's portfolio |
| `hamlet show finding <type>` | Drill into a finding |
| `hamlet impact --show units` | Impact: impacted code units |
| `hamlet impact --show gaps` | Impact: protection gaps |
| `hamlet impact --show tests` | Impact: recommended tests |
| `hamlet impact --show owners` | Impact: affected owners |
| `hamlet impact --owner NAME` | Impact: filter by owner |
| `hamlet summary` | Executive summary with trends |
| `hamlet posture` | Detailed posture with evidence |
| `hamlet portfolio` | Cost, leverage, redundancy analysis |
| `hamlet compare` | Snapshot comparison with trends |
