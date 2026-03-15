# First-Run Product Story

## The first five minutes

A developer clones a repo they've inherited, or wants to understand the test health of their own project. They install Terrain and run:

```
terrain analyze
```

### What they see first

A structured report showing:
1. **Repository shape** — languages, frameworks, test file count
2. **Posture** — five dimensions rated strong/moderate/weak with evidence
3. **Signals** — categorized findings (health, quality, migration, governance)
4. **Risk surfaces** — where risk concentrates (directories, owners)
5. **Top findings** — the most important signals with locations and explanations

This is not a wall of text. It is a focused, scannable summary that fits in a terminal window.

### What surprises them

The "wow" moment in the first run comes from structural insight they did not have before:
- "I didn't know 40% of our exported functions have no linked tests."
- "I didn't know our auth tests are all mock-heavy AND have migration blockers."
- "I didn't know 80% of our test suite is E2E."

### What they do next

The analyze output naturally leads to:

```
terrain summary        # leadership-ready overview
terrain posture        # detailed posture with evidence per dimension
terrain focus          # where to look first
```

Each command output includes a "next useful command" hint.

After posture, users can run `terrain portfolio` to see the test suite as a portfolio of investments. The portfolio view reveals which tests provide the most protection per CI minute, which overlap so heavily they are redundancy candidates, and where runtime concentrates in a small number of broad tests. This reframes test quality from "pass/fail" to "cost vs. value."

## The second session

After the initial exploration, the user runs:

```
terrain analyze --write-snapshot
```

This persists the snapshot. On the next run, `terrain compare` shows what changed. The user now has trend tracking with zero infrastructure.

## The team session

The user shares `terrain summary` output in a PR or Slack thread. It is designed to be copy-paste ready — no ANSI codes, no excessive width, clear headings.

A tech lead reads it and immediately sees:
- Overall posture
- Key numbers (test files, frameworks, signals, critical findings)
- Top risk areas
- Recommended focus

## Command flow

```
terrain analyze              → full analysis, the "what"
terrain summary              → leadership view, the "so what"
terrain posture              → evidence detail, the "show me"
terrain focus                → prioritized action, the "now what"
terrain portfolio            → test investment view, the "is it worth it"
terrain metrics              → aggregate scorecard
terrain compare              → trend tracking
terrain export benchmark     → privacy-safe export for future comparison
```

## Design constraints

- No command requires arguments to produce useful output on a repo
- Default output is human-readable; `--json` for machine consumption
- No command takes more than 30 seconds on a 500-file repo
- Error messages suggest what to do, not just what went wrong
