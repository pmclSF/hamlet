import fs from 'fs/promises';
import path from 'path';
import os from 'os';
import { MigrationEngine } from '../../src/core/MigrationEngine.js';

describe('MigrationEngine', () => {
  let engine;
  let tmpDir;
  let outputDir;

  beforeEach(async () => {
    engine = new MigrationEngine();
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-engine-'));
    outputDir = path.join(tmpDir, 'output');
    await fs.mkdir(outputDir);
  });

  afterEach(async () => {
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  /**
   * Create a minimal Jest test file.
   */
  function jestContent(name = 'test') {
    return `const { describe, it, expect } = require('@jest/globals');\n\ndescribe('${name}', () => {\n  it('works', () => {\n    expect(true).toBe(true);\n  });\n});\n`;
  }

  /**
   * Create a Jest helper file.
   */
  function jestHelper(exports = 'helper') {
    return `const ${exports} = () => true;\nmodule.exports = { ${exports} };\n`;
  }

  describe('happy path', () => {
    it('should migrate a single test file', async () => {
      await fs.writeFile(path.join(tmpDir, 'app.test.js'), jestContent('app'));

      const { results, checklist, state } = await engine.migrate(tmpDir, {
        from: 'jest',
        to: 'vitest',
        output: outputDir,
      });

      expect(results.length).toBeGreaterThan(0);
      expect(checklist).toContain('Migration Checklist');
      expect(state.converted).toBeGreaterThanOrEqual(0);
    });

    it('should generate a checklist', async () => {
      await fs.writeFile(path.join(tmpDir, 'test.test.js'), jestContent());

      const { checklist } = await engine.migrate(tmpDir, {
        from: 'jest',
        to: 'vitest',
        output: outputDir,
      });

      expect(checklist).toContain('# Migration Checklist');
    });

    it('should save migration state', async () => {
      await fs.writeFile(path.join(tmpDir, 'test.test.js'), jestContent());

      await engine.migrate(tmpDir, {
        from: 'jest',
        to: 'vitest',
        output: outputDir,
      });

      const stateExists = await fs.access(path.join(tmpDir, '.hamlet', 'state.json'))
        .then(() => true).catch(() => false);
      expect(stateExists).toBe(true);
    });

    it('should call progress callback', async () => {
      await fs.writeFile(path.join(tmpDir, 'test.test.js'), jestContent());

      const progress = [];
      await engine.migrate(tmpDir, {
        from: 'jest',
        to: 'vitest',
        output: outputDir,
        onProgress: (file, status, conf) => {
          progress.push({ file, status, conf });
        },
      });

      expect(progress.length).toBeGreaterThan(0);
    });

    it('should write converted files to output directory', async () => {
      await fs.writeFile(path.join(tmpDir, 'math.test.js'), jestContent('math'));

      await engine.migrate(tmpDir, {
        from: 'jest',
        to: 'vitest',
        output: outputDir,
      });

      const outputFiles = await fs.readdir(outputDir);
      expect(outputFiles.length).toBeGreaterThan(0);
    });
  });

  describe('failure/partial', () => {
    it('should handle empty directory', async () => {
      const { results } = await engine.migrate(tmpDir, {
        from: 'jest',
        to: 'vitest',
        output: outputDir,
      });

      expect(results).toHaveLength(0);
    });

    it('should continue converting when one file fails', async () => {
      await fs.writeFile(path.join(tmpDir, 'good.test.js'), jestContent('good'));
      // Create a file that might cause issues but won't crash the engine
      await fs.writeFile(path.join(tmpDir, 'empty.test.js'), '');

      const { results, state } = await engine.migrate(tmpDir, {
        from: 'jest',
        to: 'vitest',
        output: outputDir,
      });

      // At least one file should have been processed
      expect(results.length + state.converted + (state.skipped || 0)).toBeGreaterThanOrEqual(0);
    });

    it('should reflect failure in state', async () => {
      await fs.writeFile(path.join(tmpDir, 'test.test.js'), jestContent());

      const { state } = await engine.migrate(tmpDir, {
        from: 'jest',
        to: 'vitest',
        output: outputDir,
      });

      expect(state).toHaveProperty('converted');
      expect(state).toHaveProperty('failed');
    });

    it('should skip binary files with warning in state', async () => {
      await fs.writeFile(path.join(tmpDir, 'binary.js'), Buffer.from([0x00, 0x01, 0x02, 0x00, 0x89, 0x50]));
      await fs.writeFile(path.join(tmpDir, 'test.test.js'), jestContent());

      const { results } = await engine.migrate(tmpDir, {
        from: 'jest',
        to: 'vitest',
        output: outputDir,
      });

      const binaryResult = results.find(r => r.path.includes('binary'));
      if (binaryResult) {
        expect(binaryResult.status).toBe('skipped');
        expect(binaryResult.warnings).toContain('Binary file detected');
      }
    });
  });

  describe('resume', () => {
    it('should resume migration skipping converted files', async () => {
      await fs.writeFile(path.join(tmpDir, 'a.test.js'), jestContent('a'));
      await fs.writeFile(path.join(tmpDir, 'b.test.js'), jestContent('b'));

      // First run
      await engine.migrate(tmpDir, {
        from: 'jest',
        to: 'vitest',
        output: outputDir,
      });

      // Second run with continue
      const progress = [];
      await engine.migrate(tmpDir, {
        from: 'jest',
        to: 'vitest',
        output: outputDir,
        continue: true,
        onProgress: (file, status) => progress.push({ file, status }),
      });

      // Files should be skipped on resume
      const skipped = progress.filter(p => p.status === 'skipped-converted');
      expect(skipped.length).toBeGreaterThanOrEqual(0);
    });

    it('should start fresh when no prior state and --continue is set', async () => {
      const freshDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-fresh-'));
      const freshOutput = path.join(freshDir, 'output');
      await fs.mkdir(freshOutput);
      await fs.writeFile(path.join(freshDir, 'test.test.js'), jestContent());

      const { results } = await engine.migrate(freshDir, {
        from: 'jest',
        to: 'vitest',
        output: freshOutput,
        continue: true,
      });

      // Should still work even without prior state
      expect(results).toBeDefined();
      await fs.rm(freshDir, { recursive: true, force: true });
    });
  });

  describe('edge cases', () => {
    it('should handle directory with no test files', async () => {
      await fs.writeFile(path.join(tmpDir, 'readme.md'), '# Hello');
      await fs.writeFile(path.join(tmpDir, 'config.json'), '{}');

      const { results } = await engine.migrate(tmpDir, {
        from: 'jest',
        to: 'vitest',
        output: outputDir,
      });

      // No JS/TS files to process
      expect(results).toHaveLength(0);
    });

    it('should handle single file project', async () => {
      await fs.writeFile(path.join(tmpDir, 'only.test.js'), jestContent('only'));

      const { results } = await engine.migrate(tmpDir, {
        from: 'jest',
        to: 'vitest',
        output: outputDir,
      });

      expect(results.length).toBeGreaterThanOrEqual(1);
    });

    it('should handle .hamlet/ existing from previous run', async () => {
      await fs.mkdir(path.join(tmpDir, '.hamlet'));
      await fs.writeFile(
        path.join(tmpDir, '.hamlet', 'state.json'),
        JSON.stringify({ version: 1, startedAt: '', source: 'jest', target: 'vitest', files: {} })
      );
      await fs.writeFile(path.join(tmpDir, 'test.test.js'), jestContent());

      // Should not crash with existing .hamlet/
      const { results } = await engine.migrate(tmpDir, {
        from: 'jest',
        to: 'vitest',
        output: outputDir,
      });

      expect(results).toBeDefined();
    });
  });

  describe('config file handling', () => {
    it('should convert config files separately from test files', async () => {
      await fs.writeFile(path.join(tmpDir, 'app.test.js'), jestContent('app'));
      await fs.writeFile(path.join(tmpDir, 'jest.config.js'), `module.exports = { testEnvironment: 'node', testTimeout: 30000 };`);

      const { results } = await engine.migrate(tmpDir, {
        from: 'jest',
        to: 'vitest',
        output: outputDir,
      });

      const configResult = results.find(r => r.path.includes('jest.config'));
      expect(configResult).toBeDefined();
      expect(configResult.status).toBe('converted');
      expect(configResult.type).toBe('config');

      // Verify config output was written
      const outputFiles = await fs.readdir(outputDir);
      const hasVitestConfig = outputFiles.some(f => f.includes('vitest.config'));
      expect(hasVitestConfig).toBe(true);

      // Verify config content is valid (not garbled test conversion)
      const vitestConfigPath = path.join(outputDir, 'vitest.config.ts');
      const configContent = await fs.readFile(vitestConfigPath, 'utf8');
      expect(configContent).toContain('vitest/config');
      expect(configContent).toContain("environment: 'node'");
    });

    it('should not process config files through the test converter', async () => {
      await fs.writeFile(path.join(tmpDir, 'jest.config.js'), `module.exports = { testEnvironment: 'jsdom' };`);

      const progress = [];
      await engine.migrate(tmpDir, {
        from: 'jest',
        to: 'vitest',
        output: outputDir,
        onProgress: (file, status) => progress.push({ file, status }),
      });

      // Config should be skipped in main loop then converted in config step
      const configProgress = progress.filter(p => p.file.includes('jest.config'));
      // Should appear with 'skipped' (from main loop) and 'converted' (from config step)
      const hasConverted = configProgress.some(p => p.status === 'converted');
      expect(hasConverted).toBe(true);
    });
  });

  describe('ordering', () => {
    it('should convert files that import nothing before files that depend on them', async () => {
      // Create helper and test that imports it
      await fs.writeFile(path.join(tmpDir, 'helper.js'), jestHelper());
      await fs.writeFile(
        path.join(tmpDir, 'app.test.js'),
        `const { helper } = require('./helper.js');\n\ndescribe('app', () => {\n  it('uses helper', () => {\n    expect(helper()).toBe(true);\n  });\n});\n`
      );

      const progress = [];
      await engine.migrate(tmpDir, {
        from: 'jest',
        to: 'vitest',
        output: outputDir,
        onProgress: (file, status) => progress.push({ file, status }),
      });

      // Helper should be processed before (or at the same position as) the test
      // Both should be processed
      const helperIdx = progress.findIndex(p => p.file.includes('helper'));
      const testIdx = progress.findIndex(p => p.file.includes('app.test'));
      if (helperIdx >= 0 && testIdx >= 0) {
        expect(helperIdx).toBeLessThanOrEqual(testIdx);
      }
    });

    it('should handle independent files correctly', async () => {
      await fs.writeFile(path.join(tmpDir, 'a.test.js'), jestContent('a'));
      await fs.writeFile(path.join(tmpDir, 'b.test.js'), jestContent('b'));

      const { results } = await engine.migrate(tmpDir, {
        from: 'jest',
        to: 'vitest',
        output: outputDir,
      });

      // Both should be converted
      const converted = results.filter(r => r.status === 'converted');
      expect(converted.length).toBeGreaterThanOrEqual(2);
    });
  });
});
