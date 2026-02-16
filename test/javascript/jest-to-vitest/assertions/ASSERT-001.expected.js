import { describe, it, expect } from 'vitest';

describe('StringUtils', () => {
  it('should return the correct length', () => {
    const result = getLength('hello');
    expect(result).toBe(5);
  });

  it('should handle empty string', () => {
    expect(getLength('')).toBe(0);
  });
});
