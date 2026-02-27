/**
 * Classifies files by type, framework, and confidence.
 *
 * Types: test, helper, fixture, config, page-object, factory, type-def, setup, unknown
 * Content wins over path when they conflict.
 */

import { FrameworkDetector } from './FrameworkDetector.js';

const CONFIG_PATTERNS = [
  /jest\.config\.[jt]sx?$/,
  /vitest\.config\.[jt]sx?$/,
  /playwright\.config\.[jt]sx?$/,
  /cypress\.config\.[jt]sx?$/,
  /\.eslintrc/,
  /tsconfig.*\.json$/,
  /babel\.config/,
  /webpack\.config/,
  /rollup\.config/,
  /vite\.config\.[jt]sx?$/,
  /wdio\.conf\.[jt]sx?$/,
  /\.mocharc\.(yml|yaml|json|js|cjs)$/,
  /jasmine\.json$/,
  /testng\.xml$/,
  /pytest\.ini$/,
  /pyproject\.toml$/,
  /setup\.cfg$/,
];

const SETUP_PATTERNS = [
  /jest\.setup\.[jt]sx?$/,
  /vitest\.setup\.[jt]sx?$/,
  /setup\.[jt]sx?$/,
  /globalSetup\.[jt]sx?$/,
  /globalTeardown\.[jt]sx?$/,
  /conftest\.py$/,
  /spec_helper\.rb$/,
  /rails_helper\.rb$/,
  /test_helper\.rb$/,
];

const TYPE_DEF_PATTERNS = [/\.d\.ts$/, /\.d\.mts$/];

const TEST_FILE_PATTERNS = [
  /\.test\.[jt]sx?$/,
  /\.spec\.[jt]sx?$/,
  /\.cy\.[jt]sx?$/,
  /_test\.[jt]sx?$/,
  /_spec\.[jt]sx?$/,
  /test_.*\.[jt]sx?$/,
  /_test\.py$/,
  /test_.*\.py$/,
  /_spec\.rb$/,
];

