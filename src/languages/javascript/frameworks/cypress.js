/**
 * Cypress framework definition.
 *
 * Provides detect, parse, and emit for the Cypress E2E testing framework.
 * emit() handles conversions from WebdriverIO and TestCafe into Cypress code.
 * parse() builds a flat IR tree (one node per line, no nesting) from Cypress
 * source code. The IR is consumed by ConfidenceScorer for scoring.
 */

import {
  TestFile,
  TestSuite,
  TestCase,
  Hook,
  Assertion,
  MockCall,
  ImportStatement,
  RawCode,
  Comment,
  Modifier,
} from '../../../core/ir.js';

import { TodoFormatter } from '../../../core/TodoFormatter.js';

const formatter = new TodoFormatter('javascript');

function detect(source) {
  if (!source || !source.trim()) return 0;

  let score = 0;

  if (/\bcy\./.test(source)) score += 30;
  if (/\bcy\.visit\s*\(/.test(source)) score += 15;
  if (/\bcy\.get\s*\(/.test(source)) score += 15;
  if (/\bcy\.contains\s*\(/.test(source)) score += 10;
  if (/\bcy\.intercept\s*\(/.test(source)) score += 10;
  if (/\bcy\.request\s*\(/.test(source)) score += 5;
  if (/\.should\s*\(/.test(source)) score += 10;
  if (/\bCypress\./.test(source)) score += 10;
  if (/\bdescribe\s*\(/.test(source)) score += 3;
  if (/\bit\s*\(/.test(source)) score += 3;

  // Negative: Playwright
  if (/from\s+['"]@playwright\/test['"]/.test(source)) score -= 40;
  if (/\bpage\./.test(source)) score -= 20;

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

    if (/\.should\s*\(/.test(trimmed) || /\bexpect\s*\(/.test(trimmed)) {
      body.push(
        new Assertion({
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    if (/\bcy\./.test(trimmed)) {
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
 * Emit Cypress code from IR + original source.
 *
 * Handles WebdriverIO→Cypress and TestCafe→Cypress conversions.
 * Each source framework's patterns are isolated in a separate function
 * and gated by source detection to prevent phase interference.
 *
 * Note: _ir is currently unused — conversion operates on the source string
 * via regex. The IR is consumed by ConfidenceScorer for scoring only.
 * Future work: reconstruct output from the IR tree.
 *
 * @param {TestFile} _ir - Parsed IR tree (used by ConfidenceScorer, not here)
 * @param {string} source - Original source code
 * @returns {string} Converted Cypress source code
 */
function emit(_ir, source) {
  let result = source;

  // Detect source framework
  const isPlaywrightSource =
    /from\s+['"]@playwright\/test['"]/.test(source) ||
    /\bpage\.goto\s*\(/.test(source);
  const isSeleniumSource =
    /require\s*\(\s*['"]selenium-webdriver['"]/.test(source) ||
    /from\s+['"]selenium-webdriver['"]/.test(source);
  const isWdioSource =
    /\bbrowser\.url\s*\(/.test(source) ||
    (/\$\(/.test(source) && /\.setValue\s*\(/.test(source));
  const isPuppeteerSource = /\bpuppeteer\.launch/.test(source);
  const isTestCafeSource =
    /\bfixture\s*`/.test(source) || /from\s+['"]testcafe['"]/.test(source);

  // Phase 1: Remove source-framework imports
  if (isPlaywrightSource) {
    result = result.replace(
      /import\s+\{[^}]*\}\s+from\s+['"]@playwright\/test['"];?\n?/g,
      ''
    );
  }
  if (isSeleniumSource) {
    result = result.replace(
      /(?:const|let|var)\s+\{[^}]*\}\s*=\s*require\s*\(\s*['"]selenium-webdriver['"]\s*\)\s*;?\n?/g,
      ''
    );
    result = result.replace(
      /import\s+\{[^}]*\}\s+from\s+['"]selenium-webdriver['"];?\n?/g,
      ''
    );
    // Remove driver setup/teardown boilerplate
    result = result.replace(/\s*let\s+driver\s*;\s*\n?/g, '\n');
    result = result.replace(
      /\s*beforeAll\s*\(\s*async\s*\(\)\s*=>\s*\{[^}]*new\s+Builder[^}]*\}\s*\)\s*;?\n?/g,
      '\n'
    );
    result = result.replace(
      /\s*afterAll\s*\(\s*async\s*\(\)\s*=>\s*\{[^}]*driver\.quit[^}]*\}\s*\)\s*;?\n?/g,
      '\n'
    );
  }
  if (isWdioSource) {
    result = result.replace(
      /import\s+\{[^}]*\}\s+from\s+['"]@wdio\/globals['"];?\n?/g,
      ''
    );
    result = result.replace(
      /import\s+\{[^}]*\}\s+from\s+['"]webdriverio['"];?\n?/g,
      ''
    );
  }
  if (isPuppeteerSource) {
    result = result.replace(
      /const\s+puppeteer\s*=\s*require\s*\(\s*['"]puppeteer['"]\s*\)\s*;?\n?/g,
      ''
    );
    result = result.replace(
      /import\s+puppeteer\s+from\s+['"]puppeteer['"];?\n?/g,
      ''
    );
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
  if (isTestCafeSource) {
    result = result.replace(
      /import\s+\{[^}]*\}\s+from\s+['"]testcafe['"];?\n?/g,
      ''
    );
  }

  // Phase 2: Convert source commands
  if (isPlaywrightSource) {
    result = convertPlaywrightToCypress(result);
  }
  if (isSeleniumSource) {
    result = convertSeleniumToCypress(result);
  }
  if (isWdioSource) {
    result = convertWdioToCypress(result);
  }
  if (isPuppeteerSource) {
    result = convertPuppeteerToCypress(result);
  }
  if (isTestCafeSource) {
    result = convertTestCafeToCypress(result);
  }

  // Phase 3: Convert test structure
  if (isPlaywrightSource) {
    result = convertPlaywrightStructure(result);
  }
  if (isTestCafeSource) {
    result = convertTestCafeStructure(result);
  }
  // WDIO/Selenium/Puppeteer use describe/it — same as Cypress

  // Phase 3.5: Re-split long chains into multi-line format
  // Forward conversion joins multi-line chains for regex matching;
  // this restores multi-line style for readability and fidelity.
  if (isPlaywrightSource) {
    result = resplitChains(result);
  }

  // Phase 4: Remove async/await (Cypress is synchronous)
  result = removeAsyncAwait(result);

  // Phase 4.5: Post-async cleanup — fix patterns that depend on await removal
  if (isPlaywrightSource) {
    // cy.location('hash') → cy.hash() (Cypress has a dedicated command)
    result = result.replace(/cy\.location\('hash'\)/g, 'cy.hash()');
    // Remaining new URL(cy.url()).prop after async strip
    result = result.replace(
      /new URL\(cy\.url\(\)\)\.(\w+)/g,
      "cy.location('$1')"
    );
    // Clean up location('hash') that was just created
    result = result.replace(/cy\.location\('hash'\)/g, 'cy.hash()');
  }

  // Phase 5: Clean up
  result = result.replace(/\n{3,}/g, '\n\n').trim() + '\n';

  // Phase 6: Add Cypress reference comment (for Playwright/Selenium/Puppeteer sources)
  if (isPlaywrightSource || isSeleniumSource || isPuppeteerSource) {
    if (!result.includes('/// <reference types="cypress" />')) {
      result = '/// <reference types="cypress" />\n\n' + result;
    }
  }

  return result;
}

// ═══════════════════════════════════════════════════════════════════════
// WebdriverIO → Cypress
// ═══════════════════════════════════════════════════════════════════════

/**
 * Convert WebdriverIO commands to Cypress equivalents.
 */
function convertWdioToCypress(content) {
  let result = content;

  // --- WDIO assertions → Cypress .should() chains ---

  result = result.replace(
    /await expect\(browser\)\.toHaveUrl\(([^)]+)\)/g,
    "cy.url().should('eq', $1)"
  );
  result = result.replace(
    /await expect\(browser\)\.toHaveUrlContaining\(([^)]+)\)/g,
    "cy.url().should('include', $1)"
  );
  result = result.replace(
    /await expect\(browser\)\.toHaveTitle\(([^)]+)\)/g,
    "cy.title().should('eq', $1)"
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.toBeDisplayed\(\)/g,
    "cy.get($1).should('be.visible')"
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.not\.toBeDisplayed\(\)/g,
    "cy.get($1).should('not.be.visible')"
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.toExist\(\)/g,
    "cy.get($1).should('exist')"
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.not\.toExist\(\)/g,
    "cy.get($1).should('not.exist')"
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.toHaveText\(([^)]+)\)/g,
    "cy.get($1).should('have.text', $2)"
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.toHaveTextContaining\(([^)]+)\)/g,
    "cy.get($1).should('contain', $2)"
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.toHaveValue\(([^)]+)\)/g,
    "cy.get($1).should('have.value', $2)"
  );
  result = result.replace(
    /await expect\(\$\$\(([^)]+)\)\)\.toBeElementsArrayOfSize\(([^)]+)\)/g,
    "cy.get($1).should('have.length', $2)"
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.toBeSelected\(\)/g,
    "cy.get($1).should('be.checked')"
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.toBeEnabled\(\)/g,
    "cy.get($1).should('be.enabled')"
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.toBeDisabled\(\)/g,
    "cy.get($1).should('be.disabled')"
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.toHaveAttribute\(([^,]+),\s*([^)]+)\)/g,
    "cy.get($1).should('have.attr', $2, $3)"
  );

  // --- WDIO text selectors (before composite patterns to avoid $() catch-all) ---

  // $('=text') -> cy.contains('text')
  result = result.replace(/\$\(['"]=([\w\s]+)['"]\)/g, "cy.contains('$1')");
  // $('*=text') -> cy.contains('text')
  result = result.replace(/\$\(['"]\*=([\w\s]+)['"]\)/g, "cy.contains('$1')");

  // --- Composite $().action() chains ---

  result = result.replace(
    /await \$\(([^)]+)\)\.setValue\(([^)]+)\)/g,
    'cy.get($1).clear().type($2)'
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.click\(\)/g,
    'cy.get($1).click()'
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.doubleClick\(\)/g,
    'cy.get($1).dblclick()'
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.clearValue\(\)/g,
    'cy.get($1).clear()'
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.moveTo\(\)/g,
    "cy.get($1).trigger('mouseover')"
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.getText\(\)/g,
    "cy.get($1).invoke('text')"
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.isDisplayed\(\)/g,
    "cy.get($1).should('be.visible')"
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.isExisting\(\)/g,
    "cy.get($1).should('exist')"
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.waitForDisplayed\(\)/g,
    "cy.get($1).should('be.visible')"
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.waitForExist\(\)/g,
    "cy.get($1).should('exist')"
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.selectByVisibleText\(([^)]+)\)/g,
    'cy.get($1).select($2)'
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.selectByAttribute\(['"]value['"],\s*([^)]+)\)/g,
    'cy.get($1).select($2)'
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.getAttribute\(([^)]+)\)/g,
    "cy.get($1).invoke('attr', $2)"
  );

  // --- Standalone $() / $$() -> cy.get() ---

  result = result.replace(/\$\$\(([^)]+)\)/g, 'cy.get($1)');
  result = result.replace(/\$\(([^)]+)\)/g, 'cy.get($1)');

  // --- Navigation ---

  result = result.replace(/await browser\.url\(([^)]+)\)/g, 'cy.visit($1)');

  // --- Browser API ---

  result = result.replace(/await browser\.pause\(([^)]+)\)/g, 'cy.wait($1)');
  result = result.replace(
    /await browser\.execute\(([^)]*)\)/g,
    'cy.window().then($1)'
  );
  result = result.replace(/await browser\.refresh\(\)/g, 'cy.reload()');
  result = result.replace(/await browser\.back\(\)/g, "cy.go('back')");
  result = result.replace(/await browser\.forward\(\)/g, "cy.go('forward')");
  result = result.replace(/await browser\.getTitle\(\)/g, 'cy.title()');
  result = result.replace(/await browser\.getUrl\(\)/g, 'cy.url()');
  result = result.replace(
    /await browser\.keys\(\[([^\]]+)\]\)/g,
    "cy.get('body').type($1)"
  );

  // --- Cookies ---

  result = result.replace(
    /await browser\.deleteCookies\(\)/g,
    'cy.clearCookies()'
  );
  result = result.replace(/await browser\.getCookies\(\)/g, 'cy.getCookies()');

  // --- Unconvertible: browser.mock ---

  result = result.replace(
    /await browser\.mock\([^)]+(?:,\s*[^)]+)?\)/g,
    (match) =>
      formatter.formatTodo({
        id: 'UNCONVERTIBLE-MOCK',
        description: 'WDIO browser.mock() has no direct Cypress equivalent',
        original: match.trim(),
        action: 'Use cy.intercept() for network interception in Cypress',
      }) +
      '\n// ' +
      match.trim()
  );

  return result;
}

// ═══════════════════════════════════════════════════════════════════════
// TestCafe → Cypress
// ═══════════════════════════════════════════════════════════════════════

/**
 * Convert TestCafe commands to Cypress equivalents.
 */
function convertTestCafeToCypress(content) {
  let result = content;

  // --- TestCafe assertions (most specific first) ---

  result = result.replace(
    /await\s+t\.expect\(Selector\(([^)]+)\)\.exists\)\.ok\(\)/g,
    "cy.get($1).should('exist')"
  );
  result = result.replace(
    /await\s+t\.expect\(Selector\(([^)]+)\)\.exists\)\.notOk\(\)/g,
    "cy.get($1).should('not.exist')"
  );
  result = result.replace(
    /await\s+t\.expect\(Selector\(([^)]+)\)\.visible\)\.ok\(\)/g,
    "cy.get($1).should('be.visible')"
  );
  result = result.replace(
    /await\s+t\.expect\(Selector\(([^)]+)\)\.visible\)\.notOk\(\)/g,
    "cy.get($1).should('not.be.visible')"
  );
  result = result.replace(
    /await\s+t\.expect\(Selector\(([^)]+)\)\.count\)\.eql\(([^)]+)\)/g,
    "cy.get($1).should('have.length', $2)"
  );
  result = result.replace(
    /await\s+t\.expect\(Selector\(([^)]+)\)\.innerText\)\.eql\(([^)]+)\)/g,
    "cy.get($1).should('have.text', $2)"
  );
  result = result.replace(
    /await\s+t\.expect\(Selector\(([^)]+)\)\.innerText\)\.contains\(([^)]+)\)/g,
    "cy.get($1).should('contain', $2)"
  );
  result = result.replace(
    /await\s+t\.expect\(Selector\(([^)]+)\)\.value\)\.eql\(([^)]+)\)/g,
    "cy.get($1).should('have.value', $2)"
  );

  // Generic t.expect assertions
  result = result.replace(
    /await\s+t\.expect\(([^)]+)\)\.ok\(\)/g,
    'expect($1).to.be.ok'
  );
  result = result.replace(
    /await\s+t\.expect\(([^)]+)\)\.notOk\(\)/g,
    'expect($1).to.not.be.ok'
  );
  result = result.replace(
    /await\s+t\.expect\(([^)]+)\)\.eql\(([^)]+)\)/g,
    'expect($1).to.equal($2)'
  );
  result = result.replace(
    /await\s+t\.expect\(([^)]+)\)\.contains\(([^)]+)\)/g,
    'expect($1).to.contain($2)'
  );

  // --- t.* actions ---

  result = result.replace(
    /await\s+t\.typeText\(([^,]+),\s*([^)]+)\)/g,
    'cy.get($1).type($2)'
  );
  result = result.replace(/await\s+t\.click\(([^)]+)\)/g, 'cy.get($1).click()');
  result = result.replace(
    /await\s+t\.doubleClick\(([^)]+)\)/g,
    'cy.get($1).dblclick()'
  );
  result = result.replace(
    /await\s+t\.hover\(([^)]+)\)/g,
    "cy.get($1).trigger('mouseover')"
  );
  result = result.replace(/await\s+t\.navigateTo\(([^)]+)\)/g, 'cy.visit($1)');
  result = result.replace(/await\s+t\.wait\(([^)]+)\)/g, 'cy.wait($1)');
  result = result.replace(/await\s+t\.takeScreenshot\(\)/g, 'cy.screenshot()');
  result = result.replace(
    /await\s+t\.resizeWindow\(([^,]+),\s*([^)]+)\)/g,
    'cy.viewport($1, $2)'
  );
  result = result.replace(
    /await\s+t\.pressKey\(([^)]+)\)/g,
    "cy.get('body').type($1)"
  );

  // --- Selector chains ---

  result = result.replace(
    /Selector\(([^)]+)\)\.withText\(([^)]+)\)/g,
    'cy.contains($1, $2)'
  );
  result = result.replace(
    /Selector\(([^)]+)\)\.nth\(([^)]+)\)/g,
    'cy.get($1).eq($2)'
  );
  result = result.replace(
    /Selector\(([^)]+)\)\.find\(([^)]+)\)/g,
    'cy.get($1).find($2)'
  );

  // Standalone Selector() -> cy.get()
  result = result.replace(/Selector\(([^)]+)\)/g, 'cy.get($1)');

  // --- Unconvertible: Role, RequestMock ---

  result = result.replace(
    /const\s+\w+\s*=\s*Role\([^)]+(?:,\s*async\s+t\s*=>\s*\{[\s\S]*?\})\s*\)\s*;?/g,
    (match) =>
      formatter.formatTodo({
        id: 'UNCONVERTIBLE-ROLE',
        description: 'TestCafe Role() has no direct Cypress equivalent',
        original: match.trim(),
        action: 'Use cy.session() for auth state management in Cypress',
      }) +
      '\n// ' +
      match.trim()
  );

  result = result.replace(
    /RequestMock\(\)/g,
    (match) =>
      '/* ' +
      formatter.formatTodo({
        id: 'UNCONVERTIBLE-REQUEST-MOCK',
        description: 'TestCafe RequestMock() — use cy.intercept() in Cypress',
        original: match.trim(),
        action: 'Rewrite using cy.intercept() for network mocking',
      }) +
      ' */'
  );

  return result;
}

/**
 * Convert TestCafe test structure to Cypress describe/it.
 */
function convertTestCafeStructure(content) {
  let result = content;

  // fixture`name` -> describe('name', () => {
  result = result.replace(/fixture\s*`([^`]*)`/g, "describe('$1', () => {");

  // .page`url` -> beforeEach with cy.visit
  result = result.replace(
    /\.page\s*`([^`]*)`\s*;?/g,
    "\n  beforeEach(() => {\n    cy.visit('$1');\n  });"
  );

  // test('name', async t => { -> it('name', () => {
  result = result.replace(
    /test\(([^,]+),\s*async\s+t\s*=>\s*\{/g,
    'it($1, () => {'
  );

  return result;
}

// ═══════════════════════════════════════════════════════════════════════
// Playwright → Cypress
// ═══════════════════════════════════════════════════════════════════════

/**
 * Convert Playwright commands to Cypress equivalents.
 */
function convertPlaywrightToCypress(content) {
  let result = content;

  // --- Restore HAMLET-TODO blocks back to original cy.* calls ---

  // Catch-all 2-line format:
  // /* HAMLET-TODO: cy.xxx() has no Playwright equivalent — rewrite manually */
  // // cy.xxx(...)
  result = result.replace(
    /\/\* HAMLET-TODO: cy\.\w+\(\) has no Playwright equivalent[^*]*\*\/\s*\n\s*\/\/ (cy\.[^\n]+)/g,
    '$1'
  );

  // TodoFormatter 3-line format:
  // // HAMLET-TODO [TYPE]: description
  // // Original: cy.xxx(...)
  // // Manual action required: ...
  result = result.replace(
    /\/\/ HAMLET-TODO \[[^\]]+\]:[^\n]*\n\s*\/\/ Original: (cy\.[^\n]+)\n\s*\/\/ Manual action[^\n]*/g,
    '$1'
  );

  // --- Restore annotated custom commands ---

  // page.getByTestId("x") /* @hamlet:getBySel */ → cy.getBySel("x")
  result = result.replace(
    /page\.getByTestId\(([^)]+)\)\s*\/\* @hamlet:getBySel \*\//g,
    'cy.getBySel($1)'
  );

  // page.locator(`[data-test*=${x}]`) /* @hamlet:getBySelLike */ → cy.getBySelLike(x)
  result = result.replace(
    /page\.locator\(`\[data-test\*=\$\{([^}]+)\}\]`\)\s*\/\* @hamlet:getBySelLike \*\//g,
    'cy.getBySelLike($1)'
  );

  // await page.screenshot({ path: x }) /* @hamlet:visualSnapshot */ → cy.visualSnapshot(x)
  result = result.replace(
    /await page\.screenshot\(\{\s*path:\s*([^}]+?)\s*\}\)\s*\/\* @hamlet:visualSnapshot \*\//g,
    'cy.visualSnapshot($1)'
  );

  // route(url, ...) /* @hamlet:intercept(method).as("alias") */ → cy.intercept(method, url).as("alias")
  result = result.replace(
    /await page\.route\(([^,\n]+),\s*route\s*=>\s*route\.continue\(\)\)\s*\/\* @hamlet:intercept\(([^)]+)\)\.as\("([^"]+)"\) \*\//g,
    'cy.intercept($2, $1).as("$3")'
  );
  // route(url, ...) /* @hamlet:as("alias") */ → cy.intercept(url).as("alias")
  result = result.replace(
    /await page\.route\(([^,\n]+),\s*route\s*=>\s*route\.continue\(\)\)\s*\/\* @hamlet:as\("([^"]+)"\) \*\//g,
    'cy.intercept($1).as("$2")'
  );
  // route(url, route => route.fulfill(resp)) /* @hamlet:intercept(method).as("alias") */ → cy.intercept(method, url, resp).as("alias")
  result = result.replace(
    /await page\.route\(([^,\n]+),\s*route\s*=>\s*route\.fulfill\(([^)]+)\)\)\s*\/\* @hamlet:intercept\(([^)]+)\)\.as\("([^"]+)"\) \*\//g,
    'cy.intercept($3, $1, $2).as("$4")'
  );
  // route(url, route => route.fulfill(resp)) /* @hamlet:as("alias") */ → cy.intercept(url, resp).as("alias")
  result = result.replace(
    /await page\.route\(([^,\n]+),\s*route\s*=>\s*route\.fulfill\(([^)]+)\)\)\s*\/\* @hamlet:as\("([^"]+)"\) \*\//g,
    'cy.intercept($1, $2).as("$3")'
  );
  // Generic .as() annotation fallback
  result = result.replace(
    /\s*\/\* @hamlet:as\("([^"]+)"\) \*\//g,
    '.as("$1")'
  );

  // --- Assertions (most specific first) ---

  // toHaveURL(new RegExp(path)) → cy.location("pathname").should("eq", path)
  result = result.replace(
    /await expect\(page\)\.toHaveURL\(new RegExp\(([^)]+)\)\)/g,
    'cy.location("pathname").should("eq", $1)'
  );
  result = result.replace(
    /await expect\(page\)\.toHaveURL\(([^)]+)\)/g,
    "cy.url().should('include', $1)"
  );
  result = result.replace(
    /await expect\(page\)\.toHaveTitle\(([^)]+)\)/g,
    "cy.title().should('eq', $1)"
  );
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
    /await expect\(page\.locator\(([^)]+)\)\)\.toHaveAttribute\(([^,]+),\s*([^)]+)\)/g,
    "cy.get($1).should('have.attr', $2, $3)"
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
    /await expect\(page\.locator\(([^)]+)\)\)\.toHaveCount\(([^)]+)\)/g,
    "cy.get($1).should('have.length', $2)"
  );
  // Generic expect with locator variable
  result = result.replace(
    /await expect\(([^)]+)\)\.toBeVisible\(\)/g,
    "$1.should('be.visible')"
  );
  result = result.replace(
    /await expect\(([^)]+)\)\.toHaveText\(([^)]+)\)/g,
    "$1.should('have.text', $2)"
  );
  result = result.replace(
    /await expect\(([^)]+)\)\.toContainText\(([^)]+)\)/g,
    "$1.should('contain', $2)"
  );
  result = result.replace(
    /await expect\(([^)]+)\)\.toHaveValue\(([^)]+)\)/g,
    "$1.should('have.value', $2)"
  );

  // Chained assertions (from .and() conversion): .toContainText(x) → .and("contain", x)
  result = result.replace(
    /\.toContainText\(([^)]+)\)/g,
    '.and("contain", $1)'
  );
  result = result.replace(
    /\.toHaveText\(([^)]+)\)/g,
    '.and("have.text", $1)'
  );
  result = result.replace(
    /\.toHaveAttribute\(([^)]+)\)/g,
    '.and("have.attr", $1)'
  );
  result = result.replace(
    /\.toHaveClass\(([^)]+)\)/g,
    '.and("have.class", $1)'
  );

  // --- Wait patterns ---

  result = result.replace(
    /await page\.waitForTimeout\((\d+)\)/g,
    'cy.wait($1)'
  );
  result = result.replace(
    /await page\.waitForSelector\(([^)]+)\)/g,
    'cy.get($1)'
  );
  result = result.replace(
    /await page\.waitForURL\(([^)]+)\)/g,
    "cy.url().should('include', $1)"
  );
  result = result.replace(
    /await page\.waitForLoadState\(['"]networkidle['"]\)/g,
    'cy.wait(1000)'
  );

  // --- Navigation ---

  result = result.replace(/await page\.goto\(([^)]+)\)/g, 'cy.visit($1)');
  // Restore numeric go() from annotation before generic conversion
  result = result.replace(
    /await page\.goBack\(\)\s*\/\* @hamlet:go\((-?\d+)\) \*\//g,
    'cy.go($1)'
  );
  result = result.replace(
    /await page\.goForward\(\)\s*\/\* @hamlet:go\((\d+)\) \*\//g,
    'cy.go($1)'
  );
  result = result.replace(/await page\.goBack\(\)/g, "cy.go('back')");
  result = result.replace(/await page\.goForward\(\)/g, "cy.go('forward')");
  result = result.replace(/await page\.reload\(\)/g, 'cy.reload()');
  // new URL(page.url()).prop → cy.location('prop')
  result = result.replace(
    /new URL\(page\.url\(\)\)\.(\w+)/g,
    "cy.location('$1')"
  );
  result = result.replace(/page\.url\(\)/g, 'cy.url()');
  result = result.replace(/await page\.title\(\)/g, 'cy.title()');

  // --- Composite locator actions (before standalone locator) ---

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
    /await page\.locator\(([^)]+)\)\.focus\(\)/g,
    'cy.get($1).focus()'
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.blur\(\)/g,
    'cy.get($1).blur()'
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.hover\(\)/g,
    "cy.get($1).trigger('mouseover')"
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.scrollIntoViewIfNeeded\(\)/g,
    'cy.get($1).scrollIntoView()'
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.setInputFiles\(([^)]+)\)/g,
    'cy.get($1).selectFile($2)'
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.textContent\(\)/g,
    "cy.get($1).invoke('text')"
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.isVisible\(\)/g,
    "cy.get($1).should('be.visible')"
  );
  result = result.replace(
    /await page\.locator\(([^)]+)\)\.getAttribute\(([^)]+)\)/g,
    "cy.get($1).invoke('attr', $2)"
  );

  // --- Reverse mappings for common Playwright patterns ---

  // localStorage/sessionStorage
  result = result.replace(
    /await page\.evaluate\(\(key\) => localStorage\.removeItem\(key\),\s*([^)]+)\)\s*\/\* @hamlet:clearLocalStorage\([^)]*\) \*\//g,
    'cy.clearLocalStorage($1)'
  );
  result = result.replace(
    /await page\.evaluate\(\(\) => localStorage\.clear\(\)\)/g,
    'cy.clearLocalStorage()'
  );
  result = result.replace(
    /await page\.evaluate\(\(\) => \(\{ \.\.\.localStorage \}\)\)/g,
    'cy.getAllLocalStorage()'
  );
  result = result.replace(
    /await page\.evaluate\(\(\) => \(\{ \.\.\.sessionStorage \}\)\)/g,
    'cy.getAllSessionStorage()'
  );
  result = result.replace(
    /await page\.evaluate\(\(\) => sessionStorage\.clear\(\)\)/g,
    'cy.clearAllSessionStorage()'
  );

  // Cookies — decompose addCookies back to setCookie(name, value)
  result = result.replace(
    /await context\.addCookies\(\[\{\s*name:\s*([^,]+),\s*value:\s*([^,]+),\s*url:[^}]*\}\]\)/g,
    'cy.setCookie($1, $2)'
  );
  result = result.replace(
    /await context\.clearCookies\(\{\s*name:\s*([^}]+)\}\)/g,
    'cy.clearCookie($1)'
  );
  result = result.replace(
    /await context\.clearCookies\(\)/g,
    'cy.clearCookies()'
  );
  // context.cookies().then(cookies => cookies.find(c => c.name === x)) → cy.getCookie(x)
  result = result.replace(
    /await context\.cookies\(\)\.then\(cookies\s*=>\s*cookies\.find\(c\s*=>\s*c\.name\s*===\s*([^)]+)\)\)/g,
    'cy.getCookie($1)'
  );
  result = result.replace(/await context\.cookies\(\)/g, 'cy.getCookies()');

  // console.log → cy.log
  result = result.replace(/console\.log\(([^)]+)\)/g, 'cy.log($1)');

  // page.waitForResponse with variable → cy.wait(`@${var}`)
  result = result.replace(
    /await page\.waitForResponse\(response\s*=>\s*response\.url\(\)\.includes\((\w[^)]*)\)\)/g,
    'cy.wait(`@${$1}`)'
  );
  // page.waitForResponse with string → cy.wait('@alias')
  result = result.replace(
    /await page\.waitForResponse\(response\s*=>\s*response\.url\(\)\.includes\(["']([^"']+)["']\)\)/g,
    "cy.wait('@$1')"
  );
  // Generic waitForResponse fallback
  result = result.replace(
    /await page\.waitForResponse\([^)]+\)/g,
    "cy.wait('@alias') /* HAMLET-TODO: map response matcher to named alias */"
  );

  // page.on → cy.on
  result = result.replace(/page\.on\(([^,]+),\s*([^)]+)\)/g, 'cy.on($1, $2)');

  // page.should() → cy.window().should() (bare page used as cy.window() proxy)
  result = result.replace(/\bpage\.should\(/g, 'cy.window().should(');

  // request.method(url, { data: body }) → cy.request('METHOD', url, body)
  result = result.replace(
    /await request\.(get|post|put|patch|delete)\(([^,)]+),\s*\{\s*data:\s*([^}]+)\}\)/g,
    (match, method, url, body) =>
      `cy.request('${method.toUpperCase()}', ${url.trim()}, ${body.trim()})`
  );
  // request.method(url) /* @hamlet:explicit-method */ → cy.request('METHOD', url) (preserve explicit method)
  result = result.replace(
    /await request\.(get|post|put|patch|delete)\(([^)]+)\)\s*\/\* @hamlet:explicit-method \*\//g,
    (match, method, url) =>
      `cy.request("${method.toUpperCase()}", ${url.trim()})`
  );
  // request.get(url) → cy.request(url) (GET is default, no need to specify method)
  result = result.replace(/await request\.get\(([^)]+)\)/g, 'cy.request($1)');
  result = result.replace(
    /await request\.(post|put|patch|delete)\(([^)]+)\)/g,
    (match, method, url) =>
      `cy.request('${method.toUpperCase()}', ${url.trim()})`
  );

  // ['property'] → .its('property') — reverse of .its() conversion
  result = result.replace(/\[('[^']+')]/g, '.its($1)');
  result = result.replace(/\[("[^"]+")]/g, '.its($1)');

  // .click({ button: 'right' }) → .rightclick()
  result = result.replace(
    /\.click\(\{\s*button:\s*['"]right['"]\s*\}\)/g,
    '.rightclick()'
  );

  // .dispatchEvent('event') → .trigger('event')
  result = result.replace(/\.dispatchEvent\(([^)]+)\)/g, '.trigger($1)');

  // .scrollIntoViewIfNeeded() → .scrollIntoView()
  result = result.replace(/\.scrollIntoViewIfNeeded\(\)/g, '.scrollIntoView()');

  // .evaluate((el, prop) => el[prop], arg) → .invoke(arg)
  result = result.replace(
    /\.evaluate\(\(el,\s*prop\)\s*=>\s*el\[prop\],\s*([^)]+)\)/g,
    '.invoke($1)'
  );

  // test.step("within 'sel'", async () => { → cy.get(sel).within(() => {
  result = result.replace(
    /test\.step\("within ([^"]+)",\s*async\s*\(\)\s*=>\s*\{/g,
    'cy.get($1).within(() => {'
  );

  // .filter({ has: page.locator(sel) }) → .filter(sel)
  result = result.replace(
    /\.filter\(\{\s*has:\s*(?:page\.locator|cy\.get)\(([^)]+)\)\s*\}\)/g,
    '.filter($1)'
  );
  // .filter({ hasNot: page.locator(sel) }) → .not(sel)
  result = result.replace(
    /\.filter\(\{\s*hasNot:\s*(?:page\.locator|cy\.get)\(([^)]+)\)\s*\}\)/g,
    '.not($1)'
  );

  // --- Standalone locators ---

  result = result.replace(/page\.locator\(([^)]+)\)/g, 'cy.get($1)');
  result = result.replace(/page\.getByText\(([^)]+)\)/g, 'cy.contains($1)');
  result = result.replace(/page\.getByRole\(([^)]+)\)/g, 'cy.get(`[role=$1]`)');
  result = result.replace(
    /page\.getByTestId\(([^)]+)\)/g,
    'cy.get(`[data-testid=$1]`)'
  );

  // --- Network ---

  // page.route(url, route => route.continue()) → cy.intercept(url)
  result = result.replace(
    /await page\.route\(([^,\n]+),\s*route\s*=>\s*route\.continue\(\)\)/g,
    'cy.intercept($1)'
  );
  // page.route(url, (route) => { /* @hamlet:intercept(method) */ → cy.intercept(method, url, (req) => {
  result = result.replace(
    /await page\.route\(([^,\n]+),\s*\(route\)\s*=>\s*\{\s*\/\* @hamlet:intercept\(([^)]+)\) \*\//g,
    'cy.intercept($2, $1, (req) => {'
  );
  // page.route(url, (route) => { ... }) → cy.intercept(url, (req) => { ... })
  result = result.replace(
    /await page\.route\(([^,\n]+),\s*\(route\)\s*=>\s*\{/g,
    'cy.intercept($1, (req) => {'
  );
  // Generic page.route fallback
  result = result.replace(/await page\.route\(([^,\n]+),/g, 'cy.intercept($1,');
  result = result.replace(/await page\.screenshot\(\)/g, 'cy.screenshot()');
  // Restore named viewport presets with orientation from annotations
  result = result.replace(
    /await page\.setViewportSize\(\s*\{\s*width:\s*\d+,\s*height:\s*\d+\s*\}\s*\)\s*\/\* viewport preset: '([^']+)', '(landscape|portrait)' \*\//g,
    'cy.viewport("$1", "$2")'
  );
  // Restore named viewport presets from annotations
  result = result.replace(
    /await page\.setViewportSize\(\s*\{\s*width:\s*\d+,\s*height:\s*\d+\s*\}\s*\)\s*\/\* viewport preset: '([^']+)' \*\//g,
    "cy.viewport('$1')"
  );
  result = result.replace(
    /await page\.setViewportSize\(\s*\{\s*width:\s*(\d+),\s*height:\s*(\d+)\s*\}\s*\)/g,
    'cy.viewport($1, $2)'
  );

  // Chained .locator() → .find()
  result = result.replace(/\.locator\(([^)]+)\)/g, '.find($1)');

  // .first() / .last() / .nth()
  result = result.replace(/\.first\(\)/g, '.first()');
  result = result.replace(/\.last\(\)/g, '.last()');
  result = result.replace(/\.nth\((\d+)\)/g, '.eq($1)');

  return result;
}

/**
 * Convert Playwright test structure to Cypress structure.
 */
function convertPlaywrightStructure(content) {
  let result = content;

  result = result.replace(/test\.describe\.only\(/g, 'describe.only(');
  result = result.replace(/test\.describe\.skip\(/g, 'describe.skip(');
  // Restore context() from annotated test.describe (before generic conversion)
  result = result.replace(
    /test\.describe\(\s*\/\* @hamlet:was-context \*\//g,
    'context('
  );
  result = result.replace(/test\.describe\(/g, 'describe(');
  result = result.replace(/test\.only\(/g, 'it.only(');
  result = result.replace(/test\.skip\(/g, 'it.skip(');
  result = result.replace(/test\.beforeAll\(/g, 'before(');
  result = result.replace(/test\.afterAll\(/g, 'after(');
  result = result.replace(/test\.beforeEach\(/g, 'beforeEach(');
  result = result.replace(/test\.afterEach\(/g, 'afterEach(');
  // Match test('name', or test("name", — handle parens inside quoted strings
  result = result.replace(/\btest\(('[^']*'|"[^"]*"|`[^`]*`)\s*,/g, 'it($1,');

  // Remove { page } destructure from callback params
  result = result.replace(
    /\(\s*\{\s*page\s*(?:,\s*request\s*)?\}\s*\)\s*=>/g,
    '() =>'
  );

  return result;
}

// ═══════════════════════════════════════════════════════════════════════
// Selenium → Cypress
// ═══════════════════════════════════════════════════════════════════════

/**
 * Convert Selenium WebDriver commands to Cypress equivalents.
 */
function convertSeleniumToCypress(content) {
  let result = content;

  // --- Assertions ---

  result = result.replace(
    /expect\(await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.isDisplayed\(\)\)\.toBe\(true\)/g,
    "cy.get($1).should('be.visible')"
  );
  result = result.replace(
    /expect\(await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.isDisplayed\(\)\)\.toBe\(false\)/g,
    "cy.get($1).should('not.be.visible')"
  );
  result = result.replace(
    /expect\(await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.getText\(\)\)\.toBe\(([^)]+)\)/g,
    "cy.get($1).should('have.text', $2)"
  );
  result = result.replace(
    /expect\(await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.getText\(\)\)\.toContain\(([^)]+)\)/g,
    "cy.get($1).should('contain', $2)"
  );
  result = result.replace(
    /expect\(await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.getAttribute\('value'\)\)\.toBe\(([^)]+)\)/g,
    "cy.get($1).should('have.value', $2)"
  );
  result = result.replace(
    /expect\(await\s+driver\.getCurrentUrl\(\)\)\.toContain\(([^)]+)\)/g,
    "cy.url().should('include', $1)"
  );
  result = result.replace(
    /expect\(await\s+driver\.getCurrentUrl\(\)\)\.toBe\(([^)]+)\)/g,
    "cy.url().should('eq', $1)"
  );
  result = result.replace(
    /expect\(await\s+driver\.getTitle\(\)\)\.toBe\(([^)]+)\)/g,
    "cy.title().should('eq', $1)"
  );

  // --- Wait patterns ---

  result = result.replace(/await driver\.sleep\((\d+)\)/g, 'cy.wait($1)');
  result = result.replace(
    /await driver\.wait\(until\.elementLocated\(By\.css\(([^)]+)\)\),\s*(\d+)\)/g,
    'cy.get($1, { timeout: $2 })'
  );
  result = result.replace(
    /await driver\.wait\(until\.elementIsVisible\(([^)]+)\),\s*(\d+)\)/g,
    "$1.should('be.visible')"
  );
  result = result.replace(
    /await driver\.wait\(until\.urlContains\(([^)]+)\),\s*(\d+)\)/g,
    "cy.url().should('include', $1)"
  );

  // --- Element state assertions ---

  result = result.replace(
    /expect\(await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.isEnabled\(\)\)\.toBe\(false\)/g,
    "cy.get($1).should('be.disabled')"
  );
  result = result.replace(
    /expect\(await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.isEnabled\(\)\)\.toBe\(true\)/g,
    "cy.get($1).should('be.enabled')"
  );
  result = result.replace(
    /expect\(await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.isSelected\(\)\)\.toBe\(true\)/g,
    "cy.get($1).should('be.checked')"
  );
  result = result.replace(
    /expect\(await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.isSelected\(\)\)\.toBe\(false\)/g,
    "cy.get($1).should('not.be.checked')"
  );

  // --- Composite findElement actions ---

  result = result.replace(
    /await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.sendKeys\(([^)]+)\)/g,
    'cy.get($1).type($2)'
  );
  result = result.replace(
    /await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.click\(\)/g,
    'cy.get($1).click()'
  );
  result = result.replace(
    /await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.clear\(\)/g,
    'cy.get($1).clear()'
  );
  result = result.replace(
    /await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.getText\(\)/g,
    "cy.get($1).invoke('text')"
  );
  result = result.replace(
    /await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.isDisplayed\(\)/g,
    "cy.get($1).should('be.visible')"
  );

  // --- Navigation ---

  result = result.replace(/await driver\.get\(([^)]+)\)/g, 'cy.visit($1)');
  result = result.replace(
    /await driver\.navigate\(\)\.back\(\)/g,
    "cy.go('back')"
  );
  result = result.replace(
    /await driver\.navigate\(\)\.forward\(\)/g,
    "cy.go('forward')"
  );
  result = result.replace(
    /await driver\.navigate\(\)\.refresh\(\)/g,
    'cy.reload()'
  );
  result = result.replace(/await driver\.getCurrentUrl\(\)/g, 'cy.url()');
  result = result.replace(/await driver\.getTitle\(\)/g, 'cy.title()');

  // --- Standalone selectors ---

  result = result.replace(
    /await driver\.findElement\(By\.css\(([^)]+)\)\)/g,
    'cy.get($1)'
  );
  result = result.replace(
    /await driver\.findElement\(By\.id\(([^)]+)\)\)/g,
    'cy.get(`#${$1}`)'
  );
  result = result.replace(
    /await driver\.findElement\(By\.xpath\(([^)]+)\)\)/g,
    'cy.xpath($1)'
  );
  result = result.replace(
    /await driver\.findElements\(By\.css\(([^)]+)\)\)/g,
    'cy.get($1)'
  );
  result = result.replace(
    /driver\.findElement\(By\.css\(([^)]+)\)\)/g,
    'cy.get($1)'
  );

  // --- Interactions ---

  result = result.replace(/\.sendKeys\(([^)]+)\)/g, '.type($1)');

  return result;
}

// ═══════════════════════════════════════════════════════════════════════
// Puppeteer → Cypress
// ═══════════════════════════════════════════════════════════════════════

/**
 * Convert Puppeteer commands to Cypress equivalents.
 */
function convertPuppeteerToCypress(content) {
  let result = content;

  // --- Navigation ---

  result = result.replace(/await page\.goto\(([^)]+)\)/g, 'cy.visit($1)');
  result = result.replace(/await page\.goBack\(\)/g, "cy.go('back')");
  result = result.replace(/await page\.goForward\(\)/g, "cy.go('forward')");
  result = result.replace(/await page\.reload\(\)/g, 'cy.reload()');

  // --- Selectors and actions ---

  result = result.replace(
    /await page\.type\(([^,]+),\s*([^)]+)\)/g,
    'cy.get($1).type($2)'
  );
  result = result.replace(
    /await page\.click\(([^)]+)\)/g,
    'cy.get($1).click()'
  );
  result = result.replace(/await page\.\$\(([^)]+)\)/g, 'cy.get($1)');
  result = result.replace(/await page\.\$\$\(([^)]+)\)/g, 'cy.get($1)');
  result = result.replace(
    /await page\.waitForSelector\(([^)]+)\)/g,
    'cy.get($1)'
  );

  // --- Viewport ---

  result = result.replace(
    /await page\.setViewport\(\s*\{\s*width:\s*(\d+),\s*height:\s*(\d+)\s*\}\s*\)/g,
    'cy.viewport($1, $2)'
  );

  result = result.replace(/await page\.screenshot\(\)/g, 'cy.screenshot()');

  return result;
}

/**
 * Re-split long chains into multi-line format.
 * Forward conversion joins multi-line cy.get().should().and() chains into
 * single lines for regex matching. This restores multi-line style to
 * improve round-trip fidelity and readability.
 */
function resplitChains(content) {
  return content
    .split('\n')
    .flatMap((line) => {
      const trimmed = line.trim();
      // Only process cy.get/cy.contains chains with 2+ chained methods
      if (
        !/^(?:await\s+)?cy\.(?:get|contains)\(/.test(trimmed) &&
        !/^(?:await\s+)?(?:expect\()?cy\./.test(trimmed)
      ) {
        return [line];
      }
      // Count chain points (")." patterns)
      const chainPoints = (trimmed.match(/\)\./g) || []).length;
      // Only split if 3+ chain points, or 2+ and line is very long
      if (chainPoints < 3 && (chainPoints < 2 || trimmed.length < 120)) {
        return [line];
      }
      // Get the leading indent
      const indent = line.match(/^(\s*)/)[1];
      const contIndent = indent + '  ';
      // Split at chain points: ).method( → )\n  .method(
      const parts = [];
      let depth = 0;
      let current = '';
      for (let i = 0; i < trimmed.length; i++) {
        const ch = trimmed[i];
        if (ch === '(' || ch === '[' || ch === '{') depth++;
        if (ch === ')' || ch === ']' || ch === '}') depth--;
        current += ch;
        // Split at ")." when at depth 0 (after the closing paren)
        if (
          ch === ')' &&
          depth === 0 &&
          i + 1 < trimmed.length &&
          trimmed[i + 1] === '.'
        ) {
          parts.push(current);
          current = '';
        }
      }
      if (current) parts.push(current);
      if (parts.length < 2) return [line];
      return parts.map((p, i) => (i === 0 ? indent + p : contIndent + p));
    })
    .join('\n');
}

/**
 * Remove async/await keywords (Cypress is synchronous).
 */
function removeAsyncAwait(content) {
  let result = content;

  // Remove 'await ' keyword
  result = result.replace(/\bawait\s+/g, '');

  // Remove 'async ' from arrow functions: async () => -> () =>
  result = result.replace(/\basync\s+(\([^)]*\)\s*=>)/g, '$1');
  // Remove 'async ' from function keyword: async function -> function
  result = result.replace(/\basync\s+function\b/g, 'function');

  return result;
}

export default {
  name: 'cypress',
  language: 'javascript',
  paradigm: 'bdd-e2e',
  detect,
  parse,
  emit,
  imports: {
    globals: [
      'describe',
      'it',
      'context',
      'specify',
      'before',
      'after',
      'beforeEach',
      'afterEach',
      'cy',
      'Cypress',
      'expect',
    ],
    mockNamespace: 'cy',
  },
};
