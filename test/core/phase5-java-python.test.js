/**
 * Phase 5 tests: Java + Python parser enhancements and ir-full emitters.
 *
 * Covers:
 * - JUnit4 parser: nested IR, assertion subject/expected extraction
 * - JUnit5 parser: nested IR, assertion subject/expected extraction
 * - TestNG parser: nested IR, assertion subject/expected extraction
 * - pytest parser: nested IR, assertion subject/expected extraction
 * - unittest parser: nested IR, assertion subject/expected extraction
 * - nose2 parser: nested IR, assertion subject/expected extraction
 * - JUnit5 ir-full emitter: emitFullFile, emitAssertion, emitNode, matchesBaseline
 * - TestNG ir-full emitter: emitFullFile, emitAssertion, emitNode, matchesBaseline
 * - pytest ir-full emitter: emitFullFile, emitAssertion, emitNode, matchesBaseline
 * - unittest ir-full emitter: emitFullFile, emitAssertion, emitNode, matchesBaseline
 * - Dynamic emitter loading for all 4 targets
 */

import junit4Fw from '../../src/languages/java/frameworks/junit4.js';
import junit5Fw from '../../src/languages/java/frameworks/junit5.js';
import testngFw from '../../src/languages/java/frameworks/testng.js';
import pytestFw from '../../src/languages/python/frameworks/pytest.js';
import unittestFw from '../../src/languages/python/frameworks/unittest_fw.js';
import nose2Fw from '../../src/languages/python/frameworks/nose2.js';

import * as junit5Emitter from '../../src/core/emitters/junit5/irEmitter.js';
import * as testngEmitter from '../../src/core/emitters/testng/irEmitter.js';
import * as pytestEmitter from '../../src/core/emitters/pytest/irEmitter.js';
import * as unittestEmitter from '../../src/core/emitters/unittest/irEmitter.js';

import {
  TestFile,
  TestSuite,
  TestCase,
  Hook,
  Assertion,
  RawCode,
  Comment,
  Modifier,
} from '../../src/core/ir.js';

// ═══════════════════════════════════════════════════════════════════════
// JUnit4 Parser Tests
// ═══════════════════════════════════════════════════════════════════════

describe('JUnit4 parser — nested IR', () => {
  it('should produce TestFile with nested TestSuite > TestCase > Assertion', () => {
    const source = `
import org.junit.Test;
import static org.junit.Assert.*;

public class CalcTest {
    @Test
    public void testAdd() {
        assertEquals(4, add(2, 2));
    }

    @Test
    public void testSub() {
        assertNotEquals(0, sub(5, 3));
    }
}`;
    const ir = junit4Fw.parse(source);
    expect(ir).toBeInstanceOf(TestFile);
    expect(ir.body.length).toBeGreaterThanOrEqual(1);

    const suite = ir.body.find((n) => n instanceof TestSuite);
    expect(suite).toBeDefined();
    expect(suite.name).toBe('CalcTest');
    expect(suite.tests.length).toBe(2);

    const tc1 = suite.tests[0];
    expect(tc1).toBeInstanceOf(TestCase);
    expect(tc1.name).toBe('testAdd');
    expect(tc1.body.length).toBeGreaterThanOrEqual(1);

    const assertion = tc1.body.find((n) => n instanceof Assertion);
    expect(assertion).toBeDefined();
    expect(assertion.kind).toBe('equal');
  });

  it('should extract subject/expected from assertEquals(expected, actual)', () => {
    const source = `
import org.junit.Test;
import static org.junit.Assert.*;

public class ArgTest {
    @Test
    public void testArgs() {
        assertEquals(42, compute());
    }
}`;
    const ir = junit4Fw.parse(source);
    const suite = ir.body.find((n) => n instanceof TestSuite);
    const tc = suite.tests.find((n) => n instanceof TestCase);
    const assertion = tc.body.find((n) => n instanceof Assertion);
    expect(assertion.kind).toBe('equal');
    expect(assertion.expected).toBe('42');
    expect(assertion.subject).toBe('compute()');
  });

  it('should parse @Before and @After as hooks', () => {
    const source = `
import org.junit.Test;
import org.junit.Before;
import org.junit.After;

public class HookTest {
    @Before
    public void setUp() {
        init();
    }

    @After
    public void tearDown() {
        cleanup();
    }

    @Test
    public void testIt() {
        assertTrue(true);
    }
}`;
    const ir = junit4Fw.parse(source);
    const suite = ir.body.find((n) => n instanceof TestSuite);
    expect(suite.hooks.length).toBe(2);
    expect(suite.hooks.some((h) => h.hookType === 'beforeEach')).toBe(true);
    expect(suite.hooks.some((h) => h.hookType === 'afterEach')).toBe(true);
  });
});

