import { spawnSync } from 'child_process';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '../..');
const cliPath = path.resolve(rootDir, 'bin/hamlet.js');
const outputDir = path.resolve(__dirname, '../output/doctor');

/**
 * Run the CLI and return stdout, stderr, and exitCode without throwing.
 */
function runDoctor(args, options = {}) {
  const result = spawnSync('node', [cliPath, 'doctor', ...args], {
    encoding: 'utf8',
    timeout: 60000,
    stdio: ['ignore', 'pipe', 'pipe'],
    ...options,
  });
  return {
    stdout: result.stdout || '',
    stderr: result.stderr || '',
    exitCode: result.status,
  };
}

describe('CLI doctor command', () => {
  const withTestsDir = path.join(outputDir, 'with-tests');
  const minimalDir = path.join(outputDir, 'minimal');
  const withTsDir = path.join(outputDir, 'with-ts');
  const emptyDir = path.join(outputDir, 'empty');

  beforeAll(async () => {
    // Clean up and recreate fixture directories
    await fs.rm(outputDir, { recursive: true, force: true });

    await fs.mkdir(withTestsDir, { recursive: true });
    await fs.mkdir(minimalDir, { recursive: true });
    await fs.mkdir(withTsDir, { recursive: true });
    await fs.mkdir(emptyDir, { recursive: true });

    // with-tests: package.json with jest ^29, plus test files
    await fs.writeFile(
      path.join(withTestsDir, 'package.json'),
      JSON.stringify(
        {
          name: 'doctor-fixture-with-tests',
          devDependencies: { jest: '^29.7.0' },
        },
        null,
        2,
      ),
    );
    await fs.writeFile(
      path.join(withTestsDir, 'app.test.js'),
      `describe('app', () => { it('works', () => { expect(1).toBe(1); }); });`,
    );
    await fs.writeFile(
      path.join(withTestsDir, 'utils.test.js'),
      `describe('utils', () => { it('works', () => { expect(true).toBe(true); }); });`,
    );

    // minimal: package.json only, no deps, no test files
    await fs.writeFile(
      path.join(minimalDir, 'package.json'),
      JSON.stringify({ name: 'doctor-fixture-minimal' }, null, 2),
    );

    // with-ts: package.json + tsconfig.json + test file
    await fs.writeFile(
      path.join(withTsDir, 'package.json'),
      JSON.stringify(
        {
          name: 'doctor-fixture-ts',
          devDependencies: { vitest: '^1.0.0' },
        },
        null,
        2,
      ),
    );
    await fs.writeFile(
      path.join(withTsDir, 'tsconfig.json'),
      JSON.stringify({ compilerOptions: { strict: true } }, null, 2),
    );
    await fs.writeFile(
      path.join(withTsDir, 'example.test.js'),
      `describe('ts project', () => { it('runs', () => { expect(1).toBe(1); }); });`,
    );

    // empty: nothing inside
  });

  afterAll(async () => {
    await fs.rm(outputDir, { recursive: true, force: true });
  });

  // ── Default path ─────────────────────────────────────────────────

  it('exits 0 for default path (project root) and shows PASS', () => {
    const { stdout, exitCode } = runDoctor([], { cwd: rootDir });
    expect(exitCode).toBe(0);
    expect(stdout).toContain('PASS');
  });

  // ── Valid project with tests ─────────────────────────────────────

  it('exits 0 for a valid project with test files', () => {
    const { stdout, exitCode } = runDoctor([withTestsDir]);
    expect(exitCode).toBe(0);
    expect(stdout).toContain('PASS');
    expect(stdout).toContain('test file');
  });

  it('shows test file count in output', () => {
    const { stdout } = runDoctor([withTestsDir]);
    expect(stdout).toMatch(/2 test files? found/);
  });

  // ── Summary line ─────────────────────────────────────────────────

  it('displays a summary line with check counts', () => {
    const { stdout } = runDoctor([withTestsDir]);
    expect(stdout).toMatch(/\d+ checks:/);
    expect(stdout).toMatch(/\d+ passed/);
  });

  // ── Missing tests ────────────────────────────────────────────────

  it('exits 0 (WARN, not FAIL) when no test files found', () => {
    const { stdout, exitCode } = runDoctor([minimalDir]);
    expect(exitCode).toBe(0);
    expect(stdout).toContain('WARN');
    expect(stdout).toContain('No test files');
  });

  // ── Non-existent path ────────────────────────────────────────────

  it('exits 1 for a non-existent path', () => {
    const { stdout, exitCode } = runDoctor(['/tmp/does-not-exist-hamlet']);
    expect(exitCode).toBe(1);
    expect(stdout).toContain('FAIL');
    expect(stdout).toContain('does not exist');
  });

  // ── File path (not a directory) ──────────────────────────────────

  it('exits 1 when given a file instead of a directory', () => {
    const filePath = path.join(withTestsDir, 'package.json');
    const { stdout, exitCode } = runDoctor([filePath]);
    expect(exitCode).toBe(1);
    expect(stdout).toContain('FAIL');
    expect(stdout).toContain('not a directory');
  });

  // ── --json output ────────────────────────────────────────────────

  it('produces valid JSON with checks and summary when --json is used', () => {
    const { stdout, exitCode } = runDoctor(['--json', withTestsDir]);
    expect(exitCode).toBe(0);
    const parsed = JSON.parse(stdout);
    expect(Array.isArray(parsed.checks)).toBe(true);
    expect(parsed.summary).toBeDefined();
    expect(typeof parsed.summary.pass).toBe('number');
    expect(typeof parsed.summary.total).toBe('number');
  });

  it('includes id, label, status, detail on each JSON check', () => {
    const { stdout } = runDoctor(['--json', withTestsDir]);
    const parsed = JSON.parse(stdout);
    for (const check of parsed.checks) {
      expect(check).toHaveProperty('id');
      expect(check).toHaveProperty('label');
      expect(check).toHaveProperty('status');
      expect(check).toHaveProperty('detail');
    }
  });

  it('returns exit 1 and summary.fail > 0 for FAIL in JSON mode', () => {
    const { stdout, exitCode } = runDoctor([
      '--json',
      '/tmp/does-not-exist-hamlet',
    ]);
    expect(exitCode).toBe(1);
    const parsed = JSON.parse(stdout);
    expect(parsed.summary.fail).toBeGreaterThan(0);
  });

  it('includes remediation on FAIL checks in JSON mode', () => {
    const { stdout } = runDoctor([
      '--json',
      '/tmp/does-not-exist-hamlet',
    ]);
    const parsed = JSON.parse(stdout);
    const failCheck = parsed.checks.find((c) => c.status === 'FAIL');
    expect(failCheck).toBeDefined();
    expect(failCheck.remediation).toBeDefined();
    expect(typeof failCheck.remediation).toBe('string');
  });

  // ── --verbose ────────────────────────────────────────────────────

  it('shows extra detail with --verbose', () => {
    const { stdout } = runDoctor(['--verbose', withTestsDir]);
    expect(stdout).toContain('Scanned');
  });

  it('includes verbose field in --json --verbose output', () => {
    const { stdout } = runDoctor(['--json', '--verbose', withTestsDir]);
    const parsed = JSON.parse(stdout);
    const testFilesCheck = parsed.checks.find((c) => c.id === 'test-files');
    expect(testFilesCheck).toBeDefined();
    expect(testFilesCheck.verbose).toBeDefined();
    expect(testFilesCheck.verbose).toContain('Scanned');
  });

  // ── node-version check ───────────────────────────────────────────

  it('always includes node-version check as PASS', () => {
    const { stdout } = runDoctor(['--json', withTestsDir]);
    const parsed = JSON.parse(stdout);
    const nodeCheck = parsed.checks.find((c) => c.id === 'node-version');
    expect(nodeCheck).toBeDefined();
    expect(nodeCheck.status).toBe('PASS');
  });

  // ── TypeScript detection ─────────────────────────────────────────

  it('detects tsconfig.json when present', () => {
    const { stdout } = runDoctor(['--json', withTsDir]);
    const parsed = JSON.parse(stdout);
    const tsCheck = parsed.checks.find((c) => c.id === 'typescript');
    expect(tsCheck).toBeDefined();
    expect(tsCheck.detail).toContain('tsconfig.json');
  });

  // ── Jest ESM warning ─────────────────────────────────────────────

  it('shows jest-esm warning for jest 29 projects', () => {
    const { stdout } = runDoctor(['--json', withTestsDir]);
    const parsed = JSON.parse(stdout);
    const jestCheck = parsed.checks.find((c) => c.id === 'jest-esm');
    expect(jestCheck).toBeDefined();
    expect(jestCheck.status).toBe('WARN');
  });

  it('omits jest-esm check for non-jest projects', () => {
    const { stdout } = runDoctor(['--json', minimalDir]);
    const parsed = JSON.parse(stdout);
    const jestCheck = parsed.checks.find((c) => c.id === 'jest-esm');
    expect(jestCheck).toBeUndefined();
  });
});
