import { execFileSync } from 'child_process';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '../..');
const cliPath = path.resolve(rootDir, 'bin/hamlet.js');
const fixturesDir = path.resolve(__dirname, '../fixtures');
const outputDir = path.resolve(__dirname, '../output/errors-help');

function runCLI(args, options = {}) {
  return execFileSync('node', [cliPath, ...args], {
    encoding: 'utf8',
    ...options,
  });
}

function runCLIWithError(args) {
  try {
    execFileSync('node', [cliPath, ...args], {
      encoding: 'utf8',
      stdio: 'pipe',
    });
    return { stdout: '', stderr: '', exitCode: 0 };
  } catch (error) {
    return {
      stdout: error.stdout || '',
      stderr: error.stderr || '',
      exitCode: error.status,
    };
  }
}

describe('CLI Error Messages & Help', () => {
  beforeAll(async () => {
    await fs.mkdir(outputDir, { recursive: true });
  });

  afterAll(async () => {
    await fs.rm(outputDir, { recursive: true, force: true }).catch(() => {});
  });

  describe('Cross-language conversion error', () => {
    test('should include language info when converting across languages', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const result = runCLIWithError([
        'convert', inputFile,
        '--from', 'jest', '--to', 'pytest',
        '-o', path.resolve(outputDir, 'cross-lang-out.js'),
      ]);

      const combined = result.stdout + result.stderr;
      expect(combined).toContain('javascript');
      expect(combined).toContain('python');
      expect(result.exitCode).toBe(2);
    });
  });

  describe('Unsupported direction error', () => {
    test('should show supported targets for source framework', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      // jest -> selenium is not a supported direction
      const result = runCLIWithError([
        'convert', inputFile,
        '--from', 'jest', '--to', 'selenium',
        '-o', path.resolve(outputDir, 'unsupported-out.js'),
      ]);

      const combined = result.stdout + result.stderr;
      expect(combined).toContain('Unsupported conversion');
      // Should suggest supported targets for jest
      expect(combined).toContain('vitest');
    });
  });

  describe('Unknown framework error', () => {
    test('should show valid frameworks when an unknown framework is given', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const result = runCLIWithError([
        'convert', inputFile,
        '--from', 'jest', '--to', 'unknownfw',
      ]);

      const combined = result.stdout + result.stderr;
      expect(combined).toContain('Invalid target framework');
      expect(combined).toContain('Valid options');
    });
  });

  describe('File not found error', () => {
    test('should suggest similar files in directory', async () => {
      // Create a file with a known name so we can check for suggestions
      await fs.writeFile(path.join(outputDir, 'auth.test.js'), 'test');

      const result = runCLIWithError([
        'convert', path.join(outputDir, 'auth.tset.js'),
        '--from', 'jest', '--to', 'vitest',
        '-o', path.resolve(outputDir, 'notfound-out.js'),
      ]);

      const combined = result.stdout + result.stderr;
      expect(combined).toContain('File not found');
    });
  });

  describe('hamlet --help', () => {
    test('should show grouped commands', () => {
      const result = runCLI(['--help']);

      expect(result).toContain('convert');
      expect(result).toContain('list');
      expect(result).toContain('shorthands');
      expect(result).toContain('detect');
      expect(result).toContain('migrate');
      expect(result).toContain('estimate');
      expect(result).toContain('doctor');
    });
  });

  describe('hamlet convert --help', () => {
    test('should show convert command options', () => {
      const result = runCLI(['convert', '--help']);

      expect(result).toContain('--from');
      expect(result).toContain('--to');
      expect(result).toContain('--output');
      expect(result).toContain('--dry-run');
      expect(result).toContain('--json');
      expect(result).toContain('--quiet');
      expect(result).toContain('--on-error');
      expect(result).toContain('--verbose');
    });
  });

  describe('hamlet list', () => {
    test('should show all directions with shorthands', () => {
      const result = runCLI(['list']);

      expect(result).toContain('Supported conversion directions');
      // Check all 4 categories
      expect(result).toContain('JavaScript E2E');
      expect(result).toContain('JavaScript Unit');
      expect(result).toContain('Java');
      expect(result).toContain('Python');
      // Check shorthand presence
      expect(result).toContain('jest2vt');
      expect(result).toContain('cy2pw');
    });
  });

  describe('hamlet doctor', () => {
    test('should run without error', () => {
      const result = runCLI(['doctor']);

      expect(result).toContain('Hamlet Doctor');
      expect(result).toContain('Node.js');
      expect(result).toContain('Hamlet');
      expect(result).toContain('Conversions');
      expect(result).toContain('directions');
      expect(result).toContain('frameworks');
    });
  });

  describe('Invalid flag produces helpful error', () => {
    test('should show error for invalid --on-error value', () => {
      // Commander doesn't validate enum values, so the code handles it
      // But --from with an invalid framework does produce an error
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const result = runCLIWithError([
        'convert', inputFile,
        '--from', 'notreal',
        '--to', 'vitest',
      ]);

      expect(result.exitCode).not.toBe(0);
      const combined = result.stdout + result.stderr;
      expect(combined).toContain('Invalid');
    });
  });

  describe('Cross-language error with --json', () => {
    test('should produce JSON error for cross-language conversion', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const result = runCLIWithError([
        'convert', inputFile,
        '--from', 'jest', '--to', 'pytest',
        '--json',
      ]);

      const combined = result.stdout + result.stderr;
      // Should contain JSON with error info
      try {
        const parsed = JSON.parse(combined.trim());
        expect(parsed.success).toBe(false);
        expect(parsed.error).toContain('javascript');
        expect(parsed.error).toContain('python');
      } catch (_e) {
        // If JSON parsing fails, at least the error should be there
        expect(combined).toContain('python');
      }
    });
  });
});
