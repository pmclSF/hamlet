# Impact Analysis and Test Selection

## Stage 128 -- User Guide

### Overview

Hamlet's impact analysis commands help you understand which tests are affected by a code change and select the right tests to run. This guide covers the `hamlet impact` and `hamlet select-tests` commands.

---

### The `hamlet impact` Command

Analyze the impact of code changes on your test suite.

```bash
hamlet impact [flags]
```

#### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--root` | `.` | Repository root directory |
| `--base` | `main` | Base branch or commit to diff against |
| `--json` | `false` | Output structured JSON instead of human-readable text |
| `--show` | `summary` | Drill-down view (see below) |
| `--owner` | (none) | Filter results to a specific code owner |

#### Basic Usage

```bash
# Analyze impact of current branch vs main
hamlet impact

# Analyze against a specific base
hamlet impact --base develop

# JSON output for CI consumption
hamlet impact --json

# Filter to a specific owner
hamlet impact --owner @backend-team
```

---

### Drill-Down Views

The `--show` flag selects a specific view of the impact data.

| View | Description |
|------|-------------|
| `summary` | High-level change-risk posture, unit count, gap count, test count |
| `units` | List of impacted code units (functions, classes, modules) with their protection status |
| `gaps` | Changed code units that lack any mapped test |
| `tests` | Tests that exercise the changed code, with confidence levels |
| `owners` | Impact breakdown by code owner |
| `graph` | Dependency graph showing change-to-test paths |
| `selected` | The protective test set recommended for execution |

```bash
# See which code units are affected
hamlet impact --show units

# Find protection gaps
hamlet impact --show gaps

# See the recommended test set
hamlet impact --show selected
```

---

### The `hamlet select-tests` Command

Output a list of test files to run based on the current diff. This is the action-oriented counterpart to `hamlet impact --show selected`.

```bash
hamlet select-tests [flags]
```

```bash
# List test files to run
hamlet select-tests

# Output as newline-delimited paths (for xargs / CI)
hamlet select-tests --format paths

# Output as JSON array
hamlet select-tests --json

# Against a specific base
hamlet select-tests --base develop
```

#### CI Integration

```bash
# Run only impacted tests with Jest
hamlet select-tests --format paths | xargs npx jest

# Run only impacted tests with Go
hamlet select-tests --format paths | xargs go test
```

---

### The `hamlet pr` Command

Generate PR-oriented impact reports suitable for GitHub comments or CI annotations.

```bash
hamlet pr [flags]
```

```bash
# Markdown summary for PR comment
hamlet pr --format markdown

# CI annotations (GitHub Actions format)
hamlet pr --format annotation

# JSON for custom integrations
hamlet pr --format json
```

See the [PR Impact Workflow Guide](pr-impact-workflow.md) for detailed CI integration examples.

---

### Example Workflows

#### Local PR Review

```bash
# 1. Check overall impact
hamlet impact

# 2. Drill into gaps
hamlet impact --show gaps

# 3. Run the protective test set
hamlet select-tests --format paths | xargs npx jest
```

#### CI Integration

```bash
# In your CI pipeline:
# 1. Generate impact report
hamlet impact --json > impact.json

# 2. Post PR comment
hamlet pr --format markdown > comment.md

# 3. Run selective tests
hamlet select-tests --format paths | xargs npx jest --ci
```

---

### Output Interpretation

#### Summary Output

```
Impact Analysis: feature/add-auth vs main

  Changed files:     8
  Impacted units:   14
  Protective tests: 11
  Protection gaps:   3

  Change-risk posture: MODERATE

  Confidence: 9 exact, 5 inferred
```

- **Changed files**: number of files modified in the diff.
- **Impacted units**: code-level units (functions, classes) within those files.
- **Protective tests**: tests that exercise at least one impacted unit.
- **Protection gaps**: impacted units with no mapped test.
- **Change-risk posture**: overall risk level (LOW, MODERATE, HIGH, CRITICAL) based on gap ratio, change volume, and mapping confidence.
- **Confidence**: how many test-to-code mappings are exact (import-traced) vs. inferred (naming/co-location heuristic).

#### Gap Output

```
Protection Gaps (3 units):

  src/auth/tokenValidator.js :: validateRefreshToken
    No mapped tests. Nearest test: test/auth/tokenValidator.test.js (inferred)

  src/auth/sessionStore.js :: clearExpired
    No mapped tests.

  src/middleware/rateLimit.js :: checkBurst
    No mapped tests. Nearest test: test/middleware/rateLimit.test.js (inferred)
```

Gaps indicate areas where the changed code has no test exercising it. The "nearest test" hint suggests where coverage could be added.

#### Confidence Levels

- **Exact**: the test directly imports or references the changed unit.
- **Inferred**: the mapping is based on naming convention, co-location, or transitive dependency. Useful but may include false positives.

Filter by confidence using the drill-down views or JSON output to focus on high-certainty results.
