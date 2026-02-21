/**
 * Phase 3 tests: JS unit target emitters (Jest, Vitest, Mocha, Jasmine).
 *
 * Covers:
 * - Jest parser: nested IR with subject/expected extraction
 * - Jest ir-full emitter: describe/it blocks, assertions, mocks
 * - Vitest ir-full emitter: vitest imports, vi.* mocking
 * - Mocha ir-full emitter: Chai assertions, Sinon mocking, Mocha hooks
 * - Jasmine ir-full emitter: Jasmine matchers, spy API, fdescribe/xit
 * - Dynamic emitter loading for all four targets
 */

import {
  TestFile,
  TestSuite,
  TestCase,
  Hook,
  Assertion,
  MockCall,
  RawCode,
  Comment,
  ImportStatement,
  SharedVariable,
  Modifier,
  walkIR,
} from '../../src/core/ir.js';

// ═══════════════════════════════════════════════════════════════════════
// Jest parser: nested IR with subject/expected extraction
// ═══════════════════════════════════════════════════════════════════════

describe('Jest parser — nested IR', () => {
  let parse;

  beforeEach(async () => {
    const jestFw = (
      await import('../../src/languages/javascript/frameworks/jest.js')
    ).default;
    parse = jestFw.parse;
  });

  it('should produce nested TestSuite with TestCase children', () => {
    const ir = parse(`
describe('Math', () => {
  it('adds numbers', () => {
    expect(1 + 1).toBe(2);
  });
});
    `);
    expect(ir).toBeInstanceOf(TestFile);
    expect(ir.body.length).toBe(1);
    const suite = ir.body[0];
    expect(suite).toBeInstanceOf(TestSuite);
    expect(suite.name).toBe('Math');
    expect(suite.tests.length).toBe(1);
    expect(suite.tests[0]).toBeInstanceOf(TestCase);
    expect(suite.tests[0].name).toBe('adds numbers');
  });

  it('should nest hooks inside TestSuite.hooks[]', () => {
    const ir = parse(`
describe('Setup', () => {
  beforeEach(() => {
    jest.useFakeTimers();
  });
  afterEach(() => {
    jest.useRealTimers();
  });
  it('works', () => {});
});
    `);
    const suite = ir.body[0];
    expect(suite.hooks.length).toBe(2);
    expect(suite.hooks[0]).toBeInstanceOf(Hook);
    expect(suite.hooks[0].hookType).toBe('beforeEach');
    expect(suite.hooks[1].hookType).toBe('afterEach');
  });

  it('should nest assertions inside TestCase.body[]', () => {
    const ir = parse(`
describe('Suite', () => {
  it('checks value', () => {
    expect(result).toBe(42);
    expect(list).toHaveLength(3);
  });
});
    `);
    const tc = ir.body[0].tests[0];
    expect(tc.body.length).toBe(2);
    expect(tc.body[0]).toBeInstanceOf(Assertion);
    expect(tc.body[1]).toBeInstanceOf(Assertion);
  });

  it('should extract subject from expect(subject)', () => {
    const ir = parse(`
describe('Suite', () => {
  it('test', () => {
    expect(result).toBe(42);
  });
});
    `);
    const assertion = ir.body[0].tests[0].body[0];
    expect(assertion.subject).toBe('result');
  });

  it('should extract subject with nested parens', () => {
    const ir = parse(`
describe('Suite', () => {
  it('test', () => {
    expect(getValue(a, b)).toBe(true);
  });
});
    `);
    const assertion = ir.body[0].tests[0].body[0];
    expect(assertion.subject).toBe('getValue(a, b)');
  });

  it('should extract expected from matcher args', () => {
    const ir = parse(`
describe('Suite', () => {
  it('test', () => {
    expect(x).toBe(42);
    expect(arr).toHaveLength(5);
    expect(str).toContain('hello');
  });
});
    `);
    const body = ir.body[0].tests[0].body;
    expect(body[0].expected).toBe('42');
    expect(body[1].expected).toBe('5');
    expect(body[2].expected).toBe("'hello'");
  });

  it('should detect negation', () => {
    const ir = parse(`
describe('Suite', () => {
  it('test', () => {
    expect(x).not.toBe(null);
  });
});
    `);
    const assertion = ir.body[0].tests[0].body[0];
    expect(assertion.isNegated).toBe(true);
    expect(assertion.kind).toBe('strictEqual');
  });

  it('should handle nested describes', () => {
    const ir = parse(`
describe('Outer', () => {
  describe('Inner', () => {
    it('nested test', () => {});
  });
});
    `);
    const outer = ir.body[0];
    expect(outer.name).toBe('Outer');
    expect(outer.tests.length).toBe(1);
    const inner = outer.tests[0];
    expect(inner).toBeInstanceOf(TestSuite);
    expect(inner.name).toBe('Inner');
    expect(inner.tests.length).toBe(1);
    expect(inner.tests[0].name).toBe('nested test');
  });

  it('should extract modifiers (.only, .skip, .todo)', () => {
    const ir = parse(`
describe.only('focused', () => {
  it.skip('skipped', () => {});
  it.todo('pending');
});
    `);
    const suite = ir.body[0];
    expect(suite.modifiers[0].modifierType).toBe('only');
    const tc1 = suite.tests[0];
    expect(tc1.modifiers[0].modifierType).toBe('skip');
  });

  it('should nest MockCall inside TestCase body', () => {
    const ir = parse(`
describe('Suite', () => {
  it('mocks', () => {
    const fn = jest.fn();
    expect(fn).toHaveBeenCalled();
  });
});
    `);
    const body = ir.body[0].tests[0].body;
    expect(body[0]).toBeInstanceOf(MockCall);
    expect(body[0].kind).toBe('createMock');
    expect(body[1]).toBeInstanceOf(Assertion);
    expect(body[1].kind).toBe('called');
  });
});

