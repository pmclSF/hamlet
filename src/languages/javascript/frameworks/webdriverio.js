/**
 * WebdriverIO framework definition.
 *
 * Provides detect, parse, and emit for the WebdriverIO E2E testing framework.
 * emit() transforms Playwright and Cypress source code into WebdriverIO code.
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

  // WDIO-specific imports (strong signals)
  if (/from\s+['"]@wdio\/globals['"]/.test(source)) score += 40;
  if (/from\s+['"]webdriverio['"]/.test(source)) score += 35;

  // browser.* API
  if (/\bbrowser\.url\s*\(/.test(source)) score += 20;
  if (/\bbrowser\.execute\s*\(/.test(source)) score += 5;
  if (/\bbrowser\.pause\s*\(/.test(source)) score += 5;
  if (/\bbrowser\.getTitle\s*\(/.test(source)) score += 5;
  if (/\bbrowser\.keys\s*\(/.test(source)) score += 5;

  // WDIO element selectors
  if (/\$\(\s*['"`]/.test(source) && /\.setValue\s*\(/.test(source))
    score += 20;
  if (/\$\$\s*\(/.test(source)) score += 10;

  // WDIO element actions
  if (/\.setValue\s*\(/.test(source)) score += 15;
  if (/\.getText\s*\(/.test(source)) score += 10;
  if (/\.isDisplayed\s*\(/.test(source)) score += 10;
  if (/\.waitForDisplayed\s*\(/.test(source)) score += 10;
  if (/\.moveTo\s*\(/.test(source)) score += 5;

  // WDIO assertions
  if (/expect\(browser\)\.toHave/.test(source)) score += 15;
  if (/toBeDisplayed\(\)/.test(source)) score += 10;
  if (/toHaveUrl\(/.test(source)) score += 10;

  // Negative: Cypress
  if (/\bcy\./.test(source)) score -= 30;
  // Negative: Playwright
  if (/\bpage\.goto\s*\(/.test(source)) score -= 30;
  if (/from\s+['"]@playwright\/test['"]/.test(source)) score -= 30;
  if (/\bpage\.locator\s*\(/.test(source)) score -= 20;
  // Negative: TestCafe
  if (/\bSelector\s*\(/.test(source) && /\bfixture\s*`/.test(source))
    score -= 20;
  // Negative: Puppeteer
  if (/\bpuppeteer\.launch/.test(source)) score -= 20;

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

    if (/\bbrowser\./.test(trimmed)) {
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

    if (/\$\$?\s*\(/.test(trimmed)) {
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
 * Emit WebdriverIO code from IR + original source.
 *
 * Handles Playwright→WDIO and Cypress→WDIO conversions.
 *
 * @param {TestFile} _ir - Parsed IR tree (for scoring metadata)
 * @param {string} source - Original source code
 * @returns {string} Converted WebdriverIO source code
 */
