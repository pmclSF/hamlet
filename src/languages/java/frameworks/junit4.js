/**
 * JUnit 4 framework definition.
 *
 * Provides detect, parse, and emit for the JUnit 4 testing framework.
 * parse() builds an IR tree from JUnit 4 source code for scoring.
 * emit() is a stub — JUnit 4 is only used as a source framework.
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
 * Detect whether source code is JUnit 4.
 * Returns confidence score 0-100.
 */
function detect(source) {
  if (!source || !source.trim()) return 0;

  let score = 0;

  // Strong JUnit 4 signals
  if (/import\s+org\.junit\.Test\b/.test(source)) score += 30;
  if (/import\s+org\.junit\.Assert\b/.test(source)) score += 25;
  if (/import\s+org\.junit\.Before\b/.test(source)) score += 20;
  if (/import\s+org\.junit\.After\b/.test(source)) score += 20;
  if (/import\s+org\.junit\.BeforeClass\b/.test(source)) score += 20;
  if (/import\s+org\.junit\.AfterClass\b/.test(source)) score += 20;
  if (/import\s+org\.junit\.Ignore\b/.test(source)) score += 15;
  if (/import\s+org\.junit\.\*/.test(source)) score += 25;
  if (/import\s+static\s+org\.junit\.Assert\.\*/.test(source)) score += 25;

  // JUnit 4-specific patterns
  if (/@RunWith\s*\(/.test(source)) score += 15;
  if (/@Rule\b/.test(source)) score += 15;
  if (/@ClassRule\b/.test(source)) score += 15;
  if (/@Test\s*\(\s*expected\s*=/.test(source)) score += 15;
  if (/@Test\s*\(\s*timeout\s*=/.test(source)) score += 15;
  if (/@Category\s*\(/.test(source)) score += 10;
  if (/@Parameterized/.test(source)) score += 15;

  // Weak signals (shared with JUnit 5)
  if (/@Test\b/.test(source)) score += 5;
  if (/Assert\.assertEquals\b/.test(source)) score += 5;
  if (/Assert\.assertTrue\b/.test(source)) score += 5;

  // Negative signals: NOT JUnit 4
  if (/import\s+org\.junit\.jupiter/.test(source)) score -= 40;
  if (/import\s+org\.testng/.test(source)) score -= 40;
  if (/@BeforeEach\b/.test(source)) score -= 20;
  if (/@AfterEach\b/.test(source)) score -= 20;
  if (/@DisplayName\b/.test(source)) score -= 20;
  if (/Assertions\./.test(source)) score -= 20;

  return Math.max(0, Math.min(100, score));
}

/**
 * Parse JUnit 4 source code into a nested IR tree.
 *
 * Uses brace-depth tracking to nest TestCase/Hook inside TestSuite,
 * and Assertion/RawCode inside TestCase/Hook bodies.
 * Extracts assertion subject/expected from Assert.assertEquals(expected, actual).
 */
function parse(source) {
  const lines = source.split('\n');
  const imports = [];
  const rootBody = [];
  const stack = [{ node: null, addChild: (c) => rootBody.push(c), depth: 0 }];
  let depth = 0;
  let pendingAnnotations = []; // @Test, @Before, @Ignore, etc.

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

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();
    const loc = { line: i + 1, column: 0 };

    if (!trimmed) continue;

    const delta = countBraces(trimmed);

    // Comments
    if (
      trimmed.startsWith('//') ||
      trimmed.startsWith('/*') ||
      trimmed.startsWith('*')
    ) {
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
      depth += delta;
      continue;
    }

    // Import statements
    if (/^import\s/.test(trimmed)) {
      const sourceMatch = trimmed.match(/import\s+(?:static\s+)?([^\s;]+)/);
      const imp = new ImportStatement({
        kind: 'library',
        source: sourceMatch ? sourceMatch[1] : '',
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      });
      imports.push(imp);
      depth += delta;
      continue;
    }

    // Annotations — store as pending, consumed by next method/class
    if (/@Test\b/.test(trimmed)) {
      pendingAnnotations.push({
        type: 'test',
        loc,
        originalSource: line,
      });
      depth += delta;
      continue;
    }
    if (/@Before\b(?!Class)/.test(trimmed)) {
      pendingAnnotations.push({
        type: 'beforeEach',
        loc,
        originalSource: line,
      });
      depth += delta;
      continue;
    }
    if (/@After\b(?!Class)/.test(trimmed)) {
      pendingAnnotations.push({ type: 'afterEach', loc, originalSource: line });
      depth += delta;
      continue;
    }
    if (/@BeforeClass\b/.test(trimmed)) {
      pendingAnnotations.push({ type: 'beforeAll', loc, originalSource: line });
      depth += delta;
      continue;
    }
    if (/@AfterClass\b/.test(trimmed)) {
      pendingAnnotations.push({ type: 'afterAll', loc, originalSource: line });
      depth += delta;
      continue;
    }
    if (/@Ignore\b/.test(trimmed)) {
      pendingAnnotations.push({ type: 'skip', loc, originalSource: line });
      depth += delta;
      continue;
    }

    // Class declaration → TestSuite
    if (/\bclass\s+\w+/.test(trimmed)) {
      const suite = new TestSuite({
        name: (trimmed.match(/class\s+(\w+)/) || [])[1] || '',
        modifiers: [],
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      });
      addChild(suite);
      pendingAnnotations = [];
      depth += delta;
      stack.push({
        node: suite,
        addChild: (c) => {
          if (c instanceof Hook) suite.hooks.push(c);
          else suite.tests.push(c);
        },
        depth,
      });
      continue;
    }

    // Method declaration — consume pending annotations
    if (
      /(?:public\s+|protected\s+|private\s+)?(?:static\s+)?void\s+\w+\s*\(/.test(
        trimmed
      )
    ) {
      const methodName = (trimmed.match(/void\s+(\w+)\s*\(/) || [])[1] || '';
      const hookAnnotation = pendingAnnotations.find((a) =>
        ['beforeEach', 'afterEach', 'beforeAll', 'afterAll'].includes(a.type)
      );
      const hasTest = pendingAnnotations.some((a) => a.type === 'test');
      const hasSkip = pendingAnnotations.some((a) => a.type === 'skip');

      if (hookAnnotation) {
        const hook = new Hook({
          hookType: hookAnnotation.type,
          body: [],
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        });
        addChild(hook);
        depth += delta;
        stack.push({ node: hook, addChild: (c) => hook.body.push(c), depth });
      } else {
        const modifiers = [];
        if (hasSkip) {
          modifiers.push(
            new Modifier({ modifierType: 'skip', confidence: 'converted' })
          );
        }
        const tc = new TestCase({
          name: methodName,
          isAsync: false,
          modifiers,
          body: [],
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        });
        addChild(tc);
        depth += delta;
        stack.push({ node: tc, addChild: (c) => tc.body.push(c), depth });
      }
      pendingAnnotations = [];
      continue;
    }

    // @Rule / @ClassRule
    if (/@(?:Class)?Rule\b/.test(trimmed)) {
      addChild(
        new RawCode({
          code: line,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'unconvertible',
        })
      );
      depth += delta;
      continue;
    }

    // Assert calls — extract subject/expected
    if (
      /\bAssert\.\w+\s*\(/.test(trimmed) ||
      /\bassert\w+\s*\(/.test(trimmed)
    ) {
      const assertion = parseJUnit4Assertion(trimmed, loc, line);
      addChild(assertion);
      depth += delta;
      continue;
    }

    // Skip pure structural braces — not test content
    if (!/^[{}]+$/.test(trimmed)) {
      addChild(
        new RawCode({
          code: line,
          sourceLocation: loc,
          originalSource: line,
        })
      );
    }

    depth += delta;

    // Pop containers when braces close
    while (stack.length > 1 && depth < stack[stack.length - 1].depth) {
      stack.pop();
    }
  }

  return new TestFile({
    language: 'java',
    imports,
    body: rootBody,
  });
}

/**
 * Count net brace delta in a line, skipping strings and comments.
 */
function countBraces(line) {
  let delta = 0;
  let inString = false;
  let stringChar = '';
  let inLineComment = false;

  for (let i = 0; i < line.length; i++) {
    const ch = line[i];
    const next = i + 1 < line.length ? line[i + 1] : '';

    if (inLineComment) break;

    if (inString) {
      if (ch === '\\') {
        i++;
        continue;
      }
      if (ch === stringChar) inString = false;
      continue;
    }

    if (ch === '/' && next === '/') {
      inLineComment = true;
      continue;
    }
    if (ch === '"' || ch === "'") {
      inString = true;
      stringChar = ch;
      continue;
    }

    if (ch === '{') delta++;
    else if (ch === '}') delta--;
  }
  return delta;
}

/**
 * Parse a JUnit 4 assertion line and extract kind, subject, expected.
 * JUnit 4 arg order: assertEquals(expected, actual) or assertEquals("msg", expected, actual)
 */
function parseJUnit4Assertion(trimmed, loc, line) {
  let kind = 'equal';
  if (/assertEquals/.test(trimmed)) kind = 'equal';
  else if (/assertTrue/.test(trimmed)) kind = 'truthy';
  else if (/assertFalse/.test(trimmed)) kind = 'falsy';
  else if (/assertNull/.test(trimmed)) kind = 'isNull';
  else if (/assertNotNull/.test(trimmed)) kind = 'isDefined';
  else if (/assertSame/.test(trimmed)) kind = 'strictEqual';
  else if (/assertArrayEquals/.test(trimmed)) kind = 'deepEqual';
  else if (/assertNotEquals/.test(trimmed)) kind = 'notEqual';

  const args = extractJavaAssertArgs(trimmed);
  let subject, expected;

  if (args) {
    if (kind === 'truthy' || kind === 'falsy') {
      // assertTrue(cond) or assertTrue("msg", cond)
      subject = args.length >= 2 ? args[1] : args[0];
    } else if (kind === 'isNull' || kind === 'isDefined') {
      // assertNull(obj) or assertNull("msg", obj)
      subject = args.length >= 2 ? args[1] : args[0];
    } else if (
      kind === 'equal' ||
      kind === 'notEqual' ||
      kind === 'strictEqual' ||
      kind === 'deepEqual'
    ) {
      // assertEquals(expected, actual) or assertEquals("msg", expected, actual)
      if (args.length === 3) {
        expected = args[1];
        subject = args[2];
      } else if (args.length >= 2) {
        expected = args[0];
        subject = args[1];
      }
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
 * Extract arguments from a Java Assert.method(...) call.
 * Splits at top-level commas, respecting nested parens, strings, and generics.
 */
function extractJavaAssertArgs(trimmed) {
  const m = trimmed.match(/(?:Assert\.)?\bassert\w+\s*\(/);
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
  return splitAssertArgs(argsStr);
}

/**
 * Split assertion arguments at commas, respecting parens, strings, and generics.
 */
function splitAssertArgs(argsStr) {
  const args = [];
  let depth = 0;
  let current = '';
  let inString = false;
  let stringChar = '';

  for (let i = 0; i < argsStr.length; i++) {
    const ch = argsStr[i];

    if (inString) {
      current += ch;
      if (ch === '\\') {
        i++;
        current += argsStr[i] || '';
        continue;
      }
      if (ch === stringChar) inString = false;
      continue;
    }

    if (ch === '"' || ch === "'") {
      inString = true;
      stringChar = ch;
      current += ch;
      continue;
    }

    if (ch === '(' || ch === '<' || ch === '{' || ch === '[') {
      depth++;
      current += ch;
    } else if (ch === ')' || ch === '>' || ch === '}' || ch === ']') {
      depth--;
      current += ch;
    } else if (ch === ',' && depth === 0) {
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
 * Emit JUnit 4 code from IR + original source.
 *
 * Stub — JUnit 4 is only used as a source framework in Step 3.
 */
function emit(_ir, source) {
  return source;
}

export default {
  name: 'junit4',
  language: 'java',
  paradigm: 'xunit',
  detect,
  parse,
  emit,
  imports: {
    packages: [
      'org.junit.Test',
      'org.junit.Assert',
      'org.junit.Before',
      'org.junit.After',
    ],
  },
};
