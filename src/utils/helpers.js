import fs from 'fs/promises';
import path from 'path';
import chalk from 'chalk';

/**
 * File system helpers
 */
export const fileUtils = {
  /**
   * Check if file exists
   * @param {string} filePath - Path to file
   * @returns {Promise<boolean>} - True if file exists
   */
  async fileExists(filePath) {
    try {
      await fs.access(filePath);
      return true;
    } catch {
      return false;
    }
  },

  /**
   * Ensure directory exists
   * @param {string} dir - Directory path
   */
  async ensureDir(dir) {
    try {
      await fs.mkdir(dir, { recursive: true });
    } catch (error) {
      console.error(chalk.red('Error creating directory:'), error);
      throw error;
    }
  },

  /**
   * Copy directory recursively
   * @param {string} src - Source directory
   * @param {string} dest - Destination directory
   */
  async copyDir(src, dest) {
    await fileUtils.ensureDir(dest);
    const entries = await fs.readdir(src, { withFileTypes: true });

    for (const entry of entries) {
      const srcPath = path.join(src, entry.name);
      const destPath = path.join(dest, entry.name);

      if (entry.isDirectory()) {
        await fileUtils.copyDir(srcPath, destPath);
      } else {
        await fs.copyFile(srcPath, destPath);
      }
    }
  },

  /**
   * Get all files in directory recursively
   * @param {string} dir - Directory path
   * @param {RegExp} pattern - File pattern to match
   * @returns {Promise<string[]>} - Array of file paths
   */
  async getFiles(dir, pattern) {
    const files = [];

    async function scan(directory) {
      const entries = await fs.readdir(directory, { withFileTypes: true });

      for (const entry of entries) {
        const fullPath = path.join(directory, entry.name);

        if (entry.isDirectory()) {
          await scan(fullPath);
        } else if (!pattern || pattern.test(entry.name)) {
          files.push(fullPath);
        }
      }
    }

    await scan(dir);
    return files;
  },

  /**
   * Clean directory
   * @param {string} dir - Directory to clean
   * @param {RegExp} exclude - Pattern for files to exclude
   */
  async cleanDir(dir, exclude) {
    try {
      const entries = await fs.readdir(dir, { withFileTypes: true });

      for (const entry of entries) {
        const fullPath = path.join(dir, entry.name);

        if (exclude && exclude.test(entry.name)) {
          continue;
        }

        if (entry.isDirectory()) {
          await fs.rm(fullPath, { recursive: true });
        } else {
          await fs.unlink(fullPath);
        }
      }
    } catch (error) {
      if (error.code !== 'ENOENT') {
        throw error;
      }
    }
  },
};

/**
 * String manipulation helpers
 */
export const stringUtils = {
  /**
   * Convert camelCase to kebab-case
   * @param {string} str - String to convert
   * @returns {string} - Converted string
   */
  camelToKebab(str) {
    return str.replace(/([a-z0-9])([A-Z])/g, '$1-$2').toLowerCase();
  },

  /**
   * Convert kebab-case to camelCase
   * @param {string} str - String to convert
   * @returns {string} - Converted string
   */
  kebabToCamel(str) {
    return str.replace(/-([a-z])/g, (g) => g[1].toUpperCase());
  },

  /**
   * Calculate similarity between two strings
   * @param {string} str1 - First string
   * @param {string} str2 - Second string
   * @returns {number} - Similarity score (0-1)
   */
  calculateSimilarity(str1, str2) {
    const len1 = str1.length;
    const len2 = str2.length;
    const matrix = Array(len2 + 1)
      .fill(null)
      .map(() => Array(len1 + 1).fill(null));

    for (let i = 0; i <= len1; i++) matrix[0][i] = i;
    for (let j = 0; j <= len2; j++) matrix[j][0] = j;

    for (let j = 1; j <= len2; j++) {
      for (let i = 1; i <= len1; i++) {
        const substitute =
          matrix[j - 1][i - 1] + (str1[i - 1] !== str2[j - 1] ? 1 : 0);
        matrix[j][i] = Math.min(
          matrix[j - 1][i] + 1,
          matrix[j][i - 1] + 1,
          substitute,
        );
      }
    }

    return 1 - matrix[len2][len1] / Math.max(len1, len2);
  },
};

/**
 * Code manipulation helpers
 */
