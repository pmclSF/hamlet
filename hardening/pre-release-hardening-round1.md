# Pre-Release Hardening Register (Round 1)

Date: 2026-03-10
Branch: `fix/pre-release-review`
Scope: Engineering quality, security, runtime, usability, product alignment

## Executed Checks

- `go test ./... -count=1` (pass)
- `go test ./... -race -count=1` (pass)
- `go vet ./...` (pass)
- `staticcheck ./...` (fail: 5 findings)
- `golangci-lint run` (fail: same 5 findings)
- `go test ./internal/testdata -run 'CLI|Golden|Determinism|Schema|Adversarial|E2E' -count=1` (pass)
- `npm run lint` (pass)
- `npm test -- --runInBand` (fail: 2 suites, 34 tests; sandbox bind restriction)
- `govulncheck ./...` (pass, no vulnerabilities)
- `npm audit --omit=dev --audit-level=high` (pass, 0 vulnerabilities)
- `go run ./benchmarks/cli --repo terrain --timeout 180` (partial pass: analyze/impact/explain OK; insights/debug timeout)
- `go run ./benchmarks/cli --repo terrain --command analyze --timeout 180` (pass, runtime 86.3s solo)
- `go run ./benchmarks/cli --repo terrain --command impact --timeout 180` (pass, runtime 120.0s solo)
- `go run ./benchmarks/cli --repo terrain --command insights --timeout 180` (timeout; plus output-dir failure when disk constrained)

## Findings

## S1

1. `show --help` is broken and slow
- Evidence: `go run ./cmd/terrain show --help` returns `unknown entity type: "--help" (valid: test, unit, codeunit, owner, finding)` and exits non-zero.
- Impact: CLI discoverability and self-service usability are degraded; common help path fails.
- Suggested fix: Add explicit `--help` handling in `show` subcommand parser and print subcommand usage immediately without pipeline work.

2. Benchmark harness has no live progress for long-running commands
- Evidence: `go run ./benchmarks/cli --repo terrain --timeout 180` provides no per-command heartbeat while command batch runs.
- Impact: Operational ambiguity; looks hung; poor CI/benchmark UX.
- Suggested fix: stream per-command start/finish logs and timeout notices as each command executes.

3. Benchmark harness likely retains very large command outputs in-memory
- Evidence: `CommandResult` stores full `stdout`/`stderr` for each command; analyze/debug JSON output on large repos is massive.
- Impact: memory pressure, slowdowns, inflated artifacts, potential OOM risk.
- Suggested fix: store bounded output (head/tail + byte count), optionally full raw output behind explicit `--capture-full-output` flag.

4. Core benchmark commands exceed timeout on large real repo
- Evidence (`/tmp/terrain-hardening-bench`): `insights`, `debug:graph`, `debug:coverage`, `debug:fanout`, and `debug:duplicates` all fail at ~180s (`exit -1`).
- Impact: major feature surface appears unavailable under realistic workloads; benchmark credibility collapses.
- Suggested fix: optimize command paths, add progressive output/partial results, and set per-command timeout budgets that reflect command complexity.

5. Benchmark run can fail late after expensive execution if output directory creation fails
- Evidence: `insights` solo benchmark ended with `error creating output dir: mkdir ... no space left on device` after full command runtime.
- Impact: expensive runs can be discarded at write stage; poor resilience and observability.
- Suggested fix: validate writeability and disk availability before execution begins; fail fast.

## S2

6. Static check: unused function `analyzeTestFileContent`
- File: `internal/analysis/content_analysis.go:22`
- Impact: dead code path; maintenance burden.
- Suggested fix: remove or wire to active pipeline path with tests.

7. Static check: unused function `ingestRuntime`
- File: `internal/engine/pipeline.go:696`
- Impact: dead code path; possible logic drift from active ingest path.
- Suggested fix: remove or call from pipeline; ensure single ingestion path.

8. Static check: unused function `ingestCoverage`
- File: `internal/engine/pipeline.go:761`
- Impact: dead code path; possible logic drift from active ingest path.
- Suggested fix: remove or call from pipeline; ensure single ingestion path.

9. Full Node suite depends on local TCP bind and fails in restricted environments
- Evidence: failing suites `test/server/server-api.test.js` and `test/server/ui-smoke.test.js` with `listen EPERM: operation not permitted 127.0.0.1`.
- Impact: environment-sensitive CI hardening; false negatives in sandboxed runners.
- Suggested fix: refactor server tests to avoid real socket binding where possible (inject handler into request harness), or gracefully skip with explicit reason if bind is disallowed.

10. Analyze output relevance is weak on repos containing benchmark corpora
- Evidence: `terrain analyze --root .` discovers 59k+ test files and surfaces mostly low-priority findings; top findings are low-severity custom matcher items while risk section is massive.
- Impact: user signal-to-noise drops sharply; first-run actionability suffers.
- Suggested fix: add optional default ignore presets (`benchmarks/`, large fixture corpora) or introduce stronger prioritization weighting for top findings.

11. Version-era references remain outside changelog in docs/comments
- Evidence:
  - `docs/vscode-extension.md` contains numbered release-line references.
  - `docs/demo.md`, `docs/engineering/impact-analysis-system-map.md`, and internal comments contain numbered release-line references.
  - `docs/README.md` links legacy docs using versioned filenames (acceptable if intentionally marked legacy, but still a policy exception).
- Impact: copy-policy inconsistency and mixed product narrative.
- Suggested fix: normalize to "current architecture" wording outside `CHANGELOG.md`; keep legacy references only in `docs/legacy/**` with clear historical labeling.

12. `analyze --help` default text is duplicated for `--slow-threshold`
- Evidence: `slow test threshold in ms (default: 5000) (default 5000)`
- Impact: minor polish issue in CLI quality.
- Suggested fix: remove one source of default text formatting.

## S3

13. Static simplification opportunity in snapshot test
- File: `cmd/terrain/snapshot_test.go:51`
- Finding: nil check before `len(map)` is unnecessary.
- Suggested fix: simplify expression.

14. Static simplification opportunity in CODEOWNERS parser
- File: `internal/ownership/codeowners.go:177`
- Finding: `if strings.HasPrefix(p, "/")` can be unconditional `TrimPrefix`.
- Suggested fix: simplify for readability.

## Runtime Benchmark Snapshot

- Full command batch (`--repo terrain --timeout 180`) results:
  - `analyze` OK 57.1s
  - `impact` OK 54.4s
  - `insights` timeout 181.8s
  - `explain` OK 106.0s (credibility reduced: missing dependency path/confidence)
  - `debug:*` commands timeout around 180s
- Single-command spot checks:
  - `analyze` solo: 86.3s
  - `impact` solo: 120.0s
- Concurrent single-command reruns materially increased latency (up to ~142-147s), reinforcing contention sensitivity.