function emit(_ir, source) {
  let result = source;

  // Strip incoming HAMLET-TODO blocks (from previous round-trip step)
  result = result.replace(
    /^[ \t]*\/\/ HAMLET-TODO \[[^\]]+\]:.*\n(?:[ \t]*\n)*(?:[ \t]*\/\/ (?:Original|Manual action required):.*\n(?:[ \t]*\n)*)*/gm,
    ''
  );
  result = result.replace(
    /^[ \t]*\/\*\s*HAMLET-TODO:.*?\*\/\s*\n?/gm,
    ''
  );

  const isPlaywrightSource =
    /from\s+['"]@playwright\/test['"]/.test(source) ||
    /\bpage\.goto\s*\(/.test(source);
  const isCypressSource = /\bcy\./.test(source);

  // Phase 1: Remove source imports
  result = removeSourceImports(result, isPlaywrightSource, isCypressSource);

  // Phase 2: Convert Playwright patterns to WDIO
  result = convertPlaywrightToWdio(result);

  // Phase 3: Convert Cypress patterns to WDIO
  result = convertCypressToWdio(result);

  // Phase 4: Convert test structure
  result = convertTestStructure(result, isPlaywrightSource);

  // Phase 5: Add async/await to Cypress-sourced callbacks
  if (isCypressSource) {
    result = addAsyncAwait(result);
  }

  // Phase 6: Cleanup
  result = cleanupOutput(result);

  return result;
}

/**
 * Remove source framework imports.
 */
function removeSourceImports(content, isPlaywright, isCypress) {
  let result = content;

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

  return result;
}

/**
 * Convert Playwright commands to WDIO equivalents.
 */
function convertPlaywrightToWdio(content) {
  let result = content;

  // --- Assertion patterns (most specific first) ---

  result = result.replace(
    /await expect\(page\)\.toHaveURL\(([^)]+)\)/g,
    'await expect(browser).toHaveUrl($1)'
  );
  result = result.replace(
    /await expect\(page\)\.toHaveTitle\(([^)]+)\)/g,
    'await expect(browser).toHaveTitle($1)'
  );
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toBeVisible\(\)/g,
    'await expect($($1)).toBeDisplayed()'
  );
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toBeHidden\(\)/g,
    'await expect($($1)).not.toBeDisplayed()'
  );
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toBeAttached\(\)/g,
    'await expect($($1)).toExist()'
  );
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.not\.toBeAttached\(\)/g,
    'await expect($($1)).not.toExist()'
  );
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toHaveText\(([^)]+)\)/g,
    'await expect($($1)).toHaveText($2)'
  );
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toContainText\(([^)]+)\)/g,
    'await expect($($1)).toHaveTextContaining($2)'
  );
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toHaveValue\(([^)]+)\)/g,
    'await expect($($1)).toHaveValue($2)'
  );
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toHaveCount\(([^)]+)\)/g,
    'await expect($$$$($1)).toBeElementsArrayOfSize($2)'
  );
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toBeChecked\(\)/g,
    'await expect($($1)).toBeSelected()'
  );
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toBeEnabled\(\)/g,
    'await expect($($1)).toBeEnabled()'
  );
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toBeDisabled\(\)/g,
    'await expect($($1)).toBeDisabled()'
  );
  result = result.replace(
    /await expect\(page\.locator\(([^)]+)\)\)\.toHaveAttribute\(([^,]+),\s*([^)]+)\)/g,
    'await expect($($1)).toHaveAttribute($2, $3)'
  );

  // --- Composite action patterns ---

  result = result.replace(
    /await page\.locator\(([^)]+)\)\.fill\(([^)]+)\)/g,
    'await $($1).setValue($2)'
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.click\(\)/g,
    'await $($1).click()'
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.dblclick\(\)/g,
    'await $($1).doubleClick()'
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.hover\(\)/g,
    'await $($1).moveTo()'
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.textContent\(\)/g,
    'await $($1).getText()'
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.isVisible\(\)/g,
    'await $($1).isDisplayed()'
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.waitFor\(\)/g,
    'await $($1).waitForDisplayed()'
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.waitFor\(\{\s*state:\s*['"]visible['"]\s*\}\)/g,
    'await $($1).waitForDisplayed()'
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.clear\(\)/g,
    'await $($1).clearValue()'
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.selectOption\(\{\s*label:\s*([^}]+)\}\)/g,
    'await $($1).selectByVisibleText($2)'
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.selectOption\(([^)]+)\)/g,
    "await $($1).selectByAttribute('value', $2)"
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.check\(\)/g,
    'await $($1).click()'
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.uncheck\(\)/g,
    'await $($1).click()'
  );

  // Standalone page.locator -> $
  result = result.replace(/page\.locator\(([^)]+)\)/g, '$($1)');

  // --- Navigation ---

  result = result.replace(
    /await page\.goto\(([^)]+)\)/g,
    'await browser.url($1)'
  );

  // --- Browser API ---

  result = result.replace(
    /await page\.waitForTimeout\(([^)]+)\)/g,
    'await browser.pause($1)'
  );
  result = result.replace(/await page\.evaluate\(/g, 'await browser.execute(');
  result = result.replace(/await page\.title\(\)/g, 'await browser.getTitle()');
  result = result.replace(/await page\.url\(\)/g, 'await browser.getUrl()');
  result = result.replace(/await page\.reload\(\)/g, 'await browser.refresh()');
  result = result.replace(/await page\.goBack\(\)/g, 'await browser.back()');
  result = result.replace(
    /await page\.goForward\(\)/g,
    'await browser.forward()'
  );
  result = result.replace(
    /await page\.keyboard\.press\(([^)]+)\)/g,
    'await browser.keys([$1])'
  );
  result = result.replace(
    /await page\.setViewportSize\(([^)]+)\)/g,
    'await browser.setWindowSize($1)'
  );

  // --- Cookies ---

  result = result.replace(
    /await context\.addCookies\(/g,
    'await browser.setCookies('
  );
  result = result.replace(
    /await context\.cookies\(\)/g,
    'await browser.getCookies()'
  );
  result = result.replace(
    /await context\.clearCookies\(\)/g,
    'await browser.deleteCookies()'
  );

  // --- getByText ---

  result = result.replace(/page\.getByText\(([^)]+)\)/g, (_, arg) => {
    const text = arg.replace(/^['"]|['"]$/g, '');
    return `$(\`*=${text}\`)`;
  });

  // --- Unconvertible: page.route ---

  result = result.replace(
    /await page\.route\([^)]+,\s*[^)]+\)/g,
    (match) =>
      formatter.formatTodo({
        id: 'UNCONVERTIBLE-ROUTE',
        description: 'Playwright page.route() has no direct WDIO equivalent',
        original: match.trim(),
        action:
          'Use a mock server (e.g., msw or wiremock) for network interception',
      }) +
      '\n// ' +
      match.trim()
  );

  return result;
}

