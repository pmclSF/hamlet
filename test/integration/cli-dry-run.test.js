import { execFileSync } from 'child_process';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '../..');
const cliPath = path.resolve(rootDir, 'bin/hamlet.js');
const fixturesDir = path.resolve(__dirname, '../fixtures');
const outputDir = path.resolve(__dirname, '../output/dry-run');

function runCLI(args, options = {}) {
  return execFileSync('node', [cliPath, ...args], {
    encoding: 'utf8',
    ...options,
  });
}

describe('CLI Dry-Run Mode', () => {
  let dryRunFixtures;

  beforeAll(async () => {
    await fs.mkdir(outputDir, { recursive: true });
    dryRunFixtures = path.resolve(outputDir, 'fixtures');
    await fs.mkdir(dryRunFixtures, { recursive: true });

    await fs.writeFile(
      path.join(dryRunFixtures, 'auth.test.js'),
      `describe('Auth', () => {
  it('should login', () => {
    const fn = jest.fn();
    fn();
    expect(fn).toHaveBeenCalled();
  });
});
`,
    );

    await fs.writeFile(
      path.join(dryRunFixtures, 'utils.test.js'),
      `describe('Utils', () => {
  it('should work', () => {
    expect(1 + 1).toBe(2);
  });
});
`,
    );
  });

  afterAll(async () => {
    await fs.rm(outputDir, { recursive: true, force: true }).catch(() => {});
  });

  describe('Single file dry-run', () => {
    test('should show confidence without writing files', async () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const outFile = path.resolve(outputDir, 'dryrun-single.js');

      await fs.rm(outFile, { force: true }).catch(() => {});

      const result = runCLI([
        'convert', inputFile,
        '--from', 'jest', '--to', 'vitest',
        '-o', outFile,
        '--dry-run',
      ]);

      expect(result).toContain('Dry run');
      expect(result).toContain('Would convert');

      // No file should be created
      const exists = await fs.access(outFile).then(() => true).catch(() => false);
      expect(exists).toBe(false);
    });

    test('should include confidence level in output', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');

      const result = runCLI([
        'convert', inputFile,
        '--from', 'jest', '--to', 'vitest',
        '--dry-run',
      ]);

      expect(result).toContain('Confidence');
    });
  });

  describe('Batch dry-run', () => {
    test('should show file counts without writing', async () => {
      const outDir = path.resolve(outputDir, 'batch-dryrun-out');
      await fs.rm(outDir, { recursive: true, force: true }).catch(() => {});

      const result = runCLI([
        'convert', dryRunFixtures,
        '--from', 'jest', '--to', 'vitest',
        '-o', outDir,
        '--dry-run',
      ]);

      expect(result).toContain('Dry run');
      expect(result).toContain('Files found');
      expect(result).toContain('Would convert');

      // No output directory should be created
      const exists = await fs.access(outDir).then(() => true).catch(() => false);
      expect(exists).toBe(false);
    });

    test('should show confidence distribution', () => {
      const result = runCLI([
        'convert', dryRunFixtures,
        '--from', 'jest', '--to', 'vitest',
        '-o', path.resolve(outputDir, 'conf-dist-out'),
        '--dry-run',
      ]);

      expect(result).toContain('Confidence distribution');
      expect(result).toMatch(/High/i);
    });
  });

  describe('Migrate dry-run', () => {
    test('should show estimation without creating .hamlet/', async () => {
      const migrateDir = path.resolve(outputDir, 'migrate-dryrun');
      await fs.mkdir(migrateDir, { recursive: true });
      await fs.writeFile(
        path.join(migrateDir, 'test.test.js'),
        `describe('t', () => { it('w', () => { expect(1).toBe(1); }); });`,
      );

      const result = runCLI([
        'migrate', migrateDir,
        '--from', 'jest', '--to', 'vitest',
        '--dry-run',
      ]);

      expect(result).toContain('Dry run');
      expect(result).toContain('Estimation Summary');

      // .hamlet/ should NOT exist
      const hamletExists = await fs.access(path.join(migrateDir, '.hamlet'))
        .then(() => true).catch(() => false);
      expect(hamletExists).toBe(false);
    });
  });

  describe('No files created on disk during dry-run', () => {
    test('should not create any files', async () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const outFile = path.resolve(outputDir, 'absolutely-should-not-exist.js');

      await fs.rm(outFile, { force: true }).catch(() => {});

      runCLI(['jest2vt', inputFile, '--dry-run']);

      const exists = await fs.access(outFile).then(() => true).catch(() => false);
      expect(exists).toBe(false);
    });
  });

  describe('Dry-run with --quiet', () => {
    test('should produce no output', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');

      const result = runCLI([
        'jest2vt', inputFile, '--dry-run', '--quiet',
      ]);

      expect(result.trim()).toBe('');
    });
  });

  describe('Dry-run with shorthand', () => {
    test('should work with shorthand commands', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');

      const result = runCLI(['jest2vt', inputFile, '--dry-run']);

      expect(result).toContain('Dry run');
    });
  });

  describe('Dry-run with --json', () => {
    test('should produce valid JSON output', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');

      const result = runCLI(['jest2vt', inputFile, '--dry-run', '--json']);

      const parsed = JSON.parse(result);
      expect(parsed.success).toBe(true);
      expect(parsed.dryRun).toBe(true);
    });
  });

  describe('convert-config --dry-run', () => {
    let configFixturesDir;
    let jestConfigPath;

    beforeAll(async () => {
      configFixturesDir = path.resolve(outputDir, 'config-fixtures');
      await fs.mkdir(configFixturesDir, { recursive: true });
      jestConfigPath = path.join(configFixturesDir, 'jest.config.js');
      await fs.writeFile(
        jestConfigPath,
        `module.exports = { testEnvironment: 'node', testTimeout: 30000 };`
      );
    });

    test('should show preview without writing output file', async () => {
      const outFile = path.resolve(
        outputDir,
        'config-dryrun-output.config.ts'
      );
      await fs.rm(outFile, { force: true }).catch(() => {});

      const result = runCLI([
        'convert-config',
        jestConfigPath,
        '--from',
        'jest',
        '--to',
        'vitest',
        '--output',
        outFile,
        '--dry-run',
      ]);

      expect(result).toContain('Dry run');
      expect(result).toContain('jest');
      expect(result).toContain('vitest');

      const exists = await fs
        .access(outFile)
        .then(() => true)
        .catch(() => false);
      expect(exists).toBe(false);
    });

    test('should show stdout as output when no --output given', () => {
      const result = runCLI([
        'convert-config',
        jestConfigPath,
        '--from',
        'jest',
        '--to',
        'vitest',
        '--dry-run',
      ]);

      expect(result).toContain('Dry run');
      expect(result).toContain('(stdout)');
    });
  });

  describe('convert --plan', () => {
    test('should show file mapping for single file', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');

      const result = runCLI([
        'convert',
        inputFile,
        '--from',
        'jest',
        '--to',
        'vitest',
        '--plan',
      ]);

      expect(result).toContain('Conversion Plan');
      expect(result).toContain('\u2192');
      expect(result).toContain('Direction: jest');
      expect(result).toContain('Confidence');
    });

    test('should show file mapping table for batch', () => {
      const result = runCLI([
        'convert',
        dryRunFixtures,
        '--from',
        'jest',
        '--to',
        'vitest',
        '-o',
        path.resolve(outputDir, 'plan-batch-out'),
        '--plan',
      ]);

      expect(result).toContain('Conversion Plan');
      expect(result).toContain('files to convert');
      expect(result).toContain('Input');
      expect(result).toContain('Output');
      expect(result).toContain('\u2192');
      expect(result).toContain('Confidence');
    });

    test('should produce valid JSON with plan structure', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');

      const result = runCLI([
        'convert',
        inputFile,
        '--from',
        'jest',
        '--to',
        'vitest',
        '--plan',
        '--json',
      ]);

      const parsed = JSON.parse(result);
      expect(parsed.plan).toBe(true);
      expect(parsed.direction).toEqual({
        from: 'jest',
        to: 'vitest',
      });
      expect(parsed.files).toBeInstanceOf(Array);
      expect(parsed.files.length).toBeGreaterThan(0);
      expect(parsed.files[0]).toHaveProperty('source');
      expect(parsed.files[0]).toHaveProperty('output');
      expect(parsed.files[0]).toHaveProperty('confidence');
      expect(parsed.summary).toHaveProperty('total');
      expect(parsed.summary).toHaveProperty('confidence');
      expect(parsed.warnings).toEqual([]);
    });

    test('should imply --dry-run (no files written)', async () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');
      const outFile = path.resolve(
        outputDir,
        'plan-implies-dryrun.test.js'
      );
      await fs.rm(outFile, { force: true }).catch(() => {});

      runCLI([
        'convert',
        inputFile,
        '--from',
        'jest',
        '--to',
        'vitest',
        '-o',
        outFile,
        '--plan',
      ]);

      const exists = await fs
        .access(outFile)
        .then(() => true)
        .catch(() => false);
      expect(exists).toBe(false);
    });

    test('should work with shorthand commands', () => {
      const inputFile = path.resolve(fixturesDir, 'sample.jest.js');

      const result = runCLI(['jest2vt', inputFile, '--plan']);

      expect(result).toContain('Conversion Plan');
      expect(result).toContain('Confidence');
    });
  });

  describe('migrate --plan', () => {
    test('should show file-by-file mapping', async () => {
      const migrateDir = path.resolve(outputDir, 'migrate-plan');
      await fs.mkdir(migrateDir, { recursive: true });
      await fs.writeFile(
        path.join(migrateDir, 'test.test.js'),
        `describe('t', () => { it('w', () => { expect(1).toBe(1); }); });`
      );

      const result = runCLI([
        'migrate',
        migrateDir,
        '--from',
        'jest',
        '--to',
        'vitest',
        '--plan',
      ]);

      expect(result).toContain('Migration Plan');
      expect(result).toContain('jest');
      expect(result).toContain('vitest');
      expect(result).toContain('Input');
      expect(result).toContain('Confidence');
      expect(result).toContain('Summary');

      // .hamlet/ should NOT exist
      const hamletExists = await fs
        .access(path.join(migrateDir, '.hamlet'))
        .then(() => true)
        .catch(() => false);
      expect(hamletExists).toBe(false);
    });
  });
});
