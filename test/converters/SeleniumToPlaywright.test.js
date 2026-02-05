import { SeleniumToPlaywright } from '../../src/converters/SeleniumToPlaywright.js';

describe('SeleniumToPlaywright', () => {
  let converter;

  beforeEach(() => {
    converter = new SeleniumToPlaywright();
  });

  describe('constructor', () => {
    it('should set source and target frameworks', () => {
      expect(converter.sourceFramework).toBe('selenium');
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
        const input = `it('should work', async () => {});`;
        const result = await converter.convert(input);
        expect(result).toContain("test('should work'");
      });

      it('should add page parameter to test callbacks', async () => {
        const input = `it('test', async () => { await driver.get('/'); });`;
        const result = await converter.convert(input);
        expect(result).toContain('async ({ page })');
      });
    });

    describe('navigation', () => {
      it('should convert driver.get to page.goto', async () => {
        const input = `await driver.get('/home');`;
        const result = await converter.convert(input);
        expect(result).toContain("await page.goto('/home')");
      });

      it('should convert driver.navigate().refresh to page.reload', async () => {
        const input = `await driver.navigate().refresh();`;
        const result = await converter.convert(input);
        expect(result).toContain('await page.reload()');
      });

      it('should convert driver.navigate().back to page.goBack', async () => {
        const input = `await driver.navigate().back();`;
        const result = await converter.convert(input);
        expect(result).toContain('await page.goBack()');
      });
    });

    describe('selectors and interactions', () => {
      it('should convert driver.findElement().click to page.locator().click()', async () => {
        const input = `await driver.findElement(By.css('.btn')).click();`;
        const result = await converter.convert(input);
        expect(result).toContain("await page.locator('.btn').click()");
      });

      it('should convert driver.findElement().sendKeys to page.locator().fill()', async () => {
        const input = `await driver.findElement(By.css('#input')).sendKeys('hello');`;
        const result = await converter.convert(input);
        expect(result).toContain("await page.locator('#input').fill('hello')");
      });
    });

    describe('assertions', () => {
      it('should convert isDisplayed assertion to toBeVisible', async () => {
        const input = `expect(await (await driver.findElement(By.css('.element'))).isDisplayed()).toBe(true);`;
        const result = await converter.convert(input);
        expect(result).toContain("await expect(page.locator('.element')).toBeVisible()");
      });

      it('should convert getText assertion to toHaveText', async () => {
        const input = `expect(await (await driver.findElement(By.css('.element'))).getText()).toBe('Hello');`;
        const result = await converter.convert(input);
        expect(result).toContain("await expect(page.locator('.element')).toHaveText('Hello')");
      });

      it('should convert findElements length to toHaveCount', async () => {
        const input = `expect((await driver.findElements(By.css('.items'))).length).toBe(3);`;
        const result = await converter.convert(input);
        expect(result).toContain("await expect(page.locator('.items')).toHaveCount(3)");
      });
    });

    describe('boilerplate handling', () => {
      it('should remove Selenium imports', async () => {
        const input = `const { Builder, By } = require('selenium-webdriver');\nit('test', async () => {});`;
        const result = await converter.convert(input);
        expect(result).not.toContain("require('selenium-webdriver')");
      });

      it('should remove driver setup', async () => {
        const input = `let driver;\nbeforeAll(async () => { driver = await new Builder().forBrowser('chrome').build(); });`;
        const result = await converter.convert(input);
        expect(result).not.toContain('new Builder()');
      });

      it('should add Playwright imports', async () => {
        const input = `it('test', async () => {});`;
        const result = await converter.convert(input);
        expect(result).toContain("import { test, expect } from '@playwright/test'");
      });
    });
  });
});
