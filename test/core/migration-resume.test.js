/**
 * Migration state resume, retry, and idempotency tests.
 *
 * Uses MigrationStateManager directly (not full MigrationEngine)
 * to test state persistence, resume, and retry semantics.
 */
import fs from 'fs/promises';
import path from 'path';
import os from 'os';
import { MigrationStateManager } from '../../src/core/MigrationStateManager.js';

describe('Migration resume/retry/idempotency', () => {
  let tmpDir;

  beforeEach(async () => {
    tmpDir = await fs.mkdtemp(path.join(os.tmpdir(), 'hamlet-resume-'));
  });

  afterEach(async () => {
    await fs.rm(tmpDir, { recursive: true, force: true });
  });

  describe('initial run produces a state file', () => {
    it('should create .hamlet/state.json after init', async () => {
      const manager = new MigrationStateManager(tmpDir);
      await manager.init({ source: 'jest', target: 'vitest' });

      const statePath = path.join(tmpDir, '.hamlet', 'state.json');
      const exists = await fs
        .access(statePath)
        .then(() => true)
        .catch(() => false);
      expect(exists).toBe(true);

      const raw = await fs.readFile(statePath, 'utf8');
      const state = JSON.parse(raw);
      expect(state.version).toBe(1);
      expect(state.source).toBe('jest');
      expect(state.target).toBe('vitest');
      expect(state.files).toEqual({});
    });
  });

  describe('--continue resumes without redoing completed files', () => {
    it('should skip already-converted files on resume', async () => {
      // First run: mark some files converted, some failed
      const manager1 = new MigrationStateManager(tmpDir);
      await manager1.init({ source: 'jest', target: 'vitest' });
      manager1.markFileConverted('a.test.js', { confidence: 95 });
      manager1.markFileConverted('b.test.js', { confidence: 80 });
      manager1.markFileConverted('c.test.js', { error: 'parse error' });
      await manager1.save();

      // Resume: load existing state
      const manager2 = new MigrationStateManager(tmpDir);
      await manager2.load();

      // Converted files should be skipped
      expect(manager2.isConverted('a.test.js')).toBe(true);
      expect(manager2.isConverted('b.test.js')).toBe(true);
      // Failed file should NOT be marked as converted
      expect(manager2.isConverted('c.test.js')).toBe(false);
      // New file should not be marked
      expect(manager2.isConverted('d.test.js')).toBe(false);
    });

    it('should preserve existing state when adding new conversions', async () => {
      const manager1 = new MigrationStateManager(tmpDir);
      await manager1.init({ source: 'jest', target: 'vitest' });
      manager1.markFileConverted('a.test.js', { confidence: 95 });
      await manager1.save();

      // Resume and convert more
      const manager2 = new MigrationStateManager(tmpDir);
      await manager2.load();
      manager2.markFileConverted('b.test.js', { confidence: 88 });
      await manager2.save();

      // Verify both are persisted
      const manager3 = new MigrationStateManager(tmpDir);
      await manager3.load();
      expect(manager3.isConverted('a.test.js')).toBe(true);
      expect(manager3.isConverted('b.test.js')).toBe(true);
      const status = manager3.getStatus();
      expect(status.converted).toBe(2);
    });
  });

  describe('--retry-failed only retries failed files', () => {
    it('should identify failed files for retry', async () => {
      const manager = new MigrationStateManager(tmpDir);
      await manager.init({ source: 'jest', target: 'vitest' });
      manager.markFileConverted('pass.test.js', { confidence: 90 });
      manager.markFileConverted('fail.test.js', { error: 'timeout' });
      manager.markFileSkipped('skip.test.js', 'Binary file');
      await manager.save();

      // Reload for retry
      const retryMgr = new MigrationStateManager(tmpDir);
      await retryMgr.load();

      // Only the failed file should be identified for retry
      expect(retryMgr.isFailed('fail.test.js')).toBe(true);
      expect(retryMgr.isFailed('pass.test.js')).toBe(false);
      expect(retryMgr.isFailed('skip.test.js')).toBe(false);
    });

    it('should allow re-marking a failed file as converted', async () => {
      const manager = new MigrationStateManager(tmpDir);
      await manager.init({ source: 'jest', target: 'vitest' });
      manager.markFileConverted('flaky.test.js', { error: 'timeout' });
      await manager.save();
      expect(manager.isFailed('flaky.test.js')).toBe(true);

      // Retry succeeds
      manager.markFileConverted('flaky.test.js', { confidence: 75 });
      await manager.save();

      const reloaded = new MigrationStateManager(tmpDir);
      await reloaded.load();
      expect(reloaded.isConverted('flaky.test.js')).toBe(true);
      expect(reloaded.isFailed('flaky.test.js')).toBe(false);
    });
  });

  describe('state idempotency', () => {
    it('should yield the same state when run twice with same inputs', async () => {
      // Run 1
      const mgr1 = new MigrationStateManager(tmpDir);
      await mgr1.init({ source: 'jest', target: 'vitest' });
      mgr1.markFileConverted('a.test.js', { confidence: 90 });
      mgr1.markFileConverted('b.test.js', { error: 'fail' });
      mgr1.markFileSkipped('c.png', 'Binary');
      await mgr1.save();
      const status1 = mgr1.getStatus();

      // "Run 2" — reset and replay the same operations
      await mgr1.reset();
      const mgr2 = new MigrationStateManager(tmpDir);
      await mgr2.init({ source: 'jest', target: 'vitest' });
      mgr2.markFileConverted('a.test.js', { confidence: 90 });
      mgr2.markFileConverted('b.test.js', { error: 'fail' });
      mgr2.markFileSkipped('c.png', 'Binary');
      await mgr2.save();
      const status2 = mgr2.getStatus();

      expect(status1.converted).toBe(status2.converted);
      expect(status1.failed).toBe(status2.failed);
      expect(status1.skipped).toBe(status2.skipped);
      expect(status1.total).toBe(status2.total);
    });

    it('should produce same getStatus() after save→load round-trip', async () => {
      const mgr = new MigrationStateManager(tmpDir);
      await mgr.init({ source: 'cypress', target: 'playwright' });
      mgr.markFileConverted('auth.cy.js', { confidence: 85 });
      mgr.markFileConverted('nav.cy.js', { confidence: 92 });
      mgr.markFileConverted('broken.cy.js', { error: 'unsupported' });
      await mgr.save();

      const statusBefore = mgr.getStatus();

      const mgr2 = new MigrationStateManager(tmpDir);
      await mgr2.load();
      const statusAfter = mgr2.getStatus();

      expect(statusBefore.total).toBe(statusAfter.total);
      expect(statusBefore.converted).toBe(statusAfter.converted);
      expect(statusBefore.failed).toBe(statusAfter.failed);
      expect(statusBefore.source).toBe(statusAfter.source);
      expect(statusBefore.target).toBe(statusAfter.target);
    });
  });
});
