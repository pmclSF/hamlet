# Releasing Hamlet

## Release Flow

```
git tag v2.x.x && git push origin v2.x.x
         │
         ▼
  release.yml (trigger: tag push v*)
    ├── npm ci
    ├── npm test
    ├── Create GitHub Release (auto-generated notes)
    └── npm publish (NPM_TOKEN secret)
```

A single workflow (`release.yml`) handles both the GitHub Release and npm publish.
`publish.yml` exists as a fallback — it triggers on `release: created` events and
runs the same publish step. If `release.yml` already published successfully,
`publish.yml` will no-op (npm rejects duplicate version publishes).

### Required Secrets

| Secret | Where | Purpose |
|--------|-------|---------|
| `NPM_TOKEN` | GitHub repo → Settings → Secrets | npm automation token with publish access |

### Permissions

| Workflow | Permission | Why |
|----------|-----------|-----|
| `release.yml` | `contents: write` | Create GitHub Release |
| `publish.yml` | `contents: read`, `id-token: write` | Read repo, npm provenance (future) |

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
directory, and verifies the VERSION export matches package.json.

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

### 3. Verify consumer import (manual, if needed)

```bash
TARBALL=$(npm pack --pack-destination /tmp)
mkdir /tmp/hamlet-verify && cd /tmp/hamlet-verify
npm init -y
npm install /tmp/$TARBALL
node --input-type=module -e "import { VERSION, ConverterFactory } from 'hamlet-converter'; console.log('VERSION:', VERSION); console.log('ConverterFactory:', typeof ConverterFactory);"
cd - && rm -rf /tmp/hamlet-verify /tmp/$TARBALL
```

### 4. Tag and push

```bash
git tag v2.x.x
git push origin v2.x.x
```

### 5. Verify after push

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
