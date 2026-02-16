import { describe, it, expect } from 'vitest';

describe('Empty parameterization', () => {
  it.each([])('should not run with empty array', (value) => {
    expect(value).toBeDefined();
  });

  it('should still run normal tests', () => {
    expect(true).toBe(true);
  });
});
