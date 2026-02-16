import { describe, it, expect } from 'vitest';

const { sum } = require('./math');

describe('math', () => {
  it('adds numbers', () => {
    expect(sum(1, 2)).toBe(3);
  });

  it('handles zero', () => {
    expect(sum(0, 0)).toBe(0);
  });

  it('handles negative numbers', () => {
    expect(sum(-1, -2)).toBe(-3);
  });
});
