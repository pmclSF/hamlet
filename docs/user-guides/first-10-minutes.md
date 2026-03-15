# Your First 10 Minutes with Terrain

## Minute 0-2: See what Terrain finds

```bash
cd your-repo
terrain analyze
```

Scan the output for surprises. Common first reactions:
- "I didn't know we had 5 different test frameworks."
- "40% of our exported functions have no tests?"
- "All our flaky tests are in one directory."

## Minute 2-4: Get the leadership view

```bash
terrain summary
```

This produces a concise overview with:
- Overall posture
- Key numbers
- Top risk areas
- Dominant signal types
- Recommended focus

This output is designed to be paste-ready for Slack, PRs, or team updates.

## Minute 4-6: Understand the evidence

```bash
terrain posture
```

This breaks down each posture dimension with individual measurements, evidence strength, and limitations. If a dimension is rated "weak," you can see exactly which measurements drove that assessment.

## Minute 6-7: See test cost and leverage

```bash
terrain portfolio
```

Run `terrain portfolio` to see test cost, leverage, and redundancy insights. The portfolio view treats your test suite as a set of investments — showing which tests deliver the most protection per CI minute and which overlap so heavily they are candidates for consolidation.

## Minute 7-9: Save and track

```bash
terrain analyze --write-snapshot
```

Do this regularly to track trends. After your second snapshot:

```bash
terrain compare
```

You'll see what improved, what worsened, and what stayed the same.

## Minute 9-10: Export and share

```bash
terrain metrics          # aggregate scorecard
terrain export benchmark # privacy-safe export
```

The metrics command gives you a structured scorecard. The benchmark export produces an artifact with only aggregate counts and bands — no file paths, no symbol names, no source code.

## What's next

- Add a `.terrain/policy.yaml` to enforce team standards
- Run `terrain policy check` in CI
- Use `terrain analyze --json` to integrate with other tools
- Share `terrain summary` output in your next retrospective