// ═══════════════════════════════════════════════════════════════════════
// Jest ir-full emitter
// ═══════════════════════════════════════════════════════════════════════

describe('Jest ir-full emitter', () => {
  let emitFullFile, emitAssertion, emitNode;

  beforeEach(async () => {
    const mod = await import('../../src/core/emitters/jest/irEmitter.js');
    emitFullFile = mod.emitFullFile;
    emitAssertion = mod.emitAssertion;
    emitNode = mod.emitNode;
  });

  it('should return null for non-TestFile input', () => {
    expect(emitFullFile({})).toBeNull();
    expect(emitFullFile(null)).toBeNull();
  });

  it('should emit describe/it blocks', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'Math',
          tests: [new TestCase({ name: 'adds' })],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain("describe('Math', () => {");
    expect(result).toContain("it('adds', () => {");
  });

  it('should emit hooks without test. prefix', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'Suite',
          hooks: [new Hook({ hookType: 'beforeEach' })],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain('beforeEach(() => {');
    expect(result).not.toContain('test.beforeEach');
  });

  it('should emit Jest assertions', () => {
    const node = new Assertion({
      kind: 'strictEqual',
      subject: 'result',
      expected: '42',
    });
    const result = emitAssertion(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe('expect(result).toBe(42)');
  });

  it('should emit negated assertions', () => {
    const node = new Assertion({
      kind: 'isNull',
      subject: 'x',
      isNegated: true,
    });
    const result = emitAssertion(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe('expect(x).not.toBeNull()');
  });

  it('should emit jest.fn() for createMock', () => {
    const node = new MockCall({ kind: 'createMock' });
    const result = emitNode(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe('jest.fn()');
  });

  it('should emit jest.mock() for mockModule', () => {
    const node = new MockCall({ kind: 'mockModule', target: 'fs' });
    const result = emitNode(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe("jest.mock('fs')");
  });

  it('should emit describe.only and it.skip', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'Focused',
          modifiers: [new Modifier({ modifierType: 'only' })],
          tests: [
            new TestCase({
              name: 'skipped',
              modifiers: [new Modifier({ modifierType: 'skip' })],
            }),
          ],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain("describe.only('Focused'");
    expect(result).toContain("it.skip('skipped'");
  });
});

// ═══════════════════════════════════════════════════════════════════════
// Vitest ir-full emitter
// ═══════════════════════════════════════════════════════════════════════

describe('Vitest ir-full emitter', () => {
  let emitFullFile, emitNode;

  beforeEach(async () => {
    const mod = await import('../../src/core/emitters/vitest/irEmitter.js');
    emitFullFile = mod.emitFullFile;
    emitNode = mod.emitNode;
  });

  it('should emit vitest import', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'Suite',
          tests: [new TestCase({ name: 'test' })],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain("from 'vitest'");
    expect(result).toContain('describe');
    expect(result).toContain('it');
    expect(result).toContain('expect');
  });

  it('should include vi in import when mocks are present', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'Suite',
          tests: [
            new TestCase({
              name: 'test',
              body: [new MockCall({ kind: 'createMock' })],
            }),
          ],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain('vi');
    expect(result).toContain("from 'vitest'");
  });

  it('should include hook imports when hooks are present', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'Suite',
          hooks: [new Hook({ hookType: 'beforeEach' })],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain('beforeEach');
    expect(result).toContain("from 'vitest'");
  });

  it('should emit vi.fn() for createMock', () => {
    const node = new MockCall({ kind: 'createMock' });
    const result = emitNode(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe('vi.fn()');
  });

  it('should emit vi.mock() for mockModule', () => {
    const node = new MockCall({ kind: 'mockModule', target: 'fs' });
    const result = emitNode(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe("vi.mock('fs')");
  });

  it('should emit assertions in vitest syntax', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'Suite',
          tests: [
            new TestCase({
              name: 'checks',
              body: [
                new Assertion({
                  kind: 'strictEqual',
                  subject: 'x',
                  expected: '1',
                }),
              ],
            }),
          ],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain('expect(x).toBe(1);');
  });
});

