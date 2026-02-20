import fs from 'fs/promises';
import path from 'path';
import chalk from 'chalk';
import { chromium } from '@playwright/test';

/**
 * Validates converted Playwright tests for correctness and functionality
 */
export class TestValidator {
  constructor() {
    this.results = {
      passed: [],
      failed: [],
      skipped: [],
      errors: [],
    };

    this.validationRules = {
      syntaxCheck: this.checkSyntax.bind(this),
      importValidation: this.validateImports.bind(this),
      assertionCheck: this.checkAssertions.bind(this),
      asyncAwaitUsage: this.checkAsyncAwait.bind(this),
      selectorValidation: this.validateSelectors.bind(this),
      pageObjectUsage: this.checkPageObjects.bind(this),
      fixtureHandling: this.checkFixtures.bind(this),
      hookImplementation: this.checkHooks.bind(this),
    };
  }

  /**
   * Validate converted tests in a directory
   * @param {string} testDir - Directory containing converted tests
   * @returns {Promise<Object>} - Validation results
   */
  async validateConvertedTests(testDir) {
    try {
      console.log(chalk.blue('\nStarting test validation...'));

      // Find all test files
      const testFiles = await this.findTestFiles(testDir);

      // Run validations
      for (const file of testFiles) {
        await this.validateSingleTest(file);
      }

      // Generate report
      const report = this.generateValidationReport();

      // Log summary
      this.logValidationSummary();

      return report;
    } catch (error) {
      console.error(chalk.red('Error during test validation:'), error);
      throw error;
    }
  }

  /**
   * Find all test files in directory
   * @param {string} dir - Directory to search
   * @returns {Promise<string[]>} - Array of test file paths
   */
  async findTestFiles(dir) {
    const testFiles = [];

    async function scan(directory) {
      const entries = await fs.readdir(directory, { withFileTypes: true });

      for (const entry of entries) {
        const fullPath = path.join(directory, entry.name);

        if (entry.isDirectory()) {
          await scan(fullPath);
        } else if (
          entry.isFile() &&
          /\.(spec|test)\.(js|ts)$/.test(entry.name)
        ) {
          testFiles.push(fullPath);
        }
      }
    }

    await scan(dir);
    return testFiles;
  }

  /**
   * Validate a single test file
   * @param {string} testFile - Path to test file
   */
  async validateSingleTest(testFile) {
    try {
      console.log(chalk.blue(`\nValidating ${path.basename(testFile)}...`));

      const content = await fs.readFile(testFile, 'utf8');
      const validationResults = [];

      // Run all validation rules
      for (const [ruleName, validator] of Object.entries(
        this.validationRules
      )) {
        try {
          const result = await validator(content, testFile);
          validationResults.push({
            rule: ruleName,
            ...result,
          });
        } catch (error) {
          validationResults.push({
            rule: ruleName,
            status: 'error',
            message: error.message,
          });
        }
      }

      // Check if test can be executed
      const executionResult = await this.executeTest(testFile);
      validationResults.push(executionResult);

      // Process validation results
      this.processValidationResults(testFile, validationResults);
    } catch (error) {
      this.results.errors.push({
        file: testFile,
        error: error.message,
      });
      console.error(
        chalk.red(`✗ Validation failed for ${path.basename(testFile)}:`),
        error
      );
    }
  }

  /**
   * Check test file syntax
   * @param {string} content - Test file content
   * @returns {Object} - Validation result
   */
  async checkSyntax(content) {
    try {
      // Try to parse the content as a module
      new Function(content);
      return { status: 'passed' };
    } catch (error) {
      return {
        status: 'failed',
        message: `Syntax error: ${error.message}`,
      };
    }
  }

  /**
   * Validate test imports
   * @param {string} content - Test file content
   * @returns {Object} - Validation result
   */
  async validateImports(content) {
    const requiredImports = ['@playwright/test', 'expect'];

    const missingImports = requiredImports.filter(
      (imp) =>
        !content.includes(`from '${imp}'`) &&
        !content.includes(`require('${imp}')`)
    );

    return {
      status: missingImports.length === 0 ? 'passed' : 'failed',
      message:
        missingImports.length > 0
          ? `Missing required imports: ${missingImports.join(', ')}`
          : null,
    };
  }

