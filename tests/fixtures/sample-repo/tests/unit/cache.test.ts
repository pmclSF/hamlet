import { describe, it, expect } from 'vitest';
import { setCache, getCache, deleteCache } from '../../src/cache/redis.js';

describe('cache', () => {
  describe('setCache / getCache', () => {
    it('should store and retrieve a value', async () => {
      await setCache('key1', 'value1', 60);
      const result = await getCache('key1');
      expect(result).toBe('value1');
    });

    it('should return null for missing key', async () => {
      const result = await getCache('nonexistent');
      expect(result).toBeNull();
    });
  });

  describe('deleteCache', () => {
    it('should remove a cached value', async () => {
      await setCache('del-key', 'val', 60);
      const deleted = await deleteCache('del-key');
      expect(deleted).toBe(true);
    });

    it('should return false for missing key', async () => {
      const deleted = await deleteCache('no-such-key');
      expect(deleted).toBe(false);
    });
  });
});
