import { FileClassifier } from '../../src/core/FileClassifier.js';

describe('FileClassifier', () => {
  let classifier;

  beforeEach(() => {
    classifier = new FileClassifier();
  });

  describe('classify', () => {
    // Standard file extension detection
    it('should classify .cy.js as test/cypress', () => {
      const result = classifier.classify('tests/login.cy.js', 'describe("Login", () => { it("works", () => { cy.visit("/"); }); });');
      expect(result.type).toBe('test');
      expect(result.framework).toBe('cypress');
    });

    it('should classify .spec.ts with Playwright imports as test/playwright', () => {
      const content = `import { test, expect } from '@playwright/test';\ntest('works', async ({ page }) => { await page.goto('/'); });`;
      const result = classifier.classify('tests/login.spec.ts', content);
      expect(result.type).toBe('test');
      expect(result.framework).toBe('playwright');
    });

    it('should classify .test.js with Jest patterns as test/jest', () => {
      const content = `describe('Math', () => { it('adds', () => { expect(1+1).toBe(2); }); });`;
      const result = classifier.classify('tests/math.test.js', content);
      expect(result.type).toBe('test');
    });

    it('should classify jest.config.js as config/jest', () => {
      const result = classifier.classify('jest.config.js', 'module.exports = { testEnvironment: "node" };');
      expect(result.type).toBe('config');
      expect(result.framework).toBe('jest');
    });

    it('should classify vitest.config.ts as config/vitest', () => {
      const result = classifier.classify('vitest.config.ts', 'export default defineConfig({})');
      expect(result.type).toBe('config');
      expect(result.framework).toBe('vitest');
    });

    it('should classify playwright.config.js as config/playwright', () => {
      const result = classifier.classify('playwright.config.js', 'module.exports = { use: { baseURL: "/" } }');
      expect(result.type).toBe('config');
      expect(result.framework).toBe('playwright');
    });

    it('should classify cypress.config.js as config/cypress', () => {
      const result = classifier.classify('cypress.config.js', 'module.exports = defineConfig({})');
      expect(result.type).toBe('config');
      expect(result.framework).toBe('cypress');
    });

    // Path pattern detection
    it('should classify file in helpers/ directory as helper', () => {
      const result = classifier.classify('test/helpers/setup-db.js', 'export function setupDb() { return db.connect(); }');
      expect(result.type).toBe('helper');
    });

    it('should classify file in fixtures/ directory as fixture', () => {
      const result = classifier.classify('test/fixtures/users.json', '{ "name": "John" }');
      expect(result.type).toBe('fixture');
    });

    it('should classify file in pages/ directory as page-object', () => {
      const result = classifier.classify('test/pages/login.js', 'export class LoginPage { get username() { return page.locator("#user"); } }');
      expect(result.type).toBe('page-object');
    });

    // Content wins over path when they conflict
    it('should classify file in helpers/ with test cases as test (content wins)', () => {
      const content = `import { helper } from './utils';\ndescribe('Helper tests', () => { it('should work', () => { expect(true).toBe(true); }); });`;
      const result = classifier.classify('test/helpers/my-helper.js', content);
      expect(result.type).toBe('test');
    });

    it('should classify file in tests/ with no test patterns as unknown', () => {
      const content = 'export const data = { name: "test data" };';
      const result = classifier.classify('tests/data.js', content);
      // No test patterns found in content, not a recognized path pattern
      expect(['fixture', 'helper', 'unknown']).toContain(result.type);
    });

    // Exports helpers AND contains tests → primary role is test
    it('should classify file that exports helpers but contains tests as test', () => {
      const content = `export function createUser() { return {}; }\ndescribe('createUser', () => { it('creates', () => { expect(createUser()).toBeDefined(); }); });`;
      const result = classifier.classify('utils/createUser.js', content);
      expect(result.type).toBe('test');
    });

    // Framework API calls but no test cases → page object
    it('should classify file with framework API but no tests as page-object', () => {
      const content = `export class DashboardPage {\n  get title() { return page.locator('h1'); }\n  async navigate() { await page.goto('/dashboard'); }\n}`;
      const result = classifier.classify('e2e/DashboardPage.js', content);
      expect(result.type).toBe('page-object');
    });

    // Setup files
    it('should classify jest.setup.js as setup', () => {
      const result = classifier.classify('jest.setup.js', 'global.fetch = require("node-fetch");');
      expect(result.type).toBe('setup');
    });

    it('should classify vitest.setup.ts as setup', () => {
      const result = classifier.classify('vitest.setup.ts', 'import "@testing-library/jest-dom";');
      expect(result.type).toBe('setup');
    });

    // Type definitions
    it('should classify .d.ts files as type-def', () => {
      const result = classifier.classify('src/types/global.d.ts', 'declare module "*.css" {}');
      expect(result.type).toBe('type-def');
    });

    // Edge cases
    it('should handle empty file', () => {
      const result = classifier.classify('empty.js', '');
      expect(result.type).toBe('unknown');
      expect(result.confidence).toBe(0);
    });

    it('should handle file with only whitespace', () => {
      const result = classifier.classify('blank.js', '   \n\n  ');
      expect(result.type).toBe('unknown');
    });

    it('should handle binary content', () => {
      const binary = Buffer.from([0x89, 0x50, 0x4E, 0x47, 0x00, 0x00]).toString();
      const result = classifier.classify('image.png', binary);
      expect(result.type).toBe('unknown');
    });

    it('should handle file with shebang', () => {
      const content = `#!/usr/bin/env node\ndescribe('CLI', () => { it('runs', () => { expect(true).toBe(true); }); });`;
      const result = classifier.classify('cli.test.js', content);
      expect(result.type).toBe('test');
    });

    // Python/Ruby specifics
    it('should classify conftest.py as setup', () => {
      const result = classifier.classify('tests/conftest.py', '@pytest.fixture\ndef db():\n    return connect()');
      expect(result.type).toBe('setup');
    });

    it('should classify spec_helper.rb as setup', () => {
      const result = classifier.classify('spec/spec_helper.rb', 'RSpec.configure do |config|\nend');
      expect(result.type).toBe('setup');
    });

    // Factory detection
    it('should classify factory file by path', () => {
      const result = classifier.classify('test/factories/user.js', 'export function buildUser() { return { name: "John" }; }');
      expect(result.type).toBe('factory');
    });

    it('should classify factory file by content', () => {
      const content = 'export function createMockUser(overrides = {}) { return { id: 1, ...overrides }; }';
      const result = classifier.classify('utils/data.js', content);
      expect(result.type).toBe('factory');
    });

    // Confidence scores
    it('should return higher confidence for content-matched tests', () => {
      const content = `describe('Login', () => { it('works', () => { expect(true).toBe(true); }); });`;
      const result = classifier.classify('tests/login.test.js', content);
      expect(result.confidence).toBeGreaterThanOrEqual(85);
    });

    it('should return lower confidence for path-only classification', () => {
      const content = 'export function formatDate(d) { return d.toISOString(); }';
      const result = classifier.classify('test/helpers/date.js', content);
      expect(result.confidence).toBeLessThan(90);
    });
  });
});
