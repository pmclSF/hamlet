import {
  IRNode,
  TestFile,
  TestSuite,
  TestCase,
  Hook,
  Assertion,
  MockCall,
  ImportStatement,
  RawCode,
  Comment,
  SharedVariable,
  Modifier,
  ParameterSet,
  walkIR,
} from '../../src/core/ir.js';

describe('IR Node Types', () => {
  describe('IRNode (base)', () => {
    it('should set type and default properties', () => {
      const node = new IRNode('TestNode');
      expect(node.type).toBe('TestNode');
      expect(node.sourceLocation).toBeNull();
      expect(node.confidence).toBe('converted');
      expect(node.originalSource).toBe('');
    });

    it('should accept sourceLocation', () => {
      const loc = { line: 5, column: 0, endLine: 5, endColumn: 30 };
      const node = new IRNode('TestNode', { sourceLocation: loc });
      expect(node.sourceLocation).toEqual(loc);
    });

    it('should accept confidence flag', () => {
      const node = new IRNode('TestNode', { confidence: 'unconvertible' });
      expect(node.confidence).toBe('unconvertible');
    });

    it('should accept warning confidence', () => {
      const node = new IRNode('TestNode', { confidence: 'warning' });
      expect(node.confidence).toBe('warning');
    });

    it('should accept originalSource', () => {
      const node = new IRNode('TestNode', { originalSource: 'jest.fn()' });
      expect(node.originalSource).toBe('jest.fn()');
    });
  });

  describe('TestFile', () => {
    it('should have correct type and defaults', () => {
      const file = new TestFile();
      expect(file.type).toBe('TestFile');
      expect(file.language).toBe('javascript');
      expect(file.imports).toEqual([]);
      expect(file.body).toEqual([]);
    });

    it('should accept language, imports, and body', () => {
      const imp = new ImportStatement({ source: 'vitest' });
      const suite = new TestSuite({ name: 'MyTests' });
      const file = new TestFile({
        language: 'python',
        imports: [imp],
        body: [suite],
      });
      expect(file.language).toBe('python');
      expect(file.imports).toHaveLength(1);
      expect(file.body).toHaveLength(1);
    });
  });

  describe('TestSuite', () => {
    it('should have correct type and defaults', () => {
      const suite = new TestSuite();
      expect(suite.type).toBe('TestSuite');
      expect(suite.name).toBe('');
      expect(suite.hooks).toEqual([]);
      expect(suite.tests).toEqual([]);
      expect(suite.sharedState).toEqual([]);
      expect(suite.modifiers).toEqual([]);
    });

    it('should accept all properties', () => {
      const hook = new Hook({ hookType: 'beforeEach' });
      const test = new TestCase({ name: 'my test' });
      const modifier = new Modifier({ modifierType: 'skip' });
      const suite = new TestSuite({
        name: 'UserService',
        hooks: [hook],
        tests: [test],
        modifiers: [modifier],
      });
      expect(suite.name).toBe('UserService');
      expect(suite.hooks).toHaveLength(1);
      expect(suite.tests).toHaveLength(1);
      expect(suite.modifiers).toHaveLength(1);
    });
  });

  describe('TestCase', () => {
    it('should have correct type and defaults', () => {
      const tc = new TestCase();
      expect(tc.type).toBe('TestCase');
      expect(tc.name).toBe('');
      expect(tc.body).toEqual([]);
      expect(tc.modifiers).toEqual([]);
      expect(tc.parameters).toBeNull();
      expect(tc.isAsync).toBe(false);
    });

    it('should accept async and parameters', () => {
      const params = new ParameterSet({ paramKind: 'values' });
      const tc = new TestCase({
        name: 'should work',
        isAsync: true,
        parameters: params,
      });
      expect(tc.isAsync).toBe(true);
      expect(tc.parameters).toBe(params);
    });
  });

  describe('Hook', () => {
    it('should have correct type and defaults', () => {
      const hook = new Hook();
      expect(hook.type).toBe('Hook');
      expect(hook.hookType).toBe('beforeEach');
      expect(hook.scope).toBe('suite');
      expect(hook.body).toEqual([]);
      expect(hook.isAsync).toBe(false);
    });

    it('should accept all hook types', () => {
      for (const hookType of ['beforeAll', 'afterAll', 'beforeEach', 'afterEach', 'around']) {
        const hook = new Hook({ hookType });
        expect(hook.hookType).toBe(hookType);
      }
    });
  });

  describe('Assertion', () => {
    it('should have correct type and defaults', () => {
      const a = new Assertion();
      expect(a.type).toBe('Assertion');
      expect(a.kind).toBe('equal');
      expect(a.subject).toBe('');
      expect(a.expected).toBeNull();
      expect(a.isNegated).toBe(false);
      expect(a.message).toBeNull();
    });

    it('should accept all properties', () => {
      const a = new Assertion({
        kind: 'deepEqual',
        subject: 'result',
        expected: '{ a: 1 }',
        isNegated: true,
        message: 'should match',
      });
      expect(a.kind).toBe('deepEqual');
      expect(a.subject).toBe('result');
      expect(a.expected).toBe('{ a: 1 }');
      expect(a.isNegated).toBe(true);
      expect(a.message).toBe('should match');
    });
  });

  describe('MockCall', () => {
    it('should have correct type and defaults', () => {
      const m = new MockCall();
      expect(m.type).toBe('MockCall');
      expect(m.kind).toBe('createMock');
      expect(m.target).toBe('');
      expect(m.args).toEqual([]);
      expect(m.returnValue).toBeNull();
    });

    it('should accept all properties', () => {
      const m = new MockCall({
        kind: 'stubReturn',
        target: 'fetchUser',
        args: ['userId'],
        returnValue: '{ name: "Bob" }',
      });
      expect(m.kind).toBe('stubReturn');
      expect(m.target).toBe('fetchUser');
      expect(m.args).toEqual(['userId']);
      expect(m.returnValue).toBe('{ name: "Bob" }');
    });
  });

  describe('ImportStatement', () => {
    it('should have correct type and defaults', () => {
      const i = new ImportStatement();
      expect(i.type).toBe('ImportStatement');
      expect(i.kind).toBe('framework');
      expect(i.source).toBe('');
      expect(i.specifiers).toEqual([]);
      expect(i.isDefault).toBe(false);
      expect(i.isTypeOnly).toBe(false);
    });

    it('should accept specifiers and type-only imports', () => {
      const i = new ImportStatement({
        kind: 'library',
        source: 'vitest',
        specifiers: [
          { name: 'describe', alias: null },
          { name: 'it', alias: 'test' },
        ],
        isTypeOnly: true,
      });
      expect(i.specifiers).toHaveLength(2);
      expect(i.specifiers[1].alias).toBe('test');
      expect(i.isTypeOnly).toBe(true);
    });
  });

  describe('RawCode', () => {
    it('should have correct type and defaults', () => {
      const r = new RawCode();
      expect(r.type).toBe('RawCode');
      expect(r.code).toBe('');
      expect(r.comment).toBeNull();
    });

    it('should preserve code and optional comment', () => {
      const r = new RawCode({
        code: 'const x = 42;',
        comment: 'HAMLET-TODO: Review this',
      });
      expect(r.code).toBe('const x = 42;');
      expect(r.comment).toBe('HAMLET-TODO: Review this');
    });
  });

  describe('Comment', () => {
    it('should have correct type and defaults', () => {
      const c = new Comment();
      expect(c.type).toBe('Comment');
      expect(c.text).toBe('');
      expect(c.commentKind).toBe('inline');
      expect(c.preserveExact).toBe(false);
    });

    it('should support license headers with preserveExact', () => {
      const c = new Comment({
        text: '/* MIT License */',
        commentKind: 'license',
        preserveExact: true,
      });
      expect(c.preserveExact).toBe(true);
      expect(c.commentKind).toBe('license');
    });
  });

  describe('SharedVariable', () => {
    it('should have correct type and defaults', () => {
      const s = new SharedVariable();
      expect(s.type).toBe('SharedVariable');
      expect(s.name).toBe('');
      expect(s.initializer).toBe('');
      expect(s.isLazy).toBe(false);
      expect(s.scope).toBe('instance');
    });
  });

  describe('Modifier', () => {
    it('should have correct type and defaults', () => {
      const m = new Modifier();
      expect(m.type).toBe('Modifier');
      expect(m.modifierType).toBe('skip');
      expect(m.value).toBeNull();
      expect(m.condition).toBeNull();
    });

    it('should accept conditional skip with reason', () => {
      const m = new Modifier({
        modifierType: 'skip',
        value: 'not supported on CI',
        condition: 'process.env.CI',
      });
      expect(m.value).toBe('not supported on CI');
      expect(m.condition).toBe('process.env.CI');
    });
  });

  describe('ParameterSet', () => {
    it('should have correct type and defaults', () => {
      const p = new ParameterSet();
      expect(p.type).toBe('ParameterSet');
      expect(p.paramKind).toBe('values');
      expect(p.parameters).toEqual([]);
      expect(p.ids).toBeNull();
    });

    it('should accept parameters with ids', () => {
      const p = new ParameterSet({
        paramKind: 'csv',
        parameters: [
          { name: 'input', values: ['1', '2', '3'] },
          { name: 'expected', values: ['1', '4', '9'] },
        ],
        ids: ['one', 'two', 'three'],
      });
      expect(p.parameters).toHaveLength(2);
      expect(p.ids).toEqual(['one', 'two', 'three']);
    });
  });
});

