import fs from 'fs/promises';
import path from 'path';
import { logUtils, reportUtils } from './helpers.js';

/**
 * Handles report generation for the conversion process
 */
export class ConversionReporter {
  constructor(options = {}) {
    this.options = {
      outputDir: options.outputDir || 'reports',
      format: options.format || 'html',
      includeTimestamps: options.includeTimestamps ?? true,
      includeLogs: options.includeLogs ?? true,
      ...options,
    };

    this.logger = logUtils.createLogger('Reporter');

    this.data = {
      summary: {
        startTime: null,
        endTime: null,
        totalFiles: 0,
        convertedFiles: 0,
        skippedFiles: 0,
        errors: [],
      },
      testResults: {
        passed: [],
        failed: [],
        skipped: [],
      },
      validationResults: {
        passed: [],
        failed: [],
        warnings: [],
      },
      visualResults: {
        matches: [],
        mismatches: [],
        errors: [],
      },
      conversionSteps: [],
    };
  }

  /**
   * Initialize report
   */
  startReport() {
    this.data.summary.startTime = new Date();
    this.logger.info('Started conversion report');
  }

  /**
   * Finalize report
   */
  endReport() {
    this.data.summary.endTime = new Date();
    this.logger.info('Completed conversion report');
  }

  /**
   * Add test result
   * @param {Object} result - Test result
   */
  addTestResult(result) {
    if (result.status === 'passed') {
      this.data.testResults.passed.push(result);
    } else if (result.status === 'failed') {
      this.data.testResults.failed.push(result);
    } else {
      this.data.testResults.skipped.push(result);
    }
  }

  /**
   * Add validation result
   * @param {Object} result - Validation result
   */
  addValidationResult(result) {
    if (result.status === 'passed') {
      this.data.validationResults.passed.push(result);
    } else if (result.status === 'failed') {
      this.data.validationResults.failed.push(result);
    } else {
      this.data.validationResults.warnings.push(result);
    }
  }

  /**
   * Add one or many validation results.
   * @param {Object|Object[]} results - Validation result(s)
   */
  addValidationResults(results) {
    if (!results) return;

    if (Array.isArray(results)) {
      results.forEach((result) => this.addValidationResult(result));
      return;
    }

    if (results.details && typeof results.details === 'object') {
      (results.details.passed || []).forEach((result) =>
        this.addValidationResult({ ...result, status: 'passed' })
      );
      (results.details.failed || []).forEach((result) =>
        this.addValidationResult({ ...result, status: 'failed' })
      );
      (results.details.skipped || []).forEach((result) =>
        this.addValidationResult({ ...result, status: 'warning' })
      );
      (results.details.errors || []).forEach((result) =>
        this.addValidationResult({ ...result, status: 'warning' })
      );
      return;
    }

    this.addValidationResult(results);
  }

  /**
   * Add visual comparison result
   * @param {Object} result - Visual comparison result
   */
  addVisualResult(result) {
    if (result.matches) {
      this.data.visualResults.matches.push(result);
    } else if (result.error) {
      this.data.visualResults.errors.push(result);
    } else {
      this.data.visualResults.mismatches.push(result);
    }
  }

  /**
   * Add one or many visual comparison results.
   * @param {Object|Object[]} results - Visual result(s)
   */
  addVisualResults(results) {
    if (!results) return;
    if (Array.isArray(results)) {
      results.forEach((result) => this.addVisualResult(result));
      return;
    }
    this.addVisualResult(results);
  }

  /**
   * Record conversion step
   * @param {string} step - Step description
   * @param {string} status - Step status
   * @param {Object} details - Step details
   */
  recordStep(step, status, details = {}) {
    this.data.conversionSteps.push({
      step,
      status,
      details,
      timestamp: new Date().toISOString(),
    });
  }