// ═══════════════════════════════════════════════════════════════════════
// Mocha ir-full emitter
// ═══════════════════════════════════════════════════════════════════════

describe('Mocha ir-full emitter', () => {
  let emitFullFile, emitAssertion, emitNode;

  beforeEach(async () => {
    const mod = await import('../../src/core/emitters/mocha/irEmitter.js');
    emitFullFile = mod.emitFullFile;
    emitAssertion = mod.emitAssertion;
    emitNode = mod.emitNode;
  });

  it('should emit chai require import', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'Suite',
          tests: [new TestCase({ name: 'test' })],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain("require('chai')");
    expect(result).toContain('expect');
  });

  it('should emit sinon import when mocks are present', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'Suite',
          tests: [
            new TestCase({
              name: 'test',
              body: [new MockCall({ kind: 'createMock' })],
            }),
          ],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain("require('sinon')");
  });

  it('should emit Chai assertion chains', () => {
    const result = emitAssertion(
      new Assertion({
        kind: 'strictEqual',
        subject: 'x',
        expected: '42',
      })
    );
    expect(result.supported).toBe(true);
    expect(result.code).toBe('expect(x).to.equal(42)');
  });

  it('should emit negated Chai chains', () => {
    const result = emitAssertion(
      new Assertion({
        kind: 'isNull',
        subject: 'val',
        isNegated: true,
      })
    );
    expect(result.supported).toBe(true);
    expect(result.code).toBe('expect(val).to.not.be.null');
  });

  it('should use before/after instead of beforeAll/afterAll', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'Suite',
          hooks: [
            new Hook({ hookType: 'beforeAll' }),
            new Hook({ hookType: 'afterAll' }),
          ],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain('before(');
    expect(result).toContain('after(');
    expect(result).not.toContain('beforeAll(');
    expect(result).not.toContain('afterAll(');
  });

  it('should use function() instead of arrow functions', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'Suite',
          tests: [new TestCase({ name: 'test' })],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain('function ()');
  });

  it('should emit sinon.stub() for createMock', () => {
    const node = new MockCall({ kind: 'createMock' });
    const result = emitNode(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe('sinon.stub()');
  });
});

