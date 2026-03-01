import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import { generateReport } from '../../src/index.js';

describe('Public API: generateReport', () => {
  let tmpDir;

  beforeEach(async () => {
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-report-api-'));
  });

  afterEach(async () => {
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  it('should write to the caller-provided file path using caller-provided data', async () => {
    const outPath = path.join(tmpDir, 'custom-report.json');
    const resultPath = await generateReport(outPath, 'json', {
      summary: {
        totalFiles: 1,
      },
    });

    expect(resultPath).toBe(path.resolve(outPath));
    const raw = await fs.readFile(outPath, 'utf8');
    const report = JSON.parse(raw);
    expect(report.summary.totalFiles).toBe(1);
  });
});
