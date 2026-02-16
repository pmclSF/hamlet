import fs from 'fs/promises';
import path from 'path';
import os from 'os';
import { MigrationEstimator } from '../../src/core/MigrationEstimator.js';

describe('MigrationEstimator', () => {
  let estimator;
  let tmpDir;

  beforeEach(async () => {
    estimator = new MigrationEstimator();
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-estimate-'));
  });

  afterEach(async () => {
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  describe('estimate', () => {
    it('should estimate a simple project (all high confidence)', async () => {
      await fs.writeFile(
        path.join(tmpDir, 'simple.test.js'),
        `describe('simple', () => { it('works', () => { expect(1).toBe(1); }); });`
      );

      const result = await estimator.estimate(tmpDir, { from: 'jest', to: 'vitest' });

      expect(result.summary.totalFiles).toBe(1);
      expect(result.summary.predictedHigh).toBeGreaterThanOrEqual(1);
      expect(result.from).toBe('jest');
      expect(result.to).toBe('vitest');
    });

    it('should estimate project with complex mocks (some medium/low)', async () => {
      await fs.writeFile(
        path.join(tmpDir, 'mocked.test.js'),
        `jest.mock('./module');\njest.spyOn(obj, 'method');\njest.mock('./another');\njest.mock('./third');\njest.mock('./fourth');\n\ndescribe('mocked', () => { it('works', () => { expect(1).toBe(1); }); });`
      );

      const result = await estimator.estimate(tmpDir, { from: 'jest', to: 'vitest' });

      expect(result.summary.totalFiles).toBe(1);
      // Should detect high-complexity patterns
      expect(result.files[0].predictedConfidence).toBeLessThan(90);
    });

    it('should identify top blockers', async () => {
      await fs.writeFile(
        path.join(tmpDir, 'complex.test.js'),
        `jest.mock('./a');\njest.mock('./b');\njest.spyOn(x, 'y');\n\ndescribe('x', () => { it('w', () => { expect(1).toBe(1); }); });`
      );

      const result = await estimator.estimate(tmpDir, { from: 'jest', to: 'vitest' });

      expect(result.blockers).toBeDefined();
      expect(Array.isArray(result.blockers)).toBe(true);
    });

    it('should produce effort estimate', async () => {
      await fs.writeFile(
        path.join(tmpDir, 'test.test.js'),
        `describe('test', () => { it('works', () => { expect(1).toBe(1); }); });`
      );

      const result = await estimator.estimate(tmpDir, { from: 'jest', to: 'vitest' });

      expect(result.estimatedEffort).toBeDefined();
      expect(result.estimatedEffort.description).toBeDefined();
      expect(typeof result.estimatedEffort.estimatedManualMinutes).toBe('number');
    });

    it('should return structured data (not formatted string)', async () => {
      await fs.writeFile(
        path.join(tmpDir, 'test.test.js'),
        `describe('test', () => { it('works', () => { expect(1).toBe(1); }); });`
      );

      const result = await estimator.estimate(tmpDir, { from: 'jest', to: 'vitest' });

      expect(typeof result).toBe('object');
      expect(result.summary).toBeDefined();
      expect(result.files).toBeDefined();
      expect(Array.isArray(result.files)).toBe(true);
    });

    it('should handle empty project', async () => {
      const result = await estimator.estimate(tmpDir, { from: 'jest', to: 'vitest' });

      expect(result.summary.totalFiles).toBe(0);
      expect(result.files).toHaveLength(0);
    });

    it('should handle project with no test files', async () => {
      await fs.writeFile(path.join(tmpDir, 'readme.md'), '# Hello');
      await fs.writeFile(path.join(tmpDir, 'config.json'), '{}');

      const result = await estimator.estimate(tmpDir, { from: 'jest', to: 'vitest' });

      expect(result.summary.totalFiles).toBe(0); // No .js/.ts files
    });

    it('should NOT modify files or create .hamlet/', async () => {
      await fs.writeFile(
        path.join(tmpDir, 'test.test.js'),
        `describe('test', () => { it('works', () => { expect(1).toBe(1); }); });`
      );

      await estimator.estimate(tmpDir, { from: 'jest', to: 'vitest' });

      const hamletExists = await fs.access(path.join(tmpDir, '.hamlet'))
        .then(() => true).catch(() => false);
      expect(hamletExists).toBe(false);
    });

    it('should include file type classification in results', async () => {
      await fs.writeFile(
        path.join(tmpDir, 'app.test.js'),
        `describe('app', () => { it('works', () => { expect(1).toBe(1); }); });`
      );

      const result = await estimator.estimate(tmpDir, { from: 'jest', to: 'vitest' });

      expect(result.files[0].type).toBeDefined();
    });

    it('should count circular dependencies', async () => {
      await fs.writeFile(
        path.join(tmpDir, 'a.js'),
        `import { b } from './b.js';\nexport const a = 1;`
      );
      await fs.writeFile(
        path.join(tmpDir, 'b.js'),
        `import { a } from './a.js';\nexport const b = 2;`
      );

      const result = await estimator.estimate(tmpDir, { from: 'jest', to: 'vitest' });

      expect(result.summary.circularDependencies).toBeGreaterThan(0);
    });

    it('should handle Cypress estimation', async () => {
      await fs.writeFile(
        path.join(tmpDir, 'login.cy.js'),
        `describe('Login', () => { it('visits', () => { cy.visit('/'); cy.get('#user').type('admin'); }); });`
      );

      const result = await estimator.estimate(tmpDir, { from: 'cypress', to: 'playwright' });

      expect(result.summary.totalFiles).toBe(1);
      expect(result.from).toBe('cypress');
      expect(result.to).toBe('playwright');
    });
  });
});
