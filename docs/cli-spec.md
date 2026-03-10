# CLI Specification

## Philosophy

The CLI is the primary OSS interface to Hamlet.

It must be:
- fast
- trustworthy
- scriptable
- readable
- machine-friendly

### `hamlet version`
Purpose:
Print version, commit, and build date.

Output: `hamlet <version> (commit <sha>, built <date>)`

---

## Core commands

### `hamlet init`
Purpose:
Inspect a repository for common coverage/runtime artifacts and print a
ready-to-run `hamlet analyze` command with detected paths.

Flags:
- `--root PATH` — repository root to inspect (default: current directory)

### `hamlet analyze`
Primary command.

Purpose:
Generate a snapshot and summary of:
- frameworks
- health signals
- quality signals
- migration signals
- risk surfaces

Must support:
- human-readable output
- JSON output
- snapshot persistence

Flags:
- `--root PATH` — repository root to analyze (default: current directory)
- `--json` — output JSON snapshot
- `--write-snapshot` — persist snapshot to .hamlet/snapshots/latest.json
- `--coverage PATH` — ingest coverage data (LCOV, Istanbul JSON)

### `hamlet impact`
Purpose:
Impact analysis for changed code. Shows which code units, test gaps,
tests, and owners are affected by a git diff.

Flags:
- `--root PATH` — repository root (default: current directory)
- `--base REF` — git base ref for diff (default: HEAD~1)
- `--json` — output JSON impact result
- `--show VIEW` — drill-down: units, gaps, tests, owners
- `--owner NAME` — filter results by owner

### `hamlet posture`
Purpose:
Detailed posture breakdown with measurement evidence by dimension.

Flags:
- `--root PATH` — repository root (default: current directory)
- `--json` — output JSON posture snapshot

### `hamlet migration readiness`
Purpose:
Assess migration readiness with framework inventory, blocker taxonomy,
quality factors that compound migration risk, area-by-area safety
classification, and coverage guidance.

Must support:
- human-readable output (default)
- JSON output (`--json`)

Flags:
- `--root PATH` — repository root to analyze (default: current directory)
- `--json` — output JSON readiness summary

### `hamlet migration blockers`
Purpose:
List migration blockers by type and highest-risk areas. Focused view
for teams actively planning a framework migration.

Must support:
- human-readable output (default)
- JSON output (`--json`)

Flags:
- `--root PATH` — repository root to analyze (default: current directory)
- `--json` — output JSON blockers summary

### `hamlet migration preview`
Purpose:
Preview migration for a single file or directory scope. Shows source
framework, suggested target, blockers, safe patterns, and difficulty
assessment. Honest about limitations when preview is not possible.

Must support:
- human-readable output (default)
- JSON output (`--json`)

Flags:
- `--root PATH` — repository root to analyze (default: current directory)
- `--json` — output JSON preview result
- `--file PATH` — preview a single file (relative to root)
- `--scope DIR` — preview all files in a directory scope

### `hamlet compare`
Purpose:
Compare two snapshots and show trend changes.

Must support:
- human-readable output (default)
- JSON output (`--json`)

Flags:
- `--root PATH` — repository root (default: current directory)
- `--from PATH` — baseline snapshot JSON (default: second-latest in .hamlet/snapshots/)
- `--to PATH` — current snapshot JSON (default: latest in .hamlet/snapshots/)
- `--json` — output JSON comparison

Behavior:
- If --from and --to are not specified, uses the two most recent timestamped snapshots
- If fewer than two snapshots exist, returns a clear error message

### `hamlet policy check`
Purpose:
Evaluate current repository state against local Hamlet policy.

Must support:
- human-readable output (default)
- JSON output (`--json`)
- CI-friendly exit codes

Flags:
- `--root PATH` — repository root to analyze (default: current directory)
- `--json` — output JSON result

Exit codes:
- `0` — no policy file found, or policy exists with no violations
- `1` — policy file malformed or evaluation/runtime error
- `2` — policy violations found (CI gate signal)

Policy file:
- Loaded from `.hamlet/policy.yaml` in the analyzed repository root
- Missing policy file is not an error (exit 0, informational message)
- Malformed policy file produces an actionable error (exit 1)

Supported policy rules:
- `disallow_skipped_tests` — flag skipped tests as violations
- `disallow_frameworks` — list of framework names not permitted
- `max_test_runtime_ms` — maximum average test runtime
- `minimum_coverage_percent` — minimum coverage threshold
- `max_weak_assertions` — maximum allowed weakAssertion signals
- `max_mock_heavy_tests` — maximum allowed mockHeavyTest signals

### `hamlet metrics`
Purpose:
Output aggregate, benchmark-ready metrics scorecard.

Must support:
- human-readable output (default)
- JSON output (`--json`)

Flags:
- `--root PATH` — repository root to analyze (default: current directory)
- `--json` — output JSON metrics snapshot

Metrics categories:
- Structure: test file count, frameworks, fragmentation ratio, languages
- Health: slow/flaky/skipped/dead test counts and ratios
- Quality: weak assertions, mock-heavy tests, untested exports, coverage breaks
- Change readiness: migration blockers, deprecated patterns, dynamic generation, custom matchers
- Governance: policy violations, legacy framework usage, runtime budget exceeded
- Risk: reliability/change/speed bands, high-risk area count, critical findings

