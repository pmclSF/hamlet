/**
 * IR-driven unittest emitter.
 *
 * Generates Python unittest code from IR nodes:
 * - Class-based structure inheriting unittest.TestCase
 * - setUp/tearDown/setUpClass/tearDownClass lifecycle
 * - self.assertEqual(actual, expected) assertion style
 * - import unittest
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
    return /\bself\.assert\w+\s*\(/.test(line);
  }
  return false;
}

/**
 * Generate a complete unittest file from an IR tree.
 */
export function emitFullFile(ir) {
  if (!(ir instanceof TestFile)) return null;

  const lines = [];
  lines.push('import unittest');

  // Emit user imports — skip source framework imports
  for (const imp of ir.imports || []) {
    if (imp instanceof ImportStatement) {
      const src = imp.source || '';
      if (/^(unittest|pytest|nose|nose2|nose\.tools|nose2\.tools)$/.test(src))
        continue;
      lines.push(imp.originalSource || `import ${src}`);
    }
  }

  if (lines.length > 0) lines.push('');
  lines.push('');

  for (const node of ir.body || []) {
    emitBodyNode(node, 0, lines);
  }

  // Add main block
  lines.push('');
  lines.push('');
  lines.push("if __name__ == '__main__':");
  lines.push('    unittest.main()');

  // Clean up multiple blank lines
  let result = lines.join('\n') + '\n';
  result = result.replace(/\n{4,}/g, '\n\n\n');
  result = result.replace(/^\n+/, '');
  return result;
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
    lines.push(indent + result.code);
  } else if (node instanceof Comment) {
    lines.push(indent + (node.text || '').trim());
  } else if (node instanceof RawCode) {
    const code = (node.code || '').trim();
    if (code) lines.push(indent + code);
  } else if (node instanceof MockCall) {
    lines.push(
      indent + '# HAMLET-TODO: Mock/stub conversion not supported for unittest'
    );
  } else if (node instanceof SharedVariable) {
    lines.push(
      indent +
        `# HAMLET-TODO: Shared variable "${node.name}" conversion not yet supported`
    );
  } else {
    const original = node.originalSource
      ? node.originalSource.trim()
      : node.type || 'unknown';
    lines.push(indent + `# HAMLET-TODO: Unsupported node (${original})`);
  }
}

function emitSuite(suite, depth, lines) {
  const indent = '    '.repeat(depth);
  const name = suite.name || 'TestSuite';

  lines.push(`${indent}class ${name}(unittest.TestCase):`);
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
  const tests = suite.tests || [];
  if (tests.length === 0 && sortedHooks.length === 0) {
    lines.push(`${indent}    pass`);
  }
  for (let i = 0; i < tests.length; i++) {
    const child = tests[i];
    emitBodyNode(child, depth + 1, lines);
    if (i < tests.length - 1) {
      lines.push('');
    }
  }
}

function emitTestCase(tc, depth, lines) {
  const indent = '    '.repeat(depth);
  const name = tc.name || 'test_unnamed';
  const hasSkip = (tc.modifiers || []).some(
    (m) => m instanceof Modifier && m.modifierType === 'skip'
  );

  if (hasSkip) {
    lines.push(`${indent}@unittest.skip`);
  }
  lines.push(`${indent}def ${name}(self):`);

  const body = tc.body || [];
  if (body.length === 0) {
    lines.push(`${indent}    pass`);
  }
  for (const child of body) {
    emitBodyNode(child, depth + 1, lines);
  }
}

function emitHook(hook, depth, lines) {
  const indent = '    '.repeat(depth);
  const nameMap = {
    beforeEach: 'setUp',
    afterEach: 'tearDown',
    beforeAll: 'setUpClass',
    afterAll: 'tearDownClass',
  };
  const methodName = nameMap[hook.hookType] || 'setUp';
  const isClassLevel =
    hook.hookType === 'beforeAll' || hook.hookType === 'afterAll';

  if (isClassLevel) {
    lines.push(`${indent}@classmethod`);
    lines.push(`${indent}def ${methodName}(cls):`);
  } else {
    lines.push(`${indent}def ${methodName}(self):`);
  }

  const body = hook.body || [];
  if (body.length === 0) {
    lines.push(`${indent}    pass`);
  }
  for (const child of body) {
    emitBodyNode(child, depth + 1, lines);
  }
}

// ═══════════════════════════════════════════════════════════════════════
// Assertion emitter — unittest self.assert* style
// ═══════════════════════════════════════════════════════════════════════

export function emitAssertion(node) {
  if (!(node instanceof Assertion)) {
    return { code: '', supported: false };
  }

  const { kind, subject, expected } = node;
  const handler = UNITTEST_ASSERTIONS[kind];
  if (handler) {
    const code = handler(subject, expected);
    if (code) return { code, supported: true };
  }

  const original = node.originalSource ? node.originalSource.trim() : '';
  return {
    code:
      `# HAMLET-TODO: Unsupported assertion kind "${kind}"` +
      (original ? `\n# Original: ${original}` : ''),
    supported: false,
  };
}

// ═══════════════════════════════════════════════════════════════════════
// Assertion kind → unittest self.assert* mapping
// ═══════════════════════════════════════════════════════════════════════

const UNITTEST_ASSERTIONS = {
  equal: (subj, exp) =>
    subj && exp ? `self.assertEqual(${subj}, ${exp})` : null,
  notEqual: (subj, exp) =>
    subj && exp ? `self.assertNotEqual(${subj}, ${exp})` : null,
  strictEqual: (subj, exp) =>
    subj && exp ? `self.assertIs(${subj}, ${exp})` : null,
  deepEqual: (subj, exp) =>
    subj && exp ? `self.assertEqual(${subj}, ${exp})` : null,
  truthy: (subj) => (subj ? `self.assertTrue(${subj})` : null),
  falsy: (subj) => (subj ? `self.assertFalse(${subj})` : null),
  isNull: (subj) => (subj ? `self.assertIsNone(${subj})` : null),
  isDefined: (subj) => (subj ? `self.assertIsNotNone(${subj})` : null),
  throws: (_subj, exp) =>
    exp
      ? `with self.assertRaises(${exp}):`
      : `with self.assertRaises(Exception):`,
  contains: (subj, exp) =>
    subj && exp ? `self.assertIn(${exp}, ${subj})` : null,
  notContains: (subj, exp) =>
    subj && exp ? `self.assertNotIn(${exp}, ${subj})` : null,
  greaterThan: (subj, exp) =>
    subj && exp ? `self.assertGreater(${subj}, ${exp})` : null,
  greaterThanOrEqual: (subj, exp) =>
    subj && exp ? `self.assertGreaterEqual(${subj}, ${exp})` : null,
  lessThan: (subj, exp) =>
    subj && exp ? `self.assertLess(${subj}, ${exp})` : null,
  lessThanOrEqual: (subj, exp) =>
    subj && exp ? `self.assertLessEqual(${subj}, ${exp})` : null,
  isInstance: (subj, exp) =>
    subj && exp ? `self.assertIsInstance(${subj}, ${exp})` : null,
  fail: () => 'self.fail()',
};
