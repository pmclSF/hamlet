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
});
