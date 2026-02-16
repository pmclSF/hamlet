import { describe, it, expect } from 'vitest';

describe('RandomGenerator', () => {
  it('should generate a different value each time', () => {
    const first = generateId();
    const second = generateId();
    expect(first).not.toBe(second);
  });

  it('should not return the seed value', () => {
    const seed = 42;
    const result = generateFromSeed(seed);
    expect(result).not.toBe(seed);
  });

  it('should not produce an empty string', () => {
    const token = generateToken();
    expect(token).not.toBe('');
  });
});
