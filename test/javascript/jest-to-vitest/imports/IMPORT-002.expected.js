import { describe, it, expect } from 'vitest';
import { sum } from './math.js';

describe('math', () => {
  it('adds numbers', () => {
    expect(sum(1, 2)).toBe(3);
  });

  it('adds decimal numbers', () => {
    expect(sum(0.1, 0.2)).toBeCloseTo(0.3);
  });

  it('handles large values', () => {
    expect(sum(1000000, 2000000)).toBe(3000000);
  });
});
