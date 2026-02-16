import { BaseConverter } from "../core/BaseConverter.js";
import { PatternEngine } from "../core/PatternEngine.js";
import { directMappings as navMappings } from "../patterns/commands/navigation.js";
import { directMappings as selectorMappings } from "../patterns/commands/selectors.js";
import { directMappings as interactionMappings } from "../patterns/commands/interactions.js";
import { directMappings as assertionMappings } from "../patterns/commands/assertions.js";
import { directMappings as waitMappings } from "../patterns/commands/waits.js";

/**
 * Converts Cypress tests to Playwright format
 */
export class CypressToPlaywright extends BaseConverter {
  constructor(options = {}) {
    super(options);
    this.sourceFramework = "cypress";
    this.targetFramework = "playwright";
    this.engine = new PatternEngine();
    this.initializePatterns();
  }

  /**
   * Initialize conversion patterns
   */
  initializePatterns() {
    const direction = "cypress-playwright";

    // Register all pattern categories
    this.engine.registerPatterns("navigation", navMappings[direction] || {});
    this.engine.registerPatterns(
      "selectors",
      selectorMappings[direction] || {},
    );
    this.engine.registerPatterns(
      "interactions",
      interactionMappings[direction] || {},
    );
    this.engine.registerPatterns(
      "assertions",
      assertionMappings[direction] || {},
    );
    this.engine.registerPatterns("waits", waitMappings[direction] || {});

    // Test structure patterns
    this.engine.registerPatterns("structure", {
      "describe\\(": "test.describe(",
      "it\\(": "test(",
      "before\\(": "test.beforeAll(",
      "after\\(": "test.afterAll(",
      "beforeEach\\(": "test.beforeEach(",
      "afterEach\\(": "test.afterEach(",
      "context\\(": "test.describe(",
      "specify\\(": "test(",
      "it\\.only\\(": "test.only(",
      "it\\.skip\\(": "test.skip(",
      "describe\\.only\\(": "test.describe.only(",
      "describe\\.skip\\(": "test.describe.skip(",
    });

    // Core command patterns
    this.engine.registerPatterns("commands", {
      "cy\\.visit\\(": "await page.goto(",
      "cy\\.get\\(": "page.locator(",
      "cy\\.contains\\(": "page.getByText(",
      "cy\\.find\\(": ".locator(",
      "cy\\.focused\\(\\)": 'page.locator(":focus")',
      "cy\\.root\\(\\)": 'page.locator("html")',
      "cy\\.document\\(\\)": "page",
      "cy\\.window\\(\\)": "page",
      "cy\\.viewport\\(": "await page.setViewportSize(",
      "cy\\.screenshot\\(": "await page.screenshot(",
      "cy\\.reload\\(\\)": "await page.reload()",
      "cy\\.go\\(['\"]back['\"]\\)": "await page.goBack()",
      "cy\\.go\\(['\"]forward['\"]\\)": "await page.goForward()",
      "cy\\.url\\(\\)": "page.url()",
      "cy\\.title\\(\\)": "await page.title()",
      "cy\\.clearCookies\\(\\)": "await context.clearCookies()",
      "cy\\.clearLocalStorage\\(\\)":
        "await page.evaluate(() => localStorage.clear())",
      "cy\\.log\\(": "console.log(",
      "cy\\.pause\\(\\)": "// await page.pause() // Uncomment for debugging",
      "cy\\.debug\\(\\)": "// debugger; // Uncomment for debugging",
    });

    // Interaction patterns
    this.engine.registerPatterns("interactions", {
      "\\.type\\(": ".fill(",
      "\\.click\\(\\)": ".click()",
      "\\.dblclick\\(\\)": ".dblclick()",
      "\\.rightclick\\(\\)": '.click({ button: "right" })',
      "\\.check\\(\\)": ".check()",
      "\\.uncheck\\(\\)": ".uncheck()",
      "\\.select\\(": ".selectOption(",
      "\\.clear\\(\\)": ".clear()",
      "\\.focus\\(\\)": ".focus()",
      "\\.blur\\(\\)": ".blur()",
      "\\.trigger\\(['\"]mouseover['\"]\\)": ".hover()",
      "\\.trigger\\(['\"]mouseenter['\"]\\)": ".hover()",
      "\\.scrollIntoView\\(\\)": ".scrollIntoViewIfNeeded()",
      "\\.selectFile\\(": ".setInputFiles(",
      "\\.attachFile\\(": ".setInputFiles(",
    });

    // Assertion patterns
    this.engine.registerPatterns("assertions", {
      "\\.should\\(['\"]be\\.visible['\"]\\)":
        "); await expect(element).toBeVisible()",
      "\\.should\\(['\"]not\\.be\\.visible['\"]\\)":
        "); await expect(element).toBeHidden()",
      "\\.should\\(['\"]exist['\"]\\)":
        "); await expect(element).toBeAttached()",
      "\\.should\\(['\"]not\\.exist['\"]\\)":
        "); await expect(element).not.toBeAttached()",
      "\\.should\\(['\"]have\\.text['\"],\\s*([^)]+)\\)":
        "); await expect(element).toHaveText($1)",
      "\\.should\\(['\"]contain['\"],\\s*([^)]+)\\)":
        "); await expect(element).toContainText($1)",
      "\\.should\\(['\"]have\\.value['\"],\\s*([^)]+)\\)":
        "); await expect(element).toHaveValue($1)",
      "\\.should\\(['\"]have\\.attr['\"],\\s*([^,\n]+),?\\s*([^)]*)\\)":
        "); await expect(element).toHaveAttribute($1, $2)",
      "\\.should\\(['\"]have\\.class['\"],\\s*([^)]+)\\)":
        "); await expect(element).toHaveClass($1)",
      "\\.should\\(['\"]be\\.checked['\"]\\)":
        "); await expect(element).toBeChecked()",
      "\\.should\\(['\"]be\\.disabled['\"]\\)":
        "); await expect(element).toBeDisabled()",
      "\\.should\\(['\"]be\\.enabled['\"]\\)":
        "); await expect(element).toBeEnabled()",
      "\\.should\\(['\"]have\\.length['\"],\\s*([^)]+)\\)":
        "); await expect(element).toHaveCount($1)",
    });

    // Traversal patterns
    this.engine.registerPatterns("traversal", {
      "\\.first\\(\\)": ".first()",
      "\\.last\\(\\)": ".last()",
      "\\.eq\\((\\d+)\\)": ".nth($1)",
      "\\.parent\\(\\)": '.locator("..")',
      "\\.children\\(\\)": '.locator("> *")',
      "\\.siblings\\(\\)": '.locator("~ *")',
      "\\.next\\(\\)": '.locator("+ *")',
      "\\.prev\\(\\)": '.locator(":prev")',
    });

    // Network patterns
    this.engine.registerPatterns("network", {
      "cy\\.intercept\\(": "await page.route(",
      "cy\\.request\\(": "await request.fetch(",
      "cy\\.wait\\(['\"]@":
        'await page.waitForResponse(response => response.url().includes("',
    });
  }

