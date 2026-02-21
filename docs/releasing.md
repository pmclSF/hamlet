# Releasing Hamlet

## Release Flow

```
git tag v2.x.x && git push origin v2.x.x
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
          ├── Create GitHub Release (auto-generated notes)
          └── npm publish (NPM_TOKEN secret)
```

A single workflow (`release.yml`) handles the full pipeline: verify → release →
publish. The release job only runs if verify passes.

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
- `src/types/index.d.ts`
- `README.md`
- `LICENSE`
- `package.json`

No test files, no `.github/`, no `node_modules/`, no `.env`.

### 3. Tag and push

```bash
git tag v2.x.x
git push origin v2.x.x
```

### 4. Verify after push

- [ ] GitHub Actions: `release.yml` completed successfully
- [ ] GitHub Releases: new release exists with auto-generated notes
- [ ] npm: `npm view hamlet-converter version` shows the new version
- [ ] Install test: `npx hamlet-converter@latest --version` prints the new version

## Tag Naming Convention

Tags follow semver prefixed with `v`:

```
v2.0.0      # major
v2.1.0      # minor (new features, no breaking changes)
v2.1.1      # patch (bug fixes only)
```

The `v` prefix is required — `release.yml` triggers on `push: tags: ['v*']`.
