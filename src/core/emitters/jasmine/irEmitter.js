/**
 * IR-driven Jasmine emitter.
 *
 * Generates Jasmine code from IR nodes. Jasmine uses its own spy API
 * (spyOn, jasmine.createSpy) and expect matchers similar to Jest.
 *
 * Provides three unified exports:
 *   emitNode(node)             — dispatch by type → { code, supported }
 *   matchesBaseline(line, node) — check if a baseline line corresponds to a node
 *   emitFullFile(ir)           — generate complete Jasmine file from IR tree walk
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
    return /\bjasmine\.\w+/.test(line) || /\bspyOn\s*\(/.test(line);
  }
  return false;
}

/**
 * Generate a complete Jasmine file from an IR tree.
 */
export function emitFullFile(ir) {
  if (!(ir instanceof TestFile)) return null;

  const lines = [];

  // Emit user imports — skip source framework imports
  for (const imp of ir.imports || []) {
    if (imp instanceof ImportStatement) {
      const src = imp.source || '';
      if (
        /^(@jest\/globals|vitest|chai|sinon|chai-as-promised|sinon-chai)$/.test(
          src
        )
      )
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

  // Jasmine uses fdescribe/xdescribe for modifiers
  const keyword =
    mod === '.only' ? 'fdescribe' : mod === '.skip' ? 'xdescribe' : 'describe';
  lines.push(`${indent}${keyword}('${name}', function () {`);

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

  // Jasmine uses fit/xit for modifiers
  const keyword = mod === '.only' ? 'fit' : mod === '.skip' ? 'xit' : 'it';
  const asyncPrefix = tc.isAsync ? 'async ' : '';

  lines.push(`${indent}${keyword}('${name}', ${asyncPrefix}function () {`);

  for (const child of tc.body || []) {
    emitBodyNode(child, depth + 1, lines);
  }

  lines.push(`${indent}});`);
}

function emitHook(hook, depth, lines) {
  const indent = '  '.repeat(depth);
  const hookType = hook.hookType || 'beforeEach';
  const asyncPrefix = hook.isAsync ? 'async ' : '';

  lines.push(`${indent}${hookType}(${asyncPrefix}function () {`);

  for (const child of hook.body || []) {
    emitBodyNode(child, depth + 1, lines);
  }

  lines.push(`${indent}});`);
}

// ═══════════════════════════════════════════════════════════════════════
// Assertion emitter — Jasmine expect matchers
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
  const handler = JASMINE_ASSERTIONS[kind];
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
// MockCall emitter — Jasmine spy API
// ═══════════════════════════════════════════════════════════════════════

export function emitMockCall(node) {
  if (!(node instanceof MockCall)) {
    return { code: '', supported: false };
  }

  const { kind } = node;

  switch (kind) {
    case 'createMock':
      return { code: 'jasmine.createSpy()', supported: true };
    case 'spyOnMethod':
      return todoMock(node, 'spyOn() — needs object and method args');
    case 'fakeTimers':
      return { code: 'jasmine.clock().install()', supported: true };
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
// Assertion kind → Jasmine matcher mapping
// ═══════════════════════════════════════════════════════════════════════

const JASMINE_ASSERTIONS = {
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
  containsEqual: (subj, exp, not) => `expect(${subj})${not}.toContain(${exp})`,
  hasLength: (subj, exp, not) => `expect(${subj}.length)${not}.toBe(${exp})`,
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
  throws: (subj, _exp, not) => `expect(${subj})${not}.toThrow()`,
  called: (subj, _exp, not) => `expect(${subj})${not}.toHaveBeenCalled()`,
  calledWith: (subj, exp, not) =>
    `expect(${subj})${not}.toHaveBeenCalledWith(${exp || ''})`,
  calledTimes: (subj, exp, not) =>
    `expect(${subj}).calls.count()${not === '.not' ? ' !== ' : ' === '}${exp}`,
  snapshot: (subj, _exp, _not) =>
    `// HAMLET-TODO: Jasmine has no built-in snapshot support`,
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
