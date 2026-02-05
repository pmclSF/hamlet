import { CypressToPlaywright } from '../../src/converters/CypressToPlaywright.js';

describe('CypressToPlaywright', () => {
  let converter;

  beforeEach(() => {
    converter = new CypressToPlaywright();
  });

  describe('constructor', () => {
    it('should set source and target frameworks', () => {
      expect(converter.sourceFramework).toBe('cypress');
      expect(converter.targetFramework).toBe('playwright');
    });
  });

  describe('convert', () => {
    describe('test structure', () => {
      it('should convert describe to test.describe', async () => {
        const input = `describe('Test Suite', () => {});`;
        const result = await converter.convert(input);
        expect(result).toContain("test.describe('Test Suite'");
      });

      it('should convert it to test', async () => {
        const input = `it('should work', () => {});`;
        const result = await converter.convert(input);
        expect(result).toContain("test('should work'");
      });

      it('should convert beforeEach to test.beforeEach', async () => {
        const input = `beforeEach(() => {});`;
        const result = await converter.convert(input);
        expect(result).toContain('test.beforeEach(');
      });

      it('should convert afterEach to test.afterEach', async () => {
        const input = `afterEach(() => {});`;
        const result = await converter.convert(input);
        expect(result).toContain('test.afterEach(');
      });

      it('should add async and page parameter to test callbacks', async () => {
        const input = `it('test', () => { cy.visit('/'); });`;
        const result = await converter.convert(input);
        expect(result).toContain('async ({ page })');
      });
    });

    describe('navigation', () => {
      it('should convert cy.visit to page.goto', async () => {
        const input = `cy.visit('/home');`;
        const result = await converter.convert(input);
        expect(result).toContain("await page.goto('/home')");
      });

      it('should convert cy.reload to page.reload', async () => {
        const input = `cy.reload();`;
        const result = await converter.convert(input);
        expect(result).toContain('await page.reload()');
      });

      it('should convert cy.go("back") to page.goBack', async () => {
        const input = `cy.go('back');`;
        const result = await converter.convert(input);
        expect(result).toContain('await page.goBack()');
      });
    });

    describe('selectors and interactions', () => {
      it('should convert cy.get().click() to page.locator().click()', async () => {
        const input = `cy.get('.btn').click();`;
        const result = await converter.convert(input);
        expect(result).toContain("await page.locator('.btn').click()");
      });

      it('should convert cy.get().type() to page.locator().fill()', async () => {
        const input = `cy.get('#input').type('hello');`;
        const result = await converter.convert(input);
        expect(result).toContain("await page.locator('#input').fill('hello')");
      });

      it('should convert cy.get().clear() to page.locator().clear()', async () => {
        const input = `cy.get('#input').clear();`;
        const result = await converter.convert(input);
        expect(result).toContain("await page.locator('#input').clear()");
      });

      it('should convert cy.get().check() to page.locator().check()', async () => {
        const input = `cy.get('#checkbox').check();`;
        const result = await converter.convert(input);
        expect(result).toContain("await page.locator('#checkbox').check()");
      });

      it('should convert cy.get().select() to page.locator().selectOption()', async () => {
        const input = `cy.get('select').select('option1');`;
        const result = await converter.convert(input);
        expect(result).toContain("await page.locator('select').selectOption('option1')");
      });

      it('should convert cy.contains() to page.getByText()', async () => {
        const input = `cy.contains('Click me').click();`;
        const result = await converter.convert(input);
        expect(result).toContain("await page.getByText('Click me').click()");
      });
    });

    describe('assertions', () => {
      it('should convert should("be.visible") to toBeVisible()', async () => {
        const input = `cy.get('.element').should('be.visible');`;
        const result = await converter.convert(input);
        expect(result).toContain("await expect(page.locator('.element')).toBeVisible()");
      });

      it('should convert should("not.be.visible") to toBeHidden()', async () => {
        const input = `cy.get('.element').should('not.be.visible');`;
        const result = await converter.convert(input);
        expect(result).toContain("await expect(page.locator('.element')).toBeHidden()");
      });

      it('should convert should("have.text") to toHaveText()', async () => {
        const input = `cy.get('.element').should('have.text', 'Hello');`;
        const result = await converter.convert(input);
        expect(result).toContain("await expect(page.locator('.element')).toHaveText('Hello')");
      });

      it('should convert should("contain") to toContainText()', async () => {
        const input = `cy.get('.element').should('contain', 'Hello');`;
        const result = await converter.convert(input);
        expect(result).toContain("await expect(page.locator('.element')).toContainText('Hello')");
      });

      it('should convert should("have.value") to toHaveValue()', async () => {
        const input = `cy.get('#input').should('have.value', 'test');`;
        const result = await converter.convert(input);
        expect(result).toContain("await expect(page.locator('#input')).toHaveValue('test')");
      });

      it('should convert should("exist") to toBeAttached()', async () => {
        const input = `cy.get('.element').should('exist');`;
        const result = await converter.convert(input);
        expect(result).toContain("await expect(page.locator('.element')).toBeAttached()");
      });

      it('should convert should("be.checked") to toBeChecked()', async () => {
        const input = `cy.get('#checkbox').should('be.checked');`;
        const result = await converter.convert(input);
        expect(result).toContain("await expect(page.locator('#checkbox')).toBeChecked()");
      });

      it('should convert should("be.disabled") to toBeDisabled()', async () => {
        const input = `cy.get('#btn').should('be.disabled');`;
        const result = await converter.convert(input);
        expect(result).toContain("await expect(page.locator('#btn')).toBeDisabled()");
      });

      it('should convert should("have.length") to toHaveCount()', async () => {
        const input = `cy.get('.items').should('have.length', 3);`;
        const result = await converter.convert(input);
        expect(result).toContain("await expect(page.locator('.items')).toHaveCount(3)");
      });
    });

    describe('imports', () => {
      it('should add Playwright import', async () => {
        const input = `it('test', () => {});`;
        const result = await converter.convert(input);
        expect(result).toContain("import { test, expect } from '@playwright/test'");
      });
    });
  });

  describe('detectTestTypes', () => {
    it('should detect API tests', () => {
      const types = converter.detectTestTypes('cy.request("/api/users")');
      expect(types).toContain('api');
    });

    it('should detect component tests', () => {
      const types = converter.detectTestTypes('cy.mount(<Component />)');
      expect(types).toContain('component');
    });

    it('should default to e2e', () => {
      const types = converter.detectTestTypes('cy.visit("/")');
      expect(types).toContain('e2e');
    });
  });
});
