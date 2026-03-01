import fs from 'fs/promises';
import os from 'os';
import path from 'path';
import {
  RepositoryConverter,
  validateRepoUrl,
  convertRepository,
} from '../../src/converter/repoConverter.js';

describe('RepositoryConverter', () => {
  let converter;

  beforeEach(() => {
    converter = new RepositoryConverter();
  });

  describe('constructor', () => {
    it('should initialize with default options', () => {
      expect(converter.options.tempDir).toBe('.hamlet-temp');
      expect(converter.options.batchSize).toBe(5);
      expect(converter.options.preserveStructure).toBe(true);
      expect(converter.options.ignore).toEqual(['node_modules/**', '**/cypress/plugins/**']);
    });

    it('should accept custom options', () => {
      const custom = new RepositoryConverter({
        tempDir: '.custom-temp',
        batchSize: 10,
        preserveStructure: false
      });
      expect(custom.options.tempDir).toBe('.custom-temp');
      expect(custom.options.batchSize).toBe(10);
      expect(custom.options.preserveStructure).toBe(false);
    });

    it('should initialize stats with zero counts', () => {
      expect(converter.stats.totalFiles).toBe(0);
      expect(converter.stats.converted).toBe(0);
      expect(converter.stats.skipped).toBe(0);
      expect(converter.stats.errors).toEqual([]);
    });
  });

  describe('createBatches', () => {
    it('should create batches of the configured size', () => {
      const files = ['a.js', 'b.js', 'c.js', 'd.js', 'e.js', 'f.js', 'g.js'];
      converter.options.batchSize = 3;
      const batches = converter.createBatches(files);
      expect(batches).toHaveLength(3);
      expect(batches[0]).toEqual(['a.js', 'b.js', 'c.js']);
      expect(batches[1]).toEqual(['d.js', 'e.js', 'f.js']);
      expect(batches[2]).toEqual(['g.js']);
    });

    it('should handle empty file list', () => {
      const batches = converter.createBatches([]);
      expect(batches).toEqual([]);
    });

    it('should handle files fewer than batch size', () => {
      const files = ['a.js', 'b.js'];
      converter.options.batchSize = 5;
      const batches = converter.createBatches(files);
      expect(batches).toHaveLength(1);
      expect(batches[0]).toEqual(['a.js', 'b.js']);
    });

    it('should handle exact batch size', () => {
      const files = ['a.js', 'b.js', 'c.js', 'd.js', 'e.js'];
      converter.options.batchSize = 5;
      const batches = converter.createBatches(files);
      expect(batches).toHaveLength(1);
      expect(batches[0]).toHaveLength(5);
    });
  });

  describe('validateRepoUrl', () => {
    it('should accept valid HTTPS URLs', () => {
      expect(() =>
        validateRepoUrl('https://github.com/user/repo.git')
      ).not.toThrow();
      expect(() =>
        validateRepoUrl('https://github.com/user/repo')
      ).not.toThrow();
      expect(() =>
        validateRepoUrl('https://gitlab.com/org/sub/repo.git')
      ).not.toThrow();
    });

    it('should accept valid HTTP URLs', () => {
      expect(() =>
        validateRepoUrl('http://github.com/user/repo.git')
      ).not.toThrow();
    });

    it('should accept valid SSH URLs', () => {
      expect(() =>
        validateRepoUrl('git@github.com:user/repo.git')
      ).not.toThrow();
      expect(() =>
        validateRepoUrl('git@github.com:user/repo')
      ).not.toThrow();
    });

    it('should reject URLs with semicolons (command chaining)', () => {
      expect(() =>
        validateRepoUrl('https://github.com/user/repo; touch /tmp/pwned')
      ).toThrow('Invalid repository URL');
    });

    it('should reject URLs with backticks (command substitution)', () => {
      expect(() =>
        validateRepoUrl('https://github.com/user/`whoami`.git')
      ).toThrow('Invalid repository URL');
    });

    it('should reject URLs with pipe characters', () => {
      expect(() =>
        validateRepoUrl('https://github.com/user/repo|cat /etc/passwd')
      ).toThrow('Invalid repository URL');
    });

    it('should reject URLs with $() command substitution', () => {
      expect(() =>
        validateRepoUrl('https://github.com/user/$(whoami).git')
      ).toThrow('Invalid repository URL');
    });

    it('should reject URLs with null bytes', () => {
      expect(() =>
        validateRepoUrl('https://github.com/user/repo\0.git')
      ).toThrow('Invalid repository URL');
    });

    it('should reject non-string input', () => {
      expect(() => validateRepoUrl(null)).toThrow('Invalid repository URL');
      expect(() => validateRepoUrl(undefined)).toThrow(
        'Invalid repository URL'
      );
      expect(() => validateRepoUrl(42)).toThrow('Invalid repository URL');
      expect(() => validateRepoUrl('')).toThrow('Invalid repository URL');
    });

    it('should reject URLs with unrecognized protocols', () => {
      expect(() => validateRepoUrl('ftp://github.com/user/repo')).toThrow(
        'Invalid repository URL'
      );
      expect(() =>
        validateRepoUrl('file:///etc/passwd')
      ).toThrow('Invalid repository URL');
    });

    it('should reject malformed URLs that pass protocol check', () => {
      expect(() => validateRepoUrl('https://')).toThrow(
        'Invalid repository URL'
      );
      expect(() => validateRepoUrl('git@')).toThrow(
        'Invalid repository URL'
      );
    });
  });

  describe('generateReport', () => {
    it('should return a report with stats', () => {
      converter.stats.totalFiles = 10;
      converter.stats.converted = 8;
      converter.stats.skipped = 2;
      converter.stats.errors.push({ file: 'bad.js', error: 'parse error' });

      const report = converter.generateReport();
      expect(report.stats.totalFiles).toBe(10);
      expect(report.stats.converted).toBe(8);
      expect(report.stats.skipped).toBe(2);
      expect(report.stats.errors).toHaveLength(1);
      expect(report.timestamp).toBeDefined();
      expect(report.configuration).toBeDefined();
    });

    it('should include configuration options in report', () => {
      const report = converter.generateReport();
      expect(report.configuration.tempDir).toBe('.hamlet-temp');
      expect(report.configuration.batchSize).toBe(5);
    });
  });

  describe('convertRepository (function)', () => {
    let tmpDir;

    beforeEach(async () => {
      tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-repo-'));
    });

    afterEach(async () => {
      await fs.rm(tmpDir, { recursive: true, force: true });
    });

    it('should return testResults array and summary counts', async () => {
      const repoDir = path.join(tmpDir, 'repo');
      const outDir = path.join(tmpDir, 'out');
      const testDir = path.join(repoDir, 'cypress', 'e2e');
      await fs.mkdir(testDir, { recursive: true });
      await fs.writeFile(
        path.join(testDir, 'login.cy.js'),
        `
describe('login', () => {
  it('works', () => {
    cy.visit('/');
  });
});
`
      );

      const report = await convertRepository(repoDir, outDir, {
        report: false,
        convertPlugins: false,
      });

      expect(report).toHaveProperty('summary');
      expect(report.summary.totalFiles).toBe(1);
      expect(Array.isArray(report.testResults)).toBe(true);
      expect(report.summary.failedFiles).toBe(0);
    });

    it('should honor from/to options and discover non-Cypress test files', async () => {
      const repoDir = path.join(tmpDir, 'repo-jest');
      const outDir = path.join(tmpDir, 'out-jest');
      await fs.mkdir(repoDir, { recursive: true });
      await fs.writeFile(
        path.join(repoDir, 'sum.test.js'),
        `
import { test, expect } from '@jest/globals';
test('sum', () => {
  expect(1 + 1).toBe(2);
});
`
      );

      const report = await convertRepository(repoDir, outDir, {
        from: 'jest',
        to: 'vitest',
        report: false,
        convertPlugins: false,
      });

      expect(report.summary.totalFiles).toBe(1);
      expect(report.summary.convertedFiles).toBe(1);
      expect(report.testResults[0].output).toContain('sum.test.js');
    });

    it('should write converted plugin output as text when plugins are detected', async () => {
      const repoDir = path.join(tmpDir, 'repo-plugins');
      const outDir = path.join(tmpDir, 'out-plugins');
      await fs.mkdir(path.join(repoDir, 'cypress', 'e2e'), { recursive: true });
      await fs.mkdir(path.join(repoDir, 'cypress', 'plugins'), {
        recursive: true,
      });
      await fs.writeFile(
        path.join(repoDir, 'cypress', 'e2e', 'flow.cy.js'),
        "describe('flow', () => { it('runs', () => { cy.visit('/'); }); });"
      );
      await fs.writeFile(
        path.join(repoDir, 'cypress', 'plugins', 'index.js'),
        "const upload = require('cypress-file-upload'); module.exports = () => {};"
      );

      const report = await convertRepository(repoDir, outDir, {
        from: 'cypress',
        to: 'playwright',
        report: false,
        convertPlugins: true,
      });

      const pluginEntry = report.pluginResults.find((r) => r.status === 'success');
      expect(pluginEntry).toBeDefined();
      const pluginOutput = await fs.readFile(pluginEntry.output, 'utf8');
      expect(pluginOutput).toContain('Converted plugin helpers');
      expect(pluginOutput).toContain('export const pluginConfig');
    });
  });
});
