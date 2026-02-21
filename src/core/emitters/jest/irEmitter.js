/**
 * IR-driven Jest emitter.
 *
 * Generates Jest code from IR nodes. Provides three unified exports:
 *   emitNode(node)             — dispatch by type → { code, supported }
 *   matchesBaseline(line, node) — check if a baseline line corresponds to a node
 *   emitFullFile(ir)           — generate complete Jest file from IR tree walk
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

/**
 * Emit a single Jest code fragment from any supported IR node.
 *
 * @param {import('../../ir.js').IRNode} node
 * @returns {{ code: string, supported: boolean }}
 */
export function emitNode(node) {
  if (node instanceof Assertion) return emitAssertion(node);
  if (node instanceof MockCall) return emitMockCall(node);
  return { code: '', supported: false };
}

/**
 * Check if a baseline line corresponds to a given IR node.
 *
 * @param {string} line - Trimmed line from baseline output
 * @param {import('../../ir.js').IRNode} node
 * @returns {boolean}
 */
export function matchesBaseline(line, node) {
  if (node instanceof Assertion) {
    return /\bexpect\s*\(/.test(line);
  }
  if (node instanceof MockCall) {
    return /\bjest\.\w+/.test(line);
  }
  return false;
}

/**
 * Generate a complete Jest file from an IR tree.
 *
 * Tree-walks the IR and emits a full Jest test file including
 * describe blocks, it/test cases, hooks, assertions,
 * raw code pass-through, and HAMLET-TODO for unsupported nodes.
 *
 * @param {import('../../ir.js').TestFile} ir
 * @returns {string|null}
 */
export function emitFullFile(ir) {
  if (!(ir instanceof TestFile)) return null;

  const lines = [];

  // Emit imports — skip source framework imports (chai, sinon, etc.)
  for (const imp of ir.imports || []) {
    if (imp instanceof ImportStatement) {
      const src = imp.source || '';
      // Skip Mocha/Chai/Sinon/Jasmine-specific imports
      if (/^(chai|sinon|chai-as-promised|sinon-chai|jasmine)$/.test(src))
        continue;
      lines.push(imp.originalSource || src);
    }
  }

  if (lines.length > 0) lines.push('');

  // Emit body nodes
  for (const node of ir.body || []) {
    emitBodyNode(node, 0, lines);
  }

  return lines.join('\n') + '\n';
}

// ═══════════════════════════════════════════════════════════════════════
// Tree-walk emitters
// ═══════════════════════════════════════════════════════════════════════

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

/**
 * Emit a describe() block with hooks and tests.
 */
function emitSuite(suite, depth, lines) {
  const indent = '  '.repeat(depth);
  const name = escapeSingleQuotes(suite.name || 'unnamed');
  const mod = suiteModifier(suite);

  lines.push(`${indent}describe${mod}('${name}', () => {`);

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
 * Emit an it() block.
 */
function emitTestCase(tc, depth, lines) {
  const indent = '  '.repeat(depth);
  const name = escapeSingleQuotes(tc.name || 'unnamed');
  const mod = testModifier(tc);
  const asyncPrefix = tc.isAsync ? 'async ' : '';

  lines.push(`${indent}it${mod}('${name}', ${asyncPrefix}() => {`);

  for (const child of tc.body || []) {
    emitBodyNode(child, depth + 1, lines);
  }

  lines.push(`${indent}});`);
}

/**
 * Emit a hook block (beforeEach, afterAll, etc.).
 */
function emitHook(hook, depth, lines) {
  const indent = '  '.repeat(depth);
  const hookType = hook.hookType || 'beforeEach';
  const asyncPrefix = hook.isAsync ? 'async ' : '';

  lines.push(`${indent}${hookType}(${asyncPrefix}() => {`);

  for (const child of hook.body || []) {
    emitBodyNode(child, depth + 1, lines);
  }

  lines.push(`${indent}});`);
}

// ═══════════════════════════════════════════════════════════════════════
// Assertion emitter
// ═══════════════════════════════════════════════════════════════════════

/**
 * Emit a Jest assertion from an IR Assertion node.
 *
 * Maps IR assertion kinds to Jest expect matchers.
 *
 * @param {Assertion} node
 * @returns {{ code: string, supported: boolean }}
 */
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
  const handler = JEST_ASSERTIONS[kind];
  if (handler) {
    const code = handler(subject, expected, not);
    if (code) return { code, supported: true };
  }

  // Unsupported kind
  const original = node.originalSource ? node.originalSource.trim() : '';
  return {
    code:
      `// HAMLET-TODO: Unsupported assertion kind "${kind}"` +
      (original ? `\n// Original: ${original}` : ''),
    supported: false,
  };
}

// ═══════════════════════════════════════════════════════════════════════
// MockCall emitter
// ═══════════════════════════════════════════════════════════════════════

/**
 * Emit a Jest mock call from an IR MockCall node.
 *
 * @param {MockCall} node
 * @returns {{ code: string, supported: boolean }}
 */
export function emitMockCall(node) {
  if (!(node instanceof MockCall)) {
    return { code: '', supported: false };
  }

  const { kind, target, originalSource } = node;

  switch (kind) {
    case 'mockModule':
      if (target) {
        return { code: `jest.mock('${target}')`, supported: true };
      }
      return todoMock(node, 'jest.mock() without target');
    case 'createMock':
      return { code: 'jest.fn()', supported: true };
    case 'spyOnMethod':
      return todoMock(node, 'jest.spyOn() — needs object and method args');
    case 'fakeTimers':
      return { code: 'jest.useFakeTimers()', supported: true };
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
// Assertion kind → Jest matcher mapping
// ═══════════════════════════════════════════════════════════════════════

const JEST_ASSERTIONS = {
  strictEqual: (subj, exp, not) => `expect(${subj})${not}.toBe(${exp})`,
  deepEqual: (subj, exp, not) => `expect(${subj})${not}.toEqual(${exp})`,
  equal: (subj, exp, not) => `expect(${subj})${not}.toBe(${exp})`,
  truthy: (subj, _exp, not) => `expect(${subj})${not}.toBeTruthy()`,
  falsy: (subj, _exp, not) => `expect(${subj})${not}.toBeFalsy()`,
  isNull: (subj, _exp, not) => `expect(${subj})${not}.toBeNull()`,
  isUndefined: (subj, _exp, not) => `expect(${subj})${not}.toBeUndefined()`,
  isDefined: (subj, _exp, not) => `expect(${subj})${not}.toBeDefined()`,
  isNaN: (subj, _exp, not) => `expect(${subj})${not}.toBeNaN()`,
  instanceOf: (subj, exp, not) =>
    `expect(${subj})${not}.toBeInstanceOf(${exp})`,
  matches: (subj, exp, not) => `expect(${subj})${not}.toMatch(${exp})`,
  contains: (subj, exp, not) => `expect(${subj})${not}.toContain(${exp})`,
  containsEqual: (subj, exp, not) =>
    `expect(${subj})${not}.toContainEqual(${exp})`,
  hasLength: (subj, exp, not) => `expect(${subj})${not}.toHaveLength(${exp})`,
  hasProperty: (subj, exp, not) =>
    `expect(${subj})${not}.toHaveProperty(${exp})`,
  greaterThan: (subj, exp, not) =>
    `expect(${subj})${not}.toBeGreaterThan(${exp})`,
  lessThan: (subj, exp, not) => `expect(${subj})${not}.toBeLessThan(${exp})`,
  greaterOrEqual: (subj, exp, not) =>
    `expect(${subj})${not}.toBeGreaterThanOrEqual(${exp})`,
  lessOrEqual: (subj, exp, not) =>
    `expect(${subj})${not}.toBeLessThanOrEqual(${exp})`,
  closeTo: (subj, exp, not) => `expect(${subj})${not}.toBeCloseTo(${exp})`,
  throws: (subj, exp, not) => `expect(${subj})${not}.toThrow(${exp || ''})`,
  called: (subj, _exp, not) => `expect(${subj})${not}.toHaveBeenCalled()`,
  calledWith: (subj, exp, not) =>
    `expect(${subj})${not}.toHaveBeenCalledWith(${exp || ''})`,
  calledTimes: (subj, exp, not) =>
    `expect(${subj})${not}.toHaveBeenCalledTimes(${exp})`,
  snapshot: (subj, _exp, not) => `expect(${subj})${not}.toMatchSnapshot()`,
  resolves: (subj, exp, not) =>
    `await expect(${subj}).resolves${not}.toBe(${exp || 'undefined'})`,
  rejects: (subj, exp, not) =>
    `await expect(${subj}).rejects${not}.toThrow(${exp || ''})`,
  hasClass: (subj, exp, not) => `expect(${subj})${not}.toHaveClass(${exp})`,
  hasCount: (subj, exp, not) => `expect(${subj})${not}.toHaveCount(${exp})`,
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
      if (mod.modifierType === 'pending') return '.todo';
    }
  }
  return '';
}

function escapeSingleQuotes(str) {
  return str.replace(/'/g, "\\'");
}
