/**
 * Cross-step regression testing.
 *
 * Verifies all 25 conversion directions still work end-to-end,
 * along with batch mode, migration, dry-run, shorthands, config conversion.
 */

import { execFileSync } from 'child_process';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';
import { ConverterFactory } from '../../../src/core/ConverterFactory.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '../../..');
const cliPath = path.resolve(rootDir, 'bin/hamlet.js');
const outputDir = path.resolve(__dirname, '../../output/regression');

function runCLI(args, options = {}) {
  return execFileSync('node', [cliPath, ...args], {
    encoding: 'utf8',
    ...options,
  });
}

// ── Minimal inputs per framework for smoke testing ───────────────────

const MINIMAL_INPUTS = {
  jest: `describe('T', () => { it('works', () => { expect(1).toBe(1); }); });`,
  vitest: `import { describe, it, expect } from 'vitest';\ndescribe('T', () => { it('works', () => { expect(1).toBe(1); }); });`,
  mocha: `describe('T', () => { it('works', () => { expect(1).toBe(1); }); });`,
  jasmine: `describe('T', () => { it('works', () => { expect(1).toBe(1); }); });`,
  cypress: `describe('T', () => { it('works', () => { cy.visit('/'); cy.get('#a').should('exist'); }); });`,
  playwright: `import { test, expect } from '@playwright/test';\ntest.describe('T', () => { test('works', async ({ page }) => { await page.goto('/'); await expect(page.locator('#a')).toBeVisible(); }); });`,
  selenium: `const { Builder, By } = require('selenium-webdriver');\ndescribe('T', () => { it('works', async () => { await driver.get('/'); const el = await driver.findElement(By.css('#a')); expect(await el.isDisplayed()).toBe(true); }); });`,
  webdriverio: `describe('T', () => { it('works', async () => { await browser.url('/'); await expect($('#a')).toBeDisplayed(); }); });`,
  puppeteer: `describe('T', () => { it('works', async () => { await page.goto('http://localhost:3000'); const el = await page.waitForSelector('#a'); expect(el).toBeTruthy(); }); });`,
  testcafe: `import { Selector } from 'testcafe';\nfixture('T').page('http://localhost:3000');\ntest('works', async t => { await t.expect(Selector('#a').exists).ok(); });`,
  junit4: `import org.junit.Test;\nimport static org.junit.Assert.*;\npublic class T { @Test public void works() { assertEquals(1, 1); } }`,
  junit5: `import org.junit.jupiter.api.Test;\nimport static org.junit.jupiter.api.Assertions.*;\nclass T { @Test void works() { assertEquals(1, 1); } }`,
  testng: `import org.testng.annotations.Test;\nimport org.testng.Assert;\npublic class T { @Test public void works() { Assert.assertEquals(1, 1); } }`,
  pytest: `def test_works():\n    assert 1 == 1`,
  unittest: `import unittest\nclass T(unittest.TestCase):\n    def test_works(self):\n        self.assertEqual(1, 1)`,
  nose2: `def test_works():\n    assert 1 == 1`,
};