Privacy boundary:
- Metrics contain only aggregate counts, ratios, and qualitative bands
- No raw file paths, symbol names, source code, or user identity
- Safe for future anonymous aggregation

### `hamlet summary`
Purpose:
Executive summary — leadership-oriented risk, trend, and benchmark readiness report.

Must support:
- human-readable output (default)
- JSON output (`--json`)

Flags:
- `--root PATH` — repository root to analyze (default: current directory)
- `--json` — output JSON executive summary (ExecutiveSummary model)

Output includes:
- Overall posture by dimension (reliability, change, speed, governance)
- Key numbers (test files, frameworks, total signals, critical findings, high-risk areas)
- Top risk areas with risk type and band
- Trend highlights (if prior snapshots exist in .hamlet/snapshots/)
- Dominant signal drivers
- Evidence-based recommended focus
- Benchmark readiness (ready dimensions, limited dimensions, segmentation)

Behavior:
- Automatically attempts to load prior snapshots for trend comparison
- Gracefully degrades when no snapshot history exists
- Benchmark readiness section describes what is measurable, not how it ranks

### `hamlet export benchmark`
Purpose:
Output a benchmark-safe JSON artifact for future anonymous comparison.

Flags:
- `--root PATH` — repository root to analyze (default: current directory)

Output is always JSON (no human-readable mode — this is a machine artifact).

Export includes:
- Schema version for compatibility
- Segmentation tags (primary language, primary framework, test file bucket, framework count, coverage/runtime/policy presence)
- Full aggregate metrics (same as `hamlet metrics --json`)

Privacy boundary:
- No raw file paths, symbol names, source code, or user identity
- Only aggregate counts, ratios, qualitative bands, and segmentation tags
- Safe for anonymous aggregation and cross-repo comparison

### `hamlet insights`
Purpose:
Prioritized improvement actions with rationale. Shows what to fix first
and why, based on signal analysis.

Flags:
- `--root PATH` — repository root (default: current directory)
- `--json` — output JSON insights

### `hamlet explain`
Purpose:
Evidence chain for a specific entity — test file, code unit, owner, or finding.
Answers "why did Hamlet say this?"

Usage: `hamlet explain <test-path|test-id|code-unit|owner|finding>`

Flags:
- `--root PATH` — repository root (default: current directory)
- `--json` — output JSON explanation

### `hamlet focus`
Purpose:
Focus summary — where to concentrate testing effort based on risk,
coverage gaps, and recent changes.

Flags:
- `--root PATH` — repository root (default: current directory)
- `--json` — output JSON focus summary

### `hamlet portfolio`
Purpose:
Portfolio view of the test suite — treats the test suite as a portfolio
of investments, showing coverage breadth, test type distribution, and
risk allocation across the codebase.

Flags:
- `--root PATH` — repository root (default: current directory)
- `--json` — output JSON portfolio snapshot

### `hamlet select-tests`
Purpose:
Protective test selection for changed code. Given a git diff, returns the
minimal set of tests that should run to cover affected code paths.

Flags:
- `--root PATH` — repository root (default: current directory)
- `--base REF` — git base ref for diff (default: HEAD~1)
- `--json` — output JSON protective test set

### `hamlet pr`
Purpose:
PR analysis — combined impact, test selection, and risk summary formatted
for pull request review workflows.

Flags:
- `--root PATH` — repository root (default: current directory)
- `--base REF` — git base ref for diff (default: HEAD~1)
- `--json` — output JSON PR analysis
- `--format FORMAT` — output format: markdown, comment, annotation

### `hamlet show`
Purpose:
Drill into a specific entity — show details for a test, code unit, owner,
or finding by ID or path.

Usage: `hamlet show <test|unit|codeunit|owner|finding> <id-or-path>`

Flags:
- `--root PATH` — repository root (default: current directory)
- `--json` — output JSON

### `hamlet debug`
Purpose:
Developer debugging commands for inspecting internal analysis state.

Usage: `hamlet debug <graph|coverage|fanout|duplicates> [flags]`

Subcommands:
- `graph` — dependency graph statistics
- `coverage` — coverage attribution details
- `fanout` — import fan-out analysis
- `duplicates` — test identity collision detection

Flags:
- `--root PATH` — repository root (default: current directory)
- `--json` — output JSON
- `--changed FILES` — comma-separated changed files for impact context

### `hamlet depgraph`
Purpose:
Full dependency graph inspection with multiple sub-views.

Flags:
- `--root PATH` — repository root (default: current directory)
- `--json` — output JSON
- `--show VIEW` — sub-view: stats, coverage, duplicates, fanout, impact, profile
- `--changed FILES` — comma-separated changed files for impact analysis

---

## Output rules

### Human-readable output
Should emphasize:
- top findings
- representative examples
- next actions

### JSON output
Should be stable enough to support:
- extension rendering
- CI integration
- snapshot persistence
- future hosted ingestion
