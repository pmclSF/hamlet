# Hamlet 🎭

> To be, or not to be... in Playwright. A test converter from Cypress to Playwright.

Hamlet is a CLI tool that automates the conversion of Cypress tests to Playwright, supporting a wide range of test types and patterns.

## Features

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

#### API Tests
Transform API testing patterns including request interception and validation.
```typescript
// Cypress
cy.request('POST', '/api/users', { name: 'John' })
  .its('status')
  .should('equal', 200)

// Converted Playwright
const response = await request.post('/api/users', {
  data: { name: 'John' }
})
expect(response.status()).toBe(200)
```

#### Component Tests
Convert component-level tests with mounting support.
```typescript
// Cypress
cy.mount(Button, { props: { label: 'Click me' } })
cy.get('button').click()

// Converted Playwright
await mount(Button, { props: { label: 'Click me' } })
await page.getByRole('button').click()
```

#### Accessibility Tests
Transform accessibility testing patterns using axe-core.
```typescript
// Cypress
cy.injectAxe()
cy.checkA11y()

// Converted Playwright
await injectAxe()
await checkA11y()
```

#### Visual Tests
Convert visual regression testing scenarios.
```typescript
// Cypress
cy.get('.header').matchImageSnapshot()

// Converted Playwright
await expect(page.locator('.header')).toHaveScreenshot()
```

#### Performance Tests
Transform performance measurement tests.
```typescript
// Cypress
cy.window().its('performance').then((p) => {
  const navigation = p.getEntriesByType('navigation')[0]
  expect(navigation.domComplete).to.be.lessThan(2000)
})

// Converted Playwright
const timing = await page.evaluate(() => performance.getEntriesByType('navigation')[0])
expect(timing.domComplete).toBeLessThan(2000)
```

#### Mobile/Responsive Tests
Convert viewport-specific and responsive design tests.
```typescript
// Cypress
cy.viewport('iphone-x')
cy.get('.mobile-menu').should('be.visible')

// Converted Playwright
await page.setViewportSize({ width: 375, height: 812 })
await expect(page.locator('.mobile-menu')).toBeVisible()
```

#### Network Tests
Transform network interception and mocking patterns.
```typescript
// Cypress
cy.intercept('GET', '/api/users', { fixture: 'users.json' })

// Converted Playwright
await page.route('/api/users', async route => {
  await route.fulfill({ path: 'fixtures/users.json' })
})
```

#### Authentication Tests
Convert authentication flows and session handling.
```typescript
// Cypress
cy.login('user@example.com', 'password')
cy.getCookie('session').should('exist')

// Converted Playwright
await login('user@example.com', 'password')
await expect(await context.cookies()).toContainEqual(
  expect.objectContaining({ name: 'session' })
)
```

#### Database Tests
Transform database seeding and verification patterns.
```typescript
// Cypress
cy.task('db:seed', { users: 1 })
cy.task('db:query', 'SELECT * FROM users')

// Converted Playwright
await test.step('db operations', async () => {
  await seedDatabase({ users: 1 })
  const results = await queryDatabase('SELECT * FROM users')
})
```

#### File Upload/Download Tests
Convert file handling test scenarios.
```typescript
// Cypress
cy.get('input[type="file"]').attachFile('example.pdf')

// Converted Playwright
await page.locator('input[type="file"]').setInputFiles('example.pdf')
```

#### iFrame Tests
Transform iframe interaction patterns.
```typescript
// Cypress
cy.get('#frame').iframe().find('.content')

// Converted Playwright
const frame = page.frameLocator('#frame')
await frame.locator('.content')
```

### Additional Features

- Custom Commands Conversion
- Configuration File Migration
- Test Pattern Detection
- Error Recovery & Reporting

## Installation

```bash
npm install -g hamlet-test-converter
```

## Usage

```bash
hamlet convert ./cypress/e2e/**/*.cy.{js,ts}
```

## Configuration

Create a `hamlet.config.js` file:

```javascript
module.exports = {
  outDir: './tests',
  preserveComments: true,
  // Additional configuration options
}
```

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.