/**
 * IR-driven Vitest emitter.
 *
 * Generates Vitest code from IR nodes. Vitest is API-compatible with Jest
 * but uses vi.* for mocking and imports from 'vitest'.
 *
 * Provides three unified exports:
 *   emitNode(node)             — dispatch by type → { code, supported }
 *   matchesBaseline(line, node) — check if a baseline line corresponds to a node
 *   emitFullFile(ir)           — generate complete Vitest file from IR tree walk
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
 * Emit a single Vitest code fragment from any supported IR node.
 */
export function emitNode(node) {
  if (node instanceof Assertion) return emitAssertion(node);
  if (node instanceof MockCall) return emitMockCall(node);
  return { code: '', supported: false };
}

/**
 * Check if a baseline line corresponds to a given IR node.
 */
export function matchesBaseline(line, node) {
  if (node instanceof Assertion) {
    return /\bexpect\s*\(/.test(line);
  }
  if (node instanceof MockCall) {
    return /\bvi\.\w+/.test(line);
  }
  return false;
}

/**
 * Generate a complete Vitest file from an IR tree.
 */
export function emitFullFile(ir) {
  if (!(ir instanceof TestFile)) return null;

  const lines = [];

  // Collect needed Vitest imports
  const vitestImports = new Set(['describe', 'it', 'expect']);
  if (needsVi(ir)) vitestImports.add('vi');
  if (needsHooks(ir, 'beforeEach')) vitestImports.add('beforeEach');
  if (needsHooks(ir, 'afterEach')) vitestImports.add('afterEach');
  if (needsHooks(ir, 'beforeAll')) vitestImports.add('beforeAll');
  if (needsHooks(ir, 'afterAll')) vitestImports.add('afterAll');

  lines.push(`import { ${[...vitestImports].join(', ')} } from 'vitest';`);

  // Emit user imports — skip source framework imports
  for (const imp of ir.imports || []) {
    if (imp instanceof ImportStatement) {
      const src = imp.source || '';
      // Skip Jest/Mocha/Chai/Sinon/Jasmine-specific imports
      if (
        /^(chai|sinon|chai-as-promised|sinon-chai|jasmine|@jest\/globals)$/.test(
          src
        )
      )
        continue;
      lines.push(imp.originalSource || src);
    }
  }

  lines.push('');

  // Emit body nodes
  for (const node of ir.body || []) {
    emitBodyNode(node, 0, lines);
  }

  return lines.join('\n') + '\n';
}

// ═══════════════════════════════════════════════════════════════════════
// IR tree helpers
// ═══════════════════════════════════════════════════════════════════════

function needsVi(ir) {
  return hasMockNodes(ir.body || []);
}

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

function needsHooks(ir, hookType) {
  return hasHookType(ir.body || [], hookType);
}

function hasHookType(nodes, hookType) {
  for (const node of nodes) {
    if (node instanceof Hook && node.hookType === hookType) return true;
    if (node instanceof TestSuite) {
      if (hasHookType(node.hooks || [], hookType)) return true;
      if (hasHookType(node.tests || [], hookType)) return true;
    }
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
  const asyncPrefix = tc.isAsync ? 'async ' : '';

  lines.push(`${indent}it${mod}('${name}', ${asyncPrefix}() => {`);

  for (const child of tc.body || []) {
    emitBodyNode(child, depth + 1, lines);
  }

  lines.push(`${indent}});`);
}

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
// Assertion emitter — same matchers as Jest
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
  const handler = VITEST_ASSERTIONS[kind];
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
// MockCall emitter — uses vi.* instead of jest.*
// ═══════════════════════════════════════════════════════════════════════

export function emitMockCall(node) {
  if (!(node instanceof MockCall)) {
    return { code: '', supported: false };
  }

  const { kind, target } = node;

  switch (kind) {
    case 'mockModule':
      if (target) {
        return { code: `vi.mock('${target}')`, supported: true };
      }
      return todoMock(node, 'vi.mock() without target');
    case 'createMock':
      return { code: 'vi.fn()', supported: true };
    case 'spyOnMethod':
      return todoMock(node, 'vi.spyOn() — needs object and method args');
    case 'fakeTimers':
      return { code: 'vi.useFakeTimers()', supported: true };
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
// Assertion kind → Vitest matcher mapping (identical to Jest)
// ═══════════════════════════════════════════════════════════════════════

const VITEST_ASSERTIONS = {
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
