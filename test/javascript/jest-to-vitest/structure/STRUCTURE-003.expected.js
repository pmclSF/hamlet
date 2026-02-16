import { describe, it, expect } from 'vitest';

describe('Math', () => {
  describe('addition', () => {
    it('should add positive numbers', () => {
      expect(2 + 3).toBe(5);
    });

    it('should add negative numbers', () => {
      expect(-2 + -3).toBe(-5);
    });
  });

  describe('subtraction', () => {
    it('should subtract positive numbers', () => {
      expect(10 - 4).toBe(6);
    });

    it('should handle subtracting larger from smaller', () => {
      expect(3 - 7).toBe(-4);
    });
  });
});
