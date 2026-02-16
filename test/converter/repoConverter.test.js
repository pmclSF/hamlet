import { RepositoryConverter } from '../../src/converter/repoConverter.js';
import fs from 'fs/promises';
import path from 'path';
import os from 'os';

describe('RepositoryConverter', () => {
  let converter;

  beforeEach(() => {
    converter = new RepositoryConverter();
  });

  describe('constructor', () => {
    it('should initialize with default options', () => {
      expect(converter.options.batchSize).toBe(5);
      expect(converter.options.preserveStructure).toBe(true);
    });

    it('should accept custom options', () => {
      const custom = new RepositoryConverter({ batchSize: 10 });
      expect(custom.options.batchSize).toBe(10);
    });

    it('should initialize empty stats', () => {
      expect(converter.stats.totalFiles).toBe(0);
      expect(converter.stats.converted).toBe(0);
      expect(converter.stats.errors).toEqual([]);
    });
  });

  describe('analyzeRepository', () => {
    let tmpDir;

    beforeEach(async () => {
      tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-repo-test-'));
    });

    afterEach(async () => {
      await fs.rm(tmpDir, { recursive: true, force: true });
    });

    it('should return testFiles, configs, supportFiles, plugins', async () => {
      const result = await converter.analyzeRepository(tmpDir);
      expect(result).toHaveProperty('testFiles');
      expect(result).toHaveProperty('configs');
      expect(result).toHaveProperty('supportFiles');
      expect(result).toHaveProperty('plugins');
      expect(Array.isArray(result.testFiles)).toBe(true);
      expect(Array.isArray(result.configs)).toBe(true);
      expect(Array.isArray(result.supportFiles)).toBe(true);
      expect(Array.isArray(result.plugins)).toBe(true);
    });

    it('should find cypress test files', async () => {
      const e2eDir = path.join(tmpDir, 'cypress', 'e2e');
      await fs.mkdir(e2eDir, { recursive: true });
      await fs.writeFile(path.join(e2eDir, 'login.cy.js'), 'cy.visit("/login")');

      const result = await converter.analyzeRepository(tmpDir);
      expect(result.testFiles.length).toBe(1);
      expect(result.testFiles[0]).toContain('login.cy.js');
    });

    it('should find config files', async () => {
      await fs.writeFile(path.join(tmpDir, 'cypress.json'), '{}');

      const result = await converter.analyzeRepository(tmpDir);
      expect(result.configs.length).toBe(1);
      expect(result.configs[0]).toContain('cypress.json');
    });

    it('should find support files', async () => {
      const supportDir = path.join(tmpDir, 'cypress', 'support');
      await fs.mkdir(supportDir, { recursive: true });
      await fs.writeFile(path.join(supportDir, 'commands.js'), '// commands');

      const result = await converter.analyzeRepository(tmpDir);
      expect(result.supportFiles.length).toBe(1);
      expect(result.supportFiles[0]).toContain('commands.js');
    });

    it('should find plugin files', async () => {
      const pluginsDir = path.join(tmpDir, 'cypress', 'plugins');
      await fs.mkdir(pluginsDir, { recursive: true });
      await fs.writeFile(path.join(pluginsDir, 'index.js'), '// plugins');

      const result = await converter.analyzeRepository(tmpDir);
      expect(result.plugins.length).toBe(1);
      expect(result.plugins[0]).toContain('index.js');
    });

    it('should return empty arrays for empty repo', async () => {
      const result = await converter.analyzeRepository(tmpDir);
      expect(result.testFiles).toEqual([]);
      expect(result.configs).toEqual([]);
      expect(result.supportFiles).toEqual([]);
      expect(result.plugins).toEqual([]);
    });
  });

  describe('findCypressTests', () => {
    let tmpDir;

    beforeEach(async () => {
      tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-find-test-'));
    });

    afterEach(async () => {
      await fs.rm(tmpDir, { recursive: true, force: true });
    });

    it('should find .cy.js files', async () => {
      await fs.writeFile(path.join(tmpDir, 'test.cy.js'), '// test');
      const files = await converter.findCypressTests(tmpDir);
      expect(files.length).toBe(1);
    });

    it('should find files in cypress/e2e directory', async () => {
      const dir = path.join(tmpDir, 'cypress', 'e2e');
      await fs.mkdir(dir, { recursive: true });
      await fs.writeFile(path.join(dir, 'spec.js'), '// test');
      const files = await converter.findCypressTests(tmpDir);
      expect(files.length).toBe(1);
    });
  });

  describe('createBatches', () => {
    it('should split files into batches', () => {
      const files = ['a.js', 'b.js', 'c.js', 'd.js', 'e.js', 'f.js'];
      const batches = converter.createBatches(files);
      expect(batches.length).toBe(2);
      expect(batches[0].length).toBe(5);
      expect(batches[1].length).toBe(1);
    });

    it('should handle empty array', () => {
      expect(converter.createBatches([])).toEqual([]);
    });
  });

  describe('generateReport', () => {
    it('should include stats and timestamp', () => {
      const report = converter.generateReport();
      expect(report).toHaveProperty('stats');
      expect(report).toHaveProperty('timestamp');
      expect(report.stats.totalFiles).toBe(0);
    });
  });
});
