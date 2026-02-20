import { spawnSync } from 'child_process';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '../..');
const cliPath = path.resolve(rootDir, 'bin/hamlet.js');
const fixturesDir = path.resolve(__dirname, '../fixtures');
const outputDir = path.resolve(__dirname, '../output');

// Helper to run CLI commands safely without shell interpolation.
// Uses spawnSync with an explicit timeout and no stdin pipe to avoid
// leaving handles that delay Jest worker shutdown.
function runCLI(args, options = {}) {
  const { stdio = ['ignore', 'pipe', 'pipe'], ...rest } = options;
  const result = spawnSync('node', [cliPath, ...args], {
    encoding: 'utf8',
    timeout: 15000,
    stdio,
    ...rest,
  });
  if (result.error) throw result.error;
  if (result.status !== 0) {
    const err = new Error(
      result.stderr || `Process exited with code ${result.status}`,
    );
    err.stdout = result.stdout;
    err.stderr = result.stderr;
    err.status = result.status;
    throw err;
  }
  return result.stdout;
}

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

    await fs.writeFile(path.join(fixturesDir, 'sample.jest.js'), `
describe('Sample Test', () => {
  const mockFn = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should call the function', () => {
    mockFn('hello');
    expect(mockFn).toHaveBeenCalledWith('hello');
  });

  it('should use fake timers', () => {
    jest.useFakeTimers();
    setTimeout(() => mockFn(), 1000);
    jest.advanceTimersByTime(1000);
    expect(mockFn).toHaveBeenCalled();
    jest.useRealTimers();
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
    // Clean up output directory (recursive handles subdirs left by tests)
    await fs.rm(outputDir, { recursive: true, force: true }).catch(() => {});
    await fs.mkdir(outputDir, { recursive: true });

    // Clean up fixture subdirectories created during tests
    await fs
      .rm(path.join(fixturesDir, 'multi'), { recursive: true, force: true })
      .catch(() => {});
  });

  describe('Help Command', () => {
    test('should display help information', () => {
      const result = runCLI(['--help']);
      expect(result).toContain('hamlet');
      expect(result).toContain('convert');
    });

    test('should display version', () => {
      const result = runCLI(['--version']);
      expect(result).toMatch(/\d+\.\d+\.\d+/);
    });
  });

  describe('Convert Command - Cypress to Playwright', () => {
    test('should convert Cypress file to Playwright', async () => {
      const inputFile = path.resolve(fixturesDir, 'sample.cy.js');
      const outputFile = path.resolve(outputDir, 'sample.spec.js');

      runCLI(['convert', inputFile, '--from', 'cypress', '--to', 'playwright', '-o', outputDir]);

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
      const inputFile = path.resolve(fixturesDir, 'sample.cy.js');
      const outputFile = path.resolve(outputDir, 'sample.test.js');

      runCLI(['convert', inputFile, '--from', 'cypress', '--to', 'selenium', '-o', outputDir]);

      const output = await fs.readFile(outputFile, 'utf8');
      expect(output).toContain("require('selenium-webdriver')");
      expect(output).toContain('driver.get');
      expect(output).toContain('driver.findElement');
      expect(output).toContain('By.css');
    });
  });

  describe('Convert Command - Playwright to Cypress', () => {
    test('should convert Playwright file to Cypress', async () => {
      const inputFile = path.resolve(fixturesDir, 'sample.spec.ts');
      const outputFile = path.resolve(outputDir, 'sample.cy.js');

      runCLI(['convert', inputFile, '--from', 'playwright', '--to', 'cypress', '-o', outputDir]);

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
      const inputFile = path.resolve(fixturesDir, 'sample.spec.ts');
      const outputFile = path.resolve(outputDir, 'sample.test.js');

      // Clean up from previous test
      try {
        await fs.unlink(outputFile);
      } catch (e) {}

      runCLI(['convert', inputFile, '--from', 'playwright', '--to', 'selenium', '-o', outputDir]);

      const output = await fs.readFile(outputFile, 'utf8');
      expect(output).toContain("require('selenium-webdriver')");
      expect(output).toContain('driver.get');
      expect(output).toContain('driver.findElement');
    });
  });

  describe('Convert Command - Selenium to Cypress', () => {
    test('should convert Selenium file to Cypress', async () => {
      const inputFile = path.resolve(fixturesDir, 'sample.selenium.js');
      const outputFile = path.resolve(outputDir, 'sample.selenium.cy.js');

      // Clean up from previous test
      try {
        await fs.unlink(outputFile);
      } catch (e) {}

      runCLI(['convert', inputFile, '--from', 'selenium', '--to', 'cypress', '-o', outputDir]);

      const output = await fs.readFile(outputFile, 'utf8');
      expect(output).toContain('/// <reference types="cypress" />');
      expect(output).toContain('cy.visit');
      expect(output).toContain('cy.get');
      expect(output).toContain("should('be.visible')");
    });
  });

  describe('Convert Command - Selenium to Playwright', () => {
    test('should convert Selenium file to Playwright', async () => {
      const inputFile = path.resolve(fixturesDir, 'sample.selenium.js');
      const outputFile = path.resolve(outputDir, 'sample.selenium.spec.js');

      // Clean up from previous test
      try {
        await fs.unlink(outputFile);
      } catch (e) {}

      runCLI(['convert', inputFile, '--from', 'selenium', '--to', 'playwright', '-o', outputDir]);

      const output = await fs.readFile(outputFile, 'utf8');
      expect(output).toContain("import { test, expect } from '@playwright/test'");
      expect(output).toContain('page.goto');
      expect(output).toContain('page.locator');
      expect(output).toContain('toBeVisible');
    });
  });

  describe('Convert Command - Jest to Vitest', () => {
    test('should convert Jest file to Vitest', async () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const outputFile = path.resolve(outputDir, 'sample.jest.test.js');

      // Clean up from previous test
      try {
        await fs.unlink(outputFile);
      } catch (e) {}

      runCLI(['convert', inputFile, '--from', 'jest', '--to', 'vitest', '-o', outputDir]);

      const output = await fs.readFile(outputFile, 'utf8');
      expect(output).toContain("from 'vitest'");
      expect(output).toContain('vi.fn()');
      expect(output).toContain('vi.clearAllMocks()');
      expect(output).toContain('vi.useFakeTimers()');
      expect(output).toContain('vi.advanceTimersByTime(');
      expect(output).toContain('vi.useRealTimers()');
    });
  });

  describe('Shorthand Commands', () => {
    test('jest2vt shorthand should convert Jest to Vitest', async () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const outputFile = path.resolve(outputDir, 'sample.jest.test.js');

      // Clean up from previous test
      try {
        await fs.unlink(outputFile);
      } catch (e) {}

      runCLI(['jest2vt', inputFile, '-o', outputDir]);

      const output = await fs.readFile(outputFile, 'utf8');
      expect(output).toContain("from 'vitest'");
      expect(output).toContain('vi.fn()');
    });
  });

  describe('Error Handling', () => {
    test('should error on missing input file', () => {
      expect(() => {
        runCLI(['convert', 'nonexistent.js', '--from', 'cypress', '--to', 'playwright', '-o', outputDir]);
      }).toThrow();
    });

    test('should error on invalid source framework', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.cy.js');
      expect(() => {
        runCLI(['convert', inputFile, '--from', 'invalid', '--to', 'playwright', '-o', outputDir]);
      }).toThrow();
    });

    test('should error on invalid target framework', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.cy.js');
      expect(() => {
        runCLI(['convert', inputFile, '--from', 'cypress', '--to', 'invalid', '-o', outputDir]);
      }).toThrow();
    });

    test('should error when source and target are the same', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.cy.js');
      expect(() => {
        runCLI(['convert', inputFile, '--from', 'cypress', '--to', 'cypress', '-o', outputDir]);
      }).toThrow();
    });
  });

  describe('Directory Conversion', () => {
    test('should convert all files in a directory', async () => {
      // Create a subdirectory with multiple test files
      const subDir = path.resolve(fixturesDir, 'multi');
      await fs.mkdir(subDir, { recursive: true });

      await fs.writeFile(path.resolve(subDir, 'test1.cy.js'), `
describe('Test 1', () => {
  it('test case 1', () => {
    cy.visit('/page1');
  });
});
`);

      await fs.writeFile(path.resolve(subDir, 'test2.cy.js'), `
describe('Test 2', () => {
  it('test case 2', () => {
    cy.visit('/page2');
  });
});
`);

      const multiOutputDir = path.resolve(outputDir, 'multi');
      await fs.mkdir(multiOutputDir, { recursive: true });

      runCLI(['convert', subDir, '--from', 'cypress', '--to', 'playwright', '-o', multiOutputDir]);

      const files = await fs.readdir(multiOutputDir);
      expect(files.length).toBeGreaterThan(0);

      // Check at least one file was converted correctly
      const outputFiles = files.filter(f => f.endsWith('.spec.js'));
      expect(outputFiles.length).toBeGreaterThan(0);
    });
  });

  describe('Migrate Command', () => {
    let migrateDir;
    let migrateOutput;

    beforeEach(async () => {
      migrateDir = path.resolve(outputDir, 'migrate-src');
      migrateOutput = path.resolve(outputDir, 'migrate-out');
      await fs.mkdir(migrateDir, { recursive: true });
      await fs.mkdir(migrateOutput, { recursive: true });
    });

    afterEach(async () => {
      await fs.rm(migrateDir, { recursive: true, force: true }).catch(() => {});
      await fs.rm(migrateOutput, { recursive: true, force: true }).catch(() => {});
    });

    test('should run migrate command successfully', async () => {
      await fs.writeFile(
        path.join(migrateDir, 'app.test.js'),
        `describe('app', () => { it('works', () => { expect(1).toBe(1); }); });`
      );

      const result = runCLI([
        'migrate', migrateDir,
        '--from', 'jest', '--to', 'vitest',
        '-o', migrateOutput,
      ]);

      expect(result).toContain('Migration complete');
    });

    test('should show progress for each file', async () => {
      await fs.writeFile(
        path.join(migrateDir, 'a.test.js'),
        `describe('a', () => { it('works', () => { expect(1).toBe(1); }); });`
      );

      const result = runCLI([
        'migrate', migrateDir,
        '--from', 'jest', '--to', 'vitest',
        '-o', migrateOutput,
      ]);

      expect(result).toContain('a.test.js');
    });

    test('should handle empty project directory', async () => {
      const result = runCLI([
        'migrate', migrateDir,
        '--from', 'jest', '--to', 'vitest',
        '-o', migrateOutput,
      ]);

      expect(result).toContain('Migration complete');
    });

    test('should create .hamlet state directory', async () => {
      await fs.writeFile(
        path.join(migrateDir, 'test.test.js'),
        `describe('t', () => { it('w', () => { expect(1).toBe(1); }); });`
      );

      runCLI([
        'migrate', migrateDir,
        '--from', 'jest', '--to', 'vitest',
        '-o', migrateOutput,
      ]);

      const stateExists = await fs.access(path.join(migrateDir, '.hamlet', 'state.json'))
        .then(() => true).catch(() => false);
      expect(stateExists).toBe(true);
    });
  });

  describe('Estimate Command', () => {
    let estimateDir;

    beforeEach(async () => {
      estimateDir = path.resolve(outputDir, 'estimate-src');
      await fs.mkdir(estimateDir, { recursive: true });
    });

    afterEach(async () => {
      await fs.rm(estimateDir, { recursive: true, force: true }).catch(() => {});
    });

    test('should run estimate command and show summary', async () => {
      await fs.writeFile(
        path.join(estimateDir, 'simple.test.js'),
        `describe('simple', () => { it('works', () => { expect(1).toBe(1); }); });`
      );

      const result = runCLI([
        'estimate', estimateDir,
        '--from', 'jest', '--to', 'vitest',
      ]);

      expect(result).toContain('Estimation Summary');
      expect(result).toContain('Total files');
    });

    test('should handle empty directory', async () => {
      const result = runCLI([
        'estimate', estimateDir,
        '--from', 'jest', '--to', 'vitest',
      ]);

      expect(result).toContain('Estimation Summary');
      expect(result).toContain('Total files: 0');
    });

    test('should show effort estimate', async () => {
      await fs.writeFile(
        path.join(estimateDir, 'test.test.js'),
        `describe('test', () => { it('works', () => { expect(1).toBe(1); }); });`
      );

      const result = runCLI([
        'estimate', estimateDir,
        '--from', 'jest', '--to', 'vitest',
      ]);

      expect(result).toContain('Effort Estimate');
    });

    test('should NOT create .hamlet directory', async () => {
      await fs.writeFile(
        path.join(estimateDir, 'test.test.js'),
        `describe('t', () => { it('w', () => { expect(1).toBe(1); }); });`
      );

      runCLI([
        'estimate', estimateDir,
        '--from', 'jest', '--to', 'vitest',
      ]);

      const hamletExists = await fs.access(path.join(estimateDir, '.hamlet'))
        .then(() => true).catch(() => false);
      expect(hamletExists).toBe(false);
    });

    test('should show blockers for complex files', async () => {
      await fs.writeFile(
        path.join(estimateDir, 'complex.test.js'),
        `jest.mock('./module');\njest.spyOn(obj, 'method');\njest.mock('./another');\n\ndescribe('x', () => { it('w', () => { expect(1).toBe(1); }); });`
      );

      const result = runCLI([
        'estimate', estimateDir,
        '--from', 'jest', '--to', 'vitest',
      ]);

      expect(result).toContain('Top Blockers');
    });
  });

  describe('Status Command', () => {
    let statusDir;

    beforeEach(async () => {
      statusDir = path.resolve(outputDir, 'status-src');
      await fs.mkdir(statusDir, { recursive: true });
    });

    afterEach(async () => {
      await fs.rm(statusDir, { recursive: true, force: true }).catch(() => {});
    });

    test('should show "no migration" when .hamlet does not exist', () => {
      const result = runCLI(['status', '-d', statusDir]);
      expect(result).toContain('No migration in progress');
    });

    test('should show migration status after migrate', async () => {
      await fs.writeFile(
        path.join(statusDir, 'test.test.js'),
        `describe('t', () => { it('w', () => { expect(1).toBe(1); }); });`
      );

      const statusOutput = path.resolve(outputDir, 'status-out');
      await fs.mkdir(statusOutput, { recursive: true });

      runCLI([
        'migrate', statusDir,
        '--from', 'jest', '--to', 'vitest',
        '-o', statusOutput,
      ]);

      const result = runCLI(['status', '-d', statusDir]);

      expect(result).toContain('Migration Status');
      expect(result).toContain('jest');
      expect(result).toContain('vitest');

      await fs.rm(statusOutput, { recursive: true, force: true }).catch(() => {});
    });
  });

  describe('Checklist Command', () => {
    let checklistDir;

    beforeEach(async () => {
      checklistDir = path.resolve(outputDir, 'checklist-src');
      await fs.mkdir(checklistDir, { recursive: true });
    });

    afterEach(async () => {
      await fs.rm(checklistDir, { recursive: true, force: true }).catch(() => {});
    });

    test('should show "no migration" when .hamlet does not exist', () => {
      const result = runCLI(['checklist', '-d', checklistDir]);
      expect(result).toContain('No migration in progress');
    });

    test('should generate checklist after migrate', async () => {
      await fs.writeFile(
        path.join(checklistDir, 'test.test.js'),
        `describe('t', () => { it('w', () => { expect(1).toBe(1); }); });`
      );

      const checklistOutput = path.resolve(outputDir, 'checklist-out');
      await fs.mkdir(checklistOutput, { recursive: true });

      runCLI([
        'migrate', checklistDir,
        '--from', 'jest', '--to', 'vitest',
        '-o', checklistOutput,
      ]);

      const result = runCLI(['checklist', '-d', checklistDir]);

      expect(result).toContain('Migration Checklist');

      await fs.rm(checklistOutput, { recursive: true, force: true }).catch(() => {});
    });
  });

  describe('Reset Command', () => {
    let resetDir;

    beforeEach(async () => {
      resetDir = path.resolve(outputDir, 'reset-src');
      await fs.mkdir(resetDir, { recursive: true });
    });

    afterEach(async () => {
      await fs.rm(resetDir, { recursive: true, force: true }).catch(() => {});
    });

    test('should show "no state" when .hamlet does not exist', () => {
      const result = runCLI(['reset', '-d', resetDir, '--yes']);
      expect(result).toContain('No migration state');
    });

    test('should require --yes flag', async () => {
      await fs.mkdir(path.join(resetDir, '.hamlet'), { recursive: true });
      await fs.writeFile(
        path.join(resetDir, '.hamlet', 'state.json'),
        JSON.stringify({ version: 1, files: {} })
      );

      const result = runCLI(['reset', '-d', resetDir]);
      expect(result).toContain('Use --yes');
    });

    test('should clear .hamlet directory with --yes', async () => {
      await fs.mkdir(path.join(resetDir, '.hamlet'), { recursive: true });
      await fs.writeFile(
        path.join(resetDir, '.hamlet', 'state.json'),
        JSON.stringify({ version: 1, files: {} })
      );

      const result = runCLI(['reset', '-d', resetDir, '--yes']);
      expect(result).toContain('Migration state cleared');

      const hamletExists = await fs.access(path.join(resetDir, '.hamlet'))
        .then(() => true).catch(() => false);
      expect(hamletExists).toBe(false);
    });
  });

  describe('New Commands - Help', () => {
    test('should list migrate command in help', () => {
      const result = runCLI(['--help']);
      expect(result).toContain('migrate');
    });

    test('should list estimate command in help', () => {
      const result = runCLI(['--help']);
      expect(result).toContain('estimate');
    });

    test('should list status command in help', () => {
      const result = runCLI(['--help']);
      expect(result).toContain('status');
    });

    test('should list checklist command in help', () => {
      const result = runCLI(['--help']);
      expect(result).toContain('checklist');
    });

    test('should list reset command in help', () => {
      const result = runCLI(['--help']);
      expect(result).toContain('reset');
    });
  });
});
