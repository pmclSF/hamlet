import { describe, it, expect } from 'vitest';

describe('MathUtils', () => {
  it('should calculate the area of a circle', () => {
    const area = circleArea(5);
    expect(area).toBeCloseTo(78.54, 2);
  });

  it('should handle floating point addition', () => {
    const sum = add(0.1, 0.2);
    expect(sum).toBeCloseTo(0.3);
  });

  it('should compute tax with precision', () => {
    const tax = calculateTax(99.99, 0.0825);
    expect(tax).toBeCloseTo(8.25, 2);
  });

  it('should compute average with default precision', () => {
    const avg = average([1.1, 2.2, 3.3]);
    expect(avg).toBeCloseTo(2.2);
  });
});
