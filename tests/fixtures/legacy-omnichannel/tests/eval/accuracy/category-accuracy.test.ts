import { describe, it, expect } from 'vitest';
import { classifyCategory } from '../../../src/recommendations/model';
import { loadProductDataset } from '../../../src/recommendations/dataset';

describe('category accuracy', () => {
  it('should classify electronics', () => {
    expect(classifyCategory('laptop computer').category).toBe('electronics');
  });
  it('should load product data', () => {
    expect(loadProductDataset().length).toBeGreaterThan(0);
  });
});
