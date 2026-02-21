/**
 * IR-driven Playwright emitter.
 *
 * Generates Playwright code from IR nodes. Provides three unified exports:
 *   emitNode(node)             — dispatch by type → { code, supported }
 *   matchesBaseline(line, node) — check if a baseline line corresponds to a node
 *   emitFullFile(ir)           — stub for future full-file generation
 *
 * Also re-exports emitAssertion/emitNavigation for backward compatibility.
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
  ParameterSet,
} from '../../ir.js';

// ═══════════════════════════════════════════════════════════════════════
// Unified interface
// ═══════════════════════════════════════════════════════════════════════

/**
 * Emit a single Playwright code fragment from any supported IR node.
 *
 * @param {import('../../ir.js').IRNode} node
 * @returns {{ code: string, supported: boolean }}
 */
export function emitNode(node) {
  if (node instanceof Assertion) return emitAssertion(node);
  if (node instanceof Navigation) return emitNavigation(node);
  if (node instanceof MockCall) return emitMockCall(node);
  return { code: '', supported: false };
}

/**
 * Check if a line in baseline output corresponds to a given IR node.
 *
 * @param {string} line - Trimmed line from baseline output
 * @param {import('../../ir.js').IRNode} node - IR node (Assertion or Navigation)
 * @returns {boolean}
 */
export function matchesBaseline(line, node) {
  if (node instanceof Navigation) {
    return matchesNavigationLine(line, node);
  }

  if (node instanceof Assertion) {
    return matchesAssertionLine(line, node);
  }

  if (node instanceof MockCall) {
    return matchesMockLine(line, node);
  }

  return false;
}

/**
 * Generate a complete Playwright file from an IR tree.
 *
 * Tree-walks the IR and emits a full Playwright test file including
 * imports, test.describe blocks, test() cases, hooks, assertions,
 * navigation, raw code pass-through, and HAMLET-TODO for unsupported nodes.
 *
 * @param {import('../../ir.js').TestFile} ir
 * @returns {string|null}
 */
export function emitFullFile(ir) {
  if (!(ir instanceof TestFile)) return null;

  const lines = [];

  // Emit imports — always include Playwright test import
  lines.push("import { test, expect } from '@playwright/test';");

  // Emit any additional non-framework imports from IR
  for (const imp of ir.imports || []) {
    if (imp instanceof ImportStatement) {
      const src = imp.source || '';
      // Skip Cypress/source framework imports
      if (/cypress|@cypress/.test(src)) continue;
      lines.push(src);
    }
  }

  lines.push('');

  // Emit body nodes
  for (const node of ir.body || []) {
    emitBodyNode(node, 0, lines);
  }

  return lines.join('\n') + '\n';
}

/**
 * Emit a body-level IR node at the given indent depth.
 */
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
    const result = emitMockCall(node);
    if (result.supported) {
      lines.push(indent + result.code + ';');
    } else {
      // Multi-line code may contain embedded newlines (TODO + Original)
      for (const codeLine of result.code.split('\n')) {
        lines.push(indent + codeLine);
      }
    }
  } else if (node instanceof SharedVariable) {
    lines.push(
      indent +
        `// HAMLET-TODO: Shared variable "${node.name}" conversion not yet supported`
    );
  } else {
    // Unknown node type — emit TODO
    const original = node.originalSource
      ? node.originalSource.trim()
      : node.type || 'unknown';
    lines.push(indent + `// HAMLET-TODO: Unsupported node (${original})`);
  }
}

/**
 * Emit a test.describe block with hooks and tests.
 */
function emitSuite(suite, depth, lines) {
  const indent = '  '.repeat(depth);
  const name = escapeSingleQuotes(suite.name || 'unnamed');

  // Handle modifiers (.only, .skip)
  const mod = suiteModifier(suite);
  lines.push(`${indent}test.describe${mod}('${name}', () => {`);

  // Emit hooks first (ordered: beforeAll, beforeEach, afterEach, afterAll)
  const hookOrder = ['beforeAll', 'beforeEach', 'afterEach', 'afterAll'];
  const sortedHooks = [...(suite.hooks || [])].sort((a, b) => {
    return hookOrder.indexOf(a.hookType) - hookOrder.indexOf(b.hookType);
  });
  for (const hook of sortedHooks) {
    emitHook(hook, depth + 1, lines);
    lines.push('');
  }

  // Emit tests and nested suites
  for (let i = 0; i < (suite.tests || []).length; i++) {
    const child = suite.tests[i];
    emitBodyNode(child, depth + 1, lines);
    if (i < suite.tests.length - 1) {
      lines.push('');
    }
  }

  lines.push(`${indent}});`);
}

/**
 * Emit a test() block.
 */
