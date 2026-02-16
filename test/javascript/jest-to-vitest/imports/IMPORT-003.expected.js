import { describe, it, expect } from 'vitest';

describe('dynamic module loading', () => {
  it('should load the module dynamically', async () => {
    const mod = await import('./module');
    expect(mod.default).toBeDefined();
    expect(typeof mod.default).toBe('function');
  });

  it('should access named exports from dynamic import', async () => {
    const { multiply, divide } = await import('./math-utils');
    expect(multiply(3, 4)).toBe(12);
    expect(divide(10, 2)).toBe(5);
  });
});
