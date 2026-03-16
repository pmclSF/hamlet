# Build, Test, and Release Pipeline

> **Status:** Current (validated 2026-03-15)
> **Purpose:** Document the CI/CD pipeline, test gates, and release process for Terrain.

## CI Pipeline (`ci.yml`)

The CI pipeline runs on every push to `main` and on every pull request. All jobs are hard gates — any failure blocks the build.

### Jobs

```
ci.yml
├─ test (JS)              → format + lint + jest (Node 22.x, 24.x)
├─ go-test                → go mod tidy + build + unit tests + golden tests
│                            + CLI smoke tests + fixture matrix + benchmark smoke
├─ go-bench-compare       → benchstat base vs head (PRs only)
└─ build (integration)    → JS CLI sanity + conversion round-trip
```

### Job: `test` (JavaScript)

**Matrix:** Node 22.x, 24.x

| Step | Command | Gate |
|------|---------|------|
| Format check | `npm run format:check` | Hard |
| Lint | `npm run lint` | Hard |
| Tests | `npm test` | Hard |

### Job: `go-test` (Go)

| Step | Command | Purpose |
|------|---------|---------|
| Tidy check | `go mod tidy && git diff --exit-code` | No dependency drift |
| Build | `go build -o terrain ./cmd/terrain` | Binary compiles |
| Unit tests | `go test ./internal/... ./cmd/... -count=1 -race` | All packages pass with race detector |
| Golden tests | `go test ./cmd/terrain/ -run TestSnapshot -count=1 -v` | Snapshot stability for 4 canonical commands |
| CLI smoke | `./terrain version && analyze + insights on sample-repo` | Binary runs, produces JSON |
| Fixture matrix | `analyze --json` on 5 fixtures | Handles diverse repo shapes |
| Benchmark smoke | 4 canonical commands validated across 3 fixtures | Output structure assertions |

**Benchmark Smoke Tests** validate that the 4 canonical commands produce structurally correct output:

| Fixture | Commands Tested | Assertions |
|---------|----------------|------------|
| sample-repo | `analyze`, `insights` | testFileCount > 0, healthGrade present |
| ai-eval-suite | `analyze`, `ai list` | testFileCount > 0, scenarioCount > 0 |
| backend-api | `analyze` | 5 posture dimensions present |

### Job: `go-bench-compare` (PRs only)

Compares Go benchmarks between PR head and base using `benchstat`:
- `BenchmarkRunPipeline`
- `BenchmarkSignalDetection`
- `BenchmarkBuildImportGraph`
- `BenchmarkRiskScore`
- `BenchmarkExtractTestCases`

Results posted to GitHub Step Summary.

### Job: `build` (Integration)

Runs after `test` and `go-test` pass. Validates the JS converter engine:
- `bin/terrain.js --version` and `--help`
- Cypress → Playwright conversion round-trip
- Playwright → Cypress conversion round-trip

## PR Analysis (`terrain-pr.yml`)

Runs on every PR to `main`. Uses Terrain to analyze its own changes:

1. Builds Go binary
2. Runs `terrain pr --json` against the PR diff
3. Extracts selected Go and JS tests
4. Runs only the selected tests (targeted test execution)
5. Posts/updates a PR comment with impact summary and test results
6. Fails if selected tests fail

## Security Analysis (`codeql.yml`)

CodeQL scans on push, PR, and weekly schedule:
- **Languages:** Go, JavaScript/TypeScript, Python
- Uses GitHub's `codeql-action/init` + `autobuild` + `analyze`

## Release Pipeline

### Trigger

Tag push matching `v*` triggers `release.yml`.

### Flow

```
Tag push (v*)
├─ verify
│   ├─ Tag matches package.json version
│   └─ npm run release:verify (format + lint + test + pack)
├─ release (needs: verify)
│   ├─ npm publish --provenance
│   └─ Create GitHub Release
└─ go-release (needs: verify)
    └─ GoReleaser: cross-compile + attach binaries
```

### Artifacts

| Artifact | Format | Targets |
|----------|--------|---------|
| npm package | `terrain-testframework` | Registry: npmjs.org |
| Go binaries | `terrain` | Linux (amd64, arm64), Darwin (amd64, arm64), Windows (amd64, arm64) |
| Checksums | SHA-256 | Attached to GitHub Release |

### GoReleaser Configuration

```yaml
# .goreleaser.yaml
builds:
  - main: ./cmd/terrain
    env: [CGO_ENABLED=0]
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ldflags:
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}
```

### Safety Net (`publish.yml`)

If a GitHub Release is created manually (not via `release.yml`), runs `npm run release:verify` to surface issues. Does NOT publish.

## Test Taxonomy

| Layer | What | Where | Gate |
|-------|------|-------|------|
| JS unit tests | Jest test suite (1916 tests) | `npm test` | CI: `test` job |
| JS format | Prettier | `npm run format:check` | CI: `test` job |
| JS lint | ESLint | `npm run lint` | CI: `test` job |
| Go unit tests | All internal + cmd packages | `go test ./internal/... ./cmd/...` | CI: `go-test` job |
| Go race detection | Data race detector | `-race` flag | CI: `go-test` job |
| Golden/snapshot | 4 canonical command stability | `TestSnapshot_*` | CI: `go-test` job |
| CLI smoke | Binary produces valid output | `terrain version/analyze/insights` | CI: `go-test` job |
| Fixture matrix | 5 diverse repo shapes | `analyze --json` per fixture | CI: `go-test` job |
| Benchmark smoke | 4 commands, 3 fixtures | JSON structure assertions | CI: `go-test` job |
| Go benchmarks | Performance regression | `benchstat` base vs head | CI: `go-bench-compare` (PRs) |
| Integration | JS CLI + conversion | Round-trip conversion test | CI: `build` job |
| PR analysis | Terrain analyzes itself | `terrain pr` + targeted tests | `terrain-pr.yml` |
| Security | CodeQL static analysis | Go, JS/TS, Python | `codeql.yml` |

## Local Development Commands

```bash
# JavaScript
npm test                    # Run all JS tests
npm run lint                # Lint JS source
npm run format              # Format JS source
npm run format:check        # Check JS formatting

# Go
go test ./internal/... ./cmd/... -count=1 -race  # All Go tests
go test ./cmd/terrain/ -run TestSnapshot -v       # Golden tests only
go build -o terrain ./cmd/terrain                 # Build binary

# Benchmark
go build -o terrain-bench ./cmd/terrain-bench     # Build bench harness
./terrain-bench --repo terrain --sequential       # Run against self
./terrain-bench --output benchmarks/output/       # Write results

# Release verification
npm run release:verify      # Full pre-release check
```

## Exit Codes

| Code | Meaning | Used By |
|------|---------|---------|
| 0 | Success | All commands |
| 1 | Runtime error | All commands |
| 2 | Policy violation / usage error | `terrain policy check`, invalid args |
