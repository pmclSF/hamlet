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
  const isWdioSource =
    /\bbrowser\.url\s*\(/.test(source) ||
    (/\$\(/.test(source) && /\.setValue\s*\(/.test(source));
  const isTestCafeSource =
    /\bfixture\s*`/.test(source) || /from\s+['"]testcafe['"]/.test(source);

  // Phase 1: Remove source-framework imports
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
  if (isTestCafeSource) {
    result = result.replace(
      /import\s+\{[^}]*\}\s+from\s+['"]testcafe['"];?\n?/g,
      ''
    );
  }

  // Phase 2: Convert source commands
  if (isWdioSource) {
    result = convertWdioToCypress(result);
  }
  if (isTestCafeSource) {
    result = convertTestCafeToCypress(result);
  }

  // Phase 3: Convert test structure
  if (isTestCafeSource) {
    result = convertTestCafeStructure(result);
  }
  // WDIO uses describe/it — same as Cypress, no structural changes needed

  // Phase 4: Remove async/await (Cypress is synchronous)
  result = removeAsyncAwait(result);

  // Phase 5: Clean up
  result = result.replace(/\n{3,}/g, '\n\n').trim() + '\n';

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
