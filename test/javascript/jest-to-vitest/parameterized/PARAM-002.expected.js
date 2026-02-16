import { describe, it, expect } from 'vitest';

describe('Addition', () => {
  it.each([
    [1, 2, 3],
    [4, 5, 9],
    [10, 20, 30],
    [-1, -2, -3],
    [0, 0, 0],
  ])('add(%i, %i) = %i', (a, b, expected) => {
    expect(a + b).toBe(expected);
  });
});
