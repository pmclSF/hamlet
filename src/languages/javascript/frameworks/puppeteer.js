/**
 * Puppeteer framework definition.
 *
 * Provides detect, parse, and emit for the Puppeteer browser automation library.
 * emit() transforms Playwright source code into Puppeteer code.
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

  // Puppeteer imports (strong signals)
  if (/require\(['"]puppeteer['"]\)/.test(source)) score += 30;
  if (/from\s+['"]puppeteer['"]/.test(source)) score += 30;

  // Puppeteer lifecycle
  if (/puppeteer\.launch\s*\(/.test(source)) score += 25;
  if (/browser\.newPage\s*\(/.test(source)) score += 15;
  if (/browser\.close\s*\(/.test(source)) score += 5;

  // Puppeteer page API
  if (/page\.\$\s*\(/.test(source)) score += 10;
  if (/page\.\$\$\s*\(/.test(source)) score += 10;
  if (/page\.\$eval\s*\(/.test(source)) score += 10;
  if (/page\.\$\$eval\s*\(/.test(source)) score += 10;
  if (/page\.type\s*\(/.test(source)) score += 10;
  if (/page\.click\s*\(/.test(source)) score += 5;
  if (/page\.waitForSelector\s*\(/.test(source)) score += 10;
  if (/page\.setViewport\s*\(/.test(source)) score += 10;

  // Negative: Playwright (locator API is the key difference)
  if (/page\.locator\s*\(/.test(source)) score -= 30;
  if (/from\s+['"]@playwright\/test['"]/.test(source)) score -= 30;
  // Negative: Cypress
  if (/\bcy\./.test(source)) score -= 30;
  // Negative: TestCafe
  if (/\bSelector\s*\(/.test(source) && /\bfixture\s*`/.test(source))
    score -= 20;
  // Negative: WDIO
  if (/\bbrowser\.url\s*\(/.test(source)) score -= 10;

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

    if (/\bpage\./.test(trimmed) || /\bpuppeteer\./.test(trimmed)) {
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
 * Emit Puppeteer code from IR + original Playwright source.
 *
 * @param {TestFile} _ir - Parsed IR tree
 * @param {string} source - Original Playwright source code
 * @returns {string} Converted Puppeteer source code
 */
