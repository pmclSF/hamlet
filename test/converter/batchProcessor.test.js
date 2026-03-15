import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import {
  BatchProcessor,
  processTestFiles,
} from '../../src/converter/batchProcessor.js';

describe('BatchProcessor', () => {
  let processor;

  beforeEach(() => {
    processor = new BatchProcessor();
  });

  describe('constructor', () => {
    it('should initialize with default options', () => {
      expect(processor.options.batchSize).toBe(5);
      expect(processor.options.concurrency).toBe(3);
    });

    it('should accept custom options', () => {
      const custom = new BatchProcessor({ batchSize: 10, concurrency: 5 });
      expect(custom.options.batchSize).toBe(10);
      expect(custom.options.concurrency).toBe(5);
    });

    it('should initialize stats to zero', () => {
      expect(processor.stats.total).toBe(0);
      expect(processor.stats.processed).toBe(0);
      expect(processor.stats.failed).toBe(0);
      expect(processor.stats.skipped).toBe(0);
    });
  });

  describe('createBatches', () => {
    it('should split files into batches of configured size', () => {
      const files = ['a', 'b', 'c', 'd', 'e', 'f', 'g'];
      const batches = processor.createBatches(files);
      expect(batches).toHaveLength(2);
      expect(batches[0]).toHaveLength(5);
      expect(batches[1]).toHaveLength(2);
    });

    it('should handle empty file list', () => {
      const batches = processor.createBatches([]);
      expect(batches).toHaveLength(0);
    });

    it('should handle files less than batch size', () => {
      const batches = processor.createBatches(['a', 'b']);
      expect(batches).toHaveLength(1);
      expect(batches[0]).toEqual(['a', 'b']);
    });

    it('should respect custom batch size', () => {
      const custom = new BatchProcessor({ batchSize: 2 });
      const batches = custom.createBatches(['a', 'b', 'c', 'd', 'e']);
      expect(batches).toHaveLength(3);
    });
  });

  describe('processBatch', () => {
    it('should process all files and track stats', async () => {
      const files = ['file1.js', 'file2.js', 'file3.js'];
      const processed = [];
      const processorFn = async (file) => { processed.push(file); };

      await processor.processBatch(files, processorFn);

      expect(processor.stats.total).toBe(3);
      expect(processor.stats.processed).toBe(3);
      expect(processor.stats.failed).toBe(0);
      expect(processed).toEqual(files);
    });

    it('should track failed files', async () => {
      const files = ['good.js', 'bad.js', 'good2.js'];
      const processorFn = async (file) => {
        if (file === 'bad.js') throw new Error('failed');
      };

      await processor.processBatch(files, processorFn);

      expect(processor.stats.processed).toBe(2);
      expect(processor.stats.failed).toBe(1);
    });

    it('should handle empty file list', async () => {
      const result = await processor.processBatch([], async () => {});
      expect(processor.stats.total).toBe(0);
      expect(Array.isArray(result.results)).toBe(true);
    });

    it('should expose per-file results in return value', async () => {
      const files = ['a.js', 'b.js'];
      const result = await processor.processBatch(files, async (file) => ({
        file,
        status: 'success',
      }));

      expect(result.total).toBe(2);
      expect(result.results).toHaveLength(2);
      expect(result.results[0].status).toBe('success');
    });

    it('should respect configured concurrency limit', async () => {
      const files = ['a.js', 'b.js', 'c.js', 'd.js', 'e.js'];
      const limited = new BatchProcessor({ batchSize: 5, concurrency: 2 });
      let active = 0;
      let maxActive = 0;

      await limited.processBatch(files, async () => {
        active++;
        maxActive = Math.max(maxActive, active);
        await new Promise((resolve) => setTimeout(resolve, 20));
        active--;
        return { status: 'success' };
      });

      expect(maxActive).toBeLessThanOrEqual(2);
    });
  });

  describe('processFile', () => {
    it('should increment processed count on success', async () => {
      await processor.processFile('test.js', async () => {});
      expect(processor.stats.processed).toBe(1);
      expect(processor.stats.failed).toBe(0);
    });

    it('should increment failed count on error', async () => {
      await processor.processFile('test.js', async () => {
        throw new Error('fail');
      });
      expect(processor.stats.failed).toBe(1);
    });
  });

  describe('getStats', () => {
    it('should include success count and rate', async () => {
      const files = ['a.js', 'b.js', 'c.js'];
      await processor.processBatch(files, async () => {});

      const stats = processor.getStats();
      expect(stats.success).toBe(3);
      expect(stats.successRate).toBe('100.00%');
    });
  });
});

describe('processTestFiles', () => {
  let tmpDir;

  beforeEach(async () => {
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'terrain-batch-'));
  });

  afterEach(async () => {
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  it('should return success/failure counts and result list without crashing', async () => {
    const goodSource = path.join(tmpDir, 'good.cy.js');
    const goodOutput = path.join(tmpDir, 'good.spec.js');
    const missingSource = path.join(tmpDir, 'missing.cy.js');

    await fs.writeFile(
      goodSource,
      `
describe('good', () => {
  it('works', () => {
    cy.visit('/');
  });
});
`
    );

    const result = await processTestFiles([goodSource, missingSource], {
      getOutputPath: (file) =>
        file === goodSource ? goodOutput : path.join(tmpDir, 'missing.spec.js'),
    });

    expect(result.total).toBe(2);
    expect(result.successful).toBe(1);
    expect(result.failed).toBe(1);
    expect(result.results).toHaveLength(2);
  });
});
