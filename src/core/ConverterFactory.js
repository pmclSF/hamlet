/**
 * Supported frameworks
 */
export const FRAMEWORKS = {
  CYPRESS: 'cypress',
  PLAYWRIGHT: 'playwright',
  SELENIUM: 'selenium'
};

/**
 * Factory for creating framework converters
 */
export class ConverterFactory {
  static converters = new Map();
  static initialized = false;

  /**
   * Initialize the factory with all available converters
   * Lazy loading to avoid circular dependencies
   */
  static async initialize() {
    if (this.initialized) return;

    const converterModules = [
      ['cypress-playwright', () => import('../converters/CypressToPlaywright.js')],
      ['cypress-selenium', () => import('../converters/CypressToSelenium.js')],
      ['playwright-cypress', () => import('../converters/PlaywrightToCypress.js')],
      ['playwright-selenium', () => import('../converters/PlaywrightToSelenium.js')],
      ['selenium-cypress', () => import('../converters/SeleniumToCypress.js')],
      ['selenium-playwright', () => import('../converters/SeleniumToPlaywright.js')]
    ];

    for (const [key, loader] of converterModules) {
      this.converters.set(key, loader);
    }

    this.initialized = true;
  }

  /**
   * Create a converter for the specified frameworks
   * @param {string} from - Source framework (cypress, playwright, selenium)
   * @param {string} to - Target framework (cypress, playwright, selenium)
   * @param {Object} options - Converter options
   * @returns {Promise<BaseConverter>} - Configured converter instance
   */
  static async createConverter(from, to, options = {}) {
    await this.initialize();

    const fromLower = from.toLowerCase();
    const toLower = to.toLowerCase();

    // Validate frameworks
    const validFrameworks = Object.values(FRAMEWORKS);
    if (!validFrameworks.includes(fromLower)) {
      throw new Error(`Invalid source framework: ${from}. Valid options: ${validFrameworks.join(', ')}`);
    }
    if (!validFrameworks.includes(toLower)) {
      throw new Error(`Invalid target framework: ${to}. Valid options: ${validFrameworks.join(', ')}`);
    }
    if (fromLower === toLower) {
      throw new Error('Source and target frameworks must be different');
    }

    const key = `${fromLower}-${toLower}`;
    const loader = this.converters.get(key);

    if (!loader) {
      throw new Error(
        `Unsupported conversion: ${from} to ${to}. ` +
        `Supported conversions: ${this.getSupportedConversions().join(', ')}`
      );
    }

    try {
      const module = await loader();
      const ConverterClass = module.default || Object.values(module)[0];
      return new ConverterClass(options);
    } catch (error) {
      throw new Error(`Failed to load converter for ${from} to ${to}: ${error.message}`);
    }
  }

  /**
   * Create converter synchronously (requires pre-loaded converters)
   * @param {string} from - Source framework
   * @param {string} to - Target framework
   * @param {Object} options - Converter options
   * @param {Object} converterClasses - Pre-loaded converter classes
   * @returns {BaseConverter} - Converter instance
   */
  static createConverterSync(from, to, options = {}, converterClasses = {}) {
    const fromLower = from.toLowerCase();
    const toLower = to.toLowerCase();
    const key = `${fromLower}-${toLower}`;

    const ConverterClass = converterClasses[key];
    if (!ConverterClass) {
      throw new Error(`Converter not found for ${from} to ${to}`);
    }

    return new ConverterClass(options);
  }

  /**
   * Get all supported conversion directions
   * @returns {string[]} - Array of "from-to" strings
   */
  static getSupportedConversions() {
    return [
      'cypress-playwright',
      'cypress-selenium',
      'playwright-cypress',
      'playwright-selenium',
      'selenium-cypress',
      'selenium-playwright'
    ];
  }

  /**
   * Check if a conversion direction is supported
   * @param {string} from - Source framework
   * @param {string} to - Target framework
   * @returns {boolean}
   */
  static isSupported(from, to) {
    const key = `${from.toLowerCase()}-${to.toLowerCase()}`;
    return this.getSupportedConversions().includes(key);
  }

  /**
   * Get all supported frameworks
   * @returns {string[]}
   */
  static getFrameworks() {
    return Object.values(FRAMEWORKS);
  }

  /**
   * Get conversion matrix showing all supported directions
   * @returns {Object} - Matrix object { from: { to: boolean } }
   */
  static getConversionMatrix() {
    const frameworks = this.getFrameworks();
    const matrix = {};

    for (const from of frameworks) {
      matrix[from] = {};
      for (const to of frameworks) {
        matrix[from][to] = from !== to && this.isSupported(from, to);
      }
    }

    return matrix;
  }
}
