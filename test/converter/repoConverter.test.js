import { RepositoryConverter } from '../../src/converter/repoConverter.js';

describe('RepositoryConverter', () => {
  let converter;

  beforeEach(() => {
    converter = new RepositoryConverter();
  });

  describe('constructor', () => {
    it('should initialize with default options', () => {
      expect(converter.options.tempDir).toBe('.hamlet-temp');
      expect(converter.options.batchSize).toBe(5);
      expect(converter.options.preserveStructure).toBe(true);
      expect(converter.options.ignore).toEqual(['node_modules/**', '**/cypress/plugins/**']);
    });

    it('should accept custom options', () => {
      const custom = new RepositoryConverter({
        tempDir: '.custom-temp',
        batchSize: 10,
        preserveStructure: false
      });
      expect(custom.options.tempDir).toBe('.custom-temp');
      expect(custom.options.batchSize).toBe(10);
      expect(custom.options.preserveStructure).toBe(false);
    });

    it('should initialize stats with zero counts', () => {
      expect(converter.stats.totalFiles).toBe(0);
      expect(converter.stats.converted).toBe(0);
      expect(converter.stats.skipped).toBe(0);
      expect(converter.stats.errors).toEqual([]);
    });
  });

  describe('createBatches', () => {
    it('should create batches of the configured size', () => {
      const files = ['a.js', 'b.js', 'c.js', 'd.js', 'e.js', 'f.js', 'g.js'];
      converter.options.batchSize = 3;
      const batches = converter.createBatches(files);
      expect(batches).toHaveLength(3);
      expect(batches[0]).toEqual(['a.js', 'b.js', 'c.js']);
      expect(batches[1]).toEqual(['d.js', 'e.js', 'f.js']);
      expect(batches[2]).toEqual(['g.js']);
    });

    it('should handle empty file list', () => {
      const batches = converter.createBatches([]);
      expect(batches).toEqual([]);
    });

    it('should handle files fewer than batch size', () => {
      const files = ['a.js', 'b.js'];
      converter.options.batchSize = 5;
      const batches = converter.createBatches(files);
      expect(batches).toHaveLength(1);
      expect(batches[0]).toEqual(['a.js', 'b.js']);
    });

    it('should handle exact batch size', () => {
      const files = ['a.js', 'b.js', 'c.js', 'd.js', 'e.js'];
      converter.options.batchSize = 5;
      const batches = converter.createBatches(files);
      expect(batches).toHaveLength(1);
      expect(batches[0]).toHaveLength(5);
    });
  });

  describe('generateReport', () => {
    it('should return a report with stats', () => {
      converter.stats.totalFiles = 10;
      converter.stats.converted = 8;
      converter.stats.skipped = 2;
      converter.stats.errors.push({ file: 'bad.js', error: 'parse error' });

      const report = converter.generateReport();
      expect(report.stats.totalFiles).toBe(10);
      expect(report.stats.converted).toBe(8);
      expect(report.stats.skipped).toBe(2);
      expect(report.stats.errors).toHaveLength(1);
      expect(report.timestamp).toBeDefined();
      expect(report.configuration).toBeDefined();
    });

    it('should include configuration options in report', () => {
      const report = converter.generateReport();
      expect(report.configuration.tempDir).toBe('.hamlet-temp');
      expect(report.configuration.batchSize).toBe(5);
    });
  });
});
