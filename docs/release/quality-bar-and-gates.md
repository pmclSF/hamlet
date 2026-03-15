# Quality Bar and Release Gates

Terrain enforces a tiered gating system to catch defects at the earliest
possible stage. Each tier builds on the previous one, with faster checks
running first.

## Tier 1: PR Gates (Required on Every PR)

These checks run on every pull request and must pass before merge. They
complete in under 2 minutes.

| Gate | Command | What It Catches |
|------|---------|-----------------|
| `go vet` | `go vet ./...` | Static analysis: unused variables, formatting errors, suspicious constructs |
| `go test` | `go test ./...` | Unit test failures, integration test failures |
| `go build` | `go build ./cmd/terrain` | Compilation errors, missing dependencies |
| Golden tests | `go test ./... -run TestGolden` | Output regressions against checked-in golden files |

**Makefile targets:**

```makefile
make check       # Runs go vet
make test        # Runs go test ./...
make test-v      # Runs go test ./... -v (verbose)
make test-cover  # Runs go test ./... -coverprofile=coverage.out
```

PR gates are configured in the CI workflow and block merge on failure.
No exceptions, no manual overrides.

## Tier 2: Release Gates (Required Before Release)

Release gates include all PR gates plus additional checks that are too
slow or environment-dependent for every PR but must pass before any
release is tagged.

| Gate | What It Validates |
|------|-------------------|
| All PR gates | Everything in Tier 1 |
| Determinism suite | Same input produces byte-identical output across runs |
| Schema compatibility | JSON output conforms to published schemas in `docs/schema/` |
| Benchmark-safe export | Exported benchmark data excludes sensitive fields and meets format requirements |
| E2E scenarios | Full pipeline workflows produce correct end-to-end results |
| CLI regression | Binary builds, runs, and produces expected output for all commands |
| Docs consistency | CLI help text matches documented commands and flags |

Release gates are run manually before tagging a release or automatically
in the release CI workflow. See `docs/release/release-checklist-final.md`
for the full release process.

## Tier 3: Nightly and Optional Gates

These checks run on a schedule or on-demand. They are not required for
release but surface issues that accumulate over time.

| Gate | Frequency | Purpose |
|------|-----------|---------|
| Full benchmark suite | Nightly | Performance regression detection across all subsystems |
| Large-scale tests | Weekly | Validation against very large repositories (10k+ files) |
| Adversarial suite expansion | On-demand | Edge cases, malformed inputs, pathological repo structures |

Failures in Tier 3 create tracking issues but do not block releases
unless the failure indicates a correctness bug.

## CI Workflow Integration

The CI pipeline is defined in `.github/workflows/` and integrates with
the gate tiers as follows:

```
PR opened/updated
  --> Tier 1 gates run automatically
  --> All must pass for merge

Release branch created
  --> Tier 1 + Tier 2 gates run automatically
  --> All must pass for tag

Nightly schedule
  --> Tier 3 gates run automatically
  --> Failures create issues
```

### Key CI Configuration Points

- **Go version matrix.** Tests run against the minimum supported Go version
  and the latest stable release.
- **Race detector.** `go test -race` is enabled in CI to catch data races
  that may not manifest in single-threaded test runs.
- **Coverage threshold.** Coverage is computed but not gated -- the quality
  bar is enforced through test design (golden tests, E2E scenarios) rather
  than coverage percentages.

## Promoting a Gate

When a new test category is added:

1. Start in Tier 3 (nightly) to validate stability and runtime.
2. After 2 weeks of green runs, promote to Tier 2 (release).
3. If runtime is under 30 seconds, consider promoting to Tier 1 (PR).

Never add a flaky test to Tier 1. Fix the flakiness first, run it in
Tier 3 to confirm stability, then promote.

See also:
- `docs/release/release-checklist-final.md` for the release process
- `docs/engineering/determinism.md` for the deterministic output contract
- `docs/engineering/verification-system-map.md` for the full test layer diagram
