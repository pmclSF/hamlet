import { describe, it, expect, vi } from 'vitest';

const mockCallback = vi.fn();
const mockFormatter = vi.fn((x) => x.toUpperCase());

describe('Array utilities', () => {
  it('calls the callback for each item', () => {
    forEach([1, 2, 3], mockCallback);
    expect(mockCallback).toHaveBeenCalledTimes(3);
  });

  it('uses the mock formatter', () => {
    const result = mockFormatter('hello');
    expect(result).toBe('HELLO');
    expect(mockFormatter).toHaveBeenCalledWith('hello');
  });
});
