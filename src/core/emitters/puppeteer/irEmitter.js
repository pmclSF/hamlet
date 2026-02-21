/**
 * IR-driven Puppeteer emitter.
 *
 * Generates Puppeteer code from IR nodes. Puppeteer uses:
 * - describe/it (Jest test runner)
 * - page.$(sel), page.goto(url) for page interaction
 * - Manual assertions with expect() (Jest)
 * - browser.launch/newPage/close boilerplate
 */

import {
  Assertion,
  Navigation,
  TestFile,
  TestSuite,
  TestCase,
  Hook,
  RawCode,
  Comment,
  ImportStatement,
  MockCall,
  SharedVariable,
  Modifier,
} from '../../ir.js';

// ═══════════════════════════════════════════════════════════════════════
// Unified interface
// ═══════════════════════════════════════════════════════════════════════

export function emitNode(node) {
  if (node instanceof Assertion) return emitAssertion(node);
  if (node instanceof Navigation) return emitNavigation(node);
  return { code: '', supported: false };
}

export function matchesBaseline(line, node) {
  if (node instanceof Assertion) {
    return /\bexpect\s*\(/.test(line);
  }
  if (node instanceof Navigation) {
    return /page\.goto\s*\(/.test(line);
  }
  return false;
}

/**
 * Generate a complete Puppeteer file from an IR tree.
 */
export function emitFullFile(ir) {
  if (!(ir instanceof TestFile)) return null;

  const lines = [];

  // Puppeteer import
  lines.push("const puppeteer = require('puppeteer');");

  // Emit user imports — skip source framework imports
  for (const imp of ir.imports || []) {
    if (imp instanceof ImportStatement) {
      const src = imp.source || '';
      if (/@playwright|@wdio|webdriverio|testcafe|cypress/.test(src)) continue;
      if (/puppeteer/.test(src)) continue; // Already added
      lines.push(imp.originalSource || src);
    }
  }

  lines.push('');

  for (const node of ir.body || []) {
    emitBodyNode(node, 0, lines);
  }

  return lines.join('\n') + '\n';
}

// ═══════════════════════════════════════════════════════════════════════
// Tree-walk emitters
// ═══════════════════════════════════════════════════════════════════════

function emitBodyNode(node, depth, lines) {
  const indent = '  '.repeat(depth);

  if (node instanceof TestSuite) {
    emitSuite(node, depth, lines);
  } else if (node instanceof TestCase) {
    emitTestCase(node, depth, lines);
  } else if (node instanceof Hook) {
    emitHook(node, depth, lines);
  } else if (node instanceof Assertion) {
    const result = emitAssertion(node);
    lines.push(indent + result.code + ';');
  } else if (node instanceof Navigation) {
    const result = emitNavigation(node);
    lines.push(indent + result.code + ';');
  } else if (node instanceof Comment) {
    lines.push(indent + (node.text || '').trim());
  } else if (node instanceof RawCode) {
    const code = (node.code || '').trim();
    if (code) lines.push(indent + code);
  } else if (node instanceof MockCall) {
    const original = node.originalSource ? node.originalSource.trim() : '';
    lines.push(
      indent +
        `// HAMLET-TODO: Mock/stub conversion not supported for Puppeteer` +
        (original ? `\n${indent}// Original: ${original}` : '')
    );
  } else if (node instanceof SharedVariable) {
    lines.push(
      indent +
        `// HAMLET-TODO: Shared variable "${node.name}" conversion not yet supported`
    );
  } else {
    const original = node.originalSource
      ? node.originalSource.trim()
      : node.type || 'unknown';
    lines.push(indent + `// HAMLET-TODO: Unsupported node (${original})`);
  }
}

function emitSuite(suite, depth, lines) {
  const indent = '  '.repeat(depth);
  const name = escapeSingleQuotes(suite.name || 'unnamed');
  const mod = suiteModifier(suite);

  lines.push(`${indent}describe${mod}('${name}', () => {`);

  // Add browser/page boilerplate at suite level
  lines.push(`${indent}  let browser, page;`);
  lines.push('');
  lines.push(`${indent}  beforeAll(async () => {`);
  lines.push(`${indent}    browser = await puppeteer.launch();`);
  lines.push(`${indent}    page = await browser.newPage();`);
  lines.push(`${indent}  });`);
  lines.push('');
  lines.push(`${indent}  afterAll(async () => {`);
  lines.push(`${indent}    await browser.close();`);
  lines.push(`${indent}  });`);
  lines.push('');

  // Emit user hooks (ordered)
  const hookOrder = ['beforeAll', 'beforeEach', 'afterEach', 'afterAll'];
  const sortedHooks = [...(suite.hooks || [])].sort((a, b) => {
    return hookOrder.indexOf(a.hookType) - hookOrder.indexOf(b.hookType);
  });
  for (const hook of sortedHooks) {
    emitHook(hook, depth + 1, lines);
    lines.push('');
  }

  for (let i = 0; i < (suite.tests || []).length; i++) {
    const child = suite.tests[i];
    emitBodyNode(child, depth + 1, lines);
    if (i < suite.tests.length - 1) {
      lines.push('');
    }
  }

  lines.push(`${indent}});`);
}

function emitTestCase(tc, depth, lines) {
  const indent = '  '.repeat(depth);
  const name = escapeSingleQuotes(tc.name || 'unnamed');
  const mod = testModifier(tc);

  lines.push(`${indent}it${mod}('${name}', async () => {`);

  for (const child of tc.body || []) {
    emitBodyNode(child, depth + 1, lines);
  }

  lines.push(`${indent}});`);
}

function emitHook(hook, depth, lines) {
  const indent = '  '.repeat(depth);
  const hookType = hook.hookType || 'beforeEach';

  lines.push(`${indent}${hookType}(async () => {`);

  for (const child of hook.body || []) {
    emitBodyNode(child, depth + 1, lines);
  }

  lines.push(`${indent}});`);
}

// ═══════════════════════════════════════════════════════════════════════
// Assertion emitter — Puppeteer page-level with Jest expect
// ═══════════════════════════════════════════════════════════════════════

export function emitAssertion(node) {
  if (!(node instanceof Assertion)) {
    return { code: '', supported: false };
  }

  const { kind, subject, expected, isNegated } = node;

  if (!subject) {
    return {
      code: `// HAMLET-TODO: Assertion without subject (kind: "${kind}")`,
      supported: false,
    };
  }

  const not = isNegated ? '.not' : '';
  const handler = PUPPETEER_ASSERTIONS[kind];
  if (handler) {
    const code = handler(subject, expected, not);
    if (code) return { code, supported: true };
  }

  const original = node.originalSource ? node.originalSource.trim() : '';
  return {
    code:
      `// HAMLET-TODO: Unsupported assertion kind "${kind}"` +
      (original ? `\n// Original: ${original}` : ''),
    supported: false,
  };
}

// ═══════════════════════════════════════════════════════════════════════
// Navigation emitter — Puppeteer page API
// ═══════════════════════════════════════════════════════════════════════

export function emitNavigation(node) {
  if (!(node instanceof Navigation)) {
    return { code: '', supported: false };
  }

  switch (node.action) {
    case 'visit': {
      if (!node.url) return { code: '', supported: false };
      return { code: `await page.goto('${node.url}')`, supported: true };
    }
    case 'goBack':
      return { code: 'await page.goBack()', supported: true };
    case 'goForward':
      return { code: 'await page.goForward()', supported: true };
    case 'reload':
      return { code: 'await page.reload()', supported: true };
    default:
      return {
        code: `// HAMLET-TODO: Unsupported navigation action "${node.action}"`,
        supported: false,
      };
  }
}

// ═══════════════════════════════════════════════════════════════════════
// Assertion kind → Puppeteer mapping (manual page evaluation + Jest expect)
// ═══════════════════════════════════════════════════════════════════════

const PUPPETEER_ASSERTIONS = {
  'be.visible': (subj, _exp, not) =>
    `const el = await page.$('${subj}');\n    const isVisible = el && await el.isIntersectingViewport();\n    expect(isVisible)${not}.toBe(true)`,
  exist: (subj, _exp, not) =>
    `const el = await page.$('${subj}');\n    expect(el)${not}.not.toBeNull()`,
  contain: (subj, exp, not) =>
    `const text = await page.$eval('${subj}', el => el.textContent);\n    expect(text)${not}.toContain(${exp})`,
  'have.text': (subj, exp, not) =>
    `const text = await page.$eval('${subj}', el => el.textContent);\n    expect(text.trim())${not}.toBe(${exp})`,
  'have.value': (subj, exp, not) =>
    `const val = await page.$eval('${subj}', el => el.value);\n    expect(val)${not}.toBe(${exp})`,
  'have.attr': (subj, exp, not) =>
    `const attr = await page.$eval('${subj}', el => el.getAttribute(${exp}));\n    expect(attr)${not}.toBeTruthy()`,
  'url.include': (_subj, exp, not) =>
    `expect(page.url())${not}.toContain(${exp})`,
  'url.equal': (_subj, exp, not) => `expect(page.url())${not}.toBe(${exp})`,
  'title.equal': (_subj, exp, not) =>
    `const title = await page.title();\n    expect(title)${not}.toBe(${exp})`,
  equal: (subj, exp, not) => `expect(${subj})${not}.toBe(${exp})`,
  'be.true': (subj, _exp, not) => `expect(${subj})${not}.toBe(true)`,
  'be.false': (subj, _exp, not) => `expect(${subj})${not}.toBe(false)`,
  'be.null': (subj, _exp, not) => `expect(${subj})${not}.toBeNull()`,
  'be.undefined': (subj, _exp, not) => `expect(${subj})${not}.toBeUndefined()`,
  include: (subj, exp, not) => `expect(${subj})${not}.toContain(${exp})`,
  'have.property': (subj, exp, not) =>
    `expect(${subj})${not}.toHaveProperty(${exp})`,
};

// ═══════════════════════════════════════════════════════════════════════
// Helpers
// ═══════════════════════════════════════════════════════════════════════

function suiteModifier(suite) {
  for (const mod of suite.modifiers || []) {
    if (mod instanceof Modifier) {
      if (mod.modifierType === 'only') return '.only';
      if (mod.modifierType === 'skip') return '.skip';
    }
  }
  return '';
}

function testModifier(tc) {
  for (const mod of tc.modifiers || []) {
    if (mod instanceof Modifier) {
      if (mod.modifierType === 'only') return '.only';
      if (mod.modifierType === 'skip') return '.skip';
    }
  }
  return '';
}

function escapeSingleQuotes(str) {
  return str.replace(/'/g, "\\'");
}
