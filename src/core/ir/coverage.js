/**
 * IR Coverage Metrics — tracks how many nodes can be IR-emitted.
 *
 * Covers all emittable node types: Assertion, Navigation, MockCall,
 * Hook, TestSuite, TestCase, Modifier, ImportStatement.
 */

import {
  Assertion,
  Navigation,
  MockCall,
  Hook,
  TestSuite,
  TestCase,
  Modifier,
  ImportStatement,
  walkIR,
} from '../ir.js';

/**
 * @typedef {Object} IRCoverage
 * @property {number} totalAssertions - Total assertion nodes in the IR
 * @property {number} supportedAssertions - Assertions with known IR emission
 * @property {number} unsupportedAssertions - Assertions without IR emission
 * @property {number} totalNavigation - Total navigation nodes in the IR
 * @property {number} supportedNavigation - Navigation nodes with known IR emission
 * @property {number} totalMockCalls - Total mock call nodes
 * @property {number} supportedMockCalls - Mock calls with known IR emission
 * @property {number} totalHooks - Total hook nodes
 * @property {number} supportedHooks - Hooks with known IR emission
 * @property {number} totalTestSuites - Total test suite nodes
 * @property {number} supportedTestSuites - Test suites with known IR emission
 * @property {number} totalTestCases - Total test case nodes
 * @property {number} supportedTestCases - Test cases with known IR emission
 * @property {number} totalModifiers - Total modifier nodes
 * @property {number} supportedModifiers - Modifiers with known IR emission
 * @property {number} totalImports - Total import statement nodes
 * @property {number} supportedImports - Import statements with known IR emission
 * @property {number} coveragePercent - Percentage of all supported IR nodes (0-100)
 * @property {string[]} unsupportedKinds - Unique kinds that lack IR support
 */

/** Known assertion kinds that the IR emitter supports */
const SUPPORTED_KINDS = new Set([
  'be.visible',
  'exist',
  'contain',
  'have.text',
  'have.length',
  'have.attr',
  'have.class',
  'have.value',
  'be.checked',
  'be.disabled',
  'be.enabled',
  'be.empty',
  'be.focused',
  'match',
  'url.include',
  'url.equal',
  'title.equal',
  'equal',
  'be.true',
  'be.false',
  'be.null',
  'be.undefined',
  'include',
  'have.property',
]);

/** Known navigation actions that the IR emitter supports */
const SUPPORTED_ACTIONS = new Set(['visit', 'goBack', 'goForward', 'reload']);

/** Known mock call kinds that the IR emitter supports */
const SUPPORTED_MOCK_KINDS = new Set(['networkIntercept']);

/**
 * Compute IR coverage metrics for an IR tree.
 *
 * @param {import('../ir.js').IRNode} ir - Root of the IR tree
 * @returns {IRCoverage}
 */
export function computeIRCoverage(ir) {
  let totalAssertions = 0;
  let supportedAssertions = 0;
  let totalNavigation = 0;
  let supportedNavigation = 0;
  let totalMockCalls = 0;
  let supportedMockCalls = 0;
  let totalHooks = 0;
  let supportedHooks = 0;
  let totalTestSuites = 0;
  let supportedTestSuites = 0;
  let totalTestCases = 0;
  let supportedTestCases = 0;
  let totalModifiers = 0;
  let supportedModifiers = 0;
  let totalImports = 0;
  let supportedImports = 0;
  const unsupportedKinds = new Set();

  walkIR(ir, (node) => {
    if (node instanceof Assertion) {
      totalAssertions++;
      const kind = node.kind || '';
      const hasSubject =
        typeof node.subject === 'string' && node.subject !== '';
      if (SUPPORTED_KINDS.has(kind) && hasSubject) {
        supportedAssertions++;
      } else if (kind) {
        unsupportedKinds.add(kind);
      }
    } else if (node instanceof Navigation) {
      totalNavigation++;
      if (SUPPORTED_ACTIONS.has(node.action)) {
        if (node.action === 'visit' && !node.url) {
          unsupportedKinds.add('navigation:visit-no-url');
        } else {
          supportedNavigation++;
        }
      } else {
        unsupportedKinds.add(`navigation:${node.action}`);
      }
    } else if (node instanceof MockCall) {
      totalMockCalls++;
      const mockKind = node.kind || '';
      if (SUPPORTED_MOCK_KINDS.has(mockKind) && node.target) {
        supportedMockCalls++;
      } else if (mockKind) {
        unsupportedKinds.add(`mock:${mockKind}`);
      }
    } else if (node instanceof Hook) {
      totalHooks++;
      supportedHooks++;
    } else if (node instanceof TestSuite) {
      totalTestSuites++;
      supportedTestSuites++;
    } else if (node instanceof TestCase) {
      totalTestCases++;
      supportedTestCases++;
    } else if (node instanceof Modifier) {
      totalModifiers++;
      supportedModifiers++;
    } else if (node instanceof ImportStatement) {
      totalImports++;
      supportedImports++;
    }
  });

  // Coverage is based on emittable leaf nodes
  const totalNodes = totalAssertions + totalNavigation + totalMockCalls;
  const supportedNodes =
    supportedAssertions + supportedNavigation + supportedMockCalls;
  const percent =
    totalNodes > 0 ? Math.round((supportedNodes / totalNodes) * 100) : 100;

  return {
    totalAssertions,
    supportedAssertions,
    unsupportedAssertions: totalAssertions - supportedAssertions,
    totalNavigation,
    supportedNavigation,
    totalMockCalls,
    supportedMockCalls,
    totalHooks,
    supportedHooks,
    totalTestSuites,
    supportedTestSuites,
    totalTestCases,
    supportedTestCases,
    totalModifiers,
    supportedModifiers,
    totalImports,
    supportedImports,
    coveragePercent: percent,
    unsupportedKinds: [...unsupportedKinds],
  };
}
