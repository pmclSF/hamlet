import { OutputValidator } from '../../src/core/OutputValidator.js';

describe('OutputValidator', () => {
  let validator;

  beforeEach(() => {
    validator = new OutputValidator();
  });

  describe('validate', () => {
    it('should pass valid Vitest output', () => {
      const output = `import { describe, it, expect } from 'vitest';\n\ndescribe('Math', () => {\n  it('adds', () => {\n    expect(1 + 1).toBe(2);\n  });\n});`;
      const { valid, issues } = validator.validate(output, 'vitest');

      expect(valid).toBe(true);
      expect(issues).toHaveLength(0);
    });

    it('should detect unbalanced brackets', () => {
      const output = `describe('test', () => {\n  it('works', () => {\n    expect(true).toBe(true);\n  });\n`;
      const { valid, issues } = validator.validate(output, 'vitest');

      expect(valid).toBe(false);
      expect(issues.some(i => i.type === 'bracket')).toBe(true);
    });

    it('should detect unbalanced parentheses', () => {
      const output = `expect(foo(.toBe(true);`;
      const { valid, issues } = validator.validate(output, 'vitest');

      expect(valid).toBe(false);
      expect(issues.some(i => i.type === 'bracket')).toBe(true);
    });

    it('should detect dangling jest references in vitest output', () => {
      const output = `import { describe, it, expect } from 'vitest';\n\ndescribe('test', () => {\n  it('works', () => {\n    jest.fn();\n  });\n});`;
      const { valid, issues } = validator.validate(output, 'vitest');

      expect(valid).toBe(false);
      expect(issues.some(i => i.type === 'dangling-reference')).toBe(true);
    });

    it('should detect dangling cy references in playwright output', () => {
      const output = `test('works', async ({ page }) => {\n  cy.visit('/');\n});`;
      const { valid, issues } = validator.validate(output, 'playwright');

      expect(valid).toBe(false);
      expect(issues.some(i => i.type === 'dangling-reference')).toBe(true);
    });

    it('should detect empty test bodies', () => {
      const output = `describe('test', () => {\n  it('should work', () => {});\n});`;
      const { valid, issues } = validator.validate(output, 'vitest');

      expect(valid).toBe(false);
      expect(issues.some(i => i.type === 'empty-test')).toBe(true);
    });

    it('should detect empty import source', () => {
      const output = `import { foo } from '';\n\ndescribe('test', () => {\n  it('works', () => {\n    expect(foo()).toBe(1);\n  });\n});`;
      const { valid, issues } = validator.validate(output, 'vitest');

      expect(valid).toBe(false);
      expect(issues.some(i => i.type === 'import')).toBe(true);
    });

    it('should handle empty output', () => {
      const { valid, issues } = validator.validate('', 'vitest');

      expect(valid).toBe(false);
      expect(issues.some(i => i.type === 'empty')).toBe(true);
    });

    it('should not flag dangling references in comments', () => {
      const output = `// Previously used jest.fn() here\nimport { describe, it, expect, vi } from 'vitest';\n\ndescribe('test', () => {\n  it('works', () => {\n    const fn = vi.fn();\n    expect(fn).toBeDefined();\n  });\n});`;
      const { issues } = validator.validate(output, 'vitest');

      expect(issues.filter(i => i.type === 'dangling-reference')).toHaveLength(0);
    });

    it('should pass valid Playwright output', () => {
      const output = `import { test, expect } from '@playwright/test';\n\ntest('navigation', async ({ page }) => {\n  await page.goto('/');\n  await expect(page).toHaveTitle('Home');\n});`;
      const { valid, issues } = validator.validate(output, 'playwright');

      expect(valid).toBe(true);
      expect(issues).toHaveLength(0);
    });

    it('should handle output with only whitespace', () => {
      const { valid, issues } = validator.validate('   \n\n  ', 'vitest');

      expect(valid).toBe(false);
      expect(issues.some(i => i.type === 'empty')).toBe(true);
    });

    it('should not flag brackets inside strings', () => {
      const output = `describe('test', () => {\n  it('checks regex', () => {\n    const re = '({[';\n    expect(re).toBeDefined();\n  });\n});`;
      const { valid, issues } = validator.validate(output, 'vitest');

      // Brackets in strings should not cause issues
      expect(issues.filter(i => i.type === 'bracket')).toHaveLength(0);
    });
  });

  describe('strict validation', () => {
    it('should pass valid JS when strictValidate is enabled', () => {
      const output = `const x = 1;\nconst y = x + 2;\n`;
      const { valid, issues } = validator.validate(output, 'vitest', {
        strictValidate: true,
      });

      expect(valid).toBe(true);
      expect(issues.filter((i) => i.type === 'syntax')).toHaveLength(0);
    });

    it('should detect syntax errors when strictValidate is enabled', () => {
      const output = `const x = ;\nfoo bar baz`;
      const { valid, issues } = validator.validate(output, 'vitest', {
        strictValidate: true,
      });

      expect(valid).toBe(false);
      expect(issues.some((i) => i.type === 'syntax')).toBe(true);
    });

    it('should not check syntax when strictValidate is not set', () => {
      const output = `const x = ;\nfoo bar baz`;
      const { issues } = validator.validate(output, 'vitest');

      // Without strict validate, no syntax issues reported
      expect(issues.filter((i) => i.type === 'syntax')).toHaveLength(0);
    });
  });
});