// ═══════════════════════════════════════════════════════════════════════
// Jasmine ir-full emitter
// ═══════════════════════════════════════════════════════════════════════

describe('Jasmine ir-full emitter', () => {
  let emitFullFile, emitAssertion, emitNode;

  beforeEach(async () => {
    const mod = await import('../../src/core/emitters/jasmine/irEmitter.js');
    emitFullFile = mod.emitFullFile;
    emitAssertion = mod.emitAssertion;
    emitNode = mod.emitNode;
  });

  it('should emit describe/it blocks with function()', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'Calculator',
          tests: [new TestCase({ name: 'adds' })],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain("describe('Calculator', function () {");
    expect(result).toContain("it('adds', function () {");
  });

  it('should use fdescribe/xdescribe for modifiers', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'Focused',
          modifiers: [new Modifier({ modifierType: 'only' })],
        }),
        new TestSuite({
          name: 'Skipped',
          modifiers: [new Modifier({ modifierType: 'skip' })],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain("fdescribe('Focused'");
    expect(result).toContain("xdescribe('Skipped'");
  });

  it('should use fit/xit for test modifiers', () => {
    const ir = new TestFile({
      body: [
        new TestSuite({
          name: 'Suite',
          tests: [
            new TestCase({
              name: 'focused',
              modifiers: [new Modifier({ modifierType: 'only' })],
            }),
            new TestCase({
              name: 'skipped',
              modifiers: [new Modifier({ modifierType: 'skip' })],
            }),
          ],
        }),
      ],
    });
    const result = emitFullFile(ir);
    expect(result).toContain("fit('focused'");
    expect(result).toContain("xit('skipped'");
  });

  it('should emit Jasmine assertions', () => {
    const result = emitAssertion(
      new Assertion({
        kind: 'strictEqual',
        subject: 'x',
        expected: '42',
      })
    );
    expect(result.supported).toBe(true);
    expect(result.code).toBe('expect(x).toBe(42)');
  });

  it('should emit jasmine.createSpy() for createMock', () => {
    const node = new MockCall({ kind: 'createMock' });
    const result = emitNode(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe('jasmine.createSpy()');
  });

  it('should emit jasmine.clock().install() for fakeTimers', () => {
    const node = new MockCall({ kind: 'fakeTimers' });
    const result = emitNode(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe('jasmine.clock().install()');
  });
});

// ═══════════════════════════════════════════════════════════════════════
// Dynamic emitter loading
// ═══════════════════════════════════════════════════════════════════════

describe('Dynamic emitter loading', () => {
  it('should load jest emitter', async () => {
    const emitter = await import('../../src/core/emitters/jest/irEmitter.js');
    expect(typeof emitter.emitNode).toBe('function');
    expect(typeof emitter.emitFullFile).toBe('function');
    expect(typeof emitter.matchesBaseline).toBe('function');
  });

  it('should load vitest emitter', async () => {
    const emitter = await import('../../src/core/emitters/vitest/irEmitter.js');
    expect(typeof emitter.emitNode).toBe('function');
    expect(typeof emitter.emitFullFile).toBe('function');
    expect(typeof emitter.matchesBaseline).toBe('function');
  });

  it('should load mocha emitter', async () => {
    const emitter = await import('../../src/core/emitters/mocha/irEmitter.js');
    expect(typeof emitter.emitNode).toBe('function');
    expect(typeof emitter.emitFullFile).toBe('function');
    expect(typeof emitter.matchesBaseline).toBe('function');
  });

  it('should load jasmine emitter', async () => {
    const emitter =
      await import('../../src/core/emitters/jasmine/irEmitter.js');
    expect(typeof emitter.emitNode).toBe('function');
    expect(typeof emitter.emitFullFile).toBe('function');
    expect(typeof emitter.matchesBaseline).toBe('function');
  });
});
