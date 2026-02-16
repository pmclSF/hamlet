import fs from 'fs/promises';
import path from 'path';
import os from 'os';
import { SafetyManager } from '../../src/core/SafetyManager.js';

describe('SafetyManager', () => {
  let tmpDir;
  let manager;

  beforeEach(async () => {
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-safety-'));
    manager = new SafetyManager(tmpDir);
  });

  afterEach(async () => {
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  describe('atomicWrite', () => {
    it('should create file correctly', async () => {
      const filePath = path.join(tmpDir, 'output.js');
      await manager.atomicWrite(filePath, 'const x = 1;');

      const content = await fs.readFile(filePath, 'utf8');
      expect(content).toBe('const x = 1;');
    });

    it('should not leave tmp file after write', async () => {
      const filePath = path.join(tmpDir, 'output.js');
      await manager.atomicWrite(filePath, 'content');

      const tmpExists = await fs.access(filePath + '.hamlet-tmp')
        .then(() => true).catch(() => false);
      expect(tmpExists).toBe(false);
    });

    it('should create parent directories if needed', async () => {
      const filePath = path.join(tmpDir, 'deep', 'nested', 'output.js');
      await manager.atomicWrite(filePath, 'nested content');

      const content = await fs.readFile(filePath, 'utf8');
      expect(content).toBe('nested content');
    });

    it('should overwrite existing file', async () => {
      const filePath = path.join(tmpDir, 'output.js');
      await fs.writeFile(filePath, 'old content');
      await manager.atomicWrite(filePath, 'new content');

      const content = await fs.readFile(filePath, 'utf8');
      expect(content).toBe('new content');
    });
  });

  describe('backup', () => {
    it('should preserve original file', async () => {
      const filePath = path.join(tmpDir, 'test.js');
      await fs.writeFile(filePath, 'original content');

      const backupPath = await manager.backup(filePath);

      const backupContent = await fs.readFile(backupPath, 'utf8');
      expect(backupContent).toBe('original content');
    });

    it('should store backup in .hamlet/backups/', async () => {
      const filePath = path.join(tmpDir, 'test.js');
      await fs.writeFile(filePath, 'content');

      const backupPath = await manager.backup(filePath);

      expect(backupPath).toContain('.hamlet');
      expect(backupPath).toContain('backups');
    });

    it('should preserve nested path structure in backup', async () => {
      const nestedDir = path.join(tmpDir, 'src', 'tests');
      await fs.mkdir(nestedDir, { recursive: true });
      const filePath = path.join(nestedDir, 'test.js');
      await fs.writeFile(filePath, 'nested test');

      const backupPath = await manager.backup(filePath);

      expect(backupPath).toContain('src');
      const content = await fs.readFile(backupPath, 'utf8');
      expect(content).toBe('nested test');
    });

    it('should not modify the original file', async () => {
      const filePath = path.join(tmpDir, 'test.js');
      await fs.writeFile(filePath, 'dont touch me');

      await manager.backup(filePath);

      const original = await fs.readFile(filePath, 'utf8');
      expect(original).toBe('dont touch me');
    });
  });

  describe('isWithinProject', () => {
    it('should return true for files inside project', () => {
      expect(manager.isWithinProject(path.join(tmpDir, 'src', 'test.js'))).toBe(true);
    });

    it('should return false for files outside project', () => {
      expect(manager.isWithinProject('/tmp/other/file.js')).toBe(false);
    });

    it('should return false for parent directory traversal', () => {
      expect(manager.isWithinProject(path.join(tmpDir, '..', 'escape.js'))).toBe(false);
    });
  });

  describe('registerShutdownHandler', () => {
    it('should return a cleanup function', () => {
      const mockStateManager = { save: async () => {} };
      const cleanup = manager.registerShutdownHandler(mockStateManager);

      expect(typeof cleanup).toBe('function');
      cleanup(); // Clean up the handler
    });
  });
});
