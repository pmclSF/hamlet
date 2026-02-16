/**
 * Converts test framework configuration files between frameworks.
 *
 * Handles the 5-6 most common keys per framework pair.
 * Unrecognized keys get a HAMLET-TODO comment (not silent omission).
 */

import { TodoFormatter } from './TodoFormatter.js';

const JEST_TO_VITEST_KEYS = {
  testEnvironment: (value) => {
    if (value === 'jsdom') return { key: 'environment', value: '\'jsdom\'' };
    if (value === 'node') return { key: 'environment', value: '\'node\'' };
    return { key: 'environment', value: `'${value}'` };
  },
  setupFiles: (value) => ({ key: 'setupFiles', value }),
  setupFilesAfterFramework: (value) => ({ key: 'setupFiles', value }),
  testMatch: (value) => ({ key: 'include', value }),
  coverageThreshold: (value) => ({ key: 'coverage.thresholds', value }),
  testTimeout: (value) => ({ key: 'testTimeout', value }),
  clearMocks: (value) => ({ key: 'clearMocks', value }),
  resetMocks: (value) => ({ key: 'restoreMocks', value }),
  restoreMocks: (value) => ({ key: 'restoreMocks', value }),
};

const CYPRESS_TO_PLAYWRIGHT_KEYS = {
  baseUrl: (value) => ({ key: 'use.baseURL', value }),
  viewportWidth: (value, allConfig) => {
    const height = allConfig.viewportHeight || 720;
    return { key: 'use.viewport', value: `{ width: ${value}, height: ${height} }` };
  },
  viewportHeight: () => null, // Handled by viewportWidth
  retries: (value) => ({ key: 'retries', value }),
  specPattern: (value) => ({ key: 'testMatch', value }),
};

export class ConfigConverter {
  constructor() {
    this.formatter = new TodoFormatter('javascript');
  }

  /**
   * Convert a config file from one framework to another.
   *
   * @param {string} configContent - Content of the config file
   * @param {string} fromFramework - Source framework
   * @param {string} toFramework - Target framework
   * @returns {string} Converted config content
   */
  convert(configContent, fromFramework, toFramework) {
    const direction = `${fromFramework}-${toFramework}`;

    if (direction === 'jest-vitest') {
      return this._convertJestToVitest(configContent);
    }

    if (direction === 'cypress-playwright') {
      return this._convertCypressToPlaywright(configContent);
    }

    return this._addTodoHeader(configContent, fromFramework, toFramework);
  }

  /**
   * @param {string} content
   * @returns {string}
   */
  _convertJestToVitest(content) {
    const parsed = this._extractConfigObject(content);
    if (!parsed) {
      return this._addTodoHeader(content, 'jest', 'vitest');
    }

    const { keys } = parsed;
    const converted = [];
    const todos = [];

    converted.push('import { defineConfig } from \'vitest/config\';');
    converted.push('');
    converted.push('export default defineConfig({');
    converted.push('  test: {');

    for (const [key, value] of Object.entries(keys)) {
      const mapper = JEST_TO_VITEST_KEYS[key];
      if (mapper) {
        const result = mapper(value);
        if (result) {
          converted.push(`    ${result.key}: ${this._formatValue(result.value)},`);
        }
      } else {
        const todo = this.formatter.formatTodo({
          id: 'CONFIG-UNSUPPORTED',
          description: `Unsupported Jest config key: ${key}`,
          original: `${key}: ${JSON.stringify(value)}`,
          action: 'Manually convert this option to Vitest equivalent',
        });
        todos.push(todo);
      }
    }

    converted.push('  },');
    converted.push('});');

    if (todos.length > 0) {
      converted.push('');
      for (const todo of todos) {
        converted.push(todo);
      }
    }

    return converted.join('\n') + '\n';
  }