/**
 * Convert Cypress commands to WDIO equivalents.
 */
function convertCypressToWdio(content) {
  let result = content;

  // --- Composite cy.get().should() assertion chains ---

  result = result.replace(
    /cy\.get\(([^)]+)\)\.should\(['"]be\.visible['"]\)/g,
    'await expect($($1)).toBeDisplayed()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.should\(['"]not\.be\.visible['"]\)/g,
    'await expect($($1)).not.toBeDisplayed()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.should\(['"]exist['"]\)/g,
    'await expect($($1)).toExist()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.should\(['"]not\.exist['"]\)/g,
    'await expect($($1)).not.toExist()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.should\(['"]have\.text['"],\s*([^)]+)\)/g,
    'await expect($($1)).toHaveText($2)'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.should\(['"]contain['"],\s*([^)]+)\)/g,
    'await expect($($1)).toHaveTextContaining($2)'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.should\(['"]have\.value['"],\s*([^)]+)\)/g,
    'await expect($($1)).toHaveValue($2)'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.should\(['"]have\.length['"],\s*(\d+)\)/g,
    'await expect($$$$($1)).toBeElementsArrayOfSize($2)'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.should\(['"]be\.checked['"]\)/g,
    'await expect($($1)).toBeSelected()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.should\(['"]be\.disabled['"]\)/g,
    'await expect($($1)).toBeDisabled()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.should\(['"]be\.enabled['"]\)/g,
    'await expect($($1)).toBeEnabled()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.should\(['"]have\.attr['"],\s*([^,]+),\s*([^)]+)\)/g,
    'await expect($($1)).toHaveAttribute($2, $3)'
  );

  // --- Composite cy.get().action() chains ---

  // .clear().type() combined → setValue (must be before individual .clear() and .type())
  result = result.replace(
    /cy\.get\(([^)]+)\)\.clear\(\)\.type\(([^)]+)\)/g,
    'await $($1).setValue($2)'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.type\(([^)]+)\)/g,
    'await $($1).setValue($2)'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.click\(\)/g,
    'await $($1).click()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.dblclick\(\)/g,
    'await $($1).doubleClick()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.clear\(\)/g,
    'await $($1).clearValue()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.select\(([^)]+)\)/g,
    'await $($1).selectByVisibleText($2)'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.check\(\)/g,
    'await $($1).click()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.trigger\(['"]mouseover['"]\)/g,
    'await $($1).moveTo()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.invoke\(['"]text['"]\)/g,
    'await $($1).getText()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.invoke\(['"]attr['"],\s*([^)]+)\)/g,
    'await $($1).getAttribute($2)'
  );

  // --- cy.contains ---

  result = result.replace(/cy\.contains\(([^)]+)\)\.click\(\)/g, (_, arg) => {
    const text = arg.replace(/^['"]|['"]$/g, '');
    return `await $(\`*=${text}\`).click()`;
  });
  result = result.replace(/cy\.contains\(([^)]+)\)/g, (_, arg) => {
    const text = arg.replace(/^['"]|['"]$/g, '');
    return `$(\`*=${text}\`)`;
  });

  // --- Navigation ---

  result = result.replace(/cy\.visit\(([^)]+)\)/g, 'await browser.url($1)');
  result = result.replace(/cy\.reload\(\)/g, 'await browser.refresh()');
  result = result.replace(/cy\.go\(['"]back['"]\)/g, 'await browser.back()');
  result = result.replace(
    /cy\.go\(['"]forward['"]\)/g,
    'await browser.forward()'
  );

  // --- URL/Title assertions ---

  result = result.replace(
    /cy\.url\(\)\.should\(['"]include['"],\s*([^)]+)\)/g,
    'await expect(browser).toHaveUrlContaining($1)'
  );
  result = result.replace(
    /cy\.url\(\)\.should\(['"]eq['"],\s*([^)]+)\)/g,
    'await expect(browser).toHaveUrl($1)'
  );
  result = result.replace(
    /cy\.title\(\)\.should\(['"]eq['"],\s*([^)]+)\)/g,
    'await expect(browser).toHaveTitle($1)'
  );

  // --- Waits ---

  result = result.replace(/cy\.wait\((\d+)\)/g, 'await browser.pause($1)');

  // --- Simple commands ---

  result = result.replace(
    /cy\.clearCookies\(\)/g,
    'await browser.deleteCookies()'
  );
  result = result.replace(
    /cy\.getCookies\(\)/g,
    'await browser.getCookies()'
  );
  result = result.replace(
    /cy\.clearLocalStorage\(\)/g,
    'await browser.execute(() => localStorage.clear())'
  );
  result = result.replace(/cy\.log\(([^)]+)\)/g, 'console.log($1)');

  // --- Window/eval ---

  result = result.replace(
    /cy\.window\(\)\.then\(([^)]+)\)/g,
    'await browser.execute($1)'
  );

  // --- Intercept -> HAMLET-TODO ---

  result = result.replace(
    /cy\.intercept\([^)]+(?:,[^)]+)?\)(?:\.as\(['"][^'"]+['"]\))?/g,
    (match) =>
      formatter.formatTodo({
        id: 'UNCONVERTIBLE-INTERCEPT',
        description: 'Cypress cy.intercept() has no direct WDIO equivalent',
        original: match.trim(),
        action:
          'Use a mock server (e.g., msw or wiremock) for network interception',
      }) +
      '\n// ' +
      match.trim()
  );

  return result;
}

/**
 * Convert test structure (Playwright test.describe/test -> describe/it).
 */
function convertTestStructure(content, isPlaywright) {
  let result = content;

  if (isPlaywright) {
    // Playwright -> WDIO: test.describe -> describe, test() -> it()
    result = result.replace(/test\.describe\.only\(/g, 'describe.only(');
    result = result.replace(/test\.describe\.skip\(/g, 'describe.skip(');
    result = result.replace(/test\.describe\(/g, 'describe(');
    result = result.replace(/test\.only\(/g, 'it.only(');
    result = result.replace(/test\.skip\(/g, 'it.skip(');
    result = result.replace(/test\.beforeAll\(/g, 'before(');
    result = result.replace(/test\.afterAll\(/g, 'after(');
    result = result.replace(/test\.beforeEach\(/g, 'beforeEach(');
    result = result.replace(/test\.afterEach\(/g, 'afterEach(');
    // test( -> it( (must be after all test.* prefixed patterns)
    result = result.replace(/\btest\(([^,()\n]+),/g, 'it($1,');

    // Remove { page } / { page, request } destructure from callback params
    result = result.replace(
      /\(\s*\{\s*page\s*(?:,\s*request\s*)?\}\s*\)\s*=>/g,
      '() =>'
    );
  }
  // Cypress: describe/it stays as-is for WDIO (Mocha-based structure)

  return result;
}

/**
 * Add async/await to Cypress-sourced callbacks (Cypress is sync, WDIO is async).
 */
function addAsyncAwait(content) {
  let result = content;

  // Add async to test callbacks that don't already have it
  result = result.replace(
    /((?:it|test)\s*\([^,]+,\s*)(?!async)\(\s*\)\s*=>\s*\{/g,
    '$1async () => {'
  );
  result = result.replace(
    /((?:beforeEach|afterEach|before|after)\s*\(\s*)(?!async)\(\s*\)\s*=>\s*\{/g,
    '$1async () => {'
  );

  return result;
}

/**
 * Clean up output.
 */
function cleanupOutput(content) {
  return (
    content
      .replace(/await\s+await/g, 'await')
      .replace(/\n{3,}/g, '\n\n')
      .trim() + '\n'
  );
}

export default {
  name: 'webdriverio',
  language: 'javascript',
  paradigm: 'bdd-e2e',
  detect,
  parse,
  emit,
  imports: {
    globals: [
      'describe',
      'it',
      'before',
      'after',
      'beforeEach',
      'afterEach',
      '$',
      '$$',
      'browser',
      'expect',
    ],
    from: '@wdio/globals',
    mockNamespace: null,
  },
};
