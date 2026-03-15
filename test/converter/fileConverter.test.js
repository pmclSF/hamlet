import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import {
  convertConfig,
  convertCypressToPlaywright,
  convertFile,
} from '../../src/converter/fileConverter.js';

describe('fileConverter', () => {
  let tmpDir;

  beforeEach(async () => {
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'terrain-file-convert-'));
  });

  afterEach(async () => {
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  it('should support validate=true without throwing runtime method errors', async () => {
    const source = path.join(tmpDir, 'sample.cy.js');
    const output = path.join(tmpDir, 'sample.spec.js');

    await fs.writeFile(
      source,
      `
describe('sample', () => {
  it('works', () => {
    cy.visit('/');
  });
});
`
    );

    const result = await convertFile(source, output, { validate: true });
    expect(result.success).toBe(true);
    expect(result.outputPath).toBe(output);
    expect(result.validationResults).toHaveProperty('summary');
  });

  it('should honor from/to options for non-Cypress directions', async () => {
    const source = path.join(tmpDir, 'sample.jest.js');
    const output = path.join(tmpDir, 'sample.vitest.js');

    await fs.writeFile(
      source,
      `
import { test, expect } from '@jest/globals';
test('works', () => {
  expect(1).toBe(1);
});
`
    );

    await convertFile(source, output, { from: 'jest', to: 'vitest' });
    const converted = await fs.readFile(output, 'utf8');

    expect(converted).toContain("from 'vitest'");
    expect(converted).not.toContain('@playwright/test');
  });

  it('should convert JS Cypress config files without JSON parse errors', async () => {
    const configPath = path.join(tmpDir, 'cypress.config.js');
    await fs.writeFile(
      configPath,
      "module.exports = { e2e: { baseUrl: 'https://example.com' }, retries: 1 };"
    );

    const converted = await convertConfig(configPath, {
      from: 'cypress',
      to: 'playwright',
    });
    expect(converted).toContain("defineConfig");
    expect(converted).toContain("baseURL");
  });

  it('should not inject Playwright projects into non-converted fallback output', async () => {
    const configPath = path.join(tmpDir, 'cypress.dynamic.config.js');
    await fs.writeFile(
      configPath,
      "module.exports = makeConfig(process.env.NODE_ENV === 'ci');"
    );

    const converted = await convertConfig(configPath, {
      from: 'cypress',
      to: 'playwright',
    });

    expect(converted).toContain('TERRAIN-TODO');
    expect(converted).not.toContain('projects: [');
    expect(converted).toContain("module.exports = makeConfig");
  });

  it('should not strip valid comparison expressions or trailing lines', async () => {
    const converted = await convertCypressToPlaywright(
      "it('x',()=>{const v = a < b > c; cy.visit('/')});\nconst z = 1;"
    );

    expect(converted).toContain('a < b > c');
    expect(converted).toContain('const z = 1;');
  });
});