  /**
   * @param {string} content
   * @returns {string}
   */
  _convertCypressToPlaywright(content) {
    const parsed = this._extractConfigObject(content);
    if (!parsed) {
      return this._addTodoHeader(content, 'cypress', 'playwright');
    }

    const { keys } = parsed;
    const converted = [];
    const todos = [];

    converted.push('import { defineConfig, devices } from \'@playwright/test\';');
    converted.push('');
    converted.push('export default defineConfig({');

    for (const [key, value] of Object.entries(keys)) {
      const mapper = CYPRESS_TO_PLAYWRIGHT_KEYS[key];
      if (mapper) {
        const result = mapper(value, keys);
        if (result) {
          converted.push(`  ${result.key}: ${this._formatValue(result.value)},`);
        }
      } else {
        const todo = this.formatter.formatTodo({
          id: 'CONFIG-UNSUPPORTED',
          description: `Unsupported Cypress config key: ${key}`,
          original: `${key}: ${JSON.stringify(value)}`,
          action: 'Manually convert this option to Playwright equivalent',
        });
        todos.push(todo);
      }
    }

    converted.push('});');

    if (todos.length > 0) {
      converted.push('');
      for (const todo of todos) {
        converted.push(todo);
      }
    }

    return converted.join('\n') + '\n';
  }

  /**
   * Extract config keys from a config file (best-effort parsing).
   *
   * @param {string} content
   * @returns {{keys: Object, raw: string}|null}
   */
  _extractConfigObject(content) {
    // Try to parse as a simple object literal
    // Match module.exports = { ... } or export default { ... } or defineConfig({ ... })
    const patterns = [
      /module\.exports\s*=\s*\{([\s\S]*)\}/,
      /export\s+default\s+\{([\s\S]*)\}/,
      /defineConfig\s*\(\s*\{([\s\S]*)\}\s*\)/,
    ];

    for (const pattern of patterns) {
      const match = content.match(pattern);
      if (match) {
        const body = match[1];
        const keys = this._parseSimpleObject(body);
        if (keys) {
          return { keys, raw: body };
        }
      }
    }

    // Check for JS logic (conditional, function calls, etc.)
    if (/\bif\s*\(/.test(content) || /\bfunction\s/.test(content) || /=>\s*\{/.test(content)) {
      return null; // Too complex — will get HAMLET-TODO
    }

    return null;
  }

  /**
   * Parse a simple JS object body into key-value pairs.
   * Only handles simple literals: strings, numbers, booleans, arrays.
   *
   * @param {string} body
   * @returns {Object}
   */
  _parseSimpleObject(body) {
    const keys = {};
    // Match key: value patterns (simple values only)
    const keyValuePattern = /(\w+)\s*:\s*(?:'([^']*)'|"([^"]*)"|(\d+)|(\btrue\b|\bfalse\b)|\[([^\]]*)\])/g;

    let match;
    while ((match = keyValuePattern.exec(body)) !== null) {
      const key = match[1];
      const value = match[2] ?? match[3] ?? (match[4] ? Number(match[4]) : null) ??
                    (match[5] === 'true' ? true : match[5] === 'false' ? false : null) ??
                    (match[6] ? match[6] : null);
      if (value !== null) {
        keys[key] = value;
      }
    }

    return keys;
  }

  /**
   * Format a value for output.
   * @param {*} value
   * @returns {string}
   */
  _formatValue(value) {
    if (typeof value === 'string') {
      // Already formatted string (with quotes)
      if (value.startsWith('\'') || value.startsWith('"') || value.startsWith('{') || value.startsWith('[')) {
        return value;
      }
      return `'${value}'`;
    }
    if (typeof value === 'boolean' || typeof value === 'number') {
      return String(value);
    }
    return JSON.stringify(value);
  }

  /**
   * Add a HAMLET-TODO header to unconvertible config.
   * @param {string} content
   * @param {string} from
   * @param {string} to
   * @returns {string}
   */
  _addTodoHeader(content, from, to) {
    const todo = this.formatter.formatTodo({
      id: 'CONFIG-MANUAL',
      description: `Config conversion from ${from} to ${to} requires manual review`,
      original: `Full config file (${from})`,
      action: `Rewrite this config for ${to}`,
    });
    return todo + '\n\n' + content;
  }
}
