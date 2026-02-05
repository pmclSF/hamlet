import { BaseConverter } from '../core/BaseConverter.js';
import { PatternEngine } from '../core/PatternEngine.js';

/**
 * Converts Playwright tests to Cypress format
 */
export class PlaywrightToCypress extends BaseConverter {
  constructor(options = {}) {
    super(options);
    this.sourceFramework = 'playwright';
    this.targetFramework = 'cypress';
    this.engine = new PatternEngine();
    this.initializePatterns();
  }

  initializePatterns() {
    // Test structure patterns
    this.engine.registerPatterns('structure', {
      'test\\.describe\\(': 'describe(',
      'test\\.describe\\.only\\(': 'describe.only(',
      'test\\.describe\\.skip\\(': 'describe.skip(',
      'test\\(': 'it(',
      'test\\.only\\(': 'it.only(',
      'test\\.skip\\(': 'it.skip(',
      'test\\.beforeAll\\(': 'before(',
      'test\\.afterAll\\(': 'after(',
      'test\\.beforeEach\\(': 'beforeEach(',
      'test\\.afterEach\\(': 'afterEach('
    });

    // Navigation patterns
    this.engine.registerPatterns('navigation', {
      'await page\\.goto\\(([^)]+)\\)': 'cy.visit($1)',
      'await page\\.goBack\\(\\)': "cy.go('back')",
      'await page\\.goForward\\(\\)': "cy.go('forward')",
      'await page\\.reload\\(\\)': 'cy.reload()',
      'page\\.url\\(\\)': 'cy.url()',
      'await page\\.title\\(\\)': 'cy.title()'
    });

    // Selector patterns
    this.engine.registerPatterns('selectors', {
      'page\\.locator\\(([^)]+)\\)': 'cy.get($1)',
      'page\\.getByText\\(([^)]+)\\)': 'cy.contains($1)',
      'page\\.getByRole\\(([^)]+)\\)': 'cy.get(`[role=$1]`)',
      'page\\.getByTestId\\(([^)]+)\\)': 'cy.get(`[data-testid=$1]`)',
      'page\\.getByLabel\\(([^)]+)\\)': 'cy.get(`[aria-label=$1]`)',
      'page\\.getByPlaceholder\\(([^)]+)\\)': 'cy.get(`[placeholder=$1]`)',
      '\\.locator\\(([^)]+)\\)': '.find($1)',
      '\\.first\\(\\)': '.first()',
      '\\.last\\(\\)': '.last()',
      '\\.nth\\((\\d+)\\)': '.eq($1)'
    });

    // Interaction patterns
    this.engine.registerPatterns('interactions', {
      '\\.fill\\(([^)]+)\\)': '.type($1)',
      '\\.click\\(\\)': '.click()',
      '\\.dblclick\\(\\)': '.dblclick()',
      '\\.click\\(\\{\\s*button:\\s*[\'"]right[\'"]\\s*\\}\\)': '.rightclick()',
      '\\.check\\(\\)': '.check()',
      '\\.uncheck\\(\\)': '.uncheck()',
      '\\.selectOption\\(([^)]+)\\)': '.select($1)',
      '\\.clear\\(\\)': '.clear()',
      '\\.focus\\(\\)': '.focus()',
      '\\.blur\\(\\)': '.blur()',
      '\\.hover\\(\\)': '.trigger("mouseover")',
      '\\.scrollIntoViewIfNeeded\\(\\)': '.scrollIntoView()',
      '\\.setInputFiles\\(([^)]+)\\)': '.selectFile($1)'
    });

    // Assertion patterns
    this.engine.registerPatterns('assertions', {
      'await expect\\(([^)]+)\\)\\.toBeVisible\\(\\)': '$1.should("be.visible")',
      'await expect\\(([^)]+)\\)\\.toBeHidden\\(\\)': '$1.should("not.be.visible")',
      'await expect\\(([^)]+)\\)\\.toBeAttached\\(\\)': '$1.should("exist")',
      'await expect\\(([^)]+)\\)\\.not\\.toBeAttached\\(\\)': '$1.should("not.exist")',
      'await expect\\(([^)]+)\\)\\.toHaveText\\(([^)]+)\\)': '$1.should("have.text", $2)',
      'await expect\\(([^)]+)\\)\\.toContainText\\(([^)]+)\\)': '$1.should("contain", $2)',
      'await expect\\(([^)]+)\\)\\.toHaveValue\\(([^)]+)\\)': '$1.should("have.value", $2)',
      'await expect\\(([^)]+)\\)\\.toHaveAttribute\\(([^,\n]+),\\s*([^)]+)\\)': '$1.should("have.attr", $2, $3)',
      'await expect\\(([^)]+)\\)\\.toHaveClass\\(([^)]+)\\)': '$1.should("have.class", $2)',
      'await expect\\(([^)]+)\\)\\.toBeChecked\\(\\)': '$1.should("be.checked")',
      'await expect\\(([^)]+)\\)\\.toBeDisabled\\(\\)': '$1.should("be.disabled")',
      'await expect\\(([^)]+)\\)\\.toBeEnabled\\(\\)': '$1.should("be.enabled")',
      'await expect\\(([^)]+)\\)\\.toHaveCount\\(([^)]+)\\)': '$1.should("have.length", $2)',
      'await expect\\(page\\)\\.toHaveURL\\(([^)]+)\\)': 'cy.url().should("include", $1)',
      'await expect\\(page\\)\\.toHaveTitle\\(([^)]+)\\)': 'cy.title().should("eq", $1)'
    });

    // Wait patterns
    this.engine.registerPatterns('waits', {
      'await page\\.waitForTimeout\\((\\d+)\\)': 'cy.wait($1)',
      'await page\\.waitForSelector\\(([^)]+)\\)': 'cy.get($1)',
      'await page\\.waitForURL\\(([^)]+)\\)': 'cy.url().should("include", $1)',
      'await page\\.waitForLoadState\\([\'"]networkidle[\'"]\\)': 'cy.wait(1000)'
    });

    // Network patterns
    this.engine.registerPatterns('network', {
      'await page\\.route\\(([^,\n]+),': 'cy.intercept($1,',
      'await request\\.fetch\\(': 'cy.request('
    });
  }

