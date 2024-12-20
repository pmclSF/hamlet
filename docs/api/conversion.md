# Conversion Process

## Overview

Hamlet converts Cypress tests to Playwright using a multi-step process:

1. Test Analysis
   - Detect test type
   - Analyze dependencies
   - Extract metadata

2. Conversion
   - Transform commands
   - Convert assertions
   - Update imports
   - Handle async/await

3. Validation
   - Syntax check
   - Run converted tests
   - Compare outputs

## Supported Conversions

### Commands
| Cypress | Playwright |
|---------|------------|
| `cy.visit()` | `page.goto()` |
| `cy.get()` | `page.locator()` |
| `cy.click()` | `click()` |
| `cy.type()` | `fill()` |
| `cy.request()` | `request.fetch()` |

### Assertions
| Cypress | Playwright |
|---------|------------|
| `should('exist')` | `toBeVisible()` |
| `should('have.text')` | `toHaveText()` |
| `should('have.value')` | `toHaveValue()` |
| `should('be.checked')` | `toBeChecked()` |

### Test Structure
```javascript
// Cypress
describe('My Test', () => {
  it('should work', () => {
    cy.visit('/page');
    cy.get('button').click();
  });
});

// Playwright
test.describe('My Test', () => {
  test('should work', async ({ page }) => {
    await page.goto('/page');
    await page.locator('button').click();
  });
});
```

## Plugin Conversion

Hamlet handles common Cypress plugins:

- cypress-file-upload → Playwright built-in
- cypress-xpath → Playwright built-in
- cypress-real-events → Playwright built-in
- cypress-image-snapshot → Playwright screenshots

## Advanced Features

### Custom Command Conversion
```javascript
// Custom command definition
Cypress.Commands.add('login', (username, password) => {
  cy.visit('/login');
  cy.get('#username').type(username);
  cy.get('#password').type(password);
  cy.get('button').click();
});

// Converted to Playwright helper
async function login(page, username, password) {
  await page.goto('/login');
  await page.locator('#username').fill(username);
  await page.locator('#password').fill(password);
  await page.locator('button').click();
}
```

### Visual Testing
```javascript
// Cypress
cy.get('.header').matchImageSnapshot();

// Playwright
await expect(page.locator('.header')).toHaveScreenshot();
```

### API Testing
```javascript
// Cypress
cy.request('POST', '/api/users', { name: 'John' })
  .its('status')
  .should('equal', 200);

// Playwright
const response = await request.post('/api/users', {
  data: { name: 'John' }
});
expect(response.status()).toBe(200);
```