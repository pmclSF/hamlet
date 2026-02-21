/**
 * IR-driven Mocha+Chai+Sinon emitter.
 *
 * Generates Mocha code with Chai assertions and Sinon mocking from IR nodes.
 *
 * Provides three unified exports:
 *   emitNode(node)             — dispatch by type → { code, supported }
 *   matchesBaseline(line, node) — check if a baseline line corresponds to a node
 *   emitFullFile(ir)           — generate complete Mocha+Chai file from IR tree walk
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
  if (node instanceof MockCall) return emitMockCall(node);
  return { code: '', supported: false };
}

export function matchesBaseline(line, node) {
  if (node instanceof Assertion) {
    return /\bexpect\s*\(/.test(line);
  }
  if (node instanceof MockCall) {
    return /\bsinon\.\w+/.test(line);
  }
  return false;
}

/**
 * Generate a complete Mocha+Chai file from an IR tree.
 */
export function emitFullFile(ir) {
  if (!(ir instanceof TestFile)) return null;

  const lines = [];
  const needsSinon = hasMockNodes(ir.body || []);

  // Emit Chai import
  lines.push("const { expect } = require('chai');");
  if (needsSinon) {
    lines.push("const sinon = require('sinon');");
  }

  // Emit user imports — skip source framework imports
  for (const imp of ir.imports || []) {
    if (imp instanceof ImportStatement) {
      const src = imp.source || '';
      if (/^(@jest\/globals|vitest|jasmine)$/.test(src)) continue;
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
// Helpers
// ═══════════════════════════════════════════════════════════════════════

function hasMockNodes(nodes) {
  for (const node of nodes) {
    if (node instanceof MockCall) return true;
    if (node instanceof TestSuite) {
      if (hasMockNodes(node.tests || [])) return true;
      if (hasMockNodes(node.hooks || [])) return true;
    }
    if (node instanceof TestCase && hasMockNodes(node.body || [])) return true;
    if (node instanceof Hook && hasMockNodes(node.body || [])) return true;
  }
  return false;
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

  lines.push(`${indent}describe${mod}('${name}', function () {`);

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
  const asyncPrefix = tc.isAsync ? 'async ' : '';

  lines.push(`${indent}it${mod}('${name}', ${asyncPrefix}function () {`);

  for (const child of tc.body || []) {
    emitBodyNode(child, depth + 1, lines);
  }

  lines.push(`${indent}});`);
}

function emitHook(hook, depth, lines) {
  const indent = '  '.repeat(depth);
  // Mocha uses before/after instead of beforeAll/afterAll
  const hookMap = {
    beforeAll: 'before',
    afterAll: 'after',
    beforeEach: 'beforeEach',
    afterEach: 'afterEach',
  };
  const hookType = hookMap[hook.hookType] || hook.hookType || 'beforeEach';
  const asyncPrefix = hook.isAsync ? 'async ' : '';

  lines.push(`${indent}${hookType}(${asyncPrefix}function () {`);

  for (const child of hook.body || []) {
    emitBodyNode(child, depth + 1, lines);
  }

  lines.push(`${indent}});`);
}

// ═══════════════════════════════════════════════════════════════════════
// Assertion emitter — Chai expect chains
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
  const handler = CHAI_ASSERTIONS[kind];
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
// MockCall emitter — Sinon API
// ═══════════════════════════════════════════════════════════════════════

export function emitMockCall(node) {
  if (!(node instanceof MockCall)) {
    return { code: '', supported: false };
  }

  const { kind, target } = node;

  switch (kind) {
    case 'createMock':
      return { code: 'sinon.stub()', supported: true };
    case 'spyOnMethod':
      return todoMock(node, 'sinon.spy() — needs object and method args');
    case 'fakeTimers':
      return { code: 'sinon.useFakeTimers()', supported: true };
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
// Assertion kind → Chai expect chain mapping
// ═══════════════════════════════════════════════════════════════════════

const CHAI_ASSERTIONS = {
  strictEqual: (subj, exp, not) => `expect(${subj}).to${not}.equal(${exp})`,
  deepEqual: (subj, exp, not) => `expect(${subj}).to${not}.deep.equal(${exp})`,
  equal: (subj, exp, not) => `expect(${subj}).to${not}.equal(${exp})`,
  truthy: (subj, _exp, not) => `expect(${subj}).to${not}.be.ok`,
  falsy: (subj, _exp, not) => `expect(${subj}).to${not}.be.ok`,
  isNull: (subj, _exp, not) => `expect(${subj}).to${not}.be.null`,
  isUndefined: (subj, _exp, not) => `expect(${subj}).to${not}.be.undefined`,
  isDefined: (subj, _exp, not) => `expect(${subj}).to${not}.exist`,
  isNaN: (subj, _exp, not) => `expect(${subj}).to${not}.be.NaN`,
  instanceOf: (subj, exp, not) =>
    `expect(${subj}).to${not}.be.an.instanceOf(${exp})`,
  matches: (subj, exp, not) => `expect(${subj}).to${not}.match(${exp})`,
  contains: (subj, exp, not) => `expect(${subj}).to${not}.include(${exp})`,
  containsEqual: (subj, exp, not) =>
    `expect(${subj}).to${not}.deep.include(${exp})`,
  hasLength: (subj, exp, not) =>
    `expect(${subj}).to${not}.have.lengthOf(${exp})`,
  hasProperty: (subj, exp, not) =>
    `expect(${subj}).to${not}.have.property(${exp})`,
  greaterThan: (subj, exp, not) => `expect(${subj}).to${not}.be.above(${exp})`,
  lessThan: (subj, exp, not) => `expect(${subj}).to${not}.be.below(${exp})`,
  greaterOrEqual: (subj, exp, not) =>
    `expect(${subj}).to${not}.be.at.least(${exp})`,
  lessOrEqual: (subj, exp, not) =>
    `expect(${subj}).to${not}.be.at.most(${exp})`,
  closeTo: (subj, exp, not) => `expect(${subj}).to${not}.be.closeTo(${exp})`,
  throws: (subj, _exp, not) => `expect(${subj}).to${not}.throw`,
  called: (subj, _exp, not) => `expect(${subj}).to${not}.have.been.called`,
  calledWith: (subj, exp, not) =>
    `expect(${subj}).to${not}.have.been.calledWith(${exp || ''})`,
  calledTimes: (subj, exp, not) =>
    `expect(${subj}).to${not}.have.callCount(${exp})`,
  snapshot: (subj, _exp, _not) =>
    `// HAMLET-TODO: Mocha has no built-in snapshot support — use chai-jest-snapshot\n// expect(${subj}).to.matchSnapshot()`,
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
      if (mod.modifierType === 'pending') return '.skip';
    }
  }
  return '';
}

function escapeSingleQuotes(str) {
  return str.replace(/'/g, "\\'");
}
