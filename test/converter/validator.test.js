import { TestValidator } from '../../src/converter/validator.js';

describe('TestValidator', () => {
  let validator;

  beforeEach(() => {
    validator = new TestValidator();
  });

  describe('constructor', () => {
    it('should initialize with empty results', () => {
      expect(validator.results.passed).toEqual([]);
      expect(validator.results.failed).toEqual([]);
      expect(validator.results.skipped).toEqual([]);
      expect(validator.results.errors).toEqual([]);
    });

    it('should define validation rules', () => {
      expect(Object.keys(validator.validationRules)).toContain('syntaxCheck');
      expect(Object.keys(validator.validationRules)).toContain('importValidation');
      expect(Object.keys(validator.validationRules)).toContain('assertionCheck');
      expect(Object.keys(validator.validationRules)).toContain('asyncAwaitUsage');
      expect(Object.keys(validator.validationRules)).toContain('selectorValidation');
    });
  });

  describe('checkSyntax', () => {
    it('should pass for valid JavaScript', async () => {
      const result = await validator.checkSyntax('const x = 1; const y = 2;');
      expect(result.status).toBe('passed');
    });

    it('should fail for invalid JavaScript', async () => {
      const result = await validator.checkSyntax('const x = {;');
      expect(result.status).toBe('failed');
      expect(result.message).toContain('Syntax error');
    });
  });

  describe('validateImports', () => {
    it('should pass when required imports are present', async () => {
      const content = `
        import { test, expect } from '@playwright/test';
        import { expect } from 'expect';
      `;
      const result = await validator.validateImports(content);
      expect(result.status).toBe('passed');
    });

    it('should fail when imports are missing', async () => {
      const result = await validator.validateImports('const x = 1;');
      expect(result.status).toBe('failed');
      expect(result.message).toContain('Missing required imports');
    });
  });

  describe('checkAssertions', () => {
    it('should pass when assertions are present', async () => {
      const content = 'expect(page.locator(".btn")).toBeVisible();';
      const result = await validator.checkAssertions(content);
      expect(result.status).toBe('passed');
    });

    it('should fail when no assertions found', async () => {
      const result = await validator.checkAssertions('const x = 1;');
      expect(result.status).toBe('failed');
      expect(result.message).toContain('No assertions found');
    });
  });

  describe('checkAsyncAwait', () => {
    it('should pass for proper async/await usage', async () => {
      const content = `
        test('my test', async ({ page }) => {
          await page.goto('/');
        });
      `;
      const result = await validator.checkAsyncAwait(content);
      expect(result.status).toBe('passed');
    });

    it('should fail when async keyword is missing', async () => {
      const content = `
        test('my test', () => {
          page.goto('/');
        });
      `;
      const result = await validator.checkAsyncAwait(content);
      expect(result.status).toBe('failed');
    });
  });

  describe('validateSelectors', () => {
    it('should pass for valid selectors', async () => {
      const content = "locator('.valid-selector')";
      const result = await validator.validateSelectors(content);
      expect(result.status).toBe('passed');
    });

    it('should fail for selectors with old Cypress chain syntax', async () => {
      const content = "locator('.parent >> .child')";
      const result = await validator.validateSelectors(content);
      expect(result.status).toBe('failed');
      expect(result.message).toContain('Invalid selectors');
    });

    it('should fail for cypress-specific attributes', async () => {
      const content = "locator('[cypress-data]')";
      const result = await validator.validateSelectors(content);
      expect(result.status).toBe('failed');
    });

    it('should pass when no selectors present', async () => {
      const result = await validator.validateSelectors('const x = 1;');
      expect(result.status).toBe('passed');
    });
  });

  describe('checkPageObjects', () => {
    it('should skip when no page objects present', async () => {
      const result = await validator.checkPageObjects('const x = 1;');
      expect(result.status).toBe('skipped');
    });

    it('should pass for proper page object usage', async () => {
      const content = `
        class LoginPage {
          constructor(page) {
            this.page = page;
          }
        }
      `;
      const result = await validator.checkPageObjects(content);
      expect(result.status).toBe('passed');
    });

    it('should fail for page objects without page in constructor', async () => {
      const content = `
        class LoginPage {
          constructor() {
            this.url = '/login';
          }
        }
      `;
      const result = await validator.checkPageObjects(content);
      expect(result.status).toBe('failed');
    });
  });

  describe('checkFixtures', () => {
    it('should skip when no fixtures present', async () => {
      const result = await validator.checkFixtures('const x = 1;');
      expect(result.status).toBe('skipped');
    });
  });

  describe('checkHooks', () => {
    it('should skip when no hooks present', async () => {
      const result = await validator.checkHooks('const x = 1;');
      expect(result.status).toBe('skipped');
    });
  });

  describe('processValidationResults', () => {
    it('should add to passed when no failures', () => {
      const results = [
        { status: 'passed', rule: 'syntax' },
        { status: 'passed', rule: 'imports' }
      ];

      validator.processValidationResults('test.spec.js', results);
      expect(validator.results.passed).toHaveLength(1);
      expect(validator.results.failed).toHaveLength(0);
    });

    it('should add to failed when failures exist', () => {
      const results = [
        { status: 'passed', rule: 'syntax' },
        { status: 'failed', rule: 'imports', message: 'Missing imports' }
      ];

      validator.processValidationResults('test.spec.js', results);
      expect(validator.results.failed).toHaveLength(1);
      expect(validator.results.passed).toHaveLength(0);
    });

    it('should track skipped validations', () => {
      const results = [
        { status: 'skipped', rule: 'fixtures' }
      ];

      validator.processValidationResults('test.spec.js', results);
      expect(validator.results.skipped).toHaveLength(1);
    });
  });

  describe('generateValidationReport', () => {
    it('should generate report with correct summary', () => {
      validator.results.passed.push({ file: 'a.js' });
      validator.results.failed.push({ file: 'b.js' });

      const report = validator.generateValidationReport();
      expect(report.summary.total).toBe(2);
      expect(report.summary.passed).toBe(1);
      expect(report.summary.failed).toBe(1);
      expect(report.timestamp).toBeDefined();
    });
  });

  describe('getResults', () => {
    it('should return current results', () => {
      validator.results.passed.push({ file: 'test.js' });
      const results = validator.getResults();
      expect(results.passed).toHaveLength(1);
    });
  });
});
