/**
 * IR-driven TestNG emitter.
 *
 * Generates TestNG code from IR nodes:
 * - Class-based structure with @Test methods
 * - @BeforeMethod/@AfterMethod/@BeforeClass/@AfterClass lifecycle
 * - Assert.assertEquals(actual, expected) with actual-first arg order
 * - import org.testng.* packages
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
    return /\bAssert\.\w+\s*\(/.test(line);
  }
  return false;
}

/**
 * Generate a complete TestNG file from an IR tree.
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
  imports.add('import org.testng.annotations.Test;');

  function walk(nodes) {
    for (const node of nodes || []) {
      if (node instanceof Hook) {
        const map = {
          beforeEach: 'BeforeMethod',
          afterEach: 'AfterMethod',
          beforeAll: 'BeforeClass',
          afterAll: 'AfterClass',
        };
        const name = map[node.hookType];
        if (name) imports.add(`import org.testng.annotations.${name};`);
        walk(node.body);
      }
      if (node instanceof TestSuite) {
        walk(node.hooks);
        walk(node.tests);
      }
      if (node instanceof TestCase) {
        walk(node.body);
      }
      if (node instanceof Assertion) {
        imports.add('import org.testng.Assert;');
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
      indent + '// HAMLET-TODO: Mock/stub conversion not supported for TestNG'
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

  lines.push(`${indent}public class ${name} {`);
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
    lines.push(`${indent}@Test(enabled = false)`);
  } else {
    lines.push(`${indent}@Test`);
  }
  lines.push(`${indent}public void ${name}() {`);

  for (const child of tc.body || []) {
    emitBodyNode(child, depth + 1, lines);
  }

  lines.push(`${indent}}`);
}

function emitHook(hook, depth, lines) {
  const indent = '    '.repeat(depth);
  const annotationMap = {
    beforeEach: '@BeforeMethod',
    afterEach: '@AfterMethod',
    beforeAll: '@BeforeClass',
    afterAll: '@AfterClass',
  };
  const annotation = annotationMap[hook.hookType] || '@BeforeMethod';
  const methodName = hook.hookType || 'setup';

  lines.push(`${indent}${annotation}`);
  lines.push(`${indent}public void ${methodName}() {`);

  for (const child of hook.body || []) {
    emitBodyNode(child, depth + 1, lines);
  }

  lines.push(`${indent}}`);
}

// ═══════════════════════════════════════════════════════════════════════
// Assertion emitter — TestNG Assert API
// ═══════════════════════════════════════════════════════════════════════

export function emitAssertion(node) {
  if (!(node instanceof Assertion)) {
    return { code: '', supported: false };
  }

  const { kind, subject, expected } = node;
  const handler = TESTNG_ASSERTIONS[kind];
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
// Assertion kind → TestNG Assert mapping
// TestNG arg order: Assert.assertEquals(actual, expected)
// ═══════════════════════════════════════════════════════════════════════

const TESTNG_ASSERTIONS = {
  equal: (subj, exp) =>
    subj && exp ? `Assert.assertEquals(${subj}, ${exp})` : null,
  notEqual: (subj, exp) =>
    subj && exp ? `Assert.assertNotEquals(${subj}, ${exp})` : null,
  strictEqual: (subj, exp) =>
    subj && exp ? `Assert.assertSame(${subj}, ${exp})` : null,
  truthy: (subj) => (subj ? `Assert.assertTrue(${subj})` : null),
  falsy: (subj) => (subj ? `Assert.assertFalse(${subj})` : null),
  isNull: (subj) => (subj ? `Assert.assertNull(${subj})` : null),
  isDefined: (subj) => (subj ? `Assert.assertNotNull(${subj})` : null),
  throws: (_subj, exp) =>
    exp ? `Assert.assertThrows(${exp}, () -> {})` : null,
  contains: (subj, exp) =>
    subj && exp ? `Assert.assertTrue(${subj}.contains(${exp}))` : null,
  greaterThan: (subj, exp) =>
    subj && exp ? `Assert.assertTrue(${subj} > ${exp})` : null,
  lessThan: (subj, exp) =>
    subj && exp ? `Assert.assertTrue(${subj} < ${exp})` : null,
  isInstance: (subj, exp) =>
    subj && exp ? `Assert.assertTrue(${subj} instanceof ${exp})` : null,
  fail: () => 'Assert.fail()',
};
