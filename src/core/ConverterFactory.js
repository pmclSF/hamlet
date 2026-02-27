/**
 * Supported frameworks
 */
export const FRAMEWORKS = {
  CYPRESS: 'cypress',
  PLAYWRIGHT: 'playwright',
  SELENIUM: 'selenium',
  JEST: 'jest',
  VITEST: 'vitest',
  MOCHA: 'mocha',
  JASMINE: 'jasmine',
  JUNIT4: 'junit4',
  JUNIT5: 'junit5',
  TESTNG: 'testng',
  PYTEST: 'pytest',
  UNITTEST: 'unittest',
  NOSE2: 'nose2',
  WEBDRIVERIO: 'webdriverio',
  PUPPETEER: 'puppeteer',
  TESTCAFE: 'testcafe',
};

/**
 * Maps each framework to its language directory.
 * Used by _loadFrameworkDefinitions to resolve import paths.
 */
const FRAMEWORK_LANGUAGE = {
  cypress: 'javascript',
  playwright: 'javascript',
  selenium: 'javascript',
  jest: 'javascript',
  vitest: 'javascript',
  mocha: 'javascript',
  jasmine: 'javascript',
  junit4: 'java',
  junit5: 'java',
  testng: 'java',
  pytest: 'python',
  unittest: 'python',
  nose2: 'python',
  webdriverio: 'javascript',
  puppeteer: 'javascript',
  testcafe: 'javascript',
};

/**
 * Maps framework names to non-standard filenames when the framework name
 * cannot be used directly as a filename (e.g., Node reserved words).
 */
const FRAMEWORK_FILE_OVERRIDE = {
  unittest: "unittest_fw",
};

/**
 * Conversion directions handled by the new pipeline (v2 architecture).
 * All other directions fall back to legacy converters.
 */
const PIPELINE_DIRECTIONS = new Set([
  'cypress-playwright',
  'jest-vitest',
  'mocha-jest',
  'jasmine-jest',
  'jest-mocha',
  'jest-jasmine',
  'junit4-junit5',
  'junit5-testng',
  'testng-junit5',
  'pytest-unittest',
  'unittest-pytest',
  'nose2-pytest',
  'webdriverio-playwright',
  'webdriverio-cypress',
  'playwright-webdriverio',
  'cypress-webdriverio',
  'puppeteer-playwright',
  'playwright-puppeteer',
  'testcafe-playwright',
  'testcafe-cypress',
]);

/**
 * Factory for creating framework converters.
 *
 * Routes through the new ConversionPipeline for directions that have been
 * migrated (cypress→playwright, jest→vitest). Falls back to legacy
 * converter classes for all other directions.
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
      ["cypress-selenium", () => import("../converters/CypressToSelenium.js")],
      [
        "playwright-cypress",
        () => import("../converters/PlaywrightToCypress.js"),
      ],
      [
        "playwright-selenium",
        () => import("../converters/PlaywrightToSelenium.js"),
      ],
      ["selenium-cypress", () => import("../converters/SeleniumToCypress.js")],
      [
        "selenium-playwright",
        () => import("../converters/SeleniumToPlaywright.js"),
      ],
    ];

    for (const [key, loader] of converterModules) {
      this.converters.set(key, loader);
    }

    this.initialized = true;
  }

  /**
   * Create a converter for the specified frameworks.
   *
   * For pipeline-backed directions (cypress→playwright, jest→vitest),
   * returns a PipelineConverter. For legacy directions, returns the
   * original converter class instance.
   *
   * @param {string} from - Source framework
   * @param {string} to - Target framework
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
      throw new Error(
        `Invalid source framework: ${from}. Valid options: ${validFrameworks.join(', ')}`,
      );
    }
    if (!validFrameworks.includes(toLower)) {
      throw new Error(
        `Invalid target framework: ${to}. Valid options: ${validFrameworks.join(', ')}`,
      );
    }
    if (fromLower === toLower) {
      throw new Error("Source and target frameworks must be different");
    }

    const key = `${fromLower}-${toLower}`;

    // Pipeline-backed directions
    if (PIPELINE_DIRECTIONS.has(key)) {
      return this._createPipelineConverter(fromLower, toLower, options);
    }

    // Legacy converter directions
    const loader = this.converters.get(key);
    if (!loader) {
      throw new Error(
        `Unsupported conversion: ${from} to ${to}. ` +
          `Supported conversions: ${this.getSupportedConversions().join(', ')}`,
      );
    }

    try {
      const module = await loader();
      const ConverterClass = module.default || Object.values(module)[0];
      return new ConverterClass(options);
    } catch (error) {
      throw new Error(
        `Failed to load converter for ${from} to ${to}: ${error.message}`
      );
    }
  }

  /**
   * Create a PipelineConverter for a migrated direction.
   * Uses dynamic imports to load framework definitions lazily.
   *
   * @param {string} from - Source framework name
   * @param {string} to - Target framework name
   * @param {Object} options - Converter options
   * @returns {Promise<import('./PipelineConverter.js').PipelineConverter>}
   */
  static async _createPipelineConverter(from, to, options) {
    const { PipelineConverter } = await import("./PipelineConverter.js");

    // Load framework definitions based on direction
    const definitions = await this._loadFrameworkDefinitions(from, to);
    return new PipelineConverter(from, to, definitions, options);
  }

  /**
   * Load framework definitions for a given conversion direction.
   * @param {string} from - Source framework name
   * @param {string} to - Target framework name
   * @returns {Promise<Object[]>} Array of framework definitions
   */
  static async _loadFrameworkDefinitions(from, to) {
    const names = new Set([from, to]);
    const definitions = [];

    for (const name of names) {
      const language = FRAMEWORK_LANGUAGE[name] || "javascript";
      const fileName = FRAMEWORK_FILE_OVERRIDE[name] || name;
      const mod = await import(
        `../languages/${language}/frameworks/${fileName}.js`
      );
      definitions.push(mod.default);
    }

    return definitions;
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
      'selenium-playwright',
      'jest-vitest',
      'mocha-jest',
      'jasmine-jest',
      'jest-mocha',
      'jest-jasmine',
      'junit4-junit5',
      'junit5-testng',
      'testng-junit5',
      'pytest-unittest',
      'unittest-pytest',
      'nose2-pytest',
      'webdriverio-playwright',
      'webdriverio-cypress',
      'playwright-webdriverio',
      'cypress-webdriverio',
      'puppeteer-playwright',
      'playwright-puppeteer',
      'testcafe-playwright',
      'testcafe-cypress',
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
   * Check if a conversion direction is backed by the new pipeline.
   * @param {string} from - Source framework
   * @param {string} to - Target framework
   * @returns {boolean}
   */
  static isPipelineBacked(from, to) {
    return PIPELINE_DIRECTIONS.has(`${from.toLowerCase()}-${to.toLowerCase()}`);
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
