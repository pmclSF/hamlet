/**
 * Safety mechanisms for migration operations.
 *
 * Atomic writes, backups, shutdown handlers.
 * Never overwrites source files.
 */

import fs from "fs/promises";
import path from "path";

export class SafetyManager {
  /**
   * @param {string} projectRoot - Root directory of the project
   */
  constructor(projectRoot) {
    this.projectRoot = projectRoot;
    this.backupDir = path.join(projectRoot, ".hamlet", "backups");
  }

  /**
   * Write content to a file atomically (write to tmp, rename).
   *
   * @param {string} filePath - Target file path
   * @param {string} content - Content to write
   * @returns {Promise<void>}
   */
  async atomicWrite(filePath, content) {
    const dir = path.dirname(filePath);
    await fs.mkdir(dir, { recursive: true });

    const tmpPath = filePath + ".hamlet-tmp";
    await fs.writeFile(tmpPath, content, "utf8");
    await fs.rename(tmpPath, filePath);
  }

  /**
   * Create a backup of a file before modification.
   *
   * @param {string} filePath - Absolute path of the file to backup
   * @returns {Promise<string>} Path to the backup file
   */
  async backup(filePath) {
    await fs.mkdir(this.backupDir, { recursive: true });

    const relativePath = path.relative(this.projectRoot, filePath);
    const backupPath = path.join(this.backupDir, relativePath);

    await fs.mkdir(path.dirname(backupPath), { recursive: true });
    await fs.copyFile(filePath, backupPath);

    return backupPath;
  }

  /**
   * Check if a file is within the project root (safety check against path traversal).
   *
   * @param {string} filePath - Path to check
   * @returns {boolean}
   */
  isWithinProject(filePath) {
    const resolved = path.resolve(filePath);
    const root = path.resolve(this.projectRoot);
    return resolved.startsWith(root + path.sep) || resolved === root;
  }

  /**
   * Register a shutdown handler that saves state before exit.
   *
   * @param {import('./MigrationStateManager.js').MigrationStateManager} stateManager
   * @returns {Function} Cleanup function to remove the handler
   */
  registerShutdownHandler(stateManager) {
    const handler = async () => {
      try {
        await stateManager.save();
      } catch {
        // Best effort — don't crash on exit
      }
    };

    const syncHandler = () => {
      // Can't await in signal handlers, but we try
      handler();
    };

    process.on("SIGINT", syncHandler);
    process.on("SIGTERM", syncHandler);

    // Return cleanup function
    return () => {
      process.removeListener("SIGINT", syncHandler);
      process.removeListener("SIGTERM", syncHandler);
    };
  }
}
