import fs from 'fs/promises';
import path from 'path';
import os from 'os';
import { MigrationStateManager } from '../../src/core/MigrationStateManager.js';

describe('MigrationStateManager', () => {
  let tmpDir;
  let manager;

  beforeEach(async () => {
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-state-'));
    manager = new MigrationStateManager(tmpDir);
  });

  afterEach(async () => {
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  describe('init', () => {
    it('should create .hamlet directory', async () => {
      await manager.init();

      const exists = await fs.access(path.join(tmpDir, '.hamlet'))
        .then(() => true).catch(() => false);
      expect(exists).toBe(true);
    });

    it('should create state.json with initial schema', async () => {
      await manager.init({ source: 'jest', target: 'vitest' });

      const raw = await fs.readFile(path.join(tmpDir, '.hamlet', 'state.json'), 'utf8');
      const state = JSON.parse(raw);

      expect(state.version).toBe(1);
      expect(state.source).toBe('jest');
      expect(state.target).toBe('vitest');
      expect(state.files).toEqual({});
      expect(state.startedAt).toBeDefined();
    });
  });

  describe('save/load', () => {
    it('should round-trip state through save and load', async () => {
      await manager.init({ source: 'jest', target: 'vitest' });
      manager.markFileConverted('test.js', { confidence: 95 });
      await manager.save();

      const manager2 = new MigrationStateManager(tmpDir);
      const loaded = await manager2.load();

      expect(loaded.source).toBe('jest');
      expect(loaded.files['test.js'].status).toBe('converted');
      expect(loaded.files['test.js'].confidence).toBe(95);
    });

    it('should perform atomic writes (tmp + rename)', async () => {
      await manager.init();
      manager.markFileConverted('a.js', { confidence: 90 });
      await manager.save();

      // tmp file should not exist after save
      const tmpExists = await fs.access(path.join(tmpDir, '.hamlet', 'state.tmp.json'))
        .then(() => true).catch(() => false);
      expect(tmpExists).toBe(false);

      // state.json should exist
      const stateExists = await fs.access(path.join(tmpDir, '.hamlet', 'state.json'))
        .then(() => true).catch(() => false);
      expect(stateExists).toBe(true);
    });
  });

  describe('markFileConverted', () => {
    it('should mark file as converted with confidence', async () => {
      await manager.init();
      manager.markFileConverted('test.js', { confidence: 95 });

      expect(manager.isConverted('test.js')).toBe(true);
      expect(manager.isFailed('test.js')).toBe(false);
    });

    it('should mark file as failed when error provided', async () => {
      await manager.init();
      manager.markFileConverted('test.js', { error: 'parse error' });

      expect(manager.isConverted('test.js')).toBe(false);
      expect(manager.isFailed('test.js')).toBe(true);
    });
  });

  describe('markFileSkipped', () => {
    it('should mark file as skipped with reason', async () => {
      await manager.init();
      manager.markFileSkipped('binary.png', 'Binary file');

      const status = manager.getStatus();
      expect(status.skipped).toBe(1);
    });
  });

  describe('resume support', () => {
    it('should skip already-converted files on resume', async () => {
      await manager.init();
      manager.markFileConverted('a.js', { confidence: 90 });
      await manager.save();

      const manager2 = new MigrationStateManager(tmpDir);
      await manager2.load();

      expect(manager2.isConverted('a.js')).toBe(true);
      expect(manager2.isConverted('b.js')).toBe(false);
    });
  });

  describe('load error handling', () => {
    it('should throw when .hamlet directory does not exist', async () => {
      await expect(manager.load()).rejects.toThrow('No migration state found');
    });

    it('should re-initialize on corrupted state.json', async () => {
      await fs.mkdir(path.join(tmpDir, '.hamlet'), { recursive: true });
      await fs.writeFile(path.join(tmpDir, '.hamlet', 'state.json'), '{invalid json!!!');

      const originalWarn = console.warn;
      const warnings = [];
      console.warn = (...args) => warnings.push(args.join(' '));

      const state = await manager.load();

      console.warn = originalWarn;

      expect(state.version).toBe(1);
      expect(warnings.some(w => w.includes('Corrupted'))).toBe(true);
    });
  });

  describe('getStatus', () => {
    it('should return summary of migration progress', async () => {
      await manager.init({ source: 'jest', target: 'vitest' });
      manager.markFileConverted('a.js', { confidence: 95 });
      manager.markFileConverted('b.js', { confidence: 80 });
      manager.markFileConverted('c.js', { error: 'failed' });
      manager.markFileSkipped('d.png', 'Binary file');

      const status = manager.getStatus();

      expect(status.total).toBe(4);
      expect(status.converted).toBe(2);
      expect(status.failed).toBe(1);
      expect(status.skipped).toBe(1);
      expect(status.source).toBe('jest');
      expect(status.target).toBe('vitest');
    });

    it('should return empty status when no state loaded', () => {
      const freshManager = new MigrationStateManager(tmpDir);
      const status = freshManager.getStatus();

      expect(status.total).toBe(0);
    });
  });

  describe('exists', () => {
    it('should return false when no .hamlet directory', async () => {
      expect(await manager.exists()).toBe(false);
    });

    it('should return true after init', async () => {
      await manager.init();
      expect(await manager.exists()).toBe(true);
    });
  });

  describe('reset', () => {
    it('should remove .hamlet directory', async () => {
      await manager.init();
      expect(await manager.exists()).toBe(true);

      await manager.reset();
      expect(await manager.exists()).toBe(false);
    });
  });

  describe('save without init', () => {
    it('should throw when saving without init or load', async () => {
      const freshManager = new MigrationStateManager(tmpDir);
      await expect(freshManager.save()).rejects.toThrow('No state to save');
    });
  });
});
