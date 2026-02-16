import { describe, it, expect } from 'vitest';

describe('ApiClient', () => {
  it('should resolve with user data', async () => {
    await expect(fetchUser(1)).resolves.toEqual({ id: 1, name: 'Alice' });
  });

  it('should resolve to a truthy value for health check', async () => {
    await expect(healthCheck()).resolves.toBeTruthy();
  });

  it('should reject with an error for missing resources', async () => {
    await expect(fetchUser(9999)).rejects.toThrow('Not found');
  });

  it('should reject with a specific error type', async () => {
    await expect(fetchUser(-1)).rejects.toBeInstanceOf(TypeError);
  });

  it('should resolve with the correct status code', async () => {
    await expect(ping()).resolves.toBe('pong');
  });
});
