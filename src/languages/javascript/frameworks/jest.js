/**
 * Jest framework definition.
 *
 * Provides detect, parse, and emit for the Jest testing framework.
 * parse() builds an IR tree from Jest source code for scoring.
 * emit() generates Jest code from IR (used when Jest is the target).
 */

import {
  TestFile,
  TestSuite,
  TestCase,
  Hook,
  Assertion,
  MockCall,
  ImportStatement,
  RawCode,
  Comment,
  Modifier,
  ParameterSet,
} from '../../../core/ir.js';

/**
 * Detect whether source code is Jest.
 * Returns confidence score 0-100.
 */
function detect(source) {
  if (!source || !source.trim()) return 0;

  let score = 0;

  // Jest-specific APIs (strong signals)
  if (/jest\.fn\s*\(/.test(source)) score += 25;
  if (/jest\.mock\s*\(/.test(source)) score += 25;
  if (/jest\.spyOn\s*\(/.test(source)) score += 20;
  if (/jest\.useFakeTimers/.test(source)) score += 15;
  if (/jest\.requireActual/.test(source)) score += 15;
  if (/jest\.setTimeout/.test(source)) score += 10;
  if (/jest\.clearAllMocks|jest\.resetAllMocks|jest\.restoreAllMocks/.test(source)) score += 10;
  if (/jest\.retryTimes/.test(source)) score += 10;

  // Test structure (weak — shared with Vitest, Mocha, Jasmine)
  if (/\bdescribe\s*\(/.test(source)) score += 5;
  if (/\b(?:it|test)\s*\(/.test(source)) score += 5;
  if (/\bexpect\s*\(/.test(source)) score += 5;
  if (/\bbeforeEach\s*\(/.test(source)) score += 3;

  // Negative signal: Vitest imports mean it's NOT Jest
  if (/from\s+['"]vitest['"]/.test(source)) score -= 50;
  if (/\bvi\.fn\b/.test(source)) score -= 30;

  // Negative: Mocha/Chai
  if (/from\s+['"]chai['"]/.test(source)) score -= 20;
  if (/require\s*\(\s*['"]chai['"]\s*\)/.test(source)) score -= 20;

  return Math.max(0, Math.min(100, score));
}

/**
 * Parse Jest source code into an IR tree.
 *
 * This is a regex-based parser that identifies test structure elements
 * and creates IR nodes for scoring and conversion tracking. It doesn't
 * attempt to parse every JavaScript expression — unrecognized code
 * passes through as RawCode nodes.
 */
function parse(source) {
  const lines = source.split('\n');
  const imports = [];
  const body = [];
  const allNodes = [];

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();
    const loc = { line: i + 1, column: 0 };

    // Skip empty lines
    if (!trimmed) continue;

    // Comments
    if (trimmed.startsWith('//') || trimmed.startsWith('/*') || trimmed.startsWith('*')) {
      const isDirective = /eslint-disable|noinspection|@ts-|type:\s*ignore/.test(trimmed);
      const isLicense = /license|copyright|MIT|Apache|BSD/i.test(trimmed) && i < 5;
      const node = new Comment({
        text: line,
        commentKind: isLicense ? 'license' : isDirective ? 'directive' : 'inline',
        preserveExact: isDirective || isLicense,
        sourceLocation: loc,
        originalSource: line,
      });
      allNodes.push(node);
      continue;
    }

    // Import/require statements
    if (/^import\s/.test(trimmed) || /^const\s.*=\s*require\(/.test(trimmed)) {
      const isTypeOnly = /^import\s+type\s/.test(trimmed);
      const sourceMatch = trimmed.match(/from\s+['"]([^'"]+)['"]/) ||
                          trimmed.match(/require\s*\(\s*['"]([^'"]+)['"]\s*\)/);
      const node = new ImportStatement({
        kind: 'library',
        source: sourceMatch ? sourceMatch[1] : '',
        isTypeOnly,
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      });
      imports.push(node);
      allNodes.push(node);
      continue;
    }

    // jest.mock() at top level
    if (/^jest\.mock\s*\(/.test(trimmed)) {
      const hasFactory = /jest\.mock\s*\([^,]+,/.test(trimmed);
      const hasVirtual = /\{\s*virtual:\s*true\s*\}/.test(trimmed);
      // Collect multi-line mock
      let fullMock = line;
      if (!isBalanced(trimmed)) {
        let j = i + 1;
        while (j < lines.length && !isBalanced(fullMock)) {
          fullMock += '\n' + lines[j];
          j++;
        }
        i = j - 1;
      }
      const node = new MockCall({
        kind: hasVirtual ? 'mockModule' : hasFactory ? 'mockModule' : 'mockModule',
        target: (trimmed.match(/jest\.mock\s*\(\s*['"]([^'"]+)['"]/) || [])[1] || '',
        sourceLocation: loc,
        originalSource: fullMock,
        confidence: hasVirtual ? 'unconvertible' : 'converted',
      });
      allNodes.push(node);
      continue;
    }

    // jest.setTimeout / jest.retryTimes at top level
    if (/^jest\.setTimeout\s*\(/.test(trimmed)) {
      allNodes.push(new MockCall({
        kind: 'fakeTimers',
        target: 'setTimeout',
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      }));
      continue;
    }
    if (/^jest\.retryTimes\s*\(/.test(trimmed)) {
      allNodes.push(new MockCall({
        kind: 'fakeTimers',
        target: 'retryTimes',
        sourceLocation: loc,
        originalSource: line,
        confidence: 'warning',
      }));
      continue;
    }

    // describe / it / test / hooks / assertions inside code
    // These are tracked for scoring but we don't need to build a full tree
    // for same-paradigm conversions
    if (/\bdescribe\s*\(/.test(trimmed) || /\bdescribe\.(?:only|skip|each)\s*[\(`]/.test(trimmed)) {
      const hasSkip = /\.skip/.test(trimmed);
      const hasOnly = /\.only/.test(trimmed);
      const modifiers = [];
      if (hasSkip) modifiers.push(new Modifier({ modifierType: 'skip', sourceLocation: loc }));
      if (hasOnly) modifiers.push(new Modifier({ modifierType: 'only', sourceLocation: loc }));
      allNodes.push(new TestSuite({
        name: (trimmed.match(/describe(?:\.\w+)*\s*\(\s*['"`]([^'"`]*)['"`]/) || [])[1] || '',
        modifiers,
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      }));
      continue;
    }

    if (/\b(?:it|test)\s*\(/.test(trimmed) || /\b(?:it|test)\.(?:only|skip|todo|each)\s*[\(`]/.test(trimmed)) {
      const hasTodo = /\.todo/.test(trimmed);
      const hasSkip = /\.skip/.test(trimmed) || /^x(?:it|test)\s*\(/.test(trimmed);
      const hasOnly = /\.only/.test(trimmed) || /^f(?:it|test)\s*\(/.test(trimmed);
      const isAsync = /async\s*\(/.test(trimmed) || /async\s*\(\s*\{/.test(trimmed);
      const modifiers = [];
      if (hasTodo) modifiers.push(new Modifier({ modifierType: 'pending', sourceLocation: loc }));
      if (hasSkip) modifiers.push(new Modifier({ modifierType: 'skip', sourceLocation: loc }));
      if (hasOnly) modifiers.push(new Modifier({ modifierType: 'only', sourceLocation: loc }));
      allNodes.push(new TestCase({
        name: (trimmed.match(/(?:it|test)(?:\.\w+)*\s*\(\s*['"`]([^'"`]*)['"`]/) || [])[1] || '',
        isAsync,
        modifiers,
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      }));
      continue;
    }

    if (/\b(?:beforeEach|afterEach|beforeAll|afterAll)\s*\(/.test(trimmed)) {
      const hookType = (trimmed.match(/\b(beforeEach|afterEach|beforeAll|afterAll)/) || [])[1];
      allNodes.push(new Hook({
        hookType,
        isAsync: /async/.test(trimmed),
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      }));
      continue;
    }

    // Assertions: expect(...)
    if (/\bexpect\s*\(/.test(trimmed)) {
      const isNegated = /\.not\./.test(trimmed);
      let kind = 'equal';
      if (/\.toBe\(/.test(trimmed)) kind = 'strictEqual';
      else if (/\.toEqual\(/.test(trimmed)) kind = 'deepEqual';
      else if (/\.toBeTruthy\(/.test(trimmed)) kind = 'truthy';
      else if (/\.toBeFalsy\(/.test(trimmed)) kind = 'falsy';
      else if (/\.toBeNull\(/.test(trimmed)) kind = 'isNull';
      else if (/\.toBeUndefined\(/.test(trimmed)) kind = 'isUndefined';
      else if (/\.toBeDefined\(/.test(trimmed)) kind = 'isDefined';
      else if (/\.toBeNaN\(/.test(trimmed)) kind = 'isNaN';
      else if (/\.toBeInstanceOf\(/.test(trimmed)) kind = 'instanceOf';
      else if (/\.toMatch\(/.test(trimmed)) kind = 'matches';
      else if (/\.toContain\(/.test(trimmed)) kind = 'contains';
      else if (/\.toContainEqual\(/.test(trimmed)) kind = 'containsEqual';
      else if (/\.toHaveLength\(/.test(trimmed)) kind = 'hasLength';
      else if (/\.toHaveProperty\(/.test(trimmed)) kind = 'hasProperty';
      else if (/\.toBeGreaterThan\(/.test(trimmed)) kind = 'greaterThan';
      else if (/\.toBeLessThan\(/.test(trimmed)) kind = 'lessThan';
      else if (/\.toBeGreaterThanOrEqual\(/.test(trimmed)) kind = 'greaterOrEqual';
      else if (/\.toBeLessThanOrEqual\(/.test(trimmed)) kind = 'lessOrEqual';
      else if (/\.toBeCloseTo\(/.test(trimmed)) kind = 'closeTo';
      else if (/\.toThrow\(/.test(trimmed)) kind = 'throws';
      else if (/\.toHaveBeenCalled\b/.test(trimmed)) kind = 'called';
      else if (/\.toHaveBeenCalledWith\(/.test(trimmed)) kind = 'calledWith';
      else if (/\.toHaveBeenCalledTimes\(/.test(trimmed)) kind = 'calledTimes';
      else if (/\.toMatchSnapshot\(/.test(trimmed)) kind = 'snapshot';
      else if (/\.toMatchInlineSnapshot\(/.test(trimmed)) kind = 'snapshot';
      else if (/\.toHaveBeenLastCalledWith\(/.test(trimmed)) kind = 'calledWith';
      else if (/\.resolves\./.test(trimmed)) kind = 'resolves';
      else if (/\.rejects\./.test(trimmed)) kind = 'rejects';
      else if (/\.toStrictEqual\(/.test(trimmed)) kind = 'strictEqual';
      else if (/\.toHaveClass\(/.test(trimmed)) kind = 'hasClass';
      else if (/\.toHaveCount\(/.test(trimmed)) kind = 'hasCount';

      allNodes.push(new Assertion({
        kind,
        isNegated,
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      }));
      continue;
    }

    // jest.fn / jest.spyOn / jest.mock inline
    if (/jest\.fn\s*\(/.test(trimmed) || /jest\.spyOn\s*\(/.test(trimmed)) {
      const kind = /jest\.fn/.test(trimmed) ? 'createMock' : 'spyOnMethod';
      allNodes.push(new MockCall({
        kind,
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      }));
      continue;
    }

    // jest.useFakeTimers / jest.useRealTimers / jest.advanceTimersByTime / jest.runAllTimers
    if (/jest\.\w+/.test(trimmed)) {
      allNodes.push(new MockCall({
        kind: 'fakeTimers',
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      }));
      continue;
    }

    // expect.extend, expect.assertions, expect.addSnapshotSerializer
    if (/\bexpect\.extend\s*\(/.test(trimmed)) {
      allNodes.push(new RawCode({
        code: line,
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      }));
      continue;
    }
    if (/\bexpect\.addSnapshotSerializer\s*\(/.test(trimmed)) {
      allNodes.push(new RawCode({
        code: line,
        sourceLocation: loc,
        originalSource: line,
        confidence: 'unconvertible',
      }));
      continue;
    }

    // Everything else is raw code
    allNodes.push(new RawCode({
      code: line,
      sourceLocation: loc,
      originalSource: line,
    }));
  }

  return new TestFile({
    language: 'javascript',
    imports,
    body: allNodes.filter(n => !imports.includes(n)),
  });
}

/**
 * Emit Jest code from IR.
 * Used when Jest is the target framework (Vitest→Jest conversion).
 * Stub for now — only Jest→Vitest is implemented.
 */
function emit(_ir, _source) {
  throw new Error('Jest emit not yet implemented (Vitest→Jest direction)');
}

/**
 * Check if parentheses/brackets are balanced in a string.
 */
function isBalanced(str) {
  let depth = 0;
  for (const ch of str) {
    if (ch === '(' || ch === '{' || ch === '[') depth++;
    if (ch === ')' || ch === '}' || ch === ']') depth--;
  }
  return depth <= 0;
}

export default {
  name: 'jest',
  language: 'javascript',
  paradigm: 'bdd',
  detect,
  parse,
  emit,
  imports: {
    globals: ['describe', 'it', 'test', 'expect', 'beforeEach', 'afterEach', 'beforeAll', 'afterAll'],
    mockNamespace: 'jest',
  },
};
