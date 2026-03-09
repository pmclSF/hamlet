# V3 Release Checklist

## Build verification
- [ ] `go build ./cmd/hamlet/` succeeds
- [ ] `go vet ./...` clean
- [ ] `go test ./internal/...` all pass
- [ ] Binary runs: `hamlet --help` shows updated help text

## First-run flow
- [ ] `hamlet analyze` produces useful output on a real repo
- [ ] `hamlet summary` produces a concise leadership overview
- [ ] `hamlet posture` shows all 5 dimensions with evidence
- [ ] `hamlet metrics` produces aggregate scorecard
- [ ] Every command includes "next steps" hints

## Demo fixtures
- [ ] `fixtures/demos/healthy-balanced.json` validates
- [ ] `fixtures/demos/flaky-concentrated.json` validates
- [ ] `fixtures/demos/e2e-heavy-shallow.json` validates
- [ ] `fixtures/demos/fragmented-migration-risk.json` validates

## Portfolio
- [ ] `hamlet portfolio` produces correct output on a real repo
- [ ] Portfolio findings (redundancy, leverage, runtime concentration) surface in `hamlet analyze`
- [ ] Demo fixture `bloated-overlapping-tests.json` validates

## Snapshot and comparison
- [ ] `hamlet analyze --write-snapshot` persists correctly
- [ ] `hamlet compare` works with two snapshots
- [ ] Trend highlights render properly

## Policy
- [ ] `hamlet policy check` works with no policy file (exit 0)
- [ ] `hamlet policy check` works with policy file and violations (exit 1)

## Export
- [ ] `hamlet export benchmark` produces valid JSON
- [ ] Export contains no raw file paths or symbol names
- [ ] Schema version is "2"
- [ ] Posture bands are included in export

## Documentation
- [ ] README reflects current commands and behavior
- [ ] docs/product/ has positioning, narrative, wow moments
- [ ] docs/user-guides/ has getting-started and first-10-minutes
- [ ] docs/demos/ has fixture manifest and wow workflows

## Output quality
- [ ] No command dumps overwhelming raw data by default
- [ ] All human-readable output is scannable and concise
- [ ] JSON output is stable and well-structured
- [ ] No ANSI codes in default output (copy-paste safe)

## Version and metadata
- [ ] go.mod version is correct
- [ ] GeneratedAt timestamps are UTC
- [ ] Analysis version string is current
