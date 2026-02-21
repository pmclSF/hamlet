/**
 * IR-driven pytest emitter.
 *
 * Generates pytest code from IR nodes:
 * - Function-based structure with bare assert
 * - @pytest.fixture for hooks, @pytest.mark.skip for skips
 * - assert actual == expected (bare assert style)
 * - import pytest
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
    return /^\s*assert\s+/.test(line) || /\bpytest\.raises\b/.test(line);
  }
  return false;
}

/**
 * Generate a complete pytest file from an IR tree.
 */
export function emitFullFile(ir) {
  if (!(ir instanceof TestFile)) return null;

  const lines = [];
  const needsPytest = checkNeedsPytest(ir);

  // Emit import pytest if needed
  if (needsPytest) {
    lines.push('import pytest');
  }

  // Emit user imports — skip source framework imports
  for (const imp of ir.imports || []) {
    if (imp instanceof ImportStatement) {
      const src = imp.source || '';
      if (/^(unittest|nose|nose2|nose\.tools|nose2\.tools)$/.test(src))
        continue;
      if (/^pytest$/.test(src)) continue;
      lines.push(imp.originalSource || `import ${src}`);
    }
  }

  if (lines.length > 0) lines.push('');
  lines.push('');

  for (const node of ir.body || []) {
    emitBodyNode(node, 0, lines);
  }

  // Clean up multiple blank lines
  let result = lines.join('\n') + '\n';
  result = result.replace(/\n{4,}/g, '\n\n\n');
  result = result.replace(/^\n+/, '');
  return result;
}

// ═══════════════════════════════════════════════════════════════════════
// Check if pytest import is needed
// ═══════════════════════════════════════════════════════════════════════

function checkNeedsPytest(ir) {
  let needed = false;

  function walk(nodes) {
    for (const node of nodes || []) {
      if (node instanceof Hook) {
        needed = true;
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
        if (hasSkip) needed = true;
        walk(node.body);
      }
      if (node instanceof Assertion && node.kind === 'throws') {
        needed = true;
      }
    }
  }

  walk(ir.body || []);
  return needed;
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
      indent + '# HAMLET-TODO: Mock/stub conversion not supported for pytest'
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

  lines.push(`${indent}class ${name}:`);

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
    lines.push(`${indent}@pytest.mark.skip`);
  }
  lines.push(`${indent}def ${name}():`);

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

  if (hook.hookType === 'beforeEach' || hook.hookType === 'afterEach') {
    // Use @pytest.fixture(autouse=True)
    lines.push(`${indent}@pytest.fixture(autouse=True)`);
    const name =
      hook.hookType === 'beforeEach' ? 'setup_method' : 'teardown_method';

    if (hook.hookType === 'afterEach') {
      lines.push(`${indent}def ${name}(self):`);
      lines.push(`${indent}    yield`);
      for (const child of hook.body || []) {
        emitBodyNode(child, depth + 1, lines);
      }
    } else {
      lines.push(`${indent}def ${name}(self):`);
      const body = hook.body || [];
      if (body.length === 0) {
        lines.push(`${indent}    pass`);
      }
      for (const child of body) {
        emitBodyNode(child, depth + 1, lines);
      }
    }
  } else if (hook.hookType === 'beforeAll' || hook.hookType === 'afterAll') {
    // Use @pytest.fixture(scope="class", autouse=True) or setup_class/teardown_class
    const name =
      hook.hookType === 'beforeAll' ? 'setup_class' : 'teardown_class';
    lines.push(`${indent}@classmethod`);

    if (hook.hookType === 'afterAll') {
      lines.push(`${indent}def ${name}(cls):`);
      lines.push(`${indent}    yield`);
      for (const child of hook.body || []) {
        emitBodyNode(child, depth + 1, lines);
      }
    } else {
      lines.push(`${indent}def ${name}(cls):`);
      const body = hook.body || [];
      if (body.length === 0) {
        lines.push(`${indent}    pass`);
      }
      for (const child of body) {
        emitBodyNode(child, depth + 1, lines);
      }
    }
  }
}

// ═══════════════════════════════════════════════════════════════════════
// Assertion emitter — pytest bare assert style
// ═══════════════════════════════════════════════════════════════════════

export function emitAssertion(node) {
  if (!(node instanceof Assertion)) {
    return { code: '', supported: false };
  }

  const { kind, subject, expected } = node;
  const handler = PYTEST_ASSERTIONS[kind];
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
// Assertion kind → pytest bare assert mapping
// ═══════════════════════════════════════════════════════════════════════

const PYTEST_ASSERTIONS = {
  equal: (subj, exp) => (subj && exp ? `assert ${subj} == ${exp}` : null),
  notEqual: (subj, exp) => (subj && exp ? `assert ${subj} != ${exp}` : null),
  strictEqual: (subj, exp) => (subj && exp ? `assert ${subj} is ${exp}` : null),
  deepEqual: (subj, exp) => (subj && exp ? `assert ${subj} == ${exp}` : null),
  truthy: (subj) => (subj ? `assert ${subj}` : null),
  falsy: (subj) => (subj ? `assert not ${subj}` : null),
  isNull: (subj) => (subj ? `assert ${subj} is None` : null),
  isDefined: (subj) => (subj ? `assert ${subj} is not None` : null),
  throws: (_subj, exp) =>
    exp ? `with pytest.raises(${exp}):` : `with pytest.raises(Exception):`,
  contains: (subj, exp) => (subj && exp ? `assert ${exp} in ${subj}` : null),
  notContains: (subj, exp) =>
    subj && exp ? `assert ${exp} not in ${subj}` : null,
  greaterThan: (subj, exp) => (subj && exp ? `assert ${subj} > ${exp}` : null),
  greaterThanOrEqual: (subj, exp) =>
    subj && exp ? `assert ${subj} >= ${exp}` : null,
  lessThan: (subj, exp) => (subj && exp ? `assert ${subj} < ${exp}` : null),
  lessThanOrEqual: (subj, exp) =>
    subj && exp ? `assert ${subj} <= ${exp}` : null,
  isInstance: (subj, exp) =>
    subj && exp ? `assert isinstance(${subj}, ${exp})` : null,
  fail: () => 'pytest.fail()',
};
