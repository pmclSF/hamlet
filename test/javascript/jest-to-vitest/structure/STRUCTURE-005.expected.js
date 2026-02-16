import { describe, test, expect } from 'vitest';

describe('Array utilities', () => {
  test('returns the first element of an array', () => {
    const arr = [10, 20, 30];
    expect(arr[0]).toBe(10);
  });

  test('returns the last element of an array', () => {
    const arr = [10, 20, 30];
    expect(arr[arr.length - 1]).toBe(30);
  });

  test('flattens a nested array one level deep', () => {
    const nested = [[1, 2], [3, 4], [5]];
    const flat = nested.flat();
    expect(flat).toEqual([1, 2, 3, 4, 5]);
  });

  test('filters out falsy values', () => {
    const mixed = [0, 'hello', '', null, undefined, 42, false, 'world'];
    const truthy = mixed.filter(Boolean);
    expect(truthy).toEqual(['hello', 42, 'world']);
  });
});
