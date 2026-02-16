/**
 * Playwright framework definition.
 *
 * Provides detect, parse, and emit for the Playwright E2E testing framework.
 * emit() transforms Cypress source code into Playwright code — the core of
 * the Cypress→Playwright conversion, ported from the legacy CypressToPlaywright.js.
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
} from '../../../core/ir.js';

function detect(source) {
  if (!source || !source.trim()) return 0;

  let score = 0;

  if (/from\s+['"]@playwright\/test['"]/.test(source)) score += 40;
  if (/\bpage\.goto\s*\(/.test(source)) score += 15;
  if (/\bpage\.locator\s*\(/.test(source)) score += 15;
  if (/\bpage\.getByText\s*\(/.test(source)) score += 10;
  if (/\btest\.describe\s*\(/.test(source)) score += 10;
  if (/\bawait expect\(/.test(source)) score += 10;
  if (/\bpage\.route\s*\(/.test(source)) score += 5;
  if (/\bpage\./.test(source)) score += 5;

  // Negative: Cypress
  if (/\bcy\./.test(source)) score -= 30;

  return Math.max(0, Math.min(100, score));
}

function parse(source) {
  // Minimal parse for when Playwright is the source (Playwright→X direction).
  return new TestFile({
    language: 'javascript',
    imports: [],
    body: [new RawCode({ code: source })],
  });
}

/**
 * Emit Playwright code from IR + original Cypress source.
 *
 * This is the Cypress→Playwright converter, ported from the legacy
 * CypressToPlaywright.js. It applies regex transforms on the source text.
 *
 * @param {TestFile} ir - Parsed IR tree (for scoring metadata)
 * @param {string} source - Original Cypress source code
 * @returns {string} Converted Playwright source code
 */
function emit(ir, source) {
  let result = source;

  // Detect test types for import generation
  const testTypes = detectTestTypes(source);

  // Phase 1: Convert Cypress commands (order matters — specific patterns first)
  result = convertCypressCommands(result);

  // Phase 2: Convert test structure
  result = convertTestStructure(result);

  // Phase 3: Transform callbacks to async with page parameter
  result = transformTestCallbacks(result, testTypes);

  // Phase 4: Add imports
  const imports = getImports(testTypes);

  // Phase 5: Clean up
  result = cleanupOutput(result);

  // Combine
  result = imports.join('\n') + '\n\n' + result;

  return result;
}

/**
 * Convert Cypress commands to Playwright equivalents.
 * Specific composite patterns first, then general patterns.
 */
