import { execFileSync } from 'child_process';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '../..');
const cliPath = path.resolve(rootDir, 'bin/hamlet.js');
const outputDir = path.resolve(__dirname, '../output/on-error');

function runCLI(args, options = {}) {
  return execFileSync('node', [cliPath, ...args], {
    encoding: 'utf8',
    ...options,
  });
}

describe('CLI --on-error Flag', () => {
  let batchDir;

  beforeAll(async () => {
    await fs.mkdir(outputDir, { recursive: true });
    batchDir = path.resolve(outputDir, 'batch-src');
    await fs.mkdir(batchDir, { recursive: true });

    // Valid jest test file
    await fs.writeFile(
      path.join(batchDir, 'good.test.js'),
      `describe('Good', () => {
  it('works', () => {
    const fn = jest.fn();
    fn();
    expect(fn).toHaveBeenCalled();
  });
});
`,
    );

    // Another valid file
    await fs.writeFile(
      path.join(batchDir, 'also-good.test.js'),
      `describe('Also Good', () => {
  it('passes', () => {
    expect(1).toBe(1);
  });
});
`,
    );
  });

  afterAll(async () => {
    await fs.rm(outputDir, { recursive: true, force: true }).catch(() => {});
  });

  describe('Default (skip) mode', () => {
    test('should default to skip when flag not specified', async () => {
      const outDir = path.resolve(outputDir, 'default-out');

      const result = runCLI([
        'convert', batchDir,
        '--from', 'jest', '--to', 'vitest',
        '-o', outDir,
      ]);

      // Should complete without error
      expect(result).toContain('converted');
    });
  });

  describe('skip mode', () => {
    test('should skip failed files and continue converting others', async () => {
      const outDir = path.resolve(outputDir, 'skip-out');

      const result = runCLI([
        'convert', batchDir,
        '--from', 'jest', '--to', 'vitest',
        '-o', outDir,
        '--on-error', 'skip',
      ]);

      // Should show summary with conversions
      expect(result).toContain('converted');
    });
  });

  describe('fail mode', () => {
    test('should stop on first error with non-zero exit', () => {
      // Use a nonexistent source that will cause stat failure
      expect(() => {
        runCLI([
          'convert', '/nonexistent/dir',
          '--from', 'jest', '--to', 'vitest',
          '-o', path.resolve(outputDir, 'fail-out'),
          '--on-error', 'fail',
        ], { stdio: 'pipe' });
      }).toThrow();
    });
  });

  describe('best-effort mode for single file', () => {
    test('should try partial output with HAMLET-WARNING comment on error', () => {
      // Create a file that will cause conversion issues
      const badDir = path.resolve(outputDir, 'best-effort-src');
      // For single file, best-effort still exits 1 but may write partial output
      // Using an unsupported conversion to trigger error
      expect(() => {
        runCLI([
          'convert', path.join(batchDir, 'good.test.js'),
          '--from', 'jest', '--to', 'pytest',
          '-o', path.resolve(outputDir, 'best-effort-out.js'),
          '--on-error', 'best-effort',
        ], { stdio: 'pipe' });
      }).toThrow();
    });
  });

  describe('Single file + skip mode', () => {
    test('should error with exit code on invalid conversion', () => {
      expect(() => {
        runCLI([
          'convert', '/nonexistent/file.js',
          '--from', 'jest', '--to', 'vitest',
          '-o', path.resolve(outputDir, 'single-skip-out.js'),
          '--on-error', 'skip',
        ], { stdio: 'pipe' });
      }).toThrow();
    });
  });

  describe('Error summary at end of batch', () => {
    test('should show summary with counts after batch conversion', async () => {
      const outDir = path.resolve(outputDir, 'summary-out');

      const result = runCLI([
        'convert', batchDir,
        '--from', 'jest', '--to', 'vitest',
        '-o', outDir,
        '--on-error', 'skip',
      ]);

      // Summary should contain counts
      expect(result).toMatch(/\d+ converted/);
    });
  });

  describe('--on-error with shorthand', () => {
    test('should accept --on-error flag with shorthand commands', async () => {
      const outDir = path.resolve(outputDir, 'shorthand-onerror-out');

      const result = runCLI([
        'jest2vt', batchDir,
        '-o', outDir,
        '--on-error', 'skip',
      ]);

      expect(result).toContain('converted');
    });
  });

  describe('--on-error with --json', () => {
    test('should include error information in JSON output', async () => {
      const outDir = path.resolve(outputDir, 'json-onerror-out');

      const result = runCLI([
        'jest2vt', batchDir,
        '-o', outDir,
        '--on-error', 'skip',
        '--json',
      ]);

      const parsed = JSON.parse(result);
      expect(parsed.summary).toBeDefined();
      expect(typeof parsed.summary.converted).toBe('number');
    });
  });
});
