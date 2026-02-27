import { ProjectAnalyzer } from '../../src/core/ProjectAnalyzer.js';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const fixturesDir = path.resolve(__dirname, '../fixtures/analyze');

describe('ProjectAnalyzer', () => {
  let analyzer;

  beforeEach(() => {
    analyzer = new ProjectAnalyzer();
  });

  describe('analyze', () => {
    it('should return required top-level keys', async () => {
      const report = await analyzer.analyze(fixturesDir);
      expect(report).toHaveProperty('schemaVersion');
      expect(report).toHaveProperty('meta');
      expect(report).toHaveProperty('summary');
      expect(report).toHaveProperty('files');
    });

    it('should set schemaVersion to 1.0.0', async () => {
      const report = await analyzer.analyze(fixturesDir);
      expect(report.schemaVersion).toBe('1.0.0');
    });

    it('should include correct meta fields', async () => {
      const report = await analyzer.analyze(fixturesDir);
      expect(report.meta).toHaveProperty('hamletVersion');
      expect(report.meta).toHaveProperty('nodeVersion');
      expect(report.meta).toHaveProperty('generatedAt');
      expect(report.meta).toHaveProperty('root');
      expect(typeof report.meta.hamletVersion).toBe('string');
      expect(report.meta.nodeVersion).toBe(process.version);
      expect(report.meta.root).toBe(path.resolve(fixturesDir));
    });

    it('should sort files lexicographically by path', async () => {
      const report = await analyzer.analyze(fixturesDir);
      const paths = report.files.map((f) => f.path);
      const sorted = [...paths].sort((a, b) => (a < b ? -1 : a > b ? 1 : 0));
      expect(paths).toEqual(sorted);
    });

    it('should detect frameworks from fixture files', async () => {
      const report = await analyzer.analyze(fixturesDir);
      const frameworks = report.summary.frameworksDetected;
      expect(frameworks.length).toBeGreaterThan(0);
    });

    it('should include candidates array for each file', async () => {
      const report = await analyzer.analyze(fixturesDir);
      for (const file of report.files) {
        expect(Array.isArray(file.candidates)).toBe(true);
        for (const c of file.candidates) {
          expect(c).toHaveProperty('framework');
          expect(c).toHaveProperty('score');
          expect(typeof c.score).toBe('number');
        }
      }
    });

    it('should include warnings array for each file', async () => {
      const report = await analyzer.analyze(fixturesDir);
      for (const file of report.files) {
        expect(Array.isArray(file.warnings)).toBe(true);
      }
    });

    it('should respect maxFiles cap', async () => {
      const report = await analyzer.analyze(fixturesDir, { maxFiles: 1 });
      expect(report.files.length).toBeLessThanOrEqual(1);
    });

    it('should return empty files array for empty directory', async () => {
      const fs = await import('fs/promises');
      const emptyDir = path.join(fixturesDir, '__empty__');
      await fs.mkdir(emptyDir, { recursive: true });
      try {
        const report = await analyzer.analyze(emptyDir);
        expect(report.files).toEqual([]);
        expect(report.summary.fileCount).toBe(0);
      } finally {
        await fs.rmdir(emptyDir).catch(() => {});
      }
    });

    it('should annotate directionsSupported with pipelineBacked boolean', async () => {
      const report = await analyzer.analyze(fixturesDir);
      for (const d of report.summary.directionsSupported) {
        expect(d).toHaveProperty('from');
        expect(d).toHaveProperty('to');
        expect(typeof d.pipelineBacked).toBe('boolean');
      }
    });

    it('should count test files correctly', async () => {
      const report = await analyzer.analyze(fixturesDir);
      const testFiles = report.files.filter((f) => f.type === 'test');
      expect(report.summary.testFileCount).toBe(testFiles.length);
    });

    it('should have non-negative confidenceAvg', async () => {
      const report = await analyzer.analyze(fixturesDir);
      expect(report.summary.confidenceAvg).toBeGreaterThanOrEqual(0);
    });
  });
});