  /**
   * Generate final report
   * @param {Object|null} data - Optional report data payload
   * @param {string|null} outputPath - Optional output file or directory path
   * @returns {Promise<string>} - Path to generated report
   */
  async generateReport(data = null, outputPath = null) {
    try {
      if (data && typeof data === 'object') {
        this.data = data;
      }

      let reportPath;
      if (outputPath) {
        const resolvedOutput = path.resolve(outputPath);
        let outputStat = null;
        try {
          outputStat = await fs.stat(resolvedOutput);
        } catch (_err) {
          outputStat = null;
        }

        if (outputStat && outputStat.isDirectory()) {
          await fs.mkdir(resolvedOutput, { recursive: true });
          reportPath = path.join(
            resolvedOutput,
            `conversion-report-${Date.now()}.${this.options.format}`
          );
        } else if (outputStat && outputStat.isFile()) {
          reportPath = resolvedOutput;
          await fs.mkdir(path.dirname(reportPath), { recursive: true });
        } else {
          const hasExtension = !!path.extname(resolvedOutput);
          if (hasExtension) {
            reportPath = resolvedOutput;
            await fs.mkdir(path.dirname(reportPath), { recursive: true });
          } else {
            await fs.mkdir(resolvedOutput, { recursive: true });
            reportPath = path.join(
              resolvedOutput,
              `conversion-report-${Date.now()}.${this.options.format}`
            );
          }
        }
      } else {
        await fs.mkdir(this.options.outputDir, { recursive: true });
        reportPath = path.join(
          this.options.outputDir,
          `conversion-report-${Date.now()}.${this.options.format}`
        );
      }

      let content;
      if (this.options.format === 'html') {
        content = this.generateHtmlReport();
      } else if (this.options.format === 'json') {
        content = JSON.stringify(this.data, null, 2);
      } else if (this.options.format === 'md') {
        content = this.generateMarkdownReport();
      } else {
        throw new Error(`Unsupported report format: ${this.options.format}`);
      }

      await fs.writeFile(reportPath, content);
      this.logger.success(`Report generated at: ${reportPath}`);
      return reportPath;
    } catch (error) {
      this.logger.error('Failed to generate report');
      throw error;
    }
  }

