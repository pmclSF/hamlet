# Your First 10 Minutes with Hamlet

## Minute 0-2: See what Hamlet finds

```bash
cd your-repo
hamlet analyze
```

Scan the output for surprises. Common first reactions:
- "I didn't know we had 5 different test frameworks."
- "40% of our exported functions have no tests?"
- "All our flaky tests are in one directory."

## Minute 2-4: Get the leadership view

```bash
hamlet summary
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
hamlet posture
```

This breaks down each posture dimension with individual measurements, evidence strength, and limitations. If a dimension is rated "weak," you can see exactly which measurements drove that assessment.

## Minute 6-8: Save and track

```bash
hamlet analyze --write-snapshot
```

Do this regularly to track trends. After your second snapshot:

```bash
hamlet compare
```

You'll see what improved, what worsened, and what stayed the same.

## Minute 8-10: Export and share

```bash
hamlet metrics          # aggregate scorecard
hamlet export benchmark # privacy-safe export
```

The metrics command gives you a structured scorecard. The benchmark export produces an artifact with only aggregate counts and bands — no file paths, no symbol names, no source code.

## What's next

- Add a `.hamlet/policy.yaml` to enforce team standards
- Run `hamlet policy check` in CI
- Use `hamlet analyze --json` to integrate with other tools
- Share `hamlet summary` output in your next retrospective
