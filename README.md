# Hamlet 🎭
## To be or not to be... in Playwright

A sophisticated test migration tool that converts Cypress tests to Playwright format, with support for test management systems and CI/CD integration.

## Features
- 🔄 Automated test conversion with intelligent pattern matching
- 📊 Test management system integration (TestRail, Azure DevOps, Xray)
- 🔄 CI/CD pipeline support
- 📝 Detailed reporting and error handling
- 🧪 Support for various test types (E2E, Component, API, Accessibility)

## Installation

```bash
# Install globally
npm install -g hamlet-test-converter

# Or install in your project
npm install --save-dev hamlet-test-converter
Quick Start
bashCopy# Initialize Hamlet in your project
hamlet init

# Convert a single test
hamlet convert ./cypress/e2e/test.cy.js -o ./playwright/test.spec.js

# Convert entire test suite
hamlet convert ./cypress/e2e -o ./playwright-tests --recursive

# Convert and sync with test management system
hamlet convert ./cypress/e2e -o ./playwright-tests --sync --config .hamletrc.json
Command Reference
Convert Command
bashCopyhamlet convert <source> [options]

Options:
  -o, --output <path>    Output file or directory path
  -r, --recursive        Process directories recursively
  -s, --sync            Sync with test management system
  -c, --config <path>   Path to config file
  -v, --verbose         Show verbose output
Sync Command
bashCopyhamlet sync <source> [options]

Options:
  -c, --config <path>     Path to test management config
  -t, --type <type>       Test management system type
  --dry-run              Show what would be synced
  --create-missing       Create missing test cases
Init Command
bashCopyhamlet init [options]

Options:
  -t, --type <type>      Test management system type
Configuration
Create a .hamletrc.json file:
jsonCopy{
  "testManagement": {
    "type": "testrail",
    "config": {
      "url": "https://your-instance.testrail.com",
      "username": "api-user",
      "apiKey": "your-api-key",
      "project": "Your Project",
      "suite": "Your Suite"
    }
  },
  "conversion": {
    "patterns": "./patterns.json",
    "output": "./playwright-tests"
  }
}
Supported Test Types
E2E Tests
javascriptCopy// Cypress
describe('Login', () => {
  it('should login', () => {
    cy.visit('/login');
    cy.get('[data-test=username]').type('user');
    cy.get('[data-test=submit]').click();
  });
});

// Converted Playwright
test.describe('Login', () => {
  test('should login', async ({ page }) => {
    await page.goto('/login');
    await page.locator('[data-test=username]').fill('user');
    await page.locator('[data-test=submit]').click();
  });
});
API Tests
javascriptCopy// Cypress
cy.request('/api/users').its('status').should('eq', 200);

// Converted Playwright
const response = await request.get('/api/users');
expect(response.status()).toBe(200);
Accessibility Tests
javascriptCopy// Cypress
cy.injectAxe();
cy.checkA11y();

// Converted Playwright
await injectAxe(page);
await checkA11y(page);
Test Management Integration
TestRail
Add TestRail configuration to your .hamletrc.json:
jsonCopy{
  "testManagement": {
    "type": "testrail",
    "config": {
      "url": "https://your-instance.testrail.com",
      "username": "api-user",
      "apiKey": "your-api-key"
    }
  }
}
Azure Test Plans
jsonCopy{
  "testManagement": {
    "type": "azure",
    "config": {
      "organization": "your-org",
      "project": "your-project",
      "pat": "your-personal-access-token"
    }
  }
}
Contributing
We welcome contributions! Please see our Contributing Guide for details.
Development Setup
bashCopy# Clone the repository
git clone https://github.com/pmcISF/hamlet.git
cd hamlet

# Install dependencies
npm install

# Link for local development
npm link

# Run tests
npm test
License
MIT License - see LICENSE file for details.
Support

📚 Documentation
🐛 Issue Tracker
💬 Discussions# test
# another test
