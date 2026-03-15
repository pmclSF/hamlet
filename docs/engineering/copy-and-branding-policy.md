# Copy and Branding Policy

This policy prevents release-line branding drift in Terrain's active product surfaces.

## Rule

Do not use numbered release-line labels in active copy/content.

Applies to:
- Product and contributor Markdown docs
- CLI help output
- JSON outputs intended for automation or benchmarking

## Intentional Exceptions

These surfaces are intentionally excluded from the guardrail:
- `CHANGELOG.md` (historical release chronology)
- `docs/legacy/**` (archived historical documentation)
- `test/**`, `tests/**`, `fixtures/**` (non-customer fixture content)

Technical protocol/dependency versions are also exempt when they are contractually required, for example:
- External API path versions
- Dependency module versions

## Enforcement

Run locally:

```bash
python3 scripts/verify_copy_policy.py
```

CI enforcement:
- `.github/workflows/ci.yml` runs this check in the Go job.
- `make verify-copy-policy` runs the same check locally.
- `make pr-gate` depends on `verify-copy-policy`.

## Output-Specific Guard

In addition to text scanning, the guard validates that:
- `terrain --help` contains no release-line branding
- `terrain metrics --json` and `terrain export benchmark` contain no release-line branding
- `analysisVersion` does not use a `v<digit>` prefix
