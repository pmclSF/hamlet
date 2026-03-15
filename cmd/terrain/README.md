# cmd/terrain

Go CLI entry point for the Terrain test system intelligence platform.

## Primary Commands (Canonical User Journeys)

| Command | Question it answers |
|---------|---------------------|
| `terrain analyze` | What is the state of our test system? |
| `terrain impact` | What validations matter for this change? |
| `terrain insights` | What should we fix in our test system? |
| `terrain explain <target>` | Why did Terrain make this decision? |

## Supporting Commands

| Command | Purpose |
|---------|---------|
| `terrain init` | Detect data files and print recommended analyze command |
| `terrain summary` | Executive summary with risk, trends, benchmark readiness |
| `terrain focus` | Prioritized next actions |
| `terrain posture` | Detailed posture breakdown with measurement evidence |
| `terrain portfolio` | Portfolio intelligence: cost, breadth, leverage, redundancy |
| `terrain metrics` | Aggregate metrics scorecard |
| `terrain compare` | Compare two snapshots for trend tracking |
| `terrain select-tests` | Recommend protective test set for a change |
| `terrain pr` | PR/change-scoped analysis |
| `terrain show <entity> <id>` | Drill into test, unit, owner, or finding |
| `terrain migration <sub>` | Migration readiness, blockers, or preview |
| `terrain policy check` | Evaluate local policy rules |
| `terrain export benchmark` | Privacy-safe JSON export for benchmarking |

## Advanced / Debug Commands

| Command | Purpose |
|---------|---------|
| `terrain debug graph` | Dependency graph statistics |
| `terrain debug coverage` | Structural coverage analysis |
| `terrain debug fanout` | High-fanout node analysis |
| `terrain debug duplicates` | Duplicate test cluster analysis |
| `terrain debug depgraph` | Full dependency graph analysis (all engines) |

## Canonical Workflow

Run the four primary journeys in order:

```bash
terrain analyze                          # understand your test system
terrain insights                         # find what to improve
terrain impact --base main               # see what a change affects
terrain explain src/auth/login.test.ts   # understand why
```

This flow maps to the canonical product journeys documented in [docs/product/canonical-user-journeys.md](../../docs/product/canonical-user-journeys.md).

All commands support `--root PATH` and `--json` flags. See [docs/cli-spec.md](../../docs/cli-spec.md) for the full CLI specification.
