import fs from 'fs/promises';
import path from 'path';

import { DependencyAnalyzer } from './dependencyAnalyzer.js';
import { TestMetadataCollector } from './metadataCollector.js';
import { TestValidator } from './validator.js';
import { TypeScriptConverter } from './typescript.js';
import { PluginConverter } from './plugins.js';
import { VisualComparison } from './visual.js';
import { TestMapper } from './mapper.js';
import { ConversionReporter } from '../utils/reporter.js';
import { fileUtils, logUtils } from '../utils/helpers.js';

const logger = logUtils.createLogger('Converter');

/**
 * Detect type of Cypress test
 * @param {string} content - Test content
 * @returns {string[]} - Array of detected test types
 */
function detectTestType(content) {
  const patterns = {
    api: /cy\.request|cy\.intercept|\.then\s*\(\s*{\s*status/i,
    component: /cy\.mount|mount\(/i,
    accessibility: /cy\.injectAxe|cy\.checkA11y|aria-|role=/i,
    visual: /cy\.screenshot|matchImageSnapshot/i,
    performance: /cy\.lighthouse|performance\.|timing/i,
    mobile: /viewport|mobile|touch|swipe/i,
  };

  return Object.entries(patterns)
    .filter(([_, pattern]) => pattern.test(content))
    .map(([type]) => type);
}

/**
 * Generate required imports based on test type
 * @param {string[]} types - Array of test types
 * @returns {string[]} - Array of import statements
 */
function generateImports(types) {
  const imports = new Set(["import { test, expect } from '@playwright/test';"]);

  const typeImports = {
    api: "import { request } from '@playwright/test';",
    component: "import { mount } from '@playwright/experimental-ct-react';",
    accessibility: "import { injectAxe, checkA11y } from 'axe-playwright';",
    visual: "import { expect } from '@playwright/test';",
  };

  types.forEach((type) => {
    if (typeImports[type]) {
      imports.add(typeImports[type]);
    }
  });

  return Array.from(imports);
}
/**
 * Convert Cypress test to Playwright format
 * @param {string} cypressContent - Content of Cypress test file
 * @param {Object} options - Conversion options
 * @returns {string} - Converted Playwright test content
 */
export async function convertCypressToPlaywright(cypressContent, options = {}) {
  let playwrightContent = cypressContent;

  // Extract metadata inline from content (collectMetadata expects a file path)
  const metadataCollector =
    options.metadataCollector || new TestMetadataCollector();
  const _metadata = {
    type: metadataCollector.detectTestType(cypressContent),
    suites: metadataCollector.extractTestSuites(cypressContent),
    cases: metadataCollector.extractTestCases(cypressContent),
    tags: metadataCollector.extractTags(cypressContent),
    complexity: metadataCollector.calculateComplexity(cypressContent),
  };

  // Detect test type
  const testType = detectTestType(cypressContent);

  // Get required imports based on test type
  const imports = generateImports(testType);

  // Basic conversion patterns
  const conversions = {
    // Test Structure
    'describe\\(': 'test.describe(',
    'it\\(': 'test(',
    'cy\\.': 'await page.',
    'before\\(': 'test.beforeAll(',
    'after\\(': 'test.afterAll(',
    'beforeEach\\(': 'test.beforeEach(',
    'afterEach\\(': 'test.afterEach(',

    // Basic Commands
    'visit\\(': 'goto(',
    'get\\(': 'locator(',
    'find\\(': 'locator(',
    'type\\(': 'fill(',
    'click\\(': 'click(',
    'dblclick\\(': 'dblclick(',
    'rightclick\\(': 'click({ button: "right" })',
    'check\\(': 'check(',
    'uncheck\\(': 'uncheck(',
    'select\\(': 'selectOption(',
    'scrollTo\\(': 'scroll(',
    'scrollIntoView\\(': 'scrollIntoViewIfNeeded(',
    'trigger\\(': 'dispatchEvent(',
    'focus\\(': 'focus(',
    'blur\\(': 'blur(',
    'clear\\(': 'clear(',

    // Assertions
    "should\\('be.visible'\\)": 'toBeVisible()',
    "should\\('not.be.visible'\\)": 'toBeHidden()',
    "should\\('exist'\\)": 'toBeVisible()',
    "should\\('not.exist'\\)": 'toBeHidden()',
    "should\\('have.text',\\s*([^)]+)\\)": 'toHaveText($1)',
    "should\\('have.value',\\s*([^)]+)\\)": 'toHaveValue($1)',
    "should\\('be.checked'\\)": 'toBeChecked()',
    "should\\('be.disabled'\\)": 'toBeDisabled()',
    "should\\('be.enabled'\\)": 'toBeEnabled()',
    "should\\('have.class',\\s*([^)]+)\\)": 'toHaveClass($1)',
    "should\\('have.attr',\\s*([^)]+)\\)": 'toHaveAttribute($1)',
    "should\\('have.length'\\)": 'toHaveCount(',
    "should\\('be.empty'\\)": 'toBeEmpty()',
    "should\\('be.focused'\\)": 'toBeFocused()',

    // API Testing
    'request\\(': 'await request.fetch(',
    'intercept\\(': 'await page.route(',
    'wait\\(@([^)]+)\\)':
      'waitForResponse(response => response.url().includes($1))',

    // Component Testing
    'mount\\(': 'await mount(',
    '\\.shadow\\(\\)': '.shadowRoot()',

    // Accessibility Testing
    'injectAxe\\(': 'await injectAxe(page)',
    'checkA11y\\(': 'await checkA11y(page)',

    // Visual Testing
    'matchImageSnapshot\\(': 'screenshot({ name: ',

    // File Handling
    'readFile\\(': 'await fs.readFile(',
    'writeFile\\(': 'await fs.writeFile(',
    'fixture\\(': "await fs.readFile(path.join('fixtures', ",

    // Iframe Handling
    'iframe\\(\\)': 'frameLocator()',

    // Multiple Windows/Tabs
    'window\\(\\)': 'context.newPage()',

    // Local Storage
    'clearLocalStorage\\(': 'evaluate(() => localStorage.clear())',
    'clearCookies\\(': 'context.clearCookies()',

    // Mouse Events
    'hover\\(': 'hover(',
    'mousedown\\(': 'mouseDown(',
    'mouseup\\(': 'mouseUp(',
    'mousemove\\(': 'moveBy(',

    // Keyboard Events
    'keyboard\\(': 'keyboard.press(',
    'press\\(': 'press(',

    // Viewport/Responsive
    'viewport\\(': 'setViewportSize(',

    // Network
    'server\\(': '// Use page.route() instead of cy.server()',

    // State Management
    'window\\.store': 'await page.evaluate(() => window.store',

    // Database
    'task\\(': "await request.fetch('/api/db', ",

    // Custom Commands
    'Cypress\\.Commands\\.add\\(': '// Convert to Playwright helper function: ',

    // Extended Assertions
    "should\\('contain'\\)": 'toContain()',
    "should\\('include'\\)": 'toContain()',
    "should\\('have.length',\\s*([^)]+)\\)": 'toHaveCount($1)',
    "should\\('match'\\)": 'toMatch()',
    "should\\('be.gt'\\)": 'toBeGreaterThan()',
    "should\\('be.gte'\\)": 'toBeGreaterThanOrEqual()',
    "should\\('be.lt'\\)": 'toBeLessThan()',
    "should\\('be.lte'\\)": 'toBeLessThanOrEqual()',
    "should\\('be.null'\\)": 'toBeNull()',
    "should\\('be.undefined'\\)": 'toBeUndefined()',
    "should\\('be.true'\\)": 'toBeTruthy()',
    "should\\('be.false'\\)": 'toBeFalsy()',

    // Extended Commands
    'within\\(': 'locator(',
    'parents\\(': "locator('.. ",
    'children\\(': "locator('> ",
    'first\\(': 'first(',
    'last\\(': 'last(',
    'eq\\(': 'nth(',
    'closest\\(': 'closest(',
    'prev\\(': "locator(':prev')",
    'next\\(': "locator(':next')",
    "trigger\\('mouseover'\\)": 'hover()',
    "trigger\\('mouseenter'\\)": 'hover()',
    "trigger\\('mouseleave'\\)": 'hover({ force: false })',
    "trigger\\('focus'\\)": 'focus()',
    "trigger\\('blur'\\)": 'blur()',
    'select\\(([^)]+)\\)': (match) => `selectOption(${match[1]})`,
  };

  // Apply conversions
  for (const [cypressPattern, playwrightPattern] of Object.entries(
    conversions
  )) {
    playwrightContent = playwrightContent.replace(
      new RegExp(cypressPattern, 'g'),
      playwrightPattern
    );
  }

  // Setup test configuration based on detected types
  const setupConfig = {
    mode: 'parallel',
    timeout: options.timeout || 30000,
  };

  // Add test type specific setup
  let setup = `
  // Test type: ${testType.join(', ')}
  test.describe.configure(${JSON.stringify(setupConfig, null, 2)});
  `;

  // Clean up and format
  playwrightContent =
    playwrightContent
      // Make test callbacks async and include page parameter
      .replace(
        /test\((.*?),\s*\((.*?)\)\s*=>/g,
        'test($1, async ({ page' +
          (testType.includes('api') ? ', request' : '') +
          ' }) =>'
      )
      // Clean up any remaining text after the last closing brace
      .replace(/}[^}]*$/, '});')
      // Fix any remaining vistest to goto
      .replace(/vistest\(/g, 'goto(')
      // Remove any XML-style tags and their content
      .replace(/<\/?userStyle[^>]*>.*?<\/userStyle>/g, '')
      // Remove any other XML-style tags
      .replace(/<[^>]+>/g, '')
      // Remove any stray characters and whitespace at the end
      .replace(/[%$#@\s]+$/, '')
      // Add final newline
      .trim() + '\n';

  // Combine imports, setup, and converted content
  return imports.join('\n') + '\n\n' + setup + playwrightContent;
}

/**
 * Convert Cypress configuration to Playwright configuration
 * @param {string} configPath - Path to cypress.json
 * @param {Object} options - Conversion options
 * @returns {string} - Playwright config content
 */
export async function convertConfig(configPath, options = {}) {
  try {
    const cypressConfig = JSON.parse(await fs.readFile(configPath, 'utf8'));

    const playwrightConfig = {
      testDir: './tests',
      timeout: cypressConfig.defaultCommandTimeout || 4000,
      expect: {
        timeout: cypressConfig.defaultCommandTimeout || 4000,
      },
      use: {
        baseURL: cypressConfig.baseUrl,
        viewport:
          cypressConfig.viewportWidth && cypressConfig.viewportHeight
            ? {
                width: cypressConfig.viewportWidth,
                height: cypressConfig.viewportHeight,
              }
            : undefined,
        video: cypressConfig.video ? 'on' : 'off',
        screenshot: cypressConfig.screenshotOnFailure
          ? 'only-on-failure'
          : 'off',
        trace: options.trace || 'retain-on-failure',
        // Additional Playwright-specific options
        actionTimeout: cypressConfig.defaultCommandTimeout || 4000,
        navigationTimeout: cypressConfig.pageLoadTimeout || 30000,
        testIdAttribute: cypressConfig.testIdAttribute || 'data-testid',
      },
      projects: [
        {
          name: 'chromium',
          use: { browserName: 'chromium' },
        },
        {
          name: 'firefox',
          use: { browserName: 'firefox' },
        },
        {
          name: 'webkit',
          use: { browserName: 'webkit' },
        },
      ],
      // Additional configuration from Cypress
      retries: cypressConfig.retries || 0,
      workers: cypressConfig.numTestsKeptInMemory || undefined,
      reporter: cypressConfig.reporter
        ? [['html'], [cypressConfig.reporter]]
        : [['html']],
      reportSlowTests: {
        max: 5,
        threshold: cypressConfig.slowTestThreshold || 10000,
      },
      fullyParallel: true,
      forbidOnly: !!process.env.CI,
      maxFailures: cypressConfig.stopOnFirstFail ? 1 : 0,
    };

    // Handle Cypress plugins if they exist
    if (options.convertPlugins) {
      const pluginConverter = new PluginConverter();
      const pluginResult = await pluginConverter.convertPlugin(configPath);
      Object.assign(playwrightConfig.use, pluginResult.config);
    }

    // Generate config file content
    const configContent = `
  import { defineConfig, devices } from '@playwright/test';

  /**
   * Converted from Cypress configuration
   * @see https://playwright.dev/docs/test-configuration
   */
  export default ${JSON.stringify(playwrightConfig, null, 2)};
  `;

    // Save extended configuration if needed
    if (options.extendedConfig) {
      const extendedConfig = await generateExtendedConfig(cypressConfig);
      await fs.writeFile(
        path.join(path.dirname(configPath), 'playwright.extended.config.js'),
        extendedConfig
      );
    }

    return configContent;
  } catch (error) {
    logger.error('Failed to convert Cypress config:', error);
    throw error;
  }
}

/**
 * Convert a single file from Cypress to Playwright
 * @param {string} sourcePath - Path to source Cypress file
 * @param {string} outputPath - Path for output Playwright file
 * @param {Object} options - Conversion options
 */
export async function convertFile(sourcePath, outputPath, options = {}) {
  try {
    // Initialize collectors and analyzers
    const metadataCollector = new TestMetadataCollector();
    const dependencyAnalyzer = new DependencyAnalyzer();
    const reporter = options.reporter || new ConversionReporter();

    // Collect metadata and analyze dependencies
    const metadata = await metadataCollector.collectMetadata(sourcePath);
    const dependencies =
      await dependencyAnalyzer.analyzeDependencies(sourcePath);

    // Read and convert content
    const content = await fs.readFile(sourcePath, 'utf8');
    let converted = await convertCypressToPlaywright(content, {
      ...options,
      metadata,
      dependencies,
    });

    // Convert TypeScript if needed
    if (options.typescript && sourcePath.endsWith('.ts')) {
      const tsConverter = new TypeScriptConverter();
      converted = await tsConverter.convertContent(converted);
    }

    // Ensure output directory exists
    await fileUtils.ensureDir(path.dirname(outputPath));

    // Write converted file
    await fs.writeFile(outputPath, converted);

    // Validate if requested
    let validator = null;
    let validationResults = null;

    if (options.validate) {
      validator = new TestValidator();
      validationResults = await validator.validateTest(outputPath);
      reporter.addValidationResults(validationResults);
    }

    // Run visual comparison if requested
    let comparisonResults = null;
    if (options.compareVisuals) {
      const visualComparator = new VisualComparison();
      comparisonResults = await visualComparator.compareTest(
        sourcePath,
        outputPath
      );
      reporter.addVisualResults(comparisonResults);
    }

    // Add to test mapper
    if (options.mapTests) {
      const testMapper = new TestMapper();
      await testMapper.addMapping(sourcePath, outputPath);
    }

    logger.success(`Converted ${path.basename(sourcePath)}`);
    return {
      success: true,
      metadata,
      dependencies,
      outputPath,
      validationResults: validationResults,
      visualResults: comparisonResults,
    };
  } catch (error) {
    logger.error(`Failed to convert ${sourcePath}:`, error);
    throw error;
  }
}

/**
 * Generate extended Playwright configuration
 * @param {Object} cypressConfig - Original Cypress configuration
 * @returns {string} - Extended configuration content
 */
async function generateExtendedConfig(cypressConfig) {
  const extendedConfig = `
  import { defineConfig } from '@playwright/test';
  import baseConfig from './playwright.config';

  /**
   * Extended Playwright configuration with additional settings
   * converted from Cypress configuration
   */
  export default defineConfig({
    ...baseConfig,
    use: {
      ...baseConfig.use,
      // Additional browser context options
      contextOptions: {
        ignoreHTTPSErrors: ${!!cypressConfig.ignoreHTTPSErrors},
        bypassCSP: ${!!cypressConfig.modifyObstructiveCode},
        locale: '${cypressConfig.locale || 'en-US'}',
        timezoneId: '${cypressConfig.timezone || 'UTC'}',
        geolocation: ${cypressConfig.geolocation ? JSON.stringify(cypressConfig.geolocation) : 'undefined'},
        permissions: ${cypressConfig.permissions ? JSON.stringify(cypressConfig.permissions) : '[]'},
        offline: ${!!cypressConfig.offline}
      },
      // Screenshot options
      screenshot: {
        mode: '${cypressConfig.screenshotOnFailure ? 'only-on-failure' : 'off'}',
        fullPage: ${!!cypressConfig.screenshotOnRunFailure}
      },
      // Video options
      video: {
        mode: '${cypressConfig.video ? 'retain-on-failure' : 'off'}',
        size: ${cypressConfig.videoCompression ? JSON.stringify({ width: 1280, height: 720 }) : 'undefined'}
      }
    }
  });
  `;

  return extendedConfig;
}
