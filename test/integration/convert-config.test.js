import { execFileSync } from 'child_process';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const CLI = path.join(__dirname, '../../bin/hamlet.js');
const fixturesDir = path.join(__dirname, '../fixtures/configs');

function run(args) {
  return execFileSync('node', [CLI, ...args], {
    encoding: 'utf8',
    timeout: 15000,
  });
}

describe('convert-config CLI command', () => {
  let jestConfigPath;
  let cypressConfigPath;
  let wdioConfigPath;
  let mochaConfigPath;

  beforeAll(async () => {
    await fs.mkdir(fixturesDir, { recursive: true });

    jestConfigPath = path.join(fixturesDir, 'jest.config.js');
    await fs.writeFile(jestConfigPath, `module.exports = { testEnvironment: 'node', testTimeout: 30000 };`);

    cypressConfigPath = path.join(fixturesDir, 'cypress-simple.config.js');
    await fs.writeFile(cypressConfigPath, `module.exports = { baseUrl: 'http://localhost:3000', retries: 2 };`);

    wdioConfigPath = path.join(fixturesDir, 'wdio.conf.js');
    await fs.writeFile(wdioConfigPath, `exports.config = { baseUrl: 'http://localhost:3000', waitforTimeout: 10000 };`);

    mochaConfigPath = path.join(fixturesDir, '.mocharc.yml');
    await fs.writeFile(mochaConfigPath, `timeout: 5000\nspec: ./test/**/*.test.js`);
  });

  it('should convert Jest config to Vitest with --from and --to', () => {
    const result = run(['convert-config', jestConfigPath, '--from', 'jest', '--to', 'vitest']);

    expect(result).toContain('vitest/config');
    expect(result).toContain("environment: 'node'");
    expect(result).toContain('testTimeout: 30000');
  });

  it('should auto-detect framework from filename', () => {
    const result = run(['convert-config', jestConfigPath, '--to', 'vitest']);

    expect(result).toContain('vitest/config');
    expect(result).toContain("environment: 'node'");
  });

  it('should convert WDIO config to Playwright', () => {
    const result = run(['convert-config', wdioConfigPath, '--from', 'webdriverio', '--to', 'playwright']);

    expect(result).toContain('@playwright/test');
    expect(result).toContain('use.baseURL');
    expect(result).toContain('timeout: 10000');
  });

  it('should convert Mocha config to Jest', () => {
    const result = run(['convert-config', mochaConfigPath, '--from', 'mocha', '--to', 'jest']);

    expect(result).toContain('module.exports');
    expect(result).toContain('testTimeout: 5000');
  });

  it('should write output to file with --output', async () => {
    const outputPath = path.join(fixturesDir, 'vitest-output.config.ts');

    // Capture stderr (status messages) separately from stdout
    run(['convert-config', jestConfigPath, '--from', 'jest', '--to', 'vitest', '--output', outputPath]);

    const content = await fs.readFile(outputPath, 'utf8');
    expect(content).toContain('vitest/config');
    expect(content).toContain("environment: 'node'");

    // Cleanup
    await fs.unlink(outputPath);
  });

  it('should convert Cypress config to WebdriverIO', () => {
    const result = run(['convert-config', cypressConfigPath, '--from', 'cypress', '--to', 'webdriverio']);

    expect(result).toContain('exports.config');
    expect(result).toContain("baseUrl: 'http://localhost:3000'");
  });
});
