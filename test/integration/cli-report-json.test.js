import { spawnSync } from 'child_process';
import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';
import os from 'os';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '../..');
const cliPath = path.resolve(rootDir, 'bin/terrain.js');
const fixtureFile = path.resolve(
  __dirname,
  '../fixtures/convert/simple.test.js'
);

function assertConversionReportSchema(report) {
  expect(report).toBeDefined();
  expect(typeof report).toBe('object');

  expect(report.schemaVersion).toBe('1.0.0');
  expect(report.meta).toBeDefined();
  expect(typeof report.meta.terrainVersion).toBe('string');
  expect(typeof report.meta.nodeVersion).toBe('string');
  expect(typeof report.meta.startedAt).toBe('string');
  expect(typeof report.meta.finishedAt).toBe('string');
  expect(Number.isNaN(Date.parse(report.meta.startedAt))).toBe(false);
  expect(Number.isNaN(Date.parse(report.meta.finishedAt))).toBe(false);

  expect(report.plan).toBeDefined();
  expect(typeof report.plan.root).toBe('string');
  expect(typeof report.plan.outputDir).toBe('string');
  expect(report.plan.direction).toBeDefined();
  expect(typeof report.plan.direction.from).toBe('string');
  expect(typeof report.plan.direction.to).toBe('string');
  expect(typeof report.plan.direction.pipelineBacked).toBe('boolean');

  expect(report.results).toBeDefined();
  expect(typeof report.results.filesConverted).toBe('number');
  expect(typeof report.results.filesFailed).toBe('number');
  expect(typeof report.results.todosAdded).toBe('number');

  expect(Array.isArray(report.files)).toBe(true);
  for (const file of report.files) {
    expect(typeof file.inputPath).toBe('string');
    expect(
      file.outputPath === null || typeof file.outputPath === 'string'
    ).toBe(true);
    expect(['converted', 'failed', 'skipped']).toContain(file.status);
    expect(file.confidence === null || typeof file.confidence === 'number').toBe(
      true
    );
    expect(typeof file.todosAdded).toBe('number');
    expect(Array.isArray(file.warnings)).toBe(true);
    for (const warning of file.warnings) {
      expect(typeof warning).toBe('string');
      expect(warning.length).toBeGreaterThan(0);
    }
    if (file.status === 'converted') {
      expect(typeof file.outputPath).toBe('string');
    }
  }

  expect(
    report.results.filesConverted + report.results.filesFailed
  ).toBe(report.files.length);
}

describe('CLI --report-json on convert', () => {
  let tmpDir;

  beforeEach(async () => {
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'terrain-report-'));
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
    assertConversionReportSchema(report);
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

  it('should satisfy the full conversion report schema contract', async () => {
    const reportPath = path.join(tmpDir, 'report-schema.json');
    const outDir = path.join(tmpDir, 'out-schema');

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
    const report = JSON.parse(await fs.readFile(reportPath, 'utf8'));
    assertConversionReportSchema(report);
  });

  it('should serialize warnings as strings only (no null entries)', async () => {
    const warningFixture = path.join(tmpDir, 'warning.test.js');
    await fs.writeFile(
      warningFixture,
      `
jest.retryTimes(2);
test('works', () => {
  expect(1).toBe(1);
});
`
    );

    const reportPath = path.join(tmpDir, 'report-with-warning.json');
    const outDir = path.join(tmpDir, 'out-warning');
    const result = spawnSync(
      'node',
      [
        cliPath,
        'convert',
        warningFixture,
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
    const report = JSON.parse(await fs.readFile(reportPath, 'utf8'));
    expect(Array.isArray(report.files[0].warnings)).toBe(true);
    for (const warning of report.files[0].warnings) {
      expect(typeof warning).toBe('string');
      expect(warning.length).toBeGreaterThan(0);
    }
  });
});
