import { BaseConverter } from '../core/BaseConverter.js';
import { PatternEngine } from '../core/PatternEngine.js';

/**
 * Converts Selenium WebDriver tests to Cypress format
 */
export class SeleniumToCypress extends BaseConverter {
  constructor(options = {}) {
    super(options);
    this.sourceFramework = 'selenium';
    this.targetFramework = 'cypress';
    this.engine = new PatternEngine();
    this.initializePatterns();
  }

  initializePatterns() {
    // Navigation patterns
    this.engine.registerPatterns('navigation', {
      'await driver\\.get\\(([^)]+)\\)': 'cy.visit($1)',
      'await driver\\.navigate\\(\\)\\.to\\(([^)]+)\\)': 'cy.visit($1)',
      'await driver\\.navigate\\(\\)\\.back\\(\\)': "cy.go('back')",
      'await driver\\.navigate\\(\\)\\.forward\\(\\)': "cy.go('forward')",
      'await driver\\.navigate\\(\\)\\.refresh\\(\\)': 'cy.reload()',
      'await driver\\.getCurrentUrl\\(\\)': 'cy.url()',
      'await driver\\.getTitle\\(\\)': 'cy.title()'
    });

    // Selector patterns
    this.engine.registerPatterns('selectors', {
      'await driver\\.findElement\\(By\\.css\\(([^)]+)\\)\\)': 'cy.get($1)',
      'await driver\\.findElement\\(By\\.id\\(([^)]+)\\)\\)': 'cy.get(`#${$1}`)',
      'await driver\\.findElement\\(By\\.name\\(([^)]+)\\)\\)': 'cy.get(`[name=${$1}]`)',
      'await driver\\.findElement\\(By\\.className\\(([^)]+)\\)\\)': 'cy.get(`.${$1}`)',
      'await driver\\.findElement\\(By\\.tagName\\(([^)]+)\\)\\)': 'cy.get($1)',
      'await driver\\.findElement\\(By\\.xpath\\(([^)]+)\\)\\)': 'cy.xpath($1)',
      'await driver\\.findElement\\(By\\.linkText\\(([^)]+)\\)\\)': 'cy.contains("a", $1)',
      'await driver\\.findElement\\(By\\.partialLinkText\\(([^)]+)\\)\\)': 'cy.contains("a", $1)',
      'await driver\\.findElements\\(By\\.css\\(([^)]+)\\)\\)': 'cy.get($1)',
      '\\.findElement\\(By\\.css\\(([^)]+)\\)\\)': '.find($1)',
      'await driver\\.switchTo\\(\\)\\.activeElement\\(\\)': 'cy.focused()'
    });

    // Interaction patterns
    this.engine.registerPatterns('interactions', {
      '\\.sendKeys\\(([^)]+)\\)': '.type($1)',
      '\\.click\\(\\)': '.click()',
      '\\.clear\\(\\)': '.clear()',
      '\\.submit\\(\\)': '.submit()'
    });

    // Assertion patterns
    this.engine.registerPatterns('assertions', {
      'expect\\(await ([^.]+)\\.isDisplayed\\(\\)\\)\\.toBe\\(true\\)': '$1.should("be.visible")',
      'expect\\(await ([^.]+)\\.isDisplayed\\(\\)\\)\\.toBe\\(false\\)': '$1.should("not.be.visible")',
      'expect\\(await ([^.]+)\\.getText\\(\\)\\)\\.toBe\\(([^)]+)\\)': '$1.should("have.text", $2)',
      'expect\\(await ([^.]+)\\.getText\\(\\)\\)\\.toContain\\(([^)]+)\\)': '$1.should("contain", $2)',
      'expect\\(await ([^.]+)\\.getAttribute\\([\'"]value[\'"]\\)\\)\\.toBe\\(([^)]+)\\)': '$1.should("have.value", $2)',
      'expect\\(await ([^.]+)\\.getAttribute\\(([^)]+)\\)\\)\\.toBe\\(([^)]+)\\)': '$1.should("have.attr", $2, $3)',
      'expect\\(await ([^.]+)\\.isSelected\\(\\)\\)\\.toBe\\(true\\)': '$1.should("be.checked")',
      'expect\\(await ([^.]+)\\.isSelected\\(\\)\\)\\.toBe\\(false\\)': '$1.should("not.be.checked")',
      'expect\\(await ([^.]+)\\.isEnabled\\(\\)\\)\\.toBe\\(false\\)': '$1.should("be.disabled")',
      'expect\\(await ([^.]+)\\.isEnabled\\(\\)\\)\\.toBe\\(true\\)': '$1.should("be.enabled")'
    });

    // Wait patterns
    this.engine.registerPatterns('waits', {
      'await driver\\.sleep\\((\\d+)\\)': 'cy.wait($1)',
      'await driver\\.wait\\(until\\.elementLocated\\(By\\.css\\(([^)]+)\\)\\),\\s*(\\d+)\\)': 'cy.get($1, { timeout: $2 })',
      'await driver\\.wait\\(until\\.elementIsVisible\\(([^)]+)\\),\\s*(\\d+)\\)': '$1.should("be.visible")',
      'await driver\\.wait\\(until\\.urlContains\\(([^)]+)\\),\\s*(\\d+)\\)': 'cy.url().should("include", $1)'
    });

    // Remove Selenium imports and setup
    this.engine.registerPatterns('cleanup', {
      "const\\s*\\{[^{}\n]*Builder[^{}\n]*\\}\\s*=\\s*require\\(['\"]selenium-webdriver['\"]\\);?": '',
      "const\\s*\\{[^{}\n]*expect[^{}\n]*\\}\\s*=\\s*require\\(['\"]@jest/globals['\"]\\);?": '',
      'let\\s+driver;?': '',
      'beforeAll\\s*\\([^)]*\\)\\s*\\{[^{}\n]*new\\s+Builder[^{}\n]*\\};?': '',
      'afterAll\\s*\\([^)]*\\)\\s*\\{[^{}\n]*driver\\.quit[^{}\n]*\\};?': ''
    });
  }

