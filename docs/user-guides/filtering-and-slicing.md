# Filtering and Slicing

Hamlet supports filtering and slicing across commands to focus on the relevant part of a large repo.

## Available Filters

### Owner filtering

```bash
hamlet impact --owner team-platform     # impact scoped to one team
hamlet show owner team-platform          # drill into owner details
```

### Path/entity filtering

```bash
hamlet show test src/__tests__/auth.test.js   # specific test
hamlet show unit AuthService                   # specific code unit
hamlet show finding redundancy_candidate       # specific finding type
```

### Change-scoped filtering

```bash
hamlet pr                              # only changed files
hamlet pr --base origin/main           # changes since main
hamlet impact --base HEAD~3            # impact of last 3 commits
```

### Drill-down views

```bash
hamlet impact --show units             # impacted code units only
hamlet impact --show gaps              # protection gaps only
hamlet impact --show tests             # relevant tests only
hamlet impact --show owners            # affected owners only
```

## JSON for Custom Filtering

All commands support `--json` for programmatic filtering with `jq`:

```bash
# High-severity signals only
hamlet analyze --json | jq '.signals[] | select(.severity == "high")'

# Tests with low pass rate
hamlet analyze --json | jq '.testFiles[] | select(.runtimeStats.passRate < 0.8)'

# Untested exported code units
hamlet analyze --json | jq '.codeUnits[] | select(.exported == true and (.linkedTestFiles | length) == 0)'

# Portfolio findings for a specific owner
hamlet portfolio --json | jq '.findings[] | select(.owner == "team-platform")'

# Weak posture dimensions
hamlet posture --json | jq '.posture[] | select(.band == "WEAK" or .band == "CRITICAL")'
```

## Common Query Patterns

| Query | Command |
|-------|---------|
| What changed? | `hamlet pr` |
| Who's affected? | `hamlet impact --show owners` |
| What's untested? | `hamlet impact --show gaps` |
| Owner's health | `hamlet show owner <name>` |
| Test details | `hamlet show test <path>` |
| Unit coverage | `hamlet show unit <name>` |
| Leadership view | `hamlet summary` |
| Full evidence | `hamlet posture` |
