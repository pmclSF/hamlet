import { InputNormalizer } from '../../src/core/InputNormalizer.js';

describe('InputNormalizer', () => {
  let normalizer;

  beforeEach(() => {
    normalizer = new InputNormalizer();
  });

  describe('normalize', () => {
    it('should pass through clean content unchanged', () => {
      const content = `describe('test', () => {\n  it('works', () => {\n    expect(true).toBe(true);\n  });\n});`;
      const { normalized, issues } = normalizer.normalize(content);

      expect(normalized).toBe(content);
      expect(issues).toHaveLength(0);
    });

    it('should detect mismatched single quotes and report issue', () => {
      const content = `const x = 'hello;`;
      const { normalized, issues } = normalizer.normalize(content);

      expect(issues.some(i => i.type === 'quote')).toBe(true);
      // Should attempt to close the quote
      expect(normalized).toContain("'");
    });

    it('should detect mismatched double quotes and report issue', () => {
      const content = `const x = "hello;`;
      const { normalized, issues } = normalizer.normalize(content);

      expect(issues.some(i => i.type === 'quote')).toBe(true);
    });

    it('should report unclosed brackets', () => {
      const content = `function foo() {\n  if (true) {\n    return 1;\n  }\n`;
      const { issues } = normalizer.normalize(content);

      expect(issues.some(i => i.type === 'bracket')).toBe(true);
    });

    it('should report unclosed parentheses', () => {
      const content = `describe('test', () => {\n  it('works', () => {\n    expect(true).toBe(true);\n  });\n`;
      const { issues } = normalizer.normalize(content);

      expect(issues.some(i => i.type === 'bracket')).toBe(true);
    });

    it('should handle empty files', () => {
      const { normalized, issues } = normalizer.normalize('');

      expect(normalized).toBe('');
      expect(issues.some(i => i.type === 'empty')).toBe(true);
    });

    it('should handle null content', () => {
      const { normalized, issues } = normalizer.normalize(null);

      expect(normalized).toBe('');
      expect(issues.some(i => i.type === 'empty')).toBe(true);
    });

    it('should detect binary content', () => {
      const binary = '\x00\x01\x02\x03\x04\x05\x06\x07\x08';
      const { normalized, issues } = normalizer.normalize(binary);

      expect(issues.some(i => i.type === 'binary')).toBe(true);
      expect(normalized).toBe('');
    });

    it('should remove BOM character', () => {
      const bom = '\uFEFF';
      const content = bom + 'const x = 1;';
      const { normalized, issues } = normalizer.normalize(content);

      expect(normalized).toBe('const x = 1;');
      expect(issues.some(i => i.type === 'encoding')).toBe(true);
    });

    it('should normalize CRLF to LF', () => {
      const content = 'line1\r\nline2\r\nline3';
      const { normalized, issues } = normalizer.normalize(content);

      expect(normalized).toBe('line1\nline2\nline3');
      expect(issues.some(i => i.type === 'encoding')).toBe(true);
    });

    it('should handle content with only comments', () => {
      const content = '// This is a comment\n/* Another comment */';
      const { normalized, issues } = normalizer.normalize(content);

      expect(normalized).toBe(content);
      expect(issues.filter(i => i.type !== 'encoding')).toHaveLength(0);
    });

    it('should report line numbers for issues', () => {
      const content = `const a = 'ok';\nconst b = "broken;`;
      const { issues } = normalizer.normalize(content);

      const quoteIssue = issues.find(i => i.type === 'quote');
      expect(quoteIssue).toBeDefined();
      expect(quoteIssue.line).toBe(2);
    });

    it('should handle balanced content without false positives', () => {
      const content = `const obj = { a: [1, 2], b: { c: '(test)' } };\nfoo(bar(baz()));`;
      const { issues } = normalizer.normalize(content);

      expect(issues.filter(i => i.type === 'bracket')).toHaveLength(0);
    });

    it('should handle deeply nested brackets', () => {
      const content = `((((({})))))`;
      const { issues } = normalizer.normalize(content);

      expect(issues.filter(i => i.type === 'bracket')).toHaveLength(0);
    });

    it('should not count quotes inside opposite quote type as mismatched', () => {
      const content = `const x = "it's fine";`;
      const { issues } = normalizer.normalize(content);

      expect(issues.filter(i => i.type === 'quote')).toHaveLength(0);
    });
  });
});
