/**
 * CLI Summary Footer Tests
 *
 * Verifies that batch and single-file conversions produce a summary footer
 * with expected substrings. Uses real CLI invocations against tiny fixtures.
 */
import { execFileSync } from 'child_process';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '../..');
const cliPath = path.resolve(rootDir, 'bin/hamlet.js');
const outputDir = path.resolve(__dirname, '../output/report');
const batchFixtures = path.resolve(outputDir, 'fixtures');

function runCLI(args, options = {}) {
  return execFileSync('node', [cliPath, ...args], {
    encoding: 'utf8',
    ...options,
  });
}

describe('CLI Summary Footer', () => {
  beforeAll(async () => {
    await fs.mkdir(batchFixtures, { recursive: true });

    await fs.writeFile(
      path.join(batchFixtures, 'auth.test.js'),
      `describe('Auth', () => {
  it('should login', () => {
    expect(true).toBe(true);
  });
});
`
    );

    await fs.writeFile(
      path.join(batchFixtures, 'utils.test.js'),
      `describe('Utils', () => {
  it('should format', () => {
    expect('hello').toBe('hello');
  });
});
`
    );
  });

  afterAll(async () => {
    await fs.rm(outputDir, { recursive: true, force: true }).catch(() => {});
  });

  test('batch convert shows summary with converted count', () => {
    const outDir = path.resolve(outputDir, 'summary-batch');

    const result = runCLI([
      'convert',
      batchFixtures,
      '--from',
      'jest',
      '--to',
      'vitest',
      '-o',
      outDir,
    ]);

    expect(result).toContain('Summary:');
    expect(result).toMatch(/\d+ converted/);
  });

  test('single file convert shows completion message', () => {
    const fixture = path.resolve(rootDir, 'test/fixtures/sample.jest.js');
    const outDir = path.resolve(outputDir, 'summary-single');

    const result = runCLI([
      'convert',
      fixture,
      '--from',
      'jest',
      '--to',
      'vitest',
      '-o',
      outDir,
    ]);

    expect(result).toContain('Converted');
  });
});
