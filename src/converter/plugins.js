import chalk from 'chalk';

/**
 * Handles conversion of Cypress plugins to Playwright equivalents
 */
export class PluginConverter {
  constructor() {
    // Plugin mapping definitions
    this.pluginMappings = new Map([
      // File Upload plugins
      [
        'cypress-file-upload',
        {
          playwright: '@playwright/test',
          setup: `
          // File upload is built into Playwright
          await page.setInputFiles('input[type="file"]', 'path/to/file');
        `,
          config: {
            use: {
              acceptDownloads: true,
            },
          },
        },
      ],

      // Real Events
      [
        'cypress-real-events',
        {
          playwright: '@playwright/test',
          setup: `
          // Real events are built into Playwright
          await page.mouse.move(100, 200);
          await page.mouse.down();
          await page.mouse.up();
        `,
          config: {},
        },
      ],

      // XPath Support
      [
        'cypress-xpath',
        {
          playwright: '@playwright/test',
          setup: `
          // XPath is built into Playwright
          await page.locator('xpath=//button');
        `,
          config: {},
        },
      ],

      // Visual Testing
      [
        'cypress-image-snapshot',
        {
          playwright: '@playwright/test',
          setup: `
          // Use Playwright's built-in snapshot testing
          await expect(page).toHaveScreenshot();
        `,
          config: {
            use: {
              screenshot: 'on',
            },
          },
        },
      ],

      // Accessibility Testing
      [
        'cypress-axe',
        {
          playwright: 'axe-playwright',
          setup: `
          import { injectAxe, checkA11y } from 'axe-playwright';
          
          // Setup in beforeEach
          await injectAxe(page);
          await checkA11y(page);
        `,
          config: {
            use: {
              axePage: true,
            },
          },
        },
      ],

      // API Testing
      [
        'cypress-api',
        {
          playwright: '@playwright/test',
          setup: `
          // Use Playwright's APIRequestContext
          const request = await playwright.request.newContext();
          const response = await request.get('https://api.example.com');
        `,
          config: {},
        },
      ],

      // Browser Console
      [
        'cypress-log-to-output',
        {
          playwright: '@playwright/test',
          setup: `
          // Console logging is built into Playwright
          page.on('console', msg => console.log(msg.text()));
        `,
          config: {},
        },
      ],

      // Local Storage
      [
        'cypress-localstorage-commands',
        {
          playwright: '@playwright/test',
          setup: `
          // Local storage handling in Playwright
          await page.evaluate(() => window.localStorage.setItem('key', 'value'));
        `,
          config: {},
        },
      ],

      // Cookie Handling
      [
        'cypress-cookie',
        {
          playwright: '@playwright/test',
          setup: `
          // Cookie handling in Playwright
          await context.addCookies([{ name: 'cookie1', value: 'value1', url: 'https://example.com' }]);
        `,
          config: {},
        },
      ],

      // Database Commands
      [
        'cypress-sql-server',
        {
          playwright: 'playwright-sql',
          setup: `
          // Example SQL handling in Playwright
          import { sql } from 'playwright-sql';
          await sql.query('SELECT * FROM users');
        `,
          config: {},
        },
      ],

      // Authentication
      [
        'cypress-auth',
        {
          playwright: '@playwright/test',
          setup: `
          // Authentication handling in Playwright
          async function globalSetup() {
            const browser = await chromium.launch();
            const context = await browser.newContext();
            const page = await context.newPage();
            await page.goto('https://example.com/login');
            await page.fill('#email', 'user@example.com');
            await page.fill('#password', 'password');
            await page.click('#submit');
            await context.storageState({ path: 'auth.json' });
            await browser.close();
          }
        `,
          config: {
            use: {
              storageState: 'auth.json',
            },
          },
        },
      ],
    ]);

    // Plugin categories for better organization
    this.categories = {
      ui: ['cypress-file-upload', 'cypress-real-events', 'cypress-xpath'],
      testing: ['cypress-image-snapshot', 'cypress-axe'],
      api: ['cypress-api'],
      storage: ['cypress-localstorage-commands', 'cypress-cookie'],
      database: ['cypress-sql-server'],
      auth: ['cypress-auth'],
    };
  }

  /**
   * Convert a Cypress plugin to its Playwright equivalent
   * @param {string} pluginPath - Path to plugin file
   * @returns {Promise<Object>} - Conversion result
   */
  async convertPlugin(pluginPath) {
    try {
      // Extract plugin name from path
      const pluginName = pluginPath
        .split('/')
        .pop()
        .replace(/\.[jt]s$/, '');
      const detectedPlugins = this.detectPlugins(pluginName);

      const conversions = [];
      for (const plugin of detectedPlugins) {
        const conversion = await this.convertSinglePlugin(plugin);
        if (conversion) {
          conversions.push(conversion);
        }
      }

      return this.generatePluginOutput(conversions);
    } catch (error) {
      console.error(chalk.red(`Error converting plugin ${pluginPath}:`), error);
      throw error;
    }
  }

