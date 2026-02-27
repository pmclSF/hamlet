import { spawnSync } from 'child_process';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';
import os from 'os';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '../..');
const cliPath = path.resolve(rootDir, 'bin/hamlet.js');
const fixtureFile = path.resolve(
  __dirname,
  '../fixtures/convert/simple.test.js'
);

describe('CLI --report-json on convert', () => {
  let tmpDir;

  beforeEach(async () => {
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-report-'));
  });

  afterEach(async () => {
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  it('should write valid JSON report', async () => {
    const reportPath = path.join(tmpDir, 'report.json');
    const outDir = path.join(tmpDir, 'out');

    const result = spawnSync(
      'node',
      [
        cliPath,
        'convert',
        fixtureFile,
        '--from',
        'jest',
        '--to',
        'vitest',
        '-o',
        outDir,
        '--report-json',
        reportPath,
      ],
      { encoding: 'utf8', timeout: 30000, stdio: ['ignore', 'pipe', 'pipe'] }
    );

    expect(result.status).toBe(0);

    const raw = await fs.readFile(reportPath, 'utf8');
    const report = JSON.parse(raw);
    expect(report).toBeDefined();
  });

  it('should have required top-level keys', async () => {
    const reportPath = path.join(tmpDir, 'report.json');
    const outDir = path.join(tmpDir, 'out');

    spawnSync(
      'node',
      [
        cliPath,
        'convert',
        fixtureFile,
        '--from',
        'jest',
        '--to',
        'vitest',
        '-o',
        outDir,
        '--report-json',
        reportPath,
      ],
      { encoding: 'utf8', timeout: 30000, stdio: ['ignore', 'pipe', 'pipe'] }
    );

    const report = JSON.parse(await fs.readFile(reportPath, 'utf8'));
    expect(report).toHaveProperty('schemaVersion', '1.0.0');
    expect(report).toHaveProperty('meta');
    expect(report).toHaveProperty('plan');
    expect(report).toHaveProperty('results');
    expect(report).toHaveProperty('files');
  });

  it('should have correct direction and pipelineBacked', async () => {
    const reportPath = path.join(tmpDir, 'report.json');
    const outDir = path.join(tmpDir, 'out');

    spawnSync(
      'node',
      [
        cliPath,
        'convert',
        fixtureFile,
        '--from',
        'jest',
        '--to',
        'vitest',
        '-o',
        outDir,
        '--report-json',
        reportPath,
      ],
      { encoding: 'utf8', timeout: 30000, stdio: ['ignore', 'pipe', 'pipe'] }
    );

    const report = JSON.parse(await fs.readFile(reportPath, 'utf8'));
    expect(report.plan.direction.from).toBe('jest');
    expect(report.plan.direction.to).toBe('vitest');
    expect(typeof report.plan.direction.pipelineBacked).toBe('boolean');
  });

  it('should report correct conversion results', async () => {
    const reportPath = path.join(tmpDir, 'report.json');
    const outDir = path.join(tmpDir, 'out');

    spawnSync(
      'node',
      [
        cliPath,
        'convert',
        fixtureFile,
        '--from',
        'jest',
        '--to',
        'vitest',
        '-o',
        outDir,
        '--report-json',
        reportPath,
      ],
      { encoding: 'utf8', timeout: 30000, stdio: ['ignore', 'pipe', 'pipe'] }
    );

    const report = JSON.parse(await fs.readFile(reportPath, 'utf8'));
    expect(report.results.filesConverted).toBe(1);
    expect(report.files.length).toBe(1);
    expect(report.files[0].status).toBe('converted');
    expect(typeof report.files[0].todosAdded).toBe('number');
    expect(report.files[0]).toHaveProperty('confidence');
  });
});