describe('walkIR', () => {
  it('should visit all nodes in a tree depth-first', () => {
    const assertion = new Assertion({ kind: 'equal', sourceLocation: { line: 5 } });
    const testCase = new TestCase({ name: 'my test', body: [assertion] });
    const hook = new Hook({ hookType: 'beforeEach' });
    const suite = new TestSuite({ name: 'MySuite', hooks: [hook], tests: [testCase] });
    const imp = new ImportStatement({ source: 'vitest' });
    const file = new TestFile({ imports: [imp], body: [suite] });

    const visited = walkIR(file);
    const types = visited.map(n => n.type);

    expect(types).toEqual([
      'TestFile',
      'ImportStatement',
      'TestSuite',
      'Hook',
      'TestCase',
      'Assertion',
    ]);
  });

  it('should call visitor function for each node', () => {
    const testCase = new TestCase({ name: 'test' });
    const suite = new TestSuite({ name: 'suite', tests: [testCase] });
    const file = new TestFile({ body: [suite] });

    const visitedNames = [];
    walkIR(file, (node) => {
      visitedNames.push(node.type);
    });

    expect(visitedNames).toEqual(['TestFile', 'TestSuite', 'TestCase']);
  });

  it('should skip children when visitor returns false', () => {
    const assertion = new Assertion({ kind: 'equal' });
    const testCase = new TestCase({ name: 'test', body: [assertion] });
    const suite = new TestSuite({ name: 'suite', tests: [testCase] });
    const file = new TestFile({ body: [suite] });

    const visited = walkIR(file, (node) => {
      if (node.type === 'TestCase') return false;
    });

    const types = visited.map(n => n.type);
    expect(types).toContain('TestCase');
    expect(types).not.toContain('Assertion');
  });

  it('should handle empty tree', () => {
    const file = new TestFile();
    const visited = walkIR(file);
    expect(visited).toHaveLength(1);
    expect(visited[0].type).toBe('TestFile');
  });

  it('should visit nested suites', () => {
    const innerTest = new TestCase({ name: 'inner test' });
    const innerSuite = new TestSuite({ name: 'inner', tests: [innerTest] });
    const outerSuite = new TestSuite({ name: 'outer', tests: [innerSuite] });
    const file = new TestFile({ body: [outerSuite] });

    const visited = walkIR(file);
    const types = visited.map(n => n.type);
    expect(types).toEqual(['TestFile', 'TestSuite', 'TestSuite', 'TestCase']);
  });

  it('should visit modifiers on suites and test cases', () => {
    const mod = new Modifier({ modifierType: 'skip' });
    const tc = new TestCase({ name: 'test', modifiers: [mod] });
    const file = new TestFile({ body: [tc] });

    const visited = walkIR(file);
    const types = visited.map(n => n.type);
    expect(types).toContain('Modifier');
  });
});
