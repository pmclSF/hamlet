# Impact Analysis and Test Selection

## Stage 128 -- User Guide

### Overview

Terrain's impact analysis commands help you understand which tests are affected by a code change and select the right tests to run. This guide covers the `terrain impact` and `terrain select-tests` commands.

---

### The `terrain impact` Command

Analyze the impact of code changes on your test suite.

```bash
terrain impact [flags]
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
terrain impact

# Analyze against a specific base
terrain impact --base develop

# JSON output for CI consumption
terrain impact --json

# Filter to a specific owner
terrain impact --owner @backend-team
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
terrain impact --show units

# Find protection gaps
terrain impact --show gaps

# See the recommended test set
terrain impact --show selected
```

---

### The `terrain select-tests` Command

Output a list of test files to run based on the current diff. This is the action-oriented counterpart to `terrain impact --show selected`.

```bash
terrain select-tests [flags]
```

```bash
# List test files to run
terrain select-tests

# Output as newline-delimited paths (for xargs / CI)
terrain select-tests --format paths

# Output as JSON array
terrain select-tests --json

# Against a specific base
terrain select-tests --base develop
```

#### CI Integration

```bash
# Run only impacted tests with Jest
terrain select-tests --format paths | xargs npx jest

# Run only impacted tests with Go
terrain select-tests --format paths | xargs go test
```

---

### The `terrain pr` Command

Generate PR-oriented impact reports suitable for GitHub comments or CI annotations.

```bash
terrain pr [flags]
```

```bash
# Markdown summary for PR comment
terrain pr --format markdown

# CI annotations (GitHub Actions format)
terrain pr --format annotation

# JSON for custom integrations
terrain pr --format json
```

See the [PR Impact Workflow Guide](pr-impact-workflow.md) for detailed CI integration examples.

---

### Example Workflows

#### Local PR Review

```bash
# 1. Check overall impact
terrain impact

# 2. Drill into gaps
terrain impact --show gaps

# 3. Run the protective test set
terrain select-tests --format paths | xargs npx jest
```

#### CI Integration

```bash
# In your CI pipeline:
# 1. Generate impact report
terrain impact --json > impact.json

# 2. Post PR comment
terrain pr --format markdown > comment.md

# 3. Run selective tests
terrain select-tests --format paths | xargs npx jest --ci
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
