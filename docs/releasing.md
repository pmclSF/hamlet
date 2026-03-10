# Releasing Hamlet

## Release Flow

```
git tag vX.Y.Z && git push origin vX.Y.Z
         │
         ▼
  release.yml (trigger: tag push v*)
    ├── verify job:
    │     ├── npm ci
    │     ├── Assert tag matches package.json version
    │     └── npm run release:verify
    │           ├── format:check
    │           ├── lint
    │           ├── test (all suites)
    │           └── verify-pack.js
    │                 ├── npm pack → install in temp dir
    │                 ├── Verify JS exports (VERSION, convertFile, …)
    │                 ├── CLI smoke (--version, --help)
    │                 └── Conversion smoke (jest→vitest)
    └── release job (needs: verify):
          ├── npm ci
          ├── npm publish --provenance (NPM_TOKEN secret)
          └── Create GitHub Release (auto-generated notes)
```

A single workflow (`release.yml`) handles the full pipeline: verify → npm release →
Go binary release. Both release jobs only run if verify passes.

### Go Binary Release

The `go-release` job uses [GoReleaser](https://goreleaser.com/) to build
multi-platform binaries (Linux/macOS/Windows × amd64/arm64) and attach them
to the GitHub Release. Configuration lives in `.goreleaser.yaml`.

Binaries are stamped with version, commit, and build date via ldflags:
```
-X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}
```

Users can install via:
```bash
go install github.com/pmclSF/hamlet/cmd/hamlet@latest
```
Or download pre-built binaries from the GitHub Releases page.

The verify job includes a **tag/version guard** that aborts if the git tag
version does not match `package.json` version — preventing accidental publishes
of mismatched versions.

`publish.yml` (renamed "Verify Release") is a safety net that triggers on
`release: created` events. It runs `npm run release:verify` but does NOT
publish — this catches issues if a release is created manually outside the
tag-push flow.

### Required Secrets

| Secret | Where | Purpose |
|--------|-------|---------|
| `NPM_TOKEN` | GitHub repo → Settings → Secrets | npm automation token with publish access |

### Permissions

| Workflow | Permission | Why |
|----------|-----------|-----|
| `release.yml` | `contents: write` | Create GitHub Release |
| `release.yml` | `id-token: write` | npm provenance attestation (`npm publish --provenance`) |
| `publish.yml` | `contents: read` | Read repo (verify only, no publish) |

## Dry-Run Checklist

Follow these steps before cutting any release.

### Prerequisites

- [ ] You are on `main` with a clean working tree (`git status` shows nothing)
- [ ] CI is green on the latest commit (check GitHub Actions)
- [ ] `CHANGELOG.md` is up to date for the new version
- [ ] `package.json` version matches the tag you are about to create

### 1. Run the full verification locally

```bash
npm run release:verify
```

This runs format:check, lint, tests, packs the tarball, installs it in a temp
directory, verifies exports, runs CLI smoke tests (`--version`, `--help`), and
runs a conversion smoke test (jest→vitest).

### 2. Inspect tarball contents

```bash
npm pack --dry-run
```

Confirm only expected files are included:
- `bin/hamlet.js`
- `src/**/*.js`
- `src/types/*.d.ts`
- `README.md`
- `SECURITY.md`
- `LICENSE`
- `package.json`

No test files, no `.github/`, no `node_modules/`, no `.env`.

### 3. Tag and push

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

### 4. Verify after push

- [ ] GitHub Actions: `release.yml` completed successfully
- [ ] GitHub Releases: new release exists with auto-generated notes
- [ ] npm: `npm view hamlet-testframework version` shows the new version
- [ ] Install test: `npx hamlet-testframework@latest --version` prints the new version

## Tag Naming Convention

Tags follow semver prefixed with `v`:

```
vX.0.0      # major
vX.Y.0      # minor (new features, no breaking changes)
vX.Y.Z      # patch (bug fixes only)
```

The `v` prefix is required — `release.yml` triggers on `push: tags: ['v*']`.
