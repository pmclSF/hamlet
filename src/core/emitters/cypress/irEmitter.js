/**
 * IR-driven Cypress emitter.
 *
 * Generates Cypress code from IR nodes. Cypress uses:
 * - describe/it (globals, no imports needed)
 * - cy.get(sel).should('kind', expected) for assertions
 * - cy.visit(url), cy.go('back'), cy.reload() for navigation
 * - cy.intercept() for network mocking
 * - Synchronous API (no async/await)
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
  if (node instanceof MockCall) return emitMockCall(node);
  return { code: '', supported: false };
}

export function matchesBaseline(line, node) {
  if (node instanceof Assertion) {
    return /\.should\s*\(/.test(line) || /\bexpect\s*\(/.test(line);
  }
  if (node instanceof Navigation) {
    return (
      /cy\.visit\s*\(/.test(line) ||
      /cy\.go\s*\(/.test(line) ||
      /cy\.reload\s*\(/.test(line)
    );
  }
  if (node instanceof MockCall) {
    return /cy\.intercept\s*\(/.test(line);
  }
  return false;
}

/**
 * Generate a complete Cypress file from an IR tree.
 */
export function emitFullFile(ir) {
  if (!(ir instanceof TestFile)) return null;

  const lines = [];

  // Cypress uses globals — no framework import needed
  // Emit non-framework user imports
  for (const imp of ir.imports || []) {
    if (imp instanceof ImportStatement) {
      const src = imp.source || '';
      if (/@playwright|@wdio|webdriverio|puppeteer|testcafe/.test(src))
        continue;
      lines.push(imp.originalSource || src);
    }
  }

  if (lines.length > 0) lines.push('');

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
    lines.push(indent + result.code);
  } else if (node instanceof Navigation) {
    const result = emitNavigation(node);
    lines.push(indent + result.code);
  } else if (node instanceof Comment) {
    lines.push(indent + (node.text || '').trim());
  } else if (node instanceof RawCode) {
    const code = (node.code || '').trim();
    if (code) lines.push(indent + code);
  } else if (node instanceof MockCall) {
    const result = emitMockCall(node);
    for (const codeLine of result.code.split('\n')) {
      lines.push(indent + codeLine);
    }
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

  // Cypress is synchronous — no async
  lines.push(`${indent}it${mod}('${name}', () => {`);

  for (const child of tc.body || []) {
    emitBodyNode(child, depth + 1, lines);
  }

  lines.push(`${indent}});`);
}

function emitHook(hook, depth, lines) {
  const indent = '  '.repeat(depth);
  // Cypress uses before/after aliases
  const hookMap = {
    beforeAll: 'before',
    afterAll: 'after',
    beforeEach: 'beforeEach',
    afterEach: 'afterEach',
  };
  const hookType = hookMap[hook.hookType] || hook.hookType || 'beforeEach';

  lines.push(`${indent}${hookType}(() => {`);

  for (const child of hook.body || []) {
    emitBodyNode(child, depth + 1, lines);
  }

  lines.push(`${indent}});`);
}

// ═══════════════════════════════════════════════════════════════════════
// Assertion emitter — Cypress .should() chains
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

  const neg = isNegated ? 'not.' : '';
  const handler = CYPRESS_ASSERTIONS[kind];
  if (handler) {
    const code = handler(subject, expected, neg);
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
// Navigation emitter — Cypress cy.visit/cy.go/cy.reload
// ═══════════════════════════════════════════════════════════════════════

export function emitNavigation(node) {
  if (!(node instanceof Navigation)) {
    return { code: '', supported: false };
  }

  switch (node.action) {
    case 'visit': {
      if (!node.url) return { code: '', supported: false };
      return { code: `cy.visit('${node.url}')`, supported: true };
    }
    case 'goBack':
      return { code: "cy.go('back')", supported: true };
    case 'goForward':
      return { code: "cy.go('forward')", supported: true };
    case 'reload':
      return { code: 'cy.reload()', supported: true };
    default:
      return {
        code: `// HAMLET-TODO: Unsupported navigation action "${node.action}"`,
        supported: false,
      };
  }
}

// ═══════════════════════════════════════════════════════════════════════
// MockCall emitter — Cypress cy.intercept
// ═══════════════════════════════════════════════════════════════════════

export function emitMockCall(node) {
  if (!(node instanceof MockCall)) {
    return { code: '', supported: false };
  }

  const { kind, target, args, returnValue } = node;

  switch (kind) {
    case 'networkIntercept': {
      if (!target) {
        return todoMock(node, 'cy.intercept() without target URL');
      }
      const method = args && args[0] ? `'${args[0]}', ` : '';
      if (returnValue) {
        return {
          code: `cy.intercept(${method}'${target}', ${returnValue})`,
          supported: true,
        };
      }
      return {
        code: `cy.intercept(${method}'${target}')`,
        supported: true,
      };
    }
    default:
      return todoMock(node, `Unsupported mock kind "${kind}"`);
  }
}

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
// Assertion kind → Cypress .should() mapping
// ═══════════════════════════════════════════════════════════════════════

const CYPRESS_ASSERTIONS = {
  'be.visible': (subj, _exp, neg) =>
    `cy.get('${subj}').should('${neg}be.visible')`,
  exist: (subj, _exp, neg) => `cy.get('${subj}').should('${neg}exist')`,
  contain: (subj, exp, neg) =>
    `cy.get('${subj}').should('${neg}contain', ${exp})`,
  'have.text': (subj, exp, neg) =>
    `cy.get('${subj}').should('${neg}have.text', ${exp})`,
  'have.length': (subj, exp, neg) =>
    `cy.get('${subj}').should('${neg}have.length', ${exp})`,
  'have.attr': (subj, exp, neg) =>
    `cy.get('${subj}').should('${neg}have.attr', ${exp})`,
  'have.class': (subj, exp, neg) =>
    `cy.get('${subj}').should('${neg}have.class', ${exp})`,
  'have.value': (subj, exp, neg) =>
    `cy.get('${subj}').should('${neg}have.value', ${exp})`,
  'be.checked': (subj, _exp, neg) =>
    `cy.get('${subj}').should('${neg}be.checked')`,
  'be.disabled': (subj, _exp, neg) =>
    `cy.get('${subj}').should('${neg}be.disabled')`,
  'be.enabled': (subj, _exp, neg) =>
    `cy.get('${subj}').should('${neg}be.enabled')`,
  'be.empty': (subj, _exp, neg) => `cy.get('${subj}').should('${neg}be.empty')`,
  'be.focused': (subj, _exp, neg) =>
    `cy.get('${subj}').should('${neg}be.focused')`,
  'url.include': (_subj, exp, neg) =>
    `cy.url().should('${neg}include', ${exp})`,
  'url.equal': (_subj, exp, neg) => `cy.url().should('${neg}eq', ${exp})`,
  'title.equal': (_subj, exp, neg) => `cy.title().should('${neg}eq', ${exp})`,
  equal: (subj, exp, neg) => `expect(${subj}).to.${neg}equal(${exp})`,
  'be.true': (subj, _exp, neg) => `expect(${subj}).to.${neg}be.true`,
  'be.false': (subj, _exp, neg) => `expect(${subj}).to.${neg}be.false`,
  'be.null': (subj, _exp, neg) => `expect(${subj}).to.${neg}be.null`,
  'be.undefined': (subj, _exp, neg) => `expect(${subj}).to.${neg}be.undefined`,
  include: (subj, exp, neg) => `expect(${subj}).to.${neg}include(${exp})`,
  'have.property': (subj, exp, neg) =>
    `expect(${subj}).to.${neg}have.property(${exp})`,
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