// ═══════════════════════════════════════════════════════════════════════
// JUnit5 Parser Tests
// ═══════════════════════════════════════════════════════════════════════

describe('JUnit5 parser — nested IR', () => {
  it('should produce nested TestSuite > TestCase > Assertion', () => {
    const source = `
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

class CalculatorTest {
    @Test
    void testMultiply() {
        Assertions.assertEquals(6, multiply(2, 3));
    }
}`;
    const ir = junit5Fw.parse(source);
    expect(ir).toBeInstanceOf(TestFile);

    const suite = ir.body.find((n) => n instanceof TestSuite);
    expect(suite).toBeDefined();
    expect(suite.name).toBe('CalculatorTest');

    const tc = suite.tests.find((n) => n instanceof TestCase);
    expect(tc).toBeDefined();
    expect(tc.name).toBe('testMultiply');

    const assertion = tc.body.find((n) => n instanceof Assertion);
    expect(assertion).toBeDefined();
    expect(assertion.kind).toBe('equal');
    expect(assertion.expected).toBe('6');
    expect(assertion.subject).toBe('multiply(2, 3)');
  });

  it('should parse @Disabled as skip modifier', () => {
    const source = `
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Disabled;

class SkipTest {
    @Disabled
    @Test
    void testSkipped() {
    }
}`;
    const ir = junit5Fw.parse(source);
    const suite = ir.body.find((n) => n instanceof TestSuite);
    const tc = suite.tests.find((n) => n instanceof TestCase);
    expect(tc.modifiers.some((m) => m.modifierType === 'skip')).toBe(true);
  });

  it('should extract assertNotNull subject', () => {
    const source = `
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

class NullTest {
    @Test
    void testNotNull() {
        Assertions.assertNotNull(getValue());
    }
}`;
    const ir = junit5Fw.parse(source);
    const suite = ir.body.find((n) => n instanceof TestSuite);
    const tc = suite.tests.find((n) => n instanceof TestCase);
    const assertion = tc.body.find((n) => n instanceof Assertion);
    expect(assertion.kind).toBe('isDefined');
    expect(assertion.subject).toBe('getValue()');
  });
});

// ═══════════════════════════════════════════════════════════════════════
// TestNG Parser Tests
// ═══════════════════════════════════════════════════════════════════════

describe('TestNG parser — nested IR', () => {
  it('should produce nested TestSuite > TestCase > Assertion', () => {
    const source = `
import org.testng.annotations.Test;
import org.testng.Assert;

public class SampleTest {
    @Test
    public void testValue() {
        Assert.assertEquals(getValue(), 42);
    }
}`;
    const ir = testngFw.parse(source);
    expect(ir).toBeInstanceOf(TestFile);

    const suite = ir.body.find((n) => n instanceof TestSuite);
    expect(suite).toBeDefined();
    expect(suite.name).toBe('SampleTest');

    const tc = suite.tests.find((n) => n instanceof TestCase);
    expect(tc).toBeDefined();

    const assertion = tc.body.find((n) => n instanceof Assertion);
    expect(assertion).toBeDefined();
    expect(assertion.kind).toBe('equal');
    // TestNG: assertEquals(actual, expected)
    expect(assertion.subject).toBe('getValue()');
    expect(assertion.expected).toBe('42');
  });

  it('should parse @BeforeMethod and @AfterMethod as hooks', () => {
    const source = `
import org.testng.annotations.Test;
import org.testng.annotations.BeforeMethod;
import org.testng.annotations.AfterMethod;

public class HookTest {
    @BeforeMethod
    public void before() {
        setup();
    }

    @AfterMethod
    public void after() {
        cleanup();
    }

    @Test
    public void testIt() {
    }
}`;
    const ir = testngFw.parse(source);
    const suite = ir.body.find((n) => n instanceof TestSuite);
    expect(suite.hooks.length).toBe(2);
    expect(suite.hooks.some((h) => h.hookType === 'beforeEach')).toBe(true);
    expect(suite.hooks.some((h) => h.hookType === 'afterEach')).toBe(true);
  });
});

