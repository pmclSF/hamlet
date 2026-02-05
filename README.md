# Hamlet

[![CI](https://github.com/pmclSF/hamlet/actions/workflows/ci.yml/badge.svg)](https://github.com/pmclSF/hamlet/actions/workflows/ci.yml)
[![npm version](https://badge.fury.io/js/hamlet-converter.svg)](https://www.npmjs.com/package/hamlet-converter)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Bidirectional multi-framework test converter for **Cypress**, **Playwright**, and **Selenium**.

## Overview

Hamlet enables seamless migration of automated tests between the three most popular testing frameworks. Convert tests in any direction with a single command.

```
┌──────────┐     ┌────────────┐     ┌──────────┐
│  Cypress │ ←→  │ Playwright │ ←→  │ Selenium │
└──────────┘     └────────────┘     └──────────┘
      ↑                                   ↑
      └───────────────────────────────────┘
```

## Features

- **6 Conversion Directions**: Convert between any pair of supported frameworks
- **CLI & Programmatic API**: Use from command line or import in your code
- **TypeScript Support**: Full type definitions included
- **Auto-detection**: Automatically detect source framework from file content
- **Batch Processing**: Convert entire directories of tests
- **Config Conversion**: Convert framework configuration files
- **207 Tests**: Thoroughly tested with comprehensive edge cases

## Installation

```bash
# Global installation
npm install -g hamlet-converter

# Local installation
npm install hamlet-converter
```

## Quick Start

```bash
# Convert Cypress to Playwright
hamlet convert ./tests/login.cy.js --from cypress --to playwright -o ./output

# Convert Playwright to Selenium
hamlet convert ./tests/login.spec.ts --from playwright --to selenium -o ./output

# Convert entire directory
hamlet convert ./cypress/e2e --from cypress --to playwright -o ./playwright-tests

# List all supported conversions
hamlet list-conversions
```

## Supported Conversions

| From | To | Command |
|------|-----|---------|
| Cypress | Playwright | `--from cypress --to playwright` |
| Cypress | Selenium | `--from cypress --to selenium` |
| Playwright | Cypress | `--from playwright --to cypress` |
| Playwright | Selenium | `--from playwright --to selenium` |
| Selenium | Cypress | `--from selenium --to cypress` |
| Selenium | Playwright | `--from selenium --to playwright` |

## CLI Reference

```bash
Usage: hamlet [command] [options]

Commands:
  convert <source>      Convert tests between frameworks
  list-conversions      List all supported conversion directions
  detect <file>         Auto-detect the testing framework from a file
  validate <path>       Validate converted tests
  init                  Initialize Hamlet configuration
  cy2pw <source>        Shorthand for Cypress to Playwright

Options for convert:
  -f, --from <framework>   Source framework (cypress, playwright, selenium)
  -t, --to <framework>     Target framework (cypress, playwright, selenium)
  -o, --output <path>      Output path for converted tests
  --validate               Validate converted tests after conversion
  --dry-run                Show what would be converted without making changes
  --auto-detect            Auto-detect source framework from file content
```

## Conversion Examples

### Cypress → Playwright

**Input:**
```javascript
describe('Login', () => {
  beforeEach(() => {
    cy.visit('/login');
  });

  it('should login successfully', () => {
    cy.get('#username').type('testuser');
    cy.get('#password').type('secret123');
    cy.get('#remember').check();
    cy.get('button[type="submit"]').click();
    cy.get('.welcome').should('be.visible');
    cy.get('.welcome').should('have.text', 'Welcome!');
  });
});
```

**Output:**
```javascript
import { test, expect } from '@playwright/test';

test.describe('Login', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
  });

  test('should login successfully', async ({ page }) => {
    await page.locator('#username').fill('testuser');
    await page.locator('#password').fill('secret123');
    await page.locator('#remember').check();
    await page.locator('button[type="submit"]').click();
    await expect(page.locator('.welcome')).toBeVisible();
    await expect(page.locator('.welcome')).toHaveText('Welcome!');
  });
});
```

### Playwright → Selenium

**Input:**
```javascript
import { test, expect } from '@playwright/test';

test('should handle form', async ({ page }) => {
  await page.goto('/form');
  await page.locator('#email').fill('user@example.com');
  await page.locator('#country').selectOption('USA');
  await expect(page.locator('.result')).toBeVisible();
});
```

**Output:**
```javascript
const { Builder, By, Key, until } = require('selenium-webdriver');
const { expect } = require('@jest/globals');

let driver;

beforeAll(async () => {
  driver = await new Builder().forBrowser('chrome').build();
});

afterAll(async () => {
  await driver.quit();
});

describe('Test Suite', () => {
  it('should handle form', async () => {
    await driver.get('/form');
    await driver.findElement(By.css('#email')).sendKeys('user@example.com');
    const select = await driver.findElement(By.css('#country'));
    await select.findElement(By.css(`option[value=${'USA'}]`)).click();
    expect(await (await driver.findElement(By.css('.result'))).isDisplayed()).toBe(true);
  });
});
```

## Command Mapping

### Actions

| Cypress | Playwright | Selenium |
|---------|------------|----------|
| `cy.visit(url)` | `page.goto(url)` | `driver.get(url)` |
| `cy.get(sel)` | `page.locator(sel)` | `driver.findElement(By.css(sel))` |
| `.type(text)` | `.fill(text)` | `.sendKeys(text)` |
| `.click()` | `.click()` | `.click()` |
| `.clear()` | `.clear()` | `.clear()` |
| `.check()` | `.check()` | Conditional click |
| `.uncheck()` | `.uncheck()` | Conditional click |
| `.select(val)` | `.selectOption(val)` | Select option by value |
| `cy.reload()` | `page.reload()` | `driver.navigate().refresh()` |
| `cy.go('back')` | `page.goBack()` | `driver.navigate().back()` |
| `cy.wait(ms)` | `page.waitForTimeout(ms)` | `driver.sleep(ms)` |

### Assertions

| Cypress | Playwright | Selenium |
|---------|------------|----------|
| `.should('be.visible')` | `expect().toBeVisible()` | `expect(isDisplayed()).toBe(true)` |
| `.should('not.be.visible')` | `expect().toBeHidden()` | `expect(isDisplayed()).toBe(false)` |
| `.should('have.text', t)` | `expect().toHaveText(t)` | `expect(getText()).toBe(t)` |
| `.should('contain', t)` | `expect().toContainText(t)` | `expect(getText()).toContain(t)` |
| `.should('have.value', v)` | `expect().toHaveValue(v)` | `expect(getAttribute('value')).toBe(v)` |
| `.should('exist')` | `expect().toBeAttached()` | `expect(elements.length).toBeGreaterThan(0)` |
| `.should('not.exist')` | `expect().not.toBeAttached()` | `expect(elements.length).toBe(0)` |
| `.should('be.checked')` | `expect().toBeChecked()` | `expect(isSelected()).toBe(true)` |
| `.should('be.disabled')` | `expect().toBeDisabled()` | `expect(isEnabled()).toBe(false)` |
| `.should('have.length', n)` | `expect().toHaveCount(n)` | `expect(elements.length).toBe(n)` |

### Test Structure

| Cypress | Playwright | Selenium/Jest |
|---------|------------|---------------|
| `describe()` | `test.describe()` | `describe()` |
| `it()` | `test()` | `it()` |
| `beforeEach()` | `test.beforeEach()` | `beforeEach()` |
| `afterEach()` | `test.afterEach()` | `afterEach()` |
| `before()` | `test.beforeAll()` | `beforeAll()` |
| `after()` | `test.afterAll()` | `afterAll()` |

## Programmatic API

```javascript
import { ConverterFactory, FRAMEWORKS } from 'hamlet-converter';

// Create a converter
const converter = await ConverterFactory.createConverter(
  FRAMEWORKS.CYPRESS,
  FRAMEWORKS.PLAYWRIGHT
);

// Convert content
const cypressCode = `
  describe('Test', () => {
    it('works', () => {
      cy.visit('/');
      cy.get('#btn').click();
    });
  });
`;

const playwrightCode = await converter.convert(cypressCode);
console.log(playwrightCode);
```

### TypeScript Support

```typescript
import {
  ConverterFactory,
  IConverter,
  Framework,
  ConversionOptions
} from 'hamlet-converter';

const options: ConversionOptions = {
  preserveStructure: true,
  batchSize: 10
};

const converter: IConverter = await ConverterFactory.createConverter(
  'cypress' as Framework,
  'playwright' as Framework,
  options
);
```

### Available Exports

```javascript
import {
  // Factory
  ConverterFactory,
  FRAMEWORKS,

  // Individual converters
  CypressToPlaywright,
  PlaywrightToCypress,
  CypressToSelenium,
  SeleniumToCypress,
  PlaywrightToSelenium,
  SeleniumToPlaywright,

  // Core utilities
  BaseConverter,
  PatternEngine,
  FrameworkDetector
} from 'hamlet-converter';
```

## Configuration

Create a `.hamletrc.json` configuration file:

```bash
hamlet init
```

```json
{
  "defaultSource": "cypress",
  "defaultTarget": "playwright",
  "output": "./converted",
  "preserveStructure": true,
  "validate": true,
  "batchSize": 5,
  "ignore": ["node_modules/**", "**/fixtures/**"]
}
```

## Framework Detection

Auto-detect the testing framework from file content:

```bash
hamlet detect ./tests/unknown-test.js
```

Output:
```
Framework Detection Results:

  File: ./tests/unknown-test.js
  Detected Framework: cypress
  Confidence: 95%
  Detection Method: content

  Scores:
    cypress      ████████████████████ (95)
    playwright   ██ (10)
    selenium     █ (5)
```

## Project Structure

```
hamlet/
├── bin/hamlet.js           # CLI entry point
├── src/
│   ├── core/
│   │   ├── BaseConverter.js
│   │   ├── ConverterFactory.js
│   │   ├── FrameworkDetector.js
│   │   └── PatternEngine.js
│   ├── converters/
│   │   ├── CypressToPlaywright.js
│   │   ├── CypressToSelenium.js
│   │   ├── PlaywrightToCypress.js
│   │   ├── PlaywrightToSelenium.js
│   │   ├── SeleniumToCypress.js
│   │   └── SeleniumToPlaywright.js
│   └── types/
│       └── index.d.ts      # TypeScript definitions
└── test/                   # 207 tests
```

## Development

```bash
# Install dependencies
npm install

# Run tests
npm test

# Run linter
npm run lint

# Format code
npm run format
```

## Requirements

- Node.js >= 16.0.0

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Links

- [GitHub Repository](https://github.com/pmclSF/hamlet)
- [npm Package](https://www.npmjs.com/package/hamlet-converter)
- [Issue Tracker](https://github.com/pmclSF/hamlet/issues)
