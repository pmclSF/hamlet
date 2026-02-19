/**
 * Round-trip test helper.
 *
 * Provides convert() and semantic equivalence checking for round-trip tests.
 * Semantic equivalence means same test structure (suite/test/assertion/hook counts)
 * not identical strings (formatting, import style will differ).
 */

import { ConverterFactory } from '../../../src/core/ConverterFactory.js';

/**
 * Convert content from one framework to another.
 * @param {string} content - Source code
 * @param {string} from - Source framework
 * @param {string} to - Target framework
 * @returns {Promise<{code: string, report: Object|null}>}
 */
export async function convert(content, from, to) {
  const converter = await ConverterFactory.createConverter(from, to);
  const code = await converter.convert(content);
  const report = converter.getLastReport ? converter.getLastReport() : null;
  return { code, report };
}

// ── Counting helpers (regex-based, framework-aware) ──────────────────

const TEST_PATTERNS = {
  jest: [/\bit\s*\(/, /\btest\s*\(/, /\btest\.each/],
  vitest: [/\bit\s*\(/, /\btest\s*\(/, /\btest\.each/],
  mocha: [/\bit\s*\(/, /\bspecify\s*\(/],
  jasmine: [/\bit\s*\(/, /\bfit\s*\(/],
  cypress: [/\bit\s*\(/],
  playwright: [/\btest\s*\(/, /\btest\.describe/],
  webdriverio: [/\bit\s*\(/, /\btest\s*\(/],
  puppeteer: [/\bit\s*\(/, /\btest\s*\(/],
  testcafe: [/\btest\s*\(/],
  junit4: [/@Test/],
  junit5: [/@Test/, /@ParameterizedTest/],
  testng: [/@Test/],
  pytest: [/\bdef test_/],
  unittest: [/\bdef test_/],
};

const SUITE_PATTERNS = {
  jest: [/\bdescribe\s*\(/],
  vitest: [/\bdescribe\s*\(/],
  mocha: [/\bdescribe\s*\(/, /\bcontext\s*\(/],
  jasmine: [/\bdescribe\s*\(/, /\bfdescribe\s*\(/],
  cypress: [/\bdescribe\s*\(/, /\bcontext\s*\(/],
  playwright: [/\btest\.describe\s*\(/],
  webdriverio: [/\bdescribe\s*\(/],
  puppeteer: [/\bdescribe\s*\(/],
  testcafe: [/\bfixture\s*[(`]/],
  junit4: [/\bpublic\s+class\s+/],
  junit5: [/\bclass\s+/, /@Nested/],
  testng: [/\bpublic\s+class\s+/],
  pytest: [/\bclass\s+Test/],
  unittest: [/\bclass\s+Test/],
};

const ASSERTION_PATTERNS = {
  jest: [/\bexpect\s*\(/, /\bassert\s*[.(]/],
  vitest: [/\bexpect\s*\(/, /\bassert\s*[.(]/],
  mocha: [/\bexpect\s*\(/, /\bassert\s*[.(]/, /\.should\b/],
  jasmine: [/\bexpect\s*\(/],
  cypress: [/\.should\s*\(/, /\bexpect\s*\(/],
  playwright: [/\bexpect\s*\(/],
  webdriverio: [/\bexpect\s*\(/, /\.should\b/],
  puppeteer: [/\bexpect\s*\(/],
  testcafe: [/\bt\.expect\s*\(/],
  junit4: [/\bassert\w+\s*\(/, /\bAssert\.\w+\s*\(/],
  junit5: [/\bassert\w+\s*\(/, /\bAssertions\.\w+\s*\(/],
  testng: [/\bAssert\.\w+\s*\(/],
  pytest: [/\bassert\s+/],
  unittest: [/\bself\.assert\w+\s*\(/],
};

const HOOK_PATTERNS = {
  jest: [/\bbeforeEach\s*\(/, /\bafterEach\s*\(/, /\bbeforeAll\s*\(/, /\bafterAll\s*\(/],
  vitest: [/\bbeforeEach\s*\(/, /\bafterEach\s*\(/, /\bbeforeAll\s*\(/, /\bafterAll\s*\(/],
  mocha: [/\bbefore\s*\(/, /\bafter\s*\(/, /\bbeforeEach\s*\(/, /\bafterEach\s*\(/],
  jasmine: [/\bbeforeEach\s*\(/, /\bafterEach\s*\(/, /\bbeforeAll\s*\(/, /\bafterAll\s*\(/],
  cypress: [/\bbefore\s*\(/, /\bafter\s*\(/, /\bbeforeEach\s*\(/, /\bafterEach\s*\(/],
  playwright: [/\btest\.beforeEach\s*\(/, /\btest\.afterEach\s*\(/, /\btest\.beforeAll\s*\(/, /\btest\.afterAll\s*\(/],
  webdriverio: [/\bbefore\s*\(/, /\bafter\s*\(/, /\bbeforeEach\s*\(/, /\bafterEach\s*\(/],
  puppeteer: [/\bbeforeAll\s*\(/, /\bafterAll\s*\(/, /\bbeforeEach\s*\(/, /\bafterEach\s*\(/],
  testcafe: [/\bfixture\b.*\bbefore\b/, /\bfixture\b.*\bafter\b/],
  junit4: [/@Before\b/, /@After\b/, /@BeforeClass/, /@AfterClass/],
  junit5: [/@BeforeEach/, /@AfterEach/, /@BeforeAll/, /@AfterAll/],
  testng: [/@BeforeMethod/, /@AfterMethod/, /@BeforeClass/, /@AfterClass/],
  pytest: [/@pytest\.fixture/],
  unittest: [/\bdef setUp\b/, /\bdef tearDown\b/, /\bdef setUpClass\b/, /\bdef tearDownClass\b/],
};

/**
 * Count occurrences of any pattern in an array against content.
 */
function countMatches(content, patterns) {
  let total = 0;
  for (const pat of patterns) {
    const matches = content.match(new RegExp(pat.source, 'g' + (pat.flags || '')));
    if (matches) total += matches.length;
  }
  return total;
}

/**
 * Parse structural info from source code for a given framework.
 * @param {string} content
 * @param {string} framework
 * @returns {{suiteCount: number, testCount: number, assertionCount: number, hookCount: number}}
 */
export function parseStructure(content, framework) {
  const fw = framework.toLowerCase();
  return {
    suiteCount: countMatches(content, SUITE_PATTERNS[fw] || []),
    testCount: countMatches(content, TEST_PATTERNS[fw] || []),
    assertionCount: countMatches(content, ASSERTION_PATTERNS[fw] || []),
    hookCount: countMatches(content, HOOK_PATTERNS[fw] || []),
  };
}

/**
 * Assert semantic equivalence between original and round-tripped code.
 * @param {Object} original - parseStructure result for original
 * @param {Object} roundTripped - parseStructure result for round-tripped
 * @param {Object} [options]
 * @param {number} [options.testTolerance=0] - Allowed difference in test count
 * @param {number} [options.assertionTolerance=0] - Allowed difference in assertion count
 */
export function assertSemanticEquivalence(original, roundTripped, options = {}) {
  const testTol = options.testTolerance || 0;
  const assertTol = options.assertionTolerance || 0;

  expect(roundTripped.suiteCount).toBeGreaterThanOrEqual(original.suiteCount - 1);
  expect(roundTripped.testCount).toBeGreaterThanOrEqual(original.testCount - testTol);
  expect(roundTripped.testCount).toBeLessThanOrEqual(original.testCount + testTol);
  expect(roundTripped.assertionCount).toBeGreaterThanOrEqual(original.assertionCount - assertTol);
}
