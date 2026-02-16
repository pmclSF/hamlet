import { describe, it, expect } from 'vitest';

describe('Symbol comparisons', () => {
  it('should compare the same symbol reference', () => {
    const sym = Symbol('test');
    expect(sym).toBe(sym);
  });

  it('should distinguish different symbols with same description', () => {
    const sym1 = Symbol('id');
    const sym2 = Symbol('id');
    expect(sym1).not.toBe(sym2);
  });

  it('should handle Symbol.for with global registry', () => {
    const sym1 = Symbol.for('shared');
    const sym2 = Symbol.for('shared');
    expect(sym1).toBe(sym2);
  });

  it('should verify symbol type', () => {
    const sym = Symbol('example');
    expect(typeof sym).toBe('symbol');
    expect(sym.toString()).toBe('Symbol(example)');
  });

  it('should use symbols as object keys', () => {
    const key = Symbol('key');
    const obj = { [key]: 'value' };
    expect(obj[key]).toBe('value');
  });
});
