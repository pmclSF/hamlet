import { describe, it, expect } from 'vitest';

describe('Positive number validation', () => {
  it.each([1, 2, 3, 100, 999])('should confirm %i is positive', (value) => {
    expect(value).toBeGreaterThan(0);
  });

  it.each([0, -1, -100])('should confirm %i is not positive', (value) => {
    expect(value).toBeLessThanOrEqual(0);
  });
});
