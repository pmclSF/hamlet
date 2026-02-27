import fs from 'fs/promises';
import path from 'path';
import os from 'os';
import { generateHtmlReport } from '../../src/core/HtmlReportGenerator.js';

describe('HtmlReportGenerator', () => {
  let tmpDir;

  beforeEach(async () => {
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-html-'));
  });

  afterEach(async () => {
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  const sampleReport = {
    schemaVersion: 1,
    meta: {
      hamletVersion: '1.0.0',
      generatedAt: '2025-01-01T00:00:00.000Z',
      root: '/tmp/project',
    },
    summary: {
      fileCount: 3,
      testFileCount: 2,
      frameworksDetected: ['jest', 'mocha'],
      confidenceAvg: 85,
      directionsSupported: [
        { from: 'jest', to: 'vitest', pipelineBacked: true },
        { from: 'mocha', to: 'jest', pipelineBacked: false },
      ],
    },
    files: [
      {
        path: 'src/utils.js',
        type: 'source',
        framework: null,
        confidence: 0,
        candidates: [],
        warnings: [],
      },
      {
        path: 'test/auth.test.js',
        type: 'test',
        framework: 'jest',
        confidence: 92,
        candidates: [
          { framework: 'jest', score: 92 },
          { framework: 'jasmine', score: 30 },
        ],
        warnings: ['Mixed assertion styles detected'],
      },
      {
        path: 'test/math.test.js',
        type: 'test',
        framework: 'mocha',
        confidence: 78,
        candidates: [{ framework: 'mocha', score: 78 }],
        warnings: [],
      },
    ],
  };

  describe('generateHtmlReport', () => {
    it('should create index.html and report.json', async () => {
      const outDir = path.join(tmpDir, 'report');
      await generateHtmlReport(sampleReport, outDir);

      const htmlPath = path.join(outDir, 'index.html');
      const jsonPath = path.join(outDir, 'report.json');

      const htmlStat = await fs.stat(htmlPath);
      const jsonStat = await fs.stat(jsonPath);

      expect(htmlStat.isFile()).toBe(true);
      expect(jsonStat.isFile()).toBe(true);
    });

    it('should write valid JSON sidecar matching the input report', async () => {
      const outDir = path.join(tmpDir, 'report');
      await generateHtmlReport(sampleReport, outDir);

      const jsonContent = await fs.readFile(
        path.join(outDir, 'report.json'),
        'utf-8'
      );
      const parsed = JSON.parse(jsonContent);

      expect(parsed.schemaVersion).toBe(1);
      expect(parsed.meta.hamletVersion).toBe('1.0.0');
      expect(parsed.summary.fileCount).toBe(3);
      expect(parsed.files).toHaveLength(3);
    });

    it('should produce a self-contained HTML file', async () => {
      const outDir = path.join(tmpDir, 'report');
      await generateHtmlReport(sampleReport, outDir);

      const html = await fs.readFile(path.join(outDir, 'index.html'), 'utf-8');

      expect(html).toContain('<!DOCTYPE html>');
      expect(html).toContain('<style>');
      expect(html).toContain('<script>');
      expect(html).not.toContain('<link rel="stylesheet"');
      expect(html).not.toContain('src="http');
    });

    it('should embed the report data in the HTML', async () => {
      const outDir = path.join(tmpDir, 'report');
      await generateHtmlReport(sampleReport, outDir);

      const html = await fs.readFile(path.join(outDir, 'index.html'), 'utf-8');

      expect(html).toContain('var DATA=');
      expect(html).toContain('"jest"');
      expect(html).toContain('"mocha"');
      expect(html).toContain('test/auth.test.js');
    });

    it('should include summary cards', async () => {
      const outDir = path.join(tmpDir, 'report');
      await generateHtmlReport(sampleReport, outDir);

      const html = await fs.readFile(path.join(outDir, 'index.html'), 'utf-8');

      expect(html).toContain('Files Scanned');
      expect(html).toContain('Test Files');
      expect(html).toContain('Frameworks');
      expect(html).toContain('Avg Confidence');
      expect(html).toContain('Directions');
    });

    it('should include the sortable file table', async () => {
      const outDir = path.join(tmpDir, 'report');
      await generateHtmlReport(sampleReport, outDir);

      const html = await fs.readFile(path.join(outDir, 'index.html'), 'utf-8');

      expect(html).toContain('data-col="path"');
      expect(html).toContain('data-col="confidence"');
      expect(html).toContain('file-row');
      expect(html).toContain('search');
    });

    it('should include supported directions', async () => {
      const outDir = path.join(tmpDir, 'report');
      await generateHtmlReport(sampleReport, outDir);

      const html = await fs.readFile(path.join(outDir, 'index.html'), 'utf-8');

      expect(html).toContain('Supported Directions');
      expect(html).toContain('pipeline');
      expect(html).toContain('legacy');
    });

    it('should include download JSON button', async () => {
      const outDir = path.join(tmpDir, 'report');
      await generateHtmlReport(sampleReport, outDir);

      const html = await fs.readFile(path.join(outDir, 'index.html'), 'utf-8');

      expect(html).toContain('dl-json');
      expect(html).toContain('Download JSON');
    });

    it('should create parent directories if they do not exist', async () => {
      const outDir = path.join(tmpDir, 'deeply', 'nested', 'report');
      await generateHtmlReport(sampleReport, outDir);

      const htmlStat = await fs.stat(path.join(outDir, 'index.html'));
      expect(htmlStat.isFile()).toBe(true);
    });

    it('should HTML-escape file paths to prevent XSS', async () => {
      const xssReport = {
        ...sampleReport,
        files: [
          {
            path: '<script>alert("xss")</script>',
            type: 'test',
            framework: 'jest',
            confidence: 90,
            candidates: [],
            warnings: [],
          },
        ],
      };

      const outDir = path.join(tmpDir, 'report');
      await generateHtmlReport(xssReport, outDir);

      const html = await fs.readFile(path.join(outDir, 'index.html'), 'utf-8');

      expect(html).not.toContain('<script>alert("xss")</script>');
      expect(html).toContain('&lt;script&gt;');
    });
  });
});
