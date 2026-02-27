import { spawnSync } from 'child_process';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const rootDir = path.resolve(__dirname, '../..');
const cliPath = path.resolve(rootDir, 'bin/hamlet.js');
const fixturesDir = path.resolve(__dirname, '../fixtures/analyze');

function runCLI(args) {
  const result = spawnSync('node', [cliPath, ...args], {
    encoding: 'utf8',
    timeout: 30000,
    stdio: ['ignore', 'pipe', 'pipe'],
  });
  return result;
}

describe('CLI analyze command', () => {
  describe('--json mode', () => {
    it('should exit 0 and produce valid JSON', () => {
      const result = runCLI(['analyze', fixturesDir, '--json']);
      expect(result.status).toBe(0);
      const report = JSON.parse(result.stdout);
      expect(report).toBeDefined();
    });

    it('should include all required top-level keys', () => {
      const result = runCLI(['analyze', fixturesDir, '--json']);
      const report = JSON.parse(result.stdout);
      expect(report).toHaveProperty('schemaVersion');
      expect(report).toHaveProperty('meta');
      expect(report).toHaveProperty('summary');
      expect(report).toHaveProperty('files');
    });

    it('should have correct meta field types', () => {
      const result = runCLI(['analyze', fixturesDir, '--json']);
      const report = JSON.parse(result.stdout);
      expect(typeof report.meta.hamletVersion).toBe('string');
      expect(typeof report.meta.nodeVersion).toBe('string');
      expect(typeof report.meta.generatedAt).toBe('string');
      expect(typeof report.meta.root).toBe('string');
    });

    it('should have correct summary field types', () => {
      const result = runCLI(['analyze', fixturesDir, '--json']);
      const report = JSON.parse(result.stdout);
      expect(typeof report.summary.fileCount).toBe('number');
      expect(typeof report.summary.testFileCount).toBe('number');
      expect(Array.isArray(report.summary.frameworksDetected)).toBe(true);
      expect(Array.isArray(report.summary.directionsSupported)).toBe(true);
      expect(typeof report.summary.confidenceAvg).toBe('number');
    });

    it('should have correct file entry fields', () => {
      const result = runCLI(['analyze', fixturesDir, '--json']);
      const report = JSON.parse(result.stdout);
      expect(report.files.length).toBeGreaterThan(0);
      for (const f of report.files) {
        expect(typeof f.path).toBe('string');
        expect(typeof f.type).toBe('string');
        expect(Array.isArray(f.candidates)).toBe(true);
        expect(typeof f.confidence).toBe('number');
        expect(Array.isArray(f.warnings)).toBe(true);
      }
    });

    it('should sort files lexicographically', () => {
      const result = runCLI(['analyze', fixturesDir, '--json']);
      const report = JSON.parse(result.stdout);
      const paths = report.files.map((f) => f.path);
      const sorted = [...paths].sort((a, b) => (a < b ? -1 : a > b ? 1 : 0));
      expect(paths).toEqual(sorted);
    });

    it('should detect jest and pytest from fixtures', () => {
      const result = runCLI(['analyze', fixturesDir, '--json']);
      const report = JSON.parse(result.stdout);
      const frameworks = report.summary.frameworksDetected;
      expect(frameworks).toContain('jest');
      expect(frameworks).toContain('pytest');
    });
  });

  describe('human output mode', () => {
    it('should show Analysis Summary', () => {
      const result = runCLI(['analyze', fixturesDir]);
      expect(result.status).toBe(0);
      expect(result.stdout).toContain('Analysis Summary');
    });
  });
});
