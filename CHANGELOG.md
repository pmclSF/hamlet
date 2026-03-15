# Changelog

## 3.0.0 — V3 Signal-First Test Intelligence (unreleased)

Terrain V3 is a complete architectural shift from V2's conversion-led approach
to a signal-first test intelligence platform built in Go.

### Core Analysis
- Repository scanning with framework detection (16 JS/TS/Java/Python frameworks)
- Test file discovery and code unit extraction
- Signal-first architecture: every finding is a structured Signal with type, severity, evidence, and location
- Evidence model with strength (strong/moderate/weak), source, and confidence scoring

### Signal Detectors
- **Quality**: weak assertions, mock-heavy tests, untested exports, coverage threshold breaks
- **Migration**: deprecated patterns, dynamic test generation, custom matchers, unsupported setup, framework fragmentation
- **Governance**: policy violations, legacy framework usage, runtime budget exceeded
- **Health**: slow tests, flaky tests, skipped tests (runtime-backed)

### Risk Modeling
- Explainable risk engine with reliability, change, and speed dimensions
- Risk surfaces by file, directory, owner, and repository scope
- Heatmap model with directory and owner hotspots

### Migration Intelligence
- `terrain migration readiness` — readiness assessment with quality factors and area assessments
- `terrain migration blockers` — blockers by type and area with representative examples
- `terrain migration preview` — file-level and scope-level migration difficulty preview
- Target framework inference (jest to vitest, cypress to playwright, etc.)

### Executive Summary
- Posture summary by risk dimension with evidence-based recommendations
- Blind spots, known-limitations, and benchmark readiness sections
- Trend highlights from snapshot comparison

### Artifact Ingestion
- Runtime: JUnit XML and Jest JSON parsers
- Coverage: LCOV and Istanbul JSON parsers with code unit attribution
- Graceful degradation when artifacts are absent

### Snapshot and Comparison
- `terrain analyze --write-snapshot` — persist snapshots for trend tracking
- `terrain compare` — snapshot-to-snapshot comparison with signal deltas and risk band changes

### Impact Analysis
- `terrain impact` — change-scope analysis against git diff
- Drill-down views: units, gaps, tests, owners

### VS Code Extension
- Sidebar views: Overview, Health, Quality, Migration, Review
- TreeDataProvider implementations over CLI JSON output
- Commands: refresh, open summary, show migration blockers, reveal file

### Packaging
- goreleaser config for multi-platform binaries
- `terrain version` with build metadata
- 25 internal Go packages

---

## Unreleased (V2)

### Prepublish Hardening

- **docs**: Fix README "4 languages" → "3 languages" (JavaScript, Java, Python)
- **types**: Align `src/types/index.d.ts` with actual main entry exports — add missing `VERSION`, `DEFAULT_OPTIONS`, `SUPPORTED_TEST_TYPES`, `validateTests`, `generateReport`, `processTestFiles`, `ConversionReporter`, `BatchProcessor`, utility namespaces
- **types**: Add consumer type-test (`test/types/consumer.ts`) and `npm run typecheck` script; integrated into `release:verify`
- **cli**: Generate `DIRECTIONS` in `shorthands.js` from `ConverterFactory.getSupportedConversions()` instead of hardcoded array — single source of truth
- **test**: Add shorthands↔ConverterFactory sync test (4 assertions)
- **test**: Add all-25-directions smoke test — converter creation + non-empty output for every direction
- **test**: Add migration state resume/retry/idempotency tests (9 tests)

## 2.0.0

### New Frameworks

- **WebdriverIO** — bidirectional with Playwright and Cypress
- **Puppeteer** — bidirectional with Playwright
- **TestCafe** — converts to Playwright and Cypress
- **Mocha** — bidirectional with Jest
- **Jasmine** — bidirectional with Jest
- **JUnit 4** — converts to JUnit 5
- **JUnit 5** — bidirectional with TestNG
- **TestNG** — bidirectional with JUnit 5
- **pytest** — bidirectional with unittest
- **unittest** — bidirectional with pytest
- **nose2** — converts to pytest

### Migration Tool

- `terrain migrate` — full project migration with state tracking
- `terrain estimate` — preview migration complexity
- `terrain status` / `terrain checklist` — track migration progress
- Dependency-ordered conversion (helpers before tests)
- Resume interrupted migrations with `--continue`

### Config Conversion

- `terrain convert-config` — convert framework configuration files
- Supports Jest, Vitest, Cypress, Playwright, WebdriverIO, Mocha configs

### CLI Polish

- **50 shorthand commands** for all 25 conversion directions
- **Batch mode** — convert directories and glob patterns
- **Enhanced dry-run** — confidence reports and file counts
- **`--on-error`** — skip, fail, or best-effort error handling
- **`--json`** — machine-readable output for CI
- **`--quiet` / `--verbose`** — output control
- **`terrain list`** — categorized conversion directory
- **`terrain doctor`** — diagnostic command
- TTY-aware progress bar

### Pipeline Architecture

- Framework-neutral intermediate representation (IR)
- Confidence scoring for every conversion
- TERRAIN-TODO markers for unconvertible patterns
- Pattern-based parsing and emission

## 1.0.0

### Initial Release

- 6 conversion directions: Cypress, Playwright, Selenium (all pairs)
- CLI with `convert`, `detect`, `validate`, `init` commands
- Programmatic API via `ConverterFactory`
- TypeScript type definitions
- Auto-detection of source framework
- Batch processing for directories
