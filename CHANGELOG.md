# Changelog

## Unreleased

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

- `hamlet migrate` — full project migration with state tracking
- `hamlet estimate` — preview migration complexity
- `hamlet status` / `hamlet checklist` — track migration progress
- Dependency-ordered conversion (helpers before tests)
- Resume interrupted migrations with `--continue`

### Config Conversion

- `hamlet convert-config` — convert framework configuration files
- Supports Jest, Vitest, Cypress, Playwright, WebdriverIO, Mocha configs

### CLI Polish

- **50 shorthand commands** for all 25 conversion directions
- **Batch mode** — convert directories and glob patterns
- **Enhanced dry-run** — confidence reports and file counts
- **`--on-error`** — skip, fail, or best-effort error handling
- **`--json`** — machine-readable output for CI
- **`--quiet` / `--verbose`** — output control
- **`hamlet list`** — categorized conversion directory
- **`hamlet doctor`** — diagnostic command
- TTY-aware progress bar

### Pipeline Architecture

- Framework-neutral intermediate representation (IR)
- Confidence scoring for every conversion
- HAMLET-TODO markers for unconvertible patterns
- Pattern-based parsing and emission

## 1.0.0

### Initial Release

- 6 conversion directions: Cypress, Playwright, Selenium (all pairs)
- CLI with `convert`, `detect`, `validate`, `init` commands
- Programmatic API via `ConverterFactory`
- TypeScript type definitions
- Auto-detection of source framework
- Batch processing for directories