function emit(_ir, source) {
  let result = source;

  const isPlaywrightSource =
    /from\s+['"]@playwright\/test['"]/.test(source) ||
    /\bpage\.locator\s*\(/.test(source);

  if (!isPlaywrightSource) {
    return source;
  }

  // Phase 1: Remove Playwright imports
  result = result.replace(
    /import\s+\{[^}]*\}\s+from\s+['"]@playwright\/test['"];?\n?/g,
    ''
  );

  // Phase 2: Convert Playwright assertions to Jest-style manual assertions
  result = convertPlaywrightAssertions(result);

  // Phase 3: Convert Playwright locator actions to Puppeteer page-level actions
  result = convertPlaywrightActions(result);

  // Phase 4: Convert Playwright navigation/browser API
  result = convertPlaywrightBrowserApi(result);

  // Phase 5: Convert test structure (test.describe -> describe, test -> it)
  result = convertTestStructure(result);

  // Phase 6: Remove { page } parameter
  result = result.replace(
    /\(\s*\{\s*page\s*(?:,\s*request\s*)?\}\s*\)\s*=>/g,
    '() =>'
  );

  // Phase 7: Add Puppeteer lifecycle boilerplate
  result = addPuppeteerLifecycle(result);

  // Phase 8: Add Puppeteer import
  result = "const puppeteer = require('puppeteer');\n\n" + result;

  // Phase 9: Cleanup
  result =
    result
      .replace(/await\s+await/g, 'await')
      .replace(/\n{3,}/g, '\n\n')
      .trim() + '\n';

  return result;
}

/**
 * Convert Playwright expect assertions to Jest manual assertions.
 */
function convertPlaywrightAssertions(content) {
  let result = content;

  // await expect(page).toHaveURL(url) -> expect(page.url()).toBe(url)
  result = result.replace(
    /await expect\(page\)\.toHaveURL\(([^)]+)\)/g,
    'expect(page.url()).toBe($1)'
  );
  // await expect(page).toHaveTitle(title) -> expect(await page.title()).toBe(title)
  result = result.replace(
    /await expect\(page\)\.toHaveTitle\(([^)]+)\)/g,
    'expect(await page.title()).toBe($1)'
  );
  // await expect(page.locator(sel)).toBeVisible() -> expect(await page.$(sel)).toBeTruthy()
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toBeVisible\(\)/g,
    'expect(await page.$($1)).toBeTruthy()'
  );
  // await expect(page.locator(sel)).toBeHidden() -> expect(await page.$(sel)).toBeFalsy()
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toBeHidden\(\)/g,
    'expect(await page.$($1)).toBeFalsy()'
  );
  // await expect(page.locator(sel)).toBeAttached() -> expect(await page.$(sel)).toBeTruthy()
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toBeAttached\(\)/g,
    'expect(await page.$($1)).toBeTruthy()'
  );
  // await expect(page.locator(sel)).toHaveText(text) -> expect(await page.$eval(sel, el => el.textContent)).toBe(text)
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toHaveText\(([^)]+)\)/g,
    'expect(await page.$eval($1, el => el.textContent)).toBe($2)'
  );
  // await expect(page.locator(sel)).toContainText(text) -> expect(await page.$eval(sel, el => el.textContent)).toContain(text)
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toContainText\(([^)]+)\)/g,
    'expect(await page.$eval($1, el => el.textContent)).toContain($2)'
  );
  // await expect(page.locator(sel)).toHaveValue(val) -> expect(await page.$eval(sel, el => el.value)).toBe(val)
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toHaveValue\(([^)]+)\)/g,
    'expect(await page.$eval($1, el => el.value)).toBe($2)'
  );
  // await expect(page.locator(sel)).toHaveCount(n) -> expect((await page.$$(sel)).length).toBe(n)
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toHaveCount\(([^)]+)\)/g,
    'expect((await page.$$($1)).length).toBe($2)'
  );
  // await expect(page.locator(sel)).toBeChecked() -> expect(await page.$eval(sel, el => el.checked)).toBe(true)
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toBeChecked\(\)/g,
    'expect(await page.$eval($1, el => el.checked)).toBe(true)'
  );
  // await expect(page.locator(sel)).toHaveAttribute(attr, val) -> expect(await page.$eval(sel, (el, a) => el.getAttribute(a), attr)).toBe(val)
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toHaveAttribute\(([^,]+),\s*([^)]+)\)/g,
    'expect(await page.$eval($1, (el, a) => el.getAttribute(a), $2)).toBe($3)'
  );

  return result;
}

/**
 * Convert Playwright locator actions to Puppeteer page-level actions.
 */
function convertPlaywrightActions(content) {
  let result = content;

  // await page.locator(sel).fill(text) -> await page.type(sel, text)
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.fill\(([^)]+)\)/g,
    'await page.type($1, $2)'
  );
  // await page.locator(sel).click() -> await page.click(sel)
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.click\(\)/g,
    'await page.click($1)'
  );
  // await page.locator(sel).dblclick() -> await page.click(sel, { clickCount: 2 })
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.dblclick\(\)/g,
    'await page.click($1, { clickCount: 2 })'
  );
  // await page.locator(sel).hover() -> await page.hover(sel)
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.hover\(\)/g,
    'await page.hover($1)'
  );
  // await page.locator(sel).textContent() -> await page.$eval(sel, el => el.textContent)
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.textContent\(\)/g,
    'await page.$eval($1, el => el.textContent)'
  );
  // await page.locator(sel).isVisible() -> !!(await page.$(sel))
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.isVisible\(\)/g,
    '!!(await page.$($1))'
  );
  // await page.locator(sel).waitFor() -> await page.waitForSelector(sel)
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.waitFor\(\)/g,
    'await page.waitForSelector($1)'
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.waitFor\(\{[^}]*\}\)/g,
    'await page.waitForSelector($1)'
  );
  // await page.locator(sel).evaluate(fn) -> await page.$eval(sel, fn)
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.evaluate\(([^)]+)\)/g,
    'await page.$eval($1, $2)'
  );
  // await page.locator(sel).evaluateAll(fn) -> await page.$$eval(sel, fn)
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.evaluateAll\(([^)]+)\)/g,
    'await page.$$eval($1, $2)'
  );
  // await page.locator(sel).selectOption(val) -> await page.select(sel, val)
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.selectOption\(([^)]+)\)/g,
    'await page.select($1, $2)'
  );
  // await page.locator(sel).clear() -> await page.click(sel, { clickCount: 3 }); await page.keyboard.press('Backspace')
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.clear\(\)/g,
    "await page.click($1, { clickCount: 3 });\n    await page.keyboard.press('Backspace')"
  );

  // Standalone page.locator -> page.$ (catch remaining)
  result = result.replace(/page\.locator\(([^)]+)\)/g, 'page.$($1)');

  return result;
}

