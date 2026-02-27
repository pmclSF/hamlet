# Getting Started with Hamlet

## Installation

```bash
npm install -g hamlet-converter
```

Requires Node.js >= 22.0.0.

## Basic Usage

### Convert a single file

```bash
hamlet convert auth.test.js --from jest --to vitest -o converted/
```

Or use a shorthand:

```bash
hamlet jest2vt auth.test.js -o converted/
```

### Convert a directory

```bash
hamlet convert tests/ --from jest --to vitest -o converted/
```

### Preview before converting

```bash
hamlet estimate tests/ --from jest --to vitest
```

This shows how many files would be converted and an estimated confidence score, without writing any files.

## Multi-Framework Usage

Hamlet supports 25 conversion directions across JavaScript, Java, and Python.

### JavaScript Unit Tests

```bash
# Jest to Vitest
hamlet jest2vt auth.test.js -o converted/

# Mocha to Jest
hamlet mocha2jest utils.test.js -o converted/

# Jasmine to Jest
hamlet jas2jest spec/app.spec.js -o converted/
```

### JavaScript E2E Tests

```bash
# Cypress to Playwright
hamlet cy2pw cypress/e2e/login.cy.js -o tests/

# Playwright to Selenium
hamlet pw2sel tests/login.spec.ts -o selenium/
```

### Java

```bash
# JUnit 4 to JUnit 5
hamlet ju42ju5 LoginTest.java -o converted/

# JUnit 5 to TestNG
hamlet ju52tng LoginTest.java -o converted/
```

### Python

```bash
# pytest to unittest
hamlet pyt2ut test_auth.py -o converted/

# nose2 to pytest
hamlet nose22pyt test_utils.py -o converted/
```

Run `hamlet list` to see all 25 conversion directions with their shorthands.

## Common Scenarios

### Dry run (preview without writing)

```bash
hamlet convert tests/ --from jest --to vitest -o converted/ --dry-run
```

### Full project migration

```bash
hamlet migrate tests/ --from jest --to vitest -o converted/
```

The `migrate` command provides state tracking, dependency ordering, and config conversion on top of `convert`.

### Auto-detect source framework

```bash
hamlet detect auth.test.js
# Output: jest (confidence: 95%)
```

### Convert config files

```bash
hamlet convert-config jest.config.js --to vitest -o vitest.config.js
```

### JSON output for CI

```bash
hamlet jest2vt auth.test.js -o converted/ --json
```

## Understanding Output

### Confidence scores

Every conversion produces a confidence score (0-100%):

- **High (80-100%)**: Fully automated, ready to use
- **Medium (50-79%)**: Mostly automated, review HAMLET-TODO markers
- **Low (0-49%)**: Significant manual work needed

### HAMLET-TODO markers

When a pattern can't be automatically converted, Hamlet inserts a comment:

```javascript
// HAMLET-TODO: cy.session() has no direct equivalent in Playwright
// Original: cy.session('admin', () => { ... })
```

Search for `HAMLET-TODO` after conversion to find patterns that need manual attention.

## Next Steps

- [Migration Guide](./migration-guide.md) - full project migration workflow
- [CLI Reference](../api/cli.md) - all commands and options
- [Configuration](../api/configuration.md) - CLI flags and programmatic API
- [Conversion Process](../api/conversion.md) - how conversion works under the hood
