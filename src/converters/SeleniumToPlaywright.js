import { BaseConverter } from '../core/BaseConverter.js';
import { PatternEngine } from '../core/PatternEngine.js';

/**
 * Converts Selenium WebDriver tests to Playwright format
 */
export class SeleniumToPlaywright extends BaseConverter {
  constructor(options = {}) {
    super(options);
    this.sourceFramework = 'selenium';
    this.targetFramework = 'playwright';
    this.engine = new PatternEngine();
    this.initializePatterns();
  }

  initializePatterns() {
    // Test structure patterns
    this.engine.registerPatterns('structure', {
      'describe\\(': 'test.describe(',
      'it\\(': 'test('
    });

    // Navigation patterns
    this.engine.registerPatterns('navigation', {
      'await driver\\.get\\(([^)]+)\\)': 'await page.goto($1)',
      'await driver\\.navigate\\(\\)\\.to\\(([^)]+)\\)': 'await page.goto($1)',
      'await driver\\.navigate\\(\\)\\.back\\(\\)': 'await page.goBack()',
      'await driver\\.navigate\\(\\)\\.forward\\(\\)': 'await page.goForward()',
      'await driver\\.navigate\\(\\)\\.refresh\\(\\)': 'await page.reload()',
      'await driver\\.getCurrentUrl\\(\\)': 'page.url()',
      'await driver\\.getTitle\\(\\)': 'await page.title()'
    });

    // Selector patterns
    this.engine.registerPatterns('selectors', {
      'await driver\\.findElement\\(By\\.css\\(([^)]+)\\)\\)': 'page.locator($1)',
      'await driver\\.findElement\\(By\\.id\\(([^)]+)\\)\\)': 'page.locator(`#${$1}`)',
      'await driver\\.findElement\\(By\\.name\\(([^)]+)\\)\\)': 'page.locator(`[name=${$1}]`)',
      'await driver\\.findElement\\(By\\.className\\(([^)]+)\\)\\)': 'page.locator(`.${$1}`)',
      'await driver\\.findElement\\(By\\.tagName\\(([^)]+)\\)\\)': 'page.locator($1)',
      'await driver\\.findElement\\(By\\.xpath\\(([^)]+)\\)\\)': 'page.locator($1)',
      'await driver\\.findElement\\(By\\.linkText\\(([^)]+)\\)\\)': 'page.getByText($1)',
      'await driver\\.findElements\\(By\\.css\\(([^)]+)\\)\\)': 'page.locator($1)',
      '\\.findElement\\(By\\.css\\(([^)]+)\\)\\)': '.locator($1)',
      'await driver\\.switchTo\\(\\)\\.activeElement\\(\\)': 'page.locator(":focus")'
    });

    // Interaction patterns
    this.engine.registerPatterns('interactions', {
      '\\.sendKeys\\(([^)]+)\\)': '.fill($1)',
      '\\.click\\(\\)': '.click()',
      '\\.clear\\(\\)': '.clear()',
      '\\.submit\\(\\)': '.press("Enter")'
    });

    // Assertion patterns
    this.engine.registerPatterns('assertions', {
      'expect\\(await ([^.]+)\\.isDisplayed\\(\\)\\)\\.toBe\\(true\\)': 'await expect($1).toBeVisible()',
      'expect\\(await ([^.]+)\\.isDisplayed\\(\\)\\)\\.toBe\\(false\\)': 'await expect($1).toBeHidden()',
      'expect\\(await ([^.]+)\\.getText\\(\\)\\)\\.toBe\\(([^)]+)\\)': 'await expect($1).toHaveText($2)',
      'expect\\(await ([^.]+)\\.getText\\(\\)\\)\\.toContain\\(([^)]+)\\)': 'await expect($1).toContainText($2)',
      'expect\\(await ([^.]+)\\.getAttribute\\([\'"]value[\'"]\\)\\)\\.toBe\\(([^)]+)\\)': 'await expect($1).toHaveValue($2)',
      'expect\\(await ([^.]+)\\.getAttribute\\(([^)]+)\\)\\)\\.toBe\\(([^)]+)\\)': 'await expect($1).toHaveAttribute($2, $3)',
      'expect\\(await ([^.]+)\\.isSelected\\(\\)\\)\\.toBe\\(true\\)': 'await expect($1).toBeChecked()',
      'expect\\(await ([^.]+)\\.isSelected\\(\\)\\)\\.toBe\\(false\\)': 'await expect($1).not.toBeChecked()',
      'expect\\(await ([^.]+)\\.isEnabled\\(\\)\\)\\.toBe\\(false\\)': 'await expect($1).toBeDisabled()',
      'expect\\(await ([^.]+)\\.isEnabled\\(\\)\\)\\.toBe\\(true\\)': 'await expect($1).toBeEnabled()'
    });

    // Wait patterns
    this.engine.registerPatterns('waits', {
      'await driver\\.sleep\\((\\d+)\\)': 'await page.waitForTimeout($1)',
      'await driver\\.wait\\(until\\.elementLocated\\(By\\.css\\(([^)]+)\\)\\),\\s*(\\d+)\\)': 'await page.waitForSelector($1, { timeout: $2 })',
      'await driver\\.wait\\(until\\.elementIsVisible\\(([^)]+)\\),\\s*(\\d+)\\)': 'await page.waitForSelector($1, { state: "visible", timeout: $2 })',
      'await driver\\.wait\\(until\\.urlContains\\(([^)]+)\\),\\s*(\\d+)\\)': 'await page.waitForURL($1)'
    });

    // Remove Selenium imports and setup
    this.engine.registerPatterns('cleanup', {
      "const\\s*\\{[^}]*Builder[^}]*\\}\\s*=\\s*require\\(['\"]selenium-webdriver['\"]\\);?\\n?": '',
      "const\\s*\\{[^}]*expect[^}]*\\}\\s*=\\s*require\\(['\"]@jest/globals['\"]\\);?\\n?": '',
      'let\\s+driver;?\\n?': '',
      'beforeAll\\s*\\([^)]*\\)\\s*\\{[\\s\\S]*?new\\s+Builder[\\s\\S]*?\\};?\\n?': '',
      'afterAll\\s*\\([^)]*\\)\\s*\\{[\\s\\S]*?driver\\.quit[\\s\\S]*?\\};?\\n?': ''
    });
  }