  async convert(content, options = {}) {
    let result = content;

    // Remove Selenium imports and setup/teardown
    result = this.removeSeleniumBoilerplate(result);

    // Convert Selenium commands to Cypress
    result = this.convertSeleniumCommands(result);

    // Remove await keywords (Cypress handles async)
    result = this.removeAsyncAwait(result);

    // Transform test structure
    result = this.transformTestStructure(result);

    // Add Cypress reference
    result = this.getSetup() + result;

    // Clean up empty lines
    result = result.replace(/\n{3,}/g, '\n\n').trim() + '\n';

    this.stats.conversions++;
    return result;
  }

  /**
   * Remove Selenium boilerplate (imports, driver setup/teardown)
   * @param {string} content - Content to clean
   * @returns {string}
   */
  removeSeleniumBoilerplate(content) {
    let result = content;

    // Remove Selenium imports
    // Note: Using [^{}\n]* to prevent ReDoS (already safe, just documenting)
    result = result.replace(/const\s*\{\s*Builder[^{}\n]*\}\s*=\s*require\(['"]selenium-webdriver['"]\);?\n?/g, '');
    result = result.replace(/const\s*\{\s*expect[^{}\n]*\}\s*=\s*require\(['"]@jest\/globals['"]\);?\n?/g, '');
    result = result.replace(/import\s*\{\s*Builder[^{}\n]*\}\s*from\s*['"]selenium-webdriver['"];?\n?/g, '');

    // Remove driver variable declaration
    result = result.replace(/let\s+driver;?\n?/g, '');

    // Remove beforeAll with driver setup
    result = result.replace(/beforeAll\s*\(\s*async\s*\(\)\s*=>\s*\{[^{}\n]*new\s+Builder[^{}\n]*\}\s*\);?\n?/g, '');

    // Remove afterAll with driver quit
    result = result.replace(/afterAll\s*\(\s*async\s*\(\)\s*=>\s*\{[^{}\n]*driver\.quit[^{}\n]*\}\s*\);?\n?/g, '');

    return result;
  }

  /**
   * Convert Selenium commands to Cypress equivalents
   * @param {string} content - Content to convert
   * @returns {string}
   */
  convertSeleniumCommands(content) {
    let result = content;

    // Convert assertions first (more specific patterns)
    // expect(await (await driver.findElement(By.css(selector))).isDisplayed()).toBe(true)
    result = result.replace(
      /expect\s*\(\s*await\s*\(\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.isDisplayed\s*\(\s*\)\s*\)\.toBe\s*\(\s*true\s*\)/g,
      "cy.get($1).should('be.visible')"
    );

    result = result.replace(
      /expect\s*\(\s*await\s*\(\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.isDisplayed\s*\(\s*\)\s*\)\.toBe\s*\(\s*false\s*\)/g,
      "cy.get($1).should('not.be.visible')"
    );

    result = result.replace(
      /expect\s*\(\s*await\s*\(\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.getText\s*\(\s*\)\s*\)\.toBe\s*\(([^)]+)\)/g,
      "cy.get($1).should('have.text', $2)"
    );

    result = result.replace(
      /expect\s*\(\s*await\s*\(\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.getText\s*\(\s*\)\s*\)\.toContain\s*\(([^)]+)\)/g,
      "cy.get($1).should('contain', $2)"
    );

    result = result.replace(
      /expect\s*\(\s*await\s*\(\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.getAttribute\s*\(\s*["']value["']\s*\)\s*\)\.toBe\s*\(([^)]+)\)/g,
      "cy.get($1).should('have.value', $2)"
    );

    result = result.replace(
      /expect\s*\(\s*await\s*\(\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.isSelected\s*\(\s*\)\s*\)\.toBe\s*\(\s*true\s*\)/g,
      "cy.get($1).should('be.checked')"
    );

    result = result.replace(
      /expect\s*\(\s*await\s*\(\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.isEnabled\s*\(\s*\)\s*\)\.toBe\s*\(\s*false\s*\)/g,
      "cy.get($1).should('be.disabled')"
    );

    result = result.replace(
      /expect\s*\(\s*await\s*\(\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.isEnabled\s*\(\s*\)\s*\)\.toBe\s*\(\s*true\s*\)/g,
      "cy.get($1).should('be.enabled')"
    );

    // Convert findElements length 0 check to not.exist (must come before general length check)
    result = result.replace(
      /expect\s*\(\s*\(\s*await\s+driver\.findElements\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.length\s*\)\.toBe\s*\(0\)/g,
      "cy.get($1).should('not.exist')"
    );

    result = result.replace(
      /expect\s*\(\s*\(\s*await\s+driver\.findElements\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.length\s*\)\.toBe\s*\((\d+)\)/g,
      "cy.get($1).should('have.length', $2)"
    );

    result = result.replace(
      /expect\s*\(\s*\(\s*await\s+driver\.findElements\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.length\s*\)\.toBeGreaterThan\s*\(0\)/g,
      "cy.get($1).should('exist')"
    );

    // Convert interactions
    result = result.replace(
      /await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\.sendKeys\s*\(([^)]+)\)/g,
      'cy.get($1).type($2)'
    );

    result = result.replace(
      /await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\.click\s*\(\s*\)/g,
      'cy.get($1).click()'
    );

    result = result.replace(
      /await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\.clear\s*\(\s*\)/g,
      'cy.get($1).clear()'
    );

    // Convert navigation
    result = result.replace(
      /await\s+driver\.get\s*\(([^)]+)\)/g,
      'cy.visit($1)'
    );

    result = result.replace(/await\s+driver\.navigate\s*\(\s*\)\.refresh\s*\(\s*\)/g, 'cy.reload()');
    result = result.replace(/await\s+driver\.navigate\s*\(\s*\)\.back\s*\(\s*\)/g, "cy.go('back')");
    result = result.replace(/await\s+driver\.navigate\s*\(\s*\)\.forward\s*\(\s*\)/g, "cy.go('forward')");

    // Convert URL assertions
    result = result.replace(
      /expect\s*\(\s*await\s+driver\.getCurrentUrl\s*\(\s*\)\s*\)\.toContain\s*\(([^)]+)\)/g,
      "cy.url().should('include', $1)"
    );

    result = result.replace(
      /expect\s*\(\s*await\s+driver\.getCurrentUrl\s*\(\s*\)\s*\)\.toBe\s*\(([^)]+)\)/g,
      "cy.url().should('eq', $1)"
    );

    // Convert title assertions
    result = result.replace(
      /expect\s*\(\s*await\s+driver\.getTitle\s*\(\s*\)\s*\)\.toBe\s*\(([^)]+)\)/g,
      "cy.title().should('eq', $1)"
    );

    // Convert waits
    result = result.replace(
      /await\s+driver\.sleep\s*\((\d+)\)/g,
      'cy.wait($1)'
    );

    // Convert storage/cookies
    result = result.replace(
      /await\s+driver\.manage\s*\(\s*\)\.deleteAllCookies\s*\(\s*\)/g,
      'cy.clearCookies()'
    );

    result = result.replace(
      /await\s+driver\.executeScript\s*\(\s*["']localStorage\.clear\(\)["']\s*\)/g,
      'cy.clearLocalStorage()'
    );

    // Convert checkbox check patterns (multiline)
    result = result.replace(
      /const\s+checkbox\s*=\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\);\s*\n?\s*if\s*\(\s*!\s*\(\s*await\s+checkbox\.isSelected\s*\(\s*\)\s*\)\s*\)\s*await\s+checkbox\.click\s*\(\s*\)/g,
      'cy.get($1).check()'
    );

    // Convert checkbox uncheck patterns (multiline)
    result = result.replace(
      /const\s+checkbox\s*=\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\);\s*\n?\s*if\s*\(\s*await\s+checkbox\.isSelected\s*\(\s*\)\s*\)\s*await\s+checkbox\.click\s*\(\s*\)/g,
      'cy.get($1).uncheck()'
    );

    // Convert select patterns (multiline)
    result = result.replace(
      /const\s+select\s*=\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\);\s*\n?\s*await\s+select\.findElement\s*\(\s*By\.css\s*\(\s*`option\[value=\$\{([^}]+)\}\]`\s*\)\s*\)\.click\s*\(\s*\)/g,
      'cy.get($1).select($2)'
    );

    return result;
  }

  removeAsyncAwait(content) {
    // Keep await in cy commands since they're already converted
    // Remove await from other places
    content = content.replace(/await\s+(?!cy\.)/g, '');
    // Convert async functions to regular
    content = content.replace(/async\s+function/g, 'function');
    content = content.replace(/async\s*\(\)/g, '()');
    return content;
  }

  transformTestStructure(content) {
    // Transform test callbacks
    // Note: Using [^,()\n]+ to prevent ReDoS
    content = content.replace(
      /it\(([^,()\n]+),\s*function\(\)\s*\{/g,
      'it($1, () => {'
    );

    // Remove driver references
    content = content.replace(/const\s+driver\s*=\s*this\.driver;?\n?/g, '');

    return content;
  }

  detectTestTypes(content) {
    return ['e2e'];
  }

  getImports(testTypes) {
    return [];
  }

  getSetup() {
    return `/// <reference types="cypress" />

`;
  }

  async convertConfig(configPath, options = {}) {
    return `const { defineConfig } = require('cypress');

module.exports = defineConfig({
  e2e: {
    baseUrl: 'http://localhost:3000',
    viewportWidth: 1280,
    viewportHeight: 720,
    video: false,
    screenshotOnRunFailure: true
  }
});
`;
  }
}

export default SeleniumToCypress;
