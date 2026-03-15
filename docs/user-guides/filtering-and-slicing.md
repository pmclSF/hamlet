# Filtering and Slicing

Terrain supports filtering and slicing across commands to focus on the relevant part of a large repo.

## Available Filters

### Owner filtering

```bash
terrain impact --owner team-platform     # impact scoped to one team
terrain show owner team-platform          # drill into owner details
```

### Path/entity filtering

```bash
terrain show test src/__tests__/auth.test.js   # specific test
terrain show unit AuthService                   # specific code unit
terrain show finding redundancy_candidate       # specific finding type
```

### Change-scoped filtering

```bash
terrain pr                              # only changed files
terrain pr --base origin/main           # changes since main
terrain impact --base HEAD~3            # impact of last 3 commits
```

### Drill-down views

```bash
terrain impact --show units             # impacted code units only
terrain impact --show gaps              # protection gaps only
terrain impact --show tests             # relevant tests only
terrain impact --show owners            # affected owners only
```

## JSON for Custom Filtering

All commands support `--json` for programmatic filtering with `jq`:

```bash
# High-severity signals only
terrain analyze --json | jq '.signals[] | select(.severity == "high")'

# Tests with low pass rate
terrain analyze --json | jq '.testFiles[] | select(.runtimeStats.passRate < 0.8)'

# Untested exported code units
terrain analyze --json | jq '.codeUnits[] | select(.exported == true and (.linkedTestFiles | length) == 0)'

# Portfolio findings for a specific owner
terrain portfolio --json | jq '.findings[] | select(.owner == "team-platform")'

# Weak posture dimensions
terrain posture --json | jq '.posture[] | select(.band == "WEAK" or .band == "CRITICAL")'
```

## Common Query Patterns

| Query | Command |
|-------|---------|
| What changed? | `terrain pr` |
| Who's affected? | `terrain impact --show owners` |
| What's untested? | `terrain impact --show gaps` |
| Owner's health | `terrain show owner <name>` |
| Test details | `terrain show test <path>` |
| Unit coverage | `terrain show unit <name>` |
| Leadership view | `terrain summary` |
| Full evidence | `terrain posture` |
