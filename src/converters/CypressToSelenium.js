import { BaseConverter } from '../core/BaseConverter.js';
import { PatternEngine } from '../core/PatternEngine.js';

/**
 * Converts Cypress tests to Selenium WebDriver format
 */
export class CypressToSelenium extends BaseConverter {
  constructor(options = {}) {
    super(options);
    this.sourceFramework = 'cypress';
    this.targetFramework = 'selenium';
    this.engine = new PatternEngine();
    this.initializePatterns();
  }

  initializePatterns() {
    // Test structure patterns
    this.engine.registerPatterns('structure', {
      'describe\\(([^,\n]+),': 'describe($1,',
      'it\\(([^,\n]+),\\s*(?:async\\s*)?\\(\\)\\s*=>': 'it($1, async function()',
      'before\\(': 'beforeAll(',
      'after\\(': 'afterAll(',
      'beforeEach\\(': 'beforeEach(',
      'afterEach\\(': 'afterEach('
    });

    // Navigation patterns
    this.engine.registerPatterns('navigation', {
      'cy\\.visit\\(([^)]+)\\)': 'await driver.get($1)',
      'cy\\.go\\([\'"]back[\'"]\\)': 'await driver.navigate().back()',
      'cy\\.go\\([\'"]forward[\'"]\\)': 'await driver.navigate().forward()',
      'cy\\.reload\\(\\)': 'await driver.navigate().refresh()',
      'cy\\.url\\(\\)': 'await driver.getCurrentUrl()',
      'cy\\.title\\(\\)': 'await driver.getTitle()'
    });

    // Selector patterns
    this.engine.registerPatterns('selectors', {
      'cy\\.get\\(([^)]+)\\)': 'await driver.findElement(By.css($1))',
      'cy\\.get\\([\'"]#([^\'\"]+)[\'"]\\)': 'await driver.findElement(By.id("$1"))',
      'cy\\.contains\\(([^)]+)\\)': 'await driver.findElement(By.xpath(`//*[contains(text(),$1)]`))',
      '\\.find\\(([^)]+)\\)': '.findElement(By.css($1))',
      '\\.first\\(\\)': '[0]',
      '\\.last\\(\\)': '.slice(-1)[0]',
      '\\.eq\\((\\d+)\\)': '[$1]'
    });

    // Interaction patterns
    this.engine.registerPatterns('interactions', {
      '\\.type\\(([^)]+)\\)': '.sendKeys($1)',
      '\\.click\\(\\)': '.click()',
      '\\.clear\\(\\)': '.clear()',
      '\\.check\\(\\)': '.click()',
      '\\.uncheck\\(\\)': '.click()',
      '\\.focus\\(\\)': '.click()'
    });

    // Assertion patterns
    this.engine.registerPatterns('assertions', {
      '\\.should\\([\'"]be\\.visible[\'"]\\)': '; expect(await element.isDisplayed()).toBe(true)',
      '\\.should\\([\'"]not\\.be\\.visible[\'"]\\)': '; expect(await element.isDisplayed()).toBe(false)',
      '\\.should\\([\'"]have\\.text[\'"],\\s*([^)]+)\\)': '; expect(await element.getText()).toBe($1)',
      '\\.should\\([\'"]contain[\'"],\\s*([^)]+)\\)': '; expect(await element.getText()).toContain($1)',
      '\\.should\\([\'"]have\\.value[\'"],\\s*([^)]+)\\)': '; expect(await element.getAttribute("value")).toBe($1)',
      '\\.should\\([\'"]be\\.checked[\'"]\\)': '; expect(await element.isSelected()).toBe(true)',
      '\\.should\\([\'"]be\\.disabled[\'"]\\)': '; expect(await element.isEnabled()).toBe(false)',
      '\\.should\\([\'"]be\\.enabled[\'"]\\)': '; expect(await element.isEnabled()).toBe(true)'
    });

    // Wait patterns
    this.engine.registerPatterns('waits', {
      'cy\\.wait\\((\\d+)\\)': 'await driver.sleep($1)'
    });
  }

