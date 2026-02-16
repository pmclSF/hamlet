import path from 'path';
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
      ...options
    };

    this.stats = {
      total: 0,
      processed: 0,
      failed: 0,
      skipped: 0
    };
  }

  /**
   * Process multiple files in batches
   * @param {string[]} files - Array of file paths
   * @param {Function} processor - Processing function for each file
   * @returns {Promise<Object>} - Processing results
   */
  async processBatch(files, processor) {
    this.stats.total = files.length;
    const batches = this.createBatches(files);
    
    for (const batch of batches) {
      try {
        await Promise.all(
          batch.map(file => this.processFile(file, processor))
        );
      } catch (error) {
        logger.error('Batch processing error:', error);
      }
    }

    return this.stats;
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
      await processor(file);
      this.stats.processed++;
      logger.info(`Processed: ${path.basename(file)}`);
    } catch (error) {
      this.stats.failed++;
      logger.error(`Failed to process ${file}:`, error);
    }
  }

  /**
   * Get processing statistics
   * @returns {Object} - Processing statistics
   */
  getStats() {
    return {
      ...this.stats,
      success: this.stats.processed - this.stats.failed,
      successRate: `${((this.stats.processed - this.stats.failed) / this.stats.total * 100).toFixed(2)}%`
    };
  }
}