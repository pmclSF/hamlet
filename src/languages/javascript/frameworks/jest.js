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
import { TodoFormatter } from '../../../core/TodoFormatter.js';

const formatter = new TodoFormatter('javascript');

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
  if (
    /jest\.clearAllMocks|jest\.resetAllMocks|jest\.restoreAllMocks/.test(source)
  )
    score += 10;
  if (/jest\.retryTimes/.test(source)) score += 10;

  // Test structure (weak — shared with Vitest, Mocha, Jasmine)
  if (/\bdescribe\s*\(/.test(source)) score += 5;
  if (/\b(?:it|test)\s*\(/.test(source)) score += 5;
  if (/\bexpect\s*\(/.test(source)) score += 5;
  if (/\bbeforeEach\s*\(/.test(source)) score += 3;

  // Negative signal: Vitest imports mean it's NOT Jest
  if (/from\s+['"]vitest['"]/.test(source)) score -= 50;
  if (/\bvi\.fn\b/.test(source)) score -= 30;

  // Negative: Mocha/Chai/Sinon
  if (/from\s+['"]chai['"]/.test(source)) score -= 20;
  if (/require\s*\(\s*['"]chai['"]\s*\)/.test(source)) score -= 20;
  if (/require\s*\(\s*['"]sinon['"]\s*\)/.test(source)) score -= 20;
  if (/from\s+['"]sinon['"]/.test(source)) score -= 20;

  // Negative: Jasmine-specific
  if (/jasmine\.createSpy/.test(source)) score -= 20;
  if (/jasmine\.createSpyObj/.test(source)) score -= 20;

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
    if (
      trimmed.startsWith('//') ||
      trimmed.startsWith('/*') ||
      trimmed.startsWith('*')
    ) {
      const isDirective =
        /eslint-disable|noinspection|@ts-|type:\s*ignore/.test(trimmed);
      const isLicense =
        /license|copyright|MIT|Apache|BSD/i.test(trimmed) && i < 5;
      const node = new Comment({
        text: line,
        commentKind: isLicense
          ? 'license'
          : isDirective
            ? 'directive'
            : 'inline',
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
      const sourceMatch =
        trimmed.match(/from\s+['"]([^'"]+)['"]/) ||
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
      const hasRequireActual = /jest\.requireActual/.test(fullMock);
      const node = new MockCall({
        kind: hasVirtual
          ? 'mockModule'
          : hasFactory
            ? 'mockModule'
            : 'mockModule',
        target:
          (trimmed.match(/jest\.mock\s*\(\s*['"]([^'"]+)['"]/) || [])[1] || '',
        sourceLocation: loc,
        originalSource: fullMock,
        confidence: hasVirtual ? 'unconvertible' : 'converted',
        requiresAsync: hasRequireActual,
      });
      allNodes.push(node);
      continue;
    }

    // jest.setTimeout / jest.retryTimes at top level
    if (/^jest\.setTimeout\s*\(/.test(trimmed)) {
      allNodes.push(
        new MockCall({
          kind: 'fakeTimers',
          target: 'setTimeout',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }
    if (/^jest\.retryTimes\s*\(/.test(trimmed)) {
      allNodes.push(
        new MockCall({
          kind: 'fakeTimers',
          target: 'retryTimes',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'warning',
          frameworkSpecific: true,
        })
      );
      continue;
    }

    // describe / it / test / hooks / assertions inside code
    // These are tracked for scoring but we don't need to build a full tree
    // for same-paradigm conversions
    if (
      /\bdescribe\s*\(/.test(trimmed) ||
      /\bdescribe\.(?:only|skip|each)\s*[\(`]/.test(trimmed)
    ) {
      const hasSkip = /\.skip/.test(trimmed);
      const hasOnly = /\.only/.test(trimmed);
      const modifiers = [];
      if (hasSkip)
        modifiers.push(
          new Modifier({ modifierType: 'skip', sourceLocation: loc })
        );
      if (hasOnly)
        modifiers.push(
          new Modifier({ modifierType: 'only', sourceLocation: loc })
        );
      allNodes.push(
        new TestSuite({
          name:
            (trimmed.match(/describe(?:\.\w+)*\s*\(\s*['"`]([^'"`]*)['"`]/) ||
              [])[1] || '',
          modifiers,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    if (
      /\b(?:it|test)\s*\(/.test(trimmed) ||
      /\b(?:it|test)\.(?:only|skip|todo|each)\s*[\(`]/.test(trimmed)
    ) {
      const hasTodo = /\.todo/.test(trimmed);
      const hasSkip =
        /\.skip/.test(trimmed) || /^x(?:it|test)\s*\(/.test(trimmed);
      const hasOnly =
        /\.only/.test(trimmed) || /^f(?:it|test)\s*\(/.test(trimmed);
      const isAsync =
        /async\s*\(/.test(trimmed) || /async\s*\(\s*\{/.test(trimmed);
      const modifiers = [];
      if (hasTodo)
        modifiers.push(
          new Modifier({ modifierType: 'pending', sourceLocation: loc })
        );
      if (hasSkip)
        modifiers.push(
          new Modifier({ modifierType: 'skip', sourceLocation: loc })
        );
      if (hasOnly)
        modifiers.push(
          new Modifier({ modifierType: 'only', sourceLocation: loc })
        );
      allNodes.push(
        new TestCase({
          name:
            (trimmed.match(
              /(?:it|test)(?:\.\w+)*\s*\(\s*['"`]([^'"`]*)['"`]/
            ) || [])[1] || '',
          isAsync,
          modifiers,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    if (/\b(?:beforeEach|afterEach|beforeAll|afterAll)\s*\(/.test(trimmed)) {
      const hookType = (trimmed.match(
        /\b(beforeEach|afterEach|beforeAll|afterAll)/
      ) || [])[1];
      allNodes.push(
        new Hook({
          hookType,
          isAsync: /async/.test(trimmed),
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
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
      else if (/\.toBeGreaterThanOrEqual\(/.test(trimmed))
        kind = 'greaterOrEqual';
      else if (/\.toBeLessThanOrEqual\(/.test(trimmed)) kind = 'lessOrEqual';
      else if (/\.toBeCloseTo\(/.test(trimmed)) kind = 'closeTo';
      else if (/\.toThrow\(/.test(trimmed)) kind = 'throws';
      else if (/\.toHaveBeenCalled\b/.test(trimmed)) kind = 'called';
      else if (/\.toHaveBeenCalledWith\(/.test(trimmed)) kind = 'calledWith';
      else if (/\.toHaveBeenCalledTimes\(/.test(trimmed)) kind = 'calledTimes';
      else if (/\.toMatchSnapshot\(/.test(trimmed)) kind = 'snapshot';
      else if (/\.toMatchInlineSnapshot\(/.test(trimmed)) kind = 'snapshot';
      else if (/\.toHaveBeenLastCalledWith\(/.test(trimmed))
        kind = 'calledWith';
      else if (/\.resolves\./.test(trimmed)) kind = 'resolves';
      else if (/\.rejects\./.test(trimmed)) kind = 'rejects';
      else if (/\.toStrictEqual\(/.test(trimmed)) kind = 'strictEqual';
      else if (/\.toHaveClass\(/.test(trimmed)) kind = 'hasClass';
      else if (/\.toHaveCount\(/.test(trimmed)) kind = 'hasCount';

      allNodes.push(
        new Assertion({
          kind,
          isNegated,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // jest.fn / jest.spyOn / jest.mock inline
    if (/jest\.fn\s*\(/.test(trimmed) || /jest\.spyOn\s*\(/.test(trimmed)) {
      const kind = /jest\.fn/.test(trimmed) ? 'createMock' : 'spyOnMethod';
      allNodes.push(
        new MockCall({
          kind,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // jest.useFakeTimers / jest.useRealTimers / jest.advanceTimersByTime / jest.runAllTimers
    if (/jest\.\w+/.test(trimmed)) {
      allNodes.push(
        new MockCall({
          kind: 'fakeTimers',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // expect.extend, expect.assertions, expect.addSnapshotSerializer
    if (/\bexpect\.extend\s*\(/.test(trimmed)) {
      allNodes.push(
        new RawCode({
          code: line,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }
    if (/\bexpect\.addSnapshotSerializer\s*\(/.test(trimmed)) {
      allNodes.push(
        new RawCode({
          code: line,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'unconvertible',
        })
      );
      continue;
    }

    // Everything else is raw code
    allNodes.push(
      new RawCode({
        code: line,
        sourceLocation: loc,
        originalSource: line,
      })
    );
  }

  return new TestFile({
    language: 'javascript',
    imports,
    body: allNodes.filter((n) => !imports.includes(n)),
  });
}

/**
 * Emit Jest code from IR + original source.
 *
 * Uses IR-guided chunking: walks the IR node tree, identifies node types
 * (imports, suites, tests, hooks, assertions, mocks), and applies only
 * the relevant regex transforms per node. This avoids ordering bugs from
 * sweeping 100+ regexes across the entire file and uses IR structure for
 * correctness guarantees.
 *
 * Falls back to original source for unrecognized patterns (RawCode nodes).
 *
 * @param {TestFile} ir - Parsed IR tree from source framework
 * @param {string} source - Original source code
 * @returns {string} Converted Jest source code
 */
function emit(ir, source) {
  // Build a line→node map from the IR for guided chunking
  const nodeByLine = new Map();
  if (ir && ir.body) {
    for (const node of [...(ir.imports || []), ...ir.body]) {
      if (node.sourceLocation) {
        nodeByLine.set(node.sourceLocation.line, node);
      }
    }
  }

  // --- File-level Phase 0: Restore multi-line HAMLET-TODO blocks ---
  let result = source;
  result = result.replace(
    /\/\/ HAMLET-TODO \[UNCONVERTIBLE-SNAPSHOT\][^\n]*\n\s*\/\/ Original: ([^\n]+)\n\s*\/\/ Manual action required:[^\n]*\n\s*\/\/[^\n]*/g,
    '$1'
  );
  result = result.replace(
    /\/\/ HAMLET-TODO \[UNCONVERTIBLE-INLINE-SNAPSHOT\][^\n]*\n\s*\/\/ Original: ([^\n]+)\n\s*\/\/ Manual action required:[^\n]*\n\s*\/\/[^\n]*/g,
    '$1'
  );
  result = result.replace(
    /\/\/ HAMLET-TODO \[UNCONVERTIBLE-MODULE-MOCK\][^\n]*\n\s*\/\/ Original: ([^\n]+)\n\s*\/\/ Manual action required:[^\n]*\n\s*\/\/[^\n]*/g,
    '$1'
  );
  result = result.replace(
    /\/\/ Jasmine spies are auto-cleaned between specs/g,
    'jest.restoreAllMocks();'
  );

  // --- Per-node transforms: walk each line with its IR node type ---
  const lines = result.split('\n');
  const transformed = lines.map((line, i) => {
    const lineNum = i + 1;
    const node = nodeByLine.get(lineNum);

    if (!node) return line; // blank lines or unmapped

    switch (node.type) {
      case 'ImportStatement':
        return transformImportToJest(line);
      case 'TestSuite':
        return transformSuiteToJest(line);
      case 'TestCase':
        return transformTestCaseToJest(line);
      case 'Hook':
        return transformHookToJest(line);
      case 'Assertion':
        return transformAssertionToJest(line);
      case 'MockCall':
        return transformMockToJest(line);
      case 'Comment':
        return line; // pass-through
      default:
        // RawCode or unknown: apply conservative chain transforms
        // for multi-line assertion/mock continuations
        return transformRawCodeToJest(line);
    }
  });

  result = transformed.join('\n');

  // --- File-level cleanup ---
  result = result.replace(/\n{3,}/g, '\n\n');
  if (!result.endsWith('\n')) result += '\n';

  return result;
}

// ── Per-node transform functions (Jest target) ───────────────────────

/**
 * Transform import lines: strip chai, sinon, and related imports.
 */
function transformImportToJest(line) {
  // Remove chai imports
  if (
    /require\s*\(\s*['"]chai['"]\s*\)/.test(line) ||
    /from\s+['"]chai['"]/.test(line)
  )
    return '';
  // Remove sinon imports
  if (
    /require\s*\(\s*['"]sinon['"]\s*\)/.test(line) ||
    /from\s+['"]sinon['"]/.test(line)
  )
    return '';
  // Remove chai-as-promised, sinon-chai
  if (/chai-as-promised|sinon-chai/.test(line)) return '';
  return line;
}

/**
 * Transform test suite lines: Mocha/Jasmine aliases → Jest describe.
 */
function transformSuiteToJest(line) {
  let r = line;
  r = r.replace(/\bcontext\s*\(/g, 'describe(');
  r = r.replace(/\bfdescribe\s*\(/g, 'describe.only(');
  r = r.replace(/\bxdescribe\s*\(/g, 'describe.skip(');
  return r;
}

/**
 * Transform test case lines: Mocha specify, Jasmine f/xit → Jest it.
 */
function transformTestCaseToJest(line) {
  let r = line;
  r = r.replace(/\bspecify\s*\(/g, 'it(');
  // xit with pending → it.todo (before generic xit→it.skip)
  r = r.replace(
    /\bxit\(\s*(['"][^'"]*['"])\s*,\s*(?:\(\)\s*=>|function\s*\(\))\s*\{\s*pending\(\)\s*;?\s*\}\s*\)/g,
    'it.todo($1)'
  );
  r = r.replace(/\bxit\s*\(/g, 'it.skip(');
  r = r.replace(/\bfit\s*\(/g, 'it.only(');
  return r;
}

/**
 * Transform hook lines: Mocha before/after → Jest beforeAll/afterAll.
 */
function transformHookToJest(line) {
  let r = line;
  r = r.replace(/\bbefore\s*\((?!Each|All)/g, 'beforeAll(');
  r = r.replace(/\bafter\s*\((?!Each|All)/g, 'afterAll(');
  return r;
}

/**
 * Transform assertion lines: Chai chains + assert-style → Jest expect.
 */
function transformAssertionToJest(line) {
  let r = line;

  // Negation normalization
  r = r.replace(/\.to\.not\./g, '.not.');
  r = r.replace(/\.not\.to\./g, '.not.');

  // Chai-sinon assertion chains (most specific first)
  r = r.replace(/\.to\.have\.been\.calledOnce/g, '.toHaveBeenCalledTimes(1)');
  r = r.replace(/\.to\.have\.been\.calledTwice/g, '.toHaveBeenCalledTimes(2)');
  r = r.replace(/\.to\.have\.been\.calledWith\(/g, '.toHaveBeenCalledWith(');
  r = r.replace(/\.to\.have\.been\.called(?!\w)/g, '.toHaveBeenCalled()');
  r = r.replace(/\.not\.have\.been\.called(?!\w)/g, '.not.toHaveBeenCalled()');

  // sinon.assert → expect (classified as Assertion by mocha parser)
  r = r.replace(
    /sinon\.assert\.calledOnce\(([^)]+)\)/g,
    'expect($1).toHaveBeenCalledTimes(1)'
  );
  r = r.replace(
    /sinon\.assert\.calledTwice\(([^)]+)\)/g,
    'expect($1).toHaveBeenCalledTimes(2)'
  );
  r = r.replace(
    /sinon\.assert\.calledWith\(([^,]+),\s*([^)]*)\)/g,
    'expect($1).toHaveBeenCalledWith($2)'
  );
  r = r.replace(
    /sinon\.assert\.notCalled\(([^)]+)\)/g,
    'expect($1).not.toHaveBeenCalled()'
  );
  r = r.replace(
    /sinon\.assert\.callCount\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).toHaveBeenCalledTimes($2)'
  );
  r = r.replace(
    /sinon\.assert\.called\(([^)]+)\)/g,
    'expect($1).toHaveBeenCalled()'
  );

  // Jasmine spy .calls API inside assertions (e.g. expect(spy.calls.argsFor(0)))
  r = r.replace(/\.calls\.argsFor\((\d+)\)/g, '.mock.calls[$1]');
  r = r.replace(/\.calls\.mostRecent\(\)\.args/g, '.mock.lastCall');
  r = r.replace(/\.calls\.mostRecent\(\)/g, '.mock.lastCall');
  r = r.replace(/\.calls\.first\(\)\.args/g, '.mock.calls[0]');
  r = r.replace(/\.calls\.allArgs\(\)/g, '.mock.calls');
  r = r.replace(/\.calls\.reset\(\)/g, '.mockClear()');

  // Mock property transforms inside assertions (e.g. expect(spy.callCount))
  // MUST run after sinon.assert.callCount to avoid clobbering
  r = r.replace(/\.callCount\b/g, '.mock.calls.length');
  r = r.replace(/\.calls\.count\(\)/g, '.mock.calls.length');

  // Deep equal before equal (more specific first)
  r = r.replace(/\.to\.deep\.equal\(/g, '.toEqual(');
  r = r.replace(/\.not\.deep\.equal\(/g, '.not.toEqual(');

  // .to.equal → .toBe
  r = r.replace(/\.to\.equal\(/g, '.toBe(');
  r = r.replace(/\.not\.equal\(/g, '.not.toBe(');

  // Boolean/null/undefined terminal chains
  r = r.replace(/\.to\.be\.true/g, '.toBe(true)');
  r = r.replace(/\.not\.be\.true/g, '.not.toBe(true)');
  r = r.replace(/\.to\.be\.false/g, '.toBe(false)');
  r = r.replace(/\.not\.be\.false/g, '.not.toBe(false)');
  r = r.replace(/\.to\.be\.null/g, '.toBeNull()');
  r = r.replace(/\.not\.be\.null/g, '.not.toBeNull()');
  r = r.replace(/\.to\.be\.undefined/g, '.toBeUndefined()');
  r = r.replace(/\.not\.be\.undefined/g, '.not.toBeUndefined()');
  r = r.replace(/\.to\.be\.NaN/g, '.toBeNaN()');
  r = r.replace(/\.not\.be\.NaN/g, '.not.toBeNaN()');
  r = r.replace(/\.to\.be\.ok/g, '.toBeTruthy()');
  r = r.replace(/\.not\.be\.ok/g, '.toBeFalsy()');

  // .to.be.an.instanceOf(Y) → .toBeInstanceOf(Y)
  r = r.replace(/\.to\.be\.an?\.instanceOf\(/g, '.toBeInstanceOf(');
  r = r.replace(/\.not\.be\.an?\.instanceOf\(/g, '.not.toBeInstanceOf(');

  // .to.be.a('type') → typeof check
  const SUBJ = '([^)]*(?:\\([^)]*\\)[^)]*)*)';
  r = r.replace(
    new RegExp(
      'expect\\(' + SUBJ + '\\)\\.to\\.be\\.a\\(([\'"][^\'"]+[\'"])\\)',
      'g'
    ),
    'expect(typeof $1).toBe($2)'
  );
  r = r.replace(
    new RegExp(
      'expect\\(' + SUBJ + '\\)\\.to\\.be\\.an\\(([\'"]array[\'"])\\)',
      'g'
    ),
    'expect(Array.isArray($1)).toBe(true)'
  );
  r = r.replace(
    new RegExp(
      'expect\\(' + SUBJ + '\\)\\.to\\.be\\.an\\(([\'"][^\'"]+[\'"])\\)',
      'g'
    ),
    'expect(typeof $1).toBe($2)'
  );

  // Chain replacements with args
  r = r.replace(/\.to\.have\.lengthOf\(/g, '.toHaveLength(');
  r = r.replace(/\.not\.have\.lengthOf\(/g, '.not.toHaveLength(');
  r = r.replace(/\.to\.include\(/g, '.toContain(');
  r = r.replace(/\.not\.include\(/g, '.not.toContain(');
  r = r.replace(/\.to\.contain\(/g, '.toContain(');
  r = r.replace(/\.not\.contain\(/g, '.not.toContain(');
  r = r.replace(/\.to\.have\.property\(/g, '.toHaveProperty(');
  r = r.replace(/\.not\.have\.property\(/g, '.not.toHaveProperty(');
  r = r.replace(/\.to\.be\.above\(/g, '.toBeGreaterThan(');
  r = r.replace(/\.to\.be\.below\(/g, '.toBeLessThan(');
  r = r.replace(/\.to\.be\.at\.least\(/g, '.toBeGreaterThanOrEqual(');
  r = r.replace(/\.to\.be\.at\.most\(/g, '.toBeLessThanOrEqual(');
  r = r.replace(/\.to\.be\.closeTo\(/g, '.toBeCloseTo(');
  r = r.replace(/\.to\.throw\(/g, '.toThrow(');
  r = r.replace(/\.not\.throw\(/g, '.not.toThrow(');
  r = r.replace(/\.to\.throw(?!\w)/g, '.toThrow()');
  r = r.replace(/\.not\.throw(?!\w)/g, '.not.toThrow()');
  r = r.replace(/\.to\.match\(/g, '.toMatch(');
  r = r.replace(/\.not\.match\(/g, '.not.toMatch(');
  r = r.replace(/\.to\.exist/g, '.toBeDefined()');
  r = r.replace(/\.not\.exist/g, '.not.toBeDefined()');

  // Chai assert-style
  r = r.replace(
    /assert\.strictEqual\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).toBe($2)'
  );
  r = r.replace(
    /assert\.deepEqual\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).toEqual($2)'
  );
  r = r.replace(/assert\.equal\(([^,]+),\s*([^)]+)\)/g, 'expect($1).toBe($2)');
  r = r.replace(
    /assert\.notEqual\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).not.toBe($2)'
  );
  r = r.replace(/assert\.isTrue\(([^)]+)\)/g, 'expect($1).toBe(true)');
  r = r.replace(/assert\.isFalse\(([^)]+)\)/g, 'expect($1).toBe(false)');
  r = r.replace(/assert\.isNull\(([^)]+)\)/g, 'expect($1).toBeNull()');
  r = r.replace(/assert\.isNotNull\(([^)]+)\)/g, 'expect($1).not.toBeNull()');
  r = r.replace(
    /assert\.isUndefined\(([^)]+)\)/g,
    'expect($1).toBeUndefined()'
  );
  r = r.replace(/assert\.isDefined\(([^)]+)\)/g, 'expect($1).toBeDefined()');
  r = r.replace(/assert\.isOk\(([^)]+)\)/g, 'expect($1).toBeTruthy()');
  r = r.replace(/assert\.isNotOk\(([^)]+)\)/g, 'expect($1).toBeFalsy()');
  r = r.replace(/assert\.throws\(([^)]+)\)/g, 'expect($1).toThrow()');
  r = r.replace(/assert\.doesNotThrow\(([^)]+)\)/g, 'expect($1).not.toThrow()');
  r = r.replace(
    /assert\.include\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).toContain($2)'
  );
  r = r.replace(
    /assert\.match\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).toMatch($2)'
  );
  r = r.replace(
    /assert\.instanceOf\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).toBeInstanceOf($2)'
  );
  r = r.replace(
    /assert\.lengthOf\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).toHaveLength($2)'
  );
  r = r.replace(
    /assert\.property\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).toHaveProperty($2)'
  );

  // Jasmine asymmetric matchers (may appear in assertion args)
  r = r.replace(/\bjasmine\.any\(([^)]+)\)/g, 'expect.any($1)');
  r = r.replace(/\bjasmine\.anything\(\)/g, 'expect.anything()');
  r = r.replace(/\bjasmine\.objectContaining\(/g, 'expect.objectContaining(');
  r = r.replace(/\bjasmine\.arrayContaining\(/g, 'expect.arrayContaining(');
  r = r.replace(/\bjasmine\.stringMatching\(/g, 'expect.stringMatching(');
  r = r.replace(/\bjasmine\.stringContaining\(/g, 'expect.stringContaining(');

  return r;
}

/**
 * Transform mock call lines: Sinon/Jasmine → Jest mocking.
 */
function transformMockToJest(line) {
  let r = line;

  // sinon.stub() → jest.fn()
  r = r.replace(/\bsinon\.stub\(\)/g, 'jest.fn()');
  r = r.replace(/\bsinon\.stub\(([^,]+),\s*([^)]+)\)/g, 'jest.spyOn($1, $2)');
  r = r.replace(/\bsinon\.spy\(([^,]+),\s*([^)]+)\)/g, 'jest.spyOn($1, $2)');
  r = r.replace(/\bsinon\.spy\(\)/g, 'jest.fn()');

  // Chain methods
  r = r.replace(/\.returns\(([^)]+)\)/g, '.mockReturnValue($1)');
  r = r.replace(/\.callsFake\(([^)]*)\)/g, '.mockImplementation($1)');
  r = r.replace(/\.resolves\(([^)]+)\)/g, '.mockResolvedValue($1)');
  r = r.replace(/\.rejects\(([^)]+)\)/g, '.mockRejectedValue($1)');
  r = r.replace(
    /\.onFirstCall\(\)\.mockReturnValue\(([^)]+)\)/g,
    '.mockReturnValueOnce($1)'
  );
  r = r.replace(/\.resetHistory\(\)/g, '.mockClear()');

  // sinon.assert → expect assertions
  r = r.replace(
    /sinon\.assert\.calledOnce\(([^)]+)\)/g,
    'expect($1).toHaveBeenCalledTimes(1)'
  );
  r = r.replace(
    /sinon\.assert\.calledTwice\(([^)]+)\)/g,
    'expect($1).toHaveBeenCalledTimes(2)'
  );
  r = r.replace(
    /sinon\.assert\.called\(([^)]+)\)/g,
    'expect($1).toHaveBeenCalled()'
  );
  r = r.replace(
    /sinon\.assert\.notCalled\(([^)]+)\)/g,
    'expect($1).not.toHaveBeenCalled()'
  );
  r = r.replace(
    /sinon\.assert\.calledWith\(([^,]+),\s*([^)]*)\)/g,
    'expect($1).toHaveBeenCalledWith($2)'
  );
  r = r.replace(
    /sinon\.assert\.callCount\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).toHaveBeenCalledTimes($2)'
  );

  // Timer fakes
  r = r.replace(/\bsinon\.useFakeTimers\(\)/g, 'jest.useFakeTimers()');
  r = r.replace(/\bclock\.tick\(([^)]+)\)/g, 'jest.advanceTimersByTime($1)');
  r = r.replace(/\bclock\.restore\(\)/g, 'jest.useRealTimers()');
  r = r.replace(/\bsinon\.restore\(\)/g, 'jest.restoreAllMocks()');
  r = r.replace(/\bsandbox\.restore\(\)/g, 'jest.restoreAllMocks()');

  // Remove sandbox creation
  r = r.replace(/const\s+\w+\s*=\s*sinon\.createSandbox\(\)\s*;?/g, '');

  // .callCount → .mock.calls.length
  r = r.replace(/\.callCount\b/g, '.mock.calls.length');

  // Unconvertible sinon patterns
  if (/sinon\.fakeServer/.test(r)) {
    r = r.replace(/sinon\.fakeServer\.[^\n]+/g, (match) => {
      return (
        formatter.formatTodo({
          id: 'UNCONVERTIBLE-FAKE-SERVER',
          description: 'sinon.fakeServer has no direct Jest equivalent',
          original: match.trim(),
          action: 'Use msw (Mock Service Worker) or jest-fetch-mock instead',
        }) +
        '\n// ' +
        match.trim()
      );
    });
  }

  // Jasmine spy API
  r = r.replace(
    /jasmine\.createSpy\(\)\.and\.callFake\(([^)]+)\)/g,
    'jest.fn($1)'
  );
  r = r.replace(
    /jasmine\.createSpy\s*\(\s*(['"][^'"]*['"]\s*)?\)/g,
    'jest.fn()'
  );
  r = r.replace(
    /jasmine\.createSpyObj\s*\(\s*['"][^'"]*['"]\s*,\s*\[([^\]]+)\]\s*\)/g,
    (_match, methods) => {
      const names = methods
        .split(',')
        .map((m) => m.trim().replace(/['"]/g, ''));
      const props = names.map((n) => `${n}: jest.fn()`).join(', ');
      return `{ ${props} }`;
    }
  );
  r = r.replace(/(?<!\.)spyOn\(([^,]+),\s*([^)]+)\)/g, 'jest.spyOn($1, $2)');
  r = r.replace(/\.and\.returnValue\(([^)]+)\)/g, '.mockReturnValue($1)');
  r = r.replace(/\.and\.callFake\(([^)]*)\)/g, '.mockImplementation($1)');
  r = r.replace(
    /\.and\.throwError\(([^)]+)\)/g,
    '.mockImplementation(() => { throw $1; })'
  );
  r = r.replace(/\.and\.callThrough\(\)/g, '');
  r = r.replace(/\.calls\.count\(\)/g, '.mock.calls.length');
  r = r.replace(/\.calls\.argsFor\((\d+)\)/g, '.mock.calls[$1]');
  r = r.replace(/\.calls\.mostRecent\(\)\.args/g, '.mock.lastCall');
  r = r.replace(/\.calls\.mostRecent\(\)/g, '.mock.lastCall');
  r = r.replace(/\.calls\.reset\(\)/g, '.mockClear()');
  r = r.replace(/\.calls\.allArgs\(\)/g, '.mock.calls');
  r = r.replace(/\.calls\.first\(\)\.args/g, '.mock.calls[0]');

  // Jasmine clock
  r = r.replace(/jasmine\.clock\(\)\.install\(\)/g, 'jest.useFakeTimers()');
  r = r.replace(
    /jasmine\.clock\(\)\.tick\(([^)]+)\)/g,
    'jest.advanceTimersByTime($1)'
  );
  r = r.replace(/jasmine\.clock\(\)\.uninstall\(\)/g, 'jest.useRealTimers()');
  r = r.replace(
    /jasmine\.clock\(\)\.mockDate\(([^)]+)\)/g,
    'jest.setSystemTime($1)'
  );

  // Mocha this.timeout/this.retries/this.slow (classified as MockCall)
  if (/this\.timeout/.test(r)) {
    r = r.replace(/this\.timeout\s*\(\s*(\d+)\s*\)\s*;?/g, (match, ms) => {
      return (
        formatter.formatWarning({
          description: `Mocha this.timeout(${ms}) — use jest.setTimeout(${ms}) at top of file or per-test timeout option`,
          original: match.trim(),
        }) +
        '\n// ' +
        match.trim()
      );
    });
  }
  if (/this\.retries/.test(r)) {
    r = r.replace(/this\.retries\s*\(\s*(\d+)\s*\)\s*;?/g, (match, count) => {
      return (
        formatter.formatWarning({
          description: `Mocha this.retries(${count}) — use jest.retryTimes(${count}) at top of describe block`,
          original: match.trim(),
        }) +
        '\n// ' +
        match.trim()
      );
    });
  }
  if (/this\.slow/.test(r)) {
    r = r.replace(/this\.slow\s*\(\s*(\d+)\s*\)\s*;?/g, (match, ms) => {
      return (
        formatter.formatWarning({
          description: `Mocha this.slow(${ms}) has no Jest equivalent — remove or use a custom reporter`,
          original: match.trim(),
        }) +
        '\n// ' +
        match.trim()
      );
    });
  }

  // Unconvertible Jasmine patterns
  if (/jasmine\.addMatchers/.test(r)) {
    r = r.replace(/jasmine\.addMatchers\s*\([^)]*\)/g, (match) => {
      return (
        formatter.formatTodo({
          id: 'UNCONVERTIBLE-CUSTOM-MATCHER',
          description:
            'Jasmine custom matchers must be converted to expect.extend() in Jest',
          original: match.trim(),
          action: 'Rewrite custom matchers using expect.extend()',
        }) +
        '\n// ' +
        match.trim()
      );
    });
  }

  return r;
}

/**
 * Transform RawCode lines: conservative chain transforms for
 * multi-line assertion/mock continuations and unconvertible patterns.
 */
function transformRawCodeToJest(line) {
  let r = line;

  // Chai chain transforms (continuation lines from multi-line assertions)
  r = r.replace(/\.to\.not\./g, '.not.');
  r = r.replace(/\.not\.to\./g, '.not.');
  r = r.replace(/\.to\.deep\.equal\(/g, '.toEqual(');
  r = r.replace(/\.not\.deep\.equal\(/g, '.not.toEqual(');
  r = r.replace(/\.to\.equal\(/g, '.toBe(');
  r = r.replace(/\.not\.equal\(/g, '.not.toBe(');
  r = r.replace(/\.to\.be\.true/g, '.toBe(true)');
  r = r.replace(/\.not\.be\.true/g, '.not.toBe(true)');
  r = r.replace(/\.to\.be\.false/g, '.toBe(false)');
  r = r.replace(/\.not\.be\.false/g, '.not.toBe(false)');
  r = r.replace(/\.to\.be\.null/g, '.toBeNull()');
  r = r.replace(/\.to\.be\.undefined/g, '.toBeUndefined()');
  r = r.replace(/\.to\.be\.NaN/g, '.toBeNaN()');
  r = r.replace(/\.to\.be\.ok/g, '.toBeTruthy()');
  r = r.replace(/\.not\.be\.ok/g, '.toBeFalsy()');
  r = r.replace(/\.to\.be\.an?\.instanceOf\(/g, '.toBeInstanceOf(');
  r = r.replace(/\.to\.have\.lengthOf\(/g, '.toHaveLength(');
  r = r.replace(/\.to\.include\(/g, '.toContain(');
  r = r.replace(/\.to\.contain\(/g, '.toContain(');
  r = r.replace(/\.to\.have\.property\(/g, '.toHaveProperty(');
  r = r.replace(/\.to\.be\.above\(/g, '.toBeGreaterThan(');
  r = r.replace(/\.to\.be\.below\(/g, '.toBeLessThan(');
  r = r.replace(/\.to\.be\.at\.least\(/g, '.toBeGreaterThanOrEqual(');
  r = r.replace(/\.to\.be\.at\.most\(/g, '.toBeLessThanOrEqual(');
  r = r.replace(/\.to\.be\.closeTo\(/g, '.toBeCloseTo(');
  r = r.replace(/\.to\.throw\(/g, '.toThrow(');
  r = r.replace(/\.to\.throw(?!\w)/g, '.toThrow()');
  r = r.replace(/\.to\.match\(/g, '.toMatch(');
  r = r.replace(/\.to\.exist/g, '.toBeDefined()');
  r = r.replace(/\.not\.exist/g, '.not.toBeDefined()');
  r = r.replace(/\.to\.have\.been\.calledWith\(/g, '.toHaveBeenCalledWith(');
  r = r.replace(/\.to\.have\.been\.called(?!\w)/g, '.toHaveBeenCalled()');

  // sinon.assert on raw/continuation lines (MUST run before .callCount)
  r = r.replace(
    /sinon\.assert\.calledOnce\(([^)]+)\)/g,
    'expect($1).toHaveBeenCalledTimes(1)'
  );
  r = r.replace(
    /sinon\.assert\.calledTwice\(([^)]+)\)/g,
    'expect($1).toHaveBeenCalledTimes(2)'
  );
  r = r.replace(
    /sinon\.assert\.calledWith\(([^,]+),\s*([^)]*)\)/g,
    'expect($1).toHaveBeenCalledWith($2)'
  );
  r = r.replace(
    /sinon\.assert\.notCalled\(([^)]+)\)/g,
    'expect($1).not.toHaveBeenCalled()'
  );
  r = r.replace(
    /sinon\.assert\.callCount\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).toHaveBeenCalledTimes($2)'
  );
  r = r.replace(
    /sinon\.assert\.called\(([^)]+)\)/g,
    'expect($1).toHaveBeenCalled()'
  );

  // Mock chain continuations
  r = r.replace(/\.returns\(([^)]+)\)/g, '.mockReturnValue($1)');
  r = r.replace(/\.callsFake\(([^)]*)\)/g, '.mockImplementation($1)');
  r = r.replace(/\.resolves\(([^)]+)\)/g, '.mockResolvedValue($1)');
  r = r.replace(/\.rejects\(([^)]+)\)/g, '.mockRejectedValue($1)');
  r = r.replace(/\.resetHistory\(\)/g, '.mockClear()');
  r = r.replace(/\.and\.returnValue\(([^)]+)\)/g, '.mockReturnValue($1)');
  r = r.replace(/\.and\.callFake\(([^)]*)\)/g, '.mockImplementation($1)');
  r = r.replace(/\.and\.callThrough\(\)/g, '');
  r = r.replace(/\.calls\.count\(\)/g, '.mock.calls.length');
  r = r.replace(/\.calls\.argsFor\((\d+)\)/g, '.mock.calls[$1]');
  r = r.replace(/\.calls\.reset\(\)/g, '.mockClear()');
  r = r.replace(/\.callCount\b/g, '.mock.calls.length');

  // Standalone sinon calls on unrecognized lines
  r = r.replace(/\bsinon\.stub\(\)/g, 'jest.fn()');
  r = r.replace(/\bsinon\.stub\(([^,]+),\s*([^)]+)\)/g, 'jest.spyOn($1, $2)');
  r = r.replace(/\bsinon\.spy\(([^,]+),\s*([^)]+)\)/g, 'jest.spyOn($1, $2)');
  r = r.replace(/\bsinon\.spy\(\)/g, 'jest.fn()');
  r = r.replace(/\bsinon\.restore\(\)/g, 'jest.restoreAllMocks()');
  r = r.replace(/\bsandbox\.restore\(\)/g, 'jest.restoreAllMocks()');
  r = r.replace(/\bclock\.tick\(([^)]+)\)/g, 'jest.advanceTimersByTime($1)');
  r = r.replace(/\bclock\.restore\(\)/g, 'jest.useRealTimers()');

  // Unconvertible patterns on raw lines
  if (/chai\.use/.test(r)) {
    r = r.replace(/chai\.use\s*\([^)]+\)\s*;?/g, (match) => {
      return (
        formatter.formatTodo({
          id: 'UNCONVERTIBLE-CHAI-PLUGIN',
          description: 'Chai plugin not available in Jest',
          original: match.trim(),
          action:
            'Find a Jest-compatible alternative or implement custom matchers with expect.extend()',
        }) +
        '\n// ' +
        match.trim()
      );
    });
  }
  if (/this\.timeout/.test(r)) {
    r = r.replace(/this\.timeout\s*\(\s*(\d+)\s*\)\s*;?/g, (match, ms) => {
      return (
        formatter.formatWarning({
          description: `Mocha this.timeout(${ms}) — use jest.setTimeout(${ms}) at top of file or per-test timeout option`,
          original: match.trim(),
        }) +
        '\n// ' +
        match.trim()
      );
    });
  }
  if (/this\.retries/.test(r)) {
    r = r.replace(/this\.retries\s*\(\s*(\d+)\s*\)\s*;?/g, (match, count) => {
      return (
        formatter.formatWarning({
          description: `Mocha this.retries(${count}) — use jest.retryTimes(${count}) at top of describe block`,
          original: match.trim(),
        }) +
        '\n// ' +
        match.trim()
      );
    });
  }
  if (/this\.slow/.test(r)) {
    r = r.replace(/this\.slow\s*\(\s*(\d+)\s*\)\s*;?/g, (match, ms) => {
      return (
        formatter.formatWarning({
          description: `Mocha this.slow(${ms}) has no Jest equivalent — remove or use a custom reporter`,
          original: match.trim(),
        }) +
        '\n// ' +
        match.trim()
      );
    });
  }
  if (/\bpending\s*\(/.test(r)) {
    r = r.replace(/\bpending\s*\(\s*(['"][^'"]*['"])?\s*\)\s*;?/g, (match) => {
      return (
        formatter.formatTodo({
          id: 'UNCONVERTIBLE-PENDING',
          description: 'Jasmine pending() — use it.todo() or it.skip() in Jest',
          original: match.trim(),
          action: 'Replace with it.todo() for the containing test',
        }) +
        '\n// ' +
        match.trim()
      );
    });
  }
  if (/\bexpectAsync\s*\(/.test(r)) {
    r = r.replace(/\bexpectAsync\s*\([^)]+\)/g, (match) => {
      return (
        formatter.formatTodo({
          id: 'UNCONVERTIBLE-EXPECT-ASYNC',
          description:
            'Jasmine expectAsync() — use await expect(...).resolves/rejects in Jest',
          original: match.trim(),
          action: 'Rewrite using expect(...).resolves or expect(...).rejects',
        }) +
        '\n// ' +
        match.trim()
      );
    });
  }

  // Jasmine spy .calls API on raw lines
  r = r.replace(/\.calls\.argsFor\((\d+)\)/g, '.mock.calls[$1]');
  r = r.replace(/\.calls\.mostRecent\(\)\.args/g, '.mock.lastCall');
  r = r.replace(/\.calls\.mostRecent\(\)/g, '.mock.lastCall');
  r = r.replace(/\.calls\.first\(\)\.args/g, '.mock.calls[0]');
  r = r.replace(/\.calls\.allArgs\(\)/g, '.mock.calls');

  // Unconvertible Jasmine patterns on raw lines
  if (/jasmine\.addMatchers/.test(r)) {
    r = r.replace(/jasmine\.addMatchers\s*\([^)]*\)/g, (match) => {
      return (
        formatter.formatTodo({
          id: 'UNCONVERTIBLE-CUSTOM-MATCHER',
          description:
            'Jasmine custom matchers must be converted to expect.extend() in Jest',
          original: match.trim(),
          action: 'Rewrite custom matchers using expect.extend()',
        }) +
        '\n// ' +
        match.trim()
      );
    });
  }

  // Jasmine asymmetric matchers (may appear in raw code args)
  r = r.replace(/\bjasmine\.any\(([^)]+)\)/g, 'expect.any($1)');
  r = r.replace(/\bjasmine\.anything\(\)/g, 'expect.anything()');
  r = r.replace(/\bjasmine\.objectContaining\(/g, 'expect.objectContaining(');
  r = r.replace(/\bjasmine\.arrayContaining\(/g, 'expect.arrayContaining(');
  r = r.replace(/\bjasmine\.stringMatching\(/g, 'expect.stringMatching(');
  r = r.replace(/\bjasmine\.stringContaining\(/g, 'expect.stringContaining(');

  return r;
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
    globals: [
      'describe',
      'it',
      'test',
      'expect',
      'beforeEach',
      'afterEach',
      'beforeAll',
      'afterAll',
    ],
    mockNamespace: 'jest',
  },
};
