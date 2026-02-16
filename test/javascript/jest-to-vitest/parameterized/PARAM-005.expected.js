import { describe, it, expect } from 'vitest';

describe('String length validation', () => {
  it.each([
    { input: 'hello', expected: 5 },
    { input: 'world!', expected: 6 },
    { input: '', expected: 0 },
    { input: 'a', expected: 1 },
  ])('length of "$input" should be $expected', ({ input, expected }) => {
    expect(input.length).toBe(expected);
  });

  it.each([
    { value: 'HELLO', method: 'toLowerCase', expected: 'hello' },
    { value: 'hello', method: 'toUpperCase', expected: 'HELLO' },
  ])('$value.$method() should return "$expected"', ({ value, method, expected }) => {
    expect(value[method]()).toBe(expected);
  });
});