  /**
   * Check test assertions
   * @param {string} content - Test file content
   * @returns {Object} - Validation result
   */
  async checkAssertions(content) {
    const assertionPattern = /expect\s*\(.*?\)\s*\.\s*to[A-Z]\w+\s*\(/g;
    const assertions = content.match(assertionPattern);

    return {
      status: assertions ? 'passed' : 'failed',
      message: assertions ? null : 'No assertions found in test',
    };
  }

  /**
   * Check async/await usage
   * @param {string} content - Test file content
   * @returns {Object} - Validation result
   */
  async checkAsyncAwait(content) {
    const asyncTestPattern = /test\s*\(\s*(['"`].*?['"`])\s*,\s*async/g;
    const awaitPattern = /await\s+\w+\s*\./g;

    const hasAsyncTests = asyncTestPattern.test(content);
    const hasAwait = awaitPattern.test(content);

    return {
      status: hasAsyncTests && hasAwait ? 'passed' : 'failed',
      message: !hasAsyncTests
        ? 'Missing async test functions'
        : !hasAwait
          ? 'Missing await keywords'
          : null,
    };
  }

  /**
   * Validate selectors
   * @param {string} content - Test file content
   * @returns {Object} - Validation result
   */
  async validateSelectors(content) {
    // Use non-greedy matching with explicit character exclusion to prevent ReDoS
    const selectorPattern = /locator\s*\(\s*['"`]([^'"`\n]*)['"`]\s*\)/g;
    const selectors = Array.from(
      content.matchAll(selectorPattern),
      (m) => m[1]
    );

    const invalidSelectors = selectors.filter((selector) => {
      // Check for common issues in selectors
      return (
        selector.includes('>>') || // Old Cypress chain syntax
        /^[.#][0-9]/.test(selector) || // Invalid CSS selector
        selector.includes('cypress-') // Cypress-specific attributes
      );
    });

    return {
      status: invalidSelectors.length === 0 ? 'passed' : 'failed',
      message:
        invalidSelectors.length > 0
          ? `Invalid selectors found: ${invalidSelectors.join(', ')}`
          : null,
    };
  }

  /**
   * Check page object usage
   * @param {string} content - Test file content
   * @returns {Object} - Validation result
   */
  async checkPageObjects(content) {
    // This is optional, so we just check for proper usage if present
    const pageObjectPattern = /class\s+\w+Page\s*{/;
    const hasPageObjects = pageObjectPattern.test(content);

    if (!hasPageObjects) {
      return { status: 'skipped' };
    }

    const properUsage = content.includes('constructor(page)');
    return {
      status: properUsage ? 'passed' : 'failed',
      message: !properUsage
        ? 'Page objects should accept page in constructor'
        : null,
    };
  }

  /**
   * Check fixture handling
   * @param {string} content - Test file content
   * @returns {Object} - Validation result
   */
  async checkFixtures(content) {
    const fixturePattern =
      /test\s*\(\s*(['"`].*?['"`])\s*,\s*async\s*\(\s*{\s*\w+\s*}\s*\)/;
    const hasFixtures = fixturePattern.test(content);

    if (!hasFixtures) {
      return { status: 'skipped' };
    }

    const properUsage = content.includes('test.use(');
    return {
      status: properUsage ? 'passed' : 'failed',
      message: !properUsage
        ? 'Fixtures should be properly configured with test.use'
        : null,
    };
  }

  /**
   * Check hook implementation
   * @param {string} content - Test file content
   * @returns {Object} - Validation result
   */
  async checkHooks(content) {
    const hookPattern =
      /test\.(beforeAll|afterAll|beforeEach|afterEach)\s*\(\s*async/g;
    const hooks = content.match(hookPattern);

    if (!hooks) {
      return { status: 'skipped' };
    }

    const properUsage = hooks.every(
      (hook) =>
        content.includes(`${hook.split('.')[1]}(async`) &&
        content.includes('await')
    );

    return {
      status: properUsage ? 'passed' : 'failed',
      message: !properUsage
        ? 'Hooks should be async functions with proper await usage'
        : null,
    };
  }

  /**
   * Execute test file
   * @param {string} testFile - Path to test file
   * @returns {Object} - Execution result
   */
  async executeTest(testFile) {
    try {
      const browser = await chromium.launch();
      const context = await browser.newContext();
      const page = await context.newPage();

      // Load and execute test
      const testModule = await import(path.resolve(testFile));
      await testModule.default({ page });

      await browser.close();
      return { status: 'passed', rule: 'execution' };
    } catch (error) {
      return {
        status: 'failed',
        rule: 'execution',
        message: `Test execution failed: ${error.message}`,
      };
    }
  }

  /**
   * Process validation results
   * @param {string} testFile - Test file path
   * @param {Object[]} results - Validation results
   */
  processValidationResults(testFile, results) {
    const failed = results.filter((r) => r.status === 'failed');
    const skipped = results.filter((r) => r.status === 'skipped');

    if (failed.length === 0) {
      this.results.passed.push({
        file: testFile,
        results: results,
      });
      console.log(
        chalk.green(`✓ ${path.basename(testFile)} passed validation`)
      );
    } else {
      this.results.failed.push({
        file: testFile,
        failures: failed,
      });
      console.log(chalk.red(`✗ ${path.basename(testFile)} failed validation`));
      failed.forEach((failure) => {
        console.log(chalk.yellow(`  - ${failure.rule}: ${failure.message}`));
      });
    }

    if (skipped.length > 0) {
      this.results.skipped.push({
        file: testFile,
        skipped: skipped.map((s) => s.rule),
      });
    }
  }

  /**
   * Generate validation report
   * @returns {Object} - Validation report
   */
  generateValidationReport() {
    return {
      summary: {
        total: this.results.passed.length + this.results.failed.length,
        passed: this.results.passed.length,
        failed: this.results.failed.length,
        skipped: this.results.skipped.length,
        errors: this.results.errors.length,
      },
      details: {
        passed: this.results.passed,
        failed: this.results.failed,
        skipped: this.results.skipped,
        errors: this.results.errors,
      },
      timestamp: new Date().toISOString(),
    };
  }

  /**
   * Log validation summary
   */
  logValidationSummary() {
    const { summary } = this.generateValidationReport();

    console.log('\nValidation Summary:');
    console.log(chalk.green(`Passed: ${summary.passed}`));
    console.log(chalk.red(`Failed: ${summary.failed}`));
    console.log(chalk.yellow(`Skipped: ${summary.skipped}`));
    console.log(chalk.red(`Errors: ${summary.errors}`));
  }

  /**
   * Get validation results
   * @returns {Object} - Validation results
   */
  getResults() {
    return this.results;
  }
}
