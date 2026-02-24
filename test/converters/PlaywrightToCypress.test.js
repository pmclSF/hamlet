import { PlaywrightToCypress } from '../../src/converters/PlaywrightToCypress.js';

describe('PlaywrightToCypress', () => {
  let converter;

  beforeEach(() => {
    converter = new PlaywrightToCypress();
  });

  describe('constructor', () => {
    it('should set source and target frameworks', () => {
      expect(converter.sourceFramework).toBe('playwright');
      expect(converter.targetFramework).toBe('cypress');
    });
  });

  describe('convert', () => {
    describe('test structure', () => {
      it('should convert test.describe to describe', async () => {
        const input = `test.describe('Test Suite', () => {});`;
        const result = await converter.convert(input);
        expect(result).toContain("describe('Test Suite'");
      });

      it('should convert test to it', async () => {
        const input = `test('should work', async ({ page }) => {});`;
        const result = await converter.convert(input);
        expect(result).toContain("it('should work'");
      });

      it('should remove async and page parameter', async () => {
        const input = `test('test', async ({ page }) => { await page.goto('/'); });`;
        const result = await converter.convert(input);
        expect(result).not.toContain('async ({ page })');
      });
    });

    describe('navigation', () => {
      it('should convert page.goto to cy.visit', async () => {
        const input = `await page.goto('/home');`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.visit('/home')");
      });

      it('should convert page.reload to cy.reload', async () => {
        const input = `await page.reload();`;
        const result = await converter.convert(input);
        expect(result).toContain('cy.reload()');
      });

      it('should convert page.goBack to cy.go("back")', async () => {
        const input = `await page.goBack();`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.go('back')");
      });
    });

    describe('selectors and interactions', () => {
      it('should convert page.locator().click() to cy.get().click()', async () => {
        const input = `await page.locator('.btn').click();`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.get('.btn').click()");
      });

      it('should convert page.locator().fill() to cy.get().type()', async () => {
        const input = `await page.locator('#input').fill('hello');`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.get('#input').type('hello')");
      });

      it('should convert page.getByText().click() to cy.contains().click()', async () => {
        const input = `await page.getByText('Click me').click();`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.contains('Click me').click()");
      });
    });

    describe('assertions', () => {
      it('should convert toBeVisible() to should("be.visible")', async () => {
        const input = `await expect(page.locator('.element')).toBeVisible();`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.get('.element').should('be.visible')");
      });

      it('should convert toBeHidden() to should("not.be.visible")', async () => {
        const input = `await expect(page.locator('.element')).toBeHidden();`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.get('.element').should('not.be.visible')");
      });

      it('should convert toHaveText() to should("have.text")', async () => {
        const input = `await expect(page.locator('.element')).toHaveText('Hello');`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.get('.element').should('have.text', 'Hello')");
      });

      it('should convert toContainText() to should("contain.text")', async () => {
        const input = `await expect(page.locator('.element')).toContainText('Hello');`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.get('.element').should('contain.text', 'Hello')");
      });

      it('should convert toHaveValue() to should("have.value")', async () => {
        const input = `await expect(page.locator('#input')).toHaveValue('test');`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.get('#input').should('have.value', 'test')");
      });

      it('should convert toBeAttached() to should("exist")', async () => {
        const input = `await expect(page.locator('.element')).toBeAttached();`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.get('.element').should('exist')");
      });

      it('should convert toBeChecked() to should("be.checked")', async () => {
        const input = `await expect(page.locator('#checkbox')).toBeChecked();`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.get('#checkbox').should('be.checked')");
      });

      it('should convert toBeDisabled() to should("be.disabled")', async () => {
        const input = `await expect(page.locator('#btn')).toBeDisabled();`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.get('#btn').should('be.disabled')");
      });

      it('should convert toHaveCount() to should("have.length")', async () => {
        const input = `await expect(page.locator('.items')).toHaveCount(3);`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.get('.items').should('have.length', 3)");
      });
    });

    describe('imports', () => {
      it('should remove Playwright imports', async () => {
        const input = `import { test, expect } from '@playwright/test';\ntest('test', () => {});`;
        const result = await converter.convert(input);
        expect(result).not.toContain("import { test, expect } from '@playwright/test'");
      });

      it('should add Cypress reference', async () => {
        const input = `test('test', () => {});`;
        const result = await converter.convert(input);
        expect(result).toContain('/// <reference types="cypress" />');
      });
    });
  });
});
