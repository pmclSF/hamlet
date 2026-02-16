import { describe, it, expect } from 'vitest';

describe('Calculator', () => {
  it('should add two numbers', () => {
    const a = 5;
    const b = 3;
    const result = a + b;
    expect(result).toBe(8);
  });
});
