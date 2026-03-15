# CLI Drill-Down Commands

Terrain provides entity-inspection commands to navigate from summary views down to specific tests, code units, owners, and findings.

## Commands

### `terrain show test <path-or-id>`

Inspect a test file or test case by path or test ID.

```bash
terrain show test src/__tests__/auth.test.js
terrain show test abc123def456  # by test ID
```

Shows: framework, owner, test/assertion counts, runtime stats, coverage links, signals.

### `terrain show unit <name-or-path>`

Inspect a code unit by name, path, or unit ID.

```bash
terrain show unit AuthService
terrain show unit src/auth.js:AuthService
```

Shows: kind, exported status, owner, covering tests.

### `terrain show owner <name>`

Inspect an owner's test portfolio.

```bash
terrain show owner team-platform
```

Shows: owned files, test files, signal count, top signals.

### `terrain show finding <id-or-type>`

Inspect a portfolio finding or signal by index or type.

```bash
terrain show finding redundancy_candidate
terrain show finding 0           # by index
terrain show finding s0           # signal by index
```

Shows: type, path, confidence, explanation, suggested action.

## JSON Output

All drill-down commands support `--json` for programmatic consumption:

```bash
terrain show test src/auth.test.js --json
terrain show owner team-platform --json
```

## Navigation Pattern

The typical drill-down flow:

1. `terrain summary` — see the overview
2. `terrain posture` — understand measurement evidence
3. `terrain show owner team-platform` — focus on one team
4. `terrain show test src/__tests__/auth.test.js` — inspect a specific test
5. `terrain show unit AuthService` — see what covers this unit
