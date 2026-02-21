/**
 * IR-driven JUnit 5 emitter.
 *
 * Generates JUnit 5 (Jupiter) code from IR nodes:
 * - Class-based structure with @Test methods
 * - @BeforeEach/@AfterEach/@BeforeAll/@AfterAll lifecycle
 * - Assertions.assertEquals(expected, actual) with expected-first arg order
 * - import org.junit.jupiter.api.* packages
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
  return { code: '', supported: false };
}

export function matchesBaseline(line, node) {
  if (node instanceof Assertion) {
    return /\bAssertions\.\w+\s*\(/.test(line) || /\bassert\w+\s*\(/.test(line);
  }
  return false;
}

/**
 * Generate a complete JUnit 5 file from an IR tree.
 */
export function emitFullFile(ir) {
  if (!(ir instanceof TestFile)) return null;

  const lines = [];
  const neededImports = collectNeededImports(ir);

  // Emit imports
  for (const imp of neededImports) {
    lines.push(imp);
  }

  // Emit user imports — skip source framework imports
  for (const imp of ir.imports || []) {
    if (imp instanceof ImportStatement) {
      const src = imp.source || '';
      if (/org\.junit|org\.testng|org\.hamcrest/.test(src)) continue;
      lines.push(imp.originalSource || `import ${src};`);
    }
  }

  if (lines.length > 0) lines.push('');

  for (const node of ir.body || []) {
    emitBodyNode(node, 0, lines);
  }

  return lines.join('\n') + '\n';
}

// ═══════════════════════════════════════════════════════════════════════
// Import collection
// ═══════════════════════════════════════════════════════════════════════

function collectNeededImports(ir) {
  const imports = new Set();
  imports.add('import org.junit.jupiter.api.Test;');

  function walk(nodes) {
    for (const node of nodes || []) {
      if (node instanceof Hook) {
        const map = {
          beforeEach: 'BeforeEach',
          afterEach: 'AfterEach',
          beforeAll: 'BeforeAll',
          afterAll: 'AfterAll',
        };
        const name = map[node.hookType];
        if (name) imports.add(`import org.junit.jupiter.api.${name};`);
        walk(node.body);
      }
      if (node instanceof TestSuite) {
        walk(node.hooks);
        walk(node.tests);
      }
      if (node instanceof TestCase) {
        const hasSkip = (node.modifiers || []).some(
          (m) => m instanceof Modifier && m.modifierType === 'skip'
        );
        if (hasSkip) imports.add('import org.junit.jupiter.api.Disabled;');
        walk(node.body);
      }
      if (node instanceof Assertion) {
        imports.add('import org.junit.jupiter.api.Assertions;');
      }
    }
  }

  walk(ir.body || []);
  return [...imports].sort();
}

// ═══════════════════════════════════════════════════════════════════════
// Tree-walk emitters
// ═══════════════════════════════════════════════════════════════════════

