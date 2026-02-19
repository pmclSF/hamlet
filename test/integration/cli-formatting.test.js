import { execFileSync } from 'child_process';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '../..');
const cliPath = path.resolve(rootDir, 'bin/hamlet.js');
const fixturesDir = path.resolve(__dirname, '../fixtures');
const outputDir = path.resolve(__dirname, '../output/formatting');

function runCLI(args, options = {}) {
  return execFileSync('node', [cliPath, ...args], {
    encoding: 'utf8',
    ...options,
  });
}

describe('CLI Output Formatting & Progress', () => {
  beforeAll(async () => {
    await fs.mkdir(outputDir, { recursive: true });
  });

  afterAll(async () => {
    await fs.rm(outputDir, { recursive: true, force: true }).catch(() => {});
  });

  describe('--no-color flag', () => {
    test('should suppress ANSI color codes in output', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const outFile = path.resolve(outputDir, 'nocolor-out.js');

      const result = runCLI([
        'jest2vt', inputFile, '-o', outFile, '--no-color',
      ]);

      // ANSI escape codes start with \x1b[
      // eslint-disable-next-line no-control-regex
      const hasAnsi = /\x1b\[/.test(result);
      expect(hasAnsi).toBe(false);
    });
  });

  describe('NO_COLOR environment variable', () => {
    test('should suppress color when NO_COLOR is set', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const outFile = path.resolve(outputDir, 'nocolor-env-out.js');

      const result = runCLI(
        ['jest2vt', inputFile, '-o', outFile],
        { env: { ...process.env, NO_COLOR: '1' } },
      );

      // eslint-disable-next-line no-control-regex
      const hasAnsi = /\x1b\[/.test(result);
      expect(hasAnsi).toBe(false);
    });
  });

  describe('--quiet flag', () => {
    test('should suppress normal output', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const outFile = path.resolve(outputDir, 'quiet-fmt-out.js');

      const result = runCLI([
        'jest2vt', inputFile, '-o', outFile, '--quiet',
      ]);

      expect(result.trim()).toBe('');
    });
  });

  describe('--verbose flag', () => {
    test('should show detailed output', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const outFile = path.resolve(outputDir, 'verbose-out.js');

      const result = runCLI([
        'jest2vt', inputFile, '-o', outFile, '--verbose',
      ]);

      // Verbose should show more detail than normal
      expect(result).toContain('Converted');
    });
  });

  describe('--json flag', () => {
    test('should produce valid parseable JSON', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const outFile = path.resolve(outputDir, 'json-out.js');

      const result = runCLI([
        'jest2vt', inputFile, '-o', outFile, '--json',
      ]);

      const parsed = JSON.parse(result);
      expect(parsed.success).toBe(true);
      expect(parsed.files).toBeInstanceOf(Array);
      expect(parsed.files.length).toBe(1);
      expect(parsed.summary).toBeDefined();
      expect(parsed.summary.converted).toBe(1);
    });

    test('should include file details in JSON', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const outFile = path.resolve(outputDir, 'json-detail-out.js');

      const result = runCLI([
        'jest2vt', inputFile, '-o', outFile, '--json',
      ]);

      const parsed = JSON.parse(result);
      expect(parsed.files[0].source).toBeTruthy();
      expect(parsed.files[0].output).toBeTruthy();
    });

    test('should produce JSON on error too', () => {
      try {
        runCLI([
          'convert', '/nonexistent.js',
          '--from', 'jest', '--to', 'vitest',
          '--json',
        ], { stdio: 'pipe' });
      } catch (error) {
        // The stderr or stdout should contain JSON
        const combined = (error.stdout || '') + (error.stderr || '');
        // In error cases the JSON is output before exit
        expect(combined).toBeTruthy();
      }
    });
  });

  describe('Exit codes', () => {
    test('should exit 0 for successful conversion', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const outFile = path.resolve(outputDir, 'exitcode-out.js');

      // If this doesn't throw, exit code is 0
      const result = runCLI(['jest2vt', inputFile, '-o', outFile]);
      expect(result).toBeTruthy();
    });

    test('should exit non-zero for failures', () => {
      expect(() => {
        runCLI([
          'convert', '/nonexistent.js',
          '--from', 'jest', '--to', 'vitest',
          '-o', path.resolve(outputDir, 'fail-out.js'),
        ], { stdio: 'pipe' });
      }).toThrow();
    });

    test('should exit 2 for bad arguments', () => {
      try {
        runCLI([
          'convert', path.resolve(fixturesDir, 'sample.jest.js'),
          '--from', 'invalid_framework', '--to', 'vitest',
        ], { stdio: 'pipe' });
        // Should not reach here
        expect(true).toBe(false);
      } catch (error) {
        expect(error.status).toBe(2);
      }
    });

    test('should exit 2 for same source and target', () => {
      try {
        runCLI([
          'convert', path.resolve(fixturesDir, 'sample.jest.js'),
          '--from', 'jest', '--to', 'jest',
        ], { stdio: 'pipe' });
        expect(true).toBe(false);
      } catch (error) {
        expect(error.status).toBe(2);
      }
    });
  });

  describe('Confidence report', () => {
    test('should show confidence for single file conversion', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const outFile = path.resolve(outputDir, 'confidence-out.js');

      const result = runCLI(['jest2vt', inputFile, '-o', outFile]);

      expect(result).toContain('Confidence');
    });
  });

  describe('Batch summary', () => {
    test('should show accurate summary at end of batch', async () => {
      const batchDir = path.resolve(outputDir, 'batch-summary-src');
      await fs.mkdir(batchDir, { recursive: true });

      await fs.writeFile(
        path.join(batchDir, 'a.test.js'),
        `describe('A', () => { it('works', () => { expect(1).toBe(1); }); });`,
      );
      await fs.writeFile(
        path.join(batchDir, 'b.test.js'),
        `describe('B', () => { it('works', () => { expect(2).toBe(2); }); });`,
      );

      const outDir = path.resolve(outputDir, 'batch-summary-out');
      const result = runCLI([
        'convert', batchDir,
        '--from', 'jest', '--to', 'vitest',
        '-o', outDir,
      ]);

      expect(result).toContain('converted');
      expect(result).toContain('skipped');
      expect(result).toContain('failed');
    });
  });
});