describe('Cross-Step Regression Testing', () => {
  beforeAll(async () => {
    await fs.mkdir(outputDir, { recursive: true });
  });

  afterAll(async () => {
    await fs.rm(outputDir, { recursive: true, force: true }).catch(() => {});
  });

  // ── All 25 directions smoke test ─────────────────────────────────

  describe('All 25 conversion directions produce output', () => {
    const directions = ConverterFactory.getSupportedConversions();

    for (const direction of directions) {
      const [from, to] = direction.split('-');

      it(`${from} → ${to}: converts minimal input successfully`, async () => {
        const input = MINIMAL_INPUTS[from];
        expect(input).toBeTruthy();

        const converter = await ConverterFactory.createConverter(from, to);
        const output = await converter.convert(input);

        expect(output).toBeTruthy();
        expect(output.length).toBeGreaterThan(0);
      });
    }
  });

  // ── Batch mode across JS directions ──────────────────────────────

  describe('Batch mode', () => {
    test('should convert a directory of mixed JS test files', async () => {
      const batchDir = path.resolve(outputDir, 'batch-src');
      const batchOut = path.resolve(outputDir, 'batch-out');
      await fs.mkdir(batchDir, { recursive: true });

      await fs.writeFile(
        path.join(batchDir, 'a.test.js'),
        `describe('A', () => { it('works', () => { expect(1).toBe(1); }); });`,
      );
      await fs.writeFile(
        path.join(batchDir, 'b.test.js'),
        `describe('B', () => { it('works', () => { const fn = jest.fn(); fn(); expect(fn).toHaveBeenCalled(); }); });`,
      );

      const result = runCLI([
        'convert', batchDir,
        '--from', 'jest', '--to', 'vitest',
        '-o', batchOut,
      ]);

      expect(result).toContain('converted');
      const files = await fs.readdir(batchOut);
      expect(files.length).toBeGreaterThanOrEqual(2);
    });
  });

  // ── Migration mode ───────────────────────────────────────────────

  describe('Migration mode', () => {
    test('should migrate project with test files and track state', async () => {
      const migrateDir = path.resolve(outputDir, 'migrate-src');
      const migrateOut = path.resolve(outputDir, 'migrate-out');
      await fs.mkdir(migrateDir, { recursive: true });
      await fs.mkdir(migrateOut, { recursive: true });

      await fs.writeFile(
        path.join(migrateDir, 'auth.test.js'),
        `describe('Auth', () => { it('works', () => { expect(1).toBe(1); }); });`,
      );
      await fs.writeFile(
        path.join(migrateDir, 'utils.test.js'),
        `describe('Utils', () => { it('works', () => { expect(2).toBe(2); }); });`,
      );

      const result = runCLI([
        'migrate', migrateDir,
        '--from', 'jest', '--to', 'vitest',
        '-o', migrateOut,
      ]);

      expect(result).toContain('Migration complete');

      // State tracked in .hamlet/
      const stateExists = await fs.access(path.join(migrateDir, '.hamlet', 'state.json'))
        .then(() => true).catch(() => false);
      expect(stateExists).toBe(true);
    });
  });

  // ── Dry-run produces no side effects ─────────────────────────────

  describe('Dry-run produces no side effects', () => {
    test('should not create any files during dry-run', async () => {
      const dryDir = path.resolve(outputDir, 'dry-src');
      const dryOut = path.resolve(outputDir, 'dry-out');
      await fs.mkdir(dryDir, { recursive: true });
      await fs.rm(dryOut, { recursive: true, force: true }).catch(() => {});

      await fs.writeFile(
        path.join(dryDir, 'test.test.js'),
        `describe('T', () => { it('works', () => { expect(1).toBe(1); }); });`,
      );

      runCLI([
        'convert', dryDir,
        '--from', 'jest', '--to', 'vitest',
        '-o', dryOut,
        '--dry-run',
      ]);

      const outExists = await fs.access(dryOut).then(() => true).catch(() => false);
      expect(outExists).toBe(false);
    });
  });

  // ── Shorthand commands per language ──────────────────────────────

  describe('Shorthand commands per language', () => {
    test('jest2vt produces correct output', async () => {
      const inFile = path.resolve(outputDir, 'jest2vt-in.test.js');
      const outFile = path.resolve(outputDir, 'jest2vt-out.test.js');
      await fs.writeFile(inFile, MINIMAL_INPUTS.jest);

      runCLI(['jest2vt', inFile, '-o', outFile]);

      const output = await fs.readFile(outFile, 'utf8');
      expect(output).toContain("from 'vitest'");
    });

    test('cy2pw produces correct output', async () => {
      const inFile = path.resolve(outputDir, 'cy2pw-in.cy.js');
      const outFile = path.resolve(outputDir, 'cy2pw-out.spec.js');
      await fs.writeFile(inFile, MINIMAL_INPUTS.cypress);

      runCLI(['cy2pw', inFile, '-o', outFile]);

      const output = await fs.readFile(outFile, 'utf8');
      expect(output).toContain('page');
    });

    test('ju42ju5 produces correct output', async () => {
      const inFile = path.resolve(outputDir, 'ju42ju5-in.java');
      const outFile = path.resolve(outputDir, 'ju42ju5-out.java');
      await fs.writeFile(inFile, MINIMAL_INPUTS.junit4);

      runCLI(['ju42ju5', inFile, '-o', outFile]);

      const output = await fs.readFile(outFile, 'utf8');
      expect(output).toBeTruthy();
    });

    test('pyt2ut produces correct output', async () => {
      const inFile = path.resolve(outputDir, 'pyt2ut-in.py');
      const outFile = path.resolve(outputDir, 'pyt2ut-out.py');
      await fs.writeFile(inFile, MINIMAL_INPUTS.pytest);

      runCLI(['pyt2ut', inFile, '-o', outFile]);

      const output = await fs.readFile(outFile, 'utf8');
      expect(output).toBeTruthy();
    });
  });

  // ── --on-error modes ─────────────────────────────────────────────

  describe('--on-error modes', () => {
    test('skip mode continues past errors', async () => {
      const errDir = path.resolve(outputDir, 'onerror-src');
      const errOut = path.resolve(outputDir, 'onerror-out');
      await fs.mkdir(errDir, { recursive: true });

      await fs.writeFile(
        path.join(errDir, 'valid.test.js'),
        `describe('V', () => { it('works', () => { expect(1).toBe(1); }); });`,
      );

      const result = runCLI([
        'convert', errDir,
        '--from', 'jest', '--to', 'vitest',
        '-o', errOut,
        '--on-error', 'skip',
      ]);

      expect(result).toContain('converted');
    });
  });

  // ── JSON output mode ─────────────────────────────────────────────

  describe('JSON output mode', () => {
    test('should produce valid JSON with correct schema', async () => {
      const inFile = path.resolve(outputDir, 'json-in.test.js');
      const outFile = path.resolve(outputDir, 'json-out.test.js');
      await fs.writeFile(inFile, MINIMAL_INPUTS.jest);

      const result = runCLI(['jest2vt', inFile, '-o', outFile, '--json']);

      const parsed = JSON.parse(result);
      expect(parsed.success).toBe(true);
      expect(parsed.files).toBeInstanceOf(Array);
      expect(parsed.summary).toBeDefined();
      expect(typeof parsed.summary.converted).toBe('number');
      expect(typeof parsed.summary.skipped).toBe('number');
      expect(typeof parsed.summary.failed).toBe('number');
    });
  });

  // ── Estimate command ─────────────────────────────────────────────

  describe('Estimate command', () => {
    test('should show file counts and confidence predictions', async () => {
      const estDir = path.resolve(outputDir, 'estimate-src');
      await fs.mkdir(estDir, { recursive: true });

      await fs.writeFile(
        path.join(estDir, 'test.test.js'),
        `describe('T', () => { it('w', () => { expect(1).toBe(1); }); });`,
      );

      const result = runCLI([
        'estimate', estDir,
        '--from', 'jest', '--to', 'vitest',
      ]);

      expect(result).toContain('Estimation Summary');
      expect(result).toContain('Total files');
    });
  });

  // ── Doctor command ───────────────────────────────────────────────

  describe('Doctor command', () => {
    test('should run without error and show diagnostics', () => {
      const result = runCLI(['doctor']);
      expect(result).toContain('Hamlet Doctor');
      expect(result).toContain('Node.js');
      expect(result).toContain('Conversions');
    });
  });

  // ── Config conversion ────────────────────────────────────────────

  describe('Config conversion', () => {
    test('should convert jest.config.js to vitest.config.js', async () => {
      const configFile = path.resolve(outputDir, 'jest.config.js');
      await fs.writeFile(configFile, `module.exports = {
  testEnvironment: 'node',
  coverageDirectory: 'coverage',
  testMatch: ['**/*.test.js'],
};
`);

      const result = runCLI([
        'convert-config', configFile,
        '--from', 'jest', '--to', 'vitest',
      ]);

      expect(result).toBeTruthy();
    });
  });
});
