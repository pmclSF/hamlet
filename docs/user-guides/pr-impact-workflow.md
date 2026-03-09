# PR Impact Workflow

## Stage 128 -- PR Integration Guide

### Overview

This guide walks through integrating Hamlet's impact analysis into your pull request workflow, from local review to automated CI comments and annotations.

---

### Step-by-Step PR Impact Analysis

#### 1. Analyze the Diff

Before opening or reviewing a PR, run impact analysis against the target branch:

```bash
hamlet impact --base main
```

This produces a summary showing changed files, impacted units, protective tests, and gaps.

#### 2. Review Protection Gaps

Drill into gaps to understand where the change lacks test coverage:

```bash
hamlet impact --show gaps --base main
```

#### 3. Run Protective Tests

Execute only the tests that exercise the changed code:

```bash
hamlet select-tests --base main --format paths | xargs npx jest
```

#### 4. Generate a PR Report

Create a markdown summary suitable for posting as a PR comment:

```bash
hamlet pr --format markdown --base main
```

---

### GitHub Actions Integration

Add Hamlet impact analysis to your CI pipeline with a GitHub Actions workflow:

```yaml
name: PR Impact Analysis

on:
  pull_request:
    branches: [main, develop]

jobs:
  impact:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Full history for accurate diff

      - name: Set up Hamlet
        run: npm install -g hamlet-cli

      - name: Run impact analysis
        run: hamlet impact --base ${{ github.base_ref }} --json > impact.json

      - name: Post PR comment
        run: |
          hamlet pr --format markdown --base ${{ github.base_ref }} > comment.md
          gh pr comment ${{ github.event.pull_request.number }} --body-file comment.md
        env:
          GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}

      - name: Add CI annotations
        run: hamlet pr --format annotation --base ${{ github.base_ref }}

      - name: Run selective tests
        run: |
          hamlet select-tests --base ${{ github.base_ref }} --format paths | \
            xargs npx jest --ci --reporters=default
```

---

### PR Comment Format (`--format markdown`)

The `hamlet pr --format markdown` command produces output like:

```markdown
## Hamlet Impact Analysis

**Change-risk posture:** MODERATE

| Metric | Count |
|--------|-------|
| Changed files | 8 |
| Impacted units | 14 |
| Protective tests | 11 |
| Protection gaps | 3 |

### Protection Gaps

- `src/auth/tokenValidator.js :: validateRefreshToken` -- no mapped tests
- `src/auth/sessionStore.js :: clearExpired` -- no mapped tests
- `src/middleware/rateLimit.js :: checkBurst` -- no mapped tests

### Recommended Test Set

11 tests selected. Run with:
  hamlet select-tests --format paths | xargs npx jest
```

This comment gives reviewers an immediate picture of change risk and where to focus attention.

---

### CI Annotation Format (`--format annotation`)

The `hamlet pr --format annotation` command emits GitHub Actions annotation commands:

```
::warning file=src/auth/tokenValidator.js,line=42::Protection gap: validateRefreshToken has no mapped tests
::warning file=src/auth/sessionStore.js,line=15::Protection gap: clearExpired has no mapped tests
::notice file=test/auth/login.test.js,line=1::Protective test: exercises 3 impacted units (exact)
```

These annotations appear inline on the PR diff, highlighting gaps and protective tests directly in context.

---

### JSON Format (`--format json`)

For custom integrations, use JSON output:

```bash
hamlet pr --format json --base main
```

The JSON structure includes:

```json
{
  "posture": "MODERATE",
  "changed_files": 8,
  "impacted_units": 14,
  "protective_tests": 11,
  "gaps": 3,
  "gap_details": [...],
  "test_details": [...],
  "confidence": { "exact": 9, "inferred": 5 }
}
```

Use this to build custom dashboards, Slack notifications, or gate logic.

---

### Combining With Existing CI

Impact analysis works alongside your existing test pipeline. A typical pattern:

1. **Always run**: `hamlet pr --format annotation` for inline feedback.
2. **Selective tests**: `hamlet select-tests` for fast feedback on impacted tests.
3. **Full suite**: run the complete test suite as a separate job for final merge gating.

This gives developers fast, targeted feedback without sacrificing the safety of a full test run before merge.

---

### Threshold-Based Gating

Use JSON output to implement custom merge gates:

```bash
POSTURE=$(hamlet impact --json --base main | jq -r '.posture')
if [ "$POSTURE" = "CRITICAL" ]; then
  echo "::error::Change-risk posture is CRITICAL. Review required."
  exit 1
fi
```
