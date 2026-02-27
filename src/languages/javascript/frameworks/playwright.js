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
  MockCall,
  ImportStatement,
  RawCode,
  Comment,
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

function parse(source) {
  // Minimal parse for when Playwright is the source (Playwright→X direction).
  return new TestFile({
    language: 'javascript',
    imports: [],
    body: [new RawCode({ code: source })],
  });
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

  // Strip incoming HAMLET-TODO blocks (from previous round-trip step)
  result = result.replace(
    /^[ \t]*\/\/ HAMLET-TODO \[[^\]]+\]:.*\n(?:[ \t]*\n)*(?:[ \t]*\/\/ (?:Original|Manual action required):.*\n(?:[ \t]*\n)*)*/gm,
    ''
  );
  result = result.replace(
    /^[ \t]*\/\*\s*HAMLET-TODO:.*?\*\/\s*\n?/gm,
    ''
  );
  // Strip Cypress reference type directive (from round-trip)
  result = result.replace(/^\/\/\/\s*<reference types="cypress"\s*\/>\s*\n?/gm, '');

  // Detect source framework
  const isCypressSource = /\bcy\./.test(source);
  const isWdioSource =
    /\bbrowser\.(url|getUrl|getTitle|pause|execute|refresh|back|forward|keys|setCookies|getCookies|deleteCookies|setWindowSize)\s*\(/.test(
      source
    ) ||
    /from\s+['"]@wdio\/globals['"]/.test(source) ||
    (/\$\(/.test(source) &&
      /\.(setValue|clearValue|moveTo|getText|isDisplayed|waitForDisplayed|selectByVisibleText|doubleClick)\s*\(/.test(
        source
      ));
  const isPuppeteerSource =
    /puppeteer\.launch/.test(source) ||
    /require\(['"]puppeteer['"]\)/.test(source) ||
    /from\s+['"]puppeteer['"]/.test(source);
  const isSeleniumSource =
    /require\s*\(\s*['"]selenium-webdriver['"]/.test(source) ||
    /from\s+['"]selenium-webdriver['"]/.test(source);
  const isTestCafeSource =
    /\bfixture\s*`/.test(source) || /from\s+['"]testcafe['"]/.test(source);

  // Phase 1: Remove source-framework imports
  if (isCypressSource) {
    // Cypress uses globals — no imports to remove
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
  if (isSeleniumSource) {
    result = convertSeleniumCommands(result);
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
  if (isSeleniumSource) {
    // Selenium uses describe/it (same structure as Cypress/Mocha)
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

  // --- Pre-process: protect test/describe name strings from cy.* conversion ---
  // Test names like it('cy.scrollTo() - ...') should not have their cy.* converted.
  const nameMap = new Map();
  let nameCounter = 0;
  result = result.replace(
    /\b(it|describe|context|specify|it\.only|it\.skip|describe\.only|describe\.skip)\(\s*(['"`])((?:(?!\2).)*)\2/g,
    (match, keyword, quote, name) => {
      const placeholder = `__HAMLET_NAME_${nameCounter++}__`;
      nameMap.set(placeholder, name);
      return `${keyword}(${quote}${placeholder}${quote}`;
    }
  );

  // --- Pre-process: join multi-line cy.get() chains into single lines ---
  // This prevents standalone cy.get(selector) from hitting the catch-all
  // when the chained action is on the next line.
  result = result.replace(
    /cy\.(get|contains|find)\(([^)]*)\)\s*\n\s*\./g,
    'cy.$1($2).'
  );
  // Join continuation lines: .action()\n  .nextAction()
  result = result.replace(
    /\)\s*\n\s*\.(?=should|and|then|click|type|check|uncheck|select|clear|focus|blur|first|last|eq|find|trigger|scrollIntoView|dblclick)/g,
    ').'
  );

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
  // contain.text (more specific — must be before 'contain')
  result = result.replace(
    /cy\.get\(([^()\n]+)\)\.should\(['"]contain\.text['"],\s*([^()\n]+)\)/g,
    'await expect(page.locator($1)).toContainText($2)'
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
  // contain.text → toContainText (alias for 'contain')
  result = result.replace(
    /cy\.get\(([^()\n]+)\)\.should\(['"]contain\.text['"],\s*([^()\n]+)\)/g,
    'await expect(page.locator($1)).toContainText($2)'
  );
  // not.be.empty → element has content
  result = result.replace(
    /cy\.get\(([^()\n]+)\)\.should\(['"]not\.be\.empty['"]\)/g,
    'await expect(page.locator($1)).not.toBeEmpty()'
  );
  // have.length.greaterThan → toHaveCount with { min }
  result = result.replace(
    /cy\.get\(([^)]+)\)\.should\(['"]have\.length\.greaterThan['"],\s*(\d+)\)/g,
    (match, sel, n) => {
      const min = parseInt(n) + 1;
      return `await expect(page.locator(${sel})).toHaveCount(${min}) /* at least ${min} */`;
    }
  );
  // have.length.at.least → toHaveCount with minimum
  result = result.replace(
    /cy\.get\(([^)]+)\)\.should\(['"]have\.length\.at\.least['"],\s*(\d+)\)/g,
    'await expect(page.locator($1)).toHaveCount($2) /* at least $2 */'
  );
  // not.have.class → negated class assertion
  result = result.replace(
    /cy\.get\(([^()\n]+)\)\.should\(['"]not\.have\.class['"],\s*([^()\n]+)\)/g,
    'await expect(page.locator($1)).not.toHaveClass($2)'
  );
  // be.empty → element is empty
  result = result.replace(
    /cy\.get\(([^()\n]+)\)\.should\(['"]be\.empty['"]\)/g,
    'await expect(page.locator($1)).toBeEmpty()'
  );
  // have.css → CSS property assertion
  result = result.replace(
    /cy\.get\(([^()\n]+)\)\.should\(['"]have\.css['"],\s*([^,\n]+),\s*([^)]+)\)/g,
    'await expect(page.locator($1)).toHaveCSS($2, $3)'
  );
  // include.text → toContainText (another alias)
  result = result.replace(
    /cy\.get\(([^()\n]+)\)\.should\(['"]include\.text['"],\s*([^()\n]+)\)/g,
    'await expect(page.locator($1)).toContainText($2)'
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
  // cy.wait(`@${expr}`) — template literal alias
  result = result.replace(
    /cy\.wait\(`@\$\{([^}]+)\}`\)/g,
    'await page.waitForResponse(response => response.url().includes($1))'
  );
  // cy.wait(["@alias1", "@alias2"]) — array of aliases
  result = result.replace(/cy\.wait\(\[([^\]]+)\]\)/g, (match, aliases) => {
    const items = aliases.split(',').map((a) =>
      a
        .trim()
        .replace(/^['"]@?/, '')
        .replace(/['"]$/, '')
    );
    return items
      .map(
        (a) =>
          `await page.waitForResponse(response => response.url().includes("${a}"))`
      )
      .join('\n');
  });
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
  // Named viewport presets (Cypress built-in preset dimensions)
  const viewportPresets = {
    'iphone-3': [320, 480],
    'iphone-4': [320, 480],
    'iphone-5': [320, 568],
    'iphone-6': [375, 667],
    'iphone-6+': [414, 736],
    'iphone-7': [375, 667],
    'iphone-8': [375, 667],
    'iphone-x': [375, 812],
    'iphone-xr': [414, 896],
    'iphone-se2': [375, 667],
    'ipad-2': [768, 1024],
    'ipad-mini': [768, 1024],
    'samsung-s10': [360, 760],
    'samsung-note9': [414, 846],
    'macbook-11': [1366, 768],
    'macbook-13': [1280, 800],
    'macbook-15': [1440, 900],
    'macbook-16': [1536, 960],
  };
  // cy.viewport('preset', 'landscape') → swap width/height
  result = result.replace(
    /cy\.viewport\(['"]([^'"]+)['"],\s*['"]landscape['"]\)/g,
    (match, preset) => {
      const dims = viewportPresets[preset];
      if (dims) {
        return `await page.setViewportSize({ width: ${dims[1]}, height: ${dims[0]} }) /* viewport preset: '${preset}', 'landscape' */`;
      }
      return `await page.setViewportSize({ width: 720, height: 1280 }) /* viewport preset: '${preset}', 'landscape' */`;
    }
  );
  // cy.viewport('preset', 'portrait') → normal order (portrait is default)
  result = result.replace(
    /cy\.viewport\(['"]([^'"]+)['"],\s*['"]portrait['"]\)/g,
    (match, preset) => {
      const dims = viewportPresets[preset];
      if (dims) {
        return `await page.setViewportSize({ width: ${dims[0]}, height: ${dims[1]} }) /* viewport preset: '${preset}', 'portrait' */`;
      }
      return `await page.setViewportSize({ width: 1280, height: 720 }) /* viewport preset: '${preset}', 'portrait' */`;
    }
  );
  // cy.viewport('preset') — default orientation
  result = result.replace(
    /cy\.viewport\(['"]([^'"]+)['"]\)/g,
    (match, preset) => {
      const dims = viewportPresets[preset];
      if (dims) {
        return `await page.setViewportSize({ width: ${dims[0]}, height: ${dims[1]} }) /* viewport preset: '${preset}' */`;
      }
      return `await page.setViewportSize({ width: 1280, height: 720 }) /* viewport preset: '${preset}' */`;
    }
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
    /cy\.clearLocalStorage\(([^)]+)\)/g,
    'await page.evaluate((key) => localStorage.removeItem(key), $1) /* @hamlet:clearLocalStorage($1) */'
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
    'await page.screenshot({ path: $1 }) /* @hamlet:visualSnapshot */'
  );

  // --- Network ---

  // cy.intercept(method, url, response).as(alias) — static stub
  result = result.replace(
    /cy\.intercept\(([^,\n]+),\s*([^,\n]+),\s*([^)]+)\)\.as\(['"]([^'"]+)['"]\)/g,
    'await page.route($2, route => route.fulfill($3)) /* @hamlet:intercept($1).as("$4") */'
  );

  // cy.intercept(method, url).as(alias) — spy
  result = result.replace(
    /cy\.intercept\(([^,\n]+),\s*([^)]+)\)\.as\(['"]([^'"]+)['"]\)/g,
    'await page.route($2, route => route.continue()) /* @hamlet:intercept($1).as("$3") */'
  );

  // cy.intercept(method, url, callback) — 3-arg callback form
  result = result.replace(
    /cy\.intercept\(([^,\n]+),\s*([^,\n]+),\s*\(?(?:req|request)\)?\s*=>\s*\{/g,
    'await page.route($2, (route) => { /* @hamlet:intercept($1) */'
  );

  // cy.intercept(url, callback) — 2-arg callback form
  result = result.replace(
    /cy\.intercept\(([^,\n]+),\s*\(?(?:req|request)\)?\s*=>\s*\{/g,
    'await page.route($1, (route) => {'
  );

  // cy.intercept(url).as(alias) — bare spy with alias
  result = result.replace(
    /cy\.intercept\(([^)]+)\)\.as\(['"]([^'"]+)['"]\)/g,
    'await page.route($1, route => route.continue()) /* @hamlet:as("$2") */'
  );

  // cy.intercept(method, url) — spy without alias (no .as())
  result = result.replace(
    /cy\.intercept\((['"][A-Z]+['"],\s*[^)]+)\)/g,
    'await page.route($1, route => route.continue())'
  );

  // cy.intercept(url) — bare spy without alias
  result = result.replace(
    /cy\.intercept\(([^)]+)\)/g,
    'await page.route($1, route => route.continue())'
  );

  // --- Custom Cypress commands → HAMLET-TODO ---

  // cy.getBySel(selector) → page.getByTestId(selector) (common pattern in Cypress RWA)
  result = result.replace(
    /cy\.getBySel\(([^)]+)\)/g,
    'page.getByTestId($1) /* @hamlet:getBySel */'
  );

  // cy.getBySelLike(selector) → page.locator with data-test*= selector
  result = result.replace(
    /cy\.getBySelLike\(([^)]+)\)/g,
    'page.locator(`[data-test*=${$1}]`) /* @hamlet:getBySelLike */'
  );

  // --- Viewport (numeric args) ---

  result = result.replace(
    /cy\.go\((-\d+)\)/g,
    'await page.goBack() /* @hamlet:go($1) */'
  );
  result = result.replace(
    /cy\.go\((\d+)\)/g,
    'await page.goForward() /* @hamlet:go($1) */'
  );
  result = result.replace(/cy\.reload\([^)]+\)/g, 'await page.reload()');

  // Clean up empty screenshot args
  result = result.replace(/screenshot\(\{ path: \s*\}\)/g, 'screenshot()');

  // --- Additional direct cy.* command patterns (before catch-all) ---

  result = result.replace(/cy\.window\(\)/g, 'page');
  result = result.replace(/cy\.document\(\)/g, 'page');
  result = result.replace(/cy\.wrap\(([^)]+)\)/g, '$1');
  result = result.replace(
    /cy\.scrollTo\(['"]([^'"]+)['"]\)/g,
    "await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight)) /* scrollTo '$1' */"
  );
  result = result.replace(
    /cy\.scrollTo\(([^,]+),\s*([^)]+)\)/g,
    'await page.evaluate(() => window.scrollTo($1, $2))'
  );
  result = result.replace(/cy\.on\(([^,]+),\s*([^)]+)\)/g, 'page.on($1, $2)');

  // cy.clock() / cy.clock(timestamp) / cy.tick() — HAMLET-TODO with brief note
  result = result.replace(
    /cy\.clock\(([^)]+)\)/g,
    formatter.formatTodo({
      id: 'CLOCK',
      description: 'Use page.clock API for clock control',
      original: 'cy.clock($1)',
      action: 'await page.clock.install({ time: $1 })',
    })
  );
  result = result.replace(
    /cy\.clock\(\)/g,
    formatter.formatTodo({
      id: 'CLOCK',
      description: 'Use page.clock API for clock control',
      original: 'cy.clock()',
      action: 'await page.clock.install()',
    })
  );
  result = result.replace(
    /cy\.tick\(([^)]+)\)/g,
    formatter.formatTodo({
      id: 'TICK',
      description: 'Use page.clock API for clock control',
      original: 'cy.tick($1)',
      action: 'await page.clock.fastForward($1)',
    })
  );

  // cy.fixture / cy.exec / cy.task / cy.readFile / cy.writeFile — HAMLET-TODO
  result = result.replace(
    /cy\.fixture\(([^)]+)\)/g,
    formatter.formatTodo({
      id: 'FIXTURE',
      description: 'No direct Playwright equivalent for cy.fixture()',
      original: 'cy.fixture($1)',
      action: 'Use fs.readFileSync() or import JSON directly',
    })
  );
  result = result.replace(
    /cy\.exec\(([^)]+)\)/g,
    formatter.formatTodo({
      id: 'EXEC',
      description: 'No direct Playwright equivalent for cy.exec()',
      original: 'cy.exec($1)',
      action: 'Use child_process.execSync() or test fixtures',
    })
  );
  result = result.replace(
    /cy\.task\(([^)]+)\)/g,
    formatter.formatTodo({
      id: 'TASK',
      description: 'No direct Playwright equivalent for cy.task()',
      original: 'cy.task($1)',
      action: 'Use a helper module or test fixture',
    })
  );
  result = result.replace(
    /cy\.readFile\(([^)]+)\)/g,
    formatter.formatTodo({
      id: 'READ-FILE',
      description: 'No direct Playwright equivalent for cy.readFile()',
      original: 'cy.readFile($1)',
      action: 'Use fs.readFileSync($1)',
    })
  );
  result = result.replace(
    /cy\.writeFile\(([^)]+)\)/g,
    formatter.formatTodo({
      id: 'WRITE-FILE',
      description: 'No direct Playwright equivalent for cy.writeFile()',
      original: 'cy.writeFile($1)',
      action: 'Use fs.writeFileSync($1)',
    })
  );
  result = result.replace(
    /cy\.stub\(([^)]*)\)/g,
    formatter.formatTodo({
      id: 'STUB',
      description: 'No direct Playwright equivalent for cy.stub()',
      original: 'cy.stub($1)',
      action: 'Use page.route() for network stubs or manual test doubles',
    })
  );
  result = result.replace(
    /cy\.spy\(([^)]*)\)/g,
    formatter.formatTodo({
      id: 'SPY',
      description: 'No direct Playwright equivalent for cy.spy()',
      original: 'cy.spy($1)',
      action: 'Use page.on() or manual instrumentation',
    })
  );

  // --- Additional cookie/storage commands ---

  result = result.replace(/cy\.getAllCookies\(\)/g, 'await context.cookies()');
  result = result.replace(
    /cy\.clearCookie\(([^)]+)\)/g,
    'await context.clearCookies({ name: $1 })'
  );
  result = result.replace(
    /cy\.clearAllCookies\(\)/g,
    'await context.clearCookies()'
  );
  result = result.replace(
    /cy\.getAllLocalStorage\(\)/g,
    'await page.evaluate(() => ({ ...localStorage }))'
  );
  result = result.replace(
    /cy\.getAllSessionStorage\(\)/g,
    'await page.evaluate(() => ({ ...sessionStorage }))'
  );
  result = result.replace(
    /cy\.clearAllLocalStorage\(\)/g,
    'await page.evaluate(() => localStorage.clear())'
  );
  result = result.replace(
    /cy\.clearAllSessionStorage\(\)/g,
    'await page.evaluate(() => sessionStorage.clear())'
  );

  // --- DOM query commands ---

  result = result.replace(/cy\.focused\(\)/g, "page.locator(':focus')");
  result = result.replace(/cy\.root\(\)/g, "page.locator(':root')");
  result = result.replace(/cy\.hash\(\)/g, 'new URL(page.url()).hash');
  result = result.replace(/cy\.url\(\)/g, 'page.url()');
  result = result.replace(/cy\.title\(\)/g, 'await page.title()');

  // --- API requests ---

  // cy.request('METHOD', url, body) → await request.method(url, { data: body })
  result = result.replace(
    /cy\.request\((['"])(GET|POST|PUT|PATCH|DELETE)\1,\s*([^,)]+),\s*([^)]+)\)/g,
    (match, _q, method, url, body) =>
      `await request.${method.toLowerCase()}(${url.trim()}, { data: ${body.trim()} })`
  );
  // cy.request('METHOD', url) → await request.method(url) /* @hamlet:explicit-method */
  result = result.replace(
    /cy\.request\((['"])(GET|POST|PUT|PATCH|DELETE)\1,\s*([^)]+)\)/g,
    (match, _q, method, url) =>
      `await request.${method.toLowerCase()}(${url.trim()}) /* @hamlet:explicit-method */`
  );
  // cy.request(url) → await request.get(url)
  result = result.replace(/cy\.request\(([^{),]+)\)/g, 'await request.get($1)');
  // cy.request({ ... }) → HAMLET-TODO (complex config object)
  result = result.replace(
    /cy\.request\(\{/g,
    formatter.formatTodo({
      id: 'REQUEST',
      description:
        'Convert cy.request() config object to Playwright request API',
      original: 'cy.request({...})',
      action: 'Use request.get/post/put/delete() with appropriate options',
    }) + '\nawait request.get({'
  );

  // --- Chaining helpers ---

  // cy.then(() => { ... }) → just inline the block (Playwright is sequential)
  result = result.replace(/cy\.then\(\(\)\s*=>\s*\{/g, '{');

  // .its('property') → Playwright doesn't chain like Cypress
  // .then(callback) → handled naturally by await
  result = result.replace(/\.its\(([^)]+)\)/g, '[$1]');

  // --- Actions not yet covered ---

  result = result.replace(
    /cy\.get\(([^)]+)\)\.submit\(\)/g,
    'await page.locator($1).locator(\'[type="submit"]\').click()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.rightclick\(\)/g,
    "await page.locator($1).click({ button: 'right' })"
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.trigger\(([^)]+)\)/g,
    'await page.locator($1).dispatchEvent($2)'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.scrollIntoView\(\)/g,
    'await page.locator($1).scrollIntoViewIfNeeded()'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.invoke\(([^)]+)\)/g,
    'await page.locator($1).evaluate((el, prop) => el[prop], $2)'
  );

  // --- Traversal methods on cy.get chains ---

  result = result.replace(
    /cy\.get\(([^)]+)\)\.find\(([^)]+)\)/g,
    'page.locator($1).locator($2)'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.children\(([^)]+)\)/g,
    'page.locator($1).locator($2)'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.parent\(\)/g,
    'page.locator($1).locator(..)'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.not\(([^)]+)\)/g,
    'page.locator($1).filter({ hasNot: page.locator($2) })'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.filter\(([^)]+)\)/g,
    'page.locator($1).filter({ has: page.locator($2) })'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.closest\(([^)]+)\)/g,
    'page.locator($2).filter({ has: page.locator($1) })'
  );
  result = result.replace(
    /cy\.get\(([^)]+)\)\.within\(\s*(?:(?:\(\)\s*=>)|(?:function\s*\(\)))\s*\{/g,
    'await test.step("within $1", async () => {'
  );

  // --- Chaining: .and() is alias for .should() in Cypress ---
  // .and('be.visible') → already handled by .should() patterns after chain join
  // Remaining .and() calls → pass through as additional assertion
  result = result.replace(
    /\.and\(['"]have\.text['"],\s*([^)]+)\)/g,
    '.toHaveText($1)'
  );
  result = result.replace(
    /\.and\(['"]contain['"],\s*([^)]+)\)/g,
    '.toContainText($1)'
  );
  result = result.replace(
    /\.and\(['"]have\.attr['"],\s*([^)]+)\)/g,
    '.toHaveAttribute($1)'
  );
  result = result.replace(
    /\.and\(['"]have\.class['"],\s*([^)]+)\)/g,
    '.toHaveClass($1)'
  );
  result = result.replace(
    /\.and\((['"]include['"],\s*[^)]+)\)/g,
    ' /* .and($1) */'
  );
  result = result.replace(/\.and\((['"][^'"]+['"])\)/g, ' /* .and($1) */');

  // --- Chaining: .then() → use await/variable binding ---
  result = result.replace(/\.then\(\((\$?\w+)\)\s*=>\s*\{/g, '.then(($1) => {');

  // Standalone cy.get(selector) — no chained action on same line
  // Convert to page.locator() instead of hitting the catch-all
  result = result.replace(/cy\.get\(([^)]+)\)/g, 'page.locator($1)');

  // --- Catch-all: remaining cy.* custom commands → HAMLET-TODO ---
  // Process line-by-line to skip comment lines
  result = result
    .split('\n')
    .map((line) => {
      const trimmed = line.trim();
      // Skip comment lines
      if (
        trimmed.startsWith('//') ||
        trimmed.startsWith('/*') ||
        trimmed.startsWith('*')
      ) {
        return line;
      }
      return line.replace(/cy\.(\w+)\(([^)]*)\)/g, (match, method) => {
        return (
          `/* HAMLET-TODO: cy.${method}() has no Playwright equivalent — rewrite manually */` +
          '\n// ' +
          match.trim()
        );
      });
    })
    .join('\n');

  // --- Post-process: restore protected test/describe name strings ---
  for (const [placeholder, name] of nameMap) {
    result = result.split(placeholder).join(name);
  }

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
// Selenium → Playwright
// ═══════════════════════════════════════════════════════════════════════

/**
 * Convert Selenium WebDriver commands to Playwright equivalents.
 */
function convertSeleniumCommands(content) {
  let result = content;

  // --- Assertions ---

  result = result.replace(
    /expect\(await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.isDisplayed\(\)\)\.toBe\(true\)/g,
    'await expect(page.locator($1)).toBeVisible()'
  );
  result = result.replace(
    /expect\(await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.isDisplayed\(\)\)\.toBe\(false\)/g,
    'await expect(page.locator($1)).toBeHidden()'
  );
  result = result.replace(
    /expect\(await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.getText\(\)\)\.toBe\(([^)]+)\)/g,
    'await expect(page.locator($1)).toHaveText($2)'
  );
  result = result.replace(
    /expect\(await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.getText\(\)\)\.toContain\(([^)]+)\)/g,
    'await expect(page.locator($1)).toContainText($2)'
  );
  result = result.replace(
    /expect\(await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.getAttribute\('value'\)\)\.toBe\(([^)]+)\)/g,
    'await expect(page.locator($1)).toHaveValue($2)'
  );
  result = result.replace(
    /expect\(await\s+driver\.getCurrentUrl\(\)\)\.toContain\(([^)]+)\)/g,
    'await expect(page).toHaveURL(new RegExp($1))'
  );
  result = result.replace(
    /expect\(await\s+driver\.getCurrentUrl\(\)\)\.toBe\(([^)]+)\)/g,
    'await expect(page).toHaveURL($1)'
  );
  result = result.replace(
    /expect\(await\s+driver\.getTitle\(\)\)\.toBe\(([^)]+)\)/g,
    'await expect(page).toHaveTitle($1)'
  );

  // --- Wait patterns ---

  result = result.replace(
    /await driver\.sleep\((\d+)\)/g,
    'await page.waitForTimeout($1)'
  );
  result = result.replace(
    /await driver\.wait\(until\.elementLocated\(By\.css\(([^)]+)\)\),\s*(\d+)\)/g,
    'await page.locator($1).waitFor({ timeout: $2 })'
  );
  result = result.replace(
    /await driver\.wait\(until\.elementIsVisible\(([^)]+)\),\s*(\d+)\)/g,
    'await expect($1).toBeVisible()'
  );
  result = result.replace(
    /await driver\.wait\(until\.urlContains\(([^)]+)\),\s*(\d+)\)/g,
    'await expect(page).toHaveURL(new RegExp($1))'
  );

  // --- Composite findElement actions ---

  result = result.replace(
    /await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.sendKeys\(([^)]+)\)/g,
    'await page.locator($1).fill($2)'
  );
  result = result.replace(
    /await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.click\(\)/g,
    'await page.locator($1).click()'
  );
  result = result.replace(
    /await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.clear\(\)/g,
    'await page.locator($1).clear()'
  );
  result = result.replace(
    /await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.getText\(\)/g,
    'await page.locator($1).textContent()'
  );
  result = result.replace(
    /await\s+\(await\s+driver\.findElement\(By\.css\(([^)]+)\)\)\)\.isDisplayed\(\)/g,
    'await page.locator($1).isVisible()'
  );

  // --- Navigation ---

  result = result.replace(
    /await driver\.get\(([^)]+)\)/g,
    'await page.goto($1)'
  );
  result = result.replace(
    /await driver\.navigate\(\)\.back\(\)/g,
    'await page.goBack()'
  );
  result = result.replace(
    /await driver\.navigate\(\)\.forward\(\)/g,
    'await page.goForward()'
  );
  result = result.replace(
    /await driver\.navigate\(\)\.refresh\(\)/g,
    'await page.reload()'
  );
  result = result.replace(/await driver\.getCurrentUrl\(\)/g, 'page.url()');
  result = result.replace(/await driver\.getTitle\(\)/g, 'await page.title()');

  // --- Standalone selectors ---

  result = result.replace(
    /await driver\.findElement\(By\.css\(([^)]+)\)\)/g,
    'page.locator($1)'
  );
  result = result.replace(
    /await driver\.findElement\(By\.id\(([^)]+)\)\)/g,
    'page.locator(`#${$1}`)'
  );
  result = result.replace(
    /await driver\.findElement\(By\.xpath\(([^)]+)\)\)/g,
    'page.locator(`xpath=${$1}`)'
  );
  result = result.replace(
    /await driver\.findElements\(By\.css\(([^)]+)\)\)/g,
    'page.locator($1)'
  );
  result = result.replace(
    /driver\.findElement\(By\.css\(([^)]+)\)\)/g,
    'page.locator($1)'
  );
  result = result.replace(
    /driver\.findElement\(By\.id\(([^)]+)\)\)/g,
    'page.locator(`#${$1}`)'
  );

  // --- Interactions ---

  result = result.replace(/\.sendKeys\(([^)]+)\)/g, '.fill($1)');

  // --- Checkbox pattern: isSelected()/click() → check() ---
  result = result.replace(
    /if\s*\(\s*!\s*\(\s*await\s+(\w+)\.isSelected\(\)\s*\)\s*\)\s*await\s+\1\.click\(\)\s*;?/g,
    'await $1.check();'
  );
  result = result.replace(/\.isSelected\(\)/g, '.isChecked()');

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

  result = result.replace(/\bdescribe\.only\(/g, 'test.describe.only(');
  result = result.replace(/\bdescribe\.skip\(/g, 'test.describe.skip(');
  result = result.replace(/\bdescribe\(/g, 'test.describe(');
  result = result.replace(
    /\bcontext\(/g,
    'test.describe( /* @hamlet:was-context */'
  );
  result = result.replace(/\bit\.only\(/g, 'test.only(');
  result = result.replace(/\bit\.skip\(/g, 'test.skip(');
  result = result.replace(/\bspecify\(/g, 'test(');
  result = result.replace(/\bit\(/g, 'test(');
  result = result.replace(/\bbefore\(/g, 'test.beforeAll(');
  result = result.replace(/\bafter\(/g, 'test.afterAll(');
  result = result.replace(/\bbeforeEach\(/g, 'test.beforeEach(');
  result = result.replace(/\bafterEach\(/g, 'test.afterEach(');

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
