// Import core converters and processors
import { RepositoryConverter } from './converter/repoConverter.js';
import { BatchProcessor } from './converter/batchProcessor.js';
import { DependencyAnalyzer } from './converter/dependencyAnalyzer.js';
import { TestMetadataCollector } from './converter/metadataCollector.js';

// Import validators and specialized converters
import { TestValidator } from './converter/validator.js';
import { TypeScriptConverter } from './converter/typescript.js';
import { PluginConverter } from './converter/plugins.js';
import { VisualComparison } from './converter/visual.js';
import { TestMapper } from './converter/mapper.js';

// Import file conversion functions (extracted to break circular dependency)
import {
  convertCypressToPlaywright,
  convertConfig,
  convertFile,
} from './converter/fileConverter.js';

// Import orchestration functions
import { convertRepository } from './converter/repoConverter.js';
import { processTestFiles } from './converter/batchProcessor.js';

// Import reporters and utilities
import { ConversionReporter } from './utils/reporter.js';
import {
  fileUtils,
  stringUtils,
  codeUtils,
  testUtils,
  reportUtils,
  logUtils,
} from './utils/helpers.js';

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

// Re-export orchestration functions
export { convertRepository, processTestFiles };

// Re-export file conversion functions
export { convertCypressToPlaywright, convertConfig, convertFile };

// Re-export imported classes
export {
  RepositoryConverter,
  BatchProcessor,
  DependencyAnalyzer,
  TestMetadataCollector,
  TestValidator,
  TypeScriptConverter,
  PluginConverter,
  VisualComparison,
  TestMapper,
};

// Re-export utilities
export { fileUtils, stringUtils, codeUtils, testUtils, reportUtils, logUtils };

// Re-export reporter
export { ConversionReporter };

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