  /**
   * Convert Cypress test content to Playwright
   * @param {string} content - Cypress test content
   * @param {Object} options - Conversion options
   * @returns {Promise<string>} - Playwright test content
   */
  async convert(content, _options = {}) {
    let result = content;

    // Detect test types
    const testTypes = this.detectTestTypes(content);

    // Convert Cypress commands to Playwright (order matters!)
    result = this.convertCypressCommands(result);

    // Convert test structure (describe, it, hooks)
    result = this.convertTestStructure(result);

    // Make test callbacks async with page parameter
    result = this.transformTestCallbacks(result, testTypes);

    // Add required imports
    const imports = this.getImports(testTypes);

    // Clean up and format
    result = this.cleanupOutput(result);

    // Combine imports and content
    result = imports.join("\n") + "\n\n" + result;

    this.stats.conversions++;
    return result;
  }

  /**
   * Convert Cypress commands to Playwright equivalents
   * @param {string} content - Content to convert
   * @returns {string}
   */
  convertCypressCommands(content) {
    let result = content;

    // Convert cy.get().should() chains - assertions on elements
    // Note: Using [^()\n]+ to prevent ReDoS attacks from nested parens
    // Handle .should('be.visible')
    result = result.replace(
      /cy\.get\(([^()\n]+)\)\.should\(['"]be\.visible['"]\)/g,
      "await expect(page.locator($1)).toBeVisible()",
    );

    // Handle .should('not.be.visible')
    result = result.replace(
      /cy\.get\(([^()\n]+)\)\.should\(['"]not\.be\.visible['"]\)/g,
      "await expect(page.locator($1)).toBeHidden()",
    );

    // Handle .should('exist')
    result = result.replace(
      /cy\.get\(([^()\n]+)\)\.should\(['"]exist['"]\)/g,
      "await expect(page.locator($1)).toBeAttached()",
    );

    // Handle .should('not.exist')
    result = result.replace(
      /cy\.get\(([^()\n]+)\)\.should\(['"]not\.exist['"]\)/g,
      "await expect(page.locator($1)).not.toBeAttached()",
    );

    // Handle .should('have.text', value)
    result = result.replace(
      /cy\.get\(([^()\n]+)\)\.should\(['"]have\.text['"],\s*([^()\n]+)\)/g,
      "await expect(page.locator($1)).toHaveText($2)",
    );

    // Handle .should('contain', value)
    result = result.replace(
      /cy\.get\(([^()\n]+)\)\.should\(['"]contain['"],\s*([^()\n]+)\)/g,
      "await expect(page.locator($1)).toContainText($2)",
    );

    // Handle .should('have.value', value)
    result = result.replace(
      /cy\.get\(([^()\n]+)\)\.should\(['"]have\.value['"],\s*([^()\n]+)\)/g,
      "await expect(page.locator($1)).toHaveValue($2)",
    );

    // Handle .should('have.class', value)
    result = result.replace(
      /cy\.get\(([^()\n]+)\)\.should\(['"]have\.class['"],\s*([^()\n]+)\)/g,
      "await expect(page.locator($1)).toHaveClass($2)",
    );

    // Handle .should('be.checked')
    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]be\.checked['"]\)/g,
      "await expect(page.locator($1)).toBeChecked()",
    );

    // Handle .should('be.disabled')
    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]be\.disabled['"]\)/g,
      "await expect(page.locator($1)).toBeDisabled()",
    );

    // Handle .should('be.enabled')
    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]be\.enabled['"]\)/g,
      "await expect(page.locator($1)).toBeEnabled()",
    );

    // Handle .should('have.length', n)
    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]have\.length['"],\s*(\d+)\)/g,
      "await expect(page.locator($1)).toHaveCount($2)",
    );

    // Handle .should('have.attr', name, value)
    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]have\.attr['"],\s*([^,\n]+),\s*([^)]+)\)/g,
      "await expect(page.locator($1)).toHaveAttribute($2, $3)",
    );

    // Convert cy.get().type() chains
    result = result.replace(
      /cy\.get\(([^)]+)\)\.type\(([^)]+)\)/g,
      "await page.locator($1).fill($2)",
    );

    // Convert cy.get().click() chains
    result = result.replace(
      /cy\.get\(([^)]+)\)\.click\(\)/g,
      "await page.locator($1).click()",
    );

    // Convert cy.get().dblclick() chains
    result = result.replace(
      /cy\.get\(([^)]+)\)\.dblclick\(\)/g,
      "await page.locator($1).dblclick()",
    );

    // Convert cy.get().check() chains
    result = result.replace(
      /cy\.get\(([^)]+)\)\.check\(\)/g,
      "await page.locator($1).check()",
    );

    // Convert cy.get().uncheck() chains
    result = result.replace(
      /cy\.get\(([^)]+)\)\.uncheck\(\)/g,
      "await page.locator($1).uncheck()",
    );

    // Convert cy.get().select() chains
    result = result.replace(
      /cy\.get\(([^)]+)\)\.select\(([^)]+)\)/g,
      "await page.locator($1).selectOption($2)",
    );

    // Convert cy.get().clear() chains
    result = result.replace(
      /cy\.get\(([^)]+)\)\.clear\(\)/g,
      "await page.locator($1).clear()",
    );

    // Convert cy.get().focus() chains
    result = result.replace(
      /cy\.get\(([^)]+)\)\.focus\(\)/g,
      "await page.locator($1).focus()",
    );

    // Convert cy.get().blur() chains
    result = result.replace(
      /cy\.get\(([^)]+)\)\.blur\(\)/g,
      "await page.locator($1).blur()",
    );

    // Convert cy.contains()
    result = result.replace(
      /cy\.contains\(([^)]+)\)\.click\(\)/g,
      "await page.getByText($1).click()",
    );