  /**
   * Generate HTML report
   * @returns {string} - HTML content
   */
  generateHtmlReport() {
    const esc = (value) => this.escapeHtml(value);
    const statusClass = (value) => this.normalizeStatusClass(value);
    const duration = reportUtils.formatDuration(
      this.data.summary.endTime - this.data.summary.startTime
    );

    return `
<!DOCTYPE html>
<html>
<head>
  <title>Cypress to Playwright Conversion Report</title>
  <style>
    body {
      font-family: Arial, sans-serif;
      line-height: 1.6;
      margin: 2rem;
      color: #333;
    }
    
    .header {
      margin-bottom: 2rem;
      padding-bottom: 1rem;
      border-bottom: 2px solid #eee;
    }
    
    .summary {
      background: #f5f5f5;
      padding: 1rem;
      border-radius: 4px;
      margin-bottom: 2rem;
    }
    
    .section {
      margin-bottom: 2rem;
    }
    
    .success { color: #22c55e; }
    .error { color: #ef4444; }
    .warning { color: #f59e0b; }
    
    table {
      width: 100%;
      border-collapse: collapse;
      margin: 1rem 0;
    }
    
    th, td {
      border: 1px solid #ddd;
      padding: 0.5rem;
      text-align: left;
    }
    
    th {
      background: #f5f5f5;
    }
    
    .step {
      padding: 0.5rem;
      margin: 0.5rem 0;
      border-left: 4px solid #ddd;
    }
    
    .step.success { border-color: #22c55e; }
    .step.error { border-color: #ef4444; }
    .step.warning { border-color: #f59e0b; }
    
    .details {
      font-family: monospace;
      background: #f5f5f5;
      padding: 1rem;
      margin-top: 0.5rem;
      border-radius: 4px;
    }
  </style>
</head>
<body>
  <div class="header">
    <h1>Cypress to Playwright Conversion Report</h1>
    <p>Generated on: ${new Date().toLocaleString()}</p>
    <p>Duration: ${duration}</p>
  </div>

  <div class="summary">
    <h2>Summary</h2>
    <p>Total Files: ${this.data.summary.totalFiles}</p>
    <p>Converted: ${this.data.summary.convertedFiles}</p>
    <p>Skipped: ${this.data.summary.skippedFiles}</p>
    <p>Errors: ${this.data.summary.errors.length}</p>
  </div>

  <div class="section">
    <h2>Test Results</h2>
    <table>
      <tr>
        <th>Status</th>
        <th>Count</th>
        <th>Percentage</th>
      </tr>
      <tr class="success">
        <td>Passed</td>
        <td>${this.data.testResults.passed.length}</td>
        <td>${this.calculatePercentage(this.data.testResults.passed.length)}%</td>
      </tr>
      <tr class="error">
        <td>Failed</td>
        <td>${this.data.testResults.failed.length}</td>
        <td>${this.calculatePercentage(this.data.testResults.failed.length)}%</td>
      </tr>
      <tr class="warning">
        <td>Skipped</td>
        <td>${this.data.testResults.skipped.length}</td>
        <td>${this.calculatePercentage(this.data.testResults.skipped.length)}%</td>
      </tr>
    </table>
  </div>

  <div class="section">
    <h2>Validation Results</h2>
    <table>
      <tr>
        <th>Check</th>
        <th>Status</th>
        <th>Details</th>
      </tr>
      ${this.data.validationResults.passed
        .map(
          (result) => `
        <tr class="success">
          <td>${esc(result.check)}</td>
          <td>Passed</td>
          <td>${esc(result.details || '')}</td>
        </tr>
      `
        )
        .join('')}
      ${this.data.validationResults.failed
        .map(
          (result) => `
        <tr class="error">
          <td>${esc(result.check)}</td>
          <td>Failed</td>
          <td>${esc(result.details || '')}</td>
        </tr>
      `
        )
        .join('')}
    </table>
  </div>

  <div class="section">
    <h2>Visual Comparison Results</h2>
    <table>
      <tr>
        <th>Test</th>
        <th>Status</th>
        <th>Difference</th>
      </tr>
      ${this.data.visualResults.matches
        .map(
          (result) => `
        <tr class="success">
          <td>${esc(result.test)}</td>
          <td>Match</td>
          <td>${esc(result.difference || '0%')}</td>
        </tr>
      `
        )
        .join('')}
      ${this.data.visualResults.mismatches
        .map(
          (result) => `
        <tr class="error">
          <td>${esc(result.test)}</td>
          <td>Mismatch</td>
          <td>${esc(result.difference)}</td>
        </tr>
      `
        )
        .join('')}
    </table>
  </div>

  <div class="section">
    <h2>Conversion Steps</h2>
    ${this.data.conversionSteps
      .map(
        (step) => `
      <div class="step ${statusClass(step.status)}">
        <h3>${esc(step.step)}</h3>
        <p>Status: ${esc(step.status)}</p>
        ${
          step.details
            ? `
          <div class="details">
            <pre>${esc(JSON.stringify(step.details, null, 2))}</pre>
          </div>
        `
            : ''
        }
        ${
          this.options.includeTimestamps
            ? `
          <small>Timestamp: ${new Date(step.timestamp).toLocaleString()}</small>
        `
            : ''
        }
      </div>
    `
      )
      .join('')}
  </div>

  ${
    this.data.summary.errors.length > 0
      ? `
    <div class="section">
      <h2>Errors</h2>
      ${this.data.summary.errors
        .map(
          (error) => `
        <div class="step error">
          <h3>${esc(error.type)} Error</h3>
          <p>${esc(error.message)}</p>
          ${
            error.stack
              ? `
            <div class="details">
              <pre>${esc(error.stack)}</pre>
            </div>
          `
              : ''
          }
        </div>
      `
        )
        .join('')}
    </div>
  `
      : ''
  }
</body>
</html>`;
  }

