/**
 * CLI Contract Snapshot Tests (B6.1)
 *
 * Lightweight integration tests that protect the CLI UX from accidental regressions.
 * These assert stable substrings — never full output or timing-sensitive content.
 */
import { execFileSync, spawnSync } from 'child_process';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '../..');
const cliPath = path.resolve(rootDir, 'bin/hamlet.js');
const fixturesDir = path.resolve(__dirname, '../fixtures');
const outputDir = path.resolve(__dirname, '../output/contract');

function runCLI(args, options = {}) {
  return execFileSync('node', [cliPath, ...args], {
    encoding: 'utf8',
    ...options,
  });
}

function runCLISafe(args) {
  const result = spawnSync('node', [cliPath, ...args], {
    encoding: 'utf8',
    timeout: 60000,
    stdio: ['ignore', 'pipe', 'pipe'],
  });
  return {
    stdout: result.stdout || '',
    stderr: result.stderr || '',
    exitCode: result.status,
  };
}

describe('CLI Contract Snapshot Tests', () => {
  afterAll(async () => {
    await fs.rm(outputDir, { recursive: true, force: true }).catch(() => {});
  });

  // ── 1) hamlet --help ────────────────────────────────────────────
  describe('hamlet --help', () => {
    let helpOutput;

    beforeAll(() => {
      helpOutput = runCLI(['--help']);
    });

    test('contains grouped command headings', () => {
      expect(helpOutput).toContain('convert');
      expect(helpOutput).toContain('migrate');
      expect(helpOutput).toContain('doctor');
      expect(helpOutput).toContain('detect');
      expect(helpOutput).toContain('validate');
      expect(helpOutput).toContain('estimate');
    });

    test('contains shorthands footer reference', () => {
      expect(helpOutput).toContain('shorthands');
    });

    test('shows version flag', () => {
      expect(helpOutput).toContain('-V, --version');
    });
  });

  // ── 2) hamlet doctor ───────────────────────────────────────────
  describe('hamlet doctor on project root', () => {
    let result;

    beforeAll(() => {
      result = runCLISafe(['doctor', rootDir]);
    });

    test('exits with code 0', () => {
      expect(result.exitCode).toBe(0);
    });

    test('contains PASS/WARN/FAIL tokens', () => {
      const combined = result.stdout + result.stderr;
      // At minimum PASS should appear (Node.js check always passes in dev)
      expect(combined).toContain('PASS');
      // The token format should be present
      expect(combined).toMatch(/\[(PASS|WARN|FAIL)\]/);
    });

    test('contains checks summary', () => {
      const combined = result.stdout + result.stderr;
      expect(combined).toContain('checks:');
    });
  });

  // ── 3) hamlet convert --dry-run ────────────────────────────────
  describe('hamlet convert --dry-run', () => {
    const fixture = path.resolve(fixturesDir, 'sample.jest.js');
    let result;

    beforeAll(() => {
      result = runCLISafe([
        'convert',
        fixture,
        '--from',
        'jest',
        '--to',
        'vitest',
        '--dry-run',
      ]);
    });

    test('exits with code 0', () => {
      expect(result.exitCode).toBe(0);
    });

    test('contains dry-run marker', () => {
      expect(result.stdout).toMatch(/[Dd]ry run/);
    });

    test('does not create output files', async () => {
      // The default output path would be sample.test.js in the fixtures dir
      const possibleOutput = path.resolve(fixturesDir, 'sample.test.js');
      const exists = await fs
        .access(possibleOutput)
        .then(() => true)
        .catch(() => false);
      expect(exists).toBe(false);
    });
  });

  // ── 4) Exit code policy ──────────────────────────────────────────
  describe('exit code policy', () => {
    test('missing required argument exits with code 2', () => {
      // Convert with invalid framework → exit 2
      const result = runCLISafe([
        'convert',
        path.resolve(fixturesDir, 'sample.jest.js'),
        '--from',
        'notaframework',
        '--to',
        'vitest',
      ]);
      expect(result.exitCode).toBe(2);
    });

    test('runtime error exits with code 1', () => {
      // detect on nonexistent file → exit 1
      const result = runCLISafe([
        'detect',
        '/nonexistent/path/file.js',
      ]);
      expect(result.exitCode).toBe(1);
    });
  });
});
