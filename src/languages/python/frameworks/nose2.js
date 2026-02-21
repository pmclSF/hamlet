/**
 * nose2 framework definition.
 *
 * Provides detect, parse, and emit for the nose2 testing framework.
 * parse() builds an IR tree from nose2 source code for scoring.
 * emit() is a stub — nose2 is only used as a source framework.
 */

import {
  TestFile,
  TestSuite,
  TestCase,
  Hook,
  Assertion,
  ImportStatement,
  RawCode,
  Comment,
  Modifier,
} from '../../../core/ir.js';

/**
 * Detect whether source code is nose2.
 * Returns confidence score 0-100.
 */
function detect(source) {
  if (!source || !source.trim()) return 0;

  let score = 0;

  // Strong nose signals
  if (/from\s+nose\.tools\s+import\b/.test(source)) score += 30;
  if (/from\s+nose2\.tools\s+import\b/.test(source)) score += 30;
  if (/import\s+nose2\b/.test(source)) score += 30;
  if (/import\s+nose\b/.test(source)) score += 25;

  // nose-specific assertion functions
  if (/\bassert_equal\s*\(/.test(source)) score += 15;
  if (/\bassert_true\s*\(/.test(source)) score += 15;
  if (/\bassert_false\s*\(/.test(source)) score += 15;
  if (/\bassert_raises\s*\(/.test(source)) score += 15;
  if (/\bassert_in\s*\(/.test(source)) score += 15;
  if (/\bassert_not_equal\s*\(/.test(source)) score += 15;
  if (/\bassert_is_none\s*\(/.test(source)) score += 15;
  if (/\bassert_is_not_none\s*\(/.test(source)) score += 15;
  if (/\bassert_is_instance\s*\(/.test(source)) score += 15;

  // nose2-specific decorators
  if (/@params\s*\(/.test(source)) score += 15;
  if (/@attr\s*\(/.test(source)) score += 10;
  if (/@such\.it\b/.test(source)) score += 15;

  // Weak signals (shared with other frameworks)
  if (/def\s+test_\w+/.test(source)) score += 5;
  if (/class\s+\w+.*TestCase/.test(source)) score += 5;

  // Negative signals: NOT nose
  if (/import\s+pytest\b/.test(source)) score -= 30;
  if (/@pytest\./.test(source)) score -= 30;
  if (/import\s+unittest\b/.test(source) && !/import\s+nose/.test(source))
    score -= 20;

  return Math.max(0, Math.min(100, score));
}

/**
 * Parse nose2 source code into a nested IR tree.
 *
 * Uses indent-level tracking to nest TestCase/Hook inside TestSuite (class),
 * and Assertion/RawCode inside TestCase/Hook bodies.
 * Extracts assertion subject/expected from assert_equal(a, b).
 */
function parse(source) {
  const lines = source.split('\n');
  const imports = [];
  const rootBody = [];
  const stack = [{ node: null, addChild: (c) => rootBody.push(c), indent: -1 }];
  let pendingDecorators = [];

  function addChild(child) {
    const top = stack[stack.length - 1];
    const node = top.node;
    if (!node) {
      top.addChild(child);
    } else if (node instanceof TestSuite) {
      if (child instanceof Hook) node.hooks.push(child);
      else if (child instanceof TestCase) node.tests.push(child);
      else node.tests.push(child);
    } else if (node instanceof TestCase || node instanceof Hook) {
      node.body.push(child);
    }
  }

  function getIndent(line) {
    const m = line.match(/^(\s*)/);
    return m ? m[1].length : 0;
  }

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();
    const loc = { line: i + 1, column: 0 };
    const indent = getIndent(line);

    if (!trimmed) continue;

    // Pop stack when indent decreases
    while (stack.length > 1 && indent <= stack[stack.length - 1].indent) {
      stack.pop();
    }

    // Comments
    if (trimmed.startsWith('#')) {
      const isLicense =
        /license|copyright|MIT|Apache|BSD/i.test(trimmed) && i < 5;
      addChild(
        new Comment({
          text: line,
          commentKind: isLicense ? 'license' : 'inline',
          preserveExact: isLicense,
          sourceLocation: loc,
          originalSource: line,
        })
      );
      continue;
    }

    // Import statements
    if (/^(?:import|from)\s/.test(trimmed)) {
      const sourceMatch = trimmed.match(
        /(?:from\s+(\S+)\s+import|import\s+(\S+))/
      );
      const imp = new ImportStatement({
        kind: 'library',
        source: sourceMatch ? sourceMatch[1] || sourceMatch[2] : '',
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      });
      imports.push(imp);
      continue;
    }

    // Class declaration → TestSuite
    if (/^\s*class\s+\w+/.test(line)) {
      const suite = new TestSuite({
        name: (trimmed.match(/class\s+(\w+)/) || [])[1] || '',
        modifiers: [],
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      });
      addChild(suite);
      pendingDecorators = [];
      stack.push({
        node: suite,
        addChild: (c) => {
          if (c instanceof Hook) suite.hooks.push(c);
          else suite.tests.push(c);
        },
        indent,
      });
      continue;
    }

    // setUp / tearDown → Hook
    if (/def\s+setUp\s*\(/.test(trimmed)) {
      const hook = new Hook({
        hookType: 'beforeEach',
        body: [],
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      });
      addChild(hook);
      stack.push({ node: hook, addChild: (c) => hook.body.push(c), indent });
      continue;
    }
    if (/def\s+tearDown\s*\(/.test(trimmed)) {
      const hook = new Hook({
        hookType: 'afterEach',
        body: [],
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      });
      addChild(hook);
      stack.push({ node: hook, addChild: (c) => hook.body.push(c), indent });
      continue;
    }
    if (/def\s+setUpClass\s*\(/.test(trimmed)) {
      const hook = new Hook({
        hookType: 'beforeAll',
        body: [],
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      });
      addChild(hook);
      stack.push({ node: hook, addChild: (c) => hook.body.push(c), indent });
      continue;
    }
    if (/def\s+tearDownClass\s*\(/.test(trimmed)) {
      const hook = new Hook({
        hookType: 'afterAll',
        body: [],
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      });
      addChild(hook);
      stack.push({ node: hook, addChild: (c) => hook.body.push(c), indent });
      continue;
    }

    // Decorators — store as pending
    if (/@params\s*\(/.test(trimmed)) {
      pendingDecorators.push({
        type: 'parameterized',
        loc,
        originalSource: line,
      });
      continue;
    }
    if (/@attr\s*\(/.test(trimmed)) {
      pendingDecorators.push({ type: 'tag', loc, originalSource: line });
      continue;
    }

    // Test functions/methods
    if (/def\s+test_\w+\s*\(/.test(trimmed)) {
      const tc = new TestCase({
        name: (trimmed.match(/def\s+(test_\w+)\s*\(/) || [])[1] || '',
        isAsync: /async\s+def/.test(trimmed),
        modifiers: [],
        body: [],
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      });
      addChild(tc);
      pendingDecorators = [];
      stack.push({ node: tc, addChild: (c) => tc.body.push(c), indent });
      continue;
    }

    // nose assertion functions — extract subject/expected
    if (/\bassert_\w+\s*\(/.test(trimmed)) {
      const assertion = parseNoseAssertion(trimmed, loc, line);
      addChild(assertion);
      continue;
    }

    // Everything else
    addChild(
      new RawCode({
        code: line,
        sourceLocation: loc,
        originalSource: line,
      })
    );
  }

  return new TestFile({
    language: 'python',
    imports,
    body: rootBody,
  });
}

/**
 * Parse a nose assertion and extract kind, subject, expected.
 * assert_equal(a, b) → kind='equal', subject=a, expected=b
 */
function parseNoseAssertion(trimmed, loc, line) {
  let kind = 'equal';
  if (/assert_not_equal/.test(trimmed)) kind = 'notEqual';
  else if (/assert_equal/.test(trimmed)) kind = 'equal';
  else if (/assert_true/.test(trimmed)) kind = 'truthy';
  else if (/assert_false/.test(trimmed)) kind = 'falsy';
  else if (/assert_is_not_none/.test(trimmed)) kind = 'isDefined';
  else if (/assert_is_none/.test(trimmed)) kind = 'isNull';
  else if (/assert_not_in/.test(trimmed)) kind = 'notContains';
  else if (/assert_in/.test(trimmed)) kind = 'contains';
  else if (/assert_raises/.test(trimmed)) kind = 'throws';
  else if (/assert_is_instance/.test(trimmed)) kind = 'isInstance';

  const args = extractNoseAssertArgs(trimmed);
  let subject, expected;

  if (args) {
    if (
      kind === 'equal' ||
      kind === 'notEqual' ||
      kind === 'contains' ||
      kind === 'notContains'
    ) {
      if (args.length >= 2) {
        subject = args[0];
        expected = args[1];
      }
    } else if (
      kind === 'truthy' ||
      kind === 'falsy' ||
      kind === 'isNull' ||
      kind === 'isDefined'
    ) {
      subject = args[0];
    } else if (kind === 'throws') {
      expected = args[0];
    } else if (kind === 'isInstance' && args.length >= 2) {
      subject = args[0];
      expected = args[1];
    }
  }

  return new Assertion({
    kind,
    subject,
    expected,
    sourceLocation: loc,
    originalSource: line,
    confidence: 'converted',
  });
}

/**
 * Extract arguments from a nose assert_X(...) call.
 */
function extractNoseAssertArgs(trimmed) {
  const m = trimmed.match(/\bassert_\w+\s*\(/);
  if (!m) return null;
  const start = m.index + m[0].length;
  let depth = 1;
  let i = start;
  while (i < trimmed.length && depth > 0) {
    const ch = trimmed[i];
    if (ch === '(') depth++;
    else if (ch === ')') depth--;
    i++;
  }
  const argsStr = trimmed.slice(start, i - 1);
  if (!argsStr.trim()) return [];

  // Simple comma split respecting parens/strings
  const args = [];
  let d = 0;
  let current = '';
  let inString = null;
  for (let j = 0; j < argsStr.length; j++) {
    const ch = argsStr[j];
    const prev = j > 0 ? argsStr[j - 1] : '';
    if (inString) {
      current += ch;
      if (ch === inString && prev !== '\\') inString = null;
      continue;
    }
    if (ch === '"' || ch === "'") {
      inString = ch;
      current += ch;
      continue;
    }
    if (ch === '(' || ch === '[' || ch === '{') {
      d++;
      current += ch;
    } else if (ch === ')' || ch === ']' || ch === '}') {
      d--;
      current += ch;
    } else if (ch === ',' && d === 0) {
      args.push(current.trim());
      current = '';
    } else {
      current += ch;
    }
  }
  if (current.trim()) args.push(current.trim());
  return args;
}

/**
 * Emit nose2 code from IR + original source.
 *
 * Stub — nose2 is only used as a source framework.
 */
function emit(_ir, source) {
  return source;
}

export default {
  name: 'nose2',
  language: 'python',
  paradigm: 'xunit',
  detect,
  parse,
  emit,
  imports: {
    packages: ['nose2', 'nose.tools'],
  },
};
