import { PatternEngine } from '../../src/core/PatternEngine.js';

describe('PatternEngine', () => {
  let engine;

  beforeEach(() => {
    engine = new PatternEngine();
  });

  describe('registerPattern', () => {
    it('should register a single pattern', () => {
      engine.registerPattern('test', 'cy\\.visit', 'page.goto');
      const patterns = engine.getPatternsForCategory('test');
      expect(patterns).toBeDefined();
      expect(patterns.length).toBe(1);
    });
  });

  describe('registerPatterns', () => {
    it('should register multiple patterns at once', () => {
      engine.registerPatterns('navigation', {
        'cy\\.visit\\(': 'await page.goto(',
        'cy\\.reload\\(\\)': 'await page.reload()'
      });
      const patterns = engine.getPatternsForCategory('navigation');
      expect(patterns.length).toBe(2);
    });

    it('should handle empty patterns object', () => {
      engine.registerPatterns('empty', {});
      const patterns = engine.getPatternsForCategory('empty');
      expect(patterns).toEqual([]);
    });
  });

  describe('applyPatterns', () => {
    it('should apply registered patterns to content', () => {
      engine.registerPatterns('test', {
        'hello': 'world'
      });
      const result = engine.applyPatterns('hello there');
      expect(result).toBe('world there');
    });

    it('should apply patterns with regex special characters', () => {
      engine.registerPatterns('test', {
        'cy\\.visit\\(': 'await page.goto('
      });
      const result = engine.applyPatterns("cy.visit('/home')");
      expect(result).toBe("await page.goto('/home')");
    });

    it('should apply multiple patterns in sequence', () => {
      engine.registerPatterns('test1', {
        'foo': 'bar'
      });
      engine.registerPatterns('test2', {
        'bar': 'baz'
      });
      const result = engine.applyPatterns('foo');
      expect(result).toBe('baz');
    });

    it('should apply patterns with capture groups', () => {
      engine.registerPatterns('test', {
        'visit\\(([^)]+)\\)': 'goto($1)'
      });
      const result = engine.applyPatterns("visit('/home')");
      expect(result).toBe("goto('/home')");
    });

    it('should apply patterns globally', () => {
      engine.registerPatterns('test', {
        'test': 'spec'
      });
      const result = engine.applyPatterns('test1 test2 test3');
      expect(result).toBe('spec1 spec2 spec3');
    });

    it('should handle content with no matches', () => {
      engine.registerPatterns('test', {
        'foo': 'bar'
      });
      const result = engine.applyPatterns('no matches here');
      expect(result).toBe('no matches here');
    });

    it('should handle empty content', () => {
      engine.registerPatterns('test', {
        'foo': 'bar'
      });
      const result = engine.applyPatterns('');
      expect(result).toBe('');
    });
  });

  describe('getPatternsForCategory', () => {
    it('should return empty array for non-existent category', () => {
      const patterns = engine.getPatternsForCategory('nonexistent');
      expect(patterns).toEqual([]);
    });

    it('should return patterns for existing category', () => {
      engine.registerPatterns('test', { 'a': 'b' });
      const patterns = engine.getPatternsForCategory('test');
      expect(patterns.length).toBe(1);
    });
  });

  describe('getCategories', () => {
    it('should return all registered categories', () => {
      engine.registerPatterns('nav', { 'a': 'b' });
      engine.registerPatterns('sel', { 'c': 'd' });
      const categories = engine.getCategories();
      expect(categories).toContain('nav');
      expect(categories).toContain('sel');
    });
  });

  describe('clear', () => {
    it('should clear all patterns', () => {
      engine.registerPatterns('test1', { 'a': 'b' });
      engine.registerPatterns('test2', { 'c': 'd' });
      engine.clear();
      expect(engine.getPatternsForCategory('test1')).toEqual([]);
      expect(engine.getPatternsForCategory('test2')).toEqual([]);
    });
  });

  describe('clearCategory', () => {
    it('should clear patterns for a specific category', () => {
      engine.registerPatterns('test1', { 'a': 'b' });
      engine.registerPatterns('test2', { 'c': 'd' });
      engine.clearCategory('test1');
      expect(engine.getPatternsForCategory('test1')).toEqual([]);
      expect(engine.getPatternsForCategory('test2').length).toBe(1);
    });
  });

  describe('applyPatternsWithTracking', () => {
    it('should return result and changes', () => {
      engine.registerPatterns('test', { 'foo': 'bar' });
      const { result, changes } = engine.applyPatternsWithTracking('foo baz foo');
      expect(result).toBe('bar baz bar');
      expect(changes.length).toBe(1);
      expect(changes[0].category).toBe('test');
    });
  });

  describe('getStats', () => {
    it('should track pattern applications', () => {
      engine.registerPatterns('test', { 'foo': 'bar' });
      engine.applyPatterns('foo');
      const stats = engine.getStats();
      expect(stats.patternsApplied).toBeGreaterThan(0);
    });
  });

  describe('resetStats', () => {
    it('should reset statistics', () => {
      engine.registerPatterns('test', { 'foo': 'bar' });
      engine.applyPatterns('foo');
      engine.resetStats();
      const stats = engine.getStats();
      expect(stats.patternsApplied).toBe(0);
    });
  });

  describe('registerTransformer', () => {
    it('should register a transformer function', () => {
      const fn = (content) => content.toUpperCase();
      engine.registerTransformer('upper', fn);
      expect(engine.getTransformerNames()).toContain('upper');
    });

    it('should register transformer with options', () => {
      const fn = (content) => content.trim();
      engine.registerTransformer('trim', fn, { priority: 10, description: 'Trim whitespace' });
      expect(engine.getTransformerNames()).toContain('trim');
    });
  });

  describe('applyTransformer', () => {
    it('should apply a registered transformer', () => {
      engine.registerTransformer('upper', (content) => content.toUpperCase());
      const result = engine.applyTransformer('upper', 'hello');
      expect(result).toBe('HELLO');
    });

    it('should throw for unknown transformer', () => {
      expect(() => engine.applyTransformer('nonexistent', 'content')).toThrow('Unknown transformer: nonexistent');
    });

    it('should increment transformersApplied stat', () => {
      engine.registerTransformer('upper', (content) => content.toUpperCase());
      engine.applyTransformer('upper', 'hello');
      expect(engine.getStats().transformersApplied).toBe(1);
    });
  });

  describe('applyTransformers', () => {
    it('should apply all transformers in order', () => {
      engine.registerTransformer('trim', (content) => content.trim());
      engine.registerTransformer('upper', (content) => content.toUpperCase());
      const result = engine.applyTransformers('  hello  ');
      expect(result).toBe('HELLO');
    });

    it('should apply only specified transformers', () => {
      engine.registerTransformer('trim', (content) => content.trim());
      engine.registerTransformer('upper', (content) => content.toUpperCase());
      const result = engine.applyTransformers('  hello  ', ['trim']);
      expect(result).toBe('hello');
    });

    it('should respect priority ordering', () => {
      engine.registerTransformer('addA', (c) => c + 'A', { priority: 1 });
      engine.registerTransformer('addB', (c) => c + 'B', { priority: 10 });
      const result = engine.applyTransformers('');
      expect(result).toBe('BA');
    });

    it('should handle empty content', () => {
      engine.registerTransformer('upper', (content) => content.toUpperCase());
      const result = engine.applyTransformers('');
      expect(result).toBe('');
    });
  });

  describe('getTransformerNames', () => {
    it('should return empty array when no transformers registered', () => {
      expect(engine.getTransformerNames()).toEqual([]);
    });

    it('should return names of all registered transformers', () => {
      engine.registerTransformer('a', (c) => c);
      engine.registerTransformer('b', (c) => c);
      const names = engine.getTransformerNames();
      expect(names).toContain('a');
      expect(names).toContain('b');
      expect(names).toHaveLength(2);
    });
  });

  describe('clone', () => {
    it('should create an independent copy', () => {
      engine.registerPatterns('test', { 'foo': 'bar' });
      engine.registerTransformer('upper', (c) => c.toUpperCase());

      const cloned = engine.clone();
      expect(cloned.getPatternsForCategory('test')).toHaveLength(1);
      expect(cloned.getTransformerNames()).toContain('upper');
    });

    it('should not share state with original', () => {
      engine.registerPatterns('test', { 'foo': 'bar' });
      const cloned = engine.clone();

      engine.registerPatterns('test', { 'baz': 'qux' });
      expect(engine.getPatternsForCategory('test')).toHaveLength(2);
      expect(cloned.getPatternsForCategory('test')).toHaveLength(1);
    });

    it('should produce a working clone', () => {
      engine.registerPatterns('test', { 'hello': 'world' });
      const cloned = engine.clone();
      const result = cloned.applyPatterns('hello');
      expect(result).toBe('world');
    });
  });

  describe('merge', () => {
    it('should merge patterns from another engine', () => {
      engine.registerPatterns('nav', { 'visit': 'goto' });

      const other = new PatternEngine();
      other.registerPatterns('sel', { 'get': 'locator' });

      engine.merge(other);
      expect(engine.getCategories()).toContain('nav');
      expect(engine.getCategories()).toContain('sel');
    });

    it('should merge transformers from another engine', () => {
      engine.registerTransformer('a', (c) => c);

      const other = new PatternEngine();
      other.registerTransformer('b', (c) => c);

      engine.merge(other);
      expect(engine.getTransformerNames()).toContain('a');
      expect(engine.getTransformerNames()).toContain('b');
    });

    it('should append patterns to existing categories by default', () => {
      engine.registerPatterns('test', { 'a': 'b' });

      const other = new PatternEngine();
      other.registerPatterns('test', { 'c': 'd' });

      engine.merge(other);
      expect(engine.getPatternsForCategory('test')).toHaveLength(2);
    });

    it('should overwrite categories when overwrite option is true', () => {
      engine.registerPatterns('test', { 'a': 'b' });

      const other = new PatternEngine();
      other.registerPatterns('test', { 'c': 'd' });

      engine.merge(other, { overwrite: true });
      expect(engine.getPatternsForCategory('test')).toHaveLength(1);
      const result = engine.applyPatterns('c');
      expect(result).toBe('d');
    });

    it('should not overwrite existing transformers by default', () => {
      engine.registerTransformer('t', (c) => c + '1');

      const other = new PatternEngine();
      other.registerTransformer('t', (c) => c + '2');

      engine.merge(other);
      const result = engine.applyTransformer('t', 'x');
      expect(result).toBe('x1');
    });

    it('should overwrite transformers when overwrite is true', () => {
      engine.registerTransformer('t', (c) => c + '1');

      const other = new PatternEngine();
      other.registerTransformer('t', (c) => c + '2');

      engine.merge(other, { overwrite: true });
      const result = engine.applyTransformer('t', 'x');
      expect(result).toBe('x2');
    });
  });
});