function emitTestCase(tc, depth, lines) {
  const indent = '  '.repeat(depth);
  const name = escapeSingleQuotes(tc.name || 'unnamed');

  const mod = testModifier(tc);
  lines.push(`${indent}test${mod}('${name}', async ({ page }) => {`);

  for (const child of tc.body || []) {
    emitBodyNode(child, depth + 1, lines);
  }

  lines.push(`${indent}});`);
}

/**
 * Emit a hook block (test.beforeEach, test.afterAll, etc.).
 */
function emitHook(hook, depth, lines) {
  const indent = '  '.repeat(depth);
  const hookType = hook.hookType || 'beforeEach';
  lines.push(`${indent}test.${hookType}(async ({ page }) => {`);

  for (const child of hook.body || []) {
    emitBodyNode(child, depth + 1, lines);
  }

  lines.push(`${indent}});`);
}

/**
 * Get modifier suffix for a suite (e.g., '.only', '.skip', or '').
 */
function suiteModifier(suite) {
  for (const mod of suite.modifiers || []) {
    if (mod instanceof Modifier) {
      if (mod.modifierType === 'only') return '.only';
      if (mod.modifierType === 'skip') return '.skip';
    }
  }
  return '';
}

/**
 * Get modifier suffix for a test (e.g., '.only', '.skip', or '').
 */
function testModifier(tc) {
  for (const mod of tc.modifiers || []) {
    if (mod instanceof Modifier) {
      if (mod.modifierType === 'only') return '.only';
      if (mod.modifierType === 'skip') return '.skip';
    }
  }
  return '';
}

/**
 * Escape single quotes in a string for safe inclusion in '...' delimiters.
 */
