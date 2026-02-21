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
 * Parse Jest source code into a nested IR tree.
 *
 * Uses brace-depth tracking to nest TestSuite, TestCase, and Hook nodes:
 *   TestSuite.hooks[] / TestSuite.tests[]
 *   TestCase.body[] / Hook.body[]
 *
 * Extracts subject and expected from expect() assertions.
 */
function parse(source) {
  const lines = source.split('\n');
  const imports = [];

  // Root container — collects top-level nodes
  const rootBody = [];
  // Stack of { node, depth } for nesting containers
  const stack = [{ node: null, addChild: (c) => rootBody.push(c), depth: 0 }];
  let depth = 0;

  function addChild(child) {
    const top = stack[stack.length - 1];
    const node = top.node;

    if (!node) {
      top.addChild(child);
    } else if (node instanceof TestSuite) {
      if (child instanceof Hook) node.hooks.push(child);
      else node.tests.push(child);
    } else if (node instanceof TestCase || node instanceof Hook) {
      node.body.push(child);
    }
  }

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const trimmed = line.trim();
    const loc = { line: i + 1, column: 0 };

    // Skip empty lines
    if (!trimmed) continue;

    // Count braces outside strings and comments
    const opens = countBraces(trimmed, '{');
    const closes = countBraces(trimmed, '}');
    const newDepth = depth + opens - closes;

    // Pop containers whose blocks have closed
    while (stack.length > 1 && newDepth <= stack[stack.length - 1].depth) {
      stack.pop();
    }

    // Pure structural lines (closing braces only) — skip
    if (/^[}\s);,]+$/.test(trimmed)) {
      depth = newDepth;
      continue;
    }

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
      addChild(
        new Comment({
          text: line,
          commentKind: isLicense
            ? 'license'
            : isDirective
              ? 'directive'
              : 'inline',
          preserveExact: isDirective || isLicense,
          sourceLocation: loc,
          originalSource: line,
        })
      );
      depth = newDepth;
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
      depth = newDepth;
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
      addChild(
        new MockCall({
          kind: hasVirtual
            ? 'mockModule'
            : hasFactory
              ? 'mockModule'
              : 'mockModule',
          target:
            (trimmed.match(/jest\.mock\s*\(\s*['"]([^'"]+)['"]/) || [])[1] ||
            '',
          sourceLocation: loc,
          originalSource: fullMock,
          confidence: hasVirtual ? 'unconvertible' : 'converted',
        })
      );
      depth = newDepth;
      continue;
    }

    // jest.setTimeout / jest.retryTimes at top level
    if (/^jest\.setTimeout\s*\(/.test(trimmed)) {
      addChild(
        new MockCall({
          kind: 'fakeTimers',
          target: 'setTimeout',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      depth = newDepth;
      continue;
    }
    if (/^jest\.retryTimes\s*\(/.test(trimmed)) {
      addChild(
        new MockCall({
          kind: 'fakeTimers',
          target: 'retryTimes',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'warning',
        })
      );
      depth = newDepth;
      continue;
    }

    // describe block — extract name and push container
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
      const suite = new TestSuite({
        name:
          (trimmed.match(/describe(?:\.\w+)*\s*\(\s*['"`]([^'"`]*)['"`]/) ||
            [])[1] || '',
        modifiers,
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      });
      addChild(suite);
      if (opens > closes) {
        stack.push({ node: suite, depth });
      }
      depth = newDepth;
      continue;
    }

    // it/test block — extract name and push container
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
      const tc = new TestCase({
        name:
          (trimmed.match(/(?:it|test)(?:\.\w+)*\s*\(\s*['"`]([^'"`]*)['"`]/) ||
            [])[1] || '',
        isAsync,
        modifiers,
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      });
      addChild(tc);
      if (opens > closes) {
        stack.push({ node: tc, depth });
      }
      depth = newDepth;
      continue;
    }

    // Hook block — extract hookType and push container
    if (/\b(?:beforeEach|afterEach|beforeAll|afterAll)\s*\(/.test(trimmed)) {
      const hookType = (trimmed.match(
        /\b(beforeEach|afterEach|beforeAll|afterAll)/
      ) || [])[1];
      const hook = new Hook({
        hookType,
        isAsync: /async/.test(trimmed),
        sourceLocation: loc,
        originalSource: line,
        confidence: 'converted',
      });
      addChild(hook);
      if (opens > closes) {
        stack.push({ node: hook, depth });
      }
      depth = newDepth;
      continue;
    }

    // Assertions: expect(...) — extract subject and expected
    if (/\bexpect\s*\(/.test(trimmed)) {
      const parsed = parseJestAssertionLine(trimmed);
      addChild(
        new Assertion({
          ...parsed,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      depth = newDepth;
      continue;
    }

    // jest.fn / jest.spyOn / jest.mock inline
    if (/jest\.fn\s*\(/.test(trimmed) || /jest\.spyOn\s*\(/.test(trimmed)) {
      const kind = /jest\.fn/.test(trimmed) ? 'createMock' : 'spyOnMethod';
      addChild(
        new MockCall({
          kind,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      depth = newDepth;
      continue;
    }

    // jest.useFakeTimers / jest.useRealTimers / jest.advanceTimersByTime / jest.runAllTimers
    if (/jest\.\w+/.test(trimmed)) {
      addChild(
        new MockCall({
          kind: 'fakeTimers',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      depth = newDepth;
      continue;
    }

    // expect.extend, expect.assertions, expect.addSnapshotSerializer
    if (/\bexpect\.extend\s*\(/.test(trimmed)) {
      addChild(
        new RawCode({
          code: line,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      depth = newDepth;
      continue;
    }
    if (/\bexpect\.addSnapshotSerializer\s*\(/.test(trimmed)) {
      addChild(
        new RawCode({
          code: line,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'unconvertible',
        })
      );
      depth = newDepth;
      continue;
    }

    // Everything else is raw code
    addChild(
      new RawCode({
        code: line,
        sourceLocation: loc,
        originalSource: line,
      })
    );
    depth = newDepth;
  }

  return new TestFile({
    language: 'javascript',
    imports,
    body: rootBody,
  });
}

/**
 * Emit Jest code from IR + original source.
 *
 * Converts source code to Jest by applying framework-specific pattern
 * transforms. Handles Mocha+Chai→Jest, Jasmine→Jest, and other X→Jest
 * conversions. Each phase only matches patterns from its source framework,
 * so all phases can safely run on every input.
 *
 * @param {TestFile} _ir - Parsed IR tree (for scoring metadata)
 * @param {string} source - Original source code
 * @returns {string} Converted Jest source code
 */
function emit(_ir, source) {
  let result = source;

  // --- Phase 1: Remove source-framework imports ---

  // Remove chai imports
  result = result.replace(
    /const\s+\{[^}]*\}\s*=\s*require\s*\(\s*['"]chai['"]\s*\)\s*;?\n?/g,
    ''
  );
  result = result.replace(
    /const\s+\w+\s*=\s*require\s*\(\s*['"]chai['"]\s*\)\s*;?\n?/g,
    ''
  );
  result = result.replace(/import\s+\{[^}]*\}\s+from\s+['"]chai['"];?\n?/g, '');
  result = result.replace(/import\s+\w+\s+from\s+['"]chai['"];?\n?/g, '');

  // Remove sinon imports
  result = result.replace(
    /const\s+\w+\s*=\s*require\s*\(\s*['"]sinon['"]\s*\)\s*;?\n?/g,
    ''
  );
  result = result.replace(/import\s+\w+\s+from\s+['"]sinon['"];?\n?/g, '');
  result = result.replace(
    /import\s+\{[^}]*\}\s+from\s+['"]sinon['"];?\n?/g,
    ''
  );

  // Remove chai-as-promised, sinon-chai, etc.
  result = result.replace(
    /const\s+\w+\s*=\s*require\s*\(\s*['"]chai-as-promised['"]\s*\)\s*;?\n?/g,
    ''
  );
  result = result.replace(
    /const\s+\w+\s*=\s*require\s*\(\s*['"]sinon-chai['"]\s*\)\s*;?\n?/g,
    ''
  );
  result = result.replace(
    /import\s+\w+\s+from\s+['"]chai-as-promised['"];?\n?/g,
    ''
  );
  result = result.replace(/import\s+\w+\s+from\s+['"]sinon-chai['"];?\n?/g, '');

  // --- Phase 2: Convert Mocha aliases ---

  // context( → describe(
  result = result.replace(/\bcontext\s*\(/g, 'describe(');

  // specify( → it(
  result = result.replace(/\bspecify\s*\(/g, 'it(');

  // before( → beforeAll( (only bare before, not beforeEach)
  result = result.replace(/\bbefore\s*\((?!Each|All)/g, 'beforeAll(');

  // after( → afterAll( (only bare after, not afterEach)
  result = result.replace(/\bafter\s*\((?!Each|All)/g, 'afterAll(');

  // --- Phase 3: Convert Chai expect chains → Jest expect ---
  // Uses chain-suffix replacement to avoid nested-paren issues with expect(fn())
  // Step 1: Normalize negation (.to.not. and .not.to. → .not.)
  // Step 2: Replace Chai chains with Jest matchers

  // Negation normalization (must come before chain replacements)
  result = result.replace(/\.to\.not\./g, '.not.');
  result = result.replace(/\.not\.to\./g, '.not.');

  // Chai-sinon assertion chains (most specific first)
  result = result.replace(
    /\.to\.have\.been\.calledOnce/g,
    '.toHaveBeenCalledTimes(1)'
  );
  result = result.replace(
    /\.to\.have\.been\.calledTwice/g,
    '.toHaveBeenCalledTimes(2)'
  );
  result = result.replace(
    /\.to\.have\.been\.calledWith\(/g,
    '.toHaveBeenCalledWith('
  );
  result = result.replace(
    /\.to\.have\.been\.called(?!\w)/g,
    '.toHaveBeenCalled()'
  );
  result = result.replace(
    /\.not\.have\.been\.called(?!\w)/g,
    '.not.toHaveBeenCalled()'
  );

  // Deep equal before equal (more specific first)
  result = result.replace(/\.to\.deep\.equal\(/g, '.toEqual(');
  result = result.replace(/\.not\.deep\.equal\(/g, '.not.toEqual(');

  // .to.equal → .toBe
  result = result.replace(/\.to\.equal\(/g, '.toBe(');
  result = result.replace(/\.not\.equal\(/g, '.not.toBe(');

  // Boolean/null/undefined terminal chains
  result = result.replace(/\.to\.be\.true/g, '.toBe(true)');
  result = result.replace(/\.not\.be\.true/g, '.not.toBe(true)');
  result = result.replace(/\.to\.be\.false/g, '.toBe(false)');
  result = result.replace(/\.not\.be\.false/g, '.not.toBe(false)');
  result = result.replace(/\.to\.be\.null/g, '.toBeNull()');
  result = result.replace(/\.not\.be\.null/g, '.not.toBeNull()');
  result = result.replace(/\.to\.be\.undefined/g, '.toBeUndefined()');
  result = result.replace(/\.not\.be\.undefined/g, '.not.toBeUndefined()');
  result = result.replace(/\.to\.be\.NaN/g, '.toBeNaN()');
  result = result.replace(/\.not\.be\.NaN/g, '.not.toBeNaN()');
  result = result.replace(/\.to\.be\.ok/g, '.toBeTruthy()');
  result = result.replace(/\.not\.be\.ok/g, '.toBeFalsy()');

  // .to.be.an.instanceOf(Y) → .toBeInstanceOf(Y) (before .to.be.a)
  result = result.replace(/\.to\.be\.an?\.instanceOf\(/g, '.toBeInstanceOf(');
  result = result.replace(
    /\.not\.be\.an?\.instanceOf\(/g,
    '.not.toBeInstanceOf('
  );

  // .to.be.a('type') → needs subject restructuring, use paren-safe regex
  const SUBJ = '([^)]*(?:\\([^)]*\\)[^)]*)*)';
  const beATypeRe = new RegExp(
    'expect\\(' + SUBJ + '\\)\\.to\\.be\\.a\\(([\'"][^\'"]+[\'"])\\)',
    'g'
  );
  result = result.replace(beATypeRe, 'expect(typeof $1).toBe($2)');

  // .to.be.an('array') → special case
  const beAnArrayRe = new RegExp(
    'expect\\(' + SUBJ + '\\)\\.to\\.be\\.an\\(([\'"]array[\'"])\\)',
    'g'
  );
  result = result.replace(beAnArrayRe, 'expect(Array.isArray($1)).toBe(true)');

  // .to.be.an('type') → typeof
  const beAnTypeRe = new RegExp(
    'expect\\(' + SUBJ + '\\)\\.to\\.be\\.an\\(([\'"][^\'"]+[\'"])\\)',
    'g'
  );
  result = result.replace(beAnTypeRe, 'expect(typeof $1).toBe($2)');

  // Chain replacements with args
  result = result.replace(/\.to\.have\.lengthOf\(/g, '.toHaveLength(');
  result = result.replace(/\.not\.have\.lengthOf\(/g, '.not.toHaveLength(');
  result = result.replace(/\.to\.include\(/g, '.toContain(');
  result = result.replace(/\.not\.include\(/g, '.not.toContain(');
  result = result.replace(/\.to\.contain\(/g, '.toContain(');
  result = result.replace(/\.not\.contain\(/g, '.not.toContain(');
  result = result.replace(/\.to\.have\.property\(/g, '.toHaveProperty(');
  result = result.replace(/\.not\.have\.property\(/g, '.not.toHaveProperty(');
  result = result.replace(/\.to\.be\.above\(/g, '.toBeGreaterThan(');
  result = result.replace(/\.to\.be\.below\(/g, '.toBeLessThan(');
  result = result.replace(/\.to\.be\.at\.least\(/g, '.toBeGreaterThanOrEqual(');
  result = result.replace(/\.to\.be\.at\.most\(/g, '.toBeLessThanOrEqual(');
  result = result.replace(/\.to\.be\.closeTo\(/g, '.toBeCloseTo(');
  result = result.replace(/\.to\.throw\(/g, '.toThrow(');
  result = result.replace(/\.not\.throw\(/g, '.not.toThrow(');
  result = result.replace(/\.to\.throw(?!\w)/g, '.toThrow()');
  result = result.replace(/\.not\.throw(?!\w)/g, '.not.toThrow()');
  result = result.replace(/\.to\.match\(/g, '.toMatch(');
  result = result.replace(/\.not\.match\(/g, '.not.toMatch(');

  // Terminal chains
  result = result.replace(/\.to\.exist/g, '.toBeDefined()');
  result = result.replace(/\.not\.exist/g, '.not.toBeDefined()');

  // --- Phase 4: Convert Chai assert-style ---

  // assert.strictEqual(a, b) → expect(a).toBe(b)
  result = result.replace(
    /assert\.strictEqual\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).toBe($2)'
  );

  // assert.deepEqual(a, b) → expect(a).toEqual(b)
  result = result.replace(
    /assert\.deepEqual\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).toEqual($2)'
  );

  // assert.equal(a, b) → expect(a).toBe(b)
  result = result.replace(
    /assert\.equal\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).toBe($2)'
  );

  // assert.notEqual(a, b) → expect(a).not.toBe(b)
  result = result.replace(
    /assert\.notEqual\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).not.toBe($2)'
  );

  // assert.isTrue(x) → expect(x).toBe(true)
  result = result.replace(
    /assert\.isTrue\(([^)]+)\)/g,
    'expect($1).toBe(true)'
  );

  // assert.isFalse(x) → expect(x).toBe(false)
  result = result.replace(
    /assert\.isFalse\(([^)]+)\)/g,
    'expect($1).toBe(false)'
  );

  // assert.isNull(x) → expect(x).toBeNull()
  result = result.replace(
    /assert\.isNull\(([^)]+)\)/g,
    'expect($1).toBeNull()'
  );

  // assert.isNotNull(x) → expect(x).not.toBeNull()
  result = result.replace(
    /assert\.isNotNull\(([^)]+)\)/g,
    'expect($1).not.toBeNull()'
  );

  // assert.isUndefined(x) → expect(x).toBeUndefined()
  result = result.replace(
    /assert\.isUndefined\(([^)]+)\)/g,
    'expect($1).toBeUndefined()'
  );

  // assert.isDefined(x) → expect(x).toBeDefined()
  result = result.replace(
    /assert\.isDefined\(([^)]+)\)/g,
    'expect($1).toBeDefined()'
  );

  // assert.isOk(x) → expect(x).toBeTruthy()
  result = result.replace(
    /assert\.isOk\(([^)]+)\)/g,
    'expect($1).toBeTruthy()'
  );

  // assert.isNotOk(x) → expect(x).toBeFalsy()
  result = result.replace(
    /assert\.isNotOk\(([^)]+)\)/g,
    'expect($1).toBeFalsy()'
  );

  // assert.throws(fn) → expect(fn).toThrow()
  result = result.replace(/assert\.throws\(([^)]+)\)/g, 'expect($1).toThrow()');

  // assert.doesNotThrow(fn) → expect(fn).not.toThrow()
  result = result.replace(
    /assert\.doesNotThrow\(([^)]+)\)/g,
    'expect($1).not.toThrow()'
  );

  // assert.include(haystack, needle) → expect(haystack).toContain(needle)
  result = result.replace(
    /assert\.include\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).toContain($2)'
  );

  // assert.match(str, re) → expect(str).toMatch(re)
  result = result.replace(
    /assert\.match\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).toMatch($2)'
  );

  // assert.instanceOf(obj, cls) → expect(obj).toBeInstanceOf(cls)
  result = result.replace(
    /assert\.instanceOf\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).toBeInstanceOf($2)'
  );

  // assert.lengthOf(obj, n) → expect(obj).toHaveLength(n)
  result = result.replace(
    /assert\.lengthOf\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).toHaveLength($2)'
  );

  // assert.property(obj, prop) → expect(obj).toHaveProperty(prop)
  result = result.replace(
    /assert\.property\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).toHaveProperty($2)'
  );

  // --- Phase 5: Convert Sinon → Jest mocking ---

  // sinon.stub() → jest.fn()
  result = result.replace(/\bsinon\.stub\(\)/g, 'jest.fn()');

  // sinon.stub(obj, 'method') → jest.spyOn(obj, 'method')
  result = result.replace(
    /\bsinon\.stub\(([^,]+),\s*([^)]+)\)/g,
    'jest.spyOn($1, $2)'
  );

  // sinon.spy(obj, 'method') → jest.spyOn(obj, 'method')
  result = result.replace(
    /\bsinon\.spy\(([^,]+),\s*([^)]+)\)/g,
    'jest.spyOn($1, $2)'
  );

  // sinon.spy() → jest.fn()
  result = result.replace(/\bsinon\.spy\(\)/g, 'jest.fn()');

  // .returns(val) → .mockReturnValue(val)
  result = result.replace(/\.returns\(([^)]+)\)/g, '.mockReturnValue($1)');

  // .callsFake(fn) → .mockImplementation(fn)
  result = result.replace(/\.callsFake\(([^)]*)\)/g, '.mockImplementation($1)');

  // .resolves(val) → .mockResolvedValue(val)
  result = result.replace(/\.resolves\(([^)]+)\)/g, '.mockResolvedValue($1)');

  // .rejects(err) → .mockRejectedValue(err)
  result = result.replace(/\.rejects\(([^)]+)\)/g, '.mockRejectedValue($1)');

  // .onFirstCall().returns(val) → .mockReturnValueOnce(val)
  result = result.replace(
    /\.onFirstCall\(\)\.mockReturnValue\(([^)]+)\)/g,
    '.mockReturnValueOnce($1)'
  );

  // .resetHistory() → .mockClear()
  result = result.replace(/\.resetHistory\(\)/g, '.mockClear()');

  // sinon.assert.calledOnce(fn) → expect(fn).toHaveBeenCalledTimes(1)
  result = result.replace(
    /sinon\.assert\.calledOnce\(([^)]+)\)/g,
    'expect($1).toHaveBeenCalledTimes(1)'
  );

  // sinon.assert.calledTwice(fn) → expect(fn).toHaveBeenCalledTimes(2)
  result = result.replace(
    /sinon\.assert\.calledTwice\(([^)]+)\)/g,
    'expect($1).toHaveBeenCalledTimes(2)'
  );

  // sinon.assert.called(fn) → expect(fn).toHaveBeenCalled()
  result = result.replace(
    /sinon\.assert\.called\(([^)]+)\)/g,
    'expect($1).toHaveBeenCalled()'
  );

  // sinon.assert.notCalled(fn) → expect(fn).not.toHaveBeenCalled()
  result = result.replace(
    /sinon\.assert\.notCalled\(([^)]+)\)/g,
    'expect($1).not.toHaveBeenCalled()'
  );

  // sinon.assert.calledWith(fn, args) → expect(fn).toHaveBeenCalledWith(args)
  result = result.replace(
    /sinon\.assert\.calledWith\(([^,]+),\s*([^)]*)\)/g,
    'expect($1).toHaveBeenCalledWith($2)'
  );

  // sinon.assert.callCount(fn, n) → expect(fn).toHaveBeenCalledTimes(n)
  result = result.replace(
    /sinon\.assert\.callCount\(([^,]+),\s*([^)]+)\)/g,
    'expect($1).toHaveBeenCalledTimes($2)'
  );

  // sinon.useFakeTimers() → jest.useFakeTimers()
  result = result.replace(
    /\bsinon\.useFakeTimers\(\)/g,
    'jest.useFakeTimers()'
  );

  // clock.tick(ms) → jest.advanceTimersByTime(ms)
  result = result.replace(
    /\bclock\.tick\(([^)]+)\)/g,
    'jest.advanceTimersByTime($1)'
  );

  // clock.restore() → jest.useRealTimers()
  result = result.replace(/\bclock\.restore\(\)/g, 'jest.useRealTimers()');

  // sinon.restore() → jest.restoreAllMocks()
  result = result.replace(/\bsinon\.restore\(\)/g, 'jest.restoreAllMocks()');

  // sandbox.restore() → jest.restoreAllMocks()
  result = result.replace(/\bsandbox\.restore\(\)/g, 'jest.restoreAllMocks()');

  // sinon.createSandbox() → (remove, not needed in Jest)
  result = result.replace(
    /const\s+\w+\s*=\s*sinon\.createSandbox\(\)\s*;?\n?/g,
    ''
  );

  // .callCount → .mock.calls.length
  result = result.replace(/\.callCount\b/g, '.mock.calls.length');

  // --- Phase 6: Unconvertible Mocha/Sinon patterns ---

  // chai.use(plugin) → HAMLET-TODO
  result = result.replace(/chai\.use\s*\([^)]+\)\s*;?/g, (match) => {
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

  // sinon.fakeServer → HAMLET-TODO
  result = result.replace(/sinon\.fakeServer\.[^\n]+/g, (match) => {
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

  // this.timeout() → HAMLET-WARNING
  result = result.replace(
    /this\.timeout\s*\(\s*(\d+)\s*\)\s*;?/g,
    (match, ms) => {
      return (
        formatter.formatWarning({
          description: `Mocha this.timeout(${ms}) — use jest.setTimeout(${ms}) at top of file or per-test timeout option`,
          original: match.trim(),
        }) +
        '\n// ' +
        match.trim()
      );
    }
  );

  // this.retries() → HAMLET-WARNING
  result = result.replace(
    /this\.retries\s*\(\s*(\d+)\s*\)\s*;?/g,
    (match, count) => {
      return (
        formatter.formatWarning({
          description: `Mocha this.retries(${count}) — use jest.retryTimes(${count}) at top of describe block`,
          original: match.trim(),
        }) +
        '\n// ' +
        match.trim()
      );
    }
  );

  // this.slow() → HAMLET-WARNING
  result = result.replace(/this\.slow\s*\(\s*(\d+)\s*\)\s*;?/g, (match, ms) => {
    return (
      formatter.formatWarning({
        description: `Mocha this.slow(${ms}) has no Jest equivalent — remove or use a custom reporter`,
        original: match.trim(),
      }) +
      '\n// ' +
      match.trim()
    );
  });

  // --- Phase 7: Convert Jasmine modifiers ---

  // fdescribe → describe.only
  result = result.replace(/\bfdescribe\s*\(/g, 'describe.only(');

  // xdescribe → describe.skip
  result = result.replace(/\bxdescribe\s*\(/g, 'describe.skip(');

  // fit → it.only
  result = result.replace(/\bfit\s*\(/g, 'it.only(');

  // xit → it.skip
  result = result.replace(/\bxit\s*\(/g, 'it.skip(');

  // pending() → HAMLET-TODO (Jasmine-specific)
  result = result.replace(
    /\bpending\s*\(\s*(['"][^'"]*['"])?\s*\)\s*;?/g,
    (match) => {
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
    }
  );

  // --- Phase 8: Convert Jasmine spy API ---

  // jasmine.createSpy('name') → jest.fn()
  result = result.replace(
    /jasmine\.createSpy\s*\(\s*(['"][^'"]*['"]\s*)?\)/g,
    'jest.fn()'
  );

  // jasmine.createSpyObj('name', ['a', 'b']) → { a: jest.fn(), b: jest.fn() }
  result = result.replace(
    /jasmine\.createSpyObj\s*\(\s*['"][^'"]*['"]\s*,\s*\[([^\]]+)\]\s*\)/g,
    (_match, methods) => {
      const names = methods
        .split(',')
        .map((m) => m.trim().replace(/['"]/g, ''));
      const props = names.map((n) => `${n}: jest.fn()`).join(', ');
      return `{ ${props} }`;
    }
  );

  // spyOn(obj, 'method') → jest.spyOn(obj, 'method') (only Jasmine bare spyOn)
  // Must not match jest.spyOn which was already converted
  result = result.replace(
    /(?<!\.)spyOn\(([^,]+),\s*([^)]+)\)/g,
    'jest.spyOn($1, $2)'
  );

  // .and.returnValue(val) → .mockReturnValue(val)
  result = result.replace(
    /\.and\.returnValue\(([^)]+)\)/g,
    '.mockReturnValue($1)'
  );

  // .and.callFake(fn) → .mockImplementation(fn)
  result = result.replace(
    /\.and\.callFake\(([^)]*)\)/g,
    '.mockImplementation($1)'
  );

  // .and.throwError(err) → .mockImplementation(() => { throw err; })
  result = result.replace(
    /\.and\.throwError\(([^)]+)\)/g,
    '.mockImplementation(() => { throw $1; })'
  );

  // .and.callThrough() → (remove — Jest default behavior)
  result = result.replace(/\.and\.callThrough\(\)/g, '');

  // .calls.count() → .mock.calls.length
  result = result.replace(/\.calls\.count\(\)/g, '.mock.calls.length');

  // .calls.argsFor(n) → .mock.calls[n]
  result = result.replace(/\.calls\.argsFor\((\d+)\)/g, '.mock.calls[$1]');

  // .calls.mostRecent() → .mock.lastCall
  result = result.replace(/\.calls\.mostRecent\(\)\.args/g, '.mock.lastCall');
  result = result.replace(/\.calls\.mostRecent\(\)/g, '.mock.lastCall');

  // .calls.reset() → .mockClear()
  result = result.replace(/\.calls\.reset\(\)/g, '.mockClear()');

  // .calls.allArgs() → .mock.calls
  result = result.replace(/\.calls\.allArgs\(\)/g, '.mock.calls');

  // .calls.first().args → .mock.calls[0]
  result = result.replace(/\.calls\.first\(\)\.args/g, '.mock.calls[0]');

  // --- Phase 9: Convert Jasmine asymmetric matchers ---

  // jasmine.any(X) → expect.any(X)
  result = result.replace(/\bjasmine\.any\(([^)]+)\)/g, 'expect.any($1)');

  // jasmine.anything() → expect.anything()
  result = result.replace(/\bjasmine\.anything\(\)/g, 'expect.anything()');

  // jasmine.objectContaining → expect.objectContaining
  result = result.replace(
    /\bjasmine\.objectContaining\(/g,
    'expect.objectContaining('
  );

  // jasmine.arrayContaining → expect.arrayContaining
  result = result.replace(
    /\bjasmine\.arrayContaining\(/g,
    'expect.arrayContaining('
  );

  // jasmine.stringMatching → expect.stringMatching
  result = result.replace(
    /\bjasmine\.stringMatching\(/g,
    'expect.stringMatching('
  );

  // jasmine.stringContaining → expect.stringContaining
  result = result.replace(
    /\bjasmine\.stringContaining\(/g,
    'expect.stringContaining('
  );

  // --- Phase 10: Convert Jasmine clock ---

  // jasmine.clock().install() → jest.useFakeTimers()
  result = result.replace(
    /jasmine\.clock\(\)\.install\(\)/g,
    'jest.useFakeTimers()'
  );

  // jasmine.clock().tick(ms) → jest.advanceTimersByTime(ms)
  result = result.replace(
    /jasmine\.clock\(\)\.tick\(([^)]+)\)/g,
    'jest.advanceTimersByTime($1)'
  );

  // jasmine.clock().uninstall() → jest.useRealTimers()
  result = result.replace(
    /jasmine\.clock\(\)\.uninstall\(\)/g,
    'jest.useRealTimers()'
  );

  // jasmine.clock().mockDate(d) → jest.setSystemTime(d)
  result = result.replace(
    /jasmine\.clock\(\)\.mockDate\(([^)]+)\)/g,
    'jest.setSystemTime($1)'
  );

  // --- Phase 11: Unconvertible Jasmine patterns ---

  // jasmine.addMatchers → HAMLET-TODO
  result = result.replace(/jasmine\.addMatchers\s*\([^)]*\)/g, (match) => {
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

  // expectAsync → HAMLET-TODO
  result = result.replace(/\bexpectAsync\s*\([^)]+\)/g, (match) => {
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

  // Clean up multiple blank lines
  result = result.replace(/\n{3,}/g, '\n\n');

  // Ensure trailing newline
  if (!result.endsWith('\n')) result += '\n';

  return result;
}

// ═══════════════════════════════════════════════════════════════════════
// Parser helpers
// ═══════════════════════════════════════════════════════════════════════

/**
 * Count occurrences of a character, skipping strings and line comments.
 */
function countBraces(line, ch) {
  let count = 0;
  let inString = false;
  let stringDelim = '';
  for (let i = 0; i < line.length; i++) {
    const c = line[i];
    if (inString) {
      if (c === stringDelim && line[i - 1] !== '\\') inString = false;
    } else {
      if (c === '/' && line[i + 1] === '/') break;
      if (c === "'" || c === '"' || c === '`') {
        inString = true;
        stringDelim = c;
      } else if (c === ch) {
        count++;
      }
    }
  }
  return count;
}

/**
 * Extract the subject from expect(subject) handling nested parens.
 * Returns { subject, rest } where rest is everything after the closing paren.
 */
function extractExpectSubject(trimmed) {
  const m = trimmed.match(/\bexpect\s*\(/);
  if (!m) return { subject: '', rest: '' };
  const start = m.index + m[0].length;
  let parenDepth = 1;
  let i = start;
  while (i < trimmed.length && parenDepth > 0) {
    if (trimmed[i] === '(') parenDepth++;
    else if (trimmed[i] === ')') parenDepth--;
    i++;
  }
  return {
    subject: trimmed.slice(start, i - 1),
    rest: trimmed.slice(i),
  };
}

/**
 * Extract the first arg from a matcher call, e.g. ".toBe(value)" → "value".
 * Returns null for zero-arg matchers like ".toBeNull()".
 */
function extractMatcherArg(rest) {
  const m = rest.match(/\.\w+\((.+)\)\s*;?\s*$/);
  if (!m) return null;
  const inner = m[1].trim();
  // Handle empty parens: .toBeNull() → inner is ""
  return inner || null;
}

/**
 * Extract kind, subject, expected, isNegated from a Jest assertion line.
 *
 * Handles:
 *   expect(subject).toBe(expected)
 *   expect(subject).not.toBe(expected)
 *   expect(subject).toBeNull()
 *   expect(subject).resolves.toBe(expected)
 *   expect(subject).rejects.toThrow()
 */
function parseJestAssertionLine(trimmed) {
  const { subject, rest } = extractExpectSubject(trimmed);
  const isNegated = /\.not\./.test(rest);
  let kind = 'equal';

  if (/\.toBe\(/.test(rest)) kind = 'strictEqual';
  else if (/\.toEqual\(/.test(rest)) kind = 'deepEqual';
  else if (/\.toBeTruthy\(/.test(rest)) kind = 'truthy';
  else if (/\.toBeFalsy\(/.test(rest)) kind = 'falsy';
  else if (/\.toBeNull\(/.test(rest)) kind = 'isNull';
  else if (/\.toBeUndefined\(/.test(rest)) kind = 'isUndefined';
  else if (/\.toBeDefined\(/.test(rest)) kind = 'isDefined';
  else if (/\.toBeNaN\(/.test(rest)) kind = 'isNaN';
  else if (/\.toBeInstanceOf\(/.test(rest)) kind = 'instanceOf';
  else if (/\.toMatch\(/.test(rest)) kind = 'matches';
  else if (/\.toContain\(/.test(rest)) kind = 'contains';
  else if (/\.toContainEqual\(/.test(rest)) kind = 'containsEqual';
  else if (/\.toHaveLength\(/.test(rest)) kind = 'hasLength';
  else if (/\.toHaveProperty\(/.test(rest)) kind = 'hasProperty';
  else if (/\.toBeGreaterThan\(/.test(rest)) kind = 'greaterThan';
  else if (/\.toBeLessThan\(/.test(rest)) kind = 'lessThan';
  else if (/\.toBeGreaterThanOrEqual\(/.test(rest)) kind = 'greaterOrEqual';
  else if (/\.toBeLessThanOrEqual\(/.test(rest)) kind = 'lessOrEqual';
  else if (/\.toBeCloseTo\(/.test(rest)) kind = 'closeTo';
  else if (/\.toThrow\(/.test(rest)) kind = 'throws';
  else if (/\.toHaveBeenCalled\b/.test(rest)) kind = 'called';
  else if (/\.toHaveBeenCalledWith\(/.test(rest)) kind = 'calledWith';
  else if (/\.toHaveBeenCalledTimes\(/.test(rest)) kind = 'calledTimes';
  else if (/\.toMatchSnapshot\(/.test(rest)) kind = 'snapshot';
  else if (/\.toMatchInlineSnapshot\(/.test(rest)) kind = 'snapshot';
  else if (/\.toHaveBeenLastCalledWith\(/.test(rest)) kind = 'calledWith';
  else if (/\.resolves\./.test(rest)) kind = 'resolves';
  else if (/\.rejects\./.test(rest)) kind = 'rejects';
  else if (/\.toStrictEqual\(/.test(rest)) kind = 'strictEqual';
  else if (/\.toHaveClass\(/.test(rest)) kind = 'hasClass';
  else if (/\.toHaveCount\(/.test(rest)) kind = 'hasCount';

  const expected = extractMatcherArg(rest);

  return { kind, subject, expected, isNegated };
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
