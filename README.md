# Hamlet

[![CI](https://github.com/pmclSF/hamlet/actions/workflows/ci.yml/badge.svg)](https://github.com/pmclSF/hamlet/actions/workflows/ci.yml)
[![npm version](https://badge.fury.io/js/hamlet-converter.svg)](https://www.npmjs.com/package/hamlet-converter)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Migrate your test suites between frameworks with confidence.

**25 conversion directions** across **16 frameworks** in **4 languages** (JavaScript, Java, Python).

## Node Support

Hamlet supports **active and maintenance LTS** versions of Node.js. Currently: **Node 22** and **Node 24**.

CI tests run on both. When a Node major reaches end-of-life, it is dropped in the next minor release.

## Quick Start

```bash
npm install -g hamlet-converter

# Convert a single file
hamlet jest2vt auth.test.js -o converted/

# Preview a migration
hamlet estimate tests/ --from jest --to vitest

# Migrate your project
hamlet migrate tests/ --from jest --to vitest -o converted/
```

## Supported Conversions

### JavaScript Unit Testing

| Direction | Shorthand |
|-----------|-----------|
| Jest &rarr; Vitest | `hamlet jest2vt` |
| Mocha &rarr; Jest | `hamlet mocha2jest` |
| Jasmine &rarr; Jest | `hamlet jas2jest` |
| Jest &rarr; Mocha | `hamlet jest2mocha` |
| Jest &rarr; Jasmine | `hamlet jest2jas` |

### JavaScript E2E / Browser

| Direction | Shorthand |
|-----------|-----------|
| Cypress &harr; Playwright | `hamlet cy2pw` / `hamlet pw2cy` |
| Cypress &harr; Selenium | `hamlet cy2sel` / `hamlet sel2cy` |
| Playwright &harr; Selenium | `hamlet pw2sel` / `hamlet sel2pw` |
| Cypress &harr; WebdriverIO | `hamlet cy2wdio` / `hamlet wdio2cy` |
| Playwright &harr; WebdriverIO | `hamlet pw2wdio` / `hamlet wdio2pw` |
| Puppeteer &harr; Playwright | `hamlet pptr2pw` / `hamlet pw2pptr` |
| TestCafe &rarr; Playwright | `hamlet tcafe2pw` |
| TestCafe &rarr; Cypress | `hamlet tcafe2cy` |

### Java

| Direction | Shorthand |
|-----------|-----------|
| JUnit 4 &rarr; JUnit 5 | `hamlet ju42ju5` |
| JUnit 5 &harr; TestNG | `hamlet ju52tng` / `hamlet tng2ju5` |

### Python

| Direction | Shorthand |
|-----------|-----------|
| pytest &harr; unittest | `hamlet pyt2ut` / `hamlet ut2pyt` |
| nose2 &rarr; pytest | `hamlet nose22pyt` |

Run `hamlet list` to see all directions with their shorthand aliases.

## Commands

### Convert

Convert a single file, directory, or glob pattern:

```bash
# Single file
hamlet convert auth.test.js --from jest --to vitest -o converted/

# Directory (requires --output)
hamlet convert tests/ --from jest --to vitest -o converted/

# Glob pattern
hamlet convert "tests/**/*.test.js" --from jest --to vitest -o converted/

# Shorthand (equivalent to convert --from jest --to vitest)
hamlet jest2vt auth.test.js -o converted/
```

### Migrate

Full project migration with state tracking, dependency ordering, and config conversion:

```bash
hamlet migrate tests/ --from jest --to vitest -o converted/

# Resume an interrupted migration
hamlet migrate tests/ --from jest --to vitest -o converted/ --continue

# Retry only previously failed files
hamlet migrate tests/ --from jest --to vitest -o converted/ --retry-failed
```

### Estimate

Preview migration complexity without converting:

```bash
hamlet estimate tests/ --from jest --to vitest
```

### Dry Run

Preview what would happen without writing files:

```bash
hamlet convert tests/ --from jest --to vitest -o converted/ --dry-run
hamlet migrate tests/ --from jest --to vitest --dry-run
```

### Other Commands

```bash
hamlet list              # Show all conversion directions with shorthands
hamlet shorthands        # List all shorthand command aliases
hamlet detect file.js    # Auto-detect testing framework from a file
hamlet doctor            # Run diagnostics
hamlet status -d .       # Show current migration progress
hamlet checklist -d .    # Generate migration checklist
hamlet reset -d . --yes  # Clear migration state
```

## Options

| Option | Description |
|--------|-------------|
| `-o, --output <path>` | Output path (required for directories) |
| `-f, --from <framework>` | Source framework |
| `-t, --to <framework>` | Target framework |
| `--dry-run` | Preview without writing files |
| `--on-error <mode>` | Error handling: `skip` (default), `fail`, `best-effort` |
| `-q, --quiet` | Suppress non-error output |
| `--verbose` | Show detailed per-pattern output |
| `--json` | Machine-readable JSON output |
| `--no-color` | Disable color output |
| `--auto-detect` | Auto-detect source framework |

## JSON Output

For CI integration, use `--json` for machine-readable output:

```bash
hamlet jest2vt auth.test.js -o converted/ --json
```

```json
{
  "success": true,
  "files": [{ "source": "auth.test.js", "output": "converted/auth.test.js", "confidence": 95 }],
  "summary": { "converted": 1, "skipped": 0, "failed": 0 }
}
```

## How It Works

1. **Detect** &mdash; determine source framework from content (regex heuristics per framework)
2. **Parse** &mdash; classify source lines into IR nodes (suites, tests, hooks, assertions, raw code)
3. **Transform** &mdash; apply regex-based pattern substitutions to convert API calls and test structure
4. **Score** &mdash; walk the IR tree to calculate confidence (converted vs. unconvertible nodes)
5. **Report** &mdash; generate HAMLET-TODO markers for patterns that need manual review

> **Architecture note:** Conversion is currently regex-based string transformation.
> The IR (intermediate representation) captures test structure for confidence scoring
> but emitters operate on the source string, not the IR tree.
> See [DESIGN.md](DESIGN.md) §1 for the hybrid IR + PatternEngine design rationale.

## Confidence Scores

Every conversion produces a confidence score (0-100%):

- **High (80-100%)**: Fully automated, ready to use
- **Medium (50-79%)**: Mostly automated, review HAMLET-TODO markers
- **Low (0-49%)**: Significant manual work needed

## HAMLET-TODO Markers

When a pattern can't be automatically converted, Hamlet inserts a comment:

```javascript
// HAMLET-TODO: cy.session() has no direct equivalent in Playwright
// Original: cy.session('admin', () => { ... })
```

Search for `HAMLET-TODO` after conversion to find patterns that need manual attention.

## Config Conversion

Convert framework configuration files:

```bash
hamlet convert-config jest.config.js --to vitest -o vitest.config.js
hamlet convert-config cypress.config.js --to playwright -o playwright.config.ts
```

## Programmatic API

```javascript
import { ConverterFactory, FRAMEWORKS } from 'hamlet-converter';

const converter = await ConverterFactory.createConverter('jest', 'vitest');
const output = await converter.convert(jestCode);

// Get conversion report
const report = converter.getLastReport();
console.log(`Confidence: ${report.confidence}%`);
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Runtime error (conversion failed) |
| 2 | Invalid arguments (bad framework, missing file) |

## Development

```bash
npm install
npm test                    # Run all tests
npm run lint                # Lint source
npm run format              # Format with Prettier
```

## Requirements

- Node.js >= 16.0.0

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on adding new frameworks.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Links

- [GitHub Repository](https://github.com/pmclSF/hamlet)
- [npm Package](https://www.npmjs.com/package/hamlet-converter)
- [Issue Tracker](https://github.com/pmclSF/hamlet/issues)
