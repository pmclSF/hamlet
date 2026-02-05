import { execSync, exec } from 'child_process';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.join(__dirname, '../..');
const cliPath = path.join(rootDir, 'bin/hamlet.js');
const fixturesDir = path.join(__dirname, '../fixtures');
const outputDir = path.join(__dirname, '../output');

describe('CLI Integration Tests', () => {
  beforeAll(async () => {
    // Create fixtures directory if it doesn't exist
    await fs.mkdir(fixturesDir, { recursive: true });
    await fs.mkdir(outputDir, { recursive: true });

    // Create test fixture files
    await fs.writeFile(path.join(fixturesDir, 'sample.cy.js'), `
describe('Sample Test', () => {
  it('should navigate and click', () => {
    cy.visit('/home');
    cy.get('#button').click();
    cy.get('.result').should('be.visible');
  });
});
`);

    await fs.writeFile(path.join(fixturesDir, 'sample.spec.ts'), `
import { test, expect } from '@playwright/test';

test.describe('Sample Test', () => {
  test('should navigate and click', async ({ page }) => {
    await page.goto('/home');
    await page.locator('#button').click();
    await expect(page.locator('.result')).toBeVisible();
  });
});
`);

    await fs.writeFile(path.join(fixturesDir, 'sample.selenium.js'), `
const { Builder, By } = require('selenium-webdriver');
const { expect } = require('@jest/globals');

let driver;

beforeAll(async () => {
  driver = await new Builder().forBrowser('chrome').build();
});

afterAll(async () => {
  await driver.quit();
});

describe('Sample Test', () => {
  it('should navigate and click', async () => {
    await driver.get('/home');
    await driver.findElement(By.css('#button')).click();
    expect(await (await driver.findElement(By.css('.result'))).isDisplayed()).toBe(true);
  });
});
`);
  });

  afterAll(async () => {
    // Clean up output directory
    try {
      const files = await fs.readdir(outputDir);
      for (const file of files) {
        await fs.unlink(path.join(outputDir, file));
      }
    } catch (e) {
      // Ignore cleanup errors
    }
  });

  describe('Help Command', () => {
    test('should display help information', () => {
      const result = execSync(`node ${cliPath} --help`, { encoding: 'utf8' });
      expect(result).toContain('hamlet');
      expect(result).toContain('convert');
    });

    test('should display version', () => {
      const result = execSync(`node ${cliPath} --version`, { encoding: 'utf8' });
      expect(result).toMatch(/\d+\.\d+\.\d+/);
    });
  });

  describe('Convert Command - Cypress to Playwright', () => {
    test('should convert Cypress file to Playwright', async () => {
      const inputFile = path.join(fixturesDir, 'sample.cy.js');
      const outputFile = path.join(outputDir, 'sample.spec.js');

      execSync(`node ${cliPath} convert ${inputFile} --from cypress --to playwright -o ${outputDir}`, {
        encoding: 'utf8'
      });

      const output = await fs.readFile(outputFile, 'utf8');
      expect(output).toContain("import { test, expect } from '@playwright/test'");
      expect(output).toContain('test.describe');
      expect(output).toContain('page.goto');
      expect(output).toContain('page.locator');
      expect(output).toContain('toBeVisible');
    });
  });

  describe('Convert Command - Cypress to Selenium', () => {
    test('should convert Cypress file to Selenium', async () => {
      const inputFile = path.join(fixturesDir, 'sample.cy.js');
      const outputFile = path.join(outputDir, 'sample.test.js');

      execSync(`node ${cliPath} convert ${inputFile} --from cypress --to selenium -o ${outputDir}`, {
        encoding: 'utf8'
      });

      const output = await fs.readFile(outputFile, 'utf8');
      expect(output).toContain("require('selenium-webdriver')");
      expect(output).toContain('driver.get');
      expect(output).toContain('driver.findElement');
      expect(output).toContain('By.css');
    });
  });

  describe('Convert Command - Playwright to Cypress', () => {
    test('should convert Playwright file to Cypress', async () => {
      const inputFile = path.join(fixturesDir, 'sample.spec.ts');
      const outputFile = path.join(outputDir, 'sample.cy.js');

      execSync(`node ${cliPath} convert ${inputFile} --from playwright --to cypress -o ${outputDir}`, {
        encoding: 'utf8'
      });

      const output = await fs.readFile(outputFile, 'utf8');
      expect(output).toContain('/// <reference types="cypress" />');
      expect(output).toContain('describe(');
      expect(output).toContain('cy.visit');
      expect(output).toContain('cy.get');
      expect(output).toContain("should('be.visible')");
    });
  });

  describe('Convert Command - Playwright to Selenium', () => {
    test('should convert Playwright file to Selenium', async () => {
      const inputFile = path.join(fixturesDir, 'sample.spec.ts');
      const outputFile = path.join(outputDir, 'sample.test.js');

      // Clean up from previous test
      try {
        await fs.unlink(outputFile);
      } catch (e) {}

      execSync(`node ${cliPath} convert ${inputFile} --from playwright --to selenium -o ${outputDir}`, {
        encoding: 'utf8'
      });

      const output = await fs.readFile(outputFile, 'utf8');
      expect(output).toContain("require('selenium-webdriver')");
      expect(output).toContain('driver.get');
      expect(output).toContain('driver.findElement');
    });
  });

  describe('Convert Command - Selenium to Cypress', () => {
    test('should convert Selenium file to Cypress', async () => {
      const inputFile = path.join(fixturesDir, 'sample.selenium.js');
      const outputFile = path.join(outputDir, 'sample.selenium.cy.js');

      // Clean up from previous test
      try {
        await fs.unlink(outputFile);
      } catch (e) {}

      execSync(`node ${cliPath} convert ${inputFile} --from selenium --to cypress -o ${outputDir}`, {
        encoding: 'utf8'
      });

      const output = await fs.readFile(outputFile, 'utf8');
      expect(output).toContain('/// <reference types="cypress" />');
      expect(output).toContain('cy.visit');
      expect(output).toContain('cy.get');
      expect(output).toContain("should('be.visible')");
    });
  });

  describe('Convert Command - Selenium to Playwright', () => {
    test('should convert Selenium file to Playwright', async () => {
      const inputFile = path.join(fixturesDir, 'sample.selenium.js');
      const outputFile = path.join(outputDir, 'sample.selenium.spec.js');

      // Clean up from previous test
      try {
        await fs.unlink(outputFile);
      } catch (e) {}

      execSync(`node ${cliPath} convert ${inputFile} --from selenium --to playwright -o ${outputDir}`, {
        encoding: 'utf8'
      });

      const output = await fs.readFile(outputFile, 'utf8');
      expect(output).toContain("import { test, expect } from '@playwright/test'");
      expect(output).toContain('page.goto');
      expect(output).toContain('page.locator');
      expect(output).toContain('toBeVisible');
    });
  });

  describe('Error Handling', () => {
    test('should error on missing input file', () => {
      expect(() => {
        execSync(`node ${cliPath} convert nonexistent.js --from cypress --to playwright -o ${outputDir}`, {
          encoding: 'utf8',
          stdio: 'pipe'
        });
      }).toThrow();
    });

    test('should error on invalid source framework', () => {
      expect(() => {
        execSync(`node ${cliPath} convert ${path.join(fixturesDir, 'sample.cy.js')} --from invalid --to playwright -o ${outputDir}`, {
          encoding: 'utf8',
          stdio: 'pipe'
        });
      }).toThrow();
    });

    test('should error on invalid target framework', () => {
      expect(() => {
        execSync(`node ${cliPath} convert ${path.join(fixturesDir, 'sample.cy.js')} --from cypress --to invalid -o ${outputDir}`, {
          encoding: 'utf8',
          stdio: 'pipe'
        });
      }).toThrow();
    });

    test('should error when source and target are the same', () => {
      expect(() => {
        execSync(`node ${cliPath} convert ${path.join(fixturesDir, 'sample.cy.js')} --from cypress --to cypress -o ${outputDir}`, {
          encoding: 'utf8',
          stdio: 'pipe'
        });
      }).toThrow();
    });
  });

  describe('Directory Conversion', () => {
    test('should convert all files in a directory', async () => {
      // Create a subdirectory with multiple test files
      const subDir = path.join(fixturesDir, 'multi');
      await fs.mkdir(subDir, { recursive: true });

      await fs.writeFile(path.join(subDir, 'test1.cy.js'), `
describe('Test 1', () => {
  it('test case 1', () => {
    cy.visit('/page1');
  });
});
`);

      await fs.writeFile(path.join(subDir, 'test2.cy.js'), `
describe('Test 2', () => {
  it('test case 2', () => {
    cy.visit('/page2');
  });
});
`);

      const multiOutputDir = path.join(outputDir, 'multi');
      await fs.mkdir(multiOutputDir, { recursive: true });

      execSync(`node ${cliPath} convert ${subDir} --from cypress --to playwright -o ${multiOutputDir}`, {
        encoding: 'utf8'
      });

      const files = await fs.readdir(multiOutputDir);
      expect(files.length).toBeGreaterThan(0);

      // Check at least one file was converted correctly
      const outputFiles = files.filter(f => f.endsWith('.spec.js'));
      expect(outputFiles.length).toBeGreaterThan(0);
    });
  });
});
