/**
 * Abstract base class for all framework converters
 * All converter implementations must extend this class
 */
export class BaseConverter {
  constructor(options = {}) {
    this.sourceFramework = null;
    this.targetFramework = null;
    this.options = {
      preserveComments: true,
      addTypeAnnotations: false,
      ...options
    };
    this.patterns = {};
    this.stats = {
      conversions: 0,
      warnings: [],
      errors: []
    };
  }

  /**
   * Convert test content from source to target framework
   * @param {string} content - Source test content
   * @param {Object} options - Conversion options
   * @returns {Promise<string>} - Converted test content
   */
  async convert(content, options = {}) {
    throw new Error('convert() must be implemented by subclass');
  }

  /**
   * Convert configuration file
   * @param {string} configPath - Path to config file
   * @param {Object} options - Conversion options
   * @returns {Promise<string>} - Converted config content
   */
  async convertConfig(configPath, options = {}) {
    throw new Error('convertConfig() must be implemented by subclass');
  }

  /**
   * Get required imports for target framework
   * @param {string[]} testTypes - Detected test types (e2e, api, component, etc.)
   * @returns {string[]} - Array of import statements
   */
  getImports(testTypes = []) {
    throw new Error('getImports() must be implemented by subclass');
  }

  /**
   * Detect test types from content
   * @param {string} content - Test content
   * @returns {string[]} - Array of detected test types
   */
  detectTestTypes(content) {
    throw new Error('detectTestTypes() must be implemented by subclass');
  }

  /**
   * Validate converted output
   * @param {string} content - Converted content
   * @returns {Object} - Validation result { valid: boolean, errors: string[] }
   */
  validate(content) {
    const errors = [];

    // Basic syntax check
    try {
      new Function(content);
    } catch (e) {
      errors.push(`Syntax error: ${e.message}`);
    }

    return {
      valid: errors.length === 0,
      errors
    };
  }

  /**
   * Get conversion statistics
   * @returns {Object} - Statistics object
   */
  getStats() {
    return { ...this.stats };
  }

  /**
   * Reset converter state
   */
  reset() {
    this.stats = {
      conversions: 0,
      warnings: [],
      errors: []
    };
  }

  /**
   * Add a warning to the stats
   * @param {string} message - Warning message
   * @param {Object} context - Additional context
   */
  addWarning(message, context = {}) {
    this.stats.warnings.push({ message, context, timestamp: new Date().toISOString() });
  }

  /**
   * Add an error to the stats
   * @param {string} message - Error message
   * @param {Object} context - Additional context
   */
  addError(message, context = {}) {
    this.stats.errors.push({ message, context, timestamp: new Date().toISOString() });
  }

  /**
   * Get source framework name
   * @returns {string}
   */
  getSourceFramework() {
    return this.sourceFramework;
  }

  /**
   * Get target framework name
   * @returns {string}
   */
  getTargetFramework() {
    return this.targetFramework;
  }

  /**
   * Get conversion direction string
   * @returns {string} - e.g., "cypress-to-playwright"
   */
  getConversionDirection() {
    return `${this.sourceFramework}-to-${this.targetFramework}`;
  }
}
