# Migration Guide

## Overview

This guide helps you migrate your Cypress tests to Playwright using Hamlet.

## Migration Steps

1. **Prepare Your Project**
   - Backup your tests
   - Install Hamlet
   - Create configuration

2. **Analyze Your Tests**
   ```bash
   hamlet analyze cypress/e2e
   ```

3. **Convert Tests**
   ```bash
   hamlet convert cypress/e2e
   ```

4. **Validate Conversion**
   ```bash
   hamlet validate playwright/tests
   ```

## Common Patterns

### Async/Await
Playwright requires explicit async/await:

```javascript
// Cypress
it('test', () => {
  cy.visit('/page');
  cy.click('button');
});

// Playwright
test('test', async ({ page }) => {
  await page.goto('/page');
  await page.click('button');
});
```

### Selectors
Playwright uses different selector strategies:

```javascript
// Cypress
cy.get('[data-cy=submit]')
cy.contains('Click me')

// Playwright
await page.locator('[data-testid=submit]')
await page.getByText('Click me')
```

### Custom Commands
Convert custom commands to helper functions:

```javascript
// Cypress
Cypress.Commands.add('login', (user) => {
  cy.visit('/login');
  cy.get('#user').type(user);
});

// Playwright
async function login(page, user) {
  await page.goto('/login');
  await page.locator('#user').fill(user);
}
```

## Test Management Integration

1. **Azure DevOps**
   ```javascript
   // hamlet.config.js
   module.exports = {
     testManagement: {
       type: 'azure',
       config: {
         organization: 'your-org',
         project: 'your-project'
       }
     }
   }
   ```

2. **TestRail**
   ```javascript
   module.exports = {
     testManagement: {
       type: 'testrail',
       config: {
         host: 'your-instance.testrail.com'
       }
     }
   }
   ```

## Best Practices

1. Convert tests in small batches
2. Validate each conversion
3. Update your CI/CD pipeline
4. Keep both versions until fully migrated

## Troubleshooting

Common issues and solutions:

1. **Selector Issues**
   ```javascript
   // Problem
   cy.get('.dynamic-class')

   // Solution
   await page.locator('[data-testid=element]')
   ```

2. **Async Operations**
   ```javascript
   // Problem
   cy.wait(2000)

   // Solution
   await page.waitForSelector('.element')
   ```

3. **Plugin Dependencies**
   ```javascript
   // Problem
   cy.iframe()

   // Solution
   const frame = page.frameLocator('iframe')
   ```

## Next Steps

1. Review [test types](./test-types.md)
2. Explore [advanced features](../advanced/repository-conversion.md)
3. Check [example conversions](../examples/)