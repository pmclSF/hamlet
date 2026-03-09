# CLI Drill-Down Commands

Hamlet provides entity-inspection commands to navigate from summary views down to specific tests, code units, owners, and findings.

## Commands

### `hamlet show test <path-or-id>`

Inspect a test file or test case by path or test ID.

```bash
hamlet show test src/__tests__/auth.test.js
hamlet show test abc123def456  # by test ID
```

Shows: framework, owner, test/assertion counts, runtime stats, coverage links, signals.

### `hamlet show unit <name-or-path>`

Inspect a code unit by name, path, or unit ID.

```bash
hamlet show unit AuthService
hamlet show unit src/auth.js:AuthService
```

Shows: kind, exported status, owner, covering tests.

### `hamlet show owner <name>`

Inspect an owner's test portfolio.

```bash
hamlet show owner team-platform
```

Shows: owned files, test files, signal count, top signals.

### `hamlet show finding <id-or-type>`

Inspect a portfolio finding or signal by index or type.

```bash
hamlet show finding redundancy_candidate
hamlet show finding 0           # by index
hamlet show finding s0           # signal by index
```

Shows: type, path, confidence, explanation, suggested action.

## JSON Output

All drill-down commands support `--json` for programmatic consumption:

```bash
hamlet show test src/auth.test.js --json
hamlet show owner team-platform --json
```

## Navigation Pattern

The typical drill-down flow:

1. `hamlet summary` — see the overview
2. `hamlet posture` — understand measurement evidence
3. `hamlet show owner team-platform` — focus on one team
4. `hamlet show test src/__tests__/auth.test.js` — inspect a specific test
5. `hamlet show unit AuthService` — see what covers this unit