function emitBodyNode(node, depth, lines) {
  const indent = '    '.repeat(depth);

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
    lines.push(
      indent + '// HAMLET-TODO: Mock/stub conversion not supported for JUnit 5'
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
  const indent = '    '.repeat(depth);
  const name = suite.name || 'TestSuite';

  lines.push(`${indent}class ${name} {`);
  lines.push('');

  // Hooks
  const hookOrder = ['beforeAll', 'beforeEach', 'afterEach', 'afterAll'];
  const sortedHooks = [...(suite.hooks || [])].sort((a, b) => {
    return hookOrder.indexOf(a.hookType) - hookOrder.indexOf(b.hookType);
  });
  for (const hook of sortedHooks) {
    emitHook(hook, depth + 1, lines);
    lines.push('');
  }

  // Tests
  for (let i = 0; i < (suite.tests || []).length; i++) {
    const child = suite.tests[i];
    emitBodyNode(child, depth + 1, lines);
    if (i < suite.tests.length - 1) {
      lines.push('');
    }
  }

  lines.push(`${indent}}`);
}

function emitTestCase(tc, depth, lines) {
  const indent = '    '.repeat(depth);
  const name = tc.name || 'testUnnamed';
  const hasSkip = (tc.modifiers || []).some(
    (m) => m instanceof Modifier && m.modifierType === 'skip'
  );

  if (hasSkip) {
    lines.push(`${indent}@Disabled`);
  }
  lines.push(`${indent}@Test`);
  lines.push(`${indent}void ${name}() {`);

  for (const child of tc.body || []) {
    emitBodyNode(child, depth + 1, lines);
  }

  lines.push(`${indent}}`);
}

function emitHook(hook, depth, lines) {
  const indent = '    '.repeat(depth);
  const annotationMap = {
    beforeEach: '@BeforeEach',
    afterEach: '@AfterEach',
    beforeAll: '@BeforeAll',
    afterAll: '@AfterAll',
  };
  const annotation = annotationMap[hook.hookType] || '@BeforeEach';
  const isStatic =
    hook.hookType === 'beforeAll' || hook.hookType === 'afterAll';
  const staticMod = isStatic ? 'static ' : '';
  const methodName = hook.hookType || 'setup';

  lines.push(`${indent}${annotation}`);
  lines.push(`${indent}${staticMod}void ${methodName}() {`);

  for (const child of hook.body || []) {
    emitBodyNode(child, depth + 1, lines);
  }

  lines.push(`${indent}}`);
}

// ═══════════════════════════════════════════════════════════════════════
// Assertion emitter — JUnit 5 Assertions API
// ═══════════════════════════════════════════════════════════════════════

export function emitAssertion(node) {
  if (!(node instanceof Assertion)) {
    return { code: '', supported: false };
  }

  const { kind, subject, expected } = node;
  const handler = JUNIT5_ASSERTIONS[kind];
  if (handler) {
    const code = handler(subject, expected);
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
// Assertion kind → JUnit 5 Assertions mapping
// JUnit 5 arg order: assertEquals(expected, actual)
// ═══════════════════════════════════════════════════════════════════════

const JUNIT5_ASSERTIONS = {
  equal: (subj, exp) =>
    subj && exp ? `Assertions.assertEquals(${exp}, ${subj})` : null,
  notEqual: (subj, exp) =>
    subj && exp ? `Assertions.assertNotEquals(${exp}, ${subj})` : null,
  strictEqual: (subj, exp) =>
    subj && exp ? `Assertions.assertSame(${exp}, ${subj})` : null,
  deepEqual: (subj, exp) =>
    subj && exp ? `Assertions.assertArrayEquals(${exp}, ${subj})` : null,
  truthy: (subj) => (subj ? `Assertions.assertTrue(${subj})` : null),
  falsy: (subj) => (subj ? `Assertions.assertFalse(${subj})` : null),
  isNull: (subj) => (subj ? `Assertions.assertNull(${subj})` : null),
  isDefined: (subj) => (subj ? `Assertions.assertNotNull(${subj})` : null),
  throws: (_subj, exp) =>
    exp ? `Assertions.assertThrows(${exp}, () -> {})` : null,
  contains: (subj, exp) =>
    subj && exp ? `Assertions.assertTrue(${subj}.contains(${exp}))` : null,
  greaterThan: (subj, exp) =>
    subj && exp ? `Assertions.assertTrue(${subj} > ${exp})` : null,
  greaterThanOrEqual: (subj, exp) =>
    subj && exp ? `Assertions.assertTrue(${subj} >= ${exp})` : null,
  lessThan: (subj, exp) =>
    subj && exp ? `Assertions.assertTrue(${subj} < ${exp})` : null,
  lessThanOrEqual: (subj, exp) =>
    subj && exp ? `Assertions.assertTrue(${subj} <= ${exp})` : null,
  isInstance: (subj, exp) =>
    subj && exp ? `Assertions.assertInstanceOf(${exp}, ${subj})` : null,
  fail: () => 'Assertions.fail()',
};
