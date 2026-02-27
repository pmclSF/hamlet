/**
 * Intermediate Representation (IR) node types for Hamlet v2.
 *
 * These nodes represent a normalized test file structure that is
 * framework-agnostic. Parsers produce IR from source code, and
 * emitters produce target framework code from IR.
 */

/**
 * Base class for all IR nodes.
 */
export class IRNode {
  /**
   * @param {string} type - Node type identifier
   * @param {Object} props
   * @param {Object|null} props.sourceLocation - { line, column, endLine, endColumn }
   * @param {'converted'|'unconvertible'|'warning'} props.confidence
   * @param {string} props.originalSource - Original source text for this node
   * @param {boolean} props.requiresAsync - Node requires async transformation
   * @param {boolean} props.hasTimingDependency - Node has timing-sensitive behavior
   * @param {boolean} props.frameworkSpecific - Node uses framework-specific API with no equivalent
   */
  constructor(type, props = {}) {
    this.type = type;
    this.sourceLocation = props.sourceLocation || null;
    this.confidence = props.confidence || 'converted';
    this.originalSource = props.originalSource || '';
    this.requiresAsync = props.requiresAsync || false;
    this.hasTimingDependency = props.hasTimingDependency || false;
    this.frameworkSpecific = props.frameworkSpecific || false;
  }
}

/**
 * Root node representing an entire test file.
 */
export class TestFile extends IRNode {
  constructor(props = {}) {
    super('TestFile', props);
    this.language = props.language || 'javascript';
    this.imports = props.imports || [];
    this.body = props.body || [];
  }
}

/**
 * A named group of tests (describe block, test class, etc.)
 */
export class TestSuite extends IRNode {
  constructor(props = {}) {
    super('TestSuite', props);
    this.name = props.name || '';
    this.hooks = props.hooks || [];
    this.tests = props.tests || [];
    this.sharedState = props.sharedState || [];
    this.modifiers = props.modifiers || [];
  }
}

/**
 * An individual test case.
 */
export class TestCase extends IRNode {
  constructor(props = {}) {
    super('TestCase', props);
    this.name = props.name || '';
    this.body = props.body || [];
    this.modifiers = props.modifiers || [];
    this.parameters = props.parameters || null;
    this.isAsync = props.isAsync || false;
  }
}

/**
 * A lifecycle hook (beforeAll, beforeEach, afterEach, afterAll, around).
 */
export class Hook extends IRNode {
  constructor(props = {}) {
    super('Hook', props);
    this.hookType = props.hookType || 'beforeEach';
    this.scope = props.scope || 'suite';
    this.body = props.body || [];
    this.isAsync = props.isAsync || false;
  }
}

/**
 * An assertion statement.
 */
export class Assertion extends IRNode {
  constructor(props = {}) {
    super('Assertion', props);
    this.kind = props.kind || 'equal';
    this.subject = props.subject || '';
    this.expected = props.expected !== undefined ? props.expected : null;
    this.isNegated = props.isNegated || false;
    this.message = props.message || null;
  }
}

/**
 * A mock, spy, or stub operation.
 */
export class MockCall extends IRNode {
  constructor(props = {}) {
    super('MockCall', props);
    this.kind = props.kind || 'createMock';
    this.target = props.target || '';
    this.args = props.args || [];
    this.returnValue =
      props.returnValue !== undefined ? props.returnValue : null;
  }
}

/**
 * An import or require statement.
 */
export class ImportStatement extends IRNode {
  constructor(props = {}) {
    super('ImportStatement', props);
    this.kind = props.kind || 'framework';
    this.source = props.source || '';
    this.specifiers = props.specifiers || [];
    this.isDefault = props.isDefault || false;
    this.isTypeOnly = props.isTypeOnly || false;
  }
}

/**
 * Code that passes through unconverted.
 */
export class RawCode extends IRNode {
  constructor(props = {}) {
    super('RawCode', props);
    this.code = props.code || '';
    this.comment = props.comment || null;
  }
}

/**
 * A preserved comment.
 */
export class Comment extends IRNode {
  constructor(props = {}) {
    super('Comment', props);
    this.text = props.text || '';
    this.commentKind = props.commentKind || 'inline';
    this.preserveExact = props.preserveExact || false;
  }
}

/**
 * A shared/lazy variable (RSpec let, instance var, etc.)
 */
export class SharedVariable extends IRNode {
  constructor(props = {}) {
    super('SharedVariable', props);
    this.name = props.name || '';
    this.initializer = props.initializer || '';
    this.isLazy = props.isLazy || false;
    this.scope = props.scope || 'instance';
  }
}

/**
 * A test modifier (skip, only, timeout, tag, etc.)
 */
export class Modifier extends IRNode {
  constructor(props = {}) {
    super('Modifier', props);
    this.modifierType = props.modifierType || 'skip';
    this.value = props.value !== undefined ? props.value : null;
    this.condition = props.condition || null;
  }
}

/**
 * A set of parameters for parameterized/data-driven tests.
 */
export class ParameterSet extends IRNode {
  constructor(props = {}) {
    super('ParameterSet', props);
    this.paramKind = props.paramKind || 'values';
    this.parameters = props.parameters || [];
    this.ids = props.ids || null;
  }
}

/**
 * Walk an IR tree depth-first, calling visitor(node) for each node.
 * Returns an array of all nodes visited.
 *
 * @param {IRNode} root
 * @param {Function} visitor - Called with each node. Return false to skip children.
 * @returns {IRNode[]} All visited nodes
 */
export function walkIR(root, visitor) {
  const visited = [];

  function visit(node) {
    if (!(node instanceof IRNode)) return;
    visited.push(node);
    const result = visitor ? visitor(node) : true;
    if (result === false) return;

    const childArrays = [
      node.imports,
      node.body,
      node.hooks,
      node.tests,
      node.sharedState,
      node.modifiers,
    ];

    for (const arr of childArrays) {
      if (Array.isArray(arr)) {
        for (const child of arr) {
          visit(child);
        }
      }
    }

    if (node.parameters instanceof IRNode) {
      visit(node.parameters);
    }
  }

  visit(root);
  return visited;
}
