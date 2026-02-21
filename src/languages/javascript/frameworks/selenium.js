/**
 * Selenium WebDriver framework definition.
 *
 * Provides detect, parse, and emit for the Selenium testing framework.
 * parse() builds an IR tree from Selenium source code for scoring.
 * emit() is a stub — Selenium conversions currently use legacy converters.
 */

import {
  TestFile,
  TestSuite,
  TestCase,
  Hook,
  Assertion,
  Navigation,
  ImportStatement,
  RawCode,
  Comment,
  Modifier,
} from '../../../core/ir.js';

/**
 * Detect whether source code is Selenium WebDriver.
 * Returns confidence score 0-100.
 */
function detect(source) {
  if (!source || !source.trim()) return 0;

  let score = 0;

  // Strong Selenium signals
  if (/require\s*\(\s*['"]selenium-webdriver['"]\s*\)/.test(source))
    score += 30;
  if (/from\s+['"]selenium-webdriver['"]/.test(source)) score += 30;
  if (/\bWebDriver\b/.test(source)) score += 15;
  if (/\bnew\s+Builder\b/.test(source)) score += 20;
  if (/\bBy\.\w+\s*\(/.test(source)) score += 15;

  // Selenium-specific patterns
  if (/driver\.get\s*\(/.test(source)) score += 15;
  if (/driver\.findElement\s*\(/.test(source)) score += 15;
  if (/driver\.findElements\s*\(/.test(source)) score += 10;
  if (/driver\.wait\s*\(/.test(source)) score += 10;
  if (/driver\.quit\s*\(/.test(source)) score += 10;
  if (/\.sendKeys\s*\(/.test(source)) score += 10;
  if (/until\.elementLocated\s*\(/.test(source)) score += 10;
  if (/driver\.getCurrentUrl\s*\(/.test(source)) score += 10;
  if (/driver\.getTitle\s*\(/.test(source)) score += 10;

  // Weak signals (shared with other frameworks)
  if (/describe\s*\(/.test(source)) score += 3;
  if (/\bit\s*\(/.test(source)) score += 3;
  if (/expect\s*\(/.test(source)) score += 3;

  // Negative signals: NOT Selenium
  if (/cy\./.test(source)) score -= 30;
  if (/page\./.test(source)) score -= 25;
  if (/browser\./.test(source)) score -= 20;
  if (/@playwright\/test/.test(source)) score -= 30;

  return Math.max(0, Math.min(100, score));
}

/**
 * Parse Selenium source code into a nested IR tree.
 *
 * Minimal parser that produces an IR skeleton:
 * - describe/it blocks → TestSuite/TestCase
 * - beforeEach/afterEach → Hook
 * - driver.get() → Navigation
 * - expect() → Assertion (kind=truthy, no extraction)
 * - Everything else → RawCode
 */
function parse(source) {
  const lines = source.split('\n');
  const imports = [];
  const rootBody = [];
  const stack = [{ node: null, addChild: (c) => rootBody.push(c), indent: -1 }];

  function addChild(child) {
    const top = stack[stack.length - 1];
    const node = top.node;
    if (!node) {
      top.addChild(child);
    } else if (node instanceof TestSuite) {
      if (child instanceof Hook) node.hooks.push(child);
      else if (child instanceof TestCase) node.tests.push(child);
      else node.tests.push(child);
    } else if (node instanceof TestCase || node instanceof Hook) {
      node.body.push(child);
    }
  }

  function getIndent(line) {
    const m = line.match(/^(\s*)/);
    return m ? m[1].length : 0;
  }

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();
    const loc = { line: i + 1, column: 0 };
    const indent = getIndent(line);

    if (!trimmed) continue;

    // Pop stack when we return to a shallower indent
    while (stack.length > 1 && indent <= stack[stack.length - 1].indent) {
      stack.pop();
    }

    // Comments
    if (trimmed.startsWith('//') || trimmed.startsWith('/*')) {
      addChild(
        new Comment({
          text: line,
          commentKind: 'inline',
          sourceLocation: loc,
          originalSource: line,
        })
      );
      continue;
    }

    // Import/require
    if (
      /^(?:import|const|let|var)\s+/.test(trimmed) &&
      (/require\s*\(/.test(trimmed) || /from\s+['"]/.test(trimmed))
    ) {
      const sourceMatch = trimmed.match(
        /(?:require\s*\(\s*['"]([^'"]+)['"]|from\s+['"]([^'"]+)['"])/
      );
      imports.push(
        new ImportStatement({
          kind: 'library',
          source: sourceMatch ? sourceMatch[1] || sourceMatch[2] : '',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // describe block → TestSuite
    if (/\bdescribe\s*\(/.test(trimmed)) {
      const nameMatch = trimmed.match(/describe\s*\(\s*(['"`])([^'"`]*)\1/);
      const suite = new TestSuite({
        name: nameMatch ? nameMatch[2] : '',
        modifiers: [],
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      });
      addChild(suite);
      stack.push({
        node: suite,
        addChild: (c) => {
          if (c instanceof Hook) suite.hooks.push(c);
          else suite.tests.push(c);
        },
        indent,
      });
      continue;
    }

    // it/test block → TestCase
    if (/\b(?:it|test)\s*\(/.test(trimmed)) {
      const nameMatch = trimmed.match(/(?:it|test)\s*\(\s*(['"`])([^'"`]*)\1/);
      const tc = new TestCase({
        name: nameMatch ? nameMatch[2] : '',
        isAsync: /async/.test(trimmed),
        modifiers: [],
        body: [],
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      });
      addChild(tc);
      stack.push({ node: tc, addChild: (c) => tc.body.push(c), indent });
      continue;
    }

    // beforeEach/afterEach/beforeAll/afterAll → Hook
    if (/\b(beforeEach|afterEach|beforeAll|afterAll)\s*\(/.test(trimmed)) {
      const hookType = trimmed.match(
        /\b(beforeEach|afterEach|beforeAll|afterAll)/
      )[1];
      const hook = new Hook({
        hookType,
        body: [],
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      });
      addChild(hook);
      stack.push({ node: hook, addChild: (c) => hook.body.push(c), indent });
      continue;
    }

    // driver.get() → Navigation
    if (/driver\.get\s*\(/.test(trimmed)) {
      const urlMatch = trimmed.match(/driver\.get\s*\(\s*(['"`])([^'"`]*)\1/);
      addChild(
        new Navigation({
          action: 'goto',
          url: urlMatch ? urlMatch[2] : undefined,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // expect() → Assertion (minimal — no subject/expected extraction)
    if (/\bexpect\s*\(/.test(trimmed)) {
      addChild(
        new Assertion({
          kind: 'truthy',
          subject: undefined,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // Everything else → RawCode
    addChild(
      new RawCode({
        code: line,
        sourceLocation: loc,
        originalSource: line,
      })
    );
  }

  return new TestFile({
    language: 'javascript',
    imports,
    body: rootBody,
  });
}

/**
 * Emit Selenium code from IR + original source.
 *
 * Stub — Selenium conversions currently use legacy converters
 * (CypressToSelenium, PlaywrightToSelenium, etc.).
 */
function emit(_ir, source) {
  return source;
}

export default {
  name: 'selenium',
  language: 'javascript',
  paradigm: 'bdd',
  detect,
  parse,
  emit,
  imports: {
    packages: ['selenium-webdriver'],
  },
};