  async convert(content, options = {}) {
    let result = content;

    // Remove Playwright imports
    result = result.replace(/import\s*\{[^{}\n]*\}\s*from\s*['"]@playwright\/test['"];?\n?/g, '');

    // Convert commands using explicit patterns (before removing await)
    result = this.convertPlaywrightCommands(result);

    // Remove remaining async/await
    result = this.removeAsyncAwait(result);

    // Transform test structure
    result = this.convertTestStructure(result);

    // Transform test callbacks
    result = this.transformTestCallbacks(result);

    // Add Cypress-specific setup
    const setup = this.getSetup();
    result = setup + result;

    // Clean up
    result = this.cleanupOutput(result);

    this.stats.conversions++;
    return result;
  }

  /**
   * Convert Playwright commands to Cypress equivalents
   * @param {string} content - Content to convert
   * @returns {string}
   */
  convertPlaywrightCommands(content) {
    let result = content;

    // Convert assertions first (before removing await)
    // await expect(page.locator(selector)).toBeVisible()
    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toBeVisible\(\)/g,
      "cy.get($1).should('be.visible')"
    );

    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toBeHidden\(\)/g,
      "cy.get($1).should('not.be.visible')"
    );

    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toBeAttached\(\)/g,
      "cy.get($1).should('exist')"
    );

    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.not\.toBeAttached\(\)/g,
      "cy.get($1).should('not.exist')"
    );

    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toHaveText\(([^)]+)\)/g,
      "cy.get($1).should('have.text', $2)"
    );

    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toContainText\(([^)]+)\)/g,
      "cy.get($1).should('contain', $2)"
    );

    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toHaveValue\(([^)]+)\)/g,
      "cy.get($1).should('have.value', $2)"
    );

    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toHaveClass\(([^)]+)\)/g,
      "cy.get($1).should('have.class', $2)"
    );

    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toBeChecked\(\)/g,
      "cy.get($1).should('be.checked')"
    );

    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toBeDisabled\(\)/g,
      "cy.get($1).should('be.disabled')"
    );

    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toBeEnabled\(\)/g,
      "cy.get($1).should('be.enabled')"
    );

    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toHaveCount\((\d+)\)/g,
      "cy.get($1).should('have.length', $2)"
    );

    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toHaveAttribute\(([^,\n]+),\s*([^)]+)\)/g,
      "cy.get($1).should('have.attr', $2, $3)"
    );

    // Convert page URL/title assertions
    result = result.replace(
      /await expect\(page\)\.toHaveURL\(([^)]+)\)/g,
      "cy.url().should('include', $1)"
    );

    result = result.replace(
      /await expect\(page\)\.toHaveTitle\(([^)]+)\)/g,
      "cy.title().should('eq', $1)"
    );

    // Convert interactions
    result = result.replace(
      /await page\.locator\(([^)]+)\)\.fill\(([^)]+)\)/g,
      'cy.get($1).type($2)'
    );

    result = result.replace(
      /await page\.locator\(([^)]+)\)\.click\(\)/g,
      'cy.get($1).click()'
    );

    result = result.replace(
      /await page\.locator\(([^)]+)\)\.dblclick\(\)/g,
      'cy.get($1).dblclick()'
    );

    result = result.replace(
      /await page\.locator\(([^)]+)\)\.check\(\)/g,
      'cy.get($1).check()'
    );

    result = result.replace(
      /await page\.locator\(([^)]+)\)\.uncheck\(\)/g,
      'cy.get($1).uncheck()'
    );

    result = result.replace(
      /await page\.locator\(([^)]+)\)\.selectOption\(([^)]+)\)/g,
      'cy.get($1).select($2)'
    );

    result = result.replace(
      /await page\.locator\(([^)]+)\)\.clear\(\)/g,
      'cy.get($1).clear()'
    );

    result = result.replace(
      /await page\.getByText\(([^)]+)\)\.click\(\)/g,
      'cy.contains($1).click()'
    );

    // Convert navigation
    result = result.replace(
      /await page\.goto\(([^)]+)\)/g,
      'cy.visit($1)'
    );

    result = result.replace(/await page\.reload\(\)/g, 'cy.reload()');
    result = result.replace(/await page\.goBack\(\)/g, "cy.go('back')");
    result = result.replace(/await page\.goForward\(\)/g, "cy.go('forward')");

    // Convert viewport
    result = result.replace(
      /await page\.setViewportSize\(\{\s*width:\s*(\d+),\s*height:\s*(\d+)\s*\}\)/g,
      'cy.viewport($1, $2)'
    );

    // Convert waits
    result = result.replace(
      /await page\.waitForTimeout\((\d+)\)/g,
      'cy.wait($1)'
    );

    result = result.replace(
      /await page\.waitForSelector\(([^)]+)\)/g,
      'cy.get($1)'
    );

    // Convert cookies/storage
    result = result.replace(
      /await context\.clearCookies\(\)/g,
      'cy.clearCookies()'
    );

    result = result.replace(
      /await page\.evaluate\(\(\) => localStorage\.clear\(\)\)/g,
      'cy.clearLocalStorage()'
    );

    // Convert console.log back to cy.log
    result = result.replace(/console\.log\(([^)]+)\)/g, 'cy.log($1)');

    // Convert page.locator().first().click() chains
    result = result.replace(
      /page\.locator\(([^)]+)\)\.first\(\)\.click\(\)/g,
      'cy.get($1).first().click()'
    );

    // Convert page.locator().last().click() chains
    result = result.replace(
      /page\.locator\(([^)]+)\)\.last\(\)\.click\(\)/g,
      'cy.get($1).last().click()'
    );

    // Convert page.locator().nth(n).click() chains
    result = result.replace(
      /page\.locator\(([^)]+)\)\.nth\((\d+)\)\.click\(\)/g,
      'cy.get($1).eq($2).click()'
    );

    // Convert page.locator().first() (standalone)
    result = result.replace(
      /page\.locator\(([^)]+)\)\.first\(\)/g,
      'cy.get($1).first()'
    );

    // Convert page.locator().last() (standalone)
    result = result.replace(
      /page\.locator\(([^)]+)\)\.last\(\)/g,
      'cy.get($1).last()'
    );

    // Convert page.locator().nth(n) (standalone)
    result = result.replace(
      /page\.locator\(([^)]+)\)\.nth\((\d+)\)/g,
      'cy.get($1).eq($2)'
    );

    return result;
  }

  /**
   * Convert test structure
   * @param {string} content - Content to convert
   * @returns {string}
   */
  convertTestStructure(content) {
    let result = content;

    // Convert describe
    result = result.replace(/test\.describe\.only\(/g, 'describe.only(');
    result = result.replace(/test\.describe\.skip\(/g, 'describe.skip(');
    result = result.replace(/test\.describe\(/g, 'describe(');

    // Convert test
    result = result.replace(/test\.only\(/g, 'it.only(');
    result = result.replace(/test\.skip\(/g, 'it.skip(');
    result = result.replace(/test\(/g, 'it(');

    // Convert hooks
    result = result.replace(/test\.beforeAll\(/g, 'before(');
    result = result.replace(/test\.afterAll\(/g, 'after(');
    result = result.replace(/test\.beforeEach\(/g, 'beforeEach(');
    result = result.replace(/test\.afterEach\(/g, 'afterEach(');

    return result;
  }

  /**
   * Clean up output
   * @param {string} content - Content to clean
   * @returns {string}
   */
  cleanupOutput(content) {
    return content
      // Remove empty lines
      .replace(/\n{3,}/g, '\n\n')
      // Trim
      .trim() + '\n';
  }

  removeAsyncAwait(content) {
    // Remove await keywords (Cypress handles async automatically)
    content = content.replace(/await\s+/g, '');

    // Convert async arrow functions to regular
    content = content.replace(/async\s*\(\s*\{[^}]+\}\s*\)\s*=>/g, '() =>');
    content = content.replace(/async\s*\(\s*\)\s*=>/g, '() =>');

    return content;
  }

  transformTestCallbacks(content) {
    // Remove page/request destructuring from test callbacks
    content = content.replace(
      /it\(([^,\n]+),\s*\(\s*\{[^}]+\}\s*\)\s*=>\s*\{/g,
      'it($1, () => {'
    );

    content = content.replace(
      /it\(([^,\n]+),\s*\(\s*\)\s*=>\s*\{/g,
      'it($1, () => {'
    );

    return content;
  }

  detectTestTypes(content) {
    const types = [];
    if (/request\.fetch/.test(content)) types.push('api');
    if (/mount\(/.test(content)) types.push('component');
    if (types.length === 0) types.push('e2e');
    return types;
  }

  getImports(testTypes) {
    return []; // Cypress doesn't need explicit imports for basic tests
  }

  getSetup() {
    return `/// <reference types="cypress" />

`;
  }

  async convertConfig(configPath, options = {}) {
    const fs = await import('fs/promises');
    const content = await fs.readFile(configPath, 'utf8');

    // Extract config from Playwright config
    let pwConfig = {};
    try {
      const match = content.match(/defineConfig\s*\(\s*({[\s\S]*})\s*\)/);
      if (match) {
        pwConfig = eval(`(${match[1]})`);
      }
    } catch (e) {
      // Use defaults
    }

    const cypressConfig = {
      e2e: {
        baseUrl: pwConfig.use?.baseURL || 'http://localhost:3000',
        viewportWidth: pwConfig.use?.viewport?.width || 1280,
        viewportHeight: pwConfig.use?.viewport?.height || 720,
        video: pwConfig.use?.video === 'on',
        screenshotOnRunFailure: pwConfig.use?.screenshot !== 'off',
        defaultCommandTimeout: pwConfig.timeout || 4000
      }
    };

    return `const { defineConfig } = require('cypress');

module.exports = defineConfig(${JSON.stringify(cypressConfig, null, 2)});
`;
  }
}

export default PlaywrightToCypress;
