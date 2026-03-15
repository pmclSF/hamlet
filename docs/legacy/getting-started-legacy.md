> **Legacy document.** This describes the legacy JavaScript converter engine. For the current engine, see the [CLI spec](../cli-spec.md) and [architecture overview](../architecture/00-overview.md).

# Getting Started with Terrain (Legacy Converter)

## Installation

```bash
npm install -g terrain-converter
```

Requires Node.js >= 22.0.0.

## Basic Usage

### Convert a single file

```bash
terrain convert auth.test.js --from jest --to vitest -o converted/
```

Or use a shorthand:

```bash
terrain jest2vt auth.test.js -o converted/
```

### Convert a directory

```bash
terrain convert tests/ --from jest --to vitest -o converted/
```

### Preview before converting

```bash
terrain estimate tests/ --from jest --to vitest
```

This shows how many files would be converted and an estimated confidence score, without writing any files.

## Multi-Framework Usage

Terrain supports 25 conversion directions across JavaScript, Java, and Python.

### JavaScript Unit Tests

```bash
# Jest to Vitest
terrain jest2vt auth.test.js -o converted/

# Mocha to Jest
terrain mocha2jest utils.test.js -o converted/

# Jasmine to Jest
terrain jas2jest spec/app.spec.js -o converted/
```

### JavaScript E2E Tests

```bash
# Cypress to Playwright
terrain cy2pw cypress/e2e/login.cy.js -o tests/

# Playwright to Selenium
terrain pw2sel tests/login.spec.ts -o selenium/
```

### Java

```bash
# JUnit 4 to JUnit 5
terrain ju42ju5 LoginTest.java -o converted/

# JUnit 5 to TestNG
terrain ju52tng LoginTest.java -o converted/
```

### Python

```bash
# pytest to unittest
terrain pyt2ut test_auth.py -o converted/

# nose2 to pytest
terrain nose22pyt test_utils.py -o converted/
```

Run `terrain list` to see all 25 conversion directions with their shorthands.

## Common Scenarios

### Dry run (preview without writing)

```bash
terrain convert tests/ --from jest --to vitest -o converted/ --dry-run
```

### Full project migration

```bash
terrain migrate tests/ --from jest --to vitest -o converted/
```

The `migrate` command provides state tracking, dependency ordering, and config conversion on top of `convert`.

### Auto-detect source framework

```bash
terrain detect auth.test.js
# Output: jest (confidence: 95%)
```

### Convert config files

```bash
terrain convert-config jest.config.js --to vitest -o vitest.config.js
```

### JSON output for CI

```bash
terrain jest2vt auth.test.js -o converted/ --json
```

## Understanding Output

### Confidence scores

Every conversion produces a confidence score (0-100%):

- **High (80-100%)**: Fully automated, ready to use
- **Medium (50-79%)**: Mostly automated, review TERRAIN-TODO markers
- **Low (0-49%)**: Significant manual work needed

### TERRAIN-TODO markers

When a pattern can't be automatically converted, Terrain inserts a comment:

```javascript
// TERRAIN-TODO: cy.session() has no direct equivalent in Playwright
// Original: cy.session('admin', () => { ... })
```

Search for `TERRAIN-TODO` after conversion to find patterns that need manual attention.

## Next Steps

- [Migration Guide](./migration-guide-legacy.md) - full project migration workflow
- [CLI Reference](./cli-reference-legacy.md) - all commands and options
- [Configuration](./configuration-legacy.md) - CLI flags and programmatic API
- [Conversion Process](./conversion-process-legacy.md) - how conversion works under the hood
