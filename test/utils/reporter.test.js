import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import { ConversionReporter } from '../../src/utils/reporter.js';

describe('ConversionReporter', () => {
  let reporter;

  beforeEach(() => {
    reporter = new ConversionReporter();
  });

  describe('constructor', () => {
    it('should initialize with default options', () => {
      expect(reporter.options.outputDir).toBe('reports');
      expect(reporter.options.format).toBe('html');
      expect(reporter.options.includeTimestamps).toBe(true);
      expect(reporter.options.includeLogs).toBe(true);
    });

    it('should accept custom options', () => {
      const custom = new ConversionReporter({ format: 'json', outputDir: 'out' });
      expect(custom.options.format).toBe('json');
      expect(custom.options.outputDir).toBe('out');
    });

    it('should initialize empty data structures', () => {
      expect(reporter.data.testResults.passed).toEqual([]);
      expect(reporter.data.testResults.failed).toEqual([]);
      expect(reporter.data.testResults.skipped).toEqual([]);
      expect(reporter.data.validationResults.passed).toEqual([]);
      expect(reporter.data.conversionSteps).toEqual([]);
    });
  });

  describe('startReport', () => {
    it('should set start time', () => {
      reporter.startReport();
      expect(reporter.data.summary.startTime).toBeInstanceOf(Date);
    });
  });

  describe('endReport', () => {
    it('should set end time', () => {
      reporter.endReport();
      expect(reporter.data.summary.endTime).toBeInstanceOf(Date);
    });
  });

  describe('addTestResult', () => {
    it('should add passed result to passed array', () => {
      reporter.addTestResult({ status: 'passed', name: 'test1' });
      expect(reporter.data.testResults.passed).toHaveLength(1);
    });

    it('should add failed result to failed array', () => {
      reporter.addTestResult({ status: 'failed', name: 'test2' });
      expect(reporter.data.testResults.failed).toHaveLength(1);
    });

    it('should add other results to skipped array', () => {
      reporter.addTestResult({ status: 'skipped', name: 'test3' });
      expect(reporter.data.testResults.skipped).toHaveLength(1);
    });
  });

  describe('addValidationResult', () => {
    it('should add passed validation to passed array', () => {
      reporter.addValidationResult({ status: 'passed', check: 'syntax' });
      expect(reporter.data.validationResults.passed).toHaveLength(1);
    });

    it('should add failed validation to failed array', () => {
      reporter.addValidationResult({ status: 'failed', check: 'import' });
      expect(reporter.data.validationResults.failed).toHaveLength(1);
    });

    it('should add warnings to warnings array', () => {
      reporter.addValidationResult({ status: 'warning', check: 'style' });
      expect(reporter.data.validationResults.warnings).toHaveLength(1);
    });
  });

  describe('addValidationResults', () => {
    it('should accept a validation report details object', () => {
      reporter.addValidationResults({
        details: {
          passed: [{ check: 'syntax' }],
          failed: [{ check: 'imports' }],
          skipped: [{ check: 'hooks' }],
          errors: [{ check: 'runtime' }],
        },
      });

      expect(reporter.data.validationResults.passed).toHaveLength(1);
      expect(reporter.data.validationResults.failed).toHaveLength(1);
      expect(reporter.data.validationResults.warnings).toHaveLength(2);
    });
  });

  describe('addVisualResult', () => {
    it('should add matching result', () => {
      reporter.addVisualResult({ matches: true, test: 'login' });
      expect(reporter.data.visualResults.matches).toHaveLength(1);
    });

    it('should add error result', () => {
      reporter.addVisualResult({ error: 'missing screenshot' });
      expect(reporter.data.visualResults.errors).toHaveLength(1);
    });

    it('should add mismatch result', () => {
      reporter.addVisualResult({ matches: false, difference: '5%' });
      expect(reporter.data.visualResults.mismatches).toHaveLength(1);
    });
  });

  describe('addVisualResults', () => {
    it('should accept a single visual result object', () => {
      reporter.addVisualResults({ matches: true, test: 'dashboard' });
      expect(reporter.data.visualResults.matches).toHaveLength(1);
    });
  });

  describe('recordStep', () => {
    it('should record conversion step with timestamp', () => {
      reporter.recordStep('Converting tests', 'success', { count: 5 });
      expect(reporter.data.conversionSteps).toHaveLength(1);
      expect(reporter.data.conversionSteps[0].step).toBe('Converting tests');
      expect(reporter.data.conversionSteps[0].status).toBe('success');
      expect(reporter.data.conversionSteps[0].details.count).toBe(5);
      expect(reporter.data.conversionSteps[0].timestamp).toBeDefined();
    });
  });

  describe('calculatePercentage', () => {
    it('should calculate correct percentage', () => {
      reporter.addTestResult({ status: 'passed' });
      reporter.addTestResult({ status: 'passed' });
      reporter.addTestResult({ status: 'failed' });
      reporter.addTestResult({ status: 'skipped' });

      expect(reporter.calculatePercentage(2)).toBe('50.0');
    });

    it('should return 0 when no tests exist', () => {
      expect(reporter.calculatePercentage(0)).toBe(0);
    });
  });

  describe('generateHtmlReport', () => {
    it('should generate valid HTML', () => {
      reporter.startReport();
      reporter.addTestResult({ status: 'passed', name: 'test1' });
      reporter.endReport();

      const html = reporter.generateHtmlReport();
      expect(html).toContain('<!DOCTYPE html>');
      expect(html).toContain('Conversion Report');
      expect(html).toContain('</html>');
    });

    it('should escape untrusted HTML fields', () => {
      reporter.startReport();
      reporter.addValidationResult({
        status: 'passed',
        check: '</td><script>window.__xss__=1</script><td>',
        details: '<img src=x onerror=alert(1)>',
      });
      reporter.endReport();

      const html = reporter.generateHtmlReport();
      expect(html).not.toContain('<script>window.__xss__=1</script>');
      expect(html).toContain('&lt;script&gt;window.__xss__=1&lt;/script&gt;');
    });
  });

  describe('generateMarkdownReport', () => {
    it('should generate markdown content', () => {
      reporter.startReport();
      reporter.addTestResult({ status: 'passed', name: 'test1' });
      reporter.endReport();

      const md = reporter.generateMarkdownReport();
      expect(md).toContain('# Cypress to Playwright Conversion Report');
      expect(md).toContain('## Summary');
      expect(md).toContain('## Test Results');
    });
  });

  describe('generateReport', () => {
    let tmpDir;

    beforeEach(async () => {
      tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'terrain-reporter-'));
    });

    afterEach(async () => {
      await fs.rm(tmpDir, { recursive: true, force: true });
    });

    it('should honor explicit output file path and data payload', async () => {
      const reportPath = path.join(tmpDir, 'custom-report.json');
      const custom = new ConversionReporter({ format: 'json' });
      const output = await custom.generateReport(
        {
          summary: {
            totalFiles: 1,
            convertedFiles: 1,
            skippedFiles: 0,
            errors: [],
          },
        },
        reportPath
      );

      expect(output).toBe(path.resolve(reportPath));
      const raw = await fs.readFile(reportPath, 'utf8');
      const parsed = JSON.parse(raw);
      expect(parsed.summary.totalFiles).toBe(1);
    });

    it('should treat existing dotted directory paths as directories', async () => {
      const dottedDir = path.join(tmpDir, 'out.v1');
      await fs.mkdir(dottedDir, { recursive: true });
      const custom = new ConversionReporter({ format: 'json' });
      const output = await custom.generateReport(
        { summary: { totalFiles: 1, convertedFiles: 1, skippedFiles: 0, errors: [] } },
        dottedDir
      );

      expect(path.dirname(output)).toBe(path.resolve(dottedDir));
      const raw = await fs.readFile(output, 'utf8');
      const parsed = JSON.parse(raw);
      expect(parsed.summary.totalFiles).toBe(1);
    });
  });
});