/**
 * Convert Playwright browser API to Puppeteer equivalents.
 */
function convertPlaywrightBrowserApi(content) {
  let result = content;

  // await page.setViewportSize({w, h}) -> await page.setViewport({width: w, height: h})
  result = result.replace(
    /await page\.setViewportSize\(\{\s*width:\s*(\d+),\s*height:\s*(\d+)\s*\}\)/g,
    'await page.setViewport({ width: $1, height: $2 })'
  );

  // await page.screenshot() -> await page.screenshot()  (same API, passthrough)
  // await page.screenshot({ path: p }) -> await page.screenshot({ path: p })  (same)

  // Cookie conversion
  result = result.replace(
    /await page\.context\(\)\.addCookies\(/g,
    'await page.setCookie('
  );
  result = result.replace(
    /await page\.context\(\)\.cookies\(\)/g,
    'await page.cookies()'
  );
  result = result.replace(
    /await page\.context\(\)\.clearCookies\(\)/g,
    'await page.deleteCookie()'
  );
  result = result.replace(
    /await context\.addCookies\(/g,
    'await page.setCookie('
  );
  result = result.replace(
    /await context\.cookies\(\)/g,
    'await page.cookies()'
  );
  result = result.replace(
    /await context\.clearCookies\(\)/g,
    'await page.deleteCookie()'
  );

  // Unconvertible: page.route -> HAMLET-TODO
  result = result.replace(
    /await page\.route\([^)]+,\s*[^)]+\)/g,
    (match) =>
      formatter.formatTodo({
        id: 'UNCONVERTIBLE-ROUTE',
        description:
          'Playwright page.route() requires Puppeteer page.setRequestInterception()',
        original: match.trim(),
        action:
          "Use page.setRequestInterception(true) and page.on('request', ...) pattern",
      }) +
      '\n// ' +
      match.trim()
  );

  return result;
}

/**
 * Convert Playwright test structure to Jest/Mocha describe/it.
 */
function convertTestStructure(content) {
  let result = content;

  result = result.replace(/test\.describe\.only\(/g, 'describe.only(');
  result = result.replace(/test\.describe\.skip\(/g, 'describe.skip(');
  result = result.replace(/test\.describe\(/g, 'describe(');
  result = result.replace(/test\.only\(/g, 'it.only(');
  result = result.replace(/test\.skip\(/g, 'it.skip(');
  result = result.replace(/test\.beforeAll\(/g, 'beforeAll(');
  result = result.replace(/test\.afterAll\(/g, 'afterAll(');
  result = result.replace(/test\.beforeEach\(/g, 'beforeEach(');
  result = result.replace(/test\.afterEach\(/g, 'afterEach(');
  // test( -> it( (after all test.* prefixed)
  result = result.replace(/\btest\(([^,()\n]+),/g, 'it($1,');

  return result;
}

/**
 * Add Puppeteer browser lifecycle boilerplate.
 */
function addPuppeteerLifecycle(content) {
  // Check if there's already a describe block at the top level
  const hasDescribe = /\bdescribe\s*\(/.test(content);

  if (!hasDescribe) {
    return content;
  }

  // Insert lifecycle variables and hooks after the first describe opening
  const lifecycle = `  let browser, page;

  beforeAll(async () => {
    browser = await puppeteer.launch();
    page = await browser.newPage();
  });

  afterAll(async () => {
    await browser.close();
  });`;

  // Insert after the first describe(...) { line
  const result = content.replace(
    /(describe\s*\([^)]+,\s*(?:async\s*)?\(\s*\)\s*=>\s*\{)\n/,
    `$1\n${lifecycle}\n\n`
  );

  return result;
}

export default {
  name: 'puppeteer',
  language: 'javascript',
  paradigm: 'bdd-e2e',
  detect,
  parse,
  emit,
  imports: {
    explicit: ['puppeteer'],
    from: 'puppeteer',
    mockNamespace: null,
  },
};
