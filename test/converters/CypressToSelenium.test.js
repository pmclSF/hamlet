import { CypressToSelenium } from '../../src/converters/CypressToSelenium.js';

describe('CypressToSelenium', () => {
  let converter;

  beforeEach(() => {
    converter = new CypressToSelenium();
  });

  describe('constructor', () => {
    it('should set source and target frameworks', () => {
      expect(converter.sourceFramework).toBe('cypress');
      expect(converter.targetFramework).toBe('selenium');
    });
  });

  describe('convert', () => {
    describe('test structure', () => {
      it('should make test callbacks async', async () => {
        const input = `it('should work', () => { cy.visit('/'); });`;
        const result = await converter.convert(input);
        expect(result).toContain('async');
      });
    });

    describe('navigation', () => {
      it('should convert cy.visit to driver.get', async () => {
        const input = `cy.visit('/home');`;
        const result = await converter.convert(input);
        expect(result).toContain("await driver.get('/home')");
      });

      it('should convert cy.reload to driver.navigate().refresh', async () => {
        const input = `cy.reload();`;
        const result = await converter.convert(input);
        expect(result).toContain('await driver.navigate().refresh()');
      });

      it('should convert cy.go("back") to driver.navigate().back', async () => {
        const input = `cy.go('back');`;
        const result = await converter.convert(input);
        expect(result).toContain('await driver.navigate().back()');
      });
    });

    describe('selectors and interactions', () => {
      it('should convert cy.get().click() to driver.findElement().click()', async () => {
        const input = `cy.get('.btn').click();`;
        const result = await converter.convert(input);
        expect(result).toContain("await driver.findElement(By.css('.btn')).click()");
      });

      it('should convert cy.get().type() to driver.findElement().sendKeys()', async () => {
        const input = `cy.get('#input').type('hello');`;
        const result = await converter.convert(input);
        expect(result).toContain("await driver.findElement(By.css('#input')).sendKeys('hello')");
      });

      it('should convert cy.get().clear() to driver.findElement().clear()', async () => {
        const input = `cy.get('#input').clear();`;
        const result = await converter.convert(input);
        expect(result).toContain("await driver.findElement(By.css('#input')).clear()");
      });
    });

    describe('assertions', () => {
      it('should convert should("be.visible") to isDisplayed assertion', async () => {
        const input = `cy.get('.element').should('be.visible');`;
        const result = await converter.convert(input);
        expect(result).toContain('isDisplayed()');
        expect(result).toContain('toBe(true)');
      });

      it('should convert should("have.text") to getText assertion', async () => {
        const input = `cy.get('.element').should('have.text', 'Hello');`;
        const result = await converter.convert(input);
        expect(result).toContain('getText()');
        expect(result).toContain("toBe('Hello')");
      });

      it('should convert should("have.length") to findElements length assertion', async () => {
        const input = `cy.get('.items').should('have.length', 3);`;
        const result = await converter.convert(input);
        expect(result).toContain('findElements');
        expect(result).toContain('length');
        expect(result).toContain('toBe(3)');
      });
    });

    describe('imports and setup', () => {
      it('should add Selenium imports', async () => {
        const input = `it('test', () => {});`;
        const result = await converter.convert(input);
        expect(result).toContain("const { Builder, By, Key, until } = require('selenium-webdriver')");
      });

      it('should add Jest expect import', async () => {
        const input = `it('test', () => {});`;
        const result = await converter.convert(input);
        expect(result).toContain("const { expect } = require('@jest/globals')");
      });

      it('should add driver setup', async () => {
        const input = `it('test', () => {});`;
        const result = await converter.convert(input);
        expect(result).toContain('let driver');
        expect(result).toContain('beforeAll');
        expect(result).toContain('new Builder()');
      });

      it('should add driver teardown', async () => {
        const input = `it('test', () => {});`;
        const result = await converter.convert(input);
        expect(result).toContain('afterAll');
        expect(result).toContain('driver.quit()');
      });
    });
  });
});
