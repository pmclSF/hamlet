# Hamlet 🎭

> To be, or not to be... in Playwright. A test converter from Cypress to Playwright.

Hamlet is a CLI tool that automates the conversion of Cypress tests to Playwright, supporting a wide range of test types and patterns.

## Features

### Core Modules

#### Conversion Orchestrator
Coordinates the entire conversion process with comprehensive validation and reporting:
```javascript
const orchestrator = new ConversionOrchestrator({
  validateTests: true,
  compareVisuals: true,
  generateTypes: true
});

await orchestrator.convertProject('./cypress', './playwright');
```

#### Test Validator
Ensures converted tests maintain functionality and follow best practices:
```javascript
const validator = new TestValidator();
const results = await validator.validateConvertedTests('./tests');
```

#### Visual Comparison
Compares visual test outputs between frameworks:
```javascript
const visualComparison = new VisualComparison({
  threshold: 0.1,
  saveSnapshots: true
});

await visualComparison.compareProjects('./cypress', './playwright');
```

#### TypeScript Support
Handles TypeScript conversion with type preservation:
```javascript
const tsConverter = new TypeScriptConverter();
await tsConverter.convertProject('./cypress', './playwright');
```

#### Plugin Converter
Converts Cypress plugins to Playwright equivalents:
```javascript
const pluginConverter = new PluginConverter();
const result = await pluginConverter.convertPlugin('cypress-file-upload');
```

#### Test Mapper
Maintains bidirectional mapping between tests:
```javascript
const mapper = new TestMapper();
await mapper.setupMappings('./cypress/tests', './playwright/tests');
```

### Supported Test Types

#### E2E Tests
Convert end-to-end test scenarios with full page navigation and user flows.
```typescript
// Cypress
cy.visit('/checkout')
cy.get('[data-testid="cart-item"]').should('have.length', 2)

// Converted Playwright
await page.goto('/checkout')
await expect(page.locator('[data-testid="cart-item"]')).toHaveCount(2)
```

[Previous test type examples remain the same...]

### Additional Features

#### Comprehensive Reporting
Generate detailed conversion reports in multiple formats:
```javascript
const reporter = new ConversionReporter({
  format: 'html',
  includeTimestamps: true
});

await reporter.generateReport();
```

#### Utility Functions
Helper functions for common conversion tasks:
```javascript
// File operations
await fileUtils.ensureDir('./output');
const files = await fileUtils.getFiles('./tests', /\.spec\.js$/);

// String manipulation
const kebabCase = stringUtils.camelToKebab('myTestCase');

// Code analysis
const imports = codeUtils.extractImports(sourceCode);

// Test analysis
const testCases = testUtils.extractTestCases(testContent);
```

## Installation

```bash
npm install -g hamlet-test-converter
```

## Usage

Basic conversion:
```bash
hamlet convert ./cypress/e2e/**/*.cy.{js,ts}
```

Advanced options:
```bash
hamlet convert ./cypress/e2e --validate --visual-compare --generate-types
```

## Configuration

Create a `hamlet.config.js` file:

```javascript
module.exports = {
  outDir: './tests',
  preserveComments: true,
  validation: {
    enabled: true,
    threshold: 0.1
  },
  visual: {
    enabled: true,
    saveSnapshots: true
  },
  typescript: {
    enabled: true,
    strict: true
  },
  reporting: {
    format: 'html',
    includeTimestamps: true
  }
}
```

## Project Structure

```
src/
├── converter/
│   ├── orchestrator.js   (Main conversion coordinator)
│   ├── validator.js      (Test validation)
│   ├── typescript.js     (TypeScript handling)
│   ├── mapper.js         (Test mapping)
│   ├── plugins.js        (Plugin conversion)
│   └── visual.js         (Visual comparison)
├── utils/
│   ├── helpers.js        (Utility functions)
│   └── reporter.js       (Report generation)
└── index.js              (Main entry point)
```

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.