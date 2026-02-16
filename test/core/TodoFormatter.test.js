import { TodoFormatter } from '../../src/core/TodoFormatter.js';

describe('TodoFormatter', () => {
  describe('JavaScript comment syntax', () => {
    let formatter;

    beforeEach(() => {
      formatter = new TodoFormatter('javascript');
    });

    it('should format a HAMLET-TODO with // prefix', () => {
      const result = formatter.formatTodo({
        id: 'UNCONVERTIBLE-001',
        description: 'Jest snapshot testing has no equivalent',
        original: 'expect(tree).toMatchSnapshot()',
        action: 'Replace with explicit assertion',
      });

      expect(result).toContain('// HAMLET-TODO [UNCONVERTIBLE-001]: Jest snapshot testing has no equivalent');
      expect(result).toContain('// Original: expect(tree).toMatchSnapshot()');
      expect(result).toContain('// Manual action required: Replace with explicit assertion');
    });

    it('should handle multi-line original source', () => {
      const result = formatter.formatTodo({
        id: 'UNCONVERTIBLE-002',
        description: 'Custom command',
        original: 'Cypress.Commands.add(\n  "login",\n  () => {}\n)',
        action: 'Convert to fixture',
      });

      expect(result).toContain('// Original:');
      expect(result).toContain('//   Cypress.Commands.add(');
      expect(result).toContain('//     "login",');
    });

    it('should format a HAMLET-WARNING', () => {
      const result = formatter.formatWarning({
        description: 'vi.mock hoisting differs from jest.mock',
        original: 'jest.mock("./render")',
      });

      expect(result).toContain('// HAMLET-WARNING: vi.mock hoisting differs from jest.mock');
      expect(result).toContain('// Original: jest.mock("./render")');
    });

    it('should handle warning without original source', () => {
      const result = formatter.formatWarning({
        description: 'Snapshot file locations differ between Jest and Vitest',
      });

      expect(result).toContain('// HAMLET-WARNING:');
      expect(result).not.toContain('// Original:');
    });
  });

  describe('Python comment syntax', () => {
    let formatter;

    beforeEach(() => {
      formatter = new TodoFormatter('python');
    });

    it('should use # prefix for Python', () => {
      const result = formatter.formatTodo({
        id: 'UNCONVERTIBLE-003',
        description: 'pytest conftest.py hierarchy has no unittest equivalent',
        original: '@pytest.fixture(scope="session")',
        action: 'Move fixture logic to setUpClass or a base test class',
      });

      expect(result).toContain('# HAMLET-TODO [UNCONVERTIBLE-003]:');
      expect(result).toContain('# Original: @pytest.fixture(scope="session")');
      expect(result).toContain('# Manual action required:');
    });
  });

  describe('Ruby comment syntax', () => {
    it('should use # prefix for Ruby', () => {
      const formatter = new TodoFormatter('ruby');
      const result = formatter.formatTodo({
        id: 'UNCONVERTIBLE-004',
        description: 'RSpec shared_examples cannot be converted',
        original: 'it_behaves_like "a searchable resource"',
        action: 'Extract shared tests into a module',
      });

      expect(result).toContain('# HAMLET-TODO [UNCONVERTIBLE-004]:');
      expect(result).toContain('# Original:');
    });
  });

  describe('Java comment syntax', () => {
    it('should use // prefix for Java', () => {
      const formatter = new TodoFormatter('java');
      const result = formatter.formatTodo({
        id: 'UNCONVERTIBLE-006',
        description: 'JUnit 4 @Rule requires manual conversion',
        original: '@Rule public ExpectedException thrown = ExpectedException.none();',
        action: 'Replace with assertThrows() calls',
      });

      expect(result).toContain('// HAMLET-TODO [UNCONVERTIBLE-006]:');
      expect(result).toContain('// Original: @Rule public ExpectedException');
    });
  });

  describe('unknown language', () => {
    it('should fall back to JavaScript comment syntax', () => {
      const formatter = new TodoFormatter('unknown');
      const result = formatter.formatTodo({
        id: 'TEST-001',
        description: 'Test',
        original: 'code',
        action: 'Fix it',
      });

      expect(result).toContain('// HAMLET-TODO');
    });
  });
});
