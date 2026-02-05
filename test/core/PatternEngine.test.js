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
});