const TEST_CONTENT_PATTERNS = [
  /\bdescribe\s*\(/,
  /\bit\s*\(/,
  /\btest\s*\(/,
  /\bexpect\s*\(/,
  /\bassert\s*[.(]/,
  /\btest\.describe\s*\(/,
];

const HELPER_PATH_PATTERNS = [
  /[/\\]helpers?[/\\]/i,
  /[/\\]utils?[/\\]/i,
  /[/\\]support[/\\]/i,
  /\.helper\.[jt]sx?$/,
  /\.helpers\.[jt]sx?$/,
  /\.util\.[jt]sx?$/,
  /\.utils\.[jt]sx?$/,
];

const FIXTURE_PATH_PATTERNS = [
  /[/\\]fixtures?[/\\]/i,
  /[/\\]__fixtures__[/\\]/i,
  /[/\\]test[-_]?data[/\\]/i,
  /[/\\]mocks?[/\\]/i,
  /[/\\]__mocks__[/\\]/i,
  /[/\\]stubs?[/\\]/i,
  /\.fixture\.[jt]sx?$/,
];

const PAGE_OBJECT_PATH_PATTERNS = [
  /[/\\]pages?[/\\]/i,
  /[/\\]page[-_]?objects?[/\\]/i,
  /[/\\]pom[/\\]/i,
  /\.page\.[jt]sx?$/,
  /\.po\.[jt]sx?$/,
];

const FACTORY_PATH_PATTERNS = [
  /[/\\]factor(y|ies)[/\\]/i,
  /\.factory\.[jt]sx?$/,
  /[/\\]builders?[/\\]/i,
];

const PAGE_OBJECT_CONTENT_PATTERNS = [
  /class\s+\w+Page\b/,
  /class\s+\w+PageObject\b/,
  /get\s+\w+\s*\(\)\s*\{[^}]*(?:page\.|locator\(|getBy|cy\.get|findElement)/,
];

const FACTORY_CONTENT_PATTERNS = [
  /function\s+(?:create|build|make)\w+/,
  /export\s+(?:function|const)\s+(?:create|build|make)\w+/,
  /class\s+\w+Factory\b/,
];

const CODE_EXTENSIONS = new Set([
  '.js',
  '.jsx',
  '.ts',
  '.tsx',
  '.mjs',
  '.mts',
  '.cjs',
  '.cts',
  '.py',
  '.java',
  '.kt',
  '.kts',
  '.groovy',
  '.scala',
  '.rb',
]);

export class FileClassifier {
  /**
   * Classify a file by its path and content.
   *
   * @param {string} filePath - File path (absolute or relative)
   * @param {string} content - File content
   * @returns {{type: string, framework: string|null, confidence: number}}
   */
  classify(filePath, content) {
    // Binary detection
    if (this._isBinary(content)) {
      return { type: 'unknown', framework: null, confidence: 0 };
    }

    // Empty file
    if (!content || content.trim().length === 0) {
      return { type: 'unknown', framework: null, confidence: 0 };
    }

    // Type definitions
    if (TYPE_DEF_PATTERNS.some((p) => p.test(filePath))) {
      return { type: 'type-def', framework: null, confidence: 95 };
    }

    // Config files — check path first
    if (CONFIG_PATTERNS.some((p) => p.test(filePath))) {
      const fw = this._detectFrameworkFromConfig(filePath);
      return { type: 'config', framework: fw, confidence: 95 };
    }

    // Setup files — check path patterns
    if (SETUP_PATTERNS.some((p) => p.test(filePath))) {
      const fw = this._detectFramework(content, filePath);
      return { type: 'setup', framework: fw, confidence: 90 };
    }

    // Only apply content-based detection to code files
    const ext = filePath.match(/\.[^./\\]+$/)?.[0]?.toLowerCase() || '';
    const isCode = CODE_EXTENSIONS.has(ext);

    // Content-based detection: does it have test patterns?
    const hasTests =
      isCode && TEST_CONTENT_PATTERNS.some((p) => p.test(content));
    const hasPageObjectContent =
      isCode && PAGE_OBJECT_CONTENT_PATTERNS.some((p) => p.test(content));
    const hasFactoryContent =
      isCode && FACTORY_CONTENT_PATTERNS.some((p) => p.test(content));

    // Content wins: if file has test cases, it's a test regardless of path
    if (hasTests) {
      const fw = this._detectFramework(content, filePath);
      return { type: 'test', framework: fw, confidence: 90 };
    }

    // Page object by content (no test patterns but has PO patterns)
    if (hasPageObjectContent) {
      const fw = this._detectFramework(content, filePath);
      return { type: 'page-object', framework: fw, confidence: 85 };
    }

    // Factory by content
    if (hasFactoryContent) {
      const fw = this._detectFramework(content, filePath);
      return { type: 'factory', framework: fw, confidence: 80 };
    }

    // Path-based classification for non-test content
    if (PAGE_OBJECT_PATH_PATTERNS.some((p) => p.test(filePath))) {
      const fw = this._detectFramework(content, filePath);
      return { type: 'page-object', framework: fw, confidence: 70 };
    }

    if (FACTORY_PATH_PATTERNS.some((p) => p.test(filePath))) {
      const fw = this._detectFramework(content, filePath);
      return { type: 'factory', framework: fw, confidence: 70 };
    }

    if (FIXTURE_PATH_PATTERNS.some((p) => p.test(filePath))) {
      return { type: 'fixture', framework: null, confidence: 75 };
    }

    if (HELPER_PATH_PATTERNS.some((p) => p.test(filePath))) {
      const fw = this._detectFramework(content, filePath);
      return { type: 'helper', framework: fw, confidence: 75 };
    }

    // Test file by path pattern (but no test content)
    if (TEST_FILE_PATTERNS.some((p) => p.test(filePath))) {
      const fw = this._detectFramework(content, filePath);
      return { type: 'test', framework: fw, confidence: 70 };
    }

    // Framework API calls but no test patterns → likely a helper/utility
    if (isCode) {
      const fw = this._detectFramework(content, filePath);
      if (fw) {
        return { type: 'helper', framework: fw, confidence: 60 };
      }
    }

    return { type: 'unknown', framework: null, confidence: 0 };
  }

  /**
   * @param {string} content
   * @returns {boolean}
   */
  _isBinary(content) {
    if (!content) return false;
    // Check for null bytes or high ratio of non-printable characters
    const sample = content.slice(0, 1024);
    let nonPrintable = 0;
    for (let i = 0; i < sample.length; i++) {
      const code = sample.charCodeAt(i);
      if (code === 0) return true;
      if (code < 32 && code !== 9 && code !== 10 && code !== 13) {
        nonPrintable++;
      }
    }
    return nonPrintable / sample.length > 0.1;
  }

  /**
   * @param {string} content
   * @param {string} filePath
   * @returns {string|null}
   */
  _detectFramework(content, filePath) {
    try {
      const result = FrameworkDetector.detect(content, filePath);
      if (result.framework && result.confidence > 0.3) {
        return result.framework;
      }
    } catch {
      // FrameworkDetector may not recognize the content
    }
    // Check for Jest/Vitest patterns not in FrameworkDetector
    if (
      /\bjest\b/i.test(content) ||
      /jest\.fn\b/.test(content) ||
      /jest\.mock\b/.test(content)
    ) {
      return 'jest';
    }
    if (/\bvi\.\w+/.test(content) || /from\s+['"]vitest['"]/.test(content)) {
      return 'vitest';
    }
    return null;
  }

  /**
   * @param {string} filePath
   * @returns {string|null}
   */
  _detectFrameworkFromConfig(filePath) {
    if (/jest\.config/i.test(filePath)) return 'jest';
    if (/vitest\.config/i.test(filePath)) return 'vitest';
    if (/playwright\.config/i.test(filePath)) return 'playwright';
    if (/cypress\.config/i.test(filePath)) return 'cypress';
    if (/wdio\.conf/i.test(filePath)) return 'webdriverio';
    if (/\.mocharc/i.test(filePath)) return 'mocha';
    if (/jasmine\.json/i.test(filePath)) return 'jasmine';
    if (/testng\.xml/i.test(filePath)) return 'testng';
    if (/pytest\.ini/i.test(filePath)) return 'pytest';
    if (/pyproject\.toml/i.test(filePath)) return 'pytest';
    return null;
  }
}