  async convert(content, options = {}) {
    let result = content;

    // Remove Selenium boilerplate
    result = this.removeSeleniumBoilerplate(result);

    // Convert Selenium commands to Playwright
    result = this.convertSeleniumCommands(result);

    // Convert test structure
    result = this.convertTestStructure(result);

    // Transform test callbacks
    result = this.transformTestCallbacks(result);

    // Add imports
    const imports = this.getImports([]);
    result = imports.join('\n') + '\n\n' + result;

    // Clean up
    result = result.replace(/\n{3,}/g, '\n\n').trim() + '\n';

    this.stats.conversions++;
    return result;
  }

  /**
   * Remove Selenium boilerplate
   * @param {string} content - Content to clean
   * @returns {string}
   */
  removeSeleniumBoilerplate(content) {
    let result = content;

    // Remove Selenium imports
    result = result.replace(/const\s*\{[^}]*Builder[^}]*\}\s*=\s*require\(['"]selenium-webdriver['"]\);?\n?/g, '');
    result = result.replace(/const\s*\{[^}]*expect[^}]*\}\s*=\s*require\(['"]@jest\/globals['"]\);?\n?/g, '');

    // Remove driver variable declaration
    result = result.replace(/let\s+driver;?\n?/g, '');

    // Remove beforeAll with driver setup
    result = result.replace(/beforeAll\s*\(\s*async\s*\(\)\s*=>\s*\{[^}]*new\s+Builder[^}]*\}\s*\);?\n?/g, '');

    // Remove afterAll with driver quit
    result = result.replace(/afterAll\s*\(\s*async\s*\(\)\s*=>\s*\{[^}]*driver\.quit[^}]*\}\s*\);?\n?/g, '');

    return result;
  }

  /**
   * Convert Selenium commands to Playwright equivalents
   * @param {string} content - Content to convert
   * @returns {string}
   */
  convertSeleniumCommands(content) {
    let result = content;

    // Convert assertions (complex patterns first)
    result = result.replace(
      /expect\s*\(\s*await\s*\(\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.isDisplayed\s*\(\s*\)\s*\)\.toBe\s*\(\s*true\s*\)/g,
      'await expect(page.locator($1)).toBeVisible()'
    );

    result = result.replace(
      /expect\s*\(\s*await\s*\(\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.isDisplayed\s*\(\s*\)\s*\)\.toBe\s*\(\s*false\s*\)/g,
      'await expect(page.locator($1)).toBeHidden()'
    );

    result = result.replace(
      /expect\s*\(\s*await\s*\(\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.getText\s*\(\s*\)\s*\)\.toBe\s*\(([^)]+)\)/g,
      'await expect(page.locator($1)).toHaveText($2)'
    );

    result = result.replace(
      /expect\s*\(\s*await\s*\(\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.getText\s*\(\s*\)\s*\)\.toContain\s*\(([^)]+)\)/g,
      'await expect(page.locator($1)).toContainText($2)'
    );

    result = result.replace(
      /expect\s*\(\s*await\s*\(\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.getAttribute\s*\(\s*["']value["']\s*\)\s*\)\.toBe\s*\(([^)]+)\)/g,
      'await expect(page.locator($1)).toHaveValue($2)'
    );

    result = result.replace(
      /expect\s*\(\s*await\s*\(\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.isSelected\s*\(\s*\)\s*\)\.toBe\s*\(\s*true\s*\)/g,
      'await expect(page.locator($1)).toBeChecked()'
    );

    result = result.replace(
      /expect\s*\(\s*await\s*\(\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.isEnabled\s*\(\s*\)\s*\)\.toBe\s*\(\s*false\s*\)/g,
      'await expect(page.locator($1)).toBeDisabled()'
    );

    result = result.replace(
      /expect\s*\(\s*await\s*\(\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.isEnabled\s*\(\s*\)\s*\)\.toBe\s*\(\s*true\s*\)/g,
      'await expect(page.locator($1)).toBeEnabled()'
    );

    // Handle findElements length 0 check first (more specific pattern)
    result = result.replace(
      /expect\s*\(\s*\(\s*await\s+driver\.findElements\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.length\s*\)\.toBe\s*\(0\)/g,
      'await expect(page.locator($1)).not.toBeAttached()'
    );

    result = result.replace(
      /expect\s*\(\s*\(\s*await\s+driver\.findElements\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.length\s*\)\.toBe\s*\((\d+)\)/g,
      'await expect(page.locator($1)).toHaveCount($2)'
    );

    result = result.replace(
      /expect\s*\(\s*\(\s*await\s+driver\.findElements\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\s*\)\.length\s*\)\.toBeGreaterThan\s*\(0\)/g,
      'await expect(page.locator($1)).toBeAttached()'
    );

    // Convert interactions
    result = result.replace(
      /await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\.sendKeys\s*\(([^)]+)\)/g,
      'await page.locator($1).fill($2)'
    );

    result = result.replace(
      /await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\.click\s*\(\s*\)/g,
      'await page.locator($1).click()'
    );

    result = result.replace(
      /await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\)\.clear\s*\(\s*\)/g,
      'await page.locator($1).clear()'
    );

    // Convert navigation
    result = result.replace(
      /await\s+driver\.get\s*\(([^)]+)\)/g,
      'await page.goto($1)'
    );

    result = result.replace(/await\s+driver\.navigate\s*\(\s*\)\.refresh\s*\(\s*\)/g, 'await page.reload()');
    result = result.replace(/await\s+driver\.navigate\s*\(\s*\)\.back\s*\(\s*\)/g, 'await page.goBack()');
    result = result.replace(/await\s+driver\.navigate\s*\(\s*\)\.forward\s*\(\s*\)/g, 'await page.goForward()');

    // Convert URL assertions
    result = result.replace(
      /expect\s*\(\s*await\s+driver\.getCurrentUrl\s*\(\s*\)\s*\)\.toContain\s*\(([^)]+)\)/g,
      'await expect(page).toHaveURL(new RegExp($1))'
    );

    result = result.replace(
      /expect\s*\(\s*await\s+driver\.getCurrentUrl\s*\(\s*\)\s*\)\.toBe\s*\(([^)]+)\)/g,
      'await expect(page).toHaveURL($1)'
    );

    // Convert title assertions
    result = result.replace(
      /expect\s*\(\s*await\s+driver\.getTitle\s*\(\s*\)\s*\)\.toBe\s*\(([^)]+)\)/g,
      'await expect(page).toHaveTitle($1)'
    );

    // Convert waits
    result = result.replace(
      /await\s+driver\.sleep\s*\((\d+)\)/g,
      'await page.waitForTimeout($1)'
    );

    // Convert storage/cookies
    result = result.replace(
      /await\s+driver\.manage\s*\(\s*\)\.deleteAllCookies\s*\(\s*\)/g,
      'await context.clearCookies()'
    );

    result = result.replace(
      /await\s+driver\.executeScript\s*\(\s*["']localStorage\.clear\(\)["']\s*\)/g,
      'await page.evaluate(() => localStorage.clear())'
    );

    // Convert checkbox check patterns (multiline)
    result = result.replace(
      /const\s+checkbox\s*=\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\);\s*\n?\s*if\s*\(\s*!\s*\(\s*await\s+checkbox\.isSelected\s*\(\s*\)\s*\)\s*\)\s*await\s+checkbox\.click\s*\(\s*\)/g,
      'await page.locator($1).check()'
    );

    // Convert checkbox uncheck patterns (multiline)
    result = result.replace(
      /const\s+checkbox\s*=\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\);\s*\n?\s*if\s*\(\s*await\s+checkbox\.isSelected\s*\(\s*\)\s*\)\s*await\s+checkbox\.click\s*\(\s*\)/g,
      'await page.locator($1).uncheck()'
    );

    // Convert select patterns (multiline)
    result = result.replace(
      /const\s+select\s*=\s*await\s+driver\.findElement\s*\(\s*By\.css\s*\(([^)]+)\)\s*\);\s*\n?\s*await\s+select\.findElement\s*\(\s*By\.css\s*\(\s*`option\[value=\$\{([^}]+)\}\]`\s*\)\s*\)\.click\s*\(\s*\)/g,
      'await page.locator($1).selectOption($2)'
    );

    // Convert XPath selectors with click
    result = result.replace(
      /await\s+driver\.findElement\s*\(\s*By\.xpath\s*\(([^)]+)\)\s*\)\.click\s*\(\s*\)/g,
      'await page.locator(`xpath=$1`).click()'
    );

    // Convert XPath selectors (general)
    result = result.replace(
      /await\s+driver\.findElement\s*\(\s*By\.xpath\s*\(([^)]+)\)\s*\)/g,
      'page.locator(`xpath=$1`)'
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

    // Convert describe blocks
    result = result.replace(/describe\.only\(/g, 'test.describe.only(');
    result = result.replace(/describe\.skip\(/g, 'test.describe.skip(');
    result = result.replace(/describe\(/g, 'test.describe(');

    // Convert it blocks
    result = result.replace(/it\.only\(/g, 'test.only(');
    result = result.replace(/it\.skip\(/g, 'test.skip(');
    result = result.replace(/it\(/g, 'test(');

    // Convert hooks
    result = result.replace(/beforeAll\(/g, 'test.beforeAll(');
    result = result.replace(/afterAll\(/g, 'test.afterAll(');
    result = result.replace(/beforeEach\(/g, 'test.beforeEach(');
    result = result.replace(/afterEach\(/g, 'test.afterEach(');

    return result;
  }

  transformTestCallbacks(content) {
    // Transform test callbacks to include page
    content = content.replace(
      /test\(([^,]+),\s*(?:async\s*)?function\(\)\s*\{/g,
      'test($1, async ({ page }) => {'
    );

    content = content.replace(
      /test\(([^,]+),\s*(?:async\s*)?\(\)\s*=>\s*\{/g,
      'test($1, async ({ page }) => {'
    );

    // Fix describe callbacks (should NOT have page parameter)
    content = content.replace(
      /test\.describe\(([^,]+),\s*(?:async\s*)?\(\s*\{[^}]*\}\s*\)\s*=>\s*\{/g,
      'test.describe($1, () => {'
    );

    content = content.replace(
      /test\.describe\(([^,]+),\s*(?:async\s*)?\(\)\s*=>\s*\{/g,
      'test.describe($1, () => {'
    );

    // Transform hooks
    content = content.replace(
      /test\.(beforeEach|afterEach)\s*\(\s*(?:async\s*)?\(\)\s*=>\s*\{/g,
      'test.$1(async ({ page }) => {'
    );

    content = content.replace(
      /test\.(beforeAll|afterAll)\s*\(\s*(?:async\s*)?\(\)\s*=>\s*\{/g,
      'test.$1(async () => {'
    );

    return content;
  }

  detectTestTypes(content) {
    return ['e2e'];
  }

  getImports(testTypes) {
    return [
      "import { test, expect } from '@playwright/test';"
    ];
  }

  async convertConfig(configPath, options = {}) {
    return `import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  timeout: 30000,
  expect: {
    timeout: 5000
  },
  use: {
    baseURL: 'http://localhost:3000',
    viewport: { width: 1280, height: 720 },
    trace: 'retain-on-failure'
  },
  projects: [
    { name: 'chromium', use: { browserName: 'chromium' } },
    { name: 'firefox', use: { browserName: 'firefox' } },
    { name: 'webkit', use: { browserName: 'webkit' } }
  ]
});
`;
  }
}

export default SeleniumToPlaywright;
