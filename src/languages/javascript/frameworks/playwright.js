/**
 * Playwright framework definition.
 *
 * Provides detect, parse, and emit for the Playwright E2E testing framework.
 * emit() is the E2E hub — it handles conversions from Cypress, WebdriverIO,
 * Puppeteer, and TestCafe into Playwright code.
 */

import {
  TestFile,
  TestSuite,
  TestCase,
  Hook,
  Assertion,
  Navigation,
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

/**
 * Parse Playwright source code into a nested IR tree.
 *
 * Uses brace-depth tracking to nest test.describe, test(), and hook nodes.
 * Extracts assertions, navigation, and mock calls from body lines.
 */
function parse(source) {
  const lines = source.split('\n');
  const imports = [];

  const rootBody = [];
  const stack = [{ node: null, addChild: (c) => rootBody.push(c), depth: 0 }];
  let depth = 0;

  function addChild(child) {
    const top = stack[stack.length - 1];
    const node = top.node;

    if (!node) {
      top.addChild(child);
    } else if (node instanceof TestSuite) {
      if (child instanceof Hook) node.hooks.push(child);
      else node.tests.push(child);
    } else if (node instanceof TestCase || node instanceof Hook) {
      node.body.push(child);
    }
  }

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();
    const loc = { line: i + 1, column: 0 };

    if (!trimmed) continue;

    const opens = countBraces(trimmed, '{');
    const closes = countBraces(trimmed, '}');
    const newDepth = depth + opens - closes;

    while (stack.length > 1 && newDepth <= stack[stack.length - 1].depth) {
      stack.pop();
    }

    if (/^[}\s);,]+$/.test(trimmed)) {
      depth = newDepth;
      continue;
    }

    // Comments
    if (
      trimmed.startsWith('//') ||
      trimmed.startsWith('/*') ||
      trimmed.startsWith('*')
    ) {
      addChild(
        new Comment({ text: line, sourceLocation: loc, originalSource: line })
      );
      depth = newDepth;
      continue;
    }

    // Imports
    if (/^import\s/.test(trimmed) || /^const\s.*=\s*require\(/.test(trimmed)) {
      const sourceMatch =
        trimmed.match(/from\s+['"]([^'"]+)['"]/) ||
        trimmed.match(/require\s*\(\s*['"]([^'"]+)['"]\s*\)/);
      imports.push(
        new ImportStatement({
          source: sourceMatch ? sourceMatch[1] : '',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      depth = newDepth;
      continue;
    }

    // test.describe block
    const descMatch = trimmed.match(
      /\btest\.describe(?:\.(only|skip))?\s*\(\s*(['"`])(.+?)\2/
    );
    if (descMatch) {
      const modifiers = descMatch[1]
        ? [new Modifier({ modifierType: descMatch[1] })]
        : [];
      const suite = new TestSuite({
        name: descMatch[3],
        modifiers,
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      });
      addChild(suite);
      if (opens > closes) {
        stack.push({ node: suite, depth });
      }
      depth = newDepth;
      continue;
    }

    // test() or test.only/test.skip block
    const testMatch = trimmed.match(
      /\btest(?:\.(only|skip))?\s*\(\s*(['"`])(.+?)\2/
    );
    if (testMatch) {
      const modifiers = testMatch[1]
        ? [new Modifier({ modifierType: testMatch[1] })]
        : [];
      const tc = new TestCase({
        name: testMatch[3],
        isAsync: /async/.test(trimmed),
        modifiers,
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      });
      addChild(tc);
      if (opens > closes) {
        stack.push({ node: tc, depth });
      }
      depth = newDepth;
      continue;
    }

    // Hook: test.beforeEach, test.afterAll, etc.
    const hookMatch = trimmed.match(
      /\btest\.(beforeEach|afterEach|beforeAll|afterAll)\s*\(/
    );
    if (hookMatch) {
      const hook = new Hook({
        hookType: hookMatch[1],
        isAsync: /async/.test(trimmed),
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      });
      addChild(hook);
      if (opens > closes) {
        stack.push({ node: hook, depth });
      }
      depth = newDepth;
      continue;
    }

    // Assertions: await expect(...)
    if (/\bexpect\s*\(/.test(trimmed)) {
      const parsed = parsePlaywrightAssertion(trimmed);
      addChild(
        new Assertion({
          ...parsed,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      depth = newDepth;
      continue;
    }

    // Navigation: page.goto, page.goBack, page.goForward, page.reload
    const nav = parsePlaywrightNavigation(trimmed);
    if (nav) {
      addChild(
        new Navigation({
          ...nav,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      depth = newDepth;
      continue;
    }

    // Mock: page.route
    if (/\bpage\.route\s*\(/.test(trimmed)) {
      const routeMatch = trimmed.match(/page\.route\(\s*(['"`])([^'"]+)\1/);
      const hasFulfill = /route\.fulfill/.test(trimmed);
      addChild(
        new MockCall({
          kind: 'networkIntercept',
          target: routeMatch ? routeMatch[2] : '',
          returnValue: hasFulfill ? 'fulfill' : null,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      depth = newDepth;
      continue;
    }

    // Other page.* calls → RawCode
    addChild(
      new RawCode({ code: line, sourceLocation: loc, originalSource: line })
    );
    depth = newDepth;
  }

  return new TestFile({ language: 'javascript', imports, body: rootBody });
}

// ═══════════════════════════════════════════════════════════════════════
// Parser helpers
// ═══════════════════════════════════════════════════════════════════════

/**
 * Count occurrences of a character, skipping strings and line comments.
 */
function countBraces(line, ch) {
  let count = 0;
  let inString = false;
  let stringDelim = '';
  for (let i = 0; i < line.length; i++) {
    const c = line[i];
    if (inString) {
      if (c === stringDelim && line[i - 1] !== '\\') inString = false;
    } else {
      if (c === '/' && line[i + 1] === '/') break;
      if (c === "'" || c === '"' || c === '`') {
        inString = true;
        stringDelim = c;
      } else if (c === ch) {
        count++;
      }
    }
  }
  return count;
}

/**
 * Parse a Playwright assertion line.
 *
 * Handles:
 *   await expect(page.locator(sel)).toBeVisible()
 *   await expect(page.locator(sel)).toHaveText(val)
 *   await expect(page).toHaveURL(val)
 *   await expect(page).toHaveTitle(val)
 *   expect(val).toBe(expected)
 */
function parsePlaywrightAssertion(trimmed) {
  const isNegated = /\.not\./.test(trimmed);

  // Locator-based: expect(page.locator('sel'))
  const locatorMatch = trimmed.match(
    /expect\(page\.locator\(\s*(['"`])([^'"]+)\1\s*\)\)/
  );
  if (locatorMatch) {
    const subject = locatorMatch[2];
    const kind = extractPlaywrightMatcherKind(trimmed);
    const expected = extractPlaywrightExpected(trimmed);
    return { kind, subject, expected, isNegated };
  }

  // Page-level: expect(page).toHaveURL/toHaveTitle
  if (/expect\(page\)/.test(trimmed)) {
    if (/\.toHaveURL\(/.test(trimmed)) {
      const expected = extractPlaywrightExpected(trimmed);
      return { kind: 'url.equal', subject: 'page', expected, isNegated };
    }
    if (/\.toHaveTitle\(/.test(trimmed)) {
      const expected = extractPlaywrightExpected(trimmed);
      return { kind: 'title.equal', subject: 'page', expected, isNegated };
    }
  }

  // Value-based: expect(expr).toBe(val)
  const expectMatch = trimmed.match(/expect\((.+?)\)/);
  if (expectMatch) {
    const subject = expectMatch[1];
    const kind = extractPlaywrightMatcherKind(trimmed);
    const expected = extractPlaywrightExpected(trimmed);
    return { kind, subject, expected, isNegated };
  }

  return { kind: 'equal', subject: '', expected: null, isNegated: false };
}

/**
 * Map Playwright matcher to IR assertion kind.
 */
function extractPlaywrightMatcherKind(trimmed) {
  if (/\.toBeVisible\(/.test(trimmed)) return 'be.visible';
  if (/\.toBeAttached\(/.test(trimmed)) return 'exist';
  if (/\.toContainText\(/.test(trimmed)) return 'contain';
  if (/\.toHaveText\(/.test(trimmed)) return 'have.text';
  if (/\.toHaveCount\(/.test(trimmed)) return 'have.length';
  if (/\.toHaveAttribute\(/.test(trimmed)) return 'have.attr';
  if (/\.toHaveClass\(/.test(trimmed)) return 'have.class';
  if (/\.toHaveValue\(/.test(trimmed)) return 'have.value';
  if (/\.toBeChecked\(/.test(trimmed)) return 'be.checked';
  if (/\.toBeDisabled\(/.test(trimmed)) return 'be.disabled';
  if (/\.toBeEnabled\(/.test(trimmed)) return 'be.enabled';
  if (/\.toBeEmpty\(/.test(trimmed)) return 'be.empty';
  if (/\.toBeFocused\(/.test(trimmed)) return 'be.focused';
  if (/\.toHaveURL\(/.test(trimmed)) return 'url.equal';
  if (/\.toHaveTitle\(/.test(trimmed)) return 'title.equal';
  if (/\.toBe\(/.test(trimmed)) return 'equal';
  if (/\.toEqual\(/.test(trimmed)) return 'equal';
  if (/\.toBeTruthy\(/.test(trimmed)) return 'be.true';
  if (/\.toBeFalsy\(/.test(trimmed)) return 'be.false';
  if (/\.toBeNull\(/.test(trimmed)) return 'be.null';
  if (/\.toBeUndefined\(/.test(trimmed)) return 'be.undefined';
  return 'equal';
}

/**
 * Extract the first arg from a Playwright matcher call.
 */
function extractPlaywrightExpected(trimmed) {
  // Match the last matcher call's arg: .toSomething(value)
  const m = trimmed.match(/\.to\w+\((.+)\)\s*;?\s*$/);
  if (!m) return null;
  const inner = m[1].trim();
  return inner || null;
}

/**
 * Parse a Playwright navigation command.
 */
function parsePlaywrightNavigation(trimmed) {
  // page.goto('url')
  const gotoMatch = trimmed.match(
    /page\.goto\(\s*(['"`])([^'"]+)\1\s*(?:,\s*\{[^}]*\})?\s*\)/
  );
  if (gotoMatch) {
    return { action: 'visit', url: gotoMatch[2] };
  }
  // page.goto(variable)
  if (/page\.goto\(/.test(trimmed) && !/page\.goto\(\s*['"`]/.test(trimmed)) {
    const varMatch = trimmed.match(/page\.goto\(\s*([^,)]+)/);
    if (varMatch) {
      return { action: 'visit', url: varMatch[1].trim() };
    }
  }

  if (/page\.goBack\(/.test(trimmed)) return { action: 'goBack', url: '' };
  if (/page\.goForward\(/.test(trimmed))
    return { action: 'goForward', url: '' };
  if (/page\.reload\(/.test(trimmed)) return { action: 'reload', url: '' };

  return null;
}

/**
 * Emit Playwright code from IR + original source.
 *
 * Handles Cypress→PW, WebdriverIO→PW, Puppeteer→PW, and TestCafe→PW.
 * Each source framework's patterns are isolated in a separate function
 * and gated by source detection to prevent phase interference.
 *
 * @param {TestFile} ir - Parsed IR tree (for scoring metadata)
 * @param {string} source - Original source code
 * @returns {string} Converted Playwright source code
 */
function emit(ir, source) {
  let result = source;

  // Detect source framework
  const isCypressSource = /\bcy\./.test(source);
  const isWdioSource =
    /\bbrowser\.url\s*\(/.test(source) ||
    (/\$\(/.test(source) && /\.setValue\s*\(/.test(source));
  const isPuppeteerSource =
    /puppeteer\.launch/.test(source) ||
    /require\(['"]puppeteer['"]\)/.test(source) ||
    /from\s+['"]puppeteer['"]/.test(source);
  const isTestCafeSource =
    /\bfixture\s*`/.test(source) || /from\s+['"]testcafe['"]/.test(source);

  // Phase 1: Remove source-framework imports
  if (isCypressSource) {
    // Cypress uses globals — no imports to remove
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
      /const\s+puppeteer\s*=\s*require\(['"]puppeteer['"]\)\s*;?\n?/g,
      ''
    );
    result = result.replace(
      /import\s+puppeteer\s+from\s+['"]puppeteer['"];?\n?/g,
      ''
    );
  }
  if (isTestCafeSource) {
    result = result.replace(
      /import\s+\{[^}]*\}\s+from\s+['"]testcafe['"];?\n?/g,
      ''
    );
  }

  // Phase 2: Convert source commands (each only matches its own patterns)
  if (isCypressSource) {
    result = convertCypressCommands(result);
  }
  if (isWdioSource) {
    result = convertWdioCommands(result);
  }
  if (isPuppeteerSource) {
    result = convertPuppeteerCommands(result);
  }
  if (isTestCafeSource) {
    result = convertTestCafeCommands(result);
  }

  // Phase 3: Convert test structure
  if (isCypressSource) {
    result = convertCypressTestStructure(result);
  }
  if (isPuppeteerSource) {
    result = convertPuppeteerTestStructure(result);
  }
  if (isTestCafeSource) {
    result = convertTestCafeTestStructure(result);
  }
  // WDIO uses describe/it same as Mocha — needs same structure conversion as Cypress
  if (isWdioSource) {
    result = convertCypressTestStructure(result);
  }

  // Phase 4: Detect test types and transform callbacks
  const testTypes = detectTestTypes(source);
  result = transformTestCallbacks(result, testTypes);

  // Phase 5: Add imports
  const imports = getImports(testTypes);

  // Phase 6: Clean up
  result = cleanupOutput(result);

  // Combine
  result = imports.join('\n') + '\n\n' + result;

  return result;
}

// ═══════════════════════════════════════════════════════════════════════
// Cypress → Playwright
// ═══════════════════════════════════════════════════════════════════════

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
  result = result.replace(/cy\.contains\(([^)]+)\)/g, 'page.getByText($1)');

  // --- Navigation ---

  result = result.replace(/cy\.visit\(([^)]+)\)/g, 'await page.goto($1)');
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
  result = result.replace(
    /cy\.go\(['"]forward['"]\)/g,
    'await page.goForward()'
  );
  result = result.replace(
    /cy\.viewport\((\d+),\s*(\d+)\)/g,
    'await page.setViewportSize({ width: $1, height: $2 })'
  );
  result = result.replace(
    /cy\.screenshot\(([^)]*)\)/g,
    'await page.screenshot({ path: $1 })'
  );
  result = result.replace(
    /cy\.clearCookies\(\)/g,
    'await context.clearCookies()'
  );
  result = result.replace(
    /cy\.clearLocalStorage\(\)/g,
    'await page.evaluate(() => localStorage.clear())'
  );
  result = result.replace(/cy\.log\(([^)]+)\)/g, 'console.log($1)');

  // --- Cookies ---

  result = result.replace(
    /cy\.getCookie\(([^)]+)\)/g,
    'await context.cookies().then(cookies => cookies.find(c => c.name === $1))'
  );
  result = result.replace(/cy\.getCookies\(\)/g, 'await context.cookies()');
  result = result.replace(
    /cy\.setCookie\(([^,]+),\s*([^)]+)\)/g,
    'await context.addCookies([{ name: $1, value: $2, url: page.url() }])'
  );

  // --- Location ---

  result = result.replace(
    /cy\.location\(['"]pathname['"]\)\.should\(['"]eq['"],\s*([^)]+)\)/g,
    'await expect(page).toHaveURL(new RegExp($1))'
  );
  result = result.replace(
    /cy\.location\(['"]([^'"]+)['"]\)/g,
    'new URL(page.url()).$1'
  );
  result = result.replace(/cy\.location\(\)/g, 'new URL(page.url())');

  // --- Visual snapshot ---

  result = result.replace(
    /cy\.visualSnapshot\(([^)]*)\)/g,
    'await page.screenshot({ path: $1 })'
  );

  // --- Network ---

  // cy.intercept(method, url, response).as(alias) — static stub
  result = result.replace(
    /cy\.intercept\(([^,\n]+),\s*([^,\n]+),\s*([^)]+)\)\.as\(['"]([^'"]+)['"]\)/g,
    'await page.route($2, route => route.fulfill($3))'
  );

  // cy.intercept(method, url).as(alias) — spy
  result = result.replace(
    /cy\.intercept\(([^,\n]+),\s*([^)]+)\)\.as\(['"]([^'"]+)['"]\)/g,
    'await page.route($2, route => route.continue())'
  );

  // cy.intercept(url, callback) — callback form
  result = result.replace(
    /cy\.intercept\(([^,\n]+),\s*(?:(?:req|request)\s*=>\s*\{)/g,
    'await page.route($1, (route) => {'
  );

  // cy.intercept(url) — bare spy
  result = result.replace(
    /cy\.intercept\(([^)]+)\)\.as\(['"]([^'"]+)['"]\)/g,
    'await page.route($1, route => route.continue())'
  );

  // --- Custom Cypress commands → HAMLET-TODO ---

  // cy.getBySel(selector) → page.getByTestId(selector) (common pattern in Cypress RWA)
  result = result.replace(/cy\.getBySel\(([^)]+)\)/g, 'page.getByTestId($1)');

  // cy.getBySelLike(selector) → page.locator with data-test*= selector
  result = result.replace(
    /cy\.getBySelLike\(([^)]+)\)/g,
    'page.locator(`[data-test*=${$1}]`)'
  );

  // --- Viewport (numeric args) ---

  result = result.replace(
    /cy\.go\((-?\d+)\)/g,
    'await page.goBack() /* go($1) */'
  );
  result = result.replace(/cy\.reload\([^)]+\)/g, 'await page.reload()');

  // Clean up empty screenshot args
  result = result.replace(/screenshot\(\{ path: \s*\}\)/g, 'screenshot()');

  // --- Catch-all: remaining cy.* custom commands → HAMLET-TODO ---
  result = result.replace(/cy\.(\w+)\(([^)]*)\)/g, (match, method, args) => {
    // Skip if it's already been converted (shouldn't start with cy. anymore)
    return (
      formatter.formatTodo({
        id: 'UNCONVERTIBLE-CUSTOM-COMMAND',
        description: `Cypress custom command cy.${method}() has no Playwright equivalent`,
        original: match.trim(),
        action: 'Rewrite as a Playwright helper function or page object method',
      }) +
      '\n// ' +
      match.trim()
    );
  });

  return result;
}

// ═══════════════════════════════════════════════════════════════════════
// WebdriverIO → Playwright
// ═══════════════════════════════════════════════════════════════════════

/**
 * Convert WebdriverIO commands to Playwright equivalents.
 */
function convertWdioCommands(content) {
  let result = content;

  // --- WDIO assertions (most specific first) ---

  result = result.replace(
    /await expect\(browser\)\.toHaveUrl\(([^)]+)\)/g,
    'await expect(page).toHaveURL($1)'
  );
  result = result.replace(
    /await expect\(browser\)\.toHaveUrlContaining\(([^)]+)\)/g,
    'await expect(page).toHaveURL(new RegExp($1))'
  );
  result = result.replace(
    /await expect\(browser\)\.toHaveTitle\(([^)]+)\)/g,
    'await expect(page).toHaveTitle($1)'
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.toBeDisplayed\(\)/g,
    'await expect(page.locator($1)).toBeVisible()'
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.not\.toBeDisplayed\(\)/g,
    'await expect(page.locator($1)).toBeHidden()'
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.toExist\(\)/g,
    'await expect(page.locator($1)).toBeAttached()'
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.not\.toExist\(\)/g,
    'await expect(page.locator($1)).not.toBeAttached()'
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.toHaveText\(([^)]+)\)/g,
    'await expect(page.locator($1)).toHaveText($2)'
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.toHaveTextContaining\(([^)]+)\)/g,
    'await expect(page.locator($1)).toContainText($2)'
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.toHaveValue\(([^)]+)\)/g,
    'await expect(page.locator($1)).toHaveValue($2)'
  );
  result = result.replace(
    /await expect\(\$\$\(([^)]+)\)\)\.toBeElementsArrayOfSize\(([^)]+)\)/g,
    'await expect(page.locator($1)).toHaveCount($2)'
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.toBeSelected\(\)/g,
    'await expect(page.locator($1)).toBeChecked()'
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.toBeEnabled\(\)/g,
    'await expect(page.locator($1)).toBeEnabled()'
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.toBeDisabled\(\)/g,
    'await expect(page.locator($1)).toBeDisabled()'
  );
  result = result.replace(
    /await expect\(\$\(([^)]+)\)\)\.toHaveAttribute\(([^,]+),\s*([^)]+)\)/g,
    'await expect(page.locator($1)).toHaveAttribute($2, $3)'
  );

  // --- WDIO text selectors (before composite patterns to avoid $() catch-all) ---

  // $('=text') -> page.getByText('text')  (link text)
  result = result.replace(/\$\(['"]=([\w\s]+)['"]\)/g, "page.getByText('$1')");
  // $('*=text') -> page.getByText('text')  (partial link text)
  result = result.replace(
    /\$\(['"]\*=([\w\s]+)['"]\)/g,
    "page.getByText('$1')"
  );

  // --- Composite $().action() chains ---

  result = result.replace(
    /await \$\(([^)]+)\)\.setValue\(([^)]+)\)/g,
    'await page.locator($1).fill($2)'
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.click\(\)/g,
    'await page.locator($1).click()'
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.doubleClick\(\)/g,
    'await page.locator($1).dblclick()'
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.clearValue\(\)/g,
    'await page.locator($1).clear()'
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.moveTo\(\)/g,
    'await page.locator($1).hover()'
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.getText\(\)/g,
    'await page.locator($1).textContent()'
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.isDisplayed\(\)/g,
    'await page.locator($1).isVisible()'
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.isExisting\(\)/g,
    'await page.locator($1).isVisible()'
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.waitForDisplayed\(\)/g,
    "await page.locator($1).waitFor({ state: 'visible' })"
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.waitForExist\(\)/g,
    'await page.locator($1).waitFor()'
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.selectByVisibleText\(([^)]+)\)/g,
    'await page.locator($1).selectOption({ label: $2 })'
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.selectByAttribute\(['"]value['"],\s*([^)]+)\)/g,
    'await page.locator($1).selectOption($2)'
  );
  result = result.replace(
    /await \$\(([^)]+)\)\.getAttribute\(([^)]+)\)/g,
    'await page.locator($1).getAttribute($2)'
  );

  // --- Standalone $() / $$() -> page.locator() ---

  result = result.replace(/\$\$\(([^)]+)\)/g, 'page.locator($1)');
  result = result.replace(/\$\(([^)]+)\)/g, 'page.locator($1)');

  // --- Navigation ---

  result = result.replace(
    /await browser\.url\(([^)]+)\)/g,
    'await page.goto($1)'
  );

  // --- Browser API ---

  result = result.replace(
    /await browser\.pause\(([^)]+)\)/g,
    'await page.waitForTimeout($1)'
  );
  result = result.replace(/await browser\.execute\(/g, 'await page.evaluate(');
  result = result.replace(/await browser\.refresh\(\)/g, 'await page.reload()');
  result = result.replace(/await browser\.back\(\)/g, 'await page.goBack()');
  result = result.replace(
    /await browser\.forward\(\)/g,
    'await page.goForward()'
  );
  result = result.replace(/await browser\.getTitle\(\)/g, 'await page.title()');
  result = result.replace(/await browser\.getUrl\(\)/g, 'page.url()');
  result = result.replace(
    /await browser\.keys\(\[([^\]]+)\]\)/g,
    'await page.keyboard.press($1)'
  );

  // --- Cookies ---

  result = result.replace(
    /await browser\.setCookies\(/g,
    'await context.addCookies('
  );
  result = result.replace(
    /await browser\.getCookies\(\)/g,
    'await context.cookies()'
  );
  result = result.replace(
    /await browser\.deleteCookies\(\)/g,
    'await context.clearCookies()'
  );

  // --- Unconvertible: browser.mock ---

  result = result.replace(
    /await browser\.mock\([^)]+(?:,\s*[^)]+)?\)/g,
    (match) =>
      formatter.formatTodo({
        id: 'UNCONVERTIBLE-MOCK',
        description: 'WDIO browser.mock() has no direct Playwright equivalent',
        original: match.trim(),
        action: 'Use page.route() for network interception in Playwright',
      }) +
      '\n// ' +
      match.trim()
  );

  return result;
}

// ═══════════════════════════════════════════════════════════════════════
// Puppeteer → Playwright
// ═══════════════════════════════════════════════════════════════════════

/**
 * Convert Puppeteer commands to Playwright equivalents.
 */
function convertPuppeteerCommands(content) {
  let result = content;

  // --- Remove browser lifecycle ---

  // Remove: let browser, page; (top-level declaration)
  result = result.replace(/\s*let\s+browser\s*,\s*page\s*;?\n?/g, '\n');

  // Remove beforeAll that only does lifecycle (puppeteer.launch + newPage)
  result = result.replace(
    /\s*beforeAll\(async\s*\(\)\s*=>\s*\{\s*\n?\s*browser\s*=\s*await\s+puppeteer\.launch\([^)]*\)\s*;?\s*\n?\s*page\s*=\s*await\s+browser\.newPage\(\)\s*;?\s*\n?\s*\}\)\s*;?\n?/g,
    '\n'
  );

  // Remove afterAll that only does browser.close
  result = result.replace(
    /\s*afterAll\(async\s*\(\)\s*=>\s*\{\s*\n?\s*await\s+browser\.close\(\)\s*;?\s*\n?\s*\}\)\s*;?\n?/g,
    '\n'
  );

  // Remove standalone lifecycle lines that weren't caught by the block pattern
  result = result.replace(
    /^\s*browser\s*=\s*await\s+puppeteer\.launch\([^)]*\)\s*;?\s*$/gm,
    ''
  );
  result = result.replace(
    /^\s*page\s*=\s*await\s+browser\.newPage\(\)\s*;?\s*$/gm,
    ''
  );
  result = result.replace(/^\s*await\s+browser\.close\(\)\s*;?\s*$/gm, '');

  // --- Puppeteer assertions → Playwright assertions ---

  result = result.replace(
    /expect\(page\.url\(\)\)\.toBe\(([^)]+)\)/g,
    'await expect(page).toHaveURL($1)'
  );
  result = result.replace(
    /expect\(page\.url\(\)\)\.toContain\(([^)]+)\)/g,
    'await expect(page).toHaveURL(new RegExp($1))'
  );
  result = result.replace(
    /expect\(await\s+page\.title\(\)\)\.toBe\(([^)]+)\)/g,
    'await expect(page).toHaveTitle($1)'
  );
  result = result.replace(
    /expect\(await\s+page\.\$\(([^)]+)\)\)\.toBeTruthy\(\)/g,
    'await expect(page.locator($1)).toBeVisible()'
  );
  result = result.replace(
    /expect\(await\s+page\.\$\(([^)]+)\)\)\.toBeFalsy\(\)/g,
    'await expect(page.locator($1)).toBeHidden()'
  );
  result = result.replace(
    /expect\(await\s+page\.\$eval\(([^,]+),\s*el\s*=>\s*el\.textContent\)\)\.toBe\(([^)]+)\)/g,
    'await expect(page.locator($1)).toHaveText($2)'
  );
  result = result.replace(
    /expect\(await\s+page\.\$eval\(([^,]+),\s*el\s*=>\s*el\.textContent\)\)\.toContain\(([^)]+)\)/g,
    'await expect(page.locator($1)).toContainText($2)'
  );
  result = result.replace(
    /expect\(await\s+page\.\$eval\(([^,]+),\s*el\s*=>\s*el\.value\)\)\.toBe\(([^)]+)\)/g,
    'await expect(page.locator($1)).toHaveValue($2)'
  );
  result = result.replace(
    /expect\(\(await\s+page\.\$\$\(([^)]+)\)\)\.length\)\.toBe\(([^)]+)\)/g,
    'await expect(page.locator($1)).toHaveCount($2)'
  );

  // --- Page-level actions → locator-based ---

  result = result.replace(
    /await page\.type\(([^,]+),\s*([^)]+)\)/g,
    'await page.locator($1).fill($2)'
  );
  result = result.replace(
    /await page\.click\(([^)]+)\)/g,
    'await page.locator($1).click()'
  );
  result = result.replace(
    /await page\.hover\(([^)]+)\)/g,
    'await page.locator($1).hover()'
  );
  result = result.replace(
    /await page\.select\(([^,]+),\s*([^)]+)\)/g,
    'await page.locator($1).selectOption($2)'
  );
  result = result.replace(
    /await page\.focus\(([^)]+)\)/g,
    'await page.locator($1).focus()'
  );

  // --- Selectors ---

  result = result.replace(
    /await page\.\$eval\(([^,]+),\s*/g,
    'await page.locator($1).evaluate('
  );
  result = result.replace(
    /await page\.\$\$eval\(([^,]+),\s*/g,
    'await page.locator($1).evaluateAll('
  );
  result = result.replace(/await page\.\$\$\(([^)]+)\)/g, 'page.locator($1)');
  result = result.replace(/await page\.\$\(([^)]+)\)/g, 'page.locator($1)');

  // --- Waits ---

  result = result.replace(
    /await page\.waitForSelector\(([^)]+)\)/g,
    'await page.locator($1).waitFor()'
  );
  result = result.replace(/await page\.waitForNavigation\(\)/g, '');

  // --- Browser API ---

  result = result.replace(
    /await page\.setViewport\(\{/g,
    'await page.setViewportSize({'
  );

  // Cookie conversion
  result = result.replace(
    /await page\.setCookie\(/g,
    'await context.addCookies('
  );
  result = result.replace(
    /await page\.cookies\(\)/g,
    'await context.cookies()'
  );
  result = result.replace(
    /await page\.deleteCookie\(\)/g,
    'await context.clearCookies()'
  );

  // Standalone page.$ catch-all (after all specific patterns)
  result = result.replace(/page\.\$\(([^)]+)\)/g, 'page.locator($1)');

  return result;
}

// ═══════════════════════════════════════════════════════════════════════
// TestCafe → Playwright
// ═══════════════════════════════════════════════════════════════════════

/**
 * Convert TestCafe commands to Playwright equivalents.
 */
function convertTestCafeCommands(content) {
  let result = content;

  // --- TestCafe assertions (before action conversion) ---

  // t.expect(Selector(s).exists).ok() -> await expect(page.locator(s)).toBeAttached()
  result = result.replace(
    /await\s+t\.expect\(Selector\(([^)]+)\)\.exists\)\.ok\(\)/g,
    'await expect(page.locator($1)).toBeAttached()'
  );
  result = result.replace(
    /await\s+t\.expect\(Selector\(([^)]+)\)\.exists\)\.notOk\(\)/g,
    'await expect(page.locator($1)).not.toBeAttached()'
  );
  // t.expect(Selector(s).visible).ok() -> await expect(page.locator(s)).toBeVisible()
  result = result.replace(
    /await\s+t\.expect\(Selector\(([^)]+)\)\.visible\)\.ok\(\)/g,
    'await expect(page.locator($1)).toBeVisible()'
  );
  result = result.replace(
    /await\s+t\.expect\(Selector\(([^)]+)\)\.visible\)\.notOk\(\)/g,
    'await expect(page.locator($1)).toBeHidden()'
  );
  // t.expect(Selector(s).count).eql(n) -> await expect(page.locator(s)).toHaveCount(n)
  result = result.replace(
    /await\s+t\.expect\(Selector\(([^)]+)\)\.count\)\.eql\(([^)]+)\)/g,
    'await expect(page.locator($1)).toHaveCount($2)'
  );
  // t.expect(Selector(s).innerText).eql(text) -> await expect(page.locator(s)).toHaveText(text)
  result = result.replace(
    /await\s+t\.expect\(Selector\(([^)]+)\)\.innerText\)\.eql\(([^)]+)\)/g,
    'await expect(page.locator($1)).toHaveText($2)'
  );
  result = result.replace(
    /await\s+t\.expect\(Selector\(([^)]+)\)\.innerText\)\.contains\(([^)]+)\)/g,
    'await expect(page.locator($1)).toContainText($2)'
  );
  // t.expect(Selector(s).value).eql(val) -> await expect(page.locator(s)).toHaveValue(val)
  result = result.replace(
    /await\s+t\.expect\(Selector\(([^)]+)\)\.value\)\.eql\(([^)]+)\)/g,
    'await expect(page.locator($1)).toHaveValue($2)'
  );

  // Generic t.expect assertions
  result = result.replace(
    /await\s+t\.expect\(([^)]+)\)\.ok\(\)/g,
    'expect($1).toBeTruthy()'
  );
  result = result.replace(
    /await\s+t\.expect\(([^)]+)\)\.notOk\(\)/g,
    'expect($1).toBeFalsy()'
  );
  result = result.replace(
    /await\s+t\.expect\(([^)]+)\)\.eql\(([^)]+)\)/g,
    'expect($1).toEqual($2)'
  );
  result = result.replace(
    /await\s+t\.expect\(([^)]+)\)\.notEql\(([^)]+)\)/g,
    'expect($1).not.toEqual($2)'
  );
  result = result.replace(
    /await\s+t\.expect\(([^)]+)\)\.contains\(([^)]+)\)/g,
    'expect($1).toContain($2)'
  );
  result = result.replace(
    /await\s+t\.expect\(([^)]+)\)\.notContains\(([^)]+)\)/g,
    'expect($1).not.toContain($2)'
  );
  result = result.replace(
    /await\s+t\.expect\(([^)]+)\)\.match\(([^)]+)\)/g,
    'expect($1).toMatch($2)'
  );
  result = result.replace(
    /await\s+t\.expect\(([^)]+)\)\.gt\(([^)]+)\)/g,
    'expect($1).toBeGreaterThan($2)'
  );
  result = result.replace(
    /await\s+t\.expect\(([^)]+)\)\.lt\(([^)]+)\)/g,
    'expect($1).toBeLessThan($2)'
  );
  result = result.replace(
    /await\s+t\.expect\(([^)]+)\)\.gte\(([^)]+)\)/g,
    'expect($1).toBeGreaterThanOrEqual($2)'
  );
  result = result.replace(
    /await\s+t\.expect\(([^)]+)\)\.lte\(([^)]+)\)/g,
    'expect($1).toBeLessThanOrEqual($2)'
  );

  // --- t.* actions ---

  result = result.replace(
    /await\s+t\.typeText\(([^,]+),\s*([^)]+)\)/g,
    'await page.locator($1).fill($2)'
  );
  result = result.replace(
    /await\s+t\.click\(([^)]+)\)/g,
    'await page.locator($1).click()'
  );
  result = result.replace(
    /await\s+t\.doubleClick\(([^)]+)\)/g,
    'await page.locator($1).dblclick()'
  );
  result = result.replace(
    /await\s+t\.rightClick\(([^)]+)\)/g,
    "await page.locator($1).click({ button: 'right' })"
  );
  result = result.replace(
    /await\s+t\.hover\(([^)]+)\)/g,
    'await page.locator($1).hover()'
  );
  result = result.replace(
    /await\s+t\.pressKey\(([^)]+)\)/g,
    'await page.keyboard.press($1)'
  );
  result = result.replace(
    /await\s+t\.navigateTo\(([^)]+)\)/g,
    'await page.goto($1)'
  );
  result = result.replace(
    /await\s+t\.wait\(([^)]+)\)/g,
    'await page.waitForTimeout($1)'
  );
  result = result.replace(
    /await\s+t\.takeScreenshot\(\)/g,
    'await page.screenshot()'
  );
  result = result.replace(
    /await\s+t\.resizeWindow\(([^,]+),\s*([^)]+)\)/g,
    'await page.setViewportSize({ width: $1, height: $2 })'
  );
  result = result.replace(
    /await\s+t\.eval\(\(\)\s*=>\s*/g,
    'await page.evaluate(() => '
  );
  result = result.replace(
    /await\s+t\.setFilesToUpload\(([^,]+),\s*([^)]+)\)/g,
    'await page.locator($1).setInputFiles($2)'
  );
  result = result.replace(
    /await\s+t\.switchToIframe\(([^)]+)\)/g,
    'page.frameLocator($1)'
  );
  result = result.replace(
    /await\s+t\.switchToMainWindow\(\)/g,
    '// Back to main page'
  );

  // --- Selector chains ---

  // Selector(s).nth(n) -> page.locator(s).nth(n)
  result = result.replace(
    /Selector\(([^)]+)\)\.nth\(([^)]+)\)/g,
    'page.locator($1).nth($2)'
  );
  // Selector(s).find(child) -> page.locator(s).locator(child)
  result = result.replace(
    /Selector\(([^)]+)\)\.find\(([^)]+)\)/g,
    'page.locator($1).locator($2)'
  );
  // Selector(s).withText(text) -> page.locator(s).filter({ hasText: text })
  result = result.replace(
    /Selector\(([^)]+)\)\.withText\(([^)]+)\)/g,
    'page.locator($1).filter({ hasText: $2 })'
  );

  // Standalone Selector() -> page.locator()
  result = result.replace(/Selector\(([^)]+)\)/g, 'page.locator($1)');

  // --- Unconvertible: Role, RequestMock, ClientFunction ---

  result = result.replace(
    /const\s+\w+\s*=\s*Role\([^)]+(?:,\s*async\s+t\s*=>\s*\{[\s\S]*?\})\s*\)\s*;?/g,
    (match) =>
      formatter.formatTodo({
        id: 'UNCONVERTIBLE-ROLE',
        description: 'TestCafe Role() has no direct Playwright equivalent',
        original: match.trim(),
        action:
          'Use storageState or page.context().addCookies() for auth state in Playwright',
      }) +
      '\n// ' +
      match.trim()
  );

  result = result.replace(
    /await\s+t\.useRole\([^)]+\)/g,
    (match) =>
      formatter.formatTodo({
        id: 'UNCONVERTIBLE-USE-ROLE',
        description: 'TestCafe t.useRole() has no direct Playwright equivalent',
        original: match.trim(),
        action:
          'Use storageState or page.context().addCookies() for auth state in Playwright',
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
        description: 'TestCafe RequestMock() — use page.route() in Playwright',
        original: match.trim(),
        action: 'Rewrite using page.route() for network mocking',
      }) +
      ' */'
  );

  return result;
}

// ═══════════════════════════════════════════════════════════════════════
// Test structure converters
// ═══════════════════════════════════════════════════════════════════════

/**
 * Convert Cypress/WDIO test structure (describe/it/hooks → test.describe/test).
 */
function convertCypressTestStructure(content) {
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
 * Convert Puppeteer test structure (describe/it → test.describe/test).
 */
function convertPuppeteerTestStructure(content) {
  let result = content;

  // Same as Cypress structure conversion (Puppeteer uses Mocha/Jest runners)
  result = result.replace(/describe\.only\(/g, 'test.describe.only(');
  result = result.replace(/describe\.skip\(/g, 'test.describe.skip(');
  result = result.replace(/describe\(/g, 'test.describe(');
  result = result.replace(/it\.only\(/g, 'test.only(');
  result = result.replace(/it\.skip\(/g, 'test.skip(');
  result = result.replace(/it\(/g, 'test(');
  result = result.replace(/beforeAll\(/g, 'test.beforeAll(');
  result = result.replace(/afterAll\(/g, 'test.afterAll(');
  result = result.replace(/beforeEach\(/g, 'test.beforeEach(');
  result = result.replace(/afterEach\(/g, 'test.afterEach(');

  return result;
}

/**
 * Convert TestCafe test structure (fixture/test → test.describe/test).
 */
function convertTestCafeTestStructure(content) {
  let result = content;

  // fixture`name` -> test.describe('name', () => {
  result = result.replace(
    /fixture\s*`([^`]*)`/g,
    "test.describe('$1', () => {"
  );

  // .page`url` -> test.beforeEach with page.goto
  result = result.replace(
    /\.page\s*`([^`]*)`\s*;?/g,
    "\n  test.beforeEach(async ({ page }) => {\n    await page.goto('$1');\n  });"
  );

  // test('name', async t => { -> test('name', async ({ page }) => {
  result = result.replace(
    /test\(([^,]+),\s*async\s+t\s*=>\s*\{/g,
    'test($1, async ({ page }) => {'
  );

  return result;
}

// ═══════════════════════════════════════════════════════════════════════
// Shared helpers
// ═══════════════════════════════════════════════════════════════════════

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
 * Detect test types from source.
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
  const imports = new Set(["import { test, expect } from '@playwright/test';"]);
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
  return (
    content
      .replace(/await\s+await/g, 'await')
      .replace(/screenshot\(\{ path: \s*\}\)/g, 'screenshot()')
      .replace(/\n{3,}/g, '\n\n')
      .trim() + '\n'
  );
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
