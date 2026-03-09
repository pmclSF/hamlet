# Launch Assets Guide

This document lists the demo fixtures, sample workflows, and screenshot recommendations for the Hamlet V3 launch.

## Demo Fixtures

Four fixture files in `fixtures/demos/` simulate realistic repository profiles for demo and screenshot purposes:

| Fixture | File | Profile |
|---------|------|---------|
| Healthy Balanced | `healthy-balanced.json` | 2 frameworks (Jest + Playwright), good coverage, minimal issues -- the "what good looks like" baseline |
| Flaky Concentrated | `flaky-concentrated.json` | Instability concentrated in auth tests, otherwise healthy -- demonstrates risk concentration insight |
| E2E-Heavy Shallow | `e2e-heavy-shallow.json` | Most tests are E2E with weak unit coverage -- demonstrates coverage diversity and portfolio findings |
| Fragmented Migration Risk | `fragmented-migration-risk.json` | Many frameworks, migration blockers compounded by quality issues -- demonstrates migration readiness and structural risk |

Each fixture contains a full snapshot-compatible JSON structure with repository metadata, test files, code units, and signals. They can be used with any reporting command by loading them as snapshots.

## Sample Workflows for Screenshots

### 1. First-Run Analysis

```bash
hamlet analyze --root fixtures/demos/healthy-balanced
```

**What to capture:** The full output showing framework detection, signal summary, posture dimensions at a glance, and portfolio summary. This is the "5 seconds to insight" moment.

### 2. Executive Summary

```bash
hamlet summary --root fixtures/demos/flaky-concentrated
```

**What to capture:** Overall posture, key numbers, top risk areas, prioritized recommendations, blind spots, and benchmark readiness. This is the "paste into your team update" view.

### 3. Posture Drill-Down

```bash
hamlet posture --root fixtures/demos/e2e-heavy-shallow
```

**What to capture:** The health and coverage diversity dimensions with their measurements, evidence strength annotations, explanations, and limitations. This is the "show me the evidence" view.

### 4. Portfolio Intelligence

```bash
hamlet portfolio --root fixtures/demos/fragmented-migration-risk
```

**What to capture:** The overview (asset count, portfolio posture), findings (high-leverage, redundancy, overbroad, low-value), top findings with explanations and suggested actions, and evidence notes.

### 5. Migration Assessment

```bash
hamlet migration readiness --root fixtures/demos/fragmented-migration-risk
```

**What to capture:** Per-area assessments, blocker counts by type, quality cross-references, and actionable guidance.

### 6. Impact Analysis

```bash
# In a repo with recent changes:
hamlet impact --base HEAD~3
hamlet impact --show gaps
```

**What to capture:** Changed files, test coverage of changes, identified gaps, and per-owner breakdown.

## Key Insights to Capture in Screenshots

When selecting which output to screenshot, prioritize frames that show:

1. **Posture dimensions with mixed bands** -- not all-green, not all-red. The flaky-concentrated and e2e-heavy-shallow fixtures produce moderate/weak bands that demonstrate nuance.

2. **Evidence transparency** -- measurement lines that show evidence strength (strong, partial, weak) and limitations. This differentiates Hamlet from tools that just emit scores.

3. **Actionable recommendations** -- the prioritized recommendations section with what/why/where/evidence fields.

4. **Blind spots** -- the known blind spots section, which demonstrates honesty about what the tool cannot see without additional data.

5. **Portfolio findings** -- especially high-leverage and redundancy candidates with explanations. These are the "I didn't know that" moments.

## Tips for Terminal Recordings

- Use a terminal width of 100-120 columns for readable wrapping.
- Use a dark terminal theme with high contrast. The output is plain text (no ANSI color codes), so legibility depends on font size and contrast.
- Record the progressive drill-down flow: `analyze` then `summary` then `posture` then `portfolio`. This demonstrates the information architecture.
- Keep recordings under 60 seconds. Each command produces output in under 2 seconds on typical repositories.
- If recording `hamlet compare`, run two `hamlet analyze --write-snapshot` commands with a file change between them to show trend detection.

## README Screenshot Suggestions

For the project README, consider these three screenshots:

1. **`hamlet summary` output** on the flaky-concentrated fixture -- shows posture bands, key numbers, risk areas, and recommendations in a single compact view.

2. **`hamlet posture` output** on the e2e-heavy-shallow fixture (coverage diversity dimension only) -- shows measurements with evidence and limitations, demonstrating transparency.

3. **`hamlet portfolio` output** on the fragmented-migration-risk fixture -- shows findings with explanations and suggested actions, demonstrating the portfolio intelligence concept.

Crop screenshots to focus on the most information-dense sections. Full output screenshots tend to be too tall for README readability.
