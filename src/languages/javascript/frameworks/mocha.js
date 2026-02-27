/**
 * Mocha+Chai+Sinon framework definition.
 *
 * Provides detect, parse, and emit for the Mocha testing framework
 * (with Chai assertion library and Sinon mocking library).
 * parse() builds an IR tree from Mocha source code for scoring.
 * emit() generates Mocha+Chai+Sinon code from IR (used when Mocha is the target).
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
 * Detect whether source code is Mocha+Chai.
 * Returns confidence score 0-100.
 */
function detect(source) {
  if (!source || !source.trim()) return 0;

  let score = 0;

  // Chai imports (strong signals)
  if (/require\s*\(\s*['"]chai['"]\s*\)/.test(source)) score += 30;
  if (/from\s+['"]chai['"]/.test(source)) score += 30;

  // Sinon imports (strong signals)
  if (/require\s*\(\s*['"]sinon['"]\s*\)/.test(source)) score += 25;
  if (/from\s+['"]sinon['"]/.test(source)) score += 25;

  // Chai assertion chains (strong signals)
  if (/\.to\.equal\s*\(/.test(source)) score += 20;
  if (/\.to\.deep\.equal\s*\(/.test(source)) score += 20;
  if (/\.to\.be\.true/.test(source)) score += 15;
  if (/\.to\.be\.false/.test(source)) score += 15;
  if (/\.to\.be\.null/.test(source)) score += 15;
  if (/\.to\.have\.lengthOf\s*\(/.test(source)) score += 15;
  if (/\.to\.include\s*\(/.test(source)) score += 15;
  if (/\.to\.have\.property\s*\(/.test(source)) score += 15;
  if (/\.to\.be\.a\s*\(/.test(source)) score += 10;
  if (/\.to\.exist/.test(source)) score += 10;

  // Sinon API (strong signals)
  if (/sinon\.stub\s*\(/.test(source)) score += 20;
  if (/sinon\.spy\s*\(/.test(source)) score += 20;
  if (/sinon\.assert\./.test(source)) score += 15;
  if (/sinon\.useFakeTimers/.test(source)) score += 15;
  if (/sinon\.restore/.test(source)) score += 10;
  if (/sinon\.createSandbox/.test(source)) score += 10;

  // Mocha-specific aliases (medium signals)
  if (/\bcontext\s*\(/.test(source)) score += 10;
  if (/\bspecify\s*\(/.test(source)) score += 10;

  // Mocha hooks (weak — shared names but before/after are Mocha-specific without All/Each)
  if (
    /\bbefore\s*\(/.test(source) &&
    !/\bbeforeAll\s*\(/.test(source) &&
    !/\bbeforeEach\s*\(/.test(source)
  )
    score += 5;
  if (
    /\bafter\s*\(/.test(source) &&
    !/\bafterAll\s*\(/.test(source) &&
    !/\bafterEach\s*\(/.test(source)
  )
    score += 5;

  // Test structure (weak — shared with Jest, Vitest, Jasmine)
  if (/\bdescribe\s*\(/.test(source)) score += 3;
  if (/\bit\s*\(/.test(source)) score += 3;
  if (/\bbeforeEach\s*\(/.test(source)) score += 2;

  // Negative signals: NOT Mocha
  if (/\bjest\.fn\b/.test(source)) score -= 30;
  if (/\bjest\.mock\b/.test(source)) score -= 30;
  if (/\bvi\.fn\b/.test(source)) score -= 30;
  if (/from\s+['"]vitest['"]/.test(source)) score -= 30;
  if (/jasmine\.createSpy/.test(source)) score -= 20;
  if (/jasmine\.createSpyObj/.test(source)) score -= 20;

  return Math.max(0, Math.min(100, score));
}

/**
 * Parse Mocha+Chai source code into an IR tree.
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
    if (
      /^import\s/.test(trimmed) ||
      /^const\s.*=\s*require\(/.test(trimmed) ||
      /^var\s.*=\s*require\(/.test(trimmed)
    ) {
      const sourceMatch =
        trimmed.match(/from\s+['"]([^'"]+)['"]/) ||
        trimmed.match(/require\s*\(\s*['"]([^'"]+)['"]\s*\)/);
      const isChai = sourceMatch && /^chai/.test(sourceMatch[1]);
      const isSinon = sourceMatch && sourceMatch[1] === 'sinon';
      allNodes.push(
        new ImportStatement({
          kind: isChai || isSinon ? 'library' : 'library',
          source: sourceMatch ? sourceMatch[1] : '',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      imports.push(allNodes[allNodes.length - 1]);
      continue;
    }

    // describe / context blocks
    if (/\b(?:describe|context)\s*\(/.test(trimmed)) {
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
            (trimmed.match(
              /(?:describe|context)(?:\.\w+)*\s*\(\s*['"`]([^'"`]*)['"`]/
            ) || [])[1] || '',
          modifiers,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // it / specify / test
    if (
      /\b(?:it|specify|test)\s*\(/.test(trimmed) ||
      /\b(?:it|specify|test)\.(?:only|skip)\s*\(/.test(trimmed)
    ) {
      const hasSkip = /\.skip/.test(trimmed);
      const hasOnly = /\.only/.test(trimmed);
      const isAsync = /async/.test(trimmed);
      const hasDone =
        /function\s*\(\s*done\s*\)/.test(trimmed) ||
        /\(\s*done\s*\)\s*=>/.test(trimmed);
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
        new TestCase({
          name:
            (trimmed.match(
              /(?:it|specify|test)(?:\.\w+)*\s*\(\s*['"`]([^'"`]*)['"`]/
            ) || [])[1] || '',
          isAsync: isAsync || hasDone,
          modifiers,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // Hooks: before, after, beforeEach, afterEach
    if (
      /\b(?:before|after|beforeEach|afterEach)\s*\(/.test(trimmed) &&
      !/\b(?:beforeAll|afterAll)\s*\(/.test(trimmed)
    ) {
      const hookMatch = trimmed.match(
        /\b(before|after|beforeEach|afterEach)\s*\(/
      );
      const hookType = hookMatch ? hookMatch[1] : 'beforeEach';
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

    // Chai expect assertions
    if (/\bexpect\s*\(/.test(trimmed) && /\.to\./.test(trimmed)) {
      const isNegated =
        /\.to\.not\./.test(trimmed) || /\.not\.to\./.test(trimmed);
      let kind = 'equal';
      if (
        /\.to(?:\.not)?\.deep\.equal\s*\(/.test(trimmed) ||
        /\.not\.to\.deep\.equal\s*\(/.test(trimmed)
      )
        kind = 'deepEqual';
      else if (
        /\.to(?:\.not)?\.equal\s*\(/.test(trimmed) ||
        /\.not\.to\.equal\s*\(/.test(trimmed)
      )
        kind = 'strictEqual';
      else if (/\.to(?:\.not)?\.be\.true/.test(trimmed)) kind = 'truthy';
      else if (/\.to(?:\.not)?\.be\.false/.test(trimmed)) kind = 'falsy';
      else if (/\.to(?:\.not)?\.be\.null/.test(trimmed)) kind = 'isNull';
      else if (/\.to(?:\.not)?\.be\.undefined/.test(trimmed))
        kind = 'isUndefined';
      else if (/\.to(?:\.not)?\.be\.a\s*\(/.test(trimmed)) kind = 'typeCheck';
      else if (/\.to(?:\.not)?\.have\.lengthOf\s*\(/.test(trimmed))
        kind = 'hasLength';
      else if (
        /\.to(?:\.not)?\.include\s*\(/.test(trimmed) ||
        /\.to(?:\.not)?\.contain\s*\(/.test(trimmed)
      )
        kind = 'contains';
      else if (/\.to(?:\.not)?\.have\.property\s*\(/.test(trimmed))
        kind = 'hasProperty';
      else if (/\.to(?:\.not)?\.be\.above\s*\(/.test(trimmed))
        kind = 'greaterThan';
      else if (/\.to(?:\.not)?\.be\.below\s*\(/.test(trimmed))
        kind = 'lessThan';
      else if (/\.to(?:\.not)?\.be\.closeTo\s*\(/.test(trimmed))
        kind = 'closeTo';
      else if (/\.to(?:\.not)?\.throw/.test(trimmed)) kind = 'throws';
      else if (/\.to(?:\.not)?\.match\s*\(/.test(trimmed)) kind = 'matches';
      else if (/\.to(?:\.not)?\.be\.an?\.instanceOf\s*\(/.test(trimmed))
        kind = 'instanceOf';
      else if (/\.to(?:\.not)?\.exist/.test(trimmed)) kind = 'isDefined';
      else if (/\.to(?:\.not)?\.be\.ok/.test(trimmed)) kind = 'truthy';
      else if (/\.to(?:\.not)?\.be\.NaN/.test(trimmed)) kind = 'isNaN';

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

    // Chai assert-style assertions
    if (/\bassert\./.test(trimmed)) {
      let kind = 'equal';
      if (/assert\.equal\s*\(/.test(trimmed)) kind = 'equal';
      else if (/assert\.strictEqual\s*\(/.test(trimmed)) kind = 'strictEqual';
      else if (/assert\.deepEqual\s*\(/.test(trimmed)) kind = 'deepEqual';
      else if (/assert\.isTrue\s*\(/.test(trimmed)) kind = 'truthy';
      else if (/assert\.isFalse\s*\(/.test(trimmed)) kind = 'falsy';
      else if (/assert\.isNull\s*\(/.test(trimmed)) kind = 'isNull';
      else if (/assert\.isNotNull\s*\(/.test(trimmed)) kind = 'isDefined';
      else if (/assert\.isUndefined\s*\(/.test(trimmed)) kind = 'isUndefined';
      else if (/assert\.isDefined\s*\(/.test(trimmed)) kind = 'isDefined';
      else if (/assert\.isOk\s*\(/.test(trimmed)) kind = 'truthy';
      else if (/assert\.throws\s*\(/.test(trimmed)) kind = 'throws';
      else if (/assert\.include\s*\(/.test(trimmed)) kind = 'contains';
      else if (/assert\.match\s*\(/.test(trimmed)) kind = 'matches';
      else if (/assert\.instanceOf\s*\(/.test(trimmed)) kind = 'instanceOf';
      else if (/assert\.lengthOf\s*\(/.test(trimmed)) kind = 'hasLength';
      else if (/assert\.property\s*\(/.test(trimmed)) kind = 'hasProperty';

      allNodes.push(
        new Assertion({
          kind,
          isNegated: false,
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // Sinon calls
    if (/sinon\./.test(trimmed)) {
      let kind = 'createMock';
      if (/sinon\.stub/.test(trimmed)) kind = 'createMock';
      else if (/sinon\.spy/.test(trimmed)) kind = 'spyOnMethod';
      else if (/sinon\.assert/.test(trimmed)) kind = 'mockAssertion';
      else if (/sinon\.useFakeTimers/.test(trimmed)) kind = 'fakeTimers';
      else if (/sinon\.restore/.test(trimmed)) kind = 'restoreMock';
      else if (/sinon\.createSandbox/.test(trimmed)) kind = 'createMock';
      else if (/sinon\.fakeServer/.test(trimmed)) kind = 'fakeTimers';

      allNodes.push(
        new MockCall({
          kind,
          sourceLocation: loc,
          originalSource: line,
          confidence: /sinon\.fakeServer/.test(trimmed)
            ? 'unconvertible'
            : 'converted',
        })
      );
      continue;
    }

    // Chai-sinon assertions (expect(fn).to.have.been.calledOnce etc.)
    if (/\bexpect\s*\(/.test(trimmed) && /\.to\.have\.been\./.test(trimmed)) {
      allNodes.push(
        new Assertion({
          kind: 'called',
          isNegated: /\.not\./.test(trimmed),
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // Regular expect without Chai chains (pass-through)
    if (/\bexpect\s*\(/.test(trimmed)) {
      allNodes.push(
        new Assertion({
          kind: 'equal',
          isNegated: /\.not\./.test(trimmed),
          sourceLocation: loc,
          originalSource: line,
          confidence: 'converted',
        })
      );
      continue;
    }

    // this.timeout / this.retries / this.slow
    if (/this\.timeout\s*\(/.test(trimmed)) {
      allNodes.push(
        new MockCall({
          kind: 'fakeTimers',
          target: 'timeout',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'warning',
        })
      );
      continue;
    }
    if (/this\.retries\s*\(/.test(trimmed)) {
      allNodes.push(
        new MockCall({
          kind: 'fakeTimers',
          target: 'retries',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'warning',
        })
      );
      continue;
    }
    if (/this\.slow\s*\(/.test(trimmed)) {
      allNodes.push(
        new MockCall({
          kind: 'fakeTimers',
          target: 'slow',
          sourceLocation: loc,
          originalSource: line,
          confidence: 'warning',
        })
      );
      continue;
    }

    // chai.use
    if (/chai\.use\s*\(/.test(trimmed)) {
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
 * Emit Mocha+Chai+Sinon code from IR + original source.
 *
 * Converts Jest code to Mocha+Chai+Sinon by applying pattern transforms.
 * Used when Mocha is the target framework (Jest→Mocha conversion).
 *
 * @param {TestFile} _ir - Parsed IR tree (for scoring metadata)
 * @param {string} source - Original source code (Jest)
 * @returns {string} Converted Mocha+Chai+Sinon source code
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

  // --- File-level Phase 0: Strip incoming HAMLET-TODO blocks ---
  let result = source;
  result = result.replace(
    /^[ \t]*\/\/ HAMLET-TODO \[[^\]]+\]:.*\n(?:[ \t]*\n)*(?:[ \t]*\/\/ (?:Original|Manual action required):.*\n(?:[ \t]*\n)*)*/gm,
    ''
  );

  // --- Per-node transforms: walk each line with its IR node type ---
  const lines = result.split('\n');
  const transformed = lines.map((line, i) => {
    const lineNum = i + 1;
    const node = nodeByLine.get(lineNum);

    if (!node) return line; // blank lines or unmapped

    switch (node.type) {
      case 'ImportStatement':
        return transformImportToMocha(line);
      case 'TestSuite':
        return line; // describe is shared
      case 'TestCase':
        return line; // it/test are shared
      case 'Hook':
        return transformHookToMocha(line);
      case 'Assertion':
        return transformAssertionToMocha(line);
      case 'MockCall':
        return transformMockToMocha(line);
      case 'Comment':
        return line; // pass-through
      default:
        // RawCode or unknown: apply conservative transforms
        return transformRawCodeToMocha(line);
    }
  });

  result = transformed.join('\n');

  // --- File-level: Add chai/sinon imports ---
  result = result.replace(
    /import\s+\{[^}]*\}\s+from\s+['"]@jest\/globals['"];?\n?/g,
    ''
  );

  const needsChai = /\.to\./.test(result);
  const needsSinon = /\bsinon\./.test(result);

  const importLines = [];
  if (needsChai) {
    importLines.push("const { expect } = require('chai');");
  }
  if (needsSinon) {
    importLines.push("const sinon = require('sinon');");
  }

  if (importLines.length > 0) {
    result = prependMochaImports(result, importLines);
  }

  // --- File-level: Unconvertible patterns ---
  result = result.replace(/\bjest\.mock\s*\(([^)]+)\)\s*;?/g, (match) => {
    return (
      formatter.formatTodo({
        id: 'UNCONVERTIBLE-MODULE-MOCK',
        description:
          'Mocha does not have a built-in module mocking system like jest.mock()',
        original: match.trim(),
        action: 'Use proxyquire, rewire, or manual dependency injection',
      }) +
      '\n// ' +
      match.trim()
    );
  });

  result = result.replace(
    /expect\([^)]+\)\.toMatchSnapshot\(\)\s*;?/g,
    (match) => {
      return (
        formatter.formatTodo({
          id: 'UNCONVERTIBLE-SNAPSHOT',
          description: 'Mocha does not have built-in snapshot testing',
          original: match.trim(),
          action: 'Use chai-jest-snapshot or snap-shot-it package',
        }) +
        '\n// ' +
        match.trim()
      );
    }
  );

  result = result.replace(
    /expect\([^)]+\)\.toMatchInlineSnapshot\([^)]*\)\s*;?/g,
    (match) => {
      return (
        formatter.formatTodo({
          id: 'UNCONVERTIBLE-INLINE-SNAPSHOT',
          description: 'Mocha does not support inline snapshots',
          original: match.trim(),
          action: 'Convert to explicit assertion',
        }) +
        '\n// ' +
        match.trim()
      );
    }
  );

  // Clean up leading blank lines and multiple blank lines
  result = result.replace(/^\n+/, '');
  result = result.replace(/\n{3,}/g, '\n\n');

  // Ensure trailing newline
  if (!result.endsWith('\n')) result += '\n';

  return result;
}

// ── Per-node transform functions (Mocha target) ──────────────────────

/**
 * Transform import lines: strip Jest-specific imports.
 */
function transformImportToMocha(line) {
  if (/from\s+['"]@jest\/globals['"]/.test(line)) return '';
  return line;
}

/**
 * Transform hook lines: beforeAll/afterAll → before/after.
 */
function transformHookToMocha(line) {
  let r = line;
  r = r.replace(/\bbeforeAll\s*\(/g, 'before(');
  r = r.replace(/\bafterAll\s*\(/g, 'after(');
  return r;
}

/**
 * Transform assertion lines: Jest expect → Chai chains.
 */
function transformAssertionToMocha(line) {
  let r = line;

  // Mock assertion chain-suffix replacements (before general assertions)
  r = r.replace(/\)\.toHaveBeenCalledWith\(/g, ').to.have.been.calledWith(');
  r = r.replace(/\)\.toHaveBeenCalled\(\)/g, ').to.have.been.called');
  r = r.replace(/\)\.not\.toHaveBeenCalled\(\)/g, ').to.not.have.been.called');

  // toHaveBeenCalledTimes needs subject restructuring
  const SUBJ = '([^)]*(?:\\([^)]*\\)[^)]*)*)';
  const calledTimesRe = new RegExp(
    'expect\\(' + SUBJ + '\\)\\.toHaveBeenCalledTimes\\(([^)]+)\\)',
    'g'
  );
  r = r.replace(calledTimesRe, 'expect($1.callCount).to.equal($2)');

  // Specific toBe literal cases (must come before general .toBe)
  r = r.replace(/\)\.toBe\(true\)/g, ').to.be.true');
  r = r.replace(/\)\.toBe\(false\)/g, ').to.be.false');
  r = r.replace(/\)\.toBe\(null\)/g, ').to.be.null');
  r = r.replace(/\)\.toBe\(undefined\)/g, ').to.be.undefined');

  // Negated assertion forms (must come before general positive forms)
  r = r.replace(/\)\.not\.toEqual\(/g, ').to.not.deep.equal(');
  r = r.replace(/\)\.not\.toStrictEqual\(/g, ').to.not.deep.equal(');
  r = r.replace(/\)\.not\.toBe\(/g, ').to.not.equal(');
  r = r.replace(/\)\.not\.toThrow\(\)/g, ').to.not.throw()');
  r = r.replace(/\)\.not\.toThrow\(/g, ').to.not.throw(');
  r = r.replace(/\)\.not\.toBeNull\(\)/g, ').to.not.be.null');
  r = r.replace(/\)\.not\.toBeUndefined\(\)/g, ').to.not.be.undefined');
  r = r.replace(/\)\.not\.toBeDefined\(\)/g, ').to.not.exist');
  r = r.replace(/\)\.not\.toContain\(/g, ').to.not.include(');
  r = r.replace(/\)\.not\.toHaveLength\(/g, ').to.not.have.lengthOf(');

  // General positive assertion forms
  r = r.replace(/\)\.toEqual\(/g, ').to.deep.equal(');
  r = r.replace(/\)\.toStrictEqual\(/g, ').to.deep.equal(');
  r = r.replace(/\)\.toBe\(/g, ').to.equal(');
  r = r.replace(/\)\.toBeNull\(\)/g, ').to.be.null');
  r = r.replace(/\)\.toBeUndefined\(\)/g, ').to.be.undefined');
  r = r.replace(/\)\.toBeDefined\(\)/g, ').to.exist');
  r = r.replace(/\)\.toBeNaN\(\)/g, ').to.be.NaN');
  r = r.replace(/\)\.toBeTruthy\(\)/g, ').to.be.ok');
  r = r.replace(/\)\.toBeFalsy\(\)/g, ').to.not.be.ok');
  r = r.replace(/\)\.toBeGreaterThan\(/g, ').to.be.above(');
  r = r.replace(/\)\.toBeLessThan\(/g, ').to.be.below(');
  r = r.replace(/\)\.toBeGreaterThanOrEqual\(/g, ').to.be.at.least(');
  r = r.replace(/\)\.toBeLessThanOrEqual\(/g, ').to.be.at.most(');
  r = r.replace(/\)\.toBeCloseTo\(/g, ').to.be.closeTo(');
  r = r.replace(/\)\.toBeInstanceOf\(/g, ').to.be.an.instanceOf(');
  r = r.replace(/\)\.toHaveLength\(/g, ').to.have.lengthOf(');
  r = r.replace(/\)\.toContain\(/g, ').to.include(');
  r = r.replace(/\)\.toHaveProperty\(/g, ').to.have.property(');
  r = r.replace(/\)\.toMatch\(/g, ').to.match(');
  r = r.replace(/\)\.toThrow\(\)/g, ').to.throw()');
  r = r.replace(/\)\.toThrow\(/g, ').to.throw(');

  // Jest mock transforms on assertion lines (e.g. expect(jest.fn()).toHaveBeenCalled)
  r = r.replace(/\bjest\.fn\(\)/g, 'sinon.stub()');
  r = r.replace(/\bjest\.fn\(([^)]+)\)/g, 'sinon.stub().callsFake($1)');
  r = r.replace(/\bjest\.spyOn\(([^,]+),\s*([^)]+)\)/g, 'sinon.spy($1, $2)');
  r = r.replace(/\.mockReturnValue\(([^)]+)\)/g, '.returns($1)');
  r = r.replace(/\.mockImplementation\(([^)]*)\)/g, '.callsFake($1)');
  r = r.replace(/\.mockResolvedValue\(([^)]+)\)/g, '.resolves($1)');
  r = r.replace(/\.mockRejectedValue\(([^)]+)\)/g, '.rejects($1)');
  r = r.replace(/\.mock\.calls\.length/g, '.callCount');
  r = r.replace(/\.mock\.calls\[(\d+)\]/g, '.calls.argsFor($1)');
  r = r.replace(/\.mock\.lastCall/g, '.calls.mostRecent().args');
  r = r.replace(/\.mock\.calls/g, '.calls.allArgs()');

  return r;
}

/**
 * Transform mock lines: Jest mocks → Sinon.
 */
function transformMockToMocha(line) {
  let r = line;

  // jest.fn() → sinon.stub()
  r = r.replace(/\bjest\.fn\(\)/g, 'sinon.stub()');
  r = r.replace(/\bjest\.fn\(([^)]+)\)/g, 'sinon.stub().callsFake($1)');

  // jest.spyOn → sinon.spy
  r = r.replace(/\bjest\.spyOn\(([^,]+),\s*([^)]+)\)/g, 'sinon.spy($1, $2)');

  // Mock chain methods
  r = r.replace(/\.mockReturnValue\(([^)]+)\)/g, '.returns($1)');
  r = r.replace(
    /\.mockReturnValueOnce\(([^)]+)\)/g,
    '.onFirstCall().returns($1)'
  );
  r = r.replace(/\.mockImplementation\(([^)]*)\)/g, '.callsFake($1)');
  r = r.replace(/\.mockResolvedValue\(([^)]+)\)/g, '.resolves($1)');
  r = r.replace(/\.mockRejectedValue\(([^)]+)\)/g, '.rejects($1)');
  r = r.replace(/\.mockClear\(\)/g, '.resetHistory()');
  r = r.replace(/\.mockReset\(\)/g, '.reset()');

  // Global Jest mock API
  r = r.replace(/\bjest\.clearAllMocks\(\)/g, 'sinon.reset()');
  r = r.replace(/\bjest\.resetAllMocks\(\)/g, 'sinon.reset()');
  r = r.replace(/\bjest\.restoreAllMocks\(\)/g, 'sinon.restore()');

  // Timer mocks
  r = r.replace(/\bjest\.useFakeTimers\(\)/g, 'sinon.useFakeTimers()');
  r = r.replace(/\bjest\.useRealTimers\(\)/g, 'clock.restore()');
  r = r.replace(/\bjest\.advanceTimersByTime\(([^)]+)\)/g, 'clock.tick($1)');

  // Assertion transforms that may appear on MockCall lines
  r = r.replace(/\)\.toHaveBeenCalledWith\(/g, ').to.have.been.calledWith(');
  r = r.replace(/\)\.toHaveBeenCalled\(\)/g, ').to.have.been.called');

  // Hook transforms (beforeAll/afterAll inside mock setup blocks)
  r = r.replace(/\bbeforeAll\s*\(/g, 'before(');
  r = r.replace(/\bafterAll\s*\(/g, 'after(');

  return r;
}

/**
 * Transform RawCode lines: conservative transforms for continuation lines.
 */
function transformRawCodeToMocha(line) {
  let r = line;

  // Jest assertion chain-suffix transforms (continuation lines)
  r = r.replace(/\)\.toBe\(true\)/g, ').to.be.true');
  r = r.replace(/\)\.toBe\(false\)/g, ').to.be.false');
  r = r.replace(/\)\.toBe\(null\)/g, ').to.be.null');
  r = r.replace(/\)\.toBe\(undefined\)/g, ').to.be.undefined');
  r = r.replace(/\)\.not\.toEqual\(/g, ').to.not.deep.equal(');
  r = r.replace(/\)\.not\.toBe\(/g, ').to.not.equal(');
  r = r.replace(/\)\.toEqual\(/g, ').to.deep.equal(');
  r = r.replace(/\)\.toBe\(/g, ').to.equal(');
  r = r.replace(/\)\.toBeNull\(\)/g, ').to.be.null');
  r = r.replace(/\)\.toBeUndefined\(\)/g, ').to.be.undefined');
  r = r.replace(/\)\.toBeDefined\(\)/g, ').to.exist');
  r = r.replace(/\)\.toBeTruthy\(\)/g, ').to.be.ok');
  r = r.replace(/\)\.toBeFalsy\(\)/g, ').to.not.be.ok');
  r = r.replace(/\)\.toBeGreaterThan\(/g, ').to.be.above(');
  r = r.replace(/\)\.toBeLessThan\(/g, ').to.be.below(');
  r = r.replace(/\)\.toHaveLength\(/g, ').to.have.lengthOf(');
  r = r.replace(/\)\.toContain\(/g, ').to.include(');
  r = r.replace(/\)\.toHaveProperty\(/g, ').to.have.property(');
  r = r.replace(/\)\.toMatch\(/g, ').to.match(');
  r = r.replace(/\)\.toThrow\(\)/g, ').to.throw()');
  r = r.replace(/\)\.toThrow\(/g, ').to.throw(');
  r = r.replace(/\)\.toHaveBeenCalledWith\(/g, ').to.have.been.calledWith(');
  r = r.replace(/\)\.toHaveBeenCalled\(\)/g, ').to.have.been.called');

  // Mock chain continuations on raw lines
  r = r.replace(/\.mockReturnValue\(([^)]+)\)/g, '.returns($1)');
  r = r.replace(/\.mockImplementation\(([^)]*)\)/g, '.callsFake($1)');
  r = r.replace(/\.mockResolvedValue\(([^)]+)\)/g, '.resolves($1)');
  r = r.replace(/\.mockRejectedValue\(([^)]+)\)/g, '.rejects($1)');
  r = r.replace(/\.mockClear\(\)/g, '.resetHistory()');
  r = r.replace(/\.mockReset\(\)/g, '.reset()');
  r = r.replace(/\bjest\.fn\(\)/g, 'sinon.stub()');
  r = r.replace(/\bjest\.fn\(([^)]+)\)/g, 'sinon.stub().callsFake($1)');
  r = r.replace(/\bjest\.spyOn\(([^,]+),\s*([^)]+)\)/g, 'sinon.spy($1, $2)');
  r = r.replace(/\bjest\.restoreAllMocks\(\)/g, 'sinon.restore()');
  r = r.replace(/\bjest\.clearAllMocks\(\)/g, 'sinon.reset()');
  r = r.replace(/\bjest\.resetAllMocks\(\)/g, 'sinon.reset()');
  r = r.replace(/\bjest\.useFakeTimers\(\)/g, 'sinon.useFakeTimers()');
  r = r.replace(/\bjest\.useRealTimers\(\)/g, 'clock.restore()');
  r = r.replace(/\bjest\.advanceTimersByTime\(([^)]+)\)/g, 'clock.tick($1)');

  // Hook transforms on raw lines
  r = r.replace(/\bbeforeAll\s*\(/g, 'before(');
  r = r.replace(/\bafterAll\s*\(/g, 'after(');

  return r;
}

/**
 * Prepend Mocha imports (chai, sinon) to the source code.
 */
function prependMochaImports(source, importLines) {
  const lines = source.split('\n');
  let insertIdx = 0;

  // Skip leading comments
  for (let i = 0; i < lines.length; i++) {
    const trimmed = lines[i].trim();
    if (
      trimmed.startsWith('//') ||
      trimmed.startsWith('/*') ||
      trimmed.startsWith('*') ||
      trimmed === '*/'
    ) {
      insertIdx = i + 1;
      continue;
    }
    if (trimmed === '' && i === insertIdx) {
      insertIdx = i + 1;
      continue;
    }
    break;
  }

  lines.splice(insertIdx, 0, ...importLines);

  // Ensure blank line after imports
  const afterImports = insertIdx + importLines.length;
  if (
    afterImports < lines.length &&
    lines[afterImports].trim() !== '' &&
    !lines[afterImports].trim().startsWith('const') &&
    !lines[afterImports].trim().startsWith('import')
  ) {
    lines.splice(afterImports, 0, '');
  }

  return lines.join('\n');
}

export default {
  name: 'mocha',
  language: 'javascript',
  paradigm: 'bdd',
  detect,
  parse,
  emit,
  imports: {
    globals: [
      'describe',
      'it',
      'context',
      'specify',
      'before',
      'after',
      'beforeEach',
      'afterEach',
    ],
    externalLibs: ['chai', 'sinon'],
    mockNamespace: 'sinon',
  },
};
