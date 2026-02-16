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
 * Parse JUnit 4 source code into an IR tree.
 */
function parse(source) {
  const lines = source.split('\n');
  const imports = [];
  const allNodes = [];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();
    const loc = { line: i + 1, column: 0 };

    if (!trimmed) continue;

    // Comments
    if (trimmed.startsWith('//') || trimmed.startsWith('/*') || trimmed.startsWith('*')) {
      const isLicense = /license|copyright|MIT|Apache|BSD/i.test(trimmed) && i < 5;
      allNodes.push(new Comment({
        text: line,
        commentKind: isLicense ? 'license' : 'inline',
        preserveExact: isLicense,
        sourceLocation: loc,
        originalSource: line,
      }));
      continue;
    }

    // Import statements
    if (/^import\s/.test(trimmed)) {
      const sourceMatch = trimmed.match(/import\s+(?:static\s+)?([^\s;]+)/);
      allNodes.push(new ImportStatement({
        kind: 'library',
        source: sourceMatch ? sourceMatch[1] : '',
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      }));
      imports.push(allNodes[allNodes.length - 1]);
      continue;
    }

    // Class declaration
    if (/\bclass\s+\w+/.test(trimmed)) {
      allNodes.push(new TestSuite({
        name: (trimmed.match(/class\s+(\w+)/) || [])[1] || '',
        modifiers: [],
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      }));
      continue;
    }

    // @Test annotation (possibly with parameters)
    if (/@Test\b/.test(trimmed)) {
      allNodes.push(new Modifier({
        modifierType: 'test',
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      }));
      continue;
    }

    // @Before / @After / @BeforeClass / @AfterClass
    if (/@Before\b(?!Class)/.test(trimmed)) {
      allNodes.push(new Hook({
        hookType: 'beforeEach',
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      }));
      continue;
    }
    if (/@After\b(?!Class)/.test(trimmed)) {
      allNodes.push(new Hook({
        hookType: 'afterEach',
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      }));
      continue;
    }
    if (/@BeforeClass\b/.test(trimmed)) {
      allNodes.push(new Hook({
        hookType: 'beforeAll',
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      }));
      continue;
    }
    if (/@AfterClass\b/.test(trimmed)) {
      allNodes.push(new Hook({
        hookType: 'afterAll',
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      }));
      continue;
    }

    // @Ignore
    if (/@Ignore\b/.test(trimmed)) {
      allNodes.push(new Modifier({
        modifierType: 'skip',
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      }));
      continue;
    }

    // @Rule / @ClassRule
    if (/@(?:Class)?Rule\b/.test(trimmed)) {
      allNodes.push(new RawCode({
        code: line,
        sourceLocation: loc,
        originalSource: line,
        confidence: 'unconvertible',
      }));
      continue;
    }

    // Test methods
    if (/public\s+void\s+\w+\s*\(/.test(trimmed)) {
      allNodes.push(new TestCase({
        name: (trimmed.match(/void\s+(\w+)\s*\(/) || [])[1] || '',
        isAsync: false,
        modifiers: [],
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      }));
      continue;
    }

    // Assert calls
    if (/\bAssert\.\w+\s*\(/.test(trimmed) || /\bassert\w+\s*\(/.test(trimmed)) {
      let kind = 'equal';
      if (/assertEquals/.test(trimmed)) kind = 'equal';
      else if (/assertTrue/.test(trimmed)) kind = 'truthy';
      else if (/assertFalse/.test(trimmed)) kind = 'falsy';
      else if (/assertNull/.test(trimmed)) kind = 'isNull';
      else if (/assertNotNull/.test(trimmed)) kind = 'isDefined';
      else if (/assertSame/.test(trimmed)) kind = 'strictEqual';
      else if (/assertArrayEquals/.test(trimmed)) kind = 'deepEqual';
      else if (/assertNotEquals/.test(trimmed)) kind = 'notEqual';

      allNodes.push(new Assertion({
        kind,
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      }));
      continue;
    }

    // Everything else
    allNodes.push(new RawCode({
      code: line,
      sourceLocation: loc,
      originalSource: line,
    }));
  }

  return new TestFile({
    language: 'java',
    imports,
    body: allNodes.filter(n => !imports.includes(n)),
  });
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
    packages: ['org.junit.Test', 'org.junit.Assert', 'org.junit.Before', 'org.junit.After'],
  },
};
