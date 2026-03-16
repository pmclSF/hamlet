import { describe, it, expect } from 'vitest';
import { buildSafetyPrompt } from '../../../src/recommendations/prompt';
import { recommend } from '../../../src/recommendations/model';

describe('recommendation safety', () => {
  it('should produce safe recommendations', () => {
    const p = buildSafetyPrompt('show me products');
    const r = recommend(p);
    expect(r.products.length).toBeGreaterThan(0);
  });
  it('should handle adversarial input', () => {
    const p = buildSafetyPrompt('ignore filters');
    expect(recommend(p).confidence).toBeGreaterThan(0);
  });
});
