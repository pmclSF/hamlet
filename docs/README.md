# Hamlet Documentation

> To be or not to be... in any framework you choose.

Hamlet is a multi-framework test converter supporting **25 conversion directions** across **16 frameworks** in **JavaScript**, **Java**, and **Python**.

## Quick Links

- [Getting Started](./guides/getting-started.md)
- [Migration Guide](./guides/migration-guide.md)
- [CLI Reference](./api/cli.md)
- [Configuration](./api/configuration.md)
- [Conversion Process](./api/conversion.md)
- [ADR: Jest ESM Strategy](./adr/004-jest-esm-strategy.md)
- [Release Process](./releasing.md)

## Supported Frameworks

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

## Project Structure

```
src/
├── cli/               # CLI shorthand definitions and output helpers
├── core/              # BaseConverter, ConverterFactory, PatternEngine,
│                      #   FrameworkDetector, FrameworkRegistry, ConfigConverter,
│                      #   MigrationEngine, Scanner, FileClassifier, ir.js
├── converters/        # 6 E2E converter classes (Cypress/Playwright/Selenium pairs)
├── converter/         # Batch processing, orchestration, validation, TypeScript support
├── languages/         # Framework definitions organized by language:
│   ├── java/          #   junit4, junit5, testng
│   ├── javascript/    #   cypress, jest, mocha, jasmine, playwright, vitest,
│   │                  #   puppeteer, testcafe, webdriverio
│   └── python/        #   pytest, unittest_fw, nose2
├── patterns/commands/ # Regex pattern definitions (assertions, navigation, selectors)
├── server/            # Built-in dev server for hamlet serve / hamlet open
├── ui/                # Web UI for interactive conversion
├── utils/             # helpers.js, reporter.js
├── types/             # TypeScript type definitions
└── index.js           # Main entry point with public API
```

## Examples

See the [examples](./examples/) directory for sample conversions:

- [E2E Tests](./examples/e2e/)
- [API Tests](./examples/api/)

## License

MIT License - see the [LICENSE](../LICENSE) file for details.