function escapeSingleQuotes(str) {
  return str.replace(/'/g, "\\'");
}

// ═══════════════════════════════════════════════════════════════════════
// Per-type emitters (also exported for backward compatibility)
// ═══════════════════════════════════════════════════════════════════════

/**
 * Emit a single Playwright assertion line from an IR Assertion node.
 *
 * @param {Assertion} node - IR assertion node with kind, subject, expected, isNegated
 * @returns {{ code: string, supported: boolean }} The emitted code and whether it was fully supported
 */
export function emitAssertion(node) {
  if (!(node instanceof Assertion)) {
    return { code: '', supported: false };
  }

  const { kind, subject, expected, isNegated } = node;

  // --- Locator-based assertions (cy.get(sel).should(...)) ---
  const locatorHandler = LOCATOR_ASSERTIONS[kind];
  if (locatorHandler && subject && !subject.startsWith('cy.')) {
    const code = locatorHandler(subject, expected, isNegated);
    if (code) return { code, supported: true };
  }

  // --- URL assertions ---
  if (kind === 'url.include' || kind === 'url.equal') {
    const val = quoteIfNeeded(expected);
    const not = isNegated ? '.not' : '';
    if (kind === 'url.include') {
      return {
        code: `await expect(page)${not}.toHaveURL(new RegExp(${val}))`,
        supported: true,
      };
    }
    return {
      code: `await expect(page)${not}.toHaveURL(${val})`,
      supported: true,
    };
  }

  // --- Title assertions ---
  if (kind === 'title.equal') {
    const val = quoteIfNeeded(expected);
    const not = isNegated ? '.not' : '';
    return {
      code: `await expect(page)${not}.toHaveTitle(${val})`,
      supported: true,
    };
  }

  // --- Value-based expect() assertions ---
  const valueHandler = VALUE_ASSERTIONS[kind];
  if (valueHandler) {
    const code = valueHandler(subject, expected, isNegated);
    if (code) return { code, supported: true };
  }

  // --- Unsupported: emit HAMLET-TODO ---
  const original = node.originalSource ? node.originalSource.trim() : '';
  return {
    code:
      `// HAMLET-TODO: Unsupported assertion kind "${kind}"` +
      (original ? `\n// Original: ${original}` : ''),
    supported: false,
  };
}

/**
 * Emit a single Playwright navigation line from an IR Navigation node.
 *
 * @param {Navigation} node - IR navigation node with action, url, options
 * @returns {{ code: string, supported: boolean }}
 */
export function emitNavigation(node) {
  if (!(node instanceof Navigation)) {
    return { code: '', supported: false };
  }

  const { action, url } = node;

  switch (action) {
    case 'visit': {
      if (!url) {
        return { code: '', supported: false };
      }
      // Check if url looks like a variable reference (no quotes needed)
      const isVariable = /^[a-zA-Z_$]/.test(url);
      const urlArg = isVariable ? url : `'${url}'`;
      return { code: `await page.goto(${urlArg})`, supported: true };
    }
    case 'goBack':
      return { code: 'await page.goBack()', supported: true };
    case 'goForward':
      return { code: 'await page.goForward()', supported: true };
    case 'reload':
      return { code: 'await page.reload()', supported: true };
    default:
      return {
        code:
          `// HAMLET-TODO: Unsupported navigation action "${action}"` +
          (node.originalSource
            ? `\n// Original: ${node.originalSource.trim()}`
            : ''),
        supported: false,
      };
  }
}

/**
 * Emit a Playwright mock/intercept line from an IR MockCall node.
 *
 * @param {MockCall} node - IR mock call node with kind, target, args, returnValue
 * @returns {{ code: string, supported: boolean }}
 */
export function emitMockCall(node) {
  if (!(node instanceof MockCall)) {
    return { code: '', supported: false };
  }

  const { kind, target, args, returnValue } = node;

  switch (kind) {
    case 'networkIntercept': {
      if (!target) {
        return todoMock(node, 'networkIntercept with no target URL');
      }
      const urlArg = quoteIfNeeded(target);
      if (returnValue) {
        return {
          code: `await page.route(${urlArg}, route => route.fulfill(${returnValue}))`,
          supported: true,
        };
      }
      // Spy form — no response stub
      return {
        code: `await page.route(${urlArg}, route => route.continue())`,
        supported: true,
      };
    }
    case 'createStub':
      return todoMock(node, 'cy.stub() has no direct Playwright equivalent');
    case 'createSpy':
      return todoMock(node, 'cy.spy() has no direct Playwright equivalent');
    case 'fakeTimers':
      return todoMock(
        node,
        'cy.clock() — use page.clock or manual timer control'
      );
    case 'advanceTimers': {
      const ms = args && args[0] ? args[0] : '0';
      return todoMock(
        node,
        `cy.tick(${ms}) — use page.clock.fastForward(${ms})`
      );
    }
    default:
      return todoMock(node, `Unsupported mock kind "${kind}"`);
  }
}

/**
 * Generate a HAMLET-TODO for an unsupported mock operation.
 */
function todoMock(node, reason) {
  const original = node.originalSource ? node.originalSource.trim() : '';
  return {
    code:
      `// HAMLET-TODO: ${reason}` +
      (original ? `\n// Original: ${original}` : ''),
    supported: false,
  };
}

// ═══════════════════════════════════════════════════════════════════════
// Baseline matching (moved from ConversionPipeline)
// ═══════════════════════════════════════════════════════════════════════

/**
 * Check if a baseline line corresponds to a converted assertion.
 * @param {string} line - Trimmed line
 * @param {Assertion} node
 * @returns {boolean}
 */
function matchesAssertionLine(line, node) {
  if (!/(?:await\s+)?expect\(/.test(line)) return false;

  const subject = node.subject || '';
  if (!subject) return false;

  // For locator-based: check if the line references the selector
  if (line.includes(subject)) return true;

  // For URL/title assertions: match page-level expects
  if (
    subject === 'cy.url()' &&
    line.includes('expect(page)') &&
    line.includes('URL')
  ) {
    return true;
  }
  if (
    subject === 'cy.title()' &&
    line.includes('expect(page)') &&
    line.includes('Title')
  ) {
    return true;
  }

  return false;
}

/**
 * Check if a baseline line corresponds to a converted navigation.
 * @param {string} line - Trimmed line
 * @param {Navigation} node
 * @returns {boolean}
 */
function matchesNavigationLine(line, node) {
  switch (node.action) {
    case 'visit':
      return /page\.goto\s*\(/.test(line);
    case 'goBack':
      return /page\.goBack\s*\(/.test(line);
    case 'goForward':
      return /page\.goForward\s*\(/.test(line);
    case 'reload':
      return /page\.reload\s*\(/.test(line);
    default:
      return false;
  }
}

/**
 * Check if a baseline line corresponds to a converted mock call.
 * @param {string} line - Trimmed line
 * @param {MockCall} node
 * @returns {boolean}
 */
function matchesMockLine(line, node) {
  if (node.kind === 'networkIntercept') {
    return /page\.route\s*\(/.test(line);
  }
  return false;
}

// ═══════════════════════════════════════════════════════════════════════
// Locator-based assertions (map kind → Playwright expect method)
// ═══════════════════════════════════════════════════════════════════════

const LOCATOR_ASSERTIONS = {
  'be.visible': (sel, _exp, neg) => {
    const not = neg ? '.not' : '';
    return `await expect(page.locator('${sel}'))${not}.toBeVisible()`;
  },
  exist: (sel, _exp, neg) => {
    const not = neg ? '.not' : '';
    return `await expect(page.locator('${sel}'))${not}.toBeAttached()`;
  },
  contain: (sel, exp, neg) => {
    const not = neg ? '.not' : '';
    return `await expect(page.locator('${sel}'))${not}.toContainText(${quoteIfNeeded(exp)})`;
  },
  'have.text': (sel, exp, neg) => {
    const not = neg ? '.not' : '';
    return `await expect(page.locator('${sel}'))${not}.toHaveText(${quoteIfNeeded(exp)})`;
  },
  'have.length': (sel, exp, neg) => {
    const not = neg ? '.not' : '';
    return `await expect(page.locator('${sel}'))${not}.toHaveCount(${exp})`;
  },
  'have.attr': (sel, exp, neg) => {
    const not = neg ? '.not' : '';
    // expected may be "key, value" or just "key"
    const args = exp ? exp.toString() : '';
    return `await expect(page.locator('${sel}'))${not}.toHaveAttribute(${args})`;
  },
  'have.class': (sel, exp, neg) => {
    const not = neg ? '.not' : '';
    return `await expect(page.locator('${sel}'))${not}.toHaveClass(${quoteIfNeeded(exp)})`;
  },
  'have.value': (sel, exp, neg) => {
    const not = neg ? '.not' : '';
    return `await expect(page.locator('${sel}'))${not}.toHaveValue(${quoteIfNeeded(exp)})`;
  },
  'be.checked': (sel, _exp, neg) => {
    const not = neg ? '.not' : '';
    return `await expect(page.locator('${sel}'))${not}.toBeChecked()`;
  },
  'be.disabled': (sel, _exp, neg) => {
    const not = neg ? '.not' : '';
    return `await expect(page.locator('${sel}'))${not}.toBeDisabled()`;
  },
  'be.enabled': (sel, _exp, neg) => {
    const not = neg ? '.not' : '';
    return `await expect(page.locator('${sel}'))${not}.toBeEnabled()`;
  },
  'be.empty': (sel, _exp, neg) => {
    const not = neg ? '.not' : '';
    return `await expect(page.locator('${sel}'))${not}.toBeEmpty()`;
  },
  'be.focused': (sel, _exp, neg) => {
    const not = neg ? '.not' : '';
    return `await expect(page.locator('${sel}'))${not}.toBeFocused()`;
  },
  match: (sel, exp, neg) => {
    const not = neg ? '.not' : '';
    // exp is typically a regex like /pattern/
    return `await expect(page.locator('${sel}'))${not}.toHaveText(${exp})`;
  },
};

// ═══════════════════════════════════════════════════════════════════════
// Value-based assertions (expect(expr).to.matcher)
// ═══════════════════════════════════════════════════════════════════════

const VALUE_ASSERTIONS = {
  equal: (subj, exp, neg) => {
    const not = neg ? '.not' : '';
    return `expect(${subj})${not}.toBe(${quoteIfNeeded(exp)})`;
  },
  'be.true': (subj, _exp, neg) => {
    const not = neg ? '.not' : '';
    return `expect(${subj})${not}.toBe(true)`;
  },
  'be.false': (subj, _exp, neg) => {
    const not = neg ? '.not' : '';
    return `expect(${subj})${not}.toBe(false)`;
  },
  'be.null': (subj, _exp, neg) => {
    const not = neg ? '.not' : '';
    return `expect(${subj})${not}.toBeNull()`;
  },
  'be.undefined': (subj, _exp, neg) => {
    const not = neg ? '.not' : '';
    return `expect(${subj})${not}.toBeUndefined()`;
  },
  include: (subj, exp, neg) => {
    const not = neg ? '.not' : '';
    return `expect(${subj})${not}.toContain(${quoteIfNeeded(exp)})`;
  },
  'have.property': (subj, exp, neg) => {
    const not = neg ? '.not' : '';
    return `expect(${subj})${not}.toHaveProperty(${quoteIfNeeded(exp)})`;
  },
};

// ═══════════════════════════════════════════════════════════════════════
// Helpers
// ═══════════════════════════════════════════════════════════════════════

/**
 * Wrap a value in single quotes if it doesn't already look like a
 * quoted string, number, regex, boolean, or variable reference.
 */
function quoteIfNeeded(value) {
  if (value === null || value === undefined) return 'undefined';
  const str = String(value);
  // Already quoted
  if (/^(['"`]).*\1$/.test(str)) return str;
  // Number
  if (/^\d+(\.\d+)?$/.test(str)) return str;
  // Boolean/null/undefined keywords
  if (/^(true|false|null|undefined)$/.test(str)) return str;
  // Regex literal
  if (/^\/.*\/$/.test(str)) return str;
  // Variable reference (starts with letter/$ or parenthesized)
  if (/^[a-zA-Z_$]/.test(str) || str.startsWith('(')) return str;
  // Otherwise quote it
  return `'${str}'`;
}
