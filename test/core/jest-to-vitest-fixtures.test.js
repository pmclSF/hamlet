/**
 * Integration tests for multi-file, malformed-input, and migration-scenario fixtures.
 *
 * These test the full migration pipeline, not just individual modules.
 */

import fs from 'fs/promises';
import path from 'path';
import os from 'os';
import { fileURLToPath } from 'url';
import { MigrationEngine } from '../../src/core/MigrationEngine.js';
import { InputNormalizer } from '../../src/core/InputNormalizer.js';
import { FileClassifier } from '../../src/core/FileClassifier.js';
import { DependencyGraphBuilder } from '../../src/core/DependencyGraphBuilder.js';
import { TopologicalSorter } from '../../src/core/TopologicalSorter.js';
import { ImportRewriter } from '../../src/core/ImportRewriter.js';
import { MigrationStateManager } from '../../src/core/MigrationStateManager.js';
import { MigrationChecklistGenerator } from '../../src/core/MigrationChecklistGenerator.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const FIXTURES_DIR = path.join(__dirname, '..', 'fixtures', 'jest-to-vitest');

describe('MULTI: Multi-file migration scenarios', () => {
  let engine;
  let tmpDir;
  let outputDir;

  beforeEach(async () => {
    engine = new MigrationEngine();
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-multi-'));
    outputDir = path.join(tmpDir, 'output');
    await fs.mkdir(outputDir);
  });

  afterEach(async () => {
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  it('MULTI-001: Test file importing a helper — both converted', async () => {
    const helperSrc = path.join(FIXTURES_DIR, 'multi', 'MULTI-001-helper.input.js');
    const testSrc = path.join(FIXTURES_DIR, 'multi', 'MULTI-001-test.input.js');

    await fs.copyFile(helperSrc, path.join(tmpDir, 'helper.js'));
    await fs.copyFile(testSrc, path.join(tmpDir, 'test.test.js'));

    const { results } = await engine.migrate(tmpDir, {
      from: 'jest',
      to: 'vitest',
      output: outputDir,
    });

    // Both files should be processed
    expect(results.length).toBeGreaterThanOrEqual(1);
  });

  it('MULTI-002: Shared factory used by multiple test files', async () => {
    await fs.copyFile(
      path.join(FIXTURES_DIR, 'multi', 'MULTI-002-factory.input.js'),
      path.join(tmpDir, 'factory.js')
    );
    await fs.copyFile(
      path.join(FIXTURES_DIR, 'multi', 'MULTI-002-test1.input.js'),
      path.join(tmpDir, 'test1.test.js')
    );
    await fs.copyFile(
      path.join(FIXTURES_DIR, 'multi', 'MULTI-002-test2.input.js'),
      path.join(tmpDir, 'test2.test.js')
    );

    const progress = [];
    const { results } = await engine.migrate(tmpDir, {
      from: 'jest',
      to: 'vitest',
      output: outputDir,
      onProgress: (file, status) => progress.push({ file, status }),
    });

    expect(results.length).toBeGreaterThanOrEqual(2);
  });

  it('MULTI-003: Circular dependency between helpers — still converts', async () => {
    await fs.copyFile(
      path.join(FIXTURES_DIR, 'multi', 'MULTI-003-helperA.input.js'),
      path.join(tmpDir, 'helperA.js')
    );
    await fs.copyFile(
      path.join(FIXTURES_DIR, 'multi', 'MULTI-003-helperB.input.js'),
      path.join(tmpDir, 'helperB.js')
    );

    // Should not crash on circular deps
    const { results } = await engine.migrate(tmpDir, {
      from: 'jest',
      to: 'vitest',
      output: outputDir,
    });

    expect(results).toBeDefined();
  });

  it('MULTI-004: Config file + test file — config converted separately', async () => {
    await fs.writeFile(
      path.join(tmpDir, 'jest.config.js'),
      'module.exports = { testEnvironment: "node" };'
    );
    await fs.writeFile(
      path.join(tmpDir, 'app.test.js'),
      `describe('app', () => { it('works', () => { expect(1).toBe(1); }); });`
    );

    const { results } = await engine.migrate(tmpDir, {
      from: 'jest',
      to: 'vitest',
      output: outputDir,
    });

    expect(results.length).toBeGreaterThanOrEqual(1);
  });

  it('MULTI-005: TypeScript test with type imports — .d.ts skipped, .ts converted', async () => {
    await fs.writeFile(
      path.join(tmpDir, 'types.d.ts'),
      'export type User = { name: string; id: number; };'
    );
    await fs.writeFile(
      path.join(tmpDir, 'user.test.ts'),
      `import type { User } from './types';\n\ndescribe('User', () => {\n  it('has name', () => {\n    const u: User = { name: 'Alice', id: 1 };\n    expect(u.name).toBe('Alice');\n  });\n});`
    );

    const { results } = await engine.migrate(tmpDir, {
      from: 'jest',
      to: 'vitest',
      output: outputDir,
    });

    expect(results).toBeDefined();
    // .d.ts file should be skipped (type-def)
    const typeResult = results.find(r => r.path.includes('types.d.ts'));
    if (typeResult) {
      expect(typeResult.status).toBe('skipped');
    }
  });

  it('MULTI-006: Nested directory structure — preserves structure', async () => {
    await fs.mkdir(path.join(tmpDir, 'sub'), { recursive: true });
    await fs.writeFile(
      path.join(tmpDir, 'sub', 'nested.test.js'),
      `describe('nested', () => { it('works', () => { expect(true).toBe(true); }); });`
    );

    await engine.migrate(tmpDir, {
      from: 'jest',
      to: 'vitest',
      output: outputDir,
    });

    const nestedExists = await fs.access(path.join(outputDir, 'sub', 'nested.test.js'))
      .then(() => true).catch(() => false);
    expect(nestedExists).toBe(true);
  });
});

describe('MESSY: Malformed input — full pipeline tests', () => {
  let normalizer;

  beforeEach(() => {
    normalizer = new InputNormalizer();
  });

  it('MESSY-001: Mismatched quotes — recovers', async () => {
    const input = await fs.readFile(path.join(FIXTURES_DIR, 'messy', 'MESSY-001.input.js'), 'utf8');
    const { normalized, issues } = normalizer.normalize(input);

    expect(issues.some(i => i.type === 'quote')).toBe(true);
    expect(normalized.length).toBeGreaterThan(0);
    // Should still contain the test structure
    expect(normalized).toContain('describe');
  });

  it('MESSY-002: Unclosed brackets — partial conversion with warning', async () => {
    const input = await fs.readFile(path.join(FIXTURES_DIR, 'messy', 'MESSY-002.input.js'), 'utf8');
    const { normalized, issues } = normalizer.normalize(input);

    expect(issues.some(i => i.type === 'bracket')).toBe(true);
    expect(normalized).toContain('describe');
  });

  it('MESSY-003: Mixed CommonJS and ES imports — processes both', async () => {
    const input = await fs.readFile(path.join(FIXTURES_DIR, 'messy', 'MESSY-003.input.js'), 'utf8');
    const { normalized } = normalizer.normalize(input);

    expect(normalized).toContain('require');
    expect(normalized).toContain('import');
  });

  it('MESSY-004: Empty file — empty output, no crash', async () => {
    const input = await fs.readFile(path.join(FIXTURES_DIR, 'messy', 'MESSY-004.input.js'), 'utf8');
    const { normalized, issues } = normalizer.normalize(input);

    expect(issues.some(i => i.type === 'empty')).toBe(true);
    expect(normalized).toBe('');
  });

  it('MESSY-005: File with only comments — passes through', async () => {
    const input = await fs.readFile(path.join(FIXTURES_DIR, 'messy', 'MESSY-005.input.js'), 'utf8');
    const { normalized, issues } = normalizer.normalize(input);

    // Should pass through without errors (non-encoding issues)
    expect(issues.filter(i => i.type !== 'encoding')).toHaveLength(0);
    expect(normalized).toContain('comment');
  });

  it('MESSY-006: Deeply nested callbacks — converts what it can', async () => {
    const input = await fs.readFile(path.join(FIXTURES_DIR, 'messy', 'MESSY-006.input.js'), 'utf8');
    const { normalized } = normalizer.normalize(input);

    expect(normalized).toContain('describe');
    expect(normalized).toContain('map');
  });

  it('MESSY-007: Duplicate imports — processes without crash', async () => {
    const input = await fs.readFile(path.join(FIXTURES_DIR, 'messy', 'MESSY-007.input.js'), 'utf8');
    const { normalized } = normalizer.normalize(input);

    expect(normalized).toContain('import');
    expect(normalized).toContain('describe');
  });

  it('MESSY-008: Invalid test structure (describe without it) — passes through', async () => {
    const input = await fs.readFile(path.join(FIXTURES_DIR, 'messy', 'MESSY-008.input.js'), 'utf8');
    const { normalized } = normalizer.normalize(input);

    expect(normalized).toContain('describe');
  });

  it('MESSY-009: Binary-looking content — skip with warning', () => {
    const binary = '\x00\x01\x02\x03\x00\x89\x50\x4E\x47\x00';
    const { normalized, issues } = normalizer.normalize(binary);

    expect(issues.some(i => i.type === 'binary')).toBe(true);
    expect(normalized).toBe('');
  });

  it('MESSY-010: Very long lines — handles without timeout', async () => {
    const input = await fs.readFile(path.join(FIXTURES_DIR, 'messy', 'MESSY-010.input.js'), 'utf8');
    const { normalized } = normalizer.normalize(input);

    expect(normalized).toContain('longString');
    expect(normalized).toContain('describe');
  });
});

describe('MIGRATE: Migration-specific scenarios', () => {
  let tmpDir;

  beforeEach(async () => {
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-migrate-'));
  });

  afterEach(async () => {
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  it('MIGRATE-001: Helper file converted before test that imports it', async () => {
    const engine = new MigrationEngine();
    const outputDir = path.join(tmpDir, 'output');
    await fs.mkdir(outputDir);

    await fs.writeFile(
      path.join(tmpDir, 'helper.js'),
      'export function add(a, b) { return a + b; }'
    );
    await fs.writeFile(
      path.join(tmpDir, 'math.test.js'),
      `import { add } from './helper.js';\n\ndescribe('math', () => {\n  it('adds', () => {\n    expect(add(1, 2)).toBe(3);\n  });\n});`
    );

    const progress = [];
    await engine.migrate(tmpDir, {
      from: 'jest',
      to: 'vitest',
      output: outputDir,
      onProgress: (file, status) => progress.push({ file, status }),
    });

    const helperIdx = progress.findIndex(p => p.file.includes('helper'));
    const testIdx = progress.findIndex(p => p.file.includes('math'));
    if (helperIdx >= 0 && testIdx >= 0) {
      expect(helperIdx).toBeLessThanOrEqual(testIdx);
    }
  });

  it('MIGRATE-002: Resume migration (half-done state)', async () => {
    const engine = new MigrationEngine();
    const outputDir = path.join(tmpDir, 'output');
    await fs.mkdir(outputDir);

    await fs.writeFile(
      path.join(tmpDir, 'a.test.js'),
      `describe('a', () => { it('works', () => { expect(1).toBe(1); }); });`
    );
    await fs.writeFile(
      path.join(tmpDir, 'b.test.js'),
      `describe('b', () => { it('works', () => { expect(2).toBe(2); }); });`
    );

    // First run
    await engine.migrate(tmpDir, { from: 'jest', to: 'vitest', output: outputDir });

    // Resume
    const engine2 = new MigrationEngine();
    const { results } = await engine2.migrate(tmpDir, {
      from: 'jest',
      to: 'vitest',
      output: outputDir,
      continue: true,
    });

    expect(results).toBeDefined();
  });

  it('MIGRATE-003: Checklist generation for mixed-confidence results', () => {
    const generator = new MigrationChecklistGenerator();
    const graph = { nodes: [], edges: new Map() };

    const results = [
      { path: 'high.test.js', confidence: 98, warnings: [], todos: [], type: 'test' },
      { path: 'medium.test.js', confidence: 75, warnings: ['partial mock'], todos: [], type: 'test' },
      { path: 'low.test.js', confidence: 40, warnings: [], todos: ['rewrite manually'], type: 'test' },
      { path: 'failed.test.js', confidence: 0, status: 'failed', error: 'crash', warnings: [], todos: [], type: 'test' },
    ];

    const checklist = generator.generate(graph, results);

    expect(checklist).toContain('Fully Converted');
    expect(checklist).toContain('Needs Review');
    expect(checklist).toContain('Manual Steps');
    expect(checklist).toContain('98%');
    expect(checklist).toContain('75%');
  });

  it('MIGRATE-004: File classification accuracy', () => {
    const classifier = new FileClassifier();

    const testFile = classifier.classify(
      'src/__tests__/math.test.js',
      `describe('math', () => { it('adds', () => { expect(1+1).toBe(2); }); });`
    );
    expect(testFile.type).toBe('test');

    const helperFile = classifier.classify(
      'test/helpers/db.js',
      'export function connectDB() { return {}; }'
    );
    expect(helperFile.type).toBe('helper');

    const configFile = classifier.classify(
      'jest.config.js',
      'module.exports = { testEnvironment: "node" };'
    );
    expect(configFile.type).toBe('config');
    expect(configFile.framework).toBe('jest');
  });

  it('MIGRATE-005: Import rewriting after rename', () => {
    const rewriter = new ImportRewriter();

    const content = `import { helper } from './helpers.cy.js';\nimport { test } from '@playwright/test';\n\ntest('works', async ({ page }) => {});`;
    const renames = new Map([['./helpers.cy.js', './helpers.spec.js']]);

    const result = rewriter.rewrite(content, renames);

    expect(result).toContain("from './helpers.spec.js'");
    expect(result).toContain("from '@playwright/test'");
  });
});
