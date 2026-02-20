/**
 * Vitest framework definition.
 *
 * Provides detect, parse, and emit for the Vitest testing framework.
 * emit() is the primary function — it transforms Jest source code into
 * Vitest code by applying pattern substitutions and adding imports.
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
 * Detect whether source code is Vitest.
 * Returns confidence score 0-100.
 */
function detect(source) {
  if (!source || !source.trim()) return 0;

  let score = 0;

  // Vitest-specific (strong signals)
  if (/from\s+['"]vitest['"]/.test(source)) score += 40;
  if (/\bvi\.fn\s*\(/.test(source)) score += 25;
  if (/\bvi\.mock\s*\(/.test(source)) score += 25;
  if (/\bvi\.spyOn\s*\(/.test(source)) score += 20;
  if (/\bvi\.useFakeTimers/.test(source)) score += 15;
  if (/\bvi\.importActual/.test(source)) score += 15;

  // Test structure (weak — shared with Jest, Mocha, Jasmine)
  if (/\bdescribe\s*\(/.test(source)) score += 3;
  if (/\b(?:it|test)\s*\(/.test(source)) score += 3;
  if (/\bexpect\s*\(/.test(source)) score += 3;

  // Negative: Jest-specific
  if (/\bjest\.fn\b/.test(source)) score -= 30;
  if (/\bjest\.mock\b/.test(source)) score -= 30;

  // Negative: Mocha/Chai/Sinon
  if (/from\s+['"]chai['"]/.test(source)) score -= 20;
  if (/require\s*\(\s*['"]chai['"]\s*\)/.test(source)) score -= 20;
  if (/require\s*\(\s*['"]sinon['"]\s*\)/.test(source)) score -= 20;

  // Negative: Jasmine-specific
  if (/jasmine\.createSpy/.test(source)) score -= 20;

  return Math.max(0, Math.min(100, score));
}

/**
 * Parse Vitest source code into an IR tree.
 * Stub — used when Vitest is the source (Vitest→Jest direction).
 * For now, delegates to a minimal implementation.
 */
function parse(source) {
  // Minimal parse for detection/scoring. Full parse not yet needed.
  return new TestFile({
    language: 'javascript',
    imports: [],
    body: [new RawCode({ code: source })],
  });
}

/**
 * Emit Vitest code from IR + original source.
 *
 * For same-paradigm conversions (Jest→Vitest), this works on the
 * original source text, applying pattern transforms and adding imports.
 * The IR is used to check for unconvertible/warning patterns.
 *
 * @param {TestFile} ir - Parsed IR tree (for scoring metadata)
 * @param {string} source - Original source code (Jest)
 * @returns {string} Converted Vitest source code
 */
function emit(ir, source) {
  let result = source;
  const warnings = [];
  const todos = [];

  // --- Phase 1: Handle special cases before simple replacements ---

  // jest.requireActual → await vi.importActual (factory must become async)
  result = convertRequireActual(result);

  // jest.mock with { virtual: true } → HAMLET-TODO
  result = convertVirtualMocks(result, todos);

  // jest.retryTimes → HAMLET-TODO
  result = convertRetryTimes(result, todos);

  // jest.setTimeout → vi.setConfig
  result = result.replace(
    /jest\.setTimeout\s*\(\s*(\d+)\s*\)/g,
    'vi.setConfig({ testTimeout: $1 })'
  );

  // expect.addSnapshotSerializer → HAMLET-TODO
  result = convertSnapshotSerializer(result, todos);

  // --- Phase 2: Simple jest.* → vi.* replacements ---

  // jest.fn → vi.fn (careful not to match inside strings/comments)
  result = result.replace(/\bjest\.fn\b/g, 'vi.fn');
  result = result.replace(/\bjest\.spyOn\b/g, 'vi.spyOn');
  result = result.replace(/\bjest\.useFakeTimers\b/g, 'vi.useFakeTimers');
  result = result.replace(/\bjest\.useRealTimers\b/g, 'vi.useRealTimers');
  result = result.replace(
    /\bjest\.advanceTimersByTime\b/g,
    'vi.advanceTimersByTime'
  );
  result = result.replace(
    /\bjest\.advanceTimersToNextTimer\b/g,
    'vi.advanceTimersToNextTimer'
  );
  result = result.replace(/\bjest\.runAllTimers\b/g, 'vi.runAllTimers');
  result = result.replace(
    /\bjest\.runOnlyPendingTimers\b/g,
    'vi.runOnlyPendingTimers'
  );
  result = result.replace(/\bjest\.clearAllTimers\b/g, 'vi.clearAllTimers');
  result = result.replace(/\bjest\.clearAllMocks\b/g, 'vi.clearAllMocks');
  result = result.replace(/\bjest\.resetAllMocks\b/g, 'vi.resetAllMocks');
  result = result.replace(/\bjest\.restoreAllMocks\b/g, 'vi.restoreAllMocks');
  result = result.replace(/\bjest\.resetModules\b/g, 'vi.resetModules');
  result = result.replace(/\bjest\.isMockFunction\b/g, 'vi.isMockFunction');

  // jest.mock() → vi.mock() (simple case without special options)
  // Must run after virtual mock handling
  result = result.replace(/\bjest\.mock\b/g, 'vi.mock');

  // --- Phase 3: Add warning comments for known behavioral differences ---

  // Warn about mock hoisting differences
  if (/\bvi\.mock\s*\(/.test(result)) {
    result = addMockHostingWarning(result, warnings);
  }

  // Warn about snapshot file location
  if (/\.toMatchSnapshot\s*\(/.test(result)) {
    result = addSnapshotWarning(result, warnings);
  }

  // Warn about __mocks__ directory
  if (ir && hasAutomockPattern(ir)) {
    warnings.push(
      '__mocks__ directory convention: verify mock file resolution in Vitest'
    );
  }

  // --- Phase 4: Generate and add import statement ---

  // Remove any existing jest-related imports (e.g., @jest/globals)
  result = result.replace(
    /import\s+\{[^}]*\}\s+from\s+['"]@jest\/globals['"];?\n?/g,
    ''
  );

  const vitestImport = generateVitestImport(result);
  result = prependImport(result, vitestImport);

  // --- Phase 5: Insert TODO comments ---
  for (const todo of todos) {
    result = result.replace(
      todo.marker,
      todo.comment + '\n' + todo.replacement
    );
  }

  // Clean up multiple blank lines
  result = result.replace(/\n{3,}/g, '\n\n');

  // Ensure trailing newline
  if (!result.endsWith('\n')) result += '\n';

  return result;
}

/**
 * Convert jest.requireActual to await vi.importActual.
 * The containing factory function must become async.
 */
function convertRequireActual(source) {
  // Pattern: jest.mock('module', () => { ... jest.requireActual ... })
  // Must make factory async and requireActual → await importActual
  let result = source;

  // First, handle standalone jest.requireActual references
  result = result.replace(/\bjest\.requireActual\b/g, 'await vi.importActual');

  // If we converted any requireActual, the containing jest.mock factory
  // must become async. Look for vi.mock (already converted) with non-async factory.
  if (/await vi\.importActual/.test(result)) {
    // Convert: vi.mock('path', () => { → vi.mock('path', async () => {
    result = result.replace(
      /(vi\.mock\s*\([^,]+,\s*)(\(\)\s*=>)/g,
      '$1async $2'
    );
    // Also handle: vi.mock('path', function() → vi.mock('path', async function()
    result = result.replace(
      /(vi\.mock\s*\([^,]+,\s*)(function\s*\()/g,
      '$1async $2'
    );
  }

  return result;
}

/**
 * Convert jest.mock with { virtual: true } to HAMLET-TODO.
 */
function convertVirtualMocks(source, todos) {
  // Match jest.mock('module', factory, { virtual: true })
  const virtualPattern =
    /jest\.mock\s*\(\s*(['"][^'"]+['"]),\s*\(\)\s*=>\s*\([^)]*\),\s*\{\s*virtual:\s*true\s*\}\s*\)/g;

  return source.replace(virtualPattern, (match, modulePath) => {
    const todoComment = formatter.formatTodo({
      id: 'UNCONVERTIBLE-VIRTUAL-MOCK',
      description:
        'Vitest does not support { virtual: true } option in vi.mock()',
      original: match,
      action: 'Create the module file or use vi.mock with a manual factory',
    });
    return (
      todoComment +
      '\n' +
      match
        .replace('jest.mock', 'vi.mock')
        .replace(/,\s*\{\s*virtual:\s*true\s*\}/, '')
    );
  });
}

/**
 * Convert jest.retryTimes to HAMLET-TODO.
 */
function convertRetryTimes(source, todos) {
  return source.replace(
    /jest\.retryTimes\s*\(\s*(\d+)\s*\)\s*;?/g,
    (match, count) => {
      const todoComment = formatter.formatWarning({
        description:
          `Vitest uses \`retry\` in config or per-describe options instead of jest.retryTimes(${count}). ` +
          'Configure retry in vitest.config.ts or use describe("suite", () => {}, { retry: ' +
          count +
          ' }).',
        original: match,
      });
      return todoComment + '\n// ' + match;
    }
  );
}

/**
 * Convert expect.addSnapshotSerializer to HAMLET-TODO.
 */
function convertSnapshotSerializer(source, todos) {
  return source.replace(
    /expect\.addSnapshotSerializer\s*\([^)]*\)\s*;?/g,
    (match) => {
      return (
        formatter.formatTodo({
          id: 'UNCONVERTIBLE-SNAPSHOT-SERIALIZER',
          description:
            'Vitest snapshot serializers are configured in vitest.config.ts, not in test files',
          original: match,
          action:
            'Move serializer config to vitest.config.ts snapshotSerializers option',
        }) +
        '\n// ' +
        match
      );
    }
  );
}

/**
 * Add mock hoisting warning for vi.mock() calls that appear after imports.
 */
function addMockHostingWarning(source, warnings) {
  // Only add warning if there's a vi.mock after import statements
  const lines = source.split('\n');
  let hasImportBefore = false;
  let mockLine = -1;

  for (let i = 0; i < lines.length; i++) {
    const trimmed = lines[i].trim();
    if (/^import\s/.test(trimmed)) hasImportBefore = true;
    if (/^vi\.mock\s*\(/.test(trimmed) && hasImportBefore) {
      mockLine = i;
      break;
    }
  }

  if (mockLine >= 0 && !source.includes('HAMLET-WARNING')) {
    const warning = formatter.formatWarning({
      description:
        'vi.mock is hoisted like jest.mock, but factory function scoping ' +
        'differs. Variables defined above vi.mock() are not accessible inside the factory. ' +
        'Verify this mock works correctly.',
      original: lines[mockLine].trim(),
    });
    lines.splice(mockLine, 0, warning);
    warnings.push('Mock hoisting behavior may differ');
    return lines.join('\n');
  }

  return source;
}

/**
 * Add warning comment before toMatchSnapshot() calls.
 */
function addSnapshotWarning(source, warnings) {
  if (source.includes('HAMLET-WARNING') && source.includes('snapshot')) {
    return source; // Already warned
  }

  const lines = source.split('\n');
  for (let i = 0; i < lines.length; i++) {
    if (
      /\.toMatchSnapshot\s*\(/.test(lines[i]) &&
      !lines[i].includes('HAMLET')
    ) {
      const warning =
        '// HAMLET-WARNING: Snapshot file location and format may differ between\n' +
        '// Jest (__snapshots__/*.snap) and Vitest. Re-run tests to regenerate snapshots.';
      // Insert warning before the first snapshot assertion only
      const describeIdx = findPrecedingDescribe(lines, i);
      if (
        describeIdx >= 0 &&
        !lines.slice(describeIdx, i).some((l) => l.includes('HAMLET-WARNING'))
      ) {
        lines.splice(describeIdx, 0, warning);
      }
      warnings.push('Snapshot files need regeneration');
      break;
    }
  }
  return lines.join('\n');
}

/**
 * Find the describe() line preceding a given line index.
 */
function findPrecedingDescribe(lines, fromIdx) {
  for (let i = fromIdx - 1; i >= 0; i--) {
    if (/\bdescribe\s*\(/.test(lines[i])) return i;
  }
  return 0;
}

/**
 * Check if the IR has auto-mock patterns (__mocks__ directory usage).
 */
function hasAutomockPattern(ir) {
  // Check for jest.mock without factory (relies on __mocks__ directory)
  if (!ir || !ir.body) return false;
  for (const node of ir.body) {
    if (node instanceof MockCall && node.kind === 'mockModule') {
      // Simple jest.mock('path') without factory suggests __mocks__ usage
      if (
        node.originalSource &&
        !/jest\.mock\s*\([^,]+,/.test(node.originalSource)
      ) {
        return true;
      }
    }
  }
  return false;
}

/**
 * Analyze converted source to determine which vitest exports are used.
 * Generate the import statement.
 */
function generateVitestImport(source) {
  const used = new Set();

  // Use line-start or whitespace anchor to avoid matching 'test' inside words
  // like 'authentication'. \b alone isn't enough because 'test(' matches 'test' at
  // any word boundary including mid-identifier positions.
  if (/(?:^|[\s{;,])describe\s*[\.(]/m.test(source)) used.add('describe');
  if (/(?:^|[\s{;,])it\s*[\.(]/m.test(source)) used.add('it');
  if (/(?:^|[\s{;,])test\s*[\.(]/m.test(source)) used.add('test');
  if (/(?:^|[\s{;,])expect\s*[\.(]/m.test(source)) used.add('expect');
  if (/(?:^|[\s{;,])beforeEach\s*\(/m.test(source)) used.add('beforeEach');
  if (/(?:^|[\s{;,])afterEach\s*\(/m.test(source)) used.add('afterEach');
  if (/(?:^|[\s{;,])beforeAll\s*\(/m.test(source)) used.add('beforeAll');
  if (/(?:^|[\s{;,])afterAll\s*\(/m.test(source)) used.add('afterAll');
  if (/\bvi\./.test(source)) used.add('vi');

  if (used.size === 0) return '';

  // Sort for consistent output: describe, it, test, expect, vi, hooks
  const order = [
    'describe',
    'it',
    'test',
    'expect',
    'vi',
    'beforeEach',
    'afterEach',
    'beforeAll',
    'afterAll',
  ];
  const sorted = order.filter((name) => used.has(name));

  return `import { ${sorted.join(', ')} } from 'vitest';`;
}

/**
 * Prepend the vitest import to the source code.
 * Inserts after any existing non-vitest imports, or at the top.
 */
function prependImport(source, vitestImport) {
  if (!vitestImport) return source;

  const lines = source.split('\n');
  let insertIdx = 0;

  // Find the right position: after license/directive comments, before code
  // If there are existing imports, insert the vitest import before them
  for (let i = 0; i < lines.length; i++) {
    const trimmed = lines[i].trim();

    // Skip leading comments (license headers, directives)
    if (
      trimmed.startsWith('//') ||
      trimmed.startsWith('/*') ||
      trimmed.startsWith('*') ||
      trimmed === '*/'
    ) {
      insertIdx = i + 1;
      continue;
    }

    // Skip empty lines at the top
    if (trimmed === '' && i === insertIdx) {
      insertIdx = i + 1;
      continue;
    }

    break;
  }

  lines.splice(insertIdx, 0, vitestImport);

  // Ensure blank line after imports
  if (
    insertIdx + 1 < lines.length &&
    lines[insertIdx + 1].trim() !== '' &&
    !lines[insertIdx + 1].trim().startsWith('import')
  ) {
    lines.splice(insertIdx + 1, 0, '');
  }

  return lines.join('\n');
}

export default {
  name: 'vitest',
  language: 'javascript',
  paradigm: 'bdd',
  detect,
  parse,
  emit,
  imports: {
    explicit: [
      'describe',
      'it',
      'test',
      'expect',
      'vi',
      'beforeEach',
      'afterEach',
      'beforeAll',
      'afterAll',
    ],
    from: 'vitest',
    mockNamespace: 'vi',
  },
};
