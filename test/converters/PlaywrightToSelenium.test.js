import { PlaywrightToSelenium } from '../../src/converters/PlaywrightToSelenium.js';

describe('PlaywrightToSelenium', () => {
  let converter;

  beforeEach(() => {
    converter = new PlaywrightToSelenium();
  });

  describe('constructor', () => {
    it('should set source and target frameworks', () => {
      expect(converter.sourceFramework).toBe('playwright');
      expect(converter.targetFramework).toBe('selenium');
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

      it('should remove page parameter from test callbacks', async () => {
        const input = `test('test', async ({ page }) => { await page.goto('/'); });`;
        const result = await converter.convert(input);
        expect(result).not.toContain('({ page })');
      });
    });

    describe('navigation', () => {
      it('should convert page.goto to driver.get', async () => {
        const input = `await page.goto('/home');`;
        const result = await converter.convert(input);
        expect(result).toContain("await driver.get('/home')");
      });

      it('should convert page.reload to driver.navigate().refresh', async () => {
        const input = `await page.reload();`;
        const result = await converter.convert(input);
        expect(result).toContain('await driver.navigate().refresh()');
      });

      it('should convert page.goBack to driver.navigate().back', async () => {
        const input = `await page.goBack();`;
        const result = await converter.convert(input);
        expect(result).toContain('await driver.navigate().back()');
      });
    });

    describe('selectors and interactions', () => {
      it('should convert page.locator().click to driver.findElement().click()', async () => {
        const input = `await page.locator('.btn').click();`;
        const result = await converter.convert(input);
        expect(result).toContain("await driver.findElement(By.css('.btn')).click()");
      });

      it('should convert page.locator().fill to driver.findElement().sendKeys()', async () => {
        const input = `await page.locator('#input').fill('hello');`;
        const result = await converter.convert(input);
        expect(result).toContain("await driver.findElement(By.css('#input')).sendKeys('hello')");
      });
    });

    describe('assertions', () => {
      it('should convert toBeVisible to isDisplayed assertion', async () => {
        const input = `await expect(page.locator('.element')).toBeVisible();`;
        const result = await converter.convert(input);
        expect(result).toContain('isDisplayed()');
        expect(result).toContain('toBe(true)');
      });

      it('should convert toHaveText to getText assertion', async () => {
        const input = `await expect(page.locator('.element')).toHaveText('Hello');`;
        const result = await converter.convert(input);
        expect(result).toContain('getText()');
        expect(result).toContain("toBe('Hello')");
      });

      it('should convert toHaveCount to findElements length assertion', async () => {
        const input = `await expect(page.locator('.items')).toHaveCount(3);`;
        const result = await converter.convert(input);
        expect(result).toContain('findElements');
        expect(result).toContain('length');
        expect(result).toContain('toBe(3)');
      });
    });

    describe('imports and setup', () => {
      it('should remove Playwright imports', async () => {
        const input = `import { test, expect } from '@playwright/test';\ntest('test', async () => {});`;
        const result = await converter.convert(input);
        expect(result).not.toContain("import { test, expect } from '@playwright/test'");
      });

      it('should add Selenium imports', async () => {
        const input = `test('test', async () => {});`;
        const result = await converter.convert(input);
        expect(result).toContain("const { Builder, By, Key, until } = require('selenium-webdriver')");
      });

      it('should add driver setup', async () => {
        const input = `test('test', async () => {});`;
        const result = await converter.convert(input);
        expect(result).toContain('let driver');
        expect(result).toContain('beforeAll');
        expect(result).toContain('new Builder()');
      });

      it('should add driver teardown', async () => {
        const input = `test('test', async () => {});`;
        const result = await converter.convert(input);
        expect(result).toContain('afterAll');
        expect(result).toContain('driver.quit()');
      });
    });
  });
});
