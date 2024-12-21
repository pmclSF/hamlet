/**
 * Export all configuration settings
 */

// Default configurations
export { conversionConfig } from './defaults/conversion.js';
export { reportingConfig } from './defaults/reporting.js';
export { typescriptConfig } from './defaults/typescript.js';

// Conversion patterns
export { assertionPatterns } from './patterns/assertions.js';
export { commandPatterns } from './patterns/commands.js';
export { pluginPatterns } from './patterns/plugins.js';

// Test management configurations
export { azureConfig } from './test-management/azure.js';
export { testrailConfig } from './test-management/testrail.js';
export { xrayConfig } from './test-management/xray.js';

/**
 * Default configuration object
 */
export const defaultConfig = {
  // Base configuration
  projectRoot: process.cwd(),
  outputDir: './playwright-tests',
  
  // Feature flags
  features: {
    typescript: true,
    validation: true,
    visualComparison: true,
    reporting: true,
    testManagement: false
  },

  // Test patterns
  testMatch: [
    '**/*.cy.{js,ts}',
    '**/cypress/integration/**/*.{js,ts}',
    '**/cypress/e2e/**/*.{js,ts}',
    '**/cypress/component/**/*.{js,ts}'
  ],

  // Files to ignore
  ignore: [
    '**/node_modules/**',
    '**/dist/**',
    '**/build/**',
    '**/coverage/**'
  ],

  // Error handling
  errorHandling: {
    continueOnError: true,
    throwOnFatalError: true,
    logLevel: 'info'
  }
};

/**
 * Load and merge configuration
 * @param {Object} userConfig - User configuration
 * @returns {Object} - Merged configuration
 */
export function loadConfig(userConfig = {}) {
  return {
    ...defaultConfig,
    ...userConfig,
    features: {
      ...defaultConfig.features,
      ...userConfig.features
    }
  };
}

export default {
  defaultConfig,
  loadConfig,
  conversionConfig,
  reportingConfig,
  typescriptConfig,
  assertionPatterns,
  commandPatterns,
  pluginPatterns,
  azureConfig,
  testrailConfig,
  xrayConfig
};