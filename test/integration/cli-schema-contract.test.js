/**
 * CLI JSON Schema Contract Tests
 *
 * Validates that machine-readable outputs contain required keys with correct types.
 * Does NOT test exact values — only structural guarantees.
 */
import { execFileSync } from 'child_process';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '../..');
const cliPath = path.resolve(rootDir, 'bin/hamlet.js');

function runCLI(args, options = {}) {
  return execFileSync('node', [cliPath, ...args], {
    encoding: 'utf8',
    ...options,
  });
}

describe('CLI Schema Contract Tests', () => {
  describe('doctor --json', () => {
    let parsed;

    beforeAll(() => {
      const output = runCLI(['doctor', rootDir, '--json']);
      parsed = JSON.parse(output);
    });

    test('has required top-level key: checks (array)', () => {
      expect(Array.isArray(parsed.checks)).toBe(true);
    });

    test('has required top-level key: summary (object)', () => {
      expect(typeof parsed.summary).toBe('object');
      expect(parsed.summary).not.toBeNull();
    });

    test('summary has required fields with correct types', () => {
      expect(typeof parsed.summary.pass).toBe('number');
      expect(typeof parsed.summary.warn).toBe('number');
      expect(typeof parsed.summary.fail).toBe('number');
      expect(typeof parsed.summary.total).toBe('number');
    });

    test('each check has required fields', () => {
      for (const check of parsed.checks) {
        expect(typeof check.id).toBe('string');
        expect(typeof check.label).toBe('string');
        expect(['PASS', 'WARN', 'FAIL']).toContain(check.status);
        expect(typeof check.detail).toBe('string');
      }
    });
  });
});
