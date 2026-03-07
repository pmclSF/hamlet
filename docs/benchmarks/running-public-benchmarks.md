# Running Public Benchmarks

## Prerequisites

- Go 1.23+ (to build Hamlet)
- Python 3.9+ (for YAML parsing and summarization)
- `pyyaml` Python package: `pip3 install pyyaml`
- Git
- ~3GB free disk space for full matrix, ~500MB for smoke

## Quick start

```bash
# 1. Fetch smoke-tier repos (~100MB)
make benchmark-fetch
# Or just smoke tier:
./scripts/benchmarks/fetch_public_repos.sh --tier smoke

# 2. Run smoke benchmark (~1-2 minutes)
make benchmark-smoke

# 3. View results
cat artifacts/public-benchmarks/summary.md
```

## Tiers

| Tier | Repos | Estimated time | Disk space |
|------|-------|---------------|------------|
| smoke | express, fastify | 1-2 min | ~100MB |
| full | + jest, playwright, vue, flask, next.js | 5-15 min | ~2GB |
| stress | + storybook | 15-30 min | ~3GB |

## Commands

### Fetch repos

```bash
make benchmark-fetch                              # all repos
./scripts/benchmarks/fetch_public_repos.sh --tier smoke   # smoke only
./scripts/benchmarks/fetch_public_repos.sh --id express   # one repo
./scripts/benchmarks/fetch_public_repos.sh --force        # re-clone
./scripts/benchmarks/fetch_public_repos.sh --full-clone   # full history
```

### Run benchmarks

```bash
make benchmark-smoke                              # smoke tier
make benchmark-full                               # full tier
make benchmark-stress                             # all tiers

# Run one repo
./scripts/benchmarks/run_public_matrix.sh full --id jest

# Skip determinism check for speed
./scripts/benchmarks/run_public_matrix.sh smoke --skip-determinism
```

### View results

```bash
make benchmark-summary                            # regenerate summary
cat artifacts/public-benchmarks/summary.md        # human-readable
cat artifacts/public-benchmarks/summary.json      # machine-readable
cat artifacts/public-benchmarks/express/analyze_json.stdout  # raw output
```

### Update repos

```bash
./scripts/benchmarks/update_public_repos.sh       # pull all
./scripts/benchmarks/update_public_repos.sh --id jest  # pull one
```

## Adding a new public repo

1. Add an entry to `benchmarks/public-repos.yaml`:
   ```yaml
   - id: my-repo
     url: https://github.com/org/repo.git
     branch: main
     tier: full
     category: backend-js
     clone: shallow
     description: Brief description of why this repo is valuable.
     expected:
       min_test_files: 10
       min_code_units: 5
   ```

2. Create expectations at `benchmarks/expectations/my-repo.yaml`:
   ```yaml
   min_test_files: 10
   min_code_units: 5
   require_posture: true
   analyze_must_succeed: true
   ```

3. Fetch and test:
   ```bash
   ./scripts/benchmarks/fetch_public_repos.sh --id my-repo
   ./scripts/benchmarks/run_public_matrix.sh full --id my-repo
   ```

4. Update `docs/benchmarks/public-benchmark-matrix.md` with the new entry.

## Interpreting results

### Summary table columns

| Column | Meaning |
|--------|---------|
| Status | pass/fail/degraded — overall health |
| Duration | Wall-clock time for all commands |
| Tests | Number of test files detected |
| Units | Number of code units discovered |
| FWs | Number of frameworks detected |
| Determ | pass/fail — determinism check |
| Expect | pass/fail — expectation check |

### Failure types

- **Command failure (exit ≠ 0)**: Hamlet crashed or errored. This is likely a bug.
- **Expectation miss**: Fewer tests/units than expected. Either the repo changed or Hamlet regressed.
- **Determinism failure**: Two runs of the same command produced different structured output. Investigate — could be map ordering, timestamps leaking, or non-deterministic logic.
- **Degraded**: Not a hard failure, but something is off (usually determinism).

### Artifacts

Each repo produces artifacts under `artifacts/public-benchmarks/<repo-id>/`:

```
artifacts/public-benchmarks/express/
  analyze_json.stdout       # JSON snapshot
  analyze_json.stderr       # stderr
  analyze_json.meta         # exit code, duration, timestamp
  analyze_text.stdout       # human-readable output
  summary.stdout
  posture.stdout
  metrics_json.stdout
  export.stdout
  determinism_run1.json     # first determinism run
  determinism_run2.json     # second determinism run
  determinism.meta          # pass/fail
  expectations.meta         # expectation check results
```

## Disk space and clone time

| Repo | Shallow clone size | Full clone size |
|------|-------------------|-----------------|
| express | ~15MB | ~50MB |
| fastify | ~20MB | ~80MB |
| jest | ~100MB | ~400MB |
| playwright | ~200MB | ~1GB |
| vue | ~30MB | ~150MB |
| flask | ~10MB | ~40MB |
| next.js | ~500MB | ~3GB |
| storybook | ~300MB | ~1.5GB |

Sizes are approximate and grow over time.

## Troubleshooting

**"python3 is required"**: Install Python 3.9+ and ensure `python3` is in PATH.

**"No module named 'yaml'"**: Run `pip3 install pyyaml`.

**Clone fails**: Check network connectivity. Some repos may be temporarily unavailable. The script continues on failure.

**Analysis takes too long**: Large repos (next.js, storybook) can take minutes. Consider using `--skip-determinism` or `--tier smoke` for quick feedback.

**Shallow clone can't pull**: Run `./scripts/benchmarks/fetch_public_repos.sh --id <id> --force` to re-clone.