export const codeUtils = {
  /**
   * Extract imports from code
   * @param {string} content - Code content
   * @returns {Object[]} - Array of import information
   */
  extractImports(content) {
    const imports = [];
    const importRegex =
      /import\s+(?:{[^}]+}|\*\s+as\s+\w+|\w+)\s+from\s+['"]([^'"]+)['"]/g;
    let match;

    while ((match = importRegex.exec(content)) !== null) {
      imports.push({
        statement: match[0],
        source: match[1],
      });
    }

    return imports;
  },

  /**
   * Extract exports from code
   * @param {string} content - Code content
   * @returns {Object[]} - Array of export information
   */
  extractExports(content) {
    const exports = [];
    const exportRegex =
      /export\s+(?:default\s+)?(?:class|function|const|let|var)\s+(\w+)/g;
    let match;

    while ((match = exportRegex.exec(content)) !== null) {
      exports.push({
        statement: match[0],
        name: match[1],
      });
    }

    return exports;
  },

  /**
   * Format code with consistent style
   * @param {string} content - Code content
   * @returns {string} - Formatted code
   */
  formatCode(content) {
    return (
      content
        // Add spaces around operators
        .replace(/([+\-*/%=<>!&|])?=(?!=)/g, ' $1= ')
        // Add spaces after commas
        .replace(/,([^\s])/g, ', $1')
        // Add spaces after keywords
        .replace(/(if|for|while|switch|catch)\(/g, '$1 (')
        // Add newline after semicolons
        .replace(/;(?!\s*$)/g, ';\n')
        // Trim trailing whitespace
        .replace(/\s+$/gm, '')
        // Ensure single newline at end
        .trim() + '\n'
    );
  },
};

/**
 * Test pattern helpers
 */
export const testUtils = {
  /**
   * Extract test cases from content
   * @param {string} content - Test content
   * @returns {Object[]} - Array of test information
   */
  extractTestCases(content) {
    const tests = [];
    const testRegex =
      /(?:it|test)\s*\(\s*['"`](.*?)['"`]\s*,\s*(?:async\s*)?\([^)]*\)\s*=>\s*{([\s\S]*?)}\s*\)/g;
    let match;

    while ((match = testRegex.exec(content)) !== null) {
      tests.push({
        description: match[1],
        body: match[2].trim(),
      });
    }

    return tests;
  },

  /**
   * Extract assertions from test
   * @param {string} content - Test content
   * @returns {Object[]} - Array of assertion information
   */
  extractAssertions(content) {
    const assertions = [];
    const assertionRegex =
      /(?:expect|assert)\s*\((.*?)\)\.(?:to|should)\.(.*?)\((.*?)\)/g;
    let match;

    while ((match = assertionRegex.exec(content)) !== null) {
      assertions.push({
        target: match[1].trim(),
        assertion: match[2].trim(),
        params: match[3].trim(),
      });
    }

    return assertions;
  },
};

/**
 * Progress reporting helpers
 */
export const reportUtils = {
  /**
   * Create progress bar
   * @param {number} total - Total steps
   * @returns {Object} - Progress bar methods
   */
  createProgressBar(total) {
    let current = 0;
    const width = 40;

    return {
      update(value) {
        current = value;
        const percentage = Math.round((current / total) * 100);
        const filled = Math.round((width * current) / total);
        const bar = '█'.repeat(filled) + '-'.repeat(width - filled);
        process.stdout.write(`\r[${bar}] ${percentage}%`);
      },
      complete() {
        this.update(total);
        process.stdout.write('\n');
      },
    };
  },

  /**
   * Format duration
   * @param {number} ms - Duration in milliseconds
   * @returns {string} - Formatted duration
   */
  formatDuration(ms) {
    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);

    if (hours > 0) {
      return `${hours}h ${minutes % 60}m ${seconds % 60}s`;
    } else if (minutes > 0) {
      return `${minutes}m ${seconds % 60}s`;
    } else {
      return `${seconds}s`;
    }
  },
};

/**
 * Logging helpers
 */
export const logUtils = {
  /**
   * Log error with context
   * @param {Error} error - Error object
   * @param {string} context - Error context
   */
  logError(error, context) {
    console.error(chalk.red(`Error in ${context}:`));
    console.error(chalk.red(error.message));
    if (error.stack) {
      console.error(chalk.gray(error.stack.split('\n').slice(1).join('\n')));
    }
  },

  /**
   * Log warning with details
   * @param {string} message - Warning message
   * @param {Object} details - Warning details
   */
  logWarning(message, details = {}) {
    console.warn(chalk.yellow('Warning:'), message);
    if (Object.keys(details).length > 0) {
      console.warn(chalk.gray(JSON.stringify(details, null, 2)));
    }
  },

  /**
   * Create scoped logger
   * @param {string} scope - Logger scope
   * @returns {Object} - Scoped logger methods
   */
  createLogger(scope) {
    return {
      info: (message) => console.log(chalk.blue(`[${scope}]`), message),
      success: (message) => console.log(chalk.green(`[${scope}]`), message),
      warn: (message) => console.warn(chalk.yellow(`[${scope}]`), message),
      error: (message) => console.error(chalk.red(`[${scope}]`), message),
    };
  },
};
