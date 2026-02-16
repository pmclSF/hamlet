/* eslint-disable no-unused-vars */
// noinspection JSUnusedLocalSymbols

import { describe, it, expect } from 'vitest';

describe('Edge cases', () => {
  it('should handle unused variable patterns', () => {
    const _unused = 'this is intentionally unused';
    const result = 42;
    expect(result).toBe(42);
  });

  it('should handle intentional shadowing', () => {
    const value = 'outer';
    const fn = () => {
      const value = 'inner';
      return value;
    };
    expect(fn()).toBe('inner');
    expect(value).toBe('outer');
  });
});