    result = result.replace(/cy\.contains\(([^)]+)\)/g, "page.getByText($1)");

    // Convert cy.visit()
    result = result.replace(/cy\.visit\(([^)]+)\)/g, "await page.goto($1)");

    // Convert cy.url()
    result = result.replace(
      /cy\.url\(\)\.should\(['"]include['"],\s*([^)]+)\)/g,
      "await expect(page).toHaveURL(new RegExp($1))",
    );

    result = result.replace(
      /cy\.url\(\)\.should\(['"]eq['"],\s*([^)]+)\)/g,
      "await expect(page).toHaveURL($1)",
    );

    // Convert cy.title()
    result = result.replace(
      /cy\.title\(\)\.should\(['"]eq['"],\s*([^)]+)\)/g,
      "await expect(page).toHaveTitle($1)",
    );

    result = result.replace(
      /cy\.title\(\)\.should\(['"]include['"],\s*([^)]+)\)/g,
      "await expect(page).toHaveTitle(new RegExp($1))",
    );

    // Convert cy.wait() for aliases (network)
    result = result.replace(
      /cy\.wait\(['"]@([^'"]+)['"]\)/g,
      'await page.waitForResponse(response => response.url().includes("$1"))',
    );

    // Convert cy.wait() for time
    result = result.replace(
      /cy\.wait\((\d+)\)/g,
      "await page.waitForTimeout($1)",
    );

    // Convert cy.reload()
    result = result.replace(/cy\.reload\(\)/g, "await page.reload()");

    // Convert cy.go('back')
    result = result.replace(/cy\.go\(['"]back['"]\)/g, "await page.goBack()");

    // Convert cy.go('forward')
    result = result.replace(
      /cy\.go\(['"]forward['"]\)/g,
      "await page.goForward()",
    );

    // Convert cy.viewport()
    result = result.replace(
      /cy\.viewport\((\d+),\s*(\d+)\)/g,
      "await page.setViewportSize({ width: $1, height: $2 })",
    );

    // Convert cy.screenshot()
    result = result.replace(
      /cy\.screenshot\(([^)]*)\)/g,
      "await page.screenshot({ path: $1 })",
    );

    // Convert cy.clearCookies()
    result = result.replace(
      /cy\.clearCookies\(\)/g,
      "await context.clearCookies()",
    );

    // Convert cy.clearLocalStorage()
    result = result.replace(
      /cy\.clearLocalStorage\(\)/g,
      "await page.evaluate(() => localStorage.clear())",
    );

    // Convert cy.log()
    result = result.replace(/cy\.log\(([^)]+)\)/g, "console.log($1)");

    // Convert cy.intercept()
    result = result.replace(
      /cy\.intercept\(([^,\n]+),\s*([^)]+)\)\.as\(['"]([^'"]+)['"]\)/g,
      "await page.route($1, route => route.fulfill($2))",
    );

    // Convert cy.get().check() with options (e.g., { force: true })
    result = result.replace(
      /cy\.get\(([^)]+)\)\.check\(\{[^{}\n]*\}\)/g,
      "await page.locator($1).check()",
    );

    // Convert cy.go() with numeric arguments
    result = result.replace(
      /cy\.go\((-?\d+)\)/g,
      "await page.goBack() /* go($1) */",
    );

    // Convert cy.reload() with arguments
    result = result.replace(/cy\.reload\([^)]+\)/g, "await page.reload()");

    // Convert cy.get().first().click() chains
    result = result.replace(
      /cy\.get\(([^)]+)\)\.first\(\)\.click\(\)/g,
      "await page.locator($1).first().click()",
    );

    // Convert cy.get().last().click() chains
    result = result.replace(
      /cy\.get\(([^)]+)\)\.last\(\)\.click\(\)/g,
      "await page.locator($1).last().click()",
    );

    // Convert cy.get().eq(n).click() chains
    result = result.replace(
      /cy\.get\(([^)]+)\)\.eq\((\d+)\)\.click\(\)/g,
      "await page.locator($1).nth($2).click()",
    );

    // Convert cy.get().first() (no click)
    result = result.replace(
      /cy\.get\(([^)]+)\)\.first\(\)/g,
      "page.locator($1).first()",
    );

    // Convert cy.get().last() (no click)
    result = result.replace(
      /cy\.get\(([^)]+)\)\.last\(\)/g,
      "page.locator($1).last()",
    );

    // Convert cy.get().eq(n) (no click)
    result = result.replace(
      /cy\.get\(([^)]+)\)\.eq\((\d+)\)/g,
      "page.locator($1).nth($2)",
    );

    return result;
  }

  /**
   * Convert test structure (describe, it, hooks)
   * @param {string} content - Content to convert
   * @returns {string}
   */
  convertTestStructure(content) {
    let result = content;

    // Convert describe blocks
    result = result.replace(/describe\.only\(/g, "test.describe.only(");
    result = result.replace(/describe\.skip\(/g, "test.describe.skip(");
    result = result.replace(/describe\(/g, "test.describe(");

    // Convert context (alias for describe)
    result = result.replace(/context\(/g, "test.describe(");

    // Convert it blocks
    result = result.replace(/it\.only\(/g, "test.only(");
    result = result.replace(/it\.skip\(/g, "test.skip(");
    result = result.replace(/specify\(/g, "test(");
    result = result.replace(/it\(/g, "test(");

    // Convert hooks
    result = result.replace(/before\(/g, "test.beforeAll(");
    result = result.replace(/after\(/g, "test.afterAll(");
    result = result.replace(/beforeEach\(/g, "test.beforeEach(");
    result = result.replace(/afterEach\(/g, "test.afterEach(");

    return result;
  }

  /**
   * Transform test callbacks to async with proper parameters
   * @param {string} content - Content to transform
   * @param {string[]} testTypes - Detected test types
   * @returns {string}
   */
  transformTestCallbacks(content, testTypes) {
    const params = testTypes.includes("api") ? "{ page, request }" : "{ page }";

    // Transform test callbacks
    // Note: Using [^,()\n]+ to prevent ReDoS by excluding nested parens and commas
    content = content.replace(
      /test\(([^,()\n]+),\s*(?:async\s*)?\(\s*\)\s*=>\s*\{/g,
      `test($1, async (${params}) => {`,
    );

    // Transform describe callbacks
    content = content.replace(
      /test\.describe\(([^,()\n]+),\s*(?:async\s*)?\(\s*\)\s*=>\s*\{/g,
      "test.describe($1, () => {",
    );

    // Transform hooks
    const hookParams = "{ page }";
    content = content.replace(
      /test\.(beforeAll|afterAll|beforeEach|afterEach)\(\s*(?:async\s*)?\(\s*\)\s*=>\s*\{/g,
      `test.$1(async (${hookParams}) => {`,
    );

    return content;
  }

  /**
   * Clean up and format output
   * @param {string} content - Content to clean
   * @returns {string}
   */
  cleanupOutput(content) {
    return (
      content
        // Remove double awaits
        .replace(/await\s+await/g, "await")
        // Fix any empty screenshot path args
        .replace(/screenshot\(\{ path: \s*\}\)/g, "screenshot()")
        // Clean up empty lines
        .replace(/\n{3,}/g, "\n\n")
        // Ensure proper line endings
        .trim() + "\n"
    );
  }

  /**
   * Detect test types from content
   * @param {string} content - Test content
   * @returns {string[]} - Array of detected test types
   */
  detectTestTypes(content) {
    const types = [];

    if (/cy\.request|cy\.intercept/.test(content)) types.push("api");
    if (/cy\.mount/.test(content)) types.push("component");
    if (/cy\.injectAxe|cy\.checkA11y/.test(content))
      types.push("accessibility");
    if (/cy\.screenshot|matchImageSnapshot/.test(content)) types.push("visual");
    if (/cy\.lighthouse|performance\./.test(content)) types.push("performance");
    if (/viewport|mobile/.test(content)) types.push("mobile");

    if (types.length === 0) types.push("e2e");

    return types;
  }

  /**
   * Get required imports for target framework
   * @param {string[]} testTypes - Detected test types
   * @returns {string[]} - Array of import statements
   */
  getImports(testTypes) {
    const imports = new Set([
      "import { test, expect } from '@playwright/test';",
    ]);

    if (testTypes.includes("api")) {
      imports.add("import { request } from '@playwright/test';");
    }

    if (testTypes.includes("component")) {
      imports.add("import { mount } from '@playwright/experimental-ct-react';");
    }

    if (testTypes.includes("accessibility")) {
      imports.add("import { injectAxe, checkA11y } from 'axe-playwright';");
    }

    return Array.from(imports);
  }

  /**
   * Convert Cypress config to Playwright config
   * @param {string} configPath - Path to Cypress config
   * @param {Object} options - Conversion options
   * @returns {Promise<string>} - Playwright config content
   */
  async convertConfig(configPath, _options = {}) {
    const fs = await import("fs/promises");
    const content = await fs.readFile(configPath, "utf8");

    let cypressConfig = {};

    if (configPath.endsWith(".json")) {
      cypressConfig = JSON.parse(content);
    } else {
      // Extract values using regex (safer than eval)
      const baseUrlMatch = content.match(/baseUrl:\s*['"]([^'"]+)['"]/);
      const viewportWidthMatch = content.match(/viewportWidth:\s*(\d+)/);
      const viewportHeightMatch = content.match(/viewportHeight:\s*(\d+)/);
      const videoMatch = content.match(/video:\s*(true|false)/);
      const screenshotMatch = content.match(
        /screenshotOnRunFailure:\s*(true|false)/,
      );
      const timeoutMatch = content.match(/defaultCommandTimeout:\s*(\d+)/);

      if (baseUrlMatch) cypressConfig.baseUrl = baseUrlMatch[1];
      if (viewportWidthMatch)
        cypressConfig.viewportWidth = parseInt(viewportWidthMatch[1]);
      if (viewportHeightMatch)
        cypressConfig.viewportHeight = parseInt(viewportHeightMatch[1]);
      if (videoMatch) cypressConfig.video = videoMatch[1] === "true";
      if (screenshotMatch)
        cypressConfig.screenshotOnFailure = screenshotMatch[1] === "true";
      if (timeoutMatch)
        cypressConfig.defaultCommandTimeout = parseInt(timeoutMatch[1]);
    }

    const playwrightConfig = {
      testDir: "./tests",
      timeout: cypressConfig.defaultCommandTimeout || 30000,
      expect: {
        timeout: cypressConfig.defaultCommandTimeout || 5000,
      },
      use: {
        baseURL: cypressConfig.baseUrl,
        viewport:
          cypressConfig.viewportWidth && cypressConfig.viewportHeight
            ? {
                width: cypressConfig.viewportWidth,
                height: cypressConfig.viewportHeight,
              }
            : { width: 1280, height: 720 },
        video: cypressConfig.video ? "on" : "off",
        screenshot: cypressConfig.screenshotOnFailure
          ? "only-on-failure"
          : "off",
        trace: "retain-on-failure",
      },
      projects: [
        { name: "chromium", use: { browserName: "chromium" } },
        { name: "firefox", use: { browserName: "firefox" } },
        { name: "webkit", use: { browserName: "webkit" } },
      ],
    };

    return `import { defineConfig } from '@playwright/test';

export default defineConfig(${JSON.stringify(playwrightConfig, null, 2)});
`;
  }
}

export default CypressToPlaywright;
