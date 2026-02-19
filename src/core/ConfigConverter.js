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

const VITEST_TO_JEST_KEYS = {
  environment: (value) => ({ key: 'testEnvironment', value }),
  setupFiles: (value) => ({ key: 'setupFiles', value }),
  include: (value) => ({ key: 'testMatch', value }),
  testTimeout: (value) => ({ key: 'testTimeout', value }),
  clearMocks: (value) => ({ key: 'clearMocks', value }),
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
  defaultCommandTimeout: (value) => ({ key: 'timeout', value }),
};

const PLAYWRIGHT_TO_CYPRESS_KEYS = {
  baseURL: (value) => ({ key: 'baseUrl', value }),
  timeout: (value) => ({ key: 'defaultCommandTimeout', value }),
  retries: (value) => ({ key: 'retries', value }),
  testMatch: (value) => ({ key: 'specPattern', value }),
  testDir: (value) => ({ key: 'specPattern', value }),
};

const WDIO_TO_PLAYWRIGHT_KEYS = {
  baseUrl: (value) => ({ key: 'use.baseURL', value }),
  waitforTimeout: (value) => ({ key: 'timeout', value }),
  specs: (value) => ({ key: 'testMatch', value }),
  maxInstances: (value) => ({ key: 'workers', value }),
};

const PLAYWRIGHT_TO_WDIO_KEYS = {
  baseURL: (value) => ({ key: 'baseUrl', value }),
  timeout: (value) => ({ key: 'waitforTimeout', value }),
  testMatch: (value) => ({ key: 'specs', value }),
  workers: (value) => ({ key: 'maxInstances', value }),
};

const WDIO_TO_CYPRESS_KEYS = {
  baseUrl: (value) => ({ key: 'baseUrl', value }),
  waitforTimeout: (value) => ({ key: 'defaultCommandTimeout', value }),
  specs: (value) => ({ key: 'specPattern', value }),
};

const CYPRESS_TO_WDIO_KEYS = {
  baseUrl: (value) => ({ key: 'baseUrl', value }),
  defaultCommandTimeout: (value) => ({ key: 'waitforTimeout', value }),
  specPattern: (value) => ({ key: 'specs', value }),
};

const MOCHA_TO_JEST_KEYS = {
  timeout: (value) => ({ key: 'testTimeout', value }),
  spec: (value) => ({ key: 'testMatch', value }),
  require: (value) => ({ key: 'setupFiles', value }),
  bail: (value) => ({ key: 'bail', value }),
};

