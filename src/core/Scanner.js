/**
 * Scans a project directory and returns a list of files with metadata.
 *
 * Respects ignore patterns (node_modules, .git, dist, coverage, etc.)
 * and supports configurable include/exclude globs.
 */

import fs from 'fs/promises';
import path from 'path';

const DEFAULT_IGNORE = [
  'node_modules',
  '.git',
  'dist',
  'build',
  'coverage',
  '.nyc_output',
  '.cache',
  '.hamlet',
  '__pycache__',
  '.tox',
  '.mypy_cache',
];

/**
 * @param {string} name - File or directory name
 * @param {string[]} patterns - Glob-like patterns to match
 * @returns {boolean}
 */
function matchesAny(name, patterns) {
  for (const pattern of patterns) {
    if (pattern.startsWith('*.')) {
      const ext = pattern.slice(1);
      if (name.endsWith(ext)) return true;
    } else if (pattern.startsWith('**/*.')) {
      const ext = pattern.slice(4);
      if (name.endsWith(ext)) return true;
    } else if (name === pattern) {
      return true;
    }
  }
  return false;
}

export class Scanner {
  /**
   * Scan a directory recursively and return file metadata.
   *
   * @param {string} rootDir - Root directory to scan
   * @param {Object} [options]
   * @param {string[]} [options.ignore] - Additional directory/file names to ignore
   * @param {string[]} [options.include] - If provided, only include files matching these patterns
   * @param {string[]} [options.exclude] - Exclude files matching these patterns
   * @returns {Promise<Array<{path: string, relativePath: string, size: number}>>}
   */
  async scan(rootDir, options = {}) {
    const resolvedRoot = path.resolve(rootDir);
    const ignoreSet = new Set([...DEFAULT_IGNORE, ...(options.ignore || [])]);
    const include = options.include || [];
    const exclude = options.exclude || [];
    const results = [];

    await this._walk(resolvedRoot, resolvedRoot, ignoreSet, include, exclude, results);
    return results;
  }

  /**
   * @param {string} dir - Current directory
   * @param {string} rootDir - Root directory for relative paths
   * @param {Set<string>} ignoreSet - Directory/file names to skip
   * @param {string[]} include - Include patterns
   * @param {string[]} exclude - Exclude patterns
   * @param {Array} results - Accumulator
   */
  async _walk(dir, rootDir, ignoreSet, include, exclude, results) {
    let entries;
    try {
      entries = await fs.readdir(dir, { withFileTypes: true });
    } catch {
      // Permission denied or other read error — skip silently
      return;
    }

    for (const entry of entries) {
      const fullPath = path.join(dir, entry.name);

      if (ignoreSet.has(entry.name)) continue;

      if (entry.isDirectory()) {
        await this._walk(fullPath, rootDir, ignoreSet, include, exclude, results);
      } else if (entry.isFile()) {
        if (exclude.length > 0 && matchesAny(entry.name, exclude)) continue;
        if (include.length > 0 && !matchesAny(entry.name, include)) continue;

        let size = 0;
        try {
          const stat = await fs.stat(fullPath);
          size = stat.size;
        } catch {
          // If stat fails, still include file with size 0
        }

        results.push({
          path: fullPath,
          relativePath: path.relative(rootDir, fullPath),
          size,
        });
      }
    }
  }
}
