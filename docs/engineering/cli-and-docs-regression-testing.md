# CLI and Docs Regression Testing

Hamlet's CLI regression tests verify that the compiled binary behaves
correctly from the user's perspective. These tests live in
`internal/testdata/cli_test.go` and use `exec.Command` to build and invoke
the `hamlet` binary directly.

## How CLI Tests Work

Each test follows the same pattern:

1. Build the binary from `cmd/hamlet/` using `go build`.
2. Execute the binary with specific arguments via `exec.Command`.
3. Assert on exit code, stdout content, and stderr content.

This approach tests the real binary -- no in-process shortcutting, no
function-level invocation. If the CLI parses flags incorrectly, exits with
the wrong code, or prints malformed output, these tests catch it.

## CLI Test Inventory

| Test | Command | Validates |
|------|---------|-----------|
| Build Succeeds | `go build ./cmd/hamlet` | The binary compiles without errors |
| Help Exit Code | `hamlet --help` | Exit code 0 on help request |
| Help Contains All Commands | `hamlet --help` | Every registered subcommand appears in help text |
| Unknown Command Exit Code | `hamlet nonexistent` | Non-zero exit code for unrecognized commands |
| Analyze Testdata | `hamlet analyze <path>` | Successful analysis of the sample repo with text output |
| Analyze JSON | `hamlet analyze --format json <path>` | JSON output parses correctly and contains expected fields |
| Posture Testdata | `hamlet posture <path>` | Posture report generates without error against sample repo |
| Metrics Testdata | `hamlet metrics <path>` | Metrics output includes expected measurement names |
| Summary Testdata | `hamlet summary <path>` | Summary report renders with expected sections and data |

## The Sample Repository

CLI tests run against the sample repository at
`internal/analysis/testdata/sample-repo/`. This fixture contains:

- Multiple test files across different frameworks (Go, JavaScript, Python)
- Configuration files (go.mod, package.json, pytest.ini)
- Source files with varying coverage characteristics
- CODEOWNERS file for ownership attribution
- Intentional quality signals (duplicated tests, missing assertions, etc.)

The sample repo is small enough to analyze in under a second but rich enough
to exercise all major detectors and measurements.

## Docs-Consistency Checks

A subset of CLI tests verify that documentation stays in sync with the
actual CLI behavior:

- **Command names in help text.** Every subcommand listed in `docs/cli-spec.md`
  must appear in `hamlet --help` output. If a command is added or renamed in
  code but not updated in docs, this test fails.
- **Flag names.** Key flags documented in the CLI spec are checked against
  actual flag registration to catch drift.
- **Output format descriptions.** When docs claim a command supports
  `--format json`, the test confirms the flag is accepted.

These checks prevent a common failure mode: the CLI evolves but the docs
lag behind, leaving users with inaccurate instructions.

## Running CLI Tests

```bash
# Run all CLI tests
go test ./internal/testdata/ -run TestCLI

# Run with verbose output to see command invocations
go test ./internal/testdata/ -run TestCLI -v
```

CLI tests are included in the standard `go test ./...` run and are part of
both PR gates and release gates.

## Failure Investigation

When a CLI test fails:

1. Check the test output for the exact command that was run.
2. Run that command manually to reproduce the failure.
3. If the exit code is wrong, check command registration in `cmd/hamlet/`.
4. If the output is wrong, trace through the relevant render function.
5. If a docs-consistency check fails, update the docs or the CLI to match.

See also:
- `docs/cli-spec.md` for the full CLI specification
- `docs/engineering/e2e-scenario-testing.md` for pipeline-level E2E tests
- `docs/engineering/verification-system-map.md` for the full test layer diagram
