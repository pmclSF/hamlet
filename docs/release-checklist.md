# Release Checklist

Status checklist for Hamlet V3 release readiness.

## Product

- [x] Core commands work: analyze, summary, metrics, compare, policy check, export benchmark
- [x] Human-readable and JSON output for all major commands
- [x] Executive summary with posture, trends, focus, benchmark readiness
- [x] Sample outputs updated in examples/sample-reports/
- [x] Demo walkthrough documented (docs/demo.md)

## Engineering

- [x] All Go packages build cleanly (`go build ./internal/... ./cmd/...`)
- [x] All Go tests pass (`go test ./internal/... ./cmd/...`)
- [x] 16 internal packages with test coverage
- [x] Snapshot contract (`TestSuiteSnapshot`) stable for current features
- [x] Extension scaffold present with type definitions and view builders

## UX

- [x] `hamlet --help` output is accurate and readable
- [x] Each subcommand has clear flag descriptions
- [x] Summary command includes trend + benchmark readiness
- [x] Graceful degradation when no snapshot history exists
- [x] Graceful degradation when no policy file exists
- [ ] Extension TreeDataProvider implementations (scaffold only — pending)

## Docs

- [x] README accurate with quick start, commands, architecture
- [x] docs/README.md provides navigation index
- [x] docs/demo.md walkthrough complete
- [x] docs/cli-spec.md covers all commands
- [x] docs/roadmap.md milestones A through O documented
- [x] docs/implementation-workbook.md stages 1 through 14 documented

## Packaging

- [x] `go install` path documented
- [x] Build-from-source instructions in README
- [x] Makefile with build, test, demo targets
- [ ] Binary releases / goreleaser (not yet configured)
- [ ] Homebrew formula (not yet created)

## Hardening

- [x] Stale messages removed ("not yet active", "analysis nucleus")
- [x] Compare error messages improved (actionable guidance for missing snapshots)
- [x] CLI help text aligned with actual behavior
- [x] Doc comment in main.go matches CLI surface
- [x] Test expectations updated for output changes
- [x] Empty states reviewed across all reports

## Honest Gaps

These are intentionally not yet shipped:

- **Extension**: Scaffold only — TreeDataProvider implementations pending
- **Runtime data**: Health signals depend on runtime artifacts that most repos don't produce yet
- **Coverage data**: Coverage-based signals require coverage reports in a supported format
- **Benchmark comparison**: Export model ready, but no hosted comparison service exists
- **Cross-repo aggregation**: Metrics model supports it, but aggregation is not implemented
- **Organization features**: No auth, accounts, or portfolio management
- **Hosted dashboard**: Not built — product is local-only
- **Binary distribution**: No pre-built binaries; requires Go toolchain to install

## Stability

The JSON contract (`TestSuiteSnapshot`) is stabilizing but may evolve. Early adopters should expect minor schema changes in signal metadata and risk surface fields.

Core concepts are stable:
- Signal types and categories
- Risk bands (low, medium, high, critical)
- Snapshot structure
- Policy rule format
- Metrics aggregate model
- Benchmark export schema v1

## Recommended Next Steps

1. Extension TreeDataProvider implementation
2. Binary release automation (goreleaser)
3. Runtime artifact ingestion improvements
4. Coverage report parsing
5. First public release communication
