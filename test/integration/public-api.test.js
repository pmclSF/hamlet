/**
 * Public API export contract test.
 *
 * Verifies that the main entry point (src/index.js) exports all documented
 * public symbols. This catches accidental removals during refactoring.
 * Does NOT test behavior — only export existence and type.
 */

import * as api from '../../src/index.js';

describe('Public API exports', () => {
  describe('functions', () => {
    it.each([
      'convertFile',
      'convertRepository',
      'processTestFiles',
      'validateTests',
      'generateReport',
      'convertCypressToPlaywright',
      'convertConfig',
    ])('should export %s as a function', (name) => {
      expect(typeof api[name]).toBe('function');
    });
  });

  describe('classes', () => {
    it.each([
      'RepositoryConverter',
      'BatchProcessor',
      'DependencyAnalyzer',
      'TestMetadataCollector',
      'TestValidator',
      'TypeScriptConverter',
      'PluginConverter',
      'VisualComparison',
      'TestMapper',
      'ConversionReporter',
    ])('should export %s as a function (class)', (name) => {
      expect(typeof api[name]).toBe('function');
    });
  });

  describe('utility namespaces', () => {
    it.each([
      'fileUtils',
      'stringUtils',
      'codeUtils',
      'testUtils',
      'reportUtils',
      'logUtils',
    ])('should export %s as an object', (name) => {
      expect(typeof api[name]).toBe('object');
      expect(api[name]).not.toBeNull();
    });
  });

  describe('constants', () => {
    it('should export VERSION as a string matching semver', () => {
      expect(typeof api.VERSION).toBe('string');
      expect(api.VERSION).toMatch(/^\d+\.\d+\.\d+/);
    });

    it('should export SUPPORTED_TEST_TYPES as a non-empty array', () => {
      expect(Array.isArray(api.SUPPORTED_TEST_TYPES)).toBe(true);
      expect(api.SUPPORTED_TEST_TYPES.length).toBeGreaterThan(0);
    });

    it('should export DEFAULT_OPTIONS as an object with expected keys', () => {
      expect(typeof api.DEFAULT_OPTIONS).toBe('object');
      expect(api.DEFAULT_OPTIONS).toHaveProperty('typescript');
      expect(api.DEFAULT_OPTIONS).toHaveProperty('batchSize');
    });
  });

  it('should export exactly the expected number of symbols', () => {
    const exportCount = Object.keys(api).length;
    // 7 functions + 10 classes + 6 utility namespaces + 3 constants = 26
    expect(exportCount).toBe(26);
  });
});
