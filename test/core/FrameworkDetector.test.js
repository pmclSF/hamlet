import { FrameworkDetector } from '../../src/core/FrameworkDetector.js';

describe('FrameworkDetector', () => {
  describe('detectFromContent', () => {
    describe('Cypress detection', () => {
      it('should detect Cypress from cy.visit', () => {
        const content = `cy.visit('/home');`;
        const result = FrameworkDetector.detectFromContent(content);
        expect(result.framework).toBe('cypress');
        expect(result.confidence).toBeGreaterThan(0.5);
      });

      it('should detect Cypress from cy.get', () => {
        const content = `cy.get('.element').click();`;
        const result = FrameworkDetector.detectFromContent(content);
        expect(result.framework).toBe('cypress');
      });

      it('should detect Cypress from describe/it pattern', () => {
        const content = `describe('Test', () => { it('works', () => { cy.visit('/'); }); });`;
        const result = FrameworkDetector.detectFromContent(content);
        expect(result.framework).toBe('cypress');
      });

      it('should detect Cypress from should assertions', () => {
        const content = `cy.get('.element').should('be.visible');`;
        const result = FrameworkDetector.detectFromContent(content);
        expect(result.framework).toBe('cypress');
      });
    });

    describe('Playwright detection', () => {
      it('should detect Playwright from page.goto', () => {
        const content = `await page.goto('/home');`;
        const result = FrameworkDetector.detectFromContent(content);
        expect(result.framework).toBe('playwright');
        expect(result.confidence).toBeGreaterThan(0.5);
      });

      it('should detect Playwright from page.locator', () => {
        const content = `await page.locator('.element').click();`;
        const result = FrameworkDetector.detectFromContent(content);
        expect(result.framework).toBe('playwright');
      });

      it('should detect Playwright from test.describe', () => {
        const content = `test.describe('Test', () => { test('works', async () => {}); });`;
        const result = FrameworkDetector.detectFromContent(content);
        expect(result.framework).toBe('playwright');
      });

      it('should detect Playwright from expect assertions', () => {
        const content = `await expect(page.locator('.element')).toBeVisible();`;
        const result = FrameworkDetector.detectFromContent(content);
        expect(result.framework).toBe('playwright');
      });

      it('should detect Playwright from import', () => {
        const content = `import { test, expect } from '@playwright/test';`;
        const result = FrameworkDetector.detectFromContent(content);
        expect(result.framework).toBe('playwright');
      });
    });

    describe('Selenium detection', () => {
      it('should detect Selenium from driver.get', () => {
        const content = `await driver.get('/home');`;
        const result = FrameworkDetector.detectFromContent(content);
        expect(result.framework).toBe('selenium');
        expect(result.confidence).toBeGreaterThan(0.5);
      });

      it('should detect Selenium from driver.findElement', () => {
        const content = `await driver.findElement(By.css('.element')).click();`;
        const result = FrameworkDetector.detectFromContent(content);
        expect(result.framework).toBe('selenium');
      });

      it('should detect Selenium from WebDriver import', () => {
        const content = `const { Builder, By } = require('selenium-webdriver');`;
        const result = FrameworkDetector.detectFromContent(content);
        expect(result.framework).toBe('selenium');
      });

      it('should detect Selenium from By selector', () => {
        const content = `By.css('.element')`;
        const result = FrameworkDetector.detectFromContent(content);
        expect(result.framework).toBe('selenium');
      });
    });

    describe('Unknown detection', () => {
      it('should return null for unknown content', () => {
        const content = `console.log('hello');`;
        const result = FrameworkDetector.detectFromContent(content);
        expect(result.framework).toBeNull();
        expect(result.confidence).toBe(0);
      });

      it('should return null for empty content', () => {
        const content = '';
        const result = FrameworkDetector.detectFromContent(content);
        expect(result.framework).toBeNull();
      });
    });
  });

  describe('detectFromPath', () => {
    it('should detect Cypress from .cy.js extension', () => {
      const result = FrameworkDetector.detectFromPath('login.cy.js');
      expect(result.framework).toBe('cypress');
    });

    it('should detect Cypress from .cy.ts extension', () => {
      const result = FrameworkDetector.detectFromPath('login.cy.ts');
      expect(result.framework).toBe('cypress');
    });

    it('should detect Cypress from cypress directory', () => {
      const result = FrameworkDetector.detectFromPath('cypress/e2e/login.js');
      expect(result.framework).toBe('cypress');
    });

    it('should detect Playwright from .spec.js extension', () => {
      const result = FrameworkDetector.detectFromPath('login.spec.js');
      expect(result.framework).toBe('playwright');
    });

    it('should detect Playwright from .spec.ts extension', () => {
      const result = FrameworkDetector.detectFromPath('login.spec.ts');
      expect(result.framework).toBe('playwright');
    });

    it('should return null for generic .test.js extension', () => {
      // .test.js is ambiguous - could be any framework
      const result = FrameworkDetector.detectFromPath('login.test.js');
      // Either null or a low-confidence detection is acceptable
      expect(result.confidence).toBeLessThan(0.8);
    });

    it('should return null for unknown path patterns', () => {
      const result = FrameworkDetector.detectFromPath('random-file.js');
      expect(result.framework).toBeNull();
    });
  });

  describe('detect', () => {
    it('should combine content and path detection', () => {
      const content = `cy.visit('/home');`;
      const path = 'tests/login.cy.js';
      const result = FrameworkDetector.detect(content, path);
      expect(result.framework).toBe('cypress');
      expect(result.confidence).toBeGreaterThan(0.5);
    });

    it('should prefer content detection over path detection', () => {
      const content = `await page.goto('/home');`;
      const path = 'tests/login.cy.js'; // Cypress path but Playwright content
      const result = FrameworkDetector.detect(content, path);
      expect(result.framework).toBe('playwright');
    });

    it('should use path detection when content is ambiguous', () => {
      const content = `console.log('test');`;
      const path = 'cypress/e2e/login.cy.js';
      const result = FrameworkDetector.detect(content, path);
      expect(result.framework).toBe('cypress');
    });
  });
});
