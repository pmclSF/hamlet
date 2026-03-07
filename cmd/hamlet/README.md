# cmd/hamlet

V3 CLI entry point for the Hamlet test intelligence platform.

## Implemented Commands

| Command | Purpose |
|---------|---------|
| `hamlet analyze` | Full test suite analysis |
| `hamlet summary` | Executive summary with risk, trends, benchmark readiness |
| `hamlet posture` | Detailed posture breakdown with measurement evidence |
| `hamlet metrics` | Aggregate metrics scorecard |
| `hamlet impact` | Impact analysis for changed code |
| `hamlet compare` | Compare two snapshots for trend tracking |
| `hamlet migration readiness` | Migration readiness assessment |
| `hamlet migration blockers` | List migration blockers by type and area |
| `hamlet migration preview` | Preview migration for a file or scope |
| `hamlet policy check` | Evaluate local policy rules |
| `hamlet export benchmark` | Privacy-safe JSON export for benchmarking |
| `hamlet version` | Show version, commit, and build date |

All commands support `--root PATH` and `--json` flags. See [docs/cli-spec.md](../../docs/cli-spec.md) for the full CLI specification.
