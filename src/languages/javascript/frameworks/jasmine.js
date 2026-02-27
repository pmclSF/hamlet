/**
 * Jasmine framework definition.
 *
 * Provides detect, parse, and emit for the Jasmine testing framework.
 * parse() builds an IR tree from Jasmine source code for scoring.
 * emit() generates Jasmine code from IR (used when Jasmine is the target).
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
} from '../../../core/ir.js';
import { TodoFormatter } from '../../../core/TodoFormatter.js';

const formatter = new TodoFormatter('javascript');

/**
 * Detect whether source code is Jasmine.
 * Returns confidence score 0-100.
 */
function detect(source) {
  if (!source || !source.trim()) return 0;

  let score = 0;

  // Jasmine-specific APIs (strong signals)
  if (/jasmine\.createSpy\s*\(/.test(source)) score += 25;
  if (/jasmine\.createSpyObj\s*\(/.test(source)) score += 25;
  if (/jasmine\.clock\s*\(\)/.test(source)) score += 20;
  if (/jasmine\.any\s*\(/.test(source)) score += 15;
  if (/jasmine\.objectContaining\s*\(/.test(source)) score += 15;
  if (/jasmine\.arrayContaining\s*\(/.test(source)) score += 15;
  if (/jasmine\.stringMatching\s*\(/.test(source)) score += 15;
  if (/jasmine\.anything\s*\(\)/.test(source)) score += 10;

  // Jasmine spy chains (strong signals)
  if (/\.and\.returnValue\s*\(/.test(source)) score += 15;
  if (/\.and\.callFake\s*\(/.test(source)) score += 15;
  if (/\.and\.throwError\s*\(/.test(source)) score += 15;
  if (/\.and\.callThrough\s*\(/.test(source)) score += 10;
  if (/\.calls\.count\s*\(\)/.test(source)) score += 15;
  if (/\.calls\.argsFor\s*\(/.test(source)) score += 15;
  if (/\.calls\.mostRecent\s*\(\)/.test(source)) score += 10;
  if (/\.calls\.reset\s*\(\)/.test(source)) score += 10;

  // Jasmine focus/exclude modifiers (medium signals)
  if (/\bfdescribe\s*\(/.test(source)) score += 15;
  if (/\bxdescribe\s*\(/.test(source)) score += 15;
  if (/\bfit\s*\(/.test(source)) score += 15;
  if (/\bxit\s*\(/.test(source)) score += 15;
  if (/\bpending\s*\(/.test(source)) score += 10;

  // Test structure (weak — shared)
  if (/\bdescribe\s*\(/.test(source)) score += 3;
  if (/\bit\s*\(/.test(source)) score += 3;
  if (/\bexpect\s*\(/.test(source)) score += 3;
  if (/\bbeforeEach\s*\(/.test(source)) score += 2;
  if (/\bspyOn\s*\(/.test(source)) score += 5;

  // Negative signals: NOT Jasmine
  if (/\bjest\.fn\b/.test(source)) score -= 30;
  if (/\bjest\.mock\b/.test(source)) score -= 30;
  if (/\bvi\.fn\b/.test(source)) score -= 30;
  if (/from\s+['"]vitest['"]/.test(source)) score -= 30;
  if (/require\s*\(\s*['"]chai['"]\s*\)/.test(source)) score -= 20;
  if (/from\s+['"]chai['"]/.test(source)) score -= 20;
  if (/require\s*\(\s*['"]sinon['"]\s*\)/.test(source)) score -= 20;

  return Math.max(0, Math.min(100, score));
}

/**
 * Parse Jasmine source code into an IR tree.
 *
 * Regex-based parser that identifies test structure elements
 * and creates IR nodes for scoring and conversion tracking.
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
    if (
      trimmed.startsWith('//') ||
      trimmed.startsWith('/*') ||
      trimmed.startsWith('*')
    ) {
      const isDirective =
        /eslint-disable|noinspection|@ts-|type:\s*ignore/.test(trimmed);
      const isLicense =
        /license|copyright|MIT|Apache|BSD/i.test(trimmed) && i < 5;
      allNodes.push(
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
      continue;
    }

    // Import/require statements
    if (/^import\s/.test(trimmed) || /^const\s.*=\s*require\(/.test(trimmed)) {
      const sourceMatch =
        trimmed.match(/from\s+['"]([^'"]+)['"]/) ||
        trimmed.match(/require\s*\(\s*['"]([^'"]+)['"]\s*\)/);
      allNodes.push(
        new ImportStatement({
          kind: 'library',
          source: sourceMatch ? sourceMatch[1] : '',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      imports.push(allNodes[allNodes.length - 1]);
      continue;
    }

    // fdescribe / xdescribe
    if (/\b(?:fdescribe|xdescribe)\s*\(/.test(trimmed)) {
      const isFocused = /\bfdescribe/.test(trimmed);
      const isSkipped = /\bxdescribe/.test(trimmed);
      const modifiers = [];
      if (isFocused)
        modifiers.push(
          new Modifier({ modifierType: 'only', sourceLocation: loc })
        );
      if (isSkipped)
        modifiers.push(
          new Modifier({ modifierType: 'skip', sourceLocation: loc })
        );
      allNodes.push(
        new TestSuite({
          name:
            (trimmed.match(
              /(?:fdescribe|xdescribe)\s*\(\s*['"`]([^'"`]*)['"`]/
            ) || [])[1] || '',
          modifiers,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // describe blocks
    if (/\bdescribe\s*\(/.test(trimmed)) {
      allNodes.push(
        new TestSuite({
          name:
            (trimmed.match(/describe\s*\(\s*['"`]([^'"`]*)['"`]/) || [])[1] ||
            '',
          modifiers: [],
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // fit / xit
    if (/\b(?:fit|xit)\s*\(/.test(trimmed)) {
      const isFocused = /\bfit\s*\(/.test(trimmed);
      const isSkipped = /\bxit\s*\(/.test(trimmed);
      const modifiers = [];
      if (isFocused)
        modifiers.push(
          new Modifier({ modifierType: 'only', sourceLocation: loc })
        );
      if (isSkipped)
        modifiers.push(
          new Modifier({ modifierType: 'skip', sourceLocation: loc })
        );
      allNodes.push(
        new TestCase({
          name:
            (trimmed.match(/(?:fit|xit)\s*\(\s*['"`]([^'"`]*)['"`]/) ||
              [])[1] || '',
          isAsync: /async/.test(trimmed),
          modifiers,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // it / test
    if (/\b(?:it|test)\s*\(/.test(trimmed)) {
      const isAsync = /async/.test(trimmed);
      const hasDone =
        /function\s*\(\s*done\s*\)/.test(trimmed) ||
        /\(\s*done\s*\)\s*=>/.test(trimmed);
      allNodes.push(
        new TestCase({
          name:
            (trimmed.match(/(?:it|test)\s*\(\s*['"`]([^'"`]*)['"`]/) ||
              [])[1] || '',
          isAsync: isAsync || hasDone,
          modifiers: [],
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // Hooks: beforeEach, afterEach, beforeAll, afterAll
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

    // pending()
    if (/\bpending\s*\(/.test(trimmed)) {
      allNodes.push(
        new RawCode({
          code: line,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'warning',
        })
      );
      continue;
    }

    // jasmine.createSpy / jasmine.createSpyObj
    if (
      /jasmine\.createSpy/.test(trimmed) ||
      /jasmine\.createSpyObj/.test(trimmed)
    ) {
      allNodes.push(
        new MockCall({
          kind: 'createMock',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // spyOn
    if (/\bspyOn\s*\(/.test(trimmed)) {
      allNodes.push(
        new MockCall({
          kind: 'spyOnMethod',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // jasmine.clock()
    if (/jasmine\.clock\s*\(\)/.test(trimmed)) {
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

    // jasmine.any / jasmine.objectContaining / etc.
    if (
      /jasmine\.(?:any|objectContaining|arrayContaining|stringMatching|anything)\s*\(/.test(
        trimmed
      )
    ) {
      allNodes.push(
        new Assertion({
          kind: 'equal',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // jasmine.addMatchers
    if (/jasmine\.addMatchers\s*\(/.test(trimmed)) {
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
      else if (/\.toContain\(/.test(trimmed)) kind = 'contains';
      else if (/\.toHaveBeenCalled\b/.test(trimmed)) kind = 'called';
      else if (/\.toHaveBeenCalledWith\(/.test(trimmed)) kind = 'calledWith';
      else if (/\.toHaveBeenCalledTimes\(/.test(trimmed)) kind = 'calledTimes';
      else if (/\.toThrow\(/.test(trimmed) || /\.toThrowError\(/.test(trimmed))
        kind = 'throws';
      else if (/\.toMatch\(/.test(trimmed)) kind = 'matches';

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

    // spy chains (.and.returnValue, .calls.count, etc.)
    if (
      /\.and\.(?:returnValue|callFake|throwError|callThrough|stub)\s*\(/.test(
        trimmed
      ) ||
      /\.calls\.(?:count|argsFor|mostRecent|reset|allArgs|first)\s*\(\)/.test(
        trimmed
      )
    ) {
      allNodes.push(
        new MockCall({
          kind: 'spyOnMethod',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // Everything else
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
 * Emit Jasmine code from IR + original source.
 *
 * Converts Jest code to Jasmine by applying pattern transforms.
 * Used when Jasmine is the target framework (Jest→Jasmine conversion).
 *
 * @param {TestFile} _ir - Parsed IR tree (for scoring metadata)
 * @param {string} source - Original source code (Jest)
 * @returns {string} Converted Jasmine source code
 */
function emit(_ir, source) {
  let result = source;

  // Strip incoming HAMLET-TODO blocks (from previous round-trip step)
  result = result.replace(
    /^[ \t]*\/\/ HAMLET-TODO \[[^\]]+\]:.*\n(?:[ \t]*\n)*(?:[ \t]*\/\/ (?:Original|Manual action required):.*\n(?:[ \t]*\n)*)*/gm,
    ''
  );

  // --- Phase 1: Convert Jest mock API → Jasmine spies ---

  // jest.fn() → jasmine.createSpy()
  result = result.replace(/\bjest\.fn\(\)/g, 'jasmine.createSpy()');

  // jest.fn(impl) → jasmine.createSpy().and.callFake(impl)
  result = result.replace(
    /\bjest\.fn\(([^)]+)\)/g,
    'jasmine.createSpy().and.callFake($1)'
  );

  // jest.spyOn(obj, 'method') → spyOn(obj, 'method')
  result = result.replace(
    /\bjest\.spyOn\(([^,]+),\s*([^)]+)\)/g,
    'spyOn($1, $2)'
  );

  // .mockReturnValue(val) → .and.returnValue(val)
  result = result.replace(
    /\.mockReturnValue\(([^)]+)\)/g,
    '.and.returnValue($1)'
  );

  // .mockReturnValueOnce(val) → .and.returnValue(val) (simplified — Jasmine doesn't have once)
  result = result.replace(
    /\.mockReturnValueOnce\(([^)]+)\)/g,
    '.and.returnValue($1)'
  );

  // .mockImplementation(fn) → .and.callFake(fn)
  result = result.replace(
    /\.mockImplementation\(([^)]*)\)/g,
    '.and.callFake($1)'
  );

  // .mockResolvedValue(val) → .and.returnValue(Promise.resolve(val))
  result = result.replace(
    /\.mockResolvedValue\(([^)]+)\)/g,
    '.and.returnValue(Promise.resolve($1))'
  );

  // .mockRejectedValue(err) → .and.returnValue(Promise.reject(err))
  result = result.replace(
    /\.mockRejectedValue\(([^)]+)\)/g,
    '.and.returnValue(Promise.reject($1))'
  );

  // .mockClear() → .calls.reset()
  result = result.replace(/\.mockClear\(\)/g, '.calls.reset()');

  // .mockReset() → .calls.reset() and .and.stub()
  result = result.replace(/\.mockReset\(\)/g, '.calls.reset()');

  // .mock.calls.length → .calls.count()
  result = result.replace(/\.mock\.calls\.length/g, '.calls.count()');

  // .mock.calls[n] → .calls.argsFor(n)
  result = result.replace(/\.mock\.calls\[(\d+)\]/g, '.calls.argsFor($1)');

  // .mock.calls (bare, not followed by [ or .length) → .calls.allArgs()
  result = result.replace(/\.mock\.calls(?![\[.])/g, '.calls.allArgs()');

  // .mock.lastCall → .calls.mostRecent().args
  result = result.replace(/\.mock\.lastCall/g, '.calls.mostRecent().args');

  // --- Phase 2: Convert asymmetric matchers ---

  // expect.any(X) → jasmine.any(X)
  result = result.replace(/\bexpect\.any\(([^)]+)\)/g, 'jasmine.any($1)');

  // expect.anything() → jasmine.anything()
  result = result.replace(/\bexpect\.anything\(\)/g, 'jasmine.anything()');

  // expect.objectContaining → jasmine.objectContaining
  result = result.replace(
    /\bexpect\.objectContaining\(/g,
    'jasmine.objectContaining('
  );

  // expect.arrayContaining → jasmine.arrayContaining
  result = result.replace(
    /\bexpect\.arrayContaining\(/g,
    'jasmine.arrayContaining('
  );

  // expect.stringContaining → jasmine.stringMatching
  result = result.replace(
    /\bexpect\.stringContaining\(/g,
    'jasmine.stringMatching('
  );

  // expect.stringMatching → jasmine.stringMatching
  result = result.replace(
    /\bexpect\.stringMatching\(/g,
    'jasmine.stringMatching('
  );

  // --- Phase 3: Convert modifiers ---

  // describe.only → fdescribe
  result = result.replace(/\bdescribe\.only\s*\(/g, 'fdescribe(');

  // describe.skip → xdescribe
  result = result.replace(/\bdescribe\.skip\s*\(/g, 'xdescribe(');

  // it.only → fit
  result = result.replace(/\bit\.only\s*\(/g, 'fit(');

  // it.skip → xit
  result = result.replace(/\bit\.skip\s*\(/g, 'xit(');

  // test.only → fit
  result = result.replace(/\btest\.only\s*\(/g, 'fit(');

  // test.skip → xit
  result = result.replace(/\btest\.skip\s*\(/g, 'xit(');

  // it.todo('name') → xit('name', () => { pending(); })
  result = result.replace(
    /\bit\.todo\s*\(\s*(['"`][^'"`]*['"`])\s*\)/g,
    'xit($1, () => { pending(); })'
  );

  // test → it (Jasmine uses it, not test)
  result = result.replace(/\btest\s*\(\s*(['"`])/g, 'it($1');

  // --- Phase 4: Convert timers ---

  // jest.useFakeTimers() → jasmine.clock().install()
  result = result.replace(
    /\bjest\.useFakeTimers\(\)/g,
    'jasmine.clock().install()'
  );

  // jest.useRealTimers() → jasmine.clock().uninstall()
  result = result.replace(
    /\bjest\.useRealTimers\(\)/g,
    'jasmine.clock().uninstall()'
  );

  // jest.advanceTimersByTime(ms) → jasmine.clock().tick(ms)
  result = result.replace(
    /\bjest\.advanceTimersByTime\(([^)]+)\)/g,
    'jasmine.clock().tick($1)'
  );

  // jest.setSystemTime(date) → jasmine.clock().mockDate(date)
  result = result.replace(
    /\bjest\.setSystemTime\(([^)]+)\)/g,
    'jasmine.clock().mockDate($1)'
  );

  // jest.clearAllMocks() → (no direct equivalent, remove or comment)
  result = result.replace(
    /\bjest\.clearAllMocks\(\)\s*;?/g,
    '// Jasmine spies are auto-cleaned between specs'
  );

  // jest.resetAllMocks()
  result = result.replace(
    /\bjest\.resetAllMocks\(\)\s*;?/g,
    '// Jasmine spies are auto-cleaned between specs'
  );

  // jest.restoreAllMocks()
  result = result.replace(
    /\bjest\.restoreAllMocks\(\)\s*;?/g,
    '// Jasmine spies are auto-cleaned between specs'
  );

  // --- Phase 5: Unconvertible patterns ---

  // jest.mock(module) → HAMLET-TODO
  result = result.replace(/\bjest\.mock\s*\(([^)]+)\)\s*;?/g, (match) => {
    return (
      formatter.formatTodo({
        id: 'UNCONVERTIBLE-MODULE-MOCK',
        description:
          'Jasmine does not have a built-in module mocking system like jest.mock()',
        original: match.trim(),
        action: 'Use manual dependency injection or a module mocking library',
      }) +
      '\n// ' +
      match.trim()
    );
  });

  // toMatchSnapshot → HAMLET-TODO
  result = result.replace(
    /expect\([^)]+\)\.toMatchSnapshot\(\)\s*;?/g,
    (match) => {
      return (
        formatter.formatTodo({
          id: 'UNCONVERTIBLE-SNAPSHOT',
          description: 'Jasmine does not have built-in snapshot testing',
          original: match.trim(),
          action: 'Use jasmine-snapshot or convert to explicit assertion',
        }) +
        '\n// ' +
        match.trim()
      );
    }
  );

  // toMatchInlineSnapshot → HAMLET-TODO
  result = result.replace(
    /expect\([^)]+\)\.toMatchInlineSnapshot\([^)]*\)\s*;?/g,
    (match) => {
      return (
        formatter.formatTodo({
          id: 'UNCONVERTIBLE-INLINE-SNAPSHOT',
          description: 'Jasmine does not support inline snapshots',
          original: match.trim(),
          action: 'Convert to explicit assertion',
        }) +
        '\n// ' +
        match.trim()
      );
    }
  );

  // Remove any existing jest-related imports
  result = result.replace(
    /import\s+\{[^}]*\}\s+from\s+['"]@jest\/globals['"];?\n?/g,
    ''
  );

  // Clean up multiple blank lines
  result = result.replace(/\n{3,}/g, '\n\n');

  // Ensure trailing newline
  if (!result.endsWith('\n')) result += '\n';

  return result;
}

export default {
  name: 'jasmine',
  language: 'javascript',
  paradigm: 'bdd',
  detect,
  parse,
  emit,
  imports: {
    globals: [
      'describe',
      'it',
      'expect',
      'beforeEach',
      'afterEach',
      'beforeAll',
      'afterAll',
      'fdescribe',
      'xdescribe',
      'fit',
      'xit',
      'spyOn',
      'pending',
    ],
    mockNamespace: 'jasmine',
  },
};
