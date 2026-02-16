import { describe, it, expect } from 'vitest';

describe('Single parameter case', () => {
  it.each([[42]])('works with %i', (n) => {
    expect(n).toBe(42);
    expect(typeof n).toBe('number');
  });

  it.each([['only-value']])('handles single string %s', (str) => {
    expect(str).toBe('only-value');
    expect(str.length).toBeGreaterThan(0);
  });
});
