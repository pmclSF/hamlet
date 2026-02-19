import { ConverterFactory } from '../../src/core/ConverterFactory.js';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const fixturesDir = path.join(__dirname, '../fixtures/configs');

describe('Config File Conversion', () => {
  beforeAll(async () => {
    await fs.mkdir(fixturesDir, { recursive: true });

    // Create Cypress config fixture
    await fs.writeFile(path.join(fixturesDir, 'cypress.config.js'), `
const { defineConfig } = require('cypress');

module.exports = defineConfig({
  e2e: {
    baseUrl: 'http://localhost:3000',
    viewportWidth: 1280,
    viewportHeight: 720,
    video: true,
    screenshotOnRunFailure: true,
    defaultCommandTimeout: 10000,
    specPattern: 'cypress/e2e/**/*.cy.{js,ts}'
  }
});
`);

    // Create Playwright config fixture
    await fs.writeFile(path.join(fixturesDir, 'playwright.config.ts'), `
import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  timeout: 30000,
  expect: {
    timeout: 5000
  },
  use: {
    baseURL: 'http://localhost:3000',
    viewport: { width: 1920, height: 1080 },
    video: 'on',
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure'
  },
  projects: [
    { name: 'chromium', use: { browserName: 'chromium' } },
    { name: 'firefox', use: { browserName: 'firefox' } }
  ]
});
`);
  });

  describe('Cypress to Playwright Config', () => {
    test('should convert Cypress config to Playwright format', async () => {
      const converter = await ConverterFactory.createConverter('cypress', 'playwright');
      const configPath = path.join(fixturesDir, 'cypress.config.js');

      const result = await converter.convertConfig(configPath);

      expect(result).toContain("@playwright/test");
      expect(result).toContain('defineConfig');
      expect(result).toContain('baseURL');
      expect(result).toContain('viewport');
    });
  });

  describe('Cypress to Selenium Config', () => {
    test('should convert Cypress config to Selenium format', async () => {
      const converter = await ConverterFactory.createConverter('cypress', 'selenium');
      const configPath = path.join(fixturesDir, 'cypress.config.js');

      const result = await converter.convertConfig(configPath);

      expect(result).toContain('module.exports');
      expect(result).toContain('capabilities');
      expect(result).toContain('browserName');
    });
  });

  describe('Playwright to Cypress Config', () => {
    test('should convert Playwright config to Cypress format', async () => {
      const converter = await ConverterFactory.createConverter('playwright', 'cypress');
      const configPath = path.join(fixturesDir, 'playwright.config.ts');

      const result = await converter.convertConfig(configPath);

      expect(result).toContain('defineConfig');
      expect(result).toContain('e2e');
      expect(result).toContain('baseUrl');
    });
  });

  describe('Playwright to Selenium Config', () => {
    test('should convert Playwright config to Selenium format', async () => {
      const converter = await ConverterFactory.createConverter('playwright', 'selenium');
      const configPath = path.join(fixturesDir, 'playwright.config.ts');

      const result = await converter.convertConfig(configPath);

      expect(result).toContain('module.exports');
      expect(result).toContain('capabilities');
      expect(result).toContain('browserName');
    });
  });

  describe('Selenium to Cypress Config', () => {
    test('should generate Cypress config from Selenium defaults', async () => {
      const converter = await ConverterFactory.createConverter('selenium', 'cypress');

      // Selenium typically doesn't have a standard config file, so we generate defaults
      const result = await converter.convertConfig('', {});

      expect(result).toContain('defineConfig');
      expect(result).toContain('e2e');
    });
  });

  describe('Selenium to Playwright Config', () => {
    test('should generate Playwright config from Selenium defaults', async () => {
      const converter = await ConverterFactory.createConverter('selenium', 'playwright');

      const result = await converter.convertConfig('', {});

      expect(result).toContain("import { defineConfig } from '@playwright/test'");
      expect(result).toContain('testDir');
    });
  });
});