function convertCypressCommands(content) {
  let result = content;

  // --- Composite cy.get().should() chains (most specific first) ---

  result = result.replace(
    /cy\.get\(([^()\n]+)\)\.should\(['"]be\.visible['"]\)/g,
    'await expect(page.locator($1)).toBeVisible()'
  );
  result = result.replace(
    /cy\.get\(([^()\n]+)\)\.should\(['"]not\.be\.visible['"]\)/g,
    'await expect(page.locator($1)).toBeHidden()'
  );
  result = result.replace(
    /cy\.get\(([^()\n]+)\)\.should\(['"]exist['"]\)/g,
    'await expect(page.locator($1)).toBeAttached()'
  );
  result = result.replace(
    /cy\.get\(([^()\n]+)\)\.should\(['"]not\.exist['"]\)/g,
    'await expect(page.locator($1)).not.toBeAttached()'
  );
  result = result.replace(
    /cy\.get\(([^()\n]+)\)\.should\(['"]have\.text['"],\s*([^()\n]+)\)/g,
    'await expect(page.locator($1)).toHaveText($2)'
  );
  result = result.replace(
    /cy\.get\(([^()\n]+)\)\.should\(['"]contain['"],\s*([^()\n]+)\)/g,
    'await expect(page.locator($1)).toContainText($2)'
  );
  result = result.replace(
    /cy\.get\(([^()\n]+)\)\.should\(['"]have\.value['"],\s*([^()\n]+)\)/g,
    'await expect(page.locator($1)).toHaveValue($2)'
  );
  result = result.replace(
    /cy\.get\(([^()\n]+)\)\.should\(['"]have\.class['"],\s*([^()\n]+)\)/g,
    'await expect(page.locator($1)).toHaveClass($2)'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.should\(['"]be\.checked['"]\)/g,
    'await expect(page.locator($1)).toBeChecked()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.should\(['"]be\.disabled['"]\)/g,
    'await expect(page.locator($1)).toBeDisabled()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.should\(['"]be\.enabled['"]\)/g,
    'await expect(page.locator($1)).toBeEnabled()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.should\(['"]have\.length['"],\s*(\d+)\)/g,
    'await expect(page.locator($1)).toHaveCount($2)'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.should\(['"]have\.attr['"],\s*([^,\n]+),\s*([^)]+)\)/g,
    'await expect(page.locator($1)).toHaveAttribute($2, $3)'
  );

  // --- Composite cy.get().action() chains ---

  result = result.replace(
    /cy\.get\(([^)]+)\)\.type\(([^)]+)\)/g,
    'await page.locator($1).fill($2)'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.click\(\)/g,
    'await page.locator($1).click()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.dblclick\(\)/g,
    'await page.locator($1).dblclick()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.check\(\)/g,
    'await page.locator($1).check()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.uncheck\(\)/g,
    'await page.locator($1).uncheck()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.select\(([^)]+)\)/g,
    'await page.locator($1).selectOption($2)'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.clear\(\)/g,
    'await page.locator($1).clear()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.focus\(\)/g,
    'await page.locator($1).focus()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.blur\(\)/g,
    'await page.locator($1).blur()'
  );

  // --- Actions with options (strip force/options object) ---

  result = result.replace(
    /cy\.get\(([^)]+)\)\.check\(\{[^{}\n]*\}\)/g,
    'await page.locator($1).check()'
  );

  // --- Traversal chains ---

  result = result.replace(
    /cy\.get\(([^)]+)\)\.first\(\)\.click\(\)/g,
    'await page.locator($1).first().click()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.last\(\)\.click\(\)/g,
    'await page.locator($1).last().click()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.eq\((\d+)\)\.click\(\)/g,
    'await page.locator($1).nth($2).click()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.first\(\)/g,
    'page.locator($1).first()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.last\(\)/g,
    'page.locator($1).last()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.eq\((\d+)\)/g,
    'page.locator($1).nth($2)'
  );

  // --- cy.contains ---

  result = result.replace(
    /cy\.contains\(([^)]+)\)\.click\(\)/g,
    'await page.getByText($1).click()'
  );
  result = result.replace(
    /cy\.contains\(([^)]+)\)/g,
    'page.getByText($1)'
  );

  // --- Navigation ---

  result = result.replace(
    /cy\.visit\(([^)]+)\)/g,
    'await page.goto($1)'
  );
  result = result.replace(
    /cy\.url\(\)\.should\(['"]include['"],\s*([^)]+)\)/g,
    'await expect(page).toHaveURL(new RegExp($1))'
  );
  result = result.replace(
    /cy\.url\(\)\.should\(['"]eq['"],\s*([^)]+)\)/g,
    'await expect(page).toHaveURL($1)'
  );
  result = result.replace(
    /cy\.title\(\)\.should\(['"]eq['"],\s*([^)]+)\)/g,
    'await expect(page).toHaveTitle($1)'
  );
  result = result.replace(
    /cy\.title\(\)\.should\(['"]include['"],\s*([^)]+)\)/g,
    'await expect(page).toHaveTitle(new RegExp($1))'
  );

  // --- Waits ---

  result = result.replace(
    /cy\.wait\(['"]@([^'"]+)['"]\)/g,
    'await page.waitForResponse(response => response.url().includes("$1"))'
  );
  result = result.replace(
    /cy\.wait\((\d+)\)/g,
    'await page.waitForTimeout($1)'
  );

  // --- Simple commands ---

  result = result.replace(/cy\.reload\(\)/g, 'await page.reload()');
  result = result.replace(/cy\.go\(['"]back['"]\)/g, 'await page.goBack()');
  result = result.replace(/cy\.go\(['"]forward['"]\)/g, 'await page.goForward()');
  result = result.replace(
    /cy\.viewport\((\d+),\s*(\d+)\)/g,
    'await page.setViewportSize({ width: $1, height: $2 })'
  );
  result = result.replace(
    /cy\.screenshot\(([^)]*)\)/g,
    'await page.screenshot({ path: $1 })'
  );
  result = result.replace(/cy\.clearCookies\(\)/g, 'await context.clearCookies()');
  result = result.replace(
    /cy\.clearLocalStorage\(\)/g,
    "await page.evaluate(() => localStorage.clear())"
  );
  result = result.replace(/cy\.log\(([^)]+)\)/g, 'console.log($1)');

  // --- Network ---

  result = result.replace(
    /cy\.intercept\(([^,\n]+),\s*([^)]+)\)\.as\(['"]([^'"]+)['"]\)/g,
    'await page.route($1, route => route.fulfill($2))'
  );

  // --- Viewport (numeric args) ---

  result = result.replace(
    /cy\.go\((-?\d+)\)/g,
    'await page.goBack() /* go($1) */'
  );
  result = result.replace(
    /cy\.reload\([^)]+\)/g,
    'await page.reload()'
  );

  // Clean up empty screenshot args
  result = result.replace(/screenshot\(\{ path: \s*\}\)/g, 'screenshot()');

  return result;
}

/**
 * Convert test structure (describe, it, hooks).
 */
function convertTestStructure(content) {
  let result = content;

  result = result.replace(/describe\.only\(/g, 'test.describe.only(');
  result = result.replace(/describe\.skip\(/g, 'test.describe.skip(');
  result = result.replace(/describe\(/g, 'test.describe(');
  result = result.replace(/context\(/g, 'test.describe(');
  result = result.replace(/it\.only\(/g, 'test.only(');
  result = result.replace(/it\.skip\(/g, 'test.skip(');
  result = result.replace(/specify\(/g, 'test(');
  result = result.replace(/it\(/g, 'test(');
  result = result.replace(/before\(/g, 'test.beforeAll(');
  result = result.replace(/after\(/g, 'test.afterAll(');
  result = result.replace(/beforeEach\(/g, 'test.beforeEach(');
  result = result.replace(/afterEach\(/g, 'test.afterEach(');

  return result;
}

/**
 * Transform test callbacks to async with { page } parameter.
 */
function transformTestCallbacks(content, testTypes) {
  const params = testTypes.includes('api') ? '{ page, request }' : '{ page }';

  // Note: Using [^,()\n]+ to prevent ReDoS
  content = content.replace(
    /test\(([^,()\n]+),\s*(?:async\s*)?\(\s*\)\s*=>\s*\{/g,
    `test($1, async (${params}) => {`
  );

  content = content.replace(
    /test\.describe\(([^,()\n]+),\s*(?:async\s*)?\(\s*\)\s*=>\s*\{/g,
    'test.describe($1, () => {'
  );

  const hookParams = '{ page }';
  content = content.replace(
    /test\.(beforeAll|afterAll|beforeEach|afterEach)\(\s*(?:async\s*)?\(\s*\)\s*=>\s*\{/g,
    `test.$1(async (${hookParams}) => {`
  );

  return content;
}

/**
 * Detect test types from Cypress source.
 */
function detectTestTypes(content) {
  const types = [];
  if (/cy\.request|cy\.intercept/.test(content)) types.push('api');
  if (/cy\.mount/.test(content)) types.push('component');
  if (/cy\.injectAxe|cy\.checkA11y/.test(content)) types.push('accessibility');
  if (/cy\.screenshot|matchImageSnapshot/.test(content)) types.push('visual');
  if (types.length === 0) types.push('e2e');
  return types;
}

/**
 * Generate Playwright import statements.
 */
function getImports(testTypes) {
  const imports = new Set([
    "import { test, expect } from '@playwright/test';"
  ]);
  if (testTypes.includes('api')) {
    imports.add("import { request } from '@playwright/test';");
  }
  if (testTypes.includes('component')) {
    imports.add("import { mount } from '@playwright/experimental-ct-react';");
  }
  if (testTypes.includes('accessibility')) {
    imports.add("import { injectAxe, checkA11y } from 'axe-playwright';");
  }
  return Array.from(imports);
}

/**
 * Clean up output.
 */
function cleanupOutput(content) {
  return content
    .replace(/await\s+await/g, 'await')
    .replace(/screenshot\(\{ path: \s*\}\)/g, 'screenshot()')
    .replace(/\n{3,}/g, '\n\n')
    .trim() + '\n';
}

export default {
  name: 'playwright',
  language: 'javascript',
  paradigm: 'bdd-e2e',
  detect,
  parse,
  emit,
  imports: {
    explicit: ['test', 'expect'],
    from: '@playwright/test',
    mockNamespace: null,
  },
};
