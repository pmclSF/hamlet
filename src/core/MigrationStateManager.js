/**
 * Manages migration state in a .hamlet/ directory.
 *
 * Creates/reads/writes `.hamlet/` directory with atomic writes.
 * Supports resume: skip already-converted files.
 */

import fs from 'fs/promises';
import path from 'path';

const STATE_VERSION = 1;
const STATE_DIR = '.hamlet';
const STATE_FILE = 'state.json';
const TMP_FILE = 'state.tmp.json';

export class MigrationStateManager {
  /**
   * @param {string} projectRoot - Root directory of the project
   */
  constructor(projectRoot) {
    this.projectRoot = projectRoot;
    this.stateDir = path.join(projectRoot, STATE_DIR);
    this.statePath = path.join(this.stateDir, STATE_FILE);
    this.tmpPath = path.join(this.stateDir, TMP_FILE);
    this.state = null;
  }

  /**
   * Initialize the .hamlet directory and state.
   *
   * @param {Object} [options]
   * @param {string} [options.source] - Source framework
   * @param {string} [options.target] - Target framework
   * @returns {Promise<Object>} The initial state
   */
  async init(options = {}) {
    await fs.mkdir(this.stateDir, { recursive: true });

    this.state = {
      version: STATE_VERSION,
      startedAt: new Date().toISOString(),
      source: options.source || null,
      target: options.target || null,
      files: {},
    };

    await this.save();
    return this.state;
  }

  /**
   * Load state from disk.
   *
   * @returns {Promise<Object>} The loaded state
   * @throws {Error} If state file doesn't exist
   */
  async load() {
    try {
      const raw = await fs.readFile(this.statePath, 'utf8');
      this.state = JSON.parse(raw);
      return this.state;
    } catch (error) {
      if (error.code === 'ENOENT') {
        throw new Error(`No migration state found at ${this.statePath}. Run 'hamlet migrate' to start.`);
      }
      if (error instanceof SyntaxError) {
        // Corrupted state — re-initialize with warning
        console.warn(`Warning: Corrupted state at ${this.statePath}, re-initializing.`);
        return this.init();
      }
      throw error;
    }
  }

  /**
   * Save state to disk atomically (write tmp, rename).
   *
   * @returns {Promise<void>}
   */
  async save() {
    if (!this.state) throw new Error('No state to save. Call init() or load() first.');

    await fs.mkdir(this.stateDir, { recursive: true });
    const data = JSON.stringify(this.state, null, 2);
    await fs.writeFile(this.tmpPath, data, 'utf8');
    await fs.rename(this.tmpPath, this.statePath);
  }

  /**
   * Mark a file as converted.
   *
   * @param {string} filePath - Relative path of the file
   * @param {Object} [info]
   * @param {number} [info.confidence] - Confidence score (0-100)
   * @param {string} [info.error] - Error message if failed
   */
  markFileConverted(filePath, info = {}) {
    if (!this.state) throw new Error('No state loaded. Call init() or load() first.');

    this.state.files[filePath] = {
      status: info.error ? 'failed' : 'converted',
      convertedAt: new Date().toISOString(),
      confidence: info.confidence ?? null,
      error: info.error || null,
    };
  }

  /**
   * Mark a file as skipped.
   *
   * @param {string} filePath - Relative path of the file
   * @param {string} reason - Reason for skipping
   */
  markFileSkipped(filePath, reason) {
    if (!this.state) throw new Error('No state loaded. Call init() or load() first.');

    this.state.files[filePath] = {
      status: 'skipped',
      convertedAt: new Date().toISOString(),
      confidence: null,
      reason,
    };
  }

  /**
   * Check if a file has already been converted.
   *
   * @param {string} filePath - Relative path of the file
   * @returns {boolean}
   */
  isConverted(filePath) {
    if (!this.state) return false;
    const entry = this.state.files[filePath];
    return entry != null && entry.status === 'converted';
  }

  /**
   * Check if a file has failed conversion.
   *
   * @param {string} filePath - Relative path of the file
   * @returns {boolean}
   */
  isFailed(filePath) {
    if (!this.state) return false;
    const entry = this.state.files[filePath];
    return entry != null && entry.status === 'failed';
  }

  /**
   * Get migration status summary.
   *
   * @returns {Object}
   */
  getStatus() {
    if (!this.state) return { total: 0, converted: 0, failed: 0, skipped: 0, pending: 0 };

    const entries = Object.values(this.state.files);
    return {
      total: entries.length,
      converted: entries.filter(e => e.status === 'converted').length,
      failed: entries.filter(e => e.status === 'failed').length,
      skipped: entries.filter(e => e.status === 'skipped').length,
      source: this.state.source,
      target: this.state.target,
      startedAt: this.state.startedAt,
    };
  }

  /**
   * Check if .hamlet directory exists.
   *
   * @returns {Promise<boolean>}
   */
  async exists() {
    try {
      await fs.access(this.statePath);
      return true;
    } catch {
      return false;
    }
  }

  /**
   * Remove .hamlet directory completely.
   *
   * @returns {Promise<void>}
   */
  async reset() {
    await fs.rm(this.stateDir, { recursive: true, force: true });
    this.state = null;
  }
}
