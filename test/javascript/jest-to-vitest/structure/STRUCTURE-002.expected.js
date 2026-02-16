import { describe, it, expect } from 'vitest';

describe('StringUtils', () => {
  it('should convert a string to uppercase', () => {
    const input = 'hello world';
    expect(input.toUpperCase()).toBe('HELLO WORLD');
  });

  it('should trim whitespace from both ends', () => {
    const input = '  padded string  ';
    expect(input.trim()).toBe('padded string');
  });

  it('should replace substrings correctly', () => {
    const input = 'foo bar foo';
    const result = input.replaceAll('foo', 'baz');
    expect(result).toBe('baz bar baz');
  });
});