// ═══════════════════════════════════════════════════════════════════════
// pytest Parser Tests
// ═══════════════════════════════════════════════════════════════════════

describe('pytest parser — nested IR', () => {
  it('should produce TestFile with TestCase containing Assertion', () => {
    const source = `
import pytest

def test_addition():
    assert 2 + 2 == 4
    assert 3 + 3 != 7
`;
    const ir = pytestFw.parse(source);
    expect(ir).toBeInstanceOf(TestFile);

    const tc = ir.body.find((n) => n instanceof TestCase);
    expect(tc).toBeDefined();
    expect(tc.name).toBe('test_addition');
    expect(tc.body.length).toBe(2);

    const eq = tc.body[0];
    expect(eq).toBeInstanceOf(Assertion);
    expect(eq.kind).toBe('equal');
    expect(eq.subject).toBe('2 + 2');
    expect(eq.expected).toBe('4');

    const neq = tc.body[1];
    expect(neq.kind).toBe('notEqual');
  });

  it('should extract assert is None and assert is not None', () => {
    const source = `
def test_none():
    assert result is None
    assert other is not None
`;
    const ir = pytestFw.parse(source);
    const tc = ir.body.find((n) => n instanceof TestCase);
    expect(tc.body[0].kind).toBe('isNull');
    expect(tc.body[1].kind).toBe('isDefined');
  });

  it('should extract assert in and assert not in', () => {
    const source = `
def test_contains():
    assert "foo" in my_list
    assert "bar" not in my_list
`;
    const ir = pytestFw.parse(source);
    const tc = ir.body.find((n) => n instanceof TestCase);
    expect(tc.body[0].kind).toBe('contains');
    expect(tc.body[0].subject).toBe('my_list');
    expect(tc.body[0].expected).toBe('"foo"');
    expect(tc.body[1].kind).toBe('notContains');
  });

  it('should parse @pytest.mark.skip as modifier', () => {
    const source = `
import pytest

@pytest.mark.skip
def test_skipped():
    assert True
`;
    const ir = pytestFw.parse(source);
    const tc = ir.body.find((n) => n instanceof TestCase);
    expect(tc.modifiers.some((m) => m.modifierType === 'skip')).toBe(true);
  });

  it('should parse class-based pytest tests as TestSuite', () => {
    const source = `
class TestMath:
    def test_add(self):
        assert 1 + 1 == 2
`;
    const ir = pytestFw.parse(source);
    const suite = ir.body.find((n) => n instanceof TestSuite);
    expect(suite).toBeDefined();
    expect(suite.name).toBe('TestMath');
  });
});

// ═══════════════════════════════════════════════════════════════════════
// unittest Parser Tests
// ═══════════════════════════════════════════════════════════════════════

describe('unittest parser — nested IR', () => {
  it('should produce TestFile with TestSuite > TestCase > Assertion', () => {
    const source = `
import unittest

class TestCalc(unittest.TestCase):
    def test_add(self):
        self.assertEqual(add(2, 2), 4)

    def test_sub(self):
        self.assertNotEqual(sub(5, 3), 0)
`;
    const ir = unittestFw.parse(source);
    expect(ir).toBeInstanceOf(TestFile);

    const suite = ir.body.find((n) => n instanceof TestSuite);
    expect(suite).toBeDefined();
    expect(suite.name).toBe('TestCalc');
    expect(suite.tests.length).toBe(2);

    const tc = suite.tests[0];
    expect(tc).toBeInstanceOf(TestCase);
    expect(tc.name).toBe('test_add');

    const assertion = tc.body.find((n) => n instanceof Assertion);
    expect(assertion).toBeDefined();
    expect(assertion.kind).toBe('equal');
    expect(assertion.subject).toBe('add(2, 2)');
    expect(assertion.expected).toBe('4');
  });

  it('should parse setUp/tearDown as hooks', () => {
    const source = `
import unittest

class TestHooks(unittest.TestCase):
    def setUp(self):
        self.x = 1

    def tearDown(self):
        self.x = None

    def test_it(self):
        self.assertTrue(self.x)
`;
    const ir = unittestFw.parse(source);
    const suite = ir.body.find((n) => n instanceof TestSuite);
    expect(suite.hooks.length).toBe(2);
    expect(suite.hooks.some((h) => h.hookType === 'beforeEach')).toBe(true);
    expect(suite.hooks.some((h) => h.hookType === 'afterEach')).toBe(true);
  });

  it('should extract assertIn subject/expected', () => {
    const source = `
import unittest

class TestContains(unittest.TestCase):
    def test_in(self):
        self.assertIn("foo", my_list)
`;
    const ir = unittestFw.parse(source);
    const suite = ir.body.find((n) => n instanceof TestSuite);
    const tc = suite.tests[0];
    const assertion = tc.body.find((n) => n instanceof Assertion);
    expect(assertion.kind).toBe('contains');
  });
});

