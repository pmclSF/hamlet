# Review and CI Output Modes

Hamlet provides output formats designed for code review and CI integration.

## PR Summary Markdown

Generate a markdown summary for GitHub PR comments:

```bash
hamlet pr --format markdown
```

Output includes a posture badge, stats table, findings, recommended tests, and affected owners in clean markdown.

## Concise Review Comment

Generate a one-line summary for inline PR comments:

```bash
hamlet pr --format comment
```

Example output:
```
[WARN] Hamlet: 3 file(s) changed, 2 unit(s) impacted, 1 gap(s), 2 test(s) recommended. Posture: partially_protected.
  - 1 high-severity finding(s) require attention
  - Run: src/__tests__/auth.test.js, src/__tests__/payment.test.js
```

## CI Annotations

Generate GitHub Actions-compatible annotation output:

```bash
hamlet pr --format annotation
```

Example output:
```
::error file=src/feature.js::Exported NewFeature has no test coverage.
::warning file=src/auth.js::[weakAssertion] Low assertion density in auth tests.
```

## CI Integration Example

### GitHub Actions

```yaml
- name: Hamlet PR Analysis
  run: |
    hamlet pr --format annotation
    hamlet pr --format markdown >> $GITHUB_STEP_SUMMARY
```

### Generic CI

```bash
# Run analysis and capture exit code
hamlet pr --json > hamlet-pr.json

# Use jq to check posture
POSTURE=$(jq -r '.postureBand' hamlet-pr.json)
if [ "$POSTURE" = "high_risk" ]; then
  echo "High-risk change detected"
  exit 1
fi
```

## JSON Output

For programmatic CI integration:

```bash
hamlet pr --json | jq '.postureBand'
hamlet pr --json | jq '.newFindings[] | select(.severity == "high")'
```
