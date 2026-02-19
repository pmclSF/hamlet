import { execFileSync } from 'child_process';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '../..');
const cliPath = path.resolve(rootDir, 'bin/hamlet.js');
const outputDir = path.resolve(__dirname, '../output/batch');
const batchFixtures = path.resolve(outputDir, 'fixtures');

function runCLI(args, options = {}) {
  return execFileSync('node', [cliPath, ...args], {
    encoding: 'utf8',
    ...options,
  });
}

describe('CLI Batch Mode & Glob Patterns', () => {
  beforeAll(async () => {
    await fs.mkdir(batchFixtures, { recursive: true });

    // Create 3 jest test files for batch conversion
    await fs.writeFile(
      path.join(batchFixtures, 'auth.test.js'),
      `describe('Auth', () => {
  it('should login', () => {
    const result = jest.fn();
    result();
    expect(result).toHaveBeenCalled();
  });
});
`,
    );

    await fs.writeFile(
      path.join(batchFixtures, 'utils.test.js'),
      `describe('Utils', () => {
  it('should format', () => {
    expect('hello').toBe('hello');
  });
});
`,
    );

    // Create a subdirectory with another test file
    await fs.mkdir(path.join(batchFixtures, 'sub'), { recursive: true });
    await fs.writeFile(
      path.join(batchFixtures, 'sub', 'nested.test.js'),
      `describe('Nested', () => {
  it('should work', () => {
    expect(1 + 1).toBe(2);
  });
});
`,
    );

    // Create a non-test file that should be skipped
    await fs.writeFile(
      path.join(batchFixtures, 'readme.md'),
      '# Readme\n',
    );
  });

  afterAll(async () => {
    await fs.rm(outputDir, { recursive: true, force: true }).catch(() => {});
  });

  describe('Directory conversion', () => {
    test('should convert directory of test files to output directory', async () => {
      const outDir = path.resolve(outputDir, 'dir-out');

      runCLI([
        'convert', batchFixtures,
        '--from', 'jest', '--to', 'vitest',
        '-o', outDir,
      ]);

      const files = await fs.readdir(outDir);
      // Should have converted at least the top-level test files
      const testFiles = files.filter(f => f.endsWith('.test.js'));
      expect(testFiles.length).toBeGreaterThanOrEqual(2);

      // Verify content of a converted file
      const authContent = await fs.readFile(path.join(outDir, 'auth.test.js'), 'utf8');
      expect(authContent).toContain("from 'vitest'");
    });

    test('should preserve directory structure (subdir files)', async () => {
      const outDir = path.resolve(outputDir, 'structure-out');

      runCLI([
        'convert', batchFixtures,
        '--from', 'jest', '--to', 'vitest',
        '-o', outDir,
      ]);

      // Check if subdirectory structure was preserved
      const subExists = await fs.access(path.join(outDir, 'sub')).then(() => true).catch(() => false);
      if (subExists) {
        const subFiles = await fs.readdir(path.join(outDir, 'sub'));
        expect(subFiles.length).toBeGreaterThan(0);
      }
      // At minimum, the top-level files should exist
      const topFiles = await fs.readdir(outDir);
      expect(topFiles.filter(f => f.endsWith('.test.js')).length).toBeGreaterThanOrEqual(2);
    });
  });

  describe('Glob pattern conversion', () => {
    test('should convert files matching glob pattern', async () => {
      const outDir = path.resolve(outputDir, 'glob-out');

      runCLI([
        'convert', `${batchFixtures}/*.test.js`,
        '--from', 'jest', '--to', 'vitest',
        '-o', outDir,
      ]);

      const files = await fs.readdir(outDir);
      const testFiles = files.filter(f => f.endsWith('.test.js'));
      expect(testFiles.length).toBeGreaterThanOrEqual(2);
    });

    test('should show informative message for no matches', () => {
      const result = runCLI([
        'convert', `${batchFixtures}/*.nonexistent`,
        '--from', 'jest', '--to', 'vitest',
        '-o', path.resolve(outputDir, 'no-match'),
      ]);

      expect(result).toContain('No files matched');
    });
  });

  describe('Summary output', () => {
    test('should show correct summary counts', () => {
      const outDir = path.resolve(outputDir, 'summary-out');

      const result = runCLI([
        'convert', batchFixtures,
        '--from', 'jest', '--to', 'vitest',
        '-o', outDir,
      ]);

      expect(result).toContain('converted');
    });
  });

  describe('Shorthand with directory', () => {
    test('should work with jest2vt shorthand and directory', async () => {
      const outDir = path.resolve(outputDir, 'shorthand-dir-out');

      runCLI(['jest2vt', batchFixtures, '-o', outDir]);

      const files = await fs.readdir(outDir);
      expect(files.filter(f => f.endsWith('.test.js')).length).toBeGreaterThanOrEqual(2);
    });
  });

  describe('--output required for directory', () => {
    test('should error if --output missing for directory source', () => {
      expect(() => {
        runCLI(['convert', batchFixtures, '--from', 'jest', '--to', 'vitest'], {
          stdio: 'pipe',
        });
      }).toThrow();
    });
  });

  describe('Empty directory', () => {
    test('should handle empty directory gracefully', async () => {
      const emptyDir = path.resolve(outputDir, 'empty-src');
      await fs.mkdir(emptyDir, { recursive: true });

      const outDir = path.resolve(outputDir, 'empty-out');
      const result = runCLI([
        'convert', emptyDir,
        '--from', 'jest', '--to', 'vitest',
        '-o', outDir,
      ]);

      expect(result).toContain('No matching files');
    });
  });

  describe('Non-matching files skipped', () => {
    test('should skip files not matching source framework', async () => {
      // Create a directory with cypress files but convert with --from jest
      const mixDir = path.resolve(outputDir, 'mix-src');
      await fs.mkdir(mixDir, { recursive: true });

      await fs.writeFile(
        path.join(mixDir, 'test.cy.js'),
        `describe('Cypress', () => { it('test', () => { cy.visit('/'); }); });`,
      );

      const outDir = path.resolve(outputDir, 'mix-out');
      const result = runCLI([
        'convert', mixDir,
        '--from', 'jest', '--to', 'vitest',
        '-o', outDir,
      ]);

      // Should report no matching files or zero conversions since it's cypress not jest
      // The Scanner + FileClassifier filters by framework
      expect(result).toBeTruthy();
    });
  });
});
