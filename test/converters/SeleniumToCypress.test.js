import { SeleniumToCypress } from '../../src/converters/SeleniumToCypress.js';

describe('SeleniumToCypress', () => {
  let converter;

  beforeEach(() => {
    converter = new SeleniumToCypress();
  });

  describe('constructor', () => {
    it('should set source and target frameworks', () => {
      expect(converter.sourceFramework).toBe('selenium');
      expect(converter.targetFramework).toBe('cypress');
    });
  });

  describe('convert', () => {
    describe('navigation', () => {
      it('should convert driver.get to cy.visit', async () => {
        const input = `await driver.get('/home');`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.visit('/home')");
      });

      it('should convert driver.navigate().refresh to cy.reload', async () => {
        const input = `await driver.navigate().refresh();`;
        const result = await converter.convert(input);
        expect(result).toContain('cy.reload()');
      });

      it('should convert driver.navigate().back to cy.go("back")', async () => {
        const input = `await driver.navigate().back();`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.go('back')");
      });
    });

    describe('selectors and interactions', () => {
      it('should convert driver.findElement().click to cy.get().click()', async () => {
        const input = `await driver.findElement(By.css('.btn')).click();`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.get('.btn').click()");
      });

      it('should convert driver.findElement().sendKeys to cy.get().type()', async () => {
        const input = `await driver.findElement(By.css('#input')).sendKeys('hello');`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.get('#input').type('hello')");
      });
    });

    describe('assertions', () => {
      it('should convert isDisplayed assertion to should("be.visible")', async () => {
        const input = `expect(await (await driver.findElement(By.css('.element'))).isDisplayed()).toBe(true);`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.get('.element').should('be.visible')");
      });

      it('should convert getText assertion to should("have.text")', async () => {
        const input = `expect(await (await driver.findElement(By.css('.element'))).getText()).toBe('Hello');`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.get('.element').should('have.text', 'Hello')");
      });

      it('should convert findElements length to should("have.length")', async () => {
        const input = `expect((await driver.findElements(By.css('.items'))).length).toBe(3);`;
        const result = await converter.convert(input);
        expect(result).toContain("cy.get('.items').should('have.length', 3)");
      });
    });

    describe('boilerplate removal', () => {
      it('should remove Selenium imports', async () => {
        const input = `const { Builder, By } = require('selenium-webdriver');\nit('test', () => {});`;
        const result = await converter.convert(input);
        expect(result).not.toContain("require('selenium-webdriver')");
      });

      it('should remove driver setup', async () => {
        const input = `let driver;\nbeforeAll(async () => { driver = await new Builder().forBrowser('chrome').build(); });`;
        const result = await converter.convert(input);
        expect(result).not.toContain('new Builder()');
      });

      it('should add Cypress reference', async () => {
        const input = `it('test', () => {});`;
        const result = await converter.convert(input);
        expect(result).toContain('/// <reference types="cypress" />');
      });
    });
  });
});
