/**
 * Hamlet — Multi-framework test converter.
 *
 * Public API surface:
 *   Functions: convertFile, convertRepository, convertConfig,
 *              convertCypressToPlaywright, processTestFiles,
 *              validateTests, generateReport
 *   Classes:   RepositoryConverter, BatchProcessor, ConversionReporter
 *   Constants: VERSION, SUPPORTED_TEST_TYPES, DEFAULT_OPTIONS
 *
 * Internal/advanced classes and utilities are available via the
 * "hamlet-converter/internals" subpath export.
 */

// Core orchestration
import {
  RepositoryConverter,
  convertRepository,
} from './converter/repoConverter.js';
import {
  BatchProcessor,
  processTestFiles,
} from './converter/batchProcessor.js';

// File conversion
import {
  convertCypressToPlaywright,
  convertConfig,
  convertFile,
} from './converter/fileConverter.js';

// Validation
import { TestValidator } from './converter/validator.js';

// Reporting
import { ConversionReporter } from './utils/reporter.js';

/**
 * Validate converted tests
 * @param {string} testDir - Directory containing converted tests
 * @param {Object} options - Validation options
 * @returns {Promise<Object>} - Validation results
 */
export async function validateTests(testDir, _options = {}) {
  const validator = new TestValidator();
  return validator.validateConvertedTests(testDir);
}

/**
 * Generate conversion report
 * @param {string} outputPath - Output path for report
 * @param {string} format - Report format (html, json, markdown)
 * @param {Object} data - Report data
 * @returns {Promise<void>}
 */
export async function generateReport(outputPath, format = 'json', data = {}) {
  const reporter = new ConversionReporter({ format });
  return reporter.generateReport(data, outputPath);
}

// Re-export public functions
export { convertRepository, processTestFiles };
export { convertCypressToPlaywright, convertConfig, convertFile };

// Re-export public classes
export { RepositoryConverter, BatchProcessor, ConversionReporter };

// Constants — derive version from package.json (single source of truth)
import { createRequire } from 'module';
const __require = createRequire(import.meta.url);
export const VERSION = __require('../package.json').version;
export const SUPPORTED_TEST_TYPES = [
  'e2e',
  'component',
  'api',
  'visual',
  'accessibility',
  'performance',
  'mobile',
];

export const DEFAULT_OPTIONS = {
  typescript: false,
  validate: true,
  compareVisuals: false,
  convertPlugins: true,
  preserveStructure: true,
  report: 'json',
  batchSize: 5,
  timeout: 30000,
};
