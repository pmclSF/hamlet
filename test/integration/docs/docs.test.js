/**
 * README example verification tests.
 *
 * Ensures that code examples in the documentation actually work,
 * preventing documentation drift.
 */

import { execFileSync } from 'child_process';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '../../..');
const cliPath = path.resolve(rootDir, 'bin/hamlet.js');
const outputDir = path.resolve(__dirname, '../../output/docs');

function runCLI(args, options = {}) {
  return execFileSync('node', [cliPath, ...args], {
    encoding: 'utf8',
    ...options,
  });
}

describe('Documentation Example Verification', () => {
  beforeAll(async () => {
    await fs.mkdir(outputDir, { recursive: true });
  });

  afterAll(async () => {
    await fs.rm(outputDir, { recursive: true, force: true }).catch(() => {});
  });

  test('README example: convert single file with shorthand', async () => {
    const inputFile = path.resolve(outputDir, 'auth.test.js');
    const outFile = path.resolve(outputDir, 'auth.converted.test.js');

    await fs.writeFile(inputFile, `describe('Auth', () => {
  const mockLogin = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should authenticate user', () => {
    mockLogin('admin', 'password');
    expect(mockLogin).toHaveBeenCalledWith('admin', 'password');
  });
});
`);

    runCLI(['jest2vt', inputFile, '-o', outFile]);

    const output = await fs.readFile(outFile, 'utf8');
    expect(output).toContain("from 'vitest'");
    expect(output).toContain('vi.fn');
  });

  test('README example: preview migration with estimate', async () => {
    const estDir = path.resolve(outputDir, 'estimate-example');
    await fs.mkdir(estDir, { recursive: true });

    await fs.writeFile(
      path.join(estDir, 'user.test.js'),
      `describe('User', () => { it('creates', () => { expect(true).toBe(true); }); });`,
    );

    const result = runCLI([
      'estimate', estDir,
      '--from', 'jest', '--to', 'vitest',
    ]);

    expect(result).toContain('Estimation Summary');
    expect(result).toContain('Total files');
  });

  test('README example: list supported conversions', () => {
    const result = runCLI(['list']);

    expect(result).toContain('JavaScript E2E');
    expect(result).toContain('JavaScript Unit');
    expect(result).toContain('Java');
    expect(result).toContain('Python');
  });

  test('README example: dry-run preview', async () => {
    const inputFile = path.resolve(outputDir, 'dryrun-example.test.js');
    await fs.writeFile(inputFile, `describe('App', () => {
  it('works', () => { expect(1).toBe(1); });
});
`);

    const result = runCLI([
      'convert', inputFile,
      '--from', 'jest', '--to', 'vitest',
      '--dry-run',
    ]);

    expect(result).toContain('Dry run');
    expect(result).toContain('Would convert');
  });

  test('README example: JSON output for CI integration', async () => {
    const inputFile = path.resolve(outputDir, 'ci-example.test.js');
    const outFile = path.resolve(outputDir, 'ci-example.converted.js');
    await fs.writeFile(inputFile, `describe('CI', () => {
  it('passes', () => { expect(true).toBe(true); });
});
`);

    const result = runCLI([
      'jest2vt', inputFile, '-o', outFile, '--json',
    ]);

    const parsed = JSON.parse(result);
    expect(parsed.success).toBe(true);
    expect(parsed.files).toBeInstanceOf(Array);
  });
});