// ═══════════════════════════════════════════════════════════════════════
// nose2 Parser Tests
// ═══════════════════════════════════════════════════════════════════════

describe('nose2 parser — nested IR', () => {
  it('should produce nested TestSuite > TestCase > Assertion', () => {
    const source = `
from nose.tools import assert_equal

class TestCalc:
    def test_add(self):
        assert_equal(add(2, 2), 4)
`;
    const ir = nose2Fw.parse(source);
    expect(ir).toBeInstanceOf(TestFile);

    const suite = ir.body.find((n) => n instanceof TestSuite);
    expect(suite).toBeDefined();

    const tc = suite.tests.find((n) => n instanceof TestCase);
    expect(tc).toBeDefined();

    const assertion = tc.body.find((n) => n instanceof Assertion);
    expect(assertion).toBeDefined();
    expect(assertion.kind).toBe('equal');
    expect(assertion.subject).toBe('add(2, 2)');
    expect(assertion.expected).toBe('4');
  });

  it('should extract assert_is_none and assert_raises', () => {
    const source = `
from nose.tools import assert_is_none, assert_raises

class TestNose:
    def test_none(self):
        assert_is_none(get_value())

    def test_raises(self):
        assert_raises(ValueError)
`;
    const ir = nose2Fw.parse(source);
    const suite = ir.body.find((n) => n instanceof TestSuite);

    const tc1 = suite.tests[0];
    const a1 = tc1.body.find((n) => n instanceof Assertion);
    expect(a1.kind).toBe('isNull');
    expect(a1.subject).toBe('get_value()');

    const tc2 = suite.tests[1];
    const a2 = tc2.body.find((n) => n instanceof Assertion);
    expect(a2.kind).toBe('throws');
    expect(a2.expected).toBe('ValueError');
  });
});

// ═══════════════════════════════════════════════════════════════════════
// JUnit5 Emitter Tests
// ═══════════════════════════════════════════════════════════════════════