  async convert(content, options = {}) {
    let result = content;

    // Convert Cypress commands to Selenium
    result = this.convertCypressCommands(result);

    // Convert test structure
    result = this.convertTestStructure(result);

    // Transform test callbacks
    result = this.transformTestStructure(result);

    // Add imports and setup/teardown
    const imports = this.getImports([]);
    const setup = this.getSetupTeardown();
    result = imports.join('\n') + '\n' + setup + '\n' + result;

    // Clean up
    result = this.cleanupOutput(result);

    this.stats.conversions++;
    return result;
  }

  /**
   * Convert Cypress commands to Selenium equivalents
   * @param {string} content - Content to convert
   * @returns {string}
   */
  convertCypressCommands(content) {
    let result = content;

    // Convert assertions with elements - use inline expressions to avoid duplicate variable names
    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]be\.visible['"]\)/g,
      'expect(await (await driver.findElement(By.css($1))).isDisplayed()).toBe(true)'
    );

    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]not\.be\.visible['"]\)/g,
      'expect(await (await driver.findElement(By.css($1))).isDisplayed()).toBe(false)'
    );

    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]exist['"]\)/g,
      'expect((await driver.findElements(By.css($1))).length).toBeGreaterThan(0)'
    );

    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]not\.exist['"]\)/g,
      'expect((await driver.findElements(By.css($1))).length).toBe(0)'
    );

    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]have\.text['"],\s*([^)]+)\)/g,
      'expect(await (await driver.findElement(By.css($1))).getText()).toBe($2)'
    );

    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]contain['"],\s*([^)]+)\)/g,
      'expect(await (await driver.findElement(By.css($1))).getText()).toContain($2)'
    );

    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]have\.value['"],\s*([^)]+)\)/g,
      'expect(await (await driver.findElement(By.css($1))).getAttribute("value")).toBe($2)'
    );

    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]have\.class['"],\s*([^)]+)\)/g,
      'expect(await (await driver.findElement(By.css($1))).getAttribute("class")).toContain($2)'
    );

    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]be\.checked['"]\)/g,
      'expect(await (await driver.findElement(By.css($1))).isSelected()).toBe(true)'
    );

    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]be\.disabled['"]\)/g,
      'expect(await (await driver.findElement(By.css($1))).isEnabled()).toBe(false)'
    );

    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]be\.enabled['"]\)/g,
      'expect(await (await driver.findElement(By.css($1))).isEnabled()).toBe(true)'
    );

    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]have\.length['"],\s*(\d+)\)/g,
      'expect((await driver.findElements(By.css($1))).length).toBe($2)'
    );

    // Convert interactions
    result = result.replace(
      /cy\.get\(([^)]+)\)\.type\(([^)]+)\)/g,
      'await driver.findElement(By.css($1)).sendKeys($2)'
    );

    result = result.replace(
      /cy\.get\(([^)]+)\)\.click\(\)/g,
      'await driver.findElement(By.css($1)).click()'
    );

    result = result.replace(
      /cy\.get\(([^)]+)\)\.clear\(\)/g,
      'await driver.findElement(By.css($1)).clear()'
    );

    result = result.replace(
      /cy\.get\(([^)]+)\)\.check\(\)/g,
      'const checkbox = await driver.findElement(By.css($1));\n      if (!(await checkbox.isSelected())) await checkbox.click()'
    );

    result = result.replace(
      /cy\.get\(([^)]+)\)\.uncheck\(\)/g,
      'const checkbox = await driver.findElement(By.css($1));\n      if (await checkbox.isSelected()) await checkbox.click()'
    );

    result = result.replace(
      /cy\.get\(([^)]+)\)\.select\(([^)]+)\)/g,
      'const select = await driver.findElement(By.css($1));\n      await select.findElement(By.css(`option[value=${$2}]`)).click()'
    );

    result = result.replace(
      /cy\.contains\(([^)]+)\)\.click\(\)/g,
      "await driver.findElement(By.xpath(`//*[contains(text(),$1)]`)).click()"
    );

    // Convert navigation
    result = result.replace(
      /cy\.visit\(([^)]+)\)/g,
      'await driver.get($1)'
    );

    result = result.replace(/cy\.reload\(\)/g, 'await driver.navigate().refresh()');
    result = result.replace(/cy\.go\(['"]back['"]\)/g, 'await driver.navigate().back()');
    result = result.replace(/cy\.go\(['"]forward['"]\)/g, 'await driver.navigate().forward()');

    // Convert URL assertions
    result = result.replace(
      /cy\.url\(\)\.should\(['"]include['"],\s*([^)]+)\)/g,
      'expect(await driver.getCurrentUrl()).toContain($1)'
    );

    result = result.replace(
      /cy\.url\(\)\.should\(['"]eq['"],\s*([^)]+)\)/g,
      'expect(await driver.getCurrentUrl()).toBe($1)'
    );

    // Convert title assertions
    result = result.replace(
      /cy\.title\(\)\.should\(['"]eq['"],\s*([^)]+)\)/g,
      'expect(await driver.getTitle()).toBe($1)'
    );

    // Convert waits
    result = result.replace(
      /cy\.wait\((\d+)\)/g,
      'await driver.sleep($1)'
    );

    // Convert storage/cookies
    result = result.replace(
      /cy\.clearCookies\(\)/g,
      'await driver.manage().deleteAllCookies()'
    );

    result = result.replace(
      /cy\.clearLocalStorage\(\)/g,
      'await driver.executeScript("localStorage.clear()")'
    );

    // Convert nested selector chains: cy.get().find().click()
    result = result.replace(
      /cy\.get\(([^)]+)\)\.find\(([^)]+)\)\.click\(\)/g,
      'await (await driver.findElement(By.css($1))).findElement(By.css($2)).click()'
    );

    // Convert chained element methods
    result = result.replace(
      /cy\.get\(([^)]+)\)\.first\(\)\.click\(\)/g,
      'await (await driver.findElements(By.css($1)))[0].click()'
    );

    result = result.replace(
      /cy\.get\(([^)]+)\)\.last\(\)\.click\(\)/g,
      '(await driver.findElements(By.css($1))).slice(-1)[0].click()'
    );

    result = result.replace(
      /cy\.get\(([^)]+)\)\.eq\((\d+)\)\.click\(\)/g,
      'await (await driver.findElements(By.css($1)))[$2].click()'
    );

    // Convert cy.go with numeric argument
    result = result.replace(
      /cy\.go\((-?\d+)\)/g,
      'await driver.navigate().back() /* go($1) */'
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

    // Convert describe blocks (keep as-is for Jest/Mocha)
    result = result.replace(/describe\.only\(/g, 'describe.only(');
    result = result.replace(/describe\.skip\(/g, 'describe.skip(');

    // Convert it blocks
    result = result.replace(/it\.only\(/g, 'it.only(');
    result = result.replace(/it\.skip\(/g, 'it.skip(');

    // Make test callbacks async
    result = result.replace(
      /it\(([^,\n]+),\s*\(\)\s*=>\s*\{/g,
      'it($1, async () => {'
    );

    result = result.replace(
      /it\(([^,\n]+),\s*function\s*\(\)\s*\{/g,
      'it($1, async function() {'
    );

    // Convert hooks to async
    result = result.replace(
      /before\(\s*\(\)\s*=>\s*\{/g,
      'beforeAll(async () => {'
    );

    result = result.replace(
      /after\(\s*\(\)\s*=>\s*\{/g,
      'afterAll(async () => {'
    );

    result = result.replace(
      /beforeEach\(\s*\(\)\s*=>\s*\{/g,
      'beforeEach(async () => {'
    );

    result = result.replace(
      /afterEach\(\s*\(\)\s*=>\s*\{/g,
      'afterEach(async () => {'
    );

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

  transformTestStructure(content) {
    // Add driver variable reference in tests
    content = content.replace(
      /it\(([^,\n]+),\s*async\s*function\(\)\s*\{/g,
      'it($1, async function() {\n    const driver = this.driver;'
    );
    return content;
  }

  getSetupTeardown() {
    return `
let driver;

beforeAll(async () => {
  driver = await new Builder().forBrowser('chrome').build();
});

afterAll(async () => {
  await driver.quit();
});
`;
  }

  detectTestTypes(content) {
    return ['e2e'];
  }

  getImports(testTypes) {
    return [
      "const { Builder, By, Key, until } = require('selenium-webdriver');",
      "const { expect } = require('@jest/globals');"
    ];
  }

  async convertConfig(configPath, options = {}) {
    return `// Selenium WebDriver configuration
// Converted from Cypress config

module.exports = {
  capabilities: {
    browserName: 'chrome'
  },
  baseUrl: 'http://localhost:3000',
  timeout: 30000,
  implicit_wait: 10000
};
`;
  }
}

export default CypressToSelenium;
