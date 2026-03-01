import path from 'path';
import { convertFile } from './fileConverter.js';
import { logUtils } from '../utils/helpers.js';

const logger = logUtils.createLogger('BatchProcessor');

/**
 * Handles processing of multiple test files in batches
 */
export class BatchProcessor {
  constructor(options = {}) {
    this.options = {
      batchSize: 5,
      concurrency: 3,
      ...options,
    };

    this.stats = {
      total: 0,
      processed: 0,
      failed: 0,
      skipped: 0,
    };
  }

  /**
   * Process multiple files in batches
   * @param {string[]} files - Array of file paths
   * @param {Function} processor - Processing function for each file
   * @returns {Promise<Object>} - Processing results
   */
  async processBatch(files, processor) {
    this.stats.processed = 0;
    this.stats.failed = 0;
    this.stats.skipped = 0;
    this.stats.total = files.length;
    const batches = this.createBatches(files);
    const results = [];

    const concurrency = Math.max(1, Number(this.options.concurrency) || 1);

    for (const batch of batches) {
      try {
        for (let i = 0; i < batch.length; i += concurrency) {
          const window = batch.slice(i, i + concurrency);
          const windowResults = await Promise.all(
            window.map((file) => this.processFile(file, processor))
          );
          results.push(...windowResults);
        }
      } catch (error) {
        logger.error('Batch processing error:', error);
      }
    }

    return {
      ...this.stats,
      results,
    };
  }

  /**
   * Create batches from file list
   * @param {string[]} files - Array of files
   * @returns {Array<string[]>} - Array of batches
   */
  createBatches(files) {
    const batches = [];
    for (let i = 0; i < files.length; i += this.options.batchSize) {
      batches.push(files.slice(i, i + this.options.batchSize));
    }
    return batches;
  }

  /**
   * Process a single file
   * @param {string} file - File path
   * @param {Function} processor - Processing function
   */
  async processFile(file, processor) {
    try {
      const result = await processor(file);
      if (result?.status === 'error') {
        this.stats.failed++;
        logger.error(
          `Failed to process ${file}:`,
          result.error || 'Unknown error'
        );
      } else if (result?.status === 'skipped') {
        this.stats.skipped++;
        logger.info(`Skipped: ${path.basename(file)}`);
      } else {
        this.stats.processed++;
        logger.info(`Processed: ${path.basename(file)}`);
      }
      return result;
    } catch (error) {
      this.stats.failed++;
      logger.error(`Failed to process ${file}:`, error);
      return {
        file,
        status: 'error',
        error: error.message,
      };
    }
  }

  /**
   * Get processing statistics
   * @returns {Object} - Processing statistics
   */
  getStats() {
    const successful = this.stats.processed;
    return {
      ...this.stats,
      success: successful,
      successRate:
        this.stats.total === 0
          ? '0.00%'
          : `${((successful / this.stats.total) * 100).toFixed(2)}%`,
    };
  }
}

/**
 * Process multiple test files in parallel
 * @param {string[]} files - Array of file paths
 * @param {Object} options - Processing options
 * @returns {Promise<Object>} - Processing results
 */
export async function processTestFiles(files, options = {}) {
  const batchProcessor = new BatchProcessor(options);
  const summary = await batchProcessor.processBatch(files, async (file) => {
    try {
      const outputPath =
        options.getOutputPath?.(file) ||
        file.replace(/\.cy\.(js|ts)$/, '.spec.$1');

      const converted = await convertFile(file, outputPath, options);
      return {
        file,
        outputPath: converted.outputPath || outputPath,
        status: 'success',
        metadata: converted.metadata,
        dependencies: converted.dependencies,
      };
    } catch (error) {
      return {
        file,
        error: error.message,
        status: 'error',
      };
    }
  });
  const results = summary.results || [];

  return {
    total: files.length,
    successful: results.filter((r) => r.status === 'success').length,
    failed: results.filter((r) => r.status === 'error').length,
    results,
  };
}
