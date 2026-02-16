import fs from "fs/promises";
import { logUtils } from "../utils/helpers.js";

const logger = logUtils.createLogger("MetadataCollector");

/**
 * Collects and manages test metadata
 */
export class TestMetadataCollector {
  constructor() {
    this.metadata = new Map();
    this.testSuites = new Map();
    this.tags = new Set();
  }

  /**
   * Collect metadata from test file
   * @param {string} testPath - Path to test file
   * @returns {Promise<Object>} - Test metadata
   */
  async collectMetadata(testPath) {
    try {
      const content = await fs.readFile(testPath, "utf8");

      const metadata = {
        path: testPath,
        type: this.detectTestType(content),
        suites: this.extractTestSuites(content),
        cases: this.extractTestCases(content),
        tags: this.extractTags(content),
        complexity: this.calculateComplexity(content),
        coverage: this.extractCoverage(content),
        lastModified: await this.getLastModified(testPath),
      };

      this.metadata.set(testPath, metadata);
      return metadata;
    } catch (error) {
      logger.error(`Failed to collect metadata for ${testPath}:`, error);
      throw error;
    }
  }

  /**
   * Detect test type
   * @param {string} content - Test content
   * @returns {string} - Test type
   */
  detectTestType(content) {
    const patterns = {
      e2e: /cy\.visit|cy\.go|cy\.reload/i,
      component: /cy\.mount|mount\(/i,
      api: /cy\.request|cy\.intercept/i,
      visual: /cy\.screenshot|matchImageSnapshot/i,
      performance: /cy\.lighthouse|performance\./i,
      accessibility: /cy\.injectAxe|cy\.checkA11y/i,
    };

    for (const [type, pattern] of Object.entries(patterns)) {
      if (pattern.test(content)) {
        return type;
      }
    }

    return "unknown";
  }

  /**
   * Extract test suites
   * @param {string} content - Test content
   * @returns {Object[]} - Array of test suites
   */
  extractTestSuites(content) {
    const suites = [];
    const suiteRegex = /describe\(['"](.*?)['"],/g;

    let match;
    while ((match = suiteRegex.exec(content)) !== null) {
      suites.push({
        name: match[1],
        location: match.index,
      });
    }

    return suites;
  }

  /**
   * Extract test cases
   * @param {string} content - Test content
   * @returns {Object[]} - Array of test cases
   */
  extractTestCases(content) {
    const cases = [];
    const caseRegex = /it\(['"](.*?)['"],/g;

    let match;
    while ((match = caseRegex.exec(content)) !== null) {
      cases.push({
        name: match[1],
        location: match.index,
      });
    }

    return cases;
  }

  /**
   * Extract test tags
   * @param {string} content - Test content
   * @returns {string[]} - Array of tags
   */
  extractTags(content) {
    const tags = new Set();
    const tagRegex =
      /@(smoke|regression|e2e|api|visual|performance|accessibility)/g;

    let match;
    while ((match = tagRegex.exec(content)) !== null) {
      tags.add(match[1]);
    }

    return Array.from(tags);
  }

  /**
   * Calculate test complexity
   * @param {string} content - Test content
   * @returns {Object} - Complexity metrics
   */
  calculateComplexity(content) {
    return {
      assertions: (content.match(/expect|should|assert/g) || []).length,
      commands: (content.match(/cy\./g) || []).length,
      conditionals: (content.match(/if|else|switch|case/g) || []).length,
      hooks: (content.match(/before|after|beforeEach|afterEach/g) || []).length,
    };
  }

  /**
   * Extract coverage information
   * @param {string} content - Test content
   * @returns {Object} - Coverage information
   */
  extractCoverage(content) {
    return {
      selectors: this.extractSelectors(content),
      routes: this.extractRoutes(content),
      assertions: this.extractAssertions(content),
      interactions: this.extractInteractions(content),
    };
  }

  /**
   * Extract selectors used in test
   * @param {string} content - Test content
   * @returns {string[]} - Array of selectors
   */
  extractSelectors(content) {
    const selectors = new Set();
    const selectorRegex = /cy\.(?:get|find|contains)\(['"](.*?)['"]\)/g;

    let match;
    while ((match = selectorRegex.exec(content)) !== null) {
      selectors.add(match[1]);
    }

    return Array.from(selectors);
  }

  /**
   * Extract routes tested
   * @param {string} content - Test content
   * @returns {string[]} - Array of routes
   */
  extractRoutes(content) {
    const routes = new Set();
    const routeRegex = /cy\.(?:visit|request|intercept)\(['"](.*?)['"]\)/g;

    let match;
    while ((match = routeRegex.exec(content)) !== null) {
      routes.add(match[1]);
    }

    return Array.from(routes);
  }

  /**
   * Extract assertions used
   * @param {string} content - Test content
   * @returns {Object[]} - Array of assertions
   */
  extractAssertions(content) {
    const assertions = [];
    const assertionRegex = /(?:expect|should|assert).*?['"](.*?)['"]/g;

    let match;
    while ((match = assertionRegex.exec(content)) !== null) {
      assertions.push({
        type: this.getAssertionType(match[0]),
        value: match[1],
      });
    }

    return assertions;
  }

  /**
   * Get assertion type
   * @param {string} assertion - Assertion string
   * @returns {string} - Assertion type
   */
  getAssertionType(assertion) {
    if (assertion.includes("exist")) return "existence";
    if (assertion.includes("visible")) return "visibility";
    if (assertion.includes("have.text")) return "text";
    if (assertion.includes("have.value")) return "value";
    if (assertion.includes("have.class")) return "class";
    if (assertion.includes("have.attr")) return "attribute";
    return "other";
  }

  /**
   * Extract user interactions
   * @param {string} content - Test content
   * @returns {Object[]} - Array of interactions
   */
  extractInteractions(content) {
    const interactions = [];
    const interactionRegex = /cy\.(?:click|type|select|check|uncheck|hover)\(/g;

    let match;
    while ((match = interactionRegex.exec(content)) !== null) {
      interactions.push({
        type: match[0].replace(/cy\.|[()]/g, ""),
        location: match.index,
      });
    }

    return interactions;
  }

  /**
   * Get last modified time of test file
   * @param {string} testPath - Path to test file
   * @returns {Promise<string>} - Last modified timestamp
   */
  async getLastModified(testPath) {
    const stats = await fs.stat(testPath);
    return stats.mtime.toISOString();
  }

  /**
   * Get metadata for test file
   * @param {string} testPath - Path to test file
   * @returns {Object|null} - Test metadata
   */
  getMetadata(testPath) {
    return this.metadata.get(testPath) || null;
  }

  /**
   * Get all collected tags
   * @returns {string[]} - Array of unique tags
   */
  getAllTags() {
    return Array.from(this.tags);
  }

  /**
   * Get tests by tag
   * @param {string} tag - Tag to filter by
   * @returns {Object[]} - Array of matching tests
   */
  getTestsByTag(tag) {
    return Array.from(this.metadata.values()).filter((meta) =>
      meta.tags.includes(tag),
    );
  }

  /**
   * Generate metadata report
   * @returns {Object} - Metadata report
   */
  generateReport() {
    const tests = Array.from(this.metadata.values());

    return {
      summary: {
        totalTests: tests.length,
        types: this.summarizeTypes(tests),
        tags: this.summarizeTags(tests),
        complexity: this.summarizeComplexity(tests),
      },
      tests: tests.map((test) => ({
        path: test.path,
        type: test.type,
        tags: test.tags,
        suites: test.suites.length,
        cases: test.cases.length,
        complexity: test.complexity,
      })),
    };
  }

  /**
   * Summarize test types
   * @param {Object[]} tests - Array of test metadata
   * @returns {Object} - Type summary
   */
  summarizeTypes(tests) {
    return tests.reduce((acc, test) => {
      acc[test.type] = (acc[test.type] || 0) + 1;
      return acc;
    }, {});
  }

  /**
   * Summarize test tags
   * @param {Object[]} tests - Array of test metadata
   * @returns {Object} - Tag summary
   */
  summarizeTags(tests) {
    return tests.reduce((acc, test) => {
      test.tags.forEach((tag) => {
        acc[tag] = (acc[tag] || 0) + 1;
      });
      return acc;
    }, {});
  }

  /**
   * Summarize test complexity
   * @param {Object[]} tests - Array of test metadata
   * @returns {Object} - Complexity summary
   */
  summarizeComplexity(tests) {
    const complexities = tests.map((t) => t.complexity);

    return {
      averageAssertions: this.average(complexities.map((c) => c.assertions)),
      averageCommands: this.average(complexities.map((c) => c.commands)),
      averageConditionals: this.average(
        complexities.map((c) => c.conditionals),
      ),
      averageHooks: this.average(complexities.map((c) => c.hooks)),
    };
  }

  /**
   * Calculate average
   * @param {number[]} numbers - Array of numbers
   * @returns {number} - Average value
   */
  average(numbers) {
    return numbers.length
      ? numbers.reduce((a, b) => a + b, 0) / numbers.length
      : 0;
  }
}