  /**
   * Generate Markdown report
   * @returns {string} - Markdown content
   */
  generateMarkdownReport() {
    const duration = reportUtils.formatDuration(
      this.data.summary.endTime - this.data.summary.startTime
    );

    return `
# Cypress to Playwright Conversion Report

Generated on: ${new Date().toLocaleString()}
Duration: ${duration}

## Summary
- Total Files: ${this.data.summary.totalFiles}
- Converted: ${this.data.summary.convertedFiles}
- Skipped: ${this.data.summary.skippedFiles}
- Errors: ${this.data.summary.errors.length}

## Test Results
| Status  | Count | Percentage |
|---------|-------|------------|
| Passed  | ${this.data.testResults.passed.length} | ${this.calculatePercentage(this.data.testResults.passed.length)}% |
| Failed  | ${this.data.testResults.failed.length} | ${this.calculatePercentage(this.data.testResults.failed.length)}% |
| Skipped | ${this.data.testResults.skipped.length} | ${this.calculatePercentage(this.data.testResults.skipped.length)}% |

## Validation Results
${this.data.validationResults.passed
  .map(
    (result) => `
### ✅ ${result.check}
${result.details || 'No details provided'}
`
  )
  .join('\n')}

${this.data.validationResults.failed
  .map(
    (result) => `
### ❌ ${result.check}
${result.details || 'No details provided'}
`
  )
  .join('\n')}

## Visual Comparison Results
${this.data.visualResults.matches
  .map(
    (result) => `
### ✅ ${result.test}
- Status: Match
- Difference: ${result.difference || '0%'}
`
  )
  .join('\n')}

${this.data.visualResults.mismatches
  .map(
    (result) => `
### ❌ ${result.test}
- Status: Mismatch
- Difference: ${result.difference}
`
  )
  .join('\n')}

## Conversion Steps
${this.data.conversionSteps
  .map(
    (step) => `
### ${step.step}
- Status: ${step.status}
${step.details ? `- Details:\n\`\`\`json\n${JSON.stringify(step.details, null, 2)}\n\`\`\`` : ''}
${this.options.includeTimestamps ? `- Timestamp: ${new Date(step.timestamp).toLocaleString()}` : ''}
`
  )
  .join('\n')}

${
  this.data.summary.errors.length > 0
    ? `
## Errors
${this.data.summary.errors
  .map(
    (error) => `
### ${error.type} Error
${error.message}
${error.stack ? `\`\`\`\n${error.stack}\n\`\`\`` : ''}
`
  )
  .join('\n')}
`
    : ''
}`;
  }

  /**
   * Calculate percentage
   * @param {number} value - Value to calculate percentage for
   * @returns {number} - Calculated percentage
   */
  calculatePercentage(value) {
    const total =
      this.data.testResults.passed.length +
      this.data.testResults.failed.length +
      this.data.testResults.skipped.length;

    return total === 0 ? 0 : ((value / total) * 100).toFixed(1);
  }

  /**
   * Escape untrusted values for safe HTML output.
   * @param {unknown} value - Untrusted value
   * @returns {string}
   */
  escapeHtml(value) {
    if (value == null) return '';
    return String(value)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  /**
   * Normalize arbitrary status values to safe CSS class tokens.
   * @param {unknown} value - Status value
   * @returns {'success'|'error'|'warning'}
   */
  normalizeStatusClass(value) {
    const normalized = String(value || '').toLowerCase();
    if (normalized === 'success' || normalized === 'passed') return 'success';
    if (normalized === 'error' || normalized === 'failed') return 'error';
    return 'warning';
  }
}
