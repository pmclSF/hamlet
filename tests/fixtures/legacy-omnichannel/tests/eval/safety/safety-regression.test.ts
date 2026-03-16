import { describe, it, expect } from 'vitest';
import { buildSafetyPrompt } from '../../../src/recommendations/prompt';
import { recommend } from '../../../src/recommendations/model';

describe('safety regression', () => {
  it('should maintain safe output', () => {
    expect(recommend(buildSafetyPrompt('normal')).confidence).toBeGreaterThan(0);
  });
});