  /**
   * Detect plugins used in a file
   * @param {string} content - File content
   * @returns {string[]} - Array of detected plugins
   */
  detectPlugins(content) {
    const detected = new Set();

    // Check for require/import statements
    const importPattern =
      /(?:require|import)\s*\(['"](.*?)['"]|from\s+['"](.*?)['"]/g;
    let match;
    while ((match = importPattern.exec(content)) !== null) {
      const plugin = match[1] || match[2];
      if (this.pluginMappings.has(plugin)) {
        detected.add(plugin);
      }
    }

    // Check for plugin-specific patterns
    for (const [plugin] of this.pluginMappings) {
      if (this.hasPluginPatterns(content, plugin)) {
        detected.add(plugin);
      }
    }

    return Array.from(detected);
  }

  /**
   * Check if content contains plugin-specific patterns
   * @param {string} content - File content
   * @param {string} plugin - Plugin name
   * @returns {boolean} - Whether plugin patterns were found
   */
  hasPluginPatterns(content, plugin) {
    const patterns = {
      'cypress-file-upload': /cy\.fixture.*upload|attachFile/,
      'cypress-real-events': /realHover|realClick|realType/,
      'cypress-xpath': /cy\.xpath/,
      'cypress-image-snapshot': /matchImageSnapshot/,
      'cypress-axe': /cy\.checkA11y/,
      'cypress-api': /cy\.api/,
    };

    return patterns[plugin] ? patterns[plugin].test(content) : false;
  }

  /**
   * Convert a single plugin
   * @param {string} plugin - Plugin name
   * @returns {Promise<Object>} - Conversion details
   */
  async convertSinglePlugin(plugin) {
    const mapping = this.pluginMappings.get(plugin);
    if (!mapping) {
      return {
        original: plugin,
        status: 'unknown',
        message: 'No direct equivalent found',
      };
    }

    return {
      original: plugin,
      playwright: mapping.playwright,
      setup: mapping.setup.trim(),
      config: mapping.config,
      status: 'converted',
    };
  }

  /**
   * Generate plugin output
   * @param {Object[]} conversions - Array of plugin conversions
   * @returns {Object} - Final plugin output
   */
  generatePluginOutput(conversions) {
    // Combine configurations
    const combinedConfig = conversions.reduce((config, conversion) => {
      if (conversion.status === 'converted' && conversion.config) {
        return this.mergeConfigs(config, conversion.config);
      }
      return config;
    }, {});

    // Generate setup code
    const setupCode = conversions
      .filter((c) => c.status === 'converted')
      .map((c) => c.setup)
      .join('\n\n');

    // Generate imports
    const imports = conversions
      .filter((c) => c.status === 'converted')
      .map((c) => `import { test, expect } from '${c.playwright}';`)
      .join('\n');

    return {
      imports,
      setup: setupCode,
      config: combinedConfig,
      conversions: conversions.map((c) => ({
        original: c.original,
        status: c.status,
        playwright: c.playwright,
      })),
    };
  }

  /**
   * Merge plugin configurations
   * @param {Object} config1 - First configuration
   * @param {Object} config2 - Second configuration
   * @returns {Object} - Merged configuration
   */
  mergeConfigs(config1, config2) {
    const merged = { ...config1 };

    for (const [key, value] of Object.entries(config2)) {
      if (typeof value === 'object' && !Array.isArray(value)) {
        merged[key] = this.mergeConfigs(merged[key] || {}, value);
      } else {
        merged[key] = value;
      }
    }

    return merged;
  }

  /**
   * Get plugin mapping information
   * @param {string} plugin - Plugin name
   * @returns {Object|null} - Plugin mapping information
   */
  getPluginInfo(plugin) {
    return this.pluginMappings.get(plugin) || null;
  }

  /**
   * Get plugins by category
   * @param {string} category - Category name
   * @returns {string[]} - Array of plugin names
   */
  getPluginsByCategory(category) {
    return this.categories[category] || [];
  }

  /**
   * Check if a plugin has a known conversion
   * @param {string} plugin - Plugin name
   * @returns {boolean} - Whether plugin can be converted
   */
  canConvert(plugin) {
    return this.pluginMappings.has(plugin);
  }
}
