# Review and CI Output Modes

Terrain provides output formats designed for code review and CI integration.

## PR Summary Markdown

Generate a markdown summary for GitHub PR comments:

```bash
terrain pr --format markdown
```

Output includes a posture badge, stats table, findings, recommended tests, and affected owners in clean markdown.

## Concise Review Comment

Generate a one-line summary for inline PR comments:

```bash
terrain pr --format comment
```

Example output:
```
[WARN] Terrain: 3 file(s) changed, 2 unit(s) impacted, 1 gap(s), 2 test(s) recommended. Posture: partially_protected.
  - 1 high-severity finding(s) require attention
  - Run: src/__tests__/auth.test.js, src/__tests__/payment.test.js
```

## CI Annotations

Generate GitHub Actions-compatible annotation output:

```bash
terrain pr --format annotation
```

Example output:
```
::error file=src/feature.js::Exported NewFeature has no test coverage.
::warning file=src/auth.js::[weakAssertion] Low assertion density in auth tests.
```

## CI Integration Example

### GitHub Actions

```yaml
- name: Terrain PR Analysis
  run: |
    terrain pr --format annotation
    terrain pr --format markdown >> $GITHUB_STEP_SUMMARY
```

### Generic CI

```bash
# Run analysis and capture exit code
terrain pr --json > terrain-pr.json

# Use jq to check posture
POSTURE=$(jq -r '.postureBand' terrain-pr.json)
if [ "$POSTURE" = "high_risk" ]; then
  echo "High-risk change detected"
  exit 1
fi
```

## JSON Output

For programmatic CI integration:

```bash
terrain pr --json | jq '.postureBand'
terrain pr --json | jq '.newFindings[] | select(.severity == "high")'
```
