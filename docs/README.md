# Hamlet Documentation

> To be or not to be... in any framework you choose.

Hamlet is a bidirectional test converter supporting **Cypress**, **Playwright**, and **Selenium**.

## Quick Links

- [Getting Started](./guides/getting-started.md)
- [Migration Guide](./guides/migration-guide.md)
- [CLI Reference](./api/cli.md)
- [Configuration](./api/configuration.md)
- [Conversion API](./api/conversion.md)

## Supported Conversions

| From | To | Command |
|------|-----|---------|
| Cypress | Playwright | `hamlet convert --from cypress --to playwright` |
| Cypress | Selenium | `hamlet convert --from cypress --to selenium` |
| Playwright | Cypress | `hamlet convert --from playwright --to cypress` |
| Playwright | Selenium | `hamlet convert --from playwright --to selenium` |
| Selenium | Cypress | `hamlet convert --from selenium --to cypress` |
| Selenium | Playwright | `hamlet convert --from selenium --to playwright` |

## Examples

See the [examples](./examples/) directory for sample conversions:

- [E2E Tests](./examples/e2e/)
- [API Tests](./examples/api/)

## Project Structure

```
src/
├── converters/          # Framework-specific converters
│   ├── CypressToPlaywright.js
│   ├── CypressToSelenium.js
│   ├── PlaywrightToCypress.js
│   ├── PlaywrightToSelenium.js
│   ├── SeleniumToCypress.js
│   └── SeleniumToPlaywright.js
├── core/                # Core conversion infrastructure
│   ├── BaseConverter.js
│   ├── ConverterFactory.js
│   ├── FrameworkDetector.js
│   └── PatternEngine.js
├── patterns/            # Command pattern mappings
│   └── commands/
├── converter/           # Legacy converters and utilities
├── utils/               # Helper utilities
└── index.js             # Main entry point
```

## Contributing

Contributions are welcome! Please read our Contributing Guide for details.

## License

MIT License - see the [LICENSE](../LICENSE) file for details.