describe('JUnit5 ir-full emitter', () => {
  it('should export emitNode, matchesBaseline, emitFullFile', () => {
    expect(typeof junit5Emitter.emitNode).toBe('function');
    expect(typeof junit5Emitter.matchesBaseline).toBe('function');
    expect(typeof junit5Emitter.emitFullFile).toBe('function');
  });

  it('should emit assertEquals(expected, actual) — expected first', () => {
    const node = new Assertion({
      kind: 'equal',
      subject: 'result',
      expected: '42',
    });
    const result = junit5Emitter.emitAssertion(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe('Assertions.assertEquals(42, result)');
  });

  it('should emit assertNotNull', () => {
    const node = new Assertion({ kind: 'isDefined', subject: 'obj' });
    const result = junit5Emitter.emitAssertion(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe('Assertions.assertNotNull(obj)');
  });

  it('should emit a complete JUnit5 file from IR tree', () => {
    const ir = new TestFile({
      language: 'java',
      imports: [],
      body: [
        new TestSuite({
          name: 'MathTest',
          hooks: [
            new Hook({
              hookType: 'beforeEach',
              body: [new RawCode({ code: 'calc = new Calculator();' })],
            }),
          ],
          tests: [
            new TestCase({
              name: 'testAdd',
              body: [
                new Assertion({
                  kind: 'equal',
                  subject: 'calc.add(2, 3)',
                  expected: '5',
                }),
              ],
            }),
          ],
        }),
      ],
    });

    const output = junit5Emitter.emitFullFile(ir);
    expect(output).not.toBeNull();
    expect(output).toContain('import org.junit.jupiter.api.Test;');
    expect(output).toContain('import org.junit.jupiter.api.Assertions;');
    expect(output).toContain('import org.junit.jupiter.api.BeforeEach;');
    expect(output).toContain('class MathTest {');
    expect(output).toContain('@BeforeEach');
    expect(output).toContain('void beforeEach()');
    expect(output).toContain('@Test');
    expect(output).toContain('void testAdd()');
    expect(output).toContain('Assertions.assertEquals(5, calc.add(2, 3));');
  });

  it('should emit @Disabled for skipped tests', () => {
    const ir = new TestFile({
      language: 'java',
      imports: [],
      body: [
        new TestSuite({
          name: 'SkipTest',
          tests: [
            new TestCase({
              name: 'testSkipped',
              modifiers: [new Modifier({ modifierType: 'skip' })],
              body: [],
            }),
          ],
        }),
      ],
    });

    const output = junit5Emitter.emitFullFile(ir);
    expect(output).toContain('@Disabled');
    expect(output).toContain('import org.junit.jupiter.api.Disabled;');
  });

  it('should emit static void for beforeAll/afterAll hooks', () => {
    const ir = new TestFile({
      language: 'java',
      imports: [],
      body: [
        new TestSuite({
          name: 'StaticTest',
          hooks: [new Hook({ hookType: 'beforeAll', body: [] })],
          tests: [],
        }),
      ],
    });

    const output = junit5Emitter.emitFullFile(ir);
    expect(output).toContain('@BeforeAll');
    expect(output).toContain('static void beforeAll()');
  });

  it('should match assertion baseline lines', () => {
    const node = new Assertion({ kind: 'equal', subject: 'a', expected: 'b' });
    expect(
      junit5Emitter.matchesBaseline('Assertions.assertEquals(b, a);', node)
    ).toBe(true);
    expect(junit5Emitter.matchesBaseline('assertEquals(b, a);', node)).toBe(
      true
    );
    expect(junit5Emitter.matchesBaseline('some.other.call();', node)).toBe(
      false
    );
  });
});

// ═══════════════════════════════════════════════════════════════════════
// TestNG Emitter Tests
// ═══════════════════════════════════════════════════════════════════════

describe('TestNG ir-full emitter', () => {
  it('should export emitNode, matchesBaseline, emitFullFile', () => {
    expect(typeof testngEmitter.emitNode).toBe('function');
    expect(typeof testngEmitter.matchesBaseline).toBe('function');
    expect(typeof testngEmitter.emitFullFile).toBe('function');
  });

  it('should emit Assert.assertEquals(actual, expected) — actual first', () => {
    const node = new Assertion({
      kind: 'equal',
      subject: 'result',
      expected: '42',
    });
    const result = testngEmitter.emitAssertion(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe('Assert.assertEquals(result, 42)');
  });

  it('should emit assertNull', () => {
    const node = new Assertion({ kind: 'isNull', subject: 'val' });
    const result = testngEmitter.emitAssertion(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe('Assert.assertNull(val)');
  });

  it('should emit a complete TestNG file from IR tree', () => {
    const ir = new TestFile({
      language: 'java',
      imports: [],
      body: [
        new TestSuite({
          name: 'SampleTest',
          hooks: [],
          tests: [
            new TestCase({
              name: 'testValue',
              body: [
                new Assertion({
                  kind: 'equal',
                  subject: 'getValue()',
                  expected: '10',
                }),
              ],
            }),
          ],
        }),
      ],
    });

    const output = testngEmitter.emitFullFile(ir);
    expect(output).not.toBeNull();
    expect(output).toContain('import org.testng.annotations.Test;');
    expect(output).toContain('import org.testng.Assert;');
    expect(output).toContain('public class SampleTest {');
    expect(output).toContain('@Test');
    expect(output).toContain('public void testValue()');
    expect(output).toContain('Assert.assertEquals(getValue(), 10);');
  });

  it('should emit @Test(enabled = false) for skipped tests', () => {
    const ir = new TestFile({
      language: 'java',
      imports: [],
      body: [
        new TestSuite({
          name: 'SkipTest',
          tests: [
            new TestCase({
              name: 'testSkipped',
              modifiers: [new Modifier({ modifierType: 'skip' })],
              body: [],
            }),
          ],
        }),
      ],
    });

    const output = testngEmitter.emitFullFile(ir);
    expect(output).toContain('@Test(enabled = false)');
  });

  it('should emit @BeforeMethod/@AfterMethod for hooks', () => {
    const ir = new TestFile({
      language: 'java',
      imports: [],
      body: [
        new TestSuite({
          name: 'HookTest',
          hooks: [
            new Hook({ hookType: 'beforeEach', body: [] }),
            new Hook({ hookType: 'afterEach', body: [] }),
          ],
          tests: [],
        }),
      ],
    });

    const output = testngEmitter.emitFullFile(ir);
    expect(output).toContain('@BeforeMethod');
    expect(output).toContain('@AfterMethod');
    expect(output).toContain('import org.testng.annotations.BeforeMethod;');
    expect(output).toContain('import org.testng.annotations.AfterMethod;');
  });
});

// ═══════════════════════════════════════════════════════════════════════
// pytest Emitter Tests
// ═══════════════════════════════════════════════════════════════════════

describe('pytest ir-full emitter', () => {
  it('should export emitNode, matchesBaseline, emitFullFile', () => {
    expect(typeof pytestEmitter.emitNode).toBe('function');
    expect(typeof pytestEmitter.matchesBaseline).toBe('function');
    expect(typeof pytestEmitter.emitFullFile).toBe('function');
  });

  it('should emit assert actual == expected (bare assert)', () => {
    const node = new Assertion({
      kind: 'equal',
      subject: 'result',
      expected: '42',
    });
    const result = pytestEmitter.emitAssertion(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe('assert result == 42');
  });

  it('should emit assert x is None', () => {
    const node = new Assertion({ kind: 'isNull', subject: 'val' });
    const result = pytestEmitter.emitAssertion(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe('assert val is None');
  });

  it('should emit assert not x for falsy', () => {
    const node = new Assertion({ kind: 'falsy', subject: 'val' });
    const result = pytestEmitter.emitAssertion(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe('assert not val');
  });

  it('should emit with pytest.raises(E):', () => {
    const node = new Assertion({ kind: 'throws', expected: 'ValueError' });
    const result = pytestEmitter.emitAssertion(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe('with pytest.raises(ValueError):');
  });

  it('should emit a complete pytest file from IR tree', () => {
    const ir = new TestFile({
      language: 'python',
      imports: [],
      body: [
        new TestCase({
          name: 'test_addition',
          body: [
            new Assertion({
              kind: 'equal',
              subject: '2 + 2',
              expected: '4',
            }),
          ],
        }),
        new TestCase({
          name: 'test_none',
          body: [
            new Assertion({
              kind: 'isNull',
              subject: 'get_value()',
            }),
          ],
        }),
      ],
    });

    const output = pytestEmitter.emitFullFile(ir);
    expect(output).not.toBeNull();
    expect(output).toContain('def test_addition():');
    expect(output).toContain('assert 2 + 2 == 4');
    expect(output).toContain('def test_none():');
    expect(output).toContain('assert get_value() is None');
  });

  it('should emit @pytest.mark.skip for skipped tests', () => {
    const ir = new TestFile({
      language: 'python',
      imports: [],
      body: [
        new TestCase({
          name: 'test_skipped',
          modifiers: [new Modifier({ modifierType: 'skip' })],
          body: [],
        }),
      ],
    });

    const output = pytestEmitter.emitFullFile(ir);
    expect(output).toContain('import pytest');
    expect(output).toContain('@pytest.mark.skip');
  });

  it('should emit class-based test with fixtures', () => {
    const ir = new TestFile({
      language: 'python',
      imports: [],
      body: [
        new TestSuite({
          name: 'TestMath',
          hooks: [
            new Hook({
              hookType: 'beforeEach',
              body: [new RawCode({ code: 'self.calc = Calculator()' })],
            }),
          ],
          tests: [
            new TestCase({
              name: 'test_add',
              body: [
                new Assertion({
                  kind: 'equal',
                  subject: 'self.calc.add(2, 3)',
                  expected: '5',
                }),
              ],
            }),
          ],
        }),
      ],
    });

    const output = pytestEmitter.emitFullFile(ir);
    expect(output).toContain('import pytest');
    expect(output).toContain('class TestMath:');
    expect(output).toContain('@pytest.fixture(autouse=True)');
    expect(output).toContain('def setup_method(self):');
    expect(output).toContain('def test_add():');
    expect(output).toContain('assert self.calc.add(2, 3) == 5');
  });

  it('should match assertion baseline lines', () => {
    const node = new Assertion({ kind: 'equal', subject: 'a', expected: 'b' });
    expect(pytestEmitter.matchesBaseline('    assert a == b', node)).toBe(true);
    expect(
      pytestEmitter.matchesBaseline('    with pytest.raises(E):', node)
    ).toBe(true);
    expect(pytestEmitter.matchesBaseline('    some_func()', node)).toBe(false);
  });
});

// ═══════════════════════════════════════════════════════════════════════
// unittest Emitter Tests
// ═══════════════════════════════════════════════════════════════════════

describe('unittest ir-full emitter', () => {
  it('should export emitNode, matchesBaseline, emitFullFile', () => {
    expect(typeof unittestEmitter.emitNode).toBe('function');
    expect(typeof unittestEmitter.matchesBaseline).toBe('function');
    expect(typeof unittestEmitter.emitFullFile).toBe('function');
  });

  it('should emit self.assertEqual(actual, expected)', () => {
    const node = new Assertion({
      kind: 'equal',
      subject: 'result',
      expected: '42',
    });
    const result = unittestEmitter.emitAssertion(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe('self.assertEqual(result, 42)');
  });

  it('should emit self.assertIsNone', () => {
    const node = new Assertion({ kind: 'isNull', subject: 'val' });
    const result = unittestEmitter.emitAssertion(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe('self.assertIsNone(val)');
  });

  it('should emit self.assertIn(expected, subject) for contains', () => {
    const node = new Assertion({
      kind: 'contains',
      subject: 'my_list',
      expected: '"foo"',
    });
    const result = unittestEmitter.emitAssertion(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe('self.assertIn("foo", my_list)');
  });

  it('should emit with self.assertRaises(E):', () => {
    const node = new Assertion({ kind: 'throws', expected: 'ValueError' });
    const result = unittestEmitter.emitAssertion(node);
    expect(result.supported).toBe(true);
    expect(result.code).toBe('with self.assertRaises(ValueError):');
  });

  it('should emit a complete unittest file from IR tree', () => {
    const ir = new TestFile({
      language: 'python',
      imports: [],
      body: [
        new TestSuite({
          name: 'TestCalc',
          hooks: [
            new Hook({
              hookType: 'beforeEach',
              body: [new RawCode({ code: 'self.x = 1' })],
            }),
          ],
          tests: [
            new TestCase({
              name: 'test_add',
              body: [
                new Assertion({
                  kind: 'equal',
                  subject: 'self.x + 1',
                  expected: '2',
                }),
              ],
            }),
          ],
        }),
      ],
    });

    const output = unittestEmitter.emitFullFile(ir);
    expect(output).not.toBeNull();
    expect(output).toContain('import unittest');
    expect(output).toContain('class TestCalc(unittest.TestCase):');
    expect(output).toContain('def setUp(self):');
    expect(output).toContain('self.x = 1');
    expect(output).toContain('def test_add(self):');
    expect(output).toContain('self.assertEqual(self.x + 1, 2)');
    expect(output).toContain("if __name__ == '__main__':");
    expect(output).toContain('unittest.main()');
  });

  it('should emit @unittest.skip for skipped tests', () => {
    const ir = new TestFile({
      language: 'python',
      imports: [],
      body: [
        new TestSuite({
          name: 'TestSkip',
          tests: [
            new TestCase({
              name: 'test_skipped',
              modifiers: [new Modifier({ modifierType: 'skip' })],
              body: [],
            }),
          ],
        }),
      ],
    });

    const output = unittestEmitter.emitFullFile(ir);
    expect(output).toContain('@unittest.skip');
  });

  it('should emit @classmethod for setUpClass/tearDownClass', () => {
    const ir = new TestFile({
      language: 'python',
      imports: [],
      body: [
        new TestSuite({
          name: 'TestClassLevel',
          hooks: [
            new Hook({ hookType: 'beforeAll', body: [] }),
            new Hook({ hookType: 'afterAll', body: [] }),
          ],
          tests: [],
        }),
      ],
    });

    const output = unittestEmitter.emitFullFile(ir);
    expect(output).toContain('@classmethod');
    expect(output).toContain('def setUpClass(cls):');
    expect(output).toContain('def tearDownClass(cls):');
  });

  it('should match assertion baseline lines', () => {
    const node = new Assertion({ kind: 'equal', subject: 'a', expected: 'b' });
    expect(
      unittestEmitter.matchesBaseline('self.assertEqual(a, b)', node)
    ).toBe(true);
    expect(unittestEmitter.matchesBaseline('self.assertTrue(x)', node)).toBe(
      true
    );
    expect(unittestEmitter.matchesBaseline('some_func()', node)).toBe(false);
  });
});

// ═══════════════════════════════════════════════════════════════════════
// Dynamic Emitter Loading
// ═══════════════════════════════════════════════════════════════════════

describe('Dynamic emitter loading for Phase 5 targets', () => {
  it('should dynamically import JUnit5 emitter', async () => {
    const emitter = await import('../../src/core/emitters/junit5/irEmitter.js');
    expect(typeof emitter.emitFullFile).toBe('function');
    expect(typeof emitter.emitNode).toBe('function');
    expect(typeof emitter.matchesBaseline).toBe('function');
  });

  it('should dynamically import TestNG emitter', async () => {
    const emitter = await import('../../src/core/emitters/testng/irEmitter.js');
    expect(typeof emitter.emitFullFile).toBe('function');
    expect(typeof emitter.emitNode).toBe('function');
    expect(typeof emitter.matchesBaseline).toBe('function');
  });

  it('should dynamically import pytest emitter', async () => {
    const emitter = await import('../../src/core/emitters/pytest/irEmitter.js');
    expect(typeof emitter.emitFullFile).toBe('function');
    expect(typeof emitter.emitNode).toBe('function');
    expect(typeof emitter.matchesBaseline).toBe('function');
  });

  it('should dynamically import unittest emitter', async () => {
    const emitter = await import(
      '../../src/core/emitters/unittest/irEmitter.js'
    );
    expect(typeof emitter.emitFullFile).toBe('function');
    expect(typeof emitter.emitNode).toBe('function');
    expect(typeof emitter.matchesBaseline).toBe('function');
  });
});

// ═══════════════════════════════════════════════════════════════════════
// Round-trip: parse → emit produces valid output
// ═══════════════════════════════════════════════════════════════════════

describe('Round-trip: parse source → ir-full emit', () => {
  it('JUnit4 source → JUnit5 ir-full', () => {
    const source = `
import org.junit.Test;
import static org.junit.Assert.*;

public class RoundTripTest {
    @Test
    public void testValue() {
        assertEquals(10, getValue());
    }
}`;
    const ir = junit4Fw.parse(source);
    const output = junit5Emitter.emitFullFile(ir);
    expect(output).toContain('import org.junit.jupiter.api.Test;');
    expect(output).toContain('Assertions.assertEquals(10, getValue())');
  });

  it('JUnit5 source → TestNG ir-full', () => {
    const source = `
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.Assertions;

class J5Test {
    @Test
    void testCheck() {
        Assertions.assertTrue(isReady());
    }
}`;
    const ir = junit5Fw.parse(source);
    const output = testngEmitter.emitFullFile(ir);
    expect(output).toContain('import org.testng.annotations.Test;');
    expect(output).toContain('Assert.assertTrue(isReady())');
  });

  it('unittest source → pytest ir-full', () => {
    const source = `
import unittest

class TestExample(unittest.TestCase):
    def setUp(self):
        self.val = 42

    def test_equal(self):
        self.assertEqual(self.val, 42)
`;
    const ir = unittestFw.parse(source);
    const output = pytestEmitter.emitFullFile(ir);
    expect(output).toContain('assert');
    expect(output).toContain('== 42');
  });

  it('pytest source → unittest ir-full', () => {
    const source = `
def test_simple():
    assert 1 + 1 == 2
    assert result is not None
`;
    const ir = pytestFw.parse(source);
    const output = unittestEmitter.emitFullFile(ir);
    expect(output).toContain('import unittest');
    expect(output).toContain('self.assertEqual(1 + 1, 2)');
    expect(output).toContain('self.assertIsNotNone(result)');
  });

  it('nose2 source → pytest ir-full', () => {
    const source = `
from nose.tools import assert_equal, assert_true

def test_nose():
    assert_equal(compute(), 42)
    assert_true(is_valid())
`;
    const ir = nose2Fw.parse(source);
    const output = pytestEmitter.emitFullFile(ir);
    expect(output).toContain('assert compute() == 42');
    expect(output).toContain('assert is_valid()');
  });
});