const JASMINE_TO_JEST_KEYS = {
  spec_dir: (value) => ({ key: 'roots', value }),
  spec_files: (value) => ({ key: 'testMatch', value }),
  helpers: (value) => ({ key: 'setupFiles', value }),
  random: (value) => ({ key: 'randomize', value }),
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

    const handlers = {
      'jest-vitest': (c) => this._convertJestToVitest(c),
      'vitest-jest': (c) => this._convertVitestToJest(c),
      'cypress-playwright': (c) => this._convertCypressToPlaywright(c),
      'playwright-cypress': (c) => this._convertPlaywrightToCypress(c),
      'webdriverio-playwright': (c) => this._convertWdioToPlaywright(c),
      'playwright-webdriverio': (c) => this._convertPlaywrightToWdio(c),
      'webdriverio-cypress': (c) => this._convertWdioToCypress(c),
      'cypress-webdriverio': (c) => this._convertCypressToWdio(c),
      'mocha-jest': (c) => this._convertMochaToJest(c),
      'jasmine-jest': (c) => this._convertJasmineToJest(c),
      'pytest-unittest': (c) => this._convertPytestToUnittest(c),
      'testng-junit5': (c) => this._convertTestngToJunit5(c),
      'junit4-junit5': (c) => this._convertJunit4DepsToJunit5(c),
    };

    const handler = handlers[direction];
    if (handler) {
      return handler(configContent);
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
    return this._renderVitestConfig(keys, JEST_TO_VITEST_KEYS, 'Jest');
  }

  /**
   * @param {string} content
   * @returns {string}
   */
  _convertVitestToJest(content) {
    const parsed = this._extractConfigObject(content);
    if (!parsed) {
      return this._addTodoHeader(content, 'vitest', 'jest');
    }

    const { keys } = parsed;
    return this._renderJestConfig(keys, VITEST_TO_JEST_KEYS, 'Vitest');
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
    return this._renderPlaywrightConfig(keys, CYPRESS_TO_PLAYWRIGHT_KEYS, 'Cypress');
  }

  /**
   * @param {string} content
   * @returns {string}
   */
  _convertPlaywrightToCypress(content) {
    const parsed = this._extractConfigObject(content);
    if (!parsed) {
      return this._addTodoHeader(content, 'playwright', 'cypress');
    }

    const { keys } = parsed;
    return this._renderCypressConfig(keys, PLAYWRIGHT_TO_CYPRESS_KEYS, 'Playwright');
  }

  /**
   * @param {string} content
   * @returns {string}
   */
  _convertWdioToPlaywright(content) {
    const parsed = this._extractConfigObject(content);
    if (!parsed) {
      return this._addTodoHeader(content, 'webdriverio', 'playwright');
    }

    const { keys } = parsed;
    return this._renderPlaywrightConfig(keys, WDIO_TO_PLAYWRIGHT_KEYS, 'WebdriverIO');
  }

  /**
   * @param {string} content
   * @returns {string}
   */
  _convertPlaywrightToWdio(content) {
    const parsed = this._extractConfigObject(content);
    if (!parsed) {
      return this._addTodoHeader(content, 'playwright', 'webdriverio');
    }

    const { keys } = parsed;
    return this._renderWdioConfig(keys, PLAYWRIGHT_TO_WDIO_KEYS, 'Playwright');
  }

  /**
   * @param {string} content
   * @returns {string}
   */
  _convertWdioToCypress(content) {
    const parsed = this._extractConfigObject(content);
    if (!parsed) {
      return this._addTodoHeader(content, 'webdriverio', 'cypress');
    }

    const { keys } = parsed;
    return this._renderCypressConfig(keys, WDIO_TO_CYPRESS_KEYS, 'WebdriverIO');
  }

  /**
   * @param {string} content
   * @returns {string}
   */
  _convertCypressToWdio(content) {
    const parsed = this._extractConfigObject(content);
    if (!parsed) {
      return this._addTodoHeader(content, 'cypress', 'webdriverio');
    }

    const { keys } = parsed;
    return this._renderWdioConfig(keys, CYPRESS_TO_WDIO_KEYS, 'Cypress');
  }

  /**
   * @param {string} content
   * @returns {string}
   */
  _convertMochaToJest(content) {
    const parsed = this._parseYamlSimple(content) || this._parseJsonSimple(content) || this._extractConfigObject(content);
    if (!parsed) {
      return this._addTodoHeader(content, 'mocha', 'jest');
    }

    const { keys } = parsed;
    return this._renderJestConfig(keys, MOCHA_TO_JEST_KEYS, 'Mocha');
  }

  /**
   * @param {string} content
   * @returns {string}
   */
  _convertJasmineToJest(content) {
    const parsed = this._parseJsonSimple(content) || this._extractConfigObject(content);
    if (!parsed) {
      return this._addTodoHeader(content, 'jasmine', 'jest');
    }

    const { keys } = parsed;
    return this._renderJestConfig(keys, JASMINE_TO_JEST_KEYS, 'Jasmine');
  }

  /**
   * @param {string} content
   * @returns {string}
   */
  _convertPytestToUnittest(content) {
    const formatter = new TodoFormatter('python');
    const parsed = this._parseIniFile(content);
    if (!parsed) {
      return formatter.formatTodo({
        id: 'CONFIG-MANUAL',
        description: 'Config conversion from pytest to unittest requires manual review',
        original: 'Full config file (pytest)',
        action: 'Rewrite this config for unittest',
      }) + '\n\n' + content;
    }

    const { keys } = parsed;
    const lines = [];
    lines.push('# unittest configuration');
    lines.push('# Converted from pytest.ini by Hamlet');
    lines.push('#');

    if (keys.testpaths) {
      lines.push(`# Test discovery directory: ${keys.testpaths}`);
      lines.push(`# Run: python -m unittest discover -s ${keys.testpaths}`);
    } else {
      lines.push('# Run: python -m unittest discover');
    }

    if (keys.python_files) {
      lines.push(`# File pattern: ${keys.python_files}`);
      lines.push(`# Run with: python -m unittest discover -p "${keys.python_files}"`);
    }

    if (keys.python_classes) {
      lines.push(`# Class pattern: ${keys.python_classes}`);
    }

    if (keys.python_functions) {
      lines.push(`# Function pattern: ${keys.python_functions}`);
    }

    // Add TODO for any remaining keys
    const knownKeys = new Set(['testpaths', 'python_files', 'python_classes', 'python_functions']);
    for (const [key, value] of Object.entries(keys)) {
      if (!knownKeys.has(key)) {
        lines.push(formatter.formatTodo({
          id: 'CONFIG-UNSUPPORTED',
          description: `Unsupported pytest config key: ${key}`,
          original: `${key} = ${value}`,
          action: 'No direct unittest equivalent — handle manually',
        }));
      }
    }

    return lines.join('\n') + '\n';
  }

  /**
   * @param {string} content
   * @returns {string}
   */
  _convertTestngToJunit5(content) {
    const formatter = new TodoFormatter('java');
    const parsed = this._parseTestngXml(content);
    if (!parsed) {
      return formatter.formatTodo({
        id: 'CONFIG-MANUAL',
        description: 'Config conversion from TestNG to JUnit5 requires manual review',
        original: 'Full config file (testng.xml)',
        action: 'Rewrite this config for JUnit5',
      }) + '\n\n' + content;
    }

    const { classes, suiteName } = parsed;
    const lines = [];
    lines.push('import org.junit.platform.suite.api.Suite;');
    lines.push('import org.junit.platform.suite.api.SelectClasses;');
    lines.push('import org.junit.platform.suite.api.SuiteDisplayName;');
    lines.push('');
    lines.push('@Suite');
    lines.push(`@SuiteDisplayName("${suiteName || 'Converted Suite'}")`);

    if (classes.length > 0) {
      const classRefs = classes.map(c => `${c}.class`).join(', ');
      lines.push(`@SelectClasses({${classRefs}})`);
    }

    lines.push(`public class ${this._sanitizeClassName(suiteName || 'ConvertedSuite')}Test {`);
    lines.push('  // JUnit5 Suite — test classes are selected via @SelectClasses');
    lines.push('}');

    return lines.join('\n') + '\n';
  }

  /**
   * @param {string} content
   * @returns {string}
   */
  _convertJunit4DepsToJunit5(content) {
    const formatter = new TodoFormatter('java');

    // Maven POM
    if (content.includes('<dependency>') || content.includes('<groupId>')) {
      let result = content;
      result = result.replace(
        /<groupId>junit<\/groupId>\s*\n\s*<artifactId>junit<\/artifactId>/g,
        '<groupId>org.junit.jupiter</groupId>\n    <artifactId>junit-jupiter</artifactId>',
      );
      result = result.replace(
        /<version>4\.\d+(\.\d+)?<\/version>/g,
        '<version>5.10.0</version>',
      );
      if (result !== content) {
        return result;
      }
    }

    // Gradle
    if (content.includes('testImplementation') || content.includes('testCompile')) {
      let result = content;
      result = result.replace(
        /testImplementation\s+['"]junit:junit:4\.[^'"]+['"]/g,
        'testImplementation \'org.junit.jupiter:junit-jupiter:5.10.0\'',
      );
      result = result.replace(
        /testCompile\s+['"]junit:junit:4\.[^'"]+['"]/g,
        'testImplementation \'org.junit.jupiter:junit-jupiter:5.10.0\'',
      );
      if (result !== content) {
        return result;
      }
    }

    return formatter.formatTodo({
      id: 'CONFIG-MANUAL',
      description: 'Config conversion from JUnit4 to JUnit5 requires manual review',
      original: 'Full config file (JUnit4)',
      action: 'Update dependencies to JUnit Jupiter 5.x',
    }) + '\n\n' + content;
  }

  // --- Render helpers for common output formats ---

  /**
   * Render a Vitest config file from parsed keys.
   * @param {Object} keys
   * @param {Object} keyMap
   * @param {string} sourceName
   * @returns {string}
   */
  _renderVitestConfig(keys, keyMap, sourceName) {
    const converted = [];
    const todos = [];

    converted.push('import { defineConfig } from \'vitest/config\';');
    converted.push('');
    converted.push('export default defineConfig({');
    converted.push('  test: {');

    for (const [key, value] of Object.entries(keys)) {
      const mapper = keyMap[key];
      if (mapper) {
        const result = mapper(value, keys);
        if (result) {
          converted.push(`    ${result.key}: ${this._formatValue(result.value)},`);
        }
      } else {
        todos.push(this.formatter.formatTodo({
          id: 'CONFIG-UNSUPPORTED',
          description: `Unsupported ${sourceName} config key: ${key}`,
          original: `${key}: ${JSON.stringify(value)}`,
          action: 'Manually convert this option to Vitest equivalent',
        }));
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
   * Render a Jest config file from parsed keys.
   * @param {Object} keys
   * @param {Object} keyMap
   * @param {string} sourceName
   * @returns {string}
   */
  _renderJestConfig(keys, keyMap, sourceName) {
    const converted = [];
    const todos = [];

    converted.push('module.exports = {');

    for (const [key, value] of Object.entries(keys)) {
      const mapper = keyMap[key];
      if (mapper) {
        const result = mapper(value, keys);
        if (result) {
          converted.push(`  ${result.key}: ${this._formatValue(result.value)},`);
        }
      } else {
        todos.push(this.formatter.formatTodo({
          id: 'CONFIG-UNSUPPORTED',
          description: `Unsupported ${sourceName} config key: ${key}`,
          original: `${key}: ${JSON.stringify(value)}`,
          action: 'Manually convert this option to Jest equivalent',
        }));
      }
    }

    converted.push('};');

    if (todos.length > 0) {
      converted.push('');
      for (const todo of todos) {
        converted.push(todo);
      }
    }

    return converted.join('\n') + '\n';
  }

  /**
   * Render a Playwright config file from parsed keys.
   * @param {Object} keys
   * @param {Object} keyMap
   * @param {string} sourceName
   * @returns {string}
   */
  _renderPlaywrightConfig(keys, keyMap, sourceName) {
    const converted = [];
    const todos = [];

    converted.push('import { defineConfig, devices } from \'@playwright/test\';');
    converted.push('');
    converted.push('export default defineConfig({');

    for (const [key, value] of Object.entries(keys)) {
      const mapper = keyMap[key];
      if (mapper) {
        const result = mapper(value, keys);
        if (result) {
          converted.push(`  ${result.key}: ${this._formatValue(result.value)},`);
        }
      } else {
        todos.push(this.formatter.formatTodo({
          id: 'CONFIG-UNSUPPORTED',
          description: `Unsupported ${sourceName} config key: ${key}`,
          original: `${key}: ${JSON.stringify(value)}`,
          action: 'Manually convert this option to Playwright equivalent',
        }));
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
   * Render a Cypress config file from parsed keys.
   * @param {Object} keys
   * @param {Object} keyMap
   * @param {string} sourceName
   * @returns {string}
   */
  _renderCypressConfig(keys, keyMap, sourceName) {
    const converted = [];
    const todos = [];

    converted.push('const { defineConfig } = require(\'cypress\');');
    converted.push('');
    converted.push('module.exports = defineConfig({');
    converted.push('  e2e: {');

    for (const [key, value] of Object.entries(keys)) {
      const mapper = keyMap[key];
      if (mapper) {
        const result = mapper(value, keys);
        if (result) {
          converted.push(`    ${result.key}: ${this._formatValue(result.value)},`);
        }
      } else {
        todos.push(this.formatter.formatTodo({
          id: 'CONFIG-UNSUPPORTED',
          description: `Unsupported ${sourceName} config key: ${key}`,
          original: `${key}: ${JSON.stringify(value)}`,
          action: 'Manually convert this option to Cypress equivalent',
        }));
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
   * Render a WebdriverIO config file from parsed keys.
   * @param {Object} keys
   * @param {Object} keyMap
   * @param {string} sourceName
   * @returns {string}
   */
  _renderWdioConfig(keys, keyMap, sourceName) {
    const converted = [];
    const todos = [];

    converted.push('exports.config = {');

    for (const [key, value] of Object.entries(keys)) {
      const mapper = keyMap[key];
      if (mapper) {
        const result = mapper(value, keys);
        if (result) {
          converted.push(`  ${result.key}: ${this._formatValue(result.value)},`);
        }
      } else {
        todos.push(this.formatter.formatTodo({
          id: 'CONFIG-UNSUPPORTED',
          description: `Unsupported ${sourceName} config key: ${key}`,
          original: `${key}: ${JSON.stringify(value)}`,
          action: 'Manually convert this option to WebdriverIO equivalent',
        }));
      }
    }

    converted.push('};');

    if (todos.length > 0) {
      converted.push('');
      for (const todo of todos) {
        converted.push(todo);
      }
    }

    return converted.join('\n') + '\n';
  }

  // --- Parsers ---

  /**
   * Extract config keys from a config file (best-effort parsing).
   *
   * @param {string} content
   * @returns {{keys: Object, raw: string}|null}
   */
  _extractConfigObject(content) {
    // Try to parse as a simple object literal
    // Match module.exports = { ... } or export default { ... } or defineConfig({ ... })
    // or exports.config = { ... } (WDIO)
    const patterns = [
      /module\.exports\s*=\s*\{([\s\S]*)\}/,
      /export\s+default\s+\{([\s\S]*)\}/,
      /defineConfig\s*\(\s*\{([\s\S]*)\}\s*\)/,
      /exports\.config\s*=\s*\{([\s\S]*)\}/,
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

    // Try JSON parse
    const jsonResult = this._parseJsonSimple(content);
    if (jsonResult) return jsonResult;

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
   * Try to parse content as JSON and return keys.
   * @param {string} content
   * @returns {{keys: Object, raw: string}|null}
   */
  _parseJsonSimple(content) {
    try {
      const trimmed = content.trim();
      if (!trimmed.startsWith('{')) return null;
      const obj = JSON.parse(trimmed);
      if (typeof obj !== 'object' || obj === null || Array.isArray(obj)) return null;
      return { keys: obj, raw: trimmed };
    } catch {
      return null;
    }
  }

  /**
   * Parse simple YAML content (.mocharc.yml style).
   * Only handles flat key: value pairs — no nested structures.
   * @param {string} content
   * @returns {{keys: Object, raw: string}|null}
   */
  _parseYamlSimple(content) {
    const trimmed = content.trim();
    // Quick check: YAML files typically don't start with { or module
    if (trimmed.startsWith('{') || trimmed.startsWith('module') || trimmed.startsWith('export')) {
      return null;
    }

    const keys = {};
    let foundAny = false;

    for (const line of trimmed.split('\n')) {
      const stripped = line.trim();
      if (!stripped || stripped.startsWith('#')) continue;

      const yamlMatch = stripped.match(/^(\w[\w-]*)\s*:\s*(.+)$/);
      if (yamlMatch) {
        const key = yamlMatch[1];
        let value = yamlMatch[2].trim();

        // Parse value type
        if (value === 'true') value = true;
        else if (value === 'false') value = false;
        else if (/^\d+$/.test(value)) value = Number(value);
        else if ((value.startsWith('"') && value.endsWith('"')) ||
                 (value.startsWith('\'') && value.endsWith('\''))) {
          value = value.slice(1, -1);
        }

        keys[key] = value;
        foundAny = true;
      }
    }

    return foundAny ? { keys, raw: trimmed } : null;
  }

  /**
   * Parse simple INI-style config (pytest.ini).
   * Looks for [pytest] section and extracts key = value pairs.
   * @param {string} content
   * @returns {{keys: Object, raw: string}|null}
   */
  _parseIniFile(content) {
    const trimmed = content.trim();
    const keys = {};
    let inPytestSection = false;
    let foundAny = false;

    for (const line of trimmed.split('\n')) {
      const stripped = line.trim();
      if (!stripped || stripped.startsWith('#') || stripped.startsWith(';')) continue;

      // Section header
      if (stripped.startsWith('[')) {
        inPytestSection = /^\[pytest\]$/i.test(stripped);
        continue;
      }

      if (inPytestSection) {
        const iniMatch = stripped.match(/^([\w-]+)\s*=\s*(.+)$/);
        if (iniMatch) {
          const key = iniMatch[1].replace(/-/g, '_');
          let value = iniMatch[2].trim();

          if (/^\d+$/.test(value)) value = Number(value);
          else if (value === 'true') value = true;
          else if (value === 'false') value = false;

          keys[key] = value;
          foundAny = true;
        }
      }
    }

    return foundAny ? { keys, raw: trimmed } : null;
  }

  /**
   * Parse TestNG XML to extract suite name and test classes.
   * @param {string} content
   * @returns {{suiteName: string, classes: string[]}|null}
   */
  _parseTestngXml(content) {
    if (!content.includes('<suite') && !content.includes('<test')) {
      return null;
    }

    const suiteMatch = content.match(/<suite\s[^>]*name\s*=\s*"([^"]+)"/);
    const suiteName = suiteMatch ? suiteMatch[1] : null;

    const classes = [];
    const classPattern = /<class\s[^>]*name\s*=\s*"([^"]+)"/g;
    let classMatch;
    while ((classMatch = classPattern.exec(content)) !== null) {
      classes.push(classMatch[1]);
    }

    if (!suiteName && classes.length === 0) return null;

    return { suiteName, classes };
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
   * Sanitize a string for use as a Java class name.
   * @param {string} name
   * @returns {string}
   */
  _sanitizeClassName(name) {
    return name.replace(/[^a-zA-Z0-9]/g, '');
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
