/**
 * Selenium WebDriver (JavaScript) framework definition.
 *
 * Provides detect, parse, and emit for the Selenium WebDriver JS framework.
 * emit() transforms Cypress, Playwright, WebdriverIO, Puppeteer, and TestCafe
 * source code into Selenium WebDriver JS code.
 */

import {
  TestFile,
  TestSuite,
  TestCase,
  Hook,
  Assertion,
  ImportStatement,
  RawCode,
  Comment,
} from '../../../core/ir.js';

import { TodoFormatter } from '../../../core/TodoFormatter.js';

const formatter = new TodoFormatter('javascript');

function detect(source) {
  if (!source || !source.trim()) return 0;

  let score = 0;

  // Strong signals
  if (/require\s*\(\s*['"]selenium-webdriver['"]\s*\)/.test(source))
    score += 40;
  if (/from\s+['"]selenium-webdriver['"]/.test(source)) score += 40;
  if (/By\.css\s*\(/.test(source)) score += 20;
  if (/driver\.findElement\s*\(/.test(source)) score += 15;
  if (/driver\.get\s*\(/.test(source)) score += 15;
  if (/new\s+Builder\s*\(/.test(source)) score += 10;

  // Medium signals
  if (/driver\.wait\s*\(/.test(source)) score += 10;
  if (/until\.elementLocated\s*\(/.test(source)) score += 10;
  if (/\.sendKeys\s*\(/.test(source)) score += 10;
  if (/driver\.findElements\s*\(/.test(source)) score += 5;
  if (/driver\.navigate\(\)/.test(source)) score += 5;
  if (/driver\.getTitle\s*\(/.test(source)) score += 5;
  if (/driver\.getCurrentUrl\s*\(/.test(source)) score += 5;

  // Negative: Cypress
  if (/\bcy\./.test(source)) score -= 30;
  // Negative: Playwright
  if (/\bpage\.goto\s*\(/.test(source)) score -= 30;
  if (/from\s+['"]@playwright\/test['"]/.test(source)) score -= 30;
  // Negative: Java Selenium
  if (/import\s+org\.openqa\.selenium/.test(source)) score -= 40;
  // Negative: Python Selenium
  if (/from\s+selenium\s+import/.test(source)) score -= 40;
  // Negative: WebdriverIO
  if (/from\s+['"]@wdio\/globals['"]/.test(source)) score -= 30;
  if (/\bbrowser\.url\s*\(/.test(source)) score -= 20;

  return Math.max(0, Math.min(100, score));
}

function parse(source) {
  const lines = source.split('\n');
  const imports = [];
  const body = [];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();
    const loc = { line: i + 1, column: 0 };

    if (!trimmed) continue;

    if (
      trimmed.startsWith('//') ||
      trimmed.startsWith('/*') ||
      trimmed.startsWith('*')
    ) {
      body.push(
        new Comment({ text: line, sourceLocation: loc, originalSource: line })
      );
      continue;
    }

    if (/^import\s/.test(trimmed) || /^const\s.*=\s*require\(/.test(trimmed)) {
      imports.push(
        new ImportStatement({
          source: trimmed,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    if (/\bdescribe\s*\(/.test(trimmed)) {
      body.push(
        new TestSuite({
          name: '',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    if (/\b(?:it|test)\s*\(/.test(trimmed)) {
      body.push(
        new TestCase({
          name: '',
          isAsync: /async/.test(trimmed),
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    if (
      /\b(?:beforeEach|afterEach|beforeAll|afterAll|before|after)\s*\(/.test(
        trimmed
      )
    ) {
      body.push(
        new Hook({
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    if (/\bexpect\s*\(/.test(trimmed)) {
      body.push(
        new Assertion({
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    if (/\bdriver\./.test(trimmed)) {
      body.push(
        new RawCode({
          code: line,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    if (/\bBy\./.test(trimmed)) {
      body.push(
        new RawCode({
          code: line,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    body.push(
      new RawCode({ code: line, sourceLocation: loc, originalSource: line })
    );
  }

  return new TestFile({ language: 'javascript', imports, body });
}

/**
 * Emit Selenium WebDriver JS code from IR + original source.
 *
 * Handles Cypress→Selenium, Playwright→Selenium, WDIO→Selenium,
 * Puppeteer→Selenium, and TestCafe→Selenium conversions.
 *
 * @param {TestFile} _ir - Parsed IR tree (for scoring metadata)
 * @param {string} source - Original source code
 * @returns {string} Converted Selenium WebDriver JS source code
 */
function emit(_ir, source) {
  let result = source;

  const isPlaywrightSource =
    /from\s+['"]@playwright\/test['"]/.test(source) ||
    /\bpage\.goto\s*\(/.test(source);
  const isCypressSource = /\bcy\./.test(source);
  const isWdioSource =
    /from\s+['"]@wdio\/globals['"]/.test(source) ||
    /\bbrowser\.url\s*\(/.test(source);
  const isPuppeteerSource = /\bpuppeteer\.launch/.test(source);
  const isTestCafeSource =
    /from\s+['"]testcafe['"]/.test(source) || /\bfixture\s*`/.test(source);

  // Phase 1: Convert assertions (BEFORE selectors to match composite patterns)
  result = convertAssertions(
    result,
    isCypressSource,
    isPlaywrightSource,
    isWdioSource
  );

  // Phase 2: Convert navigation
  result = convertNavigation(
    result,
    isCypressSource,
    isPlaywrightSource,
    isWdioSource,
    isPuppeteerSource,
    isTestCafeSource
  );

  // Phase 3: Convert interactions
  result = convertInteractions(
    result,
    isCypressSource,
    isPlaywrightSource,
    isWdioSource,
    isPuppeteerSource,
    isTestCafeSource
  );

  // Phase 4: Convert selectors (standalone patterns, after composite patterns)
  result = convertSelectors(
    result,
    isCypressSource,
    isPlaywrightSource,
    isWdioSource,
    isPuppeteerSource,
    isTestCafeSource
  );

  // Phase 5: Convert test structure
  result = convertTestStructure(result, isPlaywrightSource, isTestCafeSource);

  // Phase 6: Strip source imports, add Selenium imports
  result = convertImports(
    result,
    isCypressSource,
    isPlaywrightSource,
    isWdioSource,
    isPuppeteerSource,
    isTestCafeSource
  );

  // Phase 7: Add driver setup/teardown boilerplate
  result = addDriverBoilerplate(result);

  // Phase 8: Cleanup
  result = cleanupOutput(result);

  return result;
}

// ── Phase 1: Navigation ──

function convertNavigation(
  content,
  isCypress,
  isPlaywright,
  isWdio,
  isPuppeteer,
  isTestCafe
) {
  let result = content;

  if (isCypress) {
    result = result.replace(/cy\.visit\(([^)]+)\)/g, 'await driver.get($1)');
    result = result.replace(
      /cy\.go\(['"]back['"]\)/g,
      'await driver.navigate().back()'
    );
    result = result.replace(
      /cy\.go\(['"]forward['"]\)/g,
      'await driver.navigate().forward()'
    );
    result = result.replace(
      /cy\.reload\(\)/g,
      'await driver.navigate().refresh()'
    );
  }

  if (isPlaywright) {
    result = result.replace(
      /await page\.goto\(([^)]+)\)/g,
      'await driver.get($1)'
    );
    result = result.replace(
      /await page\.goBack\(\)/g,
      'await driver.navigate().back()'
    );
    result = result.replace(
      /await page\.goForward\(\)/g,
      'await driver.navigate().forward()'
    );
    result = result.replace(
      /await page\.reload\(\)/g,
      'await driver.navigate().refresh()'
    );
    result = result.replace(
      /await page\.title\(\)/g,
      'await driver.getTitle()'
    );
    result = result.replace(
      /await page\.url\(\)/g,
      'await driver.getCurrentUrl()'
    );
    // Waits
    result = result.replace(
      /await page\.waitForTimeout\((\d+)\)/g,
      'await driver.sleep($1)'
    );
    result = result.replace(
      /await page\.waitForSelector\(([^)]+)\)/g,
      'await driver.wait(until.elementLocated(By.css($1)), 10000)'
    );
    result = result.replace(
      /await page\.waitForURL\(([^)]+)\)/g,
      'await driver.wait(until.urlContains($1), 10000)'
    );
    // Viewport
    result = result.replace(
      /await page\.setViewportSize\(\s*\{\s*width:\s*(\d+),\s*height:\s*(\d+)\s*\}\s*\)/g,
      'await driver.manage().window().setRect({ width: $1, height: $2 })'
    );
  }

  if (isWdio) {
    result = result.replace(
      /await browser\.url\(([^)]+)\)/g,
      'await driver.get($1)'
    );
    result = result.replace(
      /await browser\.back\(\)/g,
      'await driver.navigate().back()'
    );
    result = result.replace(
      /await browser\.forward\(\)/g,
      'await driver.navigate().forward()'
    );
    result = result.replace(
      /await browser\.refresh\(\)/g,
      'await driver.navigate().refresh()'
    );
    result = result.replace(
      /await browser\.getTitle\(\)/g,
      'await driver.getTitle()'
    );
    result = result.replace(
      /await browser\.getUrl\(\)/g,
      'await driver.getCurrentUrl()'
    );
  }

  if (isPuppeteer) {
    result = result.replace(
      /await page\.goto\(([^)]+)\)/g,
      'await driver.get($1)'
    );
    result = result.replace(
      /await page\.goBack\(\)/g,
      'await driver.navigate().back()'
    );
    result = result.replace(
      /await page\.goForward\(\)/g,
      'await driver.navigate().forward()'
    );
    result = result.replace(
      /await page\.reload\(\)/g,
      'await driver.navigate().refresh()'
    );
  }

  if (isTestCafe) {
    result = result.replace(
      /await t\.navigateTo\(([^)]+)\)/g,
      'await driver.get($1)'
    );
  }

  return result;
}

// ── Phase 2: Selectors ──

function convertSelectors(
  content,
  isCypress,
  isPlaywright,
  isWdio,
  isPuppeteer,
  isTestCafe
) {
  let result = content;

  if (isCypress) {
    result = result.replace(
      /cy\.get\(([^)]+)\)/g,
      'await driver.findElement(By.css($1))'
    );
    result = result.replace(
      /cy\.contains\(([^)]+)\)/g,
      'await driver.findElement(By.xpath(`//*[contains(text(),$1)]`))'
    );
  }

  if (isPlaywright) {
    // page.locator with actions handled in interactions phase
    // Standalone page.locator
    result = result.replace(
      /page\.locator\(([^)]+)\)/g,
      'driver.findElement(By.css($1))'
    );
    result = result.replace(
      /page\.getByText\(([^)]+)\)/g,
      'driver.findElement(By.xpath(`//*[contains(text(),$1)]`))'
    );
  }

  if (isWdio) {
    result = result.replace(
      /\$\(([^)]+)\)/g,
      'await driver.findElement(By.css($1))'
    );
    result = result.replace(
      /\$\$\(([^)]+)\)/g,
      'await driver.findElements(By.css($1))'
    );
  }

  if (isPuppeteer) {
    result = result.replace(
      /await page\.\$\(([^)]+)\)/g,
      'await driver.findElement(By.css($1))'
    );
    result = result.replace(
      /await page\.\$\$\(([^)]+)\)/g,
      'await driver.findElements(By.css($1))'
    );
    result = result.replace(
      /await page\.waitForSelector\(([^)]+)\)/g,
      'await driver.wait(until.elementLocated(By.css($1)), 10000)'
    );
  }

  if (isTestCafe) {
    result = result.replace(
      /Selector\(([^)]+)\)/g,
      'driver.findElement(By.css($1))'
    );
  }

  return result;
}

// ── Phase 3: Interactions ──

function convertInteractions(
  content,
  isCypress,
  isPlaywright,
  isWdio,
  isPuppeteer,
  isTestCafe
) {
  let result = content;

  if (isCypress) {
    result = result.replace(/\.type\(([^)]+)\)/g, '.sendKeys($1)');
    result = result.replace(/\.click\(\)/g, '.click()');
    result = result.replace(/\.clear\(\)/g, '.clear()');
    result = result.replace(/\.check\(\)/g, '.click()');
    result = result.replace(/\.uncheck\(\)/g, '.click()');
  }

  if (isPlaywright) {
    result = result.replace(
      /await page\.locator\(([^)]+)\)\.fill\(([^)]+)\)/g,
      'await (await driver.findElement(By.css($1))).sendKeys($2)'
    );
    result = result.replace(
      /await page\.locator\(([^)]+)\)\.click\(\)/g,
      'await (await driver.findElement(By.css($1))).click()'
    );
    result = result.replace(
      /await page\.locator\(([^)]+)\)\.clear\(\)/g,
      'await (await driver.findElement(By.css($1))).clear()'
    );
    result = result.replace(
      /await page\.locator\(([^)]+)\)\.textContent\(\)/g,
      'await (await driver.findElement(By.css($1))).getText()'
    );
    result = result.replace(
      /await page\.locator\(([^)]+)\)\.isVisible\(\)/g,
      'await (await driver.findElement(By.css($1))).isDisplayed()'
    );
  }

  if (isWdio) {
    result = result.replace(/\.setValue\(([^)]+)\)/g, '.sendKeys($1)');
    result = result.replace(/\.clearValue\(\)/g, '.clear()');
    result = result.replace(/\.doubleClick\(\)/g, '.click()');
    result = result.replace(/\.getText\(\)/g, '.getText()');
    result = result.replace(/\.isDisplayed\(\)/g, '.isDisplayed()');
  }

  if (isPuppeteer) {
    result = result.replace(
      /await page\.type\(([^,]+),\s*([^)]+)\)/g,
      'await (await driver.findElement(By.css($1))).sendKeys($2)'
    );
    result = result.replace(
      /await page\.click\(([^)]+)\)/g,
      'await (await driver.findElement(By.css($1))).click()'
    );
  }

  if (isTestCafe) {
    result = result.replace(
      /await t\.typeText\(([^,]+),\s*([^)]+)\)/g,
      'await (await driver.findElement(By.css($1))).sendKeys($2)'
    );
    result = result.replace(
      /await t\.click\(([^)]+)\)/g,
      'await (await driver.findElement(By.css($1))).click()'
    );
  }

  return result;
}

// ── Phase 4: Assertions ──

function convertAssertions(content, isCypress, isPlaywright, isWdio) {
  let result = content;

  if (isCypress) {
    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]be\.visible['"]\)/g,
      'expect(await (await driver.findElement(By.css($1))).isDisplayed()).toBe(true)'
    );
    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]exist['"]\)/g,
      'expect(await driver.findElement(By.css($1))).toBeDefined()'
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
      "expect(await (await driver.findElement(By.css($1))).getAttribute('value')).toBe($2)"
    );
    // Element state assertions
    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]be\.disabled['"]\)/g,
      'expect(await (await driver.findElement(By.css($1))).isEnabled()).toBe(false)'
    );
    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]be\.enabled['"]\)/g,
      'expect(await (await driver.findElement(By.css($1))).isEnabled()).toBe(true)'
    );
    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]be\.checked['"]\)/g,
      'expect(await (await driver.findElement(By.css($1))).isSelected()).toBe(true)'
    );
    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]not\.be\.checked['"]\)/g,
      'expect(await (await driver.findElement(By.css($1))).isSelected()).toBe(false)'
    );
    result = result.replace(
      /cy\.get\(([^)]+)\)\.should\(['"]have\.focus['"]\)/g,
      'expect(await driver.switchTo().activeElement()).toEqual(await driver.findElement(By.css($1)))'
    );
    result = result.replace(
      /cy\.url\(\)\.should\(['"]include['"],\s*([^)]+)\)/g,
      'expect(await driver.getCurrentUrl()).toContain($1)'
    );
    result = result.replace(
      /cy\.url\(\)\.should\(['"]eq['"],\s*([^)]+)\)/g,
      'expect(await driver.getCurrentUrl()).toBe($1)'
    );
    result = result.replace(
      /cy\.title\(\)\.should\(['"]eq['"],\s*([^)]+)\)/g,
      'expect(await driver.getTitle()).toBe($1)'
    );
  }

  if (isPlaywright) {
    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toBeVisible\(\)/g,
      'expect(await (await driver.findElement(By.css($1))).isDisplayed()).toBe(true)'
    );
    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toHaveText\(([^)]+)\)/g,
      'expect(await (await driver.findElement(By.css($1))).getText()).toBe($2)'
    );
    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toContainText\(([^)]+)\)/g,
      'expect(await (await driver.findElement(By.css($1))).getText()).toContain($2)'
    );
    result = result.replace(
      /await expect\(page\.locator\(([^)]+)\)\)\.toHaveValue\(([^)]+)\)/g,
      "expect(await (await driver.findElement(By.css($1))).getAttribute('value')).toBe($2)"
    );
    result = result.replace(
      /await expect\(page\)\.toHaveURL\(([^)]+)\)/g,
      'expect(await driver.getCurrentUrl()).toBe($1)'
    );
    result = result.replace(
      /await expect\(page\)\.toHaveTitle\(([^)]+)\)/g,
      'expect(await driver.getTitle()).toBe($1)'
    );
  }

  if (isWdio) {
    result = result.replace(
      /await expect\(browser\)\.toHaveUrl\(([^)]+)\)/g,
      'expect(await driver.getCurrentUrl()).toBe($1)'
    );
    result = result.replace(
      /await expect\(browser\)\.toHaveTitle\(([^)]+)\)/g,
      'expect(await driver.getTitle()).toBe($1)'
    );
    result = result.replace(
      /await expect\(\$\(([^)]+)\)\)\.toBeDisplayed\(\)/g,
      'expect(await (await driver.findElement(By.css($1))).isDisplayed()).toBe(true)'
    );
    result = result.replace(
      /await expect\(\$\(([^)]+)\)\)\.toHaveText\(([^)]+)\)/g,
      'expect(await (await driver.findElement(By.css($1))).getText()).toBe($2)'
    );
  }

  return result;
}

// ── Phase 5: Test structure ──

function convertTestStructure(content, isPlaywright, isTestCafe) {
  let result = content;

  if (isPlaywright) {
    result = result.replace(/test\.describe\.only\(/g, 'describe.only(');
    result = result.replace(/test\.describe\.skip\(/g, 'describe.skip(');
    result = result.replace(/test\.describe\(/g, 'describe(');
    result = result.replace(/test\.only\(/g, 'it.only(');
    result = result.replace(/test\.skip\(/g, 'it.skip(');
    result = result.replace(/test\.beforeAll\(/g, 'beforeAll(');
    result = result.replace(/test\.afterAll\(/g, 'afterAll(');
    result = result.replace(/test\.beforeEach\(/g, 'beforeEach(');
    result = result.replace(/test\.afterEach\(/g, 'afterEach(');
    result = result.replace(/\btest\(([^,()\n]+),/g, 'it($1,');

    // Remove { page } destructure from callback params
    result = result.replace(
      /\(\s*\{\s*page\s*(?:,\s*request\s*)?\}\s*\)\s*=>/g,
      '() =>'
    );
  }

  if (isTestCafe) {
    // fixture`Name`.page`URL` -> describe('Name', () => { ... })
    result = result.replace(
      /fixture\s*`([^`]+)`\s*\.page\s*`([^`]+)`\s*;?/g,
      (_, name, _url) => `describe('${name}', () => {`
    );
    // test('name', async t => { -> it('name', async () => {
    result = result.replace(
      /test\(([^,]+),\s*async\s+t\s*=>\s*\{/g,
      'it($1, async () => {'
    );
  }

  // Add async to test callbacks that don't already have it (for Cypress)
  result = result.replace(
    /((?:it|test)\s*\([^,]+,\s*)(?!async)\(\s*\)\s*=>\s*\{/g,
    '$1async () => {'
  );
  result = result.replace(
    /((?:beforeEach|afterEach|beforeAll|afterAll|before|after)\s*\(\s*)(?!async)\(\s*\)\s*=>\s*\{/g,
    '$1async () => {'
  );

  return result;
}

// ── Phase 6: Imports ──

function convertImports(
  content,
  isCypress,
  isPlaywright,
  isWdio,
  isPuppeteer,
  isTestCafe
) {
  let result = content;

  // Remove source-framework-specific imports
  if (isPlaywright) {
    result = result.replace(
      /import\s+\{[^}]*\}\s+from\s+['"]@playwright\/test['"];?\n?/g,
      ''
    );
  }
  if (isCypress) {
    result = result.replace(
      /\/\/\/\s*<reference\s+types=["']cypress["']\s*\/>\n?/g,
      ''
    );
  }
  if (isWdio) {
    result = result.replace(
      /import\s+\{[^}]*\}\s+from\s+['"]@wdio\/globals['"];?\n?/g,
      ''
    );
  }
  if (isPuppeteer) {
    result = result.replace(
      /const\s+puppeteer\s*=\s*require\s*\(\s*['"]puppeteer['"]\s*\)\s*;?\n?/g,
      ''
    );
    result = result.replace(
      /import\s+puppeteer\s+from\s+['"]puppeteer['"];?\n?/g,
      ''
    );
    // Remove browser launch/newPage/close boilerplate
    result = result.replace(
      /const\s+browser\s*=\s*await\s+puppeteer\.launch\([^)]*\)\s*;?\n?/g,
      ''
    );
    result = result.replace(
      /const\s+page\s*=\s*await\s+browser\.newPage\(\)\s*;?\n?/g,
      ''
    );
    result = result.replace(/await\s+browser\.close\(\)\s*;?\n?/g, '');
  }
  if (isTestCafe) {
    result = result.replace(
      /import\s+\{[^}]*\}\s+from\s+['"]testcafe['"];?\n?/g,
      ''
    );
  }

  // Add Selenium import at the top if not already present
  if (
    !/require\s*\(\s*['"]selenium-webdriver['"]/.test(result) &&
    !/from\s+['"]selenium-webdriver['"]/.test(result)
  ) {
    result =
      "const { Builder, By, Key, until } = require('selenium-webdriver');\n\n" +
      result;
  }

  return result;
}

// ── Phase 7: Driver boilerplate ──

function addDriverBoilerplate(content) {
  let result = content;

  // Only add if there's no existing driver setup
  if (!/let\s+driver\b/.test(result) && /\bdriver\./.test(result)) {
    // Find the first describe block and add beforeAll/afterAll inside
    const describeMatch = result.match(
      /(describe\s*\([^,]+,\s*(?:async\s*)?\(\)\s*=>\s*\{)\n/
    );
    if (describeMatch) {
      const driverSetup = `\n  let driver;\n\n  beforeAll(async () => {\n    driver = await new Builder().forBrowser('chrome').build();\n  });\n\n  afterAll(async () => {\n    await driver.quit();\n  });\n`;
      result = result.replace(describeMatch[0], describeMatch[0] + driverSetup);
    }
  }

  return result;
}

// ── Phase 8: Cleanup ──

function cleanupOutput(content) {
  return (
    content
      .replace(/await\s+await/g, 'await')
      .replace(/\n{3,}/g, '\n\n')
      .trim() + '\n'
  );
}

export default {
  name: 'selenium',
  language: 'javascript',
  paradigm: 'bdd-e2e',
  detect,
  parse,
  emit,
  imports: {
    packages: ['selenium-webdriver'],
    globals: [
      'describe',
      'it',
      'beforeAll',
      'afterAll',
      'beforeEach',
      'afterEach',
      'expect',
    ],
  },
};
